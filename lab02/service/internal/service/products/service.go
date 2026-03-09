package products

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/virogg/networks-course/service/internal/domain"
	domainerr "github.com/virogg/networks-course/service/internal/domain/errors"
	"github.com/virogg/networks-course/service/internal/service/dto"
	svcerr "github.com/virogg/networks-course/service/internal/service/errors"
)

type Service struct {
	txManager txManager
	repo      productRepository
	imageDir  string
}

func NewService(txManager txManager, repo productRepository, imageDir string) *Service {
	return &Service{
		txManager: txManager,
		repo:      repo,
		imageDir:  imageDir,
	}
}

func contentTypeFromExt(ext string) string {
	switch ext {
	case ".jpg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func (s *Service) CreateProduct(ctx context.Context, productDTO dto.CreateProductInput) (dto.ProductOutput, error) {
	var output dto.ProductOutput
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		product := domain.NewProduct(productDTO.Name, productDTO.Description)
		if err := product.Validate(); err != nil {
			return fmt.Errorf("%w: %v", svcerr.ErrValidation, err)
		}

		if id, err := s.repo.Create(ctx, product); err != nil {
			return fmt.Errorf("service error in `CreateCourier.Create`: %w", err)
		} else {
			var p *domain.Product
			if p, err = s.repo.GetByID(ctx, id); err != nil {
				return fmt.Errorf("service error in `CreateCourier.GetByID`: %w", err)
			}
			output = dto.ToProductOutput(p)
		}

		return nil
	})
	if err != nil {
		return output, err
	}

	return output, nil
}

func (s *Service) DeleteProduct(ctx context.Context, id int64) (dto.ProductOutput, error) {
	var output dto.ProductOutput
	if id <= 0 {
		return output, fmt.Errorf("%w: %v", svcerr.ErrInvalidInput, domainerr.ErrInvalidProductID)
	}
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		product, err := s.repo.Delete(ctx, id)
		if err != nil {
			if errors.Is(err, domainerr.ErrProductNotFound) {
				return fmt.Errorf("%w: product with id=%d not found for deletion", svcerr.ErrNotFound, id)
			}
			return fmt.Errorf("service error in `DeleteProduct`: %w", err)
		}
		output = dto.ToProductOutput(product)
		return nil
	})
	return output, err
}

func (s *Service) GetProductByID(ctx context.Context, id int64) (dto.ProductOutput, error) {
	var output dto.ProductOutput
	if id <= 0 {
		return output, fmt.Errorf("%w: %v", svcerr.ErrInvalidInput, domainerr.ErrInvalidProductID)
	}

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerr.ErrProductNotFound) {
			return output, fmt.Errorf("%w: product with id=%d not found", svcerr.ErrNotFound, id)
		}
		return output, fmt.Errorf("service error in `GetProductByID`: %w", err)
	}

	return dto.ToProductOutput(product), nil
}

func (s *Service) GetProducts(ctx context.Context) ([]dto.ProductOutput, error) {
	products, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service error in `GetProducts`: %w", err)
	}

	return dto.ToProductOutputs(products), nil
}

func (s *Service) UpdateProduct(ctx context.Context, productDTO dto.UpdateProductInput) (dto.ProductOutput, error) {
	var output dto.ProductOutput

	if productDTO.ID <= 0 {
		return output, fmt.Errorf("%w: %v", svcerr.ErrInvalidInput, domainerr.ErrInvalidProductID)
	}

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		product, err := s.repo.GetByID(ctx, productDTO.ID)
		if err != nil {
			if errors.Is(err, domainerr.ErrProductNotFound) {
				return fmt.Errorf("%w: product with id=%d not found for update", svcerr.ErrNotFound, productDTO.ID)
			}
			return fmt.Errorf("service error in `UpdateProduct`: %w", err)
		}

		if productDTO.Name != nil {
			product.Name = *productDTO.Name
		}
		if productDTO.Description != nil {
			product.Description = *productDTO.Description
		}

		if err := product.Validate(); err != nil {
			return fmt.Errorf("%w: %v", svcerr.ErrValidation, err)
		}

		if err := s.repo.Update(ctx, product); err != nil {
			if errors.Is(err, domainerr.ErrProductNotFound) {
				return fmt.Errorf("%w: courier with id=%d not found during update", svcerr.ErrNotFound, productDTO.ID)
			}
			return fmt.Errorf("service error in `UpdateProduct.Update`: %w", err)
		}

		if p, err := s.repo.GetByID(ctx, productDTO.ID); err != nil {
			return fmt.Errorf("service error in `UpdateProduct.GetByID`: %w", err)
		} else {
			output = dto.ToProductOutput(p)
		}

		return nil
	})
	if err != nil {
		return output, err
	}

	return output, nil
}

func (s *Service) UploadProductImage(ctx context.Context, id int64, data []byte, origFilename string) (dto.ProductOutput, error) {
	var output dto.ProductOutput
	if id <= 0 {
		return output, fmt.Errorf("%w: %v", svcerr.ErrInvalidInput, domainerr.ErrInvalidProductID)
	}
	if len(data) == 0 {
		return output, fmt.Errorf("%w: image data is empty", svcerr.ErrInvalidInput)
	}

	filename := filepath.Base(origFilename)
	if filename == "" || filename == "." {
		return output, fmt.Errorf("%w: invalid filename", svcerr.ErrInvalidInput)
	}
	fullPath := filepath.Join(s.imageDir, filename)

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		if _, err := s.repo.GetByID(ctx, id); err != nil {
			if errors.Is(err, domainerr.ErrProductNotFound) {
				return fmt.Errorf("%w: product with id=%d not found", svcerr.ErrNotFound, id)
			}
			return fmt.Errorf("service error in UploadProductImage.GetByID: %w", err)
		}

		if err := os.MkdirAll(s.imageDir, 0o755); err != nil {
			return fmt.Errorf("service error in UploadProductImage.MkdirAll: %w", err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("service error in UploadProductImage.WriteFile: %w", err)
		}
		if err := s.repo.SetIcon(ctx, id, filename); err != nil {
			return fmt.Errorf("service error in UploadProductImage.SetIcon: %w", err)
		}

		p, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("service error in UploadProductImage.GetByID(2): %w", err)
		}
		output = dto.ToProductOutput(p)
		return nil
	})
	return output, err
}

func (s *Service) GetProductImage(ctx context.Context, id int64) ([]byte, string, error) {
	if id <= 0 {
		return nil, "", fmt.Errorf("%w: %v", svcerr.ErrInvalidInput, domainerr.ErrInvalidProductID)
	}

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerr.ErrProductNotFound) {
			return nil, "", fmt.Errorf("%w: product with id=%d not found", svcerr.ErrNotFound, id)
		}
		return nil, "", fmt.Errorf("service error in GetProductImage.GetByID: %w", err)
	}
	if product.IconPath == "" {
		return nil, "", fmt.Errorf("%w: product with id=%d has no image", svcerr.ErrNotFound, id)
	}

	fullPath := filepath.Join(s.imageDir, product.IconPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, "", fmt.Errorf("%w: image file not found on disk", svcerr.ErrNotFound)
		}
		return nil, "", fmt.Errorf("service error in GetProductImage.ReadFile: %w", err)
	}

	ct := contentTypeFromExt(filepath.Ext(product.IconPath))
	return data, ct, nil
}
