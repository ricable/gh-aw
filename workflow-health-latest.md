# Workflow Health Dashboard - 2026-02-20

## Overview
- **Total workflows**: 153 (153 executable, 36 shared includes excluded)
- **Healthy**: 150 (98%)
- **Warning**: 3 (2%) — lockdown token failures + intermittent safe_outputs error
- **Critical**: 0 (0%)
- **Compilation coverage**: 153/153 (100% ✅)
- **Outdated lock files**: 15 (MD newer than lock — down 1 from yesterday's 16)
- **Overall health score**: 85/100 (↓ -3 from yesterday's 88)

## Status: GOOD — Active P1 Issue Ongoing

P1 lockdown token issue persists and has expanded to include Issue Monster (runs every 30 min).

### Health Assessment Summary

- ✅ **0 compilation failures** (all 153 workflows compile successfully)
- ✅ **100% compilation coverage** (153/153 lock files present)
- ⚠️ **15 outdated lock files** (MD modified after lock compilation — needs recompile)
- ❌ **P1: Lockdown token missing** — 3 workflows failing
- ⚠️ **P2: Duplicate Code Detector intermittent** — succeeded today, but alternating pattern

## Critical Issues 🚨

### ✅ NO CRITICAL P0 ISSUES

## Warnings ⚠️

### [P1] Lockdown Token Missing — 3 Workflows Failing (Score: 40/100)
- **Affected workflows**: Issue Monster (every 30min!), PR Triage Agent (every 6h), Daily Issues Report Generator (daily)
- **Issue**: #16776 (updated today with Issue Monster escalation)
- **Error**: `GH_AW_GITHUB_TOKEN` not configured; lockdown mode requires custom token
- **Impact**: ~50+ failures/day (Issue Monster 30-min schedule × ~46 failures/day)
- **Fix**: Set `GH_AW_GITHUB_TOKEN` repository secret
- **Runs**: [§22215346914](https://github.com/github/gh-aw/actions/runs/22215346914), [§22213875229](https://github.com/github/gh-aw/actions/runs/22213875229)

### [P2] Duplicate Code Detector — Intermittent safe_outputs FORBIDDEN (Score: 65/100)
- **Status**: Succeeded today (2026-02-20T03:58)! But was failing on 2/19 and 2/18
- **Pattern**: Alternating success/failure — may be transient API error
- **Issue**: #16778 (updated today noting today's success)
- **Recommendation**: Monitor for 3+ more cycles before closing

## Healthy Workflows ✅

**150 workflows (98%)** operating normally.

## Outdated Lock Files (15)

MD files modified after their lock files — should be recompiled with `make recompile`:
- ai-moderator.md
- bot-detection.md
- cli-consistency-checker.md
- daily-firewall-report.md
- daily-multi-device-docs-tester.md
- daily-regulatory.md
- daily-security-red-team.md
- functional-pragmatist.md
- glossary-maintainer.md
- lockfile-stats.md
- notion-issue-summary.md
- prompt-clustering-analysis.md
- tidy.md
- video-analyzer.md
- workflow-skill-extractor.md

## Systemic Issues

### Lockdown Token Missing (P1 — Ongoing from Feb 18)
- **Affected workflows**: 3 confirmed failing; 15+ at risk (have lockdown: true)
- **Root cause**: `GH_AW_GITHUB_TOKEN` / `GH_AW_GITHUB_MCP_SERVER_TOKEN` not set
- **Escalation**: Issue Monster failures now ~46/day (every 30 min schedule)
- **Action**: Issue #16776 updated with Issue Monster impact
- **Resolution path**: Set `GH_AW_GITHUB_TOKEN` repository secret

## Actions Taken This Run

- Added comment to issue #16776 escalating Issue Monster impact
- Added comment to issue #16778 noting Duplicate Code Detector may be self-healing
- Identified 15 outdated lock files (down from 16 yesterday — 1 recompiled)

## Engine Distribution (153 executable workflows)
- Copilot: ~72 (47%)
- Claude: ~31 (20%)
- Codex: ~9 (6%)
- Other/Unspecified: ~41 (27%)

---
> Last updated: 2026-02-20T07:30:55Z (run 22215406972)
> Next check: 2026-02-21 (daily schedule)
