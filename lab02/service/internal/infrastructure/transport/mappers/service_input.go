package mappers

import (
	"github.com/virogg/networks-course/service/internal/infrastructure/transport/dto"
	serviceDTO "github.com/virogg/networks-course/service/internal/service/dto"
)

func ToCreateProductInput(dto dto.CreateProductRequest) serviceDTO.CreateProductInput {
	return serviceDTO.CreateProductInput{
		Name:        dto.Name,
		Description: dto.Description,
	}
}

func ToUpdateProductInput(id int64, dto dto.UpdateProductRequest) serviceDTO.UpdateProductInput {
	return serviceDTO.UpdateProductInput{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
	}
}
