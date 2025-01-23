package storage_s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"jira-clone-api/common/configure"
	"jira-clone-api/common/logging"
)

var (
	global Service
	logger = logging.GetLogger()
	cfg    = configure.GetConfig()
)

type Service interface {
	InitGlobal()
}

type service struct {
	Client *s3.Client
}

func New() Service {
	s3Config, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     cfg.AwsAccessKeyId,
				SecretAccessKey: cfg.AwsSecretAccessKey,
			},
		}))
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load s3 config")
	}
	client := s3.NewFromConfig(s3Config)
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String("jira"),
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("first page results")
	for _, object := range output.Contents {
		fmt.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
	}

	return &service{
		Client: client,
	}
}

func GetGlobal() Service {
	return global
}
