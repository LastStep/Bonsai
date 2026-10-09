package format

// bonsai check --json (bonsai.check/1, spec §6): what check found in a project. Printed, never stored. Written by
// bonsai check (step 5.1.4b moves it onto Check.Encode); read by the studio, agents and CI.

import "github.com/LastStep/Bonsai/internal/schema"

// Check is check --json's document.
type Check struct {
	Findings []Finding     `json:"findings"` // each one makes the exit code 1
	Warnings []Finding     `json:"warnings"` // never the exit code
	Error    *ErrorObject  `json:"error"`    // null unless check refused or failed
	Extra    schema.Object `json:"-"`
}

// Finding is one finding or warning.
type Finding struct {
	Code    string        `json:"code"`    // a fixed word: an open list whose known words are CheckWords
	File    *string       `json:"file"`    // the project-relative file it is about, or null
	Message string        `json:"message"` // one sentence
	Next    Next          `json:"next"`
	Extra   schema.Object `json:"-"`
}

// CheckWords are the known words of a finding's and a warning's code: an open list whose one home is this table
// (formats/README.md); bonsai check --schema bonsai.check prints it, and step 5.1.10's reference page lists it. One
// word per thing check looks for (spec section 6, "bonsai check findings" and "Warnings"), each either a finding (it
// makes check exit 1, and status --json lists it under problems) or a warning (printed, never the exit code, never a
// problem). Who is who usually takes the next step; a finding names the other when its own step is the other's.
// cmd/bonsai's TestCheckTable walks this table: a word with no test case fails, and so does a next step naming a
// command Bonsai does not have or a flag its word does not take. A later piece adds a word by adding an entry here
// with its case (step 5.1.8 the stale tables, 5.2 the secret-shaped string, 5.6 the stranded machine folder: engine's
// checkLater names them); a word is added, never renamed or taken out.
var CheckWords = []Word{
	// The project's own files against the lock and bonsai.yaml (offline: the lock alone).
	{Word: "config", Kind: "finding", Who: "person", Means: "bonsai.yaml is not one Bonsai reads"},
	{Word: "lock", Kind: "finding", Who: "person", Means: ".bonsai/lock.json is missing or not one Bonsai reads"},
	{Word: "packs", Kind: "finding", Who: "person", Means: "bonsai.yaml and the lock name other packs, sources, refs or folders"},
	{Word: "changed", Kind: "finding", Who: "person", Means: "a file the lock lists was edited: a pack file, Bonsai's block in CLAUDE.md, or Bonsai's lines in .claude/settings.json"},
	{Word: "missing", Kind: "finding", Who: "person", Means: "a file the lock lists is gone"},
	{Word: "gitignore", Kind: "finding", Who: "person", Means: ".bonsai/.gitignore is missing or changed, so .bonsai/local/ could be committed"},
	{Word: "local", Kind: "finding", Who: "agent", Means: "git tracks or has staged a file from .bonsai/local/, which is never committed"},
	// Bonsai's document kinds (contract sections 2.3, 5, 6, 7).
	{Word: "format0", Kind: "finding", Who: "agent", Means: "a format-0 file the lock's format0 list fixes changed or is gone (contract section 2.3)"},
	{Word: "format0-new", Kind: "finding", Who: "agent", Means: "a task, run report or STATE file with no format: line that is not on the lock's format0 list: it takes a format: line first (contract section 2.3)"},
	{Word: "document", Kind: "finding", Who: "agent", Means: "a file of one of Bonsai's kinds (task, run, state, memory, the two tables) does not read under its format, 0 or 1"},
	{Word: "label", Kind: "finding", Who: "agent", Means: "a label's value is not of the kind its definition says, or it is on a kind of document its definition does not name (contract section 5.2)"},
	{Word: "label-twice", Kind: "finding", Who: "person", Means: "two sources (packs, or a file attached on this machine) define one label name (contract section 5.1)"},
	{Word: "approve-first", Kind: "finding", Who: "person", Means: "a task in a lane with approve_first reached running, verify or done without reading approved since it last read todo or plan, as git history shows (contract section 6)"},
	{Word: "absolute-path", Kind: "finding", Who: "agent", Means: "a committed Bonsai file (bonsai.yaml, a document of one of Bonsai's kinds) holds an absolute path in a field, a pack's source among them unless it is a remote URL (contract section 2.6; spec section 14, check 2)"},
	{Word: "block-size", Kind: "finding", Who: "person", Means: "Bonsai's block in CLAUDE.md is over its fixed 40 lines (spec section 6)"},
	{Word: "memory-index-size", Kind: "finding", Who: "agent", Means: "the memory index is over its fixed 120 lines or 12 KB (spec section 10)"},
	{Word: "memory-note-size", Kind: "finding", Who: "agent", Means: "a memory note is over 4 KB (spec section 10)"},
	{Word: "missing-path", Kind: "finding", Who: "agent", Means: "a project path named in CLAUDE.md, STATE or a memory note does not exist (Bonsai's own block left out: the lock checks it)"},
	// Claude Code's settings in the project (spec sections 5 and 7).
	{Word: "settings-rule", Kind: "finding", Who: "person", Means: "a permission rule in .claude/settings.json or .claude/settings.local.json is not valid on its own"},
	{Word: "hooks-off", Kind: "finding", Who: "person", Means: "disableAllHooks is true in .claude/settings.json or .claude/settings.local.json, which turns off every hook, Bonsai's guard among them"},
	{Word: "plugin-version", Kind: "finding", Who: "person", Means: "a pack plugin's entry in a Bonsai marketplace carries a version, which a plugin pinned by commit never does (spec section 5)"},
	{Word: "plugin", Kind: "finding", Who: "person", Means: "plugin drift: Claude Code turns on another commit of a locked pack's plugin for this checkout (spec section 5)"},
	// This machine (the home and the PATH).
	{Word: "id-changed", Kind: "finding", Who: "person", Means: "bonsai.yaml's workspace id is not the one this machine last recorded for the checkout (contract section 3)"},
	{Word: "bonsai-path", Kind: "finding", Who: "person", Means: "the bonsai on the PATH is not the installed one that install.json records (spec section 3)"},
	// Warnings.
	{Word: "claude-code-old", Kind: "warning", Who: "person", Means: "Claude Code is older than the floor: Bonsai's own or a pack's needs.claude_code, the higher (spec section 7)"},
	{Word: "claude-code-unknown", Kind: "warning", Who: "person", Means: "Claude Code's version could not be read (not on the PATH, or claude --version gave an answer Bonsai does not read)"},
	{Word: "same-id", Kind: "warning", Who: "person", Means: "another checkout on this machine holds the same workspace id (contract section 3)"},
	{Word: "run-reports", Kind: "warning", Who: "person", Means: "run reports are past bonsai.yaml's generated.run rule, listed for a person to delete (spec section 6)"},
	{Word: "approve-first-unchecked", Kind: "warning", Who: "agent", Means: "approve_first was not checked: git history is not available (a shallow clone) or could not be read; never passed"},
	{Word: "plugin-missing", Kind: "warning", Who: "person", Means: "Claude Code reports a locked pack's plugin not installed for this checkout"},
	{Word: "plugin-unchecked", Kind: "warning", Who: "person", Means: "this machine's plugins were not compared with the lock (Claude Code not on the PATH, its answer unread, a local settings file unread)"},
	{Word: "cache", Kind: "warning", Who: "person", Means: "a lock written before formats set 4 names a pack this machine's cache lacks, so Bonsai's lines in .claude/settings.json were not checked"},
	{Word: "local-unchecked", Kind: "warning", Who: "agent", Means: "git ls-files failed, so files from .bonsai/local/ in git's index were not looked for"},
}

// CheckWord finds a word in CheckWords.
func CheckWord(code string) (Word, bool) {
	for _, w := range CheckWords {
		if w.Word == code {
			return w, true
		}
	}
	return Word{}, false
}

// ReadCheck reads check --json's output.
func ReadCheck(raw []byte) (*Check, error) {
	c := &Check{}
	if err := MustLookup("check").readJSONInto(raw, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Encode writes the document, held to bonsai.check/1.
func (c *Check) Encode() ([]byte, error) { return MustLookup("check").Encode(c) }
