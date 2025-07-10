// Package tools provides strongly typed tool input/output structures and parsing methods for Claude Code hooks.
package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Tool input types

// BashInput represents input for the Bash tool.
type BashInput struct {
	Command     string `json:"command" validate:"required"`
	Timeout     *int   `json:"timeout,omitempty" validate:"omitempty,max=600000"`
	Description string `json:"description,omitempty"`
}

// EditInput represents input for the Edit tool.
type EditInput struct {
	FilePath   string `json:"file_path" validate:"required,filepath"`
	OldString  string `json:"old_string" validate:"required"`
	NewString  string `json:"new_string" validate:"required,nefield=OldString"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// MultiEditInput represents input for the MultiEdit tool.
type MultiEditInput struct {
	FilePath string      `json:"file_path" validate:"required,filepath"`
	Edits    []EditEntry `json:"edits" validate:"required,min=1,dive"`
}

// EditEntry represents a single edit operation in MultiEdit.
type EditEntry struct {
	OldString  string `json:"old_string" validate:"required"`
	NewString  string `json:"new_string" validate:"required"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// WriteInput represents input for the Write tool.
type WriteInput struct {
	FilePath string `json:"file_path" validate:"required,filepath"`
	Content  string `json:"content" validate:"required"`
}

// ReadInput represents input for the Read tool.
type ReadInput struct {
	FilePath string `json:"file_path" validate:"required,filepath"`
	Limit    *int   `json:"limit,omitempty"`
	Offset   *int   `json:"offset,omitempty"`
}

// GlobInput represents input for the Glob tool.
type GlobInput struct {
	Pattern string `json:"pattern" validate:"required"`
	Path    string `json:"path,omitempty" validate:"omitempty,dirpath"`
}

// GrepInput represents input for the Grep tool.
type GrepInput struct {
	Pattern string `json:"pattern" validate:"required"`
	Path    string `json:"path,omitempty" validate:"omitempty,dirpath"`
	Include string `json:"include,omitempty"`
}

// LSInput represents input for the LS tool.
type LSInput struct {
	Path   string   `json:"path" validate:"required,dirpath"`
	Ignore []string `json:"ignore,omitempty"`
}

// TodoWriteInput represents input for the TodoWrite tool.
type TodoWriteInput struct {
	Todos []TodoItem `json:"todos" validate:"required,min=1,dive"`
}

// TodoItem represents a single todo item.
type TodoItem struct {
	Content  string       `json:"content" validate:"required,min=1"`
	Status   TodoStatus   `json:"status" validate:"required,oneof=pending in_progress completed"`
	Priority TodoPriority `json:"priority" validate:"required,oneof=high medium low"`
	ID       string       `json:"id" validate:"required"`
}

// TodoStatus represents the status of a todo item.
type TodoStatus string

// TodoPriority represents the priority of a todo item.
type TodoPriority string

// Todo status constants.
const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

// Todo priority constants.
const (
	TodoPriorityHigh   TodoPriority = "high"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityLow    TodoPriority = "low"
)

// TodoReadInput represents input for the TodoRead tool.
type TodoReadInput struct{}

// NotebookReadInput represents input for the NotebookRead tool.
type NotebookReadInput struct {
	NotebookPath string `json:"notebook_path" validate:"required,filepath"`
	CellID       string `json:"cell_id,omitempty"`
}

// NotebookEditInput represents input for the NotebookEdit tool.
type NotebookEditInput struct {
	NotebookPath string `json:"notebook_path" validate:"required,filepath"`
	CellID       string `json:"cell_id,omitempty"`
	CellType     string `json:"cell_type,omitempty" validate:"omitempty,oneof=code markdown"`
	EditMode     string `json:"edit_mode,omitempty" validate:"omitempty,oneof=replace insert delete"`
	NewSource    string `json:"new_source" validate:"required"`
}

// WebFetchInput represents input for the WebFetch tool.
type WebFetchInput struct {
	URL    string `json:"url" validate:"required,url"`
	Prompt string `json:"prompt" validate:"required"`
}

// WebSearchInput represents input for the WebSearch tool.
type WebSearchInput struct {
	Query          string   `json:"query" validate:"required,min=2"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	BlockedDomains []string `json:"blocked_domains,omitempty"`
}

// TaskInput represents input for the Task tool.
type TaskInput struct {
	Description string `json:"description" validate:"required"`
	Prompt      string `json:"prompt" validate:"required"`
}

// ExitPlanModeInput represents input for the ExitPlanMode tool.
type ExitPlanModeInput struct {
	Plan string `json:"plan" validate:"required"`
}

// MCP tool types

// MCPTool represents a parsed MCP tool with extracted server and tool names.
type MCPTool struct {
	MCPName  string          `json:"mcp_name"`  // e.g., "weather" (extracted from mcp__weather__get_forecast)
	ToolName string          `json:"tool_name"` // e.g., "get_forecast" (part after server name)
	RawInput json.RawMessage `json:"raw_input"` // Raw JSON input for flexible parsing
}

// MCPToolOutput represents output from an MCP tool.
type MCPToolOutput struct {
	MCPName   string          `json:"mcp_name"`   // e.g., "weather"
	ToolName  string          `json:"tool_name"`  // e.g., "get_forecast"
	RawOutput json.RawMessage `json:"raw_output"` // Raw JSON output
}

// Tool output types

// BashOutput represents output from the Bash tool.
type BashOutput struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

// EditOutput represents output from the Edit tool.
type EditOutput struct {
	Success bool `json:"success"`
}

// ReadOutput represents output from the Read tool.
type ReadOutput struct {
	Content string `json:"content"`
}

// GlobOutput represents output from the Glob tool.
type GlobOutput struct {
	Files []string `json:"files"`
}

// GrepOutput represents output from the Grep tool.
type GrepOutput struct {
	Files []string `json:"files"`
}

// LSOutput represents output from the LS tool.
type LSOutput struct {
	Files []FileInfo `json:"files"`
}

// FileInfo represents information about a file or directory.
type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// EventWithToolInput represents an event that contains tool input data.
type EventWithToolInput interface {
	GetToolInput() json.RawMessage
}

// EventWithToolResponse represents an event that contains tool response data.
type EventWithToolResponse interface {
	GetToolResponse() json.RawMessage
}

// ParseBash parses the event's tool input as BashInput.
func ParseBash(e EventWithToolInput) (*BashInput, error) {
	var input BashInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseEdit parses the event's tool input as EditInput.
func ParseEdit(e EventWithToolInput) (*EditInput, error) {
	var input EditInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseMultiEdit parses the event's tool input as MultiEditInput.
func ParseMultiEdit(e EventWithToolInput) (*MultiEditInput, error) {
	var input MultiEditInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseWrite parses the event's tool input as WriteInput.
func ParseWrite(e EventWithToolInput) (*WriteInput, error) {
	var input WriteInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseRead parses the event's tool input as ReadInput.
func ParseRead(e EventWithToolInput) (*ReadInput, error) {
	var input ReadInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseGlob parses the event's tool input as GlobInput.
func ParseGlob(e EventWithToolInput) (*GlobInput, error) {
	var input GlobInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseGrep parses the event's tool input as GrepInput.
func ParseGrep(e EventWithToolInput) (*GrepInput, error) {
	var input GrepInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseLS parses the event's tool input as LSInput.
func ParseLS(e EventWithToolInput) (*LSInput, error) {
	var input LSInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseTodoWrite parses the event's tool input as TodoWriteInput.
func ParseTodoWrite(e EventWithToolInput) (*TodoWriteInput, error) {
	var input TodoWriteInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseTodoRead parses the event's tool input as TodoReadInput.
func ParseTodoRead(e EventWithToolInput) (*TodoReadInput, error) {
	var input TodoReadInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseNotebookRead parses the event's tool input as NotebookReadInput.
func ParseNotebookRead(e EventWithToolInput) (*NotebookReadInput, error) {
	var input NotebookReadInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseNotebookEdit parses the event's tool input as NotebookEditInput.
func ParseNotebookEdit(e EventWithToolInput) (*NotebookEditInput, error) {
	var input NotebookEditInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseWebFetch parses the event's tool input as WebFetchInput.
func ParseWebFetch(e EventWithToolInput) (*WebFetchInput, error) {
	var input WebFetchInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseWebSearch parses the event's tool input as WebSearchInput.
func ParseWebSearch(e EventWithToolInput) (*WebSearchInput, error) {
	var input WebSearchInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseTask parses the event's tool input as TaskInput.
func ParseTask(e EventWithToolInput) (*TaskInput, error) {
	var input TaskInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseExitPlanMode parses the event's tool input as ExitPlanModeInput.
func ParseExitPlanMode(e EventWithToolInput) (*ExitPlanModeInput, error) {
	var input ExitPlanModeInput
	return &input, json.Unmarshal(e.GetToolInput(), &input)
}

// ParseBashResponse parses the event's tool response as BashOutput.
func ParseBashResponse(e EventWithToolResponse) (*BashOutput, error) {
	var output BashOutput
	return &output, json.Unmarshal(e.GetToolResponse(), &output)
}

// ParseEditResponse parses the event's tool response as EditOutput.
func ParseEditResponse(e EventWithToolResponse) (*EditOutput, error) {
	var output EditOutput
	return &output, json.Unmarshal(e.GetToolResponse(), &output)
}

// ParseReadResponse parses the event's tool response as ReadOutput.
func ParseReadResponse(e EventWithToolResponse) (*ReadOutput, error) {
	var output ReadOutput
	return &output, json.Unmarshal(e.GetToolResponse(), &output)
}

// ParseGlobResponse parses the event's tool response as GlobOutput.
func ParseGlobResponse(e EventWithToolResponse) (*GlobOutput, error) {
	var output GlobOutput
	return &output, json.Unmarshal(e.GetToolResponse(), &output)
}

// ParseGrepResponse parses the event's tool response as GrepOutput.
func ParseGrepResponse(e EventWithToolResponse) (*GrepOutput, error) {
	var output GrepOutput
	return &output, json.Unmarshal(e.GetToolResponse(), &output)
}

// ParseLSResponse parses the event's tool response as LSOutput.
func ParseLSResponse(e EventWithToolResponse) (*LSOutput, error) {
	var output LSOutput
	return &output, json.Unmarshal(e.GetToolResponse(), &output)
}

// ParseMCPTool parses an MCP tool from the tool name and event.
func ParseMCPTool(toolName string, e EventWithToolInput) (*MCPTool, error) {
	if !strings.HasPrefix(toolName, "mcp__") {
		return nil, fmt.Errorf("not an MCP tool: %s", toolName)
	}

	// Extract MCP server name and tool name from the full tool name
	// Format: mcp__servername__toolname
	parts := strings.SplitN(toolName, "__", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid MCP tool name format: %s", toolName)
	}

	return &MCPTool{
		MCPName:  parts[1],                     // servername
		ToolName: strings.Join(parts[2:], "__"), // toolname (may contain __)
		RawInput: e.GetToolInput(),
	}, nil
}

// ParseMCPToolResponse parses an MCP tool response from the tool name and event.
func ParseMCPToolResponse(toolName string, e EventWithToolResponse) (*MCPToolOutput, error) {
	if !strings.HasPrefix(toolName, "mcp__") {
		return nil, fmt.Errorf("not an MCP tool: %s", toolName)
	}

	// Extract MCP server name and tool name from the full tool name
	parts := strings.SplitN(toolName, "__", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid MCP tool name format: %s", toolName)
	}

	return &MCPToolOutput{
		MCPName:   parts[1],                     // servername
		ToolName:  strings.Join(parts[2:], "__"), // toolname (may contain __)
		RawOutput: e.GetToolResponse(),
	}, nil
}