package workspace

// The checks' settings this machine last accepted for a workspace (5.4's second fix round, K1; Rohan, 13 Oct: make
// "stricter only" code). The ladder runner and the stop gate read bonsai.yaml as committed at the main checkout's HEAD
// (stopgate.CommittedConfig), and a commit on the base branch is history: an agent that committed a lowered floor
// there, or moved the branch to such a commit (git reset --hard, a ref moved by hand), made a red branch climb green
// and the gate let the session end. So this machine's record of the main checkout (record.go) keeps the settings Bonsai
// last accepted, its checks field, and a local climb and the stop gate refuse committed settings looser than them
// (stopgate.Accepted: the error word config-looser, the gate's rule stop-config-looser), naming each loosening.
// Settings stricter than them, or equal, are accepted, and a local climb brings the record up to them (RaiseChecks), so
// floors raised at a part's end need no person. Recording looser settings is a person's step: bonsai update (and init)
// records the settings bonsai.yaml holds (RecordCheckoutWith), and refuses looser ones in an agent's session
// (cmd/bonsai, engine.ChecksLooser), as it refuses another base branch.
//
// What is kept and compared (Checks), and what reads as looser (Loosenings), field by field of bonsai.yaml:
//   - ladder_floor: a rung taken out of it. One added is stricter.
//   - ladder: a rung no longer there; or one whose kind, command, tests, capture or base_setup changed in any way
//     (Bonsai cannot tell a stricter command from a looser one, and base_setup prepares the base on which new tests
//     must fail: a setup that breaks the base makes a test that proves nothing look new); whose required went from
//     true to false; or whose ratchet was taken out or renamed (one added to a rung that had none is stricter). Rungs
//     are matched by number, and within one number in the file's order (two rungs of one number are both climbed, so
//     a second one added is stricter). A rung added is stricter. Not kept: name and means (words a result shows, which
//     decide nothing), timeout_s (a longer limit lets a slow rung finish; it never turns a red rung green, since a
//     rung that runs out of time is red, and the lock's budget is no check) and any key this Bonsai does not know (it
//     reads none of them).
//   - ratchets: a floor lowered or taken out (a ratchet with no floor reads as 0), or one Bonsai cannot read as a
//     number changed. A floor raised, or a new one, is stricter.
//   - ci_marked_tests: a name added (a test named there may skip under --ci, and a local climb reads it as marked). One
//     taken out is stricter.
//   - protected, person_only, never_edit: a path taken out (the guard and rung 0 judge by them). One added is stricter.
//     A glob is compared as written: "a/**" put in place of "a/b/**" reads as looser (a person records it).
//   - documents: a kind's place changed or taken out. The guard's tripwires and rung 0 judge a document by its declared
//     place, and the stop gate reads the task and the run report there: a task folder moved out from under the
//     protected list would let an agent write its own grants. A kind added is stricter.
//
// Not kept, and why: generated (how long Bonsai keeps what it generates in .bonsai/local/, and the session rows; run
// reports it never cleans): no verdict rests on a generated file being kept (the stop gate needs a result proving HEAD
// now; an earlier result is read only when the log fingerprints it, and a missing one reads as none), and a shell
// command can delete those files outright; packs (they take effect through bonsai update, which writes the lock and the
// settings: the packs' protected paths come from the lock's declares, not from bonsai.yaml); id, name and format (no
// check reads them as a setting); and keys this Bonsai does not know.
//
// In the record, checks is a JSON object written byte-stable, every field always written, lists in bonsai.yaml's
// order (rungs in file order, ratchets and documents in their mapping's order):
//
//	"checks": {
//	  "ladder_floor": [0, 1, 2],
//	  "ladder": [
//	    {"rung": 1, "kind": "command", "command": "go test -json ./...", "required": true, "tests": "go-test-json",
//	     "ratchet": "go_tests", "capture": null, "base_setup": null}
//	  ],
//	  "ratchets": {"go_tests": 2},
//	  "ci_marked_tests": [],
//	  "protected": ["bonsai.yaml"],
//	  "person_only": ["bonsai.yaml"],
//	  "never_edit": [],
//	  "documents": {"task": "work/tasks", "run": "work/runs", "answers": "work/answers.md", "memory": "work/memory",
//	                "protocols": "work/protocols"}
//	}
//
// checks is null when none was recorded: a record written before step 5.4's second fix round, or one whose main
// checkout's bonsai.yaml did not read in full when it was written. A local climb and the stop gate fail closed on such
// a record until bonsai update records the settings; that first record is free, as base_branch's is.

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Checks are the settings of bonsai.yaml that set how strict the ladder, the stop gate and the guard are (this file's
// comment): what this machine's record keeps as accepted, and what a climb compares with it.
type Checks struct {
	LadderFloor   []int64
	Ladder        []CheckedRung
	Ratchets      schema.Object // each ratchet's floor, by name, as bonsai.yaml holds it (a number)
	CIMarkedTests []string
	Protected     []string
	PersonOnly    []string
	NeverEdit     []string
	Documents     schema.Object // each document kind's place, by kind, as bonsai.yaml's documents holds it
}

// CheckedRung is what a rung's checking rests on: its number and the fields that decide its verdict.
type CheckedRung struct {
	Rung      int64
	Kind      string
	Command   *string
	Required  bool
	Tests     *string
	Ratchet   any // its ratchet's name, or nil; any other value as read (the field is typed open)
	Capture   any // its patterns by name, or nil; as read
	BaseSetup *string
}

// ChecksOf gives the settings for the checks that bonsai.yaml, read in full, holds; nil for nil.
func ChecksOf(w *format.Workspace) *Checks {
	if w == nil {
		return nil
	}
	c := &Checks{LadderFloor: append([]int64{}, w.LadderFloor...), Ratchets: schema.Object{},
		CIMarkedTests: append([]string{}, w.CIMarkedTests...), Protected: append([]string{}, w.Protected...),
		PersonOnly: append([]string{}, w.PersonOnly...), NeverEdit: append([]string{}, w.NeverEdit...)}
	for _, r := range w.Ladder {
		c.Ladder = append(c.Ladder, CheckedRung{Rung: r.Rung, Kind: r.Kind, Command: r.Command, Required: r.Required, Tests: r.Tests,
			Ratchet: canon(r.Ratchet), Capture: canon(r.Capture), BaseSetup: r.BaseSetup})
	}
	for _, f := range w.Ratchets {
		c.Ratchets = append(c.Ratchets, schema.Member{Key: f.Key, Value: canon(f.Value)})
	}
	d := w.Documents
	c.Documents = schema.Object{{Key: "task", Value: d.Task}, {Key: "run", Value: d.Run}, {Key: "answers", Value: d.Answers},
		{Key: "memory", Value: d.Memory}, {Key: "protocols", Value: d.Protocols}}
	for _, f := range d.Extra {
		c.Documents = append(c.Documents, schema.Member{Key: f.Key, Value: canon(f.Value)})
	}
	return c
}

// checksOfFile gives the settings for the checks of the bonsai.yaml in the folder root as the file holds it now, nil
// when it is not there or does not read in full.
func checksOfFile(root string) *Checks {
	raw, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		return nil
	}
	c, err := ReadConfigFull(raw)
	if err != nil {
		return nil
	}
	return ChecksOf(c.Full)
}

// canon gives a JSON value as the record reads it back (written, then read): an object, a list, a json.Number, a
// string, a bool or nil, so a value read from bonsai.yaml and one read from the record compare alike.
func canon(v any) any {
	raw, err := schema.Encode(v)
	if err != nil {
		return v
	}
	back, err := schema.Decode(raw)
	if err != nil {
		return v
	}
	return back
}

// JSON gives the settings as the record writes them (this file's comment).
func (c *Checks) JSON() schema.Object {
	floor := []any{}
	for _, n := range c.LadderFloor {
		floor = append(floor, n)
	}
	rungs := []any{}
	for _, r := range c.Ladder {
		rungs = append(rungs, schema.Object{{Key: "rung", Value: r.Rung}, {Key: "kind", Value: r.Kind}, {Key: "command", Value: textOrNull(r.Command)},
			{Key: "required", Value: r.Required}, {Key: "tests", Value: textOrNull(r.Tests)}, {Key: "ratchet", Value: r.Ratchet},
			{Key: "capture", Value: r.Capture}, {Key: "base_setup", Value: textOrNull(r.BaseSetup)}})
	}
	ratchets, docs := schema.Object{}, schema.Object{}
	ratchets = append(ratchets, c.Ratchets...)
	docs = append(docs, c.Documents...)
	return schema.Object{{Key: "ladder_floor", Value: floor}, {Key: "ladder", Value: rungs}, {Key: "ratchets", Value: ratchets},
		{Key: "ci_marked_tests", Value: texts(c.CIMarkedTests)}, {Key: "protected", Value: texts(c.Protected)},
		{Key: "person_only", Value: texts(c.PersonOnly)}, {Key: "never_edit", Value: texts(c.NeverEdit)}, {Key: "documents", Value: docs}}
}

func textOrNull(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

// sameChecks reports whether two settings are the same, as the record writes them.
func sameChecks(a, b *Checks) bool {
	if a == nil || b == nil {
		return a == b
	}
	x, errA := schema.Encode(a.JSON())
	y, errB := schema.Encode(b.JSON())
	return errA == nil && errB == nil && string(x) == string(y)
}

// readChecks reads a record's checks field; why says what is wrong with one Bonsai does not read, in words that
// follow "its checks ".
func readChecks(v any) (c *Checks, why string) {
	o, ok := v.(schema.Object)
	if !ok {
		return nil, "is " + schema.Show(v) + ", not an object (or null)"
	}
	c = &Checks{Ratchets: schema.Object{}, Documents: schema.Object{}}
	seen := map[string]bool{}
	for _, f := range o {
		seen[f.Key] = true
		switch f.Key {
		case "ladder_floor":
			list, ok := f.Value.([]any)
			if !ok {
				return nil, "has a ladder_floor that is not a list"
			}
			for _, it := range list {
				n, ok := wholeNumber(it)
				if !ok {
					return nil, "has a ladder_floor holding " + schema.Show(it) + ", not a rung number"
				}
				c.LadderFloor = append(c.LadderFloor, n)
			}
		case "ladder":
			list, ok := f.Value.([]any)
			if !ok {
				return nil, "has a ladder that is not a list"
			}
			for _, it := range list {
				r, why := readCheckedRung(it)
				if why != "" {
					return nil, "has a ladder entry " + schema.Show(it) + " that " + why
				}
				c.Ladder = append(c.Ladder, r)
			}
		case "ratchets", "documents":
			m, ok := f.Value.(schema.Object)
			if !ok {
				return nil, "has a " + f.Key + " that is not an object"
			}
			if f.Key == "ratchets" {
				c.Ratchets = append(schema.Object{}, m...)
			} else {
				c.Documents = append(schema.Object{}, m...)
			}
		case "ci_marked_tests", "protected", "person_only", "never_edit":
			list, ok := textList(f.Value)
			if !ok {
				return nil, "has a " + f.Key + " that is not a list of texts"
			}
			switch f.Key {
			case "ci_marked_tests":
				c.CIMarkedTests = list
			case "protected":
				c.Protected = list
			case "person_only":
				c.PersonOnly = list
			default:
				c.NeverEdit = list
			}
		}
	}
	for _, k := range []string{"ladder_floor", "ladder", "ratchets", "ci_marked_tests", "protected", "person_only", "never_edit", "documents"} {
		if !seen[k] {
			return nil, "has no " + k
		}
	}
	return c, ""
}

// readCheckedRung reads one rung of a record's checks; why says what is wrong, in words that follow "that ".
func readCheckedRung(v any) (r CheckedRung, why string) {
	o, ok := v.(schema.Object)
	if !ok {
		return r, "is not an object"
	}
	n, _ := o.Get("rung")
	if r.Rung, ok = wholeNumber(n); !ok {
		return r, "has no rung number"
	}
	kind, ok := o.Get("kind")
	if r.Kind, ok = kind.(string); !ok {
		return r, "has no kind"
	}
	req, _ := o.Get("required")
	if r.Required, ok = req.(bool); !ok {
		return r, "has no required (true or false)"
	}
	for _, f := range []struct {
		key string
		to  **string
	}{{"command", &r.Command}, {"tests", &r.Tests}, {"base_setup", &r.BaseSetup}} {
		switch x, _ := o.Get(f.key); s := x.(type) {
		case nil:
		case string:
			*f.to = &s
		default:
			return r, "has a " + f.key + " that is not a text (or null)"
		}
	}
	r.Ratchet, _ = o.Get("ratchet")
	r.Capture, _ = o.Get("capture")
	return r, ""
}

// wholeNumber reads a rung number as the record holds it: a whole json.Number from 0.
func wholeNumber(v any) (int64, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	i, err := strconv.ParseInt(string(n), 10, 64)
	return i, err == nil && i >= 0
}

func textList(v any) ([]string, bool) {
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := []string{}
	for _, it := range list {
		s, ok := it.(string)
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

// Loosenings compares the settings now (bonsai.yaml as committed) with the ones accepted (was) and gives each way now
// is looser (this file's comment), in ASCII words, in a fixed order: the floor, the rungs, the ratchets,
// ci_marked_tests, the three path lists, the documents. None means now is stricter or the same. A nil was or now
// gives none (the caller says why there is nothing to compare).
func Loosenings(was, now *Checks) []string {
	if was == nil || now == nil {
		return nil
	}
	var out []string
	floor := map[int64]bool{}
	for _, n := range now.LadderFloor {
		floor[n] = true
	}
	for _, n := range was.LadderFloor {
		if !floor[n] {
			out = append(out, "ladder_floor no longer holds rung "+strconv.FormatInt(n, 10))
			floor[n] = true // said once
		}
	}
	out = append(out, rungLoosenings(was.Ladder, now.Ladder)...)
	for _, f := range was.Ratchets {
		name := "ratchets." + f.Key
		v, ok := now.Ratchets.Get(f.Key)
		switch {
		case !ok:
			out = append(out, name+" was taken out (its floor was "+schema.Show(f.Value)+"; a ratchet with no floor reads as 0)")
		case schema.Equal(f.Value, v):
		default:
			a, okA := floorNumber(f.Value)
			b, okB := floorNumber(v)
			switch {
			case okA && okB && b.Cmp(a) < 0:
				out = append(out, name+" fell from "+schema.Show(f.Value)+" to "+schema.Show(v))
			case !okA || !okB:
				out = append(out, name+" changed from "+schema.Show(f.Value)+" to "+schema.Show(v)+", not two numbers Bonsai compares")
			}
		}
	}
	had := map[string]bool{}
	for _, s := range was.CIMarkedTests {
		had[s] = true
	}
	for _, s := range now.CIMarkedTests {
		if !had[s] {
			out = append(out, "ci_marked_tests gained "+strconv.QuoteToASCII(s))
			had[s] = true
		}
	}
	for _, l := range []struct {
		name     string
		was, now []string
	}{{"protected", was.Protected, now.Protected}, {"person_only", was.PersonOnly, now.PersonOnly}, {"never_edit", was.NeverEdit, now.NeverEdit}} {
		has := map[string]bool{}
		for _, s := range l.now {
			has[s] = true
		}
		for _, s := range l.was {
			if !has[s] {
				out = append(out, l.name+" no longer holds "+strconv.QuoteToASCII(s))
				has[s] = true
			}
		}
	}
	for _, f := range was.Documents {
		v, ok := now.Documents.Get(f.Key)
		switch {
		case !ok:
			out = append(out, "documents."+f.Key+" was taken out (it was "+schema.Show(f.Value)+")")
		case !schema.Equal(f.Value, v):
			out = append(out, "documents."+f.Key+" changed from "+schema.Show(f.Value)+" to "+schema.Show(v))
		}
	}
	return out
}

// rungLoosenings compares the rungs (this file's comment): matched by number, and in file order within one number.
func rungLoosenings(was, now []CheckedRung) []string {
	byNumber := map[int64][]CheckedRung{}
	for _, r := range now {
		byNumber[r.Rung] = append(byNumber[r.Rung], r)
	}
	seen := map[int64]int{}
	var out []string
	for _, w := range was {
		i := seen[w.Rung]
		seen[w.Rung]++
		name := "rung " + strconv.FormatInt(w.Rung, 10)
		if i > 0 {
			name += " (its entry " + strconv.Itoa(i+1) + " in the ladder)"
		}
		list := byNumber[w.Rung]
		if i >= len(list) {
			out = append(out, name+" is no longer in the ladder")
			continue
		}
		n := list[i]
		if w.Kind != n.Kind {
			out = append(out, name+"'s kind changed from "+strconv.QuoteToASCII(w.Kind)+" to "+strconv.QuoteToASCII(n.Kind))
		}
		for _, f := range []struct {
			field    string
			was, now *string
		}{{"command", w.Command, n.Command}, {"tests", w.Tests, n.Tests}, {"base_setup", w.BaseSetup, n.BaseSetup}} {
			if !schema.Equal(textOrNull(f.was), textOrNull(f.now)) {
				out = append(out, name+"'s "+f.field+" changed from "+shortValue(textOrNull(f.was))+" to "+shortValue(textOrNull(f.now)))
			}
		}
		if !schema.Equal(w.Capture, n.Capture) {
			out = append(out, name+"'s capture changed from "+shortValue(w.Capture)+" to "+shortValue(n.Capture))
		}
		if w.Required && !n.Required {
			out = append(out, name+" is no longer required")
		}
		if w.Ratchet != nil && !schema.Equal(w.Ratchet, n.Ratchet) {
			if n.Ratchet == nil {
				out = append(out, name+"'s ratchet "+shortValue(w.Ratchet)+" was taken out")
			} else {
				out = append(out, name+"'s ratchet changed from "+shortValue(w.Ratchet)+" to "+shortValue(n.Ratchet))
			}
		}
	}
	return out
}

// floorNumber reads a ratchet's floor as a number Bonsai compares: a json.Number, exactly.
func floorNumber(v any) (*big.Rat, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return nil, false
	}
	return new(big.Rat).SetString(string(n))
}

// shortValue shows a value as one line of JSON in ASCII, cut to 80 characters.
func shortValue(v any) string {
	s := schema.Show(v)
	if len(s) > 80 {
		s = s[:77] + "..."
	}
	return s
}
