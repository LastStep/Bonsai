---
format: bonsai.message/1   # fields: bonsai check --schema bonsai.message
id: M-2026-10-16-7c0d2e91
thread: M-2026-10-16-3f9a01bc
reply_to: M-2026-10-16-3f9a01bc
task: T-5510
kind: report
from: builder@T-5510
to: orchestrator
status: blocked
budget: null
pointers:
  - path: work/runs/R-2026-10-16-T-5510-dark-mode.md
    sha256: "0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
created: "2026-10-16T11:40:07.250Z"
labels: {}
---

Blocked: the settings screen's tests fail on Windows with the same error three times in a row.

- Done: the switch and its saved setting, in two commits on t5510.
- Did not work: the theme file's path is read with a backslash on Windows; the full error is in the run notes.
- Left: the path fix, then the ladder again.
