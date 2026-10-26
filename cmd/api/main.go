package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"backend-core-go/internal/app"
	"backend-core-go/internal/config"
	"backend-core-go/internal/platform/database"
	"backend-core-go/internal/platform/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.ValidateJWT(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat, os.Stdout)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	return app.NewServer(cfg, log, pool).Run(ctx)
}
