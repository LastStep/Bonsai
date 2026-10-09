package engine

// What init and update would do (spec §6, "How update decides, per file"), worked out before anything is written:
// the preview prints it, and apply.go writes it only when nothing stops it.
//
// Per file, three fingerprints: what the lock says Bonsai last wrote, what is on disk, what the pack now gives.
//
//	on disk vs locked   new vs locked   result
//	same                same            unchanged
//	same                different       updated
//	different           same            changed: left alone; check reports it
//	different, = new    different       adopted: the lock catches up
//	different           different       conflict: exit 5, nothing written, the file named
//	missing             any             kind pack (and block, keys): restored, written again; kind once: left missing
//
// Beside the table (choices of this build, where spec §6 is silent):
//   - At a first link (no lock entry), a pack file the project already holds is adopted when it equals the pack's
//     copy and a conflict otherwise; a once file it already holds is found, left as it is and locked as found; the
//     block and Bonsai's settings lines are written on --yes (the link is the person's consent to them), but what
//     runs code needs --allow-exec as well (consent.go: a pack's hook lines, the files they run, a plugin's code).
//   - kept: the lock holds the pack's copy the person chose not to take. While the pack's copy stays that, the file
//     is kept; when the pack changes it again, it is a conflict again (spec §6), unless the file now equals it.
//   - A file the pack no longer has: kind pack is removed when unedited, a conflict when edited, and dropped from the
//     lock when already gone; kinds once and kept are released, left to the project.
//   - --keep P (a conflict or an edited pack file): the file stays as the person has it, kind kept. --adopt P: the
//     pack's copy is written (replaced) and the project's copy saved in the home's cache, never in the repo. Neither
//     keeps Bonsai's own part of CLAUDE.md or .claude/settings.json: --adopt takes Bonsai's, --keep is refused.
//   - .bonsai/.gitignore is Bonsai's (spec §6): written when missing or changed, never in the lock.
//   - The lock names one pack per file. CLAUDE.md's block and .claude/settings.json hold lines from every pack and
//     from Bonsai itself; their entries name the first pack in bonsai.yaml's order, as contract §14's example names
//     base for the settings file.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The results a file can have in a plan.
const (
	Unchanged   = "unchanged"
	Updated     = "updated"
	Created     = "created"
	Adopted     = "adopted"
	Changed     = "changed"
	Conflict    = "conflict"
	Restored    = "restored"
	LeftMissing = "left missing"
	Found       = "found"
	Kept        = "kept"
	Replaced    = "replaced"
	Removed     = "removed"
	Released    = "released"
	Dropped     = "dropped"
)

// GitignoreFile and GitignoreText: .bonsai/.gitignore, written by Bonsai (spec §6, contract §3).
const (
	GitignoreFile = ".bonsai/.gitignore"
	GitignoreText = "# .bonsai/.gitignore: written by Bonsai (spec section 6). .bonsai/local/ holds this checkout's log, questions\n" +
		"# and ladder results, never committed. Bonsai writes this file again when it is missing or changed.\n" +
		"local/\n"
)

// LocalDir is the project's folder that is never committed (contract §3).
const LocalDir = ".bonsai/local"

// Request is what a person asked init or update to do.
type Request struct {
	Command string      // "init" or "update"
	Dir     string      // the folder the command runs in
	Home    string      // the Bonsai home
	Version string      // this Bonsai's version, for the lock's written_by
	Init    *InitValues // init: the values for a new bonsai.yaml; nil to take the one there
	NewID   bool        // init --new-id
	Keep    []string    // --keep paths
	Adopt   []string    // --adopt paths
	// AllowExec is --allow-exec: the person consents to what the plan's "Runs code" lists (consent.go). Without it
	// Apply refuses a plan that lists anything there.
	AllowExec bool
}

// PackMove is one pack in a plan: its commit before (none at a first link) and after.
type PackMove struct {
	ID, Source, Version, From, To string
}

// FileResult is one file in a plan.
type FileResult struct {
	Path   string // project-relative, forward slashes
	Kind   string // pack, once, block, keys or kept; "" for Bonsai's own files outside the lock
	Pack   string // the pack it comes from; "" for none
	Result string
	Why    string // a few words on the result, for a person
	Saved  string // --adopt: where the project's copy was saved (set by Apply)

	write  []byte // the new bytes, when the file is written
	remove bool   // the file is removed
	old    []byte // the bytes on disk, nil when missing
	backup bool   // --adopt: the project's copy is saved first
	newH   string // the pack's (or Bonsai's) new fingerprint
}

// Writes reports whether the plan writes or removes the file.
func (f *FileResult) Writes() bool { return f.write != nil || f.remove }

// Plan is what init or update would do.
type Plan struct {
	Command    string
	Root       string // the checkout's top folder (system separators)
	Main       string // the main checkout's
	Home       string
	Config     *workspace.Config
	FirstLink  bool // no lock yet
	NewConfig  bool // init writes a new bonsai.yaml
	OldID      string
	Packs      []PackMove
	Files      []*FileResult
	Settings   []SettingsChange // the settings lines the plan adds, changes or removes (when it writes the file)
	Conflicts  []*FileResult
	RunsCode   []CodeItem       // what the plan writes that runs code: it needs --allow-exec as well as --yes (consent.go)
	OwnHooks   []SettingsChange // at a first link, Bonsai's own hook lines the plan adds, which --yes writes (rule 6)
	LeftHooks  []string         // a link again with the lock missing, or an unverified pack: hook lines left as the project's own
	Unverified []string         // locked packs whose lock does not match their commit and folder: no baseline for consent
	AllowExec  bool             // the request's --allow-exec
	EmptyLocal int              // init --new-id: files in .bonsai/local/ to remove; -1 for none to remove
	LockWrite  bool

	Plugins []PluginResult // what InstallPlugins did after the plan was written (or had nothing to write)

	pluginCode map[string][]CodePart // each pack's plugin code parts at its new commit, for the install step

	lock      *workspace.Lock
	lockBytes []byte
}

// PluginConsent is what the install step needs after the plan (InstallPlugins): each pack's plugin code parts at the
// commit the plan locks, the request's --allow-exec, and again, the command that repeats the run with --allow-exec.
func (p *Plan) PluginConsent(again string) PluginConsent {
	return PluginConsent{Code: p.pluginCode, AllowExec: p.AllowExec, Again: again}
}

// NewLock is the lock the plan writes (or, with nothing to change, the lock as it is).
func (p *Plan) NewLock() *workspace.Lock { return p.lock }

// Nothing reports whether the plan changes nothing on disk.
func (p *Plan) Nothing() bool {
	if p.LockWrite || p.EmptyLocal > 0 {
		return false
	}
	for _, f := range p.Files {
		if f.Writes() {
			return false
		}
	}
	return true
}

// fileError gives a refusal of one of Bonsai's files (internal/workspace's Error) as an Error with its word: a file
// that is not there is missing (exit 4), one that cannot be read is read-failed (exit 3), and one Bonsai refuses is
// refused, at exit. An error that is already an *Error is kept.
func fileError(err error, missing, refused string, exit int) error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	var we *workspace.Error
	if !errors.As(err, &we) {
		return Unexpected(err)
	}
	var pe *fs.PathError
	switch {
	case missing != "" && errors.Is(err, fs.ErrNotExist):
		return wsError(we, missing, ExitState)
	case errors.As(err, &pe):
		return wsError(we, "read-failed", ExitRuntime)
	}
	return wsError(we, refused, exit)
}

// findError gives workspace.Find's error with its word: git missing, a folder in no checkout (exit 4), or one that
// cannot be read.
func findError(err error) error {
	var we *workspace.Error
	if !errors.As(err, &we) {
		return Unexpected(err)
	}
	switch {
	case we.File == "git":
		return wsError(we, "git-missing", ExitRuntime)
	case we.Msg == "is not inside a git checkout" || strings.HasPrefix(we.Msg, "git rev-parse failed"):
		return wsError(we, "not-a-checkout", ExitState)
	case strings.HasPrefix(we.Msg, "git rev-parse gave"):
		return wsError(we, "unexpected", ExitRuntime)
	}
	return wsError(we, "read-failed", ExitRuntime)
}

// FindError is findError for status, which finds the checkout itself.
func FindError(err error) *Error { return findError(err).(*Error) }

// ConfigError gives a refusal of bonsai.yaml with its word, for status, which reads it itself: not-linked when it
// is not there, read-failed, else bad-config.
func ConfigError(err error) *Error {
	return fileError(err, "not-linked", "bad-config", ExitInput).(*Error)
}

// HomeError gives workspace.Home's error with its word.
func HomeError(err error) *Error {
	var we *workspace.Error
	if errors.As(err, &we) {
		return wsError(we, "bad-home", ExitRuntime)
	}
	e := Unexpected(err)
	e.Code = "bad-home"
	return e
}

func wsError(we *workspace.Error, code string, exit int) *Error {
	what := we.File + ": " + strings.TrimSuffix(we.Msg, ".")
	if we.Line > 0 {
		what = we.File + " line " + strconv.Itoa(we.Line) + ": " + we.Msg
	}
	if we.Code != "" {
		what += " (" + we.Code + ")"
	}
	return &Error{Code: code, Exit: exit, What: what, Next: we.Next}
}

// whose names who takes the error's next step when it is not the word's usual one.
func (e *Error) whose(who string) *Error {
	e.Who = who
	return e
}

func readFile(root, rel string) ([]byte, bool, error) {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, errorf("read-failed", ExitRuntime, "check the file's permissions, then run the command again", "%s cannot be read: %v", rel, err)
	}
	return b, true, nil
}

// Build works out the plan for a request. It fetches each pack (the network), and writes nothing in the project.
// Its error is an *Error; once bonsai.yaml was read, the error names the workspace (Error.Workspace).
func Build(req Request) (_ *Plan, err error) {
	co, err := workspace.Find(req.Dir)
	if err != nil {
		return nil, findError(err)
	}
	p := &Plan{Command: req.Command, Root: co.Root, Main: co.Main, Home: req.Home, EmptyLocal: -1, AllowExec: req.AllowExec}
	defer func() {
		if err == nil {
			return
		}
		e := Unexpected(err)
		if p.Config != nil && e.Workspace == nil {
			e.Workspace = p.WorkspaceRef()
		}
		err = e
	}()
	for _, old := range []string{".bonsai.yaml", ".bonsai-lock.yaml"} {
		if _, err := os.Stat(filepath.Join(co.Root, old)); err == nil {
			return nil, errorf("old-workspace", ExitState, "leave this project on Bonsai 0.4.3 for now: moving a 0.4.3 workspace to the new Bonsai comes later",
				"this is a Bonsai 0.4.3 workspace (%s), which this Bonsai does not read or change", old)
		}
	}
	c := cache{home: req.Home}

	// bonsai.yaml.
	rawConfig, hasConfig, err := readFile(co.Root, workspace.ConfigFile)
	if err != nil {
		return nil, err
	}
	var cfg *workspace.Config
	var newPacks []*PackData
	switch {
	case hasConfig:
		if cfg, err = workspace.ReadConfigFull(rawConfig); err != nil {
			return nil, fileError(err, "", "bad-config", ExitInput)
		}
		if req.Init != nil && !sameValues(cfg, req.Init) {
			return nil, errorf("values-differ", ExitInput, "edit bonsai.yaml for the change, then run bonsai update; or run bonsai init without --name, --source, --ref, --path and --never-edit",
				"bonsai.yaml is already here and says otherwise than init's values")
		}
	case req.Command == "update":
		return nil, errorf("not-linked", ExitState, "link the project first: bonsai init --name <name> --source <pack repository> --ref <tag or commit>",
			"this checkout has no bonsai.yaml, so it is not linked to Bonsai")
	default:
		if missing := missingValues(req.Init); len(missing) > 0 {
			return nil, errorf("missing-value", ExitInput, "run bonsai init --name <name> --source <pack repository> --ref <tag or commit> (bonsai init --help shows every flag)",
				"init needs %s to write bonsai.yaml", andList(missing))
		}
		if err := checkValues(req.Init); err != nil {
			return nil, err
		}
		ref := workspace.PackRef{Source: req.Init.Source, Path: req.Init.Path, Ref: req.Init.Ref}
		commit, err := c.fetch(ref.Source, ref.Ref)
		if err != nil {
			return nil, err
		}
		c.recordResolved(ref.Source, ref.Ref, commit)
		pd, err := c.packAt(ref, commit)
		if err != nil {
			return nil, err
		}
		id, err := NewID()
		if err != nil {
			return nil, errorf("unexpected", ExitRuntime, "run the command again", "no random id: %v", err)
		}
		rawConfig = ConfigYAML(id, *req.Init, pd.Manifest.ID)
		if cfg, err = workspace.ReadConfigFull(rawConfig); err != nil {
			return nil, fileError(err, "", "bad-value", ExitInput)
		}
		pd.Ref = cfg.Packs[0]
		newPacks = []*PackData{pd}
		p.NewConfig = true
	}
	p.Config = cfg
	origConfig := rawConfig
	if req.NewID {
		if co.Root != co.Main {
			return nil, errorf("not-main-checkout", ExitState, "run bonsai init --new-id in the main checkout, "+filepath.ToSlash(co.Main),
				"init --new-id gives a copied project its own id; a worktree shares its main checkout's id").whose("person")
		}
		if hasConfig {
			id, err := NewID()
			if err != nil {
				return nil, errorf("unexpected", ExitRuntime, "run the command again", "no random id: %v", err)
			}
			if rawConfig, err = replaceID(rawConfig, cfg, id); err != nil {
				return nil, errorf("bad-config", ExitInput, "put the id on a line of its own (id: ws-...), then run the command again", "%v", err)
			}
			p.OldID = cfg.ID
			cfg.ID = id
		}
		p.EmptyLocal = countFiles(filepath.Join(co.Main, filepath.FromSlash(LocalDir)))
	}

	// The lock.
	var lock *workspace.Lock
	if _, has, _ := readFile(co.Root, workspace.LockFile); has {
		if lock, err = workspace.LoadLock(co.Root); err != nil {
			return nil, fileError(err, "", "bad-lock", ExitState)
		}
	}
	// bonsai.yaml with no lock: the lock records which hook lines a person consented to, so update cannot tell a
	// changed hook line from one already agreed. Only init links again (spec §4, §6: it previews every settings line
	// first), judging every hook line, file a hook runs and plugin code part against what is on disk, as at a first
	// link (consent.go, rule 7); update refuses. (Part 5's verifier: deleting the lock let update --yes write a
	// changed hook line.)
	if lock == nil && req.Command == "update" {
		return nil, errorf("no-lock", ExitState, "restore the lock from git (git checkout -- "+workspace.LockFile+"); or, to link the project "+
			"again from bonsai.yaml, run bonsai init, which previews every file and settings line first",
			"bonsai.yaml is here but %s is not, so update cannot tell which hook lines were already consented to", workspace.LockFile)
	}
	p.FirstLink = lock == nil
	if lock == nil {
		lock = &workspace.Lock{Files: map[string]workspace.LockedFile{}, Format0: map[string]string{}}
	}
	inConfig := map[string]bool{}
	for _, r := range cfg.Packs {
		inConfig[r.ID] = true
	}
	lockedPack := map[string]workspace.LockedPack{}
	for _, lp := range lock.Packs {
		if !inConfig[lp.ID] {
			return nil, errorf("not-built", ExitState, "put the pack back in bonsai.yaml for now; taking a pack out of a project comes with step 5.1",
				"the lock holds the pack %s, which bonsai.yaml no longer lists", lp.ID).whose("person")
		}
		lockedPack[lp.ID] = lp
	}

	// Each pack at its ref, and at its locked commit.
	if newPacks == nil {
		for _, r := range cfg.Packs {
			was := c.lastResolved(r.Source, r.Ref)
			commit, err := c.fetch(r.Source, r.Ref)
			if err != nil {
				return nil, err
			}
			if lp, ok := lockedPack[r.ID]; ok && movedTag(r, lp, was, commit) {
				return nil, errorf("tag-moved", ExitState, "a person checks the tag (git ls-remote -- "+r.Source+" refs/tags/"+r.Ref+
					"); to take the new commit, set the pack's ref in bonsai.yaml to "+commit+" (or a new tag), then run bonsai update",
					"the pack %s's tag %s resolved to %s when the lock was written and now resolves to %s: a tag that moves is "+
						"refused (spec section 5), so nothing was written", r.ID, r.Ref, short(lp.Commit), short(commit)).whose("person")
			}
			c.recordResolved(r.Source, r.Ref, commit)
			pd, err := c.packAt(r, commit)
			if err != nil {
				return nil, err
			}
			newPacks = append(newPacks, pd)
		}
	}
	// Each locked pack at its locked commit: the baseline consent is judged against (consent.go). It is read at the
	// folder the lock records (formats set 4's path), and it is the baseline only when its content there hashes to the
	// lock's own sha256 for the pack (contentHash: every file of the pack's folder at the commit). A folder in
	// bonsai.yaml other than the lock's gives none, whatever the content (a content hash misses file modes and
	// submodules: step 5.1.1's verifier, B1's rest); so do a lock edited by hand and a commit that cannot be read: the
	// pack is unverified, and its code counts as at a first link (B1 and S1). A lock written before set 4 records no
	// folder: its pack is read at bonsai.yaml's folder and judged by the content hash alone.
	oldPacks := map[string]*PackData{}
	for _, pd := range newPacks {
		lp, ok := lockedPack[pd.Ref.ID]
		mv := PackMove{ID: pd.Ref.ID, Source: pd.Ref.Source, Version: pd.Manifest.Version, To: pd.Commit}
		if ok {
			mv.From = lp.Commit
			var od *PackData
			switch {
			case lp.PathSet && lp.Path != pd.Ref.Path:
				// The folder moved: no baseline.
			case lp.Commit == pd.Commit && lp.Source == pd.Ref.Source:
				od = pd
			default:
				if commit, err := c.fetch(lp.Source, lp.Commit); err == nil {
					old := pd.Ref
					old.Source = lp.Source
					od, _ = c.packAt(old, commit)
				}
			}
			if od != nil && od.SHA256 == lp.SHA256 {
				oldPacks[pd.Ref.ID] = od
			} else {
				p.Unverified = append(p.Unverified, pd.Ref.ID)
			}
		}
		p.Packs = append(p.Packs, mv)
	}

	p.pluginCode = map[string][]CodePart{}
	for _, pd := range newPacks {
		p.pluginCode[pd.Ref.ID] = pd.Code
	}

	// Two packs may not write one path (spec §6, "Layers").
	targets := map[string]target{}
	folded := map[string]string{}
	for _, pd := range newPacks {
		for _, fe := range pd.Manifest.Files {
			l := strings.ToLower(fe.Path)
			if other, ok := folded[l]; ok {
				return nil, errorf("packs-overlap", ExitInput, "take the path out of one of the two packs; overriding a file by path waits for 1.x",
					"two packs write %s (%s and %s)", fe.Path, other, pd.Ref.ID)
			}
			folded[l] = pd.Ref.ID
			targets[fe.Path] = target{pack: pd.Ref.ID, entry: fe, data: pd.Files[fe.Path]}
		}
	}
	// Two packs may not declare one document kind (contract §7.3: a kind's name is unique in the workspace).
	kinds := map[string]string{}
	for _, pd := range newPacks {
		for _, k := range pd.Declares.Documents {
			if other, ok := kinds[k.Kind]; ok {
				return nil, errorf("packs-overlap", ExitInput, "take one of the two packs out of bonsai.yaml; the packs' makers give the kinds names of their own",
					"two packs declare the document kind %s (%s and %s)", k.Kind, other, pd.Ref.ID)
			}
			kinds[k.Kind] = pd.Ref.ID
		}
	}
	// The files a hook runs (consent.go, rule 4): each runs item is a file a linked pack writes, and a pack's hook
	// command naming a file a linked pack writes lists it in its runs; else the plan is refused (S3).
	if err := checkRuns(newPacks, targets); err != nil {
		return nil, err
	}

	newLock := &workspace.Lock{WrittenBy: req.Version, Files: map[string]workspace.LockedFile{}, Format0: lock.Format0,
		Extra: lock.Extra}
	for _, pd := range newPacks {
		// The pack's folder, as bonsai.yaml names it (formats set 4's path; step 5.1.5): every lock written records it.
		declares, err := pd.Declares.Object()
		if err != nil {
			return nil, errorf("unexpected", ExitRuntime, "report this to Bonsai's maintainers", "the pack %s's declares cannot be written: %v", pd.Ref.ID, err)
		}
		lp := workspace.LockedPack{ID: pd.Ref.ID, Source: pd.Ref.Source, Version: pd.Manifest.Version, Commit: pd.Commit,
			SHA256: pd.SHA256, Declares: declares, Path: pd.Ref.Path, PathSet: true}
		if old, ok := lockedPack[pd.Ref.ID]; ok {
			lp.Extra = old.Extra
		}
		newLock.Packs = append(newLock.Packs, lp)
	}
	firstPack := ""
	if len(cfg.Packs) > 0 {
		firstPack = cfg.Packs[0].ID
	}

	// bonsai.yaml, when init writes it or gives it a new id.
	if p.NewConfig {
		p.Files = append(p.Files, &FileResult{Path: workspace.ConfigFile, Result: Created, Why: "this project's settings, from init's values",
			write: rawConfig})
	} else if p.OldID != "" {
		p.Files = append(p.Files, &FileResult{Path: workspace.ConfigFile, Result: Updated,
			Why: "a new id in place of " + p.OldID + " (each run draws its own)", write: rawConfig, old: origConfig})
	}

	// Pack files: every path a pack gives now, or the lock holds.
	paths := map[string]bool{}
	for path := range targets {
		paths[path] = true
	}
	for path, lf := range lock.Files {
		if lf.Kind == "pack" || lf.Kind == "once" || lf.Kind == "kept" {
			paths[path] = true
		}
	}
	sorted := make([]string, 0, len(paths))
	for path := range paths {
		sorted = append(sorted, path)
	}
	sort.Strings(sorted)
	for _, path := range sorted {
		disk, exists, err := readFile(co.Root, path)
		if err != nil {
			return nil, err
		}
		lf, locked := lock.Files[path]
		if locked && lf.Kind != "pack" && lf.Kind != "once" && lf.Kind != "kept" {
			locked = false
		}
		t, inPack := targets[path]
		f := &FileResult{Path: path, old: disk}
		var entry *workspace.LockedFile
		diskH := ""
		if exists {
			diskH = workspace.HashLF(disk)
		}
		switch {
		case inPack && t.entry.Kind == "once":
			f.Kind, f.Pack = "once", t.pack
			f.newH = workspace.HashLF(t.data)
			switch {
			case !locked && !exists:
				f.Result, f.Why, f.write = Created, "written once; from now on the project's", t.data
				entry = &workspace.LockedFile{Kind: "once", Pack: t.pack, SHA256: f.newH}
			case !locked:
				f.Result, f.Why = Found, "already in the project: left as it is, and the project's"
				entry = &workspace.LockedFile{Kind: "once", Pack: t.pack, SHA256: diskH}
			case !exists:
				f.Result, f.Why = LeftMissing, "the project's file, deleted: left missing"
				entry = &workspace.LockedFile{Kind: "once", Pack: t.pack, SHA256: lf.SHA256}
			default:
				f.Result, f.Why = Unchanged, "the project's file"
				entry = &workspace.LockedFile{Kind: "once", Pack: t.pack, SHA256: lf.SHA256}
			}
		case inPack:
			f.Kind, f.Pack = "pack", t.pack
			f.newH = workspace.HashLF(t.data)
			packEntry := &workspace.LockedFile{Kind: "pack", Pack: t.pack, SHA256: f.newH}
			switch {
			case !exists:
				f.Result, f.write, entry = Created, t.data, packEntry
				f.Why = "new from the pack"
				if locked {
					f.Result, f.Why = Restored, "missing: written again"
				}
			case !locked:
				if diskH == f.newH {
					f.Result, f.Why, entry = Adopted, "already the pack's copy: the lock catches up", packEntry
				} else {
					f.Result, f.Why = Conflict, "the project has its own file here, and the pack writes it too"
				}
			case lf.Kind == "kept":
				f.Kind = "kept"
				switch {
				case f.newH == lf.SHA256:
					f.Result, f.Why = Kept, "your kept edit; the pack's copy is unchanged"
					e := lf
					entry = &e
				case diskH == f.newH:
					f.Kind, f.Result, f.Why, entry = "pack", Adopted, "now equal to the pack's copy: the lock catches up", packEntry
				default:
					f.Result, f.Why = Conflict, "you kept your edit, and the pack changed the file again"
				}
			case diskH == lf.SHA256:
				if f.newH == lf.SHA256 {
					f.Result, f.Why, entry = Unchanged, "", packEntry
				} else {
					f.Result, f.Why, f.write, entry = Updated, "the pack changed it", t.data, packEntry
				}
			case f.newH == lf.SHA256:
				f.Result, f.Why = Changed, "you edited it; the pack did not change it, so it is left alone"
				e := lf
				e.Pack = t.pack
				entry = &e
			case diskH == f.newH:
				f.Result, f.Why, entry = Adopted, "your copy equals the pack's new one: the lock catches up", packEntry
			default:
				f.Result, f.Why = Conflict, "you edited it, and the pack changed it"
			}
			if f.Result == Conflict {
				e := lf
				entry = &e
				if !locked {
					entry = nil
				}
			}
		default: // the lock holds it; no pack gives it now
			f.Kind, f.Pack = lf.Kind, lf.Pack
			switch {
			case lf.Kind != "pack":
				f.Result, f.Why = Released, "the pack no longer has it: the project keeps it"
			case !exists:
				f.Result, f.Why = Dropped, "the pack no longer has it, and it is gone"
			case diskH == lf.SHA256:
				f.Result, f.Why, f.remove = Removed, "the pack no longer has it", true
			default:
				f.Result, f.Why = Conflict, "the pack no longer has it, and you edited it"
				e := lf
				entry = &e
			}
		}
		if entry != nil {
			if lf, ok := lock.Files[path]; ok {
				entry.Extra = lf.Extra
			}
			newLock.Files[path] = *entry
		}
		p.Files = append(p.Files, f)
	}

	// The block in CLAUDE.md.
	bd, err := readBlock(co.Root)
	if err != nil {
		return nil, err
	}
	body := blockBody(cfg.Name, newPacks)
	bf := &FileResult{Path: BlockFile, Kind: "block", Pack: firstPack, old: bd.raw, newH: regionHash(body)}
	if !bd.exists {
		bf.old = nil
	}
	lf, locked := lock.Files[BlockFile]
	blockEntry := workspace.LockedFile{Kind: "block", Pack: firstPack, SHA256: bf.newH, Extra: lf.Extra}
	switch {
	case !locked || lf.Kind != "block":
		switch {
		case bd.found && bd.region() == body:
			bf.Result, bf.Why = Adopted, "the block is already there: the lock catches up"
		case !bd.exists:
			bf.Result, bf.Why, bf.write = Created, "a new CLAUDE.md holding Bonsai's block", bd.withBlock(body)
		default:
			bf.Result, bf.Why, bf.write = Updated, "Bonsai's block added; the rest of the file is kept", bd.withBlock(body)
		}
		newLock.Files[BlockFile] = blockEntry
	case !bd.found:
		bf.Result, bf.Why, bf.write = Restored, "Bonsai's block is missing: written again", bd.withBlock(body)
		newLock.Files[BlockFile] = blockEntry
	default:
		diskH := regionHash(bd.region())
		switch {
		case diskH == lf.SHA256 && bf.newH == lf.SHA256:
			bf.Result = Unchanged
			newLock.Files[BlockFile] = blockEntry
		case diskH == lf.SHA256:
			bf.Result, bf.Why, bf.write = Updated, "the block changed", bd.withBlock(body)
			newLock.Files[BlockFile] = blockEntry
		case bf.newH == lf.SHA256:
			bf.Result, bf.Why = Changed, "you edited Bonsai's block; it is left alone"
			newLock.Files[BlockFile] = lf
		case diskH == bf.newH:
			bf.Result, bf.Why = Adopted, "the block equals the new one: the lock catches up"
			newLock.Files[BlockFile] = blockEntry
		default:
			bf.Result, bf.Why = Conflict, "you edited Bonsai's block, and the block changed"
			newLock.Files[BlockFile] = lf
		}
	}
	p.Files = append(p.Files, bf)

	// Bonsai's lines in .claude/settings.json.
	sd, err := readSettings(co.Root)
	if err != nil {
		return nil, err
	}
	var newPL, refPL []packLines
	for _, pd := range newPacks {
		newPL = append(newPL, packLines{id: pd.Ref.ID, source: pd.Ref.Source, folder: pd.Ref.Path, commit: pd.Commit,
			known: true, hooks: pd.Manifest.Hooks, deny: pd.Manifest.Deny})
		lp, ok := lockedPack[pd.Ref.ID]
		if !ok {
			continue
		}
		pl := packLines{id: lp.ID, source: lp.Source, folder: lockedFolder(lp, pd.Ref.Path), commit: lp.Commit}
		if od := oldPacks[lp.ID]; od != nil {
			pl.known, pl.hooks, pl.deny = true, od.Manifest.Hooks, od.Manifest.Deny
		}
		refPL = append(refPL, pl)
	}
	lnew := buildLines(cfg, newPL)
	slf, slocked := lock.Files[SettingsFile]
	if slocked && slf.Kind != "keys" {
		slocked = false
	}
	var lref []Line
	lockHash := ""
	if slocked {
		lref = buildLines(cfg, refPL)
		lockHash = slf.SHA256
	}
	disk := diskLines(sd.root)
	claimed, same := claim(disk, lref, p.FirstLink, lockHash)
	// The hook lines a person consented to: the lines Bonsai would have written from the lock's packs at their
	// locked commits, only when they hash to the lock's own record of the settings file, so they are what Bonsai
	// last wrote. Else none: a hook line that only comes back then counts as new code too (S1).
	var consented []Line
	if lockHash != "" && linesHash(lref) == lockHash {
		consented = lref
	}
	// A link again with the lock missing (rule 7), or a pack whose lock Bonsai cannot verify: a hook line on disk
	// that Bonsai neither claims nor writes may be a pack's earlier line, but Bonsai cannot tell it from the
	// project's own, so it stays in place; the preview names each, so a person is not surprised.
	if (p.FirstLink && hasConfig) || len(p.Unverified) > 0 {
		mine := map[string]bool{}
		for _, l := range append(append([]Line{}, claimed...), lnew...) {
			mine[l.canon()] = true
		}
		for _, d := range disk {
			if d.Kind == "hook" && !mine[d.canon()] {
				p.LeftHooks = append(p.LeftHooks, d.Text())
			}
		}
	}
	sf := &FileResult{Path: SettingsFile, Kind: "keys", Pack: firstPack, newH: linesHash(lnew)}
	if sd.exists {
		sf.old = sd.raw
	}
	changes := lineChanges(claimed, lnew, disk, consented)
	keysEntry := workspace.LockedFile{Kind: "keys", Pack: firstPack, SHA256: sf.newH, Extra: slf.Extra}
	write := false
	switch {
	case !slocked:
		write = true
		sf.Result, sf.Why = Updated, "Bonsai's lines added; the project's own entries are kept"
		if !sd.exists {
			sf.Result, sf.Why = Created, "a new file holding Bonsai's lines"
		}
		newLock.Files[SettingsFile] = keysEntry
	case !sd.exists:
		write = true
		sf.Result, sf.Why = Restored, "missing: Bonsai's lines written again"
		newLock.Files[SettingsFile] = keysEntry
	case same && sf.newH == slf.SHA256:
		sf.Result = Unchanged
		newLock.Files[SettingsFile] = keysEntry
	case same:
		write = true
		sf.Result, sf.Why = Updated, "Bonsai's lines changed; the project's own entries are kept"
		newLock.Files[SettingsFile] = keysEntry
	case sf.newH == slf.SHA256:
		sf.Result, sf.Why = Changed, "Bonsai's lines were edited by hand; they are left alone"
		newLock.Files[SettingsFile] = slf
	case linesHash(claimed) == sf.newH:
		sf.Result, sf.Why = Adopted, "Bonsai's lines already read as the new ones: the lock catches up"
		newLock.Files[SettingsFile] = keysEntry
	default:
		sf.Result, sf.Why = Conflict, "Bonsai's lines were edited by hand, and the new lines differ"
		newLock.Files[SettingsFile] = slf
	}
	if write && len(changes) == 0 && sd.exists {
		sf.Result, sf.Why = Adopted, "every line Bonsai writes is already there: the lock catches up"
		write = false
	}
	settingsBytes := func() ([]byte, error) {
		return schema.Encode(applyLines(sd.root, claimed, lnew))
	}
	if write {
		if sf.write, err = settingsBytes(); err != nil {
			return nil, errorf("unexpected", ExitRuntime, "run the command again", "%s cannot be written: %v", SettingsFile, err)
		}
		p.Settings = changes
	}
	p.Files = append(p.Files, sf)

	// .bonsai/.gitignore.
	gi, giExists, err := readFile(co.Root, GitignoreFile)
	if err != nil {
		return nil, err
	}
	gf := &FileResult{Path: GitignoreFile, old: gi, Result: Unchanged}
	switch {
	case !giExists:
		gf.Result, gf.Why, gf.write = Created, "keeps .bonsai/local/ out of git", []byte(GitignoreText)
		if !p.FirstLink {
			gf.Result, gf.Why = Restored, "missing: written again"
		}
	case workspace.HashLF(gi) != workspace.HashLF([]byte(GitignoreText)):
		gf.Result, gf.Why, gf.write = Restored, "changed: Bonsai's copy written again", []byte(GitignoreText)
	}
	p.Files = append(p.Files, gf)

	// --keep and --adopt.
	if err := p.resolve(req, newLock, claimed, lnew, consented, disk, bd, body, settingsBytes, targets); err != nil {
		return nil, err
	}
	for _, f := range p.Files {
		if f.Result == Conflict {
			p.Conflicts = append(p.Conflicts, f)
		}
	}
	// What runs code (consent.go): the settings lines the plan writes, or, while the settings file is a conflict,
	// the lines --adopt would write, so one refusal names both steps.
	consentLines := p.Settings
	if sf.write == nil && sf.Result == Conflict {
		consentLines = changes
	}
	p.consent(consentLines, newPacks, oldPacks)

	// The lock, written last, and only when it changes (written_by aside).
	p.lock = newLock
	if p.lockBytes, err = newLock.Encode(); err != nil {
		return nil, errorf("unexpected", ExitRuntime, "report this to Bonsai's maintainers", "the new lock cannot be written: %v", err)
	}
	oldRaw, hadLock, err := readFile(co.Root, workspace.LockFile)
	if err != nil {
		return nil, err
	}
	p.LockWrite = true
	if hadLock {
		same := *newLock
		same.WrittenBy = lock.WrittenBy
		if b, err := same.Encode(); err == nil && bytes.Equal(b, bytes.ReplaceAll(oldRaw, []byte("\r\n"), []byte("\n"))) {
			p.LockWrite = false
		}
	}
	return p, nil
}

// lockedFolder is the pack's folder as the lock records it, or, for a lock written before formats set 4 (no path),
// bonsai.yaml's: the folder its lines were last written with.
func lockedFolder(lp workspace.LockedPack, configured string) string {
	if lp.PathSet {
		return lp.Path
	}
	return configured
}

// target is a file a pack gives now.
type target struct {
	pack  string
	entry workspace.FileEntry
	data  []byte
}

// resolve applies --keep and --adopt to the plan's conflicts and edited files.
func (p *Plan) resolve(req Request, newLock *workspace.Lock, claimed, lnew, consented, disk []Line,
	bd *blockDoc, body string, settingsBytes func() ([]byte, error), targets map[string]target) error {
	byPath := map[string]*FileResult{}
	for _, f := range p.Files {
		byPath[f.Path] = f
	}
	conflicts := func() string {
		var names []string
		for _, f := range p.Files {
			if f.Result == Conflict || f.Result == Changed {
				names = append(names, f.Path)
			}
		}
		if len(names) == 0 {
			return "this update has none"
		}
		return strings.Join(names, ", ")
	}
	seen := map[string]bool{}
	for _, list := range []struct {
		flag  string
		paths []string
	}{{"--keep", req.Keep}, {"--adopt", req.Adopt}} {
		for _, path := range list.paths {
			path = filepath.ToSlash(path)
			f, ok := byPath[path]
			if !ok || (f.Result != Conflict && f.Result != Changed) {
				return errorf("bad-value", ExitInput, "name a file this update reports as a conflict or as edited: "+conflicts(),
					"%s %s: that file is no conflict and no edited file of this update", list.flag, path)
			}
			if seen[path] {
				return errorf("bad-value", ExitInput, "name each file once, with --keep or with --adopt", "%s is named twice", path)
			}
			seen[path] = true
			switch {
			case list.flag == "--keep" && (f.Kind == "block" || f.Kind == "keys"):
				return errorf("bad-value", ExitInput, "use --adopt "+path+" to take Bonsai's part (your copy is saved in the Bonsai home), or put Bonsai's part back by hand",
					"--keep %s: Bonsai's own part of that file cannot be kept edited", path)
			case list.flag == "--keep":
				lf := newLock.Files[path]
				if f.newH == "" { // the pack no longer has it: the project keeps it
					delete(newLock.Files, path)
					f.Result, f.Why = Released, "the pack no longer has it: you keep your copy"
				} else {
					newLock.Files[path] = workspace.LockedFile{Kind: "kept", Pack: f.Pack, SHA256: f.newH, Extra: lf.Extra}
					f.Kind, f.Result, f.Why = "kept", Kept, "you keep your edit (--keep); a later pack change to it is a conflict again"
				}
			default: // --adopt
				f.backup = true
				switch f.Kind {
				case "block":
					f.write = bd.withBlock(body)
					newLock.Files[path] = workspace.LockedFile{Kind: "block", Pack: f.Pack, SHA256: f.newH, Extra: newLock.Files[path].Extra}
				case "keys":
					b, err := settingsBytes()
					if err != nil {
						return errorf("unexpected", ExitRuntime, "run the command again", "%s cannot be written: %v", SettingsFile, err)
					}
					f.write = b
					p.Settings = lineChanges(claimed, lnew, disk, consented)
					newLock.Files[path] = workspace.LockedFile{Kind: "keys", Pack: f.Pack, SHA256: f.newH, Extra: newLock.Files[path].Extra}
				default:
					if t, ok := targets[path]; ok {
						f.write = t.data
						newLock.Files[path] = workspace.LockedFile{Kind: "pack", Pack: t.pack, SHA256: f.newH, Extra: newLock.Files[path].Extra}
						f.Kind = "pack"
					} else {
						f.remove = true
						delete(newLock.Files, path)
					}
				}
				f.Result, f.Why = Replaced, "the pack's copy taken (--adopt); yours is saved in the Bonsai home's cache"
				if f.remove {
					f.Result, f.Why = Removed, "taken out as the pack has (--adopt); yours is saved in the Bonsai home's cache"
				}
			}
		}
	}
	return nil
}

// sameValues reports whether init's values say what bonsai.yaml says.
func sameValues(cfg *workspace.Config, v *InitValues) bool {
	if v.Name != "" && v.Name != cfg.Name {
		return false
	}
	if v.Source != "" || v.Ref != "" || v.Path != "" {
		if len(cfg.Packs) != 1 {
			return false
		}
		r := cfg.Packs[0]
		if (v.Source != "" && v.Source != r.Source) || (v.Ref != "" && v.Ref != r.Ref) || v.Path != r.Path {
			return false
		}
	}
	if len(v.NeverEdit) > 0 && strings.Join(v.NeverEdit, "\n") != strings.Join(cfg.NeverEdit, "\n") {
		return false
	}
	return true
}

func missingValues(v *InitValues) []string {
	var missing []string
	if v == nil {
		v = &InitValues{}
	}
	if v.Name == "" {
		missing = append(missing, "--name")
	}
	if v.Source == "" {
		missing = append(missing, "--source")
	}
	if v.Ref == "" {
		missing = append(missing, "--ref")
	}
	return missing
}

// checkValues holds init's values to what bonsai.yaml may say, before anything is fetched.
func checkValues(v *InitValues) error {
	next := "bonsai init --help shows each value's form"
	for _, s := range append([]string{v.Name, v.Source, v.Path, v.Ref}, v.NeverEdit...) {
		for _, r := range s {
			if r < 0x20 || r == 0x7f {
				return errorf("bad-value", ExitInput, next, "init's values may not hold a control character")
			}
		}
	}
	if strings.HasPrefix(v.Source, "-") || strings.HasPrefix(v.Ref, "-") {
		return errorf("bad-value", ExitInput, next, "--source and --ref may not start with -, which git would read as an option")
	}
	// The rest (the name's form, the path, each never_edit path) bonsai.yaml's own reader checks, on the bonsai.yaml
	// init would write, before it is written.
	probe := ConfigYAML("ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", *v, "probe")
	if _, err := workspace.ReadConfigFull(probe); err != nil {
		return fileError(err, "", "bad-value", ExitInput).(*Error).withNext(next)
	}
	return nil
}

func (e *Error) withNext(next string) *Error {
	e.What = strings.Replace(e.What, "bonsai.yaml", "init's values (bonsai.yaml)", 1)
	e.Next = next
	return e
}

func countFiles(dir string) int {
	n := 0
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			n++
		}
		return nil
	})
	if _, err := os.Stat(dir); err != nil {
		return -1
	}
	return n
}

// andList joins words as a person writes a list: "a", "a and b", "a, b and c".
func andList(words []string) string {
	if len(words) < 2 {
		return strings.Join(words, "")
	}
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}
