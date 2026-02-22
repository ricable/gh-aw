# Workflow Health Dashboard - 2026-02-22

## Overview
- **Total workflows**: 158 executable (214 total - 56 shared includes)
- **Healthy**: 155 (98%)
- **Warning/Failing**: 3 (2%) — lockdown token failures
- **Critical**: 0 (0%)
- **Compilation coverage**: 158/158 (100% ✅)
- **Outdated lock files**: 0 (all up-to-date ✅)
- **Overall health score**: 83/100 (→ stable from yesterday's 82)

## Status: DEGRADED — P1 Lockdown Token Still Unresolved

The GH_AW_GITHUB_TOKEN issue persists. Currently confirmed 3 workflows failing consistently.

### Health Assessment Summary

- ✅ **0 compilation failures** (all 158 executable workflows compile)
- ✅ **100% compilation coverage** (no missing lock files)
- ✅ **0 outdated lock files** (all timestamps current)
- ❌ **P1: Lockdown token missing** — 3 workflows failing consistently
- ✅ **Smoke Gemini**: 1 success in 7-day window (recovered or quota reset?)
- ✅ **Duplicate Code Detector**: continuing healthy streak
- ✅ **Chroma Issue Indexer**: 2/2 success in recent window

## Critical Issues 🚨

### ✅ NO CRITICAL P0 ISSUES

## Warnings ⚠️

### [P1] Lockdown Token Missing — 3 Workflows Failing (Score: 35/100)
- **Affected workflows**: Issue Monster, PR Triage Agent, Daily Issues Report Generator
- **Tracking issue**: #17414 (open since 2026-02-21)
- **Error**: `GH_AW_GITHUB_TOKEN` not configured; lockdown mode requires custom token  
- **Impact**: ~50+ failures/day (Issue Monster 30-min schedule dominates)
- **Fix**: Set `GH_AW_GITHUB_TOKEN` repository secret
- **Latest runs**: 
  - Issue Monster: run #1997 failure [§22272702112](https://github.com/github/gh-aw/actions/runs/22272702112)
  - PR Triage Agent: failure [§22271898819](https://github.com/github/gh-aw/actions/runs/22271898819)

## Healthy Workflows ✅

**155 workflows (98%)** operating normally.

## Issues Tracked

- **#17414** [P1] Lockdown mode failing: GH_AW_GITHUB_TOKEN not configured — OPEN
- **#17387** Issue Monster failed (symptom tracking) — OPEN  
- **#16801** PR Triage Agent failed (symptom tracking) — OPEN
- **#17408** No-Op Runs tracker — normal/healthy

## Actions Taken This Run

- Added comment to #17414 with updated status
- Updated shared-alerts.md with current health metrics
- No new issues created (all issues already tracked)

## Run Info
- Timestamp: 2026-02-22T07:25:00Z
- Workflow run: [§22272741315](https://github.com/github/gh-aw/actions/runs/22272741315)
