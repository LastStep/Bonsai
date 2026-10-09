package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
)

// machineFile writes a file into the workspace's machine folder.
func machineFile(t *testing.T, home, main, rel, content string) string {
	t.Helper()
	dir, err := MachineDir(home, main)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// This machine's settings: agents with no file; command with its command; a file Bonsai does not read refused,
// naming the file and the step.
func TestMachineSettings(t *testing.T) {
	home, main := t.TempDir(), t.TempDir()
	if s, err := LoadMachineSettings(home, main); err != nil || s.StatusWrites != "agents" || s.StatusCommand != "" {
		t.Errorf("no settings file: %+v %v", s, err)
	}
	for _, c := range []struct {
		content, writes, command, refused string
	}{
		{`{}`, "agents", "", ""},
		{`{"status_writes": "agents", "status_command": "studio move"}`, "agents", "", ""},
		{`{"status_writes": "command", "status_command": "studio move", "later": 1}`, "command", "studio move", ""},
		{"\xef\xbb\xbf" + `{"status_writes": null}`, "agents", "", ""},
		{`{"status_writes": "command"}`, "", "", "names no command"},
		{`{"status_writes": "studio"}`, "", "", "not agents or command"},
		{`{"status_writes": "agents", "status_writes": "command"}`, "", "", "not a JSON object"},
		{`[]`, "", "", "not a JSON object"},
	} {
		machineFile(t, home, main, MachineSettingsFile, c.content)
		s, err := LoadMachineSettings(home, main)
		if c.refused != "" {
			if err == nil || !strings.Contains(err.Error(), c.refused) || !strings.Contains(err.Error(), "next: fix ") {
				t.Errorf("%s: want refused (%s), got %+v %v", c.content, c.refused, s, err)
			}
			continue
		}
		if err != nil || s.StatusWrites != c.writes || s.StatusCommand != c.command {
			t.Errorf("%s: %+v %v", c.content, s, err)
		}
	}
}

// Labels in force: each pack's from the lock's declares, then those attached on this machine by file name; an
// attached file that is not definitions, not named for its namespace, or takes a pack's (or Bonsai's) namespace is
// left out with a problem.
func TestLabelsInForce(t *testing.T) {
	home, main := t.TempDir(), t.TempDir()
	def := func(ns, name string) string {
		return "format: bonsai.labels/1\nnamespace: " + ns + "\nversion: 2\nlabels:\n  - name: " + ns + "." + name +
			"\n    kind: number\n    values: []\n    items: null\n    pattern: null\n    max: null\n    kinds: [\"task\"]\n" +
			"    set_by: outside\n    grants: false\n    description: \"A test label.\"\n"
	}
	machineFile(t, home, main, "labels/studio.yaml", def("studio", "cost"))
	machineFile(t, home, main, "labels/aaa.yaml", def("aaa", "x"))
	machineFile(t, home, main, "labels/wrong.yaml", def("other", "x"))
	machineFile(t, home, main, "labels/base.yaml", def("base", "x"))
	machineFile(t, home, main, "labels/bonsai.yaml", def("bonsai", "x"))
	machineFile(t, home, main, "labels/broken.yaml", "format: bonsai.labels/1\nnamespace: Broken\n")
	machineFile(t, home, main, "labels/readme.txt", "not a definitions file\n")
	labels, err := format.ReadLabels([]byte(def("base", "owner")))
	if err != nil {
		t.Fatal(err)
	}
	lock := lockWith(t, &format.Declares{Labels: labels})
	sets, problems := LabelsInForce(lock, home, main)
	var got []string
	for _, s := range sets {
		got = append(got, s.Namespace+"/"+s.From+"/"+s.Labels[0].Name)
	}
	if strings.Join(got, " ") != "base/base/base.owner aaa/machine/aaa.x studio/machine/studio.cost" {
		t.Errorf("labels in force %v", got)
	}
	var whys []string
	for _, p := range problems {
		whys = append(whys, p.Error())
	}
	all := strings.Join(whys, "\n")
	for _, want := range []string{"base.yaml: attaches the namespace \"base\"", "bonsai.yaml: attaches the namespace \"bonsai\"",
		"broken.yaml: is not label definitions", "wrong.yaml: holds the namespace \"other\", not its file's name"} {
		if !strings.Contains(all, want) {
			t.Errorf("no problem %q in\n%s", want, all)
		}
	}
	if len(problems) != 4 {
		t.Errorf("%d problems:\n%s", len(problems), all)
	}
	if sets, problems := LabelsInForce(nil, "", ""); len(sets) != 0 || len(problems) != 0 {
		t.Errorf("nothing: %v %v", sets, problems)
	}
}
