.PHONY: build test clean run-server run-tui docker help

BINARY_SERVER=bin/curltree-server
BINARY_TUI=bin/curltree-tui
DOCKER_IMAGE=curltree

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
BUILD_FLAGS=-v -ldflags="-s -w"
CGO_ENABLED=1

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build both server and TUI binaries
	@echo "Building server..."
	@mkdir -p bin
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_SERVER) ./cmd/server
	@echo "Building TUI..."
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_TUI) ./cmd/tui
	@echo "Build complete!"

build-server: ## Build only the server binary
	@echo "Building server..."
	@mkdir -p bin
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_SERVER) ./cmd/server

build-tui: ## Build only the TUI binary
	@echo "Building TUI..."
	@mkdir -p bin
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_TUI) ./cmd/tui

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f curltree.db

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

run-server: build-server ## Run the HTTP server
	@echo "Starting HTTP server..."
	@./$(BINARY_SERVER)

run-tui: build-tui ## Run the TUI SSH server
	@echo "Starting TUI SSH server..."
	@./$(BINARY_TUI)

dev-server: ## Run server in development mode with auto-reload
	@echo "Starting development server..."
	@go run ./cmd/server

dev-tui: ## Run TUI in development mode with auto-reload
	@echo "Starting development TUI..."
	@go run ./cmd/tui

run-dev: ## Run both server and TUI in development mode (uses port 23234)
	@echo "Starting development environment..."
	@echo "HTTP server: http://localhost:8080"
	@echo "SSH TUI: ssh -p 23234 localhost"
	@$(MAKE) -j2 dev-server dev-tui

generate-keys: ## Generate SSH host key for development
	@echo "Generating SSH host key..."
	@mkdir -p .ssh
	@ssh-keygen -t rsa -b 4096 -f .ssh/curltree_host_key -N "" -C "curltree-host-key"
	@echo "SSH host key generated: .ssh/curltree_host_key"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build ## Run application in Docker
	@echo "Running Docker container..."
	@docker run -p 8080:8080 -p 23234:23234 $(DOCKER_IMAGE)

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

install: build ## Install binaries to GOPATH/bin
	@echo "Installing binaries..."
	@cp $(BINARY_SERVER) $(GOPATH)/bin/
	@cp $(BINARY_TUI) $(GOPATH)/bin/

demo-data: ## Create demo user data
	@echo "Creating demo data..."
	@curl -X POST http://localhost:8080/api/profiles \
		-H "Content-Type: application/json" \
		-d '{"ssh_public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ demo", "full_name": "Demo User", "username": "demo", "about": "This is a demo profile", "links": [{"name": "Website", "url": "https://example.com"}, {"name": "GitHub", "url": "https://github.com/demo"}]}'

.DEFAULT_GOAL := help