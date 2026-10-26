package auth

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

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	res, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Created(c, res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err)
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.OK(c, res)
}
