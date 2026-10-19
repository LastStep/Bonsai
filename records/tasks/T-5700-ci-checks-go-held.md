---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5700
title: "5.7.0, CI's own checks, and Go held"
status: done
lane:
done_when:
  - "plan-5.md's note 5.7.0, 1-5: go.mod's toolchain Go 1.27's newest patch (a bump only if newer); each CI job prints go version first; govulncheck -show verbose clean on it (a finding named with its id); a second golangci-lint step with --build-tags=bonsai_test_fault; actionlint at an exact version over .github/workflows/ and packs/template/.github/workflows/"
  - "every uses: in ci.yml and codeql.yml pinned owner/repo@<40 hex> # vX.Y.Z (the commit read from the release tag with git ls-remote), every actions/checkout with persist-credentials: false"
  - "internal/selfcheck/workflows_test.go holds the pins, failing on a copy with actions/checkout@v6 and on a 40-hex pin with no comment; actionlint failing on a misspelt key"
  - "WSL check 10, lint, a green climb; CI green after the landing"
depends_on: []
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: [".github/workflows/ci.yml", ".github/workflows/codeql.yml", "go.mod"]
  bonsai.wants: []
  bonsai.branch: T-5700
---

# 5.7.0, CI's own checks, and Go held (T-5700)

Plan: `design/plan-5.md`, step 5.7, note "5.7.0, CI's own checks, and Go held" and its row. The director, 19 Oct: "5.7.0
lands, 5.7.4 drafted" (ahead of the plan's gate, by that word). BUG-007 in `records/bugs.md`.

- **Owns:** `.github/workflows/ci.yml` and `codeql.yml`; `go.mod`'s toolchain line only if a newer Go 1.27 patch is out;
  a new package `internal/selfcheck/` with `workflows_test.go`; `packs/template/.github/workflows/pack.yml`'s `uses:`
  pins only (the test reads that folder; not protected).
- **Grants:** `ci.yml`, `codeql.yml`, `go.mod` (protected, not person-only: a stricter change). `release.yml` is
  person-only and 5.7.2's: not touched.
