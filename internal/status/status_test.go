package status

// status --json's test: the document holds every field of bonsai.status/1 in the schema's order and validates against
// its schema; every field has a value where it applies, and null or [] only where contract §12 says it does not apply
// (doesNotApply, below: a field null or [] that is not listed fails, and so does a listed one holding something else
// than its listed value when it is empty). The walking skeleton's notBuiltYet list is gone: step 5.1.6 built its last
// two fields, needs and checks.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// doesNotApply: the fields that print null or [] where they do not apply (contract §12), with what they print then,
// and when. Every other field always holds a value on exit 0.
var doesNotApply = map[string]string{
	"packs":          "[]",   // no pack is locked yet (no update has run)
	"labels":         "[]",   // no pack declares labels, and none is attached on this machine
	"lanes":          "[]",   // no pack defines lanes
	"status_command": "null", // status_writes is agents
	"person_only":    "[]",   // bonsai.yaml names none
	"problems":       "[]",   // check finds nothing
	"checks":         "null", // without --full
	"error":          "null", // nothing refused
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
		switch want, listed := doesNotApply[m.Key]; {
		case m.Key == "error" && failed:
			// The error object (step 5.1.4b): a known word, a sentence, and a next step with who takes it.
			eo, ok := m.Value.(schema.Object)
			if !ok {
				t.Errorf("exit 3: error is %s, want the error object", shown)
				break
			}
			if _, known := format.ErrorWord(eo.String("code")); !known || eo.String("message") == "" {
				t.Errorf("exit 3: error %s has a word not in format.ErrorWords, or no message", shown)
			}
		case m.Key == "error":
			if m.Value != nil {
				t.Errorf("exit 0: error is %s, want null (nothing refused)", shown)
			}
		case failed && m.Key != "format" && m.Key != "bonsai" && m.Key != "problems":
			if m.Value != nil {
				t.Errorf("exit 3: %s is %s, want null", m.Key, shown)
			}
		case failed:
		case listed && (m.Value == nil || shown == "[]") && shown != want:
			t.Errorf("%s is %s where it does not apply, want %s", m.Key, shown, want)
		case !listed && (m.Value == nil || shown == "[]" || shown == "{}"):
			t.Errorf("%s is %s, but contract section 12 has it always apply: fill it, or list it in doesNotApply", m.Key, shown)
		}
	}
	for k := range doesNotApply {
		if _, ok := props.(schema.Object).Get(k); !ok {
			t.Errorf("doesNotApply names %s, which bonsai.status/1 does not have", k)
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
			"problems": `[".bonsai/lock.json: is missing; next: link the project again from bonsai.yaml (a pack's code also needs --allow-exec, a person's consent, which init names), run: bonsai init --yes",` +
				`".bonsai/.gitignore is missing, so .bonsai/local/ could be committed; next: run: bonsai update --yes"]`,
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
		name, yaml, want, code string
	}{
		{"no bonsai.yaml", "", "bonsai.yaml: is not in this checkout", "not-linked"},
		{"a refused bonsai.yaml", "format: bonsai.workspace/1\nid: 0755\n", "bonsai.yaml line 2: the plain value \"0755\"", "bad-config"},
		{"a newer bonsai.yaml", "format: bonsai.workspace/2\n", "format too new", "bad-config"},
		{"a format-0 bonsai.yaml", "id: x\n", "no format: line first", "bad-config"},
		{"format: and a tab", "format:\tbonsai.workspace/1\nid: ws-7kq2m4xw5r3t6y2u7p4a5c3e2b\nname: x\n", "bonsai.yaml line 1: a tab between the key's colon and its value", "bad-config"},
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
			if e, _ := doc.Get("error"); e.(schema.Object).String("code") != c.code {
				t.Errorf("error %s, want the word %s", schema.Show(e), c.code)
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
		e, _ := doc.Get("error")
		next, _ := e.(schema.Object).Get("next")
		if e.(schema.Object).String("code") != "not-a-checkout" || next.(schema.Object).String("who") != "agent" {
			t.Errorf("error %s", schema.Show(e))
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
		"Person-only paths: .claude/**, bonsai.yaml", "Problems:\n  .bonsai/lock.json: is missing; next: link the project again from bonsai.yaml",
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
	// --allow-exec: the test pack's hook line is a pack's code, which a first link writes only with it (step 5.1.1).
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test", AllowExec: true,
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

// Step 5.1.5's fields, from a project linked to a pack that declares lanes, labels and document kinds, with this
// machine's settings and a running task: documents, labels, lanes, status_writes, status_command and active_task.
func TestStatusFromTheLockAndTheMachine(t *testing.T) {
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home := testpack.Isolate(t, tmp)
	t.Setenv("BONSAI_TASK", "")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	src, shas := testpack.DeclaringPack(t, tmp)
	root := testpack.Project(t, tmp, "linked")
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test",
		Init: &engine.InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(p); err != nil {
		t.Fatal(err)
	}
	task := "---\nformat: bonsai.task/1\nid: T-0901\ntitle: x\nstatus: running\nlane: light\ndone_when: []\ndepends_on: []\n" +
		"blocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels: {}\n---\n"
	write := func(path, content string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "work", "tasks", "T-0901-x.md"), task)
	machine, err := workspace.MachineDir(home, root)
	if err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(machine, "settings.json"), `{"status_writes": "command", "status_command": "tracker move"}`)
	doc, code := Build(root, "dev")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	want := map[string]string{
		"labels":         `[{"namespace":"bonsai","from":"base","version":1}]`,
		"lanes":          `[{"name":"light","approve_first":false,"close":"agent","from":"base"},{"name":"full","approve_first":true,"close":"person","from":"base"}]`,
		"status_writes":  `"command"`,
		"status_command": `"tracker move"`,
		"active_task":    `{"id":"T-0901","how":"running","why":null}`,
	}
	for k, w := range want {
		if v, _ := doc.Get(k); schema.Show(v) != w {
			t.Errorf("%s = %s, want %s", k, schema.Show(v), w)
		}
	}
	docs, _ := doc.Get("documents")
	var kinds []string
	for _, d := range docs.([]any) {
		kinds = append(kinds, d.(schema.Object).String("kind")+"@"+d.(schema.Object).String("from"))
	}
	if strings.Join(kinds, " ") != "task@bonsai run@bonsai state@bonsai answers@bonsai memory@bonsai tasks@bonsai sessions@bonsai plan@base bugs@base" {
		t.Errorf("documents %v", kinds)
	}
	// BONSAI_TASK names a task for this project's session; another project's session does not count.
	t.Setenv("BONSAI_TASK", "T-0999")
	doc, _ = Build(root, "dev")
	if v, _ := doc.Get("active_task"); !strings.Contains(schema.Show(v), "no task file has the named id (T-0999)") {
		t.Errorf("named: %s", schema.Show(v))
	}
	t.Setenv("CLAUDE_PROJECT_DIR", testpack.Project(t, tmp, "elsewhere"))
	doc, _ = Build(root, "dev")
	if v, _ := doc.Get("active_task"); schema.Show(v) != want["active_task"] {
		t.Errorf("another project's session: %s", schema.Show(v))
	}
	if txt := Text(doc); !strings.Contains(txt, "Active task: T-0901 (running)") || !strings.Contains(txt, "Status writes: command") {
		t.Errorf("text:\n%s", txt)
	}
}

// needs: the Claude Code floor first (kind tool, never a problem), each locked pack after it; a pack's
// needs.claude_code above Bonsai's own raises the floor (spec §7), read from the lock's declares.
func TestStatusNeeds(t *testing.T) {
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home := testpack.Isolate(t, tmp)
	src, shas := testpack.Fixture(t, tmp, "needy", map[string]string{
		".claude-plugin/plugin.json": `{"name": "needy", "description": "A fixture pack."}` + "\n",
		"bonsai/pack.yaml": "format: bonsai.pack/1\nid: needy\nversion: \"0.3.0\"\nneeds:\n  claude_code: \"2.1.400\"\n" +
			"files: []\nhooks: []\ndeny: []\n",
	})
	root := testpack.Project(t, tmp, "needy-project")
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test",
		Init: &engine.InitValues{Name: "needy", Source: src, Ref: shas[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(p); err != nil {
		t.Fatal(err)
	}
	doc, code := Build(root, "dev")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	want := `[{"kind":"tool","id":null,"name":"claude-code","source":null,"version":">=2.1.400"},` +
		`{"kind":"pack","id":"needy","name":null,"source":` + schema.Show(src) + `,"version":"0.3.0"}]`
	if v, _ := doc.Get("needs"); schema.Show(v) != want {
		t.Errorf("needs %s\nwant  %s", schema.Show(v), want)
	}
	if v, _ := doc.Get("problems"); schema.Show(v) != "[]" {
		t.Errorf("a floor is never a problem: %s", schema.Show(v))
	}
}

// fakeCLI answers claude plugin list.
type fakeCLI struct {
	list []engine.InstalledPlugin
	err  error
}

func (f fakeCLI) List(string) ([]engine.InstalledPlugin, error) { return f.list, f.err }
func (f fakeCLI) Install(string, string) (engine.InstallResult, error) {
	return engine.InstallResult{}, errors.New("status never installs")
}

// --full adds checks: the pack's newer release tags (git ls-remote of the pack's own repository, no network), its
// plugin as Claude Code reports it, Claude Code's version against the floor, and the MCP servers the packs' needs
// name (none in formats set 4); a pack whose plugin is not installed here is a needs entry of kind plugin. Asking
// nothing, or failing to ask, gives unknown, never a problem.
func TestStatusFull(t *testing.T) {
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home := testpack.Isolate(t, tmp)
	pack := testpack.Build(t, tmp)
	testpack.Tag(t, pack.Source, "v0.1.0", pack.A)
	testpack.Tag(t, pack.Source, "v0.2.0", pack.B)
	testpack.Tag(t, pack.Source, "v0.10.0", pack.C)
	testpack.Tag(t, pack.Source, "other-v9.0.0", pack.D)
	root := testpack.Project(t, tmp, "full")
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test", AllowExec: true,
		Init: &engine.InitValues{Name: "full", Source: pack.Source, Ref: "v0.1.0"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(p); err != nil {
		t.Fatal(err)
	}
	doc, code := BuildWith(root, "dev", Options{Full: true, Plugins: fakeCLI{}, Claude: func() (string, error) { return "2.1.300 (Claude Code)", nil }})
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	market := engine.MarketplaceName("full", []string{pack.A})
	want := `{"packs":[{"id":"demo-pack","version":"0.1.0","ref":"v0.1.0","newer":["v0.2.0","v0.10.0"],"why":null}],` +
		`"plugins":[{"id":"demo-pack","plugin":"demo-pack@` + market + `","installed":false,"why":null}],` +
		`"claude_code":{"version":"2.1.300","floor":"2.1.294","from":"bonsai","state":"ok","why":null},"mcp":[]}`
	if v, _ := doc.Get("checks"); schema.Show(v) != want {
		t.Errorf("checks %s\nwant   %s", schema.Show(v), want)
	}
	if v, _ := doc.Get("mode"); v != ModeFull {
		t.Errorf("mode %v", v)
	}
	if v, _ := doc.Get("needs"); !strings.Contains(schema.Show(v), `{"kind":"plugin","id":"demo-pack","name":null,`) {
		t.Errorf("a plugin not installed here is not a needs entry of kind plugin: %s", schema.Show(v))
	}
	// Installed at the lock's commit: kind pack. Claude Code older than the floor: old, never a problem.
	installed := fakeCLI{list: []engine.InstalledPlugin{{ID: "demo-pack@" + market, Version: pack.A[:12], Scope: "project", Enabled: true, ProjectPath: root}}}
	doc, _ = BuildWith(root, "dev", Options{Full: true, Plugins: installed, Claude: func() (string, error) { return "2.1.100", nil }})
	checkShape(t, doc, false)
	if v, _ := doc.Get("needs"); strings.Contains(schema.Show(v), `"kind":"plugin"`) {
		t.Errorf("installed, still kind plugin: %s", schema.Show(v))
	}
	if v, _ := doc.Get("checks"); !strings.Contains(schema.Show(v), `"state":"old"`) || !strings.Contains(schema.Show(v), `"installed":true`) {
		t.Errorf("checks %s", schema.Show(v))
	}
	if v, _ := doc.Get("problems"); schema.Show(v) != "[]" {
		t.Errorf("an old Claude Code is a problem: %s", schema.Show(v))
	}
	// Nothing to ask, and a source whose tags cannot be read: unknown, with why.
	doc, _ = BuildWith(root, "dev", Options{Full: true, Tags: func(string) ([]string, error) { return nil, errors.New("git ls-remote failed: no network") }})
	checkShape(t, doc, false)
	v, _ := doc.Get("checks")
	for _, w := range []string{`"newer":[],"why":"git ls-remote failed: no network"`, `"installed":null,"why":"Claude Code could not be asked: claude is not on the PATH"`,
		`"version":null,"floor":"2.1.294","from":"bonsai","state":"unknown","why":"Claude Code is not on the PATH"`} {
		if !strings.Contains(schema.Show(v), w) {
			t.Errorf("checks %s lacks %s", schema.Show(v), w)
		}
	}
	if v, _ := doc.Get("needs"); strings.Contains(schema.Show(v), `"kind":"plugin"`) {
		t.Errorf("not asked, yet kind plugin: %s", schema.Show(v))
	}
	// Without --full: no checks, offline, and Claude Code never asked.
	asked := false
	doc, _ = BuildWith(root, "dev", Options{Claude: func() (string, error) { asked = true; return "", nil }})
	if v, _ := doc.Get("checks"); v != nil || asked {
		t.Errorf("checks without --full: %s (Claude Code asked: %v)", schema.Show(v), asked)
	}
	if v, _ := doc.Get("mode"); v != Mode {
		t.Errorf("mode %v", v)
	}
}

// status --active: the active task alone, the same answer as status --json's active_task; Bonsai not reading the
// workspace is an error with its word.
func TestStatusActive(t *testing.T) {
	root, _ := project(t, strings.Replace(cfg, "person_only:", "documents:\n  task: work/tasks\n  run: work/runs\n  answers: work/answers.md\n"+
		"  memory: work/memory\n  protocols: work/protocols\nperson_only:", 1))
	t.Setenv("BONSAI_TASK", "")
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("work/tasks/T-0901-x.md", "---\nformat: bonsai.task/1\nid: T-0901\ntitle: x\nstatus: running\n---\n")
	a, e := Active(filepath.Join(root, "sub"))
	if e != nil || schema.Show(a.Object()) != `{"id":"T-0901","how":"running","why":null}` {
		t.Fatalf("active %s %v", schema.Show(a.Object()), e)
	}
	doc, _ := Build(root, "dev")
	if at, _ := doc.Get("active_task"); !schema.Equal(at, a.Object()) {
		t.Errorf("status --json's active_task %s, --active %s", schema.Show(at), schema.Show(a.Object()))
	}
	if msgs := schema.Validate(ActiveSchema(), a.Object()); len(msgs) > 0 {
		t.Errorf("does not fit active_task's schema: %v", msgs)
	}
	write("work/tasks/T-0902-y.md", "---\nformat: bonsai.task/1\nid: T-0902\ntitle: another\nstatus: running\n---\n")
	if a, _ := Active(root); a.ID != "" || !strings.Contains(schema.Show(a.Object()), "two or more tasks read running") {
		t.Errorf("two running: %s", schema.Show(a.Object()))
	}
	outside := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(outside))
	if _, e := Active(outside); e == nil || e.Code != "not-a-checkout" {
		t.Errorf("outside git: %v", e)
	}
}

// approve_first reads git history, so status's default leaves it out of problems (cheap, for a program that runs it
// often) and --full adds it, as check reports it.
func TestStatusHistoryOnlyWithFull(t *testing.T) {
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home := testpack.Isolate(t, tmp)
	src, shas := testpack.DeclaringPack(t, tmp)
	root := testpack.Project(t, tmp, "history")
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test",
		Init: &engine.InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(p); err != nil {
		t.Fatal(err)
	}
	task := "---\nformat: bonsai.task/1\nid: T-0901\ntitle: x\nstatus: running\nlane: full\n---\n"
	if err := os.MkdirAll(filepath.Join(root, "work", "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "work", "tasks", "T-0901-x.md"), []byte(task), 0o644); err != nil {
		t.Fatal(err)
	}
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "a task running with no approval")
	doc, _ := Build(root, "dev")
	if v, _ := doc.Get("problems"); strings.Contains(schema.Show(v), "approved") {
		t.Errorf("the default read git history: %s", schema.Show(v))
	}
	doc, _ = BuildWith(root, "dev", Options{Full: true})
	if v, _ := doc.Get("problems"); !strings.Contains(schema.Show(v), "T-0901 (work/tasks/T-0901-x.md) reads running in the lane full") {
		t.Errorf("--full: %s", schema.Show(v))
	}
}
