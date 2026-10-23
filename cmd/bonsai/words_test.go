package main

import (
	"strings"
	"testing"
)

// T-5729: update leads with what it recorded when it recorded settings for the checks looser than the ones this
// machine accepted, not with "nothing to change"; an update with nothing at all to do keeps its old lead.
func TestUpdateLeadsWithWhatItRecorded(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	linked := c.read("bonsai.yaml")
	lists := `".claude/**", "bonsai.yaml", ".bonsai/lock.json"`
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 || !strings.Contains(flat(out), "nothing to change") {
		t.Fatalf("an update with nothing to do: exit %d\n%s%s", code, out, errOut)
	}
	c.write("bonsai.yaml", strings.Replace(linked, "protected: ["+lists+"]", `protected: ["bonsai.yaml", ".bonsai/lock.json"]`, 1))
	code, out, errOut := c.run("", "update", "--yes")
	first := strings.Join(strings.SplitN(out, "\n", 3)[:2], " ")
	if code != 0 || strings.Contains(first, "nothing to change") || !strings.Contains(first, "recorded") {
		t.Errorf("update that recorded looser settings, lead line %q (exit %d)\n%s%s", first, code, out, errOut)
	}
}
