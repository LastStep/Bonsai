---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Dependency Audit"
date: 2026-04-21
status: partial
---

# Routine Report — Dependency Audit

## Overview
- **Routine:** Dependency Audit
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-14
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (scans clean on Node, multiple Go stdlib + x/net CVEs found; ecosystems without tools logged N/A)
- **Duration:** ~2 minutes
- **Files Read:** 3 — `station/agent/Routines/dependency-audit.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `station/Reports/Pending/2026-04-21-dependency-audit.md` (new), `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** `govulncheck ./...`, `govulncheck -show verbose ./...`, `go list -m -u all`, `npm audit --json`, `which` (tool presence checks)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Identify package managers
- **Action:** Searched project root for manifest files and probed installed audit tooling.
- **Result:** Found `go.mod` at repo root (Go) and `website/package.json` (npm). No `requirements.txt`/`pyproject.toml`, no `Cargo.toml`, no `Gemfile`. Tools present: `govulncheck` (go1.24 scanner), `go` 1.24.13, `npm`. Absent: `pip-audit`, `cargo audit`, `safety` — logged N/A per dispatch instructions.
- **Issues:** None.

### Step 2: Run audit scans
- **Action:** Ran `govulncheck ./...` (and verbose pass), `go list -m -u all`, and `npm audit --json` inside `website/`.
- **Result:**
  - **Go vulnerabilities (govulncheck):** 11 total — 2 reachable stdlib, 3 unreachable package-level, 6 unreachable module-level. Full findings tabulated below.
  - **Go outdated modules (`go list -m -u all`):** 17 modules behind (none security-critical on their own beyond what govulncheck already flags; `golang.org/x/net v0.38.0 → v0.45.0+` doubles as the security fix for GO-2026-4441/4440).
  - **npm audit (`website/`):** 0 vulnerabilities across 455 deps (368 prod, 88 optional). Clean.
- **Issues:** None executing the scans.

### Step 3: Triage findings
- **Action:** Graded each finding by severity and reachability per the routine's triage guidance.
- **Result:**
  - **Reachable stdlib (HIGH priority, direct impact):** 2 — both fixed by bumping Go toolchain from 1.24.13 to 1.25.8+. Upgrade is low-risk (Go 1.x is backward compatible).
  - **Reachable direct module deps:** 0 (all reachable findings are stdlib).
  - **Unreachable package-level in imports:** 3 — `golang.org/x/net` (2 CVEs, fixed in v0.45.0) + 1 stdlib `internal/syscall/unix` TOCTOU on Linux. Bumping Go to 1.25.9 + `go get golang.org/x/net@latest` clears these.
  - **Unreachable module-level stdlib:** 6 — all cleared by Go 1.25.8/1.25.9 upgrade.
  - **Unmaintained deps:** None detected (all flagged modules actively maintained — `golang.org/x/*`, Charmbracelet stack, cobra, etc.).
- **Issues:** None. No critical CVEs reachable by application code; severity ranks as elevated-routine rather than incident.

### Step 4: Log results
- **Action:** Appending RoutineLog entry (see Section 4 of dispatch instructions) and writing this report.
- **Result:** Done.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Updating `last_ran` → 2026-04-21, `next_due` → 2026-04-28, `status` → `done` for Dependency Audit row in `station/agent/Core/routines.md`.
- **Result:** Done.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | HIGH (reachable) | GO-2026-4602 — FileInfo can escape from a Root in `os` (fixed in go1.25.8) | `internal/generate/scan.go:44` via `os.ReadDir` | Flagged; Go toolchain upgrade recommended. |
| 2 | HIGH (reachable) | GO-2026-4601 — Incorrect IPv6 host literal parsing in `net/url` (fixed in go1.25.8) | `cmd/guide.go:92` via `glamour.TermRenderer.Render → url.Parse` | Flagged; Go toolchain upgrade recommended. |
| 3 | MEDIUM (unreachable import) | GO-2026-4864 — TOCTOU permits root escape via `Root.Chmod` on Linux (fixed in go1.25.9) | stdlib `internal/syscall/unix` | Flagged; resolved by same Go upgrade. |
| 4 | MEDIUM (unreachable import) | GO-2026-4441 — Infinite parsing loop in `golang.org/x/net` (fixed in v0.45.0) | `golang.org/x/net@v0.38.0` | Flagged; `go get golang.org/x/net@latest` recommended. |
| 5 | MEDIUM (unreachable import) | GO-2026-4440 — Quadratic parsing complexity in `golang.org/x/net/html` (fixed in v0.45.0) | `golang.org/x/net@v0.38.0` | Flagged; same upgrade as above. |
| 6 | LOW (module-only) | GO-2026-4947 — Unexpected work during chain building in `crypto/x509` (fixed in go1.25.9) | stdlib | Flagged; cleared by Go upgrade. |
| 7 | LOW (module-only) | GO-2026-4946 — Inefficient policy validation in `crypto/x509` (fixed in go1.25.9) | stdlib | Flagged; cleared by Go upgrade. |
| 8 | LOW (module-only) | GO-2026-4870 — Unauthenticated TLS 1.3 KeyUpdate DoS in `crypto/tls` (fixed in go1.25.9) | stdlib | Flagged; cleared by Go upgrade. |
| 9 | LOW (module-only) | GO-2026-4869 — Unbounded allocation for GNU sparse in `archive/tar` (fixed in go1.25.9) | stdlib | Flagged; cleared by Go upgrade. |
| 10 | LOW (module-only) | GO-2026-4865 — JsBraceDepth context tracking XSS in `html/template` (fixed in go1.25.9) | stdlib | Flagged; cleared by Go upgrade. |
| 11 | LOW (module-only) | GO-2026-4603 — meta-content URLs not escaped in `html/template` (fixed in go1.25.8) | stdlib | Flagged; cleared by Go upgrade. |
| — | INFO | npm audit — 0 vulnerabilities across 455 deps in `website/` | `website/package.json` | None — clean. |
| — | N/A | Python ecosystem — no `requirements.txt`/`pyproject.toml` | — | Skipped (no manifest). |
| — | N/A | Rust ecosystem — no `Cargo.toml` | — | Skipped (no manifest). |
| — | N/A | Ruby ecosystem — no `Gemfile` | — | Skipped (no manifest). |

### Outdated Go modules (from `go list -m -u all`, non-security)
Notable gaps (>1 minor version behind): `golang.org/x/crypto v0.36.0 → v0.50.0`, `golang.org/x/tools v0.37.0 → v0.44.0`, `golang.org/x/sys v0.38.0 → v0.43.0`, `golang.org/x/text v0.30.0 → v0.36.0`, `golang.org/x/mod v0.28.0 → v0.35.0`, `github.com/clipperhouse/uax29/v2 v2.5.0 → v2.7.0`, `github.com/yuin/goldmark v1.7.13 → v1.8.2`. No known CVEs flagged by govulncheck for these; treat as hygiene.

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **Upgrade Go toolchain from 1.24.13 → 1.25.9 (or latest 1.25.x).** Clears 2 reachable stdlib CVEs plus 7 latent stdlib CVEs. Low-risk bump; single line change in `go.mod` (`go 1.24` directive) plus CI runner update in `.github/workflows/release.yml`. Recommend promoting to a Backlog item or Tier-1 patch this cycle.
2. **Bump `golang.org/x/net` from v0.38.0 → v0.45.0+.** Clears GO-2026-4441 and GO-2026-4440. Run `go get golang.org/x/net@latest && go mod tidy`.
3. **(Optional) Batch dependency refresh** for remaining outdated `golang.org/x/*` and charm/bubbletea ecosystem modules as a hygiene sweep once the two items above ship.

## Notes for Next Run
- Once the Go toolchain upgrade lands, govulncheck should return a clean reachable-set. Re-run to confirm and log baseline.
- If `pip-audit` / `cargo audit` ever get installed on this host and Python/Rust manifests appear, update this routine's procedure to stop logging them as N/A.
- `npm audit` on `website/` has been clean for multiple cycles — consider reducing its scan cadence to every other run if noise becomes an issue (not urgent).
