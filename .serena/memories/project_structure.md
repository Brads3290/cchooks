# Project Structure

## Root Directory Files
- `doc.go` - Package documentation with examples
- `runner.go` - Main Runner type and event dispatch logic
- `events.go` - Event type definitions
- `responses.go` - Response type definitions and helpers
- `types.go` - Core type definitions
- `testing.go` - Testing utilities for hook developers
- `transcript.go` - Transcript support for Stop events
- `mcp_test.go` - MCP (Model Context Protocol) tool tests

## Directories
- `/examples/` - Example hook implementations
  - `simple-hook/` - Basic approval hook
  - `security-hook/` - Security-focused hook with dangerous command blocking
  - `format-hook/` - Hook that formats messages
  - `mcp-hook/` - MCP tool integration example
  - `stop-once-hook/` - Example using StopOnce handler
  - `transcript-analyzer/` - Analyzes conversation transcripts
  - `debug-hook/` - Debugging helper hook

- `/internal/tools/` - Internal tool parsing utilities
  - `tools.go` - Tool input/output type definitions
  - `tools_test.go` - Tool parsing tests

- `/serena memory/` - AI memory files (created by Serena MCP)

## Configuration Files
- `go.mod` - Go module definition
- `Makefile` - Build automation
- `.mcp.json` - MCP server configuration
- `CLAUDE.md` - Project-specific Claude instructions
- `.gitignore` - Git ignore patterns

## Documentation
- `README.md` - Main project documentation
- `CHANGELOG.md` - Version history
- `RELEASE.md` - Release process documentation
- `LICENSE` - MIT License
- `claude-code-hooks-go-sdk-design.md` - SDK design documentation