package product

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
	List(ctx context.Context, f ListFilter, p pagination.Params) ([]Product, int64, error)
	FindByID(ctx context.Context, id int64) (*Product, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, id int64, fields UpdateFields) (*Product, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db database.DBTX
	qb sq.StatementBuilderType
}

func NewRepository(db database.DBTX) Repository {
	return &repository{db: db, qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

var productColumns = []string{"id", "owner_id", "name", "description", "price", "created_at", "updated_at"}

func (r *repository) List(ctx context.Context, f ListFilter, p pagination.Params) ([]Product, int64, error) {
	where := sq.And{}
	if f.Search != "" {
		where = append(where, sq.ILike{"name": "%" + f.Search + "%"})
	}
	if f.OwnerID != 0 {
		where = append(where, sq.Eq{"owner_id": f.OwnerID})
	}

	countQuery, countArgs, err := r.qb.Select("count(*)").From("products").Where(where).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count products query: %w", err)
	}
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	query, args, err := r.qb.Select(productColumns...).
		From("products").
		Where(where).
		OrderBy("id DESC").
		Limit(p.Limit()).
		Offset(p.Offset()).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list products query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, 0, fmt.Errorf("list products: %w", err)
	}
	return products, total, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*Product, error) {
	query, args, err := r.qb.Select(productColumns...).
		From("products").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find product query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find product %d: %w", id, err)
	}
	return &p, nil
}

func (r *repository) Create(ctx context.Context, p *Product) error {
	query, args, err := r.qb.Insert("products").
		Columns("owner_id", "name", "description", "price").
		Values(p.OwnerID, p.Name, p.Description, p.Price).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert product query: %w", err)
	}
	if err := r.db.QueryRow(ctx, query, args...).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return fmt.Errorf("insert product: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, id int64, fields UpdateFields) (*Product, error) {
	ub := r.qb.Update("products").
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id})
	if fields.Name != nil {
		ub = ub.Set("name", *fields.Name)
	}
	if fields.Description != nil {
		ub = ub.Set("description", *fields.Description)
	}
	if fields.Price != nil {
		ub = ub.Set("price", *fields.Price)
	}
	query, args, err := ub.Suffix("RETURNING id, owner_id, name, description, price, created_at, updated_at").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update product query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	p, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update product %d: %w", id, err)
	}
	return &p, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	query, args, err := r.qb.Delete("products").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("build delete product query: %w", err)
	}
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete product %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}
