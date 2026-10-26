package post

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"backend-core-go/internal/platform/database"
	"backend-core-go/internal/shared/pagination"
)

type Repository interface {
	List(ctx context.Context, f ListFilter, p pagination.Params) ([]Post, int64, error)
	FindByID(ctx context.Context, id int64) (*Post, error)
	Create(ctx context.Context, p *Post) error
	Update(ctx context.Context, id int64, fields UpdateFields) (*Post, error)
	Publish(ctx context.Context, id int64) (*Post, error)
	Unpublish(ctx context.Context, id int64) (*Post, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db database.DBTX
	qb sq.StatementBuilderType
}

func NewRepository(db database.DBTX) Repository {
	return &repository{db: db, qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

var postColumns = []string{"id", "author_id", "title", "slug", "body", "status", "published_at", "created_at", "updated_at"}

const returningPost = "RETURNING id, author_id, title, slug, body, status, published_at, created_at, updated_at"

func (r *repository) List(ctx context.Context, f ListFilter, p pagination.Params) ([]Post, int64, error) {
	where := sq.And{}
	if f.Search != "" {
		where = append(where, sq.ILike{"title": "%" + f.Search + "%"})
	}
	if f.Status != "" {
		where = append(where, sq.Eq{"status": f.Status})
	}
	if f.AuthorID != 0 {
		where = append(where, sq.Eq{"author_id": f.AuthorID})
	}
	if f.VisibleTo != 0 {
		where = append(where, sq.Or{sq.Eq{"status": StatusPublished}, sq.Eq{"author_id": f.VisibleTo}})
	}

	countQuery, countArgs, err := r.qb.Select("count(*)").From("posts").Where(where).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count posts query: %w", err)
	}
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	query, args, err := r.qb.Select(postColumns...).
		From("posts").
		Where(where).
		OrderBy("published_at DESC NULLS LAST", "id DESC").
		Limit(p.Limit()).
		Offset(p.Offset()).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list posts query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	posts, err := pgx.CollectRows(rows, pgx.RowToStructByName[Post])
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}
	return posts, total, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*Post, error) {
	query, args, err := r.qb.Select(postColumns...).
		From("posts").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find post query: %w", err)
	}
	return r.queryOne(ctx, id, query, args)
}

func (r *repository) Create(ctx context.Context, p *Post) error {
	query, args, err := r.qb.Insert("posts").
		Columns("author_id", "title", "slug", "body", "status").
		Values(p.AuthorID, p.Title, p.Slug, p.Body, p.Status).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert post query: %w", err)
	}
	if err := r.db.QueryRow(ctx, query, args...).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if database.IsUniqueViolation(err) {
			return ErrSlugTaken
		}
		return fmt.Errorf("insert post: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, id int64, fields UpdateFields) (*Post, error) {
	ub := r.qb.Update("posts").
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id})
	if fields.Title != nil {
		ub = ub.Set("title", *fields.Title)
	}
	if fields.Body != nil {
		ub = ub.Set("body", *fields.Body)
	}
	query, args, err := ub.Suffix(returningPost).ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update post query: %w", err)
	}
	return r.queryOne(ctx, id, query, args)
}

func (r *repository) Publish(ctx context.Context, id int64) (*Post, error) {
	// published_at dipertahankan bila post pernah dipublish sebelumnya.
	query, args, err := r.qb.Update("posts").
		Set("status", StatusPublished).
		Set("published_at", sq.Expr("COALESCE(published_at, now())")).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id}).
		Suffix(returningPost).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build publish post query: %w", err)
	}
	return r.queryOne(ctx, id, query, args)
}

func (r *repository) Unpublish(ctx context.Context, id int64) (*Post, error) {
	query, args, err := r.qb.Update("posts").
		Set("status", StatusDraft).
		Set("published_at", nil).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id}).
		Suffix(returningPost).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build unpublish post query: %w", err)
	}
	return r.queryOne(ctx, id, query, args)
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	query, args, err := r.qb.Delete("posts").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("build delete post query: %w", err)
	}
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete post %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *repository) queryOne(ctx context.Context, id int64, query string, args []any) (*Post, error) {
	rows, _ := r.db.Query(ctx, query, args...)
	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Post])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query post %d: %w", id, err)
	}
	return &p, nil
}
