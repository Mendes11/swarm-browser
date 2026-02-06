.PHONY: all build run test clean install uninstall release release-snapshot dev fmt vet lint help

# Variables
BINARY_NAME=swarm-browser
MAIN_PATH=main.go
INSTALL_PATH=/usr/local/bin
GO=go
GORELEASER=goreleaser

# Version information (for local builds)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(BUILD_DATE) \
	-X main.builtBy=make"

# Default target
all: build

## help: Display this help message
help:
	@echo "Swarm Browser - Makefile Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^##/ { \
		help_line = $$0; \
		getline; \
		if (match($$0, /^[a-zA-Z_-]+:/)) { \
			target = substr($$0, 1, index($$0, ":")-1); \
			sub(/^## /, "", help_line); \
			printf "  %-20s %s\n", target, help_line; \
		} \
	}' $(MAKEFILE_LIST) | sort

## build: Build the binary for the current platform
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: ./$(BINARY_NAME)"

## run: Run the application
run: build
	./$(BINARY_NAME)

## dev: Run in development mode with dev clusters
dev:
	@if [ -f "run-dev.sh" ]; then \
		./run-dev.sh; \
	else \
		$(GO) run $(MAIN_PATH); \
	fi

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

## fmt: Format all Go source files
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Formatting complete"

## vet: Run go vet on all packages
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet complete"

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## tidy: Run go mod tidy to clean up dependencies
tidy:
	@echo "Tidying module dependencies..."
	$(GO) mod tidy
	@echo "Module dependencies tidied"

## deps: Download module dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	@echo "Dependencies downloaded"

## clean: Remove built binaries and temporary files
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -f debug.log
	@echo "Clean complete"

## install: Install the binary to system path (requires sudo)
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@if [ -w "$(INSTALL_PATH)" ]; then \
		cp $(BINARY_NAME) $(INSTALL_PATH)/; \
		echo "Installation complete"; \
	else \
		echo "Error: $(INSTALL_PATH) is not writable. Try: sudo make install"; \
		exit 1; \
	fi

## uninstall: Remove the binary from system path (may require sudo)
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PATH)..."
	@if [ -f "$(INSTALL_PATH)/$(BINARY_NAME)" ]; then \
		if [ -w "$(INSTALL_PATH)" ]; then \
			rm -f $(INSTALL_PATH)/$(BINARY_NAME); \
			echo "Uninstall complete"; \
		else \
			echo "Error: Cannot remove $(INSTALL_PATH)/$(BINARY_NAME). Try: sudo make uninstall"; \
			exit 1; \
		fi; \
	else \
		echo "$(BINARY_NAME) not found in $(INSTALL_PATH)"; \
	fi

## release: Create a new release using goreleaser
release:
	@if command -v $(GORELEASER) >/dev/null 2>&1; then \
		echo "Creating release with goreleaser..."; \
		$(GORELEASER) release --clean; \
	else \
		echo "Error: goreleaser not installed. Install with:"; \
		echo "  go install github.com/goreleaser/goreleaser/v2@latest"; \
		echo "or"; \
		echo "  brew install goreleaser"; \
		exit 1; \
	fi

## release-snapshot: Build release artifacts locally without publishing
release-snapshot:
	@if command -v $(GORELEASER) >/dev/null 2>&1; then \
		echo "Building snapshot release..."; \
		$(GORELEASER) release --snapshot --clean; \
		echo ""; \
		echo "Snapshot artifacts created in dist/"; \
		echo ""; \
		ls -lh dist/*.tar.gz 2>/dev/null || true; \
	else \
		echo "Error: goreleaser not installed. Install with:"; \
		echo "  go install github.com/goreleaser/goreleaser/v2@latest"; \
		echo "or"; \
		echo "  brew install goreleaser"; \
		exit 1; \
	fi

## release-dry: Test the release process without publishing
release-dry:
	@if command -v $(GORELEASER) >/dev/null 2>&1; then \
		echo "Running goreleaser dry run..."; \
		$(GORELEASER) release --skip=publish --clean; \
	else \
		echo "Error: goreleaser not installed. Install with:"; \
		echo "  go install github.com/goreleaser/goreleaser/v2@latest"; \
		echo "or"; \
		echo "  brew install goreleaser"; \
		exit 1; \
	fi

## check-release: Validate goreleaser configuration
check-release:
	@if command -v $(GORELEASER) >/dev/null 2>&1; then \
		echo "Checking goreleaser configuration..."; \
		$(GORELEASER) check; \
	else \
		echo "Error: goreleaser not installed. Install with:"; \
		echo "  go install github.com/goreleaser/goreleaser/v2@latest"; \
		echo "or"; \
		echo "  brew install goreleaser"; \
		exit 1; \
	fi

## version: Display version information of built binary
version: build
	./$(BINARY_NAME) --version

## watch: Watch for file changes and rebuild (requires entr)
watch:
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -c make build; \
	else \
		echo "Error: entr not installed. Install with:"; \
		echo "  brew install entr  # macOS"; \
		echo "  apt-get install entr  # Debian/Ubuntu"; \
		exit 1; \
	fi

## update-deps: Update all dependencies to latest versions
update-deps:
	@echo "Updating dependencies to latest versions..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "Dependencies updated"

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

## ci: Run continuous integration checks
ci: deps check
	@echo "CI checks completed successfully!"

## tools: Install required development tools
tools:
	@echo "Installing development tools..."
	@echo "Installing goreleaser..."
	go install github.com/goreleaser/goreleaser/v2@latest
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

## tag: Create and push a new version tag
tag:
	@read -p "Enter version (e.g., v0.1.2): " version; \
	if [ -z "$$version" ]; then \
		echo "Error: Version cannot be empty"; \
		exit 1; \
	fi; \
	echo "Creating tag $$version..."; \
	git tag -a $$version -m "Release $$version"; \
	echo "Tag created. To push: git push origin $$version"; \
	echo "Or to push with commits: git push origin main $$version"

## changelog: Display commit history since last tag
changelog:
	@echo "Changes since last tag:"
	@echo ""
	@git log $$(git describe --tags --abbrev=0 2>/dev/null || echo "")..HEAD --oneline

# Default target when just running 'make'
.DEFAULT_GOAL := help