---
format: bonsai.bug/1   # fields: bonsai check --schema bonsai.bug; how to fill: skill base:bug
id: BUG-012
title: The settings page forgets the dark mode switch after a restart
status: fixed
area: settings
where: src/settings/store.ts:88 (pointed)
found_by: verifier@T-0901 M-2026-10-09-1a2b3c4d
filed_by: builder@T-0901
fixed_by: builder@T-0904
fixed_in: "0a1b2c3d (the switch is saved before the page closes)"
task: T-0904
was: TB-012
why: null
created: 2026-10-09
closed: 2026-10-10
labels: {}
---

A made-up bug: the example for `bonsai.bug/1`. Its frontmatter is what `bug.json` holds, as a reader returns it.

**How to repeat it:** turn dark mode on, close the game, open it again: the switch reads off.
