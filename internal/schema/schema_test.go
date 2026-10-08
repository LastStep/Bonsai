package schema

// Tests of the JSON reader, the writer and the checker. The formats test (formats/formats_test.go) runs the checker
// over every schema and example of the set; these show that each keyword also refuses what it should, so a checker
// that passed everything would fail here.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustDecode(t *testing.T, s string) any {
	t.Helper()
	v, err := Decode([]byte(s))
	if err != nil {
		t.Fatalf("Decode(%q): %v", s, err)
	}
	return v
}

func TestDecodeRefuses(t *testing.T) {
	cases := map[string]string{
		"duplicate key":        `{"a": 1, "a": 2}`,
		"nested duplicate key": `{"a": {"b": 1, "b": 1}}`,
		"trailing data":        `{"a": 1} {}`,
		"bom":                  "\xEF\xBB\xBF{}",
		"invalid utf-8":        "{\"a\": \"\xff\"}",
		"not json":             `{a: 1}`,
		"truncated":            `{"a": [1, 2`,
	}
	for name, in := range cases {
		if _, err := Decode([]byte(in)); err == nil {
			t.Errorf("%s: Decode(%q) accepted it", name, in)
		}
	}
}

func TestDecodeKeepsOrderAndNumbers(t *testing.T) {
	v := mustDecode(t, `{"z": 1, "a": 12345678901234567890, "m": [1.50, null, true, "x"], "e": {}}`)
	o := v.(Object)
	if got := strings.Join(o.Keys(), ","); got != "z,a,m,e" {
		t.Errorf("keys %s, want z,a,m,e", got)
	}
	if n, _ := o.Get("a"); n != json.Number("12345678901234567890") {
		t.Errorf("a = %#v, want the exact number", n)
	}
	if e, _ := o.Get("e"); e == nil {
		t.Errorf("an empty object reads as nil, want an empty Object")
	}
}

func TestEncodeIsByteStableASCIIAndOrdered(t *testing.T) {
	v := Object{
		{"z", "caf\xc3\xa9 \U0001F600 <b>&"},
		{"a", map[string]any{"y": 1, "b": []any{}, "a": Object{}}},
		{"n", json.Number("-0.50")},
		{"ctl", "tab\tnl\ncr\rq\"bs\\ del\x7f nul\x00"},
		{"s", []string{"x"}},
		{"m", map[string]string{"k2": "v", "k1": "w"}},
	}
	want := `{
  "z": "caf\u00e9 \ud83d\ude00 <b>&",
  "a": {
    "a": {},
    "b": [],
    "y": 1
  },
  "n": -0.50,
  "ctl": "tab\tnl\ncr\rq\"bs\\ del\u007f nul\u0000",
  "s": [
    "x"
  ],
  "m": {
    "k1": "w",
    "k2": "v"
  }
}
`
	for i := 0; i < 20; i++ { // map order differs run to run; the output must not
		got, err := Encode(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("Encode:\n%s\nwant:\n%s", got, want)
		}
	}
	back := mustDecode(t, want)
	if !Equal(back.(Object)[0].Value, "caf\xc3\xa9 \U0001F600 <b>&") {
		t.Errorf("the escaped string does not read back")
	}
}

func TestEncodeRefusesWhatItDoesNotKnow(t *testing.T) {
	for _, v := range []any{3.5, json.Number("1e"), struct{}{}, map[int]string{}} {
		if _, err := Encode(v); err == nil {
			t.Errorf("Encode(%#v) gave no error", v)
		}
	}
}

// Every example of the formats set reads and writes back to its own bytes: the writer's layout is the set's.
func TestEncodeReproducesTheSetsExamples(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "formats", "examples", "*.json"))
	if err != nil || len(paths) != 10 {
		t.Fatalf("want the ten examples, got %d (%v)", len(paths), err)
	}
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Encode(mustDecode(t, string(raw)))
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		if !bytes.Equal(got, raw) {
			t.Errorf("%s: Encode(Decode(file)) differs from the file", filepath.ToSlash(p))
		}
	}
}

func TestValidateRefusesEachKeyword(t *testing.T) {
	s := mustDecode(t, `{
	  "type": "object",
	  "required": ["kind", "n", "id"],
	  "properties": {
	    "kind": {"enum": ["pack", "once"]},
	    "n": {"type": "integer", "minimum": 0, "maximum": 3},
	    "id": {"type": "string", "pattern": "^[a-z]+$", "maxLength": 4},
	    "c": {"const": "x"},
	    "list": {"type": "array", "maxItems": 1, "items": {"type": "string"}},
	    "closed": {"type": "object", "additionalProperties": false},
	    "names": {"type": "object", "propertyNames": {"pattern": "^a"}, "additionalProperties": {"type": "boolean"}},
	    "num": {"type": ["number", "null"]}
	  }
	}`).(Object)
	good := `{"kind": "pack", "n": 2, "id": "abc", "c": "x", "list": ["a"], "closed": {}, "names": {"ab": true}, "num": 1.5}`
	if msgs := Validate(s, mustDecode(t, good)); len(msgs) != 0 {
		t.Fatalf("a valid instance failed: %v", msgs)
	}
	bad := map[string]string{
		"enum":                 `{"kind": "kept", "n": 0, "id": "a"}`,
		"required":             `{"kind": "pack", "n": 0}`,
		"integer":              `{"kind": "pack", "n": 1.5, "id": "a"}`,
		"minimum":              `{"kind": "pack", "n": -1, "id": "a"}`,
		"maximum":              `{"kind": "pack", "n": 4, "id": "a"}`,
		"type":                 `{"kind": "pack", "n": "1", "id": "a"}`,
		"pattern":              `{"kind": "pack", "n": 0, "id": "A"}`,
		"maxLength":            `{"kind": "pack", "n": 0, "id": "abcde"}`,
		"const":                `{"kind": "pack", "n": 0, "id": "a", "c": "y"}`,
		"maxItems":             `{"kind": "pack", "n": 0, "id": "a", "list": ["a", "b"]}`,
		"items":                `{"kind": "pack", "n": 0, "id": "a", "list": [1]}`,
		"additionalProperties": `{"kind": "pack", "n": 0, "id": "a", "closed": {"x": 1}}`,
		"propertyNames":        `{"kind": "pack", "n": 0, "id": "a", "names": {"b": true}}`,
		"additional schema":    `{"kind": "pack", "n": 0, "id": "a", "names": {"ab": 1}}`,
		"type list":            `{"kind": "pack", "n": 0, "id": "a", "num": "1"}`,
		"top type":             `[]`,
	}
	for name, in := range bad {
		if msgs := Validate(s, mustDecode(t, in)); len(msgs) == 0 {
			t.Errorf("%s: %s validated", name, in)
		}
	}
}

func TestCheckSchemaAndCheckOrder(t *testing.T) {
	s := mustDecode(t, `{
	  "$schema": "https://json-schema.org/draft/2020-12/schema",
	  "description": "d",
	  "type": "object",
	  "required": ["a", "b"],
	  "properties": {
	    "a": {"description": "a", "examples": [1], "type": "integer"},
	    "b": {"description": "b", "examples": ["x"], "type": "string"}
	  }
	}`).(Object)
	if msgs := CheckSchema(s); len(msgs) != 0 {
		t.Fatalf("a sound schema failed: %v", msgs)
	}
	if msgs := CheckOrder(s, mustDecode(t, `{"b": "x", "a": 1}`)); len(msgs) == 0 {
		t.Errorf("fields out of order passed CheckOrder")
	}
	if msgs := CheckOrder(s, mustDecode(t, `{"a": 1, "other": 0, "b": "x"}`)); len(msgs) != 0 {
		t.Errorf("an unknown field between known ones failed CheckOrder: %v", msgs)
	}
	broken := map[string]string{
		"unknown keyword":  `{"$schema": "x", "required": [], "properties": {}, "oneOf": []}`,
		"no description":   `{"required": ["a"], "properties": {"a": {"examples": [1]}}}`,
		"no examples":      `{"required": ["a"], "properties": {"a": {"description": "a"}}}`,
		"bad example":      `{"required": ["a"], "properties": {"a": {"description": "a", "examples": ["x"], "type": "integer"}}}`,
		"required order":   `{"required": ["b", "a"], "properties": {"a": {"description": "a", "examples": [1]}, "b": {"description": "b", "examples": [1]}}}`,
		"required missing": `{"required": ["a"], "properties": {"a": {"description": "a", "examples": [1]}, "b": {"description": "b", "examples": [1]}}}`,
		"no required":      `{"properties": {}}`,
		"bad pattern":      `{"required": [], "properties": {}, "pattern": "("}`,
	}
	for name, in := range broken {
		if msgs := CheckSchema(mustDecode(t, in).(Object)); len(msgs) == 0 {
			t.Errorf("%s: CheckSchema passed %s", name, in)
		}
	}
	if _, err := Parse([]byte(`{"$schema": "http://json-schema.org/draft-07/schema#"}`)); err == nil {
		t.Errorf("Parse took a schema of another draft")
	}
	if _, err := Parse([]byte(`[]`)); err == nil {
		t.Errorf("Parse took an array")
	}
}

func TestEqual(t *testing.T) {
	a := mustDecode(t, `{"x": [1, 2.0, {"k": null}], "y": true}`)
	b := mustDecode(t, `{"y": true, "x": [1.0, 2, {"k": null}]}`)
	if !Equal(a, b) {
		t.Errorf("equal values compared unequal")
	}
	for _, c := range []string{`{"y": true}`, `{"y": false, "x": [1, 2, {"k": null}]}`, `{"y": true, "x": [1, 2, {"k": 0}]}`} {
		if Equal(a, mustDecode(t, c)) {
			t.Errorf("Equal took %s for %s", c, Show(a))
		}
	}
}

func TestShowIsCompactJSON(t *testing.T) {
	v := Object{{"a", json.Number("1")}, {"b", []any{true, nil, "x y"}}, {"c", Object{}}, {"d", []any{}}}
	if got := Show(v); got != `{"a":1,"b":[true,null,"x y"],"c":{},"d":[]}` {
		t.Errorf("Show = %s", got)
	}
	if got := Show(3.5); got != "3.5" {
		t.Errorf("Show of a value Encode does not take = %s", got)
	}
}

func TestLenientDropsRequiredAtEveryDepth(t *testing.T) {
	s := mustDecode(t, `{"type": "object", "required": ["a", "l", "m"], "properties": {
	  "a": {"type": "string"},
	  "l": {"type": "array", "items": {"type": "object", "required": ["x"], "properties": {"x": {"type": "integer"}}}},
	  "m": {"type": "object", "additionalProperties": {"type": "object", "required": ["y"]}}}}`).(Object)
	missing := mustDecode(t, `{"l": [{}], "m": {"k": {}}}`)
	if msgs := Validate(s, missing); len(msgs) != 3 {
		t.Errorf("the writer's schema found %d missing fields, want 3: %v", len(msgs), msgs)
	}
	lenient := Lenient(s)
	if msgs := Validate(lenient, missing); len(msgs) != 0 {
		t.Errorf("the lenient schema refuses missing fields: %v", msgs)
	}
	if msgs := Validate(lenient, mustDecode(t, `{"a": 1, "l": [{"x": "y"}]}`)); len(msgs) != 2 {
		t.Errorf("the lenient schema stopped checking types: %v", msgs)
	}
	if _, ok := s.Get("required"); !ok {
		t.Errorf("Lenient changed its input")
	}
}

func TestEncodeLineIsOneCompactASCIILine(t *testing.T) {
	v := Object{{"a", "\u00e9 \"q\"\n"}, {"b", []any{true, nil}}, {"c", Object{}}, {"d", Object{{"e", json.Number("2")}}}}
	got, err := EncodeLine(v)
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\"a\":\"\\u00e9 \\\"q\\\"\\n\",\"b\":[true,null],\"c\":{},\"d\":{\"e\":2}}\n"; string(got) != want {
		t.Errorf("EncodeLine = %q, want %q", got, want)
	}
	if _, err := EncodeLine(3.5); err == nil {
		t.Errorf("EncodeLine took a float64")
	}
}
