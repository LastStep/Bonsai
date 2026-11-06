---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6017
title: "A ratchet's count under ladder --ci adds back the ci_marked_tests that skipped when the rung counts by capture, as it does when it reads the tests (BUG-138); for 1.0.3"
status: done
lane: null
done_when:
  - "internal/ladder's count (rungs.go:411-434) gives the same number locally and under --ci for a rung counting by capture whose marked tests skip in CI; a test shows it red on main first (floor 3: local 3, --ci 2) and green after; no count rises past what ran (a marked test that did not skip is not counted twice)"
  - "The floor's rungs 0-5 green at the branch HEAD by the installed bonsai (bonsai ladder --task T-6017); passed by T-6016's fresh verifier (the ladder is a gate); landed in 1.0.3's batch with CI green"
depends_on: [T-6015]
blocked_by: null
created: 2026-11-05
started: 2026-11-05
finished: 2026-11-06
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-6017-ci-ratchet-count
  workflow.model: opus
---

# The CI ratchet count (BUG-138)

The director, 5 Nov: "yes, build 1.0.3", on the studio's round 2 ask B (T-6015). The runner (`M-2026-11-05-e09bb306`):
with no capture, accepted marked skips are added back (rungs.go:430-432); with a `capture` they are not (:421-428).
Runs beside T-6016 and T-6018 (none grants a protected path).
