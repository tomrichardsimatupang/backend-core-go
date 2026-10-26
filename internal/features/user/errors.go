package user

import "backend-core-go/internal/shared/apperr"

var (
	ErrUserNotFound     = apperr.New(apperr.KindNotFound, "USER_NOT_FOUND", "user not found")
	ErrEmailTaken       = apperr.New(apperr.KindConflict, "EMAIL_TAKEN", "email already registered")
	ErrCannotDeleteSelf = apperr.New(apperr.KindForbidden, "CANNOT_DELETE_SELF", "you cannot delete your own account")
)
