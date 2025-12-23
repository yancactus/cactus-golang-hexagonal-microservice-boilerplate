package repo

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
)

// IProductRepo defines the interface for product repository operations
type IProductRepo interface {
	// Create creates a new product
	Create(ctx context.Context, product *model.Product) (*model.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *model.Product) error

	// Delete deletes a product by ID
	Delete(ctx context.Context, id string) error

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id string) (*model.Product, error)

	// GetByName retrieves a product by name
	GetByName(ctx context.Context, name string) (*model.Product, error)

	// List retrieves products with pagination
	List(ctx context.Context, offset, limit int) ([]*model.Product, int64, error)

	// UpdateStock updates product stock
	UpdateStock(ctx context.Context, id string, quantity int) error
}

// IProductCacheRepo defines the interface for product cache operations
type IProductCacheRepo interface {
	// GetByID retrieves a cached product by ID
	GetByID(ctx context.Context, id string) (*model.Product, error)

	// Set caches a product
	Set(ctx context.Context, product *model.Product) error

	// Delete removes a product from cache
	Delete(ctx context.Context, id string) error

	// Invalidate invalidates all product cache
	Invalidate(ctx context.Context) error
}
