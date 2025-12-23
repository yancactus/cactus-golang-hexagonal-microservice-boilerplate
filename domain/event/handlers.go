package event

import (
	"context"

	"go.uber.org/zap"
)

// LoggingEventHandler logs all events
type LoggingEventHandler struct {
	logger *zap.Logger
}

// NewLoggingEventHandler creates a new logging event handler
func NewLoggingEventHandler() *LoggingEventHandler {
	logger, _ := zap.NewProduction()
	return &LoggingEventHandler{
		logger: logger,
	}
}

// HandleEvent logs the event
func (h *LoggingEventHandler) HandleEvent(ctx context.Context, event Event) error {
	h.logger.Info("Event received",
		zap.String("event_id", event.EventID()),
		zap.String("event_name", event.EventName()),
		zap.String("aggregate_id", event.AggregateID()),
		zap.Time("occurred_at", event.OccurredAt()),
	)
	return nil
}

// InterestedIn returns true for all events
func (h *LoggingEventHandler) InterestedIn(eventName string) bool {
	return true
}
