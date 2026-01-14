# Claude Loop Go CLI Makefile

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

MODULE := github.com/DeukWoongWoo/claude-loop
LDFLAGS := -X $(MODULE)/internal/version.Version=$(VERSION) \
           -X $(MODULE)/internal/version.GitCommit=$(GIT_COMMIT) \
           -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)

BINARY := claude-loop
BIN_DIR := bin

.PHONY: all build test test-integration test-e2e test-all lint clean install help

all: build

## build: Build the binary
build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/claude-loop

## test: Run unit tests
test:
	@echo "Running unit tests..."
	go test -v ./internal/...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./test/integration/...

## test-e2e: Run E2E tests
test-e2e: build
	@echo "Running E2E tests..."
	go test -v ./test/e2e/...

## test-all: Run all tests (unit, integration, E2E, golden)
test-all: test test-integration test-e2e golden-test
	@echo "All tests passed!"

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html

## install: Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY)..."
	cp $(BIN_DIR)/$(BINARY) $(GOPATH)/bin/

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

## golden-test: Compare output with golden files
golden-test: build
	@echo "Running golden tests..."
	@./$(BIN_DIR)/$(BINARY) --help > /tmp/help_output.txt 2>&1 || true
	@diff -u test/golden/help.txt /tmp/help_output.txt || (echo "Help output differs from golden file" && exit 1)
	@./$(BIN_DIR)/$(BINARY) --version > /tmp/version_output.txt 2>&1 || true
	@diff -u test/golden/version.txt /tmp/version_output.txt || (echo "Version output differs from golden file" && exit 1)
	@echo "Golden tests passed!"

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
