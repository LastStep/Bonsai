//go:build linux || windows

package rungjob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// leaveCase is one way a rung leaves a process behind: its command waits until the leftover has written left.pid.
type leaveCase struct {
	name    string
	command string
}

// The leftovers' lists end by themselves within this, after the rung's end.
const goneWithin = 3 * time.Second

func TestOutputAndExit(t *testing.T) {
	dir := testDir(t)
	r := Run(context.Background(), Spec{Command: self("print", "to-out", "to-err", "3"), Dir: dir})
	if r.Exit != 3 || r.Green() || r.TimedOut || r.Stopped {
		t.Fatalf("exit %d, green %v, timed out %v, stopped %v; want exit 3, red", r.Exit, r.Green(), r.TimedOut, r.Stopped)
	}
	if string(r.Stdout) != "to-out\n" {
		t.Errorf("stdout %q, want only the standard output", r.Stdout)
	}
	if !bytes.Contains(r.Out, []byte("to-out\n")) || !bytes.Contains(r.Out, []byte("to-err\n")) || len(r.Out) != 14 {
		t.Errorf("out %q, want both streams", r.Out)
	}
	if r.OutSize != 14 || r.StdoutSize != 7 {
		t.Errorf("sizes %d, %d; want 14, 7", r.OutSize, r.StdoutSize)
	}
	tail := slices.Clone(r.Tail)
	slices.Sort(tail)
	if !slices.Equal(tail, []string{"to-err", "to-out"}) {
		t.Errorf("tail %q", r.Tail)
	}
	if r.Ended == nil || len(r.Ended) != 0 {
		t.Errorf("ended %v, want an empty list: nothing was left", r.Ended)
	}
	if r.Job != "on" || r.Reason != "" {
		t.Errorf("job %q, reason %q; want on and none", r.Job, r.Reason)
	}
	ok := Run(context.Background(), Spec{Command: self("print", "a", "b", "0"), Dir: dir})
	if ok.Exit != 0 || !ok.Green() {
		t.Errorf("exit %d green %v; want a green rung", ok.Exit, ok.Green())
	}
}

func TestEnvironment(t *testing.T) {
	dir := testDir(t)
	env := append(os.Environ(), TagEnv+"=outer", LadderEnv+"=0")
	r := Run(context.Background(), Spec{Command: self("env", LadderEnv, TagEnv), Dir: dir, Env: env})
	if !r.Green() {
		t.Fatalf("exit %d: %s", r.Exit, r.Out)
	}
	out := strings.ReplaceAll(string(r.Stdout), "\r", "")
	if !strings.Contains(out, LadderEnv+"=1\n") {
		t.Errorf("the rung's environment lacks %s=1:\n%s", LadderEnv, out)
	}
	want := regexp.MustCompile(`(?m)^` + TagEnv + `=outer:[0-9a-f]{32}$`)
	if runtime.GOOS == "windows" {
		want = regexp.MustCompile(`(?m)^` + TagEnv + `=outer$`) // the job does the work: the tag passes as it was
	}
	if !want.MatchString(out) {
		t.Errorf("the tag: want %s in\n%s", want, out)
	}
}

// Every way a rung leaves a process: the rung stays green, the process is named among what it ended, and it is
// gone when Run returns (or within goneWithin, the system's own delay).
func TestEndsWhatItLeaves(t *testing.T) {
	for i, c := range leaveCases(t.TempDir()) {
		t.Run(c.name, func(t *testing.T) {
			dir := testDir(t)
			cmd := leaveCases(dir)[i].command
			r := Run(context.Background(), Spec{Command: cmd, Dir: dir, Timeout: time.Minute})
			if !waitFile(filepath.Join(dir, "left.pid"), 0) {
				t.Fatalf("the rung left nothing: exit %d, reason %q\n%s", r.Exit, r.Reason, r.Out)
			}
			pid := readPID(t, filepath.Join(dir, "left.pid"))
			if !r.Green() {
				t.Errorf("exit %d, reason %q, notes %q; want green (a leftover never changes the colour)\n%s",
					r.Exit, r.Reason, r.Notes, r.Out)
			}
			if !endedHas(r, pid) {
				t.Errorf("the leftover (pid %d) is not among what the rung ended: %+v", pid, r.Ended)
			}
			if r.Duration > PipeCutoff/2 {
				t.Errorf("the rung took %s: its pipes were held", r.Duration)
			}
			for _, n := range r.Notes {
				t.Errorf("note: %s", n)
			}
			waitGone(t, pid, goneWithin, "the leftover")
			if strings.Contains(c.name, "pipe") && !bytes.Contains(r.Out, []byte("holding the pipe")) {
				t.Errorf("the holder's output is lost: %q", r.Out)
			}
		})
	}
}

func TestTimeout(t *testing.T) {
	dir := testDir(t)
	a, b := filepath.Join(dir, "a.pid"), filepath.Join(dir, "b.pid")
	r := Run(context.Background(), Spec{Command: timeoutCommand(a, b), Dir: dir, Timeout: time.Second})
	if !r.TimedOut || r.Exit != -1 || r.Green() {
		t.Fatalf("timed out %v, exit %d; want a red timeout", r.TimedOut, r.Exit)
	}
	want := fmt.Sprintf("timed out after 1 s; ended %d processes", len(r.Ended))
	if r.Reason != want {
		t.Errorf("reason %q, want %q", r.Reason, want)
	}
	pa, pb := readPID(t, a), readPID(t, b)
	if len(r.Ended) < 3 || !endedHas(r, pa) || !endedHas(r, pb) {
		t.Errorf("ended %+v: want the shell and both sleepers (%d, %d)", r.Ended, pa, pb)
	}
	waitGone(t, pa, goneWithin, "the background sleeper")
	waitGone(t, pb, goneWithin, "the foreground sleeper")
	if r.Duration > 5*time.Second {
		t.Errorf("a 1 s limit took %s", r.Duration)
	}
}

func TestStop(t *testing.T) {
	dir := testDir(t)
	a, b := filepath.Join(dir, "a.pid"), filepath.Join(dir, "b.pid")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		waitFile(b, 10*time.Second)
		cancel()
	}()
	r := Run(ctx, Spec{Command: timeoutCommand(a, b), Dir: dir, Timeout: time.Minute})
	if !r.Stopped || r.TimedOut || r.Green() {
		t.Fatalf("stopped %v, timed out %v; want a stop", r.Stopped, r.TimedOut)
	}
	if want := fmt.Sprintf("stopped; ended %d processes", len(r.Ended)); r.Reason != want {
		t.Errorf("reason %q, want %q", r.Reason, want)
	}
	waitGone(t, readPID(t, a), goneWithin, "the background sleeper")
	waitGone(t, readPID(t, b), goneWithin, "the foreground sleeper")
}

// A process of the same user started outside the rung, with the same command, while the rung runs, is left alone.
func TestDecoyLeftAlone(t *testing.T) {
	dir := testDir(t)
	left, decoyPID, start := filepath.Join(dir, "left.pid"), filepath.Join(dir, "decoy.pid"), filepath.Join(dir, "go")
	var decoy *exec.Cmd
	started := make(chan error, 1)
	go func() {
		if !waitFile(left, 10*time.Second) {
			started <- fmt.Errorf("the rung's sleeper never started")
			return
		}
		decoy = sleeperCmd(decoyPID)
		if err := decoy.Start(); err != nil {
			started <- err
			return
		}
		waitFile(decoyPID, 10*time.Second)
		started <- os.WriteFile(start, nil, 0o644)
	}()
	cmd := timeoutCommand(left, filepath.Join(dir, "unused.pid"))
	cmd = strings.Replace(cmd, self("sleep", filepath.Join(dir, "unused.pid")), self("waitfile", start), 1)
	r := Run(context.Background(), Spec{Command: cmd, Dir: dir, Timeout: time.Minute})
	if err := <-started; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = decoy.Process.Kill()
		_ = decoy.Wait()
	})
	if !r.Green() {
		t.Errorf("exit %d: %s", r.Exit, r.Out)
	}
	lp, dp := readPID(t, left), readPID(t, decoyPID)
	if endedHas(r, dp) {
		t.Errorf("the decoy (pid %d) is among what the rung ended", dp)
	}
	waitGone(t, lp, goneWithin, "the rung's sleeper")
	time.Sleep(200 * time.Millisecond)
	if !alive(dp) {
		t.Errorf("the decoy (pid %d), outside the rung, was ended", dp)
	}
}

// A rung that runs a runner of its own (a nested ladder): its limit ends the inner runner, the inner rung's shell
// and what that started, in the inner job (Windows: a job inside the rung's job) or the inner session (Linux: the
// inner tag carries the outer's).
func TestNestedRunner(t *testing.T) {
	dir := testDir(t)
	inner := filepath.Join(dir, "inner.cmd")
	sleeper := filepath.Join(dir, "sleeper.pid")
	if err := os.WriteFile(inner, []byte(self("sleep", sleeper)), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		waitFile(sleeper, 20*time.Second)
		cancel()
	}()
	r := Run(ctx, Spec{Command: self("runner", dir, "600", "@"+inner), Dir: dir, Timeout: time.Minute})
	if !r.Stopped {
		t.Fatalf("stopped %v, exit %d, reason %q\n%s", r.Stopped, r.Exit, r.Reason, r.Out)
	}
	sp, rp := readPID(t, sleeper), readPID(t, filepath.Join(dir, "runner.pid"))
	if !endedHas(r, sp) || !endedHas(r, rp) {
		t.Errorf("ended %+v: want the inner runner (%d) and the inner rung's sleeper (%d)", r.Ended, rp, sp)
	}
	waitGone(t, sp, goneWithin, "the inner rung's sleeper")
	waitGone(t, rp, goneWithin, "the inner runner")
}

// The runner killed mid-rung, the hard way. Windows: its job's handle closes with it and the job ends the tree.
// Linux: its shell dies with it (Pdeathsig), and what escaped the shell is named by the next run's leftovers line.
func TestRunnerKilled(t *testing.T) {
	dir := testDir(t)
	cmdFile := filepath.Join(dir, "rung.cmd")
	if err := os.WriteFile(cmdFile, []byte(killedRunnerCommand(dir)), 0o644); err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	runner := exec.Command(exe, childArg, "runner", dir, "600", "@"+cmdFile)
	if err := runner.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runner.Process.Kill(); _ = runner.Wait() })
	pa, pb := readPID(t, filepath.Join(dir, "a.pid")), readPID(t, filepath.Join(dir, "b.pid"))
	if err := killRunner(runner.Process.Pid); err != nil {
		t.Fatal(err)
	}
	_ = runner.Wait()
	if runtime.GOOS == "windows" {
		waitGone(t, pa, goneWithin, "the rung's background sleeper")
		waitGone(t, pb, goneWithin, "the rung's foreground sleeper")
		return
	}
	waitGone(t, readPID(t, filepath.Join(dir, "shell.pid")), goneWithin, "the rung's shell (Pdeathsig)")
	out, err := exec.Command(exe, childArg, "leftovers", dir).Output()
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		Checkout []struct {
			PID int `json:"pid"`
		} `json:"checkout"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	named := map[int]bool{}
	for _, p := range v.Checkout {
		named[p.PID] = true
	}
	if !named[pa] || !named[pb] {
		t.Errorf("the next leftovers line %s does not name what escaped (%d, %d)", out, pa, pb)
	}
}

func TestNotStarted(t *testing.T) {
	dir := testDir(t)
	r := Run(context.Background(), Spec{Command: self("print", "a", "b", "0"), Dir: filepath.Join(dir, "missing")})
	if r.Exit != -1 || r.Green() || !strings.HasPrefix(r.Reason, "could not start: ") {
		t.Errorf("exit %d, reason %q; want -1 and could not start", r.Exit, r.Reason)
	}
	if strings.Contains(r.Reason, `\`) {
		t.Errorf("reason %q holds a backslash", r.Reason)
	}
	e := Run(context.Background(), Spec{Command: "  ", Dir: dir})
	if e.Exit != -1 || e.Reason != "could not start: the command is empty" {
		t.Errorf("exit %d, reason %q", e.Exit, e.Reason)
	}
}

func TestCapture(t *testing.T) {
	c := &capture{}
	big := bytes.Repeat([]byte("x"), OutputCap+10)
	c.write(big, true)
	var lines bytes.Buffer
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&lines, "line %d\r\n\n", i)
	}
	c.write(lines.Bytes(), false)
	out, std, outN, stdN, tail := c.result()
	if len(out) != OutputCap || len(std) != OutputCap {
		t.Errorf("kept %d and %d bytes; want the cap, %d", len(out), len(std), OutputCap)
	}
	if outN != int64(len(big)+lines.Len()) || stdN != int64(len(big)) {
		t.Errorf("sizes %d, %d", outN, stdN)
	}
	if len(tail) != TailLines || tail[0] != "line 61" || tail[TailLines-1] != "line 100" {
		t.Errorf("tail %q", tail)
	}
	c.write([]byte("late\n"), false)
	if _, _, n, _, _ := c.result(); n != outN {
		t.Errorf("a write after the result was kept")
	}
	// A tail cut inside a line drops the cut line.
	d := &capture{}
	d.write(bytes.Repeat([]byte("y"), 3*tailBytes), false)
	d.write([]byte("\nlast\n"), false)
	if _, _, _, _, tl := d.result(); !slices.Equal(tl, []string{"last"}) {
		t.Errorf("tail %q, want only the whole line", tl)
	}
}

func TestEnvList(t *testing.T) {
	env := setEnv([]string{"A=1", "B=2", "A=3"}, "A", "4")
	if !slices.Equal(env, []string{"B=2", "A=4"}) {
		t.Errorf("setEnv: %q", env)
	}
	if getEnv([]string{"A=1", "A=2"}, "A") != "2" || getEnv(nil, "A") != "" {
		t.Errorf("getEnv")
	}
	if runtime.GOOS == "windows" && getEnv([]string{"comspec=x"}, "ComSpec") != "x" {
		t.Errorf("getEnv on Windows reads names in any case")
	}
	if seconds(600*time.Second) != "600" || seconds(1500*time.Millisecond) != "1.5" {
		t.Errorf("seconds: %s, %s", seconds(600*time.Second), seconds(1500*time.Millisecond))
	}
}
