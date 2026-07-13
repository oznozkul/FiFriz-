package auth

import (
	"context"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) (tenantID string, userID string, err error) {
	if err := db.AutoMigrate(&Tenant{}, &User{}, &TenantMember{}, &PasswordResetToken{}); err != nil {
		return "", "", err
	}

	repository := NewRepository(db)
	service := NewService(repository, JWTConfig{Secret: "migration-only", Issuer: "fifriz", Duration: 0}, nil)
	legacyTenantID, legacyUserID, err := service.EnsureLegacyIdentity(context.Background())
	if err != nil {
		return "", "", err
	}

	return legacyTenantID.String(), legacyUserID.String(), nil
}
