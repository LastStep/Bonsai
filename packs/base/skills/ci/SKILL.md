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
2. Set the three pins to the Bonsai the project uses (`bonsai --version` names it): `BONSAI_VERSION`, then
   `BONSAI_FILE` and `BONSAI_SHA256`, which come from that release's page, the file name and its line in the
   release's `checksums.txt`; the address in the step is the release's fixed one, `https://github.com/LastStep/Bonsai/releases/download/v<version>/<file>`. To move the
   pin later, change the version, the file name and the SHA-256 together.
3. Commit and push; the run's log shows `bonsai check`'s JSON, and the job fails on any finding (warnings never fail
   it).

The workflow downloads the release archive for Linux and checks it against its SHA-256 before anything in it runs; it
builds nothing and needs no Go. A project whose `bonsai` is older than 1.0 has no release to download: it builds
`bonsai` from a pinned commit of Bonsai's public repository instead (check out `LastStep/Bonsai` at the commit with
`persist-credentials: false`, set up the Go its `go.mod` names, and run `CGO_ENABLED=0 go build -o ../bonsai
./cmd/bonsai` in it, in place of "Get bonsai").

```yaml
# .github/workflows/bonsai.yml: Bonsai's check on every push and pull request. It downloads a pinned Bonsai release
# archive, checks it against the SHA-256 written beside it, and runs bonsai check --json in the project: any finding
# fails the job, warnings never do. It uses no secret and fetches no pack (the lock carries what the check reads).
# Every action is pinned by its full commit, its version in the comment beside it; move a pin only on purpose. From
# the skill base:ci, which documents it; every key carries its comment.
name: bonsai                                 # the workflow's name, as GitHub lists it

on:                                          # when it runs
  push:                                      # every push, on every branch
  pull_request:                              # every pull request, a fork's included

permissions:                                 # what the job's token may do
  contents: read                             # read the repository, nothing more

env:                                         # the pins the job reads
  # BONSAI_VERSION, BONSAI_FILE and BONSAI_SHA256: the Bonsai release the project uses, its Linux archive's name and
  # that file's SHA-256 line in the release's checksums.txt (64 lower-case hex characters). Move all three together.
  BONSAI_VERSION: "1.0.0"
  BONSAI_FILE: "bonsai_1.0.0_linux_amd64.tar.gz"
  BONSAI_SHA256: "5277cad9e36d4b1c520ea3b9be7a698965fa4757a0d4aa9a803b71c2469285e1"

jobs:                                        # one job
  check:                                     # bonsai check, on Linux
    runs-on: ubuntu-latest                   # Linux
    steps:                                   # in order
      - name: Check out the project          # the commit being checked, into project/
        uses: actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803  # v6.1.0
        with:                                # the step's inputs
          path: project                      # into project/
          persist-credentials: false         # no token left in the checkout's git settings

      - name: Get bonsai                     # the release archive, checked by its SHA-256 before anything in it runs
        run: |                               # download, check, unpack bonsai beside the project's checkout
          curl -fsSL -o "$BONSAI_FILE" "https://github.com/LastStep/Bonsai/releases/download/v${BONSAI_VERSION}/${BONSAI_FILE}"
          echo "$BONSAI_SHA256  $BONSAI_FILE" | sha256sum -c -
          tar -xzf "$BONSAI_FILE" bonsai

      - name: Run bonsai check               # the project's check: exit 0 with no finding, 1 with any
        working-directory: project           # in the project's checkout
        env:                                 # the step's environment
          BONSAI_HOME: ${{ runner.temp }}/bonsai-home   # a Bonsai home of its own, empty: CI keeps no machine record
        run: ../bonsai check --json          # the findings and warnings, as JSON in the log
```

**Before you finish:** the workflow's run is green on the commit you pushed, and `bonsai check --json` in its log
lists no finding.
