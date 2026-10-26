package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-core-go/internal/config"
	"backend-core-go/internal/features/auth"
	"backend-core-go/internal/features/product"
	"backend-core-go/internal/features/rbac"
	"backend-core-go/internal/features/user"
	"backend-core-go/internal/platform/database"
	"backend-core-go/internal/shared/httpx"
)

func registerRoutes(r *gin.Engine, cfg config.Config, db *pgxpool.Pool) {
	r.NoRoute(func(c *gin.Context) {
		httpx.Fail(c, http.StatusNotFound, "ROUTE_NOT_FOUND", "route not found")
	})
	r.GET("/healthz", healthHandler(db))

	api := r.Group("/api/v1")

	// Public
	tokens := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL, cfg.JWT.Issuer)
	auth.RegisterRoutes(api, db, tokens)

	// Protected: wajib login, otorisasi per endpoint lewat authz.Require(...)
	rbacSvc := rbac.NewService(rbac.NewRepository(db), database.NewTransactor(db), rbac.NewMemoryCache(cfg.RBACCacheTTL))
	authz := rbac.NewMiddleware(rbacSvc)

	protected := api.Group("", auth.Authenticate(tokens))
	rbac.RegisterRoutes(protected, rbacSvc, authz)
	user.RegisterRoutes(protected, db, authz)
	product.RegisterRoutes(protected, db, authz)
}

func healthHandler(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			httpx.Fail(c, http.StatusServiceUnavailable, "UNHEALTHY", "database unavailable")
			return
		}
		httpx.OK(c, gin.H{"status": "ok"})
	}
}
