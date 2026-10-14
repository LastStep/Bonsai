---
name: task
description: "The task template (bonsai.task/1): every field of a task, its allowed values, Bonsai's four labels and the template to copy. Use it to write a new task, or to check one."
---

# The task template

A task is one unit of work: what it is, where it stands, and when it is done. It is one markdown file in the format
`bonsai.task/1`: a frontmatter of fields, then a body for whoever starts on it cold. Bonsai reads tasks for its checks,
its guard (what a running task may change), the ladder (what proves it done) and its tables.

**When to use this skill.** Before writing a new task, and to check one: every field below, each value one it allows.
The same fields and values are printed by `bonsai check --schema bonsai.task`.

**Where it goes.** One file per task in the folder `bonsai.yaml` names under `documents.task` (`bonsai status --json`
lists it under `documents`), named `<id>-<slug>.md`: `T-0042-dark-mode.md`. Copy the template below, fill every field
as the table says, and keep only the pointer comment on the `format:` line: the docs stay here. A field with no value
yet is written `null` or `[]`, never left out.

## Fields

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `format` | The format and its version, always the first line; its comment points back to these docs | `bonsai.task/1` | `bonsai.task/1` |
| `id` | The task's id, unique in the project; written plain, never quoted; the file's name starts with it | `T-` and 4 to 6 digits, the pattern `^T-[0-9]{4,6}$` | `T-0042` |
| `title` | What gets built, in one line | text, one line | `Add a dark mode switch to the settings page` |
| `status` | Where the task stands; a new task starts at `todo`. Who makes each move is the project's: where `bonsai status --json` reads `status_writes: command`, a move goes through its `status_command`; otherwise an agent edits this line, making only the moves agents may make. The line carries no comment | `todo`, `plan`, `approved`, `running`, `verify`, `done`, `blocked`, `cut` | `todo` |
| `lane` | The task's lane, where a linked pack defines lanes (a lane decides which moves are a person's) | a lane a linked pack defines (`bonsai status --json` lists them), or `null` where none does | `null` |
| `done_when` | Checkable statements a verifier can test without having built the work | a list of text, each quoted; `[]` until written | `["The settings page shows a dark mode switch"]` |
| `depends_on` | Tasks that must be done first, by id | a list of task ids; `[]` for none | `[T-0041]` |
| `blocked_by` | What blocks the task, in one sentence, while its status is `blocked` | text, or `null` otherwise | `null` |
| `created` | The day the task was written | a date, `YYYY-MM-DD` | `2026-10-15` |
| `started` | The day work started, set when the task first reads `running` | a date, or `null` before | `null` |
| `finished` | The day the task reached `done` or `cut` | a date, or `null` before | `null` |
| `labels` | Everything else, by label name: Bonsai's four below, and any label a linked pack defines (the block in `CLAUDE.md` lists each with its kind) | a mapping of label names to values; `{}` for none | as in the template |
| `labels.bonsai.branch` | The branch the task is built on | text matching `^[A-Za-z0-9._/-]{1,100}$`; leave the label out to build on the base branch (an empty or null value is a check finding) | `t0042-dark-mode` |
| `labels.bonsai.ladder` | The rungs this task climbs beyond the project's floor; it adds rungs and never drops one | a list of rung numbers; `[]` for the floor alone | `[]` |
| `labels.bonsai.allows` | The protected paths this task may change, as globs; the guard grants them only while the task reads `running`. A path on the person-only list counts only when a person granted it | a list of globs from the project's top, forward slashes; `[]` for none | `["assets/themes/**"]` |
| `labels.bonsai.wants` | The person-only paths the task asks for; a person's approval or grant copies them into `bonsai.allows` | a list of globs, as `bonsai.allows`; `[]` for none | `[]` |

```markdown
---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-0000
title: <one line: what gets built>
status: todo
lane: null
done_when:
  - "<a checkable statement a verifier can test without having built it>"
depends_on: []
blocked_by: null
created: <YYYY-MM-DD>
started: null
finished: null
labels:
  bonsai.branch: <the task's branch; leave this line out to build on the base branch>
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# <the title again>

## Context

What an agent starting cold needs that is not already in `.bonsai/STATE.md` or another of the project's documents.
Point to them; do not copy them. Three sentences, not three pages.

## Scope

What this task changes. Name the files or areas, if you know them.

## Out of scope

What it leaves alone on purpose, above all anything nearby that a builder would be tempted to fix while there.

## Notes

Anything learnt that the next agent would otherwise find out again. Added to as the work goes.
```

**Before you finish:** every field in the table is in the frontmatter, in the template's order; `id` is new in the
folder and starts the file's name; each `done_when` line can be checked by someone who did not build the work; and
`bonsai check --json` reports no finding for the file.
