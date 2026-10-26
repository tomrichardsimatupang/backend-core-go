package post

const (
	PermRead    = "post:read"    // baca post published + draft milik sendiri
	PermCreate  = "post:create"  // tulis post (selalu mulai sebagai draft)
	PermUpdate  = "post:update"  // ubah post milik sendiri
	PermDelete  = "post:delete"  // hapus post milik sendiri
	PermPublish = "post:publish" // publish/unpublish post siapa pun
	PermManage  = "post:manage"  // lihat semua draft, ubah/hapus post siapa pun
)

var Permissions = []string{PermRead, PermCreate, PermUpdate, PermDelete, PermPublish, PermManage}
