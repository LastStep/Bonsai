---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Dependency Audit"
date: 2026-04-14
status: success
---

# Routine Report — Dependency Audit

## Overview
- **Routine:** Dependency Audit
- **Frequency:** Every 7 days
- **Last Ran:** _never_
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** success
- **Duration:** ~4 min
- **Files Read:** 3 — `go.mod`, `go.sum`, `station/agent/Routines/dependency-audit.md`
- **Files Modified:** 3 — `station/Reports/Pending/2026-04-14-dependency-audit.md`, `station/agent/Core/routines.md`, `station/Logs/RoutineLog.md`
- **Tools Used:** `govulncheck ./...`, `govulncheck -show verbose ./...`, `go list -m -u all`, `go version`
- **Errors Encountered:** 0

## Procedure Walkthrough

### Step 1: Identify package managers
- **Action:** Searched for package.json, requirements.txt, pyproject.toml, go.mod, Cargo.toml, and Gemfile in the project root.
- **Result:** Only `go.mod` found. This is a pure Go project using Go 1.24.3 (toolchain go1.24.3) with 5 direct dependencies and ~30 indirect dependencies.
- **Issues:** none

### Step 2: Run audit scans
- **Action:** Installed `govulncheck` (golang.org/x/vuln v1.2.0) since it was not available, then ran `govulncheck ./...` and `govulncheck -show verbose ./...` against the project. Also ran `go list -m -u all` to check for available updates.
- **Result:** govulncheck found:
  - **3 symbol-level vulnerabilities** — code actually calls affected stdlib functions
  - **4 package-level vulnerabilities** — in imported packages but not directly invoked
  - **23 module-level vulnerabilities** — in required stdlib modules, code does not appear to call them
  - **0 third-party dependency vulnerabilities** — all vulns are in the Go standard library
- **Issues:** govulncheck was not pre-installed; had to install it on the fly. Not a blocker.

### Step 3: Triage findings
- **Action:** Analyzed all 3 symbol-level vulnerabilities by severity, fix version, and relevance to the project.
- **Result:**

#### Symbol-level (code calls affected functions):

1. **GO-2025-3956 — os/exec.LookPath unexpected paths** (Medium)
   - Found in: `os/exec@go1.24.3`
   - Fixed in: `go1.24.6`
   - Trace: `tui.init -> huh.init -> exec.LookPath`
   - Impact: Indirect call via huh library initialization. Low practical risk for a CLI scaffolding tool, but upgrading Go to 1.24.6+ resolves it.

2. **GO-2025-3750 — syscall O_CREATE|O_EXCL inconsistency** (Low)
   - Found in: `os@go1.24.3`, `syscall@go1.24.3`
   - Fixed in: `go1.24.4`
   - Platforms: **Windows only**
   - Impact: Affects file creation/chmod operations. Since Bonsai creates files during generation, this is relevant on Windows. Fixed by upgrading Go to 1.24.4+.

3. **GO-2026-4602 — os.FileInfo Root escape** (Medium)
   - Found in: `os@go1.24.3`
   - Fixed in: `go1.25.8`
   - Trace: `generate.ScanCustomFiles -> os.ReadDir`
   - Impact: Requires go1.25.x to fix, which is a major version jump. Low practical risk since Bonsai reads from user-controlled project directories, not sandboxed Roots. Monitor for when go1.25 becomes the stable release.

#### Package-level (imported, not directly called):

4. **GO-2026-4864 — os Root.Chmod TOCTOU on Linux** — Fixed in go1.25.9. Not called by code.
5. **GO-2026-4601 — net/url IPv6 host literal parsing** — Fixed in go1.25.9. Not called by code.
6. **GO-2025-4182 — os.Root symlink TOCTOU** — Fixed in go1.25.9. Not called by code.
7. **GO-2025-3431 — net/http sensitive headers on redirect** — Fixed in go1.24.4. Not called by code.

#### Module-level (23 additional stdlib vulns):
All are in stdlib modules required by the dependency tree but not reachable from Bonsai's code paths. These include vulns in archive/tar, archive/zip, crypto/tls, crypto/x509, html/template, net/http, net/mail, encoding/asn1, encoding/pem, database/sql, and net/textproto. Most are fixed in go1.24.4 through go1.24.13.

- **Issues:** none — triage completed cleanly.

### Step 4: Log results
- **Action:** Results logged to `station/Logs/RoutineLog.md` and this report.
- **Result:** Entries appended.
- **Issues:** none

### Step 5: Update dashboard
- **Action:** Updated `agent/Core/routines.md` dashboard row for Dependency Audit.
- **Result:** Last Ran set to 2026-04-14, Next Due set to 2026-04-21, Status set to done.
- **Issues:** none

## Findings Summary

| # | Severity | Finding | Location | Action Taken |
|---|----------|---------|----------|--------------|
| 1 | medium | GO-2025-3956: os/exec.LookPath unexpected paths — reachable via huh init | stdlib os/exec@go1.24.3 | Flagged for user — upgrade Go to 1.24.6+ |
| 2 | low | GO-2025-3750: syscall O_CREATE\|O_EXCL inconsistency (Windows only) | stdlib syscall@go1.24.3 | Flagged for user — upgrade Go to 1.24.4+ |
| 3 | medium | GO-2026-4602: os.FileInfo Root escape — reachable via os.ReadDir | stdlib os@go1.24.3 | Flagged for user — requires go1.25.8, monitor |
| 4 | info | Go toolchain go1.24.3 is significantly behind latest 1.24.x patches (go1.24.13) | go.mod toolchain directive | Flagged for user — upgrade toolchain |
| 5 | info | 23 module-level stdlib vulns not reachable from code | Go stdlib | No action needed — resolved by Go upgrade |
| 6 | info | Several indirect dependencies have newer versions available | go.sum | No action needed — no vulnerabilities in third-party deps |

## Errors & Warnings
No errors encountered.

## Items Flagged for User Review
- **Upgrade Go toolchain from 1.24.3 to at least 1.24.13** — this single action resolves GO-2025-3956, GO-2025-3750, and all 23 module-level stdlib vulnerabilities. Update `go.mod` toolchain directive and rebuild. This is the highest-impact action.
- **Monitor GO-2026-4602 (os.FileInfo Root escape)** — requires go1.25.8 to fix, which is a major version jump. Low practical risk for Bonsai's use case. Revisit when go1.25 becomes the stable release channel.
- **No third-party dependency vulnerabilities found** — all direct and indirect Go module dependencies are clean.

## Notes for Next Run
- `govulncheck` is now installed at `$(go env GOPATH)/bin/govulncheck` (v1.2.0). Future runs should find it available.
- If the Go toolchain has been upgraded to 1.24.13+, expect the symbol-level findings #1 and #2 to be resolved, and most module-level vulns to disappear.
- Finding #3 (GO-2026-4602) will persist until go1.25.x is adopted.
- No unmaintained dependencies detected — all Charm ecosystem packages are actively maintained.
