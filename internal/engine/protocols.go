package engine

// A pack file's <protocols>/ start (set 8; plan-5 5.5.0 note 1; spec §5: session-start is "a pack file the engine
// writes to the workspace's protocols folder, imported by the block"). A pack cannot know a project's folder, so a
// files entry's path may start with <protocols>/ (workspace.PathPlaceholders, their one home): at each init and update
// the engine puts bonsai.yaml's documents.protocols in its place (workspace.ResolveFilePath), before anything else
// reads the packs' files, so every rule after it sees the path the file is written at:
//   - the resolved path passes every check a pack file's path passes (CheckRelPath, checkPackTarget), so a
//     documents.protocols of .claude, .git, .bonsai, .. or a short name such as CLAUDE~1 is refused, at a first link and
//     at a folder change alike; so is a project with no documents.protocols (bonsai.yaml's own reader refuses a null
//     or empty one, naming its line; the line left out is refused here). Each is bad-config, a person's, naming the
//     line of bonsai.yaml to set;
//   - a documents.protocols that names a file of the project, or a path through one (CLAUDE.md, bonsai.yaml), is
//     refused before anything is written (protocolsFolder), at a first link and at update alike: bad-config, as above;
//     so is one whose first segment is a file Bonsai writes at the project's top (topTargets), in any letter case,
//     even where that file is not there yet;
//   - two files of one pack that resolve to one path are refused (bad-config: the folder makes them one); two packs
//     writing one path are refused as before (packs-overlap, plan.go);
//   - the lock records the resolved path (plan.go keys each file by it), and the block imports it as it imports every
//     pack file under documents.protocols (block.go);
//   - a project that changes its folder gets, at its next update, the old file handled as one the pack no longer has
//     (removed when unedited, a conflict when edited: plan.go's table) and the new one written: nothing new for the move.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// resolveFiles puts the project's documents.protocols folder in place of each <protocols>/ start of the packs' files
// (their manifest's files entries and the bytes read for each), in place. Its error is bad-config, naming the line of
// bonsai.yaml to set.
func resolveFiles(cfg *workspace.Config, packs []*PackData) error {
	protocols, _, next := protocolsWhere(cfg)
	for _, pd := range packs {
		files := make(map[string][]byte, len(pd.Files))
		seen := map[string]string{} // a resolved path, letter case aside, to the pack's own path for it
		for i, fe := range pd.Manifest.Files {
			resolved, err := workspace.ResolveFilePath(fe.Path, protocols)
			if err != nil {
				return errorf("bad-config", ExitInput, next, "the pack %s: %v", pd.Ref.ID, err)
			}
			folded := strings.ToLower(resolved)
			if other, ok := seen[folded]; ok {
				return errorf("bad-config", ExitInput, next, "the pack %s writes one path twice under documents.protocols %s: %s and %s "+
					"(letter case aside, as Windows sees it)", pd.Ref.ID, quote(protocols), quote(other), quote(fe.Path))
			}
			seen[folded] = fe.Path
			files[resolved] = pd.Files[fe.Path]
			pd.Manifest.Files[i].Path = resolved
		}
		pd.Files = files
	}
	return nil
}

// protocolsWhere gives the project's documents.protocols, the words naming its line in bonsai.yaml, and the next step
// of a refusal of it (a person's: bonsai.yaml is theirs).
func protocolsWhere(cfg *workspace.Config) (protocols, where, next string) {
	line := 0
	if cfg.Full != nil {
		protocols = cfg.Full.Documents.Protocols
	}
	for _, d := range cfg.Documents {
		if d.Kind == "protocols" {
			line = d.Line
		}
	}
	where = "documents.protocols in bonsai.yaml"
	if line > 0 {
		where += " (line " + strconv.Itoa(line) + ")"
	}
	next = "set " + where + " to a folder of the project's own for the packs' always-on protocol files, such as " +
		DefaultDocuments.Protocols + " (not inside .claude, .git or .bonsai), then run the command again"
	return protocols, where, next
}

// topTargets are the files Bonsai writes at the project's top, which no documents.protocols may name or run through:
// the instruction block's CLAUDE.md and bonsai.yaml (init writes it). The folders it writes there, .claude/ and
// .bonsai/, are refused by resolveFiles (checkPackTarget), as for any pack file's path.
var topTargets = []string{BlockFile, workspace.ConfigFile}

// protocolsFolder refuses a documents.protocols that names a file of the project, or a path through one, before
// anything is written, at a first link and at update alike (V1's should-fix 3, 16 Oct: CLAUDE.md there left a link
// written part-way, and bonsai.yaml failed as a file that cannot be read): bad-config, naming the line to set. Each
// segment of the folder is looked at in the checkout, as it stands; one that is there and is not a folder (a file, a
// link) is refused. Then a folder whose first segment is one of topTargets, letter case aside (as Windows sees it), is
// refused whether it is there or not (V1's second round, 16 Oct: CLAUDE.md at a first link, before the block is
// written, made CLAUDE.md a folder and left the link part-way). Any other folder that is not there yet, or one of the
// project's own, passes; a value the packs' paths cannot take (.., an absolute path, .claude) is left to resolveFiles,
// which names why.
func protocolsFolder(root string, cfg *workspace.Config) error {
	protocols, where, next := protocolsWhere(cfg)
	if protocols == "" || workspace.CheckRelPath(protocols) != nil {
		return nil
	}
	parts := strings.Split(protocols, "/")
	for i := range parts {
		at := strings.Join(parts[:i+1], "/")
		fi, err := os.Lstat(filepath.Join(root, filepath.FromSlash(at)))
		if err != nil {
			break // not there yet: what is below it is made as a folder
		}
		if fi.IsDir() {
			continue
		}
		what := "a file of the project"
		if fi.Mode()&os.ModeSymlink != 0 {
			what = "a link in the project"
		}
		if at == protocols {
			return errorf("bad-config", ExitInput, next, "%s names %s, %s, not a folder for the packs' always-on protocol "+
				"files: nothing was written", where, quote(protocols), what)
		}
		return errorf("bad-config", ExitInput, next, "%s names %s, inside %s, %s, not a folder for the packs' always-on "+
			"protocol files: nothing was written", where, quote(protocols), quote(at), what)
	}
	for _, top := range topTargets {
		if !strings.EqualFold(parts[0], top) {
			continue
		}
		if len(parts) == 1 {
			return errorf("bad-config", ExitInput, next, "%s names %s, which is %s, a file Bonsai writes at the project's "+
				"top (letter case aside, as Windows sees it), not a folder for the packs' always-on protocol files: nothing "+
				"was written", where, quote(protocols), quote(top))
		}
		return errorf("bad-config", ExitInput, next, "%s names %s, inside %s, a file Bonsai writes at the project's top "+
			"(letter case aside, as Windows sees it), not a folder for the packs' always-on protocol files: nothing was "+
			"written", where, quote(protocols), quote(top))
	}
	return nil
}
