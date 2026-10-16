package main

// bonsai labels attach and detach (spec section 4; contract sections 5.1-5.3; plan-5 5.6.0 and 5.6.2): label
// definitions attached on this machine for one workspace, a bonsai.labels/1 file copied into the workspace's machine
// folder (labels/<namespace>.yaml in the Bonsai home), or taken out. Step 5.6.0 placed the word and its tables; step
// 5.6.2 builds both sub-words (the checks and the writing are internal/workspace/attach.go's). With --json each prints
// bonsai.attach/1 (format.Attach). Both refuse themselves in an agent's session (personOnly, agent.go), before they
// read or write anything.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/textui"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func init() { register(labelsWord) }

var labelsWord = &Word{
	Name:    "labels",
	Order:   17,
	Title:   "label definitions attached on this machine for this workspace (contract section 5.3).",
	Summary: "attach or detach label definitions on this machine",
	Args:    "<action>",
	About: "bonsai labels attach copies a bonsai.labels/1 file into the workspace's machine folder; bonsai labels detach\n" +
		"takes it out. Both are a person's commands. bonsai status --json lists what is attached (labels, from machine).\n",
	Exits: []Exit{
		{Code: 0, Means: "done (or the help was printed)"},
		{Code: 2, Means: "bad input: no action, one this build does not have, or the action's own (its --help)"},
		{Code: 3, Means: "a file could not be read or written (the action's)"},
		{Code: 4, Means: "not linked, a namespace taken or a definition redefined, or an agent's session (the action's)"},
	},
	Examples: []string{"bonsai labels attach --help"},
	Subs:     []*Word{labelsAttachWord, labelsDetachWord},
}

// labelsRefused gives a labels sub-word's --json document for a refusal.
func labelsRefused(command string) func(c *call, e *engine.Error) encoder {
	return func(c *call, e *engine.Error) encoder { return &format.Attach{Command: command, Error: e.Object()} }
}

var labelsAttachWord = &Word{
	Name:    "labels attach",
	Title:   "attach a file of label definitions on this machine: a person's command",
	Summary: "attach a bonsai.labels/1 file (a person's command)",
	Args:    "<file> [flags]",
	About: `The file is read as bonsai.labels/1 and copied, its line endings made LF, to the workspace's machine folder
as labels/<namespace>.yaml. A namespace that is bonsai or a locked pack's is refused (namespace-taken). Attached
again, the same content is unchanged, and a newer version that only adds definitions is updated; anything else (a
definition changed or removed, other content at the same or an older version) is refused (label-redefined): a
definition may be added, never redefined, so a person detaches the namespace first.
In an agent's session (CLAUDE_CODE_CHILD_SESSION set) it refuses itself (person-command, exit 4) and writes nothing:
a person types it in a terminal of their own, not inside Claude (its ! lines count as an agent's session).
`,
	Flags: []Flag{{Name: "--json", Help: "print the bonsai.attach/1 document (for programs) instead of plain text"}},
	Exits: []Exit{
		{Code: 0, Means: "attached, updated, or unchanged"},
		{Code: 2, Means: "bad input: no file, or one that does not read as bonsai.labels/1 (bad-value, naming the field),\n" +
			"or a flag it does not take. Nothing was written"},
		{Code: 3, Means: "the file could not be read, or the machine folder's copy written (read-failed, write-failed)"},
		{Code: 4, Means: "not linked (not-linked), a namespace taken (namespace-taken), a definition redefined\n" +
			"(label-redefined), or an agent's session (person-command). Nothing was written"},
	},
	Examples: []string{"bonsai labels attach studio-labels.yaml --json"},
	Refused:  labelsRefused("attach"),
	Run:      runLabelsAttach,
}

var labelsDetachWord = &Word{
	Name:    "labels detach",
	Title:   "detach a namespace's label definitions from this machine: a person's command",
	Summary: "detach a namespace's definitions (a person's command)",
	Args:    "<namespace> [flags]",
	About: `It removes the workspace's machine folder's labels/<namespace>.yaml (detached), or does nothing when there is
none (nothing). It runs in a checkout whose bonsai.yaml is gone too, so a project taken out can be unregistered.
In an agent's session (CLAUDE_CODE_CHILD_SESSION set) it refuses itself (person-command, exit 4) and removes nothing:
a person types it in a terminal of their own, not inside Claude (its ! lines count as an agent's session).
`,
	Flags: []Flag{{Name: "--json", Help: "print the bonsai.attach/1 document (for programs) instead of plain text"}},
	Exits: []Exit{
		{Code: 0, Means: "detached, or nothing to detach"},
		{Code: 2, Means: "bad input: no namespace, one not in a namespace's form (bad-value), or a flag it does not take"},
		{Code: 3, Means: "the file could not be removed (write-failed)"},
		{Code: 4, Means: "not in a git checkout (not-a-checkout), or an agent's session (person-command). Nothing was removed"},
	},
	Examples: []string{"bonsai labels detach studio --json"},
	Refused:  labelsRefused("detach"),
	Run:      runLabelsDetach,
}

// labelsDoc is a labels sub-word's whole --json document for a result.
func labelsDoc(command string, o *workspace.AttachOutcome) *format.Attach {
	ns, file, result := o.Namespace, o.File, o.Result
	a := &format.Attach{Command: command, Namespace: &ns, File: &file, Result: &result, Labels: o.Labels}
	if o.Version > 0 {
		v := o.Version
		a.Version = &v
	}
	return a
}

// labelsRefusal gives an attach or detach failure as an engine error and the document's known fields.
func labelsRefusal(command string, err error) (*engine.Error, *format.Attach) {
	var ae *workspace.AttachError
	if !errors.As(err, &ae) {
		var we *workspace.Error
		if errors.As(err, &we) {
			return &engine.Error{Code: "read-failed", Exit: exitRuntime, What: we.File + ": " + we.Msg, Next: we.Next}, nil
		}
		return engine.Unexpected(err), nil
	}
	exit := map[string]int{workspace.AttachBadValue: exitInput, workspace.AttachTaken: engine.ExitState,
		workspace.AttachRedefined: engine.ExitState}[ae.Code]
	if exit == 0 {
		exit = exitRuntime
	}
	e := &engine.Error{Code: ae.Code, Exit: exit, What: ae.Msg, Next: ae.Next}
	a := &format.Attach{Command: command}
	if ae.Namespace != "" {
		ns := ae.Namespace
		a.Namespace = &ns
	}
	if ae.Version > 0 {
		v := ae.Version
		a.Version = &v
	}
	return e, a
}

// labelsFail prints a failure of attach or detach: with --json the document, its error filled and what is known
// (the namespace, the version) kept; else the refusal block.
func labelsFail(c *call, command string, err error) int {
	e, a := labelsRefusal(command, err)
	if a == nil || !c.json {
		return c.refuse(e)
	}
	noted(c.word, e)
	a.Error = e.Object()
	return c.printDoc(a, e.Exit)
}

// labelsShow prints the human answer of a result (style C) or, with --json, the document.
func labelsShow(c *call, command string, o *workspace.AttachOutcome) int {
	if c.json {
		return c.printDoc(labelsDoc(command, o), exitOK)
	}
	var d textui.Doc
	title := "bonsai labels " + command
	switch o.Result {
	case workspace.AttachAttached:
		d.Title(title, "attached "+o.Namespace+", version "+itoa64(o.Version))
		d.Section("Attached", textui.R(textui.T(o.File, textui.Path), textui.T(plural(len(o.Labels), "label"), textui.New)))
	case workspace.AttachUpdated:
		d.Title(title, "updated "+o.Namespace+" to version "+itoa64(o.Version))
		d.Section("Updated", textui.R(textui.T(o.File, textui.Path), textui.T(plural(len(o.Labels), "label"), textui.New)))
	case workspace.AttachUnchanged:
		d.Title(title, o.Namespace+" is already attached as given: nothing written")
		d.Section("Attached", textui.R(textui.T(o.File, textui.Path), textui.C(plural(len(o.Labels), "label"))))
	case workspace.AttachDetached:
		d.Title(title, "detached "+o.Namespace)
		d.Section("Removed", textui.R(textui.T(o.File, textui.Path)))
	default:
		d.Title(title, "nothing to detach: "+o.Namespace+" has no definitions attached")
	}
	if o.Result == workspace.AttachAttached || o.Result == workspace.AttachUpdated || o.Result == workspace.AttachUnchanged {
		d.Hint("bonsai status --json lists them (labels, from machine)")
	}
	return c.show(&d)
}

func itoa64(n int64) string { return fmt.Sprint(n) }

// labelsHome gives the Bonsai home, or the failure to report.
func labelsHome() (string, *engine.Error) {
	home, err := workspace.Home()
	if err != nil {
		return "", engine.HomeError(err)
	}
	return home, nil
}

// runLabelsAttach is labels attach: a person's command, refused in an agent's session before anything is read; then
// the checkout must be linked, the file is read, checked and copied (workspace.AttachLabels).
func runLabelsAttach(c *call) int {
	if e := personOnly("bonsai labels attach", "it attaches label definitions on this machine", c.line()); e != nil {
		return c.refuse(e)
	}
	switch {
	case len(c.rest) == 0:
		return c.refuse(&engine.Error{Code: "missing-value", Exit: exitInput, What: "labels attach needs the file to attach",
			Next: "run: bonsai labels attach <file>"})
	case len(c.rest) > 1:
		return c.refuse(c.flagError("labels attach takes one file, not %d", len(c.rest)))
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.refuse(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"})
	}
	co, err := workspace.Find(dir)
	if err != nil {
		return c.refuse(engine.FindError(err))
	}
	if _, err := workspace.LoadConfig(co.Root); err != nil {
		return c.refuse(engine.ConfigError(err))
	}
	lock, err := workspace.LoadLock(co.Root)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return c.refuse(lockError(err))
		}
		lock = nil
	}
	home, he := labelsHome()
	if he != nil {
		return c.refuse(he)
	}
	file := c.rest[0]
	raw, err := os.ReadFile(file)
	if err != nil {
		return c.refuse(&engine.Error{Code: "read-failed", Exit: exitRuntime,
			What: engine.ASCII(filepath.ToSlash(file)) + " cannot be read: " + engine.ASCII(err.Error()),
			Next: "give a file that exists and can be read"})
	}
	out, err := workspace.AttachLabels(lock, home, co.Main, raw)
	if err != nil {
		return labelsFail(c, "attach", err)
	}
	return labelsShow(c, "attach", out)
}

// runLabelsDetach is labels detach: a person's command, refused in an agent's session; it needs only a git checkout,
// so a project whose bonsai.yaml is gone can be unregistered.
func runLabelsDetach(c *call) int {
	if e := personOnly("bonsai labels detach", "it detaches label definitions from this machine", c.line()); e != nil {
		return c.refuse(e)
	}
	switch {
	case len(c.rest) == 0:
		return c.refuse(&engine.Error{Code: "missing-value", Exit: exitInput, What: "labels detach needs the namespace to detach",
			Next: "run: bonsai labels detach <namespace> (bonsai status --json lists the namespaces attached)"})
	case len(c.rest) > 1:
		return c.refuse(c.flagError("labels detach takes one namespace, not %d", len(c.rest)))
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.refuse(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project"})
	}
	co, err := workspace.Find(dir)
	if err != nil {
		return c.refuse(engine.FindError(err))
	}
	home, he := labelsHome()
	if he != nil {
		return c.refuse(he)
	}
	out, err := workspace.DetachLabels(home, co.Main, c.rest[0])
	if err != nil {
		return labelsFail(c, "detach", err)
	}
	return labelsShow(c, "detach", out)
}
