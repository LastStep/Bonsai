---
tags: [plan, tier-1, followup, knock-off]
description: Plan 31 + Plan 29 review followup bundle — cosmetic + test-gap + security-hardening items, single dispatch.
status: Draft
---

# Plan 32 — Plan 31 + 29 review followup bundle

**Tier:** 1
**Status:** Draft
**Agent:** general-purpose (single dispatch)

## Goal

Knock off non-blocking review followups carried over from Plan 31 (PR #75) + Plan 29 (PR #72). Same shape as Plan 26 — bundle 4–6 phase-scoped fixes, single dispatch, single PR. Closes 11 Backlog items across 6 tags.

## Context

Plan 31 ship (v0.3.0, PRs #75/#76/#77, 2026-04-24) + Plan 29 ship (PR #72, 2026-04-23) generated review minors filed to Backlog as:

- `[Plan-29-cosmetic]` — 3 items
- `[Plan-29-test-gap]` — 2 items
- `[Plan-29-security-hardening]` — 4 items (item 4 unicode lookalikes is purely speculative — skipped)
- `[Plan-31-cosmetic]` — 6 items (1, 2, 6 skipped per "local convention wins" / "speculative")
- `[Plan-31-test-gap]` — 2 items
- `[Plan-31-security-hardening]` — 4 items (item 3 TOCTOU + item 4 covered by item 1 — folded)

Net: 13 items addressed, 4 deferred-as-speculative, 2 folded into others.

Per memory feedback: "Autonomous Tier-1 bundle dispatch for small P2 knock-offs" — bundle 3-5 well-scoped independent items, single dispatch, same-session merge.

## Steps

### Phase A — Shared workspace validator (Plan-29-cosmetic 3 + Plan-29-security-hardening 1)

1. **Extract** `invalidWorkspaceReason` from `internal/tui/addflow/ground.go:289` into a shared package. New file: `internal/tui/addflow/ground.go` keeps it but `internal/tui/initflow/vessel.go:111-138` swaps its inline scan to call the shared helper.
   - Decision: keep the helper in `addflow` and import from `initflow` would create a cycle. Instead, **move** `invalidWorkspaceReason` + `NormaliseWorkspace` into a NEW package `internal/wsvalidate/` with two exports: `func InvalidReason(ws string) string` and `func Normalise(v string) string`. Both call sites import `wsvalidate`.
   - `vessel.go validate()` rewrites: trim → `wsvalidate.Normalise` → `wsvalidate.InvalidReason` → empty=valid.
   - `addflow/ground.go` Update branch around line 128 swaps to `wsvalidate.InvalidReason(norm)`.
2. **Update** the user-facing error string in `wsvalidate.InvalidReason` for the absolute-path case: replace `"no leading /"` with `"absolute paths not allowed (no leading / or drive letter)"` so Windows users (`C:\foo`) aren't misled.

### Phase B — Defence-in-depth additions to wsvalidate (Plan-29-security-hardening 2 + 3)

1. **Reject backslash** — in `wsvalidate.InvalidReason`, after the IsAbs branch and before the `..` segment scan: `if strings.ContainsRune(ws, '\\') { return "backslash not allowed (use forward slash)" }`. Catches Windows-style paths slipping past POSIX `IsAbs` since `\` is a legal POSIX literal.
2. **Reject pure root** — after the `..` segment scan: if the post-Normalise value is `./` (i.e., `Clean(input) == "."`), return `"workspace cannot be project root"`. Currently `foo/..` and `.` install station at the project root — legal but probably unintended.
3. Add 4 negative tests (rejection) + 2 positive tests (acceptance) to `internal/wsvalidate/wsvalidate_test.go`:
   - `TestInvalidReason_RejectsBackslash` — `C:\foo`, `foo\bar` → non-empty reason
   - `TestInvalidReason_RejectsPureRoot` — `./`, `.`, `foo/..` → "cannot be project root"
   - `TestInvalidReason_AcceptsNestedRelative` — `nested/path/`, `./foo/`, `foo/../bar/` → empty (Clean-reduces to safe)
   - Existing addflow + initflow callers' tests stay green via re-exported convenience symbols if needed.

### Phase C — Test strengthening (Plan-29-cosmetic 1, 2 + Plan-29-test-gap 1, 2)

1. **`TestConflicts_ColorTonesDifferPerAction`** at `internal/tui/addflow/conflicts_test.go:386-403` — strengthen by also comparing Keep vs Backup (currently only Keep vs Overwrite, which differ in label-text alone). Add a Keep-vs-Backup pair assertion. The combined test now proves at least one differing palette-tone-only pair (Backup uses a different ANSI sequence than Keep even though label text could overlap with another column).
2. **`shortName` helper rename** at `conflicts_test.go:377` → `conflictsShortName`. Update the two call sites at lines 364 + 368.
3. **Vessel acceptance test** — add `TestVessel_AcceptsCleanRelative` to `internal/tui/initflow/vessel_test.go` covering `./foo`, `foo/../bar` (must validate-true after Normalise). Mirror in `internal/tui/addflow/ground_test.go` as `TestGround_AcceptsNestedRelative` covering same cases via the addflow path.
4. **Inverse chrome test** — add `TestGenerateStage_BodyOnlyDropsChrome`'s positive companion: `TestGenerateStage_DefaultIncludesChrome` in `internal/tui/initflow/generate_test.go`. Asserts that with `SetBodyOnly(false)` (default), the rendered output contains the rail glyphs OR the `BONSAI 一` footer string.

### Phase D — generate.go consolidations (Plan-31-cosmetic 4 + 5)

1. **`hasAbility` → `slices.Contains`** at `internal/generate/generate.go:1572-1581`. Replace the 6-line linear-search helper with `slices.Contains` at the two call sites (1654, 1668). Delete `hasAbility`. Add `slices` to imports.
2. **Consolidate `agentsToSlice` / `requiredToSlice`** in `internal/generate/catalog_snapshot.go:69-96`. Replace with one `compatToSlice(a catalog.AgentCompat, omitEmpty bool) []string`. When `omitEmpty=true`, empty-Names returns `nil` (was `requiredToSlice`); when `false`, returns `[]string{}` (was `agentsToSlice`). Update all 14 call sites.

### Phase E — catalog_snapshot test gaps (Plan-31-test-gap 1 + 2)

1. Add `TestWriteCatalogSnapshot_TrailingNewline` to `internal/generate/catalog_snapshot_test.go` — write snapshot, read back bytes, assert `bytes.HasSuffix(data, []byte("\n"))`.
2. Add `TestSerializeCatalog_VersionPassThrough` — table-driven cases for `""`, `"dev"`, `"v0.3.0"`. Each: `SerializeCatalog(cat, version)` → unmarshal → assert `snap.Version == version`.

### Phase F — ProjectConfig.Validate() + symlink rejection (Plan-31-security-hardening 1 + 2)

1. **Add `func (c *ProjectConfig) Validate() error`** to `internal/config/config.go` after `Load`. Scope:
   - `c.ProjectName` — must be non-empty after trim
   - `c.Workspace` field doesn't exist on ProjectConfig (per-agent in InstalledAgent.Workspace) — instead validate each `InstalledAgent.Workspace` via `wsvalidate.InvalidReason` (error wrapping "agent <name> workspace: <reason>")
   - `c.DocsPath` — same `wsvalidate.InvalidReason` if non-empty
   - Reject shell metachars in `c.ProjectName` + each `agentType` key + each `Workspace`: `"`, `` ` ``, `$`, `\`, newline, `]`, `)`, `[`, `(` — return error naming the offending field
2. **Wire `Validate()` into `Load`**: after the existing Unmarshal + `cfg.Agents == nil` guard, call `cfg.Validate()` and return the error. This single chokepoint catches every `.bonsai.yaml` read across the codebase (cmd/init, cmd/add, cmd/remove, cmd/update, cmd/list, cmd/catalog).
3. **Symlink-resistant write** at `internal/generate/catalog_snapshot.go:209`. Replace `os.WriteFile(absPath, data, 0644)` with `os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscall.O_NOFOLLOW, 0644)` + write + close. Add `syscall` import.
   - Note: same class exists at `generate.go:285, 303, 247, 221, 943` — out of scope for this plan (sweep deferred to a separate item if appetite later).
4. Add `TestProjectConfig_Validate_RejectsAbsoluteWorkspace` + `TestProjectConfig_Validate_RejectsShellMetacharInName` + `TestProjectConfig_Validate_AcceptsCleanConfig` to a new `internal/config/config_test.go` (does not currently exist — verify and create).
5. Add `TestWriteCatalogSnapshot_RefusesSymlink` to `internal/generate/catalog_snapshot_test.go` — pre-create `.bonsai/catalog.json` as a symlink to a tempfile; call `WriteCatalogSnapshot`; assert error returned + tempfile contents unchanged.

## Out of scope

- **Plan-31-cosmetic 1** — `0o755`/`0o644` octal — local convention is legacy octal across `generate.go:218,221,244,247,282,285,303,424,559,943` + `catalog_snapshot.go:192,209`. Out-of-scope to flip globally; not flipping just two.
- **Plan-31-cosmetic 2** — `WriteCatalogSnapshot` Action-Unchanged dedupe with `writeFile`. Speculative refactor without third caller.
- **Plan-31-cosmetic 3** — `[path](path)` link-text=URL in Bonsai Reference table at `generate.go:675-677`. Path-as-label is informative for agent readers scanning the rendered CLAUDE.md. Cosmetic and arguable. Skip.
- **Plan-31-cosmetic 6** — peer re-render perf benchmark. Needs real signal first, not premature.
- **Plan-29-security-hardening 4** — Unicode lookalike normalisation (NFKC). Speculative per source filing.
- **Plan-31-security-hardening 3** — `.bonsai/` dir TOCTOU perms. Minor, contents non-secret.
- **Plan-31-security-hardening 4** — Markdown link injection via tampered `DocsPath`. Same root cause as item 1; covered by `wsvalidate.InvalidReason` on DocsPath.
- **Generator-wide `O_NOFOLLOW` sweep** — defer to follow-up. This plan only hardens `WriteCatalogSnapshot` (the most-recently-added writer).

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

Plan touches workspace-path validation (Phases A+B) + config schema validation (Phase F) + filesystem write hardening (Phase F.3). Threat model:

- **Path traversal via `.bonsai.yaml`** — Phase F.1+F.2 closes the gap (TUI validates → save; hand-edited config bypassed validation until now).
- **Symlink-replace attack on `.bonsai/catalog.json`** — Phase F.3 closes.
- **Backslash + pure-root edge cases** — Phase B closes.

No secrets touched. No new dependencies.

## Verification

- [ ] `make build` — passes clean
- [ ] `go test ./...` — all tests green (existing + Phase B/C/E/F additions)
- [ ] `golangci-lint run` (CI v1) — no new findings
- [ ] `bonsai init` in `/tmp/test-32/` happy path — workspace validates, station/ generated
- [ ] `bonsai add` in same dir — second agent installed, peer-awareness refreshes
- [ ] Manual: hand-edit `.bonsai.yaml` to set `agents.tech-lead.workspace: "/etc/"` → `bonsai list` errors with workspace-validation message
- [ ] Manual: `cd /tmp/test-32 && rm .bonsai/catalog.json && ln -s /tmp/secret-target .bonsai/catalog.json && bonsai add <ability>` → returns error, tempfile unchanged
- [ ] All 6 CI checks green: test, lint, Analyze-Go, govulncheck, CodeQL, GitGuardian

## Dispatch shape

Single `general-purpose` agent, `isolation: "worktree"`, sequential phases A→F in one branch. Aim 6 commits (one per phase) for clean review. Single draft PR.

## Closes

- `[Plan-29-cosmetic]` items 1, 2, 3
- `[Plan-29-test-gap]` items 1, 2
- `[Plan-29-security-hardening]` items 1, 2, 3 (item 4 deferred)
- `[Plan-31-cosmetic]` items 4, 5 (items 1, 2, 3, 6 deferred)
- `[Plan-31-test-gap]` items 1, 2
- `[Plan-31-security-hardening]` items 1, 2 (items 3, 4 deferred/folded)
