package service

import (
	"testing"

	"cactus-golang-hexagonal-microservice-boilerplate/adapter/repository"
	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

func TestMain(m *testing.M) {
	// Initialize configuration and logging
	config.Init("../../config", "config")
	log.Init()

	repository.Clients = &repository.ClientContainer{
		PostgreSQL: repository.NewPostgreSQLClient(nil),
		Redis:      repository.NewRedisClient(),
	}

	m.Run()
}
