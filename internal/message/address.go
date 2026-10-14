package message

// An address (contract §22's from and to): <role> for the role the person's own session plays; <role>@<task> for an
// agent of that role on that task; <role>@<task>#<n> for the nth agent of one role on one task (#2 for the second).
// Its form is bonsai.message/1's pattern; its recipient folder under .bonsai/local/messages/ is the address with its @
// written as - (builder@T-0901's is builder-T-0901), a name Windows can hold: lower-case letters, digits, dashes, an
// upper-case T and one #. Beside the pattern, an address is refused where its folder would be a name Windows keeps for
// a device (con, prn, aux, nul, com0 to com9, lpt0 to lpt9), and past MaxAddress characters, so every folder a message
// is written to is one every system can make.
//
// Folders whose names differ only in case are one folder on Windows (and on any system where case does not count in
// names): a role's address ending in -t- and a task's digits (builder-t-0901) has the folder of a task's address
// (builder@T-0901's builder-T-0901) but for the case of its T. Such an address is refused, on every system alike, so
// no two addresses ever share a folder and a message never lies in another address's folder (chosen over comparing
// folder names case-folded, which would make the reader's rule differ by system and still put two addresses' messages
// in one folder on Windows).

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// MaxAddress is the longest address bonsai message writes, in characters: its folder and a message's file name stay
// well inside the paths every system holds.
const MaxAddress = 100

var addressPattern = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`^[a-z][a-z0-9-]*(@T-[0-9]{4,6}(#[1-9][0-9]*)?)?$`)
})

// idPattern is a message's id (bonsai.message/1's): M-, the UTC date, a dash and 8 lower-case hex characters.
var idPattern = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`^M-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9a-f]{8}$`)
})

// foldsOnto is a role's address whose folder, case aside, is a task's address's: the role, then -t- and a task's
// digits, at its end.
var foldsOnto = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^([a-z][a-z0-9-]*)-t-([0-9]{4,6})$`) })

// taskPattern is a task id (bonsai.task/1's, as bonsai.message/1's task holds it).
var taskPattern = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(`^T-[0-9]{4,6}$`) })

// deviceNames are the names Windows keeps for devices, which no file or folder may take.
var deviceNames = map[string]bool{"con": true, "prn": true, "aux": true, "nul": true}

func init() {
	for _, d := range "0123456789" {
		deviceNames["com"+string(d)] = true
		deviceNames["lpt"+string(d)] = true
	}
}

// CheckAddress says why addr is not an address bonsai message writes to or from, "" when it is one.
func CheckAddress(addr string) string {
	switch {
	case addr == "":
		return "is empty"
	case len(addr) > MaxAddress:
		return "is longer than " + strconv.Itoa(MaxAddress) + " characters"
	case !addressPattern().MatchString(addr):
		return "is not <role>, <role>@<task> or <role>@<task>#<n> (a role of lower-case letters, digits and dashes; a task T- and 4 to 6 digits; n from 1)"
	case deviceNames[FolderOf(addr)]:
		return "names a folder Windows keeps for a device (" + FolderOf(addr) + "), which no system may hold"
	}
	if m := foldsOnto().FindStringSubmatch(addr); m != nil {
		return "has the folder " + addr + ", which is " + m[1] + "@T-" + m[2] + "'s folder (" + m[1] + "-T-" + m[2] + ") where the case of " +
			"a name does not count, as on Windows; to address a task's agent, write " + m[1] + "@T-" + m[2]
	}
	return ""
}

// RoleOf gives an address's role: the part before its @.
func RoleOf(addr string) string {
	role, _, _ := strings.Cut(addr, "@")
	return role
}

// FolderOf gives an address's recipient folder: its @ written as -.
func FolderOf(addr string) string { return strings.Replace(addr, "@", "-", 1) }

// AddressOf gives the address a recipient folder holds messages for, and false when the folder's name is not one an
// address gives (a role holds no upper-case letter, so its -T- is the address's @).
func AddressOf(folder string) (string, bool) {
	addr := folder
	if i := strings.LastIndex(folder, "-T-"); i > 0 {
		addr = folder[:i] + "@" + folder[i+1:]
	}
	if CheckAddress(addr) != "" || FolderOf(addr) != folder {
		return "", false
	}
	return addr, true
}

// IsID reports whether s is a message's id.
func IsID(s string) bool { return idPattern().MatchString(s) }

// IsTask reports whether s is a task id a message may name.
func IsTask(s string) bool { return taskPattern().MatchString(s) }
