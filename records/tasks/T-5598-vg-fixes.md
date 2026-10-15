---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5598
title: "VG's fix round for 5.5.11 and 5.5.8's item 4: role rights only behind an allowed call, case-free types, the report found by the gate"
status: done
lane:
done_when:
  - "A: a declared role's rights given only to a child joined to an allowed Agent or Task call for it, else it is judged an undeclared type; a Skill call that forks into a declared role (context: fork, agent: <role>) judged as a call to that role; Go tests on the measured payloads"
  - "B: role and agent types matched case-insensitively, as Claude Code matches them (\"Builder\" is builder); C: the stop gate finds the report message itself, never a flag set at PreToolUse, so a failed report command leaves the session held"
  - "D: past a budget, only the installed bonsai's place or a bare bonsai counts as the report or a brief's read; E: a climb with no --task counted as BONSAI_TASK's red run; F: internal/agents' readers open regular files only, and mkfifo and mknod join the bonsai-files writes"
  - "H (5.5.8 item 4): the new-tests reruns on the base run with rungEnv(c.o.Env), so they hold neither BONSAI_TASK nor BONSAI_ROLE, with a test; I: an agent whose state does not read is moved aside and counted afresh, not left uncounted"
  - "Where they fit: G (recorded.go:118's threshold through message.AtLeastMinutes); J's literal forms in the tables; K's three; the recorder's settle wait written as a limit or checked against last_assistant_message; every limit VG judged wider than written rewritten to what holds"
  - "Each of VG's reproducers a Go test that fails first; check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VG's second round passes 5.5.11 and 5.5.8's items 4 and 5"
depends_on: [T-5596]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: "t5598"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# VG's fix round for 5.5.11 and 5.5.8's item 4

Part 5.5's work outside its pieces: VG's (`T-5596`) must-fixes A-F and H, its should-fix I and its can-wait items where
they fit, fixed forward by one fresh Opus fixer (5.5.11's builder past the ~500k hand-over line). VG's report is in
`records/runs/R-2026-10-15-T-5596-verify-vg.md`, its log at 09:42.

- **Owns:** `internal/guard/` (the role rules, the bonsai-files tables, the budgets, the known-limits comments);
  `internal/stopgate/`; `internal/agents/`; `internal/ladder/kinds.go`; `internal/message/thread.go`'s
  `AtLeastMinutes`; the recorder's settle wait in `internal/recorder/`; `docs/reference/lists.md` (regenerated); their
  tests.
- **Grants:** none: no protected path, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the landing rule; CI; VG's second round.
