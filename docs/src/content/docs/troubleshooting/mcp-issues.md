---
title: MCP Troubleshooting Guide
description: Comprehensive troubleshooting guide for Model Context Protocol (MCP) server configuration, connection, and runtime issues in GitHub Agentic Workflows.
sidebar:
  order: 150
---

This guide provides solutions to common MCP server configuration and runtime issues. Issues are organized by category with symptoms, causes, and solutions.

## Connection and Communication Issues

### MCP Server Connection Timeout

**Symptoms:**
- Workflow fails during MCP server initialization
- Error messages about connection timeouts
- Long delays before failure

**Common causes:**
1. Network connectivity issues
2. Slow Docker image pull
3. Server startup takes too long
4. Firewall blocking connections

**Solutions:**

**For Docker containers:**
```yaml wrap
# Pre-pull images or use smaller images
mcp-servers:
  my-server:
    container: "mcp/server:alpine"  # Use Alpine-based images for faster startup
    allowed: ["*"]

# Increase timeout if server legitimately needs more time
timeout-minutes: 15  # Default is 10
```

**For HTTP servers:**
```yaml wrap
# Verify URL is correct and accessible
mcp-servers:
  api-server:
    url: "https://api.example.com/mcp"  # Check spelling and protocol
    allowed: ["*"]

# Ensure domain is in network allowlist
network:
  allowed:
    - defaults
    - "api.example.com"  # Must match URL domain
```

**For stdio servers:**
```yaml wrap
# Ensure command is available and package can be installed
mcp-servers:
  node-tool:
    command: "npx"
    args: ["-y", "package-name"]
    allowed: ["*"]

# Add appropriate network access for package installation
network:
  allowed:
    - defaults
    - node  # For npm registry access
```

**Verification steps:**
1. Test the MCP server locally: `gh aw mcp inspect workflow-name`
2. Check Docker Hub for image availability
3. Verify network configuration includes all required domains
4. Review GitHub Actions logs for detailed error messages

### Container Not Found or Pull Errors

**Symptoms:**
- Error: "Unable to find image..."
- Error: "manifest unknown"
- Container pull failures

**Common causes:**
1. Image name typo or incorrect tag
2. Private registry without authentication
3. Image doesn't exist in registry
4. Network restrictions blocking registry access

**Solutions:**

**Verify image exists:**
```bash
# Test locally
docker pull mcp/server:latest

# Check Docker Hub or registry
# Visit https://hub.docker.com/r/mcp/server
```

**Fix image name:**
```yaml wrap
# Correct image reference
mcp-servers:
  my-server:
    container: "mcp/server:v1.0"  # Include correct tag
    allowed: ["*"]
```

**Add registry to network allowlist:**
```yaml wrap
network:
  allowed:
    - defaults
    - containers  # Adds Docker Hub, GHCR, and common registries
```

**For private registries:**
```yaml wrap
mcp-servers:
  private-server:
    container: "registry.example.com/my-server:latest"
    # Note: Private registries require additional auth setup
    allowed: ["*"]

network:
  allowed:
    - defaults
    - "registry.example.com"
```

### HTTP MCP Server Connection Refused

**Symptoms:**
- Error: "Connection refused"
- Error: "Failed to connect to host"
- HTTP request failures

**Common causes:**
1. URL syntax errors
2. Missing protocol (http:// or https://)
3. Domain not in network allowlist
4. Server not accessible from GitHub Actions

**Solutions:**

**Check URL syntax:**
```yaml wrap
# Correct: Full URL with protocol
mcp-servers:
  api-server:
    url: "https://api.example.com/mcp"  # Include https://
    allowed: ["*"]

# Incorrect: Missing protocol or malformed
# url: "api.example.com/mcp"  ❌
# url: "http://api.example.com:8080"  ⚠️ Port may be blocked
```

**Add domain to network configuration:**
```yaml wrap
network:
  allowed:
    - defaults
    - "api.example.com"  # Exact domain from URL
```

**Verify server accessibility:**
```bash
# Test from local machine
curl -I https://api.example.com/mcp

# Check if server supports MCP protocol
curl -X POST https://api.example.com/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
```

## Authentication and Permission Issues

### GitHub MCP Server Authentication Failed

**Symptoms:**
- Error: "Bad credentials"
- Error: "Resource not accessible by integration"
- 401 or 403 HTTP status codes

**Common causes:**
1. Missing or invalid GitHub token
2. Token lacks required scopes
3. Token precedence issues
4. Remote mode without PAT

**Solutions:**

**For remote mode, set GH_AW_GITHUB_TOKEN:**
```bash
# Create a PAT with required scopes at https://github.com/settings/tokens
# Required scopes: repo (full control)
gh aw secrets set GH_AW_GITHUB_TOKEN --value "ghp_your_token_here"
```

**Configure custom token:**
```yaml wrap
tools:
  github:
    mode: remote
    github-token: "${{ secrets.CUSTOM_PAT }}"  # Use custom token
    toolsets: [default]
```

**Token precedence (highest to lowest):**
1. `github-token` field in configuration
2. `GH_AW_GITHUB_TOKEN` secret
3. `GITHUB_TOKEN` (default, limited permissions)

**Verify token scopes:**
```bash
# Check token scopes
gh auth status

# For PAT, ensure these scopes are enabled:
# - repo (all sub-scopes)
# - read:org (if accessing organization resources)
```

### Custom MCP Server Authentication Errors

**Symptoms:**
- Error: "Unauthorized"
- Error: "Invalid API key"
- Authentication-related failures

**Common causes:**
1. Missing or incorrect environment variables
2. Secrets not configured in repository
3. Header syntax errors
4. Token format issues

**Solutions:**

**For stdio/container servers:**
```yaml wrap
mcp-servers:
  my-server:
    container: "mcp/server:latest"
    env:
      API_KEY: "${{ secrets.MY_API_KEY }}"  # Reference secret correctly
      AUTH_TOKEN: "${{ secrets.AUTH_TOKEN }}"
    allowed: ["*"]
```

**For HTTP servers:**
```yaml wrap
mcp-servers:
  api-server:
    url: "https://api.example.com/mcp"
    headers:
      Authorization: "Bearer ${{ secrets.API_TOKEN }}"  # Bearer token format
      X-API-Key: "${{ secrets.API_KEY }}"  # Custom header
    allowed: ["*"]
```

**Verify secrets are configured:**
1. Go to repository Settings → Secrets and variables → Actions
2. Ensure secret name matches exactly (case-sensitive)
3. Test secret value is correct and not expired

**Common secret format errors:**
```yaml wrap
# ✅ Correct
env:
  TOKEN: "${{ secrets.API_TOKEN }}"

# ❌ Incorrect - missing quotes
env:
  TOKEN: ${{ secrets.API_TOKEN }}

# ❌ Incorrect - wrong secret name
env:
  TOKEN: "${{ secrets.APITOKEN }}"  # Should be API_TOKEN
```

## Configuration and Validation Issues

### Tool Not Found After Configuration

**Symptoms:**
- Agent reports tool is not available
- Tool calls fail with "unknown tool" errors
- Expected functionality doesn't work

**Common causes:**
1. Tool name not in `allowed` list
2. Wrong toolset selected
3. Tool name case mismatch
4. Tool deprecated or renamed

**Solutions:**

**For GitHub tools, use toolsets:**
```yaml wrap
# ✅ Recommended - use toolsets
tools:
  github:
    toolsets: [default, actions]  # Automatically includes all related tools

# ⚠️ Not recommended - individual tools may be renamed
tools:
  github:
    allowed: [list_workflow_runs, get_workflow_run]  # May break on updates
```

**Inspect available tools:**
```bash
# List all available tools
gh aw mcp inspect my-workflow

# Get tools for specific server
gh aw mcp inspect my-workflow --server github --verbose

# List tools by name
gh aw mcp list-tools github my-workflow
```

**For custom MCP servers, verify allowed list:**
```yaml wrap
mcp-servers:
  my-server:
    container: "mcp/server:latest"
    allowed: ["search", "analyze"]  # Tool names must match exactly

# Use wildcard to allow all tools during development
mcp-servers:
  my-server:
    container: "mcp/server:latest"
    allowed: ["*"]  # All tools available
```

**Check tool name case:**
```yaml wrap
# Tool names are case-sensitive
allowed: ["list_issues"]  # ✅ Correct
allowed: ["List_Issues"]  # ❌ Wrong case
allowed: ["listIssues"]   # ❌ Wrong format
```

### Invalid Toolset Name

**Symptoms:**
- Compilation error about invalid toolset
- Error message: "invalid toolset: 'name' is not a valid toolset"

**Common causes:**
1. Toolset name typo
2. Using deprecated toolset name
3. Toolset doesn't exist

**Solution:**

Use valid toolset names:
```yaml wrap
# ✅ Valid toolsets
tools:
  github:
    toolsets: [default]  # context, repos, issues, pull_requests (action-friendly)
    # or
    toolsets: [repos, issues, pull_requests, actions]
    # or
    toolsets: [all]  # All available toolsets
```

**Complete list of valid toolsets:**
- `context` - User and team information
- `repos` - Repository operations
- `issues` - Issue management
- `pull_requests` - Pull request operations
- `users` - User profile operations
- `actions` - Workflow runs and artifacts
- `code_security` - Security scanning alerts
- `discussions` - GitHub Discussions
- `labels` - Label management
- `notifications` - Notification operations
- `orgs` - Organization operations
- `projects` - GitHub Projects
- `gists` - Gist operations
- `search` - Search operations
- `dependabot` - Dependabot alerts
- `experiments` - Experimental features
- `secret_protection` - Secret scanning
- `security_advisories` - Security advisories
- `stargazers` - Repository stars
- `default` - Common toolsets (context, repos, issues, pull_requests)
- `all` - All available toolsets

### Toolsets and Allowed Conflict

**Symptoms:**
- Fewer tools available than expected
- Tools from toolset not working
- Unexpected tool filtering

**Cause:**
When both `toolsets:` and `allowed:` are specified, `allowed:` acts as a filter, restricting tools to only those listed within the enabled toolsets.

**Solution:**

**Recommended approach:**
```yaml wrap
# Use only toolsets - simplest and most maintainable
tools:
  github:
    toolsets: [issues, pull_requests]
```

**Advanced filtering (rarely needed):**
```yaml wrap
# Enable toolset, then restrict specific tools
tools:
  github:
    toolsets: [issues]  # Enable all issue tools
    allowed: [create_issue, update_issue]  # But only allow these two
```

**Migration from allowed to toolsets:**
```yaml wrap
# Before (not recommended)
tools:
  github:
    allowed: [get_repository, list_issues, create_pull_request]

# After (recommended)
tools:
  github:
    toolsets: [default]  # Includes all of the above and more
```

### MCP Server Type Detection Failed

**Symptoms:**
- Error: "unable to determine MCP type"
- Error: "missing type, url, command, or container"

**Cause:**
MCP server configuration missing required fields to determine transport type.

**Solution:**

Specify at least one transport field:
```yaml wrap
# Docker container transport
mcp-servers:
  server1:
    container: "mcp/server:latest"  # ✅ Valid
    allowed: ["*"]

# Stdio transport
mcp-servers:
  server2:
    command: "npx"  # ✅ Valid
    args: ["-y", "package"]
    allowed: ["*"]

# HTTP transport
mcp-servers:
  server3:
    url: "https://api.example.com/mcp"  # ✅ Valid
    allowed: ["*"]

# ❌ Invalid - no transport specified
mcp-servers:
  server4:
    env:
      KEY: "value"
    allowed: ["*"]
```

### Cannot Specify Both Container and Command

**Symptoms:**
- Error: "mcp configuration cannot specify both 'container' and 'command'"

**Cause:**
MCP server configuration includes both `container:` and `command:` fields, which are mutually exclusive.

**Solution:**

Choose one transport method:
```yaml wrap
# ✅ Use container
mcp-servers:
  my-server:
    container: "mcp/server:latest"
    allowed: ["*"]

# OR

# ✅ Use command
mcp-servers:
  my-server:
    command: "node"
    args: ["server.js"]
    allowed: ["*"]

# ❌ Don't use both
mcp-servers:
  my-server:
    container: "mcp/server:latest"
    command: "node"  # This causes an error
    allowed: ["*"]
```

## Network and Access Issues

### Domain Not in Network Allowlist

**Symptoms:**
- URLs appearing as "(redacted)" in output
- Connection refused to external services
- Firewall denials in logs

**Cause:**
Domain is not included in `network.allowed` configuration.

**Solution:**

Add domain to allowlist:
```yaml wrap
network:
  allowed:
    - defaults  # Basic GitHub infrastructure
    - "api.example.com"  # Your domain
    - "*.cdn.example.com"  # Wildcard for subdomains
```

**Common ecosystem identifiers:**
```yaml wrap
network:
  allowed:
    - defaults     # GitHub, api.github.com, githubusercontent.com
    - node         # npm, yarn, pnpm registries
    - python       # PyPI, pip
    - containers   # Docker Hub, GHCR
    - go           # Go modules proxy
```

**For MCP servers needing network access:**
```yaml wrap
tools:
  github:
    mode: remote
    toolsets: [default]

mcp-servers:
  external-api:
    url: "https://api.service.com/mcp"
    allowed: ["*"]

network:
  allowed:
    - defaults
    - "api.service.com"  # Domain for MCP server
    - "cdn.service.com"  # Additional domains the server might access
```

### Package Installation Failures

**Symptoms:**
- Error: "Failed to install package"
- NPM, pip, or other package manager errors
- Registry connection timeouts

**Cause:**
Package registries not in network allowlist.

**Solution:**

Add ecosystem identifier:
```yaml wrap
mcp-servers:
  node-tool:
    command: "npx"
    args: ["-y", "package-name"]
    allowed: ["*"]

network:
  allowed:
    - defaults
    - node  # Adds registry.npmjs.org and related domains
```

**Ecosystem identifiers by language:**

| Language/Tool | Identifier | Registries Included |
|--------------|------------|---------------------|
| Node.js | `node` | npmjs.org, yarnpkg.com, pnpm.io |
| Python | `python` | pypi.org, pythonhosted.org |
| Go | `go` | proxy.golang.org, sum.golang.org |
| Ruby | `ruby` | rubygems.org |
| Containers | `containers` | Docker Hub, GHCR, Quay.io |

### Private Registry Access

**Symptoms:**
- Cannot pull from private Docker registry
- Authentication required errors
- Registry not found

**Cause:**
Private registries require authentication and explicit network access.

**Solution:**

```yaml wrap
mcp-servers:
  private-tool:
    container: "registry.company.com/tool:latest"
    env:
      # Note: Docker authentication in GitHub Actions requires additional setup
      # See: https://docs.github.com/en/actions/publishing-packages/publishing-docker-images
      REGISTRY_TOKEN: "${{ secrets.REGISTRY_TOKEN }}"
    allowed: ["*"]

network:
  allowed:
    - defaults
    - "registry.company.com"  # Private registry domain
```

**Additional steps for private Docker registries:**
1. Add a pre-step to authenticate Docker: `docker login registry.company.com -u ${{ secrets.REGISTRY_USER }} -p ${{ secrets.REGISTRY_TOKEN }}`
2. Use GitHub Container Registry (GHCR) when possible for easier authentication with `GITHUB_TOKEN`
3. Ensure registry tokens have read permissions for the images
4. See [GitHub's Docker publishing guide](https://docs.github.com/en/actions/publishing-packages/publishing-docker-images) for detailed authentication setup

## Performance and Timeout Issues

### Slow MCP Server Startup

**Symptoms:**
- Long delays before workflow starts
- Timeout warnings in logs
- Workflow taking longer than expected

**Common causes:**
1. Large Docker images
2. Cold start for HTTP servers
3. Package installation during startup
4. Network latency

**Solutions:**

**Use smaller Docker images:**
```yaml wrap
# Prefer Alpine-based images
mcp-servers:
  my-server:
    container: "mcp/server:alpine"  # Alpine is much smaller
    allowed: ["*"]
```

**Pre-install packages:**
```yaml wrap
# For stdio servers, use lockfiles to cache dependencies
mcp-servers:
  node-tool:
    command: "npx"
    args: ["-y", "package-name@1.0.0"]  # Pin version for consistency
    allowed: ["*"]

cache:
  paths:
    - ~/.npm  # Cache npm packages
  key: mcp-node-${{ hashFiles('workflow.md') }}
```

**Use remote mode for GitHub MCP:**
```yaml wrap
# Remote mode is faster than local Docker
tools:
  github:
    mode: remote  # No Docker image pull required
    toolsets: [default]
```

**Increase timeout if legitimately needed:**
```yaml wrap
timeout-minutes: 20  # Default is 10 minutes
```

### MCP Server Memory or Resource Issues

**Symptoms:**
- Out of memory errors
- Container killed unexpectedly
- Performance degradation

**Cause:**
MCP server consuming too many resources.

**Solutions:**

**For Docker containers:**

Docker resource limits are not directly configurable in workflow frontmatter. Consider these approaches:

1. **Optimize the MCP server**: Reduce memory usage in the server implementation
2. **Use smaller datasets**: Limit the amount of data processed in a single workflow run
3. **Batch processing**: Split large operations into multiple workflow runs
4. **Container optimization**: Use Alpine-based images which have lower memory footprint

```yaml wrap
mcp-servers:
  resource-heavy:
    container: "mcp/server:alpine"  # Use Alpine for lower memory usage
    allowed: ["*"]
```

**For stdio servers:**
```yaml wrap
# Limit concurrent operations or batch size in your prompt
# Example: "Process files in batches of 10"
```

**Workflow-level timeout:**
```yaml wrap
timeout-minutes: 30  # Prevent runaway processes
```

## Debugging and Investigation

### Enable Debug Logging

Add verbose output to understand what's happening:

```yaml wrap
# Workflow-level debugging
env:
  ACTIONS_STEP_DEBUG: true
  ACTIONS_RUNNER_DEBUG: true
```

Check workflow logs:
```bash
# Download and view logs
gh aw logs my-workflow

# Audit specific run
gh aw audit RUN_ID
```

### Inspect MCP Configuration

Use CLI tools to verify configuration:

```bash
# View all MCP servers
gh aw mcp inspect my-workflow

# Get detailed server information
gh aw mcp inspect my-workflow --server github --verbose

# List available tools
gh aw mcp list-tools github my-workflow

# Validate configuration without running
gh aw compile my-workflow --validate
```

### Test MCP Servers Locally

Test servers outside of GitHub Actions:

**For Docker containers:**
```bash
# Test container runs
docker run --rm -i mcp/server:latest

# Test with environment variables
docker run --rm -i -e API_KEY=test mcp/server:latest
```

**For stdio servers:**
```bash
# Test command executes
npx -y package-name --version

# Test with stdio communication
echo '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}' | npx -y package-name
```

**For HTTP servers:**
```bash
# Test server is accessible
curl -I https://api.example.com/mcp

# Test MCP initialize call
curl -X POST https://api.example.com/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'
```

### Use MCP Debug Mode

Import the MCP debug shared configuration:

```yaml wrap
imports:
  - shared/mcp-debug.md

# Provides diagnostic tools and report_diagnostics_to_pull_request safe-output
```

### Common Debugging Workflow

1. **Verify configuration compiles:**
   ```bash
   gh aw compile my-workflow
   ```

2. **Inspect MCP servers:**
   ```bash
   gh aw mcp inspect my-workflow
   ```

3. **Test locally if possible:**
   ```bash
   # Docker: docker run --rm -i image:tag
   # Stdio: npx -y package-name
   # HTTP: curl -X POST url
   ```

4. **Check workflow logs:**
   ```bash
   gh aw logs my-workflow
   ```

5. **Enable debug mode:**
   ```yaml
   env:
     ACTIONS_STEP_DEBUG: true
   ```

6. **Review specific error messages** in this guide

7. **Check related documentation:**
   - [MCP Configuration Guide](/gh-aw/guides/mcp-configuration/)
   - [Common Issues](/gh-aw/troubleshooting/common-issues/)
   - [Error Reference](/gh-aw/troubleshooting/errors/)

## Quick Reference

### Pre-flight Checklist

Before deploying a workflow with MCP servers:

- [ ] Configuration compiles without errors (`gh aw compile`)
- [ ] MCP servers are inspectable (`gh aw mcp inspect`)
- [ ] Required secrets are configured in repository settings
- [ ] Network allowlist includes all required domains
- [ ] Docker images exist and are accessible
- [ ] Toolsets or allowed lists are correctly specified
- [ ] Authentication is properly configured
- [ ] Timeout is appropriate for the workflow

### Most Common Issues

1. **Tool not found:** Use `toolsets:` instead of `allowed:` for GitHub tools
2. **Authentication failed:** Set `GH_AW_GITHUB_TOKEN` for remote mode
3. **Connection refused:** Add domain to `network.allowed`
4. **Container not found:** Check image name and tag, add to network allowlist
5. **Timeout:** Use remote mode for GitHub MCP, increase `timeout-minutes`

### When to Use Each Mode

| Requirement | Recommended Configuration |
|------------|---------------------------|
| Fast startup | `mode: remote` for GitHub MCP |
| Offline/air-gapped | `mode: local` with pre-pulled images |
| Custom version | `mode: local` with specific version tag |
| Enterprise restrictions | `mode: local` with private registry |

## Related Documentation

- [MCP Configuration Quick Start](/gh-aw/guides/mcp-configuration/) — Common patterns and examples
- [Using MCPs](/gh-aw/guides/mcps/) — Complete MCP configuration reference
- [Getting Started with MCP](/gh-aw/guides/getting-started-mcp/) — Step-by-step tutorial
- [Tools Reference](/gh-aw/reference/tools/) — All available tools and options
- [Network Configuration](/gh-aw/guides/network-configuration/) — Network access control
- [Common Issues](/gh-aw/troubleshooting/common-issues/) — General troubleshooting
- [Error Reference](/gh-aw/troubleshooting/errors/) — Detailed error explanations
- [Security Guide](/gh-aw/guides/security/) — MCP security best practices

## Getting Help

If you encounter an issue not covered in this guide:

1. Search this documentation (Ctrl+F / Cmd+F)
2. Check [Common Issues](/gh-aw/troubleshooting/common-issues/)
3. Review workflow examples in [MCP Configuration](/gh-aw/guides/mcp-configuration/)
4. Enable verbose mode: `gh aw compile --verbose`
5. Search [existing issues](https://github.com/githubnext/gh-aw/issues)
6. Create a new issue with:
   - Workflow configuration (sanitized of secrets)
   - Error messages from logs
   - Steps to reproduce
   - Expected vs actual behavior
