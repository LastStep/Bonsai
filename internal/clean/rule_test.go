package clean

import (
	"fmt"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// generated.<kind> as the cleaner reads it: the defaults when left out, null keeping, each number held to its rule.
func TestRuleOf(t *testing.T) {
	show := func(r Rule) string {
		n := func(p *int64) string {
			if p == nil {
				return "null"
			}
			return fmt.Sprint(*p)
		}
		return n(r.KeepDays) + "/" + n(r.KeepNewest)
	}
	cases := []struct {
		name, generated, kind, want string
	}{
		{"left out: log's default", "", Log, "30/null"},
		{"left out: asks' default", "", Asks, "null/null"},
		{"left out: ladder's default", "", Ladder, "7/null"},
		{"left out: sessions' default", "", Sessions, "null/null"},
		{"the kind left out", "  ladder:\n    keep_days: 3\n", Log, "30/null"},
		{"given", "  log:\n    keep_days: 5\n    keep_newest: 100\n", Log, "5/100"},
		{"a field left out is null", "  log:\n    keep_newest: 9\n", Log, "null/9"},
		{"null keeps", "  log:\n    keep_days: null\n    keep_newest: null\n", Log, "null/null"},
		{"the kind null", "  log: null\n", Log, "null/null"},
		{"zero", "  ladder:\n    keep_days: 0\n", Ladder, "0/null"},
		{"negative", "  log:\n    keep_days: -1\n", Log, "error"},
		{"a word", "  log:\n    keep_days: soon\n", Log, "error"},
		{"a decimal", "  log:\n    keep_days: 1.5\n", Log, "error"},
		{"a list", "  log: [1]\n", Log, "error"},
		{"run takes a rule but is never cleaned", "  run:\n    keep_days: 1\n", "run", "1/null"},
		{"tasks takes none", "", "tasks", "error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := project(t, c.generated)
			cfg, err := workspace.LoadConfig(l.Main)
			if err != nil {
				t.Fatal(err)
			}
			r, err := RuleOf(cfg.Doc, c.kind)
			got := show(r)
			if err != nil {
				got = "error"
				if !strings.HasPrefix(err.Error(), "clean: ") {
					t.Errorf("the error %q does not name the cleaner", err)
				}
			}
			if got != c.want {
				t.Errorf("%s: %s, want %s", c.kind, got, c.want)
			}
		})
	}
}

// The defaults RuleOf gives are format.GeneratedKinds', their one home.
func TestRuleDefaultsAreTheTable(t *testing.T) {
	l := project(t, "")
	cfg, err := workspace.LoadConfig(l.Main)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range format.GeneratedKinds {
		if !k.Rule {
			continue
		}
		r, err := RuleOf(cfg.Doc, k.Kind)
		if err != nil {
			t.Fatal(err)
		}
		if !equalPtr(r.KeepDays, k.KeepDays) || !equalPtr(r.KeepNewest, k.KeepNewest) {
			t.Errorf("%s: the default differs from format.GeneratedKinds'", k.Kind)
		}
	}
}

func equalPtr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// The clean record's event is one the log's table of events knows.
func TestCleanIsALogEvent(t *testing.T) {
	for _, w := range format.LogEvents {
		if w.Word == "clean" {
			return
		}
	}
	t.Error("format.LogEvents has no clean")
}
