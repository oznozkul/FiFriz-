package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalStorageProvider struct {
	basePath string
}

func NewLocalStorageProvider(basePath string) *LocalStorageProvider {
	return &LocalStorageProvider{basePath: filepath.Clean(basePath)}
}

func (p *LocalStorageProvider) Name() string {
	return ProviderLocal
}

func (p *LocalStorageProvider) resolvePath(key string) string {
	cleanKey := filepath.Clean(filepath.FromSlash(key))
	if filepath.IsAbs(cleanKey) {
		return cleanKey
	}
	return filepath.Join(p.basePath, cleanKey)
}

func (p *LocalStorageProvider) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	if input.Key == "" {
		return nil, errors.New("storage key is required")
	}
	if input.Reader == nil {
		return nil, errors.New("upload reader is required")
	}

	destination := p.resolvePath(input.Key)
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := io.Copy(file, input.Reader); err != nil {
		_ = os.Remove(destination)
		return nil, err
	}

	_ = ctx.Err()
	return &UploadResult{Key: input.Key}, nil
}

func (p *LocalStorageProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.Open(p.resolvePath(key))
}

func (p *LocalStorageProvider) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := os.Remove(p.resolvePath(key))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (p *LocalStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	_, err := os.Stat(p.resolvePath(key))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (p *LocalStorageProvider) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	_ = ctx
	_ = key
	_ = expiry
	return "", errors.New("presigned URLs are not supported for local storage")
}

func (p *LocalStorageProvider) BasePath() string {
	return p.basePath
}

func (p *LocalStorageProvider) LocalPath(key string) string {
	return p.resolvePath(key)
}
