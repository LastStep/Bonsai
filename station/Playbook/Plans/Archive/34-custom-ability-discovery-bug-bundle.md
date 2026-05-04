---
tags: [plan, bug-fix]
description: Fix bonsai update silent-skip of pre-registered custom abilities + sensor frontmatter parsing + non-TTY invalid-file UX + tech-lead doc that invites manual .bonsai.yaml edits.
---

# Plan 34 — Custom Ability Discovery Bug Bundle

**Tier:** 1
**Status:** Complete
**Agent:** code (general-purpose, single worktree)

## Goal

Fix the bug chain where the tech-lead agent created custom abilities, manually registered names in `.bonsai.yaml`, and `bonsai update` silently skipped discovery — leaving files un-lock-tracked with empty CLAUDE.md descriptions. Plus two adjacent gaps (sensor frontmatter, non-TTY invalid-file silence) and the doc that invited the wrong workflow.

## Background

User report 2026-05-04. Reproduced in `/tmp/bonsai-bug-test/` and `/tmp/bonsai-bug-test2/`. Four independent fixes, all small, all in the same area of code/docs.

## Steps

### Step 1 — Doc fix: `catalog/skills/bonsai-model/bonsai-model.md`

Add an explicit "do not manually edit `.bonsai.yaml`" rule. The current text ("edit-safe", "if Bonsai state, belongs in .bonsai.yaml") leads agents to register abilities by hand.

Edits in `bonsai-model.md`:

1. **Line 79** — change:
   ```
   - **`.bonsai.yaml`** — user-owned config. Lists which agents exist + which abilities each has installed. Edit-safe.
   ```
   to:
   ```
   - **`.bonsai.yaml`** — config Bonsai writes for you. Lists which agents exist + which abilities each has installed. **Read-only for agents** — never hand-edit the `skills/workflows/protocols/sensors/routines` lists or `custom_items` map; let `bonsai update` register custom files.
   ```

2. **In the "Custom files" subsection (around line 120-130)** — after the existing "drop file → run `bonsai update`" instructions, add a callout:
   ```
   > [!warning]
   > **Do not manually register custom abilities in `.bonsai.yaml`.** If you add a name to `skills:` (or any other list) without `bonsai update` discovering it, the file is never lock-tracked and `custom_items[name]` is never populated — CLAUDE.md will render the row with an empty description and the file will silently desync from the lockfile. Always: drop file with frontmatter, then `bonsai update`.
   ```

3. **Line 201** — change:
   ```
   - Don't create shadow configuration outside `.bonsai.yaml` — if it's Bonsai state, it belongs in `.bonsai.yaml`.
   ```
   to:
   ```
   - Don't create shadow configuration outside `.bonsai.yaml` — but don't hand-edit `.bonsai.yaml` either. Bonsai state belongs in `.bonsai.yaml`, and **only `bonsai` commands write it**.
   ```

### Step 2 — Code fix 2A: orphaned-registration recovery in `internal/generate/scan.go`

Current behaviour ([scan.go:50-64](../../../../internal/generate/scan.go#L50)) builds a `known` set from `installed.<Category>` and skips any file whose name is already there. This silently corrupts state when a name was manually added without `bonsai update` doing the lock+`custom_items` registration.

Change `ScanCustomFiles`:

1. Build a per-category `tracked` set from `lock.Files` keyed by relPath under `agent/<dir>/<name>.<ext>` (i.e. files with a `custom:<type>s/<name>` source already in the lock).
2. Replace the `known[name]` skip with: skip only if **(name in installed list) AND (relPath in lock.Files)**. If only one is true (orphaned registration, or stale lock entry from a deleted file), continue — still treat as discovered.
3. For orphaned registrations (name in list, not in lock), set a new `DiscoveredFile.Orphaned = true` flag. Description copy in `RunStatic`/`DiscoverStage` should mention "re-tracking" rather than "promoting" but the promotion path itself is the same: `applyCustomFileSelection` already idempotently lock-tracks + populates `custom_items`.
4. Add `Orphaned bool` field to `DiscoveredFile` struct.

Test in `internal/generate/scan_test.go`: add `TestScanCustomFiles_OrphanedRegistration` — agent has `skills: [foo]` in installed, file `agent/Skills/foo.md` exists with valid frontmatter, lock has no entry. Expect `len(discovered) == 1`, `discovered[0].Name == "foo"`, `discovered[0].Orphaned == true`, `discovered[0].Error == ""`.

### Step 3 — Code fix 2B: sensor frontmatter parser in `internal/generate/frontmatter.go`

Custom sensor files (`agent/Sensors/foo.sh`) need bash shebang at byte 0 to be executable. The current parser hard-requires `---\n` at byte 0. Extend `ParseFrontmatter` to also accept bash-comment frontmatter:

1. If content starts with `#!`, skip the first line (shebang).
2. After optional shebang, if next non-blank line is `# ---` (or `#---`), enter bash-comment mode: read lines until `# ---` closer, strip leading `# ` (or `#`) from each, parse the resulting body as YAML.
3. Otherwise fall through to existing `---\n` raw-frontmatter path.
4. Both code paths return the same `*config.CustomItemMeta`.

Test additions in `internal/generate/frontmatter_test.go`:
- `TestParseFrontmatter_BashShebang` — `#!/usr/bin/env bash\n# ---\n# name: foo\n# description: bar\n# event: SessionStart\n# ---\n...` → meta with all three fields populated.
- `TestParseFrontmatter_BashNoShebang` — `# ---\n# name: foo\n# description: bar\n# ---\n...` → same.
- `TestParseFrontmatter_RegularMd` — existing `---\nname:...\n---\n` still works (regression).

### Step 4 — UX fix 2C: non-TTY invalid-file signal in `internal/tui/updateflow/run.go`

`RunStatic` filters `d.Error == ""` ([run.go:200](../../../../internal/tui/updateflow/run.go#L200)) and drops invalid files with no signal. CI/non-TTY callers see no indication that frontmatter was malformed.

Change `RunStatic`:

1. Before the filter loop, collect every `DiscoveredFile` with `d.Error != ""` into a slice.
2. After the scan, for each invalid file print a single line to stderr (use `fmt.Fprintf(os.Stderr, ...)` directly — `tui.Warning` writes to stdout):
   ```
   warning: <relPath> — <Error>
   ```
3. Don't fail the run — invalid files are user-fixable, not blockers. The valid set still auto-promotes.

Test in `internal/tui/updateflow/run_test.go` (or a new `run_static_test.go`): create temp project with one valid + one invalid (no-frontmatter) file, capture stderr, assert the warning line appears. If existing test infra makes stderr capture awkward, accept the test gap and verify manually via `make build && /tmp/sandbox`.

## Security

> [!warning]
> Refer to [SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

Specifics:
- New `Orphaned` field is a struct-level boolean — no shell/path injection surface.
- Bash-comment frontmatter parser must reject content where the YAML body (after `# ` stripping) would contain shell metachars. The existing `wsvalidate` rules apply at the `ProjectConfig` boundary, not at meta parse — leave parsing tolerant; the consumer paths (`generate.SettingsJSON`) already render with proper escaping.
- Stderr warning emits the file's RelPath (already a normalized path under workspace) and the Error string (a constant from `scan.go`). Neither is user-controlled in a way that requires escaping.

## Verification

- [ ] `make build` succeeds.
- [ ] `go test ./internal/generate/... ./internal/tui/updateflow/...` passes (existing + new tests).
- [ ] Sandbox repro `/tmp/bonsai-bug-test2/`: starting from `.bonsai.yaml` with hand-added `skills: [manually-registered-skill]` and the file in place, `bonsai update` discovers the orphan, lock-tracks it, populates `custom_items[manually-registered-skill].description`, CLAUDE.md row shows the description.
- [ ] Sandbox repro: `agent/Sensors/foo.sh` with `#!/usr/bin/env bash\n# ---\n# name: foo\n# description: ...\n# event: SessionStart\n# ---` is discovered as a valid sensor, lock-tracked, `.claude/settings.json` gets the hook entry.
- [ ] Sandbox repro: `agent/Skills/no-frontmatter.md` triggers a `warning:` line on stderr in non-TTY `bonsai update </dev/null 2>err && grep warning: err`.
- [ ] No regressions on `bonsai init` / `bonsai add` flows (smoke run in temp dir).
- [ ] Doc Step 1 edits applied; `grep -n "Edit-safe\|edit-safe" catalog/skills/bonsai-model/bonsai-model.md` returns nothing.

## Dispatch

Single worktree. General-purpose code agent. After agent completes, dispatch independent code-review agent with this plan as input. Then merge.

## Out of Scope

- Generalizing `ScanCustomFiles` to walk subdirectories (separate decision — current top-level-only behaviour is intentional).
- Case-insensitive directory matching (`agent/skills/` vs `agent/Skills/`) — separate Backlog item if user hits it.
- Adding `bonsai validate` command to detect orphaned registrations proactively — Backlog candidate.
