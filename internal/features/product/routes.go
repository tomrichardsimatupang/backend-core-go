package product

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-core-go/internal/features/rbac"
)

func RegisterRoutes(rg *gin.RouterGroup, db *pgxpool.Pool, authz *rbac.Middleware) {
	h := NewHandler(NewService(NewRepository(db), authz))

	g := rg.Group("/products")
	g.GET("", authz.Require(PermRead), h.List)
	g.POST("", authz.Require(PermCreate), h.Create)
	g.GET("/:id", authz.Require(PermRead), h.Get)
	// Ownership dicek di service: pemilik cukup product:update/delete,
	// selain pemilik butuh product:manage.
	g.PUT("/:id", authz.Require(PermUpdate, PermManage), h.Update)
	g.DELETE("/:id", authz.Require(PermDelete, PermManage), h.Delete)
}
