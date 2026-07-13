package documents

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDOCXTextPreservesTurkishCharactersAndParagraphs(t *testing.T) {
	tempDir := t.TempDir()
	docxPath := filepath.Join(tempDir, "test.docx")

	file, err := os.Create(docxPath)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}

	writer := zip.NewWriter(file)
	documentXML, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create document.xml: %v", err)
	}

	_, err = documentXML.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Uğur BATIKAN mülk sözleşme</w:t></w:r></w:p>
    <w:p><w:r><w:t>İkinci paragraf: Kiracı Daire No</w:t></w:r></w:p>
  </w:body>
</w:document>`))
	if err != nil {
		t.Fatalf("write document.xml: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close docx: %v", err)
	}

	got, err := extractDOCXText(docxPath)
	if err != nil {
		t.Fatalf("extractDOCXText: %v", err)
	}

	want := "Uğur BATIKAN mülk sözleşme\nİkinci paragraf: Kiracı Daire No"
	if got != want {
		t.Fatalf("extractDOCXText() = %q, want %q", got, want)
	}
}

func TestExtractDOCXTextNormalizesZipEntryPaths(t *testing.T) {
	tempDir := t.TempDir()
	docxPath := filepath.Join(tempDir, "uppercase-path.docx")

	file, err := os.Create(docxPath)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}

	writer := zip.NewWriter(file)
	documentXML, err := writer.Create("Word\\Document.XML")
	if err != nil {
		t.Fatalf("create document.xml: %v", err)
	}

	_, err = documentXML.Write([]byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Uppercase Path Test</w:t></w:r></w:p></w:body></w:document>`))
	if err != nil {
		t.Fatalf("write document.xml: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close docx: %v", err)
	}

	got, err := extractDOCXText(docxPath)
	if err != nil {
		t.Fatalf("extractDOCXText: %v", err)
	}

	if got != "Uppercase Path Test" {
		t.Fatalf("extractDOCXText() = %q, want %q", got, "Uppercase Path Test")
	}
}

func TestExtractDOCXTextReadsHeadersAndFooters(t *testing.T) {
	tempDir := t.TempDir()
	docxPath := filepath.Join(tempDir, "header-footer.docx")

	file, err := os.Create(docxPath)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}

	writer := zip.NewWriter(file)
	parts := map[string]string{
		"word/document.xml": `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Vendor Payment Service</w:t></w:r></w:p></w:body></w:document>`,
		"word/header1.xml":  `<w:hdr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:p><w:r><w:t>Türkçe Başlık</w:t></w:r></w:p></w:hdr>`,
		"word/footer1.xml":  `<w:ftr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:p><w:r><w:t>Sayfa Altı</w:t></w:r></w:p></w:ftr>`,
	}

	for name, content := range parts {
		part, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		if _, err := part.Write([]byte(content)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close docx: %v", err)
	}

	got, err := extractDOCXText(docxPath)
	if err != nil {
		t.Fatalf("extractDOCXText: %v", err)
	}

	wantParts := []string{"Vendor Payment Service", "Türkçe Başlık", "Sayfa Altı"}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("extractDOCXText() = %q, want it to contain %q", got, want)
		}
	}
}
