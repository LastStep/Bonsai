package format

// Writing, where Bonsai writes the format (contract §2.2, §2.5): every field the version knows, in the schema's
// order, the document held to the full schema before a byte is written, byte-stable.

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Document gives v as the format's document: v is a pointer to the format's Go type (encoded field by field in the
// schema's order, its Extra after) or a schema.Object already in the format's shape (status --json builds one).
// Nothing is checked here; Check and Encode check.
func (f *Format) Document(v any) (schema.Object, error) {
	if o, ok := v.(schema.Object); ok {
		return normalize(o).(schema.Object), nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("format: %s's writer takes a pointer to its Go type or a schema.Object, not a %T", f.ID(), v)
	}
	return encodeStruct(f.Schema(), rv.Elem())
}

// Check holds a document to the format as a writer writes it: every field the schema names present (contract
// §2.2), each of a type and value the schema allows, in the schema's order at every depth.
func (f *Format) Check(doc schema.Object) error {
	msgs := schema.Validate(f.Schema(), doc)
	msgs = append(msgs, schema.CheckOrder(f.Schema(), doc)...)
	if len(msgs) > 0 {
		return fmt.Errorf("the %s document does not fit its schema, so it is not written: %s", f.Versioned(), ascii(msgs[0]))
	}
	return nil
}

// Problem is one place a YAML file departs from its format as a writer writes it (WriterProblems).
type Problem struct {
	Pointer string // the field's JSON pointer, "#/deny/0/why"; "#" for the document
	Field   string // the field as a person reads it, "deny[0].why"; "(the document)" for the document
	Line    int    // the field's line in the file (the deepest mapping entry on the way); 0 for none
	Msg     string // what is wrong: `required field "why" is missing`, "is null, want string"; ASCII
}

// WriterProblems holds a YAML file read whole under format 1 (m, as internal/reader returns it) to the format as a
// writer writes it, as Check holds a document Bonsai writes: the format line, every field the schema names present
// (contract §2.2), each of a type and value the schema allows, in the schema's order at every depth. It lists every
// problem, in the order the schema check finds them; none when the file fits. Bonsai writes no pack file, so this is
// for the pack's maker's: bonsai check --pack (step 5.1.9) holds a pack's pack.yaml, labels.yaml and lanes.yaml to it.
func (f *Format) WriterProblems(m *reader.Map) []Problem {
	doc, _ := reader.JSON(m).(schema.Object)
	s := f.Schema()
	msgs := append(schema.Validate(s, doc), schema.CheckOrder(s, doc)...)
	out := make([]Problem, 0, len(msgs))
	for _, msg := range msgs {
		pointer, rest, _ := strings.Cut(msg, ": ")
		pointer = strings.TrimSuffix(pointer, " (its name)")
		segs := []string{}
		if p := strings.TrimPrefix(pointer, "#"); p != "" {
			segs = strings.Split(strings.TrimPrefix(p, "/"), "/")
		}
		out = append(out, Problem{Pointer: pointer, Field: fieldName(segs), Line: lineOf(m, segs), Msg: ascii(rest)})
	}
	return out
}

// Encode writes v (see Document) as the format's bytes once Check passes: a JSON document with two-space indent, a
// JSON-lines record as one compact line (both LF-ended, ASCII, schema.Encode's byte-stable rules). A record longer
// than the format's limit (the log's 2,048 bytes, an ask's 8,192, contract §8.1, §9.1) is refused, not cut: the
// caller shortens it. YAML and table formats have their own writers (EncodeYAML, Tasks.Encode, Sessions.Encode).
func (f *Format) Encode(v any) ([]byte, error) {
	if !f.Writes {
		return nil, fmt.Errorf("format: Bonsai does not write %s, so it has no writer", f.ID())
	}
	doc, err := f.Document(v)
	if err != nil {
		return nil, err
	}
	if err := f.Check(doc); err != nil {
		return nil, err
	}
	switch f.Shape {
	case JSONDoc, Part:
		return schema.Encode(doc)
	case JSONLine:
		line, err := schema.EncodeLine(doc)
		if err != nil {
			return nil, err
		}
		if max := maxLine[f.Name]; len(line) > max {
			return nil, fmt.Errorf("the %s record is %d bytes, over the format's %d, so it is not written: shorten its text fields first",
				f.Versioned(), len(line), max)
		}
		return line, nil
	}
	return nil, fmt.Errorf("format: %s is %s; it is written by its own writer", f.ID(), f.Shape)
}

// maxLine is the longest record of each JSON-lines format, its line feed counted (contract §8.1: a log record is at
// most 2,048 bytes; §9.1: an ask record at most 8,192).
var maxLine = map[string]int{"log": 2048, "ask": 8192}

// MaxLine is the longest record the format allows, in bytes with its line feed; 0 for a format that is not JSON
// lines.
func (f *Format) MaxLine() int { return maxLine[f.Name] }
