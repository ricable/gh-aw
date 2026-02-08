# Analysis: Can `.github/aw/*.md` Files Be Imported Like `shared/` Files?

## Executive Summary

**YES** - Workflows CAN technically import files from `.github/aw/` directory.

**SHOULD YOU?** It depends on the content type:
- **Hybrid content** (safe-outputs + runtime instructions): ⚠️ CONSIDER (acceptable for reducing boilerplate)
- **Pure agent configuration**: ❌ NO (use `shared/` instead)
- **Reusable components**: ✅ YES (but create in `shared/` directory)

## Technical Feasibility

### ✅ What Works

1. **Import Syntax**: Both relative and absolute paths work:
   ```yaml
   imports:
     - ../aw/orchestration.md          # Relative from .github/workflows/
     - .github/aw/orchestration.md     # Absolute path
   ```

2. **Compilation**: Successfully compiles to `.lock.yml` files
3. **Content Merging**: Frontmatter from `.github/aw/` files merges same as `shared/` files
4. **Runtime Loading**: Content loads via `{{#runtime-import}}` macro

### 🔍 Path Resolution Logic

From `pkg/parser/remote_fetch.go`:

```go
// isWorkflowSpec checks if a path looks like a workflowspec
func isWorkflowSpec(path string) bool {
    // ...
    
    // Reject paths that start with "." (local paths like .github/workflows/...)
    if strings.HasPrefix(cleanPath, ".") {
        return false
    }
    
    // Reject paths that start with "shared/" (local shared files)
    if strings.HasPrefix(cleanPath, "shared/") {
        return false
    }
    
    // ...
}
```

**Key Point**: The code explicitly rejects `shared/` prefix as local, but does NOT reject `.github/` prefix. This means `.github/aw/` imports are treated as local file paths and work correctly.

## Directory Purpose & Architecture

### `.github/workflows/shared/` - Workflow Components

**Purpose**: Reusable workflow building blocks

**Contents**:
- Tool configurations (bash, github, web-fetch)
- MCP server setups (tavily, serena, ast-grep)
- Report formatting guidelines
- Data visualization templates
- Network and permission configurations

**Design Pattern**: Component library for composing workflows

**Import Stats**:
- 35+ shared components
- 65% of workflows (84/130) use imports
- Most imported: `reporting.md` (46 imports)

**Example Frontmatter**:
```yaml
---
# Tool configurations
tools:
  bash:
    allowed: [read, write]
  github:
    toolsets: [default]

# Report formatting
---
```

### `.github/aw/` - Agent Configuration Files

**Purpose**: Agent behavior and orchestration prompts

**Contents**:
- Agent configuration (applyTo patterns)
- Meta-instructions for creating/updating workflows
- Orchestration patterns
- Documentation references

**Design Pattern**: Agent instruction files, not workflow components

**Import Stats**:
- 15+ agent files
- **0 workflows** currently import from this directory
- Not designed for reuse across workflows

**Example Frontmatter**:
```yaml
---
name: create-agentic-workflow
description: Create new workflows
applyTo: ".github/workflows/*.md"
infer: false
---
```

## Key Differences

| Aspect | `.github/workflows/shared/` | `.github/aw/` |
|--------|----------------------------|---------------|
| **Primary Purpose** | Reusable workflow components | Agent instructions |
| **Frontmatter Type** | Tool configs, MCP servers | Agent metadata (name, applyTo) |
| **Content Type** | Technical configurations + guidance | Pure agent instructions |
| **Import Pattern** | `shared/file.md` (clean, conventional) | `../aw/file.md` or `.github/aw/file.md` (awkward) |
| **Current Usage** | Heavily imported (46 workflows) | Not imported |
| **Semantic Meaning** | "This is meant to be shared" | "This configures agent behavior" |
| **Directory Convention** | Standard for reusable components | Standard for agent configuration |

## When `.github/aw/` Imports Make Sense

### Hybrid Content: Safe Outputs + Instructions

Some files in `.github/aw/` serve **dual purposes**:

1. **Agent creation guidance**: Help agents understand how to use features
2. **Runtime instructions**: Provide guidance that accompanies safe-output configurations

**Example**: `.github/aw/orchestration.md` contains:
- `assign-to-agent` and `dispatch-workflow` safe-output configs
- Best practices for correlation IDs and tracking
- Runtime guidance on when to use each pattern

For such hybrid content, importing from `.github/aw/` could be **acceptable** because:
- ✅ Reduces boilerplate (no need to duplicate instructions)
- ✅ Keeps instructions and configuration together
- ✅ Makes workflows simpler to write

**Usage Pattern**:
```yaml
---
description: Orchestrator workflow
imports:
  - ../aw/orchestration.md  # Brings in assign-to-agent + guidance
---

# Orchestrator Agent

Analyze the issue and create sub-issues for specialized agents.

[The imported orchestration.md provides instructions on using assign_to_agent()]
```

## Why Pure Agent Config Should NOT Be Imported

For content that's purely agent-specific, importing from `.github/aw/` creates problems:

### 1. **Semantic Confusion**

Files like `create-agentic-workflow.md` are prompts for meta-workflows:
- Designed for the agent that creates workflows
- Not relevant at runtime for regular workflows
- Using these for imports blurs the distinction

### 2. **Path Clarity**

```yaml
# Clear and intentional
imports:
  - shared/reporting.md

# Awkward and unclear (for pure agent config)
imports:
  - ../aw/create-agentic-workflow.md
```

The `shared/` prefix immediately signals "this is meant to be imported."

### 3. **Convention Over Configuration**

The codebase explicitly recognizes `shared/` as the conventional path:
- Path resolution explicitly handles `shared/` prefix
- Documentation consistently uses `shared/` examples
- 65% of workflows follow this pattern

### 4. **Future-Proofing**

If `.github/aw/` becomes a special directory with different semantics (e.g., agent marketplace, compiled agent files), using it for imports could break workflows.

## Recommendations

### ✅ DO: Use `shared/` for Most Reusable Components

```yaml
---
description: My workflow
imports:
  - shared/reporting.md
  - shared/mcp/tavily.md
  - shared/jqschema.md
---
```

### ⚠️ CONSIDER: Import from `.github/aw/` for Hybrid Content

For files that provide **both** safe-output configurations **and** runtime instructions:

```yaml
---
description: Orchestrator workflow
imports:
  - ../aw/orchestration.md  # Acceptable: safe-outputs + instructions
---
```

**When to use this pattern**:
- File contains safe-output configuration (e.g., `assign-to-agent`, `dispatch-workflow`)
- File provides runtime guidance on using those safe-outputs
- Content is genuinely useful during workflow execution (not just agent creation)
- Importing reduces boilerplate and keeps config + docs together

### ❌ DON'T: Import Pure Agent Configuration

```yaml
# Avoid this pattern - pure agent prompts
imports:
  - ../aw/create-agentic-workflow.md     # NO: Agent creation prompt
  - ../aw/github-agentic-workflows.md    # NO: Agent documentation
```

### 🔄 Migration: Create Shared Wrappers (Best Practice)

If hybrid content becomes widely used, create a `shared/` wrapper:

```bash
# Extract reusable parts to shared
cp .github/aw/orchestration.md .github/workflows/shared/orchestration-patterns.md

# Keep agent-specific guidance in .github/aw/
# Update workflows to use shared version
```

This provides the clearest separation of concerns.

1. Extract the reusable parts
2. Create a new file in `shared/`
3. Keep agent-specific content in `.github/aw/`

**Example**:

```bash
# Move orchestration patterns to shared
cp .github/aw/orchestration.md .github/workflows/shared/orchestration-patterns.md

# Update workflows to use shared version
imports:
  - shared/orchestration-patterns.md
```

## Test Results

### ✅ Compilation Test

```yaml
# Test workflow: .github/workflows/test-aw-import.md
---
description: Test importing from .github/aw
on:
  workflow_dispatch:
engine: copilot
imports:
  - ../aw/orchestration.md
---
```

**Result**: ✓ Compiles successfully to `.lock.yml`

**Generated Content**:
```yaml
# Resolved workflow manifest:
#   Imports:
#     - ../aw/orchestration.md

# Runtime loading:
{{#runtime-import .github/aw/orchestration.md}}
```

**Conclusion**: Technical feasibility confirmed, but architectural guidance remains - use `shared/` for clarity.

## Related Files

- Path resolution: `pkg/parser/remote_fetch.go`
- Import processing: `pkg/parser/import_processor.go`
- Import documentation: `docs/src/content/docs/reference/imports.md`
- Shared components blog: `docs/src/content/docs/blog/2026-01-30-imports-and-sharing.md`

## Action Items

1. ✅ Document that `.github/aw/` CAN be imported (technically)
2. ✅ Recommend NOT importing from `.github/aw/` (architecturally)
3. ✅ Update guidance to use `shared/` for all reusable components
4. 📝 Consider adding linter rule to warn about `.github/aw/` imports
5. 📝 Consider documentation update to clarify directory purposes

## Special Case: Hybrid Content

### Generic Agent Instructions with Safe Outputs

**Use Case**: Some content serves BOTH purposes:
1. Agent creation guidance (e.g., how to use `assign-to-agent` or `dispatch_workflow`)
2. Runtime instructions that accompany safe outputs

**Example**: `.github/aw/orchestration.md` contains:
- Safe-output configurations (`assign-to-agent`, `dispatch_workflow`)
- Best practices for using these features
- Runtime guidance on correlation IDs and patterns

**Solution Options**:

#### Option 1: Create Shared Wrapper (Recommended)

Extract the reusable parts into `shared/` and keep agent-specific content in `.github/aw/`:

```yaml
# .github/workflows/shared/orchestration-patterns.md
---
safe-outputs:
  assign-to-agent:
    name: "copilot"
    target: "*"
  dispatch-workflow:
    allowed: ["*"]
---

# Orchestration Patterns

[Runtime instructions for using assign-to-agent and dispatch_workflow]
```

Workflows import from `shared/`:
```yaml
imports:
  - shared/orchestration-patterns.md
```

#### Option 2: Allow Hybrid Imports (Alternative)

For content that legitimately serves both purposes, importing from `.github/aw/` could be acceptable:

```yaml
imports:
  - ../aw/orchestration.md  # Hybrid: safe-outputs + instructions
```

**Guidelines for Hybrid Content**:
- ✅ DO: If file provides safe-outputs configuration + runtime guidance
- ✅ DO: Document that the file is designed for both uses
- ❌ DON'T: For pure agent prompts or meta-instructions
- ❌ DON'T: If content is only relevant during workflow creation

### Recommendation for Hybrid Content

1. **Prefer duplication over confusion**: Create separate files in `shared/` for workflow imports
2. **Document dual purpose**: If using `.github/aw/` for imports, clearly mark files as "importable"
3. **Consider migration path**: If many workflows need this, move to `shared/` permanently

## Conclusion

While workflows CAN import from `.github/aw/`, the recommendation depends on content type:

- **Pure agent configuration**: Keep in `.github/aw/`, do NOT import
- **Hybrid content** (safe-outputs + instructions): Consider `shared/` wrapper or document as importable
- **Reusable components**: Always use `shared/` directory

The `shared/` directory remains the proper location for most reusable workflow components. Use `.github/aw/` primarily for agent configuration, with exceptions for well-documented hybrid content that serves both agent creation and runtime purposes.
