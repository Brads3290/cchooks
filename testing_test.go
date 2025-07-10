package cchooks

import (
	"context"
	"errors"
	"testing"
)

func TestTestRunner(t *testing.T) {
	t.Run("TestPreToolUse", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
				if event.ToolName == "Bash" {
					bash, _ := event.AsBash()
					if bash.Command == "rm -rf /" {
						return Block("dangerous"), nil
					}
				}
				return Approve(), nil
			},
		}

		tr := NewTestRunner(runner)

		// Test approve
		resp, err := tr.TestPreToolUse("Bash", &BashInput{Command: "ls"})
		if err != nil {
			t.Fatalf("TestPreToolUse error = %v", err)
		}
		if resp.Decision != PreToolUseApprove {
			t.Errorf("Decision = %q, want %q", resp.Decision, PreToolUseApprove)
		}

		// Test block
		resp, err = tr.TestPreToolUse("Bash", &BashInput{Command: "rm -rf /"})
		if err != nil {
			t.Fatalf("TestPreToolUse error = %v", err)
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
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) (*PostToolUseResponse, error) {
				bash, _ := event.ResponseAsBash()
				if bash.ExitCode != 0 {
					return PostBlock("command failed"), nil
				}
				return Allow(), nil
			},
		}

		tr := NewTestRunner(runner)

		// Test allow
		resp, err := tr.TestPostToolUse("Bash", &BashInput{Command: "ls"}, &BashOutput{Output: "files", ExitCode: 0})
		if err != nil {
			t.Fatalf("TestPostToolUse error = %v", err)
		}
		if resp.Decision != "" {
			t.Errorf("Decision = %q, want empty", resp.Decision)
		}

		// Test block
		resp, err = tr.TestPostToolUse("Bash", &BashInput{Command: "bad"}, &BashOutput{Output: "error", ExitCode: 1})
		if err != nil {
			t.Fatalf("TestPostToolUse error = %v", err)
		}
		if resp.Decision != PostToolUseBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, PostToolUseBlock)
		}
	})

	t.Run("TestNotification", func(t *testing.T) {
		runner := &Runner{
			Notification: func(ctx context.Context, event *NotificationEvent) (*NotificationResponse, error) {
				if event.Message == "error" {
					return StopFromNotification("error occurred"), nil
				}
				return OK(), nil
			},
		}

		tr := NewTestRunner(runner)

		// Test OK
		resp, err := tr.TestNotification("info")
		if err != nil {
			t.Fatalf("TestNotification error = %v", err)
		}
		if resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected empty response")
		}

		// Test stop
		resp, err = tr.TestNotification("error")
		if err != nil {
			t.Fatalf("TestNotification error = %v", err)
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
			Stop: func(ctx context.Context, event *StopEvent) (*StopResponse, error) {
				if !event.StopHookActive {
					return BlockStop("stop not allowed"), nil
				}
				return Continue(), nil
			},
		}

		tr := NewTestRunner(runner)

		// Test continue
		resp, err := tr.TestStop(true, []interface{}{})
		if err != nil {
			t.Fatalf("TestStop error = %v", err)
		}
		if resp.Decision != "" {
			t.Errorf("Decision = %q, want empty", resp.Decision)
		}

		// Test block
		resp, err = tr.TestStop(false, []interface{}{})
		if err != nil {
			t.Fatalf("TestStop error = %v", err)
		}
		if resp.Decision != StopBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, StopBlock)
		}
	})

	t.Run("missing handlers", func(t *testing.T) {
		runner := &Runner{}
		tr := NewTestRunner(runner)

		_, err := tr.TestPreToolUse("Bash", &BashInput{Command: "ls"})
		if err == nil || err.Error() != "PreToolUse handler not set" {
			t.Errorf("expected handler not set error, got %v", err)
		}

		_, err = tr.TestPostToolUse("Bash", &BashInput{}, &BashOutput{})
		if err == nil || err.Error() != "PostToolUse handler not set" {
			t.Errorf("expected handler not set error, got %v", err)
		}

		_, err = tr.TestNotification("test")
		if err == nil || err.Error() != "Notification handler not set" {
			t.Errorf("expected handler not set error, got %v", err)
		}

		_, err = tr.TestStop(true, nil)
		if err == nil || err.Error() != "Stop handler not set" {
			t.Errorf("expected handler not set error, got %v", err)
		}
	})
}

func TestAssertionHelpers(t *testing.T) {
	t.Run("AssertPreToolUseApproves", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
				return Approve(), nil
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseApproves("Bash", &BashInput{Command: "ls"})
		if err != nil {
			t.Errorf("AssertPreToolUseApproves() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
			return Block("nope"), nil
		}
		err = tr.AssertPreToolUseApproves("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-approve response")
		}
	})

	t.Run("AssertPreToolUseBlocks", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
				return Block("blocked"), nil
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseBlocks("Bash", &BashInput{Command: "rm -rf /"})
		if err != nil {
			t.Errorf("AssertPreToolUseBlocks() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
			return Approve(), nil
		}
		err = tr.AssertPreToolUseBlocks("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-block response")
		}
	})

	t.Run("AssertPreToolUseBlocksWithReason", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
				return Block("specific reason"), nil
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
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
				return StopClaude("stop now"), nil
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseStopsClaude("Bash", &BashInput{Command: "ls"})
		if err != nil {
			t.Errorf("AssertPreToolUseStopsClaude() error = %v", err)
		}

		// Test failure case
		runner.PreToolUse = func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
			return Approve(), nil
		}
		err = tr.AssertPreToolUseStopsClaude("Bash", &BashInput{Command: "ls"})
		if err == nil {
			t.Error("expected error for non-stop response")
		}
	})

	t.Run("AssertPostToolUseAllows", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) (*PostToolUseResponse, error) {
				return Allow(), nil
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPostToolUseAllows("Bash", &BashInput{Command: "ls"}, &BashOutput{ExitCode: 0})
		if err != nil {
			t.Errorf("AssertPostToolUseAllows() error = %v", err)
		}

		// Test failure case
		runner.PostToolUse = func(ctx context.Context, event *PostToolUseEvent) (*PostToolUseResponse, error) {
			return PostBlock("nope"), nil
		}
		err = tr.AssertPostToolUseAllows("Bash", &BashInput{Command: "ls"}, &BashOutput{ExitCode: 0})
		if err == nil {
			t.Error("expected error for non-allow response")
		}
	})

	t.Run("AssertPostToolUseBlocks", func(t *testing.T) {
		runner := &Runner{
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) (*PostToolUseResponse, error) {
				return PostBlock("blocked"), nil
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
			PostToolUse: func(ctx context.Context, event *PostToolUseEvent) (*PostToolUseResponse, error) {
				return PostBlock("failed"), nil
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
			Notification: func(ctx context.Context, event *NotificationEvent) (*NotificationResponse, error) {
				return OK(), nil
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertNotificationOK("test message")
		if err != nil {
			t.Errorf("AssertNotificationOK() error = %v", err)
		}

		// Test failure case
		runner.Notification = func(ctx context.Context, event *NotificationEvent) (*NotificationResponse, error) {
			return StopFromNotification("stop"), nil
		}
		err = tr.AssertNotificationOK("test")
		if err == nil {
			t.Error("expected error for non-OK response")
		}
	})

	t.Run("AssertStopContinues", func(t *testing.T) {
		runner := &Runner{
			Stop: func(ctx context.Context, event *StopEvent) (*StopResponse, error) {
				return Continue(), nil
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertStopContinues(true, nil)
		if err != nil {
			t.Errorf("AssertStopContinues() error = %v", err)
		}
	})

	t.Run("AssertStopBlocks", func(t *testing.T) {
		runner := &Runner{
			Stop: func(ctx context.Context, event *StopEvent) (*StopResponse, error) {
				return BlockStop("no stopping"), nil
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertStopBlocks(true, nil)
		if err != nil {
			t.Errorf("AssertStopBlocks() error = %v", err)
		}
	})

	t.Run("handler errors", func(t *testing.T) {
		runner := &Runner{
			PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
				return nil, errors.New("handler failed")
			},
		}
		tr := NewTestRunner(runner)

		err := tr.AssertPreToolUseApproves("Bash", &BashInput{})
		if err == nil || err.Error() != "handler failed" {
			t.Errorf("expected handler error, got %v", err)
		}
	})
}