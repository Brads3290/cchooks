package cchooks

// Response types with event-specific decision options

// PreToolUseResponse is the response for PreToolUse events.
type PreToolUseResponse struct {
	Decision   string `json:"decision,omitempty"`
	Continue   *bool  `json:"continue,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// PostToolUseResponse is the response for PostToolUse events.
type PostToolUseResponse struct {
	Decision   string `json:"decision,omitempty"`
	Continue   *bool  `json:"continue,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// NotificationResponse is the response for Notification events.
type NotificationResponse struct {
	Continue   *bool  `json:"continue,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
}

// StopResponse is the response for Stop events.
type StopResponse struct {
	Decision   string `json:"decision,omitempty"`
	Continue   *bool  `json:"continue,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// Constants for decisions
const (
	PreToolUseApprove = "approve"
	PreToolUseBlock   = "block"
	PostToolUseBlock  = "block"
	StopBlock         = "block"
)

// Helper functions for common responses

// Approve creates a PreToolUseResponse that approves the tool use
func Approve() *PreToolUseResponse {
	return &PreToolUseResponse{Decision: PreToolUseApprove}
}

// Block creates a PreToolUseResponse that blocks the tool use with a reason
func Block(reason string) *PreToolUseResponse {
	return &PreToolUseResponse{Decision: PreToolUseBlock, Reason: reason}
}

// PostBlock creates a PostToolUseResponse that blocks the tool use with a reason
func PostBlock(reason string) *PostToolUseResponse {
	return &PostToolUseResponse{Decision: PostToolUseBlock, Reason: reason}
}

// StopClaude creates a PreToolUseResponse that stops Claude with a reason
func StopClaude(reason string) *PreToolUseResponse {
	cont := false
	return &PreToolUseResponse{Continue: &cont, StopReason: reason}
}

// StopClaudePost creates a PostToolUseResponse that stops Claude with a reason
func StopClaudePost(reason string) *PostToolUseResponse {
	cont := false
	return &PostToolUseResponse{Continue: &cont, StopReason: reason}
}

// Allow creates an empty PostToolUseResponse that allows the action
func Allow() *PostToolUseResponse {
	return &PostToolUseResponse{}
}

// OK creates an empty NotificationResponse that continues
func OK() *NotificationResponse {
	return &NotificationResponse{}
}

// StopFromNotification creates a NotificationResponse that stops Claude
func StopFromNotification(reason string) *NotificationResponse {
	cont := false
	return &NotificationResponse{Continue: &cont, StopReason: reason}
}

// Continue creates an empty StopResponse that allows the stop
func Continue() *StopResponse {
	return &StopResponse{}
}

// BlockStop creates a StopResponse that blocks the stop with a reason
func BlockStop(reason string) *StopResponse {
	return &StopResponse{Decision: StopBlock, Reason: reason}
}

// StopFromStop creates a StopResponse that stops Claude
func StopFromStop(reason string) *StopResponse {
	cont := false
	return &StopResponse{Continue: &cont, StopReason: reason}
}

