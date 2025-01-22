package local

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"jira-clone-api/database/mongo/models"
)

func (s service) SetUser(value models.User) {
	s.context.Locals(KeyUser, value)
}

func (s service) GetUser() models.User {
	if value, ok := s.context.Locals(KeyUser).(models.User); ok {
		return value
	}
	return models.User{}
}

func (s service) SetExtraBody(value []byte) {
	s.context.Locals(KeyExtraBody, value)
}

func (s service) GetExtraBody() string {
	if value, ok := s.context.Locals(KeyExtraBody).([]byte); ok {
		return string(value)
	}
	return "{}"
}

func (s service) GetStatusCode() int {
	if value, ok := s.context.Locals(KeyStatusCode).(int); ok {
		return value
	}
	return fiber.StatusInternalServerError
}

func (s service) SetStatusCode(value int) {
	s.context.Locals(KeyStatusCode, value)
}

func (s service) SetTokenId(value primitive.ObjectID) {
	s.context.Locals(KeyTokenId, value)
}

func (s service) GetTokenId() primitive.ObjectID {
	if value, ok := s.context.Locals(KeyTokenId).(primitive.ObjectID); ok {
		return value
	}
	return primitive.NilObjectID
}
