package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateRegistration(ctx context.Context, user *User, tenant *Tenant, member *TenantMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tenant).Error; err != nil {
			return err
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		member.TenantID = tenant.ID
		member.UserID = user.ID
		return tx.Create(member).Error
	})
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) TenantType(ctx context.Context, tenantID uuid.UUID) (TenantType, error) {
	var tenant Tenant
	err := r.db.WithContext(ctx).Select("type").First(&tenant, "id = ?", tenantID).Error
	if err != nil {
		return "", err
	}
	return tenant.Type, nil
}

func (r *Repository) ListMembershipsByUserID(ctx context.Context, userID uuid.UUID) ([]TenantMember, error) {
	var members []TenantMember
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

func (r *Repository) GetMembership(ctx context.Context, userID, tenantID uuid.UUID) (*TenantMember, error) {
	var member TenantMember
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		First(&member, "user_id = ? AND tenant_id = ?", userID, tenantID).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *Repository) ListMembersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]TenantMember, error) {
	var members []TenantMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("tenant_id = ?", tenantID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

func (r *Repository) UpsertLegacyIdentity(ctx context.Context, user *User, tenant *Tenant, member *TenantMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(tenant).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(user).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(member).Error
	})
}

func (r *Repository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *Repository) GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	var token PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now().UTC()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *Repository) MarkPasswordResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("id = ?", tokenID).
		Update("used_at", now).Error
}

func (r *Repository) InvalidatePasswordResetTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}

func (r *Repository) UpdateUserFullName(ctx context.Context, userID uuid.UUID, fullName string) error {
	return r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", userID).
		Update("full_name", fullName).Error
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
}

func (r *Repository) CountMembersByTenantID(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TenantMember{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return int(count), err
}

func (r *Repository) GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	var tenant Tenant
	err := r.db.WithContext(ctx).First(&tenant, "id = ?", tenantID).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}
