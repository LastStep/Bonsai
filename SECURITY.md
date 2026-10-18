# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Bonsai, please report it through [GitHub's private vulnerability reporting](https://github.com/LastStep/Bonsai/security/advisories/new).

**Do not open a public issue for security vulnerabilities.**

We will acknowledge your report within 48 hours and provide an initial assessment of the issue.

## Scope

The following are in scope for security reports on Bonsai 1.x:

- **The guard and the hook path** - the `bonsai hook` lines Bonsai writes into `.claude/settings.json`, and the guard,
  the stop gate and the recorder behind them: a way to make the guard allow what it should refuse, to run another
  program in its place, or to leave it unable to refuse (it must fail closed)
- **The installers** - `install/install.sh` and `install/install.ps1`, and the places they install to
  (`/usr/local/bin/bonsai` and `C:\Program Files\Bonsai\bonsai.exe`)
- **The release and its files** - the release workflow, `.goreleaser.yaml`, and the archives, checksums and signed
  records a release publishes
- **The program itself** - path traversal, unsafe file operations, or command injection in what `bonsai` reads and
  writes in a project, such as `bonsai.yaml`, the lock and a pack's files

## Out of Scope

- **User-customized files** - files you modify after Bonsai writes them are your responsibility
- **Third-party packs** - a pack is code you choose to install; report a problem in one to its own repository
- **Third-party dependencies** - report these to the upstream maintainer directly
- **Claude Code itself** - report issues with Claude Code at [anthropics/claude-code](https://github.com/anthropics/claude-code/issues)

## Supported Versions

The newest 1.x release is supported. Older 1.x releases are not: update to the newest.

0.4.3 is the old product's last release. It is not supported and gets no further releases, security fixes included.
