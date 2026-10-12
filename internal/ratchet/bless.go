package ratchet

// The Bless step (design/plan-5.md, 5.4.4 notes 2, 3 and 5), once a climb's result and its fingerprint are written:
//
//  1. Never under --ci (a CI result is never a task's proof), nor on a climb whose branch is not the base branch: such a
//     climb files and resolves nothing.
//  2. On the base branch, whatever its task and whatever its colour, every open ladder:bless-* ask of another task is
//     worked out again (offer.go) from that task's result in the main checkout's .bonsai/local/ladder/ and the floors as
//     they now are, and resolved when its offer is gone (HEAD moved, the climb red or dirty, no rise left), its reason
//     in the human lines. One whose offer stands stays open: it is never filed again from a result another climb wrote.
//  3. Then the climb's own task: its offer, worked out from the result this climb wrote (its bytes' SHA-256, never the
//     file read back) and the climb's own new-tests lines (Result.Checks, which its why gives), is filed
//     (asks.FileBless; filing an open key again writes a new file record, so this climb's numbers replace the earlier
//     ones); with no offer, its open ask is resolved with the reason. So a red or dirty climb on the base branch files
//     nothing, and closes the offer its own task held.
//
// The records written are the climb's bless, in the order written: the other tasks' resolves by key, then the climb's
// own. A write that fails is a note, never a colour: the result stands as written.

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Step is the climb whose Bless step this is.
type Step struct {
	Place  asks.Place // where the asks are kept (Place.Local.Main's .bonsai/local/), the climbing session and the clock
	CI     bool       // a --ci climb: no Bless at all
	Task   string     // the climbed task's id
	Result Result     // the climb's own result, as written
	Where  Where      // the base branch, the main checkout's branch and HEAD now, and the floors now
}

// Done is what the Bless step did.
type Done struct {
	Records []schema.Object // the ask records written (bonsai.ask/1, as stored), in order: the climb's bless
	Lines   []string        // what it did and why, one line each, for the human output
	Notes   []string        // what it could not do, one line each (never a colour)
}

// Bless runs the Bless step (this file's comment).
func Bless(s Step) Done {
	var d Done
	l := s.Result.Ladder
	if s.CI || l == nil {
		return d
	}
	if s.Where.Base == "" || l.Git.Branch != s.Where.Base {
		if rose(l) {
			_, why := OfferFor(s.Task, s.Result, s.Where)
			d.Lines = append(d.Lines, "bless: not offered: "+why)
		}
		return d
	}
	main := s.Place.Local.Main
	own := asks.BlessKey(s.Task)
	open := map[string]bool{}
	if book, err := asks.Read(main, ""); err != nil {
		d.Notes = append(d.Notes, "the open Bless asks cannot be read ("+oneLine(err)+"), so none was worked out again")
	} else {
		var others []string
		for _, e := range book.Entries(false) {
			if _, ok := asks.IsBless(e.Key); ok {
				open[e.Key] = true
				if e.Key != own {
					others = append(others, e.Key)
				}
			}
		}
		sort.Strings(others)
		for _, key := range others {
			task, _ := asks.IsBless(key)
			if o, why := OfferFor(task, readResult(main, task), s.Where); o == nil {
				d.resolve(s.Place, key, why)
			}
		}
	}
	if own == "" {
		d.Notes = append(d.Notes, "the task "+s.Task+" cannot make a Bless ask's key, so nothing was offered for it")
		return d
	}
	o, why := OfferFor(s.Task, s.Result, s.Where)
	switch {
	case o != nil:
		done, e := asks.FileBless(s.Place, asks.Bless{Task: s.Task, Title: Title(s.Task), Why: Why(o), Then: Then, Data: Data(o)})
		d.take(done, e, "the Bless ask "+own+" cannot be filed")
		if e == nil || done != nil {
			d.Lines = append(d.Lines, "bless: offered "+own+" ("+risesText(o)+"): a person blesses it, or raises the floor in "+
				"bonsai.yaml's ratchets to this climb's count")
		}
	case open[own]:
		d.resolve(s.Place, own, why)
	case rose(l):
		d.Lines = append(d.Lines, "bless: not offered: "+why)
	}
	return d
}

// resolve resolves key, whose offer is gone for why.
func (d *Done) resolve(p asks.Place, key, why string) {
	done, e := asks.ResolveLadder(p, key)
	d.take(done, e, "the Bless ask "+key+" cannot be resolved")
	if e == nil || done != nil {
		d.Lines = append(d.Lines, "bless: resolved "+key+": its offer is gone: "+why)
	}
}

// take keeps a record written, and notes a write that failed (a partly-written one keeps its record).
func (d *Done) take(done *asks.Done, e *asks.Error, what string) {
	if done != nil && done.Written != nil {
		d.Records = append(d.Records, done.Written)
	}
	if e != nil {
		d.Notes = append(d.Notes, what+": "+e.Error())
	}
}

// readResult reads task's result in the main checkout's ladder folder, its bytes' SHA-256 with it.
func readResult(main, task string) Result {
	rel := workspace.LocalDir + "/ladder/" + task + ".json"
	raw, err := os.ReadFile(filepath.Join(main, filepath.FromSlash(rel)))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return Result{Missing: "there is no ladder result for " + task + " (" + rel + ")"}
	case err != nil:
		return Result{Missing: rel + " cannot be read (" + oneLine(err) + ")"}
	}
	l, err := format.ReadLadder(raw)
	if err != nil {
		return Result{Missing: rel + " is not a ladder result (" + oneLine(err) + ")"}
	}
	sum := sha256.Sum256(raw)
	return Result{Ladder: l, SHA256: hex.EncodeToString(sum[:])}
}

// rose reports whether any green rung's ratchet is above its floor (was): a climb with nothing to offer says nothing.
func rose(l *format.Ladder) bool {
	for _, r := range l.Rungs {
		if r.Status == "green" && r.Ratchet != nil && Above(r.Ratchet.Now, r.Ratchet.Was) {
			return true
		}
	}
	return false
}

// risesText names an offer's rises: "go_tests 812 -> 815, go_vet 3 -> 4".
func risesText(o *Offer) string {
	var parts []string
	for _, r := range o.Rises {
		parts = append(parts, r.Name+" "+string(r.Was)+" -> "+string(r.Now))
	}
	return strings.Join(parts, ", ")
}

// oneLine is an error's first line.
func oneLine(err error) string {
	s := err.Error()
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
