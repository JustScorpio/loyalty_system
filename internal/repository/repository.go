package repository

import (
	"context"

	"github.com/JustScorpio/loyalty_system/internal/models"
)

// Интерфейс реализующий паттерн "репозиторий"
type IRepository[T models.Entity] interface {
	GetAll(ctx context.Context) ([]T, error)
	Get(ctx context.Context, id string) (*T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error

	PingDB() bool
}
