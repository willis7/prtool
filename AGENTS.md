# Agent Guidelines for prtool

## Build/Lint/Test Commands

### Building
- `make build` - Build the binary
- `make build-all` - Build for multiple architectures (Linux/macOS/Windows)
- `go build -o bin/prtool ./main.go` - Direct build command

### Testing
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage
- `go test ./...` - Run all tests directly
- `go test -run TestName ./path/to/package` - Run a single test
- `go test -v ./path/to/package` - Run tests with verbose output

### Linting and Code Quality
- `make lint` - Run golangci-lint (if installed)
- `make vet` - Run go vet
- `make check` - Run comprehensive checks (vet + test + lint)
- `golangci-lint run` - Direct linting command
- `go vet ./...` - Direct vet command
- `go mod tidy` - Clean up dependencies

## Code Style Guidelines

### General Principles
- Follow idiomatic Go practices from Effective Go, Code Review Comments, and Google's Go Style Guide
- Write simple, clear, and idiomatic Go code
- Favor clarity and simplicity over cleverness
- Keep the happy path left-aligned (minimize indentation)
- Return early to reduce nesting
- Make the zero value useful
- Document exported types, functions, methods, and packages

### Naming Conventions
- **Packages**: Use lowercase, single-word names (avoid `util`, `common`, `base`)
- **Variables/Functions**: Use mixedCaps (camelCase), keep short but descriptive
- **Exported names**: Start with capital letter
- **Unexported names**: Start with lowercase letter
- **Interfaces**: Use -er suffix when possible (e.g., `Reader`, `Writer`)
- **Constants**: MixedCaps for exported, mixedCaps for unexported
- **Struct tags**: Use for JSON, YAML, database mappings

### Code Formatting
- Always use `gofmt` to format code
- Use `goimports` to manage imports automatically
- Keep line length reasonable (no hard limit, but consider readability)
- Add blank lines to separate logical groups of code

### Comments
- Write comments in complete sentences
- Start sentences with the name of the thing being described
- Package comments start with "Package [name]"
- Use line comments (`//`) for most comments
- Document why, not what, unless the what is complex

### Error Handling
- Check errors immediately after function calls
- Don't ignore errors using `_` unless documented why
- Wrap errors with context using `fmt.Errorf` with `%w` verb
- Create custom error types when needed for specific error checking
- Place error returns as the last return value
- Name error variables `err`
- Keep error messages lowercase and don't end with punctuation

### Architecture
- Follow standard Go project layout
- Keep `main` packages in `cmd/` directory
- Put reusable packages in `pkg/` or `internal/`
- Use `internal/` for packages that shouldn't be imported externally
- Group related functionality into packages
- Avoid circular dependencies
- Accept interfaces, return concrete types

### Imports
- Group imports: standard library, third-party, internal
- Use blank lines between groups
- Remove unused imports with `go mod tidy`

### Type Safety
- Define types to add meaning and type safety
- Use pointers for large structs or when modifying receiver
- Use values for small structs and immutability
- Prefer explicit type conversions
- Use type assertions carefully with second return value check

### Concurrency
- Don't create goroutines in libraries (let caller control concurrency)
- Always know how goroutines will exit
- Use `sync.WaitGroup` or channels to wait for goroutines
- Close channels from sender side, not receiver
- Use `select` for non-blocking operations
- Prefer channels over mutexes when possible

### Testing
- Keep tests in same package (white-box testing)
- Use `_test` package suffix for black-box testing
- Name test files with `_test.go` suffix
- Use table-driven tests for multiple test cases
- Name tests descriptively: `Test_functionName_scenario`
- Use subtests with `t.Run` for better organization
- Test both success and error cases
- Mark helper functions with `t.Helper()`
- Clean up resources using `t.Cleanup()`

### Security
- Validate all external input
- Use strong typing to prevent invalid states
- Sanitize data for different contexts (HTML, SQL, shell)
- Use standard library crypto packages
- Don't implement custom cryptography
- Store passwords using bcrypt or similar
- Use TLS for network communication