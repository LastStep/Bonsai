//go:build linux

package leftovers

// Linux: the process table from /proc, read-only. These readers serve the leftovers check and rungjob, which lists a
// rung's processes with them.

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"strings"
)

// Stat is one process's /proc/<pid>/stat, the fields the ladder reads.
type Stat struct {
	PID, PPID int
	PGID, SID int    // its process group and session
	Name      string // the kernel's short name (comm)
	State     byte   // R, S, D, Z (a zombie: it has ended and waits for its parent to reap it), X, ...
	Start     uint64 // when it started, in clock ticks since boot: with the pid, it tells a process from a later one
}

// Dead reports a process that has ended (a zombie, or one being reaped).
func (s Stat) Dead() bool { return s.State == 'Z' || s.State == 'X' }

// ParseStat reads one /proc/<pid>/stat line. The name sits in parentheses and may hold spaces and parentheses
// itself, so the fields after it are read from the last ')'.
func ParseStat(line string) (Stat, bool) {
	open := strings.IndexByte(line, '(')
	end := strings.LastIndexByte(line, ')')
	if open < 1 || end < open || end+2 > len(line) {
		return Stat{}, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(line[:open]))
	if err != nil || pid <= 0 {
		return Stat{}, false
	}
	f := strings.Fields(line[end+2:])
	// f[0] is field 3 (state); ppid is 4, pgrp 5, session 6, starttime 22.
	if len(f) < 20 || len(f[0]) == 0 {
		return Stat{}, false
	}
	s := Stat{PID: pid, Name: line[open+1 : end], State: f[0][0]}
	var e1, e2, e3, e4 error
	s.PPID, e1 = strconv.Atoi(f[1])
	s.PGID, e2 = strconv.Atoi(f[2])
	s.SID, e3 = strconv.Atoi(f[3])
	s.Start, e4 = strconv.ParseUint(f[19], 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return Stat{}, false
	}
	return s, true
}

// ReadStat reads one process's stat; false when it is gone.
func ReadStat(pid int) (Stat, bool) {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return Stat{}, false
	}
	return ParseStat(string(b))
}

// Stats reads every process's stat, in /proc's order (by pid).
func Stats() ([]Stat, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	names, err := d.Readdirnames(-1)
	_ = d.Close()
	if err != nil {
		return nil, err
	}
	out := make([]Stat, 0, len(names))
	for _, n := range names {
		pid, err := strconv.Atoi(n)
		if err != nil || pid <= 0 {
			continue
		}
		if s, ok := ReadStat(pid); ok {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("/proc lists no process")
	}
	return out, nil
}

// Argv reads a process's arguments exactly (NUL-separated in /proc); nil for a kernel thread, a zombie, a gone
// process or one this user may not read.
func Argv(pid int) []string {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil || len(b) == 0 {
		return nil
	}
	b = bytes.TrimSuffix(b, []byte{0})
	return strings.Split(string(b), "\x00")
}

// EnvValue reads one variable from the environment a process was started with (/proc/<pid>/environ); false when it
// has none, or it cannot be read (only this user's processes can, and a zombie's is empty).
func EnvValue(pid int, name string) (string, bool) {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/environ")
	if err != nil {
		return "", false
	}
	key := []byte(name + "=")
	for len(b) > 0 {
		entry := b
		if i := bytes.IndexByte(b, 0); i >= 0 {
			entry, b = b[:i], b[i+1:]
		} else {
			b = nil
		}
		if bytes.HasPrefix(entry, key) {
			return string(entry[len(key):]), true
		}
	}
	return "", false
}

// link reads one of a process's links (exe, cwd), or "" (another user's process, a kernel thread, a gone one).
func link(pid int, name string) string {
	p, err := os.Readlink("/proc/" + strconv.Itoa(pid) + "/" + name)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(p, " (deleted)")
}

// snapshot reads every live process with what ties it to a folder.
func snapshot() ([]Row, style, error) {
	stats, err := Stats()
	if err != nil {
		return nil, posixStyle, err
	}
	rows := make([]Row, 0, len(stats))
	for _, s := range stats {
		if s.Dead() {
			continue
		}
		rows = append(rows, Row{PID: s.PID, PPID: s.PPID, Name: s.Name, Start: s.Start, Exe: link(s.PID, "exe"),
			Cwd: link(s.PID, "cwd"), Argv: Argv(s.PID)})
	}
	return rows, posixStyle, nil
}

func nativeStyle() style { return posixStyle }

func selfPID() int { return os.Getpid() }
