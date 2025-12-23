package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// RabbitMQEventBus implements event.EventBus using RabbitMQ
type RabbitMQEventBus struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

// NewRabbitMQEventBus creates a new RabbitMQ event bus
func NewRabbitMQEventBus() (*RabbitMQEventBus, error) {
	cfg := config.GlobalConfig.RabbitMQ
	if cfg == nil {
		return nil, fmt.Errorf("RabbitMQ configuration is missing")
	}

	// Build connection URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.VHost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		cfg.Queue,      // queue name
		cfg.RoutingKey, // routing key
		cfg.Exchange,   // exchange
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQEventBus{
		conn:     conn,
		channel:  ch,
		exchange: cfg.Exchange,
		queue:    cfg.Queue,
	}, nil
}

// Publish publishes an event to RabbitMQ
func (r *RabbitMQEventBus) Publish(ctx context.Context, evt event.Event) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	routingKey := config.GlobalConfig.RabbitMQ.RoutingKey

	err = r.channel.PublishWithContext(ctx,
		r.exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			MessageId:    evt.EventID(),
			Type:         evt.EventName(),
			Body:         payload,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Logger.Info("Event published to RabbitMQ",
		zap.String("event_name", evt.EventName()),
		zap.String("event_id", evt.EventID()),
		zap.String("exchange", r.exchange),
		zap.String("routing_key", routingKey),
	)

	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQEventBus) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}

// Subscribe is not implemented for RabbitMQ producer
func (r *RabbitMQEventBus) Subscribe(handler event.EventHandler) {
	// Not implemented - use RabbitMQConsumer for consuming
}

// Unsubscribe is not implemented for RabbitMQ producer
func (r *RabbitMQEventBus) Unsubscribe(handler event.EventHandler) {
	// Not implemented
}
