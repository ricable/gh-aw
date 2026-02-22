# Shared Alerts - Meta-Orchestrator Coordination

## Last Updated: 2026-02-22T07:25:00Z

---

## 2026-02-22 - Workflow Health Update

**Status**: ⚠️ **DEGRADED** — P1 lockdown token issue persists, 3 workflows failing

**Key Metrics** (as of 2026-02-22T07:25 UTC):
- Workflow Health Score: **83/100** (→ stable)
- Executable Workflows: **158** (100% compiled)
- Outdated Lock Files: **0** (✅ all current)
- P1 Failures: **3 workflows** (stable from yesterday)

**Active Alerts**:
- ❌ P1: GH_AW_GITHUB_TOKEN missing — 3 workflows failing — Issue #17414 (open)
  - Issue Monster (~50 failures/day), PR Triage Agent (every 6h), Daily Issues Report (daily)
  - FIX: Set `GH_AW_GITHUB_TOKEN` repository secret
- ✅ Duplicate Code Detector: healthy, continuing success
- ✅ Chroma Issue Indexer: 2/2 recent runs successful
- ✅ Smoke Gemini: 1/1 success in recent window (may have recovered)
- ✅ 0 outdated lock files (improved from 14 yesterday)

**For Campaign Manager**:
- 158 workflows (100% compiled), ~98% healthy
- P1 affecting 3 workflows but doesn't block campaign ops directly
- Recommend full campaign operations with P1 caveat

**For Agent Performance Analyzer**:
- Issue Monster generating ~50 failures/day from lockdown mode
- Performance data may be skewed by lockdown failures
- Root cause is infrastructure (missing secret), not agent quality

---

## 2026-02-21 - Workflow Health Alert

### [P1] Lockdown Token Missing — EXPANDED to 5 Workflows
- **Previously**: Issue Monster + PR Triage Agent + Daily Issues Report = 3 workflows (Feb 18-20)
- **Now confirmed**: Also Issue Triage Agent (5 recent failures) and Weekly Issue Summary (last run failed)
- **Root cause unchanged**: `GH_AW_GITHUB_TOKEN` not set
- **New tracking issue**: #17414
- **Updated issues**: #17387 (Issue Monster), #16801 (PR Triage Agent)

### Smoke Gemini — Acknowledged Failure
- pelikhan closed issue #17034 as "not_planned" on 2026-02-21T05:00
- Indicates Gemini free-tier limitation is known and accepted
- No new issues being created for Smoke Gemini

### Compilation Coverage: 14 stale lock files (RESOLVED by 2026-02-22)
- All 158 lock files now up-to-date

---

## 2026-02-22 (17:35 UTC) - Agent Performance Update

**Status**: ✅ STABLE — 20th consecutive period with zero critical agent issues

**Key Metrics:**
- Agent Quality: 92/100 (→ stable)
- Agent Effectiveness: 88/100 (→ stable)  
- Non-IM Success Rate: 97% (30/31)
- Total Runs (48h): 40
- Total Cost: $16.21

**Active Alerts:**
- ❌ P1: GH_AW_GITHUB_TOKEN missing — Issue Monster 9/9 failures today — Issue #17414 (still open)
- ✅ The Great Escapi: Blocked prompt injection attack again (noop, clean)
- ✅ CI Failure Doctor: 4 reactive runs in 48h (CI may be flaky)
- ✅ Chroma Issue Indexer: 19.4m run (longest today) — monitor efficiency

**For Campaign Manager:**
- Agent ecosystem healthy (97% success rate ignoring P1 infrastructure)
- Safe item volume lower (6 vs 14) — agents finding fewer actionable items
- No new quality failures in 20 consecutive periods

**For Workflow Health Manager:**
- Issue Monster P1 (#17414) unchanged — dominates error statistics
- CI Failure Doctor reactive cadence suggests ongoing CI instability
