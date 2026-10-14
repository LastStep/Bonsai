package main

// bonsai message and bonsai messages (step 5.5.10) through the command line, in a linked project: a brief written,
// shown (its trust line first), answered from standard input and listed; a follow-up and a closing report; a
// malformed file beside them listed and skipped; a worktree writing its main checkout's folder; every --json against
// bonsai.messages/1; every refusal with its word, its exit code and a next step; and nothing changed in the project
// but .bonsai/local/messages/ and one message log record for each message. internal/message's tests walk every rule;
// these hold the commands to them.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// messagesDoc runs a message command with --json, holds its output to bonsai.messages/1, and gives the document.
func (c *cli) messagesDoc(exit int, args ...string) schema.Object {
	c.t.Helper()
	code, out, errOut := c.run("", append(args, "--json")...)
	if code != exit {
		c.t.Fatalf("%v: exit %d, want %d\n%s%s", args, code, exit, out, errOut)
	}
	return fits(c.t, out, "messages")
}

// messageClock sets the commands' clock: a second a call from 2026-10-16 09:00 UTC.
func messageClock(t *testing.T) {
	t.Helper()
	tick := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	old := messageNow
	t.Cleanup(func() { messageNow = old })
	messageNow = func() time.Time { tick = tick.Add(time.Second); return tick }
}

func TestMessageCommands(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	messageClock(t)
	// Not linked: exit 4, not-linked naming bonsai init, the workspace null.
	code, _, errOut := c.run("", "messages")
	if code != 4 || !strings.Contains(errOut, "next: run bonsai init") {
		t.Errorf("not linked: %d %q", code, errOut)
	}
	doc := c.messagesDoc(4, "message", "--show", "M-2026-10-16-3f9a01bc")
	if w, _ := doc.Get("workspace"); w != nil || errorIn(doc, "messages").String("code") != "not-linked" {
		t.Errorf("not linked: %s", schema.Show(doc))
	}
	c.link()
	if err := os.MkdirAll(filepath.Join(c.root, "work"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c.root, "work", "plan.md"), []byte("the plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c.commit("link")
	before := c.localFiles()

	// A brief: its id and its file printed.
	code, out, errOut := c.run("", "message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--task", "T-0901",
		"--budget", "minutes=120", "--pointer", "work/plan.md", "--body", "Objective: dark mode.\nBoundaries: the settings screen.")
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if code != 0 || len(lines) != 2 || !strings.HasPrefix(lines[0], "M-2026-10-16-") || lines[1] != ".bonsai/local/messages/builder-T-0901/"+lines[0]+".md" {
		t.Fatalf("a brief: %d %q %q", code, out, errOut)
	}
	brief := lines[0]
	// --show: the trust line first, then the file as written.
	code, out, _ = c.run("", "message", "--show", brief)
	if code != 0 || !strings.HasPrefix(out, format.TrustLine("orchestrator")+"\n---\nformat: bonsai.message/1") ||
		!strings.HasSuffix(out, "---\n\nObjective: dark mode.\nBoundaries: the settings screen.\n") || !strings.Contains(out, "path: work/plan.md\n") {
		t.Errorf("--show: %d\n%s", code, out)
	}
	doc = c.messagesDoc(0, "message", "--show", brief)
	msg, _ := doc.Get("message")
	m := msg.(schema.Object)
	if m.String("trust") != format.TrustLine("orchestrator") || m.String("body") != "Objective: dark mode.\nBoundaries: the settings screen.\n" ||
		doc.String("path") != ".bonsai/local/messages/builder-T-0901/"+brief+".md" {
		t.Errorf("--show --json: %s", schema.Show(doc))
	}
	if th := threadsIn(doc); len(th) != 1 || th[0].String("state") != "open" || th[0].String("id") != brief {
		t.Errorf("--show --json's thread: %v", th)
	}
	if w, _ := doc.Get("workspace"); w.(schema.Object).String("root") != filepath.ToSlash(c.root) {
		t.Errorf("the workspace: %v", w)
	}
	// The plan changed since: --show says so.
	if err := os.WriteFile(filepath.Join(c.root, "work", "plan.md"), []byte("the plan, changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, out, _ = c.run("", "message", "--show", brief); !strings.Contains(out, "\npointer work/plan.md: changed since the message was written") {
		t.Errorf("a changed pointer:\n%s", out)
	}
	// The report, its words from standard input (the stall stop's case), into the brief's thread.
	code, out, errOut = c.run("Done: the switch.\nLeft: nothing.\n", "message", "--from", "builder@T-0901", "--to", "orchestrator",
		"--kind", "report", "--status", "done", "--thread", brief, "--body", "-")
	report := strings.Split(out, "\n")[0]
	if code != 0 || !strings.HasPrefix(report, "M-2026-10-16-") {
		t.Fatalf("the report: %d %q %q", code, out, errOut)
	}
	doc = c.messagesDoc(0, "message", "--show", report)
	msg, _ = doc.Get("message")
	m = msg.(schema.Object)
	if m.String("body") != "Done: the switch.\nLeft: nothing.\n" || m.String("reply_to") != brief || m.String("task") != "T-0901" ||
		m.String("trust") != format.TrustLine("builder@T-0901") {
		t.Errorf("the report: %s", schema.Show(m))
	}
	// --json of a write: the message, its file and its thread, open; the words from a file a Windows editor saved (a
	// byte-order mark, CRLF).
	if err := os.WriteFile(filepath.Join(c.root, "brief.md"), []byte("\xef\xbb\xbfObjective: a check.\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc = c.messagesDoc(0, "message", "--from", "orchestrator", "--to", "verifier@T-0902", "--kind", "brief", "--body-file", "brief.md")
	msg, _ = doc.Get("message")
	second := msg.(schema.Object).String("id")
	if msg.(schema.Object).String("body") != "Objective: a check.\n" || threadsIn(doc)[0].String("state") != "open" {
		t.Errorf("a write's --json: %s", schema.Show(doc))
	}
	// A malformed file beside them: listed, skipped, its thread flagged.
	bad := filepath.Join(c.root, ".bonsai", "local", "messages", "builder-T-0901", "M-2026-10-16-0a0b0c0d.md")
	if err := os.WriteFile(bad, []byte("---\nformat: bonsai.message/1\nid: M-2026-10-16-0a0b0c0d\nthread: "+brief+"\nkind: report\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, _ = c.run("", "messages")
	want := []string{
		"answered  " + brief + "  orchestrator -> builder@T-0901  T-0901  2 messages, last ",
		"  flags: malformed\n",
		"open      " + second + "  orchestrator -> verifier@T-0902  -  1 messages, last ",
		"malformed  .bonsai/local/messages/builder-T-0901/M-2026-10-16-0a0b0c0d.md: it does not read as bonsai.message/1",
	}
	for _, w := range want {
		if code != 0 || !strings.Contains(out, w) {
			t.Errorf("messages: %d, want %q in\n%s", code, w, out)
		}
	}
	doc = c.messagesDoc(0, "messages")
	if th := threadsIn(doc); len(th) != 2 || th[0].String("id") != brief || th[1].String("id") != second {
		t.Errorf("messages --json: %s", schema.Show(doc))
	}
	if mal, _ := doc.Get("malformed"); len(mal.([]any)) != 1 {
		t.Errorf("malformed: %v", mal)
	}
	for _, f := range []struct {
		args []string
		want string
	}{
		{[]string{"--open"}, second},
		{[]string{"--to", "orchestrator"}, brief},
		{[]string{"--task", "T-0901"}, brief},
		{[]string{"--to", "auditor"}, ""},
	} {
		var ids []string
		for _, th := range threadsIn(c.messagesDoc(0, append([]string{"messages"}, f.args...)...)) {
			ids = append(ids, th.String("id"))
		}
		if strings.Join(ids, " ") != f.want {
			t.Errorf("messages %v: %v", f.args, ids)
		}
	}
	if _, out, _ = c.run("", "messages", "--to", "auditor"); !strings.HasPrefix(out, "no threads that the filters keep in .bonsai/local/messages/\n") {
		t.Errorf("none kept: %q", out)
	}
	// Nothing changed but the messages and the log; one message record for each message.
	after := c.localFiles()
	for path, text := range after {
		if before[path] != text && !strings.HasPrefix(path, ".bonsai/local/messages/") && !strings.HasPrefix(path, ".bonsai/local/log/") &&
			path != "brief.md" && path != "work/plan.md" {
			t.Errorf("%s changed", path)
		}
	}
	var targets []string
	for path := range after {
		if !strings.HasPrefix(path, ".bonsai/local/log/") {
			continue
		}
		lf, err := record.ReadLog(filepath.Join(c.root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range lf.Records {
			if r.Event == "message" {
				targets = append(targets, *r.Target)
			}
		}
	}
	if strings.Join(targets, " ") != brief+" "+report+" "+second {
		t.Errorf("message records: %v", targets)
	}
}

// A worktree writes and reads its main checkout's messages, as the log and asks do.
func TestMessageFromAWorktree(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	messageClock(t)
	c.link()
	c.commit("link")
	wt := filepath.Join(c.tmp, "branch")
	testpack.Git(t, c.root, "worktree", "add", "-q", "-b", "branch", wt)
	t.Chdir(wt)
	if err := os.WriteFile(filepath.Join(wt, "notes.md"), []byte("in the worktree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := c.run("", "message", "--from", "builder@T-0901", "--to", "orchestrator", "--kind", "brief", "--pointer", "notes.md",
		"--body", "A brief from the worktree.")
	id := strings.Split(out, "\n")[0]
	if code != 0 {
		t.Fatalf("from a worktree: %d %s", code, errOut)
	}
	if _, err := os.Stat(filepath.Join(c.root, ".bonsai", "local", "messages", "orchestrator", id+".md")); err != nil {
		t.Errorf("not in the main checkout's folder: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wt, ".bonsai", "local", "messages")); !os.IsNotExist(err) {
		t.Errorf("the worktree has a messages folder of its own: %v", err)
	}
	t.Chdir(c.root)
	if _, out, _ = c.run("", "message", "--show", id); !strings.Contains(out, "\npointer notes.md: missing from this checkout now\n") {
		t.Errorf("the worktree's pointer read in the main checkout:\n%s", out)
	}
}

// Every refusal of message and messages: its exit code, its word in a document that fits bonsai.messages/1, and a
// next step in the human lines; nothing written.
func TestMessageRefusals(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	c.link()
	c.commit("link")
	brief := strings.Split(func() string {
		_, out, _ := c.run("", "message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go.")
		return out
	}(), "\n")[0]
	before := c.localFiles()
	w, ws := messageWord, messagesWord
	base := []string{"message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go."}
	with := func(args ...string) []string { return append(append([]string{}, base...), args...) }
	for _, r := range []struct {
		word *Word
		args []string
		exit int
		code string
	}{
		{w, []string{"message", "now"}, 2, "bad-flag"},
		{w, []string{"message", "--show", brief, "--to", "verifier"}, 2, "bad-flag"},
		{w, with("--body-file", "x.md"), 2, "bad-flag"},
		{w, []string{"message", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go."}, 2, "missing-value"},
		{w, []string{"message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief"}, 2, "missing-value"},
		{w, []string{"message", "--from", "orchestrator", "--to", "builder:T-0901", "--kind", "brief", "--body", "Go."}, 2, "bad-value"},
		{w, []string{"message", "--from", "orchestrator", "--to", "nul", "--kind", "brief", "--body", "Go."}, 2, "bad-value"},
		{w, []string{"message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "order", "--body", "Go."}, 2, "bad-value"},
		{w, with("--status", "done"), 2, "bad-flag"},
		{w, with("--budget", "hours=2"), 2, "bad-value"},
		{w, with("--task", "T-1"), 2, "bad-value"},
		{w, with("--pointer", "work/none.md"), 2, "pointer-missing"},
		{w, with("--thread", "M-2026-10-16-00000000"), 4, "message-not-found"},
		{w, []string{"message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body-file", "none.md"}, 2, "bad-value"},
		{w, []string{"message", "--from", "builder@T-0901", "--to", "orchestrator", "--kind", "report", "--thread", brief, "--body", "x"}, 2, "report-no-status"},
		{w, []string{"message", "--from", "builder@T-0901", "--to", "verifier", "--kind", "report", "--status", "done", "--thread", brief, "--body", "x"}, 4, "report-not-in-thread"},
		{w, []string{"message", "--from", "auditor@T-0901", "--to", "orchestrator", "--kind", "report", "--status", "done", "--thread", brief, "--body", "x"}, 4, "report-not-in-thread"},
		{w, []string{"message", "--show", "M-2026-10-16-00000000"}, 4, "message-not-found"},
		{w, []string{"message", "--show", "../x"}, 2, "bad-value"},
		{ws, []string{"messages", "now"}, 2, "bad-flag"},
		{ws, []string{"messages", "--to", "Builder"}, 2, "bad-value"},
		{ws, []string{"messages", "--task", "5510"}, 2, "bad-value"},
	} {
		wantRefusal(t, r.word, r.args, r.exit, r.code)
	}
	if after := c.localFiles(); len(after) != len(before) {
		t.Errorf("a refusal wrote: %d files, %d before", len(after), len(before))
	}
	for path, text := range c.localFiles() {
		if before[path] != text {
			t.Errorf("a refusal changed %s", path)
		}
	}
}

// threadsIn gives a messages document's threads.
func threadsIn(doc schema.Object) []schema.Object {
	v, _ := doc.Get("threads")
	var out []schema.Object
	for _, t := range v.([]any) {
		out = append(out, t.(schema.Object))
	}
	return out
}

// A brief's budget against each limit's max in force for its recipient's role, through workspace.RolesInForce: the
// pack's max, then the project's own in bonsai.yaml once committed (an uncommitted change counts for nothing); a
// limit with no max takes any number.
func TestMessageBudgetMax(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	src, shas := rolesPack(t, c.tmp)
	c.linkTo(src, shas[0])
	c.commit("link")
	brief := func(to, budget string) []string {
		return []string{"message", "--from", "orchestrator", "--to", to, "--kind", "brief", "--budget", budget, "--body", "Go."}
	}
	e := errorIn(c.messagesDoc(2, brief("builder@T-0901", "minutes=181")...), "messages")
	if e.String("code") != "budget-over-max" || !strings.Contains(e.String("message"), "minutes 181, over its max in force for the role builder, 180") {
		t.Errorf("over the pack's max: %s", schema.Show(e))
	}
	wantRefusal(t, messageWord, brief("auditor@T-0901#2", "actions=151"), 2, "budget-over-max")
	c.messagesDoc(0, brief("builder@T-0901", "minutes=180,actions=600,red_runs=4")...)
	c.messagesDoc(0, brief("auditor@T-0901", "red_runs=99")...)
	raw := c.read("bonsai.yaml")
	c.write("bonsai.yaml", strings.Replace(raw, "    roles: null\n", "    roles:\n      builder:\n        minutes:\n          max: 480\n", 1))
	c.messagesDoc(2, brief("builder@T-0901", "minutes=481")...)
	c.messagesDoc(2, brief("builder@T-0901", "minutes=480")...)
	c.commit("the builder's minutes")
	c.messagesDoc(2, brief("builder@T-0901", "minutes=481")...)
	if doc := c.messagesDoc(0, brief("builder@T-0901", "minutes=480")...); doc.String("path") == "" {
		t.Errorf("within the project's max: %s", schema.Show(doc))
	}
	// A lock whose declares do not read is refused before anything is written; --show and messages do not read it.
	c.write(".bonsai/lock.json", strings.Replace(c.read(".bonsai/lock.json"), `"roles": [`, `"roles": 7, "x": [`, 1))
	c.commit("a broken lock")
	if e := errorIn(c.messagesDoc(4, brief("builder@T-0901", "minutes=1")...), "messages"); e.String("code") != "bad-lock" {
		t.Errorf("a broken lock: %s", schema.Show(e))
	}
	c.messagesDoc(0, "messages")
}
