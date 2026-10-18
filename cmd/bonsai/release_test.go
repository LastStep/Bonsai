package main

// bonsai status's newer-release read through the command (plan-5 5.6.5): status --full reads Bonsai's tags once a day,
// every status, check and --line then say it from the home's record, offline; a home that cannot keep the record
// gets one line on stderr, never a failure.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// No test reaches the network: every status --full in this package's tests reads Bonsai's tags through this, unless
// the test sets its own (setReleaseTags).
func init() {
	releaseTags = func(string, time.Duration) ([]string, error) {
		return nil, errors.New("git ls-remote failed: tests read no network")
	}
}

// setReleaseTags makes status --full read tags, counting its calls, for this test.
func setReleaseTags(t *testing.T, calls *int, tags ...string) {
	t.Helper()
	old := releaseTags
	releaseTags = func(source string, timeout time.Duration) ([]string, error) {
		*calls++
		if source != engine.ReleaseSource || timeout != engine.ReleaseTimeout {
			t.Errorf("status --full read %s with %s", source, timeout)
		}
		return tags, nil
	}
	t.Cleanup(func() { releaseTags = old })
}

// setVersion makes this bonsai the given version for this test.
func setVersion(t *testing.T, v string) {
	t.Helper()
	old := version
	version = v
	t.Cleanup(func() { version = old })
}

func TestStatusNewerRelease(t *testing.T) {
	c := newCLI(t)
	c.link()
	setPlugins(t, fakeList{})
	setVersion(t, "1.0.0")
	calls := 0
	setReleaseTags(t, &calls, "v0.4.3", "base-v1.0.0", "v1.0.0", "v1.0.1")
	if code, out, errOut := c.run("", "status", "--json"); code != 0 || calls != 0 || !strings.Contains(out, `"state": "unknown"`) {
		t.Fatalf("status before any read (%d calls): %d\n%s%s", calls, code, out, errOut)
	}
	code, out, errOut := c.run("", "status", "--full", "--json")
	doc := fits(t, out, "status")
	rel, _ := doc.Get("bonsai_release")
	if code != 0 || calls != 1 || errOut != "" || !strings.Contains(out, `"state": "newer"`) ||
		!strings.Contains(out, `"curl -fsSLO https://github.com/LastStep/Bonsai/releases/download/v1.0.1/checksums.txt"`) || rel == nil {
		t.Fatalf("status --full --json: %d (%d calls)\n%s%s", code, calls, out, errOut)
	}
	// Offline from here: status, its text, check and the line, with no read.
	if code, out, _ := c.run("", "status"); code != 0 || !strings.Contains(flat(out), "Bonsai v1.0.1 is released this machine's is v1.0.0") {
		t.Errorf("status: %d\n%s", code, out)
	}
	if code, out, _ := c.run("", "check", "--json"); !strings.Contains(out, `"code": "bonsai-newer"`) || !strings.Contains(out,
		"a person installs Bonsai v1.0.1 on each side: the lines are in `bonsai status --full --json`, `bonsai_release.lines`") {
		t.Errorf("check --json: %d\n%s", code, out)
	}
	if code, out, _ := c.run("", "status", "--line"); code != 0 || !strings.HasSuffix(out, " bonsai:v1.0.1\n") {
		t.Errorf("status --line: %d %q", code, out)
	}
	if c.run("", "status", "--full", "--json"); calls != 1 {
		t.Errorf("status --full read again within the day: %d calls", calls)
	}
}

func TestStatusReleaseNotKept(t *testing.T) {
	c := newCLI(t)
	c.link()
	setPlugins(t, fakeList{})
	setVersion(t, "1.0.0")
	calls := 0
	setReleaseTags(t, &calls, "v1.0.1")
	home, _ := workspace.Home()
	if err := os.RemoveAll(filepath.Join(home, "cache")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "cache"), []byte("a file where the folder goes"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := c.run("", "status", "--full", "--json")
	if code != 0 || !strings.Contains(out, `"state": "newer"`) || !strings.HasPrefix(errOut, "Bonsai's release read was not kept in ") ||
		strings.Count(errOut, "\n") != 1 {
		t.Errorf("a home that cannot keep the record: %d\n%s\nstderr: %s", code, out, errOut)
	}
}
