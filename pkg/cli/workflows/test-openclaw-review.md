---
on:
  pull_request:
    types: [opened, synchronize]
permissions:
  contents: read
  pull-requests: read
  issues: read
engine: openclaw
network:
  allowed:
    - defaults
    - github
tools:
  github:
safe-outputs:
  add-comment:
    max: 1
    hide-older-comments: true
---

# OpenClaw PR Review

Review the pull request changes and post a summary comment.

## Instructions

1. Read the pull request diff to understand what changed
2. Analyze the changes for:
   - Code quality and correctness
   - Potential bugs or issues
   - Missing test coverage
   - Style and convention adherence
3. Post a concise review comment on the pull request summarizing:
   - What the PR does (1-2 sentences)
   - Key observations or concerns (bullet points)
   - Suggested improvements if any
