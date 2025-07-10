package cchooks

import (
	"encoding/json"
	"testing"
)

func TestMCPToolDetection(t *testing.T) {
	tests := []struct {
		name            string
		toolName        string
		wantIsMCP       bool
		wantMCPToolName string
	}{
		{
			name:            "standard tool",
			toolName:        "Bash",
			wantIsMCP:       false,
			wantMCPToolName: "",
		},
		{
			name:            "simple MCP tool",
			toolName:        "mcp__weather__get_forecast",
			wantIsMCP:       true,
			wantMCPToolName: "mcp__weather__get_forecast",
		},
		{
			name:            "MCP tool with underscores in tool name",
			toolName:        "mcp__myserver__get_user_data",
			wantIsMCP:       true,
			wantMCPToolName: "mcp__myserver__get_user_data",
		},
		{
			name:            "tool starting with mcp but not MCP format",
			toolName:        "mcp_tool",
			wantIsMCP:       false,
			wantMCPToolName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test PreToolUseEvent
			preEvent := &PreToolUseEvent{
				ToolName: tt.toolName,
			}

			if got := preEvent.IsMCPTool(); got != tt.wantIsMCP {
				t.Errorf("PreToolUseEvent.IsMCPTool() = %v, want %v", got, tt.wantIsMCP)
			}

			if got := preEvent.MCPToolName(); got != tt.wantMCPToolName {
				t.Errorf("PreToolUseEvent.MCPToolName() = %q, want %q", got, tt.wantMCPToolName)
			}

			// Test PostToolUseEvent
			postEvent := &PostToolUseEvent{
				ToolName: tt.toolName,
			}

			if got := postEvent.IsMCPTool(); got != tt.wantIsMCP {
				t.Errorf("PostToolUseEvent.IsMCPTool() = %v, want %v", got, tt.wantIsMCP)
			}

			if got := postEvent.MCPToolName(); got != tt.wantMCPToolName {
				t.Errorf("PostToolUseEvent.MCPToolName() = %q, want %q", got, tt.wantMCPToolName)
			}
		})
	}
}

func TestMCPToolParsing(t *testing.T) {
	t.Run("valid MCP tool parsing", func(t *testing.T) {
		input := map[string]interface{}{
			"location": "San Francisco",
			"units":    "celsius",
		}
		inputJSON, _ := json.Marshal(input)

		event := &PreToolUseEvent{
			ToolName:  "mcp__weather__get_forecast",
			ToolInput: inputJSON,
		}

		mcpTool, err := event.AsMCPTool()
		if err != nil {
			t.Fatalf("AsMCPTool() error = %v", err)
		}

		if mcpTool.MCPName != "weather" {
			t.Errorf("MCPName = %q, want %q", mcpTool.MCPName, "weather")
		}

		if mcpTool.ToolName != "get_forecast" {
			t.Errorf("ToolName = %q, want %q", mcpTool.ToolName, "get_forecast")
		}

		// Verify we can parse the raw input
		var parsedInput map[string]interface{}
		if err := json.Unmarshal(mcpTool.RawInput, &parsedInput); err != nil {
			t.Errorf("Failed to parse RawInput: %v", err)
		}

		if parsedInput["location"] != "San Francisco" {
			t.Errorf("Expected location = San Francisco, got %v", parsedInput["location"])
		}
	})

	t.Run("MCP tool with underscores in tool name", func(t *testing.T) {
		event := &PreToolUseEvent{
			ToolName:  "mcp__database__get_user_by_id",
			ToolInput: json.RawMessage(`{"id": 123}`),
		}

		mcpTool, err := event.AsMCPTool()
		if err != nil {
			t.Fatalf("AsMCPTool() error = %v", err)
		}

		if mcpTool.MCPName != "database" {
			t.Errorf("MCPName = %q, want %q", mcpTool.MCPName, "database")
		}

		if mcpTool.ToolName != "get_user_by_id" {
			t.Errorf("ToolName = %q, want %q", mcpTool.ToolName, "get_user_by_id")
		}
	})

	t.Run("non-MCP tool returns error", func(t *testing.T) {
		event := &PreToolUseEvent{
			ToolName:  "Bash",
			ToolInput: json.RawMessage(`{"command": "ls"}`),
		}

		_, err := event.AsMCPTool()
		if err == nil {
			t.Error("Expected error for non-MCP tool, got nil")
		}
	})

	t.Run("invalid MCP tool format", func(t *testing.T) {
		event := &PreToolUseEvent{
			ToolName:  "mcp__invalid",
			ToolInput: json.RawMessage(`{}`),
		}

		_, err := event.AsMCPTool()
		if err == nil {
			t.Error("Expected error for invalid MCP tool format, got nil")
		}
	})
}

func TestMCPToolResponse(t *testing.T) {
	t.Run("valid MCP tool response", func(t *testing.T) {
		response := map[string]interface{}{
			"temperature": 22,
			"condition":   "sunny",
		}
		responseJSON, _ := json.Marshal(response)

		event := &PostToolUseEvent{
			ToolName:     "mcp__weather__get_forecast",
			ToolInput:    json.RawMessage(`{"location": "SF"}`),
			ToolResponse: responseJSON,
		}

		// Test input parsing
		mcpTool, err := event.InputAsMCPTool()
		if err != nil {
			t.Fatalf("InputAsMCPTool() error = %v", err)
		}

		if mcpTool.MCPName != "weather" {
			t.Errorf("MCPName = %q, want %q", mcpTool.MCPName, "weather")
		}

		// Test response parsing
		mcpOutput, err := event.ResponseAsMCPTool()
		if err != nil {
			t.Fatalf("ResponseAsMCPTool() error = %v", err)
		}

		if mcpOutput.MCPName != "weather" {
			t.Errorf("MCPName = %q, want %q", mcpOutput.MCPName, "weather")
		}

		if mcpOutput.ToolName != "get_forecast" {
			t.Errorf("ToolName = %q, want %q", mcpOutput.ToolName, "get_forecast")
		}

		// Verify we can parse the raw output
		var parsedOutput map[string]interface{}
		if err := json.Unmarshal(mcpOutput.RawOutput, &parsedOutput); err != nil {
			t.Errorf("Failed to parse RawOutput: %v", err)
		}

		if parsedOutput["temperature"] != float64(22) {
			t.Errorf("Expected temperature = 22, got %v", parsedOutput["temperature"])
		}
	})
}
