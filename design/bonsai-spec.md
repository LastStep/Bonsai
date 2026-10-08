# Bonsai: the design the rebuild is built from

> **The master copy, moved here from the studio repo on 8 Oct 2026 on Rohan's word; the studio keeps a pointer.** It
> was written in the studio's repo: where the text says "this repo" it means the studio's, and the paths it cites
> (`studio/...`, `docs/specs/...`, `tools/...`, `app/...`, `studio-app/...`) are the studio's, not Bonsai's. Notes marked
> **Changed 8 Oct (Rohan)** carry his decisions since; the rest is the text as he confirmed it. Bonsai's own plan, built
> from this spec, is `design/plan.md`; the formats contract is `design/contract.md` (cited below as contract §N).

**Spec, drafted 7 Oct 2026 by an Opus researcher, and revised the same day after a fresh Opus review**
(`docs/specs/2026-10-07-bonsai-review.md`: 4 blockers, 13 majors, 22 minors) and a recheck
(`docs/specs/2026-10-07-bonsai-recheck.md`: 0 blockers, 5 majors, 20 minors); both answered in §20. Rohan answered
its four questions the same evening; this text carries his answers (§18, §19), and the changes from his own review of
every format (`studio/decisions/REVIEW-2026-10-07-formats.md`, round 1) with its recheck
(`docs/specs/2026-10-07-bonsai-recheck-2.md`; §16, §20). It builds on the
settled contract (`docs/specs/2026-10-06-contract.md`, cited as contract §N) and reopens none of it. What binds:
Rohan's 4 Oct tool decisions (`studio/decisions/WALKTHROUGH-2026-10-04-studio-tools.md`), his 5 Oct answers
(`studio/decisions/WALKTHROUGH-2026-10-05-bonsai-and-studio.md` rows 1-24; `studio/decisions/VISION-2026-10-05-bonsai-and-trinetra.md`
Q1-Q16), his 7 Oct split rule and picks (`studio/decisions/CHECKLIST-2026-10-07-workspace-split.md`), his answers to the
contract's questions A-D, and his 7 Oct gate: he reviews every final format himself before anything is built on them.
Where a 5 Oct proposal and the 7 Oct rule disagree, the 7 Oct rule wins. Evidence: the 7 Oct workspace map research,
the 5 Oct audits, the phase 0 probes (`studio/decisions/PROBES-2026-10-05-phase-0.md`), the Bonsai clone (read only,
`~/Servers/Bonsai`, `main` at `c6a6757`), Claude Code's docs as published on 7 Oct, and measurements made for this spec
(§15).

**How to read it.** Rohan: §1 (about 5 minutes) and your steps in §17. The list of formats for your own review is §16;
your answers to the four questions are in §19. Builders and verifiers: all of it. Hours are agent-hour estimates unless
marked measured.

## 1. In plain words

**What Bonsai is now.** The structure inside each project: what a task, a run report, a log line and a question to you
look like; your way of working (roles, lanes, protocols) as packs; the guards that stop mistakes; the recorder that keeps
a clean log; and the ladder that proves work is done. One program in Go, nothing else to install. It never names the
studio, and it never moves a task, applies an approval or starts an agent: the studio does those (your 7 Oct rule).

**Packs are Claude Code plugins** (your 7 Oct pick). Claude Code fetches them and loads their roles and skills. Bonsai
writes into the project the few things a plugin cannot carry (deny rules, the walls round your key and token files among
them; always-on instructions) and keeps a lock, so
your edits are never overwritten. Bonsai keeps its guard out of plugins on purpose, so a missing plugin can never switch
it off. Two packs to start, both public: Bonsai's `base` and your `workflow` pack (your roles, lanes and protocols,
moved out of this repo), made from a pack template with its own CI and releases.

**Where Bonsai's files live in a project** (your round 1 answers). `bonsai.yaml` at the root, the one file you edit,
with a comment on every line. Beside it `.bonsai/`: the lock, STATE and two tables Bonsai builds (tasks; sessions and
hours), committed; and `.bonsai/local/`, never committed: the log, your questions and answers, ladder results. Your
machine's home (`~/.bonsai`) keeps only what belongs to the machine: a secret salt, this machine's settings, your
personal memory, the pack cache. Every template carries its own documentation.

**The order.**
1. You review every format yourself, one at a time (your 7 Oct gate; the list is §16). Nothing is built on any of them
   before you confirm them all: not the readers task, not slice 3, not Bonsai's skeleton.
2. The formats land in this repo (the readers task). The Desk pause lifts; slice 3 goes first.
3. Bonsai's walking skeleton, on scratch copies, never touching Mimas or this repo's tools, with written stop lines. It
   runs beside slice 3 once slice 3's plan is approved: a different repo and no shared ladder. The cost is two queues of
   questions for you at the same time; if you would rather keep slice 3 strictly first, say so and it waits.
4. Your gate, on measured numbers. Then the chosen path; the studio links to Bonsai; Mimas links last.

**The hours, re-costed for Go.** The skeleton is 30-47 h of AI time, its stop line 61 h (your answer D: the log and asks
wait for step 5.2), plus about 45 minutes of hand checks that are yours (§17); the old 25-35 h was costed for a smaller
skeleton with no hooks. After the gate, Bonsai 1.0 is about 139-218 h more (222-349 on the studio's record of estimates
growing 1.6 times), plus Bonsai's own screens, a guess of 30-50 h. Writing the guard, recorder and ladder in Go costs
roughly 40-85 h more than moving the Node code would have, part of it new work; your 7 Oct split and plugins take some of
that back.

**Go on Windows, measured today.** Windows already has Go 1.26. The old Bonsai builds there natively in 23 seconds, but 29
of its tests fail only on Windows (backslash paths, symlinks, file modes); the new code is written against those from
day one. Windows Go cannot build straight from WSL's disk. A hook in Go answers in about 1.5 ms in WSL (today's Node
guard: 26 ms) and about as fast as Node on Windows, where Git Bash's start-up dominates.

**Bonsai's repo.** 0.4.3 is already tagged and released. The 29 commits after it (an unreleased 0.5.0) will never ship as
the old product. The rebuild starts on a branch whose first commit clears out the old product; with one GitHub step of
yours (§17) the public release path is closed until step 5.7. The branch reaches the public `main` only at your gate,
so if you pause, nothing public changes.

> **Changed 8 Oct (Rohan):** no branch, no marker tag, no pull request: the clear-out and the rebuild land on `main`
> directly, so if you pause, `main` holds the work so far and 0.4.3 stays at its tag and in git history (§12's note).

**What you answered on 7 Oct** (§19). A: your 4 Oct guard answer holds; a later upgrade comes after the Jev trial. B:
no admin batch: no machine-wide settings file and no Claude Code version pin; the walls are deny rules in each linked
project, and `bonsai check` warns when Claude Code is too old. C: Bonsai's work is proven by Go tests, CI and a fresh
verifier until step 5.4, then by its own ladder, with a pre-release `bonsai` you install. D: the log and asks move to
step 5.2; the skeleton's stop line is 1.3 times its top (61 h). What is left for you: your second look at
the new and changed formats (round 2 of the review file; §16) and your steps (§17).

## 2. What Bonsai is, and is not

**Is:** formats, templates, packs, guards, the recorder and the ladder runner (checklist, confirmed 7 Oct). **Is not:**
anything that applies an action. No apply command, no task-status command, no label-writing command, no act log, no
approvals file (contract §3, §5.4, §5.5, §10). No daemon, no network service.

Rohan's checklist parts, and where each lands:

| Lands as | Parts |
|---|---|
| **The program** (§3-§9) | packs, install, updates and the lock, project config (`ws-packs`, `ws-install`, `ws-update`, `ws-config`); the ladder runner, stop gate, git-integrity check, CI (`ver-ladder`, `ver-stop`, `ver-git`, `ver-ci`); the path guard and the shell's delete check (question A), deny rules, redaction, the ledger rule (`g-path`, `g-shell`, `g-deny`, `g-redact`, `g-ledger`); the statusline's workspace half (`ws-statusline`); the boot-context caps (the block's 40 lines, the memory index's 120 lines and 12 KB) and doc freshness as `bonsai check` findings (`cost-budget`, `know-fresh`); how long each generated file is kept (§6); what loads at session start and after `/compact` (`mem-boot`, `mem-compact`) |
| **Formats** (contract, plus §5, §6, §10 here) | task, lanes, run reports, STATE, memory, plans, decisions (`plan-task`, `plan-lanes`, `plan-template`, `obs-runs`, `mem-state`, `mem-long`, `know-adr`, `know-index`, `know-research`) |
| **Pack content** (§5), no engine code | roles, the orchestrator, protocols, templates, models and effort per role, how often to verify, plan critique, phases, parallel work, the leftover-sweep rule, the fresh verifier, reward-hacking rules, tools per role, memory consolidation, handover (`roles-*`, `cost-models`, `cost-effort`, `cost-verify`, `plan-critique`, `exec-phases`, `exec-parallel`, `exec-sweep`, `ver-verifier`, `ver-reward`, `ver-prompt`, `g-roles`, `mem-consolidate`, `mem-handover`); game-only protocols as a later game pack (`roles-game`) |
| **Shared with the studio or Claude Code** | estimates and actuals (`plan-estimates`): the label `workflow.estimate_h` on tasks (set from the plan) and the studio's own `trinetra.actual_h` (measured from session records), which the Desk charts; typed asks (`loop-ask`): Bonsai's format (§8), the studio's reach; Claude Code setup per machine (`mach-cc`): `status --json` lists what a workspace needs, the oldest Claude Code it works with included, the studio's move checklist checks it; plugin output (`ws-plugin`, §5); managed settings (`ws-managed`): dropped by Rohan on 7 Oct (question B, §7); the sandbox and the secret walls (`g-sandbox`, `g-walls`, §7: the walls as base's deny rules); evals of a setup (`ver-evals`): a pack runs `claude plugin eval` as an ordinary command rung (§9) |
| **Later, in the parking file** (after 1.0, on Rohan's word) | routines and their runner (`rout-format`, `rout-house`, `rout-sec`); the dispatch guard (`g-dispatch`); lane coverage of gates (`ver-gates`; it needs `bonsai.lanes/2`, Rohan's call, contract §2.2); the assistant's runtime (`roles-assistant`, decision 3: its own design session) |
| **Moved to the studio on 7 Oct** | dispatch, status moves, approvals, grants (`exec-dispatch`, `plan-status`, `plan-approve`, `g-grants`); outside trackers as Studio connectors (`plan-external`) |

**Stays out of Bonsai:** the registry, the bridge, forwarding, the Desk, notifications, deploys, machine keys (the
studio's; one bridge per machine, keys only when a second real machine exists); outside services (Studio connectors:
"putting something like that in bonsai would end up bloating it a lot"); the Trinetra band (§5); what Rohan left blank
(cost and token counting, delivery metrics, transcript retention, Desk backups, spend per task, Fable in headless runs,
the old studio copies, Bonsai-Eval, the old 0.4.3 workspaces); what others ship (vision A.1: a marketplace, a rule
compiler, machine installers, a sandbox, a telemetry pipeline, session dashboards, a terminal UI). Later, on Rohan's
word: other agents and `AGENTS.md` (Q5), polish for strangers, layer overrides, OpenTelemetry as a second record.

## 3. One program

- **Go, one binary, nothing else to install** (walkthrough row 12, Q4 b). Linux and Windows on amd64 and arm64; macOS
  builds as today, untested. Standard library plus `golang.org/x/sys` (Windows job objects). **No terminal-UI library
  and no general YAML library:** format 1 has its own reader (contract §2.4), format 0 a hand port of `yaml.mjs`. The
  old binary takes 270-460 ms just to start in WSL, from its terminal-UI libraries (measured, §15); the new one must
  start in under 5 ms on Linux, because hooks call it on every tool call.
- **Go version:** the module says `go 1.25` with a `toolchain` line. WSL's Go 1.24.2 builds it through Go's own toolchain
  download (measured: 42 s the first time, nothing installed). Windows has Go 1.26.2.
- **Where it lives, so that every launch finds it and only Rohan replaces it.** WSL: `/usr/local/bin/bonsai`, owned by
  root. Windows: `C:\Program Files\Bonsai\bonsai.exe`, on the machine PATH, behind the UAC prompt. Both are Rohan's
  steps, once per Bonsai release on each side (vision A.1: machine pieces update on his word). A 5.x build is not a
  release: before 1.0 he installs one pre-release, at step 5.4, in WSL only (question C; §17 step 8; §18).
  `/usr/local/bin` is on the PATH of what starts sessions here: Rohan's terminals and the bridge's systemd user unit
  (read from the running bridge on 7 Oct). A background session runs with the PATH of the shell that dispatched it
  (agent-view page), so a builder that pickup starts finds `bonsai` too. Each install writes `<home>/install.json` (path, version, SHA-256); the
  5.4 pre-release has none until 5.6's installer, so until then `check` has no installed copy to compare with. No agent
  installs or replaces it: Bonsai's own `CLAUDE.md` and `Makefile` forbid `go install` (builders use `go build -o` into a
  scratch folder).
- **Old binaries shadow it today.** Four 0.4.3-era binaries are on the PATH: `~/go/bin/bonsai`, `~/go/bin/Bonsai` and
  `~/.local/bin/bonsai` in WSL, `%USERPROFILE%\go\bin\bonsai.exe` on Windows, and `~/go/bin` comes first. Removing them
  is Rohan's step, before step 5.4 (§17 step 3).
- **A tripwire, not a wall.** Root ownership protects the installed file, not the PATH. In Rohan's terminals `~/go/bin`,
  `~/node_modules/.bin` and `~/.local/bin` come before `/usr/local/bin` (measured 7 Oct), and all three are his to
  write, so one `go install` or one copy there puts a `bonsai` in front again. Sessions the bridge starts are safe: its
  PATH holds no user folders. What catches it: the walls deny copies into `~/go/bin` and `~/.local/bin` (§7; they do
  not stop `go install`); `bonsai hook start` logs its own path and SHA-256; `check` and `status` report when the
  `bonsai` found on the PATH is not the installed one, so the Desk can flag a changed guard.
- **Hooks call `bonsai` by name** (a committed file holds no absolute path); the studio's bridge calls it by the full
  path kept in its registration (vision A.9). With no machine-wide settings file (question B, dropped 7 Oct), no hook
  line calls it by full path, so a `bonsai` earlier on the PATH can stand in for it: the tripwire above is what catches
  that.
- **Output.** Human output is ASCII (PowerShell 5.1 garbles UTF-8); `--json` for programs. Exit codes (vision A.1): 0 ok,
  1 check findings, 2 bad input, 3 runtime, 4 wrong state or no `--yes`, 5 conflicts. `bonsai hook` exits 0 (allow) or
  2 (block); in a project with no `bonsai.yaml` it exits 0 at once and records nothing (§7).
- **Every command runs unattended** (Rohan, 7 Oct: "completely and easily useable by any ai agent"). No command waits
  for input without a terminal, or in an agent session (`CLAUDE_CODE_CHILD_SESSION` set) even with one: a command that
  needs a confirmation prints its preview and exits 4 without `--yes`, as today (§6); the y/N question is only for a
  person at a terminal. Every command but `hook` takes `--json` (§4); `hook` speaks Claude Code's hook format. Every
  refusal and error names the next thing to do, in the human line and in the JSON: the command to run, or the step
  that is a person's (the `error` object, §16). `bonsai --help` and `bonsai <word> --help` list every flag, exit code
  and an example, so an agent needs nothing else. Who may run `init` and `update` is unchanged (§6, §18).
- **Rules learnt from Windows** (the old code's 29 Windows-only failures), a review item for every Bonsai task: every
  stored or printed path uses forward slashes; no test needs a symlink or a file mode to pass on Windows, and a test
  that cannot run on one OS says why and CI lists it; fingerprints read line endings as LF; output is byte-stable (no
  map-order iteration); Windows renames retry on busy errors, as the spool does today; a hook line never calls `bash` by
  name (on this PC, `bash` on the Windows PATH is WSL's launcher, not Git Bash).

## 4. The commands

The contract's working names are fixed here (contract §2.1). Fourteen words; the cap for 1.0 is fifteen, one in and one
out after that (vision 2.6).

| Command | What it does | Writes |
|---|---|---|
| `bonsai init [--new-id] [--json]` | Links a project: writes `bonsai.yaml` (a comment on every line), `.bonsai/` (the lock, its `.gitignore`, the tables, and `STATE.md` from base's template when there is none, kind `once`), Bonsai's settings entries, the instruction block and the always-on protocol files (§5, §6); previews each settings line first, as `update` does; ends by saying where everything lives (§6). `--new-id` gives a copy its own id and empties its `.bonsai/local/` | Project files |
| `bonsai update [--diff] [--yes] [--allow-exec] [--keep P] [--adopt P] [--json]` | Brings packs to the refs in `bonsai.yaml`; previews by default, naming every file and every settings line with what it does (§6); all or nothing; exit 5 on a conflict; brings this machine's plugin install to the locked commit (§5) | Project files |
| `bonsai unlink [--yes] [--json]` | Removes what Bonsai wrote, leaving edited files, `.bonsai/STATE.md` and `.bonsai/local/` in place | Project files |
| `bonsai status [--json] [--full] [--active] [--line]` | One workspace at a glance (contract §12), the home, `.bonsai/` and the id among it. `--active`: only `active_task` (contract §13's read-only command). `--line`: the statusline's workspace half | Nothing |
| `bonsai check [--json] [--write] [--schema F] [--pack P]` | Findings (§6); exit 1 on any; warnings never change the exit code. CI and rung 0 run it. `--write`: rebuilds the two tables in `.bonsai/`, in the main checkout only; its exit code then says only whether it wrote (§6). `--schema F`: a format with every allowed value. `--pack P`: checks a pack folder (§5) | Nothing; with `--write`, the two tables |
| `bonsai hook <name>` | The hook entry point: `guard`, `stop`, `start`, `record` (§7, §8) | The log; cleaning (§6) |
| `bonsai ladder --task T [--root P] [--ci] [--json]` | The ladder runner (§9) | The ladder result (`--ci`: its own file), the log, a Bless ask |
| `bonsai ask … [--json]` / `bonsai ask --status <key> [--json]` | Files an ask; reads its answer (contract §9) | Asks |
| `bonsai answer <key> … [--json]` | Answers an ask (a terminal; or the studio with `--by act:<id> --via desk`) | Asks |
| `bonsai asks [--json]` | Lists asks | Nothing |
| `bonsai logs [--json] [--session S]` | Reads the log | Nothing |
| `bonsai log append --label … [--json]` | One outside event (contract §8.4) | The log |
| `bonsai settings show [--machine] [--json]` / `set k=v [--machine] [--json]` | This machine's settings for the workspace: `status_writes`, `status_command` (contract §3, §10.1); with `--machine`, the home's own: `cache_keep_days` (§6). How long project files are kept is `bonsai.yaml`'s, not a machine setting | The home |
| `bonsai labels attach <file> [--json]` / `detach <namespace> [--json]` | Machine-attached label definitions (contract §5.3) | The home |

Ten commands gained `--json` on 7 Oct (§3); no word was added. `bonsai walls --print` went with the machine-wide
settings file it printed for (question B, 7 Oct). Flags, not words: `bonsai --version`, `--help` on every
word, `bonsai check --schema <format>` (contract §2.2), `check --write` and `check --pack` (Rohan's format review; still
fourteen words, one free under the cap).
`bonsai serve` (§11) comes with the screens, on a cap Rohan raises.

**Refused in an agent session** (a tripwire, decision D): `settings set` and `labels attach` and `detach` refuse when
`CLAUDE_CODE_CHILD_SESSION` is set. Claude Code sets that variable only in processes it starts; `CLAUDECODE` is also
set in IDE terminals, so it would refuse Rohan's own commands there (env-vars page). The guard refuses them too
(contract §10.6). `init`, `update` and `unlink` write person-only paths (§6), so the guard judges them as such.

**Gone from 0.4.3:** `add`, `remove`, `list`, `catalog`, `guide`, `completion`, `validate` (now `check`), the terminal flow.

## 5. Packs: Claude Code plugins with a Bonsai manifest

**A pack is a git repo (or a folder in one) that is a valid Claude Code plugin, plus a `bonsai/` folder** for what a
plugin cannot carry. Claude Code loads only its own component folders, so `bonsai/` sits beside them (contract §5.2).
Claude Code first (Q5 a): the parts marked Claude Code only get a thin adapter per agent later, on Rohan's word.

```
<pack>/
  .claude-plugin/plugin.json   the plugin manifest: its name is the pack id; it never carries "version"
  agents/<role>.md             roles (Claude Code only); they load as <pack>:<role>
  skills/<name>/SKILL.md       protocols, how-tos and templates, to the open Agent Skills standard (portable)
  hooks/register.js            only in a pack that holds a mod (Claude Code only)
  bonsai/pack.yaml             bonsai.pack/1: id, version, needs, what the engine writes, each deny rule's why
  bonsai/labels.yaml           bonsai.labels/1 (contract §5.2)
  bonsai/lanes.yaml            bonsai.lanes/1 (contract §6), workflow pack only
  bonsai/block.md              the pack's part of the instruction block
  bonsai/files/                files the engine writes into projects (kinds pack and once), always-on protocols among them
  README.md                    what the pack is; documents plugin.json, which as JSON holds no comments
  .github/workflows/pack.yml   the pack's CI and release, from the pack template (below); a pack in Bonsai's repo uses Bonsai's CI
```

**Where roles live** (Rohan's 6.7 question). A role is one file, `agents/<role>.md`, inside a pack: Rohan's builder,
verifier, researcher, producer and playtest analyst in the `workflow` pack. Claude Code's plugin system installs the
pack on the machine (`bonsai update` asks it to, below) and loads each role under the pack's name: `workflow:builder`,
`workflow:verifier`. A linked project keeps no copies: this repo's `studio/roles/` and `.claude/agents/` go when the
studio links (§14 step 7), so a stale `builder` can never shadow the pack's. To change a role, change the pack and
release it; each project takes the release by its tag (below).

**What the plugin carries** (plugin manifest reference, 7 Oct): roles, skills, commands, hooks, mods, MCP servers. A
plugin cannot carry permission rules, environment or sandbox settings, or a loaded `CLAUDE.md`; plugin roles ignore
`permissionMode`, `hooks` and `mcpServers` (sub-agents page). This repo's roles use only `name`, `description`, `model`,
`tools`, so nothing is lost. **What the engine writes into the project:** deny rules; Bonsai's hook lines (§7); the
plugin wiring; `autoMemoryEnabled: false` (§10); `disableAllHooks: false` (§7); the instruction block in `CLAUDE.md`;
the always-on protocol files. **Bonsai's hooks never ride in a plugin:** plugins could carry them, but a plugin can be
missing, disabled or not yet fetched on a machine, and a guard must not depend on that.

**What loads always, and what on demand.** Today `CLAUDE.md` imports `studio/protocols/session-start.md` (4.6 KB), so
every session and subagent gets it; the other protocols are read when a role says so. Skills load only when Claude opens
one, unless a role preloads it through its `skills:` field (sub-agents page: "the full content of each listed skill is
injected into the subagent's context at startup"). So the workflow pack keeps today's behaviour:

| Protocol | Loads | Preloaded by |
|---|---|---|
| `session-start` | Always: a pack file the engine writes to the workspace's protocols folder, imported by the block, as today | everyone |
| `lanes` | Skill | builder, verifier, researcher, producer |
| `verification-ladder`, `reward-hacking-guards` | Skill | builder, verifier |
| `reporting` | Skill | builder, verifier, researcher |
| `adr` | Skill | researcher |
| `asset-safety`, `unity-live-editor` | Skill, in a later game pack | builder in a game project |

How a role names a plugin skill (with or without the `workflow:` prefix) is checked in step 5.5.

**Every template and pack file documents itself** (Rohan, 7 Oct: "any kind of template file that bonsai will have,
inside that file there should be proper official level documentation, so any person and ai agent can easily know the
use and purpose of it"; contract §2.8). It applies to every file Bonsai or a pack ships: the templates for a task, a
run report, STATE, a plan, a memory note and the rest (each a skill: `skills/<kind>/SKILL.md`), `bonsai.yaml`'s template,
`pack.yaml`, `labels.yaml`, `lanes.yaml`, `block.md`, the files in `bonsai/files/`, and Bonsai's generated files.
- **What the docs say:** the file's purpose, when to use it, and every field: its meaning, its allowed values (or where
  the list is defined, §6), and an example.
- **In the comment form the format allows:** in YAML a `#` comment for every key, at the end of its line or on the line
  just above (a long line), with a short header; in a template
  skill a fields table (`| Field | Meaning | Allowed values | Example |`) in the skill's body, above the template; in
  other markdown an HTML comment at the top; `plugin.json`, which is JSON, in the pack's `README.md`.
- **Tokens stay small where files are read often.** A file made from a template never copies the docs. Its `format:`
  line carries the pointer back as a trailing comment, about 25 tokens:
  `format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task`. A pack kind with
  no `format:` (a plan) carries one comment line naming its template skill. The docs load only when an agent opens the
  template skill or runs the command. What is loaded into every session carries one pointer line only: `block.md`'s
  docs stay in the pack and the block the engine writes into `CLAUDE.md` has none; an always-on protocol file in
  `bonsai/files/` keeps its docs to one line, with the rest in its pack's `README.md`.
- **`bonsai check --pack <dir>` keeps the docs in step with the fields** (in each pack's CI, and Bonsai's for `base`):
  every key in a pack's YAML files has its comment; every field in a template's frontmatter has a row in its fields
  table, and every row names a field the template holds; for a Bonsai format, those fields equal the format's schema;
  every allowed-values cell that names a closed list matches it (from the schema, or the pack's own declarations); and
  every deny rule in `pack.yaml` has its `why`. A template that drifts from its docs fails the pack's CI.
- **In a project, `check` does not require the comments:** a person may delete them from `bonsai.yaml`, and a task
  without its pointer is still a valid task.

**Pinning, fetching and drift.**
- `bonsai.yaml` names each pack by source and ref (a tag); the lock records the resolved 40-character commit. A tag that
  later resolves to another commit is refused (vision A.1).
- **A pack's `plugin.json` never carries `version`, nor does the marketplace entry the engine writes;** `bonsai check`
  refuses one in either. With a `version` there, Claude Code keeps every machine on its cached copy "however many
  commits" follow; without one, a `github` or `git-subdir` source's version is the commit (plugin loading reference,
  "How Claude Code computes the version"). So a pack is validated with `claude plugin validate --json`, not `--strict`
  (which fails on a missing `version`), and fails on every warning but that one (manifest reference).
- The engine writes an inline marketplace into `.claude/settings.json` (`extraKnownMarketplaces`, source `settings`)
  **named for the workspace and its locked commits**, `bonsai-<workspace name>-<8 hex>` (the 8 characters hash the
  lock's pack commits; this repo: `bonsai-trinetra-` and 8 characters), each plugin pinned to its locked commit (`sha`),
  and `enabledPlugins` for each (`workflow@bonsai-trinetra-…`). A user can register only one marketplace per name
  (marketplace reference), and a marketplace whose source changed is fetched again (plugin loading reference), so one
  name at two commits makes the two fight: two projects, two clones of one workspace, or, on the normal update path, a
  task worktree whose `update` moved a pack beside its main checkout still at the old commit (recheck minor 11). A name
  per workspace and commit keeps each checkout on its own; checkouts at the same commits share one. Old names stay
  registered on the machine (unchecked: whether they cost more than disk). Roles still load as `workflow:builder`:
  components are namespaced by the plugin's manifest name.
- **This machine's install follows the lock.** Claude Code does not fetch a plugin that only the project's settings
  enable (plugin loading reference), and `claude plugin install` is a no-op, exit 0, once installed (plugin commands
  reference). So `bonsai update` runs `claude plugin install` under the lock's marketplace name, which is new whenever a
  locked commit changed. The docs' other route, an enable in the untracked `.claude/settings.local.json` that Claude Code
  then fetches itself, is the fallback if the skeleton finds the first unreliable.
- **Drift is reported, never silent:** `check` and `status --full` compare the version that `claude plugin list --json`
  reports (the commit's first 12 characters; for a `git-subdir` source, plus a hash of the folder) with the lock, and
  report a difference. A pack not installed on this
  machine is a `needs` entry (kind `plugin`) in `status --json`, which the studio's move checklist reads.
- Bonsai fetches each pack itself too, into `<home>/cache/`, to read `bonsai/`: once per commit, with the machine's own
  git (Rohan's packs are public and need no login; someone else's private pack would use the machine's own git login;
  no token is stored). It never reads Claude Code's plugin cache.
- **CI needs no pack:** at each update the engine copies what checks need into the lock (`declares`, §6), so `bonsai
  check` in CI runs offline, fetches nothing, and does not break when a pack's repo is unreachable.
- **Trust.** Project `extraKnownMarketplaces` apply only after Claude Code's workspace trust, and a `--bg` session in an
  untrusted folder exits "Workspace not trusted". Project marketplace entries in an untrusted folder are ignored with no
  message, `-p` runs included (settings reference), so `bonsai update`'s `claude plugin install` there may do nothing.
  A fresh plain worktree is a new folder. The skeleton checks whether its packs load and update without a prompt; if
  not, slice 3's pickup must trust each new worktree first (a note for T-0018).

**The two packs.**

| Pack | Where | Holds |
|---|---|---|
| `base` (public) | Bonsai's repo, `packs/base/`, tagged `base-vX.Y.Z` (no leading `v`, so no release fires) | The `bonsai.*` label definitions (contract §5.6); templates for task, run report, STATE, memory note and `bonsai.yaml`, as skills, each documented (above); the generated-files page as a skill (§6); a CI workflow template (kind `once`); the baseline deny rules, the walls round key, token and credential files among them (§7). No roles, no lanes |
| `workflow` (public, Rohan, 7 Oct: "why does anything have to be private here?") | Its own repo, `LastStep/bonsai-workflow` (vision §6's proposal, taken), made from the pack template | Rohan's roles, lanes, protocols, document kinds (plans, one-pagers, decisions, specs, playtests, briefs, bugs) with their documented templates, and `workflow.*` labels (contract §5.4, plus `workflow.estimate_h`), and the deny rules for the studio's own secret files (§7), moved from this repo's `studio/roles`, `studio/protocols`, `studio/templates`. Nothing in it is secret: deny rules name files, never their contents. Its id is `workflow`, so the contract's labels keep their names |

**The pack template, with CI and releases built in** (Rohan, 7 Oct: "i want proper ci/cd for it as well, so the
process of adding to the marketplace and updating is streamlined"). A folder in Bonsai's repo, `packs/template/`: a
minimal valid pack (one role, one skill, one documented template, `pack.yaml`, `labels.yaml`, `README.md`) and its
`.github/workflows/pack.yml`. A new pack starts as a copy of that folder in its own repo; `workflow` is the first, at
step 5.5. Bonsai's own CI runs the template's checks on every Bonsai commit, so the template never drifts from the
engine. A GitHub template repository can be cut from it later, if people outside the studio make packs (Rohan's word).
The pack's CI, on every push and pull request:
1. **Validate as a Claude Code plugin:** `claude plugin validate --json`; every warning but the missing `version` fails
   (no `--strict`, which fails on that one). Whether `validate` needs a Claude login on a CI runner is unchecked; step
   5.5 tries it.
2. **`bonsai check --pack .`** on its declarations: `pack.yaml`, labels, lanes, the block's 40 lines, document kinds,
   no `version` in `plugin.json`, and the template docs in step with their fields (above). The pinned `bonsai` comes
   from a Bonsai release archive checked by its SHA-256 (§12 step 8); before Bonsai's first release, built from a
   pinned Bonsai commit.
3. **Its tests:** `bonsai init` and `bonsai check` in a scratch project linked to the pack at this commit, plus the
   pack's own tests and evals (`claude plugin eval`) if it has any. Evals call a model, so they need a key: a
   repository secret, never given to pull requests from forks, or run by hand on the machine.
4. **Release on a git tag** `vX.Y.Z`: all of the above again; the tag must equal `pack.yaml`'s `version`; then a GitHub
   release with its notes. Nothing is uploaded: a pack is fetched by commit.

**Adopting a release stays a person's step:** change the tag in `bonsai.yaml`, run `bonsai update` (its preview says
what changes). **There is no central marketplace to publish to:** Claude Code's marketplaces are git sources, and the
engine writes each workspace's own marketplace entry (above), so a tagged commit on GitHub is the whole release.

**Erratum for contract §10.3, for T-0018:** plugin components are namespaced (plugin manifest reference: an agent
`reviewer` in plugin `deploy-tools` "appears as `deploy-tools:reviewer`"), so pickup names `workflow:builder`, as in
`claude --agent workflow:builder --bg`. No docs page shows that exact command, so skeleton check 8 confirms it starts.
The old project copies in `.claude/agents/` are removed when the studio links, so a stale `builder` cannot shadow it.

**Mods.** A mod is a plugin whose `hooks/register.js` registers handlers (mods overview). A pack may carry one. **The
first mod, the Trinetra band, is the studio's own plugin, not a Bonsai pack:** it shows the active task, the lane, the
last ladder and the asks waiting for Rohan, reading `bonsai status --json`, and it names the studio, which Bonsai never
does. It does not show backups (Rohan left Desk backups blank: no action). It is a studio task after slice 3, about 4-8
h, outside Bonsai's hours. A mod fails open and can approve a call that a project hook blocked (§7): never a wall.

## 6. The engine: `bonsai.yaml`, the lock, updates

**`bonsai.yaml`** (`bonsai.workspace/1`, format-1 YAML, contract §2.4), committed, person-only: intent and project
values, the Bonsai half of today's `studio/game.yaml` (vision A.6 row 11). It stays at the root: the one Bonsai file a
person edits. The studio's display facts stay in its registration. **It is written with a comment on every line**
(Rohan, 7 Oct, format review 1.4: "document it by comments properly, so its easy to understand while reading"): `init`
writes it from base's documented template (§5). This repo's, after the link (illustrative):

```yaml
# bonsai.yaml: this project's Bonsai settings (format bonsai.workspace/1). Only a person changes this file.
# Every field and its allowed values: bonsai check --schema bonsai.workspace
format: bonsai.workspace/1          # the format and its version; a newer one is refused, never guessed
id: ws-<26 base32>                  # this project's id, written once by bonsai init; a copy gets its own (init --new-id)
name: trinetra                      # a short name: lower-case letters, digits and dashes
packs:                              # the packs this project uses, applied in this order
  - id: base                        # Bonsai's own pack: core labels, templates, the walls round secret files
    source: "https://github.com/LastStep/Bonsai.git"            # the git repo the pack comes from
    path: packs/base                # the pack's folder inside that repo
    ref: base-v1.0.0                # the release tag; change it and run bonsai update to take a new release
  - id: workflow                    # your roles, lanes, protocols and templates
    source: "https://github.com/LastStep/bonsai-workflow.git"   # its own public repo
    ref: v1.0.0                     # its release tag
documents:                          # where this project keeps the documents Bonsai knows
  task: studio/tasks                # one file per task
  run: studio/runs                  # one run report per agent session
  answers: studio/answers.md        # answers to questions that name no file
  memory: studio/memory             # memory notes and their INDEX.md
  protocols: studio/protocols       # the always-on protocol files the instruction block imports
# protected: paths an agent changes only while its running task lists them in bonsai.allows
protected: [".claude/**", "bonsai.yaml", ".bonsai/lock.json", "studio/protocols/**", ".github/**", "docs/PLAN.md", "docs/research/**"]
# person_only: of those, the paths only a person grants (an approve or grant tap)
person_only: [".claude/**", "bonsai.yaml", ".bonsai/lock.json", "studio/protocols/**", ".github/**"]
never_edit: ["studio/ledger.json"]  # paths no agent ever changes; written as deny rules
ladder_floor: [0, 1, 2, 3, 4]       # the rungs every task climbs, whatever the task lists
ladder: []                          # the rungs: what each runs and what it proves (left out here)
ratchets: {}                        # counts that may only rise; they rise when a person taps Bless
ci_marked_tests: []                 # tests allowed to skip in CI only, each by name
generated:                          # how long generated files are kept, per kind (skill base:generated-files)
  log:                              # the log, in .bonsai/local/log/
    keep_days: 30                   # a log file goes 30 days after its last line, never before its rows are in
  ladder:                           # ladder results, in .bonsai/local/ladder/
    keep_days: 7                    # the result of a task not done or cut is never cleaned
```

(The other kinds and their defaults: the table under "Generated files" below.)

`person_only` is contract §18 B (c)'s list carried forward (the guards' code leaves the repo when the studio links; the
installed binary is protected by its owner, §3); Rohan confirmed it on 7 Oct. `ladder` keeps today's rung shape
(`rung`, `name`, `kind`, `command`, `required`, `timeout_s`, `ratchet`, `capture`, `means`) plus two optional fields,
`tests` and `base_setup` (§9). There is no `state` entry: STATE's place is fixed (below). **No always-on budget:** Rohan
dropped `boot_budget_kb` (7 Oct: "Don't use boot_budget_kb"); the fixed caps stay: the instruction block's 40 lines,
the memory index's 120 lines and 12 KB, a memory note's 4 KB (§10). Writers keep the comments (contract §2.4); `check`
does not require them in a project.

**`.bonsai/`, Bonsai's folder in each project** (Rohan, 7 Oct, round 1; the tree is contract §3's): `bonsai.yaml` stays
at the root; everything else Bonsai keeps in a project sits in `.bonsai/`: `.gitignore` (written by Bonsai: `local/`),
`lock.json`, `STATE.md` (agents rewrite it) and the two tables, all committed, and `local/`, never committed. Only the
lock is person-only, not the folder. (A 0.4.3 workspace also has a `.bonsai/`, holding `catalog.json`; `init` refuses
such a workspace on its `.bonsai.yaml`, so the two never mix.)

**The two tables** (Rohan, 7 Oct, format review 3.2 and 5.3: "files where we aggregate some kind of data ... should
all live in one specified folder decided by bonsai"; formats `bonsai.tasks/1` and `bonsai.sessions/1`, contract §7.5).
The task files and the log stay the truth; Bonsai rebuilds the tables from them; **a table never grants anything:** the
guard, rung 0, the stop gate and the active-task rule read the task files (contract §13). `.bonsai/tasks.md`
(illustrative; the title paraphrased):

```markdown
---
format: bonsai.tasks/1   # generated by bonsai check --write from studio/tasks/; never edit by hand
---
Active task when none is named: none (no task reads running)

| Task | Title | Status | Lane | Started | Finished |
|---|---|---|---|---|---|
| T-0072 | The bridge serves Windows-side projects | done | full | 2026-10-07 | 2026-10-07 |
```

`.bonsai/sessions.md` has one row per ended session and per ended subagent run (a builder or verifier that an
orchestrator starts with the Agent tool runs as a subagent inside its session): the session id's first 8 characters,
task, role, model, start, end, minutes; then hours per task and role (a subagent run lies inside its session's row, so
the two are shown apart, never added). **A row's task, one rule:** the task named by `--task` or the environment, if
set; else the active task (contract §13) of the checkout or worktree it ran in; else `none`; as the recorder found it
when the session or subagent run started (contract §7.5, §8.1). Its role is `BONSAI_ROLE` for a session and the
`subagent_type` for a subagent run. It replaces the run report's "Sessions" table (§16 row 28): no agent writes rows.
Its rows are only added, so hours outlive the log files they came from.
- **Who rebuilds them, and when:** `bonsai check --write` (a flag, no new word), which rebuilds the two tables and
  writes nothing else; it runs only in the main checkout and refuses in a worktree (exit 4, naming the main checkout).
  With `--write` the exit code says only whether it wrote: 0 written, 3 it could not; findings are still listed and
  change nothing. It moves no task and commits nothing, so it stays on Bonsai's side of the 7 Oct split. **In a
  managed workspace the studio runs it** in its write path just before each move's commit and commits the tables with
  the move (contract §10.5): every move happens on main, so `main`'s tables are current at every move, and a session's
  row arrives with the next move at the latest (every task ends with one).
  In an unmanaged workspace, whoever commits on the main checkout runs it first; the workflow pack's closing protocol
  says so.
- **Committed, and never on a branch** (decided here, §18). Rohan reads them on GitHub and the Desk; merge conflicts
  are avoided because only the main checkout writes them: rung 0 refuses a task branch that changes a table, and the
  guard refuses an agent's file-tool edit of one (a tripwire). Two machines appending rows to one project's sessions
  table would conflict; `check --write` keeps both sides' rows (sorted by start), and that waits for a second real
  machine.
- **Stale is a warning, everywhere** (the orchestrator, 7 Oct): a tasks table that differs from a rebuild, or a
  sessions table that lacks a row for an ended session or subagent run in the log, is a `check` warning in the main
  checkout, a worktree and CI alike. It never changes an exit code, never turns a rung red and is never one of
  `status --json`'s `problems`: the tables lag between moves by design, grant nothing, and the next `check --write`
  rebuilds them.

**`.bonsai/local/`** (Rohan, 7 Oct, round 1; contract §3): the log, the asks and the ladder results, inside the project
in the main checkout, never committed; worktrees reach it through `git rev-parse --git-common-dir`. He asked why they
lived outside the project: a moved checkout, a container or a cloud machine would leave them behind. `check` and rung 0
refuse any file from `local/` that is tracked or staged; every writer of `local/` restores a missing
`.bonsai/.gitignore`, and `check` finds one missing or changed; only Bonsai writes there (a deny rule and the guard,
§7). The outbox went with the move (§8). A `git clean -x` (or `-X`) in the main checkout deletes `local/` (lines not yet
forwarded become a gap, contract §8.5), and `git stash --all` moves it away under a running recorder (`-u` leaves
ignored files alone): the delete check refuses both as bulk deletes (§7).

**Generated files: how long each is kept** (Rohan, 7 Oct, format review 5.1: "a fresh document in bonsai which tells
how to manage all these ai generated files (runs, logs, ladder results etc), with options of cleaning X days old,
rotating and keeping fresh Y records etc."). One page, base's `generated-files` skill (`packs/base/skills/
generated-files/SKILL.md`, read on demand, never always on), says per kind what it is, who writes it, where it lives,
its rule and its default; its table of kinds and defaults is generated from Bonsai's code and checked in CI, like the
reference page below. Each kind takes `keep_days` (clean what is older), `keep_newest` (keep only the newest Y), or
both (either one cleans); the protections always win.

| Kind | Where | Default | Never cleaned |
|---|---|---|---|
| `log` | `.bonsai/local/log/` | 30 days after a file's last line (as before) | an open session's file; a file holding an ended session or subagent run with no row yet in `.bonsai/sessions.md` |
| `asks` | `.bonsai/local/asks/` | kept (as before: "never pruned") | a day file holding an open ask |
| `ladder` | `.bonsai/local/ladder/` | 7 days after the result's `finished` (Rohan, 5.1) | the result of a task not `done` or `cut` |
| `run` | the run reports' folder | kept: committed history | any, by Bonsai: past a rule, `check` lists them and a person deletes them (a bulk delete is his) |
| `sessions` | rows of `.bonsai/sessions.md` | kept | rows of an open task |
| `tasks` | `.bonsai/tasks.md` | a rebuild: nothing to clean | — |

- **Where the settings live:** every project file, `local/` included, under `generated:` in `bonsai.yaml` (committed,
  person-only, so only a person changes how long records are kept); the home's files in the machine settings
  (`bonsai settings set --machine`). The home now holds no AI-made record; its one growing folder is the pack cache:
  `cache_keep_days`, default none (deleting `cache/` by hand is always safe; it refills). `log_keep_days` left the
  machine settings.
- **When:** at the end of each session (the recorder's `session_end`, async, never blocking), after each ladder run,
  and in `check --write` (sessions rows); the cache in `update`.
- **Rows before files:** a log file waiting for its rows is kept; the cleaner never writes a table, only
  `check --write` does. In a project with no studio, nothing is lost: the files wait until someone on the main checkout
  runs `check --write` (the workflow pack's closing protocol says to, and `check` warns of the stale table).
- **Visible:** every file or row cleaned is one `clean` record in the log (`target` its project-relative path, `reason`
  the rule, contract §8.2), so `bonsai logs`, the Desk and the forwarder's gap report all show it.

**Said in plain words** (Rohan, 7 Oct, format review 1.2: "we need to be clear to the user at the start"). `init` ends
with this, and `status` shows the same (`status --json`: `workspace`, `home`, `local`, contract §12):

```text
Linked trinetra. Workspace id: ws-<26 base32> (every clone and worktree of this project shares it).
In this project:
  bonsai.yaml      this project's Bonsai settings. You edit it; it is committed.
  .bonsai/         the lock, STATE and two tables Bonsai rebuilds (tasks; sessions and hours). Committed.
  .bonsai/local/   the log, questions for you and their answers, ladder results. Never committed.
On this machine:
  ~/.bonsai             Bonsai's home: a secret salt, this machine's settings for each project, label files
                        the studio attached, your personal memory, the pack cache. None of it enters git.
A copy meant as a new project needs its own id: bonsai init --new-id
```

**Is this how established tools do it** (his question on 1.2). Git keeps a repo's own data in `.git/` inside the
project, never committed, as `.bonsai/local/` now does; Claude Code keeps each project's data in the home under a name
made from the folder's path (`~/.claude/projects/-home-rohan-Servers-Trinetra-Game-Studio/`), as Bonsai's machine
folder does. The drawbacks, a moved checkout stranding its machine folder and a copied repo copying the id, and what
Bonsai does about each: contract §3 ("Moved checkouts", "Copies"); `check` and `status` warn of a stranded folder and
name the step.

**The lock**, `.bonsai/lock.json` (Rohan, 7 Oct, format review 6.2; was `bonsai.lock.json` at the root), is contract
§14's `bonsai.lock/1`, plus one additive field per pack, `declares`: the lanes, document kinds, label definitions and
protected paths that pack declared at the locked commit. That is what lets `bonsai check` run in CI with no pack fetch.

**File kinds** (contract §14): `pack` (Bonsai's; an edit is a conflict), `once` (written once, then the project's),
`block` (a marked region in a project-owned file, such as `CLAUDE.md`), `keys` (Bonsai's own entries in a project-owned
JSON file, such as `.claude/settings.json`), `kept` (a pack file the person kept edited).

**The instruction block** in `CLAUDE.md` holds, in at most 40 lines: one line naming the workspace and its packs; the
import of the always-on protocol files; the import of the project's memory index (§10); the label definitions agents
see (contract §5.3, about 40 tokens each). Its start marker carries the one pointer line to the pack's `block.md`
docs (§5). The caps are fixed: the block's 40 lines and the memory index's 120 lines and 12 KB (§10); there is no
budget for the always-on total (Rohan dropped it, 7 Oct, above). Today's always-on text is about 21 KB.

**How `update` decides, per file**, from three fingerprints (what the lock says Bonsai last wrote, what is on disk, what
the pack now gives), after Debian's conffiles rule (vision A.1):

| On disk vs locked | New vs locked | Result |
|---|---|---|
| same | same | `unchanged` |
| same | different | `updated` |
| different | same | left alone; `check` reports it `changed` |
| different, and equal to new | different | `adopted` (the lock catches up) |
| different | different | **conflict**: exit 5, nothing written, the file named |
| missing | any | kind `pack`: written again; kind `once`: left missing |

`--keep P` applies the rest and marks P `kept`; a later pack change to P is a conflict again. `--adopt P` takes the
pack's copy and puts the project's copy in `<home>/cache/`, never in the repo. **All or nothing:** every write is staged
first; renames follow; the lock is written last, so a crash between renames is finished by running `update` again. **Code
is consented to separately:** a change to a hook line or to a file a hook runs is listed under "runs code" and needs
`--allow-exec` as well as `--yes` (vision A.1). Without a terminal and without `--yes`, `update` prints the preview and
exits 4.

**The preview names every settings line** (Rohan, 7 Oct, format review 5.4: "Make sure it is verbose so the user
knows exactly what all things will change"). Besides each file and its result (table above), `update` and `init` list
every line they would add, change or remove in `.claude/settings.json` (hook lines, deny rules, the plugin wiring,
`autoMemoryEnabled`, `disableAllHooks`), each with one plain sentence on what it does. The sentence comes from Bonsai's
own table for its entries, and from the pack's `pack.yaml` for a pack's deny rules (each carries a `why`, which `check
--pack` requires, §5). `--json` carries the same, one entry per line (`file`, `change`, `line`, `why`). Illustrative:

```text
.claude/settings.json: 3 lines
  add     deny  Read(~/.ssh/**)
          Agents cannot read your SSH keys (base: the walls).
  change  hook  PreToolUse: bonsai hook guard || exit 2
          Checks every file edit and shell command against the task's rights; blocks if bonsai is missing.
  remove  deny  Edit(studio/old-ledger.json)
          This file is no longer in never_edit in bonsai.yaml, so agents may edit it again.
```

**Updates in a managed project need a person** (contract §17 deferred it here): `bonsai.yaml`, the lock and `.claude/**`
are person-only (contract §18 B), so an agent's `update` needs a task whose wanted paths a person granted (contract §5.5).

**Layers:** packs apply in their `bonsai.yaml` order; a path written by two packs stops the command (vision A.1, path b's
rule). Overriding a file by path waits for 1.x.

**An old 0.4.3 workspace** (`.bonsai.yaml` or `.bonsai-lock.yaml` present): `init` refuses with one plain sentence.
Rohan, 7 Oct: "dont do anything there. we will update projects when the new bonsai is up". A migration is later work.

**`bonsai check` findings** (exit 1): lock against files; format-0 files changed (contract §2.3); Bonsai-kind files
failing their format; labels against definitions; `approve_first` from history ("history not available" in a shallow
clone, contract §6); an absolute path in any committed format; a changed workspace id (contract §3); each settings rule
valid on its own (vision A.4); `disableAllHooks: true` in the project's settings or a local settings file; a `version`
in a pack's `plugin.json` or its marketplace entry; plugin drift (§5); the `bonsai` on the PATH not the installed one
(§3); the block over 40 lines, a memory index over 120 lines or 12 KB, a note over 4 KB (§10); a path named in an
instruction file, STATE or a memory note that does not exist (doc freshness); a secret-shaped string in a committed
memory note; a file from `.bonsai/local/` tracked or staged; `.bonsai/.gitignore` missing or changed.

**Warnings** are printed (in `--json`, under `warnings`) and **never change the exit code**: a Claude Code older than
Bonsai or a pack needs (§7; Rohan, 7 Oct, format review 6.5: "warn"); two checkouts on this machine holding one id; a
stale generated table, anywhere (above); a machine folder stranded under an old path (contract §3); run reports past
`generated.run`'s rule, listed for a person. **Nothing refuses to run because of Claude Code's version:** no command,
hook or rung reads it to decide; `status --json` lists the floor as a `needs` entry, never as a problem.

**Every list has one home** (Rohan, 7 Oct, format review 2.2: "how do we know which all types are possible in any place
we are using something pre-defined"; contract §2.2). Each closed or open list (task statuses, run outcomes, label value
kinds, lane rules, ask ops and types, log events and categories, exit codes, the `error` object's codes, file kinds,
generated kinds) is defined in exactly one place:
- **Bonsai's built-in schemas**, in its code: `bonsai check --schema <format>` prints a format with every field and
  every allowed value (`--schema bonsai.error` for the `error` object, §3; exit codes in `bonsai --help`).
- **A pack's declarations**: lanes, document kinds (with their statuses and moves), labels (with their `choice`
  values). They are copied into the lock (`declares`) and listed by `status --json` (`lanes`, `documents`, `labels`).
- **One reference page in Bonsai's repo**, `docs/reference/lists.md`, lists every list, its values when Bonsai owns
  them, whether it is closed or open, where it is defined, and the command that prints it in a project. It is generated
  from the code (`go generate`); a Go test rebuilds it and fails when the committed page differs, so CI catches any
  drift. Bonsai's own `CLAUDE.md` says to regenerate it with any list change (§12).

## 7. Guards and walls

Decision D stands: guards are a tripwire against accidents and the known cheats, not a wall against a same-user agent.
The walls come from deny rules and, after its probe, the sandbox. Rohan named managed settings too (7 Oct), then dropped
the machine-wide file the same day (question B), so every wall here is a project's own deny rule: a tripwire as well.

> **Added 8 Oct (Rohan):** the hook lines must not be redirectable by files an agent may edit. They call the installed
> `bonsai` by a fixed path (never found through git or the project), and whatever the guard trusts to find the main
> checkout cannot be rewritten by an agent without the guard noticing (fail closed). Why: in the studio, hook commands find
> the guard scripts through `git rev-parse --git-common-dir`, and git takes that answer from files an agent may edit (a
> worktree's `.git` file, a `commondir` file inside a `.git` folder).

**The hook lines the engine writes** (shell form, so a missing or crashing binary blocks; probed 5 Oct, P1):

| Event | Line | Blocks |
|---|---|---|
| PreToolUse: Edit, Write, MultiEdit, NotebookEdit, Bash, PowerShell | `bonsai hook guard \|\| exit 2` | Yes |
| Stop (as today) | `bonsai hook stop \|\| exit 2` | Yes |
| SessionStart (startup, resume, compact) | `bonsai hook start` | No |
| The eleven recorded events (contract §8.3) | `bonsai hook record`, async | Never |

**One hook adapter** reads Claude Code's payload (tool names, backslash paths on Windows, unknown fields). A payload it
cannot read blocks. The guard keeps its own timer under the hook's `timeout`, since a hook that times out does not block
(vision A.4). Every guard decision is a `guard` record in the log (contract §8.1).

**What the guard judges.**
- **File tools:** the active task's grants (contract §13: tasks read from main, grants only while `running`), the
  person-only list (contract §5.5), and in `command` mode the tripwires of contract §10.1. It also refuses an agent's
  write into `.bonsai/local/` and a hand edit of a generated table (§6), in every mode. Every refusal says why first.
- **Shell commands:** Rohan's 4 Oct answer (option B), which he kept on 7 Oct (question A): no shell reader; never-edited
  files carried by deny rules; a small delete check that refuses a recursive or bulk delete that does not name what it
  deletes (a `git clean` with `-x` or `-X` among them, since either deletes `.bonsai/local/`, and `git stash --all`,
  which moves it away), and one that would delete a protected path ("i dont want any accidental rm rf"). What it
  leaves (shell writes
  to person-only paths and status lines) is caught after the fact by rung 0 and the Desk's check. Today's Node
  shell guard keeps running in this repo until the studio links; the readers task still updates it (contract §15).
- **Bonsai's own commands:** `init`, `update`, `unlink` count as writes to the person-only paths; `settings set` and
  `labels` are refused (§4).

**The stop gate** (`bonsai hook stop`) is today's, as the settled contract has it (contract §13): it engages for a named
task, needs a green `local` ladder result at HEAD with the floor plus the task's rungs, and blocks on a missing or broken
task file. (Rohan's 4 Oct option B retired it; his 7 Oct checklist, `ver-stop`, and the contract brought it back.)

**What a project hook cannot do** (mods admin and hooks pages, 7 Oct): a user's mod can approve a call that a
`PreToolUse` hook outside managed settings blocked, and a local or nested `--settings` switch can turn project hooks off;
only a `PreToolUse` hook in managed settings is final. **With the managed file dropped (question B), that finality is
given up:** the guard line is a tripwire against an enabled mod too (the studio's band never approves calls), and so is
every deny rule, since Claude Code loads its built-in mod guard only where a managed file exists (mods admin page). The
engine also writes `disableAllHooks: false` into the project's settings (it beats a user `true`; a local file still
beats it, so the walls deny editing local settings). The recorder reaches linked projects only (§8). **In a project with
no `bonsai.yaml`, `bonsai hook` exits 0 at once and records nothing,** kept without the managed file for a session still
running after `unlink` or on a commit from before the link. An `rm` of `bonsai.yaml` meets the delete check (it is a
protected path); another shell route is caught after the fact, as question A says.

**Deny rules the engine writes** (in `.claude/settings.json`, kind `keys`):
- `Edit(...)` for each `never_edit` path. One such rule stops the Edit, Write, Bash (redirects, in-place edits, copies)
  and PowerShell routes in bypass mode, on both machines (P2, 5 Oct); only a write wrapped in another interpreter gets
  past. Today's ledger guard becomes this rule.
- Deny rules cannot open for a granted task, so grantable protected paths stay with the guard hook.
- Base's baseline: the walls (below); `Edit` of the home's `settings.json` and `labels/` (contract §10.6); `Edit` of
  `.bonsai/local/` (the last line of the walls below). Rohan named the asks and the log (round 1); the rule covers all
  of `local/` because Bonsai is its only writer, the ladder results included, which the stop gate trusts. Bonsai writes
  there as a program, not through the file tools, so the rule does not stop it; under the sandbox it may (contract §9.5).
- **No `Bash(...)` deny rules:** on Windows any of them removes the PowerShell tool (P2), and the delete check lives in
  the hook.

**The walls: deny rules in base** (Rohan, 7 Oct, question B: no machine-wide file). Base's manifest carries them and
the engine writes them into each linked project's `.claude/settings.json` with its other deny rules (kind `keys`), so
they apply on whichever side the project runs. The list as drafted on 7 Oct; step 5.5 lists the secret files on both
sides again:

```text
"Read(~/.ssh/**)",
"Read(~/.aws/**)",
"Read(~/.config/gh/hosts.yml)",
"Read(~/.claude/.credentials.json)",
"Read(~/.docker/config.json)",
"Read(~/.git-credentials)",
"Read(~/.bonsai/salt)",
"Read(//mnt/c/Users/*/.ssh/**)",
"Read(//mnt/c/Users/*/.git-credentials)",
"Read(//mnt/c/Users/*/.claude/.credentials.json)",
"Read(//mnt/c/Users/*/.docker/config.json)",
"Edit(//**/.claude/settings.local.json)",
"Edit(~/.claude/settings.json)",
"Edit(~/go/bin/**)",
"Edit(~/.local/bin/bonsai*)",
"Edit(//**/.bonsai/local/**)"
```

- **What they cover:** the key, token and credential files named; the rules starting `//mnt/c/` stop WSL sessions
  reading Windows ones, and on Windows `~` is the Windows home. The Windows side is short: it leaves out the Windows
  home's `.trinetra` (T-0072's), `.aws` and gh's Windows login file (`%APPDATA%\GitHub CLI\hosts.yml`; which exist is
  unchecked). `//**/` makes the local-settings rule match anywhere, not only under the session's folder. The `~/go/bin`
  and `~/.local/bin` rules stop a copy shadowing the installed `bonsai` (§3), not `go install`. Read deny rules stop
  the file tools and `cat`, `head` and `tail`, not scripts or programs (`gh`, `ssh` and Claude Code still read their own
  files). They are narrow on purpose: Rohan's own sessions and T-0072's Windows-side checks read elsewhere on C:, and
  narrow rules survive the sandbox (below). The environment scrub the research's `g-walls` row named waits for the
  sandbox: on Linux `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` puts every command in it (env-vars page). The `.bonsai/local/`
  rule uses `//**/` too, so a worktree session, whose `local/` is the main checkout's, outside its own folder, meets it.
- **The studio's own secret files** (`~/.trinetra/token`, `deploy_ed25519` and `salt`) get the same rules from the
  public `workflow` pack, since base never names the studio. The rules name the files; nothing secret is in the pack.
- **Checked in step 5.5:** each rule tried once in a scratch session on both sides (`--settings` loads the same rules).
- **Given up with the managed file, as a tripwire** (question B): the walls bind linked projects only, so Mimas until it
  links (§14 step 8) and Rohan's unlinked work have none; they sit in a project file, so a shell write that removes one is
  caught after the fact, like any write to a person-only path (question A); and a mod can approve past them (above). In
  a linked project they bind Rohan's own sessions too: there Claude cannot change his user or local settings for him
  (he uses `/config`, `/permissions` or edits by hand). Nothing needs an admin step.
- **No Claude Code version pin** (question B). Claude Code updates itself as today, on its own channel. **Warn only,
  never refuse** (Rohan, 7 Oct, format review 6.5): `bonsai check` warns when it is older than needed, a warning that
  never changes the exit code (§6); no command, hook or rung refuses for it. `status --json` lists the floor as a `needs` entry (kind `tool`, id
  `claude-code`, contract §12) for the studio's move checklist. The floor is the higher of two: Bonsai's own, a constant
  each release sets to the oldest Claude Code its tests ran on (the first from the gate's measured versions, §15), and
  any pack's, a `claude_code` entry under `needs` in `bonsai/pack.yaml` (§5). Bonsai reads `claude --version` (whether
  its output form stays stable is unchecked). There is no ceiling: a newer Claude Code that changes a hook's payload
  meets the adapter, which blocks what it cannot read (above), so the failure is loud, not silent.

**The sandbox** (WSL only; native Windows has none) waits for its probe after the gate. It needs `socat` (not installed).
The probe list: the ladder; `git push`; `bonsai ask` and `bonsai ladder` writing `.bonsai/local/` past its deny rule,
from the main checkout and from a worktree (§8, contract §9.5); issue #74081 (recursive `Read` deny globs
become one bind per matching file and commands fail; the narrow walls keep that small); and launching `cmd.exe` or
`powershell.exe` from WSL, which goes through a Unix socket the sandbox must allow (sandboxing page). Then a root step for
Rohan.

**The fallback for `.bonsai/local/`, written now** (the orchestrator, 7 Oct). Under the sandbox a worktree session may
write only its own folder, and its `local/` is the main checkout's; Claude Code also turns `Edit` deny rules into the
sandbox's write denials. If the probe finds `bonsai ask` or `bonsai ladder` blocked (a blocked ladder means no result,
so the stop gate never passes), the engine adds the main checkout's `.bonsai/local/` to the sandbox's allowed write
paths in the project's settings (kind `keys`, named in the preview). The file tools stay stopped there: by the deny
rule if the probe shows the allowance does not undo it, else by the guard alone (a tripwire either way, as today).
About 1-2 h with the probe, outside step 5's totals like the probe itself.

## 8. The recorder and asks

**The recorder.** Decision 2, confirmed 7 Oct: Bonsai records, the studio forwards and shows. `bonsai hook record` writes
`bonsai.log/1` records (contract §8) for the eleven events, async, never blocking. The guard writes its `guard` records,
the runner its `ladder` records, `bonsai log append` the studio's `event` records with labels checked against the
machine's definitions (contract §8.4).
- **Redaction in Go**, as written (decision 2), over command lines, prompts, notices and question text (vision A.4). It
  fixes the older leaks TB-057, TB-058 and TB-059 (walkthrough rows 23-24). **Proof is differential:** on the shared
  corpus (today's redaction tests plus those three bugs' cases), the Go redactor hides everything the Node one hides,
  and the three bugs' cases too.
- **Where:** the main checkout's `.bonsai/local/log/`, never committed; worktrees write main's (§6, contract §3).
  Files, keeping (`generated.log`, §6) and the cursor: contract §8.5. The input hash is keyed by the home's `salt`.
- **The sessions table** (`.bonsai/sessions.md`, §6) is built from the `session_start`, `session_end`,
  `subagent_start` and `subagent_stop` records. At each start the recorder writes the active task it finds by §6's
  rule into the record's `target` (contract §8.1), so a row keeps the task its run had, whatever moved since.
- **`bonsai hook start`** prints the opening context: the active task (id, lane, branch, grants), the last ladder result,
  and the machine-attached label definitions (contract §5.3). With the `compact` matcher it re-injects the same after
  `/compact` (`mem-compact`). It also logs the binary's own path and SHA-256 (§3).
- **Linked projects only.** Q10 (b)'s managed recorder, which would have recorded every session on the PC, went with
  the managed file (question B, 7 Oct); a project is recorded from the day it links.
- A missing binary loses records silently (recorder lines never block); the studio sees a heartbeat gap. OpenTelemetry
  stays an optional second record, later (vision A.11).

**Asks.** Contract §9, implemented as written: `bonsai ask`, `answer`, `asks`, `ask --status`; types `Answer`, `Decide`,
`Look`, `Play` from agents and `Bless` from the runner only; an answer from the asking session refused
(`CLAUDE_CODE_SESSION_ID`); an answer never grants. They live in the main checkout's `.bonsai/local/asks/` (§6).

**Under the sandbox: no outbox** (Rohan, 7 Oct, round 1: "The outbox goes"). Moving the asks into the project does not
by itself let a sandboxed `bonsai ask` write them: the deny rule over `local/` and a worktree's reach to the main
checkout still stand in the way, for `bonsai ladder` too. The probe settles it, with the fallback written in §7
(contract §9.5). A same-session answer forged by a shell write stays what it is today without the sandbox: a tripwire
case, and an answer never grants.

## 9. The ladder runner

`bonsai ladder --task T` replaces `tools/ladder/ladder.mjs`. The rungs stay each project's, in `bonsai.yaml` (4 Oct: "it
should live with the project itself").
- **Rungs:** kinds `guard` (rung 0: the diff against the named task's grants, contract §13) and `command`; each in its own
  process group on Linux and its own job object on Windows, so a rung cannot leave a process for a later one (T-0053,
  T-0062); the leftovers line at the end. Requested rungs: the floor plus the task's `bonsai.ladder` (contract §5.6).
  `--ci` writes `mode: ci`, never a task's proof (contract §11).
- **Results** go to the main checkout's `.bonsai/local/ladder/<task id>.json`, wherever the ladder ran (contract §11);
  results of done or cut tasks are cleaned after 7 days (`generated.ladder`, §6). A `--ci` run writes its own file,
  `ladder/ci.json` (as today's `ci.json`), so it never overwrites a task's local proof.
- **Rung 0 also refuses** a task branch that changes a generated table, and any file from `.bonsai/local/` that is
  tracked or staged (§6).
- **One ladder at a time per machine:** a lock in the home; a second run waits, then refuses, naming the holder. The PC's
  two homes hold two locks, so the orchestrator's rule (one ladder at a time on this PC) covers a WSL ladder and a
  Windows-side one together.
- **Ratchets and Bless:** today's condition (contract §9.2): a green `local` result on a clean base branch at HEAD with a
  count above its floor files one Bless ask. The floor rises only when the studio applies Rohan's tap (contract §10.3).
- **New tests must fail on the base** (Rohan, 7 Oct). For a count that rose, the runner names the new tests from the
  rung's test output (its `tests` field: TAP, `go test -json` or JUnit), makes a temporary worktree at the merge base,
  runs the rung's `base_setup` there (installing dependencies), copies in the branch's changed test files and runs the
  rung. **A new test counts as failed on the base only when the base run reports that test by name as failed while the
  old tests ran and passed.** A setup, import or load failure is "check not run", never "failed": otherwise a copied test
  that cannot even load would read as proof, the cheat class the research names first. **Decided here (contract §11 left
  it open): the rise is offered for Bless whatever the numbers, which the card shows**, naming any new test that passed
  on the base as "proves nothing new", and saying "check not run" when it could not run. Why: ratchets are Rohan's, and
  a new test that passes on the base can still be a fair test of behaviour that already worked.
- **Git integrity** (`ver-git`): a rung kind that records the task's base commit and flags changed test files,
  `assume-unchanged` and `skip-worktree` entries, edits to `.git/info/exclude`, new stash entries and a rewritten base.
  Not required at first: it informs the verifier.
- **The fingerprint** of each result goes through the log (contract §11).

## 10. Bonsai's home and memory

**The home** is contract §3's, with three additions: `personal/` (memory's personal layer), `locks/ladder.lock` and
`install.json` (§3). Its `cache/` (contract §3's) also keeps `--adopt` copies. The outbox went (§8). **Since 7 Oct it
keeps only what belongs to the machine** (Rohan, round 1): the salt, the home's own `settings.json`
(`cache_keep_days`, §6), each workspace's machine folder, personal memory, the install record, the ladder lock and the
pack cache. The log, asks and ladder results live in each project's `.bonsai/local/` (§6). **The per-project folder
`workspaces/r-<16 hex>/` stays, for machine settings only** (decided here, §18): `workspace.json`, `settings.json`
(`status_writes`, `status_command`) and attached `labels/`, which belong to this machine and this checkout, and must not
follow the id into a copy (contract §3).

**Memory, option D** (Rohan, 7 Oct: "Bonsai index plus notes in the repo, plus a personal layer for facts about him").
The contract left the format, the budget and the personal layer's place here (contract §7.4).
- **Notes:** one markdown file per fact in the folder `bonsai.yaml` names (`documents.memory`; this repo:
  `studio/memory/`). Frontmatter `format: bonsai.memory/1`, `id: M-<slug>`, `title`, `kind` (`project`, `feedback` or
  `reference`), `updated`, `source` (who said it, when), `labels`. Body: the fact, why, how to apply (today's notes'
  shape). At most 4 KB each.
- **The index:** `INDEX.md` in the same folder, `kind: index`, one line per note. **Budget:** a fixed 120 lines and
  12 KB (§18; there is no always-on budget since 7 Oct, §6). Today's auto-memory index is 5.5 KB.
- **The personal layer:** `~/.bonsai/personal/INDEX.md` and `notes/`, the same format, budget 40 lines. Never in a
  project, never forwarded (contract §2.6). **WSL's is the canonical one;** the Windows home holds a copy, refreshed when
  a Windows-side project links (Mimas, step 8), and `check` there reports a copy older than WSL's.
- **Loading:** the block imports the project's index; the personal index is imported once per machine from the user
  memory file `~/.claude/CLAUDE.md` (`@~/.bonsai/personal/INDEX.md`), which Claude Code loads in every project and every
  subagent (sub-agents page). That avoids the project's external-import prompt (memory page) and applies the personal
  layer to every project, Bonsai or not, as option D says. Notes are read when needed.
- **One convention:** the engine sets `autoMemoryEnabled: false` in a linked workspace (research §5: "one convention
  should win"). Background sessions and sessions started by another session already run without auto memory (memory
  page), so pickup's builders would miss it anyway. Agents write notes through the workflow pack's memory skill;
  `bonsai check` enforces the budgets, the index and the secret scan (§6).
- **The trade-off, stated:** auto memory loads the first 200 lines or 25 KB of its index and drops the rest; an import
  loads whole, so the budget is a check, not a cut-off.

## 11. Bonsai's own screens

Q7 (a): Bonsai owns one web page for one workspace, which works without the studio: packs and versions, the update
preview and diff, conflicts, labels, logs, asks, security status. **Never built before the skeleton**; when, is the
gate's choice ((d) builds it inside 1.0; (a) and (b) after). Designed first: a Fable mock and one option round. `bonsai
serve` serves it on `127.0.0.1` with a one-time token in the URL, read-only at first, over the same JSON the commands
print. **One GUI, not two:** each release ships the screens as a static bundle; the Desk loads the bundle of the release
a project pins, in a sandboxed frame with no reach to the Desk's act capability, and feeds it only what may reach the
VPS (contract §2.6: no update diffs, which are free text). Hours: a guess of 30-50, until the mock sets a real number.

## 12. Bonsai's repo: tag, clear out, records

> **Changed 8 Oct (Rohan):** "we dont need the marker tag. and we don't need to go through the whole process of PR right
> now." No marker tag (step 2), no `rebuild` branch (step 3), no draft pull request (step 4), no gate merge (step 5):
> Bonsai's work lands on `main` directly. Rohan is switching off the ruleset's pull-request and required-check rules;
> force pushes and deletion stay blocked. The old product stays reachable at tag `v0.4.3` and in git history. Step 3's
> clear-out of the old product code is the skeleton's part 1, on `main`; the old agent workspace (`station/`, the old
> `CLAUDE.md`, `.claude/`, `.bonsai.yaml`, `.bonsai/`) went first, on 8 Oct, with a new short `CLAUDE.md` and `STATE.md`.
> Step 6 is done: every branch but `main` deleted and every pull request closed on his word (8 Oct), the old routine's
> cloud sessions archived. **Records** (below): Bonsai's design, plan, STATE and run reports live in Bonsai's repo from
> 8 Oct (`design/`, `STATE.md`, `records/`), not the studio's; Bonsai still joins the studio's dashboard as its own
> project only at §14 step 6, as planned. Question C's interim proof is unchanged.

**Found on 7 Oct.** v0.4.3 was released on 13 May. `main` has 29 later commits (Plans 40 and 41, last 16 Jun), 146
remote branches and 122 open pull requests: 104 from the old maintenance routine, which ran as `LastStep` (Rohan dropped
it, `rout-bot`), 17 from Dependabot and #222 (§17). `release.yml` runs on **any** `v` tag from any commit, can re-release
a tag, and pushes Homebrew with a repository secret (§17 step 4). The `main` ruleset requires a check named `test`; the
admin role bypasses it (switching that off is Rohan's optional step, §17 step 5), and agents act on GitHub as
`LastStep`, Rohan's admin account, so what follows guards against accidents (decision D). `~/Servers/Bonsai` is the
only working clone; the ZenGarden one is left untouched (§17).

**The steps** (Q8 a: tag, then clear out in the same public repo; Q8's answer is Rohan's approval of the large delete):
1. **Rohan's GitHub steps first** (§17): the tap token moves into a `release` environment open only to `v*` tags, so no
   branch's workflow can read it, and `release.yml` is switched off until step 5.7, so no tag and no manual run can
   publish a release.
2. **A marker tag** on `main`'s head, `pre-rebuild-2026-10` (no leading `v`, so no release runs).
3. **A branch, `rebuild`.** Its first commit removes `internal/`, `cmd/`, `catalog/`, `station/`, `website/`, `docs/`,
   `embed.go`, `.bonsai.yaml`, `.bonsai/`, `.claude/`, Bonsai's own `CLAUDE.md` and `.github/workflows/docs.yml`
   (the website's). It keeps `LICENSE`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `CHANGELOG.md` (a new
   "rebuild" section), `.gitattributes`, `.gitignore`, `.golangci.yml`, `assets/`, the rest of `.github/`, and
   `.goreleaser.yaml` (the new layout keeps `cmd/bonsai` as the entry point it builds). It rewrites `README.md` ("being
   rebuilt; 0.4.3 is the old product's last release"), `Makefile` (no `go install`), `go.mod` and `go.sum`, and adds a
   short `CLAUDE.md` for agents working on Bonsai, naming no studio. Its rules for builders: no `go install` (§3); the
   rules learnt from Windows (§3); **every template and pack file documents itself, and a field change updates its docs
   in the same commit** (§5; `check --pack` fails otherwise); **every list has one home, and the reference page is
   regenerated with any list change** (§6; CI fails otherwise). **It also turns `release.yml` into a manual run with
   no Homebrew step until step 5.7;** it needs no environment (the `release` one admits only `v*` tags). A tag pushed on
   a `rebuild` commit would then run that commit's harmless workflow. Git history keeps everything.
4. **CI from the first commit:** the Linux job keeps the name `test` (the ruleset requires it); a `windows-latest` job runs
   `go test ./...` beside it (public repos pay nothing for Windows minutes). A draft pull request `rebuild` into `main`
   makes CI run on every push. Builders never push tags; only the orchestrator pushes branches.
5. **The gate decides the merge.** If Rohan continues on any path, the pull request merges and `main` is the new Bonsai;
   if he pauses, `main` keeps the old product untouched. Q8's "first step" is kept literally (the clear-out is the
   rebuild's first commit), and nothing public is deleted before the measured numbers exist. Dependabot reads its
   settings from `main`, so it keeps opening pull requests there until then; the orchestrator closes them at the merge.
6. **The routine's pull requests and branches** are closed and deleted at the merge, on Rohan's word, and the routine
   itself confirmed off on claude.ai. #222 is a stranger's and is Rohan's call (§17).
7. **Homebrew** (Q8: "0.4.3 stays downloadable (Homebrew included)"): with the first new release, the tap gains a
   `bonsai@0.4` formula pinned to 0.4.3; `brew install bonsai` moves to the new product. The old website stays online on
   GitHub Pages until Rohan turns Pages off (optional). The first release is Rohan's word.
8. **Supply chain** (vision A.9), in step 5.7: actions pinned by commit, goreleaser pinned, the re-release input removed,
   immutable releases, build provenance. Projects' CI pins the archive's SHA-256 (Q9).

**Records** (Q9 b): Bonsai's tasks, plans and run reports live in this repo under Trinetra's task numbers and a one-pager
`studio/features/F-bonsai.md`, so they show under Trinetra on the Desk for a few weeks (Q9's stated cost). They move to
Bonsai's repo when Bonsai runs itself (§14 step 6), with nothing secret or private in them. **How Bonsai's work is proven
meanwhile is question C** (Rohan, 7 Oct: (c); §14): the studio's ladder and guards reach only projects that hold
`studio/game.yaml` (`tools/ladder/ladder.mjs:91-94`; the registry's header), and Bonsai's public repo must not hold a
studio file, so the interim proof runs until step 5.4 and Bonsai's own ladder after it.

## 13. What the studio repo adopts first

> **Changed 8 Oct (Rohan): Bonsai first.** The studio's readers task is cut; the studio adopts the formats when it
> links (step 7); its Desk stays on upkeep until then.

1. **Rohan's review of every format** (§16), one at a time, recorded in a file. Everything below waits for it.
2. **The readers task** (contract §15.1 step 2, 28-44 h), unchanged. Its trick files and JSON Schemas are the ones
   Bonsai's Go reader tests against; they move to Bonsai's repo with its records (contract §2.4). When it merges, the
   Desk pause lifts and slice 3 starts first.
   > **Changed 8 Oct (Rohan):** the JSON Schemas and the trick files are Bonsai's: they start in Bonsai's repo, in
   > `formats/`, as Bonsai's first job (`design/plan.md` part 0). The readers task tests against a copy pinned to a
   > Bonsai commit.
3. **Bonsai's records here:** `F-bonsai.md` from this spec, then the skeleton's task (full lane, its stop lines in
   `done_when`) and its plan, for Rohan's approval.
   > **Changed 8 Oct (Rohan):** Bonsai's records live in Bonsai's repo (§12's note): its one-pager and plan are
   > `design/one-pager.md` and `design/plan.md`.
4. **The memory move** (§10), a small task after the format review (3-5 h): `studio/memory/` with the index and the
   project notes from today's auto-memory folder; the notes about Rohan to `~/.bonsai/personal/`, imported from
   `~/.claude/CLAUDE.md`; the block's import line in this repo's `CLAUDE.md`; auto memory off in `.claude/settings.json`
   (a protected path, so the task lists it in `allows_assets`). It needs no Bonsai code.

Not adopted early: Bonsai's hooks, the lock and the packs. They arrive when the studio links (§14 step 7).

## 14. The order of work, the hours and the stop lines

> **Changed 8 Oct (Rohan): Bonsai first.** The studio's readers task is cut; the studio adopts the formats when it
> links (step 7); its Desk stays on upkeep until then.

**What the numbers rest on.** Line counts measured on 7 Oct of the Node being ported: the guards 2,368 (of which the
shell guard 1,754, which question A (a) does not port), recorder, redaction and spool 957, asks 754, ladder, rung jobs
and process lists 1,784, the YAML reader 238, the statusline 785; about 4,700 lines of tests beside them. The engine and
packs lines come from the vision's breakdown (vision A.12: R4 and the skeptic's review), cut for what the 7 Oct split and
plugins removed. Anchors: T-0053 (process trees) planned at about 12 h; redaction took four tasks and several verifier
rounds (T-0041, T-0042, T-0043, T-0061), so it gets a wide range; Bonsai's own Plan 41 (about 3,300 lines of Go and tests
merged about 70 minutes after its plan was drafted, 16 Jun, by unattended agents) shows typing Go is not the cost:
review, verification and Windows are. **Every range below is my judgment, part by part, not a rate applied to line
counts.** Estimates are session hours. The 1.6-times record (T-0046: 38.5 h planned, 52 h taken) measured plan draft to
merge, waits included, so it overstates session time. A studio-day covers about 11 estimated hours (vision A.12).

**Re-ask lines** (vision §10): each step stops at 1.3 times its high estimate and comes back to Rohan.

| # | Step | Repo | Hours | Re-ask at |
|---|---|---|---|---|
| 0 | This spec reviewed, fixed, approved; questions A-D (answered 7 Oct, §19) | studio | — | — |
| 1 | **Rohan reviews every format** (§16) | — | his time | — |
| 2 | The readers task (contract §15) | studio | 28-44 (contract's) | contract's |
| 3 | **The walking skeleton** (below), beside slice 3 once slice 3's plan is approved, unless Rohan keeps slice 3 strictly first | Bonsai, on scratch copies | **30-47** AI hours, plus Rohan's hand checks (§17 step 6) | 61 (stop line 1) |
| 4 | **Rohan's gate** on the measured numbers (§15): path (a), (b) or (d) | — | — | — |
| 5 | The chosen path; path (a) below, in this order | Bonsai | **139-218** | per part |
| 6 | Bonsai proves and guards itself from 5.4 (question C: `bonsai ladder` on its own `bonsai.yaml`, with the pre-release Rohan installs), then links its packs (5.5); its records move to its repo; it registers as its own studio project (decision 4) | Bonsai | in 5.4-5.5 | — |
| 7 | The studio links (after 5.5): `bonsai.yaml` from `game.yaml`, hooks and CI on `bonsai`, roles and protocols from the packs, the `.claude/agents/` and `studio/roles/` copies removed, `studio/STATE.md` moved to `.bonsai/STATE.md` with the references to it, the bridge reading `.bonsai/local/`, the machine installs of `bonsai` (Rohan's steps); one ordinary week on Bonsai; a plain `git revert` of the link tried once; then today's `tools/` guards, ladder and statusline retire | studio | 9-13 (my estimate; +1 for STATE's move, 7 Oct), plus contract §15.2's after-skeleton parts (22-37) from 5.2 on | — |
| 8 | Mimas links last, in one step, on Rohan's yes (Q14 a), in Mimas's repo | Mimas | its own task | — |
| 9 | Bonsai's screens: inside 5 under path (d), else after 1.0; mock first | Bonsai | 30-50 (a guess) | after the mock |
| 10 | The first public release, on Rohan's word | Bonsai | — | — |

> **Changed 8 Oct (Rohan):** the schemas and trick files move out of row 2 into Bonsai, as its first job (`design/plan.md`
> part 0, 6-10 h, the contract's figure for that part); the studio's readers task pins a copy. The skeleton (row 3)
> follows part 0 in Bonsai's repo and waits on nothing in the studio. Row 6's records move came early, on 8 Oct; Bonsai
> still registers as its own studio project at row 6.

**The skeleton (step 3), re-costed.** The 25-35 h came from the skeptic's review: the engine, the lock, a private fetch,
CI on both OSes and one pack, "No hooks yet". This one also proves the hook path on Windows and plugin delivery. The log
and asks, which the contract first put here, moved to step 5.2 on Rohan's word (question D (a), 7 Oct; contract §15.1
step 4 now says so). Recounted after his answer B: it removed nothing here, since the managed file, the pin and their
probes were never skeleton work (they sat in §13, 5.6 and 5.7). Recounted after Rohan's format review (7 Oct): about
an hour more (part 3: `init`'s plain words, the commented `bonsai.yaml`, `.bonsai/` and its `.gitignore`, the preview's
settings lines; part 4 an hour less, as the test pack is public). The tables, the reference page, template docs and
cleaning are step 5 work. The parts:

| Part | Hours |
|---|---|
| 1. The marker tag, the `rebuild` branch and its clear-out commit (release workflow made manual), the new module layout, CI on Linux (`test`) and `windows-latest`, the draft pull request (§12) | 2-4 |
| 2. The format-1 reader in Go against the readers task's trick files; `bonsai.yaml`, the pack manifest, the lock, `status --json` (a subset) | 7-10 |
| 3. The engine: `init`, `update` (staged; kinds `pack`, `once`, `block`, `keys`; `kept`; exit 4 and 5; `--diff`, `--yes`, `--keep`, `--adopt`), `check` (lock and files, a tracked or staged `.bonsai/local/` file); line-ending-blind hashes; forward slashes; `bonsai.yaml` written with a comment on every line; `.bonsai/lock.json` and `.bonsai/.gitignore`; `init`'s closing plain words (§6); the preview naming every settings line with its sentence (§6). **Rohan's:** the PowerShell console check (check 9), saving about 1 h | 10-15 |
| 4. Packs as plugins: a public test pack, `LastStep/bonsai-test-pack` (created by the orchestrator with `gh` when Rohan approves the skeleton's plan), that passes `claude plugin validate --json` with no warning but the missing `version`; a fetch at a 40-character commit, no login; the marketplace named per workspace and commit; install and update on this machine; the drift report; two scratch projects, and a worktree beside its checkout, at two commits; a fresh worktree's trust. **Rohan's:** the interactive and fresh-worktree sessions on both sides (check 8), saving about 1 h | 4-7 |
| 5. The hook path: `bonsai hook guard` with one rule; a missing binary, a crash and an over-time run each blocking; a session started with a minimal PATH blocking with a clear reason; the binary's path and hash logged. **Rohan's:** the real Windows sessions (check 11's Windows half) and which `bonsai` Git Bash finds, saving about 1 h | 3-5 |
| 6. Measurements and the gate report, both sides of the PC (§15), with a three-case `claude plugin eval` of the test pack | 4-6 |
| **All six** (AI hours; earlier figures in §20) | **30-47** |

> **Changed 8 Oct (Rohan):** part 1 has no marker tag, no `rebuild` branch and no draft pull request: the clear-out, the
> new layout and CI land on `main` (§12's note). `design/plan.md` has the parts as they now stand.

It runs only on scratch copies: a scratch project shaped like a drifted one (one old absolute hook line and one hook of
its own), a fresh empty repo, and a scratch clone of this repo with its `origin` removed, each given its own workspace id
(`init --new-id`, contract §3). No fixture names Mimas or any one project. Test sessions run with the scratch build first
on their PATH and one shared `CLAUDE_CODE_PLUGIN_CACHE_DIR` (it moves the whole plugins root, so Rohan's own plugins show
as missing in those sessions only). For Rohan's hand checks the builder leaves the scratch folders in `~/bonsai-checks`
and `%USERPROFILE%\bonsai-checks`, each with a `claude-here` launcher that sets both and starts Claude Code
(`claude-here.cmd` on Windows); the launcher and the test build read `BONSAI_TEST_FAULT` (`missing`, `crash`, `slow`,
`minimal-path`), so each fail-closed case is one line for him (§17 step 6). Only the scratch test build reads it (a Go
build tag); a released binary has no code for it, and every fault it sets can only block, never allow. Checks marked "Rohan" are his.

**It passes when** (the audit's ten checks, updated for Go, plugins and the contract):
1. `init` into the drifted scratch project: without the required values it exits 2 naming them; with them it writes
   `bonsai.yaml` first (a comment on every line), then the rest, the lock at `.bonsai/lock.json`; the project's own hook
   stays, the old absolute line is gone, Bonsai's lines are in shell form, the deny rules carry the `never_edit` list,
   the plugin wiring holds a 40-character commit; it ends with the plain-words block (§6), and `status` shows the same.
2. Nothing written holds an absolute path, the lock included; the lock's source is the pack's remote URL.
3. The lock is JSON with forward-slash keys and verifies after a checkout with CRLF endings and under Git Bash.
4. `init` again, and `update` to the same commit, change no byte.
5. `update` to a second pack commit without a terminal and without `--yes`: the preview, naming every settings line it
   would add, change or remove with its sentence, nothing written, exit 4; with
   `--yes`: one updated, one created, the rest unchanged; a hook-line change waits for `--allow-exec`; **a new session
   then loads the second commit's roles by name**, and `check` reports no drift.
6. One pack file edited, then `update --yes`: exit 5, nothing written, the file named; `--keep` applies the rest; a third
   commit changing that file is a conflict again; `--adopt` puts the project's copy in the cache, not the repo.
7. Bonsai's format-1 reader and `yaml.mjs`'s format-1 mode reach the same outcome on every trick file.
8. A fetch from the public test pack at a commit, with no login, on WSL and on Windows, read-only; the
   pack's roles and skills load in a session; **two scratch projects pinned at two commits each load their own, and so
   do a task worktree at a second commit and its main checkout at the first, sessions alternating between them**; the
   kinds of session they load in (interactive, `-p`, `--bg`, a fresh worktree) are recorded, and `claude --agent
   workflow:builder --bg` starts the plugin's role. (Rohan: the interactive and fresh-worktree sessions.)
9. (Rohan.) One run in the PowerShell 5.1 console: ASCII output, the y/N question, a command pasted from an error
   message works.
10. `go test ./...` and `go vet` pass natively on Windows (in a Windows folder on the PC) and on `windows-latest`; no
    Windows-only skip without a named reason.
11. In a real session on WSL and on Windows (Git Bash), `bonsai hook guard` blocks an Edit of a protected path and allows
    others; a missing binary, a crash and an over-time run each block; a minimal-PATH session blocks with a clear reason.
    (Rohan: the Windows half.)
12. A plain `git revert` of the link commit restores the project's old state with no Bonsai binary present.

**Stop lines** (written into the skeleton's task before it starts; judged by a fresh verifier, not the builder):
1. **Hours spent** over 61 h (1.3 times the top of 30-47, rounded, Rohan's rule; question D). Hours spent: the sum of
   every builder and verifier session's wall-clock on the task. Bonsai does not record this repo until the studio links
   (step 7), so the run report's "Sessions" table, now retired as a format (§16 row 28), gives way to the Desk's session
   records: the orchestrator sums them for the task and writes each session and subagent run (role, start, end,
   minutes) in the run report as plain text. Once a project links, `.bonsai/sessions.md` holds these rows, one per
   session and per subagent run, each with its task by §6's rule; the line then adds the builder and verifier rows of
   the task (a run whose task the rule could not find counts as `none`, and the orchestrator adds it by hand). The estimates above are
   in the same unit: they include verification. Rohan's hand checks are not counted: the line counts AI hours.
   > **Changed 8 Oct (Rohan):** with Bonsai's records in its own repo, hours come from Bonsai's run reports
   > (`records/runs/`): each builder and verifier run's start, end and minutes, summed by the orchestrator, not the
   > studio's session records.
2. More than a day lost to Go on Windows: 8 session hours on failures that happen only on Windows.
3. More than two option rounds asked of Rohan inside the skeleton (his hand checks are not option rounds).
4. Any change to Mimas, or to this repo's `tools/`.

The vision's fifth line (no Mimas work in the same days) does not apply (walkthrough row 7). If a line is crossed, Rohan
gets the numbers and three choices: continue, take the smaller cut, or pause. Changing the vehicle is not offered (vision
§8.1): only Rohan can reopen Go.

**How Bonsai's work is proven** (question C (c), Rohan, 7 Oct). The skeleton and steps 5.1-5.3 run on the interim
proof: `go test ./...` and `go vet` in WSL and natively on Windows (in a Windows folder), their output in the run report,
Bonsai's CI beside it, and a fresh verifier on every step. From step 5.4 Bonsai's tasks climb `bonsai ladder` on a
`bonsai.yaml` in Bonsai's own repo (Bonsai's format, not a studio file), run by a pre-release `bonsai` Rohan installs in
WSL (§17 step 8); Bonsai's guard then covers its own checkout too. Golden rule 8 still asks a fresh verifier for the big
steps.

**Path (a) after the gate (step 5), 139-218 h** (222-349 on the 1.6-times record), with Rohan's 7 Oct answers: A (a), B
dropped, C (c), D (a), and his format review (+16-23 h, below):

| Part | What | Hours | Re-ask at |
|---|---|---|---|
| 5.1 Formats and engine to 1.0 | Every contract format in Go and the format-0 hand port; the schema-compare rung (contract §2.2); layers, `--allow-exec`, `unlink`, the old-workspace refusal, all of §6's findings and warnings, `--schema`, `status --full` (26-40); the tasks table and `check --write`, with its stale rules (2-3); the reference page of lists, generated, and its CI test (1-2); `check --pack`, the template docs kept in step with their fields (1-2) | 30-47 | 61 |
| 5.2 Recorder, logs, asks | Redaction in Go with TB-057, 058 and 059 fixed and the differential check (10-15); the recorder, files in `.bonsai/local/`, `logs`, `log append`, and cleaning per kind with its protections, the `clean` event and the generated-files page (9-13); asks to contract §9, no outbox (4-6); the sessions table from the log, a row per session and per subagent run with its task (2-3). The studio's forwarder can start here | 25-37 | 48 |
| 5.3 Guards | The adapter (2-3); the path guard with contract §5.5, §10.1 and §13, and its refusals for `.bonsai/local/` and the tables (7-10); the delete check (3-5); the stop gate (2-3); generated deny rules and `disableAllHooks: false` (1-2); the binary check (1-2); the second Windows check: guard, delete check, recorder under concurrency, backslash paths (2-4; **Rohan's:** its real Windows session, §17 step 7, saving about 1 h) | 18-29 | 38 |
| 5.4 Ladder runner | Rungs, process groups and job objects, one ladder at a time, leftovers, `mode`, the fingerprint, results in the main checkout's `.bonsai/local/ladder/`, rung 0's refusal of branch changes to the tables and of tracked `local/` files (13-19); floors, ratchets, Bless filing (4-6); new tests must fail, by name, with `base_setup` (7-12); git integrity (3-5); Bonsai's own `bonsai.yaml`, the switch from the interim proof and the pre-release build Rohan installs (question C, 1-2) | 28-44 | 57 |
| 5.5 Packs | `base` and `workflow` from this repo's roles, protocols and templates; the always-on and skill split, the roles' `skills:` preloads; `claude plugin validate --json` in each pack's CI, failing on every warning but the missing `version`; the walls in base's deny rules and the studio's in `workflow`, each tried once on both sides (question B, 1-2) (13-21); the documentation in every template and pack file of `base` and `workflow`, and each deny rule's `why` (3-4); the pack template `packs/template/` with its CI and release, and Bonsai's CI job that runs it (3-5) | 19-30 | 39 |
| 5.6 Machine pieces | `settings` (with `--machine` and `cache_keep_days`), `labels`, the personal memory layer and its check, the stranded-folder report (6-9); `status --line`, the workspace half of today's statusline (5-8); the installers for `/usr/local/bin` and `C:\Program Files\Bonsai` with `install.json` (2-3) | 13-20 | 26 |
| 5.7 Release | The supply-chain fixes, `release.yml` back to tag runs inside the `release` environment and switched on again (Rohan's step, §17), `bonsai@0.4`, the README | 6-11 | 14 |

Rohan's answer B took out `walls --print` (5.6, about 1 h) and the probe script for pin raises (5.7, about 2-3 h), and
the uncosted work round the managed file: the orchestrator writing and checking it, the second batch's probe, and
Rohan's admin batch and pin-raise checks (§17's old steps 5 and 8). It added the walls' checks in 5.5 (1-2 h);
`check`'s version warning replaces its managed-file warning in 5.1 at no change.

**Rohan's format review (7 Oct) added 16-23 h to step 5:** 5.1 +4-7 (the tasks table, the reference page, `check
--pack`), 5.2 +3-4 (cleaning per kind, the sessions table with subagent rows, less the outbox's 1-2), 5.3 +1, 5.4 +1, 5.5 +6-9 (template
docs, the pack template), 5.6 +1. Dropping the always-on budget and the older-major rule saved nothing measurable: the fixed
caps are checked the same way, and 1.0 reads no older major. The readers task (contract §15.1) is unchanged; the
studio's adoption gains 1-2 h (contract §15.2: the tables in each move's commit).

**Totals.** Bonsai 1.0 under path (a), with the skeleton: 169-265 AI hours (270-424 on the record), about 15-24
studio-days, roughly three to five weeks for Bonsai alone, best case; the studio's own work (contract §15) comes on top.
Path (b), the smaller cut, defers the new-tests check, git integrity, `status --line` and half of 5.7: about 121-188 h
after the gate. Path (d) adds the screens: about 169-268 h after the gate. A later upgrade of the
guard's shell side waits for the Jev trial and Rohan's decision on its research (§19 A); it is not costed here.

## 15. What the gate measures

**Measured on 7 Oct, for this spec** (the baseline; scratch folders only, nothing installed; Bonsai's `main` at `c6a6757`):

| Measure | Result |
|---|---|
| Go on Windows | Go 1.26.2 in `C:\Program Files\Go`, on the PATH. An old `bonsai.exe` sits in `%USERPROFILE%\go\bin` |
| Go in WSL | 1.24.2; the clone needs 1.25 (`GOTOOLCHAIN=local` refuses); with the default it downloads go1.25.9 into a scratch cache: first build 42 s, then 0.3 s |
| Windows Go building from WSL's disk | Fails at once: Go cannot lock `go.mod` over the share ("Incorrect function"). Windows builds need a Windows folder (or CI) |
| Native Windows build (a Windows temp folder, its own caches) | Cold 23.3 s (module downloads included), warm 0.3 s, `go vet` 4.9 s, clean |
| Native Windows tests | 9.2 s; 9 of 17 packages pass, 8 fail, 29 tests, all Windows-only (in WSL all 17 pass in 3.1 s). Causes: backslash paths from `filepath.Join` in lock keys and JSON paths (the audit's known bug), symlinks that need a privilege, file modes Windows does not keep, path checks that reject backslashes, byte-exact golden files. Most of that code is cleared out; the lock and conflict paths, whose ideas survive, are among the failures |
| Cross-compile from WSL (`GOOS=windows`) | 12.4 s cold, 0.26 s warm. Proves it compiles; proves nothing about running |
| Hook start-up in WSL, per call | A minimal Go hook (reads the payload, judges a path): 1.5 ms. Today's Node path guard: 26 ms (`node -e 0`: 18 ms). The old Bonsai binary: 270-460 ms |
| Hook start-up on Windows under Git Bash, per call | Go: 65 ms. Today's Node path guard (Mimas's pinned copy): 81 ms (`node -e 0`: 62 ms). Git Bash's own process start dominates |
| Claude Code and the PATH | WSL 2.1.292, Windows 2.1.291; no managed settings file and no HKCU or HKLM policy key (none is planned since question B); the running bridge's PATH holds `/usr/local/bin` but not `~/.local/bin` or `~/go/bin` |

Not measured yet (the skeleton's first hours): a real Claude Code session on Windows calling a Go hook; Windows job
objects; Windows Defender's first-run scan of a new binary; `windows-latest` CI; plugin install and update from a pinned
commit.

**At the gate, the skeleton reports:** hours per part against its estimate, and the ratio applied to step 5's estimates;
each check, passed or not; Windows-only failures and the hours they took; hook start-up p50 and p95 on both sides,
against the baseline above; the fail-closed and minimal-PATH results; in which session kinds (interactive, `-p`, `--bg`,
a fresh worktree) the pack's roles and skills load, which decides pickup; plugin install and update time, and whether
two projects, and a worktree beside its checkout, keep two commits; the Claude Code version on each side, which sets
Bonsai's first floor for `check`'s version warning (§7); binary size; lines of Go per part; and the three-case `claude
plugin eval`, as a sample of how evals would run, not as proof that packs help. Rohan picks (a), (b) or (d).

## 16. The formats for Rohan's review

Rohan, 7 Oct: "once the final formats and contracts are decided, i want to go through each and review them myself. once
i confirm then we will move forward with this." Every format below, one at a time; nothing is built on any of them before
he confirms them all. **He read rows 1-29 on 7 Oct** (`studio/decisions/REVIEW-2026-10-07-formats.md`, round 1). What
changed on his answers is marked "Changed 7 Oct" with the round 2 item of that file that says how, and rows 30-37 are
new: both are for his second look.

**The contract's** (`docs/specs/2026-10-06-contract.md`):

| # | Format | What it is | Where |
|---|---|---|---|
| 1 | The rules every format follows | One owner and a version line in each file; changes only add; one strict YAML grammar; today's files read forever as "format 0"; only listed fields reach the VPS. Changed 7 Oct (R2.5, R2.7, R2.8, R2.11) | contract §2 |
| 2 | The Bonsai home and the workspace id | The machine folder for what never goes into a project, and the random id that names a project. Changed 7 Oct (R2.4) | contract §3 |
| 3 | The task (`bonsai.task/1`) | One unit of work: its fields and the eight statuses | contract §4 |
| 4 | Labels and their definitions (`bonsai.labels/1`), grants | Named extra values on a document, what each means, and how protected paths are asked for and granted | contract §5 |
| 5 | Lanes (`bonsai.lanes/1`) | Light, full and director, and the two rules base understands | contract §6 |
| 6 | The run report (`bonsai.run/1`) | The log of one agent session on one task | contract §7.1 |
| 7 | STATE (`bonsai.state/1`) | The one page that says where a project stands. Changed 7 Oct (R2.1) | contract §7.2 |
| 8 | Document kinds | How a project declares its plans, decisions and other documents, and who moves their statuses | contract §7.3 |
| 9 | The log record (`bonsai.log/1`) | One line per thing an agent did, cleaned of secrets as it is written. Changed 7 Oct (R2.2, R2.3, R2.6) | contract §8 |
| 10 | The ask record (`bonsai.ask/1`) | A question an agent asks you, and your answer. Changed 7 Oct (R2.3) | contract §9 |
| 11 | Status moves and Desk taps | The studio's one write path, which moves are yours, and how a move proves it was a person's (the studio's, not a Bonsai file). Changed 7 Oct (R2.2) | contract §10 |
| 12 | The ladder result (`bonsai.ladder/1`) | The proof that a task's checks passed, at a commit. Changed 7 Oct (R2.3, R2.6) | contract §11 |
| 13 | `bonsai status --json` (`bonsai.status/1`) | One workspace at a glance, for the studio and Bonsai's screens. Changed 7 Oct (R2.4) | contract §12 |
| 14 | The active task | The one rule for which task's permissions apply right now | contract §13 |
| 15 | The lock (`bonsai.lock/1`) | Exactly which pack versions and files a project holds. Changed 7 Oct (R2.1) | contract §14 |

**Added or changed by this spec:**

| # | Format | What it is | Where |
|---|---|---|---|
| 16 | `bonsai.yaml` (`bonsai.workspace/1`) | A project's packs, folders, protected lists and ladder: the Bonsai half of today's `game.yaml`. Changed 7 Oct (R2.5, R2.6) | §6 |
| 17 | The pack (`bonsai.pack/1` and the folder layout) | What a pack holds besides the plugin, its `needs` (the oldest Claude Code it works with among them), and its rule of no `version` in `plugin.json` or the marketplace entry. Changed 7 Oct (R2.9, R2.10, R2.11) | §5, §7 |
| 18 | The lock's `declares` and its place | What each pack declared, copied in so CI needs no pack (adds to contract §14). Changed 7 Oct (R2.1) | §6 |
| 19 | Memory (`bonsai.memory/1`), its index and the personal layer | One note per fact in the repo, a short index, and your personal notes on the machine. Changed 7 Oct (R2.5) | §10 |
| 20 | The instruction block in `CLAUDE.md`, and which protocols load always | The few lines Bonsai keeps in `CLAUDE.md`, and the always-on versus skill split. Changed 7 Oct (R2.5) | §5, §6 |
| 21 | Bonsai's entries in `.claude/settings.json` | Its hook lines, deny rules (the walls round secret files among them), the plugin marketplace named per workspace and commit, auto memory off, hooks not switched off. Changed 7 Oct (R2.3, R2.9) | §5, §7 |
| 22 | The home's additions | `personal/`, `locks/`, `install.json`; `--adopt` copies in `cache/` (adds to contract §3). Changed 7 Oct (R2.3, R2.4; the home's own `settings.json` holds `cache_keep_days`) | §10, §3 |
| 23 | The ladder rung's `tests` and `base_setup`, the new-tests rule for Bless, and the git-integrity rung kind | How the runner names new tests and checks they fail before the change; a third rung kind in `bonsai.yaml`'s ladder and in ladder results (adds to contract §11) | §9 |
| 24 | `status --json`'s additions | A pack not installed here, plugin drift, a swapped binary, the Claude Code floor (a `tool` need, never a problem: warn only); `--line` (adds to contract §12) | §4, §5, §6, §7 |
| 25 | ~~The ask outbox under the sandbox~~ | Gone (Rohan, 7 Oct, round 1): asks live in `.bonsai/local/`; the sandbox's probe settles how `bonsai ask` writes there, with a fallback written (§7) | §7, §8 |
| 26 | The estimate labels | `workflow.estimate_h` on tasks; the studio's `trinetra.actual_h` | §2 |
| 27 | The binary's path and SHA-256 in the log | A field `hook start` writes (adds to contract §8) | §3, §8 |
| 28 | ~~The run report's "Sessions" table~~ | Replaced (Rohan, 7 Oct, round 1) by the sessions table in `.bonsai/` (row 30), which no agent writes | §6, §14 |
| 29 | The `error` object in every command's `--json` | On a refusal or error: `code` (a fixed word), `message`, and `next`: what to do, and whether an agent or a person does it. Its codes are a list with one home (row 34) | §3 |
| 30 | `.bonsai/` and its two tables (`bonsai.tasks/1`, `bonsai.sessions/1`). New 7 Oct | The lock, STATE and two tables Bonsai rebuilds from the task files and the log (`check --write`); a sessions row per session and per subagent run, with its task by one rule; committed, written only in the main checkout, never on a branch; stale is a warning; a table grants nothing | §6, contract §3, §7.5 |
| 31 | `.bonsai/local/`. New 7 Oct | The log, asks and ladder results inside the project, never committed; worktrees share the main checkout's; only Bonsai writes there | §6, contract §3 |
| 32 | Generated files (`generated:` in `bonsai.yaml`, the `clean` log event, the page). New 7 Oct | How long each generated kind is kept (`keep_days`, `keep_newest`, protections), defaults, and every cleaning logged | §6, contract §8 |
| 33 | Template documentation. New 7 Oct | Every template and pack file documents itself; a file made from one carries a one-line pointer on its `format:` line; `check --pack` keeps docs and fields in step | §5, contract §2.8 |
| 34 | The reference page of lists. New 7 Oct | Every list, its values, closed or open, where it is defined and how to print it; generated from the code, checked in CI | §6, contract §2.2 |
| 35 | The update preview's settings lines. New 7 Oct | Every settings line `update` or `init` would add, change or remove, each with one plain sentence (`file`, `change`, `line`, `why` in `--json`) | §6 |
| 36 | The pack template and a pack's CI. New 7 Oct | `packs/template/` in Bonsai's repo and its `pack.yml`: validate as a plugin, `check --pack`, tests, release on a tag | §5 |
| 37 | `init`'s closing words, and `status`'s same block. New 7 Oct | Where the home and `.bonsai/` are, what each holds, that none of the home enters git, the id | §6, contract §3 |

Not a file format, but part of the contract he reviews: the command words and exit codes (§3, §4); still fourteen words
(`check --write` and `check --pack` are flags). The managed settings file that stood as row 26 went with question B.

**Changes to the settled contract's text**, each on Rohan's word: §10.3's erratum (his 6.7 "Agree") and §15.1 step 4
(question D (a)) on 7 Oct, the rest from his format review the same day. The contract's §21 lists the sections.
- **§2.2:** the older-major rule (60 days and two releases) removed: "a problem to be handled when we update the formats
  for the first time" (1.1). Every list defined in exactly one place, with the reference page (2.2, 3.4).
  `bonsai schema` is `bonsai check --schema`.
- **§2.4:** writers keep comments (1.4). **§2.8, new:** templates document themselves (his new rule).
- **§3:** the project's `.bonsai/` and `.bonsai/local/`; the home keeps only what belongs to the machine, and its
  per-project folder only machine settings; `log_keep_days` leaves the machine settings; moved and copied checkouts;
  `init` and `status` say where everything is (1.2; round 1, decisions 1 and 2). **§1's table:** where each format lives.
- **§7.2:** STATE at `.bonsai/STATE.md`. **§7.3:** `state` is a fixed file; the two tables are declared. **§7.5,
  new:** the generated tables (3.2, 5.3; round 1, decision 1).
- **§8, §8.2, §8.5:** the log in `.bonsai/local/log/`; the `clean` event; keeping set in `bonsai.yaml` (round 1,
  decision 2; 5.1).
- **§9.1:** asks in `.bonsai/local/asks/`, kept by default. **§9.5:** the outbox gone (round 1, decision 2).
- **§10.3:** pickup starts `workflow:builder` (the erratum, now in the text; 6.7). **§10.5:** the rebuilt tables ride
  each move's commit. **§10.6:** the wall over `.bonsai/local/` (round 1).
- **§11:** results in the main checkout's `.bonsai/local/ladder/`, cleaned after 7 days except an open task's (5.1;
  round 1). **§12:** `home` and `local`. **§13:** the stop gate reads the result from there; the tasks table grants
  nothing (5.3). **§14:** the lock at `.bonsai/lock.json` (6.2).
- **§15.1 step 4:** the log and asks arrive at step 5.2 (question D (a), unchanged since). **§15.2:** the apply core
  +1-2 h, total 48-77. **§16 item 10:** the bridge reads each registered project's `.bonsai/local/`, a Windows-side
  one's over `/mnt/e` (round 1, decision 2). **§17:** the machine folder's row.
- **After the second recheck** (`docs/specs/2026-10-07-bonsai-recheck-2.md`; technical calls by the orchestrator, 7 Oct,
  shown to Rohan in round 2): **§7.5** a stale table is a warning everywhere; a sessions row per subagent run too,
  with its task by one rule. **§8.1** `target` on `session_start` and `subagent_start` holds the active task found.
  **§8.5** a log file is kept until its rows are in. **§9.5** the sandbox fallback. **§10.5** `check --write`'s exit
  code; a move goes ahead without the tables. **§11** `--ci` writes its own file. **§13** one more fixture. **§15.1**
  the two tables' schemas come with Bonsai. **§16 item 10** a Windows-side project uses Windows `bonsai.exe`.

## 17. Your steps (not questions)

None of these is run for you. Each says what it does, why, and whose rights it needs. Bash lines go in a WSL terminal;
PowerShell lines in PowerShell, one line at a time.

**1. The outside contributor's pull request, #222** (your account; a public act). "feat(update): show guidance after
backing up conflicts", by bferanmi806-sketch, 29 Aug. The code it changes is being cleared out, so it cannot land. Close
it with a thank-you, or leave it open until the gate merge:

```bash
gh pr close 222 -R LastStep/Bonsai --comment "Thank you for this. Bonsai is being rebuilt from scratch, so this change cannot land; 0.4.3 stays the last release of the old design."
```

> **Changed 8 Oct (Rohan):** done: #222 is closed. Every other pull request is closed and every branch but `main`
> deleted, on his word.

**2. The old ZenGarden clone** (your files; nothing needs admin). `~/ZenGarden/Bonsai` holds 9 branches with commits
that are on no remote: `feat/odysseus-sync` (1 commit, 7 Jun), `pr-72` (7), `ui-ux-testing-pre-rebase` (6),
`ui-ux-testing-pre-iter2-rebase` (4), `worktree-agent-a12e0a9b` (7), `worktree-agent-a3aa71c2` (2),
`worktree-agent-a5e5f344` (2), `docs/starlight-scaffold` (1), `fix/case-insensitive-file-collision` (1), and 8 worktrees,
6 of them locked by old agent sessions (counted read-only on 7 Oct). Nothing in the rebuild needs them, and agents leave
the clone alone. If you want a private backup of every branch in one file before you ever clean it up:

```bash
git -C ~/ZenGarden/Bonsai bundle create ~/bonsai-zengarden-branches.bundle --branches
```

**3. The four old `bonsai` binaries** (your files; the Windows one needs no admin). They are 0.4.3-era and come first on
the PATH, so hooks calling `bonsai` would find them instead of the new one. Remove the three in WSL before step 5.4,
when Bonsai's own hooks start calling `bonsai` (step 8 below), and the Windows one before Mimas links, unless you still
use the old `bonsai` command in ZenGarden. A new copy in those folders would shadow it again; the walls deny agents'
copies there, and `bonsai check` reports one (§3):

```bash
rm ~/go/bin/bonsai
rm ~/go/bin/Bonsai
rm ~/.local/bin/bonsai
```

```powershell
Remove-Item "$env:USERPROFILE\go\bin\bonsai.exe"
```

**4. GitHub, before the `rebuild` branch is first pushed** (your account, admin on the repo; a secret). Today any
workflow in any branch can read the Homebrew token. These steps put it in an environment only `v*` tags reach:
- On github.com, create a new fine-grained token: Settings, Developer settings, Fine-grained tokens, repository
  `LastStep/homebrew-tap`, permission "Contents: read and write". Copy it.
- In WSL, create the environment, open it to `v*` tags only, store the new token there (paste it when asked), and delete
  the repository-wide copy:

```bash
gh api -X PUT repos/LastStep/Bonsai/environments/release -F "deployment_branch_policy[protected_branches]=false" -F "deployment_branch_policy[custom_branch_policies]=true"
gh api -X POST repos/LastStep/Bonsai/environments/release/deployment-branch-policies -f name='v*' -f type=tag
gh secret set HOMEBREW_TAP_TOKEN --env release -R LastStep/Bonsai
gh secret delete HOMEBREW_TAP_TOKEN -R LastStep/Bonsai
```

- On github.com, revoke the old tap token (Settings, Developer settings, its token list).
- In WSL, switch the release workflow off, so no tag and no manual run can publish a release until step 5.7. Off is
  per workflow file, so it should cover every branch, the `rebuild` one included (unchecked). Nothing needs a release
  before then:

```bash
gh workflow disable release.yml -R LastStep/Bonsai
```

- At step 5.7, on your word, switch it back on: `gh workflow enable release.yml -R LastStep/Bonsai`.
- Optional: a tag ruleset (the repo's Settings, Rules, Rulesets, New tag ruleset: target `v*`; restrict creations,
  updates and deletions; no bypass). You switch it off for the minute you push a release tag. Since agents act on GitHub
  as your account, this stops an accident, not a determined agent.

> **Changed 8 Oct (Rohan):** done in a cut-down form. The repository-wide `HOMEBREW_TAP_TOKEN` secret is deleted and
> `release.yml` is disabled. The `release` environment and a new tap token wait for step 5.7. The environment command
> above fails as written; the working form, untested, followed by the `v*` tag policy line above:
>
> ```bash
> echo '{"deployment_branch_policy":{"protected_branches":false,"custom_branch_policies":true}}' | gh api -X PUT repos/LastStep/Bonsai/environments/release --input -
> ```
>
> Not confirmed yet: that the old tap token was revoked on github.com. There is no `rebuild` branch (§12's note).

**5. Optional: the administrators' bypass on Bonsai's `main`** (your account, admin on the repo; any time). The ruleset
`main-protection` (read on 7 Oct) asks for a pull request and a green `test` check, refuses force pushes and deletion,
and lets the admin role bypass all of it, always. Agents act on GitHub as your account, so the bypass is theirs too.
Turning it off makes every change to `main` go through a pull request with a green `test`, yours included. **Default:
leave it as it is.** Guards are a tripwire (decision D), and with one account it only stops an accident; the gate merge
is a pull request with CI anyway (§12). If you want it off: on github.com, the repo's Settings, Rules, Rulesets,
`main-protection`, the bypass list: remove Repository admin, then save (the page's exact labels are unchecked).

> **Changed 8 Oct (Rohan):** overtaken. Work lands on `main` without pull requests (§12's note): Rohan is switching off
> the ruleset's pull-request and required-check rules; force pushes and deletion stay blocked. The orchestrator pushes
> `main` after a verifier passes and checks CI after every push.

**6. The skeleton's hand checks** (your time, about 45 minutes in one sitting, when the builder says the scratch build
is ready; they save about 3 AI hours, §14). The builder leaves the folders and a `claude-here` launcher in
`~/bonsai-checks` and `%USERPROFILE%\bonsai-checks`. Tell the orchestrator what you saw, one line per check.

*a. The pack loads in a session* (check 8). In WSL, then the PowerShell pair in a normal PowerShell:

```bash
cd ~/bonsai-checks/project-a
~/bonsai-checks/claude-here
```

```powershell
cd "$env:USERPROFILE\bonsai-checks\project-a"
& "$env:USERPROFILE\bonsai-checks\claude-here.cmd"
```

Type `/agents`: the test pack's `marker` role should say "commit A". Type `/exit`. Do the same in `project-b` ("commit
B") and in `project-a-worktree`, a fresh worktree of `project-a` moved to commit B: there, note whether a trust question
appears and whether the role shows ("commit B"). Then open `project-a` once more: still "commit A".

*b. The PowerShell console* (check 9). In a normal PowerShell:

```powershell
$env:Path = "$env:USERPROFILE\bonsai-checks\bin;$env:Path"
cd "$env:USERPROFILE\bonsai-checks\project-conflict"
bonsai update
```

The text should be plain, with no broken characters, and end in a y/N question: answer `n`. Then run the line below; it
stops on a conflict and prints a command. Paste that command: it should work.

```powershell
bonsai update --yes
```

*c. The guard in a real Windows session* (check 11). In a normal PowerShell:

```powershell
cd "$env:USERPROFILE\bonsai-checks\project-guard"
& "$env:USERPROFILE\bonsai-checks\claude-here.cmd"
```

Ask Claude to add a line to `protected.txt` (refused, with a reason), then to `free.txt` (done). Type `/exit`. Then
open four more sessions the same way, each after one of these lines, asking only for the `free.txt` edit; each must be
refused with a clear reason. Close that PowerShell window afterwards, so the switch goes with it.

```powershell
$env:BONSAI_TEST_FAULT = "missing"
$env:BONSAI_TEST_FAULT = "crash"
$env:BONSAI_TEST_FAULT = "slow"
$env:BONSAI_TEST_FAULT = "minimal-path"
```

*d. Which `bonsai` Git Bash finds* (part 5). In Git Bash on Windows (not WSL), and send the output:

```bash
which -a bonsai
```

**7. The second Windows check** (step 5.3; about 10 minutes; saves about 1 AI hour). Open a session in the same
folder, with the full guard built in:

```powershell
cd "$env:USERPROFILE\bonsai-checks\project-guard"
& "$env:USERPROFILE\bonsai-checks\claude-here.cmd"
```

Ask Claude to change a file only you may change (refused), to delete the `junk` folder with a recursive delete that
does not name what it deletes (refused), and to delete `junk/one.txt` by name (done). Tell the orchestrator.

**8. The pre-release `bonsai` at step 5.4** (root in WSL, your password; about 5 minutes; question C). From 5.4
Bonsai's tasks climb `bonsai ladder` and Bonsai's guard covers its own checkout, so WSL needs an installed `bonsai`. Do
step 3's WSL lines first. A 5.x build is not a release (§18): this is the one install before 1.0, unless the
orchestrator asks again because a later step changed the guard, the stop gate or the ladder (the same lines then).
Windows needs nothing until 1.0. The orchestrator builds it from the 5.4 commit into `~/bonsai-checks/prerelease/` and
gives you its SHA-256. In WSL:

```bash
sha256sum ~/bonsai-checks/prerelease/bonsai
sudo install -o root -g root -m 0755 ~/bonsai-checks/prerelease/bonsai /usr/local/bin/bonsai
which -a bonsai
bonsai --version
```

The first line must print the orchestrator's number; if not, stop and tell it. `which -a` must list
`/usr/local/bin/bonsai` with nothing before it. Send the orchestrator the last line. To take it out again:
`sudo rm /usr/local/bin/bonsai`.

## 18. Decided here, and why

Each decision's reason now sits in its own section. The summary table that stood here repeated them and was cut on 7
Oct after the recheck; it is in git history (`git show 131f6f8:docs/specs/2026-10-07-bonsai.md`).

Rohan's answers, 7 Oct, after the recheck:
- **Who sets up and updates Bonsai: the spec as it is.** Rohan installs and updates the program on each side (§3); an
  agent's `init`, `update` and `unlink` still need a task he approved that names those files, since they write
  person-only paths (§6). He declined "agents do it, he pastes one line per release" and "agents do everything, the
  program in the home folder". His wish that any AI agent can drive the program easily is met by §3's unattended rule.
- **Agents other than Claude Code: Claude Code first** in 1.0 (Q5 a); others later, on his word.
- **Question A, the guard: (a), his 4 Oct answer** (the file guard, deny rules, the delete check; shell writes to his
  paths caught after the fact by the first rung and the Desk). A later upgrade comes with Jev, after its trial (§19 A).
- **Question B, the admin batch: dropped** (no managed file, no pin, no second batch; Q10 and Q11 reversed). Why, in
  his words: "why do we need this. i dont think we do, seems to be overenginnering." The walls move into the base
  pack's deny rules; `check` warns on a Claude Code older than Bonsai needs (§7, §19 B).
- **Question C, Bonsai's proof: (c)**, the interim proof (Go tests on both sides, CI, a fresh verifier) until step 5.4,
  then Bonsai's own ladder; Rohan installs a pre-release `bonsai` at 5.4 (§19 C).
- **Question D, the skeleton: (a)**, the log and asks move to step 5.2 and the stop line is 60 h (1.3 times the top of
  30-46 h), recounted after B's drop, which removed nothing from it; contract §15.1 step 4 changed on his word (§19 D).
  After his format review the skeleton is 30-47 h and the same rule gives 61 h (§14, §20).
- **Work that is minutes for him and hours for an agent is his** ("so the overall process speeds up and consumes less
  ai hours"): hand checks on Windows, interactive session checks, looking at a screen (§14, §17 steps 6 and 7).

The recheck's open minors, decided here on 7 Oct (technical, so not Rohan's):
- **11, one marketplace name at two commits:** the name carries the workspace and a hash of the lock's pack commits
  (§5), so a task worktree and its main checkout, or two clones, at different commits never share one; skeleton check 8
  and Rohan's §17 step 6a test it.
- **17, the index budget:** a fixed 120 lines and 12 KB (§10). The always-on budget it was weighed against went on
  Rohan's word in his format review (1.4: "Don't use" it).
- **18, are 5.x builds releases for §3's installs:** no; before 1.0 Rohan installs one pre-release at 5.4, in WSL only,
  where Bonsai's checkout and tasks live, and again only if a later 5.x step changes the guard, the stop gate or the
  ladder (§3, §17 step 8).
- **10, the administrators' bypass on Bonsai's `main`:** Rohan's optional step, default left on (§17 step 5).

Decided inside Rohan's format review (7 Oct, round 1, and the second recheck; technical, so not his; each reason is in
its section):
- `bonsai check --write` (a flag; still fourteen words) rebuilds the tables in the main checkout only; the studio commits
  them with each move (§6, contract §10.5). Both committed, never written on a branch; stale is a warning everywhere;
  a sessions row per session and per subagent run, its task by one rule (§6, contract §7.5).
- STATE's place is fixed at `.bonsai/STATE.md` (§6, contract §7.2). The home's `r-<16 hex>` folder stays, for machine
  settings only (§10, contract §3).
- Keeping: per kind under `generated:` in `bonsai.yaml`, the cache in the home's own settings; every removal logged; a
  log file kept until its rows are in; Bonsai never deletes a committed file (§6).
- One deny rule over all of `.bonsai/local/` (§7); the outbox goes, with the sandbox fallback written (§7, §8); `--ci`
  writes its own result file (§9).
- The pack template is `packs/template/` in Bonsai's repo, run by Bonsai's CI (§5).

## 19. Answered, 7 Oct

Rohan answered the four questions on 7 Oct. The options, their costs and the recommendations are in git history:
`git show 39cdb99:docs/specs/2026-10-07-bonsai.md`.

### A. Does your 4 Oct guard answer still hold?

**(a), yes:** "a is fine now. later we will upgrade this setup with jev". The file guard, deny rules for never-edited
files and the delete check; a shell write to his paths (TB-029's case included) is caught after the fact by rung 0 and
the Desk (§7). The upgrade waits for the Jev trial (7-21 Oct) and his decision on
`studio/decisions/RESEARCH-2026-10-04-jev-guard.md`.

### B. When, and how wide, is the admin batch?

**Dropped:** "why do we need this. i dont think we do, seems to be overenginnering." No managed settings file, no
version pin, no admin step (Q10 and Q11 reversed). The walls are base's deny rules in each linked project, `check` warns
on an old Claude Code, and what only the managed file gave is given up as a tripwire (§7).

### C. How is Bonsai's own work proven until Bonsai can guard itself?

**(c).** The skeleton and steps 5.1-5.3 (about 103-160 h) run on Go tests in WSL and on Windows, CI and a fresh
verifier; from 5.4 Bonsai's tasks climb `bonsai ladder` on a `bonsai.yaml` in its own repo, with a pre-release `bonsai`
he installs (§14, §17 step 8).

### D. The skeleton's budget against your 45 h stop line

**(a).** The log and asks move to step 5.2; the skeleton is 30-46 AI hours with a 60 h stop line (1.3 times 46; after
his format review 30-47 h and 61 h, §14);
contract §15.1 step 4 now says "right after the gate", on his word (§14).

## 20. Response to the 7 Oct review, recheck and format review

**The review** was answered finding by finding in a table here, cut after the recheck: it is in git history (`git show
131f6f8:docs/specs/2026-10-07-bonsai.md`), and the recheck's last table gives each finding's state. Its B1 recovery and
M4 reason no longer hold (the recheck's N1 and N3).

**The recheck** (its Claude Code facts not fetched again) and **his answers carried through** (an Opus pass the same
evening) were listed here item by item; both lists are in git history (`git show 8ea360b:docs/specs/2026-10-07-bonsai.md`,
§20), and every decision they made stands in its own section (N1-N3 and N5's second batch went with the managed file,
question B). Step 5 was then 123-195 h, and Bonsai 1.0 with the skeleton 153-241.

**Rohan's words after the recheck** (7 Oct): his answers are in §18. Every command now runs unattended, with `--json`
on all but `hook` and an `error` object for his review (§3, §4, §16 row 29). Checks that are minutes for him and hours
for an agent moved to him (§14, §17 steps 6 and 7).

**Rohan's own review of every format, round 1** (7 Oct, `studio/decisions/REVIEW-2026-10-07-formats.md`; an Opus pass,
then the second recheck, `docs/specs/2026-10-07-bonsai-recheck-2.md`, and its fixes): each change and its reason is in
§16 (rows marked "Changed 7 Oct", rows 30-37, the contract-change list), §18 and contract §21; his second look is round
2 of the review file. **Hours:** the skeleton went 36-53 h (with the log and asks), 33-50 (his checks moved to him),
30-46 (D), 30-47 (the format review: part 3 +1-2, part 4 -1), stop line 61 h; step 5 139-218 (+16-23: 5.1 +4-7, 5.2
+3-4, 5.3 +1, 5.4 +1, 5.5 +6-9, 5.6 +1); with the skeleton 169-265 AI hours; the studio's link 9-13 and contract §15.2
48-77. Earlier entries in this section keep their own day's figures.
