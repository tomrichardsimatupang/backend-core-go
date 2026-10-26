package user

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
	List(ctx context.Context, p pagination.Params) ([]User, int64, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, u *User, passwordHash string) error
	Update(ctx context.Context, id int64, fields UpdateFields) (*User, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db database.DBTX
	qb sq.StatementBuilderType
}

func NewRepository(db database.DBTX) Repository {
	return &repository{db: db, qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

var userColumns = []string{"id", "name", "email", "created_at", "updated_at"}

func (r *repository) List(ctx context.Context, p pagination.Params) ([]User, int64, error) {
	countQuery, countArgs, err := r.qb.Select("count(*)").From("users").ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count users query: %w", err)
	}
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	query, args, err := r.qb.Select(userColumns...).
		From("users").
		OrderBy("id").
		Limit(p.Limit()).
		Offset(p.Offset()).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list users query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (*User, error) {
	query, args, err := r.qb.Select(userColumns...).
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find user query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user %d: %w", id, err)
	}
	return &u, nil
}

func (r *repository) Create(ctx context.Context, u *User, passwordHash string) error {
	query, args, err := r.qb.Insert("users").
		Columns("name", "email", "password_hash").
		Values(u.Name, u.Email, passwordHash).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert user query: %w", err)
	}
	if err := r.db.QueryRow(ctx, query, args...).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if database.IsUniqueViolation(err) {
			return ErrEmailTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, id int64, fields UpdateFields) (*User, error) {
	ub := r.qb.Update("users").
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id})
	if fields.Name != nil {
		ub = ub.Set("name", *fields.Name)
	}
	if fields.Email != nil {
		ub = ub.Set("email", *fields.Email)
	}
	query, args, err := ub.Suffix("RETURNING id, name, email, created_at, updated_at").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update user query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, ErrUserNotFound
	case database.IsUniqueViolation(err):
		return nil, ErrEmailTaken
	case err != nil:
		return nil, fmt.Errorf("update user %d: %w", id, err)
	}
	return &u, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	query, args, err := r.qb.Delete("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("build delete user query: %w", err)
	}
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete user %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
