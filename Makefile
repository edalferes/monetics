# Simple Makefile for Monetics project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
BINARY_NAME=./bin/monetics
CMD_PATH=./cmd/api

.PHONY: help build test test-integration cover clean run docker-build swagger lint mocks

## Help
help: ## Show available commands
	@echo 'Available commands:'
	@echo '  build             - Build the application'
	@echo '  test              - Run unit tests'
	@echo '  test-integration  - Run integration tests (build tag: integration)'
	@echo '  cover             - Run tests with coverage report'
	@echo '  run               - Run the application'
	@echo '  clean             - Clean build artifacts'
	@echo '  docker-build      - Build Docker image'
	@echo '  swagger           - Generate Swagger documentation'
	@echo '  lint              - Run golangci-lint'
	@echo '  mocks             - Regenerate testify mocks (mockery)'

## Build
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p ./bin
	@$(GOBUILD) -o $(BINARY_NAME) $(CMD_PATH)

## Swagger
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@$(GOCMD) run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go --parseDependency --parseInternal -o docs/openapi

## Test
test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) ./...

## Lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	@$(shell go env GOPATH)/bin/golangci-lint run --config=.golangci.yml

## Run
run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	@$(GOCMD) run $(CMD_PATH)

## Clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

## Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME) .