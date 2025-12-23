package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"cactus-golang-hexagonal-microservice-boilerplate/domain/event"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers []string
	Topic   string
}

// KafkaEventBus implements event.EventBus using Kafka
type KafkaEventBus struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaEventBus creates a new Kafka event bus
func NewKafkaEventBus(cfg *KafkaConfig) (*KafkaEventBus, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaEventBus{
		producer: producer,
		topic:    cfg.Topic,
	}, nil
}

// Publish publishes an event to Kafka
func (k *KafkaEventBus) Publish(ctx context.Context, event event.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:     k.topic,
		Value:     sarama.StringEncoder(payload),
		Timestamp: time.Now(),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_name"),
				Value: []byte(event.EventName()),
			},
			{
				Key:   []byte("event_id"),
				Value: []byte(event.EventID()),
			},
		},
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Logger.Info("Event published to Kafka",
		zap.String("event_name", event.EventName()),
		zap.String("event_id", event.EventID()),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

// Close closes the Kafka producer
func (k *KafkaEventBus) Close() error {
	if err := k.producer.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka producer: %w", err)
	}
	return nil
}

// SendMessage sends a message to Kafka (implements event.KafkaProducer interface)
func (k *KafkaEventBus) SendMessage(topic string, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.ByteEncoder(key),
		Value:     sarama.ByteEncoder(value),
		Timestamp: time.Now(),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Logger.Info("Message sent to Kafka",
		zap.String("topic", topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

// Subscribe is not implemented for Kafka (use KafkaConsumer instead)
func (k *KafkaEventBus) Subscribe(handler event.EventHandler) {}

// Unsubscribe is not implemented for Kafka
func (k *KafkaEventBus) Unsubscribe(handler event.EventHandler) {}
