// Package schema holds Bonsai's JSON documents (contract §2.5) and the small JSON Schema checker that holds them to
// the formats set's schemas (formats/schemas, embedded by package formats).
//
// What it gives:
//   - Object, an ordered JSON object: fields keep their order, so a writer writes them in a format's fixed order
//     (contract §2.2) and a test can check that order;
//   - Decode, which reads one JSON document, refuses a duplicate key (contract §2.5) and anything after the document,
//     and keeps numbers exact (json.Number);
//   - Encode, which writes a value byte-stable: two-space indent, LF, a final newline, ASCII only (anything else is
//     escaped as \uXXXX, so PowerShell 5.1 shows it unbroken), no map-order iteration;
//   - the checker (check.go): CheckSchema, Validate, CheckOrder, Equal.
//
// Standard library only. It knows no format by name: callers pass a schema (formats.Schema gives the embedded
// bytes). Moved here from formats/formats_test.go in plan part 2, so the formats test and Bonsai's code share it.
package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"unicode/utf8"
)

// Object is a JSON object that keeps its fields' order. Arrays are []any, numbers json.Number, null nil, true and
// false bool, strings string.
type Object []Member

// Member is one field of an Object.
type Member struct {
	Key   string
	Value any
}

// Get returns the value of a field and whether the object holds it.
func (o Object) Get(key string) (any, bool) {
	if i := o.Index(key); i >= 0 {
		return o[i].Value, true
	}
	return nil, false
}

// Index returns the position of a field, or -1.
func (o Object) Index(key string) int {
	for i, m := range o {
		if m.Key == key {
			return i
		}
	}
	return -1
}

// String returns a field's value when it is a string, else "".
func (o Object) String(key string) string {
	v, _ := o.Get(key)
	s, _ := v.(string)
	return s
}

// Keys lists the fields' names in order.
func (o Object) Keys() []string {
	out := make([]string, len(o))
	for i, m := range o {
		out[i] = m.Key
	}
	return out
}

// Decode reads one JSON document. It refuses invalid UTF-8, a byte order mark, a duplicate key at any depth
// (contract §2.5: Go's decoder keeps the last, so this check is the reader's own) and anything after the document.
func Decode(raw []byte) (any, error) {
	if bytes.HasPrefix(raw, []byte{0xEF, 0xBB, 0xBF}) {
		return nil, errors.New("starts with a byte order mark: a JSON document is UTF-8 with no BOM")
	}
	if !utf8.Valid(raw) {
		return nil, errors.New("is not valid UTF-8")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, err := decodeValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("something follows the JSON document")
	}
	return v, nil
}

func decodeValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return tok, nil
	}
	switch d {
	case '{':
		o := Object{}
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return nil, err
			}
			key, _ := kt.(string)
			if o.Index(key) >= 0 {
				return nil, fmt.Errorf("duplicate key %s", strconv.QuoteToASCII(key))
			}
			v, err := decodeValue(dec)
			if err != nil {
				return nil, err
			}
			o = append(o, Member{key, v})
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return o, nil
	case '[':
		arr := []any{}
		for dec.More() {
			v, err := decodeValue(dec)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return arr, nil
	}
	return nil, fmt.Errorf("unexpected %v", d)
}

// Encode writes a value as JSON, byte-stable: objects in their own order (an Object) or with keys sorted (a
// map[string]V), two-space indent, LF line ends and a final newline, ASCII only. It takes nil, bool, string, the Go
// integer types, json.Number, Object, []any, []string, map[string]any, map[string]string and any type with an
// EncodeJSON method; anything else is an error, never a guess.
func Encode(v any) ([]byte, error) {
	var b bytes.Buffer
	if err := (encoder{&b, true}).value(v, 0); err != nil {
		return nil, err
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

// Encoder is a type that turns itself into one of the values Encode takes.
type Encoder interface {
	EncodeJSON() any
}

// encoder writes JSON: pretty (two-space indent, a space after each colon) for documents, or compact (one line, no
// spaces) for messages.
type encoder struct {
	b      *bytes.Buffer
	pretty bool
}

// newline starts a line at depth, in pretty mode only.
func (e encoder) newline(depth int) {
	if !e.pretty {
		return
	}
	e.b.WriteByte('\n')
	for i := 0; i < depth; i++ {
		e.b.WriteString("  ")
	}
}

func (e encoder) value(v any, depth int) error {
	b := e.b
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case string:
		writeString(b, x)
	case int:
		b.WriteString(strconv.Itoa(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case json.Number:
		if !numberForm(string(x)) {
			return fmt.Errorf("%q is not a JSON number", string(x))
		}
		b.WriteString(string(x))
	case Object:
		if len(x) == 0 {
			b.WriteString("{}")
			return nil
		}
		b.WriteByte('{')
		for i, m := range x {
			e.newline(depth + 1)
			writeString(b, m.Key)
			b.WriteByte(':')
			if e.pretty {
				b.WriteByte(' ')
			}
			if err := e.value(m.Value, depth+1); err != nil {
				return err
			}
			if i < len(x)-1 {
				b.WriteByte(',')
			}
		}
		e.newline(depth)
		b.WriteByte('}')
	case []any:
		if len(x) == 0 {
			b.WriteString("[]")
			return nil
		}
		b.WriteByte('[')
		for i, el := range x {
			e.newline(depth + 1)
			if err := e.value(el, depth+1); err != nil {
				return err
			}
			if i < len(x)-1 {
				b.WriteByte(',')
			}
		}
		e.newline(depth)
		b.WriteByte(']')
	case []string:
		arr := make([]any, len(x))
		for i, s := range x {
			arr[i] = s
		}
		return e.value(arr, depth)
	case map[string]any:
		return e.value(sortedObject(x), depth)
	case map[string]string:
		m := make(map[string]any, len(x))
		for k, s := range x {
			m[k] = s
		}
		return e.value(sortedObject(m), depth)
	case Encoder:
		return e.value(x.EncodeJSON(), depth)
	default:
		return fmt.Errorf("cannot encode a %T as JSON", v)
	}
	return nil
}

// sortedObject turns a map into an Object with its keys sorted bytewise, so output never depends on map order.
func sortedObject(m map[string]any) Object {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	o := make(Object, len(keys))
	for i, k := range keys {
		o[i] = Member{k, m[k]}
	}
	return o
}

// writeString writes a JSON string in ASCII: printable ASCII as itself (no HTML escaping), the usual short escapes,
// and every other character as \uXXXX (a surrogate pair above U+FFFF). Invalid UTF-8 becomes U+FFFD.
func writeString(b *bytes.Buffer, s string) {
	const hex = "0123456789abcdef"
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r >= 0x20 && r < 0x7f:
			b.WriteRune(r)
		default:
			units := []rune{r}
			if r > 0xffff {
				r -= 0x10000
				units = []rune{0xd800 + (r >> 10), 0xdc00 + (r & 0x3ff)}
			}
			for _, u := range units {
				b.WriteString(`\u`)
				b.WriteByte(hex[(u>>12)&0xf])
				b.WriteByte(hex[(u>>8)&0xf])
				b.WriteByte(hex[(u>>4)&0xf])
				b.WriteByte(hex[u&0xf])
			}
		}
	}
	b.WriteByte('"')
}

// numberForm reports whether s is a JSON number as RFC 8259 writes one.
func numberForm(s string) bool {
	i := 0
	if i < len(s) && s[i] == '-' {
		i++
	}
	switch {
	case i < len(s) && s[i] == '0':
		i++
	case i < len(s) && s[i] >= '1' && s[i] <= '9':
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	default:
		return false
	}
	if i < len(s) && s[i] == '.' {
		i++
		start := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == start {
			return false
		}
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		start := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i == start {
			return false
		}
	}
	return i == len(s)
}
