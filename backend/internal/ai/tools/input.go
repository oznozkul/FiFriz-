package tools

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func parseDocumentID(input map[string]interface{}) (uuid.UUID, error) {
	raw, ok := input["documentId"].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		return uuid.Nil, fmt.Errorf("%w: documentId required", ErrInvalidToolInput)
	}
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid documentId", ErrInvalidToolInput)
	}
	return id, nil
}

func parseOptionalUUID(input map[string]interface{}, key string) (*uuid.UUID, error) {
	raw, ok := input[key].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid %s", ErrInvalidToolInput, key)
	}
	return &id, nil
}

func parseRequiredString(input map[string]interface{}, key string) (string, error) {
	raw, ok := input[key].(string)
	value := strings.TrimSpace(raw)
	if !ok || value == "" {
		return "", fmt.Errorf("%w: %s required", ErrInvalidToolInput, key)
	}
	return value, nil
}

func parseOptionalString(input map[string]interface{}, key string) string {
	raw, _ := input[key].(string)
	return strings.TrimSpace(raw)
}

func parseOptionalInt(input map[string]interface{}, key string) *int {
	switch value := input[key].(type) {
	case float64:
		v := int(value)
		return &v
	case int:
		return &value
	default:
		return nil
	}
}
