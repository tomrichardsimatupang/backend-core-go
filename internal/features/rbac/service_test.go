package rbac

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend-core-go/internal/platform/database"
)

// mockRepo meng-embed Repository: method yang tidak di-override akan panic
// jika terpanggil, sehingga test gagal bila service memanggil hal tak terduga.
type mockRepo struct {
	Repository

	roles         map[int64]Role
	permissions   []Permission
	userPerms     map[int64][]string
	userPermCalls int
	replaced      map[int64][]string
	assigned      [][2]int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		roles: map[int64]Role{
			1: {ID: 1, Name: SuperAdminRole},
			2: {ID: 2, Name: "editor"},
		},
		permissions: []Permission{{ID: 1, Code: "product:read"}, {ID: 2, Code: "product:create"}},
		userPerms:   map[int64][]string{},
		replaced:    map[int64][]string{},
	}
}

func (m *mockRepo) WithDB(database.DBTX) Repository { return m }

func (m *mockRepo) FindRoleByID(_ context.Context, id int64) (*Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, ErrRoleNotFound
	}
	return &r, nil
}

func (m *mockRepo) ListPermissions(context.Context) ([]Permission, error) {
	return m.permissions, nil
}

func (m *mockRepo) ReplaceRolePermissions(_ context.Context, roleID int64, codes []string) error {
	m.replaced[roleID] = codes
	return nil
}

func (m *mockRepo) UserPermissions(_ context.Context, userID int64) ([]string, error) {
	m.userPermCalls++
	return m.userPerms[userID], nil
}

func (m *mockRepo) AssignRole(_ context.Context, userID, roleID int64) error {
	m.assigned = append(m.assigned, [2]int64{userID, roleID})
	return nil
}

func (m *mockRepo) DeleteRole(_ context.Context, id int64) error {
	delete(m.roles, id)
	return nil
}

type fakeTx struct{}

func (fakeTx) WithinTx(_ context.Context, fn func(database.DBTX) error) error { return fn(nil) }

func newTestService(repo *mockRepo) Service {
	return NewService(repo, fakeTx{}, NewMemoryCache(time.Minute))
}

func TestHasAny_MatchesAnyAndUsesCache(t *testing.T) {
	repo := newMockRepo()
	repo.userPerms[10] = []string{"product:read"}
	svc := newTestService(repo)
	ctx := context.Background()

	ok, err := svc.HasAny(ctx, 10, "product:manage", "product:read")
	if err != nil || !ok {
		t.Fatalf("HasAny = %v, %v; want true, nil", ok, err)
	}
	ok, err = svc.HasAny(ctx, 10, "product:manage")
	if err != nil || ok {
		t.Fatalf("HasAny = %v, %v; want false, nil", ok, err)
	}
	if repo.userPermCalls != 1 {
		t.Fatalf("repository called %d times, want 1 (cached)", repo.userPermCalls)
	}
}

func TestAssignRole_InvalidatesUserCache(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	ctx := context.Background()

	if ok, _ := svc.HasPermission(ctx, 10, "product:create"); ok {
		t.Fatal("user should not have permission yet")
	}
	repo.userPerms[10] = []string{"product:create"}
	if err := svc.AssignRole(ctx, 10, 2); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}
	if ok, _ := svc.HasPermission(ctx, 10, "product:create"); !ok {
		t.Fatal("permission should be reloaded after assign")
	}
}

func TestAssignRole_UnknownRole(t *testing.T) {
	svc := newTestService(newMockRepo())
	if err := svc.AssignRole(context.Background(), 10, 99); !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("err = %v, want ErrRoleNotFound", err)
	}
}

func TestSetRolePermissions(t *testing.T) {
	tests := []struct {
		name    string
		roleID  int64
		codes   []string
		wantErr error
	}{
		{"success", 2, []string{"product:read"}, nil},
		{"super_admin protected", 1, []string{"product:read"}, ErrProtectedRole},
		{"unknown permission", 2, []string{"product:fly"}, ErrUnknownPermission},
		{"role not found", 99, nil, ErrRoleNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			svc := newTestService(repo)
			err := svc.SetRolePermissions(context.Background(), tt.roleID, tt.codes)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			_, replaced := repo.replaced[tt.roleID]
			if replaced != (tt.wantErr == nil) {
				t.Fatalf("replaced = %v, want %v", replaced, tt.wantErr == nil)
			}
		})
	}
}

func TestSetRolePermissions_InvalidatesCache(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	ctx := context.Background()

	_, _ = svc.HasPermission(ctx, 10, "product:read")
	repo.userPerms[10] = []string{"product:read"}
	if err := svc.SetRolePermissions(ctx, 2, []string{"product:read"}); err != nil {
		t.Fatalf("SetRolePermissions: %v", err)
	}
	if ok, _ := svc.HasPermission(ctx, 10, "product:read"); !ok {
		t.Fatal("cache should be invalidated after role permissions change")
	}
}

func TestDeleteRole_SuperAdminProtected(t *testing.T) {
	svc := newTestService(newMockRepo())
	if err := svc.DeleteRole(context.Background(), 1); !errors.Is(err, ErrProtectedRole) {
		t.Fatalf("err = %v, want ErrProtectedRole", err)
	}
}
