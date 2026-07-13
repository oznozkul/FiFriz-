package tags

import (
	"strings"
	"unicode/utf8"
)

func ValidateTagName(name string) (string, error) {
	normalized := normalizeTagName(name)
	if normalized == "" {
		return "", ErrInvalidTagData
	}
	if utf8.RuneCountInString(normalized) > MaxTagNameLength {
		return "", ErrTagNameTooLong
	}
	return normalized, nil
}

func ValidateAlias(alias string) (string, error) {
	normalized := strings.TrimSpace(alias)
	if normalized == "" {
		return "", ErrInvalidTagData
	}
	if utf8.RuneCountInString(normalized) > MaxTagNameLength {
		return "", ErrTagNameTooLong
	}
	return normalized, nil
}
