package format

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// A declares reads back as written, each key in DeclaresKeys' order, an unknown key kept after them; {} for none;
// a part that does not fit its format is refused, naming the key.
func TestDeclaresReadAndWrite(t *testing.T) {
	if o, err := (&Declares{}).Object(); err != nil || len(o) != 0 {
		t.Errorf("nothing declared: %s %v", schema.Show(o), err)
	}
	matcher, floor := "startup", "2.1.300"
	turns, mins := int64(2400), int64(180)
	d := &Declares{
		Deny:      []PackDeny{{Rule: "Read(x)", Why: "No."}},
		Hooks:     []PackHook{{Event: "SessionStart", Matcher: &matcher, Command: "echo hi", Runs: []string{}, Why: "Says hi."}},
		Protected: []string{"docs/**"},
		Lanes:     []Lane{{Name: "light", Close: "agent", Description: "Small."}},
		Needs:     &PackNeeds{ClaudeCode: &floor},
		Roles: []PackRole{{Name: "builder", Kind: "lead", Writes: true, MaxTurns: &turns,
			Minutes: RoleLimit{Default: &mins, Max: &mins}}},
		Extra: schema.Object{{Key: "later", Value: true}},
	}
	o, err := d.Object()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(o.Keys(), " "); got != "lanes protected hooks deny needs roles later" {
		t.Errorf("keys %q", got)
	}
	back, err := ReadDeclares(o)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := back.Object()
	if !schema.Equal(again, o) || back.Needs == nil || *back.Needs.ClaudeCode != floor || len(back.Roles) != 1 ||
		*back.Roles[0].Minutes.Max != 180 || back.Roles[0].Actions.Default != nil {
		t.Errorf("read back as %s", schema.Show(again))
	}
	// A pack whose needs names nothing declares none.
	if o, _ := (&Declares{Needs: &PackNeeds{}}).Object(); len(o) != 0 {
		t.Errorf("empty needs: %s", schema.Show(o))
	}
	for key, bad := range map[string]any{"lanes": "x", "labels": schema.Object{{Key: "version", Value: "one"}},
		"hooks": []any{schema.Object{{Key: "event", Value: true}}},
		"roles": []any{schema.Object{{Key: "kind", Value: "boss"}}}} {
		if _, err := ReadDeclares(schema.Object{{Key: key, Value: bad}}); err == nil || !strings.Contains(err.Error(), "declares") {
			t.Errorf("a bad %s: %v", key, err)
		}
	}
}
