package ratchet

// The rise step (design/plan-5.md, 5.4.4 item 5 and 5.4.5 note 1): on a local climb that is green and clean, for each
// rung's ratchet whose count is above its floor, the runner calls NewTests once, and writes what it gives into the
// ratchet's new_tests, so the result the stop gate reads and the Bless ask (its why and data.rises[].new_tests, built
// from that result) carry the numbers. On any other climb new_tests stays null, so a builder's climbs while it works
// cost nothing more. A rise is the count above was, whether or not the name is a key of ratchets (5.4.5 note 1); only a
// key of ratchets is offered for Bless (offer.go).
//
// NewTests is the one call site step 5.4.5 wires its check into: its last commit replaces the body below with
// internal/newtests' check (the new tests named from the rung's output, run on the base in a temporary worktree,
// "failed", "proves nothing new" and "check not run" told apart).

import (
	"context"
	"encoding/json"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
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
	Timeout   time.Duration   // the rung's limit, which each base run keeps
	Stop      context.Context // the climb's stop signal
}

// NewTests is the rise step's one call (this file's comment): the rise's new tests run on the base, or nil when the
// check is not wired in. It gives nil until step 5.4.5's last commit replaces this body with its check.
func NewTests(in RiseInput) *format.LadderNewTests {
	return nil
}
