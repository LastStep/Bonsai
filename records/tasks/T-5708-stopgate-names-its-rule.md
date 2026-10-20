---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5708
title: "The trial's feature: the stop gate names its rule when it blocks (BUG-089, with BUG-088)"
status: done
lane:
done_when:
  - "every stop-gate block prints its rule word in the shared hook frame, as the guard's refusals do (internal/stopgate/stopgate.go:229-231; internal/hookin/frame.go:99-100); a test wanting the rule line on a block, failing first on the code before it"
  - "the gate's own osWords: the two shared kinds from workspace.OSWords, its extra kinds kept (the director's answer B to ask agent:h-639ab913f88a)"
  - "stop-task-missing and stop-rung-missing renamed everywhere but past records, so every block can print its rule word (the director's answer B to ask agent:h-ae479b969f78); the reference page regenerated"
  - "packs/base/skills/operating-bonsai/SKILL.md:103's stale line on where the rule shows corrected; its test green"
  - "built by a named builder session under rc.1's guard, climbed by rc.1's ladder, the probes of plan-5.md note 5.7.7, 2 recorded; a fresh Opus verifier; the landing rule; CI on the branch"
depends_on: [T-5707]
blocked_by:
created: 2026-10-20
started: 2026-10-20
finished: 2026-10-20
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: T-5708
---

# The trial's feature: the stop gate names its rule (T-5708)

The director's pick on 20 Oct ("B: stop gate names its rule") from the proposals `M-2026-10-20-392e067a`. Plan:
`design/plan-5.md`, notes 5.7.7 and 5.7.8. Owns: `internal/stopgate/stopgate.go` and its tests;
`packs/base/skills/operating-bonsai/SKILL.md` and `cmd/bonsai/operating_test.go`; on the director's answer B to the second
ask, amended by the orchestrator: `internal/guard/rules.go`, `internal/guard/stop_test.go` and `docs/reference/lists.md`
(regenerated). No protected file: the grant stage
and the refusal outside it are recorded probes.
