package amqp

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// RabbitMQConsumer represents a RabbitMQ consumer
type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	handler MessageHandler
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(handler MessageHandler) (*RabbitMQConsumer, error) {
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

	// Set QoS
	if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		queue:   cfg.Queue,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Start starts consuming messages
func (c *RabbitMQConsumer) Start() error {
	msgs, err := c.channel.Consume(
		c.queue, // queue
		"",      // consumer tag
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Logger.Info("RabbitMQ message channel closed")
					return
				}

				if err := c.handler.HandleMessage(c.ctx, c.queue, []byte(msg.MessageId), msg.Body); err != nil {
					log.Logger.Error("Failed to handle message",
						zap.Error(err),
						zap.String("queue", c.queue),
						zap.String("message_id", msg.MessageId),
					)
					// Nack the message to retry
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}

			case <-c.ctx.Done():
				return
			}
		}
	}()

	log.Logger.Info("RabbitMQ consumer started", zap.String("queue", c.queue))
	return nil
}

// Stop stops the consumer gracefully
func (c *RabbitMQConsumer) Stop() error {
	c.cancel()
	c.wg.Wait()

	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	log.Logger.Info("RabbitMQ consumer stopped")
	return nil
}
