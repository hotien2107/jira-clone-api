package storage_s3

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
)

func (s *service) InitGlobal() {
	global = s
}

func (s *service) UploadObject(bucketName, objectName string, file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	_, err = s.Client.PutObject(context.Background(), bucketName, objectName, file, info.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	return err
}

func (s *service) CheckBucketExists(bucketName string) (bool, error) {
	return s.Client.BucketExists(context.Background(), bucketName)
}
