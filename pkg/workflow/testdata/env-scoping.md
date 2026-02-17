---
name: Test Env Scoping
on: workflow_dispatch
engine: copilot
env:
  TEST_VAR: "test_value"
  ANOTHER_VAR: "another_value"
---

# Test Workflow

This workflow tests that env variables are scoped to the agent job only.
