package permissions

func CanManageFolder(role Role) bool {
	return role == RoleOwner || role == RoleAdmin
}
