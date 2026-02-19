---
title: Triggers
description: Triggers in GitHub Agentic Workflows
sidebar:
  order: 400
---

The `on:` section uses standard GitHub Actions syntax to define workflow triggers. For example:

```yaml wrap
on:
  issues:
    types: [opened]
```

## Trigger Types

GitHub Agentic Workflows supports all standard GitHub Actions triggers plus additional enhancements for reactions, cost control, and advanced filtering.

### Dispatch Triggers (`workflow_dispatch:`)

Run workflows manually from the GitHub UI, API, or via `gh aw run`/`gh aw trial`. [Full syntax reference](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#on).

```yaml wrap
on:
  workflow_dispatch:
    inputs:
      topic:
        description: 'Research topic'
        required: true
        type: string
      priority:
        description: 'Task priority'
        required: false
        type: choice
        options:
          - low
          - medium
          - high
        default: medium
      deploy_env:
        description: 'Target environment'
        required: false
        type: environment
        default: staging
```

**Supported input types:** `string`, `boolean`, `choice` (dropdown with predefined options), `environment` (dropdown of repository environments; returns the name as a string, does not enforce protection rules).

#### Accessing Inputs in Markdown

Use `${{ github.event.inputs.INPUT_NAME }}` expressions in your workflow content:

```aw wrap
---
on:
  workflow_dispatch:
    inputs:
      topic:
        description: 'Research topic'
        required: true
        type: string
permissions:
  contents: read
safe-outputs:
  create-discussion:
---

# Research Assistant

Research the following topic: "${{ github.event.inputs.topic }}"

Provide a comprehensive summary with key findings and recommendations.
```

### Scheduled Triggers (`schedule:`)

Run workflows on a recurring schedule using human-friendly expressions or [cron syntax](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#schedule).

**Fuzzy Scheduling (Recommended):** The compiler deterministically assigns each workflow a unique execution time based on the workflow file path, distributing load without manual coordination:

```yaml wrap
on:
  schedule: daily                          # Scattered time
  schedule: daily around 14:00            # Scattered within ±1 hour (13:00–15:00)
  schedule: daily between 9:00 and 17:00  # Scattered within business hours
  schedule: daily between 9am and 5pm utc-5  # With UTC offset
```

**Human-Friendly Shorthand:**

```yaml wrap
on: daily                  # Recommended
on: weekly on monday
on: every 6h
```

**Fixed-Time Cron Format:**

```yaml wrap
on:
  schedule:
    - cron: "30 6 * * 1"  # Monday at 06:30 UTC
    - cron: "0 9 15 * *"  # 15th of month at 09:00 UTC
```

**Supported Formats:**

| Format | Example | Result | Notes |
|--------|---------|--------|-------|
| **Hourly (Fuzzy)** | `hourly` | `58 */1 * * *` | Compiler assigns scattered minute |
| **Daily (Fuzzy)** | `daily` | `43 5 * * *` | Compiler assigns scattered time |
| | `daily around 14:00` | `20 14 * * *` | Scattered within ±1 hour (13:00-15:00) |
| | `daily between 9:00 and 17:00` | `37 13 * * *` | Scattered within range (9:00-17:00) |
| | `daily between 9am and 5pm utc-5` | `12 18 * * *` | With UTC offset (9am-5pm EST → 2pm-10pm UTC) |
| | `daily around 3pm utc-5` | `33 19 * * *` | With UTC offset (3 PM EST → 8 PM UTC) |
| **Weekly (Fuzzy)** | `weekly` or `weekly on monday` | `43 5 * * 1` | Compiler assigns scattered time |
| | `weekly on friday around 5pm` | `18 16 * * 5` | Scattered within ±1 hour |
| **Intervals** | `every 10 minutes` | `*/10 * * * *` | Minimum 5 minutes |
| | `every 2h` | `53 */2 * * *` | Fuzzy: scattered minute offset |
| | `0 */2 * * *` | `0 */2 * * *` | Cron syntax for fixed times |

**Time formats:** `HH:MM` (24-hour), `midnight`, `noon`, `1pm`–`12pm`, `1am`–`12am`
**UTC offsets:** Add `utc+N` or `utc-N` to any time (e.g., `daily around 14:00 utc-5`)

> [!TIP]
> Complete Schedule Syntax Reference
> See the [Schedule Syntax reference](/gh-aw/reference/schedule-syntax/) for all supported formats, monthly and interval schedules, and additional UTC offset examples.

### Issue Triggers (`issues:`)

Trigger on issue events. [Full event reference](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#issues).

```yaml wrap
on:
  issues:
    types: [opened, edited, labeled]
```

#### Issue Locking (`lock-for-agent:`)

Prevent concurrent modifications to an issue by locking it at workflow start and unlocking it after completion (with `always()` to ensure unlock on failure). Safe outputs are processed after unlock.

```yaml wrap
on:
  issues:
    types: [opened, edited]
    lock-for-agent: true
```

Use when workflows make multiple sequential updates or when race conditions must be prevented. Requires `issues: write` permission (automatically added). Pull requests and already-locked issues are silently skipped.

```aw wrap
---
on:
  issues:
    types: [opened]
    lock-for-agent: true
permissions:
  contents: read
  issues: write
safe-outputs:
  add-comment:
    max: 3
---

# Issue Processor with Locking

Process the issue and make multiple updates without interference
from concurrent modifications.

Context: "${{ needs.activation.outputs.text }}"
```

### Pull Request Triggers (`pull_request:`)

Trigger on pull request events. [Full event reference](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request).

```yaml wrap
on:
  pull_request:
    types: [opened, synchronize, labeled]
    names: [ready-for-review, needs-review]
  reaction: "rocket"
```

#### Fork Filtering (`forks:`)

Pull request workflows block forks by default. Use `forks:` to allow specific patterns:

```yaml wrap
on:
  pull_request:
    types: [opened, synchronize]
    forks: ["trusted-org/*"]  # Allow forks from trusted-org
    # forks: ["*"]            # Allow all forks (use with caution)
    # forks: ["owner/repo"]   # Allow specific repository
```

Omitting `forks` restricts to same-repository PRs only. Fork detection uses repository ID comparison and is not affected by repository renames.

### Comment Triggers
```yaml wrap
on:
  issue_comment:
    types: [created]
  pull_request_review_comment:
    types: [created]
  discussion_comment:
    types: [created]
  reaction: "eyes"
```

#### Comment Locking (`lock-for-agent:`)

For `issue_comment` events, lock the parent issue during workflow execution:

```yaml wrap
on:
  issue_comment:
    types: [created, edited]
    lock-for-agent: true
```

Behavior is identical to [Issue Locking](#issue-locking-lock-for-agent). Pull request comments are silently skipped.

### Workflow Run Triggers (`workflow_run:`)

Trigger workflows after another workflow completes. [Full event reference](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#workflow_run).

```yaml wrap
on:
  workflow_run:
    workflows: ["CI"]
    types: [completed]
    branches:
      - main
      - develop
```

#### Security Protections

The compiler automatically injects repository ID and fork checks to prevent cross-repository attacks. Always include `branches` to limit which branches can trigger the event—omitting branch restrictions causes warnings (or errors in strict mode).

See the [Security Architecture](/gh-aw/introduction/architecture/) for details.

### Command Triggers (`slash_command:`)

The `slash_command:` trigger responds to `/command-name` mentions in issues, pull requests, and comments. See [Command Triggers](/gh-aw/reference/command-triggers/) for complete documentation.

```yaml wrap
on:
  slash_command:
    name: my-bot
    events: [issues, issue_comment]  # Optional event filter
# Shorthand equivalents:
on:
  slash_command: "my-bot"
on: /my-bot  # Expands to slash_command + workflow_dispatch
```

**Complete Workflow Example:**
```aw wrap
---
on:
  slash_command:
    name: code-review
    events: [pull_request, pull_request_comment]
permissions:
  contents: read
  pull-requests: write
tools:
  github:
    toolsets: [pull_requests]
safe-outputs:
  add-comment:
    max: 5
timeout-minutes: 10
---

# Code Review Assistant

When someone mentions /code-review in a pull request or PR comment,
analyze the code changes and provide detailed feedback.

The current context is: "${{ needs.activation.outputs.text }}"

Review the pull request changes and add helpful review comments on specific
lines of code where improvements can be made.
```

The command must appear as the **first word** in the comment or body text. Command workflows automatically add the "eyes" (👀) reaction and edit comments with workflow run links.

### Label Filtering (`names:`)

Filter issue and pull request triggers by label names:

```yaml wrap
on:
  issues:
    types: [labeled, unlabeled]
    names: [bug, critical, security]
```

**Shorthand syntax** automatically includes `workflow_dispatch` and compiles to standard GitHub Actions syntax:

```yaml wrap
on: issue labeled bug
on: issue labeled bug, enhancement, priority-high
on: pull_request labeled needs-review, ready-to-merge
```

Supported for `issue`, `pull_request`, and `discussion` events. See [LabelOps workflows](/gh-aw/patterns/labelops/) for examples.

### Reactions (`reaction:`)

Add an emoji reaction to the triggering item to provide visual workflow status feedback:

```yaml wrap
on:
  issues:
    types: [opened]
  reaction: "eyes"
```

For issues/PRs, a comment with the workflow run link is also created. For command workflows, the original comment is edited with the run link.

**Available reactions:** `+1` 👍, `-1` 👎, `laugh` 😄, `confused` 😕, `heart` ❤️, `hooray` 🎉, `rocket` 🚀, `eyes` 👀

### Stop After Configuration (`stop-after:`)

Automatically disable workflow triggering after a deadline to control costs:

```yaml wrap
on: weekly on monday
  stop-after: "+25h"  # 25 hours from compilation time
```

Accepts absolute dates (`YYYY-MM-DD`, `MM/DD/YYYY`, `DD/MM/YYYY`, ISO 8601, natural language like `1st June 2025`) or relative deltas (`+7d`, `+25h`, `+1d12h30m`) from compilation time. Minimum granularity is hours. Recompiling resets the stop time.

### Manual Approval Gates (`manual-approval:`)

Require manual approval before workflow execution using GitHub environment protection rules:

```yaml wrap
on:
  workflow_dispatch:
  manual-approval: production
```

Sets the `environment` on the activation job, enabling approval gates from repository or organization settings. Configure reviewers and wait timers in Settings → Environments. See [GitHub's environment documentation](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment) for setup details.

### Skip-If-Match Condition (`skip-if-match:`)

Skip workflow execution when a GitHub search query has matches—useful for preventing duplicate scheduled runs:

```yaml wrap
on: daily
  skip-if-match: 'is:issue is:open in:title "[daily-report]"'  # Skip if any match

on: weekly on monday
  skip-if-match:
    query: "is:pr is:open label:urgent"
    max: 3  # Skip if 3 or more PRs match
```

The query runs against the current repository before activation. If matches reach or exceed `max` (default `1`), the workflow is skipped. Supports all standard GitHub search qualifiers.

### Skip-If-No-Match Condition (`skip-if-no-match:`)

The inverse of `skip-if-match`: skip when a query has **no matches** (or fewer than the minimum required):

```yaml wrap
on: weekly on monday
  skip-if-no-match: 'is:pr is:open label:ready-to-deploy'  # Skip if no matches

on:
  workflow_dispatch:
  skip-if-no-match:
    query: "is:issue is:open label:urgent"
    min: 3  # Only run if 3 or more issues match
```

If matches fall below `min` (default `1`), the workflow is skipped. Can be combined with `skip-if-match` for complex conditions.

## Related Documentation

- [Schedule Syntax](/gh-aw/reference/schedule-syntax/) - Complete schedule format reference
- [Command Triggers](/gh-aw/reference/command-triggers/) - Special @mention triggers and context text
- [Frontmatter](/gh-aw/reference/frontmatter/) - Complete frontmatter configuration
- [LabelOps](/gh-aw/patterns/labelops/) - Label-based automation workflows
- [Workflow Structure](/gh-aw/reference/workflow-structure/) - Directory layout and organization
