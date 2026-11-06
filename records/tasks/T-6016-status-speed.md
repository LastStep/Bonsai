---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6016
title: "bonsai status --json fast in a repo with a large log (BUG-137): the schema's patterns compiled once, the sessions table check skipped while the log is unchanged; for 1.0.3"
status: done
lane: null
done_when:
  - "In Bonsai's repo (its 58 MB log), status --json's median over five runs of 1.0.3's build is measured and reported against 1.0.2's 7.20 s wall and 12.29 s CPU, and a run with nothing changed since the last is cheaper still; the same output, byte for byte, as 1.0.2's build on the same repo"
  - "Every finding bonsai check gives today it still gives (the sessions table's included): a test that a log or sessions-table change since the last run is still found; tests for the compiled-once patterns"
  - "The floor's rungs 0-5 green at the branch HEAD by the installed bonsai (bonsai ladder --task T-6016); a fresh Opus verifier passes it (tests and correctness); landed in 1.0.3's batch with CI green"
depends_on: [T-6015]
blocked_by: null
created: 2026-11-05
started: 2026-11-05
finished: 2026-11-06
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-6016-status-speed
  workflow.model: opus
---

# status --json fast (BUG-137)

The director, 5 Nov: "yes, build 1.0.3", on the studio's round 2 ask A (T-6015). The runner's profile
(`M-2026-11-05-e09bb306`): about 60 % of wall in the sessions table check reading all of `.bonsai/local/log` (63 files,
58 MB); `schema.validate` compiles a regexp per line (internal/schema/check.go:187), about 45 % of CPU; status runs the
full check. Runs beside T-6017 and T-6018 (none grants a protected path).
