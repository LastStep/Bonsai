---
name: note
description: "The note template: every field of a note, what it may hold, and the template to copy. Use it to write a new note or to check one."
---

# The note template

The template of the document kind `note`, which this pack declares in `bonsai/pack.yaml` (its folder, id pattern,
statuses and moves). A note records one thing worth coming back to: a decision and why, a finding, or an open
question. Write one when you are asked to note something, or when the skill `example-pack:example` says to.

**Where it goes:** one file per note, named after its id (`N-0001.md`), at the top of the folder the project's
`bonsai.yaml` gives the kind `note` under `documents` (`notes` unless the project changed it). Copy the template
below, fill every field as the table says, and keep only the pointer comment on the first line: the docs stay here.

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `id` | The note's id, unique in the folder; the file is named after it | `N-` and four digits, the pattern `^N-[0-9]{4}$` in `bonsai/pack.yaml` | `N-0001` |
| `title` | What the note is about, in one line | text, one line | `Why notes live in notes/` |
| `status` | Where the note stands; a new note starts at the first status, and an agent may move it from draft to done | `draft`, `done` | `draft` |
| `task` | The task the note belongs to, by its id | a task's id, or `null` for none | `T-0001` |
| `finished` | The day the note reached `done`, set then | a date, `YYYY-MM-DD`, or `null` until then | `2026-10-15` |
| `labels` | The note's labels, by name | a mapping of label names to values; `{}` for none | `{}` |
| `labels.example-pack.priority` | How soon a person should read the note (defined in `bonsai/labels.yaml`) | `low`, `normal`, `high` | `normal` |

```markdown
---
id: N-0000   # a note's id; every field and how to fill it: skill example-pack:note
title: <what the note is about, in one line>
status: draft
task: null
finished: null
labels:
  example-pack.priority: normal
---

# <the title again>

<What is noted, in plain words: the decision and why, the finding, or the question.>
```
