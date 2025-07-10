# Tool-Specific Inputs and Responses

Claude Code uses various tools, and cchooks provides typed access to their inputs and outputs.

## Bash Tool

The Bash tool executes shell commands.

### Input
```go
bash, err := event.AsBash()
if err != nil {
    return cchooks.Error(err)
}

// Access fields
command := bash.Command
timeout := bash.Timeout  // *int in milliseconds
```

### Response (PostToolUse only)
```go
bashResp, err := event.ResponseAsBash()
if err != nil {
    return cchooks.Error(err)
}

// Access fields
output := bashResp.Output
exitCode := bashResp.ExitCode
```

## Edit Tool

The Edit tool modifies files.

### Input
```go
edit, err := event.AsEdit()
if err != nil {
    return cchooks.Error(err)
}

// Access fields
filePath := edit.FilePath
oldString := edit.OldString
newString := edit.NewString
```

### Response (PostToolUse only)
```go
editResp, err := event.ResponseAsEdit()
if err != nil {
    return cchooks.Error(err)
}

// Check if edit succeeded
if editResp.Success {
    // Edit was successful
}
```

## Read Tool

The Read tool reads file contents.

### Input
```go
read, err := event.AsRead()
if err != nil {
    return cchooks.Error(err)
}

// Access fields
filePath := read.FilePath
offset := read.Offset  // *int
limit := read.Limit    // *int
```

### Response (PostToolUse only)
```go
readResp, err := event.ResponseAsRead()
if err != nil {
    return cchooks.Error(err)
}

// Access file content
content := readResp.Content
```

## Write Tool

The Write tool creates or overwrites files.

### Input
```go
write, err := event.AsWrite()
if err != nil {
    return cchooks.Error(err)
}

// Access fields
filePath := write.FilePath
content := write.Content
```

### Response (PostToolUse only)
```go
writeResp, err := event.ResponseAsWrite()
if err != nil {
    return cchooks.Error(err)
}

// Check if write succeeded
if writeResp.Success {
    // Write was successful
}
```

## Generic Tool Access

For tools without specific helper methods, access the raw JSON:

```go
// PreToolUse
var toolInput map[string]interface{}
err := json.Unmarshal(event.ToolInput, &toolInput)

// PostToolUse
var toolResponse map[string]interface{}
err := json.Unmarshal(event.ToolResponse, &toolResponse)
```

## Example: Tool-Specific Logic

```go
PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
    switch event.ToolName {
    case "Bash":
        bash, _ := event.AsBash()
        if isSensitiveCommand(bash.Command) {
            return cchooks.Block("sensitive command blocked")
        }
    
    case "Write":
        write, _ := event.AsWrite()
        if isSensitiveFile(write.FilePath) {
            return cchooks.Block("cannot write to sensitive file")
        }
    
    case "Edit":
        edit, _ := event.AsEdit()
        if strings.Contains(edit.NewString, "SECRET") {
            return cchooks.Block("cannot add secrets to files")
        }
    }
    
    return cchooks.Approve()
}
```