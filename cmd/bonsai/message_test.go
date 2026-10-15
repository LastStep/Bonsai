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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
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

// Where the lock declares roles, --from and --to must name one of them (role-not-declared, naming the declared ones);
// where it declares none, any role is taken.
func TestMessageDeclaredRoles(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	src, shas := rolesPack(t, c.tmp)
	c.linkTo(src, shas[0])
	c.commit("link")
	msg := func(from, to string) []string {
		return []string{"message", "--from", from, "--to", to, "--kind", "brief", "--body", "Go."}
	}
	e := errorIn(c.messagesDoc(2, msg("orchestrator", "planner@T-0901")...), "messages")
	if e.String("code") != "role-not-declared" || !strings.Contains(e.String("message"), "--to names the role planner, which the linked packs do not declare (they declare orchestrator, builder, auditor)") {
		t.Errorf("an undeclared --to: %s", schema.Show(e))
	}
	wantRefusal(t, messageWord, msg("tester@T-0901", "builder@T-0901"), 2, "role-not-declared")
	c.messagesDoc(0, msg("orchestrator", "builder@T-0901#2")...)
	c.messagesDoc(0, msg("builder@T-0901", "auditor@T-0901")...)
	// The plain test pack declares no role: any role is taken.
	p := newCLI(t)
	p.link()
	p.commit("link")
	p.messagesDoc(0, msg("tester@T-0901", "planner")...)
}

// The flags at their thresholds and not below, at Bonsai's defaults and at a project's own in bonsai.yaml (counted
// once committed): unanswered, an open thread's newest brief as old as the threshold's minutes; circling, more briefs
// one way between two addresses than the threshold.
func TestMessageThresholds(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	clock := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	old := messageNow
	t.Cleanup(func() { messageNow = old })
	messageNow = func() time.Time { return clock }
	c.link()
	c.commit("link")
	_, out, _ := c.run("", "message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go.")
	thread := strings.Split(out, "\n")[0]
	flags := func() string {
		th := threadsIn(c.messagesDoc(0, "messages"))
		f, _ := th[0].Get("flags")
		var out []string
		for _, x := range f.([]any) {
			out = append(out, x.(string))
		}
		return strings.Join(out, ",")
	}
	brief := func(from, to string) {
		t.Helper()
		if code, _, errOut := c.run("", "message", "--from", from, "--to", to, "--kind", "brief", "--thread", thread, "--body", "Again."); code != 0 {
			t.Fatalf("a brief in the thread: %s", errOut)
		}
	}
	start := clock
	// The defaults: 60 minutes, 3 briefs one way.
	clock = start.Add(59*time.Minute + 59*time.Second)
	if got := flags(); got != "" {
		t.Errorf("a second below 60 minutes: %q", got)
	}
	clock = start.Add(60 * time.Minute)
	if got := flags(); got != "unanswered" {
		t.Errorf("at 60 minutes: %q", got)
	}
	clock = start
	brief("orchestrator", "builder@T-0901")
	brief("orchestrator", "builder@T-0901")
	if got := flags(); got != "" {
		t.Errorf("three briefs one way: %q", got)
	}
	brief("orchestrator", "builder@T-0901")
	if got := flags(); got != "circling" {
		t.Errorf("four briefs one way: %q", got)
	}
	// A project's own: 5 minutes, 1 brief one way; uncommitted, the defaults still hold.
	raw := c.read("bonsai.yaml")
	if strings.Count(raw, "messages: null") != 1 {
		t.Fatalf("init wrote no messages: null line:\n%s", raw)
	}
	c.write("bonsai.yaml", strings.Replace(raw, "messages: null", "messages:\n  unanswered_minutes: 5\n  circling_briefs: 1", 1))
	_, out, _ = c.run("", "message", "--from", "orchestrator", "--to", "verifier", "--kind", "brief", "--body", "Check.")
	second := strings.Split(out, "\n")[0]
	own := func() string {
		for _, th := range threadsIn(c.messagesDoc(0, "messages")) {
			if th.String("id") == second {
				f, _ := th.Get("flags")
				var out []string
				for _, x := range f.([]any) {
					out = append(out, x.(string))
				}
				return strings.Join(out, ",")
			}
		}
		return "no thread"
	}
	clock = start.Add(5 * time.Minute)
	if got := own(); got != "" {
		t.Errorf("five minutes, the project's own not committed: %q", got)
	}
	c.commit("the project's thresholds")
	clock = start.Add(5*time.Minute - time.Second)
	if got := own(); got != "" {
		t.Errorf("a second below the project's 5 minutes: %q", got)
	}
	clock = start.Add(5 * time.Minute)
	if got := own(); got != "unanswered" {
		t.Errorf("at the project's 5 minutes: %q", got)
	}
	clock = start
	if code, _, errOut := c.run("", "message", "--from", "orchestrator", "--to", "verifier", "--kind", "brief", "--thread", second, "--body", "Again."); code != 0 {
		t.Fatal(errOut)
	}
	if got := own(); got != "circling" {
		t.Errorf("two briefs one way at the project's 1: %q", got)
	}
}

// check warns of each message file that does not read (format.MessageCheckWords' message-malformed, walked from its
// table as TestCheckTable walks check's words): a warning, never the exit code, about the file, an agent's, its next
// step naming only what Bonsai has; a message that reads, none.
func TestMessageCheckWarning(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	c.link()
	c.commit("link")
	if code, _, errOut := c.run("", "message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go."); code != 0 {
		t.Fatal(errOut)
	}
	before, out, _ := c.run("", "check", "--json")
	if strings.Contains(out, "message-malformed") {
		t.Fatalf("a message that reads is warned of:\n%s", out)
	}
	cases := map[string]string{
		"message-malformed": "builder-T-0901/M-2026-10-16-0a0b0c0d.md",
	}
	for _, w := range format.MessageCheckWords {
		rel, ok := cases[w.Word]
		if !ok {
			t.Fatalf("format.MessageCheckWords has %s, and this test no case for it", w.Word)
		}
		c.write(".bonsai/local/messages/"+rel, "---\nformat: bonsai.message/1\nid: M-2026-10-16-0a0b0c0d\n---\n")
		code, out, _ := c.run("", "check", "--json")
		doc := fits(t, out, "check")
		warnings, _ := doc.Get("warnings")
		found := false
		for _, f := range warnings.([]any) {
			fo := f.(schema.Object)
			if fo.String("code") != w.Word {
				continue
			}
			found = true
			n, _ := fo.Get("next")
			checkNext(t, w.Word, n.(schema.Object).String("do"))
			if fo.String("file") != ".bonsai/local/messages/"+rel || whoOf(fo) != "agent" || !strings.Contains(fo.String("message"), "the file has no thread, so it cannot be filed") {
				t.Errorf("%s: %s", w.Word, schema.Show(fo))
			}
		}
		if !found || code != before {
			t.Errorf("%s: exit %d (%d before), found %v:\n%s", w.Word, code, before, found, out)
		}
		if findings, _ := doc.Get("findings"); hasCode(findings.([]any), w.Word) {
			t.Errorf("%s is a finding", w.Word)
		}
		if _, out, _ := c.run("", "check"); !strings.Contains(out, "\nwarning: .bonsai/local/messages/"+rel+" does not read as a message between agents") {
			t.Errorf("check's text does not name it:\n%s", out)
		}
	}
}

// lineOf is the 1-based line of text in s that starts with prefix, 0 for none.
func lineOf(s, prefix string) int {
	for i, l := range strings.Split(s, "\n") {
		if strings.HasPrefix(l, prefix) {
			return i + 1
		}
	}
	return 0
}

// problemsOf runs status --json and gives its problems and its roles and messages fields, the document held to
// bonsai.status/1.
func (c *cli) problemsOf() ([]string, any, any) {
	c.t.Helper()
	code, out, errOut := c.run("", "status", "--json")
	if code != 0 {
		c.t.Fatalf("status: exit %d\n%s%s", code, out, errOut)
	}
	doc := fits(c.t, out, "status")
	p, _ := doc.Get("problems")
	var problems []string
	for _, x := range p.([]any) {
		problems = append(problems, x.(string))
	}
	roles, _ := doc.Get("roles")
	messages, _ := doc.Get("messages")
	return problems, roles, messages
}

// An unknown limit name under a packs entry's roles (VM's D2: minutez). A role's limits are a closed list, so it is no
// newer field to skip but a bonsai.yaml that does not read for the roles: check names it as config at bonsai.yaml's
// line, an agent's step, and still raises every other role finding; bonsai message refuses bad-config naming
// bonsai.yaml and its line (never the lock), with an agent's step that fixes it, its document naming the workspace;
// status names bonsai.yaml with the line and the step, its roles []. Fixed and committed, each reads again.
func TestMessageUnknownLimitName(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	src, shas := rolesPack(t, c.tmp)
	c.linkTo(src, shas[0])
	c.commit("link")
	good := c.read("bonsai.yaml")
	typo := strings.Replace(good, "    roles: null\n", "    roles:\n      builder:\n        minutes:\n          default: 240\n"+
		"        minutez:\n          max: 5\n", 1)
	c.write("bonsai.yaml", typo)
	line := strconv.Itoa(lineOf(typo, "        minutez:"))
	sentence := "bonsai.yaml line " + line + " sets minutez for the role builder under the pack workflow " +
		"(packs[0].roles.builder.minutez), which is not one of a role's limits (minutes, actions, red_runs)"
	step := "in bonsai.yaml line " + line + ", write minutes, actions or red_runs in place of minutez, or take it out"

	code, out, _ := c.run("", "check", "--json")
	findings, _ := fits(t, out, "check").Get("findings")
	var codes []string
	for _, f := range findings.([]any) {
		fo := f.(schema.Object)
		codes = append(codes, fo.String("code"))
		n, _ := fo.Get("next")
		if fo.String("code") == "config" && (fo.String("message") != sentence || fo.String("file") != "bonsai.yaml" || whoOf(fo) != "agent" ||
			n.(schema.Object).String("do") != step+"; then run: bonsai check") {
			t.Errorf("check's config finding: %s", schema.Show(fo))
		}
	}
	if code != 1 || !hasCode(findings.([]any), "config") || !hasCode(findings.([]any), "role-default-over-max") {
		t.Errorf("check: exit %d, findings %v: the unknown name and the builder's default over its max", code, codes)
	}

	c.commit("a limit's name mistyped")
	brief := []string{"message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go."}
	doc := c.messagesDoc(2, brief...)
	e := errorIn(doc, "messages")
	n, _ := e.Get("next")
	if e.String("code") != "bad-config" || e.String("message") != sentence || whoOf(e) != "agent" ||
		!strings.HasPrefix(n.(schema.Object).String("do"), step+", and commit it") || strings.Contains(schema.Show(e), "lock") {
		t.Errorf("bonsai message: %s", schema.Show(e))
	}
	if ws, _ := doc.Get("workspace"); ws == nil {
		t.Errorf("a refusal once bonsai.yaml was read names no workspace: %s", schema.Show(doc))
	}
	wantRefusal(t, messageWord, brief, 2, "bad-config")
	problems, roles, _ := c.problemsOf()
	named := false
	for _, p := range problems {
		named = named || strings.HasPrefix(p, sentence+"; next: "+step+", and commit it")
		if strings.Contains(p, "lock") {
			t.Errorf("a problem blames the lock: %s", p)
		}
	}
	if !named || schema.Show(roles) != "[]" {
		t.Errorf("status: roles %s, problems %q", schema.Show(roles), problems)
	}

	c.write("bonsai.yaml", strings.Replace(typo, "        minutez:\n          max: 5\n", "", 1))
	c.commit("the limit's name fixed")
	c.messagesDoc(0, brief...)
	if _, roles, _ := c.problemsOf(); len(roles.([]any)) != 3 {
		t.Errorf("status once fixed: roles %s", schema.Show(roles))
	}
}

// bonsai.yaml or the lock committed in a form that does not read (VM's D5) is never quietly replaced by the working
// copy, even one that reads: bonsai message (a brief, --show) and bonsai messages refuse naming the committed file
// (bad-config, bad-lock), their documents naming the workspace; status names it as a problem, roles [] and messages
// null for bonsai.yaml, roles [] for the lock. Where HEAD holds no such file (nothing committed yet), the working copy
// is read.
func TestMessageCommittedCopyUnread(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	src, shas := rolesPack(t, c.tmp)
	c.linkTo(src, shas[0])
	brief := []string{"message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go."}
	shown := c.messagesDoc(0, brief...) // HEAD names no commit yet: the files as they stand
	c.commit("link")
	good := c.read("bonsai.yaml")
	c.write("bonsai.yaml", strings.Replace(good, "messages: null", "messages:\n  unanswered_minutes: soon", 1))
	c.commit("a bonsai.yaml that does not read")
	c.write("bonsai.yaml", good) // the working copy reads; the committed one does not
	m, _ := shown.Get("message")
	for _, args := range [][]string{brief, {"message", "--show", m.(schema.Object).String("id")}, {"messages"}} {
		doc := c.messagesDoc(2, args...)
		e := errorIn(doc, "messages")
		if e.String("code") != "bad-config" || !strings.HasPrefix(e.String("message"), "bonsai.yaml as committed at the main checkout's HEAD does not read: bonsai.yaml") {
			t.Errorf("%v: %s", args, schema.Show(e))
		}
		if ws, _ := doc.Get("workspace"); ws == nil {
			t.Errorf("%v: no workspace", args)
		}
	}
	problems, roles, messages := c.problemsOf()
	if !strings.Contains(strings.Join(problems, "\n"), "bonsai.yaml as committed at the main checkout's HEAD does not read") ||
		schema.Show(roles) != "[]" || messages != nil {
		t.Errorf("status: roles %s, messages %s, problems %q", schema.Show(roles), schema.Show(messages), problems)
	}
	c.commit("bonsai.yaml reads again")
	c.messagesDoc(0, brief...)

	lock := c.read(".bonsai/lock.json")
	c.write(".bonsai/lock.json", "{\n")
	c.commit("a lock that does not read")
	c.write(".bonsai/lock.json", lock)
	e := errorIn(c.messagesDoc(4, brief...), "messages")
	if e.String("code") != "bad-lock" || !strings.HasPrefix(e.String("message"), "the lock as committed at the main checkout's HEAD does not read") {
		t.Errorf("a committed lock that does not read: %s", schema.Show(e))
	}
	c.messagesDoc(0, "messages") // the lock is read only for the roles
	problems, roles, _ = c.problemsOf()
	if !strings.Contains(strings.Join(problems, "\n"), "the lock as committed at the main checkout's HEAD does not read") || schema.Show(roles) != "[]" {
		t.Errorf("status with the committed lock unread: roles %s, problems %q", schema.Show(roles), problems)
	}
}

// bonsai message --json's thread is the one bonsai messages lists (5.5.11's follow-up): once an agent tied the brief,
// --show's thread carries its recipient (the agent, its counts and limits as the guard recorded them) and the flags
// that read its state (budget-spent while it is past its budget), through internal/agents' Store.Thread; before any
// tie, the recipient is null.
func TestMessageShowRecipient(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	start := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	old := messageNow
	t.Cleanup(func() { messageNow = old })
	messageNow = func() time.Time { return start }
	c.link()
	c.commit("link")
	_, out, _ := c.run("", "message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go.")
	id := strings.Split(out, "\n")[0]
	shown := func() schema.Object {
		t.Helper()
		th := threadsIn(c.messagesDoc(0, "message", "--show", id))
		if len(th) != 1 {
			t.Fatalf("%d threads", len(th))
		}
		return th[0]
	}
	if r, _ := shown().Get("recipient"); r != nil {
		t.Errorf("a recipient before any tie: %s", schema.Show(r))
	}
	s := agents.Store{Main: c.root}
	n := func(v int64) *int64 { return &v }
	h := agents.Header{Agent: "a1", Session: "s1", Type: "workflow:builder", Role: "builder", Pack: "workflow", Kind: "lead",
		Writes: true, Limits: agents.Limits{Actions: agents.Limit{Value: n(3), Max: n(3)}}, First: agents.Stamp(start)}
	if _, err := s.Create(h); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.Tie(id, "s1", "a1"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if _, err := s.Call("s1", "a1", start.Add(time.Minute), 1); err != nil {
		t.Fatal(err)
	}
	if err := s.Append("s1", "a1", agents.Event{E: agents.EventSpent, At: agents.StampMilli(start.Add(2 * time.Minute)),
		Measure: "actions", Count: 4, Of: 3}); err != nil {
		t.Fatal(err)
	}
	th := shown()
	r, _ := th.Get("recipient")
	f, _ := th.Get("flags")
	if ro, ok := r.(schema.Object); !ok || ro.String("agent") != "a1" || !strings.HasSuffix(schema.Show(r), `"spent":true}`) ||
		schema.Show(f) != `["budget-spent"]` {
		t.Errorf("--show's thread once tied and spent: recipient %s, flags %s", schema.Show(r), schema.Show(f))
	}
	listed := threadsIn(c.messagesDoc(0, "messages"))
	if len(listed) != 1 || schema.Show(listed[0]) != schema.Show(th) {
		t.Errorf("--show's thread is not the one bonsai messages lists:\n%s\n%s", schema.Show(th), schema.Show(listed))
	}
}
