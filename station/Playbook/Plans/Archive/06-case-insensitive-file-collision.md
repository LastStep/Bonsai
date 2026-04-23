# Plan 06 — Fix Case-Insensitive File Collision

**Tier:** 1
**Status:** Complete
**Agent:** tech-lead (station-only files)

## Goal

Rename `station/index.md` to `station/code-index.md` to eliminate the case-insensitive collision with `station/INDEX.md` that blocks `go install` and pkg.go.dev indexing.

## Steps

1. `git mv station/index.md station/code-index.md`

2. Update `CLAUDE.md` (root), line 57 — change project structure tree entry:
   - Old: `├── index.md              ← code index — quick-nav to Go source`
   - New: `├── code-index.md         ← code index — quick-nav to Go source`

3. Update `station/INDEX.md`, line 52 — change document map row:
   - Old: `` | `index.md` | Code index — quick-nav to Go source functions | When navigating the codebase | ``
   - New: `` | `code-index.md` | Code index — quick-nav to Go source functions | When navigating the codebase | ``

4. Verify: `make build && go test ./...`

5. After merge: re-trigger Go module proxy indexing and verify `go install github.com/LastStep/Bonsai@latest` works.

## Security

> Refer to SecurityStandards.md — no security implications for this rename.

## Verification

- [ ] `station/index.md` no longer exists
- [ ] `station/code-index.md` exists with identical content
- [ ] `grep -ri 'index\.md' CLAUDE.md station/INDEX.md` returns zero hits for the old lowercase `index.md`
- [ ] `make build` passes
- [ ] `go test ./...` passes
