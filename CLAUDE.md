# Bonsai: guide for Claude Code

Bonsai is being rebuilt from scratch. 0.4.3 (May 2026) is the old product's last release, kept at its tag and in git
history; everything in this repo now serves the rebuild. A person directs it (below, "the director"); Claude agents are
the team.

**Bonsai guards its own checkout and proves its own work** (linked at 5.4's switch, 13 Oct; its packs `base` and
`workflow` taken at 5.5.7, 16 Oct). Rules below marked "from the link" hold now. The block `bonsai update` writes into
this file sits apart, between its markers: it holds the label definitions, the packs' skills and the roles.

## What Bonsai is, and is not

- **Is:** the structure inside each project. Formats (what a task, a run report, a log line and a question to the person
  look like); packs (roles, lanes and protocols, delivered as Claude Code plugins); guards that stop mistakes; a recorder
  that keeps a clean log; a ladder that proves work is done. One Go program, nothing else to install, on Linux and
  Windows.
- **Is not:** anything that applies an action. It never moves a task, applies an approval or starts an agent. No daemon,
  no network service, no terminal UI, and no screens or web page of its own (the director, 9 Oct: a pure command-line tool;
  every visual is the studio's). Registries, dashboards, notifications and deploys belong to its users.
- **Its first user** is the director's studio, Trinetra, which adopts the formats when it links (step 7). This repo names no studio file in its code.

## Where the truth lives

| Question | File |
|---|---|
| Where does Bonsai stand, and what is next? | `.bonsai/STATE.md`. Read it first, every session |
| What do we build, in what order, proved how? | `design/plan.md` |
| What is Bonsai, exactly? | `design/bonsai-spec.md` (130 KB). Grep its `## N.` sections; never read it whole |
| What do the formats say? | `design/contract.md` (100 KB). The same: by section |
| What did the director decide on each format? | `design/format-review.md` (confirmed 7 Oct: the formats are final) |
| What is it for; what does correct mean? | `design/one-pager.md` |
| What bugs are known, open and fixed? | `records/bugs.md`, the one register (skill `workflow:bugs`); STATE keeps no bug list |
| What happened before? | `records/runs/`. Read one only when `.bonsai/STATE.md` or the plan points to it |
| The shared formats, schemas and trick files | `formats/` (from part 0) |
| What does Bonsai check here, and what may each task change? | `bonsai.yaml`; `records/tasks/`; `bonsai status --json` |

The specs were written in the studio's repo: their "this repo" means the studio's, and the paths they cite are its.

## How a session works

The way of working is the `workflow` pack's, linked here since 5.5.7 (16 Oct): the six roles, who calls whom, briefs and
reports through `bonsai message`, the lanes and their pipelines, the lenses and the models (`workflow:orchestrate`,
`workflow:lanes`, `workflow:lenses`, `workflow:reporting`, `workflow:brief`); run reports and tasks by `base:run` and
`base:task`. Open those skills; what follows is Bonsai's own, where it says more than the pack.

- **The director opens a session in `~/Servers/Bonsai`** (the main checkout, on `main`). It is the **orchestrator**: it
  reads `.bonsai/STATE.md` and the plan, keeps its own context lean, dispatches builders and verifiers, and reads their
  reports, not their file dumps. A brief carries the job, the plan's part, the sections to read and the rules below that
  apply.
- **Agents are started by their role** (`workflow:builder`, `workflow:verifier`, `workflow:researcher`,
  `workflow:auditor`, `workflow:runner`): an agent of a type no pack declares (`general-purpose`, `Explore`) is held as
  a helper, writes nothing and has a helper's budget. A session started before the plugins loaded cannot start a role:
  quit and resume it (`claude --continue`).
- **Builders** are `workflow:builder` subagents of the orchestrator's session, unnamed (the director, 12 Oct), each in a
  plain git worktree beside the clone (`git -C ~/Servers/Bonsai worktree add ~/Servers/Bonsai-<part> -b <part> main`,
  from `main` at or after the link commit), never the Agent tool's isolation worktrees, removed only with `git worktree
  remove`. They commit; they never push; they never edit a task file. A subagent shares its session's environment and
  carries no task of its own: the guard grants by the one task reading `running`, the ladder is named by `--task`, and
  the landing rule below is their gate. **A named session** (`claude -p --agent workflow:builder` through `claude-here`
  in its worktree, with `BONSAI_TASK=<id>` and `BONSAI_ROLE=builder`) has its task and run report committed on `main`
  before its worktree is made, so the guard grants by its task and the stop gate holds it until its result is green.
- **Tasks.** One per piece, `records/tasks/T-5xyy-<slug>.md` (`base:task`): piece 5.x.y is `T-5x` and `y` in two digits
  (5.4.7 is T-5407, 5.5.12 T-5512); a part's planning and other work outside its pieces take T-5x90 to T-5x99 in order;
  a piece with a letter takes the next free number, named in its title. `lane` empty; `done_when` from the section's
  proof; `bonsai.allows` the protected paths the section's "Owns" list names, `bonsai.wants` the person-only ones among
  them (under the stricter-only rule), `bonsai.ladder` `[]` (the floor holds every rung), `bonsai.branch` the piece's
  branch or no line. **The orchestrator alone writes and moves task files**, in the main checkout: `todo` when a part's
  section is approved, `running` when the piece's builder starts, `verify` when its result is green, `done` once it has
  landed with CI green, or `cut`. The moves stay uncommitted until the records commit after the landing, with the run
  report, `.bonsai/STATE.md` and the tables from `bonsai check --write`. Every file in `records/tasks/` is a task. Two
  tasks reading `running` make none active; a task whose `bonsai.allows` is not empty runs with no other `running`.
- **Every plan is reviewed by a fresh Opus agent** before it reaches the director. **A fresh verifier** (Opus) for big or
  risky work only (the director's rule): part 0, the reader, guards and hooks, CI and release, security, and the end of
  each plan step; other parts close on green tests and CI plus the orchestrator's read of the diff, which the run report
  says. Related light parts share one verifier. Nothing is done because an agent says so.
- **Models:** only the latest Opus, Sonnet, Haiku and Fable. Opus for verifiers, plan reviews, security, guards, release
  and decisions, and builders on risky code; Sonnet for sweeps, drafts and mechanical runs; Haiku for small bookkeeping
  and audit jobs (listing files, sorting or counting things, checking a list against a source, filling a run report's
  rows), reporting facts and judging nothing; Fable for visual design. Every report says which model ran what.
- **Run reports** (`base:run`), one per task, `records/runs/R-<date>-<task id>-<topic>.md`, opened before the first edit
  and appended as the work goes: a log, not a summary, with a row for every run (model, start, end, minutes): those rows
  are the hours the stop lines count. Past run reports, logs and tasks are left as they are (the director, 12 Oct).
- **The landing rule**, the lock on `main`. The orchestrator fast-forwards a piece only when
  `.bonsai/local/ladder/<task>.json` in the main checkout reads `mode: local`, green, `git.sha` the branch's HEAD,
  `git.dirty` false, every rung of the floor and of the task's `bonsai.ladder` green, and its `ladder` record in the log
  names `/usr/local/bin/bonsai` with the result's `sha256`; then, **CI on the branch** (the director, 19 Oct): the
  orchestrator pushes the branch (builders never push), waits for CI on that exact commit to finish green, every job of
  it, the two windows jobs among them, and only then fast-forwards `main` to it, pushes `main` and deletes the branch on
  GitHub (`git push origin --delete <branch>`). Red CI on the branch stops the landing: the fix goes on as new commits
  on that branch, climbed again and pushed again, never force-pushed. Check 10's Windows half (`go test ./...` and
  `go vet ./...`, plain and tagged, natively on Windows) is the fallback for what CI cannot run: a landing while GitHub
  Actions cannot run the branch (an outage, or its minutes spent), and a proof that needs a checkout made with Windows
  git (check 3's CRLF checkout). **Pieces ready together land as one batch** (the director, 15 Oct): each rebased in
  turn onto the one before, each climbed at its own tip, the top branch pushed once and CI green on its top commit,
  `main` fast-forwarded to it and pushed once. It also reads the branch's diff against the section's "Owns" list and
  refuses one that leaves it: `records/tasks/`, `design/` and `.bonsai/STATE.md` above all, free paths that rung 0 never
  judges. The ladder is always run by its installed path, `/usr/local/bin/bonsai ladder --task <id>`. The run report
  quotes each landing's `git.sha`, the result's `sha256` and `green`, and the climb's counts and rung times. Each part's
  end verifier matches every landed commit to a green local `ladder` record naming `/usr/local/bin/bonsai` with that
  `sha256`, and reads every task's `bonsai.allows` against its section's "Owns".
- **Agents first.** A builder climbs once its work is done and again after each fix, not after every edit, with
  `--json`, and asks a person only where a refusal's `next.who` says `person`.
- **`.bonsai/STATE.md`** is rewritten, never appended, whenever where Bonsai stands changes (`base:state`); `bonsai
  check` runs before its records commit (a path it names that is gone is a finding).
- **Claude Code's memory:** the link writes `autoMemoryEnabled: false` into `.claude/settings.json`; by the director's word
  of 13 Oct ("A. Keep them on"), their own local settings file (settings.local.json in .claude, out of git) adds `"autoMemoryEnabled": true`
  until 5.6's personal layer takes the notes over. No agent copies the notes anywhere or edits that file.

## `bonsai.yaml`: stricter only, from the link

- `bonsai.yaml`, the lock and Claude Code's settings (and the release files) are person-only, and in Bonsai's mode a
  person-only path is grantable as written, so the rule is the orchestrator's: a task may hold one of them in
  `bonsai.allows` only for a change of a stricter kind (a floor raised to a green climb's count; a rung, a floor entry or
  a protected path added) or one a section the director approved names. **Any other change loosens the checks and waits
  for the director's word**, marked in the run report: a floor lowered or `ratchets` emptied; a rung's `command`, `required`,
  `tests`, `ratchet` or `capture` changed; a rung, floor entry, protected or person-only path taken out; a name added to
  `ci_marked_tests`; a release file changed beyond its section. Why: an agent loosening the checks it is held to would
  be judging its own work. Each part's end verifier reads every change to `bonsai.yaml` in the part against this rule.
- **A `bonsai.yaml` change reaches the ladder only once committed on `main`:** a climb reads the rungs, floor, ratchets
  and `ci_marked_tests` from main's committed copy, and refuses while the main checkout's file differs from it
  (`config-uncommitted`, a person's step: commit or revert it).
- **Loosening is refused in code until the director re-records it:** a climb whose committed settings loosen any check (a
  floor, a rung, a ratchet, a protected or person-only path, `ci_marked_tests`) against what this machine last accepted
  is refused (`config-looser`) until the director runs `bonsai update` in the main checkout; tightening is accepted and
  recorded automatically, so raising floors needs no person.
- **`main` stays on its branch:** the main checkout keeps the base branch this machine recorded at the link checked
  out; a climb (and the stop gate) refuses while it holds another branch or a detached HEAD.
- **The floors, once a part:** after the part's last records commit, the orchestrator climbs `main` once under the
  part's last task, raises the floors to that climb's counts under a task (`T-5x9N`) granting `bonsai.yaml`, records the
  Bless numbers in its run report and gives the director the floors in the part's last line, with nothing to answer.

## Commands that stay the director's, from the link

In Bonsai's repo an agent never passes `--allow-exec`, never runs `bonsai unlink`, and never runs an update that removes
or changes Bonsai's guard or stop line; the command itself refuses. The director types them in their own terminal, in
`~/Servers/Bonsai`: `bonsai update --allow-exec --yes` when a part changes one of Bonsai's own hook lines (the
orchestrator says when), and `bonsai unlink` only if they decide to take Bonsai out of its own repo. Every other
update of Bonsai's link (one that adds no code and leaves Bonsai's lines as they are, such as base's link in 5.5) is
the orchestrator's, in the main checkout only (rung 0 refuses a branch that changes `.claude/settings.json`), under a
task granting `bonsai.yaml` and the lock.

## The way back, if the installed `bonsai` blocks work

If Bonsai's guard breaks (the installed `bonsai` missing or failing to start), every edit, shell command and agent start
in a Claude Code session here is refused: Edit, Write, Bash, PowerShell and Monitor, and Agent, Task, SendMessage and
Workflow calls; it fails closed, on purpose. Reads, searches, skills and questions still go through (the second hook
line lets every other call pass while `bonsai` cannot run), so a session can still read and explain. Then the director, in a plain WSL terminal, quits Claude Code and starts it like this instead, for as many sessions as
the fix takes (the conversation is kept; that session runs with no guard, no stop gate, no log, and without their status
line):

```bash
cd ~/Servers/Bonsai
claude --continue --settings '{"disableAllHooks": true}'
```

The orchestrator in it fixes the problem and lands the fix on the proof used before the link (the installed copy may be
the broken part): `go test ./...` and `go vet ./...` in WSL and natively on Windows, CI, and a fresh verifier. Then it
builds a new pre-release from a clean clone and sends the director its install lines and number; from the second
pre-release on, the one they had is kept in `~/bonsai-checks/prerelease/previous/`, and its install lines are the way back to it.

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
- Every command runs unattended: `--json` everywhere but `hook`, no prompt without a terminal, ASCII human output,
  coloured only on a terminal (off under `NO_COLOR` and in pipes), and every refusal names the next step.
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

- Agents act on GitHub as `LastStep`, the director's admin account, so guards there stop accidents, not intent.
- **Builders never push.** The orchestrator pushes only to land work, by the landing rule: the batch's top branch
  first, then `main` once CI on that commit is green, and it deletes the branch on GitHub once `main` holds it. It
  checks CI for every commit it pushes:
  `gh api repos/LastStep/Bonsai/commits/<sha>/check-runs --jq '.check_runs[] | [.name, .status, .conclusion] | @tsv'`
  (or `gh run list -R LastStep/Bonsai -L 10`; this machine's gh has no `--branch`). Red CI on a branch stops its
  landing; red CI on `main` is fixed forward.
- Work lands on `main` directly: no pull requests, no long-lived branches (a pushed branch lives only until its
  landing). Force pushes and deletion are blocked on `main`; a pushed branch is never force-pushed.
- **Only the director tags and releases.** They alone push a tag (Bonsai's `v*`, the packs' `base-v*`, the workflow
  repo's `v*`) and approve a release run. `release.yml` runs only on their version tag, and its publishing only inside
  the `release` environment, once they approve; a release is published by that workflow, never by hand. No agent
  tags, approves, re-runs or cancels a release run, or changes a setting, secret, ruleset or workflow switch, or
  merges or closes a pull request. Until the director switches it on (5.7's batch 1), `release.yml` stays disabled
  on GitHub.

## Working with the director

- The director writes, reviews and verifies no code. Agents own the code end to end: never leave a code problem for
  them to catch.
- Hand checks go to a Sonnet agent wherever an agent can run them; the director is asked only for what needs a person.
- Ask in plain words, context first, one question at a time, with two to four options and a recommendation. Never ask
  what a file answers.
- Hand them manual work (hand checks, settings, installs) in one batch with exact commands: bash for WSL, one command
  per line; PowerShell 5.1 for Windows, one command per line, no `&&`.
- A small fix stays small; ask before chasing edge cases.
- Write for them in plain words: what they can now do, with evidence (test counts, the CI run, the command to try).
  Never claim something is good for them; they judge that.

## Safety

- **No secrets anywhere: the repo is public.** No home-folder path, machine or tailnet name, email address, token or server
  address in any file or commit message; write `~/...` and `%USERPROFILE%`.
- **No person's name** in any file, comment or commit message (the director, 15 Oct: Bonsai is a general product used
  by everyone): write "the director" in this project's own files and "the person" in the product (code, formats, docs,
  packs); a person's pronouns are "they". Files already under `records/` and `design/` keep what they hold.
- Never touch `~/ZenGarden/Bonsai` (an old clone with branches on no remote).
- Tests use temp folders (`t.TempDir()`), never a real home (`~/.bonsai`, `~/.claude`) or a real project.
- Bonsai's work changes nothing outside this repo, its test pack, the workflow pack's repository
  (`LastStep/bonsai-workflow`: its local copy in `~/bonsai-checks/bonsai-workflow`, then GitHub once its privacy check
  P has passed) and the scratch folders (`~/bonsai-checks`, `%USERPROFILE%\bonsai-checks`), and, from the link,
  Bonsai's own home `~/.bonsai/` (its salt, this machine's record of the repo, the ladder's lock), and Claude Code's
  plugin folder, where both packs are installed for this repo (`~/.claude/plugins/`). Sessions and climbs
  in the real repo and its worktrees run with `BONSAI_HOME` unset: on a scratch home the guard finds no record of the
  real main checkout and grants nothing. Scratch projects keep their scratch homes.
- **Stop every process you start** (builds, tests, test sessions, watchers) and check with `ps -eo pid,etime,cmd` before
  you finish. Never kill what you did not start: another agent's run looks the same.

<!-- bonsai:block start: written by bonsai init and update from bonsai.yaml and each pack's bonsai/block.md (its docs); edit those, not this block. -->
This project is linked to Bonsai as workspace bonsai; its packs: base 0.1.0, workflow 0.1.0.
Always-on protocols, read in full: @records/protocols/session-start.md
Project memory, an index of notes kept in records/memory: @records/memory/INDEX.md
Labels a document may carry under its labels: field (name: the value it takes; the documents it goes on; who sets it):
- bonsai.allows: a list of text; on task; agents set it; it grants a right. Protected paths this task may change, as globs. A path on the person-only list counts only if a person granted it.
- bonsai.wants: a list of text; on task; agents set it. Person-only paths this task asks for. A person's approval or grant copies them into bonsai.allows.
- bonsai.ladder: a list of number; on task; agents set it. The rungs this task climbs. It adds rungs to the workspace's floor and never drops one.
- bonsai.branch: text matching ^[A-Za-z0-9._/-]{1,100}$; on task; agents set it. The branch the task is built on; empty means the base branch.
- workflow.owner: one of orchestrator, builder, verifier, researcher, auditor, runner; on task; agents set it. The role that owns the task's work.
- workflow.model: one of opus, sonnet, haiku, fable; on task, run; agents set it. The Claude Code model the work runs on, or ran on (Claude Code only).
- workflow.feature: text matching ^F-[a-z0-9-]+$; on task, plan; agents set it. The one-pager (kind feature) this belongs to, by its id.
- workflow.milestone: text, at most 60 characters; on task, state, feature; agents set it. The milestone this counts toward, by the project's own name for it.
- workflow.spec: text matching ^SPEC-[0-9]{4}-[0-9]{2}-[0-9]{2}-[a-z0-9-]+$; on task; agents set it. The spec (kind spec) this task carries out, by its id.
- workflow.estimate_h: a number; on task; agents set it. The task's estimate in hours, written with the task.
Bonsai's documents have their templates in base's skills; open the one you need before writing or checking such a file: base:task, base:run (a run report), base:memory (a memory note and its index), base:message (a brief or a report, written with bonsai message), base:state (.bonsai/STATE.md), base:workspace (bonsai.yaml), base:ci (a CI workflow that runs bonsai check).
What Bonsai generates in this project, and how long it keeps each kind: base:generated-files.
Open base:operating-bonsai before you run any bonsai command, and whenever Bonsai or its guard refuses something.
Open base:operating-setup before you link or update a project, edit bonsai.yaml or change a pack's version.
Roles (workflow): the session the person talks to is the orchestrator and opens `workflow:orchestrate` when work starts; builder and verifier (leads) and researcher, auditor and runner (helpers) are called with their brief, or started as `claude --agent workflow:<role>`.
Lanes light, full and director, each with its pipeline: `workflow:lanes`. Document templates: `workflow:plan`, `workflow:feature`, `workflow:decision`, `workflow:options`, `workflow:spec`, `workflow:playtest`, `workflow:bugs`, and the daily brief's in `workflow:brief`.
<!-- bonsai:block end -->
