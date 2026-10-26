package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-core-go/internal/features/rbac"
)

func RegisterRoutes(rg *gin.RouterGroup, db *pgxpool.Pool, authz *rbac.Middleware) {
	h := NewHandler(NewService(NewRepository(db)))

	g := rg.Group("/users")
	// Data milik sendiri: cukup login.
	g.GET("/me", h.Me)
	g.PUT("/me", h.UpdateMe)

	g.GET("", authz.Require(PermRead), h.List)
	g.POST("", authz.Require(PermCreate), h.Create)
	g.GET("/:id", authz.Require(PermRead), h.Get)
	g.PUT("/:id", authz.Require(PermUpdate), h.Update)
	g.DELETE("/:id", authz.Require(PermDelete), h.Delete)
}
