package engine

// check's rows for the ladder's shape in bonsai.yaml (plan-5 5.4.2 note 11; 5.4.0 note 2): their words are
// format.LadderCheckWords, their one home, beside CheckWords in the same bonsai.check/1 document. What the runner
// (internal/ladder) reads from a rung, check holds here, so a shape the runner would misread is a finding before any
// climb: two rungs with one number; a floor or a ratchet naming nothing; a capture that is not a mapping of names to
// patterns that compile (Go's RE2); a ratchet with no count (no capture of its name, and no tests form to count passed
// tests from) or one that is not a name; a command that calls bash by name. A tests word Bonsai does not know is a
// warning: the tests are not read. Each is about bonsai.yaml, a person's file, so each next step is a person's edit
// there and then the check again.

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// ladderName is the form of a ratchet's and a capture's name (bonsai.workspace/1's descriptions, as today's ladder).
var ladderName = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^[a-z_][a-z0-9_]{0,39}$`) })

// againCheck ends each ladder row's next step.
const againCheck = "; then run: bonsai check --json"

// checkLadder checks the ladder's shape in this checkout's bonsai.yaml (this file's comment).
func (r *CheckResult) checkLadder() {
	if r.Config == nil || r.Config.Full == nil {
		return
	}
	w := r.Config.Full
	file := workspace.ConfigFile
	byNumber := map[int64][]string{}
	var numbers []int64
	named := map[string]bool{} // the ratchets the rungs name
	for _, g := range w.Ladder {
		label := rungLabel(g)
		if _, ok := byNumber[g.Rung]; !ok {
			numbers = append(numbers, g.Rung)
		}
		byNumber[g.Rung] = append(byNumber[g.Rung], label)
		if g.Command != nil && format.CallsBash(*g.Command) {
			r.add("rung-bash", file, "", label+"'s command calls bash by name (spec section 3: on Windows bash is often WSL's launcher, "+
				"not Git Bash), so the rung would run another shell there", "a person writes the command without bash in bonsai.yaml's "+
				"ladder (the rung runs in sh on Linux and in cmd on Windows)"+againCheck)
		}
		caps := r.checkCapture(g, label)
		if g.Ratchet != nil {
			name, isText := g.Ratchet.(string)
			switch {
			case !isText || !ladderName().MatchString(name):
				r.add("ratchet-no-count", file, "", label+"'s ratchet is "+schema.Show(g.Ratchet)+", not a ratchet's name (a lower-case letter "+
					"or _, then up to 39 lower-case letters, digits or _), so no count can be read for it", "a person writes the ratchet's "+
					"name in bonsai.yaml's ladder (the name of a floor in ratchets)"+againCheck)
			case !caps[name] && !knownTests(g.Tests):
				r.add("ratchet-no-count", file, "", label+"'s ratchet "+name+" has no count: the rung has no capture named "+name+
					" and no tests form to count passed tests from, so a climb reads nothing for it and it is red", "a person gives "+
					label+" a capture named "+name+" (a pattern whose first group is the count) or a tests form ("+formWords()+") in "+
					"bonsai.yaml's ladder"+againCheck)
			default:
				named[name] = true
			}
		}
		if g.Tests != nil && !knownTests(g.Tests) {
			r.add("tests-unknown", file, "", label+"'s tests is "+strconv.Quote(*g.Tests)+", a form Bonsai does not know, so its tests "+
				"are not read (the known forms: "+formWords()+")", "a person writes one of "+formWords()+" as the rung's tests in "+
				"bonsai.yaml, or null; to see each form: run: bonsai check --schema bonsai.workspace")
		}
	}
	for _, n := range numbers {
		if labels := byNumber[n]; len(labels) > 1 {
			r.add("rung-twice", file, "", "bonsai.yaml's ladder has "+strconv.Itoa(len(labels))+" rungs numbered "+strconv.FormatInt(n, 10)+
				" ("+strings.Join(labels, ", ")+"), so the floor and a task's bonsai.ladder cannot name one of them alone",
				"a person gives each of those rungs its own number in bonsai.yaml's ladder"+againCheck)
		}
	}
	for _, n := range w.LadderFloor {
		if _, ok := byNumber[n]; !ok {
			r.add("floor-no-rung", file, "", "bonsai.yaml's ladder_floor names rung "+strconv.FormatInt(n, 10)+", which its ladder does "+
				"not have, so every local climb refuses (rung-not-defined)", "a person adds rung "+strconv.FormatInt(n, 10)+
				" to bonsai.yaml's ladder, or takes it out of ladder_floor"+againCheck)
		}
	}
	var keys []string
	for _, m := range w.Ratchets {
		keys = append(keys, m.Key)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !named[k] {
			r.add("ratchet-no-rung", file, "", "bonsai.yaml's ratchets holds a floor for "+strconv.Quote(k)+", which no rung's ratchet "+
				"names, so it holds back nothing", "a person names "+k+" as the ratchet of the rung it counts in bonsai.yaml's ladder, "+
				"or takes the floor out of ratchets"+againCheck)
		}
	}
}

// checkCapture holds a rung's capture to its form (a mapping from names to patterns that compile) and gives the names
// it captures.
func (r *CheckResult) checkCapture(g format.Rung, label string) map[string]bool {
	names := map[string]bool{}
	if g.Capture == nil {
		return names
	}
	file := workspace.ConfigFile
	next := "a person fixes " + label + "'s capture in bonsai.yaml: a mapping from a name to a Go (RE2) pattern whose first group is the value" +
		againCheck
	m, ok := g.Capture.(schema.Object)
	if !ok {
		r.add("capture-pattern", file, "", label+"'s capture is "+schema.Show(g.Capture)+", not a mapping from names to patterns", next)
		return names
	}
	for _, kv := range m {
		pat, isText := kv.Value.(string)
		switch {
		case !ladderName().MatchString(kv.Key):
			r.add("capture-pattern", file, "", label+"'s capture names "+strconv.Quote(kv.Key)+", not a capture's name (a lower-case "+
				"letter or _, then up to 39 lower-case letters, digits or _)", next)
		case !isText:
			r.add("capture-pattern", file, "", label+"'s capture "+kv.Key+" is "+schema.Show(kv.Value)+", not a pattern (text)", next)
		default:
			if _, err := regexp.Compile(pat); err != nil {
				r.add("capture-pattern", file, "", label+"'s capture "+kv.Key+" holds the pattern "+strconv.Quote(pat)+", which does not "+
					"compile as a Go (RE2) regular expression: "+ascii(err.Error()), next)
				continue
			}
			names[kv.Key] = true
		}
	}
	return names
}

// rungLabel names a rung in a finding: "rung 3 (unit tests)".
func rungLabel(g format.Rung) string {
	s := "rung " + strconv.FormatInt(g.Rung, 10)
	if g.Name != nil && *g.Name != "" {
		s += " (" + ascii(*g.Name) + ")"
	}
	return s
}

// knownTests reports a tests form Bonsai reads.
func knownTests(form *string) bool {
	if form == nil {
		return false
	}
	for _, w := range format.TestForms {
		if w.Word == *form {
			return true
		}
	}
	return false
}

// formWords lists the known tests forms (format.TestForms, their one home).
func formWords() string {
	var out []string
	for _, w := range format.TestForms {
		out = append(out, w.Word)
	}
	return strings.Join(out, ", ")
}
