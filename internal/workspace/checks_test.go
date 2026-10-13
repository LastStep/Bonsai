package workspace

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// checksYAML is a bonsai.yaml whose settings for the checks the tests below change one at a time.
const checksYAML = `format: bonsai.workspace/1
id: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa
name: checks-test
packs: []
documents:
  task: work/tasks
  run: work/runs
  answers: work/answers.md
  memory: work/memory
  protocols: work/protocols
protected: ["guarded/**", "bonsai.yaml"]
person_only: ["bonsai.yaml"]
never_edit: ["ledger.json"]
ladder_floor: [0, 1, 2]
ladder:
  - rung: 0
    name: "guard"
    kind: guard
    command: null
    required: true
    timeout_s: 120
    ratchet: null
    capture: null
    means: "Every change is inside the task's grants."
    tests: null
    base_setup: null
  - rung: 1
    name: "tests"
    kind: command
    command: "go test -json ./..."
    required: true
    timeout_s: 300
    ratchet: go_tests
    capture: null
    means: "Every test passes."
    tests: go-test-json
    base_setup: null
  - rung: 2
    name: "count"
    kind: command
    command: "echo 3 passed"
    required: false
    timeout_s: null
    ratchet: null
    capture:
      found: "([0-9]+) passed"
    means: "A count."
    tests: null
    base_setup: null
ratchets:
  go_tests: 2
  found: 1.5
ci_marked_tests: ["TestMarked"]
generated:
  log:
    keep_days: null
    keep_newest: null
`

// checksOf reads a bonsai.yaml's settings for the checks.
func checksOf(t *testing.T, yaml string) *Checks {
	t.Helper()
	c, err := ReadConfigFull([]byte(yaml))
	if err != nil {
		t.Fatalf("%v\n%s", err, yaml)
	}
	return ChecksOf(c.Full)
}

// addedRung is one more ladder entry of number n, for the fixture's ladder.
func addedRung(n int, kind, command string) string {
	return "  - rung: " + strconv.Itoa(n) + "\n    name: \"more\"\n    kind: " + kind + "\n    command: " + command + "\n    required: true\n" +
		"    timeout_s: null\n    ratchet: null\n    capture: null\n    means: null\n    tests: null\n    base_setup: null\n"
}

// Each looser change is named, in words a person reads; each stricter or equal one reads as none (checks.go).
func TestLoosenings(t *testing.T) {
	was := checksOf(t, checksYAML)
	if l := Loosenings(was, checksOf(t, checksYAML)); len(l) != 0 {
		t.Fatalf("the same settings read as looser: %q", l)
	}
	if Loosenings(nil, was) != nil || Loosenings(was, nil) != nil {
		t.Error("nothing to compare gave loosenings")
	}
	edit := func(old, new string) string {
		t.Helper()
		if !strings.Contains(checksYAML, old) {
			t.Fatalf("the fixture has no %q", old)
		}
		return strings.Replace(checksYAML, old, new, 1)
	}
	looser := []struct{ name, yaml, want string }{
		{"a rung taken out of the floor", edit("ladder_floor: [0, 1, 2]", "ladder_floor: [0, 2]"), "ladder_floor no longer holds rung 1"},
		{"a rung taken out of the ladder", checksYAML[:strings.Index(checksYAML, "  - rung: 2\n")] + checksYAML[strings.Index(checksYAML, "ratchets:\n"):],
			"rung 2 is no longer in the ladder"},
		{"a kind changed", edit("    kind: guard\n", "    kind: ver-git\n"), `rung 0's kind changed from "guard" to "ver-git"`},
		{"a command changed", edit(`command: "go test -json ./..."`, `command: "true"`),
			`rung 1's command changed from "go test -json ./..." to "true"`},
		{"required taken off", edit("    required: true\n    timeout_s: 300", "    required: false\n    timeout_s: 300"),
			"rung 1 is no longer required"},
		{"tests changed", edit("tests: go-test-json", "tests: null"), `rung 1's tests changed from "go-test-json" to null`},
		{"a capture changed", edit(`found: "([0-9]+) passed"`, `found: "(9) passed"`),
			`rung 2's capture changed from {"found":"([0-9]+) passed"} to {"found":"(9) passed"}`},
		{"base_setup set", edit("    tests: go-test-json\n    base_setup: null", "    tests: go-test-json\n    base_setup: \"git checkout -- .\""),
			`rung 1's base_setup changed from null to "git checkout -- ."`},
		{"a ratchet taken off a rung", edit("ratchet: go_tests", "ratchet: null"), `rung 1's ratchet "go_tests" was taken out`},
		{"a ratchet renamed", edit("ratchet: go_tests", "ratchet: other"), `rung 1's ratchet changed from "go_tests" to "other"`},
		{"a floor lowered", edit("  go_tests: 2\n", "  go_tests: 1\n"), "ratchets.go_tests fell from 2 to 1"},
		{"a fraction lowered", edit("  found: 1.5\n", "  found: 1.25\n"), "ratchets.found fell from 1.5 to 1.25"},
		{"a floor taken out", edit("  go_tests: 2\n", ""), "ratchets.go_tests was taken out (its floor was 2; a ratchet with no floor reads as 0)"},
		{"a CI mark added", edit(`ci_marked_tests: ["TestMarked"]`, `ci_marked_tests: ["TestMarked", "TestSkipped"]`),
			`ci_marked_tests gained "TestSkipped"`},
		{"a protected path taken out", edit(`protected: ["guarded/**", "bonsai.yaml"]`, `protected: ["bonsai.yaml"]`),
			`protected no longer holds "guarded/**"`},
		{"a person-only path taken out", edit(`person_only: ["bonsai.yaml"]`, `person_only: []`), `person_only no longer holds "bonsai.yaml"`},
		{"a never-edit path taken out", edit(`never_edit: ["ledger.json"]`, `never_edit: []`), `never_edit no longer holds "ledger.json"`},
		{"the task folder moved", edit("  task: work/tasks\n", "  task: elsewhere\n"), `documents.task changed from "work/tasks" to "elsewhere"`},
		{"a glob made wider reads as looser", edit(`protected: ["guarded/**", "bonsai.yaml"]`, `protected: ["guarded/*", "bonsai.yaml"]`),
			`protected no longer holds "guarded/**"`},
		// A rung added runs inside the climb (V2c's break-it): a person records it, whatever it holds.
		{"a rung added to the ladder", edit("ratchets:\n", addedRung(3, "command", `"echo more"`)+"ratchets:\n"),
			"rung 3 was added to the ladder: its command runs inside the climb"},
		{"a second rung of a number", edit("ratchets:\n", addedRung(2, "command", `"echo more"`)+"ratchets:\n"),
			"rung 2 (its entry 2 in the ladder) was added: its command runs inside the climb"},
		{"a guard rung added", edit("ratchets:\n", addedRung(4, "guard", "null")+"ratchets:\n"),
			"rung 4 was added to the ladder: its check runs inside the climb"},
	}
	for _, c := range looser {
		got := Loosenings(was, checksOf(t, c.yaml))
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%s: %q, want [%q]", c.name, got, c.want)
		}
	}
	stricter := []struct{ name, yaml string }{
		{"a rung added to the floor", edit("ladder_floor: [0, 1, 2]", "ladder_floor: [0, 1, 2, 3]")},
		{"required put on", edit("    required: false\n", "    required: true\n")},
		{"a ratchet put on a rung", edit("    ratchet: null\n    capture:\n", "    ratchet: found\n    capture:\n")},
		{"a floor raised", edit("  go_tests: 2\n", "  go_tests: 7\n")},
		{"a floor added", edit("  found: 1.5\n", "  found: 1.5\n  more: 3\n")},
		{"a CI mark taken out", edit(`ci_marked_tests: ["TestMarked"]`, `ci_marked_tests: []`)},
		{"a protected path added", edit(`protected: ["guarded/**", "bonsai.yaml"]`, `protected: ["guarded/**", "bonsai.yaml", "more/**"]`)},
		{"a person-only path added", edit(`person_only: ["bonsai.yaml"]`, `person_only: ["bonsai.yaml", "guarded/**"]`)},
		{"a never-edit path added", edit(`never_edit: ["ledger.json"]`, `never_edit: ["ledger.json", "more.json"]`)},
		{"a name, a means and a timeout changed", strings.Replace(edit(`name: "tests"`, `name: "every test"`), "timeout_s: 300", "timeout_s: 900", 1)},
		{"the keep settings changed", edit("    keep_days: null\n", "    keep_days: 1\n")},
	}
	for _, c := range stricter {
		now := checksOf(t, c.yaml)
		if got := Loosenings(was, now); len(got) != 0 {
			t.Errorf("%s reads as looser: %q", c.name, got)
		}
	}
	// A recorded rung put on the floor: its check is one the record holds.
	if got := Loosenings(checksOf(t, edit("ladder_floor: [0, 1, 2]", "ladder_floor: [0, 1]")), was); len(got) != 0 {
		t.Errorf("a recorded rung put on the floor reads as looser: %q", got)
	}
	// An entry put before a recorded one of its number: matched in the file's order, the new entry reads as the recorded
	// one changed and the recorded one as added; the entries added come after the recorded rungs' findings.
	before := edit("  - rung: 2\n", addedRung(2, "command", `"echo 3 passed"`)+"  - rung: 2\n")
	moved := []string{`rung 2's capture changed from {"found":"([0-9]+) passed"} to null`,
		"rung 2 (its entry 2 in the ladder) was added: its command runs inside the climb"}
	if got := Loosenings(was, checksOf(t, before)); strings.Join(got, "|") != strings.Join(moved, "|") {
		t.Errorf("an entry put before a recorded one: %q", got)
	}
	// Several at once: each named, in the fixed order (floor, rungs, ratchets, marks, lists, documents).
	many := strings.NewReplacer("ladder_floor: [0, 1, 2]", "ladder_floor: [0]", "  go_tests: 2\n", "  go_tests: 0\n",
		`never_edit: ["ledger.json"]`, "never_edit: []", "ratchet: go_tests", "ratchet: null").Replace(checksYAML)
	want := []string{"ladder_floor no longer holds rung 1", "ladder_floor no longer holds rung 2", `rung 1's ratchet "go_tests" was taken out`,
		"ratchets.go_tests fell from 2 to 0", `never_edit no longer holds "ledger.json"`}
	if got := Loosenings(was, checksOf(t, many)); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("several: %q", got)
	}
}

// The record keeps the settings: written from the main checkout's bonsai.yaml, read back the same, byte-stable; a
// climb's RaiseChecks brings them up to stricter ones and writes nothing for looser or the same; a record with none
// fails closed; a checks field Bonsai does not read is a damaged record.
func TestRecordedChecks(t *testing.T) {
	home, main := t.TempDir(), t.TempDir()
	const id = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"
	at := time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC)
	put := func(body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(main, ConfigFile), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dir, _ := MachineDir(home, main)
	key := filepath.Join(dir, MachineRecordFile)
	// No bonsai.yaml that reads: no settings recorded, and a climb fails closed.
	if wrote, err := RecordCheckout(home, main, id, at); !wrote || err != nil {
		t.Fatalf("first: %v %v", wrote, err)
	}
	if raw, _ := os.ReadFile(key); !strings.Contains(string(raw), "\"checks\": null\n") {
		t.Errorf("a record with no settings:\n%s", raw)
	}
	if c, tr := AcceptedChecks(home, main); c != nil || tr.Verified || !strings.Contains(tr.Why, "(it was written before Bonsai kept them)") ||
		!strings.Contains(tr.Next, "a person runs bonsai update in the main checkout") {
		t.Errorf("no settings: %+v", tr)
	}
	if _, _, err := RaiseChecks(home, main, checksOf(t, checksYAML)); err == nil || !strings.Contains(err.Error(), "holds no settings for the checks") {
		t.Errorf("a climb raised a record with no settings: %v", err)
	}
	if _, tr := AcceptedChecks(t.TempDir(), main); tr.Verified || !strings.Contains(tr.Why, "no record") {
		t.Errorf("no record: %+v", tr)
	}
	// bonsai.yaml there: recorded (bonsai update's step), and read back the same.
	put(checksYAML)
	if wrote, err := RecordCheckout(home, main, id, at); !wrote || err != nil {
		t.Fatalf("with bonsai.yaml: %v %v", wrote, err)
	}
	raw, _ := os.ReadFile(key)
	for _, s := range []string{"  \"checks\": {\n    \"ladder_floor\": [\n      0,\n      1,\n      2\n    ],", `"command": "go test -json ./..."`,
		`"ratchets": {` + "\n" + `      "go_tests": 2,` + "\n" + `      "found": 1.5`, `"task": "work/tasks"`} {
		if !strings.Contains(string(raw), s) {
			t.Errorf("the record lacks %q:\n%s", s, raw)
		}
	}
	got, tr := AcceptedChecks(home, main)
	if !tr.Verified || !sameChecks(got, checksOf(t, checksYAML)) {
		t.Fatalf("read back: %+v %+v", got, tr)
	}
	if wrote, _ := RecordCheckout(home, main, id, at); wrote {
		t.Error("the same settings again wrote")
	}
	// A climb's raise: stricter is recorded, looser and the same write nothing.
	stricter := checksOf(t, strings.Replace(checksYAML, "  go_tests: 2\n", "  go_tests: 3\n", 1))
	if raised, l, err := RaiseChecks(home, main, stricter); !raised || l != nil || err != nil {
		t.Fatalf("stricter: %v %q %v", raised, l, err)
	}
	if c, _ := AcceptedChecks(home, main); !sameChecks(c, stricter) {
		t.Error("the stricter settings were not recorded")
	}
	if raised, l, err := RaiseChecks(home, main, stricter); raised || l != nil || err != nil {
		t.Errorf("the same: %v %q %v", raised, l, err)
	}
	before, _ := os.ReadFile(key)
	if raised, l, err := RaiseChecks(home, main, checksOf(t, checksYAML)); raised || len(l) != 1 || l[0] != "ratchets.go_tests fell from 3 to 2" || err != nil {
		t.Errorf("looser: %v %q %v", raised, l, err)
	}
	if after, _ := os.ReadFile(key); string(after) != string(before) {
		t.Error("a looser raise wrote the record")
	}
	// bonsai update records what it read, looser too (its caller refuses that to an agent); nil keeps what is there.
	if wrote, err := RecordCheckoutWith(home, main, id, at, nil); wrote || err != nil {
		t.Errorf("nil settings wrote: %v %v", wrote, err)
	}
	if wrote, err := RecordCheckoutWith(home, main, id, at, checksOf(t, checksYAML)); !wrote || err != nil {
		t.Errorf("looser settings by update: %v %v", wrote, err)
	}
	if c, _ := AcceptedChecks(home, main); !sameChecks(c, checksOf(t, checksYAML)) {
		t.Error("update's settings were not recorded")
	}
	// A checks field Bonsai does not read: the record is damaged, and not written over.
	for _, bad := range []string{`7`, `{"ladder_floor": [], "ladder": []}`, `{"ladder_floor": ["x"], "ladder": [], "ratchets": {}, ` +
		`"ci_marked_tests": [], "protected": [], "person_only": [], "never_edit": [], "documents": {}}`} {
		body := `{"path": "/p", "ids": [], "base_branch": "main", "checks": ` + bad + `}`
		if err := os.WriteFile(key, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadMachineRecord(home, main); err == nil || !strings.Contains(err.Error(), "its checks") {
			t.Errorf("checks %s: %v", bad, err)
		}
		if wrote, err := RecordCheckout(home, main, id, at); wrote || err == nil {
			t.Errorf("checks %s: written over", bad)
		}
	}
}

// Two writers at once: a write whose record changed between its read and its rename reads it again and decides again,
// so the other writer's id is kept.
func TestRewriteReadsAgain(t *testing.T) {
	home, main := t.TempDir(), t.TempDir()
	const id = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"
	at := time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC)
	if err := os.WriteFile(filepath.Join(main, ConfigFile), []byte(checksYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordCheckout(home, main, id, at); err != nil {
		t.Fatal(err)
	}
	dir, _ := MachineDir(home, main)
	key := filepath.Join(dir, MachineRecordFile)
	stricter := checksOf(t, strings.Replace(checksYAML, "  go_tests: 2\n", "  go_tests: 3\n", 1))
	tries := 0
	wrote, err := rewrite(key, func(m *MachineRecord) (*MachineRecord, error) {
		tries++
		if tries == 1 {
			// Another command records a new id after this one read the record.
			if _, err := RecordCheckout(home, main, "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", at.Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
		}
		m.Checks = stricter
		return m, nil
	})
	if !wrote || err != nil || tries != 2 {
		t.Fatalf("wrote %v, %v, tries %d", wrote, err, tries)
	}
	rec, _ := LoadMachineRecord(home, main)
	if rec.Current() != "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb" || len(rec.IDs) != 2 || !sameChecks(rec.Checks, stricter) {
		t.Errorf("the other writer's id or this one's settings were lost: %+v", rec)
	}
}
