package product

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// DeleteProductUseCase handles deleting a product
type DeleteProductUseCase struct {
	productService service.IProductService
}

// NewDeleteProductUseCase creates a new DeleteProductUseCase
func NewDeleteProductUseCase(productService service.IProductService) *DeleteProductUseCase {
	return &DeleteProductUseCase{
		productService: productService,
	}
}

// Execute deletes a product
func (uc *DeleteProductUseCase) Execute(ctx context.Context, input *DeleteProductInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	return uc.productService.Delete(ctx, input.ID)
}
