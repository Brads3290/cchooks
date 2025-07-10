# CRITICAL: Document All Design Decisions
**ANY time the user states a rule, pattern, or design decision, you MUST immediately update serena memory WITHOUT BEING ASKED**

# General: Use sequential thinking
You should generally make use of sequential thinking to help you reason. Simple requests may not require this, but you should be aware that it is available when more than a simple answer is required.

## Documentation is Secondary to Tasks
**When the user provides information WHILE giving you a task:**
1. Document it quickly and quietly in the appropriate serena memory file
2. THEN IMMEDIATELY DO THE ACTUAL TASK
3. Documentation should be a brief side-action, not your main response

## Auto-Documentation Triggers
If the user's message contains any of these patterns, UPDATE SERENA MEMORY IMMEDIATELY:
- "All X should..." / "Every X must..."
- "Always..." / "Never..."
- "Make sure to..." / "Ensure that..."
- "The pattern is..." / "The convention is..."
- "Going forward..." / "From now on..."
- Any correction about how things should work
- Any statement about project standards or requirements

# Task Execution Rules

1. **Do ONLY what's asked** - Complete the exact task requested. Don't fix unrelated issues or extend the scope.
   - Example: If asked to refactor tests, don't fix failing tests unless asked.

2. **Never guess** - If something is unclear, contradictory, or complicated:
   - First check `serena memory/` for relevant information
   - If not documented and would require guessing, ASK the user
   - Be specific about what you need to know and why

3. **Analyze first** - Think through the problem before taking action.

4. **Before marking complete** - Ask yourself: "Did the user teach me something that should be documented?"
