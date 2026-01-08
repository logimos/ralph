.PHONY: build install install-local run clean test help jq-tested jq-untested jq-status release-major release-minor release-patch release

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

# Version management
# Try to get version from git tag, strip any suffix like -dirty, -1-g1234567, etc.
# Falls back to "dev" if git is not available or no tags exist
# This ensures semantic versioning works correctly for both local builds and CI/CD
GIT_DESCRIBE = $(shell git describe --tags --always --dirty 2>/dev/null)
VERSION ?= $(shell if [ -n "$(GIT_DESCRIBE)" ]; then echo "$(GIT_DESCRIBE)" | sed 's/-.*//'; else echo "dev"; fi)
GIT_TAG = $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
VERSION_MAJOR = $(shell echo $(GIT_TAG) | sed 's/v\([0-9]*\).*/\1/' | grep -q '^[0-9]' && echo $(shell echo $(GIT_TAG) | sed 's/v\([0-9]*\).*/\1/') || echo "0")
VERSION_MINOR = $(shell echo $(GIT_TAG) | sed 's/v[0-9]*\.\([0-9]*\).*/\1/' | grep -q '^[0-9]' && echo $(shell echo $(GIT_TAG) | sed 's/v[0-9]*\.\([0-9]*\).*/\1/') || echo "0")
VERSION_PATCH = $(shell echo $(GIT_TAG) | sed 's/v[0-9]*\.[0-9]*\.\([0-9]*\).*/\1/' | grep -q '^[0-9]' && echo $(shell echo $(GIT_TAG) | sed 's/v[0-9]*\.[0-9]*\.\([0-9]*\).*/\1/') || echo "0")

# Build flags
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"
BUILD_DIR=.
RELEASE_DIR=dist

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

## install-local: Install current local version for testing
install-local: build
	@echo "Installing local version of $(BINARY_NAME)..."
	@GOPATH_BIN=$$(go env GOPATH)/bin; \
	if [ -n "$$(go env GOBIN)" ]; then \
		INSTALL_DIR=$$(go env GOBIN); \
	else \
		INSTALL_DIR=$$GOPATH_BIN; \
	fi; \
	echo "Installing to $$INSTALL_DIR"; \
	cp $(BINARY_NAME) $$INSTALL_DIR/$(BINARY_NAME); \
	echo "✓ Installed to $$INSTALL_DIR/$(BINARY_NAME)"

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
	rm -rf $(RELEASE_DIR)
	rm -f .version
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

## release-major: Create a major release (builds binaries for GitHub)
release-major:
	@$(MAKE) release VERSION_TYPE=major

## release-minor: Create a minor release (builds binaries for GitHub)
release-minor:
	@$(MAKE) release VERSION_TYPE=minor

## release-patch: Create a patch release (builds binaries for GitHub)
release-patch:
	@$(MAKE) release VERSION_TYPE=patch

## release: Build release binaries for GitHub (use release-major/minor/patch)
release:
	@if [ -z "$(VERSION_TYPE)" ]; then \
		echo "Error: VERSION_TYPE not set. Use release-major, release-minor, or release-patch"; \
		exit 1; \
	fi
	@echo "Creating $(VERSION_TYPE) release..."
	@$(MAKE) _calculate-version VERSION_TYPE=$(VERSION_TYPE)
	@$(MAKE) _build-release
	@echo ""
	@echo "✓ Release binaries built in $(RELEASE_DIR)/"
	@echo "Next steps:"
	@echo "  1. Review the binaries in $(RELEASE_DIR)/"
	@echo "  2. Create and push the git tag:"
	@echo "     git tag v$$(cat .version)"
	@echo "     git push origin v$$(cat .version)"
	@echo "  3. GitHub Actions will automatically create the release and upload binaries"

_calculate-version:
	@MAJOR=$(VERSION_MAJOR); \
	MINOR=$(VERSION_MINOR); \
	PATCH=$(VERSION_PATCH); \
	if [ -z "$$MAJOR" ] || [ "$$MAJOR" = "v0.0.0" ]; then MAJOR=0; fi; \
	if [ -z "$$MINOR" ]; then MINOR=0; fi; \
	if [ -z "$$PATCH" ]; then PATCH=0; fi; \
	case "$(VERSION_TYPE)" in \
		major) \
			NEW_MAJOR=$$((MAJOR + 1)); \
			NEW_VERSION="v$$NEW_MAJOR.0.0"; \
			;; \
		minor) \
			NEW_MINOR=$$((MINOR + 1)); \
			NEW_VERSION="v$$MAJOR.$$NEW_MINOR.0"; \
			;; \
		patch) \
			NEW_PATCH=$$((PATCH + 1)); \
			NEW_VERSION="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
			;; \
		*) \
			echo "Error: Invalid VERSION_TYPE. Use major, minor, or patch"; \
			exit 1; \
			;; \
	esac; \
	echo $$NEW_VERSION > .version; \
	echo "New version: $$NEW_VERSION"

_build-release:
	@VERSION=$$(cat .version); \
	echo "Building release binaries for version $$VERSION..."; \
	rm -rf $(RELEASE_DIR); \
	mkdir -p $(RELEASE_DIR); \
	echo "Building for linux/amd64..."; \
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -ldflags "-s -w -X main.Version=$$VERSION" -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64 $(BINARY_NAME).go; \
	echo "Building for linux/arm64..."; \
	GOOS=linux GOARCH=arm64 $(GO_BUILD) -ldflags "-s -w -X main.Version=$$VERSION" -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64 $(BINARY_NAME).go; \
	echo "Building for darwin/amd64..."; \
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) -ldflags "-s -w -X main.Version=$$VERSION" -o $(RELEASE_DIR)/$(BINARY_NAME)-darwin-amd64 $(BINARY_NAME).go; \
	echo "Building for darwin/arm64..."; \
	GOOS=darwin GOARCH=arm64 $(GO_BUILD) -ldflags "-s -w -X main.Version=$$VERSION" -o $(RELEASE_DIR)/$(BINARY_NAME)-darwin-arm64 $(BINARY_NAME).go; \
	echo "Building for windows/amd64..."; \
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -ldflags "-s -w -X main.Version=$$VERSION" -o $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.exe $(BINARY_NAME).go; \
	echo "Creating checksums..."; \
	cd $(RELEASE_DIR) && sha256sum * > checksums.txt; \
	echo "✓ Release build complete"

## help: Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
	@echo ""
	@echo "Examples:"
	@echo "  make build                                    # Build the binary"
	@echo "  make install                                  # Install to GOPATH/bin"
	@echo "  make install-local                            # Install local version for testing"
	@echo "  make run ITERATIONS=5                         # Run with 5 iterations"
	@echo "  make run ITERATIONS=3 ARGS='-verbose'        # Run with verbose output"
	@echo "  make run ITERATIONS=3 ARGS='-agent cursor-agent -verbose'  # Run with custom options"
	@echo "  make test                                     # Run tests"
	@echo "  make lint                                     # Format and vet code"
	@echo "  make clean                                    # Remove build artifacts"
	@echo "  make jq-tested                                # List tested features"
	@echo "  make jq-untested                              # List untested features"
	@echo "  make jq-status                                # Show both tested and untested"
	@echo "  make release-major                            # Create major release (v1.0.0 -> v2.0.0)"
	@echo "  make release-minor                            # Create minor release (v1.0.0 -> v1.1.0)"
	@echo "  make release-patch                            # Create patch release (v1.0.0 -> v1.0.1)"

