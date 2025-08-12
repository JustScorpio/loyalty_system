package postgres

import (
	"context"
	"errors"

	"github.com/JustScorpio/loyalty_system/internal/customcontext"
	"github.com/jackc/pgx/v5"
)

type PgxTransactionManager struct {
	db *pgx.Conn
}

func NewPgxTransactionManager(db *pgx.Conn) *PgxTransactionManager {
	return &PgxTransactionManager{db: db}
}

func (tm *PgxTransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, err := tm.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tm.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx); err != nil {
		_ = tm.Rollback(ctx)
		return err
	}

	return tm.Commit(ctx)
}

func (tm *PgxTransactionManager) Begin(ctx context.Context) (context.Context, error) {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return customcontext.WithTx(ctx, tx), nil
}

func (tm *PgxTransactionManager) Commit(ctx context.Context) error {
	tx, ok := customcontext.GetTx(ctx)
	if !ok {
		return errors.New("no transaction in context")
	}
	return tx.Commit(ctx)
}

func (tm *PgxTransactionManager) Rollback(ctx context.Context) error {
	tx, ok := customcontext.GetTx(ctx)
	if !ok {
		return errors.New("no transaction in context")
	}
	return tx.Rollback(ctx)
}
