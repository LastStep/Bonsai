package textui

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
)

// sample is a Doc with every kind of block, long enough to wrap at 80 columns.
func sample() *Doc {
	var d Doc
	d.Title("bonsai demo", "a short note")
	d.Section("Packs", R(T("base", Key), C("0.1.0"), T("new, at 6023a34", New), T("https://example.org/packs/base.git", Path)),
		R(T("workflow", Key), C("0.1.0"), T("at ef5b5ae", Plain), T("https://example.org/packs/workflow.git", Path)))
	d.Table("Asks", R(C("state"), C("key"), C("title")),
		R(T("open", Warn), T("agent:k1", Path), C("Which one of the two ways should the next part of the work take, and why?")),
		R(T("closed", Dim), T("agent:k2", Path), C("Which two?")))
	d.Para("Runs code", "nothing that needs --allow-exec (no pack hook line, file a hook runs or plugin code part is added or "+
		"changed, and no change to a hook line)", Ok)
	d.Gap()
	d.Hint("every field: bonsai demo --json")
	return &d
}

// The width fits the layout: no line is wider than the look's width, at 80 and at 120, and the text is the same
// words at both.
func TestWidthFits(t *testing.T) {
	for _, w := range []int{80, 120} {
		out := sample().Render(Look{Width: w})
		for _, l := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
			if n := width(l); n > w {
				t.Errorf("width %d: a line of %d columns: %q", w, n, l)
			}
			if strings.HasSuffix(l, " ") {
				t.Errorf("width %d: a line ends in a space: %q", w, l)
			}
		}
		if got, want := strings.Join(strings.Fields(out), " "), strings.Join(strings.Fields(sample().Render(Look{Width: 200})), " "); got != want {
			t.Errorf("width %d: the words differ from a wide layout's:\n%s\n%s", w, got, want)
		}
	}
	at80, at120 := sample().Render(Look{Width: 80}), sample().Render(Look{Width: 120})
	if strings.Count(at80, "\n") <= strings.Count(at120, "\n") {
		t.Errorf("80 columns take no more lines than 120:\n%s\n%s", at80, at120)
	}
	// The columns line up: every row of the table starts its key at the same column.
	var cols []int
	for _, l := range strings.Split(at120, "\n") {
		if i := strings.Index(l, "agent:k"); i >= 0 {
			cols = append(cols, i)
		}
	}
	if len(cols) != 2 || cols[0] != cols[1] {
		t.Errorf("the keys start at columns %v:\n%s", cols, at120)
	}
	// A narrow terminal is drawn at the narrowest width.
	if out := sample().Render(Look{Width: 10}); out != sample().Render(Look{Width: minWidth}) {
		t.Errorf("a 10-column terminal is not drawn at %d:\n%s", minWidth, out)
	}
}

// The text a pipe gets: Fixed columns, no colour, ASCII, and the same bytes on every run.
func TestPipedIsPlainAndStable(t *testing.T) {
	a, b := sample().Render(Piped()), sample().Render(Piped())
	if a != b || a != sample().String() {
		t.Fatalf("two renders differ:\n%s\n%s", a, b)
	}
	for _, r := range a {
		if r > 0x7e || (r < 0x20 && r != '\n') {
			t.Fatalf("a character %q that is not printable ASCII:\n%s", r, a)
		}
	}
	want := "bonsai demo  a short note\n" +
		"Packs      base      0.1.0  new, at 6023a34  https://example.org/packs/base.git\n" +
		"           workflow  0.1.0  at ef5b5ae       https://example.org/packs/workflow.git\n" +
		"Asks       state   key       title\n" +
		"           open    agent:k1  Which one of the two ways should the next part of the work take, and\n" +
		"                             why?\n" +
		"           closed  agent:k2  Which two?\n" +
		"Runs code  nothing that needs --allow-exec (no pack hook line, file a hook runs or plugin code part\n" +
		"           is added or changed, and no change to a hook line)\n" +
		"\n" +
		"  every field: bonsai demo --json\n"
	if a != want {
		t.Errorf("piped:\n%s\nwant:\n%s", a, want)
	}
}

// Colour: codes round the text of each cell with a tone, none round plain text or spaces, and none at all without
// colour.
func TestColour(t *testing.T) {
	out := sample().Render(Look{Width: 100, Color: true})
	for _, want := range []string{"\x1b[1mbonsai demo\x1b[0m", "\x1b[2ma short note\x1b[0m", "\x1b[35mnew, at 6023a34\x1b[0m",
		"\x1b[33mopen\x1b[0m", "\x1b[36magent:k1\x1b[0m", "\x1b[32mnothing that needs", "\x1b[2mevery field: bonsai demo --json\x1b[0m"} {
		if !strings.Contains(out, want) {
			t.Errorf("no %q in\n%s", want, out)
		}
	}
	if regexp.MustCompile("\x1b\\[[0-9;]*m *\x1b\\[0m").MatchString(out) {
		t.Errorf("colour round nothing but spaces:\n%q", out)
	}
	stripped := out
	for strings.Contains(stripped, "\x1b[") {
		i := strings.Index(stripped, "\x1b[")
		j := strings.IndexByte(stripped[i:], 'm')
		stripped = stripped[:i] + stripped[i+j+1:]
	}
	if stripped != sample().Render(Look{Width: 100}) {
		t.Errorf("the coloured text without its codes is not the plain text:\n%s", stripped)
	}
}

// For: a buffer and a pipe get the piped look; a terminal its width and colour, unless NO_COLOR is set (to anything
// but empty), and Fixed when it does not say its width.
func TestFor(t *testing.T) {
	if l := For(&bytes.Buffer{}); l != Piped() {
		t.Errorf("a buffer: %+v", l)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close(); _ = w.Close() }()
	if l := For(w); l != Piped() {
		t.Errorf("a pipe: %+v", l)
	}
	saved, savedColor := terminal, colorOn
	defer func() { terminal, colorOn = saved, savedColor }()
	terminal = func(*os.File) (bool, int) { return true, 132 }
	colorOn = func(*os.File) bool { return true }
	t.Setenv("NO_COLOR", "")
	if l := For(w); l != (Look{Width: 132, Color: true}) {
		t.Errorf("a terminal: %+v", l)
	}
	t.Setenv("NO_COLOR", "1")
	if l := For(w); l != (Look{Width: 132}) {
		t.Errorf("a terminal under NO_COLOR: %+v", l)
	}
	t.Setenv("NO_COLOR", "")
	colorOn = func(*os.File) bool { return false }
	if l := For(w); l.Color {
		t.Errorf("a console that draws no colour codes: %+v", l)
	}
	terminal = func(*os.File) (bool, int) { return true, 0 }
	if l := For(w); l.Width != Fixed {
		t.Errorf("a terminal that does not say its width: %+v", l)
	}
	if l := For(nil); l != Piped() {
		t.Errorf("no writer: %+v", l)
	}
}

// The refusal block: the same shape for every refusal, what, why and next, a person's step naming them once.
func TestRefusal(t *testing.T) {
	var d Doc
	d.Refuse(Refusal{Command: "bonsai status", Code: "bad-flag", Exit: 2, Means: "bad input", Why: `status takes no "--nope"`,
		Who: "agent", Next: "run `bonsai status --help` to see its flags"})
	want := "bonsai status: refused  [bad-flag]\n" +
		"  what  exit 2, bad input\n" +
		"  why   status takes no \"--nope\"\n" +
		"  next  run `bonsai status --help` to see its flags\n"
	if got := d.String(); got != want {
		t.Errorf("refusal:\n%s\nwant:\n%s", got, want)
	}
	for _, c := range []struct{ next, want string }{
		{"run bonsai init in the project's main checkout", "  next  a person: run bonsai init in the project's main checkout\n"},
		{"a person runs it in a terminal of their own", "  next  a person runs it in a terminal of their own\n"},
	} {
		var d Doc
		d.Refuse(Refusal{Command: "bonsai ladder", Code: "not-linked", Exit: 4, Why: "not linked", Who: "person", Next: c.next})
		if got := d.String(); !strings.HasSuffix(got, c.want) || !strings.Contains(got, "  what  exit 4\n") {
			t.Errorf("a person's step:\n%s", got)
		}
	}
	// A long sentence wraps under its own column.
	var long Doc
	long.Refuse(Refusal{Command: "bonsai check", Code: "x", Exit: 3, Why: strings.Repeat("word ", 40), Next: "step"})
	for _, l := range strings.Split(strings.TrimSuffix(long.Render(Look{Width: 60}), "\n"), "\n")[2:] {
		if width(l) > 60 || !strings.HasPrefix(l, "  ") {
			t.Errorf("a refusal line: %q", l)
		}
		if strings.HasPrefix(l, "        ") && !strings.HasPrefix(l, "        word") {
			t.Errorf("a continued line not under its column: %q", l)
		}
	}
}

// Rows that span, labels wider than the gutter, the last column under the others when it has no room, and text kept
// raw.
func TestLayoutCases(t *testing.T) {
	var d Doc
	d.Section("Where", R(T("/a/rather/long/path/to/a/project/folder/that/takes/room", Path)), R(C("bonsai.yaml"), C("the settings")))
	d.Section("a-label-wider-than-the-gutter-can-hold", R(C("x"), C("y")))
	d.Raw("kept   as\n  it is")
	got := d.Render(Look{Width: 80})
	want := "Where     /a/rather/long/path/to/a/project/folder/that/takes/room\n" +
		"          bonsai.yaml  the settings\n" +
		"a-label-wider-than-the-gutter-can-hold\n" +
		"          x  y\n" +
		"kept   as\n  it is\n"
	if got != want {
		t.Errorf("layout:\n%s\nwant:\n%s", got, want)
	}
	var under Doc
	under.Section("Plugins", R(C("waiting"), C("base@a-long-marketplace-name"), C(strings.Repeat("because ", 12))))
	out := under.Render(Look{Width: 80})
	if lines := strings.Split(out, "\n"); len(lines) < 3 || !strings.HasPrefix(lines[1], strings.Repeat(" ", 14)+"because") {
		t.Errorf("the last column is not under the others:\n%s", out)
	}
	if Wrap("", 10)[0] != "" || len(Wrap("a\nb", 10)) != 2 || strings.Join(Wrap("ab abcdefghijklmnop cd", 8), "|") != "ab|abcdefghijklmnop|cd" {
		t.Errorf("wrap: %q %q %q", Wrap("", 10), Wrap("a\nb", 10), Wrap("ab abcdefghijklmnop cd", 8))
	}
	var empty Doc
	if !empty.Empty() || empty.String() != "" {
		t.Errorf("an empty Doc: %q", empty.String())
	}
}
