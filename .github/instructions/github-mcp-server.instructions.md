---
applyTo: "**"
---

# GitHub MCP Server Instructions

This document provides comprehensive guidance for using the GitHub MCP (Model Context Protocol) server in agentic workflows. It covers available tools, toolset configuration, authentication, and best practices.

**Last Updated**: 2026-02-22
**MCP Server Version**: 2.0 (Remote Mode)
**Total Available Tools**: 56 across 19 toolsets

## Quick Reference

### Default Toolsets

When no `toolsets:` configuration is specified, the following are enabled by default:
- `context` - GitHub Copilot context and support docs
- `repos` - Repository operations (read)
- `issues` - Issue management (read)
- `pull_requests` - Pull request operations (read)

### Configuration Examples

```yaml
# Use defaults
tools:
  github:

# Use all toolsets
tools:
  github:
    toolsets: [all]

# Custom selection
tools:
  github:
    toolsets: [default, actions, discussions]

# Read-only remote mode
tools:
  github:
    mode: remote
    read-only: true
    toolsets: [repos, issues]
```

## Complete Tool Reference

### context toolset
*GitHub Copilot context and support documentation. No special permissions required.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `get_copilot_space` | Get details about a specific GitHub Copilot space | `space_id` |
| `github_support_docs_search` | Search GitHub support documentation | `query` |
| `list_copilot_spaces` | List available GitHub Copilot spaces | — |

### repos toolset
*Repository operations. Requires `contents` read permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `get_file_contents` | Read file or directory contents from a repository | `owner`, `repo`, `path`, `ref` |
| `get_repository_tree` | Get the file tree of a repository | `owner`, `repo`, `ref`, `recursive` |
| `list_commits` | List commits in a repository | `owner`, `repo`, `sha`, `path` |
| `get_commit` | Get details of a specific commit | `owner`, `repo`, `sha` |
| `list_branches` | List branches in a repository | `owner`, `repo` |
| `list_tags` | List tags in a repository | `owner`, `repo` |
| `get_tag` | Get details of a specific tag | `owner`, `repo`, `tag` |
| `get_latest_release` | Get the latest release for a repository | `owner`, `repo` |
| `get_release_by_tag` | Get a release by its tag name | `owner`, `repo`, `tag` |
| `list_releases` | List all releases for a repository | `owner`, `repo` |

### issues toolset
*Issue management. Requires `issues` read/write permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `issue_read` | Read issue details and comments | `owner`, `repo`, `issue_number` |
| `list_issues` | List issues in a repository | `owner`, `repo`, `state`, `labels` |
| `list_issue_types` | List available issue types for a repository | `owner`, `repo` |
| `search_issues` | Search issues (also available via `search` toolset) | `query`, `owner`, `repo` |

### pull_requests toolset
*Pull request operations. Requires `pull-requests` read/write permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `pull_request_read` | Read pull request details, reviews, and comments | `owner`, `repo`, `pull_number` |
| `list_pull_requests` | List pull requests in a repository | `owner`, `repo`, `state`, `base` |
| `search_pull_requests` | Search pull requests (also available via `search` toolset) | `query`, `owner`, `repo` |

### actions toolset
*GitHub Actions workflows and CI/CD. Requires `actions` read permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `actions_list` | List GitHub Actions workflows and workflow runs | `owner`, `repo`, `workflow_id` |
| `actions_get` | Get details of a specific workflow run | `owner`, `repo`, `run_id` |
| `get_job_logs` | Download logs for a specific workflow job | `owner`, `repo`, `job_id` |

### code_security toolset
*Code scanning alerts. Requires `security-events` read/write permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_code_scanning_alerts` | List code scanning alerts for a repository | `owner`, `repo`, `state`, `severity` |
| `get_code_scanning_alert` | Get details of a specific code scanning alert | `owner`, `repo`, `alert_number` |

### dependabot toolset
*Dependabot vulnerability alerts. Requires `security-events` read permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_dependabot_alerts` | List Dependabot alerts for a repository | `owner`, `repo`, `state`, `severity` |
| `get_dependabot_alert` | Get details of a specific Dependabot alert | `owner`, `repo`, `alert_number` |

### discussions toolset
*GitHub Discussions. Requires `discussions` read/write permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_discussions` | List discussions in a repository | `owner`, `repo`, `category_id` |
| `get_discussion` | Get details of a specific discussion | `owner`, `repo`, `discussion_number` |
| `get_discussion_comments` | Get comments for a specific discussion | `owner`, `repo`, `discussion_number` |
| `list_discussion_categories` | List discussion categories for a repository | `owner`, `repo` |

### experiments toolset
*Experimental/preview features. May be unstable. No special permissions.*

Currently no tools are available in this toolset.

### gists toolset
*GitHub Gist operations. No special permissions required.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `get_gist` | Get a specific gist by ID | `gist_id` |
| `list_gists` | List gists for a user | `username` |

### labels toolset
*Label management. Requires `issues` read/write permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `get_label` | Get details of a specific label | `owner`, `repo`, `name` |
| `list_label` | List labels in a repository | `owner`, `repo` |

### notifications toolset
*User notification management. No special permissions required.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_notifications` | List user notifications | `all`, `participating` |
| `get_notification_details` | Get details of a specific notification | `thread_id` |

### orgs toolset
*Organization security operations. No special permissions required.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_org_repository_security_advisories` | List security advisories for repositories in an organization | `org` |

### projects toolset
*GitHub Projects (classic and new). Requires a PAT — not supported by `GITHUB_TOKEN`.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_projects` | List GitHub Projects for a user or organization | `owner`, `org` |
| `get_project` | Get details of a specific project | `project_id` |
| `list_project_items` | List items (issues, PRs, notes) in a project | `project_id` |
| `get_project_item` | Get a specific project item | `project_id`, `item_id` |
| `list_project_fields` | List fields defined in a project | `project_id` |
| `get_project_field` | Get a specific project field | `project_id`, `field_id` |

### secret_protection toolset
*Secret scanning alerts. Requires `security-events` read permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_secret_scanning_alerts` | List secret scanning alerts for a repository | `owner`, `repo`, `state` |
| `get_secret_scanning_alert` | Get details of a specific secret scanning alert | `owner`, `repo`, `alert_number` |

### security_advisories toolset
*Security advisory management. Requires `security-events` read/write permission.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_global_security_advisories` | List advisories from the GitHub Advisory Database | `ecosystem`, `severity`, `cve_id` |
| `get_global_security_advisory` | Get a specific global security advisory | `ghsa_id` |
| `list_repository_security_advisories` | List security advisories for a specific repository | `owner`, `repo`, `state` |

### stargazers toolset
*Repository star information. No special permissions required.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `list_starred_repositories` | List repositories starred by a user | `username` |

### users toolset
*User profile information. Requires additional token scopes not available via `GITHUB_TOKEN`.*

No tools currently listed. Requires explicit PAT configuration.

### search toolset
*Advanced GitHub search. No special permissions required.*

| Tool | Purpose | Key Parameters |
|------|---------|----------------|
| `search_repositories` | Search for repositories | `query`, `sort`, `order` |
| `search_code` | Search code across repositories | `query`, `owner`, `repo` |
| `search_issues` | Search issues and pull requests | `query`, `sort`, `order` |
| `search_pull_requests` | Search pull requests | `query`, `sort`, `order` |
| `search_users` | Search GitHub users | `query`, `sort`, `order` |
| `search_orgs` | Search GitHub organizations | `query` |

## Toolset Configuration Reference

### Standard Configuration

```yaml
tools:
  github:
    mode: "remote"          # "remote" (default) or "local"
    toolsets: [all]         # or specific toolsets
    read-only: false        # true = read-only mode
    github-token: "..."     # optional custom token
```

### Available Toolset Values

| Value | Description |
|-------|-------------|
| `default` | Recommended defaults: context, repos, issues, pull_requests |
| `all` | All toolsets |
| `context` | GitHub Copilot context and support |
| `repos` | Repository operations |
| `issues` | Issue management |
| `pull_requests` | Pull request operations |
| `actions` | GitHub Actions workflows |
| `code_security` | Code scanning alerts |
| `dependabot` | Dependabot alerts |
| `discussions` | GitHub Discussions |
| `experiments` | Experimental features |
| `gists` | Gist operations |
| `labels` | Label management |
| `notifications` | Notification management |
| `orgs` | Organization security |
| `projects` | GitHub Projects (requires PAT) |
| `secret_protection` | Secret scanning |
| `security_advisories` | Security advisories |
| `stargazers` | Repository stars |
| `users` | User profiles (requires PAT) |
| `search` | Advanced search |

### Recommended Defaults Rationale

The default toolsets (`context`, `repos`, `issues`, `pull_requests`) are chosen because:
- **`context`**: Provides Copilot space awareness and support docs search, useful across all workflows
- **`repos`**: Core repository access is needed for almost every workflow
- **`issues`**: Issue tracking is fundamental to most development workflows
- **`pull_requests`**: PR operations are essential for code review and CI workflows

**Specialized toolsets** that should be explicitly enabled when needed:
- `actions` — For CI/CD monitoring and log analysis
- `discussions` — For community Q&A and announcement workflows
- `search` — For cross-repository search operations
- `code_security`, `dependabot`, `secret_protection` — For security audit workflows
- `security_advisories` — For advisory management
- `projects` — For project board management (requires PAT)
- `notifications` — For notification management workflows
- `labels` — For label automation
- `gists` — For gist-based workflows
- `orgs` — For organization-level security advisory listing
- `stargazers` — For star/engagement tracking
- `users` — For user profile lookups (requires PAT)
- `experiments` — For preview/experimental features

## Authentication

### Remote Mode (Recommended)

```
Authorization: Bearer <token>
X-MCP-Readonly: true  (optional, for read-only enforcement)
```

**Token priority** (first available):
1. `github-token:` field in workflow configuration
2. `GH_AW_GITHUB_TOKEN` secret
3. `GITHUB_TOKEN` (default Actions token)

### Local Mode (Docker)

Environment variables:
- `GITHUB_PERSONAL_ACCESS_TOKEN` — Required
- `GITHUB_READ_ONLY=1` — Optional read-only mode
- `GITHUB_TOOLSETS=repos,issues` — Optional toolset filter

## Token Permissions Reference

| Toolset | Required Permission | Notes |
|---------|--------------------|-|
| `context` | None | Public data only |
| `repos` | `contents: read` | Repository file access |
| `issues` | `issues: read` | Issue/comment access |
| `pull_requests` | `pull-requests: read` | PR access |
| `actions` | `actions: read` | Workflow logs access |
| `code_security` | `security-events: read` | Code scanning |
| `dependabot` | `security-events: read` | Dependabot alerts |
| `discussions` | `discussions: read` | Discussion access |
| `gists` | None (public) or PAT | Gist access |
| `labels` | `issues: read` | Label access |
| `notifications` | None | Auth user notifications |
| `orgs` | None | Public org data |
| `projects` | PAT required | Not supported by GITHUB_TOKEN |
| `secret_protection` | `security-events: read` | Secret scanning |
| `security_advisories` | `security-events: read` | Advisory access |
| `stargazers` | None | Public star data |
| `users` | PAT required | Not supported by GITHUB_TOKEN |
| `search` | None | Public search |

## Important Notes and Limitations

### Tools Not Available via GITHUB_TOKEN

The following toolsets require a Personal Access Token (PAT) and **cannot** be used with the default `GITHUB_TOKEN`:
- `projects` — GitHub Projects require project-scope PAT
- `users` — User management requires additional OAuth scopes

### Billing and Cost Data

Detailed per-run billing costs are **not available** through the GitHub API with standard permissions. The `actions` toolset provides run duration and status but not cost data. Use the GitHub billing UI or require `admin:org` PAT for billing reports.

### Write Operations

While the JSON mapping shows `write_permissions` for many toolsets, the actual write tools (create_issue, create_pull_request, etc.) are handled through the `safe-outputs` mechanism in agentic workflows to ensure proper audit trails and permission checks.

### Rate Limits

- Authenticated REST API: ~5,000 requests/hour (PAT) or lower (GITHUB_TOKEN)
- GraphQL API: Complexity-based limits
- Design workflows to paginate and minimize unnecessary requests

## Usage Examples

### Workflow: Repository Code Review
```yaml
tools:
  github:
    toolsets: [repos, pull_requests, issues]
```

### Workflow: Security Audit
```yaml
tools:
  github:
    toolsets: [code_security, dependabot, secret_protection]
    read-only: true
```

### Workflow: CI/CD Monitor
```yaml
tools:
  github:
    toolsets: [default, actions]
```

### Workflow: Community Management
```yaml
tools:
  github:
    toolsets: [default, discussions, labels]
```

### Workflow: Full Access
```yaml
tools:
  github:
    toolsets: [all]
    github-token: "${{ secrets.CUSTOM_PAT }}"
```

## References

- [GitHub MCP Server Repository](https://github.com/github/github-mcp-server)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [GitHub Actions Documentation](https://docs.github.com/actions)
- [Agentic Workflows Reference](../.github/aw/github-agentic-workflows.md)
