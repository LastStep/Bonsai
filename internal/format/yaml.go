package format

// Format 1's YAML, written (contract §2.4): what Bonsai writes as YAML, bonsai.yaml (and the tables' frontmatter).
// The writer is small because format 1 is: a mapping is key: value lines, a nested mapping indented two spaces, a
// list of mappings "- key: value" items, a list of anything else one flow list [a, b] on its key's line, an empty
// list [] and an empty mapping {}. Text is written plain when format 1 reads it back as the same text, else in double
// quotes with format 1's five escapes. Every document is read back by internal/reader before it is returned, and
// must give the same value, so the writer can never write YAML its own reader reads differently.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// commentColumn is the column a line's comment starts at when the line leaves room (bonsai.yaml as init writes it).
const commentColumn = 38

// EncodeYAML writes v (see Document) as format 1's YAML once Check passes: the format line first, then every
// field in the schema's order. comment, when not nil, gives a line's comment by its field's path ("packs[1].ref";
// a list's items by their index); "" for none.
func (f *Format) EncodeYAML(v any, comment func(path string) string) ([]byte, error) {
	if !f.Writes || f.Shape != YAMLFile {
		return nil, fmt.Errorf("format: Bonsai does not write %s as YAML", f.ID())
	}
	doc, err := f.Document(v)
	if err != nil {
		return nil, err
	}
	if err := f.Check(doc); err != nil {
		return nil, err
	}
	return yamlDocument(doc, comment)
}

// yamlDocument writes a mapping as format 1's YAML and reads it back: the value must be the same.
func yamlDocument(doc schema.Object, comment func(string) string) ([]byte, error) {
	w := &yamlWriter{comment: comment}
	if err := w.mapping(doc, 0, ""); err != nil {
		return nil, err
	}
	out := []byte(w.b.String())
	r := reader.ReadYAML(out)
	if r.Outcome != reader.Accepted {
		why := "it has no format: line"
		if r.Refusal != nil {
			why = r.Refusal.Error()
		}
		return nil, fmt.Errorf("format: the YAML written does not read back (%s)", ascii(why))
	}
	if !schema.Equal(reader.JSON(r.Value), doc) {
		return nil, fmt.Errorf("format: the YAML written reads back as another value")
	}
	return out, nil
}

type yamlWriter struct {
	b       strings.Builder
	comment func(string) string
}

// line writes one line, with its path's comment.
func (w *yamlWriter) line(text, path string) {
	c := ""
	if w.comment != nil {
		c = w.comment(path)
	}
	if c == "" {
		w.b.WriteString(text + "\n")
		return
	}
	pad := commentColumn - len(text)
	if pad < 2 {
		pad = 2
	}
	w.b.WriteString(text + strings.Repeat(" ", pad) + "# " + c + "\n")
}

// mapping writes o's fields at indent; first, when not "", is written in place of the first line's indent (a list
// item's "- ").
func (w *yamlWriter) mapping(o schema.Object, indent int, path string) error {
	return w.mappingAt(o, strings.Repeat(" ", indent), strings.Repeat(" ", indent), path)
}

func (w *yamlWriter) mappingAt(o schema.Object, first, rest, path string) error {
	for i, m := range o {
		lead := rest
		if i == 0 {
			lead = first
		}
		p := m.Key
		if path != "" {
			p = path + "." + m.Key
		}
		key := lead + m.Key + ":"
		switch x := m.Value.(type) {
		case schema.Object:
			if len(x) == 0 {
				w.line(key+" {}", p)
				continue
			}
			w.line(key, p)
			if err := w.mapping(x, len(rest)+2, p); err != nil {
				return err
			}
		case []any:
			if err := w.list(key, x, len(rest), p); err != nil {
				return err
			}
		default:
			s, err := yamlScalar(x, false)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			w.line(key+" "+s, p)
		}
	}
	return nil
}

// list writes a list under its key: a flow list on the key's line unless an item is a mapping or a list, then one
// "- " item per line, two spaces deeper than the key.
func (w *yamlWriter) list(key string, l []any, indent int, path string) error {
	block := false
	for _, it := range l {
		switch it.(type) {
		case schema.Object, []any:
			block = true
		}
	}
	if !block {
		s, err := flowList(l)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		w.line(key+" "+s, path)
		return nil
	}
	w.line(key, path)
	dash := strings.Repeat(" ", indent+2) + "- "
	for i, it := range l {
		p := fmt.Sprintf("%s[%d]", path, i)
		switch x := it.(type) {
		case schema.Object:
			if len(x) == 0 {
				w.line(dash+"{}", p)
				continue
			}
			if err := w.mappingAt(x, dash, strings.Repeat(" ", indent+4), p); err != nil {
				return err
			}
		case []any:
			s, err := flowList(x)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			w.line(dash+s, p)
		default:
			s, err := yamlScalar(x, false)
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			w.line(dash+s, p)
		}
	}
	return nil
}

// flowList writes a list of scalars as [a, b]; text in a flow list is always quoted.
func flowList(l []any) (string, error) {
	parts := make([]string, len(l))
	for i, it := range l {
		s, err := yamlScalar(it, true)
		if err != nil {
			return "", err
		}
		parts[i] = s
	}
	return "[" + strings.Join(parts, ", ") + "]", nil
}

// yamlScalar writes null, a boolean, a number or text as format 1 reads it back.
func yamlScalar(v any, inFlow bool) (string, error) {
	switch x := v.(type) {
	case nil:
		return "null", nil
	case bool:
		if x {
			return "true", nil
		}
		return "false", nil
	case json.Number:
		if !yamlNumber(string(x)) {
			return "", fmt.Errorf("the number %s has no form format 1 reads (an integer of at most 15 digits, or a plain decimal)", x)
		}
		return string(x), nil
	case string:
		if !inFlow && plainReadsAs(x) {
			return x, nil
		}
		return yamlQuote(x), nil
	case schema.Object:
		if len(x) == 0 {
			return "{}", nil
		}
	case []any:
		if len(x) == 0 {
			return "[]", nil
		}
	}
	return "", fmt.Errorf("a %T cannot be written inside a flow list in format 1", v)
}

// yamlNumber reports whether a JSON number is one format 1 reads as a number: an integer of at most 15 digits, or
// a plain decimal (contract §2.4).
func yamlNumber(s string) bool {
	t := strings.TrimPrefix(s, "-")
	whole, frac, hasDot := strings.Cut(t, ".")
	if whole == "" || (len(whole) > 1 && whole[0] == '0') || !digits(whole) {
		return false
	}
	if hasDot {
		return frac != "" && digits(frac)
	}
	return len(whole) <= 15
}

func digits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// plainReadsAs reports whether text written plain reads back under format 1 as the same text: it is tried with
// internal/reader, the one judge of format 1's plain values.
func plainReadsAs(s string) bool {
	if s == "" || strings.TrimSpace(s) != s || strings.ContainsAny(s, "\n\r\t\"'") {
		return false
	}
	r := reader.ReadYAML([]byte("format: x/1\nk: " + s + "\n"))
	if r.Outcome != reader.Accepted {
		return false
	}
	v, _ := r.Value.Get("k")
	return v == s
}

// yamlQuote writes text in double quotes with format 1's escapes (contract §2.4: \" \\ \n \r \t).
func yamlQuote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	return `"` + r.Replace(s) + `"`
}
