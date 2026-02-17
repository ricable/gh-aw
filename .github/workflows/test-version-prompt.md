---
name: Test Versioned Prompt
description: Example workflow demonstrating prompt versioning
version: 1.0.0
engine: codex
on:
  workflow_dispatch:
---

# Test Versioned Prompt

This is a test workflow to demonstrate the new prompt versioning feature.

The version (1.0.0) will be:
1. Displayed in the compiled workflow header as a comment
2. Included in the aw_info JSON for runtime tracking
3. Logged and available for comparison, rollback, and A/B testing
