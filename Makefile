# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=portservice
DOCKER_IMAGE=portservice

# Build info
VERSION?=1.0.0
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)
BUILD_FLAGS=-v
DEV_FLAGS=$(BUILD_FLAGS) -race
PROD_FLAGS=$(BUILD_FLAGS) -trimpath -ldflags="-s -w $(LDFLAGS)"

# Output directories
BIN_DIR=bin
DIST_DIR=dist

# Local development directories
LOCAL_BIN=$(BIN_DIR)/local
DEV_BIN=$(LOCAL_BIN)/dev
PROD_BIN=$(LOCAL_BIN)/prod

# Linting
GOLINT=golangci-lint

.PHONY: all build clean test coverage lint docker-build docker-run help mod-download mod-tidy local-dev local-prod run-dev run-prod

all: test build

help: ## Display available commands
	@echo "Available commands:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application for production
	$(GOBUILD) $(PROD_FLAGS) -o $(BINARY_NAME) cmd/portservice/main.go

clean: ## Clean build files
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BIN_DIR)
	rm -rf $(DIST_DIR)

test: ## Run tests
	$(GOTEST) -v ./...

test-race: ## Run tests with race detector
	$(GOTEST) -v -race ./...

coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

coverage-func: ## Show test coverage by function
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

lint: ## Run linter
	$(GOLINT) run

docker-build: ## Build docker image
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-run: ## Run docker container
	docker run -p 8080:8080 $(DOCKER_IMAGE):latest

mod-download: ## Download Go module dependencies
	$(GOMOD) download

mod-tidy: ## Tidy Go module dependencies
	$(GOMOD) tidy

# Local development builds
local-setup: ## Create local build directories
	mkdir -p $(DEV_BIN)
	mkdir -p $(PROD_BIN)

local-dev: local-setup ## Build for local development (with race detection)
	$(GOBUILD) $(DEV_FLAGS) -o $(DEV_BIN)/$(BINARY_NAME) cmd/portservice/main.go

local-prod: local-setup ## Build for local production
	$(GOBUILD) $(PROD_FLAGS) -o $(PROD_BIN)/$(BINARY_NAME) cmd/portservice/main.go

run-dev: local-dev ## Run local development build
	$(DEV_BIN)/$(BINARY_NAME)

run-prod: local-prod ## Run local production build
	$(PROD_BIN)/$(BINARY_NAME)

# Development tools installation
tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker compose commands
compose-up: ## Start all services with docker-compose
	docker-compose up -d

compose-down: ## Stop all services
	docker-compose down

# Database migrations (if needed in future)
migrate-up: ## Run database migrations up
	@echo "No migrations implemented yet"

migrate-down: ## Run database migrations down
	@echo "No migrations implemented yet"

# Generate mocks (if needed)
generate-mocks: ## Generate mocks for testing
	@echo "No mock generation implemented yet"

# Default target
.DEFAULT_GOAL := help 