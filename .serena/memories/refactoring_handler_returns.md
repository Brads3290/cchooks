# Handler Return Refactoring

## Pattern Change
- **OLD**: Handlers return `(result, error)`
- **NEW**: Handlers return single result via interfaces

## Implementation Details
- Each handler type returns an interface (e.g., `PreToolUseResponse` interface, not struct pointer)
- `ErrorResponse` implements ALL response interfaces
- This allows `cchooks.Error(err)` to be returned from any handler

## Example:
```go
// Handler signature
PreToolUse func(context.Context, *PreToolUseEvent) PreToolUseResponse // interface, not *PreToolUseResponse

// Can return either:
return cchooks.Approve() // returns *PreToolUseResponse which implements PreToolUseResponse interface
return cchooks.Error(err) // returns *ErrorResponse which also implements PreToolUseResponse interface
```

## Status
- Need to define response interfaces
- Need to make ErrorResponse implement all interfaces
- Need to update handler signatures to return interfaces