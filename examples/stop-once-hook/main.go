package main

import (
	"context"
	"log"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		// StopOnce only triggers when stop_hook_active is false (first stop)
		StopOnce: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
			// This will only run on the first stop event
			// You could use this to save state, send notifications, etc.
			log.Printf("First stop detected! Session: %s\n", event.SessionID)

			// Block the first stop to allow cleanup or confirmation
			return cchooks.BlockStop("Please confirm you want to stop")
		},

		// Regular Stop handler for subsequent stops (when stop_hook_active is true)
		Stop: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
			// This runs on all subsequent stop attempts
			log.Printf("Stop attempt %v for session: %s\n", event.StopHookActive, event.SessionID)

			// Allow subsequent stops
			return cchooks.Continue()
		},
	}

	runner.Run()
}
