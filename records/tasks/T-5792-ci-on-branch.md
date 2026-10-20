---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5792
title: "CI on the branch: a batch's branch pushed and green on CI (its windows jobs among them) before main moves"
status: done
lane:
done_when:
  - "ci.yml runs its jobs on a push to any branch, not only main (tags excluded), so a batch's branch gets every job, the two windows jobs among them; codeql.yml unchanged; every pin and persist-credentials line kept (internal/selfcheck's workflow tests green)"
  - "CLAUDE.md's landing rule reworded: the orchestrator pushes the batch's top branch, waits for CI green on that exact commit (every job), then fast-forwards main to it and pushes main, and deletes the pushed branch; check 10's local Windows half kept only for what CI cannot run; the GitHub section's lines (builders never push; the orchestrator pushes) say the same"
  - "actionlint and the pin test green; a green climb at the branch's HEAD; CI green on the pushed branch before landing (this task's own landing is the first under the new rule)"
depends_on: []
blocked_by:
created: 2026-10-20
started: 2026-10-20
finished: 2026-10-20
labels:
  bonsai.ladder: []
  bonsai.allows: [".github/workflows/ci.yml", "CLAUDE.md"]
  bonsai.branch: T-5792
---

# CI on the branch (T-5792)

The director, 19 Oct, `workflow:correct`'s first question (Windows before a landing): **"C: CI on branch"**. The
batch's branch is pushed, CI (its windows jobs among them) runs on it, and `main` is fast-forwarded only when it is
green; the local Windows half stays as the fallback for what CI cannot run. A stricter change (CI on more pushes) and
the director's word for the landing rule's text. Before 5.7's first landing.
