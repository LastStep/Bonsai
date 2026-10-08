# Contributing to Bonsai

Bonsai is being rebuilt from scratch (see [`README.md`](README.md)). The rebuild is run by its director, Rohan, with
Claude Code agents as the team, in the order [`design/plan.md`](design/plan.md) sets. Work lands on `main` directly,
in small commits, each pushed only after its proof passes; there are no pull requests while the rebuild runs. To
report a problem or suggest something, open an [issue](https://github.com/LastStep/Bonsai/issues); for a security
problem, follow [`SECURITY.md`](SECURITY.md).

## How work runs here

Read these first, in this order:

1. [`CLAUDE.md`](CLAUDE.md): how a session works, the rules for every change, GitHub and safety. Every agent follows it.
2. [`STATE.md`](STATE.md): where Bonsai stands and the one thing to do next. It is rewritten, never appended.
3. [`design/plan.md`](design/plan.md): what is built, in what order, and how each part is proved.

Each part of the plan is built in its own git worktree and branch, recorded in a run report under
[`records/runs/`](records/runs/), and, for big or risky work, checked by a fresh verifier before it lands.

## Building and testing

Go 1.25: the module's `go` and `toolchain` lines say which; an older `go` downloads that toolchain itself. The code
uses the standard library only, plus `golang.org/x/sys` once the Windows code needs it.

```bash
go build -o ~/bonsai-checks/bin/bonsai ./cmd/bonsai
~/bonsai-checks/bin/bonsai --version
go test ./...
go vet ./...
```

`make build OUT=~/bonsai-checks/bin/bonsai` does the same build. Never `go install`: it puts a `bonsai` in front of
the installed one on the PATH.

The proof of a change, until Bonsai has its own ladder: `go test ./...` and `go vet ./...` on Linux and natively on
Windows, and CI green on both (`.github/workflows/ci.yml`: `test`, `windows`, `lint`, `govulncheck`).

## Rules every change follows

The full list is in [`CLAUDE.md`](CLAUDE.md). In short:

- Every stored or printed path uses forward slashes; output is byte-stable; fingerprints read line endings as LF; no
  test needs a symlink or a file mode to pass on Windows.
- Every command runs unattended: `--json` everywhere but `hook`, no prompt without a terminal, ASCII human output, and
  every refusal names the next step.
- `formats/` changes only together with its manifest (`formats/README.md`).
- Nothing private in any file or commit: no home-folder path, machine name, email address, token or server address.
  The repository is public.
- Commits are small, titled `area: what` in plain words.

## The old product

0.4.3's code, catalog and guides are at tag [`v0.4.3`](https://github.com/LastStep/Bonsai/tree/v0.4.3); its
contributing guide is that tag's `CONTRIBUTING.md`.
