---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5729
title: "Bonsai's words from the dry run and BUG-097"
status: done
lane: null
done_when:
  - "bonsai update that records looser settings leads with what it recorded, not nothing to change, by test"
  - "no printed line names a plan step (step 5.6 and the like) a new user cannot know"
  - "the ladder prints its result's path relative to the checkout"
  - "a printed climb no longer names other programs on the machine by name (BUG-097), by test"
  - "the climb at the branch's HEAD green, CI on the branch green"
depends_on: []
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.branch: T-5729
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# Bonsai's words from the dry run and BUG-097

## Context

T-5710 item 12 and the README dry run's Bonsai-words items 11 and 13 (`records/trial-rc1.md`, "The README's dry
run"); BUG-097 in `records/bugs.md`.
