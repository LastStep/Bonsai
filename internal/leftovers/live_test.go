package leftovers

// The check on the system's real process table. The test binary is its own child program (TestMain): a sleeper
// that writes its pid and sleeps, and a planter that starts a sleeper and exits, so the sleeper is no longer this
// process's descendant. Every process a test starts names the test's own folder in its command line, and the
// test's cleanup ends each one still alive, by pid, after checking that.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const childArg = "leftovers-child"

func TestMain(m *testing.M) {
	if len(os.Args) > 2 && os.Args[1] == childArg {
		os.Exit(child(os.Args[2:]))
	}
	os.Exit(m.Run())
}

// child runs one mode:
//
//	sleep <pidfile> [<arg>]               write its pid, sleep ten minutes
//	plant <pidfile> <dir> <arg>           start "sleep <pidfile> <arg>" in dir, wait for its pid file, exit 0
func child(args []string) int {
	switch args[0] {
	case "sleep":
		if err := os.WriteFile(args[1]+".tmp", []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
			return 3
		}
		if err := os.Rename(args[1]+".tmp", args[1]); err != nil {
			return 3
		}
		time.Sleep(10 * time.Minute)
		return 0
	case "plant":
		cmd := sleeper(args[1], args[2], args[3])
		if err := cmd.Start(); err != nil {
			return 5
		}
		if !waitFile(args[1], 10*time.Second) {
			return 4
		}
		return 0
	}
	return 2
}

func sleeper(pidfile, dir, arg string) *exec.Cmd {
	exe, _ := os.Executable()
	cmd := exec.Command(exe, childArg, "sleep", pidfile, arg)
	cmd.Dir = dir
	return cmd
}

func waitFile(file string, d time.Duration) bool {
	end := time.Now().Add(d)
	for {
		if _, err := os.Stat(file); err == nil {
			return true
		}
		if time.Now().After(end) {
			return false
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func readPID(t *testing.T, file string) int {
	t.Helper()
	if !waitFile(file, 10*time.Second) {
		t.Fatalf("no pid file %s", filepath.Base(file))
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	return pid
}

// A leftover planted in the checkout is named; this process's own child working there and a process of the same
// command outside the checkout are not.
func TestCheckFindsAPlantedLeftover(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		if v := Check(t.TempDir()); v.Kind != NotChecked || v.Why != runtime.GOOS {
			t.Errorf("%+v, want not checked (%s)", v, runtime.GOOS)
		}
		t.Skip("the process table is read on Linux and Windows only; elsewhere the check says not checked")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root, outside := filepath.Join(base, "checkout"), filepath.Join(base, "elsewhere")
	for _, d := range []string{root, outside} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { endTestProcesses(t, base) })
	exe, _ := os.Executable()
	plant := func(pidfile, dir, arg string) int {
		if out, err := exec.Command(exe, childArg, "plant", pidfile, dir, arg).CombinedOutput(); err != nil {
			t.Fatalf("plant: %v: %s", err, out)
		}
		return readPID(t, pidfile)
	}
	planted := plant(filepath.Join(base, "planted.pid"), root, filepath.Join(root, "marker"))
	decoy := plant(filepath.Join(base, "decoy.pid"), outside, filepath.Join(outside, "marker"))
	own := sleeper(filepath.Join(base, "own.pid"), root, filepath.Join(root, "marker"))
	if err := own.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = own.Process.Kill(); _ = own.Wait() })
	ownPID := readPID(t, filepath.Join(base, "own.pid"))

	v := Check(root)
	if v.Kind != Found {
		t.Fatalf("%+v, want found", v)
	}
	named := map[int]Process{}
	for _, p := range v.Checkout {
		named[p.PID] = p
	}
	p, ok := named[planted]
	if !ok {
		t.Errorf("the planted leftover (pid %d) is not named: %+v", planted, v.Checkout)
	} else if !strings.HasPrefix(p.Name, "leftovers.test") || p.Head != "" {
		t.Errorf("the planted leftover reads %+v", p)
	}
	if _, ok := named[ownPID]; ok {
		t.Errorf("this process's own child (pid %d) is named", ownPID)
	}
	if _, ok := named[decoy]; ok {
		t.Errorf("the decoy outside the checkout (pid %d) is named", decoy)
	}
	if _, ok := named[os.Getpid()]; ok {
		t.Errorf("the runner itself is named")
	}
	if v.MS < 0 || v.MS > 5000 {
		t.Errorf("the check took %d ms", v.MS)
	}
	t.Logf("the check took %d ms and named %d processes", v.MS, len(v.Checkout))
}

// endTestProcesses ends every process still alive whose pid is in a pid file under dir and whose command line
// names dir.
func endTestProcesses(t *testing.T, dir string) {
	files, _ := filepath.Glob(filepath.Join(dir, "*.pid"))
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
		if err != nil || pid <= 0 {
			continue
		}
		if namesDir(pid, dir) {
			t.Logf("cleanup: ending pid %d (%s)", pid, filepath.Base(f))
			endPID(pid)
		}
	}
}
