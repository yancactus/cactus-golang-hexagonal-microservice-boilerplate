package middleware

import (
	"os"
	"testing"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
	"cactus-golang-hexagonal-microservice-boilerplate/util/log"
)

func TestMain(m *testing.M) {
	config.Init("../../../config", "config")
	log.Init()

	exitCode := m.Run()
	os.Exit(exitCode)
}
