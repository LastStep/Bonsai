---
name: ci
description: "A GitHub Actions workflow that runs bonsai check on every push and pull request, with no pack fetched: copy it to .github/workflows/bonsai.yml. Use it to give a linked project its CI, or to check or update the one it has."
---

<!-- skills/ci/SKILL.md: base's CI workflow template, loaded as base:ci; no fields table: the workflow documents every
key in its own comments. Its frontmatter fields are documented in base's README.md ("Skill files"). -->

# Bonsai's check in CI

A workflow that runs `bonsai check --json` in the project on every push and pull request, so a finding (a document
that does not read, a label of the wrong kind, a table out of date, a lock that does not match `bonsai.yaml`) fails
the pull request that brings it. `bonsai check` needs no pack in CI: the lock carries what it reads from each pack,
so it fetches nothing, needs no network beyond getting `bonsai` itself, and uses no secret.

**When to use this skill.** When a linked project has no Bonsai CI yet, or to move its pinned Bonsai. It is a
skill, not a file Bonsai writes at the link: a workflow runs on GitHub with the repository's token, so it is added by
an agent or a person on purpose, and its pin is the project's to move.

**How to add it.**
1. Copy the workflow below to `.github/workflows/bonsai.yml` in the project (in a project that protects
   `.github/**`, under a task granting that path).
2. Set `BONSAI_COMMIT` to the Bonsai the project uses: the commit of the `bonsai` its people and agents run
   (`bonsai --version` names it), all 40 characters.
3. Commit and push; the run's log shows `bonsai check`'s JSON, and the job fails on any finding (warnings never fail
   it).

**Before Bonsai 1.0** there is no release to download, so the workflow builds `bonsai` from the pinned commit of
Bonsai's public repository with the Go its `go.mod` names. **From Bonsai 1.0**, replace the three steps marked
"until Bonsai 1.0" with the one step after the workflow, which downloads the release archive and checks it by its
SHA-256 before running it.

```yaml
# .github/workflows/bonsai.yml: Bonsai's check on every push and pull request. It builds bonsai from a pinned
# commit of Bonsai's public repository and runs bonsai check --json in the project: any finding fails the job,
# warnings never do. It uses no secret and fetches no pack (the lock carries what the check reads). Every action is
# pinned by its full commit, its version in the comment beside it; move a pin only on purpose. From the skill
# base:ci, which documents it; every key carries its comment.
name: bonsai                                 # the workflow's name, as GitHub lists it

on:                                          # when it runs
  push:                                      # every push, on every branch
  pull_request:                              # every pull request, a fork's included

permissions:                                 # what the job's token may do
  contents: read                             # read the repository, nothing more

env:                                         # the pin the job reads
  # BONSAI_COMMIT: the Bonsai commit whose bonsai checks the project, all 40 characters: the project's own Bonsai
  BONSAI_COMMIT: "e1437d09832c86c7e6ae430b9336358e1706199a"

jobs:                                        # one job
  check:                                     # bonsai check, on Linux
    runs-on: ubuntu-latest                   # Linux
    steps:                                   # in order
      - name: Check out the project          # the commit being checked, into project/
        uses: actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803  # v6.1.0
        with:                                # the step's inputs
          path: project                      # into project/
          persist-credentials: false         # no token left in the checkout's git settings

      - name: Check out Bonsai               # until Bonsai 1.0: Bonsai's source at BONSAI_COMMIT, into bonsai-src/
        uses: actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803  # v6.1.0
        with:                                # the step's inputs
          repository: LastStep/Bonsai        # Bonsai's public repository
          ref: ${{ env.BONSAI_COMMIT }}      # the pinned commit
          path: bonsai-src                   # into bonsai-src/
          persist-credentials: false         # no token left in the checkout's git settings

      - name: Set up Go                      # until Bonsai 1.0: the Go that Bonsai's go.mod names
        uses: actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16  # v6.5.0
        with:                                # the step's inputs
          go-version-file: bonsai-src/go.mod # Bonsai's go.mod, its toolchain line included
          cache: false                       # nothing to cache: Bonsai uses Go's standard library only

      - name: Build bonsai                   # until Bonsai 1.0: a static build beside the two checkouts
        working-directory: bonsai-src        # in Bonsai's source
        run: CGO_ENABLED=0 go build -o ../bonsai ./cmd/bonsai   # cgo off, as Bonsai's release builds it

      - name: Run bonsai check               # the project's check: exit 0 with no finding, 1 with any
        working-directory: project           # in the project's checkout
        env:                                 # the step's environment
          BONSAI_HOME: ${{ runner.temp }}/bonsai-home   # a Bonsai home of its own, empty: CI keeps no machine record
        run: ../bonsai check --json          # the findings and warnings, as JSON in the log
```

**From Bonsai 1.0**, in place of the steps "Check out Bonsai", "Set up Go" and "Build bonsai", with the address of the
release's Linux archive and its SHA-256 from the release's `checksums.txt`:

```yaml
      - name: Get bonsai                     # Bonsai's release archive, checked by its SHA-256 before it runs
        env:                                 # the step's environment
          BONSAI_URL: "<the address of the release's linux amd64 archive>"   # from the release's page
          BONSAI_SHA256: "<its SHA-256, from the release's checksums.txt>"    # 64 lower-case hex characters
        run: |                               # download, check, unpack bonsai beside the checkout
          curl -fsSL -o bonsai.tar.gz "$BONSAI_URL"
          echo "$BONSAI_SHA256  bonsai.tar.gz" | sha256sum -c -
          tar -xzf bonsai.tar.gz bonsai
```

**Before you finish:** the workflow's run is green on the commit you pushed, and `bonsai check --json` in its log
lists no finding.
