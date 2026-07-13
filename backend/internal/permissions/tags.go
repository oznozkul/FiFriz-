package permissions

func CanManageTags(role Role) bool {
	return role == RoleOwner || role == RoleAdmin
}

func CanCreateTags(role Role) bool {
	return role == RoleOwner || role == RoleAdmin || role == RoleMember
}
