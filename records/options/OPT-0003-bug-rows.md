---
id: OPT-0003   # an options write-up; every field and how to fill it: skill workflow:options
title: How should Bonsai give a project's bug rows as JSON, and read a register that drifts from the template?
lane: director
status: decided
author: orchestrator
created: 2026-11-19
decided: 2026-11-19
chosen: "a bugs system (the sketch with the director's notes); BUG- ids only; no import"
---

# How should Bonsai give a project's bug rows as JSON?

## Why this is being asked now

The studio asked on 19 Nov (task `T-6023`). It reads every linked project through `bonsai status --json`, and wants
to see "bug X filed" and "bug X fixed". Today Bonsai names only the register's file, so the studio parses the
markdown itself. That means knowing the register's shape, which belongs to Bonsai and the workflow pack. The contract
says the opposite today (`design/contract.md` 7.3: "Rows the studio parses out of a pack document's body (the bugs
table) are not contract"), so changing it is your call.

## What is true today

- The `bugs` kind is the workflow pack's (`bonsai/pack.yaml`): one file, no id, no statuses. Its shape is in its
  template only (`workflow:bugs`): an Open table `Id | What goes wrong | Where | Found by`, a Fixed table `Id | What
  went wrong | Fixed in | Task`, ids `BUG-` and three digits.
- Bonsai's code names no pack's kind (one home for every list: the pack's declaration). A reader of the rows needs
  the shape declared by the pack, so the work has three parts: a new field in the pack format, the workflow pack
  declaring it, and Bonsai's reader.
- `status --json` takes 0.26 s in Bonsai's repo (it was 8 s before BUG-137's fix, 1.0.3). The studio calls it every
  few seconds.
- Real registers drift. The studio's has `###` area headings under Open and a Fixed table `Id | Was | Fixed in`.
  Another project uses `TB-NNN` ids, another `B-NNN` with `What a player sees`. All of them keep the template's
  column order: id first, then what goes wrong, then where (or fixed in).

## Question 1: where the rows live

| | A. Its own command (recommended) | B. Inside `status --json` | C. In the cache Bonsai keeps |
|---|---|---|---|
| What | `bonsai rows bugs --json`: the rows of any one-file kind whose pack declares rows | Each `documents` entry of such a kind carries its rows | Status keeps the rows in `.bonsai/local/cache/`, read again only when the file changes |
| Speed for the studio | Status stays as it is; the studio already has the file's path and calls `rows` only when the file changes | Every poll parses the register (a few ms for 200 rows) and sends every row again | Cheap after the first read, but adds a cache: BUG-144 and 145 are bugs of the cache we have |
| Cost to build | Smallest: one command, one schema | Changes status's format, which every consumer reads | The most: B plus a cache |
| Undo | Easy: drop a command | Hard: consumers come to rely on status's shape | Hard |

## Question 2: a register that drifts from the template

| | A. Read by order, warn on drift (recommended) | B. Rows only for a register that fits | C. Read by the table's headers |
|---|---|---|---|
| What | Rows are read by the template's column order under `## Open` and `## Fixed`; any id like `TB-012` or `B-7` counts; a `###` heading is the area. `bonsai check` warns (`bugs-drift`, exit code unchanged) naming each header that differs from the template, with the next step | `check` fails on drift; `rows` refuses until the register fits | The pack declares, per field, the header names it accepts; a project adds its own |
| The studio's three registers | Give rows today; the warning tells each project what to move | Give nothing until each register is moved by hand | Give rows once each project's names are added |
| Risk | A register whose columns are in another order reads wrong (the warning says so) | CI goes red in projects that worked | The list of names grows with every project; drift is never seen |

## Question 3: the release

Three releases, each tagged by you: the pack format and Bonsai's reader in **Bonsai 1.0.4**, the `bugs` kind's
declaration in **workflow v1.0.3**, and **base 1.0.4** for its pins (base's ci skill downloads Bonsai by version).
The studio's projects then move their packs (`bonsai pack set --all`). You said on 6 Nov "we are not putting 1.0.4
just yet", so the timing is yours: ship it on its own once built, or hold it for the next batch.

## A row, if A and A are chosen

```json
{ "id": "BUG-144", "state": "open",
  "title": "The sessions cache's version number is not tied to the build: ...", "area": null,
  "where": "internal/sessions/cache.go (said by the verifier, not run)",
  "found_by": "verifier@T-6016 `M-2026-11-06-fb3e78c1`", "fixed_in": null, "task": null, "file": "records/bugs.md" }
{ "id": "BUG-139", "state": "fixed", "title": "`status --json`'s `bonsai_release` reads `newest` 1.0.0 ...",
  "area": null, "where": null, "found_by": null, "fixed_in": "fee1e13a (a record whose newest is ...)",
  "task": "T-6010", "file": "records/bugs.md" }
```

(Both from Bonsai's own register, cut short here with "...".) Found while writing this: BUG-137 still sits in Open
although STATE calls it fixed by T-6016; a row for the next records commit to move.

Open rows fill `where` and `found_by`, and leave `fixed_in` and `task` null. Fixed rows fill `fixed_in` and `task`
(the template's Fixed table has no where or found by). `task` is the one field the studio did not list: the
template's Fixed table keeps the commit and the task in two columns. Cells are given as written, with the bold taken
off the id.

## Recommendation

1A, 2A, and 1.0.4 on its own once it is built and verified. The studio gets rows from every register it reads today.
Each drifting project is told what to move without breaking anything. Status stays as fast as it is.

## The director's first answers (19 Nov)

- Question 1: "instead of doing some banage kind of fix, we should build this system properly. like having proper
  commands for working with bugs, tell me what that would look like"
- Question 2: "my reply to question 1 might change this, so i will hold off for that"
- Question 3: "Ship when ready"

## The proper system, sketched (for the director's read)

Bugs become documents like tasks: **one file per bug** (`records/bugs/BUG-145.md`, a frontmatter of id, status,
title, area, where, found_by, fixed_in, task, dates; a body for how to repeat it). Their format is Bonsai's own
(`bonsai.bug/1`), not a pack's. The one-page register stays, generated from the files by `bonsai check --write` as
the tasks table is, so a person still reads one page and a link to it still works. Two agents filing at once no
longer clash on one table or on the next free id.

| Command | What it does |
|---|---|
| `bonsai bug add "<what goes wrong>" --where .. --found-by .. [--area ..]` | Takes the next free id, writes the file, prints the id |
| `bonsai bug fix <id> --commit <sha> [--task <id>]` | Open to fixed; refuses a commit the repo does not have |
| `bonsai bug wontfix <id> --why ".."` | Open to "known, not a bug", with the reason |
| `bonsai bug reopen <id>` | Back to open (it came back) |
| `bonsai bug edit <id> --title/--where/--area ..` | Changes a field |
| `bonsai bug list [--open/--fixed] [--area ..] --json` | The rows: the studio's ask |
| `bonsai bug show <id> --json` | One bug in full |
| `bonsai bug import [--yes]` | Once per project: reads today's markdown register, drift and all (by column order, any `XX-NNN` id, `###` areas), shows what it will write, then writes the files |

Status and every command: `--json`, unattended, each refusal naming its next step; `bonsai check` holds each file to
its schema and the register to the files. Drift (question 2) is met once, at import, and cannot come back: the
commands write the shape.

What it touches: the contract (a new kind and format; 7.3's line on rows goes), the workflow pack's `bugs` kind and
skill (they move to Bonsai and base), and a principle: Bonsai "never moves a task" (contract 10, the studio applies
moves). A bug's move is not a task's and needs no person's approval, as `bonsai message` already writes its own
records, but it is the first status Bonsai's own command moves.

## The director's notes on the sketch (19 Nov), and the orchestrator's read

Six notes, each with the orchestrator's verdict (the chat of 19 Nov explains each in full):

1. **Each project keeps its id prefix** (TB-012, then TB-081), set in bonsai.yaml: have it. Import sets it from the
   register; a register mixing prefixes is refused at import.
2. **Import loses nothing** ("Known, not bugs" to wontfix, notes under tables placed, the preview lists what it
   cannot place): have it, and import writes nothing while anything is unplaced. Bonsai's own register also has
   "Watched, not bugs (flaky)": a fourth status `watched`, or a wontfix with a reason (a question for the plan).
3. **Every move logged with who** (filed by builder@T-0108): have it, and the file keeps `filed_by` and `fixed_by`
   too, since the log is per machine, never committed and cleaned after its keep days. Caveat: a subagent shares its
   session's environment, so "who" is as good as `BONSAI_ROLE` and `BONSAI_TASK` say.
4. **Bug moves follow status_writes, as tasks do**: have it. Once `bug` is a declared kind with statuses, contract
   10 covers it as written: under `command` the studio's move command moves it, and the guard refuses a hand edit of
   its status line. `bonsai bug fix` then writes only its fields (fixed_in, task) and names the move command; under
   `agents` it writes the move as well.
5. **The generated page marked generated, its place said**: have it, at the old path (`records/bugs.md`, cited in 15
   run reports), declared in status as a generated kind like `tasks`.
6. **status --json lists the bug folder with its path and statuses; bug list --json is for agents**: have it; it
   comes free with the declared kind and costs status nothing.

## The director's choice (19 Nov 2026)

- "all looks good, i just dont fully understand the prefix thing, tell me more about it"
- "yes we will go with bug prefix only, as we want to keep things uniform. and i dont think you need to do any kind of
  importing for the project, but rather i will tell each project agent myself to do the migration"

So: the sketch with notes 2 to 6 as read above; note 1 replaced by one prefix, `BUG-`, each old bug keeping its
number with its old id in `was`; no import command (each project's agent moves its register). The plan:
`records/plans/P-T-6023.md`.

Later the same day, on the plan's choices: "i dont want bonsai to use the studio, as bonsai is a core standalone
technology"; then "always bonsai bug" (bug moves only by `bonsai bug`, in every project, `status_writes` for tasks
only: note 4 replaced); Bonsai's own register moved by the orchestrator after the release; `watched` kept.
