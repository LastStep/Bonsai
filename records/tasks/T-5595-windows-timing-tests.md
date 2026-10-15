---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5595
title: "Windows timing tests that fail under load: their own budgets and limits, so they judge what they test, not how fast"
status: running
lane:
done_when:
  - "internal/clean's TestDefaultsWhenLeftOut (and any other test of SessionEnd that leaves the 1 s budget to chance) sets its own Budget, as TestEveryKind does since 5.5.8"
  - "internal/ladder's TestLockOfADeadHolder judges that a climb never waits for a dead holder by its words and the lock it takes, not by a 5 s wall clock"
  - "internal/stopgate's tests that met the gate's 4 s git limit under load (TestStopGateFixtures/env-other-project, TestTheRecordHoldsNoUserFolder) and the guard's TestAnAllowItCannotRecordIsBlocked run with a generous git limit of their own; the gate's own limit and behaviour unchanged"
  - "Each shown failing first under a held clock or a tiny limit in a temporary copy where it can be; 20 runs green natively on Windows; check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green"
depends_on: [T-5508]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished:
labels:
  bonsai.branch: "t5595"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# Windows timing tests that fail under load

Part 5.5's work outside its pieces: five Windows check-10 runs on 15 Oct (each with other agents' runs beside it) went
red on tests that judge a wall clock rather than what they test: `internal/clean`'s `TestDefaultsWhenLeftOut` (the
session-end cleaning's 1 s budget, as `TestEveryKind` was before 5.5.8), `internal/ladder`'s `TestLockOfADeadHolder`
(a 5 s limit), `internal/stopgate`'s `TestStopGateFixtures/env-other-project` and `TestTheRecordHoldsNoUserFolder`,
and `internal/guard`'s `TestAnAllowItCannotRecordIsBlocked` (the gate's 4 s git limit). Each passed alone. They redden
landings, and CI next; the gate's own 4 s limit stays a known limit (Rohan, 13 Oct).

- **Owns:** those tests and their test helpers; in `internal/stopgate/`, `gitWait` made a variable its tests may raise
  (its value and the gate's behaviour unchanged); nothing else.
- **Grants:** none: no protected path, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the orchestrator's read of the diff; CI.
