// Command seed menyinkronkan permission & role default (make seed).
// Dengan flag -demo, juga membuat user dan artikel contoh (make seed-demo).
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"backend-core-go/internal/app"
	"backend-core-go/internal/config"
	"backend-core-go/internal/features/rbac"
	"backend-core-go/internal/platform/database"
	"backend-core-go/internal/platform/logger"
)

func main() {
	demo := flag.Bool("demo", false, "buat user & artikel contoh untuk kasus CMS/Blog")
	flag.Parse()

	if err := run(*demo); err != nil {
		slog.Error("seed failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(demo bool) error {
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
	if err := seeder.Sync(ctx, perms, app.DefaultRoles()); err != nil {
		return err
	}
	if cfg.SeedAdminEmail != "" {
		if _, err := seeder.AssignRole(ctx, cfg.SeedAdminEmail, rbac.SuperAdminRole); err != nil {
			return err
		}
	}
	if demo {
		return seedDemo(ctx, pool, seeder, log)
	}
	return nil
}
