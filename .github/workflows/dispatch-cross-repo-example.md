---
on: issues
engine: copilot
permissions:
  contents: read

safe-outputs:
  dispatch-workflow:
    workflows:
      - test-workflow
    max: 3
    target-repo: github/gh-aw
    allowed-repos:
      - github/test-repo
      - octocat/hello-world
---

# Cross-Repository Dispatch Example

This workflow demonstrates dispatching workflows across repositories using the safe-outputs.dispatch-workflow feature.

## Same-Repository Dispatch

By default, workflows are dispatched to the current repository:

```json
{
  "type": "dispatch_workflow",
  "workflow_name": "test-workflow",
  "inputs": {
    "environment": "staging"
  }
}
```

## Cross-Repository Dispatch

To dispatch to a different repository, specify the `repo` field. The repository must be in the allowed-repos list:

```json
{
  "type": "dispatch_workflow",
  "workflow_name": "test-workflow",
  "repo": "github/test-repo",
  "inputs": {
    "environment": "production"
  }
}
```

## Security

- **Deny-by-default**: Cross-repository dispatch requires explicit allowlist
- **Same-repo always allowed**: The workflow's repository is always allowed
- **Bare names auto-qualified**: `"test-repo"` becomes `"github/test-repo"` based on target-repo's org

## Configuration Options

- `workflows`: List of workflow names to allow dispatching
- `max`: Maximum number of dispatches per run (default: 1, max: 50)
- `target-repo`: Default target repository (optional, defaults to current repo)
- `allowed-repos`: Additional repositories that can be targeted (deny-by-default)
