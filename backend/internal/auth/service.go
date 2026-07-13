package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/email"
)

const (
	LegacyTenantID = "00000000-0000-0000-0000-000000000101"
	LegacyUserID   = "00000000-0000-0000-0000-000000000102"
	LegacyMemberID = "00000000-0000-0000-0000-000000000103"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidRegistration    = errors.New("invalid registration data")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrNoTenantMembership     = errors.New("tenant membership not found")
	ErrInvalidToken           = errors.New("invalid token")
)

type JWTConfig struct {
	Secret   string
	Issuer   string
	Duration time.Duration
}

type Service struct {
	repository        *Repository
	jwtConfig         JWTConfig
	publisher         events.Publisher
	personalBootstrap PersonalTenantBootstrapper
	resetConfig       PasswordResetConfig
	emailSender       email.Sender
}

type PersonalTenantBootstrapper interface {
	BootstrapPersonalTenant(ctx context.Context, tenantID, userID uuid.UUID) error
}

func (s *Service) SetPersonalTenantBootstrapper(bootstrapper PersonalTenantBootstrapper) {
	s.personalBootstrap = bootstrapper
}

type Claims struct {
	UserID   string     `json:"userId"`
	TenantID string     `json:"tenantId"`
	Email    string     `json:"email"`
	Role     TenantRole `json:"role"`
	jwt.RegisteredClaims
}

func NewService(repository *Repository, jwtConfig JWTConfig, publisher events.Publisher) *Service {
	return &Service{
		repository: repository,
		jwtConfig:  jwtConfig,
		publisher:  publisher,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterRequest) (*AuthResponse, error) {
	input.AccountType = TenantType(strings.TrimSpace(string(input.AccountType)))
	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = normalizeEmail(input.Email)
	input.Password = strings.TrimSpace(input.Password)
	input.OrganizationName = strings.TrimSpace(input.OrganizationName)

	if input.AccountType != TenantTypePersonal && input.AccountType != TenantTypeOrganization {
		return nil, ErrInvalidRegistration
	}
	if input.FullName == "" || input.Email == "" || len(input.Password) < 6 {
		return nil, ErrInvalidRegistration
	}
	if input.AccountType == TenantTypeOrganization && input.OrganizationName == "" {
		return nil, ErrInvalidRegistration
	}

	if _, err := s.repository.GetUserByEmail(ctx, input.Email); err == nil {
		return nil, ErrEmailAlreadyRegistered
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	tenantName := input.OrganizationName
	if input.AccountType == TenantTypePersonal {
		tenantName = input.FullName
		if tenantName == "" {
			tenantName = input.Email
		}
	}

	user := &User{
		FullName:     input.FullName,
		Email:        input.Email,
		PasswordHash: string(passwordHash),
	}
	tenant := &Tenant{
		Name: tenantName,
		Type: input.AccountType,
	}
	member := &TenantMember{
		Role: TenantRoleOwner,
	}

	if err := s.repository.CreateRegistration(ctx, user, tenant, member); err != nil {
		return nil, err
	}

	if s.personalBootstrap != nil && tenant.Type == TenantTypePersonal {
		if err := s.personalBootstrap.BootstrapPersonalTenant(ctx, tenant.ID, user.ID); err != nil {
			return nil, err
		}
	}

	member.User = *user
	member.Tenant = *tenant

	userID := user.ID
	tenantID := tenant.ID
	s.publisher.PublishAfterCommit(ctx, domain.NewUserRegistered(tenantID, userID, user.Email))

	return s.buildAuthResponse(ctx, *user, *member)
}

func (s *Service) Login(ctx context.Context, input LoginRequest) (*AuthResponse, error) {
	email := normalizeEmail(input.Email)
	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	members, err := s.repository.ListMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, ErrNoTenantMembership
	}

	userID := user.ID
	tenantID := members[0].TenantID
	s.publisher.PublishAfterCommit(ctx, domain.NewUserLoggedIn(tenantID, userID, user.Email))

	return s.buildAuthResponse(ctx, *user, members[0])
}

func (s *Service) Me(ctx context.Context, authContext Context) (*MeResponse, error) {
	user, err := s.repository.GetUserByID(ctx, authContext.UserID)
	if err != nil {
		return nil, err
	}

	member, err := s.repository.GetMembership(ctx, authContext.UserID, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	members, err := s.repository.ListMembershipsByUserID(ctx, authContext.UserID)
	if err != nil {
		return nil, err
	}

	return &MeResponse{
		User:         toUserResponse(*user),
		ActiveTenant: toTenantResponse(*member),
		Tenants:      toTenantResponses(members),
	}, nil
}

func (s *Service) ListTenantMembers(ctx context.Context) ([]TenantMemberResponse, error) {
	authContext, err := RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	members, err := s.repository.ListMembersByTenantID(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	responses := make([]TenantMemberResponse, 0, len(members))
	for _, member := range members {
		responses = append(responses, TenantMemberResponse{
			UserID:   member.UserID,
			FullName: member.User.FullName,
			Email:    member.User.Email,
			Role:     member.Role,
		})
	}
	return responses, nil
}

func (s *Service) VerifyMembership(ctx context.Context, authContext Context) (Context, error) {
	member, err := s.repository.GetMembership(ctx, authContext.UserID, authContext.TenantID)
	if err != nil {
		return Context{}, err
	}

	authContext.Role = member.Role
	return authContext, nil
}

func (s *Service) ParseToken(tokenString string) (Context, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return Context{}, ErrInvalidToken
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.jwtConfig.Secret), nil
	})
	if err != nil || !token.Valid {
		return Context{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return Context{}, ErrInvalidToken
	}
	tenantID, err := uuid.Parse(claims.TenantID)
	if err != nil {
		return Context{}, ErrInvalidToken
	}

	return Context{
		UserID:   userID,
		TenantID: tenantID,
		Email:    claims.Email,
		Role:     claims.Role,
	}, nil
}

func (s *Service) buildAuthResponse(ctx context.Context, user User, activeMember TenantMember) (*AuthResponse, error) {
	members, err := s.repository.ListMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.issueToken(user, activeMember)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		User:         toUserResponse(user),
		ActiveTenant: toTenantResponse(activeMember),
		Tenants:      toTenantResponses(members),
	}, nil
}

func (s *Service) issueToken(user User, member TenantMember) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID.String(),
		TenantID: member.TenantID.String(),
		Email:    user.Email,
		Role:     member.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.jwtConfig.Issuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtConfig.Duration)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtConfig.Secret))
}

func toTenantResponses(members []TenantMember) []TenantResponse {
	responses := make([]TenantResponse, 0, len(members))
	for _, member := range members {
		responses = append(responses, toTenantResponse(member))
	}
	return responses
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func LegacyIDs() (uuid.UUID, uuid.UUID, uuid.UUID) {
	return uuid.MustParse(LegacyTenantID), uuid.MustParse(LegacyUserID), uuid.MustParse(LegacyMemberID)
}

func BuildLegacyIdentity() (*Tenant, *User, *TenantMember) {
	tenantID, userID, memberID := LegacyIDs()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)

	tenant := &Tenant{
		ID:   tenantID,
		Name: "Legacy Personal Tenant",
		Type: TenantTypePersonal,
	}
	user := &User{
		ID:           userID,
		FullName:     "Legacy User",
		Email:        "legacy@fifriz.local",
		PasswordHash: string(passwordHash),
	}
	member := &TenantMember{
		ID:       memberID,
		TenantID: tenantID,
		UserID:   userID,
		Role:     TenantRoleOwner,
	}

	return tenant, user, member
}

func (s *Service) EnsureLegacyIdentity(ctx context.Context) (uuid.UUID, uuid.UUID, error) {
	tenant, user, member := BuildLegacyIdentity()
	if err := s.repository.UpsertLegacyIdentity(ctx, user, tenant, member); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("ensure legacy identity: %w", err)
	}
	return tenant.ID, user.ID, nil
}
