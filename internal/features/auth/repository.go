package auth

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"backend-core-go/internal/platform/database"
)

type Repository interface {
	FindCredentialsByEmail(ctx context.Context, email string) (*Credentials, error)
	CreateUser(ctx context.Context, name, email, passwordHash string) (int64, error)
}

type repository struct {
	db database.DBTX
	qb sq.StatementBuilderType
}

func NewRepository(db database.DBTX) Repository {
	return &repository{db: db, qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *repository) FindCredentialsByEmail(ctx context.Context, email string) (*Credentials, error) {
	query, args, err := r.qb.Select("id", "password_hash").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find credentials query: %w", err)
	}
	rows, _ := r.db.Query(ctx, query, args...)
	creds, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Credentials])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errCredentialsNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find credentials: %w", err)
	}
	return &creds, nil
}

func (r *repository) CreateUser(ctx context.Context, name, email, passwordHash string) (int64, error) {
	query, args, err := r.qb.Insert("users").
		Columns("name", "email", "password_hash").
		Values(name, email, passwordHash).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build insert user query: %w", err)
	}
	var id int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		if database.IsUniqueViolation(err) {
			return 0, ErrEmailTaken
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}
