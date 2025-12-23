.PHONY: fmt lint test

init:
	@echo "=== Init Go Project with Pre-commit Hooks ==="
	brew install go
	brew install node
	brew install pre-commit
	brew install npm
	brew install golangci-lint
	brew upgrade golangci-lint
	npm install -g @commitlint/cli @commitlint/config-conventional

	@echo "=== Setup Pre-commit ==="
	pre-commit install
	@echo "=== Done.  ==="

fmt:
	go fmt ./...
	goimports -w -local "cactus-golang-hexagonal-microservice-boilerplate" ./

test:
	@echo "=== Prepare Dependency ==="
	go mod tidy
	@echo "=== Start Unit Test ==="
	go test -v -race -cover ./...

pre-commit.install:
	@echo "=== Setup Pre-commit ==="
	pre-commit install

precommit.rehook:
	@echo "=== Rehook Pre-commit ==="
	pre-commit autoupdate
	pre-commit install --install-hooks
	pre-commit install --hook-type commit-msg

ci.lint:
	@echo "=== Start CI Linter ==="
	golangci-lint run -v ./... --fix

all: fmt ci.lint test
