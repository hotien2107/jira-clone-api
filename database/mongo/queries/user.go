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
	"jira-clone-api/database/mongo"
	"jira-clone-api/database/mongo/models"
)

type UserQuery interface {
	GetById(id primitive.ObjectID, opts ...OptionsQuery) (user *models.User, err error)
	Create(user models.User) (newUser *models.User, err error)
	GetByUsername(username string, opts ...OptionsQuery) (user *models.User, err error)
}

type userQuery struct {
	collection *mongoDriver.Collection
	context    context.Context
}

func NewUser(ctx context.Context) UserQuery {
	return &userQuery{
		collection: mongo.NewUtilityService().GetUserCollection(),
		context:    ctx,
	}
}

func (q *userQuery) GetById(id primitive.ObjectID, opts ...OptionsQuery) (*models.User, error) {
	opt := NewOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}
	var data models.User
	optFind := &options.FindOneOptions{Projection: opt.QueryOnlyField()}
	ctx, cancel := timeoutFunc(q.context)
	defer cancel()
	if err := q.collection.FindOne(ctx, bson.M{"_id": id}, optFind).Decode(&data); err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return nil, response.NewError(fiber.StatusNotFound, response.ErrorOptions{Data: "User not found"})
		}
		logger.Error().Err(err).Str("function", "GetById").Str("functionInline", "q.collection.FindOne").Msg("userQuery")
		return nil, response.NewError(fiber.StatusInternalServerError)
	}
	return &data, nil
}

func (q *userQuery) Create(data models.User) (user *models.User, err error) {
	currentTime := time.Now()
	data.UpdatedAt = currentTime
	data.CreatedAt = currentTime
	ctx, cancel := timeoutFunc(q.context)
	defer cancel()
	result, err := q.collection.InsertOne(ctx, data)
	if err != nil {
		if mongoDriver.IsDuplicateKeyError(err) {
			return nil, response.NewError(fiber.StatusConflict, response.ErrorOptions{Data: "User already exists"})
		}
		logger.Error().Err(err).Str("function", "Create").Str("functionInline", "q.collection.InsertOne").Msg("userQuery")
		return nil, response.NewError(fiber.StatusInternalServerError)
	}
	data.Id = result.InsertedID.(primitive.ObjectID)
	return &data, nil
}

func (q *userQuery) GetByUsername(username string, opts ...OptionsQuery) (*models.User, error) {
	opt := NewOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}
	var data models.User
	optFind := &options.FindOneOptions{Projection: opt.QueryOnlyField()}
	ctx, cancel := timeoutFunc(q.context)
	defer cancel()
	if err := q.collection.FindOne(ctx, bson.M{"username": username}, optFind).Decode(&data); err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return nil, response.NewError(fiber.StatusNotFound, response.ErrorOptions{Data: "User not found"})
		}
		logger.Error().Err(err).Str("function", "GetByUsername").Str("functionInline", "q.collection.FindOne").Msg("userQuery")
		return nil, response.NewError(fiber.StatusInternalServerError)
	}
	return &data, nil
}
