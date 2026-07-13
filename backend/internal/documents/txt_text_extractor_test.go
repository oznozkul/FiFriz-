package documents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTXTText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("IMAP imported body"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := extractTXTText(path)
	if err != nil {
		t.Fatalf("extractTXTText: %v", err)
	}
	if got != "IMAP imported body" {
		t.Fatalf("extractTXTText() = %q, want %q", got, "IMAP imported body")
	}
}
