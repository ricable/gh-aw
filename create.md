# Creating Agentic Workflows and Other Actions

This prompt guides you, a coding agent, to create, debug, update or do other actions related to **GitHub Agentic Workflows (gh-aw)** in a repository.

## Step 1: Install GitHub Agentic Workflows CLI Extension

Check if `gh aw` is installed by running

```bash
gh aw version
```

If it is installed, run:

```bash
gh extension upgrade aw
```

to upgrade to latest. If it is not installed, run the installation script from the main branch of the gh-aw repository:

```bash
curl -sL https://raw.githubusercontent.com/github/gh-aw/main/install-gh-aw.sh | bash
```

**What this does**: Downloads and installs the gh-aw binary to `~/.local/share/gh/extensions/gh-aw/`

**Verify installation**:

```bash
gh aw version
```

You should see version information displayed. If you encounter an error, check that:

- GitHub CLI (`gh`) is installed and authenticated
- The installation script completed without errors
- `~/.local/share/gh/extensions` is in your PATH

## Step 2: Create the Workflow or do Other Actions

Follow the instructions in the appropriate prompt based on the user's request.

Here ROOT is the location where you found this file. For example,

* if this file is at `https://github.com/github/gh-aw/blob/main/create.md` then the ROOT is `https://github.com/github/gh-aw/blob/main/`
* if this file is at `https://github.com/github/gh-aw/blob/v0.35.1/create.md` then the ROOT is `https://github.com/github/gh-aw/v0.35.1/main/`

### Create New Workflow

**Load when**: User wants to create a new workflow from scratch, add automation, or design a workflow that doesn't exist yet

**Prompt file**: `ROOT/.github/aw/create-agentic-workflow.md`

**Use cases**:

- "Create a workflow that triages issues"
- "I need a workflow to label pull requests"
- "Design a weekly research automation"

### Update Existing Workflow  

**Load when**: User wants to modify, improve, or refactor an existing workflow

**Prompt file**: `ROOT/.github/aw/update-agentic-workflow.md` (relative to the location where you found this file. For example, if this file is at https://github.com/github/gh-aw/blob/main/create.md then the prompt file is at https://github.com/github/gh-aw/blob/main/.github/aw/update-agentic-workflow.md)

**Use cases**:

- "Add web-fetch tool to the issue-classifier workflow"
- "Update the PR reviewer to use discussions instead of issues"
- "Improve the prompt for the weekly-research workflow"

### Debug Workflow  

**Load when**: User needs to investigate, audit, debug, or understand a workflow, troubleshoot issues, analyze logs, or fix errors

**Prompt file**: `ROOT/.github/aw/debug-agentic-workflow.md` (relative to the location where you found this file. For example, if this file is at https://github.com/github/gh-aw/blob/main/create.md then the prompt file is at https://github.com/github/gh-aw/blob/main/.github/aw/debug-agentic-workflow.md)

**Use cases**:

- "Why is this workflow failing?"
- "Analyze the logs for workflow X"
- "Investigate missing tool calls in run #12345"

### Upgrade Agentic Workflows

**Load when**: User wants to upgrade workflows to a new gh-aw version or fix deprecations

**Prompt file**: `ROOT/.github/aw/upgrade-agentic-workflows.md` (relative to the location where you found this file. For example, if this file is at https://github.com/github/gh-aw/blob/main/create.md then the prompt file is at https://github.com/github/gh-aw/blob/main/.github/aw/debug-agentic-workflow.md)

**Use cases**:

- "Upgrade all workflows to the latest version"
- "Fix deprecated fields in workflows"
- "Apply breaking changes from the new release"

### Create Shared Agentic Workflow

**Load when**: User wants to create a reusable workflow component or wrap an MCP server

**Prompt file**: `ROOT/.github/aw/create-shared-agentic-workflow.md` (relative to the location where you found this file. For example, if this file is at https://github.com/github/gh-aw/blob/main/create.md then the prompt file is at https://github.com/github/gh-aw/blob/main/.github/aw/update-agentic-workflow.md)

**Use cases**:

- "Create a shared component for Notion integration"
- "Wrap the Slack MCP server as a reusable component"
- "Design a shared workflow for database queries"

If you need to clarify requirements or discuss options, and you are working in an interactive agent chat system, do so interactively with the user. If running non-interactively, make reasonable assumptions based on the repository context.

## Reusable Components

GitHub Agentic Workflows supports modular, reusable components that can be imported into your workflows. This enables:

- **DRY Principle**: Avoid duplicating configuration across multiple workflows
- **Rapid Development**: Compose workflows by importing pre-built capabilities
- **Separation of Concerns**: Maintain tool configs, permissions, and prompts independently
- **Consistent Patterns**: Use battle-tested components for common patterns

### Common Reusable Components

**Orchestration** - Coordinate multiple workflows or agents

Import `shared/orchestration.md` when your workflow needs to delegate work to other workflows or AI agents:

```yaml
imports:
  - shared/orchestration.md
safe-outputs:
  dispatch-workflow:
    workflows: [worker-a, worker-b]
    max: 10
  assign-to-agent:
    name: copilot
    max: 5
```

**Use cases**: Multi-phase initiatives, fan-out work distribution, coordinating specialized workers

**Monitoring** - Track workflow progress in GitHub Projects

Import `shared/projects.md` when your workflow should update project boards:

```yaml
imports:
  - shared/projects.md
safe-outputs:
  update-project:
    project: https://github.com/orgs/myorg/projects/123
    max: 10
  create-project-status-update:
    project: https://github.com/orgs/myorg/projects/123
    max: 1
```

**Use cases**: Tracking workflow-created issues, periodic status updates, operational dashboards

**Reporting** - Consistent report formatting

Import `shared/reporting.md` for standardized report structure:

```yaml
imports:
  - shared/reporting.md
```

**Use cases**: Daily/weekly status reports, test results, analysis summaries

### Discovering Available Components

Repositories using gh-aw often organize shared components in:

- `.github/workflows/shared/` - Core capabilities (reporting, data analysis, tool configs)
- `.github/workflows/shared/mcp/` - MCP server configurations

List available components:

```bash
ls .github/workflows/shared/
```

### Using Imports

Add imports to your workflow frontmatter:

```yaml
---
on: issues
imports:
  - shared/orchestration.md
  - shared/projects.md
  - shared/mcp/tavily.md
---
```

The compiler merges imported configurations with your workflow's config. See [Packaging & Distribution](/gh-aw/guides/packaging-imports/) for complete details on import behavior and merge semantics.

### Available After Installation

When you run `gh aw init` in a repository, several helper files become available:

**Workflow Prompt Files** (`.github/aw/`)
- `create-agentic-workflow.md` - Guides creation of new workflows
- `update-agentic-workflow.md` - Guides updates to existing workflows
- `debug-agentic-workflow.md` - Helps debug workflow issues
- `create-shared-agentic-workflow.md` - Creates reusable shared components
- `upgrade-agentic-workflows.md` - Upgrades workflows to new gh-aw versions
- `github-agentic-workflows.md` - GitHub Copilot custom instructions

**Agent Files** (`.github/agents/`)
- `agentic-workflows.agent.md` - Dispatcher agent that routes to appropriate prompts

These files are available for AI assistants to use when helping you create, update, or debug workflows. The `create.md` file (this file) references these prompts and routes you to the appropriate one based on your task.

### Learn More

- **[Orchestration Guide](/gh-aw/guides/orchestration/)** - Orchestrator/worker pattern for coordinating multiple workflows
- **[Projects & Monitoring Guide](/gh-aw/guides/monitoring/)** - Track workflow activities in GitHub Projects boards
- **[Packaging & Distribution Guide](/gh-aw/guides/packaging-imports/)** - Import system, versioning, and shared components

## Step 3: Review Changes

Check what files were changed or created:

```bash
git status
```

If creating a workflow, the actual files you created will be under `.github/workflows/`. There should be at least one workflow file and one lock file.

- `.github/workflows/<workflow-name>.md`
- `.github/workflows/<workflow-name>.lock.yml`

If creating a workflow, check the .gitattributes file and make sure it exists and contains at least the following line:

```text
.github/workflows/*.lock.yml linguist-generated=true merge=ours
```

You do not need to run `gh aw init` as part of your workflow creation. However if you did run this you may also see:

- `.github/aw/github-agentic-workflows.md`
- `.github/agents/agentic-workflows.agent.md`
- `.vscode/settings.json`
- `.vscode/mcp.json`
- And several other configuration files

Don't remove these but don't add them if not already present in the repo. Unless instructed otherwise do NOT commit the changes to ANY files except the gitattributes file and workflow files.

- `.gitattributes`
- `.github/workflows/<workflow-name>.md`
- `.github/workflows/<workflow-name>.lock.yml`

## Step 4: Commit and Push Changes

Commit the changes, e.g.

```bash
git add .gitattributes .github/workflows/<workflow-name>.md .github/workflows/<workflow-name>.lock.yml
git commit -m "Initialize repository for GitHub Agentic Workflows"
git push
```

If there is branch protection on the default branch, create a pull request instead and report the link to the pull request.

## Troubleshooting

See the separate guides on troubleshooting common issues.

## Instructions

When a user interacts with you:

1. **Identify the task type** from the user's request
2. **Fetch and read the appropriate prompt** 
3. **Follow the loaded prompt's instructions** exactly
4. **If uncertain**, ask clarifying questions to determine the right prompt

## Quick Reference

```bash
# Create a new workflow
gh aw new <workflow-name>

# Compile workflows
gh aw compile [workflow-name]

# Debug workflow runs
gh aw logs [workflow-name]
gh aw audit <run-id>

# Upgrade workflows
gh aw fix --write
gh aw compile --validate
```

## Key Features of gh-aw

- **Natural Language Workflows**: Write workflows in markdown with YAML frontmatter
- **AI Engine Support**: Copilot, Claude, Codex, or custom engines
- **MCP Server Integration**: Connect to Model Context Protocol servers for tools
- **Safe Outputs**: Structured communication between AI and GitHub API
- **Strict Mode**: Security-first validation and sandboxing
- **Shared Components**: Reusable workflow building blocks
- **Repo Memory**: Persistent git-backed storage for agents

## Important Notes

- Workflows must be compiled to `.lock.yml` files before running in GitHub Actions
- Follow security best practices: minimal permissions, explicit network access, no template injection
