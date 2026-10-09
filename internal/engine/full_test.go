package engine

// The engine and check read bonsai.yaml and pack.yaml in full (plan-5 5.1.4a): a field of bonsai.workspace/1 the
// walking skeleton did not read is held to the schema, and a pack whose hook line calls bash by name is refused
// before anything is written.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestFullConfigRead(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "full")
	e.link(t, root, e.pack.A)
	before := snapshot(t, root)
	writeFile(t, root, workspace.ConfigFile, read(t, root, workspace.ConfigFile)+
		"ladder:\n  - rung: 0\n    kind: guard\n    required: maybe\n")
	r, err := Check(root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Findings) != 1 || r.Findings[0].Code != "config" || !strings.Contains(r.Findings[0].Message, "field ladder[0].required") ||
		!strings.Contains(r.Findings[0].Next, "bonsai.yaml") {
		t.Errorf("check: %+v", r.Findings)
	}
	_, err = e.try(root, Request{})
	if err == nil || !strings.Contains(err.Error(), "field ladder[0].required") || err.(*Error).Exit != ExitInput {
		t.Errorf("update: %v", err)
	}
	after := snapshot(t, root)
	delete(before, workspace.ConfigFile)
	delete(after, workspace.ConfigFile)
	sameSnapshot(t, "update refused for bonsai.yaml", before, after)
}

func TestPackHookCallingBashIsRefused(t *testing.T) {
	e := setup(t)
	hooks := "\n  - event: SessionStart\n    matcher: startup\n    command: \"bash run/hello.sh\"\n    runs: [\"run/hello.sh\"]\n" +
		"    why: \"Says hello when a session starts; it blocks nothing.\""
	source, shas := testpack.Fixture(t, e.tmp, "bash-pack", map[string]string{
		".claude-plugin/plugin.json": strings.Replace(fixturePluginJSON, "%s", "bash-pack", 1),
		"bonsai/pack.yaml": "format: bonsai.pack/1\nid: bash-pack\nversion: \"0.1.0\"\nfiles:\n  - path: run/hello.sh\n" +
			"    from: hello.sh\n    kind: pack\nhooks:" + hooks + "\ndeny: []\n",
		"bonsai/files/hello.sh": "echo hello\n",
	})
	root := testpack.Project(t, e.tmp, "bash")
	before := snapshot(t, root)
	_, err := e.try(root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: source, Ref: shas[0]},
		AllowExec: true})
	if err == nil || !strings.Contains(err.Error(), "calls bash by name") || !strings.Contains(err.Error(), "next: ") {
		t.Fatalf("init: %v", err)
	}
	sameSnapshot(t, "init refused", before, snapshot(t, root))
}

const fixturePluginJSON = "{\"name\": \"%s\", \"description\": \"A fixture pack.\"}\n"
