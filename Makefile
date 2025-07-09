.PHONY: test build examples clean fmt lint

# Default target
all: fmt lint test

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build all examples
examples:
	go build -o bin/security-hook ./examples/security-hook
	go build -o bin/format-hook ./examples/format-hook
	go build -o bin/simple-hook ./examples/simple-hook

# Format code
fmt:
	go fmt ./...
	gofmt -w .

# Run linter
lint:
	golangci-lint run || true

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run a specific example hook for testing
run-example:
	@echo '{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}' | go run ./examples/simple-hook