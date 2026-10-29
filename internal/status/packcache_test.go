package status

// The packs' newer releases from the home's cache (T-6009's plan, steps 4 and 5; formats set 13): status --json's
// packs[] carry ref and newer; plain status reads only cache/packs.json, never a source (its Tags fails the test if
// called); --full reads each source once and writes the cache, which plain status then shows.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// cachedProject links the test pack at the tag v0.1.0 and gives the project's root, its home and the pack's source.
func cachedProject(t *testing.T) (root, home, source string) {
	t.Helper()
	tmp := t.TempDir()
	if r, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = r
	}
	home = testpack.Isolate(t, tmp)
	pack := testpack.Build(t, tmp)
	testpack.Tag(t, pack.Source, "v0.1.0", pack.A)
	root = testpack.Project(t, tmp, "cached")
	p, err := engine.Build(engine.Request{Command: "init", Dir: root, Home: home, Version: "test", AllowExec: true,
		Init: &engine.InitValues{Name: "cached", Source: pack.Source, Ref: "v0.1.0"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(p); err != nil {
		t.Fatal(err)
	}
	return root, home, pack.Source
}

// writePackCache writes cache/packs.json by hand: one source's tags read at read.
func writePackCache(t *testing.T, home, source string, tags []string, read string) {
	t.Helper()
	c := engine.PackCache{Sources: []engine.PackSourceRecord{{Source: source, Tags: tags, Read: &read}}}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(home, filepath.FromSlash(engine.PackCacheFile))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

// noTags fails the test when a source is read.
func noTags(t *testing.T) func(string) ([]string, error) {
	return func(source string) ([]string, error) {
		t.Errorf("plain status read the source %s", source)
		return nil, nil
	}
}

func firstPack(t *testing.T, doc schema.Object) schema.Object {
	t.Helper()
	pv, _ := doc.Get("packs")
	list, _ := pv.([]any)
	if len(list) != 1 {
		t.Fatalf("packs = %s, want one", schema.Show(pv))
	}
	return list[0].(schema.Object)
}

func TestStatusShowsTheCacheOffline(t *testing.T) {
	root, home, source := cachedProject(t)
	// No record: ref from bonsai.yaml, newer null.
	doc, code := BuildWith(root, "dev", Options{Tags: noTags(t)})
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	p := firstPack(t, doc)
	if p.String("ref") != "v0.1.0" {
		t.Errorf("ref = %q, want bonsai.yaml's v0.1.0", p.String("ref"))
	}
	if n, _ := p.Get("newer"); n != nil {
		t.Errorf("newer with no record = %s, want null", schema.Show(n))
	}
	now := time.Now().UTC().Format(time.RFC3339)
	cases := []struct {
		name, read string
		tags       []string
		want       string
	}{
		{"newer", now, []string{"v0.1.0", "v0.2.0", "v0.3.0", "other-v9.0.0"}, `{"tag":"v0.3.0","read":"` + now + `","state":"newer"}`},
		{"current", now, []string{"v0.1.0"}, `{"tag":"v0.1.0","read":"` + now + `","state":"current"}`},
		{"stale by age", "2020-01-01T00:00:00Z", []string{"v0.1.0"}, `{"tag":"v0.1.0","read":"2020-01-01T00:00:00Z","state":"stale"}`},
		{"stale below the lock", now, []string{"v0.0.9"}, `{"tag":"v0.1.0","read":"` + now + `","state":"stale"}`},
		{"newer though old", "2020-01-01T00:00:00Z", []string{"v0.1.0", "v0.2.0"}, `{"tag":"v0.2.0","read":"2020-01-01T00:00:00Z","state":"newer"}`},
	}
	for _, c := range cases {
		writePackCache(t, home, source, c.tags, c.read)
		doc, _ := BuildWith(root, "dev", Options{Tags: noTags(t)})
		checkShape(t, doc, false)
		n, _ := firstPack(t, doc).Get("newer")
		if got := schema.Show(n); got != c.want {
			t.Errorf("%s: newer = %s, want %s", c.name, got, c.want)
		}
		if c.name == "newer" && !strings.Contains(flat(Text(doc)), "v0.3.0 is released") {
			t.Errorf("the text names no newer release:\n%s", Text(doc))
		}
	}
	// A record of another source only: null.
	writePackCache(t, home, "https://example.org/other.git", []string{"v9.0.0"}, now)
	doc, _ = BuildWith(root, "dev", Options{Tags: noTags(t)})
	if n, _ := firstPack(t, doc).Get("newer"); n != nil {
		t.Errorf("newer from another source's record = %s, want null", schema.Show(n))
	}
}

// --full reads each source once and writes the cache; a failed read after a good one keeps the good one; plain status
// then shows it without reading.
func TestStatusFullWritesTheCache(t *testing.T) {
	root, home, source := cachedProject(t)
	calls := 0
	remote := func(s string) ([]string, error) {
		calls++
		if s != source {
			t.Errorf("read %s, want %s", s, source)
		}
		return []string{"v0.1.0", "v0.2.0"}, nil
	}
	if _, code := BuildWith(root, "dev", Options{Full: true, Tags: remote, Plugins: fakeCLI{}}); code != ExitOK {
		t.Fatalf("exit %d", code)
	}
	if calls != 1 {
		t.Errorf("--full read the source %d times, want once", calls)
	}
	c, err := engine.LoadPackCache(home)
	if err != nil || c.Record(source) == nil || strings.Join(c.Record(source).Tags, ",") != "v0.1.0,v0.2.0" {
		t.Fatalf("cache after --full: %+v, %v", c, err)
	}
	doc, _ := BuildWith(root, "dev", Options{Tags: noTags(t)})
	if n, _ := firstPack(t, doc).Get("newer"); n == nil || n.(schema.Object).String("state") != engine.PackNewer {
		t.Errorf("plain status after --full: newer = %s", schema.Show(n))
	}
	read := *c.Record(source).Read
	failing := func(string) ([]string, error) { return nil, os.ErrDeadlineExceeded }
	BuildWith(root, "dev", Options{Full: true, Tags: failing, Plugins: fakeCLI{}})
	c, _ = engine.LoadPackCache(home)
	if r := c.Record(source); r == nil || r.Failed == nil || r.Read == nil || *r.Read != read || len(r.Tags) != 2 {
		t.Errorf("a failed read after a good one: %+v, want the good one kept and the failure recorded", r)
	}
	doc, _ = BuildWith(root, "dev", Options{Tags: noTags(t)})
	if n, _ := firstPack(t, doc).Get("newer"); n == nil || n.(schema.Object).String("read") != read {
		t.Errorf("after a failed read: newer = %s, want the good read's", schema.Show(n))
	}
}
