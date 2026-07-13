package server

import (
	"context"
	"errors"

	"fifriz/internal/ai"
	"fifriz/internal/documents"
)

type summaryChatAdapter struct {
	provider ai.ChatProvider
}

func NewSummaryChatAdapter(provider ai.ChatProvider) *summaryChatAdapter {
	return &summaryChatAdapter{provider: provider}
}

func (a *summaryChatAdapter) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if a == nil || a.provider == nil {
		return "", documents.ErrSummaryProviderDisabled
	}

	response, err := a.provider.Complete(ctx, ai.ChatCompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	})
	if err != nil {
		if errors.Is(err, ai.ErrOllamaChatUnavailable) || errors.Is(err, ai.ErrChatProviderDisabled) {
			return "", documents.ErrSummaryProviderUnavailable
		}
		return "", err
	}
	if response == nil {
		return "", documents.ErrSummaryProviderDisabled
	}
	return response.Content, nil
}
