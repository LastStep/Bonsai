package message

// bonsai.yaml and the lock as committed (VM's D5): bonsai message and bonsai messages, and status's roles and messages,
// read them as committed at the main checkout's HEAD, as the guard does, so a change counts once committed. This is
// the one reading for both callers (cmd/bonsai's message.go, internal/status): the caller's working copy stands in only
// where HEAD holds no such file (held false: no commit yet, never committed there, or a folder there); a committed
// copy that does not read, or a HEAD git cannot read, is an error naming the file and its next step (bonsai message
// refuses, status lists a problem), never quietly replaced by the working copy. RolesInForce is the roles in force from
// them, its error naming bonsai.yaml or the lock, whichever does not read, with its line.

import (
	"errors"
	"strconv"

	"github.com/LastStep/Bonsai/internal/vergit"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// CommittedConfig reads bonsai.yaml as committed at main's HEAD, in full: held false, and no error, where HEAD holds
// none. A committed copy that does not read is bad-config, naming its line; a HEAD git cannot read, read-failed.
func CommittedConfig(main string) (cfg *workspace.Config, held bool, e *Error) {
	raw, held, e := committed(main, workspace.ConfigFile)
	if e != nil || !held {
		return nil, held, e
	}
	c, err := workspace.ReadConfigFull(raw)
	if err != nil {
		next := "fix bonsai.yaml where it says and commit it (bonsai check names what does not read)"
		var we *workspace.Error
		if errors.As(err, &we) && we.Next != "" {
			next = ascii(we.Next) + ", in bonsai.yaml, and commit it"
		}
		return nil, true, &Error{Code: "bad-config", Exit: ExitInput,
			What: "bonsai.yaml as committed at the main checkout's HEAD does not read: " + fileWhat(err),
			Next: next + " (it counts here as committed)"}
	}
	return c, true, nil
}

// CommittedLock reads the lock as committed at main's HEAD: held false, and no error, where HEAD holds none. A
// committed lock that does not read is bad-lock; a HEAD git cannot read, read-failed.
func CommittedLock(main string) (lock *workspace.Lock, held bool, e *Error) {
	raw, held, e := committed(main, workspace.LockFile)
	if e != nil || !held {
		return nil, held, e
	}
	l, err := workspace.ReadLock(raw)
	if err != nil {
		return nil, true, &Error{Code: "bad-lock", Exit: ExitState,
			What: "the lock as committed at the main checkout's HEAD does not read: " + fileWhat(err),
			Next: "the lock is Bonsai's to write: a person runs bonsai update in the main checkout and commits the lock it " +
				"writes; if update refuses it, the person deletes it and links the project again from bonsai.yaml (bonsai " +
				"init --yes previews every file first), then commits it"}
	}
	return l, true, nil
}

// committed reads rel as committed at main's HEAD (vergit.Committed: git's object, never the working file): held
// false where HEAD holds none (no commit yet, no such path, a folder there); an error when git cannot read what HEAD
// names or cannot run.
func committed(main, rel string) ([]byte, bool, *Error) {
	raw, err := vergit.Committed(main, rel, 0)
	switch {
	case err == nil:
		return raw, true, nil
	case errors.Is(err, vergit.ErrNoCommit), errors.Is(err, vergit.ErrNotInHead), errors.Is(err, vergit.ErrNotAFile):
		return nil, false, nil
	case errors.Is(err, vergit.ErrHeadUnreadable):
		return nil, false, &Error{Code: "read-failed", Exit: ExitRuntime, Who: "person",
			What: rel + " cannot be read as committed at the main checkout's HEAD: " + oneLine(err),
			Next: "a person runs git fsck in the main checkout, which names each object that is not there or is damaged, " +
				"then fetches them (git fetch) or makes a fresh clone of the project"}
	}
	return nil, false, &Error{Code: "read-failed", Exit: ExitRuntime,
		What: rel + " cannot be read as committed at the main checkout's HEAD: " + oneLine(err),
		Next: "check that git runs in the main checkout (git status there), then run the command again"}
}

// fileWhat is a file's read error as one sentence without its next step: the file, its line and what is wrong.
func fileWhat(err error) string {
	var we *workspace.Error
	if !errors.As(err, &we) {
		return oneLine(err)
	}
	s := we.File
	if we.Line > 0 {
		s += " line " + strconv.Itoa(we.Line)
	}
	s += ": " + we.Msg
	if we.Code != "" {
		s += " (" + we.Code + ")"
	}
	return ascii(s)
}

// RolesInForce gives the roles in force from the lock and bonsai.yaml a caller read (workspace.RolesInForceOf, the
// one function every reader of a role's limits uses), its error naming the file that does not read: bonsai.yaml with
// its line (bad-config, a step for whoever may change bonsai.yaml, then commit it), or the lock (bad-lock, a person's
// update).
func RolesInForce(lock *workspace.Lock, cfg *workspace.Config) ([]workspace.RoleInForce, *Error) {
	in, err := workspace.RolesInForceOf(lock, cfg)
	if err == nil {
		return in, nil
	}
	var re *workspace.RolesError
	switch {
	case errors.As(err, &re) && re.File == workspace.ConfigFile:
		return nil, &Error{Code: "bad-config", Exit: ExitInput, Who: "agent", What: ascii(re.Error()),
			Next: ascii(re.Next()) + ", and commit it (the roles count as committed at the main checkout's HEAD)"}
	case errors.As(err, &re):
		return nil, &Error{Code: "bad-lock", Exit: ExitState, What: ascii(re.Error()), Next: ascii(re.Next())}
	}
	return nil, unexpected(err)
}
