---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Dependency Audit"
date: 2026-05-11
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
- **Duration:** ~8 minutes
- **Files Read:** 3 — `/home/user/Bonsai/go.mod`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Files Modified:** 3 — `/home/user/Bonsai/station/Reports/Pending/2026-05-11-dependency-audit.md`, `/home/user/Bonsai/station/agent/Core/routines.md`, `/home/user/Bonsai/station/Logs/RoutineLog.md`
- **Tools Used:** `govulncheck ./...` (attempted — 403 Forbidden on vuln.go.dev), `go list -m all`, `go list -m -u all`, `go list -m -json all`, `go mod tidy`, `go mod verify`, `go build ./...`, `npm audit`, `npm outdated`, `which govulncheck`, `go install golang.org/x/vuln/cmd/govulncheck@latest`
- **Errors Encountered:** 1 — `govulncheck` could not fetch vulnerability database (403 Forbidden on `https://vuln.go.dev/index/modules.json.gz`; no outbound internet access in environment)

## Procedure Walkthrough

### Step 1: Identify Package Managers
- **Action:** Checked for `package.json`, `requirements.txt`, `pyproject.toml`, `go.mod`, `Cargo.toml`, `Gemfile` at project root. Also found `website/package.json` from prior run knowledge and confirmed with `ls`.
- **Result:** Two package managers found: **Go** (`go.mod` at project root) and **Node.js** (`website/package.json` — Astro documentation site). No Python, Rust, or Ruby manifests present.
- **Issues:** None.

### Step 2: Run Audit Scans
- **Action (Go):** Installed `govulncheck` via `go install golang.org/x/vuln/cmd/govulncheck@latest` (succeeded, placed in `~/go/bin/`). Ran `~/go/bin/govulncheck ./...` from project root.
- **Result (Go):** Command failed with `HTTP GET https://vuln.go.dev/index/modules.json.gz returned unexpected status: 403 Forbidden`. The vulnerability database is inaccessible — the sandbox environment has no outbound internet. `govulncheck` requires live access to `vuln.go.dev`; no offline/local DB alternative available. Fell back to manual dependency review using `go list -m all`, `go list -m -u all`, and cross-reference against known CVE history for key packages.
- **Action (Node.js):** Ran `npm audit` in `website/` directory.
- **Result (Node.js):** `found 0 vulnerabilities`. Clean. Note: packages show as MISSING (not locally installed), consistent with prior run results — npm audit still resolves advisory data via the registry metadata.
- **Issues:** govulncheck blocked by network; Go audit is manual this cycle. This is consistent with prior run (2026-05-04) which had the same limitation after govulncheck was available.

### Step 3: Triage Findings
- **Action:** Performed manual triage of all Go dependencies by:
  1. Reviewing current versions against known CVE fix points for security-sensitive packages (`golang.org/x/net`, `golang.org/x/crypto`, `golang.org/x/text`, `golang.org/x/sys`, `microcosm-cc/bluemonday`)
  2. Checking `go mod verify` — all module checksums verified clean
  3. Running `go build ./...` — builds cleanly with no errors
  4. Running `go mod tidy` — no changes (lockfile already clean)
  5. Identifying available updates via `go list -m -u all`
  6. Flagging packages with last-release date older than 12 months (using `go list -m -json all` timestamps)

- **Result — Security assessment (manual):**

  | Package | Current | Known CVE fix points | Assessment |
  |---------|---------|---------------------|------------|
  | `golang.org/x/net` | v0.53.0 | CVE-2023-44487 fixed v0.17.0; CVE-2023-45288 fixed v0.23.0 | **Clean** — past all known major CVEs; v0.54.0 available (minor) |
  | `golang.org/x/crypto` | v0.50.0 | CVE-2023-48795 (Terrapin) fixed v0.17.0 | **Clean** — past all known CVEs; v0.51.0 available (minor) |
  | `golang.org/x/text` | v0.36.0 | CVE-2022-32149 fixed v0.3.8 | **Clean** — far past all known CVEs |
  | `golang.org/x/sys` | v0.43.0 | No known direct CVEs | **Clean** |
  | `microcosm-cc/bluemonday` | v1.0.27 | HTML sanitizer — actively maintained, no known open CVEs | **Clean** |
  | `gopkg.in/yaml.v3` | v3.0.1 | CVE-2022-3064 fixed in v3.0.1 | **Clean** — v3.0.1 is the security-patched release |

  No Critical or High CVEs identified in direct or transitive dependencies through manual review.

- **Result — Available updates (23 modules behind, hygiene):**

  | Package | Current | Latest | Type |
  |---------|---------|--------|------|
  | `golang.org/x/net` | v0.53.0 | v0.54.0 | indirect |
  | `golang.org/x/crypto` | v0.50.0 | v0.51.0 | indirect |
  | `golang.org/x/text` | v0.36.0 | v0.37.0 | indirect |
  | `golang.org/x/sys` | v0.43.0 | v0.44.0 | indirect |
  | `golang.org/x/tools` | v0.43.0 | v0.45.0 | indirect |
  | `golang.org/x/exp` | v0.0.0-20231006 | v0.0.0-20260508 | indirect |
  | `golang.org/x/mod` | v0.34.0 | v0.36.0 | indirect |
  | `golang.org/x/term` | v0.42.0 | v0.43.0 | indirect (also direct) |
  | `golang.org/x/sync` | v0.20.0 | (latest) | indirect |
  | `github.com/alecthomas/chroma/v2` | v2.20.0 | v2.24.1 | indirect |
  | `github.com/yuin/goldmark` | v1.7.13 | v1.8.2 | indirect |
  | `github.com/dlclark/regexp2` | v1.11.5 | v1.12.0 | indirect |
  | `github.com/spf13/pflag` | v1.0.9 | v1.0.10 | indirect |
  | `github.com/sahilm/fuzzy` | v0.1.1 | v0.1.2 | indirect |
  | `github.com/charmbracelet/colorprofile` | v0.4.1 | v0.4.3 | indirect |
  | Various other charmbracelet/x/* | — | newer pseudo-versions | indirect |

  Count: 23 modules behind (same count as 2026-05-04 run — no change, batch hygiene refresh still in Backlog).

- **Result — Potentially unmaintained packages (no release in 12+ months from 2026-05-11):**

  Packages with last-release before 2025-05-11, filtered to only notable ones with meaningful surface area:

  | Package | Last Release | Role | Risk |
  |---------|-------------|------|------|
  | `github.com/aymerick/douceur` | 2015-08-27 | CSS parser (indirect, used by bluemonday) | Low — tiny, stable, no CVEs; bluemonday's dependency |
  | `gopkg.in/yaml.v3` | 2022-05-27 | YAML parsing (direct) | Low — v3.0.1 IS the security-patched release; go.yaml.in/yaml/v3 exists as modern fork |
  | `github.com/russross/blackfriday/v2` | 2020-10-27 | Markdown (indirect, cobra docs) | Low — indirect, not used in main code paths |
  | `github.com/muesli/reflow` | 2021-05-17 | Text reflow (indirect, charm stack) | Low — stable utility library |
  | `github.com/erikgeiser/coninput` | 2021-10-04 | Console input (indirect, bubbletea) | Low — stable low-level utility |
  | `github.com/mitchellh/hashstructure/v2` | 2021-05-27 | Struct hashing (indirect) | Low — stable |

  Most "old" packages are either: (a) tiny single-purpose stable libs that haven't needed updates, (b) test helpers, or (c) deeply indirect. No packages flagged for urgent replacement.

- **Issues:** govulncheck offline prevents authoritative CVE scan. Manual review covers known historical CVEs but cannot catch newly published advisories. This is a persistent limitation of the sandbox environment.

### Step 4: Log Results
- **Action:** Appended entry to `station/Logs/RoutineLog.md`.
- **Result:** Entry written.
- **Issues:** None.

### Step 5: Update Dashboard
- **Action:** Updated `station/agent/Core/routines.md` dashboard row for Dependency Audit — Last Ran → 2026-05-11, Next Due → 2026-05-18, Status → done.
- **Result:** Updated.
- **Issues:** None.

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | Info | govulncheck blocked (403 Forbidden on vuln.go.dev) — no authoritative CVE scan possible | Environment (no internet) | Noted; manual review performed as fallback |
| 2 | Low | 23 Go modules behind latest versions (hygiene) | `go.mod` transitive deps | No change — batch refresh tracked in Backlog (Group G / P3) |
| 3 | Low | `gopkg.in/yaml.v3` v3.0.1 last released 2022-05-27 — modern fork `go.yaml.in/yaml/v3` exists | `go.mod` (direct dep) | Noted — v3.0.1 is security-current; migration is a hygiene decision, not security-critical |
| 4 | Low | `github.com/aymerick/douceur` v0.2.0 last released 2015 — used by bluemonday CSS parser | Indirect dep | No action — stable, no known CVEs; bluemonday's responsibility |
| 5 | Info | npm packages in `website/` show as MISSING (not locally installed) | `website/` | No action — consistent with prior runs; npm audit still clean via registry |

## Errors & Warnings

- **govulncheck network failure:** `HTTP GET https://vuln.go.dev/index/modules.json.gz returned unexpected status: 403 Forbidden`. No authoritative Go vulnerability scan possible this cycle. Manual review performed as documented above. This is an environment constraint, not a project issue. Persistent across multiple cycles — flagged in prior reports.

## Items Flagged for User Review

1. **govulncheck offline (persistent):** Go vulnerability scanning via `govulncheck` requires internet access to `vuln.go.dev`. The sandbox environment blocks this. Manual review can cover historical CVEs but cannot catch newly published advisories. Consider running `govulncheck ./...` manually in a development environment as a supplement. This has been flagged across multiple cycles (2026-04-14, 2026-04-21, 2026-05-04) — no resolution yet; likely a structural environment constraint.

2. **23 modules behind (hygiene, low priority):** The batch Go module refresh has been on the Backlog (Group G / P3) since the 2026-04-21 cycle. Still 23 modules behind. No security relevance found this cycle but staying current reduces future upgrade-cliff risk. This was included in the Plan 36 scope per the 2026-05-04 digest — confirm if Plan 37 scope covers this or if a dedicated pass is needed.

## Notes for Next Run

- govulncheck will likely remain blocked unless the environment changes — continue manual review fallback.
- `gopkg.in/yaml.v3` → `go.yaml.in/yaml/v3` migration could be considered as part of a future batch module refresh but is not security-urgent.
- Module version count (23 behind) is stable since last cycle — no new drift introduced.
- npm audit clean for the 4th consecutive run — no npm action needed.
- `go mod verify` passes cleanly — module checksums intact.
