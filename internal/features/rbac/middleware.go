package rbac

import (
	"context"

	"github.com/gin-gonic/gin"

	"backend-core-go/internal/shared/httpx"
)

type Middleware struct {
	svc Service
}

func NewMiddleware(svc Service) *Middleware {
	return &Middleware{svc: svc}
}

// Require lolos jika user punya SALAH SATU permission yang disebut.
// Harus dipasang setelah auth.Authenticate.
func (m *Middleware) Require(perms ...string) gin.HandlerFunc {
	if len(perms) == 0 {
		panic("rbac: Require called without permissions")
	}
	return func(c *gin.Context) {
		userID, ok := httpx.CurrentUserID(c)
		if !ok {
			return
		}
		allowed, err := m.svc.HasAny(c.Request.Context(), userID, perms...)
		if err != nil {
			httpx.Error(c, err)
			return
		}
		if !allowed {
			httpx.Error(c, ErrForbidden)
			return
		}
		c.Next()
	}
}

// HasPermission dipakai feature lain (lewat interface kecil di feature
// tersebut) untuk pengecekan permission di service, mis. ownership override.
func (m *Middleware) HasPermission(ctx context.Context, userID int64, perm string) (bool, error) {
	return m.svc.HasPermission(ctx, userID, perm)
}
