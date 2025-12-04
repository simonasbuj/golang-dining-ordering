// Package s3 handles s3 storage of menu item images.
package s3

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type s3Storage struct {
	s3Client *s3.Client
	url      string
	bucket   string
}

// NewS3Storage initializes an S3/MinIO client for the specified bucket.
//
//revive:disable:unexported-return
func NewS3Storage(ctx context.Context, key, secret, url, bucket string) *s3Storage {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("loading s3 default config")
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = "eu-west-3"
		o.Credentials = aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(key, secret, ""),
		)
		o.BaseEndpoint = aws.String(url)
		o.UsePathStyle = true
	})

	return &s3Storage{
		s3Client: client,
		url:      url,
		bucket:   bucket,
	}
}

//revive:enable:unexported-return

func (s *s3Storage) StoreMenuItemImage(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close() //nolint:errcheck

	fileName := fmt.Sprintf("%s/%s", uuid.New().String(), fileHeader.Filename)

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fileName),
		Body:        file,
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", s.url, s.bucket, fileName), nil
}

func (s *s3Storage) DeleteMenuItemImage(ctx context.Context, path string) error {
	return nil
}
