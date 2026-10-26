package post

import "time"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

type Post struct {
	ID          int64      `db:"id"`
	AuthorID    int64      `db:"author_id"`
	Title       string     `db:"title"`
	Slug        string     `db:"slug"`
	Body        string     `db:"body"`
	Status      string     `db:"status"`
	PublishedAt *time.Time `db:"published_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type ListFilter struct {
	Search   string
	Status   string
	AuthorID int64
	// VisibleTo != 0 membatasi hasil ke post published + draft milik user ini.
	VisibleTo int64
}

// UpdateFields berisi kolom yang akan diubah; nil berarti tidak diubah.
// Slug sengaja tidak ikut berubah saat judul diubah agar URL tetap stabil.
type UpdateFields struct {
	Title *string
	Body  *string
}

func (f UpdateFields) IsEmpty() bool {
	return f.Title == nil && f.Body == nil
}
