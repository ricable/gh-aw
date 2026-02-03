---
# Agentic Workflows Tool Configuration
# This shared component configures the agentic-workflows MCP tool for workflows
# that need to compile, validate, audit, or inspect agentic workflow files.
#
# Usage in workflows:
#   tools:
#     agentic-workflows:
#   imports:
#     - shared/agentic-workflows.md
#
# The tool provides these capabilities:
#   - status: Show status of workflow files
#   - compile: Compile workflows to GitHub Actions YAML
#   - logs: Download and analyze workflow logs
#   - audit: Investigate workflow runs
#   - mcp-inspect: Inspect MCP servers in workflows
#   - add: Add workflows from remote repositories
#   - update: Update workflows from their sources
#   - fix: Apply automatic codemods to workflow files
permissions:
  actions: read
tools:
  agentic-workflows:
---

The agentic-workflows tool is now available. Use it to:
- **Compile workflows**: `compile` tool compiles .md files to .lock.yml
- **Check status**: `status` tool shows workflow file status
- **Analyze logs**: `logs` tool downloads and analyzes workflow runs
- **Audit runs**: `audit` tool investigates specific workflow runs
- **Inspect MCP**: `mcp-inspect` tool shows MCP servers in workflows
- **Add workflows**: `add` tool imports workflows from remote repos
- **Update workflows**: `update` tool updates workflows from sources
- **Fix workflows**: `fix` tool applies automatic codemods

All tools return structured JSON output for easy parsing and analysis.
