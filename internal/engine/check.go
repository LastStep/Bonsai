package engine

// bonsai check's findings, as far as the walking skeleton builds them (plan part 3: "check (lock and files, a tracked
// or staged .bonsai/local/ file)"; spec §6, "bonsai check findings"):
//
//   - bonsai.yaml that Bonsai refuses; the lock missing or refused; bonsai.yaml and the lock naming other packs;
//   - each file the lock lists: a pack file, the block in CLAUDE.md or Bonsai's lines in .claude/settings.json
//     edited or missing (a once or kept file is the project's: no finding); a format-0 file changed;
//   - .bonsai/.gitignore missing or changed; a file from .bonsai/local/ tracked or staged;
//   - plugin drift (plan part 4b, plugins.go): a .claude/settings.local.json that Claude Code reads for this checkout
//     turning on a locked pack's plugin from one of the workspace's marketplaces at other commits than the lock's
//     (checkLocalPlugins, offline; status shows it too); and, from bonsai check only, what Claude Code reports
//     installed (ComparePlugins, which runs claude plugin list): Check alone never runs Claude Code, so status stays
//     cheap and offline (contract §12; spec §5 puts that comparison in check and status --full).
//
// It reads, fetches nothing and writes nothing (CI runs it offline). Bonsai's lines in the settings file are told
// apart with each pack's lines at its locked commit, read from this machine's pack cache; a pack not in the cache
// leaves that file unchecked, with a warning (step 5.1 copies what checks need into the lock's declares).
// Spec §6's other findings and warnings come with step 5.1.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// Finding is one finding or warning: its code, the file it is about, a sentence, the next step and who takes it.
type Finding struct {
	Code    string // config, lock, packs, changed, missing, format0, gitignore, local, cache, plugin
	File    string
	Message string
	Next    string
	Who     string // agent or person (bonsai.check/1's next.who)
}

// Sentence gives the finding as status --json's problems hold it: one sentence naming its next step.
func (f Finding) Sentence() string {
	return ascii(strings.TrimSuffix(f.Message, ".")) + "; next: " + ascii(f.Next)
}

// PackState is a locked pack and whether this checkout's files match it (status --json's packs).
type PackState struct {
	ID, Version, Commit, State string
}

// CheckResult is what check found.
type CheckResult struct {
	Root, Main     string
	Config         *workspace.Config
	Lock           *workspace.Lock
	Findings       []Finding
	Warnings       []Finding
	Packs          []PackState
	Changed        int // files the lock lists that were edited
	Missing        int // files the lock lists that are gone
	Format0Changed int
}

// Check checks the workspace holding dir. Its error is for a folder that is no linked checkout at all (exit 4).
func Check(dir, home string) (*CheckResult, error) {
	co, err := workspace.Find(dir)
	if err != nil {
		return nil, findError(err)
	}
	r := &CheckResult{Root: co.Root, Main: co.Main}
	cfg, err := workspace.LoadConfigFull(co.Root)
	if err != nil {
		var we *workspace.Error
		if errors.As(err, &we) && errors.Is(err, os.ErrNotExist) {
			return nil, fileError(err, "not-linked", "bad-config", ExitInput)
		}
		e := fileError(err, "", "bad-config", ExitInput).(*Error)
		r.Findings = append(r.Findings, Finding{Code: "config", File: workspace.ConfigFile, Message: e.What, Next: e.Next, Who: "person"})
		return r, nil
	}
	r.Config = cfg
	lock, err := workspace.LoadLock(co.Root)
	if err != nil {
		e := fileError(err, "", "bad-lock", ExitState).(*Error)
		r.Findings = append(r.Findings, Finding{Code: "lock", File: workspace.LockFile, Message: e.What, Next: e.Next, Who: "person"})
		r.checkGitignore()
		r.checkLocal()
		return r, nil
	}
	r.Lock = lock

	// bonsai.yaml and the lock.
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
			r.find("packs", workspace.ConfigFile, "person", "bonsai.yaml lists the pack "+ref.ID+", which the lock does not hold", "run bonsai update")
		case lp.Source != ref.Source:
			r.find("packs", workspace.ConfigFile, "person", "bonsai.yaml takes "+ref.ID+" from "+ref.Source+", the lock from "+lp.Source, "run bonsai update")
		case commitPattern.MatchString(ref.Ref) && ref.Ref != lp.Commit:
			r.find("packs", workspace.ConfigFile, "person", "bonsai.yaml's ref for "+ref.ID+" is "+short(ref.Ref)+", the lock holds "+short(lp.Commit), "run bonsai update")
		}
	}
	for _, lp := range lock.Packs {
		if !listed[lp.ID] {
			r.find("packs", workspace.LockFile, "person", "the lock holds the pack "+lp.ID+", which bonsai.yaml does not list", "put it back in bonsai.yaml (taking a pack out comes with step 5.1)")
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
			raw, exists, err := readFile(co.Root, path)
			switch {
			case err != nil:
				r.find("missing", path, "agent", err.Error(), "check the file's permissions")
			case !exists:
				r.Missing++
				mark(lf.Pack, "missing")
				r.find("missing", path, "person", path+" is missing (a file of the pack "+lf.Pack+")", "run bonsai update --yes to write it again")
			case workspace.HashLF(raw) != lf.SHA256:
				r.Changed++
				mark(lf.Pack, "changed")
				r.find("changed", path, "person", path+" was edited (a file of the pack "+lf.Pack+", which an edit makes a conflict at the pack's next change)",
					"keep the edit: bonsai update --yes --keep "+path+"; or take the pack's copy back: bonsai update --yes --adopt "+path)
			}
		case "block":
			bd, err := readBlock(co.Root)
			switch {
			case err != nil:
				r.Changed++
				mark(lf.Pack, "changed")
				e := err.(*Error)
				r.find("changed", path, "person", e.What, e.Next)
			case !bd.found:
				r.Missing++
				mark(lf.Pack, "missing")
				r.find("missing", path, "person", "Bonsai's block in "+path+" is missing", "run bonsai update --yes to write it again")
			case regionHash(bd.region()) != lf.SHA256:
				r.Changed++
				mark(lf.Pack, "changed")
				r.find("changed", path, "person", "Bonsai's block in "+path+" was edited", "take Bonsai's block back: bonsai update --yes --adopt "+path+" (your copy is saved in the Bonsai home)")
			}
		case "keys":
			r.checkSettings(home, cfg, lock, lf, mark)
		}
	}
	r.checkLocalPlugins()
	for _, path := range sortedPaths(lock.Format0) {
		raw, exists, _ := readFile(co.Root, path)
		if !exists || workspace.HashLF(raw) != lock.Format0[path] {
			r.Format0Changed++
			r.find("format0", path, "agent", path+" is a format-0 file the lock fixes, and it changed", "restore it from git (git checkout -- "+path+")")
		}
	}
	for _, lp := range lock.Packs {
		s := state[lp.ID]
		if s == "" {
			s = "ok"
		}
		r.Packs = append(r.Packs, PackState{ID: lp.ID, Version: lp.Version, Commit: lp.Commit, State: s})
	}
	r.checkGitignore()
	r.checkLocal()
	return r, nil
}

func (r *CheckResult) find(code, file, who, msg, next string) {
	r.Findings = append(r.Findings, Finding{Code: code, File: file, Message: msg, Next: next, Who: who})
}

func (r *CheckResult) checkSettings(home string, cfg *workspace.Config, lock *workspace.Lock, lf workspace.LockedFile, mark func(string, string)) {
	sd, err := readSettings(r.Root)
	if err != nil {
		e := err.(*Error)
		r.Changed++
		mark(lf.Pack, "changed")
		r.find("changed", SettingsFile, "person", e.What, e.Next)
		return
	}
	if !sd.exists {
		r.Missing++
		mark(lf.Pack, "missing")
		r.find("missing", SettingsFile, "person", SettingsFile+" is missing, with Bonsai's lines in it", "run bonsai update --yes to write Bonsai's lines again")
		return
	}
	c := cache{home: home}
	folders := map[string]string{}
	for _, ref := range cfg.Packs {
		folders[ref.ID] = ref.Path
	}
	var pls []packLines
	for _, lp := range lock.Packs {
		pl := packLines{id: lp.ID, source: lp.Source, folder: folders[lp.ID], commit: lp.Commit}
		if !c.has(lp.Source, lp.Commit) {
			r.Warnings = append(r.Warnings, Finding{Code: "cache", File: SettingsFile,
				Message: "Bonsai's lines in " + SettingsFile + " were not checked: the pack " + lp.ID + " at " + short(lp.Commit) + " is not in this machine's pack cache",
				Next:    "run bonsai update once on this machine (it fetches the pack and writes nothing when nothing changed), then check again",
				Who:     "agent"})
			return
		}
		pd, err := c.packAt(workspace.PackRef{ID: lp.ID, Source: lp.Source, Path: folders[lp.ID]}, lp.Commit)
		if err != nil {
			e := err.(*Error)
			r.Warnings = append(r.Warnings, Finding{Code: "cache", File: SettingsFile, Message: e.What, Next: e.Next, Who: "person"})
			return
		}
		pl.known, pl.hooks, pl.deny = true, pd.Manifest.Hooks, pd.Manifest.Deny
		pls = append(pls, pl)
	}
	_, same := claim(diskLines(sd.root), buildLines(cfg, pls), false, lf.SHA256)
	if !same {
		r.Changed++
		mark(lf.Pack, "changed")
		r.find("changed", SettingsFile, "person", "Bonsai's lines in "+SettingsFile+" were edited, or bonsai.yaml changed since the last update",
			"run bonsai update to see the lines; to take Bonsai's lines back: bonsai update --yes --adopt "+SettingsFile+" (your copy is saved in the Bonsai home)")
	}
}

func (r *CheckResult) checkGitignore() {
	raw, exists, _ := readFile(r.Root, GitignoreFile)
	switch {
	case !exists:
		r.find("gitignore", GitignoreFile, "person", GitignoreFile+" is missing, so .bonsai/local/ could be committed", "run bonsai update --yes to write it again")
	case workspace.HashLF(raw) != workspace.HashLF([]byte(GitignoreText)):
		r.find("gitignore", GitignoreFile, "person", GitignoreFile+" was changed", "run bonsai update --yes to write Bonsai's copy again")
	}
}

// checkLocal finds files from .bonsai/local/ in git's index: tracked, or staged to be (contract §3).
func (r *CheckResult) checkLocal() {
	cmd := exec.Command("git", "-C", r.Root, "ls-files", "-z", "--", LocalDir)
	cmd.Env = gitEnv()
	out, err := cmd.Output()
	if err != nil {
		r.Warnings = append(r.Warnings, Finding{Code: "local", File: LocalDir, Message: "git ls-files failed, so tracked .bonsai/local/ files were not looked for",
			Next: "check that git works in this checkout (git status), then check again", Who: "agent"})
		return
	}
	var files []string
	for _, f := range bytes.Split(out, []byte{0}) {
		if len(f) > 0 {
			files = append(files, filepath.ToSlash(string(f)))
		}
	}
	if len(files) > 0 {
		r.find("local", LocalDir, "agent", "git tracks or has staged "+strings.Join(files, ", ")+" from .bonsai/local/, which is never committed",
			"git rm -r --cached -- .bonsai/local, then commit")
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
