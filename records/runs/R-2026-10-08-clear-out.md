# Run: part 1, the clear-out and the new layout

- Date: 2026-10-08
- Who: one builder (Opus) and one fresh verifier (Opus, before the push), briefed by the orchestrator; the
  orchestrator is this report's only writer and commits it after the fast-forward
- Branch: `clear-out`, from `main` at `d37a5fb`, in the plain worktree `~/Servers/Bonsai-clear-out`
- Plan: `design/plan.md`, parts 1-6, part 1's row and "CI and release after part 1" (2-4 AI hours; the skeleton's stop
  line is 61 h for parts 1-6)
- Outcome: part 1 done, passed by a fresh verifier before its push; on `main` at `17f2938`

> Opened before the first edit and appended to as the work went.

## Log

- 14:34 Part 0 closed: CI on `d37a5fb` (the state commit on top of `3770d04`): `test`, `lint`, CodeQL green;
  `govulncheck` red at its install step, as before part 1.
- 14:35 Worktree `~/Servers/Bonsai-clear-out` on branch `clear-out` from `main` at `d37a5fb`. Briefed one Opus
  builder: part 1's row whole, "CI and release after part 1", the rules from `CLAUDE.md`; the proof on WSL and natively
  on Windows; no push.
- 14:47 The builder's report: ten commits on `clear-out` (`b8095a4`..`28227f7`). `b8095a4` is the one removal commit,
  358 deletions and nothing else: `internal/` 130, `cmd/` 21, `catalog/` 143, `website/` 54, `docs/` 8, `embed.go`,
  `.github/workflows/docs.yml`. Kept: 164 tracked files. New `cmd/bonsai`: a standard-library stub, `--version` prints
  `bonsai <version>` (`main.version`, default `dev`, set by `-ldflags -X` as `.goreleaser.yaml` does), anything else
  exits 2 naming the next step; 7 tests. `go.mod` keeps `go 1.25.0`, `toolchain go1.25.9` and the `retract v0.1.0`;
  every dependency gone, `go.sum` empty. `Makefile` without `install`; `build` writes to `$(OUT)`. CI: `test` and
  `lint` unchanged; a new `windows` job (`go test`, `go vet`); govulncheck pinned to `v1.7.0` (the proxy's `.mod`
  files: v1.2.0-v1.7.0 need `go 1.25.0`, v1.8.0 needs 1.26; locally `@v1.7.0` builds on 1.25.9 with
  `GOTOOLCHAIN=local` and finds nothing our code calls). `release.yml`: `workflow_dispatch` only, `contents: read`, no
  secret, `goreleaser build --snapshot --clean`; `.goreleaser.yaml` unchanged (no `before:` hook, no removed path).
  `docs.yml` absent from every new commit's tree, so the push cannot run it. README, CONTRIBUTING, CHANGELOG
  ("Rebuild - Unreleased"), SECURITY (one sentence) and the feature template rewritten for the rebuild.
  Proof: baseline 18 packages ok, 660 passed; after 2 packages ok (`cmd/bonsai` 7, `formats` 15), 22 passed, 0 failed;
  `GOOS=windows` build and vet ok; golangci-lint 2.11.4 0 issues; `go mod tidy` no diff, `go mod verify` ok; the
  workflows parse; the six goreleaser targets build with its ldflags (GoReleaser itself not installed). Windows:
  `clear-out-src` (Windows git, `core.autocrlf` true) at `28227f7`, `go.exe` 1.26.2: `ok cmd/bonsai`, `ok formats`,
  22 passed, vet clean, `bonsai.exe --version` prints `bonsai dev`. Noted: govulncheck reports 20 standard-library
  vulnerabilities in Go 1.25.9 that our code does not call, fixed by 1.25.13 (a later small toolchain bump).
- 14:49 Briefed one fresh Opus verifier for part 1, before the push: part 1's row, "CI and release after part 1", the
  diff, the tests re-run on WSL and in its own Windows clone, and the private grep. It passes or fails; it fixes
  nothing.
- 14:54 The verifier's verdict: PASS, nothing must be fixed before the push. The removal equals the plan's list
  exactly (358 files; every keep-list path present; no commit touches `design/`, `records/` or `formats/`);
  `release.yml` can only build and is `disabled_manually` on GitHub; CI as the plan says; `docs.yml` absent at the
  pushed head; `go.mod` standard library only; the stub ASCII, refusals naming the next step, tests on buffers only.
  govulncheck pin re-checked: the proxy's `.mod` files and a local `go run ...@v1.7.0` under `GOTOOLCHAIN=local` pass,
  `@v1.8.0` fails on Go 1.25.9. The builder's deliberate leftovers all accepted. Proof re-run: WSL 2 packages ok, 22
  passed, 0 failed, 0 skipped; `GOOS=windows` vet and build ok; lint 0 issues; tidy no diff; its own Windows clone
  (`verify-clear-out-src`) `go.exe test ./...` 22 passed, vet clean. Private grep: none. Should-fix: `SECURITY.md:30`
  promised security updates for "the latest release", now 0.4.3, which gets none. Minor: the issue chooser links to
  Discussions, which are off; the README's present tense; CONTRIBUTING's Go 1.21 note; old `.gitattributes`
  patterns. Info: `STATE.md` named a studio task id; Dependabot opened #252 (npm in `/website`), moot after the
  clear-out.
- 14:56 The orchestrator checked `.github/dependabot.yml` (only `gomod` and `github-actions`, nothing for the website)
  and made the two one-line fixes itself on `clear-out`: `SECURITY.md` now says 0.4.3 gets no further releases,
  security fixes included; the Discussions link is dropped from the issue chooser (`17f2938`). Left the README tense,
  the CONTRIBUTING note and `.gitattributes` as loose ends in `STATE.md`. #252 is not ours to close.
- 14:58 Fast-forwarded `main` to `clear-out` (`17f2938`). Rewrote `STATE.md`: parts 0 and 1 done, the test pack next,
  the studio task id dropped, CI and the website's workflow as they now are.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder: part 1 | Opus | 14:36 | 14:47 | 11 |
| Verifier: part 1 | Opus | 14:48 | 14:54 | 6 |

Times are local (8 Oct). Skeleton hours so far (parts 1-6, stop line 61 h): 17 minutes (part 1: builder 11, verifier 6).

## Option rounds asked of Rohan inside the skeleton (stop line: more than two)

None yet.

## Windows-only failures (stop line: 8 hours)

None yet.
