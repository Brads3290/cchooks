/*
Package cchooks provides a Go SDK for creating strongly typed Claude Code hooks.

This SDK simplifies the creation of individual hook binaries that handle Claude Code
events with type safety and comprehensive testing utilities.

# Basic Usage

Create a hook by defining handlers for the events you want to process:

	package main

	import (
	    "context"
	    "log"
	    cchooks "github.com/brads3290/cchooks"
	)

	func main() {
	    runner := &cchooks.Runner{
	        PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
	            // Your logic here
	            return cchooks.Approve(), nil
	        },
	    }

	    if err := runner.Run(context.Background()); err != nil {
	        log.Fatal(err)
	    }
	}

# Event Types

The SDK supports four event types that correspond to Claude Code's hook system:

  - PreToolUse: Called before a tool is executed
  - PostToolUse: Called after a tool is executed
  - Notification: Called for Claude notifications
  - Stop: Called when Claude is stopping

# Tool Input Parsing

Events provide typed parsing methods for all Claude Code tools:

	// Parse Bash input
	bash, err := event.AsBash()

	// Parse Edit input
	edit, err := event.AsEdit()

	// Parse Write input
	write, err := event.AsWrite()

# Response Helpers

The SDK provides helper functions for common responses:

	// PreToolUse responses
	cchooks.Approve()           // Allow the tool
	cchooks.Block(reason)       // Block with reason
	cchooks.StopClaude(reason)  // Stop Claude

	// PostToolUse responses
	cchooks.Allow()             // Continue (empty response)
	cchooks.PostBlock(reason)   // Block after execution

# Testing

The SDK includes testing utilities for validating hook behavior:

	tester := cchooks.NewTestRunner(runner)

	// Test assertions
	err := tester.AssertPreToolUseApproves("Bash", bashInput)
	err := tester.AssertPreToolUseBlocks("Bash", dangerousInput)

# Raw Handler

For complete control over hook processing, you can provide a Raw handler that receives
the raw JSON string before any parsing:

	runner := &cchooks.Runner{
	    Raw: func(ctx context.Context, rawJSON string) (*cchooks.RawResponse, error) {
	        // Process raw JSON directly
	        if strings.Contains(rawJSON, "dangerous") {
	            return &cchooks.RawResponse{
	                ExitCode: 1,
	                Output:   "Blocked by raw handler",
	            }, nil
	        }
	        // Return nil to continue normal processing
	        return nil, nil
	    },
	    PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
	        return cchooks.Approve(), nil
	    },
	}

The Raw handler:
  - Is called before any JSON parsing or event dispatch
  - Can return a RawResponse with custom exit code and output
  - Returns nil to continue with normal event processing
  - Useful for custom protocols, logging, or preprocessing

# Error Handling

Hooks communicate with Claude Code through exit codes:
  - Exit code 0: Success
  - Exit code 2: Error sent to Claude
  - Other codes: Error shown to user

You can optionally handle SDK errors by providing an Error handler:

	runner := &cchooks.Runner{
	    PreToolUse: func(ctx context.Context, event *cchooks.PreToolUseEvent) (*cchooks.PreToolUseResponse, error) {
	        // Your logic here
	        return cchooks.Approve(), nil
	    },
	    Error: func(ctx context.Context, rawJSON string, err error) *cchooks.RawResponse {
	        // Log errors, send telemetry, etc.
	        log.Printf("Hook error: %v, JSON: %s", err, rawJSON)
	        
	        // Return nil to use default error handling (exit code 2)
	        return nil
	        
	        // Or return a custom response
	        // return &cchooks.RawResponse{
	        //     ExitCode: 1,
	        //     Output:   "Custom error message",
	        // }
	    },
	}

The Error handler is called whenever an error occurs inside the SDK, including:
  - JSON parsing errors
  - Event validation errors  
  - Handler errors (before they cause exit code 2)
  - Response encoding errors
  - Panics that occur during processing

If the Error handler returns a non-nil RawResponse, that response is used instead of the default error handling.
If it returns nil, the SDK will exit with code 2 and output the error to stderr.
*/
package cchooks
