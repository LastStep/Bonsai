package engine

// A pack's roles (step 5.5.0, notes 7 and 8): copied into the lock's declares, the refusals of a role's name twice and
// of two session roles, and check's findings on the limits a project sets for them (checkroles.go).

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// roleYAML is one roles item of a pack.yaml: limits are minutes, actions and red_runs, each "default max" ("null"
// for none).
func roleYAML(name, kind string, session, writes bool, maxTurns string, limits [3]string) string {
	s := "  - name: " + name + "\n    kind: " + kind + "\n    session: " + map[bool]string{true: "true", false: "false"}[session] +
		"\n    writes: " + map[bool]string{true: "true", false: "false"}[writes] + "\n    max_turns: " + maxTurns + "\n"
	for i, l := range []string{"minutes", "actions", "red_runs"} {
		d, m, _ := strings.Cut(limits[i], " ")
		s += "    " + l + ":\n      default: " + d + "\n      max: " + m + "\n"
	}
	return s
}

// workflowRoles are three of workflow's six roles, as it ships them: the session's, a lead and a helper.
var workflowRoles = roleYAML("orchestrator", "lead", true, true, "null", [3]string{"null null", "null null", "null null"}) +
	roleYAML("builder", "lead", false, true, "2400", [3]string{"180 180", "600 600", "4 4"}) +
	roleYAML("auditor", "helper", false, false, "600", [3]string{"30 30", "150 150", "null null"})

// A pack's roles reach the lock's declares at a link, as its pack.yaml declares them, and every reader of a role reads
// them from there (workspace.RolesInForce), offline.
func TestRolesInTheLock(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "workflow", workflowRoles)
	root := testpack.Project(t, e.tmp, "roles")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	l, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if keys := strings.Join(l.Packs[0].Declares.Keys(), " "); keys != "roles" {
		t.Errorf("declares keys %q", keys)
	}
	d, err := l.Packs[0].Declared()
	if err != nil {
		t.Fatal(err)
	}
	pack, err := format.ReadPack([]byte("format: bonsai.pack/1\nid: workflow\nversion: \"1\"\nroles:\n" + workflowRoles))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := (&format.Declares{Roles: d.Roles}).Object()
	want, _ := (&format.Declares{Roles: pack.Roles}).Object()
	if len(d.Roles) != 3 || !schema.Equal(got, want) {
		t.Errorf("the lock's roles\n%s\nwant\n%s", schema.Show(got), schema.Show(want))
	}
	cfg, _ := workspace.LoadConfigFull(root)
	in, err := workspace.RolesInForce(l, cfg.Full)
	if err != nil || len(in) != 3 || in[1].Role.Name != "builder" || *in[1].Actions.Max != 600 || in[1].Actions.MaxFrom != workspace.LimitFromPack {
		t.Errorf("roles in force: %v %+v", err, in)
	}
	if !strings.Contains(read(t, root, "bonsai.yaml"), "\n    roles: null\n") {
		t.Errorf("init's packs entry has no roles: null line:\n%s", read(t, root, "bonsai.yaml"))
	}
	// A pack that declares none writes no roles key, as before set 8.
	src2, shas2 := protoPack(t, e.tmp, "plain", "")
	root2 := testpack.Project(t, e.tmp, "no-roles")
	e.apply(t, root2, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src2, Ref: shas2[0]}})
	l2, _ := workspace.LoadLock(root2)
	if _, ok := l2.Packs[0].Declares.Get("roles"); ok {
		t.Errorf("a pack with roles: [] wrote roles: %v", l2.Packs[0].Declares)
	}
}

// The role rules a schema cannot say, at a link: a role's name twice in one pack and two session roles in one pack
// are bad-pack; two packs declaring a session role between them are refused naming both; two packs declaring one
// role's name each keep their own.
func TestRolesRefused(t *testing.T) {
	e := setup(t)
	link := func(t *testing.T, name string, packs ...[3]string) error {
		t.Helper()
		root := testpack.Project(t, e.tmp, name)
		y := packsYAML(packs...)
		y += "documents:\n  task: work/tasks\n  run: work/runs\n  answers: work/answers.md\n  memory: work/memory\n  protocols: work/protocols\n"
		writeFile(t, root, "bonsai.yaml", y)
		_, err := e.try(root, Request{Command: "init"})
		return err
	}
	builder := roleYAML("builder", "lead", false, true, "null", [3]string{"null null", "null null", "null null"})
	lead := roleYAML("lead", "lead", true, true, "null", [3]string{"null null", "null null", "null null"})
	twice, s1 := protoPack(t, e.tmp, "twice", builder+builder)
	wantRefused(t, link(t, "role-twice", [3]string{"twice", twice, s1[0]}), "bad-pack", ExitInput, `declares the role "builder" twice`)
	sessions, s2 := protoPack(t, e.tmp, "sessions", workflowRoles+lead)
	wantRefused(t, link(t, "two-sessions", [3]string{"sessions", sessions, s2[0]}), "bad-pack", ExitInput,
		`roles item 4 ("lead") has session true, and so does "orchestrator"`)
	wf, s3 := protoPack(t, e.tmp, "workflow", workflowRoles)
	other, s4 := protoPack(t, e.tmp, "other", lead+builder)
	wantRefused(t, link(t, "two-packs", [3]string{"workflow", wf, s3[0]}, [3]string{"other", other, s4[0]}), "packs-overlap", ExitInput,
		"workflow:orchestrator and other:lead")
	// One role's name in two packs (builder in both): each keeps its own.
	mine, s5 := protoPack(t, e.tmp, "mine", builder)
	if err := link(t, "one-name", [3]string{"workflow", wf, s3[0]}, [3]string{"mine", mine, s5[0]}); err != nil {
		t.Errorf("one role's name in two packs: %v", err)
	}
}

// check's findings on the limits bonsai.yaml sets for a pack's roles (format.RoleCheckWords): a limit for a role the
// pack does not declare, a default above its max, and a max of actions above the role's max_turns; each a finding, an
// agent's, about bonsai.yaml; and limits within bounds give none, raised or lowered.
func TestRoleFindings(t *testing.T) {
	e := setup(t)
	src, shas := protoPack(t, e.tmp, "workflow", workflowRoles)
	root := testpack.Project(t, e.tmp, "role-findings")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	linked := read(t, root, "bonsai.yaml")
	for _, c := range []struct {
		name, roles string
		want        []string // each finding: its code, then words of its message
	}{
		{"within bounds", "      builder:\n        minutes:\n          default: 240\n          max: 480\n        actions:\n          max: 2400\n" +
			"      auditor:\n        actions:\n          default: 10\n", nil},
		{"undeclared", "      tester:\n        minutes:\n          default: 5\n",
			[]string{`role-undeclared bonsai.yaml sets limits for the role "tester" under the pack workflow (packs[0].roles.tester), which ` +
				"that pack does not declare (its roles: orchestrator, builder, auditor)"}},
		{"over its max", "      builder:\n        minutes:\n          default: 240\n      orchestrator:\n        actions:\n          max: 10\n",
			[]string{"role-default-over-max the role workflow:orchestrator's actions default in force, no limit (the pack's), is above its max, 10 (bonsai.yaml's)",
				"role-default-over-max the role workflow:builder's minutes default in force, 240 (bonsai.yaml's), is above its max, 180 (the pack's)"}},
		{"actions over turns", "      builder:\n        actions:\n          max: 2401\n      auditor:\n        actions:\n          default: null\n          max: null\n",
			[]string{"role-actions-over-turns the role workflow:builder's max of actions in force, 2401 (bonsai.yaml's), is above its max_turns, 2400",
				"role-actions-over-turns the role workflow:auditor's max of actions in force, no limit (bonsai.yaml's), is above its max_turns, 600"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			writeFile(t, root, "bonsai.yaml", strings.Replace(linked, "    roles: null\n", "    roles:\n"+c.roles, 1))
			r, err := checkLocal(t, root, e.home)
			if err != nil {
				t.Fatal(err)
			}
			var got []Finding
			for _, f := range r.Findings {
				if strings.HasPrefix(f.Code, "role-") {
					got = append(got, f)
				} else {
					t.Errorf("another finding: %+v", f)
				}
			}
			if len(got) != len(c.want) {
				t.Fatalf("findings %+v, want %d", got, len(c.want))
			}
			for i, f := range got {
				code, words, _ := strings.Cut(c.want[i], " ")
				if f.Code != code || !strings.Contains(f.Message, words) || f.File != "bonsai.yaml" || f.Who != "agent" || f.Next == "" {
					t.Errorf("finding %d: %+v, want %s holding %q", i+1, f, code, words)
				}
			}
		})
	}
}
