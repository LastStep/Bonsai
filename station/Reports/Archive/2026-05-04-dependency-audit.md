---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Dependency Audit"
date: 2026-05-04
status: success
---

# Routine Report — Dependency Audit

## Overview
- **Routine:** Dependency Audit
- **Frequency:** Every 7 days
- **Last Ran:** 2026-04-21
- **Triggered By:** Tech Lead batched routine dispatch

## Execution Metadata
- **Status:** success (no reachable vulnerabilities; only unreachable imports + module-level findings remain; ecosystems without tools logged N/A)
- **Duration:** ~2 minutes
- **Files Read:** 2 — `station/agent/Routines/dependency-audit.md`, `station/Reports/Archive/2026-04-21-dependency-audit.md`
- **Files Modified:** 1 — `station/Reports/Pending/2026-05-04-dependency-audit.md` (new)
- **Tools Used:** `govulncheck ./...`, `govulncheck -show verbose ./...`, `go list -m -u all`, `npm audit --json`, `which` (tool presence checks)
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Identify package managers
- **Action:** Searched project root for manifest files and probed installed audit tooling.
- **Result:** Found `go.mod` at repo root (Go) and `website/package.json` (npm). No `requirements.txt`/`pyproject.toml`, no `Cargo.toml`, no `Gemfile`. Tools present: `govulncheck` v1.2.0 (DB updated 2026-04-21), `go` 1.25.8, `npm` (Node v22.21.1). Absent: `pip-audit`, `cargo audit`, `safety` — logged N/A per dispatch instructions.
- **Issues:** None.

### Step 2: Run audit scans
- **Action:** Ran `govulncheck ./...` (and verbose pass), `go list -m -u all`, and `npm audit --json` inside `website/`.
- **Result:**
  - **Go vulnerabilities (govulncheck):** 8 total — 0 reachable (down from 2 last cycle), 3 unreachable package-level, 5 unreachable module-level. Full findings tabulated below.
  - **Go outdated modules (`go list -m -u all`):** 23 modules behind. `golang.org/x/net v0.38.0 → v0.53.0` doubles as the security fix for GO-2026-4441/4440. No additional security-critical gaps beyond govulncheck flags.
  - **npm audit (`website/`):** 0 vulnerabilities across 455 deps (368 prod, 88 optional). Clean — third consecutive clean run.
- **Issues:** None executing the scans.

### Step 3: Triage findings
- **Action:** Graded each finding by severity and reachability per the routine's triage guidance and cross-referenced against the 2026-04-21 report.
- **Result:**
  - **Reachable stdlib (HIGH priority, direct impact):** 0 (down from 2). Both prior reachable CVEs (GO-2026-4602, GO-2026-4601) — RESOLVED by the Go 1.24.13 → 1.25.8 toolchain upgrade that landed since last audit.
  - **Reachable direct module deps:** 0.
  - **Unreachable package-level in imports:** 3 — `golang.org/x/net@v0.38.0` (2 CVEs, fixed in v0.45.0+) + 1 stdlib `internal/syscall/unix` TOCTOU on Linux (fixed in go1.25.9). All persistent from last cycle.
  - **Unreachable module-level stdlib:** 5 — all cleared by a go1.25.8 → go1.25.9 (or later) bump. Persistent from last cycle.
  - **Unmaintained deps:** None detected (all flagged modules actively maintained — `golang.org/x/*`, Charmbracelet stack, cobra, chroma, etc.).
- **Issues:** None. No critical CVEs reachable by application code; severity drops to "hygiene" rather than the elevated-routine grade from last cycle.

### Step 4: Log results
- **Action:** Per dispatch instructions, NOT updating `station/Logs/RoutineLog.md` — Tech Lead will batch-update after all routines complete.
- **Result:** Report written; routine log untouched.
- **Issues:** None.

### Step 5: Update dashboard
- **Action:** Per dispatch instructions, NOT updating `station/agent/Core/routines.md` — Tech Lead will batch-update after all routines complete.
- **Result:** Dashboard untouched.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Status vs 2026-04-21 |
|---|----------|---------|----------|----------------------|
| 1 | MEDIUM (unreachable import) | GO-2026-4864 — TOCTOU permits root escape via `Root.Chmod` on Linux (fixed in go1.25.9) | stdlib `internal/syscall/unix` | Persistent |
| 2 | MEDIUM (unreachable import) | GO-2026-4441 — Infinite parsing loop in `golang.org/x/net` (fixed in v0.45.0) | `golang.org/x/net@v0.38.0` | Persistent |
| 3 | MEDIUM (unreachable import) | GO-2026-4440 — Quadratic parsing complexity in `golang.org/x/net/html` (fixed in v0.45.0) | `golang.org/x/net@v0.38.0` | Persistent |
| 4 | LOW (module-only) | GO-2026-4947 — Unexpected work during chain building in `crypto/x509` (fixed in go1.25.9) | stdlib | Persistent |
| 5 | LOW (module-only) | GO-2026-4946 — Inefficient policy validation in `crypto/x509` (fixed in go1.25.9) | stdlib | Persistent |
| 6 | LOW (module-only) | GO-2026-4870 — Unauthenticated TLS 1.3 KeyUpdate DoS in `crypto/tls` (fixed in go1.25.9) | stdlib | Persistent |
| 7 | LOW (module-only) | GO-2026-4869 — Unbounded allocation for GNU sparse in `archive/tar` (fixed in go1.25.9) | stdlib | Persistent |
| 8 | LOW (module-only) | GO-2026-4865 — JsBraceDepth context tracking XSS in `html/template` (fixed in go1.25.9) | stdlib | Persistent |
| — | RESOLVED | GO-2026-4602 — FileInfo can escape from a Root in `os` | stdlib (was reachable via `internal/generate/scan.go`) | **Resolved** by Go 1.25.8 upgrade |
| — | RESOLVED | GO-2026-4601 — Incorrect IPv6 host literal parsing in `net/url` | stdlib (was reachable via `cmd/guide.go`) | **Resolved** by Go 1.25.8 upgrade |
| — | RESOLVED | GO-2026-4603 — meta-content URLs not escaped in `html/template` | stdlib | **Resolved** by Go 1.25.8 upgrade |
| — | INFO | npm audit — 0 vulnerabilities across 455 deps in `website/` | `website/package.json` | Clean (3rd consecutive run) |
| — | N/A | Python ecosystem — no `requirements.txt`/`pyproject.toml` | — | Skipped (no manifest) |
| — | N/A | Rust ecosystem — no `Cargo.toml` | — | Skipped (no manifest) |
| — | N/A | Ruby ecosystem — no `Gemfile` | — | Skipped (no manifest) |

### Outdated Go modules (from `go list -m -u all`, non-security)
Notable gaps (>1 minor version behind, or large drift): `golang.org/x/crypto v0.36.0 → v0.50.0`, `golang.org/x/tools v0.37.0 → v0.44.0`, `golang.org/x/text v0.30.0 → v0.36.0`, `golang.org/x/mod v0.28.0 → v0.35.0`, `golang.org/x/sync v0.17.0 → v0.20.0`, `github.com/alecthomas/chroma/v2 v2.20.0 → v2.24.1`, `github.com/yuin/goldmark v1.7.13 → v1.8.2`, `github.com/aymanbagabas/go-udiff v0.3.1 → v0.4.1`, `github.com/dlclark/regexp2 v1.11.5 → v1.12.0`, `github.com/spf13/pflag v1.0.9 → v1.0.10`. Charmbracelet `x/exp/*` modules also have newer pseudo-versions. No CVEs flagged by govulncheck for these — treat as hygiene.

### Cross-Reference vs 2026-04-21 (Resolved / Persistent / New)
- **Resolved (3):** GO-2026-4602, GO-2026-4601, GO-2026-4603 — all cleared by the Go 1.24.13 → 1.25.8 toolchain bump that landed since last audit. The two reachable stdlib CVEs are gone.
- **Persistent (8):** GO-2026-4864, GO-2026-4441, GO-2026-4440, GO-2026-4947, GO-2026-4946, GO-2026-4870, GO-2026-4869, GO-2026-4865 — all unreachable, all clear with go1.25.9 + `golang.org/x/net@v0.45.0+`.
- **New (0):** No new CVEs reported by govulncheck since 2026-04-21.

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review

1. **Bump Go toolchain go1.25.8 → go1.25.9 (or latest 1.25.x).** Clears 6 of 8 remaining stdlib findings (one package-level + five module-level). Single-line edit to `go.mod` (`go 1.25.0` directive + `toolchain go1.25.8`) plus CI runner update. All findings are unreachable, so this is hygiene-grade; no incident urgency.
2. **Bump `golang.org/x/net` from v0.38.0 → v0.45.0+ (latest is v0.53.0).** Clears the remaining 2 unreachable package-level findings (GO-2026-4441, GO-2026-4440). Run `go get golang.org/x/net@latest && go mod tidy`. Also lifts a notable outdated-module gap.
3. **(Optional) Batch dependency refresh** — 23 modules behind, mostly `golang.org/x/*` and Charmbracelet stack. Suggest a single sweep PR after the two items above ship to bring everything current. No CVEs flagged for any of these by govulncheck.

## Notes for Next Run
- Reachable-set is now clean — this baseline confirms the Go 1.25.8 upgrade landed correctly. Future runs should treat any new reachable finding as a regression.
- Items 1 and 2 above are persistent across two cycles now; consider creating a Backlog entry to track them rather than re-flagging each routine run.
- If `pip-audit` / `cargo audit` ever get installed on this host and Python/Rust manifests appear, update this routine's procedure to stop logging them as N/A.
- `npm audit` on `website/` — 3 consecutive clean runs. Could reduce cadence, but cost is negligible (~2s).
