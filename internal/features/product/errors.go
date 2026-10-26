package product

import "backend-core-go/internal/shared/apperr"

var (
	ErrProductNotFound = apperr.New(apperr.KindNotFound, "PRODUCT_NOT_FOUND", "product not found")
	ErrNotOwner        = apperr.New(apperr.KindForbidden, "PRODUCT_FORBIDDEN", "you can only modify your own products")
)
