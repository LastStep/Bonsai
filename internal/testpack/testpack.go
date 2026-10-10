// Package testpack builds, for Go tests, a local pack repository with seven commits, A to G, shaped like Bonsai's
// public test pack (github.com/LastStep/bonsai-test-pack, plan part 4a; E and F from step 5.1.1, G from 5.1.9), so
// the engine's tests fetch a pack with no network; fixture packs (Fixture) for the cases the test pack does not hold;
// and a documented pack folder (DocumentedPack) that every rule of bonsai check --pack reaches. Only tests import it.
//
// What each commit changes, as in the public pack (its README's table):
//
//	A  pack.yaml lists guide.md (kind pack), start.md (kind once), the block and one hook line "echo demo hook A"
//	B  guide.md becomes edition 2; extra.md is new (kind pack); the marker says commit B
//	C  guide.md becomes edition 3
//	D  only the hook line's command: "echo demo hook D"
//	E  the mixed update: guide.md becomes edition 4 and the hook line "echo demo hook E" (its entry now says runs: [])
//	F  only a hook of the plugin itself: hooks/hooks.json, a SessionStart hook that runs an echo
//	G  only documentation: pack.yaml gains a comment on every key that had none, and documents: [] and protected: []
//	   (every field written, as bonsai check --pack holds a pack's maker to); nothing the engine writes changes
//
// Every repository and project it makes lives in the test's temporary folder, with git isolated from the
// machine's own configuration (no global or system config, no repository above the folder).
package testpack

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// ID is the pack's id.
const ID = "demo-pack"

// Pack is a built pack repository: its source (a bare repository's folder) and its seven commits.
type Pack struct {
	Source              string
	A, B, C, D, E, F, G string
}

// PluginHooks is commit F's hooks/hooks.json: a hook of the plugin itself, which Claude Code runs on its own.
const PluginHooks = `{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          { "type": "command", "command": "echo demo plugin hook F" }
        ]
      }
    ]
  }
}
`

// Isolate points git at an empty configuration and no ignore file of the person's, stops it looking above tmp, and
// sets BONSAI_HOME to tmp/home. Every test that runs git or Bonsai calls it first.
func Isolate(t *testing.T, tmp string) (home string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH: the engine fetches packs and finds projects with git")
	}
	empty := filepath.Join(tmp, "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	// git also reads the person's own ignore file, $XDG_CONFIG_HOME/git/ignore (or ~/.config/git/ignore): none here.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "xdg"))
	t.Setenv("GIT_CEILING_DIRECTORIES", tmp)
	t.Setenv("CLAUDE_CODE_CHILD_SESSION", "")
	home = filepath.Join(tmp, "home")
	t.Setenv("BONSAI_HOME", home)
	return home
}

// Git runs git in dir, as a fixed test user, and returns its output.
func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=bonsai-test", "-c", "user.email=bonsai-test",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main", "-c", "core.autocrlf=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// PackYAML is the pack's bonsai/pack.yaml: hook is the hook line's letter, extra whether extra.md is listed, runs
// whether the hook entry says runs: [] (from E; a missing runs reads the same), documented whether every key has its
// comment and every field is written, documents: [] and protected: [] among them (from G; missing, they read the same).
func PackYAML(hook string, extra, runs, documented bool) string {
	s := `# bonsai/pack.yaml: a test pack for Bonsai's engine tests.
format: bonsai.pack/1                # the format
id: demo-pack                        # the pack's id
version: "0.1.0"                     # the same at every commit
needs:                               # what a machine needs
  claude_code: "2.1.0"               # the oldest Claude Code
block: block.md                      # the block
files:                               # the files the engine writes
  - path: demo/guide.md              # changed by B and C
    from: guide.md                   # its source
    kind: pack                       # Bonsai's file
  - path: demo/start.md              # written once
    from: start.md                   # its source
    kind: once                       # the project's after init
`
	if extra {
		s += `  - path: demo/extra.md              # new at B
    from: extra.md                   # its source
    kind: pack                       # Bonsai's file
`
	}
	s += `hooks:                               # hook lines
  - event: SessionStart              # a session starts
    matcher: startup                 # new sessions only
    command: "echo demo hook ` + hook + `"     # D and E change it
`
	if runs {
		s += `    runs: []                         # the pack files it runs: none
`
	}
	if !documented {
		return s + `    why: "Prints which demo hook line is in place; it blocks nothing."
deny:                                # deny rules
  - rule: "Edit(demo/never.txt)"     # an example rule
    why: "Agents cannot edit demo/never.txt (the demo pack's rule)."
`
	}
	return s + `    why: "Prints which demo hook line is in place; it blocks nothing."  # preview text
deny:                                # deny rules
  - rule: "Edit(demo/never.txt)"     # an example rule
    why: "Agents cannot edit demo/never.txt (the demo pack's rule)."  # preview text
documents: []                        # the document kinds it declares: none
protected: []                        # the paths it declares protected: none
`
}

const block = "<!--\nbonsai/block.md: the demo pack's block. This comment stays in the pack.\n-->\nThe demo pack is linked: its role is demo-pack:marker.\n"

// Build makes the pack repository with commits A to G and returns it.
func Build(t *testing.T, tmp string) *Pack {
	t.Helper()
	work := filepath.Join(tmp, "pack-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	Git(t, work, "init", "-q")
	write := func(rel, content string) {
		p := filepath.Join(work, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	commit := func(msg string) string {
		Git(t, work, "add", "-A")
		Git(t, work, "commit", "-q", "-m", msg)
		return Git(t, work, "rev-parse", "HEAD")
	}
	p := &Pack{}
	write(".claude-plugin/plugin.json", "{\"name\": \"demo-pack\", \"description\": \"A test pack.\"}\n")
	write("agents/marker.md", "---\nname: marker\ndescription: says commit A\n---\nSay commit A.\n")
	write("bonsai/pack.yaml", PackYAML("A", false, false, false))
	write("bonsai/block.md", block)
	write("bonsai/files/guide.md", "# Guide\n\nEdition 1.\n")
	write("bonsai/files/start.md", "# Start here\n\nThe project's own file once written.\n")
	p.A = commit("A")
	write("agents/marker.md", "---\nname: marker\ndescription: says commit B\n---\nSay commit B.\n")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 2.\n")
	write("bonsai/files/extra.md", "# Extra\n\nNew at commit B.\n")
	write("bonsai/pack.yaml", PackYAML("A", true, false, false))
	p.B = commit("B")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 3.\n")
	p.C = commit("C")
	write("bonsai/pack.yaml", PackYAML("D", true, false, false))
	p.D = commit("D")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 4.\n")
	write("bonsai/pack.yaml", PackYAML("E", true, true, false))
	p.E = commit("E")
	write("hooks/hooks.json", PluginHooks)
	p.F = commit("F")
	write("bonsai/pack.yaml", PackYAML("E", true, true, true))
	p.G = commit("G")
	p.Source = filepath.Join(tmp, "demo-pack.git")
	Git(t, tmp, "clone", "-q", "--bare", "--", work, p.Source)
	return p
}

// Fixture builds a pack repository at tmp/<name>.git from commits, each the files it writes (a path in the
// repository, forward slashes, to its content; "" removes the file), applied in path order on the commit before.
// It returns the repository's source and its commits, in order. For the cases the test pack does not hold (a pack
// with no hook line, a hook line that runs a pack file, a plugin's code parts).
func Fixture(t *testing.T, tmp, name string, commits ...map[string]string) (source string, shas []string) {
	t.Helper()
	work := filepath.Join(tmp, name+"-work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}
	Git(t, work, "init", "-q")
	for i, files := range commits {
		paths := make([]string, 0, len(files))
		for rel := range files {
			paths = append(paths, rel)
		}
		sort.Strings(paths)
		for _, rel := range paths {
			p := filepath.Join(work, filepath.FromSlash(rel))
			if files[rel] == "" {
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
				continue
			}
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(files[rel]), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		Git(t, work, "add", "-A")
		Git(t, work, "commit", "-q", "--allow-empty", "-m", "commit "+strconv.Itoa(i+1))
		shas = append(shas, Git(t, work, "rev-parse", "HEAD"))
	}
	source = filepath.Join(tmp, name+".git")
	Git(t, tmp, "clone", "-q", "--bare", "--", work, source)
	return source, shas
}

// The fixture packs: each a Claude Code plugin with a bonsai/ folder, built by Fixture at tmp/<id>.git, for the
// consent cases the test pack does not hold (plan-5, piece 5.1.1). Each returns its source and its commits.
const fixturePlugin = "{\"name\": \"%s\", \"description\": \"A fixture pack.\"}\n"

func fixtureYAML(id, files, hooks string) string {
	return "format: bonsai.pack/1\nid: " + id + "\nversion: \"0.1.0\"\nfiles:" + files + "\nhooks:" + hooks + "\ndeny: []\n"
}

// QuietPack (quiet-pack) has no hook line and no plugin code: one role and one pack file, quiet/readme.md. A first
// link to it writes Bonsai's own lines on --yes alone.
func QuietPack(t *testing.T, tmp string) (string, []string) {
	t.Helper()
	return Fixture(t, tmp, "quiet-pack", map[string]string{
		".claude-plugin/plugin.json": strings.Replace(fixturePlugin, "%s", "quiet-pack", 1),
		"agents/helper.md":           "---\nname: helper\ndescription: helps\n---\nHelp.\n",
		"bonsai/pack.yaml":           fixtureYAML("quiet-pack", "\n  - path: quiet/readme.md\n    from: readme.md\n    kind: pack", " []"),
		"bonsai/files/readme.md":     "# Quiet\n",
	})
}

// SidePack (side-pack) is a second pack beside the test pack, for taking a pack out of a project (step 5.1.7): two pack
// files (side/a.md, side/b.md), a file written once (side/once.md), a block.md, one hook line ("echo side hook", code
// at a first link) and one deny rule (Read(side/secret.txt)); its plugin has a role and no code part.
func SidePack(t *testing.T, tmp string) (string, []string) {
	t.Helper()
	yaml := "format: bonsai.pack/1\nid: side-pack\nversion: \"0.2.0\"\nblock: block.md\nfiles:\n" +
		"  - path: side/a.md\n    from: a.md\n    kind: pack\n" +
		"  - path: side/b.md\n    from: b.md\n    kind: pack\n" +
		"  - path: side/once.md\n    from: once.md\n    kind: once\n" +
		"hooks:\n  - event: SessionStart\n    matcher: startup\n    command: \"echo side hook\"\n    why: \"Says the side pack is here; it blocks nothing.\"\n" +
		"deny:\n  - rule: \"Read(side/secret.txt)\"\n    why: \"Agents cannot read side/secret.txt (the side pack's rule).\"\n"
	return Fixture(t, tmp, "side-pack", map[string]string{
		".claude-plugin/plugin.json": strings.Replace(fixturePlugin, "%s", "side-pack", 1),
		"agents/helper.md":           "---\nname: helper\ndescription: helps on the side\n---\nHelp.\n",
		"bonsai/pack.yaml":           yaml,
		"bonsai/block.md":            "The side pack is linked.\n",
		"bonsai/files/a.md":          "# Side A\n",
		"bonsai/files/b.md":          "# Side B\n",
		"bonsai/files/once.md":       "# Side, once\n",
	})
}

// RunPack (run-pack): its hook line runs one of its files, run/hello.sh (declared in runs); commit 2 changes only
// that file (bonsai/files/hello.sh), commit 3 only a file no hook runs (run/notes.md).
func RunPack(t *testing.T, tmp string) (string, []string) {
	t.Helper()
	files := "\n  - path: run/hello.sh\n    from: hello.sh\n    kind: pack\n  - path: run/notes.md\n    from: notes.md\n    kind: pack"
	hooks := "\n  - event: SessionStart\n    matcher: startup\n    command: \"sh run/hello.sh\"\n    runs: [\"run/hello.sh\"]\n" +
		"    why: \"Says hello when a session starts; it blocks nothing.\""
	return Fixture(t, tmp, "run-pack",
		map[string]string{
			".claude-plugin/plugin.json": strings.Replace(fixturePlugin, "%s", "run-pack", 1),
			"bonsai/pack.yaml":           fixtureYAML("run-pack", files, hooks),
			"bonsai/files/hello.sh":      "echo hello 1\n",
			"bonsai/files/notes.md":      "Notes 1.\n",
		},
		map[string]string{"bonsai/files/hello.sh": "echo hello 2\n"},
		map[string]string{"bonsai/files/notes.md": "Notes 2.\n"})
}

// MCPPack (mcp-pack): its plugin carries an MCP server (.mcp.json) from its first commit; commit 2 changes only a
// role, commit 3 takes the server out and changes the role again.
func MCPPack(t *testing.T, tmp string) (string, []string) {
	t.Helper()
	return Fixture(t, tmp, "mcp-pack",
		map[string]string{
			".claude-plugin/plugin.json": strings.Replace(fixturePlugin, "%s", "mcp-pack", 1),
			".mcp.json":                  "{\"mcpServers\": {\"demo\": {\"command\": \"echo\"}}}\n",
			"agents/helper.md":           "Help 1.\n",
			"bonsai/pack.yaml":           fixtureYAML("mcp-pack", " []", " []"),
		},
		map[string]string{"agents/helper.md": "Help 2.\n"},
		map[string]string{".mcp.json": "", "agents/helper.md": "Help 3.\n"})
}

// BasePackYAML is DeclaringPack's bonsai/pack.yaml: an always-on protocol file in work/protocols/, a block, a deny
// rule, the document kind plan and a protected glob, and no hook line.
const BasePackYAML = `format: bonsai.pack/1
id: base
version: "1.0.0"
block: block.md
files:
  - path: work/protocols/session-start.md
    from: session-start.md
    kind: pack
hooks: []
deny:
  - rule: "Read(secrets/**)"
    why: "Agents cannot read secrets/ (the fixture's wall)."
documents:
  - kind: plan
    path: work/plans
    file: null
    id: "^P-T-[0-9]{4,6}$"
    statuses: ["draft", "approved", "done"]
    person:
      - ["draft", "approved"]
    agent:
      - ["approved", "done"]
    stamp:
      approved: approved
    task_field: task
  - kind: bugs
    path: null
    file: work/bugs.md
    id: null
    statuses: []
    person: []
    agent: []
    stamp: {}
    task_field: null
protected: ["work/plans/**"]
`

// BaseLabelsYAML is DeclaringPack's bonsai/labels.yaml: two of Bonsai's own labels (contract §5.6), which the base
// pack declares from step 5.5; until then a fixture declares them (plan-5, 5.1.5).
const BaseLabelsYAML = `format: bonsai.labels/1
namespace: bonsai
version: 1
labels:
  - name: bonsai.allows
    kind: list
    values: []
    items: text
    pattern: null
    max: null
    kinds: ["task"]
    set_by: agent
    grants: true
    description: "Protected paths this task may change, as globs."
  - name: bonsai.branch
    kind: text
    values: []
    items: null
    pattern: "^[A-Za-z0-9._/-]{1,100}$"
    max: null
    kinds: ["task"]
    set_by: agent
    grants: false
    description: "The branch the task is built on; empty means the base branch."
`

// BaseLanesYAML is DeclaringPack's bonsai/lanes.yaml.
const BaseLanesYAML = `format: bonsai.lanes/1
lanes:
  - name: light
    approve_first: false
    close: agent
    description: "Fixes, tweaks, tests, docs."
  - name: full
    approve_first: true
    close: person
    description: "A feature: a person approves the plan first and closes the task."
`

// DeclaringPack (id base) declares one of each kind the lock's declares holds but hook lines: lanes, two document
// kinds (plan, a folder; bugs, a file), Bonsai's own label namespace (two definitions), a protected glob and a deny
// rule, with an always-on protocol file and a block. It carries no code.
func DeclaringPack(t *testing.T, tmp string) (string, []string) {
	t.Helper()
	return Fixture(t, tmp, "base", map[string]string{
		".claude-plugin/plugin.json":    strings.Replace(fixturePlugin, "%s", "base", 1),
		"bonsai/pack.yaml":              BasePackYAML,
		"bonsai/labels.yaml":            BaseLabelsYAML,
		"bonsai/lanes.yaml":             BaseLanesYAML,
		"bonsai/block.md":               "<!-- the fixture's block docs -->\nThe base fixture is linked.\n",
		"bonsai/files/session-start.md": "# Session start\n\nRead STATE first.\n",
	})
}

// Tag points the tag name at commit in a pack repository (a bare one, as Fixture and Build make), moving it if it is
// there: a pack's maker moving a release tag, which Bonsai refuses (spec §5).
func Tag(t *testing.T, source, name, commit string) {
	t.Helper()
	Git(t, source, "tag", "-f", name, commit)
}

// Project makes an empty git checkout at tmp/<name> and returns it.
func Project(t *testing.T, tmp, name string) string {
	t.Helper()
	root := filepath.Join(tmp, name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	Git(t, root, "init", "-q")
	r, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// SetRef changes the ref of bonsai.yaml's one pack, as a person would.
func SetRef(t *testing.T, root, from, to string) {
	t.Helper()
	p := filepath.Join(root, "bonsai.yaml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), from) {
		t.Fatalf("bonsai.yaml does not hold %s", from)
	}
	if err := os.WriteFile(p, []byte(strings.Replace(string(b), from, to, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

// DocumentedPack is a pack folder's files (a path in the folder, forward slashes, to its content) that every rule of
// bonsai check --pack reaches and that passes them all (step 5.1.9): a pack.yaml with a comment on every key and every
// field written (a files entry the hook line runs, listed in its runs; a deny rule with its why; the document kind
// note; a protected glob), a labels.yaml (one choice label) and a lanes.yaml (light and full), each key commented, a
// block.md, a plugin.json with no version, a plugin hook, and two template skills: note (the pack's own kind) and task
// (bonsai.task/1, its fields table equal to the schema's fields). Each test case changes one file to break one rule.
func DocumentedPack() map[string]string {
	return map[string]string{
		".claude-plugin/plugin.json": "{\"name\": \"docs-pack\", \"description\": \"A fixture pack, documented in full.\"}\n",
		"hooks/hooks.json":           `{"hooks": {"SessionStart": [{"matcher": "startup", "hooks": [{"type": "command", "command": "echo docs plugin hook"}]}]}}` + "\n",
		"bonsai/pack.yaml": `# bonsai/pack.yaml: the documented fixture pack's manifest (bonsai.pack/1).
# Every key carries its comment, and every field is written.
format: bonsai.pack/1                 # the format and its version
id: docs-pack                         # the pack's id
version: "0.1.0"                      # the pack's version
needs:                                # what a machine needs
  claude_code: null                   # no floor of its own
block: block.md                       # its part of the block
files:                                # the files the engine writes
  - path: docs/run.sh                 # the script the hook line runs
    from: run.sh                      # its source
    kind: pack                        # Bonsai's file
  - path: work/protocols/start.md     # an always-on protocol file
    from: start.md                    # its source
    kind: pack                        # Bonsai's file
hooks:                                # hook lines
  - event: SessionStart               # a session starts
    matcher: startup                  # new sessions only
    command: "sh docs/run.sh"         # runs the pack's script
    runs: ["docs/run.sh"]             # the file it runs
    # why: the line in the preview, too long to share its line
    why: "Runs the pack's greeting when a session starts; it blocks nothing."
deny:                                 # deny rules
  - rule: "Read(docs/secret.txt)"     # an example rule
    why: "Agents cannot read docs/secret.txt."  # the preview's sentence
documents:                            # the kinds it declares
  - kind: note                        # a note
    path: work/notes                  # its folder
    file: null                        # a folder, not one file
    id: "^N-[0-9]{4}$"                # a note's id
    statuses: [draft, done]           # its statuses
    person: []                        # no person's moves
    agent:                            # the moves agents make
      - [draft, done]                 # an agent finishes a note
    stamp:                            # dates set on a status
      done: finished                  # finished: set at done
    task_field: null                  # it belongs to no task
protected: ["work/notes/**"]          # paths an agent changes only with a grant
`,
		"bonsai/labels.yaml": `# bonsai/labels.yaml: the documented fixture pack's labels (bonsai.labels/1).
format: bonsai.labels/1               # the format
namespace: docs-pack                  # the pack's own id
version: 1                            # this set's version
labels:                               # one entry per label
  - name: docs-pack.size              # a choice label
    kind: choice                      # one of its values
    values: [small, large]            # its values
    items: null                       # not a list
    pattern: null                     # no pattern
    max: null                         # no length limit
    kinds: [note]                     # notes carry it
    set_by: agent                     # agents write it
    grants: false                     # it grants nothing
    description: "How big the note is."  # what agents see
`,
		"bonsai/lanes.yaml": `# bonsai/lanes.yaml: the documented fixture pack's lanes (bonsai.lanes/1).
format: bonsai.lanes/1                # the format
lanes:                                # its lanes
  - name: light                       # the light lane
    approve_first: false              # no approval first
    close: agent                      # an agent closes it
    description: "Small fixes."       # what agents read
  - name: full                        # the full lane
    approve_first: true               # a person approves first
    close: person                     # a person closes it
    description: "A feature."         # what agents read
`,
		"bonsai/block.md":       "<!--\nbonsai/block.md: the documented fixture pack's block. This comment stays in the pack.\n-->\nThe docs pack is linked.\n",
		"bonsai/files/run.sh":   "echo hello from the docs pack\n",
		"bonsai/files/start.md": "# Start\n\nRead this first.\n",
		"skills/hello/SKILL.md": "---\nname: hello\ndescription: \"Says hello; no template.\"\n---\n\n| Field | Meaning |\n|---|---|\n| `x` | not a fields table |\n",
		"skills/note/SKILL.md": "---\nname: note\ndescription: \"The note template: use it to write a note.\"\n---\n\n" +
			"<!-- The note template of the docs fixture pack. -->\n\n" +
			"| Field | Meaning | Allowed values | Example |\n|---|---|---|---|\n" +
			"| `id` | The note's id | `N-` and four digits | `N-0001` |\n" +
			"| `status` | Where it stands | `draft`, `done` | `draft` |\n" +
			"| `lane` | Its lane | `light`, `full`, or `null` | `light` |\n" +
			"| `labels` | Its labels | a mapping | `{}` |\n" +
			"| `labels.docs-pack.size` | How big it is | `small` or `large` | `small` |\n\n" +
			"```markdown\n---\nid: N-0000   # the note's id; how to fill: skill docs-pack:note\nstatus: draft\nlane: null\nlabels: {}\n---\n\n# A note\n```\n",
		"skills/task/SKILL.md": "---\nname: task\ndescription: \"The task template.\"\n---\n\n" +
			"| Field | Meaning | Allowed values | Example |\n| --- | --- | --- | --- |\n" +
			"| `format` | The format | `bonsai.task/1` | `bonsai.task/1` |\n" +
			"| `id` | The task's id | T- and digits | `T-0001` |\n" +
			"| `title` | Its title | text | `A task` |\n" +
			"| `status` | Where it stands | `todo`, `plan`, `approved`, `running`, `verify`, `done`, `blocked`, `cut` | `todo` |\n" +
			"| `lane` | Its lane | `light`, `full`, or `null` | `light` |\n" +
			"| `done_when` | What done means | a list of text | `[]` |\n" +
			"| `depends_on` | Tasks first | task ids | `[]` |\n" +
			"| `blocked_by` | What blocks it | text, or `null` | `null` |\n" +
			"| `created` | Its day | a date | `2026-10-10` |\n" +
			"| `started` | When it ran | a date, or `null` | `null` |\n" +
			"| `finished` | When it ended | a date, or `null` | `null` |\n" +
			"| `labels` | Its labels | a mapping | `{}` |\n\n" +
			"~~~markdown\n---\nformat: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill docs-pack:task\n" +
			"id: T-0000\ntitle: <a title>\nstatus: todo\nlane: null\ndone_when: []\ndepends_on: []\nblocked_by: null\n" +
			"created: <today>\nstarted: null\nfinished: null\nlabels: {}\n---\n\n# <a title>\n~~~\n",
	}
}

// WriteFolder writes files (a path in the folder, forward slashes, to its content; "" leaves the file out) into dir.
func WriteFolder(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		if content == "" {
			continue
		}
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
