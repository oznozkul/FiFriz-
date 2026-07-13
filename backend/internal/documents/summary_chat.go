package documents

import "context"

// SummaryChatCompleter generates text for document summaries.
// Implemented by the AI chat provider adapter in server wiring.
type SummaryChatCompleter interface {
	Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}
