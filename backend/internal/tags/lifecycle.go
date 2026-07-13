package tags

import "strings"

const (
	StatusActive   = "active"
	StatusArchived = "archived"
	StatusHidden   = "hidden"
)

const MaxTagNameLength = 80

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

func EffectiveLifecycleStatus(tag *Tag) string {
	if tag == nil {
		return StatusActive
	}
	if tag.DeletedAt.Valid {
		return "deleted"
	}
	return NormalizeStatus(tag.Status)
}

func IsTagVisible(tag *Tag, isAdmin bool) bool {
	if tag == nil || tag.DeletedAt.Valid {
		return false
	}
	if tag.Status == StatusHidden && !isAdmin {
		return false
	}
	if tag.Status == StatusArchived && !isAdmin {
		return false
	}
	return true
}
