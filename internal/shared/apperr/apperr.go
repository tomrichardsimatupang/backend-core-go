// Package apperr mendefinisikan tipe error domain yang dipakai lintas feature.
// Setiap feature membuat error-nya sendiri dengan New, lalu httpx memetakan
// Kind ke HTTP status.
package apperr

type Kind int

const (
	KindInternal Kind = iota
	KindInvalid
	KindUnauthenticated
	KindForbidden
	KindNotFound
	KindConflict
)

type Error struct {
	Kind    Kind
	Code    string
	Message string
}

func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

func (e *Error) Error() string {
	return e.Message
}
