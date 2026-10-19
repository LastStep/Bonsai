---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5617
title: "CI's windows job faster: the plain and tagged passes as parallel jobs"
status: running
lane:
done_when:
  - "CI's windows test job split in two jobs running side by side, one the plain pass and one the tagged pass (-tags bonsai_test_fault), each running every package and every test the single job ran (no package, test, vet or flag dropped), the job names readable in the check-runs list"
  - "Defender's settings on the runner untouched (the director, 19 Oct: \"dont turn off defender scanning\")"
  - "the windows jobs' times before (7a05c12d's run) and after quoted from CI; CI green on the landed commit, every job; a green climb"
depends_on: [T-5613]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: [".github/workflows/ci.yml"]
  bonsai.wants: []
  bonsai.branch: T-5617
---

# CI's windows job faster (T-5617)

The director, 19 Oct: "Both 1 and 2", then "actually dont turn off defender scanning": the parallel jobs only (STATE's proposal: the windows job takes ~30 min, the plain and tagged passes in a
row; `cmd/bonsai` ~11 min and `internal/engine` ~9 min there). Running the tagged pass on fewer packages stays out (it
narrows a check). Starts once T-5613's CI run is read green, so a red run has one cause.

- **Owns:** `.github/workflows/ci.yml`'s windows job; `CONTRIBUTING.md`'s list of CI jobs.
- **Grants:** `.github/workflows/ci.yml` (protected, not person-only): the checks kept whole.
