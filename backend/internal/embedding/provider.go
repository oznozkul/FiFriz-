package embedding

import (
	"context"
	"errors"
	"log"
	"strings"
)

var (
	ErrProviderDisabled = errors.New("embedding provider is disabled")
	ErrEmptyInput       = errors.New("embedding input is empty")
)

type Provider interface {
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)
	ProviderName() string
	ModelName() string
}

type Config struct {
	Provider string
	Ollama   OllamaConfig
}

func NewProvider(cfg Config) Provider {
	providerName := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if providerName == "" {
		providerName = "ollama"
	}

	switch providerName {
	case "ollama":
		provider := NewOllamaProvider(cfg.Ollama)
		baseURL := strings.TrimRight(strings.TrimSpace(cfg.Ollama.BaseURL), "/")
		if baseURL == "" {
			baseURL = defaultOllamaBaseURL
		}
		log.Printf(
			"embedding provider enabled: provider=%s base_url=%s model=%s",
			provider.ProviderName(),
			baseURL,
			provider.ModelName(),
		)
		return provider
	default:
		log.Printf("embedding provider disabled: unsupported provider %q", cfg.Provider)
		return nil
	}
}
