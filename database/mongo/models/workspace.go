package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Workspace struct {
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Name      string             `bson:"name"`
	UserId    primitive.ObjectID `bson:"user_id"`
	Id        primitive.ObjectID `bson:"_id,omitempty"`
}

func (m *Workspace) CollectionName() string {
	return "workspaces"
}
