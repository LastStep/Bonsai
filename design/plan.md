# Bonsai's plan: the formats, then the walking skeleton

- **Status:** draft. Rewritten on 8 Oct 2026 for work inside Bonsai's repo, from the studio's drafts (the skeleton's
  plan and task, reviewed twice; the readers plan's step 1, reviewed once). Reviewed on 8 Oct by a fresh Opus agent and
  fixed on its findings; waits for Rohan's approval.
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
   refused for this reason. They are tests any reader can run. Bonsai's own reader uses them now; the studio's readers
   test against them when it links (step 7).
2. **The walking skeleton (parts 1-6).** The risky parts of the new Bonsai, built on throwaway copies of projects, to
   prove before you decide whether to go on: Bonsai can write and update a project's files without overwriting your
   edits; a pack can arrive as a Claude Code plugin at an exact version, with two projects (and a worktree beside its
   checkout) each loading their own; and the guard blocks what it should, and also blocks when it is missing, crashes,
   is slow or cannot be found. It ends with a report of measured numbers, and you pick the path: the full Bonsai 1.0,
   the smaller cut, or 1.0 with its own screens.

**What you will see.** Commits on Bonsai's `main`, each pushed after its proof passed (a fresh verifier where the work
is big or risky), with green checks on GitHub (Linux, and Windows from part 1). One check, `govulncheck`, has been red
since 23 Sep for a reason outside this work (its install now needs a newer Go than the checks use): part 0's commit
will still show it red, and part 1 makes it green. Part 1 clears the old product's code out of `main`; 0.4.3 stays
downloadable and at its tag. A new public repository, `LastStep/bonsai-test-pack`, for part 4. No release and no tag.
Your hand checks come in two sittings: about 25 minutes after part 5 and about 20 minutes after part 4. Then the gate
report.

**What you must do.**
- Approve this plan and the one-pager (`design/one-pager.md`). Nothing is built before.
- Confirm that the old Homebrew tap token is revoked on github.com (spec §17 step 4; not confirmed on 8 Oct).
- Before part 5's hook checks: your step 3, the old `bonsai` binaries (spec §17 step 3 has the lines).
- Two sittings of hand checks (spec §17 step 6; your choice, 8 Oct), together about 45 minutes, saving about 3 AI
  hours. The first, about 25 minutes after part 5: the PowerShell console, the guard on Windows, and which `bonsai`
  Git Bash finds (checks b, c, d). The second, about 20 minutes after part 4: the pack loads (check a). Parts 3 and 5
  are done only once the first sitting passes, and part 4 once the second does. Each check has a plain pass condition
  below; tell the orchestrator what you saw, one line per check.
- Later, not in this plan: step 8 (the pre-release at 5.4) and, at 5.7, the `release` environment and a new tap token.

**Choices that are yours.** None in this plan beyond approving it.

**Risk.** Windows is where the old code failed (29 tests); the stop line for that is 8 hours lost to Windows-only
failures. A Windows session may not load a pack the way WSL does; finding that out is what the skeleton is for, and the
answer goes in the gate report either way. Agents act on GitHub as your account, an admin, so the plan forbids them every
release, tag and settings change. Work lands on `main` with no pull request, as you chose: if you pause, `main` holds the
work so far and the old product stays at its tag. If any of the four stop lines is crossed, work stops and you get the
numbers and three choices: continue, the smaller cut, or pause.

**Size.** Part 0: 6-10 AI hours; at 13 work stops and you are asked. The skeleton: 30-47 AI hours across six parts; at
61 work stops and you are asked. On the studio's record (estimates grow 1.6 times) expect 10-16 and 48-75: both stop
lines fall inside those ranges, so you may well be asked to choose at one of them. Plus about 45 minutes of yours, in
the two sittings. Nothing here waits on the studio.

## For the orchestrator, builders and verifiers

### How the work runs

- **The orchestrator** is the Claude Code session Rohan opens in `~/Servers/Bonsai` (the main checkout, on `main`). It
  keeps its context lean, dispatches builders and verifiers with full briefs (Bonsai has no role files: each brief
  carries the rules of `CLAUDE.md` and of this plan that the job needs), reads their reports, merges and pushes.
- **Builders** work in a plain git worktree beside the clone, one per part:
  `git -C ~/Servers/Bonsai worktree add ~/Servers/Bonsai-<part> -b <part> main` (never the Agent tool's isolation
  worktrees); part 5's is made from part 3's branch, which waits unmerged for the first sitting (below). They commit on
  that branch and never push. Builds go `go build -o` into a scratch folder
  (`~/bonsai-checks/bin` for the skeleton); never `go install`. `~/ZenGarden/Bonsai` is never touched.
- **Verifiers** are fresh Opus agents, for big or risky work only (Rohan, 6 Oct: fewer verifications): part 0 (the set
  everyone tests against), part 1 (CI and release changes, before its push), part 2 (the reader), part 5 (the hook path,
  with part 3's hook-line writes and its refusal of a hook-line change), and the last verifier at the skeleton's end
  over all twelve checks. Parts 3, 4 and 6 land on green tests on both sides (check 10) and CI plus the orchestrator's
  read of the diff, which the run report says; the last verifier covers them. A verifier reads this plan's part, the
  spec and contract sections it cites and the diff, re-runs the tests itself, and passes or fails the part. It fixes
  nothing.
- **Landing a part.** After all its proof passes, hand checks included (Rohan, 8 Oct): parts 3 and 5 land together
  once the first sitting passes; the rest of part 4 starts from `main` once they have (or from part 5's branch while the
  sitting waits) and lands after the second sitting. Then: `git -C ~/Servers/Bonsai merge --ff-only <part>`, then
  `git -C ~/Servers/Bonsai push origin main`, then CI for that commit:
  `gh api repos/LastStep/Bonsai/commits/<sha>/check-runs --jq '.check_runs[] | [.name, .status, .conclusion] | @tsv'`
  (or `gh run list -R LastStep/Bonsai -L 10`; this machine's gh, 2.4, has no `--branch`). Red CI is fixed forward with a
  new commit; force pushes are blocked. Before each push the orchestrator fetches and reads `main`'s new commits on the
  remote: one it did not make (an old cloud session, a Dependabot merge) stops the work until it is understood. Every
  commit of this work is listed in its part's run report. The fast-forward needs `main` where the branch started: the
  orchestrator commits nothing on `main` while a part is out, and if `main` moved anyway, the builder rebases its branch
  on it and the tests run again before the merge.
- **GitHub.** Agents act as `LastStep`, Rohan's admin account. Only the orchestrator pushes: Bonsai's `main` and the
  test pack's `main`. Never: a tag, a release (`gh release` anything), a workflow switched on, a setting, ruleset or
  secret changed, a pull request merged or closed, a branch deleted on GitHub. Releases are Rohan's word at step 5.7.
- **Run reports**, one per part, `records/runs/R-<date>-<part>.md`, opened before the first edit and appended as the
  work goes: a log, not a summary. Each lists every builder and verifier run with its model, start, end and minutes, and
  the orchestrator keeps a running total for part 0 and for the skeleton. These rows are the hours source for the stop
  lines: the studio's session records do not cover Bonsai. Each run report has one writer, the orchestrator, in the main
  checkout: builders and verifiers write none and put their log in their final report, which the orchestrator copies
  in. It commits the report only after the part's fast-forward (an untracked report that the branch also held would stop
  the merge), then pushes.
- **Models.** Opus for builders on the formats, CI and release (part 1), the engine, the guard, Windows and plugin
  delivery (the test pack among it), and for every verifier and plan review; Sonnet for mechanical runs (cross-compiles,
  measurements into tables, CI checks); Haiku for small bookkeeping and audit jobs, as `CLAUDE.md` has it. Each run
  report names the model of every run.
- **Nothing private and nothing studio-shaped in what this work makes.** Code, tests, `formats/`, the test pack, commit
  messages and run reports name no studio task id or studio path, no one's home folder (`/home/...`, `C:\Users\...`;
  write `~/...` and `%USERPROFILE%`), no machine or tailnet name, no email address. `design/` holds the specs as the
  studio wrote them, citing its files; that is quoting, not new work. One exception: part 0 names format 0's source,
  the studio's `tools/lib/yaml.mjs` at commit `4a05eac` (contract §2.4 defines format 0 by that file), in its run report
  and `formats/README.md`.
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

The work waits for nothing in the studio. Rohan's 8 Oct decision, "Bonsai first": the studio's readers task (T-0074) is
cut and its Desk stays on upkeep until it links (step 7). What links the two:
1. **The studio adopts `formats/` when it links (step 7)**, not before: `bonsai init` there, its bridge reading Bonsai's
   outputs, its own guards, ladder and statusline retired. Its remaining readers test against the trick files then. A
   change to `formats/` is a commit that changes its manifest, never a silent edit.
2. **Format 0's outcomes come from the studio's reader.** Format 0 means "exactly what the studio's `yaml.mjs` reads
   today" (contract §2.3-§2.4). Part 0 fills each case's format-0 outcome by running that reader as the studio froze it
   for format 0, at commit `4a05eac`, taken out with `git show` into a scratch folder (part 0 says how). Nothing is
   written or run in the studio's checkout; the orchestrator gives its path. The studio's test of its frozen copy
   re-checks every format-0 outcome when it links.
3. **Check 7** (Bonsai's reader and the studio's reach the same outcome on every trick file) has two halves: Bonsai's
   reader reaches every format-1 outcome in `formats/expect.json` (part 2's test), and the studio's reader reaches the
   same file's outcomes (when it links, step 7; until then the half is open). The gate report says whether it has run.
4. **Later:** the studio's slice 3 and its registration of projects (contract §15.2) come after the skeleton; Bonsai
   joins the studio's dashboard as its own project at spec §14 step 6; the studio links to Bonsai at step 7, Mimas last
   at step 8.
5. **Note for step 5.3 (the guard).** Rohan's 8 Oct requirement: the hook lines must not be redirectable by files an agent may edit. They call the installed
   `bonsai` by a fixed path (never found through git or the project), and whatever the guard trusts to find the main
   checkout cannot be rewritten by an agent without the guard noticing (fail closed). Why: in the studio, hook commands find
   the guard scripts through `git rev-parse --git-common-dir`, and git takes that answer from files an agent may edit (a
   worktree's `.git` file, a `commondir` file inside a `.git` folder).
6. The studio's own work may ask Rohan questions in the same days (the spec's "two queues"). Stop line 3 counts only
   this plan's option rounds.

### Part 0: the formats (6-10 h, stop line 13 h)

Bonsai's first job (Rohan, 8 Oct: "those are kind of tests which other projects can use ... it makes sense for them to
live in bonsai"). It lands on `main` before part 1, beside the old product's code, in a folder part 1's clear-out keeps.
Sources: contract §2.4 (the YAML rules and the trick files), §2.2 and §2.5 (fields, versions, JSON), and each format's
section: §4 task, §5.2 labels, §6 lanes, §7.1 run, §7.2 state, §8.1 log, §9.1 ask, §11 ladder, §12 status, §14 lock;
with the sections they lean on: §3 (the home and the workspace id), §5.1 (label names), §8.2 (the log's events and
categories) and §13 (the active task). Spec §16 adds to four of these formats, confirmed by Rohan with the rest: row 18,
the lock's `declares` (spec §6); row 23, the ladder's third rung kind, git integrity (`ver-git`, spec §9); row 24,
`status --json`'s additions (spec §4-§7); row 27, the binary's path and SHA-256 in the log (spec §3, §8). A closed list
never grows inside a major (contract §2.2), so each schema carries them from the start. Where the contract gives a field
only "as today" (the ladder's `captures`, `skipped`, `leftovers` and `proof`; the log's `kind`, `text`, `source`,
`model` and `reason`; the ask's `verdict`), the builder reads today's shape read-only at the studio's frozen commit
(`git show 4a05eac:` of `tools/ladder/ladder.mjs`, `tools/lib/spool.mjs` and `tools/lib/asks.mjs`). A shape still
unclear is typed open; a name the spec has not fixed (the log's binary path and SHA-256, format review 4.2) is not
invented; `README.md` lists both.

**What it builds, all under `formats/`:**
- `README.md`: what the folder is and who reads it, how a case is laid out, every field of `expect.json` and
  `manifest.json` with an example, the reason codes, and how the set changes (a new case, a new manifest, a new commit
  named in `STATE.md`). The folder documents itself (contract §2.8).
- `schemas/<name>.schema.json` for the ten formats: `task`, `labels`, `lanes`, `run`, `state`, `log`, `ask`, `ladder`,
  `status`, `lock`. JSON Schema draft 2020-12. **A schema describes what a writer writes** (contract §2.2): properties
  in the contract's fixed order, every field the version knows required (`null` or `[]` where it does not apply),
  closed lists as `enum`, open lists as strings. A reader is more lenient (a missing field reads as `null`, an unknown
  one is kept), so `additionalProperties` stays open. Each schema documents itself (`CLAUDE.md`, contract §2.8): a
  top-level `description` (what the format is for, who writes and reads it) and a `description` and `examples` on every
  property. **These schemas are the one home** that contract §2.2 and format review R2.8 call "Bonsai's code" for every
  list in the ten formats: from part 2 on, Bonsai's Go code embeds them and keeps no second copy of a list. Not here:
  the two generated tables (`bonsai.tasks/1`, `bonsai.sessions/1`), `bonsai.yaml` (`bonsai.workspace/1`), `pack.yaml`
  (`bonsai.pack/1`) and the `error` object (`bonsai.error`, spec §3, §16 row 29), all step 5.1; memory
  (`bonsai.memory/1`, spec §10, §16 row 19), a spec format, not one of the contract's ten. Nothing here reads or writes
  a log or an ask.
- `examples/`: one valid document per format, made-up values, each stored as the JSON a reader returns
  (`<name>.json`), which its schema validates. For the five YAML and markdown formats (`task`, `labels`, `lanes`, `run`,
  `state`) the source file sits beside it (`<name>.md` or `<name>.yaml`) and is also a yaml-1 case: `expect.json` gives
  it both outcomes like any case, and its format-1 value is that JSON. So the test checks every schema without a
  reader, and part 2's reader proves the YAML.
- `trick/yaml-1/<rule>/` and `trick/yaml-0/<oddity>/`: the input files, markdown with frontmatter or plain YAML
  (definition) files, made-up content only. Format-1 cases start with a `format:` line, as a real file would. **One
  case per rule of contract §2.4:** LF and CRLF lines; a BOM (frontmatter, and a definition file); a tab in indentation;
  `---` and `...` anywhere but the frontmatter's own markers; every line consumed, a list at its key's own indent among
  them; the key rule: keys of `[a-z][a-z0-9_]*` and a dotted label key (`<namespace>.<name>`) accepted, `Title` and
  `done-when` refused, quoted keys, complex keys, `<<`, a key twice in one mapping, the reserved words as keys (`y`, `n`,
  `yes`, `no`, `on`, `off`, `true`, `false`, `null`); nested mappings; block sequences (deeper than their key, exactly
  one space after `-`); one-line flow sequences, `[]` and `{}`; anchors, aliases, tags, flow mappings with content,
  nested and multi-line flow sequences (all refused); block scalars `|`, `|-`, `>`, `>-`, and the rest refused (`|+`,
  `>+`, an indentation indicator such as `|2`), a line starting with `#` inside one, a deeper line in `>`; comments at
  a line's start and after a space, and a `#` with no space before it; a quote character inside a plain value; quoted scalars on one line, the five escapes, a bad escape, an unescaped `"`,
  something after the closing quote; **every row of the plain-scalar table** (null forms, `true` and `false`, the
  integer at 15 digits and at 16, a decimal, both date forms read as text, a text value, the "quote this value"
  refusal); **each listed refusal** (`0755`, `55227e5`, `0x1F`, `1e3`, `1_000`, `.inf`, `TRUE`, `No`, `5 arenas`,
  `.claude/**`, a value holding `: `), and each read as text once quoted. Plus: a duplicate key quoted and unquoted;
  **the format-0 oddities** (a plain value holding `: `, unquoted hashes, `TRUE`, the last of two duplicate keys,
  document markers skipped); and **the dispatch cases**: `format:` first, `format:` first carrying contract §2.8's
  pointer comment (accepted), no `format:` key, and `format:` not first (refused).
- `expect.json`: one language-neutral file giving every case two outcomes:
  - **Format 0:** the studio's frozen reader, which reads every file as format 0, `format:` line or not. `accepted` with
    the value as JSON, or `refused` with the reader's own message.
  - **Format 1:** a reader that dispatches as §2.4 says. With `format:` first it reads by the format-1 grammar:
    `accepted` with the value as JSON (a date stays text, an integer has at most 15 digits), or `refused` with a short
    reason code (the codes are listed in `README.md`). With no top-level `format:` key the outcome is `format-0`: §2.4
    sends the file to format 0, which part 2 does not build. With `format:` anywhere but first, `refused`.
  - **The value** is the YAML file's mapping, or for markdown the frontmatter's mapping (`parseFrontmatter`'s `data`);
    the body is not compared.
  - **A case §2.4's words do not settle** (for example `''` inside single quotes) takes the outcome §2.4's own test
    gives: accepted only if a YAML 1.1 and a YAML 1.2 library both read it and read the same value, otherwise refused.
    The builder lists every such case in its run report, and the verifier rules on each.
- `manifest.json`: a set version and the SHA-256 of each file's raw bytes (never line-ending-normalised: the CRLF case is
  the point), with forward-slash paths sorted by path, for every file under `formats/` except itself and the Go files.
- In the repo's `.gitattributes`: `formats/** -text`, on a line **after** the existing `*.yaml text eol=lf` (the last
  matching line wins), so no checkout, Windows' `core.autocrlf` included, changes a byte.
- `formats/*_test.go`: standard library only, a test-only package that needs no reader and survives part 1. It checks
  that:
  - the manifest matches every file's raw bytes, lists every file and nothing more, sorted by path;
  - the CRLF case holds `\r\n` and the BOM cases start with `EF BB BF`, as checked out;
  - every rule of §2.4 has a case (a short hand list of the rules in the test), and every case has both outcomes in
    `expect.json`;
  - every schema is valid JSON, declares draft 2020-12, and has a `description` and `examples` on every property; its
    example validates under a small checker in the test that implements the keywords the schemas use, skips only the
    annotations (`title`, `description`, `examples`) and fails on any other keyword (part 2 moves the checker into the
    new layout); each YAML or markdown example's format-1 value in `expect.json` equals its `<name>.json`;
  - no file in the set matches a generic private pattern: an absolute path (`/home/`, `/mnt/`, `~/`, a drive letter),
    an email address, a tailnet host (`.ts.net`). The test names no real machine or tailnet name, which would publish
    it; the verifier's grep checks those, with the names in its brief.

**The format-0 outcomes** (link 2 above): the studio's frozen reader, taken read-only with
`git -C <the studio's checkout> show 4a05eac:tools/lib/yaml.mjs > <scratch folder>/yaml0.mjs`. A small script beside it,
never committed, reads each case as the studio's callers do (`readFileSync(path, 'utf8')`, so a BOM and CRLF reach the
reader), runs `parseYaml` on a YAML file or `parseFrontmatter` on markdown, and writes the value (for markdown, its
`data`) or the exact error message into `expect.json`. The run report records the commit and how many cases it filled.

**Done when:**
- `go test ./...` and `go vet ./...` pass in WSL: the new tests pass, and the old product's tests are no worse than before
  (part 0 changes no product code).
- A fresh Opus verifier reads every case against contract §2.4 **rule by rule, not against any reader**: one agent wrote
  both the cases and the answers. Is each rule covered; is each format-1 outcome what the grammar says; is each format-0
  outcome what the studio's reader gives (it runs the frozen reader with a runner it writes itself from the paragraph
  above, not the builder's script). It also reads each schema against its sources and greps the set for private
  strings.
- Before the push, once, natively on Windows: a Windows-git clone of the builder's branch, made from a `git bundle` of
  it copied to `%USERPROFILE%\bonsai-checks\`, in `%USERPROFILE%\bonsai-checks\src`, where Windows Go passes
  `go test ./formats/` (only this package: the old product's own tests fail on Windows). It proves the CRLF and BOM bytes
  survive a Windows checkout. From part 1 the `windows` CI job runs it on every push.
- After the push: CI's `test` and `lint` jobs green on the commit. `govulncheck` stays red, as it has been since 23 Sep
  (its install needs a newer Go than CI's), until part 1 pins it.
- `STATE.md` names the commit, and the orchestrator tells Rohan in one line.

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
| 1. The clear-out and the new layout | One commit on `main` removes the old product's code: `internal/`, `cmd/`, `catalog/`, `website/`, `docs/`, `embed.go`, `.github/workflows/docs.yml` (the website's; the old site stays on GitHub Pages until Rohan turns Pages off). It keeps `LICENSE`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md` (rewritten for the rebuild), `CHANGELOG.md` (a new "rebuild" section), `.gitattributes`, `.gitignore` (less its old-product lines: the website, `station/`, the old lock name), `.golangci.yml`, `assets/`, the rest of `.github/`, `.goreleaser.yaml`, and all 8 Oct added: `CLAUDE.md`, `STATE.md`, `design/`, `records/`, `formats/`. It rewrites `README.md` ("being rebuilt; 0.4.3 is the old product's last release"), `Makefile` (no `go install`; `build` is `go build -o`), `go.mod` (`go 1.25` with a `toolchain` line; the standard library, and `golang.org/x/sys` only once needed) and `go.sum`, and adds `cmd/bonsai` with a stub `main` (`bonsai --version`), so `go vet` and CodeQL's autobuild find a package. CI and `release.yml` as below. | Before the push, a fresh Opus verifier: the clear-out's file list holds only what it should, `release.yml` can only build, CI as below, `go vet`. After it, CI green on the pushed commit (`test`, `windows`, `govulncheck`) | 2-4 |
| 4a. The test pack | `LastStep/bonsai-test-pack`, public: the orchestrator creates it with `gh` once Rohan approves this plan; an Opus builder makes its commits in a clone at `~/bonsai-checks/bonsai-test-pack`, and the orchestrator pushes them. Its `bonsai/pack.yaml` is `bonsai.pack/1` (spec §5). Commits: **A** (a `marker` role saying "commit A", the files `init` writes into a project, one hook line); **B** (marker "commit B", one of those files changed, one new file); **C** (that file changed again); **D** (one hook line changed, nothing else). | `claude plugin validate --json` with no warning but the missing `version`; part 2's reader reads its `pack.yaml` | in part 4 |
| 2. Reader and files | The format-1 reader in Go, no general YAML library (spec §3, contract §2.4), tested on the cases in this repo's own `formats/`; `bonsai.yaml`, `pack.yaml`, the lock (`.bonsai/lock.json`), a subset of `status --json` (below). No format 0 (step 5.1). The schema checker moves from part 0's test into the new layout. | Committed Go tests reach every trick file's format-1 outcome in `formats/expect.json`; the lock and `status --json` validate against part 0's schemas; the Windows rules (forward slashes, line-ending-blind hashes, byte-stable output) | 7-10 |
| 3. Engine | The fetch in its plain form (git, at a commit, from the pack's URL); `init`, `update` (staged; kinds `pack`, `once`, `block`, `keys`; `kept`; exits 4 and 5; `--diff`, `--yes`, `--keep`, `--adopt`), `check` (lock and files, a tracked or staged `.bonsai/local/` file); `bonsai.yaml` with a comment on every line; `.bonsai/lock.json` and `.bonsai/.gitignore`; `init`'s closing plain words; the preview naming every settings line with its sentence. A hook-line change only refused: exit 4, nothing written, `--allow-exec` named as the next step (the flag itself is step 5.1). Scratch targets: a drifted project (one old absolute hook line, one hook of its own), a fresh empty repo, and a scratch clone of the studio's repo with `origin` removed (the orchestrator gives its source), each with `init --new-id`. | Checks 1-6 and 12 as scripted runs over those targets (check 5 up to its new-session clause, which needs part 4), A to B for check 5, A to B to C for check 6, D for the hook-line refusal; committed Go tests of the same behaviour on `t.TempDir()` repos; check 9 (hand check b) | 10-15 |
| 5. Hook path | `bonsai hook guard` with one rule: an Edit of a path in `bonsai.yaml`'s protected list is refused (`project-guard` lists `protected.txt`). The fault switch `BONSAI_TEST_FAULT` (`missing`, `crash`, `slow`, `minimal-path`) behind a Go build tag, every fault only blocking. The binary's path and SHA-256 logged. | Check 11 on WSL (the builder, through `claude-here`) and on Windows under Git Bash (hand checks c and d); a Go test per fault; a normal build holds no fault code; its verifier also reads part 3's hook-line writes and the refusal of a hook-line change | 3-5 |
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
- `lint` and CodeQL stay as they are, expected green. `govulncheck` has been red since 23 Sep: its install
  (`go install golang.org/x/vuln/cmd/govulncheck@latest`) now needs a newer Go than CI's. Part 1 pins it to a version
  that builds with the Go CI uses, so it is green again from part 1.
- `release.yml` keeps only `workflow_dispatch` (no tag trigger, no re-release input), `contents: read`, no
  `HOMEBREW_TAP_TOKEN`, and runs `goreleaser build --snapshot --clean`: it builds and publishes nothing, so
  `.goreleaser.yaml`'s `brews:` block (kept for 5.7's `bonsai@0.4`) is never reached. It stays disabled on GitHub until
  5.7, which is Rohan's.

**Test sessions and the launcher.**
- Scratch folders `~/bonsai-checks` and `%USERPROFILE%\bonsai-checks`. Scripted runs live in `~/bonsai-checks/scripts/`,
  never committed; their commands and output go in the run report. Every script sets `BONSAI_HOME` to
  `bonsai-checks/home` on its side (contract §3), so `init`, `update` and the hook never write the real `~/.bonsai` or
  `%USERPROFILE%\.bonsai`.
- Each side has a `claude-here` launcher (`claude-here.cmd` on Windows) that, before it starts Claude Code:
  - puts `bonsai-checks/bin`, the scratch build, **first on the PATH**: until Rohan's step 3 the old 0.4.3 `bonsai` in
    `~/go/bin`, `~/.local/bin` and `%USERPROFILE%\go\bin` comes first, and every hook and timing would hit it;
  - sets one shared `CLAUDE_CODE_PLUGIN_CACHE_DIR` (spec §14; it moves the whole plugins root, so Rohan's own plugins
    show as missing in those sessions only);
  - sets `BONSAI_HOME` to `bonsai-checks/home`, as the scripts do.
- The launcher and the test build read `BONSAI_TEST_FAULT`; only the scratch test build has code for it (a Go build tag).
- No session opens in a scratch folder except through `claude-here`.
- Windows-side scratch repos, `project-a-worktree` and check 3's CRLF checkout are made with Windows git by its full
  path.
- The guard logs its own path and SHA-256; the gate report quotes them per side, so it shows which binary answered.
- **Claude Code's own files.** Every `claude plugin install` and `claude plugin marketplace add` the skeleton runs,
  Bonsai's own `update` included, passes `--scope project` or `--scope local`: the default, `user`, writes
  `~/.claude/settings.json`. The orchestrator records the SHA-256 of `~/.claude/settings.json` and
  `%USERPROFILE%\.claude\settings.json` before part 3 and after part 4; a change stops the work until it is understood.
  Accepted writes outside the scratch folders: Claude Code's session transcripts and its folder-trust entries for the
  scratch projects.

**Check 10, natively on Windows**, at the end of parts 2, 3, 4 and 5 and on the final commit, before the push:
`go test ./...` and `go vet ./...` with Windows Go in `%USERPROFILE%\bonsai-checks\src`, a copy of the builder's tree
(`cp -a` of the worktree without `.git`) or a Windows-git clone of the builder's branch (from a `git bundle`, as in
part 0). CI does not replace that run: check 10 needs both.

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

**Rohan's hand checks** (spec §17 step 6), in two sittings (Rohan, 8 Oct): the **first**, about 25 minutes after part
5, holds checks b, c and d; the **second**, about 20 minutes after part 4, holds check a. Parts 3 and 5 close after the
first sitting passes, not before; part 4 closes after the second. For each sitting the builder leaves the folders and
the orchestrator copies the spec's exact lines into one message to him, adding
`$env:BONSAI_HOME = "$env:USERPROFILE\bonsai-checks\home"` before check b's `bonsai update` (the launcher sets it for
the others). He sends one line per check, and the run report keeps his words and the result. A check that fails is a
failed check, not a done one.

| Hand check | For | Sitting | Passes when |
|---|---|---|---|
| a. The pack loads | Check 8 (part 4) | Second | On WSL and in PowerShell, `/agents` shows the `marker` role saying "commit A" in `project-a`, "commit B" in `project-b`, "commit B" in `project-a-worktree` (he notes whether a trust question came), and "commit A" again in `project-a` |
| b. The PowerShell console | Check 9 (part 3) | First | `bonsai update` prints plain text with no broken characters and ends in y/N; `bonsai update --yes` stops on the conflict and prints a command; the pasted command works |
| c. The guard on Windows | Check 11, Windows half (part 5) | First | The `protected.txt` edit is refused with a reason, the `free.txt` edit is done, and each of the four fault sessions refuses the `free.txt` edit with a clear reason |
| d. Which `bonsai` Git Bash finds | Check 11 (part 5) | First | His `which -a bonsai` output is in the run report, and the guard's log for check c's sessions names the scratch build's path and hash, not the old `go\bin\bonsai.exe` |

### How it is proved

Bonsai has no ladder of its own until step 5.4, so the proof is the interim one (question C, spec §14): Go tests and
`go vet` in WSL and natively on Windows, CI, and a fresh verifier.

| Done when | Proved by |
|---|---|
| Part 0 | The formats tests (CI `test` and `lint`); the verifier's rule-by-rule read and its own format-0 run; before the push, the Windows-git clone's `go test ./formats/`; the commit named in `STATE.md` |
| Part 1 | Before the push, a fresh verifier on the clear-out's file list, `release.yml` build-only and the CI jobs; CI on the pushed commit, `test`, `windows` and `govulncheck` green |
| The test pack | `claude plugin validate --json`; its four commits |
| Part 2 | Committed Go tests on every trick file's format-1 outcome and on the schemas; check 10 |
| Part 3 | Checks 1-6 (5 to its session clause) and 12; check 9 (hand check b, first sitting); check 10; part 5's verifier on its hook lines |
| Part 5 | Check 11 on both sides; hand checks c and d (first sitting); the fault tests; a normal build without fault code; check 10 |
| Part 4 | Check 8, hand check a (second sitting), the recorded session kinds; check 5's last clause; check 10 |
| Part 6 | The gate report against spec §15 |
| Stop lines | The run reports' rows and tallies, judged by the verifier |
| Hand checks | Rohan's lines against the pass conditions above |
| The interim proof | At the end of parts 2, 3, 4 and 5 and on the final commit, `go test ./...` and `go vet ./...` in WSL and natively on Windows, output in the run report; the final commit's full SHA with its CI run, `test` and `windows` green; no Windows-only skip without a named reason |
| The last verifier | A fresh Opus verifier re-runs `go test ./...` and `go vet ./...` on both sides itself, and agrees on every part, the twelve checks and the four stop lines |
| Nothing private | The verifier's read of both repos' file lists and a grep of their files and commit messages for studio task ids, studio paths, home folders, machine or tailnet names and email addresses |

### Choices made, and what lost

| Choice | Alternative | Why this one |
|---|---|---|
| Records and design in Bonsai's repo | Records in the studio's repo under its task numbers | Rohan, 8 Oct |
| Work on `main`, each push after its proof | A `rebuild` branch, a draft pull request, a gate merge | Rohan, 8 Oct; force pushes and deletion stay blocked; 0.4.3 stays at its tag |
| Hand checks in two sittings: b, c and d after part 5, a after part 4 | One 45-minute sitting after part 4 | Rohan, 8 Oct; parts 3 and 5 close on their hand checks without waiting for part 4 |
| The formats' master in Bonsai's `formats/`, part 0 first | The studio's master and a copy here | Rohan, 8 Oct |
| One `expect.json` for the set | One per case folder | One file any language reads, hashed in the manifest |
| Verifiers for part 0, part 1, part 2, part 5 (with part 3's hook lines) and the end | One per part, or one at the end | Rohan's 6 Oct rule: a fresh verifier only for big or risky work; the formats, CI and release, the reader and the hooks are risky (`CLAUDE.md`), and the end checks all twelve |
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
- §18 and §19 D give question D's own figures (60 h, 30-46 h) and then the format review's (61 h, 30-47 h), which
  §14 and §20 use. This plan uses 61 and 30-47.
- The 1.3-times re-ask rule is per step. This plan adds part 0's line (13 h) and no per-part lines inside the skeleton.
- §14 puts "a fetch at a 40-character commit" in part 4, but part 3's checks 1-2 need a fetch (the plugin wiring's
  commit, the lock's remote URL): part 3 builds the plain fetch, part 4 proves it read-only and without a login on both
  sides.
- Check 5 with a mixed update (`--yes` while a hook line also changes) is unsettled between §6's "all or nothing" and
  check 5's wording; the skeleton avoids it (commit D), and step 5.1 must settle it with `--allow-exec`.
- §17 step 6 has the hand checks in one 45-minute sitting; Rohan chose two on 8 Oct (above).
- §17 step 6 lists hand check d under part 5, and check 11 does not name it; it is tied to check 11 here through the
  guard's log.
- §17 step 3 asks for the old binaries to go before step 5.4; this plan asks before part 5's hook checks. The launcher
  puts the scratch build first either way; a session started without it would find the old one.
