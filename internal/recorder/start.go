package recorder

// bonsai hook start (spec §8; design/plan-5.md, 5.2.4 notes 1 and 9), on every SessionStart (startup, resume, clear,
// compact and fork alike, so /compact and /clear bring the context back): it writes the session's session_start record
// first, so one process makes the session's log file before any tool call, with the bonsai binary's path (from ~/
// under the user's home) and SHA-256, hashed on every start so a binary swapped between a session and its resume is
// seen (spec §3's tripwire); then it prints the opening context on stdout, plain ASCII, which Claude Code adds to the
// session:
//
//   - the workspace's name and id;
//   - the active task (contract §13) with its title, status, lane, branch and grants (bonsai.allows), or "none" with
//     the function's reason;
//   - that task's last ladder result (green or not, its commit, when; step 5.4 writes them), or "none yet";
//   - the label definitions attached on this machine (contract §5.3).
//
// At most MaxContextLines lines; past that, one line naming bonsai status --json. It prints what it can and exits 0:
// an unreadable bonsai.yaml gives one line saying so, and that the guard blocks edits until a person fixes it.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/timefmt"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// MaxContextLines caps the opening context (spec §8).
const MaxContextLines = 60

// Start is one run of `bonsai hook start`: the session_start record, then the opening context on stdout. It
// returns 0, always.
func Start(o Options, stdout io.Writer) int {
	ignoreBrokenPipe()
	o.defaults()
	out := make(chan string, 1)
	within(o.Budget, func() { out <- start(o) })
	select {
	case s := <-out:
		if s != "" && stdout != nil {
			_, _ = io.WriteString(stdout, s)
		}
	default:
	}
	return 0
}

// start writes the record and gives the context to print ("" for an unlinked folder).
func start(o Options) string {
	p, proj, ok := payloadAndProject(o)
	if !ok {
		return ""
	}
	cfg, err := workspace.LoadConfig(proj)
	if err != nil {
		return lines([]string{"Bonsai: bonsai.yaml cannot be read (" + err.Error() + "), so Bonsai's guard blocks every file " +
			"edit and shell command until a person fixes it (bonsai check names the line); no active task is read."})
	}
	local := workspace.FindLocal(proj, cfg.ID)
	act := &active{o: o, cfg: cfg, local: local} // the record's target and the context's task: one read of the task folder
	if p != nil && p.Event == "SessionStart" {
		if l := build(o, p, proj, cfg, local, act); l != nil {
			path, sum := self(o)
			l.BonsaiPath, l.BonsaiSHA256 = ptr(path), ptr(sum)
			_, _ = writeLog(local.Main, l)
		}
	}
	return openingContext(cfg, local, act)
}

// self gives the running binary's path as a record keeps it (from ~/ under the user's home, forward slashes) and its
// SHA-256 in hex; "" for what cannot be found (record.Self, the binary's one reader).
func self(o Options) (string, string) {
	path, sum := record.Self(o.Executable, true)
	if path == "" {
		return "", ""
	}
	if home, err := o.UserHome(); err == nil && home != "" {
		path = underHome(path, home)
	}
	return path, sum
}

// underHome writes path from ~/ when it lies under home (case-blind on Windows), else as it is, with forward slashes:
// record.FromHome, the rule's one home.
func underHome(path, home string) string { return record.FromHome(path, home) }

// openingContext is the context hook start prints (this file's comment); act is the run's active task.
func openingContext(cfg *workspace.Config, local workspace.Local, act *active) string {
	var out []string
	add := func(format string, a ...any) { out = append(out, fmt.Sprintf(format, a...)) }
	add("Bonsai: this project is the workspace %s (%s). Its log is in .bonsai/local/ (never committed).", cfg.Name, cfg.ID)
	a, err := act.get()
	switch {
	case err != nil:
		add("Active task: none (the task folder cannot be read: %s).", err.Error())
	case a.ID == "":
		add("Active task: none: %s.", strings.TrimPrefix(a.Sentence(), "no active task: "))
	default:
		t := a.Task
		add("Active task: %s %s (%s), status %s, lane %s, branch %s.", a.ID, strconv.Quote(t.Title), a.File, t.Status,
			orNone(str(t.Lane)), orNone(taskBranch(t)))
		add("Its grants (bonsai.allows): %s.", taskAllows(t))
		add("Its last ladder result: %s.", lastLadder(local, a.ID))
	}
	out = append(out, attachedLabels(local)...)
	if len(out) > MaxContextLines {
		out = append(out[:MaxContextLines-1], "More than fits here: run bonsai status --json.")
	}
	return lines(out)
}

// lines joins lines, each made one ASCII line, with a final line feed.
func lines(ls []string) string {
	var b strings.Builder
	for _, l := range ls {
		b.WriteString(oneLine(l))
		b.WriteByte('\n')
	}
	return b.String()
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

// taskBranch is the task's bonsai.branch label (contract §4), or its format-0 worktree field.
func taskBranch(t *format.Task) string {
	if v, ok := t.Labels.Get("bonsai.branch"); ok {
		s, _ := v.(string)
		return s
	}
	if t.Format0 != nil {
		v, _ := t.Format0.Get("worktree")
		s, _ := v.(string)
		return s
	}
	return ""
}

// taskAllows says the task's bonsai.allows label (contract §5.5), or its format-0 allows_assets field, as the guard
// reads it (format.Task.Allows): its globs, "none", or that it grants nothing when it is not a list of text.
func taskAllows(t *format.Task) string {
	globs, ok := t.Allows()
	if !ok {
		return "not a list of text, so it grants nothing"
	}
	var out []string
	for _, g := range globs {
		if g != "" {
			out = append(out, g)
		}
	}
	return orNone(strings.Join(out, ", "))
}

// startZone is the zone the opening context shows times in; nil means time.Local. A test sets it (Go on Windows
// ignores TZ). The context is read by an agent, which records times in UTC: so it says the zone with the offset.
var startZone *time.Location

// shownFinish is a ladder result's stored instant as the opening context prints it: in the person's zone with its
// offset (timefmt.Show), or as stored when it is not an instant.
func shownFinish(s string) string {
	t, err := timefmt.ParseUTC(s)
	if err != nil {
		return s
	}
	loc := startZone
	if loc == nil {
		loc = time.Local
	}
	return timefmt.Show(t, loc, false)
}

// lastLadder says the task's last ladder result, read from the main checkout's .bonsai/local/ladder/<task>.json
// (contract §11): green or not, its commit, when; "none yet" when there is none.
func lastLadder(local workspace.Local, task string) string {
	raw, err := os.ReadFile(filepath.Join(local.Dir(), "ladder", task+".json"))
	if os.IsNotExist(err) {
		return "none yet"
	}
	if err != nil {
		return "it cannot be read (" + err.Error() + ")"
	}
	l, err := format.ReadLadder(raw)
	if err != nil {
		return "it does not read as bonsai.ladder/1 (" + err.Error() + ")"
	}
	green := "not green"
	if l.Green {
		green = "green"
	}
	sha := l.Git.SHA
	if len(sha) > 12 {
		sha = sha[:12]
	}
	return fmt.Sprintf("%s at commit %s (%s), finished %s", green, sha, l.Mode, shownFinish(l.Finished))
}

// attachedLabels lists the label definitions attached on this machine (contract §5.3), one line each, after a line
// naming them; none when there are none.
func attachedLabels(local workspace.Local) []string {
	home, err := workspace.Home()
	if err != nil {
		return nil
	}
	lock, _ := workspace.LoadLock(local.Main)
	sets, _ := workspace.LabelsInForce(lock, home, local.Main)
	var out []string
	for _, s := range sets {
		if s.From != workspace.FromMachine {
			continue
		}
		if out == nil {
			out = append(out, "Label definitions attached on this machine (contract section 5.3):")
		}
		for _, d := range s.Labels {
			kind := d.Kind
			if len(d.Values) > 0 {
				kind += ": " + strings.Join(d.Values, ", ")
			}
			grants := ""
			if d.Grants {
				grants = ", grants"
			}
			out = append(out, fmt.Sprintf("  %s (%s; set by %s%s; version %d): %s", d.Name, kind, d.SetBy, grants, s.Version, d.Description))
		}
	}
	return out
}

// oneLine keeps a line ASCII and unbroken: printable ASCII as itself, a line break or tab as a space, any other
// character as a Go escape (spec §3: PowerShell 5.1 garbles UTF-8).
func oneLine(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteByte(' ')
		case r >= 0x20 && r < 0x7f:
			b.WriteRune(r)
		default:
			q := strconv.QuoteRuneToASCII(r)
			b.WriteString(q[1 : len(q)-1])
		}
	}
	return b.String()
}
