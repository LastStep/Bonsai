---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6007
title: "The bug sweep: triage the open bugs in records/bugs.md and plan their fixes in batches"
status: todo
lane: null
done_when:
  - "Every bug in records/bugs.md's Open table at the start of this task (76, BUG-006 to BUG-133) is sorted, with evidence read at main, into one of: already fixed (moved to the Fixed table with its commit), not a bug or no longer real (moved out with the reason), a fix for a batch, a decision for the director, or a later piece the director already named"
  - "A plan records/plans/P-T-6007.md groups the fixes into batches, each with its own task (T-6009 on; T-6008 is the packs at 1.0.1), its files, its risk and whether it needs a fresh Opus verifier; reviewed by a fresh Opus agent and approved by the director"
  - "The director's decisions on the bugs that need one are asked one at a time and recorded in the run report in their words"
depends_on: []
blocked_by: null
created: 2026-10-28
started: 2026-10-28
finished: null
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# The bug sweep: triage the open bugs and plan their fixes in batches

The director's word, 28 Oct: on what comes after 1.0.1, "Bug sweep" (over scratch folders and the cleaner, linking
the next project, and the pi work).

## Context

`records/bugs.md` holds 76 open bugs at `d4c70ba2` (BUG-006 to BUG-133). Many were filed on 19 Oct from STATE's old
lists and reports, several marked **said** (never read in the code); some may since have been fixed by 5.6, 5.7 or
1.0.1's work without the register knowing. The sweep first sorts every one by what the code at `main` shows, then
plans the fixes in a few batches, each its own task and landing, and likely a 1.0.2 release at the end.

## Scope

- The triage: an auditor reads each open bug against `main` and reports its sort, with file:line evidence.
- The register: bugs already fixed or no longer real move out of Open, each with its evidence.
- The plan: `records/plans/P-T-6007.md`, the batches and their tasks.
- The director's decisions: one question at a time, for bugs where the fix is a choice.

## Out of scope

- Building any fix: each batch is its own task.
- The pieces the director placed after 1.0 with their own words (scratch folders `T-5793`, the cleaner role, the pi
  work, playbooks): a bug that is really one of those is sorted "later piece" and left open.
- The workflow pack's release and tags (the director's).
