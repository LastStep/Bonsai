package engine

// check's warnings on the personal memory layer in the Bonsai home (spec section 10; plan-5 5.6.3, 3;
// workspace/personal.go): read from the home whatever project check runs in, each a warning (never the exit code: the
// layer is the machine's, not the project's, so a project's CI and rung 0 never fail on it), each a person's step,
// since a person writes the layer. With no personal/ folder there is nothing to check but the Windows copy's record.
//
//   - personal-format: the index or a note does not read as bonsai.memory/1, or the index's kind is not index (or a
//     note's is);
//   - personal-budget: the index over its cap, or a note over its cap (MemoryCaps: personal-index, note);
//   - personal-index: an index line linking a note in notes/ that does not exist, or a note no index line links;
//   - personal-secret: what the redactor finds (redact.Find, as the project's secret finding), naming the file, the
//     line and the kind, never the value;
//   - personal-copy: on Windows, from cache/personal-copy.json and the copy alone, never reading the source (a read
//     through \\wsl.localhost can start WSL and take seconds, on every check and every rung 0): personal_from set
//     but never refreshed from it, the last refresh failed, or the copy changed since that refresh. The record is read
//     on every system, so a record that says a refresh failed is reported wherever it is.
//
// check writes nothing in the home.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// personalLink finds a markdown link's target.
var personalLink = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`\]\(([^)\s]+)\)`) })

// checkPersonal adds the personal layer's warnings.
func (r *CheckResult) checkPersonal() {
	if r.Home == "" {
		return
	}
	r.checkPersonalCopy()
	dir := workspace.PersonalPath(r.Home)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return
	}
	show := func(rel string) string { return filepath.ToSlash(filepath.Join(dir, filepath.FromSlash(rel))) }
	edit := func(rel, what string) string {
		return "a person " + what + " " + show(rel) + " (the personal layer is the person's to write; an agent hands over a draft)"
	}
	// The index.
	indexRaw, indexErr := os.ReadFile(filepath.Join(dir, workspace.PersonalIndex))
	linked := map[string]bool{}
	if indexErr == nil {
		r.personalDoc(workspace.PersonalIndex, indexRaw, true, show, edit)
		for _, m := range personalLink().FindAllStringSubmatch(string(indexRaw), -1) {
			target := strings.TrimPrefix(m[1], "./")
			if strings.Contains(target, "://") || !strings.HasSuffix(target, ".md") {
				continue
			}
			linked[target] = true
		}
	}
	// The notes.
	var notes []string
	entries, _ := os.ReadDir(filepath.Join(dir, workspace.PersonalNotes))
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") || !e.Type().IsRegular() {
			continue
		}
		rel := workspace.PersonalNotes + "/" + e.Name()
		notes = append(notes, rel)
		raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		r.personalDoc(rel, raw, false, show, edit)
	}
	sort.Strings(notes)
	if indexErr != nil {
		return
	}
	// The index's lines against the notes.
	have := map[string]bool{}
	for _, n := range notes {
		have[n] = true
	}
	var targets []string
	for t := range linked {
		targets = append(targets, t)
	}
	sort.Strings(targets)
	for _, t := range targets {
		if strings.HasPrefix(t, workspace.PersonalNotes+"/") && !have[t] {
			r.add("personal-index", "", "", "a line of the personal memory index "+show(workspace.PersonalIndex)+" names the note "+t+", which does not exist",
				edit(workspace.PersonalIndex, "takes the line out of, or writes the note "+show(t)+" for,"))
		}
	}
	for _, n := range notes {
		if !linked[n] {
			r.add("personal-index", "", "", "the personal note "+show(n)+" has no line in the personal memory index",
				edit(workspace.PersonalIndex, "adds a line for "+n+" to"))
		}
	}
}

// personalDoc checks one file of the layer: its format, its cap and its secrets.
func (r *CheckResult) personalDoc(rel string, raw []byte, index bool, show func(string) string, edit func(string, string) string) {
	m, err := format.ReadMemory(raw)
	switch {
	case err != nil:
		r.add("personal-format", "", "", show(rel)+" does not read as bonsai.memory/1: "+docError(err), edit(rel, "fixes"))
	case index && m.Kind != "index":
		r.add("personal-format", "", "", "the personal memory index "+show(rel)+" has kind "+ascii(m.Kind)+", not index", edit(rel, "sets kind: index in"))
	case !index && m.Kind == "index":
		r.add("personal-format", "", "", "the personal note "+show(rel)+" has kind index, which is the index's alone", edit(rel, "sets the note's kind (project, feedback or reference) in"))
	}
	if index {
		c, lines := memoryCap("personal-index"), capLines(raw)
		if lines > c.Lines || len(raw) > c.Bytes {
			r.add("personal-budget", "", "", fmt.Sprintf("the personal memory index %s is %d lines and %d bytes, over its %d lines or %s",
				show(rel), lines, len(raw), c.Lines, KB(c.Bytes)), edit(rel, fmt.Sprintf("shortens to at most %d lines and %s (one short line a note)", c.Lines, KB(c.Bytes))))
		}
	} else if c := memoryCap("note"); len(raw) > c.Bytes {
		r.add("personal-budget", "", "", fmt.Sprintf("the personal note %s is %d bytes, over its %s", show(rel), len(raw), KB(c.Bytes)),
			edit(rel, "shortens to at most "+KB(c.Bytes)+", or splits in two,"))
	}
	for _, f := range secretLines(string(raw)) {
		r.add("personal-secret", "", "", fmt.Sprintf("%s holds a secret-shaped string on line %d (%s), in the layer that loads into every session on this machine; its value is not shown here",
			show(rel), f.line, strings.Join(f.kinds, ", ")),
			edit(rel, fmt.Sprintf("takes the secret out of line %d of", f.line))+", and decides whether to rotate it")
	}
}

// checkPersonalCopy adds the personal-copy warning from the Windows copy's record (never reading its source).
func (r *CheckResult) checkPersonalCopy() {
	refresh := "on Windows, in a linked project (it refreshes the copy from personal_from), run: bonsai update --yes"
	rec, err := workspace.ReadPersonalRecord(r.Home)
	recPath := filepath.ToSlash(filepath.Join(r.Home, filepath.FromSlash(workspace.PersonalRecord)))
	if err != nil {
		r.add("personal-copy", "", "", "the personal memory copy's record "+recPath+" cannot be read: "+ascii(oneLine(err.Error())), refresh)
		return
	}
	hs, herr := workspace.LoadHomeSettings(r.Home)
	from := ""
	if herr == nil {
		from = hs.PersonalFrom
	}
	if from != "" && (rec == nil || rec.Source != from) {
		r.add("personal-copy", "", "", "the personal memory copy was never refreshed from "+ascii(from)+" (personal_from) since it was set", refresh)
		return
	}
	if rec == nil {
		return
	}
	if rec.Failed != nil {
		why := ""
		if rec.Why != nil {
			why = *rec.Why
		}
		r.add("personal-copy", "", "", "the last refresh of the personal memory copy from "+ascii(rec.Source)+" failed at "+ascii(*rec.Failed)+": "+ascii(why), refresh)
		return
	}
	if rec.Refreshed == nil {
		return
	}
	have, err := workspace.PersonalFiles(workspace.PersonalPath(r.Home))
	if err != nil {
		r.add("personal-copy", "", "", "the personal memory copy cannot be read: "+ascii(oneLine(err.Error())), refresh)
		return
	}
	if same, differs := workspace.SamePersonal(have, rec.Files); !same {
		r.add("personal-copy", "", "", "the personal memory copy changed since its refresh from "+ascii(rec.Source)+" at "+ascii(*rec.Refreshed)+
			" ("+differs+" differs); WSL's layer is the canonical one, so a change made here is lost at the next refresh",
			"make the change in WSL's layer, then "+refresh)
	}
}

// PersonalLine is one line init or update adds to its text about the personal layer; Warn when something failed.
type PersonalLine struct {
	Text string
	Warn bool
}

// PersonalAfterWrite is init's and update's step on the personal layer, run once they wrote (or had nothing to
// change) with the person's yes, never in a preview: on Windows with personal_from set, the copy refreshed from it
// (workspace.RefreshPersonal), then the empty index written when the home has none (workspace.EnsurePersonalIndex).
// It never fails the command; it gives the lines for its text, none when there is nothing to say.
func PersonalAfterWrite(home string, clock func() time.Time) []PersonalLine {
	if home == "" {
		return nil
	}
	var out []PersonalLine
	hs, err := workspace.LoadHomeSettings(home)
	switch {
	case err != nil:
		out = append(out, PersonalLine{Text: "the personal memory copy was not refreshed: " + ascii(oneLine(err.Error())), Warn: true})
	case hs.PersonalFrom != "":
		r := workspace.RefreshPersonal(home, hs.PersonalFrom, clock)
		out = append(out, PersonalLine{Text: ascii(r.Line()), Warn: r.Failed != ""})
	}
	index := filepath.ToSlash(filepath.Join(workspace.PersonalPath(home), workspace.PersonalIndex))
	wrote, err := workspace.EnsurePersonalIndex(home, clock())
	switch {
	case err != nil:
		out = append(out, PersonalLine{Text: "the empty personal memory index " + index + " was not written: " + ascii(oneLine(err.Error())), Warn: true})
	case wrote:
		out = append(out, PersonalLine{Text: "wrote the empty personal memory index " + index + " (notes about the person, for every project on this machine; a person writes them)"})
	}
	return out
}
