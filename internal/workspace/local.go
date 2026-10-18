package workspace

// A project's .bonsai/local/ (contract §3, spec §6): the folder that is never committed, its .gitignore, which every
// writer restores when it is missing, and the checkout whose local/ a writer uses, found by reading the checkout's
// .git with no git process.
//
// Contract §3: worktrees share the main checkout's .bonsai/local/; a task worktree has none of its own. Spec §6 and
// contract §3 name git rev-parse --git-common-dir for finding it; ReadGitLayout reads the same files git reads,
// without starting git, so a hook stays fast (design/plan-5.md, 5.2.2 note 3; a test holds its answer equal to
// git's, Find's, on a main checkout, a worktree, a submodule and a separate git folder):
//
//   - .git is a folder: the checkout is its own main, and .git is its git folder.
//   - .git is a file, "gitdir: <path>" (relative to the checkout, or absolute): <path> is the checkout's git folder.
//     Its commondir file names the common git folder (relative to the git folder, or absolute). When the git folder
//     is <common>/worktrees/<name> and the common folder is named .git, the checkout is a worktree and its main
//     checkout is the common folder's parent.
//   - Anything else is its own main: a git folder with no commondir (a submodule, a checkout with a separate git
//     folder), a common folder not named .git (a bare repository's worktree, or a separate git folder's), no .git,
//     or a file that cannot be read. Paths are compared case-blind on Windows.
//
// An agent can rewrite those files (a worktree's .git file, a commondir file), so the main checkout they name is used
// only when it is verified (VerifyMain; the director's 8 Oct rule, spec §7: whatever the guard trusts to find the main
// checkout "cannot be rewritten by an agent without the guard noticing (fail closed)"; design/plan-5.md, 5.3.1 note
// 2). A worktree's main checkout is verified when all three hold:
//
//   - git's own link back: <common>/worktrees/<name>/gitdir, the file git keeps beside the worktree's HEAD, names this
//     worktree's .git file;
//   - the main checkout's bonsai.yaml holds the workspace id this checkout's holds;
//   - this machine's record (record.go: <home>/workspaces/r-<key of main>/workspace.json, written only by bonsai init
//     and update run in a main checkout) names the main checkout's path and that id, and is the first record on this
//     machine to claim the id (FirstClaim: a checkout claiming it later, a decoy among them, is never honoured).
//
// A checkout that is its own main needs the third alone. The link back and the main's bonsai.yaml lie in folders an
// agent can write, so they alone prove nothing; the record lies outside every project, where only Bonsai writes (the
// guard refuses an agent's file tools there, and the engine's deny rule a shell redirect: a program the agent runs
// that forges it stays the tripwire case question A accepts). Not verified: a rewritten .git file, a forged
// commondir, a decoy main with or without a record, a worktree made to look like its own main, a clone that never
// ran bonsai update on this machine, and a second clone of one project on one machine (bonsai check warns of it).
//
// FindLocal is where every writer of .bonsai/local/ writes, and the guard's records go the same way: a worktree whose
// main checkout is verified writes the main checkout's local/; any other checkout writes its own. Every reader of
// local/ (the recorder, bonsai hook start, the asks, bonsai logs, the cleaner, check --write) takes this answer, so
// the rule has one home.

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LocalDir is the project's folder that is never committed (contract §3), relative to the main checkout.
const LocalDir = ".bonsai/local"

// GitignoreFile and GitignoreText are .bonsai/.gitignore, written by Bonsai (spec §6, contract §3): bonsai init and
// update write it, bonsai check finds it missing or changed, and every writer of .bonsai/local/ restores it when it
// is missing (EnsureGitignore). This is the text's one home. Its comment names what local/ holds (step 5.5.0 added the
// messages between agents and each agent's counts): a project linked before takes the new text at its next update,
// which check's gitignore finding names until then.
const (
	GitignoreFile = ".bonsai/.gitignore"
	GitignoreText = "# .bonsai/.gitignore: written by Bonsai (spec section 6). .bonsai/local/ holds this checkout's log, questions\n" +
		"# and their answers, ladder results, messages between agents and each agent's counts, never committed. Bonsai\n" +
		"# writes this file again when it is missing or changed.\n" +
		"local/\n"
)

// EnsureGitignore writes .bonsai/.gitignore in the checkout at root when it is missing (spec §6: "every writer of
// local/ restores a missing .bonsai/.gitignore"), whole or not at all, and reports whether it wrote. A file that is
// there is left as it is, changed or not: bonsai check finds a changed one, and bonsai update writes Bonsai's text
// again. Two writers that both find it missing both write the same bytes; the second rename replaces the first.
func EnsureGitignore(root string) (bool, error) {
	path := filepath.Join(root, filepath.FromSlash(GitignoreFile))
	_, err := os.Lstat(path)
	switch {
	case err == nil:
		return false, nil
	case !errors.Is(err, fs.ErrNotExist):
		return false, err
	}
	if err := WriteFileAtomic(path, []byte(GitignoreText)); err != nil {
		return false, err
	}
	return true, nil
}

// GitLayout is a checkout's git layout, as the files in its .git say (ReadGitLayout). Every path is absolute and
// clean, in the system's own separators: use filepath.ToSlash to store or print one.
type GitLayout struct {
	Root   string // the checkout's top folder, as given
	GitDir string // the checkout's own git folder: Root/.git, or the folder its .git file names; "" when none reads
	Common string // the common git folder its commondir names; GitDir when it names none
	Main   string // the main checkout by the layout alone: Common's parent for a worktree; else Root
}

// ReadGitLayout reads the git layout of the checkout whose top folder is root, with no git process (the rules are
// this file's comment). It never fails: what cannot be read leaves the checkout its own main.
func ReadGitLayout(root string) GitLayout {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = filepath.Clean(root)
	}
	g := GitLayout{Root: abs, Main: abs}
	dotgit := filepath.Join(abs, ".git")
	fi, err := os.Stat(dotgit)
	if err != nil {
		return g
	}
	if fi.IsDir() {
		g.GitDir, g.Common = dotgit, dotgit
		return g
	}
	if !fi.Mode().IsRegular() {
		return g
	}
	line, ok := firstLine(dotgit)
	rest, isGitdir := strings.CutPrefix(line, "gitdir:")
	gitdir := strings.TrimSpace(rest)
	if !ok || !isGitdir || gitdir == "" {
		return g
	}
	gitdir = resolve(abs, gitdir)
	if fi, err := os.Stat(gitdir); err != nil || !fi.IsDir() {
		return g
	}
	g.GitDir, g.Common = gitdir, gitdir
	common, ok := firstLine(filepath.Join(gitdir, "commondir"))
	common = strings.TrimSpace(common)
	if !ok || common == "" {
		return g // a submodule, or a checkout with a separate git folder: its own main
	}
	g.Common = resolve(gitdir, common)
	fold := runtime.GOOS == "windows"
	if sameName(filepath.Base(g.Common), ".git", fold) && childOf(gitdir, filepath.Join(g.Common, "worktrees"), fold) {
		g.Main = filepath.Dir(g.Common)
	}
	return g
}

// Branch is the checkout's branch, read from its git folder's HEAD: "" for a detached HEAD, no git folder, or a
// HEAD that cannot be read.
func (g GitLayout) Branch() string {
	if g.GitDir == "" {
		return ""
	}
	line, ok := firstLine(filepath.Join(g.GitDir, "HEAD"))
	branch, isRef := strings.CutPrefix(strings.TrimSpace(line), "ref: refs/heads/")
	if !ok || !isRef {
		return ""
	}
	return branch
}

// Worktree reports whether the layout names a main checkout other than the checkout itself.
func (g GitLayout) Worktree() bool { return !samePath(g.Main, g.Root, runtime.GOOS == "windows") }

// LinksBack reports whether git's own link back holds for a worktree: <its git folder>/gitdir, which git writes when it
// adds the worktree, names this checkout's .git file (as written, relative to that git folder or absolute; or the
// same file reached another way, a link on the path). False for a checkout that is not a worktree.
func (g GitLayout) LinksBack() bool {
	if g.GitDir == "" || !g.Worktree() {
		return false
	}
	line, ok := firstLine(filepath.Join(g.GitDir, "gitdir"))
	named := strings.TrimSpace(line)
	if !ok || named == "" {
		return false
	}
	named = resolve(g.GitDir, named)
	own := filepath.Join(g.Root, ".git")
	if samePath(named, own, runtime.GOOS == "windows") {
		return true
	}
	a, err := os.Stat(named)
	if err != nil {
		return false
	}
	b, err := os.Stat(own)
	return err == nil && os.SameFile(a, b)
}

// Trust is whether a checkout's main checkout is verified (VerifyMain), and when it is not, why and the person's step
// that would verify it.
type Trust struct {
	Verified bool
	Why      string // why not, in ASCII words that follow "its main checkout is not verified: "; never a path outside the home
	Next     string // the step that would verify it, a person's; it may name a path on this machine
}

// notVerified is a Trust that is not verified, with its reason and step.
func notVerified(why, next string) Trust { return Trust{Why: why, Next: next} }

// updateNext is the step that verifies most checkouts: this machine's record written by the main checkout itself.
const updateNext = "a person runs bonsai update in the project's main checkout, which records it on this machine"

// VerifyMain says whether the main checkout of the checkout laid out as g is verified (this file's comment): for a
// worktree, git's link back, the main checkout's id (mainID, its bonsai.yaml's, read by the caller) and this
// machine's record; for a checkout that is its own main, the record alone. id is the checkout's own bonsai.yaml's id;
// home the Bonsai home ("" when there is none). It starts no process and writes nothing.
func VerifyMain(home string, g GitLayout, id, mainID string) Trust {
	switch {
	case g.GitDir == "":
		return notVerified("the folder is not a checkout's top (it holds no .git Bonsai reads)",
			"a project's settings live at a checkout's top: a person moves this bonsai.yaml there, or removes it")
	case id == "":
		return notVerified("its bonsai.yaml holds no workspace id", "a person fixes bonsai.yaml (bonsai check names the line)")
	}
	if g.Worktree() {
		if !g.LinksBack() {
			return notVerified("git's own link back to this worktree (gitdir in its git folder) does not name it, so its .git "+
				"file may point at another checkout", "a person checks the worktree's .git file (git worktree repair, run in the "+
				"main checkout, mends git's links), or makes the worktree again")
		}
		if mainID != id {
			return notVerified("the main checkout its .git file names does not hold this workspace's id in its bonsai.yaml",
				"a person checks the worktree's .git file and the main checkout's bonsai.yaml (bonsai check in each)")
		}
	}
	if home == "" {
		return notVerified("there is no Bonsai home to read this machine's record of the main checkout from",
			"set BONSAI_HOME, or give the user a home folder")
	}
	real, err := realPath(g.Main)
	if err != nil {
		return notVerified("the main checkout's folder cannot be resolved ("+OSWords(err)+")", updateNext)
	}
	key := machineKey(real, runtime.GOOS)
	rec, err := readMachineRecord(filepath.Join(home, "workspaces", key, MachineRecordFile))
	switch {
	case err != nil:
		return notVerified("this machine's record of the main checkout, workspaces/"+key+"/"+MachineRecordFile+
			" in the Bonsai home, cannot be read", nextOf(err))
	case rec == nil:
		return notVerified("this machine has no record of the main checkout (bonsai update has not run in it here)", updateNext)
	case !samePath(filepath.FromSlash(rec.Path), real, runtime.GOOS == "windows"):
		return notVerified("this machine's record of the main checkout names another folder", updateNext)
	case rec.Current() != id:
		return notVerified("this machine's record of the main checkout holds another workspace id (bonsai update has not "+
			"run in it since the id changed)", updateNext)
	}
	first, ties, err := FirstClaim(home, id)
	switch {
	case err != nil:
		return notVerified("a record in the Bonsai home cannot be read, so which checkout on this machine claimed the "+
			"workspace id first cannot be told", nextOf(err))
	case first != nil && first.Key == key:
		return Trust{Verified: true}
	case first != nil:
		return notVerified("another checkout on this machine claimed this workspace's id before its main checkout did",
			rivalNext(home, []Claim{*first}))
	}
	return notVerified("another checkout on this machine claimed this workspace's id in the same second as its main "+
		"checkout, so neither is trusted", rivalNext(home, ties))
}

// rivalNext is a person's step when the first claim on an id is not the main checkout's alone: the checkouts that
// claimed it first, by path and machine folder.
func rivalNext(home string, rivals []Claim) string {
	var paths []string
	for _, r := range rivals {
		paths = append(paths, r.Record.Path+" (machine folder workspaces/"+r.Key+")")
	}
	return "a person checks which checkout is the project's main checkout (first to claim the id on this machine: " +
		strings.Join(paths, " and ") + ", in the Bonsai home " + filepath.ToSlash(home) + "): a copy meant as its own " +
		"project runs bonsai init --new-id in it, a checkout that is gone has its machine folder deleted, and a worktree " +
		"whose .git file points at another main checkout is mended with git worktree repair"
}

// nextOf gives an error's next step, when it is an *Error.
func nextOf(err error) string {
	var we *Error
	if errors.As(err, &we) && we.Next != "" {
		return we.Next
	}
	return "a person checks the Bonsai home's workspaces folder"
}

// realPath is a folder's real path: absolute, symbolic links resolved (MachineKey's path).
func realPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// Local is where a writer in one checkout writes .bonsai/local/ (FindLocal).
type Local struct {
	Root   string // the checkout the writer runs in: its top folder, absolute
	Main   string // the checkout whose .bonsai/local/ is written: the verified main checkout, or Root
	Branch string // Root's branch, read from its HEAD; "" when detached or unreadable
}

// FindLocal finds where a writer of workspace id, running in the checkout whose top folder is root, writes: the main
// checkout its .git names (ReadGitLayout) when it is verified (VerifyMain, with the Bonsai home Home names), else root
// itself. It starts no process and never fails.
func FindLocal(root, id string) Local {
	g := ReadGitLayout(root)
	l := Local{Root: g.Root, Main: g.Root, Branch: g.Branch()}
	if id == "" || !g.Worktree() {
		return l
	}
	home, err := Home()
	if err != nil {
		home = ""
	}
	if VerifyMain(home, g, id, configID(g.Main)).Verified {
		l.Main = g.Main
	}
	return l
}

// Dir is the folder the writer writes: Main's .bonsai/local/.
func (l Local) Dir() string { return filepath.Join(l.Main, filepath.FromSlash(LocalDir)) }

// configID reads only the id of root/bonsai.yaml, "" when the file cannot be read or holds none. It does not hold
// the rest of the file to its fields, so a main checkout whose bonsai.yaml has a problem elsewhere still counts.
func configID(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		return ""
	}
	m, err := readYAML(ConfigFile, raw, WorkspaceFormat, configNext)
	if err != nil {
		return ""
	}
	id, _ := m.Get("id")
	s, _ := id.(string)
	return s
}

// firstLine reads a small file's first line, without its line ending. A file over 4 KiB is not one of git's
// pointer files and is not read.
func firstLine(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer func() { _ = f.Close() }()
	raw, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || len(raw) > 4096 {
		return "", false
	}
	s := string(raw)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSuffix(s, "\r"), true
}

// resolve makes p, as a git file names it, absolute: relative to base when it is not absolute already.
func resolve(base, p string) string {
	p = filepath.FromSlash(p)
	if !filepath.IsAbs(p) {
		p = filepath.Join(base, p)
	}
	return filepath.Clean(p)
}

// sameName compares two names, case-blind when fold is set (Windows).
func sameName(a, b string, fold bool) bool {
	if fold {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// samePath compares two clean paths, case-blind when fold is set (Windows).
func samePath(a, b string, fold bool) bool {
	return sameName(filepath.Clean(a), filepath.Clean(b), fold)
}

// childOf reports whether path is exactly one folder below parent (parent/<name>), case-blind when fold is set.
func childOf(path, parent string, fold bool) bool {
	path, parent = filepath.Clean(path), filepath.Clean(parent)
	return !samePath(path, parent, fold) && samePath(filepath.Dir(path), parent, fold)
}
