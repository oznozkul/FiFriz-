package documents

import (
	"strings"
	"testing"
)

func TestChunkTextReturnsSingleChunkForShortText(t *testing.T) {
	text := "Kısa bir belge metni. Tek paragrafta kalması yeterli."
	chunks := chunkText(text)

	if len(chunks) != 1 {
		t.Fatalf("chunkText() returned %d chunks, want 1", len(chunks))
	}
	if chunks[0] != text {
		t.Fatalf("chunkText() = %q, want %q", chunks[0], text)
	}
}

func TestChunkTextPreservesTurkishCharacters(t *testing.T) {
	text := strings.Repeat("Uğur BATIKAN mülk sözleşmesi için önemli hüküm. ", 200)
	chunks := chunkText(text)

	if len(chunks) < 2 {
		t.Fatalf("chunkText() returned %d chunks, want at least 2", len(chunks))
	}

	for index, chunk := range chunks {
		if !strings.Contains(chunk, "Uğur") || !strings.Contains(chunk, "sözleşmesi") {
			t.Fatalf("chunk %d lost Turkish characters: %q", index, chunk)
		}
	}
}

func TestChunkTextTargetsWordRange(t *testing.T) {
	sentence := "Bu cümle OCR ve chunking testi için kullanılıyor. "
	text := strings.Repeat(sentence, 500)
	chunks := chunkText(text)

	if len(chunks) < 2 {
		t.Fatalf("chunkText() returned %d chunks, want multiple chunks", len(chunks))
	}

	for index, chunk := range chunks {
		words := countWords(chunk)
		if index < len(chunks)-1 && words < chunkTargetMinWords/2 {
			t.Fatalf("chunk %d has only %d words", index, words)
		}
		if words > chunkTargetMaxWords+150 {
			t.Fatalf("chunk %d has %d words, expected around %d-%d", index, words, chunkTargetMinWords, chunkTargetMaxWords)
		}
	}
}

func TestChunkTextDoesNotReturnEmptyChunks(t *testing.T) {
	text := strings.Repeat("Paragraf satiri.\n", 800)
	chunks := chunkText(text)

	for index, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			t.Fatalf("chunk %d is empty", index)
		}
	}
}

func TestEstimateTokens(t *testing.T) {
	got := estimateTokens("bir iki uc dort bes")
	if got <= 0 {
		t.Fatalf("estimateTokens() = %d, want > 0", got)
	}
}

func TestChunkTextCreatesMultipleChunksForLargeDocuments(t *testing.T) {
	sentence := "Bu cümle büyük belge chunking doğrulaması için kullanılıyor. "
	text := strings.Repeat(sentence, 500)
	chunks := chunkText(text)

	if len(chunks) < 2 {
		t.Fatalf("chunkText() returned %d chunks, want at least 2 for large text", len(chunks))
	}

	for index, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			t.Fatalf("chunk %d is empty", index)
		}
	}
}
