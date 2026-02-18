---
on:
  workflow_dispatch:
engine: copilot
safe-outputs:
  add-labels:
    blocked: ["~*", "\\**"]
    max: 5
---

# Test Blocked Label Patterns

This is a test workflow to verify that the `blocked` pattern matching works for the `add-labels` safe output.

The workflow is configured to block labels starting with `~` or `*` using glob patterns.

Please add the following labels to issue #1:
- "bug" (should succeed)
- "enhancement" (should succeed)
- "~triage" (should be blocked by ~* pattern)
- "*admin" (should be blocked by \** pattern)
- "documentation" (should succeed)

Expected result: Only "bug", "enhancement", and "documentation" should be added to the issue.
