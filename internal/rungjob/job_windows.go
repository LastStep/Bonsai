//go:build windows

package rungjob

// Windows: the rung's job object (the package doc). Every call not in package syscall goes through its lazily
// loaded procedures.

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/LastStep/Bonsai/internal/leftovers"
)

var (
	kernel32                      = syscall.NewLazyDLL("kernel32.dll")
	ntdll                         = syscall.NewLazyDLL("ntdll.dll")
	procCreateJobObjectW          = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject   = kernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject  = kernel32.NewProc("AssignProcessToJobObject")
	procQueryInformationJobObject = kernel32.NewProc("QueryInformationJobObject")
	procTerminateJobObject        = kernel32.NewProc("TerminateJobObject")
	procIsProcessInJob            = kernel32.NewProc("IsProcessInJob")
	procNtResumeProcess           = ntdll.NewProc("NtResumeProcess")
)

const (
	createSuspended = 0x00000004 // CREATE_SUSPENDED

	processTerminate     = 0x0001 // PROCESS_TERMINATE
	processSetQuota      = 0x0100 // PROCESS_SET_QUOTA
	processSuspendResume = 0x0800 // PROCESS_SUSPEND_RESUME

	jobObjectBasicProcessIDList       = 3      // JobObjectBasicProcessIdList
	jobObjectExtendedLimitInformation = 9      // JobObjectExtendedLimitInformation
	jobObjectLimitKillOnJobClose      = 0x2000 // JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE; no BREAKAWAY_OK flag, so none may leave
	errorMoreData                     = syscall.Errno(234)

	// endExitCode is the exit code TerminateJobObject gives the job's processes.
	endExitCode = 1
)

// basicLimit is JOBOBJECT_BASIC_LIMIT_INFORMATION.
type basicLimit struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

// extendedLimit is JOBOBJECT_EXTENDED_LIMIT_INFORMATION.
type extendedLimit struct {
	Basic                 basicLimit
	IoInfo                [6]uint64 // IO_COUNTERS
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

// makeJob makes the rung's job: unnamed, with null security attributes (so its handle is not inherited), ending
// every member when its last handle closes, with no breakaway. A test stands in a failing one.
var makeJob = func() (syscall.Handle, error) {
	h, _, err := procCreateJobObjectW.Call(0, 0)
	if h == 0 {
		return 0, fmt.Errorf("CreateJobObjectW: %w", err)
	}
	var info extendedLimit
	info.Basic.LimitFlags = jobObjectLimitKillOnJobClose
	if ok, _, err := procSetInformationJobObject.Call(h, jobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info)); ok == 0 {
		_ = syscall.CloseHandle(syscall.Handle(h))
		return 0, fmt.Errorf("SetInformationJobObject: %w", err)
	}
	return syscall.Handle(h), nil
}

// rung is one rung's job on Windows.
type rung struct {
	handle syscall.Handle // the job; 0 when off
	off    string         // why the job is off
	shell  int
	setup  time.Duration
	notes  []string
}

func newRung() *rung { return &rung{} }

func (r *rung) job() string {
	if r.off != "" {
		return "off: " + r.off
	}
	return "on"
}

func (r *rung) setupCost() time.Duration { return r.setup }

// prepare makes the job; if it cannot be made, the rung runs without one.
func (r *rung) prepare(env []string) []string {
	t := time.Now()
	h, err := makeJob()
	if err != nil {
		r.off = oneLine(err.Error())
	} else {
		r.handle = h
	}
	r.setup += time.Since(t)
	return env
}

// command builds today's line, %ComSpec% /d /s /c "<command>", started suspended when there is a job.
func (r *rung) command(command, dir string, env []string) *exec.Cmd {
	comspec := getEnv(env, "ComSpec")
	if comspec == "" || !filepath.IsAbs(comspec) {
		root := getEnv(env, "SystemRoot")
		if root == "" {
			root = os.Getenv("SystemRoot")
		}
		if root == "" {
			root = `C:\Windows`
		}
		comspec = filepath.Join(root, "System32", "cmd.exe")
	}
	cmd := exec.Command(comspec)
	cmd.Dir = dir
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:    `"` + comspec + `" /d /s /c "` + command + `"`,
		HideWindow: true,
	}
	if r.handle != 0 {
		cmd.SysProcAttr.CreationFlags |= createSuspended
	}
	return cmd
}

// launch starts the shell, suspended, puts it in the job and lets it run.
func (r *rung) launch(cmd *exec.Cmd) (<-chan error, error) {
	if err := cmd.Start(); err != nil {
		if r.handle != 0 {
			_ = syscall.CloseHandle(r.handle)
			r.handle = 0
		}
		return nil, err
	}
	r.shell = cmd.Process.Pid
	if r.handle != 0 {
		t := time.Now()
		if err := r.enter(uint32(r.shell)); err != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			_ = syscall.CloseHandle(r.handle)
			r.handle = 0
			return nil, err
		}
		r.setup += time.Since(t)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	return exited, nil
}

// enter puts the suspended shell in the job and resumes it. The shell cannot have run, so its pid is its own (os/exec
// holds its handle). A shell that cannot join the job runs without one, and the job is off; one that cannot be let
// run is an error, and the caller ends it.
func (r *rung) enter(pid uint32) error {
	h, err := syscall.OpenProcess(processSetQuota|processTerminate|processSuspendResume|leftovers.QueryLimited,
		false, pid)
	if err != nil {
		return fmt.Errorf("could not open the suspended shell: OpenProcess: %w", err)
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	if ok, _, err := procAssignProcessToJobObject.Call(uintptr(r.handle), uintptr(h)); ok == 0 {
		r.off = oneLine("AssignProcessToJobObject: " + err.Error())
		_ = syscall.CloseHandle(r.handle)
		r.handle = 0
	}
	if err := procNtResumeProcess.Find(); err != nil {
		return fmt.Errorf("could not resume the suspended shell: %w", err)
	}
	if st, _, _ := procNtResumeProcess.Call(uintptr(h)); st != 0 {
		return fmt.Errorf("could not resume the suspended shell: NtResumeProcess: NTSTATUS 0x%08X", uint32(st))
	}
	return nil
}

// end lists the job's members, ends the job, and waits until it is empty (EndWait at most). Without a job, only the
// shell is ended, at a timeout or a stop.
func (r *rung) end(cmd *exec.Cmd, alive bool) ([]Process, []string) {
	if r.handle == 0 {
		if !alive {
			return nil, nil
		}
		ended := []Process{describe(uint32(r.shell), 0)}
		_ = cmd.Process.Kill()
		return ended, []string{"job off: only the rung's shell was ended; what it started is not known"}
	}
	var notes []string
	ended := []Process{}
	pids, err := r.members()
	if err != nil {
		notes = append(notes, "could not list the job's members: "+oneLine(err.Error()))
	}
	for _, pid := range pids {
		if p, ok := r.member(pid); ok {
			ended = append(ended, p)
		}
	}
	if ok, _, err := procTerminateJobObject.Call(uintptr(r.handle), endExitCode); ok == 0 {
		notes = append(notes, "could not end the job: TerminateJobObject: "+oneLine(err.Error()))
	}
	deadline := time.Now().Add(EndWait)
	for {
		left, err := r.members()
		if err != nil || len(left) == 0 {
			break
		}
		if time.Now().After(deadline) {
			notes = append(notes, fmt.Sprintf("%d of the job's processes still alive %s s after its end", len(left),
				seconds(EndWait)))
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	return ended, notes
}

// members lists the pids in the job (nested jobs' members among them).
func (r *rung) members() ([]uint32, error) {
	ptr := unsafe.Sizeof(uintptr(0))
	n := 64
	for {
		// JOBOBJECT_BASIC_PROCESS_ID_LIST: two ULONGs, then the pids as ULONG_PTRs from byte 8.
		buf := make([]uintptr, (8+uintptr(n)*ptr+ptr-1)/ptr)
		ok, _, err := procQueryInformationJobObject.Call(uintptr(r.handle), jobObjectBasicProcessIDList,
			uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))*ptr, 0)
		if ok == 0 {
			if errors.Is(err, errorMoreData) && n < 1<<16 {
				n *= 4
				continue
			}
			return nil, fmt.Errorf("QueryInformationJobObject: %w", err)
		}
		base := unsafe.Pointer(&buf[0])
		count := *(*uint32)(unsafe.Add(base, 4))
		out := make([]uint32, 0, count)
		for i := uintptr(0); i < uintptr(count); i++ {
			out = append(out, uint32(*(*uintptr)(unsafe.Add(base, 8+i*ptr))))
		}
		return out, nil
	}
}

// member describes one listed pid, checked to be in this job (a pid that left and was taken since is not).
func (r *rung) member(pid uint32) (Process, bool) {
	h, err := syscall.OpenProcess(leftovers.QueryLimited, false, pid)
	if err != nil {
		if goneProcess(err) {
			return Process{}, false
		}
		return Process{PID: int(pid)}, true
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	var in int32
	if ok, _, _ := procIsProcessInJob.Call(uintptr(h), uintptr(r.handle), uintptr(unsafe.Pointer(&in))); ok != 0 && in == 0 {
		return Process{}, false
	}
	return describe(pid, h), true
}

// errorInvalidParameter is what OpenProcess gives for a pid that no process holds any more (it exited between the
// job's listing and the open): ERROR_INVALID_PARAMETER.
const errorInvalidParameter = syscall.Errno(87)

// goneProcess reports an OpenProcess error that means the process is gone: such a pid is no member, so it is skipped.
// Any other error (access denied and the rest) leaves a process that is there, which is still reported.
func goneProcess(err error) bool { return errors.Is(err, errorInvalidParameter) }

// describe names a process by its handle, or opens one for it when h is 0.
func describe(pid uint32, h syscall.Handle) Process {
	if h == 0 {
		var err error
		if h, err = syscall.OpenProcess(leftovers.QueryLimited, false, pid); err != nil {
			return Process{PID: int(pid)}
		}
		defer func() { _ = syscall.CloseHandle(h) }()
	}
	image, argv := leftovers.Describe(h)
	return Process{PID: int(pid), Name: leftovers.BaseName(image), Head: leftovers.HeadOf(argv)}
}

// finish closes the job's handle, its last: kill-on-close ends whatever is still in it.
func (r *rung) finish() []string {
	if r.handle != 0 {
		_ = syscall.CloseHandle(r.handle)
		r.handle = 0
	}
	return r.notes
}

// signalNumber is 0 on Windows: no process ends by a signal.
func signalNumber(*os.ProcessState) int { return 0 }
