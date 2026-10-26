package app

import (
	"backend-core-go/internal/features/post"
	"backend-core-go/internal/features/product"
	"backend-core-go/internal/features/rbac"
	"backend-core-go/internal/features/user"
)

// Permissions mengembalikan seluruh permission aplikasi untuk seeder.
// Setiap feature baru WAJIB menambahkan slice Permissions-nya di sini.
func Permissions() ([]string, error) {
	return rbac.AllPermissions(
		rbac.Permissions,
		user.Permissions,
		product.Permissions,
		post.Permissions,
	)
}
