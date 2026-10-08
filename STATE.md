# Bonsai: where it stands

Rewritten, never appended. Last rewritten 8 Oct 2026 when the test pack was published (`records/runs/R-2026-10-08-test-pack.md`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The design is settled and the plan is approved (8 Oct). Done so far:
- **Part 0, the formats** (`3770d04`): `formats/` holds a JSON Schema for each of the contract's ten formats, an example
  of each, 97 trick files with both outcomes in `expect.json`, a raw-byte manifest (124 files, set 1) and a Go test.
- **Part 1, the clear-out** (`17f2938`): the old product's code is gone from `main` (358 files; 0.4.3 stays at its tag
  and in git history). The tree holds `cmd/bonsai` (a stub that answers `--version` only), `formats/`, `design/`,
  `records/` and the repo's own files. No dependencies; standard library only.
Both were passed by a fresh Opus verifier.
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

## The one thing to do next: part 2, the reader and files

Next, in the plan's order: part 2, then 3, 5, the rest of 4, and 6 (`design/plan.md`, "Parts 1-6"). Part 2 is the
format-1 reader in Go (no general YAML library; spec §3, contract §2.4), tested on every trick file in `formats/`;
`bonsai.yaml`, `pack.yaml` (the test pack's at commit A), the lock, and the partial `status --json` (every field, a
named list of those still `null`); the schema checker moves out of part 0's test. It is "the reader": an Opus builder
and a fresh Opus verifier (with part 1's, as the plan's shared verifier), and check 10 natively on Windows before the
push. Open `records/runs/R-<date>-reader.md`, make the worktree, and brief from part 2's row and "The partial
`status --json`".

For part 2's reader, from part 0: the frontmatter is its lines with their own line endings (`formats/README.md`); cut
any other way, YAML libraries disagree on a final block scalar's newline. From the test pack: its hook line is declared
in `pack.yaml` under `hooks`, a reading of the plan's "one hook line" that step 5.1 should confirm (spec §7 lists only
Bonsai's own hook lines).

## Waiting on Rohan

- Before part 5's hook checks: spec §17 step 3, the old `bonsai` binaries (the lines are in the spec).
- Later: the two hand-check sittings (after part 5, about 25 minutes; after part 4, about 20); step 8 at 5.4; at 5.7
  the `release` environment and a new tap token (the working environment command is in spec §17 step 4's note).
- Whenever he likes: turn GitHub Pages off (the old website); close #252 if Dependabot has not.

## Loose ends

- Go 1.25.9 (the module's `toolchain` line) has 19 standard-library vulnerabilities and 1 in an imported package that
  govulncheck lists but our code does not reach; a later 1.25.x patch fixes them. A one-line `toolchain` bump, as its
  own small commit.
- Spec step 5.3 (the guard): the hook lines must not be redirectable by files an agent may edit. They call the installed
  `bonsai` by a fixed path, and what the guard trusts to find the main checkout cannot be rewritten without the guard
  noticing (fail closed). See `design/plan.md`, "What still links Bonsai and the studio", and spec §7.
- Small wording items from part 1's verifier, not fixed: `README.md` describes the designed product in the present
  tense (framed by "being rebuilt"); `CONTRIBUTING.md`'s note that an older `go` downloads the toolchain holds from
  Go 1.21; `.gitattributes` keeps old-product patterns (harmless).
- The spec's §18 and §19 D give both question D's figures (60 h, 30-46 h) and the format review's (61 h, 30-47 h);
  §14 and §20 say 61 and 30-47. The plan uses 61 and 30-47.
- The spec's §17 step 6 still has the hand checks in one sitting; the plan follows Rohan's two.
- Scratch left from parts 0 and 1 under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\` (clones, bundles,
  builds, the YAML libraries). Safe to delete; each later Windows run makes its own folder.
