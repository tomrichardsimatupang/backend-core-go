package rbac

const (
	PermRoleRead   = "role:read"
	PermRoleManage = "role:manage"
)

var Permissions = []string{PermRoleRead, PermRoleManage}
