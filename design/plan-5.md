# Bonsai's plan for step 5: path (a), Bonsai 1.0, with 5.1 in full

- **Status:** draft, 9 Oct 2026. For review by a fresh Opus agent, fixed on its findings, then Rohan's approval.
  Nothing in step 5 is built before he approves it.
- **What it follows:** Rohan's gate, 9 Oct: **path (a), the full Bonsai 1.0** (spec §14, "Path (a) after the gate
  (step 5), 139-218 h"), parts 5.1 to 5.7 in the spec's order.
- **Design:** `design/bonsai-spec.md` (the spec: §14 the parts, their hours and re-ask lines, "How Bonsai's work is
  proven"; §17 Rohan's steps), `design/contract.md` (the formats), `design/one-pager.md`, `design/format-review.md`.
- **Builds on:** `design/plan.md` (part 0 and the walking skeleton, approved 8 Oct, done) and
  `records/gate-skeleton.md` (the gate report: section 1 the measured pace, section 5 the findings step 5 inherits).
- **Hours** (AI hours, verification included, spec §14): 5.1 30-47, 5.2 25-37, 5.3 18-29, 5.4 28-44, 5.5 19-30,
  5.6 13-20, 5.7 6-11; in all 139-218. Each part stops and comes back to Rohan at its own re-ask line (below).
- **Records:** `STATE.md`; run reports in `records/runs/`.

Two readers. **Rohan** reads down to "Size" and reads no code. The **orchestrator, builders and verifiers** read the
rest. 5.1 is planned in full here; 5.2 to 5.7 are outlined, and each gets its own detailed section in this file before
it starts.

## For Rohan (plain words)

**What step 5 builds.** Bonsai 1.0, in seven parts, one after the other. What each lets you do or see:

| Part | What you will be able to do or see when it is done | Spec hours | Re-ask at |
|---|---|---|---|
| 5.1 Formats and engine to 1.0 | Bonsai reads and writes every format, old files included. An update that would change code that runs (a hook) waits for a second, separate yes. `bonsai unlink` takes Bonsai out of a project cleanly. `bonsai check` reports every problem the spec lists, `bonsai status` is complete, and one page lists every list of allowed values. | 30-47 | 61 |
| 5.2 Recorder, logs, asks | Every session in a linked project leaves a log inside the project, with secrets hidden as it is written. Agents can ask you typed questions and read your answers. Old records are cleaned by rules you set. A table of sessions and hours per task. | 25-37 | 48 |
| 5.3 Guards | The full guard: an agent may change a protected file only while its running task allows it; files only you may grant stay yours; a recursive delete that does not name what it deletes is refused; a builder cannot stop before its proof is green. You run one real Windows session (about 10 minutes). | 18-29 | 38 |
| 5.4 Ladder runner | `bonsai ladder` proves a task's work, on WSL and Windows. From here Bonsai proves and guards its own repo with it, on a pre-release you install in WSL (about 5 minutes, your password). | 28-44 | 57 |
| 5.5 Packs | Bonsai's `base` pack and your `workflow` pack (your roles, lanes, protocols and templates) as public Claude Code plugins, each with its own checks on GitHub; a template for new packs; the walls round your key and token files. | 19-30 | 39 |
| 5.6 Machine pieces | Settings per machine, label files the studio attaches, your personal memory layer, the workspace half of the statusline, and installers for both sides. | 13-20 | 26 |
| 5.7 Release | The release path made safe and switched back on with you; on your word, Bonsai 1.0. | 6-11 | 14 |
| **Step 5** | | **139-218** | each part its own |

**No screens of Bonsai's own** (your word, 9 Oct). Bonsai is a command-line tool only: its human output is each
command's plain text, and `--json` for programs. Every picture of its data is the studio's, drawn from those JSON
outputs (`status --json` and the rest). One consequence: an update's diff is free text, which the formats keep off the
VPS (contract §2.6), so you read it in the terminal (`bonsai update --diff`) unless the studio decides otherwise later.

**The order, and why.** The spec's order, which is also the order things depend on each other: the formats and the
engine first (everything reads them); then the log and asks (the guard and the ladder write into them); then the
guards; then the ladder (it uses the guard's rules for its first rung and writes what the stop gate reads); then the
packs (they are checked by 5.1's pack check and tried against 5.3's guard, and Bonsai links its own `base` pack after
5.4); then the machine pieces (the statusline's workspace half may show what the earlier parts record); then the
release. Your 8 Oct rule holds: "the proper way, no shortcuts". No two parts run at once. Inside 5.1 two small pieces
may run beside another (named below), because they share no file and no proof; each later part's section names any
such pair.

**What you will see.**
- Commits on Bonsai's `main`, each pushed only after its proof passed, with green checks on GitHub (Linux and
  Windows). A fresh Opus verifier for the risky pieces and at the end of every part.
- A run report per piece of work in `records/runs/`, and `STATE.md` rewritten at the end of each part.
- Your roadmap's Bonsai cards updated after each part (your 8 Oct decision).
- New commits on the public test pack (`LastStep/bonsai-test-pack`) in 5.1, added after its four, never rewriting them.
  A new public repo for your `workflow` pack in 5.5 (the spec names `LastStep/bonsai-workflow`), once you approve
  5.5's section.
- No release and no tag until you say so at 5.7.

**What you must do, and when.**
- **Now:** approve this plan, and answer the one choice below. Nothing in step 5 is built before.
- **In 5.3:** the second Windows check (spec §17 step 7, about 10 minutes): a real Windows session in a scratch folder,
  asking Claude for three edits and reporting what happened. A Sonnet agent runs every check an agent can first, so
  your sitting is only what needs a person typing (your 8 Oct word).
- **At 5.4:** install the pre-release `bonsai` in WSL (spec §17 step 8, about 5 minutes, your password). The
  orchestrator gives you its fingerprint first. Your step 3 (the old binaries) is already done (8 Oct). You install
  again only if a later part changes the guard, the stop gate or the ladder.
- **In 5.5:** possibly a repository secret, if a pack's checks on GitHub need a Claude login or a model key; 5.5's
  section says, with exact lines.
- **At 5.7:** the GitHub release steps that are yours (the `release` environment and a new Homebrew tap token, spec §17
  step 4's note; switching the release workflow back on), then your word for 1.0.
- **Whenever you like:** turn GitHub Pages off (the old website).
- Hand checks: a Sonnet agent runs them wherever an agent can (your 8 Oct word). Yours are only what needs a person: a
  typed session, a password, a GitHub setting.

**When the studio links, and when Bonsai joins the Desk.**
- **Bonsai joins the Desk** (spec step 6): from 5.4 Bonsai is linked to itself (its own `bonsai.yaml`, its own
  ladder and guard), so the studio can register it as its own project. The registration is the studio's own work
  (contract §15.2), not in this plan or its hours; Bonsai's side is ready at 5.4.
- **The studio links** (spec step 7) after 5.5: its own plan, in its own repo. Mimas links last (step 8). This plan
  changes nothing in either repo, and waits on neither.

**Choices that are yours.**
1. **How the later parts' plans reach you.** 5.1 is planned in full here. Before each later part starts, a planner
   writes its detailed section into this file and a fresh Opus agent reviews it; the orchestrator fixes what the
   review finds. Then:
   - **(A)** every part's section comes to you for approval before the part starts (six more approvals);
   - **(B)** a part's section comes to you only when it changes what is yours in this plan: its hours or re-ask line,
     the order of the parts, your steps, or a choice that is yours (a new public repo, a format change, a question
     only you can answer). Otherwise the part starts after the review's fixes, and you get one line with a link;
   - **(C)** no section comes to you; you hear at the end of each part.

   **Recommended: (B).** Each section is still reviewed fresh, and nothing that is yours moves without you. 5.4 (your
   install), 5.5 (your roles go public; a new repo) and 5.7 (the release) will come to you under (B) anyway, because
   each holds a step or a choice of yours.

**Risk.**
- **Windows.** The guards and the ladder (5.3, 5.4) are where Windows bites: paths, junctions, process trees. The
  skeleton lost about 18 minutes to Windows-only failures. Work stops and you are asked if step 5 loses 8 hours to them.
- **Claude Code changes under us.** In the skeleton, `/agents` disappeared in Claude Code 2.1.294, and a project's
  plugins install only after a person trusts its folder. Each part records the Claude Code version it ran on.
- **From 5.4, Bonsai guards its own repo.** A broken pre-release could block edits in Bonsai's own checkout, because
  the guard blocks when it fails. 5.4's section writes your way back (one or two lines) before the switch.
- **Your roles and protocols go public in 5.5.** 5.5 checks every file for anything private before the first push.
- **Agents act on GitHub as your account.** They never tag, release, switch a workflow on, or change a setting, secret
  or ruleset; those stay yours (5.7).
- **Stop lines.** Work stops and you get the numbers and three choices (continue, the smaller cut, or pause) if: a
  part reaches its re-ask line; step 5 loses 8 hours to Windows-only failures; more than two option rounds are asked of
  you inside one part; or anything changes in Mimas or the studio's repo. Your choice is written down before any more
  work.

**Size.** The spec's hours: 139-218 AI hours for step 5 (222-349 on the studio's record of estimates growing 1.6
times), each part with its own re-ask line (the table above). The skeleton's measured pace: 6 h 17 min of agent runs
against its 30-47 h, a ratio of 0.1337-0.2094 (gate report section 2.2). At that ratio step 5 would take 18.6-45.6 h.
**That is not a promise**, for four reasons (gate report section 1): it counts only the agents' own runs, not the
orchestrating session, your sittings or waiting; it comes from six parts over two days; it swung by part from
0.06-0.09 (the engine) to 0.32-0.53 (the hook path), and the parts that dealt with Windows and Claude Code ran highest,
which is more of step 5; and the stop lines use the spec's hours, not the ratio. Your own time: about 10 minutes at
5.3 and 5 at 5.4, the 5.7 GitHub steps, and the plan approvals your choice above sets. Nothing here waits on the
studio.

## For the orchestrator, builders and verifiers

### How the work runs in step 5

**What holds from `design/plan.md`** ("How the work runs", "Test sessions and the launcher", "Check 10, natively on
Windows", "The scratch clone of the studio's repo"), unchanged unless listed below: the orchestrator in the main
checkout dispatching with full briefs (Bonsai has no role files); builders in plain worktrees, committing, never
pushing; builds with `go build -o` into a scratch folder, never `go install`; landing by `merge --ff-only`, push, CI
read for the pushed commit, red CI fixed forward, the remote's new commits read before each push; GitHub as
`LastStep` with every tag, release, workflow switch, setting, ruleset, secret and pull-request action forbidden; run
reports written only by the orchestrator, opened before the first edit, a log, committed after the fast-forward;
nothing private in code, tests, commits or reports; tests in `t.TempDir()` with their own `BONSAI_HOME`; every
process an agent starts stopped and checked with `ps -eo pid,etime,cmd`; Windows runs under
`%USERPROFILE%\bonsai-checks\` with Windows Go and Windows git by full path; scripted runs in
`~/bonsai-checks/scripts/`, never committed, with `BONSAI_HOME` on the scratch home; test sessions only through
`claude-here`; every `claude plugin` command at project scope; the user settings files' SHA-256 recorded before and
after any part that runs Claude Code, a change stopping the work; no Claude Code session in a scratch clone of the
studio's repo; check 10 (`go test ./...` and `go vet ./...`, plain and with the fault tag, in WSL and natively on
Windows) before every push. Accepted writes outside the scratch folders stay as `plan.md` has them, plus Claude.ai's
plugin sync rewriting its own files under `~/.claude/plugins/synced/` (part 4b; gate report section 5, 5.5).

**What changes for step 5.**
1. **A detailed section per part, written before it starts.** A planner (Opus) writes the part's section into this
   file from its outline below, the spec and contract sections it cites and the state of `main`; a fresh Opus agent
   reviews it; the orchestrator fixes it on the findings; then it goes to Rohan or not, by his choice above. The
   section has 5.1's shape: pieces with their proof, hours and sources, the order, where each inherited finding is
   settled, and "done" in checks a verifier can run.
2. **Pieces.** A part is built in pieces (5.1's are below), one worktree and branch each:
   `git -C ~/Servers/Bonsai worktree add ~/Servers/Bonsai-<piece> -b <piece> main`, with `<piece>` the piece's number
   (`5.1.1`). Pieces land one at a time; a piece allowed to run beside another rebases on `main` and re-runs its proof
   if the other lands first.
3. **A run report per piece**, `records/runs/R-<date>-<piece>-<topic>.md` (`R-2026-10-10-5.1.1-consent.md`), with its
   "Runs" rows (model, start, end, minutes), the part's running total against its re-ask line, and step 5's
   Windows-only tally. A part's planning and plan-review runs go in the report of its first piece and count in its
   hours; this plan's writing and review count in 5.1's.
4. **Verifiers** (`CLAUDE.md`, Rohan's rule): a fresh Opus verifier for big or risky work (the reader, guards and
   hooks, CI and release, security) and at the end of every part. Each part's section names its risky pieces; 5.1's
   are below. The end-of-part verifier re-runs the tests on both sides itself, runs the part's "done" checks, reads the
   stop lines and passes or fails the part. It fixes nothing. That also meets spec §14's "a fresh verifier on every
   step", at the part's level.
5. **The proof changes at 5.4** (spec §14, "How Bonsai's work is proven"; question C, §19). Through 5.3: the interim
   proof, as in the skeleton. From 5.4: Bonsai's own `bonsai.yaml` in its repo, its tasks climbing `bonsai ladder` run
   by the pre-release Rohan installs in WSL (§17 step 8), and Bonsai's guard over its own checkout. CI on both sides
   and check 10 stay until a rung covers what they prove; fresh verifiers stay for the big steps (spec §14). 5.4's
   section sets the rungs, the task files Bonsai's own work then uses, how a builder's task is named, and the way back
   if the pre-release blocks work.
6. **The test pack grows by appended commits.** `LastStep/bonsai-test-pack` keeps A to D as they are; 5.1 adds E, F
   and G (below), each made by the piece's builder in `~/bonsai-checks/bonsai-test-pack` with its README updated in
   the same commit, and pushed by the orchestrator after the piece's proof passes. `internal/testpack` mirrors them.
7. **Every build whose hash is logged or installed carries its commit.** Builds made inside the WSL worktrees get no
   `vcs.revision` (gate report 2.11). The 5.4 pre-release and any binary whose hash a report quotes are built from a
   clean clone of the commit, and `go version -m` shows the stamp (5.2 settles the rest of that finding).
8. **Hand checks:** a Sonnet agent runs every check an agent can, on Rohan's 8 Oct word; Rohan's sittings hold only
   what needs a person. The orchestrator sends his lines in one batch with exact commands (bash for WSL, PowerShell 5.1
   for Windows, one command per line), and the run report keeps his words.
9. **`STATE.md`** is rewritten at the end of each part and whenever where Bonsai stands changes; Rohan's roadmap
   artifact is updated after each part.
10. **The studio's frozen files, read-only.** Where a piece needs a studio file as its reference (format 0's
    `tools/lib/yaml.mjs` at `4a05eac`, contract §2.4; 5.2's redaction tests; 5.5's roles and protocols), it is taken
    with `git show <commit>:<path>` into a scratch folder, or read from a scratch clone with its origin removed; the
    orchestrator gives the studio checkout's path. Nothing is written or run in the studio's checkout. Mimas is read
    only with Windows git by its full path. **One exception to "nothing studio-shaped"**: the format-0 port's comments
    and its run report may name `yaml.mjs` at `4a05eac`, the contract's own definition of format 0 (contract §2.4); the
    skeleton's last verifier noted the reader already does (`STATE.md`, loose ends).

### Step 5.1: formats and engine to 1.0 (30-47 h, re-ask at 61)

The spec's row (§14): "Every contract format in Go and the format-0 hand port; the schema-compare rung (contract §2.2);
layers, `--allow-exec`, `unlink`, the old-workspace refusal, all of §6's findings and warnings, `--schema`, `status
--full` (26-40); the tasks table and `check --write`, with its stale rules (2-3); the reference page of lists,
generated, and its CI test (1-2); `check --pack`, the template docs kept in step with their fields (1-2)". It also
settles every 5.1 finding of the gate report's section 5.

**What exists** (the skeleton's code, at `54b4fdd`; read each package's doc comment):
- `internal/reader`: the format-1 reader, all 116 format-1 outcomes; files with no top-level `format:` come back as
  "format 0", unread.
- `internal/schema`: ordered JSON, a strict decoder (duplicate keys refused), byte-stable ASCII encoding, the schema
  checker.
- `internal/workspace`: `bonsai.yaml` and `pack.yaml` (the fields the skeleton used; every other key kept as read),
  the lock, the home and machine folder, checkouts, atomic writes with Windows' busy retry.
- `internal/engine`: the fetch, `init`, `update` (staged, the four kinds plus `kept`, exits 4 and 5, `--diff`,
  `--yes`, `--keep`, `--adopt`), the preview's settings lines, the block, `check` (lock and files, `.gitignore`,
  tracked `local/` files, plugin drift), the plugin install step (`installed`, `waiting`, `failed`, `skipped`). The
  0.4.3-workspace refusal and the "two packs write one path" refusal are already in `plan.go`.
- `internal/status`: every field of `bonsai.status/1`, nine of them held `null` or `[]` in the test's `notBuiltYet`.
- `internal/guard`: the one-rule guard, its fault switch behind the `bonsai_test_fault` tag.
- `formats/`: set 3, ten schemas, 116 trick cases, the manifest; `formats/embed.go` embeds the schemas.
- `cmd/bonsai`: `--version`, `--help`, `status`, `init`, `update`, `check`, `hook guard`; `--allow-exec` exits 2.

**The pieces, in order.** Hours: the spec gives 26-40 for its first item as one figure; the split across pieces 5.1.1
to 5.1.7 below is the planner's judgment for sizing briefs, not a spec figure. It sums to the spec's figure (low
3+3+4+5+4+4+3 = 26; high 5+5+6+8+6+6+4 = 40). Pieces 5.1.8 to 5.1.10 carry the spec's own figures. Piece 5.1.0 is not
in the spec's row (below).

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.1.0 | **Tests that rest on a fixed time.** `TestEachFaultBlocks` (`internal/guard/fault_on_test.go`, a 200 ms budget under the fault tag, which CI runs on both sides) and `TestOverTimeBlocks` (`internal/guard/guard_test.go`, about 1.15 s, and the over-time record within 500 ms) made to wait on the event they test, not on a fixed time, as `1170c92` did for `TestRenameRetryOnWindows`. Test files only | The same tests pass plain and tagged on both sides; the diff touches only `_test.go` files (the orchestrator's read) | not in the spec's rows; its minutes count in 5.1's | The coordinator's note of 9 Oct (below) |
| 5.1.1 | **Consent to code.** `--allow-exec`; the mixed update; a pack's hook lines, the files they run and a plugin's own code parts counted as running code; the relink with the lock deleted. Test-pack commits E and F | Go tests walking a table of every consent case on the test pack's A to F; check 5 re-run as a script with E and F; the verifier's break-it over flag combinations | 3-5 | Spec §4 (`init`, `update`), §5 (what a plugin carries; mods), §6 ("How `update` decides", "Code is consented to separately", the preview), §7 (the hook lines), §14 check 5; gate report §3 row 5, §5 (5.1); `plan.md` "Stale or in tension" |
| 5.1.2 | **The format-0 reader**: a hand port of `yaml.mjs` at `4a05eac`, reached by the reader's dispatch for a file with no top-level `format:` | Every case's format-0 outcome in `formats/expect.json`; a fuzz run with no crash; a scripted run over today's files against the frozen reader with no difference | 3-5 | Contract §2.3, §2.4; `formats/README.md` ("expect.json", "How the outcomes were found"); spec §3 |
| 5.1.3 | **Formats set 4**: the schemas `workspace`, `pack`, `tasks`, `sessions`, `memory`, `error`, each documented, with an example; the two cases the reader's verifier left for a later set; contract §13's active-task fixtures with their answers; the manifest at set 4; the schema-compare test | The formats test (manifest, docs, examples, the rules list); the schema-compare test shown failing on a removal and passing on an addition; a Windows-git clone's `go test ./formats/`; the verifier's rule-by-rule read; CI | 4-6 | Contract §2.2, §2.4, §2.8, §3, §5.2, §7.3, §7.4, §7.5, §13; spec §3 and §16 row 29 (`error`), §5 (`pack.yaml`), §6 (`bonsai.yaml`, generated kinds), §10 (memory); format review 1.4, 3.5, 4.5, 6.1, R2.5, R2.6; `formats/README.md` ("How the set changes") |
| 5.1.4 | **Every format in Go**: a reader (and a writer where Bonsai writes it) for each of the sixteen formats, held to its schema; `bonsai.yaml` and `pack.yaml` read in full; the `error` object in every command's `--json`; `check --schema F`; every word's `--help` from one table of flags and exit codes | Go tests: each example read and written back byte for byte in schema order, and validated; `--schema` for all sixteen; every refusal's `--json` carries an `error` whose code is in the schema's list | 5-8 | Contract §2.1, §2.2, §2.5, each format's section (§4-§9, §11, §12, §14); spec §3 (output, unattended, `error`), §4, §5 (`pack.yaml`), §6 (`bonsai.yaml`), §16 rows 16, 17, 29 |
| 5.1.5 | **What the engine reads and writes, to 1.0**: the lock's `declares`; declared document kinds; labels in force; the active task (contract §13), one function every reader uses; this machine's settings, read; the instruction block complete; `bonsai.yaml` written with every field | Go tests on §13's fixtures (set 4) and on `t.TempDir()` projects; `check` running offline from the lock alone (no pack cache) | 4-6 | Spec §5 (pinning, "CI needs no pack"), §6 (`bonsai.yaml`, the lock, the block, layers), §10 (the index import); contract §3, §5.1-§5.3, §6, §7.3, §13, §14 |
| 5.1.6 | **`check` and `status` to 1.0**: every finding and warning of spec §6 (two moved, below); the Claude Code floor at 2.1.294; `status --json` with every field, `--full` and `--active` | One Go test per finding and warning, walked from one table so a finding with no test fails; the status test's `notBuiltYet` list empty | 4-6 | Spec §3 (the tripwire, `install.json`), §6 ("`bonsai check` findings", "Warnings"), §7 (the version floor), §10 (budgets); contract §2.3, §2.6, §3, §6, §12, §13; gate report §2.10 |
| 5.1.7 | **`unlink`, and Claude Code's own records**: `bonsai unlink`; the project-scope install record removed with it; first-time trust's `waiting` finished; Claude Code's key order kept in `.claude/settings.json`; project scope only, held by a test | Go tests; scripted runs on both sides on installed scratch targets (unlink, check 12 re-run with an installed plugin, the key-order run) | 3-4 | Spec §4 (`unlink`), §5 ("This machine's install follows the lock", "Trust"), §7 (no `bonsai.yaml`: the hook exits 0); gate report §3 rows 4 and 12, §5 (5.1); `records/runs/R-2026-10-08-plugins.md` 21:14 |
| 5.1.8 | **The tasks table and `check --write`**, with its stale rule | Go tests: the table byte-stable from format-0 and format-1 tasks; the worktree refusal; the exit codes; stale is a warning only | 2-3 | Spec §6 ("The two tables"); contract §7.5, §13 |
| 5.1.9 | **`check --pack`**: a pack folder's declarations and its documentation kept in step with its fields. Test-pack commit G | Fixtures failing each rule; the test pack at G passing | 1-2 | Spec §5 ("Every template and pack file documents itself"), §3 (no `bash` by name); contract §2.8 |
| 5.1.10 | **The reference page of lists**, `docs/reference/lists.md`, generated with `go generate`, and its test | The test rebuilds the page and fails on any difference (shown with a changed list in a temporary copy) | 1-2 | Spec §6 ("Every list has one home"), §12; contract §2.2 |
| **5.1** | | | **30-47** (re-ask 61) | |

**The order, and why.**
1. **5.1.0 first, or beside 5.1.1.** Every 5.1 push runs these tests on a loaded Windows runner; one flaked there
   already (`1170c92`). It touches test files only, in packages 5.1.1 does not change: truly independent.
2. **5.1.1 next** (the gate report and `STATE.md`: "`--allow-exec` and the mixed update first"). It settles how
   `init` and `update` judge code, which the relink, `unlink` and every preview reuse, before more code builds on the
   skeleton's "refuse only".
3. **5.1.2 may run beside 5.1.1.** It adds a mode to `internal/reader`, which 5.1.1 does not touch, and its proof
   (`expect.json`'s format-0 outcomes) does not rest on 5.1.1's. If 5.1.3 lands first, 5.1.2 rebases and its test
   picks up set 4's two new cases by itself.
4. **5.1.3 after 5.1.1:** the `pack` schema carries 5.1.1's field for the files a hook runs. Everything after reads
   set 4's schemas, the one home of every list.
5. **5.1.4 after 5.1.2 and 5.1.3:** its Go types are held to set 4's schemas; Bonsai-kind files read under format 0
   or 1.
6. **5.1.5 after 5.1.4:** the active task reads typed task files; `declares` writes typed lanes, kinds and labels.
7. **5.1.6 after 5.1.5:** the findings need `declares` (offline), lanes (`approve_first`), labels and the active task.
8. **5.1.7 after 5.1.6:** `unlink` reuses 5.1.1's preview and 5.1.4's `error`; it changes the engine's apply and
   plugin code, which 5.1.6 also touches, so they do not run together.
9. **5.1.8 after 5.1.6:** the table's header is the active task; its stale warning is one of `check`'s.
10. **5.1.9 after 5.1.4:** it reads `pack.yaml` through 5.1.4's types; placed after 5.1.8 so 5.1.10 can follow.
11. **5.1.10 last:** it lists every list, so it is generated once all of 5.1's lists exist; from then on its test
    guards every later part.

**Who builds and verifies.** Opus builders for 5.1.0 to 5.1.7 and 5.1.9 (the formats, a reader, consent to code,
the engine, Claude Code's files); Sonnet for 5.1.8 and 5.1.10 (mechanical, with the orchestrator's read). Fresh Opus
verifiers: **5.1.1** (consent to code is security: what runs on a person's machine), **5.1.2** (a reader), **5.1.3**
(the set every reader tests against, and a CI change). Pieces 5.1.0 and 5.1.4 to 5.1.10 land on green tests on both
sides, CI and the orchestrator's read of the diff, which the run report says; **the 5.1 end verifier** covers them.

#### Where each inherited finding is settled

The gate report's section 5, its 5.1 list, item by item, plus one loose end from `STATE.md` and the coordinator's
note of 9 Oct:

| Finding (gate report §5, 5.1) | Settled in | How |
|---|---|---|
| `--allow-exec`; the mixed update between §6's "all or nothing" and check 5's words | 5.1.1 | Rules 1-6 below: a run with any code to consent writes nothing without `--allow-exec` |
| A plugin's move to a new commit is not counted as running code | 5.1.1 | Rule 3 below: the plugin's own code parts are compared between the two commits |
| With the lock deleted, `init --yes` links again and writes a changed hook line | 5.1.1 | Rule 5 below: a relink is judged against what is on disk |
| First-time trust: the install step reports `waiting` until a person's session has trusted the folder | 5.1.7 | Trust stays a person's; `waiting` is finished as a person's next step in every output (below) |
| Unlink and revert: Claude Code's install record outside git | 5.1.7 | `unlink` removes the record; a plain revert's leftover record is tried and documented (below) |
| Claude Code rewrites `.claude/settings.json` in its own key order | 5.1.7 | Bonsai keeps the file's key order (below) |
| Local plugin scope leaks across worktrees | 5.1.7 | Already settled (project scope only, part 4b); 5.1.7 adds a test that every `claude` call Bonsai makes passes `--scope project` |
| `check`'s version warning: 2.1.294 is the first measured floor | 5.1.6 | Bonsai's floor constant is 2.1.294 (gate report §2.10); a pack's `needs.claude_code` raises it |
| `status --json`'s `formats` stays null; no `error` object | 5.1.4, 5.1.6 | `formats` read as below (no schema change); the `error` object in every command's `--json` |
| For a later set: a quoted `"format":` key and a tab still dispatches as format 0; lone-CR line endings get `quote-this-value`, whose next step does not help | 5.1.3 (cases), 5.1.2 or the reader (fix) | Two new cases with both outcomes; the format-1 outcome by contract §2.4's words; the reader made to reach it (below) |
| `STATE.md`: the reader's comments name `yaml.mjs` | 5.1.2 | Allowed as the contract's own reference ("What changes", item 10) |
| The coordinator, 9 Oct: tests resting on a fixed time | 5.1.0 (two), 5.3 (three) | 5.1.0 fixes the two shaped like the flake; 5.3 reviews the rest with the guard's budget (5.3's outline) |

#### Notes per piece

**5.1.0, tests that rest on a fixed time.** After `1170c92` (`TestRenameRetryOnWindows` held a file for a fixed 150 ms
and flaked once on a loaded Windows CI runner; it now holds the file until the first busy refusal), the builder of
that fix listed five tests whose outcome still rests on a fixed time budget, and changed none of them:
`TestEachFaultBlocks` (a 200 ms guard budget, tag `bonsai_test_fault`: on a loaded runner a fault case could get an
`over-time` record instead of its own); `TestOverTimeBlocks` (about 1.15 s, and the over-time record within
`recordWait`, 500 ms); `TestFaultsThroughTheHookLine/slow` (9 s against the guard's 5 s); "busy forever" in
`internal/workspace/places_test.go` (2 s, low risk); and the guard tests on the default 5 s budget (generous). 5.1.0
takes the first two, the flake's own shape. Each test must still prove what it proves: a fault case blocks with its
own record, and an over-time run blocks and is recorded. The guard's code is not changed (that is 5.3). The other
three go to 5.3, whose guard work may change the budgets they measure.

**5.1.1, consent to code.** Spec §6: "a change to a hook line or to a file a hook runs is listed under 'runs code'
and needs `--allow-exec` as well as `--yes`"; "All or nothing". The rules, which `init` and `update` share:
1. **Runs code:** a hook line added or changed in `.claude/settings.json` (Bonsai's own four kinds of line, spec §7's
   table, or a pack's); a pack file that a pack's hook line runs, changed or new; and a plugin's own code parts
   changed between the locked commit and the new one, or present at a pack's first link (rule 3). A removed hook line
   runs nothing: it is an ordinary settings line in the preview.
2. **All or nothing.** A run with anything under "runs code" and no `--allow-exec` writes nothing and exits 4, its
   preview naming each item under "Runs code" and `--allow-exec` as the next step, at a terminal too: it does not ask
   y/N for code. With `--allow-exec` and `--yes`, everything is written. `--allow-exec` without `--yes` and without a
   terminal prints the preview and exits 4, naming `--yes`. `--json` lists the items (`runs_code`). A mixed update is
   therefore refused whole without `--allow-exec`, never half written: the lock records one commit per pack, so a lock
   at B with A's hook line would describe a project that does not exist.
3. **A plugin's code parts** are what Claude Code runs on its own, without an agent's call: its hooks, MCP and LSP
   servers, and mods (spec §5), as the plugin reference lists them when the builder reads it (the version read goes in
   the run report). The engine compares those parts, the files they name included, between the locked commit and the
   new one in its pack cache. Roles, skills and commands are prompts: a change to them alone is not code. `base` and
   `workflow` carry none (Bonsai's hooks never ride in a plugin, spec §5), so in practice this is rare.
4. **The files a hook runs are declared.** A pack's hook entry in `pack.yaml` names the pack files its command runs
   (`runs`, a list of paths, `[]` for none); `check --pack` (5.1.9) refuses a hook command naming a pack file that
   `runs` does not list. Chosen over scanning commands for paths, which misses a file run through another.
5. **A relink is judged against the disk.** `init` on a project whose lock is missing compares its hook lines with the
   ones already in `.claude/settings.json`: a Bonsai line it would change counts as changed. (`update` already refuses
   when `bonsai.yaml` is there and the lock is not, naming `init` or a restore from git.)
6. **The first link.** Bonsai's own hook lines are the link's purpose and the preview names each with its sentence, so
   a first link writes them on `--yes` (or a y at a terminal). A pack's hook lines, the files they run and a plugin's
   code parts need `--allow-exec` at the first link too: they are a pack's code, not Bonsai's. A later change of
   Bonsai's own line (a new Bonsai, or 5.3's hook-line form) is "changed", so it needs `--allow-exec`.
- **Test-pack commits:** **E** changes one pack file and the hook line together (the mixed update); **F** adds a hook
  to the plugin itself (`hooks/`, an `echo` only) and nothing else. `internal/testpack` mirrors both.
- **The verifier's break-it:** every flag combination (`--yes`, `--allow-exec`, `--json`, `--keep`, `--adopt`,
  terminal or not) on C to D, D to E, E to F and a relink with the lock deleted, the tree hashed before and after each,
  as part 5's verifier did with 11 combinations (hook 19:18). No combination may write a hook line without
  `--allow-exec`.

**5.1.2, the format-0 reader.** Contract §2.4: "Bonsai's Go reader has a format-0 mode that is a hand port of
`yaml.mjs`, tested on today's files, never a general YAML library", and format 0 reads "with no new refusals".
- The source: `yaml.mjs` at `4a05eac` (`parseYaml`, `parseFrontmatter`), taken read-only with `git show` into a scratch
  folder, never committed, as part 0 took it (`plan.md` part 0, "The format-0 outcomes").
- The oracle: every case's `format0` outcome in `formats/expect.json`. An accepted case's value must equal the expected
  JSON exactly, numbers as the frozen reader gives them; a refused case must be refused. The refusal's message is not
  compared: the port names its own messages, each with a next step (spec §3).
- Today's files, as a scripted run, never committed: every tracked markdown file with frontmatter and every YAML file
  in a scratch clone of the studio's repo (origin removed, as part 3 made it), and in an export of Mimas made with
  Windows git (`git archive`, read-only), read by the frozen reader (Node) and by Bonsai's. The run report gives the
  counts and every difference by path and kind, never a file's content. Passes with no difference.
- Wiring: the dispatch already returns "format 0"; the format-0 mode is what a caller then reads with. Only Bonsai's
  kinds (`task`, `run`, `state`, contract §2.3) are read under format 0 by Bonsai's own checks.

**5.1.3, formats set 4.** One commit that changes `formats/` with its manifest (`set` 4), as `formats/README.md`'s
"How the set changes" asks.
- **New schemas**, documented as the ten are (a top-level `description`; a `description` and `examples` on every
  property; every field required, `null` or `[]` where it does not apply; closed lists as `enum`, open lists as
  strings): `workspace` (`bonsai.workspace/1`, every field of spec §6's example and its parts: packs, documents,
  `protected`, `person_only`, `never_edit`, `ladder_floor`, `ladder` with today's rung shape plus `tests` and
  `base_setup`, `ratchets`, `ci_marked_tests`, `generated` with its kinds, `keep_days` and `keep_newest`); `pack`
  (`bonsai.pack/1`: id, version, needs, the block, files with their kinds, hook entries with 5.1.1's `runs`, deny rules
  with their `why`, and the document kinds a pack declares); `tasks` and `sessions` (`bonsai.tasks/1`,
  `bonsai.sessions/1`: the data a reader returns, the frontmatter and the rows; a JSON Schema cannot describe a
  markdown table, so the writer's column order is the README's); `memory` (`bonsai.memory/1`, a note and the index,
  spec §10); `error` (`bonsai.error`: `code`, a closed list of fixed words; `message`; `next`, an object with `do`,
  what to do, and `who`, `agent` or `person`). Examples for each; the YAML ones (`workspace`, `pack`, `memory`) are
  also yaml-1 cases, as the five are.
- **Two new cases** (gate report §5, 5.1; reader 16:47): a quoted `"format":` key followed by a tab, and a file with
  lone-CR line endings. Both outcomes: format 0 from the frozen reader; format 1 by contract §2.4's words, the first
  case decided as its space-separated form already is, the second with a code or next step that names the line
  endings. A change to the README's order rules happens only here and only if §2.4's words require it; the verifier
  rules on it. The reader is made to reach both outcomes in the same commit.
- **The active-task fixtures** (contract §13: "one folder per case, run against every reader"): one folder per case
  under `formats/`, with the answer each reader must reach (`id`, `how`, `why`) in a file the README documents. The
  cases that need git (a worktree reading its task from main) are written as a layout each reader's test builds; none
  holds a `.git`. The studio's readers test against them when it links (step 7).
- **The schema-compare test** (contract §2.2: "A CI rung in Bonsai's repo fails on any schema change but
  additions"). It compares each schema with the same file at the set's base commit, read with `git show`; the base is
  a constant beside the test, moved forward in each set's commit. **An addition** is: a new schema file; a new
  property at the end of an object's `properties`, with its name added to `required`; a change to `description`,
  `examples` or `title`. Anything else fails: a property removed, renamed or moved; a type, `enum`, `pattern` or
  bound changed; a name taken out of `required`. CI checks out the history it needs; a checkout without the base
  commit skips with that reason, except under CI, where it fails. From 5.4 the same test is a rung of Bonsai's own
  ladder.
- **Chosen, so no existing schema changes:** `status --json`'s `formats` lists every format Bonsai knows, `read` the
  majors it reads (`[0, 1]` for `task`, `run`, `state`; `[1]` for the rest) and `write` the major a writer of this
  Bonsai writes, whether Bonsai itself writes the file or not: contract §12's own example gives `bonsai.task` a
  `write` of 1, though agents write tasks. Making `write` nullable would not be an addition.

**5.1.4, every format in Go.** Sixteen formats: the contract's ten and set 4's six. Each has a Go type that reads a
document (a missing field reads as `null`, an unknown one is kept byte for byte, a newer major is refused as "format
too new" and nothing else is read) and, where Bonsai writes it, writes it in the schema's field order, validated
against the embedded schema before it is written. Bonsai writes the lock, `status`, the tables, `bonsai.yaml`, the
log, asks, ladder results and the `error` object; 5.2 and 5.4 use the log, ask and ladder writers, built and tested
here. `bonsai.yaml` and `pack.yaml` are read in full and refused with a named field and next step when they break
their schema; a hook command that calls `bash` by name is refused (spec §3). `bonsai check --schema <format>` prints
the format with every field and allowed value (with `--json`, the schema), for all sixteen; an unknown name exits 2
listing the names. Every refusal and error of every word fills `error` in its `--json`, with a code from the `error`
schema's list. Every word's `--help` comes from one table of its flags and exit codes, so help and behaviour cannot
drift.

**5.1.5, what the engine reads and writes.**
- **The lock's `declares`** (spec §6, §16 row 18): per pack, the lanes, document kinds, label definitions and protected
  paths it declared at the locked commit, written at each `init` and `update`. `check` then runs from the lock alone:
  a test runs it with an empty pack cache and no network. Its inner layout stays open in the schema (format review
  6.3) and is documented in the README.
- **Document kinds** (contract §7.3): Bonsai's (`task`, `run`, `state`, `answers`, `memory`, `tasks`, `sessions`) at
  the paths `bonsai.yaml`'s `documents` names or their fixed files, and each pack's from `declares`.
- **Labels in force:** the packs' definitions from `declares`, and the ones attached on this machine, read from the
  machine folder's `labels/` (attaching is 5.6). `bonsai.*` definitions come with `base` in 5.5; until then tests use a
  fixture pack that declares them.
- **The active task** (contract §13): one function, used from here on by `status`, `check`, the tables and, later, the
  guard, rung 0 and the stop gate. Tasks are read from the main checkout's task folder; `--task` and `BONSAI_TASK`
  (only for the session's own project); none when they disagree, when two read `running`, or when a file does not
  parse. Proved on set 4's fixtures.
- **This machine's settings for the workspace**, read: `status_writes` (default `agents` with no settings file) and
  `status_command` (`null` unless `command`). `bonsai settings set` writes them in 5.6.
- **The instruction block** complete (spec §6): the workspace line, the import of the always-on protocol files, the
  import of the project's memory index, the label definitions agents see; at most 40 lines.
- **`bonsai.yaml`** written by `init` with every field and a comment on every line, from Bonsai's built-in template
  until `base` holds it (5.5).
- **Layers and the old workspace** (spec §6): the existing refusals of two packs writing one path and of a 0.4.3
  workspace (`.bonsai.yaml` or `.bonsai-lock.yaml`) get their tests.

**5.1.6, `check` and `status`.**
- **Findings** (exit 1; spec §6's list): the lock against the files and format-0 files changed (built); Bonsai-kind
  files failing their format, under format 0 or 1; labels against their definitions (a value of the wrong kind, a name
  defined twice); `approve_first` from git history, "history not available" in a shallow clone, never "passed"; an
  absolute path in any committed format; a changed workspace id (contract §3); each settings rule valid on its own;
  `disableAllHooks: true` in the project's settings or a local settings file; a `version` in a pack's `plugin.json`
  or its marketplace entry; plugin drift (built); the `bonsai` on the PATH not the installed one, when `install.json`
  exists (spec §3: until 5.6's installer writes it, nothing to compare, said as a note); the block over 40 lines, a
  memory index over 120 lines or 12 KB, a note over 4 KB; a path named in an instruction file, STATE or a memory note
  that does not exist; a tracked or staged `local/` file and `.gitignore` (built).
- **Warnings** (never the exit code): Claude Code older than the floor (Bonsai's, 2.1.294, or a pack's
  `needs.claude_code`, the higher; `claude --version` read defensively, an unreadable answer is itself a warning);
  two checkouts on this machine holding one id; a stale tasks table (5.1.8); run reports past `generated.run`'s rule.
- **Moved out of 5.1**, with reasons ("Stale or in tension"): a secret-shaped string in a committed memory note goes
  to 5.2, where the redactor's patterns have their one home; the stranded machine folder goes to 5.6, whose row names
  it.
- **`status --json`:** every field filled where it applies, `null` or `[]` only where the contract says it does not
  apply (`status_command` with `status_writes: agents`; `checks` without `--full`); the test's `notBuiltYet` list is
  emptied and replaced by that "does not apply" list. `needs`: each locked pack not installed on this machine (kind
  `plugin`), the Claude Code floor (kind `tool`, `name: claude-code`, never a problem), and the packs' declared needs.
  `--full` adds `checks`: newer tags of each pack (`git ls-remote`), Claude Code's version, and the MCP needs a pack
  declares. `--active` prints `active_task` only (contract §13's read-only command; the spec's §4 name). `--line` is
  5.6's.

**5.1.7, `unlink` and Claude Code's own records.**
- **`bonsai unlink [--yes] [--json]`** (spec §4): a preview first; without `--yes` and without a terminal, exit 4.
  It removes the pack files nobody edited, the block, Bonsai's entries in `.claude/settings.json`, the lock, the
  tables and `bonsai.yaml`. It leaves, and names: edited files, `once` and `kept` files (the project's),
  `.bonsai/STATE.md`, `.bonsai/local/`, `.bonsai/.gitignore` while `local/` holds anything (else its files would
  show to git), and the home's machine folder. A second `unlink` changes nothing; `init` after it links again.
- **The install record:** after removing, `unlink` runs `claude plugin uninstall` at project scope for each locked
  pack in this checkout, its result reported as the install's are (done, `waiting`, `failed`, `skipped`), never
  changing the exit code. A worktree's own record is its own; `unlink`'s help says so.
- **A plain `git revert`** (check 12) leaves Claude Code's record. Tried once as a script on an installed scratch
  target on each side: revert, then a `-p` session in it; the run report says whether the plugin loads (it should
  not: the revert removes `enabledPlugins`). `unlink`'s help names the leftover.
- **First-time trust:** trust stays a person's. Bonsai never answers Claude Code's trust question or registers a
  project's marketplace behind it (the skeleton's `--settings` route was a test method, not product behaviour).
  `waiting` gets its full form: in the human line and in `--json` (`next.who: person`: open a session in the folder
  and trust it, then run `bonsai update`), as a `needs` entry of kind `plugin` in `status --json`, and as a `check`
  warning.
- **Key order:** when Bonsai changes `.claude/settings.json` it keeps the file's existing key order, changing only its
  own entries; a new file is written in the order Claude Code writes, as measured by the builder on the version in
  use. If that order is not stable, the one rewrite at the first install is documented instead. Proof on both sides:
  `init`, install, commit, then `update` to the same commit and to a new one: `git diff` holds only Bonsai's changed
  lines.
- **Project scope only:** a test fails if any `claude` call Bonsai makes passes another scope.

**5.1.8, the tasks table and `check --write`.** `check --write` rebuilds `.bonsai/tasks.md` from the task files
(format 0 and 1): frontmatter `format: bonsai.tasks/1` with its pointer comment, the active task as contract §13's
step 2 finds it, then one row per task, newest id first. Main checkout only: in a worktree it refuses with exit 4
naming the main checkout. With `--write` the exit code says only whether it wrote (0 written, 3 could not); findings
are still listed. A table that differs from a rebuild is a warning everywhere, never a finding or a `problem`.
`sessions.md` joins `--write` in 5.2, from the log. The guard's refusal of a hand edit is 5.3's, rung 0's 5.4's.

**5.1.9, `check --pack`.** On a pack folder: `pack.yaml`, `labels.yaml`, `lanes.yaml` held to their schemas; every key
in them carries its comment; each template skill's fields table matches its template's frontmatter both ways, and for
a Bonsai format equals the schema; every allowed-values cell naming a closed list matches it; every deny rule has its
`why`; no `version` in `plugin.json`; the block's 40 lines; the declared document kinds well formed; no hook command
calling `bash` by name; every pack file a hook command names listed in its `runs`. Test-pack commit **G** documents
every file the rules reach (no behaviour change) and passes. `claude plugin validate` is the pack's CI, 5.5's.

**5.1.10, the reference page.** `docs/reference/lists.md`, generated by `go generate` from the code: every list (task
statuses, run outcomes, label value kinds, lane rules, ask ops and types, log events and categories, exit codes, the
`error` codes, file kinds, generated kinds), its values when Bonsai owns them, closed or open, where it is defined, and
the command that prints it in a project. Its test rebuilds the page in memory and fails on any difference. A
`.gitattributes` line keeps the page LF on every checkout. `CLAUDE.md` already tells builders to regenerate it with
any list change.

#### Proof for each piece

Every piece: its Go tests, `go vet`, plain and with the fault tag, in WSL and natively on Windows (check 10) before
the push, their counts in the run report; CI green on the pushed commit (`test`, `windows`, `lint`, `govulncheck`,
CodeQL); no Windows-only skip without a named reason; the Windows rules of `CLAUDE.md` read in the diff. Pieces that
run Claude Code (5.1.1's scripted check 5, 5.1.7) record its version and the user settings hashes before and after.
The three verifiers re-run the tests themselves.

#### 5.1 done

A fresh Opus verifier, at the end of 5.1, runs each check itself on the final commit and passes or fails 5.1:
1. **Formats in Go:** for each of the sixteen schemas, the Go type reads its example and, where Bonsai writes the
   format, writes it back byte for byte in schema order; `bonsai check --schema <name>` prints each (human and
   `--json`), and an unknown name exits 2.
2. **Format 0:** the reader reaches every `format0` outcome in `formats/expect.json`; the verifier runs the frozen
   reader itself on set 4's cases (from its own runner, as part 0's verifier did).
3. **Set 4:** the manifest matches every byte; every schema documents itself; the schema-compare test fails when the
   verifier removes a property in a temporary copy and passes when it adds one at the end.
4. **Consent:** the consent table's cases hold on a fresh build: the mixed update (D to E) with `--yes` alone exits 4
   and writes nothing, naming `--allow-exec`; with both flags it writes everything; F with `--yes` alone exits 4; a
   relink with the lock deleted and a changed hook line exits 4 without `--allow-exec`; a first link of Bonsai's own
   lines writes on `--yes`.
5. **The skeleton's checks** 1-6 and 12, re-run as scripts on the final build on scratch targets, still pass (check 5
   now with `--allow-exec`).
6. **`unlink`:** on an installed scratch target on each side, the preview, exit 4 without `--yes`, then the removal
   with what stays named; Claude Code's record for the checkout gone; a second run changes nothing.
7. **Key order:** `init`, install, commit, `update` on each side leaves only Bonsai's lines in `git diff`.
8. **`check`:** the findings-and-warnings table's test passes, one case per finding and warning of spec §6 but the two
   moved; warnings never change the exit code; a fake `claude --version` older, newer and unreadable gives the
   warning, nothing and a warning.
9. **`status`:** every field validated against the schema; nothing `null` or `[]` but what does not apply; `--full`
   and `--active` work; `formats` lists every format Bonsai knows with the majors it reads and writes.
10. **The active task:** every set-4 fixture gives its answer.
11. **`check --write`:** the table is byte-stable on two runs and on both sides; a worktree run exits 4 naming the main
    checkout; a stale table is a warning only.
12. **`check --pack`:** the test pack at G passes; each failing fixture fails on its own rule.
13. **The reference page:** `go generate` changes nothing; a changed list in a temporary copy fails the test.
14. **Unattended:** every word's `--help` lists every flag, exit code and an example; every refusal's `--json` has an
    `error` with a listed code and a `next` with `who`.
15. **Check 10 and CI:** `go test ./...` and `go vet ./...` plain and tagged, in WSL and natively on Windows, run by
    the verifier; CI green on the final commit.
16. **Stop lines:** 5.1's hours under 61; step 5's Windows-only tally; option rounds; Mimas and the studio's repo
    unchanged; the user settings hashes. **Nothing private:** a grep of the diff and the commit messages.

### Steps 5.2-5.7, outlined

Each gets its detailed section, in 5.1's shape, before it starts ("What changes", item 1). The spec rows are §14's.

**5.2 Recorder, logs, asks (25-37 h, re-ask at 48).**
- **Builds:** redaction in Go, with the three older leaks spec §8 names fixed and the differential check (10-15); the
  recorder (`bonsai hook record` for the eleven events, async; `bonsai hook start`'s opening context and the binary's
  path and hash), files in `.bonsai/local/log/`, `bonsai logs`, `bonsai log append`, cleaning per kind with its
  protections, the `clean` event and the generated-files page (9-13); asks to contract §9, no outbox (4-6); the
  sessions table from the log, a row per session and per subagent run with its task (2-3).
- **Needs:** 5.1's log, ask and sessions types; `generated` in `bonsai.yaml`; the active task (a start record's
  `target`); `check --write`.
- **Settles (gate report §5, 5.2):** the names of the log's binary path and SHA-256 fields (an addition at the end of
  `bonsai.log/1`, a set change with its manifest; format review 4.2 left them unnamed); `input_hash` keyed by the
  home's salt, which is made at first need and never leaves the machine; the hash logged once per session file (kept,
  now by `hook start`); builds without a commit stamp ("What changes", item 7); and from 5.1, the secret scan of
  memory notes, on the redactor's patterns.
- **Risks:** redaction is security (a fresh verifier). Its differential corpus is the studio's redaction tests at a
  named commit, read-only; it runs as a scripted differential like format 0's, or is copied in only after a scrub
  (made-up secrets only, no studio ids). Appends under concurrency on Windows (whole lines, busy retries). The
  generated-files page is a `base` skill in the spec (§6) and `base` arrives in 5.5: the section says where the page
  lives until then.
- **Rohan:** nothing. The studio's forwarder can start from here (its own work).

**5.3 Guards (18-29 h, re-ask at 38).**
- **Builds:** the hook adapter (2-3); the path guard with contract §5.5, §10.1 and §13 and its refusals for
  `.bonsai/local/` and the tables (7-10); the delete check (3-5); the stop gate (2-3); generated deny rules and
  `disableAllHooks: false` (1-2); the binary check (1-2); the second Windows check: guard, delete check, recorder under
  concurrency, backslash paths (2-4).
- **Needs:** 5.1's active task, lanes, `person_only` and labels; 5.2's log (guard records) and recorder.
- **Settles (gate report §5, 5.3):** how the hook line finds `bonsai`. Rohan's 8 Oct requirement (spec §7's note: a
  fixed path, nothing an agent may edit can redirect it) meets spec §3 ("Hooks call `bonsai` by name") and check 2 (no
  absolute path in a committed file). If no design meets all three, the section brings Rohan the choice as an option
  round. Also: a `.git` entry in the session's starting subfolder; `bonsai.yaml` read from the working tree (what the
  guard trusts to find the main checkout must not be rewritable unnoticed); shell `rm` and `mv` of `bonsai.yaml` or the
  lock; the Windows junction limits (dangling, swapped between check and write, the admin share: fixed or documented);
  a session's first call over 5 ms on WSL (5.73 ms at p50, the self-hash); the large-Write fail-closed test on
  Windows; and the coordinator's three remaining timed tests (`TestFaultsThroughTheHookLine/slow` at 9 s against the
  guard's 5 s, "busy forever" at 2 s, and the guard tests on the default 5 s budget), reviewed with the guard's budget.
- **Risks:** Windows; a guard that blocks real work ("cries wolf") gets worked around; a hook that times out does not
  block, so the guard's own timer stays. The whole part is guards and hooks: a fresh verifier, with a break-it run on
  both sides.
- **Rohan:** the second Windows check, spec §17 step 7, about 10 minutes, after a Sonnet agent's scripted run on both
  sides. Possibly the hook line's form as an option round.

**5.4 Ladder runner (28-44 h, re-ask at 57).**
- **Builds:** rungs, process groups and job objects, one ladder at a time, leftovers, `mode`, the fingerprint, results
  in the main checkout's `.bonsai/local/ladder/`, rung 0's refusal of branch changes to the tables and of tracked
  `local/` files (13-19); floors, ratchets, Bless filing (4-6); new tests must fail, by name, with `base_setup`
  (7-12); git integrity (3-5); Bonsai's own `bonsai.yaml`, the switch from the interim proof and the pre-release build
  Rohan installs (1-2).
- **Needs:** 5.1 (the ladder in `bonsai.yaml`, the active task, the tasks), 5.2 (the `ladder` record, the Bless ask),
  5.3 (rung 0 judges as the guard does; the stop gate reads the results).
- **Settles (gate report §5, 5.4):** Windows job objects and Windows Defender's first scan of a new binary, never
  measured; `golang.org/x/sys`, which part 5 measured at about 3 ms more on every Windows start (spec §3 plans it for
  job objects), kept off the hook's path or measured within its budget.
- **The switch, written before it happens:** Bonsai's rungs (rung 0, `go vet`, the tests plain and tagged, the schema
  compare; whether the native Windows run becomes a rung or stays beside the ladder); the task files Bonsai's own work
  then uses and how a builder's task is named (`BONSAI_TASK`), including for builders that run as subagents of the
  orchestrator's session; what `protected` and `person_only` hold in Bonsai's own repo; and the way back if the
  pre-release blocks work (a person's local settings line, or reinstalling the previous build), with exact lines for
  Rohan.
- **Risks:** Windows process trees; the stop gate binding the orchestrator's own builders; Bonsai links itself before
  `base` exists (5.5), so its tasks' `bonsai.*` labels have no definition in force until then; once Bonsai's hook
  lines are committed in its public repo, anyone opening Claude Code in a clone without `bonsai` installed is blocked
  (fail closed), which `CONTRIBUTING.md` must say. A fresh verifier for the runner and the switch.
- **Rohan:** spec §17 step 8, the pre-release install in WSL (about 5 minutes; his password), once; again only if a
  later part changes the guard, the stop gate or the ladder.

**5.5 Packs (19-30 h, re-ask at 39).**
- **Builds:** `base` and `workflow` from the studio's roles, protocols and templates; the always-on and skill split,
  the roles' `skills:` preloads; `claude plugin validate --json` in each pack's CI, failing on every warning but the
  missing `version`; the walls in base's deny rules and the studio's in `workflow`, each tried once on both sides
  (13-21); the documentation in every template and pack file, and each deny rule's `why` (3-4); the pack template
  `packs/template/` with its CI and release, and Bonsai's CI job that runs it (3-5).
- **Needs:** 5.1 (`check --pack`, the `pack` schema, `declares`), 5.3 (deny rules and the guard, for the walls), 5.4
  (Bonsai links `base` in its own repo, spec step 6).
- **Settles (gate report §5, 5.5):** `/agents` is gone, so a role is read by starting a session as it (`--agent
  <pack>:<role>`); the plugin's version shows only as the commit folder; not yet seen: a pack's role answering in an
  interactive session, a Windows `--bg` session, the Windows trust prompt; Claude Code's own writes outside the scratch
  folders; `claude --agent workflow:builder --bg` (check 8's original line); `claude plugin eval`'s needs
  (`--no-publish`, `--trust-plugin`, a path target, the miscounted `runsPerCase`). Also where `bonsai.yaml`'s and
  STATE's templates live (spec §4, §6 put them in `base`).
- **Risks:** Rohan's roles, protocols and templates move from the studio's repo, read-only, into public repos: every
  file is checked for anything private before its first push, and the section sets what may stay (studio paths and
  task ids inside his own protocols among it). Whether `validate` needs a Claude login on a CI runner (spec §5,
  unchecked) and whether evals run in CI (they need a key). A fresh verifier for the walls (security) and the pack CI.
- **Rohan:** approving the new public repo (`LastStep/bonsai-workflow`) and what goes into it; a repository secret if
  CI needs one.

**5.6 Machine pieces (13-20 h, re-ask at 26).**
- **Builds:** `settings` (with `--machine` and `cache_keep_days`), `labels`, the personal memory layer and its check,
  the stranded-folder report (6-9); `status --line`, the workspace half of today's statusline (5-8); the installers
  for `/usr/local/bin` and `C:\Program Files\Bonsai` with `install.json` (2-3).
- **Needs:** 5.1 (`status`, the home, machine settings read); what `--line` shows is set in 5.6's section, and it
  may read 5.2's asks and 5.4's ladder results.
- **Settles:** nothing new from the skeleton; from 5.1's split, the stranded-folder warning and the writes behind
  `status_writes` and attached labels.
- **Risks:** `settings set` and `labels` refuse in an agent session (`CLAUDE_CODE_CHILD_SESSION`), and the guard
  refuses them too; installers write root-owned and admin paths, so agents test them only against scratch targets;
  the personal index's import line in `~/.claude/CLAUDE.md` is a person's step (no test touches a real home);
  `--line` runs on every statusline refresh, so its speed is measured. A fresh verifier for the installers.
- **Rohan:** none inside 5.6 unless its section asks for one real install proof. The real installs on both sides come
  with 1.0 and the studio's link (spec §3, step 7).

**5.7 Release (6-11 h, re-ask at 14).**
- **Builds:** the supply-chain fixes (actions pinned by commit, GoReleaser pinned, the re-release input removed,
  immutable releases, build provenance; spec §12 step 8), `release.yml` back to tag runs inside the `release`
  environment and switched on again (Rohan's step), `bonsai@0.4` for Homebrew, the README.
- **Needs:** everything before it.
- **Settles (gate report §5, 5.7):** the Go 1.25.x toolchain bump that clears the standard-library findings
  govulncheck lists; CI's `lint` without the fault tag; the fault switch absent from the release build (checked on
  that build); the wording in `README.md`, `CONTRIBUTING.md`, `.gitattributes` and `CHANGELOG.md`.
- **Risks:** a release is public and cannot be taken back; agents act as an admin, so every tag, release, workflow
  switch, secret and setting stays Rohan's; the Homebrew tap is written only by the release workflow, with the token
  inside the `release` environment. A fresh verifier (CI and release).
- **Rohan:** the `release` environment and a new tap token (spec §17 step 4's note has the working command);
  `gh workflow enable release.yml`; the optional tag ruleset; his word for 1.0.

### Stop lines and hours

Judged by each part's end verifier, not the builder. Done means **no line crossed without Rohan's recorded choice, or,
past a line, his recorded choice to go on.**
1. **A part's hours over its re-ask line:** 5.1 61, 5.2 48, 5.3 38, 5.4 57, 5.5 39, 5.6 26, 5.7 14 (spec §14: 1.3
   times each part's high estimate, rounded: 47 x 1.3 = 61.1; 37 x 1.3 = 48.1; 29 x 1.3 = 37.7; 44 x 1.3 = 57.2; 30 x
   1.3 = 39; 20 x 1.3 = 26; 11 x 1.3 = 14.3). A part that ends under its line passes nothing on to the next.
2. **Windows-only failures over 8 hours in step 5**, carried over from the skeleton's line 2 ("more than a day lost
   to Go on Windows", spec §14). Kept for the whole step, not per part: it is the signal that Go on Windows is failing,
   the skeleton spent about 18 minutes against it, and 5.3 and 5.4 hold most of the risk. Counted as before: a failure
   also seen in WSL or on Linux CI does not count, nor copying to a Windows folder or waiting for CI.
3. **More than two option rounds asked of Rohan inside one part**, carried over per part. Step 5 has questions that
   are his (5.3's hook line, 5.5's public content); more than two in one part says its section left too much open.
   Plan-section approvals and hand checks are not option rounds.
4. **Any change to Mimas or to the studio's repo**, carried over whole. Read-only use of the studio's frozen files
   ("What changes", item 10) is not a change.

At a crossed line work stops. Rohan gets the numbers and three choices: continue, the smaller cut (spec §14's path
(b): defer the new-tests check, git integrity, `status --line` and half of 5.7), or pause. His choice is written in
the run report before any more work.

**How hours are counted.** As in the skeleton: every builder, verifier, audit and hand-check run's wall-clock, from
its "Runs" row (model, start, end, minutes) in the piece's run report; a run without a row is added by hand and marked
so; a part's planning and plan-review runs count in that part (this plan's count in 5.1). Not counted: the
orchestrator's own session, Rohan's time, waiting for CI. From 5.4, Bonsai's own `.bonsai/sessions.md` (5.2) also
holds these runs; a Haiku audit compares it with the run reports at each part's end, and the run reports stay the
source.

### How it is proved

| Done when | Proved by |
|---|---|
| 5.1.0 | The two tests pass plain and tagged on both sides; the diff holds only `_test.go` files |
| 5.1.1 | The consent table's Go tests; check 5 scripted with E and F; the verifier's break-it; check 10; CI |
| 5.1.2 | Every `format0` outcome; fuzz; no difference over today's files; the verifier's own frozen-reader run; check 10; CI |
| 5.1.3 | The formats test; the schema-compare test failing on a removal; the Windows-git clone's `go test ./formats/`; the verifier's rule-by-rule read; CI |
| 5.1.4 to 5.1.10 | Their Go tests; scripted runs where Claude Code is involved (5.1.7); check 10; CI; the orchestrator's read of the diff |
| 5.1 | The end verifier on "5.1 done" |
| 5.2 to 5.7 | Each part's section; its verifiers as outlined; its end verifier |
| The interim proof (to 5.3) | Before each push, `go test ./...` and `go vet ./...`, plain and tagged, in WSL and natively on Windows, counts in the run report; CI green on the pushed commit; no Windows-only skip without a named reason |
| The ladder proof (from 5.4) | `bonsai ladder` green on Bonsai's own `bonsai.yaml`, run by the pre-release Rohan installed; CI and check 10 beside it until a rung covers them; fresh verifiers for the big steps |
| Stop lines | The run reports' rows and tallies, judged by each part's end verifier |
| Hand checks | Rohan's lines, or the Sonnet agent's report, against each check's pass condition |
| Nothing private | Each verifier's grep of the diff and commit messages for home folders, machine or tailnet names, email addresses, studio task ids and studio paths (beyond item 10's exception) |

### Choices made, and what lost

| Choice | Alternative | Why this one |
|---|---|---|
| A mixed update without `--allow-exec` is refused whole | Write the rest, leave the hook line waiting | Spec §6 "all or nothing"; the lock records one commit per pack and must describe the project |
| A plugin's code parts compared between commits | Every pack move counts as code | Roles and skills are prompts; making every update need `--allow-exec` teaches people to pass it by habit |
| A pack names the files its hooks run (`runs`) | Scan hook commands for paths | A declared list `check --pack` can hold; a scan misses a file run through another |
| A relink judged against the disk | A link is consent, whatever is there | The hook path's verifier's finding: a changed hook line must not ride a relink |
| Bonsai's own lines written at a first link on `--yes` | `--allow-exec` for every link | They are the link's purpose and the preview names each; a pack's code still needs it |
| Format-0 refusals matched as refusals, messages not compared | The same messages | A Go port names its own messages with next steps; readers must agree on outcomes |
| Today's files read as a scripted run | Committed tests | They are private (as `plan.md`'s scratch targets) |
| The schema compare against the set's base commit, read with git | A frozen copy of each schema | A copy would be a second home for every list |
| "Addition" defined narrowly | Allow widening a type or a pattern | An older reader must read every newer writer's document |
| `formats.write` as the major a writer writes | Make `write` nullable | Contract §12's own example; nullable is not an addition |
| `error.next` as `{do, who}` | One sentence | An agent decides by `who` without parsing words (format review 4.5's question) |
| §13's fixtures in `formats/` | Go tests only | Contract §13: "run against every reader"; the studio tests against them at step 7 |
| The tables' schemas describe the rows a reader returns | The markdown text | A JSON Schema cannot describe a markdown table |
| `unlink` keeps `.bonsai/.gitignore` while `local/` stays | Remove it | Else the log and asks show to git |
| `unlink` removes Claude Code's install record | Leave it | The gate's finding (verifier N6) |
| Bonsai keeps the settings file's key order | Its own order | No churn after Claude Code's rewrite |
| Trust stays a person's; `waiting` with a person's next step | Register the marketplace headless | Trust is Claude Code's safety question for a person |
| The memory secret scan in 5.2; the stranded folder in 5.6 | All of §6 in 5.1 | One home for the secret patterns; 5.6's row names the stranded folder |
| 5.1.0 before 5.1.1 (or beside it) | Leave the timed tests to 5.3 | Every 5.1 push runs them on a loaded Windows runner; one flaked already |
| Two pieces beside another in 5.1 (5.1.0, 5.1.2) | None, or more | Rohan's 8 Oct rule: only where truly independent (no shared file, no shared proof) |
| A run report per piece | One per part | 5.1 is eleven pieces over days; each report stays one log |
| A verifier for risky pieces and at each part's end | One per piece | `CLAUDE.md`, Rohan's rule; spec §14's "on every step" met per part |
| The skeleton's Windows, option-round and Mimas lines carried over | Hours lines only | Each still guards what it was for; reasons in "Stop lines" |
| Planning and review runs count in their part | Outside the parts | The spec's estimates cover the whole part; conservative |

### Risk in the code

- **The format-0 port.** JavaScript's numbers and coercions are easy to get subtly wrong in Go. The oracle is the
  frozen reader's own outputs; today's files catch what the cases miss; the verifier runs the frozen reader itself.
- **Consent.** A code path that writes a hook line without `--allow-exec` is the failure that matters. The verifier's
  break-it hashes the tree around every flag combination.
- **Claude Code moves.** Trust, key order, `plugin uninstall`, `claude --version`'s form and the plugin reference's
  component list are read on the version in use and recorded; a change later is a finding for the part that meets it.
- **The schema compare needs git history.** A shallow checkout fails it under CI rather than passing.
- **Size.** 5.1 is the largest part with 5.4. The re-ask line is 61 h; the run reports keep its running total.
- **Windows.** New generated files (`docs/reference/lists.md`, the tables) must be byte-stable and LF on a Windows
  checkout; check 10 runs before every push.
- **The public test pack** grows by appended commits only; a force push is blocked anyway.
- **Processes.** Scripted Claude Code sessions in 5.1.1 and 5.1.7: the orchestrator sweeps after each agent and kills
  by pid only what that agent clearly left.

### Stale or in tension in the spec

- **§17 step 6** names `/agents` and one sitting. `/agents` is gone in Claude Code 2.1.294; Rohan chose two sittings
  and a Sonnet agent for hand checks (8 Oct). Step 5 reads a role by starting a session as it (`--agent
  <pack>:<role>`).
- **§17 step 8** says "Do step 3's WSL lines first": done on 8 Oct (all four old binaries removed).
- **§18 and §19 D** give 60 h and 30-46 h, then 61 h and 30-47 h; step 5 uses §14's table.
- **§14, "How Bonsai's work is proven"**, asks "a fresh verifier on every step"; `CLAUDE.md` (Rohan's rule) asks one
  for big or risky work and at the end of each plan step. This plan does both: risky pieces, and every part's end.
- **§3 against §7's 8 Oct note and check 2:** hooks call `bonsai` by name (§3); Rohan asked for a fixed path no agent
  can redirect (§7); no committed file may hold an absolute path (check 2, contract §2.6). 5.3's section settles it,
  with Rohan if no design meets all three.
- **§3's `golang.org/x/sys`** for job objects against part 5's measured 3 ms on every Windows start: 5.4.
- **§4 and §6: `init` writes `bonsai.yaml` and STATE from `base`'s templates**, but `base` arrives in 5.5. Until then
  `bonsai.yaml` comes from a built-in template and `init` writes no STATE; 5.5 settles the templates' one home.
- **§14's 5.1 row says "all of §6's findings and warnings"**, while its 5.6 row names "the stranded-folder report" and
  "the personal memory layer and its check". The stranded folder goes to 5.6; the secret scan of memory notes to 5.2
  (the redactor's patterns are its one home); the rest of §6 is 5.1's.
- **`status_writes` and `status_command`:** the status test's comment puts them in 5.1 "(bonsai settings)", but
  `settings` is 5.6's word. 5.1 reads the machine settings; 5.6 writes them.
- **Contract §12's `formats.write`** is a required integer, but Bonsai only reads some formats; read as the major a
  writer writes (contract §12's own example), so no schema changes.
- **Contract §2.2's "fails on any schema change but additions"** does not say what an addition is for a JSON Schema;
  defined in 5.1.3.
- **Format review 4.5** left two "worth a look" questions open under Rohan's "Agree": `who` as its own field (taken:
  `next.who`), and whether he wants the code words before they are built. The words are in the `error` schema and the
  reference page; the orchestrator sends him the list in one line when 5.1.3 lands, no option round, and a word he
  wants changed is changed before 5.1.4 builds on it.
- **Contract §13's working name** `bonsai task active --json` against spec §4's `status --active`: the spec fixes
  command words, so `--active`.
- **The generated-files page** is a `base` skill (§6), but cleaning is 5.2 and `base` is 5.5: 5.2's section says where
  it lives meanwhile.
- **Base's `bonsai.*` label definitions** arrive with 5.5, but Bonsai links itself at 5.4: 5.4's section settles what
  its tasks' labels are checked against meanwhile.
- **Step 6, "it registers as its own studio project"**, needs the studio's registration (contract §15.2, "after the
  skeleton"), which is studio work while its Desk is on upkeep until step 7. Bonsai's side is ready at 5.4; when it
  shows on the Desk is the studio's plan.
- **Step 7, "the studio links (after 5.5)"**, includes the machine installs of `bonsai`, but a 5.x build is not a
  release (§3, §18) and the one pre-release is WSL's, at 5.4. Before 1.0 the studio would link on that pre-release
  (WSL only) or wait for 5.7: the studio's plan decides; nothing here waits on it.
- **Bonsai's own screens are gone** (Rohan, 9 Oct: "completely remove the idea of bonsai's own screens ... this
  visual part of the job will be handled by the studio, while bonsai is a pure cli tool"). Overtaken: §11 whole
  (`bonsai serve`, the static bundle, the Fable mock), §1's and §14's 30-50 h for the screens, §4's `bonsai serve`
  line, §14 row 9 and path (d) (§14 "Totals", §15), §16 row 13's and contract §1, §8 and §12's "Bonsai's screens" as a
  reader. Nothing in step 5 builds a page; contract §2.6 still keeps update diffs off the VPS, so they are read with
  `bonsai update --diff`. The orchestrator adds the dated notes to the spec and contract on `main`; this plan only
  follows them.
- **§14 row 10** puts "the first public release" after row 9, which went with the screens. This plan reads the 1.0
  release as Rohan's word at the end of 5.7 (`CLAUDE.md`: releases are his word at step 5.7).
- **§8's three older redaction leaks** are named by studio bug ids; Bonsai's public code and tests describe the cases,
  never the ids.
