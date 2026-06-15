package nonint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// initProject runs RunInit against a fresh tmp dir with the tech-lead-only
// fixture and returns the cwd. Shared setup for the update-core tests — every
// RunUpdate test needs a materialised project (workspace + .bonsai.yaml + lock)
// on disk first.
func initProject(t *testing.T) string {
	t.Helper()
	cat := loadTestCatalog(t)
	tmp := t.TempDir()
	cfg := minimalInitCfg(t, tmp)
	if _, code, err := RunInit(tmp, filepath.Join(tmp, ".bonsai.yaml"), cfg, cat, "test"); err != nil || code != ExitOK {
		t.Fatalf("RunInit (setup): code=%d err=%v", code, err)
	}
	return tmp
}

// techLeadSkillsDir returns the path to the tech-lead's agent/Skills directory
// inside a project initialised by initProject (DocsPath defaults to station/).
func techLeadSkillsDir(t *testing.T, cwd string) string {
	t.Helper()
	cfg, err := config.Load(filepath.Join(cwd, ".bonsai.yaml"))
	if err != nil {
		t.Fatalf("load .bonsai.yaml: %v", err)
	}
	tl := cfg.Agents["tech-lead"]
	if tl == nil {
		t.Fatalf("tech-lead missing from .bonsai.yaml")
	}
	return filepath.Join(cwd, tl.Workspace, "agent", "Skills")
}

// reloadForUpdate reloads the project config + lock the way the cobra entry
// point does before calling RunUpdate.
func reloadForUpdate(t *testing.T, cwd string) (*config.ProjectConfig, *config.LockFile) {
	t.Helper()
	cfg, err := config.Load(filepath.Join(cwd, ".bonsai.yaml"))
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	lock, _ := config.LoadLockFile(cwd)
	if lock == nil {
		lock = config.NewLockFile()
	}
	return cfg, lock
}

const validSkillBody = `---
description: A user-authored custom skill
display_name: My Custom Skill
---

# My Custom Skill

Body.
`

// TestRunUpdate_PositivePath proves the RunStatic→RunUpdate lift preserved the
// non-conflict behaviour: a clean valid discovery is auto-applied and tracked
// (exit 0, a Write.Files entry created/updated, no warnings); an invalid
// (bad-frontmatter) discovery is surfaced in Result.Warnings (NOT stdout) and
// does not fail the run. Plan 41 Verification — update positive path.
func TestRunUpdate_PositivePath(t *testing.T) {
	cat := loadTestCatalog(t)
	cwd := initProject(t)
	skillsDir := techLeadSkillsDir(t, cwd)

	// Valid custom skill — clean frontmatter, untracked → discovered as valid.
	validRel := filepath.Join("station", "agent", "Skills", "my-custom.md")
	if err := os.WriteFile(filepath.Join(skillsDir, "my-custom.md"), []byte(validSkillBody), 0o644); err != nil {
		t.Fatalf("write valid custom skill: %v", err)
	}
	// Invalid custom skill — no frontmatter → discovered with Error set.
	if err := os.WriteFile(filepath.Join(skillsDir, "broken.md"), []byte("# No frontmatter\n"), 0o644); err != nil {
		t.Fatalf("write invalid custom skill: %v", err)
	}

	cfg, lock := reloadForUpdate(t, cwd)
	result, code, err := RunUpdate(cwd, cfg, cat, lock, "test", false)
	if err != nil {
		t.Fatalf("RunUpdate: %v", err)
	}
	if code != ExitOK {
		t.Fatalf("exit code: want %d, got %d", ExitOK, code)
	}

	// Applying the valid discovery is observable three ways: (a) the skill is
	// registered in the persisted config, (b) the custom file is tracked in the
	// lock, and (c) the resulting workspace re-render produces at least one
	// created/updated entry in Write.Files (the nav table in CLAUDE.md grows to
	// list the new skill). The custom file itself is user-authored on disk so
	// the generator never re-writes it as its own Write entry — tracking is the
	// signal, the CLAUDE.md update is its visible effect.
	post, _ := config.Load(filepath.Join(cwd, ".bonsai.yaml"))
	if !containsString(post.Agents["tech-lead"].Skills, "my-custom") {
		t.Errorf("valid skill not registered in config; skills=%v", post.Agents["tech-lead"].Skills)
	}
	postLock, _ := config.LoadLockFile(cwd)
	if postLock == nil || postLock.Files[validRel] == nil {
		t.Errorf("valid discovery %q not tracked in lock after update", validRel)
	}
	var sawApplied bool
	for _, f := range result.Write.Files {
		if f.Action == generate.ActionCreated || f.Action == generate.ActionUpdated {
			sawApplied = true
		}
	}
	if !sawApplied {
		t.Errorf("applying the discovery produced no created/updated Write entry; got %+v", result.Write.Files)
	}

	// Invalid discovery surfaced in Warnings (not stdout/Write).
	var sawInvalidWarning bool
	for _, w := range result.Warnings {
		if strings.Contains(w, "broken.md") {
			sawInvalidWarning = true
		}
	}
	if !sawInvalidWarning {
		t.Errorf("invalid discovery not surfaced in Warnings; warnings=%v", result.Warnings)
	}

	// And the invalid file must NOT appear on the JSONL stream.
	var buf bytes.Buffer
	if err := EmitJSONL(&buf, result); err != nil {
		t.Fatalf("EmitJSONL: %v", err)
	}
	if strings.Contains(buf.String(), "broken.md") {
		t.Errorf("invalid discovery leaked onto JSONL stream:\n%s", buf.String())
	}
}

// TestRunUpdate_CleanResync proves a fresh post-init project with no custom
// files and no user edits resyncs to exit 0 with no conflicts and empty
// warnings (the everything-unchanged steady state).
func TestRunUpdate_CleanResync(t *testing.T) {
	cat := loadTestCatalog(t)
	cwd := initProject(t)
	cfg, lock := reloadForUpdate(t, cwd)

	result, code, err := RunUpdate(cwd, cfg, cat, lock, "test", false)
	if err != nil {
		t.Fatalf("RunUpdate: %v", err)
	}
	if code != ExitOK {
		t.Fatalf("exit code: want %d, got %d", ExitOK, code)
	}
	if result.Write.HasConflicts() {
		t.Errorf("clean resync must have no conflicts; got %+v", result.Write.Conflicts())
	}
	if len(result.Warnings) != 0 {
		t.Errorf("clean resync must have no warnings; got %v", result.Warnings)
	}
}

// TestRunUpdate_ConflictExit5 is the exit-5 negative control at the core level
// (Plan 41 Verification C2): a user-edited generator-target file with no lock
// entry surfaces as ActionConflict. Without --skip-conflicts the run exits
// ExitConflict (5) and Write.Files STILL lists the conflict entry; with
// --skip-conflicts it exits 0 and the file is counted skipped (left untouched).
func TestRunUpdate_ConflictExit5(t *testing.T) {
	cat := loadTestCatalog(t)
	cwd := initProject(t)

	// Overwrite a stable generator output with user bytes and NO lock update —
	// IsModified treats an untracked-but-existing file as user-owned, so the
	// next write surfaces ActionConflict. Mirrors TestRunInit_ConflictEmittedNotForced.
	userBody := []byte("# USER EDIT — must not be overwritten\n")
	relPath := "station/agent/Core/identity.md"
	target := filepath.Join(cwd, relPath)
	if err := os.WriteFile(target, userBody, 0o644); err != nil {
		t.Fatalf("seed conflict: %v", err)
	}
	// Drop the lock entry for this path so it reads as user-owned.
	dropLockEntry(t, cwd, relPath)

	// skipConflicts=false → exit 5, conflict entry present.
	cfg, lock := reloadForUpdate(t, cwd)
	result, code, err := RunUpdate(cwd, cfg, cat, lock, "test", false)
	if err != nil {
		t.Fatalf("RunUpdate (no skip): %v", err)
	}
	if code != ExitConflict {
		t.Fatalf("exit code: want %d (ExitConflict), got %d", ExitConflict, code)
	}
	var foundConflict bool
	for _, f := range result.Write.Files {
		if f.RelPath == relPath && f.Action == generate.ActionConflict {
			foundConflict = true
		}
	}
	if !foundConflict {
		t.Errorf("expected Action==conflict entry for %q in Write.Files; got %+v", relPath, result.Write.Files)
	}
	// File untouched.
	if got, _ := os.ReadFile(target); !bytes.Equal(got, userBody) {
		t.Errorf("conflict file overwritten; want %q got %q", userBody, got)
	}

	// skipConflicts=true → exit 0, file counted as skipped (still untouched).
	cfg2, lock2 := reloadForUpdate(t, cwd)
	result2, code2, err := RunUpdate(cwd, cfg2, cat, lock2, "test", true)
	if err != nil {
		t.Fatalf("RunUpdate (skip): %v", err)
	}
	if code2 != ExitOK {
		t.Fatalf("exit code (skip): want %d, got %d", ExitOK, code2)
	}
	_, _, _, skipped, conflicts := result2.Counts()
	if skipped == 0 {
		t.Errorf("skipConflicts must count the conflict as skipped; counts skipped=%d conflicts=%d files=%+v", skipped, conflicts, result2.Write.Files)
	}
	if got, _ := os.ReadFile(target); !bytes.Equal(got, userBody) {
		t.Errorf("conflict file overwritten under skip; want %q got %q", userBody, got)
	}
}

// containsString reports whether s appears in slice.
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// dropLockEntry removes the lock entry for relPath and re-saves the lock so the
// file reads as untracked (user-owned) on the next IsModified check.
func dropLockEntry(t *testing.T, cwd, relPath string) {
	t.Helper()
	lock, _ := config.LoadLockFile(cwd)
	if lock == nil {
		return
	}
	delete(lock.Files, relPath)
	if err := lock.Save(cwd); err != nil {
		t.Fatalf("re-save lock: %v", err)
	}
}
