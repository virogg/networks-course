package dto

import "github.com/virogg/networks-course/service/internal/domain"

type CreateProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateProductInput struct {
	ID          int64   `json:"id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ProductOutput struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IconPath    string `json:"icon_path,omitempty"`
}

func ToProductOutput(product *domain.Product) ProductOutput {
	return ProductOutput{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		IconPath:    product.IconPath,
	}
}

func ToProductOutputs(products []*domain.Product) []ProductOutput {
	outputs := make([]ProductOutput, len(products))
	for i, p := range products {
		outputs[i] = ToProductOutput(p)
	}
	return outputs
}
