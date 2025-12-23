package tests

import (
	"testing"

	"cactus-golang-hexagonal-microservice-boilerplate/config"
)

func TestMain(m *testing.M) {
	config.Init("../config", "config")

	m.Run()
}
