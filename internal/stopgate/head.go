package stopgate

// The commit at a checkout's HEAD, read from its git files with no git process (design/plan-5.md, 5.3.4 note 2: "HEAD
// is read from the session's checkout's git files (a branch's ref file, packed-refs, a detached HEAD) with no git
// process; a HEAD it cannot read blocks"), as git itself keeps them:
//   - HEAD, in the checkout's own git folder (a worktree's is its own: <common>/worktrees/<name>/HEAD), holds either a
//     commit (a detached HEAD) or "ref: <name>", the branch it is on;
//   - a ref is a loose file under the git folder (refs/heads/<branch>), holding a commit or, for a symbolic ref, "ref:
//     <name>" again; a per-worktree ref (refs/worktree/, refs/bisect/, refs/rewritten/) lives in the worktree's own git
//     folder, every other in the common one;
//   - a ref with no loose file is looked up in the common folder's packed-refs, one "<commit> <name>" per line after
//     its "# pack-refs" header, a peeled tag's "^<commit>" line after its own.
//
// A commit is 40 hexadecimal characters (SHA-1) or 64 (SHA-256), as git writes it. A symbolic ref is followed at most
// five deep. HEAD cannot be read, and the gate blocks, when: the checkout has no git folder Bonsai reads; HEAD or a ref
// file does not read as one of those lines; a ref's name is not one git would write (it could reach outside the git
// folder); the branch has no commit yet; or the repository keeps its refs in a reftable (git's
// extensions.refStorage), which Bonsai does not read.

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// maxRefFile caps one of git's pointer files (HEAD, a loose ref), in bytes; maxPackedRefs caps packed-refs.
const (
	maxRefFile    = 4096
	maxPackedRefs = 64 << 20
)

// Head gives the commit at the HEAD of the checkout laid out as g, or why it cannot be read, in ASCII words.
func Head(g workspace.GitLayout) (string, error) {
	if g.GitDir == "" {
		return "", errors.New("the checkout holds no .git that Bonsai reads")
	}
	line, err := pointer(filepath.Join(g.GitDir, "HEAD"))
	if err != nil {
		return "", errors.New("HEAD in its git folder cannot be read (" + osWords(err) + ")")
	}
	if isCommit(line) {
		return strings.ToLower(line), nil
	}
	name, isRef := strings.CutPrefix(line, "ref:")
	if !isRef {
		return "", errors.New("HEAD in its git folder holds neither a commit nor a ref")
	}
	if fi, err := os.Stat(filepath.Join(g.Common, "reftable")); err == nil && fi.IsDir() {
		return "", errors.New("the repository keeps its refs in a reftable (git's extensions.refStorage), which Bonsai does not read")
	}
	return resolveRef(g, strings.TrimSpace(name), "HEAD")
}

// resolveRef follows a ref, named by from, to its commit.
func resolveRef(g workspace.GitLayout, name, from string) (string, error) {
	for depth := 0; depth < 5; depth++ {
		if !refName(name) {
			return "", errors.New(from + " names " + quote(name) + ", which is not a ref name git writes")
		}
		line, err := pointer(filepath.Join(refFolder(g, name), filepath.FromSlash(name)))
		switch {
		case err == nil:
			if isCommit(line) {
				return strings.ToLower(line), nil
			}
			next, isRef := strings.CutPrefix(line, "ref:")
			if !isRef {
				return "", errors.New("the ref file " + quote(name) + " holds neither a commit nor a ref")
			}
			from, name = "the ref "+quote(name), strings.TrimSpace(next)
			continue
		case !errors.Is(err, fs.ErrNotExist):
			return "", errors.New("the ref file " + quote(name) + " cannot be read (" + osWords(err) + ")")
		}
		sha, found, err := packed(g.Common, name)
		switch {
		case err != nil:
			return "", err
		case found:
			return sha, nil
		}
		return "", errors.New(from + " names " + quote(name) + ", which has no commit yet (no ref file, and no line in packed-refs)")
	}
	return "", errors.New(from + " is a ref that names a ref more than five deep")
}

// refFolder is the git folder a ref lives in: the checkout's own for a per-worktree ref, else the common one.
func refFolder(g workspace.GitLayout, name string) string {
	for _, p := range []string{"refs/worktree/", "refs/bisect/", "refs/rewritten/"} {
		if strings.HasPrefix(name, p) {
			return g.GitDir
		}
	}
	return g.Common
}

// packed looks a ref up in the common git folder's packed-refs: its commit, whether it is there, and why the file
// cannot be read (a file that is not there holds no ref).
func packed(common, name string) (string, bool, error) {
	f, err := os.Open(filepath.Join(common, "packed-refs"))
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, errors.New("packed-refs cannot be read (" + osWords(err) + ")")
	}
	defer func() { _ = f.Close() }()
	br := bufio.NewReader(io.LimitReader(f, maxPackedRefs+1))
	read := 0
	for {
		line, err := br.ReadString('\n')
		read += len(line)
		if read > maxPackedRefs {
			return "", false, errors.New("packed-refs is larger than 64 MiB")
		}
		line = strings.TrimRight(line, "\r\n")
		if sha, ref, ok := strings.Cut(line, " "); ok && ref == name && isCommit(sha) {
			return strings.ToLower(sha), true, nil
		}
		if err == io.EOF {
			return "", false, nil
		}
		if err != nil {
			return "", false, errors.New("packed-refs cannot be read (" + osWords(err) + ")")
		}
	}
}

// pointer reads one of git's pointer files: its first line, without its line ending or the spaces around it.
func pointer(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	if fi, err := f.Stat(); err == nil && fi.IsDir() {
		return "", errors.New("it is a folder")
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxRefFile+1))
	if err != nil {
		return "", err
	}
	if len(raw) > maxRefFile {
		return "", errors.New("it is larger than any file git writes there")
	}
	s := string(raw)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s), nil
}

// isCommit reports a commit as git writes one: 40 or 64 hexadecimal characters.
func isCommit(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

// refName reports a ref name the gate reads: refs/ and parts git allows (git check-ref-format's rules, the ones that
// keep a name inside the git folder and on one line): no empty part, no part starting with a dot or ending in .lock,
// no "..", "@{", backslash, control character, space or any of ~^:?*[.
func refName(name string) bool {
	if !strings.HasPrefix(name, "refs/") || strings.HasSuffix(name, "/") || strings.HasSuffix(name, ".") ||
		strings.Contains(name, "..") || strings.Contains(name, "@{") {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || strings.HasPrefix(part, ".") || strings.HasSuffix(part, ".lock") {
			return false
		}
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c < 0x21 || c == 0x7f || strings.IndexByte(`\~^:?*[`, c) >= 0 {
			return false
		}
	}
	return true
}
