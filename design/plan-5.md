# Bonsai's plan for step 5: path (a), Bonsai 1.0, with 5.1 in full

- **Status:** draft, 9 Oct 2026. Reviewed by a fresh Opus agent (ready after fixes) and fixed on its findings and on
  Rohan's two answers of 9 Oct (the plans of later parts, and Bonsai's scope); for Rohan's approval. Nothing in step 5
  is built before he approves it.
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
- **Now:** approve this plan. Nothing in step 5 is built before.
- **In 5.1, a look (no vote needed):** the list of error words every command's `--json` uses, and its "what next"
  part, which becomes two fields (what to do, and whether an agent or a person does it) instead of the one sentence you
  saw in format review 4.5; with it, the shapes of `check`'s and `update`'s JSON, which the studio will read. Until
  1.0 any word can still change.
- **In 5.3:** the second Windows check (spec §17 step 7, about 10 minutes): a real Windows session in a scratch folder,
  asking Claude for one edit and two deletes and reporting what happened. A Sonnet agent runs every check an agent can
  first, so your sitting is only what needs a person typing (your 8 Oct word). Possibly one question: how the hook
  lines find `bonsai`, where your 8 Oct rule (a fixed path no agent can redirect) meets the rule that no committed file
  holds a full path.
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

**Choices that are yours.** None left in this plan beyond approving it. One was decided on 9 Oct:
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
