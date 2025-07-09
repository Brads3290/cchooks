package cchooks

import "encoding/json"

// Tool input parsing methods on PreToolUseEvent
func (e *PreToolUseEvent) AsBash() (*BashInput, error) {
	var input BashInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsEdit() (*EditInput, error) {
	var input EditInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsMultiEdit() (*MultiEditInput, error) {
	var input MultiEditInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsWrite() (*WriteInput, error) {
	var input WriteInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsRead() (*ReadInput, error) {
	var input ReadInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsGlob() (*GlobInput, error) {
	var input GlobInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsGrep() (*GrepInput, error) {
	var input GrepInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsLS() (*LSInput, error) {
	var input LSInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsTodoWrite() (*TodoWriteInput, error) {
	var input TodoWriteInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsTodoRead() (*TodoReadInput, error) {
	var input TodoReadInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsNotebookRead() (*NotebookReadInput, error) {
	var input NotebookReadInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsNotebookEdit() (*NotebookEditInput, error) {
	var input NotebookEditInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsWebFetch() (*WebFetchInput, error) {
	var input WebFetchInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsWebSearch() (*WebSearchInput, error) {
	var input WebSearchInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsTask() (*TaskInput, error) {
	var input TaskInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PreToolUseEvent) AsExitPlanMode() (*ExitPlanModeInput, error) {
	var input ExitPlanModeInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

// PostToolUse can access both input and response
func (e *PostToolUseEvent) InputAsBash() (*BashInput, error) {
	var input BashInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsEdit() (*EditInput, error) {
	var input EditInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsMultiEdit() (*MultiEditInput, error) {
	var input MultiEditInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsWrite() (*WriteInput, error) {
	var input WriteInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsRead() (*ReadInput, error) {
	var input ReadInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsGlob() (*GlobInput, error) {
	var input GlobInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsGrep() (*GrepInput, error) {
	var input GrepInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsLS() (*LSInput, error) {
	var input LSInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsTodoWrite() (*TodoWriteInput, error) {
	var input TodoWriteInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsTodoRead() (*TodoReadInput, error) {
	var input TodoReadInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsNotebookRead() (*NotebookReadInput, error) {
	var input NotebookReadInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsNotebookEdit() (*NotebookEditInput, error) {
	var input NotebookEditInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsWebFetch() (*WebFetchInput, error) {
	var input WebFetchInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsWebSearch() (*WebSearchInput, error) {
	var input WebSearchInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsTask() (*TaskInput, error) {
	var input TaskInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

func (e *PostToolUseEvent) InputAsExitPlanMode() (*ExitPlanModeInput, error) {
	var input ExitPlanModeInput
	return &input, json.Unmarshal(e.ToolInput, &input)
}

// Response parsing methods
func (e *PostToolUseEvent) ResponseAsBash() (*BashOutput, error) {
	var output BashOutput
	return &output, json.Unmarshal(e.ToolResponse, &output)
}

func (e *PostToolUseEvent) ResponseAsEdit() (*EditOutput, error) {
	var output EditOutput
	return &output, json.Unmarshal(e.ToolResponse, &output)
}

func (e *PostToolUseEvent) ResponseAsRead() (*ReadOutput, error) {
	var output ReadOutput
	return &output, json.Unmarshal(e.ToolResponse, &output)
}

func (e *PostToolUseEvent) ResponseAsGlob() (*GlobOutput, error) {
	var output GlobOutput
	return &output, json.Unmarshal(e.ToolResponse, &output)
}

func (e *PostToolUseEvent) ResponseAsGrep() (*GrepOutput, error) {
	var output GrepOutput
	return &output, json.Unmarshal(e.ToolResponse, &output)
}

func (e *PostToolUseEvent) ResponseAsLS() (*LSOutput, error) {
	var output LSOutput
	return &output, json.Unmarshal(e.ToolResponse, &output)
}
