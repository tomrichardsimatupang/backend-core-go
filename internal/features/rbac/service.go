package rbac

import (
	"context"
	"fmt"
	"strings"

	"backend-core-go/internal/platform/database"
)

type Service interface {
	HasPermission(ctx context.Context, userID int64, perm string) (bool, error)
	HasAny(ctx context.Context, userID int64, perms ...string) (bool, error)

	ListPermissions(ctx context.Context) ([]Permission, error)
	ListRoles(ctx context.Context) ([]Role, error)
	GetRole(ctx context.Context, id int64) (*RoleWithPermissions, error)
	CreateRole(ctx context.Context, req CreateRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, id int64) error
	SetRolePermissions(ctx context.Context, roleID int64, codes []string) error

	UserRoles(ctx context.Context, userID int64) ([]Role, error)
	AssignRole(ctx context.Context, userID, roleID int64) error
	RevokeRole(ctx context.Context, userID, roleID int64) error
}

type service struct {
	repo  Repository
	tx    database.Transactor
	cache Cache
}

func NewService(repo Repository, tx database.Transactor, cache Cache) Service {
	return &service{repo: repo, tx: tx, cache: cache}
}

func (s *service) HasPermission(ctx context.Context, userID int64, perm string) (bool, error) {
	return s.HasAny(ctx, userID, perm)
}

func (s *service) HasAny(ctx context.Context, userID int64, perms ...string) (bool, error) {
	set, err := s.permissionsOf(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if set.Has(p) {
			return true, nil
		}
	}
	return false, nil
}

func (s *service) permissionsOf(ctx context.Context, userID int64) (PermissionSet, error) {
	if set, ok := s.cache.Get(userID); ok {
		return set, nil
	}
	codes, err := s.repo.UserPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load permissions of user %d: %w", userID, err)
	}
	set := NewPermissionSet(codes...)
	s.cache.Set(userID, set)
	return set, nil
}

func (s *service) ListPermissions(ctx context.Context) ([]Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *service) ListRoles(ctx context.Context) ([]Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *service) GetRole(ctx context.Context, id int64) (*RoleWithPermissions, error) {
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	perms, err := s.repo.RolePermissions(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RoleWithPermissions{Role: *role, Permissions: perms}, nil
}

func (s *service) CreateRole(ctx context.Context, req CreateRoleRequest) (*Role, error) {
	role := &Role{
		Name:        strings.ToLower(strings.TrimSpace(req.Name)),
		Description: strings.TrimSpace(req.Description),
	}
	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *service) DeleteRole(ctx context.Context, id int64) error {
	if err := s.ensureMutable(ctx, id); err != nil {
		return err
	}
	if err := s.repo.DeleteRole(ctx, id); err != nil {
		return err
	}
	// user_roles ikut terhapus (cascade); user mana saja bisa terdampak.
	s.cache.InvalidateAll()
	return nil
}

func (s *service) SetRolePermissions(ctx context.Context, roleID int64, codes []string) error {
	if err := s.ensureMutable(ctx, roleID); err != nil {
		return err
	}

	known, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return err
	}
	knownSet := make(PermissionSet, len(known))
	for _, p := range known {
		knownSet[p.Code] = struct{}{}
	}
	for _, code := range codes {
		if !knownSet.Has(code) {
			return fmt.Errorf("permission %q: %w", code, ErrUnknownPermission)
		}
	}

	err = s.tx.WithinTx(ctx, func(db database.DBTX) error {
		return s.repo.WithDB(db).ReplaceRolePermissions(ctx, roleID, codes)
	})
	if err != nil {
		return fmt.Errorf("set permissions of role %d: %w", roleID, err)
	}
	s.cache.InvalidateAll()
	return nil
}

func (s *service) UserRoles(ctx context.Context, userID int64) ([]Role, error) {
	return s.repo.UserRoles(ctx, userID)
}

func (s *service) AssignRole(ctx context.Context, userID, roleID int64) error {
	if _, err := s.repo.FindRoleByID(ctx, roleID); err != nil {
		return err
	}
	if err := s.repo.AssignRole(ctx, userID, roleID); err != nil {
		return err
	}
	s.cache.Invalidate(userID)
	return nil
}

func (s *service) RevokeRole(ctx context.Context, userID, roleID int64) error {
	if err := s.repo.RevokeRole(ctx, userID, roleID); err != nil {
		return err
	}
	s.cache.Invalidate(userID)
	return nil
}

// ensureMutable menolak perubahan pada role super_admin, yang dikelola seeder.
func (s *service) ensureMutable(ctx context.Context, roleID int64) error {
	role, err := s.repo.FindRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role.Name == SuperAdminRole {
		return ErrProtectedRole
	}
	return nil
}
