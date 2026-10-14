package clean

// The messages kind (step 5.5.10, note 5): a thread goes, every file of it with its clean record, once its last
// message is older than keep_days or beyond the newest keep_newest, and never while it is open, whatever its age; a
// malformed file and a staged write are left alone; null keeps; the session's end cleans it too.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// thread writes a thread in l: a brief at at, and a report a minute later when answered.
func thread(t *testing.T, l workspace.Local, at time.Time, answered bool) []string {
	t.Helper()
	clock := at
	p := message.Place{Workspace: testID, Local: l, Now: func() time.Time { return clock }, Getenv: func(string) string { return "" }}
	b, e := message.Write(p, message.Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Body: "Go.", WorkDir: l.Root},
		message.DefaultRules())
	if e != nil {
		t.Fatal(e)
	}
	paths := []string{b.Path}
	if answered {
		clock = at.Add(time.Minute)
		r, e := message.Write(p, message.Writing{From: "builder@T-0901", To: "orchestrator", Kind: "report", Status: "done",
			Thread: b.Message.ID, Body: "Done.", WorkDir: l.Root}, message.DefaultRules())
		if e != nil {
			t.Fatal(e)
		}
		paths = append(paths, r.Path)
	}
	return paths
}

// left gives the message files there are, by path.
func left(t *testing.T, l workspace.Local) []string {
	t.Helper()
	var out []string
	_ = filepath.Walk(filepath.Join(l.Main, ".bonsai", "local", "messages"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(l.Main, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func has(list []string, p string) bool {
	for _, x := range list {
		if x == p {
			return true
		}
	}
	return false
}

func TestMessagesByAge(t *testing.T) {
	l := project(t, "")
	oldOpen := thread(t, l, daysAgo(400), false)
	oldDone := thread(t, l, daysAgo(31), true)
	newDone := thread(t, l, daysAgo(29), true)
	malformed := write(t, l.Main, ".bonsai/local/messages/builder-T-0901/M-2026-01-01-0a0b0c0d.md",
		"---\nformat: bonsai.message/1\nid: M-2026-01-01-0a0b0c0d\nthread: "+filepath.Base(strings.TrimSuffix(oldDone[0], ".md"))+"\n---\n")
	stamp(t, malformed, daysAgo(400))
	staged := write(t, l.Main, ".bonsai/local/messages/builder-T-0901/.M-2026-01-01-0a0b0c0e.md.tmp-1", "staged\n")
	stamp(t, staged, daysAgo(400))
	d := Files(opts(l), Messages)
	if len(d.Errs) != 0 || d.Stopped {
		t.Fatalf("errors %v, stopped %v", d.Errs, d.Stopped)
	}
	there := left(t, l)
	for _, p := range oldDone {
		if has(there, p) {
			t.Errorf("%s of a thread 31 days old, answered, is still there", p)
		}
	}
	for _, p := range append(append(oldOpen, newDone...), ".bonsai/local/messages/builder-T-0901/M-2026-01-01-0a0b0c0d.md",
		".bonsai/local/messages/builder-T-0901/.M-2026-01-01-0a0b0c0e.md.tmp-1") {
		if !has(there, p) {
			t.Errorf("%s is gone: an open thread, a thread 29 days old, a malformed file and a staged write stay", p)
		}
	}
	// Each file gone has its clean record, the brief first: target its path, reason the rule.
	if len(d.Cleaned) != 2 || d.Cleaned[0].Target != oldDone[0] || d.Cleaned[1].Target != oldDone[1] ||
		d.Cleaned[0].Reason != "generated.messages.keep_days=30" || d.Cleaned[0].Kind != Messages {
		t.Errorf("cleaned: %+v", d.Cleaned)
	}
	want := []string{oldDone[0] + "=generated.messages.keep_days=30", oldDone[1] + "=generated.messages.keep_days=30"}
	sort.Strings(want)
	if got := cleanRecords(t, l); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("clean records: %v", got)
	}
	// An open thread is never cleaned, however old: answered, it goes at the next run.
	again := Files(opts(l), Messages)
	if len(again.Cleaned) != 0 {
		t.Errorf("a second run cleaned %+v", again.Cleaned)
	}
}

// keep_newest ranks every thread by its last message, open ones among them (a protected thread still counts among the
// newest), and never cleans an open one.
func TestMessagesKeepNewest(t *testing.T) {
	l := project(t, "  messages:\n    keep_days: null\n    keep_newest: 2\n")
	a := thread(t, l, daysAgo(10), true)
	b := thread(t, l, daysAgo(8), false)
	c := thread(t, l, daysAgo(6), true)
	dd := thread(t, l, daysAgo(4), true)
	d := Files(opts(l), Messages)
	there := left(t, l)
	gone := func(paths []string) bool { return !has(there, paths[0]) && !has(there, paths[len(paths)-1]) }
	// The newest two are dd and c; b, open, is beyond them and stays; a goes.
	if !gone(a) || !has(there, b[0]) || !has(there, c[0]) || !has(there, dd[1]) {
		t.Errorf("kept: %v (cleaned %+v)", there, d.Cleaned)
	}
	if len(d.Cleaned) != 2 || d.Cleaned[0].Reason != "generated.messages.keep_newest=2" {
		t.Errorf("cleaned: %+v", d.Cleaned)
	}
}

// generated.messages null keeps every thread; the session's end cleans the kind, within its budget.
func TestMessagesNullAndSessionEnd(t *testing.T) {
	l := project(t, "  messages: null\n")
	old := thread(t, l, daysAgo(60), true)
	if d := Files(opts(l), Messages); len(d.Cleaned) != 0 || !has(left(t, l), old[0]) {
		t.Errorf("messages: null cleaned %+v", d.Cleaned)
	}
	l = project(t, "")
	old = thread(t, l, daysAgo(60), true)
	d := SessionEnd(opts(l))
	if has(left(t, l), old[0]) || len(d.Cleaned) != 2 {
		t.Errorf("the session's end: %+v, left %v", d, left(t, l))
	}
}

// A file Windows holds open is left after one try, and its thread stops there: the brief kept, its report not
// deleted before it, so the thread still reads answered and goes at a later run.
func TestMessagesBusyFile(t *testing.T) {
	l := project(t, "")
	old := thread(t, l, daysAgo(60), true)
	defer func(f func(string) error, b func(error) bool) { removeFile, busy = f, b }(removeFile, busy)
	removeFile = func(string) error { return os.ErrPermission }
	busy = func(error) bool { return true }
	d := Files(opts(l), Messages)
	if len(d.Busy) != 1 || d.Busy[0].Target != old[0] || !has(left(t, l), old[1]) {
		t.Errorf("busy: %+v, left %v", d, left(t, l))
	}
	removeFile, busy = os.Remove, workspace.IsBusy
	if d := Files(opts(l), Messages); len(d.Cleaned) != 2 {
		t.Errorf("the later run: %+v", d)
	}
}
