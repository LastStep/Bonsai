package main

// bonsai --help --json (bonsai.help/1, plan-5 5.1.10): the whole tool as one document for an agent, read from the same
// tables as the human --help: the word registry (Word, Flag and Exit in word.go), format.ExitCodes and
// format.ErrorWords. It lists every word and every flag the registry has, which TestHelpJSONHasEveryWord holds, and
// it is held to formats/schemas/help.schema.json by format.Help's writer, so a document that does not fit it is never
// printed. Only the bare `bonsai --help --json` prints it: a word's own --help stays text whatever flags come with it
// (the word tables' tests hold that), and the document holds every word, hook's included (hook takes no --json of its
// own, it speaks Claude Code's hook format; the document describes the tool, it is not hook's output).

import (
	"io"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
)

// helpEntries is every word in the order bonsai --help lists them, each followed by its sub-words.
func helpEntries() []*Word {
	var out []*Word
	for _, w := range wordList() {
		out = append(out, w)
		out = append(out, w.Subs...)
	}
	return out
}

// helpRequest reports whether args are exactly bonsai --help --json (or -h, in either order). A word's own --help,
// with or without --json, stays the human text: every word takes --json as a flag beside --help and the tests hold
// that the help prints whatever other flags of its table come with it.
func helpRequest(args []string) bool {
	if len(args) != 2 {
		return false
	}
	return (args[0] == "--help" || args[0] == "-h") && args[1] == "--json" ||
		args[0] == "--json" && (args[1] == "--help" || args[1] == "-h")
}

// helpDoc is the document: every word, each followed by its sub-words.
func helpDoc() *format.Help {
	h := &format.Help{Bonsai: version, About: helpAbout()}
	for _, e := range format.ExitCodes {
		h.ExitCodes = append(h.ExitCodes, format.HelpExitCode{Code: int64(e.Code), Short: e.Short, Means: e.Means, Applies: e.Applies})
	}
	for _, w := range helpEntries() {
		h.Words = append(h.Words, helpWord(w))
	}
	for _, e := range format.ErrorWords {
		h.ErrorWords = append(h.ErrorWords, format.HelpErrorWord{Code: e.Word, Means: e.Means, Who: e.Who})
	}
	return h
}

func helpWord(w *Word) format.HelpWord {
	hw := format.HelpWord{Name: w.Name, Summary: w.Summary, Title: w.Title, About: w.About, Notes: w.Notes,
		Usage: strings.TrimSpace("bonsai " + w.Name + " " + w.Args), TakesJSON: w.takesJSON(),
		Examples: append([]string{}, w.Examples...)}
	if w.Later != "" {
		later := w.Later
		hw.Later = &later
	}
	for _, f := range w.Flags {
		hf := format.HelpFlag{Name: f.Name, Many: f.Many, Help: f.Help}
		if f.Value != "" {
			v := f.Value
			hf.Value = &v
		}
		if f.More != nil {
			hf.Help += "\n" + strings.Join(strings.Fields(f.More()), " ")
		}
		if f.Later != "" {
			later := f.Later
			hf.Later = &later
		}
		hw.Flags = append(hw.Flags, hf)
	}
	for _, e := range w.Exits {
		hw.Exits = append(hw.Exits, format.HelpExit{Code: int64(e.Code), Means: e.Means})
	}
	for _, s := range w.Subs {
		hw.Subs = append(hw.Subs, s.Name)
	}
	return hw
}

// runHelpJSON prints the document and returns the exit code.
func runHelpJSON(stdout, stderr io.Writer) int {
	c := &call{word: bonsaiWord, json: true, stdout: stdout, stderr: stderr}
	return c.printDoc(helpDoc(), exitOK)
}
