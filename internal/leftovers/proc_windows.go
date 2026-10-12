//go:build windows

package leftovers

// Windows: the process table from a snapshot (CreateToolhelp32Snapshot), and each process's image path, start time
// and command line read through a handle opened for limited query. The calls not in package syscall go through its
// lazily loaded procedures, so golang.org/x/sys stays out of the program (step 5.3's decision). These readers serve
// the leftovers check and rungjob, which describes a job's members with them.

import (
	"os"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	ntdll                          = syscall.NewLazyDLL("ntdll.dll")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	procNtQueryInformationProcess  = ntdll.NewProc("NtQueryInformationProcess")
)

// QueryLimited is the access right a description needs (PROCESS_QUERY_LIMITED_INFORMATION).
const QueryLimited = 0x1000

const (
	processCommandLineInformation = 60         // NtQueryInformationProcess's class for the command line (Windows 8.1 on)
	statusInfoLengthMismatch      = 0xC0000004 // NTSTATUS: the buffer is too small
	statusBufferTooSmall          = 0xC0000023
	statusBufferOverflow          = 0x80000005
)

// Created reads when a process started (its creation FILETIME), or 0 when it cannot be read.
func Created(h syscall.Handle) uint64 {
	var c, e, k, u syscall.Filetime
	if err := syscall.GetProcessTimes(h, &c, &e, &k, &u); err != nil {
		return 0
	}
	return uint64(c.HighDateTime)<<32 | uint64(c.LowDateTime)
}

// ImagePath reads a process's program path (QueryFullProcessImageNameW), or "".
func ImagePath(h syscall.Handle) string {
	buf := make([]uint16, 1024)
	for {
		n := uint32(len(buf))
		r, _, err := procQueryFullProcessImageNameW.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&n)))
		if r != 0 {
			return string(utf16.Decode(buf[:n]))
		}
		if err != syscall.ERROR_INSUFFICIENT_BUFFER || len(buf) >= 32768 {
			return ""
		}
		buf = make([]uint16, 32768)
	}
}

// unicodeString is the system's UNICODE_STRING: a length in bytes and a pointer to the text.
type unicodeString struct {
	Length, MaximumLength uint16
	Buffer                *uint16
}

// CommandLine reads a process's command line (NtQueryInformationProcess, a single call the system answers for a
// handle opened for limited query); false when it cannot.
func CommandLine(h syscall.Handle) (string, bool) {
	if procNtQueryInformationProcess.Find() != nil {
		return "", false
	}
	size := uint32(1024)
	for try := 0; try < 3; try++ {
		buf := make([]uintptr, (size+uint32(unsafe.Sizeof(uintptr(0)))-1)/uint32(unsafe.Sizeof(uintptr(0))))
		var ret uint32
		st, _, _ := procNtQueryInformationProcess.Call(uintptr(h), processCommandLineInformation,
			uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))*unsafe.Sizeof(uintptr(0)), uintptr(unsafe.Pointer(&ret)))
		switch uint32(st) {
		case 0:
			us := (*unicodeString)(unsafe.Pointer(&buf[0]))
			if us.Buffer == nil || us.Length == 0 {
				return "", true
			}
			s := string(utf16.Decode(unsafe.Slice(us.Buffer, us.Length/2)))
			return s, true
		case statusInfoLengthMismatch, statusBufferTooSmall, statusBufferOverflow:
			if ret <= size || ret > 1<<20 {
				return "", false
			}
			size = ret
		default:
			return "", false
		}
	}
	return "", false
}

// Describe reads a process's program path and its arguments (nil when the command line cannot be read).
func Describe(h syscall.Handle) (image string, argv []string) {
	image = ImagePath(h)
	if line, ok := CommandLine(h); ok {
		argv = SplitCommandLine(line)
	}
	return image, argv
}

// BaseName gives a Windows path's file name.
func BaseName(p string) string {
	if i := strings.LastIndexAny(p, `\/`); i >= 0 {
		return p[i+1:]
	}
	return p
}

// snapshot reads every process with its program, start time and arguments where this user may query it.
func snapshot() ([]Row, style, error) {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, windowsStyle, os.NewSyscallError("CreateToolhelp32Snapshot", err)
	}
	defer func() { _ = syscall.CloseHandle(snap) }()
	var e syscall.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	if err := syscall.Process32First(snap, &e); err != nil {
		return nil, windowsStyle, os.NewSyscallError("Process32First", err)
	}
	var rows []Row
	for {
		r := Row{PID: int(e.ProcessID), PPID: int(e.ParentProcessID), Name: syscall.UTF16ToString(e.ExeFile[:])}
		if e.ProcessID != 0 {
			if h, err := syscall.OpenProcess(QueryLimited, false, e.ProcessID); err == nil {
				r.Start = Created(h)
				r.Exe, r.Argv = Describe(h)
				_ = syscall.CloseHandle(h)
			}
		}
		rows = append(rows, r)
		if err := syscall.Process32Next(snap, &e); err != nil {
			if err == syscall.ERROR_NO_MORE_FILES {
				break
			}
			return nil, windowsStyle, os.NewSyscallError("Process32Next", err)
		}
	}
	return rows, windowsStyle, nil
}

func nativeStyle() style { return windowsStyle }

func selfPID() int { return os.Getpid() }
