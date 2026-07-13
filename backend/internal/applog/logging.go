package applog

import "log"

func Info(format string, args ...any) {
	log.Printf("[info] "+format, args...)
}

func Warn(format string, args ...any) {
	log.Printf("[warn] "+format, args...)
}

func PermissionDenied(action, resource string, userID, tenantID string) {
	log.Printf("[security] permission_denied action=%s resource=%s user_id=%s tenant_id=%s", action, resource, userID, tenantID)
}

func TenantMismatch(resource, userID, expectedTenantID, actualTenantID string) {
	log.Printf("[security] tenant_mismatch resource=%s user_id=%s expected_tenant_id=%s actual_tenant_id=%s", resource, userID, expectedTenantID, actualTenantID)
}
