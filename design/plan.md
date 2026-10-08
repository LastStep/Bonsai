# Bonsai's plan: the formats, then the walking skeleton

- **Status:** draft. Rewritten on 8 Oct 2026 for work inside Bonsai's repo, from the studio's drafts (the skeleton's
  plan and task, reviewed twice; the readers plan's step 1, reviewed once). Waits for a fresh review and Rohan's
  approval.
- **Design:** `design/bonsai-spec.md` (the spec: §14 the parts, the twelve checks and the stop lines; §15 what the gate
  measures; §17 Rohan's steps), `design/contract.md` (the formats: §2.4 the YAML rules and the trick files),
  `design/one-pager.md`, `design/format-review.md`.
- **Hours** (AI hours, verification included): part 0, 6-10, stop line 13; the skeleton, parts 1-6, 30-47, stop line 61.
- **Records:** `STATE.md`; run reports in `records/runs/`.

Two readers. **Rohan** reads down to "Size" and reads no code. The **orchestrator, builders and verifiers** read the
rest.

## For Rohan (plain words)

**What this does.** Two pieces of work, in this order.
1. **The formats (part 0).** Written rules for every file Bonsai and the studio both read (one schema per format), and a
   set of small, deliberately tricky example files, each with its one right answer: accepted with this value, or
   refused for this reason. They are tests any reader can run. The studio tests its own reader on a copy pinned to a
   Bonsai commit.
2. **The walking skeleton (parts 1-6).** The risky parts of the new Bonsai, built on throwaway copies of projects, to
   prove before you decide whether to go on: Bonsai can write and update a project's files without overwriting your
   edits; a pack can arrive as a Claude Code plugin at an exact version, with two projects (and a worktree beside its
   checkout) each loading their own; and the guard blocks what it should, and also blocks when it is missing, crashes,
   is slow or cannot be found. It ends with a report of measured numbers, and you pick the path: the full Bonsai 1.0,
   the smaller cut, or 1.0 with its own screens.

**What you will see.** Commits on Bonsai's `main`, each pushed after its proof passed (a fresh verifier where the work is big or risky), with green checks on
GitHub (Linux, and Windows from part 1). Part 1 clears the old product's code out of `main`; 0.4.3 stays downloadable
and at its tag. A new public repository, `LastStep/bonsai-test-pack`, for part 4. No release and no tag. Then one
sitting from you, about 45 minutes, on your hand checks. Then the gate report.

**What you must do.**
- Approve this plan and the one-pager (`design/one-pager.md`). Nothing is built before.
- Confirm that the old Homebrew tap token is revoked on github.com (spec §17 step 4; not confirmed on 8 Oct).
- Before part 5's hook checks: your step 3, the old `bonsai` binaries (spec §17 step 3 has the lines).
- One sitting of about 45 minutes when the scratch build is ready (spec §17 step 6). It saves about 3 AI hours. Each
  check has a plain pass condition below; tell the orchestrator what you saw, one line per check.
- Later, not in this plan: step 8 (the pre-release at 5.4) and, at 5.7, the `release` environment and a new tap token.

**Choices that are yours.** None in this plan beyond approving it.

**Risk.** Windows is where the old code failed (29 tests); the stop line for that is 8 hours lost to Windows-only
failures. A Windows session may not load a pack the way WSL does; finding that out is what the skeleton is for, and the
answer goes in the gate report either way. Agents act on GitHub as your account, an admin, so the plan forbids them every
release, tag and settings change. Work lands on `main` with no pull request, as you chose: if you pause, `main` holds the
work so far and the old product stays at its tag. If any of the four stop lines is crossed, work stops and you get the
numbers and three choices: continue, the smaller cut, or pause.

**Size.** Part 0: 6-10 AI hours; at 13 work stops and you are asked. The skeleton: 30-47 AI hours across six parts; at
61 work stops and you are asked. On the studio's record (estimates grow 1.6 times) expect 10-16 and 48-75. Plus about 45
minutes of yours. Nothing here waits on the studio.

## For the orchestrator, builders and verifiers

### How the work runs

- **The orchestrator** is the Claude Code session Rohan opens in `~/Servers/Bonsai` (the main checkout, on `main`). It
  keeps its context lean, dispatches builders and verifiers with full briefs (Bonsai has no role files: each brief
  carries the rules of `CLAUDE.md` and of this plan that the job needs), reads their reports, merges and pushes.
- **Builders** work in a plain git worktree beside the clone, one per part:
  `git -C ~/Servers/Bonsai worktree add ~/Servers/Bonsai-<part> -b <part> main` (never the Agent tool's isolation
  worktrees). They commit on that branch and never push. Builds go `go build -o` into a scratch folder
  (`~/bonsai-checks/bin` for the skeleton); never `go install`. `~/ZenGarden/Bonsai` is never touched.
- **Verifiers** are fresh Opus agents, for big or risky work only (Rohan, 6 Oct: fewer verifications): part 0 (the set
  everyone tests against), part 2 (the reader, sharing its verifier with part 1's CI and release changes), part 5 (the
  hook path), and the last verifier at the skeleton's end over all twelve checks. Parts 3, 4 and 6 land on green tests
  on both sides and CI plus the orchestrator's read of the diff, which the run report says; the last verifier covers
  them. A verifier reads this plan's part, the spec and contract sections it cites and the diff, re-runs the tests itself, and
  passes or fails the part. It fixes nothing.
- **Landing a part.** After its proof passes: `git -C ~/Servers/Bonsai merge --ff-only <part>`, then
  `git -C ~/Servers/Bonsai push origin main`, then CI for that commit:
  `gh api repos/LastStep/Bonsai/commits/<sha>/check-runs --jq '.check_runs[] | [.name, .status, .conclusion] | @tsv'`
  (or `gh run list -R LastStep/Bonsai -L 10`; this machine's gh, 2.4, has no `--branch`). Red CI is fixed forward with a
  new commit; force pushes are blocked. Before each push the orchestrator fetches and reads `main`'s new commits on the
  remote: one it did not make (an old cloud session, a Dependabot merge) stops the work until it is understood. Every
  commit of this work is listed in its part's run report.
- **GitHub.** Agents act as `LastStep`, Rohan's admin account. Only the orchestrator pushes: Bonsai's `main` and the
  test pack's `main`. Never: a tag, a release (`gh release` anything), a workflow switched on, a setting, ruleset or
  secret changed, a pull request merged or closed, a branch deleted on GitHub. Releases are Rohan's word at step 5.7.
- **Run reports**, one per part, `records/runs/R-<date>-<part>.md`, opened before the first edit and appended as the
  work goes: a log, not a summary. Each lists every builder and verifier run with its model, start, end and minutes, and
  the orchestrator keeps a running total for part 0 and for the skeleton. These rows are the hours source for the stop
  lines: the studio's session records do not cover Bonsai.
- **Models.** Opus for builders on the formats, the engine, the guard, Windows and plugin delivery, and for every
  verifier and plan review; Sonnet for mechanical runs (cross-compiles, measurements into tables, CI checks). Each run
  report names the model of every run.
- **Nothing private and nothing studio-shaped in what this work makes.** Code, tests, `formats/`, the test pack, commit
  messages and run reports name no studio task id or studio path, no one's home folder (`/home/...`, `C:\Users\...`;
  write `~/...` and `%USERPROFILE%`), no machine or tailnet name, no email address. `design/` holds the specs as the
  studio wrote them, citing its files; that is quoting, not new work.
- **Tests** make their own repos and homes in `t.TempDir()`: never the real `~/.bonsai`, `~/.claude` or a real project.
- **Processes.** Every agent stops what it starts and checks with `ps -eo pid,etime,cmd`. The orchestrator sweeps after
  each subagent (Go builds and tests, `claude` test sessions started from `~/bonsai-checks`) and kills by pid only what
  that agent clearly left.
- **Windows.** Windows Go cannot build from WSL's disk (it cannot lock `go.mod` over the share). Windows runs happen
  under `%USERPROFILE%\bonsai-checks\`, which WSL reaches as
  `WINHOME=$(wslpath "$(cmd.exe /c 'echo %USERPROFILE%' 2>/dev/null | tr -d '\r')")`, with Windows Go
  (`"/mnt/c/Program Files/Go/bin/go.exe"`). Every Windows-side repo, worktree and CRLF checkout is made with Windows git
  by its full path (`"/mnt/c/Program Files/Git/cmd/git.exe"`): a worktree made by Linux git stores a `/mnt/c/...` path
  that breaks on Windows. Linux git never runs in a Windows checkout.

### What still links Bonsai and the studio

The work waits for nothing in the studio. What links the two:
1. **The studio pins `formats/`.** Its readers task takes a copy at the commit `STATE.md` names when part 0 is done and
   tests its own reader against it. A fix to `formats/` after that is a new commit, named in `STATE.md` and told to
   Rohan, never a silent edit.
2. **Format 0's outcomes come from the studio's reader.** Format 0 means "exactly what the studio's `yaml.mjs` reads
   today" (contract §2.3-§2.4). Part 0 fills each case's format-0 outcome by running that reader read-only from the
   studio's checkout on this PC; the orchestrator gives the builder the path and the commit (the one the studio froze as
   its format-0 reference). The studio's test of its frozen copy re-checks every format-0 outcome when it pins.
3. **Check 7** (Bonsai's reader and the studio's reach the same outcome on every trick file) has two halves: Bonsai's
   reader reaches every format-1 outcome in `formats/expect.json` (part 2's test), and the studio's reader reaches the
   same file's outcomes (its readers task). The gate report says whether the studio's half has run.
4. **Later:** the studio's slice 3 and its registration of projects (contract §15.2) come after the skeleton; Bonsai
   joins the studio's dashboard as its own project at spec §14 step 6; the studio links to Bonsai at step 7, Mimas last
   at step 8.
5. The studio's own work may ask Rohan questions in the same days (the spec's "two queues"). Stop line 3 counts only
   this plan's option rounds.

### Part 0: the formats (6-10 h, stop line 13 h)

Bonsai's first job (Rohan, 8 Oct: "those are kind of tests which other projects can use ... it makes sense for them to
live in bonsai"). It lands on `main` before part 1, beside the old product's code, in a folder part 1's clear-out keeps.
Sources: contract §2.4 (the YAML rules and the trick files), §2.2 and §2.5 (fields, versions, JSON), and each format's
section: §4 task, §5.2 labels, §6 lanes, §7.1 run, §7.2 state, §8.1 log, §9.1 ask, §11 ladder, §12 status, §14 lock.

**What it builds, all under `formats/`:**
- `README.md`: what the folder is and who reads it, how a case is laid out, every field of `expect.json` and
  `manifest.json` with an example, the reason codes, and how the set changes (a new case, a new manifest, a new commit
  named in `STATE.md`). The folder documents itself (contract §2.8).
- `schemas/<name>.schema.json` for the ten formats: `task`, `labels`, `lanes`, `run`, `state`, `log`, `ask`, `ladder`,
  `status`, `lock`. JSON Schema draft 2020-12. Properties in the contract's fixed order, every field required (`null` or
  `[]` where it does not apply, contract §2.2), closed lists as `enum`, open lists as strings. Not here: the two
  generated tables (`bonsai.tasks/1`, `bonsai.sessions/1`), `bonsai.yaml` (`bonsai.workspace/1`) and `pack.yaml`
  (`bonsai.pack/1`), all step 5.1; memory, which has no format of its own. Nothing here reads or writes a log or an ask.
- `examples/`: one valid document per format, made-up values.
- `trick/yaml-1/<rule>/` and `trick/yaml-0/<oddity>/`: the input files, markdown with frontmatter or plain YAML
  (definition) files, made-up content only. Format-1 cases start with a `format:` line, as a real file would. **One
  case per rule of contract §2.4:** LF and CRLF lines; a BOM (frontmatter, and a definition file); a tab in indentation;
  `---` and `...` anywhere but the frontmatter's own markers; every line consumed, a list at its key's own indent among
  them; the key rule: quoted keys, complex keys, `<<`, a key twice in one mapping, the reserved words as keys (`y`, `n`,
  `yes`, `no`, `on`, `off`, `true`, `false`, `null`); nested mappings; block sequences (deeper than their key, exactly
  one space after `-`); one-line flow sequences, `[]` and `{}`; anchors, aliases, tags, flow mappings with content,
  nested and multi-line flow sequences (all refused); block scalars `|`, `|-`, `>`, `>-`, a line starting with `#`
  inside one, a deeper line in `>`; comments at a line's start and after a space, and a `#` with no space before it; a
  quote character inside a plain value; quoted scalars on one line, the five escapes, a bad escape, an unescaped `"`,
  something after the closing quote; **every row of the plain-scalar table** (null forms, `true` and `false`, the
  integer at 15 digits and at 16, a decimal, both date forms read as text, a text value, the "quote this value"
  refusal); **each listed refusal** (`0755`, `55227e5`, `0x1F`, `1e3`, `1_000`, `.inf`, `TRUE`, `No`, `5 arenas`,
  `.claude/**`, a value holding `: `), and each read as text once quoted. Plus: a duplicate key quoted and unquoted;
  **the format-0 oddities** (a plain value holding `: `, unquoted hashes, `TRUE`, the last of two duplicate keys,
  document markers skipped); and **the dispatch cases**: `format:` first, no `format:` key, and `format:` not first
  (refused).
- `expect.json`: one language-neutral file giving every case's outcome under each format, that is, the file read by a
  format-0 reader and by a format-1 reader: `accepted` with the value as JSON, or `refused` with a short reason code
  (format 1; the codes are listed in `README.md`) or the reader's own message (format 0). Under format 1 a date stays
  text and an integer has at most 15 digits.
- `manifest.json`: a set version and the SHA-256 of each file's raw bytes (never line-ending-normalised: the CRLF case is
  the point), with forward-slash paths, for every file under `formats/` except itself and the Go files.
- In the repo's `.gitattributes`: `formats/** -text`, on a line **after** the existing `*.yaml text eol=lf` (the last
  matching line wins), so no checkout, Windows' `core.autocrlf` included, changes a byte.
- `formats/*_test.go`: standard library only, a test-only package that needs no reader and survives part 1. It checks
  that:
  - the manifest matches every file's raw bytes, lists every file and nothing more;
  - the CRLF case holds `\r\n` and the BOM cases start with `EF BB BF`, as checked out;
  - every rule of §2.4 has a case (a short hand list of the rules in the test), and every case has both outcomes in
    `expect.json`;
  - every schema is valid JSON and declares draft 2020-12, and its example validates under a small checker in the test,
    limited to the keywords the schemas use (part 2 moves the checker into the new layout);
  - no file in the set holds an absolute path (`/home/`, `/mnt/`, `~/`, a drive letter), an email address or a tailnet
    name.

**The format-0 outcomes** (link 2 above): a small script in the builder's scratch folder, never committed, runs the
studio's reader on each case (`parseYaml` for a YAML file, `parseFrontmatter` for markdown) and writes the value or the
exact error message into `expect.json`. The run report records the studio commit it ran and how many cases it filled.

**Done when:**
- `go test ./...` and `go vet ./...` pass in WSL: the new tests pass, and the old product's tests are no worse than before
  (part 0 changes no product code).
- A fresh Opus verifier reads every case against contract §2.4 **rule by rule, not against any reader**: one agent wrote
  both the cases and the answers. Is each rule covered; is each format-1 outcome what the grammar says; is each format-0
  outcome what the studio's reader gives (it re-runs the script itself). It also reads each schema against its contract
  section and greps the set for private strings.
- After the push: CI's `test` job green on the commit. Then once, natively on Windows: a Windows-git clone of `main` in
  `%USERPROFILE%\bonsai-checks\src` where Windows Go passes `go test ./formats/` (only this package: the old product's
  own tests fail on Windows). It proves the CRLF and BOM bytes survive a Windows checkout. From part 1 the `windows` CI
  job runs it on every push.
- `STATE.md` names the commit (its full SHA) for the studio to pin, and the orchestrator tells Rohan in one line.

### Parts 1-6: the walking skeleton (30-47 h, stop line 61 h)

Spec §14 has the parts, the twelve checks and the stop lines; §15 what the gate measures; §17 step 6 Rohan's hand checks
with their exact lines. This section says how they are built here.

**Order: 1 → the test pack (the first piece of part 4) → 2 → 3 → 5 → the rest of 4 → 6.**
- The test pack before 2 and 3: part 2's reader reads its `pack.yaml`, and part 3's checks fetch it by its public URL
  (check 2: the lock's source is the pack's remote URL).
- 3 before 5: in a project with no `bonsai.yaml`, `bonsai hook` allows everything (spec §3, §7), so the guard needs part
  2's reader for the protected list in `bonsai.yaml` (hand check c's `protected.txt`), and a real session needs the hook
  lines part 3's `init` writes.
- 3 before the rest of 4: the plugin wiring, the marketplace entry and `update` are part 3's engine.
- 5 before the rest of 4: neither needs the other; the hook path is smaller and Windows hooks are the likeliest stop, so
  that is found first.

| Part | Built | Proved by | Hours |
|---|---|---|---|
| 1. The clear-out and the new layout | One commit on `main` removes the old product's code: `internal/`, `cmd/`, `catalog/`, `website/`, `docs/`, `embed.go`, `.github/workflows/docs.yml` (the website's; the old site stays on GitHub Pages until Rohan turns Pages off). It keeps `LICENSE`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md` (rewritten for the rebuild), `CHANGELOG.md` (a new "rebuild" section), `.gitattributes`, `.gitignore` (less its old-product lines: the website, `station/`, the old lock name), `.golangci.yml`, `assets/`, the rest of `.github/`, `.goreleaser.yaml`, and all 8 Oct added: `CLAUDE.md`, `STATE.md`, `design/`, `records/`, `formats/`. It rewrites `README.md` ("being rebuilt; 0.4.3 is the old product's last release"), `Makefile` (no `go install`; `build` is `go build -o`), `go.mod` (`go 1.25` with a `toolchain` line; the standard library, and `golang.org/x/sys` only once needed) and `go.sum`, and adds `cmd/bonsai` with a stub `main` (`bonsai --version`), so `go vet` and CodeQL's autobuild find a package. CI and `release.yml` as below. | CI green on the pushed commit (`test` and `windows`); `go vet`; `release.yml` can only build; the clear-out's file list holds only what it should | 2-4 |
| 4a. The test pack | `LastStep/bonsai-test-pack`, public, made by the orchestrator with `gh` once Rohan approves this plan. Commits: **A** (a `marker` role saying "commit A", the files `init` writes into a project, one hook line); **B** (marker "commit B", one of those files changed, one new file); **C** (that file changed again); **D** (one hook line changed, nothing else). | `claude plugin validate --json` with no warning but the missing `version`; part 2's reader reads its `pack.yaml` | in part 4 |
| 2. Reader and files | The format-1 reader in Go, no general YAML library (spec §3, contract §2.4), tested on the cases in this repo's own `formats/`; `bonsai.yaml`, `pack.yaml`, the lock (`.bonsai/lock.json`), a subset of `status --json` (below). No format 0 (step 5.1). The schema checker moves from part 0's test into the new layout. | Committed Go tests reach every trick file's format-1 outcome in `formats/expect.json`; the lock and `status --json` validate against part 0's schemas; the Windows rules (forward slashes, line-ending-blind hashes, byte-stable output) | 7-10 |
| 3. Engine | The fetch in its plain form (git, at a commit, from the pack's URL); `init`, `update` (staged; kinds `pack`, `once`, `block`, `keys`; `kept`; exits 4 and 5; `--diff`, `--yes`, `--keep`, `--adopt`), `check` (lock and files, a tracked or staged `.bonsai/local/` file); `bonsai.yaml` with a comment on every line; `.bonsai/lock.json` and `.bonsai/.gitignore`; `init`'s closing plain words; the preview naming every settings line with its sentence. A hook-line change only refused: exit 4, nothing written, `--allow-exec` named as the next step (the flag itself is step 5.1). Scratch targets: a drifted project (one old absolute hook line, one hook of its own), a fresh empty repo, and a scratch clone of the studio's repo with `origin` removed (the orchestrator gives its source), each with `init --new-id`. | Checks 1-6 and 12 as scripted runs over those targets (check 5 up to its new-session clause, which needs part 4), A to B for check 5, A to B to C for check 6, D for the hook-line refusal; committed Go tests of the same behaviour on `t.TempDir()` repos; check 9 (hand check b) | 10-15 |
| 5. Hook path | `bonsai hook guard` with one rule: an Edit of a path in `bonsai.yaml`'s protected list is refused (`project-guard` lists `protected.txt`). The fault switch `BONSAI_TEST_FAULT` (`missing`, `crash`, `slow`, `minimal-path`) behind a Go build tag, every fault only blocking. The binary's path and SHA-256 logged. | Check 11 on WSL (the builder, through `claude-here`) and on Windows under Git Bash (hand checks c and d); a Go test per fault; a normal build holds no fault code | 3-5 |
| 4b. Packs as plugins | The fetch at a 40-character commit, no login, read-only, on both sides; the marketplace named by workspace and a hash of the locked commits; install and update on this machine; the drift report; `project-a` at A, `project-b` at B, `project-a-worktree` at B beside `project-a` at A; a fresh worktree's trust. | Check 8 with the session kinds recorded (interactive, `-p`, `--bg`, a fresh worktree) and `claude --agent <plugin>:<role> --bg` starting the test pack's role; check 5's last clause (a new session loads B's roles, `check` reports no drift); hand check a | 4-7 |
| 6. Measure and report | The gate report: hours per part against estimate and the ratio for step 5; each check; Windows-only failures with hours; hook start-up p50 and p95 on both sides against the baseline (1.5 ms WSL, 65 ms Git Bash); fail-closed results; plugin install and update time; two commits kept; the Claude Code version on each side; binary size; lines of Go per part; a three-case `claude plugin eval` of the test pack; whether the studio's half of check 7 has run. | The report against spec §15's list, by the verifier | 4-6 |
| **Total** | | | **30-47** |

**The partial `status --json` (part 2): every field.** It prints every field of `bonsai.status/1` in the contract's
order, `null` (or `[]` for a list) where the skeleton has not built it yet, as contract §2.2 asks of a writer. Its test
validates every built field against part 0's schema and holds the fields still `null` or `[]` in a named list: a listed
field that gets a value, or an unlisted one that is `null`, fails it. Step 5.1 empties the list. Chosen over checking only
the fields present: the key list and order are tested from the start, and a reader sees the same shape now and at 1.0.

**The scratch clone of the studio's repo** (part 3) is for scripted `bonsai` and git runs only. No Claude Code session
opens in it: it carries the studio's own hooks, which would write into the studio's real home and show phantom sessions
on its dashboard. No committed fixture names Mimas, the studio or any one project.

**CI and release after part 1.**
- The Linux job keeps the name `test` and its shape, not a matrix (a matrix renames the check `test (ubuntu-latest)`;
  the name stays in case Rohan turns the required check on again).
- A separate job, `windows`, on `windows-latest`, runs `go test ./...` and `go vet ./...` (public repos pay nothing for
  Windows minutes).
- `lint`, `govulncheck` and CodeQL stay as they are; all are expected green.
- `release.yml` keeps only `workflow_dispatch` (no tag trigger, no re-release input), `contents: read`, no
  `HOMEBREW_TAP_TOKEN`, and runs `goreleaser build --snapshot --clean`: it builds and publishes nothing, so
  `.goreleaser.yaml`'s `brews:` block (kept for 5.7's `bonsai@0.4`) is never reached. It stays disabled on GitHub until
  5.7, which is Rohan's.

**Test sessions and the launcher.**
- Scratch folders `~/bonsai-checks` and `%USERPROFILE%\bonsai-checks`. Scripted runs live in `~/bonsai-checks/scripts/`,
  never committed; their commands and output go in the run report.
- Each side has a `claude-here` launcher (`claude-here.cmd` on Windows) that, before it starts Claude Code:
  - puts `bonsai-checks/bin`, the scratch build, **first on the PATH**: until Rohan's step 3 the old 0.4.3 `bonsai` in
    `~/go/bin`, `~/.local/bin` and `%USERPROFILE%\go\bin` comes first, and every hook and timing would hit it;
  - sets one shared `CLAUDE_CODE_PLUGIN_CACHE_DIR` (spec §14; it moves the whole plugins root, so Rohan's own plugins
    show as missing in those sessions only).
- The launcher and the test build read `BONSAI_TEST_FAULT`; only the scratch test build has code for it (a Go build tag).
- No session opens in a scratch folder except through `claude-here`.
- Windows-side scratch repos, `project-a-worktree` and check 3's CRLF checkout are made with Windows git by its full
  path.
- The guard logs its own path and SHA-256; the gate report quotes them per side, so it shows which binary answered.

**Check 10, natively on Windows**, at the end of parts 2, 3 and 5 and on the final commit: `go test ./...` and
`go vet ./...` with Windows Go in `%USERPROFILE%\bonsai-checks\src`, a copy of the builder's tree (`cp -a` of the
worktree without `.git`) or a Windows-git clone of the pushed `main`. CI does not replace that run: check 10 needs both.

**The twelve checks** are spec §14's, with these readings: check 7 as "What still links" item 3 says (Bonsai's half is
"every trick file's format-1 outcome"; the skeleton has no format 0); check 3's CRLF checkout is made with Windows git;
check 12's revert runs on the scratch copies.

**Stop lines** (judged by a fresh verifier, not the builder). Done means **no line crossed without Rohan's recorded
choice, or, past a line, his recorded choice to go on.**
1. **Hours over 61** for parts 1-6 (part 0 has its own line, 13). The sum of every builder and verifier run's
   wall-clock, from the run reports' rows; a run without a row is added by hand and marked so.
2. **Windows-only failures over 8 hours.** A running tally in the run reports of time spent on failures that happen only
   on Windows. A failure also seen in WSL or on Linux CI does not count, nor does copying to a Windows folder or waiting
   for CI.
3. **More than two option rounds** asked of Rohan inside the skeleton, counted in the run reports. Hand checks are not
   option rounds.
4. **Any change to Mimas or to the studio's repo.** This work changes nothing outside Bonsai's repo, the test pack and
   the scratch folders. The verifier reads the studio's log read-only, and checks Mimas only with Windows git by its full
   path (the orchestrator gives the checkout), never Linux git.

At a crossed line work stops. Rohan gets the numbers and three choices (continue, the smaller cut, pause); his choice is
written in the run report before any more work.

**Rohan's hand checks** (spec §17 step 6). The builder leaves the folders; the orchestrator copies the spec's exact lines
into one message to him. He sends one line per check, and the run report keeps his words and the result. A check that
fails is a failed check, not a done one.

| Hand check | For | Passes when |
|---|---|---|
| a. The pack loads | Check 8 (part 4) | On WSL and in PowerShell, `/agents` shows the `marker` role saying "commit A" in `project-a`, "commit B" in `project-b`, "commit B" in `project-a-worktree` (he notes whether a trust question came), and "commit A" again in `project-a` |
| b. The PowerShell console | Check 9 (part 3) | `bonsai update` prints plain text with no broken characters and ends in y/N; `bonsai update --yes` stops on the conflict and prints a command; the pasted command works |
| c. The guard on Windows | Check 11, Windows half (part 5) | The `protected.txt` edit is refused with a reason, the `free.txt` edit is done, and each of the four fault sessions refuses the `free.txt` edit with a clear reason |
| d. Which `bonsai` Git Bash finds | Check 11 (part 5) | His `which -a bonsai` output is in the run report, and the guard's log for check c's sessions names the scratch build's path and hash, not the old `go\bin\bonsai.exe` |

### How it is proved

Bonsai has no ladder of its own until step 5.4, so the proof is the interim one (question C, spec §14): Go tests and
`go vet` in WSL and natively on Windows, CI, and a fresh verifier.

| Done when | Proved by |
|---|---|
| Part 0 | The formats tests (CI `test`); the verifier's rule-by-rule read and its own format-0 run; the Windows-git clone's `go test ./formats/`; the commit named in `STATE.md` |
| Part 1 | CI on the pushed commit, both jobs; `release.yml` build-only; the clear-out's file list |
| The test pack | `claude plugin validate --json`; its four commits |
| Part 2 | Committed Go tests on every trick file's format-1 outcome and on the schemas; check 10 |
| Part 3 | Checks 1-6 (5 to its session clause) and 12; check 9; check 10 |
| Part 5 | Check 11 on both sides; hand checks c and d; the fault tests; a normal build without fault code; check 10 |
| Part 4 | Check 8, hand check a, the recorded session kinds; check 5's last clause |
| Part 6 | The gate report against spec §15 |
| Stop lines | The run reports' rows and tallies, judged by the verifier |
| Hand checks | Rohan's lines against the pass conditions above |
| The interim proof | At the end of parts 2, 3 and 5 and on the final commit, `go test ./...` and `go vet ./...` in WSL and natively on Windows, output in the run report; the final commit's full SHA with its CI run, `test` and `windows` green; no Windows-only skip without a named reason |
| The last verifier | A fresh Opus verifier re-runs `go test ./...` and `go vet ./...` on both sides itself, and agrees on every part, the twelve checks and the four stop lines |
| Nothing private | The verifier's read of both repos' file lists and a grep of their files and commit messages for studio task ids, studio paths, home folders, machine or tailnet names and email addresses |

### Choices made, and what lost

| Choice | Alternative | Why this one |
|---|---|---|
| Records and design in Bonsai's repo | Records in the studio's repo under its task numbers | Rohan, 8 Oct |
| Work on `main`, each push after its proof | A `rebuild` branch, a draft pull request, a gate merge | Rohan, 8 Oct; force pushes and deletion stay blocked; 0.4.3 stays at its tag |
| The formats' master in Bonsai's `formats/`, part 0 first | The studio's master and a copy here | Rohan, 8 Oct |
| One `expect.json` for the set | One per case folder | One file any language reads, hashed in the manifest |
| Verifiers for part 0, parts 1-2, part 5 and the end | One per part, or one at the end | Rohan's 6 Oct rule: a fresh verifier only for big or risky work; the formats, the reader and the hook path are risky, and the end checks all twelve |
| `status --json` with every field | Only the fields present | Contract §2.2; the shape is tested from the start |
| The hook-line change in its own test-pack commit (D) | All of check 5's changes between A and B | Spec §6 says `update` is all or nothing and a hook-line change needs `--allow-exec` as well as `--yes`; kept apart, the skeleton needs no rule for a mixed update; that rule comes with `--allow-exec` in step 5.1 |
| Scripted runs over the scratch copies | Committed tests over them | They hold private content and absolute paths, and CI cannot reach them |
| A separate Windows CI job | A matrix on `test` | The check keeps the name `test` |
| A small schema checker of Bonsai's own | A schema library | Standard library only (spec §3) |

### Risk in the code

- **Trick files wrong.** One agent writes the cases and the answers; the verifier reads each against contract §2.4, not
  against any reader, and the studio's frozen reader re-checks the format-0 side.
- Go on Windows (paths, renames, file modes; job objects are not in the skeleton). Stop line 2.
- Plugins may not load pinned commits, or two commits side by side, as the spec expects. If no plugin delivery works at a
  pinned commit, stop and report: it decides the whole design (spec §5).
- The fault switch must not ship: the verifier checks the build tag and that a normal build has no fault symbol.
- A Dependabot pull request may appear; it is left alone. Never merge or close a pull request without Rohan.
- Processes: the orchestrator sweeps after each subagent, including `claude` test sessions and Go processes started from
  `~/Servers/Bonsai-<part>` or `~/bonsai-checks`.

### Stale or in tension in the spec

- §12 steps 2-6 and §17 steps 1, 4 and 5 are overtaken by Rohan's 8 Oct decisions; the spec carries a note at each.
- §19 D still says 60 h and 30-46 h; §14 and §20 say 61 h and 30-47 h. This plan uses 61 and 30-47.
- The 1.3-times re-ask rule is per step. This plan adds part 0's line (13 h) and no per-part lines inside the skeleton.
- §14 puts "a fetch at a 40-character commit" in part 4, but part 3's checks 1-2 need a fetch (the plugin wiring's
  commit, the lock's remote URL): part 3 builds the plain fetch, part 4 proves it read-only and without a login on both
  sides.
- Check 5 with a mixed update (`--yes` while a hook line also changes) is unsettled between §6's "all or nothing" and
  check 5's wording; the skeleton avoids it (commit D), and step 5.1 must settle it with `--allow-exec`.
- §17 step 6 lists hand check d under part 5, and check 11 does not name it; it is tied to check 11 here through the
  guard's log.
- §17 step 3 asks for the old binaries to go before step 5.4; this plan asks before part 5's hook checks. The launcher
  puts the scratch build first either way; a session started without it would find the old one.
