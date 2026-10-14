package engine

// check --pack's pack-deny-relative (set 8; plan-5 5.5.0 note 9) names a form that holds after a shell cd with the
// reach the rule has (VM's wording point): a name with no slash, which matches at every depth, at every depth
// (Edit(*.env) as Edit(//**/*.env), never Edit(/*.env), which would lose every .env below the top); a path with a slash
// before its end from the project's top folder; a rule already anchored, none.

import (
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/reader"
)

func TestDenyRelativeKeepsItsReach(t *testing.T) {
	rules := []struct{ rule, holds string }{
		{"Edit(*.env)", "Edit(//**/*.env)"},
		{"Read(.env)", "Read(//**/.env)"},
		{"Edit(secrets/)", "Edit(//**/secrets/)"},
		{"Read(docs/secret.txt)", "Read(/docs/secret.txt)"},
		{"Edit(./docs/**)", "Edit(/docs/**)"},
		{"Edit(./*.env)", "Edit(/*.env)"},
		{"Write(**/keys.md)", "Write(/**/keys.md)"},
		{"Edit(/top.md)", ""}, {"Read(//etc/hosts)", ""}, {"Edit(~/.ssh/**)", ""}, {"Read(C:/x)", ""}, {"Bash(rm *)", ""},
	}
	var y strings.Builder
	y.WriteString("format: bonsai.pack/1\nid: p\ndeny:\n")
	for _, r := range rules {
		y.WriteString("  - rule: \"" + r.rule + "\"\n    why: \"A wall.\"\n")
	}
	read := reader.ReadYAML([]byte(y.String()))
	if read.Value == nil {
		t.Fatalf("the fixture does not read: %v", read.Err())
	}
	found := denyRelative("bonsai/pack.yaml", read.Value)
	i := 0
	for n, r := range rules {
		if r.holds == "" {
			continue
		}
		if i >= len(found) {
			t.Fatalf("%s: no finding (found %d)", r.rule, len(found))
		}
		f := found[i]
		i++
		item := "deny item " + strconv.Itoa(n+1) + "'s rule"
		if f.code != "pack-deny-relative" || !strings.Contains(f.msg, item+" \""+r.rule+"\"") ||
			!strings.Contains(f.msg, "the form that holds is \""+r.holds+"\"") || !strings.Contains(f.next, "as \""+r.holds+"\"") {
			t.Errorf("%s: %s\nnext: %s\nwant %s", r.rule, f.msg, f.next, r.holds)
		}
		if anyDepth := strings.HasPrefix(r.holds[strings.Index(r.holds, "(")+1:], "//**/"); anyDepth != strings.Contains(f.msg, "at any depth") ||
			anyDepth == strings.Contains(f.next, "(or with // for an absolute path)") {
			t.Errorf("%s: how it holds: %s / %s", r.rule, f.msg, f.next)
		}
	}
	if i != len(found) {
		t.Errorf("%d findings, want %d: %v", len(found), i, found)
	}
}
