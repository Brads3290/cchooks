# Code Style and Conventions

## Go Language Version
- Go 1.24.4

## Package Structure
- Main package: `cchooks`
- Internal packages: `internal/tools` for tool-specific types
- Test files follow Go convention: `*_test.go`
- Integration tests in separate package: `cchooks_test`

## Naming Conventions
- Exported types/functions: PascalCase (e.g., `Runner`, `PreToolUseEvent`)
- Unexported types/functions: camelCase (e.g., `outputResponse`, `isEmpty`)
- Methods: PascalCase for exported, camelCase for unexported
- Constants: Not heavily used, follow Go conventions when needed

## Documentation
- Package documentation in `doc.go` with comprehensive examples
- All exported types and functions have GoDoc comments
- Comments start with the name of the item being documented
- Example code included in documentation for clarity

## Error Handling
- Functions return `error` as last return value
- Custom error types not heavily used, standard errors preferred
- Error messages are descriptive and lowercase

## Testing
- Unit tests alongside code files
- Integration tests in separate `_test` package
- Test functions follow `Test*` naming convention
- Helper test types like `TestRunner` provided for users

## Code Organization
- One main type per file where practical
- Related functionality grouped together
- Clear separation between public API and internal implementation