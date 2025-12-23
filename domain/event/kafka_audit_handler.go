package event

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// KafkaProducer defines the interface for Kafka message production
type KafkaProducer interface {
	SendMessage(topic string, key, value []byte) error
}

// KafkaAuditHandler publishes domain events to Kafka for audit
type KafkaAuditHandler struct {
	producer KafkaProducer
	topic    string
	logger   *zap.Logger
}

// NewKafkaAuditHandler creates a new Kafka audit handler
func NewKafkaAuditHandler(producer KafkaProducer, topic string) *KafkaAuditHandler {
	logger, _ := zap.NewProduction()
	return &KafkaAuditHandler{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}
}

// AuditMessage represents the message sent to Kafka
type AuditMessage struct {
	ID         string                 `json:"id"`
	EventName  string                 `json:"event_name"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Payload    map[string]interface{} `json:"payload"`
	Timestamp  string                 `json:"timestamp"`
	UserID     *int                   `json:"user_id,omitempty"`
}

// HandleEvent publishes the event to Kafka
func (h *KafkaAuditHandler) HandleEvent(ctx context.Context, event Event) error {
	// Determine entity type and action from event name
	entityType, action := parseEventName(event.EventName())

	msg := AuditMessage{
		ID:         event.EventID(),
		EventName:  event.EventName(),
		EntityType: entityType,
		EntityID:   event.AggregateID(),
		Action:     action,
		Payload:    extractPayload(event),
		Timestamp:  event.OccurredAt().Format(time.RFC3339),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal audit message",
			zap.String("event_id", event.EventID()),
			zap.Error(err),
		)
		return err
	}

	if err := h.producer.SendMessage(h.topic, []byte(event.EventID()), data); err != nil {
		h.logger.Error("Failed to send audit message to Kafka",
			zap.String("event_id", event.EventID()),
			zap.String("topic", h.topic),
			zap.Error(err),
		)
		return err
	}

	log.SugaredLogger.Infof("Audit event published to Kafka: %s", event.EventName())
	return nil
}

// InterestedIn returns true for all domain events (audit everything)
func (h *KafkaAuditHandler) InterestedIn(eventName string) bool {
	// Audit all domain events
	return true
}

// parseEventName extracts entity type and action from event name
// e.g., "user.created" -> ("user", "created")
func parseEventName(eventName string) (entityType string, action string) {
	entityType = "unknown"
	action = "unknown"

	switch eventName {
	case "user.created":
		return "user", "created"
	case "user.updated":
		return "user", "updated"
	case "user.deleted":
		return "user", "deleted"
	case "product.created":
		return "product", "created"
	case "product.updated":
		return "product", "updated"
	case "product.deleted":
		return "product", "deleted"
	case "product.stock_updated":
		return "product", "stock_updated"
	case "order.created":
		return "order", "created"
	case "order.status_changed":
		return "order", "status_changed"
	case "order.cancelled":
		return "order", "canceled"
	}

	return entityType, action
}

// extractPayload extracts payload from event
func extractPayload(event Event) map[string]interface{} {
	if baseEvent, ok := event.(BaseEvent); ok {
		if payload, ok := baseEvent.Payload.(map[string]interface{}); ok {
			return payload
		}
	}

	// Default: return event name and aggregate ID
	return map[string]interface{}{
		"event_name":   event.EventName(),
		"aggregate_id": event.AggregateID(),
	}
}
