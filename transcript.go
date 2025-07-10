package cchooks

import (
	"encoding/json"
	"time"
)

// TranscriptEntry represents a single entry in the Claude Code transcript
type TranscriptEntry struct {
	ParentUUID        *string         `json:"parentUuid"`
	UUID              string          `json:"uuid"`
	IsSidechain       bool            `json:"isSidechain"`
	UserType          string          `json:"userType"`
	CWD               string          `json:"cwd"`
	SessionID         string          `json:"sessionId"`
	Version           string          `json:"version"`
	Type              string          `json:"type"` // "user" or "assistant"
	Message           json.RawMessage `json:"message"`
	Timestamp         time.Time       `json:"timestamp"`
	RequestID         string          `json:"requestId,omitempty"`
	IsMeta            bool            `json:"isMeta,omitempty"`
	ToolUseResult     interface{}     `json:"toolUseResult,omitempty"`
	IsAPIErrorMessage bool            `json:"isApiErrorMessage,omitempty"`
}

// UserMessage represents a user message in the transcript
type UserMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"` // Can be string or array of content blocks
}

// AssistantMessage represents an assistant message in the transcript
type AssistantMessage struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Model        string          `json:"model"`
	Content      json.RawMessage `json:"content"` // Array of content blocks
	StopReason   *string         `json:"stop_reason"`
	StopSequence *string         `json:"stop_sequence"`
	Usage        Usage           `json:"usage"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens              int    `json:"input_tokens"`
	OutputTokens             int    `json:"output_tokens"`
	CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
	ServiceTier              string `json:"service_tier,omitempty"`
}

// ContentBlock represents a content block in messages
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   interface{}     `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

// GetUserMessage parses the message field as a UserMessage for user type entries
func (t *TranscriptEntry) GetUserMessage() (*UserMessage, error) {
	if t.Type != "user" {
		return nil, nil
	}
	var msg UserMessage
	if err := json.Unmarshal(t.Message, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetAssistantMessage parses the message field as an AssistantMessage for assistant type entries
func (t *TranscriptEntry) GetAssistantMessage() (*AssistantMessage, error) {
	if t.Type != "assistant" {
		return nil, nil
	}
	var msg AssistantMessage
	if err := json.Unmarshal(t.Message, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// IsUserMessage returns true if this is a user message
func (t *TranscriptEntry) IsUserMessage() bool {
	return t.Type == "user"
}

// IsAssistantMessage returns true if this is an assistant message
func (t *TranscriptEntry) IsAssistantMessage() bool {
	return t.Type == "assistant"
}
