---
name: Test Agentic Workflows Tool
description: Test workflow to validate the agentic-workflows tool configuration
on: workflow_dispatch
permissions:
  contents: read
  actions: read
  issues: read
  pull-requests: read
engine: copilot
tools:
  agentic-workflows:
  github:
    toolsets: [default]
timeout-minutes: 10
imports:
  - shared/agentic-workflows.md
---

# Test Agentic Workflows Tool

This is a test workflow to validate the agentic-workflows tool configuration using the shared component.

## Test Tasks

1. **Use the status tool** to list all workflows in the repository
2. **Verify the tool is working** by checking that the output contains workflow information

## Instructions

Use the `status` tool from the agentic-workflows MCP server to list all workflows:

```
{
  "pattern": ""
}
```

The tool should return JSON with workflow information including:
- workflow: Name of each workflow
- agent: AI engine used
- compiled: Compilation status
- status: GitHub workflow status

If the tool works correctly, report success. If it fails, report the error.
