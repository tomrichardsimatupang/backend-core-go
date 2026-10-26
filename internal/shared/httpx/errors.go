package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-core-go/internal/platform/logger"
	"backend-core-go/internal/shared/apperr"
)

// statusByKind memetakan jenis error domain ke HTTP status.
var statusByKind = map[apperr.Kind]int{
	apperr.KindInvalid:         http.StatusUnprocessableEntity,
	apperr.KindUnauthenticated: http.StatusUnauthorized,
	apperr.KindForbidden:       http.StatusForbidden,
	apperr.KindNotFound:        http.StatusNotFound,
	apperr.KindConflict:        http.StatusConflict,
}

// Error memetakan error dari service ke response HTTP. Error yang bukan
// error domain dilog di server dan dikirim ke client sebagai 500 generik.
func Error(c *gin.Context, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		if status, ok := statusByKind[appErr.Kind]; ok {
			Fail(c, status, appErr.Code, appErr.Message)
			return
		}
	}

	logger.FromContext(c.Request.Context()).Error("unhandled error",
		slog.String("method", c.Request.Method),
		slog.String("route", c.FullPath()),
		slog.Any("error", err),
	)
	Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
