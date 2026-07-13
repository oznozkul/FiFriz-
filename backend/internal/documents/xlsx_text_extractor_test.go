package documents

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractXLSXTextReadsAllSheetsWithNames(t *testing.T) {
	tempDir := t.TempDir()
	xlsxPath := filepath.Join(tempDir, "workbook.xlsx")
	writeTestXLSX(t, xlsxPath, map[string]string{
		"xl/workbook.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Satis" sheetId="1" r:id="rId1"/>
    <sheet name="Ozet" sheetId="2" r:id="rId2"/>
  </sheets>
</workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`,
		"xl/worksheets/sheet1.xml": `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
  <row r="1"><c r="A1"><v>Ürün</v></c><c r="B1"><v>Adet</v></c></row>
  <row r="2"><c r="A2"><v>Elma</v></c><c r="B2"><v>10</v></c></row>
</sheetData></worksheet>`,
		"xl/worksheets/sheet2.xml": `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>
  <row r="1"><c r="A1"><v>Toplam</v></c><c r="B1"><v>10</v></c></row>
</sheetData></worksheet>`,
	})

	got, err := extractXLSXText(xlsxPath)
	if err != nil {
		t.Fatalf("extractXLSXText: %v", err)
	}

	wantParts := []string{
		"[Satis]",
		"Ürün\tAdet",
		"Elma\t10",
		"[Ozet]",
		"Toplam\t10",
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("extractXLSXText() = %q, want it to contain %q", got, want)
		}
	}
}

func TestExtractXLSXTextUsesSharedStrings(t *testing.T) {
	tempDir := t.TempDir()
	xlsxPath := filepath.Join(tempDir, "shared-strings.xlsx")
	writeTestXLSX(t, xlsxPath, map[string]string{
		"xl/workbook.xml": `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Data" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/sharedStrings.xml": `<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><si><t>Vendor Payment</t></si><si><t>Ödeme Servisi</t></si></sst>`,
		"xl/worksheets/sheet1.xml": `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c></row></sheetData></worksheet>`,
	})

	got, err := extractXLSXText(xlsxPath)
	if err != nil {
		t.Fatalf("extractXLSXText: %v", err)
	}

	want := "[Data]\nVendor Payment\tÖdeme Servisi"
	if got != want {
		t.Fatalf("extractXLSXText() = %q, want %q", got, want)
	}
}

func TestExtractXLSXTextNormalizesZipEntryPaths(t *testing.T) {
	tempDir := t.TempDir()
	xlsxPath := filepath.Join(tempDir, "uppercase-path.xlsx")
	writeTestXLSX(t, xlsxPath, map[string]string{
		"XL\\Workbook.XML": `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Sayfa1" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml": `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1"><v>Uppercase Path Test</v></c></row></sheetData></worksheet>`,
	})

	got, err := extractXLSXText(xlsxPath)
	if err != nil {
		t.Fatalf("extractXLSXText: %v", err)
	}

	if got != "[Sayfa1]\nUppercase Path Test" {
		t.Fatalf("extractXLSXText() = %q, want %q", got, "[Sayfa1]\nUppercase Path Test")
	}
}

func TestNormalizeStoredExtensionHandlesXLSXVariants(t *testing.T) {
	tests := map[string]string{
		"xlsx":              "xlsx",
		".xlsx":             "xlsx",
		"XLSX":              "xlsx",
		".XLSX":             "xlsx",
		"example.XLSX":      "xlsx",
		"example.test.xlsx": "xlsx",
	}

	for input, want := range tests {
		if got := normalizeStoredExtension(input); got != want {
			t.Fatalf("normalizeStoredExtension(%q) = %q, want %q", input, got, want)
		}
	}
}

func writeTestXLSX(t *testing.T, xlsxPath string, parts map[string]string) {
	t.Helper()

	file, err := os.Create(xlsxPath)
	if err != nil {
		t.Fatalf("create xlsx: %v", err)
	}

	writer := zip.NewWriter(file)
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
		t.Fatalf("close xlsx: %v", err)
	}
}
