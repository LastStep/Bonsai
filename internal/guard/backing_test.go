package guard

// VG's items A and B (15 Oct), each reproducer a test: a declared role's rights go only to an agent joined to a call
// the guard allowed for that role (an Agent, Task or Skill call), a Skill whose frontmatter forks into a declared role
// is a call to that role, held to who may call whom, and every agent type is matched with its letter case folded, as
// Claude Code matches a subagent_type. On the fixture's roles: lead (a lead that writes, the measured builder's place),
// helper, verifier.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// forkySkill is VG's skill: it forks into the lead role (the builder in VG's run).
const forkySkill = "---\nname: forky\ndescription: a scratch skill that runs in a forked lead.\ncontext: fork\nagent: lead\n---\n" +
	"1. Edit free.txt. 2. echo forked > forked.txt.\n"

// skillProject is a roles project holding VG's forked skill, a skill that runs inline, and a command that forks.
func skillProject(t *testing.T) string {
	t.Helper()
	return rolesProject(t, map[string]string{
		".claude/skills/forky/SKILL.md":  forkySkill,
		".claude/skills/inline/SKILL.md": "---\nname: inline\ndescription: runs in its caller's context.\n---\nRead free.txt.\n",
		".claude/commands/forkcmd.md":    "---\r\ndescription: \"a command\" # forks\r\ncontext: 'fork'\r\nagent: \"workflow:lead\" # the lead\r\n---\r\nGo.\r\n",
	})
}

// wantSecond is wantAnswer on the guard's second line, which judges the Skill tool.
func wantSecond(t *testing.T, dir, payload string, code int, rule string) string {
	t.Helper()
	got, stderr := runAt(t, dir, payload, testNow, true)
	if got != code {
		t.Fatalf("second line: exit %d, want %d (%s): %s\n%s", got, code, rule, stderr, payload)
	}
	if rule != "" {
		if r := lastRule(t, dir, payload); r != rule {
			t.Fatalf("second line: rule %q, want %q: %s", r, rule, stderr)
		}
	}
	return stderr
}

// metaFile writes Claude Code's meta file beside a subagent's transcript, as measured for a skill's fork (agentType
// and parentAgentId, no toolUseId) or an Agent call's child (toolUseId given).
func metaFile(t *testing.T, transcripts, session, agent, typ, parent, use string) {
	t.Helper()
	body := `{"agentType":"` + typ + `"`
	if use != "" {
		body += `,"toolUseId":"` + use + `"`
	}
	if parent != "" {
		body += `,"parentAgentId":"` + parent + `"`
	}
	writeFile(t, filepath.Join(transcripts, session, "subagents", "agent-"+agent+".meta.json"), body+`,"spawnDepth":2}`)
}

// A1, VG's reproducer: a helper started by the person's session runs the forked skill; the guard refuses it, and the
// forked lead, which no allowed call started, writes nothing.
func TestAForkedSkillFromAHelper(t *testing.T) {
	dir := skillProject(t)
	tr := t.TempDir()
	transcript := filepath.ToSlash(filepath.Join(tr, sessA+".jsonl"))
	started(t, dir, sessA, "helper")
	helper := func(tool string, input schema.Object) string {
		return hookCall{session: sessA, agentID: "a8e88a75e579776bc", agentType: "helper", tool: tool, transcript: transcript,
			input: input}.payload(dir)
	}
	wantAnswer(t, dir, helper("Bash", bash("git status")), 0, RuleShell)
	s := wantSecond(t, dir, helper("Skill", in("skill", "forky")), 2, RuleCallRefused)
	for _, want := range []string{"Skill refused: the skill \"forky\" forks into the lead role \"workflow:lead\"",
		"a helper, and every type no pack declares, calls no one", "next: "} {
		if !strings.Contains(s, want) {
			t.Errorf("the refusal has no %q:\n%s", want, s)
		}
	}
	// Were the fork to start anyway (a skill the guard could not read): its lead holds no rights.
	metaFile(t, tr, sessA, "a3699b6f7263650a6", "lead", "a8e88a75e579776bc", "")
	child := func(tool string, input schema.Object) string {
		return hookCall{session: sessA, agentID: "a3699b6f7263650a6", agentType: "lead", tool: tool, transcript: transcript,
			input: input}.payload(dir)
	}
	s = wantAnswer(t, dir, child("Edit", in("file_path", filepath.Join(dir, "free.txt"), "old_string", "free", "new_string", "forked")),
		2, RuleRoleWritesNothing)
	if !strings.Contains(s, "which no call the guard allowed for that role started") {
		t.Errorf("%s", s)
	}
	wantAnswer(t, dir, child("Bash", bash("echo forked > forked.txt")), 2, RuleRoleWritesNothing)
	st := stateOf(t, dir, sessA, "a3699b6f7263650a6")
	if st.Header.Role != "" || st.Header.Writes || !st.Header.Unbacked || st.Header.Kind != "helper" {
		t.Errorf("the forked lead's header %+v", st.Header)
	}
	found := false
	for _, r := range sessionRecords(t, dir, child("Read", in("file_path", "x"))) {
		found = found || r.String("rule") == RuleRoleUnbacked && strings.Contains(r.String("text"), "names the lead role")
	}
	if !found {
		t.Error("no role-unbacked record")
	}
}

// A2: the person's session runs the forked skill (allowed: it may call a lead), and the lead it forks, joined to that
// Skill call by its meta file naming no call, plays its role: its Edit of a free file goes through.
func TestAForkedSkillFromTheSession(t *testing.T) {
	dir := skillProject(t)
	tr := t.TempDir()
	transcript := filepath.ToSlash(filepath.Join(tr, sessA+".jsonl"))
	// An Agent call for a lead stays open beside it: the fork, whose meta file names no call, is joined to the Skill
	// call alone, never to that Agent call (whose own child its meta file names).
	started(t, dir, sessA, "workflow:lead")
	wantSecond(t, dir, hookCall{session: sessA, tool: "Skill", use: "toolu_01SKILL", transcript: transcript,
		input: in("skill", "forky")}.payload(dir), 0, RuleCallAllowed)
	metaFile(t, tr, sessA, "af0000000000000a1", "lead", "", "")
	wantAnswer(t, dir, hookCall{session: sessA, agentID: "af0000000000000a1", agentType: "lead", tool: "Edit", transcript: transcript,
		input: in("file_path", filepath.Join(dir, "free.txt"), "old_string", "free", "new_string", "forked")}.payload(dir), 0, RuleNotProtected)
	st := stateOf(t, dir, sessA, "af0000000000000a1")
	if st.Header.Role != "lead" || !st.Header.Writes || st.Link() == nil || st.Link().Call != "toolu_01SKILL" || st.Link().Parent != "session" {
		t.Errorf("the fork: %+v %+v", st.Header, st.Link())
	}
	// A command that forks (CRLF, quotes, comments) is read the same way: a lead may not run it (a lead calls no lead).
	started(t, dir, sessA, "workflow:lead")
	wantAnswer(t, dir, hookCall{session: sessA, agentID: "al1", agentType: "workflow:lead", tool: "Read", input: in("file_path", "x")}.payload(dir), 0, RuleOtherTool)
	wantSecond(t, dir, hookCall{session: sessA, agentID: "al1", agentType: "workflow:lead", tool: "Skill",
		input: in("skill", "forkcmd")}.payload(dir), 2, RuleCallRefused)
	wantSecond(t, dir, hookCall{session: sessA, agentID: "al1", agentType: "workflow:lead", tool: "Skill",
		input: in("skill", "forky")}.payload(dir), 2, RuleCallRefused)
	wantSecond(t, dir, hookCall{session: sessA, agentID: "al1", agentType: "workflow:lead", tool: "Skill",
		input: in("skill", "/forky", "args", "x")}.payload(dir), 2, RuleCallRefused)
	// A skill that runs inline, one named nowhere, and a plugin's (not read here) are let through to anyone.
	for _, name := range []string{"inline", "nowhere", "workflow:forky", "../forky"} {
		wantSecond(t, dir, hookCall{session: sessA, agentID: "al1", agentType: "workflow:lead", tool: "Skill",
			input: in("skill", name)}.payload(dir), 0, RuleOtherTool)
	}
	// The person's own Claude Code folder is read too (CLAUDE_CONFIG_DIR).
	cfg := t.TempDir()
	writeFile(t, filepath.Join(cfg, "skills", "mine", "SKILL.md"), forkySkill)
	p := hookCall{session: sessA, agentID: "al1", agentType: "workflow:lead", tool: "Skill", input: in("skill", "mine")}.payload(dir)
	if code, s := runAt(t, dir, p, testNow, true, "CLAUDE_CONFIG_DIR", cfg); code != 2 || !strings.Contains(s, "Skill refused") {
		t.Errorf("a personal skill that forks: %d %s", code, s)
	}
}

// A3: a subagent whose type names a declared role but whose starting call asked for another type, or which no
// allowed call started, is held as a type no pack declares.
func TestARoleOnlyBehindAnAllowedCall(t *testing.T) {
	dir := rolesProject(t, nil)
	tr := t.TempDir()
	transcript := filepath.ToSlash(filepath.Join(tr, sessA+".jsonl"))
	edit := func(id, typ string) string {
		return hookCall{session: sessA, agentID: id, agentType: typ, tool: "Edit", transcript: transcript,
			input: in("file_path", filepath.Join(dir, "free.txt"), "old_string", "f", "new_string", "g")}.payload(dir)
	}
	// No call at all.
	wantAnswer(t, dir, edit("u1", "workflow:lead"), 2, RuleRoleWritesNothing)
	// Its meta file names a call that asked for Explore.
	wantAnswer(t, dir, hookCall{session: sessA, tool: "Agent", use: "toolu_01EXPL", input: in("subagent_type", "Explore", "prompt", "x")}.payload(dir), 0, RuleCallAllowed)
	metaFile(t, tr, sessA, "u2", "workflow:lead", "", "toolu_01EXPL")
	wantAnswer(t, dir, edit("u2", "workflow:lead"), 2, RuleRoleWritesNothing)
	// Its meta file names a call for a lead another agent already holds: it does not take it.
	wantAnswer(t, dir, hookCall{session: sessA, tool: "Agent", use: "toolu_01LEAD", input: in("subagent_type", "lead", "prompt", "x")}.payload(dir), 0, RuleCallAllowed)
	metaFile(t, tr, sessA, "u3", "workflow:lead", "", "toolu_01LEAD")
	wantAnswer(t, dir, edit("u3", "workflow:lead"), 0, RuleNotProtected)
	metaFile(t, tr, sessA, "u4", "workflow:lead", "", "toolu_01LEAD")
	wantAnswer(t, dir, edit("u4", "workflow:lead"), 2, RuleRoleWritesNothing)
	// An undeclared type started by no call is as before (held as a helper), and not marked unbacked.
	wantAnswer(t, dir, edit("u5", "Explore"), 2, RuleRoleWritesNothing)
	if st := stateOf(t, dir, sessA, "u5"); st.Header.Unbacked {
		t.Errorf("an undeclared type marked unbacked: %+v", st.Header)
	}
	// A state moved aside and made afresh keeps the call it was joined to: u3 is a lead again.
	writeFile(t, filepath.Join(dir, ".bonsai", "local", "agents", sessA, "u3.state"), "not a header\n")
	wantAnswer(t, dir, edit("u3", "workflow:lead"), 0, RuleNotProtected)
}

// B, VG's reproducer: a case variant of a declared role's type is that role, as Claude Code matches it. A lead's call
// for "Lead" is a lead calling a lead (refused); the session's call for "LEAD" starts a lead (its child, typed as the
// agent's own name, plays it); an --agent session typed "Workflow:Lead", and BONSAI_ROLE "Lead", play the lead role.
func TestBTypesMatchedWithLetterCaseFolded(t *testing.T) {
	dir := rolesProject(t, nil)
	started(t, dir, sessA, "workflow:lead")
	lead := func(tool string, input schema.Object) string {
		return hookCall{session: sessA, agentID: "a1", agentType: "workflow:lead", tool: tool, input: input}.payload(dir)
	}
	wantAnswer(t, dir, lead("Read", in("file_path", "x")), 0, RuleOtherTool)
	for _, typ := range []string{"Lead", "LEAD", "Workflow:Lead", "workflow:VERIFIER"} {
		wantAnswer(t, dir, lead("Agent", in("subagent_type", typ, "prompt", "x")), 2, RuleCallRefused)
	}
	wantAnswer(t, dir, lead("Agent", in("subagent_type", "Helper", "prompt", "x")), 0, RuleCallAllowed)
	wantAnswer(t, dir, hookCall{session: sessA, tool: "Agent", input: in("subagent_type", "Workflow:Orchestrator", "prompt", "x")}.payload(dir),
		2, RuleCallRefused)
	wantAnswer(t, dir, hookCall{session: sessA, tool: "Agent", input: in("subagent_type", "LEAD", "prompt", "x")}.payload(dir), 0, RuleCallAllowed)
	wantAnswer(t, dir, hookCall{session: sessA, agentID: "b2", agentType: "lead", tool: "Edit",
		input: in("file_path", filepath.Join(dir, "free.txt"), "old_string", "f", "new_string", "g")}.payload(dir), 0, RuleNotProtected)
	if st := stateOf(t, dir, sessA, "b2"); st.Header.Role != "lead" || st.Link() == nil {
		t.Errorf("the session's LEAD call's child: %+v %+v", st.Header, st.Link())
	}
	named := hookCall{session: sessB, agentType: "Workflow:Lead", tool: "Read", input: in("file_path", "x")}.payload(dir)
	wantAnswer(t, dir, named, 0, RuleOtherTool)
	if st := stateOf(t, dir, sessB, sessB); st.Header.Role != "lead" || !st.Header.Named {
		t.Errorf("an --agent Workflow:Lead session: %+v", st.Header)
	}
	byEnv := hookCall{session: "5e3cffd5-bd72-4f4a-a089-220fd2fdbff6", tool: "Read", input: in("file_path", "x")}.payload(dir)
	wantAnswer(t, dir, byEnv, 0, RuleOtherTool, RoleEnv, "Lead")
	if st := stateOf(t, dir, "5e3cffd5-bd72-4f4a-a089-220fd2fdbff6", "5e3cffd5-bd72-4f4a-a089-220fd2fdbff6"); st.Header.Role != "lead" {
		t.Errorf("BONSAI_ROLE=Lead: %+v", st.Header)
	}
}

// The frontmatter of a skill, as skillFork reads it.
func TestSkillFrontmatter(t *testing.T) {
	for _, c := range []struct {
		raw   string
		agent string
		forks bool
	}{
		{"---\ncontext: fork\nagent: lead\n---\n", "lead", true},
		{"\ufeff---\r\ncontext: \"fork\" # yes\r\nagent: 'workflow:lead'\r\n---\r\n", "workflow:lead", true},
		{"---\ncontext: Fork\n---\nagent: lead\n", "", true},
		{"---\ncontext: fork\nmetadata:\n  agent: lead\n---\n", "", true},
		{"---\ncontext: inline\nagent: lead\n---\n", "", false},
		{"---\nname: x\n---\ncontext: fork\n", "", false},
		{"context: fork\nagent: lead\n", "", false},
		{"---\ncontext: fork\nagent: lead\n", "lead", true},
		{"", "", false},
	} {
		if agent, forks := skillFork([]byte(c.raw)); agent != c.agent || forks != c.forks {
			t.Errorf("%q: %q %v, want %q %v", c.raw, agent, forks, c.agent, c.forks)
		}
	}
}

// The row's test in TestEveryRule (rules_test.go's ruleTests).
func init() {
	ruleTests[RuleRoleUnbacked] = func(t *testing.T) {
		dir := rolesProject(t, nil)
		p := hookCall{session: sessA, agentID: "u1", agentType: "workflow:lead", tool: "Glob", input: in("pattern", "*")}.payload(dir)
		wantAnswer(t, dir, p, 0, RuleOtherTool)
		found := false
		for _, r := range sessionRecords(t, dir, p) {
			found = found || r.String("rule") == RuleRoleUnbacked && r.String("decision") == "allow"
		}
		if !found {
			t.Error("no role-unbacked record")
		}
	}
}
