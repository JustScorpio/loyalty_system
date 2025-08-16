package postgres

import (
	"context"
	"fmt"

	"github.com/JustScorpio/loyalty_system/internal/customcontext"
	"github.com/JustScorpio/loyalty_system/internal/models"
	"github.com/jackc/pgx/v5"
)

type PgUsersRepo struct {
	db *pgx.Conn
}

func NewPgUsersRepo(db *pgx.Conn) (*PgUsersRepo, error) {
	// Создание таблицы users, если её нет
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			login TEXT NOT NULL PRIMARY KEY,
			password TEXT,
			currentpoints REAL,
			withdrawnpoints REAL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &PgUsersRepo{db: db}, nil
}

func (r *PgUsersRepo) GetAll(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.Query(ctx, "SELECT login, password, currentpoints, withdrawnpoints FROM users")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err = rows.Err(); err != nil {
		return nil, err
	}

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Login, &user.Password, &user.CurrentPoints, &user.WithdrawnPoints)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *PgUsersRepo) Get(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, "SELECT login, password, currentpoints, withdrawnpoints FROM users WHERE login = $1", login).Scan(&user.Login, &user.Password, &user.CurrentPoints, &user.WithdrawnPoints)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PgUsersRepo) Create(ctx context.Context, user *models.User) error {
	err := r.execQuery(ctx, "INSERT INTO users (login, password, currentpoints, withdrawnpoints) VALUES ($1, $2, $3, $4)", &user.Login, &user.Password, &user.CurrentPoints, &user.WithdrawnPoints)
	if err != nil {
		return err
	}
	return nil
}

func (r *PgUsersRepo) Update(ctx context.Context, user *models.User) error {
	err := r.execQuery(ctx, "UPDATE users SET password = $2, currentpoints = $3, withdrawnpoints = $4 WHERE login = $1", user.Login, user.Password, user.CurrentPoints, user.WithdrawnPoints)
	return err
}

func (r *PgUsersRepo) Delete(ctx context.Context, login string) error {
	err := r.execQuery(ctx, "DELETE FROM users WHERE login = $1", login)
	return err
}

func (r *PgUsersRepo) PingDB() bool {
	err := r.db.Ping(context.Background())
	return err == nil
}

// execQuery выполняет запрос, автоматически используя транзакцию из контекста если она есть
func (r *PgUsersRepo) execQuery(ctx context.Context, query string, args ...interface{}) error {
	if tx, ok := customcontext.GetTx(ctx); ok {
		_, err := tx.Exec(ctx, query, args...)
		return err
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}
