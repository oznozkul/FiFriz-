package ai

import "testing"

func TestSessionTitleFromQuestion(t *testing.T) {
	title := sessionTitleFromQuestion("Bu sözleşmenin bitiş tarihi ne zaman?")
	if title != "Bu sözleşmenin bitiş tarihi ne zaman?" {
		t.Fatalf("unexpected title: %q", title)
	}

	longQuestion := ""
	for range 80 {
		longQuestion += "a"
	}

	title = sessionTitleFromQuestion(longQuestion)
	if len([]rune(title)) != sessionTitleMaxRunes {
		t.Fatalf("title length = %d, want %d", len([]rune(title)), sessionTitleMaxRunes)
	}
	if title[len(title)-3:] != "..." {
		t.Fatalf("title should end with ellipsis: %q", title)
	}
}

func TestDeriveSessionScopeDocument(t *testing.T) {
	documentID := "11111111-1111-1111-1111-111111111111"
	scope := deriveSessionScope(ChatRequest{DocumentID: &documentID})
	if scope.ScopeType == nil || *scope.ScopeType != ScopeTypeDocument {
		t.Fatalf("scope type = %v", scope.ScopeType)
	}
	if scope.ScopeID == nil {
		t.Fatal("scope id should be set")
	}
}
