# Companion App — Design Document

**Codename:** Greenhouse
**What:** Desktop app for managing projects and observing AI coding agents at work
**Relationship to Bonsai:** Separate open-source product. Bonsai defines agents; this app runs and observes them.
**Status:** Design phase

---

## Vision

A lightweight desktop app that gives developers a unified control plane across all their projects and AI coding agents. You see what every agent is doing, launch new sessions, track costs, review activity — without leaving one window.

Bonsai scaffolds the agent workspace. This app is where you *operate* it.

---

## Positioning in the Landscape

### Existing Tools

| Tool | Type | Stars | What it does | Gap |
|------|------|-------|-------------|-----|
| **claude-devtools** | Electron + React | 3.1k | Reads `~/.claude/` logs, shows token attribution, subagent trees, compaction viz | Read-only observer. Electron (heavy). No process management. |
| **Claude Squad** | Go TUI (tmux) | 7k | Manages multiple Claude Code instances in tmux + git worktrees | Terminal-only. No GUI. No dashboard. |
| **Mission Control** | Next.js web app | 4k | Kanban tasks, cost tracking, agent orchestration, scheduling | Web app (needs server). Not a desktop app. |
| **CC-Lens** | Web dashboard | — | Real-time monitoring reading `~/.claude/` directly | Browser-based. No process control. |
| **ToolKode** | Proprietary TUI | 2 | Terminal agent with Switchboard multi-workspace UI | Closed-source. Replaces Claude Code instead of complementing it. 2 weeks old. |

### Where this app fits

None of these are: **a fast, native desktop app that both spawns and observes Claude Code sessions across multiple projects**.

- claude-devtools is closest but read-only + Electron (150MB+ binary, 200MB+ RAM)
- Claude Squad manages processes but has no GUI
- Mission Control has the right feature scope but is a web app

This app: **Tauri (8MB binary, ~100MB RAM) + both spawn and observe modes + designed for Bonsai but agent-agnostic**.

---

## Tech Stack

### Core

| Layer | Choice | Why |
|-------|--------|-----|
| **App framework** | Tauri v2 | 8-10MB binary, native feel, Rust backend for process management, cross-platform (Win/Mac/Linux) |
| **Backend** | Rust (Tauri core) | Process spawning via tokio, local HTTP server via axum, SQLite via sqlx |
| **Frontend** | Svelte 5 (SvelteKit SPA mode) | Smallest bundle, compile-time reactivity, fastest runtime. Best Tauri pairing per community consensus. |
| **Database** | SQLite (via tauri-plugin-sql / sqlx) | Projects, sessions, activity logs. Local-first, zero setup. |
| **Settings** | tauri-plugin-store | Key-value JSON for app preferences |

### Key Libraries / Crates

| Need | Rust Crate | Why |
|------|-----------|-----|
| Process management | `tokio::process` + `command-group` | Async process spawning + clean group termination (kills all children) |
| Webhook server | `axum` | Tokio-native, lightweight, runs alongside Tauri on `tauri::async_runtime::spawn` |
| File watching | `notify` + `notify-debouncer-mini` | Watch `~/.claude/` for session changes. Cross-platform standard. |
| Reactive DB | `sqlx` + `real-time-sqlx` (optional) | Compile-time checked queries + auto-push changes to frontend via Tauri channels |

| Need | Frontend Library | Why |
|------|-----------------|-----|
| Activity feed | Custom Svelte 5 component + `svelte-virtual` | Native runes reactivity handles high-frequency streaming. Virtualized scrolling for large logs. |
| Terminal (future) | `xterm.js` + `@xterm/addon-webgl` | Only if we add embedded terminal later. Not MVP. |
| UI framework | `shadcn-svelte` or `skeleton` | Component primitives for the dashboard |

### Tauri Plugins (all official first-party)

| Plugin | Purpose |
|--------|---------|
| `tauri-plugin-shell` | Spawn Claude Code processes, stream stdout, write stdin, kill |
| `tauri-plugin-sql` | SQLite for structured data |
| `tauri-plugin-store` | App settings/preferences |
| `tauri-plugin-notification` | Desktop notifications (agent finished, needs input, error) |
| `tauri-plugin-single-instance` | Prevent multiple app instances |
| `tauri-plugin-autostart` | Optional launch-at-login |
| `tauri-plugin-window-state` | Remember window size/position |
| `tauri-plugin-updater` | Auto-update from GitHub Releases |
| `tauri-plugin-process` | Restart/relaunch after update |
| `tauri-plugin-log` | Structured logging |
| `tauri-plugin-deep-link` | Custom URL scheme (future: `companion://open?project=...`) |

---

## Integration with Claude Code

### The Big Discovery: Agent SDK

Claude Code has official Agent SDKs (Python + TypeScript) that spawn Claude Code as a subprocess with structured JSON communication over stdin/stdout. The protocol is:

```
Companion App (Rust)
    │
    ├── spawns: claude --print --output-format stream-json --input-format stream-json
    │           --include-hook-events --permission-mode <mode>
    │
    ├── writes to stdin:  {"type": "user", "message": {"role": "user", "content": "..."}}
    │
    └── reads from stdout (NDJSON, one JSON object per line):
        ├── {"type": "system", "subtype": "init", "session_id": "...", "model": "...", ...}
        ├── {"type": "assistant", "message": {"content": [...]}}
        ├── {"type": "tool_progress", "tool_name": "Edit", "tool_use_id": "...", ...}
        ├── {"type": "system", "subtype": "hook_started", ...}
        └── {"type": "result", "session_id": "...", "total_cost_usd": 0.034, ...}
```

We implement this protocol directly in Rust (no Node.js dependency needed). The Agent SDK is reference code for the protocol — we speak the same wire format.

### Two Integration Channels

**Channel 1 — Spawn Mode (primary)**

The app spawns Claude Code as a subprocess using the stream-json protocol. Full control:

- Send prompts via stdin
- Receive all events via stdout (messages, tool calls, costs, hooks)
- Kill/interrupt the process
- Session lifecycle fully managed

This is how you "launch agents from the app."

**Channel 2 — Observe Mode (passive)**

For Claude Code sessions running independently in the user's terminal:

- **File watching:** Monitor `~/.claude/projects/*/sessions/*.jsonl` for live transcript updates
- **Webhook hooks:** App runs a local HTTP server (axum on `127.0.0.1:9111`). User configures Claude Code hooks to POST events to it.
- **Session listing:** Read session metadata via the same paths the Agent SDK's `listSessions()` uses

This is how you "see agents already running in terminals."

### Session Discovery

Claude Code stores sessions at:
```
~/.claude/projects/<encoded-cwd>/sessions/<session-id>.jsonl
```

Where `<encoded-cwd>` is the absolute path with non-alphanumeric chars replaced by `-`.

Each session JSONL line is a self-contained event. The app can:
1. List all sessions across all projects by globbing `~/.claude/projects/*/sessions/*.jsonl`
2. Read session metadata (first prompt, model, timestamps)
3. Tail active session files for real-time updates
4. Cross-reference with the global history at `~/.claude/history.jsonl`

### Hooks as Structured Events

Claude Code hooks (25+ lifecycle events) can fire HTTP webhooks:

```json
// In project's .claude/settings.json (auto-configured by Bonsai sensors)
{
  "hooks": {
    "PreToolUse": [{
      "hooks": [{
        "type": "http",
        "url": "http://127.0.0.1:9111/webhook",
        "headers": { "X-Session-Id": "$SESSION_ID" }
      }]
    }],
    "PostToolUse": [{
      "hooks": [{
        "type": "http",
        "url": "http://127.0.0.1:9111/webhook"
      }]
    }],
    "SessionStart": [{
      "hooks": [{
        "type": "http",
        "url": "http://127.0.0.1:9111/webhook"
      }]
    }],
    "Stop": [{
      "hooks": [{
        "type": "http",
        "url": "http://127.0.0.1:9111/webhook"
      }]
    }]
  }
}
```

Hook payloads include: `session_id`, `cwd`, `hook_event_name`, `tool_name`, `tool_input`, `transcript_path`.

**Bonsai integration:** A new Bonsai sensor (`companion-bridge` or similar) could auto-generate these hook configs, so any Bonsai-scaffolded project automatically reports to the companion app.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Tauri App                                                   │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │  Svelte 5 Frontend                                       │ │
│  │                                                          │ │
│  │  ┌────────────┐ ┌──────────────┐ ┌──────────────────┐   │ │
│  │  │ Project    │ │ Activity     │ │ Session          │   │ │
│  │  │ Sidebar    │ │ Feed         │ │ Detail           │   │ │
│  │  │            │ │ (streaming)  │ │ (messages, tools)│   │ │
│  │  └────────────┘ └──────────────┘ └──────────────────┘   │ │
│  │  ┌────────────┐ ┌──────────────┐ ┌──────────────────┐   │ │
│  │  │ Cost       │ │ Settings     │ │ Agent            │   │ │
│  │  │ Tracker    │ │              │ │ Config Viewer    │   │ │
│  │  └────────────┘ └──────────────┘ └──────────────────┘   │ │
│  └──────────────────────────┬──────────────────────────────┘ │
│              Tauri IPC      │                                 │
│         (Commands + Channels + Events)                       │
│  ┌──────────────────────────┴──────────────────────────────┐ │
│  │  Rust Backend                                            │ │
│  │                                                          │ │
│  │  ┌─────────────────┐  ┌──────────────────────────────┐  │ │
│  │  │ Process Manager │  │ Webhook Server (axum :9111)  │  │ │
│  │  │                 │  │                               │  │ │
│  │  │ spawn/kill/     │  │ POST /webhook ──► parse ──►  │  │ │
│  │  │ stream claude   │  │ store in SQLite + emit to    │  │ │
│  │  │ processes       │  │ frontend via Channel          │  │ │
│  │  └────────┬────────┘  └──────────────────────────────┘  │ │
│  │           │                                              │ │
│  │  ┌────────▼────────┐  ┌──────────────────────────────┐  │ │
│  │  │ Session Watcher │  │ SQLite Database              │  │ │
│  │  │                 │  │                               │  │ │
│  │  │ notify crate    │  │ projects, sessions,           │  │ │
│  │  │ watches         │  │ activity_log, costs,          │  │ │
│  │  │ ~/.claude/      │  │ agent_configs                 │  │ │
│  │  └─────────────────┘  └──────────────────────────────┘  │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌──────────────┐                                            │
│  │ System Tray  │  Runs in background, shows active agents   │
│  └──────────────┘                                            │
└─────────────────────────────────────────────────────────────┘
        │                              │
        │ stdin/stdout (NDJSON)        │ HTTP POST (webhooks)
        ▼                              ▼
  ┌───────────┐                  ┌───────────┐
  │ claude    │ (spawned)        │ claude    │ (independent,
  │ --print   │                  │           │  in user's terminal)
  │ --stream  │                  │           │
  └───────────┘                  └───────────┘
```

### Data Flow

**Spawn mode:**
```
User clicks "New Session" in app
  → Rust spawns `claude --print --output-format stream-json ...`
  → stdout NDJSON lines parsed in Rust
  → Each event: stored in SQLite + streamed to frontend via Tauri Channel
  → Frontend renders activity feed in real-time
  → User sends follow-up message: frontend → Rust command → write to stdin
```

**Observe mode:**
```
User runs `claude` in their own terminal
  → Session JSONL file created at ~/.claude/projects/.../sessions/xxx.jsonl
  → notify crate detects new/modified file
  → Rust reads new lines, parses events
  → Stored in SQLite + streamed to frontend via Tauri Channel

AND/OR:

  → Claude Code hook fires HTTP POST to 127.0.0.1:9111
  → axum receives, parses, stores, emits to frontend
```

### IPC Strategy

| Pattern | Mechanism | Use |
|---------|-----------|-----|
| CRUD operations | Tauri Commands | List projects, get settings, create session |
| Real-time streaming | Tauri Channels | Process stdout → activity feed, webhook events → dashboard |
| Cross-window notifications | Tauri Events | "Agent finished in Project A" broadcast to all windows |

---

## Data Model (SQLite)

```sql
-- Projects the user has registered
CREATE TABLE projects (
    id          TEXT PRIMARY KEY,           -- UUID
    name        TEXT NOT NULL,
    path        TEXT NOT NULL UNIQUE,       -- absolute filesystem path
    agent_type  TEXT,                       -- bonsai agent type if applicable
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    last_active TEXT
);

-- Claude Code sessions (discovered or spawned)
CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,           -- Claude Code session UUID
    project_id  TEXT NOT NULL REFERENCES projects(id),
    mode        TEXT NOT NULL,              -- 'spawned' | 'observed'
    status      TEXT NOT NULL DEFAULT 'idle', -- 'idle' | 'running' | 'stopped' | 'error'
    model       TEXT,
    title       TEXT,                       -- display name
    first_prompt TEXT,
    started_at  TEXT,
    ended_at    TEXT,
    total_cost  REAL DEFAULT 0,
    input_tokens  INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    num_turns   INTEGER DEFAULT 0,
    jsonl_path  TEXT                        -- path to ~/.claude session file
);

-- Activity events (unified from both spawn + observe modes)
CREATE TABLE events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    timestamp   TEXT NOT NULL DEFAULT (datetime('now')),
    event_type  TEXT NOT NULL,              -- 'tool_use' | 'message' | 'error' | 'hook' | 'system'
    tool_name   TEXT,                       -- for tool_use events
    summary     TEXT,                       -- human-readable one-liner
    payload     TEXT,                       -- full JSON event (stored for drill-down)
    cost_usd    REAL DEFAULT 0
);

-- Cost tracking (aggregated per day per project)
CREATE TABLE daily_costs (
    date        TEXT NOT NULL,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    total_cost  REAL DEFAULT 0,
    input_tokens  INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    session_count INTEGER DEFAULT 0,
    PRIMARY KEY (date, project_id)
);
```

---

## Core Features

### MVP (v0.1)

1. **Project registry** — Add/remove projects by path. Auto-detect `.bonsai.yaml` and `CLAUDE.md` to identify agent-configured projects.

2. **Session discovery** — Scan `~/.claude/projects/*/sessions/*.jsonl` to find all sessions. Show them grouped by project with status, model, cost, timestamp.

3. **Activity feed** — For any session (spawned or observed), show a real-time streaming feed of what the agent is doing: messages, tool calls (file edits, bash commands, searches), costs per turn.

4. **Spawn sessions** — Launch a new Claude Code session from the app. User types a prompt, app spawns `claude --print --output-format stream-json`, streams results into activity feed.

5. **Cost dashboard** — Track spending per project and per session. Show daily/weekly totals.

6. **System tray** — App runs in background. Tray icon shows number of active sessions. Notifications when an agent finishes or errors.

### v0.2

7. **Observe mode (webhooks)** — Local axum server receives hook events from independently running Claude Code sessions. Richer real-time data than file watching alone.

8. **Bonsai integration** — Read `.bonsai.yaml` to show installed agent type, skills, protocols, sensors. "Open in terminal" button that launches Claude Code in the project directory.

9. **Session management** — Resume sessions, name them, tag them. Fork a session to try a different approach.

10. **Multi-turn conversations** — In spawned sessions, send follow-up messages (streaming input via stdin). Full conversational agent interaction from within the app.

### v0.3

11. **Agent config viewer** — Read and display the agent's CLAUDE.md, identity, skills, protocols. Understand what the agent is configured to do.

12. **Search across sessions** — Full-text search over all session transcripts. "Where did I fix that auth bug?"

13. **Notifications + alerts** — Desktop notifications when: agent finishes, agent errors, cost exceeds threshold, permission needed.

14. **Auto-update** — GitHub Releases + tauri-plugin-updater with signed binaries.

### Future

- Embedded terminal (xterm.js) for full interactive sessions inside the app
- Managed Agents integration (cloud sessions alongside local)
- Plugin system for non-Claude-Code agents
- Team features (shared project dashboards)
- Keyboard-driven workflow (command palette, vim-style navigation)

---

## Agent-Agnostic Design

Although we build for Claude Code first, the architecture should accommodate other agents later.

### Abstraction Layer

```rust
trait AgentBackend {
    /// Spawn a new agent session
    async fn spawn(&self, config: SpawnConfig) -> Result<SessionHandle>;
    
    /// List existing sessions for a project
    async fn discover_sessions(&self, project_path: &Path) -> Result<Vec<SessionInfo>>;
    
    /// Read events from a session (live or historical)
    async fn stream_events(&self, session_id: &str) -> Result<impl Stream<Item = AgentEvent>>;
    
    /// Send a message to a running session
    async fn send_message(&self, session_id: &str, message: &str) -> Result<()>;
    
    /// Interrupt/kill a session
    async fn interrupt(&self, session_id: &str) -> Result<()>;
}

/// Normalized event type (agent-agnostic)
enum AgentEvent {
    Message { role: Role, content: String },
    ToolUse { tool: String, input: Value, output: Option<String> },
    Cost { usd: f64, input_tokens: u64, output_tokens: u64 },
    Status { state: SessionState },
    Error { message: String },
}
```

The `ClaudeCodeBackend` implements this trait using the stream-json protocol + session file discovery. Future backends (Codex CLI, Aider, etc.) would implement the same trait.

---

## UI Concepts

### Main Window Layout

```
┌─ Sidebar ──────┬─ Main Panel ──────────────────────────────┐
│                 │                                            │
│  PROJECTS       │  Project: my-api                          │
│                 │  Agent: backend  ·  3 sessions             │
│  ● my-api    ← │                                            │
│    backend      │  ┌─ Active Session ─────────────────────┐ │
│    2 active     │  │ "Fix auth token expiry" · ● running  │ │
│                 │  │ claude-opus-4-6 · $0.12 · 4m elapsed │ │
│  ○ web-app      │  │                                      │ │
│    fullstack    │  │  > Edit src/api/auth.go:47            │ │
│    idle         │  │    + validate JWT expiry field        │ │
│                 │  │  > Bash: go test ./auth/...           │ │
│  ○ infra        │  │    ✓ 12 passed                       │ │
│    devops       │  │  > "Fixed the token validation..."   │ │
│    idle         │  │  > Edit src/api/middleware.go:23      │ │
│                 │  │    + add expiry check to middleware   │ │
│─────────────────│  │                                      │ │
│                 │  └──────────────────────────────────────┘ │
│  QUICK ACTIONS  │                                            │
│                 │  ┌─ Recent Sessions ────────────────────┐ │
│  + New Session  │  │ "Add pagination" · ○ done · $0.08    │ │
│  ⚡ Open Term   │  │ "Fix CI pipeline" · ○ done · $0.23   │ │
│                 │  └──────────────────────────────────────┘ │
│                 │                                            │
│  STATS (today)  │  ┌─ Cost Chart ────────────────────────┐ │
│  $1.24 spent    │  │  ▁▂▃▅▇█▆▃▂▁  (7 day trend)         │ │
│  47 tool calls  │  └──────────────────────────────────────┘ │
│  3 sessions     │                                            │
└─────────────────┴────────────────────────────────────────────┘
```

### Design Principles

- **Glanceable** — See status of all projects and agents at a glance from the sidebar
- **Non-intrusive** — Runs in background, notifies when attention needed
- **Fast** — Native app, sub-second startup, streaming updates at 60fps
- **Keyboard-first** — Command palette, hotkeys for common actions
- **Dark mode default** — Developers live in dark mode

---

## Relevant CLI Flags Reference

Flags used when spawning Claude Code from the app:

| Flag | Purpose |
|------|---------|
| `--print` | Non-interactive mode |
| `--output-format stream-json` | NDJSON event stream on stdout |
| `--input-format stream-json` | Accept messages as JSON on stdin |
| `--include-hook-events` | Include hook lifecycle in stream |
| `--include-partial-messages` | Token-level streaming (optional, for live typing effect) |
| `--session-id <uuid>` | Use specific session ID |
| `--resume <id>` | Resume existing session |
| `--name <title>` | Set display name |
| `--model <model>` | Set model |
| `--permission-mode <mode>` | Control permission behavior |
| `--max-budget-usd <amount>` | Spending cap |
| `--max-turns <N>` | Turn limit |
| `--allowedTools <tools>` | Pre-approve tools |
| `--cwd <path>` | Working directory (implicit from spawn) |

### Stream Event Types (stdout NDJSON)

| Type | Subtype | What it contains |
|------|---------|-----------------|
| `system` | `init` | Session ID, model, tools, cwd, version |
| `assistant` | — | Complete assistant message with content blocks |
| `user` | — | Tool results, user messages |
| `tool_progress` | — | Tool name, elapsed time |
| `result` | — | Final result: cost, tokens, duration, stop_reason |
| `system` | `api_retry` | Retry info (attempt, delay, error) |
| `system` | `hook_started/progress/response` | Hook lifecycle |
| `system` | `status` | Status changes (compacting, etc.) |
| `system` | `compact_boundary` | Context compaction occurred |
| `rate_limit_event` | — | Rate limit status |

---

## Open Questions

### Decided

- **Name:** Greenhouse — pairs with Bonsai (you grow bonsai in a greenhouse)
- **Frontend:** Svelte 5 (SvelteKit SPA mode) — smallest bundle, fastest runtime, best Tauri pairing
- **Window model:** Single window with sidebar navigation for MVP

### Open

1. **Permission handling** — When a spawned session needs permission (tool approval), how do we surface that in the UI? Inline in the activity feed with approve/deny buttons? Or always use `--permission-mode auto`?

2. **How much of the Agent SDK protocol to implement** — The SDK handles edge cases (retries, reconnection, process recovery). How much of that do we reimplement in Rust vs. keeping it simple for MVP?

3. **Observe mode fidelity** — File watching `~/.claude/` gives us transcript data but with some delay. Webhook hooks give real-time events but require user to configure hooks. What's the default experience?

4. **Bonsai sensor for auto-registration** — Should Bonsai auto-install a sensor that registers the project with the companion app? Or should the companion app discover Bonsai projects by scanning for `.bonsai.yaml`?

---

## References

- [Claude Code Agent SDK (TypeScript)](https://code.claude.com/docs/en/agent-sdk/typescript)
- [Claude Code Agent SDK (Python)](https://code.claude.com/docs/en/agent-sdk/python)
- [Claude Code Hooks Reference](https://code.claude.com/docs/en/hooks)
- [Claude Code CLI Reference](https://code.claude.com/docs/en/cli-reference)
- [Claude Code Headless Mode](https://code.claude.com/docs/en/headless)
- [Tauri v2 Documentation](https://v2.tauri.app/)
- [Tauri Shell Plugin](https://v2.tauri.app/plugin/shell/)
- [Tauri IPC (Channels)](https://v2.tauri.app/develop/calling-frontend/)
- [claude-devtools](https://github.com/matt1398/claude-devtools) — closest prior art (Electron)
- [Claude Squad](https://github.com/smtg-ai/claude-squad) — process management patterns
- [Mission Control](https://github.com/builderz-labs/mission-control) — feature scope reference
- [Claude Managed Agents](https://platform.claude.com/docs/en/managed-agents/overview) — future cloud integration
