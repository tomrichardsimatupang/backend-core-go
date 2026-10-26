package product

import (
	"context"
	"fmt"
	"strings"

	"backend-core-go/internal/shared/pagination"
)

// Authorizer adalah interface kecil yang dipenuhi *rbac.Middleware, supaya
// service tidak bergantung langsung pada implementasi rbac.
type Authorizer interface {
	HasPermission(ctx context.Context, userID int64, perm string) (bool, error)
}

type Service interface {
	List(ctx context.Context, q ListQuery, p pagination.Params) ([]Product, pagination.Meta, error)
	Get(ctx context.Context, id int64) (*Product, error)
	Create(ctx context.Context, actorID int64, req CreateProductRequest) (*Product, error)
	Update(ctx context.Context, actorID, id int64, req UpdateProductRequest) (*Product, error)
	Delete(ctx context.Context, actorID, id int64) error
}

type service struct {
	repo  Repository
	authz Authorizer
}

func NewService(repo Repository, authz Authorizer) Service {
	return &service{repo: repo, authz: authz}
}

func (s *service) List(ctx context.Context, q ListQuery, p pagination.Params) ([]Product, pagination.Meta, error) {
	filter := ListFilter{Search: strings.TrimSpace(q.Search), OwnerID: q.OwnerID}
	products, total, err := s.repo.List(ctx, filter, p)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return products, pagination.NewMeta(p, total), nil
}

func (s *service) Get(ctx context.Context, id int64) (*Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) Create(ctx context.Context, actorID int64, req CreateProductRequest) (*Product, error) {
	p := &Product{
		OwnerID:     actorID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Price:       *req.Price,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) Update(ctx context.Context, actorID, id int64, req UpdateProductRequest) (*Product, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeOwner(ctx, actorID, p); err != nil {
		return nil, err
	}

	fields := UpdateFields{Price: req.Price}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		fields.Name = &name
	}
	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		fields.Description = &desc
	}
	if fields.IsEmpty() {
		return p, nil
	}
	return s.repo.Update(ctx, id, fields)
}

func (s *service) Delete(ctx context.Context, actorID, id int64) error {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorizeOwner(ctx, actorID, p); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// authorizeOwner: pemilik selalu boleh; selain pemilik butuh product:manage.
func (s *service) authorizeOwner(ctx context.Context, actorID int64, p *Product) error {
	if p.OwnerID == actorID {
		return nil
	}
	ok, err := s.authz.HasPermission(ctx, actorID, PermManage)
	if err != nil {
		return fmt.Errorf("check %s for user %d: %w", PermManage, actorID, err)
	}
	if !ok {
		return ErrNotOwner
	}
	return nil
}
