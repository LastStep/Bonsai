package workspace

// bonsai.yaml (bonsai.workspace/1, spec §6): the project's settings, at the checkout's root, the one Bonsai file a
// person edits.
//
// Read here (what the walking skeleton's parts 2 to 5 use): format, id, name, packs (id, source, path, ref),
// protected, person_only and never_edit (part 3 writes each never_edit path as a deny rule). Every other key spec §6
// shows (documents, ladder_floor, ladder, ratchets, ci_marked_tests, generated) and any unknown key is kept in
// Config.Doc as read and not checked: their meaning and checks are step 5.1's, with bonsai.workspace/1's schema.

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// ConfigFile is bonsai.yaml's name, at a checkout's root.
const ConfigFile = "bonsai.yaml"

// WorkspaceFormat is the format: line bonsai.yaml carries.
const WorkspaceFormat = "bonsai.workspace/1"

// Config is what Bonsai reads from bonsai.yaml.
type Config struct {
	ID         string      // the workspace id: ws- and 26 characters (contract §3)
	Name       string      // a slug: [a-z][a-z0-9-]{0,39} (contract §3)
	Packs      []PackRef   // the packs, in the order they apply
	Protected  []string    // globs an agent changes only while its running task lists them (spec §6); [] for none
	PersonOnly []string    // of those, the globs only a person grants (contract §5.5); [] for none
	NeverEdit  []string    // paths no agent ever changes, written as Edit deny rules (spec §6, §7); [] for none
	Doc        *reader.Map // the whole file as read, every key kept
}

// PackRef is one entry of bonsai.yaml's packs.
type PackRef struct {
	ID     string // the pack's id, as its pack.yaml says
	Source string // the git repository the pack comes from
	Path   string // the pack's folder inside that repository, "" for its root
	Ref    string // the release tag (spec §5); the lock records the commit it resolves to
	Line   int    // the line of the entry in bonsai.yaml
}

// The id and name patterns are the ones bonsai.status/1 prints them under (formats/schemas/status.schema.json), so
// bonsai.yaml never accepts a value status --json could not print.
var (
	idnameOnce             sync.Once
	idPattern, namePattern *regexp.Regexp
)

func idName() (*regexp.Regexp, *regexp.Regexp) {
	idnameOnce.Do(func() {
		raw, err := formats.Schema("status")
		if err != nil {
			panic(err)
		}
		s, err := schema.Parse(raw)
		if err != nil {
			panic(err)
		}
		props := func(o schema.Object, key string) schema.Object {
			v, _ := o.Get(key)
			out, _ := v.(schema.Object)
			return out
		}
		ws := props(props(props(s, "properties"), "workspace"), "properties")
		idPattern = regexp.MustCompile(props(ws, "id").String("pattern"))
		namePattern = regexp.MustCompile(props(ws, "name").String("pattern"))
	})
	return idPattern, namePattern
}

const configNext = "fix that line of bonsai.yaml, or restore the file from git"

// ReadConfig reads bonsai.yaml's bytes.
func ReadConfig(raw []byte) (*Config, error) {
	m, err := readYAML(ConfigFile, raw, WorkspaceFormat, configNext)
	if err != nil {
		return nil, err
	}
	f := &fields{file: ConfigFile, next: configNext}
	c := &Config{Doc: m}
	idp, namep := idName()
	c.ID = f.text(m, "id", "the file", 0, true)
	if e, _ := m.Entry("id"); f.err == nil && !idp.MatchString(c.ID) {
		f.fail(e.Line, "the id %s is not ws- and 26 lower-case letters or digits", showValue(c.ID))
	}
	c.Name = f.text(m, "name", "the file", 0, true)
	if e, _ := m.Entry("name"); f.err == nil && !namep.MatchString(c.Name) {
		f.fail(e.Line, "the name %s is not a slug: a lower-case letter, then up to 39 lower-case letters, digits and dashes", showValue(c.Name))
	}
	items, line := f.list(m, "packs", "the file")
	c.Packs = []PackRef{}
	seen := map[string]bool{}
	for i, it := range items {
		pm, ok := it.(*reader.Map)
		if !ok {
			f.fail(line, "packs item %d is %s, not a mapping of id, source, path and ref", i+1, kindOf(it))
			break
		}
		where := "packs item " + itoa(i+1)
		first := 0
		if es := pm.Entries(); len(es) > 0 {
			first = es[0].Line
		}
		p := PackRef{Line: first}
		p.ID = f.text(pm, "id", where, first, true)
		if f.err == nil && !packIDPattern.MatchString(p.ID) {
			f.fail(first, "%s's id %s is not a pack id: a lower-case letter, then lower-case letters, digits and dashes", where, showValue(p.ID))
		}
		if f.err == nil && seen[p.ID] {
			f.fail(first, "the pack %s is listed twice", showValue(p.ID))
		}
		seen[p.ID] = true
		p.Source = f.text(pm, "source", where, first, true)
		if e, _ := pm.Entry("source"); f.err == nil && strings.HasPrefix(p.Source, "-") {
			f.fail(e.Line, "%s's source %s starts with -, which git would read as an option", where, showValue(p.Source))
		}
		p.Path = f.text(pm, "path", where, first, false)
		if f.err == nil && p.Path != "" {
			if err := CheckRelPath(p.Path); err != nil {
				e, _ := pm.Entry("path")
				f.fail(e.Line, "%s's path: %v", where, err)
			}
		}
		p.Ref = f.text(pm, "ref", where, first, true)
		if e, _ := pm.Entry("ref"); f.err == nil && strings.HasPrefix(p.Ref, "-") {
			f.fail(e.Line, "%s's ref %s starts with -, which git would read as an option", where, showValue(p.Ref))
		}
		c.Packs = append(c.Packs, p)
	}
	c.Protected = f.texts(m, "protected", "the file")
	c.PersonOnly = f.texts(m, "person_only", "the file")
	c.NeverEdit = f.texts(m, "never_edit", "the file")
	for _, key := range []string{"protected", "person_only", "never_edit"} {
		e, _ := m.Entry(key)
		list := c.Protected
		switch key {
		case "person_only":
			list = c.PersonOnly
		case "never_edit":
			list = c.NeverEdit
		}
		for _, g := range list {
			if f.err == nil && !relativeGlob(g) {
				f.fail(e.Line, "the %s glob %s is not project-relative with forward slashes", key, showValue(g))
			}
			// A never_edit path becomes the deny rule Edit(<path>) in .claude/settings.json: a parenthesis would end
			// the rule early, and a leading ~ would make it a path in the home folder, not the project.
			if f.err == nil && key == "never_edit" && (strings.ContainsAny(g, "()") || strings.HasPrefix(g, "~")) {
				f.fail(e.Line, "the never_edit path %s holds a parenthesis or starts with ~; it becomes the deny rule Edit(<path>)", showValue(g))
			}
		}
	}
	if f.err != nil {
		return nil, f.err
	}
	return c, nil
}

// relativeGlob reports whether a glob is project-relative with forward slashes: no leading slash, no backslash, no
// drive colon (contract §2.6: no absolute path in a committed format).
func relativeGlob(g string) bool {
	return g != "" && !strings.HasPrefix(g, "/") && !strings.ContainsAny(g, `\:`)
}

// LoadConfig reads root/bonsai.yaml. A missing file is an *Error whose errors.Is(err, fs.ErrNotExist) holds.
func LoadConfig(root string) (*Config, error) {
	raw, err := os.ReadFile(filepath.Join(root, ConfigFile))
	if err != nil {
		e := &Error{File: ConfigFile, Msg: "cannot be read: " + oneLine(err), Err: err,
			Next: "check the file's permissions, then run the command again"}
		if errors.Is(err, fs.ErrNotExist) {
			e.Msg = "is not in this checkout (" + filepath.ToSlash(root) + "), so it is not linked to Bonsai"
			e.Next = "run bonsai init in the project's main checkout"
		}
		return nil, e
	}
	return ReadConfig(raw)
}
