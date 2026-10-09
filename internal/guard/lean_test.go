package guard

// The guard's read of bonsai.yaml stays lean (plan-5 5.1.4a): it reads the fields it judges by through
// workspace.LoadConfig, never the full read the engine and check use (workspace.LoadConfigFull), so a field it does
// not judge by, broken, never blocks an edit, and a hook call never pays for the full read.

import (
	"errors"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// leanFields are the keys the guard's read (workspace.ReadConfig) reads and checks: the format line, the workspace
// id its records carry, the name and packs the lean read has checked since the walking skeleton, and the protected
// lists and never_edit, which it judges by. Every other key of bonsai.workspace/1 is outside it.
var leanFields = []string{"format", "id", "name", "packs", "protected", "person_only", "never_edit"}

// brokenElsewhere breaks every field of bonsai.workspace/1 outside leanFields, each in a way the full read refuses.
var brokenElsewhere = map[string]string{
	"documents":       "documents:\n  task: \"/abs/tasks\"\n",
	"ladder_floor":    "ladder_floor: [\"zero\"]\n",
	"ladder":          "ladder:\n  - rung: 1\n    kind: command\n    command: \"go test ./...\"\n    required: true\n    timeout_s: \"ten\"\n",
	"ratchets":        "ratchets:\n  tests: \"many\"\n",
	"ci_marked_tests": "ci_marked_tests: 3\n",
	"generated":       "generated:\n  log:\n    keep_days: -1\n  sessions:\n    keep_newest: \"all\"\n",
}

func TestGuardReadIsLean(t *testing.T) {
	// Every field of the schema is either read by the guard or broken below.
	props := workspaceProps(t)
	for _, p := range props {
		if !contains(leanFields, p) && brokenElsewhere[p] == "" {
			t.Errorf("bonsai.workspace/1's field %s is neither in the guard's read nor broken by this test", p)
		}
	}
	for field, broken := range brokenElsewhere {
		t.Run(field, func(t *testing.T) {
			dir := project(t, testYAML+broken)
			// The full read refuses the file, naming the field ...
			_, err := workspace.LoadConfigFull(dir)
			var we *workspace.Error
			if !errors.As(err, &we) || !strings.Contains(we.Msg, "field "+field) {
				t.Fatalf("the full read does not refuse the broken %s: %v", field, err)
			}
			// ... and the guard still decides: a protected path is refused for being protected, a free one allowed.
			code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lean", dir, "protected.txt")), 0)
			if code != 2 || !strings.Contains(stderr, "protected") || strings.Contains(stderr, "bonsai.yaml cannot be read") {
				t.Errorf("a protected edit: exit %d, %s", code, stderr)
			}
			code, stderr = guardRun(t, env(dir), strings.NewReader(editOf("lean", dir, "free.txt")), 0)
			if code != 0 {
				t.Errorf("a free edit: exit %d, %s", code, stderr)
			}
		})
	}
	// A broken field the guard does read still blocks, as it always has: the guard cannot tell what is protected.
	dir := project(t, strings.Replace(testYAML, `protected: [`, `protected: ["/abs", `, 1))
	if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lean", dir, "free.txt")), 0); code != 2 ||
		!strings.Contains(stderr, "bonsai.yaml cannot be read") {
		t.Errorf("a broken protected list: exit %d, %s", code, stderr)
	}
}

// workspaceProps lists bonsai.workspace/1's top-level fields, from the schema (their one home).
func workspaceProps(t *testing.T) []string {
	t.Helper()
	raw, err := formats.Schema("workspace")
	if err != nil {
		t.Fatal(err)
	}
	s, err := schema.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	props, _ := s.Get("properties")
	keys := props.(schema.Object).Keys()
	if len(keys) < len(leanFields)+len(brokenElsewhere) {
		t.Fatalf("the schema lists %d fields", len(keys))
	}
	return keys
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
