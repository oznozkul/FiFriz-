package documents

import "testing"

func TestParseSummaryLLMResponseJSON(t *testing.T) {
	parsed, err := parseSummaryLLMResponse(`{"summary":"Kısa özet","keyPoints":["Madde 1","Madde 2"]}`)
	if err != nil {
		t.Fatalf("parseSummaryLLMResponse() error = %v", err)
	}
	if parsed.Summary != "Kısa özet" {
		t.Fatalf("summary = %q", parsed.Summary)
	}
	if len(parsed.KeyPoints) != 2 {
		t.Fatalf("key points len = %d", len(parsed.KeyPoints))
	}
}

func TestParseSummaryLLMResponseFencedJSON(t *testing.T) {
	parsed, err := parseSummaryLLMResponse("```json\n{\"summary\":\"Özet\",\"keyPoints\":[\"A\"]}\n```")
	if err != nil {
		t.Fatalf("parseSummaryLLMResponse() error = %v", err)
	}
	if parsed.Summary != "Özet" {
		t.Fatalf("summary = %q", parsed.Summary)
	}
}
