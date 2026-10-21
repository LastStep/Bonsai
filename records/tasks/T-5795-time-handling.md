---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5795
title: "Before 1.0: one system for time (stored with its zone, shown in local time, never typed by hand), planned with the director"
status: done
lane: full
done_when:
  - "an auditor's inventory of every time Bonsai and its packs write, print or read (field, format, zone, who writes it), and a researcher's survey of better ways, the director's 'time script' idea among them"
  - "options put to the director, then a plan (workflow:plan) reviewed by a fresh Opus agent and approved by the director, its pieces as tasks"
  - "landed before the last pre-release, so 1.0 carries it"
depends_on: []
blocked_by:
created: 2026-10-21
started: 2026-10-21
finished: 2026-10-21
labels:
  bonsai.ladder: []
  bonsai.allows: []
---

# One system for time, before 1.0 (T-5795)

From the trial's analysis (`records/trial-rc1.md`, rough 11, "three clocks to line up"). Bonsai's own records (the
log, asks, messages, ladder results) keep ISO-8601 UTC with the zone; `.bonsai/sessions.md` prints UTC with the zone
in its header only; run reports' `started`/`finished` and log lines are typed by agents as `YYYY-MM-DD HH:MM` in local
time with no zone (`design/contract.md:592`); message ids take the UTC day, run report ids the local day.

The director, 21 Oct, asked "what is going on there exactly? and do we need a better system of handling time
overall? how can we solve this on a higher level?"; offered (plan now and build after 1.0; the display in rc.2; all
before 1.0), chose **"option 3"** (all of it before 1.0, the format change included), and: **"also try to once see if
there is a better way to fix this for good than what you suggested. like maybe a simple time script to run for
example?"**

The orchestrator's first sketch (to be bettered, not assumed): store UTC with the zone everywhere; show the machine's
local zone, named, wherever a person reads; agents never type a time (run reports' rows filled from the log); any time
still typed carries its zone. Run report: `records/runs/R-2026-10-21-T-5795-time-handling.md`.

**Approved 21 Oct.** The plan `records/plans/P-T-5795.md` (revision 2; reviews `M-2026-10-21-979a667b` FAIL,
`M-2026-10-21-b5a38f91` PASS); the director chose the guard's check of fresh stamps ("Guard checks the stamp") and
approved it ("Approved"). Step 0 done: `design/contract.md` section 23 "Time". Pieces: `T-5712` (A, formats set 11),
`T-5713` (B, `bonsai now`), `T-5714` (C, printed output), `T-5715` (D, the sessions table), `T-5716` (F, the guard),
`T-5717` (E, now on prompts and subagents). This task closes when the six have landed and the verifier has passed.
