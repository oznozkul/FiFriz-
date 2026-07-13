package folders

import (
	"strings"
	"time"
)

const (
	StatusActive   = "active"
	StatusArchived = "archived"
	StatusHidden   = "hidden"
)

func NormalizeStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "", StatusActive:
		return StatusActive
	case StatusArchived:
		return StatusArchived
	case StatusHidden:
		return StatusHidden
	default:
		return StatusActive
	}
}

func EffectiveLifecycleStatus(folder *Folder) string {
	if folder == nil {
		return StatusActive
	}
	if folder.DeletedAt.Valid {
		return "deleted"
	}
	if folder.IsLocked {
		return "locked"
	}
	return NormalizeStatus(folder.Status)
}

func IsFolderLocked(folder *Folder) bool {
	return folder != nil && folder.IsLocked
}

func IsFolderVisible(folder *Folder, isAdmin bool) bool {
	if folder == nil {
		return false
	}
	if folder.DeletedAt.Valid {
		return false
	}
	if folder.Status == StatusHidden && !isAdmin {
		return false
	}
	return true
}

func touchActivity(folder *Folder) {
	now := time.Now()
	folder.LastActivityAt = &now
}
