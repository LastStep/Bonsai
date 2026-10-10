package clean

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A kind's folder that is a junction (mklink /J, which needs no privilege) is never followed on Windows: nothing
// behind it is cleaned. A file link needs a privilege or developer mode there, so files linked one by one are held
// on Linux (TestEveryKind).
func TestJunctionIsNotFollowed(t *testing.T) {
	l := project(t, allRules)
	outside := filepath.Join(filepath.Dir(l.Main), "elsewhere")
	taskFile(t, l, "T-0001", "done")
	elsewhere := projectAt(t, outside)
	ladderFile(t, elsewhere, "T-0001.json", ptr("T-0001"), daysAgo(20))
	ladder := filepath.Join(l.Main, ".bonsai", "local", "ladder")
	if err := os.Remove(ladder); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", ladder, filepath.Join(outside, ".bonsai", "local", "ladder")).CombinedOutput(); err != nil {
		t.Fatalf("mklink /J: %v %s", err, out)
	}
	if d := Files(opts(l), Ladder); len(d.Cleaned) != 0 || len(d.Errs) != 0 {
		t.Errorf("cleaned %v, errors %v", d.Cleaned, d.Errs)
	}
	if _, err := os.Stat(filepath.Join(outside, ".bonsai", "local", "ladder", "T-0001.json")); err != nil {
		t.Errorf("the result behind the junction: %v", err)
	}
}
