package customcontext

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Определяем собственный тип для ключа
type contextKey int

const (
	userIDKey contextKey = iota
	txKey
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) string {
	userID := ctx.Value(userIDKey)
	if userID == nil {
		userID = ""
	}

	return userID.(string)
}

// WithTx добавляет транзакцию в контекст
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetTx извлекает транзакцию из контекста
func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}
