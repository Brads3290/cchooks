# Advanced Topics

This guide covers advanced usage patterns and features of cchooks.

## Error Handling

### Using the Error Handler

The Error handler allows you to customize how errors are reported:

```go
runner := &cchooks.Runner{
    Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
        // Log to external service
        logger.Error("Hook error", 
            "error", err,
            "input", rawJSON,
        )
        
        // Return custom response
        return &cchooks.RawResponse{
            ExitCode: 42,
            Output: "Hook temporarily unavailable",
        }
    },
}
```

### Error Response in Handlers

All handlers can return errors using the Error response:

```go
PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
    config, err := loadConfig()
    if err != nil {
        return cchooks.Error(fmt.Errorf("failed to load config: %w", err))
    }
    // ... rest of handler
}
```

## Working with Context

The context passed to handlers can be used for timeouts and cancellation:

```go
PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
    // Create a timeout for external calls
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // Make external API call
    allowed, err := checkWithAPI(ctx, event)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return cchooks.Block("External validation timed out")
        }
        return cchooks.Error(err)
    }
    
    if !allowed {
        return cchooks.Block("Not authorized by external service")
    }
    
    return cchooks.Approve()
}
```

## Raw Handler for Custom Events

The Raw handler gives you complete control over event processing:

```go
runner := &cchooks.Runner{
    Raw: func(ctx context.Context, rawJSON string) *cchooks.RawResponse {
        // Parse custom event
        var event map[string]interface{}
        if err := json.Unmarshal([]byte(rawJSON), &event); err != nil {
            return &cchooks.RawResponse{
                ExitCode: 2,
                Output: "Invalid JSON",
            }
        }
        
        // Handle custom event types
        if event["hook_event_name"] == "CustomEvent" {
            // Process custom event
            return &cchooks.RawResponse{
                ExitCode: 0,
                Output: `{"status": "processed"}`,
            }
        }
        
        // Return nil to continue normal processing
        return nil
    },
}
```

## Transcript Analysis

The Stop handler receives transcript data that can be analyzed:

```go
Stop: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
    var codeWritten bool
    var testsRun bool
    
    // Analyze transcript
    for _, entry := range event.Transcript {
        if entry.IsAssistantMessage() {
            msg, _ := entry.GetAssistantMessage()
            
            // Check content blocks
            for _, content := range msg.Content {
                if content["type"] == "text" {
                    text := content["text"].(string)
                    if strings.Contains(text, "```") {
                        codeWritten = true
                    }
                    if strings.Contains(text, "test") || strings.Contains(text, "Test") {
                        testsRun = true
                    }
                }
            }
        }
    }
    
    // Require tests if code was written
    if codeWritten && !testsRun {
        return cchooks.BlockStop("Code was written but no tests were run")
    }
    
    return cchooks.Continue()
}
```

## Security Best Practices

1. **Validate All Inputs**: Don't trust tool inputs
2. **Use Allowlists**: Prefer allowing specific safe operations over blocking dangerous ones
3. **Sanitize Paths**: Prevent directory traversal attacks
4. **Limit Resource Usage**: Implement rate limiting and resource quotas
5. **Log Security Events**: Keep audit trails of blocked operations

Example secure path validation:

```go
func isPathSafe(path string) bool {
    // Clean the path
    cleaned := filepath.Clean(path)
    
    // Must be absolute
    if !filepath.IsAbs(cleaned) {
        return false
    }
    
    // Must be within allowed directories
    allowedDirs := []string{"/home/user/project", "/tmp"}
    for _, dir := range allowedDirs {
        if strings.HasPrefix(cleaned, dir) {
            return true
        }
    }
    
    return false
}
```

## Debugging Hooks

### Development Mode

Create a debug mode for your hooks:

```go
func main() {
    debug := os.Getenv("HOOK_DEBUG") == "true"
    
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
            if debug {
                log.Printf("PreToolUse: %s - %s", event.ToolName, string(event.ToolInput))
            }
            
            // Handler logic...
        },
        
        Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
            if debug {
                // In debug mode, output full error details
                return &cchooks.RawResponse{
                    ExitCode: 2,
                    Output: fmt.Sprintf("DEBUG ERROR: %v\nInput: %s", err, rawJSON),
                }
            }
            return nil // Use default error handling
        },
    }
    
    runner.Run()
}
```

### Testing Panics

The Runner recovers from panics in handlers, but you can test panic behavior:

```go
func TestPanicRecovery(t *testing.T) {
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
            panic("test panic")
        },
        Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
            if !strings.Contains(err.Error(), "panic: test panic") {
                t.Errorf("expected panic error, got: %v", err)
            }
            return nil
        },
    }
    
    // Test will verify the panic is caught and passed to Error handler
}
```