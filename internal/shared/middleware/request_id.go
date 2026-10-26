package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	"github.com/gin-gonic/gin"

	"backend-core-go/internal/platform/logger"
)

const HeaderRequestID = "X-Request-ID"

const maxRequestIDLength = 64

// RequestID memakai X-Request-ID dari client (atau membuat baru), lalu
// menyimpan logger ber-request_id ke context request.
func RequestID(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" || len(id) > maxRequestIDLength {
			id = newRequestID()
		}
		c.Header(HeaderRequestID, id)

		ctx := logger.WithContext(c.Request.Context(), base.With(slog.String("request_id", id)))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
