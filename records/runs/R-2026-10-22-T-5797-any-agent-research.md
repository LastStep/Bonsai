---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-22-T-5797
task: T-5797
role: orchestrator
model: opus
started: 2026-10-22 12:15
finished: 2026-10-22 13:31
outcome: needs-person
commits: []
labels: {}
---

# Run: T-5797, Bonsai with any agent, pi first

- Who: the orchestrator (opus), with researchers, auditors and a verifier
- Why: the director, 22 Oct: Bonsai fails with agents other than Claude Code; a master research document for later
  plans, at the fundamental level, security a core principle

## Log

- 12:15 The director's brief; read the pi trial's `FINDINGS.md`, `PI-ADAPTER.md`, `PI-PIECES.md`, its picture and
  `AGENTS.md`; STATE.
- 12:20 Asked four questions. Answers: any agent, pi first; the 1.0 question "figure this out at the end of the
  research with your suggestions and reasoning"; pi tests in scratch; the document in `design/`.
- 12:22 Read the spec's sections 2, 5 (opening) and 7. Task `T-5797` written at `plan` (three tasks of another session
  read `running`; this one grants nothing). Scratch: `~/bonsai-checks/T-5797/`.
- 12:29 Shared rules `~/bonsai-checks/briefs/common-2026-10-22-T-5797.md`; the host matrix N1-N13
  `~/bonsai-checks/T-5797/needs-matrix.md`. Eleven helpers started, research only, writing nothing:
  - researcher (sonnet) `M-2026-10-22-2a41fbe0`: hooks of Codex CLI, Gemini CLI, Copilot CLI
  - researcher#2 (sonnet) `M-2026-10-22-21ffaccb`: Cursor, Droid, OpenCode
  - researcher#3 (sonnet) `M-2026-10-22-005e7092`: Amp, Goose, Cline, Kiro, Windsurf, Aider (shallow)
  - researcher#4 (sonnet) `M-2026-10-22-ddb1a550`: AGENTS.md, Agent Skills, shared agent/plugin formats, agent markers
  - researcher#5 (sonnet) `M-2026-10-22-940c0286`: ACP and OpenTelemetry as neutral layers
  - researcher#6 (opus) `M-2026-10-22-389008b9`: OS enforcement and the trust boundary (security)
  - researcher#7 (sonnet) `M-2026-10-22-3edf1f7b`: prior art across agents
  - auditor (sonnet) `M-2026-10-22-31cbe7b9`: Claude couplings in Bonsai's Go code
  - auditor#2 (sonnet) `M-2026-10-22-60add1a2`: couplings in formats, contract, packs, docs; versioning class
  - auditor#3 (sonnet) `M-2026-10-22-c49910b2`: pi 1.4.0's extension lifecycle
  - auditor#4 (sonnet) `M-2026-10-22-ddd89628`: pi's tools, paths, env, loading, sessions, sandbox; pi-harness
- 12:42-12:51 Live probes in scratch (`~/bonsai-checks/T-5797/probes/`, pi 1.4.0, claude-haiku-5-5): P1 stop gate in
  `-p` and json (holds reach the model; shutdown awaited), P2 handler order (a later extension's rewrite gets past a
  guard loaded earlier), P4 a 60-line prototype calling rc.1's `bonsai hook guard` in the pi trial (4 refusals, 1
  allow, 6-13 ms; records labelled `claude-code`; first run hung on stdin), P3 the shell's environment (markers pile
  up; an outer Claude session's variables, a token among them, reach pi's shell: no value written anywhere, checked).
  The trial folder left clean (`git status` empty).
- 12:39-13:00 Reports in (their `created` stamps): auditor#2 `M-2026-10-22-f7bbc3c9`, auditor#3 `M-2026-10-22-f2123ac5`, researcher#3
  `M-2026-10-22-e08055c9`, auditor#4 `M-2026-10-22-a8ac9080`, researcher#7 `M-2026-10-22-a85b1d5f`, auditor
  `M-2026-10-22-79095a89` (+ `-0bb7b366`), researcher#2 `M-2026-10-22-053bfa71`, researcher#6 `M-2026-10-22-b7f19b73`
  (read in full), researcher#4 `M-2026-10-22-5e8f7007` (stopped by its 30-minute budget, every question answered).
  Landlock ABI 7 read on this kernel (a read-only syscall). Claude Code 2.1.308 sets
  `AI_AGENT=claude-code_2-1-308_agent` in its shell (ran).
- 12:55- `design/agent-neutral.md` written section by section (sections 1, 2, 4, 5, 6, 7, 8, 9, 10 so far).
- 13:00-13:02 The last two reports: researcher#5 `M-2026-10-22-3ce2d917` (ACP, OpenTelemetry; budget-stopped,
  answered) and researcher `M-2026-10-22-b4f04faf` (Codex, Gemini, Copilot; budget-stopped, Copilot partly). No
  helper process left (`ps` checked; the CI watcher seen belongs to another session).
- 13:05 The draft complete (sections 0-10, about 81 KB). The orchestrator's own pass: section references fixed in the
  attack-path table; the director's harness unnamed; `CLAUDECODE` taken off the person-proof marker list (it reaches
  editor terminals, researcher#4) and a precondition added to the rc.2 fix; role tests added to 6.11; who climbs under a
  sandbox made an open point (6.9).
- 13:06 Two fresh Opus verifiers, 60 minutes each: verifier (facts) `M-2026-10-22-d52df3ec`, verifier#2 (security)
  `M-2026-10-22-df7e3c5c`.
- 13:06-13:20 Round 1: verifier (facts, opus) `M-2026-10-22-1e25eeb4` FAIL, 1 blocking (Copilot CLI's opt-in OS
  sandbox misread as "best-effort rules") and 18 minor; verifier#2 (security, opus) `M-2026-10-22-64dc2c2d` FAIL, 5
  blocking (B1 the rc.2 fix overclaimed and missed three of the four read sites; B2 signing holds only behind a sandbox
  read wall; B3 the rung sandbox claim unproven; B4 pi rated "enforced" though other host code and a user-writable
  transpile cache can defeat it; B5 MCP and extension tools unjudged) and 10 minor; hygiene passed. The orchestrator
  checked B4's cache (`/tmp/jiti/`, user-owned, holds the probes' and the harness's compiled extensions; ran) and the
  four read sites and the test helper (pointed). Every finding fixed in the text but one: the Landlock grade (#3, M8)
  stays "ran", since the orchestrator ran the query itself (12:5x, above). Before-copy kept at
  `~/bonsai-checks/T-5797/agent-neutral.before-vf.md`.
- 13:24 Round 2 sent to both verifiers in their threads (`M-2026-10-22-c49b48a0`, `M-2026-10-22-55516867`), 30
  minutes each.
- 13:25 Probe P5 (S14), cache pointed into scratch (`TMPDIR` + `JITI_RESPECT_TMPDIR_ENV=1`; the real
  `/tmp/jiti/` untouched): a TypeScript extension compiled once, its cached copy edited, the next start ran the edited
  copy ("TAMPERED"); a plain `.mjs` extension left no cache file. Logs `~/bonsai-checks/T-5797/probes/out/s14.log`.
  Written into sections 4.2, 4.3, 5.4 (S14), 6.10 and 10. No pi process left (ps).
- 13:27-13:28 Round 2 (their `created` stamps): verifier (facts) `M-2026-10-22-59a8f666` PASS, 4 minor; verifier#2
  (security) `M-2026-10-22-aff025dc` PASS, 10 minor. All 14 fixed in the text (the citation `rule.go:319` read by the
  orchestrator: the facts verifier was right). Section 11 written.
- 13:31 Done: `design/agent-neutral.md` (about 98 KB), not committed; the task to `verify` until the director has read
  it; then a records commit (the document, the task, this report, STATE). STATE not rewritten now: another session is
  working in this checkout.

## Processes left running

None: every helper reported none of its own (two could not run their last check after their budget; the orchestrator
checked with `ps -eo pid,etime,cmd` after each report); every pi probe ended (no `pi` process at 13:29).

## Left

- The director reads `design/agent-neutral.md` (section 0 first) and answers D1-D9 when a plan is written from it.
- rc.2's list (`T-5710`) could take section 8's two items, after the director's word.
- Scratch kept: `~/bonsai-checks/T-5797/` (notes, probes with logs, two earlier copies of the document).
- 13:32-13:59 The director: "for the next session i want a proper walkthrough for this doc, and the choices i have to make.
  my aim is to get the bonsai parts done, so it becomes provider agnostic, and ideally we can finish it in one overnight
  autonomous session". The orchestrator checked the translate route in the code (`internal/guard/hook.go:31,55,123`:
  the guard's environment is injected) and rc.2's distance (`T-5710`: about twelve items), then wrote
  `design/agent-neutral-walkthrough.md`: the honest answer (all phases do not fit a night; the core can), the tour,
  choices C1-C8 that shape the night (and which research choices wait), the night's four-lane sketch, the director's
  steps before it, the questions ready to ask. STATE and the task point to it.
- 14:00 A fresh Opus reviewer of the night's sketch (feasibility, hidden person steps, dependencies, size):
  verifier#3 `M-2026-10-22-e61e5a00`, 40 minutes.
- 14:12 Review of the night's sketch: verifier#3 (opus) `M-2026-10-22-479c8b5e` FAIL: the proof needs a model login
  (a person's step); serial chains inside the batches; the size does not fit one night (its estimate: batches 1-2 in a
  night, the rest in a second session of 6-8 h); the accidental-host refusal needs a Cursor test first; ten minor.
  The walkthrough rewritten: two sessions, C9 (the proof's login), the chains marked, every minor folded in; STATE
  corrected. Round 2 sent (`M-2026-10-22-ac6f304a`, 20 minutes).
- 14:18 Round 2 of the sketch review: verifier#3 `M-2026-10-22-22f3ed12` PASS; its five small points folded in (the
  login wording; the contract section committed with the plan so no batch task holds a grant; the named-session
  precedence for the plan to state; a PowerShell form of the marker check; the live proof loads the extension with
  `-e`, so the setup wiring is proved by tests). The walkthrough is ready for the next session. No process left
  (helpers and probes; `ps` checked).
- 2026-10-23 06:30Z (a new session, orchestrator, opus; the director present) The walkthrough: the tour given in
  short; the choices asked one at a time. **C1:** the director: "1.0 first (Recommended)": rc.2's fixes, rc.2, the
  trial and 1.0 for Claude Code; the core's two sessions start 1.1.
  **C2:** "Translate (Recommended)": a Go reader per agent into the words the guard uses today; the rename later.
  **C3, reshaped by the director:** "as we decided earlier to finish 1.0 first, then based on this discussion, you can
  create kind of a contract plan which includes all the things pi needs to build, then i will give that plan to pi and
  build it there in parallel"; and: "i have a Servers/pi-harness repo, which i want to be used as my personalised pi
  configuration setup ... so i want things built in that". After the two parts were explained (the pi side: one
  extension; the Bonsai side: identity, the seam, protection and honesty), the director chose "pi side in pi-harness
  (Recommended)": the contract plan fixes the message format; pi builds the extension in pi-harness now, with a stopgap
  mode against today's binary; the Go side is built here after 1.0 by Claude agents with full proof. The night plan's
  two sessions now cover the Go side only.
  **C4:** "Wider tripwire (Recommended)": every known agent marker refuses person-only commands, in rc.2's fixes (T-5710); the parent-process check on Linux and WSL after 1.0 with identity.
  **C8:** the director ran the marker check in their terminals: "prints nothing, list is fine": the eight markers, CLAUDECODE left out.
  **C5 and C6, settled by the pi-harness choice (stated, not asked):** plain JavaScript (`.mjs`); its own small package
  in pi-harness, its entry last in pi's `packages`; the stopgap's limit said (the file is agent-writable until
  `check` verifies it after 1.0). **C7:** "Envelope only for now": `bonsai.hook/1` fixed now with its contract
  section (pi builds against it); `status.hosts` and `workspace.agents` wait for the Go work after 1.0.
  **C9:** "Your pi login (Recommended)": the after-1.0 proof runs pi with -ne -nc -ns --no-session -e, cheapest model; the read caveat said.
- 06:48Z The choices done. `T-5710` item 13 added (the markers, C4 and C8). New task `T-5798` (the contract plan for
  the pi side, its own run report from here). An auditor (opus) briefed for today's hook interface:
  `M-2026-10-23-43472a70`.
- 23 Oct, close: `T-5797` done; the work goes on in `T-5798` (approved, done).
