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

type WorkspaceQuery interface {
	GetById(id primitive.ObjectID, opts ...OptionsQuery) (workspace *models.Workspace, err error)
	Create(workspace models.Workspace) (newWorkspace *models.Workspace, err error)
}

type workspaceQuery struct {
	collection *mongoDriver.Collection
	context    context.Context
}

func NewWorkspace(ctx context.Context) WorkspaceQuery {
	return &workspaceQuery{
		collection: mongo.NewUtilityService().GetWorkspaceCollection(),
		context:    ctx,
	}
}

func (q *workspaceQuery) GetById(id primitive.ObjectID, opts ...OptionsQuery) (*models.Workspace, error) {
	opt := NewOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}
	var data models.Workspace
	optFind := &options.FindOneOptions{Projection: opt.QueryOnlyField()}
	ctx, cancel := timeoutFunc(q.context)
	defer cancel()
	if err := q.collection.FindOne(ctx, bson.M{"_id": id}, optFind).Decode(&data); err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return nil, response.NewError(fiber.StatusNotFound, response.ErrorOptions{Data: "Workspace not found"})
		}
		logger.Error().Err(err).Str("function", "GetById").Str("functionInline", "q.collection.FindOne").Msg("workspaceQuery")
		return nil, response.NewError(fiber.StatusInternalServerError)
	}
	return &data, nil
}

func (q *workspaceQuery) Create(data models.Workspace) (workspace *models.Workspace, err error) {
	currentTime := time.Now()
	data.UpdatedAt = currentTime
	data.CreatedAt = currentTime
	ctx, cancel := timeoutFunc(q.context)
	defer cancel()
	result, err := q.collection.InsertOne(ctx, data)
	if err != nil {
		if mongoDriver.IsDuplicateKeyError(err) {
			return nil, response.NewError(fiber.StatusConflict, response.ErrorOptions{Data: "Workspace already exists"})
		}
		logger.Error().Err(err).Str("function", "Create").Str("functionInline", "q.collection.InsertOne").Msg("workspaceQuery")
		return nil, response.NewError(fiber.StatusInternalServerError)
	}
	data.Id = result.InsertedID.(primitive.ObjectID)
	return &data, nil
}
