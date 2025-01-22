package queries

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jira-clone-api/common/response"
	respErr "jira-clone-api/common/response/error"
	"jira-clone-api/database/mongo"
	"jira-clone-api/database/mongo/models"
)

type TokenQuery interface {
	Create(data models.Token) (id primitive.ObjectID, err error)
	GetById(id primitive.ObjectID, opts ...OptionsQuery) (webToken *models.Token, err error)
	DeleteById(id primitive.ObjectID) error
	DeleteByUserId(userId primitive.ObjectID) error
}

type webTokenQuery struct {
	collection *mongoDriver.Collection
	context    context.Context
}

func NewToken(ctx context.Context) TokenQuery {
	return &webTokenQuery{
		collection: mongo.NewUtilityService().GetTokenCollection(),
		context:    ctx,
	}
}

func (q *webTokenQuery) GetById(id primitive.ObjectID, opts ...OptionsQuery) (*models.Token, error) {
	opt := NewOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}
	optFind := &options.FindOneOptions{
		Projection: opt.QueryOnlyField(),
	}

	ctx, cancel := timeoutFunc(q.context)
	defer cancel()
	var webToken models.Token
	if err := q.collection.FindOne(ctx, bson.M{
		"_id": id,
	}, optFind).Decode(&webToken); err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return nil, response.NewError(fiber.StatusNotFound, response.ErrorOptions{Data: respErr.ErrResourceNotFound})
		}
		logger.Error().Err(err).Str("function", "GetById").Str("functionInline", "q.collection.FindOne.Decode").Msg("webTokenQuery")
		return nil, response.NewError(fiber.StatusInternalServerError)
	}

	return &webToken, nil
}

func (q *webTokenQuery) Create(data models.Token) (id primitive.ObjectID, err error) {
	currentTime := time.Now()
	data.UpdatedAt = currentTime
	data.CreatedAt = currentTime
	ctx, cancel := timeoutFunc(q.context)
	defer cancel()
	result, err := q.collection.InsertOne(ctx, data)
	if err != nil {
		logger.Error().Err(err).Str("function", "Create").Str("functionInline", "q.collection.InsertOne").Msg("webTokenQuery")
		return id, response.NewError(fiber.StatusInternalServerError)
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (q *webTokenQuery) DeleteById(id primitive.ObjectID) error {
	if _, err := q.collection.DeleteOne(q.context, bson.M{"_id": id}); err != nil {
		logger.Error().Err(err).Str("function", "DeleteById").Str("functionInline", "q.collection.DeleteOne").Msg("webTokenQuery")
		return response.NewError(fiber.StatusInternalServerError)
	}
	return nil
}

func (q *webTokenQuery) DeleteByUserId(userId primitive.ObjectID) error {
	if _, err := q.collection.DeleteMany(q.context, bson.M{"user_id": userId}); err != nil {
		logger.Error().Err(err).Str("function", "DeleteByUserId").Str("functionInline", "q.collection.DeleteMany").Msg("webTokenQuery")
		return response.NewError(fiber.StatusInternalServerError)
	}
	return nil
}
