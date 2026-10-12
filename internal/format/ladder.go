package format

// The ladder result (bonsai.ladder/1, contract §11): the proof that a task's checks passed, at a commit. Bonsai's
// ladder runner writes it (step 5.4, through Ladder.Encode); the stop gate and the studio read it.
//
// The ladder's words live in Go tables beside its type, their one home (set 7, plan-5 5.4.0): RungStatuses (a rung's
// status), GitFindings (what a ver-git rung's captures name), TestForms (a rung's tests) and the test-file convention
// of each form (TestFileConventions). Each is registered as a List in format.go's entries for ladder and workspace, so
// bonsai check --schema bonsai.ladder and bonsai.workspace print them and the reference page of lists reads them; the
// schemas' descriptions name those commands and copy no word.

import (
	"encoding/json"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Ladder is one ladder result.
type Ladder struct {
	Task      *string       `json:"task"`      // the task's id, or null
	Workspace string        `json:"workspace"` // the workspace id
	Mode      string        `json:"mode"`      // the closed list: local, ci
	Started   string        `json:"started"`   // UTC, to the millisecond
	Finished  string        `json:"finished"`
	Git       LadderGit     `json:"git"`
	Requested []int64       `json:"requested"` // the rungs asked for
	Green     bool          `json:"green"`
	Rungs     []LadderRung  `json:"rungs"`
	Skipped   []LadderSkip  `json:"skipped"`
	Leftovers any           `json:"leftovers"` // an object, a text or null: typed open
	Proof     *string       `json:"proof"`
	Extra     schema.Object `json:"-"`
}

// LadderGit is the commit the result proves.
type LadderGit struct {
	SHA    string        `json:"sha"`
	Branch string        `json:"branch"`
	Dirty  bool          `json:"dirty"`
	Extra  schema.Object `json:"-"`
}

// LadderRung is one rung's outcome.
type LadderRung struct {
	Rung       int64          `json:"rung"`
	Name       string         `json:"name"`
	Kind       string         `json:"kind"` // the closed list: guard, command, ver-git
	Required   bool           `json:"required"`
	Command    *string        `json:"command"`
	Means      *string        `json:"means"`
	Status     string         `json:"status"` // typed open; its known words are RungStatuses
	Reason     *string        `json:"reason"`
	DurationMS *int64         `json:"duration_ms"`
	Tests      schema.Object  `json:"tests"`    // typed open, or null
	Captures   schema.Object  `json:"captures"` // name to a number, a text or null; or null
	Ratchet    *LadderRatchet `json:"ratchet"`
	Extra      schema.Object  `json:"-"`
}

// LadderRatchet is a rung's count against its floor.
type LadderRatchet struct {
	Name     string          `json:"name"`
	Was      json.Number     `json:"was"`
	Now      json.Number     `json:"now"`
	OK       bool            `json:"ok"`
	NewTests *LadderNewTests `json:"new_tests"`
	Extra    schema.Object   `json:"-"`
}

// LadderNewTests is the new tests run on the base.
type LadderNewTests struct {
	Base         string        `json:"base"`           // the base commit, "" when none could be found (set 7)
	Count        int64         `json:"count"`          // the new tests the rung's output named
	FailedOnBase int64         `json:"failed_on_base"` // those the base run reported by name as failed
	PassedOnBase int64         `json:"passed_on_base"` // those the base run reported passed: they prove nothing new (set 7)
	NotRun       int64         `json:"not_run"`        // those the base run did not report: the check was not run (set 7)
	Extra        schema.Object `json:"-"`
}

// LadderSkip is a rung skipped by a mark.
type LadderSkip struct {
	Rung     int64         `json:"rung"`
	Name     string        `json:"name"`
	Mark     *string       `json:"mark"`
	Reason   string        `json:"reason"`
	Accepted bool          `json:"accepted"`
	Extra    schema.Object `json:"-"`
}

// ReadLadder reads a ladder result's bytes.
func ReadLadder(raw []byte) (*Ladder, error) {
	l := &Ladder{}
	if err := MustLookup("ladder").readJSONInto(raw, l); err != nil {
		return nil, err
	}
	return l, nil
}

// Encode writes the result, held to bonsai.ladder/1.
func (l *Ladder) Encode() ([]byte, error) { return MustLookup("ladder").Encode(l) }

// RungStatuses are the known words of a rung's status in a result: an open list (formats/README.md, "Typed open"),
// whose one home is this table. pending is a word today's results hold and Bonsai never writes: no rung field marks a
// rung as not built.
var RungStatuses = []Word{
	{Word: "green", Means: "the rung ran and passed"},
	{Word: "red", Means: "the rung ran and failed: its exit, its tests, its ratchet's count or its timeout; the reason says which"},
	{Word: "skipped", Means: "the rung did not run: an earlier required rung was red, or the run did not ask for it"},
	{Word: "error", Means: "the rung could not run or its outcome could not be read (the command did not start, a test output did not parse, a ver-git rung could not read git)"},
	{Word: "pending", Means: "a rung not built yet, which today's results hold; Bonsai's runner never writes it"},
}

// The tests forms: a rung's tests in bonsai.yaml names how its output is read.
const (
	TestsGoJSON = "go-test-json"
	TestsTAP    = "tap"
	TestsJUnit  = "junit"
)

// TestForms are the known words of a rung's tests in bonsai.yaml: an open list (an unknown word means the tests are
// not read, and bonsai check warns, naming these), whose one home is this table.
var TestForms = []Word{
	{Word: TestsGoJSON, Means: "the rung's standard output is go test -json: one JSON event per line, tests named by package and test, subtests by their path"},
	{Word: TestsTAP, Means: "the rung's standard output is TAP, as node's test runner prints it (node --test --test-reporter=tap): ok and not ok lines, subtests indented"},
	{Word: TestsJUnit, Means: "the rung's standard output is a JUnit XML report: testcase elements with their classname, name and file, failures and skips inside"},
}

// TestFileConvention says which files are a project's test files for one tests form, which new tests must fail on the
// base (5.4.5) and git integrity's changed-test-file finding (5.4.6) both read. Patterns are project-relative globs
// with forward slashes, in the dialect of bonsai.yaml's protected list (** is a whole segment); a path matching any
// is a test file. A form with no patterns names its files itself.
type TestFileConvention struct {
	Form     string
	Patterns []string
	Means    string // one line
}

// TestFileConventions are the test-file conventions, one per tests form, in TestForms' order.
var TestFileConventions = []TestFileConvention{
	{Form: TestsGoJSON, Patterns: []string{"**/*_test.go", "**/testdata/**"},
		Means: "the *_test.go files, and the files under a package's testdata folder"},
	{Form: TestsTAP, Patterns: tapPatterns(),
		Means: "node's test runner's default patterns, for .js, .mjs, .cjs and .ts: *.test.*, *-test.*, *_test.*, test-*.*, test.*, and files under a test folder"},
	{Form: TestsJUnit, Patterns: nil,
		Means: "the files the report's testcase elements name (their file attribute); with none named, the test files cannot be told, and the check is not run"},
}

// tapPatterns builds the tap convention's globs: node's default patterns for each extension it runs.
func tapPatterns() []string {
	var out []string
	for _, ext := range []string{"js", "mjs", "cjs", "ts"} {
		out = append(out, "**/*.test."+ext, "**/*-test."+ext, "**/*_test."+ext, "**/test-*."+ext, "**/test."+ext, "**/test/**/*."+ext)
	}
	return out
}

// TestFileWords is TestFileConventions as a list of words (a form and its convention), for bonsai check --schema and
// the reference page; it is derived, never edited.
var TestFileWords = testFileWords()

func testFileWords() []Word {
	out := make([]Word, len(TestFileConventions))
	for i, c := range TestFileConventions {
		out[i] = Word{Word: c.Form, Means: "test files: " + c.Means}
	}
	return out
}

// TestFileConventionOf finds a tests form's convention.
func TestFileConventionOf(form string) (TestFileConvention, bool) {
	for _, c := range TestFileConventions {
		if c.Form == form {
			return c, true
		}
	}
	return TestFileConvention{}, false
}

// GitFindings are the names a ver-git rung gives what it found, as keys of the rung's captures in a result: an open
// list, whose one home is this table (5.4.6 writes them, in this order; the rung informs and never turns red for
// one). The first three are what the rung records, so the task's next climb compares: base (a commit, or null when
// there is none), stash-count (a number) and exclude-hash (sha256:<hex>). Every other word is a finding, its value
// the count found, or null when the rung could not judge it (its reason says why). A finding whose line says it
// hides a change makes the result's git.dirty true: git status would not show that change.
var GitFindings = []Word{
	{Word: "base", Means: "recorded: the task's base commit, the merge base of HEAD with the base branch (on the base branch itself, the base an earlier climb of the task recorded); null when there is none"},
	{Word: "stash-count", Means: "recorded: the number of entries the stash holds, which the task's next climb compares (stash-added)"},
	{Word: "exclude-hash", Means: "recorded: sha256:<hex> of the common git folder's info/exclude with line ends read as LF (an absent file reads as empty), which the task's next climb compares (exclude-edited)"},
	{Word: "test-file-changed", Means: "a test file (by the tests form, TestFileConventions) differs between the task's base and the working tree: changed, added (untracked ones too) or deleted"},
	{Word: "assume-unchanged", Means: "the index marks a file assume-unchanged, so git stops looking for edits to it: it hides a change, and the run counts as dirty"},
	{Word: "skip-worktree", Means: "the index marks a file skip-worktree, so git keeps the index's copy whatever the working tree holds: it hides a change, and the run counts as dirty"},
	{Word: "exclude-edited", Means: "the common git folder's info/exclude (line ends read as LF) differs from the hash the task's last climb recorded"},
	{Word: "stash-added", Means: "the stash holds more entries than the task's last climb recorded"},
	{Word: "base-rewritten", Means: "the base the task's last climb recorded is no longer in HEAD's history (or no longer in a clone that is not shallow)"},
	{Word: "exclude-lines", Means: "the common git folder's info/exclude holds lines other than comments: patterns that ignore files for this clone alone, which no diff shows"},
	{Word: "exclude-hidden", Means: "a file or folder present that git ignores only by info/exclude or core.excludesFile (no .gitignore ignores it), so git status does not show it: it hides a change, and the run counts as dirty, but for one inside a .claude/ folder (Claude Code's own local files, which Claude Code ignores that way itself), named only"},
	{Word: "untracked-gitignore", Means: "a file or folder present that git ignores by a .gitignore git does not track (one that ignores itself too), so git status does not show it: it hides a change, and the run counts as dirty"},
	{Word: "filter-local", Means: "the repository's own git config defines a filter driver (filter.<name>.clean, smudge or process), which can change what git compares, so git status may not show a change"},
}
