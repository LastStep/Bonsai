package engine

// A plan's fingerprint (T-6009's plan, step 7b; bonsai.changes/1's fingerprint, set 13): the SHA-256 of the plan's
// own fields only, written canonically, so a person or the studio can tie a --yes to the preview they read:
// --yes --expect <fingerprint> writes nothing (plan-changed) when the plan built now has another.
//
// What it covers, in this order, each list in the plan's order and every object's keys in a fixed order (no map
// iteration anywhere, so it is byte-stable and the same on Windows): the command and the word (name, sub), so a
// pack preview's fingerprint is refused by update and the reverse; the workspace (id, name, root); each planned file
// (path, result, and the SHA-256 of its bytes on disk and of the bytes the plan writes, line endings read as LF;
// "" where there are none); each settings line (file, change, kind, line, was); each pack (id, the ref bonsai.yaml
// names, the commit before and after); config; loosening; pack_changes. What it leaves out, as outcome or flag
// fields that differ between a preview and its --yes run: result, error, allow_exec, plugins, memory, cache,
// commit_lines, any read time, and the fingerprint itself.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// lfHash is the SHA-256 of b with CRLF read as LF, "" for nil.
func lfHash(b []byte) string {
	if b == nil {
		return ""
	}
	sum := sha256.Sum256(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))
	return hex.EncodeToString(sum[:])
}

// Fingerprint is the plan's fingerprint (this file's comment), 64 lower-case hex characters.
func (p *Plan) Fingerprint() string {
	c := p.changesCore()
	files := []any{}
	for _, f := range p.Files {
		after := f.write
		switch {
		case f.remove:
			after = nil
		case after == nil:
			after = f.old
		}
		files = append(files, schema.Object{{Key: "path", Value: f.Path}, {Key: "result", Value: f.Result},
			{Key: "old", Value: lfHash(f.old)}, {Key: "new", Value: lfHash(after)}})
	}
	settings := []any{}
	for _, s := range c.Settings {
		settings = append(settings, schema.Object{{Key: "file", Value: s.File}, {Key: "change", Value: s.Change},
			{Key: "kind", Value: s.Kind}, {Key: "line", Value: s.Line}, {Key: "was", Value: strOrNil(s.Was)}})
	}
	refs := map[string]string{}
	if p.Config != nil {
		for _, r := range p.Config.Packs {
			refs[r.ID] = r.Ref
		}
	}
	packs := []any{}
	for _, m := range p.Packs {
		packs = append(packs, schema.Object{{Key: "id", Value: m.ID}, {Key: "ref", Value: refs[m.ID]},
			{Key: "from", Value: m.From}, {Key: "to", Value: m.To}})
	}
	var ws any
	if c.Workspace != nil {
		ws = schema.Object{{Key: "id", Value: c.Workspace.ID}, {Key: "name", Value: c.Workspace.Name}, {Key: "root", Value: c.Workspace.Root}}
	}
	var config any
	if c.Config != nil {
		config = schema.Object{{Key: "path", Value: c.Config.Path}, {Key: "line", Value: c.Config.Line},
			{Key: "before", Value: strOrNil(c.Config.Before)}, {Key: "after", Value: strOrNil(c.Config.After)},
			{Key: "comment", Value: strOrNil(c.Config.Comment)}}
	}
	loose := []any{}
	for _, l := range c.Loosening {
		loose = append(loose, schema.Object{{Key: "kind", Value: l.Kind}, {Key: "value", Value: l.Value}, {Key: "pack", Value: l.Pack}})
	}
	changes := []any{}
	for _, pc := range c.PackChanges {
		titles := make([]any, len(pc.Titles))
		for i, t := range pc.Titles {
			titles[i] = t
		}
		changes = append(changes, schema.Object{{Key: "pack", Value: pc.Pack}, {Key: "titles", Value: titles},
			{Key: "more", Value: pc.More}, {Key: "why", Value: strOrNil(pc.Why)}})
	}
	doc := schema.Object{
		{Key: "fingerprint", Value: "bonsai.changes/1 plan, T-6009"},
		{Key: "command", Value: c.Command},
		{Key: "word", Value: schema.Object{{Key: "name", Value: c.Word.Name}, {Key: "sub", Value: strOrNil(c.Word.Sub)}}},
		{Key: "workspace", Value: ws},
		{Key: "files", Value: files},
		{Key: "settings", Value: settings},
		{Key: "packs", Value: packs},
		{Key: "config", Value: config},
		{Key: "loosening", Value: loose},
		{Key: "pack_changes", Value: changes},
	}
	sum := sha256.Sum256([]byte(schema.Show(doc)))
	return hex.EncodeToString(sum[:])
}

func strOrNil(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

// The word fields a plan carries for bonsai.changes/1's word: set by bonsai pack; empty for init, update and unlink,
// whose word is their command.
func (p *Plan) word() format.ChangesWord {
	if p.WordName == "" {
		return format.ChangesWord{Name: p.Command}
	}
	w := format.ChangesWord{Name: p.WordName}
	if p.WordSub != "" {
		sub := p.WordSub
		w.Sub = &sub
	}
	return w
}
