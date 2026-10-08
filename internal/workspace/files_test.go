package workspace

// Tests of bonsai.yaml and pack.yaml: the test pack's pack.yaml at commit A reads (plan part 2), a full bonsai.yaml
// reads, and each check refuses what it should with its file, line and next step.

import (
	"crypto/sha1" //nolint:gosec // git's blob hash, to prove the fixture is the source commit's bytes
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// gitBlob is git's object id for a file's bytes (git hash-object).
func gitBlob(raw []byte) string {
	h := sha1.New() //nolint:gosec // see above
	_, _ = fmt.Fprintf(h, "blob %d\x00", len(raw))
	_, _ = h.Write(raw)
	return hex.EncodeToString(h.Sum(nil))
}

func TestReadPackTestPackA(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "test-pack-a", "pack.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// LastStep/bonsai-test-pack at 506205354b7589f82f849820987aad17dba3309d, bonsai/pack.yaml (testdata/README.md).
	if got := gitBlob(raw); got != "e679d2d5396a4cbbd69f4898ed22dab3b744f90c" {
		t.Fatalf("testdata/test-pack-a/pack.yaml is git blob %s, not the test pack's at commit A", got)
	}
	p, err := ReadPack(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := &Pack{
		ID: "test-pack", Version: "0.1.0", ClaudeCode: "2.1.0", Block: "block.md",
		Files: []FileEntry{
			{Path: "test-pack/guide.md", From: "guide.md", Kind: "pack", Line: 39},
			{Path: "test-pack/start.md", From: "start.md", Kind: "once", Line: 42},
		},
		Hooks: []HookEntry{{Event: "SessionStart", Matcher: "startup", Command: "echo test-pack hook A",
			Why: "Prints which test-pack hook line is in place when a session starts; it blocks nothing.", Line: 46}},
		Deny: []DenyEntry{{Rule: "Edit(test-pack/never.txt)",
			Why: "Agents cannot edit test-pack/never.txt (the test pack's example deny rule).", Line: 51}},
	}
	p.Doc = nil
	if got, w := fmt.Sprintf("%+v", *p), fmt.Sprintf("%+v", *want); got != w {
		t.Errorf("read\n  %s\nwant\n  %s", got, w)
	}
	// Its CRLF form reads the same: a Windows checkout without .gitattributes would give it.
	crlf := strings.ReplaceAll(string(raw), "\n", "\r\n")
	p2, err := ReadPack([]byte(crlf))
	if err != nil {
		t.Fatal(err)
	}
	p2.Doc = nil
	if fmt.Sprintf("%+v", *p2) != fmt.Sprintf("%+v", *want) {
		t.Errorf("the CRLF form reads differently: %+v", *p2)
	}
}

const packHead = "format: bonsai.pack/1\nid: p\nversion: \"1.0.0\"\n"

func TestReadPackRefuses(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"format 0", "id: p\n", "no format: line first"},
		{"another format", "format: bonsai.lanes/1\n", `format is "bonsai.lanes/1", not bonsai.pack/1`},
		{"a newer major", "format: bonsai.pack/2\n", "line 1: format too new"},
		{"a reader refusal", packHead + "Block: x\n", "line 4: the key \"Block\" is neither"},
		{"no id", "format: bonsai.pack/1\nversion: \"1\"\n", "pack.yaml has no id"},
		{"a bad id", "format: bonsai.pack/1\nid: My-Pack\nversion: \"1\"\n", "is not a pack id"},
		{"no version", "format: bonsai.pack/1\nid: p\n", "pack.yaml has no version"},
		{"a number version", "format: bonsai.pack/1\nid: p\nversion: 1\n", "version is a number, not text"},
		{"needs not a mapping", packHead + "needs: [a]\n", "needs is a list, not a mapping"},
		{"claude_code a number", packHead + "needs:\n  claude_code: 2\n", "claude_code is a number"},
		{"block outside", packHead + "block: \"../x.md\"\n", "has a . or .. segment"},
		{"files not a list", packHead + "files: x\n", "files is text, not a list"},
		{"a file not a mapping", packHead + "files:\n  - x\n", "files item 1 is text, not a mapping"},
		{"a file with no from", packHead + "files:\n  - path: a.md\n    kind: pack\n", "files item 1 has no from"},
		{"a kept file", packHead + "files:\n  - path: a.md\n    from: a.md\n    kind: kept\n", `kind "kept" is not pack or once`},
		{"a block kind", packHead + "files:\n  - path: a.md\n    from: a.md\n    kind: block\n", "is not pack or once"},
		{"a path out", packHead + "files:\n  - path: \"../a.md\"\n    from: a.md\n    kind: pack\n", "has a . or .. segment"},
		{"a .git path", packHead + "files:\n  - path: \".git/hooks/pre-commit\"\n    from: a\n    kind: pack\n", "is inside a .git folder"},
		{"bonsai.yaml", packHead + "files:\n  - path: Bonsai.yaml\n    from: a\n    kind: pack\n", "is Bonsai's own file"},
		{".bonsai", packHead + "files:\n  - path: \".bonsai/lock.json\"\n    from: a\n    kind: pack\n", "is Bonsai's own file"},
		{"an absolute from", packHead + "files:\n  - path: a\n    from: \"/etc/x\"\n    kind: pack\n", "from: the path \"/etc/x\" starts with /"},
		{"twice by case", packHead + "files:\n  - path: A.md\n    from: a\n    kind: pack\n  - path: a.md\n    from: a\n    kind: once\n", "listed twice"},
		{"a hook with no why", packHead + "hooks:\n  - event: SessionStart\n    command: \"echo\"\n", "hooks item 1 has no why"},
		{"a hook with no command", packHead + "hooks:\n  - event: SessionStart\n    why: \"x\"\n", "hooks item 1 has no command"},
		{"a deny with no why", packHead + "deny:\n  - rule: \"Edit(x)\"\n", "deny item 1 has no why"},
		{"a deny rule a list", packHead + "deny:\n  - rule: [a]\n    why: x\n", "deny item 1's rule is a list"},
		{"an empty mapping item", packHead + "deny:\n  - \n", "deny item 1 is null"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ReadPack([]byte(c.in))
			if err == nil {
				t.Fatalf("read with no error")
			}
			msg := err.Error()
			if !strings.Contains(msg, c.want) || !strings.HasPrefix(msg, "bonsai/pack.yaml") || !strings.Contains(msg, "; next: ") {
				t.Errorf("error %q, want it to name bonsai/pack.yaml, hold %q and name a next step", msg, c.want)
			}
		})
	}
}

// PackFileKinds is a subset of the lock schema's file kinds, their one home.
func TestPackFileKindsAreLockKinds(t *testing.T) {
	enum := schemaAt(t, LockSchema(), "properties", "files", "additionalProperties", "properties", "kind", "enum")
	for _, k := range PackFileKinds {
		if !strings.Contains(enum, `"`+k+`"`) {
			t.Errorf("pack file kind %s is not in the lock schema's kinds %s", k, enum)
		}
	}
}

// A bonsai.yaml as spec §6 shows it, with a comment on every line, the id made valid and its lists shortened.
const fullConfig = `# bonsai.yaml: this project's Bonsai settings (format bonsai.workspace/1). Only a person changes this file.
# Every field and its allowed values: bonsai check --schema bonsai.workspace
format: bonsai.workspace/1          # the format and its version; a newer one is refused, never guessed
id: ws-7kq2m4xw5r3t6y2u7p4a5c3e2b   # this project's id, written once by bonsai init; a copy gets its own (init --new-id)
name: example                       # a short name: lower-case letters, digits and dashes
packs:                              # the packs this project uses, applied in this order
  - id: base                        # Bonsai's own pack: core labels, templates, the walls round secret files
    source: "https://example.com/packs/base.git"    # the git repo the pack comes from
    path: packs/base                # the pack's folder inside that repo
    ref: base-v1.0.0                # the release tag; change it and run bonsai update to take a new release
  - id: workflow                    # your roles, lanes, protocols and templates
    source: "https://example.com/packs/workflow.git"   # its own public repo
    ref: v1.0.0                     # its release tag
documents:                          # where this project keeps the documents Bonsai knows
  task: work/tasks                  # one file per task
# protected: paths an agent changes only while its running task lists them in bonsai.allows
protected: [".claude/**", "bonsai.yaml", ".bonsai/lock.json", ".github/**"]
# person_only: of those, the paths only a person grants (an approve or grant tap)
person_only: [".claude/**", "bonsai.yaml"]
never_edit: ["work/ledger.json"]    # paths no agent ever changes; written as deny rules
ladder_floor: [0, 1, 2, 3, 4]       # the rungs every task climbs, whatever the task lists
ratchets: {}                        # counts that may only rise; they rise when a person taps Bless
generated:                          # how long generated files are kept, per kind (skill base:generated-files)
  log:                              # the log, in .bonsai/local/log/
    keep_days: 30                   # a log file goes 30 days after its last line, never before its rows are in
`

func TestReadConfig(t *testing.T) {
	for _, in := range []string{fullConfig, strings.ReplaceAll(fullConfig, "\n", "\r\n"), "\xEF\xBB\xBF" + fullConfig} {
		c, err := ReadConfig([]byte(in))
		if err != nil {
			t.Fatal(err)
		}
		got := fmt.Sprintf("%s %s %+v %q %q", c.ID, c.Name, c.Packs, c.Protected, c.PersonOnly)
		want := `ws-7kq2m4xw5r3t6y2u7p4a5c3e2b example [{ID:base Source:https://example.com/packs/base.git Path:packs/base Ref:base-v1.0.0 Line:7} {ID:workflow Source:https://example.com/packs/workflow.git Path: Ref:v1.0.0 Line:11}] [".claude/**" "bonsai.yaml" ".bonsai/lock.json" ".github/**"] [".claude/**" "bonsai.yaml"]`
		if got != want {
			t.Errorf("read\n  %s\nwant\n  %s", got, want)
		}
		// Every key is kept, the step-5.1 ones too, in file order.
		var keys []string
		for _, e := range c.Doc.Entries() {
			keys = append(keys, e.Key)
		}
		if strings.Join(keys, " ") != "format id name packs documents protected person_only never_edit ladder_floor ratchets generated" {
			t.Errorf("keys %v", keys)
		}
	}
	// A minimal one: lists missing read as empty.
	c, err := ReadConfig([]byte("format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: x\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Packs == nil || c.Protected == nil || c.PersonOnly == nil || len(c.Packs)+len(c.Protected)+len(c.PersonOnly) != 0 {
		t.Errorf("missing lists read as %#v %#v %#v, want empty", c.Packs, c.Protected, c.PersonOnly)
	}
}

const cfgHead = "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: x\n"

func TestReadConfigRefuses(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"format 0", "id: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\n", "bonsai.yaml line 1: no format: line first"},
		{"a newer major", "format: bonsai.workspace/2\nid: [\n", "bonsai.yaml line 1: format too new"},
		{"another format", "format: bonsai.pack/1\n", `format is "bonsai.pack/1", not bonsai.workspace/1`},
		{"a reader refusal", cfgHead + "packs:\n- id: a\n", "bonsai.yaml line 5: a sequence item"},
		{"no id", "format: bonsai.workspace/1\nname: x\n", "bonsai.yaml has no id"},
		{"an id too short", "format: bonsai.workspace/1\nid: ws-abc\nname: x\n", "line 2: the id \"ws-abc\" is not ws-"},
		{"an id with capitals", "format: bonsai.workspace/1\nid: ws-AAAAAAAAAAAAAAAAAAAAAAAAAA\nname: x\n", "is not ws-"},
		{"no name", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\n", "has no name"},
		{"a bad name", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: X_y\n", "is not a slug"},
		{"a name a number", "format: bonsai.workspace/1\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: 7\n", "name is a number, not text"},
		{"packs a mapping", cfgHead + "packs:\n  a: 1\n", "packs is a mapping, not a list"},
		{"a pack text", cfgHead + "packs: [a]\n", "packs item 1 is text, not a mapping"},
		{"a pack with no ref", cfgHead + "packs:\n  - id: a\n    source: s\n", "packs item 1 has no ref"},
		{"a pack with no source", cfgHead + "packs:\n  - id: a\n    ref: r\n", "packs item 1 has no source"},
		{"a pack id with capitals", cfgHead + "packs:\n  - id: Ab\n    source: s\n    ref: r\n", "is not a pack id"},
		{"a pack twice", cfgHead + "packs:\n  - id: a\n    source: s\n    ref: r\n  - id: a\n    source: t\n    ref: r\n", "line 8: the pack \"a\" is listed twice"},
		{"a pack path out", cfgHead + "packs:\n  - id: a\n    source: s\n    path: a/../..\n    ref: r\n", "line 7: packs item 1's path"},
		{"an absolute glob", cfgHead + "protected: [\"/etc/**\"]\n", "line 4: the protected glob \"/etc/**\" is not project-relative"},
		{"a backslash glob", cfgHead + "person_only: ['a\\b']\n", "the person_only glob"},
		{"a drive glob", cfgHead + "protected: [\"C:/x\"]\n", "is not project-relative"},
		{"a glob a number", cfgHead + "protected: [1]\n", "protected item 1 is a number, not text"},
		{"person_only text", cfgHead + "person_only: x\n", "person_only is text, not a list"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ReadConfig([]byte(c.in))
			if err == nil {
				t.Fatalf("read with no error")
			}
			msg := err.Error()
			if !strings.Contains(msg, c.want) || !strings.Contains(msg, "; next: ") {
				t.Errorf("error %q, want it to hold %q and name a next step", msg, c.want)
			}
			for i := 0; i < len(msg); i++ {
				if msg[i] < 0x20 || msg[i] > 0x7e {
					t.Fatalf("not ASCII: %q", msg)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadConfig(dir)
	if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "next: run bonsai init") {
		t.Fatalf("a missing bonsai.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bonsai.yaml"), []byte(fullConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(dir)
	if err != nil || c.Name != "example" {
		t.Fatalf("LoadConfig: %v %v", c, err)
	}
	var e *Error
	if !errors.As(err, &e) && err != nil {
		t.Errorf("not an *Error")
	}
}

// The id and name patterns come from the status schema, their one home.
func TestIDAndNamePatternsAreTheSchemas(t *testing.T) {
	id, name := idName()
	if id.String() != `^ws-[a-z0-9]{26}$` || name.String() != `^[a-z][a-z0-9-]{0,39}$` {
		t.Errorf("patterns %s %s", id, name)
	}
}
