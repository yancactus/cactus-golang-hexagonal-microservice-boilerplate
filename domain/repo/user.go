package repo

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
)

// IUserRepo defines the interface for user repository operations
type IUserRepo interface {
	// Create creates a new user
	Create(ctx context.Context, tx Transaction, user *model.User) (*model.User, error)

	// Update updates an existing user
	Update(ctx context.Context, tx Transaction, user *model.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, tx Transaction, id string) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, tx Transaction, id string) (*model.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, tx Transaction, email string) (*model.User, error)

	// List retrieves users with pagination
	List(ctx context.Context, tx Transaction, offset, limit int) ([]*model.User, int64, error)
}

// IUserCacheRepo defines the interface for user cache operations
type IUserCacheRepo interface {
	// GetByID retrieves a cached user by ID
	GetByID(ctx context.Context, id string) (*model.User, error)

	// Set caches a user
	Set(ctx context.Context, user *model.User) error

	// Delete removes a user from cache
	Delete(ctx context.Context, id string) error

	// Invalidate invalidates all user cache
	Invalidate(ctx context.Context) error
}
