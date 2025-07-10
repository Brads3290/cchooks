# Tool Usage Guidelines

## regex_replace Tool

### When to USE regex_replace:
- There are many edits to make that follow the same pattern
- You need to update multiple occurrences of a similar structure
- The pattern is consistent and can be captured with regex

### When NOT to use regex_replace:
- The edits you need to make do not follow the same pattern
- Different edits require different logic or transformations
- The changes are unique to each location
- In these cases, use normal Edit/MultiEdit tools instead

## Edit/MultiEdit Tools
- Use Edit for single changes
- Use MultiEdit for multiple different changes in the same file
- These are better when each edit is unique or doesn't follow a pattern