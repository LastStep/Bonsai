package main

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// helpOut runs bonsai --help --json and reads the document with the format's reader.
func helpOut(t *testing.T, args ...string) (*format.Help, string) {
	t.Helper()
	code, out, errOut := runArgs(args...)
	if code != 0 || errOut != "" {
		t.Fatalf("%v: exit %d, stderr %q", args, code, errOut)
	}
	h, err := format.ReadHelp([]byte(out))
	if err != nil {
		t.Fatalf("%v: the document is not bonsai.help/1: %v", args, err)
	}
	return h, out
}

// The document fits its schema, checked here against the embedded file itself and not only by the writer; it is
// ASCII, and the same in either order of the two flags.
func TestHelpJSONFitsItsSchema(t *testing.T) {
	_, out := helpOut(t, "--help", "--json")
	asciiOnly(t, "bonsai --help --json", out)
	doc, err := schema.Decode([]byte(out))
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range schema.Validate(format.MustLookup("help").Schema(), doc) {
		t.Errorf("bonsai --help --json: %s", problem)
	}
	for _, args := range [][]string{{"--json", "--help"}, {"-h", "--json"}, {"--json", "-h"}} {
		if _, again := helpOut(t, args...); again != out {
			t.Errorf("%v: another document than --help --json", args)
		}
	}
}

// The document lists every word, every sub-word, every flag (with its value, whether it may repeat, and its step when
// it is not built), every exit code with its meaning, and every example the registry has, and every error word and
// exit code the shared tables have; nothing else.
func TestHelpJSONHasEveryWord(t *testing.T) {
	h, _ := helpOut(t, "--help", "--json")
	byName := map[string]format.HelpWord{}
	var order []string
	for _, w := range h.Words {
		if _, dup := byName[w.Name]; dup {
			t.Errorf("the word %s is listed twice", w.Name)
		}
		byName[w.Name] = w
		order = append(order, w.Name)
	}
	var want []string
	for _, w := range helpEntries() {
		want = append(want, w.Name)
		got, ok := byName[w.Name]
		if !ok {
			t.Errorf("the word %s is not in the document", w.Name)
			continue
		}
		if got.Summary != w.Summary || got.Title != w.Title || got.TakesJSON != (w.Refused != nil) ||
			got.Usage != strings.TrimSpace("bonsai "+w.Name+" "+w.Args) || (got.Later != nil) != (w.Later != "") {
			t.Errorf("%s: its summary, title, usage, --json or step differ from the registry's", w.Name)
		}
		if len(got.Flags) != len(w.Flags) || len(got.Exits) != len(w.Exits) || len(got.Examples) != len(w.Examples) || len(got.Subs) != len(w.Subs) {
			t.Errorf("%s: %d flags, %d exits, %d examples, %d subs in the document; the registry has %d, %d, %d, %d", w.Name,
				len(got.Flags), len(got.Exits), len(got.Examples), len(got.Subs), len(w.Flags), len(w.Exits), len(w.Examples), len(w.Subs))
			continue
		}
		for i, f := range w.Flags {
			g := got.Flags[i]
			if g.Name != f.Name || (g.Value != nil) != (f.Value != "") || (g.Value != nil && *g.Value != f.Value) || g.Many != f.Many ||
				!strings.HasPrefix(g.Help, f.Help) || (g.Later != nil) != (f.Later != "") || (g.Later != nil && *g.Later != f.Later) {
				t.Errorf("%s %s: differs from the registry's flag", w.Name, f.Name)
			}
		}
		for i, e := range w.Exits {
			if got.Exits[i].Code != int64(e.Code) || got.Exits[i].Means != e.Means {
				t.Errorf("%s: exit %d differs from the registry's", w.Name, e.Code)
			}
		}
		for i, ex := range w.Examples {
			if got.Examples[i] != ex {
				t.Errorf("%s: example %q differs", w.Name, ex)
			}
		}
		for i, s := range w.Subs {
			if got.Subs[i] != s.Name {
				t.Errorf("%s: sub-word %q differs", w.Name, s.Name)
			}
		}
		if !w.takesJSON() && len(w.Subs) == 0 && got.TakesJSON {
			t.Errorf("%s: listed as taking --json", w.Name)
		}
	}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("the words are %v, the registry's %v", order, want)
	}
	// hook takes no --json, and the document still describes it.
	if hook, ok := byName["hook"]; !ok || hook.TakesJSON || byName["hook guard"].TakesJSON {
		t.Errorf("hook and hook guard must be in the document, taking no --json")
	}
	if len(h.ErrorWords) != len(format.ErrorWords) {
		t.Fatalf("%d error words in the document, %d in format.ErrorWords", len(h.ErrorWords), len(format.ErrorWords))
	}
	for i, e := range format.ErrorWords {
		g := h.ErrorWords[i]
		if g.Code != e.Word || g.Means != e.Means || g.Who != e.Who || g.Who == "" || g.Means == "" {
			t.Errorf("error word %s differs from format.ErrorWords", e.Word)
		}
	}
	if len(h.ExitCodes) != len(format.ExitCodes) {
		t.Errorf("%d exit codes in the document, %d in format.ExitCodes", len(h.ExitCodes), len(format.ExitCodes))
	}
	if h.Bonsai != version {
		t.Errorf("bonsai is %q, the build's version is %q", h.Bonsai, version)
	}
}

// Every word's flag in the document appears in its human --help too, so the two are one table: a flag in the human
// help is in the document, and a flag the document lists is one the word takes.
func TestHelpJSONAgreesWithHumanHelp(t *testing.T) {
	h, _ := helpOut(t, "--help", "--json")
	for _, w := range h.Words {
		if w.Later != nil {
			continue
		}
		reg := words[strings.Fields(w.Name)[0]]
		if len(strings.Fields(w.Name)) > 1 {
			for _, s := range reg.Subs {
				if s.Name == w.Name {
					reg = s
				}
			}
		}
		help := reg.Help()
		for _, f := range w.Flags {
			if f.Later != nil {
				continue
			}
			shown := f.Name
			if f.Value != nil {
				shown += " " + *f.Value
			}
			if !strings.Contains(help, "  "+shown+"  ") {
				t.Errorf("%s: the document lists %s, the human help does not", w.Name, shown)
			}
		}
	}
}

// A word's own --help stays the human text with --json beside it, and hook refuses --json as before.
func TestWordHelpWithJSONStaysText(t *testing.T) {
	code, out, _ := runArgs("check", "--help", "--json")
	if code != 0 || out != words["check"].Help() {
		t.Errorf("check --help --json: exit %d, and not the human help", code)
	}
	if code, _, _ := runArgs("--help", "--json", "check"); code != 2 {
		t.Errorf("--help --json check: exit %d, want 2 (a refusal)", code)
	}
}
