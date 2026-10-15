package main

import (
	"bytes"
	"strings"
	"testing"
)

// update's help says what a run without --yes still does (V1's should-fix 2, 16 Oct): its fetch fills the Bonsai
// home's cache, and with nothing to change it records this machine's record of the checkout and asks Claude Code to
// install the locked packs' plugins, as a run with --yes does.
func TestUpdateHelpSaysWhatAPreviewWrites(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"update", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	help := strings.Join(strings.Fields(stdout.String()), " ")
	for _, want := range []string{
		"into the Bonsai home's cache, a preview's too",
		"with nothing to change, with or without --yes, it records this machine's record of the checkout",
		"and brings this machine's plugins to the lock",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("update's help has no %q:\n%s", want, stdout.String())
		}
	}
}
