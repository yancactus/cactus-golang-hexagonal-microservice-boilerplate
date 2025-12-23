package product

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// CreateProductUseCase handles product creation
type CreateProductUseCase struct {
	productService service.IProductService
}

// NewCreateProductUseCase creates a new CreateProductUseCase
func NewCreateProductUseCase(productService service.IProductService) *CreateProductUseCase {
	return &CreateProductUseCase{
		productService: productService,
	}
}

// Execute creates a new product
func (uc *CreateProductUseCase) Execute(ctx context.Context, input *CreateProductInput) (*ProductOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	product, err := uc.productService.Create(ctx, input.Name, input.Description, input.Price, input.Stock)
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
