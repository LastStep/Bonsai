package main

// The words settings and labels as step 5.6.0 placed them (plan-5 note 5.6.0, 2): registered with their full tables
// in bonsai --help --json, their sub-words built (settings' by step 5.6.1, TestSettingsCommand; labels' by step 5.6.2,
// labels_test.go), a sub-word still marked Later refused as not-built; each refusal's --json document held to its format; and personOnly, the
// agent's-session refusal the sub-words call, with the stamped build's confinement to its scratch root (agent.go).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestSettingsAndLabelsArePlaced(t *testing.T) {
	want := map[string][]string{"settings": {"settings show", "settings set"}, "labels": {"labels attach", "labels detach"}}
	later := map[string]string{"settings": "", "labels": ""} // both built: settings in step 5.6.1, labels in step 5.6.2
	for _, name := range []string{"settings", "labels"} {
		w := words[name]
		if w == nil {
			t.Fatalf("%s is not registered", name)
		}
		if len(w.Subs) != 2 {
			t.Fatalf("%s has sub-words %d, want 2", name, len(w.Subs))
		}
		for i, s := range w.Subs {
			if s.Name != want[name][i] || s.Later != later[name] || s.Refused == nil || s.Run == nil || s.flag("--json") == nil ||
				len(s.Examples) == 0 {
				t.Errorf("%s: %+v", s.Name, s)
			}
			for _, code := range []int{0, 2, 3, 4} {
				if !s.exits(code) {
					t.Errorf("%s has no exit %d", s.Name, code)
				}
			}
			if later[name] == "" {
				continue // built: its own tests (TestSettingsCommand, labels_test.go) run it
			}
			// Refused as not built, before anything else, the --json given or not.
			var stdout, stderr bytes.Buffer
			args := append(strings.Fields(s.Name), "x=y", "--json")
			if code := run(args, &stdout, &stderr); code != exitInput || !strings.Contains(stderr.String(), "[not-built]") ||
				!strings.Contains(stderr.String(), strings.TrimPrefix(later[name], "step ")) {
				t.Errorf("bonsai %s: exit %d\n%s%s", strings.Join(args, " "), code, stdout.String(), stderr.String())
			}
		}
	}
	for _, s := range []*Word{settingsShowWord, settingsSetWord} {
		if s.flag("--machine") == nil {
			t.Errorf("%s takes no --machine", s.Name)
		}
	}
	// bonsai --help --json lists both words and their four sub-words, each sub-word's full table and its step.
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--help", "--json"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("--help --json: exit %d: %s", code, stderr.String())
	}
	var doc struct {
		Words []struct {
			Name  string  `json:"name"`
			Later *string `json:"later"`
			Flags []struct {
				Name string `json:"name"`
			} `json:"flags"`
			Exits []struct {
				Code int `json:"code"`
			} `json:"exits"`
			Subs []string `json:"subs"`
		} `json:"words"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, w := range doc.Words {
		seen[w.Name] = true
		switch w.Name {
		case "settings", "labels":
			if strings.Join(w.Subs, ",") != strings.Join(want[w.Name], ",") {
				t.Errorf("--help --json: %s's subs %v", w.Name, w.Subs)
			}
		case "settings show", "settings set", "labels attach", "labels detach":
			if w.Later != nil || len(w.Flags) == 0 || len(w.Exits) != 4 { // all four built (steps 5.6.1 and 5.6.2)
				t.Errorf("--help --json: %s: later %v, %d flags, %d exits", w.Name, w.Later, len(w.Flags), len(w.Exits))
			}
		}
	}
	for _, n := range []string{"settings", "settings show", "settings set", "labels", "labels attach", "labels detach"} {
		if !seen[n] {
			t.Errorf("--help --json does not list %s", n)
		}
	}
}

// TestSettingsAndLabelsRefusalDocuments: each sub-word's --json refusal is its format's document, held to the schema.
func TestSettingsAndLabelsRefusalDocuments(t *testing.T) {
	e := personOnlyAt("", "", "/srv/bonsai-home", func(string) string { return "1" }, "bonsai settings set", "it changes this machine's settings", "bonsai settings set a=b")
	for _, c := range []struct {
		word  *Word
		given []given
		has   string
	}{
		{settingsShowWord, []given{{name: "--machine"}}, `"scope": "machine"`},
		{settingsSetWord, nil, `"scope": "workspace"`},
		{labelsAttachWord, nil, `"command": "attach"`},
		{labelsDetachWord, nil, `"command": "detach"`},
	} {
		out, err := c.word.Refused(&call{word: c.word, given: c.given}, e).Encode()
		if err != nil || !strings.Contains(string(out), c.has) || !strings.Contains(string(out), `"code": "person-command"`) {
			t.Errorf("%s: %v\n%s", c.word.Name, err, out)
		}
	}
	if _, ok := format.ErrorWord("person-command"); !ok {
		t.Error("person-command is not an error word")
	}
}

// TestPersonOnly: in a plain build the variable alone decides, and nothing skips it; a stamped build is free only
// while the home lies inside its scratch root, and refused otherwise whether or not the variable is set.
func TestPersonOnly(t *testing.T) {
	const shown, what, line = "bonsai settings set", "it changes this machine's settings", "bonsai settings set status_writes=agents"
	agent := func(string) string { return "1" }
	person := func(string) string { return "" }
	root := t.TempDir()
	outside := t.TempDir()
	place := filepath.Join(root, "bonsai")
	cases := []struct {
		name              string
		place, root, home string
		getenv            func(string) string
		code              string // "" for no refusal
	}{
		{"plain, a person", "", "", outside, person, ""},
		{"plain, an agent", "", "", outside, agent, "person-command"},
		{"plain, an agent, the home in a scratch folder", "", "", filepath.Join(root, "home"), agent, "person-command"},
		{"stamped, the home inside, an agent", place, root, filepath.Join(root, "home"), agent, ""},
		{"stamped, the home inside, a person", place, root, filepath.Join(root, "a", "home"), person, ""},
		{"stamped, the home the root itself", place, root, root, agent, ""},
		{"stamped, the home outside, a person", place, root, outside, person, "test-build"},
		{"stamped, the home outside, an agent", place, root, outside, agent, "test-build"},
		{"stamped, a sibling of the root", place, root, root + "-home", person, "test-build"},
		{"stamped, out through ..", place, root, filepath.Join(root, "..", filepath.Base(outside)), person, "test-build"},
		{"stamped, no home found", place, root, "", person, "test-build"},
		{"stamped with the root only", "", root, filepath.Join(root, "home"), person, "test-build"},
		{"stamped with the place only", place, "", filepath.Join(root, "home"), person, "test-build"},
		{"stamped, a relative root", place, "scratch", filepath.Join("scratch", "home"), person, "test-build"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := personOnlyAt(c.place, c.root, c.home, c.getenv, shown, what, line)
			if c.code == "" {
				if e != nil {
					t.Fatalf("refused: %s: %s", e.Code, e.What)
				}
				return
			}
			if e == nil {
				t.Fatalf("not refused, want %s", c.code)
			}
			o := e.Object()
			if o.Code != c.code || e.Exit != engine.ExitState {
				t.Errorf("refused %s exit %d, want %s exit 4", o.Code, e.Exit, c.code)
			}
			if !strings.Contains(e.What, "nothing was written") {
				t.Errorf("its sentence: %s", e.What)
			}
			asciiOnly(t, "the refusal", e.What+e.Next)
			switch c.code {
			case "person-command":
				if o.Next.Who != "person" || o.Next.Do != line ||
					!strings.Contains(e.What, "type it in a terminal of your own, not inside Claude (its ! lines count as an agent's session)") {
					t.Errorf("the person's refusal: %+v", o)
				}
			case "test-build":
				if o.Next.Who != "agent" || !strings.Contains(o.Next.Do, "BONSAI_HOME") {
					t.Errorf("the test build's refusal: %+v", o)
				}
			}
		})
	}
	// personOnly reads the variable as it is now: this test binary is a plain build.
	if p, r := engine.TestBuild(); p != "" || r != "" {
		t.Skip("this test binary is stamped") // never so under go test; the cases above cover a stamped build
	}
	t.Setenv("BONSAI_HOME", t.TempDir())
	t.Setenv(guard.ChildSessionEnv, "1")
	if e := personOnly(shown, what, line); e == nil || e.Code != "person-command" {
		t.Errorf("an agent's session: %+v", e)
	}
	testpack.AsPerson(t)
	if e := personOnly(shown, what, line); e != nil {
		t.Errorf("a person's: %+v", e)
	}
}

// TestPersonOnlyLinkedRoot: a scratch root reached through a link, its home not made yet, is the home inside it (CI's
// Windows runner reaches its temporary folder by an 8.3 short name, which resolves the same way); a link inside the
// root to a folder outside it, its target not made yet, is outside.
func TestPersonOnlyLinkedRoot(t *testing.T) {
	const shown, what, line = "bonsai settings set", "it changes this machine's settings", "bonsai settings set status_writes=agents"
	agent := func(string) string { return "1" }
	real, outside := t.TempDir(), t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("no link can be made here (on Windows it needs a privilege): %v", err)
	}
	place := filepath.Join(real, "bonsai")
	for _, home := range []string{filepath.Join(link, "home"), filepath.Join(link, "a", "home"), filepath.Join(real, "home")} {
		if e := personOnlyAt(place, link, home, agent, shown, what, line); e != nil {
			t.Errorf("the root %s, the home %s: refused: %s: %s", link, home, e.Code, e.What)
		}
		if e := personOnlyAt(place, real, home, agent, shown, what, line); e != nil {
			t.Errorf("the root %s, the home %s: refused: %s: %s", real, home, e.Code, e.What)
		}
	}
	away := filepath.Join(real, "away")
	if err := os.Symlink(filepath.Join(outside, "not-made"), away); err != nil {
		t.Fatal(err)
	}
	for _, home := range []string{away, filepath.Join(away, "home")} {
		if e := personOnlyAt(place, link, home, agent, shown, what, line); e == nil || e.Code != "test-build" {
			t.Errorf("the home %s, through a link to outside the root: %+v, want test-build", home, e)
		}
	}
}

// TestCallLine: the line a person is handed is the command again, its arguments and flags quoted for a shell.
func TestCallLine(t *testing.T) {
	c := &call{word: settingsSetWord, rest: []string{"status_writes=command", "status_command=studio task move"},
		given: []given{{name: "--machine"}, {name: "--json"}}}
	if got, want := c.line(), "bonsai settings set status_writes=command 'status_command=studio task move' --machine --json"; got != want {
		t.Errorf("line %q, want %q", got, want)
	}
}

// TestSettingsCommand: settings show and set on a t.TempDir() home and project, as a person (the variable cleared):
// every key, a bad value and an unknown key refused with nothing written, all or nothing, the byte-stable file, a
// value already in force writing nothing, the workspace's file read from a worktree too; and settings set refused
// in an agent's session with nothing written, settings show not.
func TestSettingsCommand(t *testing.T) {
	c := newCLI(t)
	home := os.Getenv("BONSAI_HOME")
	asPerson(c)
	// Not linked: no bonsai.yaml yet. --machine needs none.
	if code, out, _ := c.run("", "settings", "show", "--json"); code != engine.ExitState || !strings.Contains(out, `"code": "not-linked"`) ||
		!strings.Contains(out, `"file": null`) {
		t.Errorf("not linked: exit %d\n%s", code, out)
	}
	if code, out, errOut := c.run("", "settings", "show", "--machine"); code != exitOK || !strings.Contains(out, "cache_keep_days") ||
		!strings.Contains(out, "the default") {
		t.Errorf("--machine: exit %d\n%s%s", code, out, errOut)
	}
	c.write(workspace.ConfigFile, "name: demo\n")
	settings := func(args ...string) (int, *format.Settings) {
		t.Helper()
		code, out, errOut := c.run("", append(append([]string{"settings"}, args...), "--json")...)
		doc, err := format.ReadSettings([]byte(out))
		if err != nil {
			t.Fatalf("settings %v: exit %d: %v\n%s%s", args, code, err, out, errOut)
		}
		return code, doc
	}
	key, _ := workspace.MachineKey(c.root)
	file := filepath.ToSlash(filepath.Join(home, "workspaces", key, "settings.json"))
	code, doc := settings("show")
	if code != exitOK || doc.Scope != "workspace" || doc.File == nil || *doc.File != file || len(doc.Settings) != 2 ||
		doc.Settings[0].Value != "agents" || doc.Settings[0].From != "default" || doc.Settings[1].Value != nil || len(doc.Changed) != 0 {
		t.Fatalf("show: exit %d %+v", code, doc)
	}
	for _, bad := range [][]string{
		{"set", "status_writes=sometimes"},
		{"set", "status_writes=command"},
		{"set", "status_writes=command", "status_command=move; rm -rf x"},
		{"set", "status_writes=command", "status_command=move", "nothing=1"},
		{"set", "cache_keep_days=30"},
		{"set", "cache_keep_days=0", "--machine"},
		{"set", "cache_keep_days=3651", "--machine"},
		{"set", "status_writes=agents", "--machine"},
		{"set", "novalue"},
		{"set"},
	} {
		code, doc := settings(bad...)
		if code != exitInput || doc.Error == nil || doc.Error.Code != "bad-value" || len(doc.Settings) != 0 {
			t.Errorf("%v: exit %d %+v", bad, code, doc.Error)
		}
	}
	if _, err := os.Stat(filepath.FromSlash(file)); err == nil {
		t.Fatalf("a refused set wrote %s", file)
	}
	if _, err := os.Stat(filepath.Join(home, "settings.json")); err == nil {
		t.Fatal("a refused set wrote the home's settings.json")
	}
	code, doc = settings("set", "status_writes=command", "status_command=studio task move")
	if code != exitOK || strings.Join(doc.Changed, ",") != "status_writes,status_command" || doc.Settings[1].Value != "studio task move" ||
		doc.Settings[1].From != "file" {
		t.Fatalf("set: exit %d %+v", code, doc)
	}
	if raw, _ := os.ReadFile(filepath.FromSlash(file)); string(raw) != "{\n  \"status_writes\": \"command\",\n  \"status_command\": \"studio task move\"\n}\n" {
		t.Errorf("the file:\n%s", raw)
	}
	// Again: nothing changed, nothing written.
	if code, doc = settings("set", "status_command=studio task move"); code != exitOK || len(doc.Changed) != 0 {
		t.Errorf("again: exit %d %+v", code, doc)
	}
	// The text, for a person.
	if code, out, _ := c.run("", "settings", "show"); code != exitOK || !strings.Contains(out, "studio task move") ||
		!strings.Contains(out, "from the file") || !strings.Contains(out, file) {
		t.Errorf("show's text: exit %d\n%s", code, out)
	}
	if code, out, _ := c.run("", "settings", "set", "status_writes=agents"); code != exitOK ||
		!strings.Contains(out, "written: status_writes, status_command") {
		t.Errorf("set's text: exit %d\n%s", code, out)
	}
	// The home's own: cache_keep_days.
	if code, doc = settings("set", "cache_keep_days=90", "--machine"); code != exitOK || doc.Scope != "machine" ||
		strings.Join(doc.Changed, ",") != "cache_keep_days" || fmt.Sprint(doc.Settings[0].Value) != "90" {
		t.Errorf("--machine: exit %d %+v", code, doc)
	}
	// A worktree reads the main checkout's machine folder.
	testpack.Git(t, c.root, "add", "-A")
	testpack.Git(t, c.root, "commit", "-q", "-m", "linked")
	wt := filepath.Join(c.tmp, "wt")
	testpack.Git(t, c.root, "worktree", "add", "-q", wt)
	t.Chdir(wt)
	if code, doc = settings("show"); code != exitOK || doc.File == nil || *doc.File != file {
		t.Errorf("from a worktree: exit %d %+v", code, doc)
	}
	t.Chdir(c.root)
	// A file outside its rule: read-failed, exit 3, naming the file and the key.
	if err := os.WriteFile(filepath.FromSlash(file), []byte(`{"status_writes": "command", "status_command": "move $(id)"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, doc = settings("show"); code != exitRuntime || doc.Error == nil || doc.Error.Code != "read-failed" ||
		!strings.Contains(doc.Error.Message, file) || !strings.Contains(doc.Error.Message, "status_command") {
		t.Errorf("a forged file: exit %d %+v", code, doc.Error)
	}
	// An agent's session: settings set refused, exit 4, nothing written; settings show still runs.
	asAgent(c)
	before, _ := os.ReadFile(filepath.Join(home, "settings.json"))
	code, doc = settings("set", "cache_keep_days=5", "--machine")
	if code != engine.ExitState || doc.Error == nil || doc.Error.Code != "person-command" || doc.Error.Next.Who != "person" ||
		doc.Error.Next.Do != "bonsai settings set cache_keep_days=5 --machine --json" {
		t.Errorf("an agent's set: exit %d %+v", code, doc.Error)
	}
	if after, _ := os.ReadFile(filepath.Join(home, "settings.json")); string(after) != string(before) {
		t.Errorf("an agent's set wrote:\n%s", after)
	}
	if code, doc = settings("show", "--machine"); code != exitOK || fmt.Sprint(doc.Settings[0].Value) != "90" {
		t.Errorf("an agent's show: exit %d %+v", code, doc)
	}
}
