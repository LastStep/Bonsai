---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6020
title: "One-command installers in the repo for anyone: install/get-bonsai.sh and install/get-bonsai.ps1 download a release, check its SHA-256 and provenance and run its own installer; docs/install.md and the README name them"
status: done
lane: null
done_when:
  - "install/get-bonsai.sh (Linux, WSL) and install/get-bonsai.ps1 (Windows PowerShell 5.1): the newest release or a named version; the archive's SHA-256 against its checksums.txt; gh attestation verify as docs/install.md's lines; the release's own install.sh or install.ps1; stop at the first failure, nothing installed; plain refusals naming the next step (gh missing or not logged in, a bad version, an unknown machine); self-documenting headers; no person's path or name"
  - "Tests: both parse (sh -n; the PowerShell parser in CI's windows job or a Go test that runs it where powershell exists, skipped with its reason elsewhere); the download names and the version handling; a refusal for a bad version; a dry run against the real latest release stopping before the installer, by hand, in the report"
  - "docs/install.md and README.md: a one-command section fetching the scripts from main's raw files (no 1.0.4 yet, the director's word of 6 Nov), every check kept, the manual lines left as they are; the selfcheck install test still holds"
  - "Climbed green (bonsai ladder --task T-6020); a fresh Opus verifier passes it (security); landed with CI green; no release"
depends_on: [T-6019]
blocked_by: null
created: 2026-11-06
started: 2026-11-06
finished: 2026-11-06
labels:
  bonsai.ladder: []
  bonsai.allows: [".gitattributes"]
  bonsai.branch: T-6020-get-bonsai
  workflow.model: opus
---

# One-command installers for anyone

The director, 6 Nov: "can we move these installer scripts within bonsai repo, especially if nothing is hardcoded in
them ... so anyone can use", then "go ahead with T-6020. but we are not putting 1.0.4 just yet". The director's own
scripts (5 Nov, T-6012's run report), copied into `~/bonsai-checks/T-6020/`, are the start: the only repository they
name is LastStep/Bonsai; their usage comments name the director's scratch folder, which goes.

## Out of scope

- A release (1.0.4): the scripts reach anyone from main's raw files once landed.
- Shipping the scripts inside release archives (.goreleaser.yaml is person-only).
