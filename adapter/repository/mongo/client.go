package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
)

// Client wraps the MongoDB client
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient creates a new MongoDB client based on configuration
func NewClient(ctx context.Context) (*Client, error) {
	cfg := config.GlobalConfig.MongoDB
	if cfg == nil {
		return nil, fmt.Errorf("MongoDB configuration is missing")
	}

	// Build connection URI
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.AuthSource,
	)

	// If no auth, use simpler URI
	if cfg.User == "" {
		uri = fmt.Sprintf("mongodb://%s:%d/%s",
			cfg.Host,
			cfg.Port,
			cfg.Database,
		)
	}

	// Client options
	clientOpts := options.Client().ApplyURI(uri)

	if cfg.MinPoolSize > 0 {
		clientOpts.SetMinPoolSize(uint64(cfg.MinPoolSize))
	}
	if cfg.MaxPoolSize > 0 {
		clientOpts.SetMaxPoolSize(uint64(cfg.MaxPoolSize))
	}
	if cfg.IdleTimeout > 0 {
		clientOpts.SetMaxConnIdleTime(time.Duration(cfg.IdleTimeout) * time.Second)
	}

	// Connect
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		database: client.Database(cfg.Database),
	}, nil
}

// GetDatabase returns the MongoDB database
func (c *Client) GetDatabase() *mongo.Database {
	return c.database
}

// GetCollection returns a collection from the database
func (c *Client) GetCollection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	return nil
}
