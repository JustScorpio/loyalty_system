package postgres

import (
	"context"
	"fmt"

	// "github.com/JustScorpio/loyalty_system/internal/models"
	"github.com/jackc/pgx/v5"
)

type PostgresShURLRepository struct {
	db *pgx.Conn
}

func NewPostgresShURLRepository(connStr string) (*PostgresShURLRepository, error) {

	// Подключение к базе данных
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	//Проверка подключения
	if err = db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Создание таблицы shurls, если её нет
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT NOT NULL PRIMARY KEY,
			login TEXT NOT NULL UNIQUE,
			password ,
			loyaltypoints
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table shurls: %w", err)
	}

	return &PostgresShURLRepository{db: db}, nil
}

func (r *PostgresShURLRepository) GetAll(ctx context.Context) ([]entities.ShURL, error) {
	rows, err := r.db.Query(ctx, "SELECT token, longurl, createdby FROM shurls WHERE deleted = false")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err = rows.Err(); err != nil {
		return nil, err
	}

	var shurls []entities.ShURL
	for rows.Next() {
		var shurl entities.ShURL
		err := rows.Scan(&shurl.Token, &shurl.LongURL, &shurl.CreatedBy)
		if err != nil {
			return nil, err
		}
		shurls = append(shurls, shurl)
	}

	return shurls, nil
}

func (r *PostgresShURLRepository) Get(ctx context.Context, id string) (*entities.ShURL, error) {
	var shurl entities.ShURL
	var deleted bool
	err := r.db.QueryRow(ctx, "SELECT token, longurl, createdby, deleted FROM shurls WHERE token = $1", id).Scan(&shurl.Token, &shurl.LongURL, &shurl.CreatedBy, &deleted)

	if deleted {
		return nil, errGone
	}

	if err != nil {
		return nil, err
	}
	return &shurl, nil
}

func (r *PostgresShURLRepository) Create(ctx context.Context, shurl *entities.ShURL) error {
	_, err := r.db.Exec(ctx, "INSERT INTO shurls (token, longurl, createdBy) VALUES ($1, $2, $3)", shurl.Token, shurl.LongURL, shurl.CreatedBy)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresShURLRepository) Update(ctx context.Context, shurl *entities.ShURL) error {
	_, err := r.db.Exec(ctx, "UPDATE shurls SET longurl = $2, createdby = $3 WHERE token = $1", shurl.Token, shurl.LongURL, shurl.CreatedBy)
	return err
}

func (r *PostgresShURLRepository) Delete(ctx context.Context, ids []string, userID string) error {
	_, err := r.db.Exec(ctx, "UPDATE shurls SET deleted = true WHERE token = ANY($1) AND createdby = $2", ids, userID)
	return err
}

func (r *PostgresShURLRepository) CloseConnection() {
	r.db.Close(context.Background())
}

func (r *PostgresShURLRepository) PingDB() bool {
	err := r.db.Ping(context.Background())
	return err == nil
}
