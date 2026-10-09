package engine

// The lock's path (formats set 4; step 5.1.5): every lock written records each pack's folder; a lock written before
// set 4 reads it as unknown; consent is judged against the folder as well as the content (step 5.1.1's verifier, B1's
// rest, and step 5.1.3's verifier, F7); and the plugin step's "already installed" compares the folder too.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// lockPaths gives each locked pack's path as the lock file holds it: its JSON, or "-" when the pack has none.
func lockPaths(t *testing.T, root string) string {
	t.Helper()
	l, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, p := range l.Packs {
		switch {
		case !p.PathSet:
			out = append(out, p.ID+"=-")
		case p.Path == "":
			out = append(out, p.ID+"=null")
		default:
			out = append(out, p.ID+"="+p.Path)
		}
	}
	return strings.Join(out, " ")
}

// withoutPath writes the lock back as a Bonsai before formats set 4 wrote it: no path in any pack.
func withoutPath(t *testing.T, root string) {
	t.Helper()
	lock := read(t, root, workspace.LockFile)
	var kept []string
	for _, l := range strings.Split(lock, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), `"path": `) {
			// The pack's last field: the line before loses its comma.
			kept[len(kept)-1] = strings.TrimSuffix(kept[len(kept)-1], ",")
			continue
		}
		kept = append(kept, l)
	}
	out := strings.Join(kept, "\n")
	if out == lock || strings.Contains(out, `"path"`) {
		t.Fatalf("the lock held no path to take out:\n%s", lock)
	}
	writeFile(t, root, workspace.LockFile, out)
}

// twinPack is a pack repository whose two folders, a/ and b/, are the same pack with the same files; b/ also holds
// a git submodule named hooks (a gitlink: no symbolic link, so it runs on Windows too), which a plugin's content hash
// does not see (it hashes blobs) but Claude Code would read as the plugin's hooks folder.
func twinPack(t *testing.T, tmp string) (source, commit string) {
	t.Helper()
	work := filepath.Join(tmp, "twin-sub-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	testpack.Git(t, work, "init", "-q")
	for _, dir := range []string{"a", "b"} {
		writeFile(t, work, dir+"/.claude-plugin/plugin.json", pluginJSON("twin"))
		writeFile(t, work, dir+"/agents/helper.md", "Help.\n")
		writeFile(t, work, dir+"/bonsai/pack.yaml", packYAML("twin", " []", " []"))
	}
	testpack.Git(t, work, "add", "-A")
	testpack.Git(t, work, "commit", "-q", "-m", "the pack in two folders")
	first := testpack.Git(t, work, "rev-parse", "HEAD")
	testpack.Git(t, work, "update-index", "--add", "--cacheinfo", "160000,"+first+",b/hooks")
	testpack.Git(t, work, "commit", "-q", "-m", "a submodule named hooks in b")
	commit = testpack.Git(t, work, "rev-parse", "HEAD")
	source = filepath.Join(tmp, "twin-sub.git")
	testpack.Git(t, tmp, "clone", "-q", "--bare", "--", work, source)
	return source, commit
}

// Every lock written records each pack's folder: null at the repository's root, the folder as bonsai.yaml names it.
func TestLockRecordsEachPacksFolder(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "root-pack")
	e.link(t, root, e.pack.A)
	if got := lockPaths(t, root); got != "demo-pack=null" {
		t.Errorf("a pack at its repository's root: %s", got)
	}
	if !strings.Contains(read(t, root, workspace.LockFile), `"path": null`) {
		t.Errorf("the lock does not write the root's path as null:\n%s", read(t, root, workspace.LockFile))
	}
	src, commit := twinPack(t, e.tmp)
	sub := testpack.Project(t, e.tmp, "sub-pack")
	e.apply(t, sub, Request{Command: "init", Init: &InitValues{Name: "twin", Source: src, Ref: commit, Path: "a"}})
	if got := lockPaths(t, sub); got != "twin=a" {
		t.Errorf("a pack in a folder: %s", got)
	}
}

// A lock written before formats set 4 has no path. Its pack's folder reads as unknown, not as the root: a pack in a
// folder is judged by its content hash at bonsai.yaml's folder, as before, so it does not look moved after an
// upgrade; the next update writes the path and nothing else. A folder changed in bonsai.yaml since such a lock is
// still found by the content hash when the content differs (TestFolderChangeIsCode's case).
func TestOldLockFolderIsUnknown(t *testing.T) {
	e := setup(t)
	src, commit := twinPack(t, e.tmp)
	for _, c := range []struct {
		name, ref, folder string
		values            *InitValues
	}{
		{"a pack in a folder", commit, "a", &InitValues{Name: "twin", Source: src, Ref: commit, Path: "a"}},
		{"a pack at its root", e.pack.F, "", &InitValues{Name: "demo", Source: e.pack.Source, Ref: e.pack.F}},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := testpack.Project(t, e.tmp, "old-lock-"+strings.ReplaceAll(c.name, " ", "-"))
			e.apply(t, root, Request{Command: "init", Init: c.values, AllowExec: true})
			withoutPath(t, root)
			if got := lockPaths(t, root); !strings.HasSuffix(got, "=-") {
				t.Fatalf("the old lock reads its path as %s, want unknown", got)
			}
			p := e.plan(t, root, Request{})
			if len(p.Unverified) > 0 || p.NeedsExec() || len(p.Conflicts) > 0 {
				t.Fatalf("an old lock's pack looks moved: unverified %v, runs code %s\n%s", p.Unverified, items(p), p.Preview(false))
			}
			for _, f := range p.Files {
				if f.Writes() {
					t.Errorf("an old lock's update writes %s (%s)", f.Path, f.Result)
				}
			}
			if !p.LockWrite {
				t.Errorf("the update does not write the lock's path")
			}
			if err := Apply(p); err != nil {
				t.Fatal(err)
			}
			want := "=" + c.folder
			if c.folder == "" {
				want = "=null"
			}
			if got := lockPaths(t, root); !strings.HasSuffix(got, want) {
				t.Errorf("after the update the lock's path is %s, want %s", got, want)
			}
		})
	}
}

// The submodule repro (step 5.1.1's verifier, B1's rest): two folders with the same blobs, one with a submodule named
// hooks. The content hash is the same for both, so a folder change from a to b used to pass on --yes. With the
// lock's path it has no baseline: the pack is unverified, its plugin's code parts count as at a first link (the
// submodule named among them), and --yes alone writes nothing.
func TestFolderChangeWithTheSameBlobs(t *testing.T) {
	e := setup(t)
	src, commit := twinPack(t, e.tmp)
	c := cache{home: e.home}
	if _, err := c.fetch(src, commit); err != nil {
		t.Fatal(err)
	}
	a, errA := c.packAt(workspace.PackRef{ID: "twin", Source: src, Path: "a"}, commit)
	b, errB := c.packAt(workspace.PackRef{ID: "twin", Source: src, Path: "b"}, commit)
	if errA != nil || errB != nil || a.SHA256 != b.SHA256 || len(a.Code) != 0 || len(b.Code) != 1 {
		t.Fatalf("the twin folders: %v %v, hashes %s %s, code %v %v", errA, errB, a.SHA256, b.SHA256, a.Code, b.Code)
	}
	root := testpack.Project(t, e.tmp, "twin-sub")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "twin", Source: src, Ref: commit, Path: "a"}})
	writeFile(t, root, "bonsai.yaml", strings.Replace(read(t, root, "bonsai.yaml"), `path: "a"`, `path: "b"`, 1))
	before := snapshot(t, root)
	p := e.plan(t, root, Request{})
	if strings.Join(p.Unverified, ",") != "twin" {
		t.Errorf("unverified %v, want twin", p.Unverified)
	}
	if got := items(p); got != "plugin add hooks (twin)" {
		t.Errorf("runs code %q, want the submodule named\n%s", got, p.Preview(false))
	}
	if !strings.Contains(p.Preview(false), "Unverified: the pack twin's folder in bonsai.yaml is not the lock's") {
		t.Errorf("the preview does not say why the pack is unverified:\n%s", p.Preview(false))
	}
	if err := Apply(p); err == nil {
		t.Fatal("a folder change was applied on --yes alone")
	}
	sameSnapshot(t, "a refused folder change", before, snapshot(t, root))
	r, err := Check(root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range r.Findings {
		found = found || (f.Code == "packs" && strings.Contains(f.Message, "from the folder b of its repository, the lock from a"))
	}
	if !found {
		t.Errorf("check does not find the folder change: %+v", r.Findings)
	}
	p = e.apply(t, root, Request{AllowExec: true})
	if got := lockPaths(t, root); got != "twin=b" {
		t.Errorf("with --allow-exec the lock's path is %s, want b", got)
	}
}

// The plugin step's "already installed" compares the folder as well as the commit: a pack's marketplace name pins
// its folder, so a plugin installed from the old folder at the same commit is another plugin, and a plugin that
// carries code waits for --allow-exec. A pack at its repository's root keeps the name its commit alone gives.
func TestPluginStepComparesTheFolder(t *testing.T) {
	e := setup(t)
	if MarketplaceName("demo", []string{Pin(e.pack.A, "")}) != MarketplaceName("demo", []string{e.pack.A}) {
		t.Errorf("a pack at its root changed its marketplace's name")
	}
	if MarketplaceName("demo", []string{Pin(e.pack.A, "a")}) == MarketplaceName("demo", []string{Pin(e.pack.A, "b")}) {
		t.Errorf("two folders at one commit share a marketplace")
	}
	src, commit := twinPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "twin-plugin")
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "twin", Source: src, Ref: commit, Path: "b"}, AllowExec: true})
	atB := PluginID("twin", MarketplaceName("twin", []string{Pin(commit, "b")}))
	atA := PluginID("twin", MarketplaceName("twin", []string{Pin(commit, "a")}))
	p := e.plan(t, root, Request{})
	consent := p.PluginConsent("bonsai update --allow-exec")
	if len(consent.Code["twin"]) == 0 {
		t.Fatalf("the plugin at b carries no code: %+v", consent)
	}
	for _, c := range []struct {
		installed, want string
	}{{atB, "installed"}, {atA, "waiting"}} {
		f := &fakeCLI{list: []InstalledPlugin{{ID: c.installed, Version: commit[:12], Scope: "project", Enabled: true, ProjectPath: root}}}
		got := InstallPlugins(root, p.Config, p.NewLock(), f, consent)
		if len(got) != 1 || got[0].Result != c.want || got[0].Plugin != atB {
			t.Errorf("installed %s: %+v, want %s", c.installed, got, c.want)
		}
	}
}
