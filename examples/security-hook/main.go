package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
			switch event.ToolName {
			case "Bash":
				bash, err := event.AsBash()
				if err != nil {
					return nil, err
				}

				// Block dangerous commands
				dangerous := []string{"rm -rf", "sudo rm", "dd if=", ":(){ :|: & };:"}
				for _, pattern := range dangerous {
					if strings.Contains(bash.Command, pattern) {
						return cchooks.Block(fmt.Sprintf("Dangerous command pattern detected: %s", pattern)), nil
					}
				}

				// Warn about sudo usage
				if strings.HasPrefix(bash.Command, "sudo") {
					log.Printf("WARNING: sudo command detected: %s", bash.Command)
				}

				return cchooks.Approve(), nil

			case "Edit", "Write":
				// Check file paths
				var filePath string
				if event.ToolName == "Edit" {
					edit, err := event.AsEdit()
					if err != nil {
						return nil, err
					}
					filePath = edit.FilePath
				} else {
					write, err := event.AsWrite()
					if err != nil {
						return nil, err
					}
					filePath = write.FilePath
				}

				// Block editing production files
				if strings.Contains(filePath, "/production/") {
					return cchooks.Block("Cannot edit production files"), nil
				}

				// Block editing system files
				systemPaths := []string{"/etc/", "/usr/", "/bin/", "/sbin/", "/boot/"}
				for _, systemPath := range systemPaths {
					if strings.HasPrefix(filePath, systemPath) {
						return cchooks.Block(fmt.Sprintf("Cannot edit system files in %s", systemPath)), nil
					}
				}

				return cchooks.Approve(), nil

			default:
				return cchooks.Approve(), nil
			}
		},

		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
			// Auto-format code after edits
			if event.ToolName == "Edit" || event.ToolName == "Write" {
				var filePath string
				if event.ToolName == "Edit" {
					edit, err := event.InputAsEdit()
					if err == nil {
						filePath = edit.FilePath
					}
				} else {
					write, err := event.InputAsWrite()
					if err == nil {
						filePath = write.FilePath
					}
				}

				if filePath != "" {
					switch {
					case strings.HasSuffix(filePath, ".go"):
						exec.Command("gofmt", "-w", filePath).Run()
					case strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".ts"):
						exec.Command("prettier", "--write", filePath).Run()
					case strings.HasSuffix(filePath, ".py"):
						exec.Command("black", filePath).Run()
					}
				}
			}

			return cchooks.Allow(), nil
		},

		Notification: func(ctx context.Context, event *cchooks.NotificationEvent) (*cchooks.NotificationResponse, error) {
			// Log notifications
			log.Printf("Claude notification: %s", event.Message)

			// Send desktop notification on macOS
			if strings.Contains(event.Message, "error") || strings.Contains(event.Message, "failed") {
				cmd := exec.Command("osascript", "-e",
					fmt.Sprintf(`display notification "%s" with title "Claude Code" sound name "Basso"`, event.Message))
				cmd.Run()
			}

			return cchooks.OK(), nil
		},

		Stop: func(ctx context.Context, event *cchooks.StopEvent) (*cchooks.StopResponse, error) {
			log.Printf("Claude session %s stopped", event.SessionID)
			return cchooks.Continue(), nil
		},
	}

	if err := runner.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
