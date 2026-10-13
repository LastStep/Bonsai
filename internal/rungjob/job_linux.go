//go:build linux

package rungjob

// Linux: the rung's session and group, its tag, and the runner as the reaper of what escapes (the package doc).

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/LastStep/Bonsai/internal/leftovers"
)

// prSetChildSubreaper is prctl's PR_SET_CHILD_SUBREAPER.
const prSetChildSubreaper = 36

var subreaper struct {
	once sync.Once
	err  error
}

// becomeSubreaper marks this process a child subreaper, once: a process of a rung whose parent exits is then
// handed to this process, not to init. The mark is this process's alone; its children do not inherit it.
var becomeSubreaper = func() error {
	subreaper.once.Do(func() {
		if _, _, e := syscall.RawSyscall(syscall.SYS_PRCTL, prSetChildSubreaper, 1, 0); e != 0 {
			subreaper.err = e
		}
	})
	return subreaper.err
}

// shells are the rung shells this package started and os/exec has not yet reaped, by pid: never another rung's
// adopted child, and never reaped here.
var shells sync.Map

func isShell(pid int) bool {
	_, ok := shells.Load(pid)
	return ok
}

// running are the tags of the rungs this process is running right now, from prepare to finish: an orphan handed to
// the runner that carries one of them is that rung's (two rungs run side by side in one process, as the tests' parallel
// Run calls do), and no other rung adopts it. A tag of no rung running here (another runner's, or one a rung's own
// command wrote, such as BONSAI_RUNG_TAG=other) is no rung's claim on it.
var running sync.Map

// runningHere reports whether a tag names a rung this process is running now.
func runningHere(tag string) bool {
	_, ok := running.Load(tag)
	return ok
}

// rung is one rung's job on Linux.
type rung struct {
	tag         string
	self        int
	selfSID     int
	leader      int    // the shell's pid: its session and group are numbered by it
	leaderStart uint64 // the shell's start time, which tells its pid from a later process's
	seen        map[int]uint64
	setup       time.Duration
	notes       []string
}

func newRung() *rung { return &rung{self: os.Getpid(), seen: map[int]uint64{}} }

func (r *rung) job() string { return "on" }

func (r *rung) setupCost() time.Duration { return r.setup }

// prepare marks the runner a subreaper and tags the rung's environment.
func (r *rung) prepare(env []string) []string {
	t := time.Now()
	if err := becomeSubreaper(); err != nil {
		r.notes = append(r.notes, "reaper: off: prctl(PR_SET_CHILD_SUBREAPER): "+err.Error())
	}
	if st, ok := leftovers.ReadStat(r.self); ok {
		r.selfSID = st.SID
	}
	r.tag = randomTag()
	running.Store(r.tag, true)
	tag := r.tag
	if outer := getEnv(env, TagEnv); outer != "" {
		tag = outer + ":" + r.tag
	}
	env = setEnv(env, TagEnv, tag)
	r.setup += time.Since(t)
	return env
}

func (r *rung) command(command, dir string, env []string) *exec.Cmd {
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Pdeathsig: syscall.SIGKILL}
	return cmd
}

// launch starts the shell on a locked thread and waits for it there: Pdeathsig fires when the thread that started
// the child exits, and a thread is only ever ended by the runtime when a goroutine locked to it returns locked, so
// the thread is held, locked, until the shell is reaped.
func (r *rung) launch(cmd *exec.Cmd) (<-chan error, error) {
	started := make(chan error, 1)
	exited := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if err := cmd.Start(); err != nil {
			started <- err
			return
		}
		pid := cmd.Process.Pid
		shells.Store(pid, true)
		r.leader = pid
		if st, ok := leftovers.ReadStat(pid); ok {
			r.leaderStart = st.Start
		}
		started <- nil
		err := cmd.Wait()
		shells.Delete(pid)
		exited <- err
	}()
	if err := <-started; err != nil {
		return nil, err
	}
	return exited, nil
}

// end lists the rung's processes, ends its group and each of them, and lists again until none is left (EndWait at
// most). alive says the shell may still run (a timeout or a stop).
func (r *rung) end(cmd *exec.Cmd, alive bool) ([]Process, []string) {
	var notes []string
	ended := []Process{}
	deadline := time.Now().Add(EndWait)
	for round := 0; ; round++ {
		rows, err := r.members()
		if err != nil {
			notes = append(notes, "could not list the rung's processes: "+err.Error())
			if alive && round == 0 {
				// The shell is not reaped yet, so its group's number is still its own.
				_ = syscall.Kill(-r.leader, syscall.SIGKILL)
			}
			break
		}
		for _, s := range rows {
			if start, ok := r.seen[s.PID]; ok && start == s.Start {
				continue
			}
			r.seen[s.PID] = s.Start
			ended = append(ended, Process{PID: s.PID, Name: s.Name, Head: leftovers.HeadOf(leftovers.Argv(s.PID))})
		}
		for _, s := range rows {
			if s.PGID == r.leader {
				// A group is acted on only while it has a member: its number cannot have passed to another.
				_ = syscall.Kill(-r.leader, syscall.SIGKILL)
				break
			}
		}
		killRows(rows)
		r.reap()
		if len(rows) == 0 {
			break
		}
		if time.Now().After(deadline) {
			notes = append(notes, fmt.Sprintf("%d of the rung's processes still alive %s s after its end", len(rows),
				seconds(EndWait)))
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return ended, notes
}

// members lists the rung's live processes: its session and group, every process carrying its tag, the runner's
// adopted children, and every live descendant of those; never the runner or its ancestors.
func (r *rung) members() ([]leftovers.Stat, error) {
	stats, err := leftovers.Stats()
	if err != nil {
		return nil, err
	}
	byPID := make(map[int]leftovers.Stat, len(stats))
	for _, s := range stats {
		byPID[s.PID] = s
	}
	spare := map[int]bool{}
	for x, n := r.self, 0; n < 256; n++ {
		s, ok := byPID[x]
		if !ok || spare[x] {
			break
		}
		spare[x] = true
		x = s.PPID
	}
	// The leader's number counts while no other process holds it: a live process with that pid and another start
	// time took the number after the shell was reaped and every member had gone.
	lead := r.leader
	if h, ok := byPID[r.leader]; ok && h.Start != r.leaderStart {
		lead = 0
	}
	found := map[int]bool{}
	var stack []int
	for _, s := range stats {
		if s.Dead() || spare[s.PID] {
			continue
		}
		if lead != 0 && (s.SID == lead || s.PGID == lead) || r.tagged(s.PID) ||
			s.PPID == r.self && r.adopted(s) {
			found[s.PID] = true
			stack = append(stack, s.PID)
		}
	}
	children := map[int][]int{}
	for _, s := range stats {
		if p, ok := byPID[s.PPID]; ok && !s.Dead() && !spare[s.PID] && p.Start <= s.Start {
			children[s.PPID] = append(children[s.PPID], s.PID)
		}
	}
	for len(stack) > 0 {
		x := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, c := range children[x] {
			if !found[c] {
				found[c] = true
				stack = append(stack, c)
			}
		}
	}
	var rows []leftovers.Stat
	for _, s := range stats {
		if found[s.PID] {
			rows = append(rows, s)
		}
	}
	return rows, nil
}

// tagged reports a process whose environment carries this rung's tag.
func (r *rung) tagged(pid int) bool {
	v, ok := leftovers.EnvValue(pid, TagEnv)
	if !ok {
		return false
	}
	for _, t := range strings.Split(v, ":") {
		if t == r.tag {
			return true
		}
	}
	return false
}

// adopted reports a child of the runner that this rung's processes left to it: in a session other than the
// runner's (no process can join the runner's session from the rung's), started after the rung's shell, not a
// shell this package started, and carrying no tag of another rung this process is running now (running). A tag of
// no rung running here does not keep it: a rung's command can set the variable itself (BONSAI_RUNG_TAG=other setsid
// ...), and an orphan only this runner was handed is this runner's to end. The caller has checked it is the runner's
// child.
func (r *rung) adopted(s leftovers.Stat) bool {
	if s.SID == r.selfSID || s.Start < r.leaderStart || isShell(s.PID) {
		return false
	}
	v, ok := leftovers.EnvValue(s.PID, TagEnv)
	if !ok || v == "" || r.tagged(s.PID) {
		return true
	}
	for _, t := range strings.Split(v, ":") {
		if t != r.tag && runningHere(t) {
			return false // another rung's, running beside this one
		}
	}
	return true
}

// killRows ends each listed process whose pid still has the start time it was listed with.
func killRows(rows []leftovers.Stat) {
	for _, s := range rows {
		now, ok := leftovers.ReadStat(s.PID)
		if !ok || now.Start != s.Start || now.Dead() {
			continue
		}
		_ = syscall.Kill(s.PID, syscall.SIGKILL)
	}
}

// reap waits for the runner's ended children that belong to this rung, each by its own pid and without blocking:
// never wait4(-1), which would take a child os/exec is waiting for.
func (r *rung) reap() {
	stats, err := leftovers.Stats()
	if err != nil {
		return
	}
	for _, s := range stats {
		if s.PPID != r.self || !s.Dead() || isShell(s.PID) {
			continue
		}
		start, seen := r.seen[s.PID]
		mine := seen && start == s.Start ||
			r.leader != 0 && (s.SID == r.leader || s.PGID == r.leader) ||
			s.SID != r.selfSID && s.Start >= r.leaderStart
		if !mine {
			continue
		}
		var ws syscall.WaitStatus
		for {
			_, err := syscall.Wait4(s.PID, &ws, syscall.WNOHANG, nil)
			if !errors.Is(err, syscall.EINTR) {
				break
			}
		}
	}
}

// finish reaps what the end left to the runner and gives the notes gathered on the way. The rung's tag is no longer
// one this process runs.
func (r *rung) finish() []string {
	r.reap()
	running.Delete(r.tag)
	return r.notes
}

// randomTag is a fresh tag: 16 random bytes in hex.
func randomTag() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand does not fail on the systems Bonsai runs on; the time keeps a tag distinct if it ever did.
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b)
}

// signalNumber gives the signal that ended a process, or 0.
func signalNumber(ps *os.ProcessState) int {
	if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return int(ws.Signal())
	}
	return 0
}
