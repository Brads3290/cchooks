package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/brads3290/claude-code-hooks-go"
	"github.com/brads3290/claude-code-hooks-go/internal/tools"
)

func TestPreToolUseEventParsing(t *testing.T) {
	tests := []struct {
		name      string
		toolInput string
		parser    func(*cchooks.PreToolUseEvent) (interface{}, error)
		validate  func(t *testing.T, result interface{})
	}{
		{
			name:      "AsBash",
			toolInput: `{"command": "ls -la", "timeout": 5000, "description": "List files"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsBash()
			},
			validate: func(t *testing.T, result interface{}) {
				bash := result.(*tools.BashInput)
				if bash.Command != "ls -la" {
					t.Errorf("Command = %q, want %q", bash.Command, "ls -la")
				}
				if bash.Timeout == nil || *bash.Timeout != 5000 {
					t.Errorf("Timeout = %v, want 5000", bash.Timeout)
				}
				if bash.Description != "List files" {
					t.Errorf("Description = %q, want %q", bash.Description, "List files")
				}
			},
		},
		{
			name:      "AsEdit",
			toolInput: `{"file_path": "/test.txt", "old_string": "old", "new_string": "new", "replace_all": true}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsEdit()
			},
			validate: func(t *testing.T, result interface{}) {
				edit := result.(*tools.EditInput)
				if edit.FilePath != "/test.txt" {
					t.Errorf("FilePath = %q, want %q", edit.FilePath, "/test.txt")
				}
				if edit.OldString != "old" {
					t.Errorf("OldString = %q, want %q", edit.OldString, "old")
				}
				if edit.NewString != "new" {
					t.Errorf("NewString = %q, want %q", edit.NewString, "new")
				}
				if !edit.ReplaceAll {
					t.Error("ReplaceAll = false, want true")
				}
			},
		},
		{
			name:      "AsMultiEdit",
			toolInput: `{"file_path": "/test.txt", "edits": [{"old_string": "a", "new_string": "b", "replace_all": true}]}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsMultiEdit()
			},
			validate: func(t *testing.T, result interface{}) {
				multi := result.(*tools.MultiEditInput)
				if multi.FilePath != "/test.txt" {
					t.Errorf("FilePath = %q, want %q", multi.FilePath, "/test.txt")
				}
				if len(multi.Edits) != 1 {
					t.Fatalf("len(Edits) = %d, want 1", len(multi.Edits))
				}
				if multi.Edits[0].OldString != "a" {
					t.Errorf("Edits[0].OldString = %q, want %q", multi.Edits[0].OldString, "a")
				}
			},
		},
		{
			name:      "AsWrite",
			toolInput: `{"file_path": "/new.txt", "content": "Hello World"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsWrite()
			},
			validate: func(t *testing.T, result interface{}) {
				write := result.(*tools.WriteInput)
				if write.FilePath != "/new.txt" {
					t.Errorf("FilePath = %q, want %q", write.FilePath, "/new.txt")
				}
				if write.Content != "Hello World" {
					t.Errorf("Content = %q, want %q", write.Content, "Hello World")
				}
			},
		},
		{
			name:      "AsRead",
			toolInput: `{"file_path": "/read.txt", "limit": 100, "offset": 50}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsRead()
			},
			validate: func(t *testing.T, result interface{}) {
				read := result.(*tools.ReadInput)
				if read.FilePath != "/read.txt" {
					t.Errorf("FilePath = %q, want %q", read.FilePath, "/read.txt")
				}
				if read.Limit == nil || *read.Limit != 100 {
					t.Errorf("Limit = %v, want 100", read.Limit)
				}
				if read.Offset == nil || *read.Offset != 50 {
					t.Errorf("Offset = %v, want 50", read.Offset)
				}
			},
		},
		{
			name:      "AsGlob",
			toolInput: `{"pattern": "*.go", "path": "/src"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsGlob()
			},
			validate: func(t *testing.T, result interface{}) {
				glob := result.(*tools.GlobInput)
				if glob.Pattern != "*.go" {
					t.Errorf("Pattern = %q, want %q", glob.Pattern, "*.go")
				}
				if glob.Path != "/src" {
					t.Errorf("Path = %q, want %q", glob.Path, "/src")
				}
			},
		},
		{
			name:      "AsGrep",
			toolInput: `{"pattern": "TODO", "path": "/src", "include": "*.go"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsGrep()
			},
			validate: func(t *testing.T, result interface{}) {
				grep := result.(*tools.GrepInput)
				if grep.Pattern != "TODO" {
					t.Errorf("Pattern = %q, want %q", grep.Pattern, "TODO")
				}
				if grep.Path != "/src" {
					t.Errorf("Path = %q, want %q", grep.Path, "/src")
				}
				if grep.Include != "*.go" {
					t.Errorf("Include = %q, want %q", grep.Include, "*.go")
				}
			},
		},
		{
			name:      "AsLS",
			toolInput: `{"path": "/home", "ignore": [".git", "node_modules"]}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsLS()
			},
			validate: func(t *testing.T, result interface{}) {
				ls := result.(*tools.LSInput)
				if ls.Path != "/home" {
					t.Errorf("Path = %q, want %q", ls.Path, "/home")
				}
				if len(ls.Ignore) != 2 {
					t.Fatalf("len(Ignore) = %d, want 2", len(ls.Ignore))
				}
				if ls.Ignore[0] != ".git" || ls.Ignore[1] != "node_modules" {
					t.Errorf("Ignore = %v, want [.git node_modules]", ls.Ignore)
				}
			},
		},
		{
			name:      "AsTodoWrite",
			toolInput: `{"todos": [{"id": "1", "content": "Test", "status": "pending", "priority": "high"}]}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsTodoWrite()
			},
			validate: func(t *testing.T, result interface{}) {
				todo := result.(*tools.TodoWriteInput)
				if len(todo.Todos) != 1 {
					t.Fatalf("len(Todos) = %d, want 1", len(todo.Todos))
				}
				if todo.Todos[0].ID != "1" {
					t.Errorf("ID = %q, want %q", todo.Todos[0].ID, "1")
				}
				if todo.Todos[0].Content != "Test" {
					t.Errorf("Content = %q, want %q", todo.Todos[0].Content, "Test")
				}
				if todo.Todos[0].Status != tools.TodoStatusPending {
					t.Errorf("Status = %q, want %q", todo.Todos[0].Status, tools.TodoStatusPending)
				}
				if todo.Todos[0].Priority != tools.TodoPriorityHigh {
					t.Errorf("Priority = %q, want %q", todo.Todos[0].Priority, tools.TodoPriorityHigh)
				}
			},
		},
		{
			name:      "AsTodoRead",
			toolInput: `{}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsTodoRead()
			},
			validate: func(t *testing.T, result interface{}) {
				// TodoReadInput has no fields
				_ = result.(*tools.TodoReadInput)
			},
		},
		{
			name:      "AsNotebookRead",
			toolInput: `{"notebook_path": "/nb.ipynb", "cell_id": "cell123"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsNotebookRead()
			},
			validate: func(t *testing.T, result interface{}) {
				nb := result.(*tools.NotebookReadInput)
				if nb.NotebookPath != "/nb.ipynb" {
					t.Errorf("NotebookPath = %q, want %q", nb.NotebookPath, "/nb.ipynb")
				}
				if nb.CellID != "cell123" {
					t.Errorf("CellID = %q, want %q", nb.CellID, "cell123")
				}
			},
		},
		{
			name:      "AsNotebookEdit",
			toolInput: `{"notebook_path": "/nb.ipynb", "cell_id": "cell1", "cell_type": "code", "edit_mode": "replace", "new_source": "print('hi')"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsNotebookEdit()
			},
			validate: func(t *testing.T, result interface{}) {
				nb := result.(*tools.NotebookEditInput)
				if nb.NotebookPath != "/nb.ipynb" {
					t.Errorf("NotebookPath = %q, want %q", nb.NotebookPath, "/nb.ipynb")
				}
				if nb.CellID != "cell1" {
					t.Errorf("CellID = %q, want %q", nb.CellID, "cell1")
				}
				if nb.CellType != "code" {
					t.Errorf("CellType = %q, want %q", nb.CellType, "code")
				}
				if nb.EditMode != "replace" {
					t.Errorf("EditMode = %q, want %q", nb.EditMode, "replace")
				}
				if nb.NewSource != "print('hi')" {
					t.Errorf("NewSource = %q, want %q", nb.NewSource, "print('hi')")
				}
			},
		},
		{
			name:      "AsWebFetch",
			toolInput: `{"url": "https://example.com", "prompt": "Get main content"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsWebFetch()
			},
			validate: func(t *testing.T, result interface{}) {
				web := result.(*tools.WebFetchInput)
				if web.URL != "https://example.com" {
					t.Errorf("URL = %q, want %q", web.URL, "https://example.com")
				}
				if web.Prompt != "Get main content" {
					t.Errorf("Prompt = %q, want %q", web.Prompt, "Get main content")
				}
			},
		},
		{
			name:      "AsWebSearch",
			toolInput: `{"query": "golang hooks", "allowed_domains": ["go.dev"], "blocked_domains": ["spam.com"]}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsWebSearch()
			},
			validate: func(t *testing.T, result interface{}) {
				search := result.(*tools.WebSearchInput)
				if search.Query != "golang hooks" {
					t.Errorf("Query = %q, want %q", search.Query, "golang hooks")
				}
				if len(search.AllowedDomains) != 1 || search.AllowedDomains[0] != "go.dev" {
					t.Errorf("AllowedDomains = %v, want [go.dev]", search.AllowedDomains)
				}
				if len(search.BlockedDomains) != 1 || search.BlockedDomains[0] != "spam.com" {
					t.Errorf("BlockedDomains = %v, want [spam.com]", search.BlockedDomains)
				}
			},
		},
		{
			name:      "AsTask",
			toolInput: `{"description": "Search code", "prompt": "Find all TODO comments"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsTask()
			},
			validate: func(t *testing.T, result interface{}) {
				task := result.(*tools.TaskInput)
				if task.Description != "Search code" {
					t.Errorf("Description = %q, want %q", task.Description, "Search code")
				}
				if task.Prompt != "Find all TODO comments" {
					t.Errorf("Prompt = %q, want %q", task.Prompt, "Find all TODO comments")
				}
			},
		},
		{
			name:      "AsExitPlanMode",
			toolInput: `{"plan": "1. Fix bug\n2. Add tests\n3. Update docs"}`,
			parser: func(e *cchooks.PreToolUseEvent) (interface{}, error) {
				return e.AsExitPlanMode()
			},
			validate: func(t *testing.T, result interface{}) {
				exit := result.(*tools.ExitPlanModeInput)
				if exit.Plan != "1. Fix bug\n2. Add tests\n3. Update docs" {
					t.Errorf("Plan = %q", exit.Plan)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &cchooks.PreToolUseEvent{
				SessionID: "test",
				ToolName:  "TestTool",
				ToolInput: json.RawMessage(tt.toolInput),
			}

			result, err := tt.parser(event)
			if err != nil {
				t.Fatalf("parser error = %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestPostToolUseEventParsing(t *testing.T) {
	t.Run("input parsing", func(t *testing.T) {
		event := &cchooks.PostToolUseEvent{
			SessionID:    "test",
			ToolName:     "Bash",
			ToolInput:    json.RawMessage(`{"command": "echo test"}`),
			ToolResponse: json.RawMessage(`{"output": "test", "exit_code": 0}`),
		}

		// Test all input parsers
		bash, err := event.InputAsBash()
		if err != nil {
			t.Fatalf("InputAsBash error = %v", err)
		}
		if bash.Command != "echo test" {
			t.Errorf("Command = %q, want %q", bash.Command, "echo test")
		}
	})

	t.Run("response parsing", func(t *testing.T) {
		tests := []struct {
			name         string
			toolResponse string
			parser       func(*cchooks.PostToolUseEvent) (interface{}, error)
			validate     func(t *testing.T, result interface{})
		}{
			{
				name:         "ResponseAsBash",
				toolResponse: `{"output": "Hello\nWorld", "exit_code": 0}`,
				parser: func(e *cchooks.PostToolUseEvent) (interface{}, error) {
					return e.ResponseAsBash()
				},
				validate: func(t *testing.T, result interface{}) {
					bash := result.(*tools.BashOutput)
					if bash.Output != "Hello\nWorld" {
						t.Errorf("Output = %q, want %q", bash.Output, "Hello\nWorld")
					}
					if bash.ExitCode != 0 {
						t.Errorf("ExitCode = %d, want 0", bash.ExitCode)
					}
				},
			},
			{
				name:         "ResponseAsEdit",
				toolResponse: `{"success": true}`,
				parser: func(e *cchooks.PostToolUseEvent) (interface{}, error) {
					return e.ResponseAsEdit()
				},
				validate: func(t *testing.T, result interface{}) {
					edit := result.(*tools.EditOutput)
					if !edit.Success {
						t.Error("Success = false, want true")
					}
				},
			},
			{
				name:         "ResponseAsRead",
				toolResponse: `{"content": "File contents here"}`,
				parser: func(e *cchooks.PostToolUseEvent) (interface{}, error) {
					return e.ResponseAsRead()
				},
				validate: func(t *testing.T, result interface{}) {
					read := result.(*tools.ReadOutput)
					if read.Content != "File contents here" {
						t.Errorf("Content = %q, want %q", read.Content, "File contents here")
					}
				},
			},
			{
				name:         "ResponseAsGlob",
				toolResponse: `{"files": ["main.go", "test.go", "util.go"]}`,
				parser: func(e *cchooks.PostToolUseEvent) (interface{}, error) {
					return e.ResponseAsGlob()
				},
				validate: func(t *testing.T, result interface{}) {
					glob := result.(*tools.GlobOutput)
					if len(glob.Files) != 3 {
						t.Fatalf("len(Files) = %d, want 3", len(glob.Files))
					}
					expected := []string{"main.go", "test.go", "util.go"}
					for i, file := range glob.Files {
						if file != expected[i] {
							t.Errorf("Files[%d] = %q, want %q", i, file, expected[i])
						}
					}
				},
			},
			{
				name:         "ResponseAsGrep",
				toolResponse: `{"files": ["file1.go", "file2.go"]}`,
				parser: func(e *cchooks.PostToolUseEvent) (interface{}, error) {
					return e.ResponseAsGrep()
				},
				validate: func(t *testing.T, result interface{}) {
					grep := result.(*tools.GrepOutput)
					if len(grep.Files) != 2 {
						t.Fatalf("len(Files) = %d, want 2", len(grep.Files))
					}
				},
			},
			{
				name:         "ResponseAsLS",
				toolResponse: `{"files": [{"name": "main.go", "is_dir": false, "size": 1024}, {"name": "pkg", "is_dir": true, "size": 0}]}`,
				parser: func(e *cchooks.PostToolUseEvent) (interface{}, error) {
					return e.ResponseAsLS()
				},
				validate: func(t *testing.T, result interface{}) {
					ls := result.(*tools.LSOutput)
					if len(ls.Files) != 2 {
						t.Fatalf("len(Files) = %d, want 2", len(ls.Files))
					}
					if ls.Files[0].Name != "main.go" {
						t.Errorf("Files[0].Name = %q, want %q", ls.Files[0].Name, "main.go")
					}
					if ls.Files[0].IsDir {
						t.Error("Files[0].IsDir = true, want false")
					}
					if ls.Files[0].Size != 1024 {
						t.Errorf("Files[0].Size = %d, want 1024", ls.Files[0].Size)
					}
					if ls.Files[1].Name != "pkg" {
						t.Errorf("Files[1].Name = %q, want %q", ls.Files[1].Name, "pkg")
					}
					if !ls.Files[1].IsDir {
						t.Error("Files[1].IsDir = false, want true")
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				event := &cchooks.PostToolUseEvent{
					SessionID:    "test",
					ToolName:     "TestTool",
					ToolInput:    json.RawMessage(`{}`),
					ToolResponse: json.RawMessage(tt.toolResponse),
				}

				result, err := tt.parser(event)
				if err != nil {
					t.Fatalf("parser error = %v", err)
				}

				tt.validate(t, result)
			})
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		event := &cchooks.PreToolUseEvent{
			SessionID: "test",
			ToolName:  "Bash",
			ToolInput: json.RawMessage(`{"invalid json`),
		}

		_, err := event.AsBash()
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}