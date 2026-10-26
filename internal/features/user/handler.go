package user

import (
	"github.com/gin-gonic/gin"

	"backend-core-go/internal/shared/httpx"
	"backend-core-go/internal/shared/pagination"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := httpx.CurrentUserID(c)
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toResponse(u))
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID, ok := httpx.CurrentUserID(c)
	if !ok {
		return
	}
	h.update(c, userID)
}

func (h *Handler) List(c *gin.Context) {
	users, meta, err := h.svc.List(c.Request.Context(), pagination.FromQuery(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OKWithMeta(c, toResponses(users), meta)
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toResponse(u))
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	u, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, toResponse(u))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	h.update(c, id)
}

func (h *Handler) Delete(c *gin.Context) {
	actorID, ok := httpx.CurrentUserID(c)
	if !ok {
		return
	}
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), actorID, id); err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) update(c *gin.Context, id int64) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	u, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toResponse(u))
}
