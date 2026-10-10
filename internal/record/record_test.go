package record

// Tests of the common fields' builder and of Line, where a record is cut to fit. The records are read back through
// the log type's own reader and the line's decoded object, never against a list of the log's fields held here, so a
// field added at the log's end (formats set 5) changes nothing in these tests.

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// decode reads a line back as an ordered object.
func decode(t *testing.T, line []byte) schema.Object {
	t.Helper()
	v, err := schema.Decode(bytes.TrimSuffix(line, []byte("\n")))
	if err != nil {
		t.Fatalf("%v: %s", err, line)
	}
	return v.(schema.Object)
}

// fakeRedact stands in for the redactor (internal/redact): it hides one made-up secret.
func fakeRedact(s string) string { return strings.ReplaceAll(s, "hunter2", "[redacted]") }

// New fills the common fields, in the schema's order, and leaves every other field null.
func TestNew(t *testing.T) {
	root := filepath.Join(t.TempDir(), "example-project")
	env := map[string]string{"BONSAI_TASK": "T-0001 hunter2", "BONSAI_ROLE": "builder"}
	at := time.Date(2026, 10, 10, 9, 8, 7, 654321000, time.FixedZone("east", 3600))
	l := New("event", Common{Workspace: raceWorkspace, Local: workspace.Local{Root: root, Main: root, Branch: "t0001-thing"},
		Session: raceSession, Agent: "claude-code", Getenv: func(k string) string { return env[k] }, Redact: fakeRedact, At: at})
	line, err := Line(l)
	if err != nil {
		t.Fatal(err)
	}
	doc := decode(t, line)
	want := []schema.Member{
		{Key: "format", Value: "bonsai.log/1"},
		{Key: "at", Value: "2026-10-10T08:08:07.654Z"},
		{Key: "workspace", Value: raceWorkspace},
		{Key: "session", Value: raceSession},
		{Key: "agent", Value: "claude-code"},
		{Key: "event", Value: "event"},
		{Key: "checkout", Value: "example-project"},
		{Key: "branch", Value: "t0001-thing"},
		{Key: "task", Value: "T-0001 [redacted]"},
		{Key: "role", Value: "builder"},
	}
	last := -1
	for _, m := range want {
		i := doc.Index(m.Key)
		if i < 0 || doc[i].Value != m.Value {
			t.Errorf("%s: %v, want %v", m.Key, schema.Show(doc[i].Value), m.Value)
		}
		if i < last {
			t.Errorf("%s is out of the schema's order", m.Key)
		}
		last = i
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(doc.String("id")) {
		t.Errorf("id %q is not a version 4 UUID", doc.String("id"))
	}
	if labels, _ := doc.Get("labels"); len(labels.(schema.Object)) != 0 {
		t.Errorf("labels %v, want {}", labels)
	}
	set := map[string]bool{"format": true, "id": true, "labels": true}
	for _, m := range want {
		set[m.Key] = true
	}
	for _, m := range doc {
		if !set[m.Key] && m.Value != nil {
			t.Errorf("%s is %v, want null", m.Key, schema.Show(m.Value))
		}
	}
	// No redactor: task and role are null, never raw. No session: session and agent null. No time: now.
	before := time.Now().UTC().Truncate(time.Millisecond)
	l = New("clean", Common{Workspace: raceWorkspace, Getenv: func(k string) string { return env[k] }})
	if l.Task != nil || l.Role != nil || l.Session != nil || l.Agent != nil || l.Checkout != nil || l.Branch != nil {
		t.Errorf("with nothing given: %+v", l)
	}
	if got, err := time.Parse(AtLayout, l.At); err != nil || got.Before(before) || got.After(time.Now().Add(time.Second)) {
		t.Errorf("at %q, %v; want now", l.At, err)
	}
	if New("event", Common{}).ID == New("event", Common{}).ID {
		t.Errorf("two records share an id")
	}
}

func TestCheckoutName(t *testing.T) {
	for in, want := range map[string]string{"": "", filepath.FromSlash("/"): "", filepath.FromSlash("/srv/example-project"): "example-project",
		filepath.FromSlash("/srv/example-project/"): "example-project", ".": ""} {
		if got := CheckoutName(in); got != want {
			t.Errorf("CheckoutName(%q) = %q, want %q", in, got, want)
		}
	}
}

// Line gives the log type's own line for a record that fits; cuts target and text to their schema's lengths; and
// halves the longest text while the line is over 2,048 bytes, a label's value among them, never the fields that name
// or pair a record. The record given is not changed.
func TestLine(t *testing.T) {
	max := format.MustLookup("log").MaxLine()
	l := New("event", Common{Workspace: raceWorkspace, Session: raceSession})
	line, err := Line(l)
	if err != nil {
		t.Fatal(err)
	}
	if enc, err := l.Encode(); err != nil || !bytes.Equal(enc, line) {
		t.Errorf("a record that fits: Line %s, Encode %s, %v", line, enc, err)
	}

	// Over the schema's lengths: target and text cut to 200 and 300 characters.
	target, text := strings.Repeat("t", 500), strings.Repeat("x", 1000)
	l.Target, l.Text = &target, &text
	doc := decode(t, mustLine(t, l))
	if s := doc.String("target"); utf8.RuneCountInString(s) != 200 || !strings.HasSuffix(s, "...") {
		t.Errorf("target cut to %d characters: %q", utf8.RuneCountInString(s), s)
	}
	if s := doc.String("text"); utf8.RuneCountInString(s) != 300 || !strings.HasSuffix(s, "...") {
		t.Errorf("text cut to %d characters", utf8.RuneCountInString(s))
	}
	if *l.Target != target || *l.Text != text {
		t.Errorf("Line changed the record given")
	}

	// Over 2,048 bytes: each character outside ASCII is six bytes on the line. The longest goes first, then the next.
	target, text = strings.Repeat("é", 200), strings.Repeat("é", 300)
	branch := strings.Repeat("b", 120)
	l.Target, l.Text, l.Branch = &target, &text, &branch
	l.Labels = schema.Object{{Key: "example.note", Value: strings.Repeat("è", 150)}}
	line = mustLine(t, l)
	doc = decode(t, line)
	if len(line) > max {
		t.Errorf("the line is %d bytes", len(line))
	}
	for _, k := range []string{"target", "text"} {
		if s := doc.String(k); !strings.HasSuffix(s, "...") {
			t.Errorf("%s was not halved: %d characters", k, utf8.RuneCountInString(s))
		}
	}
	if doc.String("session") != raceSession || doc.String("id") != l.ID || doc.String("at") != l.At ||
		doc.String("workspace") != raceWorkspace || doc.String("branch") != branch {
		t.Errorf("a field that names the record, or one short enough, was cut: %s", line)
	}
	t.Logf("cut to %d bytes: target %d, text %d characters", len(line), utf8.RuneCountInString(doc.String("target")),
		utf8.RuneCountInString(doc.String("text")))

	// A label's value is cut too when it is the longest.
	short := "short"
	l.Target, l.Text = &short, &short
	l.Labels = schema.Object{{Key: "example.note", Value: strings.Repeat("è", 400)}}
	doc = decode(t, mustLine(t, l))
	labels, _ := doc.Get("labels")
	if v, _ := labels.(schema.Object).Get("example.note"); !strings.HasSuffix(v.(string), "...") {
		t.Errorf("the label's value was not cut")
	}

	// What cannot be cut cannot fit: many labels, each name long and each value short.
	l.Labels = schema.Object{}
	for i := 0; i < 80; i++ {
		l.Labels = append(l.Labels, schema.Member{Key: fmt.Sprintf("example.label_%02d_%s", i, strings.Repeat("n", 20)), Value: "v"})
	}
	if _, err := Line(l); err == nil || !strings.Contains(err.Error(), "does not fit in 2048 bytes") {
		t.Errorf("a record that cannot fit: %v", err)
	}

	// A record the schema refuses is refused, not cut.
	bad := New("event", Common{Workspace: "not-an-id"})
	if _, err := Line(bad); err == nil {
		t.Errorf("a record with a bad workspace id was written")
	}
}

func mustLine(t *testing.T, l *format.Log) []byte {
	t.Helper()
	line, err := Line(l)
	if err != nil {
		t.Fatal(err)
	}
	return line
}

// lineLen counts a text's bytes on the line as the encoder writes them.
func TestLineLen(t *testing.T) {
	for _, s := range []string{"", "plain", "a\"b\\c", "tab\tnew\nline\r", "\x01\x7f", "\u00e9t\u00e9", "\U0001F600", "\xff"} {
		line, err := schema.EncodeLine(s)
		if err != nil {
			t.Fatal(err)
		}
		if want := len(line) - 3; lineLen(s) != want { // less the two quotes and the line feed
			t.Errorf("lineLen(%q) = %d, the encoder writes %d", s, lineLen(s), want)
		}
	}
}

func TestCut(t *testing.T) {
	for _, c := range []struct {
		in   string
		max  int
		want string
	}{{"abcdef", 6, "abcdef"}, {"abcdef", 5, "ab..."}, {"abcdef", 3, "abc"}, {"ééééé", 4, "é..."}} {
		if got := Cut(c.in, c.max); got != c.want {
			t.Errorf("Cut(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}
