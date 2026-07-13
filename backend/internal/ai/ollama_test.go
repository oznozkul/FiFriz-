package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fifriz/internal/config"
)

func TestOllamaChatProviderComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var payload ollamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "llama3.1:8b" {
			t.Fatalf("model = %q, want llama3.1:8b", payload.Model)
		}
		if len(payload.Messages) != 2 || payload.Messages[0].Role != "system" || payload.Messages[1].Role != "user" {
			t.Fatalf("unexpected messages: %+v", payload.Messages)
		}

		_ = json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: ollamaMessage{Role: "assistant", Content: "Merhaba"},
		})
	}))
	defer server.Close()

	provider := NewOllamaChatProvider(OllamaConfig{
		BaseURL: server.URL,
		Model:   "llama3.1:8b",
	})

	response, err := provider.Complete(t.Context(), ChatCompletionRequest{
		SystemPrompt: "Sen yardimci bir asistansin.",
		UserPrompt:   "Selam",
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if response.Content != "Merhaba" {
		t.Fatalf("Content = %q, want Merhaba", response.Content)
	}
}

func TestOllamaChatProviderUnavailable(t *testing.T) {
	provider := NewOllamaChatProvider(OllamaConfig{
		BaseURL: "http://127.0.0.1:1",
		Model:   "llama3.1:8b",
	})

	_, err := provider.Complete(t.Context(), ChatCompletionRequest{
		UserPrompt: "Selam",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Ollama servisine baglanilamadi") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewChatProviderSelectsOllama(t *testing.T) {
	provider := NewChatProvider(configWithProvider("ollama"))
	if provider == nil {
		t.Fatal("expected ollama provider")
	}
	if provider.Name() != "ollama" {
		t.Fatalf("Name() = %q, want ollama", provider.Name())
	}
}

func configWithProvider(name string) config.AIConfig {
	return config.AIConfig{
		Provider: name,
		Ollama: config.OllamaChatConfig{
			BaseURL: "http://localhost:11434",
			Model:   "llama3.1:8b",
		},
	}
}
