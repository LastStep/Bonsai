package ladder

// One ladder at a time per home (spec §9, §10: the home's locks/ladder.lock; plan-5 5.4.2 note 8). The lock is the
// system's, ended with its holder: flock on Linux and macOS (lock_unix.go); on Windows LockFileEx on one byte far past
// the file's end (lock_windows.go), so a waiting climb can still read the file's words, which a file opened with no
// sharing would not allow. Its holder writes in the file who it is: its task, its checkout, its process id, when it
// started and its budget (the sum of its rungs' timeouts), one JSON object.
//
// A second climb tries the lock; held, it reads the holder's words, says once on its notice writer whom it waits for,
// and tries again every LockPoll while the holder's process lives and the holder is inside its budget. Past the
// budget it refuses: ladder-busy, exit 4, naming the holder and the next step (the same command once that process has
// ended). A holder that died lets go of the lock with its process, so the lock is taken at once. A held lock whose
// words do not read (a holder that has just taken it, or words left by an earlier holder whose process is gone) is
// waited for LockGrace, then refused the same way. Two homes (WSL's and Windows', or a test's) hold two locks and never
// block each other. The lock file's handle is never inherited by a rung (Go opens files close-on-exec, and Windows
// handles are not inherited unless asked), so no process a rung leaves behind can hold it.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/schema"
)

// LockPoll is how often a waiting climb tries the lock again.
const LockPoll = time.Second

// LockGrace is how long a held lock whose holder has not named itself is waited for.
const LockGrace = 10 * time.Second

// LockFile is the lock's place in the home.
const LockFile = "locks/ladder.lock"

// holder is the lock holder's words in the file.
type holder struct {
	Task     string `json:"task"`     // the task it climbs for, "" for none (--ci)
	Checkout string `json:"checkout"` // the checkout it climbs, forward slashes
	PID      int    `json:"pid"`
	Started  string `json:"started"`  // when it took the lock, UTC
	BudgetS  int64  `json:"budget_s"` // the sum of its rungs' timeouts, in seconds
}

// until is when the holder's budget ends; the zero time when its words do not say.
func (h holder) until() time.Time {
	t, err := time.Parse(time.RFC3339Nano, h.Started)
	if err != nil {
		return time.Time{}
	}
	return t.Add(time.Duration(h.BudgetS) * time.Second)
}

// who names the holder in a message.
func (h holder) who() string {
	task := h.Task
	if task == "" {
		task = "no task (--ci)"
	}
	return fmt.Sprintf("process %d (the climb of %s in %s)", h.PID, task, engine.ASCII(h.Checkout))
}

// encode writes the holder's words, byte-stable.
func (h holder) encode() []byte {
	b, _ := schema.EncodeLine(schema.Object{{Key: "task", Value: h.Task}, {Key: "checkout", Value: h.Checkout},
		{Key: "pid", Value: int64(h.PID)}, {Key: "started", Value: h.Started}, {Key: "budget_s", Value: h.BudgetS}})
	return b
}

// readHolder reads the holder's words from the lock file; ok is false when they do not read.
func readHolder(path string) (holder, bool) {
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) == 0 || len(raw) > 64<<10 {
		return holder{}, false
	}
	var h holder
	if err := json.Unmarshal(raw, &h); err != nil || h.PID <= 0 || h.Started == "" {
		return holder{}, false
	}
	return h, true
}

// budget is the sum of the climb's rungs' timeouts: each rung's timeout_s, else the rung job's default.
func (c *Climb) budget() int64 {
	var s int64
	for _, r := range c.rungs {
		s += int64(rungTimeout(r) / time.Second)
	}
	return s
}

// takeLock takes the home's ladder lock for the climb (this file's comment) and gives its release. A nested climb takes
// none.
func takeLock(c *Climb) (release func(), e *engine.Error) {
	if c.o.nested() {
		return func() {}, nil
	}
	path := filepath.Join(c.Home, filepath.FromSlash(LockFile))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, lockError(c, path, err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, lockError(c, path, err)
	}
	me := holder{Task: c.TaskID(), Checkout: filepath.ToSlash(c.Root), PID: os.Getpid(), BudgetS: c.budget()}
	var (
		told         bool
		unnamedSince time.Time
	)
	for {
		ok, err := tryLock(f)
		if err != nil {
			_ = f.Close()
			return nil, lockError(c, path, err)
		}
		if ok {
			me.Started = time.Now().UTC().Format(time.RFC3339Nano)
			if err := writeHolder(f, me); err != nil {
				unlock(f)
				_ = f.Close()
				return nil, lockError(c, path, err)
			}
			return func() {
				_ = f.Truncate(0)
				unlock(f)
				_ = f.Close()
			}, nil
		}
		now := time.Now()
		h, named := readHolder(path)
		if named && alive(h.PID) {
			unnamedSince = time.Time{}
			end := h.until()
			if !end.IsZero() && now.After(end) {
				_ = f.Close()
				return nil, busy(c, h, "and is past its budget of "+strconv.FormatInt(h.BudgetS, 10)+" s, which ended at "+
					end.UTC().Format(time.RFC3339))
			}
			if !told {
				_, _ = fmt.Fprintf(c.o.Notice, "bonsai ladder: another climb holds this home's ladder lock: %s; waiting while it runs, "+
					"until %s at the latest\n", h.who(), end.UTC().Format(time.RFC3339))
				told = true
			}
		} else {
			if unnamedSince.IsZero() {
				unnamedSince = now
			}
			if now.Sub(unnamedSince) > c.o.LockGrace {
				_ = f.Close()
				what := "a process that has not written its words in " + filepath.ToSlash(LockFile)
				if named {
					what = "a process whose words name process " + strconv.Itoa(h.PID) + ", which has ended"
				}
				return nil, &engine.Error{Code: "ladder-busy", Exit: ExitState,
					What: "this home's ladder lock is held by " + what + " (waited " + c.o.LockGrace.String() + "), so nothing was run",
					Next: "wait for the other climb on this machine to end; then run: bonsai ladder " + c.taskFlag() + " --json"}
			}
		}
		select {
		case <-c.o.Stop.Done():
			_ = f.Close()
			return nil, &engine.Error{Code: "ladder-busy", Exit: ExitState, What: "the climb was stopped while it waited for this home's " +
				"ladder lock, so nothing was run", Next: "run: bonsai ladder " + c.taskFlag() + " --json"}
		case <-time.After(c.o.LockPoll):
		}
	}
}

// writeHolder writes the holder's words over the file's.
func writeHolder(f *os.File, h holder) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.WriteAt(h.encode(), 0); err != nil {
		return err
	}
	return f.Sync()
}

// busy is the refusal of a climb whose holder keeps the lock past its budget.
func busy(c *Climb, h holder, why string) *engine.Error {
	return &engine.Error{Code: "ladder-busy", Exit: ExitState,
		What: "this home's ladder lock is held by " + h.who() + ", " + why + ", so nothing was run",
		Next: "wait until process " + strconv.Itoa(h.PID) + " has ended (its budget ended at " + h.until().UTC().Format(time.RFC3339) +
			"; it may be stuck, and a person may end it); then run: bonsai ladder " + c.taskFlag() + " --json"}
}

// lockError is a lock file that cannot be made or locked.
func lockError(c *Climb, path string, err error) *engine.Error {
	return &engine.Error{Code: "write-failed", Exit: ExitRuntime,
		What: "the ladder lock " + shown(path) + " cannot be made or taken (" + oneLine(err.Error()) + "), so nothing was run",
		Next: "check that the Bonsai home can be written; then run: bonsai ladder " + c.taskFlag() + " --json"}
}
