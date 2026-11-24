---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6030
title: "BUG-158: base:bug's \"Moving an old register\" steps made whole from their first real use; CLAUDE.md's pointer to the bugs moved to records/bugs/"
status: running
lane: light
done_when:
  - "Each of BUG-158's eight gaps answered in packs/base/skills/bug/SKILL.md (or, where the right answer is the command's, in the command with a test), as the T-6028 run report's \"What did not work\" names them"
  - "CLAUDE.md's \"Where the truth lives\" row for bugs names records/bugs/, bonsai bug and base:bug (the director, 24 Nov: \"yes, update CLAUDE.md\")"
  - "The ladder green at the branch's HEAD by the installed bonsai, CI green on the branch, both windows jobs; BUG-158 moved to fixed by bonsai bug fix after the landing"
depends_on: []
blocked_by: null
created: 2026-11-24
started: 2026-11-24
finished: null
labels:
  bonsai.ladder: []
  bonsai.allows: ["CLAUDE.md"]
  bonsai.branch: t-6030-bug-skill-gaps
  workflow.model: sonnet
---

# base:bug's migration steps made whole; CLAUDE.md's bug pointer

The director, 24 Nov: "yes, update CLAUDE.md and fix BUG-158". BUG-158 is the eight gaps the first real move of a
register (T-6028, Bonsai's own) found in base:bug's steps; its run report `records/runs/R-2026-11-24-T-6028-bugs-moved.md`
names each with its step. Light lane: text first; a gap whose answer is the command's comes back as a question if it
is more than a small change.
