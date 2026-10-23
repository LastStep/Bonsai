---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5799
title: "An audit of the public repo's files: what can go before 1.0, what stays, what to ask"
status: done
lane: null
done_when:
  - "every tracked file or folder at main is in one of remove, keep or ask, each with its reason and evidence (what still points to it)"
  - "the director has the list and has said which go; the removal lands as its own task"
depends_on: []
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# An audit of the public repo's files

## Context

The director, 23 Oct, beside rc.2's fixes: "clean up any old and unnessary public files which are not useful anymore.
do an audit and tell me which ones can be removed, which ones you suggest to not remove and so on". The repo
(`LastStep/Bonsai`) is public; 1.0 is its first release of the rebuild.

## Out of scope

Removing anything: that is a later task, on the director's word. Past run reports, logs and tasks are left as they
are (the director, 12 Oct).

## Notes

- 23 Oct: the auditor's report `M-2026-10-23-c740a8b9`; the director: remove `assets/demos/` and CHANGELOG's old
  sections; design/ "Keep them all". Done when T-5730 lands.
