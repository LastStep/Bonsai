package engine

// Base's workspace and state skills, generated whole from the templates the engine writes (skills.go): each holds its
// template as init writes it, and each passes check --pack's rules for a template skill (its fields table against its
// template and its format's schema, its allowed values against the closed lists). That the committed files are what
// these build is internal/reference's TestGeneratedAreCurrent.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestGeneratedSkills(t *testing.T) {
	ws, st := string(WorkspaceSkill()), string(StateSkill())
	tmpl, err := ConfigYAML(exampleID, InitValues{Name: "example", Source: exampleSource, Ref: "v1.0.0"},
		&PackData{Manifest: &workspace.Pack{ID: "example-pack"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ws, "```yaml\n"+string(tmpl)+"```\n") {
		t.Error("the workspace skill does not hold bonsai.yaml's template as init writes it")
	}
	if !strings.Contains(st, "```markdown\n"+string(StateTemplate("YYYY-MM-DD"))+"```\n") {
		t.Error("the state skill does not hold STATE's template as init writes it")
	}
	if !strings.Contains(ws, "(skill base:generated-files)") || strings.Contains(ws, "docs/reference/generated-files.md") {
		t.Error("the comment on generated: does not name the generated-files skill")
	}
	// The example ladder reads as bonsai.yaml's ladder, its floor and its ratchets.
	yaml := strings.NewReplacer("ladder_floor: []", "", "ladder: []", "", "ratchets: {}", "").Replace(string(tmpl))
	c, err := workspace.ReadConfigFull([]byte(yaml + exampleLadder))
	if err != nil {
		t.Fatalf("the example ladder does not read: %v", err)
	}
	if len(c.Full.Ladder) != 2 || len(c.Full.LadderFloor) != 2 || c.Full.Ladder[1].Tests == nil || *c.Full.Ladder[1].Tests != "go-test-json" {
		t.Errorf("the example ladder reads as %+v", c.Full.Ladder)
	}
	for name, s := range map[string]string{"workspace": ws, "state": st} {
		if !strings.HasPrefix(s, "---\nname: "+name+"\ndescription: \"") || !strings.Contains(s, GeneratedNote) {
			t.Errorf("the %s skill's head:\n%s", name, s[:200])
		}
		for i, l := range strings.Split(s, "\n") {
			for _, r := range l {
				if r < 0x20 || r > 0x7e {
					t.Errorf("the %s skill's line %d is not ASCII: %q", name, i+1, l)
					break
				}
			}
		}
	}
	// check --pack, on a pack holding them beside the documented pack's own: no finding, four template skills.
	dir := filepath.Join(t.TempDir(), "pack")
	files := testpack.DocumentedPack()
	files["skills/workspace/SKILL.md"], files["skills/state/SKILL.md"] = ws, st
	testpack.WriteFolder(t, dir, files)
	r, e := CheckPack(dir)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Findings) != 0 {
		t.Errorf("check --pack finds %+v", r.Findings)
	}
	if len(r.Notes) != 1 || !strings.Contains(r.Notes[0], "4 template skills (skills/note, skills/state, skills/task, skills/workspace)") {
		t.Errorf("check --pack's note: %v", r.Notes)
	}
	// A template skill's table and its template held together: a field taken out of the template is found.
	files["skills/state/SKILL.md"] = strings.Replace(st, "updated_by: bonsai init\n", "", 1)
	testpack.WriteFolder(t, dir, files)
	if r, _ := CheckPack(dir); len(r.Findings) == 0 || r.Findings[0].Code != "pack-fields" {
		t.Errorf("a changed template: %+v", r.Findings)
	}
}
