package documents

import (
	"strings"
	"unicode"
)

const (
	chunkTargetMinWords      = 600
	chunkTargetMaxWords      = 900
	chunkMergeRemainderBelow = 200
)

func chunkText(fullText string) []string {
	fullText = strings.TrimSpace(fullText)
	if fullText == "" {
		return nil
	}

	totalWords := countWords(fullText)
	if totalWords <= chunkTargetMaxWords {
		return []string{fullText}
	}

	sentences := splitSentences(fullText)
	if len(sentences) == 0 {
		return []string{fullText}
	}

	chunks := make([]string, 0)
	var builder strings.Builder
	currentWords := 0

	flush := func() {
		content := strings.TrimSpace(builder.String())
		builder.Reset()
		currentWords = 0
		if content != "" {
			chunks = append(chunks, content)
		}
	}

	for _, sentence := range sentences {
		sentenceWords := countWords(sentence)
		if sentenceWords == 0 {
			continue
		}

		if currentWords > 0 && currentWords+sentenceWords > chunkTargetMaxWords && currentWords >= chunkTargetMinWords {
			flush()
		}

		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(sentence)
		currentWords += sentenceWords
	}

	flush()
	chunks = mergeTrailingShortChunk(chunks)

	if len(chunks) == 0 {
		return []string{fullText}
	}

	return chunks
}

func mergeTrailingShortChunk(chunks []string) []string {
	if len(chunks) < 2 {
		return chunks
	}

	lastIndex := len(chunks) - 1
	if countWords(chunks[lastIndex]) >= chunkMergeRemainderBelow {
		return chunks
	}

	merged := strings.TrimSpace(chunks[lastIndex-1] + " " + chunks[lastIndex])
	chunks[lastIndex-1] = merged
	return chunks[:lastIndex]
}

func splitSentences(text string) []string {
	sentences := make([]string, 0)
	paragraphs := strings.Split(text, "\n")

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		var builder strings.Builder
		for _, currentRune := range paragraph {
			builder.WriteRune(currentRune)
			if isSentenceTerminator(currentRune) {
				sentence := strings.TrimSpace(builder.String())
				if sentence != "" {
					sentences = append(sentences, sentence)
				}
				builder.Reset()
			}
		}

		remainder := strings.TrimSpace(builder.String())
		if remainder != "" {
			sentences = append(sentences, remainder)
		}
	}

	return sentences
}

func isSentenceTerminator(value rune) bool {
	switch value {
	case '.', '!', '?', '…', ';':
		return true
	default:
		return false
	}
}

func countWords(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	count := 0
	inWord := false
	for _, currentRune := range value {
		if unicode.IsLetter(currentRune) || unicode.IsNumber(currentRune) {
			if !inWord {
				count++
				inWord = true
			}
			continue
		}
		inWord = false
	}

	return count
}

func estimateTokens(content string) int {
	words := countWords(content)
	if words == 0 {
		return 0
	}

	return words + words/3
}
