---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5515
title: "5.5.5's fix round for the guard and the commands (5.5.5b): a finding's own next step allowed, read-only commands not judged as writes, a named session's minutes budget"
status: verify
lane:
done_when:
  - "S1's finding: the `.bonsai/local` tracked finding's next.do (who: agent) is a command the guard lets an agent run, or its who and next change to what an agent may do; a Go test that fails first"
  - "`bonsai message --help` and `bonsai update` without --yes (a preview that writes nothing), and any other --help or preview the guard judges as a write, let through; Go tests that fail first"
  - "A named builder session (BONSAI_TASK, BONSAI_ROLE builder) whose role's minutes budget is spent has its calls refused but the report's, as note 5.5.5's R8 says, and its log shows budget-spent; a Go test that fails first; or, if the design holds no minutes budget for a named session, the run report quotes where and the orchestrator is told"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; a fresh Opus verifier passes it"
depends_on: [T-5505]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.branch: "t5515"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.5.5's fix round for the guard and the commands (5.5.5b)

Piece 5.5.5b, the next free number: 5.5.5's real sessions (`T-5505`, report `~/bonsai-checks/T-5505/report.md`,
findings 1, 2 and 5) found three faults in code that 5.3, 5.5.11 and 5.5.10 own. Fixed forward by an Opus builder.

- **Owns:** `internal/guard/`, `internal/engine/` and `cmd/bonsai/` where the three findings lead, and their tests.
- **Grants:** none: no protected path.
- **Proof:** the lines of `done_when`.
