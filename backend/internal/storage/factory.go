package storage

import (
	"strings"

	"fifriz/internal/config"
)

func ManagerFromConfig(cfg config.StorageConfig) (*Manager, error) {
	return NewManager(Config{
		Provider:      cfg.Provider,
		LocalBasePath: cfg.LocalBasePath,
		ProcessingDir: cfg.ProcessingDir,
		MinIO: MinIOConfig{
			Endpoint:  cfg.MinIO.Endpoint,
			AccessKey: cfg.MinIO.AccessKey,
			SecretKey: cfg.MinIO.SecretKey,
			Bucket:    cfg.MinIO.Bucket,
			UseSSL:    cfg.MinIO.UseSSL,
			Region:    cfg.MinIO.Region,
		},
	})
}

func IsLocalDocument(provider string) bool {
	provider = strings.ToLower(strings.TrimSpace(provider))
	return provider == "" || provider == ProviderLocal
}
