package stopgate

// The main checkout's bonsai.yaml as committed (V2's M2): the ladder runner and the stop gate read the ladder, the
// floor, the ratchets and ci_marked_tests from the bonsai.yaml committed at the main checkout's HEAD, git's object, and
// refuse while the file in the main checkout differs from it. The file itself is a person's (person_only), but a shell
// write in an agent's session is not judged by the guard in agents mode, so an uncommitted edit there could lower a
// floor or drop a rung for one climb and one gate.
//
// And no looser than this machine accepted (5.4's second fix round, K1; the director, 13 Oct): a commit on the base branch is
// history, which an agent could make too (a shell commit in the main checkout, git reset --hard to a looser commit, a
// branch ref moved by hand), so the settings committed there are compared with the ones this machine's record of the
// main checkout last accepted (Accepted; workspace/checks.go says what is compared and what reads as looser), and
// refused when they are looser, each loosening named, or when the record holds none: a person's step, bonsai update in
// the main checkout, records them. Settings stricter or the same are accepted; a local climb then brings the record up
// to them (workspace.RaiseChecks, under the home's ladder lock), and the stop gate, a hook, only compares: every allow
// it gives rests on a green result of a climb, which compared and recorded the settings it ran by, and a stricter
// commit after that climb holds the gate by its own reading of the committed floor.

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/vergit"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// UncommittedError is CommittedConfig's refusal: the main checkout's bonsai.yaml differs from the one committed at its
// HEAD, or its HEAD holds none. Why says which, in words that end a sentence.
// Never says HEAD holds no bonsai.yaml at all (no commit yet, or none holding it): what bonsai init wrote waits for its
// first commit, which bonsai check's config-uncommitted warning names as the step (BUG-106).
type UncommittedError struct {
	Why   string
	Never bool
}

func (e *UncommittedError) Error() string { return e.Why }

// UncommittedNext is the person's step after an UncommittedError, before the command that confirms it.
const UncommittedNext = "a person commits the change to bonsai.yaml in the main checkout (git commit), or reverts it " +
	"(git checkout HEAD -- bonsai.yaml)"

// HeadUnreadableError is CommittedConfig's refusal when git cannot read bonsai.yaml at the main checkout's HEAD though
// HEAD names a commit (V2b's S6): an object HEAD needs, its commit, a tree or the file, is not in the repository or is
// damaged (a clone whose borrowed objects are gone, a fetch cut short, a fault on the disk). There is no line of
// bonsai.yaml to fix, and nothing to commit: git fsck names what is wrong. Why says what, in words that end a sentence.
type HeadUnreadableError struct{ Why string }

func (e *HeadUnreadableError) Error() string { return e.Why }

// HeadUnreadableNext is the person's step after a HeadUnreadableError, before the command that confirms it.
const HeadUnreadableNext = "a person runs git fsck in the main checkout, which names each object that is not there or is " +
	"damaged, then fetches them (git fetch) or makes a fresh clone of the project"

// CommittedConfig reads the main checkout's bonsai.yaml as committed at its HEAD (vergit.Committed: one git call, git's
// object) and in full (workspace.ReadConfigFull), and refuses (*UncommittedError) while the file in the main checkout
// differs from it, its line ends read as LF (a CRLF checkout is the same file), or is not there, or HEAD holds none
// (no commit yet, none holding it, or a folder there); and (*HeadUnreadableError) when HEAD names a commit git cannot
// read, or cannot read the file in (vergit.ErrHeadUnreadable), its id read from git's files (Head) when they name one.
// main is the main checkout's top folder; timeout limits the git call. Any other error says the file could not be
// read: git could not run, or the committed file is not one Bonsai reads (a *workspace.Error naming its line).
func CommittedConfig(main string, timeout time.Duration) (*workspace.Config, error) {
	committed, err := vergit.Committed(main, workspace.ConfigFile, timeout)
	if errors.Is(err, vergit.ErrHeadUnreadable) {
		// HEAD names a commit, but that commit or an object under it is not there or is damaged.
		named := "the main checkout's HEAD names a commit"
		if sha, herr := Head(workspace.ReadGitLayout(main)); herr == nil {
			named = "the main checkout's HEAD names commit " + short(sha) + ","
		}
		why := named + " which git cannot read (the commit is not in the repository, or is damaged), so bonsai.yaml cannot be " +
			"read as committed there"
		if !errors.Is(err, vergit.ErrCommitUnreadable) {
			why = named + " but git cannot read bonsai.yaml there (" + oneLine(err.Error()) + "): an object HEAD needs is not in " +
				"the repository or is damaged"
		}
		return nil, &HeadUnreadableError{Why: why}
	}
	switch {
	case errors.Is(err, vergit.ErrNoCommit):
		return nil, &UncommittedError{Why: "the main checkout's HEAD names no commit yet, so it holds no bonsai.yaml", Never: true}
	case errors.Is(err, vergit.ErrNotInHead):
		return nil, &UncommittedError{Why: "the main checkout's HEAD holds no bonsai.yaml (it was never committed there)", Never: true}
	case errors.Is(err, vergit.ErrNotAFile):
		return nil, &UncommittedError{Why: "the main checkout's HEAD holds bonsai.yaml as a folder or a submodule, not a file"}
	case err != nil:
		return nil, fmt.Errorf("bonsai.yaml cannot be read at the main checkout's HEAD: %v", err)
	}
	working, err := os.ReadFile(filepath.Join(main, workspace.ConfigFile))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, &UncommittedError{Why: "the main checkout's bonsai.yaml is gone, though its HEAD holds one"}
	case err != nil:
		return nil, &workspace.Error{File: workspace.ConfigFile, Msg: "cannot be read in the main checkout", Err: err,
			Next: "check the file's permissions"}
	case !bytes.Equal(lf(working), lf(committed)):
		return nil, &UncommittedError{Why: "the main checkout's bonsai.yaml differs from the one committed at its HEAD"}
	}
	return workspace.ReadConfigFull(committed)
}

// lf reads a file's line ends as LF.
func lf(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }

// LooserError is Accepted's refusal: the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser
// than the settings this machine last accepted for the workspace, or there are none to compare with. Why says which,
// in words that end a sentence; Next is the person's step, before the command that confirms it.
type LooserError struct{ Why, Next string }

func (e *LooserError) Error() string { return e.Why }

// LooserNext is the person's step after settings looser than the accepted ones.
const LooserNext = "a person reads the change to bonsai.yaml on the base branch (git log -p -- bonsai.yaml in the main " +
	"checkout): to keep the looser settings, a person runs bonsai update in the main checkout, in a terminal of their own " +
	"(it records them as accepted); else a commit there puts bonsai.yaml's settings back"

// maxLoosenings caps the loosenings a refusal names; the rest are counted.
const maxLoosenings = 8

// Accepted compares the settings for the checks in cfg (the main checkout's bonsai.yaml as committed at its HEAD, read
// in full: CommittedConfig) with the ones this machine's record of the main checkout main accepted
// (workspace.AcceptedChecks), and refuses (*LooserError) when they are looser (workspace.Loosenings), naming each, or
// when the record holds none (written before Bonsai kept them: it fails closed). home is the Bonsai home. nil means
// they are the same or stricter. It starts no process and writes nothing.
func Accepted(home, main string, cfg *workspace.Config) *LooserError {
	was, t := workspace.AcceptedChecks(home, main)
	if !t.Verified {
		return &LooserError{Why: t.Why, Next: t.Next}
	}
	var now *workspace.Checks
	if cfg != nil {
		now = workspace.ChecksOf(cfg.Full)
	}
	if now == nil {
		return &LooserError{Why: "bonsai.yaml was not read in full, so its settings for the checks cannot be compared",
			Next: "tell the person: Bonsai could not compare bonsai.yaml's settings"}
	}
	l := workspace.Loosenings(was, now)
	if len(l) == 0 {
		return nil
	}
	if len(l) > maxLoosenings {
		l = append(l[:maxLoosenings:maxLoosenings], "and "+strconv.Itoa(len(l)-maxLoosenings)+" more")
	}
	return &LooserError{Why: "the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser than the settings " +
		"this machine last accepted for the workspace: " + strings.Join(l, "; "), Next: LooserNext}
}
