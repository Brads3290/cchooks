# Task Completion Checklist

When completing any coding task in the Claude Code Hooks Go SDK, ensure you:

## 1. Code Quality
- [ ] Code follows Go idioms and best practices
- [ ] All exported types/functions have proper GoDoc comments
- [ ] No unnecessary comments in code (documentation belongs in GoDoc)
- [ ] Error handling is consistent and descriptive

## 2. Formatting and Linting
```bash
# Always run before considering task complete:
go fmt ./...
golangci-lint run
```

## 3. Testing
```bash
# Ensure all tests pass
go test -v -race ./...

# Check test coverage if adding new functionality
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 4. Verification Steps
- [ ] New features have corresponding tests
- [ ] Examples are updated if API changes
- [ ] No build warnings or linter errors
- [ ] Integration tests pass if core functionality changed

## 5. Documentation
- [ ] Update README.md if adding new features
- [ ] Update doc.go if changing core API
- [ ] Add example code for new functionality

## Important Notes
- The linter is configured to continue on error (`|| true` in Makefile)
- Always fix linter warnings even though they don't block the build
- Use `make` to run the full validation suite (fmt, lint, test)