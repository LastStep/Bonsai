package guard

// Bonsai's own files outside a project (design/plan-5.md, 5.3.2 note 5), refused in every mode and never granted:
//
//   - anything in Bonsai's home (bonsai-files): BONSAI_HOME as the hook sees it, else ~/.bonsai, and ~/.bonsai as well
//     when the two differ. Its machine records are what the guard trusts to verify a main checkout (5.3.1 note 2), and
//     its settings hold status_writes, so one write there could turn command mode's tripwires off. The engine's deny
//     rule (step 5.3.5) stops a shell redirect into ~/.bonsai; a home moved by BONSAI_HOME is guarded by the guard
//     alone.
//     On Linux, a Windows user's Bonsai home as WSL mounts it (/mnt/<drive>/Users/<name>/.bonsai) counts as Bonsai's
//     home too: Bonsai's Windows side trusts its machine records, and a WSL session can reach it (wslWindowsHome).
//   - outside every linked project, a file named as Bonsai's program (bonsai-stand-in, StandInNames), so the Write
//     tool cannot place a stand-in for Bonsai in a folder on the PATH, under every answer to the hook-line question.
//
// Bonsai writes its own files as a program, not through the file tools, so neither refusal stops Bonsai. Inside a
// project, Bonsai's own files are the floor's bonsai-files entries (floor.go).

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// StandInNames are the file names a stand-in for Bonsai's program would take (5.3.2 note 5), compared with letter case
// ignored: the list's one home, on the reference page. bonsai.com is added to the plan's five: Windows runs a .com
// before an .exe of the same name.
var StandInNames = []string{"bonsai", "bonsai.exe", "bonsai.cmd", "bonsai.bat", "bonsai.ps1", "bonsai.com"}

// standIn reports whether a path's last name is one of StandInNames.
func standIn(abs string) bool {
	base := filepath.Base(abs)
	if runtime.GOOS == "windows" {
		base = win32Names(base)
	}
	for _, n := range StandInNames {
		if strings.EqualFold(base, n) {
			return true
		}
	}
	return false
}

// homes are Bonsai's home folders the guard refuses writes into, each in every form a path into it may take: as given,
// with its links resolved, and on Windows as Windows names it (finalPath). Read once a run.
func (j *judge) homes() [][]string {
	if j.homeForms != nil {
		return j.homeForms
	}
	var dirs []string
	if j.home != "" {
		dirs = append(dirs, filepath.Clean(j.home))
	}
	if user, err := os.UserHomeDir(); err == nil && user != "" {
		dirs = appendNew(dirs, filepath.Join(user, ".bonsai"))
	}
	j.homeForms = [][]string{}
	for _, d := range dirs {
		forms := []string{d}
		forms = appendNew(forms, resolve(d))
		if f, ok, err := finalPath(d); err == nil && ok {
			forms = appendNew(forms, f)
		}
		j.homeForms = append(j.homeForms, forms)
	}
	return j.homeForms
}

// inHome gives the home folder (as given) holding abs, or "" for none. Letter case is ignored on every system: a path
// that differs from the home only in case is refused too, which costs nothing where names keep their case.
func (j *judge) inHome(abs string) string {
	for _, forms := range j.homes() {
		for _, h := range forms {
			if within(h, abs) {
				return forms[0]
			}
		}
	}
	return wslWindowsHome(abs, runtime.GOOS)
}

// wslWindowsHome gives, on Linux, the Windows user's Bonsai home a path lies in as WSL mounts it,
// /mnt/<drive>/Users/<name>/.bonsai (letter case ignored, as Windows ignores it), or "" for any other path or system.
func wslWindowsHome(abs, goos string) string {
	if goos != "linux" {
		return ""
	}
	segs := strings.Split(filepath.ToSlash(filepath.Clean(abs)), "/")
	if len(segs) < 6 || segs[0] != "" || segs[1] != "mnt" || len(segs[2]) != 1 || !isLetter(segs[2][0]) ||
		!strings.EqualFold(segs[3], "Users") || segs[4] == "" || !strings.EqualFold(segs[5], ".bonsai") {
		return ""
	}
	return strings.Join(segs[:6], "/")
}

// within reports whether abs is the folder root or lies inside it, letter case ignored.
func within(root, abs string) bool {
	r, a := strings.ToLower(filepath.Clean(root)), strings.ToLower(filepath.Clean(abs))
	if a == r {
		return true
	}
	if !strings.HasSuffix(r, string(filepath.Separator)) {
		r += string(filepath.Separator)
	}
	return strings.HasPrefix(a, r)
}
