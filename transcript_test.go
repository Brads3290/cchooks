package cchooks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTranscriptReading(t *testing.T) {
	// Create a temporary transcript file
	tmpDir, err := os.MkdirTemp("", "transcript-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	transcriptPath := filepath.Join(tmpDir, "test-transcript.jsonl")

	// Write test transcript data
	transcriptData := []string{
		`{"parentUuid":null,"uuid":"1","isSidechain":false,"userType":"external","cwd":"/test","sessionId":"test-session","version":"1.0.0","type":"user","message":{"role":"user","content":"Hello"},"timestamp":"2025-01-10T10:00:00Z"}`,
		`{"parentUuid":"1","uuid":"2","isSidechain":false,"userType":"external","cwd":"/test","sessionId":"test-session","version":"1.0.0","type":"assistant","message":{"id":"msg-1","type":"message","role":"assistant","model":"claude-3","content":[{"type":"text","text":"Hi there!"}],"stop_reason":"stop_sequence","stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"service_tier":"standard"}},"timestamp":"2025-01-10T10:00:01Z"}`,
		``, // Empty line should be skipped
		`{"parentUuid":"2","uuid":"3","isSidechain":false,"userType":"external","cwd":"/test","sessionId":"test-session","version":"1.0.0","type":"user","message":{"role":"user","content":"Goodbye"},"timestamp":"2025-01-10T10:00:02Z"}`,
	}

	file, err := os.Create(transcriptPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range transcriptData {
		file.WriteString(line + "\n")
	}
	file.Close()

	// Test readTranscript function
	entries, err := readTranscript(transcriptPath)
	if err != nil {
		t.Fatalf("readTranscript failed: %v", err)
	}

	// Should have 3 entries (empty line skipped)
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Verify first entry
	if entries[0].UUID != "1" {
		t.Errorf("First entry UUID = %s, want 1", entries[0].UUID)
	}
	if entries[0].Type != "user" {
		t.Errorf("First entry Type = %s, want user", entries[0].Type)
	}

	// Test parsing user message
	userMsg, err := entries[0].GetUserMessage()
	if err != nil {
		t.Fatalf("GetUserMessage failed: %v", err)
	}
	if userMsg == nil {
		t.Fatal("Expected user message, got nil")
	}
	if userMsg.Role != "user" {
		t.Errorf("User message role = %s, want user", userMsg.Role)
	}

	// Test parsing assistant message
	assistantMsg, err := entries[1].GetAssistantMessage()
	if err != nil {
		t.Fatalf("GetAssistantMessage failed: %v", err)
	}
	if assistantMsg == nil {
		t.Fatal("Expected assistant message, got nil")
	}
	if assistantMsg.Role != "assistant" {
		t.Errorf("Assistant message role = %s, want assistant", assistantMsg.Role)
	}
	if assistantMsg.Model != "claude-3" {
		t.Errorf("Assistant message model = %s, want claude-3", assistantMsg.Model)
	}
}

func TestStopEventWithTranscript(t *testing.T) {
	// Create a temporary transcript file
	tmpDir, err := os.MkdirTemp("", "stop-transcript-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	transcriptPath := filepath.Join(tmpDir, "stop-transcript.jsonl")

	// Write minimal transcript
	file, err := os.Create(transcriptPath)
	if err != nil {
		t.Fatal(err)
	}
	file.WriteString(`{"parentUuid":null,"uuid":"1","isSidechain":false,"userType":"external","cwd":"/test","sessionId":"test-session","version":"1.0.0","type":"user","message":{"role":"user","content":"Test message"},"timestamp":"2025-01-10T10:00:00Z"}` + "\n")
	file.Close()

	// Create runner that verifies transcript is loaded
	runner := &Runner{
		Stop: func(ctx context.Context, event *StopEvent) StopResponseInterface {
			// Verify transcript was loaded
			if len(event.Transcript) != 1 {
				t.Errorf("Expected 1 transcript entry, got %d", len(event.Transcript))
			}
			if event.TranscriptPath != transcriptPath {
				t.Errorf("TranscriptPath = %s, want %s", event.TranscriptPath, transcriptPath)
			}
			return Continue()
		},
	}

	// Test with transcript path
	input := `{"hook_event_name": "Stop", "session_id": "test", "stop_hook_active": true, "transcript_path": "` + transcriptPath + `"}`

	// Mock stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte(input))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Mock stdout
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	defer func() {
		wOut.Close()
		os.Stdout = oldStdout
	}()

	// Mock exit
	exitCode := -1
	runner.ExitFn = func(code int) {
		exitCode = code
		panic("exit")
	}

	// Run
	func() {
		defer func() {
			if r := recover(); r != nil && r != "exit" {
				panic(r)
			}
		}()
		runner.Run()
	}()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestStopEventWithMissingTranscript(t *testing.T) {
	// Test that missing transcript file doesn't cause failure
	runner := &Runner{
		Stop: func(ctx context.Context, event *StopEvent) StopResponseInterface {
			// Verify transcript is empty array, not nil
			if event.Transcript == nil {
				t.Error("Transcript should not be nil")
			}
			if len(event.Transcript) != 0 {
				t.Errorf("Expected empty transcript, got %d entries", len(event.Transcript))
			}
			return Continue()
		},
	}

	// Test with non-existent transcript path
	input := `{"hook_event_name": "Stop", "session_id": "test", "stop_hook_active": true, "transcript_path": "/non/existent/path.jsonl"}`

	// Mock stdin
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte(input))
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	// Mock stdout
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	defer func() {
		wOut.Close()
		os.Stdout = oldStdout
	}()

	// Mock exit
	exitCode := -1
	runner.ExitFn = func(code int) {
		exitCode = code
		panic("exit")
	}

	// Run
	func() {
		defer func() {
			if r := recover(); r != nil && r != "exit" {
				panic(r)
			}
		}()
		runner.Run()
	}()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestTranscriptEntryMethods(t *testing.T) {
	// Test IsUserMessage and IsAssistantMessage
	userEntry := TranscriptEntry{
		Type:    "user",
		Message: json.RawMessage(`{"role":"user","content":"Hello"}`),
	}

	if !userEntry.IsUserMessage() {
		t.Error("Expected IsUserMessage to return true")
	}
	if userEntry.IsAssistantMessage() {
		t.Error("Expected IsAssistantMessage to return false")
	}

	assistantEntry := TranscriptEntry{
		Type:    "assistant",
		Message: json.RawMessage(`{"role":"assistant","content":[{"type":"text","text":"Hi"}]}`),
	}

	if assistantEntry.IsUserMessage() {
		t.Error("Expected IsUserMessage to return false")
	}
	if !assistantEntry.IsAssistantMessage() {
		t.Error("Expected IsAssistantMessage to return true")
	}

	// Test GetUserMessage returns nil for assistant entry
	userMsg, err := assistantEntry.GetUserMessage()
	if err != nil {
		t.Errorf("GetUserMessage error: %v", err)
	}
	if userMsg != nil {
		t.Error("Expected GetUserMessage to return nil for assistant entry")
	}

	// Test GetAssistantMessage returns nil for user entry
	assistantMsg, err := userEntry.GetAssistantMessage()
	if err != nil {
		t.Errorf("GetAssistantMessage error: %v", err)
	}
	if assistantMsg != nil {
		t.Error("Expected GetAssistantMessage to return nil for user entry")
	}
}
