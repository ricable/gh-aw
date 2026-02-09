# Agentics Collection - Fixed Workflows

This directory contains the fixed versions of workflows from the [githubnext/agentics](https://github.com/githubnext/agentics) repository after applying syntax updates for the latest gh-aw version.

## What Was Fixed

All workflows in this directory have been updated to use the latest gh-aw syntax:

1. **Deprecated `bash:` syntax** → `bash: true`
   - 9 workflows affected
   
2. **Deprecated `add-comment.discussion` field** → removed
   - 5 workflows affected

## How to Apply These Fixes

To apply the same fixes to the agentics repository:

```bash
cd /path/to/agentics
gh aw fix --write
gh aw compile
```

## Verification

All workflows have been verified to compile successfully:

```bash
gh aw compile --validate --strict
✓ Compiled 18 workflow(s): 0 error(s), 11 warning(s)
```

## Report

See the full analysis report: [agentics-syntax-check-2026-02-09.md](../agentics-syntax-check-2026-02-09.md)

## Date

Fixed: 2026-02-09
