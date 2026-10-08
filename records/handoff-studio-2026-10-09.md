# Bonsai, for the studio's orchestrator: where it stands on 9 Oct 2026

Written by Bonsai's orchestrator for the studio's orchestrator, on Rohan's word, after Bonsai's walking skeleton
closed. It brings the studio up to date on Bonsai: what is done, what it changes for the studio and what it does not,
what the studio can use now, and what to read. Bonsai's own `STATE.md` stays the source of truth; where this file and
`STATE.md` differ, `STATE.md` wins.

## In short

- **Bonsai's walking skeleton is done** (8-9 Oct). Its risky pieces were built and measured on scratch copies, on WSL
  and on Windows: Bonsai writes and updates a project's files without overwriting a person's edits; a pack arrives as a
  Claude Code plugin at an exact commit, with two projects and a worktree beside one of them each loading their own;
  the guard blocks what it should, and also blocks when Bonsai is missing, crashes, is slow or is not on the PATH.
- **A fresh Opus verifier passed it** (all parts, the twelve checks, the four stop lines; tests re-run on both sides).
  The gate report is `~/Servers/Bonsai/records/gate-skeleton.md`.
- **At the gate (9 Oct) Rohan chose path (a), the full Bonsai 1.0.** The plan for step 5 is being written; nothing in
  step 5 is built before he approves it.
- **Bonsai has no screens of its own (Rohan, 9 Oct).** Bonsai is a pure command-line tool; every visual of Bonsai's
  data is the studio's (see "What it means for the studio", item 2).
- **For the studio, nothing changes yet.** By Rohan's 8 Oct decision ("Bonsai first"), the studio links to Bonsai at
  Bonsai's step 7, which comes after the gate and step 5's parts 5.1 to 5.5. Until the link, the Desk stays on upkeep
  and the studio writes no interim code toward the formats. The skeleton's end removes the first thing the link waits
  on; it does not start the link.

## What is done in Bonsai

| Piece | What it is | On `main` at |
|---|---|---|
| Part 0, the formats | `formats/`: a JSON Schema for each of the contract's ten formats, an example of each, 116 trick files with their format-0 and format-1 outcomes in `expect.json`, a raw-byte manifest | `3770d04` (set 3 at `3a1f195`) |
| Part 1, the clear-out | the old product's code removed (0.4.3 stays at its tag); CI on Linux and Windows, lint, govulncheck, CodeQL | `17f2938` |
| Part 2, the reader | Bonsai's format-1 reader in Go, with no YAML library, reaching all 116 format-1 outcomes; `bonsai status` | `3a1f195` |
| Part 3, the engine | `bonsai init`, `update`, `check`: staged writes, the lock last, a person's edited file never overwritten | `ee50971` |
| Part 4, packs as plugins | a public test pack (`LastStep/bonsai-test-pack`); each locked pack installed at project scope at its commit | `54b4fdd` |
| Part 5, the hook path | `bonsai hook guard` with one rule, failing closed under every fault | `33a6122` |
| Part 6, the gate report | every number spec §15 asks for, the twelve checks, the stop lines, the findings step 5 inherits | `5633721` |

The skeleton took 6 h 18 min of agent runs against 30-47 h estimated; about 18 minutes went on Windows-only failures.
Nothing in the studio's repo or Mimas was changed by this work (the last verifier checked both, read-only).

## What it means for the studio

1. **The link waits on Bonsai's step 5, parts 5.1 to 5.5** (formats and engine to 1.0, the recorder, the guards, the
   ladder runner, the packs). The spec costs those at 120-187 h; at the skeleton's measured pace that would be roughly
   16-39 agent hours, but the gate report's section 1 lists why that pace may not hold.
2. **Every visual of Bonsai's data is the studio's** (Rohan, 9 Oct: "this visual part of the job will be handled by the
   studio, while bonsai is a pure cli tool for overall simplicity"). The spec's plan for Bonsai's own web page (spec
   §11: one page per workspace, `bonsai serve`, a bundle the Desk would load) is dropped. The studio draws Bonsai's
   data itself, from Bonsai's JSON outputs: `bonsai status --json` (contract §12) and the formats in `formats/`.
   Update diffs are free text that contract §2.6 keeps off the VPS, so they are read in the terminal (`bonsai update
   --diff`) unless the studio decides otherwise. This moves work the spec had costed on Bonsai's side (30-50 h, a
   guess) to the studio's; when it is done is the studio's to plan.
3. **The formats are final and live in Bonsai's `formats/`.** The studio's readers test against those trick files when
   it links (that is the studio's half of check 7, still open). Do not copy them into the studio before the link: a copy
   pinned to a Bonsai commit is what the link brings.
4. **What moves at the link** is listed in spec §14, step 7's row, and contract §15.2 (the studio's adoption, costed):
   the studio's own guards, ladder and statusline retire then, and its hooks and CI run on `bonsai`. The studio's
   readers plan, cut on 8 Oct, stays the map of every studio reader for that link.
5. **Two studio problems are Bonsai's to fix, in step 5:** the redaction leaks the studio accepted are fixed by
   Bonsai's recorder (5.2), which replaces the Node redactor; the guard bug Rohan routed to Bonsai is designed out in
   Bonsai's guard (5.3), not patched in the studio.
6. **Rohan's 8 Oct requirement on hook lines** (they must not be redirectable by files an agent may edit, as `git
   rev-parse --git-common-dir` can be in the studio) is settled in Bonsai's step 5.3 (`design/plan.md`, "What still
   links Bonsai and the studio", item 5).
7. **Bonsai joins the Desk as its own project** at spec step 6, from step 5.4 on: the studio will be asked to register
   it then.
8. **Studio work that needs no Bonsai code** is the studio's to schedule: spec §13 item 4 (the memory move) is one;
   check the studio's own records for its status.

## What the skeleton learnt that studio sessions can use now

Measured on Claude Code 2.1.294, on both sides, 8-9 Oct (sources in the gate report, section 5):
- **`/agents` is gone** in 2.1.294. Read a plugin's role by starting a session as it: `claude --agent <plugin>:<role>`.
- **Local plugin scope leaks across worktrees**: a worktree reads the main checkout's `settings.local.json`, and
  `claude plugin install --scope local` in a worktree writes the main checkout's file. Project scope does not leak.
- **A project's marketplace is registered only by a session in a trusted folder**; until then `claude plugin install`
  reports "not found". A person's first session there does it; headless, a `-p` session given the project's own entry
  through `--settings` did it.
- **A first project-scope install rewrites `.claude/settings.json`** in Claude Code's own key order (on Windows, git
  showed the file modified afterwards).
- **`claude plugin eval` publishes its report to claude.ai by default**; pass `--no-publish` to keep it local, and
  `--trust-plugin` in unattended runs. A three-case eval on Haiku cost under 3 cents and half a minute.
- **Claude Code writes outside a session's folder**: transcripts, folder-trust entries, and Claude.ai's plugin sync
  rewriting `~/.claude/plugins/synced/<account>/.marketplaces.json`, even with `CLAUDE_CODE_PLUGIN_CACHE_DIR` set.
- **Hook start-up**: a Go hook through a shell line costs about 3 ms on WSL and about 66 ms under Git Bash on Windows,
  where Git Bash alone is about 35 ms. A hook line ending `|| exit 2` blocks when its program is missing or crashes.

## Working beside Bonsai

- Bonsai's sessions run in `~/Servers/Bonsai` and own it: work lands on its `main` through its orchestrator. Studio
  sessions read Bonsai's files; they do not edit, commit or push there.
- Never touch `~/ZenGarden/Bonsai` (an old clone with branches on no remote).
- `~/bonsai-checks` and `%USERPROFILE%\bonsai-checks` are Bonsai's scratch folders (test projects, launchers, builds).
  No session opens there except through Bonsai's own launchers.
- Rohan's roadmap artifact ("Trinetra Roadmap") is shared by both: Bonsai's sessions update Bonsai's cards only.

## What to read, in this order

1. `~/Servers/Bonsai/STATE.md`: where Bonsai stands now, and what waits on Rohan.
2. `~/Servers/Bonsai/records/gate-skeleton.md`: section 1 (plain words, the paths and their hours; Rohan chose (a))
   and section 5 (the findings step 5 inherits, by part).
3. `~/Servers/Bonsai/design/plan.md`: "What still links Bonsai and the studio".
4. `~/Servers/Bonsai/design/one-pager.md`: what Bonsai is for and what "correct" means.
5. `~/Servers/Bonsai/formats/README.md`: the formats and the trick files.
6. `~/Servers/Bonsai/design/bonsai-spec.md`, by section only (it is 130 KB): §13 (what the studio adopts first), §14
   (the order of work; step 7's row is the link), §15 (what the gate measured).
7. `~/Servers/Bonsai/design/contract.md`, by section only: §15.2 (the studio adopts the contract, costed).
