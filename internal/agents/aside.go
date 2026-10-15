package agents

// An agent's state that cannot be read (VG's item I, 15 Oct: such an agent was left uncounted, every later call
// "not counted"): the guard moves it aside and counts the agent afresh from its next call, so a broken file never lets
// an agent run past its budget unseen, and what was there is kept for the person to read.

import (
	"errors"
	"io/fs"
	"os"
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
