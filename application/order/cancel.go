package order

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// CancelOrderUseCase handles canceling an order
type CancelOrderUseCase struct {
	orderService service.IOrderService
}

// NewCancelOrderUseCase creates a new CancelOrderUseCase
func NewCancelOrderUseCase(orderService service.IOrderService) *CancelOrderUseCase {
	return &CancelOrderUseCase{
		orderService: orderService,
	}
}

// Execute cancels an order
func (uc *CancelOrderUseCase) Execute(ctx context.Context, input *CancelOrderInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	return uc.orderService.Cancel(ctx, input.ID)
}
