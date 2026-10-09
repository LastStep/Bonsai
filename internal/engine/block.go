package engine

// The instruction block in the project's CLAUDE.md (file kind block, contract §14; spec §5, §6): a marked region
// Bonsai owns inside a file the project owns. Everything outside the markers is the project's and is kept byte for
// byte.
//
// The block holds one line naming the workspace and its packs, then each pack's block.md text (without its leading
// HTML comment, which documents it and stays in the pack), in bonsai.yaml's order. The start marker carries the one
// pointer to the docs (spec §5: "Its start marker carries the one pointer line to the pack's block.md docs"). Spec
// §6's other parts (the always-on protocol imports, the memory index, the label definitions) and the 40-line cap's
// check come with step 5.1 and the base pack (5.5).
//
// The lock's fingerprint for CLAUDE.md is the SHA-256 of the region, from the start marker's line to the end
// marker's, each line ending in a line feed whatever the file uses.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// BlockFile is the project file that holds the instruction block.
const BlockFile = "CLAUDE.md"

const (
	blockStartPrefix = "<!-- bonsai:block start"
	blockEndPrefix   = "<!-- bonsai:block end"
	blockStart       = "<!-- bonsai:block start: written by bonsai init and update from bonsai.yaml and each pack's bonsai/block.md (its docs); edit those, not this block. -->"
	blockEnd         = "<!-- bonsai:block end -->"
)

// blockBody is the block the engine writes, LF line ends, its last line ending in a line feed.
func blockBody(name string, packs []*PackData) string {
	var names []string
	for _, p := range packs {
		names = append(names, p.Manifest.ID+" "+p.Manifest.Version)
	}
	head := "This project is linked to Bonsai as workspace " + name + "."
	if len(names) > 0 {
		head = "This project is linked to Bonsai as workspace " + name + "; its packs: " + strings.Join(names, ", ") + "."
	}
	lines := []string{blockStart, head}
	for _, p := range packs {
		if p.Block != "" {
			lines = append(lines, p.Block)
		}
	}
	lines = append(lines, blockEnd)
	return strings.Join(lines, "\n") + "\n"
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
