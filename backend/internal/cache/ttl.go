package cache

import "time"

const (
	TTLDashboard        = 60 * time.Second
	TTLFolderTree       = 5 * time.Minute
	TTLDocumentDetail   = 2 * time.Minute
	TTLDocumentWorkspace = 2 * time.Minute
	TTLDocumentList     = 60 * time.Second
	TTLTags             = 5 * time.Minute
	TTLNotifications    = 30 * time.Second
	TTLSmartCollections = 60 * time.Second
	TTLGlobalSearch     = 30 * time.Second
	TTLHybridSearch       = 30 * time.Second
	TTLWorkspaceInsights  = 60 * time.Second
	TTLNavigationPins   = 5 * time.Minute
	TTLNavigationRecent = 5 * time.Minute
)
