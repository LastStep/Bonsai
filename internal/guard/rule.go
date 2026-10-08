package guard

// The walking skeleton's one rule (plan part 5): an edit of a path on bonsai.yaml's protected list is refused. The
// person_only list is read too: it is a part of the protected list (spec §6), so a path on it alone is protected as
// well. Nothing grants a protected path in this build: tasks and their grants come with step 5.3.

import (
	"fmt"
	"strconv"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// Decide judges one call against bonsai.yaml's lists. root is the project's absolute folder, the one holding
// bonsai.yaml.
func Decide(in *Input, root string, cfg *workspace.Config) Decision {
	if _, ok := fileTools[in.Tool]; !ok {
		if in.Tool == "Bash" || in.Tool == "PowerShell" {
			return allow(RuleShell, "")
		}
		return allow(RuleOtherTool, "")
	}
	lists := []struct {
		key, rule string
		globs     []string
	}{
		{"person_only", RulePersonOnly, cfg.PersonOnly},
		{"protected", RuleProtected, cfg.Protected},
	}
	for _, l := range lists {
		for _, g := range l.globs {
			if err := checkGlob(g); err != nil {
				return deny(RuleBadGlob,
					fmt.Sprintf("%s refused: the %s glob %s in bonsai.yaml is not a glob the guard reads (%s), so it "+
						"cannot tell whether the path is protected", in.Tool, l.key, strconv.QuoteToASCII(g), oneLine(err.Error())),
					"leave the file as it is and tell the person: a person fixes that glob in bonsai.yaml")
			}
		}
	}
	rels := projectPaths(root, in.Cwd, in.Path)
	if len(rels) == 0 {
		return allow(RuleOutside, "")
	}
	fold := caseInsensitive(root)
	for _, l := range lists {
		for _, g := range l.globs {
			for _, rel := range rels {
				if !matchGlob(g, rel, fold) {
					continue
				}
				d := deny(l.rule, "", "")
				d.Target = rel
				shown := quote(rel, 120)
				if l.rule == RulePersonOnly {
					d.Why = fmt.Sprintf("%s of %s refused: it is on bonsai.yaml's person_only list (%s), so only a "+
						"person changes it", in.Tool, shown, strconv.QuoteToASCII(g))
					d.Next = "leave the file as it is and ask the person to make this change"
				} else {
					d.Why = fmt.Sprintf("%s of %s refused: it is on bonsai.yaml's protected list (%s), and this build "+
						"gives no task the right to change it", in.Tool, shown, strconv.QuoteToASCII(g))
					d.Next = "leave the file as it is and tell the person; a person changes it (a task's grants come " +
						"with Bonsai's step 5.3)"
				}
				return d
			}
		}
	}
	return allow(RuleNotProtected, rels[0])
}
