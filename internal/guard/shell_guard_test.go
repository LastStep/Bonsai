package guard

// The delete check's table through the guard itself (plan-5, "The order", item 5: "then its whole table runs"):
// every row of delTable answered by Main, the real judge, in real projects recorded on this machine (a Bonsai home of
// the test's own), not the stub. The projects mirror delFixture's: the floor and the lists docs/secret/**, config/*.key
// and notes/ in the session's project, another linked project beside it, and Bonsai's home in the home folder. Where
// the real judge answers a row otherwise than the stub, delRealAnswers says what and why; any other difference fails.
//
// With BONSAI_DELETE_TABLE set to a file, the test also writes each row it ran there, one JSON object a line (tool,
// the payload's folder relative to the project, the line, the rule), for a scripted run of the same rows through
// the built binary's hook line (never committed: the orchestrator's check).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

const delRealYAML = `format: bonsai.workspace/1
id: ws-dddddddddddddddddddddddddd
name: delete-table
packs: []
protected: ["docs/secret/**", "config/*.key", "notes/"]
person_only: []
never_edit: []
`

// delRealAnswers are the rows the real judge answers otherwise than the stub, with why.
var delRealAnswers = map[string]struct{ rule, why string }{}

// delRealFixture makes the real projects: proj (the session's, recorded), other (another project, recorded) and a home
// folder holding Bonsai's home; it gives proj's folder.
func delRealFixture(t *testing.T) string {
	t.Helper()
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	homeFolder := filepath.Join(top, "home")
	home := filepath.Join(homeFolder, ".bonsai")
	t.Setenv("HOME", homeFolder)
	t.Setenv("USERPROFILE", homeFolder)
	t.Setenv(workspace.HomeEnv, home)
	proj, other := filepath.Join(top, "proj"), filepath.Join(top, "other")
	files := map[string]string{
		"proj/bonsai.yaml": delRealYAML, "proj/.bonsai/STATE.md": "x\n",
		"proj/.bonsai/local/log/old.ndjson": "", "proj/.bonsai/tasks.md": "x\n", "proj/.git/HEAD": "ref: refs/heads/main\n",
		"proj/.git/objects/.keep": "", "proj/.git/refs/.keep": "", "proj/plant/.git/.keep": "", "proj/plant/sub/x.txt": "x\n",
		"proj/.claude/settings.json": "{}\n", "proj/.claude/notes.md": "x\n", "proj/docs/secret/plan.md": "x\n",
		"proj/docs/readme.md": "x\n", "proj/config/app.key": "x\n", "proj/config/app.yaml": "x\n", "proj/notes/a.md": "x\n",
		"proj/src/main.go": "x\n", "proj/build/out/x.o": "x\n", "proj/junk/one.txt": "x\n", "proj/junk/two.txt": "x\n",
		"proj/node_modules/pkg/index.js": "x\n", "proj/node_modules/pkg/.git/HEAD": "x\n",
		"proj/sub/bonsai.yaml": strings.Replace(delRealYAML, "ws-dddddddddddddddddddddddddd", "ws-eeeeeeeeeeeeeeeeeeeeeeeeee", 1),
		"proj/sub/src/x.go":    "x\n", "proj/vendor/lib/.git/HEAD": "x\n", "proj/file.txt": "x\n", "proj/a.txt": "x\n",
		"proj/x.yaml":       "x\n",
		"other/bonsai.yaml": strings.Replace(delRealYAML, "ws-dddddddddddddddddddddddddd", "ws-ffffffffffffffffffffffffff", 1),
		"other/.git/HEAD":   "ref: refs/heads/main\n", "other/src/x.go": "x\n",
		"home/.bonsai/settings.json": "{}\n", "home/notes.txt": "x\n",
	}
	for rel, body := range files {
		writeFile(t, filepath.Join(top, filepath.FromSlash(rel)), body)
	}
	for _, p := range []struct{ dir, id string }{{proj, "ws-dddddddddddddddddddddddddd"}, {other, "ws-ffffffffffffffffffffffffff"}} {
		if _, err := workspace.RecordCheckout(home, p.dir, p.id, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	return proj
}

func TestShellTableThroughTheGuard(t *testing.T) {
	proj := delRealFixture(t)
	var dump *os.File
	if out := os.Getenv("BONSAI_DELETE_TABLE"); out != "" {
		f, err := os.Create(out)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		dump = f
	}
	differ := 0
	for _, row := range delTable {
		if row.only == "windows" && runtime.GOOS != "windows" || row.only == "unix" && runtime.GOOS == "windows" {
			continue
		}
		cwd := proj
		if row.dir != "" {
			cwd = filepath.Join(proj, filepath.FromSlash(row.dir))
		}
		code, stderr := answer(t, proj, bashIn(cwd, row.tool, row.line))
		recs := records(t, proj, "s-r1.ndjson")
		got := recs[len(recs)-1].String("rule")
		want := row.want
		if ra, ok := delRealAnswers[row.tool+" "+row.line]; ok {
			want = ra.rule
			differ++
		}
		if got != want || (code == 2) == ruleAllows(t, got) {
			t.Errorf("%s %q: the guard answers %s (exit %d), want %s\n  %s", row.tool, row.line, got, code, want, stderr)
		}
		if dump != nil {
			b, err := json.Marshal(map[string]string{"tool": row.tool, "dir": row.dir, "line": row.line, "rule": got})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := dump.Write(append(b, '\n')); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Logf("%d rows answered otherwise by the real judge than by the stub, each named in delRealAnswers", differ)
}
