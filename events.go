package cchooks

import (
	"encoding/json"

	"github.com/brads3290/cchooks/internal/tools"
)

// Event types - data containers for each hook event
type PreToolUseEvent struct {
	SessionID string          `json:"session_id"`
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

type PostToolUseEvent struct {
	SessionID    string          `json:"session_id"`
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

type NotificationEvent struct {
	SessionID string `json:"session_id"`
	Message   string `json:"notification_message"`
}

type StopEvent struct {
	SessionID      string             `json:"session_id"`
	StopHookActive bool               `json:"stop_hook_active"`
	TranscriptPath string             `json:"transcript_path"`
	Transcript     []TranscriptEntry  `json:"transcript"`
}

// Interface implementations for tools package

// GetToolInput implements tools.EventWithToolInput for PreToolUseEvent.
func (e *PreToolUseEvent) GetToolInput() json.RawMessage {
	return e.ToolInput
}

// GetToolInput implements tools.EventWithToolInput for PostToolUseEvent.
func (e *PostToolUseEvent) GetToolInput() json.RawMessage {
	return e.ToolInput
}

// GetToolResponse implements tools.EventWithToolResponse for PostToolUseEvent.
func (e *PostToolUseEvent) GetToolResponse() json.RawMessage {
	return e.ToolResponse
}

// Convenience parsing methods for PreToolUseEvent

// AsBash parses the tool input as BashInput.
func (e *PreToolUseEvent) AsBash() (*tools.BashInput, error) {
	return tools.ParseBash(e)
}

// AsEdit parses the tool input as EditInput.
func (e *PreToolUseEvent) AsEdit() (*tools.EditInput, error) {
	return tools.ParseEdit(e)
}

// AsMultiEdit parses the tool input as MultiEditInput.
func (e *PreToolUseEvent) AsMultiEdit() (*tools.MultiEditInput, error) {
	return tools.ParseMultiEdit(e)
}

// AsWrite parses the tool input as WriteInput.
func (e *PreToolUseEvent) AsWrite() (*tools.WriteInput, error) {
	return tools.ParseWrite(e)
}

// AsRead parses the tool input as ReadInput.
func (e *PreToolUseEvent) AsRead() (*tools.ReadInput, error) {
	return tools.ParseRead(e)
}

// AsGlob parses the tool input as GlobInput.
func (e *PreToolUseEvent) AsGlob() (*tools.GlobInput, error) {
	return tools.ParseGlob(e)
}

// AsGrep parses the tool input as GrepInput.
func (e *PreToolUseEvent) AsGrep() (*tools.GrepInput, error) {
	return tools.ParseGrep(e)
}

// AsLS parses the tool input as LSInput.
func (e *PreToolUseEvent) AsLS() (*tools.LSInput, error) {
	return tools.ParseLS(e)
}

// AsTodoWrite parses the tool input as TodoWriteInput.
func (e *PreToolUseEvent) AsTodoWrite() (*tools.TodoWriteInput, error) {
	return tools.ParseTodoWrite(e)
}

// AsTodoRead parses the tool input as TodoReadInput.
func (e *PreToolUseEvent) AsTodoRead() (*tools.TodoReadInput, error) {
	return tools.ParseTodoRead(e)
}

// AsNotebookRead parses the tool input as NotebookReadInput.
func (e *PreToolUseEvent) AsNotebookRead() (*tools.NotebookReadInput, error) {
	return tools.ParseNotebookRead(e)
}

// AsNotebookEdit parses the tool input as NotebookEditInput.
func (e *PreToolUseEvent) AsNotebookEdit() (*tools.NotebookEditInput, error) {
	return tools.ParseNotebookEdit(e)
}

// AsWebFetch parses the tool input as WebFetchInput.
func (e *PreToolUseEvent) AsWebFetch() (*tools.WebFetchInput, error) {
	return tools.ParseWebFetch(e)
}

// AsWebSearch parses the tool input as WebSearchInput.
func (e *PreToolUseEvent) AsWebSearch() (*tools.WebSearchInput, error) {
	return tools.ParseWebSearch(e)
}

// AsTask parses the tool input as TaskInput.
func (e *PreToolUseEvent) AsTask() (*tools.TaskInput, error) {
	return tools.ParseTask(e)
}

// AsExitPlanMode parses the tool input as ExitPlanModeInput.
func (e *PreToolUseEvent) AsExitPlanMode() (*tools.ExitPlanModeInput, error) {
	return tools.ParseExitPlanMode(e)
}

// Convenience parsing methods for PostToolUseEvent - Input

// InputAsBash parses the tool input as BashInput.
func (e *PostToolUseEvent) InputAsBash() (*tools.BashInput, error) {
	return tools.ParseBash(e)
}

// InputAsEdit parses the tool input as EditInput.
func (e *PostToolUseEvent) InputAsEdit() (*tools.EditInput, error) {
	return tools.ParseEdit(e)
}

// InputAsMultiEdit parses the tool input as MultiEditInput.
func (e *PostToolUseEvent) InputAsMultiEdit() (*tools.MultiEditInput, error) {
	return tools.ParseMultiEdit(e)
}

// InputAsWrite parses the tool input as WriteInput.
func (e *PostToolUseEvent) InputAsWrite() (*tools.WriteInput, error) {
	return tools.ParseWrite(e)
}

// InputAsRead parses the tool input as ReadInput.
func (e *PostToolUseEvent) InputAsRead() (*tools.ReadInput, error) {
	return tools.ParseRead(e)
}

// InputAsGlob parses the tool input as GlobInput.
func (e *PostToolUseEvent) InputAsGlob() (*tools.GlobInput, error) {
	return tools.ParseGlob(e)
}

// InputAsGrep parses the tool input as GrepInput.
func (e *PostToolUseEvent) InputAsGrep() (*tools.GrepInput, error) {
	return tools.ParseGrep(e)
}

// InputAsLS parses the tool input as LSInput.
func (e *PostToolUseEvent) InputAsLS() (*tools.LSInput, error) {
	return tools.ParseLS(e)
}

// InputAsTodoWrite parses the tool input as TodoWriteInput.
func (e *PostToolUseEvent) InputAsTodoWrite() (*tools.TodoWriteInput, error) {
	return tools.ParseTodoWrite(e)
}

// InputAsTodoRead parses the tool input as TodoReadInput.
func (e *PostToolUseEvent) InputAsTodoRead() (*tools.TodoReadInput, error) {
	return tools.ParseTodoRead(e)
}

// InputAsNotebookRead parses the tool input as NotebookReadInput.
func (e *PostToolUseEvent) InputAsNotebookRead() (*tools.NotebookReadInput, error) {
	return tools.ParseNotebookRead(e)
}

// InputAsNotebookEdit parses the tool input as NotebookEditInput.
func (e *PostToolUseEvent) InputAsNotebookEdit() (*tools.NotebookEditInput, error) {
	return tools.ParseNotebookEdit(e)
}

// InputAsWebFetch parses the tool input as WebFetchInput.
func (e *PostToolUseEvent) InputAsWebFetch() (*tools.WebFetchInput, error) {
	return tools.ParseWebFetch(e)
}

// InputAsWebSearch parses the tool input as WebSearchInput.
func (e *PostToolUseEvent) InputAsWebSearch() (*tools.WebSearchInput, error) {
	return tools.ParseWebSearch(e)
}

// InputAsTask parses the tool input as TaskInput.
func (e *PostToolUseEvent) InputAsTask() (*tools.TaskInput, error) {
	return tools.ParseTask(e)
}

// InputAsExitPlanMode parses the tool input as ExitPlanModeInput.
func (e *PostToolUseEvent) InputAsExitPlanMode() (*tools.ExitPlanModeInput, error) {
	return tools.ParseExitPlanMode(e)
}

// Convenience parsing methods for PostToolUseEvent - Response

// ResponseAsBash parses the tool response as BashOutput.
func (e *PostToolUseEvent) ResponseAsBash() (*tools.BashOutput, error) {
	return tools.ParseBashResponse(e)
}

// ResponseAsEdit parses the tool response as EditOutput.
func (e *PostToolUseEvent) ResponseAsEdit() (*tools.EditOutput, error) {
	return tools.ParseEditResponse(e)
}

// ResponseAsRead parses the tool response as ReadOutput.
func (e *PostToolUseEvent) ResponseAsRead() (*tools.ReadOutput, error) {
	return tools.ParseReadResponse(e)
}

// ResponseAsGlob parses the tool response as GlobOutput.
func (e *PostToolUseEvent) ResponseAsGlob() (*tools.GlobOutput, error) {
	return tools.ParseGlobResponse(e)
}

// ResponseAsGrep parses the tool response as GrepOutput.
func (e *PostToolUseEvent) ResponseAsGrep() (*tools.GrepOutput, error) {
	return tools.ParseGrepResponse(e)
}

// ResponseAsLS parses the tool response as LSOutput.
func (e *PostToolUseEvent) ResponseAsLS() (*tools.LSOutput, error) {
	return tools.ParseLSResponse(e)
}

