package storage_s3

import (
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	UploadObject(file *multipart.FileHeader) error
}

type service struct {
	Client *minio.Client
	Bucket string
}

func New() Service {
	minioClient, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.S3AccessKeyId, cfg.S3SecretAccessKey, ""),
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Storage S3 init error")
	}
	return &service{
		Client: minioClient,
		Bucket: cfg.S3BucketName,
	}
}

func GetGlobal() Service {
	return global
}
