package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"backend-core-go/internal/platform/logger"
	"backend-core-go/internal/shared/httpx"
)

// Recovery menangkap panic, mencatat stack trace, dan mengirim 500 generik.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.FromContext(c.Request.Context()).Error("panic recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				httpx.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			}
		}()
		c.Next()
	}
}
