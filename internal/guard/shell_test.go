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
}

// delStubGlobs are the own project's guarded globs, as delRoots gives them.
var delStubGlobs = []string{"bonsai.yaml", ".bonsai/lock.json", ".bonsai/local/", ".bonsai/.gitignore",
	".bonsai/tasks.md", ".bonsai/sessions.md", ".git", ".claude/settings.json", ".claude/settings.local.json",
	"**/bonsai.yaml", "**/.bonsai/", "docs/secret/**", "config/*.key", "notes/"}

func (s *delStub) delRoots() []delRoot {
	return []delRoot{
		{Dir: s.own, Globs: delStubGlobs, Own: true},
		{Dir: s.other, Globs: []string{".git", ".bonsai", ".claude"}},
		{Dir: s.home, Whole: true},
	}
}

func (s *delStub) delMay(dir, p string) Decision {
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
		"proj/.bonsai/tasks.md", "proj/.git/HEAD", "proj/.claude/settings.json", "proj/.claude/notes.md",
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
	{tB, "", "rm -rf node_modules", delRuleNamed, ""},
	{tB, "", "rm file.txt", delRuleNamed, ""},
	{tB, "", "git rm src/main.go", delRuleNamed, ""},
	{tB, "", "mv a.txt b.txt", delRuleNamed, ""},
	{tB, "", "git stash", delRuleNone, ""},
	{tB, "", "git clean -n", delRuleNone, ""},
	{tB, "", "git clean -nd", delRuleNone, ""},
	{tB, "", "git clean -fdx --dry-run", delRuleNone, ""},
	{tB, "", "git clean -fd build", delRuleNamed, ""},
	{tB, "", "rm -rf build dist", delRuleNamed, ""},
	{tB, "", "rm -f junk/one.txt junk/two.txt", delRuleNamed, ""},
	{tB, "", "rm -rf ./build", delRuleNamed, ""},
	{tB, "", "rm -r build/", delRuleNamed, ""},
	{tB, "", "mv src/main.go src/app.go", delRuleNamed, ""},
	{tB, "", "git mv src/main.go src/app.go", delRuleNamed, ""},
	{tB, "", "rmdir build/out", delRuleNamed, ""},
	{tB, "", "rm -rf /tmp/bonsai-scratch-xyz", delRuleNamed, ""},
	{tB, "", "ls -la && git status", delRuleNone, ""},
	{tB, "", "go test ./... && go vet ./...", delRuleNone, ""},
	{tB, "", `git commit -m "drop the rm -rf junk/* step"`, delRuleNone, ""},
	{tB, "", "git commit -m \"$(cat <<'EOF'\nDelete junk with rm -rf junk/* and git clean -fdx.\nIt's done.\nEOF\n)\"", delRuleNone, ""},
	{tB, "", "echo 'rm -rf /' > notes.txt", delRuleNone, ""},
	{tB, "", "cat > clean.sh <<'EOF'\nrm -rf junk/*\nEOF\nchmod +x clean.sh", delRuleNone, ""},
	{tB, "", "git stash -u", delRuleNone, ""},
	{tB, "", "git stash pop", delRuleNone, ""},
	{tB, "", `git stash push -m "wip: -a test"`, delRuleNone, ""},
	{tB, "", "find . -name '*.go' -newer go.mod", delRuleNone, ""},
	{tB, "", "find build -name '*.o' -delete", delRuleNamed, ""},
	{tB, "", "rsync -a src/ backup/", delRuleNone, ""},
	{tB, "", "rsync -a --delete src/ backup/", delRuleNamed, ""},
	{tB, "", "rsync -a --delay-updates src/ backup/", delRuleNone, ""},
	{tB, "", "rm -rf node_modules && npm install", delRuleNamed, ""},
	{tB, "", "cd build && rm -rf out", delRuleNamed, ""},
	{tB, "", "(cd build && rm -rf out) && ls", delRuleNamed, ""},
	{tB, "", "rm -rf vendor/lib/.git", delRuleNamed, ""},
	{tB, "", "rm .bonsai/STATE.md", delRuleNamed, ""},
	{tB, "", "mv *.png assets/", delRuleNamed, ""},
	{tB, "", "git rm -r --cached build", delRuleNamed, ""},
	{tB, "", "docker run --rm -v x:/y image", delRuleNone, ""},
	{tB, "", "npm run clean", delRuleNone, ""},
	{tB, "", "rm -- -weird-name", delRuleNamed, ""},
	{tB, "", "unlink file.txt", delRuleNamed, ""},
	{tB, "", "rm -rf {build,dist}", delRuleNamed, ""},
	{tB, "", "find . -type f | xargs grep -l TODO", delRuleNone, ""},
	{tB, "", "git -C sub clean -fd src", delRuleNamed, ""},
	{tB, "", "for f in a b; do echo $f; done", delRuleNone, ""},
	{tB, "", "if [ -d build ]; then rm -rf build; fi", delRuleNamed, ""},
	{tB, "", "test -d build && rm -rf build || true", delRuleNamed, ""},
	{tB, "", "rm -rf build 2>/dev/null; mkdir -p build", delRuleNamed, ""},
	{tB, "", `rm -rf "build" 'dist'`, delRuleNamed, ""},
	{tB, "", "cd src; cd ..; rm -rf build", delRuleNamed, ""},
	{tB, "", "pushd build && rm -rf out && popd && rm -rf dist", delRuleNamed, ""},
	{tB, "", "f() { rm -rf build; }; f", delRuleNamed, ""},
	{tB, "", "timeout 60 go test ./...", delRuleNone, ""},
	{tB, "", "command -v rm", delRuleNone, ""},
	{tB, "", "robocopy src backup /MIR", delRuleNamed, ""},
	{tB, "", "robocopy src backup /MIR /L", delRuleNone, ""},
	{tB, "", "rm -rf ../proj-scratch", delRuleNamed, ""},
	{tB, "", "rm -rf ../other/src", delRuleNamed, ""},
	{tB, "", "rm -rf ~/notes-old", delRuleNamed, ""},
	{tB, "", "cd nowhere && rm -rf .git", delRuleNamed, ""},
	{tM, "", "rm -rf build && npm run build -- --watch", delRuleNamed, ""},
	{tP, "", "Remove-Item -Recurse -Force node_modules", delRuleNamed, ""},
	{tP, "", "Remove-Item file.txt", delRuleNamed, ""},
	{tP, "", "Move-Item a.txt b.txt", delRuleNamed, ""},
	{tP, "", "Rename-Item a.txt b.txt", delRuleNamed, ""},
	{tP, "", "Get-ChildItem -Recurse | Select-Object Name", delRuleNone, ""},
	{tP, "", "Remove-Item -LiteralPath 'junk/*'", delRuleNamed, ""},
	{tP, "", "rm build -Recurse", delRuleNamed, ""},
	{tP, "", `Remove-Item .\build -Recurse -Force -ErrorAction SilentlyContinue`, delRuleNamed, ""},
	{tP, "", "Remove-Item junk/one.txt, junk/two.txt", delRuleNamed, ""},
	{tP, "", "Remove-Item -Path build -Recurse -Include *.o", delRuleNamed, ""},
	{tP, "", "Remove-Item -Recurse -WhatIf *", delRuleNone, ""},
	{tP, "", "if (Test-Path build) { Remove-Item -Recurse -Force build }", delRuleNamed, ""},
	{tP, "", `cmd /c "rd /s /q build"`, delRuleNamed, ""},
	{tP, "", `cmd /c del junk\one.txt`, delRuleNamed, ""},
	{tP, "", "git clean -fd build; git status", delRuleNamed, ""},
	{tP, "", "Write-Output 'Remove-Item -Recurse junk/*'", delRuleNone, ""},

	// --- Rule 1, the splitter, and rule 2, the delete words: every spelling reaches the program.
	{tB, "", "rm -rf junk/*", delRuleUnnamed, ""},
	{tM, "", "rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "/bin/rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "RM.EXE -rf junk/*", delRuleUnnamed, ""},
	{tB, "", `\rm -rf junk/*`, delRuleUnnamed, ""},
	{tB, "", `"rm" -rf junk/*`, delRuleUnnamed, ""},
	{tB, "", "r''m -rf junk/*", delRuleUnnamed, ""},
	{tB, "", `$'\x72\x6d' -rf junk/*`, delRuleUnnamed, ""},
	{tB, "", "sudo rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "sudo -u root -E rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "env FOO=1 BAR=2 rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "FOO=1 rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "command rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "exec rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "nice -n 5 rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "nice -10 rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "nohup rm -rf junk/* &", delRuleUnnamed, ""},
	{tB, "", "time rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "timeout -s KILL 5 rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "env -S 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", "doas -u root rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "builtin command rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "ionice -c 3 rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "setsid -f rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "stdbuf -o0 rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "zsh -c 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", "dash -c 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", "ls; rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "ls && rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "false || rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "ls | rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "ls\nrm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "ls \\\n && rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "echo $(rm -rf junk/*)", delRuleUnnamed, ""},
	{tB, "", "echo `rm -rf junk/*`", delRuleUnnamed, ""},
	{tB, "", "(rm -rf junk/*)", delRuleUnnamed, ""},
	{tB, "", "{ rm -rf junk/*; }", delRuleUnnamed, ""},
	{tB, "", "f() { rm -rf junk/*; }; f", delRuleUnnamed, ""},
	{tB, "", "diff <(rm -rf junk/*) x", delRuleUnnamed, ""},
	{tB, "", "x=$(rm -rf junk/*)", delRuleUnnamed, ""},
	{tB, "", `echo "${X:-$(rm -rf junk/*)}"`, delRuleUnnamed, ""},
	{tB, "", "bash -c 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", `sh -c "rm -rf junk/*"`, delRuleUnnamed, ""},
	{tB, "", "bash -lc 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", "bash -o pipefail -c 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", `bash -c "bash -c 'rm -rf junk/*'"`, delRuleUnnamed, ""},
	{tB, "", "eval 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", "eval rm -rf 'junk/*'", delRuleUnnamed, ""},
	{tB, "", "bash <<EOF\nrm -rf junk/*\nEOF", delRuleUnnamed, ""},
	{tB, "", "bash -s <<'EOF'\nrm -rf junk/*\nEOF", delRuleUnnamed, ""},
	{tB, "", "sh <<< 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", "echo 'rm -rf junk/*' | bash", delRuleUnnamed, ""},
	{tB, "", "cat <<EOF | sh\nrm -rf junk/*\nEOF", delRuleUnnamed, ""},
	{tB, "", "cat <<EOF\n$(rm -rf junk/*)\nEOF", delRuleUnnamed, ""},
	{tB, "", `cmd /c "rd /s /q ."`, delRuleUnnamed, ""},
	{tB, "", "cmd.exe //c del /s *.log", delRuleUnnamed, ""},
	{tB, "", "cmd /c rd/s/q %TEMP%", delRuleUnnamed, ""},
	{tB, "", `powershell -Command "Remove-Item -Recurse junk/*"`, delRuleUnnamed, ""},
	{tB, "", "pwsh -NoProfile -c 'Remove-Item -r junk/*'", delRuleUnnamed, ""},
	{tB, "", "wsl rm -rf 'junk/*'", delRuleUnnamed, ""},
	{tB, "", "busybox rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "$RM -rf junk/*", delRuleUnnamed, ""},
	{tB, "", `"$(which rm)" -r junk/*`, delRuleUnnamed, ""},
	{tB, "", "if true; then rm -rf junk/*; fi", delRuleUnnamed, ""},
	{tB, "", "while false; do rm -rf junk/*; done", delRuleUnnamed, ""},
	{tB, "", "! rm -rf junk/*", delRuleUnnamed, ""},
	{tB, "", "# a comment\nrm -rf junk/* # it's gone", delRuleUnnamed, ""},
	{tP, "", "Remove-Item -Recurse junk/*", delRuleUnnamed, ""},
	{tP, "", `Remove-Item junk\* -Recurse`, delRuleUnnamed, ""},
	{tP, "", "rm -r junk/*", delRuleUnnamed, ""},
	{tP, "", "ri -Rec junk/*", delRuleUnnamed, ""},
	{tP, "", "del -recurse junk/*", delRuleUnnamed, ""},
	{tP, "", "erase junk/*", delRuleUnnamed, ""},
	{tP, "", "rd -r junk/*", delRuleUnnamed, ""},
	{tP, "", "rmdir -R junk/*", delRuleUnnamed, ""},
	{tP, "", "REMOVE-ITEM -RECURSE junk/*", delRuleUnnamed, ""},
	{tP, "", "Microsoft.PowerShell.Management\\Remove-Item -Recurse junk/*", delRuleUnnamed, ""},
	{tP, "", "Remove-Item \u2013Recurse $p", delRuleUnnamed, ""},
	{tP, "", "Remove-Item\u00a0-Recurse\u00a0junk/*", delRuleUnnamed, ""},
	{tP, "", "Remove-Item -Recurse:$true $p", delRuleUnnamed, ""},
	{tP, "", "$null = Remove-Item -Recurse junk/*", delRuleUnnamed, ""},
	{tP, "", "$r=Remove-Item -Recurse junk/*", delRuleUnnamed, ""},
	{tP, "", "& { Remove-Item -Recurse junk/* }", delRuleUnnamed, ""},
	{tP, "", "& 'Remove-Item' -Recurse junk/*", delRuleUnnamed, ""},
	{tP, "", "[void](Remove-Item -Recurse junk/*)", delRuleUnnamed, ""},
	{tP, "", `powershell -Command "Remove-Item -Recurse junk/*"`, delRuleUnnamed, ""},
	{tP, "", "pwsh -c 'Remove-Item -r junk/*'", delRuleUnnamed, ""},
	{tP, "", "powershell -NoProfile -EncodedCommand " + delEncode("Remove-Item -Recurse junk/*"), delRuleUnnamed, ""},
	{tP, "", "@'\nRemove-Item -Recurse junk/*\n'@ | powershell -Command -", delRuleUnnamed, ""},
	{tP, "", "'Remove-Item -Recurse junk/*' | iex", delRuleUnnamed, ""},
	{tP, "", "Invoke-Expression 'rm -r junk/*'", delRuleUnnamed, ""},
	{tP, "", `cmd /c "rd /s /q ."`, delRuleUnnamed, ""},
	{tP, "", "bash -c 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tP, "", "wsl -e rm -rf junk/x junk/*", delRuleUnnamed, ""},
	{tP, "", "rm.exe -rf junk/*", delRuleUnnamed, ""},
	{tP, "", "git clean -fdx", delRuleUnnamed, ""},
	{tP, "", "Get-ChildItem | ForEach-Object { if ($_.Name -like '*.log') { Remove-Item $_ } }", delRuleUnnamed, ""},
	{tP, "", "# Remove-Item is safe here\nRemove-Item -Recurse junk/* # it's gone", delRuleUnnamed, ""},
	{tP, "", "<# a block\ncomment #> Remove-Item -Recurse junk/*", delRuleUnnamed, ""},
	{tP, "", "Remove-Item `\n  -Recurse `\n  junk/*", delRuleUnnamed, ""},

	// --- Rule 3, not naming what it deletes: every form of recursive and bulk, the operand tests, the always-refused.
	{tB, "", "rm -rf *", delRuleUnnamed, ""},
	{tB, "", "rm -rf .", delRuleUnnamed, ""},
	{tB, "", "rm -rf ./", delRuleUnnamed, ""},
	{tB, "", "rm -rf ..", delRuleUnnamed, ""},
	{tB, "", "rm -rf /", delRuleUnnamed, ""},
	{tB, "", "rm -rf ~", delRuleUnnamed, ""},
	{tB, "", "rm -rf ~/", delRuleUnnamed, ""},
	{tB, "", `rm -rf ""`, delRuleUnnamed, ""},
	{tB, "", "rm -rf build/.", delRuleUnnamed, ""},
	{tB, "", "rm -rf build/..", delRuleUnnamed, ""},
	{tB, "", "rm -rf /mnt/c", delRuleUnnamed, ""},
	{tB, "", "rm -rf /c/", delRuleUnnamed, ""},
	{tB, "", `rm -rf "$DIR"`, delRuleUnnamed, ""},
	{tB, "", "rm -rf $DIR/build", delRuleUnnamed, ""},
	{tB, "", "rm -rf ${DIR}", delRuleUnnamed, ""},
	{tB, "", "rm -rf $(pwd)/build", delRuleUnnamed, ""},
	{tB, "", "rm -rf `pwd`", delRuleUnnamed, ""},
	{tB, "", "rm -rf %TEMP%", delRuleUnnamed, ""},
	{tB, "", "rm -fR junk/*", delRuleUnnamed, ""},
	{tB, "", "rm -Rf junk/*", delRuleUnnamed, ""},
	{tB, "", "rm -vfr junk/*", delRuleUnnamed, ""},
	{tB, "", "rm -r -f junk/*", delRuleUnnamed, ""},
	{tB, "", "rm --recursive junk/*", delRuleUnnamed, ""},
	{tB, "", "rm --rec --force junk/*", delRuleUnnamed, ""},
	{tB, "", "rm --r junk/*", delRuleUnnamed, ""},
	{tB, "", "rm junk/* -rf", delRuleUnnamed, ""},
	{tB, "", "rm *.log", delRuleUnnamed, ""},
	{tB, "", "rm junk/file?.txt", delRuleUnnamed, ""},
	{tB, "", "rm junk/[ab].txt", delRuleUnnamed, ""},
	{tB, "", "rm -rf ./*", delRuleUnnamed, ""},
	{tB, "", "rm -rf junk/ build/ *", delRuleUnnamed, ""},
	{tB, "", "rm -rf {junk,build}/*", delRuleUnnamed, ""},
	{tB, "", "rm -rf junk/{1..3}", delRuleUnnamed, ""},
	{tB, "", `rm -rf $'junk/\x2a'`, delRuleUnnamed, ""},
	{tB, "", "rm -rf !(keep)", delRuleUnnamed, ""},
	{tB, "", "ls | xargs rm", delRuleUnnamed, ""},
	{tB, "", "find . -name x | xargs rm -f", delRuleUnnamed, ""},
	{tB, "", "xargs -I{} rm -rf {} < list.txt", delRuleUnnamed, ""},
	{tB, "", "xargs -0 -n 1 mv -t old/ < list.txt", delRuleUnnamed, ""},
	{tB, "", "find . -delete", delRuleUnnamed, ""},
	{tB, "", "find -delete", delRuleUnnamed, ""},
	{tB, "", "find . -name '*.log' -delete", delRuleUnnamed, ""},
	{tB, "", `find . -exec rm {} \;`, delRuleUnnamed, ""},
	{tB, "", "find . -type d -execdir rm -rf {} +", delRuleUnnamed, ""},
	{tB, "", `find . -ok rm {} \;`, delRuleUnnamed, ""},
	{tB, "", `find . -exec sudo /bin/rm -f {} \;`, delRuleUnnamed, ""},
	{tB, "", "find junk/* -delete", delRuleUnnamed, ""},
	{tB, "", "find $DIR -delete", delRuleUnnamed, ""},
	{tB, "", "git clean -fd", delRuleUnnamed, ""},
	{tB, "", "git clean -f .", delRuleUnnamed, ""},
	{tB, "", "git clean -fdx build", delRuleUnnamed, ""},
	{tB, "", "git clean -fdX", delRuleUnnamed, ""},
	{tB, "", "git clean -xfd", delRuleUnnamed, ""},
	{tB, "", "git clean -d -f -x -- build", delRuleUnnamed, ""},
	{tB, "", "git -C sub clean -fdx", delRuleUnnamed, ""},
	{tB, "", "git -c core.quotepath=off clean -ffdx", delRuleUnnamed, ""},
	{tB, "", "git clean -fd :/", delRuleUnnamed, ""},
	{tB, "", "git stash --all", delRuleUnnamed, ""},
	{tB, "", "git stash -a", delRuleUnnamed, ""},
	{tB, "", "git stash push --all", delRuleUnnamed, ""},
	{tB, "", `git stash save -a "x"`, delRuleUnnamed, ""},
	{tB, "", "git stash --al", delRuleUnnamed, ""},
	{tB, "", "git stash push -ka -- src", delRuleUnnamed, ""},
	{tB, "", "rm -rf --no-preserve-root build", delRuleUnnamed, ""},
	{tB, "", "rm -rf --no-p build", delRuleUnnamed, ""},
	{tB, "", "git rm -r .", delRuleUnnamed, ""},
	{tB, "", "git rm '*.log'", delRuleUnnamed, ""},
	{tB, "", "git rm -r :/", delRuleUnnamed, ""},
	{tB, "", "git rm --pathspec-from-file=list.txt", delRuleUnnamed, ""},
	{tB, "", "rsync -a --delete src/ .", delRuleUnnamed, ""},
	{tB, "", "rsync -a --delete-after src/ ./", delRuleUnnamed, ""},
	{tB, "", "rsync -a --del src/ $DEST", delRuleUnnamed, ""},
	{tB, "", "robocopy src . /MIR", delRuleUnnamed, ""},
	{tB, "", "robocopy src . //PURGE", delRuleUnnamed, ""},
	{tB, "", `cd "$X" && rm -rf build`, delRuleUnnamed, ""},
	{tB, "", "cd $(mktemp -d); rm -rf build", delRuleUnnamed, ""},
	{tB, "", "cd -; rm -rf build", delRuleUnnamed, ""},
	{tB, "", "env -C \"$X\" rm -rf build", delRuleUnnamed, ""},
	{tB, "", `for f in *.log; do rm "$f"; done`, delRuleUnnamed, ""},
	{tB, "", "for d in */; do\n  rm -r \"$d\"\ndone", delRuleUnnamed, ""},
	{tB, "", "while read -r f; do mv \"$f\" old/; done < list.txt", delRuleUnnamed, ""},
	{tP, "", "Remove-Item -Recurse .", delRuleUnnamed, ""},
	{tP, "", "Remove-Item -Recurse $env:TEMP\\x", delRuleUnnamed, ""},
	{tP, "", "Remove-Item -Recurse \"$HOME\\build\"", delRuleUnnamed, ""},
	{tP, "", "Remove-Item -Recurse -Path (Join-Path $root 'x')", delRuleUnnamed, ""},
	{tP, "", "Get-ChildItem *.log | Remove-Item", delRuleUnnamed, ""},
	{tP, "", "gci build | ri -Recurse", delRuleUnnamed, ""},
	{tP, "", "Get-ChildItem *.txt | Move-Item -Destination old", delRuleUnnamed, ""},
	{tP, "", "Remove-Item * -Include *.log -Recurse", delRuleUnnamed, ""},
	{tP, "", "Remove-Item -Recurse:$false junk/*", delRuleUnnamed, ""},
	{tP, "", "Remove-Item @params", delRuleUnnamed, ""},
	{tP, "", "foreach ($f in $files) { Remove-Item $f }", delRuleUnnamed, ""},
	{tP, "", "$files | ForEach-Object { Remove-Item -LiteralPath $_.FullName }", delRuleUnnamed, ""},
	{tP, "", `robocopy src C:\ /MIR`, delRuleUnnamed, ""},
	{tP, "", "cmd /c for /d %i in (*) do rd /s /q %i", delRuleUnnamed, ""},
	{tP, "", "cmd /c del /s /q *.tmp", delRuleUnnamed, ""},
	{tP, "", "cmd /c \"if exist build rd /s /q %CD%\"", delRuleUnnamed, ""},
	{tP, "", "Set-Location $env:TEMP; Remove-Item -Recurse build", delRuleUnnamed, ""},

	// --- Rule 4, a guarded path: deleted, moved, or held by a folder deleted or moved.
	{tB, "", "rm bonsai.yaml", delRuleProtected, ""},
	{tB, "", "rm -f ./bonsai.yaml", delRuleProtected, ""},
	{tB, "", "rm .bonsai/lock.json", delRuleProtected, ""},
	{tB, "", "rm -rf .bonsai", delRuleProtected, ""},
	{tB, "", "rm -rf .bonsai/local", delRuleProtected, ""},
	{tB, "", "rm -rf .git", delRuleProtected, ""},
	{tB, "", "rm -rf .git/hooks", delRuleProtected, ""},
	{tB, "", "rm -rf .claude", delRuleProtected, ""},
	{tB, "", "rm .claude/settings.json", delRuleProtected, ""},
	{tB, "", "rm -rf docs/secret", delRuleProtected, ""},
	{tB, "", "rm -rf docs", delRuleProtected, ""},
	{tB, "", "rm config/app.key", delRuleProtected, ""},
	{tB, "", "rm -rf config", delRuleProtected, ""},
	{tB, "", "rm notes/a.md", delRuleProtected, ""},
	{tB, "", "rm -rf notes", delRuleProtected, ""},
	{tB, "", "rm sub/bonsai.yaml", delRuleProtected, ""},
	{tB, "", "rm -rf sub/.bonsai", delRuleProtected, ""},
	{tB, "", "mv bonsai.yaml bonsai.yaml.bak", delRuleProtected, ""},
	{tB, "", "mv x.yaml bonsai.yaml", delRuleProtected, ""},
	{tB, "", "mv settings.json .claude/", delRuleProtected, ""},
	{tB, "", "mv -t .claude settings.json", delRuleProtected, ""},
	{tB, "", "mv .git .git-old", delRuleProtected, ""},
	{tB, "", "mv * src/", delRuleProtected, ""},
	{tB, "", "mv docs/* archive/", delRuleProtected, ""},
	{tB, "", "mv .claude old-claude", delRuleProtected, ""},
	{tB, "", "git rm bonsai.yaml", delRuleProtected, ""},
	{tB, "", "git rm -r --cached .bonsai", delRuleProtected, ""},
	{tB, "", "git mv bonsai.yaml x.yaml", delRuleProtected, ""},
	{tB, "", "git clean -fd .bonsai", delRuleProtected, ""},
	{tB, "", "shred -u bonsai.yaml", delRuleProtected, ""},
	{tB, "", "unlink .bonsai/lock.json", delRuleProtected, ""},
	{tB, "", "rm -rf ../proj", delRuleProtected, ""},
	{tB, "src", "rm -rf ../../proj", delRuleProtected, ""},
	{tB, "", "cd .. && rm -rf proj", delRuleProtected, ""},
	{tB, "", "cd build; rm -f ../bonsai.yaml", delRuleProtected, ""},
	{tB, "", "cd nowhere; rm -rf .git", delRuleProtected, ""},
	{tB, "src", "rm ../bonsai.yaml", delRuleProtected, ""},
	{tB, "", "find .git -delete", delRuleProtected, ""},
	{tB, "", "find .bonsai -name '*.json' -delete", delRuleProtected, ""},
	{tB, "", "rsync -a --delete empty/ .git/", delRuleProtected, ""},
	{tB, "", "robocopy empty .git /MIR", delRuleProtected, ""},
	{tB, "", "rm -rf {bonsai.yaml,x}", delRuleProtected, ""},
	{tB, "", "rm {bonsai.yaml,}", delRuleProtected, ""},
	{tB, "", "rm -rf ~/.bonsai", delRuleProtected, ""},
	{tB, "", "rm -rf ~/.bonsai/workspaces", delRuleProtected, ""},
	{tB, "", "rm -rf ../home", delRuleProtected, ""},
	{tB, "", "rm -rf ../other", delRuleProtected, ""},
	{tB, "", "rm -rf ../other/.git", delRuleProtected, ""},
	{tB, "", "git -C .. rm -r proj/.git", delRuleProtected, ""},
	{tB, "", "bash -c 'rm bonsai.yaml'", delRuleProtected, ""},
	{tB, "", "rm -rf build bonsai.yaml", delRuleProtected, ""},
	{tB, "", "rsync -a --remove-source-files bonsai.yaml backup/", delRuleProtected, ""},
	{tB, "", "robocopy .git backup /MOVE", delRuleProtected, ""},
	{tB, "", "$RM -rf .git", delRuleProtected, ""},
	{tP, "", "Remove-Item bonsai.yaml", delRuleProtected, ""},
	{tP, "", "Remove-Item -Recurse .git -Force", delRuleProtected, ""},
	{tP, "", `Remove-Item -LiteralPath .claude\settings.json`, delRuleProtected, ""},
	{tP, "", "Remove-Item .claude", delRuleProtected, ""},
	{tP, "", "Move-Item bonsai.yaml x.yaml", delRuleProtected, ""},
	{tP, "", "Move-Item -Path x.yaml -Destination bonsai.yaml", delRuleProtected, ""},
	{tP, "", "Rename-Item bonsai.yaml x.yaml", delRuleProtected, ""},
	{tP, "", "Rename-Item -Path x.yaml -NewName bonsai.yaml", delRuleProtected, ""},
	{tP, "", "Remove-Item FileSystem::bonsai.yaml", delRuleProtected, ""},
	{tP, "", "Remove-Item -Path .bonsai\\lock.json, x.txt", delRuleProtected, ""},
	{tP, "", `cmd /c del bonsai.yaml`, delRuleProtected, ""},
	{tP, "", `cmd /c rd /s /q .git`, delRuleProtected, ""},
	{tP, "", `cmd /c "move bonsai.yaml x"`, delRuleProtected, ""},
	{tP, "", `cmd /c ren bonsai.yaml x.yaml`, delRuleProtected, ""},
	{tP, "", `cmd /c del .claude`, delRuleProtected, ""},
	{tP, "", "Set-Location src; Remove-Item ..\\bonsai.yaml", delRuleProtected, ""},
	{tP, "", "Push-Location ..; Remove-Item -Recurse proj; Pop-Location", delRuleProtected, ""},

	// --- Lines written to slip past the splitter (the builder's own break-it), each refused or rightly allowed.
	{tB, "", "{rm,-rf,junk/*}", delRuleUnnamed, ""},
	{tB, "", "rm${IFS}-rf${IFS}junk/*", delRuleUnnamed, ""},
	{tB, "", "rm${IFS}bonsai.yaml", delRuleUnnamed, ""},
	{tB, "", "bash -c \"$(cat <<'EOF'\nrm -rf junk/*\nEOF\n)\"", delRuleUnnamed, ""},
	{tB, "", "bash <(echo 'rm -rf junk/*')", delRuleUnnamed, ""},
	{tB, "", "source <(echo 'rm -rf junk/*')", delRuleUnnamed, ""},
	{tB, "", `eval "$(echo rm -rf junk/*)"`, delRuleUnnamed, ""},
	{tB, "", "trap 'rm -rf junk/*' EXIT", delRuleUnnamed, ""},
	{tB, "", "su -c 'rm -rf junk/*' root", delRuleUnnamed, ""},
	{tB, "", "watch -n 1 'rm -rf junk/*'", delRuleUnnamed, ""},
	{tB, "", "rm -rf ~home", delRuleUnnamed, ""},
	{tB, "", "rm -rf ~+", delRuleUnnamed, ""},
	{tB, "", "rm -rf ~home/.bonsai", delRuleProtected, ""},
	{tB, "", `rm "$(echo bonsai.yaml)"`, delRuleProtected, ""},
	{tB, "", "rm `echo bonsai.yaml`", delRuleProtected, ""},
	{tB, "", "$(echo rm) bonsai.yaml", delRuleProtected, ""},
	{tB, "", "$(echo rm -rf junk/*)", delRuleUnnamed, ""},
	{tB, "", "bash <<< \"$(echo rm -rf junk/*)\"", delRuleUnnamed, ""},
	{tB, "", "rm $(echo a.txt bonsai.yaml)", delRuleProtected, ""},
	{tB, "", `rm "$(echo a.txt bonsai.yaml)"`, delRuleNamed, ""},
	{tB, "", "mv */bonsai.yaml old/", delRuleProtected, ""},
	{tB, "", `find build -exec rm -rf {} bonsai.yaml \;`, delRuleProtected, ""},
	{tB, "", "git -C .. stash -a", delRuleUnnamed, ""},
	{tB, "", "mv junk/* .", delRuleNamed, ""},
	{tB, "", "find build -exec rm -rf {} +", delRuleNamed, ""},
	{tB, "", "rm -rf .git.bak .gitignore .claude/agents", delRuleNamed, ""},
	{tB, "", "git stash -- --all", delRuleNone, ""},
	{tB, "", "git clean -fde -x build", delRuleNamed, ""},
	{tP, "", `Remove-Item ("bonsai.yaml")`, delRuleProtected, ""},
	{tP, "", "Remove-Item $('bonsai.yaml')", delRuleProtected, ""},
	{tP, "", "Remove-Item -Recurse 'build','.git'", delRuleProtected, ""},
	{tP, "", "Remove-Item -Recurse @('build','.git')", delRuleUnnamed, ""},
	{tP, "", "'rm -r junk/*' | powershell -noprofile -", delRuleUnnamed, ""},
	{tP, "", "& (Get-Command Remove-Item) -Recurse junk/*", delRuleUnnamed, ""},
	{tP, "", `cmd /c "cd .. & rd /s /q proj"`, delRuleProtected, ""},
	{tP, "", "cmd /c start /b rd /s /q .", delRuleUnnamed, ""},
	{tP, "", "cmd /c call rd /s /q .", delRuleUnnamed, ""},
	{tP, "", "cmd /c r^d /s /q .", delRuleUnnamed, ""},
	{tP, "", "cmd /c rd .git /s /q", delRuleProtected, ""},
	{tP, "", `cmd /v:on /c "set X=. & rd /s /q !X!"`, delRuleUnnamed, ""},
	{tP, "", `Move-Item junk\* .`, delRuleNamed, ""},
	{tP, "", "Move-Item -Path build -Destination .git", delRuleProtected, ""},

	// --- Lines ordinary work might write that are refused, each by the plan's rule 3 (the report lists why).
	{tB, "", "git rm -r --cached .", delRuleUnnamed, ""},
	{tB, "", "find . -maxdepth 1 -name '*.tmp' -delete", delRuleUnnamed, ""},
	{tB, "", "rm -f *.pyc", delRuleUnnamed, ""},
	{tB, "", `rm -rf "$TMPDIR/bonsai-x"`, delRuleUnnamed, ""},
	{tB, "", `cd "$(git rev-parse --show-toplevel)" && rm -rf build`, delRuleUnnamed, ""},
	{tP, "", "Get-ChildItem -Filter *.log | Remove-Item", delRuleUnnamed, ""},

	// --- A line the check cannot split.
	{tB, "", "rm -rf 'junk", delRuleUnsplit, ""},
	{tB, "", `rm -rf "junk`, delRuleUnsplit, ""},
	{tB, "", "rm $(echo $(echo $(echo $(echo x))))", delRuleUnsplit, ""},
	{tB, "", "echo $(echo $(echo $(echo $(echo x))))", delRuleNone, ""},
	{tB, "", `echo "unterminated`, delRuleNone, ""},
	{tB, "", delNest("rm -rf build", 3), delRuleNamed, ""},
	{tB, "", delNest("rm -rf build", 4), delRuleUnsplit, ""},
	{tB, "", "( ( ( ( rm -rf build ) ) ) )", delRuleUnsplit, ""},
	{tB, "", "( ( ( rm -rf build ) ) )", delRuleNamed, ""},
	{tB, "", "git clean -fdx `echo", delRuleUnsplit, ""},
	{tP, "", "Remove-Item 'junk", delRuleUnsplit, ""},
	{tP, "", "<# never closed\nRemove-Item junk", delRuleUnsplit, ""},
	{tP, "", "@'\nRemove-Item -Recurse junk/*", delRuleUnsplit, ""},

	// --- System-only rows (each says why).
	// Windows reads names case-blind: BONSAI.YAML is bonsai.yaml there, and another file elsewhere.
	{tB, "", "rm BONSAI.YAML", delRuleProtected, "windows"},
	{tB, "", "rm BONSAI.YAML", delRuleNamed, "unix"},
	{tP, "", "Remove-Item .GIT -Recurse", delRuleProtected, "windows"},
	// Only on Windows does PowerShell's \ before a name stand for the drive's root folder.
	{tP, "", `Remove-Item -Recurse C:\`, delRuleUnnamed, ""},
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
	return shellDecide(row.tool, row.line, cwd, s)
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
			if !strings.Contains(d.Why, row.want) {
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
	for _, r := range delRules {
		if r.Name == rule {
			return r.Allow
		}
	}
	return false
}

// A rule with no row in the table fails.
func TestShellRulesEachHaveARow(t *testing.T) {
	for _, r := range delRules {
		found := false
		for _, row := range delTable {
			if row.want == r.Name && row.only == "" {
				found = true
			}
		}
		if !found {
			t.Errorf("the delete check's rule %s has no row in the table", r.Name)
		}
	}
}

// Each delete word of the table has a row that refuses it and, unless it deletes only by its options, one it is
// read in: no word in the table goes untested.
func TestShellWordsEachHaveARow(t *testing.T) {
	for _, w := range delWords {
		found := false
		for _, row := range delTable {
			if row.want != delRuleUnnamed && row.want != delRuleProtected {
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

// Each word that runs another program has a row whose delete it reaches.
func TestShellRunnersEachHaveARow(t *testing.T) {
	for _, rn := range delRunners {
		found := false
		for _, row := range delTable {
			if row.want != delRuleUnnamed && row.want != delRuleProtected {
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
		d := shellDecide(c.tool, c.line, s.own, s)
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
	d := shellDecide(tB, "rm -rf build", s.own, s)
	if !d.Allow || d.Target != "build" {
		t.Errorf("a named delete: %+v", d)
	}
}

// A running task's grant opens a guarded path to a delete as it does to an edit.
func TestShellGrantOpensAGuardedPath(t *testing.T) {
	s := delFixture(t)
	if d := shellDecide(tB, "rm -rf docs/secret", s.own, s); d.Allow {
		t.Fatalf("not granted, allowed: %+v", d)
	}
	s.grants = []string{"docs/secret/**"}
	for _, line := range []string{"rm -rf docs/secret", "rm -rf docs", "mv docs/secret old-secret"} {
		if d := shellDecide(tB, line, s.own, s); !d.Allow {
			t.Errorf("%q granted, refused: %s", line, d.Why)
		}
	}
	if d := shellDecide(tB, "rm -rf .git", s.own, s); d.Allow {
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
		{"cd src && rm ../bonsai.yaml", delRuleProtected},
		{"cd src; rm ../bonsai.yaml", delRuleProtected},
		{"cd src && rm -rf .git", delRuleNamed},
		{"cd sub && rm bonsai.yaml", delRuleProtected},
		{"cd sub; cd ..; rm -rf .git", delRuleProtected},
		{"(cd vendor/lib && rm -rf .git); ls", delRuleNamed},
		{"(cd vendor/lib; rm -rf .git)", delRuleProtected},
		{"cd vendor/lib && rm -rf .git && cd ../.. && ls", delRuleNamed},
		{"cd \"$(git rev-parse --show-toplevel)\" && rm -rf build", delRuleUnnamed},
		{"cd \"$X\" && rm bonsai.yaml", delRuleProtected},
		{"cd \"$X\" && rm notes.txt", delRuleNamed},
		{"pushd vendor/lib && rm -rf .git && popd && rm -rf build", delRuleNamed},
		{"cd " + filepath.ToSlash(s.own) + "/src && rm -rf ../.git", delRuleProtected},
		{"cd ~ && rm -rf .bonsai", delRuleProtected},
	}
	for _, c := range cases {
		if d := shellDecide(tB, c.line, s.own, s); d.Rule != c.want {
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
