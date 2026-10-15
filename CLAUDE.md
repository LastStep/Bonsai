# Bonsai: guide for Claude Code

Bonsai is being rebuilt from scratch. 0.4.3 (May 2026) is the old product's last release, kept at its tag and in git
history; everything in this repo now serves the rebuild. Rohan is the director; Claude agents are the team.

**From the link** (step 5.4's switch: the installed `bonsai` links this repo), Bonsai guards its own checkout and proves
its own work. A rule below marked "from the link" holds once that commit is on `main`; until then the rules before it
hold. The block `bonsai init` writes into this file sits apart, between its markers.

## What Bonsai is, and is not

- **Is:** the structure inside each project. Formats (what a task, a run report, a log line and a question to the person
  look like); packs (roles, lanes and protocols, delivered as Claude Code plugins); guards that stop mistakes; a recorder
  that keeps a clean log; a ladder that proves work is done. One Go program, nothing else to install, on Linux and
  Windows.
- **Is not:** anything that applies an action. It never moves a task, applies an approval or starts an agent. No daemon,
  no network service, no terminal UI, and no screens or web page of its own (Rohan, 9 Oct: a pure command-line tool;
  every visual is the studio's). Registries, dashboards, notifications and deploys belong to its users.
- **Its first user** is Rohan's studio, Trinetra, which adopts the formats when it links (step 7). This repo names no studio file in its code.

## Where the truth lives

| Question | File |
|---|---|
| Where does Bonsai stand, and what is next? | `STATE.md`. Read it first, every session |
| What do we build, in what order, proved how? | `design/plan.md` |
| What is Bonsai, exactly? | `design/bonsai-spec.md` (130 KB). Grep its `## N.` sections; never read it whole |
| What do the formats say? | `design/contract.md` (100 KB). The same: by section |
| What did Rohan decide on each format? | `design/format-review.md` (confirmed 7 Oct: the formats are final) |
| What is it for; what does correct mean? | `design/one-pager.md` |
| What happened before? | `records/runs/`. Read one only when `STATE.md` or the plan points to it |
| The shared formats, schemas and trick files | `formats/` (from part 0) |
| What does Bonsai check here, and what may each task change? | `bonsai.yaml`; `records/tasks/` (from the link) |

The specs were written in the studio's repo: their "this repo" means the studio's, and the paths they cite are its.

## How a session works

- **Rohan opens a session in `~/Servers/Bonsai`** (the main checkout, on `main`). It is the **orchestrator**: it reads
  `STATE.md` and the plan, keeps its own context lean, and dispatches builder and verifier subagents. Bonsai has no role
  files, so every brief carries the job, the plan's part, the sections to read and the rules below that apply. It reads
  the agents' reports, not their file dumps.
- **Builders** work in a plain git worktree beside the clone (`git -C ~/Servers/Bonsai worktree add
  ~/Servers/Bonsai-<part> -b <part> main`), never the Agent tool's isolation worktrees. They commit; they never push.
  From the link, a worktree is made from `main` at or after the link commit (a branch from before it has no
  `bonsai.yaml`, so rung 0 cannot check it), and is removed only with `git worktree remove` (a recursive delete of a
  checkout's folder is a guarded delete).
- **Builders stay subagents of the orchestrator's session, unnamed** (Rohan, 12 Oct). A subagent shares its session's
  environment and cannot carry a task of its own: the guard grants by the one task reading `running`, the ladder is
  named by `--task`, and the stop gate runs neither for a subagent nor for the orchestrator. Their gate is the landing
  rule below, the lock on what reaches `main`. How builders start changes only with 5.5's agent pipelines.
- **A named session** (a builder started as its own `claude -p` through `claude-here` in its worktree, with
  `BONSAI_TASK=<id>` and `BONSAI_ROLE=builder`): its task and its run report are committed on `main` before its worktree
  is made, so the guard grants by its task and the stop gate holds it until its result is green at its HEAD.
- **Tasks, from the link.** One task per piece, `records/tasks/T-5xyy-<slug>.md` in `bonsai.task/1`: piece 5.x.y is
  `T-5x` and `y` in two digits (5.4.7 is T-5407, 5.5.12 T-5512); a part's planning and other work outside its pieces take
  T-5x90 to T-5x99 in order; a piece with a letter takes the next free number, named in its title. `lane` empty;
  `done_when` from the section's proof. **The orchestrator alone writes and moves task files**, in the main checkout:
  `todo` when a part's section is approved, `running` when the piece's builder starts, `verify` when its result is
  green, `done` once it has landed with CI green (or `cut`). The moves stay uncommitted until the records commit after
  the landing, with the run report, `STATE.md` and the tables from `bonsai check --write`. Builders never edit a task
  file. Every file in `records/tasks/` is a task: no template or note goes there (one that does not parse makes no task
  active, for every reader).
- **Tasks side by side.** Two tasks reading `running` make no task active, so the guard refuses a protected path to
  both. A task whose `bonsai.allows` is not empty runs with no other task `running`; the section's order says when.
- **Bonsai's four labels** (no definition is in force until base links in 5.5; the guard, rung 0, the runner and the
  stop gate read them by contract section 5.6's meaning, and a value of another kind grants or adds nothing):
  - `bonsai.allows`: the protected paths the task may change, as globs: those its section's "Owns" list names.
  - `bonsai.wants`: the person-only paths among them, copied into `bonsai.allows` under the stricter-only rule.
  - `bonsai.ladder`: rungs the task climbs beyond the floor; `[]` here, since the floor holds every rung.
  - `bonsai.branch`: the piece's branch.
- **A plan is reviewed by a fresh Opus agent** before it reaches Rohan.
- **A fresh verifier** (Opus, fresh context) for big or risky work only (Rohan's rule): part 0, the reader, guards and
  hooks, CI and release, security, and the end of each plan step; other parts close on green tests and CI plus the
  orchestrator's read of the diff, which the run report says. Related light parts share one verifier. A verifier
  reads the plan's part, the cited sections and the diff, re-runs the tests itself, and
  passes or fails the work. It fixes nothing. Nothing is done because an agent says so.
- **Models:** only the latest Opus, Sonnet, Haiku and Fable. Opus for verifiers, plan reviews, security, guards, release
  and decisions, and builders on risky code; Sonnet for sweeps, drafts and mechanical runs; Haiku for small bookkeeping
  and audit jobs (listing files, sorting or counting things, checking a list against a source, filling a run report's
  rows); Fable for visual design. Haiku reports facts; it judges nothing and its findings are read before they count.
  Every report says which model ran what.
- **Run reports** in `records/runs/R-<date>-<topic>.md`, opened before the first edit and appended as the work goes: a
  log, not a summary. Each lists every run with its model, start, end and minutes; those rows are the hours the stop
  lines count. From the link, one per task, named `R-<date>-<task id>-<topic>.md`, with `bonsai.run/1` frontmatter
  (`id` R-(date)-(task id), `task`, `role: orchestrator`, `model`, `started`, `finished`, `outcome`, `commits`,
  `labels`). Past run reports, logs and tasks are left as they are (Rohan, 12 Oct): no lock, edit or cleanup.
- **Proof** until the link: `go test ./...` and `go vet ./...` in WSL and natively on Windows, CI green on Linux and
  Windows, and a fresh verifier.
- **Proof from the link: the landing rule**, the lock on `main`. The orchestrator fast-forwards a piece only when
  `.bonsai/local/ladder/<task>.json` in the main checkout reads `mode: local`, green, `git.sha` the branch's HEAD,
  `git.dirty` false, every rung of the floor and of the task's `bonsai.ladder` green, and its `ladder` record in the
  log names `/usr/local/bin/bonsai` with the result's `sha256`; plus check 10's Windows half (`go test ./...` and
  `go vet ./...`, plain and tagged, natively on Windows) before the push, and CI after it. **Pieces ready together land
  as one batch** (Rohan, 15 Oct: "land in batches"): each rebased in turn onto the one before, each climbed at its own
  tip, check 10's Windows half run once at the batch's top commit, `main` fast-forwarded to it and pushed once; a piece
  in flight is not rebased and re-checked for each other landing, but joins the next batch. It also reads the branch's
  diff against the section's "Owns" list for the piece and refuses one that leaves it: `records/tasks/`, `design/` and
  `STATE.md` above all, free paths that rung 0 never judges. The ladder is always run by its installed path,
  `/usr/local/bin/bonsai ladder --task <id>`, so no other `bonsai` can make a proof. The run report quotes each landing's
  `git.sha`, the result's `sha256` and `green`, and the climb's counts and rung times in place of check 10's WSL output.
  Each part's end verifier matches every landed commit to a green local `ladder` record naming `/usr/local/bin/bonsai`
  with that `sha256`, and reads every task's `bonsai.allows` against its section's "Owns".
- **Agents first, from the link.** A builder climbs once its work is done and again after each fix, not after every
  edit: it runs the ladder with `--json`, reads the result, fixes what is red and climbs again, and asks a person only
  where a refusal's `next.who` says `person`.
- **`STATE.md`** is rewritten, never appended, whenever where Bonsai stands changes.
- **Claude Code's memory, from the link:** the link writes `autoMemoryEnabled: false` into `.claude/settings.json`; by
  Rohan's word of 13 Oct ("A. Keep them on"), his sitting adds `"autoMemoryEnabled": true` to his own local settings
  file (.claude/settings.local.json in the main checkout, out of git, read by no Bonsai check), so this repo's Claude
  Code memory notes keep loading until 5.6's personal layer takes them over. No agent copies the notes anywhere or
  edits that file (the guard refuses it).

## `bonsai.yaml`: stricter only, from the link

- `bonsai.yaml`, the lock and Claude Code's settings (and the release files) are person-only, and in Bonsai's mode a
  person-only path is grantable as written, so the rule is the orchestrator's: a task may hold one of them in
  `bonsai.allows` only for a change of a stricter kind (a floor raised to a green climb's count; a rung, a floor entry or
  a protected path added) or one a section Rohan approved names. **Any other change loosens the checks and waits for
  Rohan's word**, marked in the run report: a floor lowered or `ratchets` emptied; a rung's `command`, `required`,
  `tests`, `ratchet` or `capture` changed; a rung, floor entry, protected or person-only path taken out; a name added to
  `ci_marked_tests`; a release file changed beyond its section. Why: an agent loosening the checks it is held to would
  be judging its own work. Each part's end verifier reads every change to `bonsai.yaml` in the part against this rule.
- **A `bonsai.yaml` change reaches the ladder only once committed on `main`:** a climb reads the rungs, floor, ratchets
  and `ci_marked_tests` from main's committed copy, and refuses while the main checkout's file differs from it
  (`config-uncommitted`, a person's step: commit or revert it).
- **Loosening is refused in code until Rohan re-records it:** a climb whose committed settings loosen any check (a
  floor, a rung, a ratchet, a protected or person-only path, `ci_marked_tests`) against what this machine last accepted
  is refused (`config-looser`) until Rohan runs `bonsai update` in the main checkout; tightening is accepted and
  recorded automatically, so raising floors needs no person.
- **`main` stays on its branch:** the main checkout keeps the base branch this machine recorded at the link checked
  out; a climb (and the stop gate) refuses while it holds another branch or a detached HEAD.
- **The floors, once a part:** after the part's last records commit, the orchestrator climbs `main` once under the
  part's last task, raises the floors to that climb's counts under a task (`T-5x9N`) granting `bonsai.yaml`, records the
  Bless numbers in its run report and gives Rohan the floors in the part's last line, with nothing to answer.

## Commands that stay Rohan's, from the link

In Bonsai's repo an agent never passes `--allow-exec`, never runs `bonsai unlink`, and never runs an update that removes
or changes Bonsai's guard or stop line; the command itself refuses. Rohan types them in his own terminal, in
`~/Servers/Bonsai`: `bonsai update --allow-exec --yes` when a part changes one of Bonsai's own hook lines (the
orchestrator says when), and `bonsai unlink` only if he decides to take Bonsai out of its own repo. Every other
update of Bonsai's link (one that adds no code and leaves Bonsai's lines as they are, such as base's link in 5.5) is
the orchestrator's, in the main checkout only (rung 0 refuses a branch that changes `.claude/settings.json`), under a
task granting `bonsai.yaml` and the lock.

## The way back, if the installed `bonsai` blocks work

If Bonsai's guard breaks, every edit and command in a Claude Code session here is refused: it fails closed, on purpose.
Then Rohan, in a plain WSL terminal, quits Claude Code and starts it like this instead, for as many sessions as the fix
takes (the conversation is kept; that session runs with no guard, no stop gate, no log, and without his status line):

```bash
cd ~/Servers/Bonsai
claude --continue --settings '{"disableAllHooks": true}'
```

The orchestrator in it fixes the problem and lands the fix on the proof used before the link (the installed copy may be
the broken part): `go test ./...` and `go vet ./...` in WSL and natively on Windows, CI, and a fresh verifier. Then it
builds a new pre-release from a clean clone and sends Rohan its install lines and number; from the second pre-release on,
the one he had is kept in `~/bonsai-checks/prerelease/previous/`, and its install lines are the way back to it.

## Rules for every change

- **No `go install`, ever:** it puts a `bonsai` in front of the installed one on the PATH. Build with
  `go build -o <scratch folder>/bonsai ./cmd/bonsai`.
- Go: the module's `go 1.25` with its `toolchain` line (WSL's Go at `/usr/local/go/bin/go` downloads it). The standard
  library only (Windows calls through `syscall`); no terminal-UI library and no general YAML library.
- **Rules learnt from Windows** (the old code failed 29 tests only there), a review item in every change: every stored
  or printed path uses forward slashes; no test needs a symlink or a file mode to pass on Windows, and a test that cannot
  run on one OS says why; fingerprints read line endings as LF; output is byte-stable (no map-order iteration); Windows
  renames retry on busy errors; a hook line never calls `bash` by name.
- **Every template and pack file documents itself** (its purpose, when to use it, every field with an example), and a
  field change updates its docs in the same commit.
- **Every list has one home** (Bonsai's schemas or a pack's declarations); the reference page is regenerated with any
  list change.
- Every command runs unattended: `--json` everywhere but `hook`, no prompt without a terminal, ASCII human output, and
  every refusal names the next step.
- **`formats/` changes** only with its manifest.
- **Windows:** Windows Go cannot build from WSL's disk. Windows runs happen under `%USERPROFILE%\bonsai-checks\` with
  `"/mnt/c/Program Files/Go/bin/go.exe"`. Every Windows-side repo, worktree and CRLF checkout is made with Windows git by
  its full path, `"/mnt/c/Program Files/Git/cmd/git.exe"`. Never run Linux git in a Windows checkout.
- **Scratch links from a guarded session, from the link:** a stamped build's `init --allow-exec` of a scratch project,
  and any other `bonsai` command the guard refuses as a person's, runs through a script file in
  `~/bonsai-checks/scripts/`, never typed on the shell tool's line (the guard's word splitter does not read scripts; the
  stamped build's scratch root is the safety there).
- Commits: small, `area: what`, plain words, ending with the session's co-author line.

## GitHub (`LastStep/Bonsai`, public)

- Agents act on GitHub as `LastStep`, Rohan's admin account, so guards there stop accidents, not intent.
- **Builders never push.** The orchestrator fast-forwards `main` and pushes it only after the work's proof passes, then checks
  CI for that commit after every push:
  `gh api repos/LastStep/Bonsai/commits/<sha>/check-runs --jq '.check_runs[] | [.name, .status, .conclusion] | @tsv'`
  (or `gh run list -R LastStep/Bonsai -L 10`; this machine's gh has no `--branch`). Red CI is fixed forward.
- Work lands on `main` directly: no pull requests, no long-lived branches. Force pushes and deletion are blocked.
- **Nobody tags or releases.** `release.yml` stays disabled; releases are Rohan's word at spec step 5.7. No agent changes
  a setting, secret, ruleset or workflow switch, or merges or closes a pull request.

## Working with Rohan

- He is the director and writes, reviews and verifies no code. Agents own the code end to end: never leave a code
  problem for him to catch.
- Hand checks go to a Sonnet agent wherever an agent can run them; Rohan is asked only for what needs a person.
- Ask in plain words, context first, one question at a time, with two to four options and a recommendation. Never ask
  what a file answers.
- Hand him manual work (hand checks, settings, installs) in one batch with exact commands: bash for WSL, one command per
  line; PowerShell 5.1 for Windows, one command per line, no `&&`.
- A small fix stays small; ask before chasing edge cases.
- Write for him in plain words: what he can now do, with evidence (test counts, the CI run, the command to try). Never
  claim something is good for him; he judges that.

## Safety

- **No secrets anywhere: the repo is public.** No home-folder path, machine or tailnet name, email address, token or server
  address in any file or commit message; write `~/...` and `%USERPROFILE%`.
- Never touch `~/ZenGarden/Bonsai` (an old clone with branches on no remote).
- Tests use temp folders (`t.TempDir()`), never a real home (`~/.bonsai`, `~/.claude`) or a real project.
- Bonsai's work changes nothing outside this repo, its test pack, the workflow pack's repository
  (`LastStep/bonsai-workflow`: its local copy in `~/bonsai-checks/bonsai-workflow`, then GitHub once its privacy check
  P has passed) and the scratch folders (`~/bonsai-checks`, `%USERPROFILE%\bonsai-checks`), and, from the link,
  Bonsai's own home `~/.bonsai/` (its salt, this machine's record of the repo, the ladder's lock). Sessions and climbs
  in the real repo and its worktrees run with `BONSAI_HOME` unset: on a scratch home the guard finds no record of the
  real main checkout and grants nothing. Scratch projects keep their scratch homes.
- **Stop every process you start** (builds, tests, test sessions, watchers) and check with `ps -eo pid,etime,cmd` before
  you finish. Never kill what you did not start: another agent's run looks the same.

<!-- bonsai:block start: written by bonsai init and update from bonsai.yaml and each pack's bonsai/block.md (its docs); edit those, not this block. -->
This project is linked to Bonsai as workspace bonsai.
Project memory, an index of notes kept in records/memory: @records/memory/INDEX.md
<!-- bonsai:block end -->
