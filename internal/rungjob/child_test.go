//go:build linux || windows

package rungjob

// The test binary is its own child program: started with childArg first, TestMain runs one of the modes below and
// exits, so the tests start real process trees with no shell tricks on Windows. Every process a test starts writes
// its pid to a file in the test's own temporary folder, and the test's cleanup ends each one still alive whose
// command line names that folder (track), so a failing test leaves nothing running.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/leftovers"
	"github.com/LastStep/Bonsai/internal/schema"
)

const childArg = "rungjob-child"

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == childArg {
		os.Exit(child(os.Args[2:]))
	}
	os.Exit(m.Run())
}

// child runs one mode:
//
//	sleep <pidfile>                    write its pid, sleep ten minutes
//	hold <pidfile>                     print a line, write its pid, sleep (it holds the rung's output pipes)
//	print <out> <err> <code>           print to standard output and error, exit with code
//	env <name>...                      print each variable as name=value, one a line
//	waitfile <file>                    wait until the file exists (ten seconds at most), exit 0
//	spawn <how> <pidfile>              start "sleep <pidfile>" in the given way, wait for its pid file, exit 0
//	runner <dir> <timeout-s> @<file>   write <dir>/runner.pid, run the file's command as a rung (Run), print the
//	                                   Result as JSON
//	leftovers <root>                   print leftovers.Check(root) as JSON
func child(args []string) int {
	if len(args) == 0 {
		return 2
	}
	switch args[0] {
	case "sleep", "hold":
		if args[0] == "hold" {
			fmt.Println("holding the pipe")
		}
		if err := writePID(args[1], os.Getpid()); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return 3
		}
		time.Sleep(10 * time.Minute)
		return 0
	case "print":
		_, _ = fmt.Fprintln(os.Stdout, args[1])
		_, _ = fmt.Fprintln(os.Stderr, args[2])
		code, _ := strconv.Atoi(args[3])
		return code
	case "env":
		for _, n := range args[1:] {
			fmt.Printf("%s=%s\n", n, os.Getenv(n))
		}
		return 0
	case "waitfile":
		if waitFile(args[1], 10*time.Second) {
			return 0
		}
		return 4
	case "spawn":
		pidfile := args[2]
		if err := spawn(args[1], pidfile); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "spawn:", err)
			return 5
		}
		if !waitFile(pidfile, 10*time.Second) {
			return 4
		}
		return 0
	case "runner":
		_ = writePID(filepath.Join(args[1], "runner.pid"), os.Getpid())
		secs, _ := strconv.ParseFloat(args[2], 64)
		command := args[3]
		if strings.HasPrefix(command, "@") {
			b, err := os.ReadFile(command[1:])
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return 3
			}
			command = string(b)
		}
		r := Run(context.Background(), Spec{Command: command, Dir: args[1], Timeout: time.Duration(secs * float64(time.Second))})
		b, _ := json.Marshal(r)
		fmt.Println(string(b))
		return 0
	case "leftovers":
		b, err := schema.EncodeLine(leftovers.Check(args[1]).Record())
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return 3
		}
		fmt.Print(string(b))
		return 0
	}
	return 2
}

// sleeperCmd starts the test binary in sleep mode with the given pid file, as a plain child (spawn's helpers).
func sleeperCmd(pidfile string) *exec.Cmd {
	exe, _ := os.Executable()
	return exec.Command(exe, childArg, "sleep", pidfile)
}

func writePID(file string, pid int) error {
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, file)
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

// testDir is a test's own folder: its name is the marker track finds the test's processes by.
func testDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	track(t, dir)
	return dir
}

// track ends, at the test's end, every process still alive whose pid is in a pid file under dir and whose command
// line names dir: by pid, after checking it is the test's own.
func track(t *testing.T, dir string) {
	t.Cleanup(func() {
		files, _ := filepath.Glob(filepath.Join(dir, "*.pid"))
		for _, f := range files {
			pid, err := readPIDFile(f, 2*time.Second)
			if err != nil {
				continue
			}
			if alive(pid) && commandNames(pid, dir) {
				t.Logf("cleanup: ending pid %d (%s), still running", pid, filepath.Base(f))
				endPID(pid)
			}
		}
	})
}

// readPID waits for a pid file (ten seconds at most) and reads it.
func readPID(t *testing.T, file string) int {
	t.Helper()
	if !waitFile(file, 10*time.Second) {
		t.Fatalf("no pid file %s", filepath.Base(file))
	}
	pid, err := readPIDFile(file, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return pid
}

// readPIDFile reads a pid file, again until d has passed when the read fails or the content is empty or not a pid.
// Windows opens a file that is being written, or renamed onto, with a sharing violation; the file is whole a moment
// later, so one failed read is not the answer.
func readPIDFile(file string, d time.Duration) (int, error) {
	end := time.Now().Add(d)
	for {
		b, err := os.ReadFile(file)
		if err == nil {
			var pid int
			if pid, err = strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
				return pid, nil
			}
			if err == nil {
				err = fmt.Errorf("pid file %s holds %d", filepath.Base(file), pid)
			}
		}
		if time.Now().After(end) {
			return 0, err
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// waitGone fails the test when pid is still alive after d.
func waitGone(t *testing.T, pid int, d time.Duration, what string) {
	t.Helper()
	end := time.Now().Add(d)
	for alive(pid) {
		if time.Now().After(end) {
			t.Errorf("%s (pid %d) still alive %s after the rung's end", what, pid, d)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// self is the test binary as a shell word, then childArg and the arguments, each quoted for the rung's shell.
func self(args ...string) string {
	exe, _ := os.Executable()
	words := []string{quote(exe), childArg}
	for _, a := range args {
		words = append(words, quote(a))
	}
	return strings.Join(words, " ")
}

// quote quotes one word for the rung's shell: single quotes for sh, double quotes for cmd.
func quote(s string) string {
	if runtime.GOOS == "windows" {
		return `"` + s + `"`
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// endedHas reports whether a Result's Ended names pid.
func endedHas(r Result, pid int) bool {
	for _, p := range r.Ended {
		if p.PID == pid {
			return true
		}
	}
	return false
}
