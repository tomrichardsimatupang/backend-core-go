package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type Params struct {
	Page    int
	PerPage int
}

// FromQuery membaca ?page= dan ?per_page=; nilai tidak valid diganti default.
func FromQuery(c *gin.Context) Params {
	return New(atoi(c.Query("page")), atoi(c.Query("per_page")))
}

func New(page, perPage int) Params {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return Params{Page: page, PerPage: perPage}
}

func (p Params) Limit() uint64 {
	return uint64(p.PerPage) //nolint:gosec // PerPage sudah di-clamp ke 1..MaxPerPage
}

func (p Params) Offset() uint64 {
	return uint64((p.Page - 1) * p.PerPage) //nolint:gosec // Page >= 1
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func NewMeta(p Params, total int64) Meta {
	pages := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	return Meta{Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: pages}
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
