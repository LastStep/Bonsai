---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5797
title: "Bonsai with any agent, pi first: the master research"
status: done
lane:
done_when:
  - "design/agent-neutral.md written: every Claude Code coupling in Bonsai by concern, what any agent host must give Bonsai, how pi and the other agents give it, the security model, the options with one recommendation each, and what (if anything) belongs before 1.0, every claim sourced"
  - "the unchecked points of the pi trial's review settled by probes in scratch (~/bonsai-checks/T-5797/), their results in the document"
  - "a fresh Opus verifier (facts and security) passed the document; the director has read it"
depends_on: []
blocked_by:
created: 2026-10-22
started: 2026-10-22
finished: 2026-10-23
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# Bonsai with any agent, pi first: the master research (T-5797)

The director, 22 Oct: "currently bonsai is dependent on being used with claude code a lot, and it fails when being used
with other agentic systems ... we need to solve it at the fundamental level, use creativity wherever we can, and
security is always a core principle". A research session, no code: its document is what later sessions plan from.

The director's answers the same day: any agent, pi first (other agents studied to test the design, not built for); the
1.0 question answered at the research's end, with reasons; pi probes in scratch under a scratch Bonsai home, no new
installs; the document in `design/`.

Evidence to start from: the pi trial in `~/bonsai-checks/pi-trial/` (`FINDINGS.md`, `PI-ADAPTER.md`, `PI-PIECES.md`,
each with a second reviewer's notes). Spec section 2 ("Later, on the director's word: other agents and AGENTS.md"),
section 5 ("Claude Code first (Q5 a): ... a thin adapter per agent later") and section 7 (decision D: guards are a
tripwire, not a wall against a same-user agent).

The task stays at `plan` and grants nothing: another session holds three tasks `running`, and a fourth would leave its
next one with no active task.

**Next (22 Oct):** the director reads the document; the next session walks them through it and asks the choices in
`design/agent-neutral-walkthrough.md`; the night's plan is written there (its own tasks), not here. This task goes to
`done` with the records commit once the director has read the document.
