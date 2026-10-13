package workspace

// This machine's record of a workspace's main checkout, <machine folder>/workspace.json (contract §3: "the main
// checkout's path, and the workspace ids it has held"). bonsai init and update write it, in the main checkout only,
// each time they end without a refusal (written or nothing to change); nothing else writes it, and it never enters a
// project. bonsai check reads it for two things: bonsai.yaml's id no longer the one recorded (the id-changed finding)
// and another checkout on this machine holding the same id (the same-id warning). From step 5.3 the guard trusts a
// main checkout only when this record names it and is the first on this machine to claim its id (design/plan-5.md,
// 5.3.1), which is why each id keeps the time this machine first recorded it. From step 5.4 it also keeps the
// workspace's base branch: the branch the main checkout held when it was recorded, which the ladder runner and the
// stop gate hold the main checkout to (BaseOf), so a main checkout switched to another branch, whose bonsai.yaml may
// say otherwise, proves nothing until it is back on its base or a person records the new one.
//
// The file, a JSON object written byte-stable (two-space indent, LF, ASCII), every field always written:
//
//	{
//	  "path": "/srv/projects/example",       the main checkout's real path, forward slashes (it never leaves the machine)
//	  "ids": [                               the ids the checkout has held here, oldest first; the last is the current one
//	    {"id": "ws-7kq2m4xw5r3t6y2u7p4a5c3e2b", "since": "2026-10-09T12:00:00Z"}
//	  ],
//	  "base_branch": "main",                 the branch the main checkout held when it was recorded; null when it was on
//	                                         a detached HEAD the first time (a record written before step 5.4 has none)
//	  "checks": {...}                        the settings for the checks Bonsai last accepted (checks.go): the floor,
//	                                         the rungs, ratchets, ci_marked_tests, the three path lists, documents;
//	                                         null when none was recorded (a record written before step 5.4's second
//	                                         fix round has none)
//	}
//
// base_branch is written from the main checkout's HEAD file each time the record is written, but for a detached HEAD,
// which keeps the base recorded before. Recording a base other than the one already recorded is a person's step (an
// agent could switch the main checkout and record the new branch, which would make the check moot): bonsai init and
// update refuse it in an agent's session (cmd/bonsai, engine.BaseChange); the first record, and the same branch again,
// are free.
//
// checks is written each time the record is written, from the bonsai.yaml init or update read in full
// (RecordCheckoutWith; RecordCheckout reads the main checkout's file as it is), and brought up by a local climb to the
// settings committed at the main checkout's HEAD when they are stricter (RaiseChecks). A local climb and the stop gate
// refuse committed settings looser than them, and fail closed on a record with none (checks.go says what looser means).
// Recording looser settings than the ones already recorded is a person's step: bonsai init and update refuse it in an
// agent's session (cmd/bonsai, engine.ChecksLooser); the first record, and settings stricter or the same, are free.
//
// Two commands may write the record at once (bonsai update, and a climb bringing its checks up; climbs in one home take
// the ladder's lock first): each write reads the file, decides, stages the new bytes, and renames them into place only
// when the file still holds what it read, else it reads and decides again (rewrite), so neither drops the other's id,
// base or settings but in the moment between that last look and the rename.
//
// since is when this machine first recorded the id for the checkout, UTC, to the second (RFC 3339). An id is added at
// the end when bonsai.yaml's differs from the last one (init --new-id; a person's change of the id, once update runs);
// an entry is never changed or taken out. A key this Bonsai does not know is kept as read. It has no format line and
// no schema in formats/: like the machine folder's settings.json it is Bonsai's own file on one machine, read by
// Bonsai alone.
//
// A record Bonsai cannot read (not a JSON object, a path or ids of the wrong kind, an entry with no id, or a since
// that is not RFC 3339) is damaged, and RecordCheckout does not write over it: a record written again would take a
// new since, and so lose its first claim on the id to any later checkout claiming it (a decoy among them). A person
// checks it and deletes it; the next init or update writes it again, its claim starting then (FirstClaim says which
// checkout claimed an id first; bonsai check's same-id warning names the others).

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/schema"
)

// MachineRecordFile is the machine folder's record of the main checkout.
const MachineRecordFile = "workspace.json"

// MachineRecord is <machine folder>/workspace.json.
type MachineRecord struct {
	Path       string        // the main checkout's real path, forward slashes
	IDs        []RecordedID  // the ids it has held on this machine, oldest first
	BaseBranch string        // the workspace's base branch: the branch the main checkout held when recorded; "" for none
	HasBase    bool          // the record has a base_branch field (a record written before step 5.4 has none)
	Checks     *Checks       // the settings for the checks Bonsai last accepted (checks.go); nil for none
	Extra      schema.Object // keys this Bonsai does not know, kept as read
}

// RecordedID is one id a checkout has held on this machine, and when this machine first recorded it.
type RecordedID struct {
	ID    string
	Since string // RFC 3339, UTC
}

// Current is the id the record holds now (its last), "" for none.
func (m *MachineRecord) Current() string {
	if m == nil || len(m.IDs) == 0 {
		return ""
	}
	return m.IDs[len(m.IDs)-1].ID
}

// readMachineRecord reads a record file; a missing one is nil, no error.
func readMachineRecord(path string) (*MachineRecord, error) {
	m, _, err := readMachineRecordRaw(path)
	return m, err
}

// readMachineRecordRaw reads a record file and gives its bytes too (nil for a missing file, which reads as no record).
func readMachineRecordRaw(path string) (*MachineRecord, []byte, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, &Error{File: filepath.ToSlash(path), Msg: "cannot be read: " + oneLine(err), Err: err, Next: "check the file's permissions"}
	}
	m, err := parseMachineRecord(path, raw)
	return m, raw, err
}

// parseMachineRecord reads a record's bytes (path names the file in its errors).
func parseMachineRecord(path string, raw []byte) (*MachineRecord, error) {
	show := filepath.ToSlash(path)
	next := "a person checks " + show + " (this machine's record of a checkout) and deletes it; bonsai update in that " +
		"checkout then writes it again, its claim on the id starting then"
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	o, isObject := v.(schema.Object)
	if err != nil || !isObject {
		return nil, &Error{File: show, Msg: "is not a JSON object Bonsai reads", Next: next}
	}
	m := &MachineRecord{}
	for _, f := range o {
		switch f.Key {
		case "path":
			s, ok := f.Value.(string)
			if !ok || s == "" {
				return nil, &Error{File: show, Msg: "its path is " + schema.Show(f.Value) + ", not a folder", Next: next}
			}
			m.Path = s
		case "ids":
			list, ok := f.Value.([]any)
			if !ok {
				return nil, &Error{File: show, Msg: "its ids is " + schema.Show(f.Value) + ", not a list", Next: next}
			}
			for _, it := range list {
				io, _ := it.(schema.Object)
				id, since := io.String("id"), io.String("since")
				if id == "" {
					return nil, &Error{File: show, Msg: "an ids entry " + schema.Show(it) + " has no id", Next: next}
				}
				if _, err := time.Parse(time.RFC3339, since); err != nil {
					raw, _ := io.Get("since")
					return nil, &Error{File: show, Msg: "the ids entry of " + showValue(id) + " has the since " + schema.Show(raw) +
						", not a time (RFC 3339)", Next: next}
				}
				m.IDs = append(m.IDs, RecordedID{ID: id, Since: since})
			}
		case "base_branch":
			switch b := f.Value.(type) {
			case nil:
			case string:
				if b == "" {
					return nil, &Error{File: show, Msg: "its base_branch is empty, not a branch's name (or null)", Next: next}
				}
				m.BaseBranch = b
			default:
				return nil, &Error{File: show, Msg: "its base_branch is " + schema.Show(f.Value) + ", not a branch's name (or null)", Next: next}
			}
			m.HasBase = true
		case "checks":
			if f.Value == nil {
				continue
			}
			c, why := readChecks(f.Value)
			if why != "" {
				return nil, &Error{File: show, Msg: "its checks " + why, Next: next}
			}
			m.Checks = c
		default:
			m.Extra = append(m.Extra, f)
		}
	}
	return m, nil
}

// LoadMachineRecord reads this machine's record of the main checkout main: nil when there is none.
func LoadMachineRecord(home, main string) (*MachineRecord, error) {
	dir, err := MachineDir(home, main)
	if err != nil {
		return nil, err
	}
	return readMachineRecord(filepath.Join(dir, MachineRecordFile))
}

// Encode writes the record byte-stable.
func (m *MachineRecord) Encode() ([]byte, error) {
	ids := []any{}
	for _, id := range m.IDs {
		ids = append(ids, schema.Object{{Key: "id", Value: id.ID}, {Key: "since", Value: id.Since}})
	}
	var base any
	if m.BaseBranch != "" {
		base = m.BaseBranch
	}
	var checks any
	if m.Checks != nil {
		checks = m.Checks.JSON()
	}
	o := append(schema.Object{{Key: "path", Value: m.Path}, {Key: "ids", Value: ids}, {Key: "base_branch", Value: base},
		{Key: "checks", Value: checks}}, m.Extra...)
	return schema.Encode(o)
}

// RecordCheckout writes this machine's record of the main checkout main: its real path, id at the end of its ids
// unless it is already the last, since at, the branch main holds now as its base_branch (a detached HEAD keeps the
// base recorded before), and the settings for the checks main's bonsai.yaml holds now, read in full (a file that is
// not there or does not read in full keeps the settings recorded before). It writes nothing when the record already
// says so. It reports whether it wrote. A damaged record (one Bonsai cannot read) is not written over: its error names
// a person's step (this file's comment says why). Whether an agent may record another base, or looser settings, is its
// caller's to refuse (cmd/bonsai, engine.BaseChange and engine.ChecksLooser). bonsai init and update record through
// RecordCheckoutWith, with the settings they read and judged.
func RecordCheckout(home, main, id string, at time.Time) (bool, error) {
	return RecordCheckoutWith(home, main, id, at, checksOfFile(main))
}

// RecordCheckoutWith is RecordCheckout with the settings for the checks to record given (checks.go: ChecksOf of the
// bonsai.yaml the command read); nil keeps the settings recorded before.
func RecordCheckoutWith(home, main, id string, at time.Time, checks *Checks) (bool, error) {
	dir, err := MachineDir(home, main)
	if err != nil {
		return false, err
	}
	real, err := filepath.EvalSymlinks(main)
	if err != nil {
		return false, &Error{File: filepath.ToSlash(main), Msg: "cannot be resolved: " + oneLine(err), Err: err,
			Next: "check that the main checkout exists and can be read"}
	}
	if abs, err := filepath.Abs(real); err == nil {
		real = abs
	}
	path := filepath.ToSlash(real)
	branch := ReadGitLayout(main).Branch()
	return rewrite(filepath.Join(dir, MachineRecordFile), func(m *MachineRecord) (*MachineRecord, error) {
		if m == nil {
			m = &MachineRecord{}
		}
		base, keep := branch, m.Checks
		if base == "" {
			base = m.BaseBranch // a detached HEAD records no base, and keeps the one recorded before
		}
		if checks != nil {
			keep = checks
		}
		if m.Path == path && m.Current() == id && m.HasBase && m.BaseBranch == base && sameChecks(m.Checks, keep) {
			return nil, nil
		}
		m.Path = path
		if m.Current() != id {
			m.IDs = append(m.IDs, RecordedID{ID: id, Since: at.UTC().Format(time.RFC3339)})
		}
		m.BaseBranch, m.HasBase, m.Checks = base, true, keep
		return m, nil
	})
}

// rewriteTries is how many times rewrite reads and decides again when the record changed under it.
const rewriteTries = 4

// rewrite writes the record file key whole: it reads it (nil when there is none), lets change give the record to write
// (nil: nothing to write), stages the new bytes, and renames them into place (Windows' busy retry) only when the file
// still holds the bytes it read; when another command wrote it meanwhile, it reads and asks change again, up to
// rewriteTries times, then writes over it (the last writer wins, as before). A record that cannot be read is not
// written over (this file's comment). It reports whether it wrote.
func rewrite(key string, change func(m *MachineRecord) (*MachineRecord, error)) (bool, error) {
	for try := 1; ; try++ {
		m, before, err := readMachineRecordRaw(key)
		if err != nil {
			var we *Error
			if errors.As(err, &we) && we.Err == nil {
				we.Msg += ", so it is not written over (a record written again would lose its first claim on the id)"
			}
			return false, err
		}
		next, err := change(m)
		if err != nil || next == nil {
			return false, err
		}
		raw, err := next.Encode()
		if err != nil {
			return false, err
		}
		failed := func(err error) error {
			return &Error{File: filepath.ToSlash(key), Msg: "cannot be written: " + oneLine(err), Err: err,
				Next: "check that the Bonsai home can be written"}
		}
		tmp, err := StageFile(key, raw)
		if err != nil {
			return false, failed(err)
		}
		now, err := os.ReadFile(key)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			_ = os.Remove(tmp)
			return false, failed(err)
		}
		if !bytes.Equal(now, before) && try < rewriteTries {
			_ = os.Remove(tmp)
			continue
		}
		if err := CommitStaged(tmp, key); err != nil {
			_ = os.Remove(tmp)
			return false, failed(err)
		}
		return true, nil
	}
}

// AcceptedChecks reads the settings for the checks this machine's record of the main checkout main holds: nil with t
// not verified when there are none to compare with, Why saying why (no record, one that cannot be read, or one written
// before Bonsai kept them) in words that end a sentence, and Next the person's step. It starts no process and writes
// nothing.
func AcceptedChecks(home, main string) (*Checks, Trust) {
	rec, err := LoadMachineRecord(home, main)
	switch {
	case err != nil:
		return nil, notVerified("this machine's record of the main checkout cannot be read", nextOf(err))
	case rec == nil:
		return nil, notVerified("this machine has no record of the main checkout", updateNext)
	case rec.Checks == nil:
		return nil, notVerified("this machine's record of the main checkout holds no settings for the checks to compare "+
			"bonsai.yaml's with (it was written before Bonsai kept them)", "a person runs bonsai update in the main checkout, "+
			"which records bonsai.yaml's settings for the checks as accepted")
	}
	return rec.Checks, Trust{Verified: true}
}

// RaiseChecks brings this machine's record of the main checkout main up to the settings for the checks now (bonsai.yaml
// as committed at main's HEAD; a local climb calls it once it holds the home's ladder lock): when now is stricter than
// the recorded settings, or differs from them in no way that loosens, it writes them and raised is true; when they are
// the same it writes nothing. When now is looser it writes nothing and gives each loosening (Loosenings): the record
// was changed since the climb compared (stopgate.Accepted). A record with no settings, or none, is not raised: err, an
// *Error naming the person's step (only bonsai update records the first settings).
func RaiseChecks(home, main string, now *Checks) (raised bool, looser []string, err error) {
	if now == nil {
		return false, nil, nil
	}
	dir, err := MachineDir(home, main)
	if err != nil {
		return false, nil, err
	}
	key := filepath.Join(dir, MachineRecordFile)
	raised, err = rewrite(key, func(m *MachineRecord) (*MachineRecord, error) {
		looser = nil
		switch {
		case m == nil:
			return nil, &Error{File: filepath.ToSlash(key), Msg: "is not there, so the settings for the checks are not raised",
				Next: updateNext}
		case m.Checks == nil:
			return nil, &Error{File: filepath.ToSlash(key), Msg: "holds no settings for the checks (it was written before Bonsai " +
				"kept them), so they are not raised", Next: "a person runs bonsai update in the main checkout, which records them"}
		}
		if looser = Loosenings(m.Checks, now); len(looser) > 0 || sameChecks(m.Checks, now) {
			return nil, nil
		}
		m.Checks = now
		return m, nil
	})
	return raised, looser, err
}

// KeyedRecord is one workspace's record on this machine, with its machine folder's name.
type KeyedRecord struct {
	Key    string // r-<16 hex>
	Record *MachineRecord
}

// machineFolders lists the home's machine folders (workspaces/r-*, folders only: Bonsai makes no other kind there), by
// name. A home with no workspaces folder has none.
func machineFolders(home string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(home, "workspaces"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "r-") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// MachineRecords reads every workspace's record in the home, by machine folder name; a folder with no record, or one
// Bonsai does not read, is left out.
func MachineRecords(home string) []KeyedRecord {
	names, _ := machineFolders(home)
	var out []KeyedRecord
	for _, n := range names {
		m, err := readMachineRecord(filepath.Join(home, "workspaces", n, MachineRecordFile))
		if err == nil && m != nil {
			out = append(out, KeyedRecord{Key: n, Record: m})
		}
	}
	return out
}

// Claim is one record's claim on a workspace id: its machine folder, the record, and when it first claimed the id
// (the earliest since among its entries holding it).
type Claim struct {
	KeyedRecord
	Since time.Time
}

// ClaimOf gives a record's claim on id, and false when its current id is another (a record that gave the id up no
// longer claims it) or it has none.
func (m *MachineRecord) ClaimOf(id string) (time.Time, bool) {
	if m == nil || id == "" || m.Current() != id {
		return time.Time{}, false
	}
	var first time.Time
	for _, e := range m.IDs {
		if e.ID != id {
			continue
		}
		at, err := time.Parse(time.RFC3339, e.Since)
		if err != nil {
			continue // readMachineRecord refuses such an entry, so none is left here
		}
		if first.IsZero() || at.Before(first) {
			first = at
		}
	}
	return first, !first.IsZero()
}

// FirstClaim says which checkout on this machine first claimed the workspace id id (design/plan-5.md, 5.3.1 note 2:
// the guard trusts a main checkout only when its record is the first on this machine to claim the id, so a checkout
// that claims it later, a decoy among them, is never honoured). Every record in the home is read; among those whose
// current id is id, the one with the earliest claim is first. first is nil when none can be named:
//   - no record claims the id (ties and first are nil);
//   - two or more claimed it in the same second, the earliest (ties holds them, by machine folder): since is kept to
//     the second, so none of them can be told first, and none is honoured (a tie never honours a decoy);
//   - a record in the home cannot be read (err, an *Error naming the file and a person's step): it might hold the
//     earliest claim (a damaged main record while a decoy's reads), so none is honoured until a person mends it.
func FirstClaim(home, id string) (first *Claim, ties []Claim, err error) {
	names, err := machineFolders(home)
	if err != nil {
		return nil, nil, &Error{File: filepath.ToSlash(filepath.Join(home, "workspaces")), Msg: "cannot be read: " + oneLine(err),
			Err: err, Next: "check the folder's permissions"}
	}
	var claims []Claim
	for _, n := range names {
		m, err := readMachineRecord(filepath.Join(home, "workspaces", n, MachineRecordFile))
		if err != nil {
			return nil, nil, err
		}
		if at, ok := m.ClaimOf(id); ok {
			claims = append(claims, Claim{KeyedRecord: KeyedRecord{Key: n, Record: m}, Since: at})
		}
	}
	for _, c := range claims {
		switch {
		case len(ties) == 0 || c.Since.Before(ties[0].Since):
			ties = []Claim{c}
		case c.Since.Equal(ties[0].Since):
			ties = append(ties, c)
		}
	}
	if len(ties) == 1 {
		return &ties[0], nil, nil
	}
	return nil, ties, nil
}

// BaseOf says whether the main checkout main (its top folder) holds the workspace's base branch, as this machine's
// record of it names it (the ladder runner and the stop gate read it, after VerifyMain): base is the recorded branch,
// and Why and Next say why not and the person's step, in words that follow "the main checkout is not on its base
// branch: ". A record written before step 5.4 (no base_branch), or on a detached HEAD the first time (base_branch
// null), fails closed: bonsai update, run by a person in the main checkout on its base branch, records it. It starts no
// process and writes nothing.
func BaseOf(home, main string) (base string, t Trust) {
	rec, err := LoadMachineRecord(home, main)
	switch {
	case err != nil:
		return "", notVerified("this machine's record of the main checkout cannot be read", nextOf(err))
	case rec == nil:
		return "", notVerified("this machine has no record of the main checkout", updateNext)
	case !rec.HasBase || rec.BaseBranch == "":
		why := "this machine's record of the main checkout names no base branch (it was written before Bonsai kept one)"
		if rec.HasBase {
			why = "this machine's record of the main checkout names no base branch (the main checkout was on a detached HEAD " +
				"when it was recorded)"
		}
		return "", notVerified(why, "a person checks out the project's base branch in the main checkout and runs bonsai update "+
			"there, which records it as the base")
	}
	now := ReadGitLayout(main).Branch()
	switch {
	case now == "":
		return rec.BaseBranch, notVerified("the main checkout is on a detached HEAD, and this machine's record names "+
			showValue(rec.BaseBranch)+" as the workspace's base branch", "a person checks out "+showValue(rec.BaseBranch)+
			" in the main checkout (git switch "+rec.BaseBranch+")")
	case now != rec.BaseBranch:
		return rec.BaseBranch, notVerified("the main checkout holds the branch "+showValue(now)+", and this machine's record "+
			"names "+showValue(rec.BaseBranch)+" as the workspace's base branch", "a person switches the main checkout back (git "+
			"switch "+rec.BaseBranch+"), or, when "+showValue(now)+" is the base from now on, runs bonsai update in the main "+
			"checkout to record it")
	}
	return rec.BaseBranch, Trust{Verified: true}
}
