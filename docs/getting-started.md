# Getting Started with cchooks

This guide will help you create your first Claude Code hook.

## Prerequisites

- Go 1.21 or later
- Basic understanding of Go programming

## Creating Your First Hook

1. **Create a new directory** for your hook:
```bash
mkdir my-first-hook
cd my-first-hook
```

2. **Initialize a Go module**:
```bash
go mod init my-first-hook
```

3. **Install cchooks**:
```bash
go get github.com/brads3290/cchooks
```

4. **Create main.go**:
```go
package main

import (
    "context"
    "strings"
    "github.com/brads3290/cchooks"
)

func main() {
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
            // Block dangerous commands
            if event.ToolName == "Bash" {
                bash, _ := event.AsBash()
                if strings.Contains(bash.Command, "rm -rf") {
                    return cchooks.Block("dangerous command detected")
                }
            }
            return cchooks.Approve()
        },
    }
    
    runner.Run()
}
```

5. **Build your hook**:
```bash
go build -o my-first-hook
```

6. **Configure Claude Code** to use your hook by updating your configuration file.

## Next Steps

- Learn about [different handler types](./handlers.md)
- Explore [tool-specific inputs](./tools.md)
- Set up [testing for your hooks](./testing.md)
- See [example hooks](./examples.md) for more complex scenarios