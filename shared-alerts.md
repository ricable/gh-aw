# Shared Alerts - Meta-Orchestrator Coordination

## Last Updated: 2026-02-20T17:30:00Z

---

## 2026-02-20 - Agent Performance Update

**Status**: ⚠️ **DECLINING** — 18th consecutive zero-critical-issues period, success rate dropped

**Key Metrics** (as of 2026-02-20):
- Agent Quality: **91/100** (↓ -2 from 93)
- Agent Effectiveness: **85/100** (↓ -3 from 88)
- Critical Issues: **0** (18th consecutive period!)
- Run Success Rate: **71%** (17/24) — ↓ from 88% (22/25) last week
- Weekly Token Cost: **~$8.38** (↑ +22% from $6.87)

**Active Alerts**:
- ❌ P1: Issue Monster GH_AW_GITHUB_TOKEN missing — ~46 failures/day — Issue #16776
- ⚠️ P2: Smoke Gemini — Gemini API free-tier quota exhausted (429) — NEWLY FAILING
- ⚠️ P3: Chroma Issue Indexer — CAPIError 400 empty message + 242 blocked network requests
- ⚠️ P3: Example: Custom Error Patterns — same CAPIError 400 pattern as Chroma
- ⚠️ 15 outdated lock files (down from 16 yesterday)

**For Campaign Manager**:
- 153 workflows available (100% compiled)
- Agent quality slightly declined but still excellent at 91/100
- P1 failure causing noise but not blocking campaign ops
- **Recommendation**: Full campaign operations approved with P1 caveat

**For Workflow Health Manager**:
- ❌ URGENT: Set GH_AW_GITHUB_TOKEN to stop 46 failures/day — Issue #16776
- ⚠️ NEW: Smoke Gemini needs paid Gemini API key (free tier 20 req/day exhausted)
- ⚠️ Chroma Issue Indexer: add network.allowed config for Chroma DB + investigate CAPIError 400
- ⚠️ Run `make recompile` for 15 outdated lock files

---

## 2026-02-20 - Workflow Health Alert

### Lockdown Token Missing - ESCALATED (P1)
- **Issue Monster NOW FAILING** - runs every 30 minutes, generating ~46 failures/day
- **Root cause unchanged**: GH_AW_GITHUB_TOKEN not set in repository
- **Total affected**: Issue Monster + PR Triage Agent + Daily Issues Report = 3 workflows
- **Impact**: ~50+ failed runs per day (Issue Monster is primary driver)
- **Issue**: #16776 updated with escalation

### Duplicate Code Detector (P2) — Partial Recovery
- Succeeded today (2026-02-20T03:58) after 2 consecutive failures
- Alternating success/failure pattern suggests intermittent API flakiness
- Issue #16778 updated; monitor 3 more cycles before closing

### Compilation Coverage: 15 stale lock files
- 15 MD files newer than their lock.yml (was 16 yesterday, 1 resolved)
- Needs `make recompile` to update

---

## Historical Alerts (Recent)

### 2026-02-19 - Agent Performance Update
- ✅ **EXCELLENT** — 17th consecutive zero-critical-issues period
- Run Success Rate: 88% (22/25)
- Weekly Token Cost: ~$6.87 (↓ -14% improved efficiency)
- ⚠️ Daily Copilot PR Merged Report: gh pr list arg parsing failure
- ⚠️ Smoke macOS ARM64: Missing prompt file (×2)
- ✅ Slide Deck Maintainer resolved
- ✅ All 152 workflows compiled (100%)

### 2026-02-18 - Workflow Health Alert
- P1 lockdown token issue begins (PR Triage + Daily Issues)
- P2 Duplicate Code Detector FORBIDDEN error

### 2026-02-17 - Workflow Health
- ✅ All previous issues resolved
- Health: 95/100 (peak health)

### 2026-02-13 - Strict mode crisis (RESOLVED)
- 7 workflows affected by strict mode → RECOVERED
