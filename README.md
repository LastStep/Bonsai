# Bonsai

Bonsai gives a project the structure Claude Code agents work inside: formats, packs, guards, a recorder and a ladder
that proves work is done. It is one Go program with nothing else to install.

[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/LastStep/Bonsai/ci.yml?label=CI)](https://github.com/LastStep/Bonsai/actions?query=workflow%3ACI)

## What Bonsai is, and is not

**It is:**

- **Formats**: what a task, a run report, a log line and a question to the person look like.
- **Packs**: roles, lanes and protocols, delivered as Claude Code plugins.
- **Guards** that stop mistakes, a **recorder** that keeps a clean log, and a **ladder** that proves work is done.

**It is not** anything that applies an action. It never moves a task, applies an approval or starts an agent. It has
no daemon, no network service, no terminal UI and no screens or web page of its own: it is a command-line tool, and
every picture of its data is its user's to draw.

## Install and check it

[`docs/install.md`](docs/install.md) has the install lines for a release, and how to check the download with `gh`
before anything from it runs.

Bonsai installs in two places, and its hook lines look nowhere else: `/usr/local/bin/bonsai` on Linux and WSL, and
`C:\Program Files\Bonsai\bonsai.exe` on Windows. Both are places only an administrator writes. The guard runs the
`bonsai` installed there, never one found on the `PATH`, so an agent cannot swap the program that judges its own
edits; while it is missing, the guard refuses every edit and shell command in a Claude Code session of a linked
project.

## A first project

The steps below take a new project from nothing to a first proof. They need `bonsai` installed as above, `git`, and
Claude Code (`bonsai status` lists the oldest version the pack needs).

1. Make a folder and a git repository in it:

   ```bash
   mkdir demo && cd demo && git init
   ```

2. Link the project to Bonsai's own pack, `base`, by the repository, the pack's folder in it and a ref (the tag
   `base-v1.0.0` once it exists, else a 40-character commit of `main`). Without `--yes` this only previews:

   ```bash
   bonsai init --name demo --source https://github.com/LastStep/Bonsai --path packs/base --ref <tag or commit>
   bonsai init --name demo --source https://github.com/LastStep/Bonsai --path packs/base --ref <tag or commit> --yes
   ```

   `base` runs no code, so `--yes` alone links it. This writes `bonsai.yaml` (a comment on every line), the lock in
   `.bonsai/`, a block in `CLAUDE.md` and Bonsai's lines in `.claude/settings.json`.

3. Look at what you have:

   ```bash
   bonsai check
   bonsai status
   ```

   `check` lists findings, each with its next step. Until Claude Code has trusted the folder it warns that the
   `base` plugin cannot be installed yet.

4. Open Claude Code in the folder, accept its question about trusting the folder (Bonsai never answers it for you),
   quit it, then run `bonsai update --yes`. That installs the `base` plugin on this machine at the locked commit.

5. Open Claude Code there again and ask for something small. Bonsai's guard now judges its edits and shell
   commands, and the recorder writes the session's log in `.bonsai/local/`.

6. To prove work with the ladder, add a rung to the `ladder` in `bonsai.yaml` (a command that tests your project) and
   list its number in `ladder_floor`, write a task file for the work, and commit both. A rung added loosens what the
   checks hold, so the first climb is refused until you run `bonsai update --yes` in a terminal of your own, which
   records it on this machine. Then climb:

   ```bash
   bonsai ladder --task <task id>
   ```

   `base`'s skills `base:workspace` and `base:task` show a rung and a task file.

Every command takes `--help`, and every command but `hook` takes `--json`. For an agent, `bonsai --help --json`
prints every command, flag and exit code as one document (`bonsai.help/1`).

## The packs

- **`base`** (`packs/base/`): Bonsai's own pack. The templates of its documents as skills, the operating skills, the
  walls that stop agents reading the files that hold keys and logins, and Bonsai's four labels. It carries no code.
- **The template** (`packs/template/`): a minimal valid pack. A new pack starts as a copy of it in its own repository.
- **`workflow`**: the roles, lanes and protocols for a team of agents, kept in its own repository,
  `LastStep/bonsai-workflow`.

## Platforms

Linux (and WSL) and Windows. CI runs the tests on both. A release is built for Linux, Windows and macOS on `amd64`
and `arm64`, but the installers cover Linux and Windows only. macOS is built and untested, with no installer; the
guard looks only in the two places above, so a Mac user would copy the program there by hand.

## The old product, 0.4.3

0.4.3 is the old product's last release and gets no further ones. It stays where it is: its source at tag
[`v0.4.3`](https://github.com/LastStep/Bonsai/tree/v0.4.3), its binaries on the
[v0.4.3 release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.3), and `brew install LastStep/tap/bonsai`
still installs it. Homebrew does not install this Bonsai, because it puts programs outside the two places the guard
trusts. A workspace made by 0.4.3 is refused by this Bonsai and left as it is (`old-workspace`, exit 4).

## Where the design lives

| What | Where |
|---|---|
| What Bonsai is for, and what correct means | [`design/one-pager.md`](design/one-pager.md) |
| What Bonsai is, exactly | [`design/bonsai-spec.md`](design/bonsai-spec.md) |
| What the formats say | [`design/contract.md`](design/contract.md) |
| The formats: a JSON Schema for each, examples, and trick files for readers | [`formats/`](formats/) |
| Every list the code holds, as a reference page | [`docs/reference/lists.md`](docs/reference/lists.md) |
| What changed, release by release | [`CHANGELOG.md`](CHANGELOG.md) |
| How work is done here | [`CONTRIBUTING.md`](CONTRIBUTING.md) |
| How to report a security problem | [`SECURITY.md`](SECURITY.md) |

## Licence

[MIT](LICENSE).
