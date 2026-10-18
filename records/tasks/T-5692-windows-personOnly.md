---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5692
title: "Windows CI fixed forward: TestPersonOnly's two stamped-home cases, red since 5.6.0 (4b41790b)"
status: done
lane:
done_when:
  - "the cause of TestPersonOnly's two failing cases on CI's windows job (stamped, the home inside, an agent and a person) found and written in the run report, from the code and CI's log"
  - "fixed in the code or the test, whichever is wrong, with no case dropped or weakened (workflow:reward-hacking-guards); the home confinement still refuses a stamped build whose home is outside its scratch root and a plain build always"
  - "check 10 on WSL plain and tagged; the Windows half natively for the packages Defender allows; a climb green at the branch's HEAD; after landing, CI's windows job green"
depends_on: [T-5600]
blocked_by:
created: 2026-10-18
started: 2026-10-18
finished: 2026-10-18
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5692
---

# Windows CI fixed forward (T-5692)

CI's `windows` job has failed on `main` since 5.6.0 landed (`4b41790b`; again at `7daf9f35`):
`TestPersonOnly/stamped,_the_home_inside,_an_agent` and `.../a_person` in `cmd/bonsai`. Red CI is fixed forward
(`CLAUDE.md`). 5.6.0's `personOnly` (`cmd/bonsai/agent.go`) is the agent-session refusal with the stamped build's
home confinement (plan 5.6.0's row; 5.3.6 note 8).

- **Owns:** `cmd/bonsai/agent.go` and its test file(s) for `personOnly`; from 18 Oct (the audit `M-2026-10-18-7371b037`), also `internal/engine/hookline.go`'s `realFolder` and `Confine` and their tests, the same path-resolution bug; anything else is a `needs-input`.
- **Grants:** none (no protected path).
- **Proof:** the lines of `done_when`.
