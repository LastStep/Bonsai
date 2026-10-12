//go:build windows

package rungjob

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"

	"github.com/LastStep/Bonsai/internal/leftovers"
)

const (
	detachedProcess        = 0x00000008 // DETACHED_PROCESS
	createNewProcessGroup  = 0x00000200 // CREATE_NEW_PROCESS_GROUP
	createBreakawayFromJob = 0x01000000 // CREATE_BREAKAWAY_FROM_JOB
	waitTimeout            = 0x00000102 // WAIT_TIMEOUT
)

// spawn starts a sleeper the given way and leaves it (the child mode's helper).
//
//	plain       a plain child
//	double      a plain child that starts the sleeper and exits: the sleeper's parent is gone
//	detached    a detached child in a new process group, with no console
//	breakaway   a child asking to leave the job: the system must refuse it, which <pidfile>.refused records
//	nested-job  a child in a job of its own, made inside the rung's (a nested job), without kill-on-close, so it
//	            outlives the job's maker
func spawn(how, pidfile string) error {
	cmd := sleeperCmd(pidfile)
	switch how {
	case "plain":
	case "double":
		exe, _ := os.Executable()
		mid := exec.Command(exe, childArg, "spawn", "plain", pidfile)
		return mid.Run()
	case "detached":
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: detachedProcess | createNewProcessGroup}
	case "breakaway":
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createBreakawayFromJob}
		if err := cmd.Start(); err != nil {
			_ = os.WriteFile(pidfile+".refused", []byte(err.Error()), 0o644)
			return err
		}
		return nil
	case "nested-job":
		return startInNestedJob(cmd)
	default:
		return errors.New("unknown way: " + how)
	}
	return cmd.Start()
}

// startInNestedJob makes a job with no limits, starts cmd suspended, puts it in the job and lets it run. The job's
// handle closes when this process exits; with no kill-on-close, its member lives on.
func startInNestedJob(cmd *exec.Cmd) error {
	h, _, err := procCreateJobObjectW.Call(0, 0)
	if h == 0 {
		return fmt.Errorf("CreateJobObjectW: %w", err)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createSuspended}
	if err := cmd.Start(); err != nil {
		return err
	}
	p, err := syscall.OpenProcess(processSetQuota|processTerminate|processSuspendResume, false, uint32(cmd.Process.Pid))
	if err != nil {
		return err
	}
	defer func() { _ = syscall.CloseHandle(p) }()
	if ok, _, err := procAssignProcessToJobObject.Call(h, uintptr(p)); ok == 0 {
		_ = cmd.Process.Kill()
		return fmt.Errorf("AssignProcessToJobObject (nested): %w", err)
	}
	if st, _, _ := procNtResumeProcess.Call(uintptr(p)); st != 0 {
		_ = cmd.Process.Kill()
		return fmt.Errorf("NtResumeProcess: 0x%08X", uint32(st))
	}
	return nil
}

// powershellStart is PowerShell's Start-Process starting a sleeper (hidden), as -EncodedCommand, so no quoting
// reaches cmd.
func powershellStart(pidfile string) string {
	exe, _ := os.Executable()
	script := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList @('%s','sleep','\"%s\"') -WindowStyle Hidden",
		strings.ReplaceAll(exe, "'", "''"), childArg, strings.ReplaceAll(pidfile, "'", "''"))
	u := utf16.Encode([]rune(script))
	b := make([]byte, 2*len(u))
	for i, c := range u {
		b[2*i] = byte(c)
		b[2*i+1] = byte(c >> 8)
	}
	ps := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	return quote(ps) + " -NoProfile -NonInteractive -EncodedCommand " + base64.StdEncoding.EncodeToString(b)
}

// leaveCases are the ways a Windows rung leaves a process behind (plan 5.4.1 note 5).
func leaveCases(dir string) []leaveCase {
	p := dir + `\left.pid`
	return []leaveCase{
		{`start "" /b`, `start "" /b ` + self("sleep", p) + " & " + self("waitfile", p)},
		{"cmd /c start", `cmd /d /c start "" /min ` + self("sleep", p) + " & " + self("waitfile", p)},
		{"PowerShell's Start-Process", powershellStart(p) + " & " + self("waitfile", p)},
		{"a detached child in a new process group", self("spawn", "detached", p)},
		{"a double fork, its parent gone", self("spawn", "double", p)},
		{"a job inside the rung's job", self("spawn", "nested-job", p)},
		{"a child holding the output pipe", `start "" /b ` + self("hold", p) + " & " + self("waitfile", p)},
	}
}

// timeoutCommand leaves one process in the background and runs another in the foreground until the limit.
func timeoutCommand(a, b string) string {
	return `start "" /b ` + self("sleep", a) + " & " + self("sleep", b)
}

// killedRunnerCommand leaves one sleeper and waits on another.
func killedRunnerCommand(dir string) string {
	return `start "" /b ` + self("sleep", dir+`\a.pid`) + " & " + self("sleep", dir+`\b.pid`)
}

func alive(pid int) bool {
	h, err := syscall.OpenProcess(syscall.SYNCHRONIZE|leftovers.QueryLimited, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	ev, _ := syscall.WaitForSingleObject(h, 0)
	return ev == waitTimeout
}

func commandNames(pid int, dir string) bool {
	h, err := syscall.OpenProcess(leftovers.QueryLimited, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	line, ok := leftovers.CommandLine(h)
	return ok && strings.Contains(strings.ToLower(line), strings.ToLower(dir))
}

func endPID(pid int) {
	h, err := syscall.OpenProcess(processTerminate|syscall.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	_ = syscall.TerminateProcess(h, 1)
	_, _ = syscall.WaitForSingleObject(h, 5000)
}

// killRunner ends the helper runner as taskkill /f does.
func killRunner(pid int) error {
	tk := filepath.Join(os.Getenv("SystemRoot"), "System32", "taskkill.exe")
	out, err := exec.Command(tk, "/f", "/pid", fmt.Sprint(pid)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskkill: %v: %s", err, out)
	}
	return nil
}
