package cchooks

import (
	"context"
	"encoding/json"
	"fmt"
)

// TestRunner provides testing utilities for hook validation
type TestRunner struct {
	runner *Runner
}

// NewTestRunner creates a new test runner
func NewTestRunner(runner *Runner) *TestRunner {
	return &TestRunner{runner: runner}
}

// TestPreToolUse tests a PreToolUse handler
func (t *TestRunner) TestPreToolUse(toolName string, toolInput interface{}) PreToolUseResponseInterface {
	inputJSON, err := json.Marshal(toolInput)
	if err != nil {
		return Error(err)
	}

	event := &PreToolUseEvent{
		SessionID: "test-session",
		ToolName:  toolName,
		ToolInput: inputJSON,
	}

	if t.runner.PreToolUse == nil {
		return Error(fmt.Errorf("PreToolUse handler not set"))
	}

	return t.runner.PreToolUse(context.Background(), event)
}

// TestPostToolUse tests a PostToolUse handler
func (t *TestRunner) TestPostToolUse(toolName string, toolInput, toolResponse interface{}) PostToolUseResponseInterface {
	inputJSON, err := json.Marshal(toolInput)
	if err != nil {
		return Error(err)
	}

	responseJSON, err := json.Marshal(toolResponse)
	if err != nil {
		return Error(err)
	}

	event := &PostToolUseEvent{
		SessionID:    "test-session",
		ToolName:     toolName,
		ToolInput:    inputJSON,
		ToolResponse: responseJSON,
	}

	if t.runner.PostToolUse == nil {
		return Error(fmt.Errorf("PostToolUse handler not set"))
	}

	return t.runner.PostToolUse(context.Background(), event)
}

// TestNotification tests a Notification handler
func (t *TestRunner) TestNotification(message string) NotificationResponseInterface {
	event := &NotificationEvent{
		SessionID: "test-session",
		Message:   message,
	}

	if t.runner.Notification == nil {
		return Error(fmt.Errorf("Notification handler not set"))
	}

	return t.runner.Notification(context.Background(), event)
}

// TestStop tests a Stop handler
func (t *TestRunner) TestStop(stopHookActive bool, transcript []TranscriptEntry) StopResponseInterface {
	event := &StopEvent{
		SessionID:      "test-session",
		StopHookActive: stopHookActive,
		Transcript:     transcript,
		TranscriptPath: "", // Empty path for test
	}

	if t.runner.Stop == nil {
		return Error(fmt.Errorf("Stop handler not set"))
	}

	return t.runner.Stop(context.Background(), event)
}

// Test assertion helpers

// AssertPreToolUseApproves asserts that a PreToolUse handler approves
func (t *TestRunner) AssertPreToolUseApproves(toolName string, toolInput interface{}) error {
	resp := t.TestPreToolUse(toolName, toolInput)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	preResp, ok := resp.(*PreToolUseResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if preResp.Decision != PreToolUseApprove {
		return fmt.Errorf("expected approve, got %s", preResp.Decision)
	}
	return nil
}

// AssertPreToolUseBlocks asserts that a PreToolUse handler blocks
func (t *TestRunner) AssertPreToolUseBlocks(toolName string, toolInput interface{}) error {
	resp := t.TestPreToolUse(toolName, toolInput)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	preResp, ok := resp.(*PreToolUseResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if preResp.Decision != PreToolUseBlock {
		return fmt.Errorf("expected block, got %s", preResp.Decision)
	}
	return nil
}

// AssertPreToolUseBlocksWithReason asserts that a PreToolUse handler blocks with a specific reason
func (t *TestRunner) AssertPreToolUseBlocksWithReason(toolName string, toolInput interface{}, expectedReason string) error {
	resp := t.TestPreToolUse(toolName, toolInput)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	preResp, ok := resp.(*PreToolUseResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if preResp.Decision != PreToolUseBlock {
		return fmt.Errorf("expected block, got %s", preResp.Decision)
	}
	if preResp.Reason != expectedReason {
		return fmt.Errorf("expected reason %q, got %q", expectedReason, preResp.Reason)
	}
	return nil
}

// AssertPreToolUseStopsClaude asserts that a PreToolUse handler stops Claude
func (t *TestRunner) AssertPreToolUseStopsClaude(toolName string, toolInput interface{}) error {
	resp := t.TestPreToolUse(toolName, toolInput)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	preResp, ok := resp.(*PreToolUseResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if preResp.Continue == nil || *preResp.Continue != false {
		return fmt.Errorf("expected continue=false")
	}
	return nil
}

// AssertPostToolUseAllows asserts that a PostToolUse handler allows
func (t *TestRunner) AssertPostToolUseAllows(toolName string, toolInput, toolResponse interface{}) error {
	resp := t.TestPostToolUse(toolName, toolInput, toolResponse)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	postResp, ok := resp.(*PostToolUseResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if postResp.Decision != "" {
		return fmt.Errorf("expected allow (empty decision), got %s", postResp.Decision)
	}
	return nil
}

// AssertPostToolUseBlocks asserts that a PostToolUse handler blocks
func (t *TestRunner) AssertPostToolUseBlocks(toolName string, toolInput, toolResponse interface{}) error {
	resp := t.TestPostToolUse(toolName, toolInput, toolResponse)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	postResp, ok := resp.(*PostToolUseResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if postResp.Decision != PostToolUseBlock {
		return fmt.Errorf("expected block, got %s", postResp.Decision)
	}
	return nil
}

// AssertPostToolUseBlocksWithReason asserts that a PostToolUse handler blocks with a specific reason
func (t *TestRunner) AssertPostToolUseBlocksWithReason(toolName string, toolInput, toolResponse interface{}, expectedReason string) error {
	resp := t.TestPostToolUse(toolName, toolInput, toolResponse)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	postResp, ok := resp.(*PostToolUseResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if postResp.Decision != PostToolUseBlock {
		return fmt.Errorf("expected block, got %s", postResp.Decision)
	}
	if postResp.Reason != expectedReason {
		return fmt.Errorf("expected reason %q, got %q", expectedReason, postResp.Reason)
	}
	return nil
}

// AssertNotificationOK asserts that a Notification handler returns OK
func (t *TestRunner) AssertNotificationOK(message string) error {
	resp := t.TestNotification(message)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	notifResp, ok := resp.(*NotificationResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if notifResp.Continue != nil || notifResp.StopReason != "" {
		return fmt.Errorf("expected empty response, got continue=%v stopReason=%s", notifResp.Continue, notifResp.StopReason)
	}
	return nil
}

// AssertStopContinues asserts that a Stop handler continues
func (t *TestRunner) AssertStopContinues(stopHookActive bool, transcript []TranscriptEntry) error {
	resp := t.TestStop(stopHookActive, transcript)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	stopResp, ok := resp.(*StopResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if stopResp.Decision != "" {
		return fmt.Errorf("expected continue (empty decision), got %s", stopResp.Decision)
	}
	return nil
}

// AssertStopBlocks asserts that a Stop handler blocks
func (t *TestRunner) AssertStopBlocks(stopHookActive bool, transcript []TranscriptEntry) error {
	resp := t.TestStop(stopHookActive, transcript)
	if errResp, ok := resp.(*ErrorResponse); ok {
		return errResp.Error
	}
	stopResp, ok := resp.(*StopResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", resp)
	}
	if stopResp.Decision != StopBlock {
		return fmt.Errorf("expected block, got %s", stopResp.Decision)
	}
	return nil
}
