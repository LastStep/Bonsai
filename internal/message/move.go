package message

// The rename that files a message (design/plan-5.md, note 5.5.10, 2): its staged file, created new under a temporary
// name in the recipient's folder, is moved to <id>.md only when no file has that name, so a message is never
// overwritten: on Windows by MoveFileW, which never replaces its target; elsewhere by a hard link, which never
// replaces either, then the staged name removed (where the file system has no hard links, by a rename once no file of
// that name is there). While Windows reports the file busy (an antivirus scan, a search indexer, an editor holding
// it: workspace.IsBusy), the move is tried again for up to moveWait, waiting 5 ms and doubling to 160 ms, as
// internal/workspace's own renames are; any other error returns at once.

import (
	"errors"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// errTaken is a move's answer when its target is there already: the message is not written under that name.
var errTaken = errors.New("a file of that name is there already")

// moveFile, busy, now and sleep are the move and its retry's clock, variables so a test can stand in a busy Windows
// file on any system and count the retry's tries against a clock of its own.
var (
	moveFile = moveNew
	busy     = workspace.IsBusy
	now      = time.Now
	sleep    = time.Sleep
)

// moveWait is how long moveRetry keeps trying while the file is busy (a variable for the tests).
var moveWait = 2 * time.Second

// moveRetry moves from to to, never replacing to, retrying while the error says another process holds the file.
func moveRetry(from, to string) error {
	deadline := now().Add(moveWait)
	delay := 5 * time.Millisecond
	for {
		err := moveFile(from, to)
		if err == nil || errors.Is(err, errTaken) || !busy(err) || now().After(deadline) {
			return err
		}
		sleep(delay)
		if delay < 160*time.Millisecond {
			delay *= 2
		}
	}
}
