package main

// check's one table of findings and warnings (format.CheckWords, plan-5 5.1.6), walked by one test: each word has a
// case here that makes a project (or this machine) where check finds it, and the test runs bonsai check on it, as an
// agent would, with --json and without. A word with no case fails, and so does a case for a word not in the table.
// For each, the word is in its own list (findings, or warnings, which never change the exit code), the output fits
// bonsai.check/1, and every next step names only commands that exist (agents first, Rohan 9 Oct): each "run: <command>"
// is a bonsai word with only flags its table has and has built, or a git or claude command of the few Bonsai names,
// and every "bonsai <word>" a next step mentions is a word Bonsai has.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// write writes a file in the project.
func (c *cli) write(rel, content string) {
	c.t.Helper()
	p := filepath.Join(c.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		c.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		c.t.Fatal(err)
	}
}

func (c *cli) read(rel string) string {
	c.t.Helper()
	b, err := os.ReadFile(filepath.Join(c.root, filepath.FromSlash(rel)))
	if err != nil {
		c.t.Fatal(err)
	}
	return string(b)
}

// link links the project to the test pack at A.
func (c *cli) link() {
	c.t.Helper()
	if code, out, errOut := c.run("", append(c.linkArgs(c.pack.A), "--yes")...); code != 0 {
		c.t.Fatalf("link: %d\n%s%s", code, out, errOut)
	}
}

// linkTo links the project to another pack repository at ref.
func (c *cli) linkTo(source, ref string) {
	c.t.Helper()
	if code, out, errOut := c.run("", "init", "--name", "demo", "--source", source, "--ref", ref, "--allow-exec", "--yes"); code != 0 {
		c.t.Fatalf("link: %d\n%s%s", code, out, errOut)
	}
}

// commit commits everything in the project.
func (c *cli) commit(msg string) {
	c.t.Helper()
	testpack.Git(c.t, c.root, "add", "-A")
	testpack.Git(c.t, c.root, "commit", "-q", "-m", msg)
}

// task writes a format-1 task file.
func (c *cli) task(id, status, lane, labels string) string {
	c.t.Helper()
	rel := "work/tasks/" + id + "-x.md"
	if lane == "" {
		lane = "null"
	}
	if labels == "" {
		labels = " {}"
	}
	c.write(rel, "---\nformat: bonsai.task/1\nid: "+id+"\ntitle: A task\nstatus: "+status+"\nlane: "+lane+
		"\ndone_when: []\ndepends_on: []\nblocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels:"+labels+"\n---\n")
	return rel
}

// fakeClaude stands in for claude --version.
func fakeClaude(t *testing.T, out string, err error) {
	t.Helper()
	saved := claudeVersion
	t.Cleanup(func() { claudeVersion = saved })
	claudeVersion = func() (string, error) { return out, err }
}

// fakeList stands in for claude plugin list.
type fakeList struct {
	list []engine.InstalledPlugin
	err  error
}

func (f fakeList) List(string) ([]engine.InstalledPlugin, error) { return f.list, f.err }
func (f fakeList) Install(string, string) (engine.InstallResult, error) {
	return engine.InstallResult{}, errors.New("not asked in check")
}

func setPlugins(t *testing.T, p engine.PluginCLI) {
	t.Helper()
	saved := pluginCLI
	t.Cleanup(func() { pluginCLI = saved })
	pluginCLI = p
}

// labelsPack is a fixture pack in Bonsai's namespace defining bonsai.allows, as the declaring pack does: two such
// packs in one workspace define one name twice.
func labelsPack(t *testing.T, tmp, id string) (string, []string) {
	t.Helper()
	return testpack.Fixture(t, tmp, id, map[string]string{
		".claude-plugin/plugin.json": `{"name": "` + id + `", "description": "A fixture pack."}` + "\n",
		"bonsai/pack.yaml":           "format: bonsai.pack/1\nid: " + id + "\nversion: \"0.1.0\"\nfiles: []\nhooks: []\ndeny: []\n",
		"bonsai/labels.yaml":         testpack.BaseLabelsYAML,
	})
}

// addPack adds a pack to bonsai.yaml and updates.
func (c *cli) addPack(id, source, ref string) {
	c.t.Helper()
	c.write("bonsai.yaml", strings.Replace(c.read("bonsai.yaml"), "\ndocuments:", "\n  - id: "+id+"\n    source: \""+
		filepath.ToSlash(source)+"\"\n    path: null\n    ref: \""+ref+"\"\ndocuments:", 1))
	if code, out, errOut := c.run("", "update", "--yes", "--allow-exec"); code != 0 {
		c.t.Fatalf("update with %s: %d\n%s%s", id, code, out, errOut)
	}
}

// declaring links the project to the declaring fixture pack (lanes light and full, Bonsai's two labels).
func (c *cli) declaring() {
	c.t.Helper()
	src, shas := testpack.DeclaringPack(c.t, c.tmp)
	c.linkTo(src, shas[0])
}

// approveCase makes a task in the lane full (approve_first) that went from todo to running with no approval, each
// move its own commit.
func (c *cli) approveCase() {
	c.t.Helper()
	c.declaring()
	c.commit("link")
	c.task("T-0901", "todo", "full", "")
	c.commit("a task")
	c.task("T-0901", "running", "full", "")
	c.commit("running, with no approval")
}

// fakeBonsai puts a bonsai on the PATH (a file named as Windows or Linux finds it; no mode needed on Windows) and
// returns it.
func fakeBonsai(t *testing.T, dir string) string {
	t.Helper()
	name := "bonsai"
	if runtime.GOOS == "windows" {
		name = "bonsai.exe"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("not a real bonsai\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return p
}

// checkCases: one case per word of format.CheckWords.
var checkCases = map[string]func(c *cli){
	"config": func(c *cli) { c.write("bonsai.yaml", "format: bonsai.workspace/1\nid: 0755\n") },
	"lock": func(c *cli) {
		c.link()
		if err := os.Remove(filepath.Join(c.root, ".bonsai", "lock.json")); err != nil {
			c.t.Fatal(err)
		}
	},
	"packs":     func(c *cli) { c.link(); testpack.SetRef(c.t, c.root, c.pack.A, c.pack.B) },
	"changed":   func(c *cli) { c.link(); c.write("demo/guide.md", "mine\n") },
	"missing":   func(c *cli) { c.link(); _ = os.Remove(filepath.Join(c.root, "demo", "guide.md")) },
	"gitignore": func(c *cli) { c.link(); _ = os.Remove(filepath.Join(c.root, ".bonsai", ".gitignore")) },
	"local": func(c *cli) {
		c.link()
		c.write(".bonsai/local/log/s-1.ndjson", "{}\n")
		testpack.Git(c.t, c.root, "add", "-f", ".bonsai/local/log/s-1.ndjson")
	},
	"format0": func(c *cli) {
		c.write("work/tasks/T-0901-old.md", "---\nid: T-0901\ntitle: An old task\nstatus: done\n---\n")
		c.link()
		c.write("work/tasks/T-0901-old.md", "---\nid: T-0901\ntitle: An old task, changed\nstatus: done\n---\n")
	},
	"format0-new": func(c *cli) { c.link(); c.write("work/tasks/T-0902-new.md", "---\nid: T-0902\nstatus: todo\n---\n") },
	"document":    func(c *cli) { c.link(); c.task("T-0901", "nonsense", "", "") },
	"label": func(c *cli) {
		c.declaring()
		c.task("T-0901", "todo", "light", "\n  bonsai.allows: \"not a list\"")
	},
	"label-twice": func(c *cli) {
		c.declaring()
		src, shas := labelsPack(c.t, c.tmp, "more-labels")
		c.addPack("more-labels", src, shas[0])
	},
	"approve-first": func(c *cli) { c.approveCase() },
	"absolute-path": func(c *cli) {
		c.link()
		c.write("work/runs/R-2026-10-01-T-0901.md", "---\nformat: bonsai.run/1\nid: R-2026-10-01-T-0901\ntask: T-0901\nrole: builder\n"+
			"model: a model\nstarted: null\nfinished: null\noutcome: merged\ncommits: []\nlabels: {}\nnotes: \"wrote /home/example/notes.txt\"\n---\n")
	},
	"block-size": func(c *cli) {
		c.link()
		claude := c.read("CLAUDE.md")
		i := strings.Index(claude, "<!-- bonsai:block end")
		c.write("CLAUDE.md", claude[:i]+strings.Repeat("an added line\n", 40)+claude[i:])
	},
	"memory-index-size": func(c *cli) {
		c.link()
		c.write("work/memory/INDEX.md", "---\nformat: bonsai.memory/1\nid: null\ntitle: Index\nkind: index\nupdated: 2026-10-09\n"+
			"source: null\nlabels: {}\n---\n"+strings.Repeat("- a note\n", 130))
	},
	"memory-note-size": func(c *cli) {
		c.link()
		c.write("work/memory/M-big.md", "---\nformat: bonsai.memory/1\nid: M-big\ntitle: A big note\nkind: project\nupdated: 2026-10-09\n"+
			"source: \"a test\"\nlabels: {}\n---\n"+strings.Repeat("A fact, why, and how to apply it.\n", 150))
	},
	"missing-path": func(c *cli) {
		c.link()
		c.write("CLAUDE.md", c.read("CLAUDE.md")+"\nSee `demo/not-there.md` and @demo/guide.md.\n")
	},
	"settings-rule": func(c *cli) {
		c.link()
		c.write(".claude/settings.local.json", `{"permissions": {"deny": ["Read(", "Edit(work/ledger.json)"]}}`)
	},
	"hooks-off": func(c *cli) { c.link(); c.write(".claude/settings.local.json", `{"disableAllHooks": true}`) },
	"plugin-version": func(c *cli) {
		c.link()
		c.write(".claude/settings.local.json", `{"extraKnownMarketplaces": {"bonsai-demo-0123abcd": {"source": {"source": "settings", `+
			`"name": "bonsai-demo-0123abcd", "plugins": [{"name": "demo-pack", "version": "1.0.0", "source": "./"}]}}}}`)
	},
	"plugin": func(c *cli) {
		c.link()
		other := engine.MarketplaceName("demo", []string{c.pack.B})
		c.write(".claude/settings.local.json", `{"enabledPlugins": {"demo-pack@`+other+`": true}}`)
	},
	"id-changed": func(c *cli) {
		c.link()
		cfg, err := workspace.LoadConfig(c.root)
		if err != nil {
			c.t.Fatal(err)
		}
		c.write("bonsai.yaml", strings.Replace(c.read("bonsai.yaml"), cfg.ID, "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", 1))
	},
	"bonsai-path": func(c *cli) {
		c.link()
		installed := filepath.Join(c.tmp, "installed", "bonsai")
		raw, _ := json.Marshal(map[string]string{"path": filepath.ToSlash(installed), "version": "1.0.0", "sha256": strings.Repeat("0", 64)})
		home, _ := workspace.Home()
		if err := os.WriteFile(filepath.Join(home, engine.InstallFile), raw, 0o644); err != nil {
			c.t.Fatal(err)
		}
		fakeBonsai(c.t, filepath.Join(c.tmp, "elsewhere"))
	},
	"claude-code-old":     func(c *cli) { c.link(); fakeClaude(c.t, "2.1.200 (Claude Code)", nil) },
	"claude-code-unknown": func(c *cli) { c.link(); fakeClaude(c.t, "Claude Code, the latest", nil) },
	"same-id": func(c *cli) {
		c.link()
		c.commit("link")
		copyRoot := filepath.Join(c.tmp, "copy")
		testpack.Git(c.t, c.tmp, "clone", "-q", "--", c.root, copyRoot)
		t := c.t
		t.Chdir(copyRoot)
		if code, out, errOut := c.run("", "update", "--yes", "--allow-exec"); code != 0 {
			t.Fatalf("update in the copy: %d\n%s%s", code, out, errOut)
		}
		t.Chdir(c.root)
	},
	"run-reports": func(c *cli) {
		c.link()
		yaml := c.read("bonsai.yaml")
		re := regexp.MustCompile(`(?m)^(  run:[^\n]*\n    keep_days: )null`)
		if !re.MatchString(yaml) {
			c.t.Fatalf("bonsai.yaml has no generated.run.keep_days:\n%s", yaml)
		}
		c.write("bonsai.yaml", re.ReplaceAllString(yaml, "${1}30"))
		c.write("work/runs/R-2020-01-01-T-0901.md", "---\nformat: bonsai.run/1\nid: R-2020-01-01-T-0901\ntask: T-0901\nrole: builder\n"+
			"model: a model\nstarted: null\nfinished: null\noutcome: merged\ncommits: []\nlabels: {}\n---\n")
	},
	"approve-first-unchecked": func(c *cli) {
		c.approveCase()
		shallow := filepath.Join(c.tmp, "shallow")
		url := filepath.ToSlash(c.root)
		if !strings.HasPrefix(url, "/") {
			url = "/" + url // a Windows path: file:///C:/...
		}
		testpack.Git(c.t, c.tmp, "clone", "-q", "--depth", "1", "--", "file://"+url, shallow)
		c.t.Chdir(shallow)
	},
	"plugin-missing":   func(c *cli) { c.link(); setPlugins(c.t, fakeList{}) },
	"plugin-unchecked": func(c *cli) { c.link(); setPlugins(c.t, fakeList{err: engine.ErrNoClaude}) },
	"cache": func(c *cli) {
		c.link()
		v, err := schema.Decode([]byte(c.read(".bonsai/lock.json")))
		if err != nil {
			c.t.Fatal(err)
		}
		lock := v.(schema.Object)
		packs, _ := lock.Get("packs")
		var old []any
		for _, p := range packs.([]any) {
			var o schema.Object
			for _, m := range p.(schema.Object) {
				switch m.Key {
				case "path":
				case "declares":
					o = append(o, schema.Member{Key: m.Key, Value: schema.Object{}})
				default:
					o = append(o, m)
				}
			}
			old = append(old, o)
		}
		lock[lock.Index("packs")].Value = old
		raw, _ := schema.Encode(lock)
		c.write(".bonsai/lock.json", string(raw))
		home, _ := workspace.Home()
		_ = os.RemoveAll(filepath.Join(home, "cache"))
	},
	"local-unchecked": func(c *cli) { c.link(); c.write(".git/index", "not an index") },
}

func TestCheckTable(t *testing.T) {
	for code := range checkCases {
		if _, ok := format.CheckWord(code); !ok {
			t.Errorf("a case for %s, which is not in format.CheckWords", code)
		}
	}
	for _, w := range format.CheckWords {
		t.Run(w.Word, func(t *testing.T) {
			setup, ok := checkCases[w.Word]
			if !ok {
				t.Fatalf("format.CheckWords has %s, and TestCheckTable no case for it", w.Word)
			}
			fakeClaude(t, "2.1.294 (Claude Code)", nil)
			c := newCLI(t)
			setup(c)
			code, out, _ := c.run("", "check", "--json")
			doc := fits(t, out, "check")
			findings, _ := doc.Get("findings")
			warnings, _ := doc.Get("warnings")
			mine, others := findings.([]any), warnings.([]any)
			if w.Kind == "warning" {
				mine, others = others, mine
			}
			if !hasCode(mine, w.Word) {
				t.Fatalf("check --json has no %s %s:\n%s", w.Kind, w.Word, out)
			}
			if hasCode(others, w.Word) {
				t.Errorf("%s is a %s, but check lists it in the other list too:\n%s", w.Word, w.Kind, out)
			}
			wantExit := 0
			if len(findings.([]any)) > 0 {
				wantExit = 1
			}
			if code != wantExit {
				t.Errorf("exit %d with %d findings (warnings never change it)", code, len(findings.([]any)))
			}
			for _, list := range []any{findings, warnings} {
				for _, f := range list.([]any) {
					fo := f.(schema.Object)
					next, _ := fo.Get("next")
					checkNext(t, fo.String("code"), next.(schema.Object).String("do"))
				}
			}
			if textCode, text, _ := c.run("", "check"); textCode != code || !strings.Contains(text, "next: ") {
				t.Errorf("check's text exits %d, its --json %d:\n%s", textCode, code, text)
			}
		})
	}
}

func hasCode(list []any, code string) bool {
	for _, f := range list {
		if f.(schema.Object).String("code") == code {
			return true
		}
	}
	return false
}

var (
	runCommand = regexp.MustCompile(`run: ([^;]+)`)
	bonsaiMention = regexp.MustCompile(`(?:^|[\s(:])bonsai ([a-z][a-z-]*)\b`)
	gitWords   = map[string]bool{"checkout": true, "rm": true, "fetch": true, "status": true}
	claudeArgs = map[string]bool{"update": true, "--version": true, "plugin uninstall": true, "plugin list": true}
)

// checkNext holds a next step to the commands Bonsai has: each "run: <command>" is one it can run as written, and each
// "bonsai <word>" names a word in the registry.
func checkNext(t *testing.T, code, do string) {
	t.Helper()
	if strings.TrimSpace(do) == "" {
		t.Errorf("%s: an empty next step", code)
		return
	}
	for _, m := range bonsaiMention.FindAllStringSubmatch(do, -1) {
		if _, ok := words[m[1]]; !ok {
			t.Errorf("%s: the next step %q names bonsai %s, which is not a word Bonsai has", code, do, m[1])
		}
	}
	for _, m := range runCommand.FindAllStringSubmatch(do, -1) {
		if err := runnable(strings.TrimSpace(m[1])); err != nil {
			t.Errorf("%s: the next step %q: %v", code, do, err)
		}
	}
}

// runnable reports why a command a next step names would not run as written, nil when it would.
func runnable(cmd string) error {
	args := shellSplit(cmd)
	if len(args) < 2 {
		return errors.New("not a command: " + strconv.Quote(cmd))
	}
	switch args[0] {
	case "git":
		if !gitWords[args[1]] {
			return errors.New("git " + args[1] + " is not one of the git commands a next step names")
		}
		return nil
	case "claude":
		if claudeArgs[args[1]] || (len(args) > 2 && claudeArgs[args[1]+" "+args[2]]) {
			return nil
		}
		return errors.New("claude " + args[1] + " is not one of the claude commands a next step names")
	case "bonsai":
	default:
		return errors.New(strconv.Quote(cmd) + " is not a bonsai, git or claude command")
	}
	w, ok := words[args[1]]
	if !ok {
		return errors.New("bonsai " + args[1] + " is not a word Bonsai has")
	}
	for i := 2; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "--") {
			return errors.New("bonsai " + w.Name + " takes no " + strconv.Quote(a))
		}
		name := a
		if j := strings.IndexByte(a, '='); j > 0 {
			name = a[:j]
		}
		f := w.flag(name)
		switch {
		case f == nil:
			return errors.New("bonsai " + w.Name + " takes no " + name)
		case f.Later != "":
			return errors.New("bonsai " + w.Name + " " + name + " is not built yet (" + f.Later + ")")
		case f.Value != "" && !strings.Contains(a, "="):
			i++
			if i >= len(args) {
				return errors.New("bonsai " + w.Name + " " + name + " needs a value")
			}
		}
	}
	return nil
}

// shellSplit splits a command line into words: spaces apart, single quotes (ShellArg's) kept together.
func shellSplit(s string) []string {
	var out []string
	var cur strings.Builder
	in, quoted := false, false
	for i := 0; i < len(s); i++ {
		switch ch := s[i]; {
		case ch == '\'' && quoted && i+1 < len(s) && s[i+1] == '\'':
			cur.WriteByte('\'')
			i++
		case ch == '\'':
			quoted, in = !quoted, true
		case ch == ' ' && !quoted:
			if in {
				out = append(out, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteByte(ch)
			in = true
		}
	}
	if in {
		out = append(out, cur.String())
	}
	return out
}

// The next-step checker itself: it passes the commands Bonsai has and fails the ones it does not.
func TestCheckNextCommands(t *testing.T) {
	for _, ok := range []string{"bonsai update --yes --adopt CLAUDE.md", "bonsai check --schema bonsai.task", "git checkout -- 'a b.md'",
		"claude plugin uninstall x@y --scope local", "bonsai init --yes", "claude update"} {
		if err := runnable(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{"bonsai frobnicate", "bonsai update --frob", "bonsai check --write", "bonsai update now", "rm -rf x",
		"git push", "bonsai check --schema"} {
		if err := runnable(bad); err == nil {
			t.Errorf("%s: passed", bad)
		}
	}
}
