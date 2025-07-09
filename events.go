package cchooks

import "encoding/json"

// Event types - data containers for each hook event
type PreToolUseEvent struct {
	SessionID string          `json:"session_id"`
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

type PostToolUseEvent struct {
	SessionID    string          `json:"session_id"`
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

type NotificationEvent struct {
	SessionID string `json:"session_id"`
	Message   string `json:"notification_message"`
}

type StopEvent struct {
	SessionID      string        `json:"session_id"`
	StopHookActive bool          `json:"stop_hook_active"`
	Transcript     []interface{} `json:"transcript"`
}

