package engine

// bonsai check --pack <folder> (plan-5 5.1.9; spec §5, "Every template and pack file documents itself" and the pack's
// CI; contract §2.8): a pack's folder held to the rules that keep its declarations and its documentation in step with
// its fields. It reads the folder as it is on disk (a pack's checkout, in its CI) and nothing else: no project, no
// Bonsai home, no git, no network, no claude (`claude plugin validate` is the pack's CI's other step, step 5.5). Each
// rule's finding has a word of format.PackCheckWords, their one home; every one is a finding (exit 1), none a warning.
// In the order CheckPack looks, file by file, what counts for each rule:
//
//   - pack-schema: bonsai/pack.yaml, and bonsai/labels.yaml and bonsai/lanes.yaml when the pack has them, as their
//     writer writes them (format.WriterProblems, the rule format.Check holds Bonsai's own writers to; a pack's maker
//     is these files' writer, contract §2.2): format 1's YAML, the format line, every field the schema names present
//     ([] or null where none applies, so a pack with no document kinds writes documents: []), each of a type and value
//     the schema allows, in the schema's order; every problem is listed. Then what the engine refuses at a link that a
//     schema cannot say: pack.yaml as workspace.ReadPack reads it (the id, a file's target, a path's form, an empty
//     text), the block and each files entry's source present in the pack, and labelsProblems and lanesProblems
//     (declares.go). The engine's read of pack.yaml stops at its first refusal, so that one is listed alone, and only
//     when pack.yaml has no writer's problem and no finding of the rules below that the engine's read also holds
//     (pack-why, pack-documents, pack-protected, pack-bash): its refusal would most often be that finding again. Fix
//     those and run it again.
//   - pack-comment: every key of those three files, at every depth (each mapping's keys, a list item's among them),
//     carries a comment: on its own line after its value (reader.KeyLineComment, as the format-1 reader sees comments:
//     a # at a line's start or after a space, outside quotes; contract §2.4), or on the line just above it, a comment
//     line (nothing but its indentation before its #). So a header's last line counts for format:, and a key whose
//     line is too long for its comment takes it on the line above (spec §5). A file the reader refuses is pack-schema's
//     alone.
//   - pack-why: each deny rule in pack.yaml has a why: present, text, not blank. The schema's problems with a deny
//     rule's why are this rule's, not pack-schema's.
//   - pack-documents: each document kind pack.yaml declares is well formed: its fields as the schema allows (the
//     schema's problems inside the documents items are this rule's, not pack-schema's); then, for a kind that fits,
//     the engine's rules (documentsProblems: a name of its own, once; exactly one of path and file, each a project
//     path; an id pattern Go's regexp reads) and two of check --pack's own, which the engine does not refuse: each
//     status once, and each move (person, agent) a pair of two of the kind's statuses, each stamp key one of them.
//   - pack-protected: each protected glob pack.yaml declares is well formed: its form as the schema allows (the
//     schema's problems inside protected are this rule's); then, for a glob that fits, project-relative with forward
//     slashes (protectedProblems, the engine's), no empty segment but a trailing slash's, no . or .. segment, and
//     each segment but ** one path.Match reads, as the guard reads a glob (internal/guard's checkGlob).
//   - pack-bash: no hook command calls bash by name (format.CallsBash, spec §3): each hooks entry's command in
//     pack.yaml, and every command string in the plugin's own hooks (hooks/hooks.json, and a hooks object in
//     .claude-plugin/plugin.json), which Claude Code runs on its own.
//   - pack-runs: each hook command in pack.yaml that names a file the pack writes (a files entry's path in the
//     project, in full or by its file name: namesPath, the engine's own reading at a link, checkRuns) lists that path
//     in its runs (step 5.1.1, rule 4). A runs item naming another pack's file is the engine's to judge at a link,
//     where it sees every linked pack.
//   - pack-block: the instruction block a project linked to this pack alone would get is at most MaxBlockLines (spec
//     §6): blockBody, the engine's own, for a workspace with init's documents (DefaultDocuments) and this pack (its
//     files in the protocols folder imported, its label definitions, its block.md without the leading comment).
//     Looked at only when pack.yaml reads as the engine reads it.
//   - pack-plugin-version: .claude-plugin/plugin.json, when the pack has one, carries no version (manifestVersion, the
//     link's own rule); one that is not a JSON object Bonsai reads is this rule's too, since its version cannot be
//     told.
//   - pack-fields and pack-values: each template skill. A template skill is skills/<name>/SKILL.md whose body (after
//     its frontmatter) holds a fields table: a markdown table whose header row's cells are Field, Meaning, Allowed
//     values and Example, in that order, letter case aside (spec §5). Its template is the first fenced code block
//     (``` or ~~~) after the table: one whose first line is --- is a markdown template, whose fields are its
//     frontmatter's top-level keys; any other is a YAML template (bonsai.yaml's, say), whose fields are its own
//     top-level keys. Keys are read line by line, since a template holds placeholders a reader would refuse: a line
//     at column 0 that is key: or key: value, the key of a field's form ([a-z][a-z0-9_]*). A row's field is its first
//     cell, its backquotes taken off; a row may name a nested field (needs.claude_code, files[].path), which counts as
//     its top-level key's.
//     pack-fields: every top-level field of the template has a row of its own name, and every row names a field the
//     template holds; a fields table with no template after it is one too. For a Bonsai format (the template's format:
//     line names bonsai.<name>/<major>), the format is one this Bonsai has, and the template's fields equal its
//     schema's, as sets.
//     pack-values: a row's allowed-values cell names a closed list when it lists values: words in backquotes, each one
//     word (null left aside: it says the field may be null). When the row's field has a closed list, the values equal
//     it, as sets: for a Bonsai format, the field's enum or const in its schema (or its items' enum); else, in a
//     template skill named after a document kind pack.yaml declares, its status field's statuses; a lane field, the
//     pack's lanes; a label's field (labels.<name>, or the label's name), a choice label's values in the pack's
//     labels.yaml. A cell that lists no value, naming where the list is defined instead, passes.
//
// Every path in a finding is the pack's own, relative to its folder, with forward slashes. Findings come in the order
// above, each file's by line, so the output is byte-stable.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The pack files check --pack reads beside bonsai/pack.yaml, labels.yaml and lanes.yaml.
const (
	pluginHooksFile = "hooks/hooks.json"
	skillsDir       = "skills"
)

// CheckPack checks the pack folder dir (check --pack). Its error is a refusal: the folder is not a pack's (exit 4,
// not-a-pack), or one of its files cannot be read (exit 3).
func CheckPack(dir string) (*CheckResult, *Error) {
	notPack := func(what string) *Error {
		return errorf("not-a-pack", ExitState, "run bonsai check --pack with a pack's folder: the folder that holds bonsai/pack.yaml "+
			"(a pack in a repository's folder is checked there)", "%s %s", ascii(filepath.ToSlash(dir)), what)
	}
	st, err := os.Stat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, notPack("is not there, so it is not a pack's folder")
	case err != nil:
		return nil, errorf("read-failed", ExitRuntime, "check the folder's permissions, then run the command again", "%s cannot be read: %v",
			ascii(filepath.ToSlash(dir)), err)
	case !st.IsDir():
		return nil, notPack("is a file, not a pack's folder")
	}
	c := &packCheck{dir: dir, r: &CheckResult{Root: dir}}
	raw, exists, e := c.read(workspace.PackFile)
	if e != nil {
		return nil, e
	}
	if !exists {
		return nil, notPack("holds no " + workspace.PackFile + ", so it is not a pack's folder")
	}
	pf := c.packYAML(raw)
	labels, lanes, e := c.declarations(pf)
	if e != nil {
		return nil, e
	}
	if e := c.block(pf, labels); e != nil {
		return nil, e
	}
	if e := c.plugin(); e != nil {
		return nil, e
	}
	skills, e := c.templates(pf, labels, lanes)
	if e != nil {
		return nil, e
	}
	what := []string{workspace.PackFile}
	for _, f := range []struct {
		name string
		has  bool
	}{{LabelsFile, labels != nil || c.has[LabelsFile]}, {LanesFile, lanes != nil || c.has[LanesFile]}} {
		if f.has {
			what = append(what, f.name)
		}
	}
	note := fmt.Sprintf("checked the pack folder %s: %s; %d %s", filepath.ToSlash(dir), strings.Join(what, ", "),
		len(skills), plural(len(skills), "template skill", "template skills"))
	if len(skills) > 0 {
		note += " (" + strings.Join(skills, ", ") + ")"
	}
	c.r.Notes = append(c.r.Notes, note)
	return c.r, nil
}

// packCheck is one run of check --pack.
type packCheck struct {
	dir string
	r   *CheckResult
	has map[string]bool // the optional files the folder holds (labels.yaml, lanes.yaml), even when they do not read
}

// packFound is one finding before it is added: its word (format.PackCheckWords), file, sentence and next step.
type packFound struct {
	code, file, msg, next string
}

// addPack adds a check --pack finding. A word that is not in format.PackCheckWords is a bug, which
// TestCheckWordsInTheCode rules out.
func (r *CheckResult) addPack(code, file, msg, next string) {
	w, ok := format.PackCheckWord(code)
	if !ok {
		panic("engine: the check --pack word " + code + " is not in format.PackCheckWords")
	}
	r.Findings = append(r.Findings, Finding{Code: code, File: file, Message: msg, Next: next, Who: w.Who})
}

func (c *packCheck) put(fs ...packFound) {
	for _, f := range fs {
		c.r.addPack(f.code, f.file, f.msg, f.next)
	}
}

// read reads a file of the pack by its path there: exists false when it is not there.
func (c *packCheck) read(rel string) ([]byte, bool, *Error) {
	raw, err := os.ReadFile(filepath.Join(c.dir, filepath.FromSlash(rel)))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, false, nil
	case err != nil:
		return nil, false, errorf("read-failed", ExitRuntime, "check the file's permissions, then run the command again",
			"the pack's %s cannot be read: %v", rel, err)
	}
	return raw, true, nil
}

// at names a place in a pack file: "bonsai/pack.yaml line 4", or the file alone for line 0.
func at(file string, line int) string {
	if line > 0 {
		return file + " line " + strconv.Itoa(line)
	}
	return file
}

// readYAML reads a pack's YAML file under format 1: its mapping, or nil with the reader's refusal as a pack-schema
// finding.
func (c *packCheck) readYAML(file, formatName string, raw []byte) *reader.Map {
	res := reader.ReadYAML(raw)
	switch res.Outcome {
	case reader.Refused:
		msg := res.Refusal.Message
		if res.Refusal.Code != "" {
			msg += " (" + res.Refusal.Code + ")"
		}
		c.put(packFound{"pack-schema", file, at(file, res.Refusal.Line) + ": " + msg, res.Refusal.Next})
		return nil
	case reader.Format0:
		v := format.MustLookup(formatName).Versioned()
		c.put(packFound{"pack-schema", file, file + " has no format: line first, so it does not read as " + v,
			"start the file with format: " + v})
		return nil
	}
	return res.Value
}

// writerFound is a writer's problem (format.WriterProblems) as a finding of rule code.
func writerFound(code, file, formatName string, p format.Problem) packFound {
	next := "fix that field in " + file + " (every field, in order, with its allowed values: bonsai check --schema bonsai." + formatName + ")"
	if name, ok := strings.CutPrefix(p.Msg, `required field "`); ok {
		name = strings.TrimSuffix(name, `" is missing`)
		where, line := "the file", 0 // the document's own field has no line to name
		if p.Field != "(the document)" {
			where, line = p.Field, p.Line
		}
		return packFound{code, file, fmt.Sprintf("%s: %s has no field %s (a writer writes every field, [] or null where none applies: "+
			"contract section 2.2)", at(file, line), where, name), "add " + name + " to " + strings.TrimPrefix(where+" in ", "the file in ") +
			file + ", with its comment ([] or null where none applies; every field, in order: bonsai check --schema bonsai." + formatName + ")"}
	}
	return packFound{code, file, fmt.Sprintf("%s: field %s: %s", at(file, p.Line), p.Field, p.Msg), next}
}

// keyComments finds the keys of a YAML file the reader read whole (m) that carry no comment (pack-comment).
func keyComments(file string, raw []byte, m *reader.Map) []packFound {
	lines := textLines(raw)
	commentLine := func(no int) bool {
		return no >= 1 && no <= len(lines) && strings.HasPrefix(strings.TrimLeft(lines[no-1], " "), "#")
	}
	var out []packFound
	var walk func(v any, where string)
	walk = func(v any, where string) {
		switch x := v.(type) {
		case *reader.Map:
			for _, e := range x.Entries() {
				name := e.Key
				if where != "" {
					name = where + "." + e.Key
				}
				if e.Line >= 1 && e.Line <= len(lines) && !reader.KeyLineComment(lines[e.Line-1]) && !commentLine(e.Line-1) {
					out = append(out, packFound{"pack-comment", file, fmt.Sprintf("%s: the key %s has no comment, at the end of its line "+
						"or on the line just above it", at(file, e.Line), name),
						"add a # comment to that line, or a comment line just above it, saying what the key means (spec section 5: " +
							"every key in a pack's YAML files documents itself)"})
				}
				walk(e.Value, name)
			}
		case []any:
			for i, it := range x {
				walk(it, where+"["+strconv.Itoa(i)+"]")
			}
		}
	}
	walk(m, "")
	return out
}

// textLines splits a file into its lines, each without its line ending, a leading BOM taken off.
func textLines(raw []byte) []string {
	s := strings.TrimPrefix(string(raw), "\ufeff")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}

// packYAMLRead is what bonsai/pack.yaml gave the rules that follow it.
type packYAMLRead struct {
	id        string                // its id as written, "" when it has none
	manifest  *workspace.Pack       // as the engine reads it at a link; nil when it does not read so
	documents []format.PackDocument // the document kinds that fit the schema, for the template skills' statuses
}

var (
	denyWhyPointer = regexp.MustCompile(`^#/deny/[0-9]+/why$`)
	denyPointer    = regexp.MustCompile(`^#/deny/[0-9]+$`)
	itemPointer    = regexp.MustCompile(`^#/(documents|protected)/([0-9]+)`)
)

// packYAML checks bonsai/pack.yaml: pack-schema, pack-comment, pack-why, pack-documents, pack-protected, pack-bash and
// pack-runs.
func (c *packCheck) packYAML(raw []byte) packYAMLRead {
	const file = workspace.PackFile
	var out packYAMLRead
	m := c.readYAML(file, "pack", raw)
	if m == nil {
		return out
	}
	if v, ok := m.Get("id"); ok {
		out.id, _ = v.(string)
	}
	doc, _ := reader.JSON(m).(schema.Object)

	// The writer's problems, each to its rule: documents' and protected's items, a deny rule's why, the rest the schema's.
	var schemaF, docsF, protF []packFound
	badItem := map[string]bool{} // "documents/0": an item the schema's problems are about
	probs := format.MustLookup("pack").WriterProblems(m)
	for _, p := range probs {
		switch {
		case itemPointer.MatchString(p.Pointer):
			sm := itemPointer.FindStringSubmatch(p.Pointer)
			badItem[sm[1]+"/"+sm[2]] = true
			if sm[1] == "documents" {
				docsF = append(docsF, writerFound("pack-documents", file, "pack", p))
			} else {
				protF = append(protF, writerFound("pack-protected", file, "pack", p))
			}
		case denyWhyPointer.MatchString(p.Pointer) || denyPointer.MatchString(p.Pointer) && p.Msg == `required field "why" is missing`:
			// pack-why's, which reads the deny rules itself.
		default:
			schemaF = append(schemaF, writerFound("pack-schema", file, "pack", p))
		}
	}

	whyF := denyWhys(file, m)
	docsF, out.documents = append(docsF, documentsWellFormed(file, doc, badItem)...), documentsThatFit(doc, badItem)
	protF = append(protF, protectedWellFormed(file, doc, badItem)...)
	bashF := hookBash(file, m)
	runsF := hookRuns(file, m)

	c.put(schemaF...)
	if len(probs) == 0 && len(whyF)+len(docsF)+len(protF)+len(bashF) == 0 {
		manifest, err := workspace.ReadPack(raw)
		var we *workspace.Error
		switch {
		case errors.As(err, &we):
			c.put(packFound{"pack-schema", file, at(file, we.Line) + ": " + we.Msg + " (the engine refuses the pack at a link)", we.Next})
		case err != nil:
			c.put(packFound{"pack-schema", file, file + ": " + err.Error(), "fix the file as the message says, then run the command again"})
		default:
			out.manifest = manifest
			if e := c.sources(manifest); e != nil {
				c.put(*e)
			}
		}
	}
	c.put(keyComments(file, raw, m)...)
	c.put(whyF...)
	c.put(docsF...)
	c.put(protF...)
	c.put(bashF...)
	c.put(runsF...)
	return out
}

// sources finds the first file pack.yaml names that the pack does not hold: its block, then each files entry's source,
// as the engine reads them at a link (packAt).
func (c *packCheck) sources(p *workspace.Pack) *packFound {
	file := workspace.PackFile
	if p.Block != "" {
		if !c.isFile("bonsai/" + p.Block) {
			return &packFound{"pack-schema", file, fmt.Sprintf("%s: its block is bonsai/%s, which is not in the pack (the engine refuses the "+
				"pack at a link)", file, p.Block), "add bonsai/" + p.Block + " to the pack, or set block to null"}
		}
	}
	for i, fe := range p.Files {
		if !c.isFile("bonsai/files/" + fe.From) {
			return &packFound{"pack-schema", file, fmt.Sprintf("%s: files item %d comes from bonsai/files/%s, which is not in the pack "+
				"(the engine refuses the pack at a link)", at(file, fe.Line), i+1, fe.From),
				"add bonsai/files/" + fe.From + " to the pack, or fix files item " + strconv.Itoa(i+1) + "'s from"}
		}
	}
	return nil
}

func (c *packCheck) isFile(rel string) bool {
	st, err := os.Stat(filepath.Join(c.dir, filepath.FromSlash(rel)))
	return err == nil && st.Mode().IsRegular()
}

// mapItems gives a mapping's list field's items that are mappings, each with its index; none when the field is not a list.
func mapItems(m *reader.Map, key string) (out []struct {
	i int
	m *reader.Map
}) {
	v, _ := m.Get(key)
	list, _ := v.([]any)
	for i, it := range list {
		if im, ok := it.(*reader.Map); ok && im.Len() > 0 {
			out = append(out, struct {
				i int
				m *reader.Map
			}{i, im})
		}
	}
	return out
}

// denyWhys finds the deny rules with no why (pack-why).
func denyWhys(file string, m *reader.Map) []packFound {
	var out []packFound
	for _, it := range mapItems(m, "deny") {
		rule, _ := it.m.Get("rule")
		name := fmt.Sprintf("deny item %d", it.i+1)
		if r, ok := rule.(string); ok {
			name += " (" + strconv.QuoteToASCII(r) + ")"
		}
		e, ok := it.m.Entry("why")
		line, what := it.m.Entries()[0].Line, ""
		switch s, isText := e.Value.(string); {
		case !ok:
			what = "has no why"
		case e.Value == nil:
			line, what = e.Line, "has a null why"
		case !isText:
			line, what = e.Line, "has a why that is not text"
		case strings.TrimSpace(s) == "":
			line, what = e.Line, "has a blank why"
		default:
			continue
		}
		out = append(out, packFound{"pack-why", file, fmt.Sprintf("%s: %s %s: each deny rule says what it stops, in the sentence "+
			"update's preview prints (spec section 5)", at(file, line), name, what),
			"give " + name + " a why in " + file + ": one plain sentence, quoted, saying what the rule stops"})
	}
	return out
}

// documentsThatFit gives the document kinds whose items fit the schema (badItem names those that do not), bound to
// their Go type; none when documents is not a list.
func documentsThatFit(doc schema.Object, badItem map[string]bool) []format.PackDocument {
	v, _ := doc.Get("documents")
	list, _ := v.([]any)
	var fit []any
	for i, it := range list {
		if !badItem["documents/"+strconv.Itoa(i)] {
			fit = append(fit, it)
		}
	}
	var p format.Pack
	if err := format.MustLookup("pack").Bind(schema.Object{{Key: "documents", Value: fit}}, &p); err != nil {
		return nil
	}
	return p.Documents
}

// documentsWellFormed finds the document kinds that fit the schema but are not well formed (pack-documents): the
// engine's rules (documentsProblems) and check --pack's own two (statuses once; moves and stamps naming them).
func documentsWellFormed(file string, doc schema.Object, badItem map[string]bool) []packFound {
	v, _ := doc.Get("documents")
	list, _ := v.([]any)
	next := "fix that document kind in " + file + " (its fields: bonsai check --schema bonsai.pack; contract section 7.3)"
	var out []packFound
	var docs []format.PackDocument
	var index []int // each kind's item number in documents
	for i, it := range list {
		if badItem["documents/"+strconv.Itoa(i)] {
			continue
		}
		var p format.Pack
		if err := format.MustLookup("pack").Bind(schema.Object{{Key: "documents", Value: []any{it}}}, &p); err != nil || len(p.Documents) != 1 {
			continue
		}
		docs = append(docs, p.Documents[0])
		index = append(index, i)
	}
	for _, p := range documentsProblems(docs) {
		// documentsProblems numbers the kinds it was given; name each by its item in the file.
		text := strings.TrimPrefix(p.text, "pack.yaml's ")
		text = strings.Replace(text, fmt.Sprintf("documents item %d ", p.item+1), fmt.Sprintf("documents item %d ", index[p.item]+1), 1)
		out = append(out, packFound{"pack-documents", file, file + ": " + text, next})
	}
	for j, k := range docs {
		where := fmt.Sprintf("%s: documents item %d (%s)", file, index[j]+1, k.Kind)
		has := map[string]bool{}
		for _, s := range k.Statuses {
			if has[s] {
				out = append(out, packFound{"pack-documents", file, fmt.Sprintf("%s lists the status %s twice", where, quote(s)), next})
			}
			has[s] = true
		}
		for _, who := range []struct {
			name  string
			moves [][]string
		}{{"person", k.Person}, {"agent", k.Agent}} {
			for n, mv := range who.moves {
				if len(mv) != 2 {
					out = append(out, packFound{"pack-documents", file, fmt.Sprintf("%s: %s move %d is %d statuses, not a pair [from, to]",
						where, who.name, n+1, len(mv)), next})
					continue
				}
				for _, s := range mv {
					if !has[s] {
						out = append(out, packFound{"pack-documents", file, fmt.Sprintf("%s: %s move %d names the status %s, which the kind "+
							"does not have", where, who.name, n+1, quote(s)), next})
					}
				}
			}
		}
		for _, m := range k.Stamp {
			if !has[m.Key] {
				out = append(out, packFound{"pack-documents", file, fmt.Sprintf("%s: stamp names the status %s, which the kind does not have",
					where, quote(m.Key)), next})
			}
		}
	}
	return out
}

// protectedWellFormed finds the protected globs that fit the schema but are not well formed (pack-protected).
func protectedWellFormed(file string, doc schema.Object, badItem map[string]bool) []packFound {
	v, _ := doc.Get("protected")
	list, _ := v.([]any)
	next := "write that glob relative to the project with forward slashes, its segments as the guard reads them (* and ? in a name, " +
		"[...] a class, ** whole segments): bonsai check --schema bonsai.pack"
	var out []packFound
	for i, it := range list {
		g, ok := it.(string)
		if !ok || badItem["protected/"+strconv.Itoa(i)] {
			continue
		}
		if ps := protectedProblems([]string{g}); len(ps) > 0 {
			out = append(out, packFound{"pack-protected", file, file + ": " + strings.TrimPrefix(ps[0].text, "pack.yaml's "), next})
			continue
		}
		segs := strings.Split(strings.TrimSuffix(g, "/"), "/")
		for _, seg := range segs {
			what := ""
			switch {
			case seg == "":
				what = "an empty segment (// or a glob of / alone)"
			case seg == "." || seg == "..":
				what = "a " + seg + " segment, which no project path holds"
			case seg != "**":
				if _, err := path.Match(seg, ""); err != nil {
					what = "the segment " + quote(seg) + ", which path.Match does not read (" + err.Error() + ")"
				}
			}
			if what != "" {
				out = append(out, packFound{"pack-protected", file, fmt.Sprintf("%s: protected item %d, %s, has %s", file, i+1, quote(g), what), next})
				break
			}
		}
	}
	return out
}

// bashNext is the next step of a hook command that calls bash by name (format.PackFromMap's).
const bashNext = "run the script with sh (sh <path>), or name it alone if it runs on its own; on Windows the bash on the PATH can be WSL's launcher, not Git Bash"

// hookBash finds pack.yaml's hook commands that call bash by name (pack-bash).
func hookBash(file string, m *reader.Map) []packFound {
	var out []packFound
	for _, it := range mapItems(m, "hooks") {
		e, ok := it.m.Entry("command")
		cmd, isText := e.Value.(string)
		if ok && isText && format.CallsBash(cmd) {
			out = append(out, packFound{"pack-bash", file, fmt.Sprintf("%s: hooks item %d's command (%s) calls bash by name, which a hook "+
				"line never does (spec section 3)", at(file, e.Line), it.i+1, strconv.QuoteToASCII(cmd)), bashNext})
		}
	}
	return out
}

// hookRuns finds pack.yaml's hook commands that name a file the pack writes that their runs does not list (pack-runs).
func hookRuns(file string, m *reader.Map) []packFound {
	var paths []string
	for _, it := range mapItems(m, "files") {
		if p, ok := it.m.Get("path"); ok {
			if s, ok := p.(string); ok && s != "" {
				paths = append(paths, s)
			}
		}
	}
	sort.Strings(paths)
	var out []packFound
	for _, it := range mapItems(m, "hooks") {
		e, ok := it.m.Entry("command")
		cmd, isText := e.Value.(string)
		if !ok || !isText {
			continue
		}
		listed := map[string]bool{}
		if r, ok := it.m.Get("runs"); ok {
			list, _ := r.([]any)
			for _, x := range list {
				if s, ok := x.(string); ok {
					listed[s] = true
				}
			}
		}
		for _, p := range paths {
			if !listed[p] && namesPath(cmd, p) {
				out = append(out, packFound{"pack-runs", file, fmt.Sprintf("%s: hooks item %d's command (%s) names %s, a file the pack writes, "+
					"but its runs does not list it, so a change to that file would run unseen (step 5.1.1)", at(file, e.Line), it.i+1,
					strconv.QuoteToASCII(cmd), p), fmt.Sprintf("add %s to hooks item %d's runs in %s (runs lists every pack file the "+
					"command runs, by its path in the project)", strconv.QuoteToASCII(p), it.i+1, file)})
			}
		}
	}
	return out
}

// declarations checks bonsai/labels.yaml and bonsai/lanes.yaml when the pack has them: pack-schema and pack-comment.
// It gives each as Bonsai reads it when it has no pack-schema finding, nil when the pack has none or it has one (a
// list that does not hold is no closed list to hold the template skills' cells to, nor a block to count).
func (c *packCheck) declarations(pf packYAMLRead) (*format.Labels, *format.Lanes, *Error) {
	c.has = map[string]bool{}
	var labels *format.Labels
	var lanes *format.Lanes
	for _, d := range []struct{ file, name string }{{LabelsFile, "labels"}, {LanesFile, "lanes"}} {
		raw, exists, e := c.read(d.file)
		if e != nil {
			return nil, nil, e
		}
		if !exists {
			continue
		}
		c.has[d.file] = true
		m := c.readYAML(d.file, d.name, raw)
		if m == nil {
			continue
		}
		probs := format.MustLookup(d.name).WriterProblems(m)
		for _, p := range probs {
			c.put(writerFound("pack-schema", d.file, d.name, p))
		}
		// The rules a schema cannot say; a file with no problem at all gives the closed lists of pack-values.
		next := "fix " + d.file + " as the sentence says (contract section 5.1; its fields: bonsai check --schema bonsai." + d.name + ")"
		if d.name == "labels" {
			if l, err := format.ReadLabels(raw); err == nil {
				var ps []string
				if pf.id != "" {
					ps = labelsProblems(l, pf.id)
				}
				for _, p := range ps {
					c.put(packFound{"pack-schema", d.file, p, next})
				}
				if len(probs)+len(ps) == 0 {
					labels = l
				}
			}
		} else if l, err := format.ReadLanes(raw); err == nil {
			ps := lanesProblems(l)
			for _, p := range ps {
				c.put(packFound{"pack-schema", d.file, p, next})
			}
			if len(probs)+len(ps) == 0 {
				lanes = l
			}
		}
		c.put(keyComments(d.file, raw, m)...)
	}
	return labels, lanes, nil
}

// block checks the instruction block a project linked to this pack alone would get (pack-block).
func (c *packCheck) block(pf packYAMLRead, labels *format.Labels) *Error {
	if pf.manifest == nil {
		return nil
	}
	pd := &PackData{Manifest: pf.manifest, Declares: &format.Declares{Labels: labels}}
	file := workspace.PackFile
	if pf.manifest.Block != "" {
		file = "bonsai/" + pf.manifest.Block
		raw, exists, e := c.read(file)
		if e != nil {
			return e
		}
		if exists {
			pd.Block = blockText(raw)
		}
	}
	cfg := &workspace.Config{Name: "example", Full: &format.Workspace{Documents: DefaultDocuments}}
	n := strings.Count(blockBody(cfg, []*PackData{pd}), "\n")
	if n <= MaxBlockLines {
		return nil
	}
	own := 0
	if pd.Block != "" {
		own = strings.Count(pd.Block, "\n") + 1
	}
	labelLines := 0
	if labels != nil && len(labels.Labels) > 0 {
		labelLines = len(labels.Labels) + 1
	}
	c.put(packFound{"pack-block", file, fmt.Sprintf("%s: a project linked to this pack alone would get an instruction block of %d lines "+
		"in CLAUDE.md, over its fixed %d (spec section 6): its block.md gives %d of them, its label definitions %d, Bonsai's own lines "+
		"the rest", file, n, MaxBlockLines, own, labelLines),
		fmt.Sprintf("shorten bonsai/block.md, or the label definitions in %s, by %d lines at least", LabelsFile, n-MaxBlockLines)})
	return nil
}

// plugin checks the plugin's manifest (pack-plugin-version) and the plugin's own hooks (pack-bash).
func (c *packCheck) plugin() *Error {
	raw, exists, e := c.read(manifestPath)
	if e != nil {
		return e
	}
	var manifest schema.Object
	if exists {
		ver, has, readable := manifestVersion(raw)
		switch {
		case !readable:
			c.put(packFound{"pack-plugin-version", manifestPath, manifestPath + " is not a JSON object Bonsai reads, so whether it carries a " +
				"version cannot be told", "write " + manifestPath + " as a JSON object (claude plugin validate --json . says what is wrong)"})
		case has:
			c.put(packFound{"pack-plugin-version", manifestPath, fmt.Sprintf("%s carries the version %s, so Claude Code would keep a machine "+
				"on its cached copy whatever the commit (spec section 5: a pack is pinned by commit)", manifestPath, schema.Show(ver)),
				"take version out of " + manifestPath + ": Claude Code takes the commit as the plugin's version only when it has none"})
		}
		if readable {
			v, _ := schema.Decode(raw)
			manifest, _ = v.(schema.Object)
		}
	}
	hooksRaw, exists, e := c.read(pluginHooksFile)
	if e != nil {
		return e
	}
	if exists {
		if v, err := schema.Decode(hooksRaw); err == nil {
			c.put(commandsCallingBash(pluginHooksFile, v, "")...)
		}
	}
	if h, ok := manifest.Get("hooks"); ok {
		if _, isObject := h.(schema.Object); isObject {
			c.put(commandsCallingBash(manifestPath, h, "hooks")...)
		}
	}
	return nil
}

// commandsCallingBash finds every command string in a plugin's hooks (a JSON value) that calls bash by name.
func commandsCallingBash(file string, v any, where string) []packFound {
	var out []packFound
	switch x := v.(type) {
	case schema.Object:
		for _, m := range x {
			name := m.Key
			if where != "" {
				name = where + "." + m.Key
			}
			if s, ok := m.Value.(string); ok && m.Key == "command" && format.CallsBash(s) {
				out = append(out, packFound{"pack-bash", file, fmt.Sprintf("%s: %s (%s) calls bash by name, which a hook command never does "+
					"(spec section 3)", file, name, strconv.QuoteToASCII(s)), bashNext})
				continue
			}
			out = append(out, commandsCallingBash(file, m.Value, name)...)
		}
	case []any:
		for i, it := range x {
			out = append(out, commandsCallingBash(file, it, where+"["+strconv.Itoa(i)+"]")...)
		}
	}
	return out
}

// templates checks each template skill (pack-fields, pack-values) and gives their paths.
func (c *packCheck) templates(pf packYAMLRead, labels *format.Labels, lanes *format.Lanes) ([]string, *Error) {
	entries, err := os.ReadDir(filepath.Join(c.dir, skillsDir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, errorf("read-failed", ExitRuntime, "check the folder's permissions, then run the command again",
			"the pack's %s folder cannot be read: %v", skillsDir, err)
	}
	var found []string
	for _, ent := range entries { // os.ReadDir sorts them by name
		if !ent.IsDir() {
			continue
		}
		file := skillsDir + "/" + ent.Name() + "/SKILL.md"
		raw, exists, e := c.read(file)
		if e != nil {
			return nil, e
		}
		if !exists {
			continue
		}
		sk, ok := readTemplateSkill(textLines(raw))
		if !ok {
			continue
		}
		found = append(found, skillsDir+"/"+ent.Name())
		lists := closedLists{kind: ent.Name(), documents: pf.documents, labels: labels, lanes: lanes}
		c.put(sk.fields(file)...)
		c.put(sk.values(file, &lists)...)
	}
	return found, nil
}

// templateSkill is a template skill's fields table and its template, as read from its SKILL.md.
type templateSkill struct {
	tableLine int           // the fields table's header row
	rows      []fieldRow    // its rows, in order
	template  bool          // a fenced code block follows the table
	keys      []templateKey // the template's top-level fields, in order
	format    string        // its format: line's value, "" for none
}

// fieldRow is one row of a fields table: its field and its allowed-values cell.
type fieldRow struct {
	line    int
	field   string
	allowed string
}

// templateKey is one top-level field of a template, with its line in the SKILL.md.
type templateKey struct {
	name string
	line int
}

var (
	templateKeyLine = regexp.MustCompile(`^([a-z][a-z0-9_]*):( |$)`)
	tableSeparator  = regexp.MustCompile(`^:?-+:?$`)
	backquoted      = regexp.MustCompile("`([^`]*)`")
)

// fieldsHeader is the header row's cells of a fields table (spec §5), in order, in lower case.
var fieldsHeader = []string{"field", "meaning", "allowed values", "example"}

// cells splits a markdown table row into its cells, trimmed ("\|" is a | inside a cell).
func cells(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimSuffix(strings.TrimPrefix(row, "|"), "|")
	parts := strings.Split(strings.ReplaceAll(row, `\|`, "\x00"), "|")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(strings.ReplaceAll(p, "\x00", "|"))
	}
	return parts
}

// readTemplateSkill reads a SKILL.md's lines: ok false when its body holds no fields table (it is no template skill).
func readTemplateSkill(lines []string) (templateSkill, bool) {
	var sk templateSkill
	i := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" { // the skill's own frontmatter
		for i = 1; i < len(lines) && strings.TrimSpace(lines[i]) != "---"; i++ {
		}
		i++
	}
	for ; i+1 < len(lines); i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
			continue
		}
		head := cells(lines[i])
		if len(head) != len(fieldsHeader) {
			continue
		}
		match := true
		for j, h := range head {
			match = match && strings.EqualFold(h, fieldsHeader[j])
		}
		sep := true
		for _, s := range cells(lines[i+1]) {
			sep = sep && tableSeparator.MatchString(s)
		}
		if match && sep {
			break
		}
	}
	if i+1 >= len(lines) {
		return sk, false
	}
	sk.tableLine = i + 1
	for i += 2; i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|"); i++ {
		cs := cells(lines[i])
		field := strings.TrimSpace(strings.Trim(cs[0], "`"))
		if field == "" {
			continue
		}
		r := fieldRow{line: i + 1, field: field}
		if len(cs) > 2 {
			r.allowed = cs[2]
		}
		sk.rows = append(sk.rows, r)
	}
	// The template: the first fenced code block after the table.
	for ; i < len(lines); i++ {
		t := strings.TrimLeft(lines[i], " ")
		if len(lines[i])-len(t) > 3 || !(strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")) {
			continue
		}
		fence := t[:len(t)-len(strings.TrimLeft(t, t[:1]))]
		var body []int // the block's lines, by index
		for i++; i < len(lines); i++ {
			if u := strings.TrimSpace(lines[i]); strings.HasPrefix(u, fence) && strings.Trim(u, fence[:1]) == "" {
				break
			}
			body = append(body, i)
		}
		sk.template = true
		sk.keys, sk.format = templateKeys(lines, body)
		break
	}
	return sk, true
}

// templateKeys reads a template's top-level fields and its format: line's value from its lines (by index): a markdown
// template's frontmatter, or a YAML template whole.
func templateKeys(lines []string, body []int) ([]templateKey, string) {
	for len(body) > 0 && strings.TrimSpace(lines[body[0]]) == "" {
		body = body[1:]
	}
	if len(body) > 0 && strings.TrimRight(lines[body[0]], " ") == "---" {
		end := len(body)
		for j := 1; j < len(body); j++ {
			if strings.TrimRight(lines[body[j]], " ") == "---" {
				end = j
				break
			}
		}
		body = body[1:end]
	}
	var keys []templateKey
	formatValue := ""
	seen := map[string]bool{}
	for _, n := range body {
		m := templateKeyLine.FindStringSubmatch(lines[n])
		if m == nil || seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		keys = append(keys, templateKey{m[1], n + 1})
		if m[1] == "format" {
			v := strings.TrimSpace(strings.TrimPrefix(lines[n], "format:"))
			if j := strings.Index(v, " #"); j >= 0 {
				v = v[:j]
			}
			formatValue = strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return keys, formatValue
}

// rowHead is the top-level field a row names: its field up to the first . or [.
func rowHead(field string) string {
	if i := strings.IndexAny(field, ".["); i >= 0 {
		return field[:i]
	}
	return field
}

// bonsaiFormat gives the Bonsai format a template's format: line names, if it names one ("bonsai.task/1"): ok false
// when it names none, f nil when it names one this Bonsai does not have.
func (sk templateSkill) bonsaiFormat() (f *format.Format, ok bool) {
	if !strings.HasPrefix(sk.format, "bonsai.") {
		return nil, false
	}
	if g, found := format.Lookup(sk.format); found && g.Versioned() == sk.format && g.Major > 0 {
		return g, true
	}
	return nil, true
}

// fields finds where a template skill's fields table and its template differ (pack-fields).
func (sk templateSkill) fields(file string) []packFound {
	next := "make the fields table and the template hold the same fields: a row for each field of the template, and no row for a " +
		"field it does not hold"
	if !sk.template {
		return []packFound{{"pack-fields", file, fmt.Sprintf("%s: a fields table with no template after it (a fenced code block holding "+
			"the template)", at(file, sk.tableLine)), "put the template after its fields table, in a fenced code block"}}
	}
	var out []packFound
	rows := map[string]bool{}
	for _, r := range sk.rows {
		rows[r.field] = true
	}
	keys := map[string]bool{}
	var noRow []string
	for _, k := range sk.keys {
		keys[k.name] = true
		if !rows[k.name] {
			noRow = append(noRow, k.name)
		}
	}
	if len(noRow) > 0 {
		out = append(out, packFound{"pack-fields", file, fmt.Sprintf("%s: the template's %s %s no row in its fields table",
			at(file, sk.tableLine), plural(len(noRow), "field", "fields"), strings.Join(noRow, ", ")+plural(len(noRow), " has", " have")), next})
	}
	for _, r := range sk.rows {
		if !keys[rowHead(r.field)] {
			out = append(out, packFound{"pack-fields", file, fmt.Sprintf("%s: the fields table has a row for %s, a field the template does not hold",
				at(file, r.line), r.field), next})
		}
	}
	f, named := sk.bonsaiFormat()
	switch {
	case named && f == nil:
		out = append(out, packFound{"pack-fields", file, fmt.Sprintf("%s: the template's format is %s, which is not one of this Bonsai's "+
			"formats", file, sk.format), "write the format: line as one of Bonsai's formats, its name and major (bonsai check --schema " +
			"<format> prints each; an unknown name lists them)"})
	case named:
		var want, missing, extra []string
		for _, p := range sk.schemaFields(f) {
			want = append(want, p)
			if !keys[p] {
				missing = append(missing, p)
			}
		}
		for _, k := range sk.keys {
			if !contains(want, k.name) {
				extra = append(extra, k.name)
			}
		}
		if len(missing)+len(extra) > 0 {
			var parts []string
			if len(missing) > 0 {
				parts = append(parts, "lacks "+strings.Join(missing, ", "))
			}
			if len(extra) > 0 {
				parts = append(parts, "holds "+strings.Join(extra, ", ")+", which the format does not have")
			}
			out = append(out, packFound{"pack-fields", file, fmt.Sprintf("%s: the template is %s, whose fields are its schema's, but it %s",
				file, f.Versioned(), strings.Join(parts, " and ")), "give the template, and its fields table, the format's fields: " +
				"bonsai check --schema " + f.ID()})
		}
	}
	return out
}

// schemaFields are a format's top-level fields, in its schema's order.
func (templateSkill) schemaFields(f *format.Format) []string {
	ps, _ := f.Schema().Get("properties")
	po, _ := ps.(schema.Object)
	out := make([]string, 0, len(po))
	for _, m := range po {
		out = append(out, m.Key)
	}
	return out
}

// closedLists finds the closed list a template's field has: in its Bonsai format's schema, or in the pack's own
// declarations (a declared kind's statuses for the skill named after it, the pack's lanes, a choice label's values).
type closedLists struct {
	kind      string                // the template skill's folder name
	documents []format.PackDocument // the kinds pack.yaml declares
	labels    *format.Labels
	lanes     *format.Lanes
}

// list gives a row's field's closed list and where it is defined; nil when the field has none.
func (l *closedLists) list(f *format.Format, field string) ([]string, string) {
	if f != nil {
		if vals := schemaList(f.Schema(), field); len(vals) > 0 {
			return vals, f.ID() + "'s schema"
		}
	}
	if field == "status" {
		for _, d := range l.documents {
			if d.Kind == l.kind && len(d.Statuses) > 0 {
				return d.Statuses, "the statuses of the kind " + d.Kind + " in " + workspace.PackFile
			}
		}
	}
	if field == "lane" && l.lanes != nil && len(l.lanes.Lanes) > 0 {
		var names []string
		for _, ln := range l.lanes.Lanes {
			names = append(names, ln.Name)
		}
		return names, "the lanes in " + LanesFile
	}
	if l.labels != nil {
		name := strings.TrimPrefix(field, "labels.")
		for _, d := range l.labels.Labels {
			if d.Name == name && d.Kind == "choice" && len(d.Values) > 0 {
				return d.Values, "the label " + d.Name + "'s values in " + LabelsFile
			}
		}
	}
	return nil, ""
}

// schemaList is a field's closed list in a format's schema: its enum or const, or its items' enum; nil for none. A
// nested field is named by its path (documents.task, files[].kind).
func schemaList(s schema.Object, field string) []string {
	cur := s
	for _, seg := range strings.Split(field, ".") {
		seg = strings.TrimSuffix(seg, "[]")
		if items, ok := cur.Get("items"); ok {
			if io, ok := items.(schema.Object); ok {
				cur = io
			}
		}
		ps, _ := cur.Get("properties")
		po, _ := ps.(schema.Object)
		next, ok := po.Get(seg)
		if !ok {
			return nil
		}
		if cur, ok = next.(schema.Object); !ok {
			return nil
		}
	}
	vals := func(v any) []string {
		list, _ := v.([]any)
		var out []string
		for _, x := range list {
			if x == nil {
				continue
			}
			if s, ok := x.(string); ok {
				out = append(out, s)
			} else {
				out = append(out, schema.Show(x))
			}
		}
		return out
	}
	if e, ok := cur.Get("enum"); ok {
		return vals(e)
	}
	if c, ok := cur.Get("const"); ok {
		return vals([]any{c})
	}
	if items, ok := cur.Get("items"); ok {
		if io, ok := items.(schema.Object); ok {
			if e, ok := io.Get("enum"); ok {
				return vals(e)
			}
		}
	}
	return nil
}

// values finds the allowed-values cells that list other values than their field's closed list (pack-values).
func (sk templateSkill) values(file string, lists *closedLists) []packFound {
	f, _ := sk.bonsaiFormat()
	var out []packFound
	for _, r := range sk.rows {
		var cell []string
		for _, m := range backquoted.FindAllStringSubmatch(r.allowed, -1) {
			if v := m[1]; v != "" && v != "null" && !strings.ContainsAny(v, " \t") {
				cell = append(cell, v)
			}
		}
		if len(cell) == 0 {
			continue
		}
		want, where := lists.list(f, r.field)
		if len(want) == 0 || sameSet(cell, want) {
			continue
		}
		out = append(out, packFound{"pack-values", file, fmt.Sprintf("%s: the allowed values of %s list %s, not its closed list %s (from %s)",
			at(file, r.line), r.field, backquote(cell), backquote(want), where),
			fmt.Sprintf("write the allowed values of %s as its closed list, %s, or name where the list is defined instead, with no value in "+
				"backquotes", r.field, backquote(want))})
	}
	return out
}

func backquote(list []string) string {
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = "`" + s + "`"
	}
	return strings.Join(out, ", ")
}

// sameSet reports whether two lists hold the same values, order and repeats aside.
func sameSet(a, b []string) bool {
	in := func(x []string) map[string]bool {
		m := map[string]bool{}
		for _, s := range x {
			m[s] = true
		}
		return m
	}
	ma, mb := in(a), in(b)
	if len(ma) != len(mb) {
		return false
	}
	for s := range ma {
		if !mb[s] {
			return false
		}
	}
	return true
}
