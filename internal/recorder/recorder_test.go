package recorder

// Payloads written fresh for every event the recorder writes (Claude Code's hooks reference, its fields only, with
// made-up values), each record read back and held to bonsai.log/1, field by field.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const testID = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"

const testYAML = "format: bonsai.workspace/1\nid: " + testID + "\nname: demo\ndocuments:\n  task: work/tasks\n"

// fakeToken is a made-up GitHub-shaped token, built at run time so no token-shaped literal sits in the repo.
func fakeToken() string { return "ghp_" + strings.Repeat("Mn4Rt8", 6) }

// project makes a linked checkout (bonsai.yaml written, git on branch main) in an isolated home.
func project(t *testing.T) (tmp, root string) {
	t.Helper()
	tmp = t.TempDir()
	testpack.Isolate(t, tmp)
	root = testpack.Project(t, tmp, "proj")
	write(t, root, "bonsai.yaml", testYAML)
	return tmp, root
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// env is a test's environment: CLAUDE_PROJECT_DIR and what else it names; nothing else is read.
func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

// opts gives Options for a payload, the project in CLAUDE_PROJECT_DIR.
func opts(root, payload string, extra map[string]string) Options {
	vars := map[string]string{ProjectEnv: root}
	for k, v := range extra {
		vars[k] = v
	}
	return Options{Stdin: strings.NewReader(payload), Getenv: env(vars),
		Executable: func() (string, error) { return "", errors.New("no binary in this test") }}
}

// records reads a session's log file in the main checkout, every line held to bonsai.log/1: the schema (every field,
// in order, each of its type and pattern) and the 2,048-byte cap.
func records(t *testing.T, main, session string) []*format.Log {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(main, ".bonsai", "local", "log", "s-"+session+".ndjson"))
	if err != nil {
		t.Fatal(err)
	}
	f := format.MustLookup("log")
	var out []*format.Log
	for _, line := range bytes.SplitAfter(raw, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		if len(line) > f.MaxLine() || !bytes.HasSuffix(line, []byte("\n")) {
			t.Fatalf("a line of %d bytes, or not whole: %q", len(line), line)
		}
		v, err := schema.Decode(bytes.TrimSuffix(line, []byte("\n")))
		if err != nil {
			t.Fatalf("%v: %s", err, line)
		}
		if errs := append(schema.Validate(f.Schema(), v), schema.CheckOrder(f.Schema(), v)...); len(errs) > 0 {
			t.Fatalf("not bonsai.log/1: %v\n%s", errs, line)
		}
		l, err := format.ReadLog(line)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, l)
	}
	return out
}

func boolText(b *bool) string {
	if b == nil {
		return "<null>"
	}
	if *b {
		return "true"
	}
	return "false"
}

func s(p *string) string {
	if p == nil {
		return "<null>"
	}
	return *p
}

// Every event the recorder writes, one payload each, written fresh: the record's event and the fields it fills.
func TestEveryEvent(t *testing.T) {
	_, root := project(t)
	write(t, root, "work/tasks/T-0901-x.md", "---\nformat: bonsai.task/1\nid: T-0901\ntitle: Make it\nstatus: running\nlane: null\n"+
		"done_when: []\ndepends_on: []\nblocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels: {}\n---\n")
	sess := "0f1e2d3c-aaaa-4bbb-8ccc-000000000001"
	base := `"session_id":"` + sess + `","transcript_path":"/tmp/t.jsonl","cwd":"` + filepath.ToSlash(root) + `","permission_mode":"default"`
	bash := `"tool_name":"Bash","tool_input":{"command":"DEPLOY_TOKEN=` + fakeToken() + ` git push origin main","description":"push"}`
	cases := []struct {
		payload string
		event   string
		check   func(l *format.Log) string
	}{
		{`{` + base + `,"hook_event_name":"UserPromptSubmit","prompt":"please use ` + fakeToken() + `"}`, "prompt",
			func(l *format.Log) string { return s(l.Kind) + " " + s(l.Text) }},
		{`{` + base + `,"hook_event_name":"UserPromptSubmit","prompt":"  <task-notification>done</task-notification>"}`, "prompt",
			func(l *format.Log) string { return s(l.Kind) + " " + s(l.Text) }},
		{`{` + base + `,"hook_event_name":"PreToolUse",` + bash + `,"tool_use_id":"toolu_01AAA"}`, "tool_start",
			func(l *format.Log) string {
				return s(l.Tool) + " " + s(l.Category) + " " + s(l.Target) + " " + s(l.ToolUseID) + " " + s(l.Kind)
			}},
		{`{` + base + `,"hook_event_name":"PermissionRequest",` + bash + `}`, "permission",
			func(l *format.Log) string {
				return s(l.Tool) + " " + s(l.Category) + " " + s(l.Target) + " " + s(l.ToolUseID)
			}},
		{`{` + base + `,"hook_event_name":"PostToolUse",` + bash + `,"tool_use_id":"toolu_01AAA","tool_response":{"stdout":"` +
			fakeToken() + `"},"duration_ms":12}`, "tool_end",
			func(l *format.Log) string { return s(l.Target) + " ok=" + boolText(l.OK) }},
		{`{` + base + `,"hook_event_name":"PostToolUseFailure",` + bash + `,"tool_use_id":"toolu_01AAA","error":"Exit code 1","is_interrupt":true}`,
			"tool_fail", func(l *format.Log) string { return s(l.Kind) + " ok=" + boolText(l.OK) }},
		{`{` + base + `,"hook_event_name":"PostToolUseFailure",` + bash + `,"tool_use_id":"toolu_01AAA","error":"Exit code 1"}`,
			"tool_fail", func(l *format.Log) string { return s(l.Kind) }},
		{`{` + base + `,"hook_event_name":"Notification","message":"Claude needs your permission, key ` + fakeToken() +
			`","notification_type":"permission_prompt","title":"x"}`, "notice",
			func(l *format.Log) string { return s(l.Kind) + " " + s(l.Text) }},
		{`{` + base + `,"hook_event_name":"SubagentStart","agent_id":"a1b2c3","agent_type":"workflow:builder"}`, "subagent_start",
			func(l *format.Log) string { return s(l.SubagentID) + " " + s(l.SubagentType) + " " + s(l.Target) }},
		{`{` + base + `,"hook_event_name":"SubagentStop","agent_id":"a1b2c3","agent_type":"workflow:builder","stop_hook_active":false,` +
			`"agent_transcript_path":"/tmp/a.jsonl","last_assistant_message":"done"}`, "subagent_stop",
			func(l *format.Log) string { return s(l.SubagentID) + " " + s(l.Target) }},
		{`{` + base + `,"hook_event_name":"Stop","stop_hook_active":false,"last_assistant_message":"all done"}`, "stop",
			func(l *format.Log) string { return s(l.Text) + " " + s(l.Kind) }},
		{`{` + base + `,"hook_event_name":"SessionEnd","reason":"prompt_input_exit"}`, "session_end",
			func(l *format.Log) string { return s(l.Reason) }},
	}
	want := []string{
		"user <null>",
		"task-notification <null>",
		"Bash Shell git push toolu_01AAA <null>",
		"Bash Shell git push <null>",
		"git push ok=true",
		"interrupt ok=false",
		"error",
		"permission_prompt Claude needs your permission, key [redacted]",
		"a1b2c3 workflow:builder T-0901",
		"a1b2c3 <null>",
		"<null> <null>",
		"prompt_input_exit",
	}
	for _, c := range cases {
		if code := Record(opts(root, c.payload, map[string]string{"BONSAI_TASK": "T-0901", "BONSAI_ROLE": "builder"})); code != 0 {
			t.Fatalf("exit %d", code)
		}
	}
	got := records(t, root, sess)
	if len(got) != len(cases) {
		t.Fatalf("%d records for %d payloads", len(got), len(cases))
	}
	raw, _ := os.ReadFile(filepath.Join(root, ".bonsai", "local", "log", "s-"+sess+".ndjson"))
	if bytes.Contains(raw, []byte(fakeToken()[4:])) || bytes.Contains(raw, []byte("please use")) || bytes.Contains(raw, []byte("all done")) {
		t.Errorf("the log holds a secret or a prompt's words:\n%s", raw)
	}
	hashes := map[string]bool{}
	for i, l := range got {
		if l.Event != cases[i].event || s(l.AgentEvent) == "<null>" || s(l.Agent) != "claude-code" || s(l.Session) != sess ||
			l.Workspace != testID || s(l.Checkout) != "proj" || s(l.Branch) != "main" || s(l.Task) != "T-0901" || s(l.Role) != "builder" ||
			l.Remote != nil || l.BonsaiPath != nil || l.BonsaiSHA256 != nil || len(l.Labels) != 0 {
			t.Errorf("%s: the common fields: %+v", cases[i].event, l)
		}
		if g := cases[i].check(l); g != want[i] {
			t.Errorf("%s: %q, want %q", cases[i].event, g, want[i])
		}
		if l.InputHash != nil {
			hashes[*l.InputHash] = true
		}
		if toolEvent(s(l.AgentEvent)) != (l.InputHash != nil) {
			t.Errorf("%s: input_hash %s", cases[i].event, s(l.InputHash))
		}
	}
	// The four tool events of one call share one input hash.
	if len(hashes) != 1 {
		t.Errorf("the tool events' hashes: %v", hashes)
	}
}

// hook start: the session_start record (source, model as a string or an object's id, the active task, the binary's
// path from ~/ and its SHA-256, the file made by this record), then the opening context, ASCII only.
func TestStart(t *testing.T) {
	tmp, root := project(t)
	write(t, root, "work/tasks/T-0902-x.md", "---\nformat: bonsai.task/1\nid: T-0902\ntitle: Ship the caf\u00e9\nstatus: running\n"+
		"lane: full\ndone_when: []\ndepends_on: []\nblocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels:\n"+
		"  bonsai.branch: t0902-ship\n  bonsai.allows:\n    - src/a.go\n    - docs/b.md\n---\n")
	userHome := filepath.Join(tmp, "user")
	exe := filepath.Join(userHome, "bin", "bonsai")
	write(t, userHome, "bin/bonsai", "a made-up binary")
	sum := sha256.Sum256([]byte("a made-up binary"))
	sess := "0f1e2d3c-aaaa-4bbb-8ccc-000000000002"
	o := opts(root, `{"session_id":"`+sess+`","hook_event_name":"SessionStart","source":"compact","model":{"id":"claude-made-up"},"cwd":"`+
		filepath.ToSlash(root)+`"}`, map[string]string{RemoteEnv: "session_0123456789abcdefXYZ"})
	o.Executable = func() (string, error) { return exe, nil }
	o.UserHome = func() (string, error) { return userHome, nil }
	var out bytes.Buffer
	if code := Start(o, &out); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := records(t, root, sess)
	if len(got) != 1 {
		t.Fatalf("%d records", len(got))
	}
	l := got[0]
	if l.Event != "session_start" || s(l.Source) != "compact" || s(l.Model) != "claude-made-up" || s(l.Target) != "T-0902" ||
		s(l.BonsaiPath) != "~/bin/bonsai" || s(l.BonsaiSHA256) != hex.EncodeToString(sum[:]) || s(l.Remote) != "session_0123456789abcdefXYZ" {
		t.Errorf("the record: %+v", l)
	}
	text := out.String()
	for _, w := range []string{
		"Bonsai: this project is the workspace demo (" + testID + "). Its log is in .bonsai/local/ (never committed).",
		`Active task: T-0902 "Ship the caf\u00e9" (work/tasks/T-0902-x.md), status running, lane full, branch t0902-ship.`,
		"Its grants (bonsai.allows): src/a.go, docs/b.md.",
		"Its last ladder result: none yet.",
	} {
		if !strings.Contains(text, w+"\n") {
			t.Errorf("the context lacks %q:\n%s", w, text)
		}
	}
	for _, r := range text {
		if r > 0x7e || (r < 0x20 && r != '\n') {
			t.Fatalf("a character that is not plain ASCII: %q", r)
		}
	}
	// A second start (a resume) hashes again and appends its own record; the file was made once.
	o.Stdin = strings.NewReader(`{"session_id":"` + sess + `","hook_event_name":"SessionStart","source":"resume","model":"m"}`)
	Start(o, &bytes.Buffer{})
	if got := records(t, root, sess); len(got) != 2 || s(got[1].BonsaiSHA256) != hex.EncodeToString(sum[:]) || s(got[1].Source) != "resume" {
		t.Errorf("the resumed start: %+v", got)
	}
	// A ladder result for the task, and a machine's label definitions, are named.
	write(t, root, ".bonsai/local/ladder/T-0902.json", ladderJSON(t))
	machine, err := workspace.MachineDir(filepath.Join(tmp, "home"), root)
	if err != nil {
		t.Fatal(err)
	}
	write(t, machine, "labels/tracker.yaml", "format: bonsai.labels/1\nnamespace: tracker\nversion: 3\nlabels:\n  - name: tracker.cost\n"+
		"    kind: number\n    values: []\n    items: null\n    pattern: null\n    max: null\n    kinds: [task]\n    set_by: outside\n"+
		"    grants: false\n    description: What the task cost.\n")
	o.Stdin = strings.NewReader(`{"session_id":"` + sess + `","hook_event_name":"SessionStart","source":"clear"}`)
	out.Reset()
	Start(o, &out)
	for _, w := range []string{"Its last ladder result: green at commit 0123456789ab (local), finished 2026-10-10T10:00:01.000Z.",
		"Label definitions attached on this machine (contract section 5.3):", "  tracker.cost (number; set by outside; version 3): What the task cost."} {
		if !strings.Contains(out.String(), w+"\n") {
			t.Errorf("the context lacks %q:\n%s", w, out.String())
		}
	}
}

// ladderJSON is a made-up green ladder result for T-0902, from the format's own example with its task and commit set.
func ladderJSON(t *testing.T) string {
	t.Helper()
	l := &format.Ladder{Task: ptr("T-0902"), Workspace: testID, Mode: "local", Started: "2026-10-10T10:00:00.000Z",
		Finished: "2026-10-10T10:00:01.000Z", Git: format.LadderGit{SHA: "0123456789abcdef0123456789abcdef01234567", Branch: "main"},
		Requested: []int64{0}, Green: true, Rungs: []format.LadderRung{}, Skipped: []format.LadderSkip{}}
	raw, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// No active task: the context says none, with the function's reason; an unreadable bonsai.yaml gives one line, and
// writes nothing; past 60 lines, one line names bonsai status --json.
func TestStartContextCases(t *testing.T) {
	tmp, root := project(t)
	var out bytes.Buffer
	Start(opts(root, `{"session_id":"s1","hook_event_name":"SessionStart","source":"startup"}`, nil), &out)
	if !strings.Contains(out.String(), "Active task: none: no task reads running.\n") || strings.Count(out.String(), "\n") != 2 {
		t.Errorf("no task:\n%s", out.String())
	}
	machine, _ := workspace.MachineDir(filepath.Join(tmp, "home"), root)
	var labels strings.Builder
	labels.WriteString("format: bonsai.labels/1\nnamespace: many\nversion: 1\nlabels:\n")
	for i := 0; i < 70; i++ {
		labels.WriteString("  - name: many.l" + strings.Repeat("x", i%5) + string(rune('a'+i%26)) + string(rune('a'+i/26)) +
			"\n    kind: text\n    values: []\n    items: null\n    pattern: null\n    max: null\n    kinds: [task]\n    set_by: agent\n" +
			"    grants: false\n    description: A label.\n")
	}
	write(t, machine, "labels/many.yaml", labels.String())
	out.Reset()
	Start(opts(root, `{"session_id":"s1","hook_event_name":"SessionStart","source":"startup"}`, nil), &out)
	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if len(lines) != MaxContextLines || lines[MaxContextLines-1] != "More than fits here: run bonsai status --json." {
		t.Errorf("%d lines, the last %q", len(lines), lines[len(lines)-1])
	}
	write(t, root, "bonsai.yaml", "format: bonsai.workspace/1\nid: [\n")
	out.Reset()
	if code := Start(opts(root, `{"session_id":"s9","hook_event_name":"SessionStart"}`, nil), &out); code != 0 ||
		!strings.HasPrefix(out.String(), "Bonsai: bonsai.yaml cannot be read (") || strings.Count(out.String(), "\n") != 1 ||
		!strings.Contains(out.String(), "guard blocks every file edit and shell command until a person fixes it") {
		t.Errorf("an unreadable bonsai.yaml: %d\n%s", code, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".bonsai", "local", "log", "s-s9.ndjson")); err == nil {
		t.Error("a record under an unreadable bonsai.yaml")
	}
	// A folder with no bonsai.yaml: nothing printed, nothing written.
	plain := testpack.Project(t, tmp, "plain")
	out.Reset()
	if code := Start(opts(plain, `{"session_id":"s1","hook_event_name":"SessionStart"}`, nil), &out); code != 0 || out.Len() != 0 {
		t.Errorf("an unlinked folder: %d %q", code, out.String())
	}
	if _, err := os.Stat(filepath.Join(plain, ".bonsai")); err == nil {
		t.Error("an unlinked folder got .bonsai/")
	}
}

// A worktree's session writes the main checkout's log (contract §3), its checkout and branch its own, and its active
// task read from main; a project found from the payload's cwd when CLAUDE_PROJECT_DIR is unset, never above a
// checkout's top.
func TestWorktreeWritesMain(t *testing.T) {
	tmp, root := project(t)
	write(t, root, "work/tasks/T-0903-x.md", "---\nformat: bonsai.task/1\nid: T-0903\ntitle: x\nstatus: running\nlane: null\n"+
		"done_when: []\ndepends_on: []\nblocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels: {}\n---\n")
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "x")
	wt := filepath.Join(tmp, "wt")
	testpack.Git(t, root, "worktree", "add", "-q", "-b", "t0903", wt)
	if err := os.RemoveAll(filepath.Join(wt, "work")); err != nil { // tasks are read from main, never the worktree's copy
		t.Fatal(err)
	}
	o := opts(wt, `{"session_id":"wt-1","hook_event_name":"SubagentStart","agent_id":"b2","agent_type":"Explore"}`, nil)
	Record(o)
	o = Options{Stdin: strings.NewReader(`{"session_id":"wt-1","hook_event_name":"Stop","cwd":"` + filepath.ToSlash(filepath.Join(wt, "sub")) + `"}`),
		Getenv: env(nil)}
	if err := os.MkdirAll(filepath.Join(wt, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	Record(o)
	got := records(t, root, "wt-1")
	if len(got) != 2 || s(got[0].Checkout) != "wt" || s(got[0].Branch) != "t0903" || s(got[0].Target) != "T-0903" || got[1].Event != "stop" {
		t.Errorf("the worktree's records: %+v", got)
	}
	if _, err := os.Stat(filepath.Join(wt, ".bonsai", "local")); err == nil {
		t.Error("the worktree has a local/ of its own")
	}
}

// Tool calls reduced and redacted: a path outside the checkout is null, a path inside relative; categories as
// contract §8.2 has them, Ladder by the command's head; AskUserQuestion keeps its questions as text and no target;
// ids held to their pattern; an input the strict reader refuses has no hash and no target, and the record is
// still written; no salt, no hash.
func TestToolCalls(t *testing.T) {
	_, root := project(t)
	type c struct{ tool, input, cat, target, text string }
	cases := []c{
		{"Read", `{"file_path":"` + filepath.ToSlash(filepath.Join(root, "src", "a.go")) + `"}`, "Read", "src/a.go", "<null>"},
		{"Read", `{"file_path":"/etc/passwd"}`, "Read", "<null>", "<null>"},
		{"Grep", `{"pattern":"x","path":"src"}`, "Search", "src", "<null>"},
		{"Glob", `{"pattern":"*.go"}`, "Search", "<null>", "<null>"},
		{"MultiEdit", `{"file_path":"b.go","edits":[]}`, "Edit", "b.go", "<null>"},
		{"NotebookEdit", `{"notebook_path":"n.ipynb"}`, "Edit", "n.ipynb", "<null>"},
		{"Write", `{"file_path":"c.go","content":"password: ` + fakeToken() + `"}`, "Write", "c.go", "<null>"},
		{"Bash", `{"command":"bonsai ladder --task T-0901"}`, "Ladder", "bonsai ladder", "<null>"},
		{"PowerShell", `{"command":"Get-ChildItem -Path C:\\x"}`, "Shell", "Get-ChildItem", "<null>"},
		{"Agent", `{"subagent_type":"Explore","prompt":"look"}`, "Agent", "Explore", "<null>"},
		{"Task", `{"subagent_type":"Plan","prompt":"look"}`, "Agent", "Plan", "<null>"},
		{"WebFetch", `{"url":"https://user:pw@example.invalid/a?b=c"}`, "Web", "example.invalid", "<null>"},
		{"WebSearch", `{"query":"what is ` + fakeToken() + `"}`, "Web", "<null>", "<null>"},
		{"mcp__tracker__list_items", `{"q":1}`, "MCP", "tracker.list_items", "<null>"},
		{"AskUserQuestion", `{"questions":[{"question":"Which way?","options":[]},{"question":"Key ` + fakeToken() + `?"}]}`,
			"Other", "<null>", "Which way? \u00b7 Key [redacted]?"},
		{"Skill", `{"skill":"memory"}`, "Other", "memory", "<null>"},
		{"Read", `{"file_path":"a","file_path":"b"}`, "Read", "<null>", "<null>"},
	}
	for i, k := range cases {
		p := `{"session_id":"tc","hook_event_name":"PreToolUse","tool_name":"` + k.tool + `","tool_input":` + k.input +
			`,"tool_use_id":"toolu_` + string(rune('A'+i)) + `","cwd":"` + filepath.ToSlash(root) + `"}`
		Record(opts(root, p, nil))
	}
	got := records(t, root, "tc")
	if len(got) != len(cases) {
		t.Fatalf("%d records", len(got))
	}
	for i, k := range cases {
		l := got[i]
		if s(l.Category) != k.cat || s(l.Target) != k.target || s(l.Text) != k.text {
			t.Errorf("%s %s: category %s, target %s, text %s", k.tool, k.input, s(l.Category), s(l.Target), s(l.Text))
		}
	}
	if got[len(got)-1].InputHash != nil || got[0].InputHash == nil {
		t.Error("a refused input has a hash, or a read one none")
	}
	// Ids held to their pattern; the subagent's type redacted; a remote id only by its pattern.
	Record(opts(root, `{"session_id":"tc","hook_event_name":"PostToolUse","tool_name":"Read","tool_input":{},"tool_use_id":"toolu/../x",`+
		`"agent_id":"a b","agent_type":"x password=`+fakeToken()+`"}`, map[string]string{RemoteEnv: "session_short"}))
	l := records(t, root, "tc")[len(cases)]
	if l.ToolUseID != nil || l.SubagentID != nil || strings.Contains(s(l.SubagentType), fakeToken()[4:]) || l.Remote != nil {
		t.Errorf("ids: %+v", l)
	}
	// No salt: the hash is null and the record is written.
	saved := salt
	salt = func() (string, error) { return "", errors.New("no home") }
	defer func() { salt = saved }()
	Record(opts(root, `{"session_id":"tc","hook_event_name":"PostToolUse","tool_name":"Read","tool_input":{"file_path":"a"}}`, nil))
	if l := records(t, root, "tc"); len(l) != len(cases)+2 || l[len(l)-1].InputHash != nil {
		t.Errorf("no salt: %+v", l[len(l)-1])
	}
}

// The input hash: keyed by the salt, over the input with keys sorted at every depth and numbers as written, so key
// order does not count and a number's text does; 16 hex.
func TestInputHash(t *testing.T) {
	dec := func(raw string) any {
		v, err := schema.Decode([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	a := InputHash("k1", dec(`{"b":{"y":1,"x":[{"q":2,"p":1}]},"a":"\u00e9"}`))
	b := InputHash("k1", dec(`{"a":"\u00e9","b":{"x":[{"p":1,"q":2}],"y":1}}`))
	if a != b || !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(a) {
		t.Errorf("key order counted: %s %s", a, b)
	}
	if InputHash("k2", dec(`{"a":1}`)) == InputHash("k1", dec(`{"a":1}`)) || InputHash("k1", dec(`{"a":1.0}`)) == InputHash("k1", dec(`{"a":1}`)) ||
		InputHash("k1", dec(`{"a":[1,2]}`)) == InputHash("k1", dec(`{"a":[2,1]}`)) {
		t.Error("the salt, a number's text or a list's order did not count")
	}
}

// Within its budget or not, the hook returns 0: a stdin that never ends stops at the budget.
func TestBudget(t *testing.T) {
	_, root := project(t)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = w.Close(); _ = r.Close() }()
	o := opts(root, "", nil)
	o.Stdin, o.Budget = r, 200*time.Millisecond
	start := time.Now()
	if code := Record(o); code != 0 || time.Since(start) > 2*time.Second {
		t.Errorf("exit %d after %s", code, time.Since(start))
	}
	var out bytes.Buffer
	o.Stdin = r
	if code := Start(o, &out); code != 0 || out.Len() != 0 {
		t.Errorf("start past its budget: %d %q", code, out.String())
	}
}

// The binary's path from ~/ under the home, case-blind only on Windows; elsewhere as it is.
func TestUnderHome(t *testing.T) {
	if got := underHome("/home/someone/bin/bonsai", "/home/someone"); got != "~/bin/bonsai" {
		t.Errorf("%s", got)
	}
	if got := underHome("/usr/local/bin/bonsai", "/home/someone"); got != "/usr/local/bin/bonsai" {
		t.Errorf("%s", got)
	}
	if got := underHome("/home/someone-else/bonsai", "/home/someone"); got != "/home/someone-else/bonsai" {
		t.Errorf("%s", got)
	}
}

// A record's line keeps its 2,048 bytes when a payload's strings are long: the text cut to its cap, then halved.
func TestLongPayload(t *testing.T) {
	_, root := project(t)
	long := strings.Repeat("word ", 4000)
	Record(opts(root, `{"session_id":"lp","hook_event_name":"Notification","message":"`+long+`","notification_type":"`+long+`"}`,
		map[string]string{"BONSAI_TASK": long, "BONSAI_ROLE": long}))
	if l := records(t, root, "lp"); len(l) != 1 || len([]rune(s(l[0].Text))) > 300 {
		t.Errorf("%+v", l)
	}
}
