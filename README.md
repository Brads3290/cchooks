# Claude Code Hooks Go SDK

A Go SDK for creating strongly typed Claude Code hooks. This SDK simplifies the creation of individual hook binaries that handle Claude Code events with type safety and testing utilities.

## Installation

```bash
go get github.com/brads3290/claude-code-hooks-go
```

## Quick Start

Create a simple hook that blocks dangerous commands:

```go
package main

import (
    "context"
    "log"
    "strings"

    cchooks "github.com/brads3290/claude-code-hooks-go"
)

func main() {
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
            }
            
            return cchooks.Approve(), nil
        },
    }
    
    if err := runner.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## Features

- **Type Safety**: Strongly typed events and responses for all Claude Code tools
- **Easy Testing**: Built-in testing framework with assertion helpers
- **Complete Tool Coverage**: Support for all 15+ Claude Code tools
- **Simple Architecture**: Each hook is a standalone binary
- **Flexible Responses**: Support for approve, block, and stop decisions

## Event Types

The SDK supports four event types:

- `PreToolUse`: Called before a tool is executed
- `PostToolUse`: Called after a tool is executed
- `Notification`: Called for Claude notifications
- `Stop`: Called when Claude is stopping

## Response Types

Each event type has specific response options:

### PreToolUse Responses
- `Approve()`: Allow the tool to execute
- `Block(reason)`: Block the tool execution
- `StopClaude(reason)`: Stop Claude from continuing

### PostToolUse Responses
- `Allow()`: Continue normally (empty response)
- `PostBlock(reason)`: Block based on tool output
- `StopClaudePost(reason)`: Stop Claude after seeing output

### Notification Responses
- `OK()`: Continue normally
- `StopFromNotification(reason)`: Stop Claude

### Stop Responses
- `Continue()`: Allow Claude to stop
- `BlockStop(reason)`: Prevent Claude from stopping

## Tool Input Parsing

The SDK provides typed parsing for all Claude Code tools:

```go
// Parse Bash input
bash, err := event.AsBash()
if err != nil {
    return nil, err
}
fmt.Println(bash.Command)

// Parse Edit input
edit, err := event.AsEdit()
if err != nil {
    return nil, err
}
fmt.Println(edit.FilePath, edit.OldString, edit.NewString)

// Parse PostToolUse input and response
input, _ := event.InputAsBash()
output, _ := event.ResponseAsBash()
if output.ExitCode != 0 {
    return cchooks.PostBlock("Command failed"), nil
}
```

## Testing

The SDK includes a comprehensive testing framework:

```go
func TestMyHook(t *testing.T) {
    runner := createMyRunner()
    tester := cchooks.NewTestRunner(runner)
    
    // Test that dangerous commands are blocked
    err := tester.AssertPreToolUseBlocks("Bash", &cchooks.BashInput{
        Command: "rm -rf /",
    })
    assert.NoError(t, err)
    
    // Test that safe commands are approved
    err = tester.AssertPreToolUseApproves("Bash", &cchooks.BashInput{
        Command: "ls -la",
    })
    assert.NoError(t, err)
}
```

## Claude Code Configuration

Configure your hooks in Claude Code's `settings.json`:

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
    ]
  }
}
```

## Examples

See the `examples/` directory for complete examples:

- `security-hook`: Blocks dangerous commands and system file edits
- `format-hook`: Auto-formats code after edits
- `simple-hook`: Basic hook with logging

## Building Hooks

Build your hook as a standard Go binary:

```bash
go build -o my-hook main.go
```

## Error Handling

- Exit code 0: Success
- Exit code 2: Error sent to Claude
- Other exit codes: Error shown to user

## Contributing

Contributions are welcome! Please submit pull requests or issues on GitHub.

## License

MIT License - see LICENSE file for details.