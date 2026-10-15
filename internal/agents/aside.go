package agents

// An agent's state that cannot be read (VG's item I, 15 Oct: such an agent was left uncounted, every later call
// "not counted"): the guard moves it aside and counts the agent afresh from its next call, so a broken file never lets
// an agent run past its budget unseen, and what was there is kept for the person to read.
//
// First, though, a state or calls file whose mode was changed (VG round 2's M1, 15 Oct: chmod a-w on an agent's calls
// file left it uncounted; chmod 000 on its state moved it aside, a fresh window each time) has its owner's access set
// back (Repair), so the count goes on where it was: Bonsai's own files are its own to read and write. What Repair
// cannot set right (an ACL, an immutable flag, a folder or FIFO in a file's place) is moved aside or refused by the
// guard.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// asideMark marks a state moved aside (MoveAside) in its file's name.
const asideMark = "-aside-"

// MoveAside moves an agent's state that cannot be read, and its calls file with it, aside in its session's folder:
// <agent>-aside-<time>.state and .calls (the time UTC to the millisecond; the agent's id cut so the name stays one a
// state file takes), renamed whatever they are (a FIFO or a folder is never opened), so they are kept and listed
// (bonsai agents names the state in skipped; the cleaning takes both with their session's folder), and the agent's
// next state is made afresh. It gives the aside state's path inside .bonsai/local/, forward slashes; "" with no error
// when the state reads now (another call side by side made it afresh first), so nothing is moved.
func (s Store) MoveAside(session, agent string, now time.Time) (string, error) {
	if err := validName(session); err != nil {
		return "", err
	}
	if err := validName(agent); err != nil {
		return "", err
	}
	if _, err := s.Read(session, agent); err == nil || errors.Is(err, ErrNoState) {
		return "", nil
	}
	base := agent
	stamp := now.UTC().Format("20060102T150405.000Z")
	stamp = stamp[:15] + stamp[16:] // no dot: a state file's name holds letters, digits, _ and - only
	if keep := 128 - len(asideMark) - len(stamp); len(base) > keep {
		base = base[:keep]
	}
	aside := base + asideMark + stamp
	from, to := s.statePath(session, agent), s.statePath(session, aside)
	if _, err := os.Lstat(to); err == nil {
		return "", errors.New("a state moved aside at the same time is already there")
	}
	if err := workspace.CommitStaged(from, to); err != nil {
		return "", err
	}
	if _, err := os.Lstat(s.callsPath(session, agent)); err == nil {
		if err := workspace.CommitStaged(s.callsPath(session, agent), s.callsPath(session, aside)); err != nil &&
			!errors.Is(err, fs.ErrNotExist) {
			return rel(session, aside+stateExt), err
		}
	}
	return rel(session, aside+stateExt), nil
}

// Repair sets back the owner's access to an agent's own files where their mode was taken away: read, write and search
// on the agents folder, the session's folder and its claims folder (and .bonsai/local/ and .bonsai/ above them), read
// and write on the agent's state and calls files. Each is looked at first and changed only when it is what it should
// be (a plain folder, a regular file: a link is never followed) and lacks one of those bits; nothing else is touched,
// and no one else's bits are added. On Windows a file's read-only attribute is cleared (os.Chmod's meaning there). It
// gives what it changed, inside .bonsai/local/, forward slashes; none when nothing needed it, or nothing could be.
func (s Store) Repair(session, agent string) []string {
	if validName(session) != nil || validName(agent) != nil {
		return nil
	}
	local := filepath.Join(s.Main, filepath.FromSlash(workspace.LocalDir))
	var changed []string
	mend := func(path string, dir bool) {
		fi, err := os.Lstat(path)
		if err != nil || (dir && !fi.IsDir()) || (!dir && !fi.Mode().IsRegular()) {
			return
		}
		need := fs.FileMode(0o600)
		if dir {
			need = 0o700
		}
		if fi.Mode().Perm()&need == need {
			return
		}
		if os.Chmod(path, fi.Mode().Perm()|need) == nil {
			name := "."
			if rel, err := filepath.Rel(local, path); err == nil {
				name = filepath.ToSlash(rel)
			}
			changed = append(changed, name)
		}
	}
	for _, d := range []string{filepath.Dir(local), local, s.Dir(), s.SessionDir(session), filepath.Join(s.SessionDir(session), ClaimsFolder)} {
		mend(d, true)
	}
	mend(s.statePath(session, agent), false)
	mend(s.callsPath(session, agent), false)
	return changed
}

// Asides counts an agent's states moved aside in its session's folder (MoveAside's names, the agent's id cut as there).
func (s Store) Asides(session, agent string) int {
	if validName(session) != nil || validName(agent) != nil {
		return 0
	}
	entries, err := os.ReadDir(s.SessionDir(session))
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), stateExt)
		if !ok {
			continue
		}
		base, _, found := strings.Cut(name, asideMark)
		if found && base != "" && strings.HasPrefix(agent, base) && (base == agent || len(name) == 128) {
			n++
		}
	}
	return n
}
