package engine

// update's (and init's, unlink's) human lines for this machine's plugins: each plugin's lines stand apart from the
// next plugin's, so one plugin's last words (its next step's "run: bonsai update") never sit straight above the next
// plugin's result word.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/textui"
)

func TestPluginLinesStandApart(t *testing.T) {
	p := &Plan{Plugins: []PluginResult{
		{Plugin: "base@bonsai-x-12345678", Result: "waiting", Message: "Claude Code has not registered this checkout's marketplace yet",
			Next: TrustNext(false), Who: "person"},
		{Plugin: "workflow@bonsai-x-12345678", Result: "waiting", Message: "the same", Next: TrustNext(true), Who: "person"},
		{Plugin: "extra@bonsai-x-12345678", Result: "installed", Message: "at 1234567"},
	}}
	for _, w := range []int{40, 60, 80, 100, 150} {
		d := &textui.Doc{}
		p.PluginsTo(d)
		text := d.Render(textui.Look{Width: w})
		lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
		results := 0
		for i, l := range lines {
			f := strings.Fields(l)
			if len(f) == 0 || (f[0] != "waiting" && f[0] != "installed") {
				continue
			}
			results++
			if results > 1 && strings.TrimSpace(lines[i-1]) != "" {
				t.Errorf("width %d: plugin %d's line %q runs on from the line before it, %q:\n%s", w, results, l, lines[i-1], text)
			}
		}
		if results != 3 {
			t.Errorf("width %d: %d plugin lines, want 3:\n%s", w, results, text)
		}
		if strings.HasSuffix(text, "\n\n") {
			t.Errorf("width %d: the plugins end with an empty line:\n%s", w, text)
		}
	}
}
