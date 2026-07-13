package ai

import "testing"

func TestDeepSeekChatProviderName(t *testing.T) {
	provider := NewDeepSeekChatProvider(DeepSeekConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.deepseek.com",
		Model:   "deepseek-chat",
	})

	if provider.Name() != "deepseek" {
		t.Fatalf("Name() = %q, want deepseek", provider.Name())
	}

	if !provider.Configured() {
		t.Fatal("expected provider to be configured")
	}
}

func TestDeepSeekChatProviderCompleteNotImplemented(t *testing.T) {
	provider := NewDeepSeekChatProvider(DeepSeekConfig{APIKey: "test-key"})

	_, err := provider.Complete(t.Context(), ChatCompletionRequest{})
	if err != ErrChatNotImplemented {
		t.Fatalf("Complete() error = %v, want ErrChatNotImplemented", err)
	}
}
