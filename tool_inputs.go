package cchooks

// Strongly typed tool input structures
type BashInput struct {
	Command     string `json:"command" validate:"required"`
	Timeout     *int   `json:"timeout,omitempty" validate:"omitempty,max=600000"`
	Description string `json:"description,omitempty"`
}

type EditInput struct {
	FilePath   string `json:"file_path" validate:"required,filepath"`
	OldString  string `json:"old_string" validate:"required"`
	NewString  string `json:"new_string" validate:"required,nefield=OldString"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

type MultiEditInput struct {
	FilePath string      `json:"file_path" validate:"required,filepath"`
	Edits    []EditEntry `json:"edits" validate:"required,min=1,dive"`
}

type EditEntry struct {
	OldString  string `json:"old_string" validate:"required"`
	NewString  string `json:"new_string" validate:"required"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

type WriteInput struct {
	FilePath string `json:"file_path" validate:"required,filepath"`
	Content  string `json:"content" validate:"required"`
}

type ReadInput struct {
	FilePath string `json:"file_path" validate:"required,filepath"`
	Limit    *int   `json:"limit,omitempty"`
	Offset   *int   `json:"offset,omitempty"`
}

type GlobInput struct {
	Pattern string `json:"pattern" validate:"required"`
	Path    string `json:"path,omitempty" validate:"omitempty,dirpath"`
}

type GrepInput struct {
	Pattern string `json:"pattern" validate:"required"`
	Path    string `json:"path,omitempty" validate:"omitempty,dirpath"`
	Include string `json:"include,omitempty"`
}

type LSInput struct {
	Path   string   `json:"path" validate:"required,dirpath"`
	Ignore []string `json:"ignore,omitempty"`
}

type TodoWriteInput struct {
	Todos []TodoItem `json:"todos" validate:"required,min=1,dive"`
}

type TodoItem struct {
	Content  string       `json:"content" validate:"required,min=1"`
	Status   TodoStatus   `json:"status" validate:"required,oneof=pending in_progress completed"`
	Priority TodoPriority `json:"priority" validate:"required,oneof=high medium low"`
	ID       string       `json:"id" validate:"required"`
}

type TodoStatus string
type TodoPriority string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"

	TodoPriorityHigh   TodoPriority = "high"
	TodoPriorityMedium TodoPriority = "medium"
	TodoPriorityLow    TodoPriority = "low"
)

type TodoReadInput struct{}

type NotebookReadInput struct {
	NotebookPath string `json:"notebook_path" validate:"required,filepath"`
	CellID       string `json:"cell_id,omitempty"`
}

type NotebookEditInput struct {
	NotebookPath string `json:"notebook_path" validate:"required,filepath"`
	CellID       string `json:"cell_id,omitempty"`
	CellType     string `json:"cell_type,omitempty" validate:"omitempty,oneof=code markdown"`
	EditMode     string `json:"edit_mode,omitempty" validate:"omitempty,oneof=replace insert delete"`
	NewSource    string `json:"new_source" validate:"required"`
}

type WebFetchInput struct {
	URL    string `json:"url" validate:"required,url"`
	Prompt string `json:"prompt" validate:"required"`
}

type WebSearchInput struct {
	Query          string   `json:"query" validate:"required,min=2"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	BlockedDomains []string `json:"blocked_domains,omitempty"`
}

type TaskInput struct {
	Description string `json:"description" validate:"required"`
	Prompt      string `json:"prompt" validate:"required"`
}

type ExitPlanModeInput struct {
	Plan string `json:"plan" validate:"required"`
}

// Tool output types
type BashOutput struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
}

type EditOutput struct {
	Success bool `json:"success"`
}

type ReadOutput struct {
	Content string `json:"content"`
}

type GlobOutput struct {
	Files []string `json:"files"`
}

type GrepOutput struct {
	Files []string `json:"files"`
}

type LSOutput struct {
	Files []FileInfo `json:"files"`
}

type FileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}
