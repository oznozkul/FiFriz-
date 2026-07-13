package ai

import "context"

type ProviderChatMessage struct {
	Role    string
	Content string
}

type ChatCompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
	Messages     []ProviderChatMessage
}

type ChatCompletionResponse struct {
	Content string
}

type ChatProvider interface {
	Name() string
	Complete(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error)
}
