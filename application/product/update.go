package product

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// UpdateProductUseCase handles updating a product
type UpdateProductUseCase struct {
	productService service.IProductService
}

// NewUpdateProductUseCase creates a new UpdateProductUseCase
func NewUpdateProductUseCase(productService service.IProductService) *UpdateProductUseCase {
	return &UpdateProductUseCase{
		productService: productService,
	}
}

// Execute updates a product
func (uc *UpdateProductUseCase) Execute(ctx context.Context, input *UpdateProductInput) (*ProductOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	product, err := uc.productService.Update(ctx, input.ID, input.Name, input.Description, input.Price)
	if err != nil {
		return nil, err
	}

	return &ProductOutput{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}, nil
}
