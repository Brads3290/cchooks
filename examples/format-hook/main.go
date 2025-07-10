package main

import (
	"context"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	cchooks "github.com/brads3290/cchooks"
)

func main() {
	runner := &cchooks.Runner{
		PostToolUse: func(ctx context.Context, event *cchooks.PostToolUseEvent) (*cchooks.PostToolUseResponse, error) {
			// Only process successful file modifications
			if event.ToolName != "Edit" && event.ToolName != "Write" && event.ToolName != "MultiEdit" {
				return cchooks.Allow(), nil
			}

			// Get file path
			var filePath string
			switch event.ToolName {
			case "Edit":
				if edit, err := event.InputAsEdit(); err == nil {
					filePath = edit.FilePath
				}
			case "MultiEdit":
				if multi, err := event.InputAsMultiEdit(); err == nil {
					filePath = multi.FilePath
				}
			case "Write":
				if write, err := event.InputAsWrite(); err == nil {
					filePath = write.FilePath
				}
			}

			if filePath == "" {
				return cchooks.Allow(), nil
			}

			// Format based on file extension
			ext := strings.ToLower(filepath.Ext(filePath))
			var cmd *exec.Cmd

			switch ext {
			case ".go":
				cmd = exec.Command("gofmt", "-w", filePath)
			case ".js", ".jsx", ".ts", ".tsx", ".json":
				// Check if prettier is available
				if _, err := exec.LookPath("prettier"); err == nil {
					cmd = exec.Command("prettier", "--write", filePath)
				}
			case ".py":
				// Check if black is available
				if _, err := exec.LookPath("black"); err == nil {
					cmd = exec.Command("black", "-q", filePath)
				}
			case ".rs":
				// Check if rustfmt is available
				if _, err := exec.LookPath("rustfmt"); err == nil {
					cmd = exec.Command("rustfmt", filePath)
				}
			case ".java":
				// Check if google-java-format is available
				if _, err := exec.LookPath("google-java-format"); err == nil {
					cmd = exec.Command("google-java-format", "-i", filePath)
				}
			case ".c", ".cpp", ".cc", ".h", ".hpp":
				// Check if clang-format is available
				if _, err := exec.LookPath("clang-format"); err == nil {
					cmd = exec.Command("clang-format", "-i", filePath)
				}
			case ".rb":
				// Check if rubocop is available
				if _, err := exec.LookPath("rubocop"); err == nil {
					cmd = exec.Command("rubocop", "-a", filePath)
				}
			}

			if cmd != nil {
				log.Printf("Formatting %s with %s", filePath, cmd.Path)
				if output, err := cmd.CombinedOutput(); err != nil {
					log.Printf("Format error: %v\nOutput: %s", err, output)
				}
			}

			return cchooks.Allow(), nil
		},
	}

	runner.Run()
}
