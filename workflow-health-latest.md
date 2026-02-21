# Workflow Health Dashboard - 2026-02-21

## Overview
- **Total workflows**: 156 (121 executable, 35 shared includes excluded)
- **Healthy**: 115 (95%)
- **Warning**: 6 (5%) — lockdown token failures
- **Critical**: 0 (0%)
- **Compilation coverage**: 121/121 (100% ✅)
- **Outdated lock files**: 14 (down from 15 yesterday)
- **Overall health score**: 82/100 (↓ -3 from yesterday's 85)

## Status: DEGRADED — P1 Lockdown Token Still Unresolved

The GH_AW_GITHUB_TOKEN issue persists and now affects 5 workflows total (expanded from 3 yesterday).

### Health Assessment Summary

- ✅ **0 compilation failures** (all 121 executable workflows compile)
- ✅ **100% compilation coverage** (no missing lock files)
- ⚠️ **14 outdated lock files** (MD modified after lock compilation)
- ❌ **P1: Lockdown token missing** — 5 workflows failing consistently
- ✅ **Duplicate Code Detector**: 2 consecutive successes (recovering)
- ✅ **Chroma Issue Indexer**: 1 failure in last 8 runs (mostly healthy)
- ❌ **Smoke Gemini**: 100% failure on main (Gemini free-tier quota exhausted, closed by pelikhan 2026-02-21)

## Critical Issues 🚨

### ✅ NO CRITICAL P0 ISSUES

## Warnings ⚠️

### [P1] Lockdown Token Missing — 5 Workflows Failing (Score: 35/100)
- **Affected workflows**: Issue Monster, PR Triage Agent, Daily Issues Report Generator, Issue Triage Agent, Weekly Issue Summary
- **Issue**: #aw_P1lock (NEW P1 issue created this run)
- **Error**: `GH_AW_GITHUB_TOKEN` not configured; lockdown mode requires custom token  
- **Impact**: ~100+ failures/day (Issue Monster 30-min schedule dominates)
- **Fix**: Set `GH_AW_GITHUB_TOKEN` repository secret
- **Existing symptom issues**: #17387 (Issue Monster), #16801 (PR Triage Agent) — both open
- **Previous tracking issue**: #16776 (closed 2026-02-20 as not_planned)
- **Runs**: [§22252647907](https://github.com/github/gh-aw/actions/runs/22252647907), [§22251839006](https://github.com/github/gh-aw/actions/runs/22251839006)

### [P3] Smoke Gemini — Gemini API Quota Exhausted (Score: 20/100)
- **Status**: 100% failure on main (4/4 recent main-branch runs)
- **Reason**: Gemini free-tier quota exhausted (429 errors)
- **Issue**: #17034 closed by pelikhan (acknowledged/accepted)
- **Note**: Closed by human owner — not creating a new issue

## Healthy Workflows ✅

**115 workflows (95%)** operating normally.

## Outdated Lock Files (14)

MD files modified after their lock files — should be recompiled with `make recompile`:
- chroma-issue-indexer.md
- daily-choice-test.md
- daily-secrets-analysis.md
- daily-testify-uber-super-expert.md
- duplicate-code-detector.md
- github-mcp-structural-analysis.md
- gpclean.md
- mergefest.md
- pdf-summary.md
- prompt-clustering-analysis.md
- repo-audit-analyzer.md
- scout.md
- slide-deck-maintainer.md
- sub-issue-closer.md

## Improving Workflows 📈

### Duplicate Code Detector
- 2 consecutive successes (2026-02-20 and 2026-02-21 morning)
- Was alternating fail/success since Feb 14
- Issue #16778 closed; monitoring continues

### Chroma Issue Indexer
- 1 failure in last 8 runs (2/20 at 16:13)
- Mostly healthy; no new issues needed

## Systemic Issues

### Lockdown Token Missing (P1 — Ongoing from Feb 18)
- **Affected workflows**: 5 confirmed failing; 13 total have lockdown: true (8 more at risk)
- **Root cause**: `GH_AW_GITHUB_TOKEN` / `GH_AW_GITHUB_MCP_SERVER_TOKEN` not set
- **Escalation**: Issue Monster failures ~50/day (every 30 min schedule)
- **Action**: New issue created #aw_P1lock
- **Resolution path**: Set `GH_AW_GITHUB_TOKEN` repository secret

## Actions Taken This Run

- Created new P1 root cause issue #aw_P1lock for lockdown token issue (5 workflows affected)
- Added comment to issue #17387 (Issue Monster) with root cause analysis
- Added comment to issue #16801 (PR Triage Agent) with status update
- Identified 14 outdated lock files (down from 15 yesterday)
- Confirmed Duplicate Code Detector is recovering (2 consecutive successes)
- Confirmed Smoke Gemini failure acknowledged by owner (issue #17034 closed)

## Lockdown Workflows At Risk (13 total, 5 currently failing)

Workflows with `lockdown: true` that need `GH_AW_GITHUB_TOKEN`:
1. ❌ Issue Monster (failing)
2. ❌ PR Triage Agent (failing)
3. ❌ Daily Issues Report Generator (failing)
4. ❌ Issue Triage Agent (failing ~50%)
5. ❌ Weekly Issue Summary (last run failed)
6. ⚠️ Discussion Task Miner (infrequent schedule)
7. ⚠️ Grumpy Reviewer (event-triggered)
8. ⚠️ Issue Arborist (infrequent)
9. ⚠️ Org Health Report (infrequent)
10. ⚠️ Refiner (event-triggered)
11. ⚠️ Stale Repo Identifier (infrequent)
12. ⚠️ Weekly Safe Outputs Spec Review (weekly)
13. ⚠️ Workflow Generator (event-triggered)

---
> Last updated: 2026-02-21T07:21:58Z (run 22252687126)
> Next check: 2026-02-22 (daily schedule)
