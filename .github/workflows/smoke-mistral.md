---
description: Smoke test workflow that validates Mistral Vibe engine functionality by reviewing recent PRs twice daily
on: 
  schedule: every 12h
  workflow_dispatch:
  pull_request:
    types: [labeled]
    names: ["smoke"]
  reaction: "rocket"
  status-comment: true
permissions:
  contents: read
  issues: read
  pull-requests: read
name: Smoke Mistral
engine: mistral-vibe
strict: true
imports:
  - shared/gh.md
  - shared/reporting.md
network:
  allowed:
    - defaults
    - github
    - playwright
tools:
  cache-memory: true
  github:
  playwright:
    allowed_domains:
      - github.com
  edit:
  bash:
    - "*"
  serena:
    languages:
      go: {}
runtimes:
  go:
    version: "1.25"
sandbox:
  mcp:
    container: "ghcr.io/github/gh-aw-mcpg"
safe-outputs:
    add-comment:
      hide-older-comments: true
      max: 2
    create-issue:
      expires: 2h
      close-older-issues: true
    add-labels:
      allowed: [smoke-mistral]
    remove-labels:
      allowed: [smoke]
    unassign-from-user:
      allowed: [githubactionagent]
      max: 1
    hide-comment:
    messages:
      footer: "> 🚀 *Powered by Mistral Vibe via [{workflow_name}]({run_url})*"
      run-started: "🚀 Mistral Vibe awakening... [{workflow_name}]({run_url}) processing this {event_type}..."
      run-success: "✨ Mission accomplished! [{workflow_name}]({run_url}) completed successfully. 🎯"
      run-failure: "⚠️ Houston, we have a problem... [{workflow_name}]({run_url}) {status}. Retrying is recommended..."
timeout-minutes: 15
---

# Smoke Test: Mistral Vibe Engine Validation

**CRITICAL EFFICIENCY REQUIREMENTS:**
- Keep ALL outputs extremely short and concise. Use single-line responses.
- NO verbose explanations or unnecessary context.
- Minimize file reading - only read what is absolutely necessary for the task.
- Use targeted, specific queries - avoid broad searches or large data retrievals.

## Test Requirements

1. **GitHub MCP Testing**: Use GitHub MCP tools to fetch details of exactly 2 merged pull requests from ${{ github.repository }} (title and number only, no descriptions)
2. **Serena MCP Testing**: 
   - Use the Serena MCP server tool `activate_project` to initialize the workspace at `${{ github.workspace }}` and verify it succeeds (do NOT use bash to run go commands)
   - After initialization, use the `find_symbol` tool to search for symbols and verify that at least 3 symbols are found in the results
3. **Playwright Testing**: Use the playwright tools to navigate to https://github.com and verify the page title contains "GitHub" (do NOT try to install playwright - use the provided MCP tools)
4. **File Writing Testing**: Create a test file `/tmp/gh-aw/agent/smoke-test-mistral-${{ github.run_id }}.txt` with content "Smoke test passed for Mistral Vibe at $(date)" (create the directory if it doesn't exist)
5. **Bash Tool Testing**: Execute bash commands to verify file creation was successful (use `cat` to read the file back)
6. **Build gh-aw**: Run `GOCACHE=/tmp/go-cache GOMODCACHE=/tmp/go-mod make build` to verify the agent can successfully build the gh-aw project (both caches must be set to /tmp because the default cache locations are not writable). If the command fails, mark this test as ❌ and report the failure.

## Output

Add a **very brief** comment (max 5-10 lines) to the current pull request with:
- PR titles only (no descriptions)
- ✅ or ❌ for each test result
- Overall status: PASS or FAIL

If all tests pass:
- Use the `add_labels` safe-output tool to add the label `smoke-mistral` to the pull request
- Use the `remove_labels` safe-output tool to remove the label `smoke` from the pull request
- Use the `unassign_from_user` safe-output tool to unassign the user `githubactionagent` from the pull request (this is a fictitious user used for testing)
