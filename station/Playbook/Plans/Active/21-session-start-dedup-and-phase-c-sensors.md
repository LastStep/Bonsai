# Plan 21 — Session-Start Context Dedup + Phase C Sensors

**Tier:** 2 (Feature)
**Status:** Active
**Agent:** general-purpose (worktree isolation)
**Source:** User request 2026-04-21 + Plan 08 Phase C (C1 + C2)

---

## Goal

Eliminate redundant context injection at session start and complete Plan 08 Phase C. After this ships:
- Session-context sensor stops dumping content that duplicates itself, catalog protocols that are procedures, or misfires at the wrong event.
- session-start.md protocol no longer tells the agent to re-read what was already injected.
- station's self-awareness.md stops carrying project-specific UX preferences (they move to memory.md Feedback, their correct home).
- New `compact-recovery` sensor re-injects minimal context after `/compact`.
- `context-guard` detects verification and planning phrases and injects targeted reminders.

---

## Context

**Redundancy diagnosis (2026-04-21 audit):**

`catalog/sensors/session-context/session-context.sh.tmpl` dumps the following on every SessionStart:
- Core files (identity, memory, self-awareness) — **kept**
- INDEX.md, Status.md — **kept**
- FieldNotes.md — **dumped even when empty (just the `---` header)** → cut
- Reports/Pending/* via full `cat` — **overkill; name + first-line summary suffices** → trim
- All Protocols via `{{ range .Protocols }}` — dumps `memory.md` (procedure) and `session-start.md` (circular: tells agent to re-read what was just dumped) → drop these two, keep `security.md` + `scope-boundaries.md`
- "REMINDER: Before ending this session" block → misfires at session **start**; context-guard already handles wrap-up → drop

`catalog/protocols/session-start/session-start.md` currently lists 10 numbered steps = re-read every file the sensor just dumped. Proof of cost: at session start 2026-04-21, protocol caused 6 extra `Read` tool calls re-reading sensor-injected content (~10k wasted tokens).

`station/agent/Core/self-awareness.md` carries a "User Preferences — UX & Collaboration" section (~50 lines) that is project-specific durable feedback — structurally belongs in `memory.md` Feedback. Catalog source is already lean (~24 lines).

**Phase C from Plan 08 remains unstarted:**
- C1: `compact-recovery` sensor — SessionStart with matcher=`compact`, re-injects minimal context (<2000 chars) after context compaction.
- C2: `context-guard` expansion — detect verification and planning phrases, inject targeted reminders.

C1 depends on the session-context sensor being given `matcher: "startup|resume|clear"` so the two don't both fire on `/compact`.

---

## Steps

### 1. Rewrite `catalog/sensors/session-context/session-context.sh.tmpl`

- Drop end-of-session reminder block (current lines 119-124) entirely.
- Replace unconditional `cat ${DOCS}Logs/FieldNotes.md` with a check: skip dump if file contains ≤1 non-blank non-header line after the top `---` separator. Otherwise dump as now.
- Replace `Reports/Pending/*` full-`cat` loop with a summary loop: for each file, print `--- <basename> ---` and the first line matching `^description:` in the YAML frontmatter (or the first non-empty non-`---` line if no frontmatter). Do not `cat` the full body.
- Change Protocols loop so `memory` and `session-start` are filtered out:
  ```
  {{ range .Protocols -}}
  {{ if and (ne . "memory") (ne . "session-start") -}}
  if [[ -f "${WORKSPACE}agent/Protocols/{{ . }}.md" ]]; then
    echo ""
    echo "=== PROTOCOL: {{ . }}.md ==="
    cat "${WORKSPACE}agent/Protocols/{{ . }}.md"
  fi
  {{ end -}}
  {{ end }}
  ```
- Keep: all Core dumps (identity, memory, self-awareness), INDEX.md, Status.md, Health checks (stale memory, Backlog P0 scan, pending reports count, log freshness).

### 2. Add matcher to `catalog/sensors/session-context/meta.yaml`

```yaml
name: session-context
description: Injects core identity, memory, protocols, and project status at session start
agents: all
required: all
event: SessionStart
matcher: "startup|resume|clear"
```

(Generator already supports `matcher` — see `internal/generate/generate.go:491`, `:757`.)

### 3. Rewrite `catalog/protocols/session-start/session-start.md`

Replace the 10-step re-read list with a short version that assumes the session-context sensor handled the heavy lifting:

```markdown
---
tags: [protocol, session]
description: Session startup — act on the context the session-context sensor already injected.
---

# Protocol: Session Start

> [!warning]
> This is a Protocol — follow it every session, no exceptions.

---

## Always

The session-context sensor injects the following at SessionStart: identity.md, memory.md, self-awareness.md, INDEX.md, Status.md, FieldNotes.md (when non-empty), Reports/Pending summary, and always-on Protocols (security, scope-boundaries). Health warnings (stale memory, Backlog P0, pending reports, log freshness) are also surfaced.

Your job:

1. **Address any flags** in the injected `memory.md` Flags section.
2. **Confirm work state** from `memory.md` — resume in-flight tasks or start fresh as appropriate.
3. **Act on health warnings** the sensor raised (stale memory, P0 items, pending reports).
4. **Process pending reports** by reading each file in `Reports/Pending/` (sensor gave summaries only).

> [!note]
> If the session-context sensor is NOT installed (headless agent), fall back to reading core files manually: identity, memory, self-awareness, INDEX, Status, then scan Backlog P0 and check Reports/Pending.

---

## Conditional (by task type)

### If executing a plan

- Read the assigned plan in full before any dispatch
- Read [Playbook/Standards/SecurityStandards.md](../../Playbook/Standards/SecurityStandards.md)
- Read relevant skills from [agent/Skills/](../Skills/)

### If starting new work

- Check for an existing plan in [Playbook/Plans/Active/](../../Playbook/Plans/Active/); ask the user if none
- Re-read scope-boundaries if touching new files

### If reviewing or reporting

- Read the relevant plan or prior report
- Read [Playbook/Standards/SecurityStandards.md](../../Playbook/Standards/SecurityStandards.md)
- Submit reports to `Reports/Pending/` using the report template
```

### 4. Add C2 patterns to `catalog/sensors/context-guard/context-guard.sh.tmpl`

In the Python block, after the wrap-up trigger detection (~line 100 of current file), add:

```python
# Verification triggers — word-boundary only, not end-anchored
verify_patterns = [
    r'\bverify\s+(everything|it\s+all)\b',
    r'\bcheck\s+(your|the)\s+work\b',
    r'\bcheck\s+if\s+you\s+missed\b',
    r'\breview\s+(your|the)\s+changes\b',
    r'\breview\s+before\s+(commit|push|ship)\b',
    r'\bdoes\s+everything\s+look\s+(right|good)\b',
]

verify_hit = any(re.search(p, normalized) for p in verify_patterns)

if verify_hit:
    checklist = (
        '\nVERIFICATION REQUESTED. Before proceeding:\n'
        '1. Re-read your own changes — check for bugs, edge cases, regressions\n'
        '2. Verify all tests pass (if applicable)\n'
        '3. Check for stale references in documentation\n'
        '4. Confirm no security issues introduced'
    )
    injection = (injection + '\n' + checklist) if injection else checklist

# Planning triggers — word-boundary only
plan_patterns = [
    r'\b(lets|let\s+us)\s+plan\b',
    r'\bplan\s+(this|the|a)\b',
    r'\bcreate\s+a\s+plan\b',
    r'\bdesign\s+(this|the|a)\b',
    r'\barchitect\s+(this|the|a)\b',
]

plan_hit = any(re.search(p, normalized) for p in plan_patterns)

if plan_hit:
    reminder = (
        '\nPLANNING DETECTED. Before drafting a plan:\n'
        '1. Load planning workflow: {workspace}agent/Workflows/planning.md\n'
        '2. Load planning-template skill: {workspace}agent/Skills/planning-template.md\n'
        '3. Follow the Tier rules and Verification requirements'
    ).replace('{workspace}', os.path.join(root, ''))
    # Note: only inject planning reminder if no wrap-up trigger already fired
    # (avoid triple-stacking injections)
    if not triggered:
        injection = (injection + '\n' + reminder) if injection else reminder
```

**False-negative guard to verify:** Prompts like "that's all I need to plan for today" must NOT fire `plan_patterns` (word-boundary prevents mid-sentence false positives for `plan`, but verify via test prompt).

### 5. Create `catalog/sensors/compact-recovery/meta.yaml`

```yaml
name: compact-recovery
description: Re-injects minimal context after /compact (Quick Triggers + Work State only)
agents: all
required: all
event: SessionStart
matcher: "compact"
```

### 6. Create `catalog/sensors/compact-recovery/compact-recovery.sh.tmpl`

```bash
#!/usr/bin/env bash
# Compact Recovery — {{ .AgentDisplayName }}
# Re-injects minimal context after Claude Code /compact. Target: <2000 chars.

ROOT="${1:-.}"
WORKSPACE="${ROOT}/{{ .Workspace }}"

echo "=== POST-COMPACT RECOVERY ==="
echo ""

# 1. Extract Quick Triggers table from workspace CLAUDE.md
if [[ -f "${WORKSPACE}CLAUDE.md" ]]; then
  echo "--- Quick Triggers (from CLAUDE.md) ---"
  awk '/^### Quick Triggers/,/^### Protocols/' "${WORKSPACE}CLAUDE.md" | sed '/^### Protocols/d'
  echo ""
fi

# 2. Extract Work State section from memory.md
if [[ -f "${WORKSPACE}agent/Core/memory.md" ]]; then
  echo "--- Work State (from memory.md) ---"
  awk '/^## Work State/,/^## Notes/' "${WORKSPACE}agent/Core/memory.md" | sed '/^## Notes/d'
  echo ""
fi

echo "Context compacted. Resume from Work State above. If uncertain, re-read memory.md and Status.md."
```

**Budget check:** Quick Triggers table ≈ 400-600 chars. Work State section ≈ 500-1500 chars depending on in-flight work. Total stays <2000 in normal conditions; >2000 only when Work State itself is verbose (acceptable — that case matters).

### 7. Station-local customization — move UX prefs to memory.md

- Edit `station/agent/Core/self-awareness.md`: delete the entire `## User Preferences — UX & Collaboration` section (current lines 26-78).
- Append those same items (unchanged bullet content) to `station/agent/Core/memory.md` under a new sub-heading inside the existing `## Feedback` section: `### Durable UX preferences (2026-04-17 dogfooding)`. Each bullet preserves its **Why** and **How to apply** structure.

### 8. Regenerate station workspace

After catalog changes land, run `bonsai update` from the repo root to:
- Sync the rewritten `session-context.sh.tmpl` → `station/agent/Sensors/session-context.sh`
- Sync the rewritten `session-start.md` → `station/agent/Protocols/session-start.md`
- Sync the new `compact-recovery.sh` sensor + regenerate `station/.claude/settings.json` to add its hook entry
- Sync the expanded `context-guard.sh`

Step 7 (self-awareness + memory.md edits) is a **custom file change**; `bonsai update` should leave it alone (lockfile conflict → skip). Verify that `station/agent/Core/self-awareness.md` stays on the post-edit content.

---

## Dependencies

- Generator must emit SessionStart hooks with distinct matchers so `session-context` (matcher=`startup|resume|clear`) and `compact-recovery` (matcher=`compact`) register independently in `.claude/settings.json`. Already supported (`internal/generate/generate.go:473,491,517`).
- No new Go code required. All changes are catalog content + template rewrites.

---

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../Standards/SecurityStandards.md) for all security requirements.

- Sensor scripts run on every SessionStart and UserPromptSubmit. No new shell invocations or subshells that could inject user content.
- `context-guard` regex patterns operate on the prompt string already processed by existing code; additions reuse the same normalization path.
- `compact-recovery` reads files under the workspace only — no network, no writes, no external commands beyond `awk`/`sed`.
- No secrets in any new content. No new dependencies.

---

## Verification

### Build & test (in agent's worktree)

- [ ] `make build` — compiles
- [ ] `go test ./...` — existing tests all pass (no Go code changes, but template-reading tests must still pass)
- [ ] `gofmt -s -l .` — clean

### Fresh-install smoke test

- [ ] `mkdir /tmp/test-21 && cd /tmp/test-21 && /path/to/bonsai init`
- [ ] Inspect `/tmp/test-21/.claude/settings.json` — verify `SessionStart` has two hook groups: one with `"matcher": "startup|resume|clear"` for `session-context`, one with `"matcher": "compact"` for `compact-recovery`
- [ ] Inspect generated `station/agent/Sensors/compact-recovery.sh` exists and is executable
- [ ] Inspect generated `station/agent/Protocols/session-start.md` — short version (no 10-step re-read list)

### Manual session-context smoke test

Run the rewritten sensor directly against the station workspace:

- [ ] `bash station/agent/Sensors/session-context.sh /home/rohan/ZenGarden/Bonsai | wc -c` — measure byte count; should be smaller than pre-change baseline (baseline was ~33.6KB per hook persisted-output)
- [ ] Grep the output — must NOT contain: `PROTOCOL: memory.md`, `PROTOCOL: session-start.md`, "REMINDER: Before ending this session"
- [ ] Grep the output — must contain: `CORE: identity.md`, `CORE: memory.md`, `CORE: self-awareness.md`, `INDEX.md`, `Status.md`, `PROTOCOL: security.md`, `PROTOCOL: scope-boundaries.md`, `SESSION HEALTH CHECK`
- [ ] With `station/Logs/FieldNotes.md` in its current (effectively-empty) state → sensor output should skip the `=== Logs/FieldNotes.md ===` section entirely
- [ ] With `station/Reports/Pending/` empty → no `Reports/Pending/` section; with a test file added → section shows name + first-line only (not full `cat`)

### Manual context-guard smoke test

Pipe JSON to the sensor and check `additionalContext`:

- [ ] Prompt "verify everything" → injection contains "VERIFICATION REQUESTED"
- [ ] Prompt "let's plan the caching layer" → injection contains "PLANNING DETECTED"
- [ ] Prompt "check your work before we commit" → injection contains "VERIFICATION REQUESTED"
- [ ] Prompt "that's all I need to plan for today" → injection contains wrap-up checklist (trailing `that's all`), does NOT contain "PLANNING DETECTED"
- [ ] Prompt "just a normal request" → no injection (unless context tier triggers)

### Manual compact-recovery smoke test

- [ ] `bash station/agent/Sensors/compact-recovery.sh /home/rohan/ZenGarden/Bonsai | wc -c` — output <2000 bytes in normal state
- [ ] Output contains the Quick Triggers table and the Work State section; does NOT dump full memory.md or INDEX.md

### Regeneration idempotency

- [ ] `bonsai update` in repo root reports custom-file conflict on `station/agent/Core/self-awareness.md` (because step 7 modified it) — user picks "skip" and the UX-stripped version is preserved
- [ ] `bonsai update` rewrites `station/agent/Sensors/session-context.sh`, `station/agent/Sensors/context-guard.sh`, and adds `station/agent/Sensors/compact-recovery.sh`

### Documentation

- [ ] `station/CLAUDE.md` Sensors table gets a new row for `compact-recovery` (regenerated automatically by `bonsai update`)
- [ ] `station/Playbook/Status.md` moved to Recently Done with today's date
- [ ] `station/Playbook/Backlog.md` — if Plan 08 has a pending reference to Phase C, archive it
- [ ] `station/agent/Core/memory.md` Work State updated with post-merge hash

---

## Out of Scope

- Plan 08 Phase C **C3 (prompt hook for intent classification)** — stays deferred per original Plan 08 verification note.
- Catalog `memory.md` protocol content — not touched. Kept available; just no longer auto-dumped.
- `self-awareness.md` in catalog — not touched (already lean). Change is station-only.
- `bonsai init` fresh-install TUI changes — none needed.
- Trimming memory.md Notes section — memory-consolidation routine's job, not this plan.
