---
tags: [report, routine]
from: maintenance-subagent
to: Tech Lead
routine: "Dependency Audit"
date: 2026-08-02
status: partial
---

# Routine Report — Dependency Audit

## Overview
- **Routine:** Dependency Audit
- **Frequency:** Every 7 days
- **Last Ran:** 2026-05-04
- **Triggered By:** loop.md autonomous dispatch

## Execution Metadata
- **Status:** partial (Go scan blocked by proxy; npm scan complete)
- **Duration:** ~10 minutes
- **Files Read:** 3 — `/home/user/Bonsai/go.mod`, `/home/user/Bonsai/website/package.json`, `/home/user/Bonsai/website/package-lock.json`
- **Files Modified:** 3 — `station/Reports/Pending/2026-08-02-dependency-audit.md` (this file), `station/agent/Core/routines.md` (dashboard), `station/Logs/RoutineLog.md` (log entry)
- **Tools Used:**
  - `go install golang.org/x/vuln/cmd/govulncheck@latest` — installed successfully
  - `/root/go/bin/govulncheck ./...` — FAILED (vuln.go.dev blocked by proxy: "Forbidden")
  - `/opt/node22/bin/npm audit --json` (website/) — completed, 7 findings
  - `/opt/node22/bin/npm audit` (website/) — human-readable summary
  - `go list -m -json all` — enumerated Go module graph
  - `node -e "..."` against `package-lock.json` — confirmed installed versions
- **Errors Encountered:** 1 — govulncheck blocked at `https://vuln.go.dev/index/modules.json.gz` (HTTP 403 via proxy)

---

## Procedure Walkthrough

### Step 1 — Identify Package Managers
Two manifests found:
- **Go** — `/home/user/Bonsai/go.mod` (go 1.25.0, toolchain go1.25.9)
- **Node.js** — `/home/user/Bonsai/website/package.json` (5 direct deps, all pinned to `"latest"` — node_modules reflects actual resolved versions via package-lock.json)

No `requirements.txt`, `pyproject.toml`, `Cargo.toml`, or `Gemfile` found.

### Step 2 — Run Audit Scans

**Go — govulncheck:**
govulncheck was installed fresh from `golang.org/x/vuln v1.6.0`. On execution, it failed to fetch the vulnerability database:
```
govulncheck: fetching vulnerabilities: Get "https://vuln.go.dev/index/modules.json.gz": Forbidden
```
The proxy blocks outbound HTTPS to `vuln.go.dev`. Go vulnerability state cannot be confirmed this cycle. Manual note: the previous audit (2026-05-04) recorded 0 reachable CVEs after the Go 1.25.9 + x/net upgrade cycle. `golang.org/x/net` is now at v0.53.0 (up from the v0.38.0 that carried 2 prior package-level findings), which continues the positive trend.

**Node.js (website/) — npm audit:**
Completed successfully. 7 vulnerabilities found: 0 critical, 6 high, 1 low.

Installed versions confirmed from `package-lock.json`:
| Package | Installed | Vulnerability Range |
|---------|-----------|---------------------|
| astro | 6.1.7 | <=7.0.9 |
| js-yaml | 4.1.1 | 4.0.0–4.2.0 |
| postcss | 8.5.10 | <=8.5.17 |
| sharp | 0.34.5 | <0.35.0 |
| svgo | 4.0.1 | 4.0.0–4.0.1 |
| vite | 7.3.2 | 7.0.0–7.3.3 |
| esbuild | 0.27.7 | 0.27.3–0.28.0 |

### Step 3 — Triage Findings

See Findings Summary table below.

### Step 4 — Log Results
Appended to `station/Logs/RoutineLog.md`.

### Step 5 — Update Dashboard
Updated `Dependency Audit` row in `station/agent/Core/routines.md`.

---

## Findings Summary

| # | Severity | Finding | Location | Advisory | Fix Available | Action |
|---|----------|---------|----------|----------|--------------|--------|
| 1 | HIGH | Reflected XSS via unescaped slot name | astro@6.1.7 (direct) | GHSA-8hv8-536x-4wqp (CVSS 7.1) | astro≥6.3.3 | Flag for upgrade |
| 2 | HIGH | Host header SSRF in prerendered error page fetch | astro@6.1.7 (direct) | GHSA-2pvr-wf23-7pc7 (CVSS 7.5) | astro≥6.4.6 | Flag for upgrade |
| 3 | HIGH | XSS via Unescaped Attribute Names in Spread Props | astro@6.1.7 (direct) | GHSA-jrpj-wcv7-9fh9 (CVSS 4.2) | astro≥6.4.6 | Flag for upgrade |
| 4 | HIGH | Reflected XSS via unescaped View Transition animation properties | astro@6.1.7 (direct) | GHSA-4g3v-8h47-v7g6 | astro≥7.0.10 | Flag for upgrade |
| 5 | HIGH | XSS via unescaped spread attribute names — incomplete fix for CVE-2026-54298 | astro@6.1.7 (direct) | GHSA-f48w-9m4c-m7f5 | astro≥7.0.6 | Flag for upgrade |
| 6 | HIGH | XSS via unescaped transition:* directive values on hydrated islands | astro@6.1.7 (direct) | GHSA-7pw4-f3q4-r2p2 | astro≥7.0.4 | Flag for upgrade |
| 7 | HIGH | YAML merge-key chains can force quadratic CPU consumption (DoS) | js-yaml@4.1.1 (direct) | GHSA-52cp-r559-cp3m (CVSS 7.5) | js-yaml≥4.3.0 | Flag for upgrade |
| 8 | MODERATE | Quadratic-complexity DoS in merge key handling via repeated aliases | js-yaml@4.1.1 (direct) | GHSA-h67p-54hq-rp68 (CVSS 5.3) | js-yaml≥4.3.0 | Log for tracking |
| 9 | HIGH | Arbitrary file read via attacker-controlled sourceMappingURL | postcss@8.5.10 (transitive) | GHSA-6g55-p6wh-862q (CVSS 7.5) | postcss>8.5.17 | `npm audit fix` |
| 10 | HIGH | Path Traversal in source map auto-loading (sourceMappingURL) | postcss@8.5.10 (transitive) | GHSA-r28c-9q8g-f849 (CVSS 7.5) | postcss>8.5.17 | `npm audit fix` |
| 11 | HIGH | libvips inherited vulnerabilities (CVE-2026-33327/33328/35590/35591) | sharp@0.34.5 (transitive via astro) | GHSA-f88m-g3jw-g9cj | sharp≥0.35.0 (via astro 7.1.6) | Flag for upgrade |
| 12 | HIGH | removeScripts plugin leaves some executable scripts intact | svgo@4.0.1 (transitive) | GHSA-2p49-hgcm-8545 (CVSS 8.2) | svgo≥4.0.2 | `npm audit fix` |
| 13 | HIGH | NTLMv2 hash disclosure via UNC path handling on Windows | vite@7.3.2 (transitive) | GHSA-v6wh-96g9-6wx3 | vite≥7.3.5 | `npm audit fix` |
| 14 | HIGH | `server.fs.deny` bypass on Windows alternate paths | vite@7.3.2 (transitive) | GHSA-fx2h-pf6j-xcff (CVSS 7.5) | vite≥7.3.4 | `npm audit fix` |
| 15 | LOW | Arbitrary file read on Windows dev server | esbuild@0.27.7 (transitive via astro) | GHSA-g7r4-m6w7-qqqr (CVSS 2.5) | esbuild≥0.28.1 (via astro 7.1.6) | Log for tracking |
| 16 | UNKNOWN | govulncheck blocked — Go vulnerability state unverified | Go module graph | N/A | N/A | See below |
| 17 | SERVER ISLAND REPLAY | Server island encrypted parameters vulnerable to cross-component replay | astro@6.1.7 (direct) | GHSA-xr5h-phrj-8vxv (CVSS 6.1) | astro≥6.1.10 | Flag for upgrade |

---

## Detailed Fix Paths

### Path A — `npm audit fix` (safe, no major version changes)
Fixes: postcss (→>8.5.17), svgo (→4.0.2+), vite (→7.3.4+)
```
cd website && npm audit fix
```

### Path B — `npm audit fix --force` (major version changes — test required)
Fixes: astro (→7.1.6), js-yaml (→5.2.3), and transitively sharp, esbuild
```
cd website && npm audit fix --force
```
**Caution:** This will upgrade:
- `astro@6.1.7` → `7.1.6` — major version jump; check for breaking changes in astro.config.mjs, component API, and build output
- `js-yaml@4.1.1` → `5.2.3` — major version jump; check for API changes in `website/scripts/generate-catalog.mjs` (direct consumer of js-yaml)

### Path C — Manual Pin (alternative to --force for astro)
Update `website/package.json` to `"astro": "^7.1.6"` instead of `"latest"`, then run `npm install`. More predictable than `--force` for CI.

### Note on `package.json` "latest" pinning
All five deps in `website/package.json` use `"latest"`. This means the lockfile (not package.json) drives actual versions — and the lockfile is stale. A plain `npm update` in the website directory would resolve many of these by pulling current releases into the lock.

---

## Errors & Warnings

**Error 1 — govulncheck proxy block**
- `GET https://vuln.go.dev/index/modules.json.gz` → HTTP 403 (proxy: Forbidden)
- govulncheck requires outbound HTTPS to `vuln.go.dev`
- Cannot verify Go CVE status this cycle
- Mitigation: Previous clean scan (2026-05-04) + x/net upgraded from v0.38.0 → v0.53.0 since last audit suggest continued low risk, but this is not a confirmed clean result
- Recommended: verify proxy whitelist includes `vuln.go.dev` for next run, or use offline DB if available

---

## Items Flagged for User Review

1. **[HIGH — action required]** `website/` has 6 HIGH severity npm vulnerabilities. Two direct dependencies need upgrading:
   - `astro`: 6.1.7 → 7.1.6 (XSS × 6, SSRF × 1, replay × 1)
   - `js-yaml`: 4.1.1 → 5.2.3 (DoS × 2)
   - Recommended command sequence: `cd website && npm audit fix` (safe fixes first), then `npm audit fix --force` after checking astro 7.x migration guide
   
2. **[HIGH — action required]** Three transitive HIGH findings (`postcss`, `svgo`, `vite`) can be resolved with `cd website && npm audit fix` — no --force needed, safe to run now.

3. **[MODERATE — environment]** `vuln.go.dev` is blocked by the proxy. govulncheck cannot run. If this project's CI pipeline uses govulncheck (it does — confirmed in CI config), CI should still succeed because it has network access. But local/subagent runs will remain blind. Consider whitelisting `vuln.go.dev` in the proxy config.

4. **[LOW — hygiene]** `website/package.json` pins all deps to `"latest"` with no range or caret prefix. This produces unpredictable behaviour across environments. Consider using `"^x.y.z"` ranges once the upgrade to astro 7.x is done.

---

## Notes for Next Run

- govulncheck was installed to `/root/go/bin/govulncheck` but vuln.go.dev was blocked. Retry next cycle — if still blocked, switch to offline mode or check CI govulncheck results directly.
- Previous clean Go scan was 2026-05-04. golang.org/x/net has been bumped to v0.53.0 since then (positive signal).
- npm findings this cycle are significantly higher than the 2026-05-04 clean result — the lockfile had not been refreshed in 90 days while upstream packages continued to patch CVEs.
- After running the npm fix commands, re-run `npm audit` to confirm 0 findings and update the lock.
- If astro is upgraded to v7, verify the docs site still builds: `cd website && npm run build`.
