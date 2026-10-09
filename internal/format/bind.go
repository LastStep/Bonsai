package format

// A document into its Go type and back, walking the format's schema beside the type, so the schema's field order is
// the order written and a test can hold each type to its schema (TestTypesFitTheirSchemas).
//
// A Go type follows these rules, which decode and encode below are the only readers of:
//   - an object with properties is a struct; each property is the field tagged with its name (`json:"name"`), in the
//     schema's order; a property with a const (a format line) needs no field: decode skips it and encode writes the
//     const. The struct's field Extra (a schema.Object) keeps every field the schema does not name, in the order
//     read, and encode writes them after the known ones (contract §2.2: a reader keeps an unknown field). In bonsai.yaml's
//     documents, whose schema allows any key beside its five, Extra holds the packs' kinds.
//   - a field the schema does not require (only status --json's documents entries have such fields: formats/README.md)
//     is left out when it is null; every other field is written, null or [] where it does not apply (contract §2.2).
//   - a field the schema allows to be null is a pointer (*string, *int64, *bool, *Struct), a slice, a schema.Object
//     or an any: nil is null. A field that cannot be null is a value: a missing one reads as its zero value, and a
//     nil slice or schema.Object writes as [] or {} (contract §2.2: null or [] where it does not apply).
//   - text is string, an integer int64, a number json.Number (its exact digits), true or false bool, a list a slice,
//     an object with no properties of its own (labels, a pack's declares) a schema.Object, and a value of several
//     types or of none (a ladder's leftovers) an any, kept as read.

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/schema"
)

var (
	objectType = reflect.TypeOf(schema.Object{})
	numberType = reflect.TypeOf(json.Number(""))
)

// structInfo is a struct type's fields by their JSON name.
type structInfo struct {
	byName map[string]int
	order  []string // the JSON names in the struct's order
	extra  int      // the index of Extra, or -1
}

var infos sync.Map // reflect.Type -> *structInfo

func infoOf(t reflect.Type) *structInfo {
	if v, ok := infos.Load(t); ok {
		return v.(*structInfo)
	}
	in := &structInfo{byName: map[string]int{}, extra: -1}
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.Name == "Extra" && sf.Type == objectType {
			in.extra = i
			continue
		}
		name, _, _ := strings.Cut(sf.Tag.Get("json"), ",")
		if name == "" || name == "-" || !sf.IsExported() {
			continue
		}
		in.byName[name] = i
		in.order = append(in.order, name)
	}
	infos.Store(t, in)
	return in
}

// sub returns a property's schema, or nil.
func sub(s schema.Object, key string) schema.Object {
	v, _ := s.Get(key)
	o, _ := v.(schema.Object)
	return o
}

func props(s schema.Object) schema.Object { return sub(s, "properties") }

// nullable reports whether a schema allows null: its type names null, its enum holds null, or it has no type.
func nullable(s schema.Object) bool {
	tv, hasType := s.Get("type")
	if e, ok := s.Get("enum"); ok {
		for _, x := range e.([]any) {
			if x == nil {
				return true
			}
		}
		if !hasType {
			return false
		}
	}
	if _, ok := s.Get("const"); ok && !hasType {
		return false
	}
	switch t := tv.(type) {
	case nil:
		return !hasType
	case string:
		return t == "null"
	case []any:
		for _, x := range t {
			if x == "null" {
				return true
			}
		}
	}
	return false
}

// decode sets dst from v, a JSON value already held to s (read.go), so a type assertion here fails only when the
// Go type and the schema disagree, which TestTypesFitTheirSchemas rules out.
func decode(s schema.Object, v any, dst reflect.Value) error {
	t := dst.Type()
	switch {
	case t == objectType:
		if v == nil {
			dst.Set(reflect.Zero(t))
			return nil
		}
		o, ok := v.(schema.Object)
		if !ok {
			return mismatch(v, t)
		}
		dst.Set(reflect.ValueOf(o))
		return nil
	case t == numberType:
		if v == nil {
			dst.SetString("")
			return nil
		}
		n, ok := v.(json.Number)
		if !ok {
			return mismatch(v, t)
		}
		dst.SetString(string(n))
		return nil
	}
	switch t.Kind() {
	case reflect.Interface:
		if v == nil {
			dst.Set(reflect.Zero(t))
		} else {
			dst.Set(reflect.ValueOf(v))
		}
		return nil
	case reflect.Pointer:
		if v == nil {
			dst.Set(reflect.Zero(t))
			return nil
		}
		p := reflect.New(t.Elem())
		if err := decode(s, v, p.Elem()); err != nil {
			return err
		}
		dst.Set(p)
		return nil
	}
	if v == nil {
		dst.Set(reflect.Zero(t))
		return nil
	}
	switch t.Kind() {
	case reflect.String:
		x, ok := v.(string)
		if !ok {
			return mismatch(v, t)
		}
		dst.SetString(x)
	case reflect.Bool:
		x, ok := v.(bool)
		if !ok {
			return mismatch(v, t)
		}
		dst.SetBool(x)
	case reflect.Int64:
		n, ok := v.(json.Number)
		if !ok {
			return mismatch(v, t)
		}
		// The schema's integer is a number with no fraction (1.0 is one), so it is read through its exact value.
		r, ok := new(big.Rat).SetString(string(n))
		if !ok || !r.IsInt() || !r.Num().IsInt64() {
			return fmt.Errorf("%s is not an integer Go holds", n)
		}
		dst.SetInt(r.Num().Int64())
	case reflect.Slice:
		list, ok := v.([]any)
		if !ok {
			return mismatch(v, t)
		}
		items := sub(s, "items")
		out := reflect.MakeSlice(t, len(list), len(list))
		for i, x := range list {
			if err := decode(items, x, out.Index(i)); err != nil {
				return err
			}
		}
		dst.Set(out)
	case reflect.Struct:
		o, ok := v.(schema.Object)
		if !ok {
			return mismatch(v, t)
		}
		return decodeStruct(s, o, dst)
	default:
		return fmt.Errorf("format: a %s field cannot hold a JSON value", t)
	}
	return nil
}

func decodeStruct(s schema.Object, o schema.Object, dst reflect.Value) error {
	in := infoOf(dst.Type())
	ps := props(s)
	extra := schema.Object{}
	for _, m := range o {
		if i, ok := in.byName[m.Key]; ok && ps.Index(m.Key) >= 0 {
			if err := decode(sub(ps, m.Key), m.Value, dst.Field(i)); err != nil {
				return fmt.Errorf("%s: %w", m.Key, err)
			}
			continue
		}
		if p := sub(ps, m.Key); p != nil {
			if _, isConst := p.Get("const"); isConst {
				continue // the format line: the type writes it
			}
			return fmt.Errorf("format: %s has no field for the property %s", dst.Type(), m.Key)
		}
		extra = append(extra, m)
	}
	if in.extra >= 0 && len(extra) > 0 {
		dst.Field(in.extra).Set(reflect.ValueOf(extra))
	}
	return nil
}

func mismatch(v any, t reflect.Type) error {
	return fmt.Errorf("format: a %T does not go in a %s field", v, t)
}

// encode gives the JSON value of src, held to s: an object's fields in the schema's order, then its Extra.
func encode(s schema.Object, src reflect.Value) (any, error) {
	t := src.Type()
	switch {
	case t == objectType:
		o := src.Interface().(schema.Object)
		if o == nil {
			if nullable(s) {
				return nil, nil
			}
			return schema.Object{}, nil
		}
		return normalize(o), nil
	case t == numberType:
		n := src.String()
		if n == "" {
			n = "0"
		}
		return json.Number(n), nil
	}
	switch t.Kind() {
	case reflect.Interface:
		if src.IsNil() {
			return nil, nil
		}
		return normalize(src.Interface()), nil
	case reflect.Pointer:
		if src.IsNil() {
			return nil, nil
		}
		return encode(s, src.Elem())
	case reflect.String:
		return src.String(), nil
	case reflect.Bool:
		return src.Bool(), nil
	case reflect.Int64:
		return jsonInt(src.Int()), nil
	case reflect.Slice:
		if src.IsNil() && nullable(s) {
			return nil, nil
		}
		items := sub(s, "items")
		out := make([]any, src.Len())
		for i := range out {
			v, err := encode(items, src.Index(i))
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case reflect.Struct:
		return encodeStruct(s, src)
	}
	return nil, fmt.Errorf("format: a %s field cannot be written as JSON", t)
}

func encodeStruct(s schema.Object, src reflect.Value) (schema.Object, error) {
	in := infoOf(src.Type())
	out := schema.Object{}
	required := map[string]bool{}
	if r, ok := s.Get("required"); ok {
		for _, name := range r.([]any) {
			required[name.(string)] = true
		}
	}
	for _, p := range props(s) {
		ps, _ := p.Value.(schema.Object)
		i, ok := in.byName[p.Key]
		if !ok {
			c, isConst := ps.Get("const")
			if !isConst {
				return nil, fmt.Errorf("format: %s has no field for the property %s", src.Type(), p.Key)
			}
			out = append(out, schema.Member{Key: p.Key, Value: c})
			continue
		}
		v, err := encode(ps, src.Field(i))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Key, err)
		}
		if v == nil && !required[p.Key] {
			continue // a field the schema does not require is left out when null (status --json's documents entries)
		}
		out = append(out, schema.Member{Key: p.Key, Value: v})
	}
	if in.extra >= 0 {
		for _, m := range src.Field(in.extra).Interface().(schema.Object) {
			if out.Index(m.Key) >= 0 {
				return nil, fmt.Errorf("format: %s's Extra holds %s, a field the schema names", src.Type(), m.Key)
			}
			out = append(out, schema.Member{Key: m.Key, Value: normalize(m.Value)})
		}
	}
	return out, nil
}

// normalize gives a value in the shapes schema.Validate reads: every integer a json.Number, every list []any.
func normalize(v any) any {
	switch x := v.(type) {
	case schema.Object:
		out := make(schema.Object, len(x))
		for i, m := range x {
			out[i] = schema.Member{Key: m.Key, Value: normalize(m.Value)}
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = normalize(e)
		}
		return out
	case []string:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = e
		}
		return out
	case int:
		return jsonInt(int64(x))
	case int64:
		return jsonInt(x)
	}
	return v
}

func jsonInt(i int64) json.Number { return json.Number(strconv.FormatInt(i, 10)) }

// fit reports how a Go type departs from a schema, for TestTypesFitTheirSchemas: a property with no field (a const
// aside), a field with no property, fields out of the schema's order, and a value field for a property that may be
// null (or a pointer for one that may not).
func fit(s schema.Object, t reflect.Type, at string) []string {
	var out []string
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch {
	case t == objectType || t == numberType || t.Kind() == reflect.Interface:
		return nil
	case t.Kind() == reflect.Slice:
		if items := sub(s, "items"); items != nil {
			return fit(items, t.Elem(), at+"[]")
		}
		return nil
	case t.Kind() != reflect.Struct:
		return nil
	}
	in := infoOf(t)
	ps := props(s)
	last := -1
	for _, p := range ps {
		pschema, _ := p.Value.(schema.Object)
		i, ok := in.byName[p.Key]
		if !ok {
			if _, isConst := pschema.Get("const"); !isConst {
				out = append(out, fmt.Sprintf("%s.%s: no field in %s", at, p.Key, t))
			}
			continue
		}
		pos := indexOf(in.order, p.Key)
		if pos < last {
			out = append(out, fmt.Sprintf("%s.%s: out of the schema's order in %s", at, p.Key, t))
		}
		last = pos
		ft := t.Field(i).Type
		canNil := ft.Kind() == reflect.Pointer || ft.Kind() == reflect.Slice || ft.Kind() == reflect.Interface ||
			ft == objectType
		if nullable(pschema) && !canNil {
			out = append(out, fmt.Sprintf("%s.%s: may be null, so its field needs a pointer", at, p.Key))
		}
		if !nullable(pschema) && ft.Kind() == reflect.Pointer {
			out = append(out, fmt.Sprintf("%s.%s: is never null, so its field is a value, not a pointer", at, p.Key))
		}
		out = append(out, fit(pschema, ft, at+"."+p.Key)...)
	}
	for _, name := range in.order {
		if ps.Index(name) < 0 {
			out = append(out, fmt.Sprintf("%s.%s: a field of %s the schema does not name", at, name, t))
		}
	}
	if in.extra < 0 {
		out = append(out, fmt.Sprintf("%s: %s has no Extra to keep unknown fields", at, t))
	}
	return out
}

func indexOf(list []string, s string) int {
	for i, x := range list {
		if x == s {
			return i
		}
	}
	return -1
}
