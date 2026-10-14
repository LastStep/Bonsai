package reference

// Every file go generate writes from Bonsai's code (Generated, their one list), and base's generated-files skill
// among them (plan-5 5.2.6a; 5.5.0 note 4: the page moved into the skill). gen writes each at its path from the
// repository's root; TestGeneratedAreCurrent rebuilds each and fails on any difference (Stale), naming go generate as
// the fix, so no generated file drifts from the code that builds it.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
)

// GeneratedFilesSkillPath is where go generate writes base's generated-files skill, from the repository's root.
const GeneratedFilesSkillPath = "packs/base/skills/generated-files/SKILL.md"

// GeneratedFile is one file go generate writes: its path from the repository's root, forward slashes, and the
// function that builds its bytes from the code.
type GeneratedFile struct {
	Path  string
	Build func() []byte
}

// Generated are every file go generate writes, in the order gen writes them: the reference page of lists, and base's
// three skills built from the code (the workspace and state skills from the templates the engine writes, the
// generated-files skill from format.GeneratedKinds).
func Generated() []GeneratedFile {
	return []GeneratedFile{
		{Path, Page},
		{engine.WorkspaceSkillPath, engine.WorkspaceSkill},
		{engine.StateSkillPath, engine.StateSkill},
		{GeneratedFilesSkillPath, GeneratedFilesSkill},
	}
}

// Stale lists the files of Generated that differ, under the repository root, from what the code builds now (a
// CRLF checkout's line ends read as LF), each with its first differing line, or that cannot be read.
func Stale(root string) []string {
	var out []string
	for _, g := range Generated() {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(g.Path)))
		if err != nil {
			out = append(out, fmt.Sprintf("%s cannot be read: %v", g.Path, err))
			continue
		}
		have, want := bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n")), g.Build()
		if bytes.Equal(have, want) {
			continue
		}
		hl, wl := strings.Split(string(have), "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(hl) || i < len(wl); i++ {
			var h, w string
			if i < len(hl) {
				h = hl[i]
			}
			if i < len(wl) {
				w = wl[i]
			}
			if h != w {
				out = append(out, fmt.Sprintf("%s is not what the code builds: first difference at line %d\n  committed: %q\n  built:     %q", g.Path, i+1, h, w))
				break
			}
		}
	}
	return out
}

// GeneratedFilesSkill builds base's generated-files skill (spec section 6, "Generated files"; read on demand, never
// always on): a skill's frontmatter, then each kind of file Bonsai or an agent generates in a project, from
// format.GeneratedKinds, and the cleaning rules (5.2.6b) in plain words. Same rules as Page: LF, ASCII, no map-order
// iteration.
func GeneratedFilesSkill() []byte {
	var b strings.Builder
	line := func(f string, a ...any) { fmt.Fprintf(&b, f+"\n", a...) }
	line("---")
	line("name: generated-files")
	line("description: \"How long Bonsai keeps each kind of file it and its agents generate in a project (the log, asks, ladder " +
		"results, run reports, the sessions table, messages between agents, each agent's state), what is never cleaned, and how " +
		"bonsai.yaml's generated: section sets it. Read it before changing generated:, or to know whether a generated file may go.\"")
	line("---")
	line("")
	line("%s", engine.GeneratedNote)
	line("")
	line("# Generated files")
	line("")
	line("What Bonsai and its agents generate inside a project, who writes each kind, how long it is kept, what is never")
	line("cleaned and when cleaning happens. Bonsai deletes only what this skill says it may. Each kind's rule is set under")
	line("`generated:` in bonsai.yaml (a person's file); a kind or the whole section left out takes the defaults below.")
	line("")
	line("## The kinds")
	line("")
	for _, k := range format.GeneratedKinds {
		line("### `%s`", k.Kind)
		line("")
		line("- What: %s.", k.What)
		line("- Where it lives: %s.", k.Where)
		line("- Written by: %s.", k.Writer)
		line("- Default: %s.", k.DefaultWords())
		line("- Never cleaned: %s.", k.Never)
		line("- Cleaned: %s.", k.When)
		if k.Rule {
			line("- Rule in bonsai.yaml: `generated.%s` takes `keep_days` and `keep_newest`.", k.Kind)
		} else {
			line("- Rule in bonsai.yaml: none; this kind takes no rule.")
		}
		line("")
	}
	line("## The rules")
	line("")
	line("1. **The protections come first.** A protected file or row is never cleaned, whatever the rule says, and it still")
	line("   counts among the newest. The protections are the \"Never cleaned\" line of each kind above.")
	line("2. **`keep_days`** cleans a file or row older than that many days. Age is a log or asks file's last record (its")
	line("   modification time when no record reads), a ladder result's `finished` (its modification time when the result")
	line("   does not read), or a row's end.")
	line("   A file is cleaned only when both its modification time and its content say it is older than the rule.")
	line("3. **`keep_newest`** keeps only that many of the newest of its kind and cleans the rest.")
	line("4. Either one cleans. `null` for both keeps everything. A kind or the whole `generated:` section left out of")
	line("   bonsai.yaml takes the defaults above; a rule that does not read cleans nothing of its kind.")
	line("5. **Only Bonsai's own files** are cleaned, by the names it writes, regular files only, inside the kind's folder.")
	line("   A link is never followed, and nothing is cleaned in a folder that is itself a link. Anything else in the folder")
	line("   is left alone and never named.")
	line("6. **A `clean` record** goes in the log for every file or row cleaned, naming its path (a row: `.bonsai/sessions.md#`")
	line("   and its key) and the rule that cleaned it, written once the file is deleted (a row: once the table is written).")
	line("   A delete that finds the file gone writes nothing. A file another process holds open (Windows), or one written to")
	line("   since it was judged, is left for a later run.")
	line("7. **A session's end** cleans ladder results, asks and the log, oldest first, within 1 second; what it leaves waits")
	line("   for the next end. A cleaning killed between a delete and its record leaves the file gone with no record.")
	line("8. **Run reports are never deleted by Bonsai.** Past their rule, `bonsai check` lists them as a warning and a person")
	line("   deletes them.")
	return []byte(b.String())
}
