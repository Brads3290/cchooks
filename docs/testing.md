# Testing Your Hooks

cchooks provides a comprehensive test SDK that makes it easy to unit test your hook handlers without dealing with process execution or JSON marshaling.

## The Test SDK

The test SDK simulates the inputs from real Claude Code hook calls, allowing you to test your handlers in isolation.

### How It Works

The `TestRunner` wraps your hook's `Runner` and provides methods to test individual handlers:

1. **Simulates Tool Inputs**: You provide Go structs (like `BashInput`), and it converts them to JSON just like Claude would send
2. **Creates Event Structures**: Generates the same event objects your handlers receive in production
3. **Handles JSON Marshaling**: Takes care of all the JSON conversion that normally happens through stdin/stdout
4. **Provides Test Context**: Supplies a context and test session IDs

### Real vs Test Comparison

**Real hook call from Claude:**
```json
{
  "hook_event_name": "PreToolUse",
  "session_id": "abc-123",
  "tool_name": "Bash",
  "tool_input": {
    "command": "ls -la",
    "timeout": 30000
  }
}
```

**Test SDK equivalent:**
```go
timeout := 30000
tr.TestPreToolUse("Bash", &cchooks.BashInput{
    Command: "ls -la",
    Timeout: &timeout,
})
```

Both result in your handler receiving the exact same `PreToolUseEvent` structure.

## Basic Testing

```go
package main

import (
    "context"
    "strings"
    "testing"
    "github.com/brads3290/cchooks"
)

func TestMyHook(t *testing.T) {
    // Create your hook's runner
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
            if event.ToolName == "Bash" {
                bash, _ := event.AsBash()
                if strings.Contains(bash.Command, "rm -rf") {
                    return cchooks.Block("dangerous command")
                }
            }
            return cchooks.Approve()
        },
    }
    
    // Wrap it in a TestRunner
    tr := cchooks.NewTestRunner(runner)
    
    // Test a dangerous command
    resp := tr.TestPreToolUse("Bash", &cchooks.BashInput{Command: "rm -rf /"})
    
    // Check for errors
    if errResp, ok := resp.(*cchooks.ErrorResponse); ok {
        t.Fatalf("handler returned error: %v", errResp.Error)
    }
    
    // Verify the response
    if preResp, ok := resp.(*cchooks.PreToolUseResponse); ok {
        if preResp.Decision != cchooks.PreToolUseBlock {
            t.Errorf("expected block, got %s", preResp.Decision)
        }
        if preResp.Reason != "dangerous command" {
            t.Errorf("expected reason 'dangerous command', got %s", preResp.Reason)
        }
    }
}
```

## Using Assertion Helpers

The TestRunner provides helper methods that make assertions cleaner:

```go
func TestHookAssertions(t *testing.T) {
    tr := cchooks.NewTestRunner(runner)
    
    // Test that dangerous commands are blocked
    err := tr.AssertPreToolUseBlocksWithReason(
        "Bash", 
        &cchooks.BashInput{Command: "rm -rf /"}, 
        "dangerous command",
    )
    if err != nil {
        t.Error(err)
    }
    
    // Test that safe commands are approved
    err = tr.AssertPreToolUseApproves(
        "Bash",
        &cchooks.BashInput{Command: "ls -la"},
    )
    if err != nil {
        t.Error(err)
    }
}
```

### Available Assertion Helpers

**PreToolUse:**
- `AssertPreToolUseApproves(toolName, toolInput)`
- `AssertPreToolUseBlocks(toolName, toolInput)`
- `AssertPreToolUseBlocksWithReason(toolName, toolInput, reason)`
- `AssertPreToolUseStopsClaude(toolName, toolInput)`

**PostToolUse:**
- `AssertPostToolUseAllows(toolName, toolInput, toolResponse)`
- `AssertPostToolUseBlocks(toolName, toolInput, toolResponse)`
- `AssertPostToolUseBlocksWithReason(toolName, toolInput, toolResponse, reason)`
- `AssertPostToolUseStopsClaude(toolName, toolInput, toolResponse)`

**Notification:**
- `AssertNotificationOK(message)`
- `AssertNotificationStopsClaude(message)`

**Stop:**
- `AssertStopContinues(stopHookActive, transcript)`
- `AssertStopBlocks(stopHookActive, transcript)`
- `AssertStopBlocksWithReason(stopHookActive, transcript, reason)`

## Testing Different Handlers

### Testing PostToolUse

```go
func TestPostToolUse(t *testing.T) {
    runner := &cchooks.Runner{
        PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
            bash, _ := event.ResponseAsBash()
            if bash.ExitCode != 0 {
                return cchooks.PostBlock("command failed")
            }
            return cchooks.Allow()
        },
    }
    
    tr := cchooks.NewTestRunner(runner)
    
    // Test successful command
    err := tr.AssertPostToolUseAllows(
        "Bash",
        &cchooks.BashInput{Command: "echo hello"},
        &cchooks.BashOutput{Output: "hello\n", ExitCode: 0},
    )
    if err != nil {
        t.Error(err)
    }
    
    // Test failed command
    err = tr.AssertPostToolUseBlocksWithReason(
        "Bash",
        &cchooks.BashInput{Command: "false"},
        &cchooks.BashOutput{Output: "", ExitCode: 1},
        "command failed",
    )
    if err != nil {
        t.Error(err)
    }
}
```

### Testing Stop Handler with Transcripts

```go
func TestStopHandler(t *testing.T) {
    runner := &cchooks.Runner{
        Stop: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
            // Check transcript for specific patterns
            for _, entry := range event.Transcript {
                if entry.IsUserMessage() {
                    userMsg, _ := entry.GetUserMessage()
                    if strings.Contains(userMsg.Content, "don't stop") {
                        return cchooks.BlockStop("user requested continuation")
                    }
                }
            }
            return cchooks.Continue()
        },
    }
    
    tr := cchooks.NewTestRunner(runner)
    
    // Create test transcript
    transcript := []cchooks.TranscriptEntry{
        {
            Type: "user",
            Message: json.RawMessage(`{"role":"user","content":"don't stop yet"}`),
        },
    }
    
    err := tr.AssertStopBlocksWithReason(true, transcript, "user requested continuation")
    if err != nil {
        t.Error(err)
    }
}
```

## Integration Testing

For full integration tests, you can also test the complete hook binary:

```go
func TestHookIntegration(t *testing.T) {
    // Build your hook
    cmd := exec.Command("go", "build", "-o", "test-hook", ".")
    if err := cmd.Run(); err != nil {
        t.Fatal(err)
    }
    defer os.Remove("test-hook")
    
    // Prepare test input
    input := map[string]interface{}{
        "hook_event_name": "PreToolUse",
        "session_id": "test-123",
        "tool_name": "Bash",
        "tool_input": map[string]interface{}{
            "command": "echo hello",
        },
    }
    
    inputJSON, _ := json.Marshal(input)
    
    // Run the hook
    cmd = exec.Command("./test-hook")
    cmd.Stdin = bytes.NewReader(inputJSON)
    
    output, err := cmd.Output()
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify output
    var response map[string]interface{}
    if err := json.Unmarshal(output, &response); err != nil {
        t.Fatal(err)
    }
    
    if response["decision"] != "approve" {
        t.Errorf("expected approve, got %v", response["decision"])
    }
}
```

## Best Practices

1. **Test Each Handler Type**: Write tests for all handler types you implement
2. **Test Edge Cases**: Include tests for error conditions, malformed inputs, etc.
3. **Use Assertion Helpers**: They make tests more readable and maintainable
4. **Mock External Dependencies**: If your hook calls external services, mock them in tests
5. **Test Error Conditions**: Ensure your hooks handle errors gracefully