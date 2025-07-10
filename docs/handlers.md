# Handler Types

cchooks supports several types of event handlers that correspond to different stages of Claude Code's execution.

## PreToolUse Handler

Called before Claude executes a tool. You can approve, modify, or block the tool execution.

```go
PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
    // Your logic here
    return cchooks.Approve()
}
```

### Response Options

- `cchooks.Approve()` - Allow the tool to execute
- `cchooks.Block(reason)` - Prevent execution with a reason
- `cchooks.StopClaude(reason)` - Stop Claude entirely
- `cchooks.Error(err)` - Return an error

## PostToolUse Handler

Called after a tool executes. You can inspect results and potentially block Claude from seeing them.

```go
PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
    bash, _ := event.ResponseAsBash()
    if bash.ExitCode != 0 {
        return cchooks.PostBlock("command failed")
    }
    return cchooks.Allow()
}
```

### Response Options

- `cchooks.Allow()` - Let Claude see the results
- `cchooks.PostBlock(reason)` - Hide results from Claude
- `cchooks.PostStopClaude(reason)` - Stop Claude entirely
- `cchooks.Error(err)` - Return an error

## Notification Handler

Called for various notifications during Claude's execution.

```go
Notification: func(ctx context.Context, event *cchooks.NotificationEvent) cchooks.NotificationResponseInterface {
    if event.Message == "Task completed" {
        // Log or process notification
    }
    return cchooks.OK()
}
```

### Response Options

- `cchooks.OK()` - Acknowledge the notification
- `cchooks.StopFromNotification(reason)` - Stop Claude
- `cchooks.Error(err)` - Return an error

## Stop Handler

Called when Claude is about to stop. You can allow or prevent the stop.

```go
Stop: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
    if !taskComplete {
        return cchooks.BlockStop("task not complete")
    }
    return cchooks.Continue()
}
```

### Response Options

- `cchooks.Continue()` - Allow Claude to stop
- `cchooks.BlockStop(reason)` - Prevent Claude from stopping
- `cchooks.Error(err)` - Return an error

## StopOnce Handler

A special variant of the Stop handler that's only called the first time Claude tries to stop (when `stop_hook_active` is false).

```go
StopOnce: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
    // Special handling for first stop attempt
    return cchooks.BlockStop("Please confirm you want to stop")
}
```

If both `Stop` and `StopOnce` are defined, `StopOnce` takes precedence when `stop_hook_active` is false.

## Raw Handler

An advanced handler that receives the raw JSON input before any parsing. Useful for custom event handling.

```go
Raw: func(ctx context.Context, rawJSON string) *cchooks.RawResponse {
    // Custom processing
    return &cchooks.RawResponse{
        ExitCode: 0,
        Output: "custom response",
    }
}
```

If Raw returns nil, normal event processing continues.

## Error Handler

Called when any error occurs during hook execution. Allows custom error handling.

```go
Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
    // Log error or custom handling
    return nil // Use default error handling
}
```