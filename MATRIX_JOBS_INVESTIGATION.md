# Investigation: Matrix Jobs and Job Dependencies for Project Boards and Orchestration

## Executive Summary

After investigating the codebase, **matrix jobs and enhanced job capabilities like `needs` are already supported** in the schema and partially implemented. The `needs` field is fully functional for custom jobs, allowing complex job dependency graphs. However, **matrix strategy is defined in the schema but not actively processed by the compiler**.

**Conclusion**: Neither project boards nor orchestration patterns would benefit from matrix jobs. These patterns are fundamentally dynamic and agentic, requiring runtime decision-making that is incompatible with matrix's static, compile-time parallelization.

## Current Capabilities

### 1. Job Dependencies (`needs`)

**Status**: ✅ **Fully Supported**

Custom jobs can already define dependencies using the `needs` field:

```yaml
jobs:
  analyze:
    runs-on: ubuntu-latest
    needs: pre_activation
    steps:
      - run: echo "Analysis"
  
  report:
    runs-on: ubuntu-latest
    needs: [analyze, agent]
    steps:
      - run: echo "Report"
```

**Implementation Details**:
- Defined in `main_workflow_schema.json` schema
- Processed in `pkg/workflow/compiler_jobs.go` (`buildCustomJobs` function)
- Supports both single string and array format
- Automatic dependency on `activation` job if no explicit needs
- Cycle detection and validation via `JobManager.ValidateDependencies()`

**Test Case**: `.github/workflows/test-custom-jobs-needs.md` demonstrates complex job dependencies with outputs.

### 2. Matrix Strategy

**Status**: ⚠️ **Schema Defined, Not Implemented**

The schema includes `strategy` field for matrix jobs:

```json
"strategy": {
  "type": "object",
  "description": "Matrix strategy for the job",
  "additionalProperties": false
}
```

However, the compiler does **not** currently parse or process the `strategy` field from custom jobs.

## Use Case Analysis

### ProjectOps Scenarios

**Current Pattern**: Sequential workflow execution
- Agent analyzes issues/PRs
- Agent calls `update_project` multiple times in sequence
- Each project update is a separate safe-output operation

**Potential Matrix Enhancement**:
```yaml
jobs:
  triage:
    strategy:
      matrix:
        repository: [repo-a, repo-b, repo-c]
        priority: [high, medium, low]
    runs-on: ubuntu-latest
    steps:
      - name: Triage ${{ matrix.repository }} with ${{ matrix.priority }}
        run: echo "Processing..."
```

**Assessment**: ❌ **Not Applicable**
- ProjectOps workflows are agentic - they make dynamic decisions based on content analysis
- Matrix jobs require predefined, static combinations known at compile time
- The agent needs to dynamically decide which projects, which items, and which fields to update
- Matrix would force premature decisions about what to process

### Orchestration Scenarios

**Current Pattern**: Dynamic worker dispatch
- Orchestrator analyzes work items
- Orchestrator calls `dispatch_workflow` for each unit of work
- Workers process independently and report back

**Example from docs**:
```yaml
safe-outputs:
  dispatch-workflow:
    workflows: [repo-triage-worker, dependency-audit-worker]
    max: 10
```

**Potential Matrix Enhancement**:
```yaml
jobs:
  dispatch_workers:
    strategy:
      matrix:
        worker: [triage-worker, audit-worker, security-worker]
        environment: [prod, staging]
    runs-on: ubuntu-latest
    steps:
      - name: Dispatch ${{ matrix.worker }} for ${{ matrix.environment }}
```

**Assessment**: ❌ **Not Applicable**
- Orchestration requires dynamic decision-making about which workers to dispatch
- Work items discovered at runtime (issues, PRs, alerts)
- Matrix requires static, predefined combinations
- The orchestrator's value is in intelligent routing, not parallel fan-out

## When Matrix Jobs WOULD Be Useful

Matrix jobs excel at **parallelizing predetermined, homogeneous tasks**:

### ✅ Good Use Cases (NOT current gh-aw patterns):
1. **Multi-platform testing**: Test across OS/version combinations
2. **Batch processing**: Process known list of items in parallel
3. **Multi-region deployment**: Deploy to predefined regions
4. **Cross-compilation**: Build for multiple architectures

### ❌ Poor Fit for gh-aw:
1. **Content-based routing**: Decisions based on issue/PR content
2. **Dynamic discovery**: Unknown number of items at compile time
3. **Conditional processing**: Whether to process depends on analysis
4. **Adaptive workflows**: Behavior changes based on intermediate results

## Recommendations

### 1. For ProjectOps: No Matrix Support Needed

**Rationale**:
- ProjectOps is inherently dynamic and content-driven
- Matrix requires static, compile-time decisions
- Current safe-output pattern (`update_project` called multiple times) is correct
- Agent intelligently determines what to update based on analysis

**Keep Current Pattern**:
```yaml
safe-outputs:
  update-project:
    max: 50  # Dynamic limit, agent decides how many
```

### 2. For Orchestration: No Matrix Support Needed

**Rationale**:
- Orchestration value is in intelligent dispatch decisions
- Workers are dispatched based on runtime analysis
- Matrix would remove the "agentic" aspect
- Current `dispatch-workflow` with `max` limit is correct

**Keep Current Pattern**:
```yaml
safe-outputs:
  dispatch-workflow:
    workflows: [worker-a, worker-b]
    max: 10  # Agent decides which workers and how many
```

### 3. Documentation Enhancement: Clarify Existing Capabilities ✅ DONE

**Completed Actions**:
1. ✅ Added comprehensive documentation of `needs` support in `docs/src/content/docs/reference/frontmatter.md`
2. ✅ Added examples of complex job dependency graphs with multiple dependencies and outputs
3. ✅ Added explanation of dynamic vs static parallelization in `docs/src/content/docs/patterns/orchestration.md`
4. ✅ Created test workflow demonstrating custom job dependencies

**Example from Documentation**:
```yaml
jobs:
  # First job: fetch external data
  fetch_data:
    needs: activation
    runs-on: ubuntu-latest
    outputs:
      api_url: ${{ steps.fetch.outputs.url }}
    steps:
      - name: Fetch configuration
        id: fetch
        run: |
          echo "url=https://api.example.com/data" >> $GITHUB_OUTPUT
  
  # Second job: depends on data fetch
  validate_data:
    needs: fetch_data
    runs-on: ubuntu-latest
    steps:
      - name: Validate URL
        run: echo "Validating ${{ needs.fetch_data.outputs.api_url }}"
  
  # Third job: runs after agent completes
  report:
    needs: [agent, fetch_data]
    if: always()
    runs-on: ubuntu-latest
    steps:
      - name: Generate report
        run: echo "Report using ${{ needs.fetch_data.outputs.api_url }}"
```

### 4. Future Consideration: Static Multi-Target Workflows

**Scenario**: If a genuine use case emerges for static parallel processing:

```yaml
# Hypothetical: Multi-repo audit (static list)
jobs:
  audit:
    strategy:
      matrix:
        repo: [repo-a, repo-b, repo-c]
    runs-on: ubuntu-latest
    steps:
      - name: Audit ${{ matrix.repo }}
        run: echo "Auditing..."
```

**Implementation Effort**: Medium
- Parse `strategy.matrix` from custom jobs in `buildCustomJobs`
- Validate matrix dimensions
- Keep as-is (GitHub Actions handles matrix expansion)
- Update documentation

**Priority**: Low - No current use cases identified

## Key Findings

1. ✅ **`needs` is already fully supported** - now documented with examples
2. ⚠️ **Matrix is in schema but not needed** - no compelling use cases found
3. ✅ **Current patterns are correct** - dynamic, agentic workflows don't benefit from static matrix
4. ✅ **Documentation updated** - users now know about `needs` support

## Changes Made

### Code Changes
- ✅ Added test workflow: `.github/workflows/test-custom-jobs-needs.md`
- ✅ Verified test workflow compiles successfully with correct job dependencies

### Documentation Changes
- ✅ Updated `docs/src/content/docs/reference/frontmatter.md`:
  - Added "Job Dependencies (`needs`)" section with comprehensive examples
  - Documented multi-dependency syntax and job output references
  - Added best practices for job dependencies
- ✅ Updated `docs/src/content/docs/patterns/orchestration.md`:
  - Added "Dynamic vs. Static Parallelization" section
  - Explained why matrix jobs are not suitable for agentic orchestration
  - Clarified when to use custom jobs with dependencies

## Bottom Line

**The question "could both project boards and orchestration benefit from matrix jobs" is answered: No.**

These patterns are fundamentally dynamic and benefit from existing capabilities:
- **`needs`** - for custom job dependencies ✅ Already supported, now documented
- **Custom jobs** - for pre/post-processing ✅ Already supported
- **Safe-outputs** - for dynamic agent decisions ✅ Already supported
- **Matrix jobs** - ❌ Not needed, incompatible with agentic patterns

The investigation revealed that the requested capability (`needs`) **already exists and works perfectly**. The gap was in documentation, not functionality. Matrix jobs remain unnecessary for agentic workflows.
