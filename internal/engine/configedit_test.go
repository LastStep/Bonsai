package engine

// bonsai.yaml's byte-preserving pack edits (configedit.go; T-6009's plan, step 1, and its proof rows "A ref moves with
// nothing else changed" and "Add and remove are byte-safe"): on this repository's own bonsai.yaml (read from the
// checkout, never written), every pack, quoted and bare refs, and a CRLF copy, and on the shapes the remove bound must
// get right: an entry followed by a comment block, by a top-level key with no blank line, the file's last entry, the
// only entry, a dash-aligned comment above an entry, and a flow packs: [].

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// repoConfig is this repository's own bonsai.yaml, read from the checkout.
func repoConfig(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "bonsai.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n")) // a Windows checkout may hold it CRLF: LF here, CRLF made below
}

func crlf(raw []byte) []byte { return bytes.ReplaceAll(raw, []byte("\n"), []byte("\r\n")) }

// diffSpan gives the bytes of a that differ from b and the bytes of b in their place: the parts between the
// common prefix and the common suffix.
func diffSpan(a, b []byte) (gone, put []byte, at int) {
	p := 0
	for p < len(a) && p < len(b) && a[p] == b[p] {
		p++
	}
	q := 0
	for q < len(a)-p && q < len(b)-p && a[len(a)-1-q] == b[len(b)-1-q] {
		q++
	}
	return a[p : len(a)-q], b[p : len(b)-q], p
}

// readRefs reads the packs' refs from bytes.
func readRefs(t *testing.T, raw []byte) map[string]string {
	t.Helper()
	cfg, err := workspace.ReadConfigFull(raw)
	if err != nil {
		t.Fatalf("does not read: %v\n%s", err, raw)
	}
	out := map[string]string{}
	for _, p := range cfg.Packs {
		out[p.ID] = p.Ref
	}
	return out
}

func TestSetRefChangesOnlyTheValue(t *testing.T) {
	repo := repoConfig(t)
	// Bare refs must start with a letter (format 1's plain values): the repository's commits become tags here.
	barefile := bytes.Replace(repo, []byte(`ref: "4b41790bb40b220785b4d18000824356d4829a6f"`), []byte(`ref: base-v1.0.0`), 1)
	barefile = bytes.Replace(barefile, []byte(`ref: "1a9f6b4ec7848b4d31e8d990f96c6a7afd68873d"`), []byte(`ref: v1.0.0`), 1)
	single := bytes.ReplaceAll(repo, []byte(`ref: "4b41790bb40b220785b4d18000824356d4829a6f"`), []byte(`ref: '4b41790bb40b220785b4d18000824356d4829a6f'`))
	files := map[string][]byte{"repo": repo, "repo-crlf": crlf(repo), "bare": barefile, "bare-crlf": crlf(barefile), "single-quoted": single}
	for name, raw := range files {
		before := readRefs(t, raw)
		for id := range before {
			for _, ref := range []string{"base-v1.0.2", "f", "0123456789abcdef0123456789abcdef01234567", "v10.20.30"} {
				edit, err := SetRef(raw, id, ref)
				if err != nil {
					t.Errorf("%s, %s to %s: %v", name, id, ref, err)
					continue
				}
				after := readRefs(t, edit.Raw)
				for other, r := range before {
					want := r
					if other == id {
						want = ref
					}
					if after[other] != want {
						t.Errorf("%s, %s to %s: %s's ref reads %s, want %s", name, id, ref, other, after[other], want)
					}
				}
				gone, put, at := diffSpan(raw, edit.Raw)
				line := bytes.Count(raw[:at], []byte("\n")) + 1
				if line != edit.Line || bytes.ContainsAny(gone, "\r\n#") || bytes.ContainsAny(put, "\r\n#") ||
					!bytes.Contains([]byte(before[id]), bytes.Trim(gone, `"'`)) || !bytes.Contains([]byte(ref), bytes.Trim(put, `"'`)) {
					t.Errorf("%s, %s to %s: changed %q to %q at line %d (the edit's line %d), want only the ref's value", name, id, ref, gone, put, line, edit.Line)
				}
				if bytes.Count(edit.Raw, []byte("\r\n")) != bytes.Count(raw, []byte("\r\n")) {
					t.Errorf("%s: line endings changed", name)
				}
				if c := commentOf(edit.Before); c == "" || edit.Comment != c || !strings.HasSuffix(edit.After, c) {
					t.Errorf("%s, %s: comment %q, before %q, after %q: want the line's comment kept", name, id, edit.Comment, edit.Before, edit.After)
				}
			}
		}
	}
	// A bare value that would read as another type is written quoted.
	edit, err := SetRef(barefile, "base", "1234567")
	if err != nil || readRefs(t, edit.Raw)["base"] != "1234567" || !strings.Contains(edit.After, `"1234567"`) {
		t.Errorf("a ref of digits: %v, %q", err, edit.After)
	}
	// Refused: a ref this editor does not write, an unknown pack.
	for _, ref := range []string{"-x", "a b", `a"b`, "a#b", ""} {
		if _, err := SetRef(repo, "base", ref); err == nil {
			t.Errorf("the ref %q was written", ref)
		}
	}
	if _, err := SetRef(repo, "nope", "v1.0.0"); err == nil {
		t.Error("an unknown pack's ref was set")
	}
}

// The shapes of packs the remove bound and the append point must get right.
const editHead = "format: bonsai.workspace/1   # f\nid: ws-aaaaaaaaaaaaaaaaaaaaaaaaaa\nname: demo\n"

var editShapes = map[string]string{
	// An entry followed by a comment block and a blank line before the next key.
	"comment-block": editHead + `packs:                # the packs
  - id: one           # first
    source: "https://example.org/one.git"
    path: null
    ref: "v1.0.0"     # its ref
    roles:
      builder:
        minutes:
          default: 240
          # a comment inside the entry
          max: 480
  # the second pack, aligned with its dash
  - id: two
    source: "https://example.org/two.git"
    path: "packs/two"
    ref: "two-v1.0.0"

# protected: the paths an agent changes only with a grant
protected: []
`,
	// A top-level key straight after the last entry, no blank line.
	"no-blank": editHead + `packs:
  - id: one
    source: "https://example.org/one.git"
    path: null
    ref: "v1.0.0"
  - id: two
    source: "https://example.org/two.git"
    path: null
    ref: "v2.0.0"
documents:
  task: work/tasks
`,
	// The last entry is the file's end, with no final line ending; its dash at column 1.
	"at-end": editHead + `protected: []
packs:
 - id: one
   source: "https://example.org/one.git"
   path: null
   ref: "v1.0.0"
 - id: two
   source: "https://example.org/two.git"
   path: null
   ref: v2.0.0     # bare`,
	// A comment at the top level between two entries belongs to neither; one deeper belongs to the first.
	"between": editHead + `packs:
    - id: one
      source: "https://example.org/one.git"
      path: null
      ref: "v1.0.0"
        # deeper: the first entry's
# top level: neither's
    - id: two
      source: "https://example.org/two.git"
      path: null
      ref: "v2.0.0"
`,
}

func TestRemovePackIsByteSafe(t *testing.T) {
	cases := map[string]map[string]string{}
	for name, s := range editShapes {
		cases[name] = map[string]string{"raw": s}
		cases[name+"-crlf"] = map[string]string{"raw": strings.ReplaceAll(s, "\n", "\r\n")}
	}
	repo := string(repoConfig(t))
	cases["repo"] = map[string]string{"raw": repo}
	cases["repo-crlf"] = map[string]string{"raw": strings.ReplaceAll(repo, "\n", "\r\n")}
	// The lines each removal must take, exactly, by shape (LF; the CRLF copies the same lines).
	gone := map[string]map[string]string{
		"comment-block": {
			"one": "  - id: one           # first\n    source: \"https://example.org/one.git\"\n    path: null\n    ref: \"v1.0.0\"     # its ref\n    roles:\n      builder:\n        minutes:\n          default: 240\n          # a comment inside the entry\n          max: 480\n",
			"two": "  # the second pack, aligned with its dash\n  - id: two\n    source: \"https://example.org/two.git\"\n    path: \"packs/two\"\n    ref: \"two-v1.0.0\"\n",
		},
		"no-blank": {
			"one": "  - id: one\n    source: \"https://example.org/one.git\"\n    path: null\n    ref: \"v1.0.0\"\n",
			"two": "  - id: two\n    source: \"https://example.org/two.git\"\n    path: null\n    ref: \"v2.0.0\"\n",
		},
		"at-end": {
			"one": " - id: one\n   source: \"https://example.org/one.git\"\n   path: null\n   ref: \"v1.0.0\"\n",
			"two": " - id: two\n   source: \"https://example.org/two.git\"\n   path: null\n   ref: v2.0.0     # bare",
		},
		"between": {
			"one": "    - id: one\n      source: \"https://example.org/one.git\"\n      path: null\n      ref: \"v1.0.0\"\n        # deeper: the first entry's\n",
			"two": "    - id: two\n      source: \"https://example.org/two.git\"\n      path: null\n      ref: \"v2.0.0\"\n",
		},
	}
	for name, c := range cases {
		raw := []byte(c["raw"])
		before := readRefs(t, raw)
		for id := range before {
			edit, err := RemovePack(raw, id)
			if err != nil {
				t.Errorf("%s, remove %s: %v", name, id, err)
				continue
			}
			after := readRefs(t, edit.Raw)
			if _, ok := after[id]; ok || len(after) != len(before)-1 {
				t.Errorf("%s, remove %s: packs now %v", name, id, after)
			}
			for other, r := range before {
				if other != id && after[other] != r {
					t.Errorf("%s, remove %s: %s's ref %s, want %s", name, id, other, after[other], r)
				}
			}
			removed, put, _ := diffSpan(raw, edit.Raw)
			if len(put) != 0 {
				t.Errorf("%s, remove %s: bytes put in: %q", name, id, put)
			}
			base := strings.TrimSuffix(name, "-crlf")
			if want, ok := gone[base][id]; ok {
				if strings.HasSuffix(name, "-crlf") {
					want = strings.ReplaceAll(want, "\n", "\r\n")
				}
				// The removed lines are one contiguous run; diffSpan may place it a line earlier when lines repeat,
				// so the check is that raw holds the run once and the result is raw without it.
				if !bytes.Equal(edit.Raw, bytes.Replace(raw, []byte(want), nil, 1)) || len(removed) != len(want) {
					t.Errorf("%s, remove %s: took %q, want exactly %q", name, id, removed, want)
				}
			}
		}
	}
	// The only entry: packs: [] keeps its comment's column.
	only := editHead + "packs:                # the packs\n  - id: one\n    source: \"https://example.org/one.git\"\n    path: null\n    ref: \"v1.0.0\"\nprotected: []\n"
	edit, err := RemovePack([]byte(only), "one")
	if err != nil {
		t.Fatal(err)
	}
	if want := editHead + "packs: []             # the packs\nprotected: []\n"; string(edit.Raw) != want {
		t.Errorf("the only entry:\n%s\nwant\n%s", edit.Raw, want)
	}
	if _, err := RemovePack([]byte(only), "two"); err == nil {
		t.Error("an unknown pack was removed")
	}
}

func TestAddPackIsByteSafe(t *testing.T) {
	cases := map[string][]byte{"repo": repoConfig(t), "repo-crlf": crlf(repoConfig(t))}
	for name, s := range editShapes {
		cases[name] = []byte(s)
		cases[name+"-crlf"] = []byte(strings.ReplaceAll(s, "\n", "\r\n"))
	}
	cases["flow"] = []byte(editHead + "packs: []             # the packs\nprotected: []\n")
	cases["flow-crlf"] = crlf(cases["flow"])
	for name, raw := range cases {
		before := readRefs(t, raw)
		for _, p := range []NewPack{{ID: "extra", Source: "https://example.org/extra.git", Ref: "v1.0.0"},
			{ID: "sub", Source: "https://example.org/mono.git", Path: "packs/sub", Ref: "sub-v0.1.0"}} {
			edit, err := AddPack(raw, p)
			if err != nil {
				t.Errorf("%s, add %s: %v", name, p.ID, err)
				continue
			}
			cfg, err := workspace.ReadConfigFull(edit.Raw)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			last := cfg.Packs[len(cfg.Packs)-1]
			if len(cfg.Packs) != len(before)+1 || last.ID != p.ID || last.Source != p.Source || last.Path != p.Path || last.Ref != p.Ref {
				t.Errorf("%s, add %s: packs %+v", name, p.ID, cfg.Packs)
			}
			gone, put, _ := diffSpan(raw, edit.Raw)
			if name == "flow" || name == "flow-crlf" {
				// The flow list opened: packs: [] becomes packs:, the comment's column kept.
				if !bytes.Contains(edit.Raw, []byte("packs:                # the packs")) {
					t.Errorf("%s: the packs line:\n%s", name, edit.Raw)
				}
			} else if len(gone) != 0 && !(name == "at-end" || name == "at-end-crlf") {
				t.Errorf("%s, add %s: bytes taken out: %q", name, p.ID, gone)
			}
			if !bytes.Contains(put, []byte("id: "+p.ID)) || !bytes.Contains(put, []byte("# the pack's id")) {
				t.Errorf("%s, add %s: put in %q, want init's form with its comments", name, p.ID, put)
			}
			if crlfFile := bytes.Contains(raw, []byte("\r\n")); crlfFile && bytes.Count(edit.Raw, []byte("\n")) != bytes.Count(edit.Raw, []byte("\r\n")) {
				t.Errorf("%s: a line ending other than the file's", name)
			}
			if !strings.Contains(edit.After, "- id: "+p.ID) {
				t.Errorf("%s: the edit's line %d reads %q", name, edit.Line, edit.After)
			}
		}
	}
	if _, err := AddPack(repoConfig(t), NewPack{ID: "base", Source: "x", Ref: "v1.0.0"}); err == nil {
		t.Error("a pack bonsai.yaml already names was added")
	}
}
