---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-0901
title: Add a dark mode switch to the settings page
status: running
lane: light
done_when:
  - "The settings page shows a dark mode switch"
  - "The choice survives a restart"
  - "Rungs 0-2 green in one ladder run"
depends_on: [T-0900]
blocked_by:
created: 2026-10-08
started: 2026-10-08
finished:
labels:
  bonsai.branch: t0901-dark-mode
  bonsai.ladder: [0, 1, 2]
  bonsai.allows: ["assets/themes/**"]
  bonsai.wants: []
  workflow.owner: builder
  workflow.model: sonnet
  workflow.estimate_h: 2.5
---

# Add a dark mode switch to the settings page

A made-up task: the example for `bonsai.task/1`. Its frontmatter is what `task.json` holds, as a reader returns it.
