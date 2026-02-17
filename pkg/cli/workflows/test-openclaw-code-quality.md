---
on:
  workflow_dispatch:
permissions:
  contents: read
  issues: read
  pull-requests: read
engine:
  id: openclaw
  model: default-agent
network:
  allowed:
    - defaults
    - github
    - node
tools:
  github:
  bash:
    - "*"
---

# OpenClaw Code Quality Analysis

Analyze repository code quality and provide a summary.

## Instructions

1. Explore the repository structure to understand the project layout
2. Run available linting or static analysis tools if present (check package.json, Makefile, etc.)
3. Review a sample of source files for:
   - Code organization and modularity
   - Error handling patterns
   - Test coverage presence
   - Documentation quality
4. Produce a brief summary of findings covering:
   - Overall code health assessment
   - Areas that could benefit from improvement
   - Positive patterns observed
