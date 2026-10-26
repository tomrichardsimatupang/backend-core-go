package post

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-core-go/internal/features/rbac"
)

func RegisterRoutes(rg *gin.RouterGroup, db *pgxpool.Pool, authz *rbac.Middleware) {
	h := NewHandler(NewService(NewRepository(db), authz))

	g := rg.Group("/posts")
	g.GET("", authz.Require(PermRead), h.List)
	g.POST("", authz.Require(PermCreate), h.Create)
	g.GET("/:id", authz.Require(PermRead), h.Get)
	// Ownership dicek di service: penulis cukup post:update/delete,
	// selain penulis butuh post:manage.
	g.PUT("/:id", authz.Require(PermUpdate, PermManage), h.Update)
	g.DELETE("/:id", authz.Require(PermDelete, PermManage), h.Delete)
	g.POST("/:id/publish", authz.Require(PermPublish), h.Publish)
	g.POST("/:id/unpublish", authz.Require(PermPublish), h.Unpublish)
}
