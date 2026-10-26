package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterRoutes mendaftarkan endpoint publik autentikasi.
func RegisterRoutes(rg *gin.RouterGroup, db *pgxpool.Pool, tokens TokenManager) {
	h := NewHandler(NewService(NewRepository(db), tokens))

	g := rg.Group("/auth")
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
}
