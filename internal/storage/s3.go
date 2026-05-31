package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"edu-license/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Uploader interface {
	Upload(ctx context.Context, key, contentType string, body io.Reader) error
	Configured() bool
}

type S3Uploader struct {
	client *s3.Client
	bucket string
}

func NewS3Uploader(ctx context.Context, cfg config.Config) (*S3Uploader, error) {
	if cfg.S3Bucket == "" {
		return &S3Uploader{}, nil
	}
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKeyID,
			cfg.S3SecretAccessKey,
			"",
		)),
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true
		}
	})
	return &S3Uploader{client: client, bucket: cfg.S3Bucket}, nil
}

func (u *S3Uploader) Configured() bool {
	return u != nil && u.client != nil && u.bucket != ""
}

func (u *S3Uploader) Upload(ctx context.Context, key, contentType string, body io.Reader) error {
	if !u.Configured() {
		return fmt.Errorf("S3/R2 storage is not configured")
	}
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

func ApplicationDocumentKey(applicationID, filename string) string {
	return path.Join("applications", applicationID, fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename))
}
