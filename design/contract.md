# The contract: the formats Bonsai and Trinetra both read

> **The master copy, moved here from the studio repo on 8 Oct 2026 on Rohan's word; the studio keeps a pointer.** It
> was written in the studio's repo: where the text says "this repo" it means the studio's, and the paths it cites
> (`studio/...`, `docs/specs/...`, `tools/...`, `studio-app/...`) are the studio's, not Bonsai's. Bonsai owns every
> format here (rule 1, §2.1), and now this text too. Notes marked **Changed 8 Oct (Rohan)** carry his decisions since;
> the rest is the text as he confirmed it. The Bonsai spec is `design/bonsai-spec.md`.

**Spec, revised 7 Oct 2026 after review and the 7 Oct split rule, then fixed after the 7 Oct recheck.** Drafted 6 Oct
by an Opus agent. Reviewed by a fresh Opus agent (`docs/specs/2026-10-06-contract-review.md`) and rechecked by another
(`docs/specs/2026-10-07-contract-recheck.md`); both are answered in §19 and §20. Revised under Rohan's rule **"Bonsai
gives a structure, the studio applies the actions"**, his 7 Oct picks (`studio/decisions/CHECKLIST-2026-10-07-workspace-split.md`),
and his two 7 Oct decisions: the Desk pause lifts when this contract's formats land in this repo, slice 3 first; and one
bridge for this PC, serving WSL and Windows. It turns the vision's §4 and its appendix A.3, A.5 and A.11
(`studio/decisions/VISION-2026-10-05-bonsai-and-trinetra.md`) into exact formats, under Rohan's answers in
`studio/decisions/WALKTHROUGH-2026-10-05-bonsai-and-studio.md`. Those bind, except where the 7 Oct rule replaces them:
decision 5's "Trinetra writes only labels" and Q3's "Bonsai applies it". The Bonsai spec builds on this one. The
multi-machine re-cut (`docs/specs/2026-10-06-multi-machine-recut.md`) meets it in §16. This spec also carries
**T-0011**, which Rohan cut into it on 6 Oct: one definition of "the active task" (§13). **Changed on Rohan's word
after his own review of every format** (7 Oct, `studio/decisions/REVIEW-2026-10-07-formats.md`): the sections are
listed in §21, each change in the Bonsai spec's §16.

**How to read it.** Rohan: §1 (about 4 minutes) and the four questions in §18 (about 8). Builders and verifiers: all
of it.

## 1. In plain words

**What the contract is.** Bonsai is the Docker inside each project. The studio (Trinetra) is the dashboard over all of
them, and since 7 Oct also the hands. Your rule: "Bonsai gives a structure, the studio applies the actions." Bonsai
defines the files: what a task looks like, its statuses, the lanes' rules, the log, the questions agents ask you. It
also keeps its guards, its recorder and its ladder. The studio does things with those files: it moves a task's status,
applies your approvals and protected-path grants, and starts agents. The contract is the set of files and outputs both
read. Bonsai's code never mentions the studio.

| Format | What it is | Written by | Read by | Lives |
|---|---|---|---|---|
| Task (§4) | Core fields plus labels (named values such as `bonsai.allows`) | Agents create it; the studio changes it | Bonsai, the studio, agents | The project's main checkout, committed |
| Label definitions (§5) | What each label means and which values it takes | Packs; the studio's own, attached on the machine | Agents, Bonsai's checks, the studio | A pack; the studio's in Bonsai's machine folder |
| Lanes (§6) | The lane slot in every task; the lanes a pack defines | Your workflow pack | Bonsai's checks, the studio | A pack |
| Run report, STATE, memory, other documents (§7) | The rest of a project's records | Agents | Bonsai, the studio | The project, committed; STATE in `.bonsai/` |
| Generated tables (§7.5) | The tasks table and the sessions table | `bonsai check --write` | People, agents, the studio | `.bonsai/`, committed |
| Log (§8) | What agents did, cleaned of secrets as it is written | Bonsai's recorder; the studio's events, through Bonsai | `bonsai logs`, the studio's forwarder | `.bonsai/local/` in the main checkout, never committed |
| Ask (§9) | A question for a person, and its answer | `bonsai ask`, `bonsai answer`, the ladder | The studio, agents | `.bonsai/local/`, never committed |
| Ladder result (§11) | A task's checks and whether they passed | Bonsai's ladder runner | The stop gate, verifiers, the studio | `.bonsai/local/`, never committed |
| `status --json` (§12) | One workspace at a glance | `bonsai status` | The studio, the assistant (Bonsai's screens dropped, Rohan 9 Oct) | Printed |
| Lock (§14) | Which pack versions and files a workspace holds | Bonsai | Bonsai, CI, verifiers; the studio through `status` | `.bonsai/lock.json`, committed |

Not a Bonsai format: **status moves and Desk taps** (§10). The studio applies them through one write path, and checks
each change against the task format.

**Five rules for every format** (§2):
1. One owner, Bonsai, and a version line in each file, such as `format: bonsai.task/1`.
2. Within a version, changes only add. A new version that old readers cannot read is your call.
3. Every program that reads the files refuses exactly the same broken or tricky files, tested on one shared set.
4. Only listed fields leave the machine for the VPS. The free text that does leave (documents, short log and question
   text) is cleaned of secrets first. No format holds a secret.
5. Today's files are "format 0": read forever, never rewritten.

**What changes for you.**
- **The Desk comes back first.** Your 7 Oct rule: the Desk pause lifts when these formats land in this repo, and slice 3
  (T-0018) goes first. Landing them is one task, with a plan for you to approve, of about 28-44 agent hours: roughly a
  week (§15).
- **Approving and closing.** From slice 3, in a project the studio manages, approving a plan, closing a full- or
  director-lane task and cutting a task are yours, and the studio applies them. Today most happen in chat: since 29 Sep,
  7 of 9 plan approvals and 7 of 8 full-lane closes came in chat, and the orchestrator edited the file. The guard will
  refuse those hand edits (a tripwire: a script could still get round it, and the Desk would flag it). **Question C**
  asks whether chat still counts (recommended: yes, marked "from chat").
- **A Desk approval starts the work.** It starts a builder in the background. Several approvals start several builders;
  slice 3's plan decides how they share the one ladder and who sweeps up after them.
- **A tap is not yet proof it was you.** Agents on the PC can tap the Desk as the PC (your 6 Oct choice). Once taps
  approve plans, grant paths and start builders, that matters more. **Question D** asks what to do (recommended: keep it,
  and the Desk shows which device tapped).
- **You can see what no tap backs.** The Desk marks any approval, close, cut, grant or raised test floor that no tap or
  chat approval of yours backs.
- **Agents stop editing task files by hand.** They run one studio command, and only in the main checkout, so the Desk
  sees every move, including a builder that is stuck. Nothing for you to do.
- **One rule for "the active task"** (T-0011, which you cut into this spec): the guards and the ladder agree which
  task's rules apply, and refuse rather than guess.

**What it asks of you:** four questions in §18. A: are task statuses fixed for every project? B: which protected-path
grants need your tap? C: do approvals you give in chat count? D: what may a tap from the PC do?

## 2. Rules every format follows

### 2.1 Owner and names

- Bonsai owns every format here. The studio never defines a Bonsai field; its facts are labels in its own namespace
  (`trinetra.*`, §5). Its own records (the tap journal, the act ids on its server, its `Studio-*` commit trailers) sit
  outside Bonsai's formats.
- Each format has one name, `bonsai.<name>`, and one major version: `format: bonsai.task/1` in YAML,
  `"format": "bonsai.log/1"` in JSON. Under Q4 (b) the recorder, the ladder runner and the engine are one Go program, so
  each format has one writer's version (vision A.3).
- Command words (`bonsai ask`, `bonsai log append`, `trinetra move` and so on) are working names. The Bonsai spec fixes
  Bonsai's; the studio's plans fix its own. The `trinetra` program named here (`trinetra move`, `trinetra lane`,
  `trinetra project add`) is one program.

### 2.2 Versions

- **Within a major, changes only add**: a new optional field, a new label, a new value in a list marked *open*. An
  unknown field in a known major: a writer (Bonsai's or the studio's) keeps it byte for byte, since both change only the
  lines they own; Bonsai and agents ignore it; the bridge drops it before the VPS (§2.6).
- **Closed lists** (task statuses, run outcomes, ask ops, lane rules) never grow inside a major. **Open lists** (log
  event names, log categories, ask types from packs) may grow; readers show an unknown value as `other`.
- **Every list, closed or open, is defined in exactly one place** (Rohan, 7 Oct, format review 2.2 and 3.4): Bonsai's
  built-in schemas (`bonsai check --schema <format>` prints a format with every allowed value), or a pack's
  declarations (lanes, document kinds, labels; copied into the lock, §14, and listed by `status --json`, §12). Bonsai's
  repo holds one reference page listing every list and where it is defined, generated from the code and checked in CI
  (Bonsai spec §6).
- **A newer major** than a reader knows: the reader says "format too new: bonsai.task/2" and parses nothing else. The
  Desk shows the file that way. Nobody guesses.
- **An older major**: how long readers keep reading it is decided when a format first gets a new major, not now (Rohan,
  7 Oct: "That would be a problem to be handled when we update the formats for the first time"). The studio reads
  format 0 forever.
- **A new major is Rohan's call** (vision §4), asked as a director-lane choice with its migration cost.
- **Each format has a JSON Schema** in every Bonsai release (`bonsai check --schema <format>`). A CI rung in Bonsai's
  repo fails on any schema change but additions (vision A.3). `status --json` says which majors this Bonsai reads and
  writes (§12, `formats`); the bridge reads a workspace only at majors it knows.
- **Present and missing fields.** A writer writes every field its version knows, in a fixed order, `null` (or `[]`)
  when it does not apply. A reader treats a missing field as `null`: an older writer of the same major. The writer's own
  version is in `status --json` (`bonsai`) and the lock (`written_by`).

### 2.3 Format 0: today's files

- A document with no `format:` key is **format 0**: today's task, run report and STATE shapes, read as they are.
- Format 0 is **frozen by git, not by a missing field** (vision A.3). Closed records (58 run reports here and 48 in
  Mimas on 5 Oct; 69 here on 7 Oct; the closed tasks, plans and decisions) are never rewritten. Open tasks move to
  format 1 in one commit per project (§15).
- **Which files `bonsai check` holds to format 0:** only Bonsai's own kinds (`task`, `run`, `state`; §7.3). When a
  workspace links, the lock lists each such file without `format:`, with the SHA-256 of its bytes after every line
  ending is made LF (§14, `format0`). Such a file must be on that list with that hash; a new or changed one is a finding
  ("give it `format: bonsai.task/1` first"). Pack kinds (plans, one-pagers, decisions, the bugs file) are never held to
  this. No commit is looked up, so shallow CI checkouts and Windows line endings both pass.

### 2.4 The YAML rules

Committed documents are markdown with YAML frontmatter; definitions are YAML files. **There are two sets of rules, one
per format.** A reader looks at the first top-level key: if it is `format:`, the file is read under that format's
rules; if there is no top-level `format:` key, under format 0's. A top-level `format:` anywhere but first is refused.

**Format 0 reads exactly as `tools/lib/yaml.mjs` reads today**, forever, with no new refusals: those files can never be
fixed. That includes the leniencies the reviews found (a plain value holding `: `, unquoted hashes, `TRUE`, the last of
two duplicate keys, document markers skipped). Bonsai's Go reader has a format-0 mode that is a hand port of
`yaml.mjs`, tested on today's files, never a general YAML library. Duplicate keys stay readable in format 0; the risk
is closed elsewhere: format-0 Bonsai-kind files cannot change (§2.3), open tasks move to format 1 (§15), and the
studio's writer refuses any file with a duplicate key. On 7 Oct, all 229 files under `studio/` here and Mimas's 128
parsed, with no duplicate top-level key.

**Format 1 reads by this grammar and refuses everything else.** Under it every value means the same to `yaml.mjs`,
Bonsai's Go reader and the studio, and a YAML 1.1 or 1.2 library reads the same value (a date stays text to every
contract reader). `yaml.mjs`'s format-1 mode implements these rules; it does not inherit today's leniencies.
- **Lines.** LF or CRLF (writers keep the file's endings). One leading BOM is ignored, in frontmatter and YAML files
  alike (today `parseYaml` throws on one). No tab in indentation. A frontmatter block and a YAML file are each one
  document: `---` and `...` are refused anywhere else. **Every line is read:** a line the reader did not consume is
  refused (today a list at its key's own indent silently drops every key after it).
- **Keys.** Core and definition fields `[a-z][a-z0-9_]*`; label names `<namespace>.<name>` (§5.1). Never quoted, never
  complex, never `<<`, never twice in one mapping, and never one of `y`, `n`, `yes`, `no`, `on`, `off`, `true`,
  `false`, `null` (a YAML 1.1 reader takes those as booleans or null).
- **Structure.** Nested mappings by indentation. Block sequences, indented deeper than their key, with exactly one
  space after `-`. One-line flow sequences of scalars (`[a, b]`); `[]` and `{}`. Block scalars `|`, `|-`, `>` and `>-`
  only; no line inside one starts with `#`; in `>` no line is indented deeper than the first. Refused: anchors, aliases,
  tags, flow mappings with content, nested or multi-line flow sequences.
- **Comments.** A comment starts at `#` at the start of a line or after a space, as YAML has it. A quote character
  inside a plain value is just a character: it opens nothing. **Writers keep comments:** a writer changes only the
  lines it owns, and a line it changes keeps its trailing comment (a format-1 task's status line has none, §10.2). Files
  Bonsai writes from a template carry comments (§2.8).
- **Quoted scalars** are always text, on one line, and nothing but a comment may follow the closing quote. In double
  quotes only five escapes exist (a backslash before `"`, `\`, `n`, `r` or `t`); any other backslash, or an unescaped
  `"` inside, is refused.
- **Plain scalars** must be one of:

  | Plain scalar | Reads as |
  |---|---|
  | empty, `null`, `~` | null |
  | `true`, `false` (lower case only) | boolean |
  | `-?(0\|[1-9][0-9]{0,14})` | integer (at most 15 digits: exact in JavaScript and Go alike) |
  | `-?(0\|[1-9][0-9]*)\.[0-9]+` | decimal |
  | `YYYY-MM-DD` or `YYYY-MM-DD HH:MM` | text, never a date or time object |
  | starts with an ASCII letter or `_`; holds no `: ` and no ` #`; does not end in `:`; is not one of the words above in any other case, nor `y`, `n`, `yes`, `no`, `on`, `off`; inside a flow sequence also holds no `,`, `[`, `]`, `{`, `}` | text |
  | anything else | refused: "quote this value" |

  So `0755`, `55227e5`, `0x1F`, `1e3`, `1_000`, `.inf`, `TRUE`, `No`, `5 arenas`, `.claude/**` and a value holding
  `: ` are refused bare, and read as text once quoted. **The rule for writers and agents: quote every text value that
  does not start with a letter. Commit hashes are always quoted.**

**The trick files.** One shared set, each file with the outcome every reader must reach under each format (accepted
with this value, or refused for this reason): one file per rule above, plus today's format-0 oddities, CRLF, a BOM in a
definition file, and a duplicate key quoted and unquoted. The set starts in this repo (§15) and moves to Bonsai's repo
with its records (Q9 b); the studio's readers test against the set of the Bonsai release they pin. Mimas's hooks and
ladder run a pinned copy (Mimas's own T-0040, tag `mimas-pin-2026-10-05`), and its Desk path reads format 0 as today,
so nothing changes for Mimas.

> **Changed 8 Oct (Rohan):** the trick files and the JSON Schemas start in Bonsai's repo, in `formats/`, as Bonsai's
> first job ("those are kind of tests which other projects can use ... it makes sense for them to live in bonsai").
> Bonsai holds the master; the studio's readers test against it when the studio links (step 7; see §15.1's note of 8 Oct).

### 2.5 JSON documents and records

`status --json`, the ladder result and the lock are JSON documents; the log and asks are JSON lines (one record per
line, LF). UTF-8, no BOM; **duplicate keys refused** by every reader (Node's `JSON.parse` and Go's decoder both keep the
last, so each reader adds a small check); fields in a fixed order (§2.2). Records are appended whole, one line per
write, with today's retry on Windows' busy errors.

### 2.6 What may reach the VPS

The VPS is Rohan's own private server, but it faces the internet and runs his other sites, so the studio keeps no
secrets there (spec 2026-09-29 §8; golden rule 1).
- **No format has a secret field.** Keys, tokens and the input-hash key stay in their own files on the machine.
- **The bridge forwards by allowlist:** known fields of known majors, and labels only when their name is on the
  studio's allowlist (today: `studio-app/server/schemas.ts`). Everything else stays on the machine.
- **Free text only in the fields listed, and only redacted:** documents (frontmatter and body, indexed as today); the
  log's `target` and `text`; an ask's `title`, `why`, `then_text`, `options` and answer `words`. Redacted when written
  (decision 2, Bonsai's) and again by the bridge, as today.
- **No absolute paths** in committed formats (a check fails on one; vision A.1). Log records carry workspace-relative
  paths and a checkout's folder name. A checkout's absolute root appears only in `status --json`, printed locally; the
  bridge keeps it in its registration and never forwards it.
- **Memory's personal layer** (§7.4) never leaves the machine.

### 2.7 Agents other than Claude Code

Claude Code is the first agent (Q5 a). Formats name no agent, except fields marked **Claude Code only** where they
appear. Another agent gets a thin adapter later (vision A.2), on Rohan's word.

### 2.8 Templates document themselves

Rohan, 7 Oct: "any kind of template file that bonsai will have, inside that file there should be proper official level
documentation, so any person and ai agent can easily know the use and purpose of it." Every template and pack file
Bonsai or a pack ships carries its own documentation: its purpose, when to use it, and every field (meaning, allowed
values, an example), in the comment form its format allows (`#` in YAML, an HTML comment or a fields table in markdown).
A file **made from** a template (a task, a run report) does not copy those docs: its `format:` line carries a one-line
pointer back, as a trailing comment. How, and how `bonsai check` keeps the docs in step with the fields: Bonsai spec §5.

## 3. The Bonsai home and the workspace id

**The workspace id** (vision A.11). `bonsai init` writes a random id, `ws-` and 26 lowercase base32 characters, into
`bonsai.yaml` (committed), with a **name**, a slug `[a-z][a-z0-9-]{0,39}` (today's project-id shape).
- The id names the project: every clone and worktree shares it, and records name it (§8).
- **The id is not a key to anything on the machine.** Machine settings are kept under the main checkout's path (below).
  Editing the id changes no setting, and a copy of the repo elsewhere is a separate, unmanaged workspace.
- **A changed id** is a problem in `bonsai check` and `status --json` ("bonsai.yaml's id changed from ws-… to ws-…").
  Bonsai keeps working in the same machine folder; the studio's bridge stops forwarding the project until it is
  registered again (§16).
- **Copies:** a clone is the same project by design: every clone and worktree shares the id. A copy meant as a new
  project (a scratch copy for the walking skeleton, Q1 c) runs `bonsai init --new-id`, which gives it its own id and,
  after its preview, empties the copy's `.bonsai/local/` (a folder copy carries the original's log, asks and ladder
  results; the original keeps its own). `bonsai check` warns when two checkouts on one machine hold one id.
- **Moved checkouts:** `.bonsai/local/` moves with the checkout. Only its machine folder stays behind under the old
  path, so `status_writes` falls back to `agents` and attached labels are missing until it is registered again (the
  bridge flags a lost `command` setting, §10.6). `bonsai check` and `status` report the stranded folder and the old
  path, and name the step: register the checkout again (§15.2), or a person's `bonsai settings set`.
- `bonsai.yaml` and the lock are on base's protected list (§14).
- The studio's project id is its own, kept in its registration; it defaults to the workspace name.

**The project's own folder, `.bonsai/`** (Rohan, 7 Oct, format review round 1): `bonsai.yaml` stays at the root, the one
Bonsai file a person edits. Beside it, `.bonsai/` holds the lock, STATE and the generated tables, committed, and
`.bonsai/local/`, never committed: the log, asks and ladder results. Only the lock is person-only, not the folder.

```
<main checkout>/
  bonsai.yaml                          the project's settings (committed, person-only)
  .bonsai/
    .gitignore                         written by Bonsai; ignores local/
    lock.json                          the lock (§14; committed, person-only)
    STATE.md                           where the project stands (§7.2; committed; agents rewrite it)
    tasks.md, sessions.md              the generated tables (§7.5; committed; never edited by hand)
    local/                             never committed; only in the main checkout
      log/s-<session>.ndjson           one file per session (§8)
      log/w-<YYYY-MM-DD>.ndjson        records outside a session (§8.4)
      asks/<YYYY-MM-DD>.ndjson         asks and answers (§9)
      ladder/<task id>.json            ladder results (§11)
```

- **Worktrees share the main checkout's `.bonsai/local/`**: a worktree finds its main checkout through git (`git
  rev-parse --git-common-dir`), as the hooks do today. A task worktree has no `local/` of its own.
- **Never committed:** Bonsai writes `.bonsai/.gitignore` itself, and `bonsai check` and rung 0 refuse any file from
  `local/` that is tracked or staged.
- **Only Bonsai writes `local/`:** a deny rule and the guard refuse an agent's file-tool write there (a tripwire,
  decision D; §10.6).
- **How long each kind is kept** is set per kind in `bonsai.yaml` (the Bonsai spec's "Generated files", §6), and every
  file Bonsai cleans is logged.

**The Bonsai home** is the machine folder for what never goes into a project and belongs to the machine, not the
project: `BONSAI_HOME`, else `~/.bonsai` (Linux, WSL included) or `%USERPROFILE%\.bonsai` (Windows). Tests always set a
temp one (golden rule 3). Each workspace's machine folder is `workspaces/r-<16 hex>/`: the first 16 hex characters of
the SHA-256 of the main checkout's real path (case-folded on Windows), shared by its worktrees. Since 7 Oct it holds
only this machine's settings for the workspace; its records live in the project's `.bonsai/local/`.

```
<BONSAI_HOME>/
  salt                                 the input-hash key; never leaves the machine (today: TRINETRA_HOME/salt)
  settings.json                        the home's own settings: cache_keep_days (Bonsai spec §6)
  workspaces/r-<16 hex>/
    workspace.json                     the main checkout's path, and the workspace ids it has held
    settings.json                      this machine's settings for it: status_writes, status_command
    labels/<namespace>.yaml            label definitions attached on this machine (§5.3)
  cache/                               the pack cache; deletable
```

(The Bonsai spec adds `personal/`, `locks/` and `install.json`, §10.) **Said in plain words:** `bonsai init` ends by
saying where the home is, what goes there and that none of it enters git, where `.bonsai/` is and what it holds, and
the workspace id; `bonsai status` shows the same (Bonsai spec §6).

**Bonsai keeps no act log**: it applies nothing; the studio's journal stays in the studio's home (§10.5).
`settings.json` and `labels/` change only through Bonsai's commands run outside an agent session (walls: §10.6).

## 4. The task

**What it is for.** One unit of work: what it is, where it stands, when it is done. Decision 1: "Bonsai owns a small,
versioned core task format as part of how agents work (id, title, status, done-when, run report) ... Trinetra reads it
and adds its own fields as labels". The 7 Oct rule gives the actions on it to the studio (§10).
**Written by** agents, who create tasks and set the fields agents may set, and the studio's write path (status moves,
lane changes, grants; §10). In a managed workspace **a task file changes only in the main checkout** (§10.1), so the
Desk, which reads the main checkout, sees every move. In a workspace the studio does not manage, agents edit task files
themselves. **Read by** Bonsai (checks, the guard, the gate, the runner), the studio, the statusline, agents.
**Lives** in the folder the workspace declares for tasks (§7.3), one markdown file per task, named `<id>-<slug>.md`.

### 4.1 Core fields (format 1)

| Field | Type | Meaning |
|---|---|---|
| `format` | `bonsai.task/1` | The version line; the first key |
| `id` | `T-` and 4 to 6 digits | Unique in the workspace; plain, never quoted |
| `title` | text, one line | What gets built |
| `status` | one of the closed list below | Where it stands; no comment on its line |
| `lane` | a lane name, or null | The lane slot (Q6, §6); null where no pack defines lanes |
| `done_when` | list of text | Checkable statements a verifier can test without having built it |
| `depends_on` | list of task ids | Tasks that must be done first |
| `blocked_by` | text or null | One sentence, while `status` is `blocked` |
| `created`, `started`, `finished` | dates or null | `YYYY-MM-DD` |
| `labels` | map of label name to value | Everything else (§5) |

**Statuses** (closed list): `todo`, `plan`, `approved`, `running`, `verify`, `done`, `blocked`, `cut`, today's eight.
Whether statuses stay fixed in core is §18 question A; this spec assumes they do. Who may make each move is §10.2.
Today most closed tasks carry a comment on the status line ("closed by Rohan ..."); in format 1 the line has none, since
the commit's trailers, its message and `blocked_by` carry the why.

**Decided here** (the vision §4: "adding `depends_on` and start and finish times is a proposal"): `depends_on`,
`created`, `started`, `finished` and `blocked_by` are core. Every task has them today and they mean the same in any
project. Decision 1's "run report" is met by the run report naming its task (§7.1); the task carries no pointer back.

**No link to outside trackers.** No core field and no `bonsai.*` label points at an outside tracker. Rohan, 7 Oct:
outside services become Studio "connectors", kept in the studio, "as putting something like that in bonsai would end
up bloating it a lot".

### 4.2 Example

A made-up task:

```yaml
---
format: bonsai.task/1
id: T-0999
title: Parity counts the projects on the Windows side of the PC
status: todo
lane: light
done_when:
  - "parity.mjs compares every project the PC's bridge serves"
  - "Rungs 0-4 green in one ladder run in WSL"
depends_on: []
blocked_by:
created: 2026-10-07
started:
finished:
labels:
  bonsai.branch: t0999-parity-windows-side
  bonsai.ladder: [0, 1, 2, 3, 4]
  bonsai.allows: []
  bonsai.wants: []
  workflow.owner: builder
  workflow.model: sonnet
  workflow.feature: F-studio-v2
  workflow.milestone: V2
---

# Parity counts the projects on the Windows side of the PC
...
```

Every value passes the format-1 grammar (§2.4), and today's `yaml.mjs` reads it as intended.

### 4.3 What today's task fields become

| Today (format 0) | Format 1 | Why |
|---|---|---|
| `id`, `title`, `status`, `done_when`, `depends_on`, `blocked_by`, `created`, `started`, `finished` | core, unchanged | Decision 1 and §4.1 |
| `lane` | core `lane` | Q6: "a lane slot in base". The walkthrough's note on Q6 (not Rohan's words): "the stop rule must hold without Trinetra attached" |
| `project` | dropped | The workspace is the project |
| `allows_assets` | label `bonsai.allows` (granting, §5.5) | Read by Bonsai's guard and rung 0 |
| `ladder` | label `bonsai.ladder` | Read by the runner and the stop gate, above the workspace's floor (§5.6) |
| `worktree` | label `bonsai.branch` | Its value is a branch name |
| `owner`, `model`, `feature`, `milestone`, `spec` | `workflow.*` labels | Rohan's way of working: his workflow pack's |
| `cost_usd` | label `trinetra.cost` | Decision 1's own example; the studio's fact |

**`bonsai.*` labels, not core fields,** for `allows`, `wants`, `ladder` and `branch`: Q4 (b) put the guard and the
runner inside Bonsai, so they are Bonsai's own labels (vision A.3 proposed labels); they are optional per task, carry
definitions agents see (§5.6), and new ones need no new task major. **`lane` is core**: it decides at every move which
moves are a person's, and a check must not depend on a label definition being attached. **`workflow`** is a placeholder
for the id of Rohan's workflow pack, which the Bonsai spec names.

**The gap between shapes** (vision A.6): today's guard and rung 0 read `allows_assets` as a top-level key
(`paths.mjs:304`, `ladder.mjs:478`). No task changes shape until every reader reads both shapes (§15).

## 5. Labels and their definitions

### 5.1 What a label is

A named value on a document: `<namespace>.<name>: <value>`, under the document's `labels:` map. Name pattern
`^[a-z][a-z0-9-]*\.[a-z][a-z0-9_]*$` (one dot). Namespaces, after Docker, OCI and Kubernetes (vision A.3):
- `bonsai.*` is reserved for Bonsai and the packs Bonsai publishes; each other pack uses its own id (`workflow.*`).
- A namespace attached on the machine (§5.3) may not equal any pack's in the workspace; Bonsai refuses the attach.
  That is how `trinetra.*` stays the studio's without Bonsai's code naming it.
- A definition may be added, never redefined (vision A.1). Two sources defining one name is a `bonsai check` finding.

### 5.2 The definition format

```yaml
format: bonsai.labels/1
namespace: trinetra
version: 1                 # the definer's own version of this set; additive
labels:
  - name: trinetra.cost
    kind: number
    kinds: [task, run]
    set_by: outside
    description: "Dollars this work cost. Agents never write it."
```

| Field | Meaning |
|---|---|
| `name` | The label, in this namespace |
| `kind` | `choice` (one of `values`), `text` (one line; optional `pattern`, `max` up to 300), `number` (integer or decimal), `list` (a sequence of `items: text` or `items: number`) |
| `values` | For `choice`: the allowed values |
| `kinds` | Which document kinds may carry it (§7.3). Not `on`: a YAML 1.1 reader takes that key as `true` |
| `set_by` | `agent`: agents write it in the file. `outside`: agents never write it; a program outside agent sessions does (in a managed workspace, the studio's door 1, §10.1) |
| `grants` | `true` when the value grants a right (§5.5). Default `false` |
| `description` | One line, at most 200 characters; what agents see (about 40 tokens) |

Four kinds, not strings only: today's fields include lists and a number, and strings would need a second list syntax.
A value of the wrong kind is a finding, and a guard ignores it (for a granting label: no grant). **Where these files sit
in a pack** is the Bonsai spec's; packs are also valid Claude Code plugins with skills to the open standard (Rohan,
7 Oct), so these files sit outside the plugin's own folders.

### 5.3 The studio's definitions live on the machine

The vision's recommendation (§10), kept by the 7 Oct rule: the studio puts no file of its own into a project.
- The studio keeps the source of its definitions in its own repo.
- Registration (`trinetra project add .`, §15.2) runs Bonsai's attach command, which checks the file and copies it to
  `<home>/workspaces/<key>/labels/trinetra.yaml`. Unregistering detaches it. The bridge re-attaches on start when its
  definitions' `version` is newer. The studio never writes into Bonsai's home directly.
- Packs' definitions are rendered into the instruction file. Machine-attached ones reach a session through its opening
  context (**Claude Code only:** the session-start hook's).

### 5.4 Writing a label

- `set_by: agent`: agents edit the value in the file like any field (in a managed workspace, in the main checkout only,
  §10.1).
- `set_by: outside`: in a managed workspace only door 1 writes it (the bridge, in-process, after the task's branch has
  merged, to the main checkout only); door 2 refuses it, and Bonsai's guard refuses an agent's edit (a tripwire).
  `bonsai check` checks every value against its definition. Bonsai has no label-writing command.

**The studio's labels, version 1:** `trinetra.cost` (number; task and run; outside; `cost_usd`'s new home; nothing
writes it yet, since spend per task is blank on Rohan's 7 Oct checklist), and, for its own log events (§8.4),
`trinetra.event` (choice: `deploy`, `registered`, `unregistered`, `applied`) and `trinetra.ref` (text, at most 60, a
commit). Display facts (a project's name and colour) stay in the studio's registration and server.

**Workflow labels, for the record:** `workflow.owner` (choice: the pack's roles), `workflow.model` (choice; **Claude
Code only** values today: opus, sonnet, fable), `workflow.feature`, `workflow.milestone`, `workflow.spec` (text, an
id). The workflow pack defines them.

### 5.5 Grants (TB-029)

A label with `grants: true` widens what an agent may do; the only one is `bonsai.allows`. TB-029: an agent can grant
itself protected paths by editing its own task. Under the 7 Oct rule the studio applies grants:
- **The person-only list.** The workspace's project values (`bonsai.yaml`; until Bonsai, `studio/game.yaml`) hold
  `person_only:` globs, shown in `status --json`. Question B decides what goes in it: nothing (B b), every protected
  path (B a), or a short list (B c).
- **Asking.** A task that needs a person-only path lists it in `bonsai.wants`. The Desk derives a card from it, as it
  derives Approve from a draft: a full-lane task's approve card lists the wanted paths, and the approve tap grants
  them; any other task gets a "grant these paths" card. The tap copies the wanted person-only paths into
  `bonsai.allows` in its one commit (the `grant` effect, §10.3).
- **In a managed workspace** the guard refuses an agent's edit that puts a person-only path into `bonsai.allows` (a
  tripwire), and door 2 refuses one (`person_only`). Paths not on the list go into `bonsai.allows` directly, as today.
- **The guard honours what the main checkout's task file says**, while the task reads `running` (§13). What stops a
  forged value is the tripwire, and the Desk's check (§10.6), which flags any person-only value that reached `main`
  without a person's move. No approvals file is needed: the proof is the commit and the server's record.
- **In an unmanaged workspace** grants count as written; `status --json` says `status_writes: agents`.
- TB-029 closes when B is answered (a) or (c) and this lands. Under (b) it stays open, as a gap the Desk shows.

### 5.6 Bonsai's own labels

```yaml
format: bonsai.labels/1
namespace: bonsai
version: 1
labels:
  - name: bonsai.allows
    kind: list
    items: text
    kinds: [task]
    set_by: agent
    grants: true
    description: "Protected paths this task may change, as globs. A path on the person-only list counts only if a person granted it."
  - name: bonsai.wants
    kind: list
    items: text
    kinds: [task]
    set_by: agent
    description: "Person-only paths this task asks for. A person's approval or grant copies them into bonsai.allows."
  - name: bonsai.ladder
    kind: list
    items: number
    kinds: [task]
    set_by: agent
    description: "The rungs this task climbs. It adds rungs to the workspace's floor and never drops one."
  - name: bonsai.branch
    kind: text
    pattern: "^[A-Za-z0-9._/-]{1,100}$"
    kinds: [task]
    set_by: agent
    description: "The branch the task is built on; empty means the base branch."
```

**`bonsai.ladder` grants nothing** because of the floor: the workspace's project values name a ladder floor, the rungs
every task climbs (`bonsai.yaml`; until Bonsai, `studio/game.yaml`). The runner and the stop gate require the floor plus
the task's list. Today a task's `ladder:` alone decides (`ladder.mjs:111`, `stop-gate.mjs:108`), so a task listing
`[0]` would skip the tests.

## 6. Lanes: a slot in base (Q6)

Rohan, Q6: "(b) plus a slot in base - Bonsai's base knows the idea of a lane (a task can carry one; checks read it)
but defines none; his workflow pack defines light, full and director and their rules."

**The slot.** Core field `lane` (§4.1). Its value must be a lane some pack in the workspace defines; with none, null.

**The definition**, a pack file:

```yaml
format: bonsai.lanes/1
lanes:
  - name: light
    approve_first: false
    close: agent
    description: "Fixes, tweaks, tests, docs. Build it, climb the ladder, merge."
  - name: full
    approve_first: true
    close: person
    description: "A feature or a contract change. A person approves the plan before code and closes the task."
  - name: director
    approve_first: false
    close: person
    description: "A decision, not code. Write the options and stop; a person chooses."
```

**The rules base understands** (a closed list):
- `approve_first: true`: a move to `running`, `verify` or `done` needs the task to have read `approved` since it last
  read `todo` or `plan`.
- `close: person`: `done` is a person's move; `agent`: an agent may make it.
- Everything else, including the director lane's "stop and write options", is the description agents read, as today
  (`studio/protocols/lanes.md`). The director lane has no gate before work, as `lanes.md` has it.
- **Stricter** means `approve_first: true` where the other is false, or `close: person` where the other is `agent`,
  and looser on neither (light, then director, then full). An **unknown lane** counts as the strictest.

**Changing a lane** (`lanes.md`: "The Producer assigns the lane; Rohan can override it"). A lane is set when the task is
created. In a managed workspace it then changes only through the studio's write path: to a stricter lane by anyone, to
a looser one only by a person (§10.2). The guard refuses a hand edit, and the Desk's check flags a loosening no
person's move backs. In an unmanaged workspace a lane change is words, like `close: person`.

**Who checks:** the studio's write path at every move (managed); Bonsai's guard against hand edits (managed);
`bonsai check` in every workspace for `approve_first`, from git history. In a checkout without full history (a shallow
CI clone) `bonsai check` says "history not available" for that rule, never "passed".

**The studio reads** each task's `lane` and the workspace's lanes from `status --json`, and shows lanes only where a
pack defines them (Q6 b). `close: person` tells it which closes are Rohan's.

## 7. Run reports, STATE and the other documents

### 7.1 The run report (`bonsai.run/1`)

The log of one agent session on one task, opened at the start and appended to as it goes. Frontmatter:

| Field | Meaning |
|---|---|
| `format` | `bonsai.run/1` |
| `id` | `R-<YYYY-MM-DD>-<task id>`, with a suffix when a task has two on one day |
| `task` | The task id: the link between the two |
| `role` | The role the agent ran as (a pack's role name) |
| `model` | The model the agent ran on, as the agent names it |
| `started`, `finished` | `YYYY-MM-DD HH:MM`, or null |
| `outcome` | `running`, `verify`, `needs-verifier`, `needs-person`, `merged`, `blocked`, `abandoned` (closed list) |
| `commits` | list of commit hashes, quoted |
| `labels` | e.g. `trinetra.cost` |

Today's `ladder_result` path is dropped (the result's place follows from the task id, §11); `needs-rohan` becomes
`needs-person`; `cost_usd` becomes `trinetra.cost`. Format-0 reports keep their words: `needs-rohan` reads as
`needs-person`, `needs-review` as `needs-verifier`, anything else as `other`. The body is the workflow pack's template.
A run report lives on the builder's branch, like the code it describes.

### 7.2 STATE (`bonsai.state/1`)

Where a project stands, one page, rewritten. Frontmatter: `format`, `updated` (date), `updated_by` (text), `labels`
(e.g. `workflow.milestone`). The body is free. **Lives** at `.bonsai/STATE.md` in every project (Rohan, 7 Oct: such
files "should all live in one specified folder decided by bonsai"); it is not a `bonsai.yaml` setting. This repo's
`studio/STATE.md` moves there when the studio links (Bonsai spec §14 step 7).

### 7.3 Document kinds, declared

The studio must not hard-code `studio/` or any one project's folders (vision A.6 row 1). So every kind of document a
workspace holds is **declared**, by Bonsai for its core kinds and by packs for theirs, and listed in `status --json`:

```json
{ "kind": "plan", "from": "workflow", "path": "studio/plans", "id": "^P-T-\\d{4,6}$",
  "statuses": ["draft", "approved", "done"], "person": [["draft", "approved"]], "agent": [["approved", "done"]],
  "stamp": { "approved": "approved" }, "task_field": "task" }
```

| Field | Meaning |
|---|---|
| `kind`, `from` | The kind's name and who declared it (`bonsai` or a pack id) |
| `path` or `file` | The folder (top level only, one file per document) or the single file, project-relative |
| `id` | The id pattern; a document's `id` must match it and resolve to exactly one file, never a path (today's `DOC_ID_RULES`) |
| `format` | Its `bonsai.*` format, for Bonsai's kinds; absent for a pack's |
| `statuses` | Its statuses, if any; the first is where a new document starts |
| `person`, `agent` | The status moves a person makes, and those agents make (§10.2) |
| `stamp` | Fields set to the day's date when a document reaches a status (`{status: field}`) |
| `task_field` | The field naming the task a document belongs to (a plan's `task:`) |

Bonsai declares `task` (its moves are the rules in §10.2; `stamp` `{"running": "started", "done": "finished", "cut":
"finished"}`, `started` only when empty), `run`, `state` (the fixed file `.bonsai/STATE.md`), `answers` (the file for
answers to asks that name no file; today `studio/answers.md`), `memory` (§7.4), and the two generated tables `tasks`
and `sessions` (fixed files, no statuses, §7.5). Rohan's workflow pack declares the rest: plans, one-pagers, decision
records, specs, playtests, briefs, the bugs file. Each project keeps its other folder names in `bonsai.yaml`; this repo
and Mimas keep `studio/`. Rows the studio parses out of a pack document's body (the bugs table) are not contract.

### 7.4 Memory (option D)

Rohan, 7 Oct (`q-memory`): "D: Bonsai index plus notes in the repo, plus a personal layer for facts about him".
- **The index and its notes** are documents in the project, committed, a kind Bonsai declares (`memory`). The studio
  indexes them like other documents. They travel with the repo.
- **The personal layer** (facts about Rohan for every project) is machine-side, never in a project, never forwarded.
- The index's format, its size budget, and where the personal layer lives are the Bonsai spec's.

### 7.5 The generated tables (`bonsai.tasks/1`, `bonsai.sessions/1`)

Rohan, 7 Oct (format review 3.2 and 5.3, round 1): data gathered from many files lives in one folder Bonsai decides.
**Written by** `bonsai check --write` only, in the main checkout (who runs it and when: Bonsai spec §6); never by hand
(the guard refuses an agent's edit, a tripwire). **Read by** people (on GitHub and the Desk), agents, the studio.
**Lives** in `.bonsai/`, committed.
- **`.bonsai/tasks.md`**: frontmatter `format: bonsai.tasks/1`; above the table, the active task as §13's step 2 finds
  it (the one task reading `running`, or none and why); then one row per task: id, title, status, lane, started,
  finished, newest id first. Built from the task files.
- **`.bonsai/sessions.md`**: frontmatter `format: bonsai.sessions/1`; one row per ended session and per ended subagent
  run (**Claude Code only:** a subagent runs inside its session and carries its environment): the session id's first 8
  characters, task, role, model, start, end, minutes (UTC, `YYYY-MM-DD HH:MM`); then hours per task and role (a
  subagent run lies inside its session's row, so the two are never added). Built from the log's `session_start`,
  `session_end`, `subagent_start` and `subagent_stop` records. **A row's task, one rule:** the task named by `--task` or
  the environment, if set; else the active task (§13) of the checkout or worktree it ran in; else `none`; as the
  recorder found it at the start (`target`, §8.1). Its role: `role` for a session, `subagent_type` for a subagent run.
  Rows are only added, so a row outlives the log file it came from (§8.5). It replaces the run report's "Sessions"
  table.
- **The task files and the log stay the truth. A table never grants anything:** the guard, rung 0, the stop gate and
  the active-task rule (§13) read the task files; nothing reads a table to decide.
- **Stale is a warning, everywhere** (a tasks table that differs from what `--write` would write, or a sessions table
  that lacks a row for an ended session or subagent run in the log): `bonsai check` warns in the main checkout, a
  worktree and CI alike; it never changes an exit code, never turns a rung red and is never one of `status --json`'s
  `problems` (§12). The tables lag between moves by design; the next `--write` rebuilds them. Branches never change a
  table (rung 0 refuses one).

## 8. The log (decision 2)

Rohan, decision 2: "Every Bonsai workspace keeps a local log of what its agents did ... redacted as it is written, kept
on the machine and never committed ... Trinetra adds only the forwarder ... Trinetra adds its own events as labels
(`trinetra.*`) in the same log." Confirmed 7 Oct (`q-record`): "Bonsai records; Studio forwards and shows".

**Written by** Bonsai's recorder (the hooks) and Bonsai's own commands. **Read by** `bonsai logs`, the studio's
forwarder (Bonsai's screens, Q7, dropped by Rohan on 9 Oct: every visual is the studio's). **Lives** in the main checkout's `.bonsai/local/log/`, never committed (§3; moved from
the Bonsai home on Rohan's word, 7 Oct). Still "kept on the machine and never committed", as decision 2 says.

### 8.1 The record (`bonsai.log/1`)

One JSON line, at most 2,048 bytes (today's cap). Fields in this order; `null` when they do not apply.

| Field | Meaning | Today's spool field |
|---|---|---|
| `format` | `bonsai.log/1` | `v: 1` |
| `id` | UUID; a replayed line is a duplicate | `id` |
| `at` | ISO-8601 UTC with milliseconds | `at` |
| `workspace` | The workspace id (the bridge maps it to the Desk's project id, §16) | `project` |
| `session` | The agent's session id, or null outside a session | `session` |
| `agent` | The agent program: `claude-code` today | new |
| `agent_event` | The agent's own event name (§8.3) | `event` |
| `event` | Bonsai's event name (§8.2) | new |
| `subagent_id`, `subagent_type` | **Claude Code only:** the subagent, when one acted | `agent_id`, `agent_type` |
| `tool_use_id`, `input_hash` | Pair a call's start and end (**Claude Code only:** the id); the hash is keyed by the machine's salt and keeps no input | same |
| `checkout`, `branch` | The working tree's folder name (never its path) and its branch | `worktree`, `branch` |
| `task`, `role` | As the session's environment names them (`BONSAI_TASK`, `BONSAI_ROLE`), raw | `task_env`, `role_env` |
| `tool`, `category`, `target` | The tool, its category (§8.2), its target (redacted, at most 200); for `session_start` and `subagent_start`, the active task the recorder found by §7.5's rule, or null | same |
| `ok` | The call succeeded, failed, or null | same |
| `kind`, `text`, `source`, `model`, `reason` | As today: a notice's kind, short redacted text (at most 300), where a session started from, its model, why it ended | same |
| `decision`, `rule` | For `guard`: `allow` or `deny`, and the rule's name | new |
| `labels` | For `event` only: the outside event's labels (§8.4); `{}` otherwise | new |
| `remote` | **Claude Code only:** the session's Remote Control id (`^session_[A-Za-z0-9]{10,60}$`), never redacted, as today | `rc_session` |

### 8.2 Event names and categories

**Events** (an open list): `session_start`, `prompt`, `tool_start`, `tool_end`, `tool_fail`, `permission`, `notice`,
`subagent_start`, `subagent_stop`, `stop`, `session_end` (from the agent's hooks); `guard`; `ladder` (a ladder result
written, §11); `ask` (an ask filed, resolved or answered: `target` is its key); `event` (an outside event, §8.4);
`clean` (a generated file Bonsai deleted: `target` is its project-relative path, `reason` the rule, §8.5).

**Categories** (an open list): `Read`, `Search`, `Edit`, `Write`, `Shell`, `Ladder`, `MCP`, `Agent`, `Web`, `Other`,
plus categories a pack declares as `<namespace>.<Name>`. Today's `Grep` and `Bash` become `Search` and `Shell`; today's
`Unity CLI` becomes a category of the pack that knows Unity, when Mimas links. The bridge maps both ways while the
server still speaks spool/1 (§15).

### 8.3 Claude Code's events (Claude Code only)

| Claude Code hook | `event` |
|---|---|
| SessionStart, SessionEnd | `session_start`, `session_end` |
| UserPromptSubmit | `prompt` |
| PreToolUse, PostToolUse, PostToolUseFailure | `tool_start`, `tool_end`, `tool_fail` |
| PermissionRequest, Notification | `permission`, `notice` |
| SubagentStart, SubagentStop, Stop | `subagent_start`, `subagent_stop`, `stop` |

### 8.4 The studio's own events

Through Bonsai's append command (working name `bonsai log append --label trinetra.event=deploy --label
trinetra.ref=c6ba392`): one `event` record in `log/w-<date>.ndjson`, with `session` null. Bonsai checks each label
against the definitions attached on the machine (§5.3). After each commit its write path makes (§10.5), the studio
appends one with `trinetra.event=applied`, `target` the act id, `text` the move and any chat words (redacted), and
`trinetra.ref` the commit, so `bonsai logs` shows the studio's actions beside the agents'.

### 8.5 Files, keeping, and the forwarder's cursor (A.11)

- One file per session, one per UTC day for records outside a session. Lines are only appended; no file is ever
  renamed or cut short.
- **Keeping:** Bonsai deletes whole files whose last line is older than `bonsai.yaml`'s `generated.log.keep_days`
  (default 30; it was the machine setting `log_keep_days`), never the file of a session still open, nor one holding an
  ended session or subagent run whose row is not yet in `.bonsai/sessions.md` (§7.5). Each deletion is a `clean` record. The rules for every generated kind are the Bonsai spec's "Generated files" (§6).
- **The cursor** is the forwarder's, kept in the studio's home: a project, a file name and a byte offset per file. A
  file deleted before the cursor reached its end is reported to the server as a gap. A line counts as evidence once a
  copy exists off the machine (vision §4).

## 9. Asks and answers

### 9.1 The record (`bonsai.ask/1`)

**What it is for:** an agent (or the ladder) asks a person something typed; a person answers. **Written by**
`bonsai ask`, `bonsai answer` (at a terminal, or run by the studio to record a Desk answer) and Bonsai's ladder runner
(Bless). **Read by** `bonsai asks`, agents (`bonsai ask --status <key>`), the studio's bridge. **Lives** in the main
checkout's `.bonsai/local/asks/` (§3), one file per UTC day, never committed. Kept forever by default
(`generated.asks.keep_days: null`, Bonsai spec §6); a day file that holds an open ask is never cleaned.

One JSON line, at most 8,192 bytes, fields in this order:

| Field | Meaning | Today (`AskRecord` v1) |
|---|---|---|
| `format` | `bonsai.ask/1` | `v: 1` |
| `id`, `at` | UUID; ISO UTC time | same |
| `op` | `file`, `resolve`, `answer` (closed list) | `file`, `resolve` |
| `key` | `<source>:<part>`: the agent's `--key`, or `h-` and 12 hex hashed from type, target and title; Bless's is `ladder:bless-<task>` | `<source>:<project>:<part>` |
| `workspace` | The workspace id | `project` |
| `source` | `agent` or `ladder` | same |
| `session` | The asking session, or null | same |
| `type` | `Answer`, `Decide`, `Look`, `Play`, or a type a pack defines | same |
| `task`, `doc` | The id it is about (never a path), or null | same |
| `title`, `why`, `then_text`, `options`, `verdict` | As today, with today's limits (title 300, why 600, then 300, at most four options of 200) | same |
| `data` | A pack type's payload (Bless: `{task, head, result_sha, rises: [{name, was, now, new_tests}]}`, §11), or null | `bless` |
| `answer` | For `op: answer`: `{by, via, choice, verdict, words}`; null otherwise | new |

`by` is `terminal`, or the caller's own reference (text, at most 60); `via` is free text naming who carried it (at most
30), or null. Bonsai stores what it is given; its code names no caller. A second `bonsai answer` with the same key and
`by` changes nothing.

**The key drops the project.** The bridge adds its project id when forwarding (`agent:trinetra:h-…`,
`ladder:trinetra:bless-T-0068`), so the server's keys keep today's shape. Their value changes at the switch: today's
hash covers the project too (`asks.mjs:106`), so an ask filed again across the switch shows once as a new card.

### 9.2 Types and who files them

Today's split stays (vision A.5): agents file `Answer`, `Decide`, `Look`, `Play`; only the ladder files `Bless`;
`Approve` (and the grant card, §5.5) is derived by the server from the files and never filed. Bless is filed by
Bonsai's runner (`source: ladder`) when a green result made on a clean base branch at HEAD shows a ratchet count that
rose (today's condition, `bless.ts:75`). Rohan, 7 Oct (`q-ratchet`): "A Bless ask whenever a count rises, plus the
new-tests-must-fail check"; each rise carries that check's numbers (§11). The floor rises only when the studio applies
a person's `bless` tap (§10.3).

### 9.3 Who may answer

- **The terminal answers by default; the Desk when managed** (vision 2.1). Both write an `answer` record.
- **An answer from the session that asked is refused** (vision 3.4). **Claude Code only:** sessions are told apart by
  `CLAUDE_CODE_SESSION_ID`.
- **An answer answers; it never grants.** Only a person's move applied by the studio grants (§10).
- **Agents read answers** with `bonsai ask --status <key>`, which returns the answer record.

### 9.4 Notes in files

A Desk answer is written two ways, as today: a **note** in the document it is about (or the answers file), written by
the studio (§10.3), and the `answer` record, written by `bonsai answer`, which the studio runs with `by: act:<id>` and
`via: desk`. The note keeps today's heading `## From the Desk` and today's fixed line (`asks.mjs` `DESK_SECTION`,
`contract.ts` `deskNote`). A terminal answer writes only the record.

### 9.5 Under Claude Code's sandbox (Claude Code only)

Hooks run outside the sandbox, so the recorder can write `.bonsai/local/`. The sandbox is not on in this WSL distro yet
(it needs `socat`, not installed, and open WSL2 bugs), so `bonsai ask` works today. **The outbox folder is gone**
(Rohan, 7 Oct, format review round 1, with the move into `.bonsai/local/`). How `bonsai ask` and `bonsai ladder`, run
through Bash, write `.bonsai/local/` under the sandbox (the deny rule over it, §10.6, may close it to the sandbox too;
a worktree session's `local/` is the main checkout's, outside its folder; both unchecked) is settled by the sandbox's
probe, before the sandbox goes on. **The fallback, if the probe finds them blocked:** the main checkout's
`.bonsai/local/` joins the sandbox's allowed write paths, and the file tools stay stopped there (by the deny rule if
the allowance leaves it working, else by the guard: a tripwire either way). Bonsai spec §7.

## 10. Status moves and Desk taps: the studio applies them

Rohan, 7 Oct (`q-status`): "bonsai is there to provide the task templates ... bonsai gives a structure, while studio
applies the actions." Confirmed the same day: the studio applies status moves, approvals and protected-path grants,
and dispatches agents; Bonsai has no apply and no task-status command; one write path, "a small studio command that
agents and the Desk both use, checked against Bonsai's format"; a project without the studio still works. This
replaces Q3 ("Bonsai applies it at once") and decision 5's "Trinetra writes only labels". What stays of Q3: taps are
instant.

### 10.1 One write path, two doors

- **The apply core** is the one write path: today's executor and writers (`studio-app/bridge/executor.ts`,
  `writers.ts`), grown to read format 1 (declared kinds and lanes from `status --json` instead of the fixed folders in
  `contract.ts` `path_outside`) and the move rules (§10.2).
- **Door 1, the bridge,** calls it in-process for Desk taps from the server's queue, as today (any other feeder is
  refused `not_from_queue`). Desk taps never go through door 2.
- **Door 2, the command,** for agents, the orchestrator and the terminal (working name `trinetra`):
  `trinetra move <id> <status> --from <status> [--why <text>]` moves any declared document by id, and
  `trinetra lane <id> <lane> --from <lane>` changes a task's lane. Agents' moves only; a person's move is refused
  `person_only`, except under question C (b), where `--via chat --words "<his words>"` makes one (refused while the
  environment names a task, §10.6).
- **Door 2 runs the main checkout's copy** of the studio (found through `git rev-parse --git-common-dir`, as the hooks
  are), with the main checkout's installed dependencies. An agent cannot edit its refusals away on its branch, and a
  fresh worktree needs no `npm ci`.
- **Main only.** Both doors write the main checkout's file, wherever they are run from; door 2 run in a worktree still
  writes main. Branches never change a task file, so merges never conflict on one, and the Desk, which reads only the
  main checkout (`bridge/reader.ts`), sees every move. The guards, rung 0 and the stop gate read tasks from the main
  checkout too (§13).
- **Managed or not** is a Bonsai machine setting, set at registration (§15.2): `status_writes: command` with
  `status_command: "trinetra move"`, or the default `status_writes: agents`. Until Bonsai, this repo carries the same
  line in `studio/game.yaml`, switched on at the end of slice 3. In `command` mode Bonsai's guard refuses an agent's
  edit of a task file outside the main checkout, of a declared document's status line (a new document may be created
  at its first status only), of the lane line, of a person-only path in `bonsai.allows`, and of an `outside` label,
  and tells the agent to run the setting's command. It also refuses writing a Bonsai-kind file that fails its format
  (one broken task file would make the active task "none" for everyone, §13). In `agents` mode agents edit task files
  themselves.
- **Checked against Bonsai's format.** Before writing, the file must pass its format: the §2.4 rules and the JSON Schema
  of the pinned Bonsai release (until Bonsai, the schema written in this repo, §15), else `format_check`. After
  writing, the re-parsed file may differ only by the effect (`contract_field`).

### 10.2 Which moves, and whose

For tasks, with the lane rules of §6:
- **A person's moves:** to `approved` (from `todo` or `plan`); to `done` in a `close: person` lane; to `cut`, from any
  open status; a lane change to a looser lane. A person may also make any agent move.
- **Agents' moves:** every other move between the open statuses (`todo`, `plan`, `approved`, `running`, `verify`,
  `blocked`); to `done` in a `close: agent` lane; a lane change to a stricter lane. In an `approve_first` lane, a move
  to `running`, `verify` or `done` needs the task to have read `approved` since it last read `todo` or `plan`.
- **Never:** out of `done` or `cut` (`transition`).
- Moving to `blocked` needs `--why`, which fills `blocked_by`; leaving `blocked` empties it.

Other kinds: the moves their declaration lists under `person` and `agent` (§7.3). A plan's approval also moves its
task to `approved` when it reads `todo` or `plan`, in one commit, through either door.

**Effects** (a closed list): the status line, written without a comment; the dates the kind stamps; `blocked_by`; the
lane line; for a grant, `bonsai.allows`. Nothing else in any file changes.

### 10.3 The Desk's taps

| Tap | Effect |
|---|---|
| `approve` | A plan or one-pager draft→approved (with its `approved:` date) and its task todo or plan→approved, one commit. The card lists the task's wanted person-only paths, and the tap grants them (§5.5) |
| `accept_adr` | A decision record proposed→accepted |
| `done` | A task verify→done; a note when there are words or `how` is `accepted` |
| `cut` | A task, from any open status, →cut; a note when there are words (today the tap takes `blocked` only, `contract.ts:1206`) |
| `note` (answer, decision, changes, unblock) | Words under `## From the Desk` (§9.4); `bonsai answer` records the answer |
| `bless` | Raises the ratchet floors the task's latest green result earned, recomputed at apply time, never lowered. The floors live in `studio/game.yaml` `ratchets:`, later `bonsai.yaml`: the one protected file a tap writes, as today |
| `grant` (only if question B is (a) or (c)) | Copies a task's wanted person-only paths into `bonsai.allows` |
| `snooze` | Stays on the server |

Everything is recomputed from the files at apply time, never read from the ask's stored data (vision A.5); words go
through today's normalising (NFC, LF, hidden characters refused, at most 2,000).

**Pickup** (confirmed 7 Oct; it lands in slice 3, T-0018). After an `approve` tap is applied, the bridge makes a plain
git worktree on the task's `bonsai.branch` and starts `claude --agent workflow:builder --bg` there (**Claude Code
only**; plugin roles carry their pack's name, Bonsai spec §5; corrected 7 Oct, format review 6.7), with
the task named in its environment (`BONSAI_TASK`; today `TRINETRA_TASK`), so the guards and the stop gate engage
(§13). Its first move is `trinetra move <id> running --from approved`, which lands on main (§10.1). A line in the
project's instructions stops it merging and pushing on its own. T-0018's plan decides how pickup stays single across
a crash, how several builders started at once share "one ladder at a time", and who sweeps their leftover processes
(TB-044). Who may trigger pickup is question D.

### 10.4 Checks and codes

Door 1 keeps today's order and codes (`contract.ts` `BRIDGE_REFUSALS`, `executor.ts:4-9`), with three new codes after
`stale_status`:

`actions_off`, `hourly_cap` → `unknown_kind`, `server_kind` → `not_from_queue` → `bad_args` → `unknown_project` →
`bad_id` → `words` → `not_main`, `git_busy` → `path_link`, `id_not_found`, `id_ambiguous`, `path_outside`,
`path_protected` → `key_not_open`, `not_ask_state` → `bless_not_offered` → `dirty`, `stale_sha`, `stale_status` →
**`transition`** (not an allowed move), **`person_only`** (a person's move or a person-only path asked through door 2),
**`format_check`** (the file fails its format) → `contract_field` → `git_failed`, `io_failed`.

Door 2 skips the queue's checks (`actions_off` to `not_from_queue`) and `stale_sha` (the bytes a person saw). Its
`--from` is checked as `stale_status`.

### 10.5 Exactly once, and the commit

- **Door 1, as today** (`executor.ts:13-23`): the journal (`TRINETRA_HOME/actions.ndjson`), the sidecar, `recover()`
  at start with all of today's cases, the 30-day trailer search, the hourly cap, the off switch. One change: before
  comparing bytes, `recover()` looks for the act's `Studio-Action:` trailer on main; if it is there, the act was
  committed, and it is reported as applied whatever the files hold now (with two doors, a door-2 move after an
  uncounted commit is routine, and must not undo an approval).
- **Door 2 keeps no journal.** `--from` makes a late retry harmless (the task has moved on, so `stale_status`). If the
  target differs from HEAD by exactly the requested effect (a crash between write and commit), door 2 commits it;
  any other uncommitted change is `dirty`. Door 2 refuses (`git_busy`) while the journal holds an unfinished act on its
  target.
- **One lock per checkout, shared by both doors:** `TRINETRA_HOME/locks/<16 hex of the checkout's real path>.lock`,
  created exclusively with the holder's pid and that process's start time, held from the checks to the commit. A lock
  whose process is gone (pid and start time no longer match) is taken over, and that is logged. Door 2 waits up to 5
  seconds for it, then answers `git_busy`.
- **`bonsai answer` is a journal step:** after the commit and before `reported`, so a restart runs it again, which is
  harmless (§9.1).
- **The commit, both doors:** only the targets (`git commit --only`), plus, once the studio links Bonsai, the two
  generated tables (§7.5): the write path rebuilds them with `bonsai check --write` just before the commit and commits
  them whatever they held, so `main`'s tables are current at every move. They are never checked as `dirty`, and
  `recover()` rebuilds them rather than comparing them. The write path reads only `--write`'s exit code, which says
  whether it wrote (0, else 3; findings never change it); if `bonsai` is missing or fails, the move commits without
  the tables (they grant nothing; the next move rebuilds them). Trailers `Studio-Action: <uuid>` (the tap's id, or a fresh one
  for door 2) and `Studio-Via: desk`, `agent` or `chat`. A chat move's commit body carries the words
  (redacted) and the session id that ran it. `Studio-Action:` is the one trailer ever searched.
- **Logged:** after each commit, an `applied` event through Bonsai (§8.4), once Bonsai exists.
- **A format-0 project** (Mimas) is read and written as today until it links (Q14); its taps go through the same apply
  core, with Windows git for its Windows-side checkout (re-cut §3.4).

### 10.6 What proves a move was a person's

- **At apply time.** A person's move is applied only from a tap in the server's queue, or from chat under question C.
  **A tap is not proof it was Rohan:** the PC holds the Desk's act capability, and agents in WSL reach the Desk as the
  PC (Rohan, 6 Oct; spec 2026-09-29 §8). Question D decides what a PC tap may do, and whether the Desk shows which
  device tapped.
- **Chat moves** are refused by door 2 while the environment names a task (every dispatched and pickup builder), and
  each carries its session id and words; the Desk lists them. The main session, or an agent that clears the variable,
  could still make one: a tripwire.
- **After the fact, on the Desk** (costed in §15.2). The bridge scans each managed project's `main`, from the commit
  that switched on `command` mode, for person moves: a status only a person sets, an `approved:` stamp, a looser lane, a
  raised floor, a person-only grant. Each must sit in a commit whose `Studio-Action:` id the server issued for that
  project, those targets and that effect, used once, or be a marked chat move. Anything else shows as **not from a
  tap**. A trailer alone proves nothing: the server's record decides. The bridge also flags a registered project whose
  `status --json` says `status_writes: agents` (a lost setting).
- **Tripwires on the machine** (decision D): Bonsai's guard in `command` mode (§10.1); door 2's refusals; and the
  guard's refusal, in an agent session, of Bonsai's settings and label attach commands.
- **Walls** (Rohan, 7 Oct: "walls from deny rules, the sandbox and managed settings"): deny rules for the file tools
  over the Bonsai home's `settings.json` and `labels/`, over each project's `.bonsai/local/` (the log, asks and ladder
  results, §3), and over the studio's home (token, journal, sidecars, locks). The sandbox is not on in WSL yet (§9.5),
  so a script can still reach them: a tripwire until it is.
- **Without the studio**, approvals rest on trust, and `status --json` says `status_writes: agents`.

## 11. The ladder result (`bonsai.ladder/1`)

**What it is for:** the proof that a task's checks passed, at a commit. **Written by** Bonsai's runner (Q4 b).
**Read by** the stop gate, verifiers, the statusline, the runner (Bless), the studio. **Lives** at
`.bonsai/local/ladder/<task id>.json` in the main checkout, never committed (§3), wherever the ladder ran: a
worktree's ladder writes the main checkout's folder, and the result's `git` block says which branch and commit it
proves. (Today: `<runs folder>/.ladder/` in the checkout the ladder ran in.) A `--ci` run writes its own file,
`ladder/ci.json` (as today's `ci.json`), never a task's. A result whose `finished` is older than
`generated.ladder.keep_days` (default 7) is cleaned, never the result of a task that is not `done` or `cut` (Bonsai spec
§6, Rohan, 7 Oct).

Today's fields stay (`ladder.mjs`): `task`, `started`, `finished`, `git {sha, branch, dirty}`, `requested`, `green`,
`rungs` (each: `rung`, `name`, `kind`, `required`, `command`, `means`, `status`, `reason`, `duration_ms`, `tests`,
`captures`, `ratchet {name, was, now, ok}`), `skipped`, `leftovers`, `proof`. Changes:
- `format: bonsai.ladder/1` added; `project` becomes `workspace` (the id).
- **Every field always present** (§2.2): `null` or `[]` where it does not apply.
- **`mode`** is `local` or `ci`. A result whose `mode` is not `local` is never a task's proof (T-0026). A format-0
  result with no `mode` reads as `local`.
- **The floor** (§5.6): `requested` is the floor plus the task's `bonsai.ladder`; the stop gate requires each green.
- **New tests must fail on the base** (Rohan, 7 Oct). When a count ratchet rises, the runner runs the new tests on the
  base commit, and the rung's `ratchet` gains `new_tests: { "base": "<sha>", "count": 3, "failed_on_base": 3 }` (null
  when no count rose). Whether a rise whose new tests did not all fail is offered for Bless is the Bonsai spec's.
- **The fingerprint goes through the log** (vision §4): after writing, the runner appends a `ladder` record with the
  result's project-relative path, `sha256:<hex>` of its bytes, its `green`, and `task`.

## 12. `bonsai status --json` (`bonsai.status/1`)

**What it is for:** one workspace at a glance, for the studio's bridge and the assistant (decision 3); Bonsai's own
screens (Q7) were dropped by Rohan on 9 Oct: every visual is the studio's. **Written by** `bonsai status --json`; printed, never stored. The default is cheap and offline, for the
bridge to run often; `--full` adds `checks` (vision A.1).

```json
{
  "format": "bonsai.status/1",
  "bonsai": "1.0.0",
  "mode": "offline",
  "workspace": { "id": "ws-<26 base32>", "name": "trinetra", "root": "/home/<user>/Servers/Trinetra-Game-Studio" },
  "home": { "path": "/home/<user>/.bonsai", "key": "r-<16 hex>" },
  "local": { "log": "<root>/.bonsai/local/log", "asks": "<root>/.bonsai/local/asks", "ladder": "<root>/.bonsai/local/ladder" },
  "formats": { "bonsai.task": { "read": [0, 1], "write": 1 }, "bonsai.log": { "read": [1], "write": 1 } },
  "documents": [ { "kind": "task", "from": "bonsai", "path": "studio/tasks", "id": "^T-\\d{4,6}$", "format": "bonsai.task" } ],
  "packs": [ { "id": "base", "version": "1.0.0", "commit": "<sha>", "state": "ok" } ],
  "files": { "changed": 0, "missing": 0, "format0_changed": 0 },
  "labels": [ { "namespace": "bonsai", "from": "bonsai", "version": 1 },
              { "namespace": "trinetra", "from": "machine", "version": 1 } ],
  "lanes": [ { "name": "light", "approve_first": false, "close": "agent", "from": "workflow" } ],
  "status_writes": "command",
  "status_command": "trinetra move",
  "person_only": [ ".claude/**", "studio/protocols/**", "bonsai.yaml" ],
  "active_task": { "id": "T-0999", "how": "running", "why": null },
  "needs": [ { "kind": "pack", "id": "workflow", "source": "<git url>", "version": "1.2.0" },
             { "kind": "tool", "name": "node", "version": ">=22" } ],
  "problems": [],
  "checks": null
}
```

| Field | Meaning |
|---|---|
| `workspace.root` | The main checkout's path: printed locally; the bridge keeps it in its registration, never forwards it |
| `home` | The Bonsai home and this workspace's machine folder in it (§3) |
| `local` | The main checkout's record folders, absolute (`<root>` is `workspace.root`), so the forwarder never builds paths; printed locally, never forwarded |
| `formats` | Which majors this Bonsai reads and writes, per format (§2.2) |
| `documents` | Every declared document kind (§7.3) |
| `packs` | From the lock (§14); `state` `ok`, `changed` or `missing` |
| `files` | Counts of edited, missing and changed format-0 files |
| `labels` | Namespaces in force, where each came from, and its version |
| `lanes` | The lanes defined, with their rules (§6) |
| `status_writes`, `status_command` | `agents` or `command`, and the command a managed workspace's guard names (§10.1) |
| `person_only` | The globs only a person may grant (§5.5); `[]` when none |
| `active_task` | The active task for this checkout and environment (§13) |
| `needs` | What the workspace needs from a machine (kinds `pack`, `tool`, `mcp`, `shell`; an open list) |
| `problems` | One sentence each; the same findings `bonsai check` reports, a changed workspace id among them |
| `checks` | `--full` only: newer pack versions, the agent's version, MCP servers reachable |

Exit codes: 0, or 3 when Bonsai cannot read the workspace at all; it then fills `format`, `bonsai` and `problems`, and
every other field is `null`.

## 13. The active task: one definition (T-0011)

**Why.** The active task decides whose rules govern an agent right now: whose grants the guard honours, which task
rung 0 judges a change against, which task the stop gate checks. Today two readers disagree (T-0011, cut into this
spec by Rohan on 6 Oct). The guards use `activeTask()` in `tools/lib/paths.mjs`: `TRINETRA_TASK`, else the single task
reading `running`, else none ("a guard must never guess which task authorised an edit"). The ladder reads only
`--task`. So rung 0 can refuse a change the running task declares, and a gate that cries wolf gets worked around.

**Where tasks are read.** Always from the project's **main checkout's** task folder, found from any worktree through
git (§10.1): that is where task files change. The project is the one holding the path, for the guard (as today,
`paths.mjs:297`); the one the ladder runs in (`--root`), for the ladder; the one holding the session's working
directory, for the stop gate.

**The definition.** Every reader takes these steps, in this order:
1. **Named.** The command was given a task (`--task`), or the environment names one (`BONSAI_TASK`; today
   `TRINETRA_TASK`). **The environment counts only for the session's own project** (the one holding its working
   directory), so a task named for one project never governs another. The named task is active if exactly one task
   file has that id and it parses under its format; otherwise there is none, with the reason. A named task that is
   missing never falls through to step 2. If `--task` and the environment name different tasks, there is none.
2. **Running.** Else the single task whose `status` is `running`.
3. **None.** Else none: no task running, two or more running, or a task file that does not parse.

Format-0 and format-1 tasks count alike.

**What each reader does with it:**
- **The guard** honours `bonsai.allows` only from an active task that reads `running` (`session-start.md` §7: grants
  hold "for as long as the task reads running").
- **Rung 0** judges the diff against the named task's grants at any status: it judges finished work.
- **The stop gate** engages only for a named task, as today (`stop-gate.mjs:4-6`), and blocks when it is missing or
  does not parse (`stop-gate.mjs:33-39`). It reads the task from main, the run report from the session's own checkout,
  where the builder writes it, and the ladder result from the main checkout's `.bonsai/local/ladder/` (§11), which must
  prove the session's own checkout's HEAD.
- **None fails closed everywhere**: no grants, so rung 0 is red on any protected change and the guard refuses one.
  Every such refusal says why first: "no active task: two tasks read running (T-0067, T-0068)".

**Changes from today**, each toward refusing rather than guessing: tasks are read from main, not the checkout's own
copy; a task file that does not parse makes the answer none (today it is skipped); `--task` and the environment
disagreeing is none; the guard honours grants only while the task reads `running`.

**Readers:** the file guard and the Bash guard; the stop gate; the ladder runner (rung 0, and which result file it
writes); `bonsai status --json` (`active_task`: `id`, `how` — `named`, `running` or null — and `why`); a small read-only
command printing the same object (working name `bonsai task active --json`). The recorder logs the raw `task` from the
environment (§8.1) and decides nothing; the active task it notes on a start record only files that run's sessions row
(§7.5). The studio may link a session to a task by other hints for display, says it is
a hint, and never grants anything by it. The tasks table (§7.5) shows step 2's answer for people; no reader reads it,
and it grants nothing.

**Fixtures,** one folder per case, run against every reader: named; named but missing; named twice and different; one
running; none running; two running; a broken file beside one running; a worktree session whose task is read from main;
a worktree session whose ladder result is read from main's `.bonsai/local/ladder/`; the environment naming a task while the path is in another project; a named task at `verify` (the guard grants
nothing, rung 0 uses its grants); the stop gate with a named broken file (blocks) and with nothing named (allows).
T-0011's four cases are among them; the readers task in §15 carries its done-when.

## 14. The lock (`bonsai.lock/1`)

**What it is for:** exactly which pack versions and files a workspace holds, so an update writes everything or
nothing and an edit is never overwritten (vision 2.2). **Written by** Bonsai's engine. **Read by** the engine,
`bonsai check` (in CI), verifiers and people reading a diff. The studio reads it only through `status --json`.

**Lives** in the project, committed: `.bonsai/lock.json` (Rohan, 7 Oct, format review 6.2: "in .bonsai folder in the
repo root ... as we can possibly add more things there"); `bonsai.yaml` stays at the root. Both are on base's protected
list and person-only: the guard refuses an agent's edit unless the task grants them. The rest of `.bonsai/` is not
person-only (§3). Bonsai 0.4.3's `.bonsai-lock.yaml` and `.bonsai.yaml` are not read by the new Bonsai (Q8 a). This repo
has no lock today.

```json
{
  "format": "bonsai.lock/1",
  "written_by": "1.0.0",
  "packs": [ { "id": "base", "source": "https://github.com/LastStep/<base pack repo>", "version": "1.0.0",
               "commit": "<sha>", "sha256": "<content hash>" } ],
  "files": { ".claude/settings.json": { "kind": "keys", "pack": "base", "sha256": "<sha256>" } },
  "format0": { "studio/runs/R-2026-09-30-T-0033.md": "<sha256, line endings LF>" }
}
```

Per pack: source, version, commit, content hash. Per file: its kind (`pack`, `once`, `block`, `keys`, `kept`; vision
A.1) and the SHA-256 with line endings made LF. `format0`: every Bonsai-kind file without `format:` at link time (§2.3).
Keys sorted; forward slashes only; no absolute path anywhere. The engine's rules are the Bonsai spec's.

## 15. From today's files to the contract

### 15.1 In what order

> **Changed 8 Oct (Rohan): Bonsai first.** The studio's readers task is cut; the studio adopts the formats when it
> links (step 7); its Desk stays on upkeep until then.

Path (c), "formats first", adopted by this repo (vision §6):
1. **This spec approved.**
2. **The formats land in this repo: one full-lane task, T-NNNN (readers),** created on approval with the next free id,
   a plan for Rohan to approve, and a fresh verifier. Its `done_when` carries T-0011's four cases and §13's fixtures.
   **When it merges, the Desk pause lifts, and slice 3 (T-0018) starts** (Rohan, 7 Oct).

   | Part | Hours |
   |---|---|
   | The JSON Schemas for the ten formats (the two tables' come with Bonsai, step 5.1); the trick files with their outcomes | 6-10 |
   | `yaml.mjs`'s format-1 mode (the grammar, the BOM fix) | 5-8 |
   | `paths.mjs` and the guards: labels, tasks read from main, §13's rules and fixtures, the `command`-mode tripwires (§10.1), off until slice 3 switches them on | 7-11 |
   | The ladder and the stop gate: `bonsai.allows`, the floor, `mode`, §13 | 3-5 |
   | Task ids of 4 to 6 digits in `asks.mjs` (`TASK_ID`, `DOC_ID_RULES`) and `contract.ts:1219`; the bridge's parsers and the server's task rows read format 1 | 5-7 |
   | A scan of every tracked frontmatter and YAML file here and in Mimas for format-1 refusals | 1 |
   | This repo's open tasks to format 1 in one commit; new run reports, STATE and ladder results in their new shapes | 1-2 |
   | **Total** | **28-44** (45-70 on the studio's 1.6-times history) |

   > **Changed 8 Oct (Rohan):** the first row (the JSON Schemas and the trick files, 6-10 h) is Bonsai's now, its
   > first job, in Bonsai's `formats/` (`design/plan.md` part 0). The readers task takes a copy pinned to a Bonsai
   > commit instead of writing the set.

   At about 11 estimated hours a studio-day (vision A.12), with the plan's approval and the verifier, roughly a week.
   Mimas runs its pin and reads format 0, so none of this reaches it.
3. **Slice 3 (T-0018):** door 2, the Desk's check, the grant card, pickup (§15.2). Its last step switches this repo to
   `command` mode; from then on the orchestrator and agents move tasks only through door 2.
4. **The machine-side formats arrive with Bonsai's program** (log, asks), right after the gate (Bonsai step 5.2) on
   scratch copies (Q1 c; each given its own id, §3), not rebuilt first in Node: Rohan's "double work". The bridge reads
   spool/1 and `bonsai.log/1` side by side. (Changed 7 Oct from "in the walking skeleton", on Rohan's word: Bonsai
   spec question D (a).)
5. **Trinetra adopts the rest** (§15.2, after the skeleton).
6. **Mimas links last** (Q14 a). Then the spool and `ask.mjs` retire. The executor stays: it is the apply core.

### 15.2 Trinetra adopts the contract (costed)

Agent hours, estimates.

| Part | When | Hours |
|---|---|---|
| Door 2: `trinetra move` and `lane`, the move rules, `--from`, its own recovery, main-only writes, the lock, run from main's copy, `Studio-Via:` and chat's words | Slice 3 | 10-14 |
| The apply core: `recover()` reads the trailer first; the cut tap from any open status; the three new codes through the server's code list and the Desk's wording | Slice 3 | 4-6 |
| The Desk's check: scan `main` from the switch-on commit; match `Studio-Action:` ids to the acts the server issued; "not from a tap" and "from chat"; flag a lost `command` setting | Slice 3 | 6-10 |
| `bonsai.wants`, the grant card and tap, the approve card's wanted paths (only if B is (a) or (c)) | Slice 3 | 4-6 |
| The device on each tap, and its mark on the Desk (only if D is (a)); one tailnet policy line for Rohan | Slice 3 | 2-4 |
| Registration, `trinetra project add .`: reads `status --json`; keeps the root, the Bonsai path and the project-id map on the machine; runs Bonsai's attach and settings commands; adds the studio's deny rules; unregistering undoes each | After the skeleton | 6-10 |
| The apply core and the bridge read format 1 through `status --json`; `bonsai answer` as a journal step; the `applied` events; `bonsai check --write` and the two tables in each move's commit (§10.5, +1-2, 7 Oct) | After the skeleton | 6-11 |
| The forwarder reads Bonsai's log and asks beside today's, from each registered project's `.bonsai/local/` (a Windows-side project's over `/mnt/e`), mapping the workspace id to the project id; the server's task rows read labels | After the skeleton | 10-16 |
| **Total** | | **48-77** (77-123 on the 1.6-times history) |

Pickup itself is costed in T-0018's plan.

## 16. Where the contract meets the re-cut

This is the one text for the seam; the re-cut's §6 restates it. Rohan, 7 Oct: "lets just use one bridge for this
machine, which works on windows and wsl. dont overcomplicate." So the PC is one machine with one bridge (WSL's unit),
serving every project on it, WSL and Windows alike (re-cut §1). The contract owns the formats, the apply core and
registration; the re-cut owns which paths the bridge serves and how it reaches them.

1. **Project ids.** Records carry the workspace id. The bridge maps a checkout to the Desk's project id through its
   registration and stamps it on everything it sends; the server stores no workspace id. Until registration exists,
   the registry's ids are the project ids, and `onThisPC` (re-cut §3.1) decides which projects the bridge serves.
2. **Registration** (`trinetra project add .`, §15.2) is the contract's, after Bonsai's walking skeleton.
3. **A changed workspace id:** the bridge stops forwarding that project until it is registered again (§3).
4. **The PC's bridge applies every tap** through the apply core (door 1), with Windows git for a Windows-side project
   (re-cut §3.4). Bonsai never applies one.
5. **One searched trailer:** `Studio-Action:`, over 30 days. `Studio-Via:` is never searched.
6. **The journal, sidecar and `recover()` stay the studio's**, in its home on the PC (§10.5). Door 2 keeps no journal,
   and Bonsai keeps no act log. Nothing is copied for acts, ever: the proof of a person's move is in git and on the
   server.
7. **Pickup runs on the PC**, on the side the project lives on; a Windows-side project's builder starts on Windows
   (slice 3 decides how). Who may trigger it is question D.
8. **A pinned project** (Mimas): its taps go through the apply core; its own agents run the pin, have no door 2, and
   edit task files themselves until it links (Q14), as an unmanaged project does.
9. **No project format names a machine or a disk path.** A checkout's root appears only in `status --json` (printed
   locally), the registration, and, until then, the studio's own `studio/registry.yaml`.
10. **Records** (logs, asks, ladder results) stay on the PC, in each project's `.bonsai/local/` (§3; since 7 Oct, not
    the Bonsai home). The bridge reads each registered project's `.bonsai/local/` from the paths `status --json` prints
    (§12): a WSL project's directly, a Windows-side project's over `/mnt/e` (or wherever its drive is), as it polls
    Mimas's `studio/` today. The Desk keeps them by project id. **For a Windows-side project the bridge runs Windows
    `bonsai.exe`** (`status`, `answer`, `check --write`), as it runs `git.exe`: Bonsai finds the main checkout's
    `local/` through git, and Linux git never runs in such a checkout.
11. **Later**, when a second real machine exists: the re-cut's §7.

## 17. Decided here, and why

| Decision | Why |
|---|---|
| One `format:` field, first, on every document and record; a missing field means an older writer | One rule for YAML and JSON; additive versions (A.3) |
| Format 0 read exactly as `yaml.mjs` reads it; format 1 by a grammar with nothing left to a library | Frozen files can never be fixed; one meaning per value from now on |
| Core task fields: decision 1's, plus `lane`, `depends_on`, `blocked_by`, the three dates; `bonsai.*` labels for the rest | All exist today; the lane decides whose moves are whose |
| A ladder floor; a task's list only adds | A task could otherwise drop the tests |
| No outside-tracker field | Rohan, 7 Oct: Studio connectors |
| One apply core, two doors (the bridge for taps, a command for agents), writing the main checkout only | The 7 Oct rule's one write path; the Desk sees every move; no merge conflicts on status |
| Managed or not is a Bonsai machine setting | A project without the studio still works; Bonsai names no studio |
| Moves as rules: a person approves, closes `close: person` work, cuts and loosens lanes; agents do the rest | Covers every move used since 29 Sep, Rohan's cuts from any status among them |
| Today's exactly-once kept, plus `--from`, the trailer read first, one lock | Two doors must not undo each other |
| Proof of a person's move: the commit plus the server's record, checked by the Desk; a PC tap is question D | Survives anything a machine loses; a tap is a device, not a person |
| `bonsai.wants` for person-only paths | The grant card needs something to list |
| Bonsai's machine folder keyed by the main checkout's path, holding only machine settings since 7 Oct; records in the project's `.bonsai/local/` | Editing the id or copying the repo cannot loosen a setting; records move with the checkout, into containers and cloud machines, one place per project (Rohan, 7 Oct) |
| Format-0 freezing only for Bonsai's kinds, by LF hashes in the lock | New plans are not findings; shallow CI and Windows pass |
| Notes stay `## From the Desk`; a terminal answer is a record only | The studio may name the Desk (overrides vision §2.1's "never in a project file") |
| Memory D: index and notes in the repo; the personal layer on the machine | Rohan, 7 Oct |
| Bless by Bonsai's runner on a clean base branch, with the new-tests numbers; the floor raised on a tap | A.6 row 12; today's rule; Rohan, 7 Oct |
| The active task: named (the environment only for its own project), else the one running, else none; read from main; grants only while running | T-0011; a guard never guesses |
| The pause lifts when the formats land; slice 3 first; machine-side formats come with Bonsai | Rohan, 7 Oct; "double work" |
| One bridge for the PC; machines later | Rohan, 7 Oct |
| `bonsai update` as a grant needing a person (vision 2.2, 3.4) | Deferred to the Bonsai spec |

## 18. Open questions for Rohan

Four, all director-lane. **All four answered by Rohan on 7 Oct: A (a), B (c), C (b), D (a).**

### A. Task statuses: fixed in Bonsai's core, or set by a pack like lanes? (answered: core)

Lanes went into your workflow pack, with a slot in base (Q6). Statuses are the same kind of thing, and the vision says
nothing about them. The studio checks every move against them (§10.2).
- **(a) Core fixes today's eight.** Every reader knows them, so the checks stay simple. A public user gets plan,
  approved and verify, used or not.
- **(b) Core fixes six kinds** (open, active, review, done, blocked, cut), and your pack names the statuses on top. One
  more mapping every reader must get right.
- **(c) A slot only, like lanes.** Your pack defines every status and which a person sets. Most flexible; the checks
  then run on pack data.

**Answered (Rohan, 7 Oct): (a), today's eight, fixed in core.** Nothing changes for you, no second user exists yet
(Q5 a), and (b) or (c) stay possible later.
*Cost of undoing:* a new task version: open tasks rewritten once, closed ones untouched.

### B. Which protected-path grants need your tap (TB-029) (answered: the short list)

A task lists the protected paths it may change. An agent can grant itself any of them by editing its own task. On 7
Oct, **21 of 45 light tasks and 9 of 25 full tasks** here listed protected paths, most often `studio/game.yaml` (14
tasks) and the guards' two test files (14). The protected list today is `.claude/**`, `studio/game.yaml`,
`studio/ledger.json`, `studio/protocols/**`, the guards' two test files, `docs/PLAN.md`, `docs/research/**` and
`.github/**`. **The guards' own code is not on it** (`tools/hooks/`, `tools/lib/`, `tools/ladder/`): an agent can
switch a guard off by editing its code with no grant at all, and in the main checkout that takes effect at once.
Whatever you choose, a card shows you the paths.
- **(a) Every grant needs a tap.** A task that wants protected paths shows "grant these paths" in Needs you; a
  full-lane task's paths ride on its plan's approve card. About one extra tap for every two light tasks here.
- **(b) Grants count as written, as today.** The Desk lists them after the fact. No extra taps; TB-029 stays open.
- **(c) Split by path.** Only a person-only list needs a tap: `.claude/**`, `studio/game.yaml` (later `bonsai.yaml` and
  the lock), `studio/protocols/**`, `.github/**` and the guards' code and tests. Other protected paths count as
  written.

**Answered (Rohan, 7 Oct): (c), with the guards' own code added to the protected list**, after walking through the cases
(a light task needing `game.yaml`, a full-lane plan, a Mimas design-page edit, a path found mid-task, a guard edit, an
unattended build). Honestly, in this repo nearly every grant
touches that list, so (c) asks about as often as (a) here; in a game project it asks far less. If you choose C (b),
you can also grant in chat. *Cost of undoing:* edit the list.

### C. Approvals, closes and cuts you give in chat (answered: chat counts, marked)

Since 29 Sep, 7 of 9 plan approvals and 7 of 8 full-lane closes came in chat or by hand; only one close came by Desk
tap. Some came in batches: one commit (`7d15635`) closed 8 tasks and cut 3, which would be 11 taps. Your evidence also
says the approval queue is the biggest bottleneck (Approve cards waited about 19 hours, plain questions 3.5 minutes).
From slice 3, in a managed project, task files change only through the studio, and approving, closing full-lane work
and cutting are yours.
- **(a) Desk tap only.** The orchestrator puts a card in Needs you and you tap it: one more step each time you decide
  in chat. A Desk approval also starts the builder by itself.
- **(b) Chat counts, and is marked.** The orchestrator applies it with your words, marked "from chat"; the Desk lists
  each chat move with the session that made it. A dispatched builder cannot make one; the main session could, so it is
  a tripwire, not a wall (your decision D). After a chat approval the orchestrator dispatches the builder itself.
- **(c) Nothing blocks.** Hand edits stay allowed and are only flagged; the flags become noise.

**Answered (Rohan, 7 Oct): (b).** It keeps how you work (no double confirms), and the Desk still shows which were taps and which came
from chat. *Cost of undoing:* one setting.

### D. What a tap from the PC may do (answered: keep it, show the device)

Once slice 3 lands, a Desk tap approves plans, grants protected paths and starts builders. On 6 Oct you let the PC act
on the Desk, accepting that agents in WSL reach it as the PC and could tap too (spec 2026-09-29 §8). Today that only
answers questions; from slice 3, an agent could approve its own plan and start a builder. The re-cut's recheck found
the same gap for pickup, and spec 2026-09-29 §14 Q1 (how a session start from the Desk is authorised) is still open;
this answer closes it for pickup.
- **(a) Keep it, and show the device.** Each tap records whether it came from the phone or the PC (a value in the
  tailnet's act capability, which no process can forge), and the Desk marks person moves tapped from the PC. Seamless
  from both; a tripwire you can see, though your own PC taps get the mark too. One policy line for you, 2-4 agent hours.
- **(b) Person moves from the phone only.** The PC still answers and snoozes, but approve, done, cut, grant and Bless
  need the phone, so no agent on the PC can make them. At your desk you pick up the phone, or decide in chat (C).
- **(c) As today, written down.** "From a tap" means from a device allowed to act. Nothing to build; nothing to see.

**Answered (Rohan, 7 Oct): (a).** It matches how you chose to work (seamless, guards as tripwires), and chat (C b) is already a
tripwire of the same strength. If an agent is ever seen tapping, (b) is one policy change away. *Cost of undoing:* one
policy line.

## 19. Response to the 6 Oct review

The review (`docs/specs/2026-10-06-contract-review.md`: 2 blockers, 8 majors, 18 minors) was answered in the 7 Oct
revision. The recheck walked each answer against the text (its first table): 25 fixed, and M4, M5 and minor 9 only
partly. Those three are finished below, as the recheck's majors 2, 4 and 3. The review's question C was added; its
question D, first marked answered, is open again as §18 D. One suggestion stays declined: moving the field detail into
an appendix (builders read each field where its rule is stated; Rohan's path is §1 and §18 either way).

## 20. Response to the 7 Oct recheck

The recheck (`docs/specs/2026-10-07-contract-recheck.md`) found 2 blockers, 8 majors and 19 minors. All are fixed; none
is declined. Its size suggestions are partly taken (§19 and §17 shortened, §2.7 cut to its rule).

| Finding | Fixed where |
|---|---|
| **B1** A worktree's moves never reach the Desk | Task files change only in the main checkout; door 2 writes main from anywhere; every reader reads tasks from main (§4, §10.1, §13) |
| **B2** No format holds the paths a task asks for | `bonsai.wants`, the derived grant card, the approve card listing wanted paths (§5.5, §5.6, §10.3) |
| **M1** A tap is not proof it was Rohan | Said in §1 and §10.6; reopened as §18 D |
| **M2** A lane downgrade skips the gates | A lane is set at creation; loosening it is a person's move, tightening anyone's; hand edits refused; the Desk flags loosening (§6, §10.2) |
| **M3** Door 2 open to agents for person-only acts | Chat moves refused while a task is named, carrying session and words (§10.6, §18 C); `outside` labels by door 1 only (§5.4); door 2 runs main's copy (§10.1); door 2 moves any kind, a chat approval writes the same two-file commit (§10.1, §10.2) |
| **M4** Exactly-once gaps between two doors | `--from`, door 2's own recovery, refusal while an act is unfinished, `recover()` reads the trailer first, `bonsai answer` as a journal step, a lock with the process start time and a 5-second wait (§10.5) |
| **M5** When the pause lifts; steps 1 and 3 uncosted; the interim | Rohan's 7 Oct rule in §1 and §15.1: the readers task (steps folded in, 28-44 h) lifts it; slice 3 switches on `command` mode; the Desk's scan starts there |
| **M6** The move table forbids 17 real moves | Rules instead of a table: cut from any open status, `running`→`done` in `close: agent` lanes, `blocked`→`verify` (§10.2); the cut tap takes any open status (§10.3) |
| **M7** Grammar gaps | Every line read; sequences deeper than their key; block scalars limited; five escapes; comments by YAML's rule; boolean words refused as keys (`on` renamed `kinds`); 15-digit integers; one-line quotes; ASCII; a trick file per rule (§2.4) |
| **M8** B overstates (c) | B's context lists the protected paths and says the guard code is not among them; (c) names its globs; the recommendation adds the guard code (§18 B) |
| m1 Close count | 7 of 8, with the batch (§1, §18 C) |
| m2 §19 overclaims | Corrected (§19) |
| m3 Pickup wording; many builders | "Stops it merging and pushing"; T-0018 decides parallel builders and sweeps (§1, §10.3) |
| m4 T-0040 | "Mimas's own T-0040" (§2.4) |
| m5 Example uses a closed task | A made-up task (§4.2) |
| m6 Status-line comments | Format 1's status line has none (§4.1, §10.2) |
| m7 `trinetra.cost` writer | No writer: spend is blank on the checklist (§5.4) |
| m8 Studio words in `answer` | `by` and `via` free text (§9.1) |
| m9 `format:` key | Top level only (§2.4) |
| m10 One broken file stops everyone | The guard refuses writing a Bonsai-kind file that fails its format (§10.1) |
| m11 History in shallow checkouts | "History not available", never "passed" (§6) |
| m12 Managed mode lost silently | The bridge flags it (§10.6) |
| m13 Where chat words go | The commit body and the `applied` event (§8.4, §10.5) |
| m14 "Document bodies" | "Documents (frontmatter and body)" (§2.6) |
| m15 Door 2's packaging | Main's copy and dependencies (§10.1); costed (§15.2) |
| m16 §15.2 omissions | The codes, the events and the readers task costed; 1.6-times totals shown (§15) |
| m17 Terms | "The apply core" is the one path; "door 2" the agents' entry (§10.1, §16) |
| m18 `move` and `applied` | `move` dropped: no project moves now (§5.4) |
| m19 "The session's own checkout" | Defined by the session's working directory (§13) |

**The seam.** The recheck's ten re-cut items were written against the two-machine design. Rohan's one bridge for the
PC (7 Oct) made most of them moot, and the rewritten re-cut's §6 restates §16.

## 21. Changed after Rohan's format review (7 Oct)

Rohan reviewed every format himself (`studio/decisions/REVIEW-2026-10-07-formats.md`) and gave his word for these
changes. Each one, with its reason, is listed in the Bonsai spec's §16 ("Changes to the settled contract's text"); his
second look at the new and changed formats is that file's round 2.
- §1's table: where each format lives.
- §2.2: the older-major rule removed (review 1.1); every list defined in one place, with a reference page (2.2, 3.4);
  `bonsai schema` is `bonsai check --schema`.
- §2.4: writers keep comments. §2.8 (new): templates document themselves.
- §3: the project's `.bonsai/` and `.bonsai/local/`; the home keeps only machine settings per workspace; moved and
  copied checkouts; `init` and `status` say where everything is (1.2; round 1).
- §7.2, §7.3: STATE at `.bonsai/STATE.md`; the tables declared. §7.5 (new): the generated tables.
- §8, §8.2, §8.5: the log in `.bonsai/local/log/`; the `clean` event; keeping set in `bonsai.yaml`.
- §9.1, §9.5: asks in `.bonsai/local/asks/`; the outbox gone.
- §10.3: pickup starts `workflow:builder` (the erratum, 6.7). §10.5: the tables in each move's commit. §10.6: the wall
  over `.bonsai/local/`.
- §11: ladder results in `.bonsai/local/ladder/`, cleaned after 7 days (5.1). §12: `home` and `local`. §13: the stop
  gate's result path; the tasks table grants nothing. §14: the lock at `.bonsai/lock.json` (6.2).
- §15.2: the apply core +1-2 h (48-77). §16 item 10: the bridge reads each project's `.bonsai/local/`. §17: the
  machine folder's row.
- After the second recheck (`docs/specs/2026-10-07-bonsai-recheck-2.md`; the orchestrator's technical calls, shown to
  Rohan in round 2): §7.5 stale is a warning everywhere, a sessions row per subagent run with one task rule; §8.1
  `target` on the two start events; §8.5 a log file kept until its rows are in; §9.5 the sandbox fallback; §10.5
  `check --write`'s exit code, a move without the tables; §11 `--ci`'s own file; §13 one more fixture; §15.1 the two
  tables' schemas; §16 item 10 Windows `bonsai.exe` for a Windows-side project.
