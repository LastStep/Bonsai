---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5737
title: "The guard's role-writes-nothing refusal says when it could not follow a cd, with a next step that fits"
status: done
lane:
done_when:
  - "a role-writes-nothing refusal of a relative write after a cd the guard cannot follow says that it could not follow the cd and names a next step that fits (name the path in full, outside the project), by test on the trial's two shapes (cd $S then > b.py; cd $S/probe then > probe.mjs); what is refused is unchanged"
  - "the climb at the branch's HEAD green, CI on the branch green, a fresh verifier (opus) passes it"
depends_on: []
blocked_by:
created: 2026-10-24
started: 2026-10-24
finished: 2026-10-24
labels:
  bonsai.branch: T-5737
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# The guard's words after a cd it cannot follow (T-5737)

T-5734 item 3: BUG-108's words (`records/bugs.md`). In LinkSink a researcher's `cd $S && cat > b.py` and the
verifier's `cd $S/probe && ... > probe.mjs` (both in Claude Code's scratchpad under /tmp) were refused as writes
inside the project (log lines 344 and 810), and the refusal's next step, "write scratch files outside the project",
did not fit what they did. Only the words change: the guard still refuses what it cannot place.
