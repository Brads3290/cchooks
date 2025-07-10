# Suggested Commands for Claude Code Hooks Go SDK

## Development Commands

### Testing
```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out -o coverage.html

# Run tests for a specific package
go test -v ./internal/tools

# Run with verbose output
go test -v ./...
```

### Formatting and Linting
```bash
# Format all Go code
go fmt ./...
gofmt -w .

# Run linter (requires golangci-lint)
golangci-lint run

# Install golangci-lint if not present
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Building
```bash
# Build a hook binary
go build -o my-hook main.go

# Build all examples
make examples

# Build with race detector
go build -race -o my-hook main.go
```

### Makefile Targets
```bash
# Run fmt, lint, and test
make

# Run tests only
make test

# Build examples
make examples

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean

# Install dev dependencies
make dev-deps

# Test an example hook
make run-example
```

## macOS/Darwin Specific Commands
```bash
# List files (macOS ls has different flags than GNU ls)
ls -la

# Find files
find . -name "*.go"

# Search in files (use ripgrep if available, otherwise grep)
grep -r "pattern" .
rg "pattern"

# Git commands work the same
git status
git diff
git commit -m "message"
```

## Hook Testing Commands
```bash
# Test a hook with JSON input
echo '{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}' | ./my-hook

# Test with a file
./my-hook < test-event.json

# Debug with environment variable
HOOK_DEBUG=true ./my-hook < test-event.json
```