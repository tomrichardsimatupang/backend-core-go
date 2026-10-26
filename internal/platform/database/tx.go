package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX dipenuhi oleh *pgxpool.Pool dan pgx.Tx, sehingga repository
// bisa dipakai di dalam maupun di luar transaksi.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Transactor membuka transaksi dari service. Commit jika fn sukses,
// rollback jika fn mengembalikan error atau panic.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(tx DBTX) error) error
}

type poolTransactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) Transactor {
	return &poolTransactor{pool: pool}
}

func (t *poolTransactor) WithinTx(ctx context.Context, fn func(tx DBTX) error) error {
	return pgx.BeginFunc(ctx, t.pool, func(tx pgx.Tx) error {
		return fn(tx)
	})
}
