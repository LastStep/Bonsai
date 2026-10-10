package format

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// The generated kinds that take a rule are bonsai.workspace/1's generated properties, in order; DefaultGenerated
// writes each kind's default, and fits the schema.
func TestGeneratedKindsAreTheSchemas(t *testing.T) {
	props, _ := MustLookup("workspace").Schema().Get("properties")
	gen, _ := props.(schema.Object).Get("generated")
	gp, _ := gen.(schema.Object).Get("properties")
	var ours []string
	for _, k := range GeneratedKinds {
		if k.Rule {
			ours = append(ours, k.Kind)
		}
		if k.Where == "" || k.What == "" || k.DefaultWords() == "" || k.Never == "" || k.Writer == "" || k.When == "" {
			t.Errorf("%s is not documented in full", k.Kind)
		}
	}
	if got, want := strings.Join(ours, " "), strings.Join(gp.(schema.Object).Keys(), " "); got != want {
		t.Errorf("generated kinds with a rule %q, the schema's %q", got, want)
	}
	doc, err := MustLookup("workspace").Document(&Workspace{Generated: DefaultGenerated()})
	if err != nil {
		t.Fatal(err)
	}
	g, _ := doc.Get("generated")
	if got := schema.Show(g); got != `{"log":{"keep_days":30,"keep_newest":null},"asks":{"keep_days":null,"keep_newest":null},`+
		`"ladder":{"keep_days":7,"keep_newest":null},"run":{"keep_days":null,"keep_newest":null},"sessions":{"keep_days":null,"keep_newest":null}}` {
		t.Errorf("defaults %s", got)
	}
	if msgs := schema.Validate(sub(sub(MustLookup("workspace").Schema(), "properties"), "generated"), g); len(msgs) > 0 {
		t.Errorf("defaults do not fit: %v", msgs)
	}
	d := DefaultGenerated()
	*d.Log.KeepDays = 1
	if *DefaultGenerated().Log.KeepDays != 30 {
		t.Errorf("DefaultGenerated shares the table's values")
	}
}

// A kind's default in words is built from its numbers, so a changed number changes every page that prints it.
func TestDefaultWordsFollowTheNumbers(t *testing.T) {
	want := map[string]string{"log": "30 days after a file's last line", "asks": "kept", "ladder": "7 days after the result's finished",
		"run": "kept", "sessions": "kept", "tasks": "a rebuild: nothing to clean"}
	for _, k := range GeneratedKinds {
		if got := k.DefaultWords(); got != want[k.Kind] {
			t.Errorf("%s default %q, want %q", k.Kind, got, want[k.Kind])
		}
	}
	k := GeneratedKind{Rule: true, KeepDays: days(8), KeepNewest: days(5), Age: "after the end"}
	if got := k.DefaultWords(); got != "8 days after the end, and the newest 5" {
		t.Errorf("both numbers: %q", got)
	}
}
