package product

import (
	"context"
	"errors"
	"testing"

	"backend-core-go/internal/shared/pagination"
)

type mockRepo struct {
	products map[int64]Product
	updated  bool
	deleted  bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{products: map[int64]Product{1: {ID: 1, OwnerID: 10, Name: "Kopi", Price: 25000}}}
}

func (m *mockRepo) List(context.Context, ListFilter, pagination.Params) ([]Product, int64, error) {
	return nil, 0, nil
}

func (m *mockRepo) FindByID(_ context.Context, id int64) (*Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	return &p, nil
}

func (m *mockRepo) Create(_ context.Context, p *Product) error {
	p.ID = int64(len(m.products) + 1)
	m.products[p.ID] = *p
	return nil
}

func (m *mockRepo) Update(_ context.Context, id int64, fields UpdateFields) (*Product, error) {
	m.updated = true
	p := m.products[id]
	if fields.Name != nil {
		p.Name = *fields.Name
	}
	if fields.Price != nil {
		p.Price = *fields.Price
	}
	return &p, nil
}

func (m *mockRepo) Delete(context.Context, int64) error {
	m.deleted = true
	return nil
}

type mockAuthz struct {
	perms map[int64][]string
}

func (a mockAuthz) HasPermission(_ context.Context, userID int64, perm string) (bool, error) {
	for _, p := range a.perms[userID] {
		if p == perm {
			return true, nil
		}
	}
	return false, nil
}

func TestCreate_SetsOwnerFromActor(t *testing.T) {
	repo := newMockRepo()
	price := int64(5000)
	p, err := NewService(repo, mockAuthz{}).Create(context.Background(), 42, CreateProductRequest{Name: " Teh ", Price: &price})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.OwnerID != 42 || p.Name != "Teh" {
		t.Fatalf("got owner=%d name=%q", p.OwnerID, p.Name)
	}
}

func TestUpdate_Ownership(t *testing.T) {
	authz := mockAuthz{perms: map[int64][]string{30: {PermManage}}}
	name := "Kopi Susu"

	tests := []struct {
		name    string
		actorID int64
		id      int64
		wantErr error
	}{
		{"owner", 10, 1, nil},
		{"non-owner without manage", 20, 1, ErrNotOwner},
		{"non-owner with manage", 30, 1, nil},
		{"not found", 10, 99, ErrProductNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			_, err := NewService(repo, authz).Update(context.Background(), tt.actorID, tt.id, UpdateProductRequest{Name: &name})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if repo.updated != (tt.wantErr == nil) {
				t.Fatalf("updated = %v, want %v", repo.updated, tt.wantErr == nil)
			}
		})
	}
}

func TestDelete_Ownership(t *testing.T) {
	repo := newMockRepo()
	err := NewService(repo, mockAuthz{}).Delete(context.Background(), 20, 1)
	if !errors.Is(err, ErrNotOwner) {
		t.Fatalf("err = %v, want ErrNotOwner", err)
	}
	if repo.deleted {
		t.Fatal("product must not be deleted by non-owner")
	}
}
