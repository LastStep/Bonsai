package guard

// Fresh stamps checked (design/contract.md section 23, "Time"; records/plans/P-T-5795.md step 6 and its amendment 2):
// Bonsai is the only clock and every recorded time is UTC, so the guard refuses (time-not-now) an Edit or Write that
// writes a fresh stamp not of Bonsai's clock. It runs in every mode, before command mode's gate in tripwire().
//
// The fields judged, by document kind (documentOf), are the stamp fields' one home, workspace.BonsaiStamps and
// workspace.PackStamps (internal/workspace/stamps.go), which bonsai check's future-stamp finding reads too: a task's
// created, started and finished; a run report's started and finished (record times); STATE's and a memory note's
// updated; a pack kind's stamp fields, read from the lock in full, only when such a document is judged.
//
// A stamp is fresh when a listed field goes from empty (absent, null or "") to a value, when a new file holds it, and,
// for STATE's and a memory note's updated (never empty: "the day it was rewritten"), whenever it changes. A fresh time
// must read as a record time and lie within stampWindow of the guard's clock in UTC; a fresh day must read as a day,
// not after the UTC today nor more than stampDays before it. Every other change passes: a correction of a value already
// there, a value taken back to null, a field left as it was, a field outside the list.
//
// The check judges what the file holds after the call as tripwire does (afterOf); when that cannot be told (a delete or
// move, a NotebookEdit, an old_string not in the file), or the frontmatter does not read, it judges nothing: the
// format is held by each kind's format elsewhere. Rung 0 never runs it (noClock): it judges finished work, hours after
// it was written, where a fresh stamp's clock window means nothing.

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/timefmt"
	"github.com/LastStep/Bonsai/internal/workspace"
)

const (
	stampWindow = workspace.StampWindow // a fresh time's distance from the guard's clock, at most
	stampDays   = 2                     // a fresh day's days before the UTC today, at most
)

// stampField is one field the check judges (workspace.StampField, the list's one home).
type stampField = workspace.StampField

// stampFields gives the fields a document kind's stamps are in; none for a kind with none.
func (j *judge) stampFields(x judged, doc *docKind) []stampField {
	if doc.pack != "" {
		return j.packStamps(x, doc)
	}
	return workspace.BonsaiStamps(doc.kind)
}

// packStamps gives the fields a pack's kind names in its stamp, each a day, in the stamp's order with none twice; none
// when the lock cannot be read (the lean read already judged it) or the kind has no stamp.
func (j *judge) packStamps(x judged, doc *docKind) []stampField {
	lock, err := workspace.LoadLock(j.listsOf(x.p).from)
	if err != nil {
		return nil
	}
	var stamp schema.Object
	for _, lp := range lock.Packs {
		if lp.ID != doc.pack {
			continue
		}
		d, err := lp.Declared()
		if err != nil {
			return nil
		}
		for _, k := range d.Documents {
			if k.Kind == doc.kind {
				stamp = append(stamp, k.Stamp...)
			}
		}
	}
	return workspace.PackStamps(stamp)
}

// freshStamps judges a change of a document by the fresh-stamp rule; false with the refusal when a fresh stamp is not
// of Bonsai's clock. broken says the file after the change also fails its kind's format (one of Bonsai's kinds), so
// that command mode's bad-format, the plainer refusal of the two, may answer in its place there.
func (j *judge) freshStamps(c change, x judged, doc *docKind) (d Decision, ok bool, broken bool) {
	if j.noClock {
		return Decision{}, true, false
	}
	fields := j.stampFields(x, doc)
	if len(fields) == 0 {
		return Decision{}, true, false
	}
	before, err := j.readBefore(x.abs)
	exists := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Decision{}, true, false
	}
	after, known := afterOf(c.in, before, exists)
	if !known {
		return Decision{}, true, false
	}
	av, read := frontValues(after)
	if !read {
		return Decision{}, true, false
	}
	var bv map[string]string
	if exists {
		if bv, read = frontValues(before); !read {
			bv = map[string]string{}
		}
	}
	now := j.clock().UTC()
	for _, f := range fields {
		v := av[f.Name]
		if v == "" {
			continue
		}
		was, had := bv[f.Name]
		fresh := !exists || !had || was == "" || (f.Always && was != v)
		if !fresh {
			continue
		}
		if why := stampProblem(f, v, now); why != "" {
			return j.stampRefusal(c, x, doc, f, v, why, now), false, doc.pack == "" && bonsaiFormat(doc.kind, after) != nil
		}
	}
	return Decision{}, true, false
}

// stampProblem says why a fresh stamp is not of Bonsai's clock at now (UTC), "" when it is.
func stampProblem(f stampField, v string, now time.Time) string {
	if f.Time {
		t, err := timefmt.ParseRecord(v)
		if err != nil {
			return "it is not a recorded time (YYYY-MM-DD HH:MM, in UTC)"
		}
		if d := t.Sub(now); d > stampWindow || d < -stampWindow {
			return fmt.Sprintf("it is %s from Bonsai's clock, more than the %d minutes a fresh time may be", distance(d),
				int(stampWindow/time.Minute))
		}
		return ""
	}
	day, err := time.Parse("2006-01-02", v)
	if err != nil {
		return "it is not a day (YYYY-MM-DD, the UTC day)"
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	switch {
	case day.After(today):
		return "it is after the UTC today"
	case day.Before(today.AddDate(0, 0, -stampDays)):
		return fmt.Sprintf("it is more than %d days before the UTC today", stampDays)
	}
	return ""
}

// distance writes a duration's size as people read it: "3 h 12 min", "25 min".
func distance(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	m := int(d / time.Minute)
	if m >= 60 {
		return fmt.Sprintf("%d h %d min", m/60, m%60)
	}
	return fmt.Sprintf("%d min", m)
}

// stampRefusal is the refusal of a fresh stamp: the field, the value seen and Bonsai's UTC now, and bonsai now as the
// next step.
func (j *judge) stampRefusal(c change, x judged, doc *docKind, f stampField, v, why string, now time.Time) Decision {
	shown := quote(x.rel, 120)
	kind, form, next := "day", timefmt.RecordDay(now), "run: bonsai now --day, and write the UTC day it prints"
	if f.Time {
		kind, form, next = "time", timefmt.Record(now), "run: bonsai now, and write the UTC time it prints"
	}
	d := deny(RuleTimeNotNow, fmt.Sprintf("%s of %s refused: the %s's %s %s is a fresh %s not of Bonsai's clock: %s "+
		"(Bonsai's UTC now: %s)", c.tool, shown, doc.kind, f.Name, quote(v, 40), kind, why, form), next)
	d.Target = target(j.recordIn(x.chain), []string{x.abs})
	return d
}

// frontValues reads a document's frontmatter under format 1, or format 0 when it has no format: line: each top-level
// field's value as text (null as ""). A frontmatter the reader refuses (a plain value such as 2026-10-21 09:16 IST
// that format 1 will not read unquoted) is scanned line by line instead, as scanDoc reads a status line, so a stamp
// written in a wrong form is still seen; false only when there is no frontmatter at all.
func frontValues(raw []byte) (map[string]string, bool) {
	var m *reader.Map
	r := reader.ReadMarkdown(raw)
	switch r.Outcome {
	case reader.Accepted:
		m = r.Value
	case reader.Format0:
		if m0, ok := reader.ReadFormat0Markdown(raw).Map(); ok {
			m = m0
		}
	}
	if m == nil {
		return scanValues(raw)
	}
	out := map[string]string{}
	for _, e := range m.Entries() {
		out[e.Key] = scalarText(e.Value)
	}
	return out, true
}

// scanValues reads the top-level key: value lines of a frontmatter that does not read, as text: a comment after " #"
// and one pair of quotes taken off, null and ~ as "". A value that opens a structure or a block ([, {, |, >) or is an
// anchor, alias or tag (&, *, !) is no stamp typed in a wrong form but a broken file, its kind's format's to judge
// (command mode's bad-format): it is left out, so the guard and rung 0 still agree on it in agents mode.
func scanValues(raw []byte) (map[string]string, bool) {
	lines := strings.Split(lf(string(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], " ") != "---" {
		return nil, false
	}
	out := map[string]string{}
	for _, l := range lines[1:] {
		if strings.TrimRight(l, " ") == "---" {
			break
		}
		key, val, ok := strings.Cut(l, ":")
		if !ok || key == "" || strings.TrimSpace(key) != key {
			continue
		}
		if i := strings.Index(val, " #"); i >= 0 {
			val = val[:i]
		}
		val = strings.TrimSpace(val)
		if val != "" && strings.ContainsRune("[{|>&*!", rune(val[0])) {
			continue
		}
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		}
		if val == "null" || val == "~" {
			val = ""
		}
		out[strings.Trim(key, `"'`)] = val
	}
	return out, true
}
