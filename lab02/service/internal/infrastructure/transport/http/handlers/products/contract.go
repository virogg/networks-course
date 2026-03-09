package products

import (
	"context"

	"github.com/virogg/networks-course/service/internal/service/dto"
)

type productsService interface {
	CreateProduct(ctx context.Context, productDTO dto.CreateProductInput) (dto.ProductOutput, error)
	DeleteProduct(ctx context.Context, id int64) (dto.ProductOutput, error)
	GetProductByID(ctx context.Context, id int64) (dto.ProductOutput, error)
	GetProducts(ctx context.Context) ([]dto.ProductOutput, error)
	UpdateProduct(ctx context.Context, productDTO dto.UpdateProductInput) (dto.ProductOutput, error)
	UploadProductImage(ctx context.Context, id int64, data []byte, origFilename string) (dto.ProductOutput, error)
	GetProductImage(ctx context.Context, id int64) ([]byte, string, error)
}
