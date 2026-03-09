package mappers

import (
	"strconv"

	"github.com/virogg/networks-course/service/internal/infrastructure/transport/dto"
	serviceDTO "github.com/virogg/networks-course/service/internal/service/dto"
)

func ToProductResponse(output serviceDTO.ProductOutput) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          strconv.FormatInt(output.ID, 10),
		Name:        output.Name,
		Description: output.Description,
		Icon:        output.IconPath,
	}
}
