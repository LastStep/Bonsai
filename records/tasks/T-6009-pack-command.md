---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6009
title: "A pack command: add, remove and move a project's packs from the command line, and see which are behind"
status: approved
lane: full
done_when:
  - "A plan records/plans/P-T-6009.md, from an auditor's read of the code and a researcher's read of how other tools show and apply updates, answers the director's design questions (asked one at a time), is reviewed by a fresh Opus agent and approved by the director"
  - "The build follows the approved plan in two stages (the core; the director's eight picks), each its own task, branch and landing by the landing rule, each passed by a fresh Opus verifier"
depends_on: [T-6008]
blocked_by: null
created: 2026-10-28
started: 2026-10-28
finished: null
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# A pack command: add, remove and move a project's packs, and see which are behind

The director's words, 28 Oct: "to smoothen out the update experience for the packs, i want to implement a new
feature in bonsai", from the studio's asks (the artifact "Asks for Bonsai", item 1, the director's decision of 29 Oct
that Bonsai gets a pack command); then "Pack feature planning".

## Context

Today adding, removing or moving a pack is a hand edit of `bonsai.yaml` (person-only), then `bonsai update` (preview,
then `--yes`). No command edits a pack's `ref`. Seeing that a pack is behind takes `bonsai status --full` (network).
Moving the studio and ZenGarden-AI-API to the packs at 1.0.1 (T-6008) took a hand edit of two lines in each.

## Scope

- `bonsai pack add`, `remove` and a move to another ref or tag (the studio's `pack set <pack> --ref <tag>`), with
  Bonsai's preview and `--yes`, and the person/agent rule the plan sets.
- A cached "newer tag" per pack in the default `status --json`, and each pack's ref there (the studio's items 1, 8).

## Out of scope

- The studio's other asks: items 2, 4, 5 are BUG-137 to 139 (the bug sweep); 3, 6, 7 are answers in the studio's
  reply.
- Applying anything for a studio: Bonsai never applies an action on its own.
