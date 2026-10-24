# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - Unreleased

> **Bonsai 1.0 is a new product.** It gives a project the structure Claude Code agents work inside: formats, packs,
> guards, a recorder and a ladder that proves work is done, from one Go program on Linux (and WSL) and Windows. It
> shares no code with 0.4.3, the old product's last release, which stays at tag `v0.4.3`.

### Added
- **The `bonsai` program** (`cmd/bonsai`), with these words: `init`, `update` and `unlink` (link a project to packs,
  bring them to the refs in `bonsai.yaml`, take Bonsai out again), `status` and `check` (one workspace at a glance;
  findings on the lock and the files), `hook` (what Claude Code's hook lines call), `ladder`, `ask`, `answer` and
  `asks`, `logs` and `log`, `message` and `messages`, `agents`, `settings` and `labels`. Every word takes `--help`;
  every word but `hook` takes `--json`; without a terminal none prompts; each refusal names its next step.
  `bonsai --help --json` prints every word, flag and exit code as one document.
- **The formats** (`formats/`): 29 JSON Schemas (`bonsai check --schema <format>` prints any), examples, trick files
  for readers with their expected outcomes, a manifest of every file's bytes, and a Go test.
- **Packs**: `base` (`packs/base/`, Bonsai's own: the templates of its documents as skills, the operating skills, the
  walls and Bonsai's four labels; it carries no code) and a template for a new pack (`packs/template/`). The
  `workflow` pack (roles, lanes and protocols) lives in its own repository.
- **The guard**: Bonsai's hook lines run the `bonsai` installed at `/usr/local/bin/bonsai` (Linux and WSL) or
  `C:\Program Files\Bonsai\bonsai.exe` (Windows), never one found on the `PATH`, and refuse every edit and shell
  command while it is missing.
- **The installers** (`install/install.sh`, `install/install.ps1`) for those two places.
- **The ladder**: `bonsai ladder --task <id>` runs the project's rungs and writes the result a named session's stop
  gate reads before the session may end.
- **A `windows` CI job**: the tests and `go vet` on `windows-latest`, beside the Linux `test` job.
- **The stop gate names its rule**: every block ends with a `rule:` line, as the guard's refusals do, and the rule
  words are checked against the words a refusal may not print.
- **One clock, in UTC**: `bonsai now` prints the time to write in a record (`--day`, `--iso`, `--local`); every time
  Bonsai writes or checks is UTC (formats set 11, `bonsai.now/1`), and printed times show in the person's zone with
  its offset (`--utc` prints them as stored). The guard refuses an edit that writes a fresh stamp far from the clock
  (`time-not-now`), and a hook line puts the time in front of the model at each prompt and subagent start.
- **Answering by letter**: `bonsai answer <key> --choice B` (or `b`, or `2`) answers a Decide with that option;
  `bonsai ask --status` shows each option's letter. The record keeps the option's whole text.
- **Other agents count as agents**: beside Claude Code's own marker, `AI_AGENT`, `PI_CODING_AGENT`,
  `CODEX_THREAD_ID`, `GEMINI_CLI`, `CURSOR_AGENT`, `OPENCODE` and `GOOSE_TERMINAL` make the person's commands refuse
  (`update --allow-exec`, `unlink`, accepting looser checks, `settings set`, `labels`), naming the marker found.
- **The stop gate reads the run report as `bonsai check` does**: a named session may not end on a run report check
  would reject (`stop-bad-run-report`).
- **`bonsai check` warns while `bonsai.yaml` is not committed** (`config-uncommitted`): on a project with no commit
  yet, or one whose `bonsai.yaml` differs from what is committed, since a climb and the stop gate refuse it then; the
  next step is the person's commit.
- **`docs/features.md`**: every capability by area, with its limits and its open bugs.
- Where to start: [`README.md`](README.md), and `docs/install.md` for the install lines.

### Changed
- **`go.mod`** needs only the standard library.
- **`make build`** builds with `-trimpath`.
- **A pack kept in a subfolder** (`base`, in `packs/base`) counts only tags that start with its name (`base-v1.0.0`)
  as its releases, so `bonsai status` no longer reads Bonsai's own `v` tags as newer `base` releases. A pack at its
  repository's top (`workflow`) keeps its plain `v` tags.
- **After a pack is added or moved**, `init` and `update` say that a Claude Code session already open in the project
  loads the pack's skills only after `/reload-plugins` or a fresh start.
- **The guard's refusal after a `cd` it cannot follow** (a role that changes nothing, writing a relative path after
  `cd $DIR`) says so and asks for the file by its full path; what it refuses is unchanged.
- **A climb's printed lines** name no other program: the leftovers line gives a count and process ids (the names stay
  in the result file and in `--json`), and no printed line names a step of Bonsai's own build plan.

### Removed
- **The old product's code** (the commands, the embedded catalog and guides, the documentation site and its deploy
  workflow) and **`make install`**: `go install` would put a `bonsai` in front of the installed one on the `PATH`.
- **The old product's demo** (`assets/demos/`): the README shows 1.0's own, in `docs/demos/`.

### A 0.4.3 workspace
`bonsai init`, `update` and `unlink` refuse a checkout that holds 0.4.3's `.bonsai.yaml` or `.bonsai-lock.yaml`
(`old-workspace`, exit 4) and change nothing in it. Moving a 0.4.3 workspace to this Bonsai is not part of 1.0: keep it
on 0.4.3, whose binaries and `brew install LastStep/tap/bonsai` stay where they are.

## The old product

The old product's changes, 0.1.0 to 0.4.3, are in this file at tag
[`v0.4.3`](https://github.com/LastStep/Bonsai/blob/v0.4.3/CHANGELOG.md). 1.0 carries none of them.
