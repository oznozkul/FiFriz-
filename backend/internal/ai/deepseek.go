package ai

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

var ErrChatNotImplemented = errors.New("deepseek chat completion is not implemented yet")

type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type DeepSeekChatProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewDeepSeekChatProvider(cfg DeepSeekConfig) *DeepSeekChatProvider {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "deepseek-chat"
	}

	return &DeepSeekChatProvider{
		apiKey:  strings.TrimSpace(cfg.APIKey),
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *DeepSeekChatProvider) Name() string {
	return "deepseek"
}

func (p *DeepSeekChatProvider) Complete(_ context.Context, _ ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return nil, ErrChatNotImplemented
}

func (p *DeepSeekChatProvider) Configured() bool {
	return p.apiKey != ""
}
