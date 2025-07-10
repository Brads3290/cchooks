# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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