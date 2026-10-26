package post

import "time"

type ListQuery struct {
	Search   string `form:"q" binding:"max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=draft published"`
	AuthorID int64  `form:"author_id" binding:"omitempty,gt=0"`
}

type CreatePostRequest struct {
	Title string `json:"title" binding:"required,min=3,max=200"`
	Body  string `json:"body" binding:"max=100000"`
}

type UpdatePostRequest struct {
	Title *string `json:"title" binding:"omitempty,min=3,max=200"`
	Body  *string `json:"body" binding:"omitempty,max=100000"`
}

type PostResponse struct {
	ID          int64      `json:"id"`
	AuthorID    int64      `json:"author_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Body        string     `json:"body"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func toResponse(p *Post) PostResponse {
	return PostResponse{
		ID:          p.ID,
		AuthorID:    p.AuthorID,
		Title:       p.Title,
		Slug:        p.Slug,
		Body:        p.Body,
		Status:      p.Status,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toResponses(posts []Post) []PostResponse {
	out := make([]PostResponse, 0, len(posts))
	for i := range posts {
		out = append(out, toResponse(&posts[i]))
	}
	return out
}
