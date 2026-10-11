package guard

// Where a delete's operand lies, for the delete check (plan-5 5.3.3 notes 3 and 4): whether it is a plain name
// (rule 3), the folder a command runs in (the cd rules), and for a recursive delete or a move of a folder the
// guarded paths under it, found from the guarded globs without reading the folder's tree (rule 4).
//
// The tree is never read: a recursive delete costs the same whether the folder holds ten files or ten thousand. A
// guarded glob with a fixed start (its first segment holds no wildcard: .git, docs/secret/**) is met against the
// operand segment by segment; where it can match a path at or under the operand (or names a folder holding it),
// the meeting gives one such path, which the guard's function then judges, so a running task's grant counts as it
// does for a file tool. A glob with no fixed start (**/bonsai.yaml) is judged against the operand itself only,
// never against what a deleted folder holds (note 4): reading the tree would cost time on every recursive delete,
// and past any cap would refuse rm -rf node_modules. What that loses: a file guarded only by such a glob inside a
// folder deleted by its name is not seen (the deletion shows in rung 0's diff).

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// delDirs is where a command's relative paths are read: the folders it may run in (more than one when a cd before
// it may have failed), and whether one of them cannot be told.
type delDirs struct {
	known   []string
	unknown bool
}

// delMaxDirs caps the folders followed; past it the folder counts as one that cannot be told.
const delMaxDirs = 8

func (a delDirs) union(b delDirs) delDirs {
	out := delDirs{unknown: a.unknown || b.unknown}
	for _, d := range append(append([]string(nil), a.known...), b.known...) {
		if len(out.known) >= delMaxDirs {
			out.unknown = true
			break
		}
		out.known = appendNew(out.known, d)
	}
	return out
}

// delHome is the user's home folder, "" when there is none.
func delHome() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// delDriveForm reads a Git Bash or WSL path to a Windows drive, /c/x, /cygdrive/c/x or /mnt/c/x, as C:\x; "" for
// any other path. It is used on Windows only.
func delDriveForm(p string) string {
	s := strings.ReplaceAll(p, `\`, "/")
	for _, pre := range []string{"/cygdrive/", "/mnt/", "/"} {
		if len(s) < len(pre)+1 || !strings.EqualFold(s[:len(pre)], pre) {
			continue
		}
		rest := s[len(pre):]
		if isLetter(rest[0]) && (len(rest) == 1 || rest[1] == '/') {
			return strings.ToUpper(rest[:1]) + `:\` + strings.ReplaceAll(strings.TrimPrefix(rest[1:], "/"), "/", `\`)
		}
	}
	return ""
}

// delTildeUser reads bash's other tilde forms: ~name/x as x in the home folder when name is its folder's name (the
// current user's), ~+/x as x in the folder the command runs in; "" for any other path.
func delTildeUser(dir, p string) string {
	if !strings.HasPrefix(p, "~") || p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		return ""
	}
	name, rest, _ := strings.Cut(p[1:], "/")
	switch name {
	case "+":
		return filepath.Join(dir, rest)
	case "", "-":
		return ""
	}
	if h := delHome(); h != "" && name == filepath.Base(h) {
		return filepath.Join(h, rest)
	}
	return ""
}

// delAbs makes an operand absolute in the folder dir, in its one most likely form: a ~ path in the home folder, a
// Git Bash or WSL drive path as Windows names it, any other against dir.
func delAbs(dir, p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		if h := delHome(); h != "" {
			return filepath.Join(h, p[1:])
		}
	}
	if runtime.GOOS == "windows" {
		if f := delDriveForm(p); f != "" {
			return filepath.Clean(f)
		}
	}
	return absolute(dir, dir, p)
}

// delForms gives every absolute form of the operand p read in the folder dir, each once and in a fixed order: as
// written, and as a path in the home folder when it starts with ~; on Windows also as Windows reads it
// (windowsForms, and WSL's /mnt/c/...); each also with its symbolic links resolved, and on Windows as Windows names
// it (finalPath). Its error says Windows opened a part of a path but could not name it.
func delForms(dir, p string) ([]string, error) {
	written := []string{p}
	if t := tildeForm(p); t != "" {
		written = append(written, t)
	}
	if t := delTildeUser(dir, p); t != "" {
		written = append(written, t)
	}
	if runtime.GOOS == "windows" {
		for _, w := range append([]string(nil), written...) {
			for _, f := range windowsForms(w) {
				written = appendNew(written, f)
			}
			if f := delDriveForm(w); f != "" {
				written = appendNew(written, f)
			}
		}
	}
	var out []string
	for _, w := range written {
		a := absolute(dir, dir, w)
		out = appendNew(out, a)
		out = appendNew(out, resolve(a))
		f, ok, err := finalPath(a)
		if err != nil {
			return nil, err
		}
		if ok {
			out = appendNew(out, f)
		}
	}
	return out, nil
}

// delRootForms gives a project's top folder in every form a path under it may be met in: as given, with its links
// resolved, and on Windows as Windows names it.
func delRootForms(root string) ([]string, error) {
	roots := []string{filepath.Clean(root)}
	if real, err := filepath.EvalSymlinks(root); err == nil {
		roots = appendNew(roots, real)
	}
	if f, ok, err := finalPath(root); err != nil {
		return nil, err
	} else if ok {
		roots = appendNew(roots, f)
	}
	return roots, nil
}

// delPlace places an operand, given in its absolute forms (delForms), against a project's top folder, given in its
// forms (delRootForms): the root-relative segments of each form that lies inside it, and whether a form is the root
// itself or a folder holding it (whole). With fold, letter case is ignored.
func delPlace(roots, forms []string, fold bool) (rels [][]string, whole bool) {
	seen := map[string]bool{}
	for _, t := range forms {
		for _, r := range roots {
			a, b := r, t
			if fold {
				a, b = strings.ToLower(a), strings.ToLower(b)
			}
			rel, err := filepath.Rel(a, b)
			if err != nil {
				continue
			}
			rel = filepath.ToSlash(rel)
			switch {
			case rel == ".":
				whole = true
			case rel == ".." || strings.HasPrefix(rel, "../"):
				if up, err := filepath.Rel(b, a); err == nil {
					up = filepath.ToSlash(up)
					if up != ".." && !strings.HasPrefix(up, "../") {
						whole = true // the operand holds the root
					}
				}
			default:
				if !seen[rel] {
					seen[rel] = true
					rels = append(rels, strings.Split(rel, "/"))
				}
			}
		}
	}
	return rels, whole
}

// delFixed reports whether a guarded glob has a fixed start: its first segment holds no wildcard.
func delFixed(glob string) bool {
	first, _, _ := strings.Cut(glob, "/")
	return first != "" && !strings.ContainsAny(first, "*?[")
}

// delGlobSegs splits a guarded glob into segments as matchGlob reads it: a trailing / means everything inside.
func delGlobSegs(glob string) []string {
	if strings.HasSuffix(glob, "/") {
		glob += "**"
	}
	return strings.Split(glob, "/")
}

// delMeet meets an operand (segments under a project's top folder, each a name or, for a glob operand, a pattern)
// with a guarded glob, and gives a path the operand takes that the glob guards, or false when there is none. With
// tree, the operand takes what it holds as well (a recursive delete, a move of a folder). The glob may also name a
// folder holding the operand, as matchGlob reads it. With fold, letter case is ignored.
func delMeet(op, glob []string, tree, fold bool) ([]string, bool) {
	var out []string
	budget := 4096 // the walk's steps: guarded globs are short, but a list of **s must not run long
	var walk func(p, g []string) bool
	walk = func(p, g []string) bool {
		budget--
		if budget < 0 {
			return false
		}
		if len(g) == 0 { // the glob is used up: it names the operand's path or a folder holding it
			for _, s := range p {
				x, ok := delSynth(s)
				if !ok {
					return false
				}
				out = append(out, x)
			}
			return true
		}
		if len(p) == 0 { // the operand is used up
			if !tree {
				for _, s := range g {
					if s != "**" {
						return false
					}
				}
				return true
			}
			for _, s := range g {
				if s == "**" {
					continue
				}
				x, ok := delSynth(s)
				if !ok {
					return false
				}
				out = append(out, x)
			}
			return true
		}
		if g[0] == "**" {
			for k := 0; k <= len(p); k++ {
				save := len(out)
				ok := true
				for _, s := range p[:k] {
					x, sok := delSynth(s)
					if !sok {
						ok = false
						break
					}
					out = append(out, x)
				}
				if ok && walk(p[k:], g[1:]) {
					return true
				}
				out = out[:save]
			}
			return false
		}
		seg, ok := delSegMeet(p[0], g[0], fold)
		if !ok {
			return false
		}
		save := len(out)
		out = append(out, seg)
		if walk(p[1:], g[1:]) {
			return true
		}
		out = out[:save]
		return false
	}
	if !walk(op, glob) {
		return nil, false
	}
	return out, true
}

// delSegMeet meets one segment of an operand with one of a guarded glob: a name both match, or false. When both
// are patterns and no simple name matches both, it gives true when they may still meet (their fixed beginnings and
// ends agree), with a name the glob matches: the check then judges that name, failing closed.
func delSegMeet(a, b string, fold bool) (string, bool) {
	ma, mb := a, b
	if fold {
		ma, mb = strings.ToLower(a), strings.ToLower(b)
	}
	litA := !strings.ContainsAny(a, "*?[")
	litB := !strings.ContainsAny(b, "*?[")
	match := func(pattern, name string) bool {
		ok, err := path.Match(pattern, name)
		if err != nil {
			return pattern == name
		}
		return ok
	}
	switch {
	case litA && litB:
		return a, ma == mb
	case litA:
		return a, match(mb, ma)
	case litB:
		return b, match(ma, mb)
	}
	for _, c := range []string{a, b} {
		if x, ok := delSynth(c); ok && match(ma, foldIf(x, fold)) && match(mb, foldIf(x, fold)) {
			return x, true
		}
	}
	pa, sa := delEdges(ma)
	pb, sb := delEdges(mb)
	if !(strings.HasPrefix(pa, pb) || strings.HasPrefix(pb, pa)) || !(strings.HasSuffix(sa, sb) || strings.HasSuffix(sb, sa)) {
		return "", false
	}
	x, ok := delSynth(b)
	if !ok {
		x = b
	}
	return x, true
}

func foldIf(s string, fold bool) string {
	if fold {
		return strings.ToLower(s)
	}
	return s
}

// delEdges gives a pattern's fixed beginning (up to its first wildcard) and fixed end (after its last).
func delEdges(p string) (string, string) {
	first := strings.IndexAny(p, "*?[")
	last := strings.LastIndexAny(p, "*?]")
	if first < 0 {
		return p, p
	}
	return p[:first], p[last+1:]
}

// delSynth gives a name that the pattern s matches (* as nothing, ? as x, a class as a character in it), or false.
func delSynth(s string) (string, bool) {
	if !strings.ContainsAny(s, "*?[\\") {
		return s, s != ""
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '*':
			if b.Len() == 0 || i+1 == len(s) {
				b.WriteByte('x') // a name, not a dot file or nothing
			}
		case '?':
			b.WriteByte('x')
		case '\\':
			if i+1 < len(s) {
				i++
				b.WriteByte(s[i])
			}
		case '[':
			end := strings.IndexByte(s[i+1:], ']')
			if end < 0 {
				b.WriteByte(c)
				continue
			}
			if end == 0 && i+2 < len(s) { // []...] : a ] first is part of the class
				if e2 := strings.IndexByte(s[i+2:], ']'); e2 >= 0 {
					end = e2 + 1
				}
			}
			class := s[i : i+1+end+1]
			picked := false
			for _, ch := range "xa0_-.Zz9" {
				if ok, err := path.Match(class, string(ch)); err == nil && ok {
					b.WriteRune(ch)
					picked = true
					break
				}
			}
			if !picked {
				for ch := byte(0x21); ch < 0x7f; ch++ {
					if ok, err := path.Match(class, string(ch)); err == nil && ok {
						b.WriteByte(ch)
						picked = true
						break
					}
				}
			}
			if !picked {
				return "", false
			}
			i += end + 1
		default:
			b.WriteByte(c)
		}
	}
	x := b.String()
	if x == "" {
		x = "x"
	}
	if ok, err := path.Match(s, x); err != nil || !ok {
		if ok2, err2 := path.Match(s, x+"x"); err2 == nil && ok2 {
			return x + "x", true
		}
		return "", false
	}
	return x, true
}

// delPlain gives why an operand is not a plain name (plan-5 5.3.3 note 3), "" when it is one: it holds a glob
// character (unless written as -LiteralPath), a variable or a substitution; it is ., .., /, ~, a drive's root,
// empty, or ends in /. or /..; a git pathspec with magic (:/, :!x).
func delPlain(op delOp) string {
	t := op.text
	switch {
	case op.expand || delPercentVar(t):
		return "holds a variable or a substitution, which the check does not expand"
	case op.glob && !op.literal:
		return "holds a glob"
	case t == "":
		return "is empty"
	case op.git && strings.HasPrefix(t, ":"):
		return "is a git pathspec with magic (:)"
	}
	s := strings.ReplaceAll(t, `\`, "/")
	trimmed := strings.TrimRight(s, "/")
	switch {
	case trimmed == "":
		return "is the root folder"
	case trimmed == "." || trimmed == "..":
		return "is " + trimmed + ", a folder by its place rather than its name"
	case trimmed == "~":
		return "is the home folder (~)"
	case strings.HasPrefix(trimmed, "~") && !strings.Contains(trimmed, "/"):
		return "is a home or working folder by its tilde name (~name, ~+, ~-)"
	case strings.HasSuffix(trimmed, "/.") || strings.HasSuffix(trimmed, "/.."):
		return "ends in /. or /.., a folder by its place rather than its name"
	case len(trimmed) == 2 && isLetter(trimmed[0]) && trimmed[1] == ':':
		return "is a drive's root"
	case delDriveRoot(trimmed):
		return "is a drive's root"
	}
	return ""
}

// delPercentVar reports a cmd.exe variable, %NAME%, in a text: refused as one in every shell, failing closed.
func delPercentVar(t string) bool {
	for {
		k := strings.IndexByte(t, '%')
		if k < 0 {
			return false
		}
		t = t[k+1:]
		end := strings.IndexByte(t, '%')
		if end < 0 {
			return false
		}
		if end > 0 && shNameStart(t[0]) && strings.Trim(t[:end], "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_") == "" {
			return true
		}
	}
}

// delDriveRoot reports a Git Bash or WSL path to a whole drive: /c, /cygdrive/c, /mnt/c.
func delDriveRoot(s string) bool {
	low := strings.ToLower(s)
	for _, pre := range []string{"/cygdrive/", "/mnt/", "/"} {
		if strings.HasPrefix(low, pre) && len(low) == len(pre)+1 && isLetter(low[len(pre)]) {
			return true
		}
	}
	return false
}

// delIsAbs reports an operand that does not depend on the folder it is read in: absolute, ~, a drive path.
func delIsAbs(p string) bool {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) || filepath.IsAbs(p) {
		return true
	}
	s := strings.ReplaceAll(p, `\`, "/")
	return strings.HasPrefix(s, "/") || len(s) >= 2 && isLetter(s[0]) && s[1] == ':'
}

// delRel gives an operand relative to the project's top folder, with forward slashes, for a record's target; ""
// when it lies outside.
func delRel(root, dir, p string) string {
	if root == "" {
		return ""
	}
	a := delAbs(dir, p)
	r, b := filepath.Clean(root), a
	if caseInsensitive(root) {
		r, b = strings.ToLower(r), strings.ToLower(b)
	}
	rel, err := filepath.Rel(r, b)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return ""
	}
	// Keep the operand's own letters where folding only placed it.
	if rel != "." && len(b) == len(a) && len(rel) <= len(a) {
		if own := filepath.ToSlash(a[len(a)-len(rel):]); strings.EqualFold(own, rel) {
			return own
		}
	}
	return rel
}
