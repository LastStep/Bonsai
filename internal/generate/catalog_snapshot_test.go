package generate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/LastStep/Bonsai/internal/catalog"
)

// buildSnapshotTestCatalog returns a small but realistic catalog covering
// every category so the snapshot test exercises all branches of the
// serializer.
func buildSnapshotTestCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	fsys := fstest.MapFS{
		// Agents
		"agents/tech-lead/agent.yaml": &fstest.MapFile{Data: []byte("name: tech-lead\ndisplay_name: Tech Lead\ndescription: orchestrator\n")},
		"agents/backend/agent.yaml":   &fstest.MapFile{Data: []byte("name: backend\ndisplay_name: Backend\ndescription: backend dev\n")},
		// Skills
		"skills/planning-template/meta.yaml":  &fstest.MapFile{Data: []byte("name: planning-template\ndescription: writing plans\nagents:\n  - tech-lead\nrequired:\n  - tech-lead\n")},
		"skills/planning-template/content.md": &fstest.MapFile{Data: []byte("content")},
		"skills/coding-standards/meta.yaml":   &fstest.MapFile{Data: []byte("name: coding-standards\ndescription: code style\nagents: all\n")},
		"skills/coding-standards/content.md":  &fstest.MapFile{Data: []byte("content")},
		// Workflows
		"workflows/code-review/meta.yaml":  &fstest.MapFile{Data: []byte("name: code-review\ndescription: reviewing code\nagents: all\n")},
		"workflows/code-review/content.md": &fstest.MapFile{Data: []byte("c")},
		// Protocols
		"protocols/security/meta.yaml":  &fstest.MapFile{Data: []byte("name: security\ndescription: security protocol\nagents: all\n")},
		"protocols/security/content.md": &fstest.MapFile{Data: []byte("c")},
		// Sensors
		"sensors/scope-guard-files/meta.yaml":                 &fstest.MapFile{Data: []byte("name: scope-guard-files\ndescription: scope guard\nevent: PreToolUse\nmatcher: Edit|Write\nagents: all\n")},
		"sensors/scope-guard-files/scope-guard-files.sh.tmpl": &fstest.MapFile{Data: []byte("#!/bin/bash\n")},
		// Routines
		"routines/backlog-hygiene/meta.yaml":               &fstest.MapFile{Data: []byte("name: backlog-hygiene\ndescription: groom backlog\nfrequency: 7 days\nagents:\n  - tech-lead\n")},
		"routines/backlog-hygiene/backlog-hygiene.md.tmpl": &fstest.MapFile{Data: []byte("c")},
	}
	cat, err := catalog.New(fsys)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	return cat
}

// TestWriteCatalogSnapshot_RoundTrip ensures that the written JSON can be
// round-tripped back into a CatalogSnapshot with all category entries intact.
func TestWriteCatalogSnapshot_RoundTrip(t *testing.T) {
	cat := buildSnapshotTestCatalog(t)
	tmpDir := t.TempDir()

	var wr WriteResult
	if err := WriteCatalogSnapshot(tmpDir, "test-1.0", cat, &wr); err != nil {
		t.Fatalf("WriteCatalogSnapshot: %v", err)
	}

	// File exists under .bonsai/
	path := filepath.Join(tmpDir, ".bonsai", "catalog.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	var snap CatalogSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if snap.Version != "test-1.0" {
		t.Errorf("version = %q, want test-1.0", snap.Version)
	}

	// Agents — tech-lead must be present.
	if !containsAgent(snap.Agents, "tech-lead") {
		t.Errorf("snap.agents missing tech-lead: %+v", snap.Agents)
	}

	// Skills — planning-template present with required + agents fields.
	var planningFound bool
	for _, s := range snap.Skills {
		if s.Name == "planning-template" {
			planningFound = true
			if len(s.Agents) == 0 || s.Agents[0] != "tech-lead" {
				t.Errorf("planning-template agents = %v, want [tech-lead]", s.Agents)
			}
			if len(s.Required) == 0 || s.Required[0] != "tech-lead" {
				t.Errorf("planning-template required = %v, want [tech-lead]", s.Required)
			}
			break
		}
	}
	if !planningFound {
		t.Errorf("skills missing planning-template: %+v", snap.Skills)
	}

	// Routines — backlog-hygiene with frequency populated.
	var routineFound bool
	for _, r := range snap.Routines {
		if r.Name == "backlog-hygiene" {
			routineFound = true
			if r.Frequency != "7 days" {
				t.Errorf("backlog-hygiene frequency = %q, want 7 days", r.Frequency)
			}
			break
		}
	}
	if !routineFound {
		t.Errorf("routines missing backlog-hygiene: %+v", snap.Routines)
	}

	// Sensor event + matcher populated.
	var sensorFound bool
	for _, s := range snap.Sensors {
		if s.Name == "scope-guard-files" {
			sensorFound = true
			if s.Event != "PreToolUse" {
				t.Errorf("sensor event = %q, want PreToolUse", s.Event)
			}
			if s.Matcher != "Edit|Write" {
				t.Errorf("sensor matcher = %q, want Edit|Write", s.Matcher)
			}
			break
		}
	}
	if !sensorFound {
		t.Errorf("sensors missing scope-guard-files: %+v", snap.Sensors)
	}

	// WriteResult — should contain one Created action for the snapshot.
	var snapshotTracked bool
	for _, f := range wr.Files {
		if f.RelPath == filepath.Join(".bonsai", "catalog.json") {
			snapshotTracked = true
			if f.Action != ActionCreated {
				t.Errorf("snapshot action = %v, want ActionCreated", f.Action)
			}
			break
		}
	}
	if !snapshotTracked {
		t.Errorf(".bonsai/catalog.json not tracked in WriteResult: %+v", wr.Files)
	}
}

// TestWriteCatalogSnapshot_Idempotent — re-writing with identical content
// emits ActionUnchanged, not another Created/Updated.
func TestWriteCatalogSnapshot_Idempotent(t *testing.T) {
	cat := buildSnapshotTestCatalog(t)
	tmpDir := t.TempDir()

	var wr1 WriteResult
	if err := WriteCatalogSnapshot(tmpDir, "test-1.0", cat, &wr1); err != nil {
		t.Fatalf("first write: %v", err)
	}

	var wr2 WriteResult
	if err := WriteCatalogSnapshot(tmpDir, "test-1.0", cat, &wr2); err != nil {
		t.Fatalf("second write: %v", err)
	}

	if len(wr2.Files) != 1 {
		t.Fatalf("want 1 tracked file, got %d", len(wr2.Files))
	}
	if wr2.Files[0].Action != ActionUnchanged {
		t.Errorf("second write action = %v, want ActionUnchanged", wr2.Files[0].Action)
	}
}

// TestWriteCatalogSnapshot_CreatesDir — .bonsai/ dir does not pre-exist; the
// write call creates it with mode 0755 (parent-observable).
func TestWriteCatalogSnapshot_CreatesDir(t *testing.T) {
	cat := buildSnapshotTestCatalog(t)
	tmpDir := t.TempDir()

	// Sanity: no .bonsai/ yet.
	if _, err := os.Stat(filepath.Join(tmpDir, ".bonsai")); !os.IsNotExist(err) {
		t.Fatalf("precondition: .bonsai/ should not exist, err=%v", err)
	}

	var wr WriteResult
	if err := WriteCatalogSnapshot(tmpDir, "v1", cat, &wr); err != nil {
		t.Fatalf("write: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmpDir, ".bonsai"))
	if err != nil {
		t.Fatalf(".bonsai/ missing after write: %v", err)
	}
	if !info.IsDir() {
		t.Errorf(".bonsai/ is not a directory")
	}
}

func containsAgent(agents []AgentEntry, name string) bool {
	for _, a := range agents {
		if a.Name == name {
			return true
		}
	}
	return false
}
