package engine

// bonsai.yaml as bonsai init writes it (spec §6: "It is written with a comment on every line"; format review 1.4),
// and init --new-id's new workspace id (contract §3).
//
// init writes bonsai.yaml from this template: base's documented template (spec §5) comes with the base pack at step
// 5.5. It holds the keys the walking skeleton reads (format, id, name, packs, protected, person_only, never_edit);
// spec §6's other keys (documents, ladder_floor, ladder, ratchets, ci_marked_tests, generated) come with the step
// that reads them.

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

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

// yamlQuote writes a text value in double quotes, with format 1's escapes (contract §2.4: \" \\ \n \r \t).
func yamlQuote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	return `"` + r.Replace(s) + `"`
}

// commented gives one line with its comment, the comments lined up at column 37 where the value leaves room.
func commented(text, comment string) string {
	pad := 36 - len(text)
	if pad < 2 {
		pad = 2
	}
	return text + strings.Repeat(" ", pad) + "# " + comment + "\n"
}

func flowList(items []string) string {
	q := make([]string, len(items))
	for i, s := range items {
		q[i] = yamlQuote(s)
	}
	return "[" + strings.Join(q, ", ") + "]"
}

// ConfigYAML is a new bonsai.yaml: id, the values init was given, and packID, the pack's id as its pack.yaml says.
func ConfigYAML(id string, v InitValues, packID string) []byte {
	var b strings.Builder
	b.WriteString("# bonsai.yaml: this project's Bonsai settings (format bonsai.workspace/1). Only a person changes this file.\n")
	b.WriteString("# bonsai init wrote it, a comment on every line. Change a value, then run bonsai update to apply it.\n")
	b.WriteString("# Every field and its allowed values: bonsai check --schema bonsai.workspace (that flag comes with step 5.1).\n")
	b.WriteString(commented("format: "+workspace.WorkspaceFormat, "the format and its version; a newer one is refused, never guessed"))
	b.WriteString(commented("id: "+id, "this project's id, written once by bonsai init; a copy gets its own (init --new-id)"))
	b.WriteString(commented("name: "+v.Name, "a short name: a lower-case letter, then lower-case letters, digits and dashes"))
	b.WriteString(commented("packs:", "the packs this project uses, applied in this order"))
	b.WriteString(commented("  - id: "+packID, "the pack's id, as its bonsai/pack.yaml says"))
	b.WriteString(commented("    source: "+yamlQuote(v.Source), "the git repository the pack comes from"))
	if v.Path != "" {
		b.WriteString(commented("    path: "+yamlQuote(v.Path), "the pack's folder inside that repository"))
	}
	b.WriteString(commented("    ref: "+yamlQuote(v.Ref), "the release tag or commit; change it and run bonsai update to take another"))
	b.WriteString("# protected: paths an agent changes only while its running task lists them in bonsai.allows\n")
	b.WriteString("protected: " + flowList(DefaultProtected) + "\n")
	b.WriteString("# person_only: of those, the paths only a person grants (an approve or grant tap)\n")
	b.WriteString("person_only: " + flowList(DefaultProtected) + "\n")
	b.WriteString("# never_edit: paths no agent ever changes; init and update write each one into .claude/settings.json\n")
	b.WriteString("# as the deny rule Edit(<path>)\n")
	b.WriteString("never_edit: " + flowList(v.NeverEdit) + "\n")
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
