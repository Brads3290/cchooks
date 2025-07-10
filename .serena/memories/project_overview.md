# Claude Code Hooks Go SDK

## Project Purpose
A Go SDK for creating strongly typed Claude Code hooks. This SDK simplifies the creation of individual hook binaries that handle Claude Code events with type safety and comprehensive testing utilities.

## Tech Stack
- **Language**: Go 1.24.4
- **Module**: github.com/brads3290/cchooks
- **Type**: Library/SDK
- **Dependencies**: None (standard library only)

## Main Features
- Strongly typed event handling for Claude Code hooks
- Support for PreToolUse, PostToolUse, Notification, and Stop events
- Type-safe tool input parsing for all Claude Code tools
- Comprehensive testing utilities
- Raw handler support for custom protocols
- Error handling with customizable responses
- StopOnce handler for handling first stop event only

## Key Components
- `Runner`: Main hook processor that handles event dispatch
- Event types: PreToolUseEvent, PostToolUseEvent, NotificationEvent, StopEvent
- Response helpers: Approve(), Block(), Allow(), etc.
- Testing utilities: TestRunner for validating hook behavior
- Tool parsing: Typed parsers for Bash, Edit, Write, and other Claude tools