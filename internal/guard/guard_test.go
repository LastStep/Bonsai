package guard

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/hookin"
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

// TestMain gives the package's tests a Bonsai home of their own, so no test reaches the real one (the salt, this
// machine's records); project gives each test its own home in turn.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "bonsai-guard-home-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv(workspace.HomeEnv, home); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

// testNow is the clock of the runs whose day's file a test reads, so a run that crosses midnight (UTC) finds its
// records in the file it wrote.
var testNow = time.Date(2026, 10, 11, 23, 59, 59, 0, time.UTC)

// project makes a linked project in a temporary folder: bonsai.yaml, a .git folder on branch main, protected.txt
// and free.txt, recorded on this machine as bonsai init records it (workspace.RecordCheckout), in a Bonsai home of the
// test's own (BONSAI_HOME), so it is a verified main checkout. It gives the folder's real path.
func project(t *testing.T, yaml string) string {
	t.Helper()
	dir := unrecordedProject(t, yaml)
	home := t.TempDir()
	t.Setenv(workspace.HomeEnv, home)
	if m := regexp.MustCompile(`(?m)^id: (ws-[a-z0-9]{26})`).FindStringSubmatch(yaml); m != nil {
		if _, err := workspace.RecordCheckout(home, dir, m[1], time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// unrecordedProject is project with no record on this machine: a clone that never ran bonsai update here.
func unrecordedProject(t *testing.T, yaml string) string {
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
	m := map[string]string{hookin.ProjectEnv: dir}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return func(k string) string { return m[k] }
}

// untimed is the budget of every run that ends on the guard's own decision: a minute, which no run comes near, so a
// loaded machine can never turn its answer into over-time. Nothing in such a run can hang (its payload is a
// strings.Reader, and a busy log file is retried for openWait at most). Only the over-time tests run on a short budget.
const untimed = time.Minute

// guardRun runs Main with a payload and gives its exit code and stderr. Its clock is testNow.
func guardRun(t *testing.T, getenv func(string) string, stdin io.Reader, budget time.Duration) (int, string) {
	t.Helper()
	var stderr bytes.Buffer
	code := Main(Options{Stdin: stdin, Stderr: &stderr, Getenv: getenv, Budget: budget, Now: func() time.Time { return testNow }})
	return code, stderr.String()
}

// decide judges one payload as a run does with root the session's project, each file read afresh.
func decide(in *hookin.Payload, root string) Decision {
	j := newJudge(guardHome())
	j.session = j.project(root)
	d, _ := j.decide(in, root)
	return d
}

// dayRecords reads the records of the one day's file (w-<date>.ndjson) in a project's log: a run with no session
// writes there, by its own clock.
func dayRecords(t *testing.T, dir string) []schema.Object {
	t.Helper()
	names, err := filepath.Glob(filepath.Join(dir, filepath.FromSlash(LogDir), "w-*.ndjson"))
	if err != nil || len(names) != 1 {
		t.Fatalf("day files %v (%v), want one", names, err)
	}
	return records(t, dir, filepath.Base(names[0]))
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
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-1", dir, "protected.txt")), untimed)
	if code != 2 {
		t.Fatalf("protected.txt: exit %d, want 2; stderr %q", code, stderr)
	}
	for _, want := range []string{`bonsai guard: Edit of "protected.txt" refused: it is on bonsai.yaml's protected list ("protected.txt")`, "\nnext: "} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
	asciiOnly(t, stderr)
	code, stderr = guardRun(t, env(dir), strings.NewReader(editOf("s-1", dir, "free.txt")), untimed)
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
	// The binary's path (from ~/ under the user's home) and its SHA-256 on the record that made the file only.
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if want := fromHome(exe); recs[0].String("bonsai_path") != want || recs[0].String("bonsai_sha256") != hex.EncodeToString(sum[:]) {
		t.Errorf("first record names %s %s, want %s %x", recs[0].String("bonsai_path"), recs[0].String("bonsai_sha256"), want, sum)
	}
	if p, _ := recs[1].Get("bonsai_path"); p != nil {
		t.Errorf("second record names the binary %v; want null", p)
	}
	if v, _ := recs[1].Get("bonsai_sha256"); v != nil {
		t.Errorf("second record names the hash %v; want null", v)
	}
	// The input hash: the salt's HMAC over the tool input, as the recorder's tool_start has it.
	key, err := workspace.Salt()
	if err != nil {
		t.Fatal(err)
	}
	in, err := hookin.Parse([]byte(editOf("s-1", dir, "free.txt")))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := recs[1].String("input_hash"), hookin.InputHash(key, in.Input); got != want || len(got) != 16 {
		t.Errorf("input_hash %q, want %q", got, want)
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
		{"glob", "Write", ".claude/agents/x.md", "", RulePersonOnly, ".claude/agents/x.md"},
		{"floor never granted", "Write", ".claude/settings.local.json", "", RuleNeverGranted, ".claude/settings.local.json"},
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
			in := &hookin.Payload{Event: "PreToolUse", Tool: c.tool, Cwd: dir}
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
			d := decide(in, dir)
			if d.Rule != c.rule || d.Target != c.target {
				t.Fatalf("%s %s: rule %s target %q, want %s %q", c.tool, c.path, d.Rule, d.Target, c.rule, c.target)
			}
			denied := c.rule == RuleProtected || c.rule == RulePersonOnly || c.rule == RuleNeverGranted
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
	t.Setenv("HOME", filepath.Dir(dir))
	t.Setenv("USERPROFILE", filepath.Dir(dir))
	d := decide(&hookin.Payload{Tool: "Edit", Path: "~/" + filepath.Base(dir) + "/protected.txt"}, dir)
	if d.Allow || d.Rule != RuleProtected {
		t.Fatalf("%+v", d)
	}
	if d := decide(&hookin.Payload{Tool: "Edit", Path: "~/" + filepath.Base(dir) + "/free.txt"}, dir); !d.Allow {
		t.Fatalf("%+v", d)
	}
}

func TestDecideOnACaseInsensitiveSystem(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Skip("letter case is folded where the file system ignores it: Windows and macOS; matchGlob's own test covers the fold")
	}
	dir := project(t, testYAML)
	for _, p := range []string{"PROTECTED.TXT", "Protected.Txt", ".CLAUDE/settings.json"} {
		d := decide(&hookin.Payload{Tool: "Edit", Path: filepath.Join(dir, p)}, dir)
		if d.Allow {
			t.Errorf("%s allowed: %+v", p, d)
		}
	}
	// The project's own folder written in another letter case is the same project, verified as this machine recorded it.
	for _, p := range []string{filepath.Join(strings.ToUpper(dir), "PROTECTED.TXT"), filepath.Join(strings.ToLower(dir), "protected.txt")} {
		if d := decide(&hookin.Payload{Tool: "Edit", Path: p}, dir); d.Allow || d.Rule != RuleProtected {
			t.Errorf("%s: %+v, want refused as protected", p, d)
		}
	}
}

func TestDecideOnWindowsForms(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("these forms name a file only on Windows; windowsForms' own test reads them on every system")
	}
	dir := project(t, testYAML)
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
		d := decide(&hookin.Payload{Tool: "Write", Path: p}, dir)
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
		d := decide(&hookin.Payload{Tool: "Write", Path: p}, dir)
		if d.Allow {
			t.Errorf("%s allowed: %+v", p, d)
		}
	}
	// The project reached through a link: CLAUDE_PROJECT_DIR is the link, the path the real folder.
	link := filepath.Join(outside, "project-link")
	d := decide(&hookin.Payload{Tool: "Edit", Path: filepath.Join(dir, "protected.txt")}, link)
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
	d := decide(&hookin.Payload{Tool: "Edit", Path: filepath.Join(dir, "free.txt")}, dir)
	if d.Allow || d.Rule != RuleBadGlob || !strings.Contains(d.Why, `"[docs"`) {
		t.Fatalf("%+v", d)
	}
}

func TestWindowsForms(t *testing.T) {
	cases := map[string][]string{
		`\\?\C:\p\protected.txt`:                      {`C:\p\protected.txt`},
		`\\.\C:\p\protected.txt`:                      {`C:\p\protected.txt`},
		`\\?\UNC\srv\share\a.txt`:                     {`\\srv\share\a.txt`},
		`/c/p/protected.txt`:                          {`C:\p\protected.txt`},
		`/cygdrive/d/p/x.txt`:                         {`D:\p\x.txt`},
		`C:\p\protected.txt.`:                         {`C:\p\protected.txt`},
		`C:\p\protected.txt . .`:                      {`C:\p\protected.txt`},
		`C:\p\protected.txt::$DATA`:                   {`C:\p\protected.txt`},
		`C:\p\protected.txt:x`:                        {`C:\p\protected.txt`},
		`C:\p.\sub \protected.txt`:                    {`C:\p\sub\protected.txt`},
		`C:/p/protected.txt`:                          {`C:\p\protected.txt`},
		`C:\p\secrets::$INDEX_ALLOCATION\key.txt`:     {`C:\p\secrets\key.txt`},
		`C:\p\secrets:$I30:$INDEX_ALLOCATION\key.txt`: {`C:\p\secrets\key.txt`},
		`secrets::$INDEX_ALLOCATION\new.txt`:          {`secrets\new.txt`},
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
	// A device or volume form is not cut down to a relative path: finalPath asks Windows where it leads.
	for _, p := range []string{`\\?\GLOBALROOT\Device\HarddiskVolume3\p\protected.txt`, `\\?\Volume{0b6f3c2e}\p\protected.txt`} {
		for _, f := range windowsForms(p) {
			if !strings.HasPrefix(f, `\\`) {
				t.Errorf("windowsForms(%q) gives %q, which reads as a relative path", p, f)
			}
		}
	}
}

// The guard reads its payload through the hook adapter (internal/hookin, whose own tests read every event and tool)
// and blocks whatever it cannot read: a payload that does not read, any fault in a value it reads, an event other
// than PreToolUse. Each refusal is bad-input, says what is wrong after "the hook input", in ASCII, and names a next
// step; a payload that reads is decided.
func TestReadsThroughTheAdapter(t *testing.T) {
	dir := project(t, testYAML)
	free := filepath.ToSlash(filepath.Join(dir, "free.txt"))
	good := map[string]string{
		"an edit":           payload("0b6f3c2e-5d1a-4e8b-9c7f-2a4d6e8f0a1b", dir, "Edit", "file_path", free),
		"a byte order mark": "\xEF\xBB\xBF" + payload("s", dir, "Edit", "file_path", free),
		"a notebook":        payload("s", dir, "NotebookEdit", "notebook_path", free),
		"a shell command":   payload("s", dir, "Bash", "command", "ls"),
		"an unknown tool":   payload("s", dir, "SomeNewTool", "anything", "x"),
		"a subagent's, with fields of later versions": `{"session_id":"s","hook_event_name":"PreToolUse","tool_name":"Write",` +
			`"tool_input":{"file_path":"` + free + `","content":"x"},"agent_id":"a1","agent_type":"workflow:builder",` +
			`"prompt_id":"p","effort":{"level":"high"},"new_field":{"deep":[1,2],"deep":"twice, unread"},"bad":"` + "\xff" + `"}`,
	}
	for name, raw := range good {
		if code, stderr := guardRun(t, env(dir), strings.NewReader(raw), untimed); code != 0 || stderr != "" {
			t.Errorf("%s: exit %d, stderr %q; want it allowed", name, code, stderr)
		}
	}
	bad := map[string]string{
		"empty":            "",
		"spaces":           "  \n",
		"not JSON":         "edit protected.txt",
		"an array":         `[{"tool_name":"Edit"}]`,
		"two documents":    `{"hook_event_name":"PreToolUse"} {}`,
		"a key twice":      `{"hook_event_name":"PreToolUse","tool_name":"Edit","tool_name":"Read","tool_input":{"file_path":"/p/a"}}`,
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
		"NUL in the cwd":   `{"cwd":"/p\u0000","hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"no command":       `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"description":"x"}}`,
		"session escapes":  `{"session_id":"../../x","hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"session slash":    `{"session_id":"a/b","hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"session dot":      `{"session_id":"a.b","hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"session number":   `{"session_id":5,"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
		"bad UTF-8":        "{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\"/p/\xff\"}}",
		"bad UTF-8, event": "{\"hook_event_name\":\"PreToolUse\xff\",\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\"/p/a\"}}",
		"notebook no path": `{"hook_event_name":"PreToolUse","tool_name":"NotebookEdit","tool_input":{"file_path":"/p/n.ipynb"}}`,
		"agent id number":  `{"agent_id":1,"hook_event_name":"PreToolUse","tool_name":"Edit","tool_input":{"file_path":"/p/a"}}`,
	}
	for name, raw := range bad {
		code, stderr := guardRun(t, env(dir), strings.NewReader(raw), untimed)
		if code != 2 || !strings.Contains(stderr, "bonsai guard: the hook input ") || !strings.Contains(stderr, "\nnext: ") {
			t.Errorf("%s: exit %d, stderr %q; want it blocked as bad input", name, code, stderr)
		}
		asciiOnly(t, stderr)
	}
	recs := dayRecords(t, dir)
	if len(recs) != len(bad) {
		t.Fatalf("%d records of bad input, want %d", len(recs), len(bad))
	}
	for _, r := range recs {
		if r.String("rule") != RuleBadInput || r.String("decision") != "deny" {
			t.Errorf("record: %s", schema.Show(r))
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
		code, stderr := guardRun(t, env(start), strings.NewReader("not even a payload"), untimed)
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
	code, stderr := guardRun(t, env(sub), strings.NewReader(editOf("s-2", dir, "protected.txt")), untimed)
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
			if k == hookin.ProjectEnv {
				panic("boom")
			}
			return ""
		}, editOf("s-3", dir, "free.txt"), RuleInternal, "the guard failed inside (boom)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, stderr := guardRun(t, c.getenv, strings.NewReader(c.stdin), untimed)
			if code != 2 || !strings.Contains(stderr, c.inErr) || !strings.Contains(stderr, "\nnext: ") {
				t.Fatalf("exit %d, stderr %q; want 2 with %q and a next step", code, stderr, c.inErr)
			}
			asciiOnly(t, stderr)
		})
	}
	// A bad input is recorded, in the day's file (no session known).
	recs := dayRecords(t, dir)
	if len(recs) != 3 || recs[0].String("rule") != RuleBadInput || recs[0].String("decision") != "deny" {
		t.Fatalf("records %d: %v", len(recs), recs)
	}
	if v, _ := recs[0].Get("session"); v != nil {
		t.Errorf("session %v, want null", v)
	}
	// A stdin that is a terminal (the guard run by hand) is not waited on.
	var stderr bytes.Buffer
	code := Main(Options{Stdin: strings.NewReader(""), StdinTerminal: true, Stderr: &stderr, Getenv: env(dir), Budget: untimed})
	if code != 2 || !strings.Contains(stderr.String(), "is a terminal") {
		t.Errorf("terminal: exit %d, %q", code, stderr.String())
	}
}

func TestBadConfigBlocksWithNoRecord(t *testing.T) {
	dir := project(t, "format: bonsai.workspace/1\nid: nonsense\n")
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-4", dir, "free.txt")), untimed)
	if code != 2 || !strings.Contains(stderr, "bonsai.yaml cannot be read") || !strings.Contains(stderr, "\nnext: tell the person: ") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, ".bonsai")); !os.IsNotExist(err) {
		t.Errorf("a record without a workspace id was attempted: %v", err)
	}
}

// A stdin that never ends: the work is still reading when the budget runs out, and the guard blocks by itself
// (over-time), its record's writer awaited for the run's RecordWait (the frame's option). Two cases:
//   - the record half: a wait far past any load, so the record is in the log when Main returns, read with no polling;
//   - the held writer: the writer held by the test (recordOverTime), so the answer comes back within the wait with
//     the record unwritten; released, the record lands.
//
// The over-time record needs bonsai.yaml read, so a run whose timer fires before the work has read it hands the frame
// no writer and writes nothing, rightly: that run is tried again on a new project with twice the budget, up to the
// guard's real Budget (a guard that cannot read bonsai.yaml in that time fails a real run too). Each answer's time is
// held only to a ceiling a hang alone crosses: the budget, the wait and the real Budget.
func TestOverTimeBlocks(t *testing.T) {
	defer func(f func(func()) func()) { recordOverTime = f }(recordOverTime)
	t.Run("the record half", func(t *testing.T) {
		const wait = time.Minute
		overTimeRun(t, wait, func(write func()) func() { return write }, func(dir string) {
			recs := dayRecords(t, dir)
			if len(recs) != 1 || recs[0].String("rule") != RuleOverTime || recs[0].String("decision") != "deny" {
				t.Fatalf("records when Main returned: %v", recs)
			}
		})
	})
	t.Run("the held writer", func(t *testing.T) {
		const wait = 100 * time.Millisecond
		release, finished := make(chan struct{}), make(chan struct{})
		hold := func(write func()) func() {
			return func() {
				<-release
				defer close(finished)
				write()
			}
		}
		overTimeRun(t, wait, hold, func(dir string) {
			if names, _ := filepath.Glob(filepath.Join(dir, filepath.FromSlash(LogDir), "*.ndjson")); len(names) != 0 {
				t.Fatalf("a record was written while its writer was held: %v", names)
			}
			close(release)
			select {
			case <-finished:
			case <-time.After(time.Minute):
				t.Fatal("the released writer did not finish within a minute")
			}
			if recs := dayRecords(t, dir); len(recs) != 1 || recs[0].String("rule") != RuleOverTime {
				t.Fatalf("records once released: %v", recs)
			}
		})
	})
}

// overTimeRun runs the guard on a stdin that never ends with the given RecordWait and stand-in for recordOverTime, from
// a 150 ms budget, doubling it while the timer fires before the work has read bonsai.yaml (no writer handed over);
// then checks the answer, and hands the project to after.
func overTimeRun(t *testing.T, wait time.Duration, stand func(func()) func(), after func(dir string)) {
	t.Helper()
	for budget := 150 * time.Millisecond; ; budget *= 2 {
		dir := project(t, testYAML)
		handed := false
		recordOverTime = func(write func()) func() { handed = true; return stand(write) }
		// The pipe's writer is never closed: the work stays blocked in its read, as a stdin that never ends would, and
		// writes nothing more into the test's folder.
		pr, _ := io.Pipe()
		var stderr bytes.Buffer
		start := time.Now()
		code := Main(Options{Stdin: pr, Stderr: &stderr, Getenv: env(dir), Budget: budget, RecordWait: wait})
		took := time.Since(start)
		if code != 2 || !strings.Contains(stderr.String(), fmt.Sprintf("the guard ran past its own %s limit", budget)) ||
			!strings.Contains(stderr.String(), "\nnext: ") {
			t.Fatalf("budget %v: exit %d, stderr %q", budget, code, stderr.String())
		}
		if ceiling := budget + wait + Budget; took > ceiling {
			t.Fatalf("budget %v, wait %v: answered after %v, past %v", budget, wait, took, ceiling)
		}
		if handed {
			after(dir)
			return
		}
		if budget > Budget {
			t.Fatalf("budget %v, past the guard's real %v: the timer still fired before the work had read bonsai.yaml", budget, Budget)
		}
		t.Logf("budget %v: the timer fired before the work had read bonsai.yaml; again with %v", budget, 2*budget)
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
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("s-5", dir, "free.txt")), untimed)
	if code != 2 || !strings.Contains(stderr, "the guard would allow this call but cannot write its record") {
		t.Fatalf("free.txt: exit %d, stderr %q", code, stderr)
	}
	code, stderr = guardRun(t, env(dir), strings.NewReader(editOf("s-5", dir, "protected.txt")), untimed)
	if code != 2 || !strings.Contains(stderr, "its record in .bonsai/local/log could not be written either") {
		t.Fatalf("protected.txt: exit %d, stderr %q", code, stderr)
	}
}

func TestLongValuesStayUnderTheRecordCap(t *testing.T) {
	dir := project(t, testYAML)
	long := strings.Repeat("d/", 600) + "x.txt"
	in := strings.Replace(editOf("s-6", dir, long), "toolu_01TEST", "toolu_"+strings.Repeat("Z", 1500), 1)
	code, stderr := guardRun(t, env(dir, "BONSAI_TASK", strings.Repeat("T", 900), "BONSAI_ROLE", "builder"), strings.NewReader(in), untimed)
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
			codes[i], _ = guardRun(t, env(dir), strings.NewReader(editOf("s-7", dir, rel)), untimed)
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

// The guard's own limit is half the timeout part 3's hook line gives it.
func TestBudgetIsHalfTheHookTimeout(t *testing.T) {
	secs, err := strconv.Atoi(engine.GuardTimeout)
	if err != nil {
		t.Fatalf("engine.GuardTimeout %q: %v", engine.GuardTimeout, err)
	}
	if 2*Budget != time.Duration(secs)*time.Second {
		t.Fatalf("Budget %v is not half of the hook line's timeout, %d s", Budget, secs)
	}
}

// The record's own small rules: categories, and the binary's path from ~/ under the user's home (letter case
// ignored on Windows), else as it is.
func TestRecordFields(t *testing.T) {
	if got := category("NotebookEdit") + category("Bash") + category("Glob"); got != "EditShellOther" {
		t.Errorf("category: %q", got)
	}
	defer func(f func() (string, error)) { userHome = f }(userHome)
	home := filepath.Join(t.TempDir(), "Me")
	userHome = func() (string, error) { return home, nil }
	cases := map[string]string{
		filepath.Join(home, "bin", "bonsai"):                    "~/bin/bonsai",
		filepath.Join(filepath.Dir(home), "Mentor", "bin", "x"): filepath.ToSlash(filepath.Join(filepath.Dir(home), "Mentor", "bin", "x")),
		home: filepath.ToSlash(home),
		filepath.Join(filepath.Dir(home), "other", "bonsai"): filepath.ToSlash(filepath.Join(filepath.Dir(home), "other", "bonsai")),
		"": "",
	}
	if runtime.GOOS == "windows" {
		cases[filepath.Join(strings.ToUpper(home), "bonsai.exe")] = "~/bonsai.exe"
	}
	for in, want := range cases {
		if got := fromHome(in); got != want {
			t.Errorf("fromHome(%q) = %q, want %q", in, got, want)
		}
	}
	userHome = func() (string, error) { return "", errors.New("no home") }
	if got := fromHome(filepath.Join(home, "bonsai")); got != filepath.ToSlash(filepath.Join(home, "bonsai")) {
		t.Errorf("with no home: %q", got)
	}
}
