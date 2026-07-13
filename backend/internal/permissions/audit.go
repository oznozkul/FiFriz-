package permissions

func CanViewAuditLogs(role Role) bool {
	return role == RoleOwner || role == RoleAdmin
}

func CanViewAuditLogsRole(role string) bool {
	return role == string(RoleOwner) || role == string(RoleAdmin)
}
