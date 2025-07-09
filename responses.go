package cchooks

// Response types with event-specific decision options
type PreToolUseResponse struct {
	Decision   string `json:"decision,omitempty"`
	Continue   *bool  `json:"continue,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type PostToolUseResponse struct {
	Decision   string `json:"decision,omitempty"`
	Continue   *bool  `json:"continue,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type NotificationResponse struct {
	Continue   *bool  `json:"continue,omitempty"`
	StopReason string `json:"stopReason,omitempty"`
}

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

