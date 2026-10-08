package guard

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const testID = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"

const testYAML = `format: bonsai.workspace/1
id: ` + testID + `
name: guard-test
packs: []
protected: [".claude/**", "bonsai.yaml", ".bonsai/lock.json", "protected.txt", "docs", "notes/"]
person_only: [".claude/**", "bonsai.yaml", ".bonsai/lock.json", "secret/*.key"]
never_edit: []
`

// project makes a linked project in a temporary folder: bonsai.yaml, a .git folder on branch main, protected.txt
// and free.txt. It gives the folder's real path.
func project(t *testing.T, yaml string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"bonsai.yaml":   yaml,
		".git/HEAD":     "ref: refs/heads/main\n",
		"protected.txt": "protected\n",
		"free.txt":      "free\n",
	}
	for name, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// payload builds a PreToolUse payload for tool with one tool_input field.
func payload(session, cwd, tool, field, value string) string {
	o := schema.Object{
		{Key: "session_id", Value: session},
		{Key: "transcript_path", Value: "/tmp/transcript.jsonl"},
		{Key: "cwd", Value: cwd},
		{Key: "permission_mode", Value: "acceptEdits"},
		{Key: "hook_event_name", Value: "PreToolUse"},
		{Key: "tool_name", Value: tool},
		{Key: "tool_input", Value: schema.Object{{Key: field, Value: value}, {Key: "old_string", Value: "a"}}},
		{Key: "tool_use_id", Value: "toolu_01TEST"},
	}
	b, err := schema.EncodeLine(o)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func editOf(session, dir, rel string) string {
	return payload(session, dir, "Edit", "file_path", filepath.Join(dir, filepath.FromSlash(rel)))
}

// env gives a Getenv with CLAUDE_PROJECT_DIR set to dir and the other variables given as pairs.
func env(dir string, kv ...string) func(string) string {
	m := map[string]string{ProjectEnv: dir}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return func(k string) string { return m[k] }
}

// guardRun runs Main with a payload and gives its exit code and stderr.
func guardRun(t *testing.T, getenv func(string) string, stdin io.Reader, budget time.Duration) (int, string) {
	t.Helper()
	var stderr bytes.Buffer
	code := Main(Options{Stdin: stdin, Stderr: &stderr, Getenv: getenv, Budget: budget})
	return code, stderr.String()
}

// records reads a log file's records, each checked against log.schema.json and for its field order.
func records(t *testing.T, dir, file string) []schema.Object {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(LogDir), file))
	if err != nil {
		t.Fatalf("the log: %v", err)
	}
	s, err := schema.Parse(mustSchema(t))
	if err != nil {
		t.Fatal(err)
	}
	var out []schema.Object
	for i, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		if len(line)+1 > MaxRecord {
			t.Errorf("record %d is %d bytes, over %d", i, len(line)+1, MaxRecord)
		}
		v, err := schema.Decode([]byte(line))
		if err != nil {
			t.Fatalf("record %d does not parse: %v\n%s", i, err, line)
		}
		if msgs := schema.Validate(s, v); len(msgs) != 0 {
			t.Errorf("record %d against log.schema.json: %v", i, msgs)
		}
		if msgs := schema.CheckOrder(s, v); len(msgs) != 0 {
			t.Errorf("record %d field order: %v", i, msgs)
		}
		o := v.(schema.Object)
		keys := o.Keys()
		if n := len(keys); n < 2 || keys[n-2] != "bonsai_path" || keys[n-1] != "bonsai_sha256" {
			t.Errorf("record %d does not end in bonsai_path, bonsai_sha256: %v", i, keys)
		}
		out = append(out, o)
	}
	return out
}

func mustSchema(t *testing.T) []byte {
	t.Helper()
	raw, err := formats.Schema("log")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestProtectedEditIsRefusedAndFreeEditAllowed(t *testing.T) {
	dir := project(t, testYAML)
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-1", dir, "protected.txt")), 0)
	if code != 2 {
		t.Fatalf("protected.txt: exit %d, want 2; stderr %q", code, stderr)
	}
	for _, want := range []string{`bonsai guard: Edit of "protected.txt" refused: it is on bonsai.yaml's protected list ("protected.txt")`, "\nnext: "} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
	asciiOnly(t, stderr)
	code, stderr = guardRun(t, env(dir), strings.NewReader(editOf("s-1", dir, "free.txt")), 0)
	if code != 0 || stderr != "" {
		t.Fatalf("free.txt: exit %d, stderr %q; want 0 and nothing", code, stderr)
	}
	recs := records(t, dir, "s-s-1.ndjson")
	if len(recs) != 2 {
		t.Fatalf("%d records, want 2", len(recs))
	}
	want := []struct{ decision, rule, target string }{{"deny", RuleProtected, "protected.txt"}, {"allow", RuleNotProtected, "free.txt"}}
	for i, w := range want {
		r := recs[i]
		if r.String("decision") != w.decision || r.String("rule") != w.rule || r.String("target") != w.target ||
			r.String("event") != "guard" || r.String("workspace") != testID || r.String("session") != "s-1" ||
			r.String("tool") != "Edit" || r.String("category") != "Edit" || r.String("branch") != "main" ||
			r.String("checkout") != filepath.Base(dir) || r.String("tool_use_id") != "toolu_01TEST" {
			t.Errorf("record %d: %s", i, schema.Show(r))
		}
	}
	if recs[0].String("text") == "" {
		t.Errorf("a refusal's record carries no text")
	}
	if v, _ := recs[1].Get("text"); v != nil {
		t.Errorf("an allow's record has text %v", v)
	}
	// The binary's path on every record, its SHA-256 on the session file's first only.
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if recs[0].String("bonsai_path") != filepath.ToSlash(exe) || recs[0].String("bonsai_sha256") != hex.EncodeToString(sum[:]) {
		t.Errorf("first record names %s %s, want %s %x", recs[0].String("bonsai_path"), recs[0].String("bonsai_sha256"), filepath.ToSlash(exe), sum)
	}
	if v, _ := recs[1].Get("bonsai_sha256"); v != nil || recs[1].String("bonsai_path") != filepath.ToSlash(exe) {
		t.Errorf("second record: path %s, hash %v; want the path and null", recs[1].String("bonsai_path"), v)
	}
}

func asciiOnly(t *testing.T, s string) {
	t.Helper()
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7e || (s[i] < 0x20 && s[i] != '\n') {
			t.Fatalf("byte %d is %#x, not printable ASCII: %q", i, s[i], s)
		}
	}
}

func TestDecide(t *testing.T) {
	dir := project(t, testYAML)
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(filepath.Dir(dir), "elsewhere.txt")
	cases := []struct {
		name, tool, path, cwd string
		rule, target          string
	}{
		{"protected", "Edit", "protected.txt", "", RuleProtected, "protected.txt"},
		{"write protected", "Write", "protected.txt", "", RuleProtected, "protected.txt"},
		{"multiedit protected", "MultiEdit", "protected.txt", "", RuleProtected, "protected.txt"},
		{"notebook protected", "NotebookEdit", "protected.txt", "", RuleProtected, "protected.txt"},
		{"free", "Edit", "free.txt", "", RuleNotProtected, "free.txt"},
		{"new free file", "Write", "sub/new.txt", "", RuleNotProtected, "sub/new.txt"},
		{"glob", "Write", ".claude/settings.local.json", "", RulePersonOnly, ".claude/settings.local.json"},
		{"person only first", "Edit", "bonsai.yaml", "", RulePersonOnly, "bonsai.yaml"},
		{"person only glob", "Write", "secret/a.key", "", RulePersonOnly, "secret/a.key"},
		{"person only glob miss", "Write", "secret/a.txt", "", RuleNotProtected, "secret/a.txt"},
		{"folder named", "Write", "docs/guide/a.md", "", RuleProtected, "docs/guide/a.md"},
		{"folder with slash", "Write", "notes/today.md", "", RuleProtected, "notes/today.md"},
		{"dot dot back in", "Edit", "sub/../protected.txt", "", RuleProtected, "protected.txt"},
		{"relative to cwd", "Edit", "../protected.txt", "sub", RuleProtected, "protected.txt"},
		{"relative to project", "Edit", "protected.txt", "-", RuleProtected, "protected.txt"},
		{"outside", "Edit", outside, "", RuleOutside, ""},
		{"the project folder itself", "Write", ".", "", RuleOutside, ""},
		{"lock", "Edit", ".bonsai/lock.json", "", RulePersonOnly, ".bonsai/lock.json"},
		{"bash", "Bash", "", "", RuleShell, ""},
		{"powershell", "PowerShell", "", "", RuleShell, ""},
		{"other tool", "Read", "", "", RuleOtherTool, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := &Input{Event: "PreToolUse", Tool: c.tool, Cwd: dir}
			switch c.cwd {
			case "-":
				in.Cwd = ""
			case "":
			default:
				in.Cwd = filepath.Join(dir, c.cwd)
			}
			if c.path != "" {
				in.Path = c.path
				if !filepath.IsAbs(c.path) && c.cwd == "" {
					in.Path = filepath.Join(dir, filepath.FromSlash(c.path))
				}
			}
			d := Decide(in, dir, cfg)
			if d.Rule != c.rule || d.Target != c.target {
				t.Fatalf("%s %s: rule %s target %q, want %s %q", c.tool, c.path, d.Rule, d.Target, c.rule, c.target)
			}
			denied := c.rule == RuleProtected || c.rule == RulePersonOnly
			if d.Allow == denied {
				t.Fatalf("allow is %v for rule %s", d.Allow, d.Rule)
			}
			if denied && (d.Why == "" || d.Next == "" || !strings.HasPrefix(d.Why, c.tool+" of ")) {
				t.Errorf("refusal without its reason first or a next step: %+v", d)
			}
		})
	}
}

// A path starting with ~ is judged in the user's home folder too.
func TestDecideReadsTilde(t *testing.T) {
	dir := project(t, testYAML)
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", filepath.Dir(dir))
	t.Setenv("USERPROFILE", filepath.Dir(dir))
	d := Decide(&Input{Tool: "Edit", Path: "~/" + filepath.Base(dir) + "/protected.txt"}, dir, cfg)
	if d.Allow || d.Rule != RuleProtected {
		t.Fatalf("%+v", d)
	}
	if d := Decide(&Input{Tool: "Edit", Path: "~/" + filepath.Base(dir) + "/free.txt"}, dir, cfg); !d.Allow {
		t.Fatalf("%+v", d)
	}
}

func TestDecideOnACaseInsensitiveSystem(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Skip("letter case is folded where the file system ignores it: Windows and macOS; matchGlob's own test covers the fold")
	}
	dir := project(t, testYAML)
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"PROTECTED.TXT", "Protected.Txt", ".CLAUDE/settings.json"} {
		d := Decide(&Input{Tool: "Edit", Path: filepath.Join(dir, p)}, dir, cfg)
		if d.Allow {
			t.Errorf("%s allowed: %+v", p, d)
		}
	}
}

func TestDecideOnWindowsForms(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("these forms name a file only on Windows; windowsForms' own test reads them on every system")
	}
	dir := project(t, testYAML)
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	slash := filepath.ToSlash(dir)
	msys := "/" + strings.ToLower(slash[:1]) + slash[2:]
	for _, p := range []string{
		`\\?\` + dir + `\protected.txt`,
		dir + `\protected.txt.`,
		dir + `\protected.txt  `,
		dir + `\protected.txt::$DATA`,
		slash + "/protected.txt",
		msys + "/protected.txt",
		dir + `\sub\..\PROTECTED.txt`,
	} {
		d := Decide(&Input{Tool: "Write", Path: p}, dir, cfg)
		if d.Allow || d.Rule != RuleProtected {
			t.Errorf("%s: %+v, want refused as protected", p, d)
		}
	}
}

// symlink makes a symbolic link, or skips the test where the system will not (Windows without the privilege).
func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symbolic links need a privilege on Windows (%v); Linux and CI run this test", err)
		}
		t.Fatal(err)
	}
}

func TestDecideFollowsLinks(t *testing.T) {
	dir := project(t, testYAML)
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	// A link to the protected file, a link to a protected file not yet made, and a link to the project folder.
	symlink(t, "protected.txt", filepath.Join(dir, "free-link.txt"))
	symlink(t, filepath.Join(dir, "secret", "new.key"), filepath.Join(dir, "dangling.txt"))
	outside := t.TempDir()
	symlink(t, dir, filepath.Join(outside, "project-link"))
	for _, p := range []string{
		filepath.Join(dir, "free-link.txt"),
		filepath.Join(dir, "dangling.txt"),
		filepath.Join(outside, "project-link", "protected.txt"),
	} {
		d := Decide(&Input{Tool: "Write", Path: p}, dir, cfg)
		if d.Allow {
			t.Errorf("%s allowed: %+v", p, d)
		}
	}
	// The project reached through a link: CLAUDE_PROJECT_DIR is the link, the path the real folder.
	link := filepath.Join(outside, "project-link")
	d := Decide(&Input{Tool: "Edit", Path: filepath.Join(dir, "protected.txt")}, link, cfg)
	if d.Allow {
		t.Errorf("a protected path under the real folder of a linked project allowed: %+v", d)
	}
}

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		glob, rel string
		fold, ok  bool
	}{
		{"protected.txt", "protected.txt", false, true},
		{"protected.txt", "Protected.txt", false, false},
		{"protected.txt", "Protected.txt", true, true},
		{".claude/**", ".claude/settings.json", false, true},
		{".claude/**", ".claude/a/b/c.json", false, true},
		{".claude/**", ".claude", false, true},
		{".claude/**", ".claudex/a", false, false},
		{"**/*.key", "a/b/c.key", false, true},
		{"**/*.key", "c.key", false, true},
		{"docs", "docs/a.md", false, true},
		{"docs", "docsx/a.md", false, false},
		{"notes/", "notes/x", false, true},
		{"*.md", "README.md", false, true},
		{"*.md", "sub/README.md", false, false},
		{"src/*/gen.go", "src/a/gen.go", false, true},
		{"src/*/gen.go", "src/a/b/gen.go", false, false},
		{"a/**/z", "a/z", false, true},
		{"a/**/z", "a/b/c/z", false, true},
		{"?.txt", "a.txt", false, true},
		{"[ab].txt", "b.txt", false, true},
		{"[ab].txt", "c.txt", false, false},
	}
	for _, c := range cases {
		if err := checkGlob(c.glob); err != nil {
			t.Fatalf("%s: %v", c.glob, err)
		}
		if got := matchGlob(c.glob, c.rel, c.fold); got != c.ok {
			t.Errorf("matchGlob(%q, %q, fold %v) = %v, want %v", c.glob, c.rel, c.fold, got, c.ok)
		}
	}
	for _, bad := range []string{"[a.txt", "a/[/b"} {
		if checkGlob(bad) == nil {
			t.Errorf("checkGlob(%q) took it", bad)
		}
	}
}

func TestBadGlobBlocks(t *testing.T) {
	dir := project(t, strings.Replace(testYAML, `"protected.txt", "docs"`, `"protected.txt", "[docs"`, 1))
	cfg, err := workspace.LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	d := Decide(&Input{Tool: "Edit", Path: filepath.Join(dir, "free.txt")}, dir, cfg)
	if d.Allow || d.Rule != RuleBadGlob || !strings.Contains(d.Why, `"[docs"`) {
		t.Fatalf("%+v", d)
	}
}

func TestWindowsForms(t *testing.T) {
	cases := map[string][]string{
		`\\?\C:\p\protected.txt`:    {`C:\p\protected.txt`},
		`\\.\C:\p\protected.txt`:    {`C:\p\protected.txt`},
		`\\?\UNC\srv\share\a.txt`:   {`\\srv\share\a.txt`},
		`/c/p/protected.txt`:        {`C:\p\protected.txt`},
		`/cygdrive/d/p/x.txt`:       {`D:\p\x.txt`},
		`C:\p\protected.txt.`:       {`C:\p\protected.txt`},
		`C:\p\protected.txt . .`:    {`C:\p\protected.txt`},
		`C:\p\protected.txt::$DATA`: {`C:\p\protected.txt`},
		`C:\p\protected.txt:x`:      {`C:\p\protected.txt`},
		`C:\p.\sub \protected.txt`:  {`C:\p\sub\protected.txt`},
		`C:/p/protected.txt`:        {`C:\p\protected.txt`},
	}
	for in, wants := range cases {
		got := windowsForms(in)
		for _, w := range wants {
			found := false
			for _, g := range got {
				found = found || g == w
			}
			if !found {
				t.Errorf("windowsForms(%q) = %q, lacks %q", in, got, w)
			}
		}
	}
	if got := win32Names(`..\x.`); got != `..\x` {
		t.Errorf("win32Names keeps .. as is: %q", got)
	}
}

func TestParseInput(t *testing.T) {
	good := payload("0b6f3c2e-5d1a-4e8b-9c7f-2a4d6e8f0a1b", "/p", "Edit", "file_path", "/p/a.txt")
	in, err := ParseInput([]byte(good))
	if err != nil || in.Tool != "Edit" || in.Path != "/p/a.txt" || in.Cwd != "/p" || in.ToolUseID != "toolu_01TEST" {
		t.Fatalf("%+v %v", in, err)
	}
	if in, err := ParseInput(append([]byte{0xEF, 0xBB, 0xBF}, good...)); err != nil || in.Path != "/p/a.txt" {
		t.Errorf("with a byte order mark: %+v %v", in, err)
	}
	if in, err := ParseInput([]byte(payload("s", "/p", "NotebookEdit", "notebook_path", "/p/n.ipynb"))); err != nil || in.Path != "/p/n.ipynb" {
		t.Errorf("NotebookEdit: %+v %v", in, err)
	}
	if in, err := ParseInput([]byte(payload("s", "/p", "Bash", "command", "ls"))); err != nil || in.Path != "" {
		t.Errorf("Bash: %+v %v", in, err)
	}
	sub := `{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"/p/x"},` +
		`"agent_id":"a1","agent_type":"workflow:builder","new_field":{"deep":[1,2]}}`
	if in, err := ParseInput([]byte(sub)); err != nil || in.AgentID != "a1" || in.AgentType != "workflow:builder" {
		t.Errorf("a subagent's payload with an unknown field: %+v %v", in, err)
	}
	bad := map[string]string{
		"empty":            "",
		"spaces":           "  \n",
		"not JSON":         "edit protected.txt",
		"an array":         `[{"tool_name":"Edit"}]`,
		"two documents":    `{"hook_event_name":"PreToolUse"} {}`,
		"duplicate key":    `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/free.txt","file_path":"/p/protected.txt"}}`,
		"no event":         `{"tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"other event":      `{"hook_event_name":"PostToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"no tool":          `{"hook_event_name":"PreToolUse","tool_input":{"file_path":"/p/a"}}`,
		"empty tool":       `{"hook_event_name":"PreToolUse","tool_name":"","tool_input":{}}`,
		"tool not text":    `{"hook_event_name":"PreToolUse","tool_name":7,"tool_input":{}}`,
		"no tool_input":    `{"hook_event_name":"PreToolUse","tool_name":"Bash"}`,
		"tool_input text":  `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":"/p/a"}`,
		"no file_path":     `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"path":"/p/a"}}`,
		"file_path null":   `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":null}}`,
		"file_path number": `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":3}}`,
		"empty file_path":  `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":""}}`,
		"NUL in path":      `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a\u0000b"}}`,
		"session escapes":  `{"session_id":"../../x","hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"session slash":    `{"session_id":"a/b","hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"bad UTF-8":        "{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\"/p/\xff\"}}",
		"notebook no path": `{"hook_event_name":"PreToolUse","tool_name":"NotebookEdit","tool_input":{"file_path":"/p/n.ipynb"}}`,
	}
	for name, raw := range bad {
		if in, err := ParseInput([]byte(raw)); err == nil {
			t.Errorf("%s: read as %+v", name, in)
		} else {
			asciiOnly(t, err.Error())
		}
	}
}

func TestNotLinkedAllowsAndRecordsNothing(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, start := range []string{dir, filepath.Join(dir, "sub")} {
		code, stderr := guardRun(t, env(start), strings.NewReader("not even a payload"), 0)
		if code != 0 || stderr != "" {
			t.Errorf("from %s: exit %d, stderr %q; want 0 and nothing", start, code, stderr)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".bonsai")); !os.IsNotExist(err) {
		t.Errorf("a project with no bonsai.yaml got .bonsai/: %v", err)
	}
}

func TestStartedInASubfolderFindsTheProject(t *testing.T) {
	dir := project(t, testYAML)
	sub := filepath.Join(dir, "sub", "deeper")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	code, stderr := guardRun(t, env(sub), strings.NewReader(editOf("s-2", dir, "protected.txt")), 0)
	if code != 2 || !strings.Contains(stderr, `"protected.txt" refused`) {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

func TestBlocksWhatItCannotReadOrDecide(t *testing.T) {
	dir := project(t, testYAML)
	cases := []struct {
		name   string
		getenv func(string) string
		stdin  string
		rule   string
		inErr  string
	}{
		{"no project variable", env(""), editOf("s-3", dir, "free.txt"), RuleNoProject, "CLAUDE_PROJECT_DIR is not set"},
		{"project folder is a file", env(filepath.Join(dir, "free.txt")), editOf("s-3", dir, "free.txt"), RuleNoProject, "cannot be read"},
		{"bad input", env(dir), `{"hook_event_name":"PreToolUse"`, RuleBadInput, "the hook input is not one JSON document"},
		{"empty input", env(dir), "", RuleBadInput, "the hook input is empty"},
		{"wrong event", env(dir), strings.Replace(editOf("s-3", dir, "free.txt"), "PreToolUse", "Stop", 1), RuleBadInput, `"Stop" payload`},
		{"a panic inside", func(k string) string {
			if k == ProjectEnv {
				panic("boom")
			}
			return ""
		}, editOf("s-3", dir, "free.txt"), RuleInternal, "the guard failed inside (boom)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, stderr := guardRun(t, c.getenv, strings.NewReader(c.stdin), 0)
			if code != 2 || !strings.Contains(stderr, c.inErr) || !strings.Contains(stderr, "\nnext: ") {
				t.Fatalf("exit %d, stderr %q; want 2 with %q and a next step", code, stderr, c.inErr)
			}
			asciiOnly(t, stderr)
		})
	}
	// A bad input is recorded, in the day's file (no session known).
	day := recordFile("", time.Now())
	recs := records(t, dir, day)
	if len(recs) != 3 || recs[0].String("rule") != RuleBadInput || recs[0].String("decision") != "deny" {
		t.Fatalf("records %d: %v", len(recs), recs)
	}
	if v, _ := recs[0].Get("session"); v != nil {
		t.Errorf("session %v, want null", v)
	}
	// A stdin that is a terminal (the guard run by hand) is not waited on.
	var stderr bytes.Buffer
	code := Main(Options{Stdin: strings.NewReader(""), StdinTerminal: true, Stderr: &stderr, Getenv: env(dir)})
	if code != 2 || !strings.Contains(stderr.String(), "is a terminal") {
		t.Errorf("terminal: exit %d, %q", code, stderr.String())
	}
}

func TestBadConfigBlocksWithNoRecord(t *testing.T) {
	dir := project(t, "format: bonsai.workspace/1\nid: nonsense\n")
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-4", dir, "free.txt")), 0)
	if code != 2 || !strings.Contains(stderr, "bonsai.yaml cannot be read") || !strings.Contains(stderr, "\nnext: tell the person: ") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, ".bonsai")); !os.IsNotExist(err) {
		t.Errorf("a record without a workspace id was attempted: %v", err)
	}
}

// A stdin that never ends: the work is still reading when the budget runs out, and the guard blocks by itself.
func TestOverTimeBlocks(t *testing.T) {
	dir := project(t, testYAML)
	// The pipe's writer is never closed: the work stays blocked in its read, as a stdin that never ends would, and
	// writes nothing more into the test's folder.
	pr, _ := io.Pipe()
	start := time.Now()
	code, stderr := guardRun(t, env(dir), pr, 150*time.Millisecond)
	took := time.Since(start)
	if code != 2 || !strings.Contains(stderr, "the guard ran past its own 150ms limit") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if took > 150*time.Millisecond+recordWait+500*time.Millisecond {
		t.Errorf("answered after %v", took)
	}
	recs := records(t, dir, recordFile("", time.Now()))
	if len(recs) != 1 || recs[0].String("rule") != RuleOverTime || recs[0].String("decision") != "deny" {
		t.Fatalf("records: %v", recs)
	}
}

func TestAnAllowItCannotRecordIsBlocked(t *testing.T) {
	dir := project(t, testYAML)
	// A file where the log folder should be.
	if err := os.MkdirAll(filepath.Join(dir, ".bonsai", "local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(LogDir)), []byte("in the way"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-5", dir, "free.txt")), 0)
	if code != 2 || !strings.Contains(stderr, "the guard would allow this call but cannot write its record") {
		t.Fatalf("free.txt: exit %d, stderr %q", code, stderr)
	}
	code, stderr = guardRun(t, env(dir), strings.NewReader(editOf("s-5", dir, "protected.txt")), 0)
	if code != 2 || !strings.Contains(stderr, "its record in .bonsai/local/log could not be written either") {
		t.Fatalf("protected.txt: exit %d, stderr %q", code, stderr)
	}
}

func TestLongValuesStayUnderTheRecordCap(t *testing.T) {
	dir := project(t, testYAML)
	long := strings.Repeat("d/", 600) + "x.txt"
	in := strings.Replace(editOf("s-6", dir, long), "toolu_01TEST", "toolu_"+strings.Repeat("Z", 1500), 1)
	code, stderr := guardRun(t, env(dir, "BONSAI_TASK", strings.Repeat("T", 900), "BONSAI_ROLE", "builder"), strings.NewReader(in), 0)
	if code != 0 {
		t.Fatalf("exit %d, %q", code, stderr)
	}
	recs := records(t, dir, "s-s-6.ndjson")
	if len(recs) != 1 || recs[0].String("role") != "builder" {
		t.Fatalf("%v", recs)
	}
}

// Calls running side by side append whole lines: none is lost or mixed with another.
func TestParallelCallsAppendWholeRecords(t *testing.T) {
	dir := project(t, testYAML)
	const n = 24
	var wg sync.WaitGroup
	codes := make([]int, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rel := "free.txt"
			if i%2 == 1 {
				rel = "protected.txt"
			}
			codes[i], _ = guardRun(t, env(dir), strings.NewReader(editOf("s-7", dir, rel)), 0)
		}(i)
	}
	wg.Wait()
	for i, c := range codes {
		if want := []int{0, 2}[i%2]; c != want {
			t.Errorf("call %d: exit %d, want %d", i, c, want)
		}
	}
	recs := records(t, dir, "s-s-7.ndjson")
	if len(recs) != n {
		t.Fatalf("%d records, want %d", len(recs), n)
	}
	hashed := 0
	for _, r := range recs {
		if r.String("bonsai_sha256") != "" {
			hashed++
		}
	}
	if hashed != 1 {
		t.Errorf("%d records carry the hash, want 1 (the file's first)", hashed)
	}
}

func TestRecordIDsAreUUIDs(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		id := newID()
		if len(id) != 36 || id[14] != '4' || !strings.ContainsRune("89ab", rune(id[19])) || seen[id] {
			t.Fatalf("id %s", id)
		}
		seen[id] = true
	}
	if got := cut("abcdef", 4); got != "a..." {
		t.Errorf("cut: %q", got)
	}
	if got := category("NotebookEdit") + category("Bash") + category("Glob"); got != "EditShellOther" {
		t.Errorf("category: %q", got)
	}
	if got := fmt.Sprint(checkoutName(string(filepath.Separator))); got != "" {
		t.Errorf("checkout at the root: %q", got)
	}
}
