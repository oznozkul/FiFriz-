package storage

import (
	"context"
	"io"
	"time"
)

const (
	ProviderLocal = "local"
	ProviderMinIO = "minio"
)

type UploadInput struct {
	Key         string
	Reader      io.Reader
	Size        int64
	ContentType string
}

type UploadResult struct {
	Key string
}

type StorageProvider interface {
	Name() string
	Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
