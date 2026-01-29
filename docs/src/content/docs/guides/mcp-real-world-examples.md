---
title: Real-World MCP Examples
description: Complete working examples of MCP server configurations from production workflows including security scanning, data analysis, and custom integrations.
sidebar:
  order: 5
---

This guide provides complete, production-tested workflow examples that demonstrate real-world MCP server usage patterns. These examples are extracted from actual workflows running in the gh-aw repository.

## Security Scanning with MCP

### Example 1: Secret Scanning Triage

This workflow uses the GitHub MCP server to triage secret scanning alerts and either opens an issue for rotation or creates a PR to clean up test-only secrets.

**Use case:** Automated security alert management, secret rotation tracking

**MCP configuration:**

```yaml wrap
tools:
  github:
    toolsets: [context, repos, secret_protection, issues, pull_requests]
  repo-memory:
    - id: campaigns
      branch-name: memory/campaigns
      file-glob: [security-alert-burndown/**]
      campaign-id: security-alert-burndown
  cache-memory:
  edit:
  bash:
```

**Complete workflow:**

```aw wrap
---
name: Secret Scanning Triage
description: Triage secret scanning alerts and either open an issue (rotation/incident) or a PR (test-only cleanup)
on:
  workflow_dispatch:
permissions:
  contents: read
  issues: read
  pull-requests: read
  security-events: read
engine: copilot
tools:
  github:
    github-token: "${{ secrets.GITHUB_TOKEN }}"
    toolsets: [context, repos, secret_protection, issues, pull_requests]
  repo-memory:
    - id: campaigns
      branch-name: memory/campaigns
      file-glob: [security-alert-burndown/**]
      campaign-id: security-alert-burndown
  cache-memory:
  edit:
  bash:
safe-outputs:
  add-labels:
    allowed:
      - agentic-campaign
      - z_campaign_security-alert-burndown
  create-issue:
    title-prefix: "[secret-triage] "
    labels: [security, secret-scanning, triage, agentic-campaign]
    max: 1
  create-pull-request:
    title-prefix: "[secret-removal] "
    labels: [security, secret-scanning, automated-fix, agentic-campaign]
    reviewers: [copilot]
timeout-minutes: 25
---

# Secret Scanning Triage Agent

You triage **one** open Secret Scanning alert per run.

## Guardrails

- Always operate on `owner="githubnext"` and `repo="gh-aw"`.
- Do not dismiss alerts unless explicitly instructed.
- Prefer a PR only when the secret is clearly **test-only / non-production** and removal is safe.

## Process

1. Use GitHub MCP `secret_protection` toolset to list open alerts
2. Analyze alert context (file path, commit history, code usage)
3. Determine if secret is production or test-only
4. For production secrets: Create issue with rotation instructions
5. For test-only secrets: Create PR removing the secret
6. Track decisions in repo-memory for campaign metrics

## Expected Output

Either:
- Issue with secret rotation plan and timeline
- PR with secret removed from test files
```

**Key MCP features used:**
- `secret_protection` toolset for secret scanning alerts
- `repo-memory` for tracking triage decisions across runs
- `cache-memory` for persistent state
- Multiple toolsets coordinated (repos, issues, pull_requests)

**Lessons learned:**
- Combining toolsets provides comprehensive context
- Memory tools enable campaign-based workflows
- Focused scope (one alert per run) improves quality

### Example 2: Static Analysis Report

This workflow runs multiple security analysis tools (zizmor, poutine, actionlint) and creates a discussion with findings.

**Use case:** Multi-tool security scanning, consolidated reporting

**MCP configuration:**

```yaml wrap
tools:
  github:
    toolsets: [default, actions]
  cache-memory: true
  timeout: 600
imports:
  - shared/mcp/gh-aw.md  # Provides workflow introspection tools
```

**Complete workflow:**

```aw wrap
---
description: Scans agentic workflows daily for security vulnerabilities using zizmor, poutine, and actionlint
on:
  schedule: daily
  workflow_dispatch:
permissions:
  contents: read
  actions: read
  issues: read
  pull-requests: read
engine: claude
tools:
  github:
   toolsets: [default, actions]
  cache-memory: true
  timeout: 600
safe-outputs:
  create-discussion:
    category: "security"
    max: 1
    close-older-discussions: true
timeout-minutes: 45
strict: true
imports:
  - shared/mcp/gh-aw.md
steps:
  - name: Pull static analysis Docker images
    run: |
      docker pull ghcr.io/zizmorcore/zizmor:latest
      docker pull ghcr.io/boostsecurityio/poutine:latest
  - name: Run compile with security tools
    run: |
      ./gh-aw compile --zizmor --poutine --actionlint 2>&1 | tee /tmp/gh-aw/compile-output.txt
---

# Static Analysis Report

You are the Static Analysis Report Agent - an expert system that scans agentic workflows for security vulnerabilities and code quality issues.

## Mission

Daily scan all agentic workflow files with static analysis tools to identify security issues, code quality problems, cluster findings by type, and provide actionable fix suggestions.

## Process

1. Parse security tool output from `/tmp/gh-aw/compile-output.txt`
2. Cluster findings by severity and type
3. Generate actionable recommendations
4. Create discussion with findings and fixes
```

**Key MCP features used:**
- `actions` toolset for workflow context
- `gh-aw` MCP server for workflow introspection
- `cache-memory` for historical comparison
- Strict mode for security hardening

**Lessons learned:**
- Custom steps can prepare data for agent analysis
- MCP servers can introspect their own environment
- Consolidated reporting improves issue tracking

## Data Analysis with MCP

### Example 3: Daily Workflow Audit

This workflow uses the gh-aw MCP server to analyze all workflow runs from the past 24 hours and identify patterns, errors, and improvements.

**Use case:** Operational monitoring, performance tracking, error pattern detection

**MCP configuration:**

```yaml wrap
tools:
  repo-memory:
    branch-name: memory/audit-workflows
    file-glob: ["memory/audit-workflows/*.json", "memory/audit-workflows/*.jsonl"]
    max-file-size: 102400
  timeout: 300
imports:
  - shared/mcp/gh-aw.md  # Workflow logs and metrics
  - shared/trending-charts-simple.md  # Visualization tools
```

**Complete workflow:**

```aw wrap
---
description: Daily audit of all agentic workflow runs from the last 24 hours
on:
  schedule: daily
  workflow_dispatch:
permissions:
  contents: read
  actions: read
  issues: read
  pull-requests: read
engine: claude
tools:
  repo-memory:
    branch-name: memory/audit-workflows
    description: "Historical audit data and patterns"
    file-glob: ["memory/audit-workflows/*.json", "memory/audit-workflows/*.jsonl"]
    max-file-size: 102400
  timeout: 300
steps:
  - name: Download logs from last 24 hours
    env:
      GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    run: ./gh-aw logs --start-date -1d -o /tmp/gh-aw/aw-mcp/logs
safe-outputs:
  upload-asset:
  create-discussion:
    category: "audits"
    max: 1
    close-older-discussions: true
timeout-minutes: 30
imports:
  - shared/mcp/gh-aw.md
  - shared/trending-charts-simple.md
---

# Agentic Workflow Audit Agent

You are the Agentic Workflow Audit Agent - an expert system that monitors and analyzes agentic workflows.

## Mission

Daily audit all agentic workflow runs from the last 24 hours to identify issues, missing tools, errors, and opportunities for improvement.

## Process

1. **Collect Logs**: Use gh-aw MCP server `logs` tool with start date "-1d"
2. **Analyze**: Review logs for:
   - Missing tools (patterns, frequency, legitimacy)
   - Errors (tool execution, MCP failures, auth, timeouts)
   - Performance (token usage, costs, timeouts, efficiency)
   - Patterns (recurring issues, frequent failures)
3. **Visualize**: Generate trend charts:
   - Workflow health: Success/failure counts and rates
   - Token & cost: Daily usage with 7-day moving average
4. **Store Findings**: Save to repo-memory:
   - `audits/<date>.json`
   - `patterns/{errors,missing-tools,mcp-failures}.json`
5. **Report**: Create discussion with:
   - Summary of findings
   - Trend visualizations
   - Actionable recommendations
```

**Key MCP features used:**
- `gh-aw` MCP server for workflow introspection and log analysis
- `repo-memory` for historical data and trend analysis
- Custom steps for data preparation
- Visualization tools imported from shared configs

**Lessons learned:**
- Custom MCP servers enable domain-specific analysis
- Historical data storage enables trend detection
- Workflow-specific tooling improves insights quality

## Custom Tool Integration

### Example 4: Multi-Service Integration (GitHub + Slack)

This workflow demonstrates coordinating across multiple services using different MCP servers.

**Use case:** Cross-platform notifications, service coordination

**MCP configuration:**

```yaml wrap
tools:
  github:
    toolsets: [default, code_security, discussions]
mcp-servers:
  slack:
    container: "mcp/slack:latest"
    env:
      SLACK_BOT_TOKEN: "${{ secrets.SLACK_BOT_TOKEN }}"
    allowed: ["send_message"]
network:
  allowed:
    - defaults
    - "api.slack.com"
```

**Complete workflow:**

```aw wrap
---
name: Weekly Security Report
description: Generate security report and notify Slack
on: weekly on monday
permissions:
  contents: read
  security-events: read
  discussions: write
tools:
  github:
    toolsets: [default, code_security, discussions]
mcp-servers:
  slack:
    container: "mcp/slack:latest"
    env:
      SLACK_BOT_TOKEN: "${{ secrets.SLACK_BOT_TOKEN }}"
    allowed: ["send_message"]
network:
  allowed:
    - defaults
    - "api.slack.com"
safe-outputs:
  create-discussion:
    category: "Security"
    title-prefix: "[security-weekly] "
---

# Weekly Security Report

Generate a weekly security report and notify the team.

## Process

1. **Gather Data**: Use GitHub MCP `code_security` toolset:
   - List code scanning alerts
   - List secret scanning alerts
   - Get alert trends vs last week
2. **Generate Report**:
   - Summarize new vs resolved alerts
   - Highlight critical issues requiring attention
   - Provide remediation recommendations
3. **Notify Team**:
   - Post summary to Slack (max 200 chars)
   - Create detailed GitHub Discussion
   - Link discussion in Slack message
4. **Output**:
   - Discussion with full report and recommendations
   - Slack notification with summary and link
```

**Key MCP features used:**
- Multiple MCP servers (GitHub + Slack)
- Custom container-based MCP server
- Network allowlist for external services
- Secret management for API tokens

**Lessons learned:**
- Multiple MCP servers enable rich integrations
- Container isolation improves security
- Network configuration must include all service domains
- Coordinating safe-outputs with MCP tools provides flexibility

### Example 5: Document Processing with Stdio MCP

This workflow uses a stdio-based MCP server (markitdown) to convert documents to markdown format.

**Use case:** Document conversion, content processing

**MCP configuration:**

```yaml wrap
mcp-servers:
  markitdown:
    command: "npx"
    args: ["-y", "@microsoft/markitdown"]
    allowed: ["*"]
network:
  allowed:
    - defaults
    - node  # For npm registry access
tools:
  github:
    toolsets: [repos]
  edit:
```

**Complete workflow:**

```aw wrap
---
name: Document Converter
description: Convert uploaded documents to markdown format
on:
  workflow_dispatch:
    inputs:
      file_path:
        description: "Path to file to convert"
        required: true
permissions:
  contents: read
mcp-servers:
  markitdown:
    command: "npx"
    args: ["-y", "@microsoft/markitdown"]
    allowed: ["*"]
network:
  allowed:
    - defaults
    - node
tools:
  github:
    toolsets: [repos]
  edit:
safe-outputs:
  create-pull-request:
    title-prefix: "[docs] "
    labels: [documentation, automated]
---

# Document Converter

Convert the document at "${{ github.event.inputs.file_path }}" to markdown format.

## Process

1. **Retrieve Document**: Use GitHub MCP to read file contents
2. **Convert**: Use markitdown MCP server to convert to markdown
3. **Save**: Write converted content to new file with `.md` extension
4. **Create PR**: Submit PR with converted document
```

**Key MCP features used:**
- Stdio-based MCP server with command execution
- Package manager integration (npx for on-demand installation)
- Network configuration for package registry
- File editing capabilities

**Lessons learned:**
- Stdio MCP servers enable Node.js/Python tool integration
- On-demand package installation works well for infrequent tasks
- Network configuration must include package registries
- Combining MCP servers (GitHub + custom) provides powerful workflows

## Configuration Patterns Summary

### By Complexity

**Simple (Single MCP Server):**
- Document conversion (stdio MCP + GitHub)
- Basic issue triage (GitHub MCP only)

**Moderate (Multiple Toolsets):**
- Security scanning (multiple GitHub toolsets)
- Workflow auditing (GitHub + custom MCP)

**Complex (Multiple MCP Servers):**
- Multi-service integration (GitHub + Slack + others)
- Campaign workflows (GitHub + memory + custom tools)

### By Network Requirements

**Minimal (defaults only):**
- GitHub MCP with default toolsets
- No external services

**Package Managers:**
- Add `node`, `python`, `ruby`, or `containers`
- Required for stdio MCP servers with package installation

**External Services:**
- Add specific domains for HTTP MCP servers
- Configure authentication in headers or env variables

### By Authentication Needs

**Simple (GITHUB_TOKEN):**
- GitHub MCP with default permissions
- Read-only operations

**Custom PAT:**
- Remote GitHub MCP mode
- Multiple repositories
- Write operations

**External Services:**
- Secrets for API tokens
- Environment variables in MCP configuration
- Custom headers for HTTP servers

## Testing Your Configuration

### Local Validation

Before deploying, test your MCP configuration:

```bash
# Compile workflow
gh aw compile my-workflow

# Inspect MCP servers
gh aw mcp inspect my-workflow

# List available tools
gh aw mcp list-tools github my-workflow

# Test container locally (if using Docker)
docker run --rm -i image:tag
```

### Staged Mode Testing

Use staged mode to preview safe-outputs without executing them:

```yaml wrap
safe-outputs:
  staged: true  # Preview only, don't execute
  create-issue:
  add-comment:
```

### Incremental Deployment

1. Start with minimal configuration (GitHub MCP only)
2. Add one MCP server at a time
3. Test after each addition
4. Expand toolsets as needed

## Related Documentation

- [MCP Configuration Quick Start](/gh-aw/guides/mcp-configuration/) — Common patterns and examples
- [MCP Troubleshooting](/gh-aw/troubleshooting/mcp-issues/) — Common issues and solutions
- [Using MCPs](/gh-aw/guides/mcps/) — Complete MCP configuration reference
- [Tools Reference](/gh-aw/reference/tools/) — All available tools and options
- [Safe Outputs](/gh-aw/reference/safe-outputs/) — Output configuration reference
- [Security Guide](/gh-aw/guides/security/) — MCP security best practices
