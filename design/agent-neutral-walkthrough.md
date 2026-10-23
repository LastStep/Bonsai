# Bonsai with any agent: the walkthrough, and the choices before the nights

- For: the session that walks the director through `design/agent-neutral.md` (the research, `T-5797`) and settles the
  choices, so that the sessions after it can build the core without the director.
- Written: 22 Oct 2026, by the orchestrator that wrote the research (Opus 5.5), at the director's word: "for the next
  session i want a proper walkthrough for this doc, and the choices i have to make. my aim is to get the bonsai parts
  done, so it becomes provider agnostic, and ideally we can finish it in one overnight autonomous session". Its night
  plan was reviewed by a fresh Opus agent (`M-2026-10-22-479c8b5e`: FAIL on the first sketch, for a hidden person
  step, hidden serial chains and an optimistic size; `M-2026-10-22-22f3ed12`: PASS on this version, its five small
  points folded in).
- How to use it: the orchestrator reads this file and section 0 of the research, then runs part 1 (the tour, about 20
  minutes with the director), part 2 (the choices, one question at a time, context first), part 3 (the plan, written
  after the choices, reviewed by a fresh Opus agent, approved), and part 4 (the director's steps before the first
  night). Everything the questions need is here; the research holds the evidence.

## 0. The honest answer to "one night"

**The core does not fit one night; it fits two autonomous sessions** (two nights, or a night and a day), if the
choices, the plan and every grant are settled in the walkthrough session first, so that neither session meets a
person's step.

- **Why not one night:** the core is about 37-68 hours of agent work (the research's section 9, phases 0, 1 and 3 with
  the parts of phase 2 the core needs). It also runs through serial chains:
  - the identity function, then the records that use it;
  - the envelope format, then pi's reader, then the hook commands;
  - the pi extension, then its setup command, then the checks that read it;
  - then the proof, then the end verifiers.

  This repo's record for a step of this size is step 5.4: about 39 agent hours over about 14.5 hours of wall clock,
  with four verifier rounds on one piece. CI on a branch takes about 27 minutes per batch (reviewer, from `gh run
  list`).
- **What makes two sessions enough:** the translate route (choice C2): a Go reader per agent turns that agent's events
  into the vocabulary Bonsai uses inside today. It works because the guard already takes its environment through an
  injected function (`internal/guard/hook.go:31,55,123`), and its judges take a payload struct, not Claude's JSON.
  That keeps the guard, stop gate and recorder mostly as they are. The full rename to neutral kinds is a later
  clean-up.
- **What pi has after each session:**
  - After the first: nothing it can use yet. Bonsai is ready underneath: identity for every agent, pi's reader, the
    hook commands for pi, Bonsai's own stop cap, pi's wiring on the floor.
  - After the second: pi guarded, held at stop, recorded and given its opening context, proved end to end, with
    Bonsai saying what holds.
- **Kept for later parts:**
  - roles on pi;
  - the sandbox, the rung sandbox, signed results;
  - the read walls and the shell-word tripwire;
  - delivery (the block in `AGENTS.md`, skills for pi);
  - the pack texts;
  - a second agent;
  - refusing other agents that run Claude's hook lines (it needs a Cursor test first).

## 1. The tour (about 20 minutes)

Walk the research section by section; for each, say the point in plain words and show the one table or line named.

| Research section | The point, in plain words | Show |
|---|---|---|
| 0 | The one-page story | Read it together |
| 1 | How it was made: eleven helpers, five live pi tests, two verifiers who failed the first draft and passed the second | The evidence table |
| 2 | Bonsai's core already works under any agent; about 11% of the Go code, 121 places and 21 pack lines assume Claude Code. **The root cause:** missing evidence is read as the highest privilege (no Claude marker means "a person"; no subagent id means "the director's own session") | 2.3, item 4 |
| 3 | The field: most agents copied Claude's hooks; most fail open; three already run Claude's hook settings; `AGENTS.md` and Agent Skills are nearly everywhere; nobody else has a stop gate or proof ladder | The table in 3.2 |
| 4 | pi in depth. It can host everything; 45 lines and today's binary refused every protected write. Its traps: an extension loaded after the guard can rewrite a call; pi waits on stdin under a script; pi runs edited copies from its compile cache (P5) | 4.2 (P1, P2, P4, P5), 4.3 |
| 5 | Security. With one OS user and no sandbox nothing is a wall, under any agent (decision D holds). A wall needs the agent's shell sandboxed with Bonsai outside it. 14 attack paths, each with its fix | 5.2, the table in 5.4 |
| 6 | The design: the host seam (the diagram in 6.3), the profile of what holds per agent (6.4), delivery, the guard and stop gate per agent, identity, walls, proof, the pi extension | The diagram; the table in 6.4 |
| 7-9 | The nine research choices; the 1.0 answer; the phases | 7, 8 |
| 10-11 | What is still open; how it was checked | 11 |

Two findings to give the director whatever they choose, because they touch their own pi setup, not Bonsai:

- **pi's compile cache runs edited code.** pi compiles TypeScript extensions into `/tmp/jiti/` and runs that copy; an
  edited copy ran (P5). Any agent with a shell can plant code in the director's harness extensions that way. Turning
  the cache off (`JITI_FS_CACHE=false` in the environment pi starts from) closes it, at some start-up cost.
- **pi started from inside a Claude Code session inherits Claude's variables**, a messaging token among them (P3).

## 2. The choices, in the order to ask them

Ask one at a time, context first, two to four options, the recommendation first (the director's rule). The first
nine shape the two sessions; the rest wait for their phase, and the session says so instead of asking them.

### Shape the sessions (ask now)

**C1. When, against 1.0.** 1.0 is a few sessions away: rc.2's fixes (`T-5710`, about twelve items), rc.2, the
director's own trial, 1.0. The two sessions change the guard and the stop gate, which 1.0 should ship already proven.
- (a) **1.0 first, then the two sessions (recommended).** 1.0 ships as built for Claude Code, as its README now says;
  the core lands as the start of 1.1. Costs the wait.
- (b) **The two sessions first, then rc.2 carries the core.** pi works sooner; 1.0 slips by the two sessions plus a
  round of release checks on the changed guard.
- (c) Both at once. Not recommended: rc.2's fixes and the core touch the same files.

**C2. How deep the seam goes.**
- (a) **Translate (recommended).** A Go reader per agent turns that agent's events into the vocabulary Bonsai uses inside
  today (`Edit`, `MultiEdit`, `Write`, `Bash`, `Read` as Bonsai's own words). The guard, stop gate and recorder change
  little; every judgement stays in Go.
- (b) **Neutral kinds** (research 6.3: actions by kind, and 19 tool names compared in 14 files rewritten). Cleaner,
  25-45 hours alone.

**C3. The shape of the two sessions.**
- (a) **Foundations, then pi wired (recommended).**
  - First session: identity, the envelope, pi's reader, the hook commands, Bonsai's stop cap and pi's wiring on the
    floor. pi's extension is built alongside, against the envelope the plan fixes, but not landed.
  - Second session: the extension landed, its setup command, the honesty checks, the end-to-end proof and the end
    verifiers.
- (b) **A vertical slice first:** a minimal pi guard and stop gate end to end in the first session, floor and honesty
  in the second. pi is usable a session sooner, with its wiring unprotected for that time; the serial chain is
  longer, so the first night is more likely to spill.
- (c) **One night, then stop:** the foundations only, and decide later. Safe, but pi gains nothing it can use yet.

**C4 (research D4). Person-only commands.** Today only Claude's marker refuses `update --yes --allow-exec`,
`unlink --yes` and looser checks.
- (a) **A wider tripwire (recommended):** every marker of C8's list refuses, and on Linux and WSL a CLI agent among
  the parents does too; the refusal names what it found. On Windows the process table gives only program names (pi
  runs as `node.exe`), so there it is markers only until a later piece; the code says so.
- (b) (a) plus an OS ceremony (`sudo -k`, Windows Hello or a key touch) for the strongest commands. A gesture every
  time; a later piece.
- An agent that hides on purpose still gets past either; the research says so (6.7).

**C5 (research D2). Code in pi's own language.** pi has no command hooks: it needs a small extension.
- (a) **A small plain-JavaScript extension (recommended),** about 45 lines, built from the same commit, carried inside
  the binary, holding no logic (it passes events and obeys). Plain JavaScript, because pi runs edited copies of
  compiled TypeScript (P5).
- (b) Nothing outside the one Go binary: pi gets no guard.

**C6. Where the pi extension lives, and who wires it in.** The safest place is root-owned beside the binary, but that is
an installer change (release files, the director's).
- (a) **A person's command `bonsai setup pi --yes` (recommended),** refused to agents. It writes the extension into
  Bonsai's home and adds it last to pi's user `packages`. `bonsai check` checks its hash and its place. At the next
  release the installer puts it root-owned beside the binary. The sessions test the command only against a scratch pi
  folder (`PI_CODING_AGENT_DIR`), never the director's.
- (b) Bonsai prints the settings line and the director adds it by hand.
- (c) Wait for the installer change before pi gets anything.

**C7. Formats and the contract.** The core adds one format and two fields, all additions:
- a new format `bonsai.hook/1`, the envelope pi's extension sends;
- `bonsai.status/1` gets a `hosts` list;
- `bonsai.workspace/1` gets `agents`, Claude Code by default.

The formats were confirmed final on 7 Oct, and every format set came with a contract section, so the plan carries a
draft contract section for these three. **Recommended:** yes to the three additions, and the director approves the
section with the plan.

**C8. The agent markers to trust.**
- **The list:** `CLAUDE_CODE_CHILD_SESSION`, `AI_AGENT`, `PI_CODING_AGENT`, `CODEX_THREAD_ID`, `GEMINI_CLI`,
  `CURSOR_AGENT`, `OPENCODE`, `GOOSE_TERMINAL`.
- **Left out on purpose:** `CLAUDECODE`, because it reaches editor terminals too.
- **Before saying yes, the director runs this in each terminal they type Bonsai commands in** (WSL, Windows Terminal,
  any editor's terminal). It must print nothing:

  ```bash
  env | grep -E '^(CLAUDE_CODE_CHILD_SESSION|AI_AGENT|PI_CODING_AGENT|CODEX_THREAD_ID|GEMINI_CLI|CURSOR_AGENT|OPENCODE|GOOSE_TERMINAL)='
  ```

  In Windows PowerShell:

  ```powershell
  Get-ChildItem Env: | Where-Object { $_.Name -match '^(CLAUDE_CODE_CHILD_SESSION|AI_AGENT|PI_CODING_AGENT|CODEX_THREAD_ID|GEMINI_CLI|CURSOR_AGENT|OPENCODE|GOOSE_TERMINAL)$' }
  ```

**C9. The proof's model login.** The end-to-end proof runs pi with a real model, and pi keeps its login in its own
settings folder (pi `dist/config.js:491-508`). Agents never touch a login.
- (a) **The proof uses the director's pi login, as this research's probes did (recommended).** pi runs with
  `-ne -nc -ns --no-session -e <the extension>`, on the cheapest model, a few cents. No settings or sessions are
  written; pi takes a lock beside its login file and may rewrite it on a token refresh, as it does for the director;
  and the model in the proof could read that login, since pi has no read walls (research S7). The extension is loaded
  with `-e`, so the wiring `bonsai setup pi` writes ("last in `packages`") is proved by Go tests and `check`, not by
  the live run.
- (b) The director logs pi in once in a scratch folder before the night
  (`PI_CODING_AGENT_DIR=~/bonsai-checks/pi-agent pi`, then `/login`).
- (c) No live model: the proof uses pi's built-in fake provider, which scripts tool calls. It needs its own small
  build piece, and then also fits CI later.

### Can wait for their phase (say so; do not ask now)

- Research D3 (walls: the sandbox, the rung sandbox, signed results, where the trusted records live): its own part,
  after the sandbox probe (it needs `socat`, a person's install).
- Research D5 (roles on pi), D7 (the second agent), D8 (the pack texts), D9 (a strict mode for unjudged tools): later
  parts.
- Research D6 (where skills go): with delivery.
- The refusal of other agents that run Claude's hook lines: after a Cursor test (research section 10).

## 3. The plan (written after the choices, then reviewed and approved)

The orchestrator writes it with `workflow:plan` in this repo's way: one task per piece, each with `done_when` from its
proof, its grants and its lane. A fresh Opus agent reviews it before the director sees it (CLAUDE.md). The sketch below
assumes C2 (a) and C3 (a); the plan sizes and orders it for real.

### Lanes, and the serial chains inside them

| Lane | Pieces, in order (each arrow a dependency) | Model | Verifier |
|---|---|---|---|
| A, identity | A1: one function for host, session, and agent or person, behind every read of Claude's variables today (4 of the child marker, 3 of the session id, 8 of the project dir); C8's markers; CLI-agent parents on Linux (Windows markers only, said in the code); refusals name what they found; the test helper clears the list; the guard's unset-watch covers it. A pi session counts as "the person's own session" only with no agent among its parents; otherwise its caller is `unknown` (so a `pi -p` an agent starts gets the least rights). The plan states what wins when a named session (`BONSAI_TASK`, `BONSAI_ROLE`) has an agent parent, as C3's proof will (it starts pi from a Claude session), and what stands in for the parent check on Windows (an outer agent's marker); C3's scratch project declares no roles unless the plan says how its session gets its role. **Then** A2: each record names its host, not `"claude-code"` (6 places, the same writers A1 touches); pi session ids mapped to the log's file rule | Opus | yes (guard) |
| B, the pi seam | B1: `bonsai.hook/1` (schema, manifest, examples, trick files, `check --schema`; its contract section is already
  committed with the plan). **Then** B2: pi's reader. It builds a Claude-shaped internal input from pi's call: pi's `edit` becomes `MultiEdit`, after pi's own argument repairs. The redactor and the input hash need that shape. It applies pi's six path rules and keeps pi's own event names for `agent_event`. Its pi tables live in the pi driver, never in `engine.GuardTools` or `recorder.Events`: those two build Claude's lines, and touching them changes Claude's settings. **Then** B3: `bonsai hook guard\|stop\|start\|record --host pi`. The envelope is read before `guard.Main` and its frame timer, because the project dir is read before the payload today. Usage comes from the envelope: a new recorder path, since today's reads only Claude's transcript. **Beside them**, B4: the stop cap in Bonsai's own stop records | Opus | yes (guard, stop gate) |
| C, the pi extension | C1, built in the first session against the envelope the plan fixes: the plain-JavaScript extension. It runs last. `tool_call` calls the guard by its fixed path within 10 s and blocks on anything but allow. `tool_result` compares the input with its judged copy. The settle handler is wrapped: an error counts as a hold, and it keeps the other handlers' entries. Start and compaction give the opening context as a system-prompt section. Records fire and forget; usage comes from `message_end`; a `--bonsai-task` flag; it knows nested agents. Its file lives under `packs/**`, which git keeps LF on Windows checkouts: its hash is checked. **Then** C2: carried in the binary; `bonsai setup pi --yes` (C6); `check` verifies its hash, its last place and no project override. **Then** C3, the proof | Opus for C1, Sonnet for C2 | C3 on the branch, then the end verifiers |
| D, protection and honesty | D1 (independent; the first batch): pi's wiring on the guard's floor (the project's `.pi/settings.json`, `.pi/extensions/**` and `.pi/sandbox.json` never granted; writes to `~/.pi/agent/` and the compile cache refused to file tools), copying `guard/floor.go` and `guard/own.go`; starts of `pi`, `codex`, `gemini`, `opencode` and `droid` from an agent's shell refused under roles, as `claude` is (`guard/rolecmds.go:107-168`). D3 (after C2): `status --json` `hosts`; `bonsai.yaml` `agents`; `check` warnings (extension not loaded, not last, pi version outside the proved range); the unhooked-session finding | Opus for D1, Sonnet for D3 | yes (guard) |

### The two sessions

**Session 1 (a night, about 10-12 hours of wall clock):**
- Batch 1: A1, then A2; B1; D1.
- Batch 2: B2, then B3; B4.
- C1 built alongside on its own branch, not landed.
- Each batch climbs the floor, then lands by CI on the branch, every job green, both Windows jobs among them. Each
  guard or stop-gate piece has its fresh verifier before landing.

**Session 2 (about 6-8 hours):**
- Batch 3: C1, then C2, then D3.
- C3, the proof, runs on batch 3's branch before it lands, as C1's own proof: the ladder has no rung for JavaScript.
  In a scratch project, with a stamped build of that branch and pi as C9 says, the pi trial's steps 2.4-2.6 are run
  again under a named session:
  - a protected edit is refused, through every path form;
  - `rm` of a protected path is refused;
  - a red ladder holds the stop until it is green;
  - the session's records name pi;
  - a later extension's rewrite is recorded as a fault.

  The person-only tripwire is proved by A1's Go tests, not by C3: a stamped build skips it, and pi started from Claude
  inherits Claude's marker.
- Then two end verifiers, one per lens (security, then tests and correctness), over every landing.

**What the director wakes up to after session 2:**
- the report, with each landing's sha and CI run;
- install lines for a pre-release built from a clean clone (theirs to install);
- one command, `bonsai setup pi --yes`, run in their own terminal;
- a two-minute check they can do under pi.

## 4. The director's steps before the first session

1. Answer C1-C9 (part 2).
2. Approve the plan, its contract section for C7, and each task's grants, once the plan's review is in (part 3).
   - The core changes neither Bonsai's own hook lines nor `bonsai.yaml`, so it needs no person-only path: "no `--host`
     flag means Claude Code", provided the pi tables stay out of `engine.GuardTools` and `recorder.Events` (B2).
   - `design/contract.md` is protected: its new section is committed with the approved plan, in the records commit
     (as `T-5795` did with contract section 23), so no batch task holds a grant and the batches can run side by side.
     The extension's file goes under `packs/**`, which git already keeps LF, so `.gitattributes` is not touched.
3. Run the marker check of C8 in each terminal.
4. Leave pi as it is (1.4.0). With C9 (a), no pi settings or sessions are written (pi may refresh its own login
   file, as it always does); the setup command is tested against a scratch pi folder.
5. Optional, any time: `JITI_FS_CACHE=false` for pi sessions (the compile-cache finding).

## 5. Questions ready to ask

The orchestrator asks these one at a time, each after its context above, with the recommended option first:

1. C1 "When should the provider-agnostic core land: after 1.0 ships, or before rc.2?"
2. C2 "How deep should the seam go: translate each agent into Bonsai's existing words, or the full rename to neutral
   kinds?"
3. C3 "How should the two sessions be split: foundations then pi wired, a vertical slice first, or one night of
   foundations only?"
4. C4 "How should Bonsai tell an agent from you for person-only commands?"
5. C5 "May Bonsai ship a small plain-JavaScript extension for pi, carried inside the binary?"
6. C6 "Where should the pi extension live, and who wires it in?"
7. C7 "May the core add the envelope format and the two fields, with their contract section?"
8. C8 "Is this the list of agent markers to trust?" Ask the terminal check first.
9. C9 "How should the end-to-end proof get a model?"
