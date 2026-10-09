package format

// bonsai check --schema <format> (contract §2.2: "bonsai check --schema <format> prints a format with every allowed
// value"): the format for a person, every field with its type and allowed values, read from the embedded schema, and
// each open list's known words from its Go table. With --json the command prints the schema itself (Format.Schema).
// The text is ASCII (spec §3), wrapped at 100 columns.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

const describeWidth = 100

// Describe gives the format as text: its name and major, what it is, how Bonsai reads and writes it, then every
// field in the schema's order, nested fields under their parent (packs[].id), each with its type, its allowed values
// and its description.
func (f *Format) Describe() string {
	s := f.Schema()
	var b strings.Builder
	title := f.Versioned()
	if f.Major == 0 {
		title += " (no format line and no major: it is part of other documents)"
	}
	b.WriteString(title + "\n")
	wrap(&b, s.String("description"), "")
	b.WriteString("\n")
	wrap(&b, fmt.Sprintf("Stored as %s. Bonsai reads %s%s; %s. Its Go type: %s. Schema: formats/schemas/%s.schema.json.",
		f.Shape, majorsText(f), map[bool]string{true: " (0 is format 0: today's files, read as they are)", false: ""}[f.Read0],
		map[bool]string{true: "Bonsai writes it", false: "Bonsai does not write it"}[f.Writes], f.GoType, f.Name), "")
	wrap(&b, "A writer writes every field below, in this order (null, or [] for a list, where one does not apply). "+
		"A reader reads a missing field as null, keeps a field it does not know, and refuses a newer major as format too new.", "")
	b.WriteString("\nFields:\n")
	f.fields(&b, s, "")
	return b.String()
}

func majorsText(f *Format) string {
	if f.Major == 0 {
		return "it wherever it appears"
	}
	var ms []string
	for _, m := range f.ReadMajors() {
		ms = append(ms, strconv.Itoa(m))
	}
	word := "major "
	if len(ms) > 1 {
		word = "majors "
	}
	return word + strings.Join(ms, " and ")
}

// fields describes an object schema's properties, and below each, its own properties and items.
func (f *Format) fields(b *strings.Builder, s schema.Object, path string) {
	for _, p := range props(s) {
		ps, _ := p.Value.(schema.Object)
		name := p.Key
		if path != "" {
			name = path + "." + p.Key
		}
		f.field(b, name, ps)
	}
	if ap := sub(s, "additionalProperties"); ap != nil && path != "" {
		f.field(b, path+".<"+keyWord(s)+">", ap)
	}
}

// keyWord names an open mapping's keys in a path: <path> for a mapping keyed by project paths, else <name>.
func keyWord(s schema.Object) string {
	if pn := sub(s, "propertyNames"); pn != nil && strings.Contains(pn.String("pattern"), `\\:`) {
		return "path"
	}
	return "name"
}

func (f *Format) field(b *strings.Builder, name string, s schema.Object) {
	b.WriteString("  " + name + "\n")
	wrap(b, "type: "+typeText(s), "      ")
	if d := s.String("description"); d != "" {
		wrap(b, d, "      ")
	}
	for _, l := range f.Lists {
		if l.Field == name {
			wrap(b, wordsText(l), "      ")
		}
	}
	if ex, ok := s.Get("examples"); ok {
		if list, ok := ex.([]any); ok && len(list) > 0 {
			e := schema.Show(list[0])
			if len(e) > 160 {
				e = e[:157] + "..."
			}
			wrap(b, "e.g. "+e, "      ")
		}
	}
	child := name
	if t := types(s); contains(t, "array") {
		if items := sub(s, "items"); items != nil {
			child += "[]"
			if props(items) != nil {
				f.fields(b, items, child)
			} else if contains(types(items), "array") {
				f.field(b, child, items)
			}
		}
		return
	}
	if props(s) != nil {
		f.fields(b, s, child)
	} else if ap := sub(s, "additionalProperties"); ap != nil && props(ap) != nil {
		f.fields(b, ap, child+".<"+keyWord(s)+">")
	}
}

// wordsText gives an open list's known words from its Go table.
func wordsText(l List) string {
	if len(*l.Words) == 0 {
		return "known words (" + l.Table + ", their one home): none yet in this build; a reader shows a word it does not know as other"
	}
	var ws []string
	for _, w := range *l.Words {
		who := ""
		switch w.Who {
		case "agent":
			who = "; next step usually an agent's"
		case "person":
			who = "; next step usually a person's"
		}
		ws = append(ws, w.Word+" ("+w.Means+who+")")
	}
	return "known words (" + l.Table + ", their one home; a reader shows any other as other): " + strings.Join(ws, "; ")
}

// types lists a schema's type names.
func types(s schema.Object) []string {
	switch t := func() any { v, _ := s.Get("type"); return v }().(type) {
	case string:
		return []string{t}
	case []any:
		var out []string
		for _, x := range t {
			if n, ok := x.(string); ok {
				out = append(out, n)
			}
		}
		return out
	}
	return nil
}

var typeWords = map[string]string{"string": "text", "integer": "an integer", "number": "a number",
	"boolean": "true or false", "array": "a list", "object": "a mapping", "null": "null"}

// typeText says what a field holds: its types, its closed list or fixed value, and its bounds.
func typeText(s schema.Object) string {
	if c, ok := s.Get("const"); ok {
		return "always " + schema.Show(c)
	}
	var parts []string
	if e, ok := s.Get("enum"); ok {
		var vals []string
		hasNull := false
		for _, v := range e.([]any) {
			if v == nil {
				hasNull = true
				continue
			}
			vals = append(vals, schema.Show(v))
		}
		t := "one of " + strings.Join(vals, ", ") + " (a closed list)"
		if hasNull {
			t += ", or null"
		}
		return t
	}
	ts := types(s)
	if len(ts) == 0 {
		return "any value (typed open)"
	}
	var words []string
	for _, t := range ts {
		w := typeWords[t]
		if p := s.String("pattern"); p != "" && t == "string" {
			w = "text matching " + p
		}
		words = append(words, w)
	}
	parts = append(parts, orList(words))
	for _, k := range []struct{ key, text string }{{"maxLength", "at most %s characters"}, {"minimum", "from %s"},
		{"maximum", "up to %s"}, {"maxItems", "at most %s items"}} {
		if v, ok := s.Get(k.key); ok {
			parts = append(parts, fmt.Sprintf(k.text, string(v.(json.Number))))
		}
	}
	if items := sub(s, "items"); items != nil {
		parts = append(parts, "each item "+typeText(items))
	}
	if pn := sub(s, "propertyNames"); pn != nil {
		parts = append(parts, "each key "+typeText(pn))
	}
	if ap := sub(s, "additionalProperties"); ap != nil && props(ap) == nil {
		parts = append(parts, "each value "+typeText(ap))
	}
	return strings.Join(parts, "; ")
}

func orList(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	}
	return strings.Join(words[:len(words)-1], ", ") + " or " + words[len(words)-1]
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// describeASCII writes a schema's prose in ASCII: its section signs and dashes as words a terminal shows, anything
// else as a Go escape.
var describeASCII = strings.NewReplacer("§", "section ", "—", "--", "–", "-", "’", "'",
	"‘", "'", "“", `"`, "”", `"`, "…", "...", "→", "->", "×", "x", "≤", "<=",
	"≥", ">=", " ", " ")

// wrap writes text in ASCII, wrapped at describeWidth, every line starting with indent.
func wrap(b *strings.Builder, text, indent string) {
	text = ascii(describeASCII.Replace(text))
	line := indent
	for _, w := range strings.Fields(text) {
		if len(line) > len(indent) && len(line)+1+len(w) > describeWidth {
			b.WriteString(line + "\n")
			line = indent
		}
		if len(line) > len(indent) {
			line += " "
		}
		line += w
	}
	if len(line) > len(indent) {
		b.WriteString(line + "\n")
	}
}
