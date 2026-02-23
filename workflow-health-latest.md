# Workflow Health Dashboard - 2026-02-23

## Overview
- **Total workflows**: 158 executable (214 total - 56 shared includes)
- **Healthy**: 155 (98%)
- **Warning/Failing**: 3 (2%) — lockdown token failures (persistent)
- **Critical**: 0 (0%)
- **Compilation coverage**: 158/158 (100% ✅)
- **Outdated lock files**: 0 (all up-to-date ✅ — 14 apparent outdated files are false positives from git checkout timing)
- **Overall health score**: 82/100 (↓ 1 from 83 — P1 persists, #17414 closed not_planned)

## Status: DEGRADED — P1 Lockdown Fix Available but Not Applied

The GH_AW_GITHUB_TOKEN tracking issue #17414 was closed as "not_planned" on Feb 22.
A fix is available in #17807 (remove `lockdown: true`), but not yet applied.

### Health Assessment Summary

- ✅ **0 compilation failures** (all 158 executable workflows compile)
- ✅ **100% compilation coverage** (no missing lock files)
- ✅ **0 outdated lock files** (all git timestamps current)
- ❌ **P1: Lockdown token missing** — 3 workflows failing (>1 week streak)
- ✅ **GitHub Remote MCP Auth Test**: 93% success rate overall (1 transient failure today)
- ✅ **Go Logger Enhancement**: 97% success rate (1 transient failure yesterday)
- ✅ **All smoke tests**: passing (Copilot, Claude, Codex, Gemini, Multi-PR)
- ✅ **Metrics Collector**: running successfully

## Critical Issues 🚨

### ✅ NO CRITICAL P0 ISSUES

## Warnings ⚠️

### [P1] Lockdown Token Missing — 3 Workflows Failing (Score: 35/100)
- **Affected workflows**: Issue Monster, PR Triage Agent, Daily Issues Report Generator
- **Tracking issue**: #17414 — CLOSED as "not_planned" on 2026-02-22
- **Fix available**: #17807 — patch to remove `lockdown: true`, compiled & ready
- **Error**: `GH_AW_GITHUB_TOKEN` not configured; explicit `lockdown: true` fails validation
- **Impact**: ~50+ failures/day (Issue Monster 30-min schedule dominates)
- **Recommended action**: Apply patch from #17807 (removes `lockdown: true` → uses automatic detection)
- **Latest runs**:
  - Issue Monster: run #2038 failure [§22296800867](https://github.com/github/gh-aw/actions/runs/22296800867)
  - PR Triage Agent: run #125 failure [§22295410778](https://github.com/github/gh-aw/actions/runs/22295410778)

### [P3] Minor Transient Failures (Not Actionable)
- GitHub Remote MCP Auth Test: 1 failure today (run #49) — 93% historical success rate
- Go Logger Enhancement: 1 failure yesterday (run #157) — 97% historical success rate
- Both resolve naturally on next run

## Healthy Workflows ✅

**155 workflows (98%)** operating normally.

Verified healthy today:
- All smoke tests (Copilot, Claude, Codex, Gemini)
- Agentic Maintenance, Auto-Triage Issues, Bot Detection
- Chroma Issue Indexer, Duplicate Code Detector
- CI Cleaner, Contribution Check, Static Analysis Report
- Metrics Collector - Infrastructure Agent

## Issues Tracked

- **#17414** [P1] Lockdown mode failing — CLOSED as "not_planned" (2026-02-22)
- **#17807** Fix: remove lockdown:true — OPEN (patch ready to apply)
- **#17387** Issue Monster failed (symptom tracking) — OPEN (updated with fix reference)
- **#16801** PR Triage Agent failed (symptom tracking) — OPEN
- **#17408** No-Op Runs tracker — normal/healthy

## Actions Taken This Run

- Added comment to #17387 noting fix available in #17807
- Updated shared-alerts.md with current health metrics
- Noted #17414 closed as "not_planned" — root cause unchanged

## Run Info
- Timestamp: 2026-02-23T07:40:00Z
- Workflow run: [§22296860578](https://github.com/github/gh-aw/actions/runs/22296860578)
- Health score: 82/100 (↓ 1 from yesterday's 83)
