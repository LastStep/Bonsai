# Contributing to Bonsai

Bonsai is being rebuilt from scratch (see [`README.md`](README.md)). The rebuild is run by its director, Rohan, with
Claude Code agents as the team, in the order [`design/plan.md`](design/plan.md) sets. Work lands on `main` directly,
in small commits, each pushed only after its proof passes; there are no pull requests while the rebuild runs. To
report a problem or suggest something, open an [issue](https://github.com/LastStep/Bonsai/issues); for a security
problem, follow [`SECURITY.md`](SECURITY.md).

## Bonsai guards its own repository

From step 5.4's switch, when this repository links itself, Bonsai runs on its own work. Its committed
`.claude/settings.json` runs the installed `bonsai` (`C:\Program Files\Bonsai\bonsai.exe`, then `/usr/local/bin/bonsai`)
on every edit and shell command in a Claude Code session here, and **refuses the call when neither is there**, on
purpose (fail closed): a guard that a missing program switched off would guard nothing. Git, editors and Go are not
affected; only Claude Code sessions are.

To work here with Claude Code, either:

- **install Bonsai at that place**, built from this repository (on Windows, as an administrator, copy a
  `GOOS=windows` build to `C:\Program Files\Bonsai\bonsai.exe` instead):

  ```bash
  CGO_ENABLED=0 go build -trimpath -o ~/bonsai-build/bonsai ./cmd/bonsai
  sudo install -o root -g root -m 0755 ~/bonsai-build/bonsai /usr/local/bin/bonsai
  ```

  With it installed, the guard also refuses changes to the protected files that [`bonsai.yaml`](bonsai.yaml) lists
  (among them `bonsai.yaml`, `CLAUDE.md`, `.gitignore`, `go.mod` and `.github/`) that no running task grants, and a
  recursive delete that does not name what it deletes;
- **or start Claude Code with Bonsai's hooks off** for your clone:

  ```bash
  claude --settings '{"disableAllHooks": true}'
  ```

  That turns off every hook in the session, Bonsai's and your own, and your own status line with them.

## How work runs here

Read these first, in this order:

1. [`CLAUDE.md`](CLAUDE.md): how a session works, the rules for every change, GitHub and safety. Every agent follows it.
2. [`STATE.md`](STATE.md): where Bonsai stands and the one thing to do next. It is rewritten, never appended.
3. [`design/plan.md`](design/plan.md): what is built, in what order, and how each part is proved.

Each piece of the plan is a task, a file in [`records/tasks/`](records/tasks/) saying what it may change and what
proves it. It is built in its own git worktree and branch, recorded in a run report under
[`records/runs/`](records/runs/), and, for big or risky work, checked by a fresh verifier before it lands.

## Building and testing

Go 1.25: the module's `go` and `toolchain` lines say which; an older `go` downloads that toolchain itself. The code
uses the standard library only (Windows calls go through `syscall`).

```bash
go build -o ~/bonsai-checks/bin/bonsai ./cmd/bonsai
~/bonsai-checks/bin/bonsai --version
go test ./...
go vet ./...
```

`make build OUT=~/bonsai-checks/bin/bonsai` does the same build. Never `go install`: it puts a `bonsai` in front of
the installed one on the PATH.

The proof of a change is `bonsai ladder --task <id>` green, run by the installed `bonsai`: the rungs in
[`bonsai.yaml`](bonsai.yaml) (the guard's check of what the task changed, `go vet`, the schema compare, every test plain
and with the test fault switch, and git integrity). Beside it, the tests run natively on Windows, and CI is green on
both sides (`.github/workflows/ci.yml`: `test`, `windows`, `lint`, `govulncheck`).

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
