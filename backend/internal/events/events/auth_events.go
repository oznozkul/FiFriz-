package domain

import "github.com/google/uuid"

const (
	UserRegisteredName          = "user.registered"
	UserLoggedInName            = "user.logged_in"
	PasswordResetRequestedName  = "password.reset_requested"
	PasswordResetCompletedName  = "password.reset_completed"
	ProfileUpdatedName          = "profile.updated"
	PasswordChangedName         = "password.changed"
)

type UserRegistered struct {
	Base
	UserID uuid.UUID
	Email  string
}

func NewUserRegistered(tenantID, userID uuid.UUID, email string) UserRegistered {
	base := NewBase(UserRegisteredName, tenantID, &userID)
	return UserRegistered{Base: base, UserID: userID, Email: email}
}

type UserLoggedIn struct {
	Base
	UserID uuid.UUID
	Email  string
}

func NewUserLoggedIn(tenantID, userID uuid.UUID, email string) UserLoggedIn {
	base := NewBase(UserLoggedInName, tenantID, &userID)
	return UserLoggedIn{Base: base, UserID: userID, Email: email}
}

type PasswordResetRequested struct {
	Base
	UserID uuid.UUID
	Email  string
}

func NewPasswordResetRequested(tenantID, userID uuid.UUID, email string) PasswordResetRequested {
	base := NewBase(PasswordResetRequestedName, tenantID, &userID)
	return PasswordResetRequested{Base: base, UserID: userID, Email: email}
}

type PasswordResetCompleted struct {
	Base
	UserID uuid.UUID
	Email  string
}

func NewPasswordResetCompleted(tenantID, userID uuid.UUID, email string) PasswordResetCompleted {
	base := NewBase(PasswordResetCompletedName, tenantID, &userID)
	return PasswordResetCompleted{Base: base, UserID: userID, Email: email}
}

type ProfileUpdated struct {
	Base
	UserID   uuid.UUID
	FullName string
}

func NewProfileUpdated(tenantID, userID uuid.UUID, fullName string) ProfileUpdated {
	base := NewBase(ProfileUpdatedName, tenantID, &userID)
	return ProfileUpdated{Base: base, UserID: userID, FullName: fullName}
}

type PasswordChanged struct {
	Base
	UserID uuid.UUID
	Email  string
}

func NewPasswordChanged(tenantID, userID uuid.UUID, email string) PasswordChanged {
	base := NewBase(PasswordChangedName, tenantID, &userID)
	return PasswordChanged{Base: base, UserID: userID, Email: email}
}
