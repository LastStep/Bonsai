package reader

// The values a read returns, and their JSON form.

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Map is a mapping, each key once, with its value and the line it was on. Read under format 1 its keys are in file
// order (the grammar refuses a key twice); read under format 0 they are in a JavaScript object's order, and a key
// written twice holds its first place, its last value and that value's line (format0.go).
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
// schema.Decode gives: a *Map becomes a schema.Object in key order, an int64 or a Decimal a json.Number. A format-0
// number (a float64) becomes the json.Number JavaScript's JSON.stringify writes, and null when it is not finite, as
// JSON.stringify writes Infinity.
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
	case float64:
		if math.IsInf(x, 0) || math.IsNaN(x) {
			return nil
		}
		return json.Number(jsNumber(x))
	case nil, bool, string:
		return x
	}
	panic(fmt.Sprintf("reader: a %T is not a value the reader returns", v))
}

// jsNumber writes a finite float64 as JavaScript's Number::toString does (ECMA-262, Number::toString with radix 10),
// which is what JSON.stringify writes: the shortest digits that read back as the same number, in plain notation from
// 1e-7 up to 1e21 and as d.ddde+n outside it, -0 as 0.
func jsNumber(f float64) string {
	if f == 0 {
		return "0"
	}
	sign := ""
	if f < 0 {
		sign, f = "-", -f
	}
	// FormatFloat's shortest digits, d.ddde±x: s is the digits (k of them), and the value is 0.s times 10^n.
	e := strconv.FormatFloat(f, 'e', -1, 64)
	at := strings.IndexByte(e, 'e')
	digits := strings.Replace(e[:at], ".", "", 1)
	x, _ := strconv.Atoi(e[at+1:])
	k, n := len(digits), x+1
	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}
	exp := "e+"
	if n-1 < 0 {
		exp = "e-"
	}
	abs := n - 1
	if abs < 0 {
		abs = -abs
	}
	if k == 1 {
		return sign + digits + exp + strconv.Itoa(abs)
	}
	return sign + digits[:1] + "." + digits[1:] + exp + strconv.Itoa(abs)
}
