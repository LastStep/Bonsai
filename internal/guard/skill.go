package guard

// The Skill tool, where the roles apply (VG's item A, 15 Oct: a helper ran a skill whose frontmatter forks into the
// builder, context: fork and agent: builder, and the forked builder wrote). A skill that forks starts an agent, so its
// call is a call to that agent's type, held to who may call whom (mayCall) like an Agent call, and kept for the
// session as a spawn call (Tool Skill), so the agent it starts is joined to it and plays its role (roles.go's pick and
// backs); a skill that runs in its caller's own context starts no one and is let through.
//
// Its frontmatter is read at the Skill call, by its name (tool_input.skill), from these files, each a regular file of
// at most maxSkillFile bytes: <folder>/.claude/skills/<name>/SKILL.md and <folder>/.claude/commands/<name>.md for the
// session's own folder (CLAUDE_PROJECT_DIR) and its main checkout, and skills/<name>/SKILL.md and commands/<name>.md
// in the person's own Claude Code folder (CLAUDE_CONFIG_DIR, else .claude in the home folder). Every one found is
// judged, so a name two of them hold is held to the stricter. Its frontmatter's top-level context: fork (letter case
// ignored) forks, into agent: <type>, or Claude Code's general-purpose agent where it names none.
//
// Known limits: a plugin's skill (<plugin>:<name>), a skill in a nested .claude/skills folder Claude Code finds below
// the session's folder or in an --add-dir folder, and a file changed between this read and Claude Code's own are not
// read here, nor is frontmatter Bonsai's line reading cannot follow (a block scalar, a flow mapping). What such a skill
// forks is never given a role's rights: its agent has no allowed call for that role behind it, so it runs as a type no
// pack declares (roles.go's backs): it changes nothing, calls no one, and runs on a helper's budget.

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/hookin"
)

// maxSkillFile caps a skill's file the guard reads, in bytes.
const maxSkillFile = 256 << 10

// skillCall judges a Skill call where the roles apply (this file's comment).
func (rr *roleRun) skillCall(w *who) Decision {
	name, args := "", ""
	if o := rr.in.InputObject(); o != nil {
		if v, ok := o.Get("skill"); ok {
			name, _ = v.(string)
		}
		if v, ok := o.Get("args"); ok {
			args, _ = v.(string)
		}
	}
	forks := rr.skillForks(name)
	if len(forks) == 0 {
		return allow(RuleOtherTool, "")
	}
	rs, err := rr.roles()
	if err != nil {
		return rolesFault(err)
	}
	for _, t := range forks {
		c := rr.resolveCallee(rs, t)
		if why := mayCall(w, c); why != "" {
			return deny(RuleCallRefused, fmt.Sprintf("Skill refused: the skill %s forks into %s (its frontmatter's context: "+
				"fork), and %s may not call it: %s (who may call whom: it may call %s)", quote(name, 60), c.shown, w.shown(), why,
				rolesCalled(w)), "do the skill's steps yourself within your role, or hand the work back to your caller in your "+
				"report (bonsai message --kind report), naming what is left")
		}
	}
	for _, t := range forks {
		role := ""
		if c := rr.resolveCallee(rs, t); c.role != nil {
			role = c.role.Role.Name
		}
		rr.addSpawn(w, agents.Spawn{Tool: "Skill", Type: t, Role: role, Briefs: briefIDs(args)})
	}
	return allow(RuleCallAllowed, orGeneral(forks[0]))
}

// skillForks gives the types a skill's files fork into (this file's comment), in the order found, each once; none
// for a skill that forks nowhere, or whose name names no file read here.
func (rr *roleRun) skillForks(name string) []string {
	if !skillName(name) {
		return nil
	}
	var files []string
	for _, root := range []string{rr.j.getenv(hookin.ProjectEnv), rr.main} {
		if root == "" {
			continue
		}
		files = append(files, filepath.Join(root, ".claude", "skills", name, "SKILL.md"),
			filepath.Join(root, ".claude", "commands", name+".md"))
	}
	if cfg := rr.claudeFolder(); cfg != "" {
		files = append(files, filepath.Join(cfg, "skills", name, "SKILL.md"), filepath.Join(cfg, "commands", name+".md"))
	}
	var out []string
	seen := map[string]bool{}
	for _, f := range files {
		if seen[f] {
			continue
		}
		seen[f] = true
		raw, err := agents.ReadRegular(f, maxSkillFile)
		if err != nil {
			continue
		}
		if t, ok := skillFork(raw); ok && !listHas(out, t) {
			out = append(out, t)
		}
	}
	return out
}

// claudeFolder is the person's own Claude Code folder: CLAUDE_CONFIG_DIR, else .claude in the home folder.
func (rr *roleRun) claudeFolder() string {
	if d := rr.j.getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	home := rr.j.getenv("HOME")
	if runtime.GOOS == "windows" || home == "" {
		if p := rr.j.getenv("USERPROFILE"); p != "" {
			home = p
		}
	}
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".claude")
}

// skillName reports a name a skill's file can be found by: letters, digits, '-', '_' and '.', not starting with a
// dot, at most 128; a plugin's (<plugin>:<name>) or any other is not read here.
func skillName(name string) bool {
	if name == "" || len(name) > 128 || name[0] == '.' {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}

// skillFork reads a skill's frontmatter for context: fork and agent (this file's comment): the type it forks into
// ("" for Claude Code's general-purpose agent), and whether it forks.
func skillFork(raw []byte) (string, bool) {
	text := strings.ReplaceAll(strings.TrimPrefix(string(raw), "\ufeff"), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", false
	}
	context, agent := "", ""
	for _, l := range lines[1:] {
		if t := strings.TrimSpace(l); t == "---" || t == "..." {
			break
		}
		if l == "" || l[0] == ' ' || l[0] == '\t' || l[0] == '#' {
			continue
		}
		k, v, ok := strings.Cut(l, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "context":
			context = yamlScalar(v)
		case "agent":
			agent = yamlScalar(v)
		}
	}
	if !strings.EqualFold(context, "fork") {
		return "", false
	}
	return agent, true
}

// yamlScalar reads a one-line YAML scalar: in single or double quotes, else up to a comment (" #"), trimmed.
func yamlScalar(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '\'' || v[0] == '"') {
		if end := strings.IndexByte(v[1:], v[0]); end >= 0 {
			return v[1 : end+1]
		}
	}
	if k := strings.Index(v, " #"); k >= 0 {
		v = v[:k]
	}
	return strings.TrimSpace(v)
}
