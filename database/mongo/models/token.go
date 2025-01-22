package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Token struct {
	UpdatedAt time.Time          `bson:"updated_at"`
	CreatedAt time.Time          `bson:"created_at"`
	ExpiredAt time.Time          `bson:"expired_at"`
	UserId    primitive.ObjectID `bson:"user_id"`
	Id        primitive.ObjectID `bson:"_id,omitempty"`
}

func (m *Token) CollectionName() string {
	return "tokens"
}
