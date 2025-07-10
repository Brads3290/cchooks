package main

import (
	"context"
	"log"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) cchooks.PreToolUseResponseInterface {
			log.Printf("PreToolUse: tool=%s session=%s", event.ToolName, event.SessionID)

			// Example: Block git push to main branch
			if event.ToolName == "Bash" {
				bash, err := event.AsBash()
				if err == nil && bash.Command == "git push origin main" {
					return cchooks.Block("Direct pushes to main branch are not allowed. Please create a pull request.")
				}
			}

			return cchooks.Approve()
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) cchooks.PostToolUseResponseInterface {
			log.Printf("PostToolUse: tool=%s session=%s", event.ToolName, event.SessionID)

			// Example: Check for test failures
			if event.ToolName == "Bash" {
				input, _ := event.InputAsBash()
				output, _ := event.ResponseAsBash()

				if input != nil && output != nil {
					if (input.Command == "npm test" || input.Command == "go test ./...") && output.ExitCode != 0 {
						return cchooks.StopClaudePost("Tests failed. Please fix the failing tests before continuing.")
					}
				}
			}

			return cchooks.Allow()
		},

		Notification: func(ctx context.Context, event *cchooks.NotificationEvent) cchooks.NotificationResponseInterface {
			log.Printf("Notification: %s", event.Message)
			return cchooks.OK()
		},

		Stop: func(ctx context.Context, event *cchooks.StopEvent) cchooks.StopResponseInterface {
			log.Printf("Session %s stopped", event.SessionID)
			return cchooks.Continue()
		},
	}

	runner.Run()
}
