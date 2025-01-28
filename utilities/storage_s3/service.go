package storage_s3

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
)

func (s *service) InitGlobal() {
	global = s
}

func (s *service) UploadObject(file *multipart.FileHeader) error {
	f, err := file.Open()
	if err != nil {
		return err
	}
	fileContents, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	_, err = s.Client.PutObject(context.Background(), s.Bucket, file.Filename, bytes.NewReader(fileContents), file.Size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	return err
}

func (s *service) CheckBucketExists(bucketName string) (bool, error) {
	return s.Client.BucketExists(context.Background(), bucketName)
}
