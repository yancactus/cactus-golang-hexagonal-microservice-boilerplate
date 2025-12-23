package product

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// ListProductsUseCase handles listing products
type ListProductsUseCase struct {
	productService service.IProductService
}

// NewListProductsUseCase creates a new ListProductsUseCase
func NewListProductsUseCase(productService service.IProductService) *ListProductsUseCase {
	return &ListProductsUseCase{
		productService: productService,
	}
}

// Execute lists products with pagination
func (uc *ListProductsUseCase) Execute(ctx context.Context, input *ListProductsInput) (*ListProductsOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	products, total, err := uc.productService.List(ctx, input.Offset, input.Limit)
	if err != nil {
		return nil, err
	}

	output := &ListProductsOutput{
		Products: make([]*ProductOutput, len(products)),
		Total:    total,
	}

	for i, p := range products {
		output.Products[i] = &ProductOutput{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	return output, nil
}
