package ai

import (
	"strings"
	"testing"

	"fifriz/internal/documents"

	"github.com/google/uuid"
)

func TestBuildRAGUserPromptIncludesChunks(t *testing.T) {
	documentID := uuid.New()
	chunks := []documents.SemanticSearchResult{
		{
			DocumentID:    documentID,
			DocumentTitle: "Kira Sözleşmesi.pdf",
			ChunkIndex:    2,
			Content:       "Aylık kira bedeli 15.000 TL'dir.",
		},
	}

	prompt := buildRAGUserPrompt("Kira bedeli ne kadar?", chunks)
	if !strings.Contains(prompt, "Kira Sözleşmesi.pdf") {
		t.Fatalf("prompt missing document title: %s", prompt)
	}
	if !strings.Contains(prompt, "15.000 TL") {
		t.Fatalf("prompt missing chunk content: %s", prompt)
	}
	if !strings.Contains(prompt, "Kira bedeli ne kadar?") {
		t.Fatalf("prompt missing question: %s", prompt)
	}
}

func TestPreviewContentTruncatesLongText(t *testing.T) {
	longText := strings.Repeat("a", contentPreviewMaxRunes+20)
	preview := previewContent(longText)
	if len([]rune(preview)) != contentPreviewMaxRunes+3 {
		t.Fatalf("preview length = %d, want %d", len([]rune(preview)), contentPreviewMaxRunes+3)
	}
	if !strings.HasSuffix(preview, "...") {
		t.Fatalf("preview should end with ellipsis: %q", preview)
	}
}

func TestToChatSourcesUsesDocumentTitle(t *testing.T) {
	documentID := uuid.New()
	chunkID := uuid.New()
	sources := toChatSources([]documents.SemanticSearchResult{
		{
			DocumentID:    documentID,
			DocumentTitle: "Fatura.pdf",
			ChunkID:       chunkID,
			ChunkIndex:    0,
			Similarity:    0.91,
			Content:       "Toplam tutar 250 TL",
		},
	})

	if len(sources) != 1 {
		t.Fatalf("sources len = %d, want 1", len(sources))
	}
	if sources[0].DocumentTitle != "Fatura.pdf" {
		t.Fatalf("DocumentTitle = %q", sources[0].DocumentTitle)
	}
	if sources[0].ContentPreview == "" {
		t.Fatal("ContentPreview should not be empty")
	}
}
