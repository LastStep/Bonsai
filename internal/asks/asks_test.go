package asks

// The asks (design/plan-5.md, 5.2.5): a table of every filing case (each type, every limit, every refusal), the
// states of a key and the answers to it, every free-text field redacted, the day files, the reader, and the writes
// that fail. Every secret here is made up for the test.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const (
	testWS = "ws-abcdefghijklmnopqrstuvwxyz"
	sessA  = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"
	sessB  = "7a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
)

// testPlace is a fresh main checkout (no git: the place is given whole) with a clock that moves one second a call
// from 2026-10-08 12:00 UTC.
func testPlace(t *testing.T, session string) Place {
	t.Helper()
	dir := t.TempDir()
	tick := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	return Place{Workspace: testWS, Local: workspace.Local{Root: dir, Main: dir, Branch: "main"}, Session: session,
		Now: func() time.Time { tick = tick.Add(time.Second); return tick }}
}

// testKinds are Bonsai's own document kinds and one a pack declares (N-<3 digits>).
func testKinds(t *testing.T) []workspace.DocKind {
	t.Helper()
	k, err := workspace.DocKinds(&format.Workspace{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return append(k, workspace.DocKind{Kind: "note", From: "notes-pack", Path: "notes", ID: `^N-[0-9]{3}$`})
}

// files are every file under dir, by slash path, with their bytes.
func files(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := os.ReadFile(p)
		rel, _ := filepath.Rel(dir, p)
		out[filepath.ToSlash(rel)] = string(raw)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// askRecords are the ask records under dir's asks folder, files by name, lines in order.
func askRecords(t *testing.T, dir string) []*format.Ask {
	t.Helper()
	var out []*format.Ask
	entries, _ := os.ReadDir(Folder(dir))
	for _, e := range entries {
		raw, err := os.ReadFile(filepath.Join(Folder(dir), e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
			a, err := format.ReadAsk([]byte(line))
			if err != nil {
				t.Fatalf("%s: %v\n%s", e.Name(), err, line)
			}
			out = append(out, a)
		}
	}
	return out
}

// logRecords are the log records under dir's log folder, by file name.
func logRecords(t *testing.T, dir string) map[string][]*format.Log {
	t.Helper()
	out := map[string][]*format.Log{}
	logDir := filepath.Join(dir, ".bonsai", "local", "log")
	entries, _ := os.ReadDir(logDir)
	for _, e := range entries {
		lf, err := record.ReadLog(filepath.Join(logDir, e.Name()))
		if err != nil || lf.Skipped > 0 {
			t.Fatalf("%s: %v, %d skipped", e.Name(), err, lf.Skipped)
		}
		out[e.Name()] = lf.Records
	}
	return out
}

var okFiling = Filing{Type: "Answer", Title: "Which colour should links use?", Why: "Two blues pass the contrast check."}

func with(mod func(*Filing)) Filing {
	f := okFiling
	mod(&f)
	return f
}

func hashKey(typ, target, title string) string {
	sum := sha256.Sum256([]byte(typ + "\x00" + target + "\x00" + title))
	return "agent:h-" + hex.EncodeToString(sum[:])[:12]
}

// Every filing case: each type, every limit, every rule refused (exit 2, the rule named, nothing written).
func TestFile(t *testing.T) {
	long := func(n int) string { return strings.Repeat("a", n) }
	options := func(o ...string) func(*Filing) {
		return func(f *Filing) { f.Type, f.Options = "Decide", o }
	}
	emoji := strings.Repeat("\U0001F600", 300)
	cases := []struct {
		name string
		f    Filing
		code string // "" when it files
		in   string // in the refusal's sentence, or the key filed
	}{
		// Each type.
		{"an Answer", okFiling, "", hashKey("Answer", "answers", okFiling.Title)},
		{"an Answer about a task", with(func(f *Filing) { f.Task = "T-0901" }), "", hashKey("Answer", "T-0901", okFiling.Title)},
		{"an Answer about Bonsai's kind of document", with(func(f *Filing) { f.Doc = "M-colours" }), "", hashKey("Answer", "M-colours", okFiling.Title)},
		{"an Answer about a pack's kind of document", with(func(f *Filing) { f.Doc = "N-007" }), "", hashKey("Answer", "N-007", okFiling.Title)},
		{"a Decide with four options", with(options("one", "two", "three", "four")), "", hashKey("Decide", "answers", okFiling.Title)},
		{"a Decide with no options", with(func(f *Filing) { f.Type = "Decide" }), "", ""},
		{"a Look with its verdict", with(func(f *Filing) { f.Type, f.Task, f.Verdict = "Look", "T-0901", "pass" }), "", ""},
		{"a Look with no verdict", with(func(f *Filing) { f.Type, f.Task = "Look", "T-0901" }), "", ""},
		{"a Play", with(func(f *Filing) { f.Type, f.Task = "Play", "T-123456" }), "", ""},
		// The types refused.
		{"no type", with(func(f *Filing) { f.Type = "" }), "missing-value", "ask needs --type: Answer, Decide, Look, Play"},
		{"Bless", with(func(f *Filing) { f.Type = "Bless" }), "bad-value", "--type Bless is the ladder's alone"},
		{"a type a pack would define", with(func(f *Filing) { f.Type = "Approve" }), "bad-value", "a type a pack defines is refused"},
		{"a type in the wrong case", with(func(f *Filing) { f.Type = "decide" }), "bad-value", `--type "decide" is not a type an agent files`},
		// Options.
		{"an option on an Answer", with(func(f *Filing) { f.Options = []string{"one"} }), "bad-flag", "--option goes only on a Decide"},
		{"five options", with(options("1", "2", "3", "4", "5")), "bad-value", "at most 4 --option, not 5"},
		{"two options the same", with(options("one", "one")), "bad-value", "each must differ"},
		{"two options the same once redacted", with(options("password=madeup-one", "password=madeup-two")), "bad-value", "each must differ"},
		{"an empty option", with(options("one", " ")), "bad-value", "an --option is empty"},
		{"an option of two lines", with(options("one\ntwo")), "bad-value", "--option must be one line"},
		{"an option of 200", with(options(long(200))), "", ""},
		{"an option of 201", with(options(long(201))), "bad-value", "--option is 201 characters once redacted, over its 200"},
		// Verdicts.
		{"a verdict on a Decide", with(func(f *Filing) { f.Type, f.Verdict = "Decide", "pass" }), "bad-flag", "--verdict goes only on a Look"},
		{"a verdict that is neither", with(func(f *Filing) { f.Type, f.Task, f.Verdict = "Look", "T-0901", "maybe" }), "bad-value", `not "maybe"`},
		// What it is about.
		{"a Look with no task", with(func(f *Filing) { f.Type = "Look" }), "missing-value", "a Look needs --task"},
		{"a Play about a document", with(func(f *Filing) { f.Type, f.Doc = "Play", "M-x" }), "missing-value", "a Play needs --task"},
		{"a task and a document", with(func(f *Filing) { f.Task, f.Doc = "T-0901", "M-x" }), "bad-flag", "--task or --doc, not both"},
		{"a task by its path", with(func(f *Filing) { f.Task = "work/tasks/T-0901.md" }), "bad-value", "never its path"},
		{"a task by a Windows path", with(func(f *Filing) { f.Task = `..\T-0901` }), "bad-value", "never its path"},
		{"a document by a file name", with(func(f *Filing) { f.Doc = "M-x.md" }), "bad-value", "never its path"},
		{"a memory id as a task", with(func(f *Filing) { f.Task = "M-colours" }), "bad-value", `--task "M-colours" is not a task's id`},
		{"a task id too short", with(func(f *Filing) { f.Task = "T-1" }), "bad-value", "is not a task's id: task (^T-[0-9]{4,6}$)"},
		{"a document of no kind", with(func(f *Filing) { f.Doc = "X-1" }), "bad-value", "matches no declared kind's id pattern"},
		{"a document id too long", with(func(f *Filing) { f.Doc = "M-" + long(119) }), "bad-value", "too long"},
		{"an id with a line feed", with(func(f *Filing) { f.Task = "T-0901\n" }), "bad-value", "is not a task's id"},
		// The title.
		{"no title", with(func(f *Filing) { f.Title = "" }), "missing-value", "--title is required"},
		{"a title of spaces", with(func(f *Filing) { f.Title = "   " }), "missing-value", "--title is required"},
		{"a title of two lines", with(func(f *Filing) { f.Title = "one\ntwo" }), "bad-value", "--title must be one line"},
		{"a title with CRLF", with(func(f *Filing) { f.Title = "one\r\ntwo" }), "bad-value", "--title must be one line"},
		{"a title of 300", with(func(f *Filing) { f.Title = long(300) }), "", ""},
		{"a title of 301", with(func(f *Filing) { f.Title = long(301) }), "bad-value", "--title is 301 characters once redacted, over its 300"},
		{"a title over 300 that the redactor brings under", with(func(f *Filing) {
			f.Title = long(260) + " password=" + strings.Repeat("x", 40)
		}), "", ""},
		// Why and then.
		{"no why", with(func(f *Filing) { f.Why = "" }), "missing-value", "--why is required"},
		{"a why of lines", with(func(f *Filing) { f.Why = "one\ntwo\r\nthree" }), "", ""},
		{"a why of 600", with(func(f *Filing) { f.Why = long(600) }), "", ""},
		{"a why of 601", with(func(f *Filing) { f.Why = long(601) }), "bad-value", "--why is 601 characters once redacted, over its 600"},
		{"a why with a lone carriage return", with(func(f *Filing) { f.Why = "one\rtwo" }), "bad-value", "(U+000D)"},
		{"a why with a tab", with(func(f *Filing) { f.Why = "one\ttwo" }), "bad-value", "(U+0009)"},
		{"a then", with(func(f *Filing) { f.Then = "The builder applies it." }), "", ""},
		{"a then of two lines", with(func(f *Filing) { f.Then = "one\ntwo" }), "bad-value", "--then must be one line"},
		{"a then of spaces", with(func(f *Filing) { f.Then = "  " }), "bad-value", "--then holds only spaces"},
		{"a then of 300", with(func(f *Filing) { f.Then = long(300) }), "", ""},
		{"a then of 301", with(func(f *Filing) { f.Then = long(301) }), "bad-value", "--then is 301 characters"},
		// Hidden characters, refused not stripped.
		{"a zero-width space", with(func(f *Filing) { f.Title = "a\u200bb" }), "bad-value", "--title holds a hidden or control character (U+200B)"},
		{"a right-to-left override", with(func(f *Filing) { f.Why = "a\u202eb" }), "bad-value", "(U+202E)"},
		{"a byte-order mark", with(func(f *Filing) { f.Title = "\ufeffa" }), "bad-value", "(U+FEFF)"},
		{"a private-use character", with(func(f *Filing) { f.Title = "a\ue000" }), "bad-value", "(U+E000)"},
		{"a line separator", with(func(f *Filing) { f.Why = "a\u2028b" }), "bad-value", "(U+2028)"},
		{"a paragraph separator", with(func(f *Filing) { f.Why = "a\u2029b" }), "bad-value", "(U+2029)"},
		{"a next-line control", with(func(f *Filing) { f.Then = "a\u0085b" }), "bad-value", "(U+0085)"},
		{"a variation selector", with(options("ok\ufe0f")), "bad-value", "--option holds a hidden or control character (U+FE0F)"},
		{"a supplementary variation selector", with(func(f *Filing) { f.Title = "a\U000e0101" }), "bad-value", "(U+E0101)"},
		{"a tag character", with(func(f *Filing) { f.Title = "a\U000e0041" }), "bad-value", "(U+E0041)"},
		{"an unassigned character", with(func(f *Filing) { f.Title = "a\u0378" }), "bad-value", "(U+0378)"},
		{"a NUL", with(func(f *Filing) { f.Title = "a\x00" }), "bad-value", "(U+0000)"},
		{"text that is not UTF-8", with(func(f *Filing) { f.Why = "a\xffb" }), "bad-value", "--why is not UTF-8 text"},
		{"an emoji and accents", with(func(f *Filing) { f.Title = "Caf\u00e9 \U0001F600 ok" }), "", ""},
		// Keys.
		{"a key", with(func(f *Filing) { f.Key = "colour-choice.v2_a" }), "", "agent:colour-choice.v2_a"},
		{"a key of 60", with(func(f *Filing) { f.Key = "k" + long(59) }), "", "agent:k" + long(59)},
		{"a key of 61", with(func(f *Filing) { f.Key = "k" + long(60) }), "bad-value", "is not a key"},
		{"a key with a colon", with(func(f *Filing) { f.Key = "a:b" }), "bad-value", "is not a key"},
		{"a key starting with a dot", with(func(f *Filing) { f.Key = ".a" }), "bad-value", "is not a key"},
		{"a key the redactor would change", with(func(f *Filing) { f.Key = "Zq7xYw3vUt9sRp1oLk5jHg2fDs8aQw4eRt6y" }), "bad-value",
			"--key looks like a secret to the redactor"},
		// The record's size.
		{"a record over 8,192 bytes", with(func(f *Filing) { f.Title, f.Why = emoji, emoji+emoji }), "bad-value",
			"bytes once written, over the 8192 an ask record may hold"},
	}
	kinds := testKinds(t)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := testPlace(t, "")
			done, e := File(p, c.f, kinds)
			if c.code != "" {
				if e == nil || e.Code != c.code || e.Exit != ExitInput || !strings.Contains(e.What, c.in) || e.Next == "" {
					t.Fatalf("want %s (exit 2) with %q, got %+v", c.code, c.in, e)
				}
				if got := files(t, p.Local.Main); len(got) != 0 {
					t.Errorf("a refusal wrote %v", got)
				}
				return
			}
			if e != nil {
				t.Fatalf("refused: %v", e)
			}
			if c.in != "" && done.Entry.Key != c.in {
				t.Errorf("key %s, want %s", done.Entry.Key, c.in)
			}
			recs := askRecords(t, p.Local.Main)
			if len(recs) != 1 || recs[0].Key != done.Entry.Key || recs[0].Op != OpFile || recs[0].Source != SourceAgent ||
				recs[0].Session != nil || val(recs[0].Type) != c.f.Type || done.Entry.State != Open || done.Written == nil {
				t.Errorf("the record: %+v", recs)
			}
			// The key can be checked from the record alone.
			if c.f.Key == "" {
				target := val(recs[0].Doc)
				if target == "" {
					target = val(recs[0].Task)
				}
				if target == "" {
					target = "answers"
				}
				if want := hashKey(val(recs[0].Type), target, val(recs[0].Title)); recs[0].Key != want {
					t.Errorf("key %s, its record's hash %s", recs[0].Key, want)
				}
			}
		})
	}
}

// The why's CRLF is stored as LF; the title kept as given past the redactor; the options in order; the doc or task.
func TestFileStores(t *testing.T) {
	p := testPlace(t, sessA)
	f := Filing{Type: "Decide", Title: "Which?", Why: "one\r\ntwo", Then: "then", Doc: "N-007", Options: []string{"b", "a"}}
	done, e := File(p, f, testKinds(t))
	if e != nil {
		t.Fatal(e)
	}
	r := askRecords(t, p.Local.Main)[0]
	if val(r.Why) != "one\ntwo" || val(r.ThenText) != "then" || val(r.Doc) != "N-007" || r.Task != nil ||
		strings.Join(r.Options, ",") != "b,a" || val(r.Session) != sessA || r.At != "2026-10-08T12:00:01.000Z" {
		t.Errorf("the record: %+v", r)
	}
	if done.Written.String("key") != r.Key {
		t.Errorf("written %s", schema.Show(done.Written))
	}
}

// The states of a key: filed, filed again, answered, the same answer again, another answer, resolved, filed again
// once closed; and the unknown key. A record is written only where the state says so, each with its log record.
func TestStates(t *testing.T) {
	p := testPlace(t, sessA)
	kinds := testKinds(t)
	file := func(title string) *Done {
		t.Helper()
		d, e := File(p, Filing{Type: "Decide", Title: title, Why: "w", Options: []string{"x", "y"}, Key: "k"}, kinds)
		if e != nil {
			t.Fatal(e)
		}
		return d
	}
	status := func(want string) *Entry {
		t.Helper()
		e, err := Status(p, "agent:k")
		if err != nil || e.State != want {
			t.Fatalf("status: %+v %v, want %s", e, err, want)
		}
		return e
	}
	answer := func(by, choice string) (*Done, *Error) {
		t.Helper()
		q := p
		q.Session = sessB
		return Answer(q, "agent:k", Answering{By: by, Choice: choice})
	}
	if _, e := Status(p, "agent:k"); e == nil || e.Code != "ask-not-open" || e.Exit != ExitState || !strings.Contains(e.What, "no ask has the key agent:k") {
		t.Fatalf("an unknown key: %v", e)
	}
	file("first")
	file("second")
	if e := status(Open); val(e.Filed.Title) != "second" {
		t.Errorf("the newest wording does not stand: %s", val(e.Filed.Title))
	}
	d, e := answer("", "y")
	if e != nil || d.Written == nil || d.Entry.State != Answered {
		t.Fatalf("the answer: %+v %v", d, e)
	}
	if e := status(Answered); e.Closed.Answer.By != Terminal || val(e.Closed.Answer.Choice) != "y" || e.Closed.Session == nil ||
		*e.Closed.Session != sessB || e.Closed.Type != nil || e.Closed.Title != nil || len(e.Closed.Options) != 0 {
		t.Errorf("the answer record: %+v", e.Closed)
	}
	before := files(t, p.Local.Main)
	// The same answer again (key and by): nothing written, exit 0, checked before the state.
	for _, choice := range []string{"y", "x"} {
		d, e = answer("terminal", choice)
		if e != nil || d.Written != nil || !strings.Contains(d.Nothing, "was answered already by terminal") {
			t.Errorf("the same answer again: %+v %v", d, e)
		}
	}
	// Another answer, and a resolve: the first answer stands.
	if _, e = answer("act:4711", "x"); e == nil || e.Code != "ask-not-open" || !strings.Contains(e.What, "answered already (by terminal") {
		t.Errorf("another answer: %v", e)
	}
	if _, e = Resolve(p, "agent:k"); e == nil || e.Code != "ask-not-open" || !strings.Contains(e.What, "answered already") {
		t.Errorf("a resolve after the answer: %v", e)
	}
	if after := files(t, p.Local.Main); !equalFiles(before, after) {
		t.Errorf("a refusal or a repeat wrote something")
	}
	// Filed again once closed: open again, and answered again.
	file("third")
	status(Open)
	if _, e = answer("act:4711", "x"); e != nil {
		t.Errorf("an answer once filed again: %v", e)
	}
	// Resolved, then resolved again and answered: neither is open.
	file("fourth")
	if d, e = Resolve(p, "agent:k"); e != nil || d.Written == nil || d.Entry.State != Resolved {
		t.Errorf("resolve: %+v %v", d, e)
	}
	before = files(t, p.Local.Main)
	if _, e = Resolve(p, "agent:k"); e == nil || e.Code != "ask-not-open" || !strings.Contains(e.What, "agent:k is resolved already (at") {
		t.Errorf("resolve again: %v", e)
	}
	if _, e = answer("", "x"); e == nil || e.Code != "ask-not-open" || !strings.Contains(e.What, "was withdrawn (resolved at") {
		t.Errorf("an answer to a resolved ask: %v", e)
	}
	if after := files(t, p.Local.Main); !equalFiles(before, after) {
		t.Errorf("a repeat or a refusal wrote something")
	}
	// Each ask record written has its one ask log record, in the same order, its kind the op and its target the key.
	recs := askRecords(t, p.Local.Main)
	logs := logRecords(t, p.Local.Main)
	ops := []string{}
	for _, r := range recs {
		ops = append(ops, r.Op)
	}
	if strings.Join(ops, ",") != "file,file,answer,file,answer,file,resolve" {
		t.Errorf("the ask records' ops: %v", ops)
	}
	var kinds2 []string
	for name, ls := range logs {
		for _, l := range ls {
			if l.Event != "ask" || val(l.Target) != "agent:k" || l.Text != nil {
				t.Errorf("%s: %+v", name, l)
			}
			kinds2 = append(kinds2, l.At+" "+val(l.Kind)+" "+val(l.Session))
		}
	}
	sort.Strings(kinds2)
	want := []string{}
	for _, r := range recs {
		want = append(want, r.At+" "+r.Op+" "+val(r.Session))
	}
	if strings.Join(kinds2, "\n") != strings.Join(want, "\n") {
		t.Errorf("the log records:\n%s\nthe ask records:\n%s", strings.Join(kinds2, "\n"), strings.Join(want, "\n"))
	}
	// The asking session's records are in its own file, the answering one's in its own.
	if len(logs["s-"+sessA+".ndjson"]) != 5 || len(logs["s-"+sessB+".ndjson"]) != 2 || len(logs) != 2 {
		t.Errorf("the log files: %v", keys(logs))
	}
}

func equalFiles(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func keys[V any](m map[string]V) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Every answer case: the session that asked refused, the choice among the options, the verdict on a Look, the
// limits of by, via and the words, an empty answer, an unknown key, a key that is not one.
func TestAnswer(t *testing.T) {
	kinds := testKinds(t)
	long := func(n int) string { return strings.Repeat("w", n) }
	cases := []struct {
		name    string
		typ     string
		session string // the answering session
		a       Answering
		key     string // "" for the filed one
		code    string
		in      string
	}{
		{"a Decide's choice", "Decide", "", Answering{Choice: "Darker"}, "", "", ""},
		{"a choice redacted as the option was", "Decide", "", Answering{Choice: "Mine password=madeup-pw"}, "", "", ""},
		{"a choice not among the options", "Decide", "", Answering{Choice: "darker"}, "", "bad-value", `--choice "darker" is not one of the ask's options: "Lighter", "Darker"`},
		{"a choice on a Look", "Look", "", Answering{Choice: "Darker"}, "", "bad-flag", "--choice answers a Decide; agent:k is a Look"},
		{"a choice on an Answer", "Answer", "", Answering{Choice: "Darker"}, "", "bad-flag", "is an Answer"},
		{"a Look's verdict", "Look", "", Answering{Verdict: "fail", Words: "The contrast is low."}, "", "", ""},
		{"a verdict on a Decide", "Decide", "", Answering{Verdict: "pass"}, "", "bad-flag", "--verdict answers a Look"},
		{"a verdict that is neither", "Look", "", Answering{Verdict: "ok"}, "", "bad-value", `not "ok"`},
		{"words of lines", "Play", "", Answering{Words: "one\r\ntwo"}, "", "", ""},
		{"words of 2,000", "Answer", "", Answering{Words: long(2000)}, "", "", ""},
		{"words of 2,001", "Answer", "", Answering{Words: long(2001)}, "", "bad-value", "--words is 2001 characters once redacted, over its 2000"},
		{"words with a hidden character", "Answer", "", Answering{Words: "a\u200db"}, "", "bad-value", "--words holds a hidden or control character (U+200D)"},
		{"a by of 60", "Answer", "", Answering{By: "act:" + long(56), Words: "w"}, "", "", ""},
		{"a by of 61", "Answer", "", Answering{By: "act:" + long(57), Words: "w"}, "", "bad-value", "--by is 61 characters"},
		{"a by of two lines", "Answer", "", Answering{By: "a\nb", Words: "w"}, "", "bad-value", "--by must be one line"},
		{"a via of 30", "Answer", "", Answering{Via: long(30), Words: "w"}, "", "", ""},
		{"a via of 31", "Answer", "", Answering{Via: long(31), Words: "w"}, "", "bad-value", "--via is 31 characters"},
		{"an empty answer", "Answer", "", Answering{By: "act:1"}, "", "missing-value", "an answer needs --choice, --verdict or --words"},
		{"the session that asked", "Answer", sessA, Answering{Words: "w"}, "", "answer-own-session", "the session that filed agent:k cannot answer it"},
		{"another session", "Answer", sessB, Answering{Words: "w"}, "", "", ""},
		{"an unknown key", "Answer", "", Answering{Words: "w"}, "agent:other", "ask-not-open", "no ask has the key agent:other"},
		{"a key that is not one", "Answer", "", Answering{Words: "w"}, "agent/k", "bad-value", `"agent/k" is not an ask's key`},
		{"a ladder's key with no ask", "Answer", "", Answering{Words: "w"}, "ladder:bless-T-0901", "ask-not-open", "no ask has the key"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := testPlace(t, sessA)
			f := Filing{Type: c.typ, Title: "t", Why: "w", Key: "k"}
			switch c.typ {
			case "Decide":
				f.Options = []string{"Lighter", "Darker", "Mine password=other-madeup"}[:2]
				if strings.HasPrefix(c.a.Choice, "Mine") {
					f.Options = []string{"Lighter", "Mine password=other-madeup"}
				}
			case "Look", "Play":
				f.Task = "T-0901"
			}
			if _, e := File(p, f, kinds); e != nil {
				t.Fatal(e)
			}
			before := files(t, p.Local.Main)
			key := c.key
			if key == "" {
				key = "agent:k"
			}
			p.Session = c.session
			done, e := Answer(p, key, c.a)
			if c.code != "" {
				want := ExitInput
				if c.code == "ask-not-open" || c.code == "answer-own-session" {
					want = ExitState
				}
				if e == nil || e.Code != c.code || e.Exit != want || !strings.Contains(e.What, c.in) {
					t.Fatalf("want %s (exit %d) with %q, got %+v", c.code, want, c.in, e)
				}
				if !equalFiles(before, files(t, p.Local.Main)) {
					t.Errorf("a refusal wrote something")
				}
				return
			}
			if e != nil {
				t.Fatalf("refused: %v", e)
			}
			a := done.Entry.Closed.Answer
			if done.Entry.State != Answered || done.Written == nil || a == nil {
				t.Fatalf("the answer: %+v", done)
			}
			if c.a.By == "" && a.By != Terminal || c.a.Via == "" && a.Via != nil || c.a.Words != "" && strings.Contains(val(a.Words), "\r") {
				t.Errorf("by %q via %v words %q", a.By, a.Via, val(a.Words))
			}
		})
	}
}

// Every free-text field is redacted before it is stored: a made-up secret in the title, why, then, an option, the
// choice, the words, by and via is in no file, and the redactor's marker is.
func TestRedacted(t *testing.T) {
	p := testPlace(t, "")
	secrets := []string{"MadeUpTitleSecret1", "MadeUpWhySecret22", "MadeUpThenSecret3", "MadeUpOptionSecret4", "MadeUpWordsSecret5",
		"MadeUpBySecret66", "MadeUpVia7"}
	f := Filing{Type: "Decide", Title: "deploy with api_key=" + secrets[0], Why: "the password: " + secrets[1],
		Then: "then token=" + secrets[2], Options: []string{"keep", "use secret=" + secrets[3]}, Key: "r"}
	if _, e := File(p, f, testKinds(t)); e != nil {
		t.Fatal(e)
	}
	if _, e := Answer(p, "agent:r", Answering{Choice: "use secret=" + secrets[3], Words: "Authorization: Bearer " + secrets[4],
		By: "act:1 token=" + secrets[5], Via: "password=" + secrets[6]}); e != nil {
		t.Fatal(e)
	}
	all := ""
	for _, v := range files(t, p.Local.Main) {
		all += v
	}
	for _, s := range secrets {
		if strings.Contains(all, s) {
			t.Errorf("the secret %s is in a file", s)
		}
	}
	if n := strings.Count(all, "[redacted]"); n < len(secrets)+1 {
		t.Errorf("%d markers:\n%s", n, all)
	}
	e, _ := Status(p, "agent:r")
	if val(e.Closed.Answer.Choice) != e.Filed.Options[1] {
		t.Errorf("the choice %q, the option %q", val(e.Closed.Answer.Choice), e.Filed.Options[1])
	}
}

// Each record goes into the file of its own UTC day, so an answer the next day goes into the next day's file, and
// its log record into its own day's file when there is no session.
func TestDayFiles(t *testing.T) {
	p := testPlace(t, "")
	at := time.Date(2026, 10, 8, 23, 59, 59, 500e6, time.UTC)
	p.Now = func() time.Time { return at }
	if _, e := File(p, Filing{Type: "Answer", Title: "t", Why: "w", Key: "d"}, nil); e != nil {
		t.Fatal(e)
	}
	at = at.Add(time.Second)
	if _, e := Answer(p, "agent:d", Answering{Words: "w"}); e != nil {
		t.Fatal(e)
	}
	got := keys(files(t, p.Local.Main))
	want := []string{".bonsai/.gitignore", ".bonsai/local/asks/2026-10-08.ndjson", ".bonsai/local/asks/2026-10-09.ndjson",
		".bonsai/local/log/w-2026-10-08.ndjson", ".bonsai/local/log/w-2026-10-09.ndjson"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("the files:\n%s", strings.Join(got, "\n"))
	}
	if e, _ := Status(p, "agent:d"); e == nil || e.State != Answered {
		t.Errorf("the state across two days: %+v", e)
	}
	// A session id that cannot name a log file is taken as none.
	p.Session = "../escape"
	if d, e := File(p, Filing{Type: "Answer", Title: "t2", Why: "w"}, nil); e != nil || d.Entry.Filed.Session != nil {
		t.Errorf("a session id that names no file: %+v %v", d, e)
	}
}

// The reader: a torn line skipped and counted, a record glued after it read, a blank line passed over, a record that
// would close a closed key changing nothing (the first answer stands), a key with no file record not listed, a file
// that is not a day's not read; the list newest first, open ones only unless all.
func TestRead(t *testing.T) {
	p := testPlace(t, "")
	for _, k := range []string{"a", "b", "c"} {
		if _, e := File(p, Filing{Type: "Answer", Title: "title " + k, Why: "w", Key: k}, nil); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := Answer(p, "agent:b", Answering{Words: "first"}); e != nil {
		t.Fatal(e)
	}
	day := filepath.Join(Folder(p.Local.Main), "2026-10-08.ndjson")
	// A second answer that raced the first, a torn line with a record glued after it, a blank line, an answer to a
	// key never filed, and a file that is not a day's.
	second := &format.Ask{ID: record.NewID(), At: "2026-10-08T12:00:09.000Z", Op: OpAnswer, Key: "agent:b", Workspace: testWS,
		Source: SourceAgent, Options: []string{}, Answer: &format.AskAnswer{By: "act:2", Words: ptr("second")}}
	stray := *second
	stray.Key = "agent:never"
	resolve := &format.Ask{ID: record.NewID(), At: "2026-10-08T12:00:10.000Z", Op: OpResolve, Key: "agent:c", Workspace: testWS,
		Source: SourceAgent, Options: []string{}}
	var add []byte
	for _, r := range []*format.Ask{second, &stray} {
		line, err := r.Encode()
		if err != nil {
			t.Fatal(err)
		}
		add = append(add, line...)
	}
	line, _ := resolve.Encode()
	add = append(append(append(add, `{"format":"bonsai.ask/1","id":"torn`...), line...), "\n"...)
	f, err := os.OpenFile(day, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(add); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	if err := os.WriteFile(filepath.Join(Folder(p.Local.Main), "notes.ndjson"), []byte("not an ask\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := Read(p.Local.Main, "")
	if err != nil {
		t.Fatal(err)
	}
	if b.Unreadable != 1 || b.Get("agent:never") != nil || b.Get("agent:c").State != Resolved {
		t.Errorf("unreadable %d, never %v, c %+v", b.Unreadable, b.Get("agent:never"), b.Get("agent:c"))
	}
	if e := b.Get("agent:b"); e.State != Answered || val(e.Closed.Answer.Words) != "first" {
		t.Errorf("the first answer does not stand: %+v", e.Closed.Answer)
	}
	var open, all []string
	for _, e := range b.Entries(false) {
		open = append(open, e.Key)
	}
	for _, e := range b.Entries(true) {
		all = append(all, e.Key)
	}
	if strings.Join(open, ",") != "agent:a" || strings.Join(all, ",") != "agent:c,agent:b,agent:a" {
		t.Errorf("open %v, all %v", open, all)
	}
	// Reading one key reads its records alone.
	if one, err := Read(p.Local.Main, "agent:b"); err != nil || one.Get("agent:a") != nil || one.Get("agent:b") == nil {
		t.Errorf("one key: %v", err)
	}
}

// The writes that fail: .bonsai/.gitignore that cannot be restored and an asks file that cannot be written write
// nothing (write-failed, exit 3); a log record that cannot be written after its ask record is partly-written (exit 3),
// the ask record standing.
func TestWriteFails(t *testing.T) {
	defer func(g func(string) (bool, error), a func(string, string, []byte, int) (bool, error),
		w func(string, *format.Log) (bool, error)) {
		ensureGitignore, appendLine, writeLog = g, a, w
	}(ensureGitignore, appendLine, writeLog)
	fail := errors.New("the disk is full")
	filing := Filing{Type: "Answer", Title: "t", Why: "w"}

	ensureGitignore = func(string) (bool, error) { return false, fail }
	p := testPlace(t, "")
	if _, e := File(p, filing, nil); e == nil || e.Code != "write-failed" || e.Exit != ExitRuntime || len(files(t, p.Local.Main)) != 0 {
		t.Errorf("no .gitignore: %v", e)
	}
	ensureGitignore = workspace.EnsureGitignore

	appendLine = func(string, string, []byte, int) (bool, error) { return false, fail }
	p = testPlace(t, "")
	if _, e := File(p, filing, nil); e == nil || e.Code != "write-failed" || !strings.Contains(e.What, "asks/2026-10-08.ndjson cannot be written") {
		t.Errorf("no asks file: %v", e)
	}
	appendLine = record.Append

	writeLog = func(string, *format.Log) (bool, error) { return false, fail }
	p = testPlace(t, "")
	if _, e := File(p, filing, nil); e == nil || e.Code != "partly-written" || e.Exit != ExitRuntime ||
		!strings.Contains(e.What, "was written, but its log record was not") {
		t.Errorf("no log record: %v", e)
	}
	if recs := askRecords(t, p.Local.Main); len(recs) != 1 {
		t.Errorf("the ask record does not stand: %d", len(recs))
	}
}

// Resolve withdraws an agent's ask only, and refuses a key that is not one.
func TestResolveKeys(t *testing.T) {
	p := testPlace(t, "")
	for key, code := range map[string]string{"ladder:bless-T-0901": "bad-value", "agent:": "bad-value", "agent:nothing": "ask-not-open"} {
		if _, e := Resolve(p, key); e == nil || e.Code != code {
			t.Errorf("%s: %v, want %s", key, e, code)
		}
	}
}

// The agent types are format.AskTypes but the ladder's, and the ladder's is in that table.
func TestAgentTypes(t *testing.T) {
	if got := strings.Join(AgentTypes(), ","); got != "Answer,Decide,Look,Play" {
		t.Errorf("agent types %s", got)
	}
	found := false
	for _, w := range format.AskTypes {
		found = found || w.Word == LadderType
	}
	if !found {
		t.Errorf("%s is not in format.AskTypes", LadderType)
	}
}

// The hidden characters: what is refused and what is not.
func TestHidden(t *testing.T) {
	for _, r := range []rune{0, '\t', '\n', '\r', 0x7f, 0x85, 0xad, 0x200b, 0x200d, 0x202e, 0x2060, 0x2028, 0x2029, 0xfeff, 0xe000,
		0xf8ff, 0xfe00, 0xfe0f, 0x180b, 0xfffe, 0xffff, 0x0378, 0xe0001, 0xe0101, 0x10fffd} {
		if !Hidden(r) {
			t.Errorf("U+%04X is not hidden", r)
		}
	}
	for _, r := range []rune{' ', 'a', 0xe9, 0x3000, 0x00a0, 0x1f600, 0x4e2d, 0x0915, 0x093f, 0xfffd} {
		if Hidden(r) {
			t.Errorf("U+%04X is hidden", r)
		}
	}
}
