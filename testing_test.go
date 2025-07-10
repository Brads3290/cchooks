package cchooks

import (
	"context"
	"errors"
	"testing"
)

func TestTestRunner(t *testing.T) {
	t.Run("TestPreToolUse", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
				if event.ToolName == "Bash" {
					bash, _ := event.AsBash()
					if bash.Command == "rm -rf /" {
						return Block("dangerous")
					}
				}
				return Approve()
			},
		}

		tr := NewTestRunner(runner)

		// Test approve
		resp := tr.TestPreToolUse("Bash", &BashInput{Command: "ls"})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestPreToolUse error = %v", errResp.Error)
		}
		if resp.Decision != PreToolUseApprove {
			t.Errorf("Decision = %q, want %q", resp.Decision, PreToolUseApprove)
		}

		// Test block
		resp = tr.TestPreToolUse("Bash", &BashInput{Command: "rm -rf /"})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestPreToolUse error = %v", errResp.Error)
		}
		if resp.Decision != PreToolUseBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, PreToolUseBlock)
		}
		if resp.Reason != "dangerous" {
			t.Errorf("Reason = %q, want %q", resp.Reason, "dangerous")
		}
	})

	t.Run("TestPostToolUse", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) *PostToolUseResponse {
				bash, _ := event.ResponseAsBash()
				if bash.ExitCode != 0 {
					return PostBlock("command failed")
				}
				return Allow()
			},
		}

		tr := NewTestRunner(runner)

		// Test allow
		resp := tr.TestPostToolUse("Bash", &BashInput{Command: "ls"}, &BashOutput{Output: "files", ExitCode: 0})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestPostToolUse error = %v", errResp.Error)
		}
		if resp.Decision != "" {
			t.Errorf("Decision = %q, want empty", resp.Decision)
		}

		// Test block
		resp = tr.TestPostToolUse("Bash", &BashInput{Command: "bad"}, &BashOutput{Output: "error", ExitCode: 1})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestPostToolUse error = %v", errResp.Error)
		}
		if resp.Decision != PostToolUseBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, PostToolUseBlock)
		}
	})

	t.Run("TestNotification", func(t *testing.T) {
		runner := &Runner{
			Notification: func(ctx context.Context, event *NotificationEvent) *NotificationResponse {
				if event.Message == "error" {
					return StopFromNotification("error occurred")
				}
				return OK()
			},
		}

		tr := NewTestRunner(runner)

		// Test OK
		resp := tr.TestNotification("info")
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestNotification error = %v", errResp.Error)
		}
		if resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected empty response")
		}

		// Test stop
		resp = tr.TestNotification("error")
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestNotification error = %v", errResp.Error)
		}
		if resp.Continue == nil || *resp.Continue != false {
			t.Error("expected continue=false")
		}
		if resp.StopReason != "error occurred" {
			t.Errorf("StopReason = %q, want %q", resp.StopReason, "error occurred")
		}
	})

	t.Run("TestStop", func(t *testing.T) {
		runner := &Runner{
			Stop: func(ctx context.Context, event *StopEvent) *StopResponse {
				if !event.StopHookActive {
					return BlockStop("stop not allowed")
				}
				return Continue()
			},
		}

		tr := NewTestRunner(runner)

		// Test continue
		resp := tr.TestStop(true, []TranscriptEntry{})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestStop error = %v", errResp.Error)
		}
		if resp.Decision != "" {
			t.Errorf("Decision = %q, want empty", resp.Decision)
		}

		// Test block
		resp = tr.TestStop(false, []TranscriptEntry{})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestStop error = %v", errResp.Error)
		}
		if resp.Decision != StopBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, StopBlock)
		}
	})

	t.Run("missing handlers", func(t *testing.T) {
		runner := &Runner{}
		tr := NewTestRunner(runner)

		resp := tr.TestPreToolUse("Bash", &BashInput{Command: "ls"})
		if errResp, ok := resp.(*ErrorResponse); !ok || errResp.Message != "PreToolUse handler not set" {
			t.Errorf("expected handler not set error, got %v", resp)
		}

		resp = tr.TestPostToolUse("Bash", &BashInput{}, &BashOutput{})
		if errResp, ok := resp.(*ErrorResponse); !ok || errResp.Message != "PostToolUse handler not set" {
			t.Errorf("expected handler not set error, got %v", resp)
		}

		resp = tr.TestNotification("test")
		if errResp, ok := resp.(*ErrorResponse); !ok || errResp.Message != "Notification handler not set" {
			t.Errorf("expected handler not set error, got %v", resp)
		}

		resp = tr.TestStop(true, []TranscriptEntry{})
		if errResp, ok := resp.(*ErrorResponse); !ok || errResp.Message != "Stop handler not set" {
			t.Errorf("expected handler not set error, got %v", resp)
		}
	})
}

func TestAssertionHelpers(t *testing.T) {
	t.Run("AssertPreToolUseApproves", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
				return Approve()
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseApproves("Bash", &BashInput{Command: "ls"})
		if err != nil {
			t.Errorf("AssertPreToolUseApproves() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
			return Block("nope")
		}
		err = tr.AssertPreToolUseApproves("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-approve response")
		}
	})

	t.Run("AssertPreToolUseBlocks", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
				return Block("blocked")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseBlocks("Bash", &BashInput{Command: "rm -rf /"})
		if err != nil {
			t.Errorf("AssertPreToolUseBlocks() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
			return Approve()
		}
		err = tr.AssertPreToolUseBlocks("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-block response")
		}
	})

	t.Run("AssertPreToolUseBlocksWithReason", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
				return Block("specific reason")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseBlocksWithReason("Bash", &BashInput{Command: "bad"}, "specific reason")
		if err != nil {
			t.Errorf("AssertPreToolUseBlocksWithReason() error = %v", err)
		}

		// Test wrong reason
		err = tr.AssertPreToolUseBlocksWithReason("Bash", &BashInput{Command: "bad"}, "wrong reason")
		if err == nil {
			t.Error("expected error for wrong reason")
		}
	})

	t.Run("AssertPreToolUseStopsClaude", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
				return StopClaude("stop now")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseStopsClaude("Bash", &BashInput{Command: "ls"})
		if err != nil {
			t.Errorf("AssertPreToolUseStopsClaude() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
			return Approve()
		}
		err = tr.AssertPreToolUseStopsClaude("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-stop response")
		}
	})

	t.Run("AssertPostToolUseAllows", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) *PostToolUseResponse {
				return Allow()
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPostToolUseAllows("Bash", &BashInput{Command: "ls"}, &BashOutput{ExitCode: 0})
		if err != nil {
			t.Errorf("AssertPostToolUseAllows() error = %v", err)
		}

		// Test failure case
		runner.PostToolUse = func(ctx context.Context, event *PostToolUseEvent) *PostToolUseResponse {
			return PostBlock("nope")
		}
		err = tr.AssertPostToolUseAllows("Bash", &BashInput{Command: "ls"}, &BashOutput{ExitCode: 0})
		if err == nil {
			t.Error("expected error for non-allow response")
		}
	})

	t.Run("AssertPostToolUseBlocks", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) *PostToolUseResponse {
				return PostBlock("blocked")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPostToolUseBlocks("Bash", &BashInput{}, &BashOutput{ExitCode: 1})
		if err != nil {
			t.Errorf("AssertPostToolUseBlocks() error = %v", err)
		}
	})

	t.Run("AssertPostToolUseBlocksWithReason", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) *PostToolUseResponse {
				return PostBlock("failed")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPostToolUseBlocksWithReason("Bash", &BashInput{}, &BashOutput{}, "failed")
		if err != nil {
			t.Errorf("AssertPostToolUseBlocksWithReason() error = %v", err)
		}

		err = tr.AssertPostToolUseBlocksWithReason("Bash", &BashInput{}, &BashOutput{}, "wrong")
		if err == nil {
			t.Error("expected error for wrong reason")
		}
	})

	t.Run("AssertNotificationOK", func(t *testing.T) {
		runner := &Runner{
			Notification: func(ctx context.Context, event *NotificationEvent) *NotificationResponse {
				return OK()
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertNotificationOK("test message")
		if err != nil {
			t.Errorf("AssertNotificationOK() error = %v", err)
		}

		// Test failure case
		runner.Notification = func(ctx context.Context, event *NotificationEvent) *NotificationResponse {
			return StopFromNotification("stop")
		}
		err = tr.AssertNotificationOK("test")
		if err == nil {
			t.Error("expected error for non-OK response")
		}
	})

	t.Run("AssertStopContinues", func(t *testing.T) {
		runner := &Runner{
			Stop: func(ctx context.Context, event *StopEvent) *StopResponse {
				return Continue()
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertStopContinues(true, []TranscriptEntry{})
		if err != nil {
			t.Errorf("AssertStopContinues() error = %v", err)
		}
	})

	t.Run("AssertStopBlocks", func(t *testing.T) {
		runner := &Runner{
			Stop: func(ctx context.Context, event *StopEvent) *StopResponse {
				return BlockStop("no stopping")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertStopBlocks(true, []TranscriptEntry{})
		if err != nil {
			t.Errorf("AssertStopBlocks() error = %v", err)
		}
	})

	t.Run("handler errors", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) *PreToolUseResponse {
				return Error(errors.New("handler failed"))
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseApproves("Bash", &BashInput{})
		if err == nil || err.Error() != "handler failed" {
			t.Errorf("expected handler error, got %v", err)
		}
	})
}
