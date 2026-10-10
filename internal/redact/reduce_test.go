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
		{"git -c http.extraHeader=\"Authorization: Bearer hunter2\" fetch origin", "git fetch"},
		{"git --git-dir=/tmp/x/.git --work-tree /tmp/y status", "git status"},
		{"/usr/bin/git show HEAD~1", "git show"},
		{"git 'status'", "git"},
		{"npm --prefix web run build", "npm run"},
		{"npm Install", "npm"},
		{"cd web && npm ci", "npm ci"},
		{"(cd web; npm test)", "npm test"},
		{"dotnet build src/App.csproj", "dotnet build"},
		{"unity --project game test", "unity test"},
		{"claude --model sonnet mcp list", "claude mcp"},
		{"claude -p \"summarise password=hunter2\"", "claude"},
		{"gh -R someone/repo issue list", "gh issue"},
		{"gh api -H \"Authorization: token hunter2\" /user", "gh api"},
		{"go -C ./tools test ./...", "go test"},
		{"bonsai check --json", "bonsai check"},
		{"API_TOKEN=hunter2 ./deploy.sh --now", "deploy"},
		{"DB_PASSWORD=\"two words\" python3 manage.py migrate", "python3"},
		{"sudo -u builder systemctl restart app", "systemctl"},
		{"env A=1 B=2 make all", "make"},
		{"nice -n 5 time make all", "make"},
		{"node -e \"console.log(process.env.SECRET)\"", "node"},
		{"ssh -i ~/.ssh/key builder@example.com uptime", "ssh"},
		{"curl -H \"Authorization: Bearer hunter2\" https://example.com", "curl"},
		{`"C:\Tools\node\node.exe" build.js`, "node"},
		{"echo hi | grep h", "echo"},
		{"echo $(cat ~/.config/token)", "echo"},
		{"# a comment only", ""},
		{"FOO=1", ""},
		{"cd /somewhere", "cd"},
		{"cd /somewhere && pushd x", "cd"},
		{"7z a out.zip in", ""},
		{"   ", ""},
		{"", ""},
		{"first\nsecond && git push", "first"},
	}
	for _, r := range rows {
		if got := CommandHead(r.cmd, false); got != r.want {
			t.Errorf("CommandHead(%q) = %q, want %q", r.cmd, got, r.want)
		}
	}
}

func TestCommandHeadPowerShell(t *testing.T) {
	rows := []struct{ cmd, want string }{
		{"Get-ChildItem -Path C:\\Users -Recurse", "Get-ChildItem"},
		{"$env:API_TOKEN = 'hunter2'; npm run build", "npm run"},
		{"& \"C:\\Program Files\\Git\\cmd\\git.exe\" status", "git status"},
		{"Set-Location C:\\work; git -C . log", "git log"},
		{"Invoke-RestMethod -Uri https://example.com -Headers @{Authorization='Bearer hunter2'}", "Invoke-RestMethod"},
		{"cd C:\\work; Get-Content notes.txt | Select-String token", "Get-Content"},
		{"$count = 3", ""},
		{"Set-Location C:\\work", "set-location"},
		{"go test ./...", "go test"},
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
		"API_TOKEN=hunter2 node -e x", "git -c http.extraHeader=hunter2 push", "ssh -i /keys/hunter2 x", "'hunter2' run",
		"\"/opt/hunter2/bin\" x", "hunter2=1 make", "./hunter2.sh", "$env:X='hunter2'; x", "env -u hunter2 X=1 y",
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
		{"Read", obj("file_path", `E:\other\x.txt`), winRoot, "", ""},
		{"Read", obj("file_path", `D:\code\app\..\secrets.txt`), winRoot, "", ""},
		{"Read", obj("file_path", "/srv/code/app/docs/a.md"), unixRoot, "", "docs/a.md"},
		{"Read", obj("file_path", "/SRV/code/app/docs/a.md"), unixRoot, "", ""},
		{"Read", obj("file_path", "/etc/passwd"), unixRoot, "", ""},
		{"Read", obj("file_path", "~/notes.txt"), unixRoot, "", ""},
		{"Edit", obj("file_path", "/srv/code/app/a.txt", "old_string", "password=hunter2"), unixRoot, "", "a.txt"},
		{"Write", obj("file_path", "/srv/code/app/out.txt", "content", "hunter2"), unixRoot, "", "out.txt"},
		{"MultiEdit", obj("file_path", "/srv/code/app/b.ts", "edits", []any{}), unixRoot, "", "b.ts"},
		{"NotebookEdit", obj("notebook_path", "/srv/code/app/n.ipynb"), unixRoot, "", "n.ipynb"},
		{"Read", obj("file_path", "/srv/code/app/password=hunter2.txt"), unixRoot, "", "password=[redacted]"},
		{"Grep", obj("pattern", "hunter2", "path", "docs"), unixRoot, "", "docs"},
		{"Grep", obj("pattern", "x", "path", "src"), unixRoot, "/srv/code/app/web", "web/src"},
		{"Grep", obj("pattern", "x", "path", "../../elsewhere"), unixRoot, "", ""},
		{"Grep", obj("pattern", "x"), unixRoot, "", ""},
		{"Glob", obj("pattern", "**/*.go", "path", `D:\code\app\internal`), winRoot, "", "internal"},
		{"Bash", obj("command", "API_TOKEN=hunter2 npm run build"), unixRoot, "", "npm run"},
		{"PowerShell", obj("command", "$env:T='hunter2'; Get-ChildItem"), winRoot, "", "Get-ChildItem"},
		{"WebFetch", obj("url", "https://someone:hunter2@API.example.com:8443/v1?api_key=hunter2", "prompt", "x"), unixRoot, "", "api.example.com"},
		{"WebFetch", obj("url", "not a url"), unixRoot, "", ""},
		{"WebSearch", obj("query", "password=hunter2"), unixRoot, "", ""},
		{"mcp__files__read_text", obj("path", "hunter2"), unixRoot, "", "files.read_text"},
		{"mcp__a__b__c", nil, unixRoot, "", "a.b__c"},
		{"mcp__only", nil, unixRoot, "", ""},
		{"Agent", obj("subagent_type", "reviewer", "prompt", "password=hunter2"), unixRoot, "", "reviewer"},
		{"Task", obj("subagent_type", "builder"), unixRoot, "", "builder"},
		{"Agent", obj("prompt", "x"), unixRoot, "", ""},
		{"Skill", obj("skill", "release-notes", "args", "token=hunter2"), unixRoot, "", "release-notes"},
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
		{q("Is password=hunter2 the right one?"), "Is password=[redacted] the right one?"},
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
