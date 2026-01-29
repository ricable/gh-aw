---
title: MCP Configuration Quick Start
description: Common MCP server configuration patterns and examples for quick setup of GitHub Agentic Workflows with Model Context Protocol servers.
sidebar:
  order: 3
---

This guide provides ready-to-use MCP server configuration patterns for common use cases. Use these examples as starting points for integrating MCP servers into your workflows.

## Common Configuration Patterns

### GitHub MCP Server (Default Configuration)

The simplest and most common configuration uses the GitHub MCP server with default toolsets:

```yaml wrap
tools:
  github:
    toolsets: [default]
```

**What it provides:**
- Repository operations (read files, list commits, search code)
- Issue management (create, update, list issues)
- Pull request operations (create, update, list PRs)
- Context information (teams, members)

**When to use:** Most workflows that need to interact with GitHub repositories, issues, and pull requests.

**Complete workflow example:**

```aw wrap
---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
tools:
  github:
    toolsets: [default]
safe-outputs:
  add-comment:
---

# Issue Triage Agent

Analyze issue #${{ github.event.issue.number }} and suggest appropriate labels based on the content.
```

### GitHub MCP Server (Remote Mode)

Remote mode uses the hosted GitHub MCP server for faster startup without Docker:

```yaml wrap
tools:
  github:
    mode: remote
    toolsets: [default]
```

**Advantages:**
- Faster startup (no Docker image pull required)
- Lower resource usage
- Easier setup in environments without Docker

**Requirements:**
- Set `GH_AW_GITHUB_TOKEN` secret with a Personal Access Token (PAT)
- Configure using: `gh aw secrets set GH_AW_GITHUB_TOKEN --value "<your-pat>"`

**Complete workflow example:**

```aw wrap
---
on:
  schedule:
    - cron: "0 9 * * *"
permissions:
  contents: read
  issues: read
  pull-requests: read
tools:
  github:
    mode: remote
    toolsets: [default, actions]
safe-outputs:
  create-issue:
    title-prefix: "[daily-report] "
---

# Daily Activity Report

Generate a summary of repository activity from the past 24 hours including new issues, merged PRs, and workflow runs.
```

### Custom MCP Server with Docker

Run custom MCP servers in Docker containers with environment variables and network restrictions:

```yaml wrap
mcp-servers:
  ast-grep:
    container: "mcp/ast-grep:latest"
    allowed: ["*"]

  azure:
    container: "mcr.microsoft.com/azure-sdk/azure-mcp:latest"
    entrypointArgs: ["server", "start"]
    env:
      AZURE_TENANT_ID: "${{ secrets.AZURE_TENANT_ID }}"
      AZURE_CLIENT_ID: "${{ secrets.AZURE_CLIENT_ID }}"
    allowed: ["*"]

network:
  allowed:
    - defaults
    - "*.azure.com"
```

**Key fields:**
- `container`: Docker image to run
- `entrypointArgs`: Arguments passed to the container's entrypoint
- `env`: Environment variables with secrets
- `allowed`: List of tool names or `["*"]` for all tools

**When to use:** Integration with third-party services, custom analysis tools, or containerized applications.

**Complete workflow example:**

```aw wrap
---
on:
  pull_request:
    types: [opened, synchronize]
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    toolsets: [repos, pull_requests]
mcp-servers:
  ast-grep:
    container: "mcp/ast-grep:latest"
    allowed: ["*"]
safe-outputs:
  add-comment:
---

# Code Pattern Analyzer

Analyze the code changes in PR #${{ github.event.pull_request.number }} using ast-grep to identify common patterns and potential issues.
```

### HTTP MCP Server with Authentication

Connect to remote HTTP MCP servers with custom authentication:

```yaml wrap
mcp-servers:
  custom-api:
    url: "https://api.example.com/mcp"
    headers:
      Authorization: "Bearer ${{ secrets.API_TOKEN }}"
      X-Custom-Header: "workflow-automation"
    allowed: ["search", "analyze", "report"]

network:
  allowed:
    - defaults
    - "api.example.com"
```

**Key fields:**
- `url`: HTTP endpoint for the MCP server
- `headers`: Custom headers including authentication tokens
- `allowed`: Specific tool names to enable

**When to use:** Cloud-based services, hosted MCP servers, or enterprise APIs.

**Complete workflow example:**

```aw wrap
---
on:
  workflow_dispatch:
    inputs:
      query:
        description: "Search query"
        required: true
permissions:
  contents: read
mcp-servers:
  external-search:
    url: "https://search.example.com/mcp"
    headers:
      Authorization: "Bearer ${{ secrets.SEARCH_API_KEY }}"
    allowed: ["search", "get_results"]
network:
  allowed:
    - defaults
    - "search.example.com"
---

# External Search Agent

Search for: "${{ github.event.inputs.query }}" using the external search API and summarize findings.
```

### Stdio-Based MCP Server (Command Execution)

Run MCP servers as command-line tools with stdin/stdout communication:

```yaml wrap
mcp-servers:
  markitdown:
    command: "npx"
    args: ["-y", "@microsoft/markitdown"]
    allowed: ["*"]

  python-tool:
    command: "uvx"
    args: ["--from", "git+https://github.com/org/tool", "tool-name"]
    env:
      TOOL_CONFIG: "${{ secrets.TOOL_CONFIG }}"
    allowed: ["analyze", "process"]

network:
  allowed:
    - defaults
    - node  # For npm registry access
    - python  # For PyPI access
```

**Key fields:**
- `command`: Executable to run (e.g., `npx`, `uvx`, `node`, `python`)
- `args`: Arguments passed to the command
- `env`: Environment variables with configuration
- `allowed`: List of tool names or `["*"]` for all tools

**When to use:** Node.js packages, Python tools, or local executables that implement MCP.

**Complete workflow example:**

```aw wrap
---
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
---

# Document Converter

Convert the file at path "${{ github.event.inputs.file_path }}" to markdown format using markitdown.
```

### Multi-Service Integration

Combine multiple MCP servers for complex workflows:

```yaml wrap
tools:
  github:
    toolsets: [default, code_security, discussions]
  
mcp-servers:
  slack:
    container: "mcp/slack:latest"
    env:
      SLACK_BOT_TOKEN: "${{ secrets.SLACK_BOT_TOKEN }}"
    allowed: ["send_message", "get_channel_history"]
  
  datadog:
    url: "https://api.datadoghq.com/mcp"
    headers:
      DD-API-KEY: "${{ secrets.DATADOG_API_KEY }}"
    allowed: ["query_metrics", "get_events"]

network:
  allowed:
    - defaults
    - "api.slack.com"
    - "api.datadoghq.com"
```

**When to use:** Complex workflows that need to coordinate across multiple services (e.g., security monitoring with notifications and metrics).

**Complete workflow example:**

```aw wrap
---
on:
  schedule:
    - cron: "0 0 * * 1"  # Every Monday at midnight
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

Generate a weekly security report summarizing code scanning alerts, post summary to Slack, and create a discussion with detailed findings.
```

### Registry-Based MCP Server

Use the GitHub MCP registry to simplify server configuration:

```yaml wrap
mcp-servers:
  notion:
    registry: "https://api.mcp.github.com/v0/servers/makenotion/notion-mcp-server"
    container: "ghcr.io/makenotion/notion-mcp-server:latest"
    env:
      NOTION_TOKEN: "${{ secrets.NOTION_TOKEN }}"
    allowed: ["search_pages", "get_page", "query_database"]

network:
  allowed:
    - defaults
```

**Key fields:**
- `registry`: URI to the server in the MCP registry (for documentation and tooling)
- `container` or `command`: How to execute the server
- Configuration inherits documentation from the registry

**CLI helper:** Use `gh aw mcp add` to add registry servers automatically:

```bash
gh aw mcp add my-workflow makenotion/notion-mcp-server
```

## Configuration by Use Case

### Issue and PR Automation

**Goal:** Automate issue triage, PR review, or project management.

**Recommended configuration:**

```yaml wrap
tools:
  github:
    toolsets: [default]  # or [issues, pull_requests, repos] for specific tools
safe-outputs:
  add-comment:
  add-labels:
    allowed: [bug, enhancement, documentation]
```

**Key toolsets:**
- `issues`: Issue management operations
- `pull_requests`: PR operations
- `default`: Both issues and PRs plus repos and context

### CI/CD Integration

**Goal:** Analyze workflow runs, artifacts, or build results.

**Recommended configuration:**

```yaml wrap
tools:
  github:
    toolsets: [default, actions]
permissions:
  actions: read
```

**Key toolsets:**
- `actions`: Workflow runs, artifacts, job logs
- `repos`: Repository and code access

### Security Workflows

**Goal:** Monitor and respond to security alerts.

**Recommended configuration:**

```yaml wrap
tools:
  github:
    toolsets: [default, code_security]
permissions:
  security-events: read
safe-outputs:
  create-issue:
    title-prefix: "[security] "
    labels: [security]
```

**Key toolsets:**
- `code_security`: Code scanning, secret scanning alerts
- `repos`: Access repository code for context

### Documentation and Content

**Goal:** Generate or update documentation, READMEs, or wikis.

**Recommended configuration:**

```yaml wrap
tools:
  github:
    toolsets: [repos]
  edit:
  bash:
safe-outputs:
  create-pull-request:
    title-prefix: "[docs] "
    labels: [documentation]
```

**Key tools:**
- `repos`: Read and search repository files
- `edit`: Modify files in the workspace
- `bash`: Run build or validation commands

### Data Analysis and Reporting

**Goal:** Analyze repository data, generate reports, or create visualizations.

**Recommended configuration:**

```yaml wrap
tools:
  github:
    toolsets: [default, actions, discussions]
  cache-memory:  # Store historical data
mcp-servers:
  jupyter:
    container: "mcp/jupyter:latest"
    allowed: ["*"]
```

**Key features:**
- Multiple toolsets for comprehensive data access
- `cache-memory`: Persist data across workflow runs
- Custom MCP servers for specialized analysis

## Quick Reference

### Common Toolset Combinations

| Use Case | Toolsets | Permissions |
|----------|----------|-------------|
| Issue triage | `[default]` | `contents: read`, `issues: read` |
| PR review | `[default]` | `contents: read`, `pull-requests: read` |
| CI/CD analysis | `[default, actions]` | `actions: read` |
| Security scanning | `[default, code_security]` | `security-events: read` |
| Full GitHub access | `[all]` | As needed |

### MCP Server Transport Types

| Transport | Configuration | Use Case |
|-----------|--------------|----------|
| **Docker** | `container: "image:tag"` | Containerized services, isolation |
| **Stdio** | `command: "cmd"` + `args: [...]` | Node.js packages, Python tools |
| **HTTP** | `url: "https://..."` | Cloud services, remote APIs |

### Network Configuration

Add domains to `network.allowed` for external access:

```yaml wrap
network:
  allowed:
    - defaults          # GitHub and basic infrastructure
    - node             # npm, yarn, pnpm registries
    - python           # PyPI, pip
    - containers       # Docker Hub, GHCR
    - "*.example.com"  # Custom domains (wildcards supported)
```

### Authentication Patterns

**GitHub MCP Server:**
- Token precedence: `github-token` field → `GH_AW_GITHUB_TOKEN` → `GITHUB_TOKEN`
- Remote mode requires `GH_AW_GITHUB_TOKEN` with PAT

**Custom MCP Servers:**
- Use `env:` for environment variables
- Store secrets in GitHub repository secrets
- Reference with `${{ secrets.SECRET_NAME }}`

## Debugging MCP Configuration

### Inspect MCP Servers

View all configured MCP servers and available tools:

```bash
# List all MCP servers in a workflow
gh aw mcp inspect my-workflow

# Get detailed information about a specific server
gh aw mcp inspect my-workflow --server github --verbose

# List available tools for a server
gh aw mcp list-tools github my-workflow
```

### Validate Configuration

Check for configuration errors before running:

```bash
# Compile with validation
gh aw compile my-workflow

# Strict validation
gh aw compile my-workflow --validate --strict
```

### Common Issues

**Tool not found:**
- Run `gh aw mcp inspect my-workflow` to see available tools
- Verify `toolsets` configuration or `allowed` list
- Check tool names match exactly (case-sensitive)

**Connection failures:**
- Verify URL syntax for HTTP servers
- Check network configuration includes required domains
- Confirm Docker images exist and are accessible
- Check environment variables are set correctly

**Authentication errors:**
- Verify secrets exist in repository settings
- Check token has required scopes
- For remote mode, ensure `GH_AW_GITHUB_TOKEN` is configured
- Verify custom headers for HTTP servers

## Architecture Overview

Understanding how MCP servers communicate helps troubleshoot issues and optimize configurations.

### MCP Gateway Communication Flow

The MCP gateway acts as a transparent proxy between the AI agent and MCP servers:

```
┌──────────────────┐
│   AI Agent       │
│  (Copilot CLI)   │
└────────┬─────────┘
         │ HTTP/JSON-RPC
         │ (All tool calls)
         ▼
┌──────────────────────────────────────┐
│        MCP Gateway                   │
│  • Routes requests to servers        │
│  • Manages server lifecycle          │
│  • Handles authentication            │
│  • Translates protocols              │
└─┬────────┬────────┬──────────────────┘
  │        │        │
  │ stdio  │ HTTP   │ stdio
  ▼        ▼        ▼
┌────────┐┌────────┐┌────────┐
│ Docker ││ Remote ││ Docker │
│  MCP   ││  HTTP  ││  MCP   │
│ Server ││ Server ││ Server │
└────────┘└────────┘└────────┘
```

**Key points:**
- All agent communication goes through the MCP gateway
- Gateway handles multiple server types (Docker, HTTP, stdio)
- Each server runs in isolation
- Gateway translates between protocols transparently

### Local vs Remote MCP Servers

**Local Mode (Docker-based):**

```
GitHub Actions Runner
┌─────────────────────────────────────┐
│  ┌──────────────┐                   │
│  │  AI Agent    │                   │
│  └──────┬───────┘                   │
│         │                           │
│         ▼                           │
│  ┌──────────────┐                   │
│  │ MCP Gateway  │                   │
│  └──────┬───────┘                   │
│         │                           │
│         ▼                           │
│  ┌──────────────────┐               │
│  │ Docker Container │               │
│  │ ┌──────────────┐ │               │
│  │ │  MCP Server  │ │               │
│  │ └──────────────┘ │               │
│  └──────────────────┘               │
└─────────────────────────────────────┘
```

**Advantages:**
- Version pinning (e.g., `version: "sha-09deac4"`)
- Offline operation possible
- Full control over server version

**Disadvantages:**
- Docker image pull time (slower startup)
- Higher resource usage
- Requires Docker in environment

---

**Remote Mode (Hosted):**

```
GitHub Actions Runner        Cloud Service
┌────────────────────┐      ┌─────────────────┐
│  ┌──────────────┐  │      │                 │
│  │  AI Agent    │  │      │                 │
│  └──────┬───────┘  │      │                 │
│         │          │      │                 │
│         ▼          │      │                 │
│  ┌──────────────┐  │      │  ┌───────────┐  │
│  │ MCP Gateway  │──┼─────→│  │ Hosted    │  │
│  └──────────────┘  │ HTTPS│  │ MCP       │  │
│                    │      │  │ Server    │  │
└────────────────────┘      │  └───────────┘  │
                            │                 │
                            └─────────────────┘
```

**Advantages:**
- Fast startup (no Docker pull)
- Lower resource usage
- Managed service (automatic updates)

**Disadvantages:**
- Requires `GH_AW_GITHUB_TOKEN` secret
- Network dependency
- Cannot pin specific versions

**When to choose:**
- **Local**: Air-gapped environments, version pinning required, offline testing
- **Remote**: Standard workflows, faster execution, managed services preferred

### Tool Communication Patterns

Different workflows have different communication patterns:

**Pattern 1: Simple Request-Response**

```
Agent → MCP Server → Response
      (Single tool call)
```

Example: Get repository information

**Pattern 2: Multi-Step Workflow**

```
Agent → MCP Server 1 → Response
      → MCP Server 2 → Response
      → MCP Server 1 → Response
      (Multiple related calls)
```

Example: Search issues, read details, post comment

**Pattern 3: Multi-Service Integration**

```
Agent → GitHub MCP  → Response
      → Slack MCP   → Response
      → DataDog MCP → Response
      (Coordinate across services)
```

Example: Security report that queries GitHub, posts to Slack, logs to DataDog

## Next Steps

- [Using MCPs](/gh-aw/guides/mcps/) — Complete MCP configuration reference
- [MCP Troubleshooting](/gh-aw/troubleshooting/mcp-issues/) — Common issues and solutions
- [Tools Reference](/gh-aw/reference/tools/) — All available tools and options
- [Getting Started with MCP](/gh-aw/guides/getting-started-mcp/) — Step-by-step tutorial
- [Security Guide](/gh-aw/guides/security/) — MCP security best practices
- [MCP Gateway Specification](/gh-aw/reference/mcp-gateway/) — Formal architecture specification
