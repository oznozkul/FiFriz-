package integrations

import (
	"mime"
	"regexp"
	"strings"
)

var unsafeIMAPFilenamePattern = regexp.MustCompile(`[^\p{L}\p{N}._\-\s]+`)

func decodeMIMEHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	decoded, err := (&mime.WordDecoder{}).DecodeHeader(value)
	if err != nil {
		return value
	}
	return strings.TrimSpace(decoded)
}

func sanitizeIMAPFilename(value string) string {
	value = decodeMIMEHeaderValue(value)
	value = unsafeIMAPFilenamePattern.ReplaceAllString(value, " ")
	value = strings.Join(strings.Fields(value), "_")
	value = strings.Trim(value, "._-")
	if value == "" {
		return "file"
	}
	if len(value) > 180 {
		return value[:180]
	}
	return value
}

func safeEmailBodyFilename(subject string) string {
	base := sanitizeIMAPFilename(subject)
	if base == "" {
		base = "email"
	}
	if !strings.HasSuffix(strings.ToLower(base), ".txt") {
		base += ".txt"
	}
	return base
}
