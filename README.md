# Bonsai

**Bonsai is being rebuilt.** 0.4.3 is the old product's last release. The new Bonsai has no release yet: this
repository holds its design, its formats and a stub `bonsai` that answers `bonsai --version` and nothing else.

[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/LastStep/Bonsai/ci.yml?label=CI)](https://github.com/LastStep/Bonsai/actions?query=workflow%3ACI)

## What the new Bonsai is

The structure inside each project that Claude Code agents work in, given by one Go program with nothing else to
install, on Linux and Windows:

- **Formats**: what a task, a run report, a log line and a question to the person look like.
- **Packs**: roles, lanes and protocols, delivered as Claude Code plugins.
- **Guards** that stop mistakes, a **recorder** that keeps a clean log, and a **ladder** that proves work is done.

It never applies an action itself: no daemon, no network service, no terminal UI.

## Where things are

| What | Where |
|---|---|
| What Bonsai is for, and what correct means | [`design/one-pager.md`](design/one-pager.md) |
| What is built, in what order, proved how | [`design/plan.md`](design/plan.md) |
| Where the rebuild stands now | [`STATE.md`](STATE.md) |
| The formats: a JSON Schema and an example for each, and trick files for readers | [`formats/`](formats/) |
| How work is done here | [`CONTRIBUTING.md`](CONTRIBUTING.md) |

## The old product, 0.4.3

It stays where it is: its source at tag [`v0.4.3`](https://github.com/LastStep/Bonsai/tree/v0.4.3), its binaries on
the [v0.4.3 release](https://github.com/LastStep/Bonsai/releases/tag/v0.4.3), and `brew install LastStep/tap/bonsai`
still installs it. Its changes are in [`CHANGELOG.md`](CHANGELOG.md).

## Licence

[MIT](LICENSE).
