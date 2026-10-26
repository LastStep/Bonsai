---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6001
title: "Bonsai's first existing project: link ZenGarden-AI-API (the pilot), retire its own agent system, and record the trial"
status: done
lane: null
done_when:
  - "The plan records/plans/P-T-6001.md was reviewed by a fresh Opus agent and approved by the director"
  - "ZenGarden-AI-API is linked to Bonsai 1.0 with the packs base (tag base-v1.0.0) and workflow (tag v1.0.0): its bonsai.yaml, lock and settings name both, and `bonsai check` and `bonsai status` there report no finding and nothing changed or missing"
  - "Its ladder has a pytest rung whose count is held by a ratchet at the count a fresh venv from requirements-test.txt gives (305 with the project's own venv on 26 Oct), and a climb on its main is green with git.dirty false"
  - "Its old agent system is sorted as the plan's table says: what Bonsai replaces and the history moved unchanged to archive/pre-bonsai/, the still-useful guides in docs/guides/ brought up to date and listed in CLAUDE.md, the lasting facts as memory notes; its CLAUDE.md keeps only the project's facts (hardware limits, architecture, how to run it) and its .bonsai/STATE.md says where the project stands"
  - "At least one real task of its own went through the workflow pack end to end there (task, run report, builder in a worktree, a green climb, landed)"
  - "records/trial-zengarden-ai-api.md in Bonsai's repo gives each of the plan's criteria a verdict with evidence, and every bug it found is in records/bugs.md"
depends_on: []
blocked_by: null
created: 2026-10-26
started: 2026-10-26
finished: 2026-10-26
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  workflow.model: opus
---

# Bonsai's first existing project: the ZenGarden-AI-API pilot

## Context

The director asked (26 Oct) how to bring ZEN and ZenGarden-AI-API onto Bonsai. Three read-only auditors probed them;
the findings and the director's two answers ("A, AI-API first") are in this task's run report
`records/runs/R-2026-10-26-T-6001-zengarden-pilot.md`. Bonsai has been linked only to itself and to a fresh project
(LinkSink, `records/trial-rc2.md`); this is its first project with history and an agent system of its own.

## Scope

The plan `records/plans/P-T-6001.md`; the link script in the scratch folder; the kickoff for the project's own
orchestrator; the trial record and bugs in Bonsai's repo. The changes inside ZenGarden-AI-API are made by the director's
own commands and by that project's own sessions, under its own tasks, never from Bonsai's session.

## Out of scope

ZEN (its own task, after this trial's record). The project's own debts the auditors found (ruff and mypy errors, the
venv's drift from its requirements, CI pushing an image when tests fail): its orchestrator may file them as its own
tasks. Any change to Bonsai's code: a bug found here goes to `records/bugs.md` and waits for the director's word.

## Notes

Numbering: the first task after 1.0's step 5 takes T-6001 (step 6, the work after 1.0).
