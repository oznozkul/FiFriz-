package documents

import "time"

// Document lifecycle states (canonical values stored in documents.status).
const (
	DocumentStatusUploading         DocumentStatus = "uploading"
	DocumentStatusProcessing        DocumentStatus = "processing"
	DocumentStatusReady             DocumentStatus = "ready"
	DocumentStatusFailed            DocumentStatus = "failed"
	DocumentStatusArchived          DocumentStatus = "archived"
	DocumentStatusDeleted           DocumentStatus = "deleted"
	DocumentStatusLocked            DocumentStatus = "locked"
	DocumentStatusExpired           DocumentStatus = "expired"
	DocumentStatusRetentionLocked   DocumentStatus = "retention_locked"

	// Legacy aliases — accepted on read, normalized on write.
	DocumentStatusUploaded  DocumentStatus = "uploaded"
	DocumentStatusProcessed DocumentStatus = "processed"
)

func NormalizeStatus(status DocumentStatus) DocumentStatus {
	switch status {
	case DocumentStatusUploaded:
		return DocumentStatusUploading
	case DocumentStatusProcessed:
		return DocumentStatusReady
	case "":
		return DocumentStatusUploading
	default:
		return status
	}
}

func IsReadyStatus(status DocumentStatus) bool {
	switch NormalizeStatus(status) {
	case DocumentStatusReady:
		return true
	default:
		return false
	}
}

func IsProcessingStatus(status DocumentStatus) bool {
	switch NormalizeStatus(status) {
	case DocumentStatusUploading, DocumentStatusProcessing:
		return true
	default:
		return false
	}
}

func isValidLifecycleStatus(status DocumentStatus) bool {
	switch NormalizeStatus(status) {
	case DocumentStatusUploading, DocumentStatusProcessing, DocumentStatusReady, DocumentStatusFailed,
		DocumentStatusArchived, DocumentStatusDeleted, DocumentStatusExpired, DocumentStatusRetentionLocked:
		return true
	default:
		return false
	}
}

// EffectiveLifecycleStatus returns the user-facing lifecycle state for badges.
func EffectiveLifecycleStatus(document *Document, inTrash bool) DocumentStatus {
	if document == nil {
		return DocumentStatusUploading
	}
	if document.IsLocked {
		return DocumentStatusLocked
	}
	if document.RetentionLockedAt != nil {
		return DocumentStatusRetentionLocked
	}
	if document.ExpiresAt != nil && document.ExpiresAt.Before(time.Now()) {
		return DocumentStatusExpired
	}
	if document.ArchivedAt != nil {
		return DocumentStatusArchived
	}
	if inTrash {
		return DocumentStatusDeleted
	}
	return NormalizeStatus(document.Status)
}
