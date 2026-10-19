---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5699
title: "5.6's end Windows run, the agent half (runbook steps 1-3, 6, 8, 9)"
status: done
lane:
done_when:
  - "the runbook's steps 1, 2, 3, 6, 8 and 9 (~/bonsai-checks/end-5.6/runbook.md) run natively on Windows at main's top, one Windows run at a time, each step's commands, exit codes and verdict against the runbook's expectation in its report"
  - "a failure reported with its test name and first lines, not fixed in this task; nothing in the repository changed"
  - "steps 4, 5, 7 and 10 left for the director, untouched"
depends_on: [T-5696, T-5698]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.6's end Windows run, the agent half (T-5699)

The director, 18 Oct: "club windows checks later in the end"; the runbook `~/bonsai-checks/end-5.6/runbook.md`. Run by a
builder (the steps make scratch repositories and stamped builds, which a runner may not), as the one Windows run going.

- **Owns:** nothing in the repository; scratch folders under `~/bonsai-checks/end-5.6/` and `%USERPROFILE%\bonsai-checks\end-5.6\`.
- **Grants:** none.
