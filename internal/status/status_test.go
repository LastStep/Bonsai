package status

// The partial status --json's test (plan part 2, "The partial status --json"): the document holds every field of
// bonsai.status/1 in the schema's order and validates against part 0's schema; every built field has a value; the
// fields not built yet are held in notBuiltYet with the exact null or [] they print. A listed field that gets a
// value fails here, and so does an unlisted one that is null: step 5.1 empties the list, and each part that builds
// a field takes it off.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// notBuiltYet: the fields the walking skeleton has not built at part 2, each with what it prints until it is.
var notBuiltYet = map[string]string{
	"formats":        "null", // which majors this Bonsai reads and writes: step 5.1
	"documents":      "[]",   // the declared document kinds (contract §7.3): step 5.1
	"labels":         "[]",   // namespaces in force: step 5.1
	"lanes":          "[]",   // from the packs' declares: step 5.1
	"status_writes":  "null", // this machine's settings: step 5.1 (bonsai settings)
	"status_command": "null", // the same
	"active_task":    "null", // contract §13: step 5.1
	"needs":          "[]",   // packs and the Claude Code floor: part 3 and step 5.1
	"checks":         "null", // --full: step 5.1
}

const cfg = `format: bonsai.workspace/1   # comments are fine
id: ws-7kq2m4xw5r3t6y2u7p4a5c3e2b
name: example
packs:
  - id: test-pack
    source: "https://github.com/LastStep/bonsai-test-pack"
    ref: v0.1.0
protected: [".claude/**", "bonsai.yaml", "protected.txt"]
person_only: [".claude/**", "bonsai.yaml"]
`

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=bonsai-test", "-c", "user.email=bonsai-test",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// project makes a git checkout with bonsai.yaml (none when yaml is "") and a temporary Bonsai home, isolated from
// the machine's git configuration and from any repository above the temporary folder.
func project(t *testing.T, yaml string) (root, home string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH: status finds a project through git")
	}
	tmp := t.TempDir()
	empty := filepath.Join(tmp, "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CEILING_DIRECTORIES", tmp)
	home = filepath.Join(tmp, "home")
	t.Setenv(workspace.HomeEnv, home)
	root = filepath.Join(tmp, "project")
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "init", "-q")
	if yaml != "" {
		if err := os.WriteFile(filepath.Join(root, "bonsai.yaml"), []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root, home
}

func real(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(r)
}

// checkShape holds a document to the schema: every field in order, valid, and the named list.
func checkShape(t *testing.T, doc schema.Object, failed bool) {
	t.Helper()
	out, err := schema.Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	back, err := schema.Decode(out)
	if err != nil {
		t.Fatal(err)
	}
	if msgs := schema.Validate(Schema(), back); len(msgs) != 0 {
		t.Errorf("does not validate against status.schema.json: %v\n%s", msgs, out)
	}
	if msgs := schema.CheckOrder(Schema(), back); len(msgs) != 0 {
		t.Errorf("out of the schema's order: %v", msgs)
	}
	props, _ := Schema().Get("properties")
	if got, want := strings.Join(back.(schema.Object).Keys(), " "), strings.Join(props.(schema.Object).Keys(), " "); got != want {
		t.Fatalf("fields\n  %s\nwant every field of bonsai.status/1 in order\n  %s", got, want)
	}
	for _, m := range back.(schema.Object) {
		shown := schema.Show(m.Value)
		switch want, listed := notBuiltYet[m.Key]; {
		case failed && m.Key != "format" && m.Key != "bonsai" && m.Key != "problems":
			if m.Value != nil {
				t.Errorf("exit 3: %s is %s, want null", m.Key, shown)
			}
		case failed:
		case listed && shown != want:
			t.Errorf("%s is listed as not built yet (%s) but holds %s: take it off notBuiltYet", m.Key, want, shown)
		case !listed && m.Value == nil:
			t.Errorf("%s is null but not listed as not built yet: build it or list it", m.Key)
		}
	}
	for k := range notBuiltYet {
		if _, ok := props.(schema.Object).Get(k); !ok {
			t.Errorf("notBuiltYet names %s, which bonsai.status/1 does not have", k)
		}
	}
}

func TestStatusOfALinkedProject(t *testing.T) {
	root, home := project(t, cfg)
	for _, dir := range []string{root, filepath.Join(root, "sub")} {
		doc, code := Build(dir, "dev")
		if code != ExitOK {
			t.Fatalf("exit %d: %s", code, schema.Show(doc))
		}
		checkShape(t, doc, false)
		r := real(t, root)
		key, _ := workspace.MachineKey(root)
		want := map[string]string{
			"format":      `"bonsai.status/1"`,
			"bonsai":      `"dev"`,
			"mode":        `"offline"`,
			"workspace":   `{"id":"ws-7kq2m4xw5r3t6y2u7p4a5c3e2b","name":"example","root":` + schema.Show(r) + `}`,
			"home":        `{"path":` + schema.Show(filepath.ToSlash(home)) + `,"key":"` + key + `"}`,
			"local":       `{"log":` + schema.Show(r+"/.bonsai/local/log") + `,"asks":` + schema.Show(r+"/.bonsai/local/asks") + `,"ladder":` + schema.Show(r+"/.bonsai/local/ladder") + `}`,
			"person_only": `[".claude/**","bonsai.yaml"]`,
			"packs":       `[]`,
			"files":       `{"changed":0,"missing":0,"format0_changed":0}`,
			// bonsai.yaml names a pack, but no update has run: check's findings (contract §12).
			"problems": `[".bonsai/lock.json: is missing; next: run bonsai update to write it",` +
				`".bonsai/.gitignore is missing, so .bonsai/local/ could be committed; next: run bonsai update --yes to write it again"]`,
		}
		for k, w := range want {
			v, _ := doc.Get(k)
			if got := schema.Show(v); got != w {
				t.Errorf("%s = %s, want %s", k, got, w)
			}
		}
		for _, k := range []string{"root", "path"} {
			for _, f := range []string{"workspace", "home"} {
				v, _ := doc.Get(f)
				if s := v.(schema.Object).String(k); strings.Contains(s, `\`) {
					t.Errorf("%s.%s holds a backslash: %s", f, k, s)
				}
			}
		}
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Errorf("status wrote the home: it writes nothing")
	}
}

// In a worktree, root, home key and local are the main checkout's; the id and name are the worktree's bonsai.yaml.
func TestStatusInAWorktree(t *testing.T) {
	root, _ := project(t, cfg)
	gitRun(t, root, "add", "bonsai.yaml")
	gitRun(t, root, "commit", "-q", "-m", "link")
	wt := filepath.Join(filepath.Dir(root), "project-wt")
	gitRun(t, root, "worktree", "add", "-q", wt)
	doc, code := Build(wt, "dev")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	ws, _ := doc.Get("workspace")
	if got := ws.(schema.Object).String("root"); got != real(t, root) {
		t.Errorf("root %s, want the main checkout %s", got, real(t, root))
	}
	mainDoc, _ := Build(root, "dev")
	h1, _ := doc.Get("home")
	h2, _ := mainDoc.Get("home")
	if !schema.Equal(h1, h2) {
		t.Errorf("a worktree has another machine folder: %s, %s", schema.Show(h1), schema.Show(h2))
	}
}

// Exit 3: format, bonsai and problems filled, every other field null, one problem naming its next step.
func TestStatusExit3(t *testing.T) {
	cases := []struct {
		name, yaml, want string
	}{
		{"no bonsai.yaml", "", "bonsai.yaml: is not in this checkout"},
		{"a refused bonsai.yaml", "format: bonsai.workspace/1\nid: 0755\n", "bonsai.yaml line 2: the plain value \"0755\""},
		{"a newer bonsai.yaml", "format: bonsai.workspace/2\n", "format too new"},
		{"a format-0 bonsai.yaml", "id: x\n", "no format: line first"},
		{"format: and a tab", "format:\tbonsai.workspace/1\nid: ws-7kq2m4xw5r3t6y2u7p4a5c3e2b\nname: x\n", "bonsai.yaml line 1: a tab between the key's colon and its value"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, _ := project(t, c.yaml)
			doc, code := Build(root, "dev")
			if code != ExitRuntime {
				t.Fatalf("exit %d, want 3", code)
			}
			checkShape(t, doc, true)
			p, _ := doc.Get("problems")
			list := p.([]any)
			if len(list) != 1 || !strings.Contains(list[0].(string), c.want) || !strings.Contains(list[0].(string), "; next: ") ||
				(c.name == "format: and a tab" && strings.Contains(list[0].(string), "no format: line first")) {
				t.Errorf("problems %s, want one holding %q and a next step", schema.Show(p), c.want)
			}
			if text := Text(doc); !strings.HasPrefix(text, "bonsai status: cannot read this workspace.\n  ") {
				t.Errorf("text %q", text)
			}
		})
	}
	t.Run("outside git", func(t *testing.T) {
		project(t, "")
		outside := t.TempDir()
		t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(outside))
		doc, code := Build(outside, "dev")
		if code != ExitRuntime {
			t.Fatalf("exit %d", code)
		}
		checkShape(t, doc, true)
		if p, _ := doc.Get("problems"); !strings.Contains(schema.Show(p), "is not inside a git checkout") {
			t.Errorf("problems %s", schema.Show(p))
		}
	})
}

// The JSON and the text are byte-stable and ASCII, whatever the paths hold.
func TestStatusIsByteStableAndASCII(t *testing.T) {
	root, _ := project(t, strings.Replace(cfg, `"protected.txt"`, `"caf`+"\xc3\xa9"+`.txt"`, 1)+"")
	nonASCII := filepath.Join(filepath.Dir(root), "proj-caf\xc3\xa9")
	if err := os.Rename(root, nonASCII); err != nil {
		t.Fatal(err)
	}
	doc, code := Build(nonASCII, "dev")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	first, _ := schema.Encode(doc)
	for i := 0; i < 5; i++ {
		again, _ := Build(nonASCII, "dev")
		out, _ := schema.Encode(again)
		if !bytes.Equal(out, first) {
			t.Fatal("two builds print differently")
		}
	}
	text := Text(doc)
	for _, s := range []string{string(first), text} {
		for i := 0; i < len(s); i++ {
			if s[i] > 0x7e || (s[i] < 0x20 && s[i] != '\n') {
				t.Fatalf("byte %d is %#x: not ASCII", i, s[i])
			}
		}
	}
	if !strings.Contains(string(first), `proj-caf\u00e9`) || !strings.Contains(text, `proj-caf\u00e9`) {
		t.Errorf("the non-ASCII path is not escaped:\n%s\n%s", first, text)
	}
	for _, want := range []string{"Workspace example, id ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", "  bonsai.yaml      this project's",
		"Person-only paths: .claude/**, bonsai.yaml", "Problems:\n  .bonsai/lock.json: is missing; next: run bonsai update",
		"This project's machine folder: workspaces/r-", "Packs: none locked\n", "A copy meant as a new project needs its own id: bonsai init --new-id\n"} {
		if !strings.Contains(text, want) {
			t.Errorf("the text lacks %q:\n%s", want, text)
		}
	}
}

// A project linked by the engine: packs from the lock with their state, file counts, and check's findings as
// problems (contract §12).
func TestStatusOfAProjectTheEngineLinked(t *testing.T) {
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home := testpack.Isolate(t, tmp)
	pack := testpack.Build(t, tmp)
	root := testpack.Project(t, tmp, "linked")
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test",
		Init: &engine.InitValues{Name: "linked", Source: pack.Source, Ref: pack.A}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(p); err != nil {
		t.Fatal(err)
	}
	doc, code := Build(root, "dev")
	if code != ExitOK {
		t.Fatalf("exit %d", code)
	}
	checkShape(t, doc, false)
	for k, want := range map[string]string{
		"packs":    `[{"id":"demo-pack","version":"0.1.0","commit":"` + pack.A + `","state":"ok"}]`,
		"files":    `{"changed":0,"missing":0,"format0_changed":0}`,
		"problems": `[]`,
	} {
		if v, _ := doc.Get(k); schema.Show(v) != want {
			t.Errorf("%s = %s, want %s", k, schema.Show(v), want)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "demo", "guide.md"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, _ = Build(root, "dev")
	checkShape(t, doc, false)
	if v, _ := doc.Get("packs"); !strings.Contains(schema.Show(v), `"state":"changed"`) {
		t.Errorf("packs %s", schema.Show(v))
	}
	if v, _ := doc.Get("files"); schema.Show(v) != `{"changed":1,"missing":0,"format0_changed":0}` {
		t.Errorf("files %s", schema.Show(v))
	}
	text := Text(doc)
	if !strings.Contains(text, "Packs: demo-pack 0.1.0 at "+pack.A[:7]+" (changed)\n") || !strings.Contains(text, "demo/guide.md was edited") {
		t.Errorf("text:\n%s", text)
	}
}
