package reader

// The values a format-1 read returns, and their JSON form.

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Map is a mapping read under format 1: its keys in file order, each once (the grammar refuses a key twice), with
// its value and the line it was on.
type Map struct {
	entries []Entry
}

// Entry is one key of a Map.
type Entry struct {
	Key   string
	Value any
	Line  int
}

func (m *Map) add(key string, v any, lineNo int) {
	m.entries = append(m.entries, Entry{key, v, lineNo})
}

// Get returns a key's value and whether the mapping holds the key.
func (m *Map) Get(key string) (any, bool) {
	e, ok := m.Entry(key)
	return e.Value, ok
}

// Entry returns a key's entry (its value and line) and whether the mapping holds the key.
func (m *Map) Entry(key string) (Entry, bool) {
	if m != nil {
		for _, e := range m.entries {
			if e.Key == key {
				return e, true
			}
		}
	}
	return Entry{}, false
}

// Entries lists the keys in file order. The slice is the caller's to keep but not to change.
func (m *Map) Entries() []Entry {
	if m == nil {
		return nil
	}
	return m.entries
}

// Len is the number of keys.
func (m *Map) Len() int {
	if m == nil {
		return 0
	}
	return len(m.entries)
}

// Decimal is a plain decimal (§2.4's -?(0|[1-9][0-9]*)\.[0-9]+), kept as its exact digits: no float rounding
// happens in the reader.
type Decimal string

// Float64 gives the decimal as the nearest float64.
func (d Decimal) Float64() float64 {
	f, _ := strconv.ParseFloat(string(d), 64)
	return f
}

// JSON turns a value the reader returned into the JSON a reader returns (formats/README.md, "value"), in the shapes
// schema.Decode gives: a *Map becomes a schema.Object in key order, an int64 or a Decimal a json.Number.
func JSON(v any) any {
	switch x := v.(type) {
	case *Map:
		o := schema.Object{}
		for _, e := range x.Entries() {
			o = append(o, schema.Member{Key: e.Key, Value: JSON(e.Value)})
		}
		return o
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = JSON(e)
		}
		return out
	case int64:
		return json.Number(strconv.FormatInt(x, 10))
	case Decimal:
		return json.Number(string(x))
	case nil, bool, string:
		return x
	}
	panic(fmt.Sprintf("reader: a %T is not a value the reader returns", v))
}
