# Bonsai under pi: the contract plan for the pi side

- Task: `T-5798`. Written 23 Oct 2026 by the orchestrator (Opus 5.5) from the director's choices in the walkthrough of
  `design/agent-neutral.md` (the choices: `records/runs/R-2026-10-22-T-5797-any-agent-research.md`, 23 Oct lines) and
  the hook facts of auditor `M-2026-10-23-2162eec5` (every fact pointed in Bonsai's source at rc.1 and pi 1.4.0's).
- For: the agent that builds the pi side, in pi, in the person's pi package `~/Servers/pi-harness`. It has no other
  context; this file is all it needs. Bonsai's side (Go) is built later in Bonsai's repo, after 1.0.
- Status: draft, for the director's approval after a fresh review.

## 1. In plain words (for the director)

pi has no hook settings, so nothing in pi calls Bonsai today. This plan builds **one small extension**, in plain
JavaScript, in its own folder of the director's pi-harness, loaded last. It does five things inside pi:

1. **Guard.** Before every tool call, it asks `bonsai` whether the call may run, and blocks it unless the answer is
   yes. No answer, a slow answer or a crash also blocks (it fails closed). It judges every call that touches a project
   linked to Bonsai, from wherever the session runs; calls that touch no linked project pass as before.
2. **Tamper check.** It notices when another extension changed a call after the guard allowed it, and from then on
   blocks the session's calls.
3. **Stop gate.** When the model wants to finish a task's work, it asks `bonsai`, and holds the model until the
   ladder is green. Like Claude Code, it lets go after 8 holds in a row with no work between them, so a model that
   stops working is never held forever; one that keeps working is held until the ladder is green.
4. **Opening context.** At the start and after a compaction, it gives the model what Bonsai gives Claude: the active
   task, its grants, its last climb.
5. **Records.** It sends the session's events to Bonsai's log.

It works in **two modes**:
- **Stopgap**, from now: it talks to today's `bonsai` (rc.1, rc.2, 1.0) by dressing pi's calls as Claude Code's, as
  the research's probe did. It works, with honest limits (section 6): the log says `claude-code`, no token usage, and
  Bonsai's roles do not apply.
- **Envelope**, after 1.0: once Bonsai's Go side lands, it sends pi's events as they are, in a new format
  `bonsai.hook/1` (section 5), and Bonsai does the translating. The extension notices which mode the installed
  `bonsai` supports by itself.

What the director approves here: the extension's behaviour (sections 3-4), the envelope format (section 5, the C7
choice: this format only, for now), the limits (section 6) and the proof (section 7).

## 2. Ground rules for the builder

- **Where you work:** only `~/Servers/pi-harness` (the person's own repo, its own git) and scratch folders under a
  temp folder. Never `~/Servers/Bonsai` or its worktrees, never `~/.bonsai`, never `~/.pi/agent/auth.json` or
  sessions. You read Bonsai's docs and `bonsai <cmd> --help`; you never edit Bonsai.
- **The layout:** a new folder `~/Servers/pi-harness/bonsai/`, its own pi package:
  - `bonsai/package.json` (a pi package manifest naming one extension);
  - `bonsai/index.mjs`, the extension: **plain JavaScript, ES module, no TypeScript, no build step, no npm
    dependency** (pi runs compiled TypeScript from a writable cache, `/tmp/jiti/`, and an edited copy there runs;
    a `.mjs` file leaves no cache copy: research P5);
  - `bonsai/lib/*.mjs` for pure functions (the mapping, the path rules, the session-id rule), so tests import them;
  - `bonsai/test/*.test.mjs`, run with `node --test` (Node 22, as pi-harness already needs);
  - `bonsai/README.md`: what it does, its modes, its limits, how to turn it off, every setting.
- **Wiring:** add `"~/Servers/pi-harness/bonsai"` as the **last** entry of `packages` in pi-harness's `settings.json`.
  pi runs `tool_call` handlers in load order, and the guard should judge the call as it will run (research P2). Last
  in `packages` is the latest place a package can take, but it is **not last overall**: after every package pi loads
  the top-level `extensions` entries of the settings, then the files in `~/.pi/agent/extensions/` and in a trusted
  project's `.pi/extensions/`, then its built-ins (pi `dist/core/package-manager.js:704-744,1986-2020`,
  `dist/core/resource-loader.js:409`). Any of those can change a call after the guard allowed it, and the last
  settle handler can undo the stop gate's hold (`dist/core/extensions/runner.js:759-800`). So: keep the settings'
  `extensions` list and `~/.pi/agent/extensions/` empty (pi-harness already loads everything as a package), and the
  README says so in one line, with why. The main pi-harness package must not also load `bonsai/` (check its
  `package.json`'s extension list).
- **Subagents:** the harness's `subagent` children are full pi processes that load the user's packages, so the guard
  loads in them too. The guard **never** turns itself off under `PI_SUBAGENT` (unlike `subagent`, `dashboard` and
  `md-log`).
- **pi-harness's own rules:** add `bonsai` (the binary, `/usr/local/bin/bonsai`) to `deps.json` and a check to
  `scripts/doctor.mjs` (`bonsai --version` runs); update pi-harness's README table and "What is in it". Small commits.
- **Platforms:** Linux and WSL first. On Windows the binary's path and the shell differ; the code keeps the path in one
  constant and says Windows is not yet proved.
- **Do not** change any other harness extension's behaviour, except a one-line note in README.

## 3. The extension, event by event (stopgap mode)

### 3.1 When it is active

`linkedProject(dir)`: walk up from `dir` to the first folder holding `bonsai.yaml` (stop at the filesystem root;
passing a `.git` is fine); return it (absolute) or none.

The **session's project**, found on `session_start` (and again on the reasons `new`, `resume`, `fork` and `reload`,
since the folder can change), in this order:
1. the **inherited** project: the `BONSAI_PI_PROJECT` a guarded **parent process** set (below), when it names a
   folder holding `bonsai.yaml`, so a child started anywhere is judged for the parent's project. Never the value
   this process exported itself: a resume opens a session in its own folder (pi
   `dist/core/agent-session-runtime.js:134-138`), and a switch from project A to project B must be judged against B.
   pi **reloads and re-evaluates the extension** on a session switch and on a change of folder
   (`agent-session-services.js:69`, `resource-loader.js:356-360`, `loader.js:92-103,475-484`), so a module-level
   constant is not enough. Two rules together:
   - every export is a pair: `BONSAI_PI_PROJECT=<folder>` and `BONSAI_PI_PROJECT_PID=<process.pid>`; a value counts as
     inherited only when `BONSAI_PI_PROJECT_PID` is set and names **another** process;
   - the inherited value is kept in `globalThis[Symbol.for("bonsai.pi.inherited")]`, set on the first load only (when
     that key is not yet present, to the inherited folder or to `null`), and read from there on every later load;
2. else `linkedProject(ctx.cwd)`;
3. else none.

When the session has a project, the extension sets `process.env.BONSAI_PI_PROJECT` to it and
`process.env.BONSAI_PI_PROJECT_PID` to `String(process.pid)` (and deletes both when a switched-to session has none),
so every pi child this
process starts (the harness's `subagent` children copy `process.env`; the model chooses a child's folder, which may
lie outside the project: `extensions/subagent/index.ts:483,489,506`; the copy of the environment is at `:385`)
inherits it. It also sets `process.env.BONSAI_PI_PARENT_SESSION` to its own Bonsai session id (3.3) for the envelope
(5.2), by the same two rules: it counts as inherited only when `BONSAI_PI_PROJECT_PID` names another process, and
the inherited value is stashed once in `globalThis[Symbol.for("bonsai.pi.inherited_parent")]`.

**Which calls are judged** (the guard's own rule is that a path is judged even when the session's folder is not
linked, `internal/guard/hook.go:115-116`):
- With a session project: every call, against that project.
- Without one: a file tool (`write`, `edit`, `read`, `grep`, `find`, `ls`) whose resolved path (3.5) lies in a linked
  project (`linkedProject` of the path's folder, or of its nearest existing parent) is judged, with
  `CLAUDE_PROJECT_DIR` set to the **session's own folder** and no `BONSAI_TASK`: Bonsai's guard then takes its own
  branch for a path in a project the session is not in (`internal/guard/hook.go:115-116`), and no task's grants
  count. Everything else passes unjudged, as before: a shell line in an unlinked session is not judged (limit, 6).

"Active" below means the call is being judged.

### 3.2 Calling `bonsai`

One helper, used by every hook:
- Program: the constant `BONSAI = "/usr/local/bin/bonsai"` (a fixed path, never a PATH lookup: an agent can put a fake
  `bonsai` earlier on the PATH).
- `child_process.spawn(BONSAI, ["hook", <name>], {stdio: ["pipe","pipe","pipe"], env})`; write the JSON payload to
  stdin and **close stdin**; read stdout and stderr; a timer of **10 s** kills it.
- Handle an `error` event on the child's stdin (an `EPIPE` when `bonsai` refuses and exits before reading a large
  input): an unhandled stream error would crash pi, and `try/catch` does not catch it. A stdin error is not itself a
  result; the exit status still decides (and a missing status blocks).
- `env`: `process.env`, then set `CLAUDE_PROJECT_DIR` to the folder 3.1 names for the call (the session's project, or
  an unlinked session's own folder) (the guard refuses with `no-project` without it, and
  an inherited one from an outer Claude Code session may name another project), and `BONSAI_TASK` (section 3.7) when
  known; remove `BONSAI_TASK` when not known. Leave `BONSAI_HOME` as inherited.
- Result: `{status, stdout, stderr, timedOut, spawnError}`.

### 3.3 The session id

Bonsai names a log file after the session id and refuses any id outside `^[A-Za-z0-9_-]{1,128}$`; a refused id makes
the guard block every call and the stop gate hold every stop. pi's default ids (UUID v7) pass; `--session-id` may hold
a `.` and has no length cap.
- pi's id: `ctx.sessionManager.getSessionId()`; the session's folder: `ctx.cwd`.
- `bonsaiSessionId(piId)`: the id itself when it matches the rule; otherwise `"pi-"` plus the first 32 hex characters
  of SHA-256 of the id (Node's `crypto`). The same id must always map the same way (a pure function, tested).
- Where pi gives no id (a `--no-session` child may still have one; if `getSessionId()` returns nothing), use
  `"pi-nosession-" + process.pid`.

### 3.4 Guard: `tool_call`

For every `tool_call` event (a call 3.1 does not judge: return nothing, keep no copy):

1. **Map** the call to Claude's shape (table below). Paths are **resolved first**, by pi's own rules (3.5), to an
   absolute path, because the guard otherwise judges `@x` or `file://...` as written, not as pi will open it.
2. **Keep a deep copy** of the event's raw `input` keyed by the tool call id (for 4's tamper check). Never modify
   `event.input`.
3. **Call** `bonsai hook guard` with:

   ```json
   {"hook_event_name": "PreToolUse", "session_id": "<bonsaiSessionId>", "cwd": "<session cwd, absolute>",
    "tool_name": "<mapped name>", "tool_input": {<mapped input>}, "tool_use_id": "<pi tool call id>"}
   ```

   No `agent_id`, `agent_type` or `transcript_path`.
4. **Answer:** exit status `0` and no timeout: allow (return nothing). **Anything else** (status 2, any other status,
   a timeout, a spawn error, a null status): return `{block: true, reason}`, where `reason` is the guard's stderr
   trimmed (it reads `bonsai guard: <why>.` / `next: <next>.` / `rule: <rule>.`), or, when there is none,
   `"Bonsai's guard gave no answer (<what happened>); the call is blocked. Check that /usr/local/bin/bonsai runs."`.
5. Wrap all of it in `try/catch`: an exception blocks, its message in the reason. (pi already blocks when a
   `tool_call` handler throws; do not rely on it.)

The mapping (pi 1.4.0's built-in tools):

| pi tool and arguments | Sent as `tool_name` | `tool_input` |
|---|---|---|
| `write {path, content}` | `Write` | `{file_path: <resolved>, content}` |
| `edit {path, edits: [{oldText, newText}]}` | `MultiEdit` | `{file_path: <resolved>, edits: [{old_string: oldText, new_string: newText}, ...]}` |
| `bash {command, timeout?}` | `Bash` | `{command: <prefix + "\n" + command, or command>}` |
| `powershell {command, timeout?}` | `PowerShell` | `{command}` (pi adds the prefix to `bash` only, `agent-session.js:2974`) |
| `read {path, offset?, limit?}` | `Read` | `{file_path: <resolved>}` |
| `grep {pattern, path?, ...}` | `Grep` | `{pattern, path: <resolved or session cwd>}` |
| `find {pattern, path?, ...}` | `Glob` | `{pattern, path: <resolved or session cwd>}` |
| `ls {path?, ...}` | `pi.ls` | `{path: <resolved or session cwd>}` |
| any other tool (another extension's: `subagent`, `web_fetch`, `ask_user_question`, MCP ...) | `pi.<name>` | the raw input |

Why these names: the guard judges by exact, case-sensitive Claude names; a name it does not know is allowed unjudged
and recorded (`rule: tool-not-judged`, `internal/guard/rules.go:47`). The shell command is judged as pi will run it:
when pi's setting `shellCommandPrefix` is set (read with `pi.getSettings()`, pi `dist/core/extensions/types.d.ts:1286`;
`dist/core/settings-manager.js:788`), pi runs a `bash` line as
`prefix + "\n" + command` (`dist/core/tools/bash.js:174`), so that is what is sent. pi's raw lowercase `write` would therefore pass with no path check, so **every
built-in file and shell tool must be mapped**. The `pi.` prefix keeps another extension's tool from posing as a
Claude tool the guard trusts differently. pi's `edit` arguments are already repaired by pi before `tool_call` (legacy
`oldText`/`newText` at the top, a JSON-string `edits`), so `edits` is an array here; if it is not, block.

A file tool with an empty or missing path, or a shell tool with no `command`, is blocked by the extension itself
(the guard would block it too).

### 3.5 pi's path rules, copied

pi resolves file paths with a function it does not export, so the extension copies it, in this order (pi
`dist/utils/paths.js:50-87`, `dist/core/tools/path-utils.js:42-44`):
1. Unicode spaces (U+00A0, U+2000-U+200A, U+202F, U+205F, U+3000) become a plain space.
2. One leading `@` is stripped.
3. On Windows only: `/c/...`, `/mnt/c/...`, `/cygdrive/c/...` become `C:\...`, except a path starting `//` or
   holding a `\`, which is left as it is (`normalizeWindowsShellPath`, `paths.js:51-58`).
4. `~` and `~/...` (and `~\...` on Windows) expand to the home folder.
5. A `file://` URL becomes a path (`url.fileURLToPath`).
6. The result is resolved against the session's `cwd` (itself normalised by rules 3-5 first), lexically
   (`path.resolve`), without following symlinks (`resolvePath`, `paths.js:83-87`).

Copy `normalizePath` and `resolvePath` line for line, exceptions included; the order above is theirs.

Each rule has a unit test with the exact pi behaviour; one test runs pi's own resolver (imported from its installed
`dist/` by path in the test only) against the copy on a table of inputs, so a pi upgrade that changes the rules
fails the test. `read`'s extra macOS fallbacks are not copied (they apply only to a file that does not exist).

### 3.6 Tamper check: `tool_result`

pi passes one mutable input object through every `tool_call` handler and re-checks nothing after; an extension loaded
after the guard can change a call it allowed (research P2), and several groups load after every package (section 2).

This check is **after the fact**: `tool_result` fires once the changed call has already run (pi
`pi-agent-core/dist/agent-loop.js:597-615`). It cannot undo that call; it stops the session from making more.

On `tool_result` for a call 3.1 judged (it carries the input as executed): compare it, by canonical JSON (keys sorted at every depth), with
the copy kept in 3.4, then drop the copy. On a difference:
- set `tampered = {tool, callId}` for the session; from then on every `tool_call` is blocked with the reason
  `"A pi extension changed a call after Bonsai's guard allowed it (<tool> <callId>); this session's calls are blocked.
  Find the extension, then /reload."`;
- tell the person: `ctx.ui.notify(..., "error")` where a UI exists, and stderr otherwise;
- in envelope mode, also send a `guard_fault` record (section 5); the stopgap has no record for it.

A `/reload` clears it (a new extension load). A result for a call 3.1 would judge but with no kept copy (a call the
guard never saw) is the same fault; a result for a call 3.1 does not judge is ignored.

### 3.7 The active task

The task, in this order: pi's flag `--bonsai-task <id>` (register it with `pi.registerFlag`; a pi started with the
flag and no extension loaded exits 1, so a launcher using the flag proves the extension loaded); else `BONSAI_TASK`
in pi's environment; else none. It is passed to every hook as `BONSAI_TASK`. With no task, the stop gate lets go (as
Bonsai's own does).

### 3.8 Stop gate: `agent_before_settle`

While the session has a project (3.1) and a task is known:
1. When the run is settling with `outcome` `aborted` or `error` (pi's outcomes are `completed`, `aborted`, `error`),
   let it settle (holding a broken run loops); count nothing.
2. Call `bonsai hook stop` with `{"hook_event_name": "Stop", "session_id": "<id>", "cwd": "<session cwd>"}`.
3. Status `0`: let go, and reset the hold count.
4. Anything else (status 2, any other, a timeout, a spawn error, an exception): **hold** by returning
   `{entries: [...event.entries, entry], continue: true}`, where `entry` is exactly
   `{type: "custom_message", customType: "bonsai", content: <text>, display: true}` and `<text>` is `"Bonsai's stop
   gate holds this session: " + <stderr trimmed, or what happened> + " Do what it says, then finish."`. pi accepts
   only the entry types `custom`, `custom_message`, `context_edit` and `compaction` here
   (`SessionBoundaryDraft`, `dist/core/extensions/types.d.ts`; `docs/extensions.md:230`); a custom message becomes a
   user message the model answers. Any other object adds nothing, and pi then drops `continue` and ends the run
   (`dist/core/agent-session.js:578-633`), which would make the gate a silent yes. Keep every other handler's
   entries (`...event.entries`, never a fresh list).
5. **The cap:** at most **8** holds in a row **with no tool call between them** (Claude Code's own cap, which today's
   Bonsai relies on: `cmd/bonsai/hook.go:131-132`): the count resets when the gate lets go and when any tool call
   has run since the last hold, so a model working through a red ladder over many rounds is held for as long as it
   keeps working. On the ninth hold in a row, let go, and tell the person loudly (`notify` error, stderr):
   `"Bonsai's stop gate held 8 times with no work between; the work is NOT proven. The session ended unproven."`. In
   envelope mode the cap is Bonsai's own (section 5) and the extension's stays as a backstop.
6. The whole handler in `try/catch`: pi lets a run settle when a settle handler throws (it fails open), so an error
   must become a hold (step 4), within the cap.

pi never calls the settle handler after Escape; that is the person stopping, and is fine.

### 3.9 Opening context: `before_agent_start` and compaction

Call `bonsai hook start` with `{"hook_event_name": "SessionStart", "session_id": "<id>", "cwd": "<cwd>",
"source": "<source>"}` once per session start reason and after each `session_compact`; keep its stdout (plain text,
at most 60 lines; exit status is always 0; empty when over its 4 s or not linked). `<source>`: pi's reason `startup`
and `new` send `startup`, `resume` and `fork` send `resume`, `reload` sends `startup`; after a compaction, `compact`.
In `before_agent_start`, set `event.systemPromptOptions.sections["bonsai"]` to it: a **system-prompt section**, not
a message (a message is summarised away by compaction; a system-prompt section survives). Nothing printed: set
nothing.

### 3.10 Records

Fire and forget: spawn `bonsai hook record`, write the payload, close stdin, do not wait, never block pi on it, and
swallow every error (detached, `unref()`). Payloads, all with `session_id` and `cwd`:

| pi event | `hook_event_name` | Extra fields |
|---|---|---|
| input from the person (`input` / the prompt) | `UserPromptSubmit` | none (Bonsai keeps no prompt text) |
| `tool_result`, not an error | `PostToolUse` | `tool_name`, `tool_input` mapped as in 3.4, `tool_use_id` |
| `tool_result`, an error | `PostToolUseFailure` | as above, `is_interrupt: false` |
| a settle that let go | `Stop` | none |
| `session_shutdown` | `SessionEnd` | `reason`, `transcript_path: ""` |

The guard already records each `tool_call` decision, so `PreToolUse` is not sent to `record`. Usage (tokens, model)
cannot reach today's Bonsai (it reads only Claude's transcript layout): the `SessionEnd` record says usage unknown.
That is a stopgap limit, closed by the envelope.

## 4. The two modes, and how the extension picks one

At `session_start` (once per process, cached): run `bonsai hook capabilities --json` with the 10 s timer.
- It prints JSON naming `"bonsai.hook/1"` in `formats` and `"pi"` in `hosts`: **envelope mode** (section 5).
- Anything else (today's binary exits with a usage error: there is no such subcommand): **stopgap mode** (section 3).

`hook capabilities` does not exist yet; Bonsai's Go side adds it with the envelope (section 5.4). The README says
which mode is running, and the extension shows it once at start (`notify` info: `Bonsai: guarding this project
(stopgap mode)`).

## 5. The envelope `bonsai.hook/1` (draft contract section)

This section is the contract both sides build to. Bonsai's side implements it after 1.0, and it moves into
`design/contract.md` with that part's approved plan. The extension builds its envelope mode against it now, tested
against a fake `bonsai` (section 7), and leaves it switched off until a real `bonsai` says it supports it (section 4).

### 5.1 The call

`bonsai hook <guard|stop|start|record> --host pi`, one JSON request on stdin (UTF-8, one object, at most 64 MiB),
stdin closed, one JSON answer on stdout. The extension sends pi's events **as they are**: Bonsai's pi reader does the
mapping and the path rules (the director's C2 choice: translate, in Go).

### 5.2 The request

```json
{
  "format": "bonsai.hook/1",
  "host": {"name": "pi", "version": "1.4.0"},
  "event": "tool_call",
  "session": {
    "id": "019a...",
    "cwd": "/abs/path",
    "subagent": "recon",
    "parent_session": "019a...",
    "parent_tool_call_id": null
  },
  "project_dir": "/abs/path/of/the/folder/holding/bonsai.yaml",
  "task": "T-0101",
  "tool": {"name": "edit", "call_id": "toolu_...", "input": {"path": "@src/a.go", "edits": [{"oldText": "a", "newText": "b"}]}},
  "result": null,
  "settle": null,
  "usage": null,
  "source": null,
  "fault": null
}
```

| Field | What it holds | Example |
|---|---|---|
| `format` | always `"bonsai.hook/1"` | `"bonsai.hook/1"` |
| `host` | the agent and its version | `{"name": "pi", "version": "1.4.0"}` |
| `event` | one of `tool_call`, `tool_result`, `settle`, `session_start`, `compact`, `prompt`, `message_end`, `session_shutdown`, `guard_fault` | `"settle"` |
| `session.id` | pi's own session id, raw (Bonsai maps it to its log rule, the same way 3.3 does) | `"019a3c..."` |
| `session.cwd` | the session's folder, absolute | `"/work/proj"` |
| `session.subagent` | `PI_SUBAGENT`'s value in a harness child, else null | `"recon"` |
| `session.parent_session` | `BONSAI_PI_PARENT_SESSION` inherited from a guarded parent (3.1), else null. The call that started the child is not known (the harness passes none), so roles on pi need more (a later part, research D5) | `"019a..."` |
| `session.parent_tool_call_id` | set for a call made from inside another call (codemode, MCP), else null | `null` |
| `project_dir` | the folder holding `bonsai.yaml` the extension found (3.1); Bonsai checks it itself too | `"/work/proj"` |
| `task` | the active task (3.7), or null | `"T-0101"` |
| `tool` | `tool_call` and `tool_result`: pi's tool name, call id and raw input, unmodified | see above |
| `result` | `tool_result` only: `{"is_error": bool, "input_executed": {...}}` | `{"is_error": false, "input_executed": {...}}` |
| `settle` | `settle` only: `{"outcome": "completed" \| "aborted" \| "error", "holds_in_a_row": n, "tool_calls_since_hold": n}` | `{"outcome": "completed", "holds_in_a_row": 0, "tool_calls_since_hold": 3}` |
| `usage` | `message_end` only: `{"model", "provider", "input", "output", "cache_read", "cache_write", "cost"}` from pi's message's `usage`; `cost` is pi's own object, copied with exactly its fields (`input`, `output`, `cacheRead`, `cacheWrite`, `total`, numbers); the builder confirms these names in pi's `session-format.md` and the Go side's plan fixes them | `{"model": "claude-haiku-5-5", "input": 1200, ..., "cost": {"total": 0.002, ...}}` |
| `source` | `session_start` only: pi's reason (`startup`, `reload`, `new`, `resume`, `fork`) | `"startup"` |
| `fault` | `guard_fault` only: `{"kind": "input-changed-after-guard", "tool": "...", "call_id": "..."}` | see 3.6 |

Absent and null are the same. Unknown fields are refused (Bonsai's formats are closed), so the extension sends
exactly these.

### 5.3 The answer

```json
{"format": "bonsai.hook/1", "decision": "block", "why": "bonsai.yaml is person-only", "next": "ask the person", "rule": "person-only", "context": null}
```

| Field | Values | Used by |
|---|---|---|
| `decision` | `allow`, `block` (guard); `hold`, `let_go` (stop); `none` (start, record) | every hook |
| `why`, `next`, `rule` | Bonsai's refusal in three parts, as its stderr gives them today; null on allow and let_go | guard, stop |
| `context` | start only: the opening context text (at most 60 lines), or null | start |

Exit status `0` with an answer that parses: obey it. **Anything else** (another status, no answer, an answer that does
not parse, a `decision` not allowed for that hook, a timeout): guard blocks, stop holds (within the extension's cap),
start adds nothing, record ignores it. Bonsai keeps its own stop cap in envelope mode (in its own records, by session).

### 5.4 `bonsai hook capabilities --json`

Prints `{"formats": ["bonsai.hook/1"], "hosts": ["pi"], "version": "<bonsai version>"}`, exit 0. Bonsai adds it with the envelope; it lets the extension pick its mode (4).
Its exact shape is fixed with the Go side's plan; the extension only needs those two lists.

### 5.5 What Bonsai's Go side owes this contract (after 1.0, for that plan)

The pi reader maps exactly the table in 3.4 (the Go copy of it, in the pi driver, never in Claude's tool lists) and
applies the rules in 3.5; maps session ids as 3.3; records `agent: "pi"` and the usage from `message_end`; treats a
`guard_fault` as a guard fault record; keeps its own stop cap; adds `capabilities`. The extension does not change when
that lands.

## 6. Limits, said plainly

**In stopgap mode:**
- The log names every pi record `agent: "claude-code"` (rc.1 writes that literal).
- No token usage or model reaches Bonsai.
- **Roles do not apply.** With no subagent id, Bonsai treats every pi session, harness children included, as the
  person's own session; in a project that declares roles (Bonsai's own repo does), that is more rights than a
  child should have. Use the stopgap in projects without roles, or accept it. Not in Bonsai's own repo.
- Tools other than pi's built-ins (`subagent`, `web_fetch`, MCP ...) are recorded, not judged.
- A pi session started outside any linked project, and not by a guarded pi parent, has its shell lines unjudged: a
  `bash` line there can still change a linked project's files. Its file tools are judged by their target (3.1).
  Children of a guarded session inherit the project and are judged in full, wherever they run.
- The guard is the last **package**, not the last extension (section 2); the tamper check catches a later change
  only after the changed call has run (3.6).
- Person-only commands: today only Claude's marker makes Bonsai's own commands refuse an agent; pi's shell carries
  `AI_AGENT=pi` instead. The guard's shell check still refuses `--allow-exec`, `unlink`, `settings set` and `labels`
  on a judged `bash` line, and rc.2 makes every agent marker refuse (`T-5710` item 13).

**In both modes:**
- **Nothing here is a wall.** pi has no sandbox and one OS user; an agent that means to get past it can (`pi -ne`,
  editing pi's settings, `PI_CODING_AGENT_DIR`, editing this extension, an `env` trick). It stops mistakes and agents
  acting in good faith; the proof (the ladder, CI on the exact commit) is what Bonsai trusts. Research section 5.
- The extension's file lives in the person's home folder, writable by any agent; until Bonsai's `check` verifies its
  hash and its place (after 1.0), git in pi-harness is the record (`git status` shows a change).
- pi's compile cache runs edited TypeScript copies of the **other** harness extensions (research P5);
  `JITI_FS_CACHE=false` in the environment pi starts from closes that, at some start-up cost.
- A pi started from inside Claude Code inherits Claude's variables, a messaging token among them (research P3).

## 7. The proof

The ladder has no rung for JavaScript, so the pi side proves itself, in pi-harness:

**Unit tests (`node --test bonsai/test`), with a fake `bonsai`** (a small script whose path the tests inject in place
of the constant, answering by a table): the mapping table, every row; the path rules, every rule plus the check
against pi's own resolver; the session-id rule (a `.`, 129 characters, a UUID); allow on 0, block on 2, on another
status, on a timeout (the fake sleeps), on a spawn error (a missing path), on an exception; the tamper check (a
changed input, a missing copy, a result for an unjudged call ignored); a stdin `EPIPE` does not crash; the stop gate:
the hold's entry is exactly the `custom_message` of 3.8, let go, an error held, the cap of 8 in a row, the count reset
by a tool call between holds, the outcome skip, other handlers' entries kept; which calls are judged (3.1: the session
project, the inherited `BONSAI_PI_PROJECT`, a file tool's target in a linked project from an unlinked session, an
unrelated call passing; a session in project A switched to project B judged against B; the extension module evaluated twice in one process (as pi's reload does) still reads the parent's original value from the stash, and a value with this process's own pid never counts as inherited; a switch to an unlinked folder deletes both variables; a grandchild inherits the grandparent's project; the parent session id follows the same rules); the `shellCommandPrefix` joined as pi joins it; the mode pick (a fake that knows `capabilities` and one
that does not); the envelope mode's requests and answers against section 5's examples, and a malformed answer blocks.

**A live run, in a scratch project** (the person's pi login, the cheapest model, a few cents; the steps of the
research's trial, 2.4-2.6):
1. A temp git repo linked with the installed `bonsai` under a scratch `BONSAI_HOME` (see `bonsai init --help`), so the
   person's `~/.bonsai` is untouched.
2. `pi -ne -nc -ns --no-session -e ~/Servers/pi-harness/bonsai/index.mjs -p "<instruction>" < /dev/null` (close
   stdin: `pi -p` waits on it), once per check:
   - writes and edits to `bonsai.yaml` refused, through each path form: relative, absolute, `@bonsai.yaml`,
     `~/...`, `file://...`; a write into `.bonsai/local/` refused; an ordinary file allowed;
   - `bash rm bonsai.yaml` refused;
   - with `--bonsai-task` naming a task whose ladder is red, the stop is held, and lets go once it is green (or at
     the cap, with the loud line);
   - a second scratch extension loaded after the guard that rewrites a write's path: caught, and the next call
     blocked;
   - `/usr/local/bin/bonsai` swapped for a missing path (by the test's path override): every call blocked;
   - a session started in a folder outside the project: a write to the project's `bonsai.yaml` by absolute path
     refused; an ordinary write in its own folder allowed;
   - the stop gate's hold reaches the model as a custom message and the run continues (not a silent end);
   - in an interactive pi in the project, `/resume` into a session whose folder is another temp linked project:
     a protected write there is judged against that project (its own refusal), not the first one;
   - the records landed in the scratch project's log (`bonsai log` or the log file), naming the session id.
3. A harness `subagent` child in the scratch project, and one the parent starts with its `cwd` in another temp
   folder: each one's write to the project's `bonsai.yaml` (absolute path) and `bash rm` of it are refused.

The builder reports the counts, each live check's outcome, and the pi and `bonsai` versions. The director's check
afterwards (two minutes): in a linked project, ask pi to edit `bonsai.yaml`; it is refused, quoting Bonsai.

## 8. Order of work

1. `bonsai/lib`: the path rules, the session-id rule, the mapping; their tests.
2. The call helper, the guard, the tamper check; tests with the fake.
3. The stop gate with its cap; start; records; tests.
4. The mode pick and the envelope mode against the fake; tests.
5. The wiring (`packages`, last), README, `deps.json`, doctor.
6. The live run (section 7), then the report.

## 9. After 1.0, on Bonsai's side (not for the pi builder)

The Go part of `design/agent-neutral-walkthrough.md`'s lanes A, B and D, reshaped by these choices: identity (A1, A2),
the envelope and pi's reader (B1-B4, section 5 here as their contract, plus `hook capabilities`), the protection and
honesty pieces (D1, D3), and then the proof of section 7 run again in envelope mode. Its plan is written then, reviewed
by a fresh Opus agent and approved by the director; this file's section 5 moves into `design/contract.md` with it.
