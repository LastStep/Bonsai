---
format: bonsai.memory/1   # fields: bonsai check --schema bonsai.memory; how to fill: skill workflow:memory
id: M-fewer-verifications
title: Fewer verifications
kind: feedback
updated: 2026-10-06
source: "the project's owner, 6 Oct"
labels: {}
---

A made-up memory note: the example for `bonsai.memory/1`. Its frontmatter is what `memory.json` holds, as a reader
returns it.

Run a fresh verifier only for big or risky work and at the end of each step.

**Why:** a verifier on every small change cost more time than the mistakes it caught.

**How to apply:** small changes close on green tests and the orchestrator's read of the diff, said in the run report.
