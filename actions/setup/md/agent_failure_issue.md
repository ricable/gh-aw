### Workflow Failure

**Workflow:** [{workflow_name}]({workflow_source_url})  
**Branch:** {branch}  
**Run URL:** {run_url}{pull_request_info}

{secret_verification_context}{assignment_errors_context}{create_discussion_errors_context}{repo_memory_validation_context}{missing_data_context}{missing_safe_outputs_context}

### Action Required

**Option 1: Assign this issue to agent using agentic-workflows**

Assign this issue to the `agentic-workflows` agent to automatically debug and fix the workflow failure.

**Option 2: Manually invoke the agent**

Debug this workflow failure using the `agentic-workflows` agent. Load the agent in your AI coding assistant by either:

- Loading `.github/agents/agentic-workflows.agent.md` from this repository, or
- Downloading the agent directly from [GitHub](https://github.com/github/gh-aw/blob/main/.github/agents/agentic-workflows.agent.md)

Then provide this prompt:

```
debug the agentic workflow {workflow_id} failure in {run_url}
```
