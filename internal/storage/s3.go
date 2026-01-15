package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
)

// S3Storage S3兼容存储实现
// 支持AWS S3、阿里云OSS、腾讯云COS、MinIO等S3兼容存储
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage 创建S3存储实例
func NewS3Storage(cfg *config.S3Config) (Storage, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:           cfg.Endpoint,
				SigningRegion: cfg.Region,
			}, nil
		}
		return aws.Endpoint{}, nil
	})

	cfgAWS, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfgAWS)

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload 上传文件到S3存储
func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, metadata *Metadata) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(metadata.ContentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

// Download 从S3存储下载文件
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	return result.Body, nil
}

// DownloadRange 下载文件的指定范围
func (s *S3Storage) DownloadRange(ctx context.Context, key string, offset int64, length int64) (io.ReadCloser, error) {
	rangeHeader := fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Range:  aws.String(rangeHeader),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object range: %w", err)
	}
	return result.Body, nil
}

// Delete 从S3存储删除文件
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// GetURL 获取S3文件的访问URL
func (s *S3Storage) GetURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.client.Options().Region, key), nil
}

// Exists 检查S3文件是否存在
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetSize 获取文件大小
func (s *S3Storage) GetSize(ctx context.Context, key string) (int64, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get object size: %w", err)
	}
	if result.ContentLength == nil {
		return 0, nil
	}
	return *result.ContentLength, nil
}

// TestConnection 测试存储连接
func (s *S3Storage) TestConnection(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to S3 bucket: %w", err)
	}
	return nil
}

