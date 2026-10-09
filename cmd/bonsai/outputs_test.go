package main

// The commands' outputs (plan-5, piece 5.1.4b): every word's table of flags and exit codes against its help and its
// command line; every refusal's error object, with a word from format.ErrorWords and who takes the next step, in a
// --json document that fits its schema; and the write that stops part-way, result failed.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// Every test of this package runs its words through run, which shows each refusal and each exit code to these: a
// refusal whose word is not in format.ErrorWords, or an exit code not in the word's table, stops the test run.
func init() {
	noteError = func(w *Word, e *engine.Error) {
		if _, ok := format.ErrorWord(e.Code); !ok {
			panic(fmt.Sprintf("bonsai %s refused with the word %q, which is not in format.ErrorWords: %s", w.Name, e.Code, e.What))
		}
		if e.Who != "" && e.Who != "agent" && e.Who != "person" {
			panic(fmt.Sprintf("bonsai %s refused naming %q as who takes the next step", w.Name, e.Who))
		}
	}
	noteExit = func(w *Word, code int) {
		if !w.exits(code) {
			panic(fmt.Sprintf("bonsai %s exited %d, which its table of exit codes does not list", w.Name, code))
		}
	}
}

// outputOf is the format each word's --json prints, the one its Refused gives.
func outputOf(t *testing.T, w *Word) string {
	t.Helper()
	b, err := w.Refused(&call{word: w}, &engine.Error{Code: "unexpected", Exit: 3, What: "x", Next: "y"}).Encode()
	if err != nil {
		t.Fatalf("%s's refusal document: %v", w.Name, err)
	}
	doc, err := schema.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	f := doc.(schema.Object).String("format")
	if f == "" {
		return "error"
	}
	name, _, _ := strings.Cut(strings.TrimPrefix(f, "bonsai."), "/")
	return name
}

// fits decodes a --json output and holds it to its format's schema as a writer writes it (every field, in order).
func fits(t *testing.T, out, name string) schema.Object {
	t.Helper()
	doc, err := schema.Decode([]byte(out))
	if err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	s := format.MustLookup(name).Schema()
	msgs := append(schema.Validate(s, doc), schema.CheckOrder(s, doc)...)
	if len(msgs) > 0 {
		t.Fatalf("the output does not fit bonsai.%s: %v\n%s", name, msgs, out)
	}
	return doc.(schema.Object)
}

// errorIn gives a document's error object: the document itself for bonsai.error.
func errorIn(doc schema.Object, name string) schema.Object {
	if name == "error" {
		return doc
	}
	e, _ := doc.Get("error")
	o, _ := e.(schema.Object)
	return o
}

func whoOf(e schema.Object) string {
	n, _ := e.Get("next")
	o, _ := n.(schema.Object)
	return o.String("who")
}

// Each word's help is written from its table: every flag (and the flags not built yet, apart), every exit code
// and an example. The word takes each flag in its table (a flag and its value, then --help, prints the help), refuses
// one not built yet as not-built, and refuses a flag its table does not have as bad-flag.
func TestWordTables(t *testing.T) {
	var all []*Word
	for _, w := range wordList() {
		all = append(all, w)
		for _, s := range w.Subs {
			if s.Later == "" {
				all = append(all, s)
			}
		}
	}
	if len(all) < 6 {
		t.Fatalf("the registry holds %d words", len(all))
	}
	for _, w := range all {
		t.Run(w.Name, func(t *testing.T) {
			line := strings.Fields(w.Name)
			code, out, errOut := runArgs(append(line, "--help")...)
			if code != 0 || errOut != "" {
				t.Fatalf("--help: exit %d, stderr %q", code, errOut)
			}
			if out != w.Help() || !strings.HasPrefix(out, "bonsai "+w.Name+": ") {
				t.Errorf("--help is not the table's help:\n%s", out)
			}
			for _, f := range append(w.Flags, Flag{Name: "--help"}) {
				want := "\n  " + flagShown(f) + "  "
				if f.Later != "" {
					want = flagShown(f) + " (" + f.Later + ")"
				}
				if !strings.Contains(out, want) {
					t.Errorf("the help does not list %s", flagShown(f))
				}
			}
			if len(w.Exits) == 0 {
				t.Errorf("no exit codes")
			}
			last := -1
			for _, e := range w.Exits {
				if !strings.Contains(out, "\n  "+strconv.Itoa(e.Code)+"  ") || e.Means == "" || e.Code <= last {
					t.Errorf("exit code %d is not listed once, in order, with its meaning", e.Code)
				}
				last = e.Code
			}
			if !strings.Contains(out, "\nExample: bonsai "+line[0]) && !strings.Contains(out, "bonsai "+w.Name+" < ") {
				t.Errorf("no example:\n%s", out)
			}
			for _, l := range strings.Split(out, "\n") {
				if len(l) > 120 && !strings.HasPrefix(l, "Example: ") {
					t.Errorf("a help line of %d characters: %q", len(l), l)
				}
			}
			asciiOnly(t, "help", out)
			for _, f := range w.Flags {
				args := append(append([]string{}, line...), f.Name)
				if f.Value != "" {
					args = append(args, "x")
				}
				code, out, errOut := runArgs(append(args, "--help")...)
				if f.Later == "" && (code != 0 || out != w.Help()) {
					t.Errorf("%v --help: exit %d, %q: the word does not take a flag of its table", args, code, errOut)
				}
				if f.Later != "" {
					wantRefusal(t, w, args, 2, "not-built")
				}
			}
			wantRefusal(t, w, append(append([]string{}, line...), "--not-in-the-table"), 2, "bad-flag")
			if w.takesJSON() != (w.flag("--json") != nil) {
				t.Errorf("--json in the table (%v) but a --json document (%v)", w.flag("--json") != nil, w.takesJSON())
			}
		})
	}
	// bonsai --help lists every word, with its usage and summary.
	_, out, _ := runArgs("--help")
	for _, w := range wordList() {
		if !strings.Contains(out, "  bonsai "+w.Name) || !strings.Contains(out, w.Summary+" ("+w.Name+" --help)") {
			t.Errorf("bonsai --help does not list %s", w.Name)
		}
	}
}

// runArgs runs bonsai in this process, as no terminal.
func runArgs(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// wantRefusal runs a refusal with and without --json: the exit code, the error object's word in a document that fits
// the word's format, and the next step in the human lines.
func wantRefusal(t *testing.T, w *Word, args []string, exit int, word string) {
	t.Helper()
	code, out, errOut := runArgs(args...)
	if code != exit || !strings.Contains(out+errOut, "\nnext: ") {
		t.Errorf("%v: exit %d, want %d, and a next step:\n%s%s", args, code, exit, out, errOut)
	}
	if !w.takesJSON() {
		return
	}
	code, out, _ = runArgs(append(args, "--json")...)
	name := outputOf(t, w)
	if e := errorIn(fits(t, out, name), name); code != exit || e.String("code") != word {
		t.Errorf("%v --json: exit %d, error %s, want %d and %s", args, code, schema.Show(e), exit, word)
	}
}

// Every refusal and error a test can reach, each word's: with --json its document fits the word's format, its error
// object names a known word, a message and a next step with who takes it; without --json the next step is named.
func TestRefusalsCarryTheirWord(t *testing.T) {
	c := newCLI(t)
	p := c.pack
	quiet, qc := testpack.QuietPack(t, c.tmp)
	nopack, nc := testpack.Fixture(t, c.tmp, "no-pack", map[string]string{"readme.md": "no bonsai/ here\n"})
	dup, dc := testpack.Fixture(t, c.tmp, "dup-pack", map[string]string{
		"bonsai/pack.yaml":       "format: bonsai.pack/1\nid: dup-pack\nversion: \"0.1.0\"\nfiles:\n  - path: quiet/readme.md\n    from: readme.md\n    kind: pack\nhooks: []\ndeny: []\n",
		"bonsai/files/readme.md": "# Mine\n",
	})
	tagged, tc := testpack.Fixture(t, c.tmp, "tagged",
		map[string]string{"bonsai/pack.yaml": "format: bonsai.pack/1\nid: tagged\nversion: \"0.1.0\"\nfiles: []\nhooks: []\ndeny: []\n"},
		map[string]string{"notes.md": "a second commit at the same version\n"})
	link := []string{"init", "--name", "demo", "--source", p.Source, "--ref", p.A}
	linked := func(c *cli) {
		if code, out, errOut := c.run("", append(link, "--yes", "--allow-exec")...); code != 0 {
			c.t.Fatalf("link: %d\n%s%s", code, out, errOut)
		}
	}
	write := func(rel, body string) func(c *cli) {
		return func(c *cli) {
			path := filepath.Join(c.root, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				c.t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				c.t.Fatal(err)
			}
		}
	}
	then := func(steps ...func(c *cli)) func(c *cli) {
		return func(c *cli) {
			for _, s := range steps {
				s(c)
			}
		}
	}
	plain := 0
	outside := func(c *cli) {
		plain++
		c.root = filepath.Join(c.tmp, "plain-"+strconv.Itoa(plain))
		if err := os.MkdirAll(c.root, 0o755); err != nil {
			c.t.Fatal(err)
		}
		c.t.Chdir(c.root)
	}
	noGit := func(c *cli) { c.t.Setenv("PATH", filepath.Join(c.tmp, "empty-path")) }
	noHome := func(c *cli) {
		c.t.Setenv("BONSAI_HOME", "")
		c.t.Setenv("HOME", "")
		c.t.Setenv("USERPROFILE", "")
	}
	twoPacks := "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n" +
		"  - id: quiet-pack\n    source: \"" + filepath.ToSlash(quiet) + "\"\n    ref: \"" + qc[0] + "\"\n" +
		"  - id: dup-pack\n    source: \"" + filepath.ToSlash(dup) + "\"\n    ref: \"" + dc[0] + "\"\n"

	cases := []struct {
		name string
		prep func(c *cli)
		args []string
		exit int
		word string
		who  string // "" for the word's usual one
	}{
		{"no word", nil, nil, 2, "unknown-command", ""},
		{"a word Bonsai does not have", nil, []string{"unlink"}, 2, "unknown-command", ""},
		{"--version with more", nil, []string{"--version", "now"}, 2, "bad-flag", ""},

		{"init with no values", nil, []string{"init"}, 2, "missing-value", ""},
		{"init, a flag it does not take", nil, []string{"init", "--nope"}, 2, "bad-flag", ""},
		{"init, a flag with no value", nil, []string{"init", "--name"}, 2, "bad-flag", ""},
		{"init, a value given to a switch", nil, []string{"init", "--yes=1"}, 2, "bad-flag", ""},
		{"init, a word left over", nil, []string{"init", "now"}, 2, "bad-flag", ""},
		{"init, a flag given twice", nil, []string{"init", "--name", "a", "--name", "b"}, 2, "bad-flag", ""},
		{"init, a bad name", nil, []string{"init", "--name", "Bad_Name", "--source", p.Source, "--ref", p.A}, 2, "bad-value", ""},
		{"init, a tag not there", nil, []string{"init", "--name", "demo", "--source", p.Source, "--ref", "no-such-tag"}, 2, "ref-not-found", ""},
		{"init, a source not there", nil, []string{"init", "--name", "demo", "--source", filepath.ToSlash(filepath.Join(c.tmp, "nowhere.git")), "--ref", "v1"}, 3, "fetch-failed", ""},
		{"init, not a pack", nil, []string{"init", "--name", "demo", "--source", nopack, "--ref", nc[0]}, 2, "bad-pack", ""},
		{"init, two packs write one path", write("bonsai.yaml", twoPacks), []string{"init"}, 2, "packs-overlap", ""},
		{"init in a 0.4.3 workspace", write(".bonsai.yaml", "agents: {}\n"), link, 4, "old-workspace", ""},
		{"init, values that differ", linked, []string{"init", "--name", "other"}, 2, "values-differ", ""},
		{"init outside a checkout", outside, link, 4, "not-a-checkout", ""},
		{"init --new-id in a worktree", func(c *cli) {
			linked(c)
			testpack.Git(c.t, c.root, "add", "-A")
			testpack.Git(c.t, c.root, "commit", "-q", "-m", "link")
			wt := c.root + "-wt"
			testpack.Git(c.t, c.root, "worktree", "add", "-q", wt)
			c.root = wt
			c.t.Chdir(wt)
		}, []string{"init", "--new-id"}, 4, "not-main-checkout", "person"},
		{"init with no --yes", nil, []string{"init", "--name", "demo", "--source", quiet, "--ref", qc[0]}, 4, "needs-yes", ""},
		{"init, code without --allow-exec", nil, append(append([]string{}, link...), "--yes"), 4, "needs-allow-exec", ""},
		{"init with no git", noGit, link, 3, "git-missing", ""},
		{"init with no home", noHome, link, 3, "bad-home", ""},

		{"update, not linked", nil, []string{"update"}, 4, "not-linked", ""},
		{"update, no lock", then(linked, func(c *cli) { _ = os.Remove(filepath.Join(c.root, ".bonsai", "lock.json")) }), []string{"update"}, 4, "no-lock", ""},
		{"update, a lock Bonsai does not read", then(linked, write(".bonsai/lock.json", "{")), []string{"update"}, 4, "bad-lock", ""},
		{"update, a bonsai.yaml Bonsai does not read", write("bonsai.yaml", "format: bonsai.workspace/1\nid: 0755\n"), []string{"update"}, 2, "bad-config", ""},
		{"update, a pack taken out", then(linked, write("bonsai.yaml", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\n")),
			[]string{"update"}, 4, "not-built", "person"},
		{"update, a settings file that is not an object", then(linked, write(".claude/settings.json", "[]")), []string{"update"}, 2, "bad-file", ""},
		{"update, a broken block", then(linked, write("CLAUDE.md", "<!-- bonsai:block start -->\n<!-- bonsai:block start -->\n")), []string{"update"}, 2, "bad-file", ""},
		{"update, conflicts", then(linked, write("demo/guide.md", "mine\n"), func(c *cli) { testpack.SetRef(c.t, c.root, p.A, p.B) }),
			[]string{"update", "--yes"}, 5, "conflicts", ""},
		{"update --keep of no conflict", linked, []string{"update", "--keep", "demo/start.md"}, 2, "bad-value", ""},
		{"update, a moved tag", func(c *cli) {
			testpack.Tag(c.t, tagged, "v0.1.0", tc[0])
			if code, out, errOut := c.run("", "init", "--name", "demo", "--source", tagged, "--ref", "v0.1.0", "--yes"); code != 0 {
				c.t.Fatalf("link: %d\n%s%s", code, out, errOut)
			}
			testpack.Tag(c.t, tagged, "v0.1.0", tc[1])
		}, []string{"update"}, 4, "tag-moved", ""},
		{"update, a flag it does not take", nil, []string{"update", "--name", "x"}, 2, "bad-flag", ""},

		{"check, not linked", nil, []string{"check"}, 4, "not-linked", ""},
		{"check outside a checkout", outside, []string{"check"}, 4, "not-a-checkout", ""},
		{"check with no git", noGit, []string{"check"}, 3, "git-missing", ""},
		{"check with no home", noHome, []string{"check"}, 3, "bad-home", ""},
		{"check --schema, a format not there", nil, []string{"check", "--schema", "bonsai.nope"}, 2, "unknown-format", ""},
		{"check --schema with no name", nil, []string{"check", "--schema"}, 2, "bad-flag", ""},
		{"check --write", nil, []string{"check", "--write"}, 2, "not-built", ""},
		{"check --pack", nil, []string{"check", "--pack", "somewhere"}, 2, "not-built", ""},
		{"check, a word left over", nil, []string{"check", "now"}, 2, "bad-flag", ""},

		{"status, not linked", nil, []string{"status"}, 3, "not-linked", ""},
		{"status outside a checkout", outside, []string{"status"}, 3, "not-a-checkout", ""},
		{"status, a bonsai.yaml Bonsai does not read", write("bonsai.yaml", "id: x\n"), []string{"status"}, 3, "bad-config", ""},
		{"status with no home", then(write("bonsai.yaml", linkedYAML), noHome), []string{"status"}, 3, "bad-home", ""},
		{"status --full", nil, []string{"status", "--full"}, 2, "not-built", ""},
		{"status, a word left over", nil, []string{"status", "now"}, 2, "bad-flag", ""},
	}
	seen := map[string]bool{}
	for i, cc := range cases {
		for _, asJSON := range []bool{true, false} {
			t.Run(cc.name+map[bool]string{true: " --json", false: ""}[asJSON], func(t *testing.T) {
				c.t = t
				c.root = testpack.Project(t, c.tmp, "refusal-"+strconv.Itoa(i)+map[bool]string{true: "-json", false: ""}[asJSON])
				t.Chdir(c.root)
				if cc.prep != nil {
					cc.prep(c)
				}
				args := append([]string{}, cc.args...)
				w := bonsaiWord
				if len(args) > 0 && words[args[0]] != nil {
					w = words[args[0]]
				}
				if asJSON {
					args = append(args, "--json")
				}
				before := tree(t, c.root)
				code, out, errOut := c.run("", args...)
				if code != cc.exit {
					t.Fatalf("exit %d, want %d\n%s%s", code, cc.exit, out, errOut)
				}
				if after := tree(t, c.root); after != before {
					t.Errorf("a refusal wrote something")
				}
				if !asJSON {
					if !strings.Contains(out+errOut, "next: ") { // status names it in its problem line
						t.Errorf("no next step:\n%s%s", out, errOut)
					}
					return
				}
				name := outputOf(t, w)
				doc := fits(t, out, name)
				e := errorIn(doc, name)
				who := cc.who
				if who == "" {
					ew, _ := format.ErrorWord(cc.word)
					who = ew.Who
				}
				n, _ := e.Get("next")
				if e.String("code") != cc.word || whoOf(e) != who || e.String("message") == "" || n.(schema.Object).String("do") == "" {
					t.Errorf("error %s, want the word %s, a message and a next step for the %s", schema.Show(e), cc.word, who)
				}
				asciiOnly(t, "message", e.String("message")+n.(schema.Object).String("do"))
				if name == "changes" {
					if r := doc.String("result"); (cc.word == "needs-yes") != (r == "preview") || (cc.word == "conflicts") != (r == "conflict") {
						t.Errorf("result %s for %s", r, cc.word)
					}
				}
				seen[cc.word] = true
			})
		}
	}
	// Every word the engine, status and the command line's shared code name is reached by a case above, but a read
	// or write failure, the unexpected, and a write stopped part-way (TestPartialWriteFails). A word named only in a
	// word's own file (check.go's unknown-format among them) is that word's tests' to reach, so a later word's file
	// and its tests need no case here.
	for word, where := range codeWords(t, func(dir, file string) bool {
		return dir != "." || file == "word.go" || file == "main.go"
	}) {
		switch word {
		case "read-failed", "write-failed", "partly-written", "unexpected":
			continue
		}
		if !seen[word] {
			t.Errorf("no case reaches the word %s (named in %s)", word, where)
		}
	}
}

const linkedYAML = "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\n"

// hook takes no --json: its refusals are lines on stderr, each with a known word (noteError) and a next step.
func TestHookRefusals(t *testing.T) {
	for _, args := range [][]string{{"hook"}, {"hook", "nap"}, {"hook", "stop"}, {"hook", "guard", "now"}, {"hook", "guard", "--json"}} {
		code, out, errOut := runArgs(args...)
		if code != 2 || out != "" || !strings.Contains(errOut, "\nnext: ") || strings.Contains(errOut, "{") {
			t.Errorf("%v: exit %d, %q %q", args, code, out, errOut)
		}
	}
}

// A write that stops part-way through its renames (here the lock, written last, cannot be renamed into place) is
// result failed, exit 3: the plan's every field, the error partly-written naming the same command again, some files
// written and the lock not; the same command then finishes the rest.
func TestPartialWriteFails(t *testing.T) {
	c := newCLI(t)
	if code, _, _ := c.run("", append(c.linkArgs(c.pack.A), "--yes")...); code != 0 {
		t.Fatal("link failed")
	}
	testpack.SetRef(t, c.root, c.pack.A, c.pack.B)
	lock := filepath.Join(c.root, ".bonsai", "lock.json")
	saved, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	defer func(a func(*engine.Plan) error) { applyPlan = a }(applyPlan)
	applyPlan = func(p *engine.Plan) error {
		if err := os.Remove(lock); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(lock, "held"), 0o755); err != nil {
			return err
		}
		return engine.Apply(p)
	}
	for _, asJSON := range []bool{true, false} {
		args := []string{"update", "--yes"}
		if asJSON {
			args = append(args, "--json")
		}
		code, out, errOut := c.run("", args...)
		if code != 3 {
			t.Fatalf("%v: exit %d\n%s%s", args, code, out, errOut)
		}
		if !asJSON {
			if !strings.Contains(errOut, "bonsai update: .bonsai/lock.json cannot be written") ||
				!strings.Contains(errOut, "\nnext: run the same command again: it finishes the rest") {
				t.Errorf("the human lines:\n%s", errOut)
			}
			continue
		}
		doc := fits(t, out, "changes")
		e := errorIn(doc, "changes")
		files, _ := doc.Get("files")
		packs, _ := doc.Get("packs")
		if doc.String("result") != "failed" || e.String("code") != "partly-written" || whoOf(e) != "agent" ||
			doc.String("lock") != "written" || len(files.([]any)) == 0 || len(packs.([]any)) != 1 {
			t.Errorf("the failed output: %s", out)
		}
		if b, _ := os.ReadFile(filepath.Join(c.root, "demo", "extra.md")); len(b) == 0 {
			t.Errorf("no file was written before the stop")
		}
		if err := os.RemoveAll(lock); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(lock, saved, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	applyPlan = engine.Apply
	if err := os.RemoveAll(lock); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lock, saved, 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out, _ := c.run("", "update", "--yes"); code != 0 || !strings.Contains(out, "bonsai update: written") {
		t.Errorf("the same command again: %d\n%s", code, out)
	}
	if code, out, _ := c.run("", "update", "--json"); code != 0 || fits(t, out, "changes").String("result") != "nothing" {
		t.Errorf("after finishing: %d\n%s", code, out)
	}
}

// Every word a refusal names in the code is in format.ErrorWords, found by reading the code: each errorf, wsError
// and fileError call and each Code: of an Error in cmd/bonsai, internal/engine and internal/status, so a refusal no
// test reaches cannot carry a word the table does not have.
func TestErrorWordsInTheCode(t *testing.T) {
	words := codeWords(t, func(string, string) bool { return true })
	for word, where := range words {
		if _, ok := format.ErrorWord(word); !ok {
			t.Errorf("%s: the word %q is not in format.ErrorWords", where, word)
		}
	}
	if len(words) < 20 {
		t.Errorf("only %d words found in the code: the reader missed some", len(words))
	}
}

// packageDir is this package's folder, where go test starts (before any test moves into a project).
var packageDir, _ = os.Getwd()

// codeWords reads the words refusals name in the non-test Go files of cmd/bonsai (dir "."), internal/engine and
// internal/status that keep says to read: each errorf, wsError and fileError call's word and each Code: of an Error
// literal, with a file that names it.
func codeWords(t *testing.T, keep func(dir, file string) bool) map[string]string {
	t.Helper()
	found := map[string]string{}
	for _, dir := range []string{".", "../../internal/engine", "../../internal/status"} {
		pkgs, err := parser.ParseDir(token.NewFileSet(), filepath.Join(packageDir, dir), func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go") && keep(dir, fi.Name())
		}, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, pkg := range pkgs {
			for name, f := range pkg.Files {
				ast.Inspect(f, func(node ast.Node) bool {
					var lits []ast.Expr
					switch x := node.(type) {
					case *ast.CallExpr:
						fn := ""
						switch f := x.Fun.(type) {
						case *ast.Ident:
							fn = f.Name
						case *ast.SelectorExpr:
							fn = f.Sel.Name
						}
						switch fn {
						case "errorf":
							lits = x.Args[:1]
						case "wsError":
							lits = x.Args[1:2]
						case "fileError":
							lits = x.Args[1:3]
						}
					case *ast.CompositeLit:
						typ := ""
						switch t := x.Type.(type) {
						case *ast.Ident:
							typ = t.Name
						case *ast.SelectorExpr:
							typ = t.Sel.Name
						}
						if typ != "Error" {
							break
						}
						for _, el := range x.Elts {
							if kv, ok := el.(*ast.KeyValueExpr); ok {
								if k, ok := kv.Key.(*ast.Ident); ok && k.Name == "Code" {
									lits = append(lits, kv.Value)
								}
							}
						}
					}
					for _, l := range lits {
						if bl, ok := l.(*ast.BasicLit); ok && bl.Kind == token.STRING {
							if word, _ := strconv.Unquote(bl.Value); word != "" {
								found[word] = filepath.ToSlash(filepath.Join(dir, filepath.Base(name)))
							}
						}
					}
					return true
				})
			}
		}
	}
	return found
}

// A first link's changes output (bonsai.changes/1): Bonsai's own hook line is in own_hooks and is a settings line
// with runs_code false (--yes writes it); the pack's hook line is a runs_code item and a settings line with
// runs_code true; the workspace, the packs, every file, the lock and the error, each in the schema's shape.
func TestChangesFirstLink(t *testing.T) {
	c := newCLI(t)
	code, out, _ := c.run("", append(c.linkArgs(c.pack.A), "--json")...)
	doc := fits(t, out, "changes")
	if code != 4 || doc.String("command") != "init" || doc.String("result") != "preview" || errorIn(doc, "changes").String("code") != "needs-yes" {
		t.Fatalf("the preview: %d\n%s", code, out)
	}
	own, _ := doc.Get("own_hooks")
	if l := own.([]any); len(l) != 1 || !strings.Contains(l[0].(string), "bonsai hook guard || exit 2") {
		t.Errorf("own_hooks %s", schema.Show(own))
	}
	settings, _ := doc.Get("settings")
	hooks := 0
	for _, s := range settings.([]any) {
		o := s.(schema.Object)
		rc, _ := o.Get("runs_code")
		switch {
		case strings.Contains(o.String("line"), "bonsai hook guard"):
			hooks++
			if rc != false {
				t.Errorf("Bonsai's own hook line at a first link has runs_code %v", rc)
			}
		case strings.Contains(o.String("line"), "echo demo hook A"):
			hooks++
			if rc != true {
				t.Errorf("the pack's hook line has runs_code %v", rc)
			}
		case rc != false:
			t.Errorf("a %s line has runs_code %v", o.String("kind"), rc)
		}
	}
	if hooks != 2 || jsonItems(t, doc) != "hook add SessionStart (startup): echo demo hook A (demo-pack)" {
		t.Errorf("settings %s", schema.Show(settings))
	}
	ws, _ := doc.Get("workspace")
	packs, _ := doc.Get("packs")
	first := packs.([]any)[0].(schema.Object)
	from, _ := first.Get("from")
	if ws.(schema.Object).String("name") != "demo" || ws.(schema.Object).String("root") != filepath.ToSlash(c.root) ||
		from != nil || first.String("to") != c.pack.A || doc.String("lock") != "written" {
		t.Errorf("workspace %s, packs %s, lock %s", schema.Show(ws), schema.Show(packs), doc.String("lock"))
	}
	code, out, _ = c.run("", append(c.linkArgs(c.pack.A), "--yes", "--json")...)
	doc = fits(t, out, "changes")
	if e, _ := doc.Get("error"); code != 0 || doc.String("result") != "applied" || e != nil {
		t.Errorf("the link: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "check", "--json")
	if doc := fits(t, out, "check"); code != 0 || schema.Show(doc) != `{"format":"bonsai.check/1","findings":[],"warnings":[],"error":null}` {
		t.Errorf("check --json: %d\n%s", code, out)
	}
}
