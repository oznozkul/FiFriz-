package documents

import (
	"errors"
	"os/exec"
	"testing"
)

func TestIsImageExtension(t *testing.T) {
	tests := map[string]bool{
		"png":         true,
		".png":        true,
		"PNG":         true,
		"jpg":         true,
		"jpeg":        true,
		".JPEG":       true,
		"pdf":         false,
		"docx":        false,
		"xlsx":        false,
		"example.png": true,
	}

	for input, want := range tests {
		if got := isImageExtension(input); got != want {
			t.Fatalf("isImageExtension(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestResolveTesseractCommandUsesEnvOverride(t *testing.T) {
	t.Setenv("TESSERACT_CMD", "tesseract")

	path, err := resolveTesseractCommand()
	if err == ErrTesseractNotAvailable {
		t.Skip("tesseract is not installed in this environment")
	}
	if err != nil {
		t.Fatalf("resolveTesseractCommand: %v", err)
	}
	if path == "" {
		t.Fatal("resolveTesseractCommand returned empty path")
	}
}

func TestExtractImageTextReturnsUnavailableWhenTesseractMissing(t *testing.T) {
	if _, err := exec.LookPath("tesseract"); err == nil {
		t.Skip("tesseract is available on PATH")
	}

	t.Setenv("TESSERACT_CMD", "")

	_, err := extractImageText("missing-image.png")
	if !errors.Is(err, ErrTesseractNotAvailable) {
		t.Fatalf("extractImageText() error = %v, want ErrTesseractNotAvailable", err)
	}
}
