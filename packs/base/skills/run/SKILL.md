---
name: run
description: "The run report template (bonsai.run/1): every field of a run report, its allowed values and the template to copy. Use it when a session starts work on a task, to open the session's report, and to check one."
---

# The run report template

A run report is the log of one agent session on one task: opened before the session's first change and added to as
the work goes, never written at the end from memory. It is one markdown file in the format `bonsai.run/1`: a short
frontmatter, then the body below. A verifier and the person read it to see what was done and how it was proved.

**When to use this skill.** At the start of a session that works on a task, to open its report; and to check one.
The same fields and values are printed by `bonsai check --schema bonsai.run`.

**Where it goes.** One file per session in the folder `bonsai.yaml` names under `documents.run` (`bonsai status
--json` lists it under `documents`), named after its id, with a slug if you like: `R-2026-10-15-T-0042.md` or
`R-2026-10-15-T-0042-dark-mode.md`. It is committed on the task's branch, beside the work it describes. Copy the
template below and keep only the pointer comment on the `format:` line: the docs stay here.

## Fields

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `format` | The format and its version, always the first line; its comment points back to these docs | `bonsai.run/1` | `bonsai.run/1` |
| `id` | The report's id: `R-`, the day the session started, and the task's id; add a suffix when a task has two reports on one day | text starting with the pattern `^R-[0-9]{4}-[0-9]{2}-[0-9]{2}-T-[0-9]{4,6}` | `R-2026-10-15-T-0042` |
| `task` | The task the session works on, by id: the link between the two (the task holds no pointer back) | a task id | `T-0042` |
| `role` | The role the agent runs as: a role a linked pack declares (`bonsai status --json` lists them), or the plain word for the session's job where none is declared | text | `builder` |
| `model` | The model the agent runs on, as the agent names it | text | `sonnet` |
| `started` | When the session started, in UTC: write what `bonsai now` prints as you open the report | `YYYY-MM-DD HH:MM`, as text | `2026-10-15 14:08` |
| `finished` | When the report was closed, in UTC: what `bonsai now` prints then | `YYYY-MM-DD HH:MM`, or `null` while it runs | `null` |
| `outcome` | How the session ended: `running` while it runs; `verify` when its work waits for its proof; `needs-verifier` when it waits for a verifier; `needs-person` when it waits for a person's word; `merged` when the work landed; `blocked` when it cannot go on; `abandoned` when the work was dropped | `running`, `verify`, `needs-verifier`, `needs-person`, `merged`, `blocked`, `abandoned` | `running` |
| `commits` | The commits the session made, each hash quoted | a list of quoted commit hashes; `[]` for none | `["1a2b3c4"]` |
| `labels` | Labels a linked pack defines for run reports, by name | a mapping of label names to values; `{}` for none | `{}` |

```markdown
---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-<YYYY-MM-DD>-T-0000
task: T-0000
role: <the role you run as>
model: <the model you run on>
started: <YYYY-MM-DD HH:MM>
finished: null
outcome: running
commits: []
labels: {}
---

# Run: <the task's title>

> Opened at the start of the session and added to as it goes. Not written at the end.

## What is different now

One paragraph, in plain words, that a person can read on a phone.

## Deviations from the plan

**First**, if there are any. What you did instead of what was approved, and why. "None" is a fine answer and should
be the usual one.

## Ladder

The proof: `bonsai ladder --task <task id> --json`, whose result Bonsai keeps in the main checkout's
`.bonsai/local/ladder/<task id>.json` (a CI run writes `ci.json` there, stamped `mode: ci`, never a task's proof).
Quote the result's `green`, its `git.sha` and each rung's line:

| Rung | Check | Result | Evidence |
|---|---|---|---|
| 0 | the task's paths | green | no path outside the task's grant |
| 1 | build | green | exit 0 |
| 2 | tests | green | 287 passed, 0 failed, 0.24 s |

Paste real output. A rung with no evidence counts as not run.

## What did not work

The thing tried and dropped, the rung that went red twice before it went green, the part that is slower or uglier
than it should be. If this section is empty, the report is not finished.

## Evidence for a human

Screenshots (where the project keeps them), the address to open, the command to run to see it yourself.

## Files changed

Grouped by area, one line each on why. Not a `git diff --stat` dump.

## Processes left running

None. Say how you checked (`ps -eo pid,etime,cmd`; on Windows `tasklist`). If something is left, list it with its
process ids and why.

## What it cost

The session's cost where it is known, and where the figure comes from; "not known" otherwise. The hours are the task's
rows in `.bonsai/sessions.md` (Bonsai's sessions table, from the log): name each run's message id and model, and type
no table of run times by hand.

## Left for next time

Concrete, with a task id if one was written.

## Verifier

> Added by the verifier, in a fresh context. Pass or fail, with reasons.
```

**Times.** Every time and day you write comes from `bonsai now` (`--day` for a day), in UTC, never from memory:
the guard refuses a fresh one far from Bonsai's clock (`time-not-now`), and `bonsai check` finds one ahead of it, however
it was written (`future-stamp`). A `finished` set later is still the time you set it, taken from `bonsai now` then. Times written before formats set 11 were local, their offset unknown: leave them.

**Before you finish the session:** set `finished` and `outcome`, list every commit in `commits`, and check that
`bonsai check --json` reports no finding for the file.
