---
tags: [log, session, statusline, sensors]
date: 2026-04-22
scope: project (personal → project scope migration)
---

# 2026-04-22 — statusLine redesign

## Context

User's ask: existing Stop-hook `systemMessage` (tokens/turns/tools metrics) wasn't useful. Wanted a persistent bottom-of-screen bar showing:

- context usage this session
- model name
- git branch
- elapsed session time
- 5h rolling budget %
- cost

Dual-surface request: "both kind of data. one that lives per session, and one for each message, as they are different data points." Mapped to: persistent `statusLine` (glance bar, every message) + `status-bar.sh` Stop hook (warnings in transcript, end of turn).

Scope evolution: started personal (`~/.claude/bonsai-statusline.sh`) to validate fast, then promoted to project (`station/agent/Sensors/statusline.sh`) once user confirmed. Phase 2 catalog port deferred (Backlog P2 Group E).

## Architecture — two surfaces

| Surface | When | Channel | Purpose |
|---|---|---|---|
| `statusline.sh` | Every prompt render (bottom bar) | stdout of `statusLine.command` | Passive glance — state at-a-glance |
| `status-bar.sh` | End of turn (Stop hook) | `{"systemMessage": "..."}` | Warnings in transcript; silent on zero |

Context-guard decoupled from status-bar: previously shared a `/tmp/bonsai-awareness-*.json` state file. Trimmed status-bar, so context-guard self-computes ctx% from `transcript_path` in stdin.

## Files changed

| File | Kind | Notes |
|---|---|---|
| `station/agent/Sensors/statusline.sh` | new | Main renderer |
| `station/agent/Sensors/status-bar.sh` | rewrite | Warnings-only, silent on zero |
| `station/agent/Sensors/context-guard.sh` | rewrite | Self-computes ctx% from transcript, no state file dep |
| `.claude/settings.json` | edit | Added top-level `statusLine` stanza |

## statusLine format

```
cave:full · tech-lead · Opus 4.7 · main*20 · ctx 42% · 5h 31% · 30m · $0.47
```

- Separator: `  ·  ` in ANSI 256-color `240` (muted grey)
- Tier color: sage `108` (<50%), sand `180` (50–70%), rose `174` (≥70%) — applied to ctx/5h/7d %
- Workspace tag: walks up 12 parents from `workspace.current_dir` looking for `.bonsai.yaml`, regex-extracts first agent key (`^\s{4}([a-z0-9-]+):`). Falls back to dir basename.
- Dirty count: `main*N` where N = `git status --porcelain | wc -l`
- Elapsed: from `cost.total_duration_ms`, rendered as `Nm` or `Nh Mm`
- Env toggles: `NO_COLOR=1` strips color, `BONSAI_STATUSLINE_HIDE=cost,5h,7d` omits fields (comma list)

## Caveman wrap

Caveman plugin ships its own `caveman-statusline.sh`. We replace rather than defer, so we wrap natively:

```bash
CAVEMAN_FILE="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/.caveman-active"
if [[ -f "$CAVEMAN_FILE" && ! -L "$CAVEMAN_FILE" ]]; then
  MODE=$(head -c 64 "$CAVEMAN_FILE" | tr -cd 'a-z0-9-')
  case "$MODE" in lite|full|ultra) CAVEMAN="cave:$MODE" ;; esac
fi
```

Security posture matches their plugin: symlink refuse, byte cap (64), whitelist (`[a-z0-9-]` + case-match against allowed modes).

## Visual iteration

User walked through three styles:

1. `🦴 full` (emoji) — didn't like emoji
2. `` full` (Nerd Font codepoints) — user has no Nerd Font installed, icons invisible
3. `cave:full` (plain text label) — picked (option A)

Also dropped emojis everywhere else ("dont use emojis anywhere actually"). All fields are plain text labels.

## Bugs hit

### 1. Heredoc stole stdin

First script used `printf '%s' "$INPUT" | python3 - <<'PY'`. The heredoc won — `json.load(sys.stdin)` read Python source text, not JSON. Render was blank.

**Fix:** pass JSON via env var instead of stdin:

```bash
STATUSLINE_JSON="$INPUT" python3 <<'PY'
import json, os
data = json.loads(os.environ.get('STATUSLINE_JSON', '') or '{}')
```

Applied same pattern to context-guard.sh rewrite.

### 2. `&&` stored as `&&` in settings.json

Project `.claude/settings.json` had `&&` serialized as unicode escapes. Edit tool's exact-match failed 4× — bare `&&`, larger context, etc.

**Fix:** round-trip via Python `json.load → mutate → json.dump`. Also normalized the file's formatting in the process.

### 3. Latent bug — not fixed, filed to backlog

`context-guard.sh` planning-reminder path uses `os.path.join(root, "")agent/...` which resolves to `{root}/agent/...` — missing the `station/` prefix. Filed as P2 Group B. Didn't fix inline (out of scope of statusLine work).

## Validation

Smoke-tested via walk-up wrapper with sample JSON payload. Rendered:

```
cave:full · tech-lead · Opus 4.7 · main*20 · ctx 42% · 5h 31% · 30m · $0.47
```

Full render will be visible on next `/clear` or fresh session.

## Follow-ups filed

- **Backlog P2 Group B:** fix context-guard planning-reminder path prefix.
- **Backlog P2 Group E:** port statusLine to catalog — **now tracked as GitHub issue [#53](https://github.com/LastStep/Bonsai/issues/53)** with full background, findings, acceptance criteria, testing plan, and proposed implementation in comments. Deferred — pickup via `/issue-to-implementation`.

## Late-session addendum (wrap phase)

After user asked to move prototype into project scope as a first step toward catalog port, investigation surfaced two more gotchas:

### Subdirectory launch gotcha

User launched Claude from `station/`, so the live `.claude/settings.json` was `station/.claude/settings.json`, not repo-root `.claude/settings.json`. First attempt at wiring the statusLine stanza went into the wrong file. Reverted, added to the live one. Memory note added.

### statusLine `$PWD` differs from hooks `$PWD`

The walk-up-to-`.bonsai.yaml` wrapper used for every hook (and copied into the statusLine command on first attempt) fails silently for statusLine. Claude Code invokes `statusLine.command` with a different working directory than hook commands. Walk terminates at `/` with no match → `bash /station/agent/Sensors/statusline.sh "/"` which doesn't exist → silent no-op.

**Fix:** absolute path at install time. `"command": "bash /abs/path/to/statusline.sh"`. This is the key finding that forced the catalog-port design away from a wrapper script and toward computing absolute paths at generate time.

### Hot-reload: no

statusLine config changes are **not** hot-reloaded. Require `/clear` or Claude Code restart to pick up edits. Not documented anywhere we found; confirmed empirically.

### Disposition

Feature request filed as issue #53 with all the above findings. Not tackling this session. `.claude/settings.json` (repo root parent) cleaned up — the leftover stanza removed. Prototype sensor files (`statusline.sh`, rewritten `status-bar.sh` + `context-guard.sh`) still uncommitted — user's call whether to commit now or roll into #53 implementation later.

## Process notes

- Direct-to-main commits per fast-iter UX convention (memory.md Feedback section) — no PRs for sensor polish.
- Personal-scope prototype (`~/.claude/bonsai-statusline.sh`) promoted to project scope once user validated feel.
- Did NOT touch user's in-flight 12-point initflow polish run during wrap-up — memory.md update narrowed to "Loose ends" line only.
