---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5733
title: "The director's own trial on rc.2 (LinkSink): an audit of whether it succeeded for Bonsai"
status: done
lane: light
done_when:
  - "Each success criterion below judged pass, partial or fail with file-and-line or log-record evidence, by two fresh auditors (opus) and the orchestrator's read"
  - "records/trial-rc2.md written for the director: the verdict, the evidence, and every gap filed as fixed in rc.3, after 1.0, or nothing in code"
depends_on: []
blocked_by:
created: 2026-10-24
started: 2026-10-24
finished: 2026-10-24
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# The trial on rc.2: did it succeed for Bonsai (T-5733)

The director, 24 Oct: "audit linksink in a way that we can know if the trial was successful for bonsai or not". The
trial (STATE, "the road to 1.0"): their own fresh project on rc.2, `~/Apps/LinkSink`, linked 24 Oct 06:37 UTC, run by
them in Claude Code. Read-only: nothing in LinkSink is changed.

Success criteria: (A) the README's first project path linked it with no help; (B) the guard refused what it should,
allowed nothing it should not, and every refusal named its next step; (C) the log is complete and well formed and holds
no secret; (D) messages valid and threaded; (E) every document Bonsai reads passes check and the tables rebuild; (F)
the packs steered the agents (roles, lanes, tasks moved, run reports, STATE); (G) what the director had to do by hand,
and what confused them or the agents.
