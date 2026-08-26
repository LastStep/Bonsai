---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Dependency Audit"
date: 2026-08-26
status: partial
---

# Routine Report — Dependency Audit

## Overview
- **Routine:** Dependency Audit
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial
- **Duration:** ~8 min
- **Files Read:** 5 — `/home/user/Bonsai/go.mod`, `/home/user/Bonsai/go.sum`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`, `/home/user/Bonsai/station/agent/Routines/dependency-audit.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Reports/Pending/2026-08-26-dependency-audit.md` (this report), `/home/user/Bonsai/station/agent/Core/routines.md` (dashboard update), `/home/user/Bonsai/station/Logs/RoutineLog.md` (log entry)
- **Tools Used:** `go version`, `go list -m all`, `go list -m -u all`, `npm audit --json`, `grep` on go.sum; govulncheck NOT installed (fallback method used)
- **Errors Encountered:** 1 — `govulncheck` not installed; Go vuln analysis performed via `go list -m -u all` + manual CVE cross-reference against prior audit baseline

---

## Procedure Walkthrough

### Step 1: Identify Package Managers
Checked project root for manifests:
- `go.mod` — FOUND (Go, primary)
- `website/package.json` — FOUND (npm, website only)
- `requirements.txt` / `pyproject.toml` — NOT FOUND (no Python)
- `Cargo.toml` — NOT FOUND (no Rust)
- `Gemfile` — NOT FOUND (no Ruby)

**Ecosystem count: 2** (Go + npm)

### Step 2: Run Audit Scans

**Go — govulncheck fallback:**
- `govulncheck` not installed
- Fallback: `go list -m -u all` to enumerate all modules and available updates
- Manual CVE cross-reference against prior audit (2026-05-04) baseline
- **Go version:** 1.25.9 linux/amd64 (same as last audit — no toolchain regression)
- **golang.org/x/net:** v0.53.0 (was v0.38.0 at last audit, latest is v0.58.0)
  - The v0.38.0 → v0.53.0 bump (likely via Plan 36 or subsequent dependency tidy) clears the 2 pkg-level unreachable findings noted in 2026-05-04 baseline
  - 6 stdlib module-level unreachable findings from last audit persist (they are in stdlib, not in x/net itself, and were already unreachable — no code change needed)
  - Net change: 8 unreachable findings → ~6 unreachable (2 pkg-level cleared by x/net bump)
- **golang.org/x/crypto:** v0.50.0 (latest v0.55.0) — 5 minor versions behind
- **golang.org/x/sys:** v0.43.0 (latest v0.47.0)
- **golang.org/x/term:** v0.42.0 (latest v0.45.0)
- **golang.org/x/text:** v0.36.0 (latest v0.41.0)
- **Total modules behind:** 32 (up from 23 at last audit)
- **Reachable CVEs confirmed:** 0 (consistent with prior audit; cannot confirm via govulncheck — recommend installation)

**npm (website/) — `npm audit --json`:**
- astro installed: **v6.1.7** (package.json specifies `"latest"`, resolved to 6.1.7 at install time)
- js-yaml installed: **v4.1.1**
- vite installed: **v7.3.2**
- **Total vulnerabilities: 8** (was 0 at 2026-05-04 — regression)
  - High: 7
  - Low: 1
  - Moderate: 0
  - Critical: 0

### Step 3: Triage Findings

**npm findings (all new since last audit):**

| # | Package | Severity | CVE / Advisory | Fix |
|---|---------|----------|----------------|-----|
| 1 | astro | HIGH (SSRF) | GHSA-2pvr-wf23-7pc7: Host header SSRF in prerendered error page fetch (CVSS 7.5) | Upgrade to v7.2.8 |
| 2 | astro | HIGH (XSS) | GHSA-8hv8-536x-4wqp: Reflected XSS via unescaped slot name (CVSS 7.1) | Upgrade to v7.2.8 |
| 3 | astro | MOD (XSS) | GHSA-jrpj-wcv7-9fh9: XSS via Unescaped Attribute Names in Spread Props | Upgrade to v7.2.8 |
| 4 | astro | MOD (XSS) | GHSA-f48w-9m4c-m7f5: XSS via unescaped spread attribute names (incomplete fix) | Upgrade to v7.2.8 |
| 5 | astro | MOD (XSS) | GHSA-4g3v-8h47-v7g6: Reflected XSS via unescaped View Transition animation props | Upgrade to v7.2.8 |
| 6 | astro | LOW (XSS) | GHSA-7pw4-f3q4-r2p2: XSS via unescaped transition:* directive values on hydrated islands | Upgrade to v7.2.8 |
| 7 | astro | LOW (replay) | GHSA-xr5h-phrj-8vxv: Server island encrypted parameters vulnerable to cross-component replay | Upgrade to v7.2.8 |
| 8 | js-yaml | HIGH (DoS) | GHSA-52cp-r559-cp3m: YAML merge-key chains force quadratic CPU consumption (HIGH) | v5.4.1 (breaking change) |
| 9 | js-yaml | HIGH (DoS) | GHSA-5p4m-2wfm-xmqj: Quadratic CPU consumption in !!omap resolution | v5.4.1 (breaking change) |
| 10 | js-yaml | MOD (DoS) | GHSA-h67p-54hq-rp68: Quadratic-complexity DoS in merge key handling | v5.4.1 (breaking change) |
| 11 | sharp | HIGH | GHSA-f88m-g3jw-g9cj: Inherited vulnerabilities from libvips (CVE-2026-33327, 33328, 35590, 35591) | Upgrade astro to v7.2.8 |
| 12 | nanoid | HIGH | GHSA-28wg-ghj8-5hjv / GHSA-2v37-7h3g-55p8: Generators loop indefinitely with negative/zero size | `npm audit fix` |
| 13 | postcss | HIGH | GHSA-6g55-p6wh-862q + GHSA-fxqj-rqcc-2cmp + GHSA-r28c-9q8g-f849: Path traversal via sourceMappingURL → arbitrary file disclosure | `npm audit fix` |
| 14 | svgo | HIGH | GHSA-2p49-hgcm-8545: removeScripts plugin leaves executable scripts intact | `npm audit fix` |
| 15 | vite | HIGH | GHSA-fx2h-pf6j-xcff: `server.fs.deny` bypass on Windows alternate paths | `npm audit fix` |
| 16 | esbuild | LOW | GHSA-g7r4-m6w7-qqqr: Arbitrary file read via dev server on Windows | Upgrade astro to v7.2.8 |

**npm severity summary for report metadata (by affected package, not per-advisory):**
- astro (direct): HIGH — 2 SSRF/XSS advisories with CVSS ≥ 7
- js-yaml (direct): HIGH — DoS via quadratic CPU (independent of astro)
- sharp, svgo, nanoid, postcss, vite (transitive): HIGH

**Context from prior audits:**
- 2026-04-21 audit noted "astro upgrade breaks build" as the blocker
- 2026-05-04 audit: npm was clean (3rd clean run) — these are all new regressions accumulated since May 2026
- The backlog-hygiene routine (2026-08-26) already flagged "6 Dependabot alerts, astro upgrade breaks build" — this audit confirms and expands that finding (8 npm audit findings, multiple Dependabot advisories merged)

**Go findings (using fallback — no govulncheck):**
- ~6 unreachable module-level stdlib findings persist (unchanged, require no action)
- x/net bump (v0.38.0 → v0.53.0) cleared the 2 previously-unreachable pkg-level x/net findings
- 32 modules behind (hygiene, no confirmed CVEs)
- No reachable CVEs confirmed via fallback method (consistent with last audit baseline)

### Steps 4–5: Log + Dashboard
Completed — see below.

---

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | **HIGH** | Astro SSRF in prerendered error page fetch (GHSA-2pvr-wf23-7pc7, CVSS 7.5) — direct dep, exploitable | `website/node_modules/astro v6.1.7` | Flagged for user review; fix = upgrade astro to v7.2.8 |
| 2 | **HIGH** | Astro Reflected XSS via unescaped slot name (GHSA-8hv8-536x-4wqp, CVSS 7.1) — direct dep | `website/node_modules/astro v6.1.7` | Flagged for user review |
| 3 | **HIGH** | js-yaml quadratic DoS via YAML merge-key chains (GHSA-52cp-r559-cp3m) | `website/node_modules/js-yaml v4.1.1` | Flagged for user review; fix = v5.4.1 (breaking) |
| 4 | **HIGH** | sharp inherited libvips CVEs (GHSA-f88m-g3jw-g9cj: CVE-2026-33327/33328/35590/35591) | `website/node_modules/sharp <0.35.0` | Flagged; fix = upgrade astro to v7.2.8 |
| 5 | **HIGH** | nanoid infinite loop with negative/zero size (2 advisories) | `website/node_modules/nanoid` | Flagged; `npm audit fix` may resolve |
| 6 | **HIGH** | postcss arbitrary file disclosure via sourceMappingURL (3 advisories) | `website/node_modules/postcss` | Flagged; `npm audit fix` may resolve |
| 7 | **HIGH** | svgo: removeScripts leaves executable scripts intact | `website/node_modules/svgo` | Flagged; `npm audit fix` may resolve |
| 8 | **HIGH** | vite: `server.fs.deny` bypass on Windows (GHSA-fx2h-pf6j-xcff) | `website/node_modules/vite v7.3.2` | Flagged; `npm audit fix` may resolve |
| 9 | **LOW** | esbuild: arbitrary file read via dev server on Windows (GHSA-g7r4-m6w7-qqqr) | `website/node_modules/esbuild` | Transitive via astro; fix = upgrade astro |
| 10 | **MEDIUM** | govulncheck not installed — Go vuln analysis limited to fallback heuristics | Tool gap | Flagged; recommend `go install golang.org/x/vuln/cmd/govulncheck@latest` |
| 11 | **LOW** | 32 Go modules behind latest (up from 23 last audit) | `go.mod` | Logged; hygiene only — no confirmed CVEs |
| 12 | **LOW** | golang.org/x/net v0.53.0 still 5 minor versions behind latest v0.58.0 | `go.mod` | Logged; prior pkg-level unreachable findings cleared by v0.38.0→v0.53.0 bump |

---

## Errors & Warnings

- `govulncheck` not installed — Go vulnerability scan fell back to `go list -m -u all` + prior audit cross-reference. Reachable CVE assessment is best-effort only. Recommend installing: `go install golang.org/x/vuln/cmd/govulncheck@latest`

---

## Items Flagged for User Review

### 1. [HIGH] npm: 8 vulnerabilities (7 high, 1 low) in website/ — regression since 2026-05-04

The website was clean at the last audit. Since May 2026, multiple astro and ecosystem advisories have accumulated:

**Recommended action path:**
```bash
cd website/
npm audit fix        # may resolve nanoid, postcss, svgo, vite automatically
# Then test build:
npm run build
# If build passes, also evaluate astro upgrade:
npm install astro@7.2.8
npm run build
```

**Risk note:** Prior audits (2026-04-21) flagged that an astro major-version upgrade breaks the build. The fix version here is `v7.2.8` (v7.x series). Installed version is `v6.1.7`. This is a major version jump that requires compatibility verification.

**js-yaml:** Fix is `v5.4.1` (breaking change from v4). Requires code migration if js-yaml is used directly. Check if it's only a transitive dep first.

**Urgency:** The 2 HIGH-severity astro advisories (SSRF CVSS 7.5, XSS CVSS 7.1) are on the public-facing website — these warrant priority attention even if the website is primarily a documentation site.

### 2. [MEDIUM] govulncheck not installed — Go vulnerability scan incomplete

Without govulncheck, the Go scan cannot confirm reachable CVE status. The prior audit baseline (0 reachable) held through Go 1.25.9, but new advisories may have been issued since May 2026.

**Recommended:** `go install golang.org/x/vuln/cmd/govulncheck@latest` before next audit.

### 3. [LOW] 32 Go modules behind latest

Up from 23 at last audit. No confirmed CVEs, but the gap is widening. A batch `go get -u ./...` + `go mod tidy` + test pass would close this hygiene debt. Consider bundling with the next plan that touches go.mod.

---

## Notes for Next Run

- Install govulncheck before running to restore full-fidelity Go vulnerability scanning
- npm regression is significant — aim to resolve astro upgrade before next audit (2026-09-02)
- If astro v7.x upgrade is confirmed build-breaking, evaluate pinning to a specific v6.x security patch or vendoring
- golang.org/x/net is now at v0.53.0 (was v0.38.0) — the 2 prior pkg-level unreachable findings should be cleared on next govulncheck run
- Cross-reference with vulnerability-scan routine (also overdue — last ran 2026-05-04) for SAST + secrets overlap
