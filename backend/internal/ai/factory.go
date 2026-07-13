package ai

import (
	"log"
	"strings"

	"fifriz/internal/config"
)

func NewChatProvider(cfg config.AIConfig) ChatProvider {
	providerName := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if providerName == "" {
		providerName = "ollama"
	}

	switch providerName {
	case "ollama":
		provider := NewOllamaChatProvider(OllamaConfig{
			BaseURL: cfg.Ollama.BaseURL,
			Model:   cfg.Ollama.Model,
			Timeout: cfg.Ollama.Timeout,
		})
		log.Printf(
			"ai chat provider enabled: provider=%s model=%s base_url=%s",
			provider.Name(),
			provider.Model(),
			cfg.Ollama.BaseURL,
		)
		return provider
	case "deepseek":
		provider := NewDeepSeekChatProvider(DeepSeekConfig{
			APIKey:  cfg.DeepSeek.APIKey,
			BaseURL: cfg.DeepSeek.BaseURL,
			Model:   cfg.DeepSeek.Model,
		})
		if !provider.Configured() {
			log.Printf("ai chat provider disabled: deepseek api key is missing")
			return nil
		}
		log.Printf("ai chat provider enabled: provider=%s model=%s", provider.Name(), cfg.DeepSeek.Model)
		return provider
	case "none", "disabled":
		log.Printf("ai chat provider disabled: configured_provider=%q", cfg.Provider)
		return nil
	default:
		log.Printf("ai chat provider disabled: unsupported provider %q", cfg.Provider)
		return nil
	}
}

func NewDeepSeekProviderFromConfig(cfg config.DeepSeekConfig) *DeepSeekChatProvider {
	return NewDeepSeekChatProvider(DeepSeekConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
}

func NewOllamaProviderFromConfig(cfg config.OllamaChatConfig) *OllamaChatProvider {
	return NewOllamaChatProvider(OllamaConfig{
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
		Timeout: cfg.Timeout,
	})
}
