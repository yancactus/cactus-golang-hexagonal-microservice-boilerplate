package repo

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
)

// IOrderRepo defines the interface for order repository operations
type IOrderRepo interface {
	// Create creates a new order with items
	Create(ctx context.Context, tx Transaction, order *model.Order) (*model.Order, error)

	// Update updates an existing order
	Update(ctx context.Context, tx Transaction, order *model.Order) error

	// Delete deletes an order by ID
	Delete(ctx context.Context, tx Transaction, id string) error

	// GetByID retrieves an order by ID with items
	GetByID(ctx context.Context, tx Transaction, id string) (*model.Order, error)

	// GetByUserID retrieves orders for a user with pagination
	GetByUserID(ctx context.Context, tx Transaction, userID string, offset, limit int) ([]*model.Order, int64, error)

	// List retrieves orders with pagination
	List(ctx context.Context, tx Transaction, offset, limit int) ([]*model.Order, int64, error)

	// UpdateStatus updates the order status
	UpdateStatus(ctx context.Context, tx Transaction, id string, status model.OrderStatus) error
}

// IOrderCacheRepo defines the interface for order cache operations
type IOrderCacheRepo interface {
	// GetByID retrieves a cached order by ID
	GetByID(ctx context.Context, id string) (*model.Order, error)

	// Set caches an order
	Set(ctx context.Context, order *model.Order) error

	// Delete removes an order from cache
	Delete(ctx context.Context, id string) error

	// InvalidateByUser invalidates all orders for a user
	InvalidateByUser(ctx context.Context, userID string) error
}
