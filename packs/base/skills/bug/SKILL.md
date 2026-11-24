---
name: bug
description: "Bugs, Bonsai's own documents from Bonsai 1.1.0: the bug file template (bonsai.bug/1), every field and every bonsai bug subcommand, the generated bugs page, and how a project moves an old hand-kept register to bug files. Use it to file a bug, move or change one, read the list, or move a register kept by hand."
---

# Bugs

A bug is one defect someone has **seen or proven**: what goes wrong as a person sees it, where, who found it and where it
stands. Each bug is **its own file** in the format `bonsai.bug/1`, named after its id (`BUG-012.md`), in the project's
bug folder. A page that lists them all, `bugs.md`, is generated from the files and never edited by hand. Bugs need
Bonsai 1.1.0 (`bonsai --version`).

**When to use this skill.** To file a bug, to mark one fixed, wontfix or watched, to reopen one, to change its text, to
list or read them, or to move an old register that a project kept by hand (the last section). The same fields and values
are printed by `bonsai check --schema bonsai.bug`.

**`bonsai bug` is the one way a bug's status changes, in every project.** A hand edit of a bug file's status line is
refused by Bonsai's guard, and so is a hand edit of the generated page. `status_writes` in `bonsai.yaml` is for tasks
only; it never applies to a bug. Every `bonsai bug` write goes to the **main checkout's** folder wherever you run it, under
a lock, and is recorded in the log with who made it. A design question or a taste is not a bug: it goes to the person.

**Where they go.** The folder `bonsai.yaml` names under `documents.bug`; when the line is absent, `bugs/` beside the
page. The page is the file under `documents.bugs` (`work/bugs.md` when both are absent, with the folder `work/bugs/`).
`bonsai status --json` lists both under `documents`. Both lines are in the template `bonsai init` writes; a project
linked before 1.1.0 has neither and gets the defaults, which is usually right (`bonsai.yaml` is a person's file: a line
there is a task that grants it).

## Fields

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `format` | The format and its version, always the first line; its comment points back to these docs | `bonsai.bug/1` | `bonsai.bug/1` |
| `id` | The bug's id, unique in the project and never reused; the file is named after it. One prefix in every project; an old id from a register with another prefix is kept in `was`. Written plain, never quoted | `BUG-` and 3 to 6 digits, the pattern `^BUG-[0-9]{3,6}$` | `BUG-012` |
| `title` | What goes wrong, as a person sees it, in plain words, on one line. A title holding `: ` (a colon and a space) or starting with a quote is written in quotes, or the file does not read (`bonsai bug` does this for you) | text, one line | `The settings page forgets the dark mode switch` |
| `status` | Where the bug stands: `open` is seen or proven and not fixed; `fixed` has a commit that fixed it; `wontfix` is known, not a bug, or not to be fixed, with its reason in `why`; `watched` was seen once and not repeated (a flaky test among them), with its reason in `why`. A new bug starts at `open`. Moves: open to fixed, wontfix or watched; watched to open or fixed; fixed and wontfix back to open. Its line carries no comment | `open`, `fixed`, `wontfix`, `watched` | `fixed` |
| `area` | The part of the project it is in, a few words, as the project names its parts | text, one line; `null` for none | `settings` |
| `where` | Where it goes wrong: a file and line, a command, a screen; add how sure the finder was where it helps (pointed: the line was read; said: only a report says so) | text; `null` when not known | `src/settings/store.ts:88 (pointed)` |
| `found_by` | Who or what found it: a role and task, a message id, a person, a CI run | text; `null` when not known | `verifier@T-0901 M-2026-10-09-1a2b3c4d` |
| `filed_by` | Who wrote this file: an agent's address (`role@task`, or a role), or `person`. Set by `bonsai bug add` | text; `null` for a bug carried over that did not say | `builder@T-0901` |
| `fixed_by` | Who marked it fixed, in the same form. Set by `bonsai bug fix` | text; `null` while it is not fixed | `builder@T-0904` |
| `fixed_in` | The commit or commits that fixed it, then a few words on how where that helps. Set by `bonsai bug fix`, which checks each commit exists | text; `null` while it is not fixed | `"0a1b2c3d (the switch is saved before the page closes)"` |
| `task` | The task that fixed it, or the task it is being fixed under | a task id, `T-` and 4 to 6 digits; `null` for none | `T-0904` |
| `was` | Its id in the register it was carried over from, when that id was another (`TB-012` became `BUG-012`); `bonsai bug show` finds a bug by it | text; `null` for a bug filed here | `TB-012` |
| `why` | Why it is `wontfix` or `watched`, one sentence; `null` for an open or fixed bug unless a reason helps | text; `null` for none | `A deliberate over-refusal, documented in the guard's tests` |
| `created` | The day the bug was filed: the UTC day, from `bonsai now --day`. A bug carried over keeps the day its old row named, else the day of the move | a date, `YYYY-MM-DD` | `2026-10-09` |
| `closed` | The day it was fixed or set `wontfix` (a move writes it): the UTC day. A fixed or wontfix bug with no day is allowed only with `was` set (an old row that named none): `bonsai check` holds that | a date, `YYYY-MM-DD`; `null` while it is open or watched | `2026-10-10` |
| `labels` | Everything else, as a map of label name to value | a mapping of label names to values; `{}` for none | `{}` |

A bug file as `bonsai bug` writes it: every field, in this order, `null` where it does not apply, then a free body (how
to repeat it, what was seen, notes). Never write one by hand; the guard refuses a new file that starts at another status
than `open`.

```markdown
---
format: bonsai.bug/1   # fields: bonsai check --schema bonsai.bug; how to fill: skill base:bug
id: BUG-<number>
title: <what goes wrong, as a person sees it>
status: open
area: <the part of the project, or null>
where: <file and line, command or screen, or null>
found_by: <role and task, message id, person or CI run, or null>
filed_by: <your address, or person>
fixed_by: null
fixed_in: null
task: null
was: null
why: null
created: <YYYY-MM-DD>
closed: null
labels: {}
---

<How to repeat it, and what was seen.>
```

## The commands

Every subcommand takes `--json` (the `bonsai.buglist/1` document) and a refusal names its next step. In an agent's
session, **`--by <role>@<task>`** (the address your brief gives you, or a role; `person` only in a person's terminal)
is required on a write that names who writes: add, fix, wontfix, watch, reopen. `bonsai bug <action> --help` lists every
flag.

| To | Command |
|---|---|
| File a bug (next free id, status `open`, created today) | `bonsai bug add --title "<what goes wrong>" --area "<part>" --where "<file:line>" --found-by "<who or what>" --by <role>@<task>` |
| Give its body | `--body "<text>"`, `--body -` (from standard input) or `--body-file <path>`, on `add` or `edit` |
| Mark it fixed | `bonsai bug fix <id> --commit <sha> --how "<a few words>" --task <task> --by <role>@<task>`; `--commit` again for more; a commit the repository lacks is refused |
| Mark it wontfix | `bonsai bug wontfix <id> --why "<why>" --by <role>@<task>` |
| Mark it watched | `bonsai bug watch <id> --why "<why>" --by <role>@<task>` |
| Reopen it (it came back, or watching is over) | `bonsai bug reopen <id> --why "<why>" --by <role>@<task>`; clears `fixed_by`, `fixed_in` and `closed` |
| Change its text (never its status) | `bonsai bug edit <id> --where "<file:line>"`; any of `--title`, `--area`, `--where`, `--found-by`, `--task`, `--why`, `--body`; `""` sets a field to none. **`--body` replaces the whole body**: read it with `show` first to add to it |
| List them, oldest id first | `bonsai bug list --json`; `--status open` (again for more) keeps some |
| Read one | `bonsai bug show <id> --json`; the id may be its old id (`was`) |

Plain words first in a title; a source someone can open in `found_by`; where the line was read, name the function too
(code moves). The next id is the highest number in the folder plus one, never one a file holds.

**The page.** `bonsai check --write` builds the bugs page (`bonsai.bugs/1`) from the files: a table per status, Open,
Watched, Fixed and Won't fix, newest id first. It writes the page **only** when it is missing or carries its generated
marker (its `format: bonsai.bugs/1` line); a hand-kept register has no marker, is never touched, and `bonsai check`
names it with the warning `bugs-not-moved` until it is moved (below). The page lags between moves by design (the warning
`tables`); `bonsai check --write` brings it up to date. Commit the bug files and the page together.

## Moving an old register

A project that kept its bugs in one hand-written file (a table of rows, as the workflow pack's old `bugs` kind or the
project's own) moves it once, and **its own agent does it**: there is no import command, since every register's columns
differ. The old page stays exactly as it is until the last step. The old ids are kept: `TB-012` or `B-012` becomes
`BUG-012` with `was: TB-012` (an old `BUG-012` stays `BUG-012`, with `was` the same), so every link and commit message
that names the old id still finds it through `bonsai bug show TB-012`.

1. **Read the whole old page first** and list what is in it: every row of every table, each area heading, the lists
   that are no table, and any text under a table. Count the rows. Work in the main checkout (writes land there).
2. **Map the columns once.** Whatever the headings are called: the id column is the old id; the column that says what
   goes wrong in a person's words (`What goes wrong`, `What a player sees`, a fixed table's `Was`: the column named `Was`
   there is the text, not `was` the field) is `--title`; `Where` is `--where`; `Found by` is `--found-by`; a task column
   is `--task`; a `Fixed in` cell is `--commit`. An area heading above a table (`### the guard`) is `--area` for every
   row under it. A cell that is long (more than 200 characters or more than one sentence), or holds several lines or links,
   goes whole into the body (`--body-file`), with a short title (its first sentence, cut at a word if there is no
   sentence break; aim for under 200 characters): the title is one line and nothing is cut. **`bonsai check` reads a
   bug's title and every field for absolute paths** (a home folder, a drive path): reword that text with `~/...` or
   `%USERPROFILE%`, and keep the old words whole in the body.
3. **Every row by one `bonsai bug add`.** An id's prefix and number: `TB-012` is `--was TB-012 --number 12`. A row of
   an Open table:

   ```bash
   bonsai bug add --by <role>@<task> --was TB-012 --number 12 --title '<what goes wrong>' \
     --area '<heading>' --where '<where>' --found-by '<found by>'
   ```

   A row of a **Fixed** table (or a row that says it was fixed) carries its status, its commit and its days where the
   row names them, with `--commit` as the cell gave it (a sha, with its words if it has them; one that names no commit
   is kept as text; a sha this repository does not hold is accepted here, since only `fix` checks), `--created` and `--closed` the days the row names *as that bug's own opened and closed days* (a day the row
   names for something else, such as the day a rule began or when a failure started, is not one: it stays in the body;
   with no such day, `--created` is the day of the move and there is no closed day):

   ```bash
   bonsai bug add --by <role>@<task> --was TB-012 --number 12 --status fixed --commit '0a1b2c3d (the words)' \
     --created 2026-10-09 --closed 2026-10-10 --title '<what went wrong>' --where '<where>' --found-by '<found by>'
   ```

   A task cell that is not a task id (`before the register`, `task not read`, `T-55xx (5.5's labels)`) has no
   `--task`: pass `--task` only for a real id, and put the rest of the cell in the body.

   A row in a list of **Watched** bugs that has an old id is `--was <id> --number <n> --status watched --why '<its
   reason>'`; a row of **Won't fix** the same with `--status wontfix --why '<its reason>'`. Always give `--why` for
   those two. `--status` needs `--was`, so a row with no old id is filed as step 4 says (add it, then `bonsai bug
   watch <the printed id> --by <role>@<task> --why '<its reason>'`). A title or cell with a quote goes in
   single quotes, a single quote inside as `'\''`. In a table cell `\|` is a plain `|`: write it so. Markdown in a cell (backquotes, bold) is kept as it is, minus the bold on the id.
4. **The list "Known, not bugs"** (behaviour that looks wrong and is meant). An item that carries an old id is a
   carry-over as won't fix, keeping that id: `--was <id> --number <n> --status wontfix --why '<why it is meant>'`, as
   in step 3. An item with no id takes a free number, so file each after every numbered row is in, then move it,
   so it keeps its reason:

   ```bash
   bonsai bug add --by <role>@<task> --title '<the behaviour>' --area '<heading>'
   bonsai bug wontfix <the printed id> --by <role>@<task> --why '<why it is meant>'
   ```

5. **Text that is no row.** A note under a table about one bug goes into that bug's body (`bonsai bug edit <id>
   --body-file <path>`, written with its old body first if `show` printed one). A note about the register as a whole,
   or the intro that only explained how to use the old page, goes into a memory note (`base:memory`) when it holds a
   fact a later session needs, and is dropped when the generated page's own intro says it. The mover's task must
   allow the memory folder; if it does not, the mover hands the note's text to its caller in its report, and the old
   page is not deleted (step 7) until the note is filed. **Nothing else is dropped:**
   if you cannot say where a line went, it is not moved.
6. **Check that nothing is missing.** Every old id must have a file, and the counts must meet:

   ```bash
   for id in <every old id, as found in step 1>; do bonsai bug show "$id" --json > /dev/null || echo "missing: $id"; done
   bonsai bug list --json
   ```

   A row carried over as fixed by `add` has no `fixed_by`: an old row does not say who fixed it, so leave it empty
   (or name the person in the body if the row does).

   The files hold one bug per old row plus one per "Known, not bugs" item, and nothing is printed as missing. Read two or
   three files (`bonsai bug show <id>`) against their old rows, one of each shape.
7. **Only then delete the old page and build the new one.** Delete the old page (`rm <documents.bugs's file>`; git
   keeps it in history), run `bonsai check --write` (it writes the generated page at the same place, so links keep
   working), then `bonsai check`: it prints no finding and no `bugs-not-moved`. Commit the bug files, the page and the
   deleted old text together, saying what moved.

If a step is refused, its message names the next one. A number already taken (`bug-id-taken`) means two old rows share
a number or the row was moved already: look before you go on. Never start a second move over a half-done one: `bonsai
bug list` tells what already arrived.
