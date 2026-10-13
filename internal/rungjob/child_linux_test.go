//go:build linux

package rungjob

import (
	"errors"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/LastStep/Bonsai/internal/leftovers"
)

// spawn starts a sleeper the given way and leaves it (the child mode's helper).
//
//	plain            a plain child
//	double           a plain child that starts the sleeper and exits: the sleeper's parent is gone
//	setsid           a child in a new session (it keeps the rung's tag)
//	clearenv-setsid  a child in a new session with an empty environment (no tag): only the reaper can tell it
//	foreign-setsid   a child in a new session whose tag is one no rung of this process runs (BONSAI_RUNG_TAG=other)
func spawn(how, pidfile string) error {
	cmd := sleeperCmd(pidfile)
	switch how {
	case "plain":
	case "double":
		exe, _ := os.Executable()
		mid := sleeperCmd(pidfile)
		mid.Args = []string{exe, childArg, "spawn", "plain", pidfile}
		return mid.Run()
	case "setsid":
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	case "clearenv-setsid":
		cmd.Env = []string{}
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	case "foreign-setsid":
		cmd.Env = append(os.Environ(), TagEnv+"=other")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	default:
		return errors.New("unknown way: " + how)
	}
	return cmd.Start()
}

// leaveCases are the ways a Linux rung leaves a process behind (plan 5.4.1 note 5).
func leaveCases(dir string) []leaveCase {
	p := dir + "/left.pid"
	cases := []leaveCase{
		{"a background child", self("sleep", p) + " & " + self("waitfile", p)},
		{"a nohup child", "nohup " + self("sleep", p) + " >/dev/null 2>&1 & " + self("waitfile", p)},
		{"a child in a new session, by setsid", self("spawn", "setsid", p)},
		{"a double fork in the shell", "( " + self("sleep", p) + " & ) ; " + self("waitfile", p)},
		{"a double fork, its parent gone", self("spawn", "double", p)},
		{"a cleared environment in a new session, its parent gone", self("spawn", "clearenv-setsid", p)},
		{"another tag in a new session, its parent gone", self("spawn", "foreign-setsid", p)},
		{"a child holding the output pipe", self("hold", p) + " & " + self("waitfile", p)},
	}
	if _, err := os.Stat("/usr/bin/setsid"); err == nil {
		cases = append(cases, leaveCase{"a setsid daemon", "setsid " + self("sleep", p) + " & " + self("waitfile", p)},
			leaveCase{"a setsid -f daemon with another tag", TagEnv + "=other setsid -f " + self("sleep", p) + "; " + self("waitfile", p)})
	}
	return cases
}

// timeoutCommand leaves one process in the background and runs another in the foreground until the limit.
func timeoutCommand(a, b string) string { return self("sleep", a) + " & " + self("sleep", b) }

// killedRunnerCommand records the shell's own pid, then leaves one sleeper and waits on another.
func killedRunnerCommand(dir string) string {
	return "echo $$ > " + quote(dir+"/shell.pid") + "; " + self("sleep", dir+"/a.pid") + " & " +
		self("sleep", dir+"/b.pid")
}

func alive(pid int) bool {
	s, ok := leftovers.ReadStat(pid)
	return ok && !s.Dead()
}

func commandNames(pid int, dir string) bool {
	for _, a := range leftovers.Argv(pid) {
		if strings.Contains(a, dir) {
			return true
		}
	}
	return false
}

// endPID ends one of the test's processes and, if it is this process's child (an adopted one), reaps it.
func endPID(pid int) {
	_ = syscall.Kill(pid, syscall.SIGKILL)
	for i := 0; i < 200; i++ {
		var ws syscall.WaitStatus
		wpid, err := syscall.Wait4(pid, &ws, syscall.WNOHANG, nil)
		if wpid == pid || err != nil && !errors.Is(err, syscall.EINTR) {
			if !alive(pid) {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// killRunner ends the helper runner the hard way: SIGKILL, which no handler sees.
func killRunner(pid int) error { return syscall.Kill(pid, syscall.SIGKILL) }
