package storage_s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func (s *service) InitGlobal() {
	global = s
}

// ListBuckets lists the buckets in the current account.
func (s *service) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	var err error
	var output *s3.ListBucketsOutput
	var buckets []types.Bucket
	bucketPaginator := s3.NewListBucketsPaginator(s.Client, &s3.ListBucketsInput{})
	for bucketPaginator.HasMorePages() {
		output, err = bucketPaginator.NextPage(ctx)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "AccessDenied" {
				logger.Error().Err(err).Msg("You don't have permission to list buckets for this account.")
				err = apiErr
			} else {
				logger.Error().Err(err).Msg("Couldn't list buckets for your account.")
			}
			break
		} else {
			buckets = append(buckets, output.Buckets...)
		}
	}
	return buckets, err
}

// BucketExists checks whether a bucket exists in the current account.
func (s *service) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			var notFound *types.NotFound
			switch {
			case errors.As(apiError, &notFound):
				logger.Info().Msg(fmt.Sprintf("Bucket %v is available.\n", bucketName))
				exists = false
				err = nil
			default:
				logger.Error().Msg(fmt.Sprintf("Either you don't have access to bucket %v or another error occurred. "+
					"Here's what happened: %v\n", bucketName, err))
			}
		}
	} else {
		logger.Info().Msg(fmt.Sprintf("Bucket %v exists and you already own it.", bucketName))
	}

	return exists, err
}

// CreateBucket creates a bucket with the specified name in the specified Region.
func (s *service) CreateBucket(ctx context.Context, name string, region string) error {
	_, err := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	})
	if err != nil {
		var owned *types.BucketAlreadyOwnedByYou
		var exists *types.BucketAlreadyExists
		if errors.As(err, &owned) {
			logger.Info().Msg(fmt.Sprintf("You already own bucket %s.\n", name))
			err = owned
		} else if errors.As(err, &exists) {
			logger.Err(err).Msg(fmt.Sprintf("You already own bucket %s.\n", name))
			err = exists
		}
	} else {
		err = s3.NewBucketExistsWaiter(s.Client).Wait(
			ctx, &s3.HeadBucketInput{Bucket: aws.String(name)}, time.Minute)
		if err != nil {
			logger.Err(err).Msg(fmt.Sprintf("You already own bucket %s.\n", name))
		}
	}
	return err
}

// UploadFile reads from a file and puts the data into an object in a bucket.
func (s *service) UploadFile(ctx context.Context, bucketName string, objectKey string, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		logger.Err(err).Msg(fmt.Sprintf("Couldn't open file %v to upload.", fileName))
	} else {
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
		_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
			Body:   file,
		})
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
				logger.Error().Err(err).Msg(fmt.Sprintf("Error while uploading object to %s. The object is too large.\n"+
					"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
					"or the multipart upload API (5TB max).", bucketName))
			} else {
				logger.Error().Err(err).Msg(fmt.Sprintf("Error while uploading object to %v:%v.", bucketName, objectKey))
			}
		} else {
			err = s3.NewObjectExistsWaiter(s.Client).Wait(
				ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(objectKey)}, time.Minute)
			if err != nil {
				logger.Err(err).Msg(fmt.Sprintf("Failed attempt to wait for object %s to exist.\n", objectKey))
			}
		}
	}
	return err
}

// UploadLargeObject uses an upload manager to upload data to an object in a bucket.
// The upload manager breaks large data into parts and uploads the parts concurrently.
func (s *service) UploadLargeObject(ctx context.Context, bucketName string, objectKey string, largeObject []byte) error {
	largeBuffer := bytes.NewReader(largeObject)
	var partMiBs int64 = 10
	uploader := manager.NewUploader(s.Client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   largeBuffer,
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			logger.Error().Err(err).Msg(fmt.Sprintf("Error while uploading object to %s. The object is too large.\n"+
				"The maximum size for a multipart upload is 5TB.", bucketName))
		} else {
			logger.Error().Msg(fmt.Sprintf("Couldn't upload large object to %v:%v.", bucketName, objectKey))
		}
	} else {
		err = s3.NewObjectExistsWaiter(s.Client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(objectKey)}, time.Minute)
		if err != nil {
			logger.Err(err).Msg(fmt.Sprintf("Failed attempt to wait for object %s to exist.\n", objectKey))
		}
	}

	return err
}

// DownloadFile gets an object from a bucket and stores it in a local file.
func (s *service) DownloadFile(ctx context.Context, bucketName string, objectKey string, fileName string) error {
	result, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			logger.Error().Err(err).Msg(fmt.Sprintf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, bucketName))
			err = noKey
		} else {
			logger.Error().Err(err).Msg(fmt.Sprintf("Couldn't get object %v:%v.", bucketName, objectKey))
		}
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(result.Body)
	file, err := os.Create(fileName)
	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("Couldn't create file %v.", fileName))
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	body, err := io.ReadAll(result.Body)
	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("Couldn't read object body from %v.", objectKey))
	}
	_, err = file.Write(body)
	return err
}

// DownloadLargeObject uses a download manager to download an object from a bucket.
// The download manager gets the data in parts and writes them to a buffer until all of
// the data has been downloaded.
func (s *service) DownloadLargeObject(ctx context.Context, bucketName string, objectKey string) ([]byte, error) {
	var partMiBs int64 = 10
	downloader := manager.NewDownloader(s.Client, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})
	buffer := manager.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(ctx, buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("Couldn't download large object from %v:%v.", bucketName, objectKey))
	}
	return buffer.Bytes(), err
}

// CopyToFolder copies an object in a bucket to a subfolder in the same bucket.
func (s *service) CopyToFolder(ctx context.Context, bucketName string, objectKey string, folderName string) error {
	objectDest := fmt.Sprintf("%v/%v", folderName, objectKey)
	_, err := s.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucketName),
		CopySource: aws.String(fmt.Sprintf("%v/%v", bucketName, objectKey)),
		Key:        aws.String(objectDest),
	})
	if err != nil {
		var notActive *types.ObjectNotInActiveTierError
		if errors.As(err, &notActive) {
			logger.Error().Err(err).Msg(fmt.Sprintf("Couldn't copy object %s from %s because the object isn't in the active tier.\n",
				objectKey, bucketName))
			err = notActive
		}
	} else {
		err = s3.NewObjectExistsWaiter(s.Client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(objectDest)}, time.Minute)
		if err != nil {
			logger.Err(err).Msg(fmt.Sprintf("Failed attempt to wait for object %s to exist.\n", objectDest))
		}
	}
	return err
}

// CopyToBucket copies an object in a bucket to another bucket.
func (s *service) CopyToBucket(ctx context.Context, sourceBucket string, destinationBucket string, objectKey string) error {
	_, err := s.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destinationBucket),
		CopySource: aws.String(fmt.Sprintf("%v/%v", sourceBucket, objectKey)),
		Key:        aws.String(objectKey),
	})
	if err != nil {
		var notActive *types.ObjectNotInActiveTierError
		if errors.As(err, &notActive) {
			logger.Error().Err(err).Msg(fmt.Sprintf("Couldn't copy object %s from %s because the object isn't in the active tier.\n",
				objectKey, sourceBucket))
			err = notActive
		}
	} else {
		err = s3.NewObjectExistsWaiter(s.Client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(destinationBucket), Key: aws.String(objectKey)}, time.Minute)
		if err != nil {
			logger.Err(err).Msg(fmt.Sprintf("Failed attempt to wait for object %s to exist.\n", objectKey))
		}
	}
	return err
}

// ListObjects lists the objects in a bucket.
func (s *service) ListObjects(ctx context.Context, bucketName string) ([]types.Object, error) {
	var err error
	var output *s3.ListObjectsV2Output
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}
	var objects []types.Object
	objectPaginator := s3.NewListObjectsV2Paginator(s.Client, input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				logger.Error().Err(err).Msg("Bucket not found")
				err = noBucket
			}
			break
		} else {
			objects = append(objects, output.Contents...)
		}
	}
	return objects, err
}

// DeleteObjects deletes a list of objects from a bucket.
func (s *service) DeleteObjects(ctx context.Context, bucketName string, objectKeys []string) error {
	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	output, err := s.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{Objects: objectIds, Quiet: aws.Bool(true)},
	})
	if err != nil || len(output.Errors) > 0 {
		logger.Error().Err(err).Msg(fmt.Sprintf("Error deleting objects from bucket %s.\n", bucketName))
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				logger.Error().Err(err).Msg("Bucket not found")
				err = noBucket
			}
		} else if len(output.Errors) > 0 {
			for _, outErr := range output.Errors {
				logger.Info().Msg(fmt.Sprintf("%s: %s\n", *outErr.Key, *outErr.Message))
			}
			err = fmt.Errorf("%s", *output.Errors[0].Message)
		}
	} else {
		for _, delObjs := range output.Deleted {
			err = s3.NewObjectNotExistsWaiter(s.Client).Wait(
				ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: delObjs.Key}, time.Minute)
			if err != nil {
				logger.Error().Err(err).Msg(fmt.Sprintf("Failed attempt to wait for object %s to be deleted.\n", *delObjs.Key))
			} else {
				logger.Info().Msg(fmt.Sprintf("Deleted %s.\n", *delObjs.Key))
			}
		}
	}
	return err
}

// DeleteBucket deletes a bucket. The bucket must be empty or an error is returned.
func (s *service) DeleteBucket(ctx context.Context, bucketName string) error {
	_, err := s.Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName)})
	if err != nil {
		var noBucket *types.NoSuchBucket
		if errors.As(err, &noBucket) {
			logger.Error().Err(err).Msg("Bucket not found")
			err = noBucket
		} else {
			logger.Error().Err(err).Msg(fmt.Sprintf("Couldn't delete bucket %v.", bucketName))
		}
	} else {
		err = s3.NewBucketNotExistsWaiter(s.Client).Wait(
			ctx, &s3.HeadBucketInput{Bucket: aws.String(bucketName)}, time.Minute)
		if err != nil {
			logger.Error().Err(err).Msg(fmt.Sprintf("Failed attempt to wait for bucket %s to be deleted.\n", bucketName))
		} else {
			logger.Info().Msg(fmt.Sprintf("Deleted %s.\n", bucketName))
		}
	}
	return err
}
