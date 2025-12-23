package product

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// GetProductUseCase handles getting a product by ID
type GetProductUseCase struct {
	productService service.IProductService
}

// NewGetProductUseCase creates a new GetProductUseCase
func NewGetProductUseCase(productService service.IProductService) *GetProductUseCase {
	return &GetProductUseCase{
		productService: productService,
	}
}

// Execute retrieves a product by ID
func (uc *GetProductUseCase) Execute(ctx context.Context, input *GetProductInput) (*ProductOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	product, err := uc.productService.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, ErrProductNotFound
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
