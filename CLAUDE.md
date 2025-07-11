# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

prtool is a Go CLI application that fetches merged GitHub pull requests and generates AI-powered summaries. It uses the Cobra framework for CLI handling and supports multiple LLM providers.

## Development Commands

### Build and Test
```bash
# Build the binary
go build ./cmd/prtool

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/github

# Run a single test
go test -run TestFunctionName ./path/to/package

# Run static analysis
go vet ./...

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run
```

### Development Workflow
```bash
# Format code
go fmt ./...

# Update dependencies
go mod tidy

# Download dependencies
go mod download

# Run the tool locally
go run ./cmd/prtool run
```

## Architecture

### Package Structure
- `cmd/prtool/` - Entry point and CLI setup
- `internal/` - Core business logic (private packages)
  - `config/` - Configuration loading and validation
  - `github/` - GitHub API client and PR fetching
  - `llm/` - LLM abstraction and implementations
  - `output/` - Markdown rendering and formatting
  - `timerange/` - Time range parsing utilities

### Key Design Patterns
1. **Interface-based LLM abstraction** - The LLM system uses interfaces to support multiple providers (OpenAI, Ollama)
2. **Three-tier configuration** - CLI flags override environment variables which override YAML config
3. **Scope resolution** - Smart logic to resolve org/team/user/repo scopes based on available GitHub permissions
4. **Golden file testing** - Use testdata/ directories with .golden files for output validation

### Configuration Hierarchy
1. CLI flags (highest priority)
2. Environment variables (PRTOOL_*)
3. YAML config file (lowest priority)

### Error Handling
- Use wrapped errors with context: `fmt.Errorf("failed to X: %w", err)`
- Return errors to the CLI layer for proper exit codes
- Use cobra.Command.SilenceUsage for command execution errors

### Testing Guidelines
- Write table-driven tests for complex logic
- Use interfaces for mocking external dependencies
- Create testdata/ directories for golden file tests
- Mock GitHub API responses for integration tests

## Implementation Status

The project is following a phased implementation plan. Check todo.md for current progress and next steps. Key phases:
- P0-P3: Core infrastructure
- P4-P7: GitHub integration and output
- P8-P9: LLM integration
- P10-P11: Polish and CI/CD