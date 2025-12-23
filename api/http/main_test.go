package http

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"cactus-golang-hexagonal-microservice-boilerplate/adapter/dependency"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/postgre"
	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository/redis"
	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

var ctx = context.Background()

func TestMain(m *testing.M) {
	// Parse command line arguments, support -short flag
	flag.Parse()

	// Initialize configuration and logging
	config.Init("../../config", "config")
	log.Init()

	// Skip integration tests in short mode
	if testing.Short() {
		fmt.Println("Skipping integration tests in short mode")
		os.Exit(0)
		return
	}

	// Use test containers
	t := &testing.T{}
	postgresConfig := postgre.SetupPostgreSQLContainer(t)
	redisConfig := redis.SetupRedisContainer(t)

	// Set global config to use test containers
	config.GlobalConfig.Postgre = postgresConfig
	config.GlobalConfig.Redis = redisConfig

	// Initialize repositories using dependency injection
	clients, err := dependency.InitializeRepositories(
		dependency.WithPostgres(),
		dependency.WithRedis(),
	)
	if err != nil {
		log.SugaredLogger.Fatalf("Failed to initialize repositories: %v", err)
	}
	repository.Clients = clients

	// Initialize services using dependency injection
	svcs, err := dependency.InitializeServices(ctx, clients, nil, dependency.WithUserService())
	if err != nil {
		log.SugaredLogger.Fatalf("Failed to initialize services: %v", err)
	}

	// Register services for API handlers
	RegisterServices(svcs)

	// Run tests
	os.Exit(m.Run())
}
