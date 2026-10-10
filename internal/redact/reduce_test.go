package redact

import (
	"regexp"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

func TestCommandHeadBash(t *testing.T) {
	rows := []struct{ cmd, want string }{
		{"git log --oneline -3", "git log"},
		{"git -C ../other -c core.pager=less diff HEAD", "git diff"},
		{"git -c http.extraheader=\"PRIVATE-TOKEN: Hq7v\" clone https://example.org/r.git", "git clone"},
		{"git --work-tree=../site --namespace v2 --git-dir ../site/.git stash pop", "git stash"},
		{"/usr/bin/git show HEAD~1", "git show"},
		{"git \"pull\" --rebase", "git"},
		{"npm --prefix web run build", "npm run"},
		{"npm Install", "npm"},
		{"cd web && npm ci", "npm ci"},
		{"(cd web; npm test)", "npm test"},
		{"dotnet build src/App.csproj", "dotnet build"},
		{"unity --project game test", "unity test"},
		{"claude --model sonnet mcp list", "claude mcp"},
		{"claude -p \"list what passwd=Hq7v unlocks\"", "claude"},
		{"gh -R someone/repo issue list", "gh issue"},
		{"gh api -H \"X-Token: Hq7v\" repos/o/r", "gh api"},
		{"go -C ./tools test ./...", "go test"},
		{"bonsai check --json", "bonsai check"},
		{"SIGNING_SECRET=Hq7v ./publish.sh --all", "publish"},
		{"ADMIN_PASSWD='two words' ruby seed.rb", "ruby"},
		{"nohup nice -n 10 rsync -a src/ dst/", "rsync"},
		{"env A=1 B=2 make all", "make"},
		{"nice -n 5 time make all", "make"},
		{"python3 -c 'import os; print(os.environ[\"API_KEY\"])'", "python3"},
		{"ssh -i ~/.ssh/key builder@example.com uptime", "ssh"},
		{"wget --header='X-Api-Key: Qm3v' -O out.json https://data.example.net/feed", "wget"},
		{`"C:\Tools\node\node.exe" build.js`, "node"},
		{"echo hi | grep h", "echo"},
		{"printf '%s' \"$(gpg -d vault.gpg)\"", "printf"},
		{"# a comment only", ""},
		{"BUILD_MODE=fast LEVEL=2", ""},
		{"pushd ../vendor/lib", "pushd"},
		{"cd /somewhere && pushd x", "cd"},
		{"7z a out.zip in", ""},
		{"   ", ""},
		{"", ""},
		{"lint\nformat && git push", "lint"},
	}
	for _, r := range rows {
		if got := CommandHead(r.cmd, false); got != r.want {
			t.Errorf("CommandHead(%q) = %q, want %q", r.cmd, got, r.want)
		}
	}
}

func TestCommandHeadPowerShell(t *testing.T) {
	rows := []struct{ cmd, want string }{
		{"Remove-Item -Force D:\\build\\tmp", "Remove-Item"},
		{"$env:NPM_SECRET = 'Hq7v'; npm run pack", "npm run"},
		{"& 'D:\\tools\\gh\\gh.exe' pr list", "gh pr"},
		{"Push-Location D:\\site; npm --workspace web run lint", "npm run"},
		{"Invoke-WebRequest -Headers @{'X-Api-Key'='Hq7v'} -Uri https://api.example.org/v1", "Invoke-WebRequest"},
		{"cd D:\\logs; Select-String -Pattern secret app.log | Out-File hits.txt", "Select-String"},
		{"$retries = 5", ""},
		{"Pop-Location", "pop-location"},
		{"go vet ./internal/...", "go vet"},
	}
	for _, r := range rows {
		if got := CommandHead(r.cmd, true); got != r.want {
			t.Errorf("CommandHead(%q, powershell) = %q, want %q", r.cmd, got, r.want)
		}
	}
}

// A head is only a program name, one subcommand word or a cmdlet: never a flag, path, value or quoted text.
func TestCommandHeadKeepsNothingElse(t *testing.T) {
	shape := regexp.MustCompile(`^[a-z][a-z0-9-]*( [a-z][a-z0-9-]*)?$|^[A-Za-z]+-[A-Za-z][A-Za-z0-9]*$`)
	cmds := []string{
		"VAULT_SECRET=Qv7s ruby -e x", "git -c http.extraheader=Qv7s push", "scp -i /keys/Qv7s a b:", "'Qv7s' run",
		"\"/srv/Qv7s/bin\" x", "Qv7s=1 make", "./Qv7s.sh", "$env:Y='Qv7s'; y", "env -u Qv7s Z=1 w",
	}
	for _, c := range cmds {
		for _, ps := range []bool{false, true} {
			h := CommandHead(c, ps)
			if h == "" {
				continue
			}
			if !shape.MatchString(h) || strings.ContainsAny(h, `/\"'=$`) {
				t.Errorf("CommandHead(%q, %v) = %q", c, ps, h)
			}
		}
	}
}

func obj(kv ...any) schema.Object {
	var o schema.Object
	for i := 0; i+1 < len(kv); i += 2 {
		o = append(o, schema.Member{Key: kv[i].(string), Value: kv[i+1]})
	}
	return o
}

func TestTarget(t *testing.T) {
	const winRoot = `D:\code\app`
	const unixRoot = "/srv/code/app"
	rows := []struct {
		tool      string
		input     schema.Object
		root, cwd string
		want      string
	}{
		{"Read", obj("file_path", `D:\code\app\README.md`), winRoot, "", "README.md"},
		{"Read", obj("file_path", "D:/code/app/src/main.go"), winRoot, "", "src/main.go"},
		{"Read", obj("file_path", `d:\CODE\App\src\main.go`), winRoot, "", "src/main.go"},
		{"Read", obj("file_path", `D:\code\app`), winRoot, "", "."},
		{"Read", obj("file_path", `D:\code\apple\x.txt`), winRoot, "", ""},
		{"Read", obj("file_path", `E:\elsewhere\x.txt`), winRoot, "", ""},
		{"Read", obj("file_path", `D:\code\app\..\secrets.txt`), winRoot, "", ""},
		{"Read", obj("file_path", "/srv/code/app/docs/a.md"), unixRoot, "", "docs/a.md"},
		{"Read", obj("file_path", "/SRV/code/app/docs/a.md"), unixRoot, "", ""},
		{"Read", obj("file_path", "/etc/passwd"), unixRoot, "", ""},
		{"Read", obj("file_path", "~/notes.txt"), unixRoot, "", ""},
		{"Edit", obj("file_path", "/srv/code/app/a.txt", "old_string", "passwd=Hq7v"), unixRoot, "", "a.txt"},
		{"Write", obj("file_path", "/srv/code/app/out.txt", "content", "Hq7v in a note"), unixRoot, "", "out.txt"},
		{"MultiEdit", obj("file_path", "/srv/code/app/b.ts", "edits", []any{}), unixRoot, "", "b.ts"},
		{"NotebookEdit", obj("notebook_path", "/srv/code/app/n.ipynb"), unixRoot, "", "n.ipynb"},
		{"Read", obj("file_path", "/srv/code/app/secret=Hq7v.log"), unixRoot, "", "secret=[redacted]"},
		{"Grep", obj("pattern", "Hq7v", "path", "notes/2026"), unixRoot, "", "notes/2026"},
		{"Grep", obj("pattern", "x", "path", "src"), unixRoot, "/srv/code/app/web", "web/src"},
		{"Grep", obj("pattern", "x", "path", "../../elsewhere"), unixRoot, "", ""},
		{"Grep", obj("pattern", "x"), unixRoot, "", ""},
		{"Glob", obj("pattern", "**/*.go", "path", `D:\code\app\internal`), winRoot, "", "internal"},
		{"Bash", obj("command", "API_TOKEN=hunter2 npm run build"), unixRoot, "", "npm run"},
		{"PowerShell", obj("command", "$env:K='Hq7v'; Get-Item .\\cfg"), winRoot, "", "Get-Item"},
		{"WebFetch", obj("url", "HTTP://ops:Wq8t@Status.Example.ORG:9000/health?token=Wq8t#top", "prompt", "check"), unixRoot, "", "status.example.org"},
		{"WebFetch", obj("url", "::: no scheme"), unixRoot, "", ""},
		{"WebSearch", obj("query", "passwd=Hq7v leaked"), unixRoot, "", ""},
		{"mcp__notes__search_text", obj("query", "Hq7v"), unixRoot, "", "notes.search_text"},
		{"mcp__a__b__c", nil, unixRoot, "", "a.b__c"},
		{"mcp__only", nil, unixRoot, "", ""},
		{"Agent", obj("subagent_type", "reviewer", "prompt", "passwd=Hq7v"), unixRoot, "", "reviewer"},
		{"Task", obj("subagent_type", "release-checker"), unixRoot, "", "release-checker"},
		{"Agent", obj("prompt", "x"), unixRoot, "", ""},
		{"Skill", obj("skill", "changelog", "args", "secret=Hq7v"), unixRoot, "", "changelog"},
		{"AskUserQuestion", obj("questions", []any{}), unixRoot, "", ""},
		{"TodoWrite", obj("todos", []any{}), unixRoot, "", ""},
		{"Read", nil, unixRoot, "", ""},
		{"Read", obj("file_path", 42), unixRoot, "", ""},
	}
	for _, r := range rows {
		if got := Target(r.tool, r.input, r.root, r.cwd); got != r.want {
			t.Errorf("Target(%s, %v) = %q, want %q", r.tool, r.input, got, r.want)
		}
	}
	long := "/srv/code/app/" + strings.Repeat("d/", 150) + "x.txt"
	if got := Target("Read", obj("file_path", long), unixRoot, ""); len([]rune(got)) != TargetCap {
		t.Errorf("a long target is %d characters", len([]rune(got)))
	}
	if got := Target("Read", obj("file_path", "/srv/code/app/a.txt"), "", ""); got != "" {
		t.Errorf("no root gave %q", got)
	}
}

func TestQuestion(t *testing.T) {
	q := func(texts ...any) schema.Object {
		var list []any
		for _, x := range texts {
			list = append(list, obj("question", x))
		}
		return obj("questions", list)
	}
	rows := []struct {
		in   schema.Object
		want string
	}{
		{q("Ship it today?"), "Ship it today?"},
		{q("First?", "Second?"), "First? \u00b7 Second?"},
		{q("Should the deploy use db_secret=Lp9w or the old one?"), "Should the deploy use db_secret=[redacted] or the old one?"},
		{q("", 7, "Only this?"), "Only this?"},
		{obj("questions", "not a list"), ""},
		{obj(), ""},
		{nil, ""},
	}
	for _, r := range rows {
		if got := Question(r.in); got != r.want {
			t.Errorf("Question(%v) = %q, want %q", r.in, got, r.want)
		}
	}
	if got := Question(q(strings.Repeat("x", 500))); len([]rune(got)) != TextCap {
		t.Errorf("a long question is %d characters", len([]rune(got)))
	}
}
