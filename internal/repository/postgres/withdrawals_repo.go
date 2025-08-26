package postgres

import (
	"context"
	"fmt"

	"github.com/JustScorpio/loyalty_system/internal/customcontext"
	"github.com/JustScorpio/loyalty_system/internal/models"
	"github.com/jackc/pgx/v5"
)

type PgWithdrawalsRepo struct {
	db *pgx.Conn
}

func NewPgWithdrawalsRepo(db *pgx.Conn) (*PgWithdrawalsRepo, error) {
	// Создание таблицы withdrawals, если её нет
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS withdrawals (
			userid TEXT NOT NULL,
			"order" TEXT NOT NULL PRIMARY KEY,
			sum REAL,
			processedat TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PgWithdrawalsRepo{db: db}, nil
}

func (r *PgWithdrawalsRepo) GetAll(ctx context.Context) ([]models.Withdrawal, error) {
	rows, err := r.db.Query(ctx, "SELECT userid, \"order\", sum, processedat FROM withdrawals")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err = rows.Err(); err != nil {
		return nil, err
	}

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal
		err := rows.Scan(&withdrawal.UserID, &withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	return withdrawals, nil
}

func (r *PgWithdrawalsRepo) Get(ctx context.Context, order string) (*models.Withdrawal, error) {
	var withdrawal models.Withdrawal
	err := r.db.QueryRow(ctx, "SELECT userid, \"order\", sum, processedat FROM withdrawals WHERE \"order\" = $1", order).Scan(&withdrawal.UserID, &withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)

	if err != nil {
		return nil, err
	}
	return &withdrawal, nil
}

func (r *PgWithdrawalsRepo) Create(ctx context.Context, withdrawal *models.Withdrawal) error {
	err := r.execQuery(ctx, "INSERT INTO withdrawals (userid, \"order\", sum, processedat) VALUES ($1, $2, $3, $4)", withdrawal.UserID, withdrawal.Order, withdrawal.Sum, withdrawal.ProcessedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *PgWithdrawalsRepo) Update(ctx context.Context, withdrawal *models.Withdrawal) error {
	err := r.execQuery(ctx, "UPDATE withdrawals SET userid = $1, sum = $3, processedat = $4 WHERE \"order\" = $2", withdrawal.UserID, withdrawal.Order, withdrawal.Sum, withdrawal.ProcessedAt)
	return err
}

func (r *PgWithdrawalsRepo) Delete(ctx context.Context, order string) error {
	err := r.execQuery(ctx, "DELETE FROM withdrawals WHERE \"order\" = $1", order)
	return err
}

func (r *PgWithdrawalsRepo) PingDB() bool {
	err := r.db.Ping(context.Background())
	return err == nil
}

// execQuery выполняет запрос, автоматически используя транзакцию из контекста если она есть
func (r *PgWithdrawalsRepo) execQuery(ctx context.Context, query string, args ...interface{}) error {
	if tx, ok := customcontext.GetTx(ctx); ok {
		_, err := tx.Exec(ctx, query, args...)
		return err
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}
