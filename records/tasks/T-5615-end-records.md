---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5615
title: "5.6 end checks 14 and 17, the orchestrator's records: the Windows-only tally, the sessions table against the run reports, correct's first run"
status: done
lane:
done_when:
  - "5.6's Windows-only tally (plan-5.md, rule 3) summed from 5.6's run reports' rows, each row it counts named, and step 5's running total carried into STATE"
  - "every 5.6 task with run rows matched to the sessions table's rows, each difference (a missing task label, differing minutes) explained by its cause in the log, or recorded as a fault in records/bugs.md"
  - "workflow:correct's first run at 5.6's end: its sources read (verifier findings, fix rounds, reverts, guard refusals), each mistake seen twice turned into a check or recorded why not, in its run report"
depends_on: [T-5611]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.6's end records (T-5615)

The end verifier T-5611 (`M-2026-10-19-c7306085`) failed 5.6 on check 14 (a) and (c) and on check 17; these are the
orchestrator's to close before the verifier's short re-check. Counting and listing by a Haiku auditor; the judgments
the orchestrator's.

- **Owns:** `records/runs/`, `records/bugs.md`, `.bonsai/STATE.md` (records only; the orchestrator's commit).
- **Grants:** none.
