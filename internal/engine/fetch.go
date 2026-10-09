package engine

// The plain fetch (plan part 3; spec §5, "Bonsai fetches each pack itself too, into <home>/cache/, to read bonsai/:
// once per commit, with the machine's own git"), and reading a pack at a commit.
//
// The cache holds one bare git repository per pack source, <home>/cache/git/<16 hex>.git, the hex the first 16
// characters of the SHA-256 of the source as bonsai.yaml writes it. A fetched commit is kept by a ref,
// refs/bonsai/commits/<commit>, so git never prunes it and a commit is fetched once. Files are read from git's
// objects (git cat-file), never from a checkout: a pack's bytes reach the project exactly as committed, whatever
// the machine's line-ending settings.
//
// git runs with the source and ref after --, so neither can be read as an option (bonsai.yaml also refuses one that
// starts with -); with GIT_TERMINAL_PROMPT=0, so a source that wants a login fails instead of waiting for input;
// with the ext transport off; and with git's own location variables cleared, so a git hook that runs bonsai cannot
// point it at another repository. The machine's own git login is used as it is: no token is stored.

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// PackData is a pack read at one commit: what the engine writes from it.
type PackData struct {
	Ref      workspace.PackRef // the bonsai.yaml entry it was read for
	Commit   string            // the 40-character commit
	Manifest *workspace.Pack   // bonsai/pack.yaml
	Files    map[string][]byte // each files entry's bytes, by its project path
	Block    string            // the block text after block.md's leading HTML comment, LF line ends; "" for none
	SHA256   string            // the content hash (contentHash)
	Tree     map[string]string // every entry of the pack's folder (the plugin's root): its path there, to its git mode and object id
	Code     []CodePart        // the plugin's own code parts (consent.go), sorted by path; none for a plugin that carries none
}

// cache is the home's pack cache.
type cache struct {
	home string
}

func (c cache) dir(source string) string {
	sum := sha256.Sum256([]byte(source))
	return filepath.Join(c.home, "cache", "git", hex.EncodeToString(sum[:])[:16]+".git")
}

// gitEnv is the environment git runs in: the process's own, less the variables that move git's repository or
// work tree, with prompts off.
func gitEnv() []string {
	drop := map[string]bool{"GIT_DIR": true, "GIT_WORK_TREE": true, "GIT_INDEX_FILE": true, "GIT_OBJECT_DIRECTORY": true,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": true, "GIT_COMMON_DIR": true, "GIT_NAMESPACE": true, "GIT_TERMINAL_PROMPT": true}
	var env []string
	for _, kv := range os.Environ() {
		k := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			k = kv[:i]
		}
		if runtime.GOOS == "windows" {
			k = strings.ToUpper(k)
		}
		if !drop[k] {
			env = append(env, kv)
		}
	}
	return append(env, "GIT_TERMINAL_PROMPT=0")
}

// git runs git on a bare repository and returns its stdout. stdin may be nil.
func git(gitDir string, stdin io.Reader, args ...string) ([]byte, error) {
	full := append([]string{"--git-dir=" + gitDir, "-c", "protocol.ext.allow=never"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = gitEnv()
	cmd.Stdin = stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, errorf(ExitRuntime, "install git, or put it on the PATH, then run the command again",
				"git is not on the PATH: Bonsai fetches packs with the machine's own git")
		}
		return stdout.Bytes(), &gitError{args: args, stderr: firstLine(stderr.String()), err: err}
	}
	return stdout.Bytes(), nil
}

type gitError struct {
	args   []string
	stderr string
	err    error
}

func (e *gitError) Error() string {
	return "git " + e.args[0] + " failed: " + e.stderr
}

func firstLine(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "hint:") {
			return ascii(l)
		}
	}
	return "(no message)"
}

// has reports whether the cache holds a commit. A cache folder that does not exist holds nothing.
func (c cache) has(source, commit string) bool {
	dir := c.dir(source)
	if _, err := os.Stat(filepath.Join(dir, "HEAD")); err != nil {
		return false
	}
	_, err := git(dir, nil, "cat-file", "-e", commit+"^{commit}")
	return err == nil
}

// fetch brings source at ref into the cache and returns the 40-character commit it resolves to. A ref that is a
// 40-character commit already in the cache fetches nothing.
func (c cache) fetch(source, ref string) (string, error) {
	dir := c.dir(source)
	if _, err := os.Stat(filepath.Join(dir, "HEAD")); err != nil {
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return "", errorf(ExitRuntime, "check that the Bonsai home can be written (BONSAI_HOME), then run again",
				"the pack cache %s cannot be made: %v", filepath.ToSlash(filepath.Dir(dir)), err)
		}
		cmd := exec.Command("git", "init", "-q", "--bare", "--", dir)
		cmd.Env = gitEnv()
		if out, err := cmd.CombinedOutput(); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return "", errorf(ExitRuntime, "install git, or put it on the PATH, then run the command again",
					"git is not on the PATH: Bonsai fetches packs with the machine's own git")
			}
			return "", errorf(ExitRuntime, "check that the Bonsai home can be written (BONSAI_HOME), then run again",
				"the pack cache %s cannot be made: %s", filepath.ToSlash(dir), firstLine(string(out)))
		}
	}
	next := "check the pack's source and ref in bonsai.yaml, and that this machine reaches " + source + ", then run again"
	if commitPattern.MatchString(ref) {
		if !c.has(source, ref) {
			_, err := git(dir, nil, "fetch", "-q", "--no-tags", "--", source, ref)
			if err != nil || !c.has(source, ref) {
				// A server that will not send a commit by its id sends it with its branches and tags.
				_, err2 := git(dir, nil, "fetch", "-q", "--no-tags", "--", source,
					"+refs/heads/*:refs/bonsai/heads/*", "+refs/tags/*:refs/bonsai/tags/*")
				if !c.has(source, ref) {
					if err2 != nil {
						return "", errorf(ExitRuntime, next, "fetching %s from %s failed: %v", ref, source, err2)
					}
					return "", errorf(ExitInput, next, "the commit %s is not in %s", ref, source)
				}
			}
		}
		return ref, c.keep(dir, ref)
	}
	if _, err := git(dir, nil, "fetch", "-q", "--no-tags", "--", source, ref); err != nil {
		var ge *gitError
		if errors.As(err, &ge) && strings.Contains(ge.stderr, "couldn't find remote ref") {
			return "", errorf(ExitInput, next, "the ref %s is not in %s", ref, source)
		}
		return "", errorf(ExitRuntime, next, "fetching %s from %s failed: %v", ref, source, err)
	}
	out, err := git(dir, nil, "rev-parse", "--verify", "-q", "FETCH_HEAD^{commit}")
	commit := strings.TrimSpace(string(out))
	if err != nil || !commitPattern.MatchString(commit) {
		return "", errorf(ExitInput, next, "the ref %s of %s is not a commit", ref, source)
	}
	return commit, c.keep(dir, commit)
}

// keep holds a fetched commit with a ref, so git never prunes it.
func (c cache) keep(dir, commit string) error {
	if _, err := git(dir, nil, "update-ref", "refs/bonsai/commits/"+commit, commit); err != nil {
		return errorf(ExitRuntime, "check that the Bonsai home can be written (BONSAI_HOME), then run again",
			"the pack cache cannot keep %s: %v", commit, err)
	}
	return nil
}

type treeEntry struct {
	mode, typ, sha, path string
}

// packAt reads the pack ref names at commit, from the cache (which must hold the commit).
func (c cache) packAt(ref workspace.PackRef, commit string) (*PackData, error) {
	dir := c.dir(ref.Source)
	where := ref.Source + " at " + short(commit)
	nextPack := "check the pack's source, path and ref in bonsai.yaml; a pack's maker fixes the pack itself"
	out, err := git(dir, nil, "ls-tree", "-r", "-z", "--full-tree", commit)
	if err != nil {
		return nil, errorf(ExitRuntime, "run the command again; if it fails again, delete the pack cache ("+
			filepath.ToSlash(dir)+") and run it again", "reading %s failed: %v", where, err)
	}
	prefix := ""
	if ref.Path != "" {
		prefix = ref.Path + "/"
	}
	entries := map[string]treeEntry{} // by path inside the pack
	for _, rec := range strings.Split(string(out), "\x00") {
		tab := strings.IndexByte(rec, '\t')
		if tab < 0 {
			continue
		}
		f := strings.Fields(rec[:tab])
		p := rec[tab+1:]
		if len(f) != 3 || !strings.HasPrefix(p, prefix) {
			continue
		}
		entries[p[len(prefix):]] = treeEntry{mode: f[0], typ: f[1], sha: f[2], path: p[len(prefix):]}
	}
	var shas []string
	for _, e := range entries {
		if e.typ == "blob" {
			shas = append(shas, e.sha)
		}
	}
	blobs, err := catBlobs(dir, shas)
	if err != nil {
		return nil, errorf(ExitRuntime, "run the command again", "reading %s failed: %v", where, err)
	}
	read := func(p string) ([]byte, error) {
		e, ok := entries[p]
		if !ok || e.typ != "blob" {
			return nil, os.ErrNotExist
		}
		if e.mode == "120000" {
			return nil, errors.New("a symbolic link, which Bonsai does not write into a project")
		}
		return blobs[e.sha], nil
	}
	raw, err := read(workspace.PackFile)
	if err != nil {
		where := where
		if ref.Path != "" {
			where += ", folder " + ref.Path
		}
		return nil, errorf(ExitInput, nextPack, "%s has no %s, so it is not a Bonsai pack", where, workspace.PackFile)
	}
	manifest, err := workspace.ReadPack(raw)
	if err != nil {
		return nil, errorf(ExitInput, nextPack, "the pack %s (%s): %v", ref.ID, where, err)
	}
	if ref.ID != "" && manifest.ID != ref.ID {
		return nil, errorf(ExitInput, "make bonsai.yaml's id for this pack "+manifest.ID+", or point it at the pack it means",
			"bonsai.yaml names the pack %s, but %s says its id is %s", ref.ID, where, manifest.ID)
	}
	pd := &PackData{Ref: ref, Commit: commit, Manifest: manifest, Files: map[string][]byte{}}
	for i, fe := range manifest.Files {
		b, err := read("bonsai/files/" + fe.From)
		if err != nil {
			return nil, errorf(ExitInput, nextPack, "the pack %s (%s): files item %d comes from bonsai/files/%s, which is %s",
				ref.ID, where, i+1, fe.From, missingOr(err))
		}
		pd.Files[fe.Path] = b
	}
	if manifest.Block != "" {
		b, err := read("bonsai/" + manifest.Block)
		if err != nil {
			return nil, errorf(ExitInput, nextPack, "the pack %s (%s): its block is bonsai/%s, which is %s",
				ref.ID, where, manifest.Block, missingOr(err))
		}
		pd.Block = blockText(b)
	}
	pd.SHA256 = contentHash(entries, blobs)
	pd.Tree = map[string]string{}
	for p, e := range entries {
		pd.Tree[p] = e.mode + " " + e.sha
	}
	pd.Code = pluginCode(entries, blobs)
	return pd, nil
}

func missingOr(err error) string {
	if errors.Is(err, os.ErrNotExist) {
		return "not there"
	}
	return err.Error()
}

// catBlobs reads blobs by their ids in one git cat-file --batch.
func catBlobs(dir string, shas []string) (map[string][]byte, error) {
	out := map[string][]byte{}
	if len(shas) == 0 {
		return out, nil
	}
	sort.Strings(shas)
	var in bytes.Buffer
	for _, s := range shas {
		in.WriteString(s + "\n")
	}
	raw, err := git(dir, &in, "cat-file", "--batch")
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(bytes.NewReader(raw))
	for range shas {
		header, err := r.ReadString('\n')
		if err != nil {
			return nil, errors.New("git cat-file --batch ended early")
		}
		f := strings.Fields(header)
		if len(f) != 3 {
			return nil, errors.New("git cat-file --batch: " + strings.TrimSpace(header))
		}
		size, err := strconv.Atoi(f[2])
		if err != nil {
			return nil, err
		}
		body := make([]byte, size+1) // the content and its trailing line feed
		if _, err := io.ReadFull(r, body); err != nil {
			return nil, err
		}
		out[f[0]] = body[:size]
	}
	return out, nil
}

// contentHash is a pack's content hash, the lock's packs[].sha256 (contract §14): the SHA-256 of one line per file
// in the pack's folder at its commit, "<SHA-256 of the file's bytes with line endings made LF>  <path in the pack>",
// sorted by path (bytewise), each ending in a line feed. It is line-ending blind, as every fingerprint in the lock
// is, and the same on every machine.
func contentHash(entries map[string]treeEntry, blobs map[string][]byte) string {
	var paths []string
	for p, e := range entries {
		if e.typ == "blob" {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	var b strings.Builder
	for _, p := range paths {
		b.WriteString(workspace.HashLF(blobs[entries[p].sha]) + "  " + p + "\n")
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// blockText is a pack's block.md without its leading HTML comment (its documentation stays in the pack, spec §5),
// with LF line ends and no blank lines at its start or end.
func blockText(raw []byte) string {
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")
	s = strings.TrimPrefix(s, "\ufeff")
	if t := strings.TrimLeft(s, " \n\t"); strings.HasPrefix(t, "<!--") {
		if end := strings.Index(t, "-->"); end >= 0 {
			s = t[end+3:]
		}
	}
	return strings.Trim(s, "\n")
}

// pluginSource is the plugin source the marketplace entry gives a pack (Claude Code's plugin sources): github for
// a pack that is a whole GitHub repository, git-subdir for a pack in a folder of a repository, url for any other
// git repository; each pinned to the locked commit by sha, with no version (spec §5).
func pluginSource(source, folder, commit string) any {
	u := gitURL(source)
	if folder != "" {
		return objectOf("source", "git-subdir", "url", u, "path", folder, "sha", commit)
	}
	if m := githubPattern.FindStringSubmatch(source); m != nil {
		return objectOf("source", "github", "repo", m[1]+"/"+m[2], "sha", commit)
	}
	return objectOf("source", "url", "url", u, "sha", commit)
}

var githubPattern = regexp.MustCompile(`^https://github\.com/([A-Za-z0-9_.-]+)/([A-Za-z0-9_.-]+?)(?:\.git)?/?$`)

// gitURL turns a source into a URL Claude Code fetches: a URL stays as it is; a local folder (a test's bare
// repository) becomes a file:// URL.
func gitURL(source string) string {
	if strings.Contains(source, "://") || strings.HasPrefix(source, "git@") {
		return source
	}
	abs, err := filepath.Abs(source)
	if err != nil {
		abs = source
	}
	p := filepath.ToSlash(abs)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return (&url.URL{Scheme: "file", Path: path.Clean(p)}).String()
}
