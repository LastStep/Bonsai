package status

// status --json's messages (set 8; step 5.5.10, note 5): the threads of messages between agents counted, open and
// flagged, as bonsai messages folds them, at bonsai.yaml's thresholds as committed; a malformed file never a problem.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/message"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestStatusMessages(t *testing.T) {
	root, _ := project(t, cfg)
	root = real(t, root)
	testpack.Git(t, root, "add", "bonsai.yaml")
	testpack.Git(t, root, "commit", "-q", "-m", "bonsai.yaml") // so its messages count once committed
	start := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	now := start
	defer func(f func() time.Time) { messagesNow = f }(messagesNow)
	messagesNow = func() time.Time { return now }
	counts := func(t *testing.T) string {
		t.Helper()
		doc, code := Build(root, "dev")
		if code != ExitOK {
			t.Fatalf("exit %d: %s", code, schema.Show(doc))
		}
		checkShape(t, doc, false)
		v, _ := doc.Get("messages")
		return schema.Show(v)
	}
	if got := counts(t); got != `{"open":0,"flagged":0}` {
		t.Errorf("no messages: %s", got)
	}
	p := message.Place{Workspace: "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", Local: workspace.Local{Root: root, Main: root},
		Now: func() time.Time { return start }, Getenv: func(string) string { return "" }}
	write := func(w message.Writing) *message.Written {
		t.Helper()
		w.Body, w.WorkDir = "Go.", root
		out, e := message.Write(p, w, message.DefaultRules())
		if e != nil {
			t.Fatal(e)
		}
		return out
	}
	a := write(message.Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief"}).Message
	write(message.Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Thread: a.ID})
	b := write(message.Writing{From: "orchestrator", To: "verifier", Kind: "brief"}).Message
	write(message.Writing{From: "verifier", To: "orchestrator", Kind: "report", Status: "done", Thread: b.ID})
	if err := os.WriteFile(filepath.Join(root, ".bonsai", "local", "messages", "verifier", "M-2026-10-16-0a0b0c0d.md"), []byte("---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// One open thread, nothing flagged at the defaults; a malformed file of no thread that reads is no problem.
	if got := counts(t); got != `{"open":1,"flagged":0}` {
		t.Errorf("at the defaults: %s", got)
	}
	now = start.Add(time.Hour)
	if got := counts(t); got != `{"open":1,"flagged":1}` {
		t.Errorf("an hour on: %s", got)
	}
	// The project's own circling_briefs: 1, counted once committed.
	now = start
	raw, err := os.ReadFile(filepath.Join(root, "bonsai.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bonsai.yaml"), []byte(string(raw)+"messages:\n  unanswered_minutes: null\n  circling_briefs: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := counts(t); got != `{"open":1,"flagged":0}` {
		t.Errorf("uncommitted: %s", got)
	}
	testpack.Git(t, root, "add", "bonsai.yaml")
	testpack.Git(t, root, "commit", "-q", "-m", "thresholds")
	if got := counts(t); got != `{"open":1,"flagged":1}` {
		t.Errorf("circling_briefs 1, committed: %s", got)
	}
}
