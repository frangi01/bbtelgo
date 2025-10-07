package repo

import (
	"context"
	"errors"
	"time"

	"github.com/frangi01/bbtelgo/internal/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MessageErrNotFound = errors.New("message not found")

type MessageRepository struct {
	col *mongo.Collection
}

func NewMessageRepository(client *mongo.Client, dbName string) (*MessageRepository, error) {
	col := client.Database(dbName).Collection("messages")

	// Indexes:
	// 1) Unique on (chat.id, messageid)
	// 2) Non-unique on date for range query (timeline)
	// 3) Non-unique on from.id to filter by sender
	_, err := col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "chat.id", Value: 1},
				{Key: "id", Value: 1}, // models.Message.MessageID -> "messageid"
			},
			Options: options.Index().SetUnique(true).SetName("uniq_chatId_messageId"),
		},
		{
			Keys:    bson.D{{Key: "date", Value: -1}}, // models.Message.Date (unix)
			Options: options.Index().SetName("idx_date"),
		},
		{
			Keys:    bson.D{{Key: "from.id", Value: 1}}, // sender Telegram
			Options: options.Index().SetName("idx_fromId"),
		},
	})
	if err != nil {
		return nil, err
	}
	return &MessageRepository{col: col}, nil
}

// Create: inserts a new message (fails if it violates the unique index)
func (r *MessageRepository) Create(ctx context.Context, m *entities.MessageEntity) (primitive.ObjectID, error) {
	if m.MongoID.IsZero() {
		m.MongoID = primitive.NewObjectID()
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now

	res, err := r.col.InsertOne(ctx, m)
	if err != nil {
		return primitive.NilObjectID, err
	}
	oid, _ := res.InsertedID.(primitive.ObjectID)
	return oid, nil
}

// UpsertByChatAndMessageID: idempotent on key (chat.id, messageid)
// If it exists -> updates Telegram payload + updatedAt
// If it does not exist -> inserts with _id and createdAt
func (r *MessageRepository) UpsertByChatAndMessageID(ctx context.Context, m *entities.MessageEntity) (created bool, oid primitive.ObjectID, err error) {
	// if m.Chat == nil {
	// 	return false, primitive.NilObjectID, fmt.Errorf("message.Chat is nil")
	// }
	now := time.Now().UTC()

	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":       primitive.NewObjectID(),
			"createdAt": now,
		},
	}

	// We serialize the entire struct, then remove the fields handled separately
	raw, err := bson.Marshal(m)
	if err != nil {
		return false, primitive.NilObjectID, err
	}
	var setDoc bson.M
	if err := bson.Unmarshal(raw, &setDoc); err != nil {
		return false, primitive.NilObjectID, err
	}
	delete(setDoc, "_id")
	delete(setDoc, "createdAt")
	setDoc["updatedAt"] = now
	update["$set"] = setDoc

	filter := bson.M{
		"chat.id":   m.Chat.ID,
		"messageid": m.Message.ID,
	}

	opts := options.Update().SetUpsert(true)
	res, err := r.col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return false, primitive.NilObjectID, err
	}

	// If entered from scratch
	if res.UpsertedID != nil {
		if id, ok := res.UpsertedID.(primitive.ObjectID); ok {
			return true, id, nil
		}
	}

	// It was an update: I take _id
	var out struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := r.col.FindOne(ctx, filter, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&out); err != nil {
		return false, primitive.NilObjectID, err
	}
	return false, out.ID, nil
}

// FindByObjectID
func (r *MessageRepository) FindByObjectID(ctx context.Context, id primitive.ObjectID) (*entities.MessageEntity, error) {
	var m entities.MessageEntity
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, MessageErrNotFound
	}
	return &m, err
}

// FindByChatAndMessageID
func (r *MessageRepository) FindByChatAndMessageID(ctx context.Context, chatID int64, messageID int) (*entities.MessageEntity, error) {
	var m entities.MessageEntity
	err := r.col.FindOne(ctx, bson.M{"chat.id": chatID, "messageid": messageID}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, MessageErrNotFound
	}
	return &m, err
}

// Update: partial update for _id + touch UpdatedAt
func (r *MessageRepository) Update(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	set["updatedAt"] = time.Now().UTC()
	res, err := r.col.UpdateByID(ctx, id, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return MessageErrNotFound
	}
	return nil
}

// Delete per _id
func (r *MessageRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return MessageErrNotFound
	}
	return nil
}

// Options di lista
type MessageListOptions struct {
	Page     int64
	PerPage  int64
	ChatID   *int64 // chat filter
	FromID   *int64 // sender filter (Telegram user id)
	DateFrom *int64 // unix seconds (Telegram Message.Date)
	DateTo   *int64 // unix seconds (exclusive)
	TextLike *string
}

// List: paginated + common filters
func (r *MessageRepository) List(ctx context.Context, opt MessageListOptions) ([]entities.MessageEntity, int64, error) {
	if opt.Page <= 0 {
		opt.Page = 1
	}
	if opt.PerPage <= 0 || opt.PerPage > 1000 {
		opt.PerPage = 50
	}

	filter := bson.M{}
	if opt.ChatID != nil {
		filter["chat.id"] = *opt.ChatID
	}
	if opt.FromID != nil {
		filter["from.id"] = *opt.FromID
	}
	// range on dates (models.Message.Date is int, unix sec)
	if opt.DateFrom != nil || opt.DateTo != nil {
		rg := bson.M{}
		if opt.DateFrom != nil {
			rg["$gte"] = *opt.DateFrom
		}
		if opt.DateTo != nil {
			rg["$lt"] = *opt.DateTo
		}
		filter["date"] = rg
	}
	if opt.TextLike != nil && *opt.TextLike != "" {
		// semplice regex case-sensitive; per i18n/case-insensitive usa collation
		filter["text"] = bson.M{"$regex": *opt.TextLike}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := options.Find().
		SetSort(bson.D{{Key: "date", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip((opt.Page - 1) * opt.PerPage).
		SetLimit(opt.PerPage)

	cur, err := r.col.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var out []entities.MessageEntity
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
