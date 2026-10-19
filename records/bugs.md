---
title: Known bugs   # the bugs register; every field and how to fill it: skill workflow:bugs
updated: 2026-10-19
---

# Known bugs

The defects someone has seen or proven and nobody has fixed yet. How to add or fix one: skill workflow:bugs.

Started 19 Oct 2026 from the audit `M-2026-10-18-39aecfea` and the register draft `M-2026-10-18-47cb853c` (auditor,
Sonnet), read at `main` 91d8a972; it replaces the bug lists `.bonsai/STATE.md` held. In "Where": **pointed** means the
line was read there; **said** means only a report's or STATE's words were read, not the code. A limit the director
accepted for after 1.0 stays Open with that word. Design questions are not here: they wait in STATE's "Waiting on the
director".

## Open

| Id | What goes wrong | Where | Found by |
|---|---|---|---|
| **BUG-006** | The 10 run reports for T-5601, T-5602, T-5603, T-5604, T-5605, T-5606, T-5691, T-5692, T-5693, T-5694 have an empty Runs table, so 5.6's hours are not counted | records/runs/R-2026-10-16-T-5601..., R-2026-10-18-T-5692... (ran: awk over the files; T-5600 and T-5607 have rows) | STATE "Small follow-ups for the end" |
| **BUG-007** | CI pins actions by tag (checkout@v6, setup-go@v6, setup-node@v6, golangci-lint-action@v9, goreleaser-action@v7, codeql-action/*@v4), so a moved tag changes CI | .github/workflows/ci.yml:18,23,52,57,100,103,108,135,138,146,155,158,183,186; codeql.yml:23-40; release.yml:17-30; 91d8a972 (pointed) | STATE "Loose ends, from tonight"; 5.7.0 |
| **BUG-008** | The packs' release job could run only on push (it runs on more events) | the workflow pack's release workflow in LastStep/bonsai-workflow; 91d8a972 (said; the repo is outside my folders) | STATE "Loose ends, from tonight" |
| **BUG-009** | check.sh validate prints "ok" in words that mislead | ~/bonsai-checks/bonsai-workflow/check.sh (workflow pack ci); 91d8a972 (said; not re-read) | STATE "Loose ends, from tonight" |
| **BUG-011** | The guard may read a `~/` redirect target as relative to the working folder (refuses more, not less). Not confirmed: shell_paths.go:107,495 (delAbs) read ~/ against the home, so it may hold for the redirect form only | internal/guard shell_paths.go delAbs; shell_words.go:1132; 91d8a972 (said; not confirmed) | ~/bonsai-checks/verify-5.5-end/report-r2.md:53-60 |
| **BUG-013** | The verifier and runner roles cannot make scratch repositories, so those steps go to a builder | the workflow pack's role files (agents/verifier.md, runner.md); 91d8a972 (said) | 5.5 end verifier (report-r2.md); STATE "How work runs now" |
| **BUG-014** | Bonsai's old tags v0.1.1-v0.4.3 read as newer versions of base in bonsai status (no pack-scoped tag form yet) | internal/engine tag reading (newer.go/release.go), cmd/bonsai/status.go; 91d8a972 (said) | STATE "Waiting on the director, Later" |
| **BUG-015** | A tag whose release is not published yet (waiting for approval, or its run failed) counts as a release, and its install lines fail to download. The orchestrator decided against a published-only check (Go's net would break TestTheBinaryLinksNoResolver) | internal/engine release.go, limits comment at line 25 (pointed) | M-2026-10-18-c3af4758; M-2026-10-18-412e8ad9; T-5605 |
| **BUG-016** | status_command's rule (letters, spaces, - and /) allows a plain harmful command such as `rm -rf /`; the plan bars only metacharacters | internal/workspace/homesettings.go, StatusCommandOK (lines 108-120, pointed) | M-2026-10-18-8ba15229 (N5); T-5693 |
| **BUG-017** | bonsai.changes/1 has no field for update's Memory lines or its cache-clean line, so --json carries nothing for them | formats/schemas/changes.schema.json (cache mentioned only for the adopt copy path, line 249; pointed) | M-2026-10-18-9b6f4d99 ("Left open"); T-5603 |
| **BUG-018** | Installers write the binary by rename and the install line tests `[ -s ]` (placed for 5.6) | install/install.sh; 91d8a972 (said) | STATE "Placed for later, 5.6" |
| **BUG-019** | The words Claude Code shows for "no bonsai is installed" are not Bonsai's | said (STATE only) | STATE "Placed for later, 5.6" |
| **BUG-020** | bonsai unlink leaves the machine folder | cmd/bonsai unlink; 91d8a972 (said) | STATE "Placed for later, 5.6" |
| **BUG-021** | install.json's keys are not settled (placed for 5.6) | install/ and internal/engine check of install.json; 91d8a972 (said) | STATE "Placed for later, 5.6" |
| **BUG-022** | The Windows personal-memory warning is not built (placed for 5.6) | said (STATE only) | STATE "Placed for later, 5.6" |
| **BUG-023** | No per-id index of the machine records | said (STATE only) | STATE "Placed for later, 5.6" |
| **BUG-025** | A pack's relative deny rule does not hold after a shell `cd` | Claude Code deny rules as written by the pack; 91d8a972 (said) | STATE "Placed for later, 5.5" |
| **BUG-026** | The operating skill's line on an added rung says a person's step (since K1); the plan's text near line ~4031 does not | packs/base/skills/operating-bonsai/SKILL.md vs design/plan-5.md ~4031; 91d8a972 (said; skill since split, not re-read) | STATE "Placed for later, 5.5" |
| **BUG-028** | A shared active-task fixture for a task file holding only `garbage: [` is missing | formats/ fixtures; 91d8a972 (said) | STATE "Placed for later, the next formats set" |
| **BUG-029** | ladder.schema.json's `leftovers` description is wrong or stale | formats/schemas/ladder.schema.json; 91d8a972 (said) | STATE "Placed for later, the next formats set" |
| **BUG-033** | The plan's text on 5.5.11's limits is looser than the code | design/plan-5.md (5.5.11) vs internal/guard/roles.go, budget.go; 91d8a972 (said) | STATE "Whenever they like" |
| **BUG-034** | 2,167 subagent ids in the log carry only a stop record | the log under .bonsai/local/log (recorder); 91d8a972 (said; STATE: "the recorder's look before 1.0") | STATE "Loose ends, from tonight" |
| **BUG-035** | ~/.pi/agent/models.json and Gemini's other files are not walled | internal/guard walls (said); 91d8a972 | STATE "Loose ends, from tonight, after 1.0 unless the director says otherwise" |
| **BUG-036** | The guard does not read `script -q /dev/null < cmds.txt` (a runner it does not unwrap) | internal/guard shell.go runner tables; 91d8a972 (said) | STATE "Loose ends, from tonight"; after 1.0 unless the director says otherwise |
| **BUG-037** | Guard, runners: an unread runner (tmux, screen, expect -c, systemd-run, nsenter, ssh, strace, ltrace, at, firejail, bwrap, parallel) and a git alias already in git's config are not read; env -S splitting rules; env -v echo piped into a shell via a runner word; the git icase variable named with quotes; a planted .git folder git does not take. After 1.0 (the director, 11 Oct) | internal/guard shell.go comment lines 47-62 (pointed); rolecmds.go:37-48 (pointed) | T-5520; STATE "From tonight" |
| **BUG-038** | Guard, names the check does not read: let, read, printf -v, mapfile, for N in, getopts, namerefs; the name as data; the name built from parts; PowerShell Env: forms and casts; cmd /c "env -i x.sh"; a backstop that refuses a text naming part of the variable. After 1.0 (the director, 11 Oct, "Scoped") | internal/guard shell_env.go comment lines 26-39 (pointed) | V2c; STATE |
| **BUG-039** | Guard, roles that change nothing: an interpreter's write, a script run by name, a path built in a variable, a program's own writes, rsync, unzip, tar -x, patch, curl -o, wget -O, scp, compiler -o, formatter -w, git words that write only its store, PowerShell cmdlets outside the list, an MCP tool's server, and a BONSAI_TASK changed where the guard cannot read it are let through. Caught after, by the diff, git.dirty and rung 0. After 1.0 | internal/guard rolecmds.go comment lines 37-48 (pointed) | T-5520; 5.5.11 verifiers (VG3, 15 Oct) |
| **BUG-040** | Guard, attribute and ACL changes it cannot read: Get-Item x \| Set-ItemProperty IsReadOnly, a member set on a variable, [IO.File]::SetAttributes, find -exec chmod, setfacl --restore, setfattr --restore, macOS xattr; Git Bash's `icacls /c //deny ... //t`. Bonsai's own counts are backed by the repair in budget.go. After 1.0 | internal/guard attrwords.go comment lines 28-34 (pointed) | VG3, 15 Oct |
| **BUG-041** | Guard, `(Get-Item x).Delete()` or `.MoveTo()` on a state starts a fresh window | internal/guard shell.go and budget.go limits (said; not re-read) | VG3, 15 Oct; after 1.0 |
| **BUG-042** | A calls file emptied in its first 11 s, or put back from a copy as long as its window's start, counts afresh; a state moved aside once counts afresh; a rare false hold when a first call's line was never written; an agent with no meta file joins an open call of the last 10 minutes | internal/guard budget.go comment lines 26-40 (pointed for the moved-aside case; the rest said) | VG3, 15 Oct; after 1.0 |
| **BUG-043** | Plain `chmod 755 .` at a project's top is refused to a builder | internal/guard (said) | VG3, 15 Oct; after 1.0 |
| **BUG-044** | The settle wait in the recorder | internal/recorder (said) | VG3, 15 Oct; after 1.0 |
| **BUG-045** | Roles: an agent gone quiet makes no call so nothing stops it; a session started where the shell judge cannot read the command reads as the person's own; records are attributed, not authenticated (a write the guard cannot read can forge a state, brief, report or log line); a role that changes nothing can write through an interpreter, script or MCP server; a named session can spend its budget on purpose to be let stop (its ladder red); agent teams' teammates are not covered; with bonsai missing the second line lets reads and skills through with no budget. After 1.0 | internal/guard roles.go comment lines 44-52 (pointed) | 5.5 "What it cannot do"; VG3 |
| **BUG-046** | Guard, skills: a plugin's skill (<plugin>:<name>), a skill in a nested .claude/skills folder or an --add-dir folder, a file changed between the read and Claude Code's own, and frontmatter Bonsai's line reading cannot follow (block scalar, flow mapping) are not read; such a fork runs as an undeclared type | internal/guard skill.go comment lines 16-23 (pointed) | VG round 2 S1; after 1.0 |
| **BUG-047** | Guard, worktree with no bonsai.yaml of its own (branch made before the link, or one that deleted it): inside .claude/worktrees/ its edits are judged by the main checkout's lists (a person_only .claude/** refuses all of them); outside the main checkout it is judged by no lists. Rung 0 catches it at the climb | internal/guard guard.go comment lines 63-68 (pointed) | 5.4.3; after 1.0 |
| **BUG-048** | Rung 0 race (V2d) | internal/rung0 (said) | V2d; STATE "Known limits from earlier, after 1.0" |
| **BUG-049** | The gate's git timeout under load | internal/stopgate (said) | STATE "Known limits from earlier, after 1.0" |
| **BUG-050** | The machine record is writable by an agent's shell | the home's machine record (said) | STATE "Known limits from earlier, after 1.0" |
| **BUG-051** | A task file's grants are read from main's working tree | internal/guard/internal/recorder (said) | STATE "Known limits from earlier, after 1.0" |
| **BUG-052** | D2: a scratch link typed on the shell line is refused | internal/guard (said) | STATE "Known limits from earlier, after 1.0" |
| **BUG-053** | K2: a fake program first on the caller's PATH fools a rung; CI to run the ladder, a task done only with it green | internal/ladder (said) | the director, 13 Oct; STATE |
| **BUG-054** | Git LFS and git-crypt checkouts read dirty | internal/vergit or ladder git.dirty (said) | STATE "Known limits from earlier, after 1.0" |
| **BUG-055** | The packs' protected paths in the lock sit outside the settings record | internal/engine lock (said) | STATE "Known limits from earlier, after 1.0" |
| **BUG-056** | A `documents` kind added moves a pack's place | internal/engine (said) | STATE "Known limits from earlier, after 1.0" |
| **BUG-057** | The redactor's limits: 5.2's L1 and the shared `=>` and `:=` forms (whether they go in a named part before the release is the director's call) | internal/redact (shapes.go, values.go; said, not re-read) | STATE "Known limits from earlier" and "Whenever they like" |
| **BUG-058** | CDPATH and cdable_vars not read by the guard; Git Bash's /proc/cygdrive/ | internal/guard (said) | the director, 12 Oct (5.3 closes on two known limits) |
| **BUG-059** | A spent agent cannot stop a process it started (exempt it, or accept) | internal/guard budget.go (said) | V1 design question; waiting on the director |
| **BUG-060** | Git's own lock files are not deletable by exact name by an agent | internal/guard delete check (said) | STATE "Whenever they like" |
| **BUG-061** | One Windows skip of TestMemberGonePid in 20 | internal/agents or guard test TestMemberGonePid (said) | 5.5 end verifier; "Can wait" |
| **BUG-062** | The TestStrandedMoved fix (84533b07) and the PowerShell installer changes were never run on Windows by a person; their only Windows proof is CI | cmd/bonsai/stranded_test.go; install/install.ps1 | M-2026-10-18-e60a57d1; M-2026-10-18-25dea7b0; end Windows run (W1-W7, E1-E5 in the builders' reports) |
| **BUG-081** | Two workflow comments point at `design/plan.md` where step 5's plan (`design/plan-5.md`) is meant | .github/workflows/ci.yml:44, release.yml:3; 91d8a972 (pointed) | M-2026-10-18-ade22b5c (T-5704's stale-phrase grep) |
| **BUG-082** | (code side fixed in 9124cdc7, T-5698; the help schema's words wait for the next formats set) The `not-built` refusal and its docs still speak "of the rebuild", though no current word or flag reaches it; two old comments say the same | cmd/bonsai/word.go:285,335; internal/format/error.go:43; docs/reference/lists.md:130; formats/schemas/help.schema.json:207,287 (with its manifest); internal/workspace/checkout.go:7; internal/record/self.go:48; 91d8a972 (pointed) | M-2026-10-18-ade22b5c |
| **BUG-083** | The pack template's workflow says "Bonsai has no release yet" (5.7.6 changes it) | packs/template/.github/workflows/pack.yml:19; 91d8a972 (pointed) | M-2026-10-18-ade22b5c |
| **BUG-086** | On Windows the guard refuses a write through a file link to a folder (and a Write naming a file link to a file not made yet) as bad-path, though Windows writes through it: an over-refusal, kept because naming such a link by its target is wrong behind a junction (a write into Bonsai's records read as a decoy outside every project, verifier round 2) | internal/engine/final_windows.go, FinalPath (ErrDangling); fe1e8f9d (pointed) | CI run 39165347544; M-2026-10-18-937e348f, M-2026-10-18-c6ee1078 (T-5696) |
| **BUG-087** | The text layout prints a stray blank line before a next step at widths 60 to 80 | internal/textui (said by T-5698's builder); 71176d92 | M-2026-10-18-d46672d6 (T-5698) |
| **BUG-088** | operating-bonsai says only `bonsai logs` shows a refusal's rule; once T-5698 lands the refusal line shows it too | packs/base/skills/operating-bonsai/SKILL.md:103; 71176d92 (said) | M-2026-10-18-d46672d6 (T-5698) |
| **BUG-089** | The stop gate's refusal prints no rule word, and its error words do not use workspace.OSWords | internal/stopgate; 71176d92 (said) | M-2026-10-18-d46672d6 (T-5698) |
| **BUG-090** | The installed govulncheck (v1.2.0) and golangci-lint (built with go1.25) cannot read Go 1.27 code; agents run the pinned versions through `go run` instead | this machine's tools, not the repo; a82d18b5 (said by T-5700's builder) | M-2026-10-18-dce878aa (T-5700) |
| **BUG-091** | Several next-step checks in engine_test.go use `HasSuffix(nextOf(out), ...)`, which a trailing row after the next step would not fail; `endsWith` (T-5612) holds the end | cmd/bonsai/engine_test.go; 90158ade (said by T-5612's builder) | M-2026-10-19-a1efbe1c (T-5612) |

## Fixed

| Id | What went wrong | Fixed in | Task |
|---|---|---|---|
| **BUG-063** | A stamped build's home confinement failed open past 64 path steps (RealPath returned the path as written); `settings set` wrote outside the scratch root | 6dbc4c87 (one walk, fails closed), 43e8b5fb | T-5694 (V1 B1; round 2 PASS M-2026-10-18-53079714) |
| **BUG-064** | A shell write target the guard could not place was let through (pathForms error ignored in rolecmds/attrwords) | 27f193eb | T-5694 |
| **BUG-065** | On Windows junctions were neither followed by EvalSymlinks nor marked as links, so a home under a junction could pass the confinement | 6dbc4c87 (FinalPath moved into engine) | T-5694 |
| **BUG-066** | The guard and the engine had two copies of the link walk (guard.resolve, engine.RealPath) | 6dbc4c87; guard paths.go:234 now calls engine.WalkLinks | T-5694 (T-5692 proposed it) |
| **BUG-067** | A relative target of a link whose target did not exist was joined to the link's folder by spelling, so a `..` was read lexically (V1 N1) | 6dbc4c87 | T-5694 |
| **BUG-068** | A stamped build's home confinement refused a home not made yet under a linked or short-named root, and let a dangling link inside the root that points outside pass (CI TestPersonOnly red on 16 Oct runs 38908665177, 38908079083) | 5648071d (branch head; engine.RealPath, FolderHolds) | T-5692 |
| **BUG-069** | update --yes never cleaned the pack cache (CleanCache had no caller) | 1b271a34 (caller at cmd/bonsai/engine.go:310) | T-5601 |
| **BUG-070** | bonsai --version did not name the commit the build was made from | 843913a1 (main.go:58 prints it) | task not read |
| **BUG-071** | install.ps1's elevated child compiled C# with Add-Type into the person's %TEMP% (V1 B2) | d4bd9f0c | T-5606 |
| **BUG-072** | The elevated child ran cmdlets (Start-Sleep, Set-Location, New-Object) a user-writable module on PSModulePath could shadow (V1 W5) | 204aa212 | T-5606 |
| **BUG-073** | install.sh decided the real place by the target's spelling (//, /./) so root-owner checks were skipped; a linked folder was followed (V1 N2) | 888c72ef | T-5606 |
| **BUG-074** | Both installers moved the copy into a target that was a folder and left an executable behind (V1 N3) | bf9e570f | T-5606 |
| **BUG-075** | install.ps1's rights check ran after the file was in place and its mask lacked 0x10 and 0x100 (V1 N4) | 8e1b0af3 | T-5606 |
| **BUG-076** | CI install job: Linux runner's /usr/local/bin was group/other writable; PowerShell's -Filter 'bonsai.exe.*' matched bonsai.exe itself; the install test wanted the target as spelled, not its long form | 2223a5d2, 8c3552f5 | T-5606 |
| **BUG-077** | TestStrandedMoved failed on Windows (t.Chdir held the checkout open so the rename was refused); a test bug | 84533b07 | T-5601 |
| **BUG-078** | TestLadderCheckWordsAndTheMachineMark hard-coded the machine-marked words without `stranded` | T-5601 branch (internal/format/lists_test.go, accepted outside the Owns; commit not read) | T-5601 |
| **BUG-079** | workflow's interview skill carried the studio's "Before you stop" paragraph | the workflow pack at a42a4f4ca171 (grep of the installed copy: absent) | the workflow pack's ref move; task not read |
| **BUG-080** | The operating skill sat at 12262 of its 12288 bytes | split in two, base:operating-bonsai now 9069 bytes (wc -c) | T-5609 |
| **BUG-005** | T-5696's own task file held a Windows home path in done_when[0] (bonsai check's one finding) | the task file, rewritten by the orchestrator 19 Oct (uncommitted until the records commit) | T-5696 |
| **BUG-027** | Contract section 5.4's workflow.model listed no haiku | design/contract.md, the "Changed 14 Oct" note under the workflow labels (it gains haiku) | T-55xx (5.5's labels change) |
| **BUG-001** | CI's windows job fails on a runner whose temp folder is an 8.3 short name (the CI runner's user folder): RealPath returns the long form where TestFolderHoldsPastTheOldLimit and TestARelativeLinkTargetIsReadFromTheLinksRealFolder want the short one, and the guard refuses a deep shell write bad-path where TestAShellWriteTheGuardCannotPlaceIsRefused wants bonsai-files | 9124cdc7 (fe1e8f9d, 6b02441d) | T-5696 |
| **BUG-002** | The guard's refusal for a path it cannot place says "on Windows the guard cannot tell where the path leads" for any walk error, so a link loop or a path past the limit on Linux gets Windows words (the refusal itself is right) | 9d7ace4e | T-5696 |
| **BUG-003** | TestEveryRule holds bad-path on Linux only through /proc (it skips unless GOOS is linux), though a link loop or deep link now reaches bad-path on Linux; the link-loop case is in deeppath_test.go:88 only | 9d7ace4e | T-5696 |
| **BUG-004** | On Windows the link walk reads a symbolic-link target that has a drive and no root (C:foo) as absolute from that drive's root, where Windows resolves it against the drive's current folder | fe1e8f9d | T-5696 |
| **BUG-010** | check's missing-path finding counts a path that git ignores as present (it only stats the file) | e39ac070 (rebased into 9124cdc7) | T-5698 |
| **BUG-012** | update's human output runs two plugin lines together | c25af2ac (rebased) | T-5698 |
| **BUG-024** | The guard's stderr refusal does not print its rule word (the skill points to bonsai logs) | 22bb2c4c (rebased) | T-5698 |
| **BUG-030** | The recorder's taskAllows and the guard's allowsOf are the same read in two places | 03d65e3c (rebased) | T-5698 |
| **BUG-031** | underHome and fromHome are two copies of one rule | 74c18b24 (rebased) | T-5698 |
| **BUG-032** | workspace.VerifyMain quotes a system error in its message | 41e0f1e4 (rebased) | T-5698 |
| **BUG-084** | A delete the guard cannot place gives "(a drive letter and folders)" as its next step on every system | 9ac0c401 (rebased) | T-5698 |
| **BUG-085** | The reference generator takes any one argument as the repository's root, so `go run ./gen -h` writes every generated file under a folder named `-h` | cc840072 (rebased) | T-5698 |

## Known, not bugs

- With bonsai missing or failing, every edit, shell command and agent start is refused (the second line lets reads, searches and skills through): CLAUDE.md "The way back".
- The guard refuses more than it must where it says so: a `~/` redirect read as relative (BUG-011 if confirmed), a link loop on Linux refused as bad-path with Windows words (BUG-002 is only the words).
- A stamped (test) build refuses a home outside its scratch root, test-build (exit 4); fail closed by design.
- A walk that cannot look at a name (chmod 000) reads it as not made, as before (43e8b5fb); the write then fails for the same user.
- Defender quarantines cmd/bonsai's test-built bonsai.exe on Windows; CI's windows job is the proof (the director, 14 Oct).
- On Windows the PowerShell tool ignores Read and Edit deny rules (STATE "Facts measured").
- A newer brief to an agent opens a fresh budget window (budget.go comment line 29).
- bonsai check's `note: no install.json in the Bonsai home` until the installer writes it (ran: bonsai check).

## Watched, not bugs (flaky, not repeated)
- CI's install (windows-latest), "the running copy ended during the install" (ci.yml:281: a `bonsai hook guard` copy started 500 ms before an install): failed once on 9124cdc7 (run 39169591437), passed on its re-run and on the five commits before.
- TestUsageWaitsForTheLastReply (internal/recorder/usage_test.go:350): failed once on CI's Windows, run 39116286027 (84533b07); passes in every later run (ran: gh --log-failed over the last 60 runs). STATE: fix if it repeats.
- TestStatusLineSpeed (cmd/bonsai/statusflags_test.go:340): failed once under load 16 (M-2026-10-18-60245142), passed on rerun and in the climb.
- internal/redact TestLinearTime under load (14 Oct): fix only if CI meets it.
