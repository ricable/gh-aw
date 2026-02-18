---
on:
  workflow_dispatch:
engine: copilot
safe-outputs:
  add-labels:
    blocked: ["~*", "\\**"]  # Block labels starting with ~ or *
    allowed: ["bug", "enhancement", "documentation", "~triage", "*admin"]  # Allowed list (but blocked patterns take precedence)
    max: 5
---

# Test Blocked Label Patterns

This is a test workflow to verify that the `blocked` pattern matching works for the `add-labels` safe output.

## Configuration

The workflow demonstrates **defense-in-depth** security:
- **Blocked patterns**: `["~*", "\\**"]` - Denies any labels starting with `~` or `*`
- **Allowed list**: Permits specific labels including `~triage` and `*admin`
- **Max count**: Limits to 5 labels per operation

## Pattern Precedence

**Blocked patterns are applied AFTER allowed list filtering**, meaning:
1. Labels are first filtered by the allowed list (if configured)
2. Then blocked patterns remove matching labels
3. Finally, the max count limit is enforced

This means even if `~triage` is in the allowed list, it will be blocked by the `~*` pattern.

## Test Case

Please add the following labels to issue #1:
- "bug" ✓ (should succeed - in allowed list, not blocked)
- "enhancement" ✓ (should succeed - in allowed list, not blocked)
- "~triage" ✗ (should be blocked by `~*` pattern despite being in allowed list)
- "*admin" ✗ (should be blocked by `\\**` pattern despite being in allowed list)
- "documentation" ✓ (should succeed - in allowed list, not blocked)
- "custom-label" ✗ (should be filtered by allowed list before reaching blocked patterns)

**Expected result**: Only "bug", "enhancement", and "documentation" should be added to the issue.

## Security Rationale

This configuration prevents agentic workflows from:
- Applying workflow trigger labels (`~*`) that could cause cascading automation
- Setting administrative labels (`*admin`, `*urgent`) reserved for maintainers

The blocked patterns provide infrastructure-level enforcement that cannot be bypassed through prompt injection attacks.
