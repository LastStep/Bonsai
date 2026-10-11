package guard

// The delete check's table (plan-5 5.3.3 note 7): command lines written fresh, each with its answer, run as Go tests
// on both sides. The guard's function that judges a path (5.3.2's, not built yet) is stood in for by delStub: the
// floor's paths (bonsai.yaml, the lock, .bonsai/local/ and the tables, .git, the two settings files at the top;
// bonsai.yaml and .bonsai/ at any depth), two protected globs and a protected folder, another linked project beside
// this one (only its top folder and floor folders guarded, 5.3.2 rule 8), and Bonsai's home, all of it guarded.

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// delStub stands in for the guard's function (5.3.2) in the delete check's tests.
type delStub struct {
	own, other, home string
	grants           []string // the globs a running task grants
	calls            int
	rootReads        int      // judgedRoots calls: made only once a delete needs them
	commands         []string // ownCommand's calls: the folder and the words
	steps            []string // personStep's calls: what each line did
}

// delStubGlobs are the own project's guarded globs, as delRoots gives them.
var delStubGlobs = []string{"bonsai.yaml", ".bonsai/lock.json", ".bonsai/local/", ".bonsai/.gitignore",
	".bonsai/tasks.md", ".bonsai/sessions.md", ".git", ".claude/settings.json", ".claude/settings.local.json",
	"**/bonsai.yaml", "**/.bonsai/", "docs/secret/**", "config/*.key", "notes/"}

func (s *delStub) judgedRoots() []judgedRoot {
	s.rootReads++
	return []judgedRoot{
		{Dir: s.own, Globs: delStubGlobs, Own: true},
		{Dir: s.other, Globs: []string{".git", ".bonsai", ".claude"}},
		{Dir: s.home, Whole: true},
	}
}

// ownCommand stands in for commands.go's in a project in agents mode: --allow-exec and unlink refused, settings set
// and labels attach or detach refused, any other word allowed. It counts the calls and keeps the last.
func (s *delStub) ownCommand(dir string, argv []string) Decision {
	s.commands = append(s.commands, dir+" "+strings.Join(argv, " "))
	for _, a := range argv[1:] {
		if strings.HasPrefix(strings.TrimLeft(strings.ToLower(a), "-"), "allow-exec") && strings.HasPrefix(a, "-") || a == "unlink" {
			return deny(RulePersonCommand, "bonsai refused: a person's command", personNext)
		}
		if a == "set" || a == "attach" || a == "detach" {
			return deny(RuleHomeCommand, "bonsai refused: it writes this machine's settings", personNext)
		}
	}
	return allow(RuleBonsaiCommand, "")
}

// personStep stands in for commands.go's in agents mode: every person's step refused. It keeps what each said.
func (s *delStub) personStep(tool, what string) Decision {
	s.steps = append(s.steps, what)
	return deny(RulePersonCommand, tool+" refused: the command line "+what, personNext)
}

func (s *delStub) mayDelete(dir, p string) Decision {
	s.calls++
	for _, root := range []string{s.home, s.own, s.other} {
		rels, err := projectPaths(root, dir, p)
		if err != nil {
			return deny(RuleBadPath, "the path cannot be named", "name it plainly")
		}
		if root == s.home {
			if len(rels) > 0 || filepath.Clean(delAbs(dir, p)) == root {
				return deny("bonsai-files", "it is in Bonsai's home, which only Bonsai writes", "leave it as it is")
			}
			continue
		}
		fold := caseInsensitive(root)
		for _, rel := range rels {
			if rule := s.rule(root, rel, fold); rule != "" {
				return deny(rule, "it is guarded ("+rule+")", "leave it as it is and tell the person")
			}
		}
	}
	return allow(RuleNotProtected, "")
}

// rule gives the stub's rule for a project-relative path, "" for a free one.
func (s *delStub) rule(root, rel string, fold bool) string {
	low := rel
	if fold {
		low = strings.ToLower(rel)
	}
	segs := strings.Split(low, "/")
	if root == s.other {
		switch segs[0] {
		case ".git", ".bonsai", ".claude":
			return "other-project"
		}
		return ""
	}
	rule := ""
	switch {
	case segs[0] == ".git":
		return "floor-never"
	case low == ".claude/settings.json" || low == ".claude/settings.local.json":
		return "floor-never"
	case segs[0] == ".bonsai" && len(segs) > 1 && (segs[1] == "local" || segs[1] == ".gitignore" || segs[1] == "tasks.md" || segs[1] == "sessions.md"):
		return "bonsai-files"
	case low == foldIf(".bonsai/STATE.md", fold):
		return ""
	case low == "bonsai.yaml" || segs[0] == ".bonsai" && len(segs) > 1:
		rule = RulePersonOnly
	case segs[len(segs)-1] == "bonsai.yaml":
		rule = RulePersonOnly
	}
	if rule == "" {
		for k := 1; k < len(segs); k++ {
			if segs[k] == ".bonsai" {
				rule = RulePersonOnly
			}
		}
	}
	if rule == "" {
		for _, g := range []string{"docs/secret/**", "config/*.key", "notes/"} {
			if matchGlob(g, rel, fold) {
				rule = RuleProtected
			}
		}
	}
	if rule != "" {
		for _, g := range s.grants {
			if matchGlob(g, rel, fold) {
				return ""
			}
		}
	}
	return rule
}

// delFixture makes the test's folders: the project, another linked project beside it, and a home holding Bonsai's.
func delFixture(t testing.TB) *delStub {
	t.Helper()
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := &delStub{own: filepath.Join(top, "proj"), other: filepath.Join(top, "other"), home: filepath.Join(top, "home", ".bonsai")}
	files := []string{
		"proj/bonsai.yaml", "proj/.bonsai/lock.json", "proj/.bonsai/STATE.md", "proj/.bonsai/local/log/s-1.ndjson",
		"proj/.bonsai/tasks.md", "proj/.git/HEAD", "proj/.git/objects/.keep", "proj/.git/refs/.keep",
		"proj/plant/.git/.keep", "proj/plant/sub/x.txt", "proj/.claude/settings.json", "proj/.claude/notes.md",
		"proj/docs/secret/plan.md", "proj/docs/readme.md", "proj/config/app.key", "proj/config/app.yaml",
		"proj/notes/a.md", "proj/src/main.go", "proj/build/out/x.o", "proj/junk/one.txt", "proj/junk/two.txt",
		"proj/node_modules/pkg/index.js", "proj/node_modules/pkg/.git/HEAD", "proj/sub/bonsai.yaml", "proj/sub/src/x.go",
		"proj/vendor/lib/.git/HEAD", "proj/file.txt", "proj/a.txt", "proj/x.yaml",
		"other/bonsai.yaml", "other/.git/HEAD", "other/src/x.go",
		"home/.bonsai/workspaces/r-0/settings.json", "home/notes.txt",
	}
	for _, f := range files {
		p := filepath.Join(top, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	home := filepath.Join(top, "home")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return s
}

// delRow is one line of the table.
type delRow struct {
	tool string // Bash, Monitor or PowerShell
	dir  string // the payload's folder, relative to the project ("" for its top)
	line string
	want string // the rule
	only string // "windows" or "unix" when the row holds on one system only, with why in its line's comment
}

const (
	tB = "Bash"
	tM = "Monitor"
	tP = "PowerShell"
)

// delTable is the delete check's table. Its rows, by rule, are counted by TestShellTable.
var delTable = []delRow{
	// --- Ordinary work: must pass.
	{tB, "", "rm -rf node_modules", RuleDeleteNamed, ""},
	{tB, "", "rm file.txt", RuleDeleteNamed, ""},
	{tB, "", "git rm src/main.go", RuleDeleteNamed, ""},
	{tB, "", "mv a.txt b.txt", RuleDeleteNamed, ""},
	{tB, "", "git stash", RuleShell, ""},
	{tB, "", "git clean -n", RuleShell, ""},
	{tB, "", "git clean -nd", RuleShell, ""},
	{tB, "", "git clean -fdx --dry-run", RuleShell, ""},
	{tB, "", "git clean -fd build", RuleDeleteNamed, ""},
	{tB, "", "rm -rf build dist", RuleDeleteNamed, ""},
	{tB, "", "rm -f junk/one.txt junk/two.txt", RuleDeleteNamed, ""},
	{tB, "", "rm -rf ./build", RuleDeleteNamed, ""},
	{tB, "", "rm -r build/", RuleDeleteNamed, ""},
	{tB, "", "mv src/main.go src/app.go", RuleDeleteNamed, ""},
	{tB, "", "git mv src/main.go src/app.go", RuleDeleteNamed, ""},
	{tB, "", "rmdir build/out", RuleDeleteNamed, ""},
	{tB, "", "rm -rf /tmp/bonsai-scratch-xyz", RuleDeleteNamed, ""},
	{tB, "", "ls -la && git status", RuleShell, ""},
	{tB, "", "go test ./... && go vet ./...", RuleShell, ""},
	{tB, "", `git commit -m "drop the rm -rf junk/* step"`, RuleShell, ""},
	{tB, "", "git commit -m \"$(cat <<'EOF'\nDelete junk with rm -rf junk/* and git clean -fdx.\nIt's done.\nEOF\n)\"", RuleShell, ""},
	{tB, "", "echo 'rm -rf /' > notes.txt", RuleShell, ""},
	{tB, "", "cat > clean.sh <<'EOF'\nrm -rf junk/*\nEOF\nchmod +x clean.sh", RuleShell, ""},
	{tB, "", "git stash -u", RuleShell, ""},
	{tB, "", "git stash pop", RuleShell, ""},
	{tB, "", `git stash push -m "wip: -a test"`, RuleShell, ""},
	{tB, "", "find . -name '*.go' -newer go.mod", RuleShell, ""},
	{tB, "", "find build -name '*.o' -delete", RuleDeleteNamed, ""},
	{tB, "", "rsync -a src/ backup/", RuleShell, ""},
	{tB, "", "rsync -a --delete src/ backup/", RuleDeleteNamed, ""},
	{tB, "", "rsync -a --delay-updates src/ backup/", RuleShell, ""},
	{tB, "", "rm -rf node_modules && npm install", RuleDeleteNamed, ""},
	{tB, "", "cd build && rm -rf out", RuleDeleteNamed, ""},
	{tB, "", "(cd build && rm -rf out) && ls", RuleDeleteNamed, ""},
	{tB, "", "rm -rf vendor/lib/.git", RuleDeleteNamed, ""},
	{tB, "", "rm .bonsai/STATE.md", RuleDeleteNamed, ""},
	{tB, "", "mv *.png assets/", RuleDeleteNamed, ""},
	{tB, "", "git rm -r --cached build", RuleDeleteNamed, ""},
	{tB, "", "docker run --rm -v x:/y image", RuleShell, ""},
	{tB, "", "npm run clean", RuleShell, ""},
	{tB, "", "rm -- -weird-name", RuleDeleteNamed, ""},
	{tB, "", "unlink file.txt", RuleDeleteNamed, ""},
	{tB, "", "rm -rf {build,dist}", RuleDeleteNamed, ""},
	{tB, "", "find . -type f | xargs grep -l TODO", RuleShell, ""},
	{tB, "", "git -C sub clean -fd src", RuleDeleteNamed, ""},
	{tB, "", "for f in a b; do echo $f; done", RuleShell, ""},
	{tB, "", "if [ -d build ]; then rm -rf build; fi", RuleDeleteNamed, ""},
	{tB, "", "test -d build && rm -rf build || true", RuleDeleteNamed, ""},
	{tB, "", "rm -rf build 2>/dev/null; mkdir -p build", RuleDeleteNamed, ""},
	{tB, "", `rm -rf "build" 'dist'`, RuleDeleteNamed, ""},
	{tB, "", "cd src; cd ..; rm -rf build", RuleDeleteNamed, ""},
	{tB, "", "pushd build && rm -rf out && popd && rm -rf dist", RuleDeleteNamed, ""},
	{tB, "", "f() { rm -rf build; }; f", RuleDeleteNamed, ""},
	{tB, "", "timeout 60 go test ./...", RuleShell, ""},
	{tB, "", "command -v rm", RuleShell, ""},
	{tB, "", "robocopy src backup /MIR", RuleDeleteNamed, ""},
	{tB, "", "robocopy src backup /MIR /L", RuleShell, ""},
	{tB, "", "rm -rf ../proj-scratch", RuleDeleteNamed, ""},
	{tB, "", "rm -rf ../other/src", RuleDeleteNamed, ""},
	{tB, "", "rm -rf ~/notes-old", RuleDeleteNamed, ""},
	{tB, "", "cd nowhere && rm -rf .git", RuleDeleteNamed, ""},
	{tM, "", "rm -rf build && npm run build -- --watch", RuleDeleteNamed, ""},
	{tP, "", "Remove-Item -Recurse -Force node_modules", RuleDeleteNamed, ""},
	{tP, "", "Remove-Item file.txt", RuleDeleteNamed, ""},
	{tP, "", "Move-Item a.txt b.txt", RuleDeleteNamed, ""},
	{tP, "", "Rename-Item a.txt b.txt", RuleDeleteNamed, ""},
	{tP, "", "Get-ChildItem -Recurse | Select-Object Name", RuleShell, ""},
	{tP, "", "Remove-Item -LiteralPath 'junk/*'", RuleDeleteNamed, ""},
	{tP, "", "rm build -Recurse", RuleDeleteNamed, ""},
	{tP, "", `Remove-Item .\build -Recurse -Force -ErrorAction SilentlyContinue`, RuleDeleteNamed, ""},
	{tP, "", "Remove-Item junk/one.txt, junk/two.txt", RuleDeleteNamed, ""},
	{tP, "", "Remove-Item -Path build -Recurse -Include *.o", RuleDeleteNamed, ""},
	{tP, "", "Remove-Item -Recurse -WhatIf *", RuleShell, ""},
	{tP, "", "if (Test-Path build) { Remove-Item -Recurse -Force build }", RuleDeleteNamed, ""},
	{tP, "", `cmd /c "rd /s /q build"`, RuleDeleteNamed, ""},
	{tP, "", `cmd /c del junk\one.txt`, RuleDeleteNamed, ""},
	{tP, "", "git clean -fd build; git status", RuleDeleteNamed, ""},
	{tP, "", "Write-Output 'Remove-Item -Recurse junk/*'", RuleShell, ""},

	// --- Rule 1, the splitter, and rule 2, the delete words: every spelling reaches the program.
	{tB, "", "rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tM, "", "rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "/bin/rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "RM.EXE -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", `\rm -rf junk/*`, RuleDeleteUnnamed, ""},
	{tB, "", `"rm" -rf junk/*`, RuleDeleteUnnamed, ""},
	{tB, "", "r''m -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", `$'\x72\x6d' -rf junk/*`, RuleDeleteUnnamed, ""},
	{tB, "", "sudo rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "sudo -u root -E rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "env FOO=1 BAR=2 rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "FOO=1 rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "command rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "exec rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "nice -n 5 rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "nice -10 rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "nohup rm -rf junk/* &", RuleDeleteUnnamed, ""},
	{tB, "", "time rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "timeout -s KILL 5 rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "env -S 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "doas -u root rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "builtin command rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "ionice -c 3 rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "setsid -f rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "stdbuf -o0 rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "zsh -c 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "dash -c 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "chronic rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "unbuffer rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "fakeroot -- rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", `ls | xargs sh -c 'rm "$1"' _`, RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item (Get-Item x).FullName -Recurse", RuleDeleteUnnamed, ""},
	// A computed operand outside a recursive or bulk delete is judged as written (note 3 refuses a variable only
	// there), so this one is not seen: the check's known limit, kept visible.
	{tP, "", "Remove-Item (Get-Item bonsai.yaml).Name", RuleDeleteNamed, ""},
	{tB, "", "ls; rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "ls && rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "false || rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "ls | rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "ls\nrm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "ls \\\n && rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "echo $(rm -rf junk/*)", RuleDeleteUnnamed, ""},
	{tB, "", "echo `rm -rf junk/*`", RuleDeleteUnnamed, ""},
	{tB, "", "(rm -rf junk/*)", RuleDeleteUnnamed, ""},
	{tB, "", "{ rm -rf junk/*; }", RuleDeleteUnnamed, ""},
	{tB, "", "f() { rm -rf junk/*; }; f", RuleDeleteUnnamed, ""},
	{tB, "", "diff <(rm -rf junk/*) x", RuleDeleteUnnamed, ""},
	{tB, "", "x=$(rm -rf junk/*)", RuleDeleteUnnamed, ""},
	{tB, "", `echo "${X:-$(rm -rf junk/*)}"`, RuleDeleteUnnamed, ""},
	{tB, "", "bash -c 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", `sh -c "rm -rf junk/*"`, RuleDeleteUnnamed, ""},
	{tB, "", "bash -lc 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "bash -o pipefail -c 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", `bash -c "bash -c 'rm -rf junk/*'"`, RuleDeleteUnnamed, ""},
	{tB, "", "eval 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "eval rm -rf 'junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "bash <<EOF\nrm -rf junk/*\nEOF", RuleDeleteUnnamed, ""},
	{tB, "", "bash -s <<'EOF'\nrm -rf junk/*\nEOF", RuleDeleteUnnamed, ""},
	{tB, "", "sh <<< 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "echo 'rm -rf junk/*' | bash", RuleDeleteUnnamed, ""},
	{tB, "", "cat <<EOF | sh\nrm -rf junk/*\nEOF", RuleDeleteUnnamed, ""},
	{tB, "", "cat <<EOF\n$(rm -rf junk/*)\nEOF", RuleDeleteUnnamed, ""},
	{tB, "", `cmd /c "rd /s /q ."`, RuleDeleteUnnamed, ""},
	{tB, "", "cmd.exe //c del /s *.log", RuleDeleteUnnamed, ""},
	{tB, "", "cmd /c rd/s/q %TEMP%", RuleDeleteUnnamed, ""},
	{tB, "", `powershell -Command "Remove-Item -Recurse junk/*"`, RuleDeleteUnnamed, ""},
	{tB, "", "pwsh -NoProfile -c 'Remove-Item -r junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "wsl rm -rf 'junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "busybox rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "$RM -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", `"$(which rm)" -r junk/*`, RuleDeleteUnnamed, ""},
	{tB, "", "if true; then rm -rf junk/*; fi", RuleDeleteUnnamed, ""},
	{tB, "", "while false; do rm -rf junk/*; done", RuleDeleteUnnamed, ""},
	{tB, "", "! rm -rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "# a comment\nrm -rf junk/* # it's gone", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item -Recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", `Remove-Item junk\* -Recurse`, RuleDeleteUnnamed, ""},
	{tP, "", "rm -r junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "ri -Rec junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "del -recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "erase junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "rd -r junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "rmdir -R junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "REMOVE-ITEM -RECURSE junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "Microsoft.PowerShell.Management\\Remove-Item -Recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item \u2013Recurse $p", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item\u00a0-Recurse\u00a0junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item -Recurse:$true $p", RuleDeleteUnnamed, ""},
	{tP, "", "$null = Remove-Item -Recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "$r=Remove-Item -Recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "& { Remove-Item -Recurse junk/* }", RuleDeleteUnnamed, ""},
	{tP, "", "& 'Remove-Item' -Recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "[void](Remove-Item -Recurse junk/*)", RuleDeleteUnnamed, ""},
	{tP, "", `powershell -Command "Remove-Item -Recurse junk/*"`, RuleDeleteUnnamed, ""},
	{tP, "", "pwsh -c 'Remove-Item -r junk/*'", RuleDeleteUnnamed, ""},
	{tP, "", "powershell -NoProfile -EncodedCommand " + delEncode("Remove-Item -Recurse junk/*"), RuleDeleteUnnamed, ""},
	{tP, "", "@'\nRemove-Item -Recurse junk/*\n'@ | powershell -Command -", RuleDeleteUnnamed, ""},
	{tP, "", "'Remove-Item -Recurse junk/*' | iex", RuleDeleteUnnamed, ""},
	{tP, "", "Invoke-Expression 'rm -r junk/*'", RuleDeleteUnnamed, ""},
	{tP, "", `cmd /c "rd /s /q ."`, RuleDeleteUnnamed, ""},
	{tP, "", "bash -c 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tP, "", "wsl -e rm -rf junk/x junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "rm.exe -rf junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "git clean -fdx", RuleDeleteUnnamed, ""},
	{tP, "", "Get-ChildItem | ForEach-Object { if ($_.Name -like '*.log') { Remove-Item $_ } }", RuleDeleteUnnamed, ""},
	{tP, "", "# Remove-Item is safe here\nRemove-Item -Recurse junk/* # it's gone", RuleDeleteUnnamed, ""},
	{tP, "", "<# a block\ncomment #> Remove-Item -Recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item `\n  -Recurse `\n  junk/*", RuleDeleteUnnamed, ""},

	// --- Rule 3, not naming what it deletes: every form of recursive and bulk, the operand tests, the always-refused.
	{tB, "", "rm -rf *", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf .", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ./", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ..", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf /", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ~", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ~/", RuleDeleteUnnamed, ""},
	{tB, "", `rm -rf ""`, RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf build/.", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf build/..", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf /mnt/c", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf /c/", RuleDeleteUnnamed, ""},
	{tB, "", `rm -rf "$DIR"`, RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf $DIR/build", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ${DIR}", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf $(pwd)/build", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf `pwd`", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf %TEMP%", RuleDeleteUnnamed, ""},
	{tB, "", "rm -fR junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm -Rf junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm -vfr junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm -r -f junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm --recursive junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm --rec --force junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm --r junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm junk/* -rf", RuleDeleteUnnamed, ""},
	{tB, "", "rm *.log", RuleDeleteUnnamed, ""},
	{tB, "", "rm junk/file?.txt", RuleDeleteUnnamed, ""},
	{tB, "", "rm junk/[ab].txt", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ./*", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf junk/ build/ *", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf {junk,build}/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf junk/{1..3}", RuleDeleteUnnamed, ""},
	{tB, "", `rm -rf $'junk/\x2a'`, RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf !(keep)", RuleDeleteUnnamed, ""},
	{tB, "", "ls | xargs rm", RuleDeleteUnnamed, ""},
	{tB, "", "find . -name x | xargs rm -f", RuleDeleteUnnamed, ""},
	{tB, "", "xargs -I{} rm -rf {} < list.txt", RuleDeleteUnnamed, ""},
	{tB, "", "xargs -0 -n 1 mv -t old/ < list.txt", RuleDeleteUnnamed, ""},
	{tB, "", "find . -delete", RuleDeleteUnnamed, ""},
	{tB, "", "find -delete", RuleDeleteUnnamed, ""},
	{tB, "", "find . -name '*.log' -delete", RuleDeleteUnnamed, ""},
	{tB, "", `find . -exec rm {} \;`, RuleDeleteUnnamed, ""},
	{tB, "", "find . -type d -execdir rm -rf {} +", RuleDeleteUnnamed, ""},
	{tB, "", `find . -ok rm {} \;`, RuleDeleteUnnamed, ""},
	{tB, "", `find . -exec sudo /bin/rm -f {} \;`, RuleDeleteUnnamed, ""},
	{tB, "", "find junk/* -delete", RuleDeleteUnnamed, ""},
	{tB, "", "find $DIR -delete", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -fd", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -f .", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -fdx build", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -fdX", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -xfd", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -d -f -x -- build", RuleDeleteUnnamed, ""},
	{tB, "", "git -C sub clean -fdx", RuleDeleteUnnamed, ""},
	{tB, "", "git -c core.quotepath=off clean -ffdx", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -fd :/", RuleDeleteUnnamed, ""},
	{tB, "", "git stash --all", RuleDeleteUnnamed, ""},
	{tB, "", "git stash -a", RuleDeleteUnnamed, ""},
	{tB, "", "git stash push --all", RuleDeleteUnnamed, ""},
	{tB, "", `git stash save -a "x"`, RuleDeleteUnnamed, ""},
	{tB, "", "git stash --al", RuleDeleteUnnamed, ""},
	{tB, "", "git stash push -ka -- src", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf --no-preserve-root build", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf --no-p build", RuleDeleteUnnamed, ""},
	{tB, "", "git rm -r .", RuleDeleteUnnamed, ""},
	{tB, "", "git rm '*.log'", RuleDeleteUnnamed, ""},
	{tB, "", "git rm -r :/", RuleDeleteUnnamed, ""},
	{tB, "", "git rm --pathspec-from-file=list.txt", RuleDeleteUnnamed, ""},
	{tB, "", "rsync -a --delete src/ .", RuleDeleteUnnamed, ""},
	{tB, "", "rsync -a --delete-after src/ ./", RuleDeleteUnnamed, ""},
	{tB, "", "rsync -a --del src/ $DEST", RuleDeleteUnnamed, ""},
	{tB, "", "robocopy src . /MIR", RuleDeleteUnnamed, ""},
	{tB, "", "robocopy src . //PURGE", RuleDeleteUnnamed, ""},
	{tB, "", `cd "$X" && rm -rf build`, RuleDeleteUnnamed, ""},
	{tB, "", "cd $(mktemp -d); rm -rf build", RuleDeleteUnnamed, ""},
	{tB, "", "cd -; rm -rf build", RuleDeleteUnnamed, ""},
	{tB, "", "env -C \"$X\" rm -rf build", RuleDeleteUnnamed, ""},
	{tB, "", `for f in *.log; do rm "$f"; done`, RuleDeleteUnnamed, ""},
	{tB, "", "for d in */; do\n  rm -r \"$d\"\ndone", RuleDeleteUnnamed, ""},
	{tB, "", "while read -r f; do mv \"$f\" old/; done < list.txt", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item -Recurse .", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item -Recurse $env:TEMP\\x", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item -Recurse \"$HOME\\build\"", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item -Recurse -Path (Join-Path $root 'x')", RuleDeleteUnnamed, ""},
	{tP, "", "Get-ChildItem *.log | Remove-Item", RuleDeleteUnnamed, ""},
	{tP, "", "gci build | ri -Recurse", RuleDeleteUnnamed, ""},
	{tP, "", "Get-ChildItem *.txt | Move-Item -Destination old", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item * -Include *.log -Recurse", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item -Recurse:$false junk/*", RuleDeleteUnnamed, ""},
	{tP, "", "Remove-Item @params", RuleDeleteUnnamed, ""},
	{tP, "", "foreach ($f in $files) { Remove-Item $f }", RuleDeleteUnnamed, ""},
	{tP, "", "$files | ForEach-Object { Remove-Item -LiteralPath $_.FullName }", RuleDeleteUnnamed, ""},
	{tP, "", `robocopy src C:\ /MIR`, RuleDeleteUnnamed, ""},
	{tP, "", "cmd /c for /d %i in (*) do rd /s /q %i", RuleDeleteUnnamed, ""},
	{tP, "", "cmd /c del /s /q *.tmp", RuleDeleteUnnamed, ""},
	{tP, "", "cmd /c \"if exist build rd /s /q %CD%\"", RuleDeleteUnnamed, ""},
	{tP, "", "Set-Location $env:TEMP; Remove-Item -Recurse build", RuleDeleteUnnamed, ""},

	// --- Rule 4, a guarded path: deleted, moved, or held by a folder deleted or moved.
	{tB, "", "rm bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "rm -f ./bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "rm .bonsai/lock.json", RuleDeleteProtected, ""},
	{tB, "", "rm -rf .bonsai", RuleDeleteProtected, ""},
	{tB, "", "rm -rf .bonsai/local", RuleDeleteProtected, ""},
	{tB, "", "rm -rf .git", RuleDeleteProtected, ""},
	{tB, "", "rm -rf .git/hooks", RuleDeleteProtected, ""},
	{tB, "", "rm -rf .claude", RuleDeleteProtected, ""},
	{tB, "", "rm .claude/settings.json", RuleDeleteProtected, ""},
	{tB, "", "rm -rf docs/secret", RuleDeleteProtected, ""},
	{tB, "", "rm -rf docs", RuleDeleteProtected, ""},
	{tB, "", "rm config/app.key", RuleDeleteProtected, ""},
	{tB, "", "rm -rf config", RuleDeleteProtected, ""},
	{tB, "", "rm notes/a.md", RuleDeleteProtected, ""},
	{tB, "", "rm -rf notes", RuleDeleteProtected, ""},
	{tB, "", "rm sub/bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "rm -rf sub/.bonsai", RuleDeleteProtected, ""},
	{tB, "", "mv bonsai.yaml bonsai.yaml.bak", RuleDeleteProtected, ""},
	{tB, "", "mv x.yaml bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "mv settings.json .claude/", RuleDeleteProtected, ""},
	{tB, "", "mv -t .claude settings.json", RuleDeleteProtected, ""},
	{tB, "", "mv .git .git-old", RuleDeleteProtected, ""},
	{tB, "", "mv * src/", RuleDeleteProtected, ""},
	{tB, "", "mv docs/* archive/", RuleDeleteProtected, ""},
	{tB, "", "mv .claude old-claude", RuleDeleteProtected, ""},
	{tB, "", "git rm bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "git rm -r --cached .bonsai", RuleDeleteProtected, ""},
	{tB, "", "git mv bonsai.yaml x.yaml", RuleDeleteProtected, ""},
	{tB, "", "git clean -fd .bonsai", RuleDeleteProtected, ""},
	{tB, "", "shred -u bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "unlink .bonsai/lock.json", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ../proj", RuleDeleteProtected, ""},
	{tB, "src", "rm -rf ../../proj", RuleDeleteProtected, ""},
	{tB, "", "cd .. && rm -rf proj", RuleDeleteProtected, ""},
	{tB, "", "cd build; rm -f ../bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "cd nowhere; rm -rf .git", RuleDeleteProtected, ""},
	{tB, "src", "rm ../bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "find .git -delete", RuleDeleteProtected, ""},
	{tB, "", "find .bonsai -name '*.json' -delete", RuleDeleteProtected, ""},
	{tB, "", "rsync -a --delete empty/ .git/", RuleDeleteProtected, ""},
	{tB, "", "robocopy empty .git /MIR", RuleDeleteProtected, ""},
	{tB, "", "rm -rf {bonsai.yaml,x}", RuleDeleteProtected, ""},
	{tB, "", "rm {bonsai.yaml,}", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ~/.bonsai", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ~/.bonsai/workspaces", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ../home", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ../other", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ../other/.git", RuleDeleteProtected, ""},
	{tB, "", "git -C .. rm -r proj/.git", RuleDeleteProtected, ""},
	{tB, "", "bash -c 'rm bonsai.yaml'", RuleDeleteProtected, ""},
	{tB, "", "rm -rf build bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "rsync -a --remove-source-files bonsai.yaml backup/", RuleDeleteProtected, ""},
	{tB, "", "robocopy .git backup /MOVE", RuleDeleteProtected, ""},
	{tB, "", "$RM -rf .git", RuleDeleteProtected, ""},
	{tP, "", "Remove-Item bonsai.yaml", RuleDeleteProtected, ""},
	{tP, "", "Remove-Item -Recurse .git -Force", RuleDeleteProtected, ""},
	{tP, "", `Remove-Item -LiteralPath .claude\settings.json`, RuleDeleteProtected, ""},
	{tP, "", "Remove-Item .claude", RuleDeleteProtected, ""},
	{tP, "", "Move-Item bonsai.yaml x.yaml", RuleDeleteProtected, ""},
	{tP, "", "Move-Item -Path x.yaml -Destination bonsai.yaml", RuleDeleteProtected, ""},
	{tP, "", "Rename-Item bonsai.yaml x.yaml", RuleDeleteProtected, ""},
	{tP, "", "Rename-Item -Path x.yaml -NewName bonsai.yaml", RuleDeleteProtected, ""},
	{tP, "", "Remove-Item FileSystem::bonsai.yaml", RuleDeleteProtected, ""},
	{tP, "", "Remove-Item -Path .bonsai\\lock.json, x.txt", RuleDeleteProtected, ""},
	{tP, "", `cmd /c del bonsai.yaml`, RuleDeleteProtected, ""},
	{tP, "", `cmd /c rd /s /q .git`, RuleDeleteProtected, ""},
	{tP, "", `cmd /c "move bonsai.yaml x"`, RuleDeleteProtected, ""},
	{tP, "", `cmd /c ren bonsai.yaml x.yaml`, RuleDeleteProtected, ""},
	{tP, "", `cmd /c del .claude`, RuleDeleteProtected, ""},
	{tP, "", "Set-Location src; Remove-Item ..\\bonsai.yaml", RuleDeleteProtected, ""},
	{tP, "", "Push-Location ..; Remove-Item -Recurse proj; Pop-Location", RuleDeleteProtected, ""},

	// --- Lines written to slip past the splitter (the builder's own break-it), each refused or rightly allowed.
	{tB, "", "{rm,-rf,junk/*}", RuleDeleteUnnamed, ""},
	{tB, "", "rm${IFS}-rf${IFS}junk/*", RuleDeleteUnnamed, ""},
	{tB, "", "rm${IFS}bonsai.yaml", RuleDeleteUnnamed, ""},
	{tB, "", "bash -c \"$(cat <<'EOF'\nrm -rf junk/*\nEOF\n)\"", RuleDeleteUnnamed, ""},
	{tB, "", "bash <(echo 'rm -rf junk/*')", RuleDeleteUnnamed, ""},
	{tB, "", "source <(echo 'rm -rf junk/*')", RuleDeleteUnnamed, ""},
	{tB, "", `eval "$(echo rm -rf junk/*)"`, RuleDeleteUnnamed, ""},
	{tB, "", "trap 'rm -rf junk/*' EXIT", RuleDeleteUnnamed, ""},
	{tB, "", "su -c 'rm -rf junk/*' root", RuleDeleteUnnamed, ""},
	{tB, "", "watch -n 1 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ~home", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ~+", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ~home/.bonsai", RuleDeleteProtected, ""},
	{tB, "", `rm "$(echo bonsai.yaml)"`, RuleDeleteProtected, ""},
	{tB, "", "rm `echo bonsai.yaml`", RuleDeleteProtected, ""},
	{tB, "", "$(echo rm) bonsai.yaml", RuleDeleteProtected, ""},
	{tB, "", "$(echo rm -rf junk/*)", RuleDeleteUnnamed, ""},
	{tB, "", "bash <<< \"$(echo rm -rf junk/*)\"", RuleDeleteUnnamed, ""},
	{tB, "", "rm $(echo a.txt bonsai.yaml)", RuleDeleteProtected, ""},
	{tB, "", `rm "$(echo a.txt bonsai.yaml)"`, RuleDeleteNamed, ""},
	{tB, "", "mv */bonsai.yaml old/", RuleDeleteProtected, ""},
	{tB, "", `find build -exec rm -rf {} bonsai.yaml \;`, RuleDeleteProtected, ""},
	{tB, "", "git -C .. stash -a", RuleDeleteUnnamed, ""},
	{tB, "", "mv junk/* .", RuleDeleteNamed, ""},
	{tB, "", "find build -exec rm -rf {} +", RuleDeleteNamed, ""},
	{tB, "", "rm -rf .git.bak .gitignore .claude/agents", RuleDeleteNamed, ""},
	{tB, "", "git stash -- --all", RuleShell, ""},
	{tB, "", "git clean -fde -x build", RuleDeleteNamed, ""},
	{tP, "", `Remove-Item ("bonsai.yaml")`, RuleDeleteProtected, ""},
	{tP, "", "Remove-Item $('bonsai.yaml')", RuleDeleteProtected, ""},
	{tP, "", "Remove-Item -Recurse 'build','.git'", RuleDeleteProtected, ""},
	{tP, "", "Remove-Item -Recurse @('build','.git')", RuleDeleteUnnamed, ""},
	{tP, "", "'rm -r junk/*' | powershell -noprofile -", RuleDeleteUnnamed, ""},
	{tP, "", "& (Get-Command Remove-Item) -Recurse junk/*", RuleDeleteUnnamed, ""},
	{tP, "", `cmd /c "cd .. & rd /s /q proj"`, RuleDeleteProtected, ""},
	{tP, "", "cmd /c start /b rd /s /q .", RuleDeleteUnnamed, ""},
	{tP, "", "cmd /c call rd /s /q .", RuleDeleteUnnamed, ""},
	{tP, "", "cmd /c r^d /s /q .", RuleDeleteUnnamed, ""},
	{tP, "", "cmd /c rd .git /s /q", RuleDeleteProtected, ""},
	{tP, "", `cmd /v:on /c "set X=. & rd /s /q !X!"`, RuleDeleteUnnamed, ""},
	{tP, "", `Move-Item junk\* .`, RuleDeleteNamed, ""},
	{tP, "", "Move-Item -Path build -Destination .git", RuleDeleteProtected, ""},

	// --- Lines ordinary work might write that are refused, each by the plan's rule 3 (the report lists why).
	{tB, "", "git rm -r --cached .", RuleDeleteUnnamed, ""},
	{tB, "", "find . -maxdepth 1 -name '*.tmp' -delete", RuleDeleteUnnamed, ""},
	{tB, "", "rm -f *.pyc", RuleDeleteUnnamed, ""},
	{tB, "", `rm -rf "$TMPDIR/bonsai-x"`, RuleDeleteUnnamed, ""},
	{tB, "", `cd "$(git rev-parse --show-toplevel)" && rm -rf build`, RuleDeleteUnnamed, ""},
	{tP, "", "Get-ChildItem -Filter *.log | Remove-Item", RuleDeleteUnnamed, ""},

	// --- Bonsai's own commands the splitter finds, and CLAUDE_CODE_CHILD_SESSION in a text it splits (the stub reads
	// them as commands.go does in a project in agents mode).
	{tB, "", "bonsai status --json", RuleBonsaiCommand, ""},
	{tB, "", "cd build && /usr/local/bin/bonsai check --write", RuleBonsaiCommand, ""},
	{tB, "", "bonsai update --allow-exec --yes", RulePersonCommand, ""},
	{tB, "", "sudo bonsai unlink --yes", RulePersonCommand, ""},
	{tB, "", "bash -c 'bonsai update --ALLOW-EXEC'", RulePersonCommand, ""},
	{tB, "", "bonsai settings set status_writes=agents", RuleHomeCommand, ""},
	{tP, "", "& 'C:\\Program Files\\Bonsai\\bonsai.exe' labels attach x", RuleHomeCommand, ""},
	{tP, "", `cmd /c "bonsai unlink --yes"`, RulePersonCommand, ""},
	{tP, "", "powershell -EncodedCommand " + delEncode("$env:CLAUDE_CODE_CHILD_SESSION = ''"), RulePersonCommand, ""},
	{tB, "", "bash <<EOF\nunset CLAUDE_CODE_CHILD_SESSION\nEOF", RulePersonCommand, ""},
	{tB, "", "exec -\u00e4c ./x.sh", RulePersonCommand, ""},
	{tB, "", "bonsai \u00e4dd --allow-exec", RulePersonCommand, ""},

	// --- V2's fix round (008b452): git's pathspec magic, the runner words before env -i, the variable's name in any
	// quoting, bash's ~+ and ~-, and bonsai where the check cannot split or find -exec or xargs runs it.
	{tB, "", "git rm :/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "git rm ':(top)config/app.key'", RuleDeleteProtected, ""},
	{tB, "", "git rm :config/app.key", RuleDeleteProtected, ""},
	{tB, "", "git rm ::config/app.key", RuleDeleteProtected, ""},
	{tB, "", "git rm ':(literal)config/app.key'", RuleDeleteProtected, ""},
	{tB, "src", "git rm :/config/app.key", RuleDeleteProtected, ""},
	{tB, "src", "git rm -r :/:docs", RuleDeleteProtected, ""},
	{tB, "", "git mv :/notes/a.md old.md", RuleDeleteProtected, ""},
	{tB, "", "git rm ':(icase)CONFIG/APP.KEY'", RuleDeleteUnnamed, ""},
	{tB, "", "git rm ':(glob)*.yaml'", RuleDeleteUnnamed, ""},
	{tB, "", "git rm ':!src'", RuleDeleteUnnamed, ""},
	{tB, "", "git rm ':^src'", RuleDeleteUnnamed, ""},
	{tB, "", "git rm ':(top,literal)config/app.key'", RuleDeleteUnnamed, ""},
	{tB, "", "git rm ':(attr:x)y'", RuleDeleteUnnamed, ""},
	{tB, "", "git clean -f ':(exclude)src'", RuleDeleteUnnamed, ""},
	{tB, "", "git rm :src/main.go", RuleDeleteNamed, ""},
	{tB, "src", "git rm :/src/main.go", RuleDeleteNamed, ""},
	{tB, "", "git clean -fd :/build", RuleDeleteNamed, ""},
	{tB, "", "timeout 5 env -i ./x.sh", RulePersonCommand, ""},
	{tB, "", "nice -n 5 env -i ./x.sh", RulePersonCommand, ""},
	{tB, "", "setsid -f env - ./x.sh", RulePersonCommand, ""},
	{tB, "", "stdbuf -o0 env --ignore-env ./x.sh", RulePersonCommand, ""},
	{tB, "", "ionice -c 3 doas exec -c ./x.sh", RulePersonCommand, ""},
	{tB, "", "timeout -k 5 10 env -u CLAUDE_CODE_CHILD_SESSION ./x.sh", RulePersonCommand, ""},
	{tB, "", "unset CLAUDE_CODE_CHILD_'SESSION'", RulePersonCommand, ""},
	{tB, "", `unset CLAUDE_CODE_CHILD_SESS\ION`, RulePersonCommand, ""},
	{tB, "", `unset "CLAUDE_CODE_CHILD"_SESSION`, RulePersonCommand, ""},
	{tB, "", "unset CLAUDE_CODE_CHILD_SESSIO{N,}", RulePersonCommand, ""},
	{tB, "", "X=CLAUDE_CODE_CHILD; unset ${X}_SESSION", RulePersonCommand, ""},
	{tB, "", "N=CLAUDE_CODE_CHILD_SESSION; unset $N", RulePersonCommand, ""},
	{tP, "", "[Environment]::SetEnvironmentVariable(('CLAUDE_CODE_CHILD_SESSION'),$null)", RulePersonCommand, ""},
	{tP, "", `[Environment]::SetEnvironmentVariable("CLAUDE_CODE_CHILD" + "_SESSION", $null)`, RulePersonCommand, ""},
	{tP, "", "$n = 'CLAUDE_CODE_CHILD_SESSION'; Remove-Item \"Env:$n\"", RulePersonCommand, ""},
	{tP, "", "Microsoft.PowerShell.Core\\Environment::CLAUDE_CODE_CHILD_SESSION | Clear-Item", RulePersonCommand, ""},
	{tB, "", "echo $CLAUDE_CODE_CHILD_SESSION; printenv CLAUDE_CODE_CHILD_SESSION", RuleShell, ""},
	{tB, "", "grep -rn CLAUDE_CODE_CHILD_SESSION internal/ && X=$CLAUDE_CODE_CHILD_SESSION", RuleShell, ""},
	{tP, "", "if ($env:CLAUDE_CODE_CHILD_SESSION) { 'agent' }; Get-ChildItem Env:CLAUDE_CODE_CHILD_SESSION", RuleShell, ""},
	{tB, "", "cat > notes.md <<'EOF'\nCLAUDE_CODE_CHILD_SESSION is set by Claude Code\nEOF", RuleShell, ""},
	// A comment naming the variable is no word the check reads, so its line fails closed (shell_env.go's known cost).
	{tB, "", "ls # CLAUDE_CODE_CHILD_SESSION", RulePersonCommand, ""},
	{tB, "", "rm ~+/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "rm ~-/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ~-/build", RuleDeleteUnnamed, ""},
	{tB, "", "rm -rf ~+", RuleDeleteUnnamed, ""},
	{tB, "", "cd ~- && rm -rf build", RuleDeleteUnnamed, ""},
	{tB, "", "rm ~+/file.txt", RuleDeleteNamed, ""},
	{tB, "", "bonsai update --allow-exec 'x", RulePersonCommand, ""},
	{tB, "", "bonsai unlink --yes \"", RulePersonCommand, ""},
	{tB, "", "bonsai settings set x 'y", RuleHomeCommand, ""},
	{tP, "", "bonsai labels attach \"x", RuleHomeCommand, ""},
	{tB, "", `find . -name x -exec bonsai unlink \;`, RulePersonCommand, ""},
	{tB, "", "find . -execdir bonsai update --allow-exec {} +", RulePersonCommand, ""},
	{tB, "", "ls | xargs bonsai", RuleBonsaiCommand, ""},
	{tB, "", "echo unlink | xargs -n1 bonsai", RuleBonsaiCommand, ""},
	{tB, "", `find build -exec bonsai status \;`, RuleBonsaiCommand, ""},

	// --- V2b's fix round (8d87bc5): PowerShell's assignment written with its value, env's quoted NAME=value and -S
	// text, ~+//, the folder stack's tildes, a .git git would not take, case-blind pathspecs, and xargs' bonsai words.
	{tP, "", "$env:CLAUDE_CODE_CHILD_SESSION=''", RulePersonCommand, ""},
	{tP, "", `$env:CLAUDE_CODE_CHILD_SESSION=""`, RulePersonCommand, ""},
	{tP, "", "$env:CLAUDE_CODE_CHILD_SESSION='x'", RulePersonCommand, ""},
	{tP, "", "$env:CLAUDE_CODE_CHILD_SESSION+='x'", RulePersonCommand, ""},
	{tP, "", "${env:CLAUDE_CODE_CHILD_SESSION}=''", RulePersonCommand, ""},
	{tP, "", "$Env:claude_code_child_session=''", RulePersonCommand, ""},
	{tP, "", "if ($true) { $env:CLAUDE_CODE_CHILD_SESSION='' }", RulePersonCommand, ""},
	{tP, "", "& { $env:CLAUDE_CODE_CHILD_SESSION='' }", RulePersonCommand, ""},
	{tP, "", "powershell -EncodedCommand " + delEncode("$env:CLAUDE_CODE_CHILD_SESSION=''"), RulePersonCommand, ""},
	{tP, "", "$n='CLAUDE_CODE_CHILD_SESSION'; [Environment]::SetEnvironmentVariable($n,$null)", RulePersonCommand, ""},
	{tP, "", `$n='CLAUDE_CODE_CHILD_SESSION'; Remove-Item "Env:$n"`, RulePersonCommand, ""},
	{tP, "", "$x = $env:CLAUDE_CODE_CHILD_SESSION = ''", RulePersonCommand, ""},
	{tP, "", "$v = $env:CLAUDE_CODE_CHILD_SESSION; Write-Output $v", RuleShell, ""},
	{tP, "", "if ($env:CLAUDE_CODE_CHILD_SESSION -eq '1') { 'agent' }", RuleShell, ""},
	{tB, "", "env 'CLAUDE_CODE_CHILD_SESSION=' ./x.sh", RulePersonCommand, ""},
	{tB, "", `env "CLAUDE_CODE_CHILD_SESSION=" ./x.sh`, RulePersonCommand, ""},
	{tB, "", "env 'FOO=1' ./x.sh", RuleShell, ""},
	{tB, "", "env -S '-i ./x.sh'", RulePersonCommand, ""},
	{tB, "", "env -S'-u CLAUDE_CODE_CHILD_SESSION' ./x.sh", RulePersonCommand, ""},
	{tB, "", "env -S '-u CLAUDE_CODE_CHILD_SESSION' bonsai unlink", RulePersonCommand, ""},
	{tB, "", "env -S '-u X' rm -rf docs", RuleDeleteProtected, ""},
	{tB, "", "env -S'-i' rm -rf build/*", RulePersonCommand, ""},
	{tB, "", "env -S '-C docs' rm secret/plan.md", RuleDeleteProtected, ""},
	{tB, "", "env -S '-u X' rm -rf build", RuleDeleteNamed, ""},
	{tB, "", "rm -rf ~+//docs", RuleDeleteProtected, ""},
	{tB, "", "rm ~+//config/app.key", RuleDeleteProtected, ""},
	{tB, "", "mv ~+//config/app.key x", RuleDeleteProtected, ""},
	{tB, "", "rm ~0/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "rm ~+0/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "rm ~-0/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "rm ~2/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "rm -rf ~1/build", RuleDeleteUnnamed, ""},
	{tB, "", "cd ~1 && rm -rf build", RuleDeleteUnnamed, ""},
	{tB, "", "cd ~0 && rm -rf build", RuleDeleteNamed, ""},
	{tB, "plant/sub", "git rm :/config/app.key", RuleDeleteProtected, ""},
	{tB, "", "git --icase-pathspecs rm config/APP.KEY", RuleDeleteUnnamed, ""},
	{tB, "", "GIT_ICASE_PATHSPECS=1 git rm Config/app.key", RuleDeleteUnnamed, ""},
	{tB, "", "export GIT_ICASE_PATHSPECS=1; git rm config/app.key", RuleDeleteUnnamed, ""},
	{tB, "", "ls | xargs -n1 bonsai check", RuleBonsaiCommand, ""},
	{tB, "", "xargs bonsai unlink < list.txt", RulePersonCommand, ""},

	// --- V2c's fix round (93167be): env's -S read as GNU env reads it, one at a time however it is written (bundled,
	// glued, long or abbreviated), and a command whose runners reach the walk's bound refused.
	{tB, "", "env -vS 'rm -rf docs'", RuleDeleteProtected, ""},
	{tB, "", "env -0S 'rm -rf junk/*'", RuleDeleteUnnamed, ""},
	{tB, "", "env -v0S 'rm config/app.key'", RuleDeleteProtected, ""},
	{tB, "", "env -vSrm -rf docs", RuleDeleteProtected, ""},
	{tB, "", "env -vS rm -rf docs", RuleDeleteProtected, ""},
	{tB, "", "env -vS 'git rm -r docs'", RuleDeleteProtected, ""},
	{tB, "", `env -vS 'sh -c "rm -rf docs"'`, RuleDeleteProtected, ""},
	{tB, "", "sudo env -vS 'rm -rf docs'", RuleDeleteProtected, ""},
	{tB, "", `bash -c "env -vS 'rm -rf docs'"`, RuleDeleteProtected, ""},
	{tB, "", "env -vS 'bonsai unlink'", RulePersonCommand, ""},
	{tB, "", "env -0S 'bonsai update --allow-exec'", RulePersonCommand, ""},
	{tB, "", "env -vS 'bonsai settings set x y'", RuleHomeCommand, ""},
	{tB, "", "env -vS '-u CLAUDE_CODE_CHILD_SESSION' ./x.sh", RulePersonCommand, ""},
	{tB, "", "env -vS -vS '-v' rm -rf docs", RuleDeleteProtected, ""},
	{tB, "", "env --split-string='rm -rf docs'", RuleDeleteProtected, ""},
	{tB, "", "env --spl 'rm -rf docs'", RuleDeleteProtected, ""},
	{tB, "", "env -v --split-s='-u CLAUDE_CODE_CHILD_SESSION' ./x.sh", RulePersonCommand, ""},
	{tB, "", "env -S '-v' -S 'rm -rf docs'", RuleDeleteProtected, ""},
	{tB, "", "env -S '-v' -S '-C docs' rm secret/plan.md", RuleDeleteProtected, ""},
	{tB, "", "env -S '-v' --split-string='-i' ./x.sh", RulePersonCommand, ""},
	{tB, "", "env -vS 'rm -rf build'", RuleDeleteNamed, ""},
	{tB, "", "env -vS 'make test'", RuleShell, ""},
	{tB, "", "env -0S 'FOO=1 BAR=2' go test ./...", RuleShell, ""},
	{tB, "", "env -vS 'bonsai status --json'", RuleBonsaiCommand, ""},
	{tB, "", strings.Repeat("nice ", 300) + "rm -rf docs", RuleDeleteUnsplit, ""},
	{tB, "", strings.Repeat("env ", 300) + "rm -rf build", RuleDeleteUnsplit, ""},
	{tB, "", strings.Repeat("command ", 300) + "bonsai status", RuleDeleteUnsplit, ""},
	{tB, "", strings.Repeat("nice ", 300) + "make", RuleDeleteUnsplit, ""},
	{tB, "", strings.Repeat("nice ", 200) + "rm -rf build", RuleDeleteNamed, ""},
	{tB, "", strings.Repeat("X=1 ", 300) + "rm -rf build", RuleDeleteNamed, ""},
	{tB, "", "find docs -exec env -S 'rm -rf' {} +", RuleDeleteProtected, ""},
	{tB, "", "find . -exec env -vS 'rm -rf' {} +", RuleDeleteUnnamed, ""},
	{tB, "", "find . -exec " + strings.Repeat("nice ", 70) + "rm -rf {} +", RuleDeleteUnnamed, ""},
	{tB, "", "find build -exec env -S 'rm -rf' {} +", RuleDeleteNamed, ""},
	{tB, "", `find . -exec env -S 'echo x' {} \;`, RuleShell, ""},

	// --- A line the check cannot split.
	{tB, "", "rm -rf 'junk", RuleDeleteUnsplit, ""},
	{tB, "", `rm -rf "junk`, RuleDeleteUnsplit, ""},
	{tB, "", "rm $(echo $(echo $(echo $(echo x))))", RuleDeleteUnsplit, ""},
	{tB, "", "echo $(echo $(echo $(echo $(echo x))))", RuleShell, ""},
	{tB, "", `echo "unterminated`, RuleShell, ""},
	{tB, "", delNest("rm -rf build", 3), RuleDeleteNamed, ""},
	{tB, "", delNest("rm -rf build", 4), RuleDeleteUnsplit, ""},
	{tB, "", "( ( ( ( rm -rf build ) ) ) )", RuleDeleteUnsplit, ""},
	{tB, "", "( ( ( rm -rf build ) ) )", RuleDeleteNamed, ""},
	{tB, "", "git clean -fdx `echo", RuleDeleteUnsplit, ""},
	{tP, "", "Remove-Item 'junk", RuleDeleteUnsplit, ""},
	{tP, "", "<# never closed\nRemove-Item junk", RuleDeleteUnsplit, ""},
	{tP, "", "@'\nRemove-Item -Recurse junk/*", RuleDeleteUnsplit, ""},

	// --- System-only rows (each says why).
	// Windows reads names case-blind: BONSAI.YAML is bonsai.yaml there, and another file elsewhere.
	{tB, "", "rm BONSAI.YAML", RuleDeleteProtected, "windows"},
	{tB, "", "rm BONSAI.YAML", RuleDeleteNamed, "unix"},
	{tP, "", "Remove-Item .GIT -Recurse", RuleDeleteProtected, "windows"},
	// Only on Windows does PowerShell's \ before a name stand for the drive's root folder.
	{tP, "", `Remove-Item -Recurse C:\`, RuleDeleteUnnamed, ""},
}

// delNest wraps a command in bash -c n times, quoting it each time.
func delNest(cmd string, n int) string {
	for k := 0; k < n; k++ {
		cmd = "bash -c '" + strings.ReplaceAll(cmd, "'", `'\''`) + "'"
	}
	return cmd
}

// delEncode gives PowerShell's -EncodedCommand text for a command.
func delEncode(s string) string {
	var b []byte
	for _, r := range s {
		b = append(b, byte(r), byte(r>>8))
	}
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var out strings.Builder
	for k := 0; k < len(b); k += 3 {
		var v uint32
		n := 0
		for j := 0; j < 3; j++ {
			v <<= 8
			if k+j < len(b) {
				v |= uint32(b[k+j])
				n++
			}
		}
		for j := 0; j < 4; j++ {
			if j <= n {
				out.WriteByte(table[(v>>(18-6*j))&63])
			} else {
				out.WriteByte('=')
			}
		}
	}
	return out.String()
}

func delRunRow(t *testing.T, s *delStub, row delRow) Decision {
	t.Helper()
	cwd := s.own
	if row.dir != "" {
		cwd = filepath.Join(s.own, filepath.FromSlash(row.dir))
	}
	return shellDecide(row.tool, row.line, cwd, s.own, s)
}

func TestShellTable(t *testing.T) {
	s := delFixture(t)
	counts := map[string]int{}
	for _, row := range delTable {
		switch row.only {
		case "windows":
			if runtime.GOOS != "windows" {
				continue
			}
		case "unix":
			if runtime.GOOS == "windows" {
				continue
			}
		}
		counts[row.want]++
		d := delRunRow(t, s, row)
		if d.Rule != row.want {
			t.Errorf("%s %q: rule %s, want %s\n  why: %s", row.tool, row.line, d.Rule, row.want, d.Why)
			continue
		}
		if d.Allow != delAllows(row.want) {
			t.Errorf("%s %q: allow %v under rule %s", row.tool, row.line, d.Allow, d.Rule)
		}
		if !d.Allow {
			for _, text := range []string{d.Why, d.Next} {
				if text == "" || strings.ContainsAny(text, "\n\r") || oneLine(text) != text {
					t.Errorf("%s %q: a refusal's reason and next step are one ASCII line each: %q / %q", row.tool, row.line, d.Why, d.Next)
				}
			}
			if strings.HasPrefix(row.want, "delete-") && !strings.Contains(d.Why, row.want) {
				t.Errorf("%s %q: the reason does not name its rule: %s", row.tool, row.line, d.Why)
			}
		}
	}
	var names []string
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		t.Logf("rows for %s: %d", n, counts[n])
	}
}

func delAllows(rule string) bool {
	for _, r := range Rules {
		if r.Name == rule {
			return r.Allow
		}
	}
	return false
}

// delCheckRules are the rules the delete check answers with (rules.go's table holds them).
var delCheckRules = []string{RuleDeleteUnnamed, RuleDeleteProtected, RuleDeleteUnsplit, RuleDeleteNamed, RuleShell,
	RuleBonsaiCommand, RulePersonCommand, RuleHomeCommand}

// A rule of the delete check with no row in the table fails.
func TestShellRulesEachHaveARow(t *testing.T) {
	for _, name := range delCheckRules {
		found := false
		for _, row := range delTable {
			if row.want == name && row.only == "" {
				found = true
			}
		}
		if !found {
			t.Errorf("the delete check's rule %s has no row in the table", name)
		}
	}
}

// Each delete word of the table has a row that refuses it and, unless it deletes only by its options, one it is
// read in: no word in the table goes untested.
func TestShellWordsEachHaveARow(t *testing.T) {
	for _, w := range DeleteWords {
		found := false
		for _, row := range delTable {
			if row.want != RuleDeleteUnnamed && row.want != RuleDeleteProtected {
				continue
			}
			for _, n := range w.Names {
				low := strings.ToLower(row.line)
				name := n
				if w.Sub != "" {
					name = "git"
				}
				if strings.Contains(low, name) && (w.Sub == "" || strings.Contains(low, w.Sub)) {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("the delete word %s has no refusing row", w.Word)
		}
	}
}

// The delete check's rules in the guard's one test per rule (rules_test.go's ruleTests, walked by TestEveryRule from
// the rules table): each answered through Main on a real project, its record holding the command's head and the
// deciding operand, never the line.
func init() {
	ruleTests[RuleDeleteUnnamed] = func(t *testing.T) {
		dir := withTask(t, "todo")
		for _, c := range []struct{ tool, line string }{{"Bash", "rm -rf junk/*"}, {"Monitor", "git clean -fdx"},
			{"PowerShell", "Remove-Item -Recurse $env:TEMP\\x"}, {"Bash", "cp ~/secret.txt x; ls | xargs rm"}} {
			stderr := wantRule(t, dir, bashIn(dir, c.tool, c.line), RuleDeleteUnnamed)
			text := delLastText(t, dir)
			if !strings.Contains(stderr, "(rule delete-unnamed)") || strings.Contains(text, "secret") || strings.Contains(text, "cp ") ||
				strings.Contains(text, "TEMP") || strings.Contains(text, c.line) {
				t.Errorf("%q: the reason names the rule and keeps no part of the line beyond its head: %q", c.line, text)
			}
		}
		if got := delLastTarget(t, dir); got != "" {
			t.Errorf("an operand from xargs names no target: %q", got)
		}
		wantRule(t, dir, bashIn(dir, "Bash", "rm -rf junk/*"), RuleDeleteUnnamed)
		if got := delLastTarget(t, dir); got != "junk/*" {
			t.Errorf("the deciding operand as target: %q", got)
		}
	}
	ruleTests[RuleDeleteProtected] = func(t *testing.T) {
		dir := withTask(t, "todo")
		for _, line := range []string{"rm bonsai.yaml", "rm -rf .git", "mv x.yaml .bonsai/", "rm -rf .claude", "rm -rf keys"} {
			stderr := wantRule(t, dir, bashIn(dir, "Bash", line), RuleDeleteProtected)
			if !strings.Contains(stderr, "(rule delete-protected)") || !strings.Contains(stderr, "refused: ") {
				t.Errorf("%q: %s", line, stderr)
			}
		}
		// A running task's grant opens a guarded path to a delete as to an edit; the floor's never-granted paths stay.
		dir = withTask(t, "running", "assets/**")
		wantRule(t, dir, bashIn(dir, "Bash", "rm -rf .git"), RuleDeleteProtected)
		wantRule(t, dir, bashIn(dir, "Bash", "rm -rf assets"), RuleDeleteNamed)
		if got := delLastTarget(t, dir); got != "assets" {
			t.Errorf("an allowed delete's target: %q", got)
		}
	}
	ruleTests[RuleDeleteUnsplit] = func(t *testing.T) {
		dir := withTask(t, "todo")
		wantRule(t, dir, bashIn(dir, "Bash", "rm -rf 'junk"), RuleDeleteUnsplit)
		wantRule(t, dir, bashIn(dir, "PowerShell", "Remove-Item \"junk"), RuleDeleteUnsplit)
		wantRule(t, dir, bashIn(dir, "Bash", "echo 'not closed"), RuleShell)
	}
	ruleTests[RuleDeleteNamed] = func(t *testing.T) {
		dir := withTask(t, "todo")
		for _, c := range []struct{ tool, line string }{{"Bash", "rm -rf node_modules build"}, {"PowerShell", "Remove-Item -Recurse -Force dist"},
			{"Monitor", "rm -f free.txt && npm run watch"}, {"Bash", "rm -rf vendor/lib/.git"}} {
			wantRule(t, dir, bashIn(dir, c.tool, c.line), RuleDeleteNamed)
		}
	}
}

// delLastText and delLastTarget read the last guard record of the test session's file in the project at dir.
func delLastText(t *testing.T, dir string) string {
	t.Helper()
	recs := records(t, dir, "s-r1.ndjson")
	return recs[len(recs)-1].String("text")
}

func delLastTarget(t *testing.T, dir string) string {
	t.Helper()
	recs := records(t, dir, "s-r1.ndjson")
	return recs[len(recs)-1].String("target")
}

// The machine's records (judgedRoots) are read only once a delete needs them: never for a call with no delete word,
// nor for a delete of named files; once for a line however many folders it deletes. A change of CLAUDE_CODE_CHILD_SESSION in
// any text it splits, nested or decoded, goes to personStep; each bonsai call to ownCommand, in its folder.
func TestShellAsksTheJudgeOnlyWhatItNeeds(t *testing.T) {
	s := delFixture(t)
	for _, line := range []string{"ls -la && git status", "rm file.txt a.txt", "mv a.txt b.txt", "git stash", "rm -rf junk/*"} {
		s.rootReads = 0
		shellDecide(tB, line, s.own, s.own, s)
		if s.rootReads != 0 {
			t.Errorf("%q read the machine's records", line)
		}
	}
	s.rootReads = 0
	shellDecide(tB, "rm -rf build dist node_modules && mv src old-src", s.own, s.own, s)
	if s.rootReads != 1 {
		t.Errorf("a line with four folders deleted or moved read the records %d times, want once", s.rootReads)
	}
	for _, c := range []struct{ tool, line string }{{tB, `bash -c 'unset CLAUDE_CODE_CHILD_SESSION'`},
		{tB, `eval "export CLAUDE_CODE_CHILD_SESSION="`}, {tP, "pwsh -EncodedCommand " + delEncode("Remove-Item Env:CLAUDE_CODE_CHILD_SESSION")}} {
		s.steps = nil
		if d := shellDecide(c.tool, c.line, s.own, s.own, s); d.Rule != RulePersonCommand || len(s.steps) != 1 {
			t.Errorf("%q: %s, person's steps %q", c.line, d.Rule, s.steps)
		}
	}
	s.commands = nil
	shellDecide(tB, "cd build && bonsai status; bonsai check", s.own, s.own, s)
	want := []string{filepath.Join(s.own, "build") + " bonsai status", filepath.Join(s.own, "build") + " bonsai check", s.own + " bonsai check"}
	if strings.Join(s.commands, "|") != strings.Join(want, "|") {
		t.Errorf("ownCommand's calls %q, want %q", s.commands, want)
	}
}

// Each word that runs another program has a row whose delete it reaches.
func TestShellRunnersEachHaveARow(t *testing.T) {
	for _, rn := range RunnerWords {
		found := false
		for _, row := range delTable {
			if row.want != RuleDeleteUnnamed && row.want != RuleDeleteProtected {
				continue
			}
			for _, f := range strings.Fields(strings.ToLower(row.line)) {
				if f == rn.Word {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("the runner %s has no refusing row", rn.Word)
		}
	}
}

// The answer names the operand and the rule, and the next step; its target is the deciding operand, relative to
// the project.
func TestShellAnswerNamesTheOperand(t *testing.T) {
	s := delFixture(t)
	cases := []struct {
		tool, line, target, next string
	}{
		{tB, "rm -rf junk/*", "junk/*", "name the files, or delete the folder by its name"},
		{tB, "rm -rf .git", ".git", "leave it as it is and tell the person"},
		{tB, "mv x.yaml bonsai.yaml", "bonsai.yaml", "leave it as it is and tell the person"},
		{tB, "git clean -fdx", "", "leave -x and -X out"},
		{tB, "rm -rf /tmp/*", "", "name the files"},
		{tP, "Remove-Item -Recurse junk/*", "junk/*", "name the files"},
	}
	for _, c := range cases {
		d := shellDecide(c.tool, c.line, s.own, s.own, s)
		if d.Allow {
			t.Errorf("%q allowed", c.line)
			continue
		}
		if d.Target != c.target {
			t.Errorf("%q: target %q, want %q", c.line, d.Target, c.target)
		}
		if !strings.HasPrefix(d.Next, c.next) {
			t.Errorf("%q: next step %q, want it to start %q", c.line, d.Next, c.next)
		}
		if c.target != "" && !strings.Contains(d.Why, c.target) {
			t.Errorf("%q: the reason does not name the operand %q: %s", c.line, c.target, d.Why)
		}
	}
	d := shellDecide(tB, "rm -rf build", s.own, s.own, s)
	if !d.Allow || d.Target != "build" {
		t.Errorf("a named delete: %+v", d)
	}
}

// A running task's grant opens a guarded path to a delete as it does to an edit.
func TestShellGrantOpensAGuardedPath(t *testing.T) {
	s := delFixture(t)
	if d := shellDecide(tB, "rm -rf docs/secret", s.own, s.own, s); d.Allow {
		t.Fatalf("not granted, allowed: %+v", d)
	}
	s.grants = []string{"docs/secret/**"}
	for _, line := range []string{"rm -rf docs/secret", "rm -rf docs", "mv docs/secret old-secret"} {
		if d := shellDecide(tB, line, s.own, s.own, s); !d.Allow {
			t.Errorf("%q granted, refused: %s", line, d.Why)
		}
	}
	if d := shellDecide(tB, "rm -rf .git", s.own, s.own, s); d.Allow {
		t.Errorf("a grant opened .git: %+v", d)
	}
}

// The folder each command runs in, through cd: a cd to a name moves it, one that may fail leaves both folders, one
// that cannot be told leaves it unknown.
func TestShellFollowsTheFolder(t *testing.T) {
	s := delFixture(t)
	cases := []struct {
		line, want string
	}{
		{"cd src && rm ../bonsai.yaml", RuleDeleteProtected},
		{"cd src; rm ../bonsai.yaml", RuleDeleteProtected},
		{"cd src && rm -rf .git", RuleDeleteNamed},
		{"cd sub && rm bonsai.yaml", RuleDeleteProtected},
		{"cd sub; cd ..; rm -rf .git", RuleDeleteProtected},
		{"(cd vendor/lib && rm -rf .git); ls", RuleDeleteNamed},
		{"(cd vendor/lib; rm -rf .git)", RuleDeleteProtected},
		{"cd vendor/lib && rm -rf .git && cd ../.. && ls", RuleDeleteNamed},
		{"cd \"$(git rev-parse --show-toplevel)\" && rm -rf build", RuleDeleteUnnamed},
		{"cd \"$X\" && rm bonsai.yaml", RuleDeleteProtected},
		{"cd \"$X\" && rm notes.txt", RuleDeleteNamed},
		{"pushd vendor/lib && rm -rf .git && popd && rm -rf build", RuleDeleteNamed},
		{"cd " + filepath.ToSlash(s.own) + "/src && rm -rf ../.git", RuleDeleteProtected},
		{"cd ~ && rm -rf .bonsai", RuleDeleteProtected},
	}
	for _, c := range cases {
		if d := shellDecide(tB, c.line, s.own, s.own, s); d.Rule != c.want {
			t.Errorf("%q: %s, want %s (%s)", c.line, d.Rule, c.want, d.Why)
		}
	}
}

// The splitter's words: quoting, nesting and here-documents, for each shell.
func TestShellSplitWords(t *testing.T) {
	cases := []struct {
		d     shDialect
		line  string
		words [][]string // each command's words, in order, nested lists after their command
	}{
		{shBash, `rm -rf "a b" 'c d' e\ f`, [][]string{{"rm", "-rf", "a b", "c d", "e f"}}},
		{shBash, `a; b && c || d | e & f`, [][]string{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}, {"f"}}},
		{shBash, "x=1 y=2 cmd arg", [][]string{{"x=1", "y=2", "cmd", "arg"}}},
		{shBash, "echo $(a $(b))", [][]string{{"echo", "$(a $(b))"}, {"a", "$(b)"}, {"b"}}},
		{shBash, "cat <<EOF\nrm -rf x\nEOF\nls", [][]string{{"cat"}, {"ls"}}},
		{shBash, "cat <<-'X'\n\tbody\n\tX\nnext", [][]string{{"cat"}, {"next"}}},
		{shBash, "a 2>/dev/null >out.txt <in.txt b", [][]string{{"a", "b"}}},
		{shBash, `echo "$(printf '%s' "x")"`, [][]string{{"echo", `$(printf '%s' "x")`}, {"printf", "%s", "x"}}},
		{shBash, `$'a\tb' $"c"`, [][]string{{"a\tb", "c"}}},
		{shPower, `Remove-Item 'a''b' "c""d" e` + "`" + ` f`, [][]string{{"Remove-Item", "a'b", `c"d`, "e f"}}},
		{shPower, "a; b | c && d", [][]string{{"a"}, {"b"}, {"c"}, {"d"}}},
		{shPower, "Get-Item (Join-Path a b)", [][]string{{"Get-Item", "(Join-Path a b)"}, {"Join-Path", "a", "b"}}},
		{shCmd, `rd /s /q "my dir" & del ^&x`, [][]string{{"rd", "/s", "/q", "my dir"}, {"del", "&x"}}},
	}
	for _, c := range cases {
		l, why := shSplit(c.line, c.d, 0)
		if why != "" {
			t.Errorf("%q: cannot split: %s", c.line, why)
			continue
		}
		var got [][]string
		var walk func(l *shList)
		walk = func(l *shList) {
			for _, cmd := range l.cmds {
				var ws []string
				for _, w := range cmd.words {
					ws = append(ws, w.text)
				}
				if len(ws) > 0 {
					got = append(got, ws)
				}
				if cmd.group != nil {
					walk(cmd.group)
				}
				for _, n := range cmd.nested {
					walk(n.list)
				}
			}
		}
		walk(l)
		if strings.Join(delFlat(got), "|") != strings.Join(delFlat(c.words), "|") {
			t.Errorf("%q: words %q, want %q", c.line, got, c.words)
		}
	}
}

func delFlat(ws [][]string) []string {
	var out []string
	for _, w := range ws {
		out = append(out, strings.Join(w, "\x00"))
	}
	return out
}

// The glob meeting: a guarded glob met against an operand, without the tree.
func TestShellMeet(t *testing.T) {
	cases := []struct {
		op, glob string
		tree     bool
		want     string // the path given, "" for none
	}{
		{"docs", "docs/secret/**", true, "docs/secret"},
		{"docs", "docs/secret/**", false, ""},
		{"docs/secret/x", "docs/secret/**", false, "docs/secret/x"},
		{"config", "config/*.key", true, "config/x.key"},
		{"config/a", "config/*.key", true, ""},
		{"", ".git", true, ".git"},
		{"*", "bonsai.yaml", true, "bonsai.yaml"},
		{"*.png", "bonsai.yaml", true, ""},
		{"*", "docs/secret/**", true, "docs/secret"},
		{"*.md", "notes/**", true, ""},
		{"docs/*", "docs/secret/**", true, "docs/secret"},
		{"x/*/y", "x/[ab]*/y", false, "x/ax/y"},
		{"src", "**/bonsai.yaml", false, ""},
		{"*/bonsai.yaml", "**/bonsai.yaml", false, "bonsai.yaml/bonsai.yaml"},
	}
	for _, c := range cases {
		var op []string
		if c.op != "" {
			op = strings.Split(c.op, "/")
		}
		rep, ok := delMeet(op, delGlobSegs(c.glob), c.tree, false)
		got := ""
		if ok {
			got = strings.Join(rep, "/")
		}
		if got != c.want {
			t.Errorf("meet(%q, %q, tree %v) = %q, want %q", c.op, c.glob, c.tree, got, c.want)
		}
	}
}
