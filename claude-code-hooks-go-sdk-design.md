# Claude Code Hooks Go SDK Design

## Overview

This document outlines the design for a Go SDK that enables developers to create strongly typed Claude Code hooks. The SDK simplifies the creation of individual hook binaries that handle Claude Code events with type safety and testing utilities.

## Key Design Principles

1. **Individual Hook Binaries**: Each hook is a standalone binary that handles one or more event types
2. **Settings.json Handles Routing**: Tool matching and event routing stays in Claude Code's configuration
3. **Type Safety**: Strong typing for all Claude Code tool inputs with compile-time validation
4. **Simple I/O**: stdin JSON parsing and stdout/stderr response handling
5. **Easy Testing**: Built-in testing utilities for hook validation

## Core Architecture

### 1. Runner with Handler Registry

```go
// Runner handles event dispatch and I/O for a single hook binary
type Runner struct {
    PreToolUse  func(context.Context, *PreToolUseEvent) (*PreToolUseResponse, error)
    PostToolUse func(context.Context, *PostToolUseEvent) (*PostToolUseResponse, error)
    Notification func(context.Context, *NotificationEvent) (*NotificationResponse, error)
    Stop        func(context.Context, *StopEvent) (*StopResponse, error)
}

// Run reads from stdin, dispatches to appropriate handler, outputs response
func (r *Runner) Run(ctx context.Context) error {
    // Read and parse JSON from stdin
    var rawEvent map[string]interface{}
    if err := json.NewDecoder(os.Stdin).Decode(&rawEvent); err != nil {
        return fmt.Errorf("failed to decode stdin: %w", err)
    }
    
    event, ok := rawEvent["event"].(string)
    if !ok {
        return fmt.Errorf("missing or invalid event field")
    }
    
    // Dispatch to appropriate handler
    switch event {
    case "PreToolUse":
        return r.handlePreToolUse(ctx, rawEvent)
    case "PostToolUse":
        return r.handlePostToolUse(ctx, rawEvent)
    case "Notification":
        return r.handleNotification(ctx, rawEvent)
    case "Stop":
        return r.handleStop(ctx, rawEvent)
    default:
        return fmt.Errorf("unknown event type: %s", event)
    }
}
```

### 2. Event Types

```go
// Event types - data containers for each hook event
type PreToolUseEvent struct {
    SessionID string          `json:"session_id"`
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"`
}

type PostToolUseEvent struct {
    SessionID    string          `json:"session_id"`
    ToolName     string          `json:"tool_name"`
    ToolInput    json.RawMessage `json:"tool_input"`
    ToolResponse json.RawMessage `json:"tool_response"`
}

type NotificationEvent struct {
    SessionID string `json:"session_id"`
    Message   string `json:"notification_message"`
}

type StopEvent struct {
    SessionID      string        `json:"session_id"`
    StopHookActive bool          `json:"stop_hook_active"`
    Transcript     []interface{} `json:"transcript"`
}
```

### 3. Response Types

```go
// Response types with event-specific decision options
type PreToolUseResponse struct {
    Decision   string `json:"decision,omitempty"`    // "approve" or "block"
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

type PostToolUseResponse struct {
    Decision   string `json:"decision,omitempty"`    // "block" only
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

type NotificationResponse struct {
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
}

type StopResponse struct {
    Decision   string `json:"decision,omitempty"`    // "block" only
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

// Constants for decisions
const (
    PreToolUseApprove = "approve"
    PreToolUseBlock   = "block"
    PostToolUseBlock  = "block"
    StopBlock         = "block"
)
```

### 4. Tool Input Types

```go
// Strongly typed tool input structures
type BashInput struct {
    Command     string `json:"command" validate:"required"`
    Timeout     *int   `json:"timeout,omitempty" validate:"omitempty,max=600000"`
    Description string `json:"description,omitempty"`
}

type EditInput struct {
    FilePath   string `json:"file_path" validate:"required,filepath"`
    OldString  string `json:"old_string" validate:"required"`
    NewString  string `json:"new_string" validate:"required,nefield=OldString"`
    ReplaceAll bool   `json:"replace_all,omitempty"`
}

type GlobInput struct {
    Pattern string `json:"pattern" validate:"required"`
    Path    string `json:"path,omitempty" validate:"omitempty,dirpath"`
}

type TodoWriteInput struct {
    Todos []TodoItem `json:"todos" validate:"required,min=1,dive"`
}

type TodoItem struct {
    Content  string       `json:"content" validate:"required,min=1"`
    Status   TodoStatus   `json:"status" validate:"required,oneof=pending in_progress completed"`
    Priority TodoPriority `json:"priority" validate:"required,oneof=high medium low"`
    ID       string       `json:"id" validate:"required"`
}

// ... additional tool input types for all 15+ Claude Code tools
```

### 5. Tool Input Parsing

```go
// Tool input parsing methods on events
func (e *PreToolUseEvent) AsBash() (*BashInput, error) {
    var input BashInput
    return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsEdit() (*EditInput, error) {
    var input EditInput
    return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsGlob() (*GlobInput, error) {
    var input GlobInput
    return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsTodoWrite() (*TodoWriteInput, error) {
    var input TodoWriteInput
    return &input, json.Unmarshal(e.ToolInput, &input)
}

// PostToolUse can access both input and response
func (e *PostToolUseEvent) InputAsBash() (*BashInput, error) {
    var input BashInput
    return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsEdit() (*EditInput, error) {
    var input EditInput
    return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) ResponseAsBash() (*BashOutput, error) {
    var output BashOutput
    return &output, json.Unmarshal(e.ToolResponse, &output)
}
```

### 6. Response Helper Functions

```go
// Helper functions for common responses
func Approve() *PreToolUseResponse {
    return &PreToolUseResponse{Decision: PreToolUseApprove}
}

func Block(reason string) *PreToolUseResponse {
    return &PreToolUseResponse{Decision: PreToolUseBlock, Reason: reason}
}

func PostBlock(reason string) *PostToolUseResponse {
    return &PostToolUseResponse{Decision: PostToolUseBlock, Reason: reason}
}

func StopClaude(reason string) *PreToolUseResponse {
    cont := false
    return &PreToolUseResponse{Continue: &cont, StopReason: reason}
}

func Allow() *PostToolUseResponse {
    return &PostToolUseResponse{} // Empty response = allow
}

func OK() *NotificationResponse {
    return &NotificationResponse{} // Empty response = continue
}

func Continue() *StopResponse {
    return &StopResponse{} // Empty response = allow stop
}

func BlockStop(reason string) *StopResponse {
    return &StopResponse{Decision: StopBlock, Reason: reason}
}
```

### 7. Testing Framework

```go
// Testing utilities
type TestRunner struct {
    runner *Runner
}

func NewTestRunner(runner *Runner) *TestRunner {
    return &TestRunner{runner: runner}
}

func (t *TestRunner) TestPreToolUse(toolName string, toolInput interface{}) (*PreToolUseResponse, error) {
    inputJSON, err := json.Marshal(toolInput)
    if err != nil {
        return nil, err
    }
    
    event := &PreToolUseEvent{
        SessionID: "test-session",
        ToolName:  toolName,
        ToolInput: inputJSON,
    }
    
    return t.runner.PreToolUse(context.Background(), event)
}

func (t *TestRunner) TestPostToolUse(toolName string, toolInput, toolResponse interface{}) (*PostToolUseResponse, error) {
    inputJSON, _ := json.Marshal(toolInput)
    responseJSON, _ := json.Marshal(toolResponse)
    
    event := &PostToolUseEvent{
        SessionID:    "test-session", 
        ToolName:     toolName,
        ToolInput:    inputJSON,
        ToolResponse: responseJSON,
    }
    
    return t.runner.PostToolUse(context.Background(), event)
}

// Test assertion helpers
func (t *TestRunner) AssertPreToolUseApproves(toolName string, toolInput interface{}) error {
    resp, err := t.TestPreToolUse(toolName, toolInput)
    if err != nil {
        return err
    }
    if resp.Decision != PreToolUseApprove {
        return fmt.Errorf("expected approve, got %s", resp.Decision)
    }
    return nil
}

func (t *TestRunner) AssertPreToolUseBlocks(toolName string, toolInput interface{}) error {
    resp, err := t.TestPreToolUse(toolName, toolInput)
    if err != nil {
        return err
    }
    if resp.Decision != PreToolUseBlock {
        return fmt.Errorf("expected block, got %s", resp.Decision)
    }
    return nil
}

func (t *TestRunner) AssertPostToolUseAllows(toolName string, toolInput, toolResponse interface{}) error {
    resp, err := t.TestPostToolUse(toolName, toolInput, toolResponse)
    if err != nil {
        return err
    }
    if resp.Decision != "" {
        return fmt.Errorf("expected allow (empty decision), got %s", resp.Decision)
    }
    return nil
}
```

## Usage Examples

### Example 1: Security and Code Formatting Hook

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os/exec"
    "strings"
    
    cchooks "github.com/your-org/claude-code-hooks"
)

func main() {
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            switch event.ToolName {
            case "Bash":
                bash, err := event.AsBash()
                if err != nil {
                    return nil, err
                }
                
                // Block dangerous commands
                dangerous := []string{"rm -rf", "sudo rm", "dd if=", ":(){ :|: & };:"}
                for _, pattern := range dangerous {
                    if strings.Contains(bash.Command, pattern) {
                        return cchooks.Block(fmt.Sprintf("Dangerous command pattern detected: %s", pattern)), nil
                    }
                }
                
                return cchooks.Approve(), nil
                
            case "Edit", "Write":
                edit, err := event.AsEdit()
                if err != nil {
                    return nil, err
                }
                
                // Block editing production files
                if strings.Contains(edit.FilePath, "/production/") {
                    return cchooks.Block("Cannot edit production files"), nil
                }
                
                return cchooks.Approve(), nil
                
            default:
                return cchooks.Approve(), nil
            }
        },
        
        PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
            // Auto-format code after edits
            if event.ToolName == "Edit" || event.ToolName == "Write" {
                edit, err := event.InputAsEdit()
                if err == nil {
                    switch {
                    case strings.HasSuffix(edit.FilePath, ".go"):
                        exec.Command("gofmt", "-w", edit.FilePath).Run()
                    case strings.HasSuffix(edit.FilePath, ".js") || strings.HasSuffix(edit.FilePath, ".ts"):
                        exec.Command("prettier", "--write", edit.FilePath).Run()
                    }
                }
            }
            
            return cchooks.Allow(), nil
        },
        
        Notification: func(ctx context.Context, event *cchooks.NotificationEvent) (*cchooks.NotificationResponse, error) {
            // Send desktop notification
            cmd := exec.Command("osascript", "-e", 
                fmt.Sprintf(`display notification "%s" with title "Claude Code"`, event.Message))
            cmd.Run()
            
            return cchooks.OK(), nil
        },
        
        Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
            log.Printf("Claude session %s stopped", event.SessionID)
            return cchooks.Continue(), nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Example 2: Testing the Hook

```go
package main

import (
    "testing"
    "github.com/stretchr/testify/assert"
    cchooks "github.com/your-org/claude-code-hooks"
)

func TestSecurityHook(t *testing.T) {
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            if event.ToolName == "Bash" {
                bash, err := event.AsBash()
                if err != nil {
                    return nil, err
                }
                
                if strings.Contains(bash.Command, "rm -rf") {
                    return cchooks.Block("Dangerous command detected"), nil
                }
                
                return cchooks.Approve(), nil
            }
            return cchooks.Approve(), nil
        },
    }
    
    tester := cchooks.NewTestRunner(runner)
    
    // Test dangerous command is blocked
    dangerousInput := &cchooks.BashInput{Command: "rm -rf /"}
    err := tester.AssertPreToolUseBlocks("Bash", dangerousInput)
    assert.NoError(t, err)
    
    // Test safe command is approved
    safeInput := &cchooks.BashInput{Command: "ls -la"}
    err = tester.AssertPreToolUseApproves("Bash", safeInput)
    assert.NoError(t, err)
}
```

### Example 3: Claude Code Configuration

The corresponding `settings.json` configuration would look like:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/your/security-hook-binary"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command", 
            "command": "/path/to/your/security-hook-binary"
          }
        ]
      }
    ],
    "Notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/your/security-hook-binary"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/your/security-hook-binary"
          }
        ]
      }
    ]
  }
}
```

## Key Benefits

1. **Type Safety**: Compile-time validation for all Claude Code tool inputs
2. **Simple Architecture**: Each hook is a standalone binary with clear responsibilities
3. **Easy Testing**: Built-in testing framework with assertion helpers
4. **Flexible Responses**: Support for all Claude Code hook response patterns
5. **Tool Coverage**: Strongly typed support for all 15+ Claude Code tools
6. **MCP Support**: Extensible design for Model Context Protocol tools
7. **Error Handling**: Proper error propagation and exit code management
8. **Documentation**: Self-documenting code with clear interfaces

## Implementation Notes

- **Response Handling**: Empty responses use exit codes, non-empty responses use JSON output
- **Error Propagation**: Exit code 2 sends errors to Claude, other codes show to user
- **Tool Validation**: Input validation using struct tags and custom validators
- **Context Support**: All handlers receive context for cancellation and timeouts
- **Extensibility**: Easy to add new tool types and response patterns

This design provides a clean, type-safe, and maintainable way to create Claude Code hooks in Go while maintaining the flexibility and power of the underlying hook system.