package engine

// bonsai check's findings and warnings (spec section 6, "bonsai check findings" and "Warnings"; step 5.1.6). Every
// one has a word in format.CheckWords, their one home, which says whether it is a finding (exit 1, and a problem in
// status --json) or a warning (never the exit code), and who usually takes its next step. Each next step is the exact
// command that fixes it, written "run: <command>", wherever one exists, and the step in plain words where a person
// must judge (Rohan, 9 Oct: Bonsai is run by agents inside projects); cmd/bonsai's TestCheckTable holds every command
// a next step names to Bonsai's words and flags.
//
// Check reads, fetches nothing and writes nothing (CI runs it offline, spec section 5: "CI needs no pack"): the lock
// alone gives every finding, with an empty pack cache and no network. What it looks at, in this order:
//
//   - check.go: bonsai.yaml and the lock (config, lock, packs); each file the lock lists (changed, missing; a once or
//     kept file is the project's); plugin drift in the local settings files (plugin, plugins.go); the lock's format0
//     list (format0); .bonsai/.gitignore and .bonsai/local/ in git's index (gitignore, local);
//   - checkdocs.go: the documents of Bonsai's kinds (format0-new, document, label, absolute-path), the labels in force
//     (label-twice), the budgets (block-size, memory-index-size, memory-note-size), the paths CLAUDE.md, STATE and the
//     memory notes name (missing-path), and the run reports past generated.run's rule (run-reports);
//   - checkhistory.go: approve_first from git history (approve-first, approve-first-unchecked);
//   - checksettings.go: Claude Code's settings in the project (settings-rule, hooks-off, plugin-version);
//   - checkmachine.go: this machine's record of the checkout and the PATH (id-changed, same-id, bonsai-path).
//
// Two more need Claude Code, so cmd/bonsai runs them after Check and status never does (status stays cheap and
// offline, contract section 12): what Claude Code reports installed (ComparePlugins, plugins.go: plugin, plugin-missing,
// plugin-unchecked) and its version against the floor (CompareClaudeCode, claude.go: claude-code-old,
// claude-code-unknown). checkLater names the three that later steps build.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Finding is one finding or warning: its word (format.CheckWords), the file it is about, a sentence, the next step
// and who takes it.
type Finding struct {
	Code    string // a word of format.CheckWords: its Kind says finding or warning
	File    string // project-relative, forward slashes; "" for none
	Message string
	Next    string // "run: <command>" where one command fixes it, else the step in plain words (which may name commands)
	Who     string // agent or person (bonsai.check/1's next.who)
}

// Sentence gives the finding as status --json's problems hold it: one sentence naming its next step.
func (f Finding) Sentence() string {
	return ascii(strings.TrimSuffix(f.Message, ".")) + "; next: " + ascii(f.Next)
}

// checkLater are spec section 6's findings and warnings that later steps build, each with its step: their words join
// format.CheckWords when they are built (TestCheckTable holds that none is there before).
var checkLater = []struct{ Code, Step, What string }{
	{"tables", "step 5.1.8", "a stale tasks table: a warning, never the exit code (check --write rebuilds it)"},
	{"secret", "step 5.2", "a secret-shaped string in a committed memory note: the redactor's patterns are its one home"},
	{"stranded", "step 5.6", "a machine folder stranded under an old path (contract section 3)"},
}

// run writes a next step that is one command: "run: <command>" (bonsai.check/1's next.do).
func run(command string) string { return "run: " + command }

// PackState is a locked pack and whether this checkout's files match it (status --json's packs).
type PackState struct {
	ID, Version, Commit, State string
}

// CheckResult is what check found.
type CheckResult struct {
	Root, Main     string
	Home           string
	Config         *workspace.Config
	Lock           *workspace.Lock
	Findings       []Finding
	Warnings       []Finding
	Notes          []string // said to a person in check's text, never a finding or a warning (no --json field)
	Packs          []PackState
	Changed        int // files the lock lists that were edited
	Missing        int // files the lock lists that are gone
	Format0Changed int

	history bool // CheckOptions.History
}

// Check checks the workspace holding dir, every finding and warning but the two that ask Claude Code. Its error is
// for a folder that is no linked checkout at all (exit 4).
func Check(dir, home string) (*CheckResult, error) { return CheckWith(dir, home, CheckOptions{History: true}) }

// CheckOptions say what Check may leave out.
type CheckOptions struct {
	// History: read git history for approve_first (checkhistory.go). status's default leaves it out to stay cheap
	// for a program that runs it often (contract §12): it reads every commit that touched the task folder.
	History bool
}

// CheckWith is Check with its options.
func CheckWith(dir, home string, o CheckOptions) (*CheckResult, error) {
	co, err := workspace.Find(dir)
	if err != nil {
		return nil, findError(err)
	}
	r := &CheckResult{Root: co.Root, Main: co.Main, Home: home, history: o.History}
	cfg, err := workspace.LoadConfigFull(co.Root)
	if err != nil {
		var we *workspace.Error
		if errors.As(err, &we) && errors.Is(err, os.ErrNotExist) {
			return nil, fileError(err, "not-linked", "bad-config", ExitInput)
		}
		e := fileError(err, "", "bad-config", ExitInput).(*Error)
		r.add("config", workspace.ConfigFile, "", e.What, e.Next)
		return r, nil
	}
	r.Config = cfg
	lock, err := workspace.LoadLock(co.Root)
	if err != nil {
		e := fileError(err, "", "bad-lock", ExitState).(*Error)
		next := r.missingLockNext()
		if !errors.Is(err, os.ErrNotExist) {
			next = "the lock is Bonsai's to write: to restore the committed one, " + lockNext + "; if that one is refused too, " +
				"a person deletes it and links the project again from bonsai.yaml (bonsai init --yes previews every file first)"
		}
		r.add("lock", workspace.LockFile, "", e.What, next)
		r.checkProject()
		return r, nil
	}
	r.Lock = lock
	r.checkLockedFiles()
	r.checkProject()
	return r, nil
}

// missingLockNext is the next step for a missing lock: git's copy when git has one, else a link again from bonsai.yaml
// (update refuses with no lock: it cannot tell which hook lines were consented to).
func (r *CheckResult) missingLockNext() string {
	cmd := exec.Command("git", "-C", r.Root, "cat-file", "-e", "HEAD:"+workspace.LockFile)
	cmd.Env = gitEnv()
	if cmd.Run() == nil {
		return run("git checkout -- " + workspace.LockFile)
	}
	return "link the project again from bonsai.yaml (a pack's code also needs --allow-exec, a person's consent, which init names), run: bonsai init --yes"
}

// checkProject runs the checks that need no lock: the documents, the settings files, this machine's record and the
// PATH, .bonsai/.gitignore and .bonsai/local/.
func (r *CheckResult) checkProject() {
	r.checkDocuments()
	if r.history {
		r.checkHistory()
	}
	r.checkSettingsFiles()
	r.checkMachine()
	r.checkGitignore()
	r.checkLocal()
}

// checkLockedFiles checks bonsai.yaml against the lock, each file the lock lists, the local settings files' plugins
// and the lock's format0 list.
func (r *CheckResult) checkLockedFiles() {
	cfg, lock := r.Config, r.Lock
	locked := map[string]workspace.LockedPack{}
	for _, lp := range lock.Packs {
		locked[lp.ID] = lp
	}
	listed := map[string]bool{}
	for _, ref := range cfg.Packs {
		listed[ref.ID] = true
		lp, ok := locked[ref.ID]
		switch {
		case !ok:
			r.add("packs", workspace.ConfigFile, "", "bonsai.yaml lists the pack "+ref.ID+", which the lock does not hold", run("bonsai update --yes"))
		case lp.Source != ref.Source:
			r.add("packs", workspace.ConfigFile, "", "bonsai.yaml takes "+ref.ID+" from "+ref.Source+", the lock from "+lp.Source, run("bonsai update --yes"))
		case commitPattern.MatchString(ref.Ref) && ref.Ref != lp.Commit:
			r.add("packs", workspace.ConfigFile, "", "bonsai.yaml's ref for "+ref.ID+" is "+short(ref.Ref)+", the lock holds "+short(lp.Commit), run("bonsai update --yes"))
		case lp.PathSet && lp.Path != ref.Path:
			r.add("packs", workspace.ConfigFile, "", "bonsai.yaml takes "+ref.ID+" from the folder "+folderName(ref.Path)+
				" of its repository, the lock from "+folderName(lp.Path)+" (update judges the new folder's code as at a first link)", run("bonsai update --yes"))
		}
	}
	for _, lp := range lock.Packs {
		if !listed[lp.ID] {
			r.add("packs", workspace.LockFile, "", "the lock holds the pack "+lp.ID+", which bonsai.yaml does not list",
				"put the pack back in bonsai.yaml for now: taking a pack out of a project comes with step 5.1.7")
		}
	}

	// Each file the lock lists.
	state := map[string]string{}
	mark := func(pack, s string) {
		if state[pack] != "missing" {
			state[pack] = s
		}
	}
	for _, path := range sortedPaths(lock.Files) {
		lf := lock.Files[path]
		switch lf.Kind {
		case "pack":
			raw, exists, err := readFile(r.Root, path)
			switch {
			case err != nil:
				r.add("missing", path, "agent", err.Error(), "check the file's permissions, then run: bonsai check")
			case !exists:
				r.Missing++
				mark(lf.Pack, "missing")
				r.add("missing", path, "", path+" is missing (a file of the pack "+lf.Pack+")", run("bonsai update --yes"))
			case workspace.HashLF(raw) != lf.SHA256:
				r.Changed++
				mark(lf.Pack, "changed")
				r.add("changed", path, "", path+" was edited (a file of the pack "+lf.Pack+", which an edit makes a conflict at the pack's next change)",
					"to keep the edit, run: bonsai update --yes --keep "+ShellArg(path)+"; to take the pack's copy back, run: bonsai update --yes --adopt "+ShellArg(path))
			}
		case "block":
			bd, err := readBlock(r.Root)
			switch {
			case err != nil:
				r.Changed++
				mark(lf.Pack, "changed")
				e := err.(*Error)
				r.add("changed", path, "", e.What, e.Next+"; or, to write Bonsai's block again, run: bonsai update --yes --adopt "+path)
			case !bd.found:
				r.Missing++
				mark(lf.Pack, "missing")
				r.add("missing", path, "", "Bonsai's block in "+path+" is missing", run("bonsai update --yes"))
			case regionHash(bd.region()) != lf.SHA256:
				r.Changed++
				mark(lf.Pack, "changed")
				r.add("changed", path, "", "Bonsai's block in "+path+" was edited (your copy is saved in the Bonsai home when you take Bonsai's back)",
					run("bonsai update --yes --adopt "+path))
			}
		case "keys":
			r.checkSettings(lf, mark)
		}
	}
	r.checkLocalPlugins()
	for _, path := range sortedPaths(lock.Format0) {
		raw, exists, _ := readFile(r.Root, path)
		switch {
		case !exists:
			r.Format0Changed++
			r.add("format0", path, "", path+" is a format-0 file the lock fixes, and it is gone", run("git checkout -- "+ShellArg(path)))
		case workspace.HashLF(raw) != lock.Format0[path] && isFormat0(raw):
			r.Format0Changed++
			kind := format0Kind(r.Config, path)
			r.add("format0", path, "", path+" is a format-0 file the lock fixes, and it changed (contract section 2.3: a format-0 file is "+
				"never rewritten)", "to undo the change, run: git checkout -- "+ShellArg(path)+"; to change it, give it a format: line first "+
				"(format: bonsai."+kind+"/1, with the fields bonsai check --schema bonsai."+kind+" lists)")
		}
	}
	for _, lp := range lock.Packs {
		s := state[lp.ID]
		if s == "" {
			s = "ok"
		}
		r.Packs = append(r.Packs, PackState{ID: lp.ID, Version: lp.Version, Commit: lp.Commit, State: s})
	}
}

// folderName names a pack's folder in its repository for a person: the folder, or "(the repository's root)".
func folderName(folder string) string {
	if folder == "" {
		return "(the repository's root)"
	}
	return folder
}

// add adds a finding or a warning, as its word's Kind in format.CheckWords says; who "" takes the word's usual one.
// A word that is not in the table is a bug, which cmd/bonsai's TestCheckWordsInTheCode rules out.
func (r *CheckResult) add(code, file, who, msg, next string) {
	w, ok := format.CheckWord(code)
	if !ok {
		panic("engine: the check word " + code + " is not in format.CheckWords")
	}
	if who == "" {
		who = w.Who
	}
	f := Finding{Code: code, File: file, Message: msg, Next: next, Who: who}
	if w.Kind == "warning" {
		r.Warnings = append(r.Warnings, f)
		return
	}
	r.Findings = append(r.Findings, f)
}

func (r *CheckResult) checkSettings(lf workspace.LockedFile, mark func(string, string)) {
	cfg, lock := r.Config, r.Lock
	sd, err := readSettings(r.Root)
	if err != nil {
		e := err.(*Error)
		r.Changed++
		mark(lf.Pack, "changed")
		r.add("changed", SettingsFile, "", e.What, e.Next)
		return
	}
	if !sd.exists {
		r.Missing++
		mark(lf.Pack, "missing")
		r.add("missing", SettingsFile, "", SettingsFile+" is missing, with Bonsai's lines in it", run("bonsai update --yes"))
		return
	}
	c := cache{home: r.Home}
	folders := map[string]string{}
	for _, ref := range cfg.Packs {
		folders[ref.ID] = ref.Path
	}
	var pls []packLines
	for _, lp := range lock.Packs {
		folder := lockedFolder(lp, folders[lp.ID])
		pl := packLines{id: lp.ID, source: lp.Source, folder: folder, commit: lp.Commit}
		if lp.PathSet {
			// The lock's declares holds the pack's hook lines and deny rules (ReadLock read them): no pack is read.
			d, err := lp.Declared()
			if err != nil {
				r.add("cache", SettingsFile, "", err.Error(), lockNext)
				return
			}
			pl.known, pl.hooks, pl.deny = true, declaredHooks(d), declaredDeny(d)
			pls = append(pls, pl)
			continue
		}
		if !c.has(lp.Source, lp.Commit) {
			r.add("cache", SettingsFile, "",
				"Bonsai's lines in "+SettingsFile+" were not checked: the pack "+lp.ID+" at "+short(lp.Commit)+" is not in this machine's pack "+
					"cache, and the lock (written before formats set 4) does not hold its hook lines",
				"write the lock again with the pack's hook lines in it, run: bonsai update --yes")
			return
		}
		pd, err := c.packAt(workspace.PackRef{ID: lp.ID, Source: lp.Source, Path: folder}, lp.Commit)
		if err != nil {
			e := err.(*Error)
			r.add("cache", SettingsFile, "", e.What, e.Next)
			return
		}
		pl.known, pl.hooks, pl.deny = true, pd.Manifest.Hooks, pd.Manifest.Deny
		pls = append(pls, pl)
	}
	_, same := claim(diskLines(sd.root), buildLines(cfg, pls), false, lf.SHA256)
	if !same {
		r.Changed++
		mark(lf.Pack, "changed")
		r.add("changed", SettingsFile, "", "Bonsai's lines in "+SettingsFile+" were edited, or bonsai.yaml changed since the last update",
			"to see the lines, run: bonsai update; to take Bonsai's lines back (your copy is saved in the Bonsai home), run: bonsai update --yes --adopt "+SettingsFile)
	}
}

func (r *CheckResult) checkGitignore() {
	raw, exists, _ := readFile(r.Root, GitignoreFile)
	switch {
	case !exists:
		r.add("gitignore", GitignoreFile, "", GitignoreFile+" is missing, so .bonsai/local/ could be committed", run("bonsai update --yes"))
	case workspace.HashLF(raw) != workspace.HashLF([]byte(GitignoreText)):
		r.add("gitignore", GitignoreFile, "", GitignoreFile+" was changed, so .bonsai/local/ could be committed", run("bonsai update --yes"))
	}
}

// checkLocal finds files from .bonsai/local/ in git's index: tracked, or staged to be (contract §3).
func (r *CheckResult) checkLocal() {
	cmd := exec.Command("git", "-C", r.Root, "ls-files", "-z", "--", LocalDir)
	cmd.Env = gitEnv()
	out, err := cmd.Output()
	if err != nil {
		r.add("local-unchecked", "", "", "git ls-files failed, so tracked .bonsai/local/ files were not looked for",
			"check that git works in this checkout, run: git status")
		return
	}
	var files []string
	for _, f := range bytes.Split(out, []byte{0}) {
		if len(f) > 0 {
			files = append(files, filepath.ToSlash(string(f)))
		}
	}
	if len(files) > 0 {
		r.add("local", LocalDir, "", "git tracks or has staged "+strings.Join(files, ", ")+" from .bonsai/local/, which is never committed",
			"take them out of git's index (the files stay on disk), run: git rm -r -q --cached -- "+LocalDir+"; then commit")
	}
}

func sortedPaths[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
