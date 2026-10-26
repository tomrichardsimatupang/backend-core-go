package rbac

import "backend-core-go/internal/shared/apperr"

var (
	ErrForbidden         = apperr.New(apperr.KindForbidden, "FORBIDDEN", "you do not have permission to perform this action")
	ErrRoleNotFound      = apperr.New(apperr.KindNotFound, "ROLE_NOT_FOUND", "role not found")
	ErrRoleExists        = apperr.New(apperr.KindConflict, "ROLE_EXISTS", "role already exists")
	ErrProtectedRole     = apperr.New(apperr.KindForbidden, "PROTECTED_ROLE", "role super_admin is managed by the seeder")
	ErrUnknownPermission = apperr.New(apperr.KindInvalid, "UNKNOWN_PERMISSION", "one or more permissions do not exist")
	ErrUserNotFound      = apperr.New(apperr.KindNotFound, "USER_NOT_FOUND", "user not found")
)
