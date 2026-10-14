package guard

// The guard's reads stay lean (plan-5 5.1.4a, widened by 5.3.1 note 3): it reads the fields of bonsai.yaml it judges
// by through workspace.LoadConfig, never the full read the engine and check use (workspace.LoadConfigFull), and of the
// lock only each pack's id and its declares' protected paths and document kinds (workspace.ReadDeclared), not the
// full read (workspace.ReadLock), which only command mode's tripwires make, for a document whose labels change or a
// new document of a pack's kind (tripwire.go). So a field it does not judge by, broken, never blocks an ordinary edit,
// and an ordinary hook call never pays for a full read.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// leanFields are the keys the guard's read (workspace.ReadConfig) reads and checks: the format line, the workspace
// id its records carry, the name and packs the lean read has checked since the walking skeleton (packs also says
// whether a missing lock is wrong), documents (the task folder and the declared kinds' places, which step 5.3.2 judges
// by), and the protected lists and never_edit, which it judges by. Every other key of bonsai.workspace/1 is outside it.
var leanFields = []string{"format", "id", "name", "packs", "documents", "protected", "person_only", "never_edit"}

// brokenElsewhere breaks every field of bonsai.workspace/1 outside leanFields, each in a way the full read refuses.
var brokenElsewhere = map[string]string{
	"ladder_floor":    "ladder_floor: [\"zero\"]\n",
	"ladder":          "ladder:\n  - rung: 1\n    kind: command\n    command: \"go test ./...\"\n    required: true\n    timeout_s: \"ten\"\n",
	"ratchets":        "ratchets:\n  tests: \"many\"\n",
	"ci_marked_tests": "ci_marked_tests: 3\n",
	"generated":       "generated:\n  log:\n    keep_days: -1\n  sessions:\n    keep_newest: \"all\"\n",
	"messages":        "messages:\n  unanswered_minutes: \"an hour\"\n  circling_briefs: null\n", // set 8
}

func TestGuardReadIsLean(t *testing.T) {
	// Every field of the schema is either read by the guard or broken below.
	props := workspaceProps(t)
	for _, p := range props {
		if !contains(leanFields, p) && brokenElsewhere[p] == "" {
			t.Errorf("bonsai.workspace/1's field %s is neither in the guard's read nor broken by this test", p)
		}
	}
	for field, broken := range brokenElsewhere {
		t.Run(field, func(t *testing.T) {
			dir := project(t, testYAML+broken)
			// The full read refuses the file, naming the field ...
			_, err := workspace.LoadConfigFull(dir)
			var we *workspace.Error
			if !errors.As(err, &we) || !strings.Contains(we.Msg, "field "+field) {
				t.Fatalf("the full read does not refuse the broken %s: %v", field, err)
			}
			// ... and the guard still decides: a protected path is refused for being protected, a free one allowed.
			code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lean", dir, "protected.txt")), untimed)
			if code != 2 || !strings.Contains(stderr, "protected") || strings.Contains(stderr, "bonsai.yaml cannot be read") {
				t.Errorf("a protected edit: exit %d, %s", code, stderr)
			}
			code, stderr = guardRun(t, env(dir), strings.NewReader(editOf("lean", dir, "free.txt")), untimed)
			if code != 0 {
				t.Errorf("a free edit: exit %d, %s", code, stderr)
			}
		})
	}
	// A broken field the guard does read still blocks, as it always has: the guard cannot tell what is protected.
	for name, yaml := range map[string]string{
		"a protected glob":    strings.Replace(testYAML, `protected: [`, `protected: ["/abs", `, 1),
		"a document's place":  testYAML + "documents:\n  task: \"/abs/tasks\"\n",
		"documents as a list": testYAML + "documents: [\"work/tasks\"]\n",
	} {
		dir := project(t, yaml)
		if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lean", dir, "free.txt")), untimed); code != 2 ||
			!strings.Contains(stderr, "bonsai.yaml cannot be read") {
			t.Errorf("%s: exit %d, %s", name, code, stderr)
		}
	}
	// documents, read: each kind and its place, in the file's order.
	cfg, err := workspace.ReadConfig([]byte(testYAML + "documents:\n  task: work/tasks\n  plan: work/plans\n"))
	if err != nil || len(cfg.Documents) != 2 || cfg.Documents[0] != (workspace.DocPlace{Kind: "task", Path: "work/tasks", Line: 9}) ||
		cfg.Documents[1].Kind != "plan" || cfg.Documents[1].Path != "work/plans" {
		t.Errorf("documents read as %+v, %v", cfg, err)
	}
}

// lockWith is a lock of one pack whose declares are the given JSON, and the given files.
func lockWith(declares, files string) string {
	return `{"format":"bonsai.lock/1","written_by":"1.0.0","packs":[{"id":"base","source":"https://example.com/base.git",` +
		`"version":"1.0.0","commit":"0a1b2c3d4e5f60718293a4b5c6d7e8f901a2b3c4","sha256":"` + strings.Repeat("a", 64) +
		`","declares":` + declares + `,"path":null}],"files":` + files + `,"format0":{}}`
}

// The lock's lean read: a pack's declared protected path is guarded; a lock broken where the guard does not read it
// (a file entry, another declares kind) still decides, though the full read refuses it; a lock broken where it reads
// (its format, the protected list, a document kind's place, the JSON itself) refuses every file tool's path as
// bad-lock, recorded; a missing lock is nothing declared when bonsai.yaml names no pack, and bad-lock when it names one.
func TestGuardReadOfTheLockIsLean(t *testing.T) {
	const declares = `{"protected":["work/protocols/**"],"documents":[{"kind":"plan","path":"work/plans","file":null}]}`
	files := `{"CLAUDE.md":{"kind":"block","pack":"base","sha256":"` + strings.Repeat("b", 64) + `"}}`
	withLock := func(t *testing.T, yaml, lock string) string {
		t.Helper()
		dir := project(t, yaml)
		if lock != "" {
			if err := os.MkdirAll(filepath.Join(dir, ".bonsai"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, ".bonsai", "lock.json"), []byte(lock), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	// Read in full, the good lock reads; the declared path is refused, a free one allowed.
	if _, err := workspace.ReadLock([]byte(lockWith(declares, files))); err != nil {
		t.Fatalf("the good lock does not read in full: %v", err)
	}
	d, err := workspace.ReadDeclared([]byte(lockWith(declares, files)))
	if err != nil || len(d) != 1 || d[0].Pack != "base" || len(d[0].Protected) != 1 || len(d[0].Documents) != 1 ||
		d[0].Documents[0] != (workspace.DocPlace{Kind: "plan", Path: "work/plans"}) {
		t.Fatalf("declared %+v, %v", d, err)
	}
	dir := withLock(t, testYAML, lockWith(declares, files))
	code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lock", dir, "work/protocols/start.md")), untimed)
	if code != 2 || !strings.Contains(stderr, `the protected paths the pack "base" declares ("work/protocols/**")`) {
		t.Errorf("a declared path: exit %d, %s", code, stderr)
	}
	if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lock", dir, "work/plans/p.md")), untimed); code != 0 {
		t.Errorf("a free path: exit %d, %s", code, stderr)
	}
	// Broken where the guard does not read: the full read refuses, the guard decides.
	for name, lock := range map[string]string{
		"a file entry":      lockWith(declares, `{"CLAUDE.md":{"kind":"nonsense","pack":"base","sha256":"x"}}`),
		"a lane":            lockWith(`{"lanes":[{"name":7}],"protected":["work/protocols/**"]}`, files),
		"the pack's commit": strings.Replace(lockWith(declares, files), `"commit":"0a1b`, `"commit":"zz`, 1),
		"an unknown key":    strings.Replace(lockWith(declares, files), `"format0":{}`, `"format0":{},"later":[1]`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			lock := lock
			dir := withLock(t, testYAML, lock)
			if name != "an unknown key" {
				if _, err := workspace.ReadLock([]byte(lock)); err == nil {
					t.Fatalf("the full read takes it")
				}
			}
			if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lock", dir, "work/protocols/a.md")), untimed); code != 2 ||
				!strings.Contains(stderr, "declares") {
				t.Errorf("a declared path: exit %d, %s", code, stderr)
			}
			if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lock", dir, "free.txt")), untimed); code != 0 {
				t.Errorf("a free path: exit %d, %s", code, stderr)
			}
		})
	}
	// Broken where the guard reads: every file tool's path refused (bad-lock), and the refusal recorded.
	for name, lock := range map[string]string{
		"not JSON":             "{",
		"a key twice":          strings.Replace(lockWith(declares, files), `"written_by"`, `"packs":[],"written_by"`, 1),
		"a newer format":       strings.Replace(lockWith(declares, files), "bonsai.lock/1", "bonsai.lock/2", 1),
		"protected not a list": lockWith(`{"protected":"work/**"}`, files),
		"an absolute glob":     lockWith(`{"protected":["/etc/**"]}`, files),
		"a place with a colon": lockWith(`{"documents":[{"kind":"plan","path":"C:/plans","file":null}]}`, files),
		"a pack with no id":    `{"format":"bonsai.lock/1","packs":[{"declares":{}}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := withLock(t, testYAML, lock)
			code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lock", dir, "free.txt")), untimed)
			if code != 2 || !strings.Contains(stderr, ".bonsai/lock.json cannot be read") || !strings.Contains(stderr, "\nnext: ") {
				t.Fatalf("exit %d, %s", code, stderr)
			}
			if recs := records(t, dir, "s-lock.ndjson"); len(recs) != 1 || recs[0].String("rule") != RuleBadLock {
				t.Errorf("records %v", recs)
			}
			// A shell call judges no path: allowed.
			if code, _ := guardRun(t, env(dir), strings.NewReader(payload("lock", dir, "Bash", "command", "ls")), untimed); code != 0 {
				t.Errorf("a shell call: exit %d", code)
			}
		})
	}
	// No lock: nothing declared with no packs; bad-lock with a pack named.
	dir = withLock(t, testYAML, "")
	if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lock", dir, "free.txt")), untimed); code != 0 {
		t.Errorf("no lock, no packs: exit %d, %s", code, stderr)
	}
	withPack := strings.Replace(testYAML, "packs: []", "packs:\n  - id: base\n    source: \"https://example.com/base.git\"\n    ref: v1", 1)
	dir = withLock(t, withPack, "")
	if code, stderr := guardRun(t, env(dir), strings.NewReader(editOf("lock", dir, "free.txt")), untimed); code != 2 ||
		!strings.Contains(stderr, ".bonsai/lock.json cannot be read (.bonsai/lock.json: is missing") {
		t.Errorf("no lock, a pack named: exit %d, %s", code, stderr)
	}
}

// workspaceProps lists bonsai.workspace/1's top-level fields, from the schema (their one home).
func workspaceProps(t *testing.T) []string {
	t.Helper()
	raw, err := formats.Schema("workspace")
	if err != nil {
		t.Fatal(err)
	}
	s, err := schema.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	props, _ := s.Get("properties")
	keys := props.(schema.Object).Keys()
	if len(keys) < len(leanFields)+len(brokenElsewhere) {
		t.Fatalf("the schema lists %d fields", len(keys))
	}
	return keys
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
