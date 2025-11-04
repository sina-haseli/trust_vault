# Trust Vault Plugin Makefile

# Variables
PLUGIN_NAME=trust-vault
PLUGIN_BINARY=trust-vault-plugin
BUILD_DIR=bin
PLUGIN_DIR=/etc/vault/plugins
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"

# Default target
.PHONY: all
all: clean build

# Build the plugin binary
.PHONY: build
build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(PLUGIN_BINARY) ./cmd/trust-vault
	@echo "Build complete: $(BUILD_DIR)/$(PLUGIN_BINARY)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(PLUGIN_BINARY)-linux-amd64 ./cmd/trust-vault
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(PLUGIN_BINARY)-linux-arm64 ./cmd/trust-vault
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(PLUGIN_BINARY)-darwin-amd64 ./cmd/trust-vault
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(PLUGIN_BINARY)-darwin-arm64 ./cmd/trust-vault
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(PLUGIN_BINARY)-windows-amd64.exe ./cmd/trust-vault
	@echo "Multi-platform build complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Calculate SHA256 checksum
.PHONY: checksum
checksum: build
	@echo "Calculating SHA256 checksum..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		shasum -a 256 $(BUILD_DIR)/$(PLUGIN_BINARY) | awk '{print $$1}' > $(BUILD_DIR)/$(PLUGIN_BINARY).sha256; \
	else \
		sha256sum $(BUILD_DIR)/$(PLUGIN_BINARY) | awk '{print $$1}' > $(BUILD_DIR)/$(PLUGIN_BINARY).sha256; \
	fi
	@echo "SHA256: $$(cat $(BUILD_DIR)/$(PLUGIN_BINARY).sha256)"

# Install plugin to Vault plugin directory (requires sudo)
.PHONY: install
install: build checksum
	@echo "Installing plugin to $(PLUGIN_DIR)..."
	@sudo mkdir -p $(PLUGIN_DIR)
	@sudo cp $(BUILD_DIR)/$(PLUGIN_BINARY) $(PLUGIN_DIR)/
	@sudo chmod +x $(PLUGIN_DIR)/$(PLUGIN_BINARY)
	@echo "Plugin installed to $(PLUGIN_DIR)/$(PLUGIN_BINARY)"
	@echo "SHA256: $$(cat $(BUILD_DIR)/$(PLUGIN_BINARY).sha256)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	@echo "Dependencies tidied"

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t trust-vault:latest -f Dockerfile .
	@echo "Docker image built: trust-vault:latest"

# Run Docker container
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker-compose up -d
	@echo "Docker container running"

# Stop Docker container
.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	docker-compose down
	@echo "Docker container stopped"

# Development build with debug symbols
.PHONY: dev
dev:
	@echo "Building development version..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(PLUGIN_BINARY) ./cmd/trust-vault
	@echo "Development build complete: $(BUILD_DIR)/$(PLUGIN_BINARY)"

# Help target
.PHONY: help
help:
	@echo "Trust Vault Plugin - Available targets:"
	@echo "  make build          - Build the plugin binary"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make checksum       - Calculate SHA256 checksum"
	@echo "  make install        - Install plugin to Vault (requires sudo)"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make tidy           - Tidy dependencies"
	@echo "  make deps           - Download dependencies"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make docker-stop    - Stop Docker container"
	@echo "  make dev            - Build development version with debug symbols"
	@echo "  make help           - Show this help message"
