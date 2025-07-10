# API Reference

This document provides a complete reference for all types and functions in the cchooks library.

## Core Types

### Runner

The main struct that handles hook execution.

```go
type Runner struct {
    Raw          func(context.Context, string) *RawResponse
    PreToolUse   func(context.Context, *PreToolUseEvent) PreToolUseResponseInterface
    PostToolUse  func(context.Context, *PostToolUseEvent) PostToolUseResponseInterface
    Notification func(context.Context, *NotificationEvent) NotificationResponseInterface
    Stop         func(context.Context, *StopEvent) StopResponseInterface
    StopOnce     func(context.Context, *StopEvent) StopResponseInterface
    Error        func(ctx context.Context, rawJSON string, err error) *RawResponse
}
```

#### Methods

- `Run()` - Reads from stdin and executes the appropriate handler
- `RunContext(ctx context.Context)` - Like Run but with a custom context

## Event Types

### PreToolUseEvent

```go
type PreToolUseEvent struct {
    SessionID string          `json:"session_id"`
    ToolName  string          `json:"tool_name"`
    ToolInput json.RawMessage `json:"tool_input"`
}
```

#### Methods
- `AsBash() (*BashInput, error)`
- `AsEdit() (*EditInput, error)`
- `AsRead() (*ReadInput, error)`
- `AsWrite() (*WriteInput, error)`

### PostToolUseEvent

```go
type PostToolUseEvent struct {
    SessionID    string          `json:"session_id"`
    ToolName     string          `json:"tool_name"`
    ToolInput    json.RawMessage `json:"tool_input"`
    ToolResponse json.RawMessage `json:"tool_response"`
}
```

#### Methods
- `InputAsBash() (*BashInput, error)`
- `InputAsEdit() (*EditInput, error)`
- `InputAsRead() (*ReadInput, error)`
- `InputAsWrite() (*WriteInput, error)`
- `ResponseAsBash() (*BashOutput, error)`
- `ResponseAsEdit() (*EditOutput, error)`
- `ResponseAsRead() (*ReadOutput, error)`
- `ResponseAsWrite() (*WriteOutput, error)`

### NotificationEvent

```go
type NotificationEvent struct {
    SessionID string `json:"session_id"`
    Message   string `json:"notification_message"`
}
```

### StopEvent

```go
type StopEvent struct {
    SessionID      string            `json:"session_id"`
    StopHookActive bool              `json:"stop_hook_active"`
    TranscriptPath string            `json:"transcript_path"`
    Transcript     []TranscriptEntry // Populated from transcript file
}
```

## Response Types

### Response Interfaces

```go
type PreToolUseResponseInterface interface {
    isPreToolUseResponse()
}

type PostToolUseResponseInterface interface {
    isPostToolUseResponse()
}

type NotificationResponseInterface interface {
    isNotificationResponse()
}

type StopResponseInterface interface {
    isStopResponse()
}
```

### Concrete Response Types

```go
type PreToolUseResponse struct {
    Decision   string `json:"decision,omitempty"`
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

type PostToolUseResponse struct {
    Decision   string `json:"decision,omitempty"`
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

type NotificationResponse struct {
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
}

type StopResponse struct {
    Decision   string `json:"decision,omitempty"`
    Continue   *bool  `json:"continue,omitempty"`
    StopReason string `json:"stopReason,omitempty"`
    Reason     string `json:"reason,omitempty"`
}

type ErrorResponse struct {
    Error   error
    Message string
}

type RawResponse struct {
    ExitCode int
    Output   string
}
```

## Response Helper Functions

### PreToolUse Responses
- `Approve() PreToolUseResponseInterface` - Approve tool execution
- `Block(reason string) PreToolUseResponseInterface` - Block with reason
- `StopClaude(reason string) PreToolUseResponseInterface` - Stop Claude

### PostToolUse Responses
- `Allow() PostToolUseResponseInterface` - Allow results to be shown
- `PostBlock(reason string) PostToolUseResponseInterface` - Block results
- `PostStopClaude(reason string) PostToolUseResponseInterface` - Stop Claude

### Notification Responses
- `OK() NotificationResponseInterface` - Acknowledge notification
- `StopFromNotification(reason string) NotificationResponseInterface` - Stop Claude

### Stop Responses
- `Continue() StopResponseInterface` - Allow Claude to stop
- `BlockStop(reason string) StopResponseInterface` - Prevent stopping

### Error Response
- `Error(err error) *ErrorResponse` - Return an error (implements all interfaces)

## Tool Input/Output Types

### Bash
```go
type BashInput struct {
    Command string `json:"command"`
    Timeout *int   `json:"timeout,omitempty"` // milliseconds
}

type BashOutput struct {
    Output   string `json:"output"`
    ExitCode int    `json:"exit_code"`
}
```

### Edit
```go
type EditInput struct {
    FilePath  string `json:"file_path"`
    OldString string `json:"old_string"`
    NewString string `json:"new_string"`
}

type EditOutput struct {
    Success bool `json:"success"`
}
```

### Read
```go
type ReadInput struct {
    FilePath string `json:"file_path"`
    Offset   *int   `json:"offset,omitempty"`
    Limit    *int   `json:"limit,omitempty"`
}

type ReadOutput struct {
    Content string `json:"content"`
}
```

### Write
```go
type WriteInput struct {
    FilePath string `json:"file_path"`
    Content  string `json:"content"`
}

type WriteOutput struct {
    Success bool `json:"success"`
}
```

## Transcript Types

```go
type TranscriptEntry struct {
    ParentUUID   *string         `json:"parentUuid"`
    UUID         string          `json:"uuid"`
    IsSidechain  bool            `json:"isSidechain"`
    UserType     string          `json:"userType"`
    CWD          string          `json:"cwd"`
    SessionID    string          `json:"sessionId"`
    Version      string          `json:"version"`
    Type         string          `json:"type"`
    Message      json.RawMessage `json:"message"`
    Timestamp    string          `json:"timestamp"`
}

type UserMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type AssistantMessage struct {
    ID         string                  `json:"id"`
    Type       string                  `json:"type"`
    Role       string                  `json:"role"`
    Model      string                  `json:"model"`
    Content    []map[string]interface{} `json:"content"`
    StopReason string                  `json:"stop_reason"`
    Usage      map[string]interface{}   `json:"usage"`
}
```

### TranscriptEntry Methods
- `IsUserMessage() bool`
- `IsAssistantMessage() bool`
- `GetUserMessage() (*UserMessage, error)`
- `GetAssistantMessage() (*AssistantMessage, error)`

## Testing Types

### TestRunner

```go
type TestRunner struct {
    // Internal runner reference
}
```

#### Constructor
- `NewTestRunner(r *Runner) *TestRunner`

#### Test Methods
- `TestPreToolUse(toolName string, toolInput interface{}) PreToolUseResponseInterface`
- `TestPostToolUse(toolName string, toolInput, toolResponse interface{}) PostToolUseResponseInterface`
- `TestNotification(message string) NotificationResponseInterface`
- `TestStop(stopHookActive bool, transcript []TranscriptEntry) StopResponseInterface`

#### Assertion Methods
- `AssertPreToolUseApproves(toolName string, toolInput interface{}) error`
- `AssertPreToolUseBlocks(toolName string, toolInput interface{}) error`
- `AssertPreToolUseBlocksWithReason(toolName string, toolInput interface{}, reason string) error`
- `AssertPreToolUseStopsClaude(toolName string, toolInput interface{}) error`
- `AssertPostToolUseAllows(toolName string, toolInput, toolResponse interface{}) error`
- `AssertPostToolUseBlocks(toolName string, toolInput, toolResponse interface{}) error`
- `AssertPostToolUseBlocksWithReason(toolName string, toolInput, toolResponse interface{}, reason string) error`
- `AssertPostToolUseStopsClaude(toolName string, toolInput, toolResponse interface{}) error`
- `AssertNotificationOK(message string) error`
- `AssertNotificationStopsClaude(message string) error`
- `AssertStopContinues(stopHookActive bool, transcript []TranscriptEntry) error`
- `AssertStopBlocks(stopHookActive bool, transcript []TranscriptEntry) error`
- `AssertStopBlocksWithReason(stopHookActive bool, transcript []TranscriptEntry, reason string) error`

## Constants

### Decision Constants
```go
const (
    PreToolUseApprove = "approve"
    PreToolUseBlock   = "block"
    PostToolUseBlock  = "block"
    StopBlock         = "block"
)
```