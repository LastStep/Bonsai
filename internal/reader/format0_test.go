package reader

// The format-0 reader's tests (contract §2.3, §2.4): every case of the formats set reaches its format0 outcome, and a
// table of edge cases beyond the set, each made up for this test, whose expected values are what the frozen yaml.mjs
// at 4a05eac gives, run by Node and written here as JSON.stringify printed them.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/schema"
)

// readCase0 reads one input file of the set under format 0, the way its extension says.
func readCase0(raw []byte, path string) Result0 {
	if strings.HasSuffix(path, ".md") {
		return ReadFormat0Markdown(raw)
	}
	return ReadFormat0YAML(raw)
}

func TestEveryCaseReachesItsFormat0Outcome(t *testing.T) {
	cases := loadExpect(t)
	reached, accepted, refused := 0, 0, 0
	for _, c := range cases {
		path := c.String("path")
		f0v, _ := c.Get("format0")
		want := f0v.(schema.Object)
		raw, err := os.ReadFile(filepath.Join(setDir, filepath.FromSlash(path)))
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		got := readCase0(raw, path)
		if got.Outcome.String() != want.String("outcome") {
			t.Errorf("%s: outcome %s, want %s (%v)", path, got.Outcome, want.String("outcome"), got.Err())
			continue
		}
		if got.Outcome == Accepted {
			wantV, _ := want.Get("value")
			gotV := JSON(got.Value)
			if !schema.Equal(gotV, wantV) || !sameOrder(gotV, wantV) || !sameNumbers(gotV, wantV) {
				t.Errorf("%s: value\n  %s\nwant\n  %s", path, schema.Show(gotV), schema.Show(wantV))
				continue
			}
			accepted++
		} else {
			if got.Refusal == nil || got.Refusal.Code != "" || got.Refusal.Next == "" {
				t.Errorf("%s: refusal %#v", path, got.Refusal)
				continue
			}
			refused++
		}
		reached++
	}
	t.Logf("format-0 outcomes reached: %d of %d (%d accepted, %d refused)", reached, len(cases), accepted, refused)
	if reached != len(cases) || len(cases) != setCases {
		t.Errorf("reached %d of %d cases; set 3 has %d", reached, len(cases), setCases)
	}
}

// sameNumbers holds two equal JSON values to the same text for every number: the frozen reader's 755 is not 755.0.
func sameNumbers(a, b any) bool {
	switch x := a.(type) {
	case json.Number:
		return x == b.(json.Number)
	case schema.Object:
		y := b.(schema.Object)
		for i := range x {
			if !sameNumbers(x[i].Value, y[i].Value) {
				return false
			}
		}
	case []any:
		y := b.([]any)
		for i := range x {
			if !sameNumbers(x[i], y[i]) {
				return false
			}
		}
	}
	return true
}

// stringify0 writes a format-0 value exactly as JSON.stringify writes the value yaml.mjs returns: compact, keys in
// the Map's order, numbers as JSON gives them, a lone surrogate (WTF-8 in the string) as \udxxx.
func stringify0(v any) string {
	var b strings.Builder
	var w func(v any)
	w = func(v any) {
		switch x := v.(type) {
		case nil:
			b.WriteString("null")
		case bool:
			b.WriteString(strconv.FormatBool(x))
		case float64:
			if j, ok := JSON(x).(json.Number); ok {
				b.WriteString(string(j))
			} else {
				b.WriteString("null")
			}
		case string:
			quoteJS(&b, x)
		case *Map:
			b.WriteByte('{')
			for i, e := range x.Entries() {
				if i > 0 {
					b.WriteByte(',')
				}
				quoteJS(&b, e.Key)
				b.WriteByte(':')
				w(e.Value)
			}
			b.WriteByte('}')
		case []any:
			b.WriteByte('[')
			for i, e := range x {
				if i > 0 {
					b.WriteByte(',')
				}
				w(e)
			}
			b.WriteByte(']')
		default:
			panic(fmt.Sprintf("stringify0: a %T", v))
		}
	}
	w(v)
	return b.String()
}

// quoteJS is JSON.stringify's QuoteJSONString on the string the Go text spells.
func quoteJS(b *strings.Builder, s string) {
	b.WriteByte('"')
	for i := 0; i < len(s); {
		if s[i] == 0xED && i+2 < len(s) && s[i+1] >= 0xA0 && s[i+1] <= 0xBF {
			fmt.Fprintf(b, `\u%04x`, rune(0xD000)|rune(s[i+1]&0x3F)<<6|rune(s[i+2]&0x3F))
			i += 3
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

type case0 struct {
	name string
	in   string
	md   bool
	want string // the value as JSON.stringify prints it, or "refused N" with the file line the refusal names
}

var cases0 = []case0{
	// The leniencies contract §2.4 names.
	{"a plain value holding ': '", "a: b: c\n", false, `{"a":"b: c"}`},
	{"a hash with no space before it", "a: x#y\n", false, `{"a":"x#y"}`},
	{"a hash after a space is a comment", "a: x #y\n", false, `{"a":"x"}`},
	{"TRUE is text, True is true", "a: TRUE\nb: True\nc: Null\nd: NULL\ne: False\n", false, `{"a":"TRUE","b":true,"c":null,"d":"NULL","e":false}`},
	{"the last of two duplicate keys, in the first's place", "a: 1\nb: 2\na: 3\n", false, `{"a":3,"b":2}`},
	{"document markers skipped", "---\na: 1\n...\nb: 2\n---\n", false, `{"a":1,"b":2}`},
	{"a list at its key's indent ends the mapping", "a:\n- x\nb: 1\n", false, `{"a":null}`},
	{"a # line inside a block scalar is an empty line", "t: |\n  a\n# c\n  b\n", false, `{"t":"a\n\nb\n"}`},

	// Keys in a JavaScript object's order.
	{"array-index keys first, ascending", "b: 1\n2: x\na: 3\n1: y\n", false, `{"1":"y","2":"x","b":1,"a":3}`},
	{"4294967295 is not an array index", "b: 1\n4294967295: x\n4294967294: y\n01: z\n", false, `{"4294967294":"y","b":1,"4294967295":"x","01":"z"}`},
	{"quoted keys", "'a b': 1\n\"c\": 2\n'': 3\n", false, `{"a b":1,"c":2,"":3}`},
	{"key characters", "a/b-c$.d_0: 1\n", false, `{"a/b-c$.d_0":1}`},
	{"no space after the colon", "a:b\n", false, "refused 1"},
	{"no space after an item's colon is text", "- a:b\n", false, `["a:b"]`},
	{"a key and a comment over a mapping", "a:   # note\n  b: 1\n", false, `{"a":{"b":1}}`},
	{"a complex key", "? a\n", false, "refused 1"},

	// __proto__ as JavaScript assigns it.
	{"__proto__ with a scalar sets nothing", "__proto__: x\na: 1\n", false, `{"a":1}`},
	{"__proto__ with a mapping sets the prototype", "__proto__:\n  a: 1\nb: 2\n", false, `{"b":2}`},
	{"__proto__ null, then __proto__ is a key", "__proto__:\n__proto__: 5\n", false, `{"__proto__":5}`},
	{"__proto__ a list keeps the setter", "__proto__: []\n__proto__: 1\n", false, `{}`},
	{"__proto__ {} then null then a key", "__proto__: {}\n__proto__:\n__proto__: 2\n", false, `{"__proto__":2}`},
	{"__proto__ through a prototype with a null prototype", "__proto__:\n  __proto__:\nx: 1\n__proto__: 3\n", false, `{"x":1,"__proto__":3}`},
	{"__proto__ through a prototype with its own __proto__", "__proto__:\n  __proto__:\n  __proto__: 1\n__proto__: 3\nx: 4\n", false, `{"__proto__":3,"x":4}`},
	{"__proto__ as a list item's first key is a key", "- __proto__: 1\n  a: 2\n", false, `[{"__proto__":1,"a":2}]`},
	{"__proto__ in a list item's later keys", "- a: 1\n  __proto__:\n  __proto__: 2\n  3: c\n", false, `[{"3":"c","a":1,"__proto__":2}]`},
	{"__proto__ set in a list item's later keys is lost", "- a: 1\n  __proto__:\n    b: 2\n", false, `[{"a":1}]`},
	{"toJSON and constructor are keys", "toJSON: 1\nconstructor: 2\nlength: 3\n", false, `{"toJSON":1,"constructor":2,"length":3}`},

	// Numbers as parseInt and parseFloat give them.
	{"leading zeros", "a: 0755\nb: 007\n", false, `{"a":755,"b":7}`},
	{"-0", "a: -0\nb: -0.0\n", false, `{"a":0,"b":0}`},
	{"decimals", "a: 1.0\nb: .5\nc: -.5\nd: 5.\ne: 1.50\n", false, `{"a":1,"b":0.5,"c":-0.5,"d":"5.","e":1.5}`},
	{"not numbers", "a: 1e3\nb: 0x1F\nc: 1_000\nd: +5\ne: 1.2.3\n", false, `{"a":"1e3","b":"0x1F","c":"1_000","d":"+5","e":"1.2.3"}`},
	{"a long integer", "a: 12345678901234567890\nb: 9007199254740993\n", false, `{"a":12345678901234567000,"b":9007199254740992}`},
	{"1e21 and 1e-7", "a: 1000000000000000000000\nb: 0.0000001\nc: 0.000001\nd: 100000000000000000000\n", false, `{"a":1e+21,"b":1e-7,"c":0.000001,"d":100000000000000000000}`},
	{"too large is Infinity, null in JSON", "a: " + strings.Repeat("9", 400) + "\nb: -" + strings.Repeat("9", 310) + ".5\n", false, `{"a":null,"b":null}`},
	{"too small is 0", "a: 0." + strings.Repeat("0", 400) + "1\n", false, `{"a":0}`},
	{"a date stays text", "d: 2026-10-09\nt: 2026-10-09 10:00\n", false, `{"d":"2026-10-09","t":"2026-10-09 10:00"}`},

	// Quoted values.
	{"single quotes", "a: 'it''s'\nb: ''''\nc: 'x' # c\n", false, `{"a":"it's","b":"'","c":"x"}`},
	{"double quotes and their five escapes", "a: \"q\\\"t\\n\\r\\t\\\\\\x\"\n", false, `{"a":"q\"t\n\r\t\\\\x"}`},
	{"a quote that does not close is text", "a: 'x\nb: \"\n", false, `{"a":"'x","b":"\""}`},
	{"text after a closing quote", "a: 'x' y\n", false, `{"a":"'x' y"}`},
	{"a hash inside quotes", "a: '#x' # y\nb: \"a # b\"\n", false, `{"a":"#x","b":"a # b"}`},
	{"an open quote hides a comment", "a: it's # c\n", false, `{"a":"it's # c"}`},

	// Flow sequences, [] and {}.
	{"a flow sequence", "a: [x, 'y, z', \"w\", 1, true, ~, ]\n", false, `{"a":["x","y, z","w",1,true,null,null]}`},
	{"empty flow forms", "a: []\nb: {}\nc: [ ]\n", false, `{"a":[],"b":{},"c":[]}`},
	{"a ] inside an item", "a: [a]b]\n", false, `{"a":["a]b"]}`},
	{"a nested flow sequence", "a: [x, [y]]\n", false, "refused 1"},
	{"a flow sequence left open", "a: [x,\n  y]\n", false, "refused 1"},
	{"a quote left open in a flow sequence", "a: [x, 'y]\n", false, "refused 1"},
	{"a flow mapping", "a: {b: c}\n", false, "refused 1"},
	{"anchors, aliases and tags", "a: 1\nb: &x 1\n", false, "refused 2"},
	{"an alias in an item", "- *x\n", false, "refused 1"},
	{"a quoted anchor is text", "a: '&x'\n", false, `{"a":"&x"}`},

	// Block scalars.
	{"the four styles", "a: |\n  x\n   y\n\n  z\nb: |-\n  x\nc: >\n  x\n  y\n\n  z\nd: >-\n  x\n  y\n", false, `{"a":"x\n y\n\nz\n","b":"x","c":"x y\nz\n","d":"x y"}`},
	{"keep is clip", "a: |+\n  x\n\n\nb: 1\n", false, `{"a":"x\n","b":1}`},
	{"an empty block is a line end", "a: |\nb: >-\nc: 1\n", false, `{"a":"\n","b":"","c":1}`},
	{"a block indicator with a number is text", "a: |2\n", false, `{"a":"|2"}`},
	{"a block indicator with a comment", "a: | # c\n  x\n", false, `{"a":"x\n"}`},
	{"a less indented line is cut at the first line's indent", "a: |\n    abcd\n  efgh\n", false, `{"a":"abcd\ngh\n"}`},
	{"the cut splits a character above U+FFFF", "a: |\n    x\n   \U0001F600y\n", false, `{"a":"x\n\ude00y\n"}`},
	{"a block ends at a document marker", "a: |\n  x\n---\n  y\nb: 1\n", false, "refused 4"},
	{"folded leading and doubled blank lines", "a: >\n\n  x\n\n\n  y\n", false, `{"a":"\nx\n\ny\n"}`},
	{"a block scalar as a list item", "- |\n  x\n- >-\n  y\n  z\n", false, `["x\n","y z"]`},

	// Lists.
	{"a top-level list", "- a\n- b: 1\n  c: 2\n-\n  - x\n- \n", false, `["a",{"b":1,"c":2},["x"],null]`},
	{"an item's keys line up two past the dash", "-   a: 1\n    b: 2\n", false, "refused 2"},
	{"a comment after a dash is a null item", "- # c\n  x: 1\n", false, "refused 2"},
	{"a list under a key, deeper", "a:\n  - x\n  - y\nb: 1\n", false, `{"a":["x","y"],"b":1}`},
	{"a key after a list item's keys", "a:\n  - k: 1\n    l: 2\n  - 3\n", false, `{"a":[{"k":1,"l":2},3]}`},

	// Indentation and tabs.
	{"unexpected indentation", "a: 1\n  b: 2\n", false, "refused 2"},
	{"an indented top level, then less", "  a: 1\nb: 2\n", false, `{"a":1}`},
	{"a tab in the indentation", "a:\n\tb: 1\n", false, "refused 2"},
	{"a tab after spaces", "a:\n  \tb: 1\n", false, "refused 2"},
	{"a tab-indented comment is a comment", "\t# c\na: 1\n", false, `{"a":1}`},
	{"a tab line past the end still refuses", "  a: 1\nb: 2\n\tc\n", false, "refused 3"},
	{"a tab after the colon", "a:\tb\n", false, `{"a":"b"}`},

	// The text: BOM, line ends, whitespace JavaScript knows, bytes that are not UTF-8.
	{"a BOM before a YAML key", "\xEF\xBB\xBFa: 1\n", false, "refused 1"},
	{"a BOM before a comment", "\xEF\xBB\xBF# c\na: 1\n", false, `{"a":1}`},
	{"CRLF", "a: 1\r\nb: |\r\n  x\r\n  y\r\nc: [p, q]\r\n", false, `{"a":1,"b":"x\ny\n","c":["p","q"]}`},
	{"a lone CR inside a line", "a: 1\rb: 2\n", false, "refused 1"},
	{"a lone CR at a line's end", "a: 1\r\nb: x\r\r\n", false, `{"a":1,"b":"x"}`},
	{"U+2028 inside a value", "a: x\u2028y\n", false, "refused 1"},
	{"U+2028 right after the colon", "a:\u2028y\n", false, `{"a":"y"}`},
	{"no-break spaces are whitespace", "a:\u00a0b\u00a0\nc: x\u00a0#d\n", false, `{"a":"b","c":"x"}`},
	{"bytes that are not UTF-8", "a: \xe2\x82x\xff\nb: \xed\xa0\x80\nc: \xf0\x9f\x98\n", false, "{\"a\":\"\ufffdx\ufffd\",\"b\":\"\ufffd\ufffd\ufffd\",\"c\":\"\ufffd\"}"},
	{"control characters are text", "a: x\x00\x1f\x7f\n", false, "{\"a\":\"x\\u0000\\u001f\x7f\"}"},
	{"an empty file", "", false, `{}`},
	{"only comments and markers", "# c\n---\n...\n", false, `{}`},

	// Frontmatter.
	{"md frontmatter", "---\na: 1\n---\nbody: [\n", true, `{"a":1}`},
	{"md with a BOM and CRLF", "\xEF\xBB\xBF---\r\na: 1\r\nb: x\r\n---\r\nbody\r\n", true, `{"a":1,"b":"x"}`},
	{"md without frontmatter", "# Title\na: 1\n", true, `{}`},
	{"md never closed", "---\na: 1\n", true, `{}`},
	{"md opener with a trailing space", "--- \na: 1\n---\n", true, `{}`},
	{"md closer at the end of the file", "---\na: 1\n---", true, `{"a":1}`},
	{"md ---- is no closer", "---\na: 1\n----\nb: 2\n---\n", true, "refused 3"},
	{"md empty frontmatter", "---\n\n---\n", true, `{}`},
	{"md a --- line right after the opener is read as a marker", "---\n---\na: 1\n---\n", true, `{"a":1}`},
	{"md a list", "---\n- x\n---\n", true, `["x"]`},
	{"md a refusal names the file's line", "---\na: 1\nb: [c\n---\n", true, "refused 3"},
	{"md a closer after a CR", "---\na: 1\r\r\n---\n", true, `{"a":1}`},
}

func TestFormat0Cases(t *testing.T) {
	for _, c := range cases0 {
		t.Run(c.name, func(t *testing.T) {
			var r Result0
			if c.md {
				r = ReadFormat0Markdown([]byte(c.in))
			} else {
				r = ReadFormat0YAML([]byte(c.in))
			}
			got := ""
			switch r.Outcome {
			case Accepted:
				got = stringify0(r.Value)
			case Refused:
				got = fmt.Sprintf("refused %d", r.Refusal.Line)
			default:
				t.Fatalf("outcome %s", r.Outcome)
			}
			if got != c.want {
				t.Errorf("got  %s\nwant %s (%v)", got, c.want, r.Err())
			}
		})
	}
}

// A refusal reads as one ASCII line naming its line and a next step, with no reason code: format 0 has none.
func TestFormat0Refusals(t *testing.T) {
	r := ReadFormat0Markdown([]byte("---\ntitle: ok\nnote: [caf\xc3\xa9, \x01\n---\n"))
	if r.Outcome != Refused || r.Value != nil {
		t.Fatalf("outcome %s, value %v", r.Outcome, r.Value)
	}
	msg := r.Err().Error()
	for i := 0; i < len(msg); i++ {
		if msg[i] < 0x20 || msg[i] > 0x7e {
			t.Fatalf("byte %d of %q is not printable ASCII", i, msg)
		}
	}
	if !strings.HasPrefix(msg, "line 3: ") || !strings.Contains(msg, "; next: close the [") || strings.Contains(msg, "()") {
		t.Errorf("message %q", msg)
	}
	if _, ok := r.Map(); ok {
		t.Errorf("a refusal gives a map")
	}
	ok := ReadFormat0YAML([]byte("a: 1\n"))
	if m, isMap := ok.Map(); !isMap || ok.Err() != nil || m.Len() != 1 {
		t.Errorf("Map and Err do not follow an accepted read: %v", ok.Err())
	}
	if e, _ := mustMap(t, ok).Entry("a"); e.Value != float64(1) || e.Line != 1 {
		t.Errorf("entry %#v", e)
	}
	list := ReadFormat0YAML([]byte("- a\n"))
	if _, isMap := list.Map(); isMap || list.Outcome != Accepted {
		t.Errorf("a top-level list reads as a map")
	}
}

func mustMap(t *testing.T, r Result0) *Map {
	t.Helper()
	m, ok := r.Map()
	if !ok {
		t.Fatalf("not a map: %v", r.Err())
	}
	return m
}

// A key's line is the line that set its value: under format 0 the last of two.
func TestFormat0Lines(t *testing.T) {
	m := mustMap(t, ReadFormat0Markdown([]byte("---\na: 1\nb:\n  c: x\na: 2\n---\n")))
	for key, want := range map[string]int{"a": 5, "b": 3} {
		if e, _ := m.Entry(key); e.Line != want {
			t.Errorf("%s: line %d, want %d", key, e.Line, want)
		}
	}
	b, _ := m.Get("b")
	if e, _ := b.(*Map).Entry("c"); e.Line != 4 {
		t.Errorf("c: line %d, want 4", e.Line)
	}
}

// Numbers print as JavaScript prints them (Number::toString), each row's text from Node.
func TestJSNumber(t *testing.T) {
	for _, c := range []struct {
		f    float64
		want string
	}{
		{0, "0"}, {math.Copysign(0, -1), "0"}, {1, "1"}, {-1, "-1"}, {755, "755"}, {0.5, "0.5"}, {-0.5, "-0.5"},
		{1.5, "1.5"}, {0.1, "0.1"}, {123456789012345, "123456789012345"}, {1e20, "100000000000000000000"},
		{1e21, "1e+21"}, {1.5e21, "1.5e+21"}, {12345678901234567890, "12345678901234567000"}, {1e-6, "0.000001"},
		{1e-7, "1e-7"}, {1.23e-18, "1.23e-18"}, {5e-324, "5e-324"}, {math.MaxFloat64, "1.7976931348623157e+308"},
		{0.000001234, "0.000001234"}, {123.456, "123.456"}, {-1e-7, "-1e-7"}, {2.5e-7, "2.5e-7"},
	} {
		if got := jsNumber(c.f); got != c.want {
			t.Errorf("%v: %s, want %s", c.f, got, c.want)
		}
	}
	if JSON(math.Inf(1)) != nil || JSON(math.Inf(-1)) != nil {
		t.Errorf("Infinity is not null")
	}
}

// The text: bytes that are not UTF-8 become U+FFFD as Node's decoder makes them, one per maximal run.
func TestDecode0(t *testing.T) {
	for in, want := range map[string]string{
		"ab":                   "ab",
		"\xe2\x82A":            "\ufffdA",
		"\xed\xa0\x80":         "\ufffd\ufffd\ufffd",
		"\xf0\x90A":            "\ufffdA",
		"\xc0\x80":             "\ufffd\ufffd",
		"\xf4\x90\x80\x80":     "\ufffd\ufffd\ufffd\ufffd",
		"\xe0\x80\x80":         "\ufffd\ufffd\ufffd",
		"\xf0\x9f\x98":         "\ufffd",
		"x\xff\xfey":           "x\ufffd\ufffdy",
		"\xf0\x9f\x98\x80\x80": "\U0001F600\ufffd",
		"\xc3":                 "\ufffd",
	} {
		if got := decode0([]byte(in)); got != want {
			t.Errorf("%q: %q, want %q", in, got, want)
		}
	}
}
