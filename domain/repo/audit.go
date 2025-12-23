package repo

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
)

// IAuditRepo defines the interface for audit log repository operations
type IAuditRepo interface {
	// Create saves a new audit log entry
	Create(ctx context.Context, audit *model.AuditLog) error

	// GetByID retrieves an audit log by ID
	GetByID(ctx context.Context, id string) (*model.AuditLog, error)

	// FindByEntityType retrieves audit logs by entity type with pagination
	// Returns the logs, the next pagination key, and any error
	FindByEntityType(ctx context.Context, entityType string, limit int, lastKey string) ([]*model.AuditLog, string, error)

	// FindByEntityID retrieves audit logs for a specific entity
	FindByEntityID(ctx context.Context, entityType, entityID string) ([]*model.AuditLog, error)
}
