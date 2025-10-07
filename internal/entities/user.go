package entities

import (
	"time"

	"github.com/go-telegram/bot/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserEntity struct {
	MongoID        	primitive.ObjectID 	`bson:"_id,omitempty" json:"id"`
	models.User 						`bson:",inline" json:",inline"`
	CreatedAt 		time.Time          	`bson:"createdAt" json:"createdAt"`
	UpdatedAt 		time.Time          	`bson:"updatedAt" json:"updatedAt"`
}