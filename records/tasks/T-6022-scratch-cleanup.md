---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6022
title: "An audit of the scratch folders (~/bonsai-checks and %USERPROFILE%\\bonsai-checks): what each holds, what is still used, what can go; then the clean-up the director approves"
status: done
lane: null
done_when:
  - "A table for the director: every top-level entry of both scratch folders with its size, last change, what it is (the task or step that made it), whether anything still uses it (STATE, a run report's way back, a launcher, a script, a kept repo), and a verdict: keep, remove, or ask"
  - "Only what the director approves is removed; what stays is named in STATE's housekeeping line"
depends_on: []
blocked_by: null
created: 2026-11-06
started: 2026-11-06
finished: 2026-11-06
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: sonnet
---

# The scratch folders' audit and clean-up

The director, 6 Nov: "do an audit of bonsai checks folder and clean up all the old and unnessary stuff. first show
me a table with overview of how things are there", "use a subagent to do it". WSL's folder: 8.4 GB, 100 folders and
26 loose files (6 Nov).
