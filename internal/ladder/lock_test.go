package ladder

// One ladder at a time per home (note 8): the lock held and released, a second climb waiting while the holder runs and
// refusing past its budget, a dead holder's lock taken at once, a holder that never names itself, and two homes that
// never block each other. The dead holder is this test binary run again as a helper process (TestHoldLockHelper),
// killed by the test: the system lets go of its lock with it, on both systems.

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
)

// holdEnv makes this test binary a lock holder: it takes the lock of the home the variable names, says "held" on
// stdout and waits to be killed.
const holdEnv = "BONSAI_TEST_HOLD_LADDER_LOCK"

// TestHoldLockHelper is the helper process's body; run as a plain test it does nothing.
func TestHoldLockHelper(t *testing.T) {
	home := os.Getenv(holdEnv)
	if home == "" {
		return
	}
	c := &Climb{Home: home, Root: home, o: Options{Getenv: noEnv}.withDefaults()}
	if _, e := takeLock(c); e != nil {
		os.Exit(3)
	}
	_, _ = os.Stdout.WriteString("held\n")
	time.Sleep(time.Hour)
}

// noEnv is the environment of a test's own lock holder: empty, so a holder never reads as a nested climb (one that
// takes no lock) when the tests themselves run inside a ladder's rung, whose environment holds BONSAI_LADDER=1.
func noEnv(string) string { return "" }

// holdLock takes the lock of a fixture's home, as a climb with a budget of budget s would, until release is called.
func holdLock(t *testing.T, f *fx, budget int) func() {
	t.Helper()
	timeout := int64(budget)
	c := &Climb{Home: f.home, Root: f.root, o: Options{Getenv: noEnv, LockPoll: f.lockPoll, LockGrace: f.lockGrace}.withDefaults()}
	c.rungs = append(c.rungs, format.Rung{Rung: 1, Kind: KindCommand, TimeoutS: &timeout})
	release, e := takeLock(c)
	if e != nil {
		t.Fatalf("taking the lock: %+v", e)
	}
	return release
}

func lockFixture(t *testing.T) *fx {
	return project(t, yamlOf("[1]", "{}", "[]", rung{n: 1, name: "one", command: "echo one", required: true}),
		map[string]string{"work/tasks/T-0901-a.md": taskText("T-0901", "running", "")})
}

// noticed is a climb's notice writer that says when the waiting line came, so a test lets go of the lock only once
// the climb waits for it, however slowly the climb gets there (a loaded Windows machine takes seconds).
type noticed struct {
	mu   sync.Mutex
	buf  bytes.Buffer
	once sync.Once
	came chan struct{}
}

func (n *noticed) Write(p []byte) (int, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if bytes.Contains(p, []byte("waiting while it runs")) {
		n.once.Do(func() { close(n.came) })
	}
	return n.buf.Write(p)
}

func (n *noticed) String() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.buf.String()
}

// A climb waits while the holder runs inside its budget, says once whom it waits for, and climbs once the holder lets
// go; the lock's file names the holder while it holds it.
func TestLockWaitsThenClimbs(t *testing.T) {
	f := lockFixture(t)
	release := holdLock(t, f, 60)
	words, _ := os.ReadFile(filepath.Join(f.home, filepath.FromSlash(LockFile)))
	if h, ok := readHolder(filepath.Join(f.home, filepath.FromSlash(LockFile))); !ok || h.PID != os.Getpid() || h.BudgetS != 60 ||
		h.Checkout != filepath.ToSlash(f.root) {
		t.Fatalf("the holder's words: %s", words)
	}
	n := &noticed{came: make(chan struct{})}
	let := make(chan time.Time, 1)
	go func() {
		select {
		case <-n.came:
			time.Sleep(200 * time.Millisecond) // the climb goes on waiting, polling, until the holder lets go
		case <-time.After(30 * time.Second):
		}
		let <- time.Now()
		release()
	}()
	a := f.climb("", Options{Task: "T-0901", Notice: n})
	done := time.Now()
	if a.Exit != 0 {
		t.Fatalf("exit %d, %+v", a.Exit, a.Error)
	}
	if released := <-let; done.Before(released) {
		t.Errorf("the climb ended %v before the holder let go", released.Sub(done))
	}
	if c := strings.Count(n.String(), "waiting while it runs"); c != 1 || !strings.Contains(n.String(), "the climb of no task") {
		t.Errorf("the waiting line, %d times:\n%s", c, n.String())
	}
	if raw, _ := os.ReadFile(filepath.Join(f.home, filepath.FromSlash(LockFile))); len(raw) != 0 {
		t.Errorf("the released lock still names a holder: %s", raw)
	}
}

// Past the holder's budget, a waiting climb refuses: ladder-busy, exit 4, naming the holder's process and the next step;
// nothing ran.
func TestLockRefusesPastTheBudget(t *testing.T) {
	f := lockFixture(t)
	start := time.Now()
	release := holdLock(t, f, 2)
	defer release()
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != ExitState || a.Error == nil || a.Error.Code != "ladder-busy" {
		t.Fatalf("exit %d, %+v", a.Exit, a.Error)
	}
	// The budget's end is read by the wall clock once (the holder's words), so a wall clock that jumps forward in that
	// moment shortens it: a second's margin.
	if took := time.Since(start); took < time.Second || took > 15*time.Second {
		t.Errorf("refused %v after the holder took the lock, its budget 2 s", took)
	}
	if !strings.Contains(a.Error.What, "process "+strconv.Itoa(os.Getpid())) || !strings.Contains(a.Error.Next, "wait until process "+strconv.Itoa(os.Getpid())+
		" has ended") || !strings.HasSuffix(a.Error.Next, "then run: bonsai ladder --task T-0901 --json") || a.Doc.Error.Next.Who != "agent" {
		t.Errorf("the refusal: %+v", a.Error)
	}
	if _, err := os.Stat(filepath.Join(f.root, ".bonsai", "local", "ladder", "T-0901.json")); err == nil {
		t.Error("a result was written")
	}
}

// A held lock whose holder names no process that runs is waited for LockGrace, then refused.
func TestLockHolderUnnamed(t *testing.T) {
	f := lockFixture(t)
	release := holdLock(t, f, 60)
	defer release()
	path := filepath.Join(f.home, filepath.FromSlash(LockFile))
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Skipf("the lock file cannot be written beside its holder here: %v", err)
	}
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != ExitState || a.Error == nil || a.Error.Code != "ladder-busy" || !strings.Contains(a.Error.What, "has not written its words") {
		t.Fatalf("exit %d, %+v", a.Exit, a.Error)
	}
}

// A holder killed while it holds the lock lets go of it with its process: the next climb takes it at once.
func TestLockOfADeadHolder(t *testing.T) {
	f := lockFixture(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestHoldLockHelper$", "-test.count=1")
	cmd.Env = append(os.Environ(), holdEnv+"="+f.home)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	held := make(chan bool, 1)
	go func() {
		line, _ := bufio.NewReader(out).ReadString('\n')
		held <- strings.TrimSpace(line) == "held"
	}()
	select {
	case ok := <-held:
		if !ok {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			t.Fatal("the helper did not take the lock")
		}
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatal("the helper did not take the lock in 30 s")
	}
	if h, ok := readHolder(filepath.Join(f.home, filepath.FromSlash(LockFile))); !ok || h.PID != cmd.Process.Pid {
		t.Errorf("the helper's words: %+v %v", h, ok)
	}
	_ = cmd.Process.Kill()
	_ = cmd.Wait()
	// Judged by what a wait would leave, not by the clock: a climb that waited for the dead holder would have told of
	// it ("waiting while it runs") and ended busy or late; this one took the lock at once, so its notice is empty. How
	// long the whole climb takes is a loaded runner's business (a 5 s limit failed at 6.8 s under load).
	a := f.climb("", Options{Task: "T-0901"})
	if a.Exit != 0 || strings.Contains(f.notice.String(), "waiting") || f.notice.Len() != 0 {
		t.Fatalf("exit %d, %+v\n%s", a.Exit, a.Error, f.notice.String())
	}
}

// Two homes hold two locks: a climb on one never waits for the other's holder.
func TestLockTwoHomes(t *testing.T) {
	f := lockFixture(t)
	other := t.TempDir()
	c := &Climb{Home: other, Root: other, o: Options{Getenv: noEnv}.withDefaults()}
	release, e := takeLock(c)
	if e != nil {
		t.Fatal(e)
	}
	defer release()
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 || f.notice.Len() != 0 {
		t.Fatalf("exit %d, %+v\n%s", a.Exit, a.Error, f.notice.String())
	}
}

// A nested climb (BONSAI_LADDER=1: inside another climb's rung) takes no lock: its outer climb holds it.
func TestNestedTakesNoLock(t *testing.T) {
	f := lockFixture(t)
	release := holdLock(t, f, 60)
	defer release()
	f.env["BONSAI_LADDER"] = "1"
	if a := f.climb("", Options{Task: "T-0901"}); a.Exit != 0 || f.notice.Len() != 0 {
		t.Fatalf("exit %d, %+v", a.Exit, a.Error)
	}
	if l := f.result("T-0901"); l.Leftovers != nil {
		t.Errorf("a nested climb's leftovers: %v", l.Leftovers)
	}
}

// The budget is the sum of the rungs' limits: each rung's timeout_s, else its kind's own default, which is also the
// limit it runs with: a command rung 600 s, rung 0 and git integrity 120 s each (a floor of both is not a wait of 20
// minutes).
func TestBudgetByKind(t *testing.T) {
	five := int64(5)
	c := &Climb{rungs: []format.Rung{{Rung: 0, Kind: KindGuard}, {Rung: 1, Kind: KindCommand}, {Rung: 2, Kind: KindVerGit},
		{Rung: 3, Kind: KindGuard, TimeoutS: &five}, {Rung: 4, Kind: KindVerGit, TimeoutS: &five}, {Rung: 5, Kind: KindCommand, TimeoutS: &five}}}
	if got := c.budget(); got != 120+600+120+5+5+5 {
		t.Errorf("the budget %d", got)
	}
	for _, r := range c.rungs[:3] {
		if want := map[string]time.Duration{KindGuard: 120 * time.Second, KindCommand: 600 * time.Second,
			KindVerGit: 120 * time.Second}[r.Kind]; rungTimeout(r) != want {
			t.Errorf("%s's limit %v, want %v", r.Kind, rungTimeout(r), want)
		}
	}
}
