# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.4.0] - 2025-01-10

### Changed
- Error handler now returns `*RawResponse` instead of void
  - Allows custom exit codes and output for error cases
  - If nil is returned, default error handling applies (exit code 2)
- SDK now recovers from panics and sends them to the error handler
  - Panics are converted to errors with "panic: " prefix
  - Error handler is called before the process exits

## [v0.3.0] - 2025-01-10

### Changed
- **BREAKING**: Renamed module from `github.com/brads3290/claude-code-hooks-go` to `github.com/brads3290/cchooks`
  - Shorter import path for better developer experience
  - All imports must be updated to use the new module name

## [v0.2.0] - 2025-01-10

### Added
- Raw handler for processing raw JSON before event dispatch
  - Executes before any JSON parsing or event handling
  - Can return custom exit codes and output
  - Useful for custom protocols, logging, and preprocessing
  - Returns nil to continue with normal event processing

## [v0.1.0] - 2025-01-10

### Added
- Initial release of Claude Code Hooks Go SDK
- Type-safe event and response handling for all Claude Code tools
- Runner for handling hook execution
- Error handler to Runner for custom error handling
  - Receives raw JSON and error object when SDK errors occur
  - Called for JSON parsing, event validation, handler, and response encoding errors
- Testing utilities with assertion helpers
- Example hooks (security, format, simple)
- Support for all Claude Code event types:
  - PreToolUse
  - PostToolUse
  - Notification
  - Stop
- Typed parsing for all 15+ Claude Code tools
- Comprehensive test coverage
- Integration tests for end-to-end validation