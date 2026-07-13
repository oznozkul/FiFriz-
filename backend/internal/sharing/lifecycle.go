package sharing

import (
	"strings"
	"time"
)

const (
	ShareTypePublicLink         = "public_link"
	ShareTypeEmailInvite        = "email_invite"
	ShareTypeInternalUser       = "internal_user"
	ShareTypeInternalWorkspace  = "internal_workspace"

	ShareStatusActive   = "active"
	ShareStatusExpired  = "expired"
	ShareStatusRevoked  = "revoked"
	ShareStatusDisabled = "disabled"

	ExpirationNever  = "never"
	Expiration1Day   = "1d"
	Expiration7Days  = "7d"
	Expiration30Days = "30d"
	ExpirationCustom = "custom"
)

func EffectiveShareStatus(link *DocumentShareLink) string {
	if link == nil {
		return ShareStatusActive
	}
	if link.DeletedAt.Valid || link.Status == ShareStatusRevoked {
		return ShareStatusRevoked
	}
	if !link.IsActive || link.Status == ShareStatusDisabled {
		return ShareStatusDisabled
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return ShareStatusExpired
	}
	if link.MaxViewCount != nil && link.ViewCount >= *link.MaxViewCount {
		return ShareStatusDisabled
	}
	return ShareStatusActive
}

func ResolveExpiration(preset string, custom *time.Time) *time.Time {
	switch strings.TrimSpace(strings.ToLower(preset)) {
	case "", ExpirationNever:
		return nil
	case Expiration1Day:
		value := time.Now().Add(24 * time.Hour)
		return &value
	case Expiration7Days:
		value := time.Now().Add(7 * 24 * time.Hour)
		return &value
	case Expiration30Days:
		value := time.Now().Add(30 * 24 * time.Hour)
		return &value
	case ExpirationCustom:
		return custom
	default:
		return custom
	}
}

func DefaultWatermarkTemplate(ownerName string) string {
	now := time.Now().Format("2006-01-02 15:04")
	if strings.TrimSpace(ownerName) == "" {
		return now + "\nShared via FiFriz"
	}
	return ownerName + "\n" + now + "\nShared via FiFriz"
}
