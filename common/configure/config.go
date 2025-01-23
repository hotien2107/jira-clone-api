package configure

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var config *Configuration

type Configuration struct {
	Host                  string        `env:"HOST" envDefault:"0.0.0.0"`
	Port                  string        `env:"PORT" envDefault:"8080"`
	TokenType             string        `env:"TOKEN_TYPE" envDefault:"Bearer"`
	TokenPublicKey        string        `env:"TOKEN_PUBLIC_KEY_PATH,file" envDefault:"certs/public.pem" envExpand:"true"`
	TokenPrivateKey       string        `env:"TOKEN_PRIVATE_KEY_PATH,file" envDefault:"certs/private.pem" envExpand:"true"`
	MongoDBJiraUri        string        `env:"MONGODB_JIRA_URI" envDefault:"mongodb://localhost:27017"`
	MongoDBJiraName       string        `env:"MONGODB_JIRA_NAME" envDefault:"db_jira"`
	AwsAccessKeyId        string        `env:"AWS_ACCESS_KEY_ID" envDefault:"!change_me!"`
	AwsSecretAccessKey    string        `env:"AWS_SECRET_ACCESS_KEY" envDefault:"!change_me!"`
	AwsRegion             string        `env:"AWS_REGION" envDefault:"!change_me!"`
	AwsEndpoint           string        `env:"AWS_ENDPOINT_URL" envDefault:"!change_me!"`
	MongoDBRequestTimeout time.Duration `env:"MONGODB_REQUEST_TIMEOUT" envDefault:"3m"`
	AccessTokenTimeout    time.Duration `env:"ACCESS_TOKEN_TIMEOUT" envDefault:"1h"`
	RefreshTokenTimeout   time.Duration `env:"REFRESH_TOKEN_TIMEOUT" envDefault:"2h"`
	PaginationMaxItem     int64         `env:"PAGINATION_MAX_ITEM" envDefault:"50"`
	APIBodyLimitSize      int           `env:"API_BODY_LIMIT_SIZE" envDefault:"1073741824"`
	Debug                 bool          `env:"DEBUG" envDefault:"true"`
	ElasticAPMEnable      bool          `env:"ELASTIC_APM_ENABLE" envDefault:"false"`
	MongoAutoIndexing     bool          `env:"MONGO_AUTO_INDEXING" envDefault:"true"`
}

func (cfg Configuration) ServerAddress() string {
	return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}

func GetConfig() Configuration {
	if config == nil {
		_ = godotenv.Load()
		config = &Configuration{}
		if err := env.Parse(config); err != nil {
			log.Fatal().Err(err).Msg("Get Config Error")
		}
	}
	return *config
}
