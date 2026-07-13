package permissions

import "fifriz/internal/auth"

func SubjectFromAuth(ctx auth.Context) Subject {
	return Subject{
		TenantID: ctx.TenantID,
		UserID:   ctx.UserID,
		Role:     Role(ctx.Role),
	}
}
