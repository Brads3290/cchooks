package main

import (
	"context"
	"log"
	"strings"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		// Analyze transcript on stop events
		Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
			// Log basic information
			log.Printf("Stop event received. Session: %s, StopHookActive: %v\n", 
				event.SessionID, event.StopHookActive)
			
			// Analyze transcript if available
			if len(event.Transcript) > 0 {
				log.Printf("Transcript contains %d entries\n", len(event.Transcript))
				
				// Count message types
				userMessages := 0
				assistantMessages := 0
				toolUses := 0
				
				for _, entry := range event.Transcript {
					if entry.IsUserMessage() {
						userMessages++
					} else if entry.IsAssistantMessage() {
						assistantMessages++
						
						// Check for tool use
						if msg, err := entry.GetAssistantMessage(); err == nil && msg != nil {
							// Simple check - in real usage you'd parse the content array
							contentStr := string(msg.Content)
							if strings.Contains(contentStr, "tool_use") {
								toolUses++
							}
						}
					}
				}
				
				log.Printf("Summary: %d user messages, %d assistant messages, %d tool uses\n",
					userMessages, assistantMessages, toolUses)
				
				// Show last user message if available
				for i := len(event.Transcript) - 1; i >= 0; i-- {
					if event.Transcript[i].IsUserMessage() {
						if msg, err := event.Transcript[i].GetUserMessage(); err == nil && msg != nil {
							// Try to extract text content
							contentStr := string(msg.Content)
							if strings.HasPrefix(contentStr, "\"") && strings.HasSuffix(contentStr, "\"") {
								contentStr = contentStr[1:len(contentStr)-1]
							}
							log.Printf("Last user message: %s\n", contentStr)
						}
						break
					}
				}
			} else {
				log.Println("No transcript available")
			}
			
			// Allow stop on subsequent attempts, block on first
			if event.StopHookActive {
				return cchooks.Continue(), nil
			} else {
				return cchooks.BlockStop("Transcript analyzed. Stop again to confirm."), nil
			}
		},
		
		// Use StopOnce for first-stop specific logic
		StopOnce: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
			log.Println("First stop detected - performing initial analysis")
			
			// Could save transcript to file, send to analytics service, etc.
			if event.TranscriptPath != "" {
				log.Printf("Transcript available at: %s\n", event.TranscriptPath)
			}
			
			return cchooks.BlockStop("First stop blocked for analysis. Stop again to exit."), nil
		},
	}

	if err := runner.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}