package engine

// The all-or-nothing write (spec §6: "every write is staged first; renames follow; the lock is written last, so a
// crash between renames is finished by running update again").
//
// Order: the project's copies --adopt replaces are saved in the home's cache first, so a person's edit is safe
// before anything in the project moves; then every new file is staged beside its target; then the renames, bonsai.yaml
// first (spec §14, check 1: "it writes bonsai.yaml first"); then the files a pack no longer has are removed, in the
// plan's order (unlink's bonsai.yaml after every other file), and the folders that leaves empty; init --new-id empties
// .bonsai/local/; the lock last, written (init, update) or removed (unlink). A run that stops between two renames
// leaves the lock as it was, so the next update finds the renamed files equal to the pack's copies (adopted) and
// finishes the rest; the next unlink finds the removed files gone and finishes the rest from the lock.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// AdoptedDir is where --adopt saves a project's copy, under the home: cache/adopted/<workspace id>/<first 12 hex
// of the copy's fingerprint>/<path in the project>. Saving the same copy twice writes the same place.
func AdoptedDir(home, id, fingerprint string) string {
	return filepath.Join(home, "cache", "adopted", id, fingerprint[:12])
}

// Apply writes a plan. It refuses a plan with conflicts, or one that runs code without the request's --allow-exec
// (consent.go): the caller checks both first. Refused, or stopped before the renames (a copy that cannot be saved in
// the home, a file that cannot be staged), it writes nothing in the project. Stopped once the renames have begun, its
// error's word is partly-written (the changes output's result failed): some files are written and the lock is not,
// so the same command again finishes the rest.
func Apply(p *Plan) error {
	if len(p.Conflicts) > 0 {
		return errorf("conflicts", ExitConflict, "run the command without --yes to see the conflicts and the commands that settle them",
			"this plan has conflicts and cannot be applied: nothing was written")
	}
	if p.NeedsExec() && !p.AllowExec {
		return errorf("needs-allow-exec", ExitState, "run the command without --yes to see what runs code; to write it, add --allow-exec as well as --yes",
			"this plan writes code that runs on this machine, and --allow-exec was not given: nothing was written")
	}
	// The project's copies first.
	for _, f := range p.Files {
		if !f.backup || f.old == nil {
			continue
		}
		dir := AdoptedDir(p.Home, p.Config.ID, workspace.HashLF(f.old))
		dest := filepath.Join(dir, filepath.FromSlash(f.Path))
		if err := workspace.WriteFileAtomic(dest, f.old); err != nil {
			return errorf("bad-home", ExitRuntime, "check that the Bonsai home can be written (BONSAI_HOME), then run the command again; nothing in the project was written",
				"your copy of %s cannot be saved in the Bonsai home: %v", f.Path, err)
		}
		f.Saved = filepath.ToSlash(dest)
	}

	// Stage every write.
	type staged struct{ tmp, path string }
	var writes []*FileResult
	for _, f := range p.Files {
		if f.write != nil {
			writes = append(writes, f)
		}
	}
	sort.SliceStable(writes, func(i, j int) bool {
		return writes[i].Path == workspace.ConfigFile && writes[j].Path != workspace.ConfigFile
	})
	var stage []staged
	cleanup := func() {
		for _, s := range stage {
			_ = os.Remove(s.tmp)
		}
	}
	for _, f := range writes {
		target := filepath.Join(p.Root, filepath.FromSlash(f.Path))
		tmp, err := workspace.StageFile(target, f.write)
		if err != nil {
			cleanup()
			return errorf("write-failed", ExitRuntime, "check the project's folders can be written, then run the command again; nothing was written",
				"%s cannot be staged: %v", f.Path, err)
		}
		stage = append(stage, staged{tmp, target})
	}

	// The renames, then the removals.
	partly := "run the same command again: it finishes the rest (the lock is written last)"
	for i, s := range stage {
		if err := workspace.CommitStaged(s.tmp, s.path); err != nil {
			for _, rest := range stage[i:] {
				_ = os.Remove(rest.tmp)
			}
			return errorf("partly-written", ExitRuntime, partly, "%s cannot be written: %v", filepath.ToSlash(s.path), err)
		}
	}
	var emptied []string
	for _, f := range p.Files {
		if f.remove {
			if err := workspace.RemoveFile(filepath.Join(p.Root, filepath.FromSlash(f.Path))); err != nil {
				return errorf("partly-written", ExitRuntime, partly, "%s cannot be removed: %v", f.Path, err)
			}
			emptied = append(emptied, f.Path)
		}
	}
	if p.EmptyLocal > 0 {
		local := filepath.Join(p.Main, filepath.FromSlash(LocalDir))
		entries, err := os.ReadDir(local)
		if err != nil {
			return errorf("partly-written", ExitRuntime, partly, "%s cannot be read: %v", LocalDir, err)
		}
		for _, e := range entries {
			if err := os.RemoveAll(filepath.Join(local, e.Name())); err != nil {
				return errorf("partly-written", ExitRuntime, partly, "%s cannot be emptied: %v", LocalDir, err)
			}
		}
	}

	// The lock, last.
	if p.LockWrite {
		if err := workspace.WriteFileAtomic(filepath.Join(p.Root, filepath.FromSlash(workspace.LockFile)), p.lockBytes); err != nil {
			return errorf("partly-written", ExitRuntime, partly, "%s cannot be written: %v", workspace.LockFile, err)
		}
	}
	if p.LockRemove {
		if err := workspace.RemoveFile(filepath.Join(p.Root, filepath.FromSlash(workspace.LockFile))); err != nil {
			return errorf("partly-written", ExitRuntime, "run the same command again: it finishes the rest (the lock is removed last)",
				"%s cannot be removed: %v", workspace.LockFile, err)
		}
		emptied = append(emptied, workspace.LockFile)
	}
	pruneEmpty(p.Root, emptied)
	return nil
}

// pruneEmpty removes each folder the removals left empty, from a removed file's folder up to the checkout's top
// (which it never removes). A folder that still holds anything stays; a folder that cannot be removed is left, since
// the project's files are already as the plan says (git does not show an empty folder).
func pruneEmpty(root string, removed []string) {
	root = filepath.Clean(root)
	for _, rel := range removed {
		dir := filepath.Dir(filepath.Join(root, filepath.FromSlash(rel)))
		for dir != root && strings.HasPrefix(dir, root+string(filepath.Separator)) {
			if os.Remove(dir) != nil {
				break
			}
			dir = filepath.Dir(dir)
		}
	}
}
