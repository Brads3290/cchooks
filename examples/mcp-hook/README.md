# MCP Hook Example

This example demonstrates how to handle MCP (Model Context Protocol) tools in Claude Code hooks.

## Features

- Detects MCP tools using the `IsMCPTool()` method
- Extracts MCP server name and tool name from the full tool identifier
- Handles MCP tools differently from built-in tools
- Shows how to parse and validate MCP tool parameters
- Demonstrates server-specific logic for different MCP servers

## MCP Tool Format

MCP tools follow the naming convention: `mcp__servername__toolname`

For example:
- `mcp__weather__get_forecast` - Weather server, get_forecast tool
- `mcp__database__query_users` - Database server, query_users tool
- `mcp__api__call_endpoint` - API server, call_endpoint tool

## Usage

```go
// Check if it's an MCP tool
if event.IsMCPTool() {
    mcpTool, err := event.AsMCPTool()
    if err != nil {
        return nil, err
    }
    
    // Access MCP tool details
    fmt.Printf("MCP Server: %s\n", mcpTool.MCPName)
    fmt.Printf("Tool Name: %s\n", mcpTool.ToolName)
    
    // Parse custom parameters
    var params map[string]interface{}
    json.Unmarshal(mcpTool.RawInput, &params)
}
```

## Example Logic

The hook in this example:

1. **Weather Server**: Validates that location is provided for forecasts
2. **Database Server**: Blocks dangerous operations like user deletions
3. **API Server**: Restricts access to admin endpoints

## Testing

To test this hook with different MCP tools:

```bash
# Weather forecast (should pass)
echo '{"hook_event_name": "PreToolUse", "session_id": "test", "tool_name": "mcp__weather__get_forecast", "tool_input": {"location": "San Francisco"}}' | go run .

# Database deletion (should block)
echo '{"hook_event_name": "PreToolUse", "session_id": "test", "tool_name": "mcp__database__delete_user", "tool_input": {"user_id": 123}}' | go run .

# API admin endpoint (should block)
echo '{"hook_event_name": "PreToolUse", "session_id": "test", "tool_name": "mcp__api__call_endpoint", "tool_input": {"endpoint": "/admin", "method": "GET"}}' | go run .
```