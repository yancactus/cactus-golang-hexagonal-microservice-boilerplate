package order

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// GetOrderUseCase handles getting an order by ID
type GetOrderUseCase struct {
	orderService service.IOrderService
}

// NewGetOrderUseCase creates a new GetOrderUseCase
func NewGetOrderUseCase(orderService service.IOrderService) *GetOrderUseCase {
	return &GetOrderUseCase{
		orderService: orderService,
	}
}

// Execute retrieves an order by ID
func (uc *GetOrderUseCase) Execute(ctx context.Context, input *GetOrderInput) (*OrderOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	order, err := uc.orderService.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, ErrOrderNotFound
	}

	return toOrderOutput(order), nil
}
