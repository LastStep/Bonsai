---
name: example
description: "How to keep a note in this project: when to write one, where it goes, its id, and how to finish it. Use it before writing or finishing a note."
---

<!-- skills/example/SKILL.md: example-pack's skill that is no template, loaded as example-pack:example and preloaded
by the role example. Its frontmatter fields are documented once, in README.md ("Skill files"). -->

# Keeping a note

A note records one thing worth coming back to: a decision and why, a finding, or an open question.

1. **Where it goes.** In the project's notes folder: the place `bonsai.yaml` gives the kind `note` under its
   `documents`, `notes` unless the project changed it. One file per note, at the folder's top, named after its id
   (`N-0001.md`).
2. **Its id.** `N-` and four digits, the next one free in the folder. The kind's id pattern is in this pack's
   `bonsai/pack.yaml`; `bonsai status --json` lists it under `documents`.
3. **Its fields.** Open the template skill `example-pack:note` and copy its template; its fields table says what
   each field means and what it may hold. Leave the docs in the skill: a note keeps only the pointer comment on its
   first line.
4. **Finishing it.** When nothing is left to add, set `status: done` and `finished:` to today's date. A note that
   belongs to a task names it in `task:`.
5. **Check it** against the template's fields table before you answer: every field there, each value one it allows.
