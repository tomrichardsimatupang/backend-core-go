package app

import (
	"backend-core-go/internal/features/post"
	"backend-core-go/internal/features/product"
	"backend-core-go/internal/features/rbac"
	"backend-core-go/internal/features/user"
)

// Nama role default untuk kasus CMS/Blog.
const (
	RoleAdmin      = "admin"
	RoleEditor     = "editor"
	RoleAuthor     = "author"
	RoleSubscriber = "subscriber"
)

// DefaultRoles adalah role yang dikelola kode dan disinkronkan oleh seeder
// (super_admin ditambahkan otomatis oleh rbac.Seeder). Perubahan permission
// role ini lewat API akan ditimpa kembali saat `make seed` dijalankan.
func DefaultRoles() []rbac.RoleSeed {
	return []rbac.RoleSeed{
		{
			Name:        RoleAdmin,
			Description: "Kelola user dan seluruh konten; tidak bisa mengubah role",
			Permissions: concat(
				user.Permissions,
				post.Permissions,
				product.Permissions,
				[]string{rbac.PermRoleRead},
			),
		},
		{
			Name:        RoleEditor,
			Description: "Review, edit, dan publish artikel siapa pun",
			Permissions: []string{
				post.PermRead, post.PermCreate, post.PermUpdate, post.PermDelete,
				post.PermPublish, post.PermManage,
			},
		},
		{
			Name:        RoleAuthor,
			Description: "Menulis dan mengelola draft artikel milik sendiri",
			Permissions: []string{post.PermRead, post.PermCreate, post.PermUpdate, post.PermDelete},
		},
		{
			Name:        RoleSubscriber,
			Description: "Hanya membaca artikel yang sudah dipublish",
			Permissions: []string{post.PermRead},
		},
	}
}

func concat(groups ...[]string) []string {
	var out []string
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}
