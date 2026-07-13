package integrations

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]IntegrationAccount, error) {
	var items []IntegrationAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *Repository) GetByID(ctx context.Context, tenantID, userID, id uuid.UUID) (*IntegrationAccount, error) {
	var account IntegrationAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, id).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) GetByIDForWorker(ctx context.Context, tenantID, id uuid.UUID) (*IntegrationAccount, error) {
	var account IntegrationAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) FindByProviderEmail(ctx context.Context, tenantID, userID uuid.UUID, provider, email string) (*IntegrationAccount, error) {
	var account IntegrationAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND provider = ? AND email_address = ?", tenantID, userID, provider, email).
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *Repository) ListDueForScheduledSync(ctx context.Context, dueBefore time.Time, providers []string) ([]IntegrationAccount, error) {
	if len(providers) == 0 {
		return []IntegrationAccount{}, nil
	}
	var items []IntegrationAccount
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("provider IN ?", providers).
		Where("status NOT IN ?", []string{StatusSyncing, StatusSyncPending}).
		Where("last_sync_at IS NULL OR last_sync_at < ?", dueBefore).
		Find(&items).Error
	if items == nil {
		items = []IntegrationAccount{}
	}
	return items, err
}

func (r *Repository) Create(ctx context.Context, account *IntegrationAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *Repository) Update(ctx context.Context, account *IntegrationAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *Repository) SoftDelete(ctx context.Context, tenantID, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, id).
		Delete(&IntegrationAccount{}).Error
}

func (r *Repository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status string, lastError *string) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}
	if lastError != nil {
		updates["last_error"] = *lastError
	}
	return r.db.WithContext(ctx).
		Model(&IntegrationAccount{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(updates).Error
}

func (r *Repository) MarkSyncCompleted(ctx context.Context, tenantID, id uuid.UUID, syncedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&IntegrationAccount{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Updates(map[string]any{
			"status":       StatusSyncCompleted,
			"last_sync_at": syncedAt,
			"last_error":   nil,
			"updated_at":   syncedAt,
		}).Error
}

func (r *Repository) CreateImportedExternal(ctx context.Context, item *ImportedExternalItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) ExistsImportedExternal(ctx context.Context, tenantID, integrationAccountID uuid.UUID, externalID, externalType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ImportedExternalItem{}).
		Where("tenant_id = ? AND integration_account_id = ? AND external_id = ? AND external_type = ?", tenantID, integrationAccountID, externalID, externalType).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountImportedByAccount(ctx context.Context, tenantID, integrationAccountID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ImportedExternalItem{}).
		Where("tenant_id = ? AND integration_account_id = ?", tenantID, integrationAccountID).
		Count(&count).Error
	return count, err
}

func (r *Repository) CreateSyncLog(ctx context.Context, log *IntegrationSyncLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *Repository) UpdateSyncLog(ctx context.Context, log *IntegrationSyncLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *Repository) ListSyncLogs(ctx context.Context, tenantID, userID, integrationID uuid.UUID) ([]IntegrationSyncLog, error) {
	var items []IntegrationSyncLog
	err := r.db.WithContext(ctx).
		Table("integration_sync_logs").
		Joins("JOIN integration_accounts ON integration_accounts.id = integration_sync_logs.integration_account_id").
		Where(
			"integration_sync_logs.tenant_id = ? AND integration_accounts.user_id = ? AND integration_sync_logs.integration_account_id = ?",
			tenantID, userID, integrationID,
		).
		Order("integration_sync_logs.created_at DESC").
		Find(&items).Error
	return items, err
}

type SyncCompletion struct {
	TotalItems    int
	ImportedItems int
	SkippedItems  int
	FailedItems   int
	Metadata      []byte
}

func (r *Repository) CompleteSync(ctx context.Context, tenantID, integrationID, syncLogID uuid.UUID, result SyncCompletion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		updates := map[string]any{
			"status":         SyncStatusCompleted,
			"finished_at":    now,
			"total_items":    result.TotalItems,
			"imported_items": result.ImportedItems,
			"skipped_items":  result.SkippedItems,
			"failed_items":   result.FailedItems,
		}
		if len(result.Metadata) > 0 {
			updates["metadata"] = result.Metadata
		}

		if err := tx.Model(&IntegrationSyncLog{}).
			Where("id = ? AND tenant_id = ?", syncLogID, tenantID).
			Updates(updates).Error; err != nil {
			return err
		}

		return tx.Model(&IntegrationAccount{}).
			Where("tenant_id = ? AND id = ?", tenantID, integrationID).
			Updates(map[string]any{
				"status":       StatusSyncCompleted,
				"last_sync_at": now,
				"last_error":   nil,
				"updated_at":   now,
			}).Error
	})
}

func (r *Repository) FailSync(ctx context.Context, tenantID, integrationID, syncLogID uuid.UUID, errorMessage string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		if err := tx.Model(&IntegrationSyncLog{}).
			Where("id = ? AND tenant_id = ?", syncLogID, tenantID).
			Updates(map[string]any{
				"status":        SyncStatusFailed,
				"finished_at":   now,
				"error_message": errorMessage,
			}).Error; err != nil {
			return err
		}

		return tx.Model(&IntegrationAccount{}).
			Where("tenant_id = ? AND id = ?", tenantID, integrationID).
			Updates(map[string]any{
				"status":     StatusSyncFailed,
				"last_error": errorMessage,
				"updated_at": now,
			}).Error
	})
}
