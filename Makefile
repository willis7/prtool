# Variables
BINARY_NAME=prtool
VERSION=v0.1.0
BUILD_DIR=build
GOFILES=$(shell find . -name '*.go' -type f)

# Default target
.PHONY: all
all: test build

# Build the binary
.PHONY: build
build:
	go build -o $(BINARY_NAME) .

# Run tests
.PHONY: test
test:
	go test ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run e2e tests
.PHONY: test-e2e
test-e2e:
	go test -v ./internal/e2e/...

# Run linters
.PHONY: lint
lint:
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf $(BUILD_DIR)

# Install the binary
.PHONY: install
install: build
	go install

# Create a new tag (does not push)
.PHONY: tag
tag:
	@echo "Creating tag $(VERSION)"
	@echo "Current tags:"
	@git tag -l
	@echo ""
	@echo "To create the tag, run:"
	@echo "  git tag -a $(VERSION) -m 'Release $(VERSION)'"
	@echo ""
	@echo "To push the tag, run:"
	@echo "  git push origin $(VERSION)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Run the binary
.PHONY: run
run: build
	./$(BINARY_NAME) run

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Run tests and build the binary (default)"
	@echo "  build        - Build the binary"
	@echo "  test         - Run all tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  test-e2e     - Run end-to-end tests"
	@echo "  lint         - Run linters"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install the binary"
	@echo "  tag          - Show instructions for creating tag $(VERSION)"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  run          - Build and run the binary"
	@echo "  help         - Show this help message"