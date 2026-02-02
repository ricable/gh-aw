---
title: "Imports & Sharing: Peli's Secret Weapon"
description: "How modular, reusable components enabled scaling our agent collection"
authors:
  - dsyme
  - pelikhan
  - mnkiefer
date: 2026-01-30
draft: true
prev:
  link: /gh-aw/blog/2026-01-27-operational-patterns/
  label: 9 Operational Patterns
next:
  link: /gh-aw/blog/2026-02-02-security-lessons/
  label: Security Lessons
---

[Previous Article](/gh-aw/blog/2026-01-27-operational-patterns/)

---

<img src="/gh-aw/peli.png" alt="Peli de Halleux" width="200" style="float: right; margin: 0 0 20px 20px; border-radius: 8px;" />

*Come with me, and you'll see* another installment in our Peli's Agent Factory series! We've already toured the [workflows](/gh-aw/blog/2026-01-13-meet-the-workflows/), learned our [lessons](/gh-aw/blog/2026-01-21-twelve-lessons/), discovered the [secret recipes](/gh-aw/blog/2026-01-24-design-patterns/), and explored [operational patterns](/gh-aw/blog/2026-01-27-operational-patterns/). Today, I shall reveal the *everlasting gobstopper* - the secret weapon that made scaling possible: imports!

Here's the truth: tending dozens of agents would be completely unsustainable without reuse. One of the most powerful features that let us scale Peli's Agent Factory is the **imports system** - a mechanism for sharing and reusing workflow components across the entire factory.

Instead of duplicating configuration, tool setup, and instructions in every single workflow, we created a library of shared components that agents can import on-demand. This isn't just about being DRY (though that's nice) - it's carefully designed to support modularization, sharing, installation, pinning, and versioning of single-file portions of agentic workflows.

Let's dive in!

## The Power of Imports

Imports transform workflow development through four key benefits:

**DRY Principle**: When we enhanced [`reporting.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/reporting.md), all 46 workflows that imported it immediately benefited. One change, 46 workflows improved.

**Composable Capabilities**: Workflows mix and match components like LEGO blocks. Need visualization, JSON processing, and web search? Just import `python-dataviz.md`, `jqschema.md`, and `mcp/tavily.md`.

**Separation of Concerns**: Infrastructure teams manage MCP servers, security teams handle network policies, data teams build visualization components, and agent authors focus on prompts.

**Rapid Experimentation**: New workflows often need just a prompt and 3-5 imports. Prototype agents in minutes:

```markdown
---
description: Analyze code patterns
imports:
  - shared/reporting.md
  - shared/mcp/serena.md
  - shared/jqschema.md
---

Analyze the codebase for common patterns...
```

## The Import Library

The factory organized shared components into two main directories:

### Core Capabilities: `.github/workflows/shared/`

35+ components providing fundamental capabilities. Top imports:

- [`reporting.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/reporting.md?plain=1) (46) - Report formatting, workflow references, consistent structure
- [`jqschema.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/jqschema.md?plain=1) (17) - JSON querying and schema validation
- [`python-dataviz.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/python-dataviz.md?plain=1) (7) - Data visualization with NumPy, Pandas, Matplotlib
- [`trending-charts-simple.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/trending-charts-simple.md?plain=1) (6) - Trend visualizations and time-series
- [`gh.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/gh.md?plain=1) (4) - Safe-input wrapper for GitHub CLI
- [`copilot-pr-data-fetch.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/copilot-pr-data-fetch.md?plain=1) (4) - Copilot data fetching and cache management

Plus specialized components for data analysis ([`charts-with-trending.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/charts-with-trending.md), [`ci-data-analysis.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/ci-data-analysis.md)), prompting ([`keep-it-short.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/keep-it-short.md)), and safe outputs ([`safe-output-app.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/safe-output-app.md)).

### MCP Server Configurations: `.github/workflows/shared/mcp/`

20+ MCP server configurations for specialized capabilities. Top servers:

- [`gh-aw.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/gh-aw.md?plain=1) (12) - Workflow debugging and metadata access
- [`tavily.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/tavily.md?plain=1) (5) - Web search and research
- [`markitdown.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/markitdown.md?plain=1) (3) - Document conversion (PDF, Office to Markdown)
- [`ast-grep.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/ast-grep.md?plain=1) (2) - Structural code search
- [`brave.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/brave.md?plain=1) (2) - Privacy-focused web search

Plus development tools ([`jupyter.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/jupyter.md), [`serena.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/serena.md)), knowledge sources ([`arxiv.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/arxiv.md), [`deepwiki.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/deepwiki.md)), and external integrations ([`slack.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/slack.md), [`notion.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/notion.md), [`sentry.md`](https://github.com/githubnext/gh-aw/tree/2c1f68a721ae7b3b67d0c2d93decf1fa5bcf7ee3/.github/workflows/shared/mcp/sentry.md)).

## Import Statistics

The factory's extensive use of imports demonstrates their value:

- **84 workflows** (65% of factory) use the imports feature
- **46 workflows** import `reporting.md` (most popular component)
- **17 workflows** import `jqschema.md` (JSON utilities)
- **12 workflows** import `mcp/gh-aw.md` (meta-analysis server)
- **35+ shared components** in `.github/workflows/shared/`
- **20+ MCP server configs** in `.github/workflows/shared/mcp/`
- **Average 2-3 imports** per workflow (some have 8+!)

## How Imports Work

### Basic Import Syntax

```markdown
---
description: My workflow
imports:
  - shared/reporting.md
  - shared/mcp/tavily.md
---

Your workflow prompt here...
```

### What Gets Imported

When a workflow imports a shared component, several things are merged:

1. **Frontmatter** - Tools, permissions, network settings
2. **Instructions** - Prompt guidance and context
3. **MCP Servers** - Tool configurations
4. **Safe Outputs** - Output templates

### Import Resolution

Imports are resolved at compile time:

1. Parse workflow frontmatter
2. Load each imported file
3. Merge configurations (workflow overrides imports)
4. Compile to final YAML

### Versioning & Pinning

Imports can be pinned to specific commits:

```markdown
imports:
  - shared/reporting.md@abc123
  - shared/mcp/tavily.md@v1.2.0
```

This ensures stability for production workflows while allowing experimentation with latest versions.

## Best Practices for Imports

**Creating shared components**: Keep them focused and single-purpose. Document configuration options, version significant changes, and test with multiple importers. Avoid monolithic components, breaking changes without versioning, or hard-coding repository-specific values.

**Using imports effectively**: Import only what you need and pin critical production workflows. Override imported settings when necessary, but understand the impact. Test after updates and document why each import is needed. Avoid conflicting components and cargo-cult import lists.

## Evolution of Shared Components

The shared component library evolved organically from duplication (workflows 1-10 copy-pasted configuration) to extraction (11-30 created `reporting.md` and `python-dataviz.md`) to ecosystem (31-80 composed existing components) to specialization (domain-specific components for Copilot analysis, security scanning).

## Impact on Velocity

The imports system dramatically accelerated development:

| Metric | Without Imports | With Imports |
| ------ | --------------- | ------------ |
| Time to create workflow | 2-4 hours | 15-30 minutes |
| Lines of configuration | 100-200 | 20-40 |
| Maintenance burden | High | Low |
| Consistency | Manual | Automatic |
| Reuse rate | ~0% | ~65% |

## Common Import Patterns

**Analyst**: `reporting.md` + `jqschema.md` + `python-dataviz.md` for read-only analysis with visualization.

**Researcher**: `reporting.md` + `mcp/tavily.md` + `mcp/arxiv.md` for web search and academic papers.

**Code Intelligence**: `reporting.md` + `mcp/serena.md` + `mcp/ast-grep.md` for semantic analysis and refactoring.

**Meta-Agent**: `reporting.md` + `mcp/gh-aw.md` + `charts-with-trending.md` for workflows analyzing workflows.

## What's Next?

The imports system enabled rapid scaling, but even the best components need proper security foundations. All the reusability in the world doesn't help if agents can accidentally cause harm.

In our next article, we'll explore the security lessons learned from operating our collection of automated agentic workflows with access to real repositories.

_More articles in this series coming soon._

[Previous Article](/gh-aw/blog/2026-01-27-operational-patterns/)
