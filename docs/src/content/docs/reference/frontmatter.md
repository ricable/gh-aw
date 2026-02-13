---
title: Frontmatter
description: Complete guide to all available frontmatter configuration options for GitHub Agentic Workflows, including triggers, permissions, AI engines, and workflow settings.
sidebar:
  order: 200
---

The [frontmatter](/gh-aw/reference/glossary/#frontmatter) (YAML configuration section between `---` markers) of GitHub Agentic Workflows includes the triggers, permissions, AI [engines](/gh-aw/reference/glossary/#engine) (which AI model/provider to use), and workflow settings. For example:

```yaml wrap
---
on:
  issues:
    types: [opened]

tools:
  edit:
  bash: ["gh issue comment"]
---
...markdown instructions...
```

## Frontmatter Elements

The frontmatter combines standard GitHub Actions properties (`on`, `permissions`, `run-name`, `runs-on`, `timeout-minutes`, `concurrency`, `env`, `environment`, `container`, `services`, `if`, `steps`, `cache`) with GitHub Agentic Workflows-specific elements (`description`, `source`, `github-token`, `imports`, `engine`, `strict`, `roles`, `features`, `plugins`, `runtimes`, `safe-inputs`, `safe-outputs`, `network`, `tools`).

Tool configurations (such as `bash`, `edit`, `github`, `web-fetch`, `web-search`, `playwright`, `cache-memory`, and custom [Model Context Protocol](/gh-aw/reference/glossary/#mcp-model-context-protocol) (MCP) [servers](/gh-aw/reference/glossary/#mcp-server)) are specified under the `tools:` key. Custom inline tools can be defined with the [`safe-inputs:`](/gh-aw/reference/safe-inputs/) (custom tools defined inline) key. See [Tools](/gh-aw/reference/tools/) and [Safe Inputs](/gh-aw/reference/safe-inputs/) for complete documentation.

### Trigger Events (`on:`)

The `on:` section uses standard GitHub Actions syntax to define workflow triggers, with additional fields for security and approval controls:

- Standard GitHub Actions triggers (push, pull_request, issues, schedule, etc.)
- `reaction:` - Add emoji reactions to triggering items
- `stop-after:` - Automatically disable triggers after a deadline
- `manual-approval:` - Require manual approval using environment protection rules
- `forks:` - Configure fork filtering for pull_request triggers

See [Trigger Events](/gh-aw/reference/triggers/) for complete documentation.

### Description (`description:`)

Provides a human-readable description of the workflow rendered as a comment in the generated lock file.

```yaml wrap
description: "Workflow that analyzes pull requests and provides feedback"
```

### Source Tracking (`source:`)

Tracks workflow origin in format `owner/repo/path@ref`. Automatically populated when using `gh aw add` to install workflows from external repositories. Optional for manually created workflows.

```yaml wrap
source: "githubnext/agentics/workflows/ci-doctor.md@v1.0.0"
```

### Labels (`labels:`)

Optional array of strings for categorizing and organizing workflows. Displayed in `gh aw status` command output and filterable using the `--label` flag.

```yaml wrap
labels: ["automation", "ci", "diagnostics"]
```

### Metadata (`metadata:`)

Optional key-value pairs for storing custom metadata compatible with the [GitHub Copilot custom agent spec](https://docs.github.com/en/copilot/reference/custom-agents-configuration).

```yaml wrap
metadata:
  author: John Doe
  version: 1.0.0
  category: automation
```

**Constraints:** Keys: 1-64 characters, Values: max 1024 characters, string values only.

### GitHub Token (`github-token:`)

Configures the default GitHub token for engine authentication, checkout steps, and safe-output operations. Precedence (highest to lowest): individual safe-output token → safe-outputs global → top-level → default `${{ secrets.GH_AW_GITHUB_TOKEN || secrets.GITHUB_TOKEN }}`.

```yaml wrap
github-token: ${{ secrets.CUSTOM_PAT }}
```

See [GitHub Tokens](/gh-aw/reference/tokens/) for complete documentation.

### Plugins (`plugins:`)

:::caution[Experimental Feature]
Plugin support is experimental and may change in future releases. Using plugins will emit a compilation warning.
:::

Specifies plugins to install before workflow execution using engine-specific CLI commands. Each plugin repository must be specified in `org/repo` format.

**Array format:**
```yaml wrap
plugins:
  - github/test-plugin
  - acme/custom-tools
```

**Object format with custom token:**
```yaml wrap
plugins:
  repos:
    - github/test-plugin
  github-token: ${{ secrets.CUSTOM_PLUGIN_TOKEN }}
```

**Token precedence:** `plugins.github-token` → top-level `github-token` → `GH_AW_PLUGINS_TOKEN` → `GH_AW_GITHUB_TOKEN` → `GITHUB_TOKEN` (default).

### Runtimes (`runtimes:`)

Override default runtime versions for languages and tools. The compiler automatically detects requirements from tool configurations and workflow steps.

**Fields:** `version` (required), `action-repo` (optional), `action-version` (optional)

**Supported runtimes**:

| Runtime | Default Version | Default Setup Action |
|---------|----------------|---------------------|
| `node` | 24 | `actions/setup-node@v6` |
| `python` | 3.12 | `actions/setup-python@v5` |
| `go` | 1.25 | `actions/setup-go@v5` |
| `uv` | latest | `astral-sh/setup-uv@v5` |
| `bun` | 1.1 | `oven-sh/setup-bun@v2` |
| `deno` | 2.x | `denoland/setup-deno@v2` |
| `ruby` | 3.3 | `ruby/setup-ruby@v1` |
| `java` | 21 | `actions/setup-java@v4` |
| `dotnet` | 8.0 | `actions/setup-dotnet@v4` |
| `elixir` | 1.17 | `erlef/setup-beam@v1` |
| `haskell` | 9.10 | `haskell-actions/setup@v2` |

**Example:**
```yaml wrap
runtimes:
  node:
    version: "22"
  python:
    version: "3.12"
    action-repo: "actions/setup-python"
    action-version: "v5"
```

Use cases include version pinning for reproducibility, testing preview/beta versions, using custom setup actions (forks, enterprise mirrors), or overriding system defaults for compatibility. Runtimes from imported shared workflows are automatically merged with your workflow's configuration.

### Permissions (`permissions:`)

Uses standard GitHub Actions permissions syntax to specify permissions for the agentic part of workflow execution. See [GitHub Actions permissions docs](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions).

```yaml wrap
permissions:
  issues: write
  contents: read
  pull-requests: write

# Or: write-all, read-all, or {}
```

Specifying any permission sets unspecified ones to `none`.

#### Permission Validation

The compiler validates workflows have sufficient permissions for configured tools. Non-strict mode (default) emits warnings, while strict mode (`gh aw compile --strict`) treats under-provisioned permissions as errors.

### Repository Access Roles (`roles:`)

Controls who can trigger workflows based on repository permission level. Defaults to `[admin, maintainer, write]`.

```yaml wrap
roles: [admin, maintainer, write]  # Default
roles: all                         # Allow any user (⚠️ use with caution)
```

Available: `admin`, `maintainer`, `write`, `read`, `all`. Workflows with unsafe triggers automatically enforce permission checks.

### Bot Filtering (`bots:`)

Configure which GitHub bot accounts can trigger workflows.

```yaml wrap
bots:
  - "dependabot[bot]"
  - "renovate[bot]"
  - "agentic-workflows-dev[bot]"
```

When specified, only listed bot accounts can trigger the workflow. The bot must be active on the repository. Combine with `roles:` for comprehensive access control. Bot filtering is not enforced when `roles: all` is set.

### Strict Mode (`strict:`)

Enables enhanced security validation for production workflows. **Enabled by default**.

```yaml wrap
strict: true   # Enable (default)
strict: false  # Disable for development/testing
```

**Enforcement:** Refuses write permissions (use [safe-outputs](/gh-aw/reference/safe-outputs/) instead), requires explicit [network configuration](/gh-aw/reference/network/), refuses wildcard `*` in domains, requires ecosystem identifiers over individual domains, requires network config for custom MCP servers with containers, enforces GitHub Actions pinned to commit SHAs, refuses deprecated frontmatter fields.

**Configuration:** Set per-workflow via frontmatter `strict: true/false` or globally via CLI flag `gh aw compile --strict` (overrides frontmatter).

See [Network Permissions - Strict Mode Validation](/gh-aw/reference/network/#strict-mode-validation) and [CLI Commands](/gh-aw/setup/cli/#compile).

### Feature Flags (`features:`)

Enable experimental or optional features as key-value pairs.

```yaml wrap
features:
  my-experimental-feature: true
  action-mode: "script"
```

> [!NOTE]
> The `features.firewall` field has been removed. The agent sandbox is now mandatory and defaults to AWF. See [Sandbox Configuration](/gh-aw/reference/sandbox/).

#### Action Mode (`features.action-mode`)

Controls how the compiler generates custom action references. Can be `"dev"` (local paths, default), `"release"` (SHA-pinned remote paths), or `"script"` (direct shell script calls).

```yaml wrap
features:
  action-mode: "script"
```

**Script mode:** Checks out `github/gh-aw` actions folder and runs setup scripts directly instead of using `uses:` syntax. Useful for testing action scripts during development or debugging installation issues.

**Precedence:** CLI flag `--action-mode` > feature flag > environment variable `GH_AW_ACTION_MODE` > auto-detection.

### AI Engine (`engine:`)

Specifies which AI engine interprets the markdown section. See [AI Engines](/gh-aw/reference/engines/) for details.

```yaml wrap
engine: copilot
```

### Network Permissions (`network:`)

Controls network access using ecosystem identifiers and domain allowlists. See [Network Permissions](/gh-aw/reference/network/) for full documentation.

```yaml wrap
network:
  allowed:
    - defaults              # Basic infrastructure
    - python               # Python/PyPI ecosystem
    - "api.example.com"    # Custom domain
```

### Safe Inputs (`safe-inputs:`)

Enables defining custom MCP tools inline using JavaScript or shell scripts. See [Safe Inputs](/gh-aw/reference/safe-inputs/) for complete documentation on creating custom tools with controlled secret access.

### Safe Outputs (`safe-outputs:`)

Enables automatic issue creation, comment posting, and other safe outputs. See [Safe Outputs Processing](/gh-aw/reference/safe-outputs/).

### Run Configuration (`run-name:`, `runs-on:`, `timeout-minutes:`)

Standard GitHub Actions properties:
```yaml wrap
run-name: "Custom workflow run name"  # Defaults to workflow name
runs-on: ubuntu-latest               # Defaults to ubuntu-latest (main job only)
timeout-minutes: 30                  # Defaults to 20 minutes
```

> [!CAUTION]
> Breaking Change: `timeout_minutes` Removed
> The underscore variant `timeout_minutes` has been removed and is no longer supported. Use `timeout-minutes` (with hyphen) instead. Workflows using `timeout_minutes` will fail compilation with an "Unknown property" error.

### Workflow Concurrency Control (`concurrency:`)

Automatically generates concurrency policies for the agent job. See [Concurrency Control](/gh-aw/reference/concurrency/).

## Environment Variables (`env:`)

Standard GitHub Actions `env:` syntax for workflow-level environment variables:

```yaml wrap
env:
  CUSTOM_VAR: "value"
  SECRET_VAR: ${{ secrets.MY_SECRET }}
```

Environment variables can be defined at multiple scopes (workflow, job, step, engine, safe-outputs, etc.) with clear precedence rules. See [Environment Variables](/gh-aw/reference/environment-variables/) for complete documentation on all 13 env scopes and precedence order.

## Secrets (`secrets:`)

Defines secret values passed to workflow execution for MCP servers, custom engines, or workflow components. Values must be GitHub Actions expressions (e.g., `${{ secrets.API_KEY }}`).

```yaml wrap
secrets:
  API_TOKEN: ${{ secrets.API_TOKEN }}
  DATABASE_URL: ${{ secrets.DB_URL }}
```

Optional descriptions:

```yaml wrap
secrets:
  API_TOKEN:
    value: ${{ secrets.API_TOKEN }}
    description: "API token for external service"
```

**Security:** Always use secret expressions, never commit plaintext secrets, use environment-specific secrets via `environment:` field, limit secret access to components that need them.

**Note:** For reusable workflows, use `jobs.<job_id>.secrets` instead. The top-level `secrets:` field is for workflow-level configuration.

## Environment Protection (`environment:`)

Specifies the environment for deployment protection rules and environment-specific secrets. Standard GitHub Actions syntax.

```yaml wrap
environment: production
```

See [GitHub Actions environment docs](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment).

## Container Configuration (`container:`)

Specifies a container to run job steps in.

```yaml wrap
container: node:18
```

See [GitHub Actions container docs](https://docs.github.com/en/actions/how-tos/write-workflows/choose-where-workflows-run/run-jobs-in-a-container).

## Service Containers (`services:`)

Defines service containers that run alongside your job (databases, caches, etc.).

```yaml wrap
services:
  postgres:
    image: postgres:13
    env:
      POSTGRES_PASSWORD: postgres
    ports:
      - 5432:5432
```

See [GitHub Actions service docs](https://docs.github.com/en/actions/using-containerized-services).

## Conditional Execution (`if:`)

Standard GitHub Actions `if:` syntax:

```yaml wrap
if: github.event_name == 'push'
```

## Custom Steps (`steps:`)

Add custom steps before agentic execution. If unspecified, a default checkout step is added automatically.

```yaml wrap
steps:
  - name: Install dependencies
    run: npm ci
```

Use custom steps to precompute data, filter triggers, or prepare context for AI agents. See [Deterministic & Agentic Patterns](/gh-aw/guides/deterministic-agentic-patterns/).

> [!CAUTION]
> Custom steps, post-steps, and jobs run OUTSIDE the firewall sandbox with standard GitHub Actions security but WITHOUT network egress controls. Use only for deterministic operations (data preparation, preprocessing, filtering, cleanup, artifact uploads, notifications) - never for agentic compute or untrusted AI execution.

## Post-Execution Steps (`post-steps:`)

Add custom steps after agentic execution. Run after AI engine completes regardless of success/failure (unless conditional expressions are used).

```yaml wrap
post-steps:
  - name: Upload Results
    if: always()
    uses: actions/upload-artifact@v4
    with:
      name: workflow-results
      path: /tmp/gh-aw/
      retention-days: 7
```

Useful for artifact uploads, summaries, cleanup, or triggering downstream workflows.

## Custom Jobs (`jobs:`)

Define custom jobs that run before agentic execution. The agentic execution job waits for all custom jobs to complete. Custom jobs can share data through artifacts or job outputs. See [Deterministic & Agentic Patterns](/gh-aw/guides/deterministic-agentic-patterns/).

```yaml wrap
jobs:
  super_linter:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - name: Run Super-Linter
        uses: super-linter/super-linter@v7
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Job Outputs

Custom jobs can expose outputs accessible in the agentic execution prompt via `${{ needs.job-name.outputs.output-name }}`:

```yaml wrap
jobs:
  release:
    outputs:
      release_id: ${{ steps.get_release.outputs.release_id }}
      version: ${{ steps.get_release.outputs.version }}
    steps:
      - id: get_release
        run: echo "version=${{ github.event.release.tag_name }}" >> $GITHUB_OUTPUT
---

Generate highlights for release ${{ needs.release.outputs.version }}.
```

Job outputs must be string values.

## Cache Configuration (`cache:`)

Cache configuration using standard GitHub Actions `actions/cache` syntax:

Single cache:
```yaml wrap
cache:
  key: node-modules-${{ hashFiles('package-lock.json') }}
  path: node_modules
  restore-keys: |
    node-modules-
```

## Related Documentation

See also: [Trigger Events](/gh-aw/reference/triggers/), [AI Engines](/gh-aw/reference/engines/), [CLI Commands](/gh-aw/setup/cli/), [Workflow Structure](/gh-aw/reference/workflow-structure/), [Network Permissions](/gh-aw/reference/network/), [Command Triggers](/gh-aw/reference/command-triggers/), [MCPs](/gh-aw/guides/mcps/), [Tools](/gh-aw/reference/tools/), [Imports](/gh-aw/reference/imports/)
