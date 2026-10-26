package post

import "backend-core-go/internal/shared/apperr"

var (
	ErrPostNotFound = apperr.New(apperr.KindNotFound, "POST_NOT_FOUND", "post not found")
	ErrSlugTaken    = apperr.New(apperr.KindConflict, "SLUG_TAKEN", "a post with the same title already exists")
	ErrInvalidTitle = apperr.New(apperr.KindInvalid, "INVALID_TITLE", "title must contain letters or digits")
	ErrNotAuthor    = apperr.New(apperr.KindForbidden, "POST_FORBIDDEN", "you can only modify your own posts")
)
