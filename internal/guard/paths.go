package guard

// What path an edit really writes, in every project-relative form it may take, and the glob match against
// bonsai.yaml's lists.
//
// The guard judges a path in all its forms and refuses when any one of them is on a list (fail closed):
//   - as written, made absolute against the payload's cwd (else the project's folder) and cleaned;
//   - with symbolic links resolved as far as the path exists (a link to a protected file is the protected file),
//     and a link whose target does not exist yet followed, since a write through it creates that target;
//   - on Windows, also as Windows itself reduces a name: the \\?\ prefix dropped, Git Bash's /c/... read as C:\...,
//     a stream after a colon cut off (protected.txt::$DATA is protected.txt), and the dots and spaces Windows drops
//     from the end of a name dropped (protected.txt. is protected.txt); resolving an existing path there also gives
//     its long name and its letter case on disk;
// each against the project's folder as given and as resolved. Letter case is folded where the project's folder is
// case-insensitive: always on Windows and macOS, and on Linux when bonsai.yaml answers to BONSAI.YAML too (a folder
// on /mnt/c in WSL).
//
// Not covered, the limits of a tripwire until step 5.3: a hard link made to a protected file, a short (8.3) name of
// a file that does not exist yet, and another route to the same disk (a network share of it).

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// projectPaths gives every project-relative form of an edit's path p (forward slashes, each once, in a fixed
// order: as written first), or none when p lies outside the project in every form. root is the project's absolute
// folder; cwd the payload's.
func projectPaths(root, cwd, p string) []string {
	roots := []string{filepath.Clean(root)}
	if real, err := filepath.EvalSymlinks(root); err == nil {
		roots = appendNew(roots, real)
	}
	written := []string{p}
	if runtime.GOOS == "windows" {
		written = append(written, windowsForms(p)...)
	}
	var targets []string
	for _, w := range written {
		a := absolute(cwd, root, w)
		targets = appendNew(targets, a)
		targets = appendNew(targets, resolve(a))
	}
	var out []string
	for _, t := range targets {
		for _, r := range roots {
			rel, err := filepath.Rel(r, t)
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)
			if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
				continue
			}
			out = appendNew(out, rel)
		}
	}
	return out
}

// absolute makes a written path absolute and clean: against cwd when it is absolute, else against the project.
func absolute(cwd, root, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	base := root
	if cwd != "" && filepath.IsAbs(cwd) {
		base = cwd
	}
	if runtime.GOOS == "windows" {
		vol := filepath.VolumeName(p)
		switch {
		case vol == "" && (strings.HasPrefix(p, `\`) || strings.HasPrefix(p, "/")):
			// \x is x at the root of the current drive.
			return filepath.Clean(filepath.VolumeName(base) + p)
		case vol != "":
			// C:x is x in drive C's current folder: the base's, when the base is on C.
			if strings.EqualFold(vol, filepath.VolumeName(base)) {
				return filepath.Join(base, p[len(vol):])
			}
			return filepath.Join(vol+`\`, p[len(vol):])
		}
	}
	return filepath.Join(base, p)
}

// resolve gives the path a write to a would reach: the symbolic links in the part of a that exists resolved, a link
// whose target does not exist yet followed, the rest kept as written. It gives a itself when nothing resolves.
func resolve(a string) string {
	var rest []string
	cur := a
	for hops := 0; hops < 64; hops++ {
		if real, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(append([]string{real}, rest...)...)
		}
		if fi, err := os.Lstat(cur); err == nil && fi.Mode()&fs.ModeSymlink != 0 {
			if t, err := os.Readlink(cur); err == nil {
				if !filepath.IsAbs(t) {
					t = filepath.Join(filepath.Dir(cur), t)
				}
				cur = filepath.Clean(t)
				continue
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		rest = append([]string{filepath.Base(cur)}, rest...)
		cur = parent
	}
	return a
}

// windowsForms gives the other forms Windows reads a written path as (the package comment lists them). It works on
// the string alone, so its tests run on any system.
func windowsForms(p string) []string {
	q := strings.ReplaceAll(p, "/", `\`)
	for _, pre := range []string{`\\?\`, `\\.\`, `\??\`} {
		if strings.HasPrefix(q, pre) {
			rest := q[len(pre):]
			if len(rest) >= 4 && strings.EqualFold(rest[:4], `UNC\`) {
				q = `\\` + rest[4:]
			} else {
				q = rest
			}
			break
		}
	}
	forms := []string{q}
	s := strings.ReplaceAll(p, `\`, "/")
	if len(s) >= len("/cygdrive/") && strings.EqualFold(s[:len("/cygdrive/")], "/cygdrive/") {
		s = s[len("/cygdrive"):]
	}
	if len(s) >= 2 && s[0] == '/' && isLetter(s[1]) && (len(s) == 2 || s[2] == '/') {
		forms = append(forms, strings.ToUpper(s[1:2])+`:\`+strings.ReplaceAll(strings.TrimPrefix(s[2:], "/"), "/", `\`))
	}
	written := append([]string(nil), forms...)
	for _, f := range written {
		forms = appendNew(forms, win32Names(f))
	}
	return forms
}

// win32Names reduces each name in a Windows path as Windows does when it opens it: the dots and spaces at the end of
// a name dropped, and in the last name a stream after a colon cut off. The volume (C: or \\server\share) is kept.
func win32Names(p string) string {
	vol := ""
	switch {
	case len(p) >= 2 && isLetter(p[0]) && p[1] == ':':
		vol = p[:2]
	case strings.HasPrefix(p, `\\`):
		parts := strings.SplitN(p[2:], `\`, 3)
		if len(parts) >= 2 {
			vol = `\\` + parts[0] + `\` + parts[1]
		}
	}
	segs := strings.Split(p[len(vol):], `\`)
	for i, s := range segs {
		if i == len(segs)-1 {
			if j := strings.IndexByte(s, ':'); j >= 0 {
				s = s[:j]
			}
		}
		if s != "." && s != ".." {
			s = strings.TrimRight(s, ". ")
		}
		segs[i] = s
	}
	return vol + strings.Join(segs, `\`)
}

func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

// caseInsensitive reports whether names in the project's folder ignore letter case.
func caseInsensitive(root string) bool {
	switch runtime.GOOS {
	case "windows", "darwin", "ios":
		return true
	}
	a, err := os.Stat(filepath.Join(root, workspace.ConfigFile))
	if err != nil {
		return false
	}
	b, err := os.Stat(filepath.Join(root, strings.ToUpper(workspace.ConfigFile)))
	return err == nil && os.SameFile(a, b)
}

// checkGlob reports a glob the guard cannot read: one whose segments path.Match refuses (an unclosed [ among them).
func checkGlob(glob string) error {
	for _, seg := range strings.Split(glob, "/") {
		if seg == "**" {
			continue
		}
		if _, err := path.Match(seg, ""); err != nil {
			return err
		}
	}
	return nil
}

// matchGlob reports whether a project-relative path is on a glob of bonsai.yaml's lists. Segments are matched as
// path.Match does (* and ? within a name, [...] a class); ** stands for any number of whole segments, none
// included; a glob ending in / means everything inside that folder. A glob that names a folder holding the path
// matches it too: protecting a folder protects what is in it. With fold, letter case is ignored. Call checkGlob
// first: a glob it refuses never matches here.
func matchGlob(glob, rel string, fold bool) bool {
	if fold {
		glob, rel = strings.ToLower(glob), strings.ToLower(rel)
	}
	if strings.HasSuffix(glob, "/") {
		glob += "**"
	}
	return matchSegs(strings.Split(glob, "/"), strings.Split(rel, "/"))
}

func matchSegs(g, r []string) bool {
	if len(g) == 0 {
		return true // the glob is used up: it named the path, or a folder holding it
	}
	if g[0] == "**" {
		for i := 0; i <= len(r); i++ {
			if matchSegs(g[1:], r[i:]) {
				return true
			}
		}
		return false
	}
	if len(r) == 0 {
		return false
	}
	if ok, err := path.Match(g[0], r[0]); err != nil || !ok {
		return false
	}
	return matchSegs(g[1:], r[1:])
}

// appendNew appends s unless the list holds it already.
func appendNew(list []string, s string) []string {
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}
