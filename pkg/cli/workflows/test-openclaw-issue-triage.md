---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
  pull-requests: read
engine:
  id: openclaw
  args: ["--thinking", "high"]
network:
  allowed:
    - defaults
    - github
tools:
  github:
safe-outputs:
  add-comment:
    max: 1
  add-labels:
    allowed: [bug, enhancement, question, documentation, good-first-issue]
---

# OpenClaw Issue Triage

Analyze newly opened issues and help categorize them.

## Instructions

1. Read the issue title and body carefully
2. Determine the issue type:
   - **bug**: Something is broken or not working as expected
   - **enhancement**: A request for new functionality
   - **question**: A question about usage or behavior
   - **documentation**: Missing or incorrect documentation
   - **good-first-issue**: Simple issues suitable for new contributors
3. Apply the appropriate label(s) to the issue
4. Post a brief comment acknowledging the issue and confirming the categorization
