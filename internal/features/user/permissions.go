package user

const (
	PermRead   = "user:read"
	PermCreate = "user:create"
	PermUpdate = "user:update"
	PermDelete = "user:delete"
)

var Permissions = []string{PermRead, PermCreate, PermUpdate, PermDelete}
