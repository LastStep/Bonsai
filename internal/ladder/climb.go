package ladder

// The climb's state: what the steps before a rung found (ladder.go), which kinds.go's functions read, and what the
// rungs give (rungs.go) on the way to the result (result.go).

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Climb is one climb of the ladder. Its exported fields are what a rung kind and a step after a rung read
// (kinds.go): set before the first rung, never changed by a kind.
type Climb struct {
	Root      string                // the checkout climbed: its top folder, absolute, system separators
	Main      string                // the verified main checkout, whose .bonsai/local/ holds the result (under --ci with main not verified, Root)
	Verified  bool                  // Main is verified (workspace.VerifyMain); false only under --ci
	Home      string                // the Bonsai home
	Config    *workspace.Config     // Main's bonsai.yaml, read in full (Config.Full): the ladder, the floor, the ratchets, the task folder
	Task      *workspace.ActiveTask // the task climbed for, named (contract §13); nil for none (--ci with none)
	CI        bool                  // --ci: mode ci
	Requested []int64               // the rungs asked for, sorted, each once
	Base      string                // the base branch: the branch Main has checked out; "" when its HEAD is detached
	MergeBase string                // git merge-base of Base and HEAD in Root; "" when there is no base or git finds none
	Prior     *format.Ladder        // the task's earlier result in Main's ladder folder (made for this task and workspace, its bytes the ones its last ladder record fingerprints), or nil
	Getenv    func(string) string   // the climb's environment reader
	Now       func() time.Time      // the climb's clock

	o          Options
	started    time.Time            // when the first rung's climb began, once the lock was taken
	workspace  *format.WorkspaceRef // the climb document's workspace
	unverified string               // under --ci, why Main is not verified ("" when it is)
	rungs      []format.Rung        // the rungs climbed, in number order (bonsai.yaml's order within one number)
	git        format.LadderGit     // HEAD at the start, its branch, dirty
	headAfter  string               // HEAD at the end
	dirtyAfter bool                 // git's status at the end
	hides      []string             // the rungs whose findings hide a change from git (kinds.go: Outcome.Hides)
	results    []format.LadderRung  // one per rung climbed
	steps      []*Step              // one per rung climbed, as the steps after a rung saw it
	skipped    []format.LadderSkip  // --ci: every skip, accepted or not
	leftovers  any                  // the leftovers record (leftovers.Value.Record)
	leftLines  []string             // the leftovers line
	stopped    bool                 // the stop signal came
	priorNote  string               // why an earlier result that is there was not read ("" when it was, or there is none)
}

// TaskID is the climbed task's id, "" for none.
func (c *Climb) TaskID() string {
	if c.Task == nil {
		return ""
	}
	return c.Task.ID
}

// PriorCaptures is the earlier result's captures for rung n (5.4.6 compares its stash count and exclude hash with
// them), or nil when there is no earlier result or it holds none for the rung.
func (c *Climb) PriorCaptures(n int64) schema.Object {
	if c.Prior == nil {
		return nil
	}
	for _, r := range c.Prior.Rungs {
		if r.Rung == n {
			return r.Captures
		}
	}
	return nil
}

// ResultRel is the result's project-relative path: .bonsai/local/ladder/<task>.json, or ci.json for a --ci climb.
func (c *Climb) ResultRel() string {
	name := c.TaskID()
	if c.CI {
		name = "ci"
	}
	return workspace.LocalDir + "/ladder/" + name + ".json"
}

// mode is the result's mode.
func (c *Climb) mode() string {
	if c.CI {
		return ModeCI
	}
	return ModeLocal
}

// prior reads the task's earlier local result in Main's ladder folder: nil when there is none, it does not read, it
// was made for another task or workspace, or its bytes are not the ones the log's last ladder record for the task
// fingerprints (lastFingerprint). The earlier result is what git integrity compares its stash count, exclude hash and
// base with, and what the rise step takes a recorded base from: a file put there by hand, or an older result copied
// back, would steer both, so only the result the last climb wrote is read. note says why a result that is there is not
// read, for the human lines ("" when it is read, or there is none).
func (c *Climb) prior() (l *format.Ladder, note string) {
	if c.Task == nil {
		return nil, ""
	}
	rel := workspace.LocalDir + "/ladder/" + c.Task.ID + ".json"
	raw, err := os.ReadFile(filepath.Join(c.Main, filepath.FromSlash(rel)))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ""
		}
		return nil, "the task's earlier result " + rel + " cannot be read (" + oneLine(err.Error()) + "), so this climb compares with no earlier one"
	}
	l, err = format.ReadLadder(raw)
	if err != nil || l.Task == nil || *l.Task != c.Task.ID || l.Workspace != c.Config.ID || l.Mode != ModeLocal {
		return nil, "the task's earlier result " + rel + " is not a local result of this task, so this climb compares with no earlier one"
	}
	sum := sha256.Sum256(raw)
	want, found, err := lastFingerprint(filepath.Join(c.Main, filepath.FromSlash(workspace.LocalDir), record.LogFolder), c.Config.ID,
		c.Task.ID, rel)
	switch {
	case err != nil:
		return nil, "the log cannot be read (" + oneLine(err.Error()) + "), so the task's earlier result " + rel + " cannot be checked " +
			"against its ladder record and this climb compares with no earlier one"
	case !found:
		return nil, "the log holds no ladder record for the task's earlier result " + rel + ", so it is not read: this climb compares " +
			"with no earlier one"
	case !want["sha256:"+hex.EncodeToString(sum[:])]:
		return nil, "the task's earlier result " + rel + " is not the one its last ladder record in the log fingerprints (another file " +
			"was put there), so it is not read: this climb compares with no earlier one"
	}
	return l, ""
}

// lastFingerprint finds the task's last ladder record in the log folder dir (every session and day file, record.ListLog):
// the record of event ladder for this workspace and task whose target is the result's path rel and whose kind is
// local, with the latest at. It gives the fingerprints (text, sha256:<hex>) of the records with that at (one, but for
// two climbs in one millisecond), and whether one was found. A line that does not read as a record is passed over, as
// every log reader does; only lines naming the result's path are read whole, so a long log costs a scan, not a parse.
func lastFingerprint(dir, workspaceID, task, rel string) (map[string]bool, bool, error) {
	files, err := record.ListLog(dir)
	if err != nil {
		return nil, false, err
	}
	names := append(append([]string{}, files.Sessions...), files.Days...)
	target := []byte(`"target":"` + rel + `"`)
	event := []byte(`"event":"ladder"`)
	var last string
	texts := map[string]bool{}
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue // cleaned while the folder was read
			}
			return nil, false, err
		}
		for _, line := range bytes.Split(raw, []byte("\n")) {
			if !bytes.Contains(line, target) || !bytes.Contains(line, event) {
				continue
			}
			l, err := format.ReadLog(bytes.TrimSuffix(line, []byte("\r")))
			if err != nil || l.Event != "ladder" || l.Workspace != workspaceID || l.Task == nil || *l.Task != task ||
				l.Target == nil || *l.Target != rel || l.Kind == nil || *l.Kind != ModeLocal || l.Text == nil {
				continue
			}
			switch {
			case l.At > last:
				last, texts = l.At, map[string]bool{*l.Text: true}
			case l.At == last:
				texts[*l.Text] = true
			}
		}
	}
	return texts, last != "", nil
}

func sortInts(list []int64) { sort.Slice(list, func(i, j int) bool { return list[i] < list[j] }) }

// sortRungs puts rungs in number order, bonsai.yaml's order kept within one number (two rungs with one number is a
// check finding, rung-twice; both are climbed).
func sortRungs(list []format.Rung) {
	sort.SliceStable(list, func(i, j int) bool { return list[i].Rung < list[j].Rung })
}
