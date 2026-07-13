package integrations

import "testing"

func TestDecodeMIMEHeaderValue(t *testing.T) {
	t.Parallel()

	subject := "=?UTF-8?Q?=C3=96deme_Entegrasyonu?="
	decoded := decodeMIMEHeaderValue(subject)
	if decoded != "Ödeme Entegrasyonu" {
		t.Fatalf("expected decoded subject, got %q", decoded)
	}
}

func TestSanitizeIMAPFilenamePreservesUnicode(t *testing.T) {
	t.Parallel()

	filename := sanitizeIMAPFilename("YouTrack: Kullanıcı Grubuna Göre İndirimi")
	if filename == "" {
		t.Fatal("expected non-empty filename")
	}
	if filename == "YouTrack__Kullan_C4_B1c_C4_B1" {
		t.Fatalf("filename should be decoded, got %q", filename)
	}
}
