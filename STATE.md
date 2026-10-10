# Bonsai: where it stands

Rewritten, never appended. Last rewritten 10 Oct 2026 at 19:55. **5.1 is done** (14:30). **5.2 is under way and paused
on Rohan's word** (19:36: "pause once these are done, dont queue others until i say go"): 5.2.0, 5.2.2, 5.2.6a and
5.2.3 have landed; 5.2.1, the redactor, was failed by its fresh verifier (three defects) and waits for its fix round;
nothing is running. **The next session (or Rohan's go) starts at "Now: step 5.2".**

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The plan is approved (8 Oct). Part 0 and all six parts of the walking skeleton
are done (gate report `records/gate-skeleton.md`). At the gate (9 Oct) Rohan chose path (a), the full Bonsai 1.0, and
approved the plan for step 5, `design/plan-5.md`; every part's section, 5.1 to 5.7, is written and approved.

- **Part 0, the formats** (`3770d04`): a JSON Schema for each format, examples, trick files, a raw-byte manifest and
  its Go test. Now set 5 (`d11f704`); 5.2.0 makes set 6.
- **The skeleton, parts 1-6** (`17f2938` to `5633721`): the clear-out, the format-1 reader, the engine (`init`,
  `update`, `check`), the hook path (`hook guard`), packs as Claude Code plugins, the gate report. The test pack
  `LastStep/bonsai-test-pack` is public, `main` at G (`cf356d3`).
- **Step 5.1, formats and engine to 1.0: done** (code `11d871c`; verified on `5fa9bd8`). 5.1.0-5.1.10: consent to code
  (`--allow-exec`), the format-0 reader, formats sets 4 and 5, a Go type for every format and `check --schema`, every
  command's `--json` and error words, the lock's `declares`, `check` and `status` to 1.0 (33 check words, `--full`,
  `--active`), `unlink` and taking a pack out, `check --write`'s tasks table, `check --pack`, the generated reference
  page `docs/reference/lists.md` and `bonsai --help --json`; the toolchain at go1.27.2. Two fresh verifiers (12:02-12:27
  stopped at the last session's close, 14:06-14:30) passed all 21 checks of "5.1 done", no defect. Tests on `5fa9bd8`:
  WSL 2065 runs plain and 2071 tagged (2 skipped, Windows/macOS-only), Windows natively 2069 and 2075 (1 skipped,
  symbolic links); vet clean; CI green on both sides.
- **Step 5.2, the recorder, logs and asks: under way, paused** (25-37 h, re-ask at 48). Landed, CI green on each:
  5.2.0 formats set 6 (`cbc34c3`), 5.2.2 appends, the reader and the salt (`0573c0b`), 5.2.6a the generated-files page
  (`ca3b0ff`), 5.2.3 the sessions table and `bonsai logs` (`8d27325`). 5.2.1, the redactor (branch `5.2.1`,
  `e9d9d93`): its fresh verifier FAILED it, three defects to fix.

Hours: part 0 45 minutes; the skeleton 6 h 18 min of 30-47 h. **5.1: 866 minutes (14 h 26 min) of its 61 h re-ask
line**, under its 30-47 h estimate. 5.2 so far: 433 minutes (7 h 13 min) of 48 h: 55 of planning on 9 Oct, 11 of the start check, then
5.2.0 23, 5.2.2 47, 5.2.6a 34, 5.2.1 140, its verifier 54, 5.2.3 69.
5.5's planning 84 minutes, 5.6's 67, 5.7's 117 (84 of planning and review, 33 of the Go switch). Step 5's Windows-only
tally: about 3 minutes (of 8 h; 5.2.2's over-strict salt test). Option rounds: one in 5.1, none in 5.2. No change to Mimas or the studio's repo. No stop line
crossed.

## Rohan's decisions

- **10 Oct, about 10:00: 5.6's section approved, with (A).** Asked "for each decision in plan 5.6, give me proper
  context and then ask me with good options", he took six decisions one at a time and chose the recommended one each
  time: (1) notes about him in his personal memory are his to save, from an agent's draft (not agents writing them);
  (2) this machine's settings and the studio's label files are all his, even how long pack copies are kept; (3) a
  newer release is known as planned (`status --full` at most daily, kept in the home's `cache/release.json`, shown
  offline); (4) the installers first run on his machines at 5.7 with 1.0; (5) `status --line` kept; (6) the personal
  index at 40 lines and 4 KB, a note 4 KB. Then "Approve": 13-20 h, re-ask 26. Spec §4's dated note on `status`'s home
  cache written. From now on a section comes to him as one question per decision, then the approval.
- **9 Oct, 18:32: 5.5's section approved, without the pin rule.** Asked whether he approves it with the new public
  `workflow` repo and a rule letting agents move a project to a newer pack version when nothing loosens, he chose
  "Approve, no pin rule": moving a pack's version, adding a pack to a linked project and taking one out stay a
  person's step; an agent prepares the change and hands him the exact line, and makes it only on his word (5.5's
  "Moving a pack's version: a person's step"). Spec §5's "Adopting a release stays a person's step" and 5.4's rule
  stand.
- **9 Oct, the gate: path (a), the full Bonsai 1.0** (spec §14, "Path (a) after the gate", 139-218 h, parts 5.1-5.7,
  each with a re-ask line at 1.3 times its high estimate).
- **9 Oct: Bonsai has no screens of its own.** "Completely remove the idea of bonsai's own screens ... this visual part
  of the job will be handled by the studio, while bonsai is a pure cli tool for overall simplicity." No web page, no
  `bonsai serve`; every visual of Bonsai's data is the studio's, from Bonsai's JSON outputs. Dated notes in the spec
  (§1, §2, §4, §11, §14, §16), the contract's reader lists, the one-pager and `CLAUDE.md`.
- **9 Oct: Bonsai's scope stands as split on 7 Oct.** Bonsai is everything inside one project (its rules, its guard,
  the record of what happened, the proof that work is done) and works with no studio; the studio is everything across
  projects and anything that acts on agents (dispatch, approvals, status moves, notifications, every visual). The
  recorder and the ladder runner stay in Bonsai; Bonsai has no scheduler.
- **9 Oct, 15:46: no self-update.** He chose a self-updating program at 15:42, then: "actually scratch out the self
  update thing. i dont want that". The program stays his install per release on each side; from 5.6 `status` and
  `check` say when a newer release exists and an agent hands him the lines. Spec §3's dated note says so, and why the
  guard cannot be served online.
- **9 Oct, 15:35: Bonsai is managed by AI agents inside projects** ("bonsai will mostly (and probably completely) be
  managed by the ai agents within a project. so make sure the support for that is top notch"). Asked whether that
  includes the program on each machine, he chose "inside projects": agents link, update, fix, check, read status and
  edit; the `bonsai` program stays his install per release (root/admin), which keeps his 5.3 (a) safe. Three additions
  follow: `bonsai --help --json`, a machine-readable list of every command, flag, exit code and error word (5.1.10);
  every `check` finding's `next.do` the exact command that fixes it where one exists (5.1.6); an "operating Bonsai"
  skill in `base` (5.5). (A self-updating program was chosen at 15:42 and dropped at 15:46, above.)
- **9 Oct, 5.3's two questions:** the hook lines name the installed `bonsai` (Windows' place, then WSL's
  `/usr/local/bin/bonsai`), never the PATH, with check 2's one named exception ((a)); and an agent may pass
  `--allow-exec`, run `unlink` or change Bonsai's guard lines only under his tap's grant in a project the studio
  manages, only a person elsewhere ((ii)), so from 5.4 such an update of Bonsai's own repo is his command. Spec §3 and
  §18's dated notes; `design/plan-5.md`, "Step 5.3".
- **9 Oct: a pack plugin that runs code is asked for on each machine** (5.1's first option round): `update` installs a
  locked pack plugin with code parts only with `--allow-exec` on that machine; and Bonsai installs or removes no
  plugin but the project's own packs' ("it shouldnt install or remove other plugins"). Spec §5's dated note.
- **9 Oct: later parts' plans reach him under (B):** a part's section comes to him only when it changes what is his
  (hours or re-ask line, the order, his steps, a new public repo or content, an option round, a format change: a new
  major or a removal). He first said (A), then chose (B) once the scope question was settled.
- 9 Oct: a handoff for the studio's orchestrator, `records/handoff-studio-2026-10-09.md`; the studio still links at
  step 7, after 5.5 (his 8 Oct "Bonsai first" stands).

8 Oct:

- Bonsai's design and records live in this repo (`design/`, `STATE.md`, `records/`); Bonsai owns its spec and the
  formats contract. Work lands on `main` directly; the old product stays at tag `v0.4.3`.
- The JSON Schemas and the trick files are Bonsai's: the master lives in `formats/`.
- **Bonsai first**: the studio adopts the formats when it links (spec step 7), not before.
- "The proper way, no shortcuts": the spec's order inside Bonsai; parallel work only where truly independent.
- Bonsai joins the studio's dashboard as its own project at spec step 6.
- Haiku for small bookkeeping and audit jobs; it reports facts and judges nothing (`CLAUDE.md`).
- The plan and one-pager approved after a fresh Opus review and its fixes.
- Label-definition fields stay required in the formats; the contract's short examples are out of date.
- Hand checks in two sittings; and (his later word) a Sonnet agent runs hand checks where an agent can.
- His roadmap is the artifact "Trinetra Roadmap" (https://claude.ai/artifact/XXKTi6geneu1pdQ4h4miw2); Bonsai's cards
  are updated after each part.

## GitHub, read 9 Oct

- `main` is the only branch; the newest tag is `v0.4.3`; no open pull request.
- `release.yml` is disabled (`disabled_manually`) and can only build; no repository secret. The `release` environment
  and a new tap token wait for step 5.7.
- Ruleset `main-protection`: blocks force pushes and deletion only.
- `LastStep/bonsai-test-pack`: public, `main` at G (`cf356d3`), no tag.
- The old website stays on GitHub Pages until Rohan turns Pages off.

## Now: step 5.2 (paused; resume on Rohan's go)

`design/plan-5.md`, "Step 5.2". Run reports `records/runs/R-2026-10-10-5.2.*`; 5.2.0's holds the part's running total
and the start check (14:06-14:17). **Every next brief is written and waits in `~/bonsai-checks/briefs/`** (the
session's own scratch folder was cleared once during a pause, so briefs live there now): `rules.md` (the rules every
builder's brief carries), `brief-5.2.1-fix.md`, `brief-5.2.4.md`, `brief-5.2.5.md`; assemble each as its brief with
`<BASE>` set to `main`'s commit and `<BESIDE>` naming what runs beside it, then `rules.md` with `<piece>` replaced.
- **Landed** (each on green check 10 on both sides, CI and the orchestrator's read; the 5.2 end verifier covers them):
  **5.2.0** (Sonnet, 23 min): formats **set 6**. **5.2.2** (Opus, 47 min): `internal/record`, `workspace.FindLocal`, the
  salt, `bonsai --version` with its commit. **5.2.6a** (Sonnet, 34 min): `docs/reference/generated-files.md`.
  **5.2.3** (Sonnet, 69 min): `internal/sessions` (`sessions.Open`, the one "open" function, for 5.2.6b), the
  sessions rows in `check --write` (a 5.1-written 8-column table rewritten, rows kept), the `tables` warning over both
  tables, `bonsai logs`. Differences from its note, each following set 6's schema: no per-task hours totals (a reader
  sums the rows), a subagent run's model null, `logs --json` oldest first.
- **5.2.1, the redactor: FAILED by its fresh verifier** (Opus, 18:43-19:37; `R-2026-10-10-5.2.1-redact.md`, 19:38).
  No leak in about 1.1 M strings otherwise, and the builder's three open points judged acceptable, but: **D1** some
  inputs take quadratic time (repeated `ſ-` 28 s at 256 KB against 100 ms; also the Kelvin sign, `eyJ-`, `AIza-`):
  measured again by the orchestrator at load 1.3, doubling the input quadruples the time, so not the machine's load;
  **D2** a token shape glued straight to a secret-named key swallows the name and leaves its value (an AWS key id
  before `PASSWORD:` Bonsai's only; webhook, Slack and `sk-` shapes shared with the studio's; note 1's rule covers
  all); **D3** test rows copied from the studio's tests or lightly swapped. **Next on Rohan's go: the fix round**
  (an Opus builder on branch `5.2.1`, `brief-5.2.1-fix.md`, about 2-4 h), then a fresh verifier again; on PASS 5.2.1
  lands.
- **Then:** 5.2.4 (the recorder; Opus; a fresh verifier) and 5.2.5 (asks and `log append`; Opus; may run beside 5.2.4,
  the words each in their own file) after 5.2.1 lands; 5.2.6b (cleaning; Opus) last, calling `sessions.Open`; then
  the 5.2 end verifier, the roadmap and `STATE.md`.
- **For the next briefs** (already in the drafts): 5.2.4: `secret` out of `checkLater`; `hook start`/`record`'s
  `Later`; `internal/redact` as `record.Common`'s redactor; `input_hash` null when the salt errors; the guard's first
  call over 5 ms (p50 7.2-7.5 ms); whether `redact.Kinds` goes on the reference page. 5.2.5: `format.AskTypes`,
  `internal/asks` in `codeWords`, orders ask 8, answer 9, asks 10, log 12 (logs is 11). 5.2.6b: `sessions.Open`,
  `Found` and `Missing`; a changed rule edits `generatedfiles.go`'s prose too. Every piece changing a listed table
  runs `go generate ./...`. **The build line** for scripted runs: `go build -buildvcs=true -o
  ~/bonsai-checks/<piece>/bonsai ./cmd/bonsai`, its `--version` showing `(commit <12 hex>)` with no `+modified`.
  Builders who rebase re-run the full check 10.
- Small, unplaced: `internal/workspace/workspace.go`'s package doc does not name `local.go` and `salt.go`;
  `formats/examples/workspace.yaml` keeps the old "skill base:generated-files" comment (the next formats set).
- The machine was heavily loaded this evening (load up to 6, another session's video render): Windows test runs took
  about 11 minutes each; timings measured then say the load.

**Carried into later briefs:** 5.6.5's brief counts a tag as a release only once its release is published; 5.7's runs
so far (84 minutes of planning and review, 33 of the Go switch) go into 5.7.0's run report; the later sections' formats
set numbers move by one (5.4.0's set 7, 5.6.0's 9); 5.7.0 checks the toolchain still holds.

**Placed for later by 5.1 and its verifiers:**
- 5.2.3: the sessions rows in `check --write` and its help's "joins in step 5.2.3"; a missing `sessions.md`.
- 5.2.4: `secret` out of `checkLater` into the check table (the memory secret scan); the guard's first call over spec
  §3's 5 ms (p50 7.2-7.5 ms on the final build, about 1 ms of it from go1.27.2).
- 5.3: the guard judging `unlink` as a person's under (ii); guarding `lock.json` (a forged lock whose files match its
  hashes is seen by nothing today); the toolchain's millisecond on the first call in its timing bar.
- 5.5: `needs.mcp` and a lock `ref` for the moved-tag gap (moved here from 5.2.0 at its start check); the operating
  skill covering `unlink` and taking a pack out; 5.1.9's limits (a template with no fields table is not seen;
  `plugin.json`'s `name` against the pack id checked nowhere; plugin hooks given as a path not walked by `pack-bash`);
  the "the pack  (...)" wording at `init --source` (`internal/engine/fetch.go`); **the marketplace observation**: in
  a scripted session Claude Code itself installed a pack plugin under its new marketplace name and ran its hook before
  Bonsai's `update --allow-exec`, so Bonsai's flag gates only Bonsai's own install; whether a person's trusted session
  does the same needs a person.
- 5.6: the machine folder `unlink` leaves (`stranded`); `install.json`'s keys; spec §10's Windows personal-memory
  warning (5.6.3).

How a session runs, for the next orchestrator: pieces and later parts' plans side by side; every brief carries its
rules in full (this session wrote them to a scratch file each builder reads first); each builder rebases on `main`
before it lands; times in the run reports from `date`; no test or script reaches the real `claude` except through
`claude-here` in a scratch target, with the settings hashes before and after; Windows' `claude.exe` kept off the PATH
of every other Windows run. Mimas's git checkout is the `MimasGame` folder (the `Mimas` folder beside it is the old
product's Plastic workspace).

## Waiting on Rohan

- **His go to resume 5.2:** the 5.2.1 fix round starts ("Now: step 5.2").
- **Claude Code's login on Windows has lapsed again** (the 5.1 verifier's Windows sessions ended "OAuth session
  expired"). 5.2.4 runs a real Windows session (or reads the log of his 5.3 session), so a login before then helps.
- **For his note:** in the verifier's first Windows session (14:21), Claude Code's failed login refresh rewrote
  `%USERPROFILE%\.claude\.credentials.json`, which is not on the accepted list of writes; its contents were not read.
  The user settings files are unchanged on both sides.
- Later, as the plan has them: at 5.4, a pre-release `bonsai` in WSL (about 5 minutes); at 5.5, reading the `workflow`
  repo and making it public; at 5.7, the release settings, the `rc.1` tag and his approval, the installs, the trial's
  analysis, his own trial on a fresh project, his word for 1.0 (about 1 h 45 min in all).
- Whenever he likes: turn GitHub Pages off (the old website).

## Loose ends

- Rohan's roadmap artifact: Bonsai's cards updated at 5.1's end (5.1 done, 5.2 under way); the studio's cards
  untouched. Next update at 5.2's end (plan item 9).
- Another of Rohan's sessions also works in `~/Servers/Bonsai` (10 Oct: two commits taking Jev out of Bonsai's plans,
  pushed). Read `origin/main`'s new commits before each push, and never stage a file this session did not write.

- The WSL user settings file's baseline is `c3978abe...e73ee7` since Rohan's `/plugin` and `/reload-plugins` at the
  start of the 10 Oct session (told him; before it `9e049dea...80d8d6`, from his `/plugin` of 9 Oct). Windows' is
  unchanged, `2b6295c1...4ff6c9`.
- From 5.1.1: S3's rest (a hook running a pack file by a glob or a built name escapes the `runs` scan) is the limit
  of the plan's rule 4, kept; `update --yes` says "nothing to change" on a forged lock while `check` finds it (5.1.6);
  the test pack's header named A-D until G (`cf356d3`, 5.1.9), which fixed it.

- Every finding and loose end the skeleton leaves is in `records/gate-skeleton.md` section 5, grouped by the step 5
  part that must settle it (5.1 to 5.7, and the spec's own text). This file no longer repeats them.
- Outside step 5's parts: the spec's §17 step 6 still names `/agents` and one sitting.
- Notes from the last verifier, for the record: part 0's run report quotes a studio task id from the plan; the reader's
  comments name `yaml.mjs`, the contract's format-0 reference (the plan's exception lists only part 0's report and
  `formats/README.md`).
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: clones, bundles, builds, targets, the
  launchers (`claude-here`, `bonsai-here`), the eval suite (`eval/`) and the last verifier's runs (`verify-final/`).
  `verify-final/scripts/out/check11-all.log` holds a full PATH with the user name: scratch only, never to be pasted
  into the repo. Safe to delete once step 5 has its own checks, but for the launchers if it reuses them.
