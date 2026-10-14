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
//   - two files of one pack that resolve to one path are refused (bad-config: the folder makes them one); two packs
//     writing one path are refused as before (packs-overlap, plan.go);
//   - the lock records the resolved path (plan.go keys each file by it), and the block imports it as it imports every
//     pack file under documents.protocols (block.go);
//   - a project that changes its folder gets, at its next update, the old file handled as one the pack no longer has
//     (removed when unedited, a conflict when edited: plan.go's table) and the new one written: nothing new for the move.

import (
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// resolveFiles puts the project's documents.protocols folder in place of each <protocols>/ start of the packs' files
// (their manifest's files entries and the bytes read for each), in place. Its error is bad-config, naming the line of
// bonsai.yaml to set.
func resolveFiles(cfg *workspace.Config, packs []*PackData) error {
	protocols, line := "", 0
	if cfg.Full != nil {
		protocols = cfg.Full.Documents.Protocols
	}
	for _, d := range cfg.Documents {
		if d.Kind == "protocols" {
			line = d.Line
		}
	}
	where := "documents.protocols in bonsai.yaml"
	if line > 0 {
		where += " (line " + strconv.Itoa(line) + ")"
	}
	next := "set " + where + " to a folder of the project's own for the packs' always-on protocol files, such as " +
		DefaultDocuments.Protocols + " (not inside .claude, .git or .bonsai), then run the command again"
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
