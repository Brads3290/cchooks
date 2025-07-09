package cchooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// osExit is a variable to allow mocking os.Exit in tests
var osExit = os.Exit

// Runner handles event dispatch and I/O for a single hook binary
type Runner struct {
	PreToolUse   func(context.Context, *PreToolUseEvent) (*PreToolUseResponse, error)
	PostToolUse  func(context.Context, *PostToolUseEvent) (*PostToolUseResponse, error)
	Notification func(context.Context, *NotificationEvent) (*NotificationResponse, error)
	Stop         func(context.Context, *StopEvent) (*StopResponse, error)
}

// Run reads from stdin, dispatches to appropriate handler, outputs response
func (r *Runner) Run(ctx context.Context) error {
	// Read and parse JSON from stdin
	var rawEvent map[string]interface{}
	if err := json.NewDecoder(os.Stdin).Decode(&rawEvent); err != nil {
		return fmt.Errorf("failed to decode stdin: %w", err)
	}

	event, ok := rawEvent["event"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid event field")
	}

	// Dispatch to appropriate handler
	switch event {
	case "PreToolUse":
		return r.handlePreToolUse(ctx, rawEvent)
	case "PostToolUse":
		return r.handlePostToolUse(ctx, rawEvent)
	case "Notification":
		return r.handleNotification(ctx, rawEvent)
	case "Stop":
		return r.handleStop(ctx, rawEvent)
	default:
		return fmt.Errorf("unknown event type: %s", event)
	}
}

func (r *Runner) handlePreToolUse(ctx context.Context, rawEvent map[string]interface{}) error {
	if r.PreToolUse == nil {
		return nil
	}

	// Parse event
	eventData, err := json.Marshal(rawEvent)
	if err != nil {
		return err
	}

	var event PreToolUseEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse PreToolUseEvent: %w", err)
	}

	// Call handler
	response, err := r.PreToolUse(ctx, &event)
	if err != nil {
		// Exit code 2 sends error to Claude
		fmt.Fprintf(os.Stderr, "%v\n", err)
		osExit(2)
	}

	// Handle response
	return outputResponse(response)
}

func (r *Runner) handlePostToolUse(ctx context.Context, rawEvent map[string]interface{}) error {
	if r.PostToolUse == nil {
		return nil
	}

	// Parse event
	eventData, err := json.Marshal(rawEvent)
	if err != nil {
		return err
	}

	var event PostToolUseEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse PostToolUseEvent: %w", err)
	}

	// Call handler
	response, err := r.PostToolUse(ctx, &event)
	if err != nil {
		// Exit code 2 sends error to Claude
		fmt.Fprintf(os.Stderr, "%v\n", err)
		osExit(2)
	}

	// Handle response
	return outputResponse(response)
}

func (r *Runner) handleNotification(ctx context.Context, rawEvent map[string]interface{}) error {
	if r.Notification == nil {
		return nil
	}

	// Parse event
	eventData, err := json.Marshal(rawEvent)
	if err != nil {
		return err
	}

	var event NotificationEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse NotificationEvent: %w", err)
	}

	// Call handler
	response, err := r.Notification(ctx, &event)
	if err != nil {
		// Exit code 2 sends error to Claude
		fmt.Fprintf(os.Stderr, "%v\n", err)
		osExit(2)
	}

	// Handle response
	return outputResponse(response)
}

func (r *Runner) handleStop(ctx context.Context, rawEvent map[string]interface{}) error {
	if r.Stop == nil {
		return nil
	}

	// Parse event
	eventData, err := json.Marshal(rawEvent)
	if err != nil {
		return err
	}

	var event StopEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse StopEvent: %w", err)
	}

	// Call handler
	response, err := r.Stop(ctx, &event)
	if err != nil {
		// Exit code 2 sends error to Claude
		fmt.Fprintf(os.Stderr, "%v\n", err)
		osExit(2)
	}

	// Handle response
	return outputResponse(response)
}

func outputResponse(response interface{}) error {
	// Check if response is empty (allow action)
	if isEmpty(response) {
		// Empty response uses exit code 0
		return nil
	}

	// Non-empty response uses JSON output
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

func isEmpty(response interface{}) bool {
	switch v := response.(type) {
	case *PreToolUseResponse:
		return v.Decision == "" && v.Continue == nil && v.StopReason == "" && v.Reason == ""
	case *PostToolUseResponse:
		return v.Decision == "" && v.Continue == nil && v.StopReason == "" && v.Reason == ""
	case *NotificationResponse:
		return v.Continue == nil && v.StopReason == ""
	case *StopResponse:
		return v.Decision == "" && v.Continue == nil && v.StopReason == "" && v.Reason == ""
	default:
		return false
	}
}
