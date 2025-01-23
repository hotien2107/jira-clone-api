package mongo

import (
	"context"

	"jira-clone-api/common/configure"
	"jira-clone-api/common/logging"

	"go.mongodb.org/mongo-driver/bson"

	"go.elastic.co/apm/module/apmmongo/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	jiraDBClient *mongo.Client
	logger       = logging.GetLogger()
	cfg          = configure.GetConfig()
	utils        = NewUtilityService()
)

func InitDatabase() {
	jiraDBClient = initClientConnection(cfg.MongoDBJiraUri, cfg.ElasticAPMEnable)
	autoIndexing()
}

func initClientConnection(mongoURI string, enableAPM bool) *mongo.Client {
	opts := options.Client()
	opts.ApplyURI(mongoURI)
	if enableAPM {
		opts.SetMonitor(apmmongo.CommandMonitor())
	}
	ctx, cancel := utils.GetContextTimeout(context.Background())
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		logger.Fatal().Err(err).Str("function", "initClientConnection").Str("functionInline", "mongo.Connect").Msg("database")
	}
	ctxPing, cancelPing := utils.GetContextTimeout(context.Background())
	defer cancelPing()
	if err = client.Ping(ctxPing, nil); err != nil {
		logger.Fatal().Err(err).Str("function", "initClientConnection").Str("functionInline", "client.Ping").Msg("database")
	}
	return client
}

func DisconnectDatabase() {
	_ = jiraDBClient.Disconnect(context.Background())
}

func autoIndexing() {
	if !cfg.MongoAutoIndexing {
		return
	}
	jiraUserIndex()
	jiraTokenIndex()
	jiraWorkspaceIndex()
}

func jiraUserIndex() {
	collIndex := utils.GetUserCollection().Indexes()
	ctxDrop, cancelDrop := utils.GetContextTimeout(context.Background())
	defer cancelDrop()
	_, _ = collIndex.DropAll(ctxDrop)
	ctx, cancel := utils.GetContextTimeout(context.Background())
	defer cancel()
	if _, err := collIndex.CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}); err != nil {
		logger.Fatal().Err(err).Msg("jiraUserIndex")
	}
}

func jiraTokenIndex() {
	collIndex := utils.GetTokenCollection().Indexes()
	ctxDrop, cancelDrop := utils.GetContextTimeout(context.Background())
	defer cancelDrop()
	_, _ = collIndex.DropAll(ctxDrop)
	ctx, cancel := utils.GetContextTimeout(context.Background())
	defer cancel()
	if _, err := collIndex.CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "expired_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}); err != nil {
		logger.Fatal().Err(err).Msg("jiraTokenIndex")
	}
}

func jiraWorkspaceIndex() {
	collIndex := utils.GetWorkspaceCollection().Indexes()
	ctxDrop, cancelDrop := utils.GetContextTimeout(context.Background())
	defer cancelDrop()
	_, _ = collIndex.DropAll(ctxDrop)
	ctx, cancel := utils.GetContextTimeout(context.Background())
	defer cancel()
	if _, err := collIndex.CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}); err != nil {
		logger.Fatal().Err(err).Msg("jiraWorkspaceIndex")
	}
}
