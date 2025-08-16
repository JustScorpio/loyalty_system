package repository

import "context"

type ITransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	//По факту не нужны. Но пусть будут
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
