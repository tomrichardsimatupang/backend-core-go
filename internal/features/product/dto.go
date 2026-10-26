package product

import "time"

type ListQuery struct {
	Search  string `form:"q" binding:"max=100"`
	OwnerID int64  `form:"owner_id" binding:"omitempty,gt=0"`
}

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=2000"`
	Price       *int64 `json:"price" binding:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=2000"`
	Price       *int64  `json:"price" binding:"omitempty,gte=0"`
}

type ProductResponse struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toResponse(p *Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		OwnerID:     p.OwnerID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toResponses(products []Product) []ProductResponse {
	out := make([]ProductResponse, 0, len(products))
	for i := range products {
		out = append(out, toResponse(&products[i]))
	}
	return out
}
