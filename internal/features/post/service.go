package post

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"backend-core-go/internal/shared/pagination"
)

// Authorizer adalah interface kecil yang dipenuhi *rbac.Middleware.
type Authorizer interface {
	HasPermission(ctx context.Context, userID int64, perm string) (bool, error)
}

type Service interface {
	List(ctx context.Context, actorID int64, q ListQuery, p pagination.Params) ([]Post, pagination.Meta, error)
	Get(ctx context.Context, actorID, id int64) (*Post, error)
	Create(ctx context.Context, actorID int64, req CreatePostRequest) (*Post, error)
	Update(ctx context.Context, actorID, id int64, req UpdatePostRequest) (*Post, error)
	Delete(ctx context.Context, actorID, id int64) error
	Publish(ctx context.Context, id int64) (*Post, error)
	Unpublish(ctx context.Context, id int64) (*Post, error)
}

type service struct {
	repo  Repository
	authz Authorizer
}

func NewService(repo Repository, authz Authorizer) Service {
	return &service{repo: repo, authz: authz}
}

// List: tanpa post:manage, user hanya melihat post published + draft miliknya.
func (s *service) List(ctx context.Context, actorID int64, q ListQuery, p pagination.Params) ([]Post, pagination.Meta, error) {
	canManage, err := s.can(ctx, actorID, PermManage)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	filter := ListFilter{Search: strings.TrimSpace(q.Search), Status: q.Status, AuthorID: q.AuthorID}
	if !canManage {
		filter.VisibleTo = actorID
	}
	posts, total, err := s.repo.List(ctx, filter, p)
	if err != nil {
		return nil, pagination.Meta{}, err
	}
	return posts, pagination.NewMeta(p, total), nil
}

// Get: draft milik orang lain disembunyikan (404) kecuali punya post:manage.
func (s *service) Get(ctx context.Context, actorID, id int64) (*Post, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status == StatusPublished || p.AuthorID == actorID {
		return p, nil
	}
	canManage, err := s.can(ctx, actorID, PermManage)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, ErrPostNotFound
	}
	return p, nil
}

func (s *service) Create(ctx context.Context, actorID int64, req CreatePostRequest) (*Post, error) {
	title := strings.TrimSpace(req.Title)
	slug := slugify(title)
	if slug == "" {
		return nil, ErrInvalidTitle
	}
	p := &Post{
		AuthorID: actorID,
		Title:    title,
		Slug:     slug,
		Body:     req.Body,
		Status:   StatusDraft,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) Update(ctx context.Context, actorID, id int64, req UpdatePostRequest) (*Post, error) {
	p, err := s.Get(ctx, actorID, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeAuthor(ctx, actorID, p); err != nil {
		return nil, err
	}

	fields := UpdateFields{Body: req.Body}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		fields.Title = &title
	}
	if fields.IsEmpty() {
		return p, nil
	}
	return s.repo.Update(ctx, id, fields)
}

func (s *service) Delete(ctx context.Context, actorID, id int64) error {
	p, err := s.Get(ctx, actorID, id)
	if err != nil {
		return err
	}
	if err := s.authorizeAuthor(ctx, actorID, p); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// Publish/Unpublish dibatasi oleh post:publish di route; editor boleh
// mempublish post siapa pun.
func (s *service) Publish(ctx context.Context, id int64) (*Post, error) {
	return s.repo.Publish(ctx, id)
}

func (s *service) Unpublish(ctx context.Context, id int64) (*Post, error) {
	return s.repo.Unpublish(ctx, id)
}

// authorizeAuthor: penulis selalu boleh; selain penulis butuh post:manage.
func (s *service) authorizeAuthor(ctx context.Context, actorID int64, p *Post) error {
	if p.AuthorID == actorID {
		return nil
	}
	ok, err := s.can(ctx, actorID, PermManage)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotAuthor
	}
	return nil
}

func (s *service) can(ctx context.Context, actorID int64, perm string) (bool, error) {
	ok, err := s.authz.HasPermission(ctx, actorID, perm)
	if err != nil {
		return false, fmt.Errorf("check %s for user %d: %w", perm, actorID, err)
	}
	return ok, nil
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

const maxSlugLength = 200

func slugify(s string) string {
	slug := strings.Trim(nonSlugChars.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if len(slug) > maxSlugLength {
		slug = strings.TrimRight(slug[:maxSlugLength], "-")
	}
	return slug
}
