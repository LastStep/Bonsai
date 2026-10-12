// Package rungjob runs one ladder rung's command so that everything it starts ends with it (spec §9; plan 5.4.1),
// and reports its exit, its output and what it ended. The runner (internal/ladder) calls Run once per command rung.
//
//	r := rungjob.Run(ctx, rungjob.Spec{Command: "go test -json ./...", Dir: root, Env: env, Timeout: 10 * time.Minute})
//	green := r.Green()
//	reason := r.Reason // "timed out after 600 s; ended 3 processes", "job: off: <why>", or ""
//
// Linux. The command runs as /bin/sh -c in a new session and process group (Setsid), with Pdeathsig, so the shell
// dies with the runner: the start runs on a locked operating-system thread that stays locked until the shell is
// reaped, since Pdeathsig fires when the thread that started the child exits. Its environment carries a random tag,
// BONSAI_RUNG_TAG, after any outer rung's (a:b), the way Jenkins tags a build. The runner marks itself a child
// subreaper (prctl(PR_SET_CHILD_SUBREAPER) through package syscall), so a process of the rung whose parent exits is
// handed to the runner, not to init. At the rung's exit, its timeout or a stop, the runner lists from /proc the
// rung's session and group, every process whose environment carries the tag, its own adopted children (a child in
// a session other than the runner's, started after the rung's shell, that is not a shell this package started and
// carries no other rung's tag) and every live descendant of those; it ends the group, then each process, and lists
// again until none is left (EndWait at most). A pid is acted on only while its start time is the one listed; a
// group only while it has a member, so its number cannot have passed to another process. Adopted orphans are reaped
// by their own pids, never by wait4(-1), which would race os/exec's own wait.
//
// Why a process group alone was not enough (the spec's words): a child that calls setsid or setpgid leaves the
// group, as today's ladder found; a daemon that double-forks leaves its parent. The session, the tag and the reaper
// each catch what the others miss: setsid leaves the session but keeps the tag; a process that clears its
// environment and leaves the session is still the runner's child once its parent is gone. What lost against
// today's ladder: its keeper, so a runner killed by SIGKILL leaves what escaped its shell (its shell dies by
// Pdeathsig), which the next run's leftovers line names (internal/leftovers).
//
// Windows. A job object made by the runner, with kill-on-close and no breakaway; its handle is not inherited, so
// the runner's handle is the only one, and a runner that dies (taskkill /f) closes it and the job ends the tree.
// The command, %ComSpec% /d /s /c "<command>" (today's line), is started suspended, put in the job, then let run
// (NtResumeProcess, on a handle opened by its pid, since os/exec closes the main thread's handle), so no process of
// the rung starts outside it. A child asking to break away is refused by the system; a job made inside the rung's
// is nested in it. At the end, its timeout or a stop, the runner lists the job's members (pids, image names and
// command heads), ends the job, and waits until it is empty (EndWait at most). If the job cannot be made, or the
// command cannot be put in it, the rung runs without one and says so (Job and Reason: "job: off: <why>"), never
// silently. The calls go through package syscall's lazily loaded procedures, so golang.org/x/sys stays out of the
// program (step 5.3's decision). No PowerShell launcher.
//
// Elsewhere (macOS and the BSDs build, untested) the command runs in its own process group, which is ended at the
// rung's end; what it ended is not listed.
//
// Everything else. The rung's environment gains BONSAI_LADDER=1, so a nested runner knows it is nested and the
// schema compare fails rather than skips. Its output is kept whole up to OutputCap (standard output also apart,
// for the test forms), plus its last TailLines lines for the reason. The output pipes are the runner's own, so
// cmd.Wait returns at the shell's exit and the tree is ended before the pipes are drained; a process the end could
// not reach that still holds them is cut off PipeCutoff after the command (os/exec's WaitDelay plays that part when
// os/exec copies the output itself, which here would hold Wait until the pipes close). A timeout ends the rung red
// with "timed out after N s; ended M processes". A failure to list or end anything never changes the colour; it
// is a note.
//
// Known limits. Linux: a process the rung had started for it by a program outside the rung (a service manager, a
// daemon it talks to) is not the rung's descendant and carries no tag; a process that clears its environment,
// leaves the session and whose parent is gone cannot be told from a stranger without root once the runner is not
// its reaper (the runner killed, or prctl refused, which a note says); and an adopted child of that kind cannot be
// told between two rungs of one runner running at once, so a runner runs its rungs one at a time and starts no
// child of its own in a new session beside a running rung. Windows: with the job off, only the shell is ended at a
// timeout, and what it started is not listed.
package rungjob

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/leftovers"
)

// DefaultTimeout is a rung's time limit when its Spec names none: 600 s, as today.
const DefaultTimeout = 600 * time.Second

// OutputCap is how much of each output stream a Result keeps whole, in bytes: 32 MiB. Bonsai's own go test -json
// of every package is a few MB; the cap keeps a runaway rung from holding the runner's memory. What is past the cap
// is counted (OutSize, StdoutSize), not kept; the tail is kept whatever the size.
const OutputCap = 32 << 20

// TailLines is how many of the output's last lines a Result keeps for the reason.
const TailLines = 40

// tailBytes bounds the tail: the last 40 lines, within the output's last 16 KiB.
const tailBytes = 16 << 10

// PipeCutoff is how long after the command's end the runner waits for its output pipes to close before it cuts
// them: a process the end could not reach (the job off, an escape) still holds them.
const PipeCutoff = 10 * time.Second

// EndWait is how long the end lists and ends again until none of the rung's processes is left.
const EndWait = 2 * time.Second

// The variables a rung's environment gains.
const (
	// LadderEnv is set to 1 in every rung's environment: a runner that sees it is nested (it prints no leftovers
	// line), and the schema compare fails rather than skips (step 5.4.0).
	LadderEnv = "BONSAI_LADDER"
	// TagEnv tags a rung's processes on Linux: the tags of every rung around a process, ':'-separated.
	TagEnv = "BONSAI_RUNG_TAG"
)

// Spec is one rung's command and how to run it.
type Spec struct {
	Command string        // the command line, run by /bin/sh -c (Linux) or %ComSpec% /d /s /c (Windows)
	Dir     string        // the folder it runs in; "" for the runner's own
	Env     []string      // its environment, NAME=value; nil for the runner's own. LadderEnv (and TagEnv) are added
	Timeout time.Duration // its limit; 0 for DefaultTimeout
}

// Process is one process the rung ended: its pid, program name and command head.
type Process = leftovers.Process

// Result is how a rung's command ended.
type Result struct {
	Exit       int       // the command's exit code; -1 when it has none (timed out, stopped, or not started)
	Out        []byte    // standard output and standard error as they came, the first OutputCap bytes
	Stdout     []byte    // standard output alone, the first OutputCap bytes (what the test forms read)
	OutSize    int64     // the whole output's size in bytes; past OutputCap, Out holds only its start
	StdoutSize int64     // the same for standard output
	Tail       []string  // the output's last lines (TailLines at most), blank lines left out
	TimedOut   bool      // the rung's limit ended it
	Stopped    bool      // the caller's stop (the context) ended it
	Ended      []Process // the rung's processes still running at its end, all ended, by pid; nil when not known
	Job        string    // "on", or "off: <why>" when the rung ran without its job (Windows) or group
	Reason     string    // the rung job's own words when there is something to say, else ""
	Notes      []string  // what the end could not do (list, end, a pipe cut): never a colour
	Duration   time.Duration
	Setup      time.Duration // the job's (Windows) or the tag's and reaper's (Linux) cost before the command ran
	End        time.Duration // the end's cost: listing, ending and waiting until none is left
}

// Green reports whether the command ran, ended by itself and exited 0.
func (r Result) Green() bool { return r.Exit == 0 && !r.TimedOut && !r.Stopped }

// Run runs one rung's command and returns how it ended. It never fails: a command that cannot start is a Result
// with Exit -1 and its reason. ctx is the stop signal: when it is done, the rung is ended as at a timeout.
func Run(ctx context.Context, s Spec) Result {
	start := time.Now()
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	env := s.Env
	if env == nil {
		env = os.Environ()
	}
	env = setEnv(env, LadderEnv, "1")
	res := Result{Exit: -1, Job: "on"}
	if strings.TrimSpace(s.Command) == "" {
		return finishNotStarted(res, "the command is empty", start)
	}

	out := &capture{}
	outR, outW, err := os.Pipe()
	if err != nil {
		return finishNotStarted(res, "no pipe for its output: "+err.Error(), start)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		_ = outR.Close()
		_ = outW.Close()
		return finishNotStarted(res, "no pipe for its output: "+err.Error(), start)
	}
	var readers sync.WaitGroup
	readers.Add(2)
	go out.read(outR, true, &readers)
	go out.read(errR, false, &readers)
	closeReads := func() { _ = outR.Close(); _ = errR.Close() }

	r := newRung()
	env = r.prepare(env)
	cmd := r.command(s.Command, s.Dir, env)
	cmd.Stdin = nil
	cmd.Stdout = outW
	cmd.Stderr = errW
	exited, err := r.launch(cmd)
	_ = outW.Close()
	_ = errW.Close()
	if err != nil {
		readers.Wait()
		closeReads()
		r.finish()
		res.Job = r.job()
		res.Setup = r.setupCost()
		return finishNotStarted(res, err.Error(), start)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	var waitErr error
	done := false
	select {
	case waitErr = <-exited:
		done = true
	case <-timer.C:
		res.TimedOut = true
	case <-ctx.Done():
		res.Stopped = true
	}
	endStart := time.Now()
	ended, notes := r.end(cmd, !done)
	if !done {
		select {
		case waitErr = <-exited:
			done = true
		case <-time.After(EndWait + 3*time.Second):
			notes = append(notes, "the rung's shell did not exit after its end")
		}
	}
	res.End = time.Since(endStart)
	commandEnd := time.Now()

	// The pipes: closed once nothing of the rung holds them, which after the end is at once.
	pipes := make(chan struct{})
	go func() { readers.Wait(); close(pipes) }()
	select {
	case <-pipes:
	case <-time.After(time.Until(commandEnd.Add(PipeCutoff))):
		out.freeze()
		notes = append(notes, fmt.Sprintf("output cut %d s after the command ended: a process outside the rung's "+
			"job holds its pipes", int(PipeCutoff/time.Second)))
	}
	closeReads()
	notes = append(notes, r.finish()...)

	if done && !res.TimedOut && !res.Stopped {
		res.Exit = exitCode(cmd, waitErr)
	}
	if ended != nil {
		sort.Slice(ended, func(i, j int) bool { return ended[i].PID < ended[j].PID })
	}
	res.Ended = ended
	res.Notes = notes
	res.Job = r.job()
	res.Setup = r.setupCost()
	res.Out, res.Stdout, res.OutSize, res.StdoutSize, res.Tail = out.result()
	var parts []string
	switch {
	case res.TimedOut:
		parts = append(parts, fmt.Sprintf("timed out after %s s; ended %d processes", seconds(timeout), len(ended)))
	case res.Stopped:
		parts = append(parts, fmt.Sprintf("stopped; ended %d processes", len(ended)))
	}
	if res.Job != "on" {
		parts = append(parts, "job: "+res.Job)
	}
	res.Reason = strings.Join(parts, "; ")
	res.Duration = time.Since(start)
	return res
}

// finishNotStarted is the Result of a command that could not start.
func finishNotStarted(res Result, why string, start time.Time) Result {
	res.Exit = -1
	res.Reason = "could not start: " + oneLine(why)
	if res.Job != "on" {
		res.Reason += "; job: " + res.Job
	}
	res.Duration = time.Since(start)
	return res
}

// exitCode reads the command's exit code; a shell ended by a signal reads as 128 plus its number, as shells say.
func exitCode(cmd *exec.Cmd, waitErr error) int {
	if cmd.ProcessState == nil {
		return -1
	}
	if c := cmd.ProcessState.ExitCode(); c >= 0 {
		return c
	}
	if n := signalNumber(cmd.ProcessState); n > 0 {
		return 128 + n
	}
	_ = waitErr
	return -1
}

// seconds writes a duration in seconds, whole when it is whole: 600, 1.5.
func seconds(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', -1, 64)
}

// oneLine keeps a message's first line, its paths with forward slashes.
func oneLine(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	s = strings.TrimSpace(s)
	if runtime.GOOS == "windows" {
		s = strings.ReplaceAll(s, `\`, "/")
	}
	return s
}

// setEnv sets one variable in an environment list, dropping any earlier entry of it (in any case on Windows).
func setEnv(env []string, key, value string) []string {
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		k, _, _ := strings.Cut(kv, "=")
		if sameKey(k, key) {
			continue
		}
		out = append(out, kv)
	}
	return append(out, key+"="+value)
}

// getEnv reads one variable from an environment list (the last entry wins, as os/exec reads it).
func getEnv(env []string, key string) string {
	v := ""
	for _, kv := range env {
		k, val, ok := strings.Cut(kv, "=")
		if ok && sameKey(k, key) {
			v = val
		}
	}
	return v
}

func sameKey(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// capture keeps a command's output: both streams as they came and standard output apart, each to OutputCap, and
// the last tailBytes of both for the tail.
type capture struct {
	mu         sync.Mutex
	out, std   []byte
	outN, stdN int64
	tail       []byte
	frozen     bool
}

// read copies one stream into the capture until it closes.
func (c *capture) read(r io.Reader, stdout bool, wg *sync.WaitGroup) {
	defer wg.Done()
	buf := make([]byte, 32<<10)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			c.write(buf[:n], stdout)
		}
		if err != nil {
			return
		}
	}
}

func (c *capture) write(b []byte, stdout bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.frozen {
		return
	}
	c.outN += int64(len(b))
	c.out = appendCapped(c.out, b)
	if stdout {
		c.stdN += int64(len(b))
		c.std = appendCapped(c.std, b)
	}
	c.tail = append(c.tail, b...)
	if len(c.tail) > 2*tailBytes {
		c.tail = append(c.tail[:0:0], c.tail[len(c.tail)-tailBytes:]...)
	}
}

// freeze stops the capture: what a reader still blocked on a held pipe reads later is not kept.
func (c *capture) freeze() {
	c.mu.Lock()
	c.frozen = true
	c.mu.Unlock()
}

func appendCapped(dst, b []byte) []byte {
	if room := OutputCap - len(dst); room < len(b) {
		b = b[:max(room, 0)]
	}
	return append(dst, b...)
}

// result gives the kept output and its tail.
func (c *capture) result() (out, stdout []byte, outN, stdN int64, tail []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.frozen = true
	t := c.tail
	cut := false
	if len(t) > tailBytes {
		t = t[len(t)-tailBytes:]
		cut = true
	}
	lines := strings.Split(strings.ToValidUTF8(string(t), "?"), "\n")
	if cut && len(lines) > 0 {
		lines = lines[1:] // the first line may have been cut
	}
	for _, l := range lines {
		l = strings.TrimRight(l, "\r")
		if strings.TrimSpace(l) == "" {
			continue
		}
		tail = append(tail, l)
	}
	if len(tail) > TailLines {
		tail = tail[len(tail)-TailLines:]
	}
	return append([]byte(nil), c.out...), append([]byte(nil), c.std...), c.outN, c.stdN, tail
}
