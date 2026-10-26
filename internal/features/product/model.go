package product

import "time"

type Product struct {
	ID          int64     `db:"id"`
	OwnerID     int64     `db:"owner_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       int64     `db:"price"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ListFilter struct {
	Search  string
	OwnerID int64
}

// UpdateFields berisi kolom yang akan diubah; nil berarti tidak diubah.
type UpdateFields struct {
	Name        *string
	Description *string
	Price       *int64
}

func (f UpdateFields) IsEmpty() bool {
	return f.Name == nil && f.Description == nil && f.Price == nil
}
