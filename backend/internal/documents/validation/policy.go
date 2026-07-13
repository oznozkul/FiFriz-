package validation

import (
	"strings"

	"github.com/google/uuid"
)

// Policy defines upload constraints. Tenant overrides merge on top of defaults.
type Policy struct {
	MaxFileSizeBytes int64
	AllowedMIMETypes map[string]struct{}
}

type TenantPolicyOverride struct {
	TenantID         uuid.UUID
	MaxFileSizeBytes *int64
	AllowedMIMETypes []string
}

type PolicyResolver struct {
	defaults  Policy
	overrides map[uuid.UUID]TenantPolicyOverride
}

func NewPolicyResolver(defaults Policy, overrides []TenantPolicyOverride) *PolicyResolver {
	overrideMap := make(map[uuid.UUID]TenantPolicyOverride, len(overrides))
	for _, item := range overrides {
		overrideMap[item.TenantID] = item
	}
	return &PolicyResolver{defaults: defaults, overrides: overrideMap}
}

func (r *PolicyResolver) Resolve(tenantID uuid.UUID) Policy {
	policy := r.defaults
	override, ok := r.overrides[tenantID]
	if !ok {
		return policy
	}
	if override.MaxFileSizeBytes != nil && *override.MaxFileSizeBytes > 0 {
		policy.MaxFileSizeBytes = *override.MaxFileSizeBytes
	}
	if len(override.AllowedMIMETypes) > 0 {
		policy.AllowedMIMETypes = parseMIMEAllowlist(override.AllowedMIMETypes)
	}
	return policy
}

func (p Policy) AllowsMIME(mimeType string) bool {
	if len(p.AllowedMIMETypes) == 0 {
		return true
	}
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	_, ok := p.AllowedMIMETypes[mimeType]
	return ok
}

func ParseMIMEAllowlist(values []string) map[string]struct{} {
	return parseMIMEAllowlist(values)
}

func parseMIMEAllowlist(values []string) map[string]struct{} {
	allowlist := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		allowlist[value] = struct{}{}
	}
	return allowlist
}
