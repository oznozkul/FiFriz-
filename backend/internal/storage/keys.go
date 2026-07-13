package storage

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func OriginalDocumentKey(tenantID, documentID uuid.UUID, storedFileName string) string {
	return fmt.Sprintf("tenants/%s/documents/%s/original/%s", tenantID, documentID, storedFileName)
}

func VersionDocumentKey(tenantID, documentID uuid.UUID, versionNumber int, storedFileName string) string {
	return fmt.Sprintf("tenants/%s/documents/%s/versions/%d/%s", tenantID, documentID, versionNumber, storedFileName)
}

func IntegrationDocumentKey(tenantID, integrationAccountID, documentID uuid.UUID, storedFileName string) string {
	return fmt.Sprintf("tenants/%s/integrations/%s/documents/%s/%s", tenantID, integrationAccountID, documentID, storedFileName)
}

func CompressedKey(key string) string {
	if strings.HasSuffix(key, ".gz") {
		return key
	}
	return key + ".gz"
}

func ValidateTenantKey(tenantID uuid.UUID, key string) bool {
	prefix := fmt.Sprintf("tenants/%s/", tenantID)
	return strings.HasPrefix(key, prefix)
}
