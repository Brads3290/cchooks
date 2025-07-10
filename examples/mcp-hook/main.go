package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
			// Handle MCP tools differently from built-in tools
			if event.IsMCPTool() {
				mcpTool, err := event.AsMCPTool()
				if err != nil {
					return nil, err
				}

				// Log MCP tool information
				log.Printf("MCP Tool detected - Server: %s, Tool: %s", mcpTool.MCPName, mcpTool.ToolName)

				// Parse raw input for inspection
				var params map[string]interface{}
				if err := json.Unmarshal(mcpTool.RawInput, &params); err != nil {
					log.Printf("Failed to parse MCP input: %v", err)
				} else {
					log.Printf("MCP Tool parameters: %+v", params)
				}

				// Apply server-specific logic
				switch mcpTool.MCPName {
				case "weather":
					// Validate weather API calls
					if mcpTool.ToolName == "get_forecast" {
						if location, ok := params["location"].(string); ok {
							if location == "" {
								return cchooks.Block("Location is required for weather forecast"), nil
							}
						}
					}

				case "database":
					// More restrictive for database operations
					if mcpTool.ToolName == "delete_user" {
						return cchooks.Block("Database deletions are not allowed via MCP"), nil
					}

				case "api":
					// Check API endpoints
					if endpoint, ok := params["endpoint"].(string); ok {
						if endpoint == "/admin" {
							return cchooks.Block("Admin endpoints are restricted"), nil
						}
					}
				}

				// Default: approve MCP tools
				return cchooks.Approve(), nil
			}

			// Handle built-in tools as usual
			switch event.ToolName {
			case "Bash":
				bash, err := event.AsBash()
				if err != nil {
					return nil, err
				}
				log.Printf("Bash command: %s", bash.Command)

			case "Edit":
				edit, err := event.AsEdit()
				if err != nil {
					return nil, err
				}
				log.Printf("Editing file: %s", edit.FilePath)
			}

			return cchooks.Approve(), nil
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
			if event.IsMCPTool() {
				// Handle MCP tool responses
				mcpTool, err := event.InputAsMCPTool()
				if err != nil {
					return nil, err
				}

				mcpOutput, err := event.ResponseAsMCPTool()
				if err != nil {
					return nil, err
				}

				log.Printf("MCP Tool completed - Server: %s, Tool: %s", mcpTool.MCPName, mcpTool.ToolName)

				// Parse and inspect the response
				var response map[string]interface{}
				if err := json.Unmarshal(mcpOutput.RawOutput, &response); err != nil {
					// Some MCP tools might return non-JSON responses
					log.Printf("MCP response (raw): %s", string(mcpOutput.RawOutput))
				} else {
					log.Printf("MCP response: %+v", response)

					// Example: Check for errors in response
					if errorMsg, ok := response["error"].(string); ok && errorMsg != "" {
						return cchooks.PostBlock(fmt.Sprintf("MCP tool error: %s", errorMsg)), nil
					}
				}
			} else {
				// Handle built-in tool responses
				log.Printf("Built-in tool %s completed", event.ToolName)
			}

			return cchooks.Allow(), nil
		},

		Notification: func(ctx context.Context, event *cchooks.NotificationEvent) (*cchooks.NotificationResponse, error) {
			// Could check for MCP-related notifications
			log.Printf("Notification: %s", event.Message)
			return cchooks.OK(), nil
		},
	}

	runner.Run()
}
