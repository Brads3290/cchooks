package cchooks

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunner_Run(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		runner      *Runner
		wantOutput  string
		wantErrCode int
	}{
		{
			name:  "PreToolUse approve",
			input: `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`,
			runner: &Runner{
				PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
					return Approve(), nil
				},
			},
			wantOutput: `{
  "decision": "approve"
}
`,
		},
		{
			name:  "PreToolUse block",
			input: `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "rm -rf /"}}`,
			runner: &Runner{
				PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
					return Block("dangerous command"), nil
				},
			},
			wantOutput: `{
  "decision": "block",
  "reason": "dangerous command"
}
`,
		},
		{
			name:  "PreToolUse empty response",
			input: `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`,
			runner: &Runner{
				PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
					return &PreToolUseResponse{}, nil
				},
			},
			wantOutput: "",
		},
		{
			name:  "PostToolUse allow",
			input: `{"event": "PostToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}, "tool_response": {"output": "file1\nfile2"}}`,
			runner: &Runner{
				PostToolUse: func(ctx context.Context, event *PostToolUseEvent) (*PostToolUseResponse, error) {
					return Allow(), nil
				},
			},
			wantOutput: "",
		},
		{
			name:  "Notification OK",
			input: `{"event": "Notification", "session_id": "test", "notification_message": "Task completed"}`,
			runner: &Runner{
				Notification: func(ctx context.Context, event *NotificationEvent) (*NotificationResponse, error) {
					return OK(), nil
				},
			},
			wantOutput: "",
		},
		{
			name:  "Stop continue",
			input: `{"event": "Stop", "session_id": "test", "stop_hook_active": true, "transcript": []}`,
			runner: &Runner{
				Stop: func(ctx context.Context, event *StopEvent) (*StopResponse, error) {
					return Continue(), nil
				},
			},
			wantOutput: "",
		},
		{
			name:        "unknown event type",
			input:       `{"event": "Unknown", "session_id": "test"}`,
			runner:      &Runner{},
			wantErrCode: 1,
		},
		{
			name:        "missing event field",
			input:       `{"session_id": "test"}`,
			runner:      &Runner{},
			wantErrCode: 1,
		},
		{
			name:  "handler returns error",
			input: `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`,
			runner: &Runner{
				PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
					return nil, errors.New("handler error")
				},
			},
			wantErrCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			w.Write([]byte(tt.input))
			w.Close()
			defer func() { os.Stdin = oldStdin }()

			// Set up stdout
			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut
			defer func() { os.Stdout = oldStdout }()

			// Set up stderr
			oldStderr := os.Stderr
			rErr, wErr, _ := os.Pipe()
			os.Stderr = wErr
			defer func() { os.Stderr = oldStderr }()

			// Capture exit code
			exitCode := 0
			oldExit := osExit
			osExit = func(code int) {
				exitCode = code
				panic("exit")
			}
			defer func() { osExit = oldExit }()

			// Run the test
			func() {
				defer func() {
					if r := recover(); r != nil && r != "exit" {
						panic(r)
					}
				}()
				err := tt.runner.Run(context.Background())
				if err != nil {
					exitCode = 1
				}
			}()

			// Close write ends
			wOut.Close()
			wErr.Close()

			// Read output
			output, _ := io.ReadAll(rOut)
			errOutput, _ := io.ReadAll(rErr)

			// Check exit code
			if exitCode != tt.wantErrCode {
				t.Errorf("exit code = %d, want %d, stderr = %s", exitCode, tt.wantErrCode, errOutput)
			}

			// Check output
			if tt.wantErrCode == 0 && string(output) != tt.wantOutput {
				t.Errorf("output = %q, want %q", string(output), tt.wantOutput)
			}
		})
	}
}

func TestEventParsing(t *testing.T) {
	t.Run("PreToolUseEvent parsing", func(t *testing.T) {
		event := &PreToolUseEvent{
			SessionID: "test",
			ToolName:  "Bash",
			ToolInput: json.RawMessage(`{"command": "ls", "timeout": 5000}`),
		}

		bash, err := event.AsBash()
		if err != nil {
			t.Fatalf("AsBash() error = %v", err)
		}
		if bash.Command != "ls" {
			t.Errorf("Command = %q, want %q", bash.Command, "ls")
		}
		if bash.Timeout == nil || *bash.Timeout != 5000 {
			t.Errorf("Timeout = %v, want 5000", bash.Timeout)
		}
	})

	t.Run("PostToolUseEvent input parsing", func(t *testing.T) {
		event := &PostToolUseEvent{
			SessionID:    "test",
			ToolName:     "Edit",
			ToolInput:    json.RawMessage(`{"file_path": "/test.txt", "old_string": "old", "new_string": "new"}`),
			ToolResponse: json.RawMessage(`{"success": true}`),
		}

		edit, err := event.InputAsEdit()
		if err != nil {
			t.Fatalf("InputAsEdit() error = %v", err)
		}
		if edit.FilePath != "/test.txt" {
			t.Errorf("FilePath = %q, want %q", edit.FilePath, "/test.txt")
		}

		editResp, err := event.ResponseAsEdit()
		if err != nil {
			t.Fatalf("ResponseAsEdit() error = %v", err)
		}
		if !editResp.Success {
			t.Errorf("Success = %v, want true", editResp.Success)
		}
	})
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		response interface{}
		want     bool
	}{
		{
			name:     "empty PreToolUseResponse",
			response: &PreToolUseResponse{},
			want:     true,
		},
		{
			name:     "non-empty PreToolUseResponse with decision",
			response: &PreToolUseResponse{Decision: "approve"},
			want:     false,
		},
		{
			name:     "non-empty PreToolUseResponse with continue",
			response: &PreToolUseResponse{Continue: func() *bool { b := false; return &b }()},
			want:     false,
		},
		{
			name:     "empty PostToolUseResponse",
			response: &PostToolUseResponse{},
			want:     true,
		},
		{
			name:     "non-empty PostToolUseResponse",
			response: &PostToolUseResponse{Decision: "block"},
			want:     false,
		},
		{
			name:     "empty NotificationResponse",
			response: &NotificationResponse{},
			want:     true,
		},
		{
			name:     "non-empty NotificationResponse",
			response: &NotificationResponse{StopReason: "done"},
			want:     false,
		},
		{
			name:     "empty StopResponse",
			response: &StopResponse{},
			want:     true,
		},
		{
			name:     "non-empty StopResponse",
			response: &StopResponse{Decision: "block"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmpty(tt.response); got != tt.want {
				t.Errorf("isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputResponse(t *testing.T) {
	tests := []struct {
		name       string
		response   interface{}
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "empty response",
			response:   &PreToolUseResponse{},
			wantOutput: "",
		},
		{
			name:       "non-empty response",
			response:   &PreToolUseResponse{Decision: "approve"},
			wantOutput: "{\n  \"decision\": \"approve\"\n}\n",
		},
		{
			name:       "response with all fields",
			response:   &PreToolUseResponse{Decision: "block", Reason: "test", StopReason: "stop", Continue: func() *bool { b := false; return &b }()},
			wantOutput: "{\n  \"decision\": \"block\",\n  \"continue\": false,\n  \"stopReason\": \"stop\",\n  \"reason\": \"test\"\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := outputResponse(tt.response)

			w.Close()
			os.Stdout = oldStdout

			output, _ := io.ReadAll(r)

			if (err != nil) != tt.wantErr {
				t.Errorf("outputResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if string(output) != tt.wantOutput {
				t.Errorf("output = %q, want %q", string(output), tt.wantOutput)
			}
		})
	}
}

func TestHandlerErrors(t *testing.T) {
	runner := &Runner{
		PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
			return nil, errors.New("test error")
		},
	}

	// Mock stdin
	input := `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte(input))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Mock stderr
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr
	defer func() { os.Stderr = oldStderr }()

	// Mock os.Exit
	exitCode := -1
	oldExit := osExit
	osExit = func(code int) {
		exitCode = code
		panic("exit")
	}
	defer func() { osExit = oldExit }()

	// Run
	func() {
		defer func() {
			if r := recover(); r != nil && r != "exit" {
				panic(r)
			}
		}()
		runner.Run(context.Background())
	}()

	// Close write end and read stderr
	wErr.Close()
	stderrOutput, _ := io.ReadAll(rErr)

	if exitCode != 2 {
		t.Errorf("expected exit code 2, got %d, stderr: %s", exitCode, stderrOutput)
	}
}

func TestRawHandler(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		runner      *Runner
		wantOutput  string
		wantErrCode int
	}{
		{
			name:  "Raw handler returns response with output",
			input: `{"event": "PreToolUse", "session_id": "test"}`,
			runner: &Runner{
				Raw: func(ctx context.Context, rawJSON string) (*RawResponse, error) {
					return &RawResponse{ExitCode: 3, Output: "custom output"}, nil
				},
			},
			wantOutput:  "custom output",
			wantErrCode: 3,
		},
		{
			name:  "Raw handler returns response without output",
			input: `{"event": "PreToolUse", "session_id": "test"}`,
			runner: &Runner{
				Raw: func(ctx context.Context, rawJSON string) (*RawResponse, error) {
					return &RawResponse{ExitCode: 5}, nil
				},
			},
			wantOutput:  "",
			wantErrCode: 5,
		},
		{
			name:  "Raw handler returns nil, continues to normal processing",
			input: `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`,
			runner: &Runner{
				Raw: func(ctx context.Context, rawJSON string) (*RawResponse, error) {
					// Return nil to continue normal processing
					return nil, nil
				},
				PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
					return Approve(), nil
				},
			},
			wantOutput: `{
  "decision": "approve"
}
`,
			wantErrCode: 0,
		},
		{
			name:  "Raw handler returns error",
			input: `{"event": "PreToolUse", "session_id": "test"}`,
			runner: &Runner{
				Raw: func(ctx context.Context, rawJSON string) (*RawResponse, error) {
					return nil, errors.New("raw handler error")
				},
			},
			wantErrCode: 2,
		},
		{
			name:  "Raw handler with Error handler",
			input: `{"event": "PreToolUse", "session_id": "test"}`,
			runner: &Runner{
				Raw: func(ctx context.Context, rawJSON string) (*RawResponse, error) {
					return nil, errors.New("raw handler error")
				},
				Error: func(ctx context.Context, rawJSON string, err error) {
					if rawJSON != `{"event": "PreToolUse", "session_id": "test"}` {
						t.Errorf("Error handler got rawJSON = %q", rawJSON)
					}
					if err.Error() != "raw handler error" {
						t.Errorf("Error handler got unexpected error: %v", err)
					}
				},
			},
			wantErrCode: 2,
		},
		{
			name:  "Raw handler with malformed JSON",
			input: `{malformed json`,
			runner: &Runner{
				Raw: func(ctx context.Context, rawJSON string) (*RawResponse, error) {
					// Can still process raw JSON even if it's malformed
					if rawJSON == `{malformed json` {
						return &RawResponse{ExitCode: 7, Output: "handled malformed"}, nil
					}
					return nil, nil
				},
			},
			wantOutput:  "handled malformed",
			wantErrCode: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			w.Write([]byte(tt.input))
			w.Close()
			defer func() { os.Stdin = oldStdin }()

			// Set up stdout
			oldStdout := os.Stdout
			rOut, wOut, _ := os.Pipe()
			os.Stdout = wOut
			defer func() { os.Stdout = oldStdout }()

			// Set up stderr
			oldStderr := os.Stderr
			rErr, wErr, _ := os.Pipe()
			os.Stderr = wErr
			defer func() { os.Stderr = oldStderr }()

			// Capture exit code
			exitCode := 0
			oldExit := osExit
			osExit = func(code int) {
				exitCode = code
				panic("exit")
			}
			defer func() { osExit = oldExit }()

			// Run the test
			func() {
				defer func() {
					if r := recover(); r != nil && r != "exit" {
						panic(r)
					}
				}()
				err := tt.runner.Run(context.Background())
				if err != nil {
					exitCode = 1
				}
			}()

			// Close write ends
			wOut.Close()
			wErr.Close()

			// Read output
			output, _ := io.ReadAll(rOut)
			errOutput, _ := io.ReadAll(rErr)

			// Check exit code
			if exitCode != tt.wantErrCode {
				t.Errorf("exit code = %d, want %d, stderr = %s", exitCode, tt.wantErrCode, errOutput)
			}

			// Check output
			if string(output) != tt.wantOutput {
				t.Errorf("output = %q, want %q", string(output), tt.wantOutput)
			}
		})
	}
}

func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		runner        *Runner
		wantErrJSON   string
		wantErrString string
	}{
		{
			name:  "invalid JSON",
			input: `{invalid json`,
			runner: &Runner{
				Error: func(ctx context.Context, rawJSON string, err error) {
					if rawJSON != `{invalid json` {
						t.Errorf("Error handler got rawJSON = %q, want %q", rawJSON, `{invalid json`)
					}
					if err == nil || !strings.Contains(err.Error(), "failed to decode stdin:") {
						t.Errorf("Error handler got unexpected error: %v", err)
					}
				},
			},
			wantErrJSON:   `{invalid json`,
			wantErrString: "failed to decode stdin:",
		},
		{
			name:  "missing event field",
			input: `{"session_id": "test"}`,
			runner: &Runner{
				Error: func(ctx context.Context, rawJSON string, err error) {
					if rawJSON != `{"session_id": "test"}` {
						t.Errorf("Error handler got rawJSON = %q, want %q", rawJSON, `{"session_id": "test"}`)
					}
					if err == nil || err.Error() != "missing or invalid event field" {
						t.Errorf("Error handler got unexpected error: %v", err)
					}
				},
			},
			wantErrJSON:   `{"session_id": "test"}`,
			wantErrString: "missing or invalid event field",
		},
		{
			name:  "handler error",
			input: `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`,
			runner: &Runner{
				PreToolUse: func(ctx context.Context, event *PreToolUseEvent) (*PreToolUseResponse, error) {
					return nil, errors.New("handler error")
				},
				Error: func(ctx context.Context, rawJSON string, err error) {
					expectedJSON := `{"event": "PreToolUse", "session_id": "test", "tool_name": "Bash", "tool_input": {"command": "ls"}}`
					var expected, actual map[string]interface{}
					json.Unmarshal([]byte(expectedJSON), &expected)
					json.Unmarshal([]byte(rawJSON), &actual)
					
					expectedBytes, _ := json.Marshal(expected)
					actualBytes, _ := json.Marshal(actual)
					
					if string(expectedBytes) != string(actualBytes) {
						t.Errorf("Error handler got rawJSON = %q, want %q", rawJSON, expectedJSON)
					}
					if err == nil || err.Error() != "handler error" {
						t.Errorf("Error handler got unexpected error: %v", err)
					}
				},
			},
			wantErrString: "handler error",
		},
		{
			name:  "unknown event type",
			input: `{"event": "UnknownEvent", "session_id": "test"}`,
			runner: &Runner{
				Error: func(ctx context.Context, rawJSON string, err error) {
					expectedJSON := `{"event": "UnknownEvent", "session_id": "test"}`
					var expected, actual map[string]interface{}
					json.Unmarshal([]byte(expectedJSON), &expected)
					json.Unmarshal([]byte(rawJSON), &actual)
					
					if err == nil || err.Error() != "unknown event type: UnknownEvent" {
						t.Errorf("Error handler got unexpected error: %v", err)
					}
				},
			},
			wantErrString: "unknown event type: UnknownEvent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock stdin
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			w.Write([]byte(tt.input))
			w.Close()
			defer func() { os.Stdin = oldStdin }()

			// Mock stderr for handler errors
			if tt.runner.PreToolUse != nil || tt.runner.PostToolUse != nil {
				oldStderr := os.Stderr
				_, wErr, _ := os.Pipe()
				os.Stderr = wErr
				defer func() {
					wErr.Close()
					os.Stderr = oldStderr
				}()

				// Mock os.Exit for handler errors
				oldExit := osExit
				osExit = func(code int) {
					panic("exit")
				}
				defer func() { osExit = oldExit }()
			}

			// Run and handle expected errors/panics
			func() {
				defer func() {
					if r := recover(); r != nil && r != "exit" {
						panic(r)
					}
				}()
				err := tt.runner.Run(context.Background())
				if err != nil && !strings.Contains(err.Error(), tt.wantErrString) {
					t.Errorf("Run() error = %v, want %v", err, tt.wantErrString)
				}
			}()
		})
	}
}
