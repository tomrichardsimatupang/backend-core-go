// Command seed menyinkronkan permission & role default (make seed).
package main

import (
	"context"
	"log/slog"
	"os"

	"backend-core-go/internal/app"
	"backend-core-go/internal/config"
	"backend-core-go/internal/features/rbac"
	"backend-core-go/internal/platform/database"
	"backend-core-go/internal/platform/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.LogLevel, cfg.LogFormat, os.Stdout)
	slog.SetDefault(log)

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	perms, err := app.Permissions()
	if err != nil {
		return err
	}

	seeder := rbac.NewSeeder(rbac.NewRepository(pool), database.NewTransactor(pool), log)
	if err := seeder.Sync(ctx, perms); err != nil {
		return err
	}
	if cfg.SeedAdminEmail != "" {
		return seeder.AssignSuperAdmin(ctx, cfg.SeedAdminEmail)
	}
	return nil
}
