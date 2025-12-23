package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action that was audited
type AuditAction string

const (
	AuditActionCreated AuditAction = "created"
	AuditActionUpdated AuditAction = "updated"
	AuditActionDeleted AuditAction = "deleted"
)

// AuditEntityType represents the type of entity being audited
type AuditEntityType string

const (
	AuditEntityUser    AuditEntityType = "user"
	AuditEntityProduct AuditEntityType = "product"
	AuditEntityOrder   AuditEntityType = "order"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         string
	EntityType string
	EntityID   string
	Action     string
	Payload    map[string]interface{}
	Timestamp  time.Time
	UserID     *int // User who performed the action (optional)
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(entityType AuditEntityType, entityID string, action AuditAction, payload map[string]interface{}, userID *int) *AuditLog {
	return &AuditLog{
		ID:         uuid.New().String(),
		EntityType: string(entityType),
		EntityID:   entityID,
		Action:     string(action),
		Payload:    payload,
		Timestamp:  time.Now(),
		UserID:     userID,
	}
}

// DomainEvent interface for domain events
type DomainEvent interface {
	EventName() string
}
