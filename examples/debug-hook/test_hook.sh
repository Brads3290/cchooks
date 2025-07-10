#!/bin/bash

echo "Testing debug hook with various inputs..."

# Build the hook
echo "Building debug hook..."
go build -o debug-hook main.go

# Test 1: Valid Stop event
echo -e "\n=== Test 1: Valid Stop event ==="
echo '{"event": "Stop", "session_id": "test-123", "stop_hook_active": true, "transcript": []}' | ./debug-hook

# Test 2: Empty JSON
echo -e "\n=== Test 2: Empty JSON ==="
echo '{}' | ./debug-hook

# Test 3: Missing event field
echo -e "\n=== Test 3: Missing event field ==="
echo '{"session_id": "test-123"}' | ./debug-hook

# Test 4: Invalid event type
echo -e "\n=== Test 4: Invalid event type ==="
echo '{"event": "InvalidEvent", "session_id": "test-123"}' | ./debug-hook

# Test 5: Malformed JSON
echo -e "\n=== Test 5: Malformed JSON ==="
echo '{invalid json' | ./debug-hook

# Test 6: Empty input
echo -e "\n=== Test 6: Empty input ==="
echo '' | ./debug-hook

echo -e "\nAll tests completed."