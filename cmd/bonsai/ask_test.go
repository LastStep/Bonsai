package main

// bonsai ask, answer and asks (step 5.2.5) through the command line, in a linked project: filing each type,
// --status, --resolve, an answer from a terminal, the asking session's answer refused, the same answer twice written
// once, Bless and a hidden character refused, every --json against bonsai.asks/1, a made-up secret in a title, an
// option and the words in no file, and nothing changed but .bonsai/local/asks/ and one ask log record for each ask
// record written. internal/asks's tests walk every rule; these hold the commands to them.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
)

const askSession = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"

// asksDoc runs an ask command with --json, holds its output to bonsai.asks/1, and gives the document's one entry.
func (c *cli) asksDoc(exit int, args ...string) (schema.Object, schema.Object) {
	c.t.Helper()
	code, out, errOut := c.run("", append(args, "--json")...)
	if code != exit {
		c.t.Fatalf("%v: exit %d, want %d\n%s%s", args, code, exit, out, errOut)
	}
	doc := fits(c.t, out, "asks")
	list, _ := doc.Get("asks")
	if l, _ := list.([]any); len(l) == 1 {
		return doc, l[0].(schema.Object)
	}
	return doc, nil
}

// localFiles gives every file of the project but .git, by path, with its bytes.
func (c *cli) localFiles() map[string]string {
	out := map[string]string{}
	for _, f := range c.files() {
		raw, _ := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(f)))
		out[f] = string(raw)
	}
	return out
}

func TestAskCommands(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	// Not linked: exit 4, not-linked naming bonsai init, the workspace null.
	code, out, errOut := c.run("", "asks")
	if code != 4 || !strings.Contains(errOut, "next: run bonsai init") {
		t.Errorf("not linked: %d %q", code, errOut)
	}
	doc, _ := c.asksDoc(4, "ask", "--status", "agent:x")
	if w, _ := doc.Get("workspace"); w != nil || errorIn(doc, "asks").String("code") != "not-linked" {
		t.Errorf("not linked: %s", schema.Show(doc))
	}
	c.link()
	c.commit("link")
	before := c.localFiles()

	// Filing each type, from a session: the key printed, the record as written.
	t.Setenv(asks.SessionEnv, askSession)
	code, out, _ = c.run("", "ask", "--type", "Answer", "--title", "Is the old theme still wanted?", "--why", "Nothing links it.")
	answerKey := strings.TrimSuffix(out, "\n")
	if code != 0 || !strings.HasPrefix(answerKey, "agent:h-") || len(answerKey) != len("agent:h-")+12 {
		t.Fatalf("an Answer: %d %q", code, out)
	}
	doc, entry := c.asksDoc(0, "ask", "--type", "Decide", "--task", "T-0901", "--title", "Which colour? api_key=MadeUpAskSecret1",
		"--why", "Both pass.\nThe check passes for both.", "--then", "The builder applies it.",
		"--option", "Lighter", "--option", "Darker password=MadeUpAskSecret2", "--key", "colour")
	written, _ := doc.Get("written")
	filed, _ := entry.Get("filed")
	if entry.String("key") != "agent:colour" || entry.String("state") != "open" || schema.Show(written) != schema.Show(filed) ||
		written.(schema.Object).String("session") != askSession {
		t.Errorf("a Decide: %s", schema.Show(doc))
	}
	if ws, _ := doc.Get("workspace"); ws.(schema.Object).String("root") != filepath.ToSlash(c.root) {
		t.Errorf("workspace %s", schema.Show(ws))
	}
	c.asksDoc(0, "ask", "--type", "Look", "--task", "T-0901", "--title", "Does the screen read well?", "--why", "A person looks.",
		"--verdict", "pass", "--key", "look")
	c.asksDoc(0, "ask", "--type", "Play", "--task", "T-0901", "--title", "Play the first level", "--why", "A person plays.", "--key", "play")

	// The asking session's answer: refused, nothing written.
	code, _, errOut = c.run("", "answer", "agent:colour", "--choice", "Lighter")
	if code != 4 || !strings.Contains(errOut, "the session that filed agent:colour cannot answer it") {
		t.Errorf("the asking session's answer: %d %q", code, errOut)
	}
	t.Setenv(asks.SessionEnv, "")

	// --status: open, then the answer.
	code, out, _ = c.run("", "ask", "--status", "agent:colour")
	if code != 0 || !strings.Contains(out, "agent:colour: open\ntype: Decide\nabout: task T-0901\nfiled: ") ||
		!strings.Contains(out, "\nwhy: Both pass.\n  The check passes for both.\n") || !strings.Contains(out, "\noption 2: Darker password=[redacted]\n") ||
		!strings.HasSuffix(out, "no answer yet\n") {
		t.Errorf("--status, open: %d\n%s", code, out)
	}
	// A terminal's answer, then the same answer again: written once.
	code, out, _ = c.run("", "answer", "agent:colour", "--choice", "Darker password=MadeUpAskSecret2", "--words", "Darker. secret=MadeUpAskSecret3")
	if code != 0 || out != "agent:colour: answered\n" {
		t.Errorf("the answer: %d %q", code, out)
	}
	doc, entry = c.asksDoc(0, "answer", "agent:colour", "--choice", "Lighter")
	if w, _ := doc.Get("written"); w != nil || entry.String("state") != "answered" {
		t.Errorf("the same answer again: %s", schema.Show(doc))
	}
	code, out, _ = c.run("", "ask", "--status", "agent:colour")
	if code != 0 || !strings.Contains(out, "agent:colour: answered\n") || !strings.Contains(out, ", by terminal\nchoice: Darker password=[redacted]\nwords: Darker. secret=[redacted]\n") {
		t.Errorf("--status, answered: %d\n%s", code, out)
	}
	_, entry = c.asksDoc(0, "ask", "--status", "agent:colour")
	closed, _ := entry.Get("closed")
	if a, _ := closed.(schema.Object).Get("answer"); a.(schema.Object).String("by") != "terminal" {
		t.Errorf("--status --json: %s", schema.Show(entry))
	}
	// Another caller's answer to it: the first answer stands.
	doc, _ = c.asksDoc(4, "answer", "agent:colour", "--by", "act:4711", "--via", "desk", "--choice", "Lighter")
	if errorIn(doc, "asks").String("code") != "ask-not-open" {
		t.Errorf("a second answer: %s", schema.Show(doc))
	}
	// A Look's verdict, by a caller, through its desk.
	_, entry = c.asksDoc(0, "answer", "agent:look", "--by", "act:4711", "--via", "desk", "--verdict", "fail", "--words", "Too dark.")
	closed, _ = entry.Get("closed")
	if a, _ := closed.(schema.Object).Get("answer"); a.(schema.Object).String("via") != "desk" || a.(schema.Object).String("verdict") != "fail" {
		t.Errorf("the Look's answer: %s", schema.Show(entry))
	}
	// Withdrawn, and withdrawn again.
	code, out, _ = c.run("", "ask", "--resolve", "agent:play")
	if code != 0 || out != "agent:play: resolved\n" {
		t.Errorf("--resolve: %d %q", code, out)
	}
	code, out, _ = c.run("", "ask", "--resolve", "agent:play")
	if code != 0 || !strings.Contains(out, "agent:play was resolved already") {
		t.Errorf("--resolve again: %d %q", code, out)
	}

	// The list: the open asks, then every key, newest first.
	code, out, _ = c.run("", "asks")
	if code != 0 || strings.Count(out, "\n") != 1 || !strings.HasPrefix(out, "open      ") || !strings.Contains(out, "  Answer  "+answerKey+"  Is the old theme") {
		t.Errorf("asks: %d\n%s", code, out)
	}
	doc, _ = c.asksDoc(0, "asks", "--all")
	list, _ := doc.Get("asks")
	var order []string
	for _, e := range list.([]any) {
		order = append(order, e.(schema.Object).String("key")+" "+e.(schema.Object).String("state"))
	}
	if strings.Join(order, ", ") != "agent:play resolved, agent:look answered, agent:colour answered, "+answerKey+" open" {
		t.Errorf("asks --all: %v", order)
	}

	// Refusals: exit 2 or 4, each with its word, nothing written.
	mid := c.localFiles()
	for _, r := range []struct {
		args []string
		exit int
		word string
	}{
		{[]string{"ask", "--type", "Bless", "--task", "T-0901", "--title", "t", "--why", "w"}, 2, "bad-value"},
		{[]string{"ask", "--type", "Answer", "--title", "a\u200bb", "--why", "w"}, 2, "bad-value"},
		{[]string{"ask", "--type", "Answer", "--title", "t", "--why", "w", "--task", "T-0901", "--doc", "M-x"}, 2, "bad-flag"},
		{[]string{"ask", "--type", "Decide", "--title", "t", "--why", "w", "--option", "a", "--option", "b", "--option", "c", "--option", "d", "--option", "e"}, 2, "bad-value"},
		{[]string{"ask", "--resolve", "agent:look", "--title", "t"}, 2, "bad-flag"},
		{[]string{"ask", "--resolve", "agent:look", "--status", "agent:look"}, 2, "bad-flag"},
		{[]string{"ask", "--status", "agent:none"}, 4, "ask-not-open"},
		{[]string{"ask", "--resolve", "agent:look"}, 4, "ask-not-open"},
		{[]string{"ask", "now"}, 2, "bad-flag"},
		{[]string{"answer"}, 2, "missing-value"},
		{[]string{"answer", "agent:look", "agent:play"}, 2, "bad-flag"},
		{[]string{"answer", "agent:play", "--words", "w"}, 4, "ask-not-open"},
		{[]string{"answer", "agent:colour"}, 2, "missing-value"},
		{[]string{"answer", "agent:look", "--choice", "Lighter", "--by", "act:9"}, 4, "ask-not-open"},
		{[]string{"asks", "now"}, 2, "bad-flag"},
	} {
		wantRefusal(t, words[r.args[0]], r.args, r.exit, r.word)
	}
	if after := c.localFiles(); !sameFiles(mid, after) {
		t.Errorf("a refusal wrote something")
	}

	// What changed: only .bonsai/local/asks/ and .bonsai/local/log/, and one ask log record for each ask record.
	after := c.localFiles()
	for f, b := range after {
		if before[f] != b && !strings.HasPrefix(f, ".bonsai/local/asks/") && !strings.HasPrefix(f, ".bonsai/local/log/") {
			t.Errorf("%s changed", f)
		}
	}
	for f := range before {
		if _, ok := after[f]; !ok {
			t.Errorf("%s went", f)
		}
	}
	var askLines, logLines []string
	for f, b := range after {
		switch {
		case strings.HasPrefix(f, ".bonsai/local/asks/"):
			for _, line := range strings.Split(strings.TrimSuffix(b, "\n"), "\n") {
				a, err := format.ReadAsk([]byte(line))
				if err != nil {
					t.Fatal(err)
				}
				askLines = append(askLines, a.At+" "+a.Op+" "+a.Key+" "+text0(a.Session))
			}
		case strings.HasPrefix(f, ".bonsai/local/log/"):
			lf, err := record.ReadLog(filepath.Join(c.root, filepath.FromSlash(f)))
			if err != nil || lf.Skipped > 0 {
				t.Fatal(f, err)
			}
			for _, l := range lf.Records {
				if l.Event != "ask" || l.Text != nil {
					t.Errorf("%s: a record %s", f, l.Event)
				}
				logLines = append(logLines, l.At+" "+text0(l.Kind)+" "+text0(l.Target)+" "+text0(l.Session))
			}
		}
	}
	sort.Strings(askLines)
	sort.Strings(logLines)
	if len(askLines) != 7 || strings.Join(askLines, "\n") != strings.Join(logLines, "\n") {
		t.Errorf("the ask records:\n%s\nthe log records:\n%s", strings.Join(askLines, "\n"), strings.Join(logLines, "\n"))
	}
	all := ""
	for _, b := range after {
		all += b
	}
	for _, s := range []string{"MadeUpAskSecret1", "MadeUpAskSecret2", "MadeUpAskSecret3"} {
		if strings.Contains(all, s) {
			t.Errorf("the made-up secret %s is in a file", s)
		}
	}
}

func sameFiles(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// A worktree's ask goes into its main checkout's .bonsai/local/, which every writer of it restores the .gitignore of.
func TestAskInAWorktree(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	c.link()
	c.commit("link")
	wt := filepath.Join(c.tmp, "task-branch")
	testpack.Git(t, c.root, "worktree", "add", "-q", "-b", "task-branch", wt)
	if err := os.Remove(filepath.Join(c.root, ".bonsai", ".gitignore")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(wt)
	code, out, errOut := c.run("", "ask", "--type", "Answer", "--title", "From the worktree", "--why", "w", "--key", "wt")
	if code != 0 || out != "agent:wt\n" {
		t.Fatalf("ask in a worktree: %d %q %q", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(wt, ".bonsai", "local")); err == nil {
		t.Errorf("the worktree has a .bonsai/local/ of its own")
	}
	if b, err := os.ReadFile(filepath.Join(c.root, ".bonsai", ".gitignore")); err != nil || !strings.Contains(string(b), "local/") {
		t.Errorf("main's .bonsai/.gitignore was not restored: %v", err)
	}
	t.Chdir(c.root)
	_, entry := c.asksDoc(0, "ask", "--status", "agent:wt")
	if entry.String("state") != "open" {
		t.Errorf("main does not see the worktree's ask: %s", schema.Show(entry))
	}
	logs, _ := filepath.Glob(filepath.Join(c.root, ".bonsai", "local", "log", "w-*.ndjson"))
	if len(logs) != 1 {
		t.Fatalf("the log files: %v", logs)
	}
	lf, err := record.ReadLog(logs[0])
	if err != nil || len(lf.Records) != 1 || text0(lf.Records[0].Checkout) != "task-branch" || text0(lf.Records[0].Branch) != "task-branch" {
		t.Errorf("the log record: %v %+v", err, lf)
	}
}
