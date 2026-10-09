package main

// Consent to code through the command line (plan-5, piece 5.1.1, rules 1-8; internal/engine/consent.go): one table
// of cases on the test pack's commits A to F and the fixture packs, each run under every flag combination that
// matters (--yes, --allow-exec, --json, a terminal or not, y or n), each asserting the exit code, that nothing or
// everything was written, and what the preview or the JSON names. Conflicts with --keep and --adopt follow it.

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// tree is every file of a checkout but .git, with its bytes' hash and its modification time.
func tree(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		sum := sha256.Sum256(raw)
		rel, _ := filepath.Rel(root, p)
		b.WriteString(filepath.ToSlash(rel) + " " + hex.EncodeToString(sum[:]) + " " + info.ModTime().Format(time.RFC3339Nano) + "\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// copyTree copies a prepared project, .git included, so each run starts from the same state.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// jsonItems gives a --json document's runs_code as "kind change item (pack)", joined by "; ".
func jsonItems(t *testing.T, doc schema.Object) string {
	t.Helper()
	v, ok := doc.Get("runs_code")
	list, isList := v.([]any)
	if !ok || !isList {
		t.Fatalf("no runs_code list: %s", schema.Show(doc))
	}
	var out []string
	for _, it := range list {
		o := it.(schema.Object)
		out = append(out, o.String("kind")+" "+o.String("change")+" "+o.String("item")+" ("+o.String("pack")+")")
		if o.String("why") == "" {
			t.Errorf("an item with no why: %s", schema.Show(o))
		}
	}
	return strings.Join(out, "; ")
}

func lockedCommit(t *testing.T, root string) string {
	t.Helper()
	lock, err := workspace.LoadLock(root)
	if err != nil || len(lock.Packs) != 1 {
		t.Fatalf("the lock: %v", err)
	}
	return lock.Packs[0].Commit
}

type consentCase struct {
	name   string
	prep   func(c *cli) // brings c.root to the case's starting state
	args   []string     // the command, without --yes, --allow-exec and --json
	items  string       // what Runs code lists ("" for nothing)
	target string       // the commit the lock holds once written
	after  func(c *cli) // a check of the written project, beyond the target commit
}

type flagCombo struct {
	flags  []string
	answer string // "" for no terminal; else the answer typed at one
}

func (f flagCombo) String() string {
	s := strings.Join(f.flags, " ")
	if s == "" {
		s = "no flags"
	}
	if f.answer != "" {
		s += ", at a terminal answering " + strings.TrimSpace(f.answer)
	}
	return s
}

var combos = []flagCombo{
	{nil, ""}, {[]string{"--yes"}, ""}, {nil, "y\n"}, {[]string{"--yes"}, "y\n"},
	{[]string{"--allow-exec"}, ""}, {[]string{"--allow-exec"}, "y\n"}, {[]string{"--allow-exec"}, "n\n"},
	{[]string{"--yes", "--allow-exec"}, ""}, {[]string{"--allow-exec", "--yes"}, "y\n"},
	{[]string{"--json"}, ""}, {[]string{"--json", "--yes"}, ""}, {[]string{"--json", "--allow-exec"}, ""},
	{[]string{"--json", "--yes", "--allow-exec"}, ""}, {[]string{"--json", "--yes", "--allow-exec"}, "y\n"},
}

// want is what a combo gives on a case: the exit code, whether it writes, and the next step's ending.
func want(code bool, f flagCombo) (exit int, writes bool, next string) {
	has := func(flag string) bool {
		for _, x := range f.flags {
			if x == flag {
				return true
			}
		}
		return false
	}
	yes, allow, json, tty := has("--yes"), has("--allow-exec"), has("--json"), f.answer != "" && !has("--json")
	switch {
	case code && !allow:
		return 4, false, " --allow-exec --yes"
	case yes:
		return 0, true, ""
	case tty && f.answer == "y\n":
		return 0, true, ""
	case tty:
		return 4, false, ""
	case json || !tty:
		if allow {
			return 4, false, " --allow-exec --yes"
		}
		return 4, false, " --yes"
	}
	return -1, false, ""
}

func TestConsentCommand(t *testing.T) {
	c := newCLI(t)
	quiet, qc := testpack.QuietPack(t, c.tmp)
	run, rc := testpack.RunPack(t, c.tmp)
	p := c.pack
	initTo := func(source, ref string) []string {
		return []string{"init", "--name", "demo", "--source", source, "--ref", ref}
	}
	linked := func(source, ref string) func(c *cli) {
		return func(c *cli) {
			if code, out, errOut := c.run("", append(initTo(source, ref), "--yes", "--allow-exec")...); code != 0 {
				t.Fatalf("link: %d\n%s%s", code, out, errOut)
			}
		}
	}
	moved := func(source, from, to string) func(c *cli) {
		return func(c *cli) { linked(source, from)(c); testpack.SetRef(t, c.root, from, to) }
	}
	relink := func(from, to string) func(c *cli) {
		return func(c *cli) {
			linked(p.Source, from)(c)
			if err := os.Remove(filepath.Join(c.root, ".bonsai", "lock.json")); err != nil {
				t.Fatal(err)
			}
			if to != from {
				testpack.SetRef(t, c.root, from, to)
			}
		}
	}
	settingsHold := func(want ...string) func(c *cli) {
		return func(c *cli) {
			s, _ := os.ReadFile(filepath.Join(c.root, ".claude", "settings.json"))
			for _, w := range want {
				if !strings.Contains(string(s), w) {
					t.Errorf("settings lack %s:\n%s", w, s)
				}
			}
		}
	}
	fileHolds := func(rel, want string) func(c *cli) {
		return func(c *cli) {
			if b, _ := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(rel))); string(b) != want {
				t.Errorf("%s holds %q, want %q", rel, b, want)
			}
		}
	}
	hookA := "hook add SessionStart (startup): echo demo hook A (demo-pack)"
	cases := []consentCase{
		{"a first link to A", nil, initTo(p.Source, p.A), hookA, p.A,
			settingsHold("echo demo hook A", "bonsai hook guard || exit 2")},
		{"a first link to the no-hook fixture", nil, initTo(quiet, qc[0]), "", qc[0], settingsHold("bonsai hook guard || exit 2")},
		{"a first link to F", nil, initTo(p.Source, p.F),
			"hook add SessionStart (startup): echo demo hook E (demo-pack); plugin add hooks/hooks.json (demo-pack)", p.F, nil},
		{"a first link to the hook-run fixture", nil, initTo(run, rc[0]),
			"hook add SessionStart (startup): sh run/hello.sh (run-pack); file add run/hello.sh (run-pack)", rc[0],
			fileHolds("run/hello.sh", "echo hello 1\n")},
		{"A to B, no code", moved(p.Source, p.A, p.B), []string{"update"}, "", p.B, fileHolds("demo/guide.md", "# Guide\n\nEdition 2.\n")},
		{"C to D, the hook line alone", moved(p.Source, p.C, p.D), []string{"update"},
			"hook change SessionStart (startup): echo demo hook D (demo-pack)", p.D, settingsHold("echo demo hook D")},
		{"D to E, the mixed update", moved(p.Source, p.D, p.E), []string{"update"},
			"hook change SessionStart (startup): echo demo hook E (demo-pack)", p.E,
			func(c *cli) {
				settingsHold("echo demo hook E")(c)
				fileHolds("demo/guide.md", "# Guide\n\nEdition 4.\n")(c)
			}},
		{"E to F, a hook of the plugin itself", moved(p.Source, p.E, p.F), []string{"update"}, "plugin add hooks/hooks.json (demo-pack)", p.F, nil},
		{"the hook-run file changed alone", moved(run, rc[0], rc[1]), []string{"update"}, "file change run/hello.sh (run-pack)", rc[1],
			fileHolds("run/hello.sh", "echo hello 2\n")},
		{"a relink, the lock deleted, the hook line changed", relink(p.C, p.D), []string{"init"},
			"hook add SessionStart (startup): echo demo hook D (demo-pack)", p.D, settingsHold("echo demo hook D")},
		{"a relink, the lock deleted, nothing changed", relink(p.C, p.C), []string{"init"}, "", p.C, nil},
		{"a relink, the lock deleted, a plugin with code", relink(p.F, p.F), []string{"init"}, "plugin add hooks/hooks.json (demo-pack)", p.F, nil},
	}
	for ci, cc := range cases {
		template := testpack.Project(t, c.tmp, "case-"+strconv.Itoa(ci))
		if cc.prep != nil {
			c.t, c.root = t, template
			t.Chdir(template)
			cc.prep(c)
		}
		for fi, f := range combos {
			t.Run(cc.name+"/"+f.String(), func(t *testing.T) {
				c.t = t
				c.root = filepath.Join(c.tmp, "run-"+strconv.Itoa(ci)+"-"+strconv.Itoa(fi))
				copyTree(t, template, c.root)
				t.Chdir(c.root)
				exit, writes, next := want(cc.items != "", f)
				before := tree(t, c.root)
				code, out, errOut := c.run(f.answer, append(append([]string{}, cc.args...), f.flags...)...)
				if code != exit {
					t.Fatalf("exit %d, want %d\n%s%s", code, exit, out, errOut)
				}
				if writes {
					if got := lockedCommit(t, c.root); got != cc.target {
						t.Errorf("the lock holds %s, want %s", got, cc.target)
					}
					if code, again, _ := c.run("", "update", "--json"); code != 0 || !strings.Contains(again, `"result": "nothing"`) {
						t.Errorf("not everything was written: update again gives %d\n%s", code, again)
					}
					if cc.after != nil {
						cc.after(c)
					}
				} else if after := tree(t, c.root); after != before {
					t.Errorf("something was written:\n%s\n---\n%s", before, after)
				}
				json := strings.HasPrefix(out, "{")
				if json {
					doc, err := schema.Decode([]byte(out))
					if err != nil {
						t.Fatalf("not JSON: %v\n%s", err, out)
					}
					o := doc.(schema.Object)
					if got := jsonItems(t, o); got != cc.items {
						t.Errorf("runs_code\n  %s\nwant\n  %s", got, cc.items)
					}
					if n, _ := o.Get("exit"); schema.Show(n) != strconv.Itoa(exit) {
						t.Errorf("JSON exit %s", schema.Show(n))
					}
					if next != "" {
						e, _ := o.Get("error")
						if eo, ok := e.(schema.Object); !ok || !strings.HasSuffix(eo.String("next"), next) {
							t.Errorf("JSON next, want it to end %q: %s", next, schema.Show(e))
						}
					}
					return
				}
				asks := f.answer != "" && !hasFlag(f, "--yes") && (cc.items == "" || hasFlag(f, "--allow-exec"))
				if strings.Contains(out, "[y/N]") != asks {
					t.Errorf("the y/N question (asked: %v):\n%s", strings.Contains(out, "[y/N]"), out)
				}
				if cc.items != "" && !strings.Contains(out, "Runs code: ") {
					t.Errorf("the preview has no Runs code:\n%s", out)
				}
				for _, it := range strings.Split(cc.items, "; ") {
					if it == "" {
						continue
					}
					parts := strings.SplitN(it, " ", 3) // kind change item (pack)
					line := strings.TrimSuffix(parts[2], ")")
					line = line[:strings.LastIndex(line, " (")]
					if !strings.Contains(out, parts[1]) || !strings.Contains(out, line) {
						t.Errorf("the preview does not name %s:\n%s", it, out)
					}
				}
				if next != "" && !strings.HasSuffix(out, next+"\n") {
					t.Errorf("the next step does not end %q:\n%s", next, out)
				}
				if cc.items == "" && !hasFlag(f, "--allow-exec") && strings.Contains(out, " --allow-exec --yes") {
					t.Errorf("--allow-exec named where nothing runs code:\n%s", out)
				}
			})
		}
	}
}

func hasFlag(f flagCombo, flag string) bool {
	for _, x := range f.flags {
		if x == flag {
			return true
		}
	}
	return false
}

// Conflicts beside code: one refusal names --allow-exec and the --keep and --adopt steps; with --allow-exec the
// conflict stops the run (exit 5); --keep and --adopt then settle it. A kept hook-run file is no code.
func TestConsentWithConflicts(t *testing.T) {
	c := newCLI(t)
	run, rc := testpack.RunPack(t, c.tmp)
	p := c.pack
	edited := testpack.Project(t, c.tmp, "edited")
	c.root = edited
	t.Chdir(edited)
	if code, _, _ := c.run("", "init", "--name", "demo", "--source", p.Source, "--ref", p.D, "--yes", "--allow-exec"); code != 0 {
		t.Fatal("link")
	}
	if err := os.WriteFile(filepath.Join(edited, "demo", "guide.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testpack.SetRef(t, edited, p.D, p.E)
	hookRun := testpack.Project(t, c.tmp, "hookrun")
	c.root = hookRun
	t.Chdir(hookRun)
	if code, _, _ := c.run("", "init", "--name", "demo", "--source", run, "--ref", rc[0], "--yes", "--allow-exec"); code != 0 {
		t.Fatal("link")
	}
	if err := os.WriteFile(filepath.Join(hookRun, "run", "hello.sh"), []byte("echo my own\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testpack.SetRef(t, hookRun, rc[0], rc[1])

	cases := []struct {
		template string
		args     []string
		answer   string
		exit     int
		writes   bool
		has      []string
	}{
		{edited, []string{"update", "--yes"}, "", 4, false, []string{"Conflicts: 1.", "Runs code: 1 item, which needs --allow-exec",
			"next: read each item under Runs code, and settle the conflicts: to keep your edits, run: bonsai update --allow-exec --yes --keep demo/guide.md\n" +
				"  or, to take the pack's copies (yours are saved in the Bonsai home's cache, never in the repo), run: bonsai update --allow-exec --yes --adopt demo/guide.md\n"}},
		{edited, []string{"update", "--yes", "--allow-exec"}, "", 5, false, []string{"Stopped: 1 conflict (demo/guide.md): nothing was written.",
			"run: bonsai update --allow-exec --yes --keep demo/guide.md\n"}},
		{edited, []string{"update", "--allow-exec"}, "y\n", 5, false, []string{"Write these changes? [y/N] Stopped: 1 conflict"}},
		{edited, []string{"update", "--yes", "--keep", "demo/guide.md"}, "", 4, false,
			[]string{"to write them with the rest, run: bonsai update --keep demo/guide.md --allow-exec --yes\n"}},
		{edited, []string{"update", "--yes", "--keep", "demo/guide.md"}, "y\n", 4, false, []string{"Refused: this update writes code"}},
		{edited, []string{"update", "--yes", "--allow-exec", "--keep", "demo/guide.md"}, "", 0, true, []string{"kept         demo/guide.md"}},
		{edited, []string{"update", "--yes", "--allow-exec", "--adopt", "demo/guide.md"}, "", 0, true, []string{"replaced     demo/guide.md"}},
		{edited, []string{"update", "--json", "--yes", "--adopt", "demo/guide.md"}, "", 4, false, []string{`"result": "refused"`, `"item": "SessionStart (startup): echo demo hook E"`}},
		{hookRun, []string{"update", "--yes"}, "", 4, false, []string{"change  file    run/hello.sh  (run-pack)",
			"A file the pack run-pack's hook line runs; a conflict: --adopt would write the pack's copy.",
			"bonsai update --allow-exec --yes --keep run/hello.sh"}},
		{hookRun, []string{"update", "--yes", "--keep", "run/hello.sh"}, "", 0, true, []string{"Runs code: nothing that needs --allow-exec"}},
		{hookRun, []string{"update", "--yes", "--adopt", "run/hello.sh"}, "", 4, false, []string{"bonsai update --adopt run/hello.sh --allow-exec --yes"}},
		{hookRun, []string{"update", "--yes", "--allow-exec", "--adopt", "run/hello.sh"}, "", 0, true, []string{"consented to with --allow-exec"}},
	}
	for i, cc := range cases {
		t.Run(strings.Join(cc.args, " ")+" "+strings.TrimSpace(cc.answer), func(t *testing.T) {
			c.t = t
			c.root = filepath.Join(c.tmp, "conflict-"+strconv.Itoa(i))
			copyTree(t, cc.template, c.root)
			t.Chdir(c.root)
			before := tree(t, c.root)
			code, out, errOut := c.run(cc.answer, cc.args...)
			if code != cc.exit {
				t.Fatalf("exit %d, want %d\n%s%s", code, cc.exit, out, errOut)
			}
			for _, h := range cc.has {
				if !strings.Contains(out, h) {
					t.Errorf("the output lacks %q:\n%s", h, out)
				}
			}
			if cc.writes {
				if code, again, _ := c.run("", "update", "--json"); code != 0 || !strings.Contains(again, `"result": "nothing"`) {
					t.Errorf("update again gives %d\n%s", code, again)
				}
			} else if after := tree(t, c.root); after != before {
				t.Errorf("something was written")
			}
		})
	}
}
