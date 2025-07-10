package cchooks

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// osExit is a variable to allow mocking os.Exit in tests
var osExit = os.Exit

// Runner handles event dispatch and I/O for a single hook binary
type Runner struct {
	// Raw is called before any other processing with the raw JSON string
	// If it returns a non-nil response, that response is used and no further processing occurs
	// If it returns nil, normal event processing continues
	Raw          func(context.Context, string) (*RawResponse, error)
	PreToolUse   func(context.Context, *PreToolUseEvent) (*PreToolUseResponse, error)
	PostToolUse  func(context.Context, *PostToolUseEvent) (*PostToolUseResponse, error)
	Notification func(context.Context, *NotificationEvent) (*NotificationResponse, error)
	Stop         func(context.Context, *StopEvent) (*StopResponse, error)
	// StopOnce is called for Stop events only when stop_hook_active is false
	// This allows hooks to handle the first stop event differently
	// If both Stop and StopOnce are defined, StopOnce takes precedence when stop_hook_active is false
	StopOnce     func(context.Context, *StopEvent) (*StopResponse, error)
	// Error is called when any error occurs inside the SDK
	// It receives the raw JSON string that was passed to the hook and the error
	// If it returns a non-nil RawResponse, that response is used instead of the default error handling
	// If it returns nil, the SDK will use exit code 2 and output the error to stderr
	Error        func(ctx context.Context, rawJSON string, err error) *RawResponse
}

// Run reads from stdin, dispatches to appropriate handler, outputs response
func (r *Runner) Run(ctx context.Context) error {
	// Read all input for error handling
	var rawJSON []byte
	rawJSON, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	// Set up panic recovery
	defer func() {
		if p := recover(); p != nil {
			// Don't catch test exit panics
			if p == "exit" {
				panic(p)
			}
			
			// Convert panic to error
			var err error
			switch v := p.(type) {
			case error:
				err = fmt.Errorf("panic: %w", v)
			case string:
				err = fmt.Errorf("panic: %s", v)
			default:
				err = fmt.Errorf("panic: %v", v)
			}

			// Handle error using handleError which will use Error handler if available
			r.handleError(ctx, string(rawJSON), err)
		}
	}()

	// Call Raw handler if provided
	if r.Raw != nil {
		response, err := r.Raw(ctx, string(rawJSON))
		if err != nil {
			r.handleError(ctx, string(rawJSON), err)
			return nil // handleError exits, so this is unreachable
		}
		
		// If Raw handler returns a response, use it and exit
		if response != nil {
			if response.Output != "" {
				fmt.Fprint(os.Stdout, response.Output)
			}
			osExit(response.ExitCode)
		}
		// If Raw handler returns nil, continue with normal processing
	}

	// Parse JSON
	var rawEvent map[string]interface{}
	if err := json.Unmarshal(rawJSON, &rawEvent); err != nil {
		err = fmt.Errorf("failed to decode stdin: %w", err)
		r.handleError(ctx, string(rawJSON), err)
		return nil // handleError exits, so this is unreachable
	}

	// Check for hook_event_name field (the actual field name used by Claude Code)
	event, ok := rawEvent["hook_event_name"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid hook_event_name field")
		r.handleError(ctx, string(rawJSON), err)
		return nil // handleError exits, so this is unreachable
	}

	// Dispatch to appropriate handler
	var dispatchErr error
	switch event {
	case "PreToolUse":
		dispatchErr = r.handlePreToolUse(ctx, rawEvent, string(rawJSON))
	case "PostToolUse":
		dispatchErr = r.handlePostToolUse(ctx, rawEvent, string(rawJSON))
	case "Notification":
		dispatchErr = r.handleNotification(ctx, rawEvent, string(rawJSON))
	case "Stop":
		dispatchErr = r.handleStop(ctx, rawEvent, string(rawJSON))
	default:
		dispatchErr = fmt.Errorf("unknown event type: %s", event)
	}
	
	if dispatchErr != nil {
		r.handleError(ctx, string(rawJSON), dispatchErr)
		return nil // handleError exits, so this is unreachable
	}
	
	return nil
}

func (r *Runner) handlePreToolUse(ctx context.Context, rawEvent map[string]interface{}, rawJSON string) error {
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
		return err
	}

	// Handle response
	if err := outputResponse(response); err != nil {
		return err
	}
	return nil
}

func (r *Runner) handlePostToolUse(ctx context.Context, rawEvent map[string]interface{}, rawJSON string) error {
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
		return err
	}

	// Handle response
	if err := outputResponse(response); err != nil {
		return err
	}
	return nil
}

func (r *Runner) handleNotification(ctx context.Context, rawEvent map[string]interface{}, rawJSON string) error {
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
		return err
	}

	// Handle response
	if err := outputResponse(response); err != nil {
		return err
	}
	return nil
}

func (r *Runner) handleStop(ctx context.Context, rawEvent map[string]interface{}, rawJSON string) error {
	// Parse event first to check stop_hook_active
	eventData, err := json.Marshal(rawEvent)
	if err != nil {
		return err
	}

	var event StopEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse StopEvent: %w", err)
	}

	// Read and parse transcript if transcript_path is provided
	if event.TranscriptPath != "" {
		transcript, err := readTranscript(event.TranscriptPath)
		if err != nil {
			// Log error but don't fail - transcript is optional enrichment
			// The handler can still work without it
			event.Transcript = []TranscriptEntry{}
		} else {
			event.Transcript = transcript
		}
	} else {
		// Ensure transcript is never nil
		event.Transcript = []TranscriptEntry{}
	}

	// Determine which handler to use
	var handler func(context.Context, *StopEvent) (*StopResponse, error)
	
	// If stop_hook_active is false and StopOnce is defined, use StopOnce
	if !event.StopHookActive && r.StopOnce != nil {
		handler = r.StopOnce
	} else if r.Stop != nil {
		// Otherwise use the regular Stop handler if defined
		handler = r.Stop
	}

	// If no appropriate handler is found, return nil
	if handler == nil {
		return nil
	}

	// Call the selected handler
	response, err := handler(ctx, &event)
	if err != nil {
		return err
	}

	// Handle response
	if err := outputResponse(response); err != nil {
		return err
	}
	return nil
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

// handleError calls the Error handler if available and handles the response
// If no Error handler or it returns nil, uses default error handling
// Default exit code is 2, except for Stop events which use 0 to avoid blocking Claude from stopping
func (r *Runner) handleError(ctx context.Context, rawJSON string, err error) {
	if r.Error != nil {
		if response := r.Error(ctx, rawJSON, err); response != nil {
			// Use the custom response
			if response.Output != "" {
				fmt.Fprint(os.Stdout, response.Output)
			}
			osExit(response.ExitCode)
			return
		}
	}
	
	// Default error handling
	fmt.Fprintf(os.Stderr, "%v\n", err)
	
	// Determine exit code based on event type
	exitCode := 2 // Default for most errors
	
	// Parse the event type from rawJSON to check if it's a Stop event
	var eventData map[string]interface{}
	if json.Unmarshal([]byte(rawJSON), &eventData) == nil {
		if event, ok := eventData["hook_event_name"].(string); ok && event == "Stop" {
			exitCode = 0 // Don't block Claude from stopping
		}
	}
	
	osExit(exitCode)
}

// readTranscript reads a JSONL transcript file and returns parsed entries
func readTranscript(path string) ([]TranscriptEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript file: %w", err)
	}
	defer file.Close()

	var entries []TranscriptEntry
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Continue on error - some lines might be malformed
			// but we want to read as much as possible
			continue
		}
		
		entries = append(entries, entry)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading transcript file: %w", err)
	}
	
	return entries, nil
}
