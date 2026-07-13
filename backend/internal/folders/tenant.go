package folders

import "fifriz/internal/auth"

func IsPersonalTenant(tenantType auth.TenantType) bool {
	return tenantType == auth.TenantTypePersonal
}

func ShouldSeedDefaultFolders(tenantType auth.TenantType) bool {
	return IsPersonalTenant(tenantType)
}

func GetFolderDisplayName(tenantType auth.TenantType) string {
	if IsPersonalTenant(tenantType) {
		return "Kategori"
	}
	return "Klasor"
}
