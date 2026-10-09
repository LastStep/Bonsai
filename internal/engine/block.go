package engine

// The instruction block in the project's CLAUDE.md (file kind block, contract §14; spec §5, §6): a marked region
// Bonsai owns inside a file the project owns. Everything outside the markers is the project's and is kept byte for
// byte.
//
// The block (spec §6), at most MaxBlockLines lines from its start marker to its end marker:
//   - the start marker, which carries the one pointer to the docs (spec §5: "Its start marker carries the one pointer
//     line to the pack's block.md docs");
//   - one line naming the workspace and its packs;
//   - the import of the always-on protocol files: each pack file the packs write into bonsai.yaml's protocols folder
//     (documents.protocols), as @<path>, on one line (Claude Code imports each @path a CLAUDE.md names);
//   - the import of the project's memory index, @<documents.memory>/INDEX.md (spec §10), whether or not it exists yet:
//     the memory skill writes it, and the block need not change when it does;
//   - the label definitions agents see (contract §5.3: the packs' definitions are rendered into the instruction file;
//     those attached on this machine reach a session through its opening context instead): a line saying how to
//     read them, then one line per label, its name, kind, the documents it goes on, who sets it, whether it grants a
//     right, and its description (about 40 tokens each);
//   - each pack's block.md text (without its leading HTML comment, which documents it and stays in the pack), in
//     bonsai.yaml's order;
//   - the end marker.
//
// A block over MaxBlockLines is refused (bad-pack: the packs' makers shorten their block.md or labels): the cap is
// fixed (spec §6), and bonsai check finds a block over it in a project (step 5.1.6).
//
// The lock's fingerprint for CLAUDE.md is the SHA-256 of the region, from the start marker's line to the end
// marker's, each line ending in a line feed whatever the file uses.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// BlockFile is the project file that holds the instruction block.
const BlockFile = "CLAUDE.md"

const (
	blockStartPrefix = "<!-- bonsai:block start"
	blockEndPrefix   = "<!-- bonsai:block end"
	blockStart       = "<!-- bonsai:block start: written by bonsai init and update from bonsai.yaml and each pack's bonsai/block.md (its docs); edit those, not this block. -->"
	blockEnd         = "<!-- bonsai:block end -->"
)

// MaxBlockLines is the instruction block's cap, its markers included (spec §6).
const MaxBlockLines = 40

// blockBody is the block the engine writes for a workspace and its packs, LF line ends, its last line ending in a line
// feed.
func blockBody(cfg *workspace.Config, packs []*PackData) string {
	var names []string
	for _, p := range packs {
		names = append(names, p.Manifest.ID+" "+p.Manifest.Version)
	}
	head := "This project is linked to Bonsai as workspace " + cfg.Name + "."
	if len(names) > 0 {
		head = "This project is linked to Bonsai as workspace " + cfg.Name + "; its packs: " + strings.Join(names, ", ") + "."
	}
	lines := []string{blockStart, head}
	var docs format.Documents
	if cfg.Full != nil {
		docs = cfg.Full.Documents
	}
	if docs.Protocols != "" {
		var imports []string
		folder := strings.TrimSuffix(docs.Protocols, "/") + "/"
		for _, p := range packs {
			for _, f := range p.Manifest.Files {
				if strings.HasPrefix(f.Path, folder) {
					imports = append(imports, "@"+f.Path)
				}
			}
		}
		if len(imports) > 0 {
			lines = append(lines, "Always-on protocols, read in full: "+strings.Join(imports, " "))
		}
	}
	if docs.Memory != "" {
		lines = append(lines, "Project memory, an index of notes kept in "+docs.Memory+": @"+strings.TrimSuffix(docs.Memory, "/")+"/"+MemoryIndex)
	}
	var labels []string
	for _, p := range packs {
		if p.Declares != nil && p.Declares.Labels != nil {
			for _, l := range p.Declares.Labels.Labels {
				labels = append(labels, labelLine(l))
			}
		}
	}
	if len(labels) > 0 {
		lines = append(lines, "Labels a document may carry under its labels: field (name: the value it takes; the documents it goes on; who sets it):")
		lines = append(lines, labels...)
	}
	for _, p := range packs {
		if p.Block != "" {
			lines = append(lines, p.Block)
		}
	}
	lines = append(lines, blockEnd)
	return strings.Join(lines, "\n") + "\n"
}

// MemoryIndex is the memory index's file name in the memory folder (spec §10).
const MemoryIndex = "INDEX.md"

// labelLine is one label definition as agents see it in the block: its name, the kind of value it takes, the
// documents it goes on, who sets it (an outside label is never written by an agent), whether it grants a right, and
// its description.
func labelLine(l format.LabelDef) string {
	kind := "a number"
	switch l.Kind {
	case "choice":
		kind = "one of " + strings.Join(l.Values, ", ")
	case "list":
		items := "text"
		if l.Items != nil {
			items = *l.Items
		}
		kind = "a list of " + items
	case "text":
		kind = "text"
		if l.Pattern != nil {
			kind += " matching " + *l.Pattern
		}
		if l.Max != nil {
			kind += fmt.Sprintf(", at most %d characters", *l.Max)
		}
	}
	who := "agents set it"
	if l.SetBy == "outside" {
		who = "set outside agent sessions: never write it"
	}
	if l.Grants {
		who += "; it grants a right"
	}
	return "- " + l.Name + ": " + kind + "; on " + strings.Join(l.Kinds, ", ") + "; " + who + ". " + l.Description
}

// checkBlock refuses a block over MaxBlockLines.
func checkBlock(body string) error {
	n := strings.Count(body, "\n")
	if n <= MaxBlockLines {
		return nil
	}
	return errorf("bad-pack", ExitInput, "the packs' makers shorten their bonsai/block.md and labels; until then, take a pack out of "+
		"bonsai.yaml or keep its ref at a release whose block fits", "the instruction block in %s would be %d lines, over its fixed %d "+
		"(spec section 6): the packs' block.md texts and label definitions are too long together", BlockFile, n, MaxBlockLines).whose("person")
}

// blockDoc is CLAUDE.md as read, with the block's place in it.
type blockDoc struct {
	raw        []byte
	exists     bool
	found      bool // the file holds a block
	start, end int  // the block's bytes: raw[start:end], from its start line to the end of its end line
}

func readBlock(root string) (*blockDoc, error) {
	raw, err := os.ReadFile(filepath.Join(root, BlockFile))
	if errors.Is(err, fs.ErrNotExist) {
		return &blockDoc{}, nil
	}
	if err != nil {
		return nil, errorf("read-failed", ExitRuntime, "check the file's permissions, then run the command again", "%s cannot be read: %v", BlockFile, err)
	}
	d := &blockDoc{raw: raw, exists: true}
	starts, ends := 0, 0
	pos := 0
	for pos < len(raw) {
		nl := bytes.IndexByte(raw[pos:], '\n')
		lineEnd := len(raw)
		if nl >= 0 {
			lineEnd = pos + nl + 1
		}
		line := strings.TrimRight(string(raw[pos:lineEnd]), "\r\n")
		switch {
		case strings.HasPrefix(line, blockStartPrefix):
			starts++
			d.start = pos
		case strings.HasPrefix(line, blockEndPrefix):
			ends++
			d.end = lineEnd
		}
		pos = lineEnd
	}
	next := "fix the markers in CLAUDE.md (one <!-- bonsai:block start line, then one <!-- bonsai:block end line), or delete both and the lines between, then run the command again"
	switch {
	case starts == 0 && ends == 0:
		return d, nil
	case starts != 1 || ends != 1 || d.end <= d.start:
		return nil, errorf("bad-file", ExitInput, next, "%s has a broken Bonsai block (%d start and %d end markers)", BlockFile, starts, ends)
	}
	d.found = true
	return d, nil
}

// region is the block's text with LF line ends, its last line ending in a line feed.
func (d *blockDoc) region() string {
	s := strings.ReplaceAll(string(d.raw[d.start:d.end]), "\r\n", "\n")
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

// regionHash is the lock's fingerprint for the block.
func regionHash(region string) string {
	sum := sha256.Sum256([]byte(region))
	return hex.EncodeToString(sum[:])
}

// withBlock gives CLAUDE.md with body as its block: in place of the old block, or after the file's text with a
// blank line between. The file's own line ends (CRLF when it holds any) are used for the block.
func (d *blockDoc) withBlock(body string) []byte {
	eol := "\n"
	if bytes.Contains(d.raw, []byte("\r\n")) {
		eol = "\r\n"
	}
	b := []byte(strings.ReplaceAll(body, "\n", eol))
	if d.found {
		out := append([]byte{}, d.raw[:d.start]...)
		out = append(out, b...)
		return append(out, d.raw[d.end:]...)
	}
	out := append([]byte{}, d.raw...)
	if len(out) > 0 {
		if !bytes.HasSuffix(out, []byte("\n")) {
			out = append(out, eol...)
		}
		if !bytes.HasSuffix(out, []byte(eol+eol)) && !bytes.Equal(out, []byte(eol)) {
			out = append(out, eol...)
		}
	}
	return append(out, b...)
}
