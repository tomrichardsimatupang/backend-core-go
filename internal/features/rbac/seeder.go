package rbac

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"backend-core-go/internal/platform/database"
)

// Seeder menyinkronkan tabel permissions dengan konstanta di kode dan
// memastikan role super_admin memiliki semua permission.
type Seeder struct {
	repo Repository
	tx   database.Transactor
	log  *slog.Logger
}

func NewSeeder(repo Repository, tx database.Transactor, log *slog.Logger) *Seeder {
	return &Seeder{repo: repo, tx: tx, log: log}
}

// Sync menambah permission baru, menghapus permission yang tidak lagi ada
// di kode, dan memberi semua permission ke super_admin.
func (s *Seeder) Sync(ctx context.Context, perms []string) error {
	err := s.tx.WithinTx(ctx, func(db database.DBTX) error {
		repo := s.repo.WithDB(db)
		if err := repo.UpsertPermissions(ctx, perms); err != nil {
			return err
		}
		if err := repo.DeletePermissionsNotIn(ctx, perms); err != nil {
			return err
		}
		role, err := repo.UpsertRole(ctx, SuperAdminRole, "Full access to every permission")
		if err != nil {
			return err
		}
		return repo.ReplaceRolePermissions(ctx, role.ID, perms)
	})
	if err != nil {
		return fmt.Errorf("sync rbac: %w", err)
	}
	s.log.Info("rbac synced", slog.Int("permissions", len(perms)), slog.String("role", SuperAdminRole))
	return nil
}

// AssignSuperAdmin memberi role super_admin ke user dengan email tertentu.
func (s *Seeder) AssignSuperAdmin(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	userID, err := s.repo.FindUserIDByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find user %q: %w", email, err)
	}
	role, err := s.repo.FindRoleByName(ctx, SuperAdminRole)
	if err != nil {
		return fmt.Errorf("find role %q: %w", SuperAdminRole, err)
	}
	if err := s.repo.AssignRole(ctx, userID, role.ID); err != nil {
		return err
	}
	s.log.Info("super_admin assigned", slog.String("email", email), slog.Int64("user_id", userID))
	return nil
}
