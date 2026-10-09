package engine

// bonsai.yaml as init writes it, from the built-in template: every field of bonsai.workspace/1, a comment on every
// line (at its end, or a long line's just above), its pack's document kinds at their defaults, and the template's
// comments held to the schema's fields both ways.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestInitWritesEveryField(t *testing.T) {
	e := setup(t)
	src, shas := testpack.DeclaringPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "template")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0], NeverEdit: []string{"work/ledger.json"}}})
	raw := read(t, root, workspace.ConfigFile)
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		commented := strings.HasPrefix(trimmed, "# ") || strings.Contains(l, "  # ") ||
			(i > 0 && strings.HasPrefix(strings.TrimSpace(lines[i-1]), "# "))
		if !commented {
			t.Errorf("line %d has no comment: %q", i+1, l)
		}
		if len(l) > 120 {
			t.Errorf("line %d is %d characters", i+1, len(l))
		}
		for _, r := range l {
			if r < 0x20 || r > 0x7e {
				t.Errorf("line %d is not ASCII: %q", i+1, l)
				break
			}
		}
	}
	if !strings.HasPrefix(raw, "# bonsai.yaml: ") || !strings.Contains(raw, "bonsai check --schema bonsai.workspace") {
		t.Errorf("no header naming the schema command:\n%s", raw)
	}
	// Every field of the schema, in its order, at the top and in each section; the pack's kinds at their defaults.
	r := reader.ReadYAML([]byte(raw))
	if r.Outcome != reader.Accepted {
		t.Fatalf("bonsai.yaml does not read: %v", r.Err())
	}
	doc := reader.JSON(r.Value).(schema.Object)
	s := format.MustLookup("workspace").Schema()
	props, _ := s.Get("properties")
	if got, want := strings.Join(doc.Keys(), " "), strings.Join(props.(schema.Object).Keys(), " "); got != want {
		t.Errorf("fields %q, want %q", got, want)
	}
	if msgs := format.MustLookup("workspace").Check(doc); msgs != nil {
		t.Errorf("bonsai.yaml does not fit the writer's schema: %v", msgs)
	}
	docs, _ := doc.Get("documents")
	if got := schema.Show(docs); got != `{"task":"work/tasks","run":"work/runs","answers":"work/answers.md","memory":"work/memory",`+
		`"protocols":"work/protocols","plan":"work/plans","bugs":"work/bugs.md"}` {
		t.Errorf("documents %s", got)
	}
	gen, _ := doc.Get("generated")
	if !schema.Equal(gen, mustDoc(t, format.DefaultGenerated())) {
		t.Errorf("generated %s", schema.Show(gen))
	}
	for _, want := range []string{`    source: "`, `    path: null`, `    ref: "` + shas[0] + `"`, `never_edit: ["work/ledger.json"]`} {
		if !strings.Contains(raw, want) {
			t.Errorf("no %q in\n%s", want, raw)
		}
	}
	// The engine and check read it in full, and the guard's lean read too.
	if _, err := workspace.ReadConfigFull([]byte(raw)); err != nil {
		t.Errorf("full read: %v", err)
	}
	if _, err := workspace.ReadConfig([]byte(raw)); err != nil {
		t.Errorf("lean read: %v", err)
	}
}

func mustDoc(t *testing.T, g format.Generated) any {
	t.Helper()
	d, err := format.MustLookup("workspace").Document(&format.Workspace{Generated: g})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := d.Get("generated")
	return v
}

// Every field of bonsai.workspace/1 has the template's comment (a pack's document kind and a generated kind's fields
// by comment's rules), and every comment names a field: a field change updates its comment in the same commit.
func TestTemplateDocumentsEveryField(t *testing.T) {
	s := format.MustLookup("workspace").Schema()
	var paths []string
	var walk func(o schema.Object, at string)
	walk = func(o schema.Object, at string) {
		props, _ := o.Get("properties")
		po, _ := props.(schema.Object)
		for _, m := range po {
			p := m.Key
			if at != "" {
				p = at + "." + m.Key
			}
			paths = append(paths, p)
			ps := m.Value.(schema.Object)
			if items, ok := ps.Get("items"); ok {
				if io, ok := items.(schema.Object); ok {
					if _, ok := io.Get("properties"); ok {
						paths = append(paths, p+"[]")
						walk(io, p+"[]")
					}
				}
			}
			walk(ps, p)
		}
	}
	walk(s, "")
	c := comment(map[string]string{"plan": "a pack's kind"})
	known := map[string]bool{}
	for _, p := range paths {
		known[p] = true
		if strings.HasPrefix(p, "format") || c(p) != "" {
			continue
		}
		t.Errorf("the field %s has no comment in the template", p)
	}
	for key := range Comments {
		if !known[key] {
			t.Errorf("the template's comment %q names no field of bonsai.workspace/1", key)
		}
		if len(Comments[key]) > 110 {
			t.Errorf("the comment for %s is %d characters", key, len(Comments[key]))
		}
	}
	if c("documents.plan") != "a pack's kind" || c("generated.log.keep_days") == "" || c("packs[3].ref") != Comments["packs[].ref"] {
		t.Errorf("comment's rules")
	}
}
