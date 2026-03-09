package products

import (
	"context"

	"github.com/virogg/networks-course/service/internal/domain"
)

type productRepository interface {
	Create(ctx context.Context, product *domain.Product) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	GetAll(ctx context.Context) ([]*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id int64) (*domain.Product, error)
	SetIcon(ctx context.Context, id int64, iconPath string) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) (err error)
}
