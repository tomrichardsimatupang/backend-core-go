package user

import "time"

type User struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// UpdateFields berisi kolom yang akan diubah; nil berarti tidak diubah.
type UpdateFields struct {
	Name  *string
	Email *string
}

func (f UpdateFields) IsEmpty() bool {
	return f.Name == nil && f.Email == nil
}
