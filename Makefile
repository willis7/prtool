.PHONY: test build clean vet lint check ci release snapshot

# Default target
all: check

# Run tests
test:
	go test ./...

# Build the binary
build:
	go build -o bin/prtool ./main.go

# Build for multiple architectures
build-all: clean
	mkdir -p dist/
	@echo "Building for multiple architectures..."
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/willis7/prtool/cmd.version=$(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)'" -o dist/prtool-linux-amd64 ./main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/willis7/prtool/cmd.version=$(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)'" -o dist/prtool-linux-arm64 ./main.go
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/willis7/prtool/cmd.version=$(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)'" -o dist/prtool-darwin-amd64 ./main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/willis7/prtool/cmd.version=$(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)'" -o dist/prtool-darwin-arm64 ./main.go
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/willis7/prtool/cmd.version=$(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)'" -o dist/prtool-windows-amd64.exe ./main.go
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w -X 'github.com/willis7/prtool/cmd.version=$(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)'" -o dist/prtool-windows-arm64.exe ./main.go
	
	@echo "Build complete! Binaries available in dist/"

# Clean build artifacts
clean:
	rm -rf bin/ dist/

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

# Release with GoReleaser (requires tag and GITHUB_TOKEN)
release:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "goreleaser not installed. Install with: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 1; \
	fi
	goreleaser release --clean

# Snapshot release (no publish)
snapshot:
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "goreleaser not installed. Install with: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 1; \
	fi
	goreleaser release --clean --snapshot --skip=publish,sign

# Install golangci-lint (for development)
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
