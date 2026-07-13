package documents

import (
	"strings"
	"testing"
)

func TestNormalizeExtractedTextRepairsCommonTurkishReplacementForms(t *testing.T) {
	input := []byte("K \uFFFD raya Veren (Mal Sah \uFFFD b \uFFFD ) T.C. K \uFFFD ml \uFFFD k No: 123 m \uFFFD lk No Da \uFFFD re K \uFFFD racı Sözleşme")

	got := normalizeExtractedText(input)
	want := "Kiraya Veren (Mal Sahibi ) T.C. Kimlik No: 123 mülk No Daire Kiracı Sözleşme"

	if got != want {
		t.Fatalf("normalizeExtractedText() = %q, want %q", got, want)
	}
}

func TestNormalizeExtractedTextRepairsFragmentedTurkishReplacementForms(t *testing.T) {
	input := []byte("dava d \uFFFD lekçes \uFFFD nde Müvekk \uFFFD l \uFFFD n \uFFFD n 15.01.2025 tar \uFFFD h \uFFFD nde banka havales \uFFFD \uFFFD le borç verd \uFFFD ğ \uFFFD n \uFFFD b \uFFFD r yaz\uFFFDl\uFFFD belge \uFFFD braz edemem \uFFFD şt \uFFFD r B \uFFFD l \uFFFD rk \uFFFD ş \uFFFD raporu banka kay\uFFFDtlar\uFFFD ve d \uFFFD \uFFFD \uFFFD tal yazışmalar b \uFFFD rl \uFFFD kte değerlend \uFFFD r \uFFFD ld \uFFFD ğ \uFFFD nde aç\uFFFDk")

	got := normalizeExtractedText(input)

	wantParts := []string{
		"dava dilekçesinde",
		"Müvekkilinin",
		"tarihinde",
		"havalesi ile",
		"verdiğini",
		"bir yazılı belge ibraz edememiştir",
		"Bilirkişi raporu",
		"kayıtları",
		"dijital yazışmalar",
		"birlikte değerlend",
		"açık",
	}

	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("normalizeExtractedText() = %q, want it to contain %q", got, want)
		}
	}
}

func TestBuildPrefixSearchQueryNormalizesTurkishCaseAndDotlessI(t *testing.T) {
	got := buildPrefixSearchQuery("uğur batıkan söz")
	want := "ugur:* & batikan:* & soz:*"

	if got != want {
		t.Fatalf("buildPrefixSearchQuery() = %q, want %q", got, want)
	}
}

func TestNormalizeStoredExtensionHandlesDOCXVariants(t *testing.T) {
	tests := map[string]string{
		"docx":              "docx",
		".docx":             "docx",
		"DOCX":              "docx",
		".DOCX":             "docx",
		"example.DOCX":      "docx",
		"example.test.docx": "docx",
	}

	for input, want := range tests {
		if got := normalizeStoredExtension(input); got != want {
			t.Fatalf("normalizeStoredExtension(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildOriginalTextSnippetPreservesTurkishCharacters(t *testing.T) {
	fullText := "Giriş metni. Doküman Kodu VendorPaymentWebService_01 Bayi Ödeme Web Servisi Türkçe içerik."

	got := formatSearchSnippet(fullText, "dokuman", "")

	wantParts := []string{"Doküman", "Ödeme", "Türkçe"}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("formatSearchSnippet() = %q, want it to contain %q", got, want)
		}
	}
}
