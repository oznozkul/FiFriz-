package validation

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrMIMETypeMismatch = errors.New("file content does not match extension")

var extensionAllowedMIMEs = map[string][]string{
	".pdf":  {"application/pdf"},
	".png":  {"image/png"},
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	".xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
}

// DetectContentType inspects magic bytes and returns a MIME type.
func DetectContentType(data []byte, extension string) string {
	if len(data) == 0 {
		return "application/octet-stream"
	}

	ext := strings.ToLower(strings.TrimSpace(extension))
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	switch ext {
	case ".pdf":
		if bytes.HasPrefix(data, []byte("%PDF")) {
			return "application/pdf"
		}
	case ".png":
		if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
			return "image/png"
		}
	case ".jpg", ".jpeg":
		if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg"
		}
	case ".docx", ".xlsx":
		if bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04}) {
			if officeMIME, ok := detectOfficeMIME(data, ext); ok {
				return officeMIME
			}
		}
	}

	return "application/octet-stream"
}

func detectOfficeMIME(data []byte, extension string) (string, bool) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", false
	}

	hasWord := false
	hasExcel := false
	for _, file := range reader.File {
		name := strings.ToLower(file.Name)
		if strings.HasPrefix(name, "word/") {
			hasWord = true
		}
		if strings.HasPrefix(name, "xl/") {
			hasExcel = true
		}
	}

	switch extension {
	case ".docx":
		if hasWord {
			return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true
		}
	case ".xlsx":
		if hasExcel {
			return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true
		}
	}
	return "", false
}

// ValidateContent validates file bytes against the declared extension using magic bytes.
func ValidateContent(data []byte, extension string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(extension))
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	allowed, ok := extensionAllowedMIMEs[ext]
	if !ok {
		return "", fmt.Errorf("unsupported extension %s", ext)
	}

	detected := DetectContentType(data, ext)
	for _, candidate := range allowed {
		if detected == candidate {
			return detected, nil
		}
	}

	return "", fmt.Errorf("%w: expected %v for %s, detected %s", ErrMIMETypeMismatch, allowed, ext, detected)
}

// ValidateReader reads up to maxBytes and validates content.
func ValidateReader(reader io.Reader, size int64, extension string, maxBytes int64) ([]byte, string, error) {
	limited := io.LimitReader(reader, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > maxBytes {
		return nil, "", fmt.Errorf("file exceeds maximum allowed size")
	}
	if size > 0 && int64(len(data)) != size {
		return data, "", fmt.Errorf("upload size mismatch")
	}

	mimeType, err := ValidateContent(data, extension)
	if err != nil {
		return data, "", err
	}
	return data, mimeType, nil
}
