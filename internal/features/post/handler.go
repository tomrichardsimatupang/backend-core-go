package post

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

func (h *Handler) List(c *gin.Context) {
	actorID, ok := httpx.CurrentUserID(c)
	if !ok {
		return
	}
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	posts, meta, err := h.svc.List(c.Request.Context(), actorID, q, pagination.FromQuery(c))
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OKWithMeta(c, toResponses(posts), meta)
}

func (h *Handler) Get(c *gin.Context) {
	actorID, ok := httpx.CurrentUserID(c)
	if !ok {
		return
	}
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.Get(c.Request.Context(), actorID, id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toResponse(p))
}

func (h *Handler) Create(c *gin.Context) {
	actorID, ok := httpx.CurrentUserID(c)
	if !ok {
		return
	}
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	p, err := h.svc.Create(c.Request.Context(), actorID, req)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, toResponse(p))
}

func (h *Handler) Update(c *gin.Context) {
	actorID, ok := httpx.CurrentUserID(c)
	if !ok {
		return
	}
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	p, err := h.svc.Update(c.Request.Context(), actorID, id, req)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toResponse(p))
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

func (h *Handler) Publish(c *gin.Context) {
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.Publish(c.Request.Context(), id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toResponse(p))
}

func (h *Handler) Unpublish(c *gin.Context) {
	id, ok := httpx.ParamID(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.Unpublish(c.Request.Context(), id)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, toResponse(p))
}
