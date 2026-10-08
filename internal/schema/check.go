package schema

// The small JSON Schema checker: the keywords the formats set's schemas use, with draft 2020-12's meaning, and
// nothing else. A schema that uses any other keyword fails CheckSchema, so a schema can never mean more than this
// checker enforces. It lived in formats/formats_test.go (plan part 0) and moved here in part 2, unchanged in what it
// checks, so the formats test and Bonsai's own code (the lock, status --json) use one checker.

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Draft202012 is the $schema every schema in the set declares.
const Draft202012 = "https://json-schema.org/draft/2020-12/schema"

// The keywords the checker implements. Annotations are skipped; any other keyword fails the schema.
var (
	annotations = map[string]bool{"title": true, "description": true, "examples": true}
	keywords    = map[string]bool{
		"type": true, "properties": true, "required": true, "additionalProperties": true, "propertyNames": true,
		"items": true, "enum": true, "const": true, "pattern": true, "maxLength": true, "minimum": true,
		"maximum": true, "maxItems": true,
	}
)

// Parse decodes a schema's bytes (formats.Schema gives them) and checks it is an object that declares draft
// 2020-12. It does not run CheckSchema: the formats test does, once, for every schema in the set.
func Parse(raw []byte) (Object, error) {
	v, err := Decode(raw)
	if err != nil {
		return nil, err
	}
	s, ok := v.(Object)
	if !ok {
		return nil, errors.New("a schema is a JSON object")
	}
	if s.String("$schema") != Draft202012 {
		return nil, fmt.Errorf("a schema declares $schema %s", Draft202012)
	}
	return s, nil
}

// CheckSchema checks a schema's own form: only known keywords, every property documented with a description and
// examples that validate against it, required naming properties in their order (all of them at the top level: a
// writer writes every field, contract §2.2). It returns one message per problem, none when the schema is sound.
func CheckSchema(s Object) []string {
	return checkSchema(s, "#", true)
}

func checkSchema(s Object, at string, top bool) []string {
	var out []string
	for _, m := range s {
		switch {
		case annotations[m.Key], keywords[m.Key]:
		case top && m.Key == "$schema":
		default:
			out = append(out, fmt.Sprintf("%s: keyword %q is not one the checker implements", at, m.Key))
		}
	}
	if p, ok := s.Get("pattern"); ok {
		if ps, ok := p.(string); !ok {
			out = append(out, at+": pattern is not a string")
		} else if _, err := regexp.Compile(ps); err != nil {
			out = append(out, fmt.Sprintf("%s: pattern %q: %v", at, ps, err))
		}
	}
	var propNames []string
	if pv, ok := s.Get("properties"); ok {
		props, ok := pv.(Object)
		if !ok {
			return append(out, at+": properties is not an object")
		}
		for _, m := range props {
			here := at + "/properties/" + m.Key
			propNames = append(propNames, m.Key)
			ps, ok := m.Value.(Object)
			if !ok {
				out = append(out, here+": not a schema object")
				continue
			}
			if d, _ := ps.Get("description"); d == nil || d == "" {
				out = append(out, here+": no description")
			}
			ev, _ := ps.Get("examples")
			if ex, ok := ev.([]any); !ok || len(ex) == 0 {
				out = append(out, here+": no examples")
			} else {
				for i, e := range ex {
					for _, msg := range validate(ps, e, fmt.Sprintf("examples[%d]", i)) {
						out = append(out, here+": its own example "+msg)
					}
				}
			}
			out = append(out, checkSchema(ps, here, false)...)
		}
	}
	if rv, ok := s.Get("required"); ok {
		req, _ := rv.([]any)
		var names []string
		for _, r := range req {
			name, _ := r.(string)
			names = append(names, name)
		}
		if top && strings.Join(names, ",") != strings.Join(propNames, ",") {
			out = append(out, at+": a writer writes every field: required must list every property, in order")
		}
		j := 0
		for _, n := range names {
			for j < len(propNames) && propNames[j] != n {
				j++
			}
			if j == len(propNames) {
				out = append(out, fmt.Sprintf("%s: required names %q out of the properties' order, or not a property", at, n))
				break
			}
		}
	} else if top {
		out = append(out, at+": no required list")
	}
	for _, k := range []string{"items", "additionalProperties", "propertyNames"} {
		if v, ok := s.Get(k); ok {
			if sub, ok := v.(Object); ok {
				out = append(out, checkSchema(sub, at+"/"+k, false)...)
			} else if _, ok := v.(bool); !ok || k != "additionalProperties" {
				out = append(out, fmt.Sprintf("%s/%s: not a schema object", at, k))
			}
		}
	}
	return out
}

// Validate checks an instance against a schema, for the keywords listed above, with draft 2020-12's meaning: each
// keyword applies only to instances of its own type. It returns one message per problem, each starting with the
// instance's JSON pointer ("#/packs/0/commit: ..."), none when the instance is valid.
func Validate(s Object, v any) []string {
	return validate(s, v, "#")
}

func validate(s Object, v any, at string) []string {
	var out []string
	if tv, ok := s.Get("type"); ok {
		var types []string
		switch x := tv.(type) {
		case string:
			types = []string{x}
		case []any:
			for _, e := range x {
				name, _ := e.(string)
				types = append(types, name)
			}
		}
		match := false
		for _, ty := range types {
			if hasType(v, ty) {
				match = true
			}
		}
		if !match {
			return append(out, fmt.Sprintf("%s: is %s, want %s", at, typeOf(v), strings.Join(types, " or ")))
		}
	}
	if c, ok := s.Get("const"); ok && !Equal(c, v) {
		out = append(out, fmt.Sprintf("%s: is %s, want %s", at, Show(v), Show(c)))
	}
	if ev, ok := s.Get("enum"); ok {
		in := false
		list, _ := ev.([]any)
		for _, e := range list {
			if Equal(e, v) {
				in = true
			}
		}
		if !in {
			out = append(out, fmt.Sprintf("%s: %s is not one of %s", at, Show(v), Show(ev)))
		}
	}
	switch x := v.(type) {
	case string:
		if p, ok := s.Get("pattern"); ok {
			ps, _ := p.(string)
			if re, err := regexp.Compile(ps); err != nil || !re.MatchString(x) {
				out = append(out, fmt.Sprintf("%s: %s does not match %s", at, Show(x), ps))
			}
		}
		if n, ok := num(s, "maxLength"); ok && big.NewRat(int64(utf8.RuneCountInString(x)), 1).Cmp(n) > 0 {
			out = append(out, fmt.Sprintf("%s: longer than %s characters", at, n.RatString()))
		}
	case json.Number:
		r, _ := new(big.Rat).SetString(string(x))
		if n, ok := num(s, "minimum"); ok && r.Cmp(n) < 0 {
			out = append(out, fmt.Sprintf("%s: %s is below %s", at, x, n.RatString()))
		}
		if n, ok := num(s, "maximum"); ok && r.Cmp(n) > 0 {
			out = append(out, fmt.Sprintf("%s: %s is above %s", at, x, n.RatString()))
		}
	case []any:
		if n, ok := num(s, "maxItems"); ok && big.NewRat(int64(len(x)), 1).Cmp(n) > 0 {
			out = append(out, fmt.Sprintf("%s: more than %s items", at, n.RatString()))
		}
		if iv, ok := s.Get("items"); ok {
			is, _ := iv.(Object)
			for i, e := range x {
				out = append(out, validate(is, e, fmt.Sprintf("%s/%d", at, i))...)
			}
		}
	case Object:
		if rv, ok := s.Get("required"); ok {
			req, _ := rv.([]any)
			for _, r := range req {
				name, _ := r.(string)
				if _, ok := x.Get(name); !ok {
					out = append(out, fmt.Sprintf("%s: required field %q is missing", at, name))
				}
			}
		}
		props, _ := s.Get("properties")
		po, _ := props.(Object)
		addl, hasAddl := s.Get("additionalProperties")
		names, hasNames := s.Get("propertyNames")
		for _, m := range x {
			here := at + "/" + m.Key
			if hasNames {
				ns, _ := names.(Object)
				out = append(out, validate(ns, m.Key, here+" (its name)")...)
			}
			if ps, ok := po.Get(m.Key); ok {
				pso, _ := ps.(Object)
				out = append(out, validate(pso, m.Value, here)...)
				continue
			}
			if hasAddl {
				switch a := addl.(type) {
				case bool:
					if !a {
						out = append(out, here+": not allowed")
					}
				case Object:
					out = append(out, validate(a, m.Value, here)...)
				}
			}
		}
	}
	return out
}

// CheckOrder checks that an instance's fields follow its schema's properties order (contract §2.2: a writer writes
// every field in a fixed order), at every depth the schema describes. Unknown fields may sit anywhere.
func CheckOrder(s Object, v any) []string {
	return checkOrder(s, v, "#")
}

func checkOrder(s Object, v any, at string) []string {
	var out []string
	switch x := v.(type) {
	case []any:
		if iv, ok := s.Get("items"); ok {
			is, _ := iv.(Object)
			for i, e := range x {
				out = append(out, checkOrder(is, e, fmt.Sprintf("%s/%d", at, i))...)
			}
		}
	case Object:
		props, _ := s.Get("properties")
		po, _ := props.(Object)
		addl, _ := s.Get("additionalProperties")
		ao, _ := addl.(Object)
		last := -1
		for _, m := range x {
			idx := po.Index(m.Key)
			if idx >= 0 {
				if idx < last {
					out = append(out, fmt.Sprintf("%s/%s: out of the schema's field order", at, m.Key))
				}
				last = idx
				ps, _ := po[idx].Value.(Object)
				out = append(out, checkOrder(ps, m.Value, at+"/"+m.Key)...)
			} else if ao != nil {
				out = append(out, checkOrder(ao, m.Value, at+"/"+m.Key)...)
			}
		}
	}
	return out
}

func hasType(v any, ty string) bool {
	switch ty {
	case "null":
		return v == nil
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := v.(json.Number)
		return ok
	case "integer":
		n, ok := v.(json.Number)
		if !ok {
			return false
		}
		r, ok := new(big.Rat).SetString(string(n))
		return ok && r.IsInt()
	case "array":
		_, ok := v.([]any)
		return ok
	case "object":
		_, ok := v.(Object)
		return ok
	}
	return false
}

func typeOf(v any) string {
	for _, ty := range []string{"null", "boolean", "string", "integer", "number", "array", "object"} {
		if hasType(v, ty) {
			return ty
		}
	}
	return "unknown"
}

func num(s Object, key string) (*big.Rat, bool) {
	v, ok := s.Get(key)
	if !ok {
		return nil, false
	}
	n, ok := v.(json.Number)
	if !ok {
		return nil, false
	}
	return new(big.Rat).SetString(string(n))
}

// Equal compares two JSON values: objects by their fields whatever the order, numbers by value.
func Equal(a, b any) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case string:
		y, ok := b.(string)
		return ok && x == y
	case json.Number:
		y, ok := b.(json.Number)
		if !ok {
			return false
		}
		rx, okx := new(big.Rat).SetString(string(x))
		ry, oky := new(big.Rat).SetString(string(y))
		return okx && oky && rx.Cmp(ry) == 0
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !Equal(x[i], y[i]) {
				return false
			}
		}
		return true
	case Object:
		y, ok := b.(Object)
		if !ok || len(x) != len(y) {
			return false
		}
		for _, m := range x {
			w, ok := y.Get(m.Key)
			if !ok || !Equal(m.Value, w) {
				return false
			}
		}
		return true
	}
	return false
}

// Show prints a value as one line of JSON, in ASCII, for messages.
func Show(v any) string {
	b, err := Encode(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	// One line: Encode indents, so fold its lines back together.
	var out strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
		out.WriteString(strings.TrimLeft(line, " "))
	}
	return out.String()
}
