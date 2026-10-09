# Bonsai's plan for step 5: path (a), Bonsai 1.0, with 5.1 in full

- **Status:** approved by Rohan, 9 Oct 2026 ("the step 5 plan is approved"). Reviewed by a fresh Opus agent (ready
  after fixes) and fixed on its findings and on Rohan's two answers of 9 Oct (the plans of later parts, and Bonsai's
  scope).
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
rest. 5.1, 5.2, 5.3 and 5.4 are planned in full here; 5.5 to 5.7 are outlined, and each gets its own detailed section
in this file before it starts.

## For Rohan (plain words)

**What step 5 builds.** Bonsai 1.0, in seven parts, one after the other. What each lets you do or see:

| Part | What you will be able to do or see when it is done | Spec hours | Re-ask at |
|---|---|---|---|
| 5.1 Formats and engine to 1.0 | Bonsai reads and writes every format, old files included. An update that would add or change code that runs (a hook) waits for a second, separate yes. A pack can be taken out of a project, and `bonsai unlink` takes Bonsai out cleanly. `bonsai check` reports the problems the spec lists (two more come with 5.2 and 5.6), `bonsai status` is complete, and one page lists every list of allowed values. | 30-47 | 61 |
| 5.2 Recorder, logs, asks | Every session in a linked project leaves a log inside the project, with secrets hidden as it is written. Agents can ask you typed questions and read your answers. Old records are cleaned by rules you set. A table of sessions and hours per task. | 25-37 | 48 |
| 5.3 Guards | The full guard: an agent may change a protected file only while its running task allows it; files only you may grant stay yours; a recursive delete that does not name what it deletes is refused; a builder cannot stop before its proof is green. You run one real Windows session (about 10 minutes). | 18-29 | 38 |
| 5.4 Ladder runner | `bonsai ladder` proves a task's work, on WSL and Windows. From here Bonsai proves and guards its own repo with it, on a pre-release you install in WSL (about 5 minutes, your password). | 28-44 | 57 |
| 5.5 Packs | Bonsai's `base` pack and your `workflow` pack (your roles, lanes, protocols and templates) as public Claude Code plugins, each with its own checks on GitHub; a template for new packs; the walls round your key and token files. | 19-30 | 39 |
| 5.6 Machine pieces | Settings per machine, label files the studio attaches, your personal memory layer, the workspace half of the statusline, and installers for both sides. | 13-20 | 26 |
| 5.7 Release | The release path made safe and switched back on with you; on your word, Bonsai 1.0. | 6-11 | 14 |
| **Step 5** | | **139-218** | each part its own |

**What Bonsai is, and what the studio is** (your 7 Oct split, which you confirmed on 9 Oct). Bonsai is everything
that lives inside one project: its rules, its guard, the record of what happened, and the proof that work is done. It
works with no studio. The studio is everything across projects and anything that acts on agents: dispatch,
approvals, status moves, notifications, and every picture of the data. Bonsai has no scheduler: starting agents is
the studio's.

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

**What you must do, and when.** Each item reaches you in one batch with exact lines, when its part gets there.
- **Now:** approve this plan. Nothing in step 5 is built before. (Approved, 9 Oct.)
- **In 5.1, a look (no vote needed):** the list of error words every command's `--json` uses, and its "what next"
  part, which becomes two fields (what to do, and whether an agent or a person does it) instead of the one sentence you
  saw in format review 4.5; with it, the shapes of `check`'s and `update`'s JSON, which the studio will read. Until
  1.0 any word can still change.
- **In 5.2, a look (no vote needed):** the names of the log's two new fields, the `bonsai` binary's path and its
  fingerprint (`bonsai_path`, `bonsai_sha256`), which format review 4.2 left for you to see before building; any change
  you want is made before they land (added 9 Oct with 5.2's section).
- **In 5.3:** the second Windows check (spec §17 step 7, about 10 minutes): a real Windows session in a scratch folder,
  asking Claude for one edit and two deletes and reporting what happened. A Sonnet agent runs every check an agent can
  first, so your sitting is only what needs a person typing (your 8 Oct word). Two questions, each an option round in
  5.3's section, answered on 9 Oct: how the hook lines find `bonsai` (your answer: (a), the installed place written into
  the lines); and who may consent to code (`--allow-exec`) and take Bonsai's guard out of a project (`unlink`) (your
  answer: (ii), your grant in a project the studio manages, only a person elsewhere). From 5.4, so, a command of yours
  when an update of Bonsai's own repo runs code or changes its guard lines, until the studio manages it.
- **At 5.4:** install the pre-release `bonsai` in WSL (spec §17 step 8, about 5 minutes, your password). The
  orchestrator gives you its fingerprint first. Your step 3 (the old binaries) is already done (8 Oct). You install
  again only if a later part changes the guard, the stop gate or the ladder.
- **In 5.5:** approve the new public repo for your `workflow` pack and what goes into it (your roles, lanes, protocols
  and templates, after a check for anything private); possibly a repository secret, if a pack's checks on GitHub need
  a Claude login or a model key.
- **In 5.6:** one line in your own `~/.claude/CLAUDE.md` that loads your personal memory in every project (spec §10).
  No agent edits that file.
- **At 5.7:** the GitHub release steps that are yours (the `release` environment and a new Homebrew tap token, spec §17
  step 4's note; switching the release workflow back on), then your word for 1.0.
- **Outside step 5:** the sandbox probe (spec §7) waits for its own small plan after step 5 and ends in a root step of
  yours (installing `socat`). Turn GitHub Pages off (the old website) whenever you like.
- Hand checks: a Sonnet agent runs them wherever an agent can (your 8 Oct word). Yours are only what needs a person: a
  typed session, a password, a GitHub setting, a file in your own home.

**When the studio links, and when Bonsai joins the Desk.**
- **Bonsai joins the Desk** (spec step 6): from 5.4 Bonsai is linked to itself (its own `bonsai.yaml`, its own
  ladder and guard). The studio's registration of a project (contract §15.2) is the studio's own work, not in this
  plan or its hours, and two of its steps use Bonsai commands that arrive in 5.6 (attaching the studio's label file,
  and this machine's settings). So the studio can show Bonsai from 5.4 and register it fully after 5.6.
- **The studio links** (spec step 7) after 5.5: its own plan, in its own repo. Mimas links last (step 8). This plan
  changes nothing in either repo, and waits on neither.

**Choices that are yours.** None open. Three were decided on 9 Oct: the two in 5.3's section, how the hook lines find
`bonsai` (your answer: (a)) and who may consent to code and take the guard out (your answer: (ii)), and this one:
- **How the later parts' plans reach you: (B).** You were offered (A) every part's section, (B) only when it changes
  what is yours, and (C) none; you first said (A), then chose (B) once the scope question above was settled. So: 5.1
  is planned in full here. Before each later part starts, a planner writes its section into this file, a fresh Opus
  agent reviews it and the orchestrator fixes what the review finds. The section comes to you when it changes what is
  yours: its hours or re-ask line; the order of the parts; your steps (the list above, or a new one); a new public
  repo or content made public (5.5); a question only you can answer (an option round, such as 5.3's hook line); or a
  format change. **A format change** means a new major of a format, or a removal: of a format, of a field, or of a
  value from a closed list. An addition (a new field, a new format, a new value in an open list) is not one and does
  not come to you by itself, unless this plan routes it to you for another reason (as with 5.1's error words).
  Otherwise the part starts after the review's fixes and you get one line with a link. Under these rules 5.4 (your
  install), 5.5 (a new repo, your content public) and 5.7 (the release) come to you.

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
  you inside one part; or step 5's own work changes anything in Mimas or the studio's repo (the studio now works in its
  own repo at the same time, so only step 5's own commands count). Your choice is written down before any more work.

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
   Windows-only tally. A part's planning and plan-review runs count in its hours: from 5.2 on they go in the report of
   its first piece; this plan's runs are in `records/runs/R-2026-10-09-plan-5.md` and are carried into 5.1's total.
   Step 5's Windows-only tally starts at 0 (the skeleton's 18 minutes are the skeleton's); while pieces run side by
   side, the orchestrator keeps one running total of it and of the part's hours, and each report quotes it.
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
   the same commit. The orchestrator pushes each once the builder has committed it and the piece's Go tests pass on
   `internal/testpack`'s mirror, before any scripted run that fetches it: appending a commit changes nothing for A to
   D. If the piece then fails, the commit stays and the fix is a later commit.
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
11. **The toolchain bump moves to where it is needed.** The Go 1.25.x bump that clears govulncheck's standard-library
    findings is listed for 5.7 (gate report §5). If `govulncheck` turns red on CI before then, the piece that meets it
    lands the one-line bump as its own commit first, and its run report says so.

**Outside step 5, named so they are not lost.**
- **The sandbox probe** (spec §7, §16 row 25; contract §9.5): it "waits for its probe after the gate", with its
  fallback for `.bonsai/local/` (about 1-2 h with the probe, "outside step 5's totals like the probe itself"). It gets
  its own small plan after step 5, and ends in a root step of Rohan's (installing `socat`).
- **Part 0's run report quotes a studio task id** (`STATE.md`, loose ends; the plan's quoted line in
  `records/runs/R-2026-10-08-formats.md`). Run reports are a log, and no piece edits them; the orchestrator may take the
  id out in a records commit of its own.

### Step 5.1: formats and engine to 1.0 (30-47 h, re-ask at 61)

The spec's row (§14): "Every contract format in Go and the format-0 hand port; the schema-compare rung (contract §2.2);
layers, `--allow-exec`, `unlink`, the old-workspace refusal, all of §6's findings and warnings, `--schema`, `status
--full` (26-40); the tasks table and `check --write`, with its stale rules (2-3); the reference page of lists,
generated, and its CI test (1-2); `check --pack`, the template docs kept in step with their fields (1-2)". It also
settles every 5.1 finding of the gate report's section 5.

**What exists** (the skeleton's code at `54b4fdd`, plus `1170c92`'s test fix for `TestRenameRetryOnWindows`; read each
package's doc comment):
- `internal/reader`: the format-1 reader, all 116 format-1 outcomes; files with no top-level `format:` come back as
  "format 0", unread.
- `internal/schema`: ordered JSON, a strict decoder (duplicate keys refused), byte-stable ASCII encoding, the schema
  checker.
- `internal/workspace`: `bonsai.yaml` and `pack.yaml` (the fields the skeleton used; every other key kept as read),
  the lock, the home and machine folder, checkouts, atomic writes with Windows' busy retry.
- `internal/engine`: the fetch, `init`, `update` (staged, the four kinds plus `kept`, exits 4 and 5, `--diff`,
  `--yes`, `--keep`, `--adopt`), the preview's settings lines, the block, `check` (lock and files, a listed format-0
  file changed, `.gitignore`, tracked `local/` files, plugin drift), the plugin install step (`installed`, `waiting`,
  `failed`, `skipped`). The 0.4.3-workspace refusal and the "two packs write one path" refusal are in `plan.go`. Two
  things it refuses or leaves for 5.1: taking a pack out of `bonsai.yaml` (`plan.go`: "taking a pack out of a project
  comes with step 5.1"), and the lock's `format0` list, which `init` writes empty and `update` only copies.
- `internal/status`: every field of `bonsai.status/1`, nine of them held `null` or `[]` in the test's `notBuiltYet`.
- `internal/guard`: the one-rule guard, its fault switch behind the `bonsai_test_fault` tag; it reads `bonsai.yaml`
  through `workspace.LoadConfig` on every call (`internal/guard/hook.go`).
- `formats/`: set 3, ten schemas, 116 trick cases, the manifest; `formats/embed.go` embeds the schemas.
- `cmd/bonsai`: `--version`, `--help`, `status`, `init`, `update`, `check`, `hook guard`; `--allow-exec` exits 2.
  `internal/engine/render.go` says the commands' JSON outputs and the `error` object get their formats in 5.1.

**The pieces, in order.** Hours: the spec gives 26-40 for its first item as one figure; the split across pieces 5.1.1
to 5.1.7 below is the planner's judgment for sizing briefs, not a spec figure. It sums to the spec's figure (low
3+3+4+3+2+4+4+3 = 26; high 5+5+6+4+4+6+6+4 = 40, pieces 5.1.1, 5.1.2, 5.1.3, 5.1.4a, 5.1.4b, 5.1.5, 5.1.6, 5.1.7).
Pieces 5.1.8 to 5.1.10 carry the spec's own figures. Piece 5.1.0 is not in the spec's row (below).

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.1.0 | **A test that rests on a fixed time.** `TestEachFaultBlocks` (`internal/guard/fault_on_test.go`, a 200 ms guard budget under the fault tag, which CI runs on both sides) made to prove each fault's own block without racing a short budget, as `1170c92` did for `TestRenameRetryOnWindows`. Test files only: a fix that needs guard code goes to 5.3 | The test passes plain and tagged on both sides; the diff touches only `_test.go` files (the orchestrator's read) | not in the spec's rows; its minutes count in 5.1's | The coordinator's note of 9 Oct (below) |
| 5.1.1 | **Consent to code.** `--allow-exec`; the mixed update; a pack's hook lines, the files they run and a plugin's own code parts counted as running code, at a first link too; the relink with the lock deleted. Test-pack commits E and F; local fixture packs for the cases the test pack does not hold | Go tests walking a table of every consent case on A to F and the fixture packs; every existing Go test and script that links the test pack updated (below); check 5 re-run as a script with E and F; the verifier's break-it, first links and relinks included | 3-5 | Spec §4 (`init`, `update`), §5 (what a plugin carries; mods), §6 ("How `update` decides", "Code is consented to separately", the preview), §7 (the hook lines), §14 checks 1-6 and 12; gate report §3 rows 1-6 and 12, §5 (5.1); `plan.md` "Stale or in tension"; `records/runs/R-2026-10-08-engine.md` 17:40 |
| 5.1.2 | **The format-0 reader**: a hand port of `yaml.mjs` at `4a05eac`, reached by the reader's dispatch for a file with no top-level `format:`. It touches `internal/reader` only | Every case's format-0 outcome in `formats/expect.json`; a fuzz run with no crash; a scripted run over today's files against the frozen reader with no difference, reported without paths | 3-5 | Contract §2.3, §2.4; `formats/README.md` ("expect.json", "How the outcomes were found"); spec §3 |
| 5.1.3 | **Formats set 4**, after 5.1.2 lands: the schemas `workspace`, `pack`, `tasks`, `sessions`, `memory` and `error`, and the two command outputs the studio reads, `check` and `changes` (the preview and result of `init`, `update` and `unlink`), each documented, with an example; `error` added at the end of `status` (an addition); the two cases the reader's verifier left for a later set, with the reader's fix; contract §13's fixtures with an answers file per reader; the manifest at set 4; the schema-compare test | The formats test (manifest, docs, examples, the rules list); the reader's and the port's tests on the new cases; the schema-compare test shown failing on a removal and passing on an addition; a Windows-git clone's `go test ./formats/`; the verifier's rule-by-rule read; CI | 4-6 | Contract §2.2, §2.4, §2.8, §3, §5.2, §7.3, §7.4, §7.5, §12, §13, §14; spec §3, §4 and §16 row 29 (`error`, the outputs), §5 (`pack.yaml`), §6 (`bonsai.yaml`, `declares`, generated kinds), §10 (memory); format review 1.4, 3.5, 4.5, 6.1, 6.3, R2.5, R2.6; `formats/README.md` ("How the set changes") |
| 5.1.4a | **Every format in Go**: a reader (and a writer where Bonsai writes it) for each of the eighteen formats, held to its schema; `bonsai.yaml` and `pack.yaml` read in full by the engine and `check`, while the guard's own read stays lean; `check --schema F` | Go tests: each example read, written back byte for byte in schema order where Bonsai writes it, and validated; `--schema` for all eighteen; a test holding the guard's read to the fields it judges; the hook's p50 and p95 on WSL before and after, in the run report | 3-4 | Contract §2.1, §2.2, §2.5, each format's section (§4-§9, §11, §12, §14); spec §3, §5 (`pack.yaml`), §6 (`bonsai.yaml`), §16 rows 16, 17 |
| 5.1.4b | **The commands' outputs**: the `error` object in every command's `--json`, its words in one Go table; `init`, `update`, `check` (and, from 5.1.7, `unlink`) writing their `--json` to set 4's `changes` and `check` schemas; every word's `--help` from one table of flags and exit codes | Every refusal's `--json` carries an `error` with a known word and `next.who`; each command's `--json` validated against its schema; a test that help and the flag table agree | 2-4 | Spec §3 (output, unattended, `error`), §4, §16 row 29; contract §12; format review 4.5 |
| 5.1.5 | **What the engine reads and writes, to 1.0**: the lock's `declares` and its `format0` list; the moved-tag refusal; declared document kinds; labels in force; the active task (contract §13), one function every reader uses; this machine's settings, read; the instruction block complete; `bonsai.yaml` written with every field | Go tests on §13's fixtures (the active-task function's answers) and on `t.TempDir()` projects; `check` running offline from the lock alone (no pack cache, no network) | 4-6 | Spec §5 (pinning, "CI needs no pack"), §6 (`bonsai.yaml`, the lock, the block, layers), §10 (the index import); contract §2.3, §3, §5.1-§5.3, §6, §7.3, §13, §14 |
| 5.1.6 | **`check` and `status` to 1.0**: every finding and warning of spec §6 (two moved, below), "a format-0 file new or changed" among them; the Claude Code floor at 2.1.294; `status --json` with every field, `--full` and `--active` | One Go test per finding and warning, walked from one table so a finding with no test fails; the status test's `notBuiltYet` list empty | 4-6 | Spec §3 (the tripwire, `install.json`), §6 ("`bonsai check` findings", "Warnings"), §7 (the version floor), §10 (budgets); contract §2.3, §2.6, §3, §6, §12, §13; gate report §2.10 |
| 5.1.7 | **`unlink`, taking a pack out, and Claude Code's own records**: `bonsai unlink`; `update` taking a pack out of a project; the project-scope install record removed with each; first-time trust's `waiting` finished; Claude Code's key order kept in `.claude/settings.json`; project scope only, held by a test | Go tests; scripted runs on both sides on installed scratch targets (`unlink`, a pack taken out, check 12 re-run with an installed plugin, the key-order run) | 3-4 | Spec §4 (`unlink`), §5 ("This machine's install follows the lock", "Trust"), §6 (layers, the preview), §7 (no `bonsai.yaml`: the hook exits 0); gate report §3 rows 4 and 12, §5 (5.1); `records/runs/R-2026-10-08-plugins.md` 21:14 |
| 5.1.8 | **The tasks table and `check --write`**, with its stale rule; `init` writing the tables | Go tests: the table byte-stable from format-0 and format-1 tasks; the worktree refusal; the exit codes; stale is a warning only; `init` writes both tables | 2-3 | Spec §4 (`init`), §6 ("The two tables"); contract §7.5, §13 |
| 5.1.9 | **`check --pack`**: a pack folder's declarations and its documentation kept in step with its fields. Test-pack commit G | Fixtures failing each rule; the test pack at G passing | 1-2 | Spec §5 ("Every template and pack file documents itself"), §3 (no `bash` by name); contract §2.8 |
| 5.1.10 | **The reference page of lists**, `docs/reference/lists.md`, generated with `go generate`, and its test | The test rebuilds the page and fails on any difference (shown with a changed list in a temporary copy) | 1-2 | Spec §6 ("Every list has one home"), §12; contract §2.2 |
| **5.1** | | | **30-47** (re-ask 61) | |

**The order, and why.**
1. **5.1.0 first, or beside 5.1.1.** Every 5.1 push runs the test on a loaded Windows runner; one test of the same
   shape flaked there already (`1170c92`). It touches one test file, in a package 5.1.1 does not change: truly
   independent.
2. **5.1.1 next** (the gate report and `STATE.md`: "`--allow-exec` and the mixed update first"). It settles how
   `init` and `update` judge code, which the relink, `unlink`, taking a pack out and every preview reuse, before more
   code builds on the skeleton's "refuse only".
3. **5.1.2 may run beside 5.1.1.** It touches `internal/reader` only, which 5.1.1 does not, and its proof
   (`expect.json`'s format-0 outcomes) does not rest on 5.1.1's.
4. **5.1.3 after 5.1.1 and 5.1.2 have landed:** the `pack` schema carries 5.1.1's `runs`; the two new cases need the
   port to test their format-0 outcome, and they share `internal/reader` and `expect.json` with 5.1.2. 5.1.3 owns the
   reader's fix for the two cases. Everything after reads set 4's schemas, the one home of every list.
5. **5.1.4a after 5.1.3:** its Go types are held to set 4's schemas; Bonsai-kind files read under format 0 or 1.
6. **5.1.4b after 5.1.4a:** the outputs are written by 5.1.4a's writers.
7. **5.1.5 after 5.1.4b:** the active task reads typed task files; `declares` writes typed lanes, kinds and labels.
8. **5.1.6 after 5.1.5:** the findings need `declares` (offline), `format0`, lanes (`approve_first`), labels and the
   active task.
9. **5.1.7 after 5.1.6:** `unlink` and taking a pack out reuse 5.1.1's consent and 5.1.4b's outputs; they change the
   engine's apply and plugin code, which 5.1.6 also touches, so they do not run together.
10. **5.1.8 after 5.1.7:** the table's header is the active task, its stale warning is one of `check`'s, and `init`
    and `unlink` both handle the tables.
11. **5.1.9 after 5.1.8:** it reads `pack.yaml` through 5.1.4a's types and checks 5.1.1's `runs`.
12. **5.1.10 last:** it lists every list, so it is generated once all of 5.1's lists exist; from then on its test
    guards every later part.

**Who builds and verifies.** Opus builders for 5.1.0 to 5.1.7 and 5.1.9 (the formats, a reader, consent to code,
the engine, Claude Code's files); Sonnet for 5.1.8 and 5.1.10 (mechanical, with the orchestrator's read). Fresh Opus
verifiers: **5.1.1** (consent to code is security: what runs on a person's machine), **5.1.2** (a reader), **5.1.3**
(the set every reader tests against, and a CI change). Pieces 5.1.0 and 5.1.4a to 5.1.10 land on green tests on both
sides, CI and the orchestrator's read of the diff, which the run report says; **the 5.1 end verifier** covers them.
5.1.4a needs no verifier of its own because it leaves the guard's path as it is (below); if it cannot, it gets one
(guards and hooks).

#### Where each inherited finding is settled

The gate report's section 5, its 5.1 list, item by item; then what the skeleton's code itself leaves for 5.1, one
loose end from `STATE.md`, and the coordinator's note of 9 Oct:

| Finding | Settled in | How |
|---|---|---|
| `--allow-exec`; the mixed update between §6's "all or nothing" and check 5's words (gate §5) | 5.1.1 | Rules 1-8 below: a run with any code to consent writes nothing without `--allow-exec` |
| A plugin's move to a new commit is not counted as running code (gate §5) | 5.1.1 | Rule 3 below: the plugin's own code parts are compared between the two commits |
| With the lock deleted, `init --yes` links again and writes a changed hook line (gate §5) | 5.1.1 | Rule 7 below: a relink is judged against what is on disk |
| First-time trust: the install step reports `waiting` until a person's session has trusted the folder (gate §5) | 5.1.7 | Trust stays a person's; `waiting` is finished as a person's next step in every output (below) |
| Unlink and revert: Claude Code's install record outside git (gate §5) | 5.1.7 | `unlink` and taking a pack out remove the record; a plain revert's leftover record is tried and documented (below) |
| Claude Code rewrites `.claude/settings.json` in its own key order (gate §5) | 5.1.7 | Bonsai keeps the file's key order (below) |
| Local plugin scope leaks across worktrees (gate §5) | 5.1.7 | Already settled (project scope only, part 4b); 5.1.7 adds a test that every `claude plugin install` and `uninstall` Bonsai runs passes `--scope project` |
| `check`'s version warning: 2.1.294 is the first measured floor (gate §5) | 5.1.6 | Bonsai's floor constant is 2.1.294 (gate report §2.10); a pack's `needs.claude_code` raises it |
| `status --json`'s `formats` stays null; no `error` object (gate §5) | 5.1.4a, 5.1.4b, 5.1.6 | `formats` read as below (no schema change); the `error` object in every command's `--json`, `status`'s included |
| For a later set: a quoted `"format":` key and a tab still dispatches as format 0; lone-CR line endings get `quote-this-value`, whose next step does not help (gate §5) | 5.1.3 | Two new cases with both outcomes; the format-1 outcome by contract §2.4's words; the reader made to reach it in the same commit (below) |
| The code: "taking a pack out of a project comes with step 5.1" (`plan.go`, `check.go`) | 5.1.7 | `update` takes a pack out (below) |
| The code: the lock's `format0` list is never filled, so a new format-0 file is never found | 5.1.5, 5.1.6 | Written at a first link; "new or changed" is a finding (below) |
| The moved tag: "a tag that later resolves to another commit is refused" (spec §5; `R-2026-10-08-engine.md` 17:40: not built) | 5.1.5 | Refused at `update` (below) |
| `STATE.md`: the reader's comments name `yaml.mjs` | 5.1.2 | Allowed as the contract's own reference ("What changes", item 10) |
| The coordinator, 9 Oct: tests resting on a fixed time | 5.1.0 (one), 5.3 (four) | 5.1.0 takes `TestEachFaultBlocks`; `TestOverTimeBlocks` rests on a guard constant, so it goes to 5.3 with the other three (5.3's outline) |

#### Notes per piece

**5.1.0, a test that rests on a fixed time.** After `1170c92` (`TestRenameRetryOnWindows` held a file for a fixed
150 ms and flaked once on a loaded Windows CI runner; it now holds the file until the first busy refusal), the builder
of that fix listed five tests whose outcome still rests on a fixed time budget, and changed none of them:
`TestEachFaultBlocks` (a 200 ms guard budget, tag `bonsai_test_fault`: on a loaded runner a fault case could get an
`over-time` record instead of its own); `TestOverTimeBlocks` (about 1.15 s, and the over-time record within
`recordWait`, 500 ms); `TestFaultsThroughTheHookLine/slow` (9 s against the guard's 5 s); "busy forever" in
`internal/workspace/places_test.go` (2 s, low risk); and the guard tests on the default 5 s budget (generous). 5.1.0
takes the first, the flake's own shape, in its test file only: each fault case must still block with its own record.
`TestOverTimeBlocks`'s record half rests on `recordWait`, a constant in `internal/guard/hook.go`, so changing it is
guard code: it goes to 5.3 with the other three, whose guard work may change the budgets they measure.

**5.1.1, consent to code.** Spec §6: "a change to a hook line or to a file a hook runs is listed under 'runs code'
and needs `--allow-exec` as well as `--yes`"; "All or nothing". The rules, which `init` and `update` share:
1. **Runs code:** a hook line added or changed in `.claude/settings.json` (Bonsai's own four kinds of line, spec §7's
   table, or a pack's); a pack file that a pack's hook line runs, changed or new; and a plugin's own code parts
   changed between the locked commit and the new one, or present at a pack's first link (rule 6). A removed hook line
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
   (`runs`, a list of paths, `[]` for none). A hook entry with no `runs` (the test pack's A to D) reads as `[]`, the
   reader rule for a missing field. `check --pack` (5.1.9) refuses a hook command naming a pack file that `runs` does
   not list. Chosen over scanning commands for paths, which misses a file run through another.
5. **A hook-run file changed alone** (the hook line the same) is "runs code" by rule 1: a local fixture pack holds this
   case (a hook line that runs a pack file, then a commit changing only that file), since the test pack has none.
6. **The first link.** Bonsai's own hook lines are the link's purpose and the preview names each with its sentence, so
   a first link writes them on `--yes` (or a y at a terminal). A pack's hook lines, the files they run and a plugin's
   code parts need `--allow-exec` at the first link too: they are a pack's code, not Bonsai's. **So every first link
   to the test pack now needs `--allow-exec`:** its commit A carries a hook line ("echo demo hook A"), which the
   skeleton's first link wrote on `--yes` alone (`R-2026-10-08-engine.md` 17:40: "a first link consents to its hook
   lines"). Every Go test and script that links the test pack passes `--allow-exec` at its first link, and the run
   report lists each one changed: among them `TestCheck1InitIntoADriftedProject` to `TestCheck6ConflictKeepAdopt`,
   `TestHookLineChangeIsRefused` and `TestCheck12RevertTheLink` (`internal/engine/engine_test.go`), `TestInitCommand`
   (`cmd/bonsai/engine_test.go`), `TestStatusOfAProjectTheEngineLinked` (`internal/status/status_test.go`), the
   plugin tests in `internal/engine/plugins_test.go`, and part 3's scripted checks under `~/bonsai-checks/scripts/`.
   That a first link of Bonsai's own lines alone writes on `--yes` is shown with a local fixture pack that has no hook
   line. A later change of Bonsai's own line (a new Bonsai, or 5.3's hook-line form) is "changed", so it needs
   `--allow-exec`.
7. **A relink is judged against the disk.** `init` on a project whose lock is missing compares every hook line, and
   every pack file a hook runs, with what is already on disk: one that would change counts as changed, one already
   there and equal is not new code. A plugin's code parts have no locked commit to compare with, so at a relink they
   count as at a first link (rule 6). (`update` already refuses when `bonsai.yaml` is there and the lock is not,
   naming `init` or a restore from git.)
8. **Taking a pack out** (5.1.7) removes its hook lines, which runs nothing, so it needs `--yes` only.
- **Test-pack commits:** **E** changes one pack file and the hook line together (the mixed update); **F** adds a hook
  to the plugin itself (`hooks/`, an `echo` only) and nothing else. `internal/testpack` mirrors both. They are pushed
  before the scripted check 5 runs ("What changes", item 6).
- **The verifier's break-it:** every flag combination (`--yes`, `--allow-exec`, `--json`, `--keep`, `--adopt`,
  terminal or not) on a first link to A, a first link to the no-hook fixture, C to D, D to E, E to F, the hook-run-file
  fixture, and a relink with the lock deleted, the tree hashed before and after each, as part 5's verifier did with 11
  combinations (hook 19:18). The rule it holds: **no pack hook line, no file a hook runs, no plugin code part, and no
  change to an existing hook line is written without `--allow-exec`; Bonsai's own lines at a first link are the one
  thing `--yes` alone writes.**

**5.1.2, the format-0 reader.** Contract §2.4: "Bonsai's Go reader has a format-0 mode that is a hand port of
`yaml.mjs`, tested on today's files, never a general YAML library", and format 0 reads "with no new refusals". It
touches `internal/reader` only.
- The source: `yaml.mjs` at `4a05eac` (`parseYaml`, `parseFrontmatter`), taken read-only with `git show` into a scratch
  folder, never committed, as part 0 took it (`plan.md` part 0, "The format-0 outcomes").
- The oracle: every case's `format0` outcome in `formats/expect.json`. An accepted case's value must equal the expected
  JSON exactly, numbers as the frozen reader gives them; a refused case must be refused. The refusal's message is not
  compared: the port names its own messages, each with a next step (spec §3).
- Today's files, as a scripted run, never committed: every tracked markdown file with frontmatter and every YAML file
  in a scratch clone of the studio's repo (origin removed, as part 3 made it), and in an export of Mimas made with
  Windows git (`git archive`, read-only), read by the frozen reader (Node) and by Bonsai's. The run report gives
  counts and each difference by its index in the run's file list and its kind (a value or an outcome), never a path or
  a file's content: the paths name studio and Mimas files and task ids, so the list stays in the scratch folder.
  Passes with no difference.
- Wiring: the dispatch already returns "format 0"; the format-0 mode is what a caller then reads with. Only Bonsai's
  kinds (`task`, `run`, `state`, contract §2.3) are read under format 0 by Bonsai's own checks.

**5.1.3, formats set 4.** One commit that changes `formats/` with its manifest (`set` 4), as `formats/README.md`'s
"How the set changes" asks. It starts after 5.1.2 has landed.
- **New schemas**, documented as the ten are (a top-level `description`; a `description` and `examples` on every
  property; every field required, `null` or `[]` where it does not apply; closed lists as `enum`, open lists as
  strings):
  - `workspace` (`bonsai.workspace/1`): every field of spec §6's example and its parts: packs, documents,
    `protected`, `person_only`, `never_edit`, `ladder_floor`, `ladder` with today's rung shape plus `tests` and
    `base_setup`, `ratchets`, `ci_marked_tests`, `generated` with its kinds, `keep_days` and `keep_newest`;
  - `pack` (`bonsai.pack/1`): id, version, needs, the block, files with their kinds, hook entries with 5.1.1's
    `runs`, deny rules with their `why`, and what a pack declares besides its `labels.yaml` and `lanes.yaml`: its
    document kinds and its protected paths, the two that spec §6's `declares` copies into the lock with the lanes and
    labels;
  - `tasks` and `sessions` (`bonsai.tasks/1`, `bonsai.sessions/1`): the data a reader returns, the frontmatter and the
    rows; a JSON Schema cannot describe a markdown table, so the writer's column order is the README's;
  - `memory` (`bonsai.memory/1`): a note and the index (spec §10);
  - `error` (`bonsai.error`): `code`, an **open** list of fixed words (a string; readers show an unknown word as
    `other`), so later parts add words as additions (format review 4.5: "Adding a word is an addition"); `message`;
    `next`, an object with `do` (what to do) and `who` (`agent` or `person`, closed);
  - the two command outputs the studio reads, now that Bonsai has no screens: `check` (`bonsai.check/1`: findings,
    warnings, and `error`) and `changes` (`bonsai.changes/1`: the preview or result of `init`, `update` and `unlink`:
    packs, files with their results, settings lines with `file`, `change`, `line` and `why`, `runs_code`, the plugin
    step, and `error`). Later parts add their commands' outputs (`asks`, `logs`, `ladder`) the same way, each a new
    schema file, an addition;
  - `status` gains `error` at the end of its properties (an addition): `null` unless the command refuses.
  Examples for each. The sources that are files are also yaml-1 cases, as the five are: `workspace.yaml` and
  `pack.yaml` (YAML), and `memory.md`, `tasks.md` and `sessions.md` (markdown with frontmatter).
- **Where an open list's known words live.** The ten formats' open lists name their known values in the schema's
  description. The `error` codes, and from 5.2 the log's events, are read by code at many places, so their known words
  live in one Go table (their one home), which `check --schema` prints, and the reference page too once 5.1.10 lands;
  the schema's description names those commands and does not copy the words.
- **Two new cases** (gate report §5, 5.1; reader 16:47): a quoted `"format":` key followed by a tab, and a file with
  lone-CR line endings. Both outcomes: format 0 from the frozen reader, which the port (5.1.2) must also reach; format
  1 by contract §2.4's words, the first case decided as its space-separated form already is, the second with a code or
  next step that names the line endings. A change to the README's order rules happens only here and only if §2.4's
  words require it; the verifier rules on it. 5.1.3 owns the reader's fix, in the same commit.
- **Contract §13's fixtures** ("one folder per case, run against every reader"): one folder per case under
  `formats/`, and an answers file the README documents, with one section per reader: the active-task function (`id`,
  `how`, and `why` as a reason code from a short list, never free text), the guard, rung 0 and the stop gate. 5.1
  tests the function's section; 5.3 tests the guard's and the stop gate's, 5.4 rung 0's. The cases that need git (a
  worktree reading its task from main) are written as a layout each reader's test builds; none holds a `.git`. The
  studio's readers test against them when it links (step 7).
- **The schema-compare test** (contract §2.2: "A CI rung in Bonsai's repo fails on any schema change but
  additions"). It compares each schema with the same file at the set's base commit, read with `git show`; the base is
  a constant beside the test, moved forward in each set's commit. **An addition** is: a new schema file; a new
  property at the end of an object's `properties`, with its name added to `required`; a change to `description`,
  `examples` or `title`. Anything else fails: a property removed, renamed or moved; a type, `enum`, `pattern` or
  bound changed; a name taken out of `required`. CI checks out the history it needs; a checkout without the base
  commit skips with that reason, except under CI, where it fails. From 5.4 the same test is a rung of Bonsai's own
  ladder. A new field is required in its schema because a schema describes what a writer writes (the README); a
  reader still reads it as `null` when an older writer left it out, which is contract §2.2's "a new optional field".
- **Chosen, so no existing schema changes:** `status --json`'s `formats` lists every format Bonsai knows, `read` the
  majors it reads (`[0, 1]` for `task`, `run`, `state`; `[1]` for the rest) and `write` the major a writer of this
  Bonsai writes, whether Bonsai itself writes the file or not: contract §12's own example gives `bonsai.task` a
  `write` of 1, though agents write tasks. Making `write` nullable would not be an addition.

**5.1.4a, every format in Go.** Eighteen formats: the contract's ten and set 4's eight. Each has a Go type that reads
a document (a missing field reads as `null`, an unknown one is kept byte for byte, a newer major is refused as "format
too new" and nothing else is read) and, where Bonsai writes it, writes it in the schema's field order, validated
against the embedded schema before it is written. Bonsai writes the lock, `status`, the two outputs, the tables,
`bonsai.yaml`, the log, asks, ladder results and the `error` object; 5.2 and 5.4 use the log, ask and ladder writers,
built and tested here.
- **A reader refuses** a present field whose type or value its schema does not allow (a missing field is `null`, an
  unknown one is kept), naming the field and the next step. A hook command that calls `bash` by name is refused (spec
  §3).
- **The guard's read stays lean.** The guard reads `bonsai.yaml` on every call (`internal/guard/hook.go`). The engine
  and `check` read it in full; the guard keeps a read of only the fields it judges by (the format line, the protected
  lists, `never_edit`), so an error in an unrelated field (a ladder rung, a generated kind) does not block every edit,
  and a first call does not pay for the full read. A test holds the guard's read to those fields; the run report gives
  the hook's p50 and p95 on WSL before and after (part 5's harness; spec §3's 5 ms; a session's first call is already
  5.73 ms at p50, gate report §2.5).
- `bonsai check --schema <format>` prints the format with every field and allowed value (with `--json`, the schema),
  for all eighteen, and an open list's known words from its Go table; an unknown name exits 2 listing the names.

**5.1.4b, the commands' outputs.** Every refusal and error of every word fills `error` in its `--json`, with a word
from the one Go table and `next.who`. `init`, `update` and `check` write their `--json` as set 4's `changes` and `check`
documents (`unlink` from 5.1.7), validated against them in tests. Every word's `--help` comes from one table of its
flags and exit codes, so help and behaviour cannot drift. When this piece lands, the orchestrator sends Rohan the error
words and the two output shapes as a look (his part, "What you must do"); a word he wants changed is changed before
1.0, since nothing has been released.

**5.1.5, what the engine reads and writes.**
- **The lock's `declares`** (spec §6, §16 row 18): per pack, the lanes, document kinds, label definitions and protected
  paths it declared at the locked commit, written at each `init` and `update`. `check` then runs from the lock alone:
  a test runs it with an empty pack cache and no network. Its inner layout stays open in the schema (format review
  6.3) and is documented in the README.
- **The lock's `format0`** (contract §2.3, §14): at a first link, `init` lists every file of Bonsai's kinds (`task`,
  `run`, `state`, at the paths `documents` names and `.bonsai/STATE.md`) that has no `format:` line, with the SHA-256
  of its bytes with line endings made LF. `update` keeps the list as it is.
- **The moved tag** (spec §5: "A tag that later resolves to another commit is refused"): when `bonsai.yaml`'s ref is
  the one the lock holds but the tag now resolves to another commit, `update` refuses with exit 4, writes nothing and
  names the step (a person checks the pack's tag; to take the new commit, a new tag in `bonsai.yaml`).
- **Document kinds** (contract §7.3): Bonsai's (`task`, `run`, `state`, `answers`, `memory`, `tasks`, `sessions`) at
  the paths `bonsai.yaml`'s `documents` names or their fixed files, and each pack's from `declares`.
- **Labels in force:** the packs' definitions from `declares`, and the ones attached on this machine, read from the
  machine folder's `labels/` (attaching is 5.6). `bonsai.*` definitions come with `base` in 5.5; until then tests use a
  fixture pack that declares them.
- **The active task** (contract §13): one function, used from here on by `status`, `check`, the tables and, later, the
  guard, rung 0 and the stop gate. Tasks are read from the main checkout's task folder; `--task` and `BONSAI_TASK`
  (only for the session's own project); none when they disagree, when two read `running`, or when a file does not
  parse. Proved on the function's section of set 4's fixtures.
- **This machine's settings for the workspace**, read: `status_writes` (default `agents` with no settings file) and
  `status_command` (`null` unless `command`). `bonsai settings set` writes them in 5.6.
- **The instruction block** complete (spec §6): the workspace line, the import of the always-on protocol files, the
  import of the project's memory index, the label definitions agents see; at most 40 lines.
- **`bonsai.yaml`** written by `init` with every field and a comment on every line, from Bonsai's built-in template
  until `base` holds it (5.5).
- **Layers and the old workspace** (spec §6): the existing refusals of two packs writing one path and of a 0.4.3
  workspace (`.bonsai.yaml` or `.bonsai-lock.yaml`) get their tests.

**5.1.6, `check` and `status`.**
- **Findings** (exit 1; spec §6's list): the lock against the files (built); a format-0 file of Bonsai's kinds that
  changed (built) or is new, not on the lock's `format0` list (contract §2.3: "give it `format: bonsai.task/1`
  first"); Bonsai-kind files failing their format, under format 0 or 1; labels against their definitions (a value of
  the wrong kind, a name defined twice); `approve_first` from git history, "history not available" in a shallow clone,
  never "passed"; an absolute path in any committed format; a changed workspace id (contract §3); each settings rule
  valid on its own; `disableAllHooks: true` in the project's settings or a local settings file; a `version` in a
  pack's `plugin.json` or its marketplace entry; plugin drift (built); the `bonsai` on the PATH not the installed one,
  when `install.json` exists (spec §3: until 5.6's installer writes it, nothing to compare, said as a note); the block
  over 40 lines, a memory index over 120 lines or 12 KB, a note over 4 KB; a path named in an instruction file, STATE
  or a memory note that does not exist; a tracked or staged `local/` file and `.gitignore` (built).
- **Warnings** (never the exit code): Claude Code older than the floor (Bonsai's, 2.1.294, or a pack's
  `needs.claude_code`, the higher; `claude --version` read defensively, an unreadable answer is itself a warning);
  two checkouts on this machine holding one id; a stale tasks table (5.1.8); run reports past `generated.run`'s rule.
- **Moved out of 5.1**, with reasons ("Stale or in tension"): a secret-shaped string in a committed memory note goes
  to 5.2, where the redactor's patterns have their one home; the stranded machine folder goes to 5.6, whose row names
  it.
- **`status --json`:** every field filled where it applies, `null` or `[]` only where the contract says it does not
  apply (`status_command` with `status_writes: agents`; `checks` without `--full`; `error` when nothing refused); the
  test's `notBuiltYet` list is emptied and replaced by that "does not apply" list. `needs`: each locked pack not
  installed on this machine (kind `plugin`), the Claude Code floor (kind `tool`, `name: claude-code`, never a problem),
  and the packs' declared needs. `--full` adds `checks` as contract §12 lists them: newer tags of each pack (`git
  ls-remote`), Claude Code's version, and whether the MCP servers the packs' `needs` name (kind `mcp`) are reachable,
  as far as Bonsai can tell without starting a session (how it reads that is the builder's to measure and record; a
  server it cannot judge is shown as unknown, never as a problem). `--active` prints `active_task` only (contract
  §13's read-only command; the spec's §4 name). `--line` is 5.6's.

**5.1.7, `unlink`, taking a pack out, and Claude Code's own records.**
- **`bonsai unlink [--yes] [--json]`** (spec §4): a preview first; without `--yes` and without a terminal, exit 4.
  It removes the pack files nobody edited, the block, Bonsai's entries in `.claude/settings.json`, the lock, the
  tables and `bonsai.yaml`. `bonsai.yaml` goes even when a person edited it (spec §7: with no `bonsai.yaml` a session
  still running gets no guard), and the preview names it as removed with its edits. It leaves, and names: other
  edited files, `once` and `kept` files (the project's), `.bonsai/STATE.md`, `.bonsai/local/`, `.bonsai/.gitignore`
  while `local/` holds anything (else its files would show to git), and the home's machine folder. A second `unlink`
  changes nothing; `init` after it links again.
- **Taking a pack out:** a pack dropped from `bonsai.yaml` is taken out by `update`, which today refuses it. The
  preview names everything the pack wrote: its files (removed if nobody edited them, else left and named), its part of
  the block, its settings lines and deny rules, its lock entry and `declares`. A removed hook line runs nothing (rule
  8), so `--yes` is enough unless the same run adds or changes code. Its plugin is uninstalled as below. `check`'s
  finding about a pack in the lock and not in `bonsai.yaml` then names `bonsai update` as the step.
- **The install record:** `unlink` and taking a pack out run `claude plugin uninstall` at project scope for each pack
  they take out, in this checkout, the result reported as the install's are (done, `waiting`, `failed`, `skipped`),
  never changing the exit code. The builder measures whether the uninstall works before or after Bonsai's own entries
  leave `.claude/settings.json`, and orders them so it does, recording the Claude Code version. A worktree's own
  record is its own; `unlink`'s help says so.
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
- **Project scope only:** a test fails if any `claude plugin install` or `uninstall` Bonsai runs passes another scope.

**5.1.8, the tasks table and `check --write`.** `check --write` rebuilds `.bonsai/tasks.md` from the task files
(format 0 and 1): frontmatter `format: bonsai.tasks/1` with its pointer comment, the active task as contract §13's
step 2 finds it, then one row per task, newest id first. Main checkout only: in a worktree it refuses with exit 4
naming the main checkout. With `--write` the exit code says only whether it wrote (0 written, 3 could not); findings
are still listed. A table that differs from a rebuild is a warning everywhere, never a finding or a `problem`. From
this piece `init` writes the tables too (spec §4): `tasks.md` as `--write` builds it, and `sessions.md` with its
frontmatter and no rows; 5.2 fills the sessions table from the log and adds it to `--write`. The guard's refusal of a
hand edit is 5.3's, rung 0's 5.4's.

**5.1.9, `check --pack`.** On a pack folder: `pack.yaml`, `labels.yaml`, `lanes.yaml` held to their schemas; every key
in them carries its comment; each template skill's fields table matches its template's frontmatter both ways, and for
a Bonsai format equals the schema; every allowed-values cell naming a closed list matches it; every deny rule has its
`why`; no `version` in `plugin.json`; the block's 40 lines; the declared document kinds and protected paths well
formed; no hook command calling `bash` by name; every pack file a hook command names listed in its `runs`. Test-pack
commit **G** documents every file the rules reach (no behaviour change) and passes. `claude plugin validate` is the
pack's CI, 5.5's.

**5.1.10, the reference page.** `docs/reference/lists.md`, generated by `go generate` from the code: every list (task
statuses, run outcomes, label value kinds, lane rules, ask ops and types, log events and categories, exit codes, the
`error` words, the active task's reason codes, file kinds, generated kinds), its values when Bonsai owns them, closed
or open, where it is defined, and the command that prints it in a project. Its test rebuilds the page in memory and
fails on any difference. A `.gitattributes` line keeps the page LF on every checkout. `CLAUDE.md` already tells
builders to regenerate it with any list change.

#### Proof for each piece

Every piece: its Go tests, `go vet`, plain and with the fault tag, in WSL and natively on Windows (check 10) before
the push, their counts in the run report; CI green on the pushed commit (`test`, `windows`, `lint`, `govulncheck`,
CodeQL); no Windows-only skip without a named reason; the Windows rules of `CLAUDE.md` read in the diff. Pieces that
run Claude Code (5.1.1's scripted check 5, 5.1.7) record its version and the user settings hashes before and after.
The three verifiers re-run the tests themselves.

#### 5.1 done

A fresh Opus verifier, at the end of 5.1, runs each check itself on the final commit and passes or fails 5.1:
1. **Formats in Go:** for each of the eighteen schemas, the Go type reads its example and, where Bonsai writes the
   format, writes it back byte for byte in schema order; `bonsai check --schema <name>` prints each (human and
   `--json`), and an unknown name exits 2.
2. **Format 0:** the reader reaches every `format0` outcome in `formats/expect.json`; the verifier runs the frozen
   reader itself on set 4's cases (from its own runner, as part 0's verifier did).
3. **Set 4:** the manifest matches every byte; every schema documents itself; the schema-compare test fails when the
   verifier removes a property in a temporary copy and passes when it adds one at the end.
4. **Consent, on the final build:** the verifier re-runs 5.1.1's break-it itself (first links, relinks, C to F, the
   hook-run-file fixture), adding `unlink` and taking a pack out, since 5.1.5 and 5.1.7 changed the write path after
   5.1.1's verifier. Among its cases: the mixed update (D to E) with `--yes` alone exits 4 and writes nothing, naming
   `--allow-exec`, and with both flags writes everything; F with `--yes` alone exits 4; a first link to A with `--yes`
   alone exits 4; a first link to the no-hook fixture writes Bonsai's own lines on `--yes`; a relink with the lock
   deleted and a changed hook line exits 4 without `--allow-exec`.
5. **The skeleton's checks** 1-6 and 12, re-run as scripts on the final build on scratch targets, still pass, every
   first link to the test pack now with `--allow-exec` (rule 6).
6. **`unlink`:** on an installed scratch target on each side, the preview, exit 4 without `--yes`, then the removal
   with what stays named (`bonsai.yaml` removed even when edited); Claude Code's record for the checkout gone; a second
   run changes nothing.
7. **Taking a pack out:** a project linked to the test pack and a fixture pack; the fixture pack dropped from
   `bonsai.yaml`; `update` previews and, with `--yes`, removes exactly what that pack wrote, its plugin uninstalled;
   `check` clean.
8. **Key order:** `init`, install, commit, `update` on each side leaves only Bonsai's lines in `git diff`.
9. **`check`:** the findings-and-warnings table's test passes, one case per finding and warning of spec §6 but the two
   moved; a new format-0 file is found; warnings never change the exit code; a fake `claude --version` older, newer
   and unreadable gives the warning, nothing and a warning.
10. **Offline:** `check` in a linked project with an empty pack cache and no network gives the same findings as with
    them.
11. **What `init` writes:** `bonsai.yaml` with every field of the schema and a comment on every line; the block with
    its imports (the protocols and the memory index) and the label definitions, at most 40 lines; the lock's
    `declares` and `format0`; the two tables.
12. **The moved tag:** a fixture pack whose tag is moved to another commit: `update` exits 4 and writes nothing.
13. **`status`:** every field validated against the schema; nothing `null` or `[]` but what does not apply; `--full`
    and `--active` work; `formats` lists every format Bonsai knows with the majors it reads and writes.
14. **The active task:** every case in the function's section of set 4's fixtures gives its answer.
15. **`check --write`:** the table is byte-stable on two runs and on both sides; a worktree run exits 4 naming the main
    checkout; a stale table is a warning only.
16. **`check --pack`:** the test pack at G passes; each failing fixture fails on its own rule.
17. **The reference page:** `go generate` changes nothing; a changed list in a temporary copy fails the test.
18. **Unattended:** every word's `--help` lists every flag, exit code and an example; every refusal's `--json` has an
    `error` with a known word and a `next` with `who`; `init`, `update`, `unlink` and `check`'s `--json` validate
    against their schemas.
19. **The guard's path:** the guard's read holds only its fields; the hook's p50 and p95 on WSL against 5.1.4a's
    before and after.
20. **Check 10 and CI:** `go test ./...` and `go vet ./...` plain and tagged, in WSL and natively on Windows, run by
    the verifier; CI green on the final commit.
21. **Stop lines:** 5.1's hours under 61; step 5's Windows-only tally; option rounds; nothing in Mimas or the studio's
    repo changed by step 5's work (as "Stop lines" says); the user settings hashes. **Nothing private:** a grep of the
    diff and the commit messages.

### Step 5.2: the recorder, logs and asks (25-37 h, re-ask at 48)

**Rohan's (B).** Nothing in this section is his, so it reaches him as one line with a link once the review's fixes
are in. That line carries one look, no vote, as 5.1 sends him its error words: the names chosen for the binary's path
and hash in the log, `bonsai_path` and `bonsai_sha256` (format review 4.2 asked him "Want to see them before
building?", still open); a name he wants changed is changed before 5.2.0 lands. The hours (25-37) and the re-ask line
(48) are the spec's. The order of the parts and his steps do not change: 5.2 needs no sitting of his. An agent runs
the recorder's real Windows session once Claude Code's Windows login is back (`STATE.md`, "Waiting on Rohan"), or
reads the log of the session he already runs in 5.3 (spec §17 step 7); nothing more is asked of him. No repo is new
and no studio content goes public: every redaction test Bonsai commits is written fresh, and the studio's are only
read, in a scratch run. No option round is asked. Every format change is an addition: two fields at the end of the
log record, two new command outputs, one field at the end of a sessions row if set 4 lacks it, new error words, and
the log's known events kept in a Go table.

The spec's row (§14), with its three bug ids put in words: "Redaction in Go with [the three older leaks §8 names]
fixed and the differential check (10-15); the recorder, files in `.bonsai/local/`, `logs`, `log append`, and cleaning
per kind with its protections, the `clean` event and the generated-files page (9-13); asks to contract §9, no outbox
(4-6); the sessions table from the log, a row per session and per subagent run with its task (2-3). The studio's
forwarder can start here". It also settles the gate report's 5.2 findings and two items 5.1 hands on (the memory
secret scan; the sessions table's rows).

**What exists** on `main` at `3427a07` (read each package's doc comment):
- `internal/guard` writes one `guard` record per decision (`record.go`): every field of contract §8.1 in order,
  `input_hash` null (no salt yet), then two fields outside the schema, `bonsai_path` on every record and
  `bonsai_sha256` on the record that made its file. Records go to the session's own project folder's
  `.bonsai/local/log/` (`s-<session>.ndjson`, or `w-<UTC date>.ndjson` without a session), opened with `O_APPEND`,
  one write per line, retried for 1 s while Windows reports the file busy, cut to 2,048 bytes by halving the longest
  text. Its note: a worktree writes its own folder, not the main checkout's, because finding main in a way no agent
  can redirect is 5.3's. Its categories are Edit, Write, Shell and Other.
- `internal/workspace`: the home (`BONSAI_HOME`, else `~/.bonsai`), the machine key, `HashLF`, atomic writes and
  `RemoveFile` with Windows' busy retries (2 s), `IsBusy`, and `Find` (the main checkout through a git process; its
  note: not for the guard).
- `internal/engine`: `ownHooks` holds the guard's line only (its note: the other three come as they are built);
  consent (5.1.1): a hook line added or changed needs `--allow-exec`, except Bonsai's own lines at a first link;
  `.bonsai/.gitignore`'s text; `init --new-id` empties `.bonsai/local/`; `check` finds a tracked or staged `local/`
  file and a missing or changed `.gitignore`.
- `cmd/bonsai/hook.go`: `hook start`, `stop` and `record` refuse as not built.
- `formats/` (set 3): `log.schema.json` (`bonsai.log/1`, 29 fields; the README's "A name not invented" leaves the
  binary's two fields out) and `ask.schema.json` (`bonsai.ask/1`, 18 fields, `answer` an object).
- `internal/status`: `local`'s three folders, absolute.

**What 5.1 will have added** (from this plan's 5.1 notes; 5.2's start re-reads each against what landed, and 5.2.0's
run report records any difference that changes a note below):
- 5.1.3, set 4: the `workspace` schema with `generated` (its kinds, `keep_days`, `keep_newest`); `sessions` (the rows
  a reader returns); `error`; `check`; the schema-compare test (only additions pass).
- 5.1.4a: a Go type for every format, with writers for the log and ask records, validated against their schemas;
  `bonsai.yaml` read in full, `generated` among it, beside the guard's lean read of only the fields it judges, which
  the hooks here use too, each reading only what it needs; `check --schema`, printing an open list's known words from
  its Go table.
- 5.1.4b: the `error` object in every command's `--json`, its words in one Go table; every word's `--help` from one
  table of flags and exit codes.
- 5.1.5: the active-task function (contract §13); labels in force (the packs' `declares` and this machine's
  `labels/`); document kinds with their id patterns; the built-in `bonsai.yaml` template `init` writes, one comment of
  which 5.2.6a changes.
- 5.1.6: `check`'s findings and warnings in one table, a test per row; the memory caps; the generated kinds and their
  defaults, in set 4's `workspace` schema and in the Go table its warning on run reports reads, which 5.2.6a extends.
- 5.1.7: `unlink` and taking a pack out, which remove Bonsai's own lines and so must remove 5.2's new ones (5.2.4
  tests both with them).
- 5.1.8: `check --write` rebuilding the tasks table; `init` writing `.bonsai/sessions.md` with its frontmatter and no
  rows.
- 5.1.10: `docs/reference/lists.md`, generated, with its test.

**The pieces.** Hours: the spec gives four figures. 5.2.1 carries the redaction row whole (10-15). The other three
rows, 15-22 together, are spread over seven pieces by the planner's judgment, for sizing briefs, not a spec figure:
the recorder row (9-13) is 5.2.0, 5.2.2, 5.2.4, 5.2.6a and 5.2.6b (low 1+2+3+1+1 = 8, high 2+3+4+1+2 = 12) plus about
an hour for `log append` in 5.2.5; the asks row (4-6) is the rest of 5.2.5; the sessions row (2-3) is 5.2.3. Three
things sit outside their own row: `logs` (the recorder row's) is in 5.2.3, a thin reader over 5.2.2's; set 5's `asks`
schema and error words (the asks row's) are in 5.2.0; and the memory secret scan, handed on by 5.1 with no hours of its
own, counts in 5.2.4. 5.2.4 is the thinnest piece for its load (two hooks, the engine's own lines under consent, two
`check` rows, real sessions): its run report keeps its running hours against its 4, and past 4 the orchestrator says
so before its verifier is briefed. In all: low 1+10+2+2+3+5+1+1 = 25; high 2+15+3+3+4+7+1+2 = 37.

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.2.0 | **Formats set 5 and the log's lists**: `bonsai_path` and `bonsai_sha256` at the end of `bonsai.log/1`; two command outputs, `bonsai.asks/1` and `bonsai.logs/1`; a sessions row's `subagent` field if set 4's rows cannot tell a subagent run from a session; the log's events and categories in one Go table; 5.2's new error words; `asks` and `logs` in `status --json`'s `formats` and in every list of the schemas; the reference page regenerated | The formats test (manifest, docs, examples); the schema-compare test passing on set 5 and failing on a removal in a temporary copy; `check --schema bonsai.log` printing the table's words; `status --json`'s `formats` naming both new formats; the reference page's test; check 10; CI | 1-2 | Contract §2.2, §7.5, §8.1, §8.2, §9.1; spec §3 (`error`), §8, §16 rows 27, 29 and 30; format review 4.1-4.3 and 4.5; `formats/README.md`; this plan's 5.1.3 and 5.1.4b notes |
| 5.2.1 | **The redactor** (`internal/redact`): every rule of the studio's redactor, the three leak classes fixed, a fixed point; and what a tool call is reduced to (command heads in bash and PowerShell, paths, a host, an MCP tool, a subagent type, a skill, an ask's question) | Go tests written fresh by class, with made-up secrets; `FuzzRedact`; time on long inputs; the scripted differential against the studio's redactor at `25b6450` (note 9); the verifier's own differential and break-it | 10-15 | Spec §8, §3 (rules learnt from Windows); contract §2.6, §8.1; "What changes", item 10; read-only, the studio's files at `25b6450`: `tools/lib/redact.mjs`, `tools/lib/redact.test.mjs`, `studio-app/bridge/redact.test.ts`, the three leak rows of `studio/bugs.md`, and the generated shapes in the run report of the change that made that commit (the brief names it) |
| 5.2.2 | **Appends, reads and the salt** (`internal/record`): one append path for every writer of `.bonsai/local/` (whole lines, busy retries); the reader (a file's records, its last line, the files; a torn line skipped and counted); the main checkout's `local/`; `.bonsai/.gitignore` restored by every writer; the home's `salt`; `bonsai --version` with the build's commit stamp | Go tests: eight processes appending at once, on both sides; a Windows file held busy, then released; two first writers racing for the salt; the main-checkout cases (main, a worktree, a `.git` file pointing elsewhere); a torn line; check 10; CI | 2-3 | Contract §2.5, §3, §8.5; spec §3 (where it lives; the tripwire), §6 (`.bonsai/local/`); `internal/guard/record.go`; `internal/workspace` (`write.go`, `home.go`, `checkout.go`); gate report §2.11 |
| 5.2.3 | **The sessions table and `bonsai logs`**: a session's spans and its subagent runs, read from the log; `check --write` adding their rows, the hours below them; the stale-table warning; `bonsai logs` | Go tests on fixture logs written fresh, one per span rule; the table byte-stable on two runs and on both sides; `logs --json` against its schema; check 10; CI | 2-3 | Spec §6 ("The two tables"), §8; contract §7.5, §8.1, §8.5, §13; this plan's 5.1.8 note |
| 5.2.4 | **The recorder**: `bonsai hook record` for ten events and `bonsai hook start` (its opening context; on every start, the `session_start` record with the binary's path and hash); Bonsai's own hook lines in the engine; the input hash; the two `check` rows that wait for them: the memory secret scan, and Bonsai's own lines out of date | Go tests on payloads written fresh for every event; every record against `bonsai.log/1`; no byte on stdout from `hook record`, exit 0 on any input; the engine's consent tests for the new lines, and `unlink` and taking a pack out with them; scripted sessions on WSL, one in a worktree (note 12); the studio's recorded payloads through both recorders, scripted (note 13); the hooks' timings on both sides; the verifier's break-it | 3-4 | Spec §3, §6 (findings; the preview), §7 (the hook lines), §8; contract §5.3, §8, §13; this plan's 5.1.1 rules 1 and 6; Claude Code's hooks reference (the version read goes in the run report); read-only, the studio's `tools/hooks/event-sink.mjs` and `tools/lib/spool.mjs` at `25b6450` (today's event mapping); gate report §2.5, §2.15, §5 |
| 5.2.5 | **Asks, and `log append`**: `bonsai ask` (filing, `--resolve`, `--status`), `bonsai answer`, `bonsai asks`, each ask record with its `ask` log record; `bonsai log append` | Go tests walking a table of every ask case (types, limits, states, the asking session refused, a repeated answer, Bless refused); every free-text field redacted; `--json` against `bonsai.asks/1` and `bonsai.logs/1`; check 10; CI | 5-7 | Spec §4, §8; contract §2.6, §5.2-§5.4, §8.4, §9; format review 4.3; read-only, the studio's `tools/lib/asks.mjs` and `tools/studio/ask.mjs` at `25b6450` (today's flags and limits) |
| 5.2.6a | **The generated-files page**: `docs/reference/generated-files.md`, generated from the generated kinds' Go table (extended with where each kind lives, who writes it, what is never cleaned and when it is cleaned), its test and its `.gitattributes` line; the comment on `generated:` in `init`'s built-in `bonsai.yaml` naming it | The page's test, failing on a changed default in a temporary copy; `go generate` changing nothing; check 10; CI | 1-1 | Spec §6 ("Generated files", "Every list has one home"); format review R2.6; this plan's 5.1.10 note |
| 5.2.6b | **Cleaning per kind**: `log`, `asks`, `ladder` and the sessions rows by `generated:`'s `keep_days` and `keep_newest`, the protections first; a `clean` record per file or row; at a session's end, in `check --write`, and a call 5.4's runner makes | A fixture project with old, new, protected and decoy files of every kind: exactly the right ones go; check 10; CI | 1-2 | Spec §6 ("Generated files"); contract §7.5, §8.2, §8.5, §9.1, §11 |
| **5.2** | | | **25-37** (re-ask 48) | |

**The order, side by side where truly independent.** Rohan, 9 Oct: "if you can orchestrate work in parallel do that
whenever possible"; his 8 Oct bar stands: no shared file, and neither's proof resting on the other's. 5.2 starts once
5.1's end verifier has passed 5.1 (no two parts run at once). A piece that runs beside another rebases on `main` and
re-runs its proof if the other lands first ("What changes", item 2). The files each piece owns:

| Piece | Owns |
|---|---|
| 5.2.0 | `formats/` (the two new schemas, `log.schema.json`, `sessions` if needed, examples, README, manifest); the log type's Go file from 5.1.4a, with the table of events and categories; the error words' Go table; wherever 5.1.4a registers an open list's table for `--schema`, the test that lists every schema, and `status --json`'s `formats` list (all three gain `asks` and `logs`); `docs/reference/lists.md` |
| 5.2.1 | `internal/redact/` only |
| 5.2.2 | `internal/record/` (new: the append path, the reader, and the one builder of every record's common fields); in `internal/workspace/`, the salt and the main checkout's `local/`; the `.gitignore` text, moved from `internal/engine/plan.go` to its one home; `cmd/bonsai/main.go` for `--version` only |
| 5.2.3 | `internal/sessions/` (new); 5.1.8's `check --write` code; 5.1.6's table of findings and warnings (one warning); the `logs` word in `cmd/bonsai` (its dispatch, its entry in the flag table, `logs.go`) |
| 5.2.4 | `internal/recorder/` (new); `cmd/bonsai/hook.go` and the hook's entry in the flag table; `internal/engine/settings.go` and the engine tests that hold today's lines; 5.1.6's table (one finding, one warning) |
| 5.2.5 | `internal/asks/` (new); the `ask`, `answer`, `asks` and `log` words in `cmd/bonsai` (dispatch, flag-table entries, `ask.go`, `log.go`) |
| 5.2.6a | The generated kinds' Go table (and `docs/reference/lists.md` if its list of kinds changes); `docs/reference/generated-files.md`, its generator and test; one `.gitattributes` line; the comment on `generated:` in `init`'s built-in `bonsai.yaml` |
| 5.2.6b | `internal/clean/` (new); its calls in `internal/recorder/` and in the sessions part of `check --write` |

1. **5.2.0, 5.2.1 and 5.2.2 start together, side by side, once the orchestrator has checked their files.** 5.2.0 sets
   every shared list and schema first, so no later piece edits a shared list file; 5.2.1 is the longest and riskiest
   piece, so it starts at once; 5.2.2 is the base every writer stands on. By the table above they share no file, but
   5.2.0's row names files 5.1.4a and 5.1.4b have not landed yet (where `--schema` registers an open list's table, the
   test that lists every schema, `status --json`'s `formats`): at 5.2's start the orchestrator reads where they
   landed, and if any is a file 5.2.2 also touches, 5.2.2 starts after 5.2.0 lands; the run report says which. None's
   proof rests on another's: the formats test and the schema compare need no redactor and no appender; the redactor
   is text in, text out; the appender takes bytes, and the reader keeps a field it does not know (contract §2.2), so
   neither needs set 5.
2. **5.2.3 after 5.2.0 and 5.2.2 have landed, beside the rest of 5.2.1.** It reads the log through 5.2.2's reader and
   writes set 5's `sessions` row and `logs` output; it redacts nothing (records are redacted when written), so its
   proof does not rest on 5.2.1, and 5.2.1 touches only `internal/redact/`.
3. **5.2.6a after 5.2.0 has landed, beside 5.2.1 and 5.2.3.** It writes the page from the generated kinds' table and
   changes one comment in the engine's built-in `bonsai.yaml` (5.1.5's file, not `plan.go`, which 5.2.2 changes): no
   piece of that wave touches either, and its test rests on its own table only. It waits for 5.2.0 only because 5.2.0
   owns the reference page, which a change to the list of kinds would regenerate. If the orchestrator finds the
   kinds' table in a file 5.2.0 or 5.2.3 touches, 5.2.6a runs after them instead.
4. **5.2.4 after 5.2.1 and 5.2.3 have landed.** It redacts targets and text (5.2.1), writes 5.2.0's fields through
   5.2.2's append path and salt, and adds two rows to the `check` table 5.2.3 also changes.
5. **5.2.5 after 5.2.1 and 5.2.3 have landed; beside 5.2.4 when the flag table allows it.** Asks redact every string
   (5.2.1) and add words in `cmd/bonsai`, as 5.2.3 does. 5.2.4 and 5.2.5 share no proof: an ask needs no recorder,
   and the recorder writes no ask. They share no file if 5.1.4b's flag table keeps each word's entry in a file of its
   own; if it is one file, 5.2.5 starts after 5.2.4 lands. The orchestrator decides from what 5.1.4b landed, and the
   run reports say which.
6. **5.2.6b last.** It protects what 5.2.3 (a span still open, a session with no row yet), 5.2.4 (the session end it
   runs at) and 5.2.5 (an open ask) define, it calls into their files, and it reads 5.2.6a's table for the defaults.

**Who builds and verifies.** Opus builders for 5.2.1 (redaction), 5.2.2 (appends under concurrency on Windows, the
salt), 5.2.4 (hooks every session runs, and the engine's own lines under consent), 5.2.5 (its refusals, and every string
redacted) and 5.2.6b (it deletes files); Sonnet for 5.2.0, 5.2.3 and 5.2.6a, whose shapes and rules this section fixes,
with the orchestrator's read. Fresh Opus verifiers: **5.2.1** (redaction is security: what may leave the machine), with
its own differential and break-it; **5.2.4** (guards and hooks: every session in a linked project runs these lines, and
they change what `update` writes). 5.2.0, 5.2.2, 5.2.3, 5.2.5, 5.2.6a and 5.2.6b land on green tests on both sides, CI
and the orchestrator's read of the diff, which the run report says; **the 5.2 end verifier** covers them, and re-runs
5.2.1's differential and 5.2.4's sessions on the final build. This section's planning and review runs count in 5.2's
hours, carried in 5.2.0's run report ("What changes", item 3).

#### Where each inherited finding is settled

The gate report's section 5, its 5.2 list, item by item; then the outline's "Settles" and what 5.1 and the code hand
on:

| Finding | Settled in | How |
|---|---|---|
| The log's `bonsai_path` and `bonsai_sha256` are Bonsai's own names, outside the log schema (gate §5) | 5.2.0 | The guard's two names added at the end of `bonsai.log/1`: an addition, set 5 with its manifest; the README's "A name not invented" becomes a choice made |
| `input_hash` is null until the salt (gate §5) | 5.2.2, 5.2.4 | The home's `salt` made at first need, and never leaving the machine (5.2.2); the recorder fills `input_hash` on its four tool events (5.2.4). The guard's records keep null until 5.3 changes guard code |
| The binary's hash is logged once per session file (gate §5) | 5.2.4 | `hook start` hashes the binary on every start (`startup`, `resume`, `clear`, `compact`; about 3-5 ms each) and writes the path and hash on its `session_start`, so a binary swapped between a session and its resume is seen. It also makes the session's file first, so in a main checkout's session the guard finds the file made and hashes nothing, and the self-hash leaves the session's first tool call (in a worktree's session the guard still hashes once in its own folder until 5.3 moves its records; 5.3 measures the guard's first call again) |
| Format review 4.2's question to Rohan on those two names, "Want to see them before building?", still open | 5.2.0; the orchestrator | The names go to him as a look, no vote, in the one line that sends him this section; a change he wants is made before 5.2.0 lands |
| Builds without a commit stamp (gate §5; "What changes", item 7) | 5.2.2 | `bonsai --version` names the build's commit or says it has none; the builder finds why worktree builds lack it and writes the build line every scripted run then uses; the end verifier ties a logged hash to its commit |
| The secret scan of memory notes, handed on by 5.1.6 | 5.2.4 | A `check` finding on the redactor's patterns (`redact.Find`), naming the note and line, never the value |
| The log's events, an open list read by code at many places (outline; 5.1.3's note) | 5.2.0 | Their known words, and the categories', in one Go table, printed by `check --schema bonsai.log` and the reference page; the schema's descriptions name the command and copy no word |
| `asks` and `logs` get `--json` schemas as additions (outline) | 5.2.0 | `bonsai.asks/1` and `bonsai.logs/1`, new schema files, each documented with an example |
| §8's three older leaks, named by studio bug ids (this plan, "Stale or in tension") | 5.2.1 | Fixed as classes; code and tests describe their shapes, never the ids |
| The generated-files page is a `base` skill, and `base` comes in 5.5 (this plan, "Stale or in tension") | 5.2.6a | `docs/reference/generated-files.md` meanwhile, generated from the code and tested; 5.5 moves its words into the skill |
| "5.2 fills the sessions table from the log and adds it to `--write`" (5.1.8's note) | 5.2.3 | Rows from the log's spans, only added, byte-stable; the stale warning |
| The other three of Bonsai's own lines come "as they are built" (`internal/engine/settings.go`) | 5.2.4 | `hook start` and `hook record` lines added; a project linked before takes them at `update --allow-exec --yes`. The stop gate's line stays 5.3's |
| `hook start`, `stop` and `record` refuse as not built (`cmd/bonsai/hook.go`) | 5.2.4 | `start` and `record` built; `stop` stays 5.3's |

#### Notes per piece

**5.2.0, formats set 5 and the log's lists.** One commit to `formats/` with its manifest at the set after 5.1's last
(set 5 here, 5.1.3's being set 4), as `formats/README.md`'s "How the set changes" asks; additions only, so the
schema-compare test passes. Its builder first reads set 4 and 5.1.4b's error words as they landed.
- **The binary's two fields** (spec §16 row 27; format review 4.2 left the names open and asked Rohan whether he
  wants to see them first: they go to him as a look, the (B) paragraph above): `bonsai_path` and `bonsai_sha256`, at
  the end of `bonsai.log/1` after `remote`: the names the guard has written since part 5, so the records already
  written stay valid. Chosen over new names, which would leave those records with two unknown fields. `bonsai_path` is
  the binary's own path with forward slashes: from `~/` when it lies under the person's home folder, so no user name
  is in it (contract §2.6), else as it is (`/usr/local/bin/bonsai`, `C:/Program Files/Bonsai/bonsai.exe`); or null.
  `hook start` writes it on every `session_start` (5.2.4); the guard writes it on its records in its own absolute form
  until 5.3 changes guard code. `bonsai_sha256` is 64 lower-case hex: on every `session_start` (5.2.4 hashes the binary
  at every start) and on the guard record that made its log file; else null.
- **The log's lists in one Go table**, beside the log's Go type: the events (contract §8.2: the eleven agent events,
  `guard`, `ladder`, `ask`, `event`, `clean`) and the categories (`Read`, `Search`, `Edit`, `Write`, `Shell`,
  `Ladder`, `MCP`, `Agent`, `Web`, `Other`; a pack's own as `<namespace>.<Name>`), each word with one line on what it
  means. The schema's `event` and `category` descriptions then name `bonsai check --schema bonsai.log` and the
  reference page and copy no word (5.1.3's rule).
- **Every description that changes with what Bonsai now writes** (a description change is an addition): `event` and
  `category` (above); `text`: a notice's message or an ask's question, never a prompt's words (5.2.4, note 5); `kind`:
  a prompt's `user` or `task-notification`, a failure's `interrupt` or `error`, a notice's type, an `ask` record's op
  (5.2.5); `task` and `role`: raw from the environment, and redacted like every free-text field; `labels`: checked
  against the labels in force, not the machine's alone (5.2.5); `target`: a path outside the checkout is null, an
  `ask` record's is its key, a `clean` record's the path it deleted; `bonsai_path` and `bonsai_sha256` as above.
- **`asks` and `logs` wherever the formats are listed:** where 5.1.4a registers the schemas for `--schema`, the test
  that lists every schema, and `status --json`'s `formats` (`read` `[1]`, `write` 1 for each), all in this piece, so
  5.2.3 and 5.2.5 add none.
- **`bonsai.asks/1`**, the `--json` of `ask`, `ask --resolve`, `ask --status`, `answer` and `asks`: `format`;
  `workspace` (null when the command refused before reading one); `asks`, a list of `{key, state, filed, closed}`,
  where `state` is `open`, `answered` or `resolved` (a closed list), `filed` the key's latest `file` record, and
  `closed` the `answer` or `resolve` record that closed it, or null, both `bonsai.ask/1` records as stored; `written`,
  the record this run appended, or null; `error`.
- **`bonsai.logs/1`**, the `--json` of `logs` and `log append`: `format`; `workspace`; `files`, a list or null, each
  `{file, session, day, first, last, records, unreadable, ended, task, role}` (a day file has `session` null and `day`
  its date; `ended`, `task` and `role` are a session's, else null); `records`, one file's records as written, or
  null; `written`, the record `log append` appended, or null; `error`.
- **A sessions row tells a subagent run from a session** (contract §7.5: the two are "shown apart, never added"). If
  set 4's row cannot, `subagent` is added at the end of a row: the subagent run's own id, its first 8 characters, or
  null for a session's row (`-` in the table). It also keeps two subagent runs of one session apart in the row's key
  (5.2.3).
- **The error words 5.2 needs**, added here so no later piece edits the words' table: `ask-not-open` (no such key, or
  it is answered or resolved; the message says which), `answer-own-session` (the session that asked answers it),
  `label-not-defined` (`log append` names a label no definition in force has, or gives a value of the wrong kind),
  `session-not-found` (`logs --session` matches no session, or several). Everything else reuses 5.1.4b's words (bad
  input, not linked, could not write), mapped by the builder from the table as it landed. A word a later piece finds
  missing is added there, and a piece beside it rebases and regenerates the reference page.
- An example for each new schema, and the log's gains its two fields; `docs/reference/lists.md` regenerated.

**5.2.1, the redactor.** Spec §8: "Redaction in Go, as written (decision 2), over command lines, prompts, notices and
question text"; "Proof is differential: on the shared corpus (today's redaction tests plus those three bugs' cases),
the Go redactor hides everything the Node one hides, and the three bugs' cases too."
1. **Not a line-for-line port.** The studio's redactor is regular expressions leaning on lookbehind and lookahead,
   which Go's `regexp` (RE2) does not have, and Bonsai takes no other regular-expression library (spec §3: the
   standard library and `golang.org/x/sys`). Its leaks share one cause: one rule's value swallows the next rule's name
   before that rule runs, which the studio patched shape by shape. Bonsai's redactor is a scanner that **finds every
   name first, then each name's value**. The names are the studio's: a secret-named key before `:` or `=` (its keyword
   list at `25b6450`, inside a run of name characters such as `DB_TOKEN` or `x-api-key`); a flag ending in such a word
   (`--password`, `--client-secret`), starting a word, its value on the flag's line; an Authorization header of any
   scheme; a Bearer value; `extraheader=`.
   - **A value keeps today's shapes:** a quoted value ends at its closing quote on its line and takes what is glued
     after the quote, and a quote never closed runs to the line's end; a value may sit on the next line, with the
     studio's exceptions (not when that line starts with a name on its own line; not a lone word on the string's last
     line); a key's bare value stops at `,` `;` `&` `}` `)`.
   - **A name that stands where another name's value starts, or inside that value, is part of that value:** its own
     value is redacted first, then the outer value takes the inner name too, one marker over both, as the studio's
     second sweep writes it. `a bearer password: "hunter2"` becomes `a bearer [redacted]: [redacted]`; `a bearer
     db_password=hunter2 ok` becomes `a bearer [redacted] ok`; `password: mysecret:123` keeps nothing of its value. No
     name's value is left because another rule reached it first, and what one pass writes a second pass leaves
     (note 5).
   - **Then every rule that takes a whole value in the studio's redactor at `25b6450`**: among them private-key
     blocks, webhook URLs, credentials inside a URL, the known token shapes (Anthropic, OpenAI, GitHub, Slack,
     Stripe-style live and test keys, npm, AWS, Google, JWT), and the long random run with its exceptions as they
     stand there (pure hex, `toolu_` ids, CamelCase names). The list here is examples; the studio's file is the
     measure.

   Chosen over a port, whose bugs would port too, and over translating the expressions, which RE2 cannot express.
2. **The three leak classes**, by their shapes (spec §8 names them by the studio's ids; Bonsai's code and tests never
   do):
   - **After an Authorization header's scheme word**, a quoted token is a value like any other (`Authorization:
     Bearer "..."`, the same after `token` and `Basic`), and a `token` followed by `:` or `=` is a name, not the
     scheme, so its value goes too; also after `=`, a tab, `authorization=`, `Proxy-Authorization:` and
     `x-authorization:`.
   - **A second name not at the very start of a value**: behind punctuation (`{`, `(`, `[`, `<`, `*`, `|`, `>`, a
     backtick, curly quotes, guillemets, an inverted question mark), one character in, behind another
     `extraHeader=`, or with any whitespace before its `:` or `=` (a vertical tab, a form feed, a no-break space).
     Its value goes.
   - **The shapes the studio's last change opened**: a value-taking name, a second name, then a third behind
     punctuation; a value-taking name that takes `extraheader=` as its value; an open quote before a name at a line's
     end, which swapped which value leaked; and the two shapes where a second pass changed the output.
   With every name found first, and every name inside another's value taken into it, the three are one rule: every
   name's value is redacted, wherever the name stands.
3. **Whitespace and case.** Between a name and its `:` or `=`, any character JavaScript's `\s` matches (U+FEFF and
   U+2028 among them). A keyword matches in any case under Unicode's simple case folding, so the long s (U+017F) and
   the Kelvin sign count as `s` and `k`. JavaScript's `i` flag without `u` folds neither, so Bonsai hides more there,
   and the differential reports it as an over-redaction class. Chosen so no Unicode form slips past. Caps count
   characters (code points), as JSON Schema's `maxLength` does.
4. **What it keeps**, as today: git SHAs and other pure hex, UUIDs, GUIDs, `toolu_` ids, CamelCase test names, prose
   saying "token" or "password" with no value after it, words ending in `-secret` or `-token`, a flag at a line's end
   with a name below it, text already redacted. The marker stays `[redacted]`, which the studio's bridge and Desk
   know.
5. **A fixed point, also cut.** Redacting its own output changes nothing, and neither does redacting that output cut
   at any length: the recorder cuts a field to its cap after redacting, and the studio's bridge redacts again
   (contract §2.6). A cut tail that is the start of the marker, or of a scheme word, is kept, as today.
6. **Linear time.** No input makes it slow: tests on 26 KB of secret words glued by `_` and on 1 MB of mixed text,
   each under a limit far above what it takes and far below what a quadratic rule would; `FuzzRedact` (no panic, a
   fixed point, a planted secret never kept).
7. **The second job: what a tool call is reduced to** (today's `targetOf`, `commandHead` and `questionText`), here
   because it is the first line of defence. A shell command is kept only as its head: the program, plus one
   subcommand word for `git`, `npm`, `dotnet`, `unity`, `claude`, `gh`, `go` and `bonsai`; a PowerShell cmdlet as
   written; so `TOKEN=... cmd`, `git -c http.extraHeader=...` and `node -e "..."` never reach the record. A file or
   search path inside the checkout is kept relative to it, and any other is null, as contract §2.6 asks (records carry
   workspace-relative paths) and as the guard already writes. A web fetch
   keeps its host, an MCP call `server.tool`, the Agent tool its subagent type, a skill its name, AskUserQuestion its
   questions joined (the ask itself, at most 300). Every one passes the redactor and its cap. A Windows path is read
   with backslashes and, on a drive-letter root, case-blind, as today. Categories stay the log's (5.2.0's table;
   5.2.4 maps them).
8. **One home for the patterns.** `internal/redact` gives `Text` (redact a string), `Find` (the spans it would redact,
   each with its rule's kind, for the memory scan) and the reductions. No other package holds a secret pattern.
9. **How the studio's corpus is used, without copying it in.** The differential is a scripted run, never committed,
   as 5.1.2's today's-files run was. The studio's files are taken at `25b6450` (the commit on its `main` that holds
   the redactor and its tests as they still are, the three leak rows, and the run report with their generated shapes)
   with `git show` into a scratch folder, or from a scratch clone with its origin removed ("What changes", item 10).
   The scripts live in `~/bonsai-checks/scripts/`; Node runs the frozen redactor, and a scratch Go driver inside the
   worktree's module runs Bonsai's (never committed, as 5.1.2's verifier's driver was). The corpus:
   - (a) the studio's test tables: secret rows, keep rows, the rows of each earlier fix, the grid of value-taking
     names before second names, cut tails, bash and PowerShell heads, targets, question text;
   - (b) the bridge test's spool lines with planted secrets;
   - (c) the three classes: the leak rows' examples and the generated shapes (two framings, about 4,200 strings),
     made again by the script, and Bonsai's own grid over names, punctuation, whitespace, quotes and case folds, a
     distinct made-up secret in every value slot;
   - (d) real text: every text file of `studio/`, `docs/` and `studio-app/test/fixtures/` in the scratch clone at
     `25b6450` (the set the studio's own corpus test reads), as it is, and with each secret row planted on its own
     line, at a line's end, and after a line ending in a name still waiting for its value;
   - (e) about a million generated strings mixing names, separators, quotes, whitespace and values, a distinct
     made-up secret in every value slot, and the fuzz finds.

   **It passes when**, over every string: no labelled or planted secret leaks from Bonsai's output, a secret counting as
   leaked when any word of it, or any run of 6 or more of its characters, survives (a word is a maximal run of letters
   and digits; the studio's misses on the three classes are counted apart); for every word the studio's redactor took
   out, Bonsai's output keeps no more copies of it than the studio's does; every keep row comes back unchanged; Bonsai's
   output is a fixed point at every cut; no string takes over 100 ms, or over 1 ms a KB for a document past 100 KB.
   Judged, not counted: the strings where Bonsai takes out a word the studio's keeps, grouped by the rule that fired,
   each group read by the builder and then the verifier (a secret shape the studio missed is right; ordinary words taken
   out are a failure to fix). Reported only: how often the studio's redactor changes Bonsai's output (the bridge's
   second pass). The run report gives counts and classes, never a string from the studio's files ((a), (b) and (d) hold
   studio paths and names); strings Bonsai's own script made may be quoted.

   **Bonsai's committed tests are its own:** every row written fresh by class, with made-up secrets (`hunter2`, AWS's
   documented example key, random strings made by the test); none is copied from the studio's tables, which the
   differential covers instead. No studio path, project name, task or bug id, or person's name enters the repo; the
   verifier greps for them. Chosen over a scrubbed copy of the corpus, or of the rows that look clean, since Bonsai's
   public repo holds no studio file (spec §12) and this section sends no studio content public; and over the
   differential alone, which CI could not re-run.
10. **Windows:** the redactor is string logic, so the differential runs in WSL only; its Go tests run natively on
    Windows with check 10.

**5.2.2, appends, reads and the salt.**
1. **One append path** (`internal/record`) for every Bonsai writer of `.bonsai/local/`: the recorder, `hook start`,
   `log append`, asks and `clean` records; the guard keeps its own writer until 5.3 changes guard code. A line is
   opened with `O_APPEND` (with `O_CREATE|O_EXCL` first, so the writer knows when it made the file) and written in one
   `Write` of the whole line with its LF, never longer than its format's cap (2,048 bytes for the log, 8,192 for
   asks); a busy error on Windows is retried for up to 1 s with a doubling wait, as the guard's `openRetry` does. On
   Linux an `O_APPEND` write of a line this size lands whole; on Windows Go opens an `O_APPEND` file for appending
   only, so each write lands whole at the end too. The builder shows both by test, not by reading. Chosen over a lock
   file, which is slower on every hook and leaves a stale lock after a crash. Not tested, and said so: a project on a
   Windows drive mounted in WSL (`/mnt/c/...`), written through WSL's bridge to that drive.
2. **The proof under concurrency:** the test binary started again as 8 child processes, each appending 250 records to
   one file; then every line parses, every id is unique and there are 2,000; on WSL and natively on Windows, in
   `t.TempDir()`. On Windows a second test holds the file open with no write sharing until the first refusal, then
   lets go, as `1170c92`'s test holds a rename's target: the append succeeds after its retry. No test needs a symbolic
   link or a file mode.
3. **The main checkout's `local/`** (contract §3: worktrees use main's). Found by reading `.git` with no git process,
   so the hooks stay fast: a folder means this checkout is main; a file names its `gitdir`, whose `commondir` file
   names the common folder; when `gitdir` lies under `<common>/worktrees/`, main is the common folder's parent. That
   main is used only when its `bonsai.yaml` holds the same workspace id; else the checkout's own folder is used. No
   `commondir` (a submodule, or a checkout with a separate git folder) means the checkout is its own main. Paths are
   compared case-blind on Windows. A test holds the finder's answer equal to `workspace.Find`'s (git's own) on a main
   checkout, a worktree, a submodule and a separate git folder. An agent can rewrite a worktree's `.git` file, so at
   worst a record lands in another checkout of the same project, never in a folder outside one: a tripwire, as the log
   is (spec §7). Spec §6 names `git rev-parse --git-common-dir`; reading the files it reads, with no process, keeps the
   hooks fast ("Stale or in tension"). This finder is the recorder's, the asks' and the cleaner's; the guard keeps its
   own folder until 5.3 decides what the hook path may trust, and may then take this one over.
4. **Every writer restores a missing `.bonsai/.gitignore`** (spec §6), from the text's one home, moved out of
   `internal/engine/plan.go`.
5. **The salt** (contract §3: "the input-hash key; never leaves the machine"): `<home>/salt`, 64 lower-case hex
   characters from 32 random bytes, made at first need by writing a whole temporary file beside it and hard-linking
   it into place, so two first writers cannot both win and no reader sees half a file; the loser reads the winner's,
   as the studio's sink does today. Mode 0600 on Linux (no test needs it on Windows). If it cannot be made or read,
   `input_hash` is null and the record is still written. Base's walls deny reading it (spec §7; 5.5).
6. **The reader:** a file's records in order, each through 5.1.4a's log type; a line that does not parse (a torn
   last line after a crash, or a line written by hand) is skipped and counted, never fatal; a file's last record
   without reading the file whole; a folder's session files and day files by their names, nothing else.
7. **`bonsai --version` names its commit:** `bonsai <version> (commit <12 hex>)`, `+modified` when the tree was dirty,
   or `(no commit stamp)`, from the build's own `vcs.revision`. The builder finds why the builds made in the WSL
   worktrees carry no stamp (gate report §2.11) and writes in the run report the build line every scripted run then
   uses: `go build -buildvcs=true`, which refuses rather than build without a stamp, or a build from a clean clone at
   the commit ("What changes", item 7).
8. **One builder of every record's common fields**, in this package, used by the recorder, `log append`, the `ask`
   records and the `clean` records, so each writes the schema's order the same way: `format`, `id`, `at`,
   `workspace`, `session`, `agent`, `checkout`, `branch`, `task`, `role`, `labels`, `remote`, and the rest null unless
   the caller fills them. `checkout` is the folder name of the checkout the writer runs in (a hook's project, a
   command's working folder; for a `clean` record at a session's end, that session's checkout, and from `check
   --write`, the main checkout); `branch` is that checkout's, read from its HEAD through the same `.git` reading,
   never a git process, on every record; `task` and `role` are `BONSAI_TASK` and `BONSAI_ROLE`, redacted, when set,
   else null. On a hook's record `session` and `agent` come from the payload; on an `ask` record, from
   `CLAUDE_CODE_SESSION_ID` (`agent` `claude-code` when it is set), else both null; on a `log append` record and a
   `clean` record both are null (contract §8.4; a clean is Bonsai's own act).

**5.2.3, the sessions table and `bonsai logs`.** Spec §6 and contract §7.5 fix the table's columns; these are the rules
they leave open.
1. **Spans.** In a session's file, a `session_start` opens a span unless its `source` is `compact` (a compaction is
   the same session going on); a `session_end` closes it. A start while a span is open (a lost end line) closes the
   open one at its last record. A `subagent_start` opens a subagent run, closed by the `subagent_stop` with its
   `subagent_id`; one with no stop ends with its session's span. A span with no end whose file's last line is more
   than 24 hours old ends at its last record (a killed or crashed session); otherwise it is **open**: no row yet, and
   its file is never cleaned. One function says "open", for the table and for the cleaner (5.2.6b). Chosen over
   waiting for an end line forever, which would keep a crashed session's file and lose its hours.
2. **A row:** the session id's first 8 characters; the task its start record found (`target`), else `none`; the role
   (`role` for a session, `subagent_type` for a subagent run); the model (`-` for a subagent run unless its records
   name one); start and end in UTC as `YYYY-MM-DD HH:MM`; minutes as end minus start as shown, so a person can check a
   row by eye; and `subagent` (5.2.0). A row's key is its id, `subagent` and start. A key already in the table is
   never added again or rewritten, and a row whose log file is gone stays (rows are only added, spec §6). Rows are
   sorted by start, then key.
3. **Hours** below the rows: for each task and role, session hours and subagent hours in two columns, never added
   together, to one decimal; then each task's total of each. A row whose task is `none` counts under `none`.
4. **Written only by `check --write`**, in the main checkout only (5.1.8). A table that does not read back (a hand
   edit that broke it) makes `check --write` refuse with exit 3, naming the line, and never drop a row. The stale
   warning (spec §6): an ended span in the log with no row; a warning everywhere, never a finding.
5. **`bonsai logs [--session S] [--day D] [--json]`.** With no filter, one line per file, newest first: a session's
   id, first and last time, ended or open, its task and role, records and unreadable lines; a day file's date and
   records. `--session` takes a full id or a prefix of at least 8 characters (the table's), and exits 4
   (`session-not-found`) when it matches none or several, naming them; `--day` reads that UTC day's file (outside
   events, `clean` records). `--day` is a flag, not a word. `logs` reads main's `local/` through 5.2.2's finder,
   writes nothing, and prints records as stored: they were redacted when written.

**5.2.4, the recorder.**
1. **Bonsai's own lines** (spec §7's table, with the departures below), added to `ownHooks`, each with its preview
   sentence:

   | Event | Line | How |
   |---|---|---|
   | SessionStart (every source) | `bonsai hook start` | synchronous, timeout 10; prints the opening context and writes the `session_start` record, with the binary's path and hash, every time |
   | UserPromptSubmit, PreToolUse (every tool), PermissionRequest, PostToolUse, PostToolUseFailure, Notification, SubagentStart, SubagentStop, Stop | `bonsai hook record` | `async`, timeout 10 |
   | SessionEnd | `bonsai hook record` | synchronous, timeout 5, then cleaning (5.2.6b) |

   All eleven events are recorded. SessionStart's record is written by `hook start` itself, so one process makes the
   session's file first, with the binary's path and hash, before any tool call; chosen over a second line on
   SessionStart, which would race `hook start` to make the file and write two records for one event. SessionEnd's line
   is synchronous because an async hook is killed as the session exits, so its line would be lost: the studio measured
   that for its own recorder, which registers SessionEnd the same way. SessionStart carries no matcher: after `/clear`
   the agent has lost its opening context as after `/compact`, so `clear` counts with the spec's `startup`, `resume`
   and `compact`. No line ends in `|| exit 2`: none of them blocks.
2. **Consent.** A first link writes these lines on `--yes` (5.1.1 rule 6). A project linked before 5.2 gets them at
   its next `update` only with `--allow-exec` and `--yes` (rule 1: a hook line added). `check` warns, never finds,
   when this build's own lines differ from the project's ("run `bonsai update --allow-exec --yes`"). `unlink` and
   taking a pack out (5.1.7) remove the new lines with the guard's: 5.2.4 tests both with them in place. Every engine
   test and scripted check that holds today's lines is updated, and the run report lists each.
3. **Which project.** `CLAUDE_PROJECT_DIR`, else the payload's `cwd`; never a later `cd`. The project is found as the
   guard finds it; in a folder with no `bonsai.yaml` both hooks exit 0 at once and write nothing (spec §3, §7). Each
   reads `bonsai.yaml` through 5.1.4a's lean read, taking only what it needs (the id, the task folder for the active
   task, and at a session's end `generated`). The records go to the main checkout's `local/` (5.2.2).
4. **What each event fills**, beside the fields every record has (`format`, `id`, `at`, `workspace`, `session`,
   `agent` as `claude-code`, `agent_event`, the subagent's id and type when the payload names them, `checkout` as the
   folder's name, `branch`, `task` and `role` from `BONSAI_TASK` and `BONSAI_ROLE`, `labels` as `{}`, and `remote`
   when `CLAUDE_CODE_BRIDGE_SESSION_ID` matches its pattern, never redacted):

   | Claude Code event | `event` | Also filled |
   |---|---|---|
   | SessionStart (`hook start`) | `session_start` | `source`, `model`, `target` (the active task found, contract §13), `bonsai_path`, and `bonsai_sha256` when it made the file |
   | UserPromptSubmit | `prompt` | `kind`: `user` or `task-notification`; no words |
   | PreToolUse, PermissionRequest | `tool_start`, `permission` | `tool`, `category`, `target`, `tool_use_id`, `input_hash`; for AskUserQuestion, `text` as its questions |
   | PostToolUse, PostToolUseFailure | `tool_end`, `tool_fail` | the same, and `ok` true or false; a failure's `kind` `interrupt` or `error` |
   | Notification | `notice` | `kind` as its type, `text` as its message |
   | SubagentStart, SubagentStop | `subagent_start`, `subagent_stop` | on the start, `target` as the active task found |
   | Stop | `stop` | nothing more |
   | SessionEnd | `session_end` | `reason` |

   Categories: Read; Grep and Glob as Search; Edit, MultiEdit and NotebookEdit as Edit; Write; Bash and PowerShell as
   Shell, or Ladder when the head is `bonsai ladder`; Agent and Task as Agent; WebFetch and WebSearch as Web; `mcp__`
   tools as MCP; anything else Other. The studio's `Unity CLI` becomes a pack's category later (contract §8.2).
5. **No prompt's words** are kept: a `prompt` record has its `kind` only, as today's sink writes it. Chosen because the
   bridge forwards `text` (contract §2.6) and nothing today shows a prompt's words; keeping them would be a privacy
   change nobody asked for. What lost: the "prompt's start" the schema's description allowed (5.2.0 rewrites it).
6. **Every free-text string passes the redactor, then its cap** (`target` 200, `text` 300), then the record's 2,048
   bytes, by halving the longest text, as the guard does: `target`, `text`, `task`, `role`, `subagent_type`, `model`,
   `source`, `reason`, `kind`, `branch` and `checkout`. Ids (`session`, `subagent_id`, `tool_use_id`) are held to
   their patterns instead, and `remote` to its own; `bonsai_path` is the binary's own path, from `~/` under the home
   (5.2.0), and is not redacted.
7. **The input hash** (contract §8.1): HMAC-SHA-256 keyed by the salt's text, over the tool input with its object
   keys sorted at every depth and the strict decoder's numbers as their source text; the first 16 hex characters.
   Bonsai's own canonical form, not byte-equal to the studio's: the two never share a salt, so no hash is compared
   across them, and matching JavaScript's way of printing numbers would be work for nothing. The tool input is the
   same bytes in a call's three events (the studio measured it), so a call's records pair.
8. **`hook record` never interferes:** nothing on stdout, ever (an async hook's stdout can reach the conversation);
   exit 0 whatever happens; a payload it cannot read writes nothing. Its payload reader is the recorder's own for now:
   spec §14 gives "the hook adapter" to 5.3, which then makes one reader for the guard and the recorder.
9. **`hook start`'s opening context** (spec §8), in plain ASCII on stdout, which Claude Code adds to the session: the
   workspace's name and id; the active task with its status, lane, branch and grants (`bonsai.allows`), or "none"
   with the function's reason; that task's last ladder result (green or not, its commit, when), or "none yet" (5.4
   writes them); the label definitions attached on this machine (contract §5.3). At most 60 lines; past that, one line
   naming `bonsai status --json`. It prints what it can and exits 0: an unreadable `bonsai.yaml` gives one line saying
   so, and that the guard blocks edits until a person fixes it. The same on every source, so `/compact` and `/clear`
   bring it back.
10. **Time:** `hook start` and `hook record`, p50 and p95, on WSL and Windows with part 5's harness, in the run report.
    `hook start` hashes the binary at every start (about 3-5 ms); `hook record` reads a PostToolUse payload whole,
    tool output included, so large payloads are timed too.
11. **The memory secret scan** (spec §6; handed on by 5.1.6): a `check` finding for a memory note (the notes and index
    in the folder `documents.memory` names, in the working tree) holding anything `redact.Find` finds. It names the
    note, the line and the kind of secret, never the value; its next step: take it out of the note, and a person
    decides whether to rotate it.
12. **Real sessions, scripted** (part 4b's method: `claude-here` in a scratch project linked by this build and
    holding one running task; the user settings hashes before and after; the Claude Code version recorded): on WSL,
    `-p` sessions that read, edit, search, run a shell command holding a made-up secret, start a subagent and end; one
    resumed; and one started in a worktree of the scratch project (spec §8: worktrees write main's), whose recorder
    lines land in the main checkout's `local/` and whose guard records in the worktree's own folder (documented until
    5.3), and which `logs` and `check --write` in the main checkout then show. The log is read back against the table
    above; the made-up secret is in no record; `hook start`'s context reached the session (the session is asked to
    quote its active task). Three Claude Code facts are measured and recorded: SessionEnd's real time limit, which the
    1 s cleaning budget must fit (5.2.6b), and what a cleaning killed between a delete and its record leaves; whether
    `CLAUDE_CODE_SESSION_ID` reaches a tool's environment and equals the hooks' `session_id` (the session prints it
    through its shell tool; 5.2.5's own-session refusal rests on it, and the end verifier runs `bonsai ask` inside a
    session); and when `CLAUDE_CODE_BRIDGE_SESSION_ID` is set for a hook (`remote` stays null without Remote Control).
    On Windows: the Go tests natively, and payloads piped into the Windows
    build; a real Windows session only if Claude Code's login there is back (`STATE.md`, "Waiting on Rohan"), else it
    is part of 5.3's Windows check, whose row already holds "the recorder under concurrency" (spec §14).
13. **A scripted differential of records:** the studio's recorded payloads (`studio-app/test/fixtures/hooks/recorded/`
    at `25b6450`, 62 files) through today's sink's exported `buildRecord` and through Bonsai's recorder, compared field
    by field after the mappings above (names, categories, a target outside the checkout). The sink's `main()` is never
    run: on SessionStart it would start the studio's bridge and write the studio's home. The script imports
    `buildRecord` from the scratch copy and calls it with a fixed salt and an environment of its own (`TRINETRA_HOME` a
    scratch folder, no studio variable); each payload's placeholder `cwd` is replaced by a scratch linked project, and
    `CLAUDE_PROJECT_DIR` points there for Bonsai's side. Counts and field names only in the run report.

**5.2.5, asks and `log append`.** Contract §9 as written; these are the command lines and the rules it leaves open.
1. **The commands.**
   - `bonsai ask --type TYPE --title TEXT --why TEXT [--then TEXT] [--task ID or --doc ID] [--option TEXT]...
     [--verdict pass|fail] [--key KEY] [--json]` files an ask and prints its key; `TYPE` is `Answer`, `Decide`,
     `Look` or `Play`.
   - `bonsai ask --resolve <key> [--json]` withdraws an open ask.
   - `bonsai ask --status <key> [--json]` gives its state and, when answered, the answer.
   - `bonsai answer <key> [--choice TEXT] [--verdict pass|fail] [--words TEXT] [--by TEXT] [--via TEXT] [--json]`:
     `by` is `terminal` unless given (at most 60 characters), `via` null unless given (at most 30).
   - `bonsai asks [--all] [--json]`: the open asks, newest first; `--all` every key.
   - `bonsai log append --label name=value... [--target T] [--text T] [--json]`.

   The flags follow today's studio command, so the workflow pack's protocols change little (5.5). `--resolve`,
   `--status` and `--all` are flags, not words.
2. **What is checked when filing**, each refused with exit 2 and the rule named: the four agent types only (Bless is
   the runner's, 5.4; a type a pack defines is refused until a pack can declare one, "Stale or in tension" below);
   the title and every option one line; at most four options, each different, and only on a Decide; `--verdict` only
   on a Look; Look and Play need `--task`; `--task` and `--doc` not both, each an id matching a declared kind's id
   pattern (5.1.5), never a path; CRLF made LF, then a hidden character refused, not stripped, as today (Unicode's
   control, format, private-use, surrogate and unassigned characters, except a line feed in `why` and in the words;
   and the line and paragraph separators); then every free-text field redacted; then the limits (title 300, why 600,
   then 300, an option 200, words 2,000), refused when over, never cut. No NFC normalising: it needs a library outside
   the standard one, and the bridge may normalise.
3. **Keys** (contract §9.1): `agent:<--key>` (`[A-Za-z0-9][A-Za-z0-9._-]{0,59}`), else `agent:h-` and 12 hex of the
   SHA-256 over the type, the doc or task id (or `answers`) and the stored, redacted title, joined by NUL, so a key
   can be checked from its record alone.
4. **A key's state is its latest record:** `file` opens it, `resolve` resolves it, `answer` answers it. Filing an open
   key again appends a new `file` (the newest wording stands); filing a closed key opens it again, as today. The same
   `answer` again (same key and `by`) writes nothing and exits 0 (contract §9.1), checked first, so a retry is never
   read as `ask-not-open`. Any other `answer` on a key that is not open exits 4 (`ask-not-open`), saying whether it is
   unknown, answered or resolved: the first answer stands. An answer from the session that asked is refused with exit 4
   (`answer-own-session`), comparing `CLAUDE_CODE_SESSION_ID` with the ask's `session` (contract §9.3). A Decide's
   `--choice` must be one of its options; a Look's `--verdict` is `pass` or `fail`. An answer grants nothing: it writes
   its record and its log record (note 6), nothing else.
5. **Where:** the main checkout's `.bonsai/local/asks/<UTC day>.ndjson`, through 5.2.2's finder and append path, at
   most 8,192 bytes a record, the day of the record's own `at`, so an answer goes into its own day's file.
6. **Each ask record has its log record** (contract §8.2: `ask`, "an ask filed, resolved or answered: `target` is its
   key"): one `ask` record in the log for each `file`, `resolve` and `answer` written, after it, with `kind` the op
   and `text` null (the words stay in the asks file); in the asking or answering session's file when
   `CLAUDE_CODE_SESSION_ID` is set, else in today's day file (5.2.2, note 8). An answer that writes nothing writes no
   log record either. The runner's Bless (5.4) gets one through the same function.
7. **`log append`** (contract §8.4): one `event` record in today's `w-` file, `session` and `agent` null; each label
   checked against the labels in force (the packs' and this machine's, 5.1.5), its value read by its definition's
   kind; a name no definition has, or a wrong value, exits 2 (`label-not-defined`). `--target` and `--text` are
   redacted and capped. Chosen over the machine's definitions alone (spec §8's words): one set of definitions for
   every label check, and the studio's are attached on the machine anyway.
8. **Exit codes:** 0 done, or nothing to do; 2 bad input; 3 could not write; 4 wrong state (not open, the asking
   session, an unknown key for `--status`, or no `bonsai.yaml`, naming `bonsai init`). No ask command waits for input.

**5.2.6a, the generated-files page** (spec §6 makes it base's `generated-files` skill, and `base` comes with 5.5):
`docs/reference/generated-files.md` in Bonsai's repo until then, generated by `go generate` from the one Go table of
generated kinds, which this piece extends (each kind, where it lives, who writes it, its default, what is never
cleaned, when it is cleaned), with 5.2.6b's rules below in plain words; a test rebuilds it and fails on any
difference, and a `.gitattributes` line keeps it LF. 5.5 moves the words into the skill and keeps its table generated
from the same Go table. Until then the comment on `generated:` in the `bonsai.yaml` `init` writes names the page's
address in Bonsai's repo instead of the skill. If 5.2.6b changes a rule, it changes the page with it.

**5.2.6b, cleaning per kind.**
1. **Kinds** (spec §6's table): `log` files, `asks` day files and `ladder` results, by file; the sessions table's rows,
   by row. Run reports are never deleted by Bonsai (5.1.6 lists those past their rule as a warning); `tasks` is a
   rebuild. The defaults when `generated` or a kind is absent: log 30 days, asks kept, ladder 7 days, sessions kept.
   `null` keeps.
2. **The rule:** a file or row goes when it is older than `keep_days`, or beyond the newest `keep_newest` of its kind,
   either one cleaning; the protections always win, and a protected one still counts among the newest. Age: a log or
   asks file's last record's `at` (its modification time if the last line does not read); a ladder result's
   `finished`; a row's end.
3. **The protections:** a log file holding an open span (5.2.3's one function), or an ended span or subagent run with
   no row yet in `.bonsai/sessions.md` (rows before files); an asks day file holding the `file` record of an open ask
   (5.2.5's state); a ladder result whose task is not `done` or `cut`, read from main's task files, a task not found
   or not read counting as not done; a row whose task is still open (`none` is never protected).
4. **When:** at a session's end, after its `session_end` line, the recorder cleans `log`, `asks` and `ladder` within a
   budget of 1 s (less if SessionEnd's measured limit asks, 5.2.4 note 12) and leaves the rest to the next end;
   `check --write` cleans the rows, as the table's only writer; 5.4's runner calls the `ladder` kind after each run.
   Chosen over a detached process, which would outlive the session. Guard records a worktree's session wrote in the
   worktree's own folder are not cleaned until 5.3 moves them: the cleaner reads main's `local/` only.
5. **Only Bonsai's own files:** the names it writes (`s-*.ndjson`, `w-<date>.ndjson`, `<date>.ndjson`, `<task>.json`),
   regular files only (a link is never followed), inside the kind's folder under main's `local/`. Anything else is
   left alone and never named.
6. **The `clean` record** (contract §8.2): one per file or row, written after the delete succeeded, in today's day
   file with `session` null; `target` the project-relative path (a row: `.bonsai/sessions.md#` and its key); `reason`
   the rule (`generated.log.keep_days=30`, `generated.asks.keep_newest=50`). A delete that finds the file gone writes
   nothing: another session's cleaner took it. On Windows a file another process holds open is skipped after one try
   and cleaned at a later end. A cleaning killed between a delete and its record leaves the file gone with no record:
   the budget stays well inside SessionEnd's measured limit so that this does not happen in practice, and the run
   report names the gap.

#### Proof for each piece

Every piece: its Go tests and `go vet`, plain and with the fault tag, in WSL and natively on Windows (check 10) before
the push, their counts in the run report; CI green on the pushed commit; no Windows-only skip without a named reason;
the Windows rules of `CLAUDE.md` read in the diff (forward slashes in every stored path, byte-stable output, busy
retries, no `bash` by name). Pieces that run Claude Code (5.2.4, and the end verifier) record its version and the user
settings hashes before and after. The scripted runs (5.2.1's differential; 5.2.4's sessions and its differential of
records) live in `~/bonsai-checks/scripts/`, never committed, and report counts and classes, never a studio string.
The two piece verifiers and the end verifier re-run the tests themselves.

#### 5.2 done

A fresh Opus verifier, at the end of 5.2, runs each check itself on the final commit and passes or fails 5.2:
1. **Set 5:** the manifest matches every byte; `bonsai_path` and `bonsai_sha256` close `bonsai.log/1`;
   `bonsai.asks/1`, `bonsai.logs/1` (and a sessions row's `subagent`, if added) document themselves, each with an
   example; the schema-compare test passes, and fails when the verifier removes a property in a temporary copy;
   `bonsai check --schema` prints both new formats.
2. **Lists:** `bonsai check --schema bonsai.log` prints every event and category from the Go table, and the schema's
   descriptions copy none; `go generate` changes nothing under `docs/reference/`; a changed default in a temporary
   copy fails the generated-files page's test.
3. **Redaction, the differential:** the verifier's own runner (the studio's redactor at `25b6450` by `git show` into
   scratch) and its own Go driver, over 5.2.1's corpus (note 9): no labelled or planted secret leaks by note 9's count
   (no word of it, and no run of 6 or more of its characters, survives); never more copies of a word the studio's
   redactor took out; keep rows unchanged; a fixed point at every cut; the three classes with no leak; each
   over-redaction class (the case folds among them) read and judged.
4. **Redaction, break-it:** the verifier's own shapes (case folds, every kind of whitespace, quotes left open,
   punctuation, names in a row, very long input); `FuzzRedact` for 5 minutes with no failure; the time on 26 KB of
   glued names and on 1 MB of mixed text.
5. **No studio content:** a grep of 5.2's diff, `internal/redact`'s tests above all, for studio paths, project names,
   task and bug ids and private strings; no test row is copied from the studio's tables (the verifier compares
   them).
6. **Appends:** the eight-process append test passes on WSL and natively on Windows; the Windows busy-file test; the
   reader skips and counts a torn line; the salt race; the finder agrees with `workspace.Find` on a main checkout, a
   worktree, a submodule and a separate git folder; every writer of `local/` (`hook start`, `hook record`, `log
   append`, `ask`, `answer`, the cleaner) restores a missing `.bonsai/.gitignore`.
7. **Real sessions** (WSL; Windows too if Claude Code's login there is back), in a fresh scratch project linked by the
   final build and holding one running task, through `claude-here`: sessions that read, edit, search, run a command
   holding a made-up secret, run `bonsai ask`, start a subagent and end; one resumed; and one in a worktree of the
   project. Every record validates against `bonsai.log/1`; the guard's and the recorder's records share a main
   checkout's session file, every line whole; the four tool events of one call (`tool_start`, `permission`,
   `tool_end`, `tool_fail`) share one `input_hash` (the guard's records hold null until 5.3); every `session_start`,
   the resumed one's too, names the binary's path (from `~/` under the home) and a SHA-256 equal to the file's, and
   `bonsai --version` and `go version -m` give its commit; the made-up secret is in no record; the session quoted
   `hook start`'s active task; nothing from `hook record` reached the conversation; the `ask` record carries the
   session's own id. The worktree's recorder lines are in main's `local/`, its guard records in its own folder, and
   `logs` and `check --write` in the main checkout show it. The three Claude Code facts of 5.2.4's note 12 are in the
   run report.
8. **Never blocks:** `hook record` and `hook start` exit 0, and `hook record` prints nothing, on an empty, broken,
   huge or unknown payload, with no `CLAUDE_PROJECT_DIR`, in a folder with no `bonsai.yaml`, with an unreadable
   `bonsai.yaml`, and with a log folder that cannot be written; SessionEnd's line lands in ten `-p` sessions out of
   ten; the hooks' p50 and p95 on both sides, against 5.2.4's.
9. **Bonsai's own lines and consent:** a first link writes them on `--yes`; a project linked by a 5.1 build (5.1's
   last commit, built from a clean clone) gets them only with `--allow-exec` and `--yes` (exit 4 without, nothing
   written), and `check` warns until then; `unlink` and taking a pack out remove them; 5.1's consent break-it cases
   still pass on this build.
10. **The sessions table:** from check 7's logs, `check --write` adds one row per session and per subagent run, each
    with its task; a second run changes no byte; the same log on Windows gives the same bytes; the stale warning
    before, none after; the hours keep subagent time apart from session time.
11. **`logs`:** the listing; `--session` with an 8-character prefix; `--day`; an ambiguous prefix exits 4 naming the
    matches; `--json` validates against `bonsai.logs/1`.
12. **Asks:** filing, `--status`, `--resolve`, and an answer from a terminal; the asking session's answer refused
    (exit 4, nothing written); the same answer twice writes one record; `--type Bless` refused; a hidden character
    refused; a made-up secret in a title, an option and the words absent from the file; every `--json` validates
    against `bonsai.asks/1`; nothing changes but `.bonsai/local/asks/` and one `ask` log record for each ask record
    written.
13. **`log append`:** a defined label accepted, an undefined one exits 2; the record in today's day file, its text
    redacted.
14. **Cleaning:** the verifier's own fixture project, with old, new and protected files and rows of every kind, and a
    decoy of each (a name Bonsai does not write; on Linux, a link): exactly the unprotected old ones go, each with
    one `clean` record; a second run removes nothing; a session end's cleaning stops within its budget on a folder of
    many files.
15. **The memory scan:** a note holding a token shape is a finding naming the file and line, not the value; a clean
    note is not.
16. **Check 10 and CI:** `go test ./...` and `go vet ./...`, plain and tagged, on WSL and natively on Windows, run by
    the verifier; CI green on the final commit.
17. **Stop lines:** 5.2's hours under 48, this section's planning and review included; step 5's Windows-only tally;
    option rounds; nothing written or run in the studio's checkout or in Mimas (the scripts in
    `~/bonsai-checks/scripts/` and the run reports' commands read); the user settings hashes around every Claude Code
    run. **Nothing private:** a grep of the diff and the commit messages.

#### Risk in the code, 5.2

- **A secret that slips through.** Redaction is the part's security. The corpus, the three classes as one rule, the
  fuzz test and two differentials (the builder's and the verifier's) are the guard against it; the reduction of a
  tool call keeps most secrets out of a record before the redactor sees it.
- **Too much taken out.** A redactor that hides ordinary words makes the record useless to the Desk; the keep rows and
  the reported over-redaction classes hold it.
- **JavaScript and Go read text differently:** whitespace, case folding, a character against a UTF-16 unit; note 3
  of 5.2.1 fixes each, and the differential's grid holds the edge cases.
- **Two more processes on every tool call** (an async `record` on PreToolUse and on PostToolUse) beside the guard. On
  Windows each start costs about 65 ms through Git Bash (gate report §2.5), in the background; many quick calls could
  pile them up. 5.2.4 measures them; if Claude Code waits on them or they pile up, the measure comes to the
  orchestrator before the piece lands.
- **SessionEnd waits** for its line and up to 1 s of cleaning.
- **Large payloads:** `hook record` reads a PostToolUse payload whole, tool output included, up to the guard's 64 MiB.
- **The log is a tripwire.** A shell write can forge a record (deny rules stop only the file tools), so the sessions
  table and the hours built from it are evidence, not proof; the run reports stay the hours' source ("How hours are
  counted").
- **A worktree session's records are split until 5.3:** the recorder writes main's `local/`, the guard its own folder.
- **Engine tests churn:** every test that holds Bonsai's own lines changes in 5.2.4; a missed one fails loudly.
- **Claude Code moves:** the payloads' fields, `async`, SessionEnd's time, SubagentStart's fields, what SessionStart's
  stdout does; each read on the version in use and recorded.
- **Cleaning deletes:** only Bonsai's names, regular files, in their own folders, protections first; the fixture
  project carries a decoy of each kind.
- **The salt** can be read by a script an agent runs (the walls stop the file tools and `cat`, not a program); it keys
  a hash that pairs records, nothing more.
- **Processes:** the scripted sessions of 5.2.4 and the end verifier's; the orchestrator sweeps after each agent.

#### Stale or in tension in the spec, for 5.2

- **§7's table** puts the eleven recorded events on `bonsai hook record`, async. Here SessionStart's record is written
  by `hook start`, and SessionEnd's line is synchronous, since an async hook is killed at exit (5.2.4, note 1).
- **§7's "SessionStart (startup, resume, compact)":** `clear` counts too (5.2.4, note 1).
- **§6, cleaning "at the end of each session (the recorder's `session_end`, async, never blocking)":** the end's line is
  synchronous, so cleaning there has a budget of 1 s (5.2.6b, note 4).
- **§6's generated-files page is base's skill;** until 5.5 it is `docs/reference/generated-files.md`, and the comment in
  `bonsai.yaml` names that page, where §6's example names the skill (5.2.6a).
- **The outline's "the hash logged once per session file (kept, now by `hook start`)":** `hook start` hashes the
  binary at every start, so a binary swapped between a session and its resume is seen (5.2.4; spec §3's tripwire).
- **`bonsai_path`:** the guard writes the binary's absolute path, a user folder's name in it, while contract §2.6 keeps
  absolute paths out of records the bridge may forward. 5.2's records write it from `~/` under the home and as it is
  elsewhere (an installed path holds no user's name); the guard's keep their form until 5.3 (5.2.0).
- **§8: `log append` checks labels "against the machine's definitions";** 5.2.5 checks them against every label in
  force, the machine's among them.
- **Contract §8.1 gives `task` and `role` "raw";** Bonsai passes every free-text string through the redactor (raw
  meaning not interpreted); a value with no secret shape comes back unchanged.
- **A target outside the checkout:** today's sink keeps the absolute path; contract §2.6 says records carry
  workspace-relative paths. Here null, as the guard writes (5.2.1, note 7); what lost: the Desk no longer sees where
  an outside read went, which today's sink shows.
- **Three departures from the spec's words:** `logs --day` and `asks --all` are flags beyond spec §4's table; 5.2.2's
  finder reads the `.git` files itself where spec §6 names `git rev-parse --git-common-dir` (the hooks stay fast, and a
  test holds the two equal); the memory scan reads the notes in the working tree where spec §6 says "a committed
  memory note" (what is about to be committed is caught before it is).
- **`log.schema.json`'s `text` names "a prompt's start";** Bonsai keeps no prompt's words (5.2.4, note 5), and 5.2.0
  rewrites the description.
- **§14 gives the hook adapter to 5.3,** while the recorder reads ten events' payloads in 5.2 with its own reader; 5.3
  makes one for both.
- **Contract §3: worktrees write main's `local/`;** the guard writes its own folder until 5.3 (its `record.go`), so a
  worktree session's records are split until then.
- **Contract §9.1's ask `type` allows "a type a pack defines",** but no format lets a pack declare an ask type (spec
  §5's declarations are lanes, document kinds, labels and protected paths): refused until one does, which would be an
  addition.
- **Contract §7.5's columns** do not say how a subagent run's row is told from its session's: the `subagent` field
  (5.2.0).
- **Contract §2.6 has the bridge redact again** with the studio's redactor, which differs from Bonsai's: its second
  pass may change Bonsai's text (counted in 5.2.1's differential; the studio's to weigh at step 7).
- **§8's proof names "the Node one" as the measure;** its output is a floor, not an oracle: Bonsai must hide at least
  what it hides, not write what it writes (5.2.1, note 9).
- **Today's asks normalise words to NFC;** Bonsai's do not (the standard library has no NFC).
- **§6: rows "are only added",** yet `sessions` is a generated kind with `keep_days`: rows go only by a person's rule in
  `bonsai.yaml` (default kept), in `check --write`.

### Step 5.3: the guards (18-29 h, re-ask at 38)

**Rohan's (B).** This section comes to Rohan, because it holds two questions only he can answer, each an option round
(5.3's two; stop line 3 allows two, so 5.3 has no room for a third). The first: how the hook lines find `bonsai`, where
his 8 Oct rule (a fixed place no agent can redirect) meets two older ones (hooks call `bonsai` by name; no committed
file holds a full path); one design meets all three, a machine-wide Claude Code settings file, which he dropped on 7
Oct. The second: who may consent to code (`--allow-exec`) and take Bonsai's guard out of a project (`unlink`): an agent
inside a task he approved, as spec §18 has it, or only a person. Both are written in plain words just below, each with a
recommendation, and go to him with this section, so the answers are in before 5.3 starts. What else changes depends on
them: under the first question's (b), his steps gain an administrator's step on each side during 5.3, and 5.3 becomes
19-30 hours (re-ask 39); under the second question's (i), or (ii) wherever the studio does not manage a project, his
steps gain a typed command whenever an update of one of his linked repos runs code or takes the guard out. Otherwise
nothing of his changes: the hours (18-29) and the re-ask line (38) are the spec's; the order of the parts stands; his
other steps are the ones already on his lists (Claude Code's Windows login back first, `STATE.md`, "Waiting on Rohan";
then the second Windows check, about 10 minutes, after a Sonnet agent's run on both sides); no repo is new and nothing
goes public; no format changes (the guard's new rule names are values of the log's free `rule` string, and `check`'s new
findings are additions). Only 5.3.6, which builds both answers, waits for them; if the second is late, 5.3.2 builds rule
7 as (i), the strictest, and 5.3.6 relaxes it to his answer.

#### The question for Rohan: how the hook lines find `bonsai`

**Rohan's answer, 9 Oct: (a), the installed place written into the lines.**

**What the line is today.** Every time an agent edits a file or runs a command in a linked project, Claude Code runs a
short line that Bonsai wrote into the project's `.claude/settings.json`: today `bonsai hook guard || exit 2`, and from
this part a second one when a session ends. The shell finds `bonsai` by its name, looking through the PATH, a list of
folders, in order. The `|| exit 2` turns a missing or crashing `bonsai` into a refusal.

**Three rules, and why they clash.** On 8 Oct you asked that the lines call the installed `bonsai` at a fixed place, so
that nothing an agent can edit redirects them. Two rules written the day before say the lines call `bonsai` by name
(spec §3), because no file committed to a project may hold a full path (the skeleton's check 2: a full path differs from
machine to machine and can name a user). A name can be redirected: a `bonsai` put into a folder that comes earlier in
the PATH (in your WSL terminals `~/go/bin`, `~/node_modules/.bin` and `~/.local/bin` all come before `/usr/local/bin`,
and one `go install` puts a `bonsai` there), or a PATH set in a Claude Code settings file. And a fixed place in the
project's own line is a full path in a committed file.

**One design meets all three, and you dropped it on 7 Oct:** a machine-wide Claude Code settings file, outside every
project, holding the line with the full path, while the project's own line stays by name. You dropped that file on 7 Oct
("why do we need this. i dont think we do, seems to be overenginnering"), so it is option (b) below, not the plan's
answer.

**What stays open whatever you choose.** A Claude Code settings file can name a wrapper program that Claude Code runs in
place of every hook line (on both sides); on Windows it can also name a start-up script that Git Bash runs before the
line, or which `bash.exe` runs it. Each of these needs a write to a settings file. An agent's own tools are refused
there (by the guard for the project's files, and from 5.5 by the walls for yours), but a shell command gets through, and
Claude Code applies the change at once. From this part, `bonsai check` reports any such setting in a project. Only (b)
might close these routes too (a machine-wide file's values beat a project's), which is not yet measured.

- **(a) The installed place, written into the project's lines (recommended).** Each line names Windows' installed file
  (`C:\Program Files\Bonsai\bonsai.exe`, which only an administrator can place) and, failing that, WSL's
  (`/usr/local/bin/bonsai`, which only root can place); if neither is there, the call is refused with a line saying a
  person installs Bonsai. Nothing on the PATH is read. It meets your 8 Oct rule. It breaks the letter of the two older
  rules but keeps their purpose: the two places are the same on every machine and name no user, and check 2 gains that
  one named exception. Its cost: Bonsai works in a linked project only where it is installed at exactly those places.
  Anywhere else, every edit and every shell command in a linked project is refused, and a session cannot end until
  Claude Code gives up after eight refusals in a row (a missing `bonsai` does the same today, but today a copy anywhere
  on the PATH also counts). So: no real linked project works on a machine before your install there (WSL at 5.4, Windows
  at 1.0; today only scratch projects are linked, and they use test builds); a Homebrew install (`brew install bonsai`,
  spec §12) lands in a folder the line never looks in, so 5.7 must settle Homebrew; Git for Windows must be installed
  for all users, since in a per-user install the WSL place, as Git Bash reads it, is a folder you can write (`check`
  tests this); and test builds name their own scratch place, in scratch projects only. About 1-2 hours, inside 5.3's.
- **(b) A machine-wide settings file holding the full path; the project's line stays by name.** It meets all three
  rules, and it is the only option whose guard no other settings file can switch off: Claude Code lets no user,
  project or local settings file turn a machine-wide line off, and only such a line stands against a mod's approval
  (spec §7). The file sits outside every repo on each machine and only an administrator writes it. Claude Code runs
  every line that matches, so each guarded call starts a second `bonsai` beside the project's (side by side, so no
  slower, but one more process each time). Its biggest cost: the machine-wide line runs in every Claude Code session on
  that machine, so a missing or broken `bonsai` (a bad pre-release, say) refuses every edit and command in all your
  work, linked or not, until you edit the file as administrator. It reverses your no of 7 Oct. Its values might also pin
  the settings routes above (unmeasured). It moves your installs forward: during 5.3, before 5.3.6 measures, you put a
  build at the two places and write the file on each side as administrator (about 10 minutes a side), and from then on
  it runs in all your sessions. About 1 hour more: 5.3 becomes 19-30 hours, and by the plan's rule (1.3 times the high
  figure) its re-ask line 39.
- **(c) By name, and Bonsai refuses to guard from anywhere but its installed place.** The lines stay as they are. A real
  Bonsai found elsewhere (a `go install` copy, an old build) refuses every call and says where it runs from. It keeps
  the two older rules. Against yours it stops a stray copy of Bonsai, but not a program that is not Bonsai and answers
  to the name (another tool called `bonsai`, or a stand-in an agent writes with a shell command; the guard refuses an
  agent's file tools writing one, under every option). The same install requirement as (a). About 1-2 hours.
- **(d) By name, as the spec has it now.** Each session's start record names the `bonsai` that started it and its
  fingerprint, and once the installer writes its record of the installed copy (5.6), `bonsai check` and `status` say
  when the `bonsai` found is not that one: all after the fact. It keeps the two older rules and does not meet yours.
  Nothing more now.

Not yet measured, and tried first in 5.3: whether, on Windows, Git Bash runs as a program an ordinary file that starts
with `#!`. If it does, a two-line file named `bonsai`, which a shell command can write into a folder on the PATH,
answers for Bonsai under (c) and (d), and that is the plainest reason for (a).

**Recommendation: (a).** It is what you asked for on 8 Oct, and it shuts the PATH to every program, not only to copies
of Bonsai. Against (b): (b) keeps every rule's letter, and only its guard cannot be switched off from a project, but it
brings back the machine-wide file you dropped as overengineering, adds an administrator's step on each machine, and a
broken `bonsai` would stop all your Claude Code work, linked or not. The two older rules exist to keep user names and
machine-specific paths out of a repo, and the two installed places are neither. Whichever you choose, a project linked
before 5.3 takes 5.3's lines (the stop gate's new line, and `Monitor` on the guard's, under every option; under (a),
every line's new form too) with one `bonsai update --allow-exec --yes`. Who may type that command is the second
question.

#### The second question for Rohan: who may consent to code, and take the guard out

**Rohan's answer, 9 Oct: (ii), his grant in a project the studio manages; only a person elsewhere.**

**What these commands do.** `bonsai update --allow-exec --yes` writes code into a project: a pack's hook lines, the
files they run, a pack plugin that carries code (5.1.1), and Bonsai's own changed lines. The flag is the consent; on 9
Oct you decided that such a plugin is installed on a machine only with it there. `bonsai unlink --yes` takes Bonsai out
of a project, its guard's and stop gate's lines with it.

**Who may run them today.** The spec lets an agent run `init`, `update` and `unlink` inside "a task he approved that
names those files" (spec §18), and asks that every command be "completely and easily useable by any ai agent" (spec §3).
In a project the studio manages (from the studio's link, step 7), "a task he approved" can be proved: only your tap
grants the files those commands write. Everywhere else, Bonsai's own repo among them until the studio manages it, an
agent's own task can name those files, so in practice any agent may: it can consent to code by itself, and with two
edits and one command take the guard out.

- **(i) Only a person types them.** An agent never passes `--allow-exec`, never runs `unlink`, and never runs an update
  that removes or changes Bonsai's own guard or stop line, in any linked project. Cost: a recurring step of yours, a
  command typed in a terminal not started from Claude Code, whenever an update of one of your linked repos runs code or
  takes the guard out: Bonsai's own from 5.4, the studio's from step 7, Mimas from step 8, at each Bonsai upgrade that
  changes a line and each pack change that adds code. Agents' scratch work is untouched: their test builds are confined
  to scratch folders.
- **(ii) Your grant in a managed project; only a person elsewhere (recommended).** An agent may, inside a task whose
  grant of `bonsai.yaml` and the lock came from your tap (a project the studio manages); everywhere else, as (i). Cost:
  (i)'s step wherever the studio does not manage the project (Bonsai's own repo until it does); in a managed project
  your approval tap is the consent.
- **(iii) As the spec reads today.** An agent may, inside its own task. Cost: nothing more for you; but the consent to
  code and the guard's lines become tripwires only, which an agent in an unmanaged project can pass by itself.

**Recommendation: (ii).** It keeps the spec's promise wherever your approval can be proved, and asks you only where it
cannot; (i) asks you even where your tap has already said yes; (iii) lets an agent consent to code for you, which your 9
Oct answer was about. Under (i) or (ii), Bonsai holds the rule in the command itself, not only in the guard, so a
script or a variable cannot get round it.

The spec's row (§14): "The adapter (2-3); the path guard with contract §5.5, §10.1 and §13, and its refusals for
`.bonsai/local/` and the tables (7-10); the delete check (3-5); the stop gate (2-3); generated deny rules and
`disableAllHooks: false` (1-2); the binary check (1-2); the second Windows check: guard, delete check, recorder under
concurrency, backslash paths (2-4; **Rohan's:** its real Windows session, §17 step 7, saving about 1 h)". It also
settles the gate report's 5.3 findings and what 5.1 and 5.2 hand on.

**What exists** on `main` at `d341a5a` (read each package's doc comment):
- `internal/guard`: `bonsai hook guard` with the skeleton's one rule: an Edit, Write, MultiEdit or NotebookEdit of a
  path on `bonsai.yaml`'s `protected` or `person_only` list is refused, and nothing grants. It finds the session's
  project from `CLAUDE_PROJECT_DIR`, walking up to the first folder holding `bonsai.yaml` but stopping at one holding
  `.git` (`findProject`, `hook.go`). A path is judged in every form (`paths.go`: links, `~`, and on Windows `\\?\`,
  `/c/`, streams, trailing dots and spaces, short names, case, and the final path by handle through `syscall`,
  `final_windows.go`). Bash and PowerShell are allowed (`shell-not-judged`). Its own 5 s timer (half the line's 10 s
  timeout); a panic or the timer blocks; the over-time answer waits `recordWait` (500 ms) for its record. One `guard`
  record per decision in the session project's own `.bonsai/local/log/` (a worktree's own folder), `input_hash` null,
  `bonsai_path` absolute, `bonsai_sha256` on the record that made the file (`record.go`). The fault switch behind the
  `bonsai_test_fault` tag (`fault_on.go`). Its rules are a `const` block in `guard.go`, their one home.
- `internal/guard/input.go`: the PreToolUse payload reader (the strict decoder, a BOM, a NUL, a 64 MiB cap; the file
  tools' path only).
- `cmd/bonsai/hook.go`: `guard` built; `start`, `stop` and `record` refuse as not built.
- `internal/engine/settings.go`: `ownHooks` holds the guard's line, `bonsai hook guard || exit 2` on
  `Edit|Write|MultiEdit|NotebookEdit|Bash|PowerShell`, timeout 10; a deny rule for each `never_edit` path and each
  pack's; `autoMemoryEnabled: false` and `disableAllHooks: false` (part 3). `isBonsaiHook` claims a line starting
  `bonsai hook`; `isOldBonsaiHook` takes any line calling Bonsai by a full path for a 0.4.3 line, which `init` removes.
  Consent (5.1.1): a hook line added or changed needs `--allow-exec`, but for Bonsai's own lines at a first link.
- `internal/workspace`: `Home` (`BONSAI_HOME`, else `~/.bonsai`), `MachineKey` (the machine folder's name from the
  main checkout's real path), `Find` through git (its note: not for the guard), `IsBusy`.
- `formats/` (set 4): contract §13's sixteen fixture cases in `formats/active-task/`, with `answers.json`'s `guard` and
  `stop_gate` sections, which 5.3 tests (the README, "The active-task fixtures").
- Tests: `internal/guard` (`guard_test.go`, `guard_windows_test.go`, `fault_on_test.go` as 5.1.0 left it) and
  `cmd/bonsai/hook_test.go` (the line through the shell, the four faults).

**What 5.1 and 5.2 will have added** (from this plan's notes; none is built yet. 5.3's start re-reads each against what
landed, and 5.3.0's run report records any difference that changes a note below):
- 5.1.4a: a Go type for each format (the ladder result's among them, which the stop gate reads); the guard's lean read
  of `bonsai.yaml` (the format line, the protected lists, `never_edit`), held by a test.
- 5.1.5: the active-task function (contract §13), one for every reader, proved on its fixture section; the lock's
  `declares` (each pack's protected paths, document kinds, lanes and labels); declared document kinds with their
  folders and id patterns; labels in force; this machine's settings for the workspace, read (`status_writes`,
  `status_command`).
- 5.1.6: `check`'s findings and warnings in one table, among them `disableAllHooks: true` in a project or local
  settings file, an absolute path in a committed format, and the `bonsai` on the PATH against `install.json`.
- 5.1.7: `unlink` and taking a pack out remove Bonsai's own lines (from 5.3, the stop line and the deny rules too).
- 5.1.8: `check --write`, the tables' only writer. 5.1.10: `docs/reference/lists.md`, generated, with its test.
- 5.2.1: `internal/redact`, with the reduction of a shell command to its head.
- 5.2.2: `internal/record`: one append path for every writer of `.bonsai/local/`, the reader, the salt, one builder of
  every record's common fields, and the finder of the main checkout's `local/` (reading `.git` with no git process; main
  used only when its `bonsai.yaml` holds the same id). Its note: the guard keeps its own writer and folder "until 5.3
  decides what the hook path may trust, and may then take this one over".
- 5.2.4: `bonsai hook start` (it makes a session's file first, with the binary's path and hash) and `bonsai hook
  record`, with their own payload reader; the input hash (HMAC-SHA-256 keyed by the salt); their lines in `ownHooks`.
  It leaves to 5.3 the one adapter, the guard's `input_hash`, the guard's records in a worktree and the stop line.
- 5.2.6b: the cleaner, which reads main's `local/` only (guard records in a worktree's own folder wait for 5.3).

**The pieces.** Hours: the spec gives seven figures. Six pieces carry one each; the path guard's 7-10 is split in two by
the planner's judgment, for sizing briefs (5.3.1 3-4, 5.3.2 4-6). In all: low 2+3+4+3+2+1+1+2 = 18; high
3+4+6+5+3+2+2+4 = 29 (pieces 5.3.0 to 5.3.7). This section's planning and review runs count in 5.3's hours, carried in
5.3.0's run report ("What changes", item 3).

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.3.0 | **The hook adapter and its frame**: one reader of Claude Code's hook payloads for every `bonsai hook` word (the events, tools and tool inputs Bonsai reads; backslash paths kept as given; unknown fields ignored), replacing the guard's `input.go` and 5.2.4's own reader; one frame for the hooks that block (the guard, and the stop gate from 5.3.4): its own timer, a panic or an unread payload blocking, the over-time record awaited for a wait given as an option; the four timed tests | Go tests on payloads written fresh for every event and tool from the hooks reference on the version in use; one reader (no other package decodes a payload); the four timed tests 20 times in a row on both sides under load; the guard's timings before and after; check 10; CI | 2-3 | Spec §3, §7 ("One hook adapter"); contract §8.3; Claude Code's hooks reference (the version read goes in the run report); this plan's 5.1.0 note and 5.2.4 notes 1, 4 and 8; `records/runs/R-2026-10-09-5.1.0-fault-test.md` (11:25) |
| 5.3.1 | **Where the guard reads from**: a path judged by the projects holding it (contract §13), the nearest alone only when it is a verified checkout; the main checkout found from `.git` with no git process and **verified** against git's own link back and this machine's record, else no grants; the lists from main's `bonsai.yaml` and the lock's `declares`; records through 5.2.2's append path into main's `local/`, with `input_hash` and `bonsai_path` from `~/`; the Windows junction limits; the guard's timings | Go tests: the project cases (a `.git` in a subfolder, a nested repo, a planted `bonsai.yaml` or `.git`, a path in another project); the main cases (a worktree, a rewritten `.git` file, a forged `commondir`, a decoy main with and without a record, a submodule); a worktree session's records in main's `local/`, cleaned by 5.2.6b's cleaner; Windows junction tests (`mklink /J`, no privilege); timings on both sides; V1's break-it | 3-4 | Spec §3, §6 (`.bonsai/local/`), §7 (the 8 Oct note); contract §2.6, §3, §8.1, §13; gate report §2.5, §5 (5.3); `design/plan.md` "What still links Bonsai and the studio", item 5; this plan's 5.2.2 notes 3 and 8, 5.2.4 notes 3, 6 and 7; `records/runs/R-2026-10-08-hook.md` (19:18, 19:47) |
| 5.3.2 | **What the guard judges**: one function, "may this agent change this path", that the file tools, the delete check and rung 0 (5.4) call: grants from the active task (contract §13: read from main, honoured only while it reads `running`); `person_only` (contract §5.5); the floor of paths guarded whatever the lists say, at any depth for `bonsai.yaml`, `.bonsai/` and `.git`; Bonsai's own files (`.bonsai/local/`, the tables, Bonsai's home) refused in every mode; `command` mode's tripwires (contract §10.1); Bonsai's own commands in a shell call | One Go test per rule, walked from the guard's rules table so a rule with no test fails; the guard's section of set 4's fixtures; `agents` and `command` mode each; a forged machine record and a planted `bonsai.yaml` refused; V2's break-it | 4-6 | Spec §4 (Bonsai's own commands), §6 (`bonsai.yaml`'s lists, the tables), §7 ("What the guard judges"); contract §3 (the home), §5.2, §5.5, §5.6, §7.3, §10.1, §10.6, §13; `formats/README.md` ("The active-task fixtures") |
| 5.3.3 | **The delete check**: in a Bash, PowerShell or Monitor call in the session's own project, a recursive or bulk delete that does not name what it deletes is refused (`git clean -x` or `-X` and `git stash --all` among them), and a delete or move of a guarded path is judged by 5.3.2's function; a word splitter, not a shell reader | A table of command lines (bash, PowerShell, `cmd /c`, nested `bash -c`, here-documents, abbreviated options), each with its answer; `rm -rf node_modules` on about 10,000 entries allowed; the splitter fuzzed; V2's break-it on both sides through the hook line | 3-5 | Spec §7 ("Shell commands"), §19 A; contract §3 (`local/`); gate report §5 (5.3) |
| 5.3.4 | **The stop gate**: `bonsai hook stop` on contract §13: it engages only for a task the environment names for the session's own project; blocks a task missing, named twice or broken; needs a green `local` result in main's `.bonsai/local/ladder/` proving the session's checkout's HEAD, with the floor and the task's rungs. Its line in `ownHooks`, under consent | Go tests on the stop gate's section of set 4's fixtures and on fixture results (green, red, another commit, `mode: ci`, a floor rung missing); the line's consent tests; `unlink` removing it; a scripted session on WSL blocked until a fixture result proves HEAD | 2-3 | Spec §7 (the stop gate), §9 (results); contract §5.6 (the floor), §11, §13; read-only, today's stop gate (`stop-gate.mjs`, as contract §13 cites it) at the studio's commit the brief names; Claude Code's hooks reference (Stop) |
| 5.3.5 | **Bonsai's own deny rules, and `disableAllHooks: false`**: the engine writes, as its own lines beside its hook lines, deny rules over the files only Bonsai writes (`.bonsai/local/`, the two tables, Bonsai's home), each with its sentence; `disableAllHooks: false`, written since part 3, held | Go tests (written, previewed, the file's key order kept, removed by `unlink`); each rule tried once in a real `-p` session on both sides: a file-tool write and a shell redirect refused by Claude Code, while `bonsai check --write` still writes | 1-2 | Spec §6 ("Generated files", the tables), §7 ("Deny rules the engine writes", "What a project hook cannot do"); contract §3, §10.6; this plan's 5.1.1 rule 8 |
| 5.3.6 | **The hook lines' form and the binary check**, on Rohan's answer (above): the form of every line of Bonsai's; how a test build names its place; the guard's matcher held to the tools reference (`Monitor` added); `check`'s findings for a settings `env` that can redirect or silence a hook, and for Windows without Git Bash; the second question's rule in `init`, `update` and `unlink` | Go tests, the shipped line's text run through `sh -c` and Git Bash among them; scripted runs on both sides with decoys and settings routes (note 4); a project linked by the build before taking the new lines only with `--allow-exec --yes`; part 3's checks 1 and 2 re-run as scripts; V3's break-it | 1-2 | Spec §3 (where it lives, the tripwire, "Hooks call `bonsai` by name"), §7 (the 8 Oct note), §14 check 2; contract §2.6; gate report §5 (5.3); Claude Code's hooks and settings references (the shell, `env`, reloads) |
| 5.3.7 | **The second Windows check**: a Sonnet agent's scripted run on both sides with the final build (the guard, the delete check, the stop gate, the recorder under parallel subagents, backslash and `/c/` paths, junctions, the large Write under `missing`), then Rohan's session (spec §17 step 7), then the timings | The agent's report, each case against its answer; Rohan's words; the timings table | 2-4 | Spec §14 (5.3's row), §17 step 7; gate report §2.5, §2.6; "Rohan's sitting" below |
| **5.3** | | | **18-29** (re-ask 38) | |

**The order, side by side where truly independent.** Rohan, 9 Oct: "if you can orchestrate work in parallel do that
whenever possible"; his 8 Oct bar stands: no shared file, and neither's proof resting on the other's. 5.3 starts once
5.2's end verifier has passed 5.2 (no two parts run at once). A piece that runs beside another rebases on `main` and
re-runs its proof if the other lands first ("What changes", item 2). The files each piece owns:

| Piece | Owns |
|---|---|
| 5.3.0 | A new package for the adapter and the frame (`internal/hookin/`, a working name); `internal/guard/input.go` (moved there) and `hook.go`'s timer and answer; 5.2.4's payload reader in `internal/recorder/`; the comment in `internal/guard/fault_on.go`; the four timed tests (in `internal/guard/guard_test.go`, `cmd/bonsai/hook_test.go` and `internal/workspace/places_test.go`) |
| 5.3.1 | `internal/guard/` (`hook.go`, `record.go`, `paths.go`, `final_windows.go` and their tests); the finder's verification in `internal/record/`; the machine record in `internal/workspace/`, and its write at `init` and `update` in `internal/engine/` if 5.1 has not written it; 5.1.4a's test of the guard's lean read; one line of `init`'s closing words |
| 5.3.2 | `internal/guard/rule.go` and new files beside it, one per judgment; the guard's rules table and the floor's table; `docs/reference/lists.md` |
| 5.3.3 | New files only, `internal/guard/shell*.go` and their tests, until its last commit, made on 5.3.2's branch once 5.3.2's builder has finished (order, item 5): it wires them into `rule.go`, adds its rule names and the delete words to the tables and regenerates `docs/reference/lists.md` |
| 5.3.4 | A new package, `internal/stopgate/`; the `stop` word in `cmd/bonsai/hook.go`; `ownHooks` in `internal/engine/settings.go` and the engine tests that hold Bonsai's lines; and a last commit, made after 5.3.2 has landed, adding its rule names to the guard's rules table and regenerating `docs/reference/lists.md` |
| 5.3.5 | `internal/engine/settings.go` (Bonsai's own deny lines) and the engine tests that hold Bonsai's lines |
| 5.3.6 | `internal/engine/settings.go` (the lines' form, `isBonsaiHook`, `isOldBonsaiHook`); 5.1.6's table of findings (one finding, and the absolute-path finding's exception), with its list of settings keys on `docs/reference/lists.md`; the build stamp in `cmd/bonsai`; the refusal of `--allow-exec` and `unlink` in an agent's session, in `cmd/bonsai` and `internal/engine`; the build line of the scripts in `~/bonsai-checks/scripts/` |
| 5.3.7 | Nothing in the repo but what a found failure needs (a fix is its own commit, landed as its piece's) |

1. **5.3.0, 5.3.3 and 5.3.5 start together, side by side,** once the orchestrator has checked their files against what
   landed (above all: where 5.2.4 put its payload reader, and whether 5.1's engine writes the machine record). 5.3.0
   first, since every later hook piece reads payloads through it; 5.3.3 is text in, an answer out, in new files, its
   call to 5.3.2's function a stub in its own tests; 5.3.5 writes settings lines and tries them in real sessions, which
   read no payload. None's proof rests on another's.
2. **5.3.1 after 5.3.0 and 5.3.5 have landed.** It changes `hook.go` and the record path that 5.3.0 reframes, and
   `internal/engine` (`init`'s closing words, the machine record's write), where 5.3.5 works too; 5.3.5 is the shorter,
   so this rarely waits.
3. **5.3.2 after 5.3.1 lands.** It judges what 5.3.1 reads.
4. **5.3.4 after 5.3.1 has landed, beside 5.3.2 and the rest of 5.3.3.** It uses 5.3.0's frame and 5.3.1's verified
   finder, and adds its line to `ownHooks`, which 5.3.5 changed. Until its last commit it shares no file with 5.3.2 or
   5.3.3 (its own package), and its proof (its fixtures and fixture results) rests on no grant.
5. **5.3.3's wiring, V2, then landing.** Once 5.3.2's builder has finished, 5.3.3 rebases onto 5.3.2's branch and makes
   its last commit there: the check wired into `rule.go`, its names and the delete words in the tables, the reference
   page regenerated; then its whole table runs. V2 verifies both on that branch; 5.3.2 lands, and 5.3.3 fast-forwards
   after it. 5.3.4 lands after them, rebased, with a last commit that puts its own names into the rules table and
   regenerates the page.
6. **5.3.6 after 5.3.3 and 5.3.4 have landed and Rohan has answered both questions.** It changes the form of every line
   in `ownHooks` and adds lists to the reference page. If his answer is late, 5.3.6 waits; nothing else does.
7. **5.3.7 last**, on the final build, so the Windows check proves the lines as they ship.

**Who builds and verifies.** Opus builders for every code piece, 5.3.0 to 5.3.6 (guards, hooks, Windows). A Sonnet agent
runs 5.3.7's scripted run (a hand check, Rohan's 8 Oct word); the orchestrator reads its report. Fresh Opus verifiers,
each with a break-it on both sides: **V1 at 5.3.1**, covering 5.3.0's adapter as landed (a payload's shapes, a
redirected main checkout, a planted `bonsai.yaml` or `.git`, cross-project paths, junctions); **V2 on 5.3.2's branch
with 5.3.3 wired on it** (order, item 5), covering both (everything the guard judges, file tools and shell); **V3 at
5.3.6**, covering 5.3.4 and 5.3.5 too (every line Bonsai writes: the stop gate, the deny rules, the lines' form). If
Rohan's answer has not come by the time 5.3.4 and 5.3.5 have landed, a fresh verifier takes those two then, and V3 later
takes 5.3.6 alone, so nothing waits on `main` unverified for his answer. 5.3.0, 5.3.4 and 5.3.5 land before their
verifier on green tests on both sides, CI and the orchestrator's read of the diff, which the run report says; a must-fix
found later is fixed forward. **The 5.3 end verifier** runs "5.3 done" on the final commit, after Rohan's sitting.

#### Where each inherited finding is settled

The gate report's section 5, its 5.3 list, item by item; then the outline's "Settles" and what 5.1 and 5.2 hand on:

| Finding | Settled in | How |
|---|---|---|
| How the hook line finds `bonsai`: by name on the PATH, which a settings `env` could redirect; spec §7's 8 Oct note against §3 and check 2 (gate §5) | Rohan; 5.3.6 | The option round above: one design meets all three, the machine-wide file Rohan dropped on 7 Oct, offered as (b); 5.3.6 builds his answer, and `check` finds a redirecting settings `env` whatever he chooses |
| A `.git` entry in the session's starting subfolder reads the project as unlinked (gate §5) | 5.3.1, 5.3.2 | A path is judged by the projects holding it, walking past any `.git` (contract §13); the nearest wins alone only when it is a verified checkout, so a planted `bonsai.yaml` or `.git` narrows nothing, and the floor refuses an agent's tools writing either at any depth. The starting folder only names the session's own project. A session started in a subfolder, or moved by `/cd`, loads none of the project's settings, so no guard runs there at all (Claude Code's settings reference, measured on the version in use): `init`'s closing words say so |
| `bonsai.yaml` read from the working tree; what the guard trusts to find the main checkout must not be rewritable unnoticed (gate §5; spec §7's 8 Oct note) | 5.3.1, 5.3.2, 5.3.3 | The main checkout verified (5.3.1 note 2); a worktree's lists come from main's `bonsai.yaml`, so a branch's copy changes nothing; the floor guards `bonsai.yaml`, the lock, the settings files and `.git` whatever the lists say (a nested `bonsai.yaml`, `.bonsai/` or `.git` too, and a machine record in Bonsai's home); the delete check refuses deleting or moving them; a shell rewrite in main's own tree stays a tripwire case (rung 0 and `check`, question A) |
| Shell `rm` or `mv` of `bonsai.yaml` or the lock is not judged (gate §5) | 5.3.3 | A delete or move of a guarded path is judged as an edit of it |
| Windows: a dangling junction into a protected folder allowed; a junction swapped between check and write; the admin share (gate §5) | 5.3.1 | A dangling junction refused (`bad-path`: the write through it would fail anyway); this machine's admin share read as its drive and judged in that form; the swap race and other shares documented as the limits of a hook that runs before the call (note 6) |
| WSL: a session's first call over 5 ms (5.73 at p50, the self-hash); the guard's p95 about 0.6 ms above part 5's, under load (gate §5) | 5.3.1, 5.3.7 | The self-hash left the guard's first call with 5.2.4 (`hook start` makes the file); tasks are read only for a guarded path; timings beside a minimal Go hook in the same harness, with the load average, so load reads out; the bar: a typical call's p95 under 5 ms on WSL |
| The large-Write fail-closed test ran on WSL only (gate §5) | 5.3.7 | A Write past the 64 KiB pipe buffer under `missing`, in a scripted Windows session |
| `golang.org/x/sys`: about 3 ms more on every Windows start; spec §3 plans it for 5.4's job objects (gate §5) | Decided here | Not linked. One binary links every package, so "off the hook's path" means out of the binary: Windows calls go through `syscall`'s lazily loaded DLL procedures, as the guard's final path does. 5.4 may link it only with the hook line's Windows p95 within 1 ms of the build before, both in its run report. The end check: `go list -deps ./cmd/bonsai` names no `golang.org/x/sys` |
| The four timed tests: `TestOverTimeBlocks` with `recordWait`; `TestFaultsThroughTheHookLine/slow`; "busy forever"; the guard tests on the default 5 s budget (the coordinator, 9 Oct; this plan's 5.1.0 note) | 5.3.0 | 5.3.0 note 3 |
| 5.1.0's leftovers: `fault_on.go`'s comment "a test's waits here for the timer" describes no test; the answer waiting for its record cannot be proved without racing 500 ms (`R-2026-10-09-5.1.0-fault-test.md`, 11:25) | 5.3.0 | The comment rewritten; the wait an option of the frame (note 3) |
| The guard's and the stop gate's sections of set 4's fixtures (outline; 5.1.3) | 5.3.2, 5.3.4 | Every case, built as `formats/README.md` says, on both sides |
| The guard's `input_hash` null, its `bonsai_path` absolute, its own writer, and a worktree's guard records in the worktree's own folder, uncleaned (5.2.0, 5.2.2, 5.2.4, 5.2.6b) | 5.3.1 | The salt's HMAC (5.2.4's function); `bonsai_path` from `~/`; 5.2.2's append path and record builder; records in the verified main's `local/`, which the cleaner reads |
| The one hook adapter (spec §14 gives it to 5.3; 5.2.4 reads its ten events with its own reader) | 5.3.0 | One package; the recorder and `hook start` move onto it |
| `hook stop` and the stop gate's line (5.2.4) | 5.3.4 | Built; the line added under consent |
| "The guard's refusal of a hand edit is 5.3's" (5.1.8's note) | 5.3.2 | The tables refused to file tools in every mode, granted or not |
| A later change of Bonsai's own line, "5.3's hook-line form", needs `--allow-exec` (5.1.1 rule 6) | 5.3.4, 5.3.6 | Tested: a project linked by the build before takes the stop line and the new form only with `--allow-exec --yes` |
| The guard's lean read holds the format line, the protected lists and `never_edit` (5.1.4a) | 5.3.1 | Widened to the fields the guard now judges by (note 3); 5.1.4a's test updated |
| The recorder's real Windows session, if Claude Code's Windows login was not back during 5.2 (5.2.4 note 12) | 5.3.7 | The agent runs 5.2.4 note 12's sessions on Windows too, read back as 5.2's done check 7 reads them |

#### Notes per piece

**5.3.0, the hook adapter and its frame.** Spec §7: "One hook adapter reads Claude Code's payload (tool names, backslash
paths on Windows, unknown fields). A payload it cannot read blocks."
1. **One reader** (`internal/hookin`, a working name). Each payload is decoded once by the strict decoder, with the
   guard's rules for a BOM, a NUL in a string read, bad UTF-8, duplicate keys and the 64 MiB cap, into one type: the
   common fields (`session_id`, `transcript_path`, `cwd`, `permission_mode`, `hook_event_name`, and `agent_id` and
   `agent_type` inside a subagent) and each event's own (a tool event's `tool_name`, `tool_use_id` and `tool_input`;
   Stop's `stop_hook_active`; SessionStart's `source` and `model`; SessionEnd's `reason`; Notification's message and
   type; the subagent events' fields). The tool inputs Bonsai reads: the file tools' path (`file_path`, NotebookEdit's
   `notebook_path`); what Edit, MultiEdit and Write would write (`old_string`, `new_string`, `replace_all`, `edits`,
   `content`), which `command` mode needs (5.3.2 rule 6); and the `command` of Bash, PowerShell and Monitor (5.3.6 note
   5). Unknown fields are ignored; an unknown tool is passed on by name. A path is kept as written, backslashes and all:
   its forms are the guard's (`paths.go`). The fields come from the hooks reference on the version in use, named in the
   run report.
2. **What an unread payload does is the caller's:** the guard and the stop gate block; `hook record` writes nothing and
   exits 0; `hook start` prints what it can (5.2.4, notes 8 and 9). With one reader, a field Claude Code adds or renames
   is read in one place; a test fails if another package decodes a payload.
3. **The four timed tests**, with the guard's budget kept at 5 s, half the line's 10 s timeout (Claude Code lets a timed
   out hook's call through; `TestBudgetIsHalfTheHookTimeout` holds the two together):
   - `TestOverTimeBlocks`: `recordWait` becomes the frame's option (500 ms by default). The record half passes a wait
     far past any load and finds the record there when `Main` returns; a second case holds the record's writer (a test
     hook) and finds the answer back within the option's wait. The bound on the answer's time becomes a ceiling only a
     hang crosses (the budget, the wait and 5 s).
   - `TestFaultsThroughTheHookLine/slow`: kept. Its 9 s is its point (the timer must answer before Claude Code's 10 s
     timeout) and its margin is about 3.5 s over the 5.5 s it needs; the 9 becomes the line's timeout constant less
     1 s, so the two cannot drift.
   - "busy forever" (`internal/workspace/places_test.go`): proves the retry gives up by counting its tries against an
     injected clock, not by a 2 s wall-clock bound.
   - The guard tests on the default budget: every test not about time passes a budget of a minute, so a loaded runner
     can never turn its answer into `over-time`; only the over-time tests use short budgets.

   Proof: each passes `-count=20`, plain and tagged, on both sides with eight busy loops beside it.
4. **The frame** (today's `Main`, made shareable): its own timer first, the work in a goroutine, a panic as
   `internal-error`, the answer with its exit code (0, or 2 with the reason and the next step on stderr), the over-time
   record awaited for the option's wait. The stop gate uses it with its own budget (5.3.4 note 4). The comment in
   `fault_on.go` is rewritten to what its tests now do.
5. **No slower:** `hook guard`'s p50 and p95 on WSL before and after, with part 5's harness, in the run report.

**5.3.1, where the guard reads from.**
1. **The projects holding the path** (contract §13: "the project is the one holding the path, for the guard"). For a
   file tool's path, in each of its forms, the guard collects every folder at or above it that holds `bonsai.yaml`,
   walking past any `.git` entry to the volume's root. The nearest one wins alone when it is a **verified checkout**
   (note 2): Claude Code puts its own worktrees inside the main checkout (`.claude/worktrees/`), where the parent's
   `.claude/**` would refuse every edit, and those worktrees are verified. When the nearest is not a verified checkout
   (a project folder with no `.git` of its own, a `bonsai.yaml` someone planted, a nested repo never linked on this
   machine), the lists of every enclosing project apply too, each read against its own folder, and grants come only from
   the nearest verified one. So a `bonsai.yaml` with empty lists placed in `docs/`, or a `.git` placed there, narrows
   nothing (and the floor refuses an agent's tools writing either, 5.3.2 rule 2). Bonsai's own repo holds such files
   (`formats/active-task/*/main/bonsai.yaml`): their lists apply beside Bonsai's own, and their task files grant
   nothing. A path in another linked project is judged by that project's lists and running task. The session's own
   project (from `CLAUDE_PROJECT_DIR`, found the same way) says where the environment's task counts (contract §13: "only
   for the session's own project"), which project a shell call is judged in (5.3.2 rule 8) and where a shell call's
   record goes. A path in no linked project is judged by 5.3.2 rule 5 alone (Bonsai's home, a stand-in `bonsai`), else
   allowed (`outside-project`); when neither the path's project nor the session's is linked, the guard allows and
   records nothing (spec §7). Chosen over stopping at `.git` (today), which reads a nested repo, or a `.git` an agent
   made, as unlinked.
   - **A session started in a subfolder** loads none of the project's settings (Claude Code's settings reference reads
     `.claude/settings.json` from the starting folder only), so no Bonsai line runs there; and a person's `/cd` to
     another folder (Claude Code 2.1.246 on) reads the settings again from the new folder, so the same holds after it.
     The builder confirms both on the version in use; `init`'s closing words gain one line ("start Claude Code in this
     folder, and do not `/cd` out of it: a session elsewhere loads none of these settings, so nothing guards it"), and
     the guard's `--help` says the same.
2. **The main checkout, verified.** Tasks are read from main (contract §13) and records go to main's `local/` (contract
   §3). The guard takes over 5.2.2's finder (`.git` read with no git process: a folder means this checkout is main; a
   file names its `gitdir`, whose `commondir` names the common folder, and main is that folder's parent when `gitdir`
   lies under `<common>/worktrees/`) and adds what Rohan's 8 Oct note asks: whatever the guard trusts to find main
   "cannot be rewritten by an agent without the guard noticing (fail closed)". Main is **verified** when all three hold:
   - git's own link back: `<common>/worktrees/<name>/gitdir` names this worktree's `.git` file;
   - main's `bonsai.yaml` holds the workspace id this checkout's holds;
   - **this machine's record** names it: `<home>/workspaces/r-<key of main>/workspace.json` (contract §3: "the main
     checkout's path, and the workspace ids it has held") holds main's path and that id, and it is the **first** record
     on this machine to claim that id: a record keeps the time it was first written, and a later checkout claiming the
     same id (a decoy among them) is not honoured; `check` already warns when two checkouts on one machine hold one id.
     The record lies outside every project and is written only by `init` and `update` run in a main checkout (this piece
     writes it there if 5.1 has not). The guard refuses an agent's tools writing anywhere in Bonsai's home (5.3.2 rule
     5), the engine's deny rule stops a shell redirect there (5.3.5), and a decoy main that `bonsai init` registers
     later is never the first record for the id.

   A checkout that is its own main (`.git` a folder, or no `commondir`) needs the third alone. Verified, the guard reads
   main's `bonsai.yaml`, lock and task files and writes main's `local/`. Not verified (a rewritten `.git` file, a decoy
   main, a forged `commondir`, a worktree made to look like its own main, or a clone that never ran `bonsai update` on
   this machine), it judges by this checkout's own `bonsai.yaml` and lock, every enclosing project's (note 1) and the
   floor, honours **no grant** (rule `main-not-verified`), and writes its record in this checkout's own `local/`; the
   refusal names the step ("a person runs `bonsai update` in the main checkout"). The recorder, asks and the cleaner
   (5.2.2's other users) take the same verified answer, so the rule has one home. Chosen over git's answer (git reads
   files an agent edits: Rohan's reason) and over the link back alone (an agent can forge both ends, inside folders it
   can write). What lost: a clone that never ran `bonsai update` on this machine gets no grants until a person does,
   which its plugin step already asks (`waiting`, 5.1.7); a second clone of one project on one machine gets none in its
   worktrees, and `check`'s warning names it. A forged record needs a write into Bonsai's home by a program the agent
   runs, past the guard and the deny rule: the tripwire case question A accepts. A `BONSAI_HOME` moved by a settings
   file is 5.3.6's `check` finding.
3. **What it reads:** from the (verified main's) `bonsai.yaml`, 5.1.4a's lean read widened to the fields the guard now
   judges by: `format`, `id`, `protected`, `person_only`, `never_edit` and `documents` (the task folder, and in
   `command` mode the declared kinds' folders); from the lock, `declares`, lean (each pack's protected paths and
   document kinds); this machine's `status_writes` only when a judgment needs the mode. 5.1.4a's test that holds the
   guard's read is updated to these fields. A worktree's own copies of `bonsai.yaml` and the lock change nothing while
   main is verified: rung 0 judges a branch's change to them (5.4).
4. **Records:** through 5.2.2's append path and its builder of the common fields (the guard's own writer and
   `openRetry` go), into the `local/` of the project whose rules decided: a file tool's path's project, or a shell
   call's session project. `input_hash` is the salt's HMAC over the tool input (5.2.4's function, so a call's `guard`
   and `tool_start` records share it); `bonsai_path` from `~/` under the home (5.2.0's rule); `bonsai_sha256` on the
   record that made the file (rare now: `hook start` makes a session's file first); `target` the project-relative path
   judged, the shell operand that decided, or null; `text` a refusal's reason. A worktree session's guard records now
   land in main's `local/` (contract §3), so `logs`, the sessions table and the cleaner (5.2.6b) see them; a test shows
   the cleaner removing an old one.
5. **The walk's cost:** the walk up from a path goes to the volume's root, an `Lstat` or two per folder; the session
   project's walk is done once per call; the home's machine records (one small file per linked checkout) are read only
   when a grant is needed.
6. **Windows junctions** (gate report §5): a path whose longest existing part is a junction or link that does not
   resolve is refused (`bad-path`; Windows would fail the write through it, so nothing is lost); a path through this
   machine's admin share (`\\localhost\C$\`, `\\127.0.0.1\C$\`, `\\?\UNC\localhost\C$\`, and the machine's own name
   from `os.Hostname`) is read as its drive (`C:\...`) and judged in that form too. Documented, not fixed: a junction
   swapped between the guard's check and Claude Code's write (a race no hook that runs before the call can close; it
   needs a second process the agent runs), a share of a folder under another name, and a hard link made to a protected
   file (part 5's limit). Tests make junctions with `mklink /J`, which needs no privilege, as part 5's do.
7. **Timings** (gate report §2.5, §5): the guard's p50 and p95 on WSL, and the hook line through Git Bash natively on
   Windows, for a typical call, a call on a guarded path (tasks read) and a session's first call, beside a minimal Go
   hook in the same harness and the load average, before and after, in the run report. Tasks are read only when a
   guarded path is touched (or, in `command` mode, a declared document): a typical call reads `bonsai.yaml`, the lock's
   `declares` and the walk up. The bar: a typical call's p95 under 5 ms on WSL (spec §3); a guarded-path call is
   reported. If the lock's read alone breaks the bar, the measure comes to the orchestrator before the piece goes on.

**5.3.2, what the guard judges.** Every refusal says why first and names the next step (spec §3, §7).
1. **One function** answers "may this agent change this path now": from the path's projects (5.3.1 note 1), their lists,
   the packs' `declares`, the floor and the active task, a decision. The file tools call it for their path; the delete
   check (5.3.3) for each operand; rung 0 (5.4) with the named task's grants at any status (contract §13: rung 0 judges
   finished work), so the guard and rung 0 judge alike.
2. **The floor**: paths the guard guards whatever the project's lists say, because its own rules and lines live in them,
   creation included. At a checkout's top: `bonsai.yaml` and `.bonsai/lock.json` count as person-only (granted as rule 3
   says); `.claude/settings.json`, `.claude/settings.local.json` and `.git` (the folder or file, and all under it) are
   **never granted**, so an agent's own task cannot open them in `agents` mode, where two file-tool edits (a grant, then
   the edit) would otherwise remove the guard's line; a person, or `bonsai update`, changes them. `.bonsai/local/`,
   `.bonsai/.gitignore` and the two tables are Bonsai's alone (rule 5); `.bonsai/STATE.md` stays free (agents rewrite
   it, spec §6). Below the top, at any depth: a `bonsai.yaml` or anything in a `.bonsai/` folder counts as person-only,
   and a `.git` entry is never granted, so no agent's tool plants a project or a checkout inside another (5.3.1 note 1).
   Chosen over the lists alone: a project may leave them short, and a guard whose own files an agent may unguard through
   the same guard guards nothing. What lost: a project cannot let its agents edit the settings files or `.git`; a task
   that changes a nested `bonsai.yaml` (Bonsai's own `formats/active-task/` fixtures) names it in its grants. The floor
   is a Go table, its one home, on the reference page. For deletes, the floor's `.git` is the project's own top one
   (5.3.3 rule 4).
3. **Grants** (contract §13, §5.6): a protected path (the lists, the packs' declared paths, the floor's person-only
   paths) is allowed when the active task, found by 5.1.5's function in the nearest verified project, with `BONSAI_TASK`
   counting only when that project is the session's own, reads `running` and its `bonsai.allows` holds a glob matching
   the path. Else it is refused (`protected`, or `person-only`), the reason naming the task or why there is none ("no
   active task: two tasks read running"). The guard takes no `--task`. Set 4's `guard` section is the measure.
4. **Person-only** (contract §5.5): in `command` mode (this machine's `status_writes: command`) a person-only path
   counts in `bonsai.allows` only as a person granted it. The guard cannot tell who wrote a grant, so what stops a
   forged one is rule 6's refusal of an agent's edit that puts a person-only path into `bonsai.allows`, and the Desk's
   check (contract §10.6). In `agents` mode grants count as written (`status --json` says so), so a person-only path is
   grantable like any protected one: the contract's own gap in an unmanaged workspace, which the floor's never-granted
   paths do not share.
5. **Bonsai's own files, in every mode, never granted** (`bonsai-files`): an agent's file-tool write into
   `.bonsai/local/` (this checkout's or main's), of `.bonsai/.gitignore` or of a generated table (`.bonsai/tasks.md`,
   `.bonsai/sessions.md`), naming `bonsai check --write` for the tables (spec §6, §7); anywhere in **Bonsai's home**
   (`BONSAI_HOME` as the hook sees it, else `~/.bonsai`, and `~/.bonsai` as well when the two differ), whose machine
   records the guard trusts (5.3.1 note 2) and whose settings hold `status_writes`, so that one write there cannot turn
   `command` mode's tripwires off; and, outside every project, of a file named `bonsai`, `bonsai.exe`, `bonsai.cmd`,
   `bonsai.bat` or `bonsai.ps1` (`bonsai-stand-in`), under every option of the hook-line question, so the Write tool
   cannot place a stand-in on the PATH. Bonsai writes its own files as a program, not through the file tools. A test
   shows a forged machine record refused.
6. **`command` mode's tripwires** (contract §10.1), only when the path's project reads `command`: an agent's edit of a
   task file outside the main checkout; of a declared document's status line (a new document may be created at its first
   status only); of the lane line; one that puts a person-only path into `bonsai.allows`; one that sets a label whose
   definition reads `set_by: outside` (contract §5.2); and a write of a Bonsai-kind file (task, run, state) that fails
   its format. Each names `status_command` as the next step. The guard judges the file as it would be after the call:
   Write's `content`; Edit's and MultiEdit's replacements applied to the file on disk as Claude Code applies them (an
   `old_string` not found changes nothing, and Claude Code fails that call itself); read under format 0 or 1. Only files
   in declared document folders are read this way, so other calls pay nothing.
7. **Bonsai's own commands in a shell call** (spec §4, §7; contract §10.6; the second question): `bonsai settings set`
   and `bonsai labels attach` or `detach` are refused in an agent session (5.6's commands refuse themselves too).
   `bonsai init`, `update` and `unlink` are judged as one edit of `bonsai.yaml` and `.bonsai/lock.json`; under the
   second question's (i) or (ii), `--allow-exec`, `unlink`, and a line that unsets or blanks `CLAUDE_CODE_CHILD_SESSION`
   (`env -u`, `unset`, `$env:` or `Remove-Item Env:`) are also refused as a person's commands (under (ii), not with a
   person's grant in a project in `command` mode). The command itself refuses too (5.3.6 note 8), since a variable,
   `xargs`, a script or a subprocess can hide a flag from the guard; the guard is the second layer. Every other word is
   allowed. 5.3.2 builds this judgment, as (i) if Rohan's answer is not in yet; 5.3.3's last commit calls it for each
   `bonsai` its splitter finds, whatever folder `bonsai` is called from.
8. **Which project a shell call is judged in.** 5.3.3 rule 3 (a delete that does not name what it deletes) applies
   wherever its operand points: an unnamed operand (`rm -rf $X/*`, `rm -rf /tmp/*`) cannot be placed. Rule 8 bounds only
   what can be placed, 5.3.3 rule 4 and rule 7: a named operand is judged in the session's own project (from a worktree,
   also its verified main) or a folder holding it (`rm -rf ..`); and in any other linked project on this machine's
   record (verified as 5.3.1 note 2 says), only a delete or move of that project's top folder or its floor folders
   (`.git`, `.bonsai`, `.claude`). A folder that cannot be told (after a `cd` to anything but a plain name, `env -C`, a
   `cd` inside a subshell) counts as the session's own project. Chosen because Bonsai's own builders, from 5.4 working
   in sessions inside its linked repo, make and delete linked scratch projects all day (on scratch homes' records, not
   this machine's), while a named `rm -rf` of another real project of Rohan's (the studio's repo from step 7, Mimas from
   step 8) is the accident he named. What lost: a shell delete of a guarded path inside another linked project, below
   its top and floor folders, and any shell delete in a linked project not on this machine's record; their file tools
   are still judged (rule 3).
9. **The rules table:** each rule's name and one line on it, in the place of today's `const` block, its one home;
   `docs/reference/lists.md` lists it (5.1.10's generator gains it if it has not). Names are kebab-case, as today's
   (`protected`, `person-only`, `granted`, `bonsai-files`, `bonsai-stand-in`, `main-not-verified`, one per
   `command`-mode tripwire, the delete check's, the stop gate's); the builder fixes them.
10. **Crying wolf, by test:** an ordinary edit in a running task's worktree, a granted path, a free path in a project
    with a long protected list, a new file in a free folder and `.bonsai/STATE.md` all pass; the run report lists every
    refusal the builder's own sessions met that should not have happened.

**5.3.3, the delete check.** Spec §7: "a small delete check that refuses a recursive or bulk delete that does not name
what it deletes (a `git clean` with `-x` or `-X` among them, since either deletes `.bonsai/local/`, and `git stash
--all`, which moves it away), and one that would delete a protected path". Rohan's 4 Oct answer, kept on 7 Oct (question
A): no shell reader.
1. **A word splitter, not a reader.** A command line (Bash's, PowerShell's or Monitor's) is split into simple commands
   at `;`, `&&`, `||`, `|`, `&` and line ends, and into words by the shell's quoting (bash: single and double quotes and
   the backslash; PowerShell: single and double quotes and the backtick). The text inside `$( )`, backticks, `( )` and
   `{ }`, the string given to `bash -c`, `sh -c`, `eval`, `powershell -Command`, `pwsh -c` or `cmd /c`, and a
   here-document or here-string fed to a shell (`bash <<EOF`, `bash -s`, a PowerShell `@' '@` piped to `powershell`) are
   split again, to a depth of three. To reach the program it skips leading `NAME=value` words wherever a command starts,
   and words that run another program, with their options and arguments: `sudo`, `env`, `command`, `exec`, `nice`,
   `nohup`, `time`, `timeout`, and `xargs` (which also makes a delete bulk); and git's global options before its
   subcommand (`git -C dir clean -fdx`, `-c`, `--git-dir`, `--work-tree`). Nothing is expanded or run. A line it cannot
   split (a quote left open, a depth past three) is refused only when it holds a delete word; else it is not judged.
2. **Delete words**, matched by base name, case-folded, with `.exe` dropped (`/bin/rm`, `RM.EXE`, `cmd.exe /c`): `rm`,
   `rmdir`, `unlink`, `shred`, `git rm`, `git clean`, `git stash` with `--all` or `-a`, `find` with `-delete` or with
   `-exec`, `-execdir` or `-ok` running `rm`; moves, since a move takes a path away: `mv`, `git mv`, PowerShell's
   `Move-Item` (`mv`, `mi`, `move`) and `Rename-Item` (`ren`, `rni`), `cmd`'s `move`, `ren` and `rename`; PowerShell's
   `Remove-Item` and its aliases (`rm`, `ri`, `del`, `erase`, `rd`, `rmdir`); `cmd`'s `del`, `erase`, `rd` and `rmdir`;
   and, as bulk deletes of their destination, `rsync` with any `--delete` option and `robocopy` with `/MIR` or `/PURGE`.
   A Go table, its one home, on the reference page.
3. **Not naming what it deletes**, refused (`delete-unnamed`): a recursive delete (`-r`, `-R`, `--recursive`, short
   options bundled in any order such as `-rf` or `-fR`, GNU long options by any unambiguous prefix such as `--rec`,
   PowerShell's `-Recurse` case-blind and by any unambiguous prefix such as `-r` or `-Rec`, `/s`) or a bulk one (`find`,
   `xargs`, `git clean`, `rsync`, `robocopy`, a pipeline into `Remove-Item`, `-Include`, `-Exclude` or `-Filter`) with
   an operand that is not a plain name: one holding a glob character, a variable or a substitution (`$`, a backtick,
   `%NAME%`), or one that is `.`, `..`, `/`, `~`, a drive's root, empty or ends in `/.`; a delete of several files by a
   glob, recursive or not (`rm *.log` is bulk); `git clean` with no path; and, always, `git clean` with `-x` or `-X`,
   `git stash --all`, and `--no-preserve-root` or its prefixes. `-LiteralPath` names its operand as written. A recursive
   delete of a folder by its name (`rm -rf build`) names what it deletes and is judged by rule 4 alone: refusing it
   would cry wolf on routine work.
4. **A guarded path:** every operand of a delete word, and both sides of a move, made absolute against the payload's
   `cwd`, is judged by 5.3.2's function as an edit of that path (`delete-protected`), where 5.3.2 rule 8 places it. A
   `cd` to a plain name earlier in the same line moves the folder; a `cd` to anything else, `env -C` or a `cd` inside a
   subshell leaves the folder unknown: a recursive delete's relative operand then counts as unnamed (rule 3), and any
   other is judged as if in the payload's folder, in the session's own project. A recursive delete or a move of a folder
   is judged as an edit of the guarded paths under it found without reading its tree: a list glob whose fixed start (up
   to its first wildcard) lies under the folder or above it, and the floor's top paths. A glob with no fixed start
   (`**/x`) is judged against the operand itself only, never against what a deleted folder holds. Chosen over reading
   the tree, which costs time on every recursive delete and, past any cap, would refuse `rm -rf node_modules` or
   `target`; what lost: a file guarded only by a `**/` glob inside a folder deleted by name is not seen (the deletion
   shows in rung 0's diff). The folders holding the floor's top paths count as guarded (the project's own `.git`,
   `.bonsai` and `.claude`, and the project's top folder); a `.git` deeper down (inside `node_modules`, say) does not,
   for deletes.
5. **What it leaves** (question A): a shell write to a guarded path (a redirect, `cp` over it, an interpreter) is not
   judged; rung 0 and the Desk catch it after the fact. A delete inside a script run by name is not seen either, nor, in
   another linked project, a named delete below its top and floor folders (5.3.2 rule 8).
6. **The answer** names the operand and the rule ("`rm -rf junk/*` does not name what it deletes (a glob): name the
   files, or delete the folder by its name"); the record keeps the command's head only (5.2.1's reduction) and the
   operand that decided as `target`.
7. **Proof:** a table of lines written fresh, each with its answer, run as Go tests on both sides: every form of rules 1
   to 3 (prefix words, base names, `.exe`, abbreviated and bundled options, moves, here-documents, `rsync --delete`,
   `robocopy /MIR`), and among the calls ordinary work must pass, `rm -rf node_modules` on a folder of about 10,000
   entries with a nested `.git`, within the budget; the splitter fuzzed (no panic; an unsplittable line holding a delete
   word never allowed); V2's own lines through the hook line on both sides, Rohan's two asks among them.

**5.3.4, the stop gate.** Spec §7: "it engages for a named task, needs a green `local` ladder result at HEAD with the
floor plus the task's rungs, and blocks on a missing or broken task file"; contract §13 says from where.
1. **When it engages:** only when `BONSAI_TASK` names a task and the session's own project is linked (a hook takes no
   `--task`); else it allows with no record. The task is read from the verified main (5.3.1 note 2); when main is not
   verified, it blocks, since it cannot trust what it would read.
2. **It blocks** (exit 2; the reason first and the next step, which Claude Code gives the agent as the reason to go on)
   when: the named task is missing, named twice, or does not parse under its format; there is no result at main's
   `.bonsai/local/ladder/<task id>.json`, or it does not read as `bonsai.ladder/1`; its `mode` is not `local`; `green`
   is false; `git.sha` is not the session's checkout's HEAD, or `git.dirty` is true; `requested` lacks a rung of the
   workspace's `ladder_floor` or of the task's `bonsai.ladder` (contract §5.6), or one of them is not green. HEAD is
   read from the session's checkout's git files (a branch's ref file, `packed-refs`, a detached HEAD) with no git
   process; a HEAD it cannot read blocks. When every check holds, it allows. Its next step names `bonsai ladder --task
   <id>`.
3. **The run report:** contract §13 has the gate read "the run report from the session's own checkout, where the
   builder writes it". Today's gate is the measure of what it checks there: the builder reads it read-only at the
   studio's commit, builds the same check and writes it in the run report. A check today's gate makes that the contract
   does not name comes to the orchestrator before it is built.
4. **Its own timer** (5.3.0's frame): a Stop hook that times out does not block either, so the gate blocks itself at
   half its line's timeout.
5. **Claude Code's own limit:** after eight blocks in a row with no tool call between them (a tool call starts the count
   again), Claude Code ends the turn (its Stop cap, `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`, hooks reference). The gate does
   not read `stop_hook_active`; it blocks each time and each block is recorded. So "a builder cannot stop before its
   proof is green" holds up to that cap, and the log shows every block. A settings file that changes the cap is a
   `check` finding (5.3.6 note 3).
6. **Its line:** `bonsai hook stop || exit 2` on Stop, timeout 10, no matcher (spec §7's table), in `ownHooks` with its
   sentence, in 5.3.6's form once that lands. A first link writes it on `--yes`; a project linked before takes it only
   with `--allow-exec --yes` (5.1.1 rule 1); `unlink` and taking a pack out remove it (5.1.7).
7. **Its record:** when it engages, a `guard` record (`agent_event` `Stop`), allow or deny, its rule from the rules
   table, through 5.3.1's record path. When it does not engage, it writes none (the recorder's `stop` record stands).
8. **Subagents:** the line is on Stop only, so a builder run as another session's subagent (SubagentStop) is not
   gated, and a session whose environment names a task is gated at its own end. How Bonsai's own builders are named is
   5.4's (its outline: "the stop gate binding the orchestrator's own builders").
9. **Proof:** the stop gate's section of set 4's fixtures (each case built as the README says; each `task: ok` case
   given a fixture result proving its HEAD), and one test per blocking case of note 2; a scripted `-p` session on WSL
   with `BONSAI_TASK` set is blocked, then stops once a fixture result proving HEAD is written. 5.4's runner writes real
   results.

**5.3.5, Bonsai's own deny rules and `disableAllHooks: false`.**
1. **What is left of the row.** Part 3 already writes a deny rule for each `never_edit` path and each pack's, and
   `disableAllHooks: false`. 5.3.5 adds Bonsai's own rules over the files only Bonsai writes, as its own lines (marked
   `Own`, in the engine's table beside `ownHooks`), each with its sentence in the preview: `Edit(//**/.bonsai/local/**)`
   (the `//**/` form, so a worktree session, whose `local/` is main's, outside its folder, meets it; spec §7), an `Edit`
   rule for each generated table, and `Edit(~/.bonsai/**)` over Bonsai's home, whose machine records the guard trusts
   (5.3.1 note 2; a home moved by `BONSAI_HOME` is guarded by the guard alone, 5.3.2 rule 5). A deny rule runs no code,
   so `--yes` writes it (5.1.1 rule 1); `unlink` removes it.
2. **Why the engine's and not `base`'s** (spec §7 lists the `local/` rule and the home's `settings.json` and `labels/`
   among base's): from 5.4 the stop gate trusts the ladder results in `local/`, and Bonsai links its own repo at 5.4,
   before `base` exists (5.5); a project may also link without `base`. When 5.5 writes base's walls, those lines stay
   the engine's, their one home (`base` keeps `Read(~/.bonsai/salt)`, a secret's wall); 5.5 reads this note. The walls
   round secret files stay base's, as spec §7 has them.
3. **The form that anchors to the project.** Claude Code reads a rule's path relative to the folder, from the settings
   file's folder (`/`), or absolute (`//`). The builder tries the forms in a real session and uses the one that holds
   for a session started in the project's top folder; part 3's `never_edit` rules take the same form if theirs does not
   hold.
4. **Proof in real sessions** (`-p` through `claude-here`, on both sides; the version and the user settings hashes
   recorded): a file-tool write and a shell redirect into `.bonsai/local/` and into each table are refused by Claude
   Code; `bonsai check --write` run through the shell tool still writes the tables (a program, not a tool write; if the
   version refuses it, the tables' rules are dropped and the guard alone covers them, written in the run report).
   `disableAllHooks: false` is written and kept, and `check` finds `true` in a project or local file (5.1.6); that it
   beats a `true` in the user's own settings is Claude Code's documented order and is not tried, since no test touches
   a real user's settings file.

**5.3.6, the hook lines' form and the binary check.**
1. **The designs weighed**, against Rohan's 8 Oct rule, spec §3 ("by name"), check 2 (no full path committed) and
   Windows. The routes every design leaves open but 8, and perhaps 5: a settings file's `env` naming a wrapper that
   Claude Code runs in place of each shell-form line (`CLAUDE_CODE_SHELL_PREFIX`, both sides), and on Windows a script
   Git Bash runs first (`BASH_ENV`) or the `bash.exe` that runs the line (`CLAUDE_CODE_GIT_BASH_PATH`, any file of that
   name); each needs a settings write, and note 3's `check` finding reports it.

   | Design | Rohan's rule | §3 | Check 2 | Verdict |
   |---|---|---|---|---|
   | 1. By name on the PATH (today) | No: any `bonsai` earlier on the PATH, or a PATH set in a settings file, answers | Yes | Yes | Option (d) |
   | 2. By name; Bonsai refuses unless it runs from its installed place | Partly: a real Bonsai found elsewhere refuses; a program that is not Bonsai but answers to the name does not | Yes | Yes | Option (c) |
   | 3. The installed places in the project's line, Windows' first, then WSL's | Yes, but for the settings routes above | No | In letter only: two system places, the same on every machine, naming no user | Option (a), recommended |
   | 4. The line sets its own PATH to the installed folders, then calls by name | As 3 | In form only | As 3 | Dropped: 3's costs with nothing gained |
   | 5. The full path in a machine-wide Claude Code settings file outside every repo (the managed file), the project's line by name | Yes, on a machine holding the file; its `env` values might also pin the settings routes, a managed value beating a project's (unmeasured) | Yes | Yes | Meets all three: option (b). Rohan dropped the machine-wide file on 7 Oct (question B, "overengineering") |
   | 6. A small launcher found by name that runs the installed place | No: the launcher is found by name | Yes | Yes | Dropped: moves the problem one step |
   | 7. The guard checks its own path and SHA-256 against `install.json` (5.6) | As 2, and only once the installer writes it; the home moves with `BONSAI_HOME` | Yes | Yes | Folded into 2, the installed place a constant; the hash of a root-owned file adds nothing |
   | 8. Claude Code's exec form (`args`, no shell) with the full path | Yes, and it closes the settings routes too: exec-form lines run with no shell and no wrapper (the env-vars reference) | No | No | Dropped: with no shell there is no `\|\| exit 2`, so a binary that cannot start does not refuse the call (Claude Code reads a failed start as a non-blocking error), nor does a crash killed by a signal (the crash fault); one exec line cannot serve both sides; and on Windows the path `/usr/local/bin/bonsai` names `C:\usr\local\bin\bonsai.exe`, a folder a normal user can make |

2. **Built on Rohan's answer.**
   - **(a):** the guard's line becomes `b="/c/Program Files/Bonsai/bonsai.exe"; [ -x "$b" ] || b=/usr/local/bin/bonsai;
     [ -x "$b" ] || { echo "bonsai guard: bonsai is not installed in C:/Program Files/Bonsai or /usr/local/bin, so the
     call is refused; next: a person installs it" >&2; exit 2; }; "$b" hook guard || exit 2`, and every other word's
     line the same (the recorder's lines end without `|| exit 2` and print nothing when Bonsai is missing, as 5.2.4 has
     them). Windows' place comes first: on Linux a `/c` folder at the root needs root to make, while under Git Bash
     `/usr/local/bin` is a folder inside Git's own install, which a per-user Git for Windows lets the user write. So (a)
     needs Git for Windows installed for all users, and `check` on Windows finds a per-user one (a Git install outside
     `C:\Program Files`). `[`, `echo` and `exit` are built into `sh` and Git Bash, so the line starts no extra process.
     The two places are Go constants, their one home, which 5.6's installers and `check` read too. `isBonsaiHook` claims
     exactly this form; `isOldBonsaiHook` still takes every other line calling Bonsai by a full path (a 0.4.3 line
     naming `/usr/local/bin/bonsai` among them), only not this form. `check`'s absolute-path finding (5.1.6) excepts the
     two places inside Bonsai's own lines in this form, and nothing else.
     - **Test builds** need their own place: a build stamped `go build -ldflags "-X <package>.hookPlace=<its own path>
       -X <package>.scratchRoot=<a folder>"` writes `b='<its path>'` alone in its lines, refuses `init`, `update` and
       `unlink` in any folder outside its stamped scratch root, so a stamped line never reaches a real project, and its
       `check` also excepts its own place. A test holds both stamps empty in a plain build. A Go test runs the shipped
       line's exact text, made by the same function with the two places swapped for temporary folders (one with a space
       in its name), through `sh -c` and, on Windows, Git Bash: the first place present; only the second; neither (exit
       2, and the message); and a decoy `bonsai` first on the PATH, ignored. Part 3's checks 1 and 2 are re-run as
       scripts on the unstamped final build. The real places are proved at 5.4 (Rohan's WSL install) and at 1.0
       (Windows).
     - This piece also reworks the scratch launchers' faults (`claude-here`, never committed) and `fault_on.go`'s words
       for them: `missing` becomes a stamped place holding nothing; `minimal-path` shows a bare PATH changes nothing.
   - **(b):** the project's lines stay by name; the guard's and the stop gate's lines, in (a)'s form, go into Claude
     Code's managed settings file on each side (`/etc/claude-code/managed-settings.json` on Linux and WSL, `C:\Program
     Files\ClaudeCode\managed-settings.json` on Windows; where WSL's file must live and who may write it is read from
     the managed-settings reference first), with `env` values pinning the settings routes if a measure shows a managed
     value beats a project's for a hook. The exact lines go on a page in `docs/`. During 5.3, before this piece
     measures, Rohan puts a build at the two installed places and writes the file on each side as administrator (about
     10 minutes a side; 5.6's installers do it later); from then on it runs in all his sessions, and a broken build
     there stops all his Claude Code work until he edits the file. About 1 hour more than (a): this piece 2-3 hours, 5.3
     19-30, re-ask 39.
   - **(c):** the lines stay by name. `bonsai hook <word>` first compares its own path (`os.Executable`, links resolved,
     case-folded on Windows) with the installed places (the same constants); when it differs, the guard and the stop
     gate refuse, naming where they run from and the installed place, and `start` and `record` exit 0 having written
     nothing. A test build is stamped as in (a) to skip the comparison.
   - **(d):** the lines do not change; `check`'s and `status`'s comparison with the installed copy waits for
     `install.json` (5.6, spec §3).

   Whichever: a project linked by the build before takes the new lines only with `--allow-exec --yes`, tested.
3. **`check`'s finding for settings that redirect a hook** (every option): a `.claude/settings.json` or
   `.claude/settings.local.json` whose `env` sets `PATH`, `BASH_ENV`, `ENV`, `CLAUDE_CODE_SHELL_PREFIX`,
   `CLAUDE_CODE_GIT_BASH_PATH`, `CLAUDE_PROJECT_DIR`, `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`, `CLAUDE_CODE_CHILD_SESSION` or
   any `BONSAI_` name is a finding, naming the file and the key, never the value. Why: Claude Code gives a settings
   file's `env` to every hook and applies a change to it, and to the hooks, in the running session (its settings and
   env-vars references, read 9 Oct). Those keys wrap or redirect a hook line, choose the shell that runs it, move the
   project the guard reads, change the stop gate's cap, hide an agent's session from Bonsai's commands (note 8), or move
   the home and the task. The list is a Go table, its one home, on the reference page. The user's own settings file is
   not read: it is the person's, and from 5.5 base's walls refuse an agent's tools there.
4. **Measured first, on both sides, in scripted sessions, and recorded:** that `CLAUDE_CODE_SHELL_PREFIX`, `BASH_ENV`
   (Git Bash), `CLAUDE_CODE_GIT_BASH_PATH` and a project `env` PATH holding a decoy each reach a hook line, set before
   the session and changed during it; whether `env` can change `CLAUDE_PROJECT_DIR` as a hook sees it, or blank
   `CLAUDE_CODE_CHILD_SESSION` for a tool's commands; on Windows, whether Git Bash runs a two-line file named `bonsai`
   (no `.exe`) that starts with `#!`, from a folder on the PATH; and the decoys: a script `bonsai`, and a real Bonsai
   build, first on the PATH. Under (a) the guard still refuses a protected edit with either decoy; under (c) the real
   Bonsai refuses and the script gets through (the residual); under (d) both get through. A result that changes what the
   round told Rohan goes to the orchestrator before the piece goes on.
5. **The matcher.** The guard's line matches the tools that write files or run commands, from a Go table read from
   Claude Code's tools reference on the version in use (the version in the run report). It gains `Monitor`, which runs a
   command under Bash's permission rules (the tools reference), and 5.3.0's adapter reads Monitor's command as it reads
   Bash's, so the delete check judges it. A changed matcher is a changed hook line, so a project linked before takes it
   only with `--allow-exec --yes`.
6. **Windows without Git Bash.** Claude Code then runs a line in PowerShell. Windows PowerShell 5.1 cannot read `||`,
   and no PowerShell reads (a)'s form, so a line fails to parse, and a hook that fails does not block: the call goes
   through, under every option (today's line still blocks under PowerShell 7, which reads `||`). So `check` and `status`
   give a finding on Windows when Git Bash cannot be found where Claude Code looks for it, proved through an injected
   lookup in a Go test. What Claude Code does without Git Bash, and what the hooks' documented `"shell": "bash"` field
   does then, is read from its docs, not measured: Git Bash is installed on this PC, and an invalid
   `CLAUDE_CODE_GIT_BASH_PATH` is ignored and Git Bash found anyway (the env-vars reference). If a machine without Git
   Bash turns up, it is tried there; if the docs say `"shell": "bash"` makes a line refuse, Bonsai's lines carry it.
7. **Hours.** This piece carries a lot for 1-2 hours (the form, the stamps, the line test, the matcher, two findings,
   the measures, note 8). Its run report keeps its running hours; past 2, the orchestrator says so before V3 is briefed.
8. **The second question, built.** Under (i) or (ii): `bonsai init`, `update` and `unlink` themselves refuse, when
   `CLAUDE_CODE_CHILD_SESSION` is set, `--allow-exec`, `unlink`, and an update that removes or changes Bonsai's own
   guard or stop line (counted as code here, against 5.1.1 rule 1's "a removed hook line runs nothing"); under (ii), not
   when the project reads `command` and the running task's grant of `bonsai.yaml` and the lock stands (5.3.2 rule 3: a
   person's tap). Claude Code sets that variable in every process its Bash, PowerShell and Monitor tools and its hook
   commands start, so scripts and subprocesses inherit it (the env-vars reference). The refusal names the next step ("a
   person runs this command in a terminal of their own"). A stamped scratch build skips it: its scratch root confines it
   (note 2), so builders' links to the test pack keep working. The guard's rule 7 stays as the second layer. A test
   holds `--allow-exec` to one spelling with no environment or file equivalent (the parser takes `--allow-exec` alone
   and refuses `--allow-exec=true`); note 4 measures whether a project `env` can blank the variable (note 3 finds it).
   Tests on both sides: an agent-shaped run (the variable set) of `update --allow-exec`, `unlink` and an update removing
   the stop line is refused, through a variable (`F=--allow-exec; bonsai update $F --yes`), `xargs`, a script and a
   subprocess too; the same with the variable unset (a person) is written. Under (iii), none of this: rule 7 judges the
   grants, and the command is free.

**5.3.7, the second Windows check.** Spec §14's row: "guard, delete check, recorder under concurrency, backslash paths".
Claude Code's login on Windows has expired (`STATE.md`, "Waiting on Rohan"): the agent's Windows run and Rohan's
session both need it back, so the orchestrator asks him for it at 5.3's start if it is not back by then.
1. **The agent's run first** (Sonnet; part 4b's method: `claude-here` in scratch projects linked by the final build,
   the Claude Code version and the user settings hashes recorded). On WSL and on Windows, `-p` sessions that: edit a
   free file (done), a protected one (refused), a person-only one (refused), a protected path a running task grants
   (done) and the same once the task reads `verify` (refused); write into `.bonsai/local/` and a table (refused); run
   `rm -rf junk/*` (refused), `rm junk/one.txt` (done), `rm bonsai.yaml`, `mv .bonsai/lock.json x`, `git clean -fdx`
   and `git stash --all` (each refused), and on Windows `Remove-Item -Recurse junk\*` and `cmd /c rd /s /q .bonsai`
   (refused), and `rm -rf junk/*` through Monitor (refused); name a path with backslashes, `.\`, a drive letter and Git
   Bash's `/c/` form (each judged as the plain path); edit through a junction into a protected folder and through a
   dangling one (refused); stop with `BONSAI_TASK` set and no result (blocked), then with a fixture result proving HEAD
   (stops); start four subagents that edit and run commands at once (every log line whole, every record valid); and on
   Windows, under `missing`, ask for a Write of about 2,000 lines (refused, the file never made: the case the first
   sitting left open). If 5.2 could not run the recorder's real Windows session, the agent runs 5.2.4 note 12's sessions
   on Windows here too. The report gives each answer against the expected one, with the records' rules.
2. **The folder for Rohan:** `%USERPROFILE%\bonsai-checks\project-guard` linked again by the final build, holding
   `person.txt` on `person_only`, a free file, `junk\` with `one.txt` and `two.txt`, and no running task; his launcher
   is `%USERPROFILE%\bonsai-checks\bin\claude-here.cmd`.
3. **Rohan's session**, below.
4. **Timings** (gate report §2.5's table, both sides): the hook line's p50 and p95 for the guard's typical call, a
   guarded path, a shell call and the stop gate, with `hook start` and `hook record` beside them and a minimal Go hook
   in the same harness, against the gate's figures.

#### Proof for each piece

Every piece: its Go tests and `go vet`, plain and with the fault tag, on WSL and natively on Windows (check 10) before
the push, their counts in the run report; CI green on the pushed commit; no Windows-only skip without a named reason (a
junction test runs on Windows only and says so); the Windows rules of `CLAUDE.md` read in the diff (forward slashes in
every stored path, byte-stable output, busy retries, no `bash` by name). Pieces that run Claude Code (5.3.4's session,
5.3.5, 5.3.6, 5.3.7 and every verifier) record its version and the user settings hashes before and after. Scripted runs
live in `~/bonsai-checks/scripts/`, never committed. V1, V2, V3 and the end verifier re-run the tests themselves, on
both sides.

#### 5.3 done

A fresh Opus verifier, at the end of 5.3 and after Rohan's sitting, runs each check itself on the final commit, on both
sides, and passes or fails 5.3:
1. **One adapter:** a payload of each event Bonsai reads goes through the one reader, and no other package decodes one;
   a payload it cannot read blocks the guard and the stop gate, and leaves `hook record` silent with exit 0.
2. **The timed tests:** the four of 5.3.0 note 3 pass `-count=20`, plain and tagged, on both sides with eight busy loops
   beside them; no guard test but the over-time ones runs on a budget under a minute.
3. **The fixtures:** every case of the `guard` and `stop_gate` sections of `formats/active-task/answers.json` gives its
   answer, the worktree cases built with git (Windows git on Windows).
4. **Grants and lists:** a running task's `bonsai.allows` opens a protected path, and the path closes when the task
   reads `verify`; no active task refuses with its reason; a pack's declared protected path is guarded; the floor holds
   where the project's lists leave its paths out, at any depth for `bonsai.yaml`, `.bonsai/` and `.git` (an agent's
   Write of `docs/bonsai.yaml` or `docs/.git` refused; the same placed by a shell narrows nothing);
   `.claude/settings.json`, `.claude/settings.local.json` and `.git` stay refused under a running task's grant;
   `.bonsai/local/`, `.bonsai/.gitignore`, the tables, Bonsai's home (a forged machine record among them) and a stand-in
   `bonsai` outside a project are refused in both modes, granted or not; `.bonsai/STATE.md` is free.
5. **`command` mode:** each tripwire of contract §10.1 refused with its reason first and `status_command` as the next
   step; in `agents` mode each allowed.
6. **Where the guard reads from:** a `.git` in the starting subfolder changes nothing; a path in another linked project
   is judged by its lists and its running task, the environment counting only for the session's own; a worktree
   session's grants come from main's task files and its records land in main's `local/`; with the worktree's `.git` file
   pointed at a decoy main (with and without the link back forged, with and without the decoy's `bonsai.yaml` holding
   the id, and with a decoy record written later than main's), no grant is honoured and the refusal names the step.
7. **Records:** every guard and stop-gate record validates against `bonsai.log/1`; a call's `guard` record and the
   recorder's `tool_start` share one `input_hash`; `bonsai_path` is from `~/` under the home; the cleaner removes an old
   file of guard records from main's `local/`.
8. **The delete check:** the verifier's own table of lines (bash, PowerShell, `cmd /c`, Monitor, nested `bash -c`,
   here-documents) through the hook line on both sides: every unnamed recursive or bulk delete refused, wherever it
   points, in every spelling of 5.3.3 rules 1 to 3; `git clean -x` and `-X` and `git stash --all` refused; a delete or
   move of a guarded path, or of a folder holding one, refused; a named delete of a free path allowed; `rm -rf
   node_modules` on about 10,000 entries allowed within the budget; a named delete of another linked project's top
   folder or its `.git`, `.bonsai` or `.claude` refused when that project is on this machine's record, and allowed for a
   scratch project on a scratch home's record; after a `cd` that cannot be told, operands judged in the session's own
   project. No wrong allow; every refusal of an ordinary command listed and judged.
9. **Bonsai's own commands,** as Rohan's second answer has it: under (i) or (ii), `--allow-exec`, `unlink` and an update
   removing Bonsai's guard or stop line refused by the command itself whenever `CLAUDE_CODE_CHILD_SESSION` is set
   (through a variable, `xargs`, a script and a subprocess too) and by the guard, a line unsetting the variable among
   them; under (ii), allowed with a person's grant in a project in `command` mode; a stamped scratch build free;
   `--allow-exec` one spelling; under (iii), allowed under a grant of `bonsai.yaml` and the lock. Always: `settings set`
   and `labels` refused.
10. **The stop gate:** each blocking case of 5.3.4 note 2 blocks with its reason; green at HEAD allows; nothing named
    allows with no record; over its budget it blocks; its line is written at a first link, added to an older project
    only with `--allow-exec --yes`, and removed by `unlink`.
11. **Deny rules:** Bonsai's own rules written and previewed with their sentences; in a real session on each side a
    file-tool write and a shell redirect into `.bonsai/local/` and into a table are refused, and `check --write` still
    writes.
12. **The lines' form**, as Rohan chose: 5.3.6 note 2's line test and note 4's decoy runs give his option's answers on
    both sides; a stamped build refuses to link outside its scratch root, and a plain build's stamps are empty; a
    project linked by the build before takes the new lines, `Monitor` on the guard's matcher among them, only with
    `--allow-exec --yes`; part 3's checks 1 and 2 pass as scripts on the unstamped final build.
13. **Settings that redirect a hook:** `check` finds each key of 5.3.6 note 3 in a project and in a local settings file,
    naming the key and not its value; on Windows, `check`'s and `status`'s finding for Git Bash missing proved through
    an injected lookup in a Go test (the real behaviour without Git Bash is read from Claude Code's docs, not measured,
    unless a machine without it turns up).
14. **Windows:** junctions into a protected folder (from inside, from outside, past 260 characters) and a dangling one
    refused; this machine's admin share read as its drive and judged in that form (a protected path through it refused,
    a free one allowed); backslash, `.\`, drive-letter and `/c/` forms judged as the plain path; the large Write under
    `missing` refused in a real session; the recorder's lines whole under parallel subagents.
15. **Timings:** the guard's typical call on WSL under 5 ms at p95 beside the minimal Go hook; 5.3.7 note 4's other
    figures against the gate's; `go list -deps ./cmd/bonsai` names no `golang.org/x/sys`.
16. **The sittings:** the Sonnet agent's report and Rohan's words in the run report, each case as expected.
17. **The break-it:** the verifier's own inputs on both sides beyond the builders' tables (path forms, cross-project
    paths, a redirected main, shell lines, payload shapes): no wrong allow.
18. **Check 10 and CI:** `go test ./...` and `go vet ./...`, plain and tagged, on WSL and natively on Windows, run by
    the verifier; CI green on the final commit.
19. **Stop lines:** 5.3's hours under 38, this section's planning and review included; step 5's Windows-only tally;
    option rounds (two, the hook lines' and who may consent to code: stop line 3's limit); nothing written or run in the
    studio's checkout or in Mimas; the user settings hashes around every Claude Code run. **Nothing private:** a grep of
    the diff and the commit messages.

#### Rohan's sitting (spec §17 step 7)

What reaches him, in one batch with exact lines, once the agent's run has passed (the orchestrator sends it; the text
here is the plan's). If Claude Code's Windows login is not back, that line goes to him at 5.3's start, on its own,
since the agent's Windows run needs it too. In PowerShell:

```powershell
claude
```

then type `/login`, sign in, and type `/exit`.

The check, about 10 minutes, in a normal PowerShell:

```powershell
cd "$env:USERPROFILE\bonsai-checks\project-guard"
& "$env:USERPROFILE\bonsai-checks\bin\claude-here.cmd"
```

If Claude Code asks whether to trust this folder, say yes. Then ask Claude three things, one at a time, in this order:
1. "Add the line hello to person.txt." It should be refused: only a person changes that file.
2. "Delete the file junk/one.txt." It should be done.
3. "Run exactly this command, as written: rm -rf junk/*" It should be refused: the delete does not name what it deletes.
   Note the command Claude actually ran; if it ran another (say `rm -rf junk`, which names the folder and is allowed),
   say so.

Type `/exit`, and send the orchestrator what Claude showed for each, in your words or copied.

Why these stay his (his 8 Oct word: his sittings hold only what needs a person typing): a typed, interactive session is
the kind he works in, which the agent's `-p` sessions are not (the folder's trust, the screen the refusal reaches);
everything else the agent has run first. If one ask goes wrong, the fix is a later commit and only that ask is repeated.
The third ask names its command, and comes last, because a recursive delete of a folder by its name names what it
deletes and is allowed (5.3.3 rule 3): a rephrased one would empty `junk` before the named delete could be tried.

#### Risk in the code, 5.3

- **A guard that cries wolf gets worked around** (outline). The floor, the delete check and `command` mode refuse more
  than today. Each piece's tests hold the ordinary calls that must pass, and every verifier lists each false refusal it
  meets.
- **Speed.** Every edit and shell call runs the guard. Each read it gains (the lock's `declares`, main's files, the
  machine record, tasks for a guarded path) threatens spec §3's 5 ms; timings are taken before and after each piece
  that adds one.
- **A redirected main.** Verification rests on this machine's record; a program the agent runs that forges it in
  Bonsai's home, past the guard and the deny rule, stays the tripwire case question A accepts.
- **Settings reload mid-session.** Claude Code applies a settings file's hooks and `env` in the running session (its
  references, read 9 Oct): a shell write that removes a line, sets `disableAllHooks: true` in a local file or sets a
  redirecting key switches the guard off or around for that session. The guard refuses file-tool edits of the files,
  the delete check their deletion, and `check` finds the settings; a shell write stays question A's case.
- **Windows.** Junctions, short names, shares, Git Bash's path forms and PowerShell's quoting; without Git Bash every
  line fails open, so `check` finds it missing (5.3.6 note 6). Every Windows-only test says why.
- **The stop gate trapping a session:** bounded by Claude Code's cap of eight blocks; every block is recorded. With
  Bonsai missing, the stop line refuses every session's end up to that cap, as the guard's line refuses every call: the
  fail-closed design, which `init`'s and the hook's help name.
- **Engine tests churn:** 5.3.4, 5.3.5 and 5.3.6 change Bonsai's lines; every engine test and script that holds them
  is updated, and the run reports list each.
- **Claude Code moves:** the payloads' fields, the Stop cap, how settings reload, which shell runs a line, how a deny
  rule's path anchors; each read on the version in use and recorded.
- **Processes:** the scripted sessions of 5.3.4 to 5.3.7 and the verifiers'; the orchestrator sweeps after each agent.

#### Stale or in tension in the spec, for 5.3

- **§3, "Hooks call `bonsai` by name", and check 2, against §7's 8 Oct note:** one design meets all three, a
  machine-wide settings file, which Rohan dropped on 7 Oct; his option round (above) offers it as (b).
- **§7: "In a project with no `bonsai.yaml`, `bonsai hook` exits 0 at once",** against contract §13's "the project is
  the one holding the path": the guard judges a path in a linked project even from a session whose own folder is not
  linked, and allows with no record only when neither is (5.3.1 note 1).
- **§7 puts the `.bonsai/local/` deny rule, and the home's `settings.json` and `labels/` rules, in `base`;** here the
  engine writes them, with the tables' (5.3.5 note 2), and 5.5's walls leave those lines to the engine.
- **Contract §13: "the project is the one holding the path, for the guard":** kept for the file tools, with every
  enclosing project's lists when the nearest is not a verified checkout (5.3.1 note 1); a shell call is judged in the
  session's own project only (5.3.2 rule 8).
- **Spec §18 and §3: an agent may run `init`, `update` and `unlink` inside a task he approved, and every command is
  usable by any agent:** in `agents` mode an agent's own task makes that any agent, which can then consent to code by
  itself and take the guard out; Rohan's second question (above).
- **§14's "generated deny rules"** names work part 3 did in part (the `never_edit` rules, `disableAllHooks: false`);
  read here as the deny rules the engine generates over Bonsai's own generated files (5.3.5).
- **§6 and contract §3: worktrees reach main "through `git rev-parse --git-common-dir`";** the hook path reads the
  `.git` files with no process and verifies the answer against git's link back and this machine's record (5.3.1 note
  2), as Rohan's 8 Oct note asks.
- **Contract §5.5 gives the lists to `bonsai.yaml`;** the floor guards `bonsai.yaml`, the lock, the settings files and
  `.git` whatever they say (5.3.2 rule 2). In `agents` mode a person-only path is grantable by an agent's own
  `bonsai.allows` (grants count as written): `bonsai.yaml` and the lock among them, never the settings files or `.git`,
  which no grant opens.
- **Contract §10.1's "writing a Bonsai-kind file that fails its format"** stands among `command` mode's refusals; read
  as `command` mode only (5.3.2 rule 6).
- **§7: the stop gate "is today's";** Claude Code ends a turn after eight blocks in a row, so the gate binds up to that
  cap (5.3.4 note 5).
- **§3's `golang.org/x/sys` for job objects:** not linked (decided above); 5.4 follows.
- **A session started in a subfolder** loads none of the project's settings (Claude Code's settings reference), so §7's
  guard covers sessions started in the project's top folder (5.3.1 note 1).
- **§17 step 7: "delete the `junk` folder with a recursive delete that does not name what it deletes":** a recursive
  delete of a folder by its name names it (5.3.3 rule 3), so Rohan's ask names `rm -rf junk/*`, and comes after the
  named delete.

### Step 5.4: the ladder runner and the switch (28-44 h, re-ask at 57)

**Rohan's (B), and his 9 Oct direction.** This section comes to Rohan, because it holds his install (spec §17 step 8).
It is written for his direction of 9 Oct, 15:35: "bonsai will mostly (and probably completely) be managed by the ai
agents within a project. so make sure the support for that is top notch. including things like installing, updating,
fixing, checking status, editing etc", which he placed "inside projects": agents link, update, fix, check, read status
and edit, while the program itself stays his install per release. So agents do everything in 5.4 that no safety reason
keeps from them: they run the ladder, read its results and fix what fails; they link Bonsai's own repo at the switch;
they raise the test-count floors. Every refusal of `bonsai ladder` names its next step as a command an agent can run as
written. Three steps stay his, each for a reason given in plain words below: the pre-release install (a root-owned file,
at the place the hook lines name, which no agent may replace); under his answer (ii), `--allow-exec`, `unlink` and an
update that changes Bonsai's own guard or stop line in Bonsai's repo, rarely; and the way back, only if the guard
breaks. Nothing else of his changes: the hours (28-44) and the re-ask line (57) are the spec's; the order of the parts
stands; no repo is new and nothing from the studio goes public; no option round is asked. Every format change is an
addition: one new command output (`bonsai.climb/1`), the active task's last result at the end of `status --json`, two
counts at the end of a ladder result's `new_tests`, new words in open lists, and descriptions. The shapes of a rung's
`ratchet` and `capture` are fixed without giving them a schema type, so `bonsai.workspace/1` needs no new major (note
5.4.0).

#### For Rohan, in plain words

**What 5.4 gives you.** `bonsai ladder`: one command that proves a piece of work is done, as the studio's ladder does
today, in WSL and natively on Windows. It runs a project's checks in order, cheapest first. Everything a check starts
ends with it, so nothing it started is left running (on Windows through the system's "job objects"). It saves the result
in the project's `.bonsai/local/ladder/` with a fingerprint in the log, which is what 5.3's stop gate reads before it
lets a named builder stop. It also keeps test counts from falling, offers each rise for "Bless" (with a check that each
new test fails on the code from before the change), and reports git tricks that hide changes. Then Bonsai starts using
all of it on its own repo.

**Your sitting, once (about 5 minutes; your password).** When the code is done and checked, the orchestrator builds the
pre-release and sends you its fingerprint, a 64-character number. Open a second WSL terminal, so the orchestrator's
session stays open, and type:

```bash
sha256sum ~/bonsai-checks/prerelease/bonsai
sudo install -o root -g root -m 0755 ~/bonsai-checks/prerelease/bonsai /usr/local/bin/bonsai
which -a bonsai
bonsai --version
```

- The first line must print the orchestrator's number. If not, stop and tell it.
- `which -a bonsai` must list `/usr/local/bin/bonsai` and nothing else. If not, stop and tell it.
- `bonsai --version` names the commit it was built from. Send the orchestrator that line.

Why this is yours: the copy at `/usr/local/bin/bonsai` is the one Bonsai's guard lines run in every session, so only
root may place it (your password); an agent that could replace it could replace the guard that holds it. Everything
after the install is the agents': the orchestrator links Bonsai's own repo with the copy you installed (`bonsai init`,
which writes the guard and its other lines and needs no `--allow-exec`, since Bonsai's repo uses no pack yet), commits
it, and checks the guard is live. If Claude Code has not taken up the new lines in its running session, it asks you to
restart it (`/exit`, then `claude` in `~/Servers/Bonsai`). Windows needs nothing until 1.0.

**What changes after the switch.**
- Every piece of work is a task file in `records/tasks/`, numbered after its piece (`T-5501` is piece 5.5.1), saying
  what it may change and what proves it. The orchestrator writes and moves them.
- A piece is proved by `bonsai ladder`, run by the copy you installed: Bonsai's guard check on everything the piece
  changed, `go vet`, the formats check, every test twice (plain, and with the test fault switch), and a git check. The
  orchestrator lands a piece only when its result is green at the exact commit it lands. GitHub's checks and the native
  Windows test run stay as they are (the ladder runs in WSL; Windows gets its installed copy at 1.0).
- Bonsai's guard covers every Claude Code session in `~/Servers/Bonsai`, yours included. It refuses changes to
  `bonsai.yaml`, its lock, Claude Code's settings files, the GitHub workflows, `CLAUDE.md`, `go.mod`, the lint and
  release settings and the four approved design documents (the spec, the contract, the one-pager and the format review),
  unless the running task allows that file. It refuses a recursive delete that does not name what it deletes. It logs
  each session in `.bonsai/local/` (never committed). If you ask Claude yourself to change one of those files, it is
  refused too: change it by hand, or ask the orchestrator, which opens a task for it.
- Claude Code's own automatic memory is switched off in Bonsai's repo (Bonsai's memory replaces it, spec §10). The one
  note it holds there (agents run the hand checks) moves into the repo as Bonsai's first memory note.
- Test counts may only rise. At each part's end the orchestrator raises the floors to the counts the part reached and
  tells you the numbers; nothing to answer. Raising a floor only makes the checks stricter. Lowering one, or loosening
  any other check in `bonsai.yaml` (a rung taken out, a protected file freed), waits for your word: an agent loosening
  the checks it is held to would be judging its own work.
- Anyone who opens Claude Code in a clone of Bonsai without Bonsai installed has every edit and command refused, on
  purpose. `CONTRIBUTING.md` says so and how to work round it.

**Commands that stay yours in Bonsai's repo** (your answer (ii): the studio does not manage Bonsai's repo yet). There an
agent never passes `--allow-exec`, never runs `bonsai unlink`, and never runs an update that removes or changes Bonsai's
guard or stop line; the command itself refuses. Why: each one either lets new code run on your machine or takes the
guard out, and an agent doing that would be consenting for you, or switching off the check on itself. So:
- When a later part changes one of Bonsai's own hook lines, one more line in the sitting where you install its
  pre-release: `cd ~/Servers/Bonsai`, then `bonsai update --allow-exec --yes`. Rare: no later part plans such a change
  before 1.0, and the orchestrator tells you when one does.
- `bonsai unlink`, only if you ever decide to take Bonsai out of its own repo. Nothing plans it.

Every other update of Bonsai's repo is the agents' (linking Bonsai's `base` pack in 5.5 adds no code that runs).

**The way back, if the pre-release blocks work.** If Bonsai's guard breaks, every edit and command in a Claude Code
session in `~/Servers/Bonsai` is refused: it fails closed, on purpose. This is yours because, while it lasts, no agent
there can run a command at all. In a plain WSL terminal (not Claude):
- **One session with Bonsai off, nothing written.** Quit Claude Code, then start it like this instead of plain `claude`:
  ```bash
  cd ~/Servers/Bonsai
  claude --settings '{"disableAllHooks": true}'
  ```
  That session runs with no guard, no stop gate and no log. The orchestrator in it fixes the problem and sends you a new
  build to install; then go back to plain `claude`. An agent tries this line first, in a scratch project, before your
  sitting.
- **Several days with Bonsai off**, for every session in that folder. Only if
  `ls ~/Servers/Bonsai/.claude/settings.local.json` says there is no such file (else ask the orchestrator for the line):
  ```bash
  cd ~/Servers/Bonsai
  printf '{\n  "disableAllHooks": true\n}\n' > .claude/settings.local.json
  ```
  Git ignores that file, and `bonsai check` reports it until you remove it with
  `rm ~/Servers/Bonsai/.claude/settings.local.json`.
- **Back to the build before** (from the second pre-release on; the orchestrator keeps the one you had and gives you its
  number again):
  ```bash
  sha256sum ~/bonsai-checks/prerelease/previous/bonsai
  sudo install -o root -g root -m 0755 ~/bonsai-checks/prerelease/previous/bonsai /usr/local/bin/bonsai
  ```

**Hours, order and your other steps.** 28-44 hours, re-ask at 57, as the spec has them; 5.4 after 5.3 and before 5.5.
Your install was already on your list; new is only the rare update line above. The way back is only for a breakdown.

#### The pieces and their order

The spec's row (§14): "Rungs, process groups and job objects, one ladder at a time, leftovers, `mode`, the fingerprint,
results in the main checkout's `.bonsai/local/ladder/`, rung 0's refusal of branch changes to the tables and of tracked
`local/` files (13-19); floors, ratchets, Bless filing (4-6); new tests must fail, by name, with `base_setup` (7-12);
git integrity (3-5); Bonsai's own `bonsai.yaml`, the switch from the interim proof and the pre-release build Rohan
installs (question C, 1-2)". It also settles the gate report's 5.4 findings and what 5.1, 5.2 and 5.3 hand on.

**What exists** on `main` at `0c88cde` (read each package's doc comment):
- `internal/format/ladder.go` (5.1.4a): the result's Go type (`Ladder`, `LadderGit`, `LadderRung`, `LadderRatchet`,
  `LadderNewTests`, `LadderSkip`), `ReadLadder`, and `Encode`, which writes in schema order held to `bonsai.ladder/1`.
  No code calls the writer yet.
- `internal/format/workspace.go`: `bonsai.yaml` read in full; a `Rung` with `Ratchet` and `Capture` typed `any`, `Tests`
  and `BaseSetup` strings; `LadderFloor`, `Ratchets` (a name to a number), `CIMarkedTests` (names), `Generated.Ladder`.
- `formats/` (set 4): `ladder.schema.json` (every field; a rung's `kind` closed: `guard`, `command`, `ver-git`; its
  `status`, `tests` and the result's `leftovers` typed open); `workspace.schema.json`, where a rung's `ratchet` and
  `capture` have no type at all and `tests` is an open string, "step 5.4 fixes their shape, in a later major if it must"
  (`formats/README.md`); `active-task/answers.json`'s `rung0` section (sixteen cases: rung 0 judges against the
  function's named task); `compare_test.go`, which skips without the base commit unless `CI` is set.
- `.github/workflows/ci.yml`: `test` (Linux, full history: plain and tagged tests, `go vet`, a Windows cross-build),
  `windows`, `lint`, `govulncheck`; CodeQL beside it.
- `cmd/bonsai` has no `ladder` word. Bonsai's repo holds no `bonsai.yaml`, `.bonsai/` or `.claude/`; `CLAUDE.md` and
  `CONTRIBUTING.md` describe the interim proof; `go.mod` requires nothing (the standard library only).
- Today's studio ladder, the measure where the spec says "as today": `tools/ladder/ladder.mjs`, `tools/lib/rungjob.mjs`,
  `tools/lib/proclist.mjs`, `tools/hooks/stop-gate.mjs`, `studio-app/bridge/bless.ts` and the ladder in
  `studio/game.yaml`, at the studio's `b5f7cbf`, read only with `git show` ("What changes", item 10). There a rung's
  `ratchet` is a name and its `capture` maps that name to a pattern with one group (`test_count: '^# pass (\d+)'`);
  `ci_marked_tests` entries carry a name, a mark and a file.

**What 5.1, 5.2 and 5.3 will have added** (from this plan's notes; none of these is built yet. 5.4's start re-reads each
against what landed, and 5.4.0's run report records any difference that changes a note below):
- 5.1.4b: the `error` words' Go table and the flag table every word's `--help` comes from. 5.1.5: the active-task
  function (with `--task` and `BONSAI_TASK`), labels in force, declared document kinds with their id patterns, the
  instruction block, `init` reading an existing `bonsai.yaml`. 5.1.6: `check`'s findings and warnings in one table.
  5.1.8: `check --write`. 5.1.10: `docs/reference/lists.md`, generated.
- 5.2.0 (set 5): `bonsai_path` and `bonsai_sha256` at the end of `bonsai.log/1`; the log's events (`ladder` among them)
  and categories (`Ladder`) in one Go table. 5.2.1: the redactor and a command's head. 5.2.2: one append path, the
  record builder, the main checkout's `local/`. 5.2.4: `hook start`'s opening context, which already reads a task's last
  result ("5.4 writes them"); the `Ladder` category when a shell call's head is `bonsai ladder`. 5.2.5: the ask filing
  function, `agent:` and `ladder:` keys, the `ask` log record, `--type Bless` refused for agents. 5.2.6b: the cleaner,
  with a call for the `ladder` kind "5.4's runner makes".
- 5.3.0: the hook frame. 5.3.1: the verified main checkout and this machine's record. 5.3.2: the one function "may this
  agent change this path", which rung 0 calls with the named task's grants at any status; the floor. 5.3.4: the stop
  gate, reading `.bonsai/local/ladder/<task>.json` (green, `mode: local`, HEAD, clean, the floor and the task's rungs).
  5.3.5: Bonsai's own deny rules (`.bonsai/local/`, the tables, `~/.bonsai`). 5.3.6: the lines in Rohan's (a) form, the
  stamped test builds confined to their scratch root, (ii) held in `init`, `update` and `unlink`; `golang.org/x/sys` not
  linked.

**The pieces.** Hours: the spec gives five figures. The runner core's 13-19 is split over four pieces by the planner's
judgment, for sizing briefs, not a spec figure (5.4.0 1-2, 5.4.1 5-7, 5.4.2 4-6, 5.4.3 3-4); the other four rows carry
the spec's own. In all: low 1+5+4+3+4+7+3+1 = 28; high 2+7+6+4+6+12+5+2 = 44 (pieces 5.4.0 to 5.4.7). This section's
planning and review runs count in 5.4's hours, carried in 5.4.0's run report ("What changes", item 3).

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.4.0 | **Formats set 6 and the ladder's lists**: `bonsai.climb/1`, the `--json` of `bonsai ladder`; two counts at the end of a result's `new_tests`; the descriptions that fix a rung's `ratchet`, `capture` and `tests` and say what a `ladder` log record carries; the ladder's words in Go tables (the `tests` forms, rung statuses, git integrity's findings); 5.4's error words; `climb` wherever the formats are listed; the active task's last result at the end of `status --json`; the schema compare failing, not skipping, under the ladder | The formats test (manifest, docs, examples); the schema-compare test passing on set 6 and failing on a removal in a temporary copy; `check --schema bonsai.climb` and `bonsai.workspace` printing the new words; the reference page's test; check 10; CI | 1-2 | Contract §2.2, §8.1, §8.2, §11; spec §6 (the rung's shape), §9, §16 row 23; `formats/README.md` ("Typed open", "How the set changes"); this plan's 5.1.3 and 5.2.0 notes |
| 5.4.1 | **Rung jobs** (`internal/rungjob/`, new): one rung's command run so that everything it starts ends with it: on Linux in its own session and process group, tagged, with the runner as the reaper of what escapes; on Windows in its own job object, the command started suspended and put in the job before it runs; its timeout; its output; and **the leftovers line** (`internal/leftovers/`, new) | Go tests on real process trees on both sides (detached and daemonised grandchildren, a held pipe, a timeout, a child asking to leave the job, a nested job, the runner killed), a decoy outside the rung left alone; scripted runs on both sides; the job's cost and Windows Defender's first scan measured; V1's break-it | 5-7 | Spec §3 (the standard library, Windows rules), §9; gate report §5 (5.4), §2.5; this plan's 5.3 decision on `golang.org/x/sys`; read-only, `tools/lib/rungjob.mjs` and `tools/lib/proclist.mjs` at `b5f7cbf` |
| 5.4.2 | **The runner, `bonsai ladder`**: the test outputs read (`internal/testout/`, new: `go test -json`, TAP, JUnit); the rungs in order with `required`, the requested set, skips and marks, the ratchet's count; the result in the verified main checkout's `.bonsai/local/ladder/`, `mode`, `--ci`; the fingerprint record; one ladder at a time per home; the `ladder` cleaning call; `--json`; every refusal's next step a command an agent runs as written, and each failing test's own command; `status --json`'s `ladder`; `check`'s rows for the ladder's shape | Go tests on fixture projects (each requested set, a required red, every skip rule, each test form, a result from a worktree landing in main, the lock held and released, a dead holder); every result and `--json` against its schema; 5.3's stop gate accepting a green local result at HEAD and refusing a `ci` one; scripted runs on both sides; V1 | 4-6 | Spec §3, §4 (`ladder`), §6 (the ladder, generated files), §9, §10 (the home's `locks/`); contract §3, §5.6, §8.1, §8.2, §11, §13; this plan's 5.2.2, 5.2.6b and 5.3.4 notes; read-only, `tools/ladder/ladder.mjs` at `b5f7cbf` |
| 5.4.3 | **Rung 0** (`internal/rung0/`, new): what the branch and the working tree changed, each path judged by 5.3.2's function with the named task's grants at any status; a branch that changes a generated table; a tracked or staged `local/` file; `check`'s findings about the commit; `bonsai.*` labels read by contract §5.6's meaning with no definition in force | The `rung0` section of set 4's fixtures, every case, on both sides; fixture repositories for each refusal; a table of paths on which the guard and rung 0 agree; V1 | 3-4 | Spec §6 (the tables, `local/`), §7, §9; contract §5.5, §5.6, §13; this plan's 5.1.8 and 5.3.1-5.3.2 notes; read-only, `ladder.mjs`'s `runGuard` at `b5f7cbf` |
| 5.4.4 | **Floors, ratchets and Bless** (`internal/ratchet/`, new): a count below its floor is red; a count that cannot be read is red; a green, clean local run on the base branch at its HEAD with a count above its floor files one Bless ask; the ask resolved when the offer is gone | Go tests walking a table of every case (each condition false in turn); the ask and its `ask` log record against their schemas; `--ci` never blesses; check 10; CI; V2 | 4-6 | Spec §9 ("Ratchets and Bless"); contract §9.1, §9.2, §10.3, §11; this plan's 5.2.5 notes 2-6; read-only, `studio-app/bridge/bless.ts` at `b5f7cbf` |
| 5.4.5 | **New tests must fail on the base** (`internal/newtests/`, new): the new tests named from the rung's output; a temporary worktree at the base; `base_setup`; the base run as it is, then with the branch's changed test files; "failed", "proves nothing new" and "check not run" told apart | Fixture repositories for each form (`go test -json`, node's TAP, JUnit): a new test failing at the base, one passing there, one that cannot load there, a broken `base_setup`; the worktree removed after; the counts in the result and on the Bless ask; check 10; CI; V2 | 7-12 | Spec §9 ("New tests must fail on the base"); contract §9.1 (`data`), §11; format review 5.2 (spec §16 row 23) |
| 5.4.6 | **Git integrity** (`internal/vergit/`, new): the `ver-git` rung kind: the task's base commit recorded; changed test files, assume-unchanged and skip-worktree entries, edits to `.git/info/exclude`, new stash entries and a rewritten base found and listed; the runner's `dirty` counting what git hides | Fixture repositories, one per finding, on both sides (Windows git there); a clean repository finds nothing; the rung informs and never turns red for a finding; check 10; CI; V2 | 3-5 | Spec §9 ("Git integrity"), §16 row 23; contract §11 |
| 5.4.7 | **The switch**: Bonsai's own `bonsai.yaml`, its task and memory folders, `CLAUDE.md`'s and `CONTRIBUTING.md`'s rules, `.gitignore`; the way back tried; the pre-release; Rohan's install; the link by the orchestrator, committed; the first climbs | A scratch clone of the switch linked and climbed by the final build; the way back in a scratch session; Rohan's lines and words; the end verifier's checks on the real repo | 1-2 | Spec §3, §6, §7, §14 ("How Bonsai's work is proven"), §17 step 8, §19 C; contract §4, §5.5, §5.6, §7.1, §7.4; this plan's "What changes", items 5 and 7, and 5.3's two answers |
| **5.4** | | | **28-44** (re-ask 57) | |

**The order, side by side where truly independent.** Rohan, 9 Oct: "if you can orchestrate work in parallel do that
whenever possible"; his 8 Oct bar stands: no shared file, and neither's proof resting on the other's. 5.4 starts once
5.3's end verifier has passed 5.3 (no two parts run at once). A piece that runs beside another rebases on `main` and
re-runs its proof if the other lands first ("What changes", item 2). The runner calls each rung kind and each step after
a rung through one small file, `internal/ladder/kinds.go`, which 5.4.2 makes; a piece built beside the runner in a
package of its own wires itself in with one last commit there, as 5.3.3 wired into the guard. The files each piece owns:

| Piece | Owns |
|---|---|
| 5.4.0 | `formats/` (`climb.schema.json` and its example; the descriptions in `workspace.schema.json`, `ladder.schema.json` and `log.schema.json`; `new_tests`' two counts; README; manifest; `compare_test.go`'s ladder marker); the error words' Go table; the Go tables of the `tests` forms, rung statuses and git integrity's findings, beside the ladder's type in `internal/format/`; `status.schema.json`'s new `ladder`; wherever 5.1.4a registers a schema and an open list's table, the test that lists every schema, and `status --json`'s `formats`; `docs/reference/lists.md` |
| 5.4.1 | `internal/rungjob/` and `internal/leftovers/` (new) and their tests |
| 5.4.2 | `internal/testout/` and `internal/ladder/` (new); `cmd/bonsai/ladder.go`, the `ladder` word's dispatch and its entry in the flag table; 5.1.6's findings table (rows for the ladder's shape); `status --json`'s `ladder` in `internal/status/`; the stamped build's confinement where 5.3.6 put it (one more word confined) |
| 5.4.3 | `internal/rung0/` (new); the reading of `bonsai.*` labels where 5.1.5 put labels in force, if 5.3 left a gap (note 5.4.3); its last commit: `internal/ladder/kinds.go` |
| 5.4.4 | `internal/ratchet/` (new); its last commit: `internal/ladder/kinds.go` |
| 5.4.5 | `internal/newtests/` (new); its last commit: the rise step in `internal/ratchet/` |
| 5.4.6 | `internal/vergit/` (new); its last commit: `internal/ladder/kinds.go` and the runner's `dirty` |
| 5.4.7 | `bonsai.yaml`, `records/tasks/`, `records/memory/`, `CLAUDE.md`, `CONTRIBUTING.md`, `.gitignore`; then, from the link's `init`, `.bonsai/`, `.claude/settings.json` and the block in `CLAUDE.md` |

1. **5.4.0 and 5.4.1 start together**, once the orchestrator has read where 5.1.4a, 5.1.4b and 5.2.0 put the tables
   5.4.0 extends. 5.4.0 sets every shared list and schema first, so no later piece edits a shared list file; 5.4.1 is
   the longest and riskiest piece, so it starts at once. 5.4.1 is new packages only and needs no format; 5.4.0 runs no
   process. Neither's proof rests on the other's.
2. **5.4.3 and 5.4.6 after 5.4.0 has landed, beside the rest of 5.4.1.** Each reads 5.4.0's tables (rung statuses; git
   integrity's finding names) and works in a new package, judging a fixture repository and returning a rung's outcome.
   Neither runs a rung job, and each is proved on its own fixtures, so neither's proof rests on 5.4.1's, on the runner's
   or on the other's.
3. **5.4.2 after 5.4.0 and 5.4.1 have landed**, beside the rest of 5.4.3 and 5.4.6. It runs every rung through 5.4.1 and
   writes 5.4.0's shapes. It makes `kinds.go` with `guard` and `ver-git` answering "not built yet" (`error`), so it can
   land before either.
4. **5.4.3's and 5.4.6's last commits after 5.4.2 has landed**, one at a time, 5.4.3 first (rung 0 is on every floor):
   each rebases on `main`, wires its kind into `kinds.go` (5.4.6 also the runner's `dirty`), and re-runs its whole proof
   through the runner. **V1** then verifies 5.4.1 to 5.4.3 on `main` (below).
5. **5.4.4 and 5.4.5 after 5.4.3's wiring has landed, side by side.** 5.4.4 wires the ratchet step and Bless into
   `kinds.go`; 5.4.5 works in its own package on fixture repositories and needs only 5.4.2's test outputs, already
   landed. 5.4.5's last commit, after 5.4.4 lands, hangs its check on 5.4.4's rise step. Until then they share no file,
   and 5.4.5's fixtures prove it alone. If 5.4.6's wiring has not landed when 5.4.4 is ready, 5.4.4 waits for it (both
   touch `kinds.go`).
6. **5.4.7's preparation beside 5.4.4 and 5.4.5.** Its files hold no Go code (`bonsai.yaml`, `CLAUDE.md`,
   `CONTRIBUTING.md`, the task and memory folders, `.gitignore`), and it tries the way back in a scratch session with
   any recent build. Its proof (a scratch clone of the switch climbed by the final build, which sets the floors) rests
   on every code piece, so it lands only after **V2** has passed the code (below).
7. **Then, in order:** the pre-release built from `main` once 5.4.7 has landed; Rohan's sitting; the link committed; the
   first climbs; the end verifier.

**Who builds and verifies.** Opus builders for 5.4.1 (process trees, Windows), 5.4.2 (the runner every later proof rests
on), 5.4.3 (rung 0 judges as the guard does), 5.4.4 (it decides a proof's colour and files asks), 5.4.5 (temporary
worktrees and base runs), 5.4.6 (it holds a cheat class) and 5.4.7 (Rohan's main repo, Claude Code's settings); Sonnet
for 5.4.0, whose shapes this section fixes, with the orchestrator's read. Fresh Opus verifiers, each re-running the
tests on both sides: **V1 after 5.4.3's wiring**, covering 5.4.1, 5.4.2 and 5.4.3 (Windows process trees, the runner,
rung 0's judgments), with a break-it on both sides (note 5.4.1's cases and its own); **V2 on the code before the
pre-release**, covering 5.4.4, 5.4.5 and 5.4.6 and running "5.4 done" checks 1 to 15 on the last code commit, so Rohan
installs code a verifier passed; **the 5.4 end verifier** after Rohan's sitting, running "5.4 done" checks 16 to 23 on
the final commit and the installed binary, with V2's report for the rest (the build's rule keeps the Go code V2 passed).
5.4.1 to 5.4.3 land before V1, and 5.4.4 to 5.4.6 before V2, on green tests on both sides, CI and the orchestrator's
read of the diff; a must-fix V1 or V2 finds is fixed forward before the pre-release is built. 5.4.0 lands on green tests
on both sides, CI and the orchestrator's read of the diff, which the run report says. A Sonnet agent runs 5.4.7's
scripted Claude Code sessions (its notes 1 and 2) and the agent's session of "5.4 done" check 14, Rohan's 8 Oct word;
the builder and the orchestrator read its report.

#### Where each inherited finding is settled

The gate report's section 5, its 5.4 list; then the outline's "Settles" and the switch, and what 5.1, 5.2 and 5.3 hand
on:

| Finding | Settled in | How |
|---|---|---|
| Windows job objects, never measured (gate §5, 5.4; spec §15 "Not measured yet") | 5.4.1 | Built and measured: a job per rung, kill on close, no breakaway; its cost per rung, a detached grandchild, a child asking to leave, a job nested inside the rung's (Bonsai's own runner tests climbing inside a rung) and the runner killed, each in the run report and V1's break-it |
| Windows Defender's first scan of a new binary, never measured (gate §5, 5.4) | 5.4.1 | Measured: the first and second start of a freshly built `bonsai.exe`, and of a freshly built test binary under `go test`, each ten times; the default rung timeout keeps that margin (note 5.4.1) |
| `golang.org/x/sys` for job objects; 5.3: not linked, "5.4 may link it only with the hook line's Windows p95 within 1 ms" (gate §5, 5.3; outline) | 5.4.1 | Not linked: the job calls go through `syscall`'s lazily loaded `kernel32` and `ntdll` procedures, as the guard's final path does. `go list -deps ./cmd/bonsai` still names no `golang.org/x/sys`, and the hook line's timings on both sides are compared with 5.3's (the binary grows) |
| Rung 0's section of set 4's active-task fixtures (outline; 5.1.3) | 5.4.3 | Every case, built as `formats/README.md` says, on both sides |
| "From 5.4 the same test is a rung of Bonsai's own ladder" (5.1.3, the schema compare) | 5.4.0, 5.4.7 | 5.4.0 makes the test fail rather than skip under the ladder; 5.4.7 makes it rung 2 of Bonsai's ladder |
| The ladder writer built for 5.4 (5.1.4a) | 5.4.2 | Every result written through it, validated before it is written |
| "rung 0's [refusal of a hand edit of the tables is] 5.4's" (5.1.8) | 5.4.3 | A branch that changes `.bonsai/tasks.md` or `.bonsai/sessions.md` is red |
| A rung's `ratchet` and `capture` untyped, its `tests` words open, "step 5.4 fixes their shape, in a later major if it must" (`formats/README.md`) | 5.4.0 | Fixed by description and by Go readers, with `check` rows; no type added, so no new major (note 5.4.0) |
| `hook start` shows a task's last ladder result, "5.4 writes them" (5.2.4 note 9) | 5.4.2 | Results written where it reads them; a done check reads the context back |
| Bless is the runner's; its `ask` log record "through the same function" (5.2.5 notes 2 and 6) | 5.4.4 | Filed through 5.2.5's function with `source: ladder` |
| "5.4's runner calls the `ladder` kind" of cleaning after each run (5.2.6b note 4) | 5.4.2 | Called after the result and its record are written; the runner's own result is never cleaned in the run that wrote it |
| The `Ladder` category when a shell call's head is `bonsai ladder` (5.2.4) | 5.4.7 | Bonsai's own builders call the ladder by its installed path; 5.4's start reads whether 5.2.1's head keeps `bonsai` for `/usr/local/bin/bonsai ladder`, and a gap is fixed in 5.4.2 |
| "rung 0 judges a branch's change to [a worktree's `bonsai.yaml` and lock]" (5.3.1 note 3); rung 0 calls the guard's one function (5.3.2 rule 1) | 5.4.3 | Both, with the named task's grants at any status |
| How Bonsai's own builders are named, "5.4's" (5.3.4 note 8; outline: "the stop gate binding the orchestrator's own builders") | 5.4.7 | Builders stay subagents, unnamed and not gated; a builder started as its own session is named and gated (the switch, "Tasks and names") |
| "5.4's runner writes real results" for the stop gate (5.3.4 note 9) | 5.4.2 | The stop gate's checks run on real results in "5.4 done" |
| The real hook places "proved at 5.4 (Rohan's WSL install)" (5.3.6 note 2) | 5.4.7 | The line in Bonsai's own settings runs `/usr/local/bin/bonsai`; a protected edit is refused in a real session |
| Stamped test builds confined to their scratch root (5.3.6 note 2) | 5.4.2 | A stamped build's `ladder` refuses to write a result outside its scratch root, as its `init` refuses to link there |
| Bonsai's builders make and delete linked scratch projects from inside Bonsai's linked repo (5.3.2 rule 8) | 5.4.7 | Tried by the end verifier from a session in Bonsai's main checkout: a scratch link on a scratch home and its named delete allowed, an unnamed one refused |
| "The proof changes at 5.4" ("What changes", item 5); the pre-release "built from a clean clone of the commit" with its stamp (item 7) | 5.4.7 | The switch below; the build line in "Rohan's sitting" |
| Base's `bonsai.*` label definitions arrive with 5.5, while Bonsai links at 5.4 (this plan, "Stale or in tension") | 5.4.3, 5.4.7 | The readers act on contract §5.6's meaning with no definition in force; `CLAUDE.md` says what the four mean until base's block does (note 5.4.3) |
| "From 5.4, Bonsai's own `.bonsai/sessions.md` also holds these runs; a Haiku audit compares it with the run reports" ("How hours are counted") | 5.4.7 | The table fills from the switch; the first audit, at 5.4's end, covers the sessions after it |
| A broken pre-release could block edits; "5.4's section writes your way back" (Rohan's "Risk") | 5.4.7 | Rohan's part above; the first form tried in a scratch session before his sitting |
| "From 5.4, a command of yours when an update of Bonsai's own repo runs code or changes its guard lines" (Rohan's list) | 5.4.7 | Rohan's part above: only the rare `update --allow-exec --yes` after a reinstall that changes Bonsai's own lines; the link itself is the orchestrator's |
| Rohan, 9 Oct, 15:35: agents manage Bonsai inside projects, "top notch", "installing, updating, fixing, checking status, editing"; the program stays his install per release | 5.4.0, 5.4.2, 5.4.4, 5.4.7 | `bonsai ladder` unattended, every refusal's `next.do` a command an agent runs as written, with `who` (5.4.2 note 12); a red climb names each failing test with the command that runs it alone; the active task's last result in `status --json`; agents link Bonsai's repo, raise its floors and edit `bonsai.yaml` by the stricter-only rule; a person's step only for the install, his (ii) and the way back, each with its reason (the (B) paragraph); proved by an agent alone climbing, reading and fixing ("5.4 done" check 14) |
| `check`'s finding for a `bonsai` on the PATH that is not the installed one waits for `install.json` (5.1.6; spec §3) | Unchanged | The pre-release has no `install.json` until 5.6's installer; Rohan's `which -a` is the check meanwhile |
| Part 0's run report quotes a studio task id; "the orchestrator may take the id out" (`STATE.md`) | 5.4.7 | Before the link or never: after it, the report is a format-0 file on the lock's list, and any change to it is a `check` finding |
| Claude Code's auto memory holds one note for Bonsai's repo, and `init` writes `autoMemoryEnabled: false` (spec §10) | 5.4.7 | The note moves into `records/memory/` as a `bonsai.memory/1` note, with its index |
| `CLAUDE.md` and `CONTRIBUTING.md` name `golang.org/x/sys` and describe the interim proof | 5.4.7 | Both rewritten for the ladder proof and the standard library only |

#### The switch, written before it happens

**Bonsai's own `bonsai.yaml`**, written by 5.4.7's builder in format 1 with a comment on every line, its id made once by
the engine's `NewID`, held to `bonsai.workspace/1` (illustrative; the builder writes the comments, and each timeout at
about three times the time it measured):

```yaml
format: bonsai.workspace/1
id: ws-<made once by 5.4.7>
name: bonsai
packs: []
documents:
  task: records/tasks
  run: records/runs
  answers: records/answers.md
  memory: records/memory
  protocols: records/protocols
protected: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**", ".github/**", "CLAUDE.md", "go.mod", "go.sum", ".golangci.yml", ".goreleaser.yaml", "design/bonsai-spec.md", "design/contract.md", "design/one-pager.md", "design/format-review.md"]
person_only: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**", ".github/workflows/release.yml", ".goreleaser.yaml"]
never_edit: []
ladder_floor: [0, 1, 2, 3, 4, 5]
ladder:
  - rung: 0
    name: guard
    kind: guard
    # ... every field of every rung written out, as the table below gives them
ratchets:
  go_tests: <the count at the switch>
  go_tests_tagged: <the count at the switch>
ci_marked_tests: []
generated:
  # ... every kind written out, with the defaults
```

| Rung | `name` | `kind` | `command` | `required` | `timeout_s` | `ratchet`; `tests` | `means` |
|---|---|---|---|---|---|---|---|
| 0 | guard | `guard` | null | true | null | null; null | Every change is inside the task's grants; no table changed on a branch; no `local/` file tracked; no `check` finding |
| 1 | vet | `command` | `"go vet ./... && go vet -tags bonsai_test_fault ./..."` | true | 300 | null; null | `go vet` finds nothing, plain or with the fault switch |
| 2 | schemas | `command` | `"go test -count=1 -run TestSchema ./formats/"` | true | 300 | null; null | Every schema changed only by additions since its set's base (contract §2.2) |
| 3 | tests | `command` | `"go test -json -count=1 ./..."` | true | 900 | `go_tests`; `go-test-json` | Every test passes, and no fewer pass than the floor |
| 4 | tests with the fault switch | `command` | `"go test -json -count=1 -tags bonsai_test_fault ./..."` | true | 900 | `go_tests_tagged`; `go-test-json` | The same, with the guard's test faults compiled in |
| 5 | git integrity | `ver-git` | null | false | 120 | null; null | What git might hide from the other rungs is listed for the verifier |

Every rung's `capture` and `base_setup` are null (Go needs no setup at the base: the module cache serves it).

- **The rungs.** Rung 0 (`guard`); rung 1, `go vet` plain and tagged; rung 2, the schema compare (contract §2.2: "A CI
  rung in Bonsai's repo fails on any schema change but additions"), which 5.4.0 makes fail rather than skip under the
  ladder; rungs 3 and 4, every test plain and tagged, read as `go test -json` so each test is named, each feeding a
  ratchet whose count is the tests that passed; rung 5, git integrity, not required. Cheapest first; `&&` reads the same
  in `sh` and `cmd`, and no command holds a quote for the shell, so the file reads the same on both sides. `-count=1`,
  so a climb's tests run in that climb. The ratchets' floors are the counts of the switch's own climb, taken after
  Rohan's install, since an installed `bonsai` changes what a few tests find on WSL.
- **The native Windows run stays beside the ladder, not a rung.** Windows Go cannot build from WSL's disk, so a rung
  would need a Windows-git clone of the commit made by a script that names this PC's Windows places, and the result
  would rest on WSL's link to Windows, which a clone elsewhere lacks. CI's `windows` job proves the same on every push.
  What lost: the result the stop gate reads covers WSL only; check 10's Windows half and CI cover Windows, as today. At
  1.0, with `bonsai.exe` installed, a climb in a Windows clone can replace check 10's Windows half; 5.7's section
  decides.
- **CI and check 10, until a rung covers them.** CI is unchanged in 5.4: all its jobs stay, and no `ladder --ci` job is
  added (it would run every test twice on each push). Check 10's WSL half (`go test` and `go vet`, plain and tagged) is
  covered by rungs 1, 3 and 4: from the switch a run report quotes the result (its path, `sha256`, the counts and each
  rung's time) instead of the raw output. Its Windows half stays: natively, before every push, as now.
- **`protected` and `person_only`.** Protected: the files that change how Bonsai is built, checked, released or told to
  work, and the four documents Rohan approved. Person-only: Bonsai's own link (`bonsai.yaml`, the lock, Claude Code's
  settings) and the release files 5.7 changes with Rohan. Not protected, on purpose: `formats/` (the schema compare and
  the manifest test hold it, and pieces change it often), `design/plan*.md` (planners write them beside builders),
  `STATE.md`, `records/` (the orchestrator's log) and every code folder (the rungs prove them). The guard's floor adds
  `.git`, `.bonsai/` and every nested `bonsai.yaml` whatever the lists say (5.3.2 rule 2), so `formats/active-task/`
  fixtures need a task's grant. `never_edit` is empty: Bonsai has no file that no agent may ever change (the closed run
  reports are held by the lock's `format0` hashes and rung 0), and a deny rule would stop the orchestrator's records
  commits too. In `agents` mode (Bonsai's, until the studio manages it) a person-only path is grantable as written
  (5.3.2 rule 4), so the rule is the orchestrator's, written in `CLAUDE.md`, and it is **stricter-only**: a task may
  hold a person-only path in `bonsai.allows` when its change only makes the guard or the ladder stricter (a floor raised
  to a green count, a rung or a protected path added) or when a section Rohan approved names it; a change that loosens
  them (a floor lowered; a rung, a floor entry, a protected or a person-only path taken out; a release file changed
  beyond its section) waits for his word, marked in the run report (contract §18 C). Why: an agent loosening the checks
  it is held to would be judging its own work. Each part's end verifier reads every change to `bonsai.yaml` in the part
  against this rule.
- **`documents`.** Tasks in `records/tasks/`, run reports in `records/runs/` (where they are), memory in
  `records/memory/`; answers and protocols name places that hold nothing yet. `STATE.md` stays at the root: contract
  §7.2 fixes STATE at `.bonsai/STATE.md`, but `init` writes it only from base's template (5.5), and the root file is not
  a Bonsai kind, so `check` does not read it. 5.5 moves it, with its frontmatter.

**Tasks and names.**
- **One task per piece**, `records/tasks/T-5xyy-<slug>.md` in format 1 (contract §4): piece 5.x.y is `T-5x` and `y` in
  two digits (5.4.7 is `T-5407`, 5.5.12 `T-5512`); a part's planning and other work outside its pieces take `T-5x90` to
  `T-5x99` in order; a piece with a letter takes the next free number, named in its title. Ids that read as their piece,
  never as a studio task's. `lane: null` (no pack declares lanes); `done_when` from the section's proof; labels
  `bonsai.branch` (the piece's branch), `bonsai.ladder` (`[]`: the floor holds every rung), `bonsai.allows` (the
  protected paths the section's "Owns" list names) and `bonsai.wants` (the person-only ones, copied into `allows` by the
  stricter-only rule).
- **The orchestrator alone writes and moves task files**, in the main checkout: `todo` when a part's section is
  approved, `running` when the piece's builder starts, `verify` when its result is green, `done` once it has landed with
  CI green (or `cut`). The moves stay uncommitted in main's working tree until its records commit after the landing,
  with the run report, `STATE.md` and the tables from `bonsai check --write`. So a builder's branch never needs a rebase
  for a status move, and grants are read from main's files as contract §13 asks. Builders never edit a task file. The
  one exception is `T-5407`, the first, which the switch's own commit brings before the rule exists. **Every file in
  `records/tasks/` is a task:** one that does not parse makes the active task none for every reader (contract §13, fail
  closed), so no template or note lives there.
- **Builders stay subagents of the orchestrator's session, unnamed.** A subagent runs inside its session's process and
  shares its environment, so it cannot carry a `BONSAI_TASK` of its own. So: the guard grants by contract §13's step 2
  (the one task reading `running`); the ladder is named by `--task`; and the stop gate, on Stop only, does not run for a
  subagent (Claude Code fires SubagentStop), nor for the orchestrator's own session, which never has `BONSAI_TASK` set.
  **The stop gate cannot trap the orchestrator or its builders.** Their gate is the orchestrator's landing rule (below).
  Two tasks `running` side by side make no task active, so the guard refuses a protected path to both, with that reason:
  a task whose `bonsai.allows` is not empty runs with no other task `running`, and the section's order names when. Rare:
  the protected list is short, and most pieces touch none of it.
- **A builder started as its own session is named and gated.** One started from the orchestrator's shell (`claude -p`
  through `claude-here`, in its worktree) gets `BONSAI_TASK=<id>` and `BONSAI_ROLE=builder` in its environment: the
  guard grants by its task even beside other running tasks, its records and sessions rows carry the task, and the stop
  gate keeps it from stopping until its result is green at its HEAD, up to Claude Code's cap of eight blocks (5.3.4 note
  5). Today's way, subagents, stays the default; the end verifier proves this one once.
- **The landing rule** (`CLAUDE.md`, from the switch): the orchestrator fast-forwards a piece only when
  `.bonsai/local/ladder/<task>.json` reads `mode: local`, green, `git.sha` the branch's HEAD, `git.dirty` false, every
  rung of the floor and the task's `bonsai.ladder` green, and its `ladder` log record names `/usr/local/bin/bonsai` with
  the result's `sha256`; plus check 10's Windows half, and CI after the push. The task moves to `done` once CI is green.
  The ladder is always run by its installed path, `/usr/local/bin/bonsai ladder --task <id>`, so no other `bonsai` on
  the PATH can make a proof. A builder climbs once its work is done, and again after a fix, not after every edit: each
  climb's minutes count in its run.
- **The floors, once a part.** Bless needs a green run in the main checkout, on the base branch, at its HEAD, with a
  clean tree (5.4.4), and the orchestrator's open run report and task moves keep the main checkout's tree dirty while
  pieces run. So at a part's end, after its last records commit, the orchestrator climbs `main` once under the part's
  last task: that run checks `main` as the part leaves it and files a Bless ask when the part's work raised a count. The
  orchestrator then raises the floors to that run's counts itself, under a task (`T-5x9N`) granting `bonsai.yaml`, by
  the stricter-only rule; records the ask's numbers (new tests that failed at the base, proved nothing new, or could not
  run) in its run report; and gives Rohan the numbers in the part's last line, with nothing to answer. The next climb on
  `main` resolves the ask (5.4.4 note 3). Chosen over a climb of `main` after every landing, which would need a records
  commit before each and run every test again for a floor that rises once a part anyway; what lost: Bless's new-tests
  numbers cover the part's last task, while every piece's own green climbs carry its own for the verifiers.
- **Run reports** stay the orchestrator's, one per task, in the main checkout. From the switch each carries
  `bonsai.run/1` frontmatter (contract §7.1: `id` `R-<date>-<task id>`, `task`, `role: orchestrator`, `model`,
  `started`, `finished`, `outcome`, `commits`, `labels`), and its file is `R-<date>-<task id>-<topic>.md` (in place of
  "What changes", item 3's `R-<date>-<piece>-<topic>.md`), so 5.3.4's check for a report naming the task finds it. Every
  report from before is closed and committed before the link, whose `init` lists them on the lock's `format0` list with
  their hashes: they are frozen from then, and the orchestrator's last edit of part 0's report (the studio task id)
  happens before it or not at all. If `check` holds the old reports to more than their hash, `documents.run` names a new
  folder instead and the switch's run report says why.

**Labels with no definition in force until `base` (5.5).** Bonsai's repo links no pack at 5.4, so no definition of
`bonsai.allows`, `bonsai.wants`, `bonsai.ladder` or `bonsai.branch` is in force (5.1.5 puts base's in force in 5.5). The
four readers that act on them (the guard, rung 0, the runner and the stop gate) read them by contract §5.6's meaning
whether a definition is in force or not: a value of another kind (a grant that is not a list of text) grants or adds
nothing. `check` judges a label against its definition only where one is in force, as contract §5.2 has it, and an
undefined label is neither a finding nor a warning (no rule names one). The instruction block shows no definition;
`CLAUDE.md` gives the four in a line each until base's block does. Chosen over linking a stand-in `base` early, which is
5.5's work and a plugin install, and over attaching them on the machine, which 5.6's command does and which contract
§5.1 rules out for `bonsai.*` (reserved for Bonsai and the packs it publishes). Note 5.4.3 says what changes if 5.3's
guard grants only from a defined label.

**`CONTRIBUTING.md`**, rewritten: Bonsai guards its own repo. Its committed `.claude/settings.json` runs the installed
`bonsai` (`C:\Program Files\Bonsai\bonsai.exe`, then `/usr/local/bin/bonsai`) on every edit and shell command in a
Claude Code session, and refuses the call when neither is there, on purpose (fail closed): a guard that a missing
program switched off would guard nothing. Git, editors and Go are not affected. To work here with Claude Code: build
Bonsai from this repo and install it at that place, or start Claude Code with Bonsai's hooks off for your clone
(`claude --settings '{"disableAllHooks": true}'`, or `.claude/settings.local.json` holding `"disableAllHooks": true`,
which `bonsai check` reports). The proof of a change is `bonsai ladder --task <id>` green, run by the installed
`bonsai`, plus CI on both sides; the standard library only (Windows calls through `syscall`); never `go install`.

**`CLAUDE.md`**, changed at the switch (its own task grants it; Rohan's approval of this section is his word for it):
the "Proof" line becomes the landing rule above; agents first: a builder runs the ladder, reads its `--json`, fixes what
is red and climbs again, and asks a person only where a refusal's `who` says so; the stricter-only rule for
`bonsai.yaml` and person-only paths; "How a session works" gains the tasks, the names and the rule for running tasks
side by side; the run reports' frontmatter; the four labels' lines; the (ii) commands that are Rohan's; the way back;
the removal of a worktree only with `git worktree remove` (a `rm -rf` of a checkout's folder is a guarded delete);
`bonsai update` of Bonsai's own link only in the main checkout (rung 0 refuses a branch that changes
`.claude/settings.json`); and the Go line without `golang.org/x/sys`. The block `init` writes sits apart, between its
markers.

**`.gitignore`** gains `.claude/settings.local.json`, Claude Code's own convention: the way back's file and Rohan's
permission choices stay out of git, and a main checkout holding one still climbs clean.

**`records/memory/`**: `INDEX.md` and one note, the hand-checks preference Claude Code's auto memory holds for this repo
(in Rohan's words), as `bonsai.memory/1`. `init` then imports the index in the block, and `autoMemoryEnabled: false`
loses nothing. The orchestrator gives the builder the note's words: no agent reads Claude Code's memory folder. The note
names no path and no person's detail beyond that preference.

**How Bonsai links itself, under Rohan's (a) and (ii).** After Rohan's install, the orchestrator links Bonsai's repo
with the installed copy, in the main checkout: first `/usr/local/bin/bonsai init --json` (the preview: every settings
line with its sentence; nothing written; exit 4 without `--yes`), each line read against this section, then
`/usr/local/bin/bonsai init --yes --json`. Under (a) the lines it writes name `C:\Program Files\Bonsai\bonsai.exe`, then
`/usr/local/bin/bonsai`, and the PATH is never read; a stamped test build cannot write them here, since it refuses to
link outside its scratch root (5.3.6 note 2), so only the installed build links the real repo. It needs no
`--allow-exec`: `packs: []` brings no pack code, and Bonsai's own lines at a first link are written on `--yes` alone
(5.1.1 rule 6). (ii) leaves it to an agent: a first link puts the guard in, and the command refuses only `--allow-exec`,
`unlink` and an update that removes or changes Bonsai's own lines (5.3.6 note 8). Chosen over Rohan typing it: no safety
reason needs a person to put a guard in, and his 9 Oct direction gives linking to agents. From then on, in Bonsai's
repo: an update that adds no code and leaves Bonsai's lines as they are (base's link in 5.5) is the orchestrator's, in
the main checkout, under a task granting `bonsai.yaml` and the lock; `--allow-exec`, `unlink` and an update changing
Bonsai's own lines are Rohan's, typed in his own terminal, until the studio manages Bonsai's repo.

**The order of the switch.**
1. 5.4.7's builder prepares its files in its worktree (an agent tries the way back in a scratch session, note 5.4.7),
   links a scratch clone of its branch with the final code's stamped build (a scratch home, inside the scratch root),
   climbs it, and reads the counts and times (the floors and timeouts come from these). It lands on the interim proof,
   the last landing that does: check 10 and CI.
2. Before the build: `main`'s working tree holds nothing uncommitted but 5.4.7's open run report, which is format 1;
   every other run report is closed and committed; part 0's report has been edited or left; Claude Code's version and
   the user settings hashes are recorded.
3. The orchestrator builds the pre-release ("Rohan's sitting", below) and sends Rohan his batch.
4. Rohan installs it and sends `bonsai --version`'s line.
5. The orchestrator links the repo (above) and reads what `init` wrote (`git status`, `git diff`: the lock,
   `.claude/settings.json` with every line in (a)'s form, the block in `CLAUDE.md`, `.bonsai/.gitignore`, the two
   tables), runs `bonsai check --json` (no finding), commits it on `main` (`bonsai: Bonsai links itself`), pushes and
   reads CI. If its session has not taken up the hooks (a subagent's file-tool edit of `bonsai.yaml` is not refused), it
   asks Rohan to restart it.
6. The first climb: after a records commit, with the main checkout clean, `/usr/local/bin/bonsai ladder --task T-5407`
   in it is green, and its result, its `ladder` record and `hook start`'s next context agree. `T-5407` moves to `done`
   in a records commit with the tables.
7. The orchestrator writes `T-5490` (5.4's end verification, running) and commits it with the verifier's opened run
   report, so a gated session in a fresh worktree finds both; the end verifier runs "5.4 done". Then `STATE.md` is
   rewritten and 5.5's planning starts under the ladder proof.

#### Notes per piece

**5.4.0, formats set 6 and the ladder's lists.** One commit to `formats/` with its manifest, at the set after the last
landed (set 5 is 5.2.0's; 5.3 changes no format), as `formats/README.md`'s "How the set changes" asks; additions only,
so the schema-compare test passes.
1. **`bonsai.climb/1`**, the `--json` of `bonsai ladder` (5.1.3: "Later parts add their commands' outputs (asks, logs,
   ladder) the same way"): `format`; `workspace` (null when it refused before reading one); `task` (or null); `mode`;
   `result`, the project-relative path written, or null; `sha256` of its bytes, or null; `ladder`, the result as written
   (`bonsai.ladder/1`), or null; `bless`, the ask record this run wrote (`bonsai.ask/1`, a `file` or a `resolve`), or
   null; `cleaned`, the project-relative paths cleaned after the run; `next`, `{do, who}` as the `error` object's, the
   step after a red run (5.4.2 note 12), or null when green; `error`. A new schema with an example. Named for the act,
   since `bonsai.ladder/1` is the result's name.
2. **A rung's `ratchet`, `capture` and `tests`**, fixed by their descriptions in `workspace.schema.json` (a description
   change is an addition) and held by 5.4.2's reader, as today's ladder has them:
   - `ratchet`: the ratchet's name (`[a-z_][a-z0-9_]{0,39}`, today's), naming a floor in `ratchets` (none reads as 0),
     or null.
   - `capture`: a map from a name (the same form) to a pattern, a Go regular expression (RE2) matched line by line
     against the rung's output, whose first group is the value (a number when it reads as one), or null. Today's
     patterns read the same in RE2; one that does not compile is a `check` finding naming it.
   - A ratchet's count is the rung's capture of the same name; on a rung with `tests` and no such capture, the number of
     tests the output names as passed. A ratchet with neither is a `check` finding, and red at run time.
   - `tests`: an open list in a Go table, `go-test-json`, `tap` (node's test runner) and `junit` (read from the rung's
     standard output), each with one line on what it reads. An unknown word: the tests are not read, and `check` warns
     naming the known ones (5.4.2's rows).

   Chosen over a type in the schema, which the compare refuses inside a major ("a `type` ... added"), so it would need
   `bonsai.workspace/2`, a new major and Rohan's call (contract §2.2), for a shape today's files already have. What
   lost: a reader holding only the JSON Schema (the studio's) accepts a shape Bonsai's `check` refuses. The README's
   "Typed open" entry is rewritten to say so.
3. **The result's additions:** `passed_on_base` and `not_run` at the end of `new_tests` (counts beside `count` and
   `failed_on_base`; 5.4.5); the `captures` description gains "for a `ver-git` rung, what it found, by name" (5.4.6); a
   rung `status`'s description names the words' table; the `tests` description (typed open) names a failed test's `run`,
   the command that runs it alone (5.4.2 note 12). No field is removed or retyped.
4. **The log's `ladder` record**, by description (contract §11: "the result's project-relative path, `sha256:<hex>` of
   its bytes, its `green`, and `task`"): `target` the path, `text` `sha256:<hex>`, `ok` its `green`, `task` its task,
   `kind` its `mode`, `category` `Ladder`, and `bonsai_path` and `bonsai_sha256` the runner's own (so the record names
   which `bonsai` made the result).
5. **Go tables**, beside the ladder's type, each word with its line: the `tests` forms; rung statuses (`green`, `red`,
   `skipped`, `error`, and `pending`, which today's results hold and Bonsai never writes: no rung field marks one as not
   built); git integrity's findings (5.4.6's names). Printed by `check --schema` and the reference page; the schemas'
   descriptions name the command and copy no word (5.1.3's rule).
6. **The error words 5.4 needs**, added here so no later piece edits the words' table, reusing 5.1.4b's and 5.3's where
   one fits: another ladder holds this home's lock past its wait; no task named for a local run, or the named task is
   none (with the function's reason code); a requested rung not in `bonsai.yaml`, or no ladder there; the main checkout
   not verified, so no result can be written where the stop gate reads; a stamped build's ladder outside its scratch
   root. The builder fixes the words.
7. **The schema compare under the ladder:** `compare_test.go` fails, as under `CI`, when `BONSAI_LADDER` is set (the
   marker 5.4.1 puts in every rung's environment), so rung 2 can never pass by skipping.
8. `climb` wherever the formats are listed (the registry, the test listing every schema, `status --json`'s `formats`:
   `read` `[1]`, `write` 1); `docs/reference/lists.md` regenerated.
9. **`status --json` gains `ladder`** at the end of its properties (an addition): the active task's last result,
   `{task, path, green, mode, sha, finished, stale}`, `stale` true when `sha` is not this checkout's HEAD; null when
   there is no active task or no result. So an agent reads where its proof stands without running anything (Rohan's 9
   Oct "checking status"); 5.4.2 fills it.

**5.4.1, rung jobs.** Spec §9: "each in its own process group on Linux and its own job object on Windows, so a rung
cannot leave a process for a later one ...; the leftovers line at the end". One function runs one command with a folder,
an environment, a timeout and a stop signal, and returns its exit, its output (standard output apart, for the test
forms) and what it ended.
1. **Linux.** The command runs as `/bin/sh -c` in a new session and process group (`Setsid`), with `Pdeathsig` so the
   shell dies with the runner, and a random tag in its environment (`BONSAI_RUNG_TAG`, added to any outer rung's). The
   runner marks itself a child subreaper (`prctl`), so a process whose parent exits is handed to the runner, not to
   init. At the rung's exit, its timeout or a stop, it lists from `/proc` the rung's session and group, every process
   whose environment carries the tag, the runner's adopted children and every live descendant of those; ends the group,
   then each; and lists again until none is left (2 s at most). A pid is acted on only while its start time is the one
   listed; a group only while it has a member. Chosen over a process group alone (the spec's words), which a `setsid` or
   `setpgid` child leaves, as today's ladder found; what lost against today's: its keeper, so a runner killed by
   `SIGKILL` leaves what escaped its shell, which the next run's leftovers line names. The known limit, as today: a
   process that clears its environment, leaves the session and whose parent is gone cannot be told from a stranger
   without root.
2. **Windows.** A job object made by the runner with kill-on-close and no breakaway, its handle not inherited; the
   command (`%ComSpec% /d /s /c "<command>"`, today's line) is started suspended, put in the job, then let run, so no
   process of the rung starts outside it. At the end, its timeout or a stop, the runner lists the job's members (their
   pids and image names), ends the job, and reports them. If the job cannot be made, the rung runs without one and says
   so (`job: off: <why>` in its reason), never silently. The calls go through `syscall`'s lazily loaded procedures
   (`CreateJobObjectW`, `SetInformationJobObject`, `AssignProcessToJobObject`, `QueryInformationJobObject`,
   `TerminateJobObject`, and a resume of the suspended process), so `golang.org/x/sys` stays out, as 5.3 decided. No
   PowerShell launcher (today's costs about 150 ms a rung to compile its calls).
3. **Everything else:** the rung's environment gains `BONSAI_LADDER=1` (a nested runner prints no leftovers line; the
   schema compare fails rather than skip, 5.4.0); output kept whole for the test forms up to a cap the builder sets and
   names, and its last 40 lines for the reason; a timeout ends the rung red with "timed out after N s; ended M
   processes"; the default timeout is 600 s, as today, unless the Defender measure asks for more.
4. **The leftovers line** (read-only, never a colour): processes still running from this checkout that the runner did
   not start, at the run's end, each by pid, program name and the head of its command (5.2.1's reduction, never a whole
   command line, which may hold a secret); the runner, its ancestors and its descendants left out. Linux from `/proc` (a
   process whose working folder, program or arguments lie in the checkout); Windows from a process snapshot
   (`CreateToolhelp32Snapshot`) and each image's path, with the command line where the system gives it cheaply. The
   result's `leftovers` is `{checkout: [...], ms}`, `{unknown: <why>, ms}`, or `not checked (<platform>)`; today's
   `shared_deploy` group is the studio's deploy kit's and is not kept. Skipped under `--ci` and in a nested ladder.
5. **Measured, in the run report:** a rung's added cost on each side (the job or the group, the listing, the end); on
   Windows a detached grandchild (`start "" /b`, PowerShell's `Start-Process`, `cmd /c start`), a child asking to break
   away (refused), a job inside the rung's job (a rung running Bonsai's own runner tests), the runner killed with
   `taskkill /f` (the job ends the tree); **Defender's first scan**: a freshly built `bonsai.exe` started ten times
   after its build and ten times later, and a package's tests under `go test` with a fresh test binary each time; on
   Linux a double fork, a `setsid` daemon, a `nohup` child, and a process holding the rung's output pipe after its shell
   exits. Each figure against the gate report's section 2.5 where it has one.

**5.4.2, the runner.** `bonsai ladder --task T [--root P] [--ci] [--json]` (spec §4).
1. **Which task and where.** `--root` is the checkout to run in (default: the one holding the working folder); tasks are
   read from its verified main (5.3.1) by the active-task function with `--task` and `BONSAI_TASK` (contract §13: the
   environment only for the session's own project; disagreeing names give none). A local run needs a task: none, or a
   named task the function cannot find, exits 2 with the reason. `--ci` may name none.
2. **Which rungs.** Requested: the floor plus the task's `bonsai.ladder`, sorted, each once (contract §5.6); with `--ci`
   and no task, every rung of the ladder. A requested rung `bonsai.yaml` does not define exits 2 before anything runs.
   Rungs run in number order; a red `required` rung marks the rest `skipped` ("an earlier required rung failed"); the
   run is green only when every rung it took up is green (`pending` never occurs: no rung field marks one). Rung 0 and
   `ver-git` call 5.4.3 and 5.4.6 through `kinds.go`; a command rung runs through 5.4.1.
3. **Test output** (`internal/testout/`): `go test -json` events (each test and subtest by its full name, passed,
   failed, skipped with its reason; a package that does not build names no test and marks the rung red), node's TAP
   (today's `parseTap`: subtests joined by `" > "`, SKIP and TODO with their reasons, a failure's file and error), and
   JUnit XML (`encoding/xml`: `classname` and `name`, failures, errors, skips). The result's `tests` keeps today's shape
   (`seen`, `skipped`, `todo`, `failed` with its first 50 and `failed_count`).
4. **Skips and marks, as today** (`ladder.mjs`'s `judgeTests`): a mark is a skip reason starting `<word>-only:`. In a
   local run, a skipped test with a mark, or one named in `ci_marked_tests`, makes its rung red (a task's proof runs
   them all); an unmarked skip passes. Under `--ci` a skip counts only when its name is in `ci_marked_tests` and its
   reason carries a mark (Bonsai's list holds names, not today's `{name, mark, file}`); any other skip, and any todo, is
   red. `skipped` is filled under `--ci` only.
5. **The ratchet's count** (5.4.0 note 2) is set on the rung; 5.4.4 judges it.
6. **The result**, written through 5.1.4a's writer into the verified main's `.bonsai/local/ladder/<task>.json` (`--ci`:
   `ci.json`), wherever the ladder ran, by a whole temporary file renamed into place (Windows' busy retry). `git`: HEAD,
   its branch, and `dirty` (git's status, plus what 5.4.6 finds git hiding). Every free-text value (a reason, a failed
   test's error, a capture's text) passes 5.2.1's redactor before it is written. A main that is not verified writes no
   result (exit 4: "a person runs `bonsai update` in the main checkout"). `--ci`: `mode: ci`, today's `proof` sentence,
   no Bless, no leftovers line; it runs anywhere (today's checks for GitHub Actions only; a variable anyone can set
   proves nothing, and `mode` alone keeps a CI result from proving a task).
7. **The fingerprint:** after the result, one `ladder` record (5.4.0 note 4) through 5.2.2's append path, in the
   session's file when `CLAUDE_CODE_SESSION_ID` is set, else today's day file.
8. **One ladder at a time per home** (spec §9): `<home>/locks/ladder.lock`, held by an operating-system lock that ends
   with its holder (`flock` on Linux; on Windows the file opened with no sharing), and holding the holder's task,
   checkout, pid, start time and its budget (the sum of its requested rungs' timeouts). A second run prints the holder
   and waits while the holder lives and is inside its budget, then refuses (exit 4, naming the holder and the next
   step). A dead holder's lock is taken at once. The orchestrator's rule (one ladder at a time on this PC) still covers
   WSL's and Windows' homes together.
9. **After the run:** 5.2.6b's cleaner for the `ladder` kind (never the result just written); the human output (ASCII:
   one line per rung with its status, time and ratchet, the red rungs' reasons and failing tests, the leftovers line,
   the result's path); `--json` as `bonsai.climb/1`. Exit codes: 0 green, 1 red, 2 bad input, 3 could not run or write,
   4 wrong state.
10. **A stamped build** (5.3.6 note 2) refuses to write a result outside its stamped scratch root, as its `init` does: a
    test build never makes a proof in a real project.
11. **`check`'s rows for the ladder's shape**, in 5.1.6's table: two rungs with one number; a floor or a ratchet naming
    nothing; a capture pattern that does not compile; a ratchet with no count; a command calling `bash` by name.
12. **Agents first** (Rohan, 9 Oct: agents manage Bonsai inside projects, and the support is to be "top notch"). The
    ladder runs unattended end to end: no prompt, `--json` always whole, the exit code its colour. Every refusal and
    error carries the `error` object; its `next.do` is a command an agent can run as written, with every value Bonsai
    knows filled in (the task id, the rung, the path, the holder's pid), or, where the fix is an edit, the file and the
    change named and then the command that confirms it; `next.who` is `agent` but where a safety reason makes the step a
    person's, which the words' table says. The refusals: no task named (`bonsai status --active --json`, then
    `bonsai ladder --task <its id> --json`, the id filled in when one task runs); the lock held (the same command again
    once the holder's pid has ended, with its budget's end); a rung not defined (`bonsai check --json`, which names the
    rung and the line); a stamped build outside its scratch root (`/usr/local/bin/bonsai ladder --task <id> --json`);
    the main checkout not verified (as 5.3.1 sets its step and `who`). **A red climb says how to fix it:** `climb`'s
    `next` names the first red rung's fix and then the climb again; each failing test of a `go-test-json` or `tap` rung
    carries in `tests.failed[].run` the command that runs it alone, quoted for the shell the rung runs in
    (`go test -count=1 -run '^TestX$/^sub$' ./pkg`, or `node --test --test-name-pattern=...` with its file); each path
    rung 0 refuses comes with its rule and the grant or revert it needs; a ratchet below its floor with both counts.
    `who` is `agent` there too, but for a person-only path in `command` mode (a person's grant). `status --json`'s
    `ladder` (5.4.0 note 9) tells an agent where its proof stands without a climb.

**5.4.3, rung 0.** Spec §9: "the diff against the named task's grants"; "Rung 0 also refuses a task branch that changes
a generated table, and any file from `.bonsai/local/` that is tracked or staged"; spec §4: "CI and rung 0 run
[`check`]".
1. **What changed:** on a branch other than the base, every path the branch's commits changed since the merge base
   (`git diff --name-only <base>...HEAD`, both sides of a rename), and on any branch what is staged, unstaged or
   untracked and not ignored (today's `changedFiles`). **The base branch** is the branch the verified main checkout has
   checked out, read from its HEAD file (5.2.2's reading): Bonsai's schema has no `base_branch`, and Bonsai's work and
   the studio's both land on the branch their main checkout holds. A main on a detached HEAD gives no base: the branch
   part is "not judged" in the reason, and the working tree is still judged.
2. **Each path judged by 5.3.2's one function** as an edit by the named task, with its grants at any status (contract
   §13: rung 0 judges finished work), so rung 0 and the guard can never disagree: a protected path the task does not
   grant, a person-only one, the floor's (`.claude/settings.json`, `.claude/settings.local.json` and `.git` never
   granted; `bonsai.yaml`, the lock, a nested `bonsai.yaml` or `.bonsai/` granted only as rule 3 says), and Bonsai's own
   files. None named (`--ci` with no task): no grants, so any protected change is red (contract §13: "None fails closed
   everywhere").
3. **Bonsai's own refusals:** on a branch other than the base, a changed `.bonsai/tasks.md` or `.bonsai/sessions.md`;
   anywhere, a tracked or staged file under `.bonsai/local/`.
4. **`check`'s findings** in the checkout it runs in, from 5.1.6's table (its one home), but those about this machine
   rather than the commit (plugin drift, a pack not installed, the PATH's `bonsai`), which the reason lists as "not
   judged here": a proof is of a commit. Warnings never count.
5. **Its outcome:** green with the count of paths checked (in `captures`, `checked`); red with each refused path and its
   rule, at most 50, in the reason. The result never holds a whole diff.
6. **`bonsai.*` labels with no definition in force:** 5.4's start reads how 5.1.5 and 5.3.2 landed. If the guard and the
   stop gate already read `bonsai.allows` and `bonsai.ladder` by contract §5.6's meaning without a definition, this
   piece reads them through the same code. If they grant only from a defined label, this piece adds one function, the
   value of a `bonsai.*` label on a task held to §5.6's kind with or without a definition, used by all four readers, and
   V1 covers the guard's change (guards and hooks).
7. **Proof:** the `rung0` section of set 4's fixtures, every case, the worktree cases built with git (Windows git on
   Windows); fixture repositories for each refusal (a protected path granted, not granted, granted while the task reads
   `verify`; `.claude/settings.json` changed under a grant; a table on a branch and on the base; a staged `local/` file;
   a `check` finding); a table of paths judged by the guard and by rung 0 with the same answer.

**5.4.4, floors, ratchets and Bless.** Spec §9: "a green `local` result on a clean base branch at HEAD with a count
above its floor files one Bless ask. The floor rises only when the studio applies Rohan's tap."
1. **The ratchet:** `was` is `ratchets[name]` (0 when absent), `now` the rung's count; `ok` when `now` is at least
   `was`; below it the rung is red ("`go_tests` went backwards: 812 -> 805"), even with every test green. A count that
   cannot be read is red too: a ratchet that measures nothing is none. Under `--ci`, accepted marked skips count, as
   today (`pinned_skips` is today's field and is not kept; the reason says how many).
2. **The offer** (today's `blessOffer`): the run is local and green; it ran in the main checkout on the base branch at
   its HEAD with a clean tree; and at least one green rung's ratchet has `now` above `was`, its name a key of
   `ratchets`. Then the runner files one ask through 5.2.5's function: key `ladder:bless-<task>`, `source: ladder`,
   `type: Bless`, `task`, a title ("Bless <task>'s ratchets") and `why` naming each rise and its new-tests numbers
   (5.4.5), `then` "Blessing raises the floor in `bonsai.yaml` from that run; it runs nothing", and `data`
   `{task, head, result_sha, rises: [{name, was, now, new_tests}]}` (contract §9.1). Filing an open key again writes a
   new `file` (5.2.5 note 4), so a later run's numbers replace the earlier ones.
3. **Resolved when the offer goes**, as today's bridge does: a run on the base branch, whatever its task, works the
   offer out again for every open `ladder:bless-*` ask from that task's result and the floors as they now are, and
   resolves each whose offer is gone, with its reason (HEAD moved, the run red or dirty, no rise left).
4. **Who raises a floor.** Bonsai's code never writes `ratchets`. Where the studio manages a project, it applies a
   person's `bless` tap (contract §10.3). Elsewhere an agent may raise a floor to a green, clean run's count, under a
   task that grants `bonsai.yaml`, since a raised floor only makes the ladder stricter; lowering one waits for a person
   (the switch's stricter-only rule, which 5.5's `workflow` pack carries for other projects). In Bonsai's own repo the
   orchestrator does it at a part's end (the switch, "Tasks and names"); HEAD has moved by then, so the next climb on
   `main` resolves the ask (note 3).
5. **Never:** under `--ci`, on a red or dirty run, on a branch, or from a result another run wrote.

**5.4.5, new tests must fail on the base.** Spec §9: "A new test counts as failed on the base only when the base run
reports that test by name as failed while the old tests ran and passed. A setup, import or load failure is 'check not
run', never 'failed'."
1. **When:** on a local run that is green and clean, for each ratchet whose `now` is above `was`; on any other run
   `new_tests` is null. So a builder's climbs while it works cost nothing more, and the run the stop gate accepts and
   the run that offers Bless both carry the numbers.
2. **The base:** the merge base of HEAD with the base branch (5.4.3 note 1). On the base branch itself, where that is
   HEAD, the base recorded by the task's last result on its branch (5.4.6); none recorded: "check not run: no base".
3. **The two base runs**, in one temporary worktree at the base (`git worktree add --detach` in the system's temporary
   folder, removed after, `git worktree prune` on a failure), each through 5.4.1 with the rung's own timeout: first
   `base_setup` if any, then the rung as it is (the base's test names, and that they pass); then the branch's changed
   test files copied in and the rung again. **New tests** are the names the HEAD run reports that the first base run
   does not. **Failed on the base:** a new test the second run reports failed by name, while every test of the first run
   that the second also ran passed. **Proves nothing new:** a new test the second run reports passed. **Check not run:**
   a new test the second run does not report (its package did not build, its file did not load), every new test when
   `base_setup` failed or an old test failed, and every one when the base could not be made. A Go test that calls code
   the branch adds does not build at the base, so it reads "check not run", which is right: it proves nothing about the
   base.
4. **Test files, by the form's own convention:** `go-test-json`, the `*_test.go` files and files under a package's
   `testdata/`; `tap`, node's test runner's default patterns (`*.test.*`, `*-test.*`, `*_test.*`, `test-*.*`, `test.*`
   and files under `test/`, for `.js`, `.mjs`, `.cjs` and `.ts`); `junit`, the files the XML names; none found: "check
   not run: the test files cannot be told".
5. **Written:** `new_tests` `{base, count, failed_on_base, passed_on_base, not_run}` on the ratchet (5.4.0 note 3); the
   names, at most 20 per group, in the human output and the Bless ask's `why`. The rise is offered for Bless whatever
   the numbers (spec §9: "ratchets are Rohan's").

**5.4.6, git integrity.** Spec §9: "records the task's base commit and flags changed test files, `assume-unchanged` and
`skip-worktree` entries, edits to `.git/info/exclude`, new stash entries and a rewritten base. Not required at first: it
informs the verifier."
1. **What it records**, in the rung's `captures` (5.4.0 note 3): `base` (the merge base, 5.4.3 note 1), and the stash
   count and `info/exclude`'s hash, so the task's next climb can compare.
2. **What it finds**, each by name and count, the names (at most 20 a kind) in the human output: changed test files
   (5.4.5's convention); `assume-unchanged` and `skip-worktree` entries (`git ls-files -v`); lines in the common git
   folder's `info/exclude` other than git's own comments; stash entries the task's earlier result did not count; a base
   the earlier result recorded that is no longer an ancestor of HEAD.
3. **It informs; it never turns red for what it finds.** Its status is `green` with its findings listed, or `error` when
   git cannot be read. A finding that hides a change from git also makes the runner's `dirty` true: an
   `assume-unchanged` or `skip-worktree` entry, or a file ignored by `info/exclude` alone. Otherwise a test run on a
   change git does not show would claim a clean commit, so the stop gate (which refuses `dirty`) catches it while this
   rung stays informative. Chosen over a red rung, which the spec defers, and over leaving `dirty` to git's status,
   which those entries fool.

**5.4.7, the switch.** Written above, in "The switch, written before it happens". The builder's proof (notes 1 and 2 run
by a Sonnet agent, as "Who builds and verifies" says):
1. **The way back, tried:** in a scratch project linked by a build stamped with an empty place (every call refused,
   5.3.6's `missing`), a session started through `claude-here` with `--settings '{"disableAllHooks": true}'` runs a free
   edit and a shell call; a second session with `.claude/settings.local.json` holding the same; `bonsai check` reports
   that file. The Claude Code version and the user settings hashes recorded. If `--settings` does not switch the
   project's hooks off on the version in use, the file is the only way back offered and Rohan's part says so before his
   sitting.
2. **Settings taken up mid-session:** in a scratch session, whether a `.claude/settings.json` with hook lines written
   after the session started is applied in that session, applied after review, or only at the next start. The
   orchestrator's step 5 above follows the answer.
3. **A scratch climb of the switch:** a clone of the branch in the scratch root, linked by the final code's stamped
   build on a scratch home (`init --new-id`, `packs: []`, no `--allow-exec`), `check` with no finding, then
   `ladder --task T-5407` green, every rung, with the counts and times the floors and timeouts take.
4. **The files:** `bonsai.yaml` validates; every line commented; the task file and the memory note parse under their
   formats; nothing private in any of them.

#### Rohan's sitting (spec §17 step 8)

**Before it, the orchestrator** builds the pre-release from a clean clone of `main` at the switch's commit, never from a
worktree (gate report §2.11; "What changes", item 7):

```bash
git clone ~/Servers/Bonsai ~/bonsai-checks/prerelease/src-<commit>
git -C ~/bonsai-checks/prerelease/src-<commit> checkout --detach <commit>
cd ~/bonsai-checks/prerelease/src-<commit> && go build -trimpath -buildvcs=true -o ~/bonsai-checks/prerelease/bonsai ./cmd/bonsai
go version -m ~/bonsai-checks/prerelease/bonsai
sha256sum ~/bonsai-checks/prerelease/bonsai
```

`go version -m` must show that `vcs.revision` and `vcs.modified=false`, and no fault tag. `-trimpath` keeps the build's
folder out of the binary and makes it the same bytes wherever it is built, so the end verifier's own build from its own
clone matches the fingerprint. The commit's Go code must be the code V2 passed
(`git diff --stat <V2's commit>..<commit> -- '*.go'` empty). Before a later pre-release, the build Rohan has is copied
to `~/bonsai-checks/prerelease/previous/` first. Then the orchestrator sends Rohan the batch in his part above, with the
number in place, in one message.

**After it:** his words and `bonsai --version`'s line go in 5.4.7's run report; the orchestrator runs the switch's steps
5 and 6.

#### Proof for each piece

Every piece: its Go tests and `go vet`, plain and with the fault tag, on WSL and natively on Windows (check 10) before
the push, their counts in the run report; CI green on the pushed commit; no Windows-only skip without a named reason (a
job-object test runs on Windows only and says so; a `/proc` test on Linux only); the Windows rules of `CLAUDE.md` read
in the diff (forward slashes in every stored and printed path, byte-stable output, busy retries, no `bash` by name, no
test that needs a symbolic link or a file mode). Tests that start process trees stop every process they start, and the
builder's `ps` check names any left. Pieces that run Claude Code (5.4.7, the end verifier) record its version and the
user settings hashes before and after. Scripted runs live in `~/bonsai-checks/scripts/`, never committed. V1, V2 and the
end verifier re-run the tests themselves, on both sides. From the switch, 5.4.7's own last climb is by the installed
`bonsai`.

#### 5.4 done

V2 runs checks 1 to 15 on the last code commit, on both sides, before the pre-release is built. A fresh Opus verifier,
at the end of 5.4 and after Rohan's sitting, runs checks 16 to 23 itself on the final commit and the installed binary,
reads V2's report for 1 to 15 (the build's rule keeps the Go code the same), and passes or fails 5.4:
1. **Set 6:** the manifest matches every byte; `bonsai.climb/1` documents itself with an example; `new_tests` ends with
   its two counts; the schema-compare test passes, fails on a removal in a temporary copy, and fails rather than skips
   with `BONSAI_LADDER` set and no base; `check --schema` prints the `tests` forms, rung statuses and git integrity's
   findings from their tables; `go generate` changes nothing under `docs/reference/`.
2. **Rung jobs, Linux:** a rung leaving a `setsid` daemon, a double fork, a `nohup` child and a process holding its
   pipe: each ended at the rung's end and at its timeout, and named; a decoy started outside the ladder, with the same
   command, left running; the runner killed with `SIGKILL` mid-rung: the shell ends, what escaped is named by the next
   run's leftovers line.
3. **Rung jobs, Windows natively:** `start "" /b`, `Start-Process`, `cmd /c start` and a child asking to break away,
   each ended or refused; a nested job; the runner ended with `taskkill /f` ends the tree; the job's cost and Defender's
   first-scan figures measured again against 5.4.1's.
4. **The runner:** the requested set (floor plus the task's, each once); a required red skips the rest; any red makes it
   red; a run from a worktree writes the main checkout's `.bonsai/local/ladder/<task>.json`; every result and every
   `--json` validates; the `ladder` record's `sha256` equals the file's and names the runner's path; `--ci` writes
   `ci.json`, `mode: ci` and the proof sentence, files nothing and prints no leftovers line; a local run with no task
   exits 2; a stamped build refuses outside its scratch root; `hook start`'s next context names the task's last result.
5. **One ladder at a time:** two runs on one home: the second waits while the holder lives and is inside its budget,
   then exits 4 naming the holder; a killed holder's lock is taken at once; two homes do not block each other.
6. **Test outputs and skips:** fixtures of `go test -json` (subtests, a package that does not build), node's TAP
   (subtests, SKIP, TODO, a failure's file) and JUnit give their names and outcomes; locally a marked skip is red and an
   unmarked one passes; under `--ci` a skip passes only when named in `ci_marked_tests` with a mark.
7. **Rung 0:** every case of the `rung0` section; a protected change granted, not granted, and granted at `verify`; the
   floor's never-granted paths red under any grant; a table changed on a branch red and on the base not; a staged
   `local/` file red; a `check` finding red and a machine finding listed as not judged; no task under `--ci`: a
   protected change red; the guard and rung 0 agree on the verifier's own table of paths.
8. **Ratchets:** a count below its floor red; a ratchet with no count red and a `check` finding; a capture that does not
   compile a `check` finding.
9. **Bless:** in a fixture main checkout on its base branch, clean, at HEAD, green and above a floor: one ask
   `ladder:bless-<task>` with `source: ladder` and contract §9.1's `data`, and its `ask` log record; none on a branch,
   dirty, red or under `--ci`; a later run with no rise resolves it; `bonsai ask --type Bless` still refused.
10. **New tests:** for each form, a new test failing at the base counted; one passing there counted as proving nothing
    new; a Go test calling new code counted as not run; a failing `base_setup` makes every one not run; the temporary
    worktree gone and `git worktree list` as before; the numbers in the result and on the Bless ask.
11. **Git integrity:** each finding on its fixture repository (Windows git on Windows); a clean repository finds
    nothing; the rung green with findings; `dirty` true with an `assume-unchanged` entry and with a file ignored by
    `info/exclude` alone.
12. **No `golang.org/x/sys`:** `go list -deps ./cmd/bonsai` names none; the hook line's p50 and p95 on both sides
    against 5.3's.
13. **Unattended:** each refusal of `bonsai ladder` (no task, the lock held past its wait, an undefined rung, a stamped
    build outside its root, an unverified main) exits with its code and an `error` whose `next.do` the verifier runs as
    written, and whose `who` is `agent` but where the words' table names a safety reason; with no terminal nothing waits
    for input; a red climb's `--json` gives `next`, and each failing test's `run` runs that test alone on both sides;
    `status --json`'s `ladder` shows the active task's last result, and `stale` after a new commit.
14. **An agent alone:** a scripted `-p` session through `claude-here`, run by a Sonnet agent, in a scratch project
    linked by the build, holding a running task whose test fails and a ratchet at the fixture's count, told only "make
    the ladder green for <task>": it climbs, reads the result, fixes the test and climbs green, with no person and no
    prompt (deleting the test stays red: the count falls below its floor). The run report keeps the session's commands.
    On WSL, and on Windows if Claude Code's login there is back.
15. **The break-it:** the verifier's own process trees on both sides and its own repositories for rung 0, ratchets and
    git integrity, beyond the builders' tables: no rung leaves a process for the next, and no change outside the named
    task's grants climbs green.
16. **The pre-release:** the installed file's SHA-256 equals the number Rohan was sent and the verifier's own
    `-trimpath` build of the commit from its own clone; `go version -m` shows that commit, unmodified; `which -a bonsai`
    lists `/usr/local/bin/bonsai` alone; `bonsai --version` names the commit.
17. **The link:** Bonsai's committed `bonsai.yaml` holds the rungs, lists and floors this section sets and validates;
    the lock, `.claude/settings.json` (every Bonsai line in Rohan's (a) form, naming the two places and no other path),
    the block and the tables are committed; `bonsai check` in the main checkout and in a fresh worktree: no finding;
    `.gitignore` holds `.claude/settings.local.json`; `CONTRIBUTING.md` says a clone without `bonsai` is refused and how
    to work round it; `CLAUDE.md` holds the landing rule, agents first, the stricter-only rule, the task and name rules,
    the (ii) commands and the way back; the link made by the orchestrator with the installed build.
18. **Guarded:** in a scripted session through `claude-here` in a fresh worktree of Bonsai, with only `T-5490` running
    (it grants nothing): a file-tool edit of `bonsai.yaml`, `go.mod` and `.github/workflows/ci.yml` refused, each naming
    why; a free file edited; `rm -rf junk/*` refused; a scratch project linked on a scratch home and deleted by name,
    allowed; the records in the main checkout's `.bonsai/local/log/`. In a scratch clone of Bonsai linked by the
    installed build on a scratch home, a running task granting `go.mod`: that edit allowed.
19. **Proven and gated:** `T-5407`'s green result in the main checkout, its record and `hook start`'s context agreeing;
    in a fresh worktree on a branch of its own, `T-5490` climbed by `/usr/local/bin/bonsai` green at its HEAD; a `-p`
    session there with `BONSAI_TASK=T-5490` is blocked at its stop while no result proves its HEAD, and stops once one
    does; a session with nothing named stops freely (the orchestrator's case); a subagent's end is not gated.
20. **Rohan's commands held:** in a scratch clone of Bonsai linked by the installed build on a scratch home (never the
    real checkout), an agent-shaped run (`CLAUDE_CODE_CHILD_SESSION` set) of `update --allow-exec --yes`, `unlink` and
    an update changing Bonsai's guard line is refused by the command; the same with the variable unset is written.
21. **The way back:** the lines in Rohan's part are exactly those 5.4.7 tried, and the verifier tries the first once
    more in a scratch project with every call refused.
22. **Check 10, the ladder and CI:** `go test ./...` and `go vet ./...`, plain and tagged, on WSL and natively on
    Windows, run by the verifier; `/usr/local/bin/bonsai ladder` green on the final commit; CI green on it.
23. **Stop lines and records:** 5.4's hours under 57, this section's planning and review included; step 5's Windows-only
    tally; option rounds (none); nothing written or run in the studio's checkout or in Mimas (the scripts and the run
    reports' commands read); the user settings hashes around every Claude Code run; the Haiku audit of
    `.bonsai/sessions.md` against the run reports after the switch; every change to `bonsai.yaml` since the switch
    stricter or naming Rohan's word. **Nothing private:** a grep of the diff, the commit messages, the task files and
    the memory note.

#### Risk in the code, 5.4

- **Windows process trees.** A process started outside the job, between the start and the assignment, would escape: the
  suspended start closes that window, and V1 tries it. A job that cannot be made runs the rung without one and says so,
  never silently.
- **Linux escapes.** A `setsid` child that clears its environment and whose parent is gone cannot be told from a
  stranger without root, as today; a runner killed by `SIGKILL` leaves what escaped its shell. The leftovers line names
  both.
- **A broken pre-release stops all work in Bonsai's checkout** (fail closed). The way back is written and tried before
  the switch, and V2 passes the code before Rohan installs it.
- **The guard crying wolf in Bonsai's own repo.** Rohan's own sessions are refused on the protected files too, and two
  tasks running at once leave no active task. The list is short, the rule for running side by side is the
  orchestrator's, and the end verifier lists every refusal its sessions met that should not have happened.
- **A non-task file in `records/tasks/`** makes the active task none for every reader: no grants anywhere (fail closed).
  The folder holds task files only.
- **Frozen run reports.** From the link, every old report is on the lock's `format0` list: any later edit is a `check`
  finding and turns rung 0 red. Reports open at the link are format 1 from their first line.
- **The stop gate and Bonsai's builders.** Subagents are never gated (the gate runs on Stop only), so the orchestrator's
  landing rule is their gate, as the interim proof's was; a named session is gated up to Claude Code's cap of eight.
- **Agents loosening their own checks.** Agents now link Bonsai's repo, raise its floors and edit `bonsai.yaml`. A
  change that loosens the guard or the ladder waits for Rohan's word (the stricter-only rule), and each part's end
  verifier reads every `bonsai.yaml` change against it. Bonsai's code never writes `ratchets`.
- **Scratch clones of Bonsai.** A clone carries Bonsai's committed lines, which run the installed pre-release, not the
  build under test, until a stamped build relinks it in the scratch root (a changed line: `--allow-exec`, which a
  stamped build may pass). A Windows clone made for check 10 carries them too: no Claude Code session ever opens there.
- **Time.** Every climb runs every test twice, and the new-tests check twice more at the base; the climbs' minutes count
  inside each builder's run. A builder climbs once its work is done; the measured climb goes in 5.4.7's run report, and
  a climb past 15 minutes comes to the orchestrator.
- **Defender** scans every new binary, test binaries included: Windows rungs run slower the first time; the measured
  margin sets the default timeout.
- **Claude Code moves:** hook snapshots at a session's start, `--settings` against project settings, `disableAllHooks`
  in a local file, `autoMemoryEnabled`, SubagentStop against Stop; each read on the version in use and recorded.
- **Processes:** rung-job tests start real process trees on both sides, and the scripted sessions run Claude Code; the
  orchestrator sweeps with `ps` after each agent and stops by pid only what that agent clearly left.

#### Stale or in tension in the spec, for 5.4

- **§3: "Standard library plus `golang.org/x/sys` (Windows job objects)":** 5.3 kept it out of the binary, and 5.4's job
  objects go through `syscall`'s lazily loaded procedures; `CLAUDE.md` and `CONTRIBUTING.md` change at the switch. The
  orchestrator adds the spec's dated note.
- **§9: "each in its own process group on Linux":** a group alone lets a `setsid` child escape; Bonsai adds a session, a
  tag and the runner as reaper, as today's ladder goes beyond the group (5.4.1 note 1), without today's keeper.
- **§9 and contract §11: Bless "on a clean base branch at HEAD"** with "the floor rises only when the studio applies
  Rohan's tap": that holds where the studio manages a project. Elsewhere, Bonsai's own repo among them, an agent raises
  a floor to a green count itself (Rohan, 9 Oct: agents manage Bonsai inside projects), since a raised floor only makes
  the ladder stricter; lowering one stays a person's (5.4.4 note 4).
- **Spec §18: an agent's `init`, `update` and `unlink` "need a task he approved that names those files":** his 9 Oct
  direction gives linking, updating, fixing and editing to agents, and his answer (ii) keeps `--allow-exec`, `unlink`
  and changes to Bonsai's own lines a person's where the studio does not manage the project; Bonsai's own link is the
  orchestrator's (the switch).
- **Contract §11: "When a count ratchet rises, the runner runs the new tests on the base commit":** Bonsai runs them
  only on green, clean runs (5.4.5 note 1), so a climb while work goes on has `new_tests` null though a count is above
  its floor. The schema's description says so (an addition).
- **Contract §11's `new_tests` `{base, count, failed_on_base}`** cannot tell "proves nothing new" from "check not run",
  which spec §9's card shows: two counts added at its end (5.4.0 note 3).
- **Spec §9: "the merge base"** of a run on the base branch is HEAD itself: such a run takes the task's recorded base
  (5.4.5 note 2). Neither the spec nor the contract defines the base branch for a Bonsai project (today's `game.yaml`
  has `base_branch`): it is the branch the main checkout holds (5.4.3 note 1).
- **`ver-git` "not required at first: it informs"** against `bonsai.ladder/1`'s "any red rung makes the run red,
  required or not": it never goes red for a finding, and what hides a change makes `dirty` true instead (5.4.6 note 3).
- **Spec §4: "CI and rung 0 run [`check`]":** rung 0 counts the findings about the commit, not those about this
  machine's install (5.4.3 note 4).
- **Spec §6's example and today's `ci_marked_tests`** hold `{name, mark, file}`; Bonsai's schema holds names, so a CI
  skip is accepted for a listed name whose reason carries any mark (5.4.2 note 4).
- **Today's rung fields `exit`, `output_tail`, `job` and `checked`** are not contract §11's: Bonsai writes the failing
  tail in `reason` and rung 0's count in `captures`.
- **Today's `--ci` runs only on GitHub Actions;** Bonsai's runs anywhere, and `mode` keeps it from proving a task (5.4.2
  note 6).
- **Spec §4: `bonsai ladder --task T`:** the task may also come from `BONSAI_TASK` (contract §13; the `rung0` fixtures'
  `named-env` case), and `--ci` may name none.
- **`workspace.schema.json`: a rung's `ratchet` is "not in Bonsai's sources":** today's ladder, read at `b5f7cbf`, has
  it as a name with a capture of that name; 5.4.0 fixes it by description, with no new major.
- **Contract §5.6 and this plan's 5.1.5: `bonsai.*` definitions come with base:** until 5.5, Bonsai's readers act on
  §5.6's meaning with no definition in force (the switch, "Labels").
- **Contract §7.1: "A run report lives on the builder's branch":** Bonsai's are the orchestrator's, in the main
  checkout, format 1 from the switch.
- **Contract §7.2: STATE at `.bonsai/STATE.md`:** Bonsai's stays at the root until base's template (5.5).
- **Spec §14, "How Bonsai's work is proven":** "Bonsai's guard then covers its own checkout too": the guard and the
  delete check cover every session there; the stop gate covers only named sessions, so the orchestrator's builders are
  held by its landing rule (the switch, "Tasks and names").
- **Spec §17 step 8:** the install stays Rohan's; the link after it is the orchestrator's. Under (a) the hook lines
  never read the PATH, so `which -a` guards what an agent runs by name (the ladder is called by its full path anyway).

### Steps 5.2-5.7, outlined

Each gets its detailed section, in 5.1's shape, before it starts ("What changes", item 1). The spec rows are §14's.

**5.2 Recorder, logs, asks (25-37 h, re-ask at 48).**
- **Planned in full** in "Step 5.2" above; this outline is kept as it was written.
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
  memory notes, on the redactor's patterns. The log's events, an open list read by code at many places, keep their
  known words in one Go table, as 5.1.3 does for the `error` words; `asks` and `logs` get their `--json` schemas as
  additions.
- **Risks:** redaction is security (a fresh verifier). Its differential corpus is the studio's redaction tests at a
  named commit, read-only; it runs as a scripted differential like format 0's, or is copied in only after a scrub
  (made-up secrets only, no studio ids). Appends under concurrency on Windows (whole lines, busy retries). The
  generated-files page is a `base` skill in the spec (§6) and `base` arrives in 5.5: the section says where the page
  lives until then.
- **Rohan:** nothing. The studio's forwarder can start from here (its own work).

**5.3 Guards (18-29 h, re-ask at 38).**
- **Planned in full** in "Step 5.3" above; this outline is kept as it was written.
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
  a session's first call over 5 ms on WSL (5.73 ms at p50, the self-hash), and the guard's WSL p95 about 0.6 ms
  above part 5's (3.48 and 3.46 ms against 2.90, measured under load); the large-Write fail-closed test on Windows;
  `golang.org/x/sys`, which part 5 tried and dropped for about 3 ms more on every Windows start, while spec §3 plans it
  for 5.4's job objects: 5.3 decides how it stays off the hook's path or fits its budget; and the coordinator's four
  remaining timed tests (`TestOverTimeBlocks`, whose record half rests on the guard's `recordWait` constant;
  `TestFaultsThroughTheHookLine/slow` at 9 s against the guard's 5 s; "busy forever" at 2 s; the guard tests on the
  default 5 s budget), reviewed with the guard's budget. 5.3 also tests the guard's and the stop gate's sections of
  set 4's active-task fixtures.
- **Risks:** Windows; a guard that blocks real work ("cries wolf") gets worked around; a hook that times out does not
  block, so the guard's own timer stays. The whole part is guards and hooks: a fresh verifier, with a break-it run on
  both sides.
- **Rohan:** the second Windows check, spec §17 step 7, about 10 minutes, after a Sonnet agent's scripted run on both
  sides. Possibly the hook line's form as an option round.

**5.4 Ladder runner (28-44 h, re-ask at 57).**
- **Planned in full** in "Step 5.4" above; this outline is kept as it was written.
- **Builds:** rungs, process groups and job objects, one ladder at a time, leftovers, `mode`, the fingerprint, results
  in the main checkout's `.bonsai/local/ladder/`, rung 0's refusal of branch changes to the tables and of tracked
  `local/` files (13-19); floors, ratchets, Bless filing (4-6); new tests must fail, by name, with `base_setup`
  (7-12); git integrity (3-5); Bonsai's own `bonsai.yaml`, the switch from the interim proof and the pre-release build
  Rohan installs (1-2).
- **Needs:** 5.1 (the ladder in `bonsai.yaml`, the active task, the tasks), 5.2 (the `ladder` record, the Bless ask),
  5.3 (rung 0 judges as the guard does; the stop gate reads the results).
- **Settles (gate report §5, 5.4):** Windows job objects and Windows Defender's first scan of a new binary, never
  measured. Job objects use `golang.org/x/sys` as 5.3 decided it may be used. 5.4 also tests rung 0's section of set
  4's active-task fixtures.
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
- **Rohan:** the one line in his `~/.claude/CLAUDE.md` that imports his personal memory index (spec §10), sent with
  exact text; one real install proof only if 5.6's section asks for it. The real installs on both sides come with 1.0
  and the studio's link (spec §3, step 7).

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
4. **Any change to Mimas or to the studio's repo made by step 5's work**, carried over. The studio now plans and
   commits its own work in its repo at the same time, so a new commit there is not by itself a crossing, and the
   skeleton's test (the repo's last commit date) no longer works. What the verifier checks instead: no command in a
   step 5 run report or in `~/bonsai-checks/scripts/` writes or runs in either checkout; the studio's files are taken
   only with `git show <commit>:<path>` or from a scratch clone with its origin removed ("What changes", item 10); Mimas
   is read only with Windows git's read commands (`git archive`, `git log`, `git show`); and no Bonsai session or
   builder starts with either checkout as its folder. Read-only use is not a change.

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
| 5.1.0 | The test passes plain and tagged on both sides; the diff holds only `_test.go` files |
| 5.1.1 | The consent table's Go tests; every test and script that links the test pack passing with `--allow-exec`; check 5 scripted with E and F; the verifier's break-it, first links and relinks included; check 10; CI |
| 5.1.2 | Every `format0` outcome; fuzz; no difference over today's files, reported without paths; the verifier's own frozen-reader run; check 10; CI |
| 5.1.3 | The formats test; the reader's and the port's tests on the two new cases; the schema-compare test failing on a removal; the Windows-git clone's `go test ./formats/`; the verifier's rule-by-rule read; CI |
| 5.1.4a | Its Go tests; the guard's lean read held by a test; the hook's p50 and p95 on WSL before and after; check 10; CI; the orchestrator's read |
| 5.1.4b to 5.1.10 | Their Go tests; scripted runs where Claude Code is involved (5.1.7); check 10; CI; the orchestrator's read of the diff |
| 5.1 | The end verifier on "5.1 done" |
| 5.2 | Its section's "Proof for each piece"; fresh verifiers for 5.2.1 and 5.2.4; the end verifier on "5.2 done" |
| 5.3 | Its section's "Proof for each piece"; fresh verifiers at 5.3.1, at 5.3.2 with 5.3.3, and at 5.3.6; the end verifier on "5.3 done", after Rohan's sitting |
| 5.4 | Its section's "Proof for each piece"; fresh verifiers V1 (5.4.1 to 5.4.3) and V2 (the code, before the pre-release); the end verifier on "5.4 done", after Rohan's install and link |
| 5.5 to 5.7 | Each part's section; its verifiers as outlined; its end verifier |
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
| Bonsai's own lines written at a first link on `--yes`; a pack's hook lines need `--allow-exec` there too | `--allow-exec` for every link, or `--yes` for every first link (the skeleton) | Bonsai's lines are the link's purpose and the preview names each; a pack's hook line is someone else's code. Cost: every first link to the test pack, and every test and script making one, gains `--allow-exec` |
| Format-0 refusals matched as refusals, messages not compared | The same messages | A Go port names its own messages with next steps; readers must agree on outcomes |
| Today's files read as a scripted run | Committed tests | They are private (as `plan.md`'s scratch targets) |
| The schema compare against the set's base commit, read with git | A frozen copy of each schema | A copy would be a second home for every list |
| "Addition" defined narrowly | Allow widening a type or a pattern | An older reader must read every newer writer's document |
| `formats.write` as the major a writer writes | Make `write` nullable | Contract §12's own example; nullable is not an addition |
| `error.next` as `{do, who}` | One sentence | An agent decides by `who` without parsing words (format review 4.5's question) |
| `error.code` an open list, its words in one Go table | A closed `enum` | 5.2 to 5.7 add words; a closed list never grows inside a major, and the schema compare would fail each addition; format review 4.5: "Adding a word is an addition" |
| `check` and `changes` output schemas in set 4; `error` added to `status` | Shapes documented in `--help` only | With no screens of its own, the studio reads these outputs; a schema holds them as the formats are held |
| The guard keeps a lean read of `bonsai.yaml` | The full schema read on every hook call | An unrelated field's error must not block every edit, and the hook's start-up budget is already tight on a first call |
| 5.1.3 after 5.1.2 lands, owning the reader's fix | Both at once | They share `internal/reader` and `expect.json`, and the new cases need the port |
| 5.1.4 in two pieces (types; outputs) | One piece | Eighteen formats' types and writers, `--schema`, `error` in every command and the help table are too much for one brief |
| `update` takes a pack out, with its plugin | Keep refusing | `plan.go` and `check.go` defer it to 5.1; `unlink` needs the same removal anyway |
| §13's fixtures in `formats/` | Go tests only | Contract §13: "run against every reader"; the studio tests against them at step 7 |
| The tables' schemas describe the rows a reader returns | The markdown text | A JSON Schema cannot describe a markdown table |
| `unlink` keeps `.bonsai/.gitignore` while `local/` stays | Remove it | Else the log and asks show to git |
| `unlink` removes Claude Code's install record | Leave it | The gate's finding (verifier N6) |
| Bonsai keeps the settings file's key order | Its own order | No churn after Claude Code's rewrite |
| Trust stays a person's; `waiting` with a person's next step | Register the marketplace headless | Trust is Claude Code's safety question for a person |
| The memory secret scan in 5.2; the stranded folder in 5.6 | All of §6 in 5.1 | One home for the secret patterns; 5.6's row names the stranded folder |
| 5.1.0 before 5.1.1 (or beside it), with `TestEachFaultBlocks` only | Leave every timed test to 5.3 | Every 5.1 push runs it on a loaded Windows runner; a test of its shape flaked already. `TestOverTimeBlocks` needs guard code, so it waits for 5.3 |
| Two pieces beside another in 5.1 (5.1.0, 5.1.2) | None, or more | Rohan's 8 Oct rule: only where truly independent (no shared file, no shared proof) |
| A run report per piece | One per part | 5.1 is twelve pieces over days; each report stays one log |
| A verifier for risky pieces and at each part's end | One per piece | `CLAUDE.md`, Rohan's rule; spec §14's "on every step" met per part |
| The skeleton's Windows, option-round and Mimas lines carried over | Hours lines only | Each still guards what it was for; reasons in "Stop lines" |
| Planning and review runs count in their part | Outside the parts | The spec's estimates cover the whole part; conservative |
| The Mimas and studio line judged by step 5's own commands | By either repo's last commit | The studio now works in its repo in parallel |
| Later parts' sections reach Rohan under (B), with "format change" defined | (A) every section; (C) none | Rohan, 9 Oct |

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
- **F's plugin hook runs in every later scripted session** at F or G, on both sides: an `echo` only, so harmless,
  but it shows in those sessions' output and logs; scripts that read session output allow for it.
- **The guard's path.** 5.1.4a changes how `bonsai.yaml` is read; the guard keeps its lean read, held by a test, and
  the hook's timings are compared before and after.
- **Every first link now needs `--allow-exec`** where a pack carries a hook line: a test or script missed in 5.1.1
  fails loudly (exit 4), not silently.
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
- **§3's `golang.org/x/sys`** for job objects against part 5's measured 3 ms on every Windows start: the gate lists it
  under 5.3, which decides; 5.4's job objects follow that decision.
- **§4 lists `init [--new-id] [--json]`**, with no `--yes` and no `--allow-exec`, while §4 also says `init` previews
  every settings line "as `update` does", the skeleton's `init` takes `--yes`, and §6's rule for code applies to a
  first link's pack hook lines. This plan gives `init` both flags (5.1.1, rule 6).
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
- **Contract §2.2 says "a new optional field"**, while every schema requires every field. The reading (5.1.3): a schema
  describes what a writer writes, so a new field is required there; a reader still reads it as `null` when an older
  writer left it out, so it is optional to readers.
- **Format review 4.5** left two "worth a look" questions open under Rohan's "Agree": `who` as its own field (taken:
  `next.who`), and whether he wants the code words before they are built. The words are an open list (5.1.3), kept in
  one Go table and printed by `check --schema bonsai.error` (and by the reference page once 5.1.10 lands); the
  orchestrator sends him the list as a look when 5.1.4b lands, no option round, and a word he wants changed is changed
  before 1.0.
- **Contract §12's `--full`** lists "MCP servers reachable"; the plan keeps that wording, for the MCP servers the packs'
  `needs` name, shown as unknown where Bonsai cannot tell without a session (5.1.6).
- **Contract §13's working name** `bonsai task active --json` against spec §4's `status --active`: the spec fixes
  command words, so `--active`.
- **The generated-files page** is a `base` skill (§6), but cleaning is 5.2 and `base` is 5.5: 5.2's section says where
  it lives meanwhile.
- **Base's `bonsai.*` label definitions** arrive with 5.5, but Bonsai links itself at 5.4: 5.4's section settles what
  its tasks' labels are checked against meanwhile.
- **Step 6, "it registers as its own studio project"**, needs the studio's registration (contract §15.2, "after the
  skeleton"), which is studio work while its Desk is on upkeep until step 7, and which runs `bonsai labels attach` and
  `bonsai settings`, both 5.6's. Bonsai links itself at 5.4; full registration waits for 5.6; when it shows on the Desk
  is the studio's plan.
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
