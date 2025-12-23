package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// KafkaConsumer represents a Kafka consumer
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topics   []string
	handler  MessageHandler
	ready    chan bool
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// MessageHandler is the interface for handling Kafka messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, topic string, key, value []byte) error
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(handler MessageHandler) (*KafkaConsumer, error) {
	cfg := config.GlobalConfig.Kafka
	if cfg == nil {
		return nil, fmt.Errorf("Kafka configuration is missing")
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	if cfg.Consumer.AutoCommit {
		saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
	}

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.ConsumerGroup, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaConsumer{
		consumer: consumer,
		topics:   []string{cfg.Topics.AuditEvents},
		handler:  handler,
		ready:    make(chan bool),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start starts consuming messages
func (c *KafkaConsumer) Start() error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumer.Consume(c.ctx, c.topics, c); err != nil {
				log.Logger.Error("Error from consumer", zap.Error(err))
			}
			if c.ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready
	log.Logger.Info("Kafka consumer started", zap.Strings("topics", c.topics))
	return nil
}

// Stop stops the consumer gracefully
func (c *KafkaConsumer) Stop() error {
	c.cancel()
	c.wg.Wait()
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}
	log.Logger.Info("Kafka consumer stopped")
	return nil
}

// Setup is run at the beginning of a new session
func (c *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session
func (c *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition
func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Logger.Info("Message channel was closed")
				return nil
			}

			if err := c.handler.HandleMessage(c.ctx, message.Topic, message.Key, message.Value); err != nil {
				log.Logger.Error("Failed to handle message",
					zap.Error(err),
					zap.String("topic", message.Topic),
					zap.Int32("partition", message.Partition),
					zap.Int64("offset", message.Offset),
				)
			} else {
				session.MarkMessage(message, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

// AuditEventMessage represents an audit event message from Kafka
type AuditEventMessage struct {
	ID         string                 `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Payload    map[string]interface{} `json:"payload"`
	Timestamp  string                 `json:"timestamp"`
	UserID     *int                   `json:"user_id,omitempty"`
}

// ParseAuditEventMessage parses an audit event message from JSON
func ParseAuditEventMessage(data []byte) (*AuditEventMessage, error) {
	var msg AuditEventMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit event: %w", err)
	}
	return &msg, nil
}
