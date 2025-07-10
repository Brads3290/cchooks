package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cchooks "github.com/brads3290/cchooks"
)

func TestSecurityHook(t *testing.T) {
	runner := createRunner()
	tester := cchooks.NewTestRunner(runner)

	// Test dangerous command is blocked
	dangerousInputs := []struct {
		name    string
		command string
	}{
		{"rm -rf root", "rm -rf /"},
		{"sudo rm", "sudo rm /important/file"},
		{"dd command", "dd if=/dev/zero of=/dev/sda"},
		{"fork bomb", ":(){ :|: & };:"},
	}

	for _, input := range dangerousInputs {
		t.Run(input.name, func(t *testing.T) {
			err := tester.AssertPreToolUseBlocks("Bash", &cchooks.BashInput{Command: input.command})
			if err != nil {
				t.Errorf("expected dangerous command to be blocked: %v", err)
			}
		})
	}

	// Test safe command is approved
	safeInputs := []struct {
		name    string
		command string
	}{
		{"ls", "ls -la"},
		{"echo", "echo 'Hello World'"},
		{"git status", "git status"},
		{"npm install", "npm install"},
	}

	for _, input := range safeInputs {
		t.Run(input.name, func(t *testing.T) {
			err := tester.AssertPreToolUseApproves("Bash", &cchooks.BashInput{Command: input.command})
			if err != nil {
				t.Errorf("expected safe command to be approved: %v", err)
			}
		})
	}

	// Test file path blocking
	blockedPaths := []struct {
		name     string
		filePath string
	}{
		{"production file", "/production/config.yaml"},
		{"etc file", "/etc/passwd"},
		{"usr file", "/usr/bin/bash"},
		{"boot file", "/boot/grub/grub.cfg"},
	}

	for _, path := range blockedPaths {
		t.Run("Edit "+path.name, func(t *testing.T) {
			err := tester.AssertPreToolUseBlocks("Edit", &cchooks.EditInput{
				FilePath:  path.filePath,
				OldString: "old",
				NewString: "new",
			})
			if err != nil {
				t.Errorf("expected edit to be blocked: %v", err)
			}
		})

		t.Run("Write "+path.name, func(t *testing.T) {
			err := tester.AssertPreToolUseBlocks("Write", &cchooks.WriteInput{
				FilePath: path.filePath,
				Content:  "content",
			})
			if err != nil {
				t.Errorf("expected write to be blocked: %v", err)
			}
		})
	}

	// Test allowed file paths
	allowedPaths := []string{
		"/home/user/project/main.go",
		"/tmp/test.txt",
		"/var/log/myapp.log",
	}

	for _, path := range allowedPaths {
		t.Run("Edit allowed "+path, func(t *testing.T) {
			err := tester.AssertPreToolUseApproves("Edit", &cchooks.EditInput{
				FilePath:  path,
				OldString: "old",
				NewString: "new",
			})
			if err != nil {
				t.Errorf("expected edit to be approved: %v", err)
			}
		})
	}
}

func createRunner() *cchooks.Runner {
	// Copy the actual logic from main.go
	return &cchooks.Runner{
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
	}
}
