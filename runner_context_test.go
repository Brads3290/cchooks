package cchooks

import (
	"context"
	"io"
	"os"
	"testing"
)

func TestRunContext(t *testing.T) {
	// Test that RunContext works with a custom context
	runner := &Runner{
		PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
			// Verify context is passed through
			if ctx == nil {
				t.Error("context should not be nil")
			}
			return Approve(), nil
		},
	}

	// Mock stdin
	input := `{"hook_event_name": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte(input))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Mock stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	defer func() {
		os.Stdout = oldStdout
	}()

	// Mock os.Exit
	exitCode := 0 // Default to success
	runner.ExitFn = func(code int) {
		exitCode = code
		panic("exit")
	}

	// Run with custom context
	ctx := context.WithValue(context.Background(), "test", "value")
	func() {
		defer func() {
			if r := recover(); r != nil && r != "exit" {
				panic(r)
			}
		}()
		runner.RunContext(ctx)
	}()

	// Close and read output
	wOut.Close()
	output, _ := io.ReadAll(rOut)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	expectedOutput := `{
  "decision": "approve"
}
`
	if string(output) != expectedOutput {
		t.Errorf("output = %q, want %q", string(output), expectedOutput)
	}
}
