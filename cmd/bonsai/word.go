package main

// The command words' one table (spec §3: "bonsai --help and bonsai <word> --help list every flag, exit code and an
// example, so an agent needs nothing else"; plan-5 5.1.4b): each word is a Word, its flags and exit codes in one
// table from which its --help is written and its command line is read, so help and behaviour cannot drift.
//
// One word, one file: init.go, update.go, check.go, status.go and hook.go each hold their word's Word and register it
// (register, in the file's init). To add a word, add its file with its Word and its run function: nothing else
// learns of it, since bonsai --help and the dispatch read the registry. To add a flag, add a Flag to the word's
// table and read it in the word's run function (call.has, call.value, call.all); its help line comes from the table.
// A flag the spec names that is not built yet stays in the table with Later set: its help says so, and the word
// refuses it as not-built until the later step clears Later and reads it. An exit code the word returns is in its
// Exits: the tests fail on a code that is not, and on a refusal whose word is not in format.ErrorWords.
//
// Every refusal goes through call.refuse: with --json, the word's own document with its error object filled (the
// changes output for init and update, check's for check, status's for status), written by internal/format's
// writers; else the sentence and its next step on stderr. A refusal before any word (none given, or one this Bonsai
// does not have) prints, with --json, the error object alone (bonsai.error). hook takes no --json: it speaks Claude
// Code's hook format.

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
)

// Word is one command word: its help, its flags and exit codes, and how it runs.
type Word struct {
	Name    string // the word as typed: "init"; a sub-word gives its parent's too: "hook guard"
	Order   int    // its row in spec §4's table of commands, the order bonsai --help lists the words in
	Title   string // the help's first line, after "bonsai <name>: "
	Summary string // its line in bonsai --help
	Args    string // what follows the word in its usage line: "[flags]", "<name>"
	About   string // the help's text before the flags, as printed (lines end in \n)
	Flags   []Flag // every flag it takes but --help, which every word takes
	Notes   string // the help's text after the flags, as printed; "" for none
	Exits   []Exit // every exit code it returns
	// Examples are full command lines, each printed as "Example: <line>".
	Examples []string
	// Subs are its sub-words (hook's guard): the word's first argument names one, and the rest is the sub-word's
	// command line.
	Subs []*Word
	// Later names the step that builds a sub-word not built yet ("step 5.3"): it is listed, and refused as not-built.
	Later string
	// LaterNext is the next step of the refusal of a sub-word not built yet ("" for the generic one).
	LaterNext string
	// FlagNext is the next step of a refusal of its command line ("" for the generic one: its --help).
	FlagNext string
	// Refused gives the word's --json document for a refusal: nil for a word that takes no --json (hook).
	Refused func(c *call, e *engine.Error) encoder
	// Run runs the word once its command line is read; it returns the exit code.
	Run func(c *call) int
}

// Flag is one flag of a word's table.
type Flag struct {
	Name  string // "--name"
	Value string // the value's placeholder in the help ("N"); "" for a switch
	Many  bool   // it may be given more than once (each value kept); else a second one is refused
	Help  string // what it does, as printed (a \n starts a line under the first)
	// Need names its value in the refusal of a flag given without one ("a format's name"); "" for "a value".
	Need string
	// NeedNext is the next step of that refusal; nil for the word's own.
	NeedNext func() string
	// More is printed under its help (check --schema's list of formats); nil for none.
	More func() string
	// Later names the step that builds it ("step 5.1.6") while it is not built: the help lists it apart, and the word
	// refuses it as not-built.
	Later string
}

// Exit is one exit code a word returns, and what it means for that word.
type Exit struct {
	Code  int
	Means string // as printed (a \n starts a line under the first)
}

// encoder is a document a word prints with --json: its format's writer holds it to its schema first.
type encoder interface {
	Encode() ([]byte, error)
}

// words is the registry: every word, by name.
var words = map[string]*Word{}

// register adds a word to the registry; each word's file calls it from its init.
func register(w *Word) {
	if _, dup := words[w.Name]; dup {
		panic("bonsai: the word " + w.Name + " is registered twice")
	}
	words[w.Name] = w
}

// wordList is every word in spec §4's order.
func wordList() []*Word {
	out := make([]*Word, 0, len(words))
	for _, w := range words {
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out
}

// takesJSON reports whether the word prints a --json document.
func (w *Word) takesJSON() bool { return w.Refused != nil }

func (w *Word) flag(name string) *Flag {
	for i := range w.Flags {
		if w.Flags[i].Name == name {
			return &w.Flags[i]
		}
	}
	return nil
}

func (w *Word) exits(code int) bool {
	for _, e := range w.Exits {
		if e.Code == code {
			return true
		}
	}
	return false
}

// Help is the word's --help, from its table: its title and text, its usage, its sub-words, every flag with its help,
// the flags not built yet, every exit code and an example.
func (w *Word) Help() string {
	var b strings.Builder
	fmt.Fprintf(&b, "bonsai %s: %s\n", w.Name, w.Title)
	b.WriteString(w.About)
	fmt.Fprintf(&b, "Usage: %s\n", strings.TrimSpace("bonsai "+w.Name+" "+w.Args))
	if len(w.Subs) > 0 {
		b.WriteString("This build answers:\n")
		var later []string
		for _, s := range w.Subs {
			if s.Later != "" {
				later = append(later, strings.TrimPrefix(s.Name, w.Name+" ")+" ("+s.Later+")")
				continue
			}
			fmt.Fprintf(&b, "  %-20s %s (bonsai %s --help)\n", "bonsai "+s.Name, s.Summary, s.Name)
		}
		if len(later) > 0 {
			fmt.Fprintf(&b, "Not built yet: %s.\n", strings.Join(later, ", "))
		}
	}
	b.WriteString("Flags:\n")
	flags := append([]Flag{}, w.Flags...)
	flags = append(flags, Flag{Name: "--help", Help: "print this help"})
	width := 0
	for _, f := range flags {
		if n := len(flagShown(f)); n > width {
			width = n
		}
	}
	var later []string
	for _, f := range flags {
		if f.Later != "" {
			later = append(later, flagShown(f)+" ("+f.Later+")")
			continue
		}
		lines := strings.Split(f.Help, "\n")
		fmt.Fprintf(&b, "  %-*s  %s\n", width, flagShown(f), lines[0])
		for _, l := range lines[1:] {
			fmt.Fprintf(&b, "  %-*s  %s\n", width, "", l)
		}
		if f.More != nil {
			b.WriteString(wrapWords(f.More(), strings.Repeat(" ", width+4), 100))
		}
	}
	b.WriteString(w.Notes)
	if len(later) > 0 {
		fmt.Fprintf(&b, "Not built yet in this build: %s.\n", strings.Join(later, ", "))
	}
	b.WriteString("Exit codes:\n")
	for _, e := range w.Exits {
		lines := strings.Split(e.Means, "\n")
		fmt.Fprintf(&b, "  %d  %s\n", e.Code, lines[0])
		for _, l := range lines[1:] {
			fmt.Fprintf(&b, "     %s\n", l)
		}
	}
	for _, ex := range w.Examples {
		fmt.Fprintf(&b, "Example: %s\n", ex)
	}
	return b.String()
}

func flagShown(f Flag) string {
	if f.Value == "" {
		return f.Name
	}
	return f.Name + " " + f.Value
}

// wrapWords wraps space-separated words at width, each line starting with indent.
func wrapWords(text, indent string, width int) string {
	var b strings.Builder
	line := ""
	for _, word := range strings.Fields(text) {
		if line != "" && len(line)+1+len(word) > width {
			b.WriteString(line + "\n")
			line = ""
		}
		if line == "" {
			line = indent + word
		} else {
			line += " " + word
		}
	}
	if line != "" {
		b.WriteString(line + "\n")
	}
	return b.String()
}

// call is one run of a word: its command line as read against its table, where it prints, and whether --json was
// asked for.
type call struct {
	word           *Word
	given          []given  // each flag given, in order
	rest           []string // the arguments that are not flags
	json           bool     // --json given (anywhere on the line) to a word that takes it
	stdout, stderr io.Writer
}

// given is one flag as given on the command line.
type given struct {
	name, value string
}

func (c *call) has(name string) bool {
	for _, g := range c.given {
		if g.name == name {
			return true
		}
	}
	return false
}

// value is the flag's value, "" when it was not given.
func (c *call) value(name string) string {
	v := ""
	for _, g := range c.given {
		if g.name == name {
			v = g.value
		}
	}
	return v
}

// all is every value of a flag given more than once.
func (c *call) all(name string) []string {
	var out []string
	for _, g := range c.given {
		if g.name == name {
			out = append(out, g.value)
		}
	}
	return out
}

// runWord reads a word's command line against its table and runs it: its help, a sub-word, a refusal of a flag the
// table does not have, or the word's own run. It returns the exit code.
func runWord(w *Word, args []string, stdout, stderr io.Writer) int {
	c := &call{word: w, stdout: stdout, stderr: stderr}
	if w.takesJSON() {
		for _, a := range args {
			if a == "--json" {
				c.json = true
			}
		}
	}
	if len(w.Subs) > 0 {
		return c.runSub(args)
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--help" || a == "-h" {
			return c.print(w.Help())
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			c.rest = append(c.rest, a)
			continue
		}
		name, value, hasValue := a, "", false
		if j := strings.IndexByte(a, '='); j > 0 && strings.HasPrefix(a, "--") {
			name, value, hasValue = a[:j], a[j+1:], true
		}
		f := w.flag(name)
		switch {
		case f == nil:
			return c.refuse(c.flagError("%s takes no %+q", w.Name, a))
		case f.Later != "":
			drop := 1
			if f.Value != "" && !hasValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				drop = 2
			}
			return c.refuse(&engine.Error{Code: "not-built", Exit: exitInput,
				What: fmt.Sprintf("%s %s is not built yet (it comes with %s of the rebuild)", w.Name, f.Name, f.Later),
				Next: "run `" + c.without(args, i, drop) + "` without " + f.Name})
		case f.Value == "" && hasValue:
			return c.refuse(c.flagError("%s takes no value for %+q", w.Name, name))
		case f.Value != "" && !hasValue:
			if i+1 >= len(args) || (strings.HasPrefix(args[i+1], "-") && args[i+1] != "-") {
				need := f.Need
				if need == "" {
					need = "a value (" + f.Value + ")"
				}
				e := c.flagError("%s %s needs %s", w.Name, f.Name, need)
				if f.NeedNext != nil {
					e.Next = f.NeedNext()
				}
				return c.refuse(e)
			}
			i++
			value = args[i]
		}
		if f.Value != "" && !f.Many && c.has(f.Name) {
			return c.refuse(c.flagError("%s takes %s once", w.Name, f.Name))
		}
		c.given = append(c.given, given{f.Name, value})
	}
	return w.Run(c)
}

// runSub runs a word's sub-word (hook guard): its first argument names it.
func (c *call) runSub(args []string) int {
	w := c.word
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		return c.print(w.Help())
	}
	if len(args) == 0 {
		return c.refuse(&engine.Error{Code: "missing-value", Exit: exitInput, What: "bonsai " + w.Name + " needs a " + w.Name + " " + strings.Trim(w.Args, "<>"),
			Next: "run `bonsai " + w.Name + " --help`"})
	}
	for _, s := range w.Subs {
		if s.Name != w.Name+" "+args[0] {
			continue
		}
		if s.Later != "" {
			next := s.LaterNext
			if next == "" {
				next = "run `bonsai " + w.Name + " --help`"
			}
			return c.refuse(&engine.Error{Code: "not-built", Exit: exitInput,
				What: "bonsai " + s.Name + " is not built yet (it comes with a later part of the rebuild: " + s.Later + ")", Next: next})
		}
		return runWord(s, args[1:], c.stdout, c.stderr)
	}
	return c.refuse(&engine.Error{Code: "bad-value", Exit: exitInput, What: fmt.Sprintf("%+q is not a %s", args[0], w.Name),
		Next: "run `bonsai " + w.Name + " --help`"})
}

// flagError is a refusal of the command line, naming the word's help or its own next step.
func (c *call) flagError(format string, args ...any) *engine.Error {
	next := c.word.FlagNext
	if next == "" {
		next = "run `bonsai " + c.word.Name + " --help` to see its flags"
	}
	return &engine.Error{Code: "bad-flag", Exit: exitInput, What: fmt.Sprintf(format, args...), Next: next}
}

// without is the command line again without the n arguments from i.
func (c *call) without(args []string, i, n int) string {
	line := "bonsai " + c.word.Name
	for j, a := range args {
		if j < i || j >= i+n {
			line += " " + engine.ShellArg(a)
		}
	}
	return line
}

// print prints the word's answer on stdout; one that cannot be written exits 3, or 2 for a word that returns no 3
// (hook).
func (c *call) print(s string) int {
	if write(c.stdout, s) != exitOK {
		if c.word.exits(exitRuntime) {
			return exitRuntime
		}
		return exitInput
	}
	return exitOK
}

// refuse prints a refusal and returns its exit code: with --json the word's document with the error object filled,
// else the sentence and its next step on stderr ("bonsai: " for the command line, "bonsai <word>: " once the word
// runs).
func (c *call) refuse(e *engine.Error) int {
	noted(c.word, e)
	if c.json {
		return c.printDoc(c.word.Refused(c, e), e.Exit)
	}
	_, _ = fmt.Fprintf(c.stderr, "bonsai: %s.\nnext: %s\n", strings.TrimSuffix(e.What, "."), e.Next)
	return e.Exit
}

// fail is refuse for a failure once the word runs: the human line names the word.
func (c *call) fail(e *engine.Error) int {
	noted(c.word, e)
	if c.json {
		return c.printDoc(c.word.Refused(c, e), e.Exit)
	}
	_, _ = fmt.Fprintf(c.stderr, "bonsai %s: %s.\nnext: %s\n", c.word.Name, strings.TrimSuffix(e.What, "."), e.Next)
	return e.Exit
}

// printDoc prints a --json document through its format's writer and returns exit. A document that does not fit its
// schema is a bug: the word's refusal document is printed in its place, its error unexpected and naming what did not
// fit, and the command exits 3.
func (c *call) printDoc(doc encoder, exit int) int {
	b, err := doc.Encode()
	if err != nil {
		e := &engine.Error{Code: "unexpected", Exit: exitRuntime, What: "Bonsai's own --json document did not fit its schema: " + err.Error(),
			Next: "run the command again without --json, and report this to Bonsai's maintainers"}
		noted(c.word, e)
		if b, err = c.word.Refused(c, e).Encode(); err != nil {
			_, _ = fmt.Fprintf(c.stderr, "bonsai %s: %v\n", c.word.Name, err)
			return exitRuntime
		}
		exit = exitRuntime
	}
	if write(c.stdout, string(b)) != exitOK {
		return exitRuntime
	}
	return exit
}

// noteError, when set (by the tests), sees every refusal before it is printed: its word, exit code and error.
var noteError func(w *Word, e *engine.Error)

// noteExit, when set (by the tests), sees every word's exit code.
var noteExit func(w *Word, code int)

func noted(w *Word, e *engine.Error) {
	if noteError != nil {
		noteError(w, e)
	}
}

// errorOnly is the error object alone, the --json of a refusal before any word.
type errorOnly struct{ e *engine.Error }

func (o errorOnly) Encode() ([]byte, error) { return o.e.Object().Encode() }
