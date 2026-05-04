package validate

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
)

// projectFixture is the canonical test scaffold — it materialises a temp
// directory with a tech-lead workspace tree (agent/Skills/Workflows/...)
// and returns the project root + a minimally populated InstalledAgent.
// Each subtest layers its own failure mode on top via direct file writes
// and config.InstalledAgent edits.
type projectFixture struct {
	root    string
	cfg     *config.ProjectConfig
	lock    *config.LockFile
	tlAgent *config.InstalledAgent
}

func newFixture(t *testing.T) *projectFixture {
	t.Helper()
	root := t.TempDir()
	for _, sub := range []string{"agent/Skills", "agent/Workflows", "agent/Protocols", "agent/Sensors", "agent/Routines"} {
		if err := os.MkdirAll(filepath.Join(root, "station", sub), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
	}
	tl := &config.InstalledAgent{
		AgentType:   "tech-lead",
		Workspace:   "station",
		CustomItems: map[string]*config.CustomItemMeta{},
	}
	cfg := &config.ProjectConfig{
		ProjectName: "demo",
		Agents:      map[string]*config.InstalledAgent{"tech-lead": tl},
	}
	return &projectFixture{
		root:    root,
		cfg:     cfg,
		lock:    config.NewLockFile(),
		tlAgent: tl,
	}
}

// writeFM is a helper that materialises a custom ability file with a
// minimal valid frontmatter block. itemType selects between markdown
// (.md) and bash-comment (.sh) frontmatter styles.
func writeFM(t *testing.T, root, rel, itemType, description string, extraFields map[string]string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
	}
	var body string
	if itemType == "sensor" {
		body = "#!/usr/bin/env bash\n# ---\n# description: " + description + "\n"
		for k, v := range extraFields {
			body += "# " + k + ": " + v + "\n"
		}
		body += "# ---\necho hi\n"
	} else {
		body = "---\ndescription: " + description + "\n"
		for k, v := range extraFields {
			body += k + ": " + v + "\n"
		}
		body += "---\nbody\n"
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", abs, err)
	}
}

// findIssue returns the first issue matching category + name, or nil if
// none — failures are reported via the calling test to keep the helper
// noise-free.
func findIssue(issues []Issue, cat Category, name string) *Issue {
	for i := range issues {
		if issues[i].Category == cat && issues[i].Name == name {
			return &issues[i]
		}
	}
	return nil
}

// TestRun_CleanProject checks the empty-project happy path. A project
// with one agent and zero abilities should produce zero issues.
func TestRun_CleanProject(t *testing.T) {
	f := newFixture(t)
	report, err := Run(f.root, f.cfg, nil, f.lock, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.HasIssues() {
		t.Fatalf("expected clean project, got issues: %+v", report.Issues)
	}
	if got := report.AgentsScanned; len(got) != 1 || got[0] != "tech-lead" {
		t.Fatalf("AgentsScanned = %v, want [tech-lead]", got)
	}
}

// TestRun_OrphanedRegistration verifies the Plan 34 repro pattern:
// installed.Skills lists "foo", file exists with valid frontmatter, but
// neither the lock nor custom_items[foo] is populated.
func TestRun_OrphanedRegistration(t *testing.T) {
	f := newFixture(t)
	f.tlAgent.Skills = []string{"foo"}
	writeFM(t, f.root, "station/agent/Skills/foo.md", "skill", "foo skill", nil)
	// Note: NOT calling lock.Track and NOT setting custom_items[foo].

	report, err := Run(f.root, f.cfg, nil, f.lock, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	iss := findIssue(report.Issues, CategoryOrphanedRegistration, "foo")
	if iss == nil {
		t.Fatalf("expected orphaned_registration for foo, got: %+v", report.Issues)
	}
	if iss.Severity != SeverityError {
		t.Errorf("severity = %s, want error", iss.Severity)
	}
	if iss.AbilityType != "skill" {
		t.Errorf("ability_type = %s, want skill", iss.AbilityType)
	}
	if iss.AgentName != "tech-lead" {
		t.Errorf("agent = %s, want tech-lead", iss.AgentName)
	}
	if iss.Path != "station/agent/Skills/foo.md" {
		t.Errorf("path = %s", iss.Path)
	}
}

// TestRun_OrphanedRegistration_LockButNoCustomItems covers the variant
// where the lock entry exists (so file is "tracked") but custom_items
// metadata is missing or empty — also an orphan, just at a different
// stage of breakage.
func TestRun_OrphanedRegistration_LockButNoCustomItems(t *testing.T) {
	f := newFixture(t)
	f.tlAgent.Skills = []string{"bar"}
	writeFM(t, f.root, "station/agent/Skills/bar.md", "skill", "bar skill", nil)
	f.lock.Track("station/agent/Skills/bar.md", []byte("dummy"), "custom:skills/bar")
	// custom_items left empty — should be flagged as orphan.

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	iss := findIssue(report.Issues, CategoryOrphanedRegistration, "bar")
	if iss == nil {
		t.Fatalf("expected orphaned_registration for bar, got: %+v", report.Issues)
	}
}

// TestRun_MissingFile verifies the case where installed.<Cat> has a
// name but the corresponding file is absent on disk.
func TestRun_MissingFile(t *testing.T) {
	f := newFixture(t)
	f.tlAgent.Skills = []string{"ghost"}
	// no file written

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	iss := findIssue(report.Issues, CategoryMissingFile, "ghost")
	if iss == nil {
		t.Fatalf("expected missing_file for ghost, got: %+v", report.Issues)
	}
	if iss.Severity != SeverityError {
		t.Errorf("severity = %s, want error", iss.Severity)
	}
	if iss.AbilityType != "skill" {
		t.Errorf("ability_type = %s, want skill", iss.AbilityType)
	}
}

// TestRun_StaleLockEntry verifies a custom: lock entry whose file was
// deleted produces a stale_lock_entry warning.
func TestRun_StaleLockEntry(t *testing.T) {
	f := newFixture(t)
	// File never written, but lock thinks it tracked one.
	f.lock.Track("station/agent/Skills/zombie.md", []byte("dummy"), "custom:skills/zombie")

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	iss := findIssue(report.Issues, CategoryStaleLockEntry, "zombie")
	if iss == nil {
		t.Fatalf("expected stale_lock_entry for zombie, got: %+v", report.Issues)
	}
	if iss.Severity != SeverityWarning {
		t.Errorf("severity = %s, want warning", iss.Severity)
	}
	if iss.AbilityType != "skill" {
		t.Errorf("ability_type = %s, want skill", iss.AbilityType)
	}
}

// TestRun_UntrackedCustomFile verifies a valid-frontmatter file dropped
// in agent/Skills/ but not yet registered emits a warning.
func TestRun_UntrackedCustomFile(t *testing.T) {
	f := newFixture(t)
	writeFM(t, f.root, "station/agent/Skills/new.md", "skill", "new untracked skill", nil)

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	iss := findIssue(report.Issues, CategoryUntrackedCustomFile, "new")
	if iss == nil {
		t.Fatalf("expected untracked_custom_file for new, got: %+v", report.Issues)
	}
	if iss.Severity != SeverityWarning {
		t.Errorf("severity = %s, want warning", iss.Severity)
	}
}

// TestRun_InvalidFrontmatter verifies a file without frontmatter produces
// an error, and a file with frontmatter but missing description also
// produces an error.
func TestRun_InvalidFrontmatter(t *testing.T) {
	f := newFixture(t)
	bad := filepath.Join(f.root, "station/agent/Skills/bad.md")
	if err := os.WriteFile(bad, []byte("no frontmatter at all\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Empty-description case.
	writeFM(t, f.root, "station/agent/Skills/empty.md", "skill", "", nil)

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	if iss := findIssue(report.Issues, CategoryInvalidFrontmatter, "bad"); iss == nil {
		t.Fatalf("expected invalid_frontmatter for bad, got: %+v", report.Issues)
	} else if iss.Severity != SeverityError {
		t.Errorf("bad severity = %s, want error", iss.Severity)
	}
	if iss := findIssue(report.Issues, CategoryInvalidFrontmatter, "empty"); iss == nil {
		t.Fatalf("expected invalid_frontmatter for empty, got: %+v", report.Issues)
	}
}

// TestRun_InvalidFrontmatter_SensorMissingEvent verifies sensor-specific
// validation: a sensor with valid frontmatter but missing `event` is
// flagged.
func TestRun_InvalidFrontmatter_SensorMissingEvent(t *testing.T) {
	f := newFixture(t)
	// Sensor frontmatter without an event field.
	writeFM(t, f.root, "station/agent/Sensors/no-event.sh", "sensor", "missing event", nil)

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	iss := findIssue(report.Issues, CategoryInvalidFrontmatter, "no-event")
	if iss == nil {
		t.Fatalf("expected invalid_frontmatter for no-event, got: %+v", report.Issues)
	}
	if iss.AbilityType != "sensor" {
		t.Errorf("ability_type = %s, want sensor", iss.AbilityType)
	}
}

// TestRun_InvalidFrontmatter_RoutineMissingFrequency verifies the
// routine-specific validation: missing frequency is flagged.
func TestRun_InvalidFrontmatter_RoutineMissingFrequency(t *testing.T) {
	f := newFixture(t)
	writeFM(t, f.root, "station/agent/Routines/no-freq.md", "routine", "missing frequency", nil)

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	iss := findIssue(report.Issues, CategoryInvalidFrontmatter, "no-freq")
	if iss == nil {
		t.Fatalf("expected invalid_frontmatter for no-freq, got: %+v", report.Issues)
	}
	if iss.AbilityType != "routine" {
		t.Errorf("ability_type = %s, want routine", iss.AbilityType)
	}
}

// TestRun_WrongExtensionInCategory verifies an .md file in agent/Sensors/
// or a .sh file in agent/Skills/ produces a wrong_extension warning.
func TestRun_WrongExtensionInCategory(t *testing.T) {
	f := newFixture(t)
	// .md file in Sensors dir — wrong.
	if err := os.WriteFile(filepath.Join(f.root, "station/agent/Sensors/notes.md"), []byte("md in sensors\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// .sh file in Skills dir — wrong.
	if err := os.WriteFile(filepath.Join(f.root, "station/agent/Skills/script.sh"), []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")

	// Both should appear.
	if iss := findIssue(report.Issues, CategoryWrongExtension, "notes"); iss == nil {
		t.Fatalf("expected wrong_extension for notes (md in Sensors), got: %+v", report.Issues)
	} else if iss.AbilityType != "sensor" {
		t.Errorf("notes ability_type = %s, want sensor", iss.AbilityType)
	}
	if iss := findIssue(report.Issues, CategoryWrongExtension, "script"); iss == nil {
		t.Fatalf("expected wrong_extension for script (sh in Skills), got: %+v", report.Issues)
	} else if iss.AbilityType != "skill" {
		t.Errorf("script ability_type = %s, want skill", iss.AbilityType)
	}
}

// TestRun_AgentFilter verifies --agent restricts the audit. A two-agent
// project with issues in both agents, filtered to one, should only report
// that agent's issues.
func TestRun_AgentFilter(t *testing.T) {
	f := newFixture(t)
	// Add a second agent at workspace "code".
	if err := os.MkdirAll(filepath.Join(f.root, "code/agent/Skills"), 0o755); err != nil {
		t.Fatalf("mkdir code agent: %v", err)
	}
	codeAgent := &config.InstalledAgent{
		AgentType:   "backend",
		Workspace:   "code",
		Skills:      []string{"orphan-code"},
		CustomItems: map[string]*config.CustomItemMeta{},
	}
	f.cfg.Agents["backend"] = codeAgent
	writeFM(t, f.root, "code/agent/Skills/orphan-code.md", "skill", "code skill", nil)

	// And an orphan in the tech-lead agent too.
	f.tlAgent.Skills = []string{"orphan-tl"}
	writeFM(t, f.root, "station/agent/Skills/orphan-tl.md", "skill", "tl skill", nil)

	// Filter to tech-lead only.
	report, err := Run(f.root, f.cfg, nil, f.lock, "tech-lead")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := report.AgentsScanned; len(got) != 1 || got[0] != "tech-lead" {
		t.Fatalf("AgentsScanned = %v, want [tech-lead]", got)
	}
	for _, iss := range report.Issues {
		if iss.AgentName != "tech-lead" {
			t.Fatalf("filter leak: issue from agent %q in tech-lead-only run: %+v", iss.AgentName, iss)
		}
	}
	if findIssue(report.Issues, CategoryOrphanedRegistration, "orphan-tl") == nil {
		t.Fatalf("expected orphan-tl issue, got: %+v", report.Issues)
	}
	if findIssue(report.Issues, CategoryOrphanedRegistration, "orphan-code") != nil {
		t.Fatalf("orphan-code should be filtered out, got: %+v", report.Issues)
	}
}

// TestRun_AgentFilter_UnknownAgent verifies a filter naming a
// non-installed agent returns an error.
func TestRun_AgentFilter_UnknownAgent(t *testing.T) {
	f := newFixture(t)
	if _, err := Run(f.root, f.cfg, nil, f.lock, "ghost-agent"); err == nil {
		t.Fatalf("expected error for unknown agent filter")
	}
}

// TestRun_MultipleCategoriesAtOnce stresses the report aggregation:
// a single project with three different failure modes should report all
// three. Also covers the Report.HasErrors / HasIssues distinction.
func TestRun_MultipleCategoriesAtOnce(t *testing.T) {
	f := newFixture(t)
	// 1. Missing file.
	f.tlAgent.Workflows = []string{"phantom"}
	// 2. Untracked custom (warning).
	writeFM(t, f.root, "station/agent/Skills/new.md", "skill", "new", nil)
	// 3. Wrong-extension warning.
	if err := os.WriteFile(filepath.Join(f.root, "station/agent/Sensors/oops.md"), []byte("md in sensors\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")

	if !report.HasIssues() {
		t.Fatalf("expected issues, got none")
	}
	if !report.HasErrors() {
		t.Fatalf("expected at least one error (missing_file)")
	}
	// Each of the three categories should appear at least once.
	cats := make(map[Category]bool)
	for _, iss := range report.Issues {
		cats[iss.Category] = true
	}
	for _, want := range []Category{CategoryMissingFile, CategoryUntrackedCustomFile, CategoryWrongExtension} {
		if !cats[want] {
			t.Fatalf("expected category %q in report, got: %+v", want, report.Issues)
		}
	}
}

// TestRun_CatalogTrackedItemNotFlaggedAsOrphan verifies the "tracked but
// not custom" branch — a catalog-shipped item with a non-"custom:" lock
// source must NOT be flagged as orphan even when custom_items lacks an
// entry. Catalog items legitimately have empty custom_items.
func TestRun_CatalogTrackedItemNotFlaggedAsOrphan(t *testing.T) {
	f := newFixture(t)
	f.tlAgent.Skills = []string{"planning-template"}
	writeFM(t, f.root, "station/agent/Skills/planning-template.md", "skill", "catalog skill", nil)
	// Source is the catalog format ("skills/foo"), not "custom:...".
	f.lock.Track("station/agent/Skills/planning-template.md", []byte("dummy"), "skills/planning-template")

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	if iss := findIssue(report.Issues, CategoryOrphanedRegistration, "planning-template"); iss != nil {
		t.Fatalf("catalog-tracked item must not be flagged as orphan: %+v", iss)
	}
}

// TestRun_TopLevelOnly verifies subdirectories under agent/Skills/ are
// ignored — same scoping rule as ScanCustomFiles.
func TestRun_TopLevelOnly(t *testing.T) {
	f := newFixture(t)
	// File in a subdir — should be ignored entirely.
	if err := os.MkdirAll(filepath.Join(f.root, "station/agent/Skills/nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	writeFM(t, f.root, "station/agent/Skills/nested/inner.md", "skill", "inner", nil)

	report, _ := Run(f.root, f.cfg, nil, f.lock, "")
	if report.HasIssues() {
		t.Fatalf("nested file should be ignored, got: %+v", report.Issues)
	}
}

// TestRun_NilConfig defensively checks the public API rejects nil cfg
// rather than panicking.
func TestRun_NilConfig(t *testing.T) {
	if _, err := Run(t.TempDir(), nil, nil, nil, ""); err == nil {
		t.Fatalf("expected error on nil cfg")
	}
}

// TestRun_AgentsScannedSorted confirms AgentsScanned comes back sorted
// — non-deterministic Go map iteration must not leak into the report.
func TestRun_AgentsScannedSorted(t *testing.T) {
	f := newFixture(t)
	for _, n := range []string{"zebra", "alpha", "middle"} {
		ws := n + "-ws"
		if err := os.MkdirAll(filepath.Join(f.root, ws, "agent/Skills"), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", ws, err)
		}
		f.cfg.Agents[n] = &config.InstalledAgent{AgentType: "backend", Workspace: ws}
	}
	report, err := Run(f.root, f.cfg, nil, f.lock, "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	got := append([]string(nil), report.AgentsScanned...)
	want := append([]string(nil), got...)
	sort.Strings(want)
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("AgentsScanned not sorted: got %v, want %v", got, want)
		}
	}
}
