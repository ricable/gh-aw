# Workflow Health Dashboard - 2026-02-19

## Overview
- **Total workflows**: 152 (152 executable)
- **Healthy**: 149 (98%)
- **Warning**: 3 (2%) — failing scheduled runs
- **Critical**: 0 (0%)
- **Compilation coverage**: 152/152 (100% ✅)
- **Outdated lock files**: 16 (MD newer than lock)
- **Overall health score**: 88/100 (↓ -7 from 95)

## Status: GOOD with Notable Issues

System is mostly healthy. Three scheduled workflows failing today.

### Health Assessment Summary

- ✅ **0 compilation failures** (all 152 workflows compile successfully)
- ✅ **100% compilation coverage** (152/152 lock files present)
- ⚠️ **16 workflows with outdated lock files** (MD modified after lock compilation)
- ❌ **3 scheduled workflow failures** (2 lockdown auth, 1 safe_outputs permission)

## Critical Issues 🚨

### ✅ NO CRITICAL ISSUES (P0)

## Warnings ⚠️

### [P1] PR Triage Agent + Daily Issues Report Generator - Lockdown Token Missing (Score: 40/100)
- **Status:** Failing (PR Triage: 2+ consecutive days; Daily Issues: today)
- **Error:** Lockdown mode enabled but GH_AW_GITHUB_TOKEN/GH_AW_GITHUB_MCP_SERVER_TOKEN not set
- **Impact:** PR categorization paused, daily issue analytics not running
- **Action:** Issue created - fix requires setting GH_AW_GITHUB_TOKEN secret
- **Runs:** 22171130204 (PR Triage), 22165491563 (Daily Issues)

### [P2] Duplicate Code Detector - Safe Outputs Permission Error (Score: 60/100)
- **Status:** Failing today
- **Error:** FORBIDDEN when assigning Copilot to issue #16739 via GraphQL (replaceActorsForAssignable)
- **Impact:** Duplicate code analysis not publishing results
- **Action:** Issue created for investigation
- **Run:** 22167968348

## Healthy Workflows ✅

**149 workflows (98%)** operating normally today.

## Systemic Issues

### Lockdown Mode Token Missing
- **Affected workflows:** 2 (PR Triage Agent, Daily Issues Report Generator)
- **Pattern:** Both have `tools.github.lockdown: true` requiring custom token
- **Root cause:** GH_AW_GITHUB_TOKEN secret not configured in repository
- **Other workflows at risk:** 15 additional workflows with lockdown enabled may fail if triggered

## Outdated Lock Files (16)

These MD files were modified after their lock files were compiled. They should be recompiled:
- agent-performance-analyzer.md
- archie.md
- breaking-change-checker.md
- daily-cli-tools-tester.md
- daily-observability-report.md
- dev-hawk.md
- docs-noob-tester.md
- mcp-inspector.md
- org-health-report.md
- poem-bot.md
- repo-audit-analyzer.md
- schema-consistency-checker.md
- sergo.md
- smoke-copilot.md
- static-analysis-report.md
- ubuntu-image-analyzer.md

## Actions Taken This Run

- Created 2 new issues (P1 lockdown token issue, P2 safe_outputs permission)
- Identified 16 outdated lock files
- No existing health issues needed updating

## Engine Distribution (152 executable workflows)
- Copilot: ~72 (47%)
- Claude: ~31 (20%)
- Codex: ~9 (6%)
- Other/Unspecified: ~40 (26%)

---
> Last updated: 2026-02-19T07:32:00Z (run 22172610317)
> Next check: 2026-02-20 (daily schedule)
