package sharing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShareLinkAnalytics struct {
	ShareLinkID      uuid.UUID  `json:"shareLinkId"`
	DocumentID       uuid.UUID  `json:"documentId"`
	OpenCount        int        `json:"openCount"`
	DownloadCount    int        `json:"downloadCount"`
	LastOpenedAt     *time.Time `json:"lastOpenedAt"`
	LastDownloadedAt *time.Time `json:"lastDownloadedAt"`
	LastIPAddress    *string    `json:"lastIpAddress"`
	LastUserAgent    *string    `json:"lastUserAgent"`
	CreatedByUserID  uuid.UUID  `json:"createdByUserId"`
	CreatedByName    string     `json:"createdByName"`
	Status           string     `json:"status"`
}

func (r *Repository) RecordShareOpen(ctx context.Context, id uuid.UUID, ip, userAgent string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"view_count":     gorm.Expr("view_count + 1"),
		"last_opened_at": now,
	}
	if ip != "" {
		updates["last_ip_address"] = ip
	}
	if userAgent != "" {
		updates["last_user_agent"] = userAgent
	}
	return r.db.WithContext(ctx).
		Model(&DocumentShareLink{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) RecordShareDownload(ctx context.Context, id uuid.UUID, ip, userAgent string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"download_count":     gorm.Expr("download_count + 1"),
		"last_downloaded_at": now,
	}
	if ip != "" {
		updates["last_ip_address"] = ip
	}
	if userAgent != "" {
		updates["last_user_agent"] = userAgent
	}
	return r.db.WithContext(ctx).
		Model(&DocumentShareLink{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *Repository) Revoke(ctx context.Context, tenantID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&DocumentShareLink{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]interface{}{
			"is_active": false,
			"status":    ShareStatusRevoked,
		}).Error
}

func (r *Repository) MarkExpired(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&DocumentShareLink{}).
		Where("id = ?", id).
		Update("status", ShareStatusExpired).Error
}

func (r *Repository) MarkAccessLimitReached(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&DocumentShareLink{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active": false,
			"status":    ShareStatusDisabled,
		}).Error
}

func (r *Repository) AnalyticsForLink(ctx context.Context, tenantID, id uuid.UUID) (*ShareLinkAnalytics, error) {
	link, err := r.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return &ShareLinkAnalytics{
		ShareLinkID:      link.ID,
		DocumentID:       link.DocumentID,
		OpenCount:        link.ViewCount,
		DownloadCount:    link.DownloadCount,
		LastOpenedAt:     link.LastOpenedAt,
		LastDownloadedAt: link.LastDownloadedAt,
		LastIPAddress:    link.LastIPAddress,
		LastUserAgent:    link.LastUserAgent,
		CreatedByUserID:  link.CreatedByUserID,
		Status:           EffectiveShareStatus(link),
	}, nil
}
