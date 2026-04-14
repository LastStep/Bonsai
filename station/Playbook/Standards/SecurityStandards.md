---
tags: [standards, security]
description: Hard security rules for all agents and application code.
---

# Bonsai — Security Standards

> [!warning]
> **Rule Zero — Directory Isolation**
> Every agent operates ONLY in its designated directory. No exceptions.
> Reading files outside your directory for context is permitted. Creating, editing, or deleting files outside your directory is NEVER permitted.
> If a task seems to require crossing a directory boundary, STOP and ask the user.

---

## Domain 1 — Secrets Management

- NEVER commit `.env`, credentials, API keys, or tokens to git
- NEVER log, print, or echo secrets — not even in debug mode
- NEVER hardcode secrets in source files — always use environment variables
- NEVER include real credentials in code examples, comments, or documentation
- If you encounter a secret in code, flag it immediately and STOP

---

## Domain 2 — Destructive Operations

- NEVER run `rm -rf`, `git clean -f`, `git reset --hard`, or `git checkout .` without explicit user approval
- NEVER force-push without explicit user approval
- NEVER delete git branches without explicit user approval
- If a destructive action seems like the only path forward, STOP and ask the user

---

## Domain 3 — Scope Boundaries

- Each agent operates ONLY in its designated directory
- No agent modifies `CLAUDE.md`, `.env`, or infrastructure config without explicit user approval
- No agent installs new dependencies without explicit user approval — propose it, don't do it
- No agent makes architectural decisions — encounter a design fork, STOP and flag it

---

## Domain 4 — User File Safety

Bonsai generates files in user projects. Protect user work:

- NEVER silently overwrite a file the user has modified — use the lock file conflict system
- ALWAYS track generated files in `.bonsai-lock.yaml` with content hashes
- ALWAYS offer skip/overwrite/backup options when conflicts are detected
- Scaffolding files are write-once — if the file exists, skip it regardless of content
- NEVER delete user files during `bonsai remove` without explicit confirmation

---

## Domain 5 — Template Safety

Catalog templates render into user projects:

- NEVER include executable code in `.md` templates — templates produce documentation, not scripts
- Sensor `.sh.tmpl` files ARE executable — ensure they exit cleanly on all paths, never hang
- Validate all template variables exist before rendering — missing vars should error, not produce empty output
- NEVER use `{{ .Variable }}` in a way that could inject shell commands when rendered into `.sh` files

---

## Domain 6 — Dependencies

- Pin dependency versions in `go.mod` — no floating ranges
- Review changelogs before upgrading major versions
- Never install packages from untrusted sources
- Keep dependencies up to date — run `go vet` and check for known vulnerabilities
- The Charm ecosystem (BubbleTea, Huh, LipGloss) is the approved TUI stack — do not introduce alternative TUI libraries

---

## Domain 7 — Embedded Catalog Integrity

- The catalog is embedded via `embed.FS` — it ships inside the binary
- NEVER allow runtime modification of embedded catalog files
- All catalog changes go through the source tree and a rebuild
- Catalog items must be self-contained — no external downloads or network calls during generation
