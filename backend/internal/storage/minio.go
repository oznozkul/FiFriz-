package storage

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
}

type MinIOStorageProvider struct {
	client *minio.Client
	bucket string
	region string
}

func NewMinIOStorageProvider(cfg MinIOConfig) (*MinIOStorageProvider, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("minio endpoint is required")
	}
	if cfg.Bucket == "" {
		return nil, errors.New("minio bucket is required")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	provider := &MinIOStorageProvider{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}
	if err := provider.EnsureBucket(context.Background()); err != nil {
		return nil, err
	}
	return provider, nil
}

func (p *MinIOStorageProvider) Name() string {
	return ProviderMinIO
}

func (p *MinIOStorageProvider) Bucket() string {
	return p.bucket
}

func (p *MinIOStorageProvider) EnsureBucket(ctx context.Context) error {
	exists, err := p.client.BucketExists(ctx, p.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return p.client.MakeBucket(ctx, p.bucket, minio.MakeBucketOptions{Region: p.region})
}

func (p *MinIOStorageProvider) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	if input.Key == "" {
		return nil, errors.New("storage key is required")
	}
	if input.Reader == nil {
		return nil, errors.New("upload reader is required")
	}

	_, err := p.client.PutObject(ctx, p.bucket, input.Key, input.Reader, input.Size, minio.PutObjectOptions{
		ContentType: input.ContentType,
	})
	if err != nil {
		return nil, err
	}
	return &UploadResult{Key: input.Key}, nil
}

func (p *MinIOStorageProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := p.client.GetObject(ctx, p.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	if _, err := object.Stat(); err != nil {
		_ = object.Close()
		return nil, err
	}
	return object, nil
}

func (p *MinIOStorageProvider) Delete(ctx context.Context, key string) error {
	return p.client.RemoveObject(ctx, p.bucket, key, minio.RemoveObjectOptions{})
}

func (p *MinIOStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.StatObject(ctx, p.bucket, key, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	if minio.ToErrorResponse(err).Code == "NoSuchKey" {
		return false, nil
	}
	return false, err
}

func (p *MinIOStorageProvider) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := p.client.PresignedGetObject(ctx, p.bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
