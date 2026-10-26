package product

const (
	PermRead   = "product:read"
	PermCreate = "product:create"
	PermUpdate = "product:update" // ubah produk milik sendiri
	PermDelete = "product:delete" // hapus produk milik sendiri
	PermManage = "product:manage" // ubah/hapus produk milik siapa pun
)

var Permissions = []string{PermRead, PermCreate, PermUpdate, PermDelete, PermManage}
