package user

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"backend-core-go/internal/shared/pagination"
)

type mockRepo struct {
	users      map[int64]User
	hashes     map[int64]string
	total      int64
	updateArgs *UpdateFields
	deleted    []int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: map[int64]User{1: {ID: 1, Name: "Budi", Email: "budi@example.com"}}, hashes: map[int64]string{}}
}

func (m *mockRepo) List(context.Context, pagination.Params) ([]User, int64, error) {
	out := make([]User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, u)
	}
	return out, m.total, nil
}

func (m *mockRepo) FindByID(_ context.Context, id int64) (*User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &u, nil
}

func (m *mockRepo) Create(_ context.Context, u *User, hash string) error {
	u.ID = int64(len(m.users) + 1)
	m.users[u.ID] = *u
	m.hashes[u.ID] = hash
	return nil
}

func (m *mockRepo) Update(_ context.Context, id int64, fields UpdateFields) (*User, error) {
	m.updateArgs = &fields
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	if fields.Name != nil {
		u.Name = *fields.Name
	}
	if fields.Email != nil {
		u.Email = *fields.Email
	}
	m.users[id] = u
	return &u, nil
}

func (m *mockRepo) Delete(_ context.Context, id int64) error {
	if _, ok := m.users[id]; !ok {
		return ErrUserNotFound
	}
	m.deleted = append(m.deleted, id)
	return nil
}

func TestCreate_NormalizesAndHashes(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	u, err := svc.Create(context.Background(), CreateUserRequest{Name: " Ani ", Email: " ANI@Example.com", Password: "secret123"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.Name != "Ani" || u.Email != "ani@example.com" {
		t.Fatalf("got name=%q email=%q", u.Name, u.Email)
	}
	if bcrypt.CompareHashAndPassword([]byte(repo.hashes[u.ID]), []byte("secret123")) != nil {
		t.Fatal("password should be bcrypt hashed")
	}
}

func TestUpdate_EmptyRequestDoesNotWrite(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	u, err := svc.Update(context.Background(), 1, UpdateUserRequest{})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updateArgs != nil {
		t.Fatal("repository Update should not be called for empty request")
	}
	if u.ID != 1 {
		t.Fatalf("got user %d, want 1", u.ID)
	}
}

func TestUpdate_NormalizesEmail(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	email := " New@Example.COM "

	u, err := svc.Update(context.Background(), 1, UpdateUserRequest{Email: &email})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if u.Email != "new@example.com" || repo.updateArgs.Name != nil {
		t.Fatalf("unexpected update: email=%q name=%v", u.Email, repo.updateArgs.Name)
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		actorID, id int64
		wantErr     error
	}{
		{"success", 2, 1, nil},
		{"cannot delete self", 1, 1, ErrCannotDeleteSelf},
		{"not found", 1, 99, ErrUserNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewService(newMockRepo()).Delete(context.Background(), tt.actorID, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestList_BuildsMeta(t *testing.T) {
	repo := newMockRepo()
	repo.total = 45
	_, meta, err := NewService(repo).List(context.Background(), pagination.New(2, 20))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if meta.Total != 45 || meta.TotalPages != 3 || meta.Page != 2 {
		t.Fatalf("unexpected meta %+v", meta)
	}
}
