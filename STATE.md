# Bonsai: where it stands

Rewritten, never appended. Last rewritten 8 Oct 2026 when Rohan's first sitting passed (`records/runs/R-2026-10-08-hook.md`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The design is settled and the plan is approved (8 Oct). Done so far:
- **Part 0, the formats** (`3770d04`): `formats/` holds a JSON Schema for each of the contract's ten formats, an example
  of each, the trick files with both outcomes in `expect.json`, a raw-byte manifest and a Go test. Now at **set 3**
  (`3a1f195`): 116 cases, 138 files; part 2 added 14 cases and the codes `format-too-new` and `not-text`.
- **Part 1, the clear-out** (`17f2938`): the old product's code is gone from `main` (358 files; 0.4.3 stays at its tag
  and in git history). The tree holds `cmd/bonsai` (a stub that answers `--version` only), `formats/`, `design/`,
  `records/` and the repo's own files. No dependencies; standard library only.
Both were passed by a fresh Opus verifier.
- **Part 2, the reader and files** (`3a1f195`): `internal/reader` (the format-1 reader, no YAML library; it reaches all
  116 format-1 outcomes), `internal/schema` (JSON, the byte-stable writer, the schema checker; `formats/embed.go` embeds
  the schemas), `internal/workspace` (`bonsai.yaml`, `pack.yaml`, the lock, the home, the checkout, path checks, atomic
  writes), `internal/status`, and `bonsai status [--json]` (every field of `bonsai.status/1`; those not built yet in a
  named, tested list). 667 tests on WSL and natively on Windows. Passed by a fresh Opus verifier after one fix round
  (51,000 break-it inputs: no accepted value differs from both YAML libraries; no crash).
- **Part 3, the engine** (`ee50971`): `internal/engine` and `bonsai init`, `update`, `check`: the plain fetch (git at a
  commit, source and ref after `--`), staged writes with the lock last, the kinds `pack`, `once`, `block`, `keys`,
  `kept`, exits 4 and 5, `--diff`, `--yes`, `--keep`, `--adopt`, the preview naming every settings line. A hook-line
  change is only refused (exit 4, `--allow-exec` named; the flag is step 5.1). Checks 1-6 and 12 pass as scripted runs
  over three scratch targets (check 5 up to part 4's clause). 703 tests on WSL and natively on Windows. It landed on
  green tests and the orchestrator's read of the diff; part 5's verifier also read its hook-line writes and refusal.
- **Part 5, the hook path** (`33a6122`): `internal/guard` and `bonsai hook guard`: one rule (an Edit, Write, MultiEdit
  or NotebookEdit of a path on `bonsai.yaml`'s protected or person-only list is refused, exit 2, reason and next
  step), failing closed (its own 5 s timer, half the hook's 10 s; a panic, bad input or an unrecordable allow blocks),
  paths judged in every form (on Windows the final path Windows reaches: streams, junctions, device paths, short
  names), one `bonsai.log/1` record per decision with the binary's path and hash. The fault switch only under
  `-tags bonsai_test_fault`; CI runs that tag too. `update` refuses when the lock is missing. Check 11 passes on WSL
  through `claude-here`. Hook start-up: WSL p50 2.4 ms (baseline 1.5), Git Bash 63 ms (baseline 65). Passed by a
  fresh Opus verifier after one fix round.
- **Rohan's first sitting passed** (hand checks b, c, d, 8 Oct): parts 3 and 5 are closed.
- **Part 4a, the test pack**: `LastStep/bonsai-test-pack`, public, commits A `5062053`, B `ab09e4b`, C `abfb5de`, D
  `1d4f46f` (`main`; full hashes in its run report). A shows a `marker` role saying "commit A"; B changes one pack file,
  adds one and says "commit B"; C changes that file again; D changes only the hook line. Its README says which check
  uses each commit.
Nothing else of the new Bonsai is built.

## Rohan's decisions, 8 Oct

- Bonsai's design and records live in this repo now (`design/`, `STATE.md`, `records/`); the studio keeps a pointer.
  Bonsai owns its spec and the formats contract.
- Work lands on `main` directly: no marker tag, no `rebuild` branch, no pull-request process. The old product stays
  reachable at tag `v0.4.3` and in git history.
- The JSON Schemas and the trick files are Bonsai's ("those are kind of tests which other projects can use"): the master
  lives in `formats/`.
- **Bonsai first**, to avoid interim code: the studio no longer adopts the formats early in its own Node code; it adopts
  them when it links (spec step 7): `bonsai init` there, its bridge reading Bonsai's outputs, its own guards, ladder and
  statusline retired. Nothing in the studio waits on Bonsai's parts; Bonsai's own order is unchanged.
- "The proper way, no shortcuts": the spec's order inside Bonsai; parallel work only where it is truly independent.
- Bonsai joins the studio's dashboard as its own project at spec step 6, as planned.
- Haiku joins the models, for small bookkeeping and audit jobs (listing, sorting, counting, checking a list against a
  source); it reports facts and judges nothing (`CLAUDE.md`).
- Rohan approved `design/plan.md` and `design/one-pager.md`, after a fresh Opus review and its fixes.
- Label-definition fields stay required in the formats: Bonsai's packs write every field; the contract's short examples
  are out of date (the plan's "Stale or in tension" list).
- Rohan's hand checks come in two sittings, not one: checks b, c and d (about 25 minutes) after part 5, check a (about
  20 minutes) after part 4. Parts 3 and 5 close once the first sitting passes, part 4 once the second does.

## GitHub, read 8 Oct

- `main`: part 1 (`17f2938`) on top of part 0, pushed 8 Oct. The only branch besides Dependabot's; the newest tag is
  `v0.4.3`.
- Open pull request #252: Dependabot's bump of an npm package in `/website`, which part 1 removed. Moot; Dependabot
  usually closes such a pull request itself. No agent merges or closes it.
- `release.yml` is disabled on GitHub and, from part 1, can only build (`workflow_dispatch`, `contents: read`, no
  secret, `goreleaser build --snapshot --clean`). No repository secret (the old `HOMEBREW_TAP_TOKEN` is deleted, and
  Rohan removed the token itself from GitHub). The only environment is `github-pages`; the `release` environment and a
  new tap token wait for step 5.7.
- Ruleset `main-protection`: blocks force pushes and deletion only, so pushes to `main` are plain pushes.
- Workflows: CI (`test` on Linux, `windows`, `lint`, `govulncheck` pinned to `v1.7.0`, the newest that builds with CI's
  Go 1.25), CodeQL, Dependabot (`gomod`, `github-actions`). The old website's Deploy Docs workflow is removed; the old
  site stays on GitHub Pages until Rohan turns Pages off.

## The one thing to do next: part 4b, packs as plugins

Part 4b (4-7 AI hours): the fetch at a 40-character commit, no login, read-only, on both sides; the marketplace named
by workspace and a hash of the locked commits; install and update on this machine (`--scope project` or `local` only,
never user); the drift report; `project-a` at A, `project-b` at B, `project-a-worktree` at B beside `project-a` at A; a
fresh worktree's trust. Check 8 with the session kinds recorded (interactive, `-p`, `--bg`, a fresh worktree) and
`claude --agent <plugin>:<role> --bg` starting the test pack's role; check 5's last clause (a new session loads B's
roles, `check` reports no drift). Check 10 on Windows. It lands on green tests on both sides and the orchestrator's read
of the diff. Recheck the user settings hashes after it (WSL `7b515457...a025a7`, Windows `2b6295c1...4ff6c9`). Then
Rohan's second sitting (hand check a, about 20 minutes), then part 6, the gate report.

Carried into part 4b: the residual fail-open risk the first sitting did not test (under the `missing` fault, a Write
whose payload exceeds the pipe buffer while the shell reads no stdin; Claude Code may turn the broken pipe into a
non-blocking status) gets an agent's headless run on WSL; and the scratch `claude-here.ps1` fails at `Get-FileHash`
(line 70) when it shows the found binary's hash, fixed before the second sitting.

## Waiting on Rohan

- Later: the second sitting (hand check a, about 20 minutes, after part 4); step 8 at 5.4; at 5.7
  the `release` environment and a new tap token (the working environment command is in spec §17 step 4's note).
- Whenever he likes: turn GitHub Pages off (the old website); close #252 if Dependabot has not.

## Loose ends

- Go 1.25.9 (the module's `toolchain` line) has 19 standard-library vulnerabilities and 1 in an imported package that
  govulncheck lists but our code does not reach; a later 1.25.x patch fixes them. A one-line `toolchain` bump, as its
  own small commit.
- Spec §17 step 3 is done: all four old `bonsai` binaries are gone (8 Oct).
- For step 5.3, from part 5's verifier: a `.git` entry in the session's starting subfolder makes the guard read the
  project as unlinked; `bonsai.yaml` (what is protected) is read from the working tree; `bonsai` is found by the PATH,
  which a settings `env` could redirect; shell `rm` or `mv` of `bonsai.yaml` or the lock is not judged; a dangling
  junction into a protected folder is allowed (the write fails); a junction swapped between check and write is not
  caught; the admin share `\\localhost\C$` is a documented limit. For step 5.1: a plugin's move to a new commit is
  not counted as running code (spec §5 lets plugins carry hooks); the log's `bonsai_path` and `bonsai_sha256` names
  wait for step 5.2; CI's `lint` does not use the fault tag.
- Spec step 5.3 (the guard): the hook lines must not be redirectable by files an agent may edit. They call the installed
  `bonsai` by a fixed path, and what the guard trusts to find the main checkout cannot be rewritten without the guard
  noticing (fail closed). See `design/plan.md`, "What still links Bonsai and the studio", and spec §7.
- From part 2's verifier, minor, not fixed: a quoted `"format":` key followed by a tab still dispatches as format 0 (a
  plain key with a tab goes to format 1); a file with lone-CR line endings is refused as `quote-this-value`, whose next
  step does not help (before the order rule it said `not-text`). Both want a case in a later set. The status text lacks
  spec §6's `bonsai init --new-id` line until `init` exists (part 3).
- `status --json`'s `formats` field stays null: its schema needs a `write` major of at least 1, so a format Bonsai only
  reads (`pack.yaml`) cannot be listed; step 5.1 says what `formats` lists. No `error` object until step 5.1.
- Small wording items from part 1's verifier, not fixed: `README.md` describes the designed product in the present
  tense (framed by "being rebuilt"); `CONTRIBUTING.md`'s note that an older `go` downloads the toolchain holds from
  Go 1.21; `.gitattributes` keeps old-product patterns (harmless).
- The spec's §18 and §19 D give both question D's figures (60 h, 30-46 h) and the format review's (61 h, 30-47 h);
  §14 and §20 say 61 and 30-47. The plan uses 61 and 30-47.
- The spec's §17 step 6 still has the hand checks in one sitting; the plan follows Rohan's two.
- Scratch left from parts 0 and 1 under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\` (clones, bundles,
  builds, the YAML libraries). Safe to delete; each later Windows run makes its own folder.
