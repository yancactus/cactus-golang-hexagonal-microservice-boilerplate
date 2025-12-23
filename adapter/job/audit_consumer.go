package job

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/model"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/service"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// AuditConsumerHandler handles audit messages from Kafka and saves to DynamoDB
type AuditConsumerHandler struct {
	auditService service.IAuditService
	logger       *zap.Logger
}

// NewAuditConsumerHandler creates a new audit consumer handler
func NewAuditConsumerHandler(auditService service.IAuditService) *AuditConsumerHandler {
	logger, _ := zap.NewProduction()
	return &AuditConsumerHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// AuditEventMessage represents the audit event message from Kafka
type AuditEventMessage struct {
	ID         string                 `json:"id"`
	EventName  string                 `json:"event_name"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Payload    map[string]interface{} `json:"payload"`
	Timestamp  string                 `json:"timestamp"`
	UserID     *int                   `json:"user_id,omitempty"`
}

// HandleMessage processes a Kafka message and saves to DynamoDB
func (h *AuditConsumerHandler) HandleMessage(ctx context.Context, topic string, key, value []byte) error {
	var msg AuditEventMessage
	if err := json.Unmarshal(value, &msg); err != nil {
		h.logger.Error("Failed to unmarshal audit message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return err
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, msg.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	// Create audit log entry
	auditLog := &model.AuditLog{
		ID:         msg.ID,
		EntityType: msg.EntityType,
		EntityID:   msg.EntityID,
		Action:     msg.Action,
		Payload:    msg.Payload,
		Timestamp:  timestamp,
		UserID:     msg.UserID,
	}

	// Save to DynamoDB via AuditService
	if err := h.auditService.Create(ctx, auditLog); err != nil {
		h.logger.Error("Failed to save audit log to DynamoDB",
			zap.String("id", msg.ID),
			zap.String("entity_type", msg.EntityType),
			zap.Error(err),
		)
		return err
	}

	log.SugaredLogger.Infof("Audit log saved to DynamoDB: %s (%s.%s)", msg.ID, msg.EntityType, msg.Action)
	return nil
}
