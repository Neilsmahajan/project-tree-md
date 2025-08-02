# Project Tree Markdown - Makefile

# Variables
BINARY_NAME=project-tree-md
MAIN_PATH=cmd/main.go
BUILD_DIR=.

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete"

# Run the application with current directory
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME) on current directory..."
	./$(BINARY_NAME)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

# Install to system (requires sudo on some systems)
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Uninstall from system
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from /usr/local/bin..."
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Development target - format, vet, test, and build
.PHONY: dev
dev: fmt vet test build

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Build and run on current directory"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  vet        - Vet code"
	@echo "  deps       - Install dependencies"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  install    - Install binary to /usr/local/bin"
	@echo "  uninstall  - Remove binary from /usr/local/bin"
	@echo "  dev        - Run fmt, vet, test, and build"
	@echo "  help       - Show this help message"
