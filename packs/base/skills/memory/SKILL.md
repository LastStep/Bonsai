---
name: memory
description: "The memory note template (bonsai.memory/1): every field of a note, its allowed values, the template to copy, and the index line. Use it to save a fact worth keeping across sessions, to change or drop one, or to check the notes and their index."
---

# The memory note template

The project's memory is a folder of notes, one fact each, and an index, `INDEX.md`, one line per note. The block in
the project's `CLAUDE.md` imports the index, so every session sees the lines; a note itself is read when its line
says it matters. Notes and index are in the format `bonsai.memory/1`, committed, and travel with the repository.

**When to use this skill.** When you learn something a later session would otherwise have to find out again: a fact
about the project, how the person wants the work done, or where something is found. Also to change a note that has
gone stale (rewrite it and its index line) or to drop one (delete the file and its line). The same fields and values
are printed by `bonsai check --schema bonsai.memory`.

**Where it goes.** In the folder `bonsai.yaml` names under `documents.memory` (`bonsai status --json` lists it under
`documents`): one file per note, named after its id (`M-fewer-verifications.md`), and the index beside them,
`INDEX.md`. Copy the template below and keep only the pointer comment on the `format:` line: the docs stay here.

**What a note holds.** One fact, then why it holds, then how to apply it. Keep it to what a later session needs to act
on. **Never a secret:** no key, token, password or login, nor where a secret's value can be read; notes are committed.

**The budgets** (`bonsai check` finds a note or an index over them): a note at most 4 KB; the index at most 120 lines
and 12 KB, since every session loads it whole. When the index nears its budget, merge notes that say the same thing
and drop the ones no longer true. These caps' one home is Bonsai's table of memory caps, on its reference page of
lists.

**The personal layer: notes about the person, written by the person.** Beside the project's memory, each machine keeps
a personal layer in Bonsai's home: `personal/INDEX.md` and `personal/notes/`, in the same format, imported into every
session on the machine from the person's own Claude Code memory file, never in a project and never committed.
`bonsai init --yes` and `bonsai update --yes` write its empty index when there is none. Agents never write it: Bonsai's
guard and its deny rules refuse an agent's file-tool write anywhere in the home, since a note there loads into every
session in every project. When you learn a fact about the person (how they like replies, a tool they use everywhere),
draft the note with this template and hand it to them in your report, with the index line to add; they save it by
hand. A fact about how the person wants work done in this project goes in this project's memory, which you write.
Its caps: the personal index at most 40 lines and 4 KB, a note at most 4 KB. `bonsai check` warns, never fails, on the
layer: a file not in the format, one over its cap, an index line naming no note or a note with no line, a
secret-shaped string, and on Windows a copy of WSL's layer that is not known to match it.

## Fields

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `format` | The format and its version, always the first line; its comment points back to these docs | `bonsai.memory/1` | `bonsai.memory/1` |
| `id` | The note's id, unique in the folder; the file is named after it. The index has none | `M-` and a slug of lower-case letters, digits and dashes, the pattern `^M-[a-z0-9][a-z0-9-]*$`; `null` for the index | `M-fewer-verifications` |
| `title` | What the note is about, in one line; the index line repeats it | text, one line | `Fewer verifications` |
| `kind` | What kind of fact: about the project; how the person wants the work done; where something is found. `index` is the index's alone | `project`, `feedback`, `reference`, `index` | `feedback` |
| `updated` | The day the note last changed: the UTC day, from `bonsai now --day` | a date, `YYYY-MM-DD` | `2026-10-15` |
| `source` | Who said it, and when, so a later session can weigh it | text; `null` for the index | `"the person, 15 Oct"` |
| `labels` | Labels a linked pack defines for memory notes, by name | a mapping of label names to values; `{}` for none | `{}` |

```markdown
---
format: bonsai.memory/1   # fields: bonsai check --schema bonsai.memory; how to fill: skill base:memory
id: M-<slug>
title: <what the note is about, in one line>
kind: <project, feedback or reference>
updated: <YYYY-MM-DD>
source: "<who said it, and when>"
labels: {}
---

<The fact, in one or two sentences.>

**Why:** <why it holds: the reason given, or what happened.>

**How to apply:** <what a session does differently because of it.>
```

## The index

`INDEX.md`, in the same folder, carries the same frontmatter with `id: null`, `kind: index` and `source: null`, then
one line per note: a link to the note's file, its title, and a few words on when it matters. Add the line when you
write a note, change it when the note changes, and take it out when the note goes.

```markdown
---
format: bonsai.memory/1   # fields: bonsai check --schema bonsai.memory; how to fill: skill base:memory
id: null
title: Memory index
kind: index
updated: <YYYY-MM-DD>
source: null
labels: {}
---

- [Fewer verifications](M-fewer-verifications.md): a fresh verifier only for big or risky work
```

**Before you finish:** the note's file is named after its id, its line is in `INDEX.md`, and `bonsai check --json`
reports no finding for either.
