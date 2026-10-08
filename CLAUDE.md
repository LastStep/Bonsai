# Bonsai: guide for Claude Code

Bonsai is being rebuilt from scratch. 0.4.3 (May 2026) is the old product's last release, kept at its tag and in git
history; everything in this repo now serves the rebuild. Rohan is the director; Claude agents are the team.

## What Bonsai is, and is not

- **Is:** the structure inside each project. Formats (what a task, a run report, a log line and a question to the person
  look like); packs (roles, lanes and protocols, delivered as Claude Code plugins); guards that stop mistakes; a recorder
  that keeps a clean log; a ladder that proves work is done. One Go program, nothing else to install, on Linux and
  Windows.
- **Is not:** anything that applies an action. It never moves a task, applies an approval or starts an agent. No daemon,
  no network service, no terminal UI. Registries, dashboards, notifications and deploys belong to its users.
- **Its first user** is Rohan's studio, Trinetra, which adopts the formats when it links (step 7). This repo names no studio file in its code.

## Where the truth lives

| Question | File |
|---|---|
| Where does Bonsai stand, and what is next? | `STATE.md`. Read it first, every session |
| What do we build, in what order, proved how? | `design/plan.md` |
| What is Bonsai, exactly? | `design/bonsai-spec.md` (130 KB). Grep its `## N.` sections; never read it whole |
| What do the formats say? | `design/contract.md` (100 KB). The same: by section |
| What did Rohan decide on each format? | `design/format-review.md` (confirmed 7 Oct: the formats are final) |
| What is it for; what does correct mean? | `design/one-pager.md` |
| What happened before? | `records/runs/`. Read one only when `STATE.md` or the plan points to it |
| The shared formats, schemas and trick files | `formats/` (from part 0) |

The specs were written in the studio's repo: their "this repo" means the studio's, and the paths they cite are its.

## How a session works

- **Rohan opens a session in `~/Servers/Bonsai`** (the main checkout, on `main`). It is the **orchestrator**: it reads
  `STATE.md` and the plan, keeps its own context lean, and dispatches builder and verifier subagents. Bonsai has no role
  files, so every brief carries the job, the plan's part, the sections to read and the rules below that apply. It reads
  the agents' reports, not their file dumps.
- **Builders** work in a plain git worktree beside the clone (`git -C ~/Servers/Bonsai worktree add
  ~/Servers/Bonsai-<part> -b <part> main`), never the Agent tool's isolation worktrees. They commit; they never push.
- **A plan is reviewed by a fresh Opus agent** before it reaches Rohan.
- **A fresh verifier** (Opus, fresh context) for big or risky work only (Rohan's rule): part 0, the reader, guards and
  hooks, CI and release, security, and the end of each plan step; other parts close on green tests and CI plus the
  orchestrator's read of the diff, which the run report says. Related light parts share one verifier. A verifier
  reads the plan's part, the cited sections and the diff, re-runs the tests itself, and
  passes or fails the work. It fixes nothing. Nothing is done because an agent says so.
- **Models:** only the latest Opus, Sonnet, Haiku and Fable. Opus for verifiers, plan reviews, security, guards, release
  and decisions, and builders on risky code; Sonnet for sweeps, drafts and mechanical runs; Haiku for small bookkeeping
  and audit jobs (listing files, sorting or counting things, checking a list against a source, filling a run report's
  rows); Fable for visual design. Haiku reports facts; it judges nothing and its findings are read before they count.
  Every report says which model ran what.
- **Run reports** in `records/runs/R-<date>-<topic>.md`, opened before the first edit and appended as the work goes: a
  log, not a summary. Each lists every run with its model, start, end and minutes; those rows are the hours the stop
  lines count.
- **Proof** until Bonsai has its own ladder (spec step 5.4): `go test ./...` and `go vet ./...` in WSL and natively on
  Windows, CI green on Linux and Windows, and a fresh verifier.
- **`STATE.md`** is rewritten, never appended, whenever where Bonsai stands changes.

## Rules for every change

- **No `go install`, ever:** it puts a `bonsai` in front of the installed one on the PATH. Build with
  `go build -o <scratch folder>/bonsai ./cmd/bonsai`.
- Go: the module's `go 1.25` with its `toolchain` line (WSL's Go at `/usr/local/go/bin/go` downloads it). The standard
  library plus `golang.org/x/sys`; no terminal-UI library and no general YAML library.
- **Rules learnt from Windows** (the old code failed 29 tests only there), a review item in every change: every stored
  or printed path uses forward slashes; no test needs a symlink or a file mode to pass on Windows, and a test that cannot
  run on one OS says why; fingerprints read line endings as LF; output is byte-stable (no map-order iteration); Windows
  renames retry on busy errors; a hook line never calls `bash` by name.
- **Every template and pack file documents itself** (its purpose, when to use it, every field with an example), and a
  field change updates its docs in the same commit.
- **Every list has one home** (Bonsai's schemas or a pack's declarations); the reference page is regenerated with any
  list change.
- Every command runs unattended: `--json` everywhere but `hook`, no prompt without a terminal, ASCII human output, and
  every refusal names the next step.
- **`formats/` changes** only with its manifest.
- **Windows:** Windows Go cannot build from WSL's disk. Windows runs happen under `%USERPROFILE%\bonsai-checks\` with
  `"/mnt/c/Program Files/Go/bin/go.exe"`. Every Windows-side repo, worktree and CRLF checkout is made with Windows git by
  its full path, `"/mnt/c/Program Files/Git/cmd/git.exe"`. Never run Linux git in a Windows checkout.
- Commits: small, `area: what`, plain words, ending with the session's co-author line.

## GitHub (`LastStep/Bonsai`, public)

- Agents act on GitHub as `LastStep`, Rohan's admin account, so guards there stop accidents, not intent.
- **Builders never push.** The orchestrator fast-forwards `main` and pushes it only after the work's proof passes, then checks
  CI for that commit after every push:
  `gh api repos/LastStep/Bonsai/commits/<sha>/check-runs --jq '.check_runs[] | [.name, .status, .conclusion] | @tsv'`
  (or `gh run list -R LastStep/Bonsai -L 10`; this machine's gh has no `--branch`). Red CI is fixed forward.
- Work lands on `main` directly: no pull requests, no long-lived branches. Force pushes and deletion are blocked.
- **Nobody tags or releases.** `release.yml` stays disabled; releases are Rohan's word at spec step 5.7. No agent changes
  a setting, secret, ruleset or workflow switch, or merges or closes a pull request.

## Working with Rohan

- He is the director and writes, reviews and verifies no code. Agents own the code end to end: never leave a code
  problem for him to catch.
- Ask in plain words, context first, one question at a time, with two to four options and a recommendation. Never ask
  what a file answers.
- Hand him manual work (hand checks, settings, installs) in one batch with exact commands: bash for WSL, one command per
  line; PowerShell 5.1 for Windows, one command per line, no `&&`.
- A small fix stays small; ask before chasing edge cases.
- Write for him in plain words: what he can now do, with evidence (test counts, the CI run, the command to try). Never
  claim something is good for him; he judges that.

## Safety

- **No secrets anywhere: the repo is public.** No home-folder path, machine or tailnet name, email address, token or server
  address in any file or commit message; write `~/...` and `%USERPROFILE%`.
- Never touch `~/ZenGarden/Bonsai` (an old clone with branches on no remote).
- Tests use temp folders (`t.TempDir()`), never a real home (`~/.bonsai`, `~/.claude`) or a real project.
- Bonsai's work changes nothing outside this repo, its test pack and the scratch folders (`~/bonsai-checks`,
  `%USERPROFILE%\bonsai-checks`).
- **Stop every process you start** (builds, tests, test sessions, watchers) and check with `ps -eo pid,etime,cmd` before
  you finish. Never kill what you did not start: another agent's run looks the same.
