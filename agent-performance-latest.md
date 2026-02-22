# Agent Performance Analysis - 2026-02-22

**Run:** [§22281821807](https://github.com/github/gh-aw/actions/runs/22281821807)  
**Status:** ✅ STABLE — 20th consecutive zero-critical-issues period; non-IM success rate 97%  
**Analysis Period:** February 21-22, 2026 (~48-hour window)

## Executive Summary

- **Agent Quality:** 92/100 (→ stable)
- **Agent Effectiveness:** 88/100 (→ stable)
- **Critical Agent Issues:** 0 (20th consecutive period 🎉)
- **Run Success Rate (non-IM):** 97% (30/31 completed) ↑ from 89%
- **Total Tokens:** 36.6M | **Estimated Cost:** ~$16.21 (48h window)
- **Total Runs:** 40 (31 non-IM success + 9 Issue Monster failures)
- **Total Safe Items:** 6 (↓ from 14 — lighter daily cadence or fewer actionable findings)

## Key Metrics

| Metric | Current | Previous | Change |
|--------|---------|----------|--------|
| Agent Quality | 92/100 | 92/100 | → stable |
| Agent Effectiveness | 88/100 | 88/100 | → stable |
| Non-IM Success Rate | 97% (30/31) | 89% (16/18) | ↑ +8% |
| Critical Issues | 0 | 0 | ✅ 20th period |
| Session Cost | ~$16.21 | ~$6.35 (partial day) | (48h window) |
| Safe Items | 6 | 14 | ↓ -57% (fewer actions needed) |
| Avg Run Duration | 7.1m | ~7m | → stable |

## 🛡️ Standout: The Great Escapi — Prompt Injection Blocked Again

A prompt injection attack disguised as "security testing" was again detected and blocked.
The agent correctly identified prohibited actions (sandbox escape, DNS tunneling, network
evasion, reconnaissance) and filed a clean noop. **Security posture: Excellent.**

## 🔥 P1 Still Burning: Issue Monster (9/9 failures today)

GH_AW_GITHUB_TOKEN still missing. 9 failures in 48h window.
Tracking issue: #17414 (open).

## Top Performing Agents

1. **AI Moderator ×3 (93/100):** 3/3 success, 2 turns each, ~200K tokens/run — highest efficiency
2. **The Great Escapi (95/100):** Security agent, blocked prompt injection, 75K tokens, clean noop
3. **CI Failure Doctor ×4 (91/100):** 4 reactive runs, all success — CI health maintainer
4. **Daily Safe Outputs Conformance Checker (90/100):** 8.6m, clean run
5. **Contribution Check (89/100):** 4.5m, clean
6. **Semantic Function Refactoring (87/100):** 7.4m, Claude engine

## Agents With Issues

1. **Issue Monster (0/100 this period — infrastructure):** 9/9 failures, GH_AW_GITHUB_TOKEN missing
   - Not a quality issue; P1 infrastructure (#17414) still unresolved
   - Generates ~50+ failures/day; skews overall success metrics

## Observations

- CI Failure Doctor ran 4× in 48h — CI instability persists
- Chroma Issue Indexer: 19.4m (longest run) — worth monitoring for efficiency
- Daily Security Red Team Agent: 14.0m (2nd longest) — expected for deep analysis
- Safe items fell to 6 from 14; likely fewer actionable findings in this period
- Engine distribution: claude (8 runs), copilot (11 runs), codex (4 runs)

## Active Issues

- ❌ **P1:** [#17414](https://github.com/github/gh-aw/issues/17414) — GH_AW_GITHUB_TOKEN missing
