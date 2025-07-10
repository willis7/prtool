.PHONY: test build clean vet lint check ci

# Default target
all: check

# Run tests
test:
	go test ./...

# Build the binary
build:
	go build -o bin/prtool ./main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Run go vet
vet:
	go vet ./...

# Run golangci-lint (if installed)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Tidy dependencies
tidy:
	go mod tidy

# Run comprehensive checks (for CI)
check: tidy vet test lint

# CI target - runs all checks and builds
ci: check build

# Install golangci-lint (for development)
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
