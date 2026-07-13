package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOllamaProviderCreateEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/embed" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		var payload ollamaEmbedRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "bge-m3" {
			t.Fatalf("model = %q, want bge-m3", payload.Model)
		}
		if payload.Input == "" {
			t.Fatal("expected input in request")
		}

		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(ollamaEmbedResponse{
			Model: "bge-m3",
			Embeddings: [][]float32{
				makeEmbeddingVector(1024, 0.42),
			},
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
		Model:   "bge-m3",
	})

	embedding, err := provider.CreateEmbedding(context.Background(), "Uğur BATIKAN sözleşme metni")
	if err != nil {
		t.Fatalf("CreateEmbedding: %v", err)
	}

	if len(embedding) != 1024 {
		t.Fatalf("embedding length = %d, want 1024", len(embedding))
	}
}

func TestOllamaProviderCreateEmbeddingLegacyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(ollamaEmbedResponse{
			Embedding: makeEmbeddingVector(1024, 0.11),
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
		Model:   "bge-m3",
	})

	embedding, err := provider.CreateEmbedding(context.Background(), "legacy response")
	if err != nil {
		t.Fatalf("CreateEmbedding: %v", err)
	}

	if len(embedding) != 1024 {
		t.Fatalf("embedding length = %d, want 1024", len(embedding))
	}
}

func TestOllamaProviderCreateEmbeddingModelError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(writer).Encode(ollamaErrorResponse{
			Error: `model "bge-m3" not found, try pulling it first`,
		})
	}))
	defer server.Close()

	provider := NewOllamaProvider(OllamaConfig{
		BaseURL: server.URL,
		Model:   "bge-m3",
	})

	_, err := provider.CreateEmbedding(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for missing model")
	}
	if got := err.Error(); !strings.Contains(got, "bge-m3") || !strings.Contains(got, "not found") {
		t.Fatalf("error = %q, want model not found details", got)
	}
}

func makeEmbeddingVector(size int, value float32) []float32 {
	vector := make([]float32, size)
	for index := range vector {
		vector[index] = value
	}
	return vector
}

func TestNewProviderDefaultsToOllama(t *testing.T) {
	provider := NewProvider(Config{})
	if provider == nil {
		t.Fatal("expected ollama provider by default")
	}

	if _, ok := provider.(*OllamaProvider); !ok {
		t.Fatalf("expected *OllamaProvider, got %T", provider)
	}

	if provider.ProviderName() != "ollama" {
		t.Fatalf("provider name = %q, want ollama", provider.ProviderName())
	}

	if provider.ModelName() != defaultOllamaEmbeddingModel {
		t.Fatalf("model name = %q, want %s", provider.ModelName(), defaultOllamaEmbeddingModel)
	}
}
