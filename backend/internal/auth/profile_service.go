package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"fifriz/internal/email"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	domain "fifriz/internal/events/events"
)

const (
	forgotPasswordSuccessMessage = "Eğer bu email kayıtlıysa şifre sıfırlama bağlantısı gönderildi."
)

var (
	ErrWeakPassword           = errors.New("password does not meet requirements")
	ErrInvalidResetToken      = errors.New("invalid or expired reset token")
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
	ErrInvalidProfile         = errors.New("invalid profile data")
)

type PasswordResetConfig struct {
	FrontendURL string
	TokenTTL    time.Duration
	AppEnv      string
}

func (s *Service) SetPasswordResetConfig(cfg PasswordResetConfig) {
	s.resetConfig = cfg
}

func (s *Service) SetEmailSender(sender email.Sender) {
	s.emailSender = sender
}

func (s *Service) ForgotPassword(ctx context.Context, input ForgotPasswordRequest) (*MessageResponse, error) {
	emailAddr := normalizeEmail(input.Email)
	if emailAddr == "" {
		return &MessageResponse{Message: forgotPasswordSuccessMessage}, nil
	}

	user, err := s.repository.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &MessageResponse{Message: forgotPasswordSuccessMessage}, nil
		}
		return nil, err
	}

	plainToken, tokenHash, err := generateResetToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(s.resetTokenTTL())
	resetToken := &PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	if err := s.repository.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return nil, err
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(s.frontendURL(), "/"), plainToken)
	if s.emailSender != nil {
		if err := s.emailSender.SendPasswordResetEmail(ctx, user.Email, resetLink); err != nil {
			return nil, err
		}
	}

	members, err := s.repository.ListMembershipsByUserID(ctx, user.ID)
	if err == nil && len(members) > 0 {
		userID := user.ID
		tenantID := members[0].TenantID
		s.publisher.PublishAfterCommit(ctx, domain.NewPasswordResetRequested(tenantID, userID, user.Email))
	}

	return &MessageResponse{Message: forgotPasswordSuccessMessage}, nil
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordRequest) (*MessageResponse, error) {
	token := strings.TrimSpace(input.Token)
	newPassword := strings.TrimSpace(input.NewPassword)
	if token == "" || newPassword == "" {
		return nil, ErrInvalidResetToken
	}
	if err := validatePassword(newPassword); err != nil {
		return nil, err
	}

	resetToken, err := s.repository.GetValidPasswordResetToken(ctx, hashResetToken(token))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidResetToken
		}
		return nil, err
	}

	user, err := s.repository.GetUserByID(ctx, resetToken.UserID)
	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if err := s.repository.UpdateUserPassword(ctx, user.ID, string(passwordHash)); err != nil {
		return nil, err
	}
	if err := s.repository.MarkPasswordResetTokenUsed(ctx, resetToken.ID); err != nil {
		return nil, err
	}
	_ = s.repository.InvalidatePasswordResetTokens(ctx, user.ID)

	members, err := s.repository.ListMembershipsByUserID(ctx, user.ID)
	if err == nil && len(members) > 0 {
		userID := user.ID
		tenantID := members[0].TenantID
		s.publisher.PublishAfterCommit(ctx, domain.NewPasswordResetCompleted(tenantID, userID, user.Email))
	}

	return &MessageResponse{Message: "Şifreniz başarıyla güncellendi."}, nil
}

func (s *Service) GetProfile(ctx context.Context, authContext Context) (*MeResponse, error) {
	return s.Me(ctx, authContext)
}

func (s *Service) UpdateProfile(ctx context.Context, authContext Context, input UpdateProfileRequest) (*MeResponse, error) {
	fullName := strings.TrimSpace(input.FullName)
	if fullName == "" {
		return nil, ErrInvalidProfile
	}

	if err := s.repository.UpdateUserFullName(ctx, authContext.UserID, fullName); err != nil {
		return nil, err
	}

	userID := authContext.UserID
	tenantID := authContext.TenantID
	s.publisher.PublishAfterCommit(ctx, domain.NewProfileUpdated(tenantID, userID, fullName))

	return s.Me(ctx, authContext)
}

func (s *Service) ChangePassword(ctx context.Context, authContext Context, input ChangePasswordRequest) (*MessageResponse, error) {
	currentPassword := strings.TrimSpace(input.CurrentPassword)
	newPassword := strings.TrimSpace(input.NewPassword)
	if currentPassword == "" || newPassword == "" {
		return nil, ErrWeakPassword
	}
	if err := validatePassword(newPassword); err != nil {
		return nil, err
	}

	user, err := s.repository.GetUserByID(ctx, authContext.UserID)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return nil, ErrInvalidCurrentPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if err := s.repository.UpdateUserPassword(ctx, user.ID, string(passwordHash)); err != nil {
		return nil, err
	}

	userID := user.ID
	tenantID := authContext.TenantID
	s.publisher.PublishAfterCommit(ctx, domain.NewPasswordChanged(tenantID, userID, user.Email))

	return &MessageResponse{Message: "Şifreniz değiştirildi."}, nil
}

func (s *Service) GetProfileTenant(ctx context.Context, authContext Context) (*TenantProfileResponse, error) {
	tenant, err := s.repository.GetTenantByID(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	member, err := s.repository.GetMembership(ctx, authContext.UserID, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	memberCount, err := s.repository.CountMembersByTenantID(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	return &TenantProfileResponse{
		TenantID:    tenant.ID,
		TenantName:  tenant.Name,
		TenantType:  tenant.Type,
		UserRole:    member.Role,
		MemberCount: memberCount,
		CreatedAt:   tenant.CreatedAt,
	}, nil
}

func (s *Service) frontendURL() string {
	if s.resetConfig.FrontendURL != "" {
		return s.resetConfig.FrontendURL
	}
	return "http://localhost:5173"
}

func (s *Service) resetTokenTTL() time.Duration {
	if s.resetConfig.TokenTTL > 0 {
		return s.resetConfig.TokenTTL
	}
	return 30 * time.Minute
}
