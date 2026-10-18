package status

// status --json's bonsai_release (set 10; plan-5 5.6.5): at the document's end, always filled on exit 0, from the
// home's record of Bonsai's releases alone, which --full reads again first when a read is due; status's text line;
// --line's marker.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
)

// releaseTags is a fake read of Bonsai's tags that counts its calls.
func releaseTags(calls *int, tags ...string) func(string, time.Duration) ([]string, error) {
	return func(source string, timeout time.Duration) ([]string, error) {
		*calls++
		if source != engine.ReleaseSource || timeout != engine.ReleaseTimeout {
			return nil, errors.New("read " + source + " with " + timeout.String())
		}
		return tags, nil
	}
}

func TestStatusRelease(t *testing.T) {
	root, home := project(t, cfg)
	doc, code := Build(real(t, root), "1.0.0")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	keys := doc.Keys()
	if keys[len(keys)-1] != "bonsai_release" {
		t.Errorf("status --json ends with %s, want bonsai_release (set 10)", keys[len(keys)-1])
	}
	rel, _ := doc.Get("bonsai_release")
	if got := schema.Show(rel); got != `{"version":"1.0.0","newest":null,"state":"unknown","read":null,`+
		`"why":"Bonsai's releases have not been read on this machine yet: bonsai status --full reads them","lines":null}` {
		t.Errorf("never read: %s", got)
	}
	if !strings.Contains(flat(Text(doc)), "newer releases unknown") {
		t.Errorf("the text has no release line:\n%s", Text(doc))
	}
	// --full reads, the default and a second --full within the day do not.
	calls := 0
	o := Options{Full: true, Tags: func(string) ([]string, error) { return nil, errors.New("no pack read here") },
		Release: releaseTags(&calls, "v0.4.3", "base-v2.0.0", "v1.0.0", "v1.0.1")}
	doc, _ = BuildWith(real(t, root), "1.0.0", o)
	checkShape(t, doc, false)
	rel, _ = doc.Get("bonsai_release")
	r := rel.(schema.Object)
	if calls != 1 || r.String("state") != engine.ReleaseNewer || r.String("newest") != "1.0.1" || !strings.Contains(schema.Show(r),
		`"linux":["cd \"$(mktemp -d)\"","curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/bonsai_1.0.1_linux_`) {
		t.Errorf("after --full's read (%d calls): %s", calls, schema.Show(rel))
	}
	if txt := flat(Text(doc)); !strings.Contains(txt, "v1.0.1 is released") || !strings.Contains(txt, "bonsai_release.lines") {
		t.Errorf("the text's release line:\n%s", txt)
	}
	doc, _ = Build(real(t, root), "1.0.0")
	if rel, _ := doc.Get("bonsai_release"); rel.(schema.Object).String("state") != engine.ReleaseNewer {
		t.Errorf("the default did not read the record: %s", schema.Show(rel))
	}
	if BuildWith(real(t, root), "1.0.0", o); calls != 1 {
		t.Errorf("--full read again within the day: %d calls", calls)
	}
	// The installed copy's version is the one compared: current.
	if err := os.WriteFile(filepath.Join(home, engine.InstallFile), []byte(`{"path": "/usr/local/bin/bonsai", "version": "1.0.1", "sha256": ""}`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, _ = Build(real(t, root), "dev")
	if rel, _ := doc.Get("bonsai_release"); rel.(schema.Object).String("state") != engine.ReleaseCurrent || !strings.Contains(flat(Text(doc)), "v1.0.1, no newer release") {
		t.Errorf("installed 1.0.1: %s\n%s", schema.Show(rel), Text(doc))
	}
	if err := os.Remove(filepath.Join(home, engine.InstallFile)); err != nil {
		t.Fatal(err)
	}
	doc, _ = Build(real(t, root), "dev")
	if rel, _ := doc.Get("bonsai_release"); rel.(schema.Object).String("state") != engine.ReleaseNotRelease ||
		!strings.Contains(flat(Text(doc)), "this bonsai is a build from source, so nothing is compared; the newest release is v1.0.1") {
		t.Errorf("dev: %s\n%s", schema.Show(rel), Text(doc))
	}
	// --line's marker, from the record alone.
	if l := BuildLineWith(root, "1.0.0"); l.Release == nil || *l.Release != "1.0.1" || !strings.HasSuffix(l.Line, " bonsai:v1.0.1") {
		t.Errorf("the line: %+v", l)
	}
	if l := BuildLineWith(root, "1.0.1"); l.Release != nil || strings.Contains(l.Line, "bonsai:") {
		t.Errorf("the line of a current bonsai: %q", l.Line)
	}
	if l := BuildLine(root); l.Release != nil {
		t.Errorf("BuildLine compared no version, yet marked one: %q", l.Line)
	}
}

// A home where the record cannot be written: --full's answer is still given, one line says it was not kept.
func TestStatusReleaseNotKept(t *testing.T) {
	root, home := project(t, cfg)
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "cache"), []byte("a file where the folder goes"), 0o644); err != nil {
		t.Fatal(err)
	}
	calls, said := 0, []string{}
	doc, code := BuildWith(real(t, root), "1.0.0", Options{Full: true, Tags: func(string) ([]string, error) { return nil, nil },
		Release: releaseTags(&calls, "v1.0.1"), NotKept: func(l string) { said = append(said, l) }})
	rel, _ := doc.Get("bonsai_release")
	if code != ExitOK || rel.(schema.Object).String("state") != engine.ReleaseNewer || len(said) != 1 ||
		!strings.HasPrefix(said[0], "Bonsai's release read was not kept in "+engine.ReleasePath(home)) {
		t.Errorf("exit %d, %s, said %q", code, schema.Show(rel), said)
	}
}
