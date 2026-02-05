---
name: test-amplifier-integration
description: Test workflow to validate Amplifier shared workflow integration
engine: copilot

imports:
  - shared/mcp/amplifier.md

on:
  workflow_dispatch:
---

# Test Amplifier Integration

This is a test workflow to validate that the Amplifier shared workflow component is properly configured and can be used in agentic workflows.

Please perform the following tasks:

1. Verify that the `amplifier` command is available in your environment
2. Run a simple amplifier command to test the installation: `amplifier --version` or `amplifier --help`
3. Test that UV package manager is installed and working
4. Report the versions of both tools and confirm they are functioning correctly

If any step fails, report the error details so we can fix the shared workflow configuration.
