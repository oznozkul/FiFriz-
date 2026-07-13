package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOllamaBaseURL        = "http://localhost:11434"
	defaultOllamaEmbeddingModel = "bge-m3"
)

type OllamaConfig struct {
	BaseURL string
	Model   string
}

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaProvider(cfg OllamaConfig) *OllamaProvider {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultOllamaBaseURL
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultOllamaEmbeddingModel
	}

	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *OllamaProvider) ProviderName() string {
	return "ollama"
}

func (p *OllamaProvider) ModelName() string {
	return p.model
}

type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ollamaEmbedResponse struct {
	Model      string        `json:"model"`
	Embeddings [][]float32   `json:"embeddings"`
	Embedding  []float32     `json:"embedding"`
}

type ollamaErrorResponse struct {
	Error string `json:"error"`
}

func (p *OllamaProvider) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, ErrEmptyInput
	}

	payload, err := json.Marshal(ollamaEmbedRequest{
		Model: p.model,
		Input: text,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("ollama embed request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		message := parseOllamaErrorMessage(body)
		if message == "" {
			message = strings.TrimSpace(string(body))
		}
		return nil, fmt.Errorf("ollama embed request failed: status=%d model=%s message=%s", response.StatusCode, p.model, message)
	}

	var parsed ollamaEmbedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("ollama embed response parse failed: %w body=%s", err, truncateForLog(string(body), 256))
	}

	embeddingValues := firstEmbeddingVector(parsed)
	if len(embeddingValues) == 0 {
		if message := parseOllamaErrorMessage(body); message != "" {
			return nil, fmt.Errorf("ollama embed failed: model=%s message=%s", p.model, message)
		}
		return nil, fmt.Errorf("ollama embed response did not contain embedding data model=%s body=%s", p.model, truncateForLog(string(body), 256))
	}

	return embeddingValues, nil
}

func firstEmbeddingVector(parsed ollamaEmbedResponse) []float32 {
	if len(parsed.Embeddings) > 0 && len(parsed.Embeddings[0]) > 0 {
		return parsed.Embeddings[0]
	}
	return parsed.Embedding
}

func parseOllamaErrorMessage(body []byte) string {
	var parsed ollamaErrorResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Error)
}

func truncateForLog(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}
