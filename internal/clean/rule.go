package clean

// generated.<kind>, read from bonsai.yaml as read (workspace.Config.Doc), so the lean read the hooks use serves: the
// cleaner judges by this section alone, and a problem elsewhere in the file (bonsai check names it) does not stop
// it.
//
//   - generated: left out, or a kind left out of it: the kind's default (format.GeneratedKinds).
//   - generated: null, or a kind null: keeps (no rule).
//   - keep_days or keep_newest null, or left out of a kind that is given: that rule is off (the schema: "a reader
//     treats a missing field as null").
//   - Anything else (a text, a negative number, a list): the kind's rule does not read, and nothing of that kind is
//     cleaned until a person fixes it; bonsai check's full read names the line.

import (
	"fmt"
	"strconv"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
)

// Rule is one kind's rule: nil for a rule that is off.
type Rule struct {
	KeepDays   *int64
	KeepNewest *int64
}

// Keeps reports a rule that cleans nothing: both off.
func (r Rule) Keeps() bool { return r.KeepDays == nil && r.KeepNewest == nil }

// RuleOf reads kind's rule from bonsai.yaml as read (doc: workspace.Config.Doc).
func RuleOf(doc *reader.Map, kind string) (Rule, error) {
	def, ok := format.GeneratedKindOf(kind)
	if !ok || !def.Rule {
		return Rule{}, fmt.Errorf("clean: %q is not a generated kind that takes a rule", kind)
	}
	g, ok := doc.Get("generated")
	if !ok {
		return Rule{KeepDays: def.KeepDays, KeepNewest: def.KeepNewest}, nil
	}
	if g == nil {
		return Rule{}, nil
	}
	gm, isMap := g.(*reader.Map)
	if !isMap {
		return Rule{}, fmt.Errorf("clean: bonsai.yaml's generated is not a mapping of kinds")
	}
	v, ok := gm.Get(kind)
	if !ok {
		return Rule{KeepDays: def.KeepDays, KeepNewest: def.KeepNewest}, nil
	}
	if v == nil {
		return Rule{}, nil
	}
	km, isMap := v.(*reader.Map)
	if !isMap {
		return Rule{}, fmt.Errorf("clean: bonsai.yaml's generated.%s is not a mapping of keep_days and keep_newest", kind)
	}
	var rule Rule
	var err error
	if rule.KeepDays, err = count(km, kind, "keep_days"); err != nil {
		return Rule{}, err
	}
	if rule.KeepNewest, err = count(km, kind, "keep_newest"); err != nil {
		return Rule{}, err
	}
	return rule, nil
}

// count reads one rule's number: nil when it is null or left out; a whole number of 0 or more otherwise.
func count(m *reader.Map, kind, key string) (*int64, error) {
	v, ok := m.Get(key)
	if !ok || v == nil {
		return nil, nil
	}
	n, isInt := v.(int64)
	if !isInt || n < 0 {
		return nil, fmt.Errorf("clean: bonsai.yaml's generated.%s.%s is not a whole number of 0 or more, nor null", kind, key)
	}
	return &n, nil
}

// reason names a rule as a clean record says it: generated.log.keep_days=30.
func reason(kind, key string, n int64) string {
	return "generated." + kind + "." + key + "=" + strconv.FormatInt(n, 10)
}
