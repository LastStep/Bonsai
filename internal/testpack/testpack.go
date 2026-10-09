// Package testpack builds, for Go tests, a local pack repository with six commits, A to F, shaped like Bonsai's
// public test pack (github.com/LastStep/bonsai-test-pack, plan part 4a; E and F from step 5.1.1), so the engine's
// tests fetch a pack with no network; and fixture packs (Fixture) for the cases the test pack does not hold. Only
// tests import it.
//
// What each commit changes, as in the public pack (its README's table):
//
//	A  pack.yaml lists guide.md (kind pack), start.md (kind once), the block and one hook line "echo demo hook A"
//	B  guide.md becomes edition 2; extra.md is new (kind pack); the marker says commit B
//	C  guide.md becomes edition 3
//	D  only the hook line's command: "echo demo hook D"
//	E  the mixed update: guide.md becomes edition 4 and the hook line "echo demo hook E" (its entry now says runs: [])
//	F  only a hook of the plugin itself: hooks/hooks.json, a SessionStart hook that runs an echo
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

// Pack is a built pack repository: its source (a bare repository's folder) and its six commits.
type Pack struct {
	Source           string
	A, B, C, D, E, F string
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
// whether the hook entry says runs: [] (from E; a missing runs reads the same).
func PackYAML(hook string, extra, runs bool) string {
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
	s += `    why: "Prints which demo hook line is in place; it blocks nothing."
deny:                                # deny rules
  - rule: "Edit(demo/never.txt)"     # an example rule
    why: "Agents cannot edit demo/never.txt (the demo pack's rule)."
`
	return s
}

const block = "<!--\nbonsai/block.md: the demo pack's block. This comment stays in the pack.\n-->\nThe demo pack is linked: its role is demo-pack:marker.\n"

// Build makes the pack repository with commits A to D and returns it.
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
	write("bonsai/pack.yaml", PackYAML("A", false, false))
	write("bonsai/block.md", block)
	write("bonsai/files/guide.md", "# Guide\n\nEdition 1.\n")
	write("bonsai/files/start.md", "# Start here\n\nThe project's own file once written.\n")
	p.A = commit("A")
	write("agents/marker.md", "---\nname: marker\ndescription: says commit B\n---\nSay commit B.\n")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 2.\n")
	write("bonsai/files/extra.md", "# Extra\n\nNew at commit B.\n")
	write("bonsai/pack.yaml", PackYAML("A", true, false))
	p.B = commit("B")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 3.\n")
	p.C = commit("C")
	write("bonsai/pack.yaml", PackYAML("D", true, false))
	p.D = commit("D")
	write("bonsai/files/guide.md", "# Guide\n\nEdition 4.\n")
	write("bonsai/pack.yaml", PackYAML("E", true, true))
	p.E = commit("E")
	write("hooks/hooks.json", PluginHooks)
	p.F = commit("F")
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
