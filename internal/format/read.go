package format

// Reading, the same for every format (contract §2.2, §2.5).

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// ReadError is a document Bonsai refuses to read: what is wrong, where, and the next step (spec §3: every refusal
// names one). Its text is ASCII.
type ReadError struct {
	Format string // the format read, "bonsai.task/1"
	Field  string // the field refused, "ladder[1].timeout_s"; "" when the document as a whole is refused
	Line   int    // the line in a YAML or markdown file (a markdown file's opening --- is line 1); 0 for none
	Code   string // internal/reader's reason code when the YAML itself was refused, else ""
	TooNew bool   // the document's format is a newer major: nothing else was read (contract §2.2)
	Msg    string // what is wrong
	Next   string // what to do
}

func (e *ReadError) Error() string {
	var b strings.Builder
	b.WriteString(e.Format)
	if e.Line > 0 {
		b.WriteString(" line " + strconv.Itoa(e.Line))
	}
	b.WriteString(": " + e.Msg)
	if e.Code != "" {
		b.WriteString(" (" + e.Code + ")")
	}
	b.WriteString("; next: " + e.Next)
	return ascii(b.String())
}

func (f *Format) refuse(line int, field, format string, args ...any) *ReadError {
	return &ReadError{Format: f.Versioned(), Field: field, Line: line, Msg: fmt.Sprintf(format, args...), Next: f.next()}
}

// tooNew refuses a document whose format line names a newer major of this format.
func (f *Format) tooNew(line int, found string) *ReadError {
	return &ReadError{Format: f.Versioned(), Line: line, TooNew: true,
		Msg:  fmt.Sprintf("format too new: %s (this Bonsai reads %s)", strconv.QuoteToASCII(found), f.Versioned()),
		Next: "read it with a Bonsai that reads " + found}
}

// checkFormat checks a document's format value: this format's name and major. A newer major is "format too new".
func (f *Format) checkFormat(v any, present bool, line int) *ReadError {
	if f.Major == 0 {
		return nil
	}
	s, isText := v.(string)
	switch {
	case !present:
		return f.refuse(line, "format", "has no format field, so it is not a %s document", f.Versioned())
	case s == f.Versioned():
		return nil
	case isText && strings.HasPrefix(s, f.ID()+"/"):
		if m, err := strconv.Atoi(s[len(f.ID())+1:]); err == nil && m > f.Major && !strings.HasPrefix(s[len(f.ID())+1:], "0") {
			return f.tooNew(line, s)
		}
	}
	return f.refuse(line, "format", "its format is %s, not %s", schema.Show(v), f.Versioned())
}

// ReadJSON reads one JSON document or JSON-lines record of the format: UTF-8 with no duplicate key (contract §2.5),
// its format checked first, then every field held to the schema as a reader holds it. It gives the document as read,
// unknown fields kept.
func (f *Format) ReadJSON(raw []byte) (schema.Object, error) {
	v, err := schema.Decode(raw)
	if err != nil {
		return nil, f.refuse(0, "", "is not a JSON document Bonsai reads: %s", oneLine(err.Error()))
	}
	doc, ok := v.(schema.Object)
	if !ok {
		return nil, f.refuse(0, "", "is not a JSON object")
	}
	fv, present := doc.Get("format")
	if err := f.checkFormat(fv, present, 0); err != nil {
		return nil, err
	}
	return f.hold(doc, nil)
}

// ReadYAML reads a YAML definition file (the format's Shape is YAMLFile) or a markdown file's frontmatter (Markdown,
// Table) under format 1. It gives the document as a reader returns it (JSON values, unknown fields kept) and the
// mapping as read, whose entries carry their lines. A file with no format: line is format 0: Format0 says so, and
// for task, run and state the caller reads it with ReadFormat0; every other format refuses it.
func (f *Format) ReadYAML(raw []byte) (doc schema.Object, m *reader.Map, err error) {
	var r reader.Result
	if f.Shape == YAMLFile {
		r = reader.ReadYAML(raw)
	} else {
		r = reader.ReadMarkdown(raw)
	}
	switch r.Outcome {
	case reader.Refused:
		e := &ReadError{Format: f.Versioned(), Line: r.Refusal.Line, Code: r.Refusal.Code, Msg: r.Refusal.Message,
			Next: r.Refusal.Next}
		if r.Refusal.Code == reader.CodeFormatTooNew {
			e.TooNew = true
		}
		return nil, nil, e
	case reader.Format0:
		where, start := "the file", "start the file with format: "+f.Versioned()
		if f.Shape != YAMLFile {
			where, start = "its frontmatter", "start the frontmatter with format: "+f.Versioned()
		}
		return nil, nil, &ReadError{Format: f.Versioned(), Line: 1, Msg: fmt.Sprintf("%s has no format: line first, so it reads "+
			"as format 0, which %s never had", where, f.ID()), Next: start, Field: errFormat0}
	}
	doc, err = f.HoldMap(r.Value)
	return doc, r.Value, err
}

// errFormat0 marks the refusal ReadYAML gives a format-0 file, which a reader of task, run and state turns into a
// format-0 read.
const errFormat0 = "\x00format0"

// isFormat0 reports whether err is ReadYAML's refusal of a format-0 file.
func isFormat0(err error) bool {
	e, ok := err.(*ReadError)
	return ok && e.Field == errFormat0
}

// HoldMap holds a mapping read under format 1 (by internal/reader) to the format: its format: line, then every
// field, with the line of the field refused.
func (f *Format) HoldMap(m *reader.Map) (schema.Object, error) {
	e, present := m.Entry("format")
	if err := f.checkFormat(e.Value, present, e.Line); err != nil {
		return nil, err
	}
	return f.hold(reader.JSON(m).(schema.Object), m)
}

// ReadFormat0 reads a markdown file of one of Bonsai's own kinds (task, run, state) under format 0, as yaml.mjs
// reads it (contract §2.3, §2.4: no new refusals): its frontmatter's mapping as read. It is never held to format 1's
// schema.
func (f *Format) ReadFormat0(raw []byte) (*reader.Map, error) {
	if !f.Read0 {
		return nil, f.refuse(1, "", "is format 0, which %s never had", f.ID())
	}
	r := reader.ReadFormat0Markdown(raw)
	if r.Outcome == reader.Refused {
		return nil, &ReadError{Format: f.ID() + " (format 0)", Line: r.Refusal.Line, Msg: r.Refusal.Message, Next: r.Refusal.Next}
	}
	m, ok := r.Map()
	if !ok {
		return nil, &ReadError{Format: f.ID() + " (format 0)", Line: 2, Msg: "its frontmatter is a list, not a mapping",
			Next: "write the frontmatter as key: value lines"}
	}
	return m, nil
}

// hold holds a document to the format's schema as a reader does (contract §2.2): a known field that is null where
// the schema allows no null reads as missing, so as null; a missing field is never refused; an unknown field is
// kept; any other field whose type or value the schema does not allow is refused, naming it. m, when the document
// came from YAML, gives the refused field's line.
func (f *Format) hold(doc schema.Object, m *reader.Map) (schema.Object, error) {
	doc = dropNull(f.Schema(), doc).(schema.Object)
	msgs := schema.Validate(f.ReadSchema(), doc)
	if len(msgs) == 0 {
		return doc, nil
	}
	pointer, rest, _ := strings.Cut(msgs[0], ": ")
	pointer = strings.TrimSuffix(pointer, " (its name)")
	segs := []string{}
	if p := strings.TrimPrefix(pointer, "#"); p != "" {
		segs = strings.Split(strings.TrimPrefix(p, "/"), "/")
	}
	field := fieldName(segs)
	return nil, f.refuse(lineOf(m, segs), field, "field %s: %s", field, rest)
}

// dropNull removes, at every depth the schema describes, a known field whose value is null where its schema allows
// no null: a reader reads it as a missing field (contract §2.2), never as a wrong type.
func dropNull(s schema.Object, v any) any {
	switch x := v.(type) {
	case schema.Object:
		ps := props(s)
		if ps == nil {
			return x
		}
		out := make(schema.Object, 0, len(x))
		for _, mem := range x {
			p := sub(ps, mem.Key)
			if p != nil && mem.Value == nil && !nullable(p) {
				continue
			}
			if p != nil {
				mem.Value = dropNull(p, mem.Value)
			}
			out = append(out, mem)
		}
		return out
	case []any:
		items := sub(s, "items")
		if items == nil {
			return x
		}
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = dropNull(items, e)
		}
		return out
	}
	return v
}

// fieldName writes a JSON pointer's segments as a person reads a field: ladder[1].timeout_s.
func fieldName(segs []string) string {
	var b strings.Builder
	for _, s := range segs {
		if _, err := strconv.Atoi(s); err == nil {
			b.WriteString("[" + s + "]")
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(s)
	}
	if b.Len() == 0 {
		return "(the document)"
	}
	return b.String()
}

// lineOf finds the line of the field a pointer names in a mapping read from YAML: the deepest mapping entry on the
// way (a list item has no line of its own, so it is its list's). 0 when m is nil.
func lineOf(m *reader.Map, segs []string) int {
	var cur any = m
	line := 0
	for _, s := range segs {
		switch x := cur.(type) {
		case *reader.Map:
			e, ok := x.Entry(s)
			if !ok {
				return line
			}
			if e.Line > 0 {
				line = e.Line
			}
			cur = e.Value
		case []any:
			i, err := strconv.Atoi(s)
			if err != nil || i < 0 || i >= len(x) {
				return line
			}
			cur = x[i]
		default:
			return line
		}
	}
	if mm, ok := cur.(*reader.Map); ok && line == 0 && mm.Len() > 0 {
		line = mm.Entries()[0].Line
	}
	return line
}

// Bind reads a document already held to the format (ReadJSON, ReadYAML) into its Go type, dst a pointer to it.
func (f *Format) Bind(doc schema.Object, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("format: Bind takes a pointer to a struct, not a %T", dst)
	}
	return decodeStruct(f.Schema(), doc, rv.Elem())
}

// bindFitting reads a format-0 document into the format's Go type as far as it fits: each known field whose value
// fits the field's schema is read, any other reads as missing (format 0 is read as it is, never refused for a
// field, contract §2.3); an unknown field is kept in Extra.
func (f *Format) bindFitting(doc schema.Object, dst any) error {
	ps := props(f.ReadSchema())
	kept := schema.Object{}
	for _, mem := range doc {
		if p := sub(ps, mem.Key); p != nil {
			if _, isConst := p.Get("const"); isConst || len(schema.Validate(p, mem.Value)) > 0 {
				continue
			}
		}
		kept = append(kept, mem)
	}
	return f.Bind(kept, dst)
}

// readJSONInto reads a JSON document of the format into dst.
func (f *Format) readJSONInto(raw []byte, dst any) error {
	doc, err := f.ReadJSON(raw)
	if err != nil {
		return err
	}
	return f.Bind(doc, dst)
}

// readYAMLInto reads a YAML file or a markdown file's frontmatter of the format into dst. For task, run and state,
// a format-0 file is read with ReadFormat0 and bound as far as it fits; its mapping as read is returned.
func (f *Format) readYAMLInto(raw []byte, dst any) (*reader.Map, error) {
	doc, _, err := f.ReadYAML(raw)
	if isFormat0(err) && f.Read0 {
		m0, err := f.ReadFormat0(raw)
		if err != nil {
			return nil, err
		}
		return m0, f.bindFitting(reader.JSON(m0).(schema.Object), dst)
	}
	if isFormat0(err) {
		err.(*ReadError).Field = ""
	}
	if err != nil {
		return nil, err
	}
	return nil, f.Bind(doc, dst)
}

// oneLine keeps an error's first line.
func oneLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}

// ascii keeps printable ASCII and writes any other character as a Go escape, so the text prints unbroken in
// PowerShell 5.1 (spec §3).
func ascii(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		c := s[i]
		if c >= 0x20 && c < 0x7f {
			b.WriteByte(c)
			i++
			continue
		}
		j := i + 1
		for j < len(s) && (s[j] < 0x20 || s[j] >= 0x7f) {
			j++
		}
		q := strconv.QuoteToASCII(s[i:j])
		b.WriteString(q[1 : len(q)-1])
		i = j
	}
	return b.String()
}

// lf gives text with every CRLF made LF, for comparing a document's bytes as Bonsai's fingerprints do.
func lf(raw []byte) []byte { return bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n")) }

func itoa(i int) string { return strconv.Itoa(i) }

func quoteASCII(s string) string { return strconv.QuoteToASCII(s) }
