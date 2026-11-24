---
format: bonsai.memory/1   # fields: bonsai check --schema bonsai.memory; how to fill: skill base:memory
id: M-bug-register-history
title: Bonsai's bug register, its history and its words
kind: project
updated: 2026-11-24
source: "the old hand-kept records/bugs.md (19 Oct to 24 Nov 2026), moved to bug files by T-6028"
labels: {}
---

Bonsai's bugs were one hand-kept page, records/bugs.md, from 19 Oct 2026 (the audit `M-2026-10-18-39aecfea` and the
draft `M-2026-10-18-47cb853c`, auditor on Sonnet, read at `main` 91d8a972, replacing STATE's bug lists) until T-6028
moved it to one file per bug in records/bugs/ on 24 Nov; records/bugs.md is now the generated page. Re-checked 28 Oct
(T-6007) at `d4c70ba2`: an auditor (Sonnet) read every open row, a fresh verifier (Opus) read 30 blind, two were run
natively on Windows; 8 rows to Fixed, 10 to Known, not bugs (now wontfix), BUG-057 split (BUG-135). The evidence per
row: records/runs/R-2026-10-28-T-6007-bug-sweep.md.

**Why:** the old page's intro held this, and its conventions still read in the carried-over bugs' text.

**How to apply:** in a bug's `where`, **pointed** means the line was read there; **said** means only a report's or
STATE's words were read, not the code. A limit the director accepted for after 1.0 stays open with that word. A
design question is never a bug: it waits in STATE's "Waiting on the director".
