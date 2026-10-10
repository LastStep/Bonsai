package main

// bonsai log append (spec section 4, section 8; contract section 8.4; plan-5 5.2.5 note 7): one outside event in the
// log, an event record in today's day file (.bonsai/local/log/w-<UTC date>.ndjson), session and agent null, its
// labels checked against every label definition in force: the locked packs' and those attached on this machine
// (workspace.LabelsInForce). Chosen over the machine's definitions alone (spec section 8's words): one set of
// definitions for every label check. --target and --text are redacted and capped, as every log record's are. With
// --json it prints bonsai.logs/1 with written filled (format.Logs). Its tables are logWord and logAppendWord.
//
// A label's value is read by its definition's kind (contract section 5.2): a choice one of its values; text one line,
// redacted, then held to its pattern and max; a number an integer or a decimal; a list its items separated by commas,
// each read by the list's items kind (no JSON on the command line: PowerShell 5.1 strips its quotes).

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(logWord) }

var logWord = &Word{
	Name:    "log",
	Order:   12,
	Title:   "write to the log of this project (contract section 8.4).",
	Summary: "write one outside event to the log",
	Args:    "<action>",
	About:   "bonsai logs reads the log; bonsai log append adds one outside event to it.\n",
	Exits: []Exit{
		{Code: 0, Means: "written (or the help was printed)"},
		{Code: 2, Means: "bad input: no action, one this build does not have, or log append's (its --help)"},
		{Code: 3, Means: "the record could not be written (log append)"},
		{Code: 4, Means: "not in a git checkout, not linked, or the lock unreadable (log append)"},
	},
	Examples: []string{"bonsai log append --help"},
	Subs:     []*Word{logAppendWord},
}

var logAppendWord = &Word{
	Name:    "log append",
	Title:   "one outside event in the log, its labels checked (contract section 8.4)",
	Summary: "one outside event, its labels checked against their definitions",
	Args:    "[flags]",
	About: `It appends one event record to today's day file in the main checkout's .bonsai/local/log/
(w-<UTC date>.ndjson; a worktree writes its main checkout's), with session and agent null. Each label must be
defined by a definition in force, a locked pack's or one attached on this machine (bonsai status --json lists
them), and its value of that definition's kind: a choice one of its values; text one line, redacted, then held to
its pattern and max; a number an integer or a decimal; a list its items separated by commas.
`,
	Flags: []Flag{
		{Name: "--label", Value: "NAME=VALUE", Many: true, Need: "a label, as name=value",
			Help: "a label of the event, <namespace>.<name>=<value>; one or more, each name once"},
		{Name: "--target", Value: "TEXT", Help: "what the event is about (an id, a key), redacted, cut to 200 characters"},
		{Name: "--text", Value: "TEXT", Help: "a short text, redacted, cut to 300 characters"},
		{Name: "--json", Help: "print the bonsai.logs/1 document (for programs) instead of plain text; written is the record"},
	},
	Exits: []Exit{
		{Code: 0, Means: "the record was written"},
		{Code: 2, Means: "bad input: no --label, a label not as name=value or given twice, a label no definition in force has\n" +
			"or a value of the wrong kind (label-not-defined); bonsai.yaml refused. Nothing was written"},
		{Code: 3, Means: "the record could not be written, or the Bonsai home cannot be used"},
		{Code: 4, Means: "not in a git checkout, not linked (bonsai init), or the lock unreadable"},
	},
	Examples: []string{"bonsai log append --label ops.event=deploy --label ops.ref=c6ba392 --target T-0901 --text \"deployed\" --json"},
	Refused:  func(c *call, e *engine.Error) encoder { return &format.Logs{Error: e.Object()} },
	Run:      runLogAppend,
}

// labelName is a label's name (contract section 5.1): <namespace>.<name>.
var labelName = regexp.MustCompile(`^[a-z][a-z0-9-]*\.[a-z][a-z0-9_]*$`)

// labelNumber is a number label's value: an integer or a decimal (contract section 5.2).
var labelNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)

// logAppendNow is the clock the record's at is read from; a test sets it.
var logAppendNow = time.Now

func runLogAppend(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("log append takes no %+q", redact.Text(c.rest[0])))
	}
	given := c.all("--label")
	if len(given) == 0 {
		return c.refuse(&engine.Error{Code: "missing-value", Exit: exitInput, What: "log append needs at least one --label",
			Next: "give the event's labels: --label <namespace>.<name>=<value>"})
	}
	type pair struct{ name, value string }
	var pairs []pair
	seen := map[string]bool{}
	for _, l := range given {
		name, value, ok := strings.Cut(l, "=")
		switch {
		case !ok:
			return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: fmt.Sprintf("--label %+q is not name=value", redact.Text(l)),
				Next: "give each label as --label <namespace>.<name>=<value>"})
		case seen[name]:
			return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: fmt.Sprintf("the label %+q is given twice", redact.Text(name)),
				Next: "give each label once"})
		}
		seen[name] = true
		pairs = append(pairs, pair{name, value})
	}
	site, e := askSite()
	if e != nil {
		return c.fail(e)
	}
	lock, err := workspace.LoadLock(site.co.Root)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return c.fail(lockError(err))
		}
		lock = nil
	}
	home, err := workspace.Home()
	if err != nil {
		return c.fail(engine.HomeError(err))
	}
	sets, problems := workspace.LabelsInForce(lock, home, site.co.Main)
	defs := map[string]format.LabelDef{}
	var names []string
	for _, s := range sets {
		for _, d := range s.Labels {
			if _, twice := defs[d.Name]; !twice {
				defs[d.Name] = d // a name defined twice is check's finding; the first source's stands (contract section 5.1)
				names = append(names, d.Name)
			}
		}
	}
	labels := schema.Object{}
	for _, p := range pairs {
		d, ok := defs[p.name]
		if !ok || !labelName.MatchString(p.name) {
			what := fmt.Sprintf("no label definition in force has %+q", redact.Text(p.name))
			if len(names) > 0 {
				what += " (in force: " + strings.Join(names, ", ") + ")"
			} else {
				what += " (none is in force)"
			}
			if len(problems) > 0 {
				what += "; and " + engine.ASCII(problems[0].Error())
			}
			return c.fail(&engine.Error{Code: "label-not-defined", Exit: exitInput, What: what,
				Next: "use a label a locked pack or this machine defines (bonsai status --json lists the sets in force)"})
		}
		v, why := labelValue(d, p.value)
		if why != "" {
			return c.fail(&engine.Error{Code: "label-not-defined", Exit: exitInput,
				What: fmt.Sprintf("the label %s's value %+q is %s (its definition: %s)", p.name, redact.Text(p.value), why, d.Kind),
				Next: "give " + p.name + " a value its definition takes"})
		}
		labels = append(labels, schema.Member{Key: p.name, Value: v})
	}
	l := record.New("event", record.Common{Workspace: site.cfg.ID, Local: site.place.Local, Redact: redact.Text,
		At: logAppendNow()})
	l.Labels = labels
	if t := c.value("--target"); t != "" {
		t = redact.Capped(t, redact.TargetCap)
		l.Target = &t
	}
	if t := c.value("--text"); t != "" {
		t = redact.Capped(t, redact.TextCap)
		l.Text = &t
	}
	line, err := record.Line(l)
	if err != nil {
		return c.fail(engine.Unexpected(err))
	}
	main := site.place.Local.Main
	if _, err := workspace.EnsureGitignore(main); err != nil {
		return c.fail(&engine.Error{Code: "write-failed", Exit: exitRuntime,
			What: workspace.GitignoreFile + " is missing and cannot be written, so nothing is written: " + engine.ASCII(err.Error()),
			Next: "check that .bonsai/ can be written, then run the command again"})
	}
	if _, err := record.WriteLog(main, l); err != nil {
		return c.fail(&engine.Error{Code: "write-failed", Exit: exitRuntime,
			What: "the event record cannot be written: " + engine.ASCII(err.Error()),
			Next: "check that .bonsai/local/log/ can be written, then run the command again"})
	}
	if c.json {
		v, err := schema.Decode(line)
		written, ok := v.(schema.Object)
		if err != nil || !ok {
			return c.fail(engine.Unexpected(fmt.Errorf("the record written does not read back: %v", err)))
		}
		return c.printDoc(&format.Logs{Workspace: site.ref, Written: written}, exitOK)
	}
	at, _ := time.Parse(record.AtLayout, l.At)
	name, _ := record.LogFileName("", at)
	return c.print("written: " + filepath.ToSlash(filepath.Join(workspace.LocalDir, record.LogFolder, name)) + " (an event record, " +
		plural(len(labels), "label") + ")\n")
}

// labelValue reads a label's value from the command line by its definition's kind (contract section 5.2): the value
// as the record holds it, or why it does not fit ("" when it fits). Text is redacted before it is held to the
// definition, as every free-text string Bonsai writes is.
func labelValue(d format.LabelDef, raw string) (any, string) {
	switch d.Kind {
	case "choice":
		for _, v := range d.Values {
			if raw == v {
				return raw, ""
			}
		}
		return nil, "not one of " + strings.Join(d.Values, ", ")
	case "text":
		return labelText(d, redact.Text(raw))
	case "number":
		if !labelNumber.MatchString(raw) {
			return nil, "not a number (an integer or a decimal)"
		}
		return json.Number(raw), ""
	case "list":
		items := []any{}
		if raw != "" {
			for _, it := range strings.Split(raw, ",") {
				if d.Items != nil && *d.Items == "number" {
					if !labelNumber.MatchString(it) {
						return nil, fmt.Sprintf("a list whose item %+q is not a number", it)
					}
					items = append(items, json.Number(it))
					continue
				}
				it = redact.Text(it)
				if it == "" || strings.ContainsAny(it, "\r\n") {
					return nil, "a list with an empty item, or one that is not one line of text"
				}
				items = append(items, it)
			}
		}
		if d.Max != nil && int64(len(items)) > *d.Max {
			return nil, fmt.Sprintf("a list of %d items, more than %d", len(items), *d.Max)
		}
		return items, ""
	}
	return nil, "of a kind this Bonsai does not know (" + d.Kind + ")"
}

// labelText holds a text value, redacted, to its definition: one line, its max, its pattern (whole).
func labelText(d format.LabelDef, s string) (any, string) {
	switch {
	case strings.ContainsAny(s, "\r\n"):
		return nil, "not one line"
	case d.Max != nil && int64(len([]rune(s))) > *d.Max:
		return nil, fmt.Sprintf("longer than %d characters", *d.Max)
	case d.Pattern != nil:
		re, err := regexp.Compile(*d.Pattern)
		if err != nil {
			return nil, "not checkable: its definition's pattern does not compile"
		}
		if loc := re.FindStringIndex(s); loc == nil || loc[0] != 0 || loc[1] != len(s) {
			return nil, "not matching " + *d.Pattern
		}
	}
	return s, ""
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}
