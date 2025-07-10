# Claude Code Hooks Go SDK

A Go SDK for creating strongly typed Claude Code hooks. This SDK simplifies the creation of individual hook binaries that handle Claude Code events with type safety and testing utilities.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Events](#events)
  - [Responses](#responses)
  - [Tool Parsing](#tool-parsing)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Advanced Features](#advanced-features)
  - [Raw Handler](#raw-handler)
  - [Stateful Hooks](#stateful-hooks)
  - [StopOnce Handler](#stoponce-handler)
  - [External Service Integration](#external-service-integration)
- [API Reference](#api-reference)
- [Examples](#examples)

## Installation

```bash
go get github.com/brads3290/cchooks
```

## Quick Start

Create a simple hook that approves all tool usage:

```go
package main

import (
    "context"
    "log"
    cchooks "github.com/brads3290/cchooks"
)

func main() {
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            return cchooks.Approve(), nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

Build and test:

```bash
go build -o my-hook main.go
echo '{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}' | ./my-hook
```

## Core Concepts

### Events

Claude Code hooks receive four event types:

- **PreToolUse**: Called before tool execution, can approve/block/stop
- **PostToolUse**: Called after tool execution with the result
- **Notification**: Receives Claude notifications
- **Stop**: Called when Claude is stopping

Implement handlers for the events you want to process:

```go
runner := &cchooks.Runner{
    PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
        log.Printf("Tool: %s, Session: %s", event.ToolName, event.SessionID)
        return cchooks.Approve(), nil
    },
    PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
        return cchooks.Allow(), nil
    },
    Notification: func(ctx context.Context, event *cchooks.NotificationEvent) (*cchooks.NotificationResponse, error) {
        return cchooks.OK(), nil
    },
    Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
        return cchooks.Continue(), nil
    },
}
```

### Responses

The SDK provides helper functions for common responses:

```go
// PreToolUse responses
cchooks.Approve()           // Allow the tool to run
cchooks.Block(reason)       // Block with reason
cchooks.StopClaude(reason)  // Stop Claude entirely

// PostToolUse responses  
cchooks.Allow()             // Continue normally
cchooks.PostBlock(reason)   // Block after execution
cchooks.StopClaudePost(reason)

// Notification responses
cchooks.OK()                // Acknowledge
cchooks.StopFromNotification(reason)

// Stop responses
cchooks.Continue()          // Allow stop
cchooks.BlockStop(reason)   // Prevent stop
```

### Tool Parsing

Events provide typed parsing for all Claude Code tools:

```go
func handlePreToolUse(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
    switch event.ToolName {
    case "Bash":
        bash, err := event.AsBash()
        if err != nil {
            return nil, err
        }
        // bash.Command, bash.Timeout, etc.
        
    case "Edit":
        edit, err := event.AsEdit()
        if err != nil {
            return nil, err
        }
        // edit.FilePath, edit.OldString, edit.NewString
        
    case "Write":
        write, err := event.AsWrite()
        if err != nil {
            return nil, err
        }
        // write.FilePath, write.Content
    }
    
    return cchooks.Approve(), nil
}
```

All 15+ Claude Code tools are supported with full type safety.

## Error Handling

The SDK uses exit codes to communicate with Claude Code:

- **0**: Success
- **2**: Error sent to Claude (default for errors)
- **Other**: Error shown to user

Implement custom error handling with the Error handler:

```go
runner := &cchooks.Runner{
    PreToolUse: handlePreToolUse,
    Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
        log.Printf("Error: %v, JSON: %s", err, rawJSON)
        
        // Return nil for default handling (exit code 2, or 0 for Stop events)
        // Or return custom response:
        return &cchooks.RawResponse{
            ExitCode: 1,
            Output:   "Custom error message",
        }
    },
}
```

The Error handler receives:
- JSON parsing errors
- Event validation errors
- Handler errors
- Panics (converted to errors with "panic:" prefix)

**Note**: Stop event errors use exit code 0 by default to avoid blocking Claude from stopping.

## Testing

The SDK includes comprehensive testing utilities:

```go
func TestSecurityHook(t *testing.T) {
    runner := createSecurityRunner()
    tester := cchooks.NewTestRunner(runner)
    
    // Test specific behaviors
    err := tester.AssertPreToolUseBlocks("Bash", &cchooks.BashInput{
        Command: "rm -rf /",
    })
    assert.NoError(t, err)
    
    err = tester.AssertPreToolUseApproves("Bash", &cchooks.BashInput{
        Command: "ls -la",
    })
    assert.NoError(t, err)
    
    // Test with raw input
    output, exitCode, err := tester.RunWithInput(`{
        "event": "PreToolUse",
        "session_id": "test",
        "tool_name": "Bash",
        "tool_input": {"command": "pwd"}
    }`)
    assert.Equal(t, 0, exitCode)
}
```

## Advanced Features

### Raw Handler

Process events before JSON parsing for complete control:

```go
runner := &cchooks.Runner{
    Raw: func(ctx context.Context, rawJSON string) (*cchooks.RawResponse, error) {
        // Log all events
        log.Printf("[RAW] %s", rawJSON)
        
        // Block specific patterns
        if strings.Contains(rawJSON, "FORBIDDEN") {
            return &cchooks.RawResponse{
                ExitCode: 1,
                Output:   "Forbidden pattern detected",
            }, nil
        }
        
        // Continue normal processing
        return nil, nil
    },
    PreToolUse: handlePreToolUse,
}
```

### Stateful Hooks

Track state across multiple tool invocations:

```go
type RateLimiter struct {
    mu       sync.Mutex
    counts   map[string]int
    window   map[string]time.Time
}

func (r *RateLimiter) CheckAndIncrement(sessionID string) bool {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    now := time.Now()
    if last, ok := r.window[sessionID]; ok && now.Sub(last) < time.Minute {
        r.counts[sessionID]++
        if r.counts[sessionID] > 100 {
            return false // Rate limit exceeded
        }
    } else {
        r.counts[sessionID] = 1
        r.window[sessionID] = now
    }
    return true
}

// Use in your hook:
rateLimiter := &RateLimiter{
    counts: make(map[string]int),
    window: make(map[string]time.Time),
}

runner := &cchooks.Runner{
    PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
        if !rateLimiter.CheckAndIncrement(event.SessionID) {
            return cchooks.Block("Rate limit exceeded"), nil
        }
        return cchooks.Approve(), nil
    },
}
```

### StopOnce Handler

The `StopOnce` handler allows you to handle the first stop event differently from subsequent ones. It only triggers when `stop_hook_active` is false:

```go
runner := &cchooks.Runner{
    // StopOnce only triggers when stop_hook_active is false (first stop)
    StopOnce: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
        // This will only run on the first stop event
        // You could use this to save state, send notifications, etc.
        log.Printf("First stop detected! Session: %s\n", event.SessionID)
        
        // Block the first stop to allow cleanup or confirmation
        return cchooks.BlockStop("Please confirm you want to stop"), nil
    },
    
    // Regular Stop handler for subsequent stops (when stop_hook_active is true)
    Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
        // This runs on all subsequent stop attempts
        log.Printf("Stop attempt for session: %s\n", event.SessionID)
        
        // Allow subsequent stops
        return cchooks.Continue(), nil
    },
}
```

**Behavior:**
- If both `Stop` and `StopOnce` are defined, `StopOnce` takes precedence when `stop_hook_active` is false
- If only `StopOnce` is defined, it will only handle the first stop event
- If only `Stop` is defined, it handles all stop events regardless of `stop_hook_active`

### External Service Integration

Integrate with security services or policy engines:

```go
type PolicyClient struct {
    endpoint string
    client   *http.Client
}

func (p *PolicyClient) CheckPolicy(ctx context.Context, event *cchooks.PreToolUseEvent) (bool, string, error) {
    payload, _ := json.Marshal(map[string]interface{}{
        "tool":       event.ToolName,
        "session_id": event.SessionID,
        "input":      event.ToolInput,
    })
    
    req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewReader(payload))
    if err != nil {
        return false, "", err
    }
    
    resp, err := p.client.Do(req)
    if err != nil {
        return false, "", err
    }
    defer resp.Body.Close()
    
    var result struct {
        Allowed bool   `json:"allowed"`
        Reason  string `json:"reason"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return false, "", err
    }
    
    return result.Allowed, result.Reason, nil
}
```


## API Reference

### Core Types

#### Runner

The main type for creating hooks:

```go
type Runner struct {
    // Raw handler - called before any parsing
    Raw          func(context.Context, string) (*RawResponse, error)
    
    // Event handlers
    PreToolUse   func(context.Context, *PreToolUseEvent) (*PreToolUseResponse, error)
    PostToolUse  func(context.Context, *PostToolUseEvent) (*PostToolUseResponse, error)
    Notification func(context.Context, *NotificationEvent) (*NotificationResponse, error)
    Stop         func(context.Context, *StopEvent) (*StopResponse, error)
    
    // StopOnce - called for Stop events only when stop_hook_active is false
    StopOnce     func(context.Context, *StopEvent) (*StopResponse, error)
    
    // Error handler - called on any SDK error
    Error        func(ctx context.Context, rawJSON string, err error) *RawResponse
}
```

#### Event Types

```go
// PreToolUseEvent - before tool execution
type PreToolUseEvent struct {
    Event     string          `json:"event"`      // Always "PreToolUse"
    SessionID string          `json:"session_id"`
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"`
}

// PostToolUseEvent - after tool execution
type PostToolUseEvent struct {
    Event        string          `json:"event"`         // Always "PostToolUse"
    SessionID    string          `json:"session_id"`
    ToolName     string          `json:"tool_name"`
    ToolInput    json.RawMessage `json:"tool_input"`
    ToolResponse json.RawMessage `json:"tool_response"`
}

// NotificationEvent - Claude notifications
type NotificationEvent struct {
    Event                string `json:"event"`                 // Always "Notification"
    SessionID            string `json:"session_id"`
    NotificationMessage  string `json:"notification_message"`
}

// StopEvent - when Claude is stopping
type StopEvent struct {
    Event          string        `json:"event"`            // Always "Stop"
    SessionID      string        `json:"session_id"`
    StopHookActive bool          `json:"stop_hook_active"`
    Transcript     []interface{} `json:"transcript"`
}
```

#### Response Types

```go
// PreToolUseResponse
type PreToolUseResponse struct {
    Decision   string `json:"decision,omitempty"`    // "approve", "block", or empty
    Continue   *bool  `json:"continue,omitempty"`    // For stop decisions
    StopReason string `json:"stopReason,omitempty"`  // Reason for stopping
    Reason     string `json:"reason,omitempty"`      // Reason for blocking
}

// PostToolUseResponse
type PostToolUseResponse struct {
    Decision   string `json:"decision,omitempty"`
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

// NotificationResponse
type NotificationResponse struct {
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
}

// StopResponse
type StopResponse struct {
    Decision   string `json:"decision,omitempty"`
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

// RawResponse - for Raw and Error handlers
type RawResponse struct {
    ExitCode int    // Process exit code
    Output   string // Stdout output
}
```

### Response Helper Functions

```go
// PreToolUse helpers
func Approve() *PreToolUseResponse
func Block(reason string) *PreToolUseResponse
func StopClaude(reason string) *PreToolUseResponse

// PostToolUse helpers
func Allow() *PostToolUseResponse
func PostBlock(reason string) *PostToolUseResponse
func StopClaudePost(reason string) *PostToolUseResponse

// Notification helpers
func OK() *NotificationResponse
func StopFromNotification(reason string) *NotificationResponse

// Stop helpers
func Continue() *StopResponse
func BlockStop(reason string) *StopResponse
```

### Tool Input Types

All tool input types are available with corresponding parsing methods:

```go
// Parsing methods on PreToolUseEvent
func (e *PreToolUseEvent) AsBash() (*BashInput, error)
func (e *PreToolUseEvent) AsEdit() (*EditInput, error)
func (e *PreToolUseEvent) AsWrite() (*WriteInput, error)
func (e *PreToolUseEvent) AsRead() (*ReadInput, error)
func (e *PreToolUseEvent) AsMultiEdit() (*MultiEditInput, error)
func (e *PreToolUseEvent) AsGlob() (*GlobInput, error)
func (e *PreToolUseEvent) AsGrep() (*GrepInput, error)
func (e *PreToolUseEvent) AsLS() (*LSInput, error)
func (e *PreToolUseEvent) AsWebSearch() (*WebSearchInput, error)
func (e *PreToolUseEvent) AsWebFetch() (*WebFetchInput, error)
func (e *PreToolUseEvent) AsNotebookEdit() (*NotebookEditInput, error)
func (e *PreToolUseEvent) AsNotebookRead() (*NotebookReadInput, error)
func (e *PreToolUseEvent) AsTodoRead() (*TodoReadInput, error)
func (e *PreToolUseEvent) AsTodoWrite() (*TodoWriteInput, error)
func (e *PreToolUseEvent) AsTask() (*TaskInput, error)
func (e *PreToolUseEvent) AsExitPlanMode() (*ExitPlanModeInput, error)

// Similar methods exist on PostToolUseEvent for both input and response
func (e *PostToolUseEvent) InputAsBash() (*BashInput, error)
func (e *PostToolUseEvent) ResponseAsBash() (*BashResponse, error)
// ... and so on for all tools
```

### Testing API

```go
// Create a test runner
func NewTestRunner(runner *Runner) *TestRunner

// Test methods
func (t *TestRunner) RunWithInput(input string) (output string, exitCode int, err error)
func (t *TestRunner) RunPreToolUse(toolName string, toolInput interface{}) (*PreToolUseResponse, error)
func (t *TestRunner) RunPostToolUse(toolName string, toolInput, toolResponse interface{}) (*PostToolUseResponse, error)

// Assertion helpers
func (t *TestRunner) AssertPreToolUseApproves(toolName string, toolInput interface{}) error
func (t *TestRunner) AssertPreToolUseBlocks(toolName string, toolInput interface{}) error
func (t *TestRunner) AssertPreToolUseStops(toolName string, toolInput interface{}) error
func (t *TestRunner) AssertPostToolUseAllows(toolName string, toolInput, toolResponse interface{}) error
func (t *TestRunner) AssertPostToolUseBlocks(toolName string, toolInput, toolResponse interface{}) error
```

### Exit Codes

- **0**: Success (hook executed successfully, or Error handler returned nil for Stop events)
- **2**: Error sent to Claude (default for Error handler returning nil on non-Stop events)  
- **Other**: Error shown to user (custom exit codes via RawResponse)

### Configuration

Configure hooks in Claude Code's settings:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/your/hook-binary"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/post-hook"
          }
        ]
      }
    ]
  }
}
```

### Best Practices

1. **Always log to stderr**: stdout is reserved for hook responses
2. **Handle errors gracefully**: Return errors from handlers rather than panicking
3. **Test thoroughly**: Use the testing utilities to verify all code paths
4. **Keep hooks focused**: Each hook should have a single, clear purpose
5. **Fail safely**: Decide whether to fail open (approve) or closed (block) when external services are unavailable

## Examples

Complete working examples are available in the `examples/` directory:

- `security-hook`: Production-ready security controls
- `format-hook`: Auto-formatting for various file types
- `simple-hook`: Minimal example with logging
- `advanced-hook`: Demonstrates all features including Raw and Error handlers

## Contributing

Contributions are welcome! Please submit pull requests or issues on GitHub.

## License

MIT License - see LICENSE file for details.