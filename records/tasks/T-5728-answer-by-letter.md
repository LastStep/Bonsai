---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5728
title: "bonsai answer takes an option's letter or number"
status: done
lane: null
done_when:
  - "a Decide ask answered with a letter (A) or number (1) records that option's full text in choice, by test"
  - "a letter or number with no such option is refused with a next step naming the options"
  - "the ask format's fields are unchanged; help, the reference page and the docs say how"
  - "the climb at the branch's HEAD green, CI on the branch green"
depends_on: []
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.branch: T-5728
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# bonsai answer takes an option's letter or number

## Context

T-5710 item 11 (the director, 20 Oct; the trial's rough 4, `records/trial-rc1.md`): `--choice` needs the option's
whole text, so the director typed `--words "B"`.
