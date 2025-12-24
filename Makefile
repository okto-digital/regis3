# regis3 Makefile

# Variables
BINARY_NAME := regis3
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
PKG := github.com/okto-digital/regis3/internal/cli
LDFLAGS := -ldflags "-s -w -X $(PKG).version=$(VERSION) -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(BUILD_DATE)"

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := gofmt

# Directories
CMD_DIR := ./cmd/regis3
BIN_DIR := ./bin
DIST_DIR := ./dist

.PHONY: all build install test test-cover lint fmt clean release snapshot help

# Default target
all: fmt lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BIN_DIR)/$(BINARY_NAME)"

# Install to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) $(CMD_DIR)

# Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -v -race ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -w .

# Check formatting
fmt-check:
	@echo "Checking formatting..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Files need formatting:" && $(GOFMT) -l . && exit 1)

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not installed, skipping..." && exit 0)
	golangci-lint run

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out coverage.html

# Cross-platform release builds (manual)
release:
	@echo "Building release binaries..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Release binaries in: $(DIST_DIR)/"

# Build snapshot with goreleaser (for testing)
snapshot:
	@echo "Building snapshot with goreleaser..."
	goreleaser build --snapshot --clean

# Full release with goreleaser
goreleaser:
	@echo "Running goreleaser..."
	goreleaser release --clean

# Run the application
run:
	@$(GOCMD) run $(CMD_DIR)

# Show help
help:
	@echo "regis3 Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       - Build the binary to ./bin/"
	@echo "  make install     - Install to GOPATH/bin"
	@echo "  make test        - Run all tests"
	@echo "  make test-cover  - Run tests with coverage report"
	@echo "  make test-race   - Run tests with race detector"
	@echo "  make fmt         - Format code with gofmt"
	@echo "  make fmt-check   - Check if code needs formatting"
	@echo "  make lint        - Run golangci-lint"
	@echo "  make tidy        - Tidy go.mod dependencies"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make release     - Build cross-platform release binaries"
	@echo "  make snapshot    - Build snapshot with goreleaser"
	@echo "  make goreleaser  - Full release with goreleaser"
	@echo "  make run         - Run the application"
	@echo "  make all         - Format, lint, test, and build"
