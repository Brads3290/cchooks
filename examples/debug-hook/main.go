package main

import (
	"context"
	"log"
	"os"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	// Create a logger that writes to stderr
	logger := log.New(os.Stderr, "[debug-hook] ", log.LstdFlags)
	logger.Println("Debug hook started")

	runner := &cchooks.Runner{
		// Raw handler to see exactly what's being received
		Raw: func(ctx context.Context, rawJSON string) (*cchooks.RawResponse, error) {
			logger.Printf("Raw JSON received: %s", rawJSON)
			logger.Printf("JSON length: %d bytes", len(rawJSON))

			// Continue with normal processing
			return nil, nil
		},

		// Error handler to log any errors
		Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
			logger.Printf("ERROR: %v", err)
			logger.Printf("Raw JSON that caused error: %s", rawJSON)
			logger.Printf("Raw JSON length: %d", len(rawJSON))

			// Return nil to exit with code 0 (success)
			return nil
		},

		// Handle all event types
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
			logger.Printf("PreToolUse: Tool=%s, Session=%s", event.ToolName, event.SessionID)
			return cchooks.Approve(), nil
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
			logger.Printf("PostToolUse: Tool=%s, Session=%s", event.ToolName, event.SessionID)
			return cchooks.Allow(), nil
		},

		Notification: func(ctx context.Context, event *cchooks.NotificationEvent) (*cchooks.NotificationResponse, error) {
			logger.Printf("Notification: Session=%s", event.SessionID)
			return cchooks.OK(), nil
		},

		Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
			logger.Printf("Stop: SessionID=%s, StopHookActive=%v",
				event.SessionID, event.StopHookActive)
			return cchooks.Continue(), nil
		},
	}

	logger.Println("Running hook...")
	runner.Run()
	logger.Println("Hook completed")
}
