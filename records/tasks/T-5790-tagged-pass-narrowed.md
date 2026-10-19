---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5790
title: "The tagged pass narrowed to the packages the bonsai_test_fault tag can reach, held there by a test"
status: done
lane:
done_when:
  - "the ladder's rungs 1 and 4 and CI's tagged steps (test job, windows (fault tag), lint's --build-tags) vet, test and lint with -tags bonsai_test_fault only the packages that hold a file naming the tag or reach such a package in their normal or test build (today ./cmd/bonsai ./internal/guard ./internal/ladder ./internal/reference ./internal/reference/gen ./internal/rung0 ./internal/stopgate; audit M-2026-10-19-3fa8b9f8); the plain passes unchanged"
  - "a Go test that derives that set from go list (deps with tests, with and without the tag) and fails when any copy of the list in bonsai.yaml or ci.yml differs from it; shown failing first on a copy where a package off the list imports internal/guard, and on a copy with a stale list"
  - "bonsai.yaml's go_tests_tagged ratchet set to the narrowed climb's count (2120 at the audit), no other floor or rung field changed; marked in the run report as a loosening on the director's word"
  - "a green climb before landing (main's committed rungs); after landing and the director's bonsai update, a green climb of main under this task with the narrowed rung 4; CI green on the landed commit, every job"
depends_on: [T-5617]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".github/workflows/ci.yml"]
  bonsai.wants: ["bonsai.yaml"]
  bonsai.branch: T-5790
---

# The tagged pass narrowed (T-5790)

The director, 19 Oct: "yes look into that so we don't do same work twice", then, on the audit's facts (the tag
changes only `internal/guard`'s file set; 27 of 34 packages cannot reach it, and their 2,642 tagged tests repeat the
plain pass exactly), **"Yes, narrow it"**: every fault test kept, the repeats dropped, the floor `go_tests_tagged`
4762 to the narrowed count, their `bonsai update` after it lands. A loosening in CLAUDE.md's terms (a rung's command
and a floor), made on the director's word.

- **Owns:** `bonsai.yaml`'s rungs 1 and 4 commands and the `go_tests_tagged` ratchet; `.github/workflows/ci.yml`'s
  tagged steps and lint line; one new test (its package the builder's choice, named in the audit); the docs naming
  the tagged pass (CONTRIBUTING.md, the operating or ladder skill lines if they quote the command).
- **Grants:** `bonsai.yaml` (person-only: the director's word, 19 Oct), `.github/workflows/ci.yml`.
