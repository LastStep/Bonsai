package vergit

// The attributes that convert a file's bytes on their way into git, other than filters and line ends (V2b's H1): the
// contents' read hashes each file as git reads it, through those conversions, so different bytes that convert to
// HEAD's blob read as unchanged. Each tracked file such an attribute applies to is read again here, by its bytes.

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// The attributes git converts a file by on its way in besides a filter (switched off for the contents' read) and line
// ends (text, eol, crlf, with core.autocrlf and core.eol: a known limit, see Tree): git's convert_to_git applies a
// filter, then working-tree-encoding, then line ends, then ident, and no other.
const (
	attrIdent    = "ident"                 // $Id: <anything on one line> $ reads as $Id$
	attrEncoding = "working-tree-encoding" // the file's bytes are read from that encoding into UTF-8
)

// treeEntry is one entry of HEAD's tree (git ls-tree -r).
type treeEntry struct {
	mode, kind, id, path string
}

// headTree reads HEAD's tree once per judgment, every entry at any depth (none when HEAD names no commit yet).
func (j *judgment) headTree() ([]treeEntry, error) {
	if j.headRead || j.head == "" {
		return j.headEntries, nil
	}
	out, err := j.git.run(nil, "ls-tree", "-r", "-z", "--full-tree", j.head)
	if err != nil {
		return nil, fmt.Errorf("git cannot read HEAD's tree: %v", err)
	}
	for _, e := range splitZ(out) {
		meta, p, ok := strings.Cut(e, "\t")
		f := strings.Fields(meta)
		if !ok || len(f) != 3 {
			return nil, fmt.Errorf("git ls-tree gave an entry this package does not read: %q", e)
		}
		j.headEntries = append(j.headEntries, treeEntry{mode: f[0], kind: f[1], id: f[2], path: p})
	}
	j.headRead = true
	return j.headEntries, nil
}

// converted reads again, by their bytes, the tracked files the contents' read found unchanged (changed holds the ones
// it found changed) that ident or working-tree-encoding applies to, as the fresh index's runner g reads attributes (the
// working tree's .gitattributes files, an ignored one too, the fresh index's, info/attributes), with one git check-attr
// call for every regular file of HEAD's tree. Such a file reads as unchanged only when its bytes, or its bytes with
// CRLF read as LF, are HEAD's blob's: git reads any $Id:...$ keyword as $Id$, and other bytes under an encoding as the
// same text, so different bytes would read as HEAD's. As with a filter (switched off for the read), an honest checkout
// such an attribute converts (an expanded $Id$, a UTF-16 file) therefore reads as changed too: compared with git's own
// checkout of HEAD's blob instead, a keyword's text or an encoding set in info/attributes, which no diff shows, would
// pass. Each such file is named with the attribute; one that cannot be read counts too.
func (j *judgment) converted(g *gitRunner, changed map[string]bool) ([]Hidden, error) {
	entries, err := j.headTree()
	if err != nil {
		return nil, err
	}
	ids := map[string]string{}
	var stdin bytes.Buffer
	for _, e := range entries {
		if e.kind != "blob" || (e.mode != "100644" && e.mode != "100755") || changed[e.path] {
			continue
		}
		ids[e.path] = e.id
		stdin.WriteString(e.path)
		stdin.WriteByte(0)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	out, err := g.run(stdin.Bytes(), "check-attr", "-z", "--stdin", attrIdent, attrEncoding)
	if err != nil {
		return nil, fmt.Errorf("git cannot read the tracked files' attributes: %v", err)
	}
	fields := splitZ(out)
	if len(fields)%3 != 0 {
		return nil, fmt.Errorf("git check-attr gave an answer this package does not read: %q", firstLine(string(out)))
	}
	attrs := map[string][]string{}
	for i := 0; i < len(fields); i += 3 {
		p, attr, value := fields[i], fields[i+1], fields[i+2]
		switch {
		case attr == attrIdent && value == "set":
		case attr == attrEncoding && value != "unspecified" && value != "unset" && value != "set":
		default:
			continue
		}
		attrs[p] = append(attrs[p], attr)
	}
	paths := make([]string, 0, len(attrs))
	for p := range attrs {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var hidden []Hidden
	for _, p := range paths {
		file := filepath.Join(j.git.dir, filepath.FromSlash(p))
		fi, err := os.Lstat(file)
		if err != nil || !fi.Mode().IsRegular() {
			continue // gone, or of another kind: the contents' read finds it changed
		}
		b, err := os.ReadFile(file)
		switch {
		case err != nil:
			hidden = append(hidden, Hidden{Path: p, Why: "a file under " + attrWords(attrs[p]) + " that cannot be read: " + osWords(err)})
		case blobID(b, len(ids[p])) == ids[p] || blobID(lf(b), len(ids[p])) == ids[p]: // HEAD's bytes
		default:
			verb := " make"
			if len(attrs[p]) == 1 {
				verb = " makes"
			}
			hidden = append(hidden, Hidden{Path: p, Why: "its bytes differ from HEAD's, though " + attrWords(attrs[p]) + verb +
				" git read them as HEAD's"})
		}
	}
	return hidden, nil
}

// attrWords names a file's converting attributes: "the ident attribute", "the ident and working-tree-encoding
// attributes".
func attrWords(attrs []string) string {
	if len(attrs) == 1 {
		return "the " + attrs[0] + " attribute"
	}
	return "the " + strings.Join(attrs, " and ") + " attributes"
}

// blobID is git's object id of a blob holding b, by the repository's hash: SHA-1 for a 40-digit id, SHA-256 for a
// 64-digit one.
func blobID(b []byte, digits int) string {
	head := "blob " + strconv.Itoa(len(b)) + "\x00"
	if digits == 64 {
		h := sha256.New()
		h.Write([]byte(head))
		h.Write(b)
		return hex.EncodeToString(h.Sum(nil))
	}
	h := sha1.New()
	h.Write([]byte(head))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// lf reads a file's line ends as LF.
func lf(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) }
