package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	AccountType      TenantType `json:"accountType"`
	FullName         string     `json:"fullName"`
	Email            string     `json:"email"`
	Password         string     `json:"password"`
	OrganizationName string     `json:"organizationName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"fullName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type TenantResponse struct {
	ID   uuid.UUID  `json:"id"`
	Name string     `json:"name"`
	Type TenantType `json:"type"`
	Role TenantRole `json:"role"`
}

type AuthResponse struct {
	AccessToken  string           `json:"accessToken"`
	User         UserResponse     `json:"user"`
	ActiveTenant TenantResponse   `json:"activeTenant"`
	Tenants      []TenantResponse `json:"tenants"`
}

type MeResponse struct {
	User         UserResponse     `json:"user"`
	ActiveTenant TenantResponse   `json:"activeTenant"`
	Tenants      []TenantResponse `json:"tenants"`
}

type TenantMemberResponse struct {
	UserID   uuid.UUID  `json:"userId"`
	FullName string     `json:"fullName"`
	Email    string     `json:"email"`
	Role     TenantRole `json:"role"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

type UpdateProfileRequest struct {
	FullName string `json:"fullName"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type TenantProfileResponse struct {
	TenantID    uuid.UUID  `json:"tenantId"`
	TenantName  string     `json:"tenantName"`
	TenantType  TenantType `json:"tenantType"`
	UserRole    TenantRole `json:"userRole"`
	MemberCount int        `json:"memberCount"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func toUserResponse(user User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func toTenantResponse(member TenantMember) TenantResponse {
	return TenantResponse{
		ID:   member.Tenant.ID,
		Name: member.Tenant.Name,
		Type: member.Tenant.Type,
		Role: member.Role,
	}
}
