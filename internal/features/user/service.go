package user

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"backend-core-go/internal/shared/pagination"
)

type Service interface {
	List(ctx context.Context, p pagination.Params) ([]User, pagination.Meta, error)
	Get(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, req CreateUserRequest) (*User, error)
	Update(ctx context.Context, id int64, req UpdateUserRequest) (*User, error)
	Delete(ctx context.Context, actorID, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, p pagination.Params) ([]User, pagination.Meta, error) {
	users, total, err := s.repo.List(ctx, p)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return users, pagination.NewMeta(p, total), nil
}

func (s *service) Get(ctx context.Context, id int64) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u := &User{Name: strings.TrimSpace(req.Name), Email: normalizeEmail(req.Email)}
	if err := s.repo.Create(ctx, u, string(hash)); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *service) Update(ctx context.Context, id int64, req UpdateUserRequest) (*User, error) {
	var fields UpdateFields
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		fields.Name = &name
	}
	if req.Email != nil {
		email := normalizeEmail(*req.Email)
		fields.Email = &email
	}
	if fields.IsEmpty() {
		return s.repo.FindByID(ctx, id)
	}
	return s.repo.Update(ctx, id, fields)
}

func (s *service) Delete(ctx context.Context, actorID, id int64) error {
	if actorID == id {
		return ErrCannotDeleteSelf
	}
	return s.repo.Delete(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
