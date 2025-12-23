package order

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// UpdateStatusUseCase handles updating order status
type UpdateStatusUseCase struct {
	orderService service.IOrderService
}

// NewUpdateStatusUseCase creates a new UpdateStatusUseCase
func NewUpdateStatusUseCase(orderService service.IOrderService) *UpdateStatusUseCase {
	return &UpdateStatusUseCase{
		orderService: orderService,
	}
}

// Execute updates the status of an order
func (uc *UpdateStatusUseCase) Execute(ctx context.Context, input *UpdateStatusInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	status := model.OrderStatus(input.Status)
	return uc.orderService.UpdateStatus(ctx, input.ID, status)
}
