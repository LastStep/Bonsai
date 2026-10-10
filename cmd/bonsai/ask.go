package main

// bonsai ask (spec section 4, section 8; contract section 9; plan-5 5.2.5): an agent files a typed ask for a person,
// withdraws it (--resolve) or reads where it stands (--status). A thin command over internal/asks, which holds every
// rule; it reads the checkout and bonsai.yaml, and with --json prints bonsai.asks/1 (format.Asks). Its table is
// askWord. The asks themselves live in the main checkout's .bonsai/local/asks/, never committed.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(askWord) }

// asksNow is the clock an ask, answer or resolve record's at is read from; a test sets it.
var asksNow = time.Now

// askFiling are the flags that file an ask, which --resolve and --status do not take.
var askFiling = []string{"--type", "--title", "--why", "--then", "--task", "--doc", "--option", "--verdict", "--key"}

var askWord = &Word{
	Name:    "ask",
	Order:   8,
	Title:   "file a typed ask for a person, withdraw it, or read where it stands (contract section 9).",
	Summary: "file an ask for a person; --status reads its answer",
	Args:    "[flags]",
	About: `An agent asks a person something typed: Answer (a question), Decide (a choice among options), Look (a person
looks at the task's result) or Play (a person plays it). Bless is the ladder's alone. It appends one ask record to
the main checkout's .bonsai/local/asks/<UTC day>.ndjson (a worktree writes its main checkout's) and one ask record
to the log, and prints the ask's key. Every text is checked before anything is written: a line feed only in --why,
no hidden or control character (refused, not stripped), every secret taken out by the redactor, then the limit,
refused when over (never cut). CLAUDE_CODE_SESSION_ID names the asking session, whose own answer is refused.
`,
	Flags: []Flag{
		{Name: "--type", Value: "TYPE", Help: "the ask's type: Answer, Decide, Look or Play (Bless is the ladder's)"},
		{Name: "--title", Value: "TEXT", Help: "the question, one line, at most 300 characters"},
		{Name: "--why", Value: "TEXT", Help: "why it is asked, at most 600 characters (it may hold line feeds)"},
		{Name: "--then", Value: "TEXT", Help: "what happens once it is answered, one line, at most 300 characters"},
		{Name: "--task", Value: "ID", Help: "the task it is about (Look and Play need one); an id, never a path"},
		{Name: "--doc", Value: "ID", Help: "or the document it is about, by an id a declared kind's pattern matches; not with --task"},
		{Name: "--option", Value: "TEXT", Many: true, Help: "a Decide's choice, one line, at most 200 characters; up to four, each different"},
		{Name: "--verdict", Value: "pass|fail", Help: "a Look's own verdict"},
		{Name: "--key", Value: "KEY", Help: "the ask's key is agent:KEY ([A-Za-z0-9][A-Za-z0-9._-], at most 60); without it, a key\n" +
			"made from the type, the task or doc and the title, so the same ask filed twice is one ask"},
		{Name: "--resolve", Value: "KEY", Help: "withdraw an open ask; it takes no other flag but --json"},
		{Name: "--status", Value: "KEY", Help: "print where an ask stands and, once answered, the answer; it takes no other flag but --json"},
		{Name: "--json", Help: "print the bonsai.asks/1 document (for programs) instead of plain text"},
	},
	Notes: `Filing an open key again appends the new wording; filing a closed one opens it again. An agent reads the answer
with --status; an answer never grants anything (contract section 9.3).
`,
	Exits: []Exit{
		{Code: 0, Means: "filed (the key is printed), withdrawn, read, or nothing to do (an ask resolved already)"},
		{Code: 2, Means: "bad input: a flag ask does not take, a missing or wrong value, a rule of contract section 9 broken\n" +
			"(the message names it); bonsai.yaml refused. Nothing was written"},
		{Code: 3, Means: "the ask could not be written or read (partly-written: the ask record stands, its log record not)"},
		{Code: 4, Means: "not in a git checkout or not linked (bonsai init); a key with no ask, or --resolve on an answered\n" +
			"ask (ask-not-open); the lock unreadable when --doc needs the packs' kinds. Nothing was written"},
	},
	Examples: []string{
		`bonsai ask --type Decide --task T-0901 --title "Which colour for links?" --why "Both pass the contrast check." --option "Lighter blue" --option "Darker blue"`,
		"bonsai ask --status agent:h-3e9f0a1b2c4d --json",
	},
	Refused: asksRefused,
	Run:     runAsk,
}

// asksRefused is the asks document of a refusal: the workspace null, asks [], the error filled.
func asksRefused(c *call, e *engine.Error) encoder {
	return &format.Asks{Asks: []format.AskEntry{}, Error: e.Object()}
}

func runAsk(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("ask takes no %+q", redact.Text(c.rest[0])))
	}
	mode := ""
	for _, m := range []string{"--resolve", "--status"} {
		if !c.has(m) {
			continue
		}
		if mode != "" {
			return c.refuse(c.flagError("ask takes --resolve or --status, not both"))
		}
		mode = m
	}
	for _, f := range askFiling {
		if mode != "" && c.has(f) {
			e := c.flagError("ask %s takes only a key (and --json), not %s", mode, f)
			e.Next = "run `bonsai ask " + mode + " <key>` alone"
			return c.refuse(e)
		}
	}
	site, e := askSite()
	if e != nil {
		return c.fail(e)
	}
	switch mode {
	case "--status":
		entry, ae := asks.Status(site.place, c.value("--status"))
		if ae != nil {
			return c.fail(askError(ae))
		}
		return c.asksOut(site, &asks.Done{Entry: entry}, statusText(entry))
	case "--resolve":
		done, ae := asks.Resolve(site.place, c.value("--resolve"))
		if ae != nil {
			return c.fail(askError(ae))
		}
		return c.asksOut(site, done, doneText(done, "resolved"))
	}
	kinds, e := site.kinds(c.value("--doc") != "")
	if e != nil {
		return c.fail(e)
	}
	done, ae := asks.File(site.place, asks.Filing{Type: c.value("--type"), Title: c.value("--title"), Why: c.value("--why"),
		Then: c.value("--then"), Task: c.value("--task"), Doc: c.value("--doc"), Options: c.all("--option"),
		Verdict: c.value("--verdict"), Key: c.value("--key")}, kinds)
	if ae != nil {
		return c.fail(askError(ae))
	}
	return c.asksOut(site, done, done.Entry.Key+"\n")
}

// asksSite is where an ask, answer or asks command runs: its checkout, bonsai.yaml as read, and where the asks are
// kept.
type asksSite struct {
	co    *workspace.Checkout
	cfg   *workspace.Config
	place asks.Place
	ref   *format.WorkspaceRef
}

// askSite finds the checkout the command runs in (git), reads its bonsai.yaml in full, and the main checkout whose
// .bonsai/local/ holds the asks (workspace.FindLocal).
func askSite() (*asksSite, *engine.Error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, &engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"}
	}
	co, err := workspace.Find(dir)
	if err != nil {
		return nil, engine.FindError(err)
	}
	cfg, err := workspace.LoadConfigFull(co.Root)
	if err != nil {
		return nil, engine.ConfigError(err)
	}
	local := workspace.FindLocal(co.Root, cfg.ID)
	return &asksSite{co: co, cfg: cfg,
		place: asks.Place{Workspace: cfg.ID, Local: local, Session: os.Getenv(asks.SessionEnv), Now: asksNow},
		ref:   &format.WorkspaceRef{ID: cfg.ID, Name: cfg.Name, Root: filepath.ToSlash(local.Main)}}, nil
}

// kinds are the document kinds an ask's ids are matched against: Bonsai's own, and with packs the locked packs'
// (their declares, in the lock).
func (s *asksSite) kinds(packs bool) ([]workspace.DocKind, *engine.Error) {
	var lock *workspace.Lock
	if packs {
		l, err := workspace.LoadLock(s.co.Root)
		switch {
		case err == nil:
			lock = l
		case !errors.Is(err, fs.ErrNotExist):
			return nil, lockError(err)
		}
	}
	kinds, err := workspace.DocKinds(s.cfg.Full, lock)
	if err != nil {
		return nil, lockError(err)
	}
	return kinds, nil
}

// lockError is a lock Bonsai cannot read where a command needs the packs' declarations (exit 4).
func lockError(err error) *engine.Error {
	e := &engine.Error{Code: "bad-lock", Exit: engine.ExitState, What: ".bonsai/lock.json cannot be read: " + engine.ASCII(err.Error()),
		Next: "restore the lock from git (git checkout -- .bonsai/lock.json), or run bonsai update"}
	var we *workspace.Error
	if errors.As(err, &we) {
		e.What, e.Next = engine.ASCII(we.File+": "+we.Msg), we.Next
	}
	return e
}

// askError is internal/asks's refusal as the command's error.
func askError(e *asks.Error) *engine.Error {
	return &engine.Error{Code: e.Code, Exit: e.Exit, What: e.What, Next: e.Next, Who: e.Who}
}

// asksOut prints a command's answer: with --json the asks document (the key's entry, and the record written), else
// text.
func (c *call) asksOut(site *asksSite, done *asks.Done, text string) int {
	if c.json {
		doc := &format.Asks{Workspace: site.ref, Asks: []format.AskEntry{done.Entry.Entry()}, Written: done.Written}
		return c.printDoc(doc, exitOK)
	}
	return c.print(text)
}

// doneText is the human line of a resolve or an answer: the key and what was done, or why nothing was.
func doneText(done *asks.Done, did string) string {
	if done.Nothing != "" {
		return engine.ASCII(done.Nothing) + "\n"
	}
	return done.Entry.Key + ": " + did + "\n"
}

// statusText is an ask as ask --status prints it: its state, the ask, and the record that closed it.
func statusText(e *asks.Entry) string {
	var b strings.Builder
	f := e.Filed
	line := func(label, s string) {
		if s == "" {
			return
		}
		lines := strings.Split(s, "\n")
		b.WriteString(label + ": " + engine.ASCII(lines[0]) + "\n")
		for _, l := range lines[1:] {
			b.WriteString("  " + engine.ASCII(l) + "\n")
		}
	}
	b.WriteString(e.Key + ": " + e.State + "\n")
	about := "the answers file (no task or document)"
	switch {
	case f.Task != nil:
		about = "task " + *f.Task
	case f.Doc != nil:
		about = "document " + *f.Doc
	}
	session := "no session"
	if f.Session != nil {
		session = "session " + *f.Session
	}
	line("type", text0(f.Type))
	line("about", about)
	line("filed", f.At+", "+session)
	line("title", text0(f.Title))
	line("why", text0(f.Why))
	line("then", text0(f.ThenText))
	for i, o := range f.Options {
		line("option "+string(rune('1'+i)), o)
	}
	line("verdict", text0(f.Verdict))
	if e.Closed == nil {
		if e.State == asks.Open {
			b.WriteString("no answer yet\n")
		}
		return b.String()
	}
	if e.Closed.Op == asks.OpResolve {
		line("resolved", e.Closed.At)
		return b.String()
	}
	if a := e.Closed.Answer; a != nil {
		via := ""
		if a.Via != nil {
			via = ", via " + *a.Via
		}
		line("answered", e.Closed.At+", by "+a.By+via)
		line("choice", text0(a.Choice))
		line("their verdict", text0(a.Verdict))
		line("words", text0(a.Words))
	}
	return b.String()
}

func text0(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
