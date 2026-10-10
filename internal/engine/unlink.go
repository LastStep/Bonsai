package engine

// bonsai unlink (spec section 4; plan-5, step 5.1.7): take Bonsai out of a checkout, from what the lock says Bonsai
// wrote. Like init and update it works out a plan first (BuildUnlink), which the preview prints and Apply writes.
//
// What it removes: the pack files nobody edited (their bytes still hash to the lock's), Bonsai's block in CLAUDE.md
// (the file too when the block was all it held), Bonsai's lines in .claude/settings.json (the file too when they were
// all it held), the two tables Bonsai rebuilds (.bonsai/tasks.md and .bonsai/sessions.md), .bonsai/.gitignore while
// .bonsai/local/ is empty, bonsai.yaml, and the lock, last. bonsai.yaml goes even when a person edited it: with no
// bonsai.yaml, Bonsai's hook lets every call through (spec section 7), a session still running included, so the
// preview says so. Folders the removals leave empty go too.
//
// What it leaves, and names: a pack file edited here, and the files written once (kind once) or kept (kind kept),
// now the project's (result released); Bonsai's block when a person edited it (between its markers, for them to take
// out); .bonsai/STATE.md; .bonsai/local/ (this checkout's log, asks and ladder results), with .bonsai/.gitignore while
// local/ holds anything, so git does not show its files; and this machine's folder for the checkout in the Bonsai home.
//
// Then, after the write, cmd/bonsai asks Claude Code to remove its record of each pack's plugin install for this
// checkout (UninstallPlugins, at project scope, after Bonsai's lines left .claude/settings.json: measured on Claude
// Code 2.1.295, plugins.go). A worktree's record is its own: unlink in each checkout. A plain git revert of the link
// commit leaves Claude Code's record, which unlink's help says.
//
// The order makes a run that stops part-way finish when run again: every file first, bonsai.yaml next, the lock last.
// With the lock there, unlink works from it whether or not bonsai.yaml is (a run that stopped after removing it reads
// Bonsai's lines from the lock alone); with neither, there is nothing to change (a second unlink); with bonsai.yaml and
// no lock, it refuses (no-lock): it cannot tell Bonsai's files from the project's.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The words unlink's plan gives the lock (bonsai.changes/1's lock field, an open list): removed.
const LockRemoved = "removed"

// BuildUnlink works out what unlink would do in the checkout holding dir. It writes nothing, and fetches a pack only
// for a lock written before formats set 4 (whose declares do not hold the pack's settings lines). Its error is an
// *Error.
func BuildUnlink(dir, home string) (_ *Plan, err error) {
	co, err := workspace.Find(dir)
	if err != nil {
		return nil, findError(err)
	}
	p := &Plan{Command: "unlink", Root: co.Root, Main: co.Main, Home: home, EmptyLocal: -1}
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
	rawConfig, hasConfig, err := readFile(co.Root, workspace.ConfigFile)
	if err != nil {
		return nil, err
	}
	cfg := &workspace.Config{}
	if hasConfig {
		if cfg, err = workspace.ReadConfigFull(rawConfig); err != nil {
			return nil, fileError(err, "", "bad-config", ExitInput).(*Error).withUnlinkNext()
		}
		p.Config = cfg
	}
	var lock *workspace.Lock
	if _, has, _ := readFile(co.Root, workspace.LockFile); has {
		if lock, err = workspace.LoadLock(co.Root); err != nil {
			e := fileError(err, "", "bad-lock", ExitState).(*Error)
			e.Next = "the lock is Bonsai's to write: restore the committed one, run: git checkout -- " + workspace.LockFile + "; then run: bonsai unlink"
			return nil, e
		}
	}
	switch {
	case lock == nil && !hasConfig:
		return p, nil // not linked: nothing to change
	case lock == nil:
		return nil, errorf("no-lock", ExitState, "restore the lock from git, run: git checkout -- "+workspace.LockFile+"; then run: bonsai unlink",
			"bonsai.yaml is here but %s is not, so unlink cannot tell the files Bonsai wrote from the project's", workspace.LockFile)
	}
	c := cache{home: home}
	p.LockRemove = true
	p.Removed = lock.Packs
	if cfg.Name != "" && len(lock.Packs) > 0 {
		p.OldMarket = lockMarket(cfg, lock)
	}
	for _, lp := range lock.Packs {
		p.Packs = append(p.Packs, PackMove{ID: lp.ID, Source: lp.Source, Version: lp.Version, From: lp.Commit})
	}

	// The files the lock lists of the packs' kinds.
	for _, path := range sortedPaths(lock.Files) {
		lf := lock.Files[path]
		if lf.Kind != "pack" && lf.Kind != "once" && lf.Kind != "kept" {
			continue
		}
		disk, exists, err := readFile(co.Root, path)
		if err != nil {
			return nil, err
		}
		f := &FileResult{Path: path, Kind: lf.Kind, Pack: lf.Pack, old: disk}
		switch {
		case !exists:
			f.Result, f.Why = Dropped, "already gone"
		case lf.Kind != "pack":
			f.Result, f.Why = Released, "the project's file: it stays"
		case workspace.HashLF(disk) == lf.SHA256:
			f.Result, f.Why, f.remove = Removed, "a pack file nobody edited", true
		default:
			f.Result, f.Why = Released, "you edited it: it stays, and is the project's now"
		}
		p.Files = append(p.Files, f)
	}

	// Bonsai's block in CLAUDE.md.
	if lf, ok := lock.Files[BlockFile]; ok && lf.Kind == "block" {
		bd, err := readBlock(co.Root)
		if err != nil {
			return nil, againUnlink(err)
		}
		f := &FileResult{Path: BlockFile, Kind: "block", Pack: lf.Pack, old: bd.raw}
		switch {
		case !bd.found:
			f.Result, f.Why = Dropped, "Bonsai's block is already gone"
		case regionHash(bd.region()) != lf.SHA256:
			f.Result, f.Why = Released, "you edited Bonsai's block, so it stays: take it out by hand, from its bonsai:block start line to its end line"
		default:
			rest := bd.withoutBlock()
			if len(bytes.TrimSpace(rest)) == 0 {
				f.Result, f.Why, f.remove = Removed, "it held only Bonsai's block", true
			} else {
				f.Result, f.Why, f.write = Updated, "Bonsai's block taken out; the rest of the file is kept", rest
			}
		}
		p.Files = append(p.Files, f)
	}

	// Bonsai's lines in .claude/settings.json: the lines it last wrote, from the lock's packs (their declares) and
	// bonsai.yaml, found in the file as update finds them (claim), and taken out.
	if lf, ok := lock.Files[SettingsFile]; ok && lf.Kind == "keys" {
		sd, err := readSettings(co.Root)
		if err != nil {
			return nil, againUnlink(err)
		}
		f := &FileResult{Path: SettingsFile, Kind: "keys", Pack: lf.Pack}
		switch {
		case !sd.exists:
			f.Result, f.Why = Dropped, "already gone"
		default:
			f.old = sd.raw
			var refPL []packLines
			for _, lp := range lock.Packs {
				pl, err := takenOutLines(c, lp)
				if err != nil {
					return nil, err
				}
				refPL = append(refPL, pl)
			}
			disk := diskLines(sd.root)
			claimed, same := claim(disk, buildLines(cfg, refPL), false, lf.SHA256)
			doc := applyLines(sd.root, claimed, nil)
			p.Settings = lineChanges(claimed, nil, disk, nil)
			switch {
			case len(claimed) == 0:
				f.Result, f.Why = Unchanged, "no line of Bonsai's is left in it"
			case len(doc) == 0:
				f.Result, f.Why, f.remove = Removed, "it held only Bonsai's lines", true
			default:
				b, err := schema.EncodeUTF8(doc)
				if err != nil {
					return nil, errorf("unexpected", ExitRuntime, "run the command again", "%s cannot be written: %v", SettingsFile, err)
				}
				f.Result, f.Why, f.write = Updated, "Bonsai's lines taken out; the project's own entries are kept, in their order", b
			}
			if !same && len(claimed) > 0 {
				f.Why += " (some of Bonsai's lines were edited by hand: the ones Bonsai still knows as its own go; read the file for the rest)"
			}
		}
		p.Files = append(p.Files, f)
	}

	// The tables Bonsai rebuilds, and .bonsai/.gitignore.
	for _, path := range []string{workspace.TasksTableFile, workspace.SessionsTableFile} {
		if raw, exists, err := readFile(co.Root, path); err != nil {
			return nil, err
		} else if exists {
			p.Files = append(p.Files, &FileResult{Path: path, Result: Removed, Why: "a table Bonsai rebuilds", old: raw, remove: true})
		}
	}
	if raw, exists, err := readFile(co.Root, workspace.GitignoreFile); err != nil {
		return nil, err
	} else if exists {
		f := &FileResult{Path: workspace.GitignoreFile, Result: Removed, Why: "Bonsai's; .bonsai/local/ holds nothing", old: raw, remove: true}
		if countFiles(filepath.Join(co.Root, filepath.FromSlash(LocalDir))) > 0 {
			f.Result, f.Why, f.remove = Released, "it stays while .bonsai/local/ holds files, so git does not show them; delete both "+
				"when you no longer need them", false
		}
		p.Files = append(p.Files, f)
	}

	// bonsai.yaml, after every other file and before the lock.
	if hasConfig {
		p.Files = append(p.Files, &FileResult{Path: workspace.ConfigFile, Result: Removed, old: rawConfig, remove: true,
			Why: "removed even if a person edited it: with no bonsai.yaml, Bonsai's hook lets every call through here, a session " +
				"still running included (spec section 7)"})
	}
	p.leftInPlace()
	return p, nil
}

// againUnlink gives a refusal whose next step ends "run the command again" unlink's exact command.
func againUnlink(err error) error {
	if e, ok := err.(*Error); ok {
		e.Next = strings.Replace(e.Next, "run the command again", "run: bonsai unlink", 1)
	}
	return err
}

// withUnlinkNext gives a bonsai.yaml refusal unlink's next step.
func (e *Error) withUnlinkNext() *Error {
	e.Next = "fix bonsai.yaml (" + strings.TrimSuffix(e.Next, ".") + "), or restore the committed one, run: git checkout -- " +
		workspace.ConfigFile + "; then run: bonsai unlink"
	return e
}

// leftInPlace lists what unlink leaves that is not one of the plan's files: .bonsai/STATE.md, .bonsai/local/, and this
// machine's folder for the checkout in the Bonsai home, each when it is there.
func (p *Plan) leftInPlace() {
	if _, err := os.Stat(filepath.Join(p.Root, filepath.FromSlash(workspace.StateFile))); err == nil && !p.lists(workspace.StateFile) {
		p.Left = append(p.Left, workspace.StateFile+": the project's STATE")
	}
	if n := countFiles(filepath.Join(p.Root, filepath.FromSlash(LocalDir))); n >= 0 {
		p.Left = append(p.Left, LocalDir+"/: this checkout's log, asks and ladder results, never committed")
	}
	if dir, err := workspace.MachineDir(p.Home, p.Main); err == nil {
		if _, err := os.Stat(dir); err == nil {
			p.Left = append(p.Left, filepath.ToSlash(dir)+"/: this machine's record and settings for the checkout, in the Bonsai home")
		}
	}
}

// lists reports whether the plan names a file.
func (p *Plan) lists(path string) bool {
	for _, f := range p.Files {
		if f.Path == path {
			return true
		}
	}
	return false
}

// withoutBlock gives CLAUDE.md without Bonsai's block, as withBlock found or put it: the block's lines go, and so does
// the blank line withBlock put before a block it added at the end of the file.
func (d *blockDoc) withoutBlock() []byte {
	if !d.found {
		return d.raw
	}
	before, after := d.raw[:d.start], d.raw[d.end:]
	if len(after) == 0 {
		switch {
		case bytes.HasSuffix(before, []byte("\r\n\r\n")):
			before = before[:len(before)-2]
		case bytes.HasSuffix(before, []byte("\n\n")):
			before = before[:len(before)-1]
		}
	}
	out := append([]byte{}, before...)
	return append(out, after...)
}

// LeftText is what unlink leaves that is not among its files, for a person: "" when there is nothing.
func (p *Plan) LeftText() string {
	if len(p.Left) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Left in place:\n")
	for _, l := range p.Left {
		b.WriteString("  " + ascii(l) + "\n")
	}
	return b.String()
}
