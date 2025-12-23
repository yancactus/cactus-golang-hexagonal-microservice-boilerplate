package order

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
)

// CreateOrderUseCase handles order creation
type CreateOrderUseCase struct {
	orderService service.IOrderService
}

// NewCreateOrderUseCase creates a new CreateOrderUseCase
func NewCreateOrderUseCase(orderService service.IOrderService) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderService: orderService,
	}
}

// Execute creates a new order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, input *CreateOrderInput) (*OrderOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	items := make([]model.OrderItem, len(input.Items))
	for i, item := range input.Items {
		items[i] = model.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	order, err := uc.orderService.Create(ctx, input.UserID, items)
	if err != nil {
		return nil, err
	}

	return toOrderOutput(order), nil
}

func toOrderOutput(order *model.Order) *OrderOutput {
	items := make([]OrderItemOutput, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemOutput{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &OrderOutput{
		ID:        order.ID,
		UserID:    order.UserID,
		Items:     items,
		Total:     order.Total,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}
