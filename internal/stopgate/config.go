package stopgate

// The main checkout's bonsai.yaml as committed (V2's M2): the ladder runner and the stop gate read the ladder, the
// floor, the ratchets and ci_marked_tests from the bonsai.yaml committed at the main checkout's HEAD, git's object, and
// refuse while the file in the main checkout differs from it. The file itself is a person's (person_only), but a shell
// write in an agent's session is not judged by the guard in agents mode, so an uncommitted edit there could lower a
// floor or drop a rung for one climb and one gate; a committed one is history a person reads.

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/LastStep/Bonsai/internal/vergit"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// UncommittedError is CommittedConfig's refusal: the main checkout's bonsai.yaml differs from the one committed at its
// HEAD, or its HEAD holds none. Why says which, in words that end a sentence.
type UncommittedError struct{ Why string }

func (e *UncommittedError) Error() string { return e.Why }

// UncommittedNext is the person's step after an UncommittedError, before the command that confirms it.
const UncommittedNext = "a person commits the change to bonsai.yaml in the main checkout (git commit), or reverts it " +
	"(git checkout HEAD -- bonsai.yaml)"

// CommittedConfig reads the main checkout's bonsai.yaml as committed at its HEAD (vergit.Committed: one git call, git's
// object) and in full (workspace.ReadConfigFull), and refuses (*UncommittedError) while the file in the main checkout
// differs from it, its line ends read as LF (a CRLF checkout is the same file), or is not there, or HEAD holds none
// (no commit yet, or none holding it).
// main is the main checkout's top folder; timeout limits the git call. Any other error says the file could not be
// read: git could not run, or the committed file is not one Bonsai reads (a *workspace.Error naming its line).
func CommittedConfig(main string, timeout time.Duration) (*workspace.Config, error) {
	committed, err := vergit.Committed(main, workspace.ConfigFile, timeout)
	switch {
	case errors.Is(err, vergit.ErrNoCommit):
		return nil, &UncommittedError{Why: "the main checkout's HEAD names no commit yet, so it holds no bonsai.yaml"}
	case errors.Is(err, vergit.ErrNotInHead):
		return nil, &UncommittedError{Why: "the main checkout's HEAD holds no bonsai.yaml (it was never committed there)"}
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
