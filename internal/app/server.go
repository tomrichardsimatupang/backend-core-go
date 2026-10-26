package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-core-go/internal/config"
	"backend-core-go/internal/shared/middleware"
)

type Server struct {
	http            *http.Server
	log             *slog.Logger
	shutdownTimeout time.Duration
}

func NewServer(cfg config.Config, log *slog.Logger, db *pgxpool.Pool) *Server {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	useJSONFieldNamesInValidation()

	r := gin.New()
	r.Use(middleware.RequestID(log), middleware.Logger(), middleware.Recovery())
	registerRoutes(r, cfg, db)

	return &Server{
		http: &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		log:             log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Run menjalankan server sampai ctx dibatalkan, lalu melakukan graceful shutdown.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("http server listening", slog.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	s.log.Info("shutting down http server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}

// useJSONFieldNamesInValidation membuat error validasi memakai nama field
// JSON/form (mis. "email") alih-alih nama field struct ("Email").
func useJSONFieldNamesInValidation() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		for _, tag := range []string{"json", "form"} {
			name, _, _ := strings.Cut(f.Tag.Get(tag), ",")
			if name != "" && name != "-" {
				return name
			}
		}
		return f.Name
	})
}
