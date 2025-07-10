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
func (t *TestRunner) TestPreToolUse(toolName string, toolInput interface{}) (*PreToolUseResponse, error) {
	inputJSON, err := json.Marshal(toolInput)
	if err != nil {
		return nil, err
	}

	event := &PreToolUseEvent{
		SessionID: "test-session",
		ToolName:  toolName,
		ToolInput: inputJSON,
	}

	if t.runner.PreToolUse == nil {
		return nil, fmt.Errorf("PreToolUse handler not set")
	}

	return t.runner.PreToolUse(context.Background(), event)
}

// TestPostToolUse tests a PostToolUse handler
func (t *TestRunner) TestPostToolUse(toolName string, toolInput, toolResponse interface{}) (*PostToolUseResponse, error) {
	inputJSON, err := json.Marshal(toolInput)
	if err != nil {
		return nil, err
	}

	responseJSON, err := json.Marshal(toolResponse)
	if err != nil {
		return nil, err
	}

	event := &PostToolUseEvent{
		SessionID:    "test-session",
		ToolName:     toolName,
		ToolInput:    inputJSON,
		ToolResponse: responseJSON,
	}

	if t.runner.PostToolUse == nil {
		return nil, fmt.Errorf("PostToolUse handler not set")
	}

	return t.runner.PostToolUse(context.Background(), event)
}

// TestNotification tests a Notification handler
func (t *TestRunner) TestNotification(message string) (*NotificationResponse, error) {
	event := &NotificationEvent{
		SessionID: "test-session",
		Message:   message,
	}

	if t.runner.Notification == nil {
		return nil, fmt.Errorf("Notification handler not set")
	}

	return t.runner.Notification(context.Background(), event)
}

// TestStop tests a Stop handler
func (t *TestRunner) TestStop(stopHookActive bool, transcript []interface{}) (*StopResponse, error) {
	event := &StopEvent{
		SessionID:      "test-session",
		StopHookActive: stopHookActive,
		Transcript:     transcript,
	}

	if t.runner.Stop == nil {
		return nil, fmt.Errorf("Stop handler not set")
	}

	return t.runner.Stop(context.Background(), event)
}

// Test assertion helpers

// AssertPreToolUseApproves asserts that a PreToolUse handler approves
func (t *TestRunner) AssertPreToolUseApproves(toolName string, toolInput interface{}) error {
	resp, err := t.TestPreToolUse(toolName, toolInput)
	if err != nil {
		return err
	}
	if resp.Decision != PreToolUseApprove {
		return fmt.Errorf("expected approve, got %s", resp.Decision)
	}
	return nil
}

// AssertPreToolUseBlocks asserts that a PreToolUse handler blocks
func (t *TestRunner) AssertPreToolUseBlocks(toolName string, toolInput interface{}) error {
	resp, err := t.TestPreToolUse(toolName, toolInput)
	if err != nil {
		return err
	}
	if resp.Decision != PreToolUseBlock {
		return fmt.Errorf("expected block, got %s", resp.Decision)
	}
	return nil
}

// AssertPreToolUseBlocksWithReason asserts that a PreToolUse handler blocks with a specific reason
func (t *TestRunner) AssertPreToolUseBlocksWithReason(toolName string, toolInput interface{}, expectedReason string) error {
	resp, err := t.TestPreToolUse(toolName, toolInput)
	if err != nil {
		return err
	}
	if resp.Decision != PreToolUseBlock {
		return fmt.Errorf("expected block, got %s", resp.Decision)
	}
	if resp.Reason != expectedReason {
		return fmt.Errorf("expected reason %q, got %q", expectedReason, resp.Reason)
	}
	return nil
}

// AssertPreToolUseStopsClaude asserts that a PreToolUse handler stops Claude
func (t *TestRunner) AssertPreToolUseStopsClaude(toolName string, toolInput interface{}) error {
	resp, err := t.TestPreToolUse(toolName, toolInput)
	if err != nil {
		return err
	}
	if resp.Continue == nil || *resp.Continue != false {
		return fmt.Errorf("expected continue=false")
	}
	return nil
}

// AssertPostToolUseAllows asserts that a PostToolUse handler allows
func (t *TestRunner) AssertPostToolUseAllows(toolName string, toolInput, toolResponse interface{}) error {
	resp, err := t.TestPostToolUse(toolName, toolInput, toolResponse)
	if err != nil {
		return err
	}
	if resp.Decision != "" {
		return fmt.Errorf("expected allow (empty decision), got %s", resp.Decision)
	}
	return nil
}

// AssertPostToolUseBlocks asserts that a PostToolUse handler blocks
func (t *TestRunner) AssertPostToolUseBlocks(toolName string, toolInput, toolResponse interface{}) error {
	resp, err := t.TestPostToolUse(toolName, toolInput, toolResponse)
	if err != nil {
		return err
	}
	if resp.Decision != PostToolUseBlock {
		return fmt.Errorf("expected block, got %s", resp.Decision)
	}
	return nil
}

// AssertPostToolUseBlocksWithReason asserts that a PostToolUse handler blocks with a specific reason
func (t *TestRunner) AssertPostToolUseBlocksWithReason(toolName string, toolInput, toolResponse interface{}, expectedReason string) error {
	resp, err := t.TestPostToolUse(toolName, toolInput, toolResponse)
	if err != nil {
		return err
	}
	if resp.Decision != PostToolUseBlock {
		return fmt.Errorf("expected block, got %s", resp.Decision)
	}
	if resp.Reason != expectedReason {
		return fmt.Errorf("expected reason %q, got %q", expectedReason, resp.Reason)
	}
	return nil
}

// AssertNotificationOK asserts that a Notification handler returns OK
func (t *TestRunner) AssertNotificationOK(message string) error {
	resp, err := t.TestNotification(message)
	if err != nil {
		return err
	}
	if resp.Continue != nil || resp.StopReason != "" {
		return fmt.Errorf("expected empty response, got continue=%v stopReason=%s", resp.Continue, resp.StopReason)
	}
	return nil
}

// AssertStopContinues asserts that a Stop handler continues
func (t *TestRunner) AssertStopContinues(stopHookActive bool, transcript []interface{}) error {
	resp, err := t.TestStop(stopHookActive, transcript)
	if err != nil {
		return err
	}
	if resp.Decision != "" {
		return fmt.Errorf("expected continue (empty decision), got %s", resp.Decision)
	}
	return nil
}

// AssertStopBlocks asserts that a Stop handler blocks
func (t *TestRunner) AssertStopBlocks(stopHookActive bool, transcript []interface{}) error {
	resp, err := t.TestStop(stopHookActive, transcript)
	if err != nil {
		return err
	}
	if resp.Decision != StopBlock {
		return fmt.Errorf("expected block, got %s", resp.Decision)
	}
	return nil
}