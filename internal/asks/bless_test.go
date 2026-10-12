package asks

// The ladder's own asks (design/plan-5.md, 5.4.4 notes 2 and 3): the Bless ask filed with source ladder, its key, its
// task and its data, and its ask log record, each held to its schema as written; a second filing replacing the first's
// numbers; the ladder's resolve; every refusal writing nothing; agents still refused the type and the ladder's keys.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
)

// blessData is an offer as the ladder files it: one rise, its new tests given or null.
func blessData(now int64, newTests bool) schema.Object {
	var nt any
	if newTests {
		nt = schema.Object{{Key: "base", Value: "3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e"}, {Key: "count", Value: int64(3)},
			{Key: "failed_on_base", Value: int64(3)}, {Key: "passed_on_base", Value: int64(0)}, {Key: "not_run", Value: int64(0)}}
	}
	return schema.Object{{Key: "task", Value: "T-0901"}, {Key: "head", Value: "5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f80"},
		{Key: "result_sha", Value: strings.Repeat("0f", 32)},
		{Key: "rises", Value: []any{schema.Object{{Key: "name", Value: "go_tests"}, {Key: "was", Value: json.Number("120")},
			{Key: "now", Value: json.Number(jsonInt(now))}, {Key: "new_tests", Value: nt}}}}}
}

func jsonInt(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

var okBless = Bless{Task: "T-0901", Title: "Bless T-0901's ratchets", Why: "go_tests rose from 120 to 123.",
	Then: "Blessing raises the floor in bonsai.yaml from that run; it runs nothing."}

// asWritten holds an asks file's lines and a log file's lines to their schemas as a writer writes them: every field,
// in order.
func asWritten(t *testing.T, name string, raw []byte) {
	t.Helper()
	s := format.MustLookup(name).Schema()
	for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		doc, err := schema.Decode([]byte(line))
		if err != nil {
			t.Fatalf("not JSON: %v\n%s", err, line)
		}
		if msgs := append(schema.Validate(s, doc), schema.CheckOrder(s, doc)...); len(msgs) > 0 {
			t.Fatalf("not bonsai.%s as written: %v\n%s", name, msgs, line)
		}
	}
}

// The Bless ask: source ladder, type Bless, key ladder:bless-<task>, its task, words and data; its ask log record in
// the climbing session's file; a second filing appends a new file record whose numbers stand; the ladder's resolve
// closes it, every field of the ask null; a second resolve is ask-not-open; filing again opens it again.
func TestFileBless(t *testing.T) {
	p := testPlace(t, sessA)
	b := okBless
	b.Data = blessData(123, true)
	d, e := FileBless(p, b)
	if e != nil {
		t.Fatal(e)
	}
	r := d.Entry.Filed
	if d.Entry.Key != "ladder:bless-T-0901" || r.Source != SourceLadder || val(r.Type) != LadderType || val(r.Task) != "T-0901" ||
		val(r.Title) != okBless.Title || val(r.Why) != okBless.Why || val(r.ThenText) != okBless.Then || r.Doc != nil ||
		len(r.Options) != 0 || r.Verdict != nil || r.Answer != nil || val(r.Session) != sessA || r.Op != OpFile {
		t.Errorf("the record: %+v", r)
	}
	if got, want := schema.Show(r.Data), schema.Show(blessData(123, true)); got != want {
		t.Errorf("data %s, want %s", got, want)
	}
	// A later climb's numbers replace the first's: a new file record, and the entry's latest wins.
	b.Data, b.Why = blessData(125, false), "go_tests rose from 120 to 125."
	if _, e := FileBless(p, b); e != nil {
		t.Fatal(e)
	}
	got, err := Read(p.Local.Main, "")
	if err != nil {
		t.Fatal(err)
	}
	if en := got.Get("ladder:bless-T-0901"); en == nil || en.State != Open || schema.Show(en.Filed.Data) != schema.Show(blessData(125, false)) ||
		val(en.Filed.Why) != b.Why {
		t.Errorf("the second filing does not stand: %+v", en)
	}
	// The ladder resolves it once its offer has gone: a resolve record with every field of the ask null.
	d, e = ResolveLadder(p, "ladder:bless-T-0901")
	if e != nil {
		t.Fatal(e)
	}
	c := d.Entry.Closed
	if d.Entry.State != Resolved || c.Op != OpResolve || c.Source != SourceLadder || c.Type != nil || c.Task != nil || c.Title != nil ||
		c.Why != nil || c.ThenText != nil || c.Data != nil || len(c.Options) != 0 || val(c.Session) != sessA {
		t.Errorf("the resolve record: %+v", c)
	}
	if _, e := ResolveLadder(p, "ladder:bless-T-0901"); e == nil || e.Code != "ask-not-open" || e.Exit != ExitState ||
		!strings.Contains(e.What, "is resolved already") {
		t.Errorf("a second resolve: %v", e)
	}
	// Filing a closed key opens it again.
	if _, e := FileBless(p, b); e != nil {
		t.Fatal(e)
	}
	if got, _ := Read(p.Local.Main, ""); got.Get("ladder:bless-T-0901").State != Open {
		t.Error("filing a resolved key did not open it again")
	}
	// Every record and every log record as written; each ask record has its one ask log record, in the session's file.
	ops := []string{}
	for _, r := range askRecords(t, p.Local.Main) {
		ops = append(ops, r.Op+":"+r.Source+":"+val(r.Type))
	}
	if strings.Join(ops, ",") != "file:ladder:Bless,file:ladder:Bless,resolve:ladder:,file:ladder:Bless" {
		t.Errorf("the ask records: %v", ops)
	}
	entries, _ := os.ReadDir(Folder(p.Local.Main))
	for _, en := range entries {
		raw, _ := os.ReadFile(filepath.Join(Folder(p.Local.Main), en.Name()))
		asWritten(t, "ask", raw)
	}
	logs := logRecords(t, p.Local.Main)
	file := "s-" + sessA + ".ndjson"
	if len(logs) != 1 || len(logs[file]) != 4 {
		t.Fatalf("the log files: %v", keys(logs))
	}
	var kinds []string
	for _, l := range logs[file] {
		if l.Event != "ask" || val(l.Target) != "ladder:bless-T-0901" || l.Text != nil || val(l.Agent) != Agent {
			t.Errorf("a log record: %+v", l)
		}
		kinds = append(kinds, val(l.Kind))
	}
	if strings.Join(kinds, ",") != "file,file,resolve,file" {
		t.Errorf("the log records' kinds: %v", kinds)
	}
	raw, _ := os.ReadFile(filepath.Join(p.Local.Main, ".bonsai", "local", "log", file))
	asWritten(t, "log", raw)
}

// Without a session the Bless ask's session is null and its log record goes to the day's file.
func TestFileBlessNoSession(t *testing.T) {
	p := testPlace(t, "")
	b := okBless
	b.Data = blessData(123, false)
	d, e := FileBless(p, b)
	if e != nil {
		t.Fatal(e)
	}
	if d.Entry.Filed.Session != nil {
		t.Errorf("session %v", *d.Entry.Filed.Session)
	}
	if logs := logRecords(t, p.Local.Main); len(logs) != 1 || len(logs["w-2026-10-08.ndjson"]) != 1 {
		t.Errorf("the log files: %v", keys(logs))
	}
}

// Every refusal of the ladder's filing names its rule and writes nothing.
func TestFileBlessRefused(t *testing.T) {
	long := strings.Repeat("x", WhyMax+1)
	for _, c := range []struct {
		name, code, says string
		mod              func(*Bless)
	}{
		{"no task", "bad-value", "cannot make a Bless ask's key", func(b *Bless) { b.Task = "" }},
		{"a path for a task", "bad-value", "cannot make a Bless ask's key", func(b *Bless) { b.Task = "tasks/T-0901" }},
		{"a task too long for a key", "bad-value", "cannot make a Bless ask's key", func(b *Bless) { b.Task = strings.Repeat("T", 55) }},
		{"a space in the task", "bad-value", "cannot make a Bless ask's key", func(b *Bless) { b.Task = "T 0901" }},
		{"no data", "missing-value", "carries its offer in data", func(b *Bless) { b.Data = nil }},
		{"no title", "missing-value", "the Bless ask's title is required", func(b *Bless) { b.Title = " " }},
		{"a title on two lines", "bad-value", "the Bless ask's title must be one line", func(b *Bless) { b.Title = "a\nb" }},
		{"no why", "missing-value", "the Bless ask's why is required", func(b *Bless) { b.Why = "" }},
		{"a why too long", "bad-value", "the Bless ask's why is 601 characters", func(b *Bless) { b.Why = long }},
		{"a hidden character", "bad-value", "hidden or control character (U+200B)", func(b *Bless) { b.Why = "a​b" }},
		{"a then on two lines", "bad-value", "the Bless ask's then must be one line", func(b *Bless) { b.Then = "a\nb" }},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := testPlace(t, "")
			b := okBless
			b.Data = blessData(123, false)
			c.mod(&b)
			d, e := FileBless(p, b)
			if e == nil || d != nil || e.Code != c.code || e.Exit != ExitInput || !strings.Contains(e.What, c.says) || e.Next == "" {
				t.Errorf("%v, want %s %q", e, c.code, c.says)
			}
			if got := files(t, p.Local.Main); len(got) != 0 {
				t.Errorf("a refusal wrote %v", got)
			}
		})
	}
}

// The ladder resolves only its own keys; agents still cannot file a Bless nor withdraw the ladder's ask; an answered
// Bless is not resolved.
func TestLadderKeys(t *testing.T) {
	p := testPlace(t, "")
	for key, code := range map[string]string{"agent:k": "bad-value", "ladder:": "bad-value", "bless-T-0901": "bad-value",
		"ladder:bless-T-0999": "ask-not-open"} {
		if _, e := ResolveLadder(p, key); e == nil || e.Code != code {
			t.Errorf("%s: %v, want %s", key, e, code)
		}
	}
	b := okBless
	b.Data = blessData(123, false)
	if _, e := FileBless(p, b); e != nil {
		t.Fatal(e)
	}
	if _, e := Resolve(p, "ladder:bless-T-0901"); e == nil || e.Code != "bad-value" || !strings.Contains(e.What, "bonsai ladder resolves itself") {
		t.Errorf("an agent withdrawing the ladder's ask: %v", e)
	}
	if _, e := File(p, with(func(f *Filing) { f.Type = LadderType; f.Task = "T-0901" }), testKinds(t)); e == nil || e.Code != "bad-value" {
		t.Errorf("an agent filing a Bless: %v", e)
	}
	if _, e := Answer(testPlace2(p, sessB), "ladder:bless-T-0901", Answering{Words: "yes"}); e != nil {
		t.Fatal(e)
	}
	if _, e := ResolveLadder(p, "ladder:bless-T-0901"); e == nil || e.Code != "ask-not-open" || !strings.Contains(e.What, "is answered already") {
		t.Errorf("resolving an answered Bless: %v", e)
	}
}

// testPlace2 is p's main checkout, written from another session.
func testPlace2(p Place, session string) Place {
	p.Session = session
	return p
}

// BlessKey makes a key only from a task's id that fits one; IsBless reads it back.
func TestBlessKey(t *testing.T) {
	for task, want := range map[string]string{"T-0901": "ladder:bless-T-0901", "T-5407": "ladder:bless-T-5407", "": "",
		"a/b": "", "a b": "", strings.Repeat("T", 54): "ladder:bless-" + strings.Repeat("T", 54), strings.Repeat("T", 55): ""} {
		if got := BlessKey(task); got != want {
			t.Errorf("BlessKey(%q) = %q, want %q", task, got, want)
		}
	}
	if task, ok := IsBless("ladder:bless-T-0901"); !ok || task != "T-0901" {
		t.Errorf("IsBless: %q %v", task, ok)
	}
	for _, key := range []string{"ladder:bless-", "agent:bless-T-0901", "ladder:other"} {
		if _, ok := IsBless(key); ok {
			t.Errorf("IsBless(%q)", key)
		}
	}
}

// A Bless record written whose log record was not is partly-written (exit 3), and its Done keeps the record written, so
// the climb's bless names it; a filing that writes nothing gives no Done.
func TestBlessPartlyWritten(t *testing.T) {
	defer func(w func(string, *format.Log) (bool, error)) { writeLog = w }(writeLog)
	p := testPlace(t, "")
	b := okBless
	b.Data = blessData(123, false)
	writeLog = func(string, *format.Log) (bool, error) { return false, os.ErrPermission }
	d, e := FileBless(p, b)
	if e == nil || e.Code != "partly-written" || e.Exit != ExitRuntime || d == nil || d.Written == nil {
		t.Fatalf("filing: %+v, %v", d, e)
	}
	d, e = ResolveLadder(p, "ladder:bless-T-0901")
	if e == nil || e.Code != "partly-written" || d == nil || d.Written == nil || d.Entry.State != Resolved {
		t.Fatalf("resolving: %+v, %v", d, e)
	}
	if recs := askRecords(t, p.Local.Main); len(recs) != 2 {
		t.Errorf("the records do not stand: %d", len(recs))
	}
	appendLine = func(string, string, []byte, int) (bool, error) { return false, os.ErrPermission }
	defer func() { appendLine = record.Append }()
	if d, e := FileBless(p, b); e == nil || e.Code != "write-failed" || d != nil {
		t.Errorf("no asks file: %+v, %v", d, e)
	}
}
