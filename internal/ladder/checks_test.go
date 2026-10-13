package ladder

// The settings for the checks through the runner and the stop gate (K1; workspace/checks.go): each looser kind
// committed on the base branch refuses the climb (config-looser) and blocks the gate (stop-config-looser), naming it;
// each stricter kind is accepted, and the climb brings this machine's record up to it, while the gate only compares; a
// nested climb compares and writes nothing; a record that holds no settings fails closed until bonsai update.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/rungjob"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// checksBase is the fixture's bonsai.yaml: rung 1 counts "3 passed" into the ratchet found (floor 3), rung 2 echoes
// and is not required, both on the floor; a CI mark, and a path on each of the three lists.
func checksBase(extra ...rung) string {
	rungs := append([]rung{{n: 1, name: "one", command: "echo 3 passed", required: true, ratchet: "found",
		capture: "\n      found: \"([0-9]+) passed\""}, {n: 2, name: "two", command: "echo two"}}, extra...)
	return strings.NewReplacer("protected: []", `protected: ["guarded/**"]`, "person_only: []", `person_only: ["bonsai.yaml"]`,
		"never_edit: []", `never_edit: ["ledger.json"]`).Replace(yamlOf("[1, 2]", "\n  found: 3", `["TestMarked"]`, rungs...))
}

// inRung2 changes old to new within rung 2's lines of the yaml.
func inRung2(t *testing.T, yaml, old, new string) string {
	t.Helper()
	i := strings.Index(yaml, "  - rung: 2\n")
	j := strings.Index(yaml, "ratchets:")
	if i < 0 || j < i || !strings.Contains(yaml[i:j], old) {
		t.Fatalf("rung 2 has no %q:\n%s", old, yaml)
	}
	return yaml[:i] + strings.Replace(yaml[i:j], old, new, 1) + yaml[j:]
}

// replaced is yaml with old replaced by new once, which must be there.
func replaced(t *testing.T, yaml, old, new string) string {
	t.Helper()
	if !strings.Contains(yaml, old) {
		t.Fatalf("no %q in:\n%s", old, yaml)
	}
	return strings.Replace(yaml, old, new, 1)
}

// accepted is this machine's record's settings for the checks, as JSON.
func (f *fx) accepted() string {
	f.t.Helper()
	c, tr := workspace.AcceptedChecks(f.home, f.root)
	if !tr.Verified {
		f.t.Fatalf("no settings recorded: %+v", tr)
	}
	return schema.Show(c.JSON())
}

func TestLooserKindsRefused(t *testing.T) {
	base := checksBase()
	f := project(t, base, map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != ExitGreen {
		t.Fatalf("the base climbs %d, %+v\n%s", a.Exit, a.Error, f.out.String())
	}
	cases := []struct{ name, yaml, want string }{
		{"a rung out of the floor", replaced(t, base, "ladder_floor: [1, 2]", "ladder_floor: [1]"), "ladder_floor no longer holds rung 2"},
		{"a rung out of the ladder", inRung2(t, base, "  - rung: 2\n", "  - rung: 3\n"), "rung 2 is no longer in the ladder"},
		{"a kind changed", inRung2(t, base, "kind: command", "kind: ver-git"), `rung 2's kind changed from "command" to "ver-git"`},
		{"a command changed", inRung2(t, base, `command: "echo two"`, `command: "echo 2"`), `rung 2's command changed from "echo two" to "echo 2"`},
		{"required taken off", replaced(t, base, "    required: true\n", "    required: false\n"), "rung 1 is no longer required"},
		{"tests changed", inRung2(t, base, "tests: null", "tests: tap"), `rung 2's tests changed from null to "tap"`},
		{"a ratchet taken off", replaced(t, base, "ratchet: found", "ratchet: null"), `rung 1's ratchet "found" was taken out`},
		{"a ratchet renamed", replaced(t, base, "ratchet: found", "ratchet: other"), `rung 1's ratchet changed from "found" to "other"`},
		{"a capture changed", replaced(t, base, `"([0-9]+) passed"`, `"(3) passed"`), `rung 1's capture changed from`},
		{"base_setup set", inRung2(t, base, "base_setup: null", `base_setup: "echo setup"`), `rung 2's base_setup changed from null to "echo setup"`},
		{"a floor lowered", replaced(t, base, "\n  found: 3", "\n  found: 2"), "ratchets.found fell from 3 to 2"},
		{"a floor taken out", replaced(t, base, "ratchets: \n  found: 3", "ratchets: {}"), "ratchets.found was taken out (its floor was 3"},
		{"a CI mark added", replaced(t, base, `["TestMarked"]`, `["TestMarked", "TestOther"]`), `ci_marked_tests gained "TestOther"`},
		{"a protected path taken out", replaced(t, base, `protected: ["guarded/**"]`, "protected: []"), `protected no longer holds "guarded/**"`},
		{"a person-only path taken out", replaced(t, base, `person_only: ["bonsai.yaml"]`, "person_only: []"), `person_only no longer holds "bonsai.yaml"`},
		{"a never-edit path taken out", replaced(t, base, `never_edit: ["ledger.json"]`, "never_edit: []"), `never_edit no longer holds "ledger.json"`},
		{"the task folder moved", replaced(t, base, "  task: work/tasks\n", "  task: work/elsewhere\n"),
			`documents.task changed from "work/tasks" to "work/elsewhere"`},
	}
	was := f.accepted()
	for _, c := range cases {
		f.write("bonsai.yaml", c.yaml)
		f.commit(c.name)
		a := f.climb("", Options{Task: "T-0901"})
		if a.Exit != ExitState || a.Error == nil || a.Error.Code != "config-looser" || a.Error.Who != "person" ||
			!strings.Contains(a.Error.What, "sets the checks looser than the settings this machine last accepted for the workspace: "+c.want) {
			t.Errorf("%s: exit %d, %+v", c.name, a.Exit, a.Error)
		}
		if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "sets the checks looser") ||
			!strings.Contains(stderr, c.want) {
			t.Errorf("%s: the stop gate %d %q", c.name, code, stderr)
		}
		if f.accepted() != was {
			t.Errorf("%s: the record changed", c.name)
		}
		testpack.Git(t, f.root, "reset", "-q", "--hard", "HEAD~1")
	}
	// Several at once: each named, in a fixed order, the list cut at eight.
	many := strings.NewReplacer("ladder_floor: [1, 2]", "ladder_floor: []", `["TestMarked"]`, `["A", "B", "C", "D", "E", "F", "G", "H"]`).Replace(base)
	f.write("bonsai.yaml", many)
	f.commit("many")
	a := f.climb("", Options{Task: "T-0901"})
	if a.Error == nil || !strings.Contains(a.Error.What, "ladder_floor no longer holds rung 1; ladder_floor no longer holds rung 2; "+
		`ci_marked_tests gained "A"`) || !strings.Contains(a.Error.What, `ci_marked_tests gained "F"; and 2 more, so nothing was run`) {
		t.Errorf("many: %+v", a.Error)
	}
}

func TestStricterKindsAccepted(t *testing.T) {
	base := checksBase()
	f := project(t, base, map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	three := rung{n: 3, name: "three", command: "echo three", required: true}
	cases := []struct{ name, yaml string }{
		{"a rung added to the floor", replaced(t, checksBase(three), "ladder_floor: [1, 2]", "ladder_floor: [1, 2, 3]")},
		{"a rung added to the ladder", checksBase(three)},
		{"required put on", inRung2(t, base, "required: false", "required: true")},
		{"a ratchet put on a rung", inRung2(t, base, "ratchet: null", "ratchet: found")},
		{"a floor raised", replaced(t, base, "\n  found: 3", "\n  found: 4")},
		{"a floor added", replaced(t, base, "\n  found: 3", "\n  found: 3\n  more: 1")},
		{"a CI mark taken out", replaced(t, base, `["TestMarked"]`, "[]")},
		{"a protected path added", replaced(t, base, `protected: ["guarded/**"]`, `protected: ["guarded/**", "more/**"]`)},
		{"a person-only path added", replaced(t, base, `person_only: ["bonsai.yaml"]`, `person_only: ["bonsai.yaml", "guarded/**"]`)},
		{"a never-edit path added", replaced(t, base, `never_edit: ["ledger.json"]`, `never_edit: ["ledger.json", "more.json"]`)},
		{"a name, a means and a timeout changed", strings.Replace(inRung2(t, base, `name: "two"`, `name: "second"`), "timeout_s: null",
			"timeout_s: 900", 1)},
	}
	was := f.accepted()
	for _, c := range cases {
		f.write("bonsai.yaml", c.yaml)
		f.commit(c.name)
		// The gate compares and writes nothing: the result proves the commit before, and the settings are not looser.
		if code, stderr := gate(t, f.root, "T-0901"); code != 2 || strings.Contains(stderr, "looser") || f.accepted() != was {
			t.Errorf("%s: the stop gate %d %q, or it wrote the record", c.name, code, stderr)
		}
		f.out.Reset()
		a := f.climb("", Options{Task: "T-0901"})
		c2, err := workspace.ReadConfigFull([]byte(c.yaml))
		if err != nil {
			t.Fatal(err)
		}
		if a.Exit == ExitState || a.Exit == ExitInput || f.accepted() != schema.Show(workspace.ChecksOf(c2.Full).JSON()) {
			t.Errorf("%s: exit %d, %+v, or not recorded:\n%s", c.name, a.Exit, a.Error, f.out.String())
		}
		raised := strings.Contains(f.out.String(), "note: this machine's record now holds bonsai.yaml's settings for the checks as "+
			"committed, stricter than the ones it held")
		if c.name == "a name, a means and a timeout changed" {
			raised = !raised // nothing it keeps changed: nothing written, no note
		}
		if !raised {
			t.Errorf("%s: the note:\n%s", c.name, f.out.String())
		}
		// The base again reads as looser than what was recorded (but for words it does not keep): a person records it.
		testpack.Git(t, f.root, "reset", "-q", "--hard", "HEAD~1")
		if a := f.climb("", Options{Task: "T-0901"}); (a.Exit == ExitState) == (c.name == "a name, a means and a timeout changed") {
			t.Errorf("%s: back to the base: exit %d, %+v", c.name, a.Exit, a.Error)
		}
		f.accept()
	}
	// A nested climb (inside a rung) compares and writes nothing: the outer climb holds the lock and wrote them.
	f.write("bonsai.yaml", replaced(t, base, "\n  found: 3", "\n  found: 3\n  more: 1"))
	f.commit("a floor added")
	f.env[rungjob.LadderEnv] = "1"
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit == ExitState || f.accepted() != was {
		t.Errorf("a nested climb: exit %d, %+v, or it wrote the record", a.Exit, a.Error)
	}
	delete(f.env, rungjob.LadderEnv)
	// A --ci climb neither compares nor writes: a looser floor climbs there (a CI run proves no task).
	f.write("bonsai.yaml", replaced(t, base, "ladder_floor: [1, 2]", "ladder_floor: [1]"))
	f.commit("a looser floor")
	if a := f.climb("", Options{CI: true}); a.Exit != ExitGreen || f.accepted() != was {
		t.Errorf("--ci: exit %d, %+v, or it wrote the record", a.Exit, a.Error)
	}
}

// A record that holds no settings for the checks (written before Bonsai kept them) fails closed until bonsai update.
func TestRecordWithoutChecks(t *testing.T) {
	f := project(t, checksBase(), map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
	dir, err := workspace.MachineDir(f.home, f.root)
	if err != nil {
		t.Fatal(err)
	}
	old := `{"path": "` + filepath.ToSlash(f.root) + `", "ids": [{"id": "` + testID + `", "since": "2026-10-13T07:00:00Z"}], ` +
		`"base_branch": "main"}`
	if err := os.WriteFile(filepath.Join(dir, workspace.MachineRecordFile), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != ExitState || a.Error == nil || a.Error.Code != "config-looser" || a.Error.Who != "person" ||
		!strings.Contains(a.Error.What, "this machine's record of the main checkout holds no settings for the checks to compare "+
			"bonsai.yaml's with (it was written before Bonsai kept them), so nothing was run") ||
		a.Error.Next != "a person runs bonsai update in the main checkout, which records bonsai.yaml's settings for the checks as "+
			"accepted; then run: bonsai ladder --task T-0901 --json" {
		t.Fatalf("exit %d, %+v", a.Exit, a.Error)
	}
	if code, stderr := gate(t, f.root, "T-0901"); code != 2 || !strings.Contains(stderr, "holds no settings for the checks") {
		t.Errorf("the stop gate: %d %q", code, stderr)
	}
	if _, err := workspace.RecordCheckout(f.home, f.root, testID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != ExitGreen {
		t.Errorf("recorded by bonsai update: exit %d, %+v", a.Exit, a.Error)
	}
}
