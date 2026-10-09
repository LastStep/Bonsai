package engine

// The lock's declares (spec §6, §16 row 18): what each pack declares, written at each init and update; the rules
// the engine holds a pack's declarations to; and bonsai check running from the lock alone (spec §5: "CI needs no
// pack"), with an empty pack cache and no network.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// linkBase links a project to the declaring fixture pack (testpack.DeclaringPack), at its first commit.
func (e *env) linkBase(t *testing.T, name string) (root, source string) {
	t.Helper()
	src, shas := testpack.DeclaringPack(t, e.tmp)
	root = testpack.Project(t, e.tmp, name)
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	return root, src
}

// A pack's declarations reach the lock's declares at a link, each kind under its key in format.DeclaresKeys' order,
// every field in its format's order; they read back as the pack declared them; and a pack that declares nothing has {}.
func TestDeclaresInTheLock(t *testing.T) {
	e := setup(t)
	root, _ := e.linkBase(t, "declares")
	l, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(l.Packs[0].Declares.Keys(), " "); got != "lanes documents labels protected deny" {
		t.Errorf("declares keys %q", got)
	}
	d, err := l.Packs[0].Declared()
	if err != nil {
		t.Fatal(err)
	}
	lanes, _ := format.ReadLanes([]byte(testpack.BaseLanesYAML))
	labels, _ := format.ReadLabels([]byte(testpack.BaseLabelsYAML))
	pack, _ := format.ReadPack([]byte(testpack.BasePackYAML))
	want := &format.Declares{Lanes: lanes.Lanes, Documents: pack.Documents, Labels: labels, Protected: pack.Protected,
		Hooks: pack.Hooks, Deny: pack.Deny}
	got, _ := d.Object()
	wantDoc, _ := want.Object()
	if !schema.Equal(got, wantDoc) {
		t.Errorf("declares read back as\n%s\nwant\n%s", schema.Show(got), schema.Show(wantDoc))
	}
	// Written in each part's schema order, at every depth.
	lanesDoc, _ := l.Packs[0].Declares.Get("lanes")
	if msgs := schema.CheckOrder(format.MustLookup("lanes").Schema(), schema.Object{{Key: "format", Value: "bonsai.lanes/1"},
		{Key: "lanes", Value: lanesDoc}}); len(msgs) > 0 {
		t.Errorf("lanes out of order: %v", msgs)
	}
	// The test pack declares nothing but its hook line and deny rule; a pack that declares nothing at all has {}.
	q := testpack.Project(t, e.tmp, "quiet")
	src, shas := testpack.QuietPack(t, e.tmp)
	e.apply(t, q, Request{Command: "init", Init: &InitValues{Name: "quiet", Source: src, Ref: shas[0]}})
	if !strings.Contains(read(t, q, workspace.LockFile), `"declares": {}`) {
		t.Errorf("a pack that declares nothing:\n%s", read(t, q, workspace.LockFile))
	}
	// An update to the same commit changes no byte of the lock.
	before := read(t, root, workspace.LockFile)
	if p := e.apply(t, root, Request{}); p.LockWrite || read(t, root, workspace.LockFile) != before {
		t.Errorf("an update with nothing new wrote the lock")
	}
}

// The rules a pack's declarations are held to (declares.go): each refused at the link as bad-pack (exit 2), or, for
// two packs, packs-overlap, with nothing written.
func TestDeclaresRefused(t *testing.T) {
	e := setup(t)
	cases := []struct {
		name, labels, packYAML, word, words string
	}{
		{"labels in another pack's namespace", strings.Replace(strings.ReplaceAll(testpack.BaseLabelsYAML, "name: bonsai.", "name: workflow."),
			"namespace: bonsai", "namespace: workflow", 1), testpack.BasePackYAML, "bad-pack", "a pack's labels use its own id"},
		{"a label outside its namespace", strings.Replace(testpack.BaseLabelsYAML, "name: bonsai.branch", "name: other.branch", 1),
			testpack.BasePackYAML, "bad-pack", "is not in its namespace bonsai"},
		{"a label defined twice", strings.Replace(testpack.BaseLabelsYAML, "name: bonsai.branch", "name: bonsai.allows", 1),
			testpack.BasePackYAML, "bad-pack", "defines the label \"bonsai.allows\" twice"},
		{"a kind named like Bonsai's own", testpack.BaseLabelsYAML, strings.Replace(testpack.BasePackYAML, "kind: plan", "kind: task", 1),
			"bad-pack", "takes the name of Bonsai's own task"},
		{"a kind with both path and file", testpack.BaseLabelsYAML, strings.Replace(testpack.BasePackYAML, "path: null\n    file: work/bugs.md",
			"path: work/bugs\n    file: work/bugs.md", 1), "bad-pack", "both path and file"},
		{"an id pattern Go does not read", testpack.BaseLabelsYAML, strings.Replace(testpack.BasePackYAML, `"^P-T-[0-9]{4,6}$"`, `"^P-(T"`, 1),
			"bad-pack", "is not a regular expression Bonsai reads"},
	}
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src, shas := testpack.Fixture(t, e.tmp, "bad-declares-"+string(rune('a'+i)), map[string]string{
				".claude-plugin/plugin.json":    pluginJSON("base"),
				"bonsai/pack.yaml":              c.packYAML,
				"bonsai/labels.yaml":            c.labels,
				"bonsai/block.md":               "x\n",
				"bonsai/files/session-start.md": "x\n",
			})
			root := testpack.Project(t, e.tmp, "bad-declares-project-"+string(rune('a'+i)))
			_, err := e.try(root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
			var ee *Error
			if !errors.As(err, &ee) || ee.Code != c.word || ee.Exit != ExitInput || !strings.Contains(ee.What, c.words) || ee.Next == "" {
				t.Fatalf("want %s holding %q, got %v", c.word, c.words, err)
			}
		})
	}
	// Two packs declaring one document kind.
	a, sa := testpack.DeclaringPack(t, e.tmp)
	other := strings.NewReplacer("id: base", "id: other", "work/protocols/session-start.md", "work/other.md").Replace(testpack.BasePackYAML)
	b, sb := testpack.Fixture(t, e.tmp, "other", map[string]string{
		".claude-plugin/plugin.json":    pluginJSON("other"),
		"bonsai/pack.yaml":              other,
		"bonsai/block.md":               "x\n",
		"bonsai/files/session-start.md": "x\n",
	})
	root := testpack.Project(t, e.tmp, "two-kinds")
	writeFile(t, root, "bonsai.yaml", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\npacks:\n"+
		"  - id: base\n    source: \""+filepath.ToSlash(a)+"\"\n    ref: \""+sa[0]+"\"\n"+
		"  - id: other\n    source: \""+filepath.ToSlash(b)+"\"\n    ref: \""+sb[0]+"\"\n")
	_, err := e.try(root, Request{Command: "init"})
	var ee *Error
	if !errors.As(err, &ee) || ee.Code != "packs-overlap" || !strings.Contains(ee.What, "two packs declare the document kind plan (base and other)") {
		t.Fatalf("two packs, one kind: %v", err)
	}
}

// A lock whose declares Bonsai does not read (edited by hand) is refused as the lock: update exits 4 (bad-lock) and
// check reports it, each naming git checkout of the lock as the step.
func TestBadDeclaresIsABadLock(t *testing.T) {
	e := setup(t)
	root, _ := e.linkBase(t, "bad-lock")
	writeFile(t, root, workspace.LockFile, strings.Replace(read(t, root, workspace.LockFile), `"close": "agent"`, `"close": "nobody"`, 1))
	_, err := e.try(root, Request{})
	var ee *Error
	if !errors.As(err, &ee) || ee.Code != "bad-lock" || ee.Exit != ExitState || !strings.Contains(ee.What, "declares.lanes") ||
		!strings.Contains(ee.Next, "git checkout -- .bonsai/lock.json") {
		t.Fatalf("update with a bad declares: %v", err)
	}
	r, err := Check(root, e.home)
	if err != nil || len(r.Findings) == 0 || r.Findings[0].Code != "lock" {
		t.Fatalf("check with a bad declares: %v %+v", err, r)
	}
}

// Check 10 of 5.1's end (plan-5): bonsai check in a linked project with an empty pack cache and no network gives the
// same findings as with them. The pack's source is moved away (no fetch could reach it) and the home's cache
// emptied; findings and warnings are compared, a clean project and one with every kind of edit check finds.
func TestCheckOfflineFromTheLock(t *testing.T) {
	e := setup(t)
	root, src := e.linkBase(t, "offline")
	demo := testpack.Project(t, e.tmp, "offline-demo")
	e.link(t, demo, e.pack.D, "ledger.json")
	type run struct{ findings, warnings string }
	checkAll := func() map[string]run {
		out := map[string]run{}
		for _, r := range []string{root, demo} {
			res, err := Check(r, e.home)
			if err != nil {
				t.Fatal(err)
			}
			var f, w []string
			for _, x := range res.Findings {
				f = append(f, x.Code+" "+x.File+": "+x.Message)
			}
			for _, x := range res.Warnings {
				w = append(w, x.Code+" "+x.File+": "+x.Message)
			}
			out[r] = run{strings.Join(f, "\n"), strings.Join(w, "\n")}
		}
		return out
	}
	edit := func() {
		// An edited pack file, a deny rule of the pack's taken out, the never_edit rule taken out, a hook line edited.
		writeFile(t, demo, "demo/guide.md", "edited\n")
		writeFile(t, root, SettingsFile, strings.Replace(read(t, root, SettingsFile), `"Read(secrets/**)"`, `"Read(secret/**)"`, 1))
		writeFile(t, demo, SettingsFile, strings.Replace(strings.Replace(read(t, demo, SettingsFile), `"Edit(ledger.json)",`, "", 1),
			"echo demo hook D", "echo demo hook X", 1))
	}
	offline := func() {
		if err := os.RemoveAll(filepath.Join(e.home, "cache")); err != nil {
			t.Fatal(err)
		}
		for _, s := range []string{src, e.pack.Source} {
			if err := os.Rename(s, s+".away"); err != nil {
				t.Fatal(err)
			}
		}
	}
	clean := checkAll()
	for r, c := range clean {
		if c.findings != "" || c.warnings != "" {
			t.Errorf("%s, linked and untouched: %+v", filepath.Base(r), c)
		}
	}
	edit()
	online := checkAll()
	if online[root].findings == "" || !strings.HasPrefix(online[root].findings, "changed .claude/settings.json") ||
		len(strings.Split(online[demo].findings, "\n")) != 2 {
		t.Fatalf("online, the edits are not all found: %+v", online)
	}
	offline()
	if got := checkAll(); got[root] != online[root] || got[demo] != online[demo] {
		t.Errorf("offline differs from online:\n%+v\nwant\n%+v", got, online)
	}
	if _, err := os.Stat(filepath.Join(e.home, "cache")); !os.IsNotExist(err) {
		t.Errorf("check wrote the pack cache")
	}
}

// A lock written before step 5.1.5 (no path, no hook lines in its declares) still has its settings judged with the
// pack from this machine's cache, and, with none there, a warning, as before: never a false finding.
func TestCheckOldLockUsesTheCache(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "old-lock-check")
	e.link(t, root, e.pack.A)
	withoutPath(t, root)
	writeFile(t, root, workspace.LockFile, strings.Replace(read(t, root, workspace.LockFile),
		read(t, root, workspace.LockFile)[strings.Index(read(t, root, workspace.LockFile), `"declares": {`):strings.Index(read(t, root, workspace.LockFile), `"files"`)],
		"\"declares\": {}\n    }\n  ],\n  ", 1))
	if r, err := Check(root, e.home); err != nil || len(r.Findings) != 0 || len(r.Warnings) != 0 {
		t.Fatalf("an old lock, the pack in the cache: %v %+v %+v", err, r.Findings, r.Warnings)
	}
	if err := os.RemoveAll(filepath.Join(e.home, "cache")); err != nil {
		t.Fatal(err)
	}
	if r, err := Check(root, e.home); err != nil || len(r.Findings) != 0 || len(r.Warnings) != 1 || r.Warnings[0].Code != "cache" {
		t.Fatalf("an old lock, no cache: %v %+v %+v", err, r.Findings, r.Warnings)
	}
}
