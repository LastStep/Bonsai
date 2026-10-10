package engine

// bonsai.yaml as bonsai init writes it, from Bonsai's built-in template (spec §6: "It is written with a comment on
// every line"; format review 1.4; base's own documented template replaces it at step 5.5), and init --new-id's new
// workspace id (contract §3).
//
// The template is two things here, its one home: the values init writes into each field of bonsai.workspace/1 that
// it is not given (DefaultDocuments, DefaultProtected, the generated kinds' defaults from format.GeneratedKinds, and
// [] or {} for the rest), and Comments, the comment written on every field's line (or, for a long one, the line
// just above), for a person and an agent who edit the file unaided. Every line is a field's or a comment: a header
// says what the file is and how to change it, and each field's comment says what it means, its allowed values or
// the command that lists them, and an example when the field is empty. TestTemplateDocumentsEveryField holds
// Comments to the schema's fields both ways, so a field change updates its comment in the same commit. Writing goes
// through format.Workspace.EncodeYAML, which holds the document to the schema and reads it back before a byte is
// written.

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"regexp"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// InitValues are the values init needs to write a new bonsai.yaml: spec §14 check 1's "required values".
type InitValues struct {
	Name      string   // the workspace's name, a slug
	Source    string   // the pack's git repository
	Path      string   // the pack's folder inside it, "" for its root
	Ref       string   // the release tag or 40-character commit to take
	NeverEdit []string // paths no agent ever changes
}

// DefaultProtected is the protected list and person_only list init writes: bonsai.yaml, the lock and Claude Code's
// project folder (contract §14: "bonsai.yaml and the lock are on base's protected list"; spec §6's person_only,
// contract §18 B (c)). A person adds to them.
var DefaultProtected = []string{".claude/**", "bonsai.yaml", ".bonsai/lock.json"}

// NewID returns a new workspace id: ws- and 26 lower-case base32 characters (contract §3), from the system's
// random source.
func NewID() (string, error) {
	b := make([]byte, 17)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "ws-" + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))[:26], nil
}

// DefaultDocuments is where init's bonsai.yaml keeps the documents Bonsai knows (contract §7.3): a work/ folder at the
// project's top. A project changes any of them in bonsai.yaml (an existing project keeps its own folder names).
var DefaultDocuments = format.Documents{Task: "work/tasks", Run: "work/runs", Answers: "work/answers.md",
	Memory: "work/memory", Protocols: "work/protocols"}

// ConfigHeader is the comment block at the top of init's bonsai.yaml: what the file is, who changes it, how a change
// takes effect, and where every field is documented.
var ConfigHeader = []string{
	"bonsai.yaml: this project's Bonsai settings (format bonsai.workspace/1), the one Bonsai file a person edits.",
	"bonsai init wrote it from Bonsai's built-in template, a comment on every key. It is committed and person-only:",
	"an agent changes it only while its running task's bonsai.allows holds it, which a person grants.",
	"To change a setting, edit its value, then run: bonsai update (it previews every change; --yes writes them).",
	"Every field, its allowed values and an example: bonsai check --schema bonsai.workspace",
}

// Comments are the template's comment for each field, by its path in the document (a list's items as []: a pack's
// fields are packs[].id). A pack's document kind under documents and a generated kind's fields take theirs from
// comment's rules below. Kept under 110 characters, so the line above holds a long line's comment.
var Comments = map[string]string{
	"format":              "the format and its version; a newer one is refused, never guessed",
	"id":                  "this project's id, written once by bonsai init; a copy meant as a new project: bonsai init --new-id",
	"name":                "a short name: a lower-case letter, then up to 39 lower-case letters, digits and dashes",
	"packs":               "the packs this project uses, applied in this order (a path two packs write stops update)",
	"packs[]":             "one pack",
	"packs[].id":          "the pack's id, as its bonsai/pack.yaml says",
	"packs[].source":      "the git repository the pack comes from, fetched with this machine's git",
	"packs[].path":        "the pack's folder inside that repository; null for the repository's top",
	"packs[].ref":         "the release tag or a 40-character commit; change it, then run bonsai update to take it",
	"documents":           "where this project keeps the documents Bonsai knows: a folder (one file each) or one file",
	"documents.task":      "the tasks' folder: one file per task, named <id>-<slug>.md (bonsai.task/1)",
	"documents.run":       "the run reports' folder: one report per agent session (bonsai.run/1)",
	"documents.answers":   "the file for answers to questions that name no file of their own",
	"documents.memory":    "the memory notes' folder, with their INDEX.md, which the block in CLAUDE.md imports",
	"documents.protocols": "the folder the packs' always-on protocol files go to; the block in CLAUDE.md imports each",
	"protected":           "paths (globs) an agent changes only while its running task lists them in bonsai.allows",
	"person_only":         "of those, the paths only a person grants (an approve or grant tap)",
	"never_edit":          `paths no agent ever changes, each a deny rule Edit(<path>) in .claude/settings.json; e.g. ["work/ledger.json"]`,
	"ladder_floor":        "the rungs every task climbs, whatever its bonsai.ladder lists; e.g. [0, 1]",
	"ladder":              "the rungs: what each runs and proves (rung, name, kind, command, required, timeout_s and more)",
	"ladder[]":            "one rung; its fields: bonsai check --schema bonsai.workspace",
	"ladder[].rung":       "the rung's number; ladder_floor and a task's bonsai.ladder name rungs by it",
	"ladder[].name":       "its short name, or null",
	"ladder[].kind":       "guard (rung 0: the diff against the task's grants), command, or ver-git",
	"ladder[].command":    "the shell command a command rung runs (never bash by name); null for the other kinds",
	"ladder[].required":   "true: a red result here skips the rungs after it",
	"ladder[].timeout_s":  "the longest it may run, in seconds; null for the runner's default",
	"ladder[].ratchet":    "the ratchet count it feeds (see ratchets); null for none",
	"ladder[].capture":    "the patterns that pick numbers or text out of its output, by name; null for none",
	"ladder[].means":      "one line: what it proves when green; null for none",
	"ladder[].tests":      "the form of its test output: TAP, go test -json or JUnit; null when it prints no tests",
	"ladder[].base_setup": "the command run first in a worktree at the merge base (installing dependencies); null for none",
	"ratchets":            "counts that may only rise, by name, each the floor a rung's count may not fall below; e.g. tests: 120",
	"ci_marked_tests":     `tests allowed to skip in CI only, each by its name in the test output; e.g. ["TestNeedsADisplay"]`,
	"generated":           "how long generated files are kept; null keeps all; docs/reference/generated-files.md in LastStep/Bonsai",
}

// comment gives a line's comment by its field's path ("packs[0].ref"): Comments' entry for the path with its list
// indexes written as [], a pack's document kind's from its declaration, a generated kind's from format.GeneratedKinds.
func comment(kinds map[string]string) func(path string) string {
	return func(path string) string {
		key := listIndex.ReplaceAllString(path, "[]")
		if c, ok := Comments[key]; ok {
			return c
		}
		if kind, ok := strings.CutPrefix(key, "documents."); ok {
			if c, ok := kinds[kind]; ok {
				return c
			}
		}
		if rest, ok := strings.CutPrefix(key, "generated."); ok {
			kind, field, _ := strings.Cut(rest, ".")
			g, ok := format.GeneratedKindOf(kind)
			if !ok {
				return ""
			}
			switch field {
			case "":
				return g.What + ", in " + g.Where + "; default: " + g.Default
			case "keep_days":
				return "clean what is older than this many days; null: no age rule"
			case "keep_newest":
				return "keep only this many of the newest; null: no count rule"
			}
		}
		return ""
	}
}

// quoted says which values init writes in double quotes: a pack's source, folder and ref, which a person edits (a
// commit pasted over a tag stays text).
func quoted(path string) bool {
	switch listIndex.ReplaceAllString(path, "[]") {
	case "packs[].source", "packs[].path", "packs[].ref":
		return true
	}
	return false
}

var listIndex = regexp.MustCompile(`\[[0-9]+\]`)

// ConfigYAML is a new bonsai.yaml from the template: id, init's values and the pack read at its ref (its id, and its
// document kinds, each written under documents at the pack's default place).
func ConfigYAML(id string, v InitValues, pack *PackData) ([]byte, error) {
	w := &format.Workspace{ID: id, Name: v.Name, Documents: DefaultDocuments, Protected: DefaultProtected,
		PersonOnly: DefaultProtected, NeverEdit: v.NeverEdit, LadderFloor: []int64{}, Ladder: []format.Rung{},
		Ratchets: schema.Object{}, CIMarkedTests: []string{}, Generated: format.DefaultGenerated()}
	if w.NeverEdit == nil {
		w.NeverEdit = []string{}
	}
	ref := format.WorkspacePack{Source: v.Source, Ref: v.Ref}
	if v.Path != "" {
		path := v.Path
		ref.Path = &path
	}
	kinds := map[string]string{}
	w.Documents.Extra = nil
	if pack != nil {
		ref.ID = pack.Manifest.ID
		if pack.Declares != nil {
			for _, k := range pack.Declares.Documents {
				what, place := "folder (one file per document)", k.Path
				if k.File != nil {
					what, place = "file", k.File
				}
				if place == nil {
					continue
				}
				w.Documents.Extra = append(w.Documents.Extra, schema.Member{Key: k.Kind, Value: *place})
				kinds[k.Kind] = "the pack " + ref.ID + "'s document kind " + k.Kind + ": its " + what
			}
		}
	}
	w.Packs = []format.WorkspacePack{ref}
	body, err := w.EncodeYAMLWith(format.YAMLOptions{Comment: comment(kinds), Quote: quoted})
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for _, h := range ConfigHeader {
		b.WriteString("# " + h + "\n")
	}
	b.Write(body)
	return []byte(b.String()), nil
}

// probeYAML is a bonsai.yaml holding only init's values, which init reads with bonsai.yaml's own reader before it
// fetches anything, so a value in a form bonsai.yaml refuses is refused with the reader's own words.
func probeYAML(v InitValues) []byte {
	q := func(s string) string {
		r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
		return `"` + r.Replace(s) + `"`
	}
	var b strings.Builder
	b.WriteString("format: " + workspace.WorkspaceFormat + "\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: " + v.Name + "\npacks:\n")
	b.WriteString("  - id: probe\n    source: " + q(v.Source) + "\n")
	if v.Path != "" {
		b.WriteString("    path: " + q(v.Path) + "\n")
	}
	b.WriteString("    ref: " + q(v.Ref) + "\n")
	var never []string
	for _, p := range v.NeverEdit {
		never = append(never, q(p))
	}
	b.WriteString("never_edit: [" + strings.Join(never, ", ") + "]\n")
	return []byte(b.String())
}

// replaceID gives bonsai.yaml's bytes with a new id on the id line, every other byte as it was.
func replaceID(raw []byte, cfg *workspace.Config, newID string) ([]byte, error) {
	e, ok := cfg.Doc.Entry("id")
	if !ok || e.Line < 1 {
		return nil, fmt.Errorf("bonsai.yaml has no id line")
	}
	lines := bytes.SplitAfter(raw, []byte("\n"))
	if e.Line > len(lines) {
		return nil, fmt.Errorf("bonsai.yaml has no line %d", e.Line)
	}
	l := lines[e.Line-1]
	i := bytes.Index(l, []byte(cfg.ID))
	if i < 0 {
		return nil, fmt.Errorf("bonsai.yaml line %d does not hold the id %s as written", e.Line, cfg.ID)
	}
	var out bytes.Buffer
	for j, x := range lines {
		if j == e.Line-1 {
			out.Write(l[:i])
			out.WriteString(newID)
			out.Write(l[i+len(cfg.ID):])
			continue
		}
		out.Write(x)
	}
	return out.Bytes(), nil
}
