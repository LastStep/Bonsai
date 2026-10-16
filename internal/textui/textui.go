// Package textui draws Bonsai's human output, the text a word prints without --json, in one look (style C, the
// director's choice of 16 Oct): a title line (the word and a short faint note), labelled sections whose rows line up
// in columns behind a label gutter, the one refusal block (what, why, next) and faint hint lines at the end.
//
// A word describes its output as a Doc, each cell marked by what it means (a Tone: ok, worth a look, refused, a path
// or a command, a hint, an addition), never by how it looks; Render chooses the look. Colour and the width come from
// a Look: For gives the one for where the text goes. On a terminal it takes the terminal's width (golang.org/x/term)
// and colour, unless NO_COLOR is set; anywhere else (a pipe, a file, a test's buffer) it is Piped: 100 columns and no
// colour, so the text is byte-stable. Without colour the text is ASCII as long as what the word gives it is.
//
// When to use it: every human answer and every refusal a word prints goes through a Doc. --json output never does.
//
//	var d textui.Doc
//	d.Title("bonsai asks", "2 open")
//	d.Table("Asks", textui.R(textui.C("state"), textui.C("key")), textui.R(textui.T("open", textui.Warn), textui.T("agent:k1", textui.Path)))
//	d.Hint("every field: bonsai asks --json")
//	fmt.Print(d.Render(textui.For(os.Stdout)))
package textui

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Tone says what a cell means; the look (a colour, or nothing) is Render's.
type Tone int

// The tones.
const (
	Plain Tone = iota // ordinary text
	Key               // a label or a heading: bold
	Ok                // a good state: green
	Warn              // worth a look: yellow
	Bad               // a refusal or an error: red
	Dim               // a hint or a secondary note: faint
	Path              // a path, a command or an id: cyan
	New               // something that is or will be added: magenta
)

// sgr is a tone's ANSI select-graphic-rendition code; "" for none.
func (t Tone) sgr() string {
	switch t {
	case Key:
		return "1"
	case Ok:
		return "32"
	case Warn:
		return "33"
	case Bad:
		return "1;31"
	case Dim:
		return "2"
	case Path:
		return "36"
	case New:
		return "35"
	}
	return ""
}

// Cell is one cell of a row: its text and what it means.
type Cell struct {
	Text string
	Tone Tone
}

// Row is one row of a section: its cells, in columns.
type Row []Cell

// C is a plain cell.
func C(text string) Cell { return Cell{Text: text} }

// T is a cell with a tone.
func T(text string, tone Tone) Cell { return Cell{Text: text, Tone: tone} }

// R is a row of cells.
func R(cells ...Cell) Row { return Row(cells) }

// Fixed is the width in columns when the output is not a terminal (or the terminal does not say its width).
const Fixed = 100

// minWidth is the narrowest layout drawn: a terminal narrower than this is drawn at this width.
const minWidth = 40

// Look is how a Doc is drawn: its width in columns and whether it is coloured.
type Look struct {
	Width int
	Color bool
}

// Piped is the look away from a terminal: Fixed columns, no colour.
func Piped() Look { return Look{Width: Fixed} }

type kind int

const (
	kTitle kind = iota
	kSection
	kTable
	kPara
	kHint
	kGap
	kRaw
	kRefusal
)

type block struct {
	kind  kind
	label string
	text  string
	tone  Tone
	rows  []Row
	ref   Refusal
}

// Doc is one word's human output, block by block in the order added.
type Doc struct {
	blocks []block
}

// Title is the first line: the command ("bonsai check") in bold and a short faint note beside it (or under it, when
// it does not fit).
func (d *Doc) Title(command, note string) {
	d.blocks = append(d.blocks, block{kind: kTitle, label: command, text: note})
}

// Section is a labelled group of rows, laid out in columns: every column but the last as wide as its widest cell,
// the last wrapped to the room left (or under the others, indented, when that room is too small). The label stands
// in the gutter beside the first row.
func (d *Doc) Section(label string, rows ...Row) {
	d.blocks = append(d.blocks, block{kind: kSection, label: label, rows: rows})
}

// Table is a Section whose first row is a header of column names, drawn faint.
func (d *Doc) Table(label string, header Row, rows ...Row) {
	d.blocks = append(d.blocks, block{kind: kTable, label: label, rows: append([]Row{header}, rows...)})
}

// Para is a labelled paragraph: text wrapped to the width beside the gutter (a \n starts a new line).
func (d *Doc) Para(label, text string, tone Tone) {
	d.blocks = append(d.blocks, block{kind: kPara, label: label, text: text, tone: tone})
}

// Hint is a faint line at the end, indented two columns: where to look next ("every field: bonsai status --json").
func (d *Doc) Hint(text string) {
	d.blocks = append(d.blocks, block{kind: kHint, text: text})
}

// Gap is a blank line between groups of sections.
func (d *Doc) Gap() { d.blocks = append(d.blocks, block{kind: kGap}) }

// Raw is text kept exactly as given (a document shown as it is stored, a schema's description): no wrapping, no
// colour. A missing final newline is added.
func (d *Doc) Raw(text string) {
	d.blocks = append(d.blocks, block{kind: kRaw, text: text})
}

// Refusal is one refusal, the same shape for every word: the command and "refused" with its error word, then what
// (its exit code), why (the sentence) and next (the step, naming a person when it is theirs).
type Refusal struct {
	Command string // "bonsai status"
	Code    string // the error object's word: "bad-flag"
	Exit    int    // the exit code
	Means   string // the exit code in a few words, as bonsai --help lists it: "bad input"
	Why     string // the refusal's sentence
	Who     string // who takes the next step: "agent" or "person" (a person's step says so)
	Next    string // the next step (a \n starts a new line)
}

// Refuse adds a refusal block.
func (d *Doc) Refuse(r Refusal) { d.blocks = append(d.blocks, block{kind: kRefusal, ref: r}) }

// Empty reports whether nothing was added.
func (d *Doc) Empty() bool { return len(d.blocks) == 0 }

// String is the Doc drawn Piped.
func (d *Doc) String() string { return d.Render(Piped()) }

// width is a text's width in columns: one per character.
func width(s string) int { return utf8.RuneCountInString(s) }

// Wrap breaks text into lines of at most w columns at spaces; a \n starts a new line. A word wider than w (a long
// path, a URL) stands whole on a line of its own, wider than w, so it can still be copied. It always gives at least
// one line.
func Wrap(text string, w int) []string {
	w = max(w, 8)
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		line := ""
		for _, word := range strings.Fields(para) {
			switch {
			case line == "":
				line = word
			case width(line)+1+width(word) <= w:
				line += " " + word
			default:
				lines = append(lines, line)
				line = word
			}
		}
		lines = append(lines, line)
	}
	return lines
}

// line builds one output line from pieces, each with its tone, so trailing spaces can be dropped and colour put
// round the text only.
type line struct {
	parts []Cell
}

func (l *line) add(text string, t Tone) { l.parts = append(l.parts, Cell{text, t}) }

func (l *line) pad(n int) {
	if n > 0 {
		l.add(strings.Repeat(" ", n), Plain)
	}
}

func (l *line) cols() int {
	n := 0
	for _, p := range l.parts {
		n += width(p.Text)
	}
	return n
}

func (l *line) write(b *strings.Builder, look Look) {
	parts := l.parts
	for len(parts) > 0 {
		last := &parts[len(parts)-1]
		trimmed := strings.TrimRight(last.Text, " ")
		if trimmed != "" {
			parts[len(parts)-1].Text = trimmed
			break
		}
		parts = parts[:len(parts)-1]
	}
	for _, p := range parts {
		code := p.Tone.sgr()
		if !look.Color || code == "" || strings.TrimSpace(p.Text) == "" {
			b.WriteString(p.Text)
			continue
		}
		b.WriteString("\x1b[" + code + "m" + p.Text + "\x1b[0m")
	}
	b.WriteString("\n")
}

// gutterFor is the label gutter: the widest section label plus two, at least 10, at most a quarter of the width (a
// wider label stands on its own line, and does not widen the gutter).
func (d *Doc) gutterFor(w int) int {
	most := max(10, w/4)
	g := 10
	for _, b := range d.blocks {
		if n := width(b.label) + 2; (b.kind == kSection || b.kind == kTable || b.kind == kPara) && n <= most {
			g = max(g, n)
		}
	}
	return g
}

// Render draws the Doc in a look.
func (d *Doc) Render(look Look) string {
	if look.Width <= 0 {
		look.Width = Fixed
	}
	look.Width = max(look.Width, minWidth)
	gutter := d.gutterFor(look.Width)
	var b strings.Builder
	for _, bl := range d.blocks {
		switch bl.kind {
		case kGap:
			b.WriteString("\n")
		case kRaw:
			b.WriteString(bl.text)
			if bl.text != "" && !strings.HasSuffix(bl.text, "\n") {
				b.WriteString("\n")
			}
		case kTitle:
			title(&b, look, bl)
		case kHint:
			for _, t := range Wrap(bl.text, look.Width-2) {
				l := &line{}
				l.add("  ", Plain)
				l.add(t, Dim)
				l.write(&b, look)
			}
		case kPara:
			para(&b, look, bl, gutter)
		case kSection, kTable:
			section(&b, look, bl, gutter)
		case kRefusal:
			refusal(&b, look, bl.ref)
		}
	}
	return b.String()
}

func title(b *strings.Builder, look Look, bl block) {
	l := &line{}
	l.add(bl.label, Key)
	if bl.text == "" {
		l.write(b, look)
		return
	}
	if width(bl.label)+2+width(bl.text) <= look.Width {
		l.add("  ", Plain)
		l.add(bl.text, Dim)
		l.write(b, look)
		return
	}
	l.write(b, look)
	for _, t := range Wrap(bl.text, look.Width-2) {
		l := &line{}
		l.add("  ", Plain)
		l.add(t, Dim)
		l.write(b, look)
	}
}

// labelLine starts a line with the gutter: the label (on the first line) padded to the gutter. A label wider than
// the gutter is written on a line of its own first.
func labelLine(b *strings.Builder, look Look, label string, gutter int) *line {
	l := &line{}
	if width(label) > gutter-2 {
		l.add(label, Key)
		l.write(b, look)
		l = &line{}
		l.pad(gutter)
		return l
	}
	l.add(label, Key)
	l.pad(gutter - width(label))
	return l
}

func para(b *strings.Builder, look Look, bl block, gutter int) {
	for i, t := range Wrap(bl.text, look.Width-gutter) {
		var l *line
		if i == 0 {
			l = labelLine(b, look, bl.label, gutter)
		} else {
			l = &line{}
			l.pad(gutter)
		}
		l.add(t, bl.tone)
		l.write(b, look)
	}
}

// section lays a section's rows out in columns. A row with fewer cells than the widest row spans: its last cell
// runs from its column to the end of the line, and does not widen that column.
func section(b *strings.Builder, look Look, bl block, gutter int) {
	ncol := 0
	for _, r := range bl.rows {
		ncol = max(ncol, len(r))
	}
	if ncol == 0 {
		labelLine(b, look, bl.label, gutter).write(b, look)
		return
	}
	widths := make([]int, ncol)
	for _, r := range bl.rows {
		n := len(r) - 1 // the cells before a row's last are aligned; the last spans unless the row is full
		if len(r) == ncol {
			n = ncol - 1
		}
		for j := 0; j < n; j++ {
			widths[j] = max(widths[j], width(r[j].Text))
		}
	}
	// colStart is where column j starts.
	colStart := func(j int) int {
		at := gutter
		for k := 0; k < j; k++ {
			if widths[k] > 0 {
				at += widths[k] + 2
			}
		}
		return at
	}
	start := colStart(ncol - 1)
	avail := look.Width - start
	maxLast, maxWord := 0, 0 // the widest cell of the last column, and its widest word, in columns
	for _, r := range bl.rows {
		if len(r) == ncol {
			maxLast = max(maxLast, width(r[ncol-1].Text))
			for _, w := range strings.Fields(r[ncol-1].Text) {
				maxWord = max(maxWord, width(w))
			}
		}
	}
	// The first columns alone do not fit: each row's cells one after another, wrapped as one text.
	stack := ncol > 1 && avail < 4
	// Too little room for the last column beside the others (under 24 columns; a word of it wider than the room; or
	// under 40 columns, with a cell of it that would take more than three lines there): it goes under them, indented,
	// in every row.
	under := !stack && ncol > 1 && maxLast > avail && (avail < 24 || maxWord > avail || (avail < 40 && maxLast > 3*avail))
	for i, r := range bl.rows {
		toneOf := func(j int) Tone { return r[j].Tone }
		if bl.kind == kTable && i == 0 {
			toneOf = func(int) Tone { return Dim }
		}
		label := ""
		if i == 0 {
			label = bl.label
		}
		l := labelLine(b, look, label, gutter)
		if len(r) == 0 {
			l.write(b, look)
			continue
		}
		if stack {
			var texts []string
			for _, c := range r {
				if c.Text != "" {
					texts = append(texts, c.Text)
				}
			}
			for k, t := range Wrap(strings.Join(texts, "  "), look.Width-gutter-2) {
				if k > 0 {
					l = &line{}
					l.pad(gutter + 2)
				}
				l.add(t, toneOf(0))
				l.write(b, look)
			}
			continue
		}
		last := len(r) - 1
		for j := 0; j < last; j++ {
			if widths[j] == 0 && r[j].Text == "" {
				continue
			}
			l.add(r[j].Text, toneOf(j))
			l.pad(widths[j] - width(r[j].Text) + 2)
		}
		text := r[last].Text
		at := colStart(last)
		if len(r) == ncol && under {
			l.write(b, look)
			if text != "" {
				for _, t := range Wrap(text, look.Width-gutter-4) {
					l := &line{}
					l.pad(gutter + 4)
					l.add(t, toneOf(last))
					l.write(b, look)
				}
			}
			continue
		}
		if text == "" {
			l.write(b, look)
			continue
		}
		for k, t := range Wrap(text, look.Width-at) {
			if k > 0 {
				l = &line{}
				l.pad(at)
			}
			l.add(t, toneOf(last))
			l.write(b, look)
		}
	}
}

// nextText is a refusal's next step as the block says it: a person's step names the person first, unless its own
// words already do.
func nextText(who, next string) string {
	if who == "person" && !strings.HasPrefix(next, "a person") {
		return "a person: " + next
	}
	return next
}

func refusal(b *strings.Builder, look Look, r Refusal) {
	l := &line{}
	l.add(r.Command, Key)
	l.add(": ", Plain)
	l.add("refused", Bad)
	if r.Code != "" {
		l.add("  ", Plain)
		l.add("["+r.Code+"]", Dim)
	}
	l.write(b, look)
	what := "exit " + strconv.Itoa(r.Exit)
	if r.Means != "" {
		what += ", " + r.Means
	}
	rows := [][3]any{
		{"what", what, Plain},
		{"why", r.Why, Plain},
		{"next", nextText(r.Who, r.Next), Path},
	}
	for _, row := range rows {
		label, text, tone := row[0].(string), row[1].(string), row[2].(Tone)
		for k, t := range Wrap(text, look.Width-8) {
			l := &line{}
			l.pad(2)
			if k == 0 {
				l.add(label, Key)
				l.pad(6 - width(label))
			} else {
				l.pad(6)
			}
			l.add(t, tone)
			l.write(b, look)
		}
	}
}
