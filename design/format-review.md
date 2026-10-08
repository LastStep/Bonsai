---
id: REVIEW-2026-10-07-formats
date: 2026-10-07
status: confirmed
owner: orchestrator
---

# Formats for Rohan's review, 7 Oct

> **Moved here from the studio repo on 8 Oct 2026 on Rohan's word, whole; the studio keeps a pointer.** Rohan's own
> review of every format, rounds 1 and 2, confirmed on 7 Oct: the formats are final. Section numbers cite the Bonsai
> spec (`design/bonsai-spec.md`) and the contract (`design/contract.md`); other paths are the studio's.

Every format below is one you confirm before anything is built on it. Six rounds; answer each format in its **Rohan:** line, or the whole round in its **Group answer:** line. Sources: Bonsai spec §16, the contract.

## Group 1: Ground rules and Bonsai's own project files

### 1.1 The rules every format follows (contract §2)

**What it is:** Five rules shared by all formats. One owner (Bonsai) and a version line in each file. Changes inside a version only add things. One strict YAML grammar (format 1). Today's files read forever as "format 0". Only listed fields reach the VPS.

**Why:** Without them each format drifts on its own, an old file breaks a new reader, and a secret can leak to the VPS by accident.

**Example:**
```
format: bonsai.task/1
```
That one line is the version line. A reader that sees `bonsai.task/2` says "format too new" and reads nothing else.

**Cost:** Writers add one line per file. A new major version (`/2`) is your call, and needs a reader for both versions for at least 60 days and two minor releases.

**Worth a look:**
- Is 60 days of reading the old version enough?
- Format 1 refuses sloppy YAML that today's files get away with. Fine for new files only (old ones stay as they are)?

**Rohan (7 Oct, his own read of this file):** The format versioning is fine. We don't need to overcomplicate it more with things like keeping both versions and such. That would be a problem to be handled when we update the formats for the first time.

### 1.2 The Bonsai home and the workspace id (contract §3)

**What it is:** The home is one folder on the machine for what must never go into a project (log, asks, settings, a secret salt). The workspace id is a random name for a project, written in its `bonsai.yaml`.

**Why:** Secrets and logs stay out of git. The id lets every clone and worktree of a project count as one project.

**Example:**
```
<home>/workspaces/r-<16 hex>/
  settings.json   labels/   log/   asks/
```
```
id: ws-<26 base32>
name: trinetra
```

**Cost:** `bonsai init` writes the id once. Changing the layout later means moving files on every machine.

**Worth a look:**
- The home is `~/.bonsai` (Windows: `%USERPROFILE%\.bonsai`). Happy with the name and place?
- A changed id is flagged, and the bridge stops forwarding until you register again. Strict enough?

**Rohan (7 Oct, his own read of this file):** is this how official software do something like this? what could be the drawbacks here? and if we are going with this, then we need to be clear to the user at the start.

### 1.3 The home's additions (spec §10, §3, §8)

**What it is:** Four more things in the home: `personal/` (notes about you), `locks/` (one ladder at a time), `install.json` (path, version and SHA-256 of the installed `bonsai`), and each workspace's `outbox/`. `--adopt` also copies a replaced file into `cache/`.

**Why:** Without them there is no home for your personal memory, no way to stop two ladders colliding, no way to spot a swapped binary.

**Example:**
```
<home>/
  personal/INDEX.md
  locks/ladder.lock
  install.json
  workspaces/r-<16 hex>/outbox/
```

**Cost:** Bonsai writes them. Adding a folder later costs nothing; renaming one costs a migration on each machine.

**Worth a look:** Nothing open.

**Rohan (7 Oct, his own read of this file):** Agree

### 1.4 `bonsai.yaml`, the project's settings (`bonsai.workspace/1`, spec §6)

**What it is:** One committed file in each project: which packs it uses, where its tasks, runs and notes live, which paths are protected, and its ladder. It is the Bonsai half of today's `game.yaml`.

**Why:** Without it Bonsai hard-codes folders like `studio/`, and the studio could not serve a second project.

**Example:**
```
format: bonsai.workspace/1
id: ws-<26 base32>
name: trinetra
packs:
  - id: workflow
    source: "https://github.com/LastStep/bonsai-workflow.git"
    ref: v1.0.0
documents:
  task: studio/tasks
protected: [".claude/**", "bonsai.yaml"]
person_only: [".claude/**", "bonsai.yaml"]
boot_budget_kb: 24
ladder_floor: [0, 1, 2, 3, 4]
```
(shortened; the spec's full example also lists run, state, answers, memory, protocols, never_edit, ladder, ratchets)

**Cost:** You change it by hand, or an agent with a granted task does (it is person-only). It is read on every command. New fields are optional additions.

**Worth a look:**
- `person_only` is the short list you chose (the guard files, `bonsai.yaml`, the lock, `.github/**`, protocols). Right list?
- `boot_budget_kb: 24`: the most always-on text a session may start with. Today it is about 21 KB. Right size?

**Rohan (7 Oct, his own read of this file):** Make sure to document it by comments properly, so its easy to understand while reading. Don't use boot_budget_kb. Person list is fine.

### 1.5 The instruction block in `CLAUDE.md`, and what loads always (spec §5, §6)

**What it is:** A marked block of at most 40 lines that Bonsai keeps inside `CLAUDE.md`: the workspace and its packs, imports of the always-on protocol files and of the memory index, and the label definitions agents see. The rest of `CLAUDE.md` stays yours. Only `session-start` loads always; other protocols are skills, read when a role asks.

**Why:** It is how every session and subagent learns the house rules without reading everything.

**Example (what the block holds, in order):**
```
1. One line: workspace trinetra, packs base and workflow
2. Import: studio/protocols/session-start.md
3. Import: the memory index
4. The label definitions agents see (about 40 tokens each)
```
(the spec fixes these contents and the 40-line cap, not the exact marker text)

**Cost:** `bonsai update` rewrites only inside the block. Every always-on byte costs tokens in every session.

**Worth a look:**
- Is "only session-start always, the rest on demand" the right split?
- 40 lines for the block: enough?

**Rohan (7 Oct, his own read of this file):** Agree

**Group 1 answer:**

## Group 2: Tasks and status

### 2.1 The task (`bonsai.task/1`, contract §4)

**What it is:** One markdown file per unit of work: id, title, status, lane, what "done" means, what it depends on, dates, and a `labels` map for everything else. Eight statuses: todo, plan, approved, running, verify, done, blocked, cut.

**Why:** Every check, guard and the Desk reads this. Without one shape, each tool guesses.

**Example:**
```
format: bonsai.task/1
id: T-0072
title: The bridge serves Windows-side projects
status: done
lane: full
done_when:
  - "Rungs 0-4 green in one ladder run in WSL"
depends_on: []
blocked_by:
created: 2026-10-07
started: 2026-10-07
finished: 2026-10-07
labels: {}
```
(`title` and `done_when` here are paraphrased for the example)

**Cost:** Agents write it; the studio moves its status. The status list never grows inside a version. Task ids are 4 to 6 digits. Open tasks move to format 1 in one commit; closed ones stay as they are.

**Worth a look:**
- Are the eight statuses final? (You chose to fix them in Bonsai's core.)
- The `finished` and `started` dates are core fields. Keep?

**Rohan (7 Oct, his own read of this file):** Agree

### 2.2 Labels and their definitions (`bonsai.labels/1`, grants, contract §5)

**What it is:** A label is a named extra value on a document, like `bonsai.branch`. A definition file says what each means, its kind (choice, text, number, list), and who may write it: `agent` or `outside` (only a program, never an agent). Labels with `grants: true` widen what an agent may do; only `bonsai.allows` does. A task that needs a person-only path lists it in `bonsai.wants`; your approve tap copies it into `bonsai.allows`.

**Why:** Closes the hole where an agent grants itself protected paths by editing its own task (TB-029). Keeps the studio's facts out of Bonsai's own fields.

**Example:**
```
format: bonsai.labels/1
namespace: trinetra
version: 1
labels:
  - name: trinetra.cost
    kind: number
    kinds: [task, run]
    set_by: outside
    description: "Dollars this work cost. Agents never write it."
```

**Cost:** The studio keeps its own definitions and attaches them to the machine. A new label is an addition. Redefining one is not allowed.

**Worth a look:**
- Question B's answer stands: only a short list of paths needs your tap. Still right?
- `trinetra.cost` exists but nothing writes it yet. Leave it?

**Rohan (7 Oct, his own read of this file):** Agree. Where do the definitions for something like "kinds" exist? What i'm asking is how do we know which all types are possible in any place we are using something pre-defined.

### 2.3 The estimate labels (spec §2)

**What it is:** `workflow.estimate_h` on a task is the planned hours (set from the plan). `trinetra.actual_h` is the hours measured from session records, which the Desk charts.

**Why:** Without them your "estimate against actual" chart has nothing to read, and hour stop lines cannot be checked.

**Example:**
```
labels:
  workflow.estimate_h: 30
```
(30 is a made-up number)
(`trinetra.actual_h` is set by the studio, outside agent sessions)

**Cost:** The planner writes the estimate once. The studio writes the actual. Both are additions to existing label sets.

**Worth a look:** Is hours the right unit (AI session hours, including verification)?

**Rohan (7 Oct, his own read of this file):** Agree

### 2.4 Lanes (`bonsai.lanes/1`, contract §6)

**What it is:** Light, full and director. Bonsai's base knows the idea of a lane; your workflow pack defines the three. Base understands two rules only: `approve_first` (a person approves before work) and `close` (`person` or `agent` closes). Stricter means more approval, never less.

**Why:** The stop rule (director lane: write options, stop) must hold without the studio attached, and the Desk needs to know which closes are yours.

**Example:**
```
format: bonsai.lanes/1
lanes:
  - name: full
    approve_first: true
    close: person
    description: "A feature or a contract change. A person approves the plan before code and closes the task."
```

**Cost:** A lane is set when a task is created. Looser later only by a person's move. The two rules are a closed list.

**Worth a look:**
- Only two rules exist in base. Is that enough, or should a third (say, "needs a fresh verifier") exist now?
- An unknown lane counts as the strictest. Right default?

**Rohan (7 Oct, his own read of this file):** Agree

### 2.5 Status moves and Desk taps (contract §10)

**What it is:** The studio is the one writer of status moves. Two doors: the bridge (Desk taps) and a command (`trinetra move`, for agents and chat). Taps are approve, accept an ADR, done, cut, note, bless, grant, snooze. Your moves are: approve, done in a person-close lane, cut, and loosening a lane.

**Why:** One checked write path stops agents approving their own work. Bonsai only gives the structure.

**Example:**
```
trinetra move T-0072 running --from approved
```
The commit carries `Studio-Action: <uuid>` and `Studio-Via: agent`.

**Cost:** This is the studio's code, not a Bonsai file. Adding a tap later is an addition. A tap from the PC cannot prove it was you, so the Desk shows which device tapped (question D).

**Worth a look:**
- Are these eight taps the right set?
- Chat approvals count, marked as chat (question C). Still what you want?

**Rohan (7 Oct, his own read of this file):** What is this for exactly? like are we doing the work through cli commands here?

**Group 2 answer:**

## Group 3: Records people read

### 3.1 The run report (`bonsai.run/1`, contract §7.1)

**What it is:** The log of one agent session on one task, started first and added to as it goes. Frontmatter names the task, role, model, times, outcome and commits.

**Why:** If a session dies halfway, this file is all that survives. It is also what a verifier reads.

**Example:**
```
format: bonsai.run/1
id: R-2026-10-07-T-0072
task: T-0072
role: builder
model: opus
started: 2026-10-07 14:08
finished: 2026-10-07 15:25
outcome: needs-person
commits: ['4e5f515', 'e551c86']
labels: {}
```

**Cost:** Written by the builder once per session. Outcomes are a closed list: running, verify, needs-verifier, needs-person, merged, blocked, abandoned. The old `needs-rohan` reads as `needs-person`.

**Worth a look:** Is that outcome list complete? (It cannot grow inside version 1.)

**Rohan (7 Oct, his own read of this file):** Agree

### 3.2 The run report's "Sessions" table (spec §14)

**What it is:** A table in the run report body: one row per session with role, start, end and minutes. It adds the hours that the walking skeleton's stop line counts.

**Why:** Without it the 60-hour stop line cannot be checked against anything.

**Example:**
```
| role | start | end | minutes |
|---|---|---|---|
| builder | 2026-10-07 14:08 | 2026-10-07 15:25 | 77 |
```

**Cost:** Each builder and verifier adds a row. It is checked against the Desk's session records.

**Worth a look:** Are minutes enough, or do you want the model per row too?

**Rohan (7 Oct, his own read of this file):** Agree. Where would this type of data aggregation file live? like reports will have their own folder, same for logs, runs etc. but for files where we aggregate some kind of data, like session table, state etc, they should all live in one specified folder decided by bonsai.

### 3.3 STATE (`bonsai.state/1`, contract §7.2)

**What it is:** The one page that says where a project stands. Frontmatter: `format`, `updated`, `updated_by`, `labels`. The body is free.

**Why:** Every new session reads it first. It replaces a human handover.

**Example:**
```
format: bonsai.state/1
updated: 2026-10-07
updated_by: orchestrator (opus)
labels:
  workflow.milestone: V2
```

**Cost:** Rewritten (never appended) by the orchestrator whenever things change. Frontmatter fields are fixed; the body is not checked.

**Worth a look:** Today's STATE also names the `project`. In format 1 the workspace is the project, so that line goes. Fine?

**Rohan (7 Oct, his own read of this file):** Agree

### 3.4 Document kinds (contract §7.3)

**What it is:** A project declares each kind of document it keeps (task, plan, decision, spec ...): its folder, id pattern, statuses, which status moves a person makes and which an agent makes. Bonsai declares task, run, state, answers and memory; your workflow pack declares the rest.

**Why:** Without it the studio hard-codes `studio/` and cannot serve another project's layout.

**Example:**
```
kind: plan
path: studio/plans
id: ^P-T-\d{4,6}$
statuses: [draft, approved, done]
person: [[draft, approved]]
agent: [[approved, done]]
task_field: task
```
(the spec shows this in JSON; shown here as lines, same fields, shortened)

**Cost:** The workflow pack writes the declarations. A new kind is an addition. Changing a kind's statuses changes who can move what.

**Worth a look:** Plans move draft, approved, done, and only you approve. Right?

**Rohan (7 Oct, his own read of this file):** Agree. Oh does this clear my 2.2 query?

### 3.5 Memory (`bonsai.memory/1`, spec §10)

**What it is:** One note per fact, in the repo, plus a short `INDEX.md`. Notes carry `id: M-<slug>`, `title`, `kind` (project, feedback, reference), `updated`, `source`. Your personal notes live on the machine, never in a project, never sent anywhere.

**Why:** Gives all projects and agents one memory convention, and turns Claude's own auto memory off so two do not fight.

**Example:**
```
format: bonsai.memory/1
id: M-fewer-verifications
title: Fewer verifications
kind: feedback
updated: 2026-10-06
source: Rohan, 6 Oct
labels: {}
```
(`id`, `title` and the name are from your memory index; the spec fixes only the field names)

**Cost:** Agents write notes (4 KB at most each). The index has a 120-line, 12 KB budget, but the always-on budget (24 KB) wins; here that leaves about 6 KB. `bonsai check` enforces it.

**Worth a look:**
- Is about 6 KB of index in this repo enough?
- Personal layer is canonical in WSL and copied to Windows. OK?

**Rohan (7 Oct, his own read of this file):** Agree

**Group 3 answer:**

## Group 4: Machine records

### 4.1 The log record (`bonsai.log/1`, contract §8)

**What it is:** One JSON line per thing an agent did (a tool call, a guard decision, a ladder result). Cleaned of secrets as it is written, kept on the machine for 30 days, never committed. The studio forwards it.

**Why:** It is the evidence behind the Desk's activity and the "who did what" you ask about.

**Example (shortened):**
```
{"format":"bonsai.log/1","id":"<uuid>",
 "at":"2026-10-07T14:08:00.000Z",
 "workspace":"ws-<26 base32>","agent":"claude-code",
 "event":"tool_end","task":"T-0072","role":"builder",
 "tool":"Edit","category":"Edit","ok":true}
```

**Cost:** The recorder writes it on every tool call. Event names and categories are open lists (readers show unknown ones as `other`); the fields only grow. Lines are at most 2,048 bytes.

**Worth a look:**
- 30 days of keeping (`log_keep_days`): right?
- Free text reaches the VPS only in `target` and `text`, redacted. OK?

**Rohan (7 Oct, his own read of this file):** Agree

### 4.2 The binary's path and SHA-256 in the log (spec §3, §8)

**What it is:** When a session starts, `bonsai hook start` logs the path and SHA-256 of the `bonsai` program that ran.

**Why:** Four old `bonsai` copies are on your PATH. If one shadows the real one, the guard is not the guard. This is how the Desk notices.

**Example:**
```
"event":"session_start", "agent":"claude-code"
(plus the binary's path and its SHA-256, as new fields)
```
(the spec does not fix these two field names; that is the open point)

**Cost:** One more field on each session start. Adding it now is cheaper than later.

**Worth a look:** The two field names are not yet fixed in the contract. Want to see them before building?

**Rohan (7 Oct, his own read of this file):** Agree

### 4.3 The ask record (`bonsai.ask/1`, contract §9)

**What it is:** A question an agent files for you, and your answer, as JSON lines in the home. Types: Answer, Decide, Look, Play (agents) and Bless (the ladder only). Operations: `file`, `resolve`, `answer`.

**Why:** It is what shows in Needs you on the Desk. An answer never grants a right.

**Example (shortened):**
```
{"format":"bonsai.ask/1","op":"file",
 "key":"agent:h-<12 hex>","workspace":"ws-<26 base32>",
 "source":"agent","type":"Decide","task":"T-0072",
 "title":"...","why":"...","options":["...","..."]}
```

**Cost:** Agents and the ladder file; you or the studio answer. Operations are a closed list. Limits: title 300, why 600, four options of 200. Changing any limit is an addition only.

**Worth a look:**
- An answer from the session that asked is refused. Right rule?
- Keys no longer name the project; the bridge adds it. Existing open asks show once as new cards at the switch. OK?

**Rohan (7 Oct, his own read of this file):** Agree

### 4.4 The ask outbox under the sandbox (spec §8)

**What it is:** When the sandbox is on, a sandboxed session may only drop question requests into an `outbox/` folder; the recorder moves them into `asks/`.

**Why:** Otherwise `bonsai ask` stops working once the sandbox is turned on.

**Example:**
```
<home>/workspaces/r-<16 hex>/outbox/
  (file requests only, op: file)
```

**Cost:** Nothing now: the sandbox is off in WSL. Choosing an MCP ask tool instead would be more code for a small risk.

**Worth a look:** Outbox folder, or an ask tool? The spec picks the outbox. Keep?

**Rohan (7 Oct, his own read of this file):** I don't understand what this is. Fully explain to me the whole idea and flow and connectivity for all things in group 4.

### 4.5 The `error` object in every command's `--json` (spec §3)

**What it is:** When a command refuses or fails, its JSON has `error` with `code` (a fixed word), `message`, and `next`: what to do, and whether an agent or a person does it.

**Why:** You asked for commands any agent can use unattended. An error that names the next step lets it carry on instead of guessing.

**Example:**
```
{"error":{"code":"<fixed word>",
  "message":"update needs --yes",
  "next":"run: bonsai update --yes (a person's step)"}}
```
(shape only: the spec does not list the code words, so the values are made up)

**Cost:** Every command must fill it. The code words become a closed list for agents to read. Adding a word is an addition.

**Worth a look:**
- Should `next` say "agent" or "person" as its own field, not inside the sentence?
- Want the list of code words before it is built?

**Rohan (7 Oct, his own read of this file):** Agree

**Group 4 answer:**

## Group 5: The ladder, the active task and the guards' settings

### 5.1 The ladder result (`bonsai.ladder/1`, contract §11)

**What it is:** The proof that a task's checks passed, at a commit: one JSON file per task, written by Bonsai's runner, kept out of git.

**Why:** The stop gate, verifiers and the Desk trust this file instead of an agent's word.

**Example (shortened):**
```
{"format":"bonsai.ladder/1",
 "task":"T-0072",
 "workspace":"ws-<26 base32>",
 "mode":"local",
 "git":{"sha":"49e9867","branch":"main","dirty":false},
 "requested":[0,1,2,3,4],
 "green":true}
```
(the sha is from T-0072's run report; the real file also holds `rungs`, `skipped`, `leftovers`, `proof` and more)

**Cost:** Written every ladder run. Every field is always present (`null` or `[]` if unused). A `mode` other than `local` (CI) is never a task's proof. After writing, a fingerprint of the file goes into the log.

**Worth a look:**
- `requested` is the project's floor plus the task's own rungs, so a task cannot skip tests. Keep?
- Old results with no `mode` read as `local`. Fine?

**Rohan (7 Oct, his own read of this file):** Agree. Should be cleaned if older than 7 days. But for this actually we might need a fresh document in bonsai which tells how to manage all these ai generated files (runs, logs, ladder results etc), with options of cleaning X days old, rotating and keeping fresh Y records etc.

### 5.2 The ladder rung's `tests` and `base_setup`, the new-tests rule, and the git-integrity rung (spec §9)

**What it is:** A rung may name how to find new tests (`tests`: TAP, Go test JSON or JUnit) and how to prepare the old code (`base_setup`). When a test count rises, the runner runs the new tests on the base commit and records `new_tests: {base, count, failed_on_base}`. A third rung kind, git-integrity, flags tampering (changed test files, hidden-file tricks, rewritten base). Not required at first.

**Why:** Closes the cheat of adding tests that always pass. This is your 7 Oct rule.

**Example:**
```
new_tests:
  base: "<sha>"
  count: 3
  failed_on_base: 3
```

**Cost:** Slower ladder runs when a count rises (a temporary worktree, one more run). A setup or load failure reads as "check not run", never "failed".

**Worth a look:**
- A rise is offered for your Bless tap whatever the numbers; the card shows "proves nothing new". Or should the offer be withheld then?
- Git-integrity informs the verifier, not required. Make it required later?

**Rohan (7 Oct, his own read of this file):** Why do we need this. explain with some examples.

### 5.3 The active task (contract §13)

**What it is:** One rule for "which task's permissions apply right now", used by the guard, rung 0, the stop gate and status. Steps: the task named by `--task` or the environment; else the single task reading `running`; else none. Tasks are read from the main checkout. None means no grants and refusals that say why.

**Why:** Today two readers disagree (T-0011), so rung 0 can refuse what the guard allowed.

**Example:**
```
"active_task": {"id": "T-0072",
  "how": "named", "why": null}
```

**Cost:** One definition, tested by the same fixtures everywhere. Changes make it refuse rather than guess: a task file that does not parse means none.

**Worth a look:** Two tasks both `running` and nothing named gives "none". Strict enough, or too strict for parallel builders?

**Rohan (7 Oct, his own read of this file):** This looks like it should be some kind of task table, similar to 3.2

### 5.4 Bonsai's entries in `.claude/settings.json` (spec §5, §7)

**What it is:** The lines Bonsai writes into the project's settings: hook lines (guard, stop, start, record), deny rules (never-edit files and the walls round key and credential files), the plugin marketplace named per workspace and commit, `autoMemoryEnabled: false` and `disableAllHooks: false`.

**Why:** These make the guards run and keep secrets out of reach. They are tripwires, not walls: the same user can still reach around them.

**Example:**
```
"hooks": PreToolUse  ->  bonsai hook guard || exit 2
"deny":  "Read(~/.ssh/**)", "Read(~/.bonsai/salt)",
         "Edit(~/go/bin/**)"
"autoMemoryEnabled": false
"disableAllHooks": false
```
(shown as plain lines; the real file is JSON)

**Cost:** `bonsai update` rewrites only its own lines. A new deny rule is an addition. The Windows list is shorter than the WSL one.

**Worth a look:**
- Are the walls the right set, and are any too wide for your own sessions?
- Updating a pack moves the marketplace name (it hashes the locked commits). Fine?

**Rohan (7 Oct, his own read of this file):** Agree. Make sure it is verbose so the user knows exactly what all things will change

**Group 5 answer:**

## Group 6: Packs, the lock, status and the command words

### 6.1 The pack (`bonsai.pack/1` and the folder layout, spec §5, §7)

**What it is:** A pack is a Git repo that is also a valid Claude Code plugin, plus a `bonsai/` folder: `pack.yaml` (id, version, `needs`, what the engine writes), labels, lanes, a block, and files. `needs` includes the oldest Claude Code it works with. The plugin's `plugin.json` and the marketplace entry never carry a `version`.

**Why:** One shape for roles, skills and rules, shared by all projects. With a `version` there, Claude Code would keep old copies.

**Example:**
```
<pack>/
  .claude-plugin/plugin.json
  agents/   skills/
  bonsai/pack.yaml   labels.yaml   lanes.yaml
  bonsai/block.md    files/
```
Two packs: `base` (public) and `workflow` (private, `LastStep/bonsai-workflow`).

**Cost:** You maintain the workflow pack. `bonsai check` refuses a `version` in the wrong place. New files are additions.

**Worth a look:**
- Is `workflow` private, with `base` public, the split you want?
- Pack versions are the Git tag, not a number in the plugin. Fine?

**Rohan (7 Oct, his own read of this file):** why does anything have to be private here? and would the pack be a different git repo in itself? i want proper ci/cd for it as well, so the process of adding to the marketplace and updating is streamlined

### 6.2 The lock (`bonsai.lock/1`, contract §14)

**What it is:** A committed file `bonsai.lock.json` listing exactly which pack versions and files a project holds, with a hash for each.

**Why:** An update writes everything or nothing, and never overwrites an edit you made.

**Example:**
```
{"format":"bonsai.lock/1","written_by":"1.0.0",
 "packs":[{"id":"base","version":"1.0.0",
   "commit":"<sha>","sha256":"<content hash>"}],
 "files":{".claude/settings.json":
   {"kind":"keys","pack":"base","sha256":"<sha256>"}},
 "format0":{}}
```

**Cost:** Bonsai's engine writes it on each update; it is person-only. Keys are sorted and paths use forward slashes.

**Worth a look:** The file is `bonsai.lock.json` at the repo root, beside `bonsai.yaml`. Fine name and place?

**Rohan (7 Oct, his own read of this file):** Agree. Would it not be better to have it in .bonsai folder in the repo root? as we can possibly add more things there

### 6.3 The lock's `declares` and its file name (spec §6)

**What it is:** For each pack, the lock also copies in what it declared: lanes, document kinds, label definitions and protected paths.

**Why:** CI can then run `bonsai check` without fetching any pack, so a private pack needs no secret in CI.

**Example:**
```
"packs":[{"id":"workflow","version":"1.0.0",
  "declares":{"lanes":["..."],"labels":["..."]}}]
```
(shape only: the spec lists the four kinds, not the inner layout)

**Cost:** The lock gets bigger. The engine fills it on update. A new kind of declaration is an addition.

**Worth a look:** Nothing open.

**Rohan (7 Oct, his own read of this file):** Agree

### 6.4 `bonsai status --json` (`bonsai.status/1`, contract §12)

**What it is:** A printed (never stored) snapshot of one workspace: its id, packs, document kinds, lanes, person-only list, active task, needs and problems. Cheap and offline by default; `--full` adds checks.

**Why:** The studio and Bonsai's screens read it so they never build paths or guess rules.

**Example (shortened):**
```
{"format":"bonsai.status/1","bonsai":"1.0.0",
 "workspace":{"id":"ws-<26 base32>","name":"trinetra"},
 "status_writes":"command",
 "person_only":[".claude/**","bonsai.yaml"],
 "active_task":{"id":"T-0072","how":"named","why":null},
 "problems":[]}
```

**Cost:** Written by `bonsai`, read often. New fields are additions. Exit 3 if the workspace cannot be read at all.

**Worth a look:** The checkout's absolute path is printed here but never forwarded. OK?

**Rohan (7 Oct, his own read of this file):** Agree

### 6.5 `status --json`'s additions (spec §4-§7)

**What it is:** More to report: a pack not installed on this machine, plugin drift, a swapped `bonsai` binary, and the Claude Code floor, all as `needs` entries or `problems`; plus `--line`, the statusline's workspace half.

**Why:** The Desk can then flag a changed guard or a machine that is not set up, before you notice by accident.

**Example:**
```
"needs":[{"kind":"tool","name":"claude-code",
          "version":">=<floor>"}]
```
(`<floor>` is set by each release; the spec gives no number)

**Cost:** A few more fields; `--full` also asks the plugin list. No pin on Claude Code: it only warns when too old.

**Worth a look:** Warn only, or refuse to run, when Claude Code is older than the floor?

**Rohan (7 Oct, his own read of this file):** Agree, and warn

### 6.6 The command words and exit codes (spec §3, §4)

**What it is:** Fourteen words: `init`, `update`, `unlink`, `status`, `check`, `hook`, `ladder`, `ask`, `answer`, `asks`, `logs`, `log append`, `settings`, `labels`. Exit codes: 0 ok, 1 findings, 2 bad input, 3 runtime, 4 wrong state or no `--yes`, 5 conflict. Hooks exit 0 (allow) or 2 (block). Every command but `hook` has `--json` and runs unattended.

**Why:** Agents and CI read the exit code, so it must never change meaning.

**Example:**
```
bonsai update --diff        # preview
bonsai update --yes         # apply
(no --yes, no terminal: preview, exit 4)
```

**Cost:** Cap of fifteen words for 1.0. A new word means one out, one in. Exit codes are fixed.

**Worth a look:**
- Is fourteen words right?
- `settings set` and `labels` are refused inside an agent session. Keep?

**Rohan (7 Oct, his own read of this file):** Looks good, explain each command word usage to me, and the scenario it will be used in

### 6.7 Changes to the settled contract's text (spec §16, §5, contract §15.1)

**What it is:** Two edits to text you already settled. The §10.3 erratum: pickup starts `workflow:builder` (plugin roles carry the pack's name as a prefix). §15.1 step 4: the log and asks formats arrive with Bonsai's program (step 5.2), not in the skeleton.

**Why:** The old text would send the wrong agent name, and would build the log twice.

**Example:**
```
claude --agent workflow:builder --bg
```
(was `--agent builder`; skeleton check 8 confirms it starts)

**Cost:** Already changed on your word (question D (a), 7 Oct). The studio's old `.claude/agents/` copies go when it links.

**Worth a look:** Nothing open.

**Rohan (7 Oct, his own read of this file):** Agree. teach me where would the bonsi agent definitions live

**Group 6 answer:**

## Round 1 decisions (7 Oct, after his read)

Asked in chat after his answers above, one at a time:
- **One folder for Bonsai's files (3.2, 5.3, 6.2):** `bonsai.yaml` stays at the root (the one file a person edits); the
  lock, STATE and generated tables (tasks with the active task; sessions with hours per task, built from the log) live
  in `.bonsai/`. Bonsai rebuilds the tables from the task files and the log, which stay the truth; `bonsai check` fails
  on a stale table.
- **Where the log, asks and ladder results live:** in `.bonsai/local/`, inside the project and never committed (Bonsai
  writes `.bonsai/.gitignore` itself; `check` and rung 0 refuse anything staged from `local/`); worktrees share the main
  checkout's through git. Machine-only things stay in `~/.bonsai`: the salt, machine settings, personal memory, the
  install record, the pack cache. The outbox goes. He asked why they lived outside ("that might be an issue for a lot of
  projects"): moves, containers and cloud machines, and two places per project.
- **The workflow pack's name:** stays `workflow` (offered `crew`, `grove`, `agent-workspace`; his own idea was
  BonsaiAgentWorkspace).
- **The workflow pack is public** (6.1: "why does anything have to be private here?"), with a pack template carrying
  CI/CD (validate as a plugin, `bonsai check`, tests, release on a tag).

## Round 2: what changed, for your second look

Your round 1 answers are now in the Bonsai spec and the contract (spec §16 lists every change; contract §21). Below,
each format that changed or is new. Answer each in its **Rohan:** line.

### R2.1 One folder for Bonsai's files in a project: `.bonsai/` (3.2, 5.3, 6.2)

**What it is:** `bonsai.yaml` stays at the root: the one file you edit. Beside it, `.bonsai/` holds the lock, STATE and
two tables Bonsai builds. All committed except `local/` (R2.3). Only the lock is yours alone; agents still rewrite STATE.
**Why:** You asked for one folder, decided by Bonsai, for files that gather data.
**Example:**
```
bonsai.yaml
.bonsai/  lock.json  STATE.md  tasks.md  sessions.md  .gitignore  local/
```
**Worth a look:** STATE moves from `studio/STATE.md` to `.bonsai/STATE.md` when the studio links. Fine?

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.2 The two tables: tasks and sessions (3.2, 5.3)

**What it is:** `tasks.md`: every task, its status and lane, and which task is active. `sessions.md`: one row per
agent session and one per builder or verifier the orchestrator starts inside it (task, role, model, start, end,
minutes), then hours per task and role. A row's task is the task the agent was told, else the one task reading
`running` when it started, else "none". Bonsai rebuilds both from the task files and the log. Session rows are only
added, and a log file is kept until its rows are in, so the hours outlive the log's 30 days. Nobody edits them by
hand, and they grant nothing: the guards still read the task files.
**Why:** Your 3.2 and 5.3 answers. It replaces the run report's "Sessions" table.
**Example:**
```
| Task | Title | Status | Lane | Started | Finished |
| T-0072 | The bridge serves Windows-side projects | done | full | 2026-10-07 | 2026-10-07 |
```
**Cost:** The studio rebuilds them with every status move and commits them with it. Branches never touch them, so
worktrees never clash on them. Between moves they can lag a little; a lagging table is only a warning, never a
failed check.
**Worth a look:** Committed, so you can read them on GitHub. OK?

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.3 `.bonsai/local/`: the log, your questions and ladder results (round 1)

**What it is:** They now live inside the project, in `.bonsai/local/`, never committed. Bonsai writes the
`.gitignore`, and `check` refuses if anything from it is staged. Worktrees use the main checkout's. Only Bonsai writes
there: an agent's own edit is refused (a deny rule and the guard: a tripwire), while the `bonsai ask` and
`bonsai ladder` an agent runs write through Bonsai.
**Why:** Your round 1 answer: a moved folder, a container or a cloud machine would have left them behind.
**Example:** `.bonsai/local/log/`, `.bonsai/local/asks/`, `.bonsai/local/ladder/`
**Cost:** The outbox (4.4) is gone. A `git clean -x` (or `-X`) would delete them and `git stash --all` would move them
away, so the guard treats both as bulk deletes. The bridge reads each project's `.bonsai/local/` (Mimas's over
`/mnt/e`, once it links, through the Windows `bonsai.exe`). When the sandbox is turned on later, one more settings
line lets Bonsai's own commands write there; agents' edits stay refused.
**Worth a look:** Nothing open.

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.4 Your machine's home, and what `init` tells you (1.2)

**What it is:** `~/.bonsai` keeps only machine things: the secret salt, this machine's settings for each project, label
files the studio attached, your personal memory, the install record, the pack cache. `bonsai init` ends by saying
where everything is, in plain words; `bonsai status` shows the same.
**Example:**
```
.bonsai/local/        the log, questions for you and their answers, ladder results. Never committed.
~/.bonsai             Bonsai's home: ... None of it enters git.
```
**Your question, is this how others do it:** yes. Git keeps its own data in `.git/` inside the project. Claude Code
keeps each project's data in your home, under the folder's path.
**Drawbacks:** a moved folder loses this machine's settings for it until the studio registers it again; Bonsai names
the old folder. A copied folder keeps the same id: a new project runs `bonsai init --new-id`; Bonsai warns when two
folders share one id.

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.5 `bonsai.yaml`, with a comment on every line (1.4)

**What it is:** `init` writes it with a comment on every line. `boot_budget_kb` is gone; the fixed limits stay (the
instruction block 40 lines, the memory index 120 lines and 12 KB). The person-only list is as you agreed. A new
`generated:` part sets how long files are kept (R2.6).
**Example:**
```
ref: v1.0.0                     # the release tag; change it and run bonsai update to take a new release
ladder_floor: [0, 1, 2, 3, 4]   # the rungs every task climbs, whatever the task lists
```
**Worth a look:** If you delete the comments, nothing breaks. OK?

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.6 Generated files: how long each is kept (5.1)

**What it is:** One Bonsai page (a skill in the base pack) explains every file AI work generates. `generated:` in
`bonsai.yaml` sets, per kind, "clean after X days" or "keep only the newest Y".
**Defaults:** log 30 days, never before its session rows are in; ladder results 7 days, never an open task's; questions
kept; session rows kept; run reports kept (committed history: Bonsai never deletes them itself, it only lists old ones
for you).
**Example:**
```
ladder:
  keep_days: 7     # the result of a task not done or cut is never cleaned
```
**Cost:** Every file cleaned is a line in the log, so you see it on the Desk.
**Worth a look:** Are the defaults right?

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.7 Every template explains itself (your new rule)

**What it is:** Every template and pack file carries its own documentation: what it is for, when to use it, and every
field with its meaning, allowed values and an example. YAML files get `#` comments; markdown templates a table of fields.
**Tokens:** a task made from a template does not copy all that. It carries one short pointer on its `format:` line:
```
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
```
**Kept true:** `bonsai check --pack` fails a pack whose docs and fields disagree, in that pack's CI.
**Worth a look:** Is one pointer line per file enough?

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.8 Where each list of allowed values is written (2.2, 3.4)

**What it is:** Each list (task statuses, label kinds, lanes, ask types, exit codes, error codes...) is defined in one
place only: Bonsai's code, or a pack's declarations. `bonsai check --schema bonsai.task` prints a format with every
allowed value; `bonsai status --json` lists a project's lanes, document kinds and label sets (each label's values are
in its pack's definitions, copied into the lock).
**New:** one reference page in Bonsai's repo lists every list and where it is defined. It is generated from the code,
and CI fails if it falls out of date.
**Your 3.4 question:** document kinds answer part of 2.2; this page answers the rest.

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.9 The update preview names every settings line (5.4)

**What it is:** Before `bonsai update` (or `init`) changes `.claude/settings.json`, it lists every line it would add,
change or remove, each with one plain sentence on what it does.
**Example:**
```
add     deny  Read(~/.ssh/**)
        Agents cannot read your SSH keys (base: the walls).
```
**Cost:** Each pack must give a reason for every deny rule, or its CI fails.
**Worth a look:** Enough detail?

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.10 The workflow pack is public, and every pack gets CI (6.1)

**What it is:** `workflow` (`LastStep/bonsai-workflow`) is public; nothing in it is secret. Yes, `workflow` is its own
git repo (`base` lives inside Bonsai's repo, `packs/base/`). A pack template (a folder in Bonsai's repo) gives each pack CI built in: check it is a valid Claude Code plugin,
`bonsai check --pack`, its tests, and a release on each `vX.Y.Z` tag.
**Updating a project stays:** change the tag in `bonsai.yaml`, run `bonsai update`. There is no central marketplace to
publish to: Bonsai writes each project's own marketplace entry, so a tagged commit is the release.
**Worth a look:** The template as a folder in Bonsai's repo now, a GitHub template repo later if others make packs. OK?

**Rohan (7 Oct):** confirmed ("round 2 looks good").

### R2.11 Smaller changes

- **1.1:** the "read the old version for 60 days" rule is gone; it is decided when a format first changes.
- **6.5:** warn only. An old Claude Code gives a warning; nothing refuses to run.
- **6.7, where roles live:** in packs, as `agents/builder.md` in the workflow pack. Claude Code installs the pack and
  loads the role as `workflow:builder`. A linked project keeps no copies.
- **Hours per task (3.2):** they now count the builders and verifiers the orchestrator starts, not only whole
  sessions, since that is where most of the work runs. A run with no task to name counts as "none".
- **Hours:** the skeleton 30-47 h (was 30-46), stop line 61 h (was 60). Bonsai 1.0 after the gate 139-218 h (was
  123-195), mostly template docs, the pack template and the tables.

**Rohan (7 Oct):** confirmed ("round 2 looks good").

## Confirmed

Rohan, 7 Oct, 23:20 IST: "round 2 looks good". Every format in this file is confirmed, as round 1 answered them and
round 2 changed them. The formats are final; the readers task (contract §15), the formats landing in the studio repo,
slice 3 and Bonsai's skeleton may be planned on them.

