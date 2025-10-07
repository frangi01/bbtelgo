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


var UserErrNotFound = errors.New("user not found")

type UserRepository struct {
	col *mongo.Collection
}

func NewUserRepository(client *mongo.Client, dbName string) (*UserRepository, error) {
	col := client.Database(dbName).Collection("users")

	// Unique index on Telegram ID (field "id" in the BSON doc)
	_, err := col.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_telegram_id"),
		},
		// Optional: index on username (not unique: it can change)
		{
			Keys: bson.D{{Key: "username", Value: 1}},
			Options: options.Index().
				SetName("idx_username").
				SetSparse(true),
		},
	})
	if err != nil {
		return nil, err
	}
	return &UserRepository{col: col}, nil
}

// Create: insert document complete; if you want to prevent duplicates, use UpsertByTelegramID
func (r *UserRepository) Create(ctx context.Context, u *entities.UserEntity) (primitive.ObjectID, error) {
	if u.MongoID.IsZero() {
		u.MongoID = primitive.NewObjectID()
	}
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now

	res, err := r.col.InsertOne(ctx, u)
	if err != nil {
		return primitive.NilObjectID, err
	}
	oid, _ := res.InsertedID.(primitive.ObjectID)
	return oid, nil
}

// UpsertByTelegramID: create/update from the Telegram payload (Telegram ID is the key)
func (r *UserRepository) UpsertByTelegramID(ctx context.Context, u *entities.UserEntity) (created bool, oid primitive.ObjectID, err error) {
	now := time.Now().UTC()

	// $set with the entire Telegram struct inline; setOnInsert for _id/createdAt
	update := bson.M{
		"$set": bson.M{
			// Inline all Telegram fields (the driver will encode the fields at the root)
			// Passing 'u' directly is fine: driver encoding -> root keys.
			// If you want to be more explicit: replace with a bson.D and map the fields.
		},
		"$setOnInsert": bson.M{
			"_id":       primitive.NewObjectID(),
			"createdAt": now,
		},
	}

	// WARNING: to use $set with a struct at the root, use bson.Raw from Marshal or use $replaceRoot in an aggregation.
	//  The simple and safe way is: serialize it yourself into a document and then merge it into the keys of $set:

	raw, err := bson.Marshal(u)
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

	

	opts := options.Update().SetUpsert(true)
	res, err := r.col.UpdateOne(ctx, bson.M{"id": u.ID}, update, opts)
	if err != nil {
		// duplicate key -> could indicate a collision on username if you make it unique; here we rethrow the error
		return false, primitive.NilObjectID, err
	}

	// If inserted anew, UpsertedID is populated
	if res.UpsertedID != nil {
		if id, ok := res.UpsertedID.(primitive.ObjectID); ok {
			return true, id, nil
		}
	}

	// Otherwise it was update: retrieve _id
	var doc struct {
		MongoID primitive.ObjectID `bson:"_id"`
	}
	if err := r.col.FindOne(ctx, bson.M{"id": u.ID}, options.FindOne().SetProjection(bson.M{"_id": 1})).Decode(&doc); err != nil {
		return false, primitive.NilObjectID, err
	}
	return false, doc.MongoID, nil
}

// FindByObjectID
func (r *UserRepository) FindByObjectID(ctx context.Context, id primitive.ObjectID) (*entities.UserEntity, error) {
	var u entities.UserEntity
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, UserErrNotFound
	}
	return &u, err
}

// FindByTelegramID
func (r *UserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*entities.UserEntity, error) {
	var u entities.UserEntity
	err := r.col.FindOne(ctx, bson.M{"id": telegramID}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, UserErrNotFound
	}
	return &u, err
}

// FindByUsername (default case-sensitive; for case-insensitive use collation)
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entities.UserEntity, error) {
	var u entities.UserEntity
	err := r.col.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, UserErrNotFound
	}
	return &u, err
}

// Update partial (by _id) with $set and touch UpdatedAt
func (r *UserRepository) Update(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	set["updatedAt"] = time.Now().UTC()
	res, err := r.col.UpdateByID(ctx, id, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return UserErrNotFound
	}
	return nil
}

// Delete (by _id)
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return UserErrNotFound
	}
	return nil
}

// List with filters
type ListOptions struct {
	Page          int64
	PerPage       int64
	UsernamePrefix *string // es. autocompletion
}

func (r *UserRepository) List(ctx context.Context, opt ListOptions) ([]entities.UserEntity, int64, error) {
	if opt.Page <= 0 {
		opt.Page = 1
	}
	if opt.PerPage <= 0 || opt.PerPage > 1000 {
		opt.PerPage = 20
	}
	filter := bson.M{}
	if opt.UsernamePrefix != nil && *opt.UsernamePrefix != "" {
		filter["username"] = bson.M{"$regex": "^" + *opt.UsernamePrefix}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpt := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip((opt.Page - 1) * opt.PerPage).
		SetLimit(opt.PerPage)

	cur, err := r.col.Find(ctx, filter, findOpt)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var out []entities.UserEntity
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}