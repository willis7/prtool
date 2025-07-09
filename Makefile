.PHONY: test build clean vet lint

# Default target
all: test

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
	golangci-lint run

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Tidy dependencies
tidy:
	go mod tidy
