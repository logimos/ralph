.PHONY: build install run clean test help jq-tested jq-untested jq-status

# Variables
BINARY_NAME=ralph
PLAN_FILE?=plan.json
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_INSTALL=$(GO_CMD) install
GO_TEST=$(GO_CMD) test
GO_FMT=$(GO_CMD) fmt
GO_VET=$(GO_CMD) vet
GO_MOD=$(GO_CMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_DIR=.

# Default target
.DEFAULT_GOAL := help

## build: Build the ralph binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO_BUILD) $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_NAME).go
	@echo "Build complete: ./$(BINARY_NAME)"

## install: Install ralph to $GOPATH/bin or $GOBIN
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO_INSTALL) $(LDFLAGS) $(BINARY_NAME).go
	@echo "Installation complete"

## run: Run ralph with arguments (requires -iterations flag)
run:
	@if [ -z "$(ITERATIONS)" ]; then \
		echo "Error: ITERATIONS is required. Example: make run ITERATIONS=5"; \
		echo "Additional args can be passed: make run ITERATIONS=5 ARGS='-verbose -agent cursor-agent'"; \
		exit 1; \
	fi
	./$(BINARY_NAME) -iterations $(ITERATIONS) $(ARGS)

## test: Run tests
test:
	@echo "Running tests..."
	$(GO_TEST) -v ./...

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GO_FMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO_VET) ./...

## lint: Run fmt and vet
lint: fmt vet

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	@echo "Clean complete"

## tidy: Tidy go.mod
tidy:
	@echo "Tidying go.mod..."
	$(GO_MOD) tidy

## jq-tested: List tested features (id, category, description)
jq-tested:
	@echo "=== Tested Features (from $(PLAN_FILE)) ==="
	@jq -r '.[] | select(.tested == true) | "\(.id) | \(.category) | \(.description)"' $(PLAN_FILE) | column -t -s '|' || echo "No tested features found"

## jq-untested: List untested features (id, category, description)
jq-untested:
	@echo "=== Untested Features (from $(PLAN_FILE)) ==="
	@jq -r '.[] | select(.tested == false) | "\(.id) | \(.category) | \(.description)"' $(PLAN_FILE) | column -t -s '|' || echo "No untested features found"

## jq-status: Show both tested and untested features
jq-status: jq-tested jq-untested

## help: Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
	@echo ""
	@echo "Examples:"
	@echo "  make build                                    # Build the binary"
	@echo "  make install                                  # Install to GOPATH/bin"
	@echo "  make run ITERATIONS=5                         # Run with 5 iterations"
	@echo "  make run ITERATIONS=3 ARGS='-verbose'        # Run with verbose output"
	@echo "  make run ITERATIONS=3 ARGS='-agent cursor-agent -verbose'  # Run with custom options"
	@echo "  make test                                     # Run tests"
	@echo "  make lint                                     # Format and vet code"
	@echo "  make clean                                    # Remove build artifacts"
	@echo "  make jq-tested                                # List tested features"
	@echo "  make jq-untested                              # List untested features"
	@echo "  make jq-status                                # Show both tested and untested"

