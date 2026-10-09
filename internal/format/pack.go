package format

// A pack's manifest, bonsai/pack.yaml (bonsai.pack/1, spec §5): what Bonsai's engine reads from a pack. A pack
// writes it; Bonsai reads it in full (the engine, through workspace.ReadPack; check --pack, step 5.1.9) and writes
// none. Beside its schema, one rule a schema cannot say: a hook command never calls bash by name (spec §3).

import (
	"strings"
	"unicode"

	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Pack is a pack.yaml, every field.
type Pack struct {
	ID        string         `json:"id"`
	Version   string         `json:"version"`
	Needs     PackNeeds      `json:"needs"`
	Block     *string        `json:"block"` // the pack's part of the instruction block, a file in bonsai/, or null
	Files     []PackFile     `json:"files"`
	Hooks     []PackHook     `json:"hooks"`
	Deny      []PackDeny     `json:"deny"`
	Documents []PackDocument `json:"documents"`
	Protected []string       `json:"protected"`
	Extra     schema.Object  `json:"-"`
}

// PackNeeds is what a machine needs for the pack.
type PackNeeds struct {
	ClaudeCode *string       `json:"claude_code"` // the oldest Claude Code it works with, or null
	Extra      schema.Object `json:"-"`
}

// PackFile is one file the engine writes into a project.
type PackFile struct {
	Path  string        `json:"path"`
	From  string        `json:"from"`
	Kind  string        `json:"kind"` // the closed list: pack, once
	Extra schema.Object `json:"-"`
}

// PackHook is one hook line.
type PackHook struct {
	Event   string        `json:"event"`
	Matcher *string       `json:"matcher"`
	Command string        `json:"command"` // never calls bash by name
	Runs    []string      `json:"runs"`    // the pack files the command runs, by their path in the project
	Why     string        `json:"why"`
	Extra   schema.Object `json:"-"`
}

// PackDeny is one deny rule.
type PackDeny struct {
	Rule  string        `json:"rule"`
	Why   string        `json:"why"`
	Extra schema.Object `json:"-"`
}

// PackDocument is one document kind the pack declares (contract §7.3).
type PackDocument struct {
	Kind      string        `json:"kind"`
	Path      *string       `json:"path"`
	File      *string       `json:"file"`
	ID        *string       `json:"id"`
	Statuses  []string      `json:"statuses"`
	Person    [][]string    `json:"person"`
	Agent     [][]string    `json:"agent"`
	Stamp     schema.Object `json:"stamp"`
	TaskField *string       `json:"task_field"`
	Extra     schema.Object `json:"-"`
}

// ReadPack reads a pack.yaml's bytes in full.
func ReadPack(raw []byte) (*Pack, error) {
	f := MustLookup("pack")
	_, m, err := f.ReadYAML(raw)
	if err != nil {
		if isFormat0(err) {
			err.(*ReadError).Field = ""
		}
		return nil, err
	}
	return PackFromMap(m)
}

// PackFromMap reads a pack.yaml in full from its mapping as internal/reader read it (workspace.ReadPack keeps it):
// its schema, then the rule that a hook command never calls bash by name.
func PackFromMap(m *reader.Map) (*Pack, error) {
	f := MustLookup("pack")
	doc, err := f.HoldMap(m)
	if err != nil {
		return nil, err
	}
	p := &Pack{}
	if err := f.Bind(doc, p); err != nil {
		return nil, err
	}
	for i, h := range p.Hooks {
		if CallsBash(h.Command) {
			field := "hooks[" + itoa(i) + "].command"
			return nil, &ReadError{Format: f.Versioned(), Field: field, Line: lineOf(m, []string{"hooks", itoa(i), "command"}),
				Msg: "field " + field + " calls bash by name (" + quoteASCII(h.Command) + "), which a hook line never does " +
					"(spec section 3: on Windows the bash on the PATH can be WSL's launcher, not Git Bash)",
				Next: "run the script with sh (sh <path>), or name it alone if it runs on its own; then release the pack again"}
		}
	}
	return p, nil
}

// CallsBash reports whether a hook command calls bash by name (spec §3: "a hook line never calls bash by name"): a
// word of the command, its quotes taken off, that is bash or bash.exe, or a path whose last part is, in any letter
// case. Words are split at spaces and at the shell's own punctuation, so bash after ;, &&, |, $( or a backquote is
// found, and so is bash given to another program (env bash, xargs bash, sh -c "bash x"). A word bash that would
// only be an argument (echo bash) is refused too: a hook line can always say it another way, and the rule is kept
// simple enough to hold.
func CallsBash(command string) bool {
	words := strings.FieldsFunc(command, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(";&|()`<>{}$=", r)
	})
	for _, w := range words {
		w = strings.ToLower(strings.Trim(w, `"'`))
		if i := strings.LastIndexAny(w, `/\`); i >= 0 {
			w = w[i+1:]
		}
		if w == "bash" || w == "bash.exe" {
			return true
		}
	}
	return false
}
