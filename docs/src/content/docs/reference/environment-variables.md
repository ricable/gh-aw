---
title: Environment Variables
description: Complete guide to environment variable precedence and merge behavior across all workflow scopes
sidebar:
  order: 650
---

Environment variables in GitHub Agentic Workflows can be defined at multiple scopes, each serving a specific purpose in the workflow lifecycle. Variables defined at more specific scopes override those at more general scopes, following GitHub Actions conventions while adding AWF-specific contexts.

## Environment Variable Scopes

GitHub Agentic Workflows supports environment variables in 13 distinct contexts:

| Scope | Syntax | Context | Typical Use |
|-------|--------|---------|-------------|
| **Frontmatter `env:`** | `env:` | Agent job only | Agent job configuration |
| **Job-level** | `jobs.<job_id>.env` | All steps in job | Job-specific config |
| **Step-level** | `steps[*].env` | Single step | Step-specific config |
| **Engine** | `engine.env` | AI engine | Engine secrets, timeouts |
| **Container** | `container.env` | Container runtime | Container settings |
| **Services** | `services.<id>.env` | Service containers | Database credentials |
| **Sandbox Agent** | `sandbox.agent.env` | Sandbox runtime | Sandbox configuration |
| **Sandbox MCP** | `sandbox.mcp.env` | Model Context Protocol (MCP) gateway | MCP debugging |
| **MCP Tools** | `tools.<name>.env` | MCP server process | MCP server secrets |
| **Safe Inputs** | `safe-inputs.<name>.env` | Safe-input execution | Tool-specific tokens |
| **Safe Outputs Global** | `safe-outputs.env` | All safe-output jobs | Shared safe-output config |
| **Safe Outputs Job** | `safe-outputs.jobs.<name>.env` | Specific safe-output job | Job-specific config |
| **GitHub Actions Step** | `githubActionsStep.env` | Pre-defined steps | Step configuration |

### Example Configurations

**Frontmatter `env:` (agent job only):**
```yaml wrap
---
env:
  NODE_ENV: production
  API_ENDPOINT: https://api.example.com
---
```

> [!NOTE]
> The frontmatter `env:` section applies **only to the agent job**, not to all jobs in the workflow. To set environment variables for custom jobs, use `jobs.<job_id>.env`. To set environment variables for safe-output jobs, use `safe-outputs.env` or `safe-outputs.jobs.<job_name>.env`.

**Job-specific overrides:**
```yaml wrap
---
jobs:
  validation:
    env:
      VALIDATION_MODE: strict
    steps:
      - run: npm run build
        env:
          BUILD_ENV: production  # Overrides job and workflow levels
---
```

**AWF-specific contexts:**
```yaml wrap
---
# Engine configuration
engine:
  id: copilot
  env:
    OPENAI_API_KEY: ${{ secrets.CUSTOM_KEY }}

# MCP server with secrets
tools:
  database:
    command: npx
    args: ["-y", "mcp-server-postgres"]
    env:
      DATABASE_URL: ${{ secrets.DATABASE_URL }}

# Safe outputs with custom PAT
safe-outputs:
  create-issue:
  env:
    GITHUB_TOKEN: ${{ secrets.CUSTOM_PAT }}
---
```

## Precedence Rules

Environment variables follow a **most-specific-wins** model, consistent with GitHub Actions. Variables at more specific scopes completely override variables with the same name at less specific scopes.

### General Precedence (Highest to Lowest)

1. **Step-level** (`steps[*].env`, `githubActionsStep.env`)
2. **Job-level** (`jobs.<job_id>.env`)
3. **Frontmatter `env:`** (applies to agent job only)

> [!NOTE]
> The frontmatter `env:` section is **not** workflow-level. It applies only to the agent job. Custom jobs must define their own environment variables using `jobs.<job_id>.env`.

### Safe Outputs Precedence

1. **Job-specific** (`safe-outputs.jobs.<job_name>.env`)
2. **Global** (`safe-outputs.env`)
3. **Frontmatter `env:`** (if the safe-output job inherits from agent job)

### Context-Specific Scopes

These scopes are independent and operate in different contexts: `engine.env`, `container.env`, `services.<id>.env`, `sandbox.agent.env`, `sandbox.mcp.env`, `tools.<tool>.env`, `safe-inputs.<tool>.env`.

### Override Example

```yaml wrap
---
env:
  API_KEY: agent-key      # Applied to agent job only
  DEBUG: "false"

jobs:
  test:
    env:
      API_KEY: test-key    # Custom job defines its own env
      EXTRA: "value"
    steps:
      - run: |
          # In 'test' job:
          # API_KEY = "test-key" (job-level)
          # EXTRA = "value" (job-level)
          # DEBUG is NOT available (frontmatter env not inherited)

# In agent job:
# API_KEY = "agent-key" (from frontmatter env)
# DEBUG = "false" (from frontmatter env)
---
```

## Common Patterns

**Agent job configuration:**
```yaml wrap
---
env:
  NODE_ENV: production  # Applied to agent job only
---
```

**Custom jobs define their own env:**
```yaml wrap
---
env:
  AGENT_VAR: value  # Agent job only

jobs:
  test:
    env:
      NODE_ENV: test  # Test job environment
---
```

**Safe outputs with custom PAT:**
```yaml wrap
---
safe-outputs:
  create-issue:
  env:
    GITHUB_TOKEN: ${{ secrets.CUSTOM_PAT }}
---
```

**Engine and MCP configuration:**
```yaml wrap
---
engine:
  env:
    OPENAI_API_KEY: ${{ secrets.OPENAI_KEY }}

tools:
  database:
    command: npx
    args: ["-y", "mcp-server-postgres"]
    env:
      DATABASE_URL: ${{ secrets.DATABASE_URL }}
---
```

## Best Practices

**Always use secrets for sensitive data:**
```yaml wrap
# ✅ Correct
env:
  API_KEY: ${{ secrets.API_KEY }}

# ❌ Never hardcode secrets
env:
  API_KEY: "sk-1234567890abcdef"
```

**Define variables at the narrowest scope needed:**
```yaml wrap
# ✅ Job-specific variable
jobs:
  build:
    env:
      BUILD_MODE: production
```

**Use consistent naming conventions:**
- `SCREAMING_SNAKE_CASE` format
- Descriptive names: `API_KEY` not `KEY`
- Service prefixes: `POSTGRES_PASSWORD`, `REDIS_PORT`

## GitHub Actions Integration

During compilation, AWF extracts environment variables from frontmatter, preserves GitHub Actions expressions (`${{ ... }}`), and renders them to the appropriate scope in `.lock.yml` files. Secret syntax is validated to ensure `${{ secrets.NAME }}` format.

**Generated lock file structure:**
```yaml
# No workflow-level env section

jobs:
  agent:
    env:
      # Frontmatter env variables merged with system env
      CUSTOM_VAR: ${{ secrets.CUSTOM_SECRET }}
      GH_AW_SAFE_OUTPUTS: /opt/gh-aw/safeoutputs/outputs.jsonl
      GH_AW_WORKFLOW_ID_SANITIZED: myworkflow
    steps:
      - name: Execute
        env:
          STEP_VAR: value
```

> [!NOTE]
> The frontmatter `env:` variables are merged into the agent job's `env:` section alongside system-generated variables like `GH_AW_SAFE_OUTPUTS` and `GH_AW_WORKFLOW_ID_SANITIZED`.

## Debugging Environment Variables

**View all available variables in agent job:**
```yaml wrap
---
env:
  TEST_VAR: agent-value
---

# Add this to your workflow markdown to debug
```

The agent job will have access to `TEST_VAR` along with system variables.

**Custom jobs define their own env:**
```yaml wrap
---
env:
  AGENT_VAR: agent-value  # Only in agent job

jobs:
  debug:
    env:
      DEBUG_VAR: custom-value  # Only in debug job
    steps:
      - run: env | sort
---
```

## Related Documentation

- [Frontmatter Reference](/gh-aw/reference/frontmatter/) - Complete frontmatter configuration
- [Safe Outputs](/gh-aw/reference/safe-outputs/) - Safe output environment configuration
- [Sandbox](/gh-aw/reference/sandbox/) - Sandbox environment variables
- [Tools](/gh-aw/reference/tools/) - MCP tool configuration
- [Safe Inputs](/gh-aw/reference/safe-inputs/) - Safe input tool configuration
- [GitHub Actions Environment Variables](https://docs.github.com/en/actions/learn-github-actions/variables) - GitHub Actions documentation
