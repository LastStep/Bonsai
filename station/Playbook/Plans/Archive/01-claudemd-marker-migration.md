# Plan 01 — CLAUDE.md Marker Migration

**Tier:** 1 (Patch)
**Status:** Complete — shipped 2026-04-15
**Agent:** tech-lead (single domain: generator)

## Goal

When `WorkspaceClaudeMD()` encounters an existing CLAUDE.md without `<!-- BONSAI_START/END -->` markers, migrate the file (backup + overwrite with markers) instead of falling through to lock-aware write which conflicts and leaves nav tables stale.

## Context

The marker-based splice system in `WorkspaceClaudeMD()` works correctly when markers exist. The gap is the fallback path (lines 676-679) for marker-less files — it calls `writeFile()` which detects user modification → returns `ActionConflict` → user skips → nav tables go stale. This affects any workspace where the CLAUDE.md was customized and markers were removed (e.g., station/CLAUDE.md in this project).

## Steps

### 1. Modify `WorkspaceClaudeMD()` in `internal/generate/generate.go`

**Current code (lines 674-680):**
```go
	}

	// No markers or no existing file — full generation with markers, use lock-aware write
	fullContent := []byte(bonsaiStartMarker + "\n" + generatedContent + bonsaiEndMarker + "\n")
	r := writeFile(projectRoot, relPath, fullContent, "generated:workspace-claude-md", lock, force)
	result.Add(r)
	return nil
```

**Replace with:**
```go
	}

	// File exists but has no markers — migrate: backup old file, overwrite with markers
	fullContent := []byte(bonsaiStartMarker + "\n" + generatedContent + bonsaiEndMarker + "\n")

	if existing, readErr := os.ReadFile(absPath); readErr == nil {
		// Backup the user's customized file before overwriting
		_ = os.WriteFile(absPath+".bak", existing, 0644)
	}

	if err := os.WriteFile(absPath, fullContent, 0644); err != nil {
		return err
	}
	lock.Track(relPath, fullContent, "generated:workspace-claude-md")
	result.Add(FileResult{RelPath: relPath, Action: ActionUpdated, Source: "generated:workspace-claude-md"})
	return nil
```

Wait — we also need to handle the case where the file does NOT exist (first generation). The current code handles both "no markers" and "no file" in the same branch. After the change:

- If file exists with markers → splice (lines 654-672, unchanged)
- If file exists without markers → backup + overwrite with markers (new code)
- If file does not exist → create with markers via `writeFile()` (existing behavior)

**Revised replacement:**
```go
	} else if _, statErr := os.Stat(absPath); statErr == nil {
		// File exists but has no markers — migrate: backup + overwrite with markers
		if old, readErr := os.ReadFile(absPath); readErr == nil {
			_ = os.WriteFile(absPath+".bak", old, 0644)
		}
		fullContent := []byte(bonsaiStartMarker + "\n" + generatedContent + bonsaiEndMarker + "\n")
		if err := os.WriteFile(absPath, fullContent, 0644); err != nil {
			return err
		}
		lock.Track(relPath, fullContent, "generated:workspace-claude-md")
		result.Add(FileResult{RelPath: relPath, Action: ActionUpdated, Source: "generated:workspace-claude-md"})
		return nil
	}

	// No existing file — create with markers via lock-aware write
	fullContent := []byte(bonsaiStartMarker + "\n" + generatedContent + bonsaiEndMarker + "\n")
	r := writeFile(projectRoot, relPath, fullContent, "generated:workspace-claude-md", lock, force)
	result.Add(r)
	return nil
```

This keeps the existing `writeFile()` path only for truly new files (where it will `ActionCreated`).

### 2. Add test in `internal/generate/generate_test.go`

Add `TestClaudeMDMigratesMarkerlessFile`:
1. Create a CLAUDE.md without markers (write arbitrary content)
2. Call `WorkspaceClaudeMD()`
3. Assert: `.bak` file was created with old content
4. Assert: new file has markers
5. Assert: new file contains correct nav tables for installed items

### 3. Verification

- `make build` — compiles
- `go test ./...` — all tests pass including new test
- Manual smoke test:
  ```bash
  mkdir /tmp/bonsai-test && cd /tmp/bonsai-test && git init
  ./bonsai init   # creates CLAUDE.md with markers
  # Edit CLAUDE.md to remove markers
  ./bonsai add    # add an item
  # Verify: CLAUDE.md has markers, .bak exists, nav tables are current
  ```

## Security

> Refer to `Playbook/Standards/SecurityStandards.md` — no security-sensitive changes in this patch.

## Verification

- [ ] `make build` passes
- [ ] `go test ./...` passes (including new marker migration test)
- [ ] Existing tests `TestClaudeMDHasMarkers` and `TestClaudeMDPreservesUserContent` still pass
- [ ] CLI smoke test: marker-less CLAUDE.md gets migrated with backup on `bonsai add`/`bonsai update`
