package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

// MockDB is a mock implementation of gorm.DB
type MockDB struct {
	mock.Mock
}

func TestNewPostgreSQLClient(t *testing.T) {
	// Create PostgreSQL client with nil db
	pgClient := repository.NewPostgreSQLClient(nil)

	// Verify client is not nil
	assert.NotNil(t, pgClient)

	// Verify GetDB returns nil (because db itself is nil)
	ctx := context.Background()
	assert.Nil(t, pgClient.GetDB(ctx))

	// Verify Close method doesn't return an error
	err := pgClient.Close(ctx)
	assert.NoError(t, err)
}

func TestNewRedisClient(t *testing.T) {
	// Create Redis client
	redisClient := repository.NewRedisClient()

	// Verify client is not nil
	assert.NotNil(t, redisClient)

	// Verify Close method doesn't return an error
	ctx := context.Background()
	err := redisClient.Close(ctx)
	assert.NoError(t, err)
}

func TestNewMongoDBClient(t *testing.T) {
	// Create MongoDB client
	mongoClient := repository.NewMongoDBClient()

	// Verify client is not nil
	assert.NotNil(t, mongoClient)

	// Verify Close method doesn't return an error
	ctx := context.Background()
	err := mongoClient.Close(ctx)
	assert.NoError(t, err)
}

func TestNewDynamoDBClient(t *testing.T) {
	// Create DynamoDB client
	dynamoClient := repository.NewDynamoDBClient()

	// Verify client is not nil
	assert.NotNil(t, dynamoClient)
}

func TestClientContainer_Close(t *testing.T) {
	// Create ClientContainer and set test clients
	container := &repository.ClientContainer{
		PostgreSQL: repository.NewPostgreSQLClient(nil),
		Redis:      repository.NewRedisClient(),
		MongoDB:    repository.NewMongoDBClient(),
		DynamoDB:   repository.NewDynamoDBClient(),
	}

	// Verify Close method doesn't panic
	ctx := context.Background()
	assert.NotPanics(t, func() {
		container.Close(ctx)
	})
}

func TestClose(t *testing.T) {
	// Save original Logger
	originalLogger := log.Logger
	defer func() {
		// Restore original Logger after test
		log.Logger = originalLogger
	}()

	// Set a temporary Logger
	log.Logger = zap.NewNop()

	// Ensure Close function doesn't panic
	ctx := context.Background()
	assert.NotPanics(t, func() {
		repository.Close(ctx)
	})
}
