package format

// Set 13's command output (T-6009; formats/schemas/packs.schema.json): bonsai pack outdated --json and bonsai pack
// show --json (bonsai.packs/1), each locked pack with its source, ref, commit, version, needs, what it brings and its
// newer release from the home's cache. Printed, never stored. This file holds its type, reader and writer;
// cmd/bonsai's pack word fills it.

import "github.com/LastStep/Bonsai/internal/schema"

// Packs is the document of bonsai pack outdated and show with --json. On a refusal Packs is [] and Error is set.
type Packs struct {
	Command string        `json:"command"` // the closed list: outdated, show
	Packs   []PacksPack   `json:"packs"`
	Error   *ErrorObject  `json:"error"`
	Extra   schema.Object `json:"-"`
}

// PacksPack is one locked pack.
type PacksPack struct {
	ID        string        `json:"id"`
	Source    string        `json:"source"`
	Path      *string       `json:"path"` // null for the repository's top
	Ref       *string       `json:"ref"`  // bonsai.yaml's, null when it no longer names the pack
	Commit    string        `json:"commit"`
	Version   string        `json:"version"`
	Needs     *PacksNeeds   `json:"needs"`  // null when the pack needs nothing
	Skills    []string      `json:"skills"` // from the home's clone at the locked commit; nil (null) when it is not there
	Roles     []string      `json:"roles"`
	Deny      []string      `json:"deny"`
	Protected []string      `json:"protected"`
	Kinds     []string      `json:"kinds"`
	Newer     *StatusNewer  `json:"newer"` // the home's cache's record, or null
	Why       *string       `json:"why"`   // why outdated's live read failed, or null
	Extra     schema.Object `json:"-"`
}

// PacksNeeds is what a pack needs.
type PacksNeeds struct {
	ClaudeCode *string       `json:"claude_code"`
	Extra      schema.Object `json:"-"`
}

// ReadPacks reads bonsai.packs/1's document.
func ReadPacks(raw []byte) (*Packs, error) {
	p := &Packs{}
	if err := MustLookup("packs").readJSONInto(raw, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Encode writes the document, held to bonsai.packs/1.
func (p *Packs) Encode() ([]byte, error) { return MustLookup("packs").Encode(p) }
