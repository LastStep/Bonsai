package main

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// init links several packs in one first link (BUG-122): --source and --ref once per pack, the n-th --ref (and --path,
// when given) going with the n-th --source. Two packs with no hook lines of their own link with --yes alone, both in
// bonsai.yaml and the lock; --ref or --path given another number of times than --source is refused, nothing written.
func TestInitSeveralPacks(t *testing.T) {
	c := newCLI(t)
	quiet, qc := testpack.QuietPack(t, c.tmp)
	decl, dc := testpack.DeclaringPack(t, c.tmp)
	for _, bad := range [][]string{
		{"--source", quiet, "--ref", qc[0], "--source", decl},
		{"--source", quiet, "--ref", qc[0], "--source", decl, "--ref", dc[0], "--path", "x"},
	} {
		code, out, errOut := c.run("", append([]string{"init", "--name", "demo", "--yes"}, bad...)...)
		if code != 2 || !strings.Contains(out+errOut, "--source") {
			t.Errorf("init %q: exit %d\n%s%s", bad, code, out, errOut)
		}
		if files := c.files(); len(files) != 0 && strings.Contains(strings.Join(files, " "), "bonsai.yaml") {
			t.Errorf("a refused init wrote: %q", files)
		}
	}
	code, out, errOut := c.run("", "init", "--name", "demo", "--source", quiet, "--ref", qc[0], "--source", decl, "--ref", dc[0], "--yes")
	if code != 0 {
		t.Fatalf("init with two packs: exit %d\n%s%s", code, out, errOut)
	}
	cfg, lock := c.read("bonsai.yaml"), c.read(".bonsai/lock.json")
	for _, id := range []string{"quiet-pack", "base"} {
		if !strings.Contains(cfg, "  - id: "+id+" ") || !strings.Contains(lock, `"id": "`+id+`"`) {
			t.Errorf("the pack %s is not linked:\n%s\n%s", id, cfg, lock)
		}
	}
	if strings.Index(cfg, "id: quiet-pack") > strings.Index(cfg, "id: base") {
		t.Errorf("the packs are not in the order given:\n%s", cfg)
	}
	// Run again with the same values: nothing to change; with another pack's values: values-differ.
	if code, out, errOut := c.run("", "init", "--name", "demo", "--source", quiet, "--ref", qc[0], "--source", decl, "--ref", dc[0], "--yes"); code != 0 {
		t.Errorf("init again: exit %d\n%s%s", code, out, errOut)
	}
	if code, out, errOut := c.run("", "init", "--name", "demo", "--source", quiet, "--ref", qc[0], "--yes"); code != 2 ||
		!strings.Contains(out+errOut, "values-differ") && !strings.Contains(out+errOut, "says otherwise") {
		t.Errorf("init with one of the two: exit %d\n%s%s", code, out, errOut)
	}
}
