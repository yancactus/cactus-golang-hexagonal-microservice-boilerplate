package product

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// UpdateStockUseCase handles updating product stock
type UpdateStockUseCase struct {
	productService service.IProductService
}

// NewUpdateStockUseCase creates a new UpdateStockUseCase
func NewUpdateStockUseCase(productService service.IProductService) *UpdateStockUseCase {
	return &UpdateStockUseCase{
		productService: productService,
	}
}

// Execute updates the stock of a product
func (uc *UpdateStockUseCase) Execute(ctx context.Context, input *UpdateStockInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	return uc.productService.UpdateStock(ctx, input.ID, input.Quantity)
}
