package auth

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"backend-core-go/internal/platform/logger"
	"backend-core-go/internal/shared/authctx"
	"backend-core-go/internal/shared/httpx"
)

// Authenticate memverifikasi bearer token dan menyimpan user_id ke context.
func Authenticate(tokens TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			httpx.Error(c, ErrMissingToken)
			return
		}

		userID, err := tokens.Parse(token)
		if err != nil {
			httpx.Error(c, err)
			return
		}

		ctx := authctx.SetUserID(c.Request.Context(), userID)
		ctx = logger.WithContext(ctx, logger.FromContext(ctx).With(slog.Int64("user_id", userID)))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
