# Installing Bonsai

<!-- docs/install.md: how a person installs Bonsai 1.0.1 on each side, what each line proves, what to do when one
fails, the way out, and how to check an installed copy. The two install blocks below are held byte for byte to the one
table Bonsai writes its install lines from (internal/engine/release.go, releaseLines) by the test
internal/selfcheck/install_test.go: change a line there and here in the same commit. For a newer release, Bonsai
itself gives the same seven lines with that release's number: `bonsai status --full --json`, `bonsai_release.lines`. -->

Bonsai is one program, `bonsai`, installed once per release on each side: Linux (and WSL) and Windows. You type the
install lines yourself, one at a time; Bonsai never installs or updates itself, and an agent hands you these lines but
never runs them. Each side takes seven lines. The first four download the release and check its fingerprint; the fifth
checks that GitHub built it from Bonsai's own release workflow; only then are the files unpacked and the installer run.

## What you need

- **Linux and WSL:** `curl`, `sha256sum`, `tar` and `sh`, and `sudo` for the installed place, `/usr/local/bin/bonsai`.
- **Windows:** PowerShell 5.1 (the one Windows ships) and `curl.exe` (in Windows 10 and later). The installer asks
  one Windows question to write to `C:\Program Files\Bonsai`.
- **Both sides:** GitHub's command line, `gh`, **version 2.97 or newer**, logged in (`gh auth login`; `gh --version`
  names the version). The fifth line needs it: `--signer-workflow` came in 2.51 and `--source-ref` in 2.68, and 2.97
  fixed GHSA-mm27-mwq9-fr5g, where a lookalike workflow name could pass the check.
- `gh`, like `sh` and `tar`, is found on the PATH, so the one that runs is the first the PATH names: `which -a gh` (on
  Windows, `where.exe gh`) should name the system's own `gh` first.

## Install on Linux and WSL

In a terminal, one line at a time:

```bash
cd "$(mktemp -d)"
curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/bonsai_1.0.1_linux_amd64.tar.gz
curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/checksums.txt
sha256sum -c --ignore-missing checksums.txt
gh attestation verify bonsai_1.0.1_linux_amd64.tar.gz -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.1 --deny-self-hosted-runners
tar -xzf bonsai_1.0.1_linux_amd64.tar.gz
sh install.sh
```

The fourth line must print `OK`; the fifth, `Verification succeeded`. The last ends with Bonsai's fingerprint, the
`bonsai` programs the PATH finds (`/usr/local/bin/bonsai` alone) and `bonsai 1.0.1 (commit ...)`. On an arm64
machine, write `arm64` for `amd64` in the archive's name, on every line that names it.

## Install on Windows

In a normal PowerShell (not as administrator), one line at a time:

```powershell
cd (New-Item -ItemType Directory (Join-Path $env:TEMP ("bonsai-" + [guid]::NewGuid())))
curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/bonsai_1.0.1_windows_amd64.zip
curl.exe -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/checksums.txt
(Get-FileHash bonsai_1.0.1_windows_amd64.zip -Algorithm SHA256).Hash -eq ((Select-String -Path checksums.txt -SimpleMatch bonsai_1.0.1_windows_amd64.zip).Line -split ' ')[0]
gh attestation verify bonsai_1.0.1_windows_amd64.zip -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --source-ref refs/tags/v1.0.1 --deny-self-hosted-runners
Expand-Archive bonsai_1.0.1_windows_amd64.zip -DestinationPath . -Force
powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
```

The fourth line must print `True`; the fifth, `Verification succeeded` (PowerShell may show a stray character before
it). The last line brings one Windows question, "Do you want to allow this app to make changes to your device?":
choose Yes. It ends with the fingerprint, the `bonsai` programs Windows finds (`C:\Program Files\Bonsai\bonsai.exe`
first) and `bonsai 1.0.1 (commit ...)`. On an arm64 machine, write `arm64` for `amd64` in the archive's name.

## What each line proves

1. A new, empty folder of its own, so no file already there is unpacked or run.
2. The release's archive, downloaded from Bonsai's releases on GitHub over HTTPS (`-f`: a failed download stops
   there, with no file).
3. The release's list of fingerprints, `checksums.txt`, from the same release.
4. The archive is the one the release lists: its SHA-256 matches the line for it in `checksums.txt`. Alone, this
   proves the two files agree, not who made them.
5. The archive was built by Bonsai's release workflow (`.github/workflows/release.yml`) in its repository,
   `LastStep/Bonsai`, for this version's tag (`refs/tags/v1.0.1`), on GitHub's own machines (with
   `--deny-self-hosted-runners`, a record made on anyone's own machine fails). It runs before unpacking because the
   installer is inside the archive and runs with administrator rights or your password: nothing from the archive runs
   before its origin is checked.
6. The archive unpacked: the program and its installer side by side.
7. The installer: it puts the program in its one installed place, checking that the copy is the file it checked,
   writes Bonsai's record of the install, and lists every `bonsai` the PATH finds, so a second copy in front of it
   shows.

## When a line fails

Stop at the first line that does not do what this page says, and run nothing after it: send what it printed to
whoever looks after your setup (or the project's issues). Nothing from the archive has run before line 7.

- **Line 2 or 3 fails:** the release, or that file in it, is not there yet (a release waiting for approval, or one
  whose run failed), or the network failed. Try again later.
- **Line 4 prints anything else** (`FAILED`, or `False`): the archive does not match its release's list. Do not go on.
- **Line 5 fails:** an exit code of 1 means the archive's origin did not match (another repository or workflow, another
  tag, a machine of anyone's own), or it has no record at all: do not go on. An exit code of 4 means `gh` is not logged
  in: run `gh auth login`, then line 5 again. An error naming a flag means `gh` is older than 2.97: update it, then
  line 5 again.
- **Line 7 fails:** the installer says why and what to do next; do that, or send what it printed.

## The way out

**Linux and WSL.** While a project relies on Bonsai's guard, never remove the installed program
(`sh install.sh --remove`): with no `bonsai` there, every edit and shell command in that project's Claude Code
sessions is refused. Go back to the build you had before instead. Before you install, keep a copy of it and note the
fingerprint the second line prints:

```bash
cp /usr/local/bin/bonsai ~/bonsai-previous
sha256sum ~/bonsai-previous
```

To go back, put the copy in place; the second line must print the fingerprint you noted:

```bash
sudo install -o root -g root -m 0755 ~/bonsai-previous /usr/local/bin/bonsai
sha256sum /usr/local/bin/bonsai
```

After it, `bonsai check` reports that the installed program is not the one Bonsai's install record names: expected,
and cleared by the next install through the installer. Where no project relies on the guard, `sh install.sh --remove`,
from the folder the archive was unpacked in, takes the program and its record out.

**Windows.** If you chose No at the Windows question, run the last line again and choose Yes. To take Bonsai out, in
the same window (one more question):

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1 -Remove
```

If that fails too, in a PowerShell opened as administrator, then take `C:\Program Files\Bonsai` out of the system
`Path` (System Properties, Environment Variables):

```powershell
Remove-Item -Recurse -LiteralPath 'C:\Program Files\Bonsai'
```

## Checking an installed copy

Every program in a release is signed the same way, so you can check the installed copy at any time, not only while
installing. On Linux and WSL:

```bash
gh attestation verify /usr/local/bin/bonsai -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --deny-self-hosted-runners
```

On Windows:

```powershell
gh attestation verify 'C:\Program Files\Bonsai\bonsai.exe' -R LastStep/Bonsai --signer-workflow LastStep/Bonsai/.github/workflows/release.yml --deny-self-hosted-runners
```

Each prints `Verification succeeded` and exits 0 for a program Bonsai's release workflow built. `bonsai --version`
names the version and commit; `bonsai check` says when the installed program is not the one the install record names.

## A newer release

When a newer release exists, `bonsai check` warns of it (`bonsai-newer`) and `bonsai status --full --json` gives both
sides' seven lines for it, in `bonsai_release.lines`: these same lines, with that release's number.
