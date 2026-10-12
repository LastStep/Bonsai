package ratchet

// The Bless offer (design/plan-5.md, 5.4.4 note 2; contract §9.1, §9.2; today's blessOffer): worked out from one task's
// result and things as they now are (the main checkout's branch and HEAD, bonsai.yaml's floors), never from an ask's
// stored data. There is an offer when the result is the task's, this workspace's and local; green; clean; made on the
// base branch, in the main checkout; the main checkout is still on that branch, at the result's commit; and at least
// one green rung's ratchet has its count above its floor, its name a key of ratchets. Otherwise there is none, and why
// says the first condition that fails, in that order.
//
// The ask's words and data (note 2): key ladder:bless-<task>, title "Bless <task>'s ratchets", why naming each rise and
// its new-tests numbers, then "Blessing raises the floor in bonsai.yaml from that run; it runs nothing.", and data
// {task, head, result_sha, rises: [{name, was, now, new_tests}]} (contract §9.1), result_sha the result's SHA-256 in 64
// lower-case hex, as the climb document's sha256 gives it. A rise's new-tests words in the why are the climb's own
// check in one line (internal/newtests' Summary: the counts, the base, and why a check was not run) when the offer is
// worked out from the result that climb wrote (Result.Checks); an offer worked out again from a result read back has
// only the counts, and says them.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// MaxRises is how many rises one Bless ask carries, as today; the why says how many more rose.
const MaxRises = 10

// WhyMax is the most a Bless ask's why holds, the ask's own limit (asks.WhyMax).
const WhyMax = 600

// Then is a Bless ask's then (note 2).
const Then = "Blessing raises the floor in bonsai.yaml from that run; it runs nothing."

// Rise is one ratchet above its floor: an item of the offer's rises.
type Rise struct {
	Name     string
	Was      json.Number            // its floor, ratchets[name] as bonsai.yaml now holds it
	Now      json.Number            // the result's count
	NewTests *format.LadderNewTests // the result's new_tests for it, nil when the check did not run
	// Checked is the climb's own new-tests check in one line (internal/newtests' Summary), which the why gives in place
	// of the counts' words; "" for an offer worked out again from a result read back, which keeps the counts' words.
	Checked string
}

// Offer is a Bless offer: the ask's data.
type Offer struct {
	Task      string
	Head      string // the commit the result proves, which is the main checkout's HEAD
	ResultSHA string // the SHA-256 of the result's bytes, 64 lower-case hex
	Base      string // the base branch the result was made on
	Rises     []Rise // at most MaxRises, in the result's rung order
	More      int    // how many more rose than the ask carries
}

// Result is a task's ladder result, as an offer is worked out from it.
type Result struct {
	Ladder *format.Ladder // nil when there is none, or it does not read
	SHA256 string         // the SHA-256 of its bytes, 64 lower-case hex
	// Elsewhere says the climb that wrote it ran outside the main checkout (a worktree holding the base branch). Only the
	// climb's own result knows it; one read back from its file was made on the base branch, which the main checkout
	// holds.
	Elsewhere bool
	Missing   string // why there is no Ladder ("" when there is one)
	// Checks are the climb's own new-tests checks in one line each (internal/newtests' Summary), one per rung of
	// Ladder.Rungs in its order, "" for a rung with none. Only the climb that wrote the result has them: nil for a
	// result read back from its file, whose offer keeps the counts' words.
	Checks []string
}

// Where is what an offer is worked out against, as things now are.
type Where struct {
	Workspace string        // this workspace's id
	Base      string        // the base branch: the branch the main checkout held when the climb began; "" when detached
	Branch    string        // the branch the main checkout holds now ("" when detached)
	Head      string        // the main checkout's HEAD now
	Floors    schema.Object // bonsai.yaml's ratchets as they now are
}

// OfferFor works out task's Bless offer from its result (this file's comment): the offer, or nil and why there is
// none.
func OfferFor(task string, r Result, w Where) (*Offer, string) {
	l := r.Ladder
	short := func(sha string) string {
		if len(sha) > 8 {
			return sha[:8]
		}
		if sha == "" {
			return "(no commit)"
		}
		return sha
	}
	switch {
	case l == nil:
		if r.Missing != "" {
			return nil, r.Missing
		}
		return nil, "there is no ladder result for " + task
	case l.Task == nil || *l.Task != task:
		named := "no task"
		if l.Task != nil {
			named = strconv.QuoteToASCII(*l.Task)
		}
		return nil, "the result for " + task + " names " + named
	case l.Workspace != w.Workspace:
		return nil, "the result for " + task + " is another workspace's (" + strconv.QuoteToASCII(l.Workspace) + ")"
	case l.Mode != "local":
		return nil, "the last climb for " + task + " was a " + strconv.QuoteToASCII(l.Mode) + " climb, which is never a task's proof"
	case !l.Green:
		return nil, "the last climb for " + task + " was not green"
	case l.Git.Dirty:
		return nil, "the last climb for " + task + " was dirty (a dirty climb proves no commit)"
	case w.Base == "":
		return nil, "the main checkout's HEAD is detached, so there is no base branch"
	case l.Git.Branch != w.Base:
		return nil, "the last climb for " + task + " was on " + strconv.QuoteToASCII(l.Git.Branch) + ", not the base branch " + w.Base
	case r.Elsewhere:
		return nil, "the last climb for " + task + " ran in a worktree, not the main checkout"
	case w.Branch != w.Base:
		return nil, "the main checkout is no longer on " + w.Base
	case !strings.EqualFold(l.Git.SHA, w.Head) || w.Head == "":
		return nil, "the last climb for " + task + " was made at " + short(l.Git.SHA) + ", and HEAD is " + short(w.Head) + " (HEAD moved)"
	}
	o := &Offer{Task: task, Head: strings.ToLower(l.Git.SHA), ResultSHA: strings.ToLower(r.SHA256), Base: w.Base}
	seen := map[string]bool{}
	for i, rung := range l.Rungs {
		rt := rung.Ratchet
		if rung.Status != "green" || rt == nil || !IsName(rt.Name) || seen[rt.Name] {
			continue
		}
		was, ok := floorOf(w.Floors, rt.Name)
		if !ok || !Above(rt.Now, was) {
			continue
		}
		seen[rt.Name] = true
		if len(o.Rises) == MaxRises {
			o.More++
			continue
		}
		checked := ""
		if i < len(r.Checks) && rt.NewTests != nil {
			checked = r.Checks[i]
		}
		o.Rises = append(o.Rises, Rise{Name: rt.Name, Was: was, Now: rt.Now, NewTests: rt.NewTests, Checked: checked})
	}
	if len(o.Rises) == 0 {
		return nil, "the last climb for " + task + " measured no ratchet above its floor in bonsai.yaml's ratchets"
	}
	return o, "offered"
}

// floorOf is a key of ratchets read as a number: ok is false when ratchets has no such key, or its value is not a
// number (such a floor offers nothing).
func floorOf(floors schema.Object, name string) (json.Number, bool) {
	v, ok := floors.Get(name)
	if !ok {
		return "", false
	}
	var n json.Number
	switch x := v.(type) {
	case json.Number:
		n = x
	case string:
		n = json.Number(x)
	case int64:
		n = json.Number(strconv.FormatInt(x, 10))
	case int:
		n = json.Number(strconv.Itoa(x))
	case float64:
		n = json.Number(strconv.FormatFloat(x, 'f', -1, 64))
	default:
		return "", false
	}
	if _, ok := Number(n); !ok {
		return "", false
	}
	return n, true
}

// Title is a Bless ask's title (note 2).
func Title(task string) string { return "Bless " + task + "'s ratchets" }

// Why is a Bless ask's why: the climb, then each rise with its new-tests numbers, in at most WhyMax characters. Each
// rise's numbers are its check's own line (Rise.Checked) when every rise fits with them, else the counts' words. Rises
// that do not fit are counted as in the ask's data, and rises past the MaxRises an ask carries as beyond it; a base
// branch's name past 100 characters is shown by its first 100.
func Why(o *Offer) string {
	if s, all := why(o, true); all {
		return s
	}
	s, _ := why(o, false)
	return s
}

// why is Why with each rise's numbers from its check's line (checked) or from its counts; all says every rise fit.
func why(o *Offer, checked bool) (string, bool) {
	head := o.Head
	if len(head) > 8 {
		head = head[:8]
	}
	base := []rune(o.Base)
	if len(base) > 100 {
		base = append(base[:100], []rune("...")...)
	}
	start := "A green, clean climb of " + string(base) + " at " + head + ": "
	var parts []string
	for _, r := range o.Rises {
		words := newTestsText(r.NewTests)
		if checked && r.Checked != "" {
			words = r.Checked
		}
		parts = append(parts, r.Name+" rose from "+string(r.Was)+" to "+string(r.Now)+", "+words)
	}
	beyond := ""
	if o.More > 0 {
		beyond = fmt.Sprintf("; and %d more %s, beyond the %d an ask carries", o.More, plural(o.More, "rise", "rises"), MaxRises)
	}
	for kept := len(parts); kept > 0; kept-- {
		s := start + strings.Join(parts[:kept], "; ")
		if n := len(parts) - kept; n > 0 {
			s += fmt.Sprintf("; and %d more %s in the ask's data", n, plural(n, "rise", "rises"))
		}
		if s += beyond + "."; len([]rune(s)) <= WhyMax {
			return s, kept == len(parts)
		}
	}
	return start + fmt.Sprintf("%d %s rose, each in the ask's data.", len(parts), plural(len(parts), "ratchet", "ratchets")), len(parts) == 0
}

// newTestsText is one rise's new-tests numbers in words (5.4.5's check), or that they were not checked.
func newTestsText(nt *format.LadderNewTests) string {
	if nt == nil {
		return "its new tests not checked on the base"
	}
	if nt.Count == 0 {
		return "no new test named"
	}
	s := fmt.Sprintf("%d new %s: %d failed on the base, %d passed there (they prove nothing new), %d not run", nt.Count,
		plural(int(nt.Count), "test", "tests"), nt.FailedOnBase, nt.PassedOnBase, nt.NotRun)
	if nt.Base == "" {
		s += " (no base was found)"
	}
	return s
}

// Data is a Bless ask's data (contract §9.1): {task, head, result_sha, rises: [{name, was, now, new_tests}]}.
func Data(o *Offer) schema.Object {
	rises := []any{}
	for _, r := range o.Rises {
		var nt any
		if r.NewTests != nil {
			nt = schema.Object{{Key: "base", Value: r.NewTests.Base}, {Key: "count", Value: r.NewTests.Count},
				{Key: "failed_on_base", Value: r.NewTests.FailedOnBase}, {Key: "passed_on_base", Value: r.NewTests.PassedOnBase},
				{Key: "not_run", Value: r.NewTests.NotRun}}
		}
		rises = append(rises, schema.Object{{Key: "name", Value: r.Name}, {Key: "was", Value: r.Was}, {Key: "now", Value: r.Now},
			{Key: "new_tests", Value: nt}})
	}
	return schema.Object{{Key: "task", Value: o.Task}, {Key: "head", Value: o.Head}, {Key: "result_sha", Value: o.ResultSHA},
		{Key: "rises", Value: rises}}
}
