---
description: The headless CLI contract — flags, serializations, and exit codes for driving Bonsai non-interactively from an AI agent, CI script, or MCP wrapper.
---

# Agent Interface — the headless CLI contract

This is the canonical, generic contract for driving `bonsai` without a TTY:
which flags switch each command into a non-interactive mode, the two output
serializations (streaming JSONL for mutating commands, single-document JSON for
read commands), and the per-command exit-code table. It is the single source of
truth for AI integrators, CI pipelines, and the planned `bonsai mcp` server
(which wraps these same cores).

The headless cores live in `internal/nonint` (mutating commands) and
`internal/generate` (read serializers). Each mutating core is a pure function:
typed options in, a structured `Result` out — no prompts, no `os.Exit`, no data
on stdout from inside the core. The CLI adapter serializes the `Result` to
JSONL on stdout and routes warnings to stderr; a future MCP adapter serializes
the same `Result` to structured content.

> **Exit-code source of truth.** The mutating exit-code constants
> (`ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4`,
> `ExitConflict=5`) are defined in `internal/nonint/runner.go`. That file is
> canonical — if the table below ever disagrees with those constants, the
> constants win.

---

## Stream discipline

A single rule governs both serializations:

- **Data goes to stdout.** Mutating commands write pure JSON Lines; read
  commands write a single indent-2 JSON document. Nothing else is written to
  stdout on the headless path.
- **Diagnostics and warnings go to stderr.** Operational errors print to stderr
  (the process then exits non-zero with no stdout). Non-fatal warnings (e.g. a
  lock-save failure, an invalid custom-file discovery) print to stderr as plain
  `warning: …` lines.
- **`warning` events never appear on stdout.** Warnings ride in `Result.Warnings`
  and are emitted by the CLI adapter to stderr only. The JSONL stream carries
  exactly two event kinds: `file` and `summary`. This keeps stdout pure protocol
  — safe to pipe straight into `jq` or an MCP stdio transport.

Capture the two streams separately:

```bash
bonsai init --non-interactive --from-config cfg.yaml >events.jsonl 2>diag.log
```

---

## Per-command flags

| Command | Headless trigger | Other headless flags |
|---------|------------------|----------------------|
| `init` | `--non-interactive` + `--from-config <path>` (both required together) | — |
| `add` | `--non-interactive` + `--from-config <path>` (both required together) | — |
| `update` | `--non-interactive` (force headless on a TTY) **or** a non-TTY stdin | `--skip-conflicts` |
| `remove` | `--non-interactive` / `-y`,`--yes` (bypass the confirm prompt) **or** a non-TTY stdin | `--from <agent>`, `-d`,`--delete-files` |
| `list` | `--json` | — |
| `catalog` | `--json` | `-a`,`--agent <type>` |
| `validate` | `--json` | `-a`,`--agent <type>` |

Notes:

- `update` and `remove` auto-detect a non-TTY stdin and switch to headless
  automatically; the explicit flag is only needed to force headless on a real
  TTY. `init`/`add` require the explicit flag pair.
- Mutating commands have **no `--json` flag** — headless mode always emits
  JSONL, so a `--json` toggle would be redundant with `--non-interactive`.
- `remove --yes` bypasses the human **confirmation** gate, never **validation**:
  an empty target or a literal `*` is rejected (exit 2, zero mutation). Under
  `--yes --delete-files`, each delete target is checked and the operation
  refuses if any is a symlink (exit 2, zero deletion).
- `remove --from <agent>` scopes an item removal to one owning agent; it is
  required to disambiguate an item installed in more than one agent. `--from` is
  not an escape hatch around required-item protection — removing a required item
  via `--from` is still rejected (exit 2, zero mutation).

---

## Input shapes

- **`init` / `add`** read a declarative overlay file via `--from-config <path>`:
  a YAML document with the same shape as `.bonsai.yaml`. `init` accepts a single
  `tech-lead` entry under `agents:`; `add` accepts exactly one agent in the
  overlay (loop the command for multi-agent setups), and its non-`agents` fields
  (`project_name`, `docs_path`, `scaffolding`) must either be omitted or match
  the existing `.bonsai.yaml` exactly.
- **`update` / `remove`** take **no config file**. `update` reconciles from the
  existing `.bonsai.yaml` plus a workspace scan. `remove` is imperative and
  takes positional arguments: `bonsai remove <agent>` or
  `bonsai remove <type> <name>` (`type` ∈ skill | workflow | protocol | sensor |
  routine).

There is no new untrusted-input surface: `update`/`remove` ingest no
caller-supplied file, and every JSON/JSONL byte is marshaled
(`json.Marshal` / `json.MarshalIndent`), never string-concatenated.

---

## Serialization 1 — JSONL (mutating commands)

`init`, `add`, `update`, and `remove` stream **JSON Lines** to stdout: one JSON
object per line, no enclosing array. There are exactly two event kinds.

### `file` event

One per file outcome, emitted in write order. Empty optional fields are omitted.

```json
{"event":"file","path":"station/CLAUDE.md","action":"created","source":"tech-lead"}
```

| field | type | notes |
|-------|------|-------|
| `event` | string | always `"file"` |
| `path` | string | repo-relative path of the affected file (omitted if empty) |
| `action` | string | one of `created`, `updated`, `unchanged`, `skipped`, `conflict` (omitted if empty) |
| `source` | string | origin label (e.g. the owning agent type); omitted if empty |

### `summary` event

Exactly one terminal line. **All five count keys are always present**, even when
zero — downstream consumers can parse the shape unconditionally.

```json
{"event":"summary","created":12,"updated":0,"unchanged":3,"skipped":0,"conflicts":0}
```

| field | type | notes |
|-------|------|-------|
| `event` | string | always `"summary"` |
| `created` | int | files newly written |
| `updated` | int | files re-rendered over a Bonsai-managed prior version |
| `unchanged` | int | files already up to date |
| `skipped` | int | files left untouched (e.g. conflicts skipped under `--skip-conflicts`) |
| `conflicts` | int | files that conflict with user edits and were not overwritten |

The all-installed / no-op case (e.g. `add` against an already-fully-installed
agent) emits a lone zero-count summary line and no `file` events.

---

## Serialization 2 — single-document JSON (read commands)

`list`, `catalog`, and `validate` with `--json` emit a **single** indent-2 JSON
document to stdout (not JSONL). All three share the same serialization
conventions: `json.MarshalIndent(v, "", "  ")`, stdout-only.

- **`list --json`** — a `ListSnapshot`: the installed agents and the abilities
  each carries, read from `.bonsai.yaml`. Ability lists are always present
  (an agent with no skills serializes `"skills": []`, never `null`).

  ```json
  {
    "version": "v0.5.0",
    "docs_path": "station/",
    "agents": [
      {
        "type": "tech-lead",
        "workspace": "station/",
        "skills": ["planning-template"],
        "workflows": ["planning"],
        "protocols": ["memory"],
        "sensors": ["status-bar"],
        "routines": ["status-hygiene"]
      }
    ]
  }
  ```

- **`catalog --json`** — the embedded catalog (available agents, skills,
  workflows, protocols, sensors, routines). Honours `-a <agent>` filtering.
  Scaffolding is intentionally excluded (it is project-config data, not catalog
  data).

- **`validate --json`** — a single audit `Report` object (orphaned
  registrations, stale lock entries, untracked customs, frontmatter problems).

---

## Exit codes

### Mutating commands

Reachability is per command — a code not listed for a command is unreachable on
its headless path. The constants below are defined in
`internal/nonint/runner.go`.

| Code | Constant | Meaning | Reachable by |
|------|----------|---------|--------------|
| `0` | `ExitOK` | success | init, add, update, remove |
| `2` | `ExitInvalidConfig` | caller input rejected before any mutation — bad overlay shape, unknown agent type, missing tech-lead, multi-owner item with no `--from`, last-tech-lead removal, empty/`*` target, symlinked delete target | init, add, update, remove |
| `3` | `ExitRuntime` | a generator or filesystem error occurred mid-run | init, add, update, remove |
| `4` | `ExitWrongCWDForInit` | wrong working-directory state — `.bonsai.yaml` already present (init) or missing (add / update / remove) | init, add, update, remove |
| `5` | `ExitConflict` | unresolved file conflicts; re-run with `--skip-conflicts` or interactively | **update only** |

Per-command reachable set:

| Command | 0 | 2 | 3 | 4 | 5 |
|---------|---|---|---|---|---|
| `init` | yes | yes | yes | yes (`.bonsai.yaml` exists) | — |
| `add` | yes | yes | yes | yes (no `.bonsai.yaml`) | — |
| `update` | yes | yes | yes | yes (no `.bonsai.yaml`) | yes (conflict, no `--skip-conflicts`) |
| `remove` | yes | yes (not-found / multi-owner / last-tech-lead) | yes | yes (no `.bonsai.yaml`) | — |

`init` and `add` never reach `5`: under `--non-interactive` they force conflicts
to skip (reported as `action:"conflict"` file events with the files left
untouched) and exit `0`.

### Read commands

Read commands do **not** use the mutating constants and do **not** all exit `0`:

| Command | Exit codes |
|---------|------------|
| `catalog --json` | `0` |
| `list --json` | `0` |
| `validate --json` | `0` clean · `1` issues found · `2` config-load / internal error |

---

## Quick recipes

```bash
# init from a YAML overlay, parse the stream
bonsai init --non-interactive --from-config cfg.yaml | jq -c 'select(.event=="file")'

# update headlessly, treat conflicts as a hard stop (exit 5)
bonsai update --non-interactive; echo "exit=$?"

# update, skipping conflicts (exit 0, conflicts counted as skipped)
bonsai update --non-interactive --skip-conflicts

# remove an item owned by two agents — disambiguate with --from
bonsai remove skill planning-template --from backend --yes

# read the installed state as one JSON document
bonsai list --json | jq '.agents[].type'

# audit; exit 1 means issues were found
bonsai validate --json; echo "exit=$?"
```
