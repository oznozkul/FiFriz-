package integrations

const (
	ProviderGmail        = "gmail"
	ProviderOutlook      = "outlook"
	ProviderIMAP         = "imap"
	ProviderGoogleDrive  = "google_drive"
	ProviderOneDrive     = "onedrive"
	ProviderDropbox      = "dropbox"
	ProviderManualUpload = "manual_upload"
)

const (
	StatusDisconnected  = "disconnected"
	StatusConnected     = "connected"
	StatusSyncPending   = "sync_pending"
	StatusSyncing       = "syncing"
	StatusSyncFailed    = "sync_failed"
	StatusSyncCompleted = "sync_completed"
)

const (
	SyncTypeManual         = "manual"
	SyncTypeScheduled      = "scheduled"
	SyncTypeInitialHistory = "initial_history"
)

const (
	SyncStatusPending   = "pending"
	SyncStatusRunning   = "running"
	SyncStatusCompleted = "completed"
	SyncStatusFailed    = "failed"
)

const (
	ExternalTypeEmailBody        = "email_body"
	ExternalTypeEmailAttachment  = "email_attachment"
	ExternalTypeDriveFile        = "drive_file"
)

func IsValidProvider(provider string) bool {
	switch provider {
	case ProviderGmail, ProviderOutlook, ProviderIMAP, ProviderGoogleDrive, ProviderOneDrive, ProviderDropbox, ProviderManualUpload:
		return true
	default:
		return false
	}
}

func IsMailProvider(provider string) bool {
	switch provider {
	case ProviderGmail, ProviderOutlook, ProviderIMAP:
		return true
	default:
		return false
	}
}

func IsSyncableProvider(provider string) bool {
	switch provider {
	case ProviderGmail, ProviderOutlook, ProviderIMAP, ProviderGoogleDrive, ProviderOneDrive:
		return true
	default:
		return false
	}
}
