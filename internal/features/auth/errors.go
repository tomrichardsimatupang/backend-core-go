package auth

import (
	"errors"

	"backend-core-go/internal/shared/apperr"
)

var (
	ErrInvalidCredentials = apperr.New(apperr.KindUnauthenticated, "INVALID_CREDENTIALS", "invalid email or password")
	ErrEmailTaken         = apperr.New(apperr.KindConflict, "EMAIL_TAKEN", "email already registered")
	ErrMissingToken       = apperr.New(apperr.KindUnauthenticated, "MISSING_TOKEN", "missing bearer token")
	ErrInvalidToken       = apperr.New(apperr.KindUnauthenticated, "INVALID_TOKEN", "invalid or expired token")

	// errCredentialsNotFound tidak dikirim ke client; service mengubahnya
	// menjadi ErrInvalidCredentials agar tidak membocorkan email terdaftar.
	errCredentialsNotFound = errors.New("credentials not found")
)
