# cchooks

A Go SDK for building Claude Code hooks - custom integrations that can intercept and control Claude's tool usage.

## Features

- 🛡️ **Security Controls** - Block dangerous commands before execution
- 📝 **Audit Logging** - Track all tool usage with detailed logs  
- 🔧 **Tool Interception** - Approve, modify, or block any tool Claude uses
- ✅ **Format Enforcement** - Ensure code follows your team's standards
- 🧪 **Comprehensive Testing** - Built-in test SDK for unit testing hooks
- 🔄 **Stop Control** - Manage when Claude can stop execution

## Quick Start

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

## Installation

```bash
go get github.com/brads3290/cchooks
```

## Documentation

- **[Getting Started](docs/getting-started.md)** - Create your first hook
- **[Handler Types](docs/handlers.md)** - Learn about different event handlers
- **[Tool-Specific Inputs](docs/tools.md)** - Work with Bash, Edit, Read, Write tools
- **[Testing Your Hooks](docs/testing.md)** - Unit test your hooks with the test SDK
- **[Example Hooks](docs/examples.md)** - Real-world examples and patterns
- **[API Reference](docs/api-reference.md)** - Complete type and function reference
- **[Advanced Topics](docs/advanced.md)** - Context, debugging, and security

## Examples

The repository includes several example hooks in the `examples/` directory:

- `simple-hook` - Basic example showing core functionality
- `security-hook` - Advanced security controls and command filtering
- `format-hook` - Code formatting enforcement

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.