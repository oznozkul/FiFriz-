package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	defaultOllamaChatBaseURL = "http://localhost:11434"
	defaultOllamaChatModel   = "llama3.1:8b"
)

var (
	ErrOllamaChatUnavailable = errors.New("ollama chat service is unavailable")
	ErrEmptyChatInput        = errors.New("chat request is empty")
)

type OllamaConfig struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

type OllamaChatProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaChatProvider(cfg OllamaConfig) *OllamaChatProvider {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultOllamaChatBaseURL
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultOllamaChatModel
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	return &OllamaChatProvider{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *OllamaChatProvider) Name() string {
	return "ollama"
}

func (p *OllamaChatProvider) Model() string {
	return p.model
}

func (p *OllamaChatProvider) Complete(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error) {
	messages := normalizeChatMessages(request)
	if len(messages) == 0 {
		return nil, ErrEmptyChatInput
	}

	payload, err := json.Marshal(ollamaChatRequest{
		Model:    p.model,
		Messages: toOllamaMessages(messages),
		Stream:   false,
	})
	if err != nil {
		return nil, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(httpRequest)
	if err != nil {
		return nil, userFriendlyOllamaChatError(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama chat response read failed: %w", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		message := parseOllamaChatErrorMessage(body)
		if message == "" {
			message = strings.TrimSpace(string(body))
		}
		return nil, fmt.Errorf("%w: %s", ErrOllamaChatUnavailable, message)
	}

	var parsed ollamaChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("ollama chat response parse failed: %w", err)
	}

	content := strings.TrimSpace(parsed.Message.Content)
	if content == "" {
		if message := parseOllamaChatErrorMessage(body); message != "" {
			return nil, fmt.Errorf("%w: %s", ErrOllamaChatUnavailable, message)
		}
		return nil, fmt.Errorf("%w: empty response from model %s", ErrOllamaChatUnavailable, p.model)
	}

	return &ChatCompletionResponse{Content: content}, nil
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
}

type ollamaChatErrorResponse struct {
	Error string `json:"error"`
}

func normalizeChatMessages(request ChatCompletionRequest) []ProviderChatMessage {
	if len(request.Messages) > 0 {
		return request.Messages
	}

	messages := make([]ProviderChatMessage, 0, 2)
	if systemPrompt := strings.TrimSpace(request.SystemPrompt); systemPrompt != "" {
		messages = append(messages, ProviderChatMessage{Role: "system", Content: systemPrompt})
	}
	if userPrompt := strings.TrimSpace(request.UserPrompt); userPrompt != "" {
		messages = append(messages, ProviderChatMessage{Role: "user", Content: userPrompt})
	}
	return messages
}

func toOllamaMessages(messages []ProviderChatMessage) []ollamaMessage {
	result := make([]ollamaMessage, 0, len(messages))
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		content := strings.TrimSpace(message.Content)
		if role == "" || content == "" {
			continue
		}
		result = append(result, ollamaMessage{Role: role, Content: content})
	}
	return result
}

func parseOllamaChatErrorMessage(body []byte) string {
	var parsed ollamaChatErrorResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Error)
}

func userFriendlyOllamaChatError(err error) error {
	if err == nil {
		return nil
	}

	message := strings.ToLower(err.Error())
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled),
		strings.Contains(message, "timeout"),
		errors.Is(err, net.ErrClosed):
		return fmt.Errorf("%w: istek zaman asimina ugradi", ErrOllamaChatUnavailable)
	case strings.Contains(message, "connection refused"),
		strings.Contains(message, "no such host"),
		strings.Contains(message, "dial tcp"):
		return fmt.Errorf("%w: Ollama servisine baglanilamadi", ErrOllamaChatUnavailable)
	default:
		return fmt.Errorf("%w: %v", ErrOllamaChatUnavailable, err)
	}
}
