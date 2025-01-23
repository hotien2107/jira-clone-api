package mongo

import (
	"context"

	mongoModels "jira-clone-api/database/mongo/models"

	"go.mongodb.org/mongo-driver/mongo"
)

type UtilityService interface {
	GetContextTimeout(ctx context.Context) (context.Context, context.CancelFunc)
	GetUserCollection() (coll *mongo.Collection)
	GetTokenCollection() (coll *mongo.Collection)
	GetWorkspaceCollection() (coll *mongo.Collection)
}

type utilityService struct{}

func NewUtilityService() UtilityService {
	service := utilityService{}
	return &service
}

func (s *utilityService) GetContextTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, cfg.MongoDBRequestTimeout)
}

func (s *utilityService) getJiraDB() (db *mongo.Database) {
	return jiraDBClient.Database(cfg.MongoDBJiraName)
}

func (s *utilityService) GetUserCollection() (coll *mongo.Collection) {
	return s.getJiraDB().Collection(new(mongoModels.User).CollectionName())
}

func (s *utilityService) GetTokenCollection() (coll *mongo.Collection) {
	return s.getJiraDB().Collection(new(mongoModels.Token).CollectionName())
}

func (s *utilityService) GetWorkspaceCollection() (coll *mongo.Collection) {
	return s.getJiraDB().Collection(new(mongoModels.Workspace).CollectionName())
}
