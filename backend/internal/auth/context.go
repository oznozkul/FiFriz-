package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type contextKey string

const authContextKey contextKey = "auth_context"

var ErrMissingAuthContext = errors.New("missing auth context")

type Context struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Email    string
	Role     TenantRole
}

func WithContext(ctx context.Context, authContext Context) context.Context {
	return context.WithValue(ctx, authContextKey, authContext)
}

func FromContext(ctx context.Context) (Context, bool) {
	value, ok := ctx.Value(authContextKey).(Context)
	return value, ok
}

func RequireContext(ctx context.Context) (Context, error) {
	authContext, ok := FromContext(ctx)
	if !ok || authContext.UserID == uuid.Nil || authContext.TenantID == uuid.Nil {
		return Context{}, ErrMissingAuthContext
	}
	return authContext, nil
}
