package rbac

import (
	"github.com/gin-gonic/gin"

	"backend-core-go/internal/shared/httpx"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toPermissionResponses(perms))
}

func (h *Handler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toRoleResponses(roles))
}

func (h *Handler) GetRole(c *gin.Context) {
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	role, err := h.svc.GetRole(c.Request.Context(), id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	res := toRoleResponse(&role.Role)
	res.Permissions = role.Permissions
	httpx.OK(c, res)
}

func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, toRoleResponse(role))
}

func (h *Handler) DeleteRole(c *gin.Context) {
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) SetRolePermissions(c *gin.Context) {
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	var req SetPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	if err := h.svc.SetRolePermissions(c.Request.Context(), id, req.Permissions); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) UserRoles(c *gin.Context) {
	userID, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	roles, err := h.svc.UserRoles(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toRoleResponses(roles))
}

func (h *Handler) AssignRole(c *gin.Context) {
	userID, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	if err := h.svc.AssignRole(c.Request.Context(), userID, req.RoleID); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) RevokeRole(c *gin.Context) {
	userID, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	roleID, ok := httpx.ParamID(c, "roleId")
	if !ok {
		return
	}
	if err := h.svc.RevokeRole(c.Request.Context(), userID, roleID); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}
