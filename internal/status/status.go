// Package status builds bonsai status's document, bonsai.status/1 (contract §12): one workspace at a glance.
//
// Plan part 2 builds the partial status: every field of bonsai.status/1, in the schema's order (the schema in
// formats/schemas, embedded by package formats, is the one list of fields), filled where the walking skeleton has
// built it and null (or [] for a list) where it has not. Built here: format, bonsai, mode, workspace, home, local,
// person_only and problems. The rest wait for later parts, which fill them as they are built; this package's test
// holds them in a named list.
//
// Exit codes (contract §12, spec §3): 0, or 3 when Bonsai cannot read the workspace at all; the document then
// fills format, bonsai and problems, and every other field is null.
package status

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Format is the document's format.
const Format = "bonsai.status/1"

// Mode is how this status is gathered: offline, cheap, no network (contract §12).
const Mode = "offline"

var (
	schemaOnce   sync.Once
	statusSchema schema.Object
)

// Schema is bonsai.status/1's schema, from formats/schemas (embedded).
func Schema() schema.Object {
	schemaOnce.Do(func() {
		raw, err := formats.Schema("status")
		if err != nil {
			panic(err)
		}
		if statusSchema, err = schema.Parse(raw); err != nil {
			panic(err)
		}
	})
	return statusSchema
}

// Exit codes.
const (
	ExitOK      = 0
	ExitRuntime = 3 // Bonsai cannot read the workspace at all (contract §12)
)

// Build gathers the status of the workspace holding dir; version is this Bonsai's own. It returns the document and
// its exit code.
func Build(dir, version string) (schema.Object, int) {
	built, problem := gather(dir)
	if problem != "" {
		return document(map[string]any{
			"format": Format, "bonsai": version, "problems": []any{problem},
		}, true), ExitRuntime
	}
	built["format"], built["bonsai"] = Format, version
	return document(built, false), ExitOK
}

// gather reads what part 2 builds. A problem is one sentence that names its next step.
func gather(dir string) (map[string]any, string) {
	co, err := workspace.Find(dir)
	if err != nil {
		return nil, err.Error()
	}
	cfg, err := workspace.LoadConfig(co.Root)
	if err != nil {
		return nil, err.Error()
	}
	home, err := workspace.Home()
	if err != nil {
		return nil, err.Error()
	}
	key, err := workspace.MachineKey(co.Main)
	if err != nil {
		return nil, fmt.Sprintf("the main checkout %s cannot be resolved: %v; next: check that it exists",
			filepath.ToSlash(co.Main), err)
	}
	root := filepath.ToSlash(co.Main)
	personOnly := make([]any, len(cfg.PersonOnly))
	for i, g := range cfg.PersonOnly {
		personOnly[i] = g
	}
	return map[string]any{
		"mode": Mode,
		"workspace": schema.Object{
			{Key: "id", Value: cfg.ID}, {Key: "name", Value: cfg.Name}, {Key: "root", Value: root},
		},
		"home": schema.Object{{Key: "path", Value: filepath.ToSlash(home)}, {Key: "key", Value: key}},
		"local": schema.Object{
			{Key: "log", Value: root + "/.bonsai/local/log"},
			{Key: "asks", Value: root + "/.bonsai/local/asks"},
			{Key: "ladder", Value: root + "/.bonsai/local/ladder"},
		},
		"person_only": personOnly,
		"problems":    []any{},
	}, ""
}

// document writes every field of the schema, in its order: a built field's value, else [] for a field whose type
// allows a list and null for any other (contract §2.2: a writer writes every field, null or [] where it does not
// apply). On exit 3 every field not built is null.
func document(built map[string]any, failed bool) schema.Object {
	props, _ := Schema().Get("properties")
	doc := schema.Object{}
	for _, p := range props.(schema.Object) {
		v, ok := built[p.Key]
		if !ok && !failed && allowsList(p.Value.(schema.Object)) {
			v = []any{}
		}
		doc = append(doc, schema.Member{Key: p.Key, Value: v})
	}
	return doc
}

func allowsList(prop schema.Object) bool {
	t, _ := prop.Get("type")
	switch x := t.(type) {
	case string:
		return x == "array"
	case []any:
		for _, e := range x {
			if e == "array" {
				return true
			}
		}
	}
	return false
}

// Text renders the document for a person: plain ASCII, the same facts as the JSON (spec §6's closing words of
// init, which status shows too). On exit 3 it gives the problems alone.
func Text(doc schema.Object) string {
	var b strings.Builder
	problems, _ := doc.Get("problems")
	ws, _ := doc.Get("workspace")
	w, ok := ws.(schema.Object)
	if !ok {
		b.WriteString("bonsai status: cannot read this workspace.\n")
		for _, p := range problems.([]any) {
			b.WriteString("  " + ascii(p.(string)) + "\n")
		}
		return b.String()
	}
	hv, _ := doc.Get("home")
	home := hv.(schema.Object)
	fmt.Fprintf(&b, "Workspace %s, id %s (every clone and worktree of this project shares it).\n",
		ascii(w.String("name")), ascii(w.String("id")))
	fmt.Fprintf(&b, "In this project, %s:\n", ascii(w.String("root")))
	b.WriteString("  bonsai.yaml      this project's Bonsai settings. You edit it; it is committed.\n")
	b.WriteString("  .bonsai/         the lock, STATE and two tables Bonsai rebuilds (tasks; sessions and hours). Committed.\n")
	b.WriteString("  .bonsai/local/   the log, questions for you and their answers, ladder results. Never committed.\n")
	b.WriteString("On this machine:\n")
	fmt.Fprintf(&b, "  %s\n", ascii(home.String("path")))
	b.WriteString("    Bonsai's home: a secret salt, this machine's settings for each project, label files attached\n")
	b.WriteString("    on this machine, your personal memory, the pack cache. None of it enters git.\n")
	fmt.Fprintf(&b, "    This project's machine folder: workspaces/%s\n", ascii(home.String("key")))
	po, _ := doc.Get("person_only")
	var globs []string
	for _, g := range po.([]any) {
		globs = append(globs, ascii(g.(string)))
	}
	if len(globs) == 0 {
		globs = []string{"none"}
	}
	fmt.Fprintf(&b, "Person-only paths: %s\n", strings.Join(globs, ", "))
	list := problems.([]any)
	if len(list) == 0 {
		b.WriteString("Problems: none.\n")
	} else {
		b.WriteString("Problems:\n")
		for _, p := range list {
			b.WriteString("  " + ascii(p.(string)) + "\n")
		}
	}
	b.WriteString("This build of Bonsai's rebuild shows the workspace, its folders and the home; packs, files, labels,\n")
	b.WriteString("lanes, the active task and needs come with its later parts.\n")
	return b.String()
}

// ascii keeps printable ASCII and writes any other character as a Go escape, so the text prints unbroken in
// PowerShell 5.1 (spec §3).
func ascii(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x20 && r < 0x7f {
			b.WriteRune(r)
			continue
		}
		q := fmt.Sprintf("%+q", string(r))
		b.WriteString(q[1 : len(q)-1])
	}
	return b.String()
}
