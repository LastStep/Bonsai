---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-20-T-5705
task: T-5705
role: orchestrator
model: opus
started: 2026-10-20 16:06
finished:
outcome:
commits: []
labels: {}
---

# Run: T-5705, 5.7.5, the releases checked

- Who: the orchestrator
- Why: `design/plan-5.md`, step 5.7, piece 5.7.5

> Opened before the first edit and appended to as the work goes.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Orchestrator, batch 1 read back | opus | 2026-10-20 16:06 | 2026-10-20 16:06 | 2 |
| Runner M-2026-10-20-8b4f4162 (rc.1, WSL) | sonnet | 2026-10-20 16:10 | 2026-10-20 16:15 | 5 |
| Orchestrator, rc.1's Windows side and the LF rebuild | opus | 2026-10-20 16:16 | 2026-10-20 16:22 | 6 |
| V2 M-2026-10-20-aea4c214 (two rounds) | opus | 2026-10-20 16:10 | 2026-10-20 16:27 | 17 |
| Orchestrator, the install read-backs | opus | 2026-10-20 16:38 | 2026-10-20 16:40 | 2 |

## Log

- 16:06 **Batch 1 done by the director** ("done. verify everything once. and the homebrew token was deleted").
  Read back by the orchestrator with GET requests only, changing nothing:
  - environment `release`: required_reviewers (one User reviewer, the repository's owner account; prevent_self_review
    false; no wait timer) and branch_policy; deployment_branch_policy custom_branch_policies true, protected_branches
    false; one policy, `v* tag`; no environment secret.
  - immutable releases: Bonsai `true`, bonsai-workflow `true`.
  - rulesets: Bonsai `main-protection` (branch, active, ~DEFAULT_BRANCH, deletion and non_fast_forward; one bypass
    actor, as before batch 1) and `pack-tags` (tag, active, refs/tags/base-v*, update and deletion, no bypass);
    bonsai-workflow `release-tags` (tag, active, refs/tags/v*, update and deletion, no bypass).
  - workflows: ci.yml, codeql.yml and release.yml `active`.
  - no repository secret on either repository; no v1 tag and no base-v tag yet; newest release v0.4.3.
  - gh 2.102.0, /usr/bin/gh first. The Homebrew token: deleted (the director's word).
  Every read-back as batch 1 says.
- 16:16 **Batch 2**: the director tagged v1.0.0-rc.1 on 73919ddc (ref read back: 73919ddcc546fa9848f13fcba8a2707169cd6ca3); release run 39410125869 started by the tag push. Go: go1.27.2 (T-5794 cut, the director's word).
- 16:25 The director approved the run. Run 39410125869 success: build (tag check, tests, vet, both tagged,
  the notes, GoReleaser, the dist checks, the hand-over) and publish (gh version, SHA-256s, sign, draft, draft checked,
  publish, checked as published; the draft deletion skipped). Release v1.0.0-rc.1: prerelease true, draft false,
  immutable true, six archives and checksums.txt.
- 16:25 Before the director's install (note 5.7.5, 5): the installed /usr/local/bin/bonsai (bf7cc3d, sha256
  70516eb8...a2478, as STATE records) copied to ~/bonsai-checks/prerelease/previous/ with its INSTALL.md (the way
  back); the 13 Oct one moved to ~/bonsai-checks/prerelease/older-a88436a/. Clean clones at the tag made by the
  orchestrator on both sides (~/bonsai-checks/rc1/clone; %USERPROFILE%\bonsai-checks\rc1\clone), since write-nothing
  roles cannot clone. Runner (sonnet) `M-2026-10-20-8b4f4162`, item 2's WSL side and item 5; V2 (opus, fresh)
  `M-2026-10-20-aea4c214`, item 2 itself and the run's log; Windows by the orchestrator.
- 16:33 Runner's report `M-2026-10-20-37e741fc` (sonnet, WSL; `~/bonsai-checks/rc1/wsl/`): gh release verify and 7
  verify-asset exit 0; gh attestation verify 12 of 12 (six archives, six programs) exit 0; sha256sum -c 6 of 6 OK;
  TestDist 9 PASS; install.sh from the real archive to a scratch target: exit 1 first (the target folder missing),
  then exit 0, --version "bonsai 1.0.0-rc.1 (commit 73919ddcc546)"; **the rebuild** linux_amd64 and windows_amd64
  (cross-built from WSL) byte-identical to the release (9883e243..., 06b7a457...); a plain scratch build's
  bonsai_release: newest null, not-a-release (its own version dev); the tap unchanged (pushed_at 2026-05-13T08:58:15Z,
  Formula/bonsai.rb sha 1625e90a...); the rc.1 lines from docs/install.md by sed: 9 of 22 lines differ, the version
  token only, the same count; every name and the tag match the release.
- 16:33 **Windows side, by the orchestrator** (%USERPROFILE%\bonsai-checks\rc1\win, Windows Go 1.27.2, gh 2.102.0):
  download exit 0; release verify and 7 verify-asset exit 0; six SHA-256s True; attestation of six archives and
  bonsai.exe exit 0; TestDist PASS (ok 4.076 s); install.ps1 from the real zip with -Target and a scratch BONSAI_HOME:
  exit 0, --version "bonsai 1.0.0-rc.1 (commit 73919ddcc546)", install.json's sha256 06b7a457...; the native rebuild
  from a CRLF checkout (core.autocrlf true) differed (c12fd565...) with identical go version -m settings; from an LF
  checkout (core.autocrlf false) **byte-identical** (06b7a457...): the difference is the checkout's line endings in
  embedded files and raw strings, not the release.
- 16:36 V2's first report `M-2026-10-20-f3e186e4` (needs-input, Windows ready): the run's log read step by step against note 5.7.2, 2 (own suite 106792300319 left out; the artifact digest equal both ends; one attestation of 13 subjects; the draft's seven SHA-256s equal); WSL: verify, verify-asset, sha256, attestation 12/12 with three negative controls failing, TestDist, the installer, all six programs rebuilt byte-identical, the archives' scripts equal the tag's, status newest null, the tap untouched. Its risk: the published notes end with CHANGELOG.md:48's line about the old product's sections, which do not follow. Its Windows commands run by the orchestrator as written (`%USERPROFILE%\bonsai-checks\rc1\v2win\`): every check rc=0, the wrong-ref control rc=1, TestDist PASS, install.ps1 rc=0; the CRLF clone's rebuild c12fd565... (as the orchestrator's own run), the LF fact passed to V2 to weigh.
- 16:37 **V2: PASS** `M-2026-10-20-3434f47c` (opus, fresh): fit for the director's install; both sides' checks pass; the WSL rebuilds byte-identical; the CRLF clone's Windows rebuild differs only inside the Go build ID (40 and 59 bytes, same sizes, same settings). Risk, for rc.2: the notes end with CHANGELOG.md:48's line about the old product's sections, which the release text does not carry. Batch 3 sent.
- 16:41 **Batch 3, the director's installs** ("both installed"). WSL read back: /usr/local/bin/bonsai "bonsai
  1.0.0-rc.1 (commit 73919ddcc546)", root:root 755, sha256 9883e243508c9f3e2a6c2bb8ef812fa23631a980274eed38c0d14c406f86a307
  (the release's), which -a names it alone, install.json the same; `bonsai check --json` in ~/Servers/Bonsai: no
  finding (one tables warning; "releases not read yet" note): Bonsai's own lines unchanged, no (ii) line. Windows read
  back: Get-Command names C:\Program Files\Bonsai\bonsai.exe first, "bonsai 1.0.0-rc.1 (commit 73919ddcc546)", sha256
  06b7a457... (the release's), install.json the same, gh attestation verify exit 0. Check 10's Windows half after rc.1
  (note 5.7.5, 8): still to run.
- 17:48 Left for later releases (T-5705 stays `verify`): check 10's Windows half after rc.1 (note 5.7.5, 8); rc.2's and 1.0's checks.
