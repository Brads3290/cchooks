# Example Hooks

This page showcases various example hooks to demonstrate different use cases and patterns.

## Simple Security Hook

Blocks potentially dangerous commands:

```go
package main

import (
    "context"
    "strings"
    "github.com/brads3290/cchooks"
)

func main() {
    dangerousPatterns := []string{
        "rm -rf /",
        "dd if=/dev/zero",
        ":(){ :|:& };:",  // Fork bomb
        "> /dev/sda",
    }
    
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
            if event.ToolName == "Bash" {
                bash, _ := event.AsBash()
                for _, pattern := range dangerousPatterns {
                    if strings.Contains(bash.Command, pattern) {
                        return cchooks.Block("Dangerous command detected: " + pattern)
                    }
                }
            }
            return cchooks.Approve()
        },
    }
    
    runner.Run()
}
```

## Format Enforcement Hook

Ensures code follows formatting standards:

```go
package main

import (
    "context"
    "os/exec"
    "strings"
    "github.com/brads3290/cchooks"
)

func main() {
    runner := &cchooks.Runner{
        PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
            // Check if a Go file was written or edited
            if event.ToolName == "Write" || event.ToolName == "Edit" {
                var filePath string
                
                if event.ToolName == "Write" {
                    write, _ := event.InputAsWrite()
                    filePath = write.FilePath
                } else {
                    edit, _ := event.InputAsEdit()
                    filePath = edit.FilePath
                }
                
                if strings.HasSuffix(filePath, ".go") {
                    // Run gofmt check
                    cmd := exec.Command("gofmt", "-l", filePath)
                    output, _ := cmd.Output()
                    
                    if len(output) > 0 {
                        return cchooks.PostBlock("Go file is not formatted. Please run gofmt.")
                    }
                }
            }
            
            return cchooks.Allow()
        },
    }
    
    runner.Run()
}
```

## Logging Hook

Logs all tool usage to a file:

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "time"
    "github.com/brads3290/cchooks"
)

func main() {
    // Open log file
    logFile, err := os.OpenFile("claude-tools.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer logFile.Close()
    
    logger := log.New(logFile, "", log.LstdFlags)
    
    runner := &cchooks.Runner{
        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
            logEntry := map[string]interface{}{
                "timestamp": time.Now().UTC(),
                "event":     "pre_tool_use",
                "tool":      event.ToolName,
                "session":   event.SessionID,
            }
            
            // Log tool-specific details
            switch event.ToolName {
            case "Bash":
                if bash, err := event.AsBash(); err == nil {
                    logEntry["command"] = bash.Command
                }
            case "Write":
                if write, err := event.AsWrite(); err == nil {
                    logEntry["file"] = write.FilePath
                }
            }
            
            jsonEntry, _ := json.Marshal(logEntry)
            logger.Println(string(jsonEntry))
            
            return cchooks.Approve()
        },
    }
    
    runner.Run()
}
```

## Stop Confirmation Hook

Requires confirmation before allowing Claude to stop:

```go
package main

import (
    "context"
    "strings"
    "github.com/brads3290/cchooks"
)

func main() {
    runner := &cchooks.Runner{
        StopOnce: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
            // On first stop attempt, check if important tasks are mentioned
            importantTasks := []string{"test", "commit", "deploy", "review"}
            
            for _, entry := range event.Transcript {
                if entry.IsUserMessage() {
                    msg, _ := entry.GetUserMessage()
                    content := strings.ToLower(msg.Content)
                    
                    for _, task := range importantTasks {
                        if strings.Contains(content, task) {
                            return cchooks.BlockStop("Important task '" + task + "' was mentioned. Are you sure all tasks are complete?")
                        }
                    }
                }
            }
            
            return cchooks.Continue()
        },
        
        Stop: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
            // After first attempt, always allow
            return cchooks.Continue()
        },
    }
    
    runner.Run()
}
```

## Combined Hook with Multiple Handlers

A comprehensive hook that combines multiple safety features:

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
            switch event.ToolName {
            case "Bash":
                bash, _ := event.AsBash()
                // Block dangerous commands
                if isDangerous(bash.Command) {
                    return cchooks.Block("Command blocked for safety")
                }
                // Warn about sudo
                if strings.Contains(bash.Command, "sudo") {
                    return cchooks.Block("Sudo commands require manual execution")
                }
                
            case "Write", "Edit":
                // Protect system files
                var path string
                if event.ToolName == "Write" {
                    write, _ := event.AsWrite()
                    path = write.FilePath
                } else {
                    edit, _ := event.AsEdit()
                    path = edit.FilePath
                }
                
                if isSystemFile(path) {
                    return cchooks.Block("Cannot modify system files")
                }
            }
            
            return cchooks.Approve()
        },
        
        PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
            if event.ToolName == "Bash" {
                bash, _ := event.ResponseAsBash()
                // Hide sensitive output
                if containsSensitiveInfo(bash.Output) {
                    return cchooks.PostBlock("Output contains sensitive information")
                }
            }
            return cchooks.Allow()
        },
        
        Notification: func(ctx context.Context, event *cchooks.NotificationEvent) cchooks.NotificationResponseInterface {
            // Could log notifications or send alerts
            return cchooks.OK()
        },
    }
    
    runner.Run()
}

func isDangerous(command string) bool {
    dangerous := []string{
        "rm -rf /",
        "dd if=/dev/zero",
        "mkfs",
        "fdisk",
    }
    
    for _, d := range dangerous {
        if strings.Contains(command, d) {
            return true
        }
    }
    return false
}

func isSystemFile(path string) bool {
    systemPaths := []string{
        "/etc/",
        "/sys/",
        "/proc/",
        "/boot/",
    }
    
    for _, sp := range systemPaths {
        if strings.HasPrefix(path, sp) {
            return true
        }
    }
    return false
}

func containsSensitiveInfo(output string) bool {
    // Check for common patterns
    patterns := []string{
        "password:",
        "secret:",
        "api_key:",
        "private_key:",
    }
    
    lower := strings.ToLower(output)
    for _, p := range patterns {
        if strings.Contains(lower, p) {
            return true
        }
    }
    return false
}
```

## More Examples

The cchooks repository includes additional examples in the `examples/` directory:

- **simple-hook**: Basic example showing core functionality
- **security-hook**: Advanced security features
- **format-hook**: Code formatting enforcement

Each example includes tests demonstrating proper usage of the test SDK.