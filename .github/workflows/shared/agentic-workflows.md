---
# Agentic Workflows Tool Configuration
#
# This shared component configures the agentic-workflows MCP tool for workflows
# that need to compile, validate, audit, or inspect agentic workflow files.
#
# ## Usage in workflows:
#
# Add to your workflow frontmatter:
#
#   tools:
#     agentic-workflows:
#   imports:
#     - shared/agentic-workflows.md
#
# ## How it works:
#
# 1. The tools.agentic-workflows field enables the gh-aw MCP server
# 2. The compiler automatically:
#    - Installs gh-aw extension via GitHub CLI (if not already installed)
#    - Copies the binary to /opt/gh-aw for containerization
#    - Configures the MCP server with stdio transport in Alpine container
#    - Mounts workspace and temp directories for file access
#    - Passes GITHUB_TOKEN for GitHub API access
#
# 3. Dev mode support:
#    - Works identically in both dev and release modes
#    - No manual configuration needed
#    - Action mode detection is automatic
#
# ## Available Tools:
#
#   - status: Show status of workflow files
#   - compile: Compile workflows to GitHub Actions YAML
#   - logs: Download and analyze workflow logs
#   - audit: Investigate workflow runs
#   - mcp-inspect: Inspect MCP servers in workflows
#   - add: Add workflows from remote repositories
#   - update: Update workflows from their sources
#   - fix: Apply automatic codemods to workflow files
#
# ## Permissions:
#
# This component adds 'actions: read' permission for accessing workflow runs.
# Your workflow may need additional permissions depending on the tools used
# (e.g., issues:read, pull-requests:read for GitHub MCP server).
#
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
