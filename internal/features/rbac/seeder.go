package rbac

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"backend-core-go/internal/platform/database"
)

// Seeder menyinkronkan tabel permissions dengan konstanta di kode,
// memastikan super_admin memiliki semua permission, dan menyinkronkan
// role default.
type Seeder struct {
	repo Repository
	tx   database.Transactor
	log  *slog.Logger
}

func NewSeeder(repo Repository, tx database.Transactor, log *slog.Logger) *Seeder {
	return &Seeder{repo: repo, tx: tx, log: log}
}

// Sync menambah permission baru, menghapus permission yang tidak lagi ada
// di kode, memberi semua permission ke super_admin, lalu menyamakan
// permission setiap role default dengan definisinya di kode.
func (s *Seeder) Sync(ctx context.Context, perms []string, roles []RoleSeed) error {
	if err := validateRoleSeeds(perms, roles); err != nil {
		return err
	}

	err := s.tx.WithinTx(ctx, func(db database.DBTX) error {
		repo := s.repo.WithDB(db)
		if err := repo.UpsertPermissions(ctx, perms); err != nil {
			return err
		}
		if err := repo.DeletePermissionsNotIn(ctx, perms); err != nil {
			return err
		}

		all := append([]RoleSeed{{Name: SuperAdminRole, Description: "Full access to every permission", Permissions: perms}}, roles...)
		for _, seed := range all {
			role, err := repo.UpsertRole(ctx, seed.Name, seed.Description)
			if err != nil {
				return err
			}
			if err := repo.ReplaceRolePermissions(ctx, role.ID, seed.Permissions); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("sync rbac: %w", err)
	}
	s.log.Info("rbac synced", slog.Int("permissions", len(perms)), slog.Int("roles", len(roles)+1))
	return nil
}

// AssignRole memberi role ke user berdasarkan email dan mengembalikan ID user.
func (s *Seeder) AssignRole(ctx context.Context, email, roleName string) (int64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	userID, err := s.repo.FindUserIDByEmail(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("find user %q: %w", email, err)
	}
	role, err := s.repo.FindRoleByName(ctx, roleName)
	if err != nil {
		return 0, fmt.Errorf("find role %q: %w", roleName, err)
	}
	if err := s.repo.AssignRole(ctx, userID, role.ID); err != nil {
		return 0, err
	}
	s.log.Info("role assigned", slog.String("email", email), slog.String("role", roleName), slog.Int64("user_id", userID))
	return userID, nil
}

func validateRoleSeeds(perms []string, roles []RoleSeed) error {
	known := NewPermissionSet(perms...)
	seen := map[string]bool{SuperAdminRole: true}
	for _, r := range roles {
		if seen[r.Name] {
			return fmt.Errorf("role seed %q: duplicate or reserved name", r.Name)
		}
		seen[r.Name] = true
		for _, p := range r.Permissions {
			if !known.Has(p) {
				return fmt.Errorf("role seed %q: unknown permission %q", r.Name, p)
			}
		}
	}
	return nil
}
