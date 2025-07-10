package cchooks

import (
	"context"
	"errors"
	"testing"
)

func TestTestRunner(t *testing.T) {
	t.Run("TestPreToolUse", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
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
		if preResp, ok := resp.(*PreToolUseResponse); ok {
			if preResp.Decision != PreToolUseApprove {
				t.Errorf("Decision = %q, want %q", preResp.Decision, PreToolUseApprove)
			}
		} else {
			t.Error("expected *PreToolUseResponse")
		}

		// Test block
		resp = tr.TestPreToolUse("Bash", &BashInput{Command: "rm -rf /"})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestPreToolUse error = %v", errResp.Error)
		}
		if preResp, ok := resp.(*PreToolUseResponse); ok {
			if preResp.Decision != PreToolUseBlock {
				t.Errorf("Decision = %q, want %q", preResp.Decision, PreToolUseBlock)
			}
			if preResp.Reason != "dangerous" {
				t.Errorf("Reason = %q, want %q", preResp.Reason, "dangerous")
			}
		} else {
			t.Error("expected *PreToolUseResponse")
		}
	})

	t.Run("TestPostToolUse", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) PostToolUseResponseInterface {
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
		if postResp, ok := resp.(*PostToolUseResponse); ok {
			if postResp.Decision != "" {
				t.Errorf("Decision = %q, want empty", postResp.Decision)
			}
		} else {
			t.Error("expected *PostToolUseResponse")
		}

		// Test block
		resp = tr.TestPostToolUse("Bash", &BashInput{Command: "bad"}, &BashOutput{Output: "error", ExitCode: 1})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestPostToolUse error = %v", errResp.Error)
		}
		if postResp, ok := resp.(*PostToolUseResponse); ok {
			if postResp.Decision != PostToolUseBlock {
				t.Errorf("Decision = %q, want %q", postResp.Decision, PostToolUseBlock)
			}
		} else {
			t.Error("expected *PostToolUseResponse")
		}
	})

	t.Run("TestNotification", func(t *testing.T) {
		runner := &Runner{
			Notification: func(ctx context.Context, event *NotificationEvent) NotificationResponseInterface {
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
		if notifResp, ok := resp.(*NotificationResponse); ok {
			if notifResp.Continue != nil || notifResp.StopReason != "" {
				t.Error("expected empty response")
			}
		} else {
			t.Error("expected *NotificationResponse")
		}

		// Test stop
		resp = tr.TestNotification("error")
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestNotification error = %v", errResp.Error)
		}
		if notifResp, ok := resp.(*NotificationResponse); ok {
			if notifResp.Continue == nil || *notifResp.Continue != false {
				t.Error("expected continue=false")
			}
			if notifResp.StopReason != "error occurred" {
				t.Errorf("StopReason = %q, want %q", notifResp.StopReason, "error occurred")
			}
		} else {
			t.Error("expected *NotificationResponse")
		}
	})

	t.Run("TestStop", func(t *testing.T) {
		runner := &Runner{
			Stop: func(ctx context.Context, event *StopEvent) StopResponseInterface {
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
		if stopResp, ok := resp.(*StopResponse); ok {
			if stopResp.Decision != "" {
				t.Errorf("Decision = %q, want empty", stopResp.Decision)
			}
		} else {
			t.Error("expected *StopResponse")
		}

		// Test block
		resp = tr.TestStop(false, []TranscriptEntry{})
		if errResp, ok := resp.(*ErrorResponse); ok {
			t.Fatalf("TestStop error = %v", errResp.Error)
		}
		if stopResp, ok := resp.(*StopResponse); ok {
			if stopResp.Decision != StopBlock {
				t.Errorf("Decision = %q, want %q", stopResp.Decision, StopBlock)
			}
		} else {
			t.Error("expected *StopResponse")
		}
	})

	t.Run("missing handlers", func(t *testing.T) {
		runner := &Runner{}
		tr := NewTestRunner(runner)

		resp := tr.TestPreToolUse("Bash", &BashInput{Command: "ls"})
		if errResp, ok := resp.(*ErrorResponse); !ok || errResp.Message != "PreToolUse handler not set" {
			t.Errorf("expected handler not set error, got %v", resp)
		}

		resp2 := tr.TestPostToolUse("Bash", &BashInput{}, &BashOutput{})
		if errResp, ok := resp2.(*ErrorResponse); !ok || errResp.Message != "PostToolUse handler not set" {
			t.Errorf("expected handler not set error, got %v", resp2)
		}

		resp3 := tr.TestNotification("test")
		if errResp, ok := resp3.(*ErrorResponse); !ok || errResp.Message != "Notification handler not set" {
			t.Errorf("expected handler not set error, got %v", resp3)
		}

		resp4 := tr.TestStop(true, []TranscriptEntry{})
		if errResp, ok := resp4.(*ErrorResponse); !ok || errResp.Message != "Stop handler not set" {
			t.Errorf("expected handler not set error, got %v", resp4)
		}
	})
}

func TestAssertionHelpers(t *testing.T) {
	t.Run("AssertPreToolUseApproves", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
				return Approve()
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseApproves("Bash", &BashInput{Command: "ls"})
		if err != nil {
			t.Errorf("AssertPreToolUseApproves() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
			return Block("nope")
		}
		err = tr.AssertPreToolUseApproves("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-approve response")
		}
	})

	t.Run("AssertPreToolUseBlocks", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
				return Block("blocked")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseBlocks("Bash", &BashInput{Command: "rm -rf /"})
		if err != nil {
			t.Errorf("AssertPreToolUseBlocks() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
			return Approve()
		}
		err = tr.AssertPreToolUseBlocks("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-block response")
		}
	})

	t.Run("AssertPreToolUseBlocksWithReason", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
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
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
				return StopClaude("stop now")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseStopsClaude("Bash", &BashInput{Command: "ls"})
		if err != nil {
			t.Errorf("AssertPreToolUseStopsClaude() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
			return Approve()
		}
		err = tr.AssertPreToolUseStopsClaude("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-stop response")
		}
	})

	t.Run("AssertPostToolUseAllows", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) PostToolUseResponseInterface {
				return Allow()
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPostToolUseAllows("Bash", &BashInput{Command: "ls"}, &BashOutput{ExitCode: 0})
		if err != nil {
			t.Errorf("AssertPostToolUseAllows() error = %v", err)
		}

		// Test failure case
		runner.PostToolUse = func(ctx context.Context, event *PostToolUseEvent) PostToolUseResponseInterface {
			return PostBlock("nope")
		}
		err = tr.AssertPostToolUseAllows("Bash", &BashInput{Command: "ls"}, &BashOutput{ExitCode: 0})
		if err == nil {
			t.Error("expected error for non-allow response")
		}
	})

	t.Run("AssertPostToolUseBlocks", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) PostToolUseResponseInterface {
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
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) PostToolUseResponseInterface {
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
			Notification: func(ctx context.Context, event *NotificationEvent) NotificationResponseInterface {
				return OK()
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertNotificationOK("test message")
		if err != nil {
			t.Errorf("AssertNotificationOK() error = %v", err)
		}

		// Test failure case
		runner.Notification = func(ctx context.Context, event *NotificationEvent) NotificationResponseInterface {
			return StopFromNotification("stop")
		}
		err = tr.AssertNotificationOK("test")
		if err == nil {
			t.Error("expected error for non-OK response")
		}
	})

	t.Run("AssertStopContinues", func(t *testing.T) {
		runner := &Runner{
			Stop: func(ctx context.Context, event *StopEvent) StopResponseInterface {
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
			Stop: func(ctx context.Context, event *StopEvent) StopResponseInterface {
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
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) PreToolUseResponseInterface {
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
