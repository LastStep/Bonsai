package ratchet

// The rise step (design/plan-5.md, 5.4.4 item 5 and 5.4.5 note 1): on a local climb that is green and clean, for each
// rung's ratchet whose count is above its floor, the runner calls NewTests once, and writes what it gives into the
// ratchet's new_tests, so the result the stop gate reads and the Bless ask (its data.rises[].new_tests, built from that
// result, and its why, from the check's own line) carry the numbers. On any other climb new_tests stays null, so a
// builder's climbs while it works cost nothing more. A rise is the count above was, whether or not the name is a key
// of ratchets (5.4.5 note 1); only a key of ratchets is offered for Bless (offer.go).
//
// NewTests is internal/newtests' check (step 5.4.5): the new tests named from the rung's output at HEAD, the rung run
// twice on the base in a temporary worktree (as it is, then with the branch's changed test files copied in), and
// "failed on the base", "proves nothing new" and "check not run" told apart. The base is newtests.ChooseBase's: the
// merge base of HEAD with the base branch, else (on the base branch itself) the base the task's last result recorded,
// else none ("check not run: no base"). Each base run keeps the rung's limit and the climb's environment, and the
// climb's stop ends a running one: a check stopped or past its limit gives new_tests with every new test "check not
// run", never an error, and leaves no worktree.

import (
	"context"
	"encoding/json"
	"math"
	"math/big"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/newtests"
	"github.com/LastStep/Bonsai/internal/testout"
)

// RiseInput is one ratchet that rose on a local, green, clean climb: what the new-tests check reads.
type RiseInput struct {
	Name      string          // the ratchet's name
	Was, Now  json.Number     // its floor (ratchets[name], 0 when it names none) and this climb's count, above it
	Rung      format.Rung     // the rung as bonsai.yaml defines it: its command, tests form, base_setup and timeout
	Tests     *testout.Report // the rung's tests as this climb read them at HEAD; nil when it has no tests form Bonsai reads
	Root      string          // the checkout climbed: its top folder, absolute, system separators
	Main      string          // the verified main checkout
	Task      string          // the climbed task's id
	Head      string          // HEAD, the commit climbed (full, lower-case)
	Base      string          // the base branch: the branch the main checkout holds; "" when its HEAD is detached
	MergeBase string          // git merge-base of Base and HEAD; "" when there is no base or git finds none
	Prior     *format.Ladder  // the task's earlier result, or nil (a climb on the base branch reads the base it recorded)
	Env       []string        // the rungs' environment
	Timeout   time.Duration   // the rung's limit, which each base run keeps (newtests reads it from Rung, the same)
	Stop      context.Context // the climb's stop signal
}

// Checked is one rise's check, as the runner carries it.
type Checked struct {
	NewTests *format.LadderNewTests // the ratchet's new_tests
	Lines    []string               // the human lines, under the rung's: the summary, the names (at most 20 a group), the notes
	Summary  string                 // the check in one line, for the Bless ask's why (Result.Checks)
	Took     time.Duration          // the check's whole time
}

// check is the new-tests check, a variable for the tests.
var check = newtests.Check

// NewTests is the rise step's one call (this file's comment): the rise's new tests run on the base.
func NewTests(in RiseInput) Checked {
	var head testout.Report
	if in.Tests != nil {
		head = *in.Tests
	}
	r := check(newtests.Input{Root: in.Root, Rung: in.Rung, Base: newtests.ChooseBase(in.MergeBase, in.Head, in.Prior), Head: head,
		Rise: riseOf(in.Now, in.Was), Env: in.Env, Stop: in.Stop})
	nt := r.NewTests
	return Checked{NewTests: &nt, Lines: r.Lines(), Summary: r.Summary(), Took: r.Took}
}

// riseOf is how far a count rose above its floor, as a count of tests (new_tests' count when the new tests cannot be
// named): now - was, a part of one counted as one (a capture's 87.5 above 87 rose), 0 when it did not rise or either
// does not read as a number, and at most the largest int64.
func riseOf(now, was json.Number) int64 {
	a, okA := Number(now)
	b, okB := Number(was)
	if !okA || !okB || a.Cmp(b) <= 0 {
		return 0
	}
	n, acc := new(big.Float).SetPrec(256).Sub(a, b).Int(nil)
	if acc == big.Below {
		n.Add(n, big.NewInt(1)) // the part of one
	}
	if !n.IsInt64() {
		return math.MaxInt64
	}
	return n.Int64()
}
