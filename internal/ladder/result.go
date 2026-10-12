package ladder

// The result and what follows it (notes 6, 7 and 9): bonsai.ladder/1 written through its writer at the main checkout's
// .bonsai/local/ladder/<task>.json (ci.json under --ci), whole; the fingerprint, one ladder record in the log naming
// the runner's own path and SHA-256; the cleaner's ladder kind, never the result just written; the Bless step
// (kinds.go); and the climb document, bonsai.climb/1.

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"

	"github.com/LastStep/Bonsai/internal/clean"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// ProofSentence is a --ci result's proof (today's words): a CI result is never a task's proof.
const ProofSentence = "none: a CI result is never a task's proof; a local climb is"

// timeLayout is a result's started and finished: UTC, to the millisecond (contract §11, as the log's at).
const timeLayout = record.AtLayout

// writeFile writes the result's bytes whole (a variable for the tests: a result that cannot be written).
var writeFile = workspace.WriteFileAtomic

// writeLog appends the fingerprint record (a variable for the tests: a record that cannot be written).
var writeLog = record.WriteLog

// userHome is the user's home folder, which the record's bonsai_path is written from (a variable for the tests).
var userHome = os.UserHomeDir

// executable finds the runner's own binary (a variable for the tests).
var executable = os.Executable

// result builds the climb's result (contract §11): every field, free text redacted where it was read.
func (c *Climb) result(finished string) *format.Ladder {
	g := c.git
	if c.dirtyAfter || (c.headAfter != g.SHA) || len(c.hides) > 0 {
		g.Dirty = true
	}
	green := len(c.results) > 0
	for _, r := range c.results {
		if r.Status != statusGreen {
			green = false
		}
	}
	l := &format.Ladder{Workspace: c.Config.ID, Mode: c.mode(), Started: c.started.UTC().Format(timeLayout), Finished: finished,
		Git: g, Requested: append([]int64{}, c.Requested...), Green: green, Rungs: c.results, Skipped: c.skipped,
		Leftovers: c.leftovers}
	if l.Rungs == nil {
		l.Rungs = []format.LadderRung{}
	}
	if l.Skipped == nil {
		l.Skipped = []format.LadderSkip{}
	}
	if c.Task != nil {
		id := c.Task.ID
		l.Task = &id
	}
	if c.CI {
		p := ProofSentence
		l.Proof = &p
	}
	return l
}

// finish writes the result, its record, the cleaning and the Bless step, prints the summary, and gives the outcome.
func (c *Climb) finish() *Answer {
	finished := c.Now().UTC()
	l := c.result(finished.Format(timeLayout))
	raw, err := l.Encode()
	if err != nil {
		return refused(c, unexpected("the ladder result did not fit bonsai.ladder/1", err))
	}
	rel := c.ResultRel()
	if _, err := workspace.EnsureGitignore(c.Main); err != nil {
		c.printNote("note: " + workspace.GitignoreFile + " could not be restored: " + oneLine(err.Error()))
	}
	if err := writeFile(filepath.Join(c.Main, filepath.FromSlash(rel)), raw); err != nil {
		e := &engine.Error{Code: "write-failed", Exit: ExitRuntime,
			What: "the ladder result " + rel + " cannot be written in the main checkout (" + oneLine(err.Error()) + "), so this climb proves nothing",
			Next: "check that " + shown(filepath.Join(c.Main, ".bonsai", "local")) + " can be written; then run: bonsai ladder " + c.taskFlag() + " --json"}
		return refused(c, e)
	}
	sum := sha256.Sum256(raw)
	w := &Written{Rel: rel, SHA256: hex.EncodeToString(sum[:]), Result: l}
	doc := c.doc(w)
	var fail *engine.Error
	if err := c.fingerprint(w, finished); err != nil {
		fail = &engine.Error{Code: "write-failed", Exit: ExitRuntime,
			What: "the ladder result " + rel + " was written, but its ladder record cannot be written to the log (" + oneLine(err.Error()) +
				"), and a result with no record proves nothing",
			Next: "check that " + shown(filepath.Join(c.Main, ".bonsai", "local", "log")) + " can be written; then run: bonsai ladder " +
				c.taskFlag() + " --json"}
	}
	done := clean.LadderResults(clean.Options{Local: c.local(), Getenv: c.Getenv, Now: c.Now, Keep: []string{rel}})
	for _, it := range done.Cleaned {
		doc.Cleaned = append(doc.Cleaned, it.Target)
	}
	for _, err := range done.Errs {
		c.printNote("note: the cleaner: " + oneLine(err.Error()))
	}
	doc.Bless = afterResult(c, w)
	if doc.Bless == nil {
		doc.Bless = []schema.Object{}
	}
	exit := ExitGreen
	if !l.Green {
		exit = ExitRed
		doc.Next = c.next()
	}
	c.printSummary(l, w, doc)
	if fail != nil {
		doc.Next = nil
		doc.Error = fail.Object()
		return &Answer{Doc: doc, Exit: fail.Exit, Error: fail}
	}
	return &Answer{Doc: doc, Exit: exit}
}

// local is where the climb's records go: the climbed checkout, writing the main checkout's .bonsai/local/.
func (c *Climb) local() workspace.Local {
	return workspace.Local{Root: c.Root, Main: c.Main, Branch: workspace.ReadGitLayout(c.Root).Branch()}
}

// fingerprint appends the ladder record (note 7, the log schema's ladder record): target the result's path, text
// sha256:<hex>, ok its green, task its task, kind its mode, category Ladder, and bonsai_path and bonsai_sha256 the
// runner's own (bonsai_path from ~/ under the user's home); in the session's log file when CLAUDE_CODE_SESSION_ID names
// one Bonsai can name a file after, else the day's file.
func (c *Climb) fingerprint(w *Written, at time.Time) error {
	session := c.Getenv(SessionEnv)
	if _, err := record.LogFileName(session, at); err != nil {
		session = ""
	}
	agent := ""
	if session != "" {
		agent = "claude-code"
	}
	l := record.New("ladder", record.Common{Workspace: c.Config.ID, Local: c.local(), Session: session, Agent: agent,
		Getenv: c.Getenv, Redact: redact.Text, At: at})
	l.Category = strPtr("Ladder")
	l.Target = strPtr(w.Rel)
	ok := w.Result.Green
	l.OK = &ok
	l.Kind = strPtr(w.Result.Mode)
	l.Text = strPtr("sha256:" + w.SHA256)
	l.Task = nil
	if w.Result.Task != nil {
		l.Task = strPtr(redact.Text(*w.Result.Task))
	}
	path, sum := record.Self(executable, true)
	home, _ := userHome()
	l.BonsaiPath, l.BonsaiSHA256 = strPtr(record.FromHome(path, home)), strPtr(sum)
	_, err := writeLog(c.Main, l)
	return err
}

// doc is the climb document of a climb that wrote its result (bonsai.climb/1).
func (c *Climb) doc(w *Written) *format.Climb {
	d := &format.Climb{Workspace: c.workspace, Mode: strPtr(c.mode()), Result: strPtr(w.Rel), SHA256: strPtr(w.SHA256),
		Bless: []schema.Object{}, Cleaned: []string{}}
	if c.Task != nil {
		d.Task = strPtr(c.Task.ID)
	}
	if lo, err := format.MustLookup("ladder").Document(w.Result); err == nil {
		d.Ladder = lo
	}
	return d
}

// refused is the outcome of a climb that refused or failed before it wrote a result: the climb document with what was
// known filled (the workspace, the task, the mode), its error object, and nothing else.
func refused(c *Climb, e *engine.Error) *Answer {
	d := &format.Climb{Bless: []schema.Object{}, Cleaned: []string{}, Error: e.Object()}
	if c != nil {
		d.Workspace = c.workspace
		if c.Task != nil {
			d.Task = strPtr(c.Task.ID)
		}
		if c.workspace != nil {
			d.Mode = strPtr(c.mode())
		}
	}
	return &Answer{Doc: d, Exit: e.Exit, Error: e}
}

// Refused is the climb document of a refusal before the climb read anything (the command line): every field null or
// empty but the error object.
func Refused(e *engine.Error) *format.Climb { return refused(nil, e).Doc }
