# Shared Alerts - Meta-Orchestrator Coordination

## Last Updated: 2026-02-21T07:22:00Z

---

## 2026-02-21 - Workflow Health Update

**Status**: ⚠️ **DEGRADED** — P1 lockdown token issue expanded to 5 workflows, Smoke Gemini acknowledged failure

**Key Metrics** (as of 2026-02-21T07:22 UTC):
- Workflow Health Score: **82/100** (↓ -3 from 85)
- Executable Workflows: **121** (100% compiled)
- Outdated Lock Files: **14** (down from 15)
- P1 Failures: **5 workflows** (up from 3 yesterday)

**Active Alerts**:
- ❌ P1: GH_AW_GITHUB_TOKEN missing — 5 workflows failing — New issue #aw_P1lock
  - Issue Monster (~50 failures/day), PR Triage Agent (every 6h), Daily Issues Report (daily), Issue Triage Agent (daily), Weekly Issue Summary
  - FIX: Set `GH_AW_GITHUB_TOKEN` repository secret
- ✅ Duplicate Code Detector: 2 consecutive successes — RECOVERING
- ✅ Chroma Issue Indexer: Mostly healthy (1/8 recent failures)
- ❌ Smoke Gemini: 100% failure on main, issue #17034 closed by pelikhan (accepted)
- ⚠️ 14 outdated lock files — needs `make recompile`

**For Campaign Manager**:
- 121 workflows (100% compiled), ~95% healthy
- P1 affecting 5 workflows but doesn't block campaign ops directly
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
- **New tracking issue**: #aw_P1lock
- **Updated issues**: #17387 (Issue Monster), #16801 (PR Triage Agent)

### Smoke Gemini — Acknowledged Failure
- pelikhan closed issue #17034 as "not_planned" on 2026-02-21T05:00
- Indicates Gemini free-tier limitation is known and accepted
- No new issues being created for Smoke Gemini

### Compilation Coverage: 14 stale lock files
- 14 MD files newer than their lock.yml (was 15 yesterday, 1 resolved)
- Needs `make recompile` to update

---

## Previous Alerts (2026-02-20)

### Agent Performance Update (2026-02-20)
- Agent Quality: 91/100 (↓ -2 from 93)
- Run Success Rate: 71% (17/24) — ↓ from 88%
- Weekly Token Cost: ~$8.38 (↑ +22%)
- P1: Issue Monster GH_AW_GITHUB_TOKEN missing — Issue #16776 (now closed, replaced by #aw_P1lock)
- P2: Smoke Gemini — Gemini API free-tier quota exhausted (now acknowledged/closed)

