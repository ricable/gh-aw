---
description: Smoke test workflow that validates OpenClaw engine functionality
on:
  schedule: daily
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
  discussions: read
  actions: read

name: Smoke OpenClaw
engine:
  id: openclaw
strict: true
features:
  experimental-engines: true
imports:
  - shared/gh.md
  - shared/github-queries-safe-input.md
sandbox:
  mcp:
    container: "ghcr.io/github/gh-aw-mcpg"
network:
  allowed:
    - defaults
    - github
tools:
  github:
    toolsets: [repos, pull_requests]
  bash:
    - "*"
safe-outputs:
  add-comment:
    hide-older-comments: true
    max: 2
  create-issue:
    expires: 2h
    group: true
    close-older-issues: true
  add-labels:
    allowed: [smoke-openclaw]
  messages:
    footer: "> 🦀 *[Mission Complete] — Powered by [{workflow_name}]({run_url})*"
    run-started: "🦀 **DEPLOY!** [{workflow_name}]({run_url}) activating for this {event_type}! *[Claw engaged...]*"
    run-success: "🎯 **TARGET ACQUIRED** — [{workflow_name}]({run_url}) **ALL CLEAR!** Systems nominal! ✨"
    run-failure: "⚠️ **RETREAT...** [{workflow_name}]({run_url}) {status}! The claw missed its target..."
timeout-minutes: 15
---

# Smoke Test: OpenClaw Engine Validation

**IMPORTANT: Keep all outputs extremely short and concise. Use single-line responses where possible. No verbose explanations.**

## Test Requirements

1. **GitHub MCP Testing**: Review the last 2 merged pull requests in ${{ github.repository }}
2. **Safe Inputs GH CLI Testing**: Use the `safeinputs-gh` tool to query 2 pull requests from ${{ github.repository }} (use args: "pr list --repo ${{ github.repository }} --limit 2 --json number,title,author")
3. **File Writing Testing**: Create a test file `/tmp/gh-aw/agent/smoke-test-openclaw-${{ github.run_id }}.txt` with content "Smoke test passed for OpenClaw at $(date)" (create the directory if it doesn't exist)
4. **Bash Tool Testing**: Execute bash commands to verify file creation was successful (use `cat` to read the file back)
5. **Discussion Interaction Testing**:
   - Use the `github-discussion-query` safe-input tool with params: `limit=1, jq=".[0]"` to get the latest discussion from ${{ github.repository }}
   - Extract the discussion number from the result (e.g., if the result is `{"number": 123, "title": "...", ...}`, extract 123)
   - Use the `add_comment` tool with `discussion_number: <extracted_number>` to add a brief comment stating that the OpenClaw smoke test agent was here

## Output

1. **Create an issue** with a summary of the smoke test run:
   - Title: "Smoke Test: OpenClaw - ${{ github.run_id }}"
   - Body should include:
     - Test results (✅ or ❌ for each test)
     - Overall status: PASS or FAIL
     - Run URL: ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}
     - Timestamp

2. Add a **very brief** comment (max 5-10 lines) to the current pull request with:
   - ✅ or ❌ for each test result
   - Overall status: PASS or FAIL

3. Use the `add_comment` tool to add a brief comment to the latest discussion (using the `discussion_number` you extracted in step 5)

If all tests pass, add the label `smoke-openclaw` to the pull request.
