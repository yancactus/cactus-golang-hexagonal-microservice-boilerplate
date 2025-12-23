package order

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// ListOrdersUseCase handles listing orders
type ListOrdersUseCase struct {
	orderService service.IOrderService
}

// NewListOrdersUseCase creates a new ListOrdersUseCase
func NewListOrdersUseCase(orderService service.IOrderService) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		orderService: orderService,
	}
}

// Execute lists orders with pagination
func (uc *ListOrdersUseCase) Execute(ctx context.Context, input *ListOrdersInput) (*ListOrdersOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	orders, total, err := uc.orderService.List(ctx, input.Offset, input.Limit)
	if err != nil {
		return nil, err
	}

	output := &ListOrdersOutput{
		Orders: make([]*OrderOutput, len(orders)),
		Total:  total,
	}

	for i, o := range orders {
		output.Orders[i] = toOrderOutput(o)
	}

	return output, nil
}

// GetUserOrdersUseCase handles getting orders by user
type GetUserOrdersUseCase struct {
	orderService service.IOrderService
}

// NewGetUserOrdersUseCase creates a new GetUserOrdersUseCase
func NewGetUserOrdersUseCase(orderService service.IOrderService) *GetUserOrdersUseCase {
	return &GetUserOrdersUseCase{
		orderService: orderService,
	}
}

// Execute retrieves orders for a user
func (uc *GetUserOrdersUseCase) Execute(ctx context.Context, input *GetUserOrdersInput) (*ListOrdersOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	orders, total, err := uc.orderService.GetByUserID(ctx, input.UserID, input.Offset, input.Limit)
	if err != nil {
		return nil, err
	}

	output := &ListOrdersOutput{
		Orders: make([]*OrderOutput, len(orders)),
		Total:  total,
	}

	for i, o := range orders {
		output.Orders[i] = toOrderOutput(o)
	}

	return output, nil
}
