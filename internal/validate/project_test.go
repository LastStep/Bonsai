package validate

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
)

// --- Project-level pass fixtures (Plan 40 Phase 2) -------------------------

// validManifest is the canonical good manifest body. Tests mutate one field
// at a time to build a negative control per rule.
const validManifest = `schema_version: 1
name: Demo Project
slug: demo
status: active
tags: []
description: a demo
links: {}
created: 2026-06-13
memory_dir: station/Memory
`

// validNote returns a well-formed note body for the given permalink, scope,
// and an optional relations/superseded tail. The frontmatter matches the
// frozen v1 schema.
func validNote(permalink, scope, supersededBy, relations string) string {
	sb := "null"
	if supersededBy != "" {
		sb = supersededBy
	}
	body := "---\n" +
		"schema_version: 1\n" +
		"title: A Note\n" +
		"type: decision\n" +
		"permalink: " + permalink + "\n" +
		"tags: []\n" +
		"scope: " + scope + "\n" +
		"valid_from: 2026-06-13\n" +
		"superseded_by: " + sb + "\n" +
		"---\n" +
		"## Observations\n- [decision] chose X #arch\n" +
		"## Relations\n"
	if relations != "" {
		body += relations + "\n"
	}
	return body
}

// writeProjectFile materialises a file under root, creating parent dirs.
func writeProjectFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", abs, err)
	}
}

// seedMemoryDirs creates the empty Memory/{decisions,notes} tree so a
// manifest-present fixture has a tree to walk even before notes are added.
func seedMemoryDirs(t *testing.T, root string) {
	t.Helper()
	for _, sub := range []string{"station/Memory/decisions", "station/Memory/notes"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(sub)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}
}

// runProject runs the full Run() against a fixture whose manifest + memory
// tree the caller has already written, and returns the report.
func runProject(t *testing.T, f *projectFixture) *Report {
	t.Helper()
	report, err := Run(f.root, f.cfg, nil, f.lock, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return report
}

// findProjectIssue returns the first issue matching category (and optional
// name) whose AgentName is empty — project-level issues are always unscoped.
func findProjectIssue(issues []Issue, cat Category, name string) *Issue {
	for i := range issues {
		if issues[i].Category == cat && issues[i].AgentName == "" && (name == "" || issues[i].Name == name) {
			return &issues[i]
		}
	}
	return nil
}

// TestProject_ValidTree is the zero-issues control: a good manifest + a small
// valid memory graph (including a resolved relation + a resolved
// superseded_by) must produce ZERO issues.
func TestProject_ValidTree(t *testing.T) {
	f := newFixture(t)
	writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
	seedMemoryDirs(t, f.root)
	// note-a supersedes note-b and relates to note-b; both exist → resolved.
	writeProjectFile(t, f.root, "station/Memory/decisions/note-a.md",
		validNote("note-a", "project/demo", "note-b", "- relates_to [[note-b]]"))
	writeProjectFile(t, f.root, "station/Memory/notes/note-b.md",
		validNote("note-b", "project/demo", "", ""))

	report := runProject(t, f)
	if report.HasIssues() {
		t.Fatalf("expected zero issues for valid tree, got: %+v", report.Issues)
	}
	// Project pass must NOT pollute AgentsScanned.
	if len(report.AgentsScanned) != 1 || report.AgentsScanned[0] != "tech-lead" {
		t.Fatalf("AgentsScanned = %v, want [tech-lead]", report.AgentsScanned)
	}
}

// TestProject_RuleTable is the fixture↔rule↔(category,severity) table. Each
// case mutates exactly one thing on top of the valid baseline and asserts the
// expected category + severity fires. setup writes the manifest + tree.
func TestProject_RuleTable(t *testing.T) {
	cases := []struct {
		name     string
		setup    func(t *testing.T, f *projectFixture)
		cat      Category
		severity Severity
		// issueName, when non-empty, is asserted on the matched issue.
		issueName string
	}{
		{
			name: "manifest schema_version 2",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml",
					strings.Replace(validManifest, "schema_version: 1", "schema_version: 2", 1))
			},
			cat:      CategoryInvalidManifest,
			severity: SeverityError,
		},
		{
			name: "manifest status bogus",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml",
					strings.Replace(validManifest, "status: active", "status: bogus", 1))
			},
			cat:      CategoryInvalidManifest,
			severity: SeverityError,
		},
		{
			name: "manifest memory_dir traversal",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml",
					strings.Replace(validManifest, "memory_dir: station/Memory", "memory_dir: ../escape", 1))
			},
			cat:      CategoryInvalidManifest,
			severity: SeverityError,
		},
		{
			name: "manifest missing required slug",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml",
					strings.Replace(validManifest, "slug: demo\n", "", 1))
			},
			cat:      CategoryInvalidManifest,
			severity: SeverityError,
		},
		{
			name: "note missing required frontmatter (no permalink/scope/...)",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				writeProjectFile(t, f.root, "station/Memory/notes/broken.md",
					"---\nschema_version: 1\ntitle: x\n---\nbody\n")
			},
			cat:      CategoryInvalidNote,
			severity: SeverityError,
		},
		{
			name: "note bad schema_version",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				n := strings.Replace(validNote("nbad", "project/demo", "", ""), "schema_version: 1", "schema_version: 2", 1)
				writeProjectFile(t, f.root, "station/Memory/notes/nbad.md", n)
			},
			cat:       CategoryInvalidNote,
			severity:  SeverityError,
			issueName: "nbad",
		},
		{
			name: "note out-of-charset permalink",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				writeProjectFile(t, f.root, "station/Memory/notes/badlink.md",
					validNote("Bad_Link!", "project/demo", "", ""))
			},
			cat:      CategoryInvalidNote,
			severity: SeverityError,
		},
		{
			name: "note bad type",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				n := strings.Replace(validNote("ntype", "project/demo", "", ""), "type: decision", "type: musing", 1)
				writeProjectFile(t, f.root, "station/Memory/notes/ntype.md", n)
			},
			cat:       CategoryInvalidNote,
			severity:  SeverityError,
			issueName: "ntype",
		},
		{
			name: "note bad scope (slug mismatch)",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				writeProjectFile(t, f.root, "station/Memory/notes/nscope.md",
					validNote("nscope", "project/other", "", ""))
			},
			cat:       CategoryInvalidNote,
			severity:  SeverityError,
			issueName: "nscope",
		},
		{
			name: "note dangling superseded_by",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				writeProjectFile(t, f.root, "station/Memory/decisions/ndang.md",
					validNote("ndang", "project/demo", "does-not-exist", ""))
			},
			cat:       CategoryInvalidNote,
			severity:  SeverityError,
			issueName: "ndang",
		},
		{
			name: "note unresolved relation (warning)",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				writeProjectFile(t, f.root, "station/Memory/notes/nrel.md",
					validNote("nrel", "project/demo", "", "- relates_to [[ghost-note]]"))
			},
			cat:       CategoryUnresolvedRelation,
			severity:  SeverityWarning,
			issueName: "nrel",
		},
		{
			name: "MEMORY.md over budget",
			setup: func(t *testing.T, f *projectFixture) {
				writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
				seedMemoryDirs(t, f.root)
				writeProjectFile(t, f.root, "station/MEMORY.md", strings.Repeat("line\n", 250))
			},
			cat:      CategoryMemoryIndexTooLarge,
			severity: SeverityWarning,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			tc.setup(t, f)
			report := runProject(t, f)
			iss := findProjectIssue(report.Issues, tc.cat, tc.issueName)
			if iss == nil {
				t.Fatalf("expected %s issue (name=%q), got: %+v", tc.cat, tc.issueName, report.Issues)
			}
			if iss.Severity != tc.severity {
				t.Errorf("severity = %s, want %s", iss.Severity, tc.severity)
			}
			if iss.AgentName != "" {
				t.Errorf("project issue carried AgentName %q, want empty", iss.AgentName)
			}
		})
	}
}

// TestProject_ManifestAbsentButMemoryPresent asserts the warning fires AND
// frontmatter errors still surface (scope-match is skipped, everything else
// is still linted).
func TestProject_ManifestAbsentButMemoryPresent(t *testing.T) {
	f := newFixture(t)
	seedMemoryDirs(t, f.root)
	// A note with a bad schema_version — a non-scope error must still surface.
	bad := strings.Replace(validNote("nabsent", "project/whatever", "", ""), "schema_version: 1", "schema_version: 2", 1)
	writeProjectFile(t, f.root, "station/Memory/notes/nabsent.md", bad)

	report := runProject(t, f)

	if findProjectIssue(report.Issues, CategoryMissingManifest, "") == nil {
		t.Fatalf("expected missing_manifest warning, got: %+v", report.Issues)
	}
	if iss := findProjectIssue(report.Issues, CategoryMissingManifest, ""); iss != nil && iss.Severity != SeverityWarning {
		t.Errorf("missing_manifest severity = %s, want warning", iss.Severity)
	}
	// Frontmatter still linted: the bad schema_version must surface as an error.
	if iss := findProjectIssue(report.Issues, CategoryInvalidNote, "nabsent"); iss == nil {
		t.Fatalf("expected invalid_note for nabsent even with manifest absent, got: %+v", report.Issues)
	} else if iss.Severity != SeverityError {
		t.Errorf("invalid_note severity = %s, want error", iss.Severity)
	}
	// Scope-match check is skipped when slug is unknown — a project/whatever
	// scope must NOT raise a scope-mismatch error.
	for _, iss := range report.Issues {
		if iss.Category == CategoryInvalidNote && strings.Contains(iss.Detail, "does not match manifest slug") {
			t.Fatalf("scope-match check should be skipped when manifest absent, got: %+v", iss)
		}
	}
}

// TestProject_ManifestAbsentNoMemory verifies the common case: no manifest and
// no memory tree → no project-level issues at all (the warning only fires when
// a tree actually exists).
func TestProject_ManifestAbsentNoMemory(t *testing.T) {
	f := newFixture(t)
	report := runProject(t, f)
	if findProjectIssue(report.Issues, CategoryMissingManifest, "") != nil {
		t.Fatalf("missing_manifest should not fire without a memory tree, got: %+v", report.Issues)
	}
}

// TestProject_RunsRegardlessOfAgentFilter confirms the project-level pass
// fires even when an --agent filter is applied (non-error path). The filtered
// agent has no issues, but the manifest does.
func TestProject_RunsRegardlessOfAgentFilter(t *testing.T) {
	f := newFixture(t)
	writeProjectFile(t, f.root, ".bonsai/project.yaml",
		strings.Replace(validManifest, "schema_version: 1", "schema_version: 2", 1))

	report, err := Run(f.root, f.cfg, nil, f.lock, "tech-lead")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if findProjectIssue(report.Issues, CategoryInvalidManifest, "") == nil {
		t.Fatalf("project pass must run under --agent filter, got: %+v", report.Issues)
	}
}

// TestProject_UnknownAgentFilterStillErrorsEarly confirms the project pass does
// NOT run when the agent filter is unknown — Run errors out before reaching it.
func TestProject_UnknownAgentFilterStillErrorsEarly(t *testing.T) {
	f := newFixture(t)
	writeProjectFile(t, f.root, ".bonsai/project.yaml",
		strings.Replace(validManifest, "schema_version: 1", "schema_version: 2", 1))
	if _, err := Run(f.root, f.cfg, nil, f.lock, "ghost"); err == nil {
		t.Fatalf("expected early error for unknown agent filter")
	}
}

// TestProject_SymlinkEscapeRejected verifies an adversarial symlinked note
// whose target escapes memory_dir is refused (symlink_escape error) and never
// read. POSIX-only — Windows symlink creation requires elevated privileges.
func TestProject_SymlinkEscapeRejected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	f := newFixture(t)
	writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
	seedMemoryDirs(t, f.root)

	// A secret file OUTSIDE the memory tree.
	outside := filepath.Join(f.root, "secret.md")
	if err := os.WriteFile(outside, []byte("top secret\n"), 0o644); err != nil {
		t.Fatalf("write outside: %v", err)
	}
	// A symlink inside the tree pointing at it.
	link := filepath.Join(f.root, "station/Memory/notes/leak.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	report := runProject(t, f)
	iss := findProjectIssue(report.Issues, CategorySymlinkEscape, "")
	if iss == nil {
		t.Fatalf("expected symlink_escape, got: %+v", report.Issues)
	}
	if iss.Severity != SeverityError {
		t.Errorf("severity = %s, want error", iss.Severity)
	}
}

// TestProject_SupersededByAbsentIsNotSuperseded verifies the absent ≡ null ≡
// not-superseded rule: a note with no superseded_by key produces no
// missing-key error and no dangling-resolution error.
func TestProject_SupersededByAbsentIsNotSuperseded(t *testing.T) {
	f := newFixture(t)
	writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
	seedMemoryDirs(t, f.root)
	// Note body with NO superseded_by line at all.
	body := "---\nschema_version: 1\ntitle: A\ntype: note\npermalink: nokey\ntags: []\nscope: project/demo\n---\n## Observations\n- [note] x\n"
	writeProjectFile(t, f.root, "station/Memory/notes/nokey.md", body)

	report := runProject(t, f)
	if report.HasIssues() {
		t.Fatalf("absent superseded_by must not raise issues, got: %+v", report.Issues)
	}
}

// TestProject_PermalinkOutOfCharsetNotIndexed verifies an out-of-charset
// permalink note is errored AND not indexed — so it cannot satisfy another
// note's relation (which would then warn as unresolved).
func TestProject_PermalinkOutOfCharsetNotIndexed(t *testing.T) {
	f := newFixture(t)
	writeProjectFile(t, f.root, ".bonsai/project.yaml", validManifest)
	seedMemoryDirs(t, f.root)
	// Bad-permalink note + a second note relating to its sanitized form.
	writeProjectFile(t, f.root, "station/Memory/notes/bad.md", validNote("Bad!", "project/demo", "", ""))
	writeProjectFile(t, f.root, "station/Memory/notes/ref.md", validNote("ref", "project/demo", "", "- relates_to [[bad]]"))

	report := runProject(t, f)
	// The bad permalink errors.
	if findProjectIssue(report.Issues, CategoryInvalidNote, "") == nil {
		t.Fatalf("expected invalid_note for out-of-charset permalink, got: %+v", report.Issues)
	}
	// And ref's relation to "bad" stays unresolved (not indexed).
	if findProjectIssue(report.Issues, CategoryUnresolvedRelation, "ref") == nil {
		t.Fatalf("out-of-charset permalink must not be indexed; ref's relation should be unresolved, got: %+v", report.Issues)
	}
}

// ensure config import is used even if a future refactor drops the direct
// reference — keeps the test file self-documenting about its dependency.
var _ = config.NewLockFile
