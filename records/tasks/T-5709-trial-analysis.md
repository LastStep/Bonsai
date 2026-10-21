---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5709
title: "5.7.9, the trial's analysis: records/trial-rc1.md for the director, and an agent's dry run of the README's first-project steps"
status: verify
lane:
done_when:
  - "records/trial-rc1.md as plan-5.md note 5.7.9 says: every claim pointing at a file, a record or a command's output; a Haiku listing of files and records checked against it; the privacy grep"
  - "an agent's dry run of the README's first-project steps on scratch with the installed rc.1, its report"
  - "the director's reply recorded (batch 4)"
depends_on: [T-5708]
blocked_by:
created: 2026-10-20
started: 2026-10-21
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
---

# 5.7.9, the trial's analysis (T-5709)

The evidence: `records/runs/R-2026-10-20-T-5707-trial-on-rc1.md` (the orchestrator's log, every finding so far) and
`R-2026-10-20-T-5708-stopgate-names-its-rule.md` (the builder's); the builder's report `M-2026-10-20-504c5465`; the
verifier's `M-2026-10-20-57b77ed9`; the session's stream `~/bonsai-checks/T-5708/stream.jsonl`; the files before and
after `~/bonsai-checks/T-5708/files-{before,after}.txt`; the asks in `.bonsai/local/asks/2026-10-20.ndjson`; the log
`.bonsai/local/log/s-64096e53-....ndjson`; the climbs in `.bonsai/local/ladder/T-5708.json` and the log.
