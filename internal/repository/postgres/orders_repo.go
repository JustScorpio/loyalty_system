package postgres

import (
	"context"
	"fmt"

	"github.com/JustScorpio/loyalty_system/internal/customcontext"
	"github.com/JustScorpio/loyalty_system/internal/models"
	"github.com/jackc/pgx/v5"
)

type PgOrdersRepo struct {
	db *pgx.Conn
}

func NewPgOrdersRepo(db *pgx.Conn) (*PgOrdersRepo, error) {
	// Создание таблицы orders, если её нет
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS orders (
			userid TEXT NOT NULL,
			number TEXT NOT NULL PRIMARY KEY,
			accrual REAL,
			status TEXT,
			uploadedat TIMESTAMP
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PgOrdersRepo{db: db}, nil
}

func (r *PgOrdersRepo) GetAll(ctx context.Context) ([]models.Order, error) {
	rows, err := r.db.Query(ctx, "SELECT userid, number, accrual, status, uploadedat FROM orders")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err = rows.Err(); err != nil {
		return nil, err
	}

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.UserID, &order.Number, &order.Accrual, &order.Status, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *PgOrdersRepo) Get(ctx context.Context, number string) (*models.Order, error) {
	var order models.Order
	err := r.db.QueryRow(ctx, "SELECT userid, number, accrual, status, uploadedat FROM orders WHERE number = $1", number).Scan(&order.UserID, &order.Number, &order.Accrual, &order.Status, &order.UploadedAt)

	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *PgOrdersRepo) Create(ctx context.Context, order *models.Order) error {
	err := r.execQuery(ctx, "INSERT INTO orders (userid, number, accrual, status, uploadedat) VALUES ($1, $2, $3, $4, $5)", order.UserID, order.Number, order.Accrual, order.Status, order.UploadedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *PgOrdersRepo) Update(ctx context.Context, order *models.Order) error {
	err := r.execQuery(ctx, "UPDATE orders SET userid = $1, number = $2, accrual = $3, status = $4, uploadedat = $5 WHERE number = $2", order.UserID, order.Number, order.Accrual, order.Status, order.UploadedAt)
	return err
}

func (r *PgOrdersRepo) Delete(ctx context.Context, number string) error {
	err := r.execQuery(ctx, "DELETE FROM orders WHERE number = $1", number)
	return err
}

func (r *PgOrdersRepo) CloseConnection() {
	r.db.Close(context.Background())
}

func (r *PgOrdersRepo) PingDB() bool {
	err := r.db.Ping(context.Background())
	return err == nil
}

// execQuery выполняет запрос, автоматически используя транзакцию из контекста если она есть
func (r *PgOrdersRepo) execQuery(ctx context.Context, query string, args ...interface{}) error {
	if tx, ok := customcontext.GetTx(ctx); ok {
		_, err := tx.Exec(ctx, query, args...)
		return err
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}
