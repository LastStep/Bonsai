package vergit

// Files under Git LFS (BUG-054): the contents' read switches every filter off, so a file Git LFS keeps reads as its
// real bytes against HEAD's blob, which is a small pointer naming them, and would always count as changed. Such a
// file is read here against that pointer instead, the way Git LFS itself defines it, trusting neither git-lfs (which
// need not be installed) nor any filter, the index or the user's settings: HEAD's pointer names the file's SHA-256
// and size, and a file whose bytes hash to them is the file HEAD's commit names. Which files are tried comes from
// HEAD's committed attributes only.

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// lfsSpec is the first line of a Git LFS pointer, by the spec's version 1.
const lfsSpec = "version https://git-lfs.github.com/spec/v1"

// lfsMaxPointer is the size under which Git LFS reads a blob as a pointer at all (its spec: "less than 1024 bytes").
const lfsMaxPointer = 1024

// lfsPointer reads a Git LFS pointer by the spec's version 1, strictly: under 1024 bytes; exactly three lines, each
// ending in a line feed: the version line, "oid sha256:" and 64 lower-case hex digits, and "size " and a decimal
// with no sign and no leading zero. A pointer with any other line (a pointer extension's ext- line, another key),
// another version (the old hawser one), another hash, CRs or anything after the size is not read: its file is then
// read as it is (and counts as changed when its bytes differ from the blob's).
func lfsPointer(blob []byte) (oid string, size int64, ok bool) {
	if len(blob) >= lfsMaxPointer || !bytes.HasSuffix(blob, []byte("\n")) {
		return "", 0, false
	}
	lines := strings.Split(strings.TrimSuffix(string(blob), "\n"), "\n")
	if len(lines) != 3 || lines[0] != lfsSpec {
		return "", 0, false
	}
	hexID, found := strings.CutPrefix(lines[1], "oid sha256:")
	if !found || len(hexID) != 64 || strings.Trim(hexID, "0123456789abcdef") != "" {
		return "", 0, false
	}
	digits, found := strings.CutPrefix(lines[2], "size ")
	if !found || digits == "" || strings.Trim(digits, "0123456789") != "" || (len(digits) > 1 && digits[0] == '0') {
		return "", 0, false
	}
	n, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return "", 0, false
	}
	return hexID, n, true
}

// lfsRead is what the read of the candidates found: the paths whose bytes are those HEAD's Git LFS pointer names
// (verified), and for each other candidate under a filter, why it still counts as changed (why) and that filter's
// name (filter).
type lfsRead struct {
	verified map[string]bool
	why      map[string]string
	filter   map[string]string
}

// lfsFilter is the value of the filter attribute that puts a file under Git LFS.
const lfsFilter = "lfs"

// lfs reads the candidates (tracked paths a status found changed in the working tree only, never in the index:
// " M"), through the fresh index's runner g (whose index is HEAD's tree). A candidate is verified, and so reads as
// unchanged, only when all of these hold:
//   - HEAD's committed attributes put it under the filter lfs: git check-attr --cached, with core.attributesFile set
//     to the null device and attr.tree to HEAD's commit (git 2.42 and later; older git has no such setting), so
//     neither a configured attributes file, nor a tree the repository's configuration names, nor a .gitattributes
//     file in the working tree (ignored or not) can add one; and none at all while the common git folder's
//     info/attributes holds a rule (a line not blank and no comment), which no commit shows and which git always reads;
//   - its HEAD entry is a regular file's blob (mode 100644 or 100755) that reads as a pointer (lfsPointer);
//   - the file, opened first, is the regular file a look at its path finds (no link, no folder, not swapped in
//     between), of the pointer's size, with HEAD's executable bit wherever this package pins core.fileMode (where it
//     does not, as on Windows, git cannot see the bit either);
//   - its bytes, streamed from that open file, hash to the pointer's SHA-256.
//
// A file under another filter (git-crypt and the like), which Bonsai cannot check without its program or key, keeps
// counting as changed, its reason saying so. The read's deadline is checked between files.
func (j *judgment) lfs(g *gitRunner, candidates []string) (lfsRead, error) {
	r := lfsRead{verified: map[string]bool{}, why: map[string]string{}, filter: map[string]string{}}
	if j.head == "" || len(candidates) == 0 {
		return r, nil
	}
	entries, err := j.headTree()
	if err != nil {
		return r, err
	}
	head := map[string]treeEntry{}
	for _, e := range entries {
		head[e.path] = e
	}
	var paths []string
	for _, p := range sortedSet(candidates) {
		if e, ok := head[p]; ok && e.kind == "blob" && (e.mode == "100644" || e.mode == "100755") {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return r, nil
	}
	attrs, err := committedFilters(g, j.head, paths)
	if err != nil {
		return r, err
	}
	localRule := infoAttributesRule(filepath.FromSlash(j.common))
	var lfsPaths []string
	for _, p := range paths {
		switch f := attrs[p]; {
		case f == "":
		case f != lfsFilter:
			r.filter[p] = f
			r.why[p] = "under the filter " + shown(f) + ", which Bonsai cannot check; a checkout with it reads dirty"
		case localRule:
			r.filter[p] = f
			r.why[p] = "under Git LFS, not checked while the repository's info/attributes holds a rule, which no commit shows"
		default:
			lfsPaths = append(lfsPaths, p)
		}
	}
	if len(lfsPaths) == 0 {
		return r, nil
	}
	blobs, err := smallBlobs(g, lfsPaths, head)
	if err != nil {
		return r, err
	}
	modePinned := false
	for _, s := range g.fs {
		if s == "core.fileMode=true" {
			modePinned = true
		}
	}
	for _, p := range lfsPaths {
		if g.ctx.Err() != nil {
			return r, fmt.Errorf("the Git LFS files could not all be read in time")
		}
		r.filter[p] = lfsFilter
		oid, size, ok := lfsPointer(blobs[p])
		if !ok {
			r.why[p] = "under Git LFS, but HEAD's blob is no Git LFS pointer this package reads, so its bytes are read as they are"
			continue
		}
		why := lfsMatch(filepath.Join(j.git.dir, filepath.FromSlash(p)), blobs[p], oid, size, head[p].mode == "100755", modePinned)
		if why == "" {
			r.verified[p] = true
			continue
		}
		r.why[p] = "under Git LFS, " + why
	}
	return r, nil
}

// lfsMatch reads the file at name against a pointer's oid and size, and HEAD's executable bit (exec) where
// modePinned: "" when it is the file the pointer names, or the pointer itself (HEAD's blob, as a checkout made without
// Git LFS holds it, which git status shows changed while the index keeps another size), else why not, in a few words.
func lfsMatch(name string, pointer []byte, oid string, size int64, exec, modePinned bool) string {
	before, err := os.Lstat(name)
	if err != nil {
		return "a file that cannot be read: " + osWords(err)
	}
	if !before.Mode().IsRegular() {
		return "not a regular file"
	}
	f, err := os.Open(name)
	if err != nil {
		return "a file that cannot be read: " + osWords(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	switch {
	case err != nil:
		return "a file that cannot be read: " + osWords(err)
	case !fi.Mode().IsRegular() || !os.SameFile(before, fi):
		return "not a regular file"
	case modePinned && (fi.Mode().Perm()&0o100 != 0) != exec:
		return "its executable bit differs from HEAD's"
	case fi.Size() != size && fi.Size() != int64(len(pointer)):
		return "its bytes differ from those HEAD's pointer names"
	}
	var head []byte // the file's first bytes, when it is the pointer's size
	if fi.Size() == int64(len(pointer)) {
		head = make([]byte, len(pointer))
		if _, err := io.ReadFull(f, head); err != nil {
			return "a file that cannot be read: " + osWords(err)
		}
		if bytes.Equal(head, pointer) {
			if more, _ := f.Read(make([]byte, 1)); more == 0 {
				return ""
			}
			return "its bytes differ from those HEAD's pointer names"
		}
		if fi.Size() != size {
			return "its bytes differ from those HEAD's pointer names"
		}
	}
	h := sha256.New()
	h.Write(head)
	n, err := io.Copy(h, f)
	n += int64(len(head))
	switch {
	case err != nil:
		return "a file that cannot be read: " + osWords(err)
	case n != size || hex.EncodeToString(h.Sum(nil)) != oid:
		return "its bytes differ from those HEAD's pointer names"
	}
	return ""
}

// committedFilters gives each path's filter attribute as HEAD's commit sets it ("" for none: unspecified, unset or
// set without a name), read by one git check-attr --cached call through g, whose index is HEAD's tree.
func committedFilters(g *gitRunner, head string, paths []string) (map[string]string, error) {
	c := *g
	c.extra = append(append([]string{}, g.extra...), "core.attributesFile="+os.DevNull, "attr.tree="+head)
	var stdin bytes.Buffer
	for _, p := range paths {
		stdin.WriteString(p)
		stdin.WriteByte(0)
	}
	out, err := c.run(stdin.Bytes(), "check-attr", "--cached", "-z", "--stdin", "filter")
	if err != nil {
		return nil, fmt.Errorf("git cannot read the committed attributes: %v", err)
	}
	fields := splitZ(out)
	if len(fields)%3 != 0 {
		return nil, fmt.Errorf("git check-attr gave an answer this package does not read: %q", firstLine(string(out)))
	}
	attrs := map[string]string{}
	for i := 0; i < len(fields); i += 3 {
		if v := fields[i+2]; fields[i+1] == "filter" && v != "unspecified" && v != "unset" && v != "set" {
			attrs[fields[i]] = v
		}
	}
	return attrs, nil
}

// infoAttributesRule reports whether the common git folder's info/attributes holds a rule: a line that is neither
// blank nor a comment. One that cannot be read (other than missing) counts as holding one.
func infoAttributesRule(common string) bool {
	b, err := os.ReadFile(filepath.Join(common, "info", "attributes"))
	if err != nil {
		return !os.IsNotExist(err)
	}
	for _, l := range strings.Split(string(b), "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			return true
		}
	}
	return false
}

// smallBlobs reads HEAD's blobs of the paths that could be a pointer (under lfsMaxPointer bytes), by two git cat-file
// calls through g: their sizes, then their bytes. A path whose blob is larger is left out (no pointer).
func smallBlobs(g *gitRunner, paths []string, head map[string]treeEntry) (map[string][]byte, error) {
	var ids []string
	seen := map[string]bool{}
	for _, p := range paths {
		if id := head[p].id; !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	out, err := g.run([]byte(strings.Join(ids, "\n")+"\n"), "cat-file", "--batch-check")
	if err != nil {
		return nil, fmt.Errorf("git cannot read HEAD's blobs: %v", err)
	}
	var small []string
	for _, l := range strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") {
		f := strings.Fields(l)
		if len(f) != 3 || f[1] != "blob" {
			return nil, fmt.Errorf("git cannot read HEAD's blob: %q", firstLine(l))
		}
		if n, err := strconv.ParseInt(f[2], 10, 64); err == nil && n < lfsMaxPointer {
			small = append(small, f[0])
		}
	}
	got := map[string][]byte{}
	if len(small) > 0 {
		out, err = g.run([]byte(strings.Join(small, "\n")+"\n"), "cat-file", "--batch")
		if err != nil {
			return nil, fmt.Errorf("git cannot read HEAD's blobs: %v", err)
		}
		rd := bufio.NewReader(bytes.NewReader(out))
		for range small {
			l, err := rd.ReadString('\n')
			f := strings.Fields(l)
			if err != nil || len(f) != 3 || f[1] != "blob" {
				return nil, fmt.Errorf("git cat-file gave an answer this package does not read: %q", firstLine(l))
			}
			n, err := strconv.Atoi(f[2])
			if err != nil || n < 0 || n >= lfsMaxPointer {
				return nil, fmt.Errorf("git cat-file gave an answer this package does not read: %q", firstLine(l))
			}
			b := make([]byte, n+1)
			if _, err := io.ReadFull(rd, b); err != nil || b[n] != '\n' {
				return nil, fmt.Errorf("git cat-file gave an answer this package does not read: %q", firstLine(l))
			}
			got[f[0]] = b[:n]
		}
	}
	blobs := map[string][]byte{}
	for _, p := range paths {
		if b, ok := got[head[p].id]; ok {
			blobs[p] = b
		}
	}
	return blobs, nil
}

// sortedSet gives the paths sorted, each once.
func sortedSet(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}
