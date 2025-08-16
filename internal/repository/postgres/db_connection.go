package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewDBConnection(connStr string) (*pgx.Conn, error) {

	// Подключение к базе данных
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	//Проверка подключения
	if err = db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
