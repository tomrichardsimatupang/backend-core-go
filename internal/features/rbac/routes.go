package rbac

import "github.com/gin-gonic/gin"

// RegisterRoutes menerima Service (bukan pool) karena instance yang sama,
// beserta cache-nya, juga dipakai oleh Middleware.
func RegisterRoutes(rg *gin.RouterGroup, svc Service, authz *Middleware) {
	h := NewHandler(svc)

	g := rg.Group("/rbac")
	g.GET("/permissions", authz.Require(PermRoleRead), h.ListPermissions)

	g.GET("/roles", authz.Require(PermRoleRead), h.ListRoles)
	g.POST("/roles", authz.Require(PermRoleManage), h.CreateRole)
	g.GET("/roles/:id", authz.Require(PermRoleRead), h.GetRole)
	g.DELETE("/roles/:id", authz.Require(PermRoleManage), h.DeleteRole)
	g.PUT("/roles/:id/permissions", authz.Require(PermRoleManage), h.SetRolePermissions)

	g.GET("/users/:id/roles", authz.Require(PermRoleRead), h.UserRoles)
	g.POST("/users/:id/roles", authz.Require(PermRoleManage), h.AssignRole)
	g.DELETE("/users/:id/roles/:roleId", authz.Require(PermRoleManage), h.RevokeRole)
}
