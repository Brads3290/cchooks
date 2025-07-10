package cchooks

import (
	"github.com/brads3290/cchooks/internal/tools"
)

// Re-export commonly used tool types for testing and external usage

// Tool input types
type BashInput = tools.BashInput
type EditInput = tools.EditInput
type MultiEditInput = tools.MultiEditInput
type EditEntry = tools.EditEntry
type WriteInput = tools.WriteInput
type ReadInput = tools.ReadInput
type GlobInput = tools.GlobInput
type GrepInput = tools.GrepInput
type LSInput = tools.LSInput
type TodoWriteInput = tools.TodoWriteInput
type TodoItem = tools.TodoItem
type TodoStatus = tools.TodoStatus
type TodoPriority = tools.TodoPriority
type TodoReadInput = tools.TodoReadInput
type NotebookReadInput = tools.NotebookReadInput
type NotebookEditInput = tools.NotebookEditInput
type WebFetchInput = tools.WebFetchInput
type WebSearchInput = tools.WebSearchInput
type TaskInput = tools.TaskInput
type ExitPlanModeInput = tools.ExitPlanModeInput

// Tool output types
type BashOutput = tools.BashOutput
type EditOutput = tools.EditOutput
type ReadOutput = tools.ReadOutput
type GlobOutput = tools.GlobOutput
type GrepOutput = tools.GrepOutput
type LSOutput = tools.LSOutput
type FileInfo = tools.FileInfo

// Todo constants
const (
	TodoStatusPending    = tools.TodoStatusPending
	TodoStatusInProgress = tools.TodoStatusInProgress
	TodoStatusCompleted  = tools.TodoStatusCompleted
	TodoPriorityHigh     = tools.TodoPriorityHigh
	TodoPriorityMedium   = tools.TodoPriorityMedium
	TodoPriorityLow      = tools.TodoPriorityLow
)