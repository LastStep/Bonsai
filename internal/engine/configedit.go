package engine

// bonsai.yaml's byte-preserving pack edits (T-6009's plan, step 1): the three edits bonsai pack makes, each on the raw
// bytes, with no general YAML library and no re-encoding (re-encoding with init's template would drop the person's
// own comments). Each edit changes only the bytes it means to:
//
//   - SetRef: one pack's ref value on its ref: line, searched only after the ref: key, quoted as it was written (a
//     bare value that would not read back as the same text is written in double quotes), the line's comment and every
//     other byte kept.
//   - AddPack: one entry appended after the last entry of packs, in init's form with init's comments, indented as the
//     file's own entries are (a flow packs: [] is opened into a block list).
//   - RemovePack: one entry's lines, from its first line to its last: the last content line before the next entry or
//     the next key at the entry's indentation or less, with the comments and blank lines after it that are not
//     indented deeper than the entry's dash (those belong to what follows), and the comment lines aligned with its
//     dash directly above it (they belong to it). Removing the last entry leaves packs: [].
//
// Line endings are kept as the file has them: a CRLF file stays CRLF, new lines taking the file's first line ending.
// Every result is read again with workspace.ReadConfigFull and held to exactly the intended change (verifyEdit): the
// reader's JSON of the new file must equal the old file's JSON with that one change made, so an edit that would move
// or alter any other value is refused before anything is written.

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// safeRef is the text a ref this editor writes may hold: a git tag, branch or commit's characters, nothing a YAML
// line could read otherwise (no quote, no #, no space, no backslash). bonsai.yaml's reader then holds it to its own
// rules (a ref never starts with -).
var safeRef = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/@+-]{0,199}$`) })

// ConfigEdit is one edit's result: the new bytes, and the edited line as a person reads it (the line's number in
// the old file, its text before and after with no line ending, and the comment the line kept), for the preview and
// bonsai.changes/1's config. Before is "" for an added entry, After "" for a removed one; for those two the line is
// the entry's first line.
type ConfigEdit struct {
	Raw     []byte
	Line    int
	Before  string
	After   string
	Comment string
}

// configLines splits raw after each \n, every line keeping its ending.
func configLines(raw []byte) [][]byte { return bytes.SplitAfter(raw, []byte("\n")) }

// lineEnding is the file's first line ending, "\n" when it has none.
func lineEnding(raw []byte) string {
	if i := bytes.IndexByte(raw, '\n'); i > 0 && raw[i-1] == '\r' {
		return "\r\n"
	}
	return "\n"
}

// bare is a line with its ending taken off.
func bare(l []byte) string { return strings.TrimRight(string(l), "\r\n") }

// indentOf is a line's leading spaces.
func indentOf(s string) int { return len(s) - len(strings.TrimLeft(s, " ")) }

// packOf finds the pack's entry in the read file: its index in packs and its reader mapping.
func packOf(cfg *workspace.Config, id string) (int, *reader.Map, error) {
	v, _ := cfg.Doc.Get("packs")
	items, _ := v.([]any)
	for i, it := range items {
		pm, ok := it.(*reader.Map)
		if !ok {
			continue
		}
		if got, _ := pm.Get("id"); got == id {
			return i, pm, nil
		}
	}
	return -1, nil, fmt.Errorf("bonsai.yaml names no pack %s", id)
}

// commentOf gives a line's trailing comment, from its # on, when the value before it ends; "" for none.
func commentOf(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '#' && i > 0 && s[i-1] == ' ' {
			return s[i:]
		}
	}
	return ""
}

// SetRef gives bonsai.yaml's bytes with the pack id's ref set to ref, every other byte as it was.
func SetRef(raw []byte, id, ref string) (ConfigEdit, error) {
	if !safeRef().MatchString(ref) {
		return ConfigEdit{}, fmt.Errorf("the ref %q is not a tag, branch or commit this command writes: letters, digits and . _ / @ + -, starting with a letter or digit", ref)
	}
	cfg, err := workspace.ReadConfigFull(raw)
	if err != nil {
		return ConfigEdit{}, err
	}
	i, pm, err := packOf(cfg, id)
	if err != nil {
		return ConfigEdit{}, err
	}
	e, ok := pm.Entry("ref")
	lines := configLines(raw)
	if !ok || e.Line < 1 || e.Line > len(lines) {
		return ConfigEdit{}, fmt.Errorf("bonsai.yaml's pack %s has no ref line", id)
	}
	l := lines[e.Line-1]
	text := bare(l)
	// The value starts after the ref: key on this line, never before it (a short ref such as f must not match the key).
	k := strings.Index(text, "ref:")
	if k < 0 || strings.TrimLeft(text[:k], " -") != "" {
		return ConfigEdit{}, fmt.Errorf("bonsai.yaml line %d does not start with the ref: key", e.Line)
	}
	start := k + len("ref:")
	for start < len(text) && text[start] == ' ' {
		start++
	}
	old, _ := e.Value.(string)
	var value, newValue string
	switch {
	case start < len(text) && (text[start] == '"' || text[start] == '\''):
		q := text[start]
		end := strings.IndexByte(text[start+1:], q)
		if end < 0 {
			return ConfigEdit{}, fmt.Errorf("bonsai.yaml line %d's ref has no closing quote", e.Line)
		}
		value = text[start : start+end+2]
		newValue = string(q) + ref + string(q)
	default:
		value = old
		if !strings.HasPrefix(text[start:], old) {
			return ConfigEdit{}, fmt.Errorf("bonsai.yaml line %d does not hold the ref %s as written", e.Line, old)
		}
		newValue = ref
		// A bare value that would read back as something else (true, null, a number) is written quoted.
		if probe, err := workspace.ReadConfig([]byte("format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: probe\npacks: []\nx: " + ref + "\n")); err != nil {
			newValue = `"` + ref + `"`
		} else if got, _ := probe.Doc.Get("x"); got != ref {
			newValue = `"` + ref + `"`
		}
	}
	after := text[:start] + newValue + text[start+len(value):]
	out := replaceLine(lines, e.Line, after+string(l[len(text):]))
	edit := ConfigEdit{Raw: out, Line: e.Line, Before: text, After: after, Comment: commentOf(text[start+len(value):])}
	return edit, verifyEdit(raw, out, func(packs []any) ([]any, error) {
		packs[i].(schema.Object)[packs[i].(schema.Object).Index("ref")].Value = ref
		return packs, nil
	})
}

// replaceLine gives lines joined with line n (1-based) replaced.
func replaceLine(lines [][]byte, n int, with string) []byte {
	var out bytes.Buffer
	for j, x := range lines {
		if j == n-1 {
			out.WriteString(with)
			continue
		}
		out.Write(x)
	}
	return out.Bytes()
}

// isComment says whether a line is a comment line (# after its indentation).
func isComment(s string) bool { return strings.HasPrefix(strings.TrimLeft(s, " "), "#") }

// isBlank says whether a line holds only spaces.
func isBlank(s string) bool { return strings.TrimSpace(s) == "" }

// entryBounds gives the 0-based first and last line (inclusive) of the packs entry starting at line first (1-based,
// the line of its dash), and the dash's column. See this file's comment for the rule.
func entryBounds(lines [][]byte, first int) (lo, hi, dash int, err error) {
	start := first - 1
	if start < 0 || start >= len(lines) {
		return 0, 0, 0, fmt.Errorf("bonsai.yaml has no line %d", first)
	}
	text := bare(lines[start])
	dash = indentOf(text)
	if !strings.HasPrefix(text[dash:], "- ") {
		return 0, 0, 0, fmt.Errorf("bonsai.yaml line %d does not start a packs entry with - ", first)
	}
	// The entry's last line: the last content line, or comment deeper than its dash, before a content line at its
	// dash's indentation or less (the next entry, or the next key).
	hi = start
	for j := start + 1; j < len(lines); j++ {
		s := bare(lines[j])
		if isBlank(s) {
			continue
		}
		ind := indentOf(s)
		if isComment(s) {
			if ind > dash {
				hi = j
			}
			continue
		}
		if ind <= dash {
			break
		}
		hi = j
	}
	// The comment lines aligned with its dash directly above it belong to it.
	lo = start
	for j := start - 1; j >= 0; j-- {
		s := bare(lines[j])
		if !isComment(s) || indentOf(s) != dash {
			break
		}
		lo = j
	}
	return lo, hi, dash, nil
}

// RemovePack gives bonsai.yaml's bytes with the pack id's entry taken out, every other byte as it was (and packs:
// written packs: [] when it was the last entry).
func RemovePack(raw []byte, id string) (ConfigEdit, error) {
	cfg, err := workspace.ReadConfigFull(raw)
	if err != nil {
		return ConfigEdit{}, err
	}
	i, _, err := packOf(cfg, id)
	if err != nil {
		return ConfigEdit{}, err
	}
	lines := configLines(raw)
	lo, hi, _, err := entryBounds(lines, cfg.Packs[i].Line)
	if err != nil {
		return ConfigEdit{}, err
	}
	var out bytes.Buffer
	for j, x := range lines {
		if j < lo || j > hi {
			out.Write(x)
		}
	}
	result := out.Bytes()
	if len(cfg.Packs) == 1 {
		// The last entry: packs: with nothing under it would read null, so it becomes the empty list.
		pe, _ := cfg.Doc.Entry("packs")
		kept := configLines(result)
		if pe.Line < 1 || pe.Line > len(kept) {
			return ConfigEdit{}, fmt.Errorf("bonsai.yaml has no packs line")
		}
		text := bare(kept[pe.Line-1])
		k := strings.Index(text, "packs:")
		if k < 0 {
			return ConfigEdit{}, fmt.Errorf("bonsai.yaml line %d does not hold the packs: key", pe.Line)
		}
		rest := text[k+len("packs:"):]
		pad := strings.TrimPrefix(rest, "   ") // keep the comment's column where three spaces allow it
		if len(pad) == len(rest) || !strings.HasPrefix(pad, " ") {
			pad = rest
		}
		result = replaceLine(kept, pe.Line, text[:k]+"packs: []"+pad+string(kept[pe.Line-1][len(text):]))
	}
	edit := ConfigEdit{Raw: result, Line: lo + 1, Before: bare(lines[cfg.Packs[i].Line-1])}
	return edit, verifyEdit(raw, result, func(packs []any) ([]any, error) {
		return append(packs[:i:i], packs[i+1:]...), nil
	})
}

// NewPack is the entry AddPack writes: init's fields for one pack.
type NewPack struct {
	ID, Source, Path, Ref string
}

// packEntryLines is one packs entry in init's form, with init's comments, its dash at column 2, each line with no
// line ending.
func packEntryLines(p NewPack) ([]string, error) {
	wp := format.WorkspacePack{ID: p.ID, Source: p.Source, Ref: p.Ref}
	if p.Path != "" {
		path := p.Path
		wp.Path = &path
	}
	w := &format.Workspace{ID: "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "entry", Packs: []format.WorkspacePack{wp},
		Documents: DefaultDocuments, Protected: []string{}, PersonOnly: []string{}, NeverEdit: []string{},
		LadderFloor: []int64{}, Ladder: []format.Rung{}, Ratchets: schema.Object{}, CIMarkedTests: []string{},
		Generated: format.DefaultGenerated(), NeverRead: []string{}}
	w.Documents.Extra = nil
	body, err := w.EncodeYAMLWith(format.YAMLOptions{Comment: comment(nil), Quote: quoted})
	if err != nil {
		return nil, err
	}
	var out []string
	in := false
	for _, s := range strings.Split(string(body), "\n") {
		switch {
		case strings.HasPrefix(s, "packs:"):
			in = true
		case in && s != "" && s[0] != ' ':
			return out, nil
		case in:
			out = append(out, s)
		}
	}
	return out, nil
}

// AddPack gives bonsai.yaml's bytes with one entry appended to packs, every other byte as it was.
func AddPack(raw []byte, p NewPack) (ConfigEdit, error) {
	if !safeRef().MatchString(p.Ref) {
		return ConfigEdit{}, fmt.Errorf("the ref %q is not a tag, branch or commit this command writes: letters, digits and . _ / @ + -, starting with a letter or digit", p.Ref)
	}
	cfg, err := workspace.ReadConfigFull(raw)
	if err != nil {
		return ConfigEdit{}, err
	}
	if _, _, err := packOf(cfg, p.ID); err == nil {
		return ConfigEdit{}, fmt.Errorf("bonsai.yaml already names the pack %s", p.ID)
	}
	entry, err := packEntryLines(p)
	if err != nil {
		return ConfigEdit{}, err
	}
	nl := lineEnding(raw)
	lines := configLines(raw)
	var out bytes.Buffer
	var at int // the 0-based line the entry goes after
	dash := 2
	if len(cfg.Packs) > 0 {
		last := cfg.Packs[len(cfg.Packs)-1]
		_, hi, d, err := entryBounds(lines, last.Line)
		if err != nil {
			return ConfigEdit{}, err
		}
		at, dash = hi, d
		for j := 0; j <= at; j++ {
			out.Write(lines[j])
		}
	} else {
		// packs: [] opened into a block list.
		pe, _ := cfg.Doc.Entry("packs")
		if pe.Line < 1 || pe.Line > len(lines) {
			return ConfigEdit{}, fmt.Errorf("bonsai.yaml has no packs line")
		}
		text := bare(lines[pe.Line-1])
		k := strings.Index(text, "packs:")
		v := strings.Index(text, "[]")
		if k < 0 || v < k || strings.TrimSpace(text[k+len("packs:"):v]) != "" {
			return ConfigEdit{}, fmt.Errorf("bonsai.yaml line %d's packs is not [] or a block list", pe.Line)
		}
		dash = indentOf(text) + 2
		at = pe.Line - 1
		for j := 0; j < at; j++ {
			out.Write(lines[j])
		}
		opened := text[:k+len("packs:")]
		if strings.TrimSpace(text[v+2:]) != "" {
			opened += strings.Repeat(" ", v+2-len(opened)) + text[v+2:] // the comment keeps its column
		}
		out.WriteString(opened + string(lines[at][len(text):]))
	}
	if b := out.Bytes(); len(b) > 0 && b[len(b)-1] != '\n' {
		out.WriteString(nl)
	}
	first := 0
	for n, s := range entry {
		if n == 0 {
			first = bytes.Count(out.Bytes(), []byte("\n")) + 1
		}
		switch {
		case dash > 2:
			s = strings.Repeat(" ", dash-2) + s
		case dash < 2:
			s = s[2-dash:]
		}
		out.WriteString(s + nl)
	}
	for j := at + 1; j < len(lines); j++ {
		out.Write(lines[j])
	}
	result := out.Bytes()
	edit := ConfigEdit{Raw: result, Line: first, After: strings.TrimRight(configLinesAt(result, first), "\r\n")}
	return edit, verifyEdit(raw, result, func(packs []any) ([]any, error) {
		probe := strings.Join(append([]string{"format: bonsai.workspace/1", "id: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", "name: probe", "packs:"}, entry...), "\n") + "\n"
		pc, err := workspace.ReadConfig([]byte(probe))
		if err != nil {
			return nil, fmt.Errorf("the new entry does not read: %v", err)
		}
		v, _ := pc.Doc.Get("packs")
		return append(packs, reader.JSON(v).([]any)...), nil
	})
}

// configLinesAt is line n (1-based) of raw, with its ending.
func configLinesAt(raw []byte, n int) string {
	lines := configLines(raw)
	if n < 1 || n > len(lines) {
		return ""
	}
	return string(lines[n-1])
}

// verifyEdit reads the edited bytes again and holds them to exactly the intended change: the reader's JSON of the new
// file equals the old file's with packs changed by change, and nothing else.
func verifyEdit(old, edited []byte, change func(packs []any) ([]any, error)) error {
	oc, err := workspace.ReadConfigFull(old)
	if err != nil {
		return err
	}
	nc, err := workspace.ReadConfigFull(edited)
	if err != nil {
		return fmt.Errorf("the edited bonsai.yaml does not read (nothing is written): %v", err)
	}
	want := reader.JSON(oc.Doc).(schema.Object)
	at := want.Index("packs")
	if at < 0 {
		return fmt.Errorf("bonsai.yaml has no packs")
	}
	packs, _ := want[at].Value.([]any)
	packs, err = change(append([]any{}, packs...))
	if err != nil {
		return err
	}
	if packs == nil {
		packs = []any{}
	}
	want[at].Value = packs
	if got := schema.Show(reader.JSON(nc.Doc)); got != schema.Show(want) {
		return fmt.Errorf("the edited bonsai.yaml would change more than the pack's entry (nothing is written): it reads %s, want %s", got, schema.Show(want))
	}
	return nil
}
