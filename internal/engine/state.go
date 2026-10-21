package engine

// STATE at init (spec §4: "STATE.md from base's template when there is none, kind once"; contract §7.2; plan-5 5.5.0
// note 2), and its template's one home: the engine writes it, since a project may link without base, and base's state
// skill is generated whole from it and from bonsai.state/1's schema (skills.go, go generate), never copied by hand.
//
// init writes .bonsai/STATE.md when none is there: format 1 (format: bonsai.state/1 with its pointer comment, the one
// line of docs a file made from a template carries, contract §2.8), updated today, updated_by bonsai init, labels {},
// and a body of a few headings for the orchestrator to fill; kind once in the lock, so from then on it is the
// project's. Neither init nor update ever writes over one that is there (one found at init is locked as found);
// update leaves a missing one missing (spec §6's table: a once file deleted is left missing), so a project linked
// before step 5.5 adds one from base's state skill; unlink leaves it (step 5.1.7). The preview names it.

import (
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// StatePointer is the comment STATE's format: line carries: where its fields are documented and how to fill it.
const StatePointer = "fields: bonsai check --schema bonsai.state; how to fill: skill base:state"

// StateFields are the template's docs for each field of bonsai.state/1, in its schema's order: what the field means,
// for the state skill's fields table (skills.go). Their one home; TestStateTemplate holds them to the schema's fields
// both ways.
var StateFields = []struct{ Field, Meaning string }{
	{"format", "The format and its version, always the first line; its comment points back to these docs"},
	{"updated", "The day the page was last rewritten, as text: rewrite it with the page"},
	{"updated_by", "Who rewrote it: a role and its model, a person's name, or bonsai init for the first page"},
	{"labels", "The labels the page carries, by name (a label definition in force says which go on STATE); {} for none"},
}

// stateBody is the body init writes under STATE's frontmatter: a few headings, no paths.
const stateBody = `
# Where this project stands

Rewritten whole whenever where the project stands changes; never appended to. A new session reads it first.

## Now

## Next

## Open questions

## Decided lately
`

// StateTemplate is .bonsai/STATE.md as init writes it, updated on the given day (YYYY-MM-DD).
func StateTemplate(day string) []byte {
	return []byte("---\nformat: bonsai.state/1   # " + StatePointer + "\nupdated: " + day + "\nupdated_by: bonsai init\nlabels: {}\n---\n" +
		stateBody)
}

// today is the day init writes into a new STATE's updated: the UTC date, YYYY-MM-DD (Bonsai records in UTC). Tests set it.
var today = func() string { return stateClock().UTC().Format("2006-01-02") }

// stateClock is the clock today reads; a test replaces it.
var stateClock = time.Now

// stateFile is STATE in a plan (note 2): written at init when it is missing, found when init meets one, kept as the
// lock holds it, left missing at update; nil when update meets a project whose lock has no STATE (linked before step
// 5.5): nothing is written or locked then. pack names the lock entry's pack, as the block's and the settings' do (the
// first pack in bonsai.yaml).
func stateFile(command, root string, lock, newLock *workspace.Lock, pack string) (*FileResult, error) {
	old, exists, err := readFile(root, workspace.StateFile)
	if err != nil {
		return nil, err
	}
	lf, locked := lock.Files[workspace.StateFile]
	locked = locked && lf.Kind == "once"
	f := &FileResult{Path: workspace.StateFile, Kind: "once", Pack: pack, old: old}
	keep := func() {
		e := lf
		e.Pack = pack
		newLock.Files[workspace.StateFile] = e
	}
	switch {
	case command == "init" && !exists:
		b := StateTemplate(today())
		f.Result, f.Why, f.write, f.newH = Created, "where the project stands, from Bonsai's template: written once, then the project's", b,
			workspace.HashLF(b)
		newLock.Files[workspace.StateFile] = workspace.LockedFile{Kind: "once", Pack: pack, SHA256: f.newH, Extra: lf.Extra}
	case locked && !exists:
		f.Result, f.Why = LeftMissing, "the project's STATE, deleted: left missing (bonsai init writes a new one)"
		keep()
	case locked:
		f.Result, f.Why = Unchanged, "the project's STATE"
		keep()
	case command == "init":
		f.Result, f.Why = Found, "already in the project: left as it is, and the project's"
		newLock.Files[workspace.StateFile] = workspace.LockedFile{Kind: "once", Pack: pack, SHA256: workspace.HashLF(old), Extra: lf.Extra}
	default:
		return nil, nil
	}
	return f, nil
}
