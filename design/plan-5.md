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
  5.6 13-20, 5.7 9.5-18.5 (the spec's 6-11 and the trial of Bonsai on itself Rohan asked for on 10 Oct); in all
  142.5-225.5. Each part stops and comes back to Rohan at its own re-ask line (below).
- **Records:** `STATE.md`; run reports in `records/runs/`.

Two readers. **Rohan** reads down to "Size" and reads no code. The **orchestrator, builders and verifiers** read the
rest. 5.1 to 5.7 are planned in full here.

## For Rohan (plain words)

**What step 5 builds.** Bonsai 1.0, in seven parts, one after the other. What each lets you do or see:

| Part | What you will be able to do or see when it is done | Spec hours | Re-ask at |
|---|---|---|---|
| 5.1 Formats and engine to 1.0 | Bonsai reads and writes every format, old files included. An update that would add or change code that runs (a hook) waits for a second, separate yes. A pack can be taken out of a project, and `bonsai unlink` takes Bonsai out cleanly. `bonsai check` reports the problems the spec lists (two more come with 5.2 and 5.6), `bonsai status` is complete, and one page lists every list of allowed values. | 30-47 | 61 |
| 5.2 Recorder, logs, asks | Every session in a linked project leaves a log inside the project, with secrets hidden as it is written. Agents can ask you typed questions and read your answers. Old records are cleaned by rules you set. A table of sessions and hours per task. | 25-37 | 48 |
| 5.3 Guards | The full guard: an agent may change a protected file only while its running task allows it; files only you may grant stay yours; a recursive delete that does not name what it deletes is refused; a builder cannot stop before its proof is green. You run one real Windows session (about 10 minutes). | 18-29 | 38 |
| 5.4 Ladder runner | `bonsai ladder` proves a task's work, on WSL and Windows. From here Bonsai proves and guards its own repo with it, on a pre-release you install in WSL (about 5 minutes, your password). | 28-44 | 57 |
| 5.5 Packs | Bonsai's `base` pack and your `workflow` pack (your roles, lanes, protocols and templates) as public Claude Code plugins, each with its own checks on GitHub; a template for new packs; the walls round your key and token files. | 19-30 | 39 |
| 5.6 Machine pieces | Bonsai's settings for each project on this computer, and the label files the studio attaches (set by you or the studio, never by an agent); your personal memory, which every Claude session on this computer reads; Bonsai's part of the status line; installers for both sides; and Bonsai telling you when a newer release exists, with the exact lines to install it. | 13-20 | 26 |
| 5.7 Release | The release path made safe and switched back on with you; a pre-release you install, and a trial of Bonsai on itself on it, analysed for you; your own trial on a fresh project; on your word, Bonsai 1.0. | 9.5-18.5 (the spec's 6-11, plus the trial) | 24 |
| **Step 5** | | **142.5-225.5** | each part its own |

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
  A new repo for your `workflow` pack in 5.5 (`LastStep/bonsai-workflow`, the spec's name), once you approve 5.5's
  section: created private, and made public only by your own line after you have read it.
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
- **Before 5.3 (overdue since 5.1):** renew Claude Code's login on Windows. In PowerShell, one line at a time:
  `claude`, then type `/login` and sign in, then `/exit` (about 2 minutes). 5.3's Windows runs and your 5.3 sitting
  need it, and 5.5's Windows sessions after. If it lapses again, the orchestrator sends these lines again.
- **In 5.3:** the second Windows check (spec §17 step 7, about 10 minutes): a real Windows session in a scratch folder,
  asking Claude for one edit and two deletes and reporting what happened. A Sonnet agent runs every check an agent can
  first, so your sitting is only what needs a person typing (your 8 Oct word). Two questions, each an option round in
  5.3's section, answered on 9 Oct: how the hook lines find `bonsai` (your answer: (a), the installed place written into
  the lines); and who may consent to code (`--allow-exec`) and take Bonsai's guard out of a project (`unlink`) (your
  answer: (ii), your grant in a project the studio manages, only a person elsewhere). From 5.4, so, a command of yours
  when an update of Bonsai's own repo runs code or changes its guard lines, until the studio manages it.
- **At 5.4:** install the pre-release `bonsai` in WSL (spec §17 step 8, about 5 minutes, your password): four lines,
  `sudo install`, `sha256sum -c` on the installed file against the number the orchestrator commits first, `which -a`
  and `--version` (5.4's section, "For Rohan"). Your step 3 (the old binaries) is already done (8 Oct). You install
  again only if a later part changes the guard, the stop gate or the ladder, and then, under your 5.3 answer (ii), one
  more line if it changes Bonsai's own hook lines: `bonsai update --allow-exec --yes` in `~/Servers/Bonsai` (none is
  planned before 1.0). From the switch: your word before an agent loosens a check in Bonsai's `bonsai.yaml`; your own
  Claude sessions there refused on its protected files; the floor numbers in a line at each part's end (nothing to
  answer); the way back only if the guard breaks.
- **In 5.5:** approve its section (you did, on 9 Oct). It makes your roles, lanes, protocols and templates public in a
  new repo, `LastStep/bonsai-workflow`, and puts three of your templates (task, run report, STATE), reworded and naming
  no one, into Bonsai's own public `base` pack; your approval is your word for those three, which go public when their
  pieces land, each after a fresh agent's check for anything private.
  - The new repo: a fresh agent checks every file and every commit for anything private first. It is created private,
    and you read it on GitHub (your phone works) with a file-by-file list of what changed from your studio's copies,
    asking for any change. When you are happy, and the orchestrator has said every change since the check was read,
    one line in WSL makes it public: `gh repo edit LastStep/bonsai-workflow --visibility public`. Then (optional,
    recommended) one more line stops anyone rewriting or deleting its `main`. Its commits carry the same author line
    (your name and email address) that Bonsai's public history already shows: nothing new is shown.
  - Once, about 3 minutes on Windows: start Claude Code through the scratch launcher in a scratch folder the
    orchestrator names, say yes to the trust question, and tell the orchestrator what it said.
  - Your choice when you approved it: moving a project to another version of a pack stays your step, and so do adding a
    pack to a linked project and taking one out. An agent prepares the change and hands you the exact line; it makes
    the change only on your word.
  - No repository secret, no install and no password in 5.5. If the studio links the `workflow` pack before 1.0, it
    first needs a newer pre-release `bonsai` installed in WSL (5.4's four lines again, about 5 minutes, your password),
    since 5.4's cannot read that pack.
  - From 5.5's end, Claude in `~/Servers/Bonsai` (your own sessions included) cannot read your key, token and login
    files, nor change your own Claude settings, `~/.claude/CLAUDE.md`, your shell's start files or git's and SSH's
    settings: you change those yourself. Bonsai's STATE moves to `.bonsai/STATE.md`.
- **In 5.6:** its section approved (10 Oct), with your answer on who saves notes about you that every project reads:
  you, from a draft an agent hands you. No install, no password and no Windows sitting in
  5.6. From 5.6 three kinds of step are yours, each handed to you by an agent as exact lines it never runs itself:
  saving a note about you into your personal memory (by hand, or by asking Claude in a session opened outside any
  linked project, such as your home folder); this computer's Bonsai settings for a project and the studio's label files
  (yours, or the studio's registration's; typed in a terminal of your own, not inside Claude); and installing a newer
  Bonsai release when `bonsai status` or `bonsai check` says one exists (six lines a side, run one at a time; if the
  fingerprint line does not print `OK`, or `True` in PowerShell, stop and send the orchestrator what it printed).
- **At 5.7** (its section, "Step 5.7", comes back to you for approval with its new hours: you answered its decisions on
  10 Oct and asked for a trial). Five batches, about an hour and a half in all, plus about 10 minutes for each further
  pre-release and your own trial, untimed. **Batch 1, the settings** (about 10 minutes, in WSL, each setting read back
  by a line after it): your `gh` checked (you updated it on 10 Oct), a look on github.com for the old Homebrew token,
  the `release` environment with your approval on each release, GitHub's lock on published releases, the packs' tags
  locked, and switching the release workflow back on. **Batch 2** (about 3 minutes): your tag for a pre-release,
  `1.0.0-rc.1`, and your approval on GitHub. **Batch 3** (about 15 minutes): installing rc.1 on both sides with the new
  installers (seven lines a side, one more than 5.6's: it checks GitHub's signed record of how the file was built; in
  WSL your password, on Windows one "allow this app to make changes" prompt), so Bonsai's own repo runs it; then your
  pick of one of two or three small features for the trial. **Batch 4** (about 30 minutes): the trial's analysis to
  read, and "happy" or what to fix; each fix round a new pre-release (your tag, approval and WSL install again). **Your
  own trial** on a fresh project, on your own. **Batch 5** (about 25 minutes): your word for 1.0, its tag and approval,
  installing 1.0 on both sides, one line in WSL that makes your own `~/.claude/CLAUDE.md` load your personal memory in
  every project (spec §10; no agent edits that file; an agent then checks that it loads), and the first release tags
  of the two packs: `base-v1.0.0` on Bonsai's repo and `v1.0.0` on the workflow repo, whose checks then publish its
  GitHub release. Until then projects name each pack by its exact commit.
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

**Choices that are yours.** 5.7's decisions, answered on 10 Oct: Homebrew stays at the old product, with a Mac and
each platform's own package manager named for after step 5; your approval holds each release; the packs' tags are
locked; and, before 1.0, a trial of Bonsai on itself on a pre-release you install, then your own trial on a fresh
project. Go 1.27 for the release ("we can go with go 1.27 latest") and a newer `gh` in WSL, which you installed, were
settled earlier that day. The section comes back to you for approval with its new hours. 5.6's section was approved on
10 Oct,
with (A) for who saves notes about you that every project reads: you save them, from an agent's draft (5.6's "Who saves
notes about you"). The five other decisions you confirmed that day: this machine's settings and the studio's label files
are all yours, even `cache_keep_days`; a newer release is shown as planned, `status --full` reading Bonsai's tags at
most daily and keeping the answer in the home's `cache/release.json`, which `status`, `check` and `--line` show offline;
the installers first run on your machines at 5.7 with 1.0; `status --line` is kept; the personal index is 40 lines and
4 KB, a note 4 KB. Three were decided on 9 Oct: the two in 5.3's section, how the hook lines find `bonsai` (your answer:
(a)) and who may consent to code and take the guard out (your answer: (ii)), and this one:
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
- **Your roles and protocols go public in 5.5.** A fresh agent checks every file and every commit for anything private
  before the first push. The repo starts private, and only your own line makes it public, after your read. Three of
  your templates go public earlier, reworded inside Bonsai's `base` pack, each after its own check.
- **Agents act on GitHub as your account.** They never tag, release, switch a workflow on, or change a setting, secret
  or ruleset; those stay yours (5.7).
- **Stop lines.** Work stops and you get the numbers and three choices (continue, the smaller cut, or pause) if: a
  part reaches its re-ask line; step 5 loses 8 hours to Windows-only failures; more than two option rounds are asked of
  you inside one part; or step 5's own work changes anything in Mimas or the studio's repo (the studio now works in its
  own repo at the same time, so only step 5's own commands count). Your choice is written down before any more work.

**Size.** The spec's hours: 139-218 AI hours for step 5, 142.5-225.5 with 5.7's trial (228-361 on the studio's record of
estimates growing 1.6 times), each part with its own re-ask line (the table above). The skeleton's measured pace:
6 h 17 min of agent runs against its 30-47 h, a ratio of 0.1337-0.2094 (gate report section 2.2). At that ratio step 5
would take 19.1-47.2 h. **That is not a promise**, for four reasons (gate report section 1): it counts only the agents'
own runs, not the orchestrating session, your sittings or waiting; it comes from six parts over two days; it swung by
part from 0.06-0.09 (the engine) to 0.32-0.53 (the hook path), and the parts that dealt with Windows and Claude Code ran
highest, which is more of step 5; and the stop lines use the spec's hours, not the ratio. Your own time: about 2 minutes
for the Windows login before 5.3, about 10 at 5.3 and 5 at 5.4, your read of the workflow repo and about 3 minutes on
Windows at 5.5, about an hour and a half over five batches at 5.7 (the GitHub settings, a pre-release and its installs,
the trial's pick and its analysis, your word, the 1.0 installs and the pack tags) plus about 10 minutes for each further
pre-release and your own trial on a fresh project, untimed, and the plan approvals your choice above sets. Nothing here
waits on the studio.

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
    findings is listed for 5.7 (gate report §5). Go 1.25's fixes ended on 19 Aug, and on Rohan's word (10 Oct, "we can
    go with go 1.27 latest") the toolchain moves to Go 1.27's newest patch as its own piece, landed before 5.7 starts;
    5.7.0 checks it still holds. If `govulncheck` turns red on CI before then, the piece that meets it lands the
    one-line bump as its own commit first, and its run report says so.

**Outside step 5, named so they are not lost.**
- **The sandbox probe** (spec §7, §16 row 25; contract §9.5): it "waits for its probe after the gate", with its
  fallback for `.bonsai/local/` (about 1-2 h with the probe, "outside step 5's totals like the probe itself"). It gets
  its own small plan after step 5, and ends in a root step of Rohan's (installing `socat`).
- **Part 0's run report quotes a studio task id** (`STATE.md`, loose ends; the plan's quoted line in
  `records/runs/R-2026-10-08-formats.md`). Run reports are a log, and no piece edits them; the orchestrator may take the
  id out in a records commit of its own.
- **Bonsai on a Mac** (Rohan, 10 Oct, with 5.7's answer on Homebrew): not in 1.0, because every project's guard runs
  Bonsai only from its two fixed places (5.3's (a): `/usr/local/bin/bonsai` and `C:\Program Files\Bonsai\bonsai.exe`),
  neither of which a Mac has as a place only an administrator writes, and 1.0's macOS build is untested and has no
  installer; it needs its own place, installer and tests.
- **Bonsai from each platform's own package manager** (Rohan, 10 Oct: "adding bonsai on official distributers for each
  platform", such as an apt package or winget): not in 1.0, because each must install into the two places the guard
  trusts (5.3's (a)), which Homebrew, for one, cannot, and each is a new channel to sign, publish and keep, beyond the
  installers 1.0 ships.

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
- **Agents first** (Rohan, 9 Oct, 15:35: Bonsai is managed by agents inside projects): every finding's and warning's
  `next.do` is the exact command that fixes it, runnable as written, wherever one exists (`bonsai update --yes`, `bonsai
  check --write`, a `git rm --cached` line), and prose only where a person must judge; a test walks the findings table
  and fails on a `next.do` that names a command Bonsai does not have or a flag its word does not take.
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
builders to regenerate it with any list change. **And `bonsai --help --json`** (Rohan, 9 Oct, 15:35: Bonsai is managed
by agents inside projects): one machine-readable document of every command word, its flags, exit codes and example,
and every error word with its meaning and usual `who`, read from the same tables (5.1.4b's word registry,
`format.ErrorWords`), so an agent learns the tool without reading a page; its shape documented in `formats/` as an
addition with its manifest in the same piece, and held by a test that it lists every word and flag the registry has.

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
breaks. His word is also needed before an agent loosens a check in `bonsai.yaml`. Nothing else of his changes: the hours
(28-44) and the re-ask line (57) are the spec's; the order of the parts stands; no repo is new and nothing from the
studio goes public; no option round is asked. Every format change is an addition: one new command output
(`bonsai.climb/1`), the active task's last result at the end of `status --json`, two counts at the end of a ladder
result's `new_tests`, new words in open lists, and descriptions. The shapes of a rung's `ratchet` and `capture` are
fixed without giving them a schema type, so `bonsai.workspace/1` needs no new major (note 5.4.0).

#### For Rohan, in plain words

**What 5.4 gives you.** `bonsai ladder`: one command that proves a piece of work is done, as the studio's ladder does
today, in WSL and natively on Windows. It runs a project's checks in order, cheapest first. Everything a check starts
ends with it, so nothing it started is left running (on Windows through the system's "job objects"). It saves the result
in the project's `.bonsai/local/ladder/` with a fingerprint in the log, which is what 5.3's stop gate reads before it
lets a named builder stop. It also keeps test counts from falling, offers each rise for "Bless" (with a check that each
new test fails on the code from before the change), and reports git tricks that hide changes. Then Bonsai starts using
all of it on its own repo.

**Your sitting, once (about 5 minutes; your password).** When the code is done and checked, the orchestrator builds the
pre-release, records its fingerprint (a 64-character number) in a committed run report, and sends you these lines with
the number already in the second. Open a second WSL terminal, so the orchestrator's session stays open, and type them
one at a time:

```bash
sudo install -o root -g root -m 0755 ~/bonsai-checks/prerelease/bonsai /usr/local/bin/bonsai
echo "<the number>  /usr/local/bin/bonsai" | sha256sum -c
which -a bonsai
/usr/local/bin/bonsai --version
```

- The second line must print `OK`. If it prints anything else, stop and tell the orchestrator: the file you installed is
  not the one it built. The check reads the installed copy itself, which only root can change, not the file in the
  scratch folder, which an agent could change between the build and your install.
- `which -a bonsai` must list `/usr/local/bin/bonsai` and nothing else. If not, stop and tell it.
- The last line names the commit it was built from. Send the orchestrator that line.

Why this is yours: the copy at `/usr/local/bin/bonsai` is the one Bonsai's guard lines run in every session, so only
root may place it (your password); an agent that could replace it could replace the guard that holds it. Everything
after the install is the agents': the orchestrator links Bonsai's own repo with the copy you installed (`bonsai init`,
which writes the guard and its other lines and needs no `--allow-exec`, since Bonsai's repo uses no pack yet), commits
it, and checks the guard is live. Claude Code most likely takes up the new lines in the running session (an agent
measures it first); if not, the orchestrator asks you to restart it without losing its conversation: `/exit`, then
`claude --continue` in `~/Servers/Bonsai`. Windows needs nothing until 1.0.

**What changes after the switch.**
- Every piece of work is a task file in `records/tasks/`, numbered after its piece (`T-5501` is piece 5.5.1), saying
  what it may change and what proves it. The orchestrator writes and moves them.
- A piece is proved by `bonsai ladder`, run by the copy you installed: Bonsai's guard check on everything the piece
  changed, `go vet`, the formats check, every test twice (plain, and with the test fault switch), and a git check. The
  orchestrator lands a piece only when its result is green at the exact commit it lands, and its run report keeps each
  landing's proof so a verifier can check it later. GitHub's checks and the native Windows test run stay as they are
  (the ladder runs in WSL; Windows gets its installed copy at 1.0).
- Bonsai's guard covers every Claude Code session in `~/Servers/Bonsai`, yours included. It refuses changes to
  `bonsai.yaml`, its lock, Claude Code's settings files, the GitHub workflows, `CLAUDE.md`, `.gitignore`, `go.mod`, the
  lint and release settings and the four approved design documents (the spec, the contract, the one-pager and the format
  review), unless the running task allows that file. It refuses a recursive delete that does not name what it deletes.
  It logs each session in `.bonsai/local/` (never committed). If you ask Claude yourself to change one of those files,
  it is refused too: change it by hand, or ask the orchestrator, which opens a task for it.
- Bonsai keeps its machine data in `~/.bonsai/` from the switch (a secret salt, this machine's record of the repo, the
  ladder's lock); none of it enters git.
- Claude Code's own automatic memory is switched off in Bonsai's repo (Bonsai's memory replaces it, spec §10). The one
  note it holds there (agents run the hand checks; you are asked only for what needs a person) is already a rule in this
  plan, and becomes a line in `CLAUDE.md`.
- Test counts may only rise. At each part's end the orchestrator raises the floors to the counts the part reached and
  tells you the numbers; nothing to answer. Raising a floor only makes the checks stricter. Any other change to the
  checks in `bonsai.yaml` (a floor lowered, say because a refactor merges tests; a rung taken out or its command
  changed; a protected file freed) waits for your word: an agent loosening the checks it is held to would be judging its
  own work.
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
- **Bonsai off, nothing written.** Quit Claude Code, then start it like this instead of plain `claude`, for as many
  sessions and days as the fix takes:
  ```bash
  cd ~/Servers/Bonsai
  claude --continue --settings '{"disableAllHooks": true}'
  ```
  `--continue` keeps the orchestrator's conversation. That session runs with no guard, no stop gate and no log, and also
  without your own status line, since the switch turns every hook off. The orchestrator in it fixes the problem, lands
  the fix on the checks used before the switch (the installed copy may be the broken part), and sends you a new build to
  install; then start sessions as usual again. An agent tries this line first, in a scratch project, before your
  sitting.
- **Back to the build before** (from the second pre-release on; the orchestrator keeps the one you had and sends you its
  number from the run report that recorded it):
  ```bash
  sudo install -o root -g root -m 0755 ~/bonsai-checks/prerelease/previous/bonsai /usr/local/bin/bonsai
  echo "<its number>  /usr/local/bin/bonsai" | sha256sum -c
  ```

**Hours, order and your other steps.** 28-44 hours, re-ask at 57, as the spec has them; 5.4 after 5.3 and before 5.5.
Your install was already on your list. New for you: the rare update line above; your word when a check in `bonsai.yaml`
must be loosened (a floor lowered, a rung changed); your own Claude Code sessions in `~/Servers/Bonsai` refused on the
protected files; perhaps one restart at the switch; Claude Code's automatic memory off in this repo; and the floor
numbers in the orchestrator's line at each part's end. The way back is only for a breakdown.

#### The pieces and their order

The spec's row (§14): "Rungs, process groups and job objects, one ladder at a time, leftovers, `mode`, the fingerprint,
results in the main checkout's `.bonsai/local/ladder/`, rung 0's refusal of branch changes to the tables and of tracked
`local/` files (13-19); floors, ratchets, Bless filing (4-6); new tests must fail, by name, with `base_setup` (7-12);
git integrity (3-5); Bonsai's own `bonsai.yaml`, the switch from the interim proof and the pre-release build Rohan
installs (question C, 1-2)". It also settles the gate report's 5.4 findings and what 5.1, 5.2 and 5.3 hand on.

**What exists** on `main` at `f47c085` (read each package's doc comment):
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
- 5.1.4b (`c2864c0`): the error words in `internal/format/error.go` (`ErrorWords`, each with who usually takes the next
  step), every refusal's `--json` carrying its word, and one table per command word in `cmd/bonsai/word.go`, from which
  `--help` comes.
- `cmd/bonsai` has no `ladder` word. Bonsai's repo holds no `bonsai.yaml`, `.bonsai/` or `.claude/`; `CLAUDE.md` and
  `CONTRIBUTING.md` describe the interim proof; `go.mod` requires nothing (the standard library only).
- Today's studio ladder, the measure where the spec says "as today": its runner (`ladder.mjs`, contract §11), its rung
  jobs and leftovers line, its stop gate (`stop-gate.mjs`, contract §13), its Bless offer (`bless.ts`, contract §9.2)
  and the ladder in its project file, at the studio's commit the brief names, read only with `git show` ("What changes",
  item 10). There a rung's `ratchet` is a name and its `capture` maps that name to a pattern with one group
  (`test_count: '^# pass (\d+)'`); `ci_marked_tests` entries carry a name, a mark and a file.

**What 5.1, 5.2 and 5.3 will have added** (from this plan's notes; none of these is built yet. 5.4's start re-reads each
against what landed, and 5.4.0's run report records any difference that changes a note below):
- 5.1.5: the active-task function (with `--task` and `BONSAI_TASK`), labels in force, declared document kinds with their
  id patterns, the instruction block, `init` reading an existing `bonsai.yaml`. 5.1.6: `check`'s findings and warnings
  in one table. 5.1.8: `check --write`. 5.1.10: `docs/reference/lists.md`, generated.
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
| 5.4.1 | **Rung jobs** (`internal/rungjob/`, new): one rung's command run so that everything it starts ends with it: on Linux in its own session and process group, tagged, with the runner as the reaper of what escapes; on Windows in its own job object, the command started suspended and put in the job before it runs; its timeout; its output; and **the leftovers line** (`internal/leftovers/`, new) | Go tests on real process trees on both sides (detached and daemonised grandchildren, a held pipe, a timeout, a child asking to leave the job, a nested job, the runner killed), a decoy outside the rung left alone; scripted runs on both sides; the job's cost and Windows Defender's first scan measured; V1's break-it | 5-7 | Spec §3 (the standard library, Windows rules), §9; gate report §5 (5.4), §2.5; this plan's 5.3 decision on `golang.org/x/sys`; read-only, today's rung jobs and leftovers line at the studio's commit the brief names |
| 5.4.2 | **The runner, `bonsai ladder`**: the test outputs read (`internal/testout/`, new: `go test -json`, TAP, JUnit); the rungs in order with `required`, the requested set, skips and marks, the ratchet's count; the result in the verified main checkout's `.bonsai/local/ladder/`, `mode`, `--ci`; the fingerprint record; one ladder at a time per home; the `ladder` cleaning call; `--json`; every refusal's next step a command an agent runs as written, and each failing test's own command; `status --json`'s `ladder`; `check`'s rows for the ladder's shape | Go tests on fixture projects (each requested set, a required red, every skip rule, each test form, a result from a worktree landing in main, the lock held and released, a dead holder); every result and `--json` against its schema; 5.3's stop gate accepting a green local result at HEAD and refusing a `ci` one; scripted runs on both sides; V1 | 4-6 | Spec §3, §4 (`ladder`), §6 (the ladder, generated files), §9, §10 (the home's `locks/`); contract §3, §5.6, §8.1, §8.2, §11, §13; this plan's 5.2.2, 5.2.6b and 5.3.4 notes; read-only, today's `ladder.mjs` at the studio's commit the brief names |
| 5.4.3 | **Rung 0** (`internal/rung0/`, new): what the branch and the working tree changed, each path judged by 5.3.2's function with the named task's grants at any status; a branch that changes a generated table; a tracked or staged `local/` file; `check`'s findings about the commit; `bonsai.*` labels read by contract §5.6's meaning with no definition in force | The `rung0` section of set 4's fixtures, every case, on both sides; fixture repositories for each refusal; a table of paths on which the guard and rung 0 agree; V1 | 3-4 | Spec §6 (the tables, `local/`), §7, §9; contract §5.5, §5.6, §13; this plan's 5.1.8 and 5.3.1-5.3.2 notes; read-only, today's rung 0 in `ladder.mjs` at the studio's commit the brief names |
| 5.4.4 | **Floors, ratchets and Bless** (`internal/ratchet/`, new): a count below its floor is red; a count that cannot be read is red; a green, clean local run on the base branch at its HEAD with a count above its floor files one Bless ask; the ask resolved when the offer is gone | Go tests walking a table of every case (each condition false in turn); the ask and its `ask` log record against their schemas; `--ci` never blesses; check 10; CI; V2 | 4-6 | Spec §9 ("Ratchets and Bless"); contract §9.1, §9.2, §10.3, §11; this plan's 5.2.5 notes 2-6; read-only, today's `bless.ts` at the studio's commit the brief names |
| 5.4.5 | **New tests must fail on the base** (`internal/newtests/`, new): the new tests named from the rung's output; a temporary worktree at the base; `base_setup`; the base run as it is, then with the branch's changed test files; "failed", "proves nothing new" and "check not run" told apart | Fixture repositories for each form (`go test -json`, node's TAP, JUnit): a new test failing at the base, one passing there, one that cannot load there, a broken `base_setup`; the worktree removed after; the counts in the result and on the Bless ask; check 10; CI; V2 | 7-12 | Spec §9 ("New tests must fail on the base"); contract §9.1 (`data`), §11; format review 5.2 (spec §16 row 23) |
| 5.4.6 | **Git integrity** (`internal/vergit/`, new): the `ver-git` rung kind: the task's base commit recorded; changed test files, assume-unchanged and skip-worktree entries, edits to `.git/info/exclude`, new stash entries and a rewritten base found and listed; the runner's `dirty` counting what git hides | Fixture repositories, one per finding, on both sides (Windows git there); a clean repository finds nothing; the rung informs and never turns red for a finding; check 10; CI; V2 | 3-5 | Spec §9 ("Git integrity"), §16 row 23; contract §11 |
| 5.4.7 | **The switch**: Bonsai's own `bonsai.yaml`, its task folder, `CLAUDE.md`'s and `CONTRIBUTING.md`'s rules, `.gitignore`; the way back tried; the pre-release; Rohan's install; the link by the orchestrator, committed; the first climbs | A scratch clone of the switch linked and climbed by the final build; the way back in a scratch session; Rohan's lines and words; the end verifier's checks on the real repo | 1-2 | Spec §3, §6, §7, §14 ("How Bonsai's work is proven"), §17 step 8, §19 C; contract §4, §5.5, §5.6, §7.1, §7.4; this plan's "What changes", items 5 and 7, and 5.3's two answers |
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
| 5.4.7 | `bonsai.yaml`, `records/tasks/`, `records/memory/INDEX.md` (empty, only if `check` needs it), `CLAUDE.md`, `CONTRIBUTING.md`, `.gitignore`; then, from the link's `init`, `.bonsai/`, `.claude/settings.json` and the block in `CLAUDE.md` |

1. **5.4.0 and 5.4.1 start together**, once the orchestrator has read where 5.1.4a, 5.1.4b and 5.2.0 put the tables
   5.4.0 extends. 5.4.0 sets every shared list and schema first, so no later piece edits a shared list file; 5.4.1 is
   the longest and riskiest piece, so it starts at once. 5.4.1 is new packages only and needs no format; 5.4.0 runs no
   process. Neither's proof rests on the other's.
2. **5.4.3 and 5.4.6 after 5.4.0 has landed, beside the rest of 5.4.1.** Each reads 5.4.0's tables (rung statuses; git
   integrity's finding names; the test-file convention per `tests` form) and works in a new package, judging a fixture
   repository and returning a rung's outcome. Each takes the base branch and the merge base as inputs, which the runner
   computes once per run (5.4.2 note 13); its own tests compute them from their fixtures with git. Neither runs a rung
   job, and each is proved on its own fixtures, so neither's proof rests on 5.4.1's, on the runner's or on the other's.
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
   `CONTRIBUTING.md`, the task folder, `.gitignore`), and an agent tries the way back in a scratch session with any
   recent build. Its proof (a scratch clone of the switch climbed by the final build, which sets the timeouts) rests on
   every code piece, so it lands only after **V2** has passed the code (below).
7. **Then, in order:** the pre-release built from `main` once 5.4.7 has landed; Rohan's sitting; the link committed; the
   first climbs, which set the floors; the end verifier.

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
read of the diff; a must-fix V1 or V2 finds is fixed forward before the pre-release is built, and V2 passes the fix
round's commit. The builders' briefs carry 5.4.1 note 6 and 5.4.2 note 8's Go details. 5.4.0 lands on green tests on
both sides, CI and the orchestrator's read of the diff, which the run report says. A Sonnet agent runs 5.4.7's scripted
Claude Code sessions (its notes 1 and 2) and the agent's session of "5.4 done" check 14, Rohan's 8 Oct word; the builder
and the orchestrator read its report.

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
| Claude Code's auto memory holds one note for Bonsai's repo, and `init` writes `autoMemoryEnabled: false` (spec §10) | 5.4.7 | Not copied: contract §7.4 keeps facts about Rohan out of projects, and the preference is already this plan's rule; one line in `CLAUDE.md`'s "Working with Rohan" says it ("No memory note") |
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
protected: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**", ".github/**", "CLAUDE.md", ".gitignore", "go.mod", "go.sum", ".golangci.yml", ".goreleaser.yaml", "design/bonsai-spec.md", "design/contract.md", "design/one-pager.md", "design/format-review.md"]
person_only: ["bonsai.yaml", ".bonsai/lock.json", ".claude/**", ".github/workflows/release.yml", ".goreleaser.yaml"]
never_edit: []
ladder_floor: [0, 1, 2, 3, 4, 5]
ladder:
  - rung: 0
    name: guard
    kind: guard
    # ... every field of every rung written out, as the table below gives them
ratchets: {}
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
  so a climb's tests run in that climb. **The floors are set after the link:** `bonsai.yaml` lands with `ratchets: {}`
  (every floor reads as 0), and the switch writes the first green climb's counts in a commit under `T-5407`, then climbs
  at that commit (the switch's step 6). Counts taken before Rohan's install could sit above the first real climb's,
  since some tests skip once `/usr/local/bin/bonsai` exists (`cmd/bonsai/hook_test.go`'s bare-PATH case), and lowering a
  floor would then wait for Rohan right after his sitting.
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
  work (`.gitignore` among them: a line there hides a file from git, and so from rung 0), and the four documents Rohan
  approved. Person-only: Bonsai's own link (`bonsai.yaml`, the lock, Claude Code's settings) and the release files 5.7
  changes with Rohan. Not protected, on purpose: `formats/` (the schema compare and the manifest test hold it, and
  pieces change it often), `design/plan*.md` (planners write them beside builders), `STATE.md`, `records/` (the
  orchestrator's log) and every code folder (the rungs prove them). The guard's floor adds `.git`, `.bonsai/` and every
  nested `bonsai.yaml` whatever the lists say (5.3.2 rule 2), so `formats/active-task/` fixtures need a task's grant.
  `never_edit` is empty: Bonsai has no file that no agent may ever change (the closed run reports are held by the lock's
  `format0` hashes and rung 0), and a deny rule would stop the orchestrator's records commits too. In `agents` mode
  (Bonsai's, until the studio manages it) a person-only path is grantable as written (5.3.2 rule 4), so the rule is the
  orchestrator's, written in `CLAUDE.md`, and it is **stricter-only**, for the paths on the top-level `person_only` list
  (Bonsai's own link and the release files): a task may hold one in `bonsai.allows` for a change of a stricter kind (a
  floor raised to a green count; a rung, a floor entry or a protected path added) or one a section Rohan approved names.
  **Any other change to them loosens and waits for his word**, marked in the run report (contract §18 C): a floor
  lowered or `ratchets` emptied; a rung's `command`, `required`, `tests`, `ratchet` or `capture` changed; a rung, floor
  entry, protected or person-only path taken out; a name added to `ci_marked_tests`; a release file changed beyond its
  section. Why: an agent loosening the checks it is held to would be judging its own work. Each part's end verifier
  reads every change to `bonsai.yaml` in the part against this rule. The copies the floor makes person-only below the
  top (a fixture's `bonsai.yaml` under `formats/active-task/`, a nested `.bonsai/`) are not Bonsai's link: a task is
  granted them like protected paths, when its section's "Owns" names them.
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
  5). Its task and its run report are committed on `main` before its worktree is made, so the stop gate, which reads the
  run report in the session's own checkout (contract §13), finds it. Today's way, subagents, stays the default; the end
  verifier proves this one once (`T-5490`).
- **The landing rule** (`CLAUDE.md`, from the switch): the orchestrator fast-forwards a piece only when
  `.bonsai/local/ladder/<task>.json` reads `mode: local`, green, `git.sha` the branch's HEAD, `git.dirty` false, every
  rung of the floor and the task's `bonsai.ladder` green, and its `ladder` log record names `/usr/local/bin/bonsai` with
  the result's `sha256`; plus check 10's Windows half, and CI after the push. The orchestrator also reads the branch's
  diff against the section's "Owns" list for the piece and refuses one that leaves it: `records/tasks/`, `design/` and
  `STATE.md` above all, since they are free paths that rung 0 never judges, and a branch editing a task's grants or the
  plan's done checks would pass it. The run report quotes each landing's `git.sha`, the result's `sha256` and `green`,
  so the gate can be audited after the result file is overwritten or cleaned: each part's end verifier matches every
  landed commit to a green local `ladder` record in the log naming `/usr/local/bin/bonsai` with that `sha256`, and reads
  every task's `bonsai.allows` against its section's "Owns". The task moves to `done` once CI is green. The ladder is
  always run by its installed path, `/usr/local/bin/bonsai ladder --task <id>`, so no other `bonsai` on the PATH can
  make a proof. A builder climbs once its work is done, and again after a fix, not after every edit: each climb's
  minutes count in its run.
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
- **Bonsai's real home.** The link writes `~/.bonsai/` (the salt, this machine's record of the main checkout, and from
  the first climb `locks/ladder.lock`): the first write of Bonsai's work outside the repo and the scratch folders,
  accepted from the link on and named in `CLAUDE.md`'s Safety line. Sessions and climbs in the real repo and its
  worktrees run with `BONSAI_HOME` unset, the one exception to the rule that every run sets a scratch home: on a scratch
  home there is no record of the real main, so the guard would read it as not verified and grant nothing. Scratch
  projects keep their scratch homes, as before.

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
(`claude --settings '{"disableAllHooks": true}'`). The proof of a change is `bonsai ladder --task <id>` green, run by
the installed `bonsai`, plus CI on both sides; the standard library only (Windows calls through `syscall`); never
`go install`.

**`CLAUDE.md`**, changed at the switch (its own task grants it; Rohan's approval of this section is his word for it),
each rule worded as holding "from the link", since the file lands before it: the "Proof" line becomes the landing rule
above, with its audit; agents first: a builder runs the ladder, reads its `--json`, fixes what is red and climbs again,
and asks a person only where a refusal's `who` says so; the stricter-only rule; "How a session works" gains the tasks,
the names, the rule for running tasks side by side and a named session's task and run report committed first; the run
reports' frontmatter; the four labels' lines; the (ii) commands that are Rohan's; the way back, and a fix landing on the
interim proof while it is in use; the Safety line naming `~/.bonsai/` and `BONSAI_HOME` unset in the real repo; a
stamped build's scratch link with `--allow-exec`, and any other `bonsai` command the guard refuses as a person's, run
through a script file in `~/bonsai-checks/scripts/` (the guard's word splitter does not read scripts; the stamped
build's scratch root is the safety there); the removal of a worktree only with `git worktree remove` (a `rm -rf` of a
checkout's folder is a guarded delete); `bonsai update` of Bonsai's own link only in the main checkout (rung 0 refuses a
branch that changes `.claude/settings.json`); the Go line without `golang.org/x/sys`; and in "Working with Rohan" one
line: hand checks go to a Sonnet agent wherever an agent can, and Rohan is asked only for what needs a person. The block
`init` writes sits apart, between its markers.

**`.gitignore`** gains `.claude/settings.local.json`, Claude Code's own convention (this machine's global git excludes
already hold it, but a repo's own rule should not rest on one machine's file): Rohan's permission choices, and the
fallback file of 5.4.7 note 1, stay out of git.

**No memory note.** Claude Code's auto memory holds one note for this repo (agents run the hand checks). Contract §7.4
keeps facts about Rohan in the machine-side personal layer, never in a project, and the preference is already this
plan's rule ("What changes", item 8), so it becomes the one `CLAUDE.md` line above, and nothing is copied from Claude
Code's memory folder. `records/memory/` holds an empty `INDEX.md` only if `check` needs the index that `documents`
names.

**How Bonsai links itself, under Rohan's (a) and (ii).** After Rohan's install, the orchestrator links Bonsai's repo
with the installed copy, in the main checkout: first `/usr/local/bin/bonsai init --json` (the preview: every settings
line with its sentence; nothing written; exit 4 without `--yes`), each line read against this section, then
`/usr/local/bin/bonsai init --yes --json`. Under (a) the lines it writes name `C:\Program Files\Bonsai\bonsai.exe`, then
`/usr/local/bin/bonsai`, and the PATH is never read; a stamped test build cannot write them here, since it refuses to
link outside its scratch root (5.3.6 note 2), so only the installed build links the real repo. It needs no
`--allow-exec`: `packs: []` brings no pack code, and Bonsai's own lines at a first link are written on `--yes` alone
(5.1.1 rule 6). `init` keeps the committed `bonsai.yaml` as it is, comments included, byte for byte (the scratch climb's
`git diff` shows it, note 5.4.7). (ii) leaves it to an agent: a first link puts the guard in, and the command refuses
only `--allow-exec`, `unlink` and an update that removes or changes Bonsai's own lines (5.3.6 note 8). Chosen over Rohan
typing it: no safety reason needs a person to put a guard in, and his 9 Oct direction gives linking to agents. From then
on, in Bonsai's repo: an update that adds no code and leaves Bonsai's lines as they are (base's link in 5.5) is the
orchestrator's, in the main checkout, under a task granting `bonsai.yaml` and the lock; `--allow-exec`, `unlink` and an
update changing Bonsai's own lines are Rohan's, typed in his own terminal, until the studio manages Bonsai's repo.

**The order of the switch.**
1. 5.4.7's builder prepares its files in its worktree (an agent tries the way back in a scratch session, note 5.4.7),
   links a scratch clone of its branch with the final code's stamped build (a scratch home, inside the scratch root),
   climbs it, and reads the times (the timeouts come from these; the floors wait for the real climb, step 6). It lands
   on the interim proof, the last landing that does: check 10 and CI.
2. Before the build: `main`'s working tree holds nothing uncommitted but 5.4.7's open run report, which is format 1;
   every other run report is closed and committed; part 0's report has been edited or left; Claude Code's version and
   the user settings hashes are recorded.
3. The orchestrator builds the pre-release ("Rohan's sitting", below) and sends Rohan his batch.
4. Rohan installs it and sends `bonsai --version`'s line.
5. The orchestrator links the repo (above) and reads what `init` wrote (`git status`, `git diff`: the lock,
   `.claude/settings.json` with every line in (a)'s form, the block in `CLAUDE.md`, `.bonsai/.gitignore`, the two
   tables), runs `bonsai check --json` (no finding), commits it on `main` (`bonsai: Bonsai links itself`), pushes and
   reads CI. Claude Code most likely takes up the new lines in the running session (5.4.7 note 2 measures it); if not (a
   subagent's file-tool edit of `bonsai.yaml` is not refused), it asks Rohan to restart it with `claude --continue`.
6. The first climbs: after a records commit, with the main checkout clean and `T-5407` running (its `bonsai.allows`
   holds `bonsai.yaml`), `/usr/local/bin/bonsai ladder --task T-5407` in it is green with every floor at 0. The
   orchestrator writes that climb's counts into `ratchets` in a commit under `T-5407` (stricter-only: floors raised from
   0), and climbs again at that commit: green, each ratchet at its floor; its result, its `ladder` record and
   `hook start`'s next context agree. `T-5407` moves to `done` in a records commit with the tables.
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
   - `ratchet`: the ratchet's name (`[a-z_][a-z0-9_]{0,39}`, as today), naming a floor in `ratchets` (none reads as 0),
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
   the command that runs it alone (5.4.2 note 12); `new_tests.base`'s description says it is `""` when no base could be
   found (5.4.5 note 2), since writing `null` there would retype it, a new major. No field is removed or retyped.
4. **The log's `ladder` record**, by description (contract §11: "the result's project-relative path, `sha256:<hex>` of
   its bytes, its `green`, and `task`"): `target` the path, `text` `sha256:<hex>`, `ok` its `green`, `task` its task,
   `kind` its `mode`, `category` `Ladder`, and `bonsai_path` and `bonsai_sha256` the runner's own (so the record names
   which `bonsai` made the result).
5. **Go tables**, beside the ladder's type, each word with its line: the `tests` forms; rung statuses (`green`, `red`,
   `skipped`, `error`, and `pending`, which today's results hold and Bonsai never writes: no rung field marks one as not
   built); git integrity's findings (5.4.6's names); the test-file convention per `tests` form, which 5.4.5 and 5.4.6
   both read (5.4.5 note 4). Printed by `check --schema` and the reference page; the schemas' descriptions name the
   command and copy no word (5.1.3's rule).
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
6. **Go details for the brief** (the review of this section): on Linux, `Pdeathsig` fires when the operating-system
   thread that started the child exits, so the start runs on a locked thread (`runtime.LockOSThread`); adopted orphans
   are reaped by their own pids, never by `wait4(-1)`, which would race `os/exec`'s own wait; `cmd.WaitDelay` bounds a
   child that holds the output pipe after the shell exits. On Windows, `os/exec` closes the main thread's handle, so the
   suspended process is resumed through `NtResumeProcess` on a handle opened by its pid.

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
4. **Skips and marks, as today's runner judges them:** a mark is a skip reason starting `<word>-only:`. In a local run,
   a skipped test with a mark, or one named in `ci_marked_tests`, makes its rung red (a task's proof runs them all); an
   unmarked skip passes. Under `--ci` a skip counts only when its name is in `ci_marked_tests` and its reason carries a
   mark (Bonsai's list holds names, not today's `{name, mark, file}`); any other skip, and any todo, is red. `skipped`
   is filled under `--ci` only.
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
   with its holder (`flock` on Linux; `LockFileEx` on a byte range on Windows, so a waiting run can still read the file,
   which a file opened with no sharing would not allow), and holding the holder's task, checkout, pid, start time and
   its budget (the sum of its requested rungs' timeouts). A second run prints the holder and waits while the holder
   lives and is inside its budget, then refuses (exit 4, naming the holder and the next step). A dead holder's lock is
   taken at once. The orchestrator's rule (one ladder at a time on this PC) still covers WSL's and Windows' homes
   together.
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
    `ladder` (5.4.0 note 9) tells an agent where its proof stands without a climb. The `ladder` word and its error words
    reach `bonsai --help --json` (5.1.10) from its flag-table entry; 5.5's "operating Bonsai" skill in `base` teaches
    this loop (climb, read the `--json`, fix, climb again), named here so 5.5 carries it.
13. **The base, once per run.** The runner works out the base branch and the merge base once, before any rung, and
    passes both to every kind through `kinds.go`, so rung 0, the new-tests check and git integrity judge from one
    answer. **The base branch** is the branch the verified main checkout has checked out, read from its HEAD file
    (5.2.2's reading): Bonsai's schema has no `base_branch`, and Bonsai's work and the studio's both land on the branch
    their main checkout holds. A main on a detached HEAD gives no base, which each kind reports as "not judged".

**5.4.3, rung 0.** Spec §9: "the diff against the named task's grants"; "Rung 0 also refuses a task branch that changes
a generated table, and any file from `.bonsai/local/` that is tracked or staged"; spec §4: "CI and rung 0 run
[`check`]".
1. **What changed:** on a branch other than the base, every path the branch's commits changed since the merge base
   (`git diff --name-only <base>...HEAD`, both sides of a rename), and on any branch what is staged, unstaged or
   untracked and not ignored (today's `changedFiles`). The base branch and the merge base come from the runner (5.4.2
   note 13). With no base, the branch part is "not judged" in the reason, and the working tree is still judged.
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
   judged here": a proof is of a commit. Warnings never count. One local finding does count: `disableAllHooks: true` in
   a local settings file of the checkout or of its verified main (where Bonsai's subagent builders take their hooks
   from) makes rung 0 red, since work done with the guard off is not proved by it. So nothing lands while such a file
   exists; Rohan's way back is the `--settings` line, which writes nothing, and the file is only 5.4.7 note 1's
   fallback.
5. **Its outcome:** green with the count of paths checked (in `captures`, `checked`); red with each refused path and its
   rule, at most 50, in the reason. The result never holds a whole diff.
6. **`bonsai.*` labels with no definition in force:** 5.4's start reads how 5.1.5 and 5.3.2 landed. If the guard and the
   stop gate already read `bonsai.allows` and `bonsai.ladder` by contract §5.6's meaning without a definition, this
   piece reads them through the same code. If they grant only from a defined label, this piece adds one function, the
   value of a `bonsai.*` label on a task held to §5.6's kind with or without a definition, used by all four readers, and
   V1 covers the guard's change (guards and hooks). The runner (5.4.2) reads `bonsai.ladder` through it too, so this
   piece lands that function first, in a commit of its own, and 5.4.2 waits for it.
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
2. **The base:** the merge base of HEAD with the base branch, from the runner (5.4.2 note 13). On the base branch
   itself, where that is HEAD, the base recorded by the task's last result on its branch (5.4.6); none recorded: "check
   not run: no base", with `new_tests.base` `""` (5.4.0 note 3).
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
4. **Test files, by the form's own convention**, from 5.4.0's table: `go-test-json`, the `*_test.go` files and files
   under a package's `testdata/`; `tap`, node's test runner's default patterns (`*.test.*`, `*-test.*`, `*_test.*`,
   `test-*.*`, `test.*` and files under `test/`, for `.js`, `.mjs`, `.cjs` and `.ts`); `junit`, the files the XML names;
   none found: "check not run: the test files cannot be told".
5. **Written:** `new_tests` `{base, count, failed_on_base, passed_on_base, not_run}` on the ratchet (5.4.0 note 3); the
   names, at most 20 per group, in the human output and the Bless ask's `why`. The rise is offered for Bless whatever
   the numbers (spec §9: "ratchets are Rohan's").

**5.4.6, git integrity.** Spec §9: "records the task's base commit and flags changed test files, `assume-unchanged` and
`skip-worktree` entries, edits to `.git/info/exclude`, new stash entries and a rewritten base. Not required at first: it
informs the verifier."
1. **What it records**, in the rung's `captures` (5.4.0 note 3): `base` (the merge base, from the runner, 5.4.2 note
   13), and the stash count and `info/exclude`'s hash, so the task's next climb can compare.
2. **What it finds**, each by name and count, the names (at most 20 a kind) in the human output: changed test files
   (5.4.0's convention); `assume-unchanged` and `skip-worktree` entries (`git ls-files -v`); lines in the common git
   folder's `info/exclude` other than git's own comments; a file present that only `info/exclude` or the global excludes
   (`core.excludesFile`) ignore, by the source `git check-ignore -v` names; stash entries the task's earlier result did
   not count; a base the earlier result recorded that is no longer an ancestor of HEAD.
3. **It informs; it never turns red for what it finds.** Its status is `green` with its findings listed, or `error` when
   git cannot be read. A finding that hides a change from git also makes the runner's `dirty` true: an
   `assume-unchanged` or `skip-worktree` entry, or a file ignored by `info/exclude` or `core.excludesFile` alone (the
   fix is a line in the repo's own `.gitignore`, which is protected, so it shows in the diff). Otherwise a test run on a
   change git does not show would claim a clean commit, so the stop gate (which refuses `dirty`) catches it while this
   rung stays informative. Chosen over a red rung, which the spec defers, and over leaving `dirty` to git's status,
   which those entries fool.

**5.4.7, the switch.** Written above, in "The switch, written before it happens". The builder's proof (notes 1, 2 and 5
run by a Sonnet agent, as "Who builds and verifies" says):
1. **The way back, tried:** in a scratch project linked by a build stamped with an empty place (every call refused,
   5.3.6's `missing`), a session started through `claude-here` with `--continue --settings '{"disableAllHooks": true}'`
   runs a free edit and a shell call, and keeps the earlier session's conversation; the status line's absence in such a
   session is noted for Rohan's part. The fallback, tried only if `--settings` does not switch the project's hooks off
   on the version in use: `.claude/settings.local.json` holding the same, which `bonsai check` reports and rung 0 counts
   (5.4.3 note 4); Rohan's part then offers that file, with those words, before his sitting. The Claude Code version and
   the user settings hashes recorded.
2. **Settings taken up mid-session:** in a scratch session, whether a `.claude/settings.json` with hook lines written
   after the session started is applied in that session (Claude Code watches its settings files), applied after review,
   or only at the next start. The orchestrator's step 5 above follows the answer.
3. **A scratch climb of the switch:** a clone of the branch in the scratch root, linked by the final code's stamped
   build on a scratch home (`init --new-id`, `packs: []`, no `--allow-exec`); `git diff` shows `bonsai.yaml` unchanged,
   comments and all; `check` with no finding; then `ladder --task T-5407` green, every rung, with the times the timeouts
   take.
4. **The files:** `bonsai.yaml` validates, every line commented; the task file parses under its format; nothing private
   in any of them.
5. **Scratch links from a guarded session:** from a session in Bonsai's linked repo, a stamped build's
   `init --allow-exec` of a scratch project run through a script file in `~/bonsai-checks/scripts/` links it, while the
   same command typed on the shell tool's line is refused by the guard (5.3.2 rule 7), as `CLAUDE.md` then says.
6. **Hours.** The piece carries a lot for 1-2 hours (the files, the tries, the scratch climb, the link and the first
   climbs). Its run report keeps its running hours; past 2, the orchestrator says so before the end verifier is briefed.

#### Rohan's sitting (spec §17 step 8)

**Before it, the orchestrator** builds the pre-release from a clean clone of `main` at the switch's commit, never from a
worktree (gate report §2.11; "What changes", item 7):

```bash
git clone ~/Servers/Bonsai ~/bonsai-checks/prerelease/src-<commit>
git -C ~/bonsai-checks/prerelease/src-<commit> checkout --detach <commit>
cd ~/bonsai-checks/prerelease/src-<commit> && CGO_ENABLED=0 go build -trimpath -buildvcs=true -o ~/bonsai-checks/prerelease/bonsai ./cmd/bonsai
go version -m ~/bonsai-checks/prerelease/bonsai
sha256sum ~/bonsai-checks/prerelease/bonsai
```

`go version -m` must show that `vcs.revision`, `vcs.modified=false`, `CGO_ENABLED=0` and no fault tag. `-trimpath` and
`CGO_ENABLED=0` (the end verifier's build pins both too) keep the build's folder and the C toolchain out of the binary,
so it is the same bytes wherever it is built and the end verifier's own build matches the fingerprint. Between V2's
commit and this one nothing may change but `records/`, `design/`, `STATE.md`, `CLAUDE.md`, `CONTRIBUTING.md`,
`bonsai.yaml` and `.gitignore` (`git diff --stat <V2's commit>..<commit>` read against that list): embedded schemas,
`go.mod` and `go.sum` count as code. A fix after V2 is passed by V2 before the build. The fingerprint goes into 5.4.7's
run report, committed, before Rohan's batch is sent; before a later pre-release, the build Rohan has is copied to
`~/bonsai-checks/prerelease/previous/` and its number is already in that earlier report. Then the orchestrator sends
Rohan the batch in his part above, with the number in place, in one message.

**After it:** his words and `/usr/local/bin/bonsai --version`'s line go in 5.4.7's run report; the orchestrator runs the
switch's steps 5 and 6.

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
   `local/` file red; a `check` finding red and a machine finding listed as not judged; `disableAllHooks: true` in a
   local settings file of the checkout or its main red; no task under `--ci`: a protected change red; the guard and rung
   0 agree on the verifier's own table of paths.
8. **Ratchets:** a count below its floor red; a ratchet with no count red and a `check` finding; a capture that does not
   compile a `check` finding.
9. **Bless:** in a fixture main checkout on its base branch, clean, at HEAD, green and above a floor: one ask
   `ladder:bless-<task>` with `source: ladder` and contract §9.1's `data`, and its `ask` log record; none on a branch,
   dirty, red or under `--ci`; a later run with no rise resolves it; `bonsai ask --type Bless` still refused.
10. **New tests:** for each form, a new test failing at the base counted; one passing there counted as proving nothing
    new; a Go test calling new code counted as not run; a failing `base_setup` makes every one not run; the temporary
    worktree gone and `git worktree list` as before; the numbers in the result and on the Bless ask.
11. **Git integrity:** each finding on its fixture repository (Windows git on Windows); a clean repository finds
    nothing; the rung green with findings; `dirty` true with an `assume-unchanged` entry, with a file ignored by
    `info/exclude` alone and with one ignored by `core.excludesFile` alone; a file the repo's `.gitignore` ignores
    changes nothing.
12. **No `golang.org/x/sys`:** `go list -deps ./cmd/bonsai` names none; the hook line's p50 and p95 on both sides
    against 5.3's.
13. **Unattended:** each refusal of `bonsai ladder` (no task, the lock held past its wait, an undefined rung, a stamped
    build outside its root, an unverified main) exits with its code and an `error` whose `next.do` the verifier runs as
    written, and whose `who` is `agent` but where the words' table names a safety reason; with no terminal nothing waits
    for input; a red climb's `--json` gives `next`, and each failing test's `run` runs that test alone on both sides;
    `status --json`'s `ladder` shows the active task's last result, and `stale` after a new commit.
14. **An agent alone:** a scripted `-p` session through `claude-here`, run by a Sonnet agent, in a scratch project
    linked by the build, holding a running task whose test is right and whose code has a bug, with a ratchet at the
    fixture's count. The agent is told only "make the ladder green for <task>; the `bonsai` to use is <its path>, and
    `<its path> --help --json` lists its commands" (no operating skill until 5.5). It climbs, reads the result, fixes
    the code and climbs green, with no person and no prompt; the verifier confirms the test file unchanged (git
    integrity lists changed test files), and deleting the test would stay red (the count falls below its floor). The run
    report keeps the session's commands. On WSL, and on Windows if Claude Code's login there is back.
15. **The break-it:** the verifier's own process trees on both sides and its own repositories for rung 0, ratchets and
    git integrity, beyond the builders' tables: no rung leaves a process for the next, and no change outside the named
    task's grants climbs green.
16. **The pre-release:** the installed file's SHA-256 equals the number Rohan was sent and the verifier's own
    `CGO_ENABLED=0 -trimpath` build of the commit from its own clone; `go version -m` shows that commit, unmodified;
    `which -a bonsai` lists `/usr/local/bin/bonsai` alone; `bonsai --version` names the commit.
17. **The link:** Bonsai's committed `bonsai.yaml` holds the rungs and lists this section sets and validates, and its
    floors are the first green climb's counts, set in a commit under `T-5407` after the link, with a green climb at that
    commit; the lock, `.claude/settings.json` (every Bonsai line in Rohan's (a) form, naming the two places and no other
    path), the block and the tables are committed; `bonsai check` in the main checkout and in a fresh worktree: no
    finding; `.gitignore` holds `.claude/settings.local.json`; `CONTRIBUTING.md` says a clone without `bonsai` is
    refused and how to work round it; `CLAUDE.md` holds the landing rule, agents first, the stricter-only rule, the task
    and name rules, the (ii) commands, the way back, the real home and the script rule for scratch links; the link made
    by the orchestrator with the installed build; `~/.bonsai/` holding the salt, the machine record and the ladder's
    lock.
18. **Guarded:** with `BONSAI_HOME` unset (the real home, as every session in the real repo), in a scripted session
    through `claude-here` in a fresh worktree of Bonsai, with only `T-5490` running (it grants nothing): a file-tool
    edit of `bonsai.yaml`, `go.mod` and `.github/workflows/ci.yml` refused, each naming why; a free file edited;
    `rm -rf junk/*` refused; a scratch project linked on a scratch home and deleted by name, allowed; the records in the
    main checkout's `.bonsai/local/log/`. In a scratch clone of Bonsai linked by the installed build on a scratch home,
    a running task granting `go.mod`: that edit allowed.
19. **Proven and gated:** `T-5407`'s green result in the main checkout, its record and `hook start`'s context agreeing;
    in a fresh worktree on a branch of its own (`BONSAI_HOME` unset), a `-p` session with `BONSAI_TASK=T-5490`, started
    before any climb of `T-5490` or after a new commit, so no result proves its HEAD, is blocked at its stop, and stops
    once `/usr/local/bin/bonsai ladder --task T-5490` is green at its HEAD; a session with nothing named stops freely
    (the orchestrator's case); a subagent's end is not gated.
20. **Rohan's commands held:** in a scratch clone of Bonsai linked by the installed build on a scratch home (never the
    real checkout), an agent-shaped run (`CLAUDE_CODE_CHILD_SESSION` set) of `update --allow-exec --yes`, `unlink` and
    an update changing Bonsai's guard line is refused by the command; the same with the variable unset is written, run
    through a script file in `~/bonsai-checks/scripts/` from the verifier's guarded session (the guard refuses those
    words on the shell tool's line, 5.3.2 rule 7); 5.4.7 note 5's scratch link through a script works from a guarded
    session.
21. **The way back:** the lines in Rohan's part are exactly those 5.4.7 tried, and the verifier tries the first once
    more in a scratch project with every call refused.
22. **Check 10, the ladder and CI:** `go test ./...` and `go vet ./...`, plain and tagged, on WSL and natively on
    Windows, run by the verifier; `/usr/local/bin/bonsai ladder --task T-5490` green on the final commit; CI green on
    it.
23. **Stop lines and records:** 5.4's hours under 57, this section's planning and review included; step 5's Windows-only
    tally; option rounds (none); nothing written or run in the studio's checkout or in Mimas (the scripts and the run
    reports' commands read); nothing written outside the repo and the scratch folders but `~/.bonsai/` from the link;
    every landing since the switch matched to its green `ladder` record and each task's `bonsai.allows` to its section's
    "Owns" (the landing rule's audit); the user settings hashes around every Claude Code run; the Haiku audit of
    `.bonsai/sessions.md` against the run reports after the switch; every change to `bonsai.yaml` since the switch
    stricter or naming Rohan's word. **Nothing private:** a grep of the diff, the commit messages and the task files.

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
- **Bonsai's real home.** From the link, Bonsai's work writes `~/.bonsai/`, and the real repo's sessions run with
  `BONSAI_HOME` unset; a scratch home set by habit in a real worktree reads main as not verified and grants nothing
  (fail closed, with its reason).
- **Breakdowns land on the interim proof.** While the way back is in use the installed runner may be the broken part, so
  the fix lands on check 10, CI and a fresh verifier, as before the switch.
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
- **`workspace.schema.json`: a rung's `ratchet` is "not in Bonsai's sources":** today's ladder, read at the studio's
  commit, has it as a name with a capture of that name; 5.4.0 fixes it by description, with no new major.
- **Contract §5.6 and this plan's 5.1.5: `bonsai.*` definitions come with base:** until 5.5, Bonsai's readers act on
  §5.6's meaning with no definition in force (the switch, "Labels").
- **Contract §7.1: "A run report lives on the builder's branch":** Bonsai's are the orchestrator's, in the main
  checkout, format 1 from the switch.
- **Contract §7.2: STATE at `.bonsai/STATE.md`:** Bonsai's stays at the root until base's template (5.5).
- **Spec §14, "How Bonsai's work is proven":** "Bonsai's guard then covers its own checkout too": the guard and the
  delete check cover every session there; the stop gate covers only named sessions, so the orchestrator's builders are
  held by its landing rule (the switch, "Tasks and names").
- **Spec §17 step 8:** the install stays Rohan's; the link after it is the orchestrator's. The fingerprint is checked on
  the installed file with `sha256sum -c`, not on the scratch file before the install, which an agent could change in
  between. Under (a) the hook lines never read the PATH, so `which -a` guards what an agent runs by name (the ladder is
  called by its full path anyway).

### Step 5.5: the packs (19-30 h, re-ask at 39)

**Rohan's (B), and what else of his changes.** This section comes to Rohan: it makes his roles, lanes, protocols and
templates public in a new repo, `LastStep/bonsai-workflow`, and three of his templates (task, run report, STATE),
reworded, public in Bonsai's own `base` pack. Beyond that approval it changes these things of his:
- He reads the `workflow` repo on GitHub while it is still private and makes it public with one line; an optional
  second line, after it, protects its `main`.
- About 3 minutes on Windows: Claude Code started once through the scratch launcher, in a scratch folder linked to both
  packs, to say yes to its question about trusting the folder. His Windows login is due before 5.3 (already on his
  list); at 5.5 it is repeated only if it has lapsed.
- The first release tags of both packs are his word, with 1.0 at 5.7.
- From 5.5's end his own Claude sessions in Bonsai's repo cannot read his key, token and login files, nor change his
  own Claude settings, his `~/.claude/CLAUDE.md`, his shell's start files or git's and SSH's settings: he changes those
  himself. Bonsai's STATE moves to `.bonsai/STATE.md`.
- If the studio links the `workflow` pack before 1.0, it first needs a newer pre-release `bonsai` installed in WSL
  (5.4's cannot read the pack): his install, at the studio's link (step 7), not in 5.5.

**Approved on 9 Oct, with a rule of his kept as 5.4 and the spec have it.** An earlier draft let agents move a project
to a newer version of a pack themselves; Rohan declined it. Moving a project to another version of a pack stays his
step, and so do adding a pack to a linked project and taking one out: an agent prepares the change and hands him the
exact line, and makes it only on his word ("Moving a pack's version: a person's step", below).

No repository secret (the "possibly" in his list is settled: none), no install and no password in 5.5 itself (it
changes no guard, stop gate or ladder code), no option round. The hours (19-30) and the re-ask line (39) are the
spec's; the order of the parts stands. Every format change is an addition: a pack can put its always-read file in the
folder a project keeps its protocols in (a pack file's path may start with `<protocols>/`), and descriptions.

#### For Rohan, in plain words

**What 5.5 gives you.** Bonsai's own pack and your way of working become two Claude Code plugins that any of your
projects can take with a few lines in its `bonsai.yaml`:
- **`base`**, Bonsai's own pack, in Bonsai's repo (`packs/base/`). It holds the meaning of Bonsai's four task labels,
  the templates for a task, a run report, STATE, a memory note and `bonsai.yaml`, the page on how long records are kept,
  a starter CI file, and **the walls**: rules that stop Claude reading your key, token and password files (SSH keys,
  your GitHub and cloud logins, Claude's own login and global files, Bonsai's secret salt), or changing your own Claude
  settings and memory and the files that start programs (your shell's start files, git's and SSH's settings). It also
  holds an **"operating Bonsai" skill**: the instructions an agent needs to link a project, update it, fix what `bonsai
  check` finds, read where things stand, edit `bonsai.yaml`, prove its work with the ladder and ask you a question,
  without asking you how. It carries no code that runs.
- **`workflow`**, your pack, in its own new public repo: your five roles (builder, verifier, researcher, producer,
  playtest analyst), your three lanes, your protocols (the session-start checklist that every session reads, and the
  rest read when needed), your document templates (plan, one-pager, decision record, options write-up, playtest, brief,
  and two new short ones for a spec and the bugs file), your labels, and walls for the studio's own secret files.
- **A template for new packs** (`packs/template/`), with its own checks on GitHub and a release on a tag, so the next
  pack (a game pack for Unity, later) starts from a copy. Bonsai's own checks run the template on every commit.

**The new public repo, and exactly what goes into it.** `LastStep/bonsai-workflow`, the name the spec gives (the pack's
id is `workflow`, and it sits beside `LastStep/bonsai-test-pack`). Under the MIT license, as Bonsai is. From your
studio's files at commit `7017d63`, read only (nothing is written in the studio's repo):

| From your studio | Count | Into the pack | What changes |
|---|---|---|---|
| Roles (`studio/roles/`) | 5 | `agents/`, one file each, loaded as `workflow:builder` and so on | Studio paths and tools become Bonsai's (below); each role names the protocol skills it loads at start |
| Protocols (`studio/protocols/`) | 6 of 8 | `session-start` as the one always-on file; `lanes`, `verification-ladder`, `reward-hacking-guards`, `reporting`, `adr` as skills | The same; the two Unity protocols stay out (a later game pack) |
| Templates (`studio/templates/`) | 6 of 10 | plan, one-pager, decision record, options, playtest, brief, as documented templates | Each gains a table of its fields; the `project:` field goes (a project is its repo) |
| Templates for Bonsai's own files | 3 of 10 | task, run report, STATE go to `base` instead (Bonsai's formats), in Bonsai's own public repo | Rewritten to the formats you confirmed on 7 Oct, naming no person ("a person"); public when their pieces land, each after a fresh agent's check for anything private, before your read of the new repo |
| `studio/templates/game.yaml` | 1 | nothing | Replaced by `base`'s `bonsai.yaml` template |
| The role copies in `.claude/agents/` | 5 | nothing | Copies of the roles; the studio removes them when it links (spec step 7) |
| New | 12 files | the plugin's manifest, the pack's manifest, the lanes file, the labels file, its block, its README, its license, its line-ending rule, its two check files (from the template), and short templates for a spec and the bugs file | Written from your lanes protocol, the formats contract and the shape of today's files |

**What is rewritten, and why.** Nothing in those files is secret: no password, no home folder, no machine name, no email
address. What is studio-shaped is rewritten so the pack works in any of your projects: the studio's folder names
(`studio/tasks/` and the rest) become "the folders your `bonsai.yaml` names"; the studio's own tools become Bonsai's
commands (its ladder script becomes `bonsai ladder`, its task variable becomes Bonsai's, its "allows_assets" becomes the
task's grant label); the studio's services (Discord, email, the bridge, the Desk) leave the protocols, since
notifications and screens are the studio's; the "golden rule 8" the protocols cite is written out in full; an example
task id in a template becomes a made-up one; the ledger lines stay, worded "where a project keeps a ledger". Your name
stays in the `workflow` texts, as in Bonsai's own public files; `base`, Bonsai's pack for any project, names no one and
says "a person". The game wording (players, playtests, "Rohan plays it") stays: it is your workflow. So do your weekly
rituals and the daily brief's hour, in the producer's role and the reporting protocol: once public they describe your
week, so say when you read the repo if any should go (the list you get marks them). `base`'s STATE template leaves them
out: they are your studio's way, not every project's. Left out: the two Unity protocols (they name Mimas's folders; they
go into a game pack later) and `game.yaml`'s template. The walls in `workflow` name three files in the studio's secret
folder (`~/.trinetra/token`, `deploy_ed25519` and `salt`): names only, which Bonsai's public spec already holds.

**How you see it before it is public.** A fresh Opus agent reads every file, and every commit from the first, for
anything private first. Then the orchestrator creates the repo **private**, pushes it, and sends you its link with a
short list, file by file, of what changed from your studio's version and why. You read it on GitHub (your phone works).
Ask for any change; it is made, read again for anything private, and pushed. The repo's commits carry the same author
line as Bonsai's own public commits (your name and email address, already public in Bonsai's history): nothing new is
shown. When you are happy, and the orchestrator has told you that every change since the check has been read, one line
in WSL makes it public:

```bash
gh repo edit LastStep/bonsai-workflow --visibility public
```

(The orchestrator checks this machine's `gh` version before it sends the line. Newer versions of `gh` ask for one more
flag, `--accept-visibility-change-consequences`; on one of those, the line it sends carries it.)

Then, optional and recommended, one more line that stops anyone, an agent or you, rewriting or deleting the repo's
`main`, which matters because projects pin the pack by its exact commit. It comes after the public line: GitHub offers
this protection on a private repo only on a paid plan. Unlike Bonsai's own `main-protection`, which lets an admin
through (so an agent acting as your account too), it lets no one through; you can remove it in the repo's settings:

```bash
echo '{"name":"main-protection","target":"branch","enforcement":"active","conditions":{"ref_name":{"include":["~DEFAULT_BRANCH"],"exclude":[]}},"rules":[{"type":"deletion"},{"type":"non_fast_forward"}]}' | gh api -X POST repos/LastStep/bonsai-workflow/rulesets --input -
```

Why these are yours: making a repo public and adding a ruleset are GitHub settings, which agents never change.

**No secret.** The packs' checks on GitHub need no Claude login and no model key. If Claude Code's plugin check
(`claude plugin validate`) turns out to need a login on GitHub's machines, it runs on this PC instead, before every
pack change lands, and the GitHub checks do the rest (a pack's own `bonsai check`, and a test project linked to it).
Evals (tests of a pack that call a model) also run on this PC, never on GitHub, so no key leaves it.

**Your Windows sitting (about 3 minutes).** 5.5 needs Windows sessions (the walls must be tried on both sides), and a
project's plugins install on Windows only once a person has trusted its folder. When the orchestrator asks, in a normal
PowerShell, one line at a time. If Claude Code's Windows login has lapsed again (the orchestrator says so; it is due
before 5.3), first, in your home folder:

```powershell
claude
```

then type `/login`, sign in, and type `/exit`. Then:

```powershell
cd "$env:USERPROFILE\bonsai-checks\packs\project-w"
& "$env:USERPROFILE\bonsai-checks\bin\claude-here.cmd"
```

Claude Code asks whether to trust this folder: say yes. Then type `/exit`, and tell the orchestrator whether the
question came and what it said. The launcher starts Claude Code on the scratch folders' own Bonsai home and plugin
folder, so nothing of yours is touched; the orchestrator records your Windows Claude settings file's fingerprint before
and after. Why yours: trusting a folder is Claude Code's question for a person, and Bonsai never answers it. Your 5.3
sitting may already have shown the question on Windows; this time the folder is linked to both packs, whose plugins wait
for that yes. Everything else on Windows is the agents'.

**The first release tags are yours, later.** Until then projects name each pack by its exact commit, which works the
same. `base-v1.0.0` (on Bonsai's repo; it starts no release) and `v1.0.0` (on the workflow repo; its checks then make a
GitHub release) come with your word for 1.0 at 5.7; agents never tag.

**What changes for your projects.**
- **Bonsai's own repo** takes `base` at the end of 5.5 (spec step 6). An agent links it; it needs no consent to code
  (`--allow-exec`), since `base` carries no code that runs, so under your answer (ii) it is not one of your commands.
  After it, in `~/Servers/Bonsai`, Claude (yours included) cannot open your key, token or login files or your global
  Claude file (`~/.claude.json`, which can hold tokens), and cannot change your own Claude settings
  (`~/.claude/settings.json`), `~/.claude.json`, your own `~/.claude/CLAUDE.md`, your shell's start files (`~/.bashrc`
  and the like), git's settings (`~/.gitconfig`) or your SSH folder: change those yourself (`/config`, `/permissions`,
  `/memory`, or by hand). Programs such as `gh`, `git` and `ssh` still read their own files. Claude Code installs the
  `base` plugin for Bonsai's repo in its own plugin folder (instructions only). Bonsai's `STATE.md` moves to
  `.bonsai/STATE.md`, where every project keeps it.
- **The studio and Mimas:** nothing changes until each links (steps 7 and 8, their own plans). Then their roles come
  from the `workflow` pack, and their own copies go.
- `workflow` is not linked into Bonsai's own repo: Bonsai's team works by `CLAUDE.md`'s briefs, and your roles are for
  your studio's projects. It can be, later, on your word.

**Hours, order and your other steps.** 19-30 hours, re-ask at 39, as the spec has them; 5.5 after 5.4 and before 5.6.
Your steps in 5.5: approve this section (you did, on 9 Oct); read the private repo and make it public (one line, plus
the optional one); about 3 minutes on Windows, starting Claude Code once through the scratch launcher and saying yes to
its trust question (with the login first only if it has lapsed). None needs a password. Later, in any project linked to
a pack (Bonsai's own repo included), moving it to another version of a pack, adding a pack or taking one out is yours:
an agent hands you the exact line, and makes the change on your word.

#### What exists, and what 5.1 to 5.4 will have added

**On `main` at `07320fe`** (5.1.0 to 5.1.5 landed; read each package's doc comment):
- No `packs/` folder. The test pack, `LastStep/bonsai-test-pack`, at F (`f8e8f18`), is the shape of a pack today:
  `.claude-plugin/plugin.json`, `agents/`, `skills/<name>/SKILL.md`, `hooks/`, `bonsai/pack.yaml` (every key
  commented), `bonsai/block.md`, `bonsai/files/`, `README.md`, `.gitattributes`.
- `formats/schemas/pack.schema.json` (set 4): `files`, `hooks`, `deny`, `documents`, `protected`; a `files` entry's
  `path` is project-relative and may not be Bonsai's own, `CLAUDE.md`, or a place Claude Code or git loads on its own.
  `labels.schema.json` and `lanes.schema.json` for `bonsai/labels.yaml` and `bonsai/lanes.yaml`.
- **The reader refuses `<` today.** `internal/workspace/pack.go` checks every `files` entry's `path` with
  `CheckRelPath` (`internal/workspace/paths.go`), which refuses `<`, `>` and the other characters Windows cannot hold,
  though the schema's pattern allows them. So today's `bonsai`, and the pre-release Rohan installs at 5.4 (built before
  5.5), refuse a `pack.yaml` with a `<protocols>/` path: `workflow`'s. The lock's paths are resolved ones, which
  `CheckRelPath` keeps checking.
- `internal/engine/declares.go` (5.1.5): what a pack declares, copied into the lock (lanes, document kinds, labels,
  protected paths, hooks, deny rules); a pack's labels in its own namespace or `bonsai`.
- `internal/engine/block.go` (5.1.5): the instruction block imports, as `@<path>`, each pack file whose path lies under
  `bonsai.yaml`'s `documents.protocols`; it renders the packs' label definitions; at most 40 lines.
- `internal/engine/config.go` (5.1.5): `init` writes `bonsai.yaml` from a built-in template with a comment on every
  line; documents default to `work/` (`work/tasks`, `work/protocols` and the rest); `packs`, `ladder` and
  `ladder_floor` empty. `init` takes `--name`, `--source`, `--ref` (a tag or a 40-character commit) and `--path` to
  link a first pack. No code writes `.bonsai/STATE.md`.
- `internal/workspace/documents.go`: Bonsai's own document kinds and their places, the one home of that table;
  `.bonsai/STATE.md` is the `state` kind's fixed file.
- 5.1.5's run report places one item here: "a project's own `documents.protocols` folder and a pack's always-on files:
  5.5".
- CI: `test`, `windows`, `lint`, `govulncheck`, CodeQL.

**What 5.1 to 5.4 will have added** (from this plan's notes; none of it is built yet. 5.5's start re-reads each against
what landed, and 5.5.0's run report records any difference that changes a note below):
- 5.1.6: `check`'s findings in one table, among them labels against their definitions, a `version` in a pack's
  `plugin.json`, the block over 40 lines, a named path that does not exist (the block's `INDEX.md` import exempt);
  every finding's `next.do` a command. 5.1.7: `unlink`, taking a pack out, trust's `waiting` as a person's step, the
  key order of `.claude/settings.json`. 5.1.8: the tables and `check --write`. **5.1.9: `check --pack`** (the pack's
  files held to their schemas; every YAML key commented; each template skill's fields table equal to its template's
  frontmatter and, for a Bonsai format, to the schema; each deny rule's `why`; no `version` in `plugin.json`; the
  block's 40 lines; no `bash` by name; every file a hook runs in `runs`) and test-pack commit G. **5.1.10:**
  `docs/reference/lists.md`, generated, and **`bonsai --help --json`**, every word, flag, exit code and error word.
- 5.2: the salt (5.2.2), asks (`bonsai ask`, `answer`, `asks`; 5.2.5), and **the generated-files page,
  `docs/reference/generated-files.md`**, generated from the one Go table of generated kinds (5.2.6a), which "5.5 moves
  ... into the skill".
- 5.3: **Bonsai's own deny rules**, written by the engine, not a pack: `Edit(//**/.bonsai/local/**)`, one per
  generated table, and `Edit(~/.bonsai/**)` (5.3.5 note 2: "When 5.5 writes base's walls, those lines stay the
  engine's ... `base` keeps `Read(~/.bonsai/salt)`"); the guard's floor (5.3.2 rule 2: `.bonsai/STATE.md` stays free);
  the hook lines in Rohan's (a) form; (ii) held in `init`, `update` and `unlink`.
- 5.4: `bonsai ladder`; **the switch**: Bonsai's own `bonsai.yaml` (`packs: []`; documents under `records/`;
  `protected` and `person_only`), its tasks `T-5xyy` in `records/tasks/`, the landing rule (a piece lands on a green
  climb by `/usr/local/bin/bonsai` at its exact commit), the stricter-only rule for `bonsai.yaml`, one task with grants
  `running` at a time, `CLAUDE.md`'s four label lines "until base's block does", `STATE.md` at the root ("5.5 moves
  it, with its frontmatter"), Bonsai's real home `~/.bonsai/`, and "linking Bonsai's `base` pack in 5.5 adds no code
  that runs".

#### The pieces and their order

The spec's row (§14): "`base` and `workflow` from this repo's roles, protocols and templates; the always-on and skill
split, the roles' `skills:` preloads; `claude plugin validate --json` in each pack's CI, failing on every warning but
the missing `version`; the walls in base's deny rules and the studio's in `workflow`, each tried once on both sides
(question B, 1-2) (13-21); the documentation in every template and pack file of `base` and `workflow`, and each deny
rule's `why` (3-4); the pack template `packs/template/` with its CI and release, and Bonsai's CI job that runs it
(3-5)".

**Hours.** The template row's 3-5 is 5.5.1's. The first row's 13-21 is split over seven pieces by the planner's
judgment, for sizing briefs (5.5.0 2-3, 5.5.2 2-3, 5.5.3 2-3, 5.5.4 4-6, 5.5.5 1-2, 5.5.6 1-2 the spec's own figure,
5.5.7 1-2). The documentation row's 3-4 is not a piece: a pack file is documented as it is written (`check --pack` fails
otherwise, and `CLAUDE.md` asks a field change and its docs in one commit), so it is spread over the two packs (5.5.2 1,
5.5.4 2-3). In all: low 2+3+3+2+6+1+1+1 = 19; high 3+5+4+3+9+2+2+2 = 30 (pieces 5.5.0 to 5.5.7). The operating skill is
5.5.3, inside those hours. This section's planning and review runs count in 5.5's hours, carried in 5.5.0's run report
("What changes", item 3). Tasks: piece 5.5.y is `T-550y` (5.4's rule); `T-5590` onward, in order, for the work outside
the pieces: `CLAUDE.md`'s Safety line before 5.5.4 (order, item 2), the privacy check P, V1, the floors and the end
verification. P0's reads run under the task of the piece they read.

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.5.0 | **The engine's part for packs** (Go): a pack file's path may start with `<protocols>/` (the pack reader accepts it; today's refuses `<`), written into the project's `documents.protocols` folder and imported by the block; `init` writes `.bonsai/STATE.md` when there is none; the templates of the two files the engine writes (`bonsai.yaml`, STATE) keep their one home in the engine, and base's `workspace` and `state` skills are generated from them; the generated-files page moves into base's `generated-files` skill; formats set N (descriptions) | Go tests (a project whose protocols folder is not the default; a moved folder; each refusal); `init`'s STATE in a `t.TempDir()` project, never overwritten; `go generate` changes nothing and a changed template in a temporary copy fails the test; the formats test; P0's privacy read of STATE's template before it lands; check 10 and the ladder; CI | 2-3 | Spec §4 (`init`), §5 ("What loads always", "Every template and pack file documents itself"), §6 (`bonsai.yaml`, "Generated files", the block, "How `update` decides"); contract §2.2, §2.8, §7.2, §14; `formats/README.md` ("How the set changes"); this plan's 5.1.5, 5.1.9 and 5.2.6a notes; 5.1.5's run report, item 5 |
| 5.5.1 | **The pack template** `packs/template/` (one role, one skill, one documented template, `pack.yaml`, `labels.yaml`, README, `.gitattributes`), its checks in one script (`ci/check.sh`), its `pack.yml` (validate, `check --pack`, a scratch link, release on a `v*` tag), and Bonsai's CI running the script on every pack folder | `check --pack` and `claude plugin validate --json` on the template (only the missing `version`); the script green on both sides locally; the release subcommand run locally with made-up tag names (one unlike the version fails, the matching one passes; no tag, repository or release made); Bonsai's `packs` job and `windows` job green on CI; validate's need of a login measured; check 10 and the ladder | 3-5 | Spec §5 (the folder layout, "The pack template", pinning, "Adopting a release", mods), §12 step 8 (supply chain); gate report §2.13 (the eval's flags); the test pack's files; this plan's 5.1.1 rules 3-4 and 5.1.9 note |
| 5.5.2 | **`base`** in `packs/base/`: the plugin manifest; `pack.yaml` with the walls, each with its `why`, and base's protected paths; `labels.yaml`, contract §5.6's four; `block.md`; README; the template skills `task`, `run`, `memory` and `ci`; a Go test holding base free of code and its labels equal to the contract's | `check --pack` and validate; Bonsai's `packs` job; the Go test; a scratch project linked to base with `--yes` alone (nothing under "Runs code"), its block and deny rules read back; P0's privacy read of the task and run-report templates (after a Haiku grep) before it lands; the orchestrator's read; check 10 and the ladder | 3-4 | Spec §5 ("The two packs", template docs), §6 (the preview's sentences), §7 (the walls, deny rules), §10 (memory); contract §4, §5.6, §7.1, §7.2, §7.4; this plan's 5.3.5 note 2; the studio's three templates at `7017d63` (below), read only |
| 5.5.3 | **The "operating Bonsai" skill** in `base` (`skills/operating-bonsai/`), and a Go test that keeps it in step with the command words, flags and error words | The test (a skill line naming an unknown word or flag fails; a word the registry has and the skill never names fails); its size; its real sessions in 5.5.5 | 2-3 | Spec §3 (unattended, `error`), §4, §6, §9; contract §10.1-§10.2 (status moves in both modes); this section's "Moving a pack's version"; this plan's 5.1.6 and 5.1.10 notes, 5.2.5 note 1, 5.3.6 note 8, 5.4.2 note 12, 5.4.3 note 2 and 5.4.4 note 4; Rohan's 9 Oct direction (`STATE.md`) |
| 5.5.4 | **`workflow`**, its own repository, made from the template: five roles with their `skills:` preloads; `session-start` as the always-on file, five protocol skills; `lanes.yaml`; `labels.yaml`; eight document kinds with documented templates; the walls for the studio's secret files; block, README, LICENSE, its CI pinned to a Bonsai commit | `check --pack` and validate locally; a scratch project linked to it alone with `--yes` alone, and with `base`, its block under 40 lines; the privacy verifier P on every commit from the root before its first push; every later commit grepped and read before Rohan is told it can go public; its CI green on both sides after the push; Rohan's read | 6-9 | Spec §5 ("Where roles live", "What loads always", "The two packs"), §6 (the tables' closing step), §7 (the studio's secret files); contract §4.3, §5.4, §6, §7.1-§7.3, §9, §10.2, §11, §13; the studio's roles, protocols and templates at `7017d63` (below), read only |
| 5.5.5 | **The packs in real sessions**, by a Sonnet agent: a role's preloaded skills, a role as a subagent, the always-on file, the operating skill's four scenarios, `--agent workflow:builder --bg`, an interactive session through a terminal multiplexer; on Windows after Rohan's trust | Each session's transcript (stream JSON) and answers in the run report, with Claude Code's version and the user settings hashes before and after | 1-2 | Gate report §2.7, §5 (5.5); this plan's 5.1.7 note (trust); `design/plan.md` ("Test sessions and the launcher") |
| 5.5.6 | **The walls, tried once on both sides**, by a Sonnet agent: the secret files listed again on each side (names only), each rule's form tried on decoys, each real rule tried where no secret can be shown | The run report's table, one row per rule and side; the user settings hashes; V1 | 1-2 | Spec §7 ("The walls", "Deny rules the engine writes"); this plan's 5.3.5 notes 3-4 |
| 5.5.7 | **Bonsai links `base`**, and the rest of the switch's hand-offs: `bonsai.yaml` gains the pack and one protected path; the update; `STATE.md` to `.bonsai/STATE.md`; `CLAUDE.md`'s lines; the workflow repo fetched from GitHub with no login, on both sides | The preview read before `--yes` (no "Runs code"); `check` with no finding; a green climb at the link's commit; `status --json` showing the four labels in force; the end verifier's checks on the real repo | 1-2 | Spec §6, §14 step 6; this plan's 5.4 section ("The switch", "Tasks and names", "Labels with no definition in force until `base`") and 5.3's two answers |
| **5.5** | | | **19-30** (re-ask 39) | |

**The order, side by side where truly independent.** Rohan, 9 Oct: "if you can orchestrate work in parallel do that
whenever possible"; his 8 Oct bar stands: no shared file, and neither's proof resting on the other's. 5.5 starts once
5.4's end verifier has passed 5.4 (no two parts run at once). A piece that runs beside another rebases on `main` and
re-runs its proof if the other lands first. One rule from 5.4 shapes the order: a task whose `bonsai.allows` is not
empty runs with no other task `running`, so a protected path is changed only in such a window. The files each piece
owns:

| Piece | Owns |
|---|---|
| 5.5.0 | `internal/engine/` (the `<protocols>/` start, `init`'s STATE, the generators), `internal/workspace/pack.go` and `internal/workspace/paths.go` (the reader's check of a `files` path), the placeholder's Go table beside the pack's type in `internal/format/`, `formats/` (the pack schema's descriptions, README, manifest), `docs/reference/lists.md` (regenerated), `docs/reference/generated-files.md` (removed), `packs/base/skills/workspace/`, `packs/base/skills/state/`, `packs/base/skills/generated-files/` (generated whole), `.gitattributes`; the rule for `<protocols>/` where 5.1.9 put `check --pack` |
| 5.5.1 | `packs/template/` (all); its last commit: the `packs` job and one step of the `windows` job in `.github/workflows/ci.yml` |
| 5.5.2 | `packs/base/` but 5.5.0's three skill folders and `skills/operating-bonsai/`; `packs/packs_test.go` |
| 5.5.3 | `packs/base/skills/operating-bonsai/`; its line in base's `block.md` and its row in base's README; `cmd/bonsai/operating_test.go` |
| 5.5.4 | The `workflow` repository (every file), worked in `~/bonsai-checks/bonsai-workflow`; nothing in Bonsai's repo |
| 5.5.5, 5.5.6 | No repository file: scripts in `~/bonsai-checks/scripts/`; a finding goes back to the piece that owns the file, as a fix round |
| 5.5.7 | `bonsai.yaml`, `.bonsai/lock.json`, `.claude/settings.json` (by `update`), `CLAUDE.md`, `STATE.md` moved to `.bonsai/STATE.md`, the two tables (by `check --write`) |
| `T-5590` (the orchestrator, before 5.5.2 and 5.5.4 start) | `CLAUDE.md`'s Safety line only |

1. **5.5.0 and 5.5.1 start together.** 5.5.0 is engine code and generated skills; 5.5.1 is a new folder and a script.
   They share no file: the template holds no always-on file, so 5.5.1's checks do not rest on `<protocols>/`; the
   template's own `.gitattributes` sits inside its folder. 5.5.1's last commit, the CI jobs in `ci.yml` (a protected
   path), is made while it is the only task reading `running`: once 5.5.0 has landed, or with 5.5.0's task at `verify`.
   5.5.0 lands only after P0 has passed STATE's template.
2. **5.5.2 and 5.5.4 after both have landed, side by side.** Before either starts, with no other task `running`, the
   orchestrator changes `CLAUDE.md`'s Safety line under `T-5590` (granting `CLAUDE.md`): Bonsai's work now also writes
   the workflow pack's repository, `LastStep/bonsai-workflow` (its local copy in the scratch folder, then GitHub from P
   on), and lands it as a records commit before 5.5.4's builder starts. `base` needs 5.5.0's three generated skills
   beside its own and 5.5.1's CI job for its proof; `workflow` is made from the template (5.5.1) and its session-start
   file needs `<protocols>/` (5.5.0) in the Bonsai commit its CI pins. They live in different repositories, and
   neither's proof reads the other: `workflow`'s CI builds Bonsai at a commit that holds no `base`, and its scratch
   links prove it alone (the link with both packs is 5.5.5's and the end verifier's). 5.5.2 lands only after P0 has
   passed its two templates.
3. **5.5.3 after 5.5.2 has landed**, beside the rest of 5.5.4: it adds a skill to `base` and edits base's `block.md` and
   README, which 5.5.2 makes.
4. **P, the privacy check, once 5.5.4's builder has committed;** then the orchestrator creates the repository private
   and pushes it, and its CI runs on both sides. Rohan's read starts then and may run beside what follows; his changes
   are a fix round of 5.5.4, each commit grepped and read before it is pushed (P's paragraph). He gets the public line
   only once every commit since P has been read.
5. **5.5.5 and 5.5.6 after 5.5.3 has landed and `workflow`'s CI is green**, run by one Sonnet agent, WSL first, one
   after the other (two agents running Claude Code at once would confuse whose session changed a user settings file).
   Their Windows halves wait for Rohan's Windows sitting. Their scratch projects link both packs from local clones, so
   they need neither repository public.
6. **V1, the walls and the packs' CI**, after 5.5.6.
7. **5.5.7 after V1**: the link of Bonsai's repo; its last step, the fetch of `workflow` from GitHub with no login on
   both sides, waits for Rohan's line making it public. Then the floors (5.4's "once a part") and the end verifier.

**Who builds and verifies.** Opus builders for 5.5.0 (the engine), 5.5.1 (CI and release), 5.5.2 (the walls are
security), 5.5.3 (the skill every agent in a linked project leans on) and 5.5.4 (Rohan's content, rewritten with
judgment); a Sonnet agent for 5.5.5 and 5.5.6 (hand checks an agent can run, Rohan's 8 Oct word); the orchestrator links
Bonsai's repo in 5.5.7, with a Sonnet agent for its fetch on both sides. A Haiku agent greps for the privacy list,
reporting facts only: before each P0 (base's studio-derived templates, at 5.5.0 and 5.5.2), every new file of `base`
before 5.5.2 lands, `workflow`'s history before P, and each commit made after P. Fresh Opus verifiers, each re-running
what it judges itself: **P0**, P's rules on base's three studio-derived templates, before 5.5.0 and 5.5.2 land; **P**,
the privacy check, on `workflow`'s whole history before its first push (Rohan's content goes public; "Risk" in his
part); **V1** after 5.5.6, on the walls (security: what an agent may read on this machine), on 5.5.0's `<protocols>/`
resolution (where pack files land) and on the packs' CI and release (the template's `pack.yml`, `ci/check.sh`, Bonsai's
`packs` job, the workflow repo's CI); **the 5.5 end verifier** on "5.5 done", after 5.5.7. 5.5.0 to 5.5.3 land on the
landing rule (a green climb by the installed `bonsai` at the exact commit), check 10's Windows half, CI and the
orchestrator's read of the diff, which the run report says. 5.5.0 needs no verifier of its own: it touches no guard,
stop gate or ladder code (if it must, it gets one), and V1 break-tests its `<protocols>/` resolution, the one place it
changes where pack files land. It lands before V1 on its proof and the orchestrator's read; a must-fix V1 finds is fixed
forward.

#### Where each inherited finding is settled

The gate report's section 5, its 5.5 list; then the outline's "Settles", and what 5.1 to 5.4 and Rohan's 9 Oct decisions
hand on:

| Finding | Settled in | How |
|---|---|---|
| `/agents` is gone in 2.1.294; a role is read by starting a session as it, `--agent <plugin>:<role>` (gate §5, 5.5) | 5.5.4, 5.5.5 | `workflow`'s README and roles say how a role is started (`claude --agent workflow:builder`) and dispatched (the Agent tool's `workflow:builder`); 5.5.5 reads each route. Spec §17 step 6's text is outside step 5's parts (`STATE.md`); the orchestrator's dated note |
| The plugin carries no version; the version shows only as the commit folder; `validate` warns only of the missing `version` (gate §5) | 5.5.1, 5.5.2, 5.5.4 | Each pack's README says the version is the commit and `pack.yaml`'s `version` names tags only; the template's validate step fails on every warning but that one, matched on the version in use |
| Not seen: a pack's role answering in an interactive session (gate §5, §2.7) | 5.5.5 | On WSL, a real interactive session driven through `tmux` by the Sonnet agent: `claude --agent workflow:builder` in a scratch project, a question only the role's preloaded skill answers, the pane captured |
| Not seen: the Windows trust prompt (gate §5, §2.7) | Rohan's sitting | Through `claude-here.cmd` in a scratch project linked to both packs (his login first only if it has lapsed since 5.3): he answers the trust question; what it said goes in the run report, beside what his 5.3 sitting recorded |
| Not seen: a Windows `--bg` session; `claude --agent workflow:builder --bg`, check 8's original line (gate §5) | 5.5.5 | In that trusted Windows project, and on WSL: backgrounded, `claude logs` shows the role's answer, `claude stop` |
| Claude Code's own writes outside the scratch folders (gate §5) | 5.5.5, 5.5.6, 5.5.7 | Recorded again around each session (user settings hashes, the plugin tree's counts, the sync files); the accepted writes stay `design/plan.md`'s (Claude Code's session transcripts and its folder-trust entries for the scratch projects) and the plugin sync's own files; before 5.5.4 `CLAUDE.md`'s Safety line names `LastStep/bonsai-workflow` among the places Bonsai's work writes (order, item 2); 5.5.7 adds one accepted write, base's install for Bonsai's real repo in `~/.claude/plugins/`, named in the same line |
| `claude plugin eval`'s needs: `--no-publish`, `--trust-plugin`, a path target, `runsPerCase` miscounted (gate §5, §2.13) | 5.5.1 | The template's README gives the working command and its flags; evals run on this PC only, never in CI |
| Where `bonsai.yaml`'s and STATE's templates live (outline; this plan's "Stale or in tension": "5.5 settles the templates' one home") | 5.5.0 | In the engine, which writes both; base's `workspace` and `state` skills are generated from them (note 5.5.0, 3) |
| `bonsai.*` definitions "come with `base` in 5.5" (5.1.5) | 5.5.2, 5.5.7 | Base's `labels.yaml` holds contract §5.6's four, held equal by a test; in force in Bonsai's repo from the link |
| "A project's own `documents.protocols` folder and a pack's always-on files: 5.5" (5.1.5's run report, item 5) | 5.5.0 | A pack file's path may start with `<protocols>/` (note 5.5.0, 1) |
| Template defaults "`ladder_floor` and `ladder` empty, for 5.4 and 5.5" (5.1.5's run report, item 4) | 5.5.0 | Kept empty (note 5.5.0, 5); base's `workspace` skill shows a ladder as an example |
| "`claude plugin validate` is the pack's CI, 5.5's" (5.1.9) | 5.5.1 | In `ci/check.sh`, run by each pack's CI and by Bonsai's; its login need measured first (note 5.5.1, 5) |
| The generated-files page moves into base's skill, "its table generated from the same Go table"; the `generated:` comment names the page until then (5.2.6a) | 5.5.0, 5.5.7 | Generated as base's `generated-files` skill; the old page removed; `init`'s comment names the skill (5.5.0); Bonsai's own `bonsai.yaml` comment changed at the link (5.5.7) |
| "Base's walls deny reading [the salt]" (5.2.2 note 5) | 5.5.2 | `Read(~/.bonsai/salt)` and its twins (note 5.5.2, 2) |
| The ask flags "follow today's studio command, so the workflow pack's protocols change little" (5.2.5) | 5.5.4 | The protocols name `bonsai ask` where today's name the studio's way of asking |
| "Whoever commits on the main checkout runs `check --write` first; the workflow pack's closing protocol says so" (spec §6) | 5.5.4 | `session-start`'s closing section says so |
| Bonsai's own deny rules are the engine's, not base's; "`base` keeps `Read(~/.bonsai/salt)`" (5.3.5 note 2) | 5.5.2 | Base's walls leave out `.bonsai/local/`, the tables and `~/.bonsai/**`; a test fails if base repeats one of the engine's lines |
| "From 5.5 base's walls refuse an agent's tools [in the user's own settings file]" (5.3.6 note 3; Rohan's 5.3 option round) | 5.5.2, 5.5.6 | `Edit(~/.claude/settings.json)`, its Windows twin and the local settings rule; tried on both sides |
| Labels with no definition in force until `base`; `CLAUDE.md` gives the four in a line each (5.4, the switch) | 5.5.7 | The link puts base's definitions in force; the four lines leave `CLAUDE.md` (the block holds them); every task in `records/tasks/` read against them first |
| `STATE.md` at the root: "5.5 moves it, with its frontmatter" (5.4, the switch) | 5.5.7 | To `.bonsai/STATE.md` with `bonsai.state/1` frontmatter; `CLAUDE.md` names the new place |
| "Linking Bonsai's `base` pack in 5.5 adds no code that runs" (5.4, Rohan's part) | 5.5.2, 5.5.7 | Kept true: a test holds base free of code; the link's preview lists nothing under "Runs code" before `--yes` |
| The ladder's loop, "climb, read the `--json`, fix, climb again", named so 5.5 carries it (5.4.2 note 12) | 5.5.3 | A section of the operating skill; scenario S2 in 5.5.5 |
| The stricter-only rule, "which 5.5's `workflow` pack carries for other projects" (5.4.4 note 4) | 5.5.3, 5.5.4, 5.5.7 | For linked projects it is taught by the operating skill in `base`, which reaches every one; `workflow`'s ladder skill points to it (this section, "Stale or in tension"). It stands as 5.4 has it: moving a pack's version, adding a pack and taking one out wait for a person's word, an agent handing over the line ("Moving a pack's version", below, stated once) |
| "No operating skill until 5.5" (5.4 done, check 14) | 5.5.5 | Scenario S2 repeats check 14 with the skill and without naming `--help --json` |
| The floors at a part's end (5.4, "Tasks and names") | After 5.5.7 | As 5.4 has it: one climb of `main`, the floors raised to its counts under a `T-559x` task, the numbers in Rohan's last line |
| Agents manage Bonsai inside projects, "top notch" (Rohan, 9 Oct, 15:35) | 5.5.3, 5.5.5 | The operating skill, kept in step with `--help --json` by a test, tried in four real scenarios on WSL and in S1 on Windows |
| Rohan's (a) and (ii): hook lines name the installed `bonsai`; `--allow-exec`, `unlink` and changes to Bonsai's own lines are a person's where the studio does not manage (5.3) | 5.5.2, 5.5.3, 5.5.4, 5.5.7 | Neither pack carries code, so neither ever needs `--allow-exec`; the skill teaches the person's gates and why, and scenario S3 holds that an agent hands the line over |
| A pack plugin that runs code is asked for on each machine; Bonsai installs no plugin but the project's own packs' (Rohan, 9 Oct) | 5.5.1 | The template's README says what makes a pack run code (hooks, servers, monitors, a mod) and that each machine then asks; base and workflow carry none |
| No screens of Bonsai's own (Rohan, 9 Oct) | This section | Nothing in 5.5 draws a page; Rohan reads his pack on GitHub |
| No self-update; the program is his install per release (Rohan, 9 Oct, 15:46) | 5.5.3 | The skill hands a person the install lines and never installs `bonsai` |
| A pack's version stays a person's step; agents hand over the line (Rohan, 9 Oct, 18:32, approving this section) | 5.5.3, 5.5.5, 5.5.7 | Stated once below ("Moving a pack's version"); the operating skill teaches it; S3 and S4 try it; `CLAUDE.md` carries it from 5.5.7 |

#### Moving a pack's version: a person's step, stated once

Rohan's choice of 9 Oct (18:32), when he approved this section: "keep moving a pack's version as your step only; agents
hand you the line". So 5.4's stricter-only rule stands as it is (an agent makes only stricter changes to `bonsai.yaml`,
or one a section Rohan approved names; **"any other change ... loosens and waits for his word"**), and so does spec §5's
"Adopting a release stays a person's step". For Bonsai's own repo and every project linked to `base`:

- **A person's step:** a pack's `ref` changed (another commit or tag), its `source` or `path` changed, a pack added to a
  project already linked, and a pack taken out of `packs`. An agent never makes one of these on its own judgment,
  whatever the preview would show.
- **What the agent does instead: it prepares the change and hands it over, in one message:** what is waiting (`bonsai
  status --full --json` lists each pack's newer tags; before a pack's first tag, the commit the person or the pack's
  README names), the exact edit to `bonsai.yaml` (the line, old and new; for a pack added or taken out, its entry), and
  why. It changes nothing first: `bonsai update --json` previews only what `bonsai.yaml` already says, so a move's
  preview is read after the edit (next item).
- **On the person's word for that edit, and only then:** their yes to the line handed over (in the session, or as their
  answer to a `Decide` ask; where a studio manages the project, their grant through it, 5.3's (ii)), or a plan they
  approved that names it (in Bonsai's repo a section Rohan approved, as 5.4 has it: 5.5.7's `base`). Under a running
  task whose `bonsai.allows` grants `bonsai.yaml` and `.bonsai/lock.json`, in the main checkout on its base branch, the
  agent makes the edit, runs `bonsai update --json` and reads the preview line by line. A move (a `ref`, `source` or
  `path`) whose preview removes a wall or other deny rule, a label definition or a protected path (a changed one counts
  as removed) is put back and handed to the person again with what the preview listed, before any `--yes`; anything
  under "Runs code" needs consent to code, the person's under 5.3's (ii). Otherwise `--yes`, `bonsai check --json` with
  no finding, and one commit of everything the command wrote.
- **What agents still do unaided** (Rohan's 15:35 direction: agents link, update, fix, check, read status and edit):
  `bonsai update` to apply what is already decided (a repair, a re-apply on this machine, the project's own stricter
  change), stopping and handing over any preview that moves, adds or takes out a pack the person's word does not cover;
  `bonsai check` and its fixes; `bonsai status`; the stricter edits of `bonsai.yaml` that 5.4 has. A first link (`bonsai
  init`) stays the agent's too, at the packs and refs the person named (their request, the task or a plan they
  approved); a ref is a choice of version, so where none is named the agent hands the person the `--ref` it proposes
  (the newest tag, or before a pack's first tag the head commit of its default branch) and links on their yes.

Stated here once: the operating skill teaches it to every linked project (note 5.5.3, 2), `CLAUDE.md` carries it for
Bonsai's own work from 5.5.7, and 5.5.7 note 4 and `workflow`'s ladder skill point to it. No line of the spec or of 5.4
changes for it.

#### Notes per piece

**5.5.0, the engine's part for packs.**
1. **`<protocols>/`.** Spec §5: `session-start` is "a pack file the engine writes to the workspace's protocols folder,
   imported by the block". A pack cannot know a project's folder (the studio's is `studio/protocols`, `init`'s default
   `work/protocols`), so a `files` entry's `path` may start with `<protocols>/`: the engine writes the file under
   `bonsai.yaml`'s `documents.protocols`, and the block, which already imports the pack files under that folder, imports
   it. Rules: only as the first segment, and no other `<...>` anywhere (`check --pack` refuses one); the pack reader
   (`internal/workspace/pack.go`) accepts the placeholder there and checks the rest of the path with `CheckRelPath`,
   which keeps refusing `<` everywhere else; the resolved path passes every check a path passes today
   (`checkPackTarget`, `CheckRelPath`), so a `documents.protocols` of `.claude`, `.git`, `.bonsai`, `..` or a short
   name such as `CLAUDE~1` is refused, at a first link and at a folder change alike; the lock records the resolved
   path; a project whose `documents.protocols` is null or empty is refused with the word the engine uses for a bad
   `bonsai.yaml`, its `next.do` naming the line to set; a project that changes its folder gets, at its next `update`,
   the old file handled as one the pack no longer has (removed when unedited, a conflict when edited, as `plan.go` does
   today) and the new one written, so nothing new is built for the move. The angle bracket cannot start a real folder
   on Windows, so the start never reads as one. The placeholders are a Go table, one value now, on the reference page;
   `pack.schema.json`'s `files.path` description gains the rule (an addition). Chosen over a new `pack.yaml` field for
   always-on files (a new field in every reader and in `check --pack`, for one file) and over asking each project to use
   the pack's folder (the studio keeps `studio/protocols`). What lost: a pack names no other project folder this way;
   another placeholder is a later addition.
2. **STATE at `init`** (spec §4: "`STATE.md` from base's template when there is none, kind `once`"): `init` writes
   `.bonsai/STATE.md` when none is there, in format 1 (`format: bonsai.state/1` with its pointer comment, `updated`
   today, `updated_by: bonsai init`, `labels: {}`), its body a few headings, kind `once` in the lock; never over an
   existing file. `update` leaves a missing one missing (kind `once`, spec §6's table), so a project linked before 5.5
   adds one from `base:state` (the operating skill says so); `unlink` leaves it (5.1.7). The preview names it.
3. **One home for each template.** A template the engine writes (`bonsai.yaml`, STATE) lives in the engine, since a
   project may link without `base` (Bonsai's own did at 5.4); base's `workspace` and `state` skills are generated whole
   from it and from the schema by `go generate` (frontmatter, purpose, when to use, the fields table, the template as
   `init` writes it), and a test rebuilds them and fails on any difference, as the reference page's does. A template
   agents fill (task, run report, memory note, and `workflow`'s) lives in its skill, and `check --pack` holds its fields
   table to its frontmatter and to the schema. Chosen over base's skill as the one home with the engine reading it from
   the pack cache (an `init` without `base`, or offline, would have no template) and over two hand copies (drift).
4. **The generated-files page.** `go generate` writes `packs/base/skills/generated-files/SKILL.md` (a skill's
   frontmatter, then 5.2.6a's page) from the same Go table; `docs/reference/generated-files.md` and its `.gitattributes`
   line go; its test is retargeted; the comment `init` writes on `generated:` names `skill base:generated-files`; every
   code comment naming the old page is changed. The skill is read on demand, never always on (spec §6).
5. **Defaults kept:** `init`'s template keeps `packs: []`, `ladder: []` and `ladder_floor: []`. A project's rungs are
   its own (only it knows its test command), and a default rung 0 would make every named session's stop gate demand a
   climb in projects that chose none; base's `workspace` skill shows a two-rung ladder as an example, and the operating
   skill says how to add one (a stricter change). Before 1.0 there is no tag of `base` to name, so `init` names no pack;
   whether 1.0's template names `base-v1.0.0` is 5.7's.
6. **Formats set N:** the set after the last landed (`formats/README.md`, "How the set changes"), one commit with its
   manifest: the pack schema's `files.path` description and anything 5.5.0's start finds the schemas say about the
   templates' home; additions only, no new major. `.gitattributes` gains `packs/** text eol=lf`, so every pack file
   in Bonsai's repo is LF on a Windows checkout.

**5.5.1, the pack template, its checks and Bonsai's CI job.**
1. **The folder** (spec §5: "a minimal valid pack (one role, one skill, one documented template, `pack.yaml`,
   `labels.yaml`, `README.md`)"): `.claude-plugin/plugin.json` (`name: example-pack`, no `version`);
   `agents/example.md`; `skills/example/SKILL.md`; `skills/note/SKILL.md`, the template of a kind `note` the pack
   declares, with its fields table; `bonsai/pack.yaml` (every key commented; one document kind, one deny rule with its
   `why`, `files: []`, `hooks: []`); `bonsai/labels.yaml` (one label in the pack's namespace); `README.md`;
   `.gitattributes` (`* text=auto eol=lf`); `ci/check.sh`; `.github/workflows/pack.yml`. No always-on file and no hook,
   so the template carries no code and links with `--yes` alone; the README shows how to add each.
2. **The README** documents `plugin.json` field by field (JSON holds no comments, spec §5); how to start a pack (copy
   the folder into a new repository; change the id in `plugin.json`, `pack.yaml` and `labels.yaml`); that the version
   is the commit and `pack.yaml`'s `version` names tags only; what makes a pack run code (hooks, MCP or LSP servers,
   monitors, a mod, a file a hook runs) and that each machine then asks for `--allow-exec` (Rohan, 9 Oct); always-on
   files and `<protocols>/`; the CI and the release; evals run on the machine, with the gate's working command
   (`claude plugin eval <path> --no-publish --trust-plugin ...`, a path target, `runsPerCase` read with care).
   **How each kind of file documents itself** (spec §5), the same in `base` and `workflow`: a YAML file (`pack.yaml`,
   `labels.yaml`, `lanes.yaml`, `pack.yml`) a `#` comment on every key, with a short header; a template skill its
   fields table; a role file, and a skill that is no template (a protocol, the operating skill), an HTML comment just
   after its frontmatter, as the studio's `.claude/agents` copies carry one: a line or two (its purpose, when it is
   used, where its fields are documented), since the body is read into the session; `ci/check.sh` its header. The
   README documents the frontmatter fields of roles (`name`, `description`, `model`, `skills`) and of skills (`name`,
   `description`) once, each with its meaning, values and an example, and gives each role and skill file a row: each
   file's own docs, held equal to the folders by `ci/check.sh` (every file in `agents/` and `skills/` has its row, and
   every row names a file there). A list with a home elsewhere (the walls, lanes, labels, document kinds, the roles'
   preloads) is never copied into a README: it says where the list lives.
3. **`ci/check.sh`, the one home of a pack's checks**, so Bonsai's job runs exactly what each pack's CI runs (spec §5:
   "so the template never drifts from the engine"). Written for `sh` and always run as `sh ci/check.sh ...` (it runs on
   Ubuntu and in Git Bash on Windows; it is CI, not a hook line), so no step and no test needs its executable mode,
   which a Windows checkout does not keep. Its header documents every subcommand and argument: the `bonsai` to use, the
   pack folder, the source and commit to link, and for `release` the tag name. Steps: `bonsai check --pack <folder>
   --json` with no finding; a scratch project in a temporary folder (`git init`; `bonsai init --name ci --source
   <source> --path <folder> --ref <commit> --yes --json` with `claude` off the PATH, so the plugin step reports
   `skipped`); `bonsai check --json` there with no finding; then, when asked, `claude plugin validate --json <folder>`,
   failing on every warning and error but the missing `version`, matched on the field names the version in use prints
   (read and recorded by the builder). **The release check is its own subcommand**, `sh ci/check.sh release <tag>
   <folder>`: it takes the tag name as an argument, reads no git tag, and fails unless the name is `v` and `pack.yaml`'s
   `version`. So it is tried anywhere with a made-up name, and no agent makes a tag, a GitHub repository or a release to
   test it.
4. **`pack.yml`**, on every push and pull request, and on `v*` tags: a `check` job on `ubuntu-latest` and
   `windows-latest` (Bonsai cloned at a pinned 40-character commit, `BONSAI_COMMIT`, built with `CGO_ENABLED=0 go build`
   and Go from that commit's `go.mod`; then the script); a `validate` job on Ubuntu (Node and the pinned Claude Code,
   `CLAUDE_CODE_VERSION`; then the script's validate step); a `release` job: on a `v*` tag, after both, it runs `sh
   ci/check.sh release "$GITHUB_REF_NAME" .` and creates the GitHub release with its notes (`gh release create`,
   `contents: write` in that job only); on every other push it runs the same subcommand with the name the version needs
   (`v` and `pack.yaml`'s `version`) and stops before creating anything, so the release path runs on every commit. Every
   action pinned by its full commit with a version comment; `permissions: contents: read` at the top; no secret
   anywhere, so a fork's pull request runs the same checks; a `#` comment on every key, with a short header (spec §5:
   every pack file documents itself, in YAML a comment for every key). Before Bonsai's first release the pinned `bonsai`
   is built from a commit; from 1.0 the spec's archive with its SHA-256 (§5, §12 step 8) replaces the clone, a 5.7
   change to the template.
5. **Validate's login, measured first:** locally with an empty Claude configuration folder (a scratch
   `CLAUDE_CONFIG_DIR`, never the real one), then on the first CI run. If it needs a login, no secret is added: the
   `validate` job leaves `pack.yml`, the script's validate step runs on this PC (by the piece's builder before a pack's
   commit lands, and before each release tag), the run report records it, and the README says so. Rohan's approval of
   this section is his word for that fallback: a long-lived Claude login stored on GitHub, for a check that reads
   files, is not worth its exposure.
6. **Bonsai's CI** (the last commit, while 5.5.1 is the only task `running`): a `packs` job on Ubuntu builds `bonsai`
   from the commit, installs the pinned Claude Code and runs `sh packs/template/ci/check.sh` on each `packs/*/` folder
   holding `bonsai/pack.yaml`, the commit itself as the ref and the checkout as the source, so `base` is checked from
   the commit it lands in; the `windows` job gains one step running the script's `check --pack` and scratch link on each
   pack folder (no validate there). `CLAUDE.md`'s rule that a template's field change updates its docs needs nothing
   new: the job fails otherwise.
7. **Proof** beyond the table: `sh ci/check.sh release` run locally on the template with a made-up tag name unlike its
   `version` fails, and with the matching name passes (no tag, GitHub repository or release is made by any agent, here
   or later in 5.5); a copy of the template in a temporary folder with a `version` in `plugin.json`, an uncommented key
   or a deny rule without `why` each fails `check --pack` through the script.

**5.5.2, `base`.**
1. **Files.** `.claude-plugin/plugin.json` (`name: base`, a description, the author and repository; no `version`);
   `bonsai/pack.yaml`: `id: base`, `version: "0.1.0"` until 5.7 sets `1.0.0`, `needs.claude_code: "2.1.294"` (Bonsai's
   floor), `block: block.md`, `files: []`, `hooks: []`, `deny:` the walls, `documents: []` (Bonsai's own kinds are
   built in), `protected: ["bonsai.yaml", ".bonsai/lock.json"]` (the schema: "Base declares `bonsai.yaml` and the
   lock"); `bonsai/labels.yaml`, contract §5.6's four with namespace `bonsai`, each key commented; no `lanes.yaml` and
   no `agents/` ("No roles, no lanes"); `bonsai/block.md`, at most four lines, naming base's skills (the operating skill
   from 5.5.3, the templates, the generated-files page), its docs in an HTML comment the engine leaves out; `README.md`
   (what base is, `plugin.json`'s fields, a row per skill, where the walls live and what they are for, pointing at
   `bonsai/pack.yaml`'s `deny`, where each rule's `why` sits, never a copy of the list; how a project takes it; that it
   carries no code). Every file documents itself as note 5.5.1, 2 says. Its files are LF through the root's
   `.gitattributes` line (5.5.0).
2. **The walls** (spec §7), each a deny rule with its `why` (the preview's sentence), in `pack.yaml`, their one home.
   **What is walled, one rule for all:** Read, the files and folders that commonly hold keys, tokens or logins; Edit,
   the person's own Claude Code files and the files whose change starts code. An exact file where one is enough, a
   folder only where the tool keeps several. `base` is public and serves projects beyond this PC, so the rule, not
   this PC's listing, decides. **Each wall has its twins:** `~/` is the session's own home (on Windows the Windows
   home); from WSL the Windows home is `//mnt/c/Users/*/`; from Windows the WSL home is its network path (below).
   - Read, secrets: `~/.ssh/**`, `~/.aws/**`, `~/.config/gh/hosts.yml`, `~/.claude/.credentials.json`,
     `~/.docker/config.json`, `~/.git-credentials`, `~/.bonsai/salt` (the spec's); `~/.netrc`, `~/.npmrc`, `~/.pypirc`,
     `~/.kube/config`, `~/.config/git/credentials` (git's other credential file), `~/.gnupg/**`, `~/.config/gcloud/**`,
     `~/.azure/**` (added, by the rule); `~/.claude.json` (added: Claude Code's global file, whose user and local MCP
     servers' `env` values often hold tokens); on Windows, where the same tools keep them elsewhere,
     `~/AppData/Roaming/GitHub CLI/hosts.yml`, `~/AppData/Roaming/gcloud/**` and `~/AppData/Roaming/gnupg/**` (added).
   - Read, the Windows home from WSL: `//mnt/c/Users/*/` with `.ssh/**`, `.git-credentials`,
     `.claude/.credentials.json`, `.docker/config.json` (the spec's); and with every other Read wall above that has a
     place in a Windows home: `.aws/**`, `.azure/**`, `.claude.json`, `.netrc`, `.npmrc`, `.pypirc`, `.kube/config`,
     `.bonsai/salt`, `AppData/Roaming/GitHub CLI/hosts.yml`, `AppData/Roaming/gcloud/**`, `AppData/Roaming/gnupg/**`
     (added: spec §7 says "the Windows side is short" and asks 5.5 to list it again).
   - Read, the WSL home from Windows: a native Windows session reaches WSL's files through
     `\\wsl.localhost\<distro>\home\<user>\` and `\\wsl$\<distro>\home\<user>\`, the twin of the `//mnt/c/Users/*/`
     rules. Each `~/` Read wall's twin there, in the form 5.5.6 finds holds; the forms to try first:
     `//wsl.localhost/*/home/*/.ssh/**` and `//wsl$/*/home/*/.ssh/**` (and the same with a third leading slash). If
     no form holds on Windows, base's README says so plainly: a native Windows session can read WSL's key files by that
     path, and the walls do not stop it.
   - Edit, the person's own Claude Code: `//**/.claude/settings.local.json`, `~/.claude/settings.json` (the spec's);
     `~/.claude.json` (added: its MCP servers are code Claude Code starts on its own); `~/.claude/CLAUDE.md` (added: the
     user memory file every project loads; 5.6's import line in it is the person's, "No agent edits that file"); from
     WSL, `//mnt/c/Users/*/` with `.claude/settings.json`, `.claude.json` and `.claude/CLAUDE.md` (added: the Windows
     user's own files).
   - Edit, files that start code: `~/go/bin/**`, `~/.local/bin/bonsai*` (the spec's: a copy that would shadow the
     installed `bonsai`); `~/.ssh/**` (added: `config` can run a command at every connection and `authorized_keys` lets
     someone in; a Read wall stops a read, not the write of a new file); `~/.gitconfig` and `~/.config/git/config`
     (added: aliases, a hooks path and helpers run commands in every repository); `~/.bashrc`, `~/.bash_profile`,
     `~/.profile`, `~/.zshrc`, `~/.zprofile` (added: run at every shell start); on Windows `~/Documents/PowerShell/**`
     and `~/Documents/WindowsPowerShell/**` (added: PowerShell's profiles and modules, run at its start); from WSL,
     `//mnt/c/Users/*/` with `.ssh/**`, `.gitconfig`, `Documents/PowerShell/**` and `Documents/WindowsPowerShell/**`.
     A place left out is written in the README with its reason (a Documents folder moved elsewhere, say, by a sync
     tool, which a rule on `~/Documents/` does not reach).
   - **Left to the engine** (5.3.5 note 2): `Edit(//**/.bonsai/local/**)`, the tables' rules and `Edit(~/.bonsai/**)`;
     `packs/packs_test.go` fails if base repeats one. Narrow on purpose, as spec §7 says (exact files where a folder is
     not needed; no `Bash(...)` rule). The final forms follow 5.5.6's tries: a form that does not hold on one side is
     replaced by one that does, or documented in the README as one side's, with the reason. Every folder rule is a
     recursive glob, which spec §7 notes the sandbox makes into one bind per file (issue #74081): the few folders here
     are for the sandbox probe to measure.
3. **The template skills** `task`, `run`, `memory`: each `skills/<kind>/SKILL.md` with a `name` and a `description`
   saying when to use it; the purpose; the fields table (`| Field | Meaning | Allowed values | Example |`), equal to the
   schema for its format; then the template, its `format:` line carrying the pointer comment (spec §5). Bodies: the
   task's from today's studio template (`studio/templates/task.md`: Context, Scope, Out of scope, Notes), its fields
   contract §4.1's with `labels:` showing the four `bonsai.*` labels; the run report's from
   `studio/templates/run-report.md` (What is different now, Deviations from the plan, Ladder, What did not work,
   Evidence for a human, Files changed, Processes left running, Left for next time, Verifier), its frontmatter contract
   §7.1's (no `ladder_result`, no `cost_usd`; `needs-person`); the memory note's from `bonsai.memory/1` (spec §10: the
   fact, why, how to apply; the index line; 4 KB a note, 120 lines and 12 KB the index; never a secret). STATE's
   template is 5.5.0's. The studio's wording is rewritten by 5.5.4 note 3's rules (no studio path, no `<mimas>`, no
   studio task id), and `base` names no person ("a person", never Rohan) and carries no studio ritual. P0 reads these
   two, and STATE's, before their pieces land (P0's paragraph).
4. **The `ci` skill** (spec §5's table: "a CI workflow template (kind `once`)"): a GitHub Actions workflow a project
   copies to `.github/workflows/bonsai.yml`, every key commented, running `bonsai check --json` with no pack fetched
   (spec §5: "CI needs no pack"), its `bonsai` built from a pinned Bonsai commit before 1.0 and taken from the release
   archive checked by its SHA-256 after (§12 step 8). **A skill, not a file `init` writes**: a workflow file runs on
   GitHub with the repository's token, so writing it at every link would put code that runs into every project without a
   word from anyone, outside what `--allow-exec` asks about (it counts what runs on the machine); a pinned version in a
   `once` file is never updated by the pack; and Bonsai's own repo, which links `base`, has its own CI under
   `.github/**`, a protected path. What lost: a project gets its CI when an agent or a person adds it (the operating
   skill says how), not at the link.
5. **Free of code, held by a test** (`packs/packs_test.go`, Go, reading the folder): `hooks: []`, `files: []`; no
   `hooks/`, `.mcp.json`, `.lsp.json`, monitors or mod file; each skill folder holds `SKILL.md` and nothing else (no
   script a skill could run); `labels.yaml` equal to contract §5.6's four, field by field; none of the engine's own
   deny rules repeated. So 5.4's "linking base adds no code that runs" stays true while base grows.

**5.5.3, the "operating Bonsai" skill** (`packs/base/skills/operating-bonsai/SKILL.md`, loaded as
`base:operating-bonsai`; Rohan, 9 Oct, 15:35; the outline's "Agents first").
1. **Its description** says when to load it: linking, updating, a `bonsai check` finding, reading status, editing
   `bonsai.yaml`, proving a task with the ladder, asking a person, or any refusal from Bonsai or its guard. Base's
   `block.md` names it in one line, so every session in a project linked to base sees that it exists (about 20 tokens);
   its text loads only when opened.
2. **What it teaches**, in plain steps with exact command lines, never a copy of a list that has its home elsewhere:
   - **Learn the tool:** `bonsai --help --json` (5.1.10) lists every word, flag, exit code and error word with its usual
     `who`; `bonsai check --schema <format>` prints a format. Read those instead of guessing.
   - **Every answer is `--json`:** on a refusal read `error.code` and `error.next`: when `next.who` is `agent`, run
     `next.do` as written; when it is `person`, stop and hand the person `next.do` exactly, with one line on why; never
     find a way round it.
   - **Where things stand:** `bonsai status --json` (`problems`, `needs`, `active_task`, `ladder`, `documents`,
     `labels`), `--active`, `--full`.
   - **Check and fix:** `bonsai check --json`; for each finding its `next.do` by its `who`; check again until no
     finding; warnings never block; `bonsai check --write` only in the main checkout, before a commit there.
   - **Link and update, only in the main checkout on its base branch:** `init` and `update` write
     `.claude/settings.json`, which rung 0 refuses on any other branch (5.4.3 note 2). A first link with `bonsai init
     --json` (a preview; nothing written; exit 4), each settings line read, then `--yes`; a pack named by `--source`,
     `--path` and `--ref`, or in `bonsai.yaml` before `init`, at the packs and refs the person named (where no ref is
     named, the agent hands the person the one it proposes and links on their yes). An update applies what is already
     decided (a repair, a re-apply on this machine, the project's own stricter change): `bonsai update --json`, its
     preview read line by line, then `--yes`; a pack's version moved, a pack added or a pack taken out is the person's
     step (next item). Exit 5 (a conflict): keep (`--keep P`) when the project's edit should win, adopt (`--adopt P`)
     when the pack's should, and ask when unsure. Exit 4 with "Runs code": consent to code (`--allow-exec`) is never the
     agent's own choice; where a studio manages the project the command accepts it under the person's grant (the task's
     grant of `bonsai.yaml` and the lock came from the person); elsewhere the agent stops and hands the person the exact
     line (5.3's (ii)). `waiting` (trust): a person opens a session in the folder and trusts it, then `bonsai update`.
     Then `bonsai check --json` with no finding, and everything the command wrote (`bonsai.yaml`, the lock,
     `.claude/settings.json`, the block, the pack's files, the tables) in one commit.
   - **Edit `bonsai.yaml`:** it is person-only; an agent edits it only under a running task whose `bonsai.allows` grants
     it, and on its own only for a stricter change (a rung or a floor entry added, a floor raised to a green, clean
     climb's count, a protected path added). **Anything else loosens a check and waits for a person's word:** a floor
     lowered; a rung's command, `required`, tests, ratchet or capture changed or removed; a protected or person-only
     path taken out; a test marked to skip in CI. **A pack's version is the person's step too, whatever the preview
     would show:** a pack's `ref`, `source` or `path` changed, a pack added to a project already linked, a pack taken
     out of `packs`. For these the agent reads what is waiting (`bonsai status --full --json`, each pack's newer tags),
     hands the person the exact edit (the line, old and new) and why in one message, and changes nothing; on their yes
     it makes the edit, reads `bonsai update --json`'s preview line by line, hands back a move that removes a wall, deny
     rule, label definition or protected path, and otherwise runs `--yes`. Why: an agent loosening the checks it is
     held to, or choosing the packs that hold it, would be judging its own work. `bonsai check --json` after every
     edit. (This section's "Moving a pack's version" is the rule's text; 5.4's switch the rest.)
   - **Status moves:** `bonsai status --json` gives `status_writes`. Under `agents`, an agent edits a task's status line
     itself, by the moves contract §10.2 gives agents. Under `command` (a project a studio manages), the guard refuses
     an agent's edit of a status line, a lane line, a task file outside the main checkout, and a person-only path put
     into `bonsai.allows`; the agent runs the command `status --json` names in `status_command` instead, as the refusal
     says (contract §10.1).
   - **Prove a task:** `bonsai ladder --task <id> --json`: read the climb's `next`, run each failing test alone with its
     `run` line, fix the code (never the test, the ladder or `bonsai.yaml` to get green), climb again; a green local
     result at HEAD is the proof the stop gate reads; a `--ci` result proves nothing. The proof that must be trusted is
     made by the installed copy the hook lines name (on Linux `/usr/local/bin/bonsai`, on Windows
     `C:\Program Files\Bonsai\bonsai.exe`).
   - **Ask a person:** `bonsai ask --type Answer|Decide|Look|Play ... --json`, then `bonsai ask --status <key> --json`;
     never answer one's own ask (`bonsai answer` is a person's).
   - **The guard:** a refused edit or command says why first; a protected path needs the running task's grant; never
     write it another way (a shell redirect, a script); never set `disableAllHooks`; a session where every call is
     refused because `bonsai` is missing stops and tells the person (the install is theirs). `bonsai hook guard`, `hook
     start`, `hook stop` and `hook record` are Claude Code's, run from the settings lines on each call: an agent never
     runs them by hand.
   - **The person's gates, and why:** consent to code (`--allow-exec`), `bonsai unlink`, a change to Bonsai's own hook
     lines (each lets new code run on the person's machine or takes the guard out; where a studio manages the project,
     the person's grant through it; elsewhere the person types it), every loosening of `bonsai.yaml`, every pack's
     version moved and every pack added or taken out (above), person-only files under `command` mode, Bless where a
     studio manages the project (elsewhere an agent raises a floor itself, by the rule above), and installing the
     `bonsai` program (root or admin; the agent sends the person the lines and the fingerprint check, never installs).
   - **Records:** `bonsai logs --json`; `.bonsai/STATE.md` (add one from `base:state` if missing); the tables are
     rebuilt, never edited; how long things are kept: `base:generated-files`; templates: `base:task`, `base:run`,
     `base:memory`, `base:workspace`, `base:state`, `base:ci`.
3. **Size:** at most 10 KB, so an agent reads it whole; detail lives in `--help --json` and the template skills.
4. **Kept in step** (`cmd/bonsai/operating_test.go`, in package `main`, where the word registry lives): every `bonsai
   ...` line in a code block parses against the registry (the word exists; every flag is the word's); every error word
   the skill names is in `format.ErrorWords`; every command word the registry has is named at least once, the `hook`
   words among them (the skill names them as never run by hand; the test leaves no word out), so a new word fails the
   test until the skill teaches it. `CLAUDE.md` gains the rule at 5.5.7 (a command word, flag or error word changes the
   skill in the same commit). Base is pinned per project, so a project's skill describes the `bonsai` of its base
   commit; it says so, and sends the agent to `--help --json` for the copy it runs.
5. **Tried in real sessions** in 5.5.5: scenarios S1 to S4 on WSL, S1 on Windows.

**5.5.4, `workflow`.**
1. **The repository.** `LastStep/bonsai-workflow`, made from `packs/template/` at the commit 5.5.1 landed (copied,
   renamed `workflow`), worked in a new local repository, `~/bonsai-checks/bonsai-workflow`, on `main`, in small
   commits; the builder never pushes and nothing is pushed before P. The studio's files are read only, with
   `git show 7017d63:<path>` ("What changes", item 10); nothing is written or run in the studio's checkout.
2. **What moves, what is rewritten, what is left out** (the studio's repo at `7017d63`, 28 files read):

   | Studio file | Into `workflow` | Loads |
   |---|---|---|
   | `studio/roles/builder.md` | `agents/builder.md` | as `workflow:builder`; preloads `lanes`, `verification-ladder`, `reward-hacking-guards`, `reporting` |
   | `studio/roles/verifier.md` | `agents/verifier.md` | preloads the same four |
   | `studio/roles/researcher.md` | `agents/researcher.md` | preloads `lanes`, `reporting`, `adr` |
   | `studio/roles/producer.md` | `agents/producer.md` | preloads `lanes` |
   | `studio/roles/playtest-analyst.md` | `agents/playtest-analyst.md` | no preload |
   | `studio/protocols/session-start.md` | `bonsai/files/session-start.md`, written to `<protocols>/session-start.md` (kind `pack`) | always, imported by the block |
   | `studio/protocols/lanes.md`, `verification-ladder.md`, `reward-hacking-guards.md`, `reporting.md`, `adr.md` | `skills/<name>/SKILL.md` | on demand, or preloaded as above |
   | `studio/templates/plan.md`, `one-pager.md`, `adr.md`, `options.md`, `playtest.md`, `brief.md` | `skills/plan/`, `skills/feature/`, `skills/decision/`, `skills/options/`, `skills/playtest/`, `skills/brief/` | on demand |
   | `studio/templates/task.md`, `run-report.md`, `state.md` | `base` (5.5.2; STATE 5.5.0) | |
   | `studio/protocols/asset-safety.md`, `unity-live-editor.md` | left out: a later game pack (spec §5's table) | |
   | `studio/templates/game.yaml` | left out: `base`'s `bonsai.yaml` template replaces it | |
   | `.claude/agents/*.md` (5) | left out: copies of the roles, spelled for the studio's repo | |

   New: `.claude-plugin/plugin.json`, `bonsai/pack.yaml`, `bonsai/lanes.yaml`, `bonsai/labels.yaml`, `bonsai/block.md`,
   `README.md`, `LICENSE`, `.gitattributes`, `ci/check.sh` and `.github/workflows/pack.yml` (from the template),
   `skills/spec/SKILL.md` and `skills/bugs/SKILL.md` (12 files). The preloads are spec §5's table; where today's
   `applies_to` names more (the producer reads `verification-ladder`, `reward-hacking-guards` and `reporting`; the
   playtest analyst `reporting`), the role's body names those skills to read when needed, so today's behaviour holds
   without loading more into every start.
3. **The rewrite rules**, the same for `base`'s three templates (which also name no person). What the grep found in the
   20 files that move (the 17 above and the three for `base`): studio folder paths on 20 lines in 8 files and
   `<studio>`, `<trinetra>` on 7 more; the studio's task variable once; its project file (`game.yaml`) on 4 lines; the
   ledger on 10 lines; Discord and email on 4; "golden rule 8" on 3; the bridge once; the studio's name on 4 lines;
   Mimas on 12 lines, mostly the `<mimas>` placeholder; Unity on 2; one example task id; `cost_usd` and `allows_assets`
   twice each; Rohan's name on 54 lines; no home folder, machine or tailnet name, email address, token or address. The
   section's review found more, each checked at `7017d63`: the paths `docs/design/index.html#anchor` (the one-pager
   template's `design_anchor`), `roadmap/<M2>.md` (STATE's template), `artifacts/...png` (the run report's) and the
   decision protocol's "Mimas: `docs/decisions.md`"; the frontmatter values `projects: [mimas, trinetra]` (the brief's),
   `scope: <studio | mimas>` (the decision record's) and `project: <mimas | trinetra | studio>` (the options
   write-up's); the lines `node <trinetra>/tools/ladder/ladder.mjs --project mimas` (the ladder protocol), `systemctl
   --user list-units` for the bridge's unit (session-start's section 7) and the producer's "Runs on a schedule"; and
   Rohan's weekly rituals and the daily brief's hour (the producer's role, the reporting protocol, STATE's template).
   So:
   - A studio path becomes the document kind it means ("the plan, in the folder `bonsai status --json` names for
     `plan`"); `studio/STATE.md` becomes `.bonsai/STATE.md`; ladder results `.bonsai/local/ladder/<task>.json`
     (contract §11), `ci.json` stamped `mode: ci` for a CI run. A project's own places become what they hold, with no
     example path: "the design section the one-pager cites", "the milestone's roadmap, where the project keeps one",
     "screenshots, where the project keeps them", "the project's own decision log".
   - The studio's tools become Bonsai's: its ladder script `bonsai ladder --task <id> --json`; its task variable
     `BONSAI_TASK` or `bonsai status --active --json` (contract §13); `allows_assets` the task's `bonsai.allows`;
     `game.yaml`'s rungs, protected list and ratchets `bonsai.yaml`'s; `cost_usd` a body line (a cost label is the
     studio's own to define, contract §7.1; `workflow` names none); asking Rohan, `bonsai ask` (contract §9).
   - The studio's services leave: Discord, email, the bridge's unit (with its `systemctl` line), the Desk, the registry
     (notifications and screens are the studio's; the producer writes the brief and the studio delivers it). The
     producer's "Runs on a schedule" says who starts it: a studio's scheduler or a person (Bonsai has none).
   - "Golden rule 8" is written out where it is cited: a fresh verifier only for big or risky work (guards, the
     ladder's proof logic, tokens and security, deploys, a public contract, the full and director lanes); related light
     tasks share one; the rest close on a green ladder and the orchestrator's read of the diff.
   - The ledger stays, worded "where the project keeps a ledger of `done_when` entries": the studio and Mimas keep
     theirs, and a project without one loses nothing.
   - `project:` and `projects:` fields and `<mimas>` placeholders go (a project is its repository); a field whose
     values name the studio's projects (the decision record's `scope:`) keeps its meaning with values that name none;
     field values that name Rohan in a list a reader parses become `person` (as contract §7.1's `needs-person`). In
     `workflow`, prose keeps his name and game wording stays; `base`'s templates name no person ("a person") and carry
     no studio ritual. His weekly rituals and the brief's hour stay in `workflow`, marked in the builder's list for his
     read (note 10).
   - **Status moves, in session-start's section 7** ("Before you stop, set the task's status"): for both modes, as the
     operating skill says (note 5.5.3, 2): where `bonsai status --json` reads `status_writes: command`, the move goes
     through its `status_command` (the guard refuses a status-line edit there); otherwise the agent edits the status
     line, by the moves contract §10.2 gives agents. Its `allows_assets` paragraph becomes the task's `bonsai.allows`,
     which grants only while the task reads `running`.
   - An example id becomes a made-up one; a frontmatter line gains its pointer comment (spec §5); a template's fields
     move into its fields table.
   - The always-on file stays at or under today's size (4.6 KB), its own docs one line, the rest in the README (spec
     §5: what loads in every session carries one pointer line).
4. **`skills:` and how a role names a skill** (spec §5: "checked in step 5.5"): at 5.5.4's start the orchestrator has a
   Sonnet agent try both forms (`lanes` and `workflow:lanes`) on a fixture role in a scratch session, as a subagent and
   as an `--agent` session (a builder runs as a subagent and starts no agent of its own); the builder writes the form
   that loads, and the run report records which. A role's body also names the skills it reads, so a route that ignores
   `skills:` still gets them.
5. **Declarations.** `bonsai/lanes.yaml`: contract §6's three lanes, the descriptions from today's `lanes.md`.
   `bonsai/labels.yaml`, namespace `workflow`: `workflow.owner` (choice: the five roles; tasks), `workflow.model`
   (choice: `opus`, `sonnet`, `haiku`, `fable`; tasks and run reports; `haiku` added to contract §5.4's list for
   Rohan's 8 Oct rule), `workflow.feature` (text, `^F-[a-z0-9-]+$`; tasks and plans), `workflow.milestone` (text;
   tasks, STATE, one-pagers), `workflow.spec` (text, an id; tasks), `workflow.estimate_h` (number; tasks; spec §16 row
   26). `pack.yaml`'s `documents`, eight kinds, each with today's id rule and statuses, its moves by contract §10.2 (a
   person approves; agents make the rest) and a default place under `work/`, as `init`'s other defaults:

   | Kind | Default place | Id | Statuses |
   |---|---|---|---|
   | `plan` | `work/plans` | `^P-T-[0-9]{4,6}$` | `draft`, `approved`, `rejected`, `superseded`; stamp `approved`; `task_field: task` |
   | `feature` (one-pagers) | `work/features` | `^F-[a-z0-9-]+$` | `draft`, `approved`, `building`, `verified`, `done`, `cut` |
   | `decision` (records) | `work/decisions` | `^ADR-[0-9]{4}$` | `proposed`, `accepted`, `superseded` |
   | `options` | `work/options` | `^OPT-[0-9]{4}$` | `open`, `decided`, `deferred` |
   | `spec` | `work/specs` | `^SPEC-[0-9]{4}-[0-9]{2}-[0-9]{2}-[a-z0-9-]+$` | `draft`, `approved` |
   | `playtest` | `work/playtests` | none (file `<date>-<who>.md`) | none |
   | `brief` | `work/briefs` | none (file `<date>.md`) | none |
   | `bugs` | the file `work/bugs.md` | none | none |

   Options write-ups get a kind and a folder of their own: today they share `studio/decisions/` with the records, and
   one kind has one list of statuses. Whether Bonsai's readers accept two kinds in one folder (the studio may point
   both at `studio/decisions`) is tried in a scratch project and written in the README either way. The `spec` and
   `bugs` templates are new and short, from the shape of today's specs' frontmatter and of the bugs register's header,
   none of their content.
6. **The walls for the studio's secret files** (spec §7: "base never names the studio"): `Read(~/.trinetra/token)`,
   `Read(~/.trinetra/deploy_ed25519)`, `Read(~/.trinetra/salt)`, and their twins by base's rule (note 5.5.2, 2): the
   same three under the Windows home from WSL (`//mnt/c/Users/*/.trinetra/...`) and under the WSL home from Windows, in
   the form 5.5.6 finds holds; 5.5.6's listing may add one there (names only). Each with its `why`.
7. **`block.md`**, at most four lines: the roles and how to start one (`claude --agent workflow:builder`), the lanes
   skill, the document templates by name. With `base`'s block, the label lines and the imports, the whole block stays
   under 40 lines (measured on a project linked to both).
8. **The README**: what the pack is and where it came from (Rohan's studio's roles, protocols and templates, 9 Oct
   2026, rewritten for Bonsai's formats); how a project takes it (its `bonsai.yaml` lines); the always-on file; a row
   per role and skill file (note 5.5.1, 2); where each list lives, pointing at its file and never copying it: each
   role's preloads on its own `skills:` line, the lanes in `bonsai/lanes.yaml`, the labels in `bonsai/labels.yaml`, the
   document kinds in `pack.yaml`'s `documents`, the walls in its `deny`; `plugin.json`'s fields; the version is the
   commit; its checks and release; that it carries no code. Every file documents itself as note 5.5.1, 2 says.
   `LICENSE`: MIT, the same holder as Bonsai's.
9. **CI pinned** to a Bonsai commit that holds 5.5.0 and 5.5.1, and to the Claude Code version 5.5.1 measured.
10. **For Rohan:** the builder's report ends with a list, file by file, of what changed from the studio's version and
    why, in plain words, with the kept lines that describe his week (the rituals, the brief's hour) marked for his
    read: the orchestrator sends it with the repository's link.
11. **Proof** beyond the table: `sh ci/check.sh` on both sides locally (on Windows, a Windows-git clone under
    `%USERPROFILE%\bonsai-checks\` and Bonsai built with Windows Go), validate locally; a scratch project linked to
    `workflow` alone with `--yes` alone writes `work/protocols/session-start.md` and the block's import of it; one
    whose `documents.protocols` is `docs/agents` gets it there.

**P0, P's rules run early, on `base`'s three templates from the studio.** `base` lands on Bonsai's public `main`
before `workflow`'s first push, so its studio-derived templates are read before they land: a fresh Opus agent reads, by
P's rules below, STATE's template before 5.5.0 lands (the engine's template and base's generated `state` skill), and
the task's and the run report's before 5.5.2 lands (base's `task` and `run` skills): each file in full against its
studio original, every commit on the piece's branch that touches it (`git log -p main..<branch> -- <files>`) with its
message, and its own grep; also that each names no person ("a person", never Rohan) and carries no studio ritual. A
Haiku grep runs first and reports facts only. P0 passes or fails; it fixes nothing; the piece lands only after a pass.
Its runs count in that piece's hours and run under its task.

**P, the privacy check.** A fresh Opus agent, before the first push, reads what a public repository publishes: its
whole history, not only its last tree. It reads every commit's diff from the root (`git log -p --root`) and every commit
message, then every file of `workflow` at the builder's last commit in full, and base's three studio-derived templates
as they landed; it compares each file with its studio original (`git show 7017d63:<path>`, read only) against the
builder's list; it runs its own grep over the history and the tree (home folders, `C:\Users\` with a name, machine and
tailnet names, `ts.net`, email addresses, token and key shapes, studio task and bug ids, `studio/`, `<studio>`,
`tools/`, the studio's services, Mimas's paths; the walls' three file names in `~/.trinetra/` are the one studio name
this section allows); and it checks that the commits' author line is the one Bonsai's public commits already carry. A
Haiku grep runs first and reports facts only. P passes or fails; it fixes nothing. Then the orchestrator creates the
repository private (`gh repo create LastStep/bonsai-workflow --private`), pushes `main`, reads CI, and sends Rohan the
link and the list.

**Every commit after P**, a fix round's and those made on Rohan's change requests alike, gets a Haiku grep (facts
only) and the orchestrator's read of its diff and message before it is pushed. The orchestrator tells Rohan the repo is
ready to go public, with his line, only when every commit since P has had both, and the run report lists each commit
with its grep and read. The end verifier's check 8 reads the whole history again.

**5.5.5, the packs in real sessions** (a Sonnet agent; spec §5; gate report §2.7).
1. **The scratch projects:** under `~/bonsai-checks/packs/` and `%USERPROFILE%\bonsai-checks\packs\`, linked by the
   stamped test build (confined to its scratch root, 5.3.6) on a scratch `BONSAI_HOME` to `base` (a local clone of
   Bonsai at the landed commit, `--path packs/base`) and `workflow` (its local repository at the pushed commit);
   Claude Code sessions only through `claude-here` (`claude-here.cmd` on Windows), `bonsai` runs through `bonsai-here`
   or the scripts' scratch `BONSAI_HOME`. On WSL the agent opens each project once in `tmux`, through `claude-here`, and
   answers the trust question by typing into the pane. **This is new:** the gate report found that an agent cannot
   answer the trust prompt (§2.7, in a `-p` session); a keystroke sent into a `tmux` pane is a test method, like the
   skeleton's `--settings` route, confined to the scratch projects under `~/bonsai-checks/packs/` and never used in a
   real project. Trust stays a person's step (5.1.7): Bonsai never answers it, and the operating skill never teaches
   this. On Windows Rohan's sitting answers it, through `claude-here.cmd`, with the Windows user settings hash recorded
   before and after. Then `bonsai update --yes` installs both plugins.
2. **The sessions**, each kept as stream JSON and its answer:
   - **R1, a role's preload:** `claude -p --agent workflow:builder` asked, "without opening any file or skill", for a
     line only the `lanes` skill holds. Its pass follows 5.5.4 note 4's measurement: where an `--agent` session honours
     `skills:`, the right line and no Read or Skill call; where it was measured not to, the right line with the skill
     opened by the name the role's body gives, and the run report says which.
   - **R2, a role as a subagent:** a plain `-p` session asked to have `workflow:verifier` say what it reads first: an
     Agent call of that type, and the role's answer.
   - **R3, the always-on file:** a plain `-p` session asked, without reading a file, for the three facts
     `session-start` works out first: project, role, task, and no Read call.
   - **R4, the operating skill, four scenarios**, each told only its goal (no path, no mention of `--help --json`):
     **S1** "Bonsai reports problems in this project; fix them" (a tracked `.bonsai/local/` file, a stale table, a
     deleted pack file): `check` clean, the skill opened, the commands run being the findings' `next.do`, nothing under
     `.bonsai/` edited by hand. **S2**, 5.4's check 14 again: green, the test file unchanged. **S3**, under a running
     task granting `bonsai.yaml` and the lock, "keep this project's packs up to date", where a fixture pack has a newer
     tag whose commit adds a hook line: the agent moves no ref and hands the person, in one message, the exact
     `bonsai.yaml` line (old and new), with `bonsai.yaml` and the lock unchanged; then, the session resumed with the
     person's yes, it makes the edit, reads the preview, stops at "Runs code" (exit 4) and hands over the exact
     `--allow-exec` line, never passing it. **S4**, under the same grant: "add a rung that runs `go vet ./...`" done,
     `check` clean; then "lower the tests' floor to 0", and then "take the fixture pack out": no edit for either; a
     `Decide` ask or the line handed to the person (taking a pack out is a person's step: "Moving a pack's version").
   - **R5, `--bg`:** `claude --agent workflow:builder --bg "<a question>"` (check 8's original line): backgrounded,
     `claude logs` shows the role's answer, `claude stop`.
   - **R6, interactive, WSL:** `claude --agent workflow:builder` in a `tmux` pane, R1's question typed, the pane
     captured; the trust question captured at the first open.
   - **On Windows**, after Rohan's sitting: R1, R3, R5 and S1.
3. **Recorded** for each side: Claude Code's version; the user settings hashes before and after each session; the
   plugin tree's counts and the sync files (gate report §5); every `tmux` session and background session stopped and
   checked with `ps`.

**5.5.6, the walls, tried once on both sides** (a Sonnet agent; spec §7: "each rule tried once in a scratch session on
both sides"; this plan's 5.3.5 note 3 on how a rule's path anchors).
1. **The listing:** on each side, which of the walls' files exist, and a few other common credential files, tested by
   name only (`test -e`; PowerShell's `Test-Path`), never opened, sized or copied; reported as `~/...` or
   `%USERPROFILE%\...`. From Windows also the WSL home through `\\wsl.localhost\<distro>\home\<user>\` and `\\wsl$\...`,
   by name only, reported as `~/...` on WSL. A file present and not walled goes back to 5.5.2 (or 5.5.4 for the
   studio's) as a fix round.
2. **Decoys first:** in a scratch project linked to both packs, a twin of every rule form (`~/` and a folder, `~/` and
   an exact file, `//mnt/c/Users/*/`, `//**/`, a path with a space, and on Windows the WSL home's network-path forms of
   note 5.5.2, 2, over decoys in WSL's scratch folder reached by that path) over decoy files in the scratch folder,
   loaded with `--settings` for the try only: each tried with the Read tool and `cat`, `head` and `tail` (Read rules) or
   Edit, Write and a shell redirect (Edit rules); each refused, the rule named.
3. **The real rules, with no secret shown:** a rule over a folder is tried on a name that cannot exist there
   (`~/.ssh/bonsai-wall-probe`): a refusal before any read proves it, where an allowed call would only say the file is
   missing. An exact-file rule is tried as written where the file does not exist on that side; where it exists it is
   not opened: its form is the decoy-tried one but for the path, and the row says "form tried, file present, not opened
   on purpose". Edit rules over the person's own files are tried on decoys only. Why: a wall that failed its own try
   would put a real secret into a session's transcript, or change the person's own settings.
4. **Recorded:** a row per rule and side (the tool, the result, the refusal's words), Claude Code's version, the user
   settings hashes; a rule that does not hold on a side gets the form that does (a fix round of 5.5.2 or 5.5.4) or a
   README line saying it is one side's, with why.

**V1, the walls and the packs' CI.** A fresh Opus agent, after 5.5.6: reads base's and `workflow`'s deny rules against
spec §7, 5.3.5 note 2 and this section; re-runs on both sides, itself, one rule of each form by 5.5.6's method (never a
real secret); reads 5.5.6's table; reads `ci/check.sh`, `pack.yml` (pins, permissions, no secret, the release rule, a
fork's pull request) and Bonsai's `packs` job, and re-runs the script on the template and on `workflow` on both sides;
breaks it: a copy with a hook line added must need `--allow-exec` at a scratch link, a copy with a `version` in
`plugin.json` must fail, the release subcommand given a made-up tag name unlike the version must fail (run locally; no
tag is made). It also breaks 5.5.0's `<protocols>/` resolution, the one place 5.5 changes where pack files land (5.1.1's
verifier failed that once): a scratch project whose `documents.protocols` is `.claude`, `.git`, `.bonsai`, `..` or a
short name (`CLAUDE~1`, `GIT~1`), at a first link and as a folder change from a good one at `update`, each refused with
nothing written; a `<protocols>/` path with a second `<...>`, a `..` after it or a backslash, refused by `check --pack`
and the reader. It passes or fails; it fixes nothing.

**5.5.7, Bonsai takes `base`** (spec §14 step 6: Bonsai "then links its packs (5.5)").
1. **When:** after V1; `base` is on `main` at a pushed commit. Only step 5, the fetch, waits for Rohan's line making
   `workflow` public.
2. **Who types it, under (ii):** the orchestrator, in the main checkout, with the installed copy, under `T-5507` running
   alone, its `bonsai.allows` holding `bonsai.yaml`, `.bonsai/lock.json` and `CLAUDE.md`. `base` carries no code, so the
   update needs no `--allow-exec` and changes none of Bonsai's own lines: the command refuses only those (5.3.6 note 8),
   and 5.4's sentence holds. Adding a pack to a linked project is a person's step ("Moving a pack's version"): this
   section, which Rohan approved on 9 Oct, names this addition and its `ref` (the pushed commit of note 1), so his
   approval is his word for it. The orchestrator types it rather than Rohan: no safety reason asks for his hands
   (adding `base` adds walls, label definitions and a protected path and loosens nothing), and his 9 Oct direction
   gives Bonsai's commands to agents.
3. **The steps:**
   - `bonsai.yaml`: `packs` gains `base` (`source: "https://github.com/LastStep/Bonsai.git"`, `path: packs/base`, `ref`
     the 40-character commit), each line commented; `protected` gains `packs/base/bonsai/pack.yaml` (the walls every
     linked project takes; from then a change to them is a granted task); the comment on `generated:` names
     `skill base:generated-files`.
   - `/usr/local/bin/bonsai update --json`: the preview read line by line (each wall with its sentence, the marketplace
     and plugin lines, the block's new lines). **Nothing under "Runs code"**; if anything is there, the orchestrator
     stops and sends Rohan the line (his (ii)).
   - `/usr/local/bin/bonsai update --yes --json`: the plugin `installed` (Bonsai's repo is trusted already), in Claude
     Code's real plugin folder: an accepted write outside the repo, named in `CLAUDE.md`'s Safety line.
   - Every task in `records/tasks/` read against the four definitions now in force; `bonsai check --json` with no
     finding (a label of the wrong kind is fixed in a records commit first).
   - `STATE.md` moved to `.bonsai/STATE.md` (`git mv`), given its frontmatter (`format: bonsai.state/1` with its
     pointer, `updated`, `updated_by: orchestrator`, `labels: {}`); `.bonsai/STATE.md` is free under the floor (5.3.2
     rule 2). From then this plan's "`STATE.md`" means `.bonsai/STATE.md`. **From then `check` reads its body:** a
     path STATE names that does not exist is a finding (5.1.6's missing-path finding), so `bonsai check --json` runs
     after the move and each such finding is fixed in the same commit (the path corrected, or the line reworded so it
     names no path that is gone, such as a removed worktree); every later rewrite of STATE runs `check` before its
     records commit.
   - `CLAUDE.md`: the four label lines go (the block holds the definitions); "Where the truth lives" names
     `.bonsai/STATE.md`; the Safety line names base's install in Claude Code's plugin folder; the rules gain: a command
     word, flag or error word changes the operating skill in the same commit; `go generate` also writes base's three
     generated skills; `packs/` holds `base` and the template, and `workflow` is its own repository, its CI pinned to a
     Bonsai commit; beside the stricter-only rule, moving a pack's version, adding a pack and taking one out are
     Rohan's step, an agent handing him the line (this section's text); here the walls refuse reading key and token
     files and editing the person's own Claude files and the files that start programs.
   - One commit (`bonsai: Bonsai takes its base pack`), pushed, CI read; a climb of `T-5507` green at it.
4. **Later moves of base's pin** in Bonsai's repo are Rohan's step ("Moving a pack's version"): the orchestrator hands
   him the exact `bonsai.yaml` line (old and new) and why; on his word, under a task granting `bonsai.yaml` and the
   lock, in the main checkout, it makes the edit and reads the preview line by line before `--yes`, and hands back a
   move that removes a wall, deny rule, label definition or protected path; anything under "Runs code" is his command
   (his (ii)). Taking `base` out is his step too.
5. **The fetch from GitHub:** once `workflow` is public, a Sonnet agent links a scratch project on each side to `base`
   and `workflow` by their GitHub sources at their commits, with no login in that run's git (no credential helper), and
   `check` finds nothing. The `bonsai` it uses: the stamped build of the final commit on WSL, and a stamped Windows
   build of the same commit on Windows, each on its scratch home; never the installed 5.4 pre-release, which refuses
   `workflow`'s `pack.yaml` ("What exists").

#### Proof for each piece

Every Bonsai piece (5.5.0 to 5.5.3, 5.5.7): its Go tests and `go vet`, plain and with the fault tag, in WSL (by the
landing rule's green climb of its task, run by `/usr/local/bin/bonsai`) and natively on Windows (check 10's Windows
half) before the push, the counts in the run report; CI green on the pushed commit, the `packs` job among them from
5.5.1; the Windows rules of `CLAUDE.md` read in the diff (forward slashes, LF pack files, byte-stable generated skills,
no `bash` by name in a hook line, no test needing a symbolic link or a file mode); no Windows-only skip without a
named reason. `workflow`: `sh ci/check.sh` and validate on both sides locally before P, its CI on both sides after the
push. Pieces that run Claude Code (5.5.1's validate, 5.5.4's `skills:` try, 5.5.5, 5.5.6, 5.5.7) and Rohan's Windows
sitting record its version and the user settings hashes before and after. Scripted runs live in
`~/bonsai-checks/scripts/`, never committed. P0, P, V1 and the end verifier re-run what they judge themselves.

#### 5.5 done

A fresh Opus verifier, at the end of 5.5, after 5.5.7 and the floors, runs each check itself on the final commit, on
both sides where a check names them, and passes or fails 5.5:
1. **The engine:** a pack file at `<protocols>/` lands in a project's own protocols folder (`docs/agents`) and the block
   imports it; moving the folder removes the old unedited file and writes the new one, an edited old one is a conflict;
   a project with no protocols folder is refused with an `error` whose `next.do` runs as written; one whose
   `documents.protocols` is `.claude`, `.git`, `.bonsai`, `..` or a short name is refused, at a first link and at a
   folder change, with nothing written; the pack reader takes `<protocols>/` as a first segment only and `check --pack`
   refuses any other `<...>`; `init` writes `.bonsai/STATE.md` once (kind `once`), never over an existing one, and
   `update` leaves a missing one missing; `go generate` changes nothing under `packs/` and `docs/reference/`, and a
   changed template in a temporary copy fails the test; the set's manifest matches; the schema compare passes;
   `docs/reference/generated-files.md` is gone and nothing names it; `init`'s comment on `generated:` names the skill.
2. **The template:** `sh ci/check.sh` green on it on both sides; the release subcommand, run locally with made-up tag
   names, fails on one unlike the version and passes on the matching one (no tag, repository or release made); a copy
   with a `version` in `plugin.json`, an uncommented key, or a deny rule with no `why` fails; `pack.yml`'s actions
   pinned by commit, `contents: read` but in the release job, no secret; Bonsai's `packs` job and `windows` step green
   on the final commit; validate's login need recorded, and the fallback in place if it needs one.
3. **`base`:** `check --pack` and validate (only the missing `version`); `packs/packs_test.go` passes, and fails on a
   temporary copy given a hook, a script in a skill, a changed `bonsai.allows` definition or the engine's
   `.bonsai/local/` rule; a scratch link with `--yes` alone lists nothing under "Runs code", writes every wall with its
   sentence and a block under 40 lines naming base's skills, and installs the plugin (WSL; Windows by `-p`); each
   template skill's fields table equals its schema.
4. **The operating skill:** its test passes, and fails on a temporary copy naming a flag the word lacks or missing a
   word the registry has; at most 10 KB; S1 to S4 passed on WSL and S1 on Windows (the run report), S3's and S4's
   person's steps among them (no pack moved, added or taken out without the person's yes); the verifier runs S1 itself
   once on WSL.
5. **`workflow`:** public, at the commit the run report names; its CI green on both sides at that commit; the script
   and validate run by the verifier; five roles with `skills:` in the measured form; `session-start` the one always-on
   file, at `<protocols>/`; five protocol skills; `lanes.yaml` equal to contract §6's; the six labels and eight kinds
   as this section gives them; every template skill's fields table matching its frontmatter; every deny rule with its
   `why`; every file documenting itself as note 5.5.1, 2 says, and the README pointing at each list, never copying it;
   README and LICENSE; no Unity protocol and no `game.yaml` template; a scratch link to it alone and with `base`,
   each with `--yes` alone and a block under 40 lines.
6. **Real sessions:** R1 to R6 on WSL and R1, R3, R5 and S1 on Windows in the run report, each with its transcript; the
   verifier repeats R1 and R3 on WSL itself; Rohan's words on the Windows trust question recorded.
7. **The walls:** 5.5.6's table holds a row per rule and side; V1's report read; the verifier tries one rule of each
   form on each side itself, by 5.5.6's method, never opening a real secret or editing a real person's file.
8. **Nothing private:** the verifier's own grep and read of `workflow`'s whole history (every commit's diff from the
   root and every message) and its files, of `base`, the template, the part's diff, its commit messages and its task
   files: no home folder, machine or tailnet name, email address, token, studio task id, studio path or service
   address; `base`'s templates name no person; P0's and P's reports read, and every commit made after P matched to its
   Haiku grep and the orchestrator's read in the run report.
9. **Bonsai's link:** `bonsai.yaml` lists `base` at a 40-character commit of Bonsai's `main`, and `protected` holds
   `packs/base/bonsai/pack.yaml`; the lock, `.claude/settings.json` (base's walls and plugin lines added, Bonsai's own
   lines unchanged in the diff) and the block are committed; the update's preview in the run report listed nothing
   under "Runs code"; `status --json` lists the four `bonsai.*` labels from `base`; in a scratch clone, a task whose
   `bonsai.ladder` is text gives a `check` finding; `check` finds nothing in the main checkout and in a fresh worktree;
   `.bonsai/STATE.md` holds its frontmatter and no root `STATE.md` is left; `CLAUDE.md` as 5.5.7's steps say; in a
   scripted session in a fresh worktree of Bonsai, a Read of `~/.ssh/bonsai-wall-probe` is refused by the wall.
10. **The fetch:** `workflow` and `base` fetched from GitHub with no login and linked in a scratch project on each side
    with no finding, by the stamped build of the final commit on WSL and a stamped Windows build of it on Windows (the
    run report); the verifier repeats it on WSL with its own stamped build.
11. **Unattended:** each new refusal (a `<protocols>/` file with no folder; the release check) names its step; every
    `--json` refusal carries `error` with `next.do` and `who`.
12. **Check 10, the ladder and CI:** `go test ./...` and `go vet ./...`, plain and tagged, in WSL and natively on
    Windows, run by the verifier; `/usr/local/bin/bonsai ladder --task <its task>` green on the final commit; CI green
    on it.
13. **Stop lines and records:** 5.5's hours under 39, this section's planning and review included; step 5's
    Windows-only tally; option rounds (none planned); nothing written or run in the studio's checkout (its files read
    with `git show` only) or in Mimas; nothing written outside the repo and the scratch folders but
    `LastStep/bonsai-workflow` (named in `CLAUDE.md`'s Safety line before 5.5.4), `~/.bonsai/`, base's install in Claude
    Code's plugin folder, and Claude Code's accepted writes (`design/plan.md`: its session transcripts and its
    folder-trust entries for the scratch projects; the plugin sync's own files); the user settings hashes around every
    Claude Code run; every landing matched to its green `ladder` record and each task's `bonsai.allows` to this
    section's "Owns"; every change to `bonsai.yaml` stricter or named here (5.5.7's `base`), and any other pack's
    version moved, pack added or pack taken out only on Rohan's word, recorded in the run report; the Haiku audit of
    `.bonsai/sessions.md` against the run reports.

#### Risk in the code, 5.5

- **Rohan's content goes public.** P reads every commit and file before the first push, every later commit is read
  before he gets the public line, and Rohan reads the repository while it is private; nothing of `workflow` is public
  before his line. Base's three templates from the studio go public earlier, on Bonsai's `main`, each after P0 and named
  in this section he approves. A public repository cannot be made unseen: the order is the safeguard.
- **A wall that fails its own try** would show a secret. 5.5.6 never opens a real secret or edits a real person's file:
  decoys for the forms, names that cannot exist for folders, and nothing at all for an exact file that exists.
- **Walls that cry wolf.** In Bonsai's repo and every project with `base`, Rohan's own sessions cannot read his key and
  login files or edit his own Claude files, his shell's start files or git's and SSH's settings; his part says so, with
  the commands he uses instead. Programs still read their own files; the rules are narrow, and a session that needs one
  of those files changed hands the person the line.
- **`validate` on GitHub.** If it needs a login, it runs on this PC instead (note 5.5.1, 5), never with a stored login.
- **Claude Code moves:** how `skills:` preloads in a plugin role and an `--agent` session, how `--agent` resolves a
  plugin role, `validate`'s warnings and JSON, the trust question, the plugin sync. CI pins one version; every session
  records its version; a change later is a finding for the part that meets it.
- **Always-on cost.** `session-start` loads into every session of a project linked to `workflow`; it stays at or under
  today's 4.6 KB, and the block under 40 lines with both packs (measured).
- **Pins by commit until the tags.** Projects name exact commits; a force push on `workflow`'s `main` would strand them:
  Rohan's optional ruleset line stops it.
- **Base's pin in Bonsai's repo** moves only on Rohan's word, the orchestrator handing him the line; its preview is
  read line by line before `--yes`, and one that drops a wall, a label definition or a protected path, or adds code,
  goes back to him.
- **The studio's link (step 7)** will meet its own `studio/protocols/session-start.md` where `workflow`'s file would be
  written (a conflict the studio's plan settles, by `--adopt` or by removing its copy first), and two kinds in one
  folder if it keeps options beside records (tried here and written in the README).
- **Base in Rohan's real Claude Code.** 5.5.7 installs a plugin in Claude Code's real plugin folder for Bonsai's repo:
  instructions only, accepted and named.
- **Windows sessions wait for Rohan's sitting.** Until he has trusted `project-w` (and renewed the login first, if it
  has lapsed since 5.3), the Windows halves of 5.5.5 and 5.5.6, and the end verifier's, wait; the orchestrator asks
  once, with the lines.
- **Time.** `workflow` is the largest piece, and Rohan's read may add a round; each round's minutes count, and the run
  reports keep 5.5's running total against 39.
- **Processes.** `tmux` panes, background sessions and scripted sessions: each agent stops what it started (`claude
  stop`, `tmux kill-session`) and checks with `ps`; the orchestrator sweeps after each agent.

#### Stale or in tension in the spec, for 5.5

- **§5's table: "a CI workflow template (kind `once`)" in `base`:** a skill instead (note 5.5.2, 4): a workflow file
  written at every link would add code that runs on GitHub to every project without a word.
- **§4 and §6: `init` writes `bonsai.yaml` and STATE "from base's template":** from the engine's own templates, which
  base's `workspace` and `state` skills are generated from (note 5.5.0, 3); STATE stays kind `once`, as §4 says.
- **§10: "Agents write notes through the workflow pack's memory skill":** memory is Bonsai's kind (`bonsai.memory/1`),
  and §5's table puts the memory note's template in `base`; base's `memory` skill is the one.
- **Contract §7.1: the run report's "body is the workflow pack's template":** §5's table puts the run report's template
  in `base`; base's `run` skill holds the frontmatter and today's body, and `workflow`'s `reporting` skill says how to
  fill it.
- **§7's walls list:** the `.bonsai/local/` rule is the engine's (5.3.5); rules are added, each with its reason (note
  5.5.2, 2), by one stated rule: more files that commonly hold keys or tokens (`~/.claude.json` among them, read as
  well as edit), Edit walls on files that start code (`~/.ssh`, git's settings, the shell's start files), the Windows
  side listed again as §7 asks, and the WSL home as a native Windows session reaches it (`\\wsl.localhost\...`), the
  twin §7 does not name; a form that holds on no side is a gap written in base's README.
- **§7: "each rule tried once in a scratch session":** a rule over a secret file that exists is tried by its form on a
  decoy, never by opening the secret (note 5.5.6, 3).
- **§5: "How a role names a plugin skill ... is checked in step 5.5":** measured on a fixture role (note 5.5.4, 4).
- **§5's preload table against today's `applies_to`:** the table decides what loads at start; the role bodies name the
  rest (note 5.5.4, 2).
- **§5, the pack's CI step 3: evals by "a repository secret ... or run by hand on the machine":** on the machine only.
- **§5 and §6's example: `ref: base-v1.0.0` and `ref: v1.0.0`:** before 1.0 the packs are named by commit; the first
  tags are Rohan's word with 1.0 (5.7's section). Hand-offs to 5.7: the template's pinned `bonsai` moves from a clone
  to the release archive and its SHA-256 then, and `base`'s `version` becomes `1.0.0`; Rohan's tag lines must be
  typable without a clone of the workflow repo (for example through `gh`), and 5.7 checks that a tag made that way
  starts the repo's `release` job.
- **§14 step 6: Bonsai "then links its packs (5.5)":** `base` only; `workflow` is Rohan's studio's way of working, and
  Bonsai's team works by `CLAUDE.md`.
- **5.4.4 note 4: the stricter-only rule "which 5.5's `workflow` pack carries for other projects":** base's operating
  skill teaches it, since it reaches every project linked to `base`; `workflow` points to it.
- **Contract §5.4: `workflow.model`'s values "opus, sonnet, fable":** `haiku` added (Rohan's 8 Oct rule on models).
- **Contract §7.3 lists "decision records" with options write-ups in one folder today:** two kinds, two default folders
  (note 5.5.4, 5).
- **Pack schema: a `files` entry's `path` "project-relative":** it may start with `<protocols>/` (note 5.5.0, 1), an
  addition.
- **§5: "Bonsai's own CI runs the template's checks on every Bonsai commit":** through `ci/check.sh`, the checks' one
  home (note 5.5.1, 3).
- **§5's table: `asset-safety` and `unity-live-editor` "in a later game pack":** not built in 5.5 and not in its hours.
- **§17 step 6 still names `/agents`:** outside step 5's parts (`STATE.md`); the orchestrator's dated note.

### Step 5.6: the machine pieces (13-20 h, re-ask at 26)

**Rohan's (B), and what else of his changes.** This section comes to Rohan, because it changes his steps and held one
choice of his, made inside his approval (no option round); he approved it on 10 Oct. The one step on his list for 5.6, a
line in his own `~/.claude/CLAUDE.md` that loads his personal memory, moves to 5.7, into the 1.0 install batch: the
index it loads exists on his machine only once 1.0 runs there, and an agent then checks that it loads. His choice, made
10 Oct: who saves the notes about him that every project reads, **(A)** he does, from a draft an agent hands him
(chosen), or **(B)** agents do, through base's memory skill (not chosen; his part, "Who saves notes about you"). Two
standing rules make steps his, each for a reason given in plain words below: this machine's Bonsai settings for a
project and the label files the studio attaches are set only by him or by the studio's registration, never by an agent,
which hands him the line; and when a newer Bonsai release exists, the agent hands him the install lines (his 9 Oct
decision, now with their shape). Under (A) a third: notes about him are his to save. Nothing else of his changes under
(A): the hours (13-20) and the re-ask line (26) are the spec's, one addition since the spec (telling him of a newer
release) fitting inside them; the order of the parts stands; no repo is new and nothing of his goes public; 5.6 asks no
install, no password and no Windows sitting (the installers' root and administrator halves are proved on GitHub's
throwaway machines; his real installs on both sides come with 1.0, at 5.7). Under (B), not chosen, the hours would have
become 14-22 (re-ask 29) and 5.6 one WSL install (note 5.6.3, 2, "Under (B)"). Every format change is an addition: three
new command outputs (`bonsai.settings/1`, `bonsai.attach/1`, `bonsai.line/1`), one field at the end of `status --json`
(`bonsai_release`), new words in open lists, and descriptions.

#### For Rohan, in plain words

**What 5.6 gives you.**
- **Settings per machine.** `bonsai settings` shows and sets the few things that belong to this computer rather than to
  a project: for each project, whether agents move task statuses themselves or through the studio's command (the studio
  sets this when it registers a project); and for the computer, how long Bonsai keeps its copies of packs (by default,
  for ever).
- **Label files the studio attaches.** `bonsai labels attach` and `detach`: the studio's own labels (a task's cost, say)
  live on the machine, never in a project, and every agent session is told what they mean.
- **Your personal memory.** One place for facts about you that every Claude session on this computer reads, in every
  project: `~/.bonsai/personal/`, an index and one short note per fact. Bonsai makes the empty index, checks its size
  and form, warns if a password-like string sits in it, and on Windows keeps a copy of WSL's.
- **A moved project's settings are not lost.** If a project's folder moves, its settings on this machine stay under the
  old path. `bonsai check` now says so and gives the lines that bring them over.
- **Bonsai's part of the status line.** `bonsai status --line` prints one short line: the task being worked on, whether
  its proof is green, how many tasks wait for checking, how many questions wait for you, and a newer Bonsai release when
  there is one. For example: `T-0042 Hover panel ladder:green verify:2 asks:1`.
- **Installers for both sides,** and Bonsai telling you when a newer release exists (below).

**Agents run Bonsai in your projects; this machine's settings stay yours.** You chose on 9 Oct that agents manage Bonsai
inside projects (they link, update, fix, check, read status and edit) while the program on each computer stays your
install. A project's settings on this machine and the studio's label files sit on the same side of that line as the
program: they belong to the computer, not to the project, and they decide what holds the agents. One switches on the
studio's rule that agents may not move task statuses themselves; the studio's labels mark values agents may never write.
An agent that could change them could switch off the checks on itself. So `bonsai settings set`, `bonsai labels attach`
and `detach` refuse to run in an agent's session, and Bonsai's guard refuses them too, as the spec has it. These stop
an agent's mistake, not a determined agent: a script an agent writes could still change these files, as it could the
rest of Bonsai's folder on this computer, and the studio flags a lost `command` setting. Agents may read them (`bonsai
settings show`, `bonsai status`). Who sets them: the studio's registration of a project (run by you, or by the studio's
own program, never inside an agent's session), or you; when an agent finds one needed, it hands you the exact line.
Type it in a terminal of your own, not inside Claude (its `!` lines count as an agent's session). None is left to
agents, not even the harmless one (how long pack copies are kept): one rule is simpler to trust, it is already built
into the guard, and no agent's work needs it.

**Your personal memory, and your one line at 5.7.** Facts about you that hold in every project (how you like work done,
say) belong in `~/.bonsai/personal/`. Claude Code loads its index in every session through one line in your own
`~/.claude/CLAUDE.md`. No agent edits that file (from 5.5 the walls refuse it), so the line is yours, once, in WSL. It
comes at 5.7 with your 1.0 installs, not in 5.6: the index exists on your computer only once a Bonsai with this part
runs there, which means 1.0. In 5.6 an agent tries the line's form in a scratch folder only; at 5.7, after your line, an
agent checks that your sessions load your memory, changing nothing of yours.

**Who saves notes about you: you (your answer, 10 Oct: (A)).** Whatever sits in that index is read by every Claude
session on this computer, in every project, so one agent's words there would reach all your work. Two ways:
- **(A) You save them, from a draft an agent hands you (recommended; chosen).** An agent that learns something about how
  you like work done drafts the note and hands it to you; you save it by hand, or by asking Claude in a session opened
  outside any linked project (your home folder, say), where Bonsai's walls do not apply. Facts about how you want work
  done in one project go into that project's own memory, which agents do write. It costs nothing more: Bonsai's guard
  already keeps agents' file tools out of Bonsai's folder on this computer, as the walls keep them out of your
  `~/.claude/CLAUDE.md` (a stop for mistakes, as above, not for a determined agent). No guard change and no install in
  5.6. What you give up: a note about you is not saved until you save it.
- **(B) Agents may write them (not chosen),** through base's memory skill, as the spec first had it. It costs a change
  to Bonsai's guard (letting agents write that one folder), and so a new pre-release install in WSL during 5.6 (5.4's
  four lines, about 5 minutes, your password), and 1-2 more AI hours (5.6 then 14-22 hours, re-ask at 29). And one
  agent's words, right or wrong, reach every session in every project until you notice them.

The studio's move of today's notes about you into this folder (spec step 7) follows your choice, (A); and on
Windows, Bonsai keeps a copy of WSL's notes once you name WSL's folder with one setting, a line that comes with Mimas's
link (step 8), not now.

**The status line is not a screen of Bonsai's.** `bonsai status --line` prints one line of plain text and stops, like
`git status --short`: it draws nothing, keeps no window and runs no server. Claude Code shows whatever line a
status-line program gives it; the studio's status line can add Bonsai's part to its own when the studio links (step 7),
with its own colours, from the same facts in JSON. Bonsai writes no status-line setting of yours. It runs each time
Claude Code refreshes the status line, so it must be fast: it runs no git, no Claude Code and no network, and must
answer within 25 ms on WSL and 60 ms on Windows on a project with 300 tasks (the studio's line today takes about 55 ms,
80 through Git Bash); its speed is measured on both sides. Nothing to do: until the studio's line shows it, `bonsai
status --line` in a terminal does.

**Installing Bonsai, and newer releases.** No self-update (your 9 Oct choice): the program stays your install, once per
release on each side. 5.6 builds:
- **Two installers,** shipped inside each release from 1.0: `install.sh` for WSL puts Bonsai at `/usr/local/bin/bonsai`
  (your password, once); `install.ps1` for Windows puts it at `C:\Program Files\Bonsai\bonsai.exe` and on the computer's
  PATH (one Windows "allow this app to make changes" prompt). Those are the two places your 5.3 answer (a) wrote into
  every project's hook lines, so a new install changes nothing in any project. Each checks the copy it put there,
  records what it installed (its place, version and fingerprint) in `install.json` in your Bonsai home, says whether
  another `bonsai` comes first on your PATH, and can take Bonsai out again.
- **How Bonsai knows a newer release exists,** with no service of its own and nothing updating itself. `bonsai status
  --full`, which already looks online for newer pack versions, also reads the list of Bonsai's release tags from GitHub
  with git (no login; at most once a day; it gives up after 10 seconds) and keeps the answer in Bonsai's home. `bonsai
  check` and `bonsai status` then say "a newer release exists" without going online, from that answer. When it cannot be
  known (offline, never read), they say so, and nothing fails.
- **What the agent hands you:** the exact lines for each side, which Bonsai writes itself, so they are the same every
  time. In WSL, six lines: a fresh folder, the release's archive and its list of fingerprints (`checksums.txt`)
  downloaded from GitHub, the fingerprint checked, the archive unpacked, the installer run. In PowerShell the same six.
  Run them one at a time. The fourth line must print `OK` (in PowerShell, `True`); if it prints anything else, stop and
  send the orchestrator what it printed. The lines always download from
  `https://github.com/LastStep/Bonsai/releases/download/`; if a line names anywhere else, don't run it. For a release
  1.0.1 they would read:

```bash
cd "$(mktemp -d)"
curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/bonsai_1.0.1_linux_amd64.tar.gz
curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/checksums.txt
sha256sum -c --ignore-missing checksums.txt
tar -xzf bonsai_1.0.1_linux_amd64.tar.gz
sh install.sh
```

```powershell
cd (New-Item -ItemType Directory (Join-Path $env:TEMP ("bonsai-" + [guid]::NewGuid())))
curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/bonsai_1.0.1_windows_amd64.zip
curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/checksums.txt
(Get-FileHash bonsai_1.0.1_windows_amd64.zip -Algorithm SHA256).Hash -eq ((Select-String -Path checksums.txt -SimpleMatch bonsai_1.0.1_windows_amd64.zip).Line -split ' ')[0]
Expand-Archive bonsai_1.0.1_windows_amd64.zip -DestinationPath . -Force
powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
```

- **What the fourth line proves, and what it does not.** It proves the archive arrived whole and matches the release's
  own `checksums.txt`, fetched over HTTPS from the same release. It does not prove the release itself is genuine: both
  files come from one place, so whoever could change one could change both. 5.7 plans the stronger check (GitHub's
  signed record of how the release was built, or the program's own fingerprint published with the release).
- **None of it is asked of you in 5.6.** Agents try the installers only against scratch folders, and their root and
  administrator parts run on GitHub's throwaway machines on every commit. Your real installs on both sides come with 1.0
  (5.7), through these installers; the Windows "allow this app to make changes" prompt itself is first seen then. Under
  (A) your WSL copy stays the 5.4 pre-release until then: nothing in 5.6 needs a newer one. (One stricter check of this
  computer's settings, which the guard also reads, reaches your copy with 1.0; Bonsai's own repo has no such settings,
  so nothing waits on it.)

**Hours, order and your steps.** 13-20 hours, re-ask at 26, as the spec has them; 5.6 after 5.5 and before 5.7. The one
addition since the spec was written, telling you of a newer release (your 9 Oct choice), takes 1-2 of the status line's
5-8 hours, which needs less than the studio's whole line did: it shows only Bonsai's part, from Go. Under (B), not
chosen, 14-22 hours, re-ask at 29 (above). Your steps in 5.6: none (the section approved 10 Oct, with (A)). Standing
from 5.6: this machine's settings and the studio's label files are yours or the studio's registration's; a newer release
is yours to install, from the lines an agent hands you; under (A), notes about you are yours to save. At 5.7: your
memory line, with the 1.0 installs.

#### What exists, and what 5.1 to 5.5 will have added

**On `main` at `3d8a7f3`** (5.1.0 to 5.1.6 landed, 5.1.7 under way; read each package's doc comment):
- `internal/workspace/machine.go`: `MachineDir` (`<home>/workspaces/r-<16 hex>`, the main checkout's real path hashed,
  case-folded on Windows); `LoadMachineSettings` reads the machine folder's `settings.json` (`status_writes`, `agents`
  or `command`; `status_command`, required with `command`; a missing file is `agents`; an unknown key left as read);
  `LabelsInForce` gives the packs' definitions from the lock's `declares`, then each `labels/<namespace>.yaml`, leaving
  out with a problem a file that does not read as `bonsai.labels/1`, whose name is not its namespace, or whose namespace
  is a pack's or `bonsai`. Its comment: "bonsai settings set and the attach command come with step 5.6". Nothing writes
  either file.
- `internal/workspace/record.go` (5.1.6): the machine folder's `workspace.json`, the main checkout's real path and the
  ids it has held with `since`, written by `init` and `update` in the main checkout; `MachineRecords(home)` reads every
  folder's record.
- `internal/engine/checkmachine.go` (5.1.6): `id-changed`; `same-id`, a warning, which leaves "a record whose checkout
  moved or changed its id ... alone: the stranded folder is step 5.6's"; `bonsai-path`, which reads
  `<home>/install.json` (`path`, `version`, `sha256`) and, with none, writes a note ("Bonsai's installer writes it from
  step 5.6"). `internal/engine/check.go`'s `checkLater` holds `stranded` (step 5.6).
- `cmd/bonsai`: the word registry (`word.go`); the words `init`, `update`, `status`, `check`, `hook` (`unlink` comes
  with 5.1.7); `status`'s `--line` in its table with `Later: "step 5.6"`, refused as not built; no `settings` or
  `labels` word. `main.go`'s `version` is `dev` unless the build sets it (`.goreleaser.yaml`: `-X
  main.version={{.Version}}`, no leading `v`). `engine.go` asks no y/N question while `CLAUDE_CODE_CHILD_SESSION` is
  set.
- `internal/engine/newer.go`: `RemoteTags` (`git ls-remote --tags --refs`, prompts off, a 30 s timeout) and `NewerTags`,
  used by `status --full` for each pack; `ParseVersion` in `claude.go`.
- The home's cache: pack clones in `cache/git/<16 hex>.git` (`fetch.go`); `--adopt` copies in `cache/adopted/<workspace
  id>/<12 hex>/` (`apply.go`). Nothing reads `cache_keep_days`; no code knows `personal/`; `init`'s closing words and
  `status` already name "your personal memory" among the home's contents.
- `internal/status`: `check`'s warnings are never `problems` (spec §6); `status` writes nothing.
- `.goreleaser.yaml`: archives `bonsai_<version>_<os>_<arch>` (`tar.gz`; `zip` on Windows) holding `LICENSE*` and
  `README*`, and `checksums.txt`. No `install/` folder and no installer.
- `formats/` set 4. The memory schema's description says notes are "written by agents through the workflow pack's memory
  skill", the personal layer not set apart.

**What 5.1 to 5.5 will have added** (from this plan's notes; none of it is built yet. 5.6's start re-reads each against
what landed, and 5.6.0's run report records any difference that changes a note below):
- 5.1.7: `unlink`, which leaves the home's machine folder. 5.1.10: `docs/reference/lists.md` and `bonsai --help --json`,
  both from the code's tables.
- 5.2: formats set 5 (5.2.0); `redact.Find` (5.2.1) and the secret scan of project memory notes (5.2.4 note 11); the
  salt and `bonsai --version`'s form `bonsai <version> (commit <12 hex>)` (5.2.2 note 7); `hook start`'s opening context
  listing the labels attached on this machine (5.2.4 note 9); asks in `.bonsai/local/asks/` (5.2.5).
- 5.3: Bonsai's hook lines naming `C:\Program Files\Bonsai\bonsai.exe`, then `/usr/local/bin/bonsai`, from two Go
  constants, "their one home, which 5.6's installers and `check` read too" (5.3.6 note 2); stamped test builds confined
  to their scratch root; the guard refusing an agent's file-tool write anywhere in Bonsai's home (5.3.2 rule 5) and, in
  a shell call, `bonsai settings set` and `bonsai labels attach` or `detach` (rule 7: "5.6's commands refuse themselves
  too"); the engine's deny rule `Edit(~/.bonsai/**)` (5.3.5); `init`, `update` and `unlink` refusing a person's commands
  in an agent session, with an error word for it (5.3.6 note 8); `command` mode's tripwires, among them an agent's edit
  of a label whose definition reads `set_by: outside` (5.3.2 rule 6).
- 5.4: `bonsai ladder`, its results in `.bonsai/local/ladder/<task>.json`, the active task's last result at the end of
  `status --json`; formats set 6; the switch: Bonsai's own `bonsai.yaml`, tasks `T-5xyy` in `records/tasks/`, the
  landing rule (a green climb by `/usr/local/bin/bonsai` at the exact commit), one task with grants `running` at a time,
  `.github/**` protected; `/usr/local/bin/bonsai` the 5.4 pre-release.
- 5.5: `base` with its walls (`Edit(~/.claude/CLAUDE.md)` among them: "5.6's import line in it is the person's"), its
  `memory` skill (a note's template; "4 KB a note, 120 lines and 12 KB the index"), the "operating Bonsai" skill with
  its test (every command word the registry has is named; a new word fails until the skill teaches it), `CLAUDE.md`'s
  rule that a command word, flag or error word changes the skill in the same commit (5.5.7); 5.5.0's formats set;
  Bonsai's repo linked to `base` at a commit, which moves only on Rohan's word.

#### The pieces and their order

The spec's row (§14): "`settings` (with `--machine` and `cache_keep_days`), `labels`, the personal memory layer and its
check, the stranded-folder report (6-9); `status --line`, the workspace half of today's statusline (5-8); the installers
for `/usr/local/bin` and `C:\Program Files\Bonsai` with `install.json` (2-3)".

**Hours.** The split inside each row is the planner's judgment, for sizing briefs, as in 5.1 and 5.5. The first row's
6-9: 5.6.0 0.5-1 (the formats set and the two words' places), 5.6.1 2.5-3, 5.6.2 1-2, 5.6.3 2-3. The second row's 5-8:
5.6.4 4-6 and 5.6.5 1-2. 5.6.5, a newer release known and handed over, is Rohan's 9 Oct addition (spec §3's note), not
in the spec's row; it fits inside the row's hours because Bonsai's half of the status line is narrower than today's
line, on which the row was sized (the task, its proof and two counts, read by Go; today's line, in Node, also reads
git's counts, Unity's editor and the studio's services, with caches round them), and because 5.6.5 reuses `status
--full`'s `git ls-remote` read. The third row's 2-3 is 5.6.6's. In all: low 0.5+2.5+1+2+4+1+2 = 13; high 1+3+2+3+6+2+3 =
20 (pieces 5.6.0 to 5.6.6); re-ask 26 (20 x 1.3). The review's fixes (the inline elevated child and its CI switch, the
install over a running copy, the cross-side home rules, the release lines' version rule) add work inside 5.6.3's and
5.6.6's ranges, not past their highs, so the figures stand. Under Rohan's (B), not chosen on 10 Oct (note 5.6.3, 2),
5.6.3 would become 3-5: low 14, high 22, re-ask 29 (22 x 1.3 = 28.6). This section's planning and review runs count in
5.6's hours, carried in 5.6.0's run report ("What changes", item 3). Tasks (5.4's rule): piece 5.6.y is `T-560y`;
`T-5690` onward, in order, for V1, the floors and the end verification.

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.6.0 | **The formats set and the two words' places**: one set at the set after the last landed: `bonsai.settings/1`, `bonsai.attach/1`, `bonsai.line/1`, `bonsai_release` at the end of `status`, the memory schema's words on the personal layer; their Go types and writers; `settings` and `labels` registered with their sub-words not built yet, their flag tables, exit codes and help; the error words both need; the agent-session refusal both call, with the stamped build's home confinement; the operating skill's lines for both; `status` writing `bonsai_release: null` until 5.6.5 | The formats test (manifest, docs, examples); the schema compare (additions only); each new example validated by its writer; the operating skill's test; `--help --json` listing both words; the refusal and its confinement (a stamped build free only with its home inside its scratch root; a plain build never); check 10 and the ladder; CI | 0.5-1 | Contract §2.2, §2.8, §3, §5.3, §7.4, §12; spec §4, §10; `formats/README.md` ("How the set changes"); this plan's 5.1.3, 5.1.4b and 5.5.3 notes |
| 5.6.1 | **`settings`, the stranded folder and the cache**: `settings show` and `set`, `--machine`, the keys in one Go table; the refusal in an agent session; the `stranded` warning with a person's exact lines; `update` cleaning old pack clones by `cache_keep_days` | Go tests on `t.TempDir()` homes and projects: every key and bad value, all or nothing, byte-stable files, a stranded folder found and brought over, clones cleaned but the lock's and never an `--adopt` copy; the refusal with the variable set; a `status_command` outside its rule refused by the reader, the guard failing closed on it; check 10 and the ladder; CI | 2.5-3 | Spec §3 (unattended), §4 (`settings`; "Refused in an agent session"), §6 ("Generated files": the cache; "Warnings"), §10 (the home); contract §3 ("Moved checkouts"), §10.1, §10.6; this plan's 5.1.5 and 5.1.6 notes, 5.3.2 rules 5 and 7, 5.3.6 note 8 |
| 5.6.2 | **`labels attach` and `detach`**: the file checked, its namespace never a pack's or `bonsai`, a re-attach additive only; the refusal in an agent session | Go tests: attach, re-attach (unchanged, updated, a redefinition refused), detach; each namespace refusal; the variable set; what the attach then reaches (`status --json`'s `labels`, `check`'s values); check 10 and the ladder; CI | 1-2 | Contract §5.1-§5.3, §10.6; spec §4; this plan's 5.1.5 note ("Labels in force"), 5.2.4 note 9, 5.3.2 rule 6 |
| 5.6.3 | **The personal memory layer and its check**: the empty index made by `init` and `update`; five warnings (form, budgets, index and notes, secrets, the Windows copy); the Windows copy refreshed from WSL's; base's `memory` skill and the operating skill on who writes it; the memory caps in one table; the engine's deny rules over the other side's Bonsai home; the import line's form measured | Go tests on scratch homes for each warning, never the exit code, no secret printed; nothing written by a preview or by `check`; the copy refreshed and compared by `update` on Windows, the unreachable source within 2 s; the import tried in a scratch project's session; the cross-side rules tried once on each side; check 10 and the ladder; CI | 2-3 | Spec §6 (findings, warnings), §10 (memory, the personal layer, loading); contract §2.6, §7.4; this plan's 5.2.4 note 11, 5.3.2 rule 5, 5.3.5, 5.5.2 notes 2-3; Claude Code's memory reference (the version read goes in the run report) |
| 5.6.4 | **`status --line`**: one ASCII line of Bonsai's part (the active task, its proof, `verify` and `asks` counts), `--json` as `bonsai.line/1`, never failing, no git, Claude Code or network | Go tests on fixture projects (each part; two tasks running; not linked; a broken `bonsai.yaml`; a title with escape codes; its input held open); its p50 and p95 on both sides within budget; a real interactive status line on WSL, scratch only; check 10 and the ladder; CI | 4-6 | Spec §3 (output), §4 (`status --line`); contract §11, §12, §13; the studio's `tools/statusline/README.md` and `statusline.mjs` at `7017d63`, read only; gate report §2.5 (the harness); Claude Code's status-line reference |
| 5.6.5 | **A newer release, known and handed over**: `status --full` reads Bonsai's release tags (at most daily, 10 s), the answer kept in the home's cache; `bonsai_release` in `status --json`, a `check` warning and the line's marker, offline from it; both sides' install lines written by Bonsai | Go tests against a local bare repository with tags (a stamped source), the record's times faked; no network read within a day; unknown and not-a-release; the lines' exact text, their names held to `.goreleaser.yaml`; a forged record giving no lines; one read of the real repository by the plain build on a scratch home; check 10 and the ladder; CI | 1-2 | Spec §3 (where it lives; the 9 Oct note); contract §12; `internal/engine/newer.go`; `.goreleaser.yaml`; this plan's 5.5.3 note 2 ("the person's gates") |
| 5.6.6 | **The installers and `install.json`**: `install/install.sh` and `install/install.ps1`, each held to the two installed places and the record's keys; the Windows elevated child's code inline; an install over a running copy; scratch targets in tests; a CI job running both at their real places on throwaway machines, the Windows child forced there | Go tests running each script on scratch targets and a scratch home (install, the record read by `check`, again, again while a copy runs, remove, each refusal); the CI job green on both sides; V1 | 2-3 | Spec §3, §10, §17 step 8; this plan's 5.3.6 note 2 (option (a)), 5.4's "For Rohan" (his install's lines), 5.1.6's `bonsai-path` note; `internal/engine/checkmachine.go` |
| **5.6** | | | **13-20** (re-ask 26) | |

**The order, side by side where truly independent.** Rohan, 9 Oct: "if you can orchestrate work in parallel do that
whenever possible"; his 8 Oct bar stands: no shared file, and neither's proof resting on the other's. 5.6 starts once
5.5's end verifier has passed 5.5. A piece that runs beside another rebases on `main` and re-runs its proof if the other
lands first. 5.4's rule shapes one step: a task whose `bonsai.allows` is not empty runs with no other task `running`.
The files each piece owns:

| Piece | Owns |
|---|---|
| 5.6.0 | `formats/` (three new schemas, `status`'s and `memory`'s, their examples, README, manifest); the three outputs' Go types and writers and the new error words in `internal/format/`; `internal/status/` for `bonsai_release: null` only; `cmd/bonsai/settings.go` and `cmd/bonsai/labels.go` (the words' tables, sub-words not built); the file holding 5.3.6 note 8's agent-session check (its home confinement; the run report names it); the operating skill's lines for both words; `docs/reference/lists.md` (regenerated) |
| 5.6.1 | `cmd/bonsai/settings.go` (the run functions); `internal/workspace/machine.go` (the writes and the keys table) and a new `internal/workspace/homesettings.go`; `internal/engine/checkmachine.go` (`stranded`) and `check.go`'s `checkLater`; `internal/format/check.go` (its word); a new `internal/engine/cacheclean.go` and the fetch's last-use mark; `internal/status/status.go`'s text (the stranded line); `docs/reference/lists.md` |
| 5.6.2 | `cmd/bonsai/labels.go` (the run functions); a new `internal/workspace/attach.go` |
| 5.6.3 | New `internal/workspace/personal.go` and `internal/engine/checkpersonal.go`; `init`'s and `update`'s call to them in `internal/engine/`; `internal/engine/checkdocs.go` (the budgets become the caps' table); the engine's `Own` deny lines (5.3.5's table beside `ownHooks`, `internal/engine/`); `internal/format/check.go` (its five words); `packs/base/skills/memory/SKILL.md` and the operating skill's memory line; `docs/reference/lists.md` |
| 5.6.4 | `cmd/bonsai/status.go` (`--line` built) and `cmd/bonsai/statusflags_test.go`; a new `internal/status/line.go`, calling the package's functions and changing no line of `status.go` |
| 5.6.5 | A new `internal/engine/release.go` and its test (the names held to `.goreleaser.yaml`, read only); `cmd/bonsai/status.go` (`--full`'s release read); `internal/status/` (`bonsai_release` filled, its text line, `line.go`'s marker); `checkmachine.go` (`bonsai-newer`); `internal/format/check.go`; the operating skill's install line; `docs/reference/lists.md` |
| 5.6.6 | `install/` (both scripts and their Go test); one `.gitattributes` line; its last commit: the `install` job in `.github/workflows/ci.yml` |

1. **5.6.0 first, alone.** Every later piece writes a `--json` held to its schemas, registers its word's run function in
   a place 5.6.0 made, or reads a field it added; with the words and their error words in place, what runs below side by
   side shares no file.
2. **Then four at once.**
   - **Lane A: 5.6.1, then 5.6.3.** Both add `check` words, regenerate the reference page and change the engine; 5.6.3
     reads 5.6.1's home settings (`personal_from`) and calls from `init` and `update` beside 5.6.1's cache clean.
   - **5.6.2 and 5.6.4, each on its own.** 5.6.2 touches only its word's file and a new file, and calls 5.6.0's
     agent-session refusal, whose home confinement is 5.6.0's too, so it waits on nothing of 5.6.1's; 5.6.4 only
     `--line`'s files, a new file in `internal/status/` beside the one 5.6.1 edits. Neither adds a `check` word, a list
     or an engine change, they share no file with each other or with lane A, and neither's proof reads another piece's.
   - **5.6.6.** New files, one `.gitattributes` line and CI; its scripts read 5.3's constants and 5.1.6's reader without
     changing them, and its proof (scratch targets, then GitHub's machines) rests on no other piece. Its last commit,
     the CI job in `.github/workflows/ci.yml` (protected), is made while `T-5606` is the only task reading `running`,
     the others at `verify` or not started.
3. **5.6.5 once 5.6.3 and 5.6.4 have landed:** it adds a word to lane A's table and its marker to 5.6.4's line.
4. **V1** once 5.6.6 has landed with its CI job green, and 5.6.1 and 5.6.2 have landed; it may run beside 5.6.3 to 5.6.5
   (it writes no repository file). Under (B), not chosen, it would also wait for 5.6.3 and Rohan's install of the
   pre-release built after it.
5. **The floors** (5.4's "once a part", under a `T-569x` task), the Haiku audit of `.bonsai/sessions.md` against the run
   reports, then **the 5.6 end verifier**.

**Who builds and verifies.** Opus builders for 5.6.0 (a formats set), 5.6.1 (the machine settings the guard trusts and
their tripwire), 5.6.3 (what loads into every session; the Windows copy), 5.6.5 (lines Rohan types; the network) and
5.6.6 (root and administrator steps, CI); Sonnet builders for 5.6.2 and 5.6.4 (rules written in full here, measured
work, the orchestrator's read and V1 behind them); a Sonnet agent for the scripted runs that need Claude Code (5.6.3's
import try and cross-side rules, 5.6.4's status line in a real session). Every piece lands on the landing rule (a green
climb of its task by `/usr/local/bin/bonsai` at the exact commit), check 10's Windows half, CI and the orchestrator's
read of the diff, which the run report says. **V1**, a fresh Opus verifier, on the installers (root, administrator, CI),
the machine's tripwires (5.6.1's and 5.6.2's refusals) and the machine settings reader the guard shares (note 5.6.1, 1);
**the 5.6 end verifier**, fresh Opus, on "5.6 done". Under (A) 5.6.3 needs no verifier of its own: it adds warnings, an
empty file and deny rules (each tried once in a session), and changes no guard, stop gate or ladder code; the end
verifier breaks its warnings and the Windows copy and reads its rules. Under (B), not chosen, V1 would also read 5.6.3's
guard change, and so wait for 5.6.3 to land.

#### Where each inherited finding is settled

The gate report's section 5, its 5.6 list; then the outline's "Settles", and what 5.1 to 5.5 and Rohan's 9 Oct decisions
hand on:

| Finding | Settled in | How |
|---|---|---|
| "Nothing new from the skeleton" (gate §5, 5.6) | This section | Nothing to settle |
| The stranded-folder warning, from 5.1's split (outline; 5.1.6: "the stranded machine folder goes to 5.6"; `checkLater`'s `stranded`; this plan's "Choices made") | 5.6.1 | A `check` warning with the old path, what stayed there and a person's exact lines (note 5.6.1, 5); `status`'s text too |
| The writes behind `status_writes` (outline; 5.1.5: "`bonsai settings set` writes them in 5.6"; "Stale or in tension": "5.1 reads the machine settings; 5.6 writes them") | 5.6.1 | `settings set`, all or nothing, refused in an agent session (notes 5.6.1, 3-4) |
| The writes behind attached labels (outline; 5.1.5: "read from the machine folder's `labels/` (attaching is 5.6)") | 5.6.2 | `labels attach` and `detach` (note 5.6.2) |
| `settings set` and `labels` "refuse in an agent session ... and the guard refuses them too" (outline, "Risks"; 5.3.2 rule 7: "5.6's commands refuse themselves too") | 5.6.1, 5.6.2, V1 | The commands refuse themselves with the variable set, the guard is the second layer; stated for Rohan with how it fits "agents manage Bonsai" (his part) |
| `status --line` "is 5.6's" (5.1.6 note; the flag's `Later: "step 5.6"`); what it shows "is set in 5.6's section, and it may read 5.2's asks and 5.4's ladder results" (outline) | 5.6.4 | Its parts and order, `bonsai.line/1`, its speed budget (note 5.6.4) |
| `--line` "runs on every statusline refresh, so its speed is measured" (outline, "Risks") | 5.6.4 | Budgets on both sides, measured with part 5's harness; a cache only if over (note 5.6.4, 4) |
| The personal memory layer and its check (spec §14's 5.6 row; "Stale or in tension": §6's split) | 5.6.3 | The empty index, five warnings, who writes it (note 5.6.3) |
| The Windows personal copy "older than WSL's" (spec §10), "in no 5.1 piece" (5.1.6's run report) | 5.6.3 | `personal_from`, the copy refreshed by `init` and `update` on Windows, the `personal-copy` warning (note 5.6.3, 4) |
| The secret scan of memory notes, in 5.2 on the redactor's patterns ("Choices made") | 5.6.3 | The personal layer scanned with the same `redact.Find`, a warning there (note 5.6.3, 3) |
| `install.json`'s keys `path`, `version`, `sha256` (5.1.6's run report); the `bonsai-path` finding "when `install.json` exists ... until 5.6's installer writes it" (5.1.6 note) | 5.6.6 | Both installers write exactly those keys, held by a test against the reader (note 5.6.6, 1) |
| "The two places are Go constants, their one home, which 5.6's installers and `check` read too" (5.3.6 note 2) | 5.6.6 | A test holds both scripts to the constants (note 5.6.6, 1) |
| Design 7, "the guard checks its own path and SHA-256 against `install.json` (5.6)": folded into 2 (5.3.6 note 1) | Unchanged | The hook lines never read `install.json`; it is the tripwire `check` and `status` read (note 5.6.6, 6) |
| `check`'s comparison "waits for `install.json` (5.1.6; spec §3)"; "the pre-release has no `install.json` until 5.6's installer" (5.4) | 5.6.6, 5.7 | Written at the next real install, which is 1.0's (5.7); in Bonsai's own repo `check` keeps its note until then |
| The import line in `~/.claude/CLAUDE.md` "is a person's step (no test touches a real home)" (outline); "5.6's import line in it is the person's, 'No agent edits that file'" (5.5.2 note 2) | 5.6.3, 5.7 | Its form measured in a scratch project (note 5.6.3, 5); his line moves to 5.7's 1.0 install batch, when the index exists on his machine, and an agent then checks that it loads ("Hand-offs to 5.7") |
| "One real install proof only if 5.6's section asks for it" (outline) | This section | None asked: the root and administrator halves run on GitHub's throwaway machines, the Windows `RunAs` child forced there by a test-only switch (note 5.6.6, 4); only the UAC prompt itself waits for his 1.0 install |
| "The installers write root-owned and admin paths, so agents test them only against scratch targets" (outline) | 5.6.6, V1 | Scratch targets in tests; the real places only on CI's throwaway machines; no agent runs an installer at a real place on this computer |
| A fresh verifier for the installers (outline) | V1 | After 5.6.6 has landed with its CI job green, beside the machine's tripwires |
| No self-update; "from 5.6 `status` and `check` say when a newer release exists and an agent hands him the lines" (Rohan, 9 Oct, 15:46; spec §3's note) | 5.6.5 | Read by `status --full` only, offline after; the lines written by Bonsai for both sides (note 5.6.5) |
| Agents manage Bonsai inside projects; the program stays his install (Rohan, 9 Oct, 15:35) | 5.6.1, 5.6.2, 5.6.5, 5.6.6 | Machine settings and attached labels are the machine's, set by a person or the studio's registration; agents read them and hand over the line; agents hand over install lines and never install |
| The operating skill names every word; a new word changes it in the same commit (5.5.3 note 4; 5.5.7's `CLAUDE.md` rule) | 5.6.0, 5.6.3, 5.6.5 | Its lines for `settings` and `labels`, for the personal layer and for the install lines |
| No screens (Rohan, 9 Oct) | 5.6.4 | `--line` prints text and stops; the studio draws Bonsai's part from `bonsai.line/1` (his part, "not a screen") |
| Rohan's (a): the hook lines name the installed places, never the PATH (5.3) | 5.6.6 | The installers place the binary exactly there; no project changes at an install |
| Rohan's (ii): `--allow-exec`, `unlink` and changes to Bonsai's own lines are a person's where the studio does not manage (5.3) | 5.6.5 | Unchanged: an installer touches no project; a release that changes Bonsai's lines shows as `check`'s finding, its step the person's line (note 5.6.5, 6) |
| A pack's version, a pack added or taken out stay his step; agents hand over the line (Rohan, 9 Oct, 18:32, approving 5.5) | This section | The same shape for machine settings, attached labels and installs: the agent prepares and hands over the exact line |
| The studio's registration "runs `bonsai labels attach` and `bonsai settings`, both 5.6's"; "full registration waits for 5.6" (this plan's "Stale or in tension"; contract §15.2) | 5.6.1, 5.6.2 | Built; both refuse in an agent session, so the registration is run by a person or the studio's own program (a note for the studio's plan, step 7) |
| A later install only when a part changes the guard, the stop gate or the ladder (spec §17 step 8; 5.4) | This section | Under (A), 5.6 changes none of them (the guard's rule 7 is 5.3's; `--line`, `settings` and `labels` are no hook), so 5.6 asks no install; but for the machine settings reader the guard shares, which now refuses a `status_command` outside its rule (note 5.6.1, 1): that reaches the installed guard with 1.0, and nothing waits on it, since Bonsai's own repo has no machine settings. Under (B), not chosen on 10 Oct, 5.6.3 would change the guard's rule 5, so one WSL install (note 5.6.3, 2) |
| The floors at a part's end (5.4, "Tasks and names") | After 5.6.5 | One climb of `main`, the floors raised to its counts under a `T-569x` task, the numbers in Rohan's last line |

#### Notes per piece

**5.6.0, the formats set and the two words' places.**
1. **The set:** one commit to `formats/` with its manifest, at the set after the last landed (set 5 is 5.2.0's, set 6
   5.4.0's, then 5.5.0's; so set 8 if those land as planned), as `formats/README.md`'s "How the set changes" asks;
   additions only, so the schema-compare test passes. Like every later part's, it follows 5.2.0's set 5.
   - **`bonsai.settings/1`**, the `--json` of `settings show` and `set`: `format`; `scope` (`workspace` or `machine`,
     closed); `file`, the settings file's place (forward slashes; printed locally, never committed); `settings`, one
     entry per key of the scope, each `key`, `value` (null when unset) and `from` (`file` or `default`, closed);
     `changed`, the keys `set` changed (`[]` for `show` and for a set that changed nothing); `error`.
   - **`bonsai.attach/1`**, the `--json` of `labels attach` and `detach`: `format`; `command` (`attach` or `detach`);
     `namespace`; `version` (the definitions' own, or null); `file` (the machine folder's file, forward slashes);
     `result` (`attached`, `updated`, `unchanged`, `detached`, `nothing`; closed); `labels`, the names attached after
     the command (`[]` after a detach); `error`.
   - **`bonsai.line/1`**, the `--json` of `status --line`: `format`; `workspace` (`id` and `name`, or null when the
     folder is not linked); `task` (`id`, `title`, `status`, or null); `running` (how many tasks read `running` when no
     one task is active, else null); `ladder` (`green`, `red`, `old`, `none`, or null with no task; closed); `verify`
     and `asks` (counts); `release` (a newer release's version, or null); `line` (the text, `""` when not linked);
     `error`.
   - **`status` gains `bonsai_release`** at the end of its properties: `version` (the version compared, or null),
     `newest` (the newest release read, or null), `state` (`current`, `newer`, `unknown`, `not-a-release`; closed),
     `read` (when last read successfully, RFC 3339 UTC, or null), `why` (one line when `unknown`, else null), `lines`
     (`linux` and `windows`, each a list of command lines, when `newer`; else null; `linux` because `install.sh` serves
     Linux and WSL alike).
   - **The memory schema's description** says the personal layer has its own caps, named on the reference page (their
     one home is note 5.6.3, 3's table, so the description holds no number), and who writes the layer as Rohan chose
     (note 5.6.3, 2); the project's notes keep their writer, base's `memory` skill (5.5's "Stale or in tension").
   - Not given a schema: the machine folder's and the home's `settings.json`, `install.json`, the home's release record
     (note 5.6.5, 2), the Windows copy's record (note 5.6.3, 4) and the line's cache if one is needed: each is Bonsai's
     own file on one machine, read by Bonsai alone, as `workspace.json` is (5.1.6), and documented in the Go file that
     writes it.
2. **The words:** `settings` and `labels` in the registry at their rows of spec §4's table, each with its sub-words
   (`settings show`, `settings set`; `labels attach`, `labels detach`) marked `Later` with their piece, their full flag
   tables, exit codes and examples, so `--help` and `bonsai --help --json` show the final shape at once; the error words
   both need, in `format.ErrorWords` with their `who` (`namespace-taken` and `label-redefined`, both a person's; the
   rest are words that exist: `bad-value` for an unknown key or a bad value, `not-linked`, `read-failed`,
   `write-failed`, and 5.3.6 note 8's word for a person's command in an agent session); and 5.3.6 note 8's check of the
   variable made callable from any word's run function if it is not, so 5.6.1 and 5.6.2 each call it and neither writes
   its own. **That function also holds the stamped build's confinement for these words:** a stamped scratch build skips
   the refusal only while Bonsai's home (`BONSAI_HOME`, else `~/.bonsai`) is inside its stamped scratch root (5.3.6 note
   2's confinement, carried from the project to the home), so scripted runs on scratch homes work and a stamped build
   never writes a real home's settings; a plain build never skips; a test holds both. Its refusal's sentence says where
   to type the command instead: "type it in a terminal of your own, not inside Claude (its `!` lines count as an agent's
   session)". Being 5.6.0's, it lets 5.6.1 and 5.6.2 run side by side. The operating skill gains, in the same commit,
   its lines for both (read with `settings show`; `settings set` and `labels` are a person's: hand over the exact line),
   as `CLAUDE.md` asks from 5.5.7, and its test passes. The reference page is regenerated.
3. **`status` writes `bonsai_release: null`** until 5.6.5 fills it, its test naming the field and that step as 5.1's
   `notBuiltYet` did; `--line` stays refused as not built until 5.6.4.

**5.6.1, `settings`, the stranded folder and the cache.**
1. **The keys, in one Go table** (their one home; the reference page, `settings show` and `--help` read it). For the
   workspace, in its machine folder's `settings.json`: `status_writes` (`agents` or `command`; default `agents`) and
   `status_command` (with `command` only: one line of at most 100 characters of letters, digits, spaces and `-_./:`,
   because the guard prints it in every `command`-mode refusal and agents run it as told; default none). For the
   machine, with `--machine`, in the home's `settings.json`: `cache_keep_days` (a whole number of days from 1 to 3650,
   or `none`; default `none`, spec §6) and `personal_from` (on Windows only: WSL's personal folder as Windows reaches
   it, such as `//wsl.localhost/<distro>/home/<user>/.bonsai/personal`, or `none`; never the home's own `personal/` or
   a folder inside it, its own target; read by 5.6.3). **The reader applies the same rules, from the same table:**
   `LoadMachineSettings` (`internal/workspace/machine.go`) refuses a `status_command` outside its rule as it refuses a
   file it cannot read, an error naming the file and the key, so a forged or hand-edited file never puts `;`, `$(...)`
   or a newline into a guard's refusal or an agent's next command. The guard meets that error on its existing
   fail-closed path (a call whose judgment needs the mode is refused, naming the file; 5.3.2's reads); `check` and
   `status` report it as they report an unreadable settings file. This changes the reader the guard uses, so it reaches
   the installed guard only with 1.0's install (5.7); Bonsai's own repo has no machine settings, so nothing waits on
   it. V1 reads it (guards and hooks).
2. **`settings show [--machine] [--json]`**: every key of the scope with its value in force and where it came from (the
   file, or the default), and the file's place. Agents may run it: it never refuses in an agent session. Without
   `--machine` it needs a linked checkout and reads the main checkout's machine folder from a worktree too. Exit codes:
   0; 2 (a bad flag or argument, `bad-flag`); 3 (a settings file it cannot read, or one holding a value outside its
   rule, `read-failed`, naming the file and the key); 4 (outside a linked checkout without `--machine`, `not-linked`).
3. **`settings set k=v [k=v ...] [--machine] [--json]`**: every value checked first, then all or nothing; an unknown key
   or a value outside its rule exits 2 (`bad-value`), naming the key's allowed values or listing the keys;
   `status_writes=command` needs a `status_command`, given or already set; `status_writes=agents` drops
   `status_command`; `personal_from` is refused off Windows and when it names the home's own `personal/`. The file is
   written byte-stable (two-space indent, LF, ASCII, the table's order, a key this Bonsai does not know kept as read),
   staged and renamed with Windows' busy retries; a value already in force writes nothing (exit 0, `changed: []`). It
   asks nothing and takes no `--yes`: each `k=v` is the change, typed. Exit codes: 0, 2, 3 (could not write), 4 (not
   linked; an agent session).
4. **In an agent session it refuses itself** (spec §4: "a tripwire, decision D"; 5.3.2 rule 7 is the guard's second
   layer): with `CLAUDE_CODE_CHILD_SESSION` set, `settings set` exits 4 with 5.3.6 note 8's word (`who: person`) and
   writes nothing, through 5.6.0's shared function (its stamped-build confinement is there, note 5.6.0, 2); its
   `next.do` is the same command line, and its sentence says "type it in a terminal of your own, not inside Claude (its
   `!` lines count as an agent's session)". The variable is read as Claude Code sets it in every process its tools start
   (5.3.6 note 8), so a script or subprocess inherits the refusal. **What still gets through,** as spec §4's "a
   tripwire" says: a shell, `cp` or interpreter writing the settings file directly (the guard judges file tools, not a
   shell's writes, 5.3.3 note 5); a script that unsets the variable before calling `bonsai settings set` (the guard
   refuses only an unset it can read on the shell line, rule 7); a stamped build whose scratch root covers the real
   home. These stop an agent's mistake, not a determined agent; the studio's bridge flags a lost `command` setting
   (contract §10.6). V1 and "5.6 done" item 4 try each and record it. Go tests set and clear the variable with
   `t.Setenv`; no scripted run unsets it in a shell line (the installed guard refuses such a line, rule 7).
5. **The stranded folder** (contract §3, "Moved checkouts"; spec §6's warning): `check` warns `stranded`, never the exit
   code, for each other machine folder whose record (`workspace.json`) holds this `bonsai.yaml`'s id as its last, whose
   recorded path holds no `bonsai.yaml` at all (the folder moved or gone), and which holds a `settings.json` or attached
   labels. A recorded path that holds a `bonsai.yaml` with another id is not stranded: that folder is the machine folder
   of the checkout now at that path, whose own `check` gives `id-changed` (5.1.6; contract §3), and this rule leaves it
   alone. The sentence names the old path and what stayed there (`status_writes`, its command, each attached namespace)
   against what this checkout has now; `next.do` (`who: person`) gives, in order, the exact commands that bring them
   over and then clear the old folder: `bonsai settings set status_writes=command "status_command=<its command>"` (only
   when the old file passes the reader's rule, note 1; else its path, to read), `bonsai labels attach '<old
   folder>/labels/<namespace>.yaml'` for each, and the old folder's removal (`rm -r '<path>'` on Linux, `Remove-Item
   -Recurse -LiteralPath '<path>'` on Windows); or the studio's registration again (contract §15.2). **The removal's
   path is built only from the folder's own name,** once verified as `r-<16 hex>` directly under `<home>/workspaces/`,
   never from the record and never the home or `workspaces/` itself; a folder of any other name gets no removal line.
   **Every path in the lines is single-quoted for its shell** (a home under `%USERPROFILE%` may hold a space; a `'`
   inside is doubled for PowerShell and written `'\''` for bash). A folder holding only its record loses nothing and
   gives nothing. `status`'s text names it too (contract §3: "`bonsai check` and `status` report the stranded folder");
   `status --json` carries no warning (5.1.6: `check`'s warnings are never `problems`), and the lost setting shows in
   its `status_writes`, which the studio's bridge already flags (contract §10.6). `stranded` moves from `checkLater`
   into `format.CheckWords` with its case in `TestCheckTable`; the `next.do` test accepts `labels attach` as a word the
   registry lists (5.6.0), and by 5.6's end it is built.
6. **The cache, cleaned in `update`** (spec §6: "the cache in `update`"): with `cache_keep_days` set, `update --yes`
   (never a preview, never `check` or `status`) removes, after its own run and never changing its exit code or result,
   each pack clone in `cache/git/` whose last use is older than that many days, never one this project's lock or
   `bonsai.yaml` names. A clone's last use is a mark the fetch touches whenever it fetches or reads it, so a clone
   another project is using now is never older than a day. The `--adopt` copies in `cache/adopted/` are a person's saved
   edits and are never cleaned ("Stale or in tension", below). What it removed is one line in `update`'s text: no
   `clean` record, since a clone is a refillable copy outside every project and a `clean` record's `target` is
   project-relative (contract §8.2).
7. **Proof** beyond the table: on `t.TempDir()` homes, two projects sharing clones, one clone past its days and in
   neither lock (removed), one past its days and in a lock (kept), an `--adopt` copy years old (kept); a moved fixture
   checkout (its machine folder written, the folder renamed, `init` there): `stranded` with both lines, which, run with
   the variable cleared, bring the settings and labels over and clear the warning; a recorded path now holding a
   `bonsai.yaml` with another id (no `stranded`); a record whose folder name is not `r-<16 hex>` (no removal line); a
   home path with a space and a quote (the lines run as printed, on both sides); an `update` preview with
   `cache_keep_days` set (nothing removed). The reader: a `status_command` holding `;`, `$`, a backtick, a newline or
   past 100 characters, written straight into a scratch home's file, refused by `LoadMachineSettings`, and the guard's
   test refusing a call that needs the mode, naming the file.

**5.6.2, `labels attach` and `detach`** (contract §5.1-§5.3).
1. **`labels attach <file> [--json]`**, in a linked checkout (`not-linked` otherwise): the file read as
   `bonsai.labels/1` by 5.1.4a's reader (a refusal names the field: `bad-value`); a namespace that is `bonsai` or a
   locked pack's refused (exit 4, `namespace-taken`, `who: person`; contract §5.1: "Bonsai refuses the attach"); copied,
   line endings made LF, to `<machine folder>/labels/<namespace>.yaml`, staged and renamed.
2. **Attached again** (contract §5.3: "the bridge re-attaches on start when its definitions' `version` is newer"): the
   same content is `unchanged`; a newer `version` that keeps every attached definition as it is and only adds is
   `updated`; anything else, a definition changed or removed, or different content at the same or an older version, is
   refused (exit 4, `label-redefined`): "a definition may be added, never redefined" (contract §5.1). Its `next.do`
   names `labels detach <namespace>` and then `attach`: a person's deliberate step.
3. **`labels detach <namespace> [--json]`**: the file removed (`detached`), or `nothing` when there is none. It also
   runs in a checkout whose `bonsai.yaml` is gone, so a project taken out can be unregistered.
4. **In an agent session both refuse themselves**, through 5.6.0's shared function, exactly as `settings set` does
   (note 5.6.1, 4).
5. **What an attach reaches, already built, tried once by the end verifier:** `status --json`'s `labels` lists the
   namespace `from: machine` (5.1.5); `check` holds values to it (5.1.6); `hook start`'s opening context shows it (5.2.4
   note 9); in `command` mode the guard refuses an agent's edit setting a label whose definition reads `set_by: outside`
   (5.3.2 rule 6).

**5.6.3, the personal memory layer** (spec §10; contract §7.4).
1. **Its place:** `<home>/personal/INDEX.md` and `<home>/personal/notes/`, in `bonsai.memory/1`, never in a project,
   never forwarded (contract §2.6). `init` and `update` write an empty index when there is none, only when they write
   (`--yes`, whether or not the project changed), never in a preview and never from `check` or `status` (`format:
   bonsai.memory/1` with its pointer comment, `id: null`, `title: Personal memory`, `kind: index`, `updated` today,
   `source: null`, `labels: {}`, and one body line saying what goes there and who writes it), never over an existing
   one; `unlink` leaves it (the machine's, not the project's).
2. **Who writes it: a person (Rohan's choice, 10 Oct: (A))** (his part, "Who saves notes about you").
   - **Under (A), recommended and chosen: a person.** 5.3.2 rule 5 and 5.3.5's `Edit(~/.bonsai/**)` refuse an agent's
     file-tool writes anywhere in the home, in every linked project, and that stays: the index loads into every session
     in every project on the machine, so a note one agent wrote would carry its words into all of them, as a write to
     `~/.claude/CLAUDE.md` would (which base walls, 5.5.2 note 2). An agent drafts a note about the person with base's
     `memory` skill and hands it over; the person saves it by hand, or in a session opened outside any linked project. A
     fact about how the person wants work done in one project goes into that project's memory, which agents write.
     Base's `memory` skill and the operating skill say so (this piece edits their lines; `check --pack` and the
     operating test pass). As with the machine's settings, the walls stop a mistake, not a determined agent (a shell
     write gets through, 5.3.3 note 5). What lost: a note about the person is not saved until they save it.
   - **Under (B), not chosen on 10 Oct: agents write it, through base's `memory` skill.** What would change:
     - **The guard,** 5.3.2 rule 5 (`internal/guard/`, owned by 5.6.3): an agent's file-tool write in the home is
       allowed for `<home>/personal/INDEX.md` and `<home>/personal/notes/<name>.md` only (a plain name ending `.md`, no
       deeper folder), and still refused everywhere else in the home; tests both ways. V1 reads it (guards and hooks),
       so V1 waits for 5.6.3 to land.
     - **The engine's deny rules:** `Edit(~/.bonsai/**)` and its cross-side twins (item 6) replaced by rules over the
       home's other entries (`workspaces/**`, `settings.json`, `install.json`, `salt`, `cache/**`, `locks/**`), each
       tried once in a session on both sides; an entry a later home adds is then walled by the guard alone.
     - **Base's `memory` skill and the operating skill:** agents save notes about the person in the personal layer,
       within its caps (item 3), one fact a note, never a secret, and say in their report what they saved.
     - **An install:** the guard changed, so a pre-release is built once 5.6.3 has landed, its fingerprint committed in
       a run report, and installed by Rohan in WSL (5.4's four lines, about 5 minutes, his password) before V1; the
       inherited row on later installs says so.
     - **Hours:** 5.6.3 3-5, so 5.6 14-22, re-ask 29 (22 x 1.3 = 28.6), changed together wherever they are quoted:
       this section's heading, header and "Hours"; his part ("Who saves notes about you", "Hours, order and your
       steps"); the pieces' table total; the plan's header line and top table (5.6's row, and the step's 139-218,
       which becomes 140-220, with the figures "Size" derives from it); the outline's 5.6 heading; "Stop lines" item 1
       (29; 22 x 1.3 = 28.6); "5.6 done" item 14.
     - **His list:** "In 5.6" gains the install (5.4's four lines, about 5 minutes, your password); "Your own time"
       gains 5 minutes at 5.6; "Stale or in tension"'s item on §10's memory skill reads "as the spec has it, in base's
       skill".
3. **Its check,** warnings only, never the exit code (the layer is the machine's, not the project's, so a project's CI
   and rung 0 must not fail on it), each `who: person`, read from the home whatever project `check` runs in:
   `personal-format` (the index or a note not `bonsai.memory/1`, or the index's `kind` not `index`); `personal-budget`
   (the index over 40 lines or 4 KB; a note over 4 KB); `personal-index` (an index line naming a note that does not
   exist, or a note no index line names); `personal-secret` (anything 5.2's `redact.Find` finds, naming the file, the
   line and the kind of secret, never the value: 5.2.4 note 11's rule for the layer that loads everywhere);
   `personal-copy` (item 4). With no `personal/` folder, nothing. **The caps' one home** is a Go table of memory caps:
   the project index's 120 lines and 12 KB, the personal index's 40 lines and 4 KB, and 4 KB a note in either layer;
   today's `MaxIndexLines`, `MaxIndexBytes` and `MaxNoteBytes` (`internal/engine/checkdocs.go`) become its rows. `check`
   reads it, 5.1.10's reference page lists it, a test in `internal/engine/` holds base's `memory` skill's numbers to it,
   and the memory schema's description names no number (note 5.6.0, 1). Why 4 KB for the personal index, not the
   project's 12 KB: 40 lines at the project index's own rate (120 lines in 12 KB, about 100 bytes a line), so a 40-line
   index of ordinary lines fits and a few overlong lines cannot hide under the line count.
4. **The Windows copy** (spec §10: "WSL's is the canonical one; the Windows home holds a copy, refreshed when a
   Windows-side project links ..., and `check` there reports a copy older than WSL's"): on Windows, once a person has
   named WSL's folder in the home setting `personal_from` (5.6.1; at Mimas's link, step 8), `init --yes` and `update
   --yes` (never a preview, never `check` or `status`) compare the copy with the source and refresh it when it differs.
   They copy only `INDEX.md` and `notes/*.md`, regular files, following no link, each at most 64 KB and at most 500
   notes (past either, nothing is copied and their text says why); written whole into the home's `personal/`, a note
   the source no longer has removed, staged and renamed with busy retries; one line in their text. A source that is the
   home's own `personal/`, or inside it, is refused (its own target). Each refresh records its outcome in
   `cache/personal-copy.json` (the source, each copied file's SHA-256, when, the last failure and its reason). "Older"
   is read as "not the same": the source is the canonical one, so any difference is the copy's. **`check` never reaches
   WSL:** a read through `\\wsl.localhost` can start WSL and take up to 2 s, which would land on every `check` and every
   rung 0. So `check` warns `personal-copy` from the record and the copy alone: never refreshed since `personal_from`
   was set, the last refresh failed (when, why), or the copy changed since that refresh. A change on WSL's side shows
   at the next `update`. In a refresh, each read of the source has 2 s; past them, a note (not compared, not
   refreshed, recorded as a failure), never a wait.
5. **The import line's form, measured** (spec §10: `@~/.bonsai/personal/INDEX.md`, "imported once per machine from the
   user memory file"): the builder reads Claude Code's memory reference on the version in use (an import with `~`, how
   deep imports go, what a missing file does, when the external-import prompt appears), and the Sonnet agent tries the
   form that will be used, in a scratch project only, through `claude-here` in an interactive `tmux` session (5.5.5's
   method): the project's own `CLAUDE.md` importing `@~/bonsai-checks/<scratch>/INDEX.md` (the `@~/` form, a file
   outside the project, as the real line's is) and a missing one, `@~/bonsai-checks/<scratch>/missing/INDEX.md`, read
   with `/memory`; the real `~/.claude` is never touched. The run report records the version, what loaded, what the
   missing file did, and any external-import prompt with its words (spec §10 expects the user memory file to avoid it;
   a project's file may not). Rohan's own line waits for 5.7 ("Hand-offs to 5.7"): the index exists on his machine only
   at 1.0, and an agent checks there that it loads.
6. **The other side's Bonsai home.** The engine's own deny rules (5.3.5's `Own` lines; not the guard, so no new
   install) gain the twins of `Edit(~/.bonsai/**)` for the other side's home, as base's walls have theirs (5.5.2 note
   2): `//mnt/c/Users/*/.bonsai/**` (a WSL session reaching the Windows home), and `//wsl.localhost/*/home/*/.bonsai/**`
   and `//wsl$/*/home/*/.bonsai/**` (a native Windows session reaching WSL's), in the forms 5.5.6 found to hold (a third
   leading slash where it found that). Each is tried once in a scratch session on its side by 5.5.6's method: a decoy
   rule of the same form over a scratch folder (`.../bonsai-checks/<scratch>/.bonsai/**`), loaded with `--settings` for
   the try only, a Write and an Edit refused with the rule named; never tried against a real home. A form that holds on
   no side is left out and written in the run report, as base's README does for its twins. Why: the Windows copy (item
   4) makes WSL's home a source of what loads on Windows, and a session on either side could otherwise write the other
   side's machine settings and personal layer by that path. Named for step 8 (Mimas) too ("Hand-offs").
7. **Proof** beyond the table: each warning from a fixture home on both sides; a secret-shaped decoy (made up) found,
   its value in no output; `init --yes` twice writes the index once, and an `init` or `update` preview and `check`
   write nothing in the home (hashed before and after); a Windows test whose `personal_from` is a scratch folder
   (refreshed by `update --yes`; the copy then edited, and `check`'s warning from the record; the source changed, and
   the next `update` refreshing it), one that never answers (a note within 2 s, recorded, then `check`'s warning), and
   a source holding a link, a 65 KB note, or naming the home's own `personal/` (each skipped or refused as item 4
   says); the caps' table read by `check`, and the skill's numbers held to it; the cross-side rules' tries in the run
   report.

**5.6.4, `status --line`** (spec §4: "the statusline's workspace half"; its reference, read only with `git show`, the
studio's `tools/statusline/README.md` and `statusline.mjs` at `7017d63`).
1. **What it prints:** one line of ASCII, at most 100 characters, its parts in this order, each left out when it says
   nothing: the active task (contract §13's function, 5.1.5, with `BONSAI_TASK` from the environment Claude Code gives
   the status-line command) as its id and title, the title cut to 30 characters with `...`; `N running` when two or
   more tasks read `running` and none is named (the function's own answer, contract §13), so the line never hides them
   behind `no task`; else `no task`;
   `ladder:green`, `ladder:red`, `ladder:old` (its commit is not HEAD) or `ladder:none`, for that task's last local
   result (5.4); `verify:N`, the tasks reading `verify` (N above 0); `asks:N`, the open asks waiting for a person
   (5.2.5; N above 0); `bonsai:vX.Y.Z` when a newer release is recorded (5.6.5 adds it). It is the studio's own task
   part (`T-0042 Hover panel ladder:green verify:2`, read today from `studio/tasks/*.md` and the ladder's results), with
   the asks Bonsai now holds.
2. **It never breaks a status line:** it runs in the folder it is started in (where Claude Code starts a status-line
   command is measured, and the help says what a person's script does if it is not the project); it never reads stdin,
   so a script that calls it never waits (a test runs it with its input held open and sees it return within its
   budget). Outside a linked checkout it prints nothing; a `bonsai.yaml` it cannot read
   prints `bonsai.yaml unreadable`; either way exit 0. Only bad flags exit 2 (`--line` with `--full` or `--active`).
   Text from task files is stripped of control characters and anything not ASCII, so a task title cannot send escape
   codes to a terminal (the studio's line has the same rule).
3. **`--line --json`** prints `bonsai.line/1` (5.6.0), the parts as fields with the line, so the studio's status line
   draws Bonsai's part in its own colours. Bonsai writes no `statusLine` setting anywhere; `--help` shows the line a
   person adds to their own Claude settings to use it alone.
4. **Speed, measured** (it runs at every status-line refresh): no git process, no Claude Code, no network; HEAD, for
   `ladder:old`, read from git's files (a worktree's `.git` file followed). Budget, with part 5's harness (gate report
   §2.5), on a fixture project of 300 task files, two years of ask day files (five open) and a ladder result: p95 at
   most 25 ms on WSL's own disk and at most 60 ms natively on Windows (the studio's Node line: 54-63 ms, about 80 ms
   through Git Bash). Measured, not budgeted: from WSL on a project on a Windows drive (`/mnt/...`), and through Git
   Bash as Claude Code runs it on Windows. Over budget, a cache in the home, `cache/line/<machine key>.json`, keyed by
   the files' sizes and times (refillable; never a write in the project; a write that fails, a read-only home say,
   never fails the line); still over, the orchestrator has the numbers before the piece lands.
5. **Seen once in a real session,** by the Sonnet agent: on WSL, a scratch project's `.claude/settings.local.json`
   naming the stamped build's `status --line` as its status line, an interactive session in `tmux` through
   `claude-here`, the pane captured with the line in it, the Claude Code version and the user settings hashes before and
   after recorded. On Windows the timing through Git Bash stands in (an interactive Windows session needs a person).
6. **Hours:** 4-6 of the row's 5-8; 5.6.5 takes the rest ("Hours", above).

**5.6.5, a newer release, known and handed over** (Rohan, 9 Oct, 15:46; spec §3's note).
1. **What is read:** the tags of Bonsai's public repository, `https://github.com/LastStep/Bonsai.git`, with the
   machine's git (`RemoteTags`, `internal/engine/newer.go`): no login, no API token, no service of Bonsai's. Only plain
   `vX.Y.Z` tags at or above 1.0.0 count: the old product's `v0.x` tags and `base-v*` pack tags do not. The source is a
   Go constant; a stamped scratch build may name a local bare repository instead (5.3.6 note 2's stamps; a plain build's
   stamp empty, held by a test), so no test reaches GitHub.
2. **When, and how often:** only `bonsai status --full` goes online, as it already does for packs. It reads at most once
   a day after a read that worked and once an hour after one that failed, each read given 10 s, beside the packs' reads.
   The answer goes into the home's `cache/release.json` (the newest release, when it was read, the source, and the last
   failure with its time and reason): a refillable copy (deleting it costs one read), Bonsai's own file, documented
   where it is written. A write that fails (a read-only home, say) never fails `status`: the answer is still printed,
   and one line says it was not kept. Plain `status`, `check` and `--line` read only that file, never the network.
3. **What is compared:** the installed copy's version from `install.json` (5.6.6), else the running binary's. A build
   that is not a release (`dev`, or a pre-release such as 5.4's) is `not-a-release`, which still names the newest
   release when one exists.
4. **What is said,** from the record: `status --json`'s `bonsai_release` (`current`; `newer`, with the newest version
   and the lines; `unknown`, never read or the last read failed, with when and why; or `not-a-release`); one line in
   `status`'s text; a `check` warning, `bonsai-newer` (`who: person`), its `next.do` "a person installs Bonsai vX.Y.Z on
   each side: the lines are in `bonsai status --full --json`, `bonsai_release.lines`"; when the record is missing or
   older than 7 days, a `check` note naming `bonsai status --full`, never a warning; `--line`'s `bonsai:vX.Y.Z`. No exit
   code changes and nothing refuses to run for any of it.
5. **The lines,** written by Bonsai from one Go table, so every agent hands over the same, for both sides at once: bash
   for WSL and PowerShell 5.1 for Windows, one command per line, no `&&` (Rohan's part shows them for 1.0.1). **Built
   only from a version that parses as three whole numbers, `X.Y.Z`** (digits and two dots, nothing else), checked when
   the tags are read and again whenever the record is read: a record whose `newest` is anything else (a forged or broken
   `cache/release.json`, a command in it, say) gives no lines and state `unknown`, its `why` naming the file. Every URL
   is the fixed prefix `https://github.com/LastStep/Bonsai/releases/download/v`, then that version, then a fixed file
   name; nothing in a line comes from the record but the checked version. The archive and checksum names follow
   `.goreleaser.yaml` today (`bonsai_<version>_<os>_<arch>`, `checksums.txt`), the architecture this machine's; a test
   in this piece reads the two templates from `.goreleaser.yaml` as text and fails if the lines' names differ, so a
   change there in 5.7 cannot leave the lines behind. The installers sit at the archive's top beside the binary, which
   5.7 makes so (below).
   The operating skill's "installing the `bonsai` program" line points at these lines (this piece edits it): the agent
   hands them over, with the release's notes link, and never runs them.
6. **After an install,** nothing in a project changes (5.3's (a)). If a new release changes Bonsai's own hook lines, its
   `check` gives the finding for lines out of date (5.2.4), whose step is a person's line under 5.3's (ii) where the
   studio does not manage the project.

**5.6.6, the installers and `install.json`** (spec §3; 5.3's (a)).
1. **Two scripts in `install/`:** `install.sh` (POSIX `sh`, for Linux and WSL) and `install.ps1` (Windows PowerShell
   5.1), each documented in its header (what it does, when to run it, every option with an example, the way out), ASCII
   only, LF through one `.gitattributes` line. Each installs the `bonsai` beside it (the archive's top, 5.7). A Go test
   in `install/` holds each to the two installed places as 5.3's constants spell them and to `install.json`'s three keys
   as `check`'s reader reads them (`internal/engine/checkmachine.go`), so neither has a second home.
2. **`install.sh [--target <file>] [--remove]`**, run by the person as themselves, never under `sudo` (it refuses as
   root, since `~` would then be root's home and `install.json` land there): checks that the `bonsai` beside it runs
   (`--version`) and takes its SHA-256; refuses first a target folder that group or others can write; installs it under
   a temporary name in the target's folder (`.bonsai.new.<pid>`), with `sudo install -o root -g root -m 0755` when the
   folder is not the person's to write (the real place, `/usr/local/bin/bonsai`, the password once), else with `install
   -m 0755` (a scratch target); checks the temporary copy's SHA-256 against the source's (a source changed meanwhile is
   caught here, in a folder only root writes, and the copy removed); then renames it over the target (`sudo mv -f`, one
   rename), so a hook starting meanwhile finds the old copy or the new one, never no file, and a running `bonsai` keeps
   its old file; then checks, at the real place, that the file and its folder are root's and writable by no one else;
   writes `${BONSAI_HOME:-$HOME/.bonsai}/install.json` (`path`, `version` from the installed copy's `--version`,
   `sha256`) through a temporary file and a rename; prints the fingerprint, `which -a bonsai` (naming any `bonsai` that
   comes first, spec §17 step 3) and the installed `--version`. `--remove` takes out the installed file (`sudo rm` at
   the real place), a leftover temporary copy and `install.json`. Exit 0, or not, with one plain sentence and its next
   step.
3. **`install.ps1 [-Target <file>] [-Remove]`**, run by the person in a normal PowerShell as `powershell -NoProfile
   -ExecutionPolicy Bypass -File .\install.ps1` (Windows runs no downloaded script otherwise): checks that the
   `bonsai.exe` beside it runs and takes its SHA-256; at the real place, `C:\Program Files\Bonsai\bonsai.exe`, the
   administrator steps run in one elevated child (`Start-Process -Verb RunAs -Wait`, one UAC prompt), or in place when
   the session is already elevated (the same steps, from the same text).
   - **The child's code is inline, never a file.** It is `powershell.exe` by its full path (from
     `[Environment]::SystemDirectory`, never found by name) with `-NoProfile -NonInteractive -EncodedCommand <base64>`,
     the text built by the parent from fixed text plus three values: the source file's path, the target's path and the
     expected SHA-256, each checked first (the paths absolute, the hash 64 hex) and embedded as a single-quoted literal
     (a `'` doubled). No script in a user-writable folder runs elevated, and the child reads no file there but the one
     it copies. It sets its own working folder (the target's folder, once made) and relies on no `$env:` value the
     parent changed (an elevated child does not inherit them).
   - **What the child does, in order:** the folder made; the file copied in under a temporary name (`bonsai.exe.new`);
     that copy's SHA-256 compared with the expected one, and on a mismatch the copy removed and the child exits
     non-zero (a source swapped after the parent hashed it is caught here, inside a folder only administrators write);
     leftover `bonsai.exe.old-*` files from an earlier install removed (one still running is skipped); the old
     `bonsai.exe`, which a hook may be running, renamed aside to `bonsai.exe.old-<n>` (Windows lets a running program
     be renamed, not overwritten), with busy retries; the new copy renamed into place, so a hook starting between the
     two renames finds no file and refuses (it fails closed), never runs another; `C:\Program Files\Bonsai` added once
     to the machine's PATH, read raw (`DoNotExpandEnvironmentNames`) and written back as `ExpandString`, never through
     `[Environment]::SetEnvironmentVariable`, which writes the expanded value as plain text; the change announced to
     running programs; the folder's ACL checked (no write to Users, Authenticated Users or Everyone). Its exit code is
     the parent's answer: non-zero is a refusal naming the step that failed.
   - **A declined UAC prompt** (`Start-Process` throws) is a refusal with its next step: "the Windows prompt was
     declined, so nothing was installed; next: run the last line again and choose Yes".
   - **Then the parent, as the person,** checks the installed hash again; writes `install.json` in `$env:BONSAI_HOME`,
     else `%USERPROFILE%\.bonsai`, its path with forward slashes (`C:/Program Files/Bonsai/bonsai.exe`); prints the
     fingerprint, `Get-Command bonsai -All` in order and `--version`. A scratch target needs no elevation and changes no
     PATH, and runs the same copy, hash and rename steps in place: the PATH merge is a function the tests run on
     strings (`%SystemRoot%` entries kept, no entry twice, the separators right), so no test writes the registry.
     `-Remove` undoes each step, the PATH entry taken out the same careful way.
4. **Proof without this computer's real places.** Go tests run each script on scratch targets and a scratch
   `BONSAI_HOME` (the `.sh` on Linux, the `.ps1` on Windows, each skipped elsewhere with its reason): install; the
   record read back by `check`'s reader; a second install (the same fingerprint, the record unchanged); a second
   install while a `bonsai` from the first is still running (a `hook guard` waiting on its input, inside its budget):
   it succeeds, the running process is untouched, and on Windows the old file sits aside until the next install
   removes it; remove; each refusal (a `bonsai` beside it that does not run; a target folder others can write, on
   Linux; `install.sh` as root, through a stub `id`; a source changed between the hash and the copy, through a stub
   `install` first on the PATH that copies another file on Linux, and on Windows the child's text built with a wrong
   expected hash: neither script has a path for tests but `-ForceChild`, below). The `install.sh` tests put a stub
   `sudo` first on the PATH that fails the test if it is ever called, so no test reaches the real one. The child's text
   is built by a function the Windows tests call directly: each value embedded and quoted, a path holding a `'` and a
   space, a bad hash refused before any text is built. **The real places run on GitHub's throwaway machines:** an
   `install` job on `ubuntu-latest` (its runner has passwordless `sudo`) and on `windows-latest` (its runner is an
   administrator, elevated with UAC off) builds Bonsai, runs each installer at its real place, checks the owner, mode or
   ACL, the PATH (the entry once, `%...%` entries kept), `install.json`, `which -a` or `Get-Command`, and `--version`
   from a fresh shell, installs again while a `bonsai` from the first install runs, then removes it and checks it is
   gone. **A test-only switch forces the child on CI:** an elevated runner would otherwise run the steps in place and
   never start the child, so `-ForceChild` (documented in the header as for tests; refused unless the session is already
   elevated, so it can never raise a prompt; it only chooses the child over in place, and skips no step or check) makes
   the Windows job take the `RunAs` path, proving the child's code, arguments, working folder and exit code, and once
   more with a wrong expected hash (the child exits non-zero and leaves no file). Only the prompt itself waits for
   Rohan's 1.0 install (a declined one is tested through a stub of the start call); V1 reads the child line by line, and
   5.7 plans that first run with the way out (`-Remove`, or the folder deleted and the PATH entry taken out by hand).
5. **No agent runs either installer at a real place** (Rohan's 15:35 choice; spec §3: "No agent installs or replaces
   it"). Agents run as the same user, so a file in a download folder can be changed between Rohan's fingerprint check
   and his install by a shell command: the guard's `bonsai-stand-in` rule stops the file tools writing a `bonsai`
   outside a project, not a shell. That is spec §3's "tripwire, not a wall", and the risk below.
6. **`install.json`** (spec §3, §10) is in the installing person's home, never root's, so each user who runs Bonsai on
   the machine has their own after their own install. It is the tripwire `check` and `status` read (5.1.6's
   `bonsai-path`), in a folder the person can write: a forged one only silences that tripwire, since the hook lines
   never read it (5.3's (a), design 7 folded), and the guard refuses an agent's file-tool write to it (the home).
7. **The CI job is the piece's last commit,** made while `T-5606` is the only task `running` (`.github/**` is protected
   in Bonsai's repo, 5.4).

**V1, the installers and the machine's tripwires.** A fresh Opus agent, once 5.6.6 has landed with its `install` job
green and 5.6.1 and 5.6.2 have landed: reads both installers against spec §3, 5.3's (a) and this section; re-runs them
on scratch targets on both sides itself; reads the `install` job's logs on the pushed commit (the real places on
throwaway machines, the forced child's run among them); reads the elevated child and the PATH merge line by line: the
child's code inline (`-EncodedCommand`), no file it reads user-writable but the source it copies, its working folder
its own, no `$env:` of the parent's relied on, its copy hashed inside the target folder and removed on a mismatch, a
declined prompt a refusal; breaks them: a target folder others can write refused, a `bonsai` beside the script that
does not run refused, a PATH value holding `%SystemRoot%` entries kept as they were, a second install unchanged, a
second install while a `bonsai` from the first runs (both sides), a remove complete. And the machine's tripwires:
`settings set`, `labels attach` and `detach` refused with `CLAUDE_CODE_CHILD_SESSION` set, through a variable, `xargs`,
a script and a subprocess, with nothing written; a stamped build free only while its home is inside its scratch root;
`status_command` refused with a `;`, a `$`, a backtick, a newline, or past 100 characters, by `settings set` and by the
reader the guard shares (note 5.6.1, 1: a file written straight into a scratch home, and the guard's refusal of a call
that needs the mode, naming the file); an attach of `bonsai`, of a locked pack's namespace, or a redefinition refused.
Then the three known routes past those refusals, each on scratch homes only, recorded "gets through, as stated" where
it does: a script, `cp` or an interpreter writing the machine folder's `settings.json`; a script that unsets
`CLAUDE_CODE_CHILD_SESSION` and then runs `bonsai settings set`; a stamped build whose scratch root covers a scratch
folder standing in for the real home. It passes or fails; it fixes nothing; a must-fix is fixed forward.

#### Proof for each piece

Every piece: its Go tests and `go vet`, plain and with the fault tag, in WSL (by the landing rule's green climb of its
task, run by `/usr/local/bin/bonsai`) and natively on Windows (check 10's Windows half) before the push, the counts in
the run report; CI green on the pushed commit, the `install` job among them from 5.6.6; the Windows rules of `CLAUDE.md`
read in the diff (forward slashes in every stored and printed path, `install.json`'s and the settings' among them;
byte-stable files; LF scripts; no test needing a symbolic link or a file mode, the `.sh` mode checks on Linux only, with
the reason; Windows renames with busy retries); no Windows-only skip without a named reason; no test touching a real
home (`BONSAI_HOME` a temporary folder in every test, scratch homes in every script). Pieces that run Claude Code
(5.6.3's import try and cross-side rules, 5.6.4's session) record its version and the user settings hashes before and
after. Scripted runs live in `~/bonsai-checks/scripts/`, never committed. V1 and the end verifier re-run what they judge
themselves.

#### 5.6 done

A fresh Opus verifier, at the end of 5.6, after 5.6.5 and the floors, runs each check itself on the final commit, on
both sides where a check names them, and passes or fails 5.6:
1. **The set:** the manifest matches every byte; the three new schemas and `status`'s `bonsai_release` documented, each
   with an example its writer reproduces; the schema compare passes; `check --schema` prints each new format.
2. **The words:** `bonsai --help` lists the fourteen words of spec §4; `bonsai --help --json` gives `settings` and
   `labels` with every flag, exit code and error word; the operating skill's test passes and fails on a temporary copy
   that drops either word.
3. **`settings`:** on a scratch project and with `--machine`, `show` and `set` for every key; each bad value and unknown
   key exits 2 naming what is allowed; a set is all or nothing, byte-stable on two runs and on both sides, and keeps an
   unknown key; `status_writes=command` then makes `status --json` read `command`, and in a scratch session on WSL the
   guard refuses an agent's edit of a task's status line, naming the command.
4. **The tripwire:** `settings set`, `labels attach` and `detach` with `CLAUDE_CODE_CHILD_SESSION` set exit 4 with `who:
   person` and write nothing (the home hashed before and after), through a variable, `xargs`, a script and a subprocess;
   `settings show` answers; a stamped build is free only with its home inside its scratch root; a `status_command`
   outside its rule, written straight into a scratch home's file, refused by the reader and by the guard. The three
   known routes, on scratch homes only, each recorded "gets through, as stated" where it does: a script, `cp` or an
   interpreter writing the settings file; a script unsetting `CLAUDE_CODE_CHILD_SESSION` before `bonsai settings set`;
   a stamped build whose scratch root covers a scratch folder standing in for the real home.
5. **The stranded folder:** a fixture checkout moved: `check` warns `stranded` with the old path and its exact lines,
   never the exit code; the lines, run as a person, bring the settings and labels over and clear it; `status`'s text
   names it.
6. **The cache:** `update` with `cache_keep_days` removes an unused clone past its days, keeps the lock's and every
   `--adopt` copy, and says so in one line.
7. **`labels`:** attach, the same again (`unchanged`), a newer additive version (`updated`), a redefinition, `bonsai`
   and a locked pack's namespace each refused, detach, detach again (`nothing`); after an attach, `status --json` lists
   it `from: machine`, a scratch session's opening context shows it, and in `command` mode an agent's edit setting its
   `outside` label is refused.
8. **The personal layer:** `init --yes` writes the empty index once and never over one; a preview and `check` write
   nothing in the home; each of the five warnings from a fixture, none changing the exit code, the caps from their one
   table; the decoy secret's value in no output; on Windows the copy refreshed by `update --yes` (only `INDEX.md` and
   `notes/*.md`, within the limits), `personal-copy` from the record after the copy changes, `check` never reaching the
   source, and an unreachable source answered with a note within 2 s; the import's measurement in the run report (the
   `@~/` form, a missing file, any external-import prompt); the cross-side deny rules written by `update` and their
   tries in the run report, one form re-tried by the verifier on each side by 5.5.6's method; under (B), not chosen,
   also an agent's file-tool write of a personal note allowed and of any other file in the home refused.
9. **`--line`:** each part on fixtures, `N running` among them; it returns with its input held open; nothing printed
   and exit 0 outside a linked checkout; a broken `bonsai.yaml`
   gives its short line and exit 0; a title holding an escape code printed without it; `--json` valid against
   `bonsai.line/1`; the verifier's own p50 and p95 on both sides within the budgets on the fixture project; the real
   session's capture in the run report.
10. **The newer release:** against a stamped local source with tags `v0.4.3`, `v1.0.0`, `v1.0.1` and `base-v2.0.0`, a
    build stamped `1.0.0`: `status --full` records `v1.0.1`, `state: newer`, and both sides' lines exactly as the table
    gives them; plain `status`, `check` (`bonsai-newer`) and `--line` say it offline; a second `--full` within the day
    reads nothing (the source made unreachable changes no answer); an unreachable source with no record gives `unknown`
    with why, exit 0; a `dev` build gives `not-a-release`; a record whose `newest` holds a command gives no lines and
    `unknown`; every line's URL starts with `https://github.com/LastStep/Bonsai/releases/download/v`; the plain build,
    on a scratch `BONSAI_HOME`, reads the real repository once within 10 s (today: no release at or above 1.0.0).
11. **The installers:** on scratch targets on both sides, install, the record read by `check`, again unchanged, again
    while a `bonsai` from the first runs, remove, each refusal; the test holding both scripts to the installed places
    and the record's keys passes, and fails on a temporary copy naming another place; the `install` job green on both
    sides on the final commit, its log showing the forced `RunAs` child and the wrong-hash child leaving no file; V1's
    report read.
12. **Unattended:** every new refusal's `--json` carries `error` with a known word and `next` with `who`; every
    `next.do` that is a command runs as written.
13. **Check 10, the ladder and CI:** `go test ./...` and `go vet ./...`, plain and tagged, in WSL and natively on
    Windows, run by the verifier; `/usr/local/bin/bonsai ladder --task <its task>` green on the final commit; CI green
    on it.
14. **Stop lines and records:** 5.6's hours under 26, this section's planning and review included; step 5's Windows-only
    tally; option rounds (none planned); nothing written or run in the studio's checkout (its files read with `git show`
    only) or in Mimas; nothing written outside the repo and the scratch folders but `~/.bonsai/` by the orchestrator's
    own real commands and Claude Code's accepted writes; no `install.json`, `personal/`, home `settings.json`, machine
    `settings.json`, `labels/`, `cache/release.json` or `cache/personal-copy.json` in the real home written by a 5.6
    run; the user settings hashes around every Claude Code run; every landing matched to its green `ladder` record and
    each task's `bonsai.allows` to this section's "Owns"; `bonsai.yaml` changed only by the floors; the Haiku audit of
    `.bonsai/sessions.md` against the run reports. **Nothing private:** a grep of the diff, the commit messages and the
    task files.

#### Risk in the code, 5.6

- **Root and administrator steps.** The installers are the only code in step 5 that runs as root or administrator. Each
  keeps its elevated part to the copy, the PATH and the checks after; the Windows child's code is inline, never a file
  an agent could rewrite, and it hashes its copy inside the folder only administrators write. The real places run only
  on throwaway CI machines before Rohan's 1.0 install, the `RunAs` child forced there by `-ForceChild`; only the UAC
  prompt from an unelevated session runs first at his install, so 5.7 plans its way out.
- **An install over a running `bonsai`.** Hooks run Bonsai all day: Linux renames a whole new file over the old one;
  Windows renames the running one aside first. A hook starting in the instant between Windows' two renames refuses
  (fails closed); none runs a half-written file.
- **The machine's PATH on Windows.** A careless edit can expand or cut it. Read raw, written as `ExpandString`, the
  merge tested on strings, the real edit tried on CI's machine and undone by `-Remove`.
- **A download changed before the install.** Agents run as the same user and could change a file in the download folder
  between Rohan's fingerprint check and his install: spec §3's tripwire, not a wall. The lines run one after another in
  one fresh folder; each installer hashes the source first and its root- or administrator-owned copy before it is put
  in place, which catches a change after that hash; the installer prints the installed copy's fingerprint,
  `install.json` records it, and `check` compares the PATH's `bonsai` with it. What the fingerprint line cannot catch:
  a release changed at its source, both files at once (5.7's provenance, "Hand-offs to 5.7").
- **The release record feeds lines a person runs.** Only a version of three whole numbers ever reaches a line, checked
  at every read, inside fixed GitHub URLs; a forged record gives no lines.
- **A tripwire, not a wall,** for the machine's settings and labels: an agent's shell can still write the home (the
  sandbox is not on, contract §10.6). The commands refuse themselves, the guard refuses them and the walls refuse the
  file tools; a script writing the file, a script unsetting the variable, or a stamped build rooted over the real home
  still gets through (V1 records each). The studio's bridge flags a lost `command` setting, and should also compare
  `status_command` with what its registration set (a note for its plan).
- **`status_command` is printed by the guard and run by agents:** its narrow characters, checked when set and whenever
  read (the guard's reader included), stop a person's typo, or a forged file, from becoming a command line with `;` or
  `$(...)` in it.
- **What loads into every session:** the personal index. Person-written under (A); budgets and the secret scan as
  warnings; the import line is the person's, at 5.7.
- **`--line` runs on every refresh.** A slow line slows every session's status line; budgets measured on both sides, a
  cache only if needed; a project on a Windows drive read from WSL is measured, not promised.
- **Text into a terminal:** task titles in `--line`, stripped of control characters.
- **The network:** one read a day at most, 10 s, in `status --full` only; GitHub unreachable never fails a command. A
  tag pushed before its release's files are up makes the download line fail plainly ("try again in a few minutes").
- **WSL from Windows:** reading `\\wsl.localhost` may start WSL; 2 s, then a note, and only in `init --yes` and `update
  --yes`, never in `check` or rung 0. Each side's Bonsai home is walled from the other's sessions by the engine's
  cross-side deny rules, where a form holds.
- **Shared files:** what runs side by side shares none; 5.6.5 waits for 5.6.3 and 5.6.4. The reference page and
  `format.CheckWords` are written in lane A and 5.6.5 only, one after the other.
- **Processes:** `tmux` sessions (5.6.3, 5.6.4): each agent stops what it started and checks with `ps`; the orchestrator
  sweeps after each agent.

#### Stale or in tension in the spec, for 5.6

- **§4's table: `status` "Writes: nothing":** `status --full` writes one file, the home's `cache/release.json`, a
  refillable copy of its network read (note 5.6.5, 2), and `--line` may write `cache/line/<machine key>.json` if its
  budget ever needs a cache (note 5.6.4, 4); nothing in the project, and a write that fails (a read-only home) never
  fails `status`. Spec §4's row gets a dated note, written by the orchestrator on `main` (this plan does not edit the
  spec).
- **§6: "deleting `cache/` by hand is always safe; it refills", against §10: "Its `cache/` also keeps `--adopt`
  copies":** true of the pack clones only; `cache_keep_days` cleans `cache/git/` and never `cache/adopted/`, which holds
  a person's saved edits (note 5.6.1, 6).
- **§6: every file cleaned is "one `clean` record in the log":** the cache's clones are not logged (refillable, outside
  every project, and a `clean` record's target is project-relative, contract §8.2); `update` names them in its text.
- **Contract §3: "`bonsai check` and `status` report the stranded folder":** a `check` warning (spec §6) and a line in
  `status`'s text; `status --json` carries no warning (5.1.6), and the lost setting shows in its `status_writes`.
- **§10: the Windows copy "refreshed when a Windows-side project links", and "a copy older than WSL's":** refreshed by
  `init --yes` and `update --yes` on Windows once a person names WSL's folder in a new home setting, `personal_from`,
  beside §4's `cache_keep_days`; "older" read as "not the same as WSL's" (note 5.6.3, 4); and `check` reports it from
  the last refresh's record, never reaching WSL (up to 2 s on every `check` and rung 0 otherwise), so a change on WSL's
  side shows at the next `update`, not at the next `check`.
- **§10's personal layer, "budget 40 lines":** 40 lines and 4 KB the index (40 lines at the project index's rate), 4 KB
  a note, in one Go table with the project's caps (note 5.6.3, 3); its checks are warnings, not findings, as the layer
  is the machine's.
- **§10: "Agents write notes through the workflow pack's memory skill"** (5.5 moved it to base's): true of the project's
  notes. For the personal layer it is settled by Rohan's choice in this section, (A), the person's, 5.3's guard
  and deny rule refusing agents in the home (note 5.6.3, 2); (B), as the spec has it, in base's skill, with a
  guard change, was not chosen. The memory schema's description says which (5.6.0).
- **§13 item 4, the memory move "to `~/.bonsai/personal/`":** follows Rohan's choice, (A): done by a person, or in a
  session outside any linked project, the studio's plan's choice (step 7); under (B), not chosen, an agent could do it.
- **Contract §15.2: registration "runs Bonsai's attach and settings commands":** both refuse in an agent's session, so
  the studio's registration is run by a person or by the studio's own program outside one; a note for the studio's plan.
- **§3: "Each install writes `<home>/install.json`":** the installing person's home, never root's (the installer refuses
  `sudo`), so per user (note 5.6.6, 6).
- **§3 and §14 have no word on a newer release:** Rohan's 9 Oct note in §3 adds it; built in 5.6.5 inside the row's
  hours (this section's "Hours").
- **§4: every command but `hook` takes `--json`; `--line` is a line:** `--line --json` prints `bonsai.line/1`.
- **§14's 5.6 row, "the workspace half of today's statusline":** today's task part with Bonsai's asks and release
  marker; the git counts, Unity's editor and the studio's services stay the studio's line's.
- **§17 step 8 gives the pre-release install as typed lines:** from 1.0 the installers replace them (5.7); 5.6 asks no
  install.
- **Hand-offs to 5.7:**
  - Each archive holds `install.sh` and `install.ps1` at its top, beside the binary (`.goreleaser.yaml`'s
    `archives.files`); 5.6.5's test holds the lines' names to `.goreleaser.yaml`, so 5.7 keeps it passing.
  - Rohan's 1.0 installs on both sides go through the installers and those lines, the UAC prompt's first real run among
    them, with its way out; his list's "At 5.7" says so.
  - **His memory import line**, in the 1.0 install batch, once the first `init --yes` or `update --yes` by the installed
    1.0 has made the index on his machine (about a minute, in WSL):

    ```bash
    printf '\n@~/.bonsai/personal/INDEX.md\n' >> ~/.claude/CLAUDE.md
    tail -n 2 ~/.claude/CLAUDE.md
    ```

    The second line must end with `@~/.bonsai/personal/INDEX.md`; he sends the orchestrator what it printed (the first
    line adds an empty line before the import, so it never joins his file's last line). Then an agent checks that it
    loads, writing nothing of his: the scratch launcher (`claude-here`) sets the PATH, the plugin cache and
    `BONSAI_HOME` but not Claude Code's configuration folder (`design/plan.md`, "Test sessions and the launcher"), so a
    session it starts reads his real `~/.claude/CLAUDE.md`, and the import's `~` is his real home whatever
    `BONSAI_HOME` says. In a scratch project, an interactive `tmux` session through `claude-here`: `/memory` lists
    `~/.bonsai/personal/INDEX.md` among the loaded files (or, if that version's `/memory` does not list imports, the
    session is asked for the loaded index's `title:`, `Personal memory`, nothing private); any external-import prompt
    recorded with its words (if one appears, his own sessions will show it too, and the orchestrator tells him); the
    index's notes never copied into a report; Claude Code's version and the user settings hashes before and after
    recorded.
  - The install lines' stronger check: once build provenance exists (spec §12 step 8), the lines gain `gh attestation
    verify <archive> -R LastStep/Bonsai` (5.7 checks it runs on both sides, and with what login), or the release
    publishes the binaries' own SHA-256, so the installed file is checked as 5.4's line checked it.
  - Homebrew's place against 5.3's (a) stays 5.7's.
- **Hand-offs to the studio's plan (step 7) and Mimas's (step 8):** registration outside agent sessions; its bridge
  comparing `status_command`, not only `status_writes`, with what its registration set (an agent could point it at a
  project script while the mode still reads `command`); its status line calling `bonsai status --line --json`; the
  memory move as Rohan chose (note 5.6.3, 2); at Mimas's link, `personal_from` on Windows, the same import line in the
  Windows `~/.claude/CLAUDE.md` (Rohan's), and the engine's cross-side deny rules over each side's Bonsai home (note
  5.6.3, 6), re-tried on that machine.

### Step 5.7: the release (9.5-18.5 h with the trial, re-ask at 24)

**Rohan's (B), and what else of his changes.** This section comes to Rohan: it holds his steps (the GitHub settings,
switching the release workflow on, every tag, the installs of a pre-release and of 1.0, the trial's pick and its
review, his memory line), a public release that cannot be taken back, and his decisions, all answered on 10 Oct
(11:19-11:29; "Decisions for Rohan", below): Homebrew stays at the old product (1 (A)); his approval holds each release
(4 (A)); the packs' tags are locked (5 (A)); and, in place of a plain rehearsal, **a trial of Bonsai on itself on
`1.0.0-rc.1`** before 1.0, with his own trial on a fresh project after it (3). Two more were settled earlier that day:
Go 1.27 for the release ("we can go with go 1.27 latest"; the bump is its own piece, landed before 5.7 starts) and a
newer `gh` in WSL (2 (A), done). The trial takes 5.7 past the spec's hours: **9.5-18.5 hours, re-ask at 24**, against
the spec's 6-11 and 14 ("Hours, order and your steps"); the section comes back to him for approval with those figures.
The release files, `.github/workflows/release.yml` and `.goreleaser.yaml`, are person-only in Bonsai's repo (5.4's
switch, the stricter-only rule): his approval of this section is his word for the changes it names to them, and for
nothing more. The order of the parts stands: 5.7 is the last. No repo is new. What goes public: Bonsai 1.0 (its files,
the README, the release notes), one or more pre-releases (`1.0.0-rc.N`), the trial's feature as an ordinary commit on
`main`, and the first tags of the two packs. No format changes are planned: the install lines Bonsai prints gain one
line (a list's entry, so `bonsai_release.lines` keeps its schema), and `base`'s `version` becomes `1.0.0`; a trial
finding that needs one comes to him first ("Risk in the code, 5.7").

#### For Rohan, in plain words

**What 5.7 gives you.** Bonsai 1.0, released on your word once Bonsai has worked on itself cleanly, and installed by
you on both sides.
- **The release path made safe before it is switched on.** Every action and tool Bonsai's workflows name is fixed to
  one exact version, so no update of theirs slips in unseen (GitHub's own machines and its code-scanning bundle still
  update themselves). A release runs only from a tag on a commit already on `main` whose checks are green. The program
  is built and checked in a step that cannot publish anything or sign anything; the step that publishes runs nothing
  from the build, and it waits for your approval inside a GitHub setting (the `release` environment). GitHub signs a
  record of how each file was built (by Bonsai's own release workflow, from that tag, on GitHub's own machines), and a
  GitHub lock (immutable releases) stops anyone changing a published release's files or moving its tag. No secret is
  needed at all, so nothing in a release can write outside Bonsai's repo.
- **A rehearsal on every push.** GitHub builds every file a release would publish, checks each (the right files in each
  archive, no test-only code in the program, the installer working from its archive) and publishes nothing.
- **A stronger check in your install lines.** One more line on each side, before anything from the download runs,
  checks GitHub's signed record (your WSL `gh` is new enough for it). What it proves: the files were built by Bonsai's
  release workflow from that tag, on GitHub's own machines. What it does not: that the code is good. A commit on `main`
  is built and signed like any other, whoever made it; the guard, the ladder and the verifiers stand there, as before.
- **The trial you asked for: Bonsai on itself, on a pre-release.** You install `1.0.0-rc.1` with the real install lines,
  so Bonsai's own repo runs it. Agents propose two or three small real features for Bonsai; you pick one; it is built as
  an ordinary Bonsai task through the whole pipeline (its task file, Bonsai's guard, the ladder, the log of every
  session, the tables, any question to you, a verifier, the landing). Then you get an analysis in plain words with the
  evidence: which files Bonsai created and changed, what the log and the tables hold, how the agents worked (sessions,
  runs, refusals, questions, ladder results), and anything rough. Fixes go out as `rc.2` and so on, the trial or its
  affected part repeated, until you are happy. Then your own trial on a fresh project, on your own, and only then your
  word for 1.0, from the same code as the last pre-release.
- **Go kept current** (your word, 10 Oct). 1.0 is built with Go 1.27's newest patch: Go 1.25 got its last security fix
  on 19 Aug, when Go 1.27 came out. The bump lands before 5.7 starts; 5.7 checks it still holds at release time and
  that the vulnerability check is clean.
- **The words brought up to date.** The README (what Bonsai is, how to install and check it, and the steps to link a
  first project, which your own trial follows), the contributing guide, the changelog with 1.0's notes, the security
  policy, the issue templates, and `bonsai --help`, which still says "rebuild in progress".
- **The packs' first versions.** `base` and `workflow` at `1.0.0`, their checks taking Bonsai from the 1.0 release by
  its fingerprint, then tagged by you.

**What stays yours, and why.** Agents act on GitHub as your account, so every step that makes something public or
changes a setting is yours, typed by you: the settings, switching the release workflow on, every tag, and approving
each release. From 1.0, `CLAUDE.md` tells every agent the same in one rule: only you tag and approve a release; no
agent approves, re-runs or cancels a release run, or changes a setting. A release is public and stays: GitHub's lock
means its files can never be changed, only followed by a newer version. Installing the program stays yours (your 9 Oct
choice), so every pre-release Bonsai runs on itself is your install too; so do the trial's pick and the judgement of its
results, your own trial on a fresh project, and the memory line in your own `~/.claude/CLAUDE.md` (no agent edits that
file). Everything else is the agents': the code, every check on every file, reading each setting back after you change
it, checking each release before you install it, running the trial and writing its analysis.

**How 1.0 happens, in order.** Each batch reaches you in one message, with every number filled in.
1. Agents build and check everything; a fresh verifier passes the release path, and reads every line of batch 1
   against GitHub's documentation, before anything is switched on.
2. **Batch 1, the settings** (about 10 minutes, in WSL).
3. **Batch 2, `1.0.0-rc.1`** (about 3 minutes): your tag and your approval; agents and a fresh verifier check every
   published file on both sides.
4. **Batch 3, rc.1 installed** (about 15 minutes): on WSL, so Bonsai's own repo runs it, and on Windows, so the Windows
   installer's first real run (its one "allow this app" question) happens on a pre-release, where a problem costs a new
   `rc` and not a 1.0.1; then your pick of the trial's feature.
5. **The trial on Bonsai** (agents; nothing of yours unless the work asks you a question, which reaches you as Bonsai's
   own question to a person).
6. **Batch 4, the analysis** (about 30 minutes to read): you say "happy", or what to fix. Each fix round is a new
   pre-release: your tag, your approval and your WSL install again (about 10 minutes; Windows too only if the fix
   touches Windows), then the trial or its affected part again and a short analysis.
7. **Your own trial on a fresh project**, on your own and in your own time, once you are happy with the trial on
   Bonsai: the README's steps link a new project to Bonsai and its `base` pack; agents help only if you ask.
8. **Batch 5, 1.0** (about 25 minutes, in two or three sittings the same day): your word; the tag, on the same code as
   the last pre-release, and your approval; once agents have checked the published files, the installs (WSL, then your
   memory line, then Windows); and last, the two pack tags.

**Batch 1, the settings.** In WSL, one line at a time. None of these can be tried without changing a setting, so a
fresh verifier has read each against GitHub's documentation before it reaches you. Several print nothing when they work,
so each setting is followed by a line that reads it back and must print what is written beside it. If a line prints an
error, or a read-back prints something else, stop and send the orchestrator what it printed. Afterwards an agent reads
every setting back once more (changing none) and tells you what it found.
- **Your `gh`** (updated on 10 Oct to 2.102.0, GitHub's package): `gh --version` must still print 2.97 or newer (2.97
  fixed a flaw in how older ones match the release workflow's name), and `which -a gh` must print `/usr/bin/gh` first
  (a `/bin/gh` after it is the same file, as on 10 Oct).
- **The old Homebrew token,** never confirmed revoked: on github.com, your Settings, Developer settings, Personal
  access tokens (both lists). Delete any token made for the Homebrew tap, and tell the orchestrator what you found: a
  live one could still change the tap.
- **The `release` environment,** open only to tags starting `v`, and waiting for your approval. The first line prints
  the environment; the second prints its tag rule; the last two must print `v* tag`, then `required_reviewers` and
  `branch_policy` (in either order):

  ```bash
  printf '{"wait_timer":0,"prevent_self_review":false,"reviewers":[{"type":"User","id":%s}],"deployment_branch_policy":{"protected_branches":false,"custom_branch_policies":true}}' "$(gh api user --jq .id)" | gh api -X PUT repos/LastStep/Bonsai/environments/release --input -
  gh api -X POST repos/LastStep/Bonsai/environments/release/deployment-branch-policies -f name='v*' -f type=tag
  gh api repos/LastStep/Bonsai/environments/release/deployment-branch-policies --jq '.branch_policies[] | .name + " " + .type'
  gh api repos/LastStep/Bonsai/environments/release --jq '.protection_rules[].type'
  ```

- **GitHub's lock on published releases.** The first line prints nothing when it works; the second must print `true`:

  ```bash
  gh api -X PUT repos/LastStep/Bonsai/immutable-releases
  gh api repos/LastStep/Bonsai/immutable-releases --jq .enabled
  ```

- **The packs' tags locked.** In Bonsai's repo a rule that stops `base-v*` tags being moved or deleted; in the workflow
  repo the same rule for its `v*` tags, and GitHub's lock on its published releases. The first two lines print each
  rule; the third prints nothing; the last three must print `pack-tags tag`, `release-tags tag` and `true`:

  ```bash
  echo '{"name":"pack-tags","target":"tag","enforcement":"active","bypass_actors":[],"conditions":{"ref_name":{"include":["refs/tags/base-v*"],"exclude":[]}},"rules":[{"type":"update"},{"type":"deletion"}]}' | gh api -X POST repos/LastStep/Bonsai/rulesets --input -
  echo '{"name":"release-tags","target":"tag","enforcement":"active","bypass_actors":[],"conditions":{"ref_name":{"include":["refs/tags/v*"],"exclude":[]}},"rules":[{"type":"update"},{"type":"deletion"}]}' | gh api -X POST repos/LastStep/bonsai-workflow/rulesets --input -
  gh api -X PUT repos/LastStep/bonsai-workflow/immutable-releases
  gh api repos/LastStep/Bonsai/rulesets --jq '.[] | select(.target == "tag") | .name + " " + .target'
  gh api repos/LastStep/bonsai-workflow/rulesets --jq '.[] | select(.target == "tag") | .name + " " + .target'
  gh api repos/LastStep/bonsai-workflow/immutable-releases --jq .enabled
  ```

- **The release workflow switched back on, last.** The second line must print `active`:

  ```bash
  gh workflow enable release.yml -R LastStep/Bonsai
  gh api repos/LastStep/Bonsai/actions/workflows/release.yml --jq .state
  ```

**Batch 2, `1.0.0-rc.1`.** One line, with the commit the orchestrator names:

```bash
gh api -X POST repos/LastStep/Bonsai/git/refs -f ref=refs/tags/v1.0.0-rc.1 -f sha=<the commit the orchestrator names>
```

Then open the link the orchestrator sends (the release run on GitHub; your phone works), choose "Review deployments",
tick `release` and "Approve and deploy". About ten minutes later a pre-release, 1.0.0-rc.1, is on Bonsai's Releases
page. Agents check every file in it on both sides, a fresh verifier checks their work, and you get a short report.
**If a run fails before it publishes** (the orchestrator tells you; nothing is public then but the tag), the fix lands
first, then one line takes the tag away and the tag line is typed again on the fixed commit, so the number is kept:

```bash
gh api -X DELETE repos/LastStep/Bonsai/git/refs/tags/<the tag the orchestrator names>
```

**Batch 3, rc.1 installed, and your pick.**
- **WSL**, once the orchestrator says rc.1's files are checked. Seven lines, one at a time (for 1.0, batch 5, the same
  lines name `1.0.0`):

  ```bash
  cd "$(mktemp -d)"
  curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.0-rc.1/bonsai_1.0.0-rc.1_linux_amd64.tar.gz
  curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.0-rc.1/checksums.txt
  sha256sum -c --ignore-missing checksums.txt
  gh attestation verify bonsai_1.0.0-rc.1_linux_amd64.tar.gz -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.0-rc.1 --deny-self-hosted-runners
  tar -xzf bonsai_1.0.0-rc.1_linux_amd64.tar.gz
  sh install.sh
  ```

  The fourth line must print `OK`; the fifth, `Verification succeeded`; the last ends with Bonsai's fingerprint, a list
  of `bonsai` programs naming `/usr/local/bin/bonsai` alone, and `bonsai 1.0.0-rc.1 (commit ...)`. Anything else: stop
  and send the orchestrator what it printed. The way back in WSL is not the installer's remove (without a `bonsai`
  there, every edit in Bonsai's own repo is refused): it is 5.4's "back to the build before", whose two lines the
  orchestrator sends with the 5.4 pre-release's number. After it, `bonsai check` reports that the installed program is
  not the one Bonsai's install record names: expected, and cleared by the next install through the installer.
- **Only if the orchestrator says the pre-release changed Bonsai's own hook lines** (your answer (ii)):

  ```bash
  cd ~/Servers/Bonsai
  /usr/local/bin/bonsai update --allow-exec --yes
  ```

- **Windows**, in a normal PowerShell (not as administrator), one line at a time:

  ```powershell
  cd (New-Item -ItemType Directory (Join-Path $env:TEMP ("bonsai-" + [guid]::NewGuid())))
  curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.0-rc.1/bonsai_1.0.0-rc.1_windows_amd64.zip
  curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.0-rc.1/checksums.txt
  (Get-FileHash bonsai_1.0.0-rc.1_windows_amd64.zip -Algorithm SHA256).Hash -eq ((Select-String -Path checksums.txt -SimpleMatch bonsai_1.0.0-rc.1_windows_amd64.zip).Line -split ' ')[0]
  gh attestation verify bonsai_1.0.0-rc.1_windows_amd64.zip -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.0-rc.1 --deny-self-hosted-runners
  Expand-Archive bonsai_1.0.0-rc.1_windows_amd64.zip -DestinationPath . -Force
  powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
  ```

  The fourth line must print `True`; the fifth, `Verification succeeded` (PowerShell may show a stray character before
  it). The last line brings one Windows question, "Do you want to allow this app to make changes to your device?":
  choose Yes. It ends with the fingerprint, the `bonsai` programs Windows finds (`C:\Program Files\Bonsai\bonsai.exe`
  first) and `bonsai 1.0.0-rc.1 (commit ...)`. **The way out:** if you chose No, run the last line again and choose
  Yes. To take it out, in the same window (one more question):

  ```powershell
  powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1 -Remove
  ```

  If that fails too, in a PowerShell opened as administrator, then take `C:\Program Files\Bonsai` out of the system
  `Path` (System Properties, Environment Variables):

  ```powershell
  Remove-Item -Recurse -LiteralPath 'C:\Program Files\Bonsai'
  ```

  No Windows project uses Bonsai yet, so a failed install there blocks no work.
- **Your pick.** The orchestrator sends two or three small features for Bonsai, each in plain words: what it does for
  you, what it touches, its hours, and which parts of the pipeline it puts to work. Reply with one.

**The trial and its analysis (batch 4).** The orchestrator tells you when the feature has landed. The analysis is a page
in `records/` (and, if you like, a page you can read on your phone): what Bonsai created and changed, with the files
named; what its log and tables hold; how the agents worked, run by run; every refusal, question and ladder result; and
anything that was rough, with what the plan proposes for each. Reply "happy", or what to fix. For each fix round:

```bash
gh api -X POST repos/LastStep/Bonsai/git/refs -f ref=refs/tags/v1.0.0-rc.2 -f sha=<the commit the orchestrator names>
```

then your approval, then the WSL lines again with the new number (Windows too only if the orchestrator says the fix
touches Windows); the trial or its affected part runs again, and a shorter analysis follows.

**Your own trial on a fresh project.** Once you are happy with the trial on Bonsai, in your own time: a new folder, git,
and the README's "first project" steps (link it to Bonsai and its `base` pack, check it, open Claude Code there and ask
for something small), with the last pre-release installed. Tell the orchestrator what you found, or nothing at all:
your word for 1.0 is what counts. Agents help only if you ask.

**Batch 5, 1.0.**
- **Your word.** Reply "release 1.0", or say what to change first (a change means another pre-release first).
- **The tag and your approval**, as in batch 2, with `v1.0.0` and the commit named then: the last pre-release's code,
  with the notes' date the only change:

  ```bash
  gh api -X POST repos/LastStep/Bonsai/git/refs -f ref=refs/tags/v1.0.0 -f sha=<the commit the orchestrator names>
  ```

- **WSL and Windows**, once the orchestrator says 1.0's files are checked: batch 3's lines with `1.0.0` in place of
  `1.0.0-rc.1` (the orchestrator sends them written out). Each install replaces the pre-release in place.
- **Your memory line**, once the orchestrator says your memory's index exists:

  ```bash
  printf '\n@~/.bonsai/personal/INDEX.md\n' >> ~/.claude/CLAUDE.md
  tail -n 2 ~/.claude/CLAUDE.md
  ```

  The second line must end with `@~/.bonsai/personal/INDEX.md`; send the orchestrator what it printed. An agent then
  checks that your Claude sessions load it, changing nothing of yours.
- **Last, the two pack tags**, once the orchestrator says the packs' checks take Bonsai from 1.0 (the same day or
  later):

  ```bash
  gh api -X POST repos/LastStep/Bonsai/git/refs -f ref=refs/tags/base-v1.0.0 -f sha=<the commit the orchestrator names>
  gh api -X POST repos/LastStep/bonsai-workflow/git/refs -f ref=refs/tags/v1.0.0 -f sha=<the commit the orchestrator names>
  ```

  The first starts nothing; the second starts the workflow repo's own checks, which publish its release (that repo has
  no `release` environment, so no approval). Neither tag can then be moved or deleted. Projects may then name the packs
  by these tags. Bonsai's own repo keeps naming `base` by its commit; moving it to the tag waits for your word, whenever
  you like (5.5's rule), not planned here.

**Hours, order and your steps.** **9.5-18.5 hours, re-ask at 24, past the spec's 6-11 and 14 because of the trial you
asked for.** The release work itself stays at the spec's 6-11 (it gained the gate's findings, the stronger install
check, the packs' move to the release and the review's fixes, and lost the Homebrew work and the Go bump, which is its
own piece before 5.7). The trial adds 3.5-7.5 hours: setting Bonsai up on rc.1 and running the feature through the
pipeline (0.5-1), the feature itself (1-3, as you pick: each proposal names its own figure), the analysis (1-1.5), and
one round of fixes as `rc.2` with the trial's affected part again (1-2). Low 6 + 3.5 = 9.5; high 11 + 7.5 = 18.5; re-ask
18.5 x 1.3 = 24.05, so 24. Each further round (`rc.3` and on) is 1-2 hours more: at the high figures two more fit under
24 (18.5 + 2 + 2 = 22.5) and a third would cross it, when you get the numbers and the three choices. Step 5 becomes
142.5-225.5 hours. This section's planning and review count in the figures. If a stop line were crossed, the smaller
cut's "half of 5.7" would be the packs' move (5.7.6) and further pre-release rounds: 1.0 would still go out with every
check, and projects would keep naming the packs by commit. Your time: about 10 minutes for batch 1, 3 for batch 2, 15
for batch 3 with your pick, 30 to read the analysis, 25 for batch 5: about an hour and a half; plus about 10 minutes for
each further pre-release, and your own trial on a fresh project, untimed.

**Settled without you (technical):** the release notes come from the changelog, not from commit titles; builds for
Linux and Windows (and macOS, built but untested, with no installer: `install.sh` refuses there); GitHub's signed record
covers every archive, the fingerprint list and every program, so an installed copy can be checked at any time; a
release must come from a commit on `main` whose checks are green; the build and the publishing are separate steps, the
second running nothing from the first; Dependabot stays, its pull requests landed by the orchestrator's own commits (it
closes its own); the native Windows test run stays beside the ladder (no Windows climb for 1.0); `bonsai init` still
names no pack by default, its template showing `base-v1.0.0` as the example to add; rc.1 is installed on Windows too,
for the installer's first real run, though the trial on Bonsai runs in WSL.

**Decisions for Rohan.** All answered on 10 Oct (11:19-11:29); each keeps its context and the options not taken, as
the record.

1. **Homebrew for the new Bonsai. Your answer: (A).** The spec moved `brew install bonsai` to the new product and kept
   the old one as `bonsai@0.4`. But your 5.3 answer (a) has every project's guard run Bonsai only from
   `/usr/local/bin/bonsai` (WSL) or `C:\Program Files\Bonsai\bonsai.exe` (Windows), and Homebrew puts programs in its
   own folder (`/home/linuxbrew/.linuxbrew/bin` on Linux): a Homebrew-installed Bonsai would refuse every edit in every
   linked project until the installer were run as well.
   - **(A) Leave Homebrew at the old product (chosen).** The tap stays exactly as it is: `brew install
     LastStep/tap/bonsai` keeps giving 0.4.3, so 0.4.3 stays downloadable, as you asked (spec §12, Q8). The new Bonsai
     installs only with its installers. No tap token, no secret in the `release` environment, and no release writes
     outside Bonsai's repo. Lost: no `brew install` for the new Bonsai, and so no install route at all on a Mac (the
     installer refuses there; a Mac user copies the program by hand); the README says both. With it you named two
     later pieces of work, a Mac and each platform's own package manager, now in the plan's list of work outside step 5.
   - **(B) As the spec had it (not chosen):** the tap's `bonsai` moved to 1.0 and `bonsai@0.4` keeping 0.4.3, at the
     cost of a token you would make and store (a long-lived key that can write to the tap), a release writing to a
     second repo, and 0.5-1 AI hour more; a Homebrew install still guarding nothing by itself.
2. **The stronger check on your WSL side. Your answer: (A), done 10 Oct: you updated `gh` in WSL.** 5.6's fingerprint
   line proves the download arrived whole, not that the release is genuine (you were told on 10 Oct: "5.7 adds the
   stronger check before your first real install"). 5.7 makes GitHub sign a record of how each file was built, and the
   line `gh attestation verify ...` checks it before anything from the download runs (the installer runs with your
   password or administrator rights). Measured on 10 Oct: Windows' `gh` (2.102, logged in) does the check; WSL's was
   2.4, from 2022, which cannot. Your read-back that day: WSL's is now 2.102.0, GitHub's package, still logged in, at
   `/usr/bin/gh`. Weighed:
   - **(A) Update `gh` in WSL from GitHub's package source (chosen; done 10 Oct):** both sides run the same install
     lines, now and at every later release; one more package source on your WSL (GitHub's), kept current by `apt
     upgrade`.
   - **(B) In WSL, run the check with Windows' `gh` by its full path:** nothing installed; that one line longer and
     different from what Bonsai prints, at every release (note 5.7.3, 6 writes it out).
   - **(C) Skip the check line in WSL:** an agent checks the same file's record before you install; on WSL you would
     trust the agent's report, not GitHub's signature.
3. **Before 1.0. Your answer: a trial of Bonsai on itself on `1.0.0-rc.1`, then your own trial on a fresh project.**
   Your words: "first when we have a proper build going, i want to setup bonsai on itself. then do a test through
   adding a small feature, which goes through everything in the pipeline. then we analyze the results, like see what
   files were created, how the agents worked etc. then once im happy with everything, then we do the 1.0 release. so
   the main point is that bonsai should work on itself cleanly"; and "okay lets do rc.1, and i also want to do a trial
   on a fresh project, which i will do independently once im satisfied with the trial on bonsai". What was put to you:
   a release's first real run needs a real tag, and a defect found after publishing costs a version number.
   - **(A) A public pre-release first, `1.0.0-rc.1`, checked by agents (recommended then):** kept, and grown into the
     trial: rc.1 is installed and used for real, on Bonsai itself, and fixes go out as further pre-releases.
   - **(B) Straight to 1.0 (not chosen):** one batch fewer; a defect that only a published release shows would have
     cost a version number.
4. **What stops an accidental release. Your answer: (A).** Agents act on GitHub as your account, and the rules alone
   stopped them tagging.
   - **(A) Each release waits for your approval (chosen):** after a tag, the release does nothing until you click
     "Approve and deploy" on GitHub (the orchestrator sends the link). One click per release. An agent could approve
     through GitHub's API too, which the rules forbid: it stops accidents, not intent.
   - **(B) The spec's optional tag ruleset (not chosen):** no `v` tag could be made at all until you switched the
     ruleset off, for each release, and on again after.
   - **(C) Neither (not chosen):** a tag would start the release at once.
5. **Locking the packs' tags. Your answer: (A).** Bonsai's own releases get GitHub's lock, which also stops their tags
   being moved or deleted once published. The packs' tags got no such lock: `base-v1.0.0` has no release at all, and the
   workflow repo publishes its release on any `v` tag with no approval and no lock. A project already linked notices a
   moved tag (Bonsai refuses it, 5.1), but a new link would take the moved one.
   - **(A) Lock both (chosen):** in Bonsai's repo a tag rule that stops `base-v*` tags being moved or deleted; in the
     workflow repo the same rule for its `v*` tags, and GitHub's lock on its published releases. Creating a tag stays
     free, so nothing is switched per release. Six lines in batch 1 (about 3 minutes); a pack tag made by mistake stays
     for good (a later version supersedes it; deleting it means switching the rule off for a minute).
   - **(B) The workflow repo's releases locked only (not chosen):** `base-v*` tags and the workflow repo's tags before
     its release is published would stay movable.
   - **(C) Neither (not chosen).**

#### What exists, and what 5.1 to 5.6 will have added

**On `main` at `692cdae`** (5.1.0 to 5.1.8 landed; read each file; 5.1.8 changed none of them):
- `.goreleaser.yaml` (v2): `./cmd/bonsai` for linux, darwin and windows on amd64 and arm64, `CGO_ENABLED=0`, ldflags
  `-s -w -X main.version={{.Version}}`, `mod_timestamp`, no `-trimpath`; archives `bonsai_<version>_<os>_<arch>`
  (`tar.gz`; `zip` on Windows) holding `LICENSE*` and `README*`; `checksums.txt`; a changelog from commit titles with
  the old product's filters; `brews:` writing `Formula/bonsai.rb` into `LastStep/homebrew-tap` with
  `HOMEBREW_TAP_TOKEN`.
- `.github/workflows/release.yml`: build-only, `workflow_dispatch` with no inputs, `contents: read`, the tests, then
  `goreleaser build --snapshot --clean` through `goreleaser/goreleaser-action@v7` at `version: "~> v2"`; disabled on
  GitHub. `ci.yml`: `test` (Linux: tests and vet, plain and tagged, a Windows cross-compile), `windows`, `lint`
  (golangci-lint v2.11.4, `verify: false`, untagged), `govulncheck` (v1.7.0, pinned because later releases need Go
  1.26). `codeql.yml`. Every action by a moving major tag (`@v6`, `@v4`). `dependabot.yml`: weekly `gomod` and
  `github-actions` updates.
- `go.mod`: `go 1.25.0`, `toolchain go1.25.9`, no requirement. Read 10 Oct from go.dev: Go 1.25.14 (19 Aug 2026) is Go
  1.25's last patch, released with Go 1.27.0; Go fixes its two newest lines only; the newest are 1.26.9 and 1.27.2 (8
  Oct).
- `cmd/bonsai/main.go`: `version` from `-X main.version`, else `dev`; its doc comment and `usage()` speak of the
  rebuild in progress. `cmd/bonsai/hook_test.go`'s `TestNormalBuildHasNoFaultCode` builds both ways and holds six
  markers out of the normal build's bytes (strings survive `-s -w`; symbols do not, so `go tool nm` cannot check a
  release build).
- Words: `README.md` (the rebuild, a stub `bonsai`, `design/plan.md`, a root `STATE.md`, Homebrew giving 0.4.3),
  `CONTRIBUTING.md` (Go 1.25, the interim proof, `golang.org/x/sys` "once the Windows code needs it"), `CHANGELOG.md`
  ("Rebuild - Unreleased": the stub and the build-only `release.yml`), `SECURITY.md` (0.4.3's scope, "no release
  yet"), `.gitattributes` (`*.sh.tmpl`, comments on generated hooks and preview goldens), `Makefile` (`-s -w`, no
  `-trimpath`).
- **GitHub, read 10 Oct (read only):** environments: `github-pages` only; repository secrets: none; immutable releases:
  off (`gh api repos/LastStep/Bonsai/immutable-releases`: `enabled: false`); rulesets: `main-protection` (branch);
  tags `v0.1.0` to `v0.4.3`, the 0.4.x releases mutable, seven files each; open pull requests: none.
  `LastStep/homebrew-tap`: public, `Formula/bonsai.rb` at 0.4.3 (GoReleaser's, both systems and both chips), last
  pushed 13 May.
- **`gh`, measured 10 Oct:** WSL's was 2.4.0 (Ubuntu's package, March 2022: no `attestation`, no `release verify`, no
  `--branch`); Rohan updated it that day from GitHub's package source to 2.102.0 (read back: still logged in,
  `/usr/bin/gh`, `--deny-self-hosted-runners` among `attestation verify`'s flags). Windows' is 2.102.0, logged in.
  Windows' `gh attestation verify` on a release file GitHub's own CLI signs (`gh_2.102.0_linux_amd64.tar.gz`, `-R
  cli/cli`) exited 0 (with `--signer-workflow cli/cli/.github/workflows/deployment.yml` too), exited 4 with no login
  ("To get started with GitHub CLI, please run: gh auth login") and 1 with the wrong repository; it printed nothing when
  its output was not a terminal.

**What 5.1 to 5.6 will have added** (from this plan's notes; 5.7's start re-reads each against what landed, and
5.7.0's run report records any difference that changes a note below):
- 5.2.2: `bonsai --version` prints `bonsai <version> (commit <12 hex>)` from the build's `vcs.revision`.
- 5.3: the hook lines name `C:\Program Files\Bonsai\bonsai.exe`, then `/usr/local/bin/bonsai`, from two Go constants;
  perhaps `golang.org/x/sys` in `go.mod` (5.3 decides, 5.4's job objects follow).
- 5.4: Bonsai's `bonsai.yaml`: protected `.github/**`, `go.mod`, `go.sum`, `.golangci.yml`, `.goreleaser.yaml`,
  `CLAUDE.md`, `.gitignore` among others; person-only `.github/workflows/release.yml` and `.goreleaser.yaml`, changed
  only as an approved section names; tasks `T-5xyy` in `records/tasks/`; the landing rule; a task with grants runs with
  no other task `running`; `/usr/local/bin/bonsai` is the 5.4 pre-release, its SHA-256 in 5.4.7's run report.
- 5.5: `packs/template/` with `pack.yml` (actions pinned by commit; `bonsai` built from a pinned Bonsai commit; a
  `release` job on `v*` tags) and `ci/check.sh`; Bonsai's `packs` job; `packs/base/` at `version: "0.1.0"`, its `ci`
  skill pinning a Bonsai commit, its `pack.yaml` protected; `LastStep/bonsai-workflow`, public, its CI pinned to a
  Bonsai commit; Bonsai's repo linked to `base` by commit; `STATE.md` at `.bonsai/STATE.md`.
- 5.6: `install/install.sh` and `install/install.ps1` (with `-ForceChild`, for tests) and their `.gitattributes` line;
  CI's `install` job (both at their real places on GitHub's machines); `internal/engine/release.go` (the release tags
  read by `status --full`, `bonsai_release`, six lines a side from one Go table, a test holding their file names to
  `.goreleaser.yaml` as text); `install.json` written by the installers; the import line's form measured.
- Before 5.7, on Rohan's word (10 Oct, "we can go with go 1.27 latest"): `go.mod`'s `toolchain` at Go 1.27's newest
  patch (`go1.27.2` on 10 Oct), the `go 1.25.0` line kept, govulncheck re-pinned to a release that reads Go 1.27 and
  golangci-lint re-pinned if it must be; its own piece on branch `go1.27`, its hours in its own run report. And WSL's
  `gh` updated by Rohan from GitHub's package source (decision 2, done 10 Oct: 2.102.0).

#### The pieces and their order

The spec's row (§14): "The supply-chain fixes, `release.yml` back to tag runs inside the `release` environment and
switched on again (Rohan's step, §17), `bonsai@0.4`, the README" (6-11). The gate report's 5.7 findings, 5.5's and
5.6's hand-offs are added to it, and the trial of Bonsai on itself that Rohan asked for on 10 Oct (pieces 5.7.7 to
5.7.10) comes on top ("Hours, order and your steps").

**Hours.** The split is the planner's judgment, for sizing briefs, as in every part. **The release pieces:** 5.7.0
0.5-0.75, 5.7.1 1.5-2.5, 5.7.2 1-2.25, 5.7.3 0.5-1, 5.7.4 1-1.5, 5.7.5 1-2, 5.7.6 0.5-1: low 0.5+1.5+1+0.5+1+1+0.5 = 6;
high 0.75+2.5+2.25+1+1.5+2+1 = 11, the spec's 6-11. **The trial** (Rohan, 10 Oct): 5.7.7 0.5-1, 5.7.8 1-3 (the
feature he picks; each proposal names its own figure inside that range, and a pick above it is said to him with the new
total before it starts), 5.7.9 1-1.5, 5.7.10 1-2 (one fix round, to `rc.2`): low 0.5+1+1+1 = 3.5; high 1+3+1.5+2 =
7.5. **In all:** low 6 + 3.5 = 9.5; high 11 + 7.5 = 18.5; re-ask 24 (18.5 x 1.3 = 24.05). Each further fix round
(`rc.3` and on, `T-5711` and on) is 1-2 hours more: two fit under 24 at the high figures (18.5 + 2 + 2 = 22.5), a third
crosses it. Since the first draft: the Go bump left 5.7.0 (done before 5.7 on Rohan's word, its hours its own
piece's), a quarter of an hour off its high; the review's fixes (two jobs, the draft checked before publishing, the
tag-check query) add a quarter to 5.7.2's. The verifiers (V1, V2, the end verifier), the floors and this section's
planning and review count inside, carried in 5.7.0's run report ("What changes", item 3). Homebrew's (B), not chosen,
would have put 0.5-1 hour more in 5.7.1. Tasks (5.4's rule): piece 5.7.y is `T-570y` (5.7.10 is `T-5710`, each
further fix round the next); `T-5790` onward, in order, for V1, V2, each fix to the release path, the date of the notes,
the floors and the end verification.

| # | What is built | What proves it | Hours | Reads |
|---|---|---|---|---|
| 5.7.0 | **CI's own checks, and Go held**: the Go 1.27 toolchain (landed before 5.7) checked to hold, at Go 1.27's newest patch at release time, govulncheck clean on it; `lint` plain and with the fault tag; `actionlint` over every workflow; every action in `ci.yml` and `codeql.yml` pinned by its full commit with its version in a comment; `persist-credentials: false` on every checkout; a test holding every workflow file to those pins | `govulncheck -show verbose ./...` with no finding in the standard library on the toolchain in use; the pin test failing on a temporary copy that names `@v6`; both lint steps and `actionlint` green; check 10; the ladder; CI, each job's `go version` read | 0.5-0.75 | Gate §5 (5.7); this plan's "What changes", item 11; the Go piece's run report; spec §12 step 8 |
| 5.7.1 | **The release build and its rehearsal**: `.goreleaser.yaml` (GoReleaser pinned; `-trimpath`; the installers at each archive's top; no GitHub release and no changelog of its own; `base-v*` tags ignored; Homebrew by decision 1); a `release-check` job building every file in snapshot mode on each push and checking it; the dist checks in Go (no fault code, build settings, each archive's files, `checksums.txt`, the Linux installer run from its archive); the notes script; `install.sh` refusing all but Linux | The dist checks green in the job and locally on both sides against a local snapshot, and failing on each doctored copy; a snapshot passing with a scratch `base-v*` tag on its commit; `goreleaser check`; 5.6.5's name test green; 5.6.6's installer tests with the new refusal; check 10; the ladder; CI | 1.5-2.5 | Spec §3, §12 steps 7-8; GoReleaser's documentation at the pinned version (read, its version recorded); 5.6.5 note 5; 5.6.6 notes 1-4 |
| 5.7.2 | **The release workflow**: `release.yml` on the two tag patterns only, in two jobs: `build` (read-only, no environment: the tag and its commit's checks, the tests, GoReleaser, the dist checks, the files handed over) and `publish` (inside the `release` environment, running nothing from the build: GitHub's signed record, a draft, the draft checked, publish, the published release checked); the issue and pull request templates; `CLAUDE.md`'s release lines | `actionlint` and the pin test; the tag-check script run locally against Bonsai's own history (a commit on `main` passes, a scratch branch's fails, a malformed name fails) and its checks query against a commit whose CI is still running; the draft-reading way measured first; V1's line-by-line read; the first real run (batch 2, `rc.1`) | 1-2.25 | Spec §12 step 8, §17 step 4 and its 8 Oct note; GitHub's documentation on environments, check runs, immutable releases and artifact attestations (read, dated in the run report) |
| 5.7.3 | **The stronger check in the install lines**: 5.6.5's table gains `gh attestation verify` of the archive before it is unpacked, on each side; `docs/install.md` (the lines for 1.0.0, what each proves, the way out, checking an installed copy) held to the table; the operating skill's install line | 5.6.5's exact-text test with the new line, built only from the checked version; the doc test failing on a changed line; the line run on both sides against a known signed file (its words and exit codes recorded); the operating skill's test; check 10; the ladder; CI | 0.5-1 | 5.6.5 notes 1-5; 5.6's "Hand-offs to 5.7"; `gh attestation verify --help` on the versions in use |
| 5.7.4 | **The words for 1.0**: README; CONTRIBUTING; CHANGELOG (the rebuild's section replaced by `## [1.0.0]`, its notes); SECURITY; `.gitattributes`; `bonsai --help`'s and `main.go`'s words; the Makefile's build flags | A stale-phrase list grepped to nothing (Haiku lists, the orchestrator reads each hit); the privacy grep; `--help`'s tests; the notes script extracting 1.0's section; check 10; the ladder; CI; the orchestrator's read | 1-1.5 | Gate §5 (wording); spec §1-§3; `design/one-pager.md`; 5.4's CONTRIBUTING line; this section |
| 5.7.5 | **The releases, checked**: each pre-release (`rc.1`, and each fix round's) and 1.0, every published file checked on both sides; V2 on rc.1; Rohan's installs read back (rc.1 on both sides, each later rc in WSL, then 1.0); Bonsai's repo on each; the date of the notes at his word; the memory line's check | Each check's output in the run report; V2's report; Rohan's lines and words; `check` in Bonsai's repo; the memory check as 5.6's hand-off writes it | 1-2 | This section's batches; 5.4's way back; 5.6's "Hand-offs to 5.7" |
| 5.7.6 | **The packs at 1.0**: the template's `pack.yml`, base's `ci` skill and the workflow repo's `pack.yml` take `bonsai` from the 1.0 archive checked by its SHA-256; `base` and `workflow` at `version: "1.0.0"`; `init`'s template comment; the two tag lines; the workflow repo's release checked | The template's checks green on both sides with the archive; Bonsai's `packs` job and the workflow repo's CI green; after the tags, the workflow repo's release made by its own job, and nothing started by `base-v1.0.0` | 0.5-1 | Spec §5 (the pack's CI, step 2), §12 step 8; 5.5.0 note 5, 5.5.1 note 4, 5.5.2, 5.5's "Stale or in tension" |
| 5.7.7 | **Bonsai on itself on rc.1** (Rohan, 10 Oct): Bonsai's repo brought to rc.1 (its lines, `check`, a climb by rc.1); two or three small real features proposed; the one he picks run as an ordinary Bonsai task end to end; the trial's records | `check` clean in Bonsai's repo on rc.1; the proposals sent; the chosen task through every stage, each with its evidence (note 5.7.7, 4) | 0.5-1 | Spec §14 step 6; 5.4's switch and "Tasks and names"; 5.5.7; `CLAUDE.md`; this plan's pipeline as built by 5.7 |
| 5.7.8 | **The trial's feature**: Rohan's pick, built by a named builder session under rc.1's guard, climbed by rc.1's ladder (new tests failing first), recorded, verified and landed | Its own proof, by its kind; the landing rule; check 10; CI; its verifier | 1-3 | Its proposal; the parts of the spec and plan it touches |
| 5.7.9 | **The trial's analysis**: a records page for Rohan, in plain words with evidence: what Bonsai created and changed, the log and the tables, how the agents worked, every refusal, ask and ladder result, what was rough and what is proposed | The page's every claim pointing at a file, a record or a command's output; a Haiku listing of files and records checked against it; Rohan's reply recorded | 1-1.5 | The trial's run reports, `.bonsai/local/` (read, never committed), the tables, `git log` |
| 5.7.10 | **One fix round, to `rc.2`** (each further round the same, numbered on): the analysis's fixes as tasks, the release path's untouched or read again by V1's brief, a new pre-release, its files checked, Rohan's WSL install, the trial or its affected part again, a short analysis | As each fix's kind asks; 5.7.5's checks on the new pre-release; the repeated part's evidence; Rohan's reply | 1-2 | The analysis; this section |
| **5.7** | | | **9.5-18.5** (re-ask 24; the spec's 6-11 and 14, with the trial) | |

**The order, side by side where truly independent.** Rohan, 9 Oct: "if you can orchestrate work in parallel do that
whenever possible"; his 8 Oct bar stands: no shared file, and neither's proof resting on the other's. 5.7 starts once
5.6's end verifier has passed 5.6. 5.4's rule shapes most of it: a task whose `bonsai.allows` is not empty runs with no
other task `running`, and 5.7.0, 5.7.1, 5.7.2 and 5.7.6 each change protected files (`go.mod`, `.github/**`,
`.goreleaser.yaml`, `CLAUDE.md`, base's `pack.yaml`). The files each piece owns:

| Piece | Owns |
|---|---|
| 5.7.0 | `.github/workflows/ci.yml` and `codeql.yml`; `go.mod`'s toolchain line only if a newer Go 1.27 patch is out at release time; a new package `internal/selfcheck/` (its doc comment: checks of Bonsai's own repository files, run by every climb) with `workflows_test.go` |
| 5.7.1 | `.goreleaser.yaml`; `ci.yml` (the `release-check` job); `.github/scripts/release-notes.sh` (new); `internal/selfcheck/dist_test.go` and the fault markers' one home, `internal/selfcheck/markers.go` (moved from `cmd/bonsai/hook_test.go`, which then reads it); `install/install.sh` and its test (the refusal off Linux) |
| 5.7.2 | `.github/workflows/release.yml`; `.github/scripts/release-tag-check.sh` (new); `.github/ISSUE_TEMPLATE/` and `.github/pull_request_template.md`; `CLAUDE.md` (its GitHub lines) |
| 5.7.3 | `internal/engine/release.go`'s lines table and its test; `docs/install.md` (new) and its test; the operating skill's install line (`packs/base/skills/operating-bonsai/SKILL.md`) |
| 5.7.4 | `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `SECURITY.md`, `.gitattributes`, `Makefile`; `cmd/bonsai/main.go`'s doc comment and `usage()`, and their tests |
| 5.7.5 | `records/` only (its run report; the orchestrator's date commit of `CHANGELOG.md`'s 1.0 heading is under its own task, `T-579x`) |
| 5.7.6 | `packs/template/.github/workflows/pack.yml` and `packs/template/ci/check.sh`; `packs/base/bonsai/pack.yaml` (protected) and `packs/base/skills/ci/SKILL.md`; the engine's `bonsai.yaml` template comment; in the workflow repo, `.github/workflows/pack.yml` and `bonsai/pack.yaml` |
| 5.7.7 | `records/` (the trial's run report, the proposals); Bonsai's own link files only as an update to rc.1's lines needs them (`bonsai.yaml`, the lock, Claude Code's settings: person-only, under 5.5.7 note 4's rule, anything that runs code Rohan's line) |
| 5.7.8 | Whatever the chosen feature's proposal names, written into its task's grants before it starts |
| 5.7.9 | `records/trial-rc1.md` (new) |
| 5.7.10 | Whatever each fix names, as its own task |

1. **5.7.0 first, alone** (the Go 1.27 piece already landed): every later workflow change meets its pin test.
2. **Then 5.7.3 and 5.7.4 side by side.** Neither changes a protected file, they share no file, and neither's proof
   reads the other's: the README links `docs/install.md` by name (5.7.3's), and the changelog's notes describe this
   section's release path, not 5.7.1's code.
3. **Then 5.7.1, alone**: its notes script reads the 1.0 section 5.7.4 wrote.
4. **Then 5.7.2, alone**: its job runs 5.7.1's dist checks and notes script.
5. **V1** once 5.7.0 to 5.7.4 have landed with CI green, the `release-check` job among it. Then **batch 1**, its
   settings read back by an agent.
6. **`rc.1`** (batch 2), then 5.7.5's checks and **V2**. A fix to the release path is a task of its own (`T-579x`),
   read by V1's brief again (a fresh Opus agent), and a new pre-release follows, since the tag's commit holds the
   workflow that runs.
7. **Batch 3**: Rohan installs rc.1 on both sides; 5.7.5 reads it back; **5.7.7** brings Bonsai's repo to rc.1 and
   sends the proposals; he picks.
8. **5.7.8**, the feature, alone (it is the trial: nothing else runs beside it, so its sessions, log and tables hold
   only its work); then **5.7.9**, the analysis, and **batch 4**, his review.
9. **5.7.10**, a fix round per "what to fix" (`rc.2`, then on), each with its pre-release, his WSL install and the
   trial or its affected part again, until he says "happy". Then **his own trial on a fresh project**, on his own.
10. **Batch 5**: his word, the date commit, the `v1.0.0` tag on the last pre-release's code, 5.7.5's checks of 1.0's
    files, then his installs and his memory line.
11. **5.7.6** once 1.0's files are checked; its task is then the only one `running` (5.7.5's install checks, which
    change no repository file, go on under `T-5705` at `verify`). Then batch 5's last two lines.
12. **The floors** (5.4's "once a part", under a `T-579x` task, climbed by the installed 1.0), the Haiku audit of
   `.bonsai/sessions.md` against the run reports, then **the 5.7 end verifier**. Then `.bonsai/STATE.md` rewritten and
   Rohan's roadmap updated: Bonsai 1.0.

**Who builds and verifies.** Opus builders for 5.7.1, 5.7.2 (the release path, security) and 5.7.3 (lines Rohan
types); Sonnet builders for 5.7.0 (mechanical; V1 reads it), 5.7.4 (prose; the orchestrator's read and the greps
behind it) and 5.7.6 (pins written in full here); Sonnet agents for 5.7.5's scripted checks on both sides and the
memory check; for the trial (5.7.7 to 5.7.10) the orchestrator itself, the feature's builder by its kind (Opus for a
guard, hook or release change, Sonnet otherwise) as a named session, a fresh verifier for it by `CLAUDE.md`'s rule, an
Opus agent writing the analysis and a Haiku agent listing the files and records it rests on; a Haiku agent for the
stale-phrase and privacy listings and the sessions audit. Every piece lands on the
landing rule (a green climb of its task by `/usr/local/bin/bonsai` at the exact commit), check 10's Windows half, CI and
the orchestrator's read of the diff, which the run report says. **V1**, a fresh Opus verifier, on the release path
before anything is switched on (CI and release, Rohan's rule); **V2**, a fresh Opus verifier, on rc.1's published
files before Rohan installs it (the release path's first real run); **the 5.7 end verifier**, fresh Opus, on "5.7
done".

#### Where each inherited finding is settled

The gate report's section 5, its 5.7 list; then the outline's "Settles", and what 5.3 to 5.6 and the plan hand on:

| Finding | Settled in | How |
|---|---|---|
| Go 1.25.9's standard-library findings; "a later 1.25.x patch fixes them" (gate §5, 5.7; "What changes", item 11) | Before 5.7; 5.7.0 | Go 1.25 has had no fix since 19 Aug (1.25.14 its last); on Rohan's word (10 Oct) the toolchain moved to Go 1.27's newest patch as its own piece before 5.7, the `go` line kept; 5.7.0 checks it holds, moves to a newer 1.27 patch if one is out, and names any finding left with no fix (note 5.7.0, 1-2) |
| CI's `lint` without the fault tag (gate §5) | 5.7.0 | A second lint step with `--build-tags=bonsai_test_fault`; both must pass |
| The fault switch absent from the release build, checked on that build (gate §5; 2.11) | 5.7.1, 5.7.2, 5.7.5 | The dist checks on every program GoReleaser builds: in `release-check` on every push, in the release job before signing, and on the published files by 5.7.5 and V2 (note 5.7.1, 4) |
| The wording in `README.md`, `CONTRIBUTING.md`, `.gitattributes`, `CHANGELOG.md` (gate §5) | 5.7.4 | Each rewritten for 1.0; `SECURITY.md`, `--help` and the Makefile with them; a stale-phrase grep finds nothing |
| Rohan's at 5.7: the `release` environment and a new tap token (gate §5; `STATE.md`; spec §17 step 4's note) | Batch 1; decision 1 | The environment by the note's working form, open to `v*` tags; no new token under 1 (A); the old token's revocation confirmed |
| Actions pinned by commit (spec §12 step 8) | 5.7.0, 5.7.2 | Every `uses:` at a full commit with its version; a test holds every workflow file, the pack template's included |
| GoReleaser pinned (§12 step 8) | 5.7.1 | An exact version; how its download is checked recorded (note 5.7.1, 1) |
| The re-release input removed (§12 step 8) | 5.7.2 | `release.yml` runs on tags only: no `workflow_dispatch` at all (the old input went in part 1; the manual build goes now, its rehearsal moving to `release-check`) |
| Immutable releases (§12 step 8) | Batch 1; 5.7.2 | A repository setting, so Rohan's line, read back; the `publish` job makes a draft, checks it, publishes it only after signing, then checks the release as published |
| Build provenance (§12 step 8) | 5.7.2, 5.7.3 | GitHub's attestation over every archive, `checksums.txt` and every program; the install lines check it |
| "Projects' CI pins the archive's SHA-256" (§12 step 8; spec §5's pack CI, step 2) | 5.7.6 | The template's `pack.yml`, base's `ci` skill and the workflow repo's `pack.yml` take the 1.0 archive, checked by its SHA-256 |
| `bonsai@0.4` and `brew install bonsai` moving (§12 step 7; outline "Builds"); "the Homebrew tap is written only by the release workflow" (outline, "Risks") | Decision 1 | Under (A): the tap untouched, no token; under (B): GoReleaser's formula and `bonsai@0.4`, the token inside the `release` environment (note 5.7.1, 2) |
| "A Homebrew install ... lands in a folder the line never looks in, so 5.7 must settle Homebrew" (5.3's (a)); "Homebrew's place against 5.3's (a) stays 5.7's" (5.6) | Decision 1 | Stated for Rohan; (A) recommended |
| `gh workflow enable release.yml`; the optional tag ruleset (§17 step 4; outline) | Batch 1; decision 4 | His enable line; his answer 4 (A): the environment's approval, no ruleset on creating tags |
| The release files person-only, "the release files 5.7 changes with Rohan" (5.4, the switch) | This section | Its approval is his word for the changes named here to `release.yml` and `.goreleaser.yaml`, under the stricter-only rule |
| "At 1.0, with `bonsai.exe` installed, a climb in a Windows clone can replace check 10's Windows half; 5.7's section decides" (5.4) | This section | Not for 1.0: check 10's Windows half and CI's `windows` job stay; a Windows climb would need a machine-specific script and its result is not what the stop gate reads |
| The way back if a new install blocks work (5.4) | 5.7.5 | The 5.4 pre-release kept in `~/bonsai-checks/prerelease/previous/` before his WSL install, its number from 5.4.7's report; 5.4's two lines are the way back, never `--remove` in WSL |
| Archives hold `install.sh` and `install.ps1` at their top (5.6, "Hand-offs to 5.7") | 5.7.1 | `archives.files` with `strip_parent`; the dist checks hold each archive's files |
| 5.6.5's test holds the lines' names to `.goreleaser.yaml` (5.6) | 5.7.1, 5.7.3 | The name template unchanged; the test stays green; its table gains the check line |
| His 1.0 installs through the installers, the UAC prompt's first real run, with its way out (5.6) | Batches 3 and 5; 5.7.5 | His first real installs are rc.1's (batch 3), the UAC prompt's first real run among them; then 1.0's (batch 5); the way out on each side written |
| His memory import line and the agent's check that it loads (5.6) | Batch 5; 5.7.5 | His two lines at 1.0, the index made by rc.1's first `--yes` in Bonsai's repo; the check as 5.6 wrote it (note 5.7.5, 7) |
| The install lines' stronger check; `gh attestation verify`'s login (5.6) | 5.7.2, 5.7.3; decision 2 | Measured: it needs `gh` logged in; 2.97 is the floor (`--signer-workflow` came in 2.51, `--source-ref` in 2.68, and 2.97 fixed GHSA-mm27-mwq9-fr5g, a signer match built without escaping); Windows' is 2.102, WSL's updated by Rohan on 10 Oct; signed records for archives and programs; `--deny-self-hosted-runners` so "on GitHub's own machines" is checked; the check runs before anything from the download |
| `-ForceChild` "for 5.7's verifier to read" (5.6.6 note 4) | V1 | Read in the shipped `install.ps1`: documented as for tests, refused unless already elevated, raising no prompt, skipping no step |
| `check`'s comparison "waits for `install.json`" until the first real install (5.6; spec §3) | 5.7.5 | His 1.0 install writes it; `check` in Bonsai's repo then reads it with no `bonsai-path` note |
| The machine settings reader's stricter rule "reaches the installed guard only with 1.0's install" (note 5.6.1, 1) | 5.7.5 | Nothing waits on it (Bonsai's repo has no machine settings); `check` there reads clean on 1.0 |
| The packs' first tags with 1.0; tag lines "typable without a clone of the workflow repo (for example through `gh`)", and "5.7 checks that a tag made that way starts the repo's `release` job" (5.5) | Batch 5; 5.7.6 | Every tag by `gh api .../git/refs`; rc.1 proves a tag made so starts a tag workflow in Bonsai's repo, and 5.7.6 checks the workflow repo's release after its tag; under his answer 5 (A) neither pack tag can then move |
| The template's pinned `bonsai` moves to the release archive; `base`'s `version` becomes `1.0.0`; base's `ci` skill takes the archive "after" 1.0 (5.5) | 5.7.6 | Note 5.7.6, 1-2 |
| "Whether 1.0's template names `base-v1.0.0` is 5.7's" (5.5.0 note 5) | 5.7.6 | No: `init` keeps `packs: []` for 5.5.0's reasons (offline, nothing installed unasked); the template's comment shows `base` at `base-v1.0.0` as the line to add |
| "Nobody tags or releases. `release.yml` stays disabled" (`CLAUDE.md`) | 5.7.2; his part | Rewritten for after 1.0 (note 5.7.2, 6), and said to Rohan in one line ("What stays yours") |
| A tag with no published release (waiting for approval, a run failed before publishing, a stray tag) read by 5.6.5 as a newer release whose lines fail (the review) | 5.7.2; a hand-off back to 5.6 | The failure path deletes the tag (Rohan's line) and tags the fixed commit again; the run deletes its own draft; named in the risks; 5.6.5 to count only a tag whose release is published ("Stale or in tension") |
| The issue and pull request templates ("Homebrew / binary download / `go install`", "Select ...") going public with 1.0 (the review) | 5.7.2 | Brought to 1.0 under its grant, `.github/**` being protected (note 5.7.2, 5) |
| `base-v*` and the workflow repo's tags left unlocked (the review) | Decision 5; batch 1 | Under (A), tag rules against moving or deleting them, and the lock on the workflow repo's releases |
| GoReleaser reading a `base-v*` tag as the current tag in `release-check` (the review, from memory, not measured) | 5.7.1 | `git.ignore_tags` (or the current tag set by hand), proved on a doctored copy (note 5.7.1, 2 and 7) |
| The old tap token "not confirmed" revoked (spec §17 step 4's note) | Batch 1 | His look on github.com, his words in the run report |
| §14 row 10, "the first public release, on Rohan's word" | Batch 5 | His word after the trial on Bonsai (5.7.7 to 5.7.10), his review of its analysis and his own trial on a fresh project |
| The floors at a part's end (5.4, "Tasks and names") | After 5.7.6 | One climb of `main` by the installed 1.0, the floors raised under a `T-579x` task, the numbers in Rohan's last line |
| GitHub Pages, the old website ("whenever you like") | Not 5.7 | Unchanged; his whenever he likes |
| Rohan's trial of Bonsai on itself on `1.0.0-rc.1` before 1.0 (10 Oct: "bonsai should work on itself cleanly") | 5.7.7 to 5.7.10; batches 2-4 | rc.1 installed by him on both sides; his pick of two or three features run through the whole pipeline; the analysis; fix rounds as further pre-releases |
| His own trial on a fresh project, "independently once im satisfied with the trial on bonsai" | His; 5.7.4 | His time, untimed; the README's "first project" steps are what the plan gives him; agents help only if he asks |
| His future items with 1 (A): "support for mac" and "adding bonsai on official distributers for each platform" | "Outside step 5" | Each named with why it is not in 1.0 |

#### Notes per piece

**5.7.0, CI's own checks, and Go held.**
1. **Go, already moved** (Rohan, 10 Oct: "we can go with go 1.27 latest"): the toolchain went to Go 1.27's newest patch
   as its own piece before 5.7, the `go 1.25.0` line kept (the language level, so `CLAUDE.md`'s "the module's go 1.25
   with its toolchain line" still reads true), govulncheck and, if it had to, golangci-lint re-pinned; its run report
   holds check 10's counts before and after. 5.7.0 checks it holds: `go.mod`'s `toolchain` is Go 1.27's newest patch
   at this piece's start (a newer one is the one-line bump, landed here); each CI job prints `go version` first, so the
   log shows the toolchain setup-go installed from `go.mod`. Go 1.27 is fixed until Go 1.29 comes out (about August
   2027); the end verifier's note names that date.
2. **govulncheck:** locally, `govulncheck -show verbose ./...` on the toolchain in use: no finding in the standard
   library, reached or not, and none in a module. A finding with no fixed Go 1.27 patch yet is named in the run report
   with its id and why it is not reached; the orchestrator weighs it before V1, and it goes to Rohan if it is reached.
3. **Lint:** a second `golangci-lint` step with `--build-tags=bonsai_test_fault` beside the plain one (gate §5), both
   required. **`actionlint`** in the same job, run as `go run github.com/rhysd/actionlint/cmd/actionlint@<exact
   version>` (Go's checksum database checks what it fetches), over every file in `.github/workflows/` and
   `packs/template/.github/workflows/`.
4. **Pins:** every `uses:` in `ci.yml` and `codeql.yml` written `owner/repo@<40 hex> # vX.Y.Z`, the commit read from
   the action's release tag with `git ls-remote` (the tag peeled to its commit) and the version the exact release that
   commit is; every `actions/checkout` with `persist-credentials: false` (no job here pushes). **The test**
   (`internal/selfcheck/workflows_test.go`, from the module root): every `uses:` in those folders is a local action
   (`./...`) or names 40 hex characters followed by a `# v` comment; anything else fails, naming the file and line.
   Dependabot stays: it proposes pin moves as pull requests, which no agent merges or closes; the orchestrator reads the
   action's change, lands the same move as its own commit with CI green, and Dependabot closes its own pull request once
   `main` holds it.
5. **Proof** beyond the table: the pin test failing on a temporary copy of `ci.yml` with `actions/checkout@v6` and on
   one with a 40-hex pin and no comment; `actionlint` failing on a temporary copy with a misspelt key.

**5.7.1, the release build and its rehearsal.**
1. **GoReleaser pinned:** `goreleaser/goreleaser-action` at a commit, `version:` an exact `v2.x.y` (the newest at the
   piece's start), `distribution: goreleaser`. The builder reads how that action fetches and checks GoReleaser at that
   version and records it; if it checks nothing, the job runs `go run github.com/goreleaser/goreleaser/v2@<the same
   version>` instead, so Go's checksum database checks it. `goreleaser check` passes on the config.
2. **`.goreleaser.yaml`**, every key commented:
   - **builds:** as today, plus `flags: [-trimpath]`; ldflags kept (`-s -w -X main.version={{.Version}}`, no leading
     `v`), which 5.6.5's lines and 5.2.2's `--version` read; GoReleaser builds in the checkout, so `vcs.revision` and
     `vcs.modified=false` are stamped; `CGO_ENABLED=0`; linux, windows and darwin on amd64 and arm64 (spec §3: macOS as
     today, untested).
   - **archives:** the name template unchanged (5.6.5's test reads it); `files`: `LICENSE`, `README.md`, and
     `install/install.sh` and `install/install.ps1` at the archive's top (`strip_parent: true`), `install.sh` with mode
     0755 in the tar archives. Each archive carries both installers (a zip's `install.sh` is harmless; one archive entry
     keeps one name template).
   - **checksum:** `checksums.txt`, SHA-256, as today.
   - **changelog:** `disable: true`; the notes come from `CHANGELOG.md` (item 3), so no commit title becomes release
     text.
   - **release:** `disable: true`. GoReleaser builds, archives and writes `checksums.txt`, and makes no GitHub release:
     5.7.2's `publish` job does that with `gh`, so GoReleaser never runs where a release can be written or a record
     signed, and needs no token (the builder confirms at the pinned version that it asks for none; else
     `--skip=publish`).
   - **git:** `ignore_tags: ["base-v*"]`, so a `base-v*` tag on or before a commit is never taken for the current or
     previous tag: from the pack tags on, GoReleaser would otherwise meet a tag it cannot read as a version on every
     push (the review's point, from memory; the builder measures it on a doctored copy, item 7). If the pinned
     open-source version takes no pattern there, each job sets the current tag itself (`GORELEASER_CURRENT_TAG`).
   - **Homebrew** (Rohan's answer 1 (A)): `brews:` removed; nothing names the tap or a token. **Under 1 (B), not chosen
     and kept as the record:** the formula by what
     the pinned GoReleaser supports (its documentation lists `brews` as deprecated; the builder reads its replacement
     and records which runs on Linux), written into `dist/` and never pushed by GoReleaser (`skip_upload: true`); and
     `bonsai@0.4`: a file in Bonsai's repo, `packaging/homebrew/bonsai@0.4.rb` (0.4.3's URLs and SHA-256s from today's
     `Formula/bonsai.rb`, `class BonsaiAT04`, `keg_only :versioned_formula`). The `publish` job pushes both (note 5.7.2,
     4). That is (B)'s 0.5-1 hour, all of it in this piece.
3. **The notes:** `.github/scripts/release-notes.sh <tag> <out>` (POSIX `sh`, LF, its header documenting it): for a tag
   `vX.Y.Z` or `vX.Y.Z-rc.N`, the section of `CHANGELOG.md` whose heading starts `## [X.Y.Z]`, up to the next `## `,
   written to `<out>`; no such section, or an empty one, fails naming the file. `<out>` is always outside the checkout
   (`$RUNNER_TEMP` in a job): an untracked file there would make GoReleaser call the tree dirty and stamp
   `vcs.modified=true`. 5.7.2's `publish` job gives it to `gh release create --notes-file`. `release-check` runs it for
   the newest section, so a broken changelog shows on the push that broke it.
4. **The dist checks** (`internal/selfcheck/dist_test.go`), run only when `BONSAI_CHECK_DIST` names a folder of
   archives and `checksums.txt` (GoReleaser's `dist/`, or a folder of downloaded release files), skipped otherwise with
   that reason; each archive unpacked into a temporary folder:
   - **each program:** none of the fault markers in its bytes (their one home moves from `cmd/bonsai/hook_test.go` to a
     place both tests read; `TestNormalBuildHasNoFaultCode` keeps proving the markers are visible in a fault build); its
     build settings read with `debug/buildinfo`: no `-tags`, `CGO_ENABLED=0`, `-trimpath=true`, `vcs.modified=false`,
     a `vcs.revision`, equal to `BONSAI_CHECK_COMMIT` when that is set;
   - **each archive:** exactly the program (`bonsai`, or `bonsai.exe` in a zip), `install.sh` (mode 0755 in a tar),
     `install.ps1`, `README.md` and `LICENSE`, all at its top; named as `.goreleaser.yaml`'s template gives;
   - **`checksums.txt`:** one line per archive, each SHA-256 right, nothing else;
   - **the installer from its archive:** on Linux the linux/amd64 archive's `install.sh --target <temp>/bonsai` with a
     temporary `BONSAI_HOME` and 5.6.6's stub `sudo` first on the PATH (it fails the test if called); `install.json`
     read back by `check`'s reader; the installed `--version` naming the version. On Windows, the windows/amd64 zip's
     `install.ps1 -Target` the same way (no prompt: a scratch target needs none), against a snapshot built in WSL and
     copied to `%USERPROFILE%\bonsai-checks\`.
5. **`release-check`** (a `ci.yml` job on `ubuntu-latest`, every push and pull request): full history; the pinned
   GoReleaser `release --snapshot --clean` (it publishes nothing; the job's token is `contents: read`); the notes
   script into `$RUNNER_TEMP`; `BONSAI_CHECK_DIST=dist go test -count=1 ./internal/selfcheck/`. It uploads nothing.
6. **`install.sh` off Linux:** it refuses on anything `uname -s` does not call `Linux`, with one sentence (macOS builds
   are untested and have no installer: copy `bonsai` to a folder only root writes), before anything else; 5.6.6's test
   gains the case with a stub `uname`.
7. **Proof** beyond the table: a local snapshot built in WSL, the dist checks green on it in WSL and, its files copied
   to `%USERPROFILE%\bonsai-checks\`, natively on Windows (the dist test run from a Windows-git clone of the commit with
   Windows Go); and failing on doctored copies of a snapshot: a program built with `-tags bonsai_test_fault`, an archive
   without `install.sh`, an archive with one file more, a wrong line in `checksums.txt`, a program built from a dirty
   tree. And a scratch clone with a scratch `base-v9.9.9` tag on its HEAD (and one on an older commit): its snapshot
   still builds and passes them. GoReleaser runs locally as `go run github.com/goreleaser/goreleaser/v2@<the pinned
   version>`, never installed.

**5.7.2, the release workflow** (`release.yml`, every key commented, its header saying what starts it, what each job
may write and who switches it on).
1. **What starts it:** `on: push: tags: ['v[0-9]+.[0-9]+.[0-9]+', 'v[0-9]+.[0-9]+.[0-9]+-rc.[0-9]+']` and nothing else:
   no `workflow_dispatch`, no other event, so no manual run and no input can publish or re-publish. A `base-v*` tag
   matches neither pattern and starts nothing.
2. **Two jobs, so nothing the build makes runs where a release can be written or a record signed.** Top-level
   `permissions: {}`; `concurrency: release-${{ github.ref }}`, never cancelled; both on `ubuntu-latest` (GitHub's own
   machines, which the signed record names and `--deny-self-hosted-runners` checks). A token is given to a step only in
   that step's `env`, and only to the steps listed as using it; never at job level.
   - **`build`:** no environment; permissions `contents: read`, `checks: read`, `actions: read`. Its steps:
     1. checkout at the tag, full history, `persist-credentials: false`;
     2. **the tag checked** by `.github/scripts/release-tag-check.sh` (POSIX `sh`, its header documenting it; the one
        step given `GH_TOKEN`): its name against `^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$` again; its commit on `main`
        (`git merge-base --is-ancestor "$GITHUB_SHA" origin/main`); and the commit's checks, read from
        `repos/LastStep/Bonsai/commits/<sha>/check-runs?per_page=100`: CI's `test`, `windows`, `lint`, `govulncheck`,
        `release-check`, `packs` and `install`, and CodeQL's `Analyze Go`, each `completed` with `success`; every other
        check run neither `failure`, `cancelled` nor `timed_out` (one `neutral`, `skipped` or still running passes);
        this run's own check suite left out, by its `check_suite_id` read from `actions/runs/$GITHUB_RUN_ID` (the
        reason for `actions: read`). A named job missing, or a run failed, stops the release, naming it;
     3. setup-go from `go.mod` with `cache: false` (no cache another run could have written);
     4. `go test ./...` and `go vet ./...`, plain and tagged;
     5. the notes (5.7.1's script) into `$RUNNER_TEMP`;
     6. GoReleaser (5.7.1's pinned way), no token: the programs, archives and `checksums.txt` in `dist/`;
     7. the dist checks on `dist/` with `BONSAI_CHECK_COMMIT=$GITHUB_SHA` (they run the built program and `install.sh`:
        here, where nothing can be written to GitHub and no signing token can be asked for);
     8. the archives, `checksums.txt`, the programs and the notes uploaded as one artifact (the upload action pinned);
        `checksums.txt`'s SHA-256 as the job's output.
   - **`publish`:** `needs: build`; `environment: release` (it waits for Rohan's approval before any step runs, his
     answer 4 (A)); permissions `contents: write`, `id-token: write`, `attestations: write`; it runs no file from the
     build. Its steps:
     1. the artifact downloaded (the download action pinned); `checksums.txt`'s SHA-256 equal to `build`'s output, and
        each archive's equal to its line in `checksums.txt`;
     2. **the signed record:** GitHub's attestation action, pinned, over every archive, `checksums.txt` and every
        program;
     3. **a draft** (`GH_TOKEN`): `gh release create "$GITHUB_REF_NAME" --draft --verify-tag --title "Bonsai <version>"
        --notes-file <notes>` with every archive and `checksums.txt`; for an `-rc.N` tag `--prerelease --latest=false`;
     4. **the draft checked before anyone sees it** (`GH_TOKEN`): every file as GitHub holds it equal, by SHA-256, to
        the one in hand. The builder's first measurement is whether `gh release download` reads a draft; if not, each
        asset's `digest` in the release's asset list; failing that, each asset fetched through the API with the token.
        Then `gh attestation verify` on each archive and program in hand, with `--signer-workflow
        LastStep/Bonsai/.github/workflows/release.yml --source-ref "$GITHUB_REF" --deny-self-hosted-runners`;
     5. **publish** (`GH_TOKEN`): `gh release edit "$GITHUB_REF_NAME" --draft=false`; GitHub's lock then holds it;
     6. **checked as published** (`GH_TOKEN`): `gh release verify "$GITHUB_REF_NAME"` and `gh release verify-asset` for
        each archive, retried for up to a minute while GitHub writes its own record of the release. The run's log is
        the release's first record;
     7. on a failure before step 5, a last step (`if: failure()`, `GH_TOKEN`) deletes this run's own draft, so no
        half-made release is left.
3. **When a run fails.** Before publishing, nothing is public but the tag: the fix lands on `main`, Rohan's line deletes
   the tag (`gh api -X DELETE repos/LastStep/Bonsai/git/refs/tags/<tag>`, his batch 2 text) and he tags the fixed commit
   with the same name, so no number is lost; until then 5.6.5's newer-release read may take the tag for a release whose
   files are not there (risks below; a hand-off back to 5.6). At step 6, after publishing: an alarm about a public
   release: the orchestrator sends Rohan the log at once, and nothing is installed from it until the cause is known; a
   defect found after publishing is fixed under the next number. A failure outside the code (GitHub down) is re-run by
   Rohan's line, `gh run rerun <id> -R LastStep/Bonsai`, which the orchestrator sends; no agent re-runs, cancels or
   approves a run.
4. **Under 1 (B), not chosen, kept as the record:** a last step of `publish`, after step 6 and the only step given
   `HOMEBREW_TAP_TOKEN` (from the environment), pushes the formula GoReleaser wrote and, when the tap lacks it,
   `bonsai@0.4.rb` (note 5.7.1, 2): files from the build, but text, written to another repo, never run.
5. **The issue and pull request templates** (`.github/ISSUE_TEMPLATE/`, `.github/pull_request_template.md`), public
   with 1.0 and under `.github/**`, so this piece's grant: the bug report asks for `bonsai --version`, the side (WSL,
   Linux or Windows), how it was installed (the installer, or by hand) and Claude Code's version; no `go install`, no
   Homebrew under 1 (A), no "Select ..." left from the old product. Every action pinned (5.7.0's test covers this file).
6. **`CLAUDE.md`'s GitHub lines:** "Nobody tags or releases. `release.yml` stays disabled; releases are Rohan's word at
   spec step 5.7" becomes: only Rohan tags (`v*`, `base-v*`, the workflow repo's) and approves a release run;
   `release.yml` runs only on his tag, its publishing inside the `release` environment; no agent tags, approves, re-runs
   or cancels a release run, or changes a setting, secret, ruleset or workflow switch; a release is published by the
   workflow, never by hand. True from the landing (before batch 1 the workflow is still off on GitHub). Rohan's part
   says it in one line ("What stays yours").
7. **Proof** beyond the table: the builder runs the tag-check script locally against Bonsai's own history with
   `GITHUB_SHA` set by hand: a commit on `main` passes, a commit on a scratch branch fails, a name `v1.0` or
   `v1.0.0-beta` fails. Its checks query runs read-only against a real `main` commit while that commit's CI is still
   running (a named job still running fails it), and against a saved copy of a real answer (the script reads one when
   given it, for this test) with one more run marked `in_progress`: left out when its suite is the run's own, a failure
   when it is a named job's. Nothing else can run before a real tag; V1 reads every line, the token scoping step by
   step.

**5.7.3, the stronger check in the install lines.**
1. **The line:** 5.6.5's table gains, on each side, after the fingerprint line and before unpacking, `gh attestation
   verify <archive> -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref
   refs/tags/v<version> --deny-self-hosted-runners` (the same words in bash and PowerShell), so seven lines a side,
   every part fixed but the checked version (5.6.5 note 5). Why before unpacking: the installer is inside the archive
   and runs with the person's password or administrator rights, so nothing from the archive runs before its origin is
   checked. Why `--deny-self-hosted-runners`: without it a record made on a machine of anyone's own would pass, and
   "built on GitHub's own machines" would be a claim, not a check.
2. **What it needs, and proves:** `gh` logged in (measured 10 Oct: no login exits 4, the wrong repository 1, success
   0), and **2.97 or newer**: `--signer-workflow` came in 2.51 and `--source-ref` in 2.68, and 2.97 (31 Jul 2026) fixed
   GHSA-mm27-mwq9-fr5g, a matcher built from `--signer-workflow` without escaping, so a lookalike name could pass (read
   from `cli/cli`'s advisory and release notes; only 2.102 was run). It proves the archive was built by `release.yml` in
   `LastStep/Bonsai`, for that tag, on GitHub's own machines. Its human words (`Verification succeeded`) show only in a
   terminal: the agent sees them in a `tmux` pane on WSL and records them; on Windows it records the exit code, and
   Rohan's real PowerShell shows the words.
3. **`docs/install.md`** (new): the lines for 1.0.0 on both sides; what each line proves; what to do when one fails; the
   way out on each side (WSL: never `--remove` while a project relies on the guard; the build before instead, after
   which `check` flags the install record until the next install); and how to check an installed copy at any time (`gh
   attestation verify /usr/local/bin/bonsai -R LastStep/Bonsai --signer-workflow
   LastStep/Bonsai/.github/workflows/release.yml --deny-self-hosted-runners`, and the Windows path), since every program
   is signed too (5.7.2). Its test gives the table 1.0.0 and holds the page's two blocks to the lines, byte for byte. It
   names `gh` 2.97 or newer, logged in, as needs; and that `gh`, like `sh` and `tar`, is found on the PATH, so `which -a
   gh` should name the system's own first.
4. **The operating skill's install line** names `docs/install.md` and `bonsai status --full --json`'s
   `bonsai_release.lines`; the agent hands the lines over and never runs them (unchanged).
5. **Proof** beyond the table: an agent runs the line's form on both sides against `gh`'s own signed release file (as
   measured above), against the same file with the wrong repository, and with `--deny-self-hosted-runners`, recording
   words and exit codes on the `gh` versions in use (2.102.0 on both sides on 10 Oct; batch 1 reads WSL's back
   again).
6. **Decision 2's (B), written out for the record** (not chosen; V1 checks it would have run): from WSL, Windows' `gh`
   by its full path, the archive's path translated for it:

   ```bash
   "/mnt/c/Program Files/GitHub CLI/gh.exe" attestation verify "$(wslpath -w bonsai_1.0.0_linux_amd64.tar.gz)" -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.0 --deny-self-hosted-runners
   ```

**5.7.4, the words for 1.0.**
1. **README:** what Bonsai is and is not (from the one-pager, in plain words); install and check it (a pointer to
   `docs/install.md`; the two installed places and why the guard needs them; `gh` for the check); **a "first project"
   section**, the steps Rohan's own trial follows (a new folder and git; `bonsai init` naming `base` by its commit until
   `base-v1.0.0` exists; `bonsai check` and `status`; Claude Code opened there, the folder trusted, something small
   asked for; `bonsai ladder` once a rung is added), which an agent runs once as written in a scratch folder with the
   installed pre-release before Rohan's own trial, writing what it met in 5.7.9's records; then `bonsai --help --json`
   for agents; the packs (`base`, the
   template, `workflow`); platforms (Linux and WSL, Windows; macOS built, untested, with no installer, so a Mac user
   copies the program by hand); the old product (0.4.3 at its tag; Homebrew as decision 1 leaves it); where the design
   lives; licence. ASCII.
2. **CONTRIBUTING:** Go (the `toolchain` line; an older Go downloads it); the proof is `bonsai ladder`, with CI and the
   native Windows run beside it; a clone without Bonsai installed has every Claude Code edit refused (5.4's line, kept);
   releases are Rohan's tags; `golang.org/x/sys` described as it stands.
3. **CHANGELOG:** "Rebuild - Unreleased" replaced by `## [1.0.0] - Unreleased` (the orchestrator's commit dates it at
   Rohan's word, 5.7.5): what 1.0 is in a few lines, then Added, the install pointer, and what a 0.4.3 workspace meets
   (the builder states what the code does, from the spec). The unreleased 0.5.0 section and the older ones stay, under
   one line saying the rebuild does not carry them.
4. **SECURITY:** in scope for 1.0 (the guard and hook path, the installers, the release and its files); supported
   versions (the newest 1.x); 0.4.3 unsupported; private reports as now.
5. **`.gitattributes`:** the old product's lines and comments out (`*.sh.tmpl`, generated hooks, preview goldens); each
   kept line's comment true; `formats/** -text` stays last.
6. **`bonsai --help` and `main.go`'s doc comment:** no "rebuild in progress" or "what has been built so far"; their
   tests updated.
7. **Makefile:** `make build` with `-trimpath`, as the release builds.
8. **The greps:** a stale-phrase list ("being rebuilt", "rebuild in progress", "stub", "no release yet", a root
   `STATE.md`, `design/plan.md` where step 5's plan is meant, "0.4.3, the latest") over the repo but `records/` and
   `design/`, each hit read by the orchestrator; the privacy grep of the diff ("How it is proved").

**5.7.5, the releases, checked** (each pre-release, then 1.0).
1. **Before each tag:** the orchestrator names the commit: `main`'s, V1 passed at it or before it, with `git diff
   --stat <V1's commit>..<commit>` holding no change to the release path (`.github/**`, `.goreleaser.yaml`, `install/`,
   the dist checks) that V1's brief has not read again, and its CI green, `release-check` included. For 1.0: the last
   pre-release's commit plus only the notes' date (item 4).
2. **After each tag:** the orchestrator reads the run's log; a Sonnet agent, in a fresh scratch folder on each side,
   downloads every file of the release and runs: `gh release verify` and `gh release verify-asset` for each; `gh
   attestation verify` for each archive and program (the signer workflow, the tag's ref and
   `--deny-self-hosted-runners`); `sha256sum -c` against `checksums.txt`; the dist checks with `BONSAI_CHECK_DIST` on
   the download and `BONSAI_CHECK_COMMIT` the tag's commit; each side's installer from the real archive at a scratch
   target with a scratch `BONSAI_HOME` (on Windows `-Target`, so no prompt); **a rebuild compared:** a clean clone at
   the tag, `go build` with the release's flags for linux/amd64 and windows/amd64 (`go version -m` of the release's
   program gives them), its SHA-256 against the release's program; a difference is shown setting by setting from `go
   version -m` side by side and is a finding unless explained; and a plain scratch build's `status --full --json` on a
   scratch home reading the real tags (a pre-release changes nothing; after 1.0, `newest` reads 1.0.0). The tap read
   before and after (`pushed_at`, `Formula/bonsai.rb`'s SHA): unchanged.
3. **V2** (fresh Opus) runs item 2's checks itself on rc.1's files and reads the run's log, before Rohan installs rc.1.
   A later pre-release gets V2's brief again (a fresh agent) only if the release path changed since rc.1.
4. **Rohan's word for 1.0,** after his own trial: the orchestrator sends the notes' link. On his word, its own commit
   under a `T-579x` task dates the heading (`## [1.0.0] - <date>`); `git diff --stat` from the last pre-release's commit
   holds only that line, `records/` and `.bonsai/STATE.md`; CI green; then batch 5's tag line with that commit.
5. **The install lines.** Bonsai's table gives lines only for a version of three whole numbers (5.6.5 note 5), so a
   pre-release's lines are `docs/install.md`'s with the version written in by the orchestrator, and an agent checks
   each against the release's file names and its tag before they go out; 1.0's are `docs/install.md`'s as they stand.
   Before Rohan's first WSL install (rc.1): `/usr/local/bin/bonsai` (the 5.4 pre-release) copied to
   `~/bonsai-checks/prerelease/previous/bonsai`, its SHA-256 matched to 5.4.7's run report; before each later one, the
   installed rc copied there the same way, its SHA-256 from `install.json` and the run report.
6. **After each WSL install:** his output read; `/usr/local/bin/bonsai --version`, its owner and mode, `install.json`,
   `which -a bonsai`; then in `~/Servers/Bonsai`, `/usr/local/bin/bonsai check --json`: `bonsai-path` with no note, and
   either nothing or the finding for Bonsai's own lines out of date, which sends him the (ii) lines (`cd
   ~/Servers/Bonsai`, then `/usr/local/bin/bonsai update --allow-exec --yes`, by its full path). Otherwise the
   orchestrator runs `bonsai update --json` there and reads the preview: when it changes nothing in the project,
   `bonsai update --yes --json` writes only the home's empty personal index; when it changes Bonsai's own files, 5.5.7
   note 4's rule holds (a task granting `bonsai.yaml` and the lock, the preview read line by line, anything under "Runs
   code" his line). At 1.0, once `~/.bonsai/personal/INDEX.md` exists (rc.1 made it), his memory line goes out.
7. **The memory check**, after his `tail` output at 1.0, exactly as 5.6's "Hand-offs to 5.7" writes it: a Sonnet
   agent, a scratch project, an interactive `tmux` session through `claude-here` (which reads his real
   `~/.claude/CLAUDE.md`), `/memory` listing `~/.bonsai/personal/INDEX.md` (or the session asked for the index's
   `title:`, `Personal memory`), any external-import prompt recorded with its words, nothing of the index copied into a
   report, Claude Code's version and the user settings hashes before and after.
8. **After each Windows install** (rc.1, then 1.0): an agent reads `Get-Command bonsai -All`, `install.json`,
   `--version`, and runs `gh attestation verify` on `C:\Program Files\Bonsai\bonsai.exe`; after rc.1, check 10's
   Windows half runs once more and its counts are compared with the last run's (a test that changes once `bonsai.exe` is
   installed is named with its reason).
9. **Bonsai's repo on each install:** from his WSL install the ladder is climbed by that copy; the first climb's counts
   are read against the floors (a fall is a finding before any floor moves).
10. **Processes:** each agent stops what it started (`tmux` sessions, downloads) and checks with `ps`; the orchestrator
    sweeps after each.

**5.7.6, the packs at 1.0** (spec §5, the pack's CI, step 2: "The pinned `bonsai` comes from a Bonsai release
archive checked by its SHA-256").
1. **In Bonsai's repo:** the template's `pack.yml` replaces its clone and build of `BONSAI_COMMIT` with the 1.0 archive
   for the runner's system (`bonsai_1.0.0_linux_amd64.tar.gz`; `bonsai_1.0.0_windows_amd64.zip`), downloaded from the
   fixed release URL and checked against the SHA-256 written beside it (from `checksums.txt`, checked by 5.7.5), each
   with a comment saying how to move it; `ci/check.sh` takes the `bonsai` it is given; base's `ci` skill the same for
   Linux; base's `pack.yaml` at `version: "1.0.0"` (protected: its task's grant names it); the engine's `bonsai.yaml`
   template keeps `packs: []`, its comment showing `base` at `ref: base-v1.0.0` as the lines to add. Bonsai's own
   `packs` job still builds `bonsai` from the commit it checks, so the template never drifts from the engine.
2. **In the workflow repo** (its clone under `~/bonsai-checks/`, as 5.5 made it): its `pack.yml` as the template's new
   one; its `bonsai/pack.yaml` at `version: "1.0.0"`; one commit, pushed by the orchestrator, its CI green on both
   sides.
3. **The tag lines** for batch 5's end name those two commits. After them: the workflow repo's `release` job green and
   its release made (`gh release view v1.0.0 -R LastStep/bonsai-workflow`); no run started by `base-v1.0.0` in Bonsai's
   repo (`gh run list -R LastStep/Bonsai -L 10`).
4. **Proof** beyond the table: the template's checks on both sides with the archive (a temporary copy with a wrong
   SHA-256 fails before running anything); `check --pack` on base at 1.0.0.

**5.7.7, Bonsai on itself on rc.1** (Rohan, 10 Oct: "the main point is that bonsai should work on itself cleanly").
1. **Set up:** once rc.1 is installed and read back (note 5.7.5, 6), Bonsai's repo runs it: `check` clean; Bonsai's
   own lines brought to rc.1's (his (ii) line if they changed); a climb of `main` by rc.1, green, its counts against the
   floors. Nothing else is `running` from here until the trial's feature lands (5.4's one-task rule, and so that the
   log, the sessions table and the trial's records hold only the trial).
2. **The proposals:** two or three small real features for Bonsai, from what the plan has left for later or what 5.1
   to 5.6 named as worth having, each in plain words for Rohan: what it does for him, what it touches, its hours
   (inside 1-3; one above that is said with the new total), and which parts of the pipeline it puts to work. Between
   them they reach: a protected file under a task's grant; a new test that must fail on the code before it (the
   ladder's new-tests rung and a Bless); a question to the person through Bonsai's asks, where the work truly has one;
   and the tables rebuilt by `check --write`. None needs a format change (a removal or a new major) or touches the
   release path.
3. **The run:** Rohan's pick as an ordinary task, by the plan's own pipeline as 5.4 and 5.5.7 left it: the task file in
   `records/tasks/` (`T-5708`, its grants from the proposal), moved by the orchestrator; a worktree; the builder started
   as 5.4's named session (`BONSAI_TASK` and `BONSAI_ROLE=builder`), so the guard, the recorder and the stop gate bind
   it; its climbs by `/usr/local/bin/bonsai` (rc.1); a fresh verifier by `CLAUDE.md`'s rule; the landing rule; CI;
   `check --write` for the tables; the run report. The orchestrator changes nothing in how Bonsai works for the trial:
   a step that does not work is a finding, not a workaround.
4. **Its evidence**, kept for 5.7.9: the commits; the task file's moves; every session's log file (`.bonsai/local/`,
   read in place, never committed or copied whole); the sessions and tasks tables; each ladder result and its log
   record; every guard refusal and stop-gate block; each ask and its answer; the verifier's report; the minutes per
   run; the files Bonsai itself wrote or changed (in the repo, in `.bonsai/local/`, in `~/.bonsai/`), listed before and
   after with their sizes.

**5.7.8, the trial's feature.** Rohan's pick, built and proved by its own proposal's terms. Its builder by its kind
(Opus for a guard, hook or release change; Sonnet otherwise); its proof check 10, the ladder by rc.1 and CI, plus what
its kind asks. It lands on `main` as an ordinary commit and goes out with the next pre-release.

**5.7.9, the trial's analysis** (`records/trial-rc1.md`, then a short page per later round).
1. **For Rohan, in plain words with evidence:** what Bonsai created and changed, every file named (in the repo, in
   `.bonsai/local/`, in his Bonsai home), and why each exists; what the log and the tables hold, with a few lines quoted
   (nothing private, no secret: the log is already redacted); how the agents worked, run by run: who ran, how long, what
   each was refused and why, every question to him, every ladder result; what was rough, each with what the plan
   proposes (a fix in the next pre-release, a change after 1.0, or nothing, and why).
2. **Checked before he reads it:** each claim points at a file, a record or a command's output; a Haiku agent lists the
   files and records and the orchestrator compares them with the page; the privacy grep runs over it. The orchestrator
   may also publish it as a page he can read on his phone (an artifact), the same text.
3. **His reply** goes in the run report in his words: "happy", or the list of what to fix, each then a 5.7.10 task.

**5.7.10, a fix round** (`rc.2`, then `rc.3` and on).
1. Each fix a task of its own, by the pipeline, landed with its proof. A fix to the release path gets V1's brief again.
   A fix that would change a format by removal or a new major comes to Rohan first (his (B) on format changes); an
   addition does not.
2. The next pre-release (Rohan's tag and approval), 5.7.5's checks on its files, his WSL install (Windows only if a fix
   touches Windows), then the trial again, or only its affected part when the fix is narrow (the orchestrator says which
   and why), and a short analysis.
3. The hours of each round go against the re-ask line ("Hours"); the round that would cross it waits for Rohan's choice.

**V1, the release path.** A fresh Opus agent, once 5.7.0 to 5.7.4 have landed with CI green: reads every workflow file
and `.goreleaser.yaml` line by line against spec §12 step 8 and this section (the triggers; the two jobs, their
permissions, which steps get a token and that none gets one at job level; that `publish` runs no file from the build;
the environment; no `workflow_dispatch`; no cache in either job; `persist-credentials`; every pin and its comment;
GoReleaser's exact version, how its download is checked, that it needs no token and ignores `base-v*` tags; the tag
check and its query, the own suite left out; the order of sign, draft, draft checked, publish and verify; the draft's
deletion on failure); runs the snapshot and the dist checks itself on both sides, and breaks them with a doctored copy
each; reads the shipped `install.ps1`'s `-ForceChild` (documented as for tests, refused unless already elevated,
raising no prompt, skipping no step) and `install.sh`'s refusal off Linux; reads 5.7.3's lines (the 2.97 floor,
`--deny-self-hosted-runners`) and their test, and runs decision 2's (B) line once to see that it would have worked;
reads every line of Rohan's batch 1, and each read-back line after it, against GitHub's REST documentation for
environments (reviewers on a public repository among it), deployment branch policies, immutable releases, rulesets and
workflow states (dated), and the delete-tag line. It passes or fails; it fixes nothing; a must-fix is fixed forward and
read again by a fresh agent with V1's brief.

**V2, the published files.** A fresh Opus agent, after rc.1's run and before Rohan installs it: runs note 5.7.5, 2's
checks itself on both sides, reads the run's log step by step against note 5.7.2, 2, and passes or fails the files as
fit for his install. A later pre-release gets V2's brief again only if the release path changed since rc.1.

#### Proof for each piece

Every piece: its Go tests and `go vet`, plain and with the fault tag, in WSL (by the landing rule's green climb of its
task, run by `/usr/local/bin/bonsai`) and natively on Windows (check 10's Windows half) before the push, the counts in
the run report; CI green on the pushed commit, `release-check` among it from 5.7.1; the Windows rules of `CLAUDE.md`
read in the diff (forward slashes in every stored and printed path; byte-stable output; LF scripts; no test needing a
symbolic link or a file mode, the tar modes checked on the archive's own headers, not on a Windows file; Windows renames
with busy retries; no `bash` by name in a hook line); no Windows-only skip without a named reason; no test touching a
real home. Every GitHub change is Rohan's typed line, and the run report holds his words and what an agent read back.
Pieces that run Claude Code (5.7.5's memory check) record its version and the user settings hashes before and after.
Scripted runs live in `~/bonsai-checks/scripts/`, never committed. V1, V2 and the end verifier re-run what they judge
themselves.

#### 5.7 done

A fresh Opus verifier, at the end of 5.7, after 5.7.6, the pack tags and the floors, runs each check itself on the final
commit and on the published 1.0, on both sides where a check names them, and passes or fails 5.7:
1. **Go:** `go.mod`'s toolchain is Go 1.27's newest patch at release time, `go 1.25.0` kept, and the release's programs
   were built with it (`go version -m`); `govulncheck -show verbose ./...` shows no
   standard-library or module finding on it (or each named, unreached, with the orchestrator's recorded reason); each
   CI job's log names that Go.
2. **CI:** every `uses:` in every workflow file pinned to a full commit with its version (the test passes, and fails on
   a temporary copy with a tag); lint plain and tagged, `actionlint`, `release-check` and every other job green on the
   final commit.
3. **The release build:** a local snapshot (GoReleaser run as `go run github.com/goreleaser/goreleaser/v2@<the pinned
   version>`, never installed) passes the dist checks on both sides; each doctored copy of note 5.7.1, 7 fails them, and
   the copy with a scratch `base-v*` tag passes; `goreleaser check` passes at the pinned version; 5.6.5's name test
   passes.
4. **`release.yml`:** only the two tag patterns start it; no `workflow_dispatch`; top-level `permissions: {}`; `build`
   with no environment and exactly `contents: read`, `checks: read` and `actions: read`; `publish` needing `build`,
   inside `environment: release`, with exactly `contents: write`, `id-token: write` and `attestations: write`, running
   no file from the build; a token in a step's `env` only, on the steps note 5.7.2, 2 names; no cache; the steps in
   that note's order, the draft deleted on a failure before publishing; the tag-check query tried against a commit
   whose CI is still running; no secret named anywhere.
5. **The settings,** read back: the `release` environment open to `v*` tags only, Rohan its reviewer; immutable releases
   on; the `base-v*` tag rule on Bonsai's repo and the `v*` tag rule and immutable releases on the workflow repo;
   `release.yml` active; no secret in the repository or the environment; WSL's `gh --version` 2.97 or newer and `which
   -a gh` naming `/usr/bin/gh` first (`/bin/gh`, the same file, may follow); Rohan's word on the old token in the run
   report.
6. **1.0 as published:** `gh release verify v1.0.0`; `gh release verify-asset` and `gh attestation verify` (signer
   workflow, `refs/tags/v1.0.0`, `--deny-self-hosted-runners`) for every archive, and `gh attestation verify` for every
   program; `checksums.txt`
   right; the dist checks green on the downloaded files; the notes equal `CHANGELOG.md`'s 1.0 section; 1.0 is "latest";
   every `1.0.0-rc.N` listed as a pre-release, never "latest"; the tap unchanged since 13 May.
7. **A rebuild:** the verifier's own build from a clean clone at `v1.0.0`, with the release's flags, matches the
   release's linux/amd64 and windows/amd64 programs byte for byte, or every difference is explained from `go version
   -m`.
8. **The installs:** on WSL `/usr/local/bin/bonsai --version` reads `bonsai 1.0.0 (commit <the tag's>)`, root's, mode
   0755; `install.json` matches its path, version and SHA-256; `which -a bonsai` names it alone; `gh attestation verify`
   passes on it; on Windows the same through `Get-Command bonsai -All` and `C:\Program Files\Bonsai\bonsai.exe`;
   `check` in Bonsai's repo gives no `bonsai-path` note and no finding; `status --full --json` reads `state: current`.
9. **The memory line:** the last line of Rohan's `~/.claude/CLAUDE.md` (`tail -n 1`, nothing else of the file read;
   never edited) is the import; the memory check's report shows the index loaded and records any prompt.
10. **The install lines:** Bonsai's table gives seven lines a side, the check before unpacking, for a version given;
    `docs/install.md` equals the table's lines for 1.0.0 (its test, and failing on a temporary copy with one changed
    character); the operating skill names it.
11. **The packs:** `base-v1.0.0` names a commit whose `packs/base/bonsai/pack.yaml` reads `1.0.0`; the workflow repo's
    `v1.0.0` names a commit at `1.0.0` and its release exists, made by its own job; the template's `pack.yml`, base's
    `ci` skill and the workflow repo's `pack.yml` name the 1.0 archives with the SHA-256s in `checksums.txt`; their CI
    green.
12. **The trial on Bonsai:** the trial's feature (`T-5708`) landed on `main` by the pipeline while the installed copy
    was a pre-release (its ladder records name `/usr/local/bin/bonsai` with that pre-release's SHA-256, matched to
    `install.json` and the run report); its task file moved through every status by the orchestrator; its builder a
    named session (its sessions rows carry the task); its guard refusals, stop-gate blocks, asks and Bless, where they
    happened, each in the log; its verifier's report; the tables rebuilt by `check --write`; nothing else `running`
    from the set-up to the landing. Each later round's repeat recorded the same way.
13. **The analysis:** `records/trial-rc1.md` (and each round's page) exists; each claim checked against the file,
    record or output it names; Rohan's reply in his words; every fix he asked for landed or, with his word, left for
    after 1.0.
14. **His word for 1.0:** in the run report in his words, after his reply "happy" and after his own trial on a fresh
    project (his report of it, or his word that it is done); `v1.0.0`'s commit equal to the last pre-release's but for
    the notes' date line, `records/` and `.bonsai/STATE.md` (`git diff --stat`).
15. **The words:** the stale-phrase list finds nothing outside `records/` and `design/`; `bonsai --help` says nothing of
    a rebuild; `SECURITY.md` names 1.0; the issue templates ask for nothing of the old product.
16. **Check 10, the ladder and CI:** `go test ./...` and `go vet ./...`, plain and tagged, in WSL and natively on
    Windows, run by the verifier; `/usr/local/bin/bonsai ladder --task <its task>` green on the final commit, by 1.0; CI
    green.
17. **Stop lines and records:** 5.7's hours under 24, this section's planning and review included (the
    Go piece's in its own report); step
    5's Windows-only tally; option rounds (none planned); nothing written or run in the studio's checkout or in Mimas;
    nothing written outside the repo, the workflow repo and the scratch folders but `~/.bonsai/` by the orchestrator's
    real commands, Claude Code's accepted writes and Rohan's own lines (his installs, his `gh`, his memory line); every
    tag, approval, enable, setting and re-run on GitHub matched to a line of Rohan's in a run report, none to an agent;
    every landing matched to its green `ladder` record and each task's `bonsai.allows` to this section's "Owns"; the
    release files changed only as this section names; the Haiku audit of `.bonsai/sessions.md` against the run reports.
    **Nothing private:** a grep of the diff, the commit messages, the task files, the release notes and the README.

#### Rohan's sittings

His lines are in his part, above; each batch goes in one message. The orchestrator's side:
- **Before batch 1:** V1 passed, every line and its read-back read against GitHub's documentation. **After:** Rohan's
  read-backs read, then an agent reads everything again, changing nothing: `gh api
  repos/LastStep/Bonsai/environments/release`, its `deployment-branch-policies`, both repos' `immutable-releases` and
  tag rulesets, the workflow's state, and in WSL `gh --version`, `which -a gh` and `gh auth status`; then the check
  line's form against `gh`'s own signed file on WSL's `gh` (note 5.7.3, 5). A line that failed gets a corrected line,
  read against GitHub's documentation, in a new message.
- **Before batch 2:** note 5.7.5, 1. **After:** note 5.7.5, 2-3 (V2).
- **Batch 3:** note 5.7.5, 5 before his installs, 6, 8 and 9 after them; note 5.7.7, 1-2; the proposals sent with
  their hours; his pick recorded.
- **Batch 4:** the analysis sent (note 5.7.9) when the feature has landed; his reply recorded; each fix round (note
  5.7.10) with its tag line, its install lines and its short analysis, in one message each.
- **His own trial:** nothing is sent unless he asks; his report, if any, recorded.
- **Batch 5:** note 5.7.5, 4 at his word; then the tag; note 5.7.5, 2 on 1.0's files; notes 5.7.5, 5-9 around his
  installs and his memory line; 5.7.6, then the pack tags.
- His words, what each line printed, and what was read back go in the run report of the piece they belong to (5.7.5's,
  the trial's).

#### Risk in the code, 5.7

- **A release is public and cannot be taken back.** GitHub's lock makes a published release final. So: the rehearsal
  job on every push, V1 before anything is switched on, V2 on rc.1, the trial on Bonsai and Rohan's own trial before
  his word, and the draft published only after signing and the dist checks.
- **A tag runs the workflow its commit holds.** A failure before publishing is fixed on `main`, the tag deleted by
  Rohan's line and made again on the fixed commit, keeping its number; a defect found after publishing needs the next
  number: before 1.0 that is only the next `rc`.
- **A tag with no published release.** While a tag waits for Rohan's approval, after a run failed before publishing,
  or if a stray tag is ever made, 5.6.5's newer-release read (`status --full` reads tags with git) takes it for a
  release, and the install lines it hands over fail to download ("try again in a few minutes" is all 5.6 says). The
  failure path deletes the tag; the run deletes its own draft; and a hand-off back to 5.6 ("Stale or in tension")
  asks 5.6.5 to count only a tag whose release is published.
- **Agents act as Rohan's admin account.** Every tag, approval, enable, setting and re-run is his typed line; a tag
  waits for his approval. An agent could still do any of it through GitHub's API: the rules and the end
  verifier's audit stop accidents, not intent (decision D).
- **The publishing job can write releases and sign records.** It runs only on GitHub's own machines, only after the
  `build` job has passed a tag on a commit on `main` whose checks are green, inside the `release` environment; it runs
  no file from the build (the program and `install.sh` run only in `build`, which can neither write to GitHub nor ask
  for a signing token); tokens go only to the steps that call `gh`; no cache, every action pinned by commit,
  `persist-credentials` off; it holds no secret and can write nowhere but Bonsai's own releases.
- **What the signed record proves.** That the files came from `release.yml` for that tag on GitHub's machines; not that
  the code is good: a commit on `main` made by anyone acting as Rohan is built and signed like any other.
- **The install lines run with root or administrator rights.** The archive's origin is checked before anything from it
  runs; between that check and the installer, the files sit in a folder an agent's shell could write (spec §3's
  tripwire, not a wall): the installers hash their source and their protected copy (5.6.6), and the installed program's
  own signed record can be checked afterwards.
- **`gh` and the PATH.** Rohan's WSL now has GitHub's package source (decision 2, done 10 Oct); the orchestrator's own
  `gh` commands moved from 2.4 to a current version (their JSON and flags read again where a run report's command uses
  them). The install lines find `gh`, like `sh`, `tar` and `curl`, on the PATH, and folders Rohan and agents can write
  may come before `/usr/bin` there (spec §3's note on `~/go/bin` and `~/.local/bin`): a `gh` put there first would
  answer the check line. Batch 1 reads `which -a gh` back; the walls refuse agents' file tools writing those folders,
  not a shell (spec §3's tripwire).
- **Go 1.27.** The move from Go 1.25 (Rohan's word, its own piece before 5.7) is checked by that piece's check 10 on
  both sides, counts before and after; 5.7.0 checks it still holds at release time.
- **Bonsai's own guard on a pre-release.** From his rc.1 install, every Claude Code session in Bonsai's repo, the
  orchestrator's included, runs rc.1's guard, stop gate and recorder; if it blocks work, 5.4's way back with the copy
  kept for it (note 5.7.5, 5). A change to Bonsai's own hook lines shows as `check`'s finding and is his (ii) line.
- **The trial is real work on the real repo.** Its feature lands on `main` and ships in the next pre-release and in
  1.0; a feature that turns out wrong is reverted by a task like any other. The trial runs alone, so a fault it finds
  stops Bonsai's other work until it is fixed or worked round by 5.4's way back.
- **What the trial finds may be large.** A finding that needs a format change by removal or a new major comes to Rohan
  first (his (B)); one that needs a guard, hook or release-path change gets its verifier (V1's brief for the release
  path) and its own pre-release; one that is not a defect of 1.0 may wait for after 1.0, on his word, written in the
  analysis.
- **How many pre-releases.** rc.2 is in the figures; two more rounds fit under 24 at the high figures, and the round
  that would cross it waits for Rohan's choice (continue, a smaller cut, or pause). Each round costs him about 10
  minutes.
- **His own trial, on his own.** Agents do not watch it; what he finds reaches the plan only through his words.
- **The UAC prompt's first real run** is at his rc.1 install: its way out is written, nothing on Windows depends on
  Bonsai yet, and a problem found there costs a new `rc`, not a 1.0.1.
- **Windows tests after the Windows install.** A test that assumes no installed `bonsai.exe` may change; check 10's
  Windows half is run again and compared.
- **Dependabot and pinned actions.** Its pull requests wait for no agent: the orchestrator lands the same move, and
  Dependabot closes its own.
- **Public words.** The README, the notes and the changelog go out with 1.0: the privacy grep and the orchestrator's
  read before each lands.
- **Shared files.** 5.7.3 and 5.7.4 share none; every piece with a protected file runs alone.

#### Stale or in tension in the spec, for 5.7

The orchestrator writes these dated notes on `main` once Rohan has approved the section (this plan does not edit the
spec); they carry his answers of 10 Oct.
- **§3, "Go version: the module says `go 1.25` with a `toolchain` line"; the gate's "a later 1.25.x patch fixes
  them":** Go 1.25's last patch was 1.25.14 (19 Aug 2026). Note: "> **Changed <date> (step 5.7's section):** the `go`
  line stays `go 1.25.0`; on Rohan's word (10 Oct, "we can go with go 1.27 latest") the `toolchain` line moves to Go
  1.27's newest patch, because Go 1.25 has had no security fix since Go 1.27 came out (19 Aug 2026). govulncheck's pin
  moves with it. The next such bump comes when Go 1.29 is out."
- **§12 step 7, Homebrew; §14's 5.7 row, "`bonsai@0.4`":** note in §12: "> **Changed 10 Oct (Rohan, step 5.7's
  section, 1 (A)):** no Homebrew for the new Bonsai. A Homebrew install lands outside the two places every
  project's hook lines name (§3's 9 Oct note, (a)), so it could never guard. The tap keeps `bonsai` at 0.4.3, unchanged,
  so 0.4.3 stays downloadable (Q8); there is no `bonsai@0.4` and no tap token, and no release writes outside Bonsai's
  repo. The new Bonsai installs with its installers. Later, outside step 5: a Mac, and each platform's own package
  manager, each installing into the two places the guard trusts." §14's row gets a one-line pointer to it.
- **§12 step 8, "immutable releases, build provenance":** "> **Changed <date> (step 5.7's section):** immutable releases
  are a repository setting, switched on by Rohan's line; GoReleaser builds in a job that cannot publish or sign; a
  second job, in the `release` environment and running nothing from the build, signs every archive, `checksums.txt` and
  every program with GitHub's artifact attestations, makes a draft, checks it, publishes it, and checks the release as
  published. The re-release input went in part 1 and the manual run now: `release.yml` runs on `vX.Y.Z` and
  `vX.Y.Z-rc.N` tags only, for a commit on `main` whose checks are green. The install lines check the record before
  unpacking (`gh attestation verify`)."
- **§17 step 4:** "> **Changed <date> (step 5.7's section, Rohan's answers of 10 Oct):** no new tap token: the `release`
  environment holds no secret. The environment asks Rohan's approval before each release run, in place of the optional
  tag ruleset. Immutable releases are switched on with `gh api -X PUT repos/LastStep/Bonsai/immutable-releases`. Tag
  rules stop `base-v*` tags in Bonsai's repo and `v*` tags in the workflow repo being moved or deleted, and the workflow
  repo's releases are locked too. The old token's revocation [is confirmed / was not found], <date>." Also the line "At
  step 5.7, on your word, switch it back on" stands, with the date it was done.
- **§14 row 10, "the first public release, on Rohan's word", and §14's 5.7 row (6-11 h):** "> **Changed 10 Oct
  (Rohan):** before 1.0, a trial of Bonsai on itself on a public pre-release, `v1.0.0-rc.1`, which he installs: a small
  real feature he picks, built through the whole pipeline on Bonsai's own repo, its results analysed for him, fixes as
  further pre-releases; then his own trial on a fresh project; then his word for 1.0 on the same code. 5.7 becomes
  9.5-18.5 hours, re-ask at 24 (the trial adds 3.5-7.5); step 5 142.5-225.5." (`design/plan-5.md`, "Step 5.7".)
- **§5, the pack's CI step 2, and §5 and §6's examples `ref: base-v1.0.0`, `ref: v1.0.0`:** done at 5.7.6, no note
  needed beyond the plan's; `init`'s template still names no pack (note 5.7.6, 1).
- **§3, "No agent installs or replaces it":** unchanged; from 1.0 the installers and Bonsai's printed lines replace
  §17 step 8's typed lines (5.6's note), with the check of §12 step 8 among them.
- **This plan's own text:** 5.6's "six lines a side" reads seven from 5.7.3 (Rohan's list says so); the outline's
  "Go 1.25.x toolchain bump" is Go 1.27's, done before 5.7 on Rohan's word (above).
- **A hand-off back to 5.6** (its section is approved and not yet built, so it is not edited here; the orchestrator puts
  it in 5.6.5's brief, or brings it to Rohan if it changes what is his): 5.6.5 counts a tag as a release only when that
  release is published, for example when its `checksums.txt` answers at the fixed download address (one HTTPS request,
  no login, inside the same 10 s), so a tag waiting for approval, or one whose run failed, never yields install lines
  that fail. Its cost: a second kind of network read beside git's, which 5.6.5's builder weighs; until it is built, the
  risk stands as "Risk in the code, 5.7" writes it.

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
- **Planned in full** in "Step 5.5" above; this outline is kept as it was written.
- **Builds:** `base` and `workflow` from the studio's roles, protocols and templates; the always-on and skill split,
  the roles' `skills:` preloads; `claude plugin validate --json` in each pack's CI, failing on every warning but the
  missing `version`; the walls in base's deny rules and the studio's in `workflow`, each tried once on both sides
  (13-21); the documentation in every template and pack file, and each deny rule's `why` (3-4); the pack template
  `packs/template/` with its CI and release, and Bonsai's CI job that runs it (3-5).
- **Agents first** (Rohan, 9 Oct, 15:35): an **"operating Bonsai" skill in `base`**, so any agent in a linked project
  can link, update, fix, check, read status and edit `bonsai.yaml` unaided: the commands with `--json`, what to do on
  each error word and finding (from `bonsai --help --json`), the person's gates and why (consent to code under his
  (ii), person-only files, the program install), and the ladder's climb-read-fix loop (5.4). Documented as every pack
  file is, and tried in a real session.
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
- **Planned in full** in "Step 5.6" above; this outline is kept as it was written.
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
- **Planned in full** in "Step 5.7" above (9.5-18.5 h with the trial Rohan asked for on 10 Oct, re-ask at 24); this
  outline is kept as it was written.
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
  `gh workflow enable release.yml`; the optional tag ruleset; his word for 1.0; then 1.0 installed on both sides with
  5.6's installers (in WSL about 5 minutes and his password; on Windows about 5 minutes and one UAC prompt, its first
  real run, with its way out), and with them his memory import line in WSL's `~/.claude/CLAUDE.md`, an agent then
  checking that it loads (5.6's "Hand-offs to 5.7").

### Stop lines and hours

Judged by each part's end verifier, not the builder. Done means **no line crossed without Rohan's recorded choice, or,
past a line, his recorded choice to go on.**
1. **A part's hours over its re-ask line:** 5.1 61, 5.2 48, 5.3 38, 5.4 57, 5.5 39, 5.6 26, 5.7 24 (spec §14: 1.3 times
   each part's high estimate, rounded: 47 x 1.3 = 61.1; 37 x 1.3 = 48.1; 29 x 1.3 = 37.7; 44 x 1.3 = 57.2; 30 x 1.3 =
   39; 20 x 1.3 = 26; 5.7's 18.5, the spec's 11 and the trial Rohan asked for on 10 Oct, x 1.3 = 24.05). A part that
   ends under its line passes nothing on to the next.
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
| 5.5 | Its section's "Proof for each piece"; fresh verifiers P0 (privacy, base's three templates from the studio before they land), P (privacy, `workflow`'s whole history before its first push) and V1 (the walls, the `<protocols>/` resolution and the packs' CI); the end verifier on "5.5 done", after Bonsai takes `base` |
| 5.6 | Its section's "Proof for each piece"; fresh verifier V1 (the installers and the machine's tripwires, after 5.6.6 with its CI job green); the end verifier on "5.6 done" |
| 5.7 | Its section's "Proof for each piece"; fresh verifiers V1 (the release path, before anything is switched on) and V2 (rc.1's published files, before Rohan installs it); the trial on Bonsai and its analysis, reviewed by Rohan; the end verifier on "5.7 done", after his word, the 1.0 installs and the pack tags |
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
  `bonsai.yaml` comes from a built-in template and `init` writes no STATE. Settled in 5.5's section (note 5.5.0, 3):
  both templates live in the engine, and base's `workspace` and `state` skills are generated from them.
- **§14's 5.1 row says "all of §6's findings and warnings"**, while its 5.6 row names "the stranded-folder report" and
  "the personal memory layer and its check". The stranded folder goes to 5.6; the secret scan of memory notes to 5.2
  (the redactor's patterns are its one home); the rest of §6 is 5.1's. Settled in 5.6's section: the stranded folder in
  note 5.6.1, 5; the personal layer's check in note 5.6.3, 3, reusing 5.2's redactor.
- **`status_writes` and `status_command`:** the status test's comment puts them in 5.1 "(bonsai settings)", but
  `settings` is 5.6's word. 5.1 reads the machine settings; 5.6 writes them (note 5.6.1, 3).
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
  is the studio's plan. Both commands refuse in an agent's session (5.6's section), so the registration is run by a
  person or by the studio's own program outside one.
- **Step 7, "the studio links (after 5.5)"**, includes the machine installs of `bonsai`, but a 5.x build is not a
  release (§3, §18) and the one pre-release is WSL's, at 5.4. That pre-release refuses the `workflow` pack (its reader
  refuses the `<protocols>/` path, which 5.5.0 adds), so before 1.0 the studio's link of `workflow` needs a pre-release
  built after 5.5.0, a new install of Rohan's in WSL (5.4's four lines), or waits for 1.0 at 5.7: the studio's plan
  decides; nothing here waits on it.
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
