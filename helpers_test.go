package cchooks

import (
	"testing"
)

func TestHelpers(t *testing.T) {
	t.Run("Approve", func(t *testing.T) {
		resp := Approve()
		if resp.Decision != PreToolUseApprove {
			t.Errorf("Decision = %q, want %q", resp.Decision, PreToolUseApprove)
		}
		if resp.Reason != "" || resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected only Decision field to be set")
		}
	})

	t.Run("Block", func(t *testing.T) {
		resp := Block("test reason")
		if resp.Decision != PreToolUseBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, PreToolUseBlock)
		}
		if resp.Reason != "test reason" {
			t.Errorf("Reason = %q, want %q", resp.Reason, "test reason")
		}
		if resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected only Decision and Reason fields to be set")
		}
	})

	t.Run("PostBlock", func(t *testing.T) {
		resp := PostBlock("post reason")
		if resp.Decision != PostToolUseBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, PostToolUseBlock)
		}
		if resp.Reason != "post reason" {
			t.Errorf("Reason = %q, want %q", resp.Reason, "post reason")
		}
	})

	t.Run("StopClaude", func(t *testing.T) {
		resp := StopClaude("stop reason")
		if resp.Continue == nil || *resp.Continue != false {
			t.Error("expected Continue to be false")
		}
		if resp.StopReason != "stop reason" {
			t.Errorf("StopReason = %q, want %q", resp.StopReason, "stop reason")
		}
		if resp.Decision != "" || resp.Reason != "" {
			t.Error("expected only Continue and StopReason fields to be set")
		}
	})

	t.Run("StopClaudePost", func(t *testing.T) {
		resp := StopClaudePost("stop post")
		if resp.Continue == nil || *resp.Continue != false {
			t.Error("expected Continue to be false")
		}
		if resp.StopReason != "stop post" {
			t.Errorf("StopReason = %q, want %q", resp.StopReason, "stop post")
		}
	})

	t.Run("Allow", func(t *testing.T) {
		resp := Allow()
		if resp.Decision != "" || resp.Reason != "" || resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected all fields to be empty")
		}
	})

	t.Run("OK", func(t *testing.T) {
		resp := OK()
		if resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected all fields to be empty")
		}
	})

	t.Run("StopFromNotification", func(t *testing.T) {
		resp := StopFromNotification("notif stop")
		if resp.Continue == nil || *resp.Continue != false {
			t.Error("expected Continue to be false")
		}
		if resp.StopReason != "notif stop" {
			t.Errorf("StopReason = %q, want %q", resp.StopReason, "notif stop")
		}
	})

	t.Run("Continue", func(t *testing.T) {
		resp := Continue()
		if resp.Decision != "" || resp.Reason != "" || resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected all fields to be empty")
		}
	})

	t.Run("BlockStop", func(t *testing.T) {
		resp := BlockStop("block stop")
		if resp.Decision != StopBlock {
			t.Errorf("Decision = %q, want %q", resp.Decision, StopBlock)
		}
		if resp.Reason != "block stop" {
			t.Errorf("Reason = %q, want %q", resp.Reason, "block stop")
		}
		if resp.Continue != nil || resp.StopReason != "" {
			t.Error("expected only Decision and Reason fields to be set")
		}
	})

	t.Run("StopFromStop", func(t *testing.T) {
		resp := StopFromStop("final stop")
		if resp.Continue == nil || *resp.Continue != false {
			t.Error("expected Continue to be false")
		}
		if resp.StopReason != "final stop" {
			t.Errorf("StopReason = %q, want %q", resp.StopReason, "final stop")
		}
		if resp.Decision != "" || resp.Reason != "" {
			t.Error("expected only Continue and StopReason fields to be set")
		}
	})
}
