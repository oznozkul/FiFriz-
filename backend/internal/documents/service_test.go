package documents

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"fifriz/internal/auth"
	"fifriz/internal/storage"

	"github.com/google/uuid"
)

func testStorageManager(t *testing.T) *storage.Manager {
	t.Helper()
	manager, err := storage.NewManager(storage.Config{
		Provider:      storage.ProviderLocal,
		LocalBasePath: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("storage manager: %v", err)
	}
	return manager
}

func TestDocumentStorageLocationUsesStoredFileName(t *testing.T) {
	storageDir := t.TempDir()
	storedFileName := uuid.New().String() + ".docx"
	filePath := filepath.Join(storageDir, storedFileName)

	if err := os.WriteFile(filePath, []byte("placeholder"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	manager, err := storage.NewManager(storage.Config{
		Provider:      storage.ProviderLocal,
		LocalBasePath: storageDir,
	})
	if err != nil {
		t.Fatalf("storage manager: %v", err)
	}

	document := &Document{
		StoredFileName:  storedFileName,
		StorageProvider: storage.ProviderLocal,
		StorageKey:      storedFileName,
		FilePath:        "legacy/relative/path.docx",
	}

	got, err := documentStorageLocation(manager, document)
	if err != nil {
		t.Fatalf("documentStorageLocation: %v", err)
	}
	want := filepath.Join(storageDir, storedFileName)
	if manager.LocalProvider().LocalPath(got.Key) != want {
		t.Fatalf("local path = %q, want %q", manager.LocalProvider().LocalPath(got.Key), want)
	}
}

func TestExtractDocumentTextFromStoredDOCXFile(t *testing.T) {
	storageDir := t.TempDir()
	storedFileName := uuid.New().String() + ".docx"
	filePath := filepath.Join(storageDir, storedFileName)

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("create docx: %v", err)
	}

	writer := zip.NewWriter(file)
	documentXML, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatalf("create document.xml: %v", err)
	}

	_, err = documentXML.Write([]byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Service integration test</w:t></w:r></w:p></w:body></w:document>`))
	if err != nil {
		t.Fatalf("write document.xml: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close docx: %v", err)
	}

	got, err := extractDocumentText("docx", filePath)
	if err != nil {
		t.Fatalf("extractDocumentText: %v", err)
	}

	if got != "Service integration test" {
		t.Fatalf("extractDocumentText() = %q, want %q", got, "Service integration test")
	}
}

func TestExtractDocumentTextFromStoredXLSXFile(t *testing.T) {
	storageDir := t.TempDir()
	storedFileName := uuid.New().String() + ".xlsx"
	xlsxPath := filepath.Join(storageDir, storedFileName)

	writeTestXLSX(t, xlsxPath, map[string]string{
		"xl/workbook.xml":            `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Data" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1"><v>Service integration test</v></c></row></sheetData></worksheet>`,
	})

	got, err := extractDocumentText("xlsx", xlsxPath)
	if err != nil {
		t.Fatalf("extractDocumentText: %v", err)
	}

	if got != "[Data]\nService integration test" {
		t.Fatalf("extractDocumentText() = %q, want %q", got, "[Data]\nService integration test")
	}
}

func TestSemanticSearchRejectsEmptyQuery(t *testing.T) {
	service := NewService(nil, testStorageManager(t), nil, nil, nil, nil, nil, nil)

	_, err := service.SemanticSearch(testAuthContext(), "   ")
	if !errors.Is(err, ErrInvalidSearchQuery) {
		t.Fatalf("SemanticSearch() error = %v, want ErrInvalidSearchQuery", err)
	}
}

func TestSemanticSearchRequiresEmbeddingProvider(t *testing.T) {
	service := NewService(nil, testStorageManager(t), nil, nil, nil, nil, nil, nil)

	_, err := service.SemanticSearch(testAuthContext(), "kira sözleşmesi")
	if !errors.Is(err, ErrEmbeddingProviderDisabled) {
		t.Fatalf("SemanticSearch() error = %v, want ErrEmbeddingProviderDisabled", err)
	}
}

func testAuthContext() context.Context {
	return auth.WithContext(context.Background(), auth.Context{
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		Email:    "test@example.com",
		Role:     auth.TenantRoleOwner,
	})
}
