package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend-core-go/internal/app"
	"backend-core-go/internal/features/post"
	"backend-core-go/internal/features/rbac"
	"backend-core-go/internal/features/user"
	"backend-core-go/internal/platform/database"
)

// demoPassword dipakai semua user contoh. Hanya untuk development.
const demoPassword = "password123"

type demoUser struct {
	Name  string
	Email string
	Role  string
}

var demoUsers = []demoUser{
	{"Super Admin", "superadmin@example.com", rbac.SuperAdminRole},
	{"Andi Admin", "admin@example.com", app.RoleAdmin},
	{"Eka Editor", "editor@example.com", app.RoleEditor},
	{"Ayu Author", "author@example.com", app.RoleAuthor},
	{"Sari Subscriber", "subscriber@example.com", app.RoleSubscriber},
}

type demoPost struct {
	AuthorEmail string
	Title       string
	Body        string
	Publish     bool
}

var demoPosts = []demoPost{
	{
		AuthorEmail: "editor@example.com",
		Title:       "Selamat Datang di Blog Kami",
		Body:        "Ini adalah artikel pertama di blog. Semua user yang login bisa membacanya.",
		Publish:     true,
	},
	{
		AuthorEmail: "author@example.com",
		Title:       "Belajar Go untuk Pemula",
		Body:        "Ditulis oleh author, lalu direview dan dipublish oleh editor.",
		Publish:     true,
	},
	{
		AuthorEmail: "author@example.com",
		Title:       "Tips Query Builder dengan Squirrel",
		Body:        "Masih draft: hanya terlihat oleh penulisnya, editor, admin, dan super admin.",
		Publish:     false,
	},
}

// seedDemo membuat user & artikel contoh. Aman dijalankan berulang kali:
// user/artikel yang sudah ada dilewati.
func seedDemo(ctx context.Context, pool *pgxpool.Pool, seeder *rbac.Seeder, log *slog.Logger) error {
	users := user.NewService(user.NewRepository(pool))
	userIDs := make(map[string]int64, len(demoUsers))

	for _, du := range demoUsers {
		_, err := users.Create(ctx, user.CreateUserRequest{Name: du.Name, Email: du.Email, Password: demoPassword})
		switch {
		case errors.Is(err, user.ErrEmailTaken):
			log.Info("demo user already exists", slog.String("email", du.Email))
		case err != nil:
			return fmt.Errorf("create demo user %s: %w", du.Email, err)
		}
		id, err := seeder.AssignRole(ctx, du.Email, du.Role)
		if err != nil {
			return err
		}
		userIDs[du.Email] = id
	}

	rbacSvc := rbac.NewService(rbac.NewRepository(pool), database.NewTransactor(pool), rbac.NewMemoryCache(0))
	posts := post.NewService(post.NewRepository(pool), rbac.NewMiddleware(rbacSvc))

	for _, dp := range demoPosts {
		p, err := posts.Create(ctx, userIDs[dp.AuthorEmail], post.CreatePostRequest{Title: dp.Title, Body: dp.Body})
		if errors.Is(err, post.ErrSlugTaken) {
			log.Info("demo post already exists", slog.String("title", dp.Title))
			continue
		}
		if err != nil {
			return fmt.Errorf("create demo post %q: %w", dp.Title, err)
		}
		if dp.Publish {
			if _, err := posts.Publish(ctx, p.ID); err != nil {
				return fmt.Errorf("publish demo post %q: %w", dp.Title, err)
			}
		}
		log.Info("demo post created", slog.String("title", dp.Title), slog.Bool("published", dp.Publish))
	}

	log.Info("demo data ready", slog.Int("users", len(demoUsers)), slog.String("password", demoPassword))
	return nil
}
