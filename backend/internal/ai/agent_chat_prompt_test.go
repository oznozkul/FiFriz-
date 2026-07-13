package ai

import "testing"

func TestParseAgentLLMResponse(t *testing.T) {
	raw := `{"answer":"Merhaba","actions":[{"toolName":"add_tag_to_document","title":"Etiket ekle","description":"test","input":{"documentId":"abc","tagName":"sözleşme"}}]}`
	parsed, err := parseAgentLLMResponse(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Answer != "Merhaba" {
		t.Fatalf("answer = %q", parsed.Answer)
	}
	if len(parsed.Actions) != 1 || parsed.Actions[0].ToolName != "add_tag_to_document" {
		t.Fatalf("actions = %+v", parsed.Actions)
	}
}

func TestParseAgentLLMResponseMarkdown(t *testing.T) {
	raw := "```json\n{\"answer\":\"Tamam\",\"actions\":[]}\n```"
	parsed, err := parseAgentLLMResponse(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Answer != "Tamam" {
		t.Fatalf("answer = %q", parsed.Answer)
	}
}
