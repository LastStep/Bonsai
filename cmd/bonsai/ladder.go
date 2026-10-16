package main

// bonsai ladder (spec §4, §9; contract §11, §13; plan-5 5.4.2): the ladder runner, internal/ladder. Its table of flags
// and exit codes is ladderWord; with --json it prints bonsai.climb/1 (format.Climb), always whole, its error object
// filled when the climb refused or failed.

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/ladder"
)

func init() { register(ladderWord) }

var ladderWord = &Word{
	Name:    "ladder",
	Order:   7,
	Title:   "climb the ladder: prove a task's work done (spec section 9).",
	Summary: "climb the project's rungs and write the result the stop gate reads",
	Args:    "[flags]",
	About: `It runs the project's rungs (bonsai.yaml's ladder, in the main checkout) in number order, cheapest first: the floor
(ladder_floor) plus the task's bonsai.ladder label, each once. Each command rung runs in its own process group (Linux)
or job object (Windows), so everything it starts ends with it; its tests are read from its output (its tests form), a
skip with a mark is red (a local climb runs every marked test), and its ratchet's count is judged against its floor
(bonsai.yaml's ratchets): below it, with no count read, or with a floor that is not a number, the rung is red. A
required rung that is not green skips the rest. The rungs, the floor and the ratchets are bonsai.yaml's as committed
at the main checkout's HEAD; a local climb refuses settings looser than the ones this machine last accepted (a person
records those with bonsai update) and records stricter ones itself. On a local climb that is green and clean, each
count that rose has its new tests run on the base, in a temporary worktree: the result's new_tests counts those
failing there, those passing there (they prove nothing new) and those not run. The result goes to the main checkout's
.bonsai/local/ladder/<task>.json, wherever the climb ran (a worktree's included), with a ladder record in the log
naming its sha256 and this bonsai's own path and hash: that is what the stop gate reads before a named session may
end. A green, clean climb in the main checkout on the base branch at HEAD whose count rose files one Bless ask
(ladder:bless-<task>); every climb on the base branch resolves the Bless asks whose offer has gone; --ci files
nothing. The task comes from --task, or BONSAI_TASK in this project's session; a local climb needs one. One climb at a
time per Bonsai home: a second waits while the first runs inside its budget, then refuses. A red climb's next step
names the first red rung's fix (each failing test carries the command that runs it alone), then the climb again.
`,
	Flags: []Flag{
		{Name: "--task", Value: "<id>", Need: "a task's id (T- and 4 to 6 digits)",
			Help: "the task to climb for (contract section 13): its rungs are the floor plus its bonsai.ladder,\nits result " +
				".bonsai/local/ladder/<id>.json; BONSAI_TASK names it too, in the session's project"},
		{Name: "--root", Value: "<folder>", Need: "a checkout's folder",
			Help: "climb the checkout holding <folder> instead of the one holding the working folder"},
		{Name: "--ci", Help: "a CI climb: mode ci, written to .bonsai/local/ladder/ci.json, never a task's proof; it may name no\n" +
			"task (then every rung), skips may pass only for ci_marked_tests with a mark, no leftovers line"},
		{Name: "--json", Help: "print the bonsai.climb/1 document (for programs) instead of text: always whole, the result\ninside it, " +
			"next naming the fix of a red climb"},
	},
	Exits: []Exit{
		{Code: 0, Means: "green: every rung climbed is green, and the result was written with its record"},
		{Code: 1, Means: "red: a rung is not green; the result was written (red), and next names the fix"},
		{Code: 2, Means: "bad input: a flag ladder does not take, no task named for a local climb or the task named is none\n" +
			"(no-task), a rung asked for that bonsai.yaml does not define (rung-not-defined)"},
		{Code: 3, Means: "could not run or write: git missing, the Bonsai home, a file that cannot be read, the result or its\nrecord that cannot be written"},
		{Code: 4, Means: "wrong state: not a linked checkout, its main checkout not verified on this machine or not on its base\n" +
			"branch (main-not-verified), the main checkout's bonsai.yaml not as committed at its HEAD (config-uncommitted)\n" +
			"or its committed settings looser than the ones this machine accepted (config-looser), another climb holding\n" +
			"this home's lock past its budget (ladder-busy), a test build outside its scratch folder (test-build)"},
	},
	Examples: []string{"bonsai ladder --task T-0901 --json", "bonsai ladder --ci --json"},
	Refused:  func(c *call, e *engine.Error) encoder { return ladder.Refused(e) },
	Run:      runLadder,
}

// runLadder climbs once. A signal (Ctrl-C, SIGTERM) ends the running rung as at its timeout and skips the rest, and the
// climb then writes its red result; during the new-tests check, after the last rung, it ends the check, whose new
// tests read "check not run", and the climb writes its result as the rungs left it.
func runLadder(c *call) int {
	if len(c.rest) > 0 {
		return c.refuse(c.flagError("ladder takes no %+q", c.rest[0]))
	}
	dir, err := os.Getwd()
	if err != nil {
		return c.fail(&engine.Error{Code: "read-failed", Exit: exitRuntime, What: "the current folder cannot be read",
			Next: "run it from a folder inside the project, or name the checkout with --root"})
	}
	stop, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	o := ladder.Options{Dir: dir, Root: c.value("--root"), Task: c.value("--task"), CI: c.has("--ci"), Stop: stop,
		Notice: c.stderr}
	if !c.json {
		o.Out = c.stdout
	}
	a := ladder.Run(o)
	if a.Error != nil {
		noted(c.word, a.Error)
	}
	if c.json {
		return c.printDoc(climbDoc{a.Doc}, a.Exit)
	}
	if a.Error != nil {
		e := *a.Error
		e.What, e.Next = engine.ASCII(trimStop(e.What)), engine.ASCII(e.Next)
		_, _ = io.WriteString(c.stderr, refusalText(c.word, &e, c.stderr))
	}
	return a.Exit
}

// climbDoc is the climb document for printDoc: format.Climb's writer holds it to bonsai.climb/1.
type climbDoc struct{ *format.Climb }

func (d climbDoc) Encode() ([]byte, error) { return d.Climb.Encode() }

// trimStop takes a sentence's final full stop off.
func trimStop(s string) string {
	for len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}
