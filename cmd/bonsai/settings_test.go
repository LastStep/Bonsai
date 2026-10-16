package main

// The words settings and labels as step 5.6.0 places them (plan-5 note 5.6.0, 2): registered with their sub-words
// marked Later and refused as not-built, their full tables in bonsai --help --json, each refusal's --json document
// held to its format; and personOnly, the agent's-session refusal the sub-words call, with the stamped build's
// confinement to its scratch root (agent.go).

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
)

func TestSettingsAndLabelsArePlaced(t *testing.T) {
	want := map[string][]string{"settings": {"settings show", "settings set"}, "labels": {"labels attach", "labels detach"}}
	later := map[string]string{"settings": "step 5.6.1", "labels": "step 5.6.2"}
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
			if w.Later == nil || !strings.HasPrefix(*w.Later, "step 5.6.") || len(w.Flags) == 0 || len(w.Exits) != 4 {
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
	t.Setenv(guard.ChildSessionEnv, "")
	if e := personOnly(shown, what, line); e != nil {
		t.Errorf("a person's: %+v", e)
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
