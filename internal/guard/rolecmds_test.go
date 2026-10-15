package guard

// The roles' rules on an agent's shell line (rolecmds.go): a start of Claude Code, the writes a role that changes
// nothing may not make, Bonsai's own records refused to every agent's readable shell writes, git's write words, bonsai
// message's --from and --kind brief, and the report's exact form.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// shellAs is a Bash call (PowerShell when ps) from an agent of a type ("" for the person's session).
func shellAs(dir, agentID, typ, command string, ps bool) string {
	tool := "Bash"
	if ps {
		tool = "PowerShell"
	}
	return hookCall{session: sessA, agentID: agentID, agentType: typ, tool: tool, input: bash(command)}.payload(dir)
}

func TestAStartOfClaudeCode(t *testing.T) {
	dir := rolesProject(t, nil)
	starts := []string{"claude -p hi", "/usr/local/bin/claude --agent workflow:lead -p x", "CLAUDE.EXE -p x",
		`"C:\Users\x\.local\bin\claude.exe" -p x`, "claude-here -p x", "claude-here.cmd -p x", "env BONSAI_ROLE=lead claude -p x",
		"npx @anthropic-ai/claude-code -p x", "npx -y @anthropic-ai/claude-code@2.1.300 -p x", "pnpx @anthropic-ai/claude-code",
		"npm exec @anthropic-ai/claude-code -- -p x", "node ~/.npm/node_modules/@anthropic-ai/claude-code/cli.js -p x",
		"bash -c 'claude -p x'", "echo a && (cd /tmp && claude -p x)", "nohup claude -p x &", "x=$(claude -p hi)"}
	for _, line := range starts {
		for _, typ := range []string{"workflow:lead", "workflow:helper", "Explore"} {
			wantAnswer(t, dir, shellAs(dir, "a1", typ, line, false), 2, RuleCallRefused)
		}
		// The person's own session starts Claude Code freely.
		wantAnswer(t, dir, shellAs(dir, "", "", line, false), 0, "")
	}
	// A named session (BONSAI_ROLE, no agent id) is refused too.
	wantAnswer(t, dir, shellAs(dir, "", "", "claude -p x", false), 2, RuleCallRefused, RoleEnv, "lead")
	// Not a start: a word that only holds the name.
	for _, line := range []string{"echo claude", "cat claude.md", "grep -r claude-code docs", "npx prettier x", "node cli.js",
		"ls ~/.claude"} {
		wantAnswer(t, dir, shellAs(dir, "a1", "workflow:lead", line, false), 0, "")
	}
}

func TestWritesOfARoleThatChangesNothing(t *testing.T) {
	dir := rolesProject(t, nil)
	// Paths on a Bash line in forward slashes: Git Bash reads a backslash as an escape (C:\Users\x is C:Usersx, a
	// name in the folder the command runs in), so Windows' own form would name a file inside the project.
	out := filepath.ToSlash(t.TempDir())
	slash := func(parts ...string) string { return filepath.ToSlash(filepath.Join(parts...)) }
	refused := []string{"echo x > notes.txt", "echo x >> free.txt", "echo x &> log.txt", "echo x 2> err.txt", "echo x >| f",
		"printf x | tee a.txt", "tee -a free.txt < /dev/null", "cp free.txt copy.txt", "cp /etc/hosts .", "ln -s free.txt l",
		"sed -i s/a/b/ free.txt", "sed --in-place=.bak -e s/a/b/ free.txt", "rm free.txt", "mv free.txt moved.txt",
		"git commit -m x", "git -C . add -A", "git checkout -- free.txt", "git stash", "git branch new", "git tag v1",
		"git reset --hard", "bonsai check --write", "bonsai update --yes", "cd sub && echo x > y.txt",
		"echo x > " + slash(dir, "abs.txt"), "{ echo x; } > grouped.txt", "touch new.txt", "truncate -s 0 free.txt",
		"install -m 644 free.txt inst.txt", "install -d newdir", "dd if=/dev/zero of=free.txt count=1",
		"bash -c 'echo x > nested.txt'", "x=$(echo y > sub.txt)", "eval 'cp free.txt e.txt'", "git branch -v new",
		"git config alias.ci commit", "git -c alias.ci=commit ci -m x", "git stash push -m list"}
	for i, line := range refused {
		// A fresh agent each time: an undeclared type runs on the smallest helper's budget (tiny's 3 actions).
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("h", i), "workflow:helper", line, false), 2, RuleRoleWritesNothing)
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("e", i), "Explore", line, false), 2, RuleRoleWritesNothing)
		// A lead that writes: as before (the delete check alone).
		if code, s := runs(t, dir, shellAs(dir, "a1", "workflow:lead", line, false)); code != 0 && !strings.Contains(line, "bonsai update") {
			t.Errorf("a lead: %s: %s", line, s)
		}
	}
	allowed := []string{"echo x > " + slash(out, "a.txt"), "go build -o " + out + "/b ./...", "git status", "git diff HEAD~1",
		"git log --oneline", "git show HEAD", "git branch", "git branch -a", "git tag -l", "git stash list", "cat free.txt",
		"grep -r x .", "echo x 2>&1", "ls >&2", "cp free.txt " + out, "sed s/a/b/ free.txt", "bonsai check", "bonsai status --json",
		"echo x > /dev/null", "touch " + slash(out, "t"), "dd if=free.txt of=" + slash(out, "d"),
		"git config --get user.name", "git -c core.pager=cat log -1"}
	for _, line := range allowed {
		wantAnswer(t, dir, shellAs(dir, "h1", "workflow:helper", line, false), 0, "")
	}
	// PowerShell's forms.
	for _, line := range []string{"Set-Content -Path free.txt -Value x", "'x' | Out-File notes.txt", "Add-Content free.txt x",
		"Copy-Item free.txt -Destination c.txt", "New-Item -Path n.txt", "echo x > p.txt", "Remove-Item free.txt"} {
		wantAnswer(t, dir, shellAs(dir, "h1", "workflow:helper", line, true), 2, RuleRoleWritesNothing)
	}
	wantAnswer(t, dir, shellAs(dir, "h1", "workflow:helper", "Get-Content free.txt", true), 0, "")
}

func TestBonsaiFilesOnAnAgentsShellLine(t *testing.T) {
	dir := rolesProject(t, nil)
	refused := []string{"echo 0 > .bonsai/local/agents/x.calls", "echo x >> .bonsai/local/log/s-x.ndjson",
		"cp brief.md .bonsai/local/messages/builder-T-0001/M-2026-10-15-0000000a.md", "tee .bonsai/local/agents/ties/x < /dev/null",
		"sed -i s/x/y/ .bonsai/lock.json", "cp lock.json .bonsai/lock.json", "cp lock.json .bonsai/", "ln -sf /dev/null .bonsai/local/ladder/T-0001.json",
		"mv x .bonsai/local/agents/y", "echo x > .bonsai/tasks.md", "cd .bonsai/local && echo x > a"}
	for i, line := range refused {
		rule := RuleBonsaiFiles
		if strings.HasPrefix(line, "mv ") {
			rule = RuleDeleteProtected // the delete check judges a move's destination first, as before
		}
		for _, typ := range []string{"workflow:lead", "workflow:helper", "Explore"} {
			wantAnswer(t, dir, shellAs(dir, fmt.Sprint(typ[:1], i), typ, line, false), 2, rule)
		}
		// A named session too.
		wantAnswer(t, dir, shellAs(dir, "", "", line, false), 2, rule, RoleEnv, "lead")
	}
	// The person's own session keeps today's guard: a redirect there is not judged (the deny rule stands behind it).
	wantAnswer(t, dir, shellAs(dir, "", "", "echo x > .bonsai/local/agents/x", false), 0, "")
	// Bonsai's own commands write there as a program.
	wantAnswer(t, dir, shellAs(dir, "a1", "workflow:lead", "bonsai ladder --task T-0001 --json", false), 0, RuleBonsaiCommand)
	// From a worktree folder outside the main checkout, a path into the main checkout's .bonsai/local/.
	wt := filepath.Join(filepath.Dir(dir), "wt")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	line := "cd " + filepath.ToSlash(wt) + " && echo x > ../proj/.bonsai/local/agents/x" // forward slashes: Git Bash's form
	wantAnswer(t, dir, shellAs(dir, "a1", "workflow:lead", line, false), 2, RuleBonsaiFiles)
}

func TestBonsaiMessageFrom(t *testing.T) {
	dir := rolesProject(t, map[string]string{".bonsai/local/messages/lead-T-0001/M-2026-10-15-0000000a.md": briefFile(
		"M-2026-10-15-0000000a", "M-2026-10-15-0000000a", "orchestrator", "lead@T-0001", "2026-10-15T03:00:00.000Z", "")})
	agent := func(id, typ, command string) string {
		return hookCall{session: sessA, agentID: id, agentType: typ, tool: "Bash", input: bash(command)}.payload(dir)
	}
	// The session starts a lead with the brief named in its prompt; the lead's first call ties it: its address is
	// lead@T-0001.
	wantAnswer(t, dir, hookCall{session: sessA, tool: "Agent", use: "toolu_01LEAD", input: in("subagent_type", "workflow:lead",
		"prompt", "Your brief is M-2026-10-15-0000000a.")}.payload(dir), 0, RuleCallAllowed)
	wantAnswer(t, dir, agent("a1", "workflow:lead", "git status"), 0, RuleShell)
	if w := stateOf(t, dir, sessA, "a1").Window(); w.Tie == nil || w.Tie.To != "lead@T-0001" {
		t.Fatalf("the brief was not tied: %+v", w)
	}
	report := "bonsai message --kind report --from lead@T-0001 --to orchestrator --thread M-2026-10-15-0000000a --status done --body 'ok'"
	wantAnswer(t, dir, agent("a1", "workflow:lead", report), 0, RuleBonsaiCommand)
	wantAnswer(t, dir, agent("a1", "workflow:lead", strings.Replace(report, "--from lead@T-0001", "--from lead@T-0002", 1)), 2, RuleCallRefused)
	wantAnswer(t, dir, agent("a1", "workflow:lead", strings.Replace(report, "--from lead@T-0001", "--from=lead", 1)), 2, RuleCallRefused)
	wantAnswer(t, dir, agent("a1", "workflow:lead", strings.Replace(report, "--from lead@T-0001 ", "", 1)), 2, RuleCallRefused)
	// A brief from the lead: to a helper, allowed; to a lead, refused.
	b := "bonsai message --kind brief --from lead@T-0001 --to helper@T-0001 --body 'go'"
	wantAnswer(t, dir, agent("a1", "workflow:lead", b), 0, RuleBonsaiCommand)
	wantAnswer(t, dir, agent("a1", "workflow:lead", strings.Replace(b, "helper@", "verifier@", 1)), 2, RuleCallRefused)
	// An agent with no tied brief, and an undeclared type, have no address: every message they write is refused; a
	// read is free.
	wantAnswer(t, dir, agent("h1", "workflow:helper", "bonsai message --kind report --from helper --to lead --body 'x'"), 2, RuleCallRefused)
	wantAnswer(t, dir, agent("e1", "Explore", "bonsai message --kind report --from x --to lead --body 'x'"), 2, RuleCallRefused)
	wantAnswer(t, dir, agent("e1", "Explore", "bonsai message --show M-2026-10-15-0000000a"), 0, RuleBonsaiCommand)
	wantAnswer(t, dir, agent("e1", "Explore", "bonsai messages --json"), 0, RuleBonsaiCommand)
	// The person's session writes from its role's name.
	wantAnswer(t, dir, shellAs(dir, "", "", "bonsai message --kind brief --from orchestrator --to lead@T-0002 --body 'x'", false), 0, RuleBonsaiCommand)
	wantAnswer(t, dir, shellAs(dir, "", "", "bonsai message --kind brief --from lead@T-0001 --to helper --body 'x'", false), 2, RuleCallRefused)
	// A named session's address is <role>@<BONSAI_TASK>.
	wantAnswer(t, dir, shellAs(dir, "", "", strings.Replace(report, "lead@T-0001", "lead@T-0009", 1), false), 0, RuleBonsaiCommand,
		RoleEnv, "lead", "BONSAI_TASK", "T-0009")
	wantAnswer(t, dir, shellAs(dir, "", "", report, false), 2, RuleCallRefused, RoleEnv, "lead", "BONSAI_TASK", "T-0009")
}

// briefFile is a brief's message file; budget is its budget's lines ("" for null).
func briefFile(id, thread, from, to, created, budget string) string {
	b := "null"
	if budget != "" {
		b = "\n" + budget
	}
	return "---\nformat: bonsai.message/1   # fields: bonsai check --schema bonsai.message\nid: " + id + "\nthread: " + thread +
		"\nreply_to: null\ntask: T-0001\nkind: brief\nfrom: " + from + "\nto: " + to + "\nstatus: null\nbudget: " + b +
		"\npointers: []\ncreated: \"" + created + "\"\nlabels: {}\n---\nThe brief's body.\n"
}

// The report's exact form, past a budget.
func TestReportCall(t *testing.T) {
	addr := "helper@T-0001"
	ok := []string{
		"bonsai message --kind report --from helper@T-0001 --to lead@T-0001 --thread M-2026-10-15-0000000a --status blocked --body 'what I have; what is left'",
		"/usr/local/bin/bonsai message --kind=report --from=helper@T-0001 --to lead --status done --body 'a $HOME `x` $(y)'",
		"bonsai message --kind report --from helper@T-0001 --to lead --status done --body - <<'EOF'\nline $HOME $(rm -rf x)\nEOF\n",
		"bonsai message --kind report --from helper@T-0001 --to lead --status done --body - <<\"EOF\"\nline\nEOF\n",
		"bonsai message --kind report --from helper@T-0001 --to lead --status done --pointer notes.md --body 'x'",
	}
	for _, c := range ok {
		if !reportCall("Bash", c, addr) {
			t.Errorf("not read as the report: %s", c)
		}
	}
	bad := []string{
		"bonsai message --kind report --from helper@T-0001 --to lead --body 'x'; rm -rf y",
		"bonsai message --kind report --from helper@T-0001 --to lead --body 'x' && true",
		"bonsai message --kind report --from helper@T-0001 --to lead --body 'x' | cat",
		"bonsai message --kind report --from helper@T-0001 --to lead --body \"$(cat x)\"",
		"bonsai message --kind report --from helper@T-0001 --to lead --body \"x\"",
		"bonsai message --kind report --from helper@T-0001 --to lead --body x",
		"bonsai message --kind report --from helper@T-0001 --to $LEAD --body 'x'",
		"bonsai message --kind report --from helper@T-0001 --to lead --body - <<EOF\n$(rm x)\nEOF\n",
		"bonsai message --kind report --from helper@T-0001 --to lead --body-file notes.md",
		"bonsai message --kind report --from helper@T-0002 --to lead --body 'x'",
		"bonsai message --kind brief --from helper@T-0001 --to lead --body 'x'",
		"bonsai message --kind report --from helper@T-0001 --to lead --body 'x' > out.txt",
		"bonsai message --kind report --from helper@T-0001 --to lead --body 'x' &",
		"(bonsai message --kind report --from helper@T-0001 --to lead --body 'x')",
		"bonsai message --kind report --kind report --from helper@T-0001 --to lead --body 'x'",
		"bonsai message --kind report --from helper@T-0001 --to lead --body 'x' --body 'y'",
		"bonsai message --kind report --from helper@T-0001 --to lead --body - <<< 'x'",
		"echo x | bonsai message --kind report --from helper@T-0001 --to lead --body -",
		"sh -c \"bonsai message --kind report --from helper@T-0001 --to lead --body 'x'\"",
		"bonsai ladder --task T-0001",
		"bonsai message --show M-2026-10-15-0000000a",
	}
	for _, c := range bad {
		if reportCall("Bash", c, addr) {
			t.Errorf("read as the report: %s", c)
		}
	}
	if !reportCall("PowerShell", "bonsai message --kind report --from helper@T-0001 --to lead --status done --body 'it''s done'", addr) {
		t.Error("PowerShell's single-quoted body")
	}
	if reportCall("PowerShell", "bonsai message --kind report --from helper@T-0001 --to lead --body \"$x\"", addr) {
		t.Error("PowerShell's double-quoted body")
	}
	if briefShow("Bash", "bonsai message --show M-2026-10-15-0000000a --json") != "M-2026-10-15-0000000a" ||
		briefShow("Bash", "bonsai message --show M-2026-10-15-0000000a; rm x") != "" || briefShow("Bash", "bonsai message --show $ID") != "" {
		t.Error("briefShow")
	}
}

// Every word of the tables is documented, and the git words' read forms read.
func TestTheRoleWordTables(t *testing.T) {
	for _, g := range GitWriteWords {
		if g.Sub == "" || strings.ToLower(g.Sub) != g.Sub {
			t.Errorf("%+v", g)
		}
	}
	// Each form, as git reads it (measured on git 2.x, 15 Oct: branch -v <name> and tag --sort=<key> <name> make the ref).
	words := func(line string) []shWord {
		var out []shWord
		for _, f := range strings.Fields(line) {
			out = append(out, shWord{text: f})
		}
		return out
	}
	for _, c := range []struct {
		line  string
		write bool
	}{
		{"status", false}, {"log --oneline", false}, {"diff HEAD", false},
		{"branch", false}, {"branch -v", false}, {"branch -vv --color", false}, {"branch -a", false}, {"branch -av", false},
		{"branch --list feat*", false}, {"branch --contains HEAD", false}, {"branch --show-current", false},
		{"branch new", true}, {"branch -v new", true}, {"branch -vv new", true}, {"branch --sort=refname new", true},
		{"branch -d x", true}, {"branch -dr origin/x", true}, {"branch -l -D x", true}, {"branch --list --delete x", true},
		{"branch --set-upstream-to=origin/x", true}, {"branch -m a b", true}, {"branch -- -v", true},
		{"tag", false}, {"tag -l", false}, {"tag -n", false}, {"tag --sort=refname", false}, {"tag -v v1", false},
		{"tag v1", true}, {"tag --sort=refname v1", true}, {"tag -a v1 -m x", true}, {"tag -l -d v1", true},
		{"stash", true}, {"stash list", false}, {"stash show -p", false}, {"stash push -m list", true}, {"stash -p", true},
		{"worktree list --porcelain", false}, {"worktree add list", true}, {"worktree", true},
		{"notes list", false}, {"notes add -m show", true},
		{"apply --check x.patch", false}, {"apply --stat x.patch", false}, {"apply --stat --apply x.patch", true}, {"apply x.patch", true},
		{"clean -n", false}, {"clean -fn", false}, {"clean -f", true}, {"rm -n x", false}, {"rm x", true},
		{"push --dry-run", false}, {"push origin main", true},
		{"replace -l", false}, {"replace -l -d x", true}, {"replace a b", true},
		{"bisect log", false}, {"bisect start", true}, {"bisect good", true}, {"clone url x", true}, {"init", true},
		{"submodule status", false}, {"submodule update --init", true}, {"sparse-checkout list", false},
		{"sparse-checkout set a", true}, {"read-tree -u HEAD", true}, {"checkout-index -a", true},
		{"merge-file -p a b c", false}, {"merge-file a b c", true}, {"format-patch --stdout -1", false},
		{"format-patch -1", true}, {"rerere status", false}, {"rerere forget x", true},
		{"config --get user.name", false}, {"config -l", false}, {"config get user.name", false}, {"config list", false},
		{"config alias.ci commit", true}, {"config user.name", true}, {"config --get-all x --unset x", true}, {"config set a b", true},
	} {
		sub, rest := gitSub(words(c.line))
		if got := gitWrite(sub, rest) != ""; got != c.write {
			t.Errorf("git %s: a write %v, want %v", c.line, got, c.write)
		}
	}
	for line, want := range map[string]bool{"-c alias.ci=commit ci -m x": true, "-calias.st=status st": true,
		"--config-env=alias.x=Y x": true, "-c core.pager=cat log": false, "-C . -c Alias.X=commit X": true, "log -c alias.x=y": false} {
		if got := gitInlineAlias(words(line)); got != want {
			t.Errorf("git %s: an inline alias %v, want %v", line, got, want)
		}
	}
	if claudeStart("claude", []shWord{{text: "claude"}}) == "" || claudeStart("claudette", []shWord{{text: "claudette"}}) != "" {
		t.Error("claudeStart")
	}
}
