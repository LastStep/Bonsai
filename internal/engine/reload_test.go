package engine

// BUG-107's words (T-5736): once init or update installs a pack's plugin under a new marketplace name (a first link,
// a pack added or moved, the workspace renamed), a Claude Code session already open in the checkout loads the pack's
// skills only after /reload-plugins or a fresh start, and the install result says so; a plugin still waiting says it
// for when it is installed, and so does check's plugin-trust warning. An install under the same name says nothing.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

func TestPluginReloadWords(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "reload")
	p := e.link(t, root, e.pack.A)
	first := p.PluginConsent("")
	if !first.Fresh {
		t.Fatalf("a first link's consent is not fresh: %+v", first)
	}
	got := InstallPlugins(root, p.Config, p.NewLock(), &fakeCLI{}, first)
	if len(got) != 1 || got[0].Result != "installed" || !strings.HasSuffix(got[0].Message, "; "+ReloadWords) {
		t.Errorf("a first link's install: %+v", got)
	}
	if !strings.Contains(ReloadWords, "/reload-plugins") || !strings.Contains(ReloadWords, "a fresh start of Claude Code") {
		t.Errorf("the words: %q", ReloadWords)
	}

	// An update with nothing to change keeps the marketplace's name: an install there is a no-op, and says nothing.
	same := e.apply(t, root, Request{})
	if c := same.PluginConsent(""); c.Fresh {
		t.Errorf("an update that keeps the name is fresh: %+v", c)
	} else if got := InstallPlugins(root, same.Config, same.NewLock(), &fakeCLI{}, c); len(got) != 1 || strings.Contains(got[0].Message, "reload") {
		t.Errorf("an install under the same name: %+v", got)
	}

	// The pack moved to another commit: a new name, and the words again.
	testpack.SetRef(t, root, e.pack.A, e.pack.B)
	moved := e.apply(t, root, Request{})
	c := moved.PluginConsent("")
	if !c.Fresh {
		t.Fatalf("an update that moves the name is not fresh: %+v", c)
	}
	if got := InstallPlugins(root, moved.Config, moved.NewLock(), &fakeCLI{}, c); len(got) != 1 || !strings.HasSuffix(got[0].Message, "; "+ReloadWords) {
		t.Errorf("an install under a new name: %+v", got)
	}

	// Waiting for the folder's trust, or for --allow-exec: the words for when it is installed.
	want := PluginID(testpack.ID, MarketplaceName("demo", []string{e.pack.B}))
	f := &fakeCLI{install: map[string]InstallResult{want: {Outcome: "failed", FailureCode: "not_found"}}}
	if got := InstallPlugins(root, moved.Config, moved.NewLock(), f, c); len(got) != 1 || got[0].Result != "waiting" ||
		!strings.HasSuffix(got[0].Message, "; once it is installed, "+ReloadWords) {
		t.Errorf("waiting for trust: %+v", got)
	}
	code := c
	code.Code = map[string][]CodePart{testpack.ID: {{Path: "hooks/hooks.json"}}}
	if got := InstallPlugins(root, moved.Config, moved.NewLock(), &fakeCLI{}, code); len(got) != 1 || got[0].Result != "waiting" ||
		!strings.HasSuffix(got[0].Message, "; once it is installed, "+ReloadWords) {
		t.Errorf("waiting for --allow-exec: %+v", got)
	}

	// check's plugin-trust warning says it too.
	r, err := checkLocal(t, root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	ComparePlugins(r, &fakeCLI{markets: []string{"another-marketplace"}})
	found := false
	for _, w := range r.Warnings {
		if w.Code == "plugin-trust" {
			found = true
			if !strings.Contains(w.Message, "; once update installs it, "+ReloadWords) {
				t.Errorf("plugin-trust: %+v", w)
			}
		}
	}
	if !found {
		t.Errorf("no plugin-trust warning: %+v", r.Warnings)
	}
}
