# Claude Code Hooks Go SDK

A Go SDK for creating strongly typed Claude Code hooks. This SDK simplifies the creation of individual hook binaries that handle Claude Code events with type safety and testing utilities.

## Table of Contents

- [Installation](#installation)
- [Tutorial](#tutorial)
  - [Your First Hook](#your-first-hook)
  - [Understanding Events](#understanding-events)
  - [Working with Different Tools](#working-with-different-tools)
  - [Handling Errors](#handling-errors)
  - [Testing Your Hooks](#testing-your-hooks)
  - [Advanced Patterns](#advanced-patterns)
- [API Reference](#api-reference)

## Installation

```bash
go get github.com/brads3290/cchooks
```

## Tutorial

### Your First Hook

Let's start with the simplest possible hook - one that approves all actions:

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
            // Approve every tool use
            return cchooks.Approve(), nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

Build and test your hook:

```bash
# Build the hook
go build -o my-first-hook main.go

# Test it with sample input
echo '{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}' | ./my-first-hook

# Should output nothing (empty response = approve)
```

### Understanding Events

Claude Code hooks receive four types of events. Let's add logging to see what we're working with:

```go
package main

import (
    "context"
    "log"
    "os"
    
    cchooks "github.com/brads3290/cchooks"
)

func main() {
    // Create a logger that writes to stderr (stdout is reserved for responses)
    logger := log.New(os.Stderr, "[hook] ", log.LstdFlags)
    
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            logger.Printf("PreToolUse: Tool=%s, Session=%s", event.ToolName, event.SessionID)
            return cchooks.Approve(), nil
        },
        PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
            logger.Printf("PostToolUse: Tool=%s, Session=%s", event.ToolName, event.SessionID)
            return cchooks.Allow(), nil  // Empty response
        },
        Notification: func(ctx context.Context, event *cchooks.NotificationEvent) (*cchooks.NotificationResponse, error) {
            logger.Printf("Notification: %s", event.NotificationMessage)
            return cchooks.OK(), nil
        },
        Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
            logger.Printf("Stop event received")
            return cchooks.Continue(), nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Working with Different Tools

Now let's create a security hook that examines specific tools:

```go
package main

import (
    "context"
    "log"
    "strings"
    
    cchooks "github.com/brads3290/cchooks"
)

func main() {
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            switch event.ToolName {
            case "Bash":
                return handleBash(event)
            case "Write", "Edit":
                return handleFileOperation(event)
            default:
                // Approve other tools by default
                return cchooks.Approve(), nil
            }
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}

func handleBash(event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
    // Parse the Bash input
    bash, err := event.AsBash()
    if err != nil {
        return nil, err
    }
    
    // Check for dangerous commands
    dangerous := []string{"rm -rf", "dd if=", "mkfs", "> /dev/"}
    for _, pattern := range dangerous {
        if strings.Contains(bash.Command, pattern) {
            return cchooks.Block("Dangerous command detected: " + pattern), nil
        }
    }
    
    return cchooks.Approve(), nil
}

func handleFileOperation(event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
    // Different tools have different input types
    var filePath string
    
    switch event.ToolName {
    case "Write":
        write, err := event.AsWrite()
        if err != nil {
            return nil, err
        }
        filePath = write.FilePath
    case "Edit":
        edit, err := event.AsEdit()
        if err != nil {
            return nil, err
        }
        filePath = edit.FilePath
    }
    
    // Block writes to system files
    if strings.HasPrefix(filePath, "/etc/") || 
       strings.HasPrefix(filePath, "/sys/") ||
       strings.HasPrefix(filePath, "/proc/") {
        return cchooks.Block("Cannot modify system files"), nil
    }
    
    return cchooks.Approve(), nil
}
```

### Handling Errors

Let's add proper error handling and see how the Error handler works:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    cchooks "github.com/brads3290/cchooks"
)

func main() {
    logger := log.New(os.Stderr, "[hook] ", log.LstdFlags)
    
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            // Simulate an error condition
            if event.ToolName == "DEBUG_ERROR" {
                return nil, fmt.Errorf("simulated error for testing")
            }
            
            // Simulate a panic
            if event.ToolName == "DEBUG_PANIC" {
                panic("simulated panic for testing")
            }
            
            return cchooks.Approve(), nil
        },
        Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
            // Log the error details
            logger.Printf("Error occurred: %v", err)
            logger.Printf("Raw JSON: %s", rawJSON)
            
            // You can return nil to use default error handling
            // (exit code 2 for most events, 0 for Stop events)
            // Or return a custom response:
            if strings.Contains(err.Error(), "panic:") {
                return &cchooks.RawResponse{
                    ExitCode: 1,
                    Output:   "Hook crashed! Please check the logs.",
                }
            }
            
            // Use default handling for other errors
            return nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Testing Your Hooks

Testing is crucial for hooks. Here's how to write comprehensive tests:

```go
package main

import (
    "testing"
    
    cchooks "github.com/brads3290/cchooks"
    "github.com/stretchr/testify/assert"
)

func TestSecurityHook(t *testing.T) {
    runner := createSecurityRunner()  // Your runner creation logic
    tester := cchooks.NewTestRunner(runner)
    
    t.Run("blocks dangerous bash commands", func(t *testing.T) {
        dangerousCommands := []string{
            "rm -rf /",
            "dd if=/dev/zero of=/dev/sda",
            "mkfs.ext4 /dev/sda",
        }
        
        for _, cmd := range dangerousCommands {
            err := tester.AssertPreToolUseBlocks("Bash", &cchooks.BashInput{
                Command: cmd,
            })
            assert.NoError(t, err, "Should block command: %s", cmd)
        }
    })
    
    t.Run("allows safe bash commands", func(t *testing.T) {
        safeCommands := []string{
            "ls -la",
            "git status",
            "npm install",
        }
        
        for _, cmd := range safeCommands {
            err := tester.AssertPreToolUseApproves("Bash", &cchooks.BashInput{
                Command: cmd,
            })
            assert.NoError(t, err, "Should approve command: %s", cmd)
        }
    })
    
    t.Run("blocks system file writes", func(t *testing.T) {
        err := tester.AssertPreToolUseBlocks("Write", &cchooks.WriteInput{
            FilePath: "/etc/passwd",
            Content:  "malicious content",
        })
        assert.NoError(t, err)
    })
    
    t.Run("handles PostToolUse events", func(t *testing.T) {
        // Test with command failure
        response, err := tester.RunPostToolUse("Bash", 
            &cchooks.BashInput{Command: "test-cmd"},
            &cchooks.BashResponse{ExitCode: 1, Output: "command failed"})
        
        assert.NoError(t, err)
        assert.NotNil(t, response)
        // Add assertions based on your PostToolUse logic
    })
}

// Test the runner directly for more control
func TestRunnerWithRawInput(t *testing.T) {
    runner := createSecurityRunner()
    tester := cchooks.NewTestRunner(runner)
    
    // Test with raw JSON input
    output, exitCode, err := tester.RunWithInput(`{
        "event": "PreToolUse",
        "session_id": "test-123",
        "tool_name": "Bash",
        "tool_input": {"command": "rm -rf /"}
    }`)
    
    assert.NoError(t, err)
    assert.Equal(t, 0, exitCode)
    assert.Contains(t, output, `"decision": "block"`)
}
```

### Advanced Patterns

#### Pattern 1: Stateful Hooks with Context

Sometimes you need to track state across multiple tool uses:

```go
package main

import (
    "context"
    "sync"
    "time"
    
    cchooks "github.com/brads3290/cchooks"
)

type SessionTracker struct {
    mu       sync.Mutex
    sessions map[string]*SessionInfo
}

type SessionInfo struct {
    CommandCount int
    LastCommand  time.Time
    Blocked      bool
}

func main() {
    tracker := &SessionTracker{
        sessions: make(map[string]*SessionInfo),
    }
    
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            tracker.mu.Lock()
            defer tracker.mu.Unlock()
            
            // Get or create session info
            session, exists := tracker.sessions[event.SessionID]
            if !exists {
                session = &SessionInfo{}
                tracker.sessions[event.SessionID] = session
            }
            
            // Rate limiting: block if too many commands
            if time.Since(session.LastCommand) < time.Second && session.CommandCount > 10 {
                session.Blocked = true
                return cchooks.Block("Rate limit exceeded"), nil
            }
            
            session.CommandCount++
            session.LastCommand = time.Now()
            
            return cchooks.Approve(), nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

#### Pattern 2: Raw Handler for Custom Protocols

Use the Raw handler for complete control over processing:

```go
runner := &cchooks.Runner{
    Raw: func(ctx context.Context, rawJSON string) (*cchooks.RawResponse, error) {
        // Example: Add custom telemetry
        log.Printf("[TELEMETRY] Event received: %d bytes", len(rawJSON))
        
        // Example: Block based on raw patterns
        if strings.Contains(rawJSON, "FORBIDDEN_PATTERN") {
            return &cchooks.RawResponse{
                ExitCode: 1,
                Output:   "Forbidden pattern detected",
            }, nil
        }
        
        // Example: Transform the input before processing
        if strings.Contains(rawJSON, "legacy_format") {
            // Transform to new format...
            // Continue with normal processing
            return nil, nil
        }
        
        // Return nil to continue with normal event processing
        return nil, nil
    },
    PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
        return cchooks.Approve(), nil
    },
}
```

#### Pattern 3: Integration with External Services

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "time"
    
    cchooks "github.com/brads3290/cchooks"
)

type SecurityService struct {
    client  *http.Client
    baseURL string
}

func main() {
    service := &SecurityService{
        client:  &http.Client{Timeout: 5 * time.Second},
        baseURL: "https://security-api.example.com",
    }
    
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
            // Check with external service
            allowed, reason, err := service.CheckCommand(ctx, event)
            if err != nil {
                // Fail open or closed based on your security posture
                return cchooks.Approve(), nil  // Fail open
            }
            
            if !allowed {
                return cchooks.Block(reason), nil
            }
            
            return cchooks.Approve(), nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}

func (s *SecurityService) CheckCommand(ctx context.Context, event *cchooks.PreToolUseEvent) (bool, string, error) {
    payload, _ := json.Marshal(map[string]interface{}{
        "tool":       event.ToolName,
        "session_id": event.SessionID,
        "input":      event.ToolInput,
    })
    
    req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/check", bytes.NewReader(payload))
    if err != nil {
        return false, "", err
    }
    
    resp, err := s.client.Do(req)
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

### Error Handler

The Error handler is called for:
- JSON parsing errors
- Event validation errors
- Handler errors
- Response encoding errors
- Panics during processing (converted to errors with "panic:" prefix)

### Best Practices

1. **Always log to stderr**: stdout is reserved for hook responses
2. **Handle errors gracefully**: Return errors from handlers rather than panicking
3. **Test thoroughly**: Use the testing utilities to verify all code paths
4. **Keep hooks focused**: Each hook should have a single, clear purpose
5. **Fail safely**: Decide whether to fail open (approve) or closed (block) when external services are unavailable

## Configuration

Configure your hooks in Claude Code's settings:

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