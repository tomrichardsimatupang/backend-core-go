package auth

type Credentials struct {
	ID           int64  `db:"id"`
	PasswordHash string `db:"password_hash"`
}
