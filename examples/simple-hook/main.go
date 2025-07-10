package main

import (
	"context"
	"log"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
			log.Printf("PreToolUse: tool=%s session=%s", event.ToolName, event.SessionID)

			// Example: Block git push to main branch
			if event.ToolName == "Bash" {
				bash, err := event.AsBash()
				if err == nil && bash.Command == "git push origin main" {
					return cchooks.Block("Direct pushes to main branch are not allowed. Please create a pull request."), nil
				}
			}

			return cchooks.Approve(), nil
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
			log.Printf("PostToolUse: tool=%s session=%s", event.ToolName, event.SessionID)

			// Example: Check for test failures
			if event.ToolName == "Bash" {
				input, _ := event.InputAsBash()
				output, _ := event.ResponseAsBash()

				if input != nil && output != nil {
					if (input.Command == "npm test" || input.Command == "go test ./...") && output.ExitCode != 0 {
						return cchooks.StopClaudePost("Tests failed. Please fix the failing tests before continuing."), nil
					}
				}
			}

			return cchooks.Allow(), nil
		},

		Notification: func(ctx context.Context, event *cchooks.NotificationEvent) (*cchooks.NotificationResponse, error) {
			log.Printf("Notification: %s", event.Message)
			return cchooks.OK(), nil
		},

		Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
			log.Printf("Session %s stopped", event.SessionID)
			return cchooks.Continue(), nil
		},
	}

	if err := runner.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
