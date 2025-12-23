package service

import (
	"context"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/repo"
)

// IAuditService defines the interface for audit service operations
type IAuditService interface {
	Create(ctx context.Context, audit *model.AuditLog) error
	GetByID(ctx context.Context, id string) (*model.AuditLog, error)
	GetByEntityType(ctx context.Context, entityType string, limit int, lastKey string) ([]*model.AuditLog, string, error)
	GetByEntityID(ctx context.Context, entityType, entityID string) ([]*model.AuditLog, error)
}

// AuditService implements IAuditService
type AuditService struct {
	repo repo.IAuditRepo
}

// NewAuditService creates a new audit service
func NewAuditService(repo repo.IAuditRepo) *AuditService {
	return &AuditService{repo: repo}
}

// Create saves a new audit log entry
func (s *AuditService) Create(ctx context.Context, audit *model.AuditLog) error {
	return s.repo.Create(ctx, audit)
}

// GetByID retrieves an audit log by ID
func (s *AuditService) GetByID(ctx context.Context, id string) (*model.AuditLog, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEntityType retrieves audit logs by entity type with pagination
func (s *AuditService) GetByEntityType(ctx context.Context, entityType string, limit int, lastKey string) ([]*model.AuditLog, string, error) {
	return s.repo.FindByEntityType(ctx, entityType, limit, lastKey)
}

// GetByEntityID retrieves audit logs for a specific entity
func (s *AuditService) GetByEntityID(ctx context.Context, entityType, entityID string) ([]*model.AuditLog, error) {
	return s.repo.FindByEntityID(ctx, entityType, entityID)
}
