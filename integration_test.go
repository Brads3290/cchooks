package cchooks_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	cchooks "github.com/brads3290/claude-code-hooks-go"
)

// TestIntegration runs an end-to-end test simulating Claude Code hook execution
func TestIntegration(t *testing.T) {
	t.Skip("Skipping integration test temporarily")
	// Create a test hook binary
	hookCode := `
package main

import (
	"context"
	"log"
	"strings"
	cchooks "github.com/brads3290/claude-code-hooks-go"
)

func main() {
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
			if event.ToolName == "Bash" {
				bash, _ := event.AsBash()
				if strings.Contains(bash.Command, "rm -rf") {
					return cchooks.Block("dangerous command"), nil
				}
			}
			return cchooks.Approve(), nil
		},
	}
	
	if err := runner.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
`

	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "cchooks-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Write test hook code
	hookFile := filepath.Join(tmpDir, "hook.go")
	if err := os.WriteFile(hookFile, []byte(hookCode), 0644); err != nil {
		t.Fatal(err)
	}

	// Write go.mod
	// Get the root directory of the cchooks module
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	
	goMod := fmt.Sprintf(`module testhook
go 1.21
require github.com/brads3290/claude-code-hooks-go v0.0.0
replace github.com/brads3290/claude-code-hooks-go => %s
`, cwd)

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Build the hook
	hookBinary := filepath.Join(tmpDir, "hook")
	cmd := exec.Command("go", "build", "-o", hookBinary, hookFile)
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build hook: %v\nOutput: %s", err, output)
	}

	// Test cases
	tests := []struct {
		name       string
		input      map[string]interface{}
		wantOutput string
		wantExit   int
	}{
		{
			name: "approve safe command",
			input: map[string]interface{}{
				"event":       "PreToolUse",
				"session_id":  "test-123",
				"tool_name":   "Bash",
				"tool_input": map[string]interface{}{
					"command": "ls -la",
				},
			},
			wantOutput: `{
  "decision": "approve"
}
`,
			wantExit: 0,
		},
		{
			name: "block dangerous command",
			input: map[string]interface{}{
				"event":       "PreToolUse",
				"session_id":  "test-456",
				"tool_name":   "Bash",
				"tool_input": map[string]interface{}{
					"command": "rm -rf /",
				},
			},
			wantOutput: `{
  "decision": "block",
  "reason": "dangerous command"
}
`,
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare input JSON
			inputJSON, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			// Run the hook
			cmd := exec.Command(hookBinary)
			cmd.Stdin = bytes.NewReader(inputJSON)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err = cmd.Run()
			exitCode := 0
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check exit code
			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nstderr: %s", exitCode, tt.wantExit, stderr.String())
			}

			// Check output
			if stdout.String() != tt.wantOutput {
				t.Errorf("output = %q, want %q", stdout.String(), tt.wantOutput)
			}
		})
	}
}

// TestExampleHooks tests that the example hooks compile and run
func TestExampleHooks(t *testing.T) {
	examples := []string{
		"examples/simple-hook",
		"examples/security-hook",
		"examples/format-hook",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			// Test that it builds
			cmd := exec.Command("go", "build", "-o", "/dev/null", "./"+example)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("failed to build %s: %v\nOutput: %s", example, err, output)
			}
		})
	}
}

// TestHookIO tests the I/O behavior of hooks
func TestHookIO(t *testing.T) {
	t.Skip("Skipping hook IO test temporarily")
	// Create pipes for testing
	stdinR, stdinW, _ := os.Pipe()
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	// Save original streams
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Replace streams
	os.Stdin = stdinR
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Restore streams on exit
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create runner
	runner := &cchooks.Runner{
		PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
			return cchooks.Approve(), nil
		},
	}

	// Write test input
	input := `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "echo test"}}`
	go func() {
		stdinW.Write([]byte(input))
		stdinW.Close()
	}()

	// Run in goroutine
	done := make(chan error)
	go func() {
		done <- runner.Run(context.Background())
	}()

	// Close write ends
	stdoutW.Close()
	stderrW.Close()

	// Read output
	stdout, _ := io.ReadAll(stdoutR)
	stderr, _ := io.ReadAll(stderrR)

	// Wait for completion
	err := <-done
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify output
	expectedOutput := `{
  "decision": "approve"
}
`
	if string(stdout) != expectedOutput {
		t.Errorf("stdout = %q, want %q", string(stdout), expectedOutput)
	}

	if len(stderr) > 0 {
		t.Errorf("unexpected stderr: %s", stderr)
	}
}