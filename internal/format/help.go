package format

// bonsai --help --json (bonsai.help/1, plan-5 5.1.10): the whole tool for a program, so an agent managing a project
// learns Bonsai without reading a page: every command word with its flags, exit codes and examples, the exit codes
// every word shares, and every word of the error object's code with its meaning and usual who. Printed, never stored.
// Written by cmd/bonsai from its word registry (Word, Flag, Exit) and from ExitCodes and ErrorWords, so it reads the
// tables that make the human --help and the behaviour and cannot drift from them.

import "github.com/LastStep/Bonsai/internal/schema"

// Help is bonsai --help --json's document.
type Help struct {
	Bonsai     string          `json:"bonsai"`      // the version of this build
	About      string          `json:"about"`       // what to know before the words
	ExitCodes  []HelpExitCode  `json:"exit_codes"`  // the codes every word shares (ExitCodes)
	Words      []HelpWord      `json:"words"`       // every command word, in the order bonsai --help lists them
	ErrorWords []HelpErrorWord `json:"error_words"` // every word of the error object's code (ErrorWords)
	Extra      schema.Object   `json:"-"`
}

// HelpExitCode is one exit code every word shares (or hook's).
type HelpExitCode struct {
	Code    int64         `json:"code"`
	Short   string        `json:"short"`
	Means   string        `json:"means"`
	Applies string        `json:"applies"` // every word but hook, or hook
	Extra   schema.Object `json:"-"`
}

// HelpWord is one command word.
type HelpWord struct {
	Name      string        `json:"name"`       // as typed: "check", "hook guard"
	Summary   string        `json:"summary"`    // its line in bonsai --help
	Title     string        `json:"title"`      // the first line of its --help, after "bonsai <name>: "
	Usage     string        `json:"usage"`      // its usage line
	About     string        `json:"about"`      // its --help's text before the flags
	TakesJSON bool          `json:"takes_json"` // it prints a --json document (every word but hook)
	Later     *string       `json:"later"`      // the step that builds it, when this build does not have it; else null
	Flags     []HelpFlag    `json:"flags"`      // every flag it takes but --help, which every word takes
	Notes     string        `json:"notes"`      // its --help's text after the flags
	Exits     []HelpExit    `json:"exits"`      // every exit code it returns
	Examples  []string      `json:"examples"`   // full command lines
	Subs      []string      `json:"subs"`       // its sub-words' names ("hook guard"), each also a word in this list
	Extra     schema.Object `json:"-"`
}

// HelpFlag is one flag of a word.
type HelpFlag struct {
	Name  string        `json:"name"`  // "--json"
	Value *string       `json:"value"` // its value's placeholder ("N"), or null for a switch
	Many  bool          `json:"many"`  // it may be given more than once
	Help  string        `json:"help"`  // what it does
	Later *string       `json:"later"` // the step that builds it while it is not built (the word refuses it); else null
	Extra schema.Object `json:"-"`
}

// HelpExit is one exit code a word returns, and what it means for that word.
type HelpExit struct {
	Code  int64         `json:"code"`
	Means string        `json:"means"`
	Extra schema.Object `json:"-"`
}

// HelpErrorWord is one known word of the error object's code.
type HelpErrorWord struct {
	Code  string        `json:"code"`  // the word
	Means string        `json:"means"` // what it means
	Who   string        `json:"who"`   // who usually takes the next step: agent or person
	Extra schema.Object `json:"-"`
}

// ReadHelp reads bonsai --help --json's output.
func ReadHelp(raw []byte) (*Help, error) {
	h := &Help{}
	if err := MustLookup("help").readJSONInto(raw, h); err != nil {
		return nil, err
	}
	return h, nil
}

// Encode writes the document, held to bonsai.help/1.
func (h *Help) Encode() ([]byte, error) { return MustLookup("help").Encode(h) }
