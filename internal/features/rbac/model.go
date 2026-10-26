package rbac

import "time"

// SuperAdminRole selalu mendapat semua permission melalui seeder.
const SuperAdminRole = "super_admin"

type Role struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

type Permission struct {
	ID          int64  `db:"id"`
	Code        string `db:"code"`
	Description string `db:"description"`
}

type RoleWithPermissions struct {
	Role
	Permissions []string
}
